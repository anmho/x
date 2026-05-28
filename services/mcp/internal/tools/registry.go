package tools

import (
	"context"
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
	linear *LinearClient
	slack  *SlackBridge
	proxy  *ProviderProxyClient
}

// NewRegistry creates a new Registry with the given repo root.
func NewRegistry(root string) *Registry {
	if abs, err := filepath.Abs(root); err == nil {
		root = abs
	}
	return &Registry{
		root:   root,
		collab: NewCollabStore(""),
		proxy:  NewProviderProxyClientFromEnv(),
	}
}

func (r *Registry) CollabStore() *CollabStore {
	return r.collab
}

func (r *Registry) SetLinearClient(client *LinearClient) {
	r.linear = client
}

func (r *Registry) SetSlackBridge(bridge *SlackBridge) {
	r.slack = bridge
}

func (r *Registry) SetProviderProxyClient(client *ProviderProxyClient) {
	r.proxy = client
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
	tools := []Tool{
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
			Name:        "collab_get_or_create_linear_channel",
			Description: "Resolve or create a collaboration channel from a Linear issue identifier",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "issue_id": {"type": "string", "description": "Linear issue identifier such as ANM-190"},
    "participants": {"type": "array", "items": {"type": "string"}, "description": "Agent or user ids participating in the channel"},
    "metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional channel metadata overrides"}
  },
  "required": ["issue_id"]
}`,
		},
		{
			Name:        "collab_get_channel_by_external_ref",
			Description: "Resolve a collaboration channel by an external reference such as a Slack thread or Linear issue",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "source": {"type": "string", "description": "External source name such as slack or linear"},
    "external_id": {"type": "string", "description": "Source-specific external identifier"},
    "parent_id": {"type": "string", "description": "Optional parent identifier such as Slack channel id"}
  },
  "required": ["source", "external_id"]
}`,
		},
		{
			Name:        "collab_get_or_create_external_channel",
			Description: "Resolve or create a collaboration channel keyed to an external provider reference",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "source": {"type": "string", "description": "External source such as github, google_docs, google_search, perplexity, glean, yahoo, yahoo_finance, or coinbase"},
    "external_id": {"type": "string", "description": "Stable provider-specific identifier"},
    "parent_id": {"type": "string", "description": "Optional parent identifier such as a repository, result type, or grouping id"},
    "title": {"type": "string", "description": "Human-readable channel title"},
    "url": {"type": "string", "description": "Canonical provider URL"},
    "participants": {"type": "array", "items": {"type": "string"}, "description": "Agent or user ids participating in the channel"},
    "metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional channel metadata"},
    "ref_metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional provider-specific external-ref metadata"}
  },
  "required": ["source", "external_id"]
}`,
		},
		{
			Name:        "collab_get_or_create_github_channel",
			Description: "Resolve or create a collaboration channel for a GitHub issue, pull request, or discussion",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "repository": {"type": "string", "description": "Repository in owner/name form"},
    "number": {"type": "string", "description": "Issue, pull request, or discussion number"},
    "kind": {"type": "string", "description": "github object kind such as issue, pull_request, or discussion"},
    "title": {"type": "string", "description": "Optional GitHub title"},
    "url": {"type": "string", "description": "Optional GitHub URL"},
    "participants": {"type": "array", "items": {"type": "string"}, "description": "Agent or user ids participating in the channel"},
    "metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional channel metadata"}
  },
  "required": ["repository", "number"]
}`,
		},
		{
			Name:        "collab_link_slack_thread",
			Description: "Attach a Slack channel/thread reference to an existing collaboration channel",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "channel_id": {"type": "string", "description": "Target collaboration channel id"},
    "slack_channel_id": {"type": "string", "description": "Slack channel id"},
    "slack_thread_ts": {"type": "string", "description": "Slack thread ts"},
    "title": {"type": "string", "description": "Optional Slack thread title"},
    "url": {"type": "string", "description": "Optional Slack permalink"}
  },
  "required": ["channel_id", "slack_channel_id", "slack_thread_ts"]
}`,
		},
		{
			Name:        "collab_link_external_ref",
			Description: "Attach an arbitrary external provider reference to an existing collaboration channel",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "channel_id": {"type": "string", "description": "Target collaboration channel id"},
    "source": {"type": "string", "description": "External source such as github, google_docs, google_search, perplexity, glean, yahoo, yahoo_finance, or coinbase"},
    "external_id": {"type": "string", "description": "Stable provider-specific identifier"},
    "parent_id": {"type": "string", "description": "Optional parent identifier"},
    "title": {"type": "string", "description": "Optional external reference title"},
    "url": {"type": "string", "description": "Optional canonical URL"},
    "metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional provider-specific external-ref metadata"}
  },
  "required": ["channel_id", "source", "external_id"]
}`,
		},
		{
			Name:        "collab_link_google_doc",
			Description: "Attach a Google Docs document reference to an existing collaboration channel",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "channel_id": {"type": "string", "description": "Target collaboration channel id"},
    "doc_id": {"type": "string", "description": "Google Docs document id"},
    "title": {"type": "string", "description": "Optional document title"},
    "url": {"type": "string", "description": "Optional Google Docs URL"},
    "metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional document metadata"}
  },
  "required": ["channel_id", "doc_id"]
}`,
		},
		{
			Name:        "collab_link_glean_result",
			Description: "Attach a Glean result or document reference to an existing collaboration channel",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "channel_id": {"type": "string", "description": "Target collaboration channel id"},
    "result_id": {"type": "string", "description": "Stable Glean result or document id"},
    "title": {"type": "string", "description": "Optional result title"},
    "url": {"type": "string", "description": "Optional result URL"},
    "datasource": {"type": "string", "description": "Optional Glean datasource"},
    "metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional result metadata"}
  },
  "required": ["channel_id", "result_id"]
}`,
		},
		{
			Name:        "collab_link_google_search_result",
			Description: "Attach a Google Search result reference to an existing collaboration channel",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "channel_id": {"type": "string", "description": "Target collaboration channel id"},
    "result_id": {"type": "string", "description": "Stable result identifier; defaults to url when omitted"},
    "query": {"type": "string", "description": "Search query used to find the result"},
    "title": {"type": "string", "description": "Optional result title"},
    "url": {"type": "string", "description": "Canonical search result URL"},
    "metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional result metadata"}
  },
  "required": ["channel_id", "url"]
}`,
		},
		{
			Name:        "collab_link_perplexity_result",
			Description: "Attach a Perplexity answer or result reference to an existing collaboration channel",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "channel_id": {"type": "string", "description": "Target collaboration channel id"},
    "result_id": {"type": "string", "description": "Stable Perplexity result identifier"},
    "query": {"type": "string", "description": "Original search/query prompt"},
    "title": {"type": "string", "description": "Optional result title"},
    "url": {"type": "string", "description": "Optional result URL"},
    "metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional result metadata"}
  },
  "required": ["channel_id", "result_id"]
}`,
		},
		{
			Name:        "collab_link_yahoo_result",
			Description: "Attach a Yahoo result reference to an existing collaboration channel",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "channel_id": {"type": "string", "description": "Target collaboration channel id"},
    "result_id": {"type": "string", "description": "Stable Yahoo result identifier; defaults to url when omitted"},
    "query": {"type": "string", "description": "Original search query"},
    "title": {"type": "string", "description": "Optional result title"},
    "url": {"type": "string", "description": "Optional result URL"},
    "metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional result metadata"}
  },
  "required": ["channel_id", "url"]
}`,
		},
		{
			Name:        "collab_link_yahoo_finance_symbol",
			Description: "Attach a Yahoo Finance symbol reference to an existing collaboration channel",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "channel_id": {"type": "string", "description": "Target collaboration channel id"},
    "symbol": {"type": "string", "description": "Ticker symbol such as AAPL or BTC-USD"},
    "title": {"type": "string", "description": "Optional company or asset title"},
    "url": {"type": "string", "description": "Optional Yahoo Finance URL"},
    "quote_type": {"type": "string", "description": "Optional quote type such as equity, crypto, or ETF"},
    "metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional symbol metadata"}
  },
  "required": ["channel_id", "symbol"]
}`,
		},
		{
			Name:        "collab_link_coinbase_asset",
			Description: "Attach a Coinbase asset or product reference to an existing collaboration channel",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "channel_id": {"type": "string", "description": "Target collaboration channel id"},
    "asset_id": {"type": "string", "description": "Stable Coinbase asset identifier"},
    "product_id": {"type": "string", "description": "Optional Coinbase product identifier such as BTC-USD"},
    "title": {"type": "string", "description": "Optional asset title"},
    "url": {"type": "string", "description": "Optional Coinbase URL"},
    "metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional asset metadata"}
  },
  "required": ["channel_id", "asset_id"]
}`,
		},
		{
			Name:        "collab_mark_status",
			Description: "Update the status for a collaboration channel and optionally emit a status message",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "channel_id": {"type": "string", "description": "Target collaboration channel id"},
    "status": {"type": "string", "description": "New status: open, blocked, waiting_approval, or done"},
    "sender": {"type": "string", "description": "Sender for an optional status message"},
    "body": {"type": "string", "description": "Optional status message body"},
    "metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional structured metadata"}
  },
  "required": ["channel_id", "status"]
}`,
		},
		{
			Name:        "collab_set_agent_focus",
			Description: "Register or update agent routing metadata used for collaboration dispatch",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "agent_id": {"type": "string", "description": "Stable agent identifier"},
    "aliases": {"type": "array", "items": {"type": "string"}, "description": "Optional aliases for matching"},
    "topics": {"type": "array", "items": {"type": "string"}, "description": "Current focus topics"},
    "capabilities": {"type": "array", "items": {"type": "string"}, "description": "Capabilities used for routing"},
    "active_channel_id": {"type": "string", "description": "Optional current channel id"},
    "metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional routing metadata such as run_id"}
  },
  "required": ["agent_id"]
}`,
		},
		{
			Name:        "collab_list_agents",
			Description: "List registered agent routing metadata, optionally filtered by query",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "query": {"type": "string", "description": "Filter by id, alias, topic, capability, or metadata"},
    "limit": {"type": "number", "description": "Maximum number of agents to return (default 50)"}
  },
  "required": []
}`,
		},
		{
			Name:        "collab_route_message",
			Description: "Route a message to one or more agent mailbox channels using explicit targets or fuzzy matching",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "sender": {"type": "string", "description": "Sender for the routed message"},
    "kind": {"type": "string", "description": "Message kind such as message, status, handoff, or artifact"},
    "body": {"type": "string", "description": "Message body"},
    "channel_id": {"type": "string", "description": "Optional source channel id"},
    "topic": {"type": "string", "description": "Optional routing topic"},
    "query": {"type": "string", "description": "Optional fuzzy query for agent selection"},
    "target_agent_ids": {"type": "array", "items": {"type": "string"}, "description": "Explicit target agent ids"},
    "exclude_agent_ids": {"type": "array", "items": {"type": "string"}, "description": "Agent ids to exclude"},
    "limit": {"type": "number", "description": "Maximum number of deliveries (default 25)"},
    "dry_run": {"type": "boolean", "description": "Preview routing without writing mailbox messages"},
    "delivery_mode": {"type": "string", "description": "Dispatcher hint such as append_context or interrupt_and_replan"},
    "metadata": {"type": "object", "additionalProperties": {"type": "string"}, "description": "Optional metadata attached to delivered mailbox messages"}
  },
  "required": ["sender", "body"]
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
			Name:        "proxy_call_tool",
			Description: "Call a configured upstream MCP tool through a provider-specific proxy endpoint",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "provider": {"type": "string", "description": "Configured provider name such as github, google_docs, google_search, perplexity, glean, yahoo, yahoo_finance, or coinbase"},
    "tool": {"type": "string", "description": "Optional upstream tool override"},
    "arguments": {"type": "object", "description": "Arguments forwarded to the upstream MCP tool", "additionalProperties": true}
  },
  "required": ["provider"]
}`,
		},
		{
			Name:        "github_proxy_call",
			Description: "Call the configured upstream GitHub MCP proxy tool",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "tool": {"type": "string", "description": "Optional upstream tool override"},
    "arguments": {"type": "object", "description": "Arguments forwarded to the upstream tool", "additionalProperties": true}
  },
  "required": []
}`,
		},
		{
			Name:        "google_docs_proxy_call",
			Description: "Call the configured upstream Google Docs MCP proxy tool",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "tool": {"type": "string", "description": "Optional upstream tool override"},
    "arguments": {"type": "object", "description": "Arguments forwarded to the upstream tool", "additionalProperties": true}
  },
  "required": []
}`,
		},
		{
			Name:        "glean_search",
			Description: "Call the configured upstream Glean search MCP tool",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "query": {"type": "string", "description": "Search query"},
    "limit": {"type": "number", "description": "Optional result limit"}
  },
  "required": ["query"]
}`,
		},
		{
			Name:        "perplexity_search_query",
			Description: "Call the configured upstream Perplexity MCP tool",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "query": {"type": "string", "description": "Search or research query"},
    "limit": {"type": "number", "description": "Optional result limit"}
  },
  "required": ["query"]
}`,
		},
		{
			Name:        "google_search_query",
			Description: "Call the configured upstream Google Search MCP tool",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "query": {"type": "string", "description": "Search query"},
    "limit": {"type": "number", "description": "Optional result limit"}
  },
  "required": ["query"]
}`,
		},
		{
			Name:        "yahoo_search_query",
			Description: "Call the configured upstream Yahoo search MCP tool",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "query": {"type": "string", "description": "Search query"},
    "limit": {"type": "number", "description": "Optional result limit"}
  },
  "required": ["query"]
}`,
		},
		{
			Name:        "yahoo_finance_quote",
			Description: "Call the configured upstream Yahoo Finance MCP tool",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "symbol": {"type": "string", "description": "Ticker symbol such as AAPL or BTC-USD"}
  },
  "required": ["symbol"]
}`,
		},
		{
			Name:        "coinbase_lookup",
			Description: "Call the configured upstream Coinbase MCP tool",
			InputSchemaJSON: `{
  "type": "object",
  "properties": {
    "asset_id": {"type": "string", "description": "Asset identifier such as BTC"},
    "product_id": {"type": "string", "description": "Optional product identifier such as BTC-USD"}
  },
  "required": []
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
	return append(tools, workspaceTools()...)
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
	getObject := func(key string) map[string]any {
		raw, ok := args[key]
		if !ok || raw == nil {
			return nil
		}
		obj, ok := raw.(map[string]interface{})
		if !ok {
			return nil
		}
		out := make(map[string]any, len(obj))
		for k, v := range obj {
			out[k] = v
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
	getBool := func(key string, def bool) bool {
		v, ok := args[key]
		if !ok {
			return def
		}
		switch b := v.(type) {
		case bool:
			return b
		case string:
			switch strings.ToLower(strings.TrimSpace(b)) {
			case "true", "1", "yes":
				return true
			case "false", "0", "no":
				return false
			}
		}
		return def
	}
	callProxy := func(provider, tool string, forwarded map[string]any) (string, bool, error) {
		if r.proxy == nil {
			return "provider proxy client is not configured", true, nil
		}
		out, err := r.proxy.CallTool(context.Background(), provider, tool, forwarded)
		if err != nil {
			return err.Error(), true, nil
		}
		return out, false, nil
	}

	switch name {
	case "workspace_read_file", "workspace_list_files", "workspace_search", "workspace_apply_patch", "git_status", "git_diff", "git_checkout":
		return r.callWorkspaceTool(name, args)
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

	case "collab_get_or_create_linear_channel":
		issueID := strings.ToUpper(strings.TrimSpace(getString("issue_id")))
		if issueID == "" {
			return "issue_id is required", true, nil
		}
		issue := &LinearIssue{Identifier: issueID}
		if r.linear != nil && r.linear.Enabled() {
			enriched, err := r.linear.GetIssue(context.Background(), issueID)
			if err != nil {
				return err.Error(), true, nil
			}
			issue = enriched
		}
		channel, err := r.collab.GetOrCreateLinearIssueChannel(
			issue,
			getStringSlice("participants"),
			getStringMap("metadata"),
		)
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channel), false, nil

	case "collab_get_channel_by_external_ref":
		channel, err := r.collab.GetChannelByExternalRef(
			getString("source"),
			getString("external_id"),
			getString("parent_id"),
		)
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channel), false, nil

	case "collab_get_or_create_external_channel":
		channel, err := r.collab.GetOrCreateExternalChannel(
			getString("source"),
			getString("external_id"),
			getString("parent_id"),
			getString("title"),
			getString("url"),
			getStringSlice("participants"),
			getStringMap("metadata"),
			getStringMap("ref_metadata"),
		)
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channel), false, nil

	case "collab_get_or_create_github_channel":
		repository := strings.TrimSpace(getString("repository"))
		number := strings.TrimSpace(getString("number"))
		if repository == "" {
			return "repository is required", true, nil
		}
		if number == "" {
			return "number is required", true, nil
		}
		kind := strings.TrimSpace(getString("kind"))
		if kind == "" {
			kind = "issue"
		}
		title := strings.TrimSpace(getString("title"))
		if title == "" {
			title = repository + "#" + number
		} else {
			title = repository + "#" + number + " · " + title
		}
		metadata := getStringMap("metadata")
		if metadata == nil {
			metadata = map[string]string{}
		}
		metadata["github_repository"] = repository
		metadata["github_number"] = number
		metadata["github_kind"] = kind
		if url := strings.TrimSpace(getString("url")); url != "" {
			metadata["github_url"] = url
		}
		channel, err := r.collab.GetOrCreateExternalChannel(
			"github",
			repository+"#"+number,
			kind,
			title,
			getString("url"),
			getStringSlice("participants"),
			metadata,
			map[string]string{"kind": kind, "repository": repository, "number": number},
		)
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channel), false, nil

	case "collab_link_slack_thread":
		channel, err := r.collab.LinkExternalRef(getString("channel_id"), &CollabExternalRef{
			Source:     "slack",
			ExternalID: getString("slack_thread_ts"),
			ParentID:   getString("slack_channel_id"),
			Title:      getString("title"),
			URL:        getString("url"),
			Metadata: map[string]string{
				"kind": "thread",
			},
		})
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channel), false, nil

	case "collab_link_external_ref":
		channel, err := r.collab.LinkExternalRef(getString("channel_id"), &CollabExternalRef{
			Source:     getString("source"),
			ExternalID: getString("external_id"),
			ParentID:   getString("parent_id"),
			Title:      getString("title"),
			URL:        getString("url"),
			Metadata:   getStringMap("metadata"),
		})
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channel), false, nil

	case "collab_link_google_doc":
		channel, err := r.collab.LinkExternalRef(getString("channel_id"), &CollabExternalRef{
			Source:     "google_docs",
			ExternalID: getString("doc_id"),
			Title:      getString("title"),
			URL:        getString("url"),
			Metadata:   getStringMap("metadata"),
		})
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channel), false, nil

	case "collab_link_glean_result":
		metadata := getStringMap("metadata")
		if metadata == nil {
			metadata = map[string]string{}
		}
		if datasource := strings.TrimSpace(getString("datasource")); datasource != "" {
			metadata["datasource"] = datasource
		}
		channel, err := r.collab.LinkExternalRef(getString("channel_id"), &CollabExternalRef{
			Source:     "glean",
			ExternalID: getString("result_id"),
			Title:      getString("title"),
			URL:        getString("url"),
			Metadata:   metadata,
		})
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channel), false, nil

	case "collab_link_google_search_result":
		externalID := strings.TrimSpace(getString("result_id"))
		if externalID == "" {
			externalID = strings.TrimSpace(getString("url"))
		}
		metadata := getStringMap("metadata")
		if metadata == nil {
			metadata = map[string]string{}
		}
		if query := strings.TrimSpace(getString("query")); query != "" {
			metadata["query"] = query
		}
		channel, err := r.collab.LinkExternalRef(getString("channel_id"), &CollabExternalRef{
			Source:     "google_search",
			ExternalID: externalID,
			Title:      getString("title"),
			URL:        getString("url"),
			Metadata:   metadata,
		})
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channel), false, nil

	case "collab_link_perplexity_result":
		metadata := getStringMap("metadata")
		if metadata == nil {
			metadata = map[string]string{}
		}
		if query := strings.TrimSpace(getString("query")); query != "" {
			metadata["query"] = query
		}
		channel, err := r.collab.LinkExternalRef(getString("channel_id"), &CollabExternalRef{
			Source:     "perplexity",
			ExternalID: getString("result_id"),
			Title:      getString("title"),
			URL:        getString("url"),
			Metadata:   metadata,
		})
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channel), false, nil

	case "collab_link_yahoo_result":
		externalID := strings.TrimSpace(getString("result_id"))
		if externalID == "" {
			externalID = strings.TrimSpace(getString("url"))
		}
		metadata := getStringMap("metadata")
		if metadata == nil {
			metadata = map[string]string{}
		}
		if query := strings.TrimSpace(getString("query")); query != "" {
			metadata["query"] = query
		}
		channel, err := r.collab.LinkExternalRef(getString("channel_id"), &CollabExternalRef{
			Source:     "yahoo",
			ExternalID: externalID,
			Title:      getString("title"),
			URL:        getString("url"),
			Metadata:   metadata,
		})
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channel), false, nil

	case "collab_link_yahoo_finance_symbol":
		metadata := getStringMap("metadata")
		if metadata == nil {
			metadata = map[string]string{}
		}
		if quoteType := strings.TrimSpace(getString("quote_type")); quoteType != "" {
			metadata["quote_type"] = quoteType
		}
		channel, err := r.collab.LinkExternalRef(getString("channel_id"), &CollabExternalRef{
			Source:     "yahoo_finance",
			ExternalID: strings.ToUpper(strings.TrimSpace(getString("symbol"))),
			Title:      getString("title"),
			URL:        getString("url"),
			Metadata:   metadata,
		})
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channel), false, nil

	case "collab_link_coinbase_asset":
		metadata := getStringMap("metadata")
		if metadata == nil {
			metadata = map[string]string{}
		}
		if productID := strings.TrimSpace(getString("product_id")); productID != "" {
			metadata["product_id"] = productID
		}
		channel, err := r.collab.LinkExternalRef(getString("channel_id"), &CollabExternalRef{
			Source:     "coinbase",
			ExternalID: getString("asset_id"),
			Title:      getString("title"),
			URL:        getString("url"),
			Metadata:   metadata,
		})
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(channel), false, nil

	case "collab_mark_status":
		metadata := getStringMap("metadata")
		channel, err := r.collab.MarkChannelStatus(
			getString("channel_id"),
			getString("status"),
			metadata,
		)
		if err != nil {
			return err.Error(), true, nil
		}
		body := strings.TrimSpace(getString("body"))
		if body == "" {
			return mustJSON(channel), false, nil
		}
		sender := strings.TrimSpace(getString("sender"))
		if sender == "" {
			return "sender is required when body is provided", true, nil
		}
		msgMetadata := cloneMap(metadata)
		if msgMetadata == nil {
			msgMetadata = map[string]string{}
		}
		msgMetadata["channel_status"] = channel.Status
		message, err := r.collab.PostMessage(
			getString("channel_id"),
			sender,
			"status",
			body,
			msgMetadata,
		)
		if err != nil {
			return err.Error(), true, nil
		}
		if r.slack != nil {
			if err := r.slack.MirrorMessage(context.Background(), message); err != nil {
				return err.Error(), true, nil
			}
		}
		return mustJSON(map[string]any{
			"channel": channel,
			"message": message,
		}), false, nil

	case "collab_set_agent_focus":
		agent, err := r.collab.SetAgentFocus(
			getString("agent_id"),
			getStringSlice("aliases"),
			getStringSlice("topics"),
			getStringSlice("capabilities"),
			getString("active_channel_id"),
			getStringMap("metadata"),
		)
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(agent), false, nil

	case "collab_list_agents":
		agents, err := r.collab.ListAgents(
			getString("query"),
			getInt("limit", 50),
		)
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(agents), false, nil

	case "collab_route_message":
		result, err := r.collab.RouteMessage(CollabRouteRequest{
			Sender:          getString("sender"),
			Kind:            getString("kind"),
			Body:            getString("body"),
			ChannelID:       getString("channel_id"),
			Topic:           getString("topic"),
			Query:           getString("query"),
			TargetAgentIDs:  getStringSlice("target_agent_ids"),
			ExcludeAgentIDs: getStringSlice("exclude_agent_ids"),
			Limit:           getInt("limit", 25),
			DryRun:          getBool("dry_run", false),
			Metadata:        getStringMap("metadata"),
			DeliveryMode:    getString("delivery_mode"),
		})
		if err != nil {
			return err.Error(), true, nil
		}
		return mustJSON(result), false, nil

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
		if r.slack != nil {
			if err := r.slack.MirrorMessage(context.Background(), message); err != nil {
				return err.Error(), true, nil
			}
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

	case "proxy_call_tool":
		provider := getString("provider")
		if strings.TrimSpace(provider) == "" {
			return "provider is required", true, nil
		}
		return callProxy(provider, getString("tool"), getObject("arguments"))

	case "github_proxy_call":
		return callProxy("github", getString("tool"), getObject("arguments"))

	case "google_docs_proxy_call":
		return callProxy("google_docs", getString("tool"), getObject("arguments"))

	case "glean_search":
		forwarded := map[string]any{"query": getString("query")}
		if limit := getInt("limit", 0); limit > 0 {
			forwarded["limit"] = limit
		}
		return callProxy("glean", "", forwarded)

	case "perplexity_search_query":
		forwarded := map[string]any{"query": getString("query")}
		if limit := getInt("limit", 0); limit > 0 {
			forwarded["limit"] = limit
		}
		return callProxy("perplexity", "", forwarded)

	case "google_search_query":
		forwarded := map[string]any{"query": getString("query")}
		if limit := getInt("limit", 0); limit > 0 {
			forwarded["limit"] = limit
		}
		return callProxy("google_search", "", forwarded)

	case "yahoo_search_query":
		forwarded := map[string]any{"query": getString("query")}
		if limit := getInt("limit", 0); limit > 0 {
			forwarded["limit"] = limit
		}
		return callProxy("yahoo", "", forwarded)

	case "yahoo_finance_quote":
		return callProxy("yahoo_finance", "", map[string]any{"symbol": strings.ToUpper(strings.TrimSpace(getString("symbol")))})

	case "coinbase_lookup":
		forwarded := map[string]any{}
		if assetID := strings.TrimSpace(getString("asset_id")); assetID != "" {
			forwarded["asset_id"] = assetID
		}
		if productID := strings.TrimSpace(getString("product_id")); productID != "" {
			forwarded["product_id"] = productID
		}
		return callProxy("coinbase", "", forwarded)

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
