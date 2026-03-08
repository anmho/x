package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const controlPlaneConfigFileName = "platform.controlplane.json"

type controlPlaneConfig struct {
	Version  string                `json:"version"`
	Accounts []controlPlaneAccount `json:"accounts,omitempty"`
	Projects []controlPlaneProject `json:"projects"`
}

type controlPlaneAccount struct {
	Name    string       `json:"name"`
	Secrets []secretSpec `json:"secrets,omitempty"`
}

type controlPlaneProject struct {
	Name         string           `json:"name"`
	DesiredState string           `json:"desired_state,omitempty"` // present|absent
	GCP          *controlPlaneGCP `json:"gcp,omitempty"`
	Deployments  []deploymentSpec `json:"deployments,omitempty"`
	Domains      []domainZoneSpec `json:"domains,omitempty"`
	Secrets      []secretSpec     `json:"secrets,omitempty"`
}

type controlPlaneGCP struct {
	ProjectID string   `json:"project_id"`
	Region    string   `json:"region,omitempty"`
	Services  []string `json:"services,omitempty"`
}

type deploymentSpec struct {
	Name         string `json:"name"`
	Provider     string `json:"provider,omitempty"`      // gcp-cloud-run
	Service      string `json:"service,omitempty"`       // cloud run service name
	Region       string `json:"region,omitempty"`        // deployment region
	DesiredState string `json:"desired_state,omitempty"` // present|absent
}

type secretSpec struct {
	Name         string            `json:"name"`
	Provider     string            `json:"provider,omitempty"`      // legacy default platform
	DesiredState string            `json:"desired_state,omitempty"` // present|absent
	SourceEnv    string            `json:"source_env,omitempty"`    // local env var for secret value
	Shares       []secretShareSpec `json:"shares,omitempty"`
}

type secretShareSpec struct {
	Platform     string   `json:"platform"` // gcp|vercel
	ProjectID    string   `json:"project_id,omitempty"`
	Name         string   `json:"name,omitempty"`
	TargetType   string   `json:"target_type,omitempty"` // project|application|deployment
	TargetID     string   `json:"target_id,omitempty"`
	Project      string   `json:"project,omitempty"`      // logical project owner for application/deployment shares
	Environments []string `json:"environments,omitempty"` // vercel-only
}

func runControlPlane(args []string) error {
	if len(args) == 0 {
		return errors.New("missing control-plane subcommand (init|show|plan|apply|destroy)")
	}
	switch args[0] {
	case "init":
		return runControlPlaneInit(args[1:])
	case "show":
		return runControlPlaneShow()
	case "plan":
		return runControlPlanePlanOrApply(args[1:], true)
	case "apply":
		return runControlPlanePlanOrApply(args[1:], false)
	case "destroy":
		return runControlPlaneDestroy(args[1:])
	default:
		return fmt.Errorf("unknown control-plane subcommand %q", args[0])
	}
}

func runControlPlaneInit(args []string) error {
	fs := flag.NewFlagSet("control-plane init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	force := fs.Bool("force", false, "overwrite existing config")
	if err := fs.Parse(args); err != nil {
		return err
	}

	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}
	path := filepath.Join(repoRoot, controlPlaneConfigFileName)
	if _, err := os.Stat(path); err == nil && !*force {
		return fmt.Errorf("%s already exists (use --force to overwrite)", path)
	}

	cfg, err := sampleControlPlaneConfigFromProjects()
	if err != nil {
		return err
	}
	body, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return err
	}
	fmt.Printf("created %s\n", path)
	return nil
}

func runControlPlaneShow() error {
	cfg, path, err := loadControlPlaneConfig(true)
	if err != nil {
		return err
	}
	body, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("config: %s\n", path)
	fmt.Println(string(body))
	return nil
}

