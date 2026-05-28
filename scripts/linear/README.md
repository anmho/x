# Linear Integration

Create Linear issues via **Linear MCP** (Model Context Protocol). No JS script—agents use the MCP tool directly.

## Setup

1. Add Linear MCP in Cursor: [cursor.com/agents](https://cursor.com/agents) → MCP → Linear
2. URL: `https://mcp.linear.app/mcp`
3. Auth: OAuth (complete in dashboard) or API key

## Usage

When instructed to create Linear tickets, use the Linear MCP tool (e.g. `create_issue` or equivalent).

## Payloads

Backlog payloads live in `docs/backlog/*.json`. Each has a `tickets` array. Use Linear MCP to create issues from these payloads when asked.
