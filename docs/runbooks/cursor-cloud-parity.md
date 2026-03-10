# Cursor Cloud Parity with Local Setup

Configure Cursor Cloud Agents to match local Cursor capabilities: MCP servers and secrets.

**Note:** Cursor does not provide a CLI or API to set secrets or MCPs programmatically. Use the dashboard.

## Quick links

- **Secrets:** [cursor.com/dashboard?tab=cloud-agents](https://cursor.com/dashboard?tab=cloud-agents) → Secrets
- **MCPs:** [cursor.com/agents](https://cursor.com/agents) → MCP dropdown

---

## Secrets

Add these in **Dashboard → Cloud Agents → Secrets**:

| Key | Purpose | Scopes / notes |
|-----|---------|----------------|
| `GITHUB_PERSONAL_ACCESS_TOKEN` | GitHub MCP (PR creation, repo access) | `repo`, `read:packages` |
| `LINEAR_API_KEY` | Linear MCP (OAuth preferred) or CLI fallback (create-issues when MCP unavailable) | [linear.app/settings/api](https://linear.app/settings/api) — Ralph loop friendly |
| `GREPTILE_API_KEY` | Greptile MCP (code search/analysis) | [app.greptile.com/settings/api](https://app.greptile.com/settings/api) |
| `LINEAR_API_KEY` | Linear CLI fallback (create-issues when MCP unavailable) | [linear.app/settings/api](https://linear.app/settings/api) — Ralph loop friendly |

Create a GitHub PAT at [github.com/settings/tokens/new](https://github.com/settings/tokens/new).

---

## MCPs

Add and enable these in **cursor.com/agents** → MCP dropdown:

### 1. GitHub MCP

- **Type:** HTTP (Streamable HTTP)
- **URL:** `https://api.githubcopilot.com/mcp/`
- **Auth:** Bearer token — use `GITHUB_PERSONAL_ACCESS_TOKEN` from Secrets (or configure OAuth if available)

### 2. Linear MCP

- **Type:** HTTP
- **URL:** `https://mcp.linear.app/mcp`
- **Auth:** OAuth (per-user; complete in dashboard)

### 3. Greptile MCP

- **Type:** HTTP
- **URL:** `https://api.greptile.com/mcp`
- **Auth:** Bearer token — use `GREPTILE_API_KEY` from Secrets
- If startup reports missing `GREPTILE_API_KEY`, the secret is absent from the active agent/session environment and the agent must be restarted after adding it.

See [greptile-mcp-setup.md](./greptile-mcp-setup.md) for full setup (Cursor, Claude Code, Codex).

---

## Local reference

Local config lives in `.cursor/mcp.json` (gitignored):

```json
{
  "mcpServers": {
    "linear": {
      "url": "https://mcp.linear.app/mcp"
    },
    "greptile": {
      "url": "https://api.greptile.com/mcp",
      "headers": {
        "Authorization": "Bearer YOUR_GREPTILE_API_KEY"
      }
    },
    "github": {
      "url": "https://api.githubcopilot.com/mcp/",
      "headers": {
        "Authorization": "Bearer YOUR_GITHUB_PAT"
      }
    }
  }
}
```

Cloud agents pick up MCP from the dashboard; project-level `.cursor/mcp.json` is ignored when gitignored.

### Adding Linear MCP locally

1. Ensure `.cursor/mcp.json` includes the Linear entry (see Local reference above).
2. Restart Cursor.
3. Complete OAuth: when first using Linear tools, Cursor will prompt you to log in at [mcp.linear.app](https://mcp.linear.app/mcp).
4. If the URL transport fails, use the stdio fallback in `.cursor/mcp.json`:
   ```json
   "linear": {
     "command": "npx",
     "args": ["-y", "mcp-remote", "https://mcp.linear.app/mcp"]
   }
   ```

---

## Verification

1. Add secrets and MCPs in the dashboard.
2. Restart or launch a new cloud agent.
3. Confirm the restarted agent no longer reports a missing `GREPTILE_API_KEY` for Greptile MCP.
4. Ask the agent to list GitHub repos or Linear issues to confirm MCP access.