func runControlPlanePlanOrApply(args []string, dryRun bool) error {
	fs := flag.NewFlagSet("control-plane plan/apply", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	projectName := fs.String("project", "", "project name")
	prune := fs.Bool("prune", false, "apply deletions for desired_state=absent")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, _, err := loadControlPlaneConfig(true)
	if err != nil {
		return err
	}
	return reconcileControlPlane(cfg, *projectName, dryRun, *prune)
}

func runControlPlaneDestroy(args []string) error {
	fs := flag.NewFlagSet("control-plane destroy", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	projectName := fs.String("project", "", "project name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*projectName) == "" {
		return errors.New("usage: platform control-plane destroy --project <name>")
	}

	cfg, _, err := loadControlPlaneConfig(true)
	if err != nil {
		return err
	}
	return reconcileControlPlane(cfg, *projectName, false, true)
}

func loadControlPlaneConfig(required bool) (*controlPlaneConfig, string, error) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return nil, "", err
	}
	path := filepath.Join(repoRoot, controlPlaneConfigFileName)
	body, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && !required {
			return nil, path, nil
		}
		if errors.Is(err, os.ErrNotExist) {
			return nil, "", fmt.Errorf("%s not found. Run: ./platform control-plane init", controlPlaneConfigFileName)
		}
		return nil, "", err
	}

	var cfg controlPlaneConfig
	if err := json.Unmarshal(body, &cfg); err != nil {
		return nil, "", fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return &cfg, path, nil
}

func sampleControlPlaneConfigFromProjects() (*controlPlaneConfig, error) {
	projectCfg, _, err := loadProjectConfig(true)
	if err != nil {
		return nil, err
	}
	out := &controlPlaneConfig{Version: "1", Projects: []controlPlaneProject{}}
	out.Accounts = []controlPlaneAccount{
		{
			Name:    "default",
			Secrets: []secretSpec{},
		},
	}
	for _, project := range projectCfg.Projects {
		entry := controlPlaneProject{
			Name:         project.Name,
			DesiredState: "present",
		}
		if len(project.Targets.GCP) > 0 {
			target := project.Targets.GCP[0]
			entry.GCP = &controlPlaneGCP{
				ProjectID: target.ProjectID,
				Region:    target.Region,
				Services:  normalizeGCPServiceAPIs(target.Services),
			}
		}
		out.Projects = append(out.Projects, entry)
	}
	return out, nil
}

