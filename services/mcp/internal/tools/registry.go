package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Tool represents a tool definition.
type Tool struct {
	Name            string
	Description     string
	InputSchemaJSON string
}

// Registry holds available tools and a reference to the repo root.
type Registry struct {
	root string
}

// NewRegistry creates a new Registry with the given repo root.
func NewRegistry(root string) *Registry {
	return &Registry{root: root}
}

// controlPlane mirrors the structure of platform.controlplane.json.
type controlPlane struct {
	Projects []struct {
		Name string `json:"name"`
		GCP  *struct {
			ProjectID string `json:"project_id"`
			Region    string `json:"region"`
		} `json:"gcp"`
		Deployments []struct {
			Service  string `json:"service"`
			Provider string `json:"provider"`
		} `json:"deployments"`
		Domains []struct {
			Provider string `json:"provider"`
			Name     string `json:"name"`
		} `json:"domains"`
	} `json:"projects"`
}

func (r *Registry) loadControlPlane() (*controlPlane, error) {
	path := filepath.Join(r.root, "platform.controlplane.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loadControlPlane: %w", err)
	}
	var cp controlPlane
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, fmt.Errorf("loadControlPlane: %w", err)
	}
	return &cp, nil
}

func runCmd(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("%s %v: timed out after 30s", name, args)
		}
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return "", err
		}
		return "", fmt.Errorf("%s %v: %s", name, args, msg)
	}
	return strings.TrimSpace(string(out)), nil
}

// Tools returns the list of all available tool definitions.
func (r *Registry) Tools() []Tool {
	return []Tool{
		{
			Name:        "gcp_configured_projects",
			Description: "List all GCP-backed projects configured in platform.controlplane.json",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {},
  "required": []
}`,
		},
		{
			Name:        "gcp_recent_logs",
			Description: "Fetch recent GCP Cloud Logging entries for a project",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "project_id": {"type": "string", "description": "GCP project ID"},
    "filter": {"type": "string", "description": "Log filter expression (optional)"},
    "limit": {"type": "number", "description": "Maximum number of log entries (default 50)"}
  },
  "required": ["project_id"]
}`,
		},
		{
			Name:        "gcp_run_services",
			Description: "List Cloud Run services in a GCP project and region",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "project_id": {"type": "string", "description": "GCP project ID"},
    "region": {"type": "string", "description": "GCP region (e.g. us-central1)"}
  },
  "required": ["project_id", "region"]
}`,
		},
		{
			Name:        "gcp_active_account",
			Description: "Get the currently active gcloud account",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {},
  "required": []
}`,
		},
		{
			Name:        "vercel_configured_projects",
			Description: "List all Vercel-backed projects configured in platform.controlplane.json",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {},
  "required": []
}`,
		},
		{
			Name:        "vercel_recent_deployments",
			Description: "List recent Vercel deployments for a project",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "project_id": {"type": "string", "description": "Vercel project ID or name"}
  },
  "required": ["project_id"]
}`,
		},
		{
			Name:        "vercel_inspect_url",
			Description: "Inspect a Vercel deployment URL",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "url": {"type": "string", "description": "Vercel deployment URL to inspect"}
  },
  "required": ["url"]
}`,
		},
		{
			Name:        "vercel_domains",
			Description: "List all Vercel domains",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {},
  "required": []
}`,
		},
	}
}

// Call invokes a tool by name with the given JSON arguments string.
// Returns (result, isError, error).
func (r *Registry) Call(name, argumentsJSON string) (string, bool, error) {
	var args map[string]interface{}
	if argumentsJSON != "" {
		if err := json.Unmarshal([]byte(argumentsJSON), &args); err != nil {
			return "", true, fmt.Errorf("failed to parse arguments: %w", err)
		}
	}
	if args == nil {
		args = map[string]interface{}{}
	}

	getString := func(key string) string {
		v, _ := args[key].(string)
		return v
	}
	getInt := func(key string, def int) int {
		v, ok := args[key]
		if !ok {
			return def
		}
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
		return def
	}

	switch name {
	case "gcp_configured_projects":
		cp, err := r.loadControlPlane()
		if err != nil {
			return err.Error(), true, nil
		}
		var lines []string
		for _, p := range cp.Projects {
			if p.GCP != nil {
				lines = append(lines, fmt.Sprintf("name=%s project_id=%s region=%s", p.Name, p.GCP.ProjectID, p.GCP.Region))
			}
		}
		if len(lines) == 0 {
			return "No GCP-backed projects found.", false, nil
		}
		return strings.Join(lines, "\n"), false, nil

	case "gcp_recent_logs":
		projectID := getString("project_id")
		if projectID == "" {
			return "project_id is required", true, nil
		}
		limit := getInt("limit", 50)
		filter := getString("filter")
		cmdArgs := []string{
			"logging", "read",
			"--project", projectID,
			fmt.Sprintf("--limit=%d", limit),
			"--format=json",
		}
		if filter != "" {
			cmdArgs = append(cmdArgs, filter)
		}
		out, err := runCmd("gcloud", cmdArgs...)
		if err != nil {
			return err.Error(), true, nil
		}
		return out, false, nil

	case "gcp_run_services":
		projectID := getString("project_id")
		region := getString("region")
		if projectID == "" {
			return "project_id is required", true, nil
		}
		if region == "" {
			return "region is required", true, nil
		}
		out, err := runCmd("gcloud", "run", "services", "list",
			"--project", projectID,
			"--region", region,
			"--format=json",
		)
		if err != nil {
			return err.Error(), true, nil
		}
		return out, false, nil

	case "gcp_active_account":
		out, err := runCmd("gcloud", "auth", "list",
			"--filter=status:ACTIVE",
			"--format=value(account)",
		)
		if err != nil {
			return err.Error(), true, nil
		}
		return out, false, nil

	case "vercel_configured_projects":
		cp, err := r.loadControlPlane()
		if err != nil {
			return err.Error(), true, nil
		}
		var lines []string
		for _, p := range cp.Projects {
			for _, d := range p.Deployments {
				if strings.EqualFold(d.Provider, "vercel") {
					lines = append(lines, fmt.Sprintf("name=%s service=%s", p.Name, d.Service))
				}
			}
		}
		if len(lines) == 0 {
			return "No Vercel-backed projects found.", false, nil
		}
		return strings.Join(lines, "\n"), false, nil

	case "vercel_recent_deployments":
		projectID := getString("project_id")
		if projectID == "" {
			return "project_id is required", true, nil
		}
		out, err := runCmd("vercel", "list", projectID)
		if err != nil {
			return err.Error(), true, nil
		}
		return out, false, nil

	case "vercel_inspect_url":
		url := getString("url")
		if url == "" {
			return "url is required", true, nil
		}
		out, err := runCmd("vercel", "inspect", url)
		if err != nil {
			return err.Error(), true, nil
		}
		return out, false, nil

	case "vercel_domains":
		out, err := runCmd("vercel", "domains", "ls")
		if err != nil {
			return err.Error(), true, nil
		}
		return out, false, nil

	default:
		return fmt.Sprintf("unknown tool: %s", name), true, nil
	}
}
