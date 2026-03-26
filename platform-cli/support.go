package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const projectConfigFileName = "platform.projects.json"

type projectConfig struct {
	Version  string        `json:"version"`
	Projects []projectSpec `json:"projects"`
}

type projectSpec struct {
	Name    string         `json:"name"`
	Targets projectTargets `json:"targets,omitempty"`
}

type projectTargets struct {
	GCP []projectGCPTarget `json:"gcp,omitempty"`
}

type projectGCPTarget struct {
	ProjectID string   `json:"project_id"`
	Region    string   `json:"region,omitempty"`
	Services  []string `json:"services,omitempty"`
}

func findRepoRoot() (string, error) {
	if wd, err := os.Getwd(); err == nil {
		if root := findRepoRootFrom(wd); root != "" {
			return root, nil
		}
	}

	if exe, err := os.Executable(); err == nil {
		if root := findRepoRootFrom(filepath.Dir(exe)); root != "" {
			return root, nil
		}
	}

	return "", errors.New("repo root not found")
}

func findRepoRootFrom(start string) string {
	dir := start
	for {
		for _, marker := range []string{"nx.json", projectConfigFileName, controlPlaneConfigFileName} {
			if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func loadProjectConfig(required bool) (*projectConfig, string, error) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return nil, "", err
	}

	path := filepath.Join(repoRoot, projectConfigFileName)
	body, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && !required {
			return nil, path, nil
		}
		if errors.Is(err, os.ErrNotExist) {
			return nil, "", fmt.Errorf("%s not found. Materialize configs with python3 scripts/ci/materialize_platform_configs.py", projectConfigFileName)
		}
		return nil, "", err
	}

	var cfg projectConfig
	if err := json.Unmarshal(body, &cfg); err != nil {
		return nil, "", fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return &cfg, path, nil
}

func normalizeGCPServiceAPIs(values []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, value := range values {
		api := strings.TrimSpace(value)
		if api == "" {
			continue
		}
		if !strings.Contains(api, ".") {
			api += ".googleapis.com"
		}
		if _, ok := seen[api]; ok {
			continue
		}
		seen[api] = struct{}{}
		out = append(out, api)
	}
	return out
}

func runCommandStreaming(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCommandOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return "", err
		}
		return "", fmt.Errorf("%w: %s", err, msg)
	}
	return strings.TrimSpace(string(out)), nil
}

func gcloudProjectExists(projectID string) (bool, error) {
	out, err := runCommandOutput("gcloud", "projects", "describe", projectID, "--format=value(projectId)")
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "not found") || strings.Contains(msg, "does not exist") {
			return false, nil
		}
		return false, fmt.Errorf("control-plane: failed checking gcp project %s: %w", projectID, err)
	}
	return strings.TrimSpace(out) != "", nil
}

func gcloudCreateProject(projectID, name string, _ int) error {
	args := []string{"projects", "create", projectID}
	if label := strings.TrimSpace(name); label != "" {
		args = append(args, "--name", label)
	}
	if err := runCommandStreaming("gcloud", args...); err != nil {
		return fmt.Errorf("control-plane: failed creating gcp project %s: %w", projectID, err)
	}
	return nil
}

func gcloudSetActiveProject(projectID string) error {
	if err := runCommandStreaming("gcloud", "config", "set", "project", projectID); err != nil {
		return fmt.Errorf("control-plane: failed linking active gcloud project %s: %w", projectID, err)
	}
	return nil
}

func gcloudEnableServices(projectID string, apis []string) error {
	normalized := normalizeGCPServiceAPIs(apis)
	if len(normalized) == 0 {
		return nil
	}
	args := append([]string{"services", "enable"}, normalized...)
	args = append(args, "--project", projectID)
	if err := runCommandStreaming("gcloud", args...); err != nil {
		return fmt.Errorf("control-plane: failed enabling APIs for %s: %w", projectID, err)
	}
	return nil
}