func reconcileControlPlane(cfg *controlPlaneConfig, selectedProject string, dryRun bool, prune bool) error {
	if len(cfg.Projects) == 0 && len(cfg.Accounts) == 0 {
		return errors.New("control-plane config has no projects or accounts")
	}
	selectedProject = strings.TrimSpace(selectedProject)

	for _, account := range cfg.Accounts {
		accountName := strings.TrimSpace(account.Name)
		if accountName == "" {
			accountName = "default"
		}
		fmt.Printf("control-plane: account=%s secrets=%d\n", accountName, len(account.Secrets))
		if !dryRun && secretSetNeedsGCP(account.Secrets, "", selectedProject, true) && !commandAvailable("gcloud") {
			return errors.New("gcloud is required for configured GCP control-plane actions")
		}
		if !dryRun && secretSetNeedsVercel(account.Secrets, "", selectedProject, true) && !commandAvailable("vercel") {
			return errors.New("vercel CLI is required for configured Vercel secret sharing actions")
		}
		if err := reconcileSecretSet(secretReconcileOptions{
			OwnerLabel:        "account " + accountName,
			Secrets:           account.Secrets,
			FallbackProjectID: "",
			DryRun:            dryRun,
			Prune:             prune,
			SelectedProject:   selectedProject,
			AccountLevel:      true,
		}); err != nil {
			return err
		}
	}

	for _, project := range cfg.Projects {
		if selectedProject != "" && project.Name != selectedProject {
			continue
		}

		state := strings.TrimSpace(project.DesiredState)
		if state == "" {
			state = "present"
		}

		projectID := ""
		apis := []string{}
		if project.GCP != nil {
			projectID = strings.TrimSpace(project.GCP.ProjectID)
			apis = normalizeGCPServiceAPIs(project.GCP.Services)
		}
		displayProjectID := projectID
		if displayProjectID == "" {
			displayProjectID = "n/a"
		}
		fmt.Printf("control-plane: project=%s gcp=%s desired=%s\n", project.Name, displayProjectID, state)

		if !dryRun && projectNeedsGCP(project, projectID) && !commandAvailable("gcloud") {
			return errors.New("gcloud is required for configured GCP control-plane actions")
		}
		if !dryRun && projectNeedsVercel(project) && !commandAvailable("vercel") {
			return errors.New("vercel CLI is required for configured Vercel secret sharing actions")
		}

		if dryRun {
			if state == "absent" {
				if projectID != "" {
					fmt.Printf("control-plane: would delete gcp project %s\n", projectID)
				}
				if err := reconcileSecretSet(secretReconcileOptions{
					OwnerLabel:        "project " + project.Name,
					Secrets:           project.Secrets,
					FallbackProjectID: projectID,
					DryRun:            true,
					Prune:             prune,
				}); err != nil {
					return err
				}
				continue
			}
			if projectID != "" {
				fmt.Printf("control-plane: would ensure gcp project exists %s\n", projectID)
				fmt.Printf("control-plane: would link gcloud config to %s\n", projectID)
				if len(apis) > 0 {
					fmt.Printf("control-plane: would enable APIs for %s: %s\n", projectID, strings.Join(apis, ", "))
				}
			} else {
				fmt.Printf("control-plane: no gcp.project_id for %s; skipping gcp infra operations\n", project.Name)
			}
			if err := reconcileSecretSet(secretReconcileOptions{
				OwnerLabel:        "project " + project.Name,
				Secrets:           project.Secrets,
				FallbackProjectID: projectID,
				DryRun:            true,
				Prune:             prune,
			}); err != nil {
				return err
			}
			if projectID != "" {
				if err := reconcileDeploymentState(project, projectID, true, prune); err != nil {
					return err
				}
			} else if len(project.Deployments) > 0 {
				fmt.Printf("control-plane: skip deployments for %s (no gcp.project_id)\n", project.Name)
			}
			if len(project.Domains) > 0 {
				if err := reconcileDomainSet(project.Name, project.Domains, true, prune); err != nil {
					return err
				}
			}
			continue
		}

		if state == "absent" {
			if !prune {
				fmt.Printf("control-plane: skip deletion for %s (use --prune or destroy)\n", project.Name)
				continue
			}
			if err := reconcileSecretSet(secretReconcileOptions{
				OwnerLabel:        "project " + project.Name,
				Secrets:           project.Secrets,
				FallbackProjectID: projectID,
				DryRun:            false,
				Prune:             true,
			}); err != nil {
				return err
			}
			if projectID != "" {
				if err := gcloudDeleteProject(projectID); err != nil {
					return err
				}
				fmt.Printf("control-plane: deleted gcp project %s\n", projectID)
			} else {
				fmt.Printf("control-plane: no gcp project to delete for %s\n", project.Name)
			}
			continue
		}

		if projectID == "" {
			fmt.Printf("control-plane: no gcp.project_id for %s; skipping gcp infra operations\n", project.Name)
		} else {
			exists, err := gcloudProjectExists(projectID)
			if err != nil {
				return err
			}
			if !exists {
				if dryRun {
					fmt.Printf("control-plane: would create gcp project %s\n", projectID)
				} else {
					if err := gcloudCreateProject(projectID, project.Name, 1); err != nil {
						return err
					}
					fmt.Printf("control-plane: created gcp project %s\n", projectID)
				}
			}

			if err := gcloudSetActiveProject(projectID); err != nil {
				return err
			}
			if len(apis) > 0 {
				if err := gcloudEnableServices(projectID, apis); err != nil {
					return err
				}
			}
		}

		if err := reconcileSecretSet(secretReconcileOptions{
			OwnerLabel:        "project " + project.Name,
			Secrets:           project.Secrets,
			FallbackProjectID: projectID,
			DryRun:            dryRun,
			Prune:             prune,
		}); err != nil {
			return err
		}
		if projectID != "" {
			if err := reconcileDeploymentState(project, projectID, dryRun, prune); err != nil {
				return err
			}
		} else if len(project.Deployments) > 0 {
			fmt.Printf("control-plane: skip deployments for %s (no gcp.project_id)\n", project.Name)
		}
		if len(project.Domains) > 0 {
			if err := reconcileDomainSet(project.Name, project.Domains, dryRun, prune); err != nil {
				return err
			}
		}
	}
	return nil
}

