# MCP service

`services/mcp` exposes the repository's internal observability tools over ConnectRPC and MCP JSON-RPC.

## Local development

Portable container setup via platform CLI:

```bash
platform mcp setup
docker build -f services/mcp/Dockerfile -t x-mcp .
docker run --rm -p 8765:8765 --env-file "$HOME/.x-mcp/mcp.env" -v "$HOME/.x-mcp:/root/.x-mcp" x-mcp
```

`platform mcp setup` writes a Docker-ready env file under `~/.x-mcp/mcp.env` by default, including the gateway API key, optional Slack/Linear credentials, and optional upstream MCP proxy settings for providers like Google Docs and Perplexity.

From repo root:

```bash
npx nx run mcp:generate-proto
npx nx run mcp:test
npx nx run mcp:build
npx nx run mcp:build-cli
./scripts/deploy-preflight mcp
```

Run the server:

```bash
cd services/mcp
GOCACHE=/tmp/go-cache go run ./cmd/server
```

On first startup the server generates a key in `~/.x-mcp/keys.json` and prints it once. Use that key with the standalone CLI:

```bash
./bin/mcp --server http://localhost:8765 --key <api-key> tools list
./bin/mcp --server http://localhost:8765 --key <api-key> keys list
```

If `platform` is installed and version-aligned, prefer it for examples. The repo-local `./platform` wrapper remains a fallback when you want to force the checked-out copy:

```bash
platform mcp tools list --server http://localhost:8765 --key <api-key>
# fallback: ./platform mcp tools list --server http://localhost:8765 --key <api-key>
```

## Collaboration mailbox tools

The MCP service now includes mailbox-style collaboration tools intended for agent-to-agent coordination via normal MCP tool calls.

Preferred mailbox tool names:

- `mail_find_channels`
- `mail_get_channel_for_agent`
- `mail_send`
- `mail_read`

Compatibility aliases remain available under the older `collab_*` names:

- `collab_get_or_create_channel`
- `collab_list_channels`
- `collab_find_channels_by_agent`
- `collab_post_message`
- `collab_read_messages`
- `collab_get_or_create_linear_channel`
- `collab_get_or_create_external_channel`
- `collab_get_or_create_github_channel`
- `collab_get_channel_by_external_ref`
- `collab_link_slack_thread`
- `collab_link_external_ref`
- `collab_link_google_doc`
- `collab_link_glean_result`
- `collab_link_google_search_result`
- `collab_link_perplexity_result`
- `collab_link_yahoo_result`
- `collab_link_yahoo_finance_symbol`
- `collab_link_coinbase_asset`
- `collab_mark_status`
- `collab_set_agent_focus`
- `collab_list_agents`
- `collab_route_message`

The mailbox model remains the source of truth. Slack threads and Linear issues are attached to channels as external references instead of replacing mailbox state.

Mailbox state is persisted locally by default at `~/.x-mcp/collab.json`. Override with `MCP_MAILBOX_FILE` or `MCP_COLLAB_STORE` if needed.

Example flow:

```bash
platform mcp --server http://localhost:8765 --key <api-key> \
  tools call mail_get_channel_for_agent agent_id=agent-a

# use the returned channel id from the previous call
platform mcp --server http://localhost:8765 --key <api-key> \
  tools call mail_send channel_id=<channel-id> sender=agent-b body='Picked up the storage slice'

platform mcp --server http://localhost:8765 --key <api-key> \
  tools call mail_read channel_id=<channel-id> after_sequence=0 limit=20
```

Linear-backed channel flow:

```bash
platform mcp --server http://localhost:8765 --key <api-key> \
  tools call collab_get_or_create_linear_channel issue_id=ANM-190 participants=agent-a

platform mcp --server http://localhost:8765 --key <api-key> \
  tools call collab_mark_status channel_id=<channel-id> status=waiting_approval sender=agent-a body='Waiting on human approval'
```

Slack thread linking:

```bash
platform mcp --server http://localhost:8765 --key <api-key> \
  tools call collab_link_slack_thread channel_id=<channel-id> slack_channel_id=C123 slack_thread_ts=1742800000.100000
```

External provider channel/linking examples:

```bash
# create a GitHub-backed collaboration channel
platform mcp --server http://localhost:8765 --key <api-key> \
  tools call collab_get_or_create_github_channel repository=anmho/x number=212 kind=pull_request \
  title='External adapters'

# link an existing channel to a Google Doc
platform mcp --server http://localhost:8765 --key <api-key> \
  tools call collab_link_google_doc channel_id=<channel-id> doc_id=<google-doc-id> \
  url='https://docs.google.com/document/d/<google-doc-id>/edit'

# link a search or finance artifact into the same channel
platform mcp --server http://localhost:8765 --key <api-key> \
  tools call collab_link_google_search_result channel_id=<channel-id> \
  url='https://example.com/result' query='agent mailbox mcp'

platform mcp --server http://localhost:8765 --key <api-key> \
  tools call collab_link_yahoo_finance_symbol channel_id=<channel-id> symbol=BTC-USD
```

Supported external-ref source names now include:

- `github`
- `google_docs`
- `google_search`
- `perplexity`
- `glean`
- `yahoo`
- `yahoo_finance`
- `coinbase`

Agent routing flow for dispatcher delivery:

```bash
# register a run-aware agent so routed mailbox deliveries can resolve to an active run
platform mcp --server http://localhost:8765 --key <api-key> \
  tools call collab_set_agent_focus agent_id=agent-a topics=routing capabilities=interrupt metadata.run_id=<run-id>

# inspect registered agent routing state
platform mcp --server http://localhost:8765 --key <api-key> \
  tools call collab_list_agents query=routing

# route a mailbox message to matching agents and include dispatcher delivery hints
platform mcp --server http://localhost:8765 --key <api-key> \
  tools call collab_route_message sender=control-plane body='Pick up the routing slice' \
  topic=routing delivery_mode=interrupt_and_replan
```

`collab_route_message` writes normal mailbox messages into each target agent mailbox, but it also attaches structured metadata such as `target_agent_id`, `delivery_mode`, and route scoring. `services/agent-control-api` can tail `/mailbox/events`, resolve that metadata back to `run_id`, and forward the message through `PushRunEvent`.

## Mailbox watcher endpoint

The MCP service now exposes an authenticated watcher endpoint for replay plus live tail:

```text
GET /mailbox/events?channel_id=<channel-id>&after_sequence=<sequence>&replay_limit=<n>
```

- Use `Authorization: Bearer <api-key>` or `X-Api-Key: <api-key>`.
- `channel_id` is optional. Omit it to watch the ordered mailbox feed across all channels.
- `after_sequence` resumes from the last sequence you have already processed.
- `Last-Event-ID` is also accepted as the replay cursor when reconnecting SSE clients.

Example:

```bash
curl -N \
  -H "Authorization: Bearer <api-key>" \
  "http://localhost:8765/mailbox/events?channel_id=<channel-id>&after_sequence=0"
```

The stream emits a `ready` event first, then ordered `message` events with SSE `id` values set to the mailbox sequence. Slow listeners may be disconnected if they stop draining their stream; reconnect with `after_sequence` or `Last-Event-ID` to replay what you missed.

For worker delivery, the watcher stays observer-oriented. The dispatcher integration lives in `services/agent-control-api` and is enabled with:

- `MCP_MAILBOX_EVENTS_URL`
- `MCP_API_KEY`
- `MCP_MAILBOX_FILE`
- `MCP_DISPATCH_STATE_FILE`

When `MCP_MAILBOX_EVENTS_URL` is configured, the control plane reuses the same `PushRunEvent` path used by the public API, so mailbox-originated routed messages and direct control-plane push events behave consistently for local active runs.

## Slack bridge

The MCP server can also accept Slack Events API callbacks:

```text
POST /slack/events
```

Behavior:

- If a Slack thread is already linked, inbound replies append to the existing mailbox channel.
- If the Slack message includes a Linear issue identifier such as `ANM-190`, the bridge resolves or creates a channel keyed to `linear:ANM-190`.
- If `LINEAR_API_KEY` is configured, issue title, URL, state, and team metadata are added to the mailbox channel.
- If `SLACK_BOT_TOKEN` is configured, mailbox messages posted through MCP tools are mirrored back into the linked Slack thread unless the message originated from Slack.

