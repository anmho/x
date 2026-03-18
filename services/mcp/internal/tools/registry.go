package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Tool represents a tool definition.
type Tool struct {
	Name            string
	Description     string
	InputSchemaJSON string
}

// Registry holds available tools and a reference to the repo root.
type Registry struct {
	root   string
	collab *CollabStore
}

// NewRegistry creates a new Registry with the given repo root.
func NewRegistry(root string) *Registry {
	return &Registry{root: root, collab: NewCollabStore("")}
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
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s %v: %s", name, args, string(exitErr.Stderr))
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Tools returns the list of all available tool definitions.
func (r *Registry) Tools() []Tool {
	return []Tool{
		{
			Name:        "collab_get_or_create_channel",
			Description: "Create or select a collaboration channel using a stable key, title, participants, and metadata",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "channel_key": {"type": "string", "description": "Stable logical key for the channel, such as a task or thread id"},
    "title": {"type": "string", "description": "Human-readable title"},
    "participants": {"type": "array", "items": {"type": "string"}, "description": "Agent or user ids participating in the channel"},
    "metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional channel metadata"}
  },
  "required": []
}`,
		},
		{
			Name:        "collab_list_channels",
			Description: "List collaboration channels, optionally filtering by participant or query",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "participant": {"type": "string", "description": "Return only channels containing this participant"},
    "query": {"type": "string", "description": "Filter by channel id, key, title, or participant"},
    "limit": {"type": "number", "description": "Maximum number of channels to return (default 50)"}
  },
  "required": []
}`,
		},
		{
			Name:        "collab_find_channels_by_agent",
			Description: "Find collaboration channels for a specific agent participant",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "agent_id": {"type": "string", "description": "Agent identifier to match"},
    "limit": {"type": "number", "description": "Maximum number of channels to return (default 50)"}
  },
  "required": ["agent_id"]
}`,
		},
		{
			Name:        "collab_post_message",
			Description: "Post a message into a collaboration channel",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "channel_id": {"type": "string", "description": "Target channel id"},
    "sender": {"type": "string", "description": "Agent or user id sending the message"},
    "kind": {"type": "string", "description": "Message kind such as message, status, handoff, or artifact"},
    "body": {"type": "string", "description": "Message body"},
    "metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional structured metadata"}
  },
  "required": ["channel_id", "sender", "body"]
}`,
		},
		{
			Name:        "collab_read_messages",
			Description: "Read recent messages from a collaboration channel",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "channel_id": {"type": "string", "description": "Target channel id"},
    "limit": {"type": "number", "description": "Maximum number of messages to return (default 50)"}
  },
  "required": ["channel_id"]
}`,
		},
		{
			Name:        "mail_find_channels",
			Description: "List mailbox channels, optionally filtering by participant or query",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "participant": {"type": "string", "description": "Return only channels containing this participant"},
    "query": {"type": "string", "description": "Filter by channel id, key, title, or participant"},
    "limit": {"type": "number", "description": "Maximum number of channels to return (default 50)"}
  },
  "required": []
}`,
		},
		{
			Name:        "mail_get_channel_for_agent",
			Description: "Get or create the stable mailbox channel for a specific agent",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "agent_id": {"type": "string", "description": "Agent identifier to match"}
  },
  "required": ["agent_id"]
}`,
		},
		{
			Name:        "mail_send",
			Description: "Send a mailbox message into a channel",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "channel_id": {"type": "string", "description": "Target channel id"},
    "sender": {"type": "string", "description": "Agent or user id sending the message"},
    "kind": {"type": "string", "description": "Message kind such as message, status, handoff, or artifact"},
    "body": {"type": "string", "description": "Message body"},
    "metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional structured metadata"}
  },
  "required": ["channel_id", "sender", "body"]
}`,
		},
		{
			Name:        "mail_read",
			Description: "Read ordered mailbox messages from a channel with replay support",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "channel_id": {"type": "string", "description": "Target channel id"},
    "after_sequence": {"type": "number", "description": "Only return messages after this sequence"},
    "limit": {"type": "number", "description": "Maximum number of messages to return (default 50)"}
  },
  "required": ["channel_id"]
}`,
		},
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
	getStringMap := func(key string) map[string]string {
		raw, ok := args[key]
		if !ok || raw == nil {
			return nil
		}
		obj, ok := raw.(map[string]interface{})
		if !ok {
			return nil
		}
		out := make(map[string]string, len(obj))
		for k, v := range obj {
			if s, ok := v.(string); ok {
				out[k] = s
			}
		}
		return out
	}
	getStringSlice := func(key string) []string {
		raw, ok := args[key]
		if !ok || raw == nil {
			return nil
		}
		items, ok := raw.([]interface{})
		if !ok {
			return nil
		}
		out := make([]string, 0, len(items))
		for _, item := range items {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
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
	getInt64 := func(key string, def int64) int64 {
		v, ok := args[key]
		if !ok {
			return def
		}
		switch n := v.(type) {
		case float64:
			return int64(n)
		case int:
			return int64(n)
		case int64:
			return n
		}
		return def
	}

	switch name {
	case "collab_get_or_create_channel":
		channel, err := r.collab.GetOrCreateChannel(
			getString("channel_key"),
			getString("title"),
			getStringSlice("participants"),
			getStringMap("metadata"),
		)
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channel), false, nil

	case "collab_list_channels":
		channels, err := r.collab.ListChannels(
			getString("participant"),
			getString("query"),
			getInt("limit", 50),
		)
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channels), false, nil

	case "collab_find_channels_by_agent":
		agentID := getString("agent_id")
		if agentID == "" {
			return "agent_id is required", true, nil
		}
		channels, err := r.collab.ListChannels(agentID, "", getInt("limit", 50))
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channels), false, nil

	case "mail_find_channels":
		channels, err := r.collab.ListChannels(
			getString("participant"),
			getString("query"),
			getInt("limit", 50),
		)
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channels), false, nil

	case "mail_get_channel_for_agent":
		agentID := getString("agent_id")
		if agentID == "" {
			return "agent_id is required", true, nil
		}
		channel, err := r.collab.GetChannelForAgent(agentID)
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channel), false, nil

	case "collab_post_message", "mail_send":
		message, err := r.collab.PostMessage(
			getString("channel_id"),
			getString("sender"),
			getString("kind"),
			getString("body"),
			getStringMap("metadata"),
		)
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(message), false, nil

	case "collab_read_messages", "mail_read":
		messages, err := r.collab.ReadMessages(
			getString("channel_id"),
			getInt64("after_sequence", 0),
			getInt("limit", 50),
		)
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(messages), false, nil

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

func mustJSON(v any) string {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}
	return string(raw)
}