func gcloudDeleteProject(projectID string) error {
	if err := runCommandStreaming("gcloud", "projects", "delete", projectID, "--quiet"); err != nil {
		return fmt.Errorf("control-plane: failed to delete project %s: %w", projectID, err)
	}
	return nil
}

func reconcileDeploymentState(project controlPlaneProject, projectID string, dryRun bool, prune bool) error {
	for _, deployment := range project.Deployments {
		provider := strings.TrimSpace(deployment.Provider)
		if provider == "" {
			provider = "gcp-cloud-run"
		}
		state := strings.TrimSpace(deployment.DesiredState)
		if state == "" {
			state = "present"
		}
		if provider != "gcp-cloud-run" {
			fmt.Printf("control-plane: skip deployment %s (unsupported provider %s)\n", deployment.Name, provider)
			continue
		}

		service := strings.TrimSpace(deployment.Service)
		region := strings.TrimSpace(deployment.Region)
		if service == "" {
			fmt.Printf("control-plane: skip deployment %s (missing service)\n", deployment.Name)
			continue
		}
		if region == "" {
			region = "us-central1"
		}

		if state == "absent" {
			if !prune {
				fmt.Printf("control-plane: skip deletion for deployment %s (use --prune or destroy)\n", deployment.Name)
				continue
			}
			if dryRun {
				fmt.Printf("control-plane: would delete Cloud Run service %s (%s)\n", service, region)
				continue
			}
			if err := runCommandStreaming("gcloud", "run", "services", "delete", service, "--region", region, "--project", projectID, "--quiet"); err != nil {
				return fmt.Errorf("control-plane: failed deleting deployment %s: %w", deployment.Name, err)
			}
			fmt.Printf("control-plane: deleted Cloud Run service %s (%s)\n", service, region)
			continue
		}

		fmt.Printf("control-plane: declared deployment %s (%s/%s)\n", deployment.Name, provider, service)
	}
	return nil
}

type secretReconcileOptions struct {
	OwnerLabel        string
	Secrets           []secretSpec
	FallbackProjectID string
	DryRun            bool
	Prune             bool
	SelectedProject   string
	AccountLevel      bool
}