Relevant environment variables:

- `SLACK_SIGNING_SECRET`
- `SLACK_BOT_TOKEN`
- `LINEAR_API_KEY`

## Upstream provider passthroughs

The MCP gateway can also proxy selected provider tools to another HTTP MCP endpoint. This keeps auth centralized in `services/mcp` while letting you reuse upstream/open-source MCP servers instead of duplicating provider-specific API clients locally.

Named passthrough tools:

- `proxy_call_tool`
- `github_proxy_call`
- `google_docs_proxy_call`
- `glean_search`
- `perplexity_search_query`
- `google_search_query`
- `yahoo_search_query`
- `yahoo_finance_quote`
- `coinbase_lookup`

Configuration model:

- Shared defaults:
  - `MCP_PROXY_DEFAULT_URL`
  - `MCP_PROXY_DEFAULT_KEY`
  - `MCP_PROXY_DEFAULT_HEADER`
  - `MCP_PROXY_DEFAULT_SCHEME`
- Provider-specific overrides:
  - `MCP_PROXY_GITHUB_*`
  - `MCP_PROXY_GOOGLE_DOCS_*`
  - `MCP_PROXY_GOOGLE_SEARCH_*`
  - `MCP_PROXY_PERPLEXITY_*`
  - `MCP_PROXY_GLEAN_*`
  - `MCP_PROXY_YAHOO_*`
  - `MCP_PROXY_YAHOO_FINANCE_*`
  - `MCP_PROXY_COINBASE_*`

Each provider accepts:

- `URL`
- `KEY`
- `HEADER`
- `SCHEME`
- `TOOL`

Examples:

```bash
# named Google Search wrapper
platform mcp --server http://localhost:8765 --key <api-key> \
  tools call google_search_query query='agent mailbox mcp' limit=5

# generic passthrough with nested arguments via raw JSON-RPC
curl -sS \
  -H "Authorization: Bearer <api-key>" \
  -H "Content-Type: application/json" \
  http://localhost:8765/mcp \
  -d '{
    "jsonrpc": "2.0",
    "id": "github-search",
    "method": "tools/call",
    "params": {
      "name": "proxy_call_tool",
      "arguments": {
        "provider": "github",
        "tool": "github_search",
        "arguments": {
          "query": "repo:anmho/x mailbox"
        }
      }
    }
  }'

# named Google Docs wrapper via raw JSON-RPC for nested args
curl -sS \
  -H "Authorization: Bearer <api-key>" \
  -H "Content-Type: application/json" \
  http://localhost:8765/mcp \
  -d '{
    "jsonrpc": "2.0",
    "id": "google-doc",
    "method": "tools/call",
    "params": {
      "name": "google_docs_proxy_call",
      "arguments": {
        "tool": "docs_document_get",
        "arguments": {
          "doc_id": "<google-doc-id>"
        }
      }
    }
  }'
```

For Google Docs specifically, prefer pointing `google_docs_proxy_call` at an upstream Google Workspace or Google Docs MCP instead of embedding Google Docs API logic directly in this gateway. As of March 25, 2026, credible open-source options include Google’s `google/mcp` workspace catalog and `ngs/google-mcp-server`.

## Docker

Build from the repo root:

```bash
docker build -f services/mcp/Dockerfile -t x-mcp .
```

Run locally with the key store persisted:

```bash
docker run --rm -p 8765:8765 \
  -v "$HOME/.x-mcp:/root/.x-mcp" \
  x-mcp
```

The image now includes `platform.controlplane.json` and sets `MCP_ROOT=/app`, so config-backed tools work without an extra bind mount. Tools that shell out to `gcloud` or `vercel` still require those CLIs and their auth to be available in the runtime environment.

## Deployment model

This repo already uses a declarative platform workflow rather than ArgoCD:

1. Edit `infra/platform/declarative_spec.py`.
2. Materialize generated configs with `python3 scripts/ci/materialize_platform_configs.py`.
3. Review with `platform control-plane plan --project mcp`.
4. Dry-run deployment with `platform deploy --project mcp --dry-run`.

That keeps `mcp` aligned with the same generated `platform.projects.json` and `platform.controlplane.json` flow used by the other services.
