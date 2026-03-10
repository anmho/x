# Greptile MCP Setup

Greptile provides AI-powered code search and analysis via MCP. Configure it for Cursor, Claude Code, and Codex.

**Prerequisite:** Get your API key from [app.greptile.com/settings/api](https://app.greptile.com/settings/api)

If MCP startup says `Environment variable GREPTILE_API_KEY ... is not set`, the project config is present but the shell that launched the client did not export the key yet. Fix the environment first, then restart the client from a fresh shell.

---

## Cursor

1. Open `.cursor/mcp.json` (or Settings → Tools & MCP)
2. Add the Greptile entry (or replace `YOUR_GREPTILE_API_KEY` with your key):

```json
"greptile": {
  "url": "https://api.greptile.com/mcp",
  "headers": {
    "Authorization": "Bearer YOUR_GREPTILE_API_KEY"
  }
}
```

3. Restart Cursor

---

## Claude Code (CLI)

Add via CLI:

```bash
claude mcp add --transport http greptile https://api.greptile.com/mcp \
  --header "Authorization: Bearer YOUR_GREPTILE_API_KEY"
```

Or use project-level `.mcp.json` (uses `GREPTILE_API_KEY` env var):

```bash
export GREPTILE_API_KEY=your-api-key-here
claude
```

To make it persistent across new terminals:

```bash
echo 'export GREPTILE_API_KEY=your-api-key-here' >> ~/.zshrc
exec zsh -l
```

---

## Codex (CLI)

Add via CLI:

```bash
codex mcp add greptile --url https://api.greptile.com/mcp \
  --bearer-token-env-var GREPTILE_API_KEY
export GREPTILE_API_KEY=your-api-key-here
```

Or use project-level `.codex/config.toml` (already in repo):

```bash
export GREPTILE_API_KEY=your-api-key-here
codex
```

If Codex already started before the export was added, restart Codex from a shell that can already run `printenv GREPTILE_API_KEY`.

---

## Claude Desktop

Uses the stdio `greptile-mcp-server` package. Add to `claude_desktop_config.json`:

```json
"greptile": {
  "command": "npx",
  "args": ["greptile-mcp-server"],
  "env": {
    "GREPTILE_API_KEY": "your_greptile_api_key_here",
    "GITHUB_TOKEN": "your_github_token_here"
  }
}
```

---

## Verify

```bash
zsh -lc 'printenv GREPTILE_API_KEY'
curl -X POST https://api.greptile.com/mcp \
  -H "Authorization: Bearer YOUR_GREPTILE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"ping"}'
```

Expected:

- First command prints a non-empty key value.
- Second command returns `{"jsonrpc":"2.0","id":1,"result":{}}`

---

## Project Files

| File | Purpose |
|------|---------|
| `.cursor/mcp.json` | Cursor (gitignored; add Greptile manually) |
| `.mcp.json` | Claude Code project config (uses `GREPTILE_API_KEY`) |
| `.codex/config.toml` | Codex project config (uses `GREPTILE_API_KEY`) |