func reconcileSecretSet(opts secretReconcileOptions) error {
	ownerLabel := strings.TrimSpace(opts.OwnerLabel)
	if ownerLabel == "" {
		ownerLabel = "scope"
	}
	for _, secret := range opts.Secrets {
		state := strings.TrimSpace(secret.DesiredState)
		if state == "" {
			state = "present"
		}
		name := strings.TrimSpace(secret.Name)
		if name == "" {
			fmt.Printf("control-plane: skip secret with empty name in %s\n", ownerLabel)
			continue
		}

		shares := normalizeSecretShares(secret, opts.FallbackProjectID)
		if opts.AccountLevel && strings.TrimSpace(opts.SelectedProject) != "" {
			shares = filterSecretSharesForProject(shares, opts.SelectedProject)
			if len(shares) == 0 {
				fmt.Printf("control-plane: skip secret %s in %s (no shares for project %s)\n", name, ownerLabel, opts.SelectedProject)
				continue
			}
		}
		if len(shares) == 0 {
			fmt.Printf("control-plane: skip secret %s (no share targets)\n", name)
			continue
		}

		sourceEnv := strings.TrimSpace(secret.SourceEnv)
		sourceValue := strings.TrimSpace(os.Getenv(sourceEnv))
		hasSourceValue := sourceEnv != "" && sourceValue != ""

		for _, share := range shares {
			targetPlatform := normalizeSecretPlatform(share.Platform)
			targetName := strings.TrimSpace(share.Name)
			if targetName == "" {
				targetName = name
			}

			switch targetPlatform {
			case "gcp":
				targetProjectID := strings.TrimSpace(share.ProjectID)
				if targetProjectID == "" {
					return fmt.Errorf("control-plane: secret %s has gcp share without project_id", name)
				}

				if state == "absent" {
					if !opts.Prune {
						fmt.Printf("control-plane: skip deletion for secret %s (use --prune or destroy)\n", targetName)
						continue
					}
					if opts.DryRun {
						fmt.Printf("control-plane: would delete gcp secret %s in project %s\n", targetName, targetProjectID)
						continue
					}
					if err := gcloudDeleteSecret(targetProjectID, targetName); err != nil {
						return err
					}
					fmt.Printf("control-plane: deleted gcp secret %s in project %s\n", targetName, targetProjectID)
					continue
				}

				if opts.DryRun {
					fmt.Printf("control-plane: would ensure gcp secret exists %s (project %s)\n", targetName, targetProjectID)
					if sourceEnv != "" {
						if hasSourceValue {
							fmt.Printf("control-plane: would add gcp secret version for %s from env %s (project %s)\n", targetName, sourceEnv, targetProjectID)
						} else {
							fmt.Printf("control-plane: source env %s not set; would skip version add for %s (project %s)\n", sourceEnv, targetName, targetProjectID)
						}
					}
					continue
				}

				exists, err := gcloudSecretExists(targetProjectID, targetName)
				if err != nil {
					return err
				}
				if !exists {
					if err := gcloudCreateSecret(targetProjectID, targetName); err != nil {
						return err
					}
					fmt.Printf("control-plane: created gcp secret %s in project %s\n", targetName, targetProjectID)
				}

				if sourceEnv == "" {
					continue
				}
				if !hasSourceValue {
					fmt.Printf("control-plane: source env %s not set; skip version add for %s (project %s)\n", sourceEnv, targetName, targetProjectID)
					continue
				}
				if err := gcloudAddSecretVersion(targetProjectID, targetName, sourceValue); err != nil {
					return err
				}
				fmt.Printf("control-plane: added gcp secret version for %s from %s (project %s)\n", targetName, sourceEnv, targetProjectID)

			case "vercel":
				targetProjectID := strings.TrimSpace(share.ProjectID)
				if targetProjectID == "" {
					return fmt.Errorf("control-plane: secret %s has vercel share without project_id", name)
				}
				environments := normalizeVercelEnvironments(share.Environments)

				if state == "absent" {
					if !opts.Prune {
						fmt.Printf("control-plane: skip deletion for vercel secret %s (use --prune or destroy)\n", targetName)
						continue
					}
					for _, environment := range environments {
						if opts.DryRun {
							fmt.Printf("control-plane: would delete vercel env %s (%s, project %s)\n", targetName, environment, targetProjectID)
							continue
						}
						if err := vercelDeleteEnv(targetProjectID, targetName, environment); err != nil {
							return err
						}
						fmt.Printf("control-plane: deleted vercel env %s (%s, project %s)\n", targetName, environment, targetProjectID)
					}
					continue
				}

				for _, environment := range environments {
					if opts.DryRun {
						fmt.Printf("control-plane: would ensure vercel env %s (%s, project %s)\n", targetName, environment, targetProjectID)
						if sourceEnv != "" {
							if hasSourceValue {
								fmt.Printf("control-plane: would upsert vercel env %s from %s (%s, project %s)\n", targetName, sourceEnv, environment, targetProjectID)
							} else {
								fmt.Printf("control-plane: source env %s not set; would skip vercel upsert for %s (%s, project %s)\n", sourceEnv, targetName, environment, targetProjectID)
							}
						}
						continue
					}

					if sourceEnv == "" {
						continue
					}
					if !hasSourceValue {
						fmt.Printf("control-plane: source env %s not set; skip vercel upsert for %s (%s, project %s)\n", sourceEnv, targetName, environment, targetProjectID)
						continue
					}
					if err := vercelUpsertEnv(targetProjectID, targetName, environment, sourceValue); err != nil {
						return err
					}
					fmt.Printf("control-plane: upserted vercel env %s (%s, project %s)\n", targetName, environment, targetProjectID)
				}

			default:
				fmt.Printf("control-plane: skip secret %s share target (unsupported platform %s)\n", targetName, targetPlatform)
				continue
			}
		}
	}
	return nil
}

func projectNeedsGCP(project controlPlaneProject, fallbackProjectID string) bool {
	if strings.TrimSpace(fallbackProjectID) != "" {
		return true
	}
	for _, deployment := range project.Deployments {
		provider := strings.TrimSpace(deployment.Provider)
		if provider == "" || provider == "gcp-cloud-run" {
			return true
		}
	}
	return secretSetNeedsGCP(project.Secrets, fallbackProjectID, "", false)
}

func projectNeedsVercel(project controlPlaneProject) bool {
	return secretSetNeedsVercel(project.Secrets, "", "", false)
}

func secretSetNeedsGCP(secrets []secretSpec, fallbackProjectID string, selectedProject string, accountLevel bool) bool {
	for _, secret := range secrets {
		shares := normalizeSecretShares(secret, fallbackProjectID)
		if accountLevel && strings.TrimSpace(selectedProject) != "" {
			shares = filterSecretSharesForProject(shares, selectedProject)
		}
		for _, share := range shares {
			if normalizeSecretPlatform(share.Platform) == "gcp" {
				return true
			}
		}
	}
	return false
}

func secretSetNeedsVercel(secrets []secretSpec, fallbackProjectID string, selectedProject string, accountLevel bool) bool {
	for _, secret := range secrets {
		shares := normalizeSecretShares(secret, fallbackProjectID)
		if accountLevel && strings.TrimSpace(selectedProject) != "" {
			shares = filterSecretSharesForProject(shares, selectedProject)
		}
		for _, share := range shares {
			if normalizeSecretPlatform(share.Platform) == "vercel" {
				return true
			}
		}
	}
	return false
}

func commandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func normalizeSecretShares(secret secretSpec, fallbackProjectID string) []secretShareSpec {
	if len(secret.Shares) == 0 {
		platform := normalizeSecretPlatform(secret.Provider)
		if platform == "" {
			platform = "gcp"
		}
		out := secretShareSpec{
			Platform:   platform,
			ProjectID:  strings.TrimSpace(fallbackProjectID),
			Name:       strings.TrimSpace(secret.Name),
			TargetType: "project",
		}
		if platform == "vercel" {
			out.Environments = normalizeVercelEnvironments(nil)
		}
		return []secretShareSpec{out}
	}

	out := make([]secretShareSpec, 0, len(secret.Shares))
	for _, share := range secret.Shares {
		platform := normalizeSecretPlatform(share.Platform)
		if platform == "" {
			platform = normalizeSecretPlatform(secret.Provider)
		}
		if platform == "" {
			platform = "gcp"
		}
		target := secretShareSpec{
			Platform:   platform,
			ProjectID:  strings.TrimSpace(share.ProjectID),
			Name:       strings.TrimSpace(share.Name),
			TargetType: normalizeSecretTargetType(share.TargetType),
			TargetID:   strings.TrimSpace(share.TargetID),
			Project:    strings.TrimSpace(share.Project),
		}
		if target.ProjectID == "" {
			target.ProjectID = strings.TrimSpace(fallbackProjectID)
		}
		if target.Name == "" {
			target.Name = strings.TrimSpace(secret.Name)
		}
		if target.TargetType == "" {
			target.TargetType = "project"
		}
		if platform == "vercel" {
			target.Environments = normalizeVercelEnvironments(share.Environments)
		}
		out = append(out, target)
	}
	return out
}

func normalizeSecretPlatform(platform string) string {
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case "":
		return ""
	case "gcp", "gcp-secret-manager", "google-secret-manager":
		return "gcp"
	case "vercel", "vercel-env", "vercel-project-env":
		return "vercel"
	default:
		return strings.ToLower(strings.TrimSpace(platform))
	}
}

func normalizeSecretTargetType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return ""
	case "project", "projects":
		return "project"
	case "application", "app", "apps":
		return "application"
	case "deployment", "deploy", "service":
		return "deployment"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func filterSecretSharesForProject(shares []secretShareSpec, selectedProject string) []secretShareSpec {
	project := strings.TrimSpace(selectedProject)
	if project == "" {
		return shares
	}
	filtered := make([]secretShareSpec, 0, len(shares))
	for _, share := range shares {
		if secretShareMatchesProject(share, project) {
			filtered = append(filtered, share)
		}
	}
	return filtered
}

func secretShareMatchesProject(share secretShareSpec, selectedProject string) bool {
	project := strings.TrimSpace(selectedProject)
	if project == "" {
		return true
	}
	shareProject := strings.TrimSpace(share.Project)
	if shareProject != "" {
		return shareProject == project
	}
	targetType := normalizeSecretTargetType(share.TargetType)
	targetID := strings.TrimSpace(share.TargetID)
	if targetType == "project" && targetID != "" {
		return targetID == project
	}
	return true
}

func normalizeVercelEnvironments(values []string) []string {
	if len(values) == 0 {
		return []string{"development", "preview", "production"}
	}
	seen := map[string]struct{}{}
	out := []string{}
	for _, value := range values {
		env := strings.ToLower(strings.TrimSpace(value))
		if env == "" {
			continue
		}
		switch env {
		case "dev":
			env = "development"
		case "prod":
			env = "production"
		}
		if _, ok := seen[env]; ok {
			continue
		}
		seen[env] = struct{}{}
		out = append(out, env)
	}
	if len(out) == 0 {
		return []string{"development", "preview", "production"}
	}
	return out
}

func gcloudSecretExists(projectID, name string) (bool, error) {
	out, err := runCommandOutput("gcloud", "secrets", "describe", name, "--project", projectID, "--format=value(name)")
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "not found") || strings.Contains(msg, "was not found") {
			return false, nil
		}
		return false, fmt.Errorf("control-plane: failed checking secret %s: %w", name, err)
	}
	return strings.TrimSpace(out) != "", nil
}

func gcloudCreateSecret(projectID, name string) error {
	if err := runCommandStreaming("gcloud", "secrets", "create", name, "--replication-policy=automatic", "--project", projectID); err != nil {
		return fmt.Errorf("control-plane: failed creating secret %s: %w", name, err)
	}
	return nil
}

func gcloudDeleteSecret(projectID, name string) error {
	if err := runCommandStreaming("gcloud", "secrets", "delete", name, "--project", projectID, "--quiet"); err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "not found") || strings.Contains(msg, "was not found") {
			return nil
		}
		return fmt.Errorf("control-plane: failed deleting secret %s: %w", name, err)
	}
	return nil
}

func gcloudAddSecretVersion(projectID, name, value string) error {
	cmd := exec.Command("gcloud", "secrets", "versions", "add", name, "--data-file=-", "--project", projectID)
	cmd.Stdin = strings.NewReader(value)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("control-plane: failed adding secret version for %s: %w", name, err)
	}
	return nil
}

func vercelUpsertEnv(projectID, name, environment, value string) error {
	_ = vercelDeleteEnv(projectID, name, environment)
	cmd := exec.Command("vercel", "env", "add", name, environment, "--project", projectID)
	cmd.Stdin = strings.NewReader(value + "\n")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("control-plane: failed upserting vercel env %s (%s, project %s): %w", name, environment, projectID, err)
	}
	return nil
}

func vercelDeleteEnv(projectID, name, environment string) error {
	if err := runCommandStreaming("vercel", "env", "rm", name, environment, "--project", projectID, "--yes"); err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "not found") || strings.Contains(msg, "does not exist") || strings.Contains(msg, "no such") {
			return nil
		}
		return fmt.Errorf("control-plane: failed deleting vercel env %s (%s, project %s): %w", name, environment, projectID, err)
	}
	return nil
}
