# Build MCP external provider adapters and upstream proxy passthroughs

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-212` ([issue link](https://linear.app/anmho/issue/ANM-212/medium-add-mcp-external-provider-adapters-and-upstream-proxy)).

## Purpose / Big Picture

Extend `services/mcp` so collaboration channels can attach first-class references for GitHub, Google Docs, Google Search, Perplexity, Glean, Yahoo, Yahoo Finance, and Coinbase, while also adding a generic upstream MCP passthrough client for provider-backed tools that should be centralized behind repo auth rather than hardcoded direct API calls.

## Progress

- [x] (2026-03-24T09:09:26Z) Created `ANM-212`, moved it to `In Progress`, assigned it to `me`, and recorded the active Codex session UUID in the ticket description.
- [x] (2026-03-24T09:09:26Z) Reviewed the existing mailbox, Slack, and Linear bridge implementation to reuse the external-ref model instead of adding a parallel provider-specific store.
- [x] (2026-03-26T07:06:53Z) Added generic external-ref helpers plus named MCP tools for GitHub, Google Docs, Google Search, Perplexity, Glean, Yahoo, Yahoo Finance, and Coinbase.
- [x] (2026-03-26T07:06:53Z) Added upstream MCP passthrough client support and named provider proxy tools with centralized auth configuration.
- [x] (2026-03-26T07:06:53Z) Added focused Go tests, updated docs/env examples, and completed the post-change audit plus ticket reconciliation.
- [x] (2026-03-26T07:06:53Z) Fixed validation blocker `ANM-217` so `workspace_apply_patch` preserves unified-diff trailing newlines before calling `git apply`.

## Surprises & Discoveries

- Observation: the current collaboration layer already models Slack and Linear cleanly as external references attached to mailbox channels.
  Evidence: `services/mcp/internal/tools/collab.go` stores `ExternalRefs` per channel and `services/mcp/internal/tools/registry.go` already exposes Linear and Slack tool wrappers.

- Observation: current observability tools are mostly shell-backed, but the collaboration bridge already includes one direct HTTP integration (`LinearClient`) and one event bridge (`SlackBridge`).
  Evidence: `services/mcp/internal/tools/collab_linear.go` and `services/mcp/internal/tools/collab_slack.go`.

- Observation: Google Docs does not need a bespoke server in this repo because credible upstream/open-source MCP servers already exist.
  Evidence: reviewed Google’s `google/mcp` catalog and `ngs/google-mcp-server` on March 25, 2026, then aligned the implementation toward passthrough configuration instead of embedding Docs API logic here.

- Observation: package-level MCP validation exposed an unrelated `workspace_apply_patch` defect in the same service package.
  Evidence: `go test ./...` failed with `git apply --whitespace=nowarn --recount --unidiff-zero -: error: corrupt patch at line 7` until `workspace.go` preserved the trailing patch newline.

## Decision Log

- Decision: extend the existing external-ref abstraction rather than adding one-off provider-specific channel tables or stores.
  Rationale: mailbox state remains the canonical collaboration record, and all provider links can stay queryable through the existing `GetChannelByExternalRef` flow.
  Date/Author: 2026-03-24 / Codex

- Decision: implement a generic upstream MCP JSON-RPC passthrough client and expose named provider wrappers on top of it.
  Rationale: the user wants centralized auth and flexibility to point some integrations at another MCP proxy instead of baking unstable provider-specific APIs directly into the service.
  Date/Author: 2026-03-24 / Codex

## Outcomes & Retrospective

Completed outcomes:

- Collaboration channels can now be keyed or linked to GitHub, Google Docs, Google Search, Perplexity, Glean, Yahoo, Yahoo Finance, and Coinbase references using the same mailbox-backed external-ref model already used for Slack and Linear.
- `services/mcp` now includes an upstream MCP proxy client plus named passthrough tools, so Google Docs and the other providers can be centralized behind env/Secret-Manager-backed auth instead of bespoke direct API code in this service.
- Focused provider proxy tests and full `services/mcp` Go validation pass after the implementation and the `workspace_apply_patch` newline fix.

Remaining gaps:

- `platform mcp` still has the existing nested structured-args limitation tracked in `ANM-204`, so generic passthrough examples for nested `arguments` continue to use raw JSON-RPC instead of the CLI.
- This slice does not provision or run upstream provider MCP servers; it only gives the gateway the adapter and auth layer needed to call them once configured.

## Context And Orientation

Relevant files for this task:

- `services/mcp/internal/tools/collab.go`
- `services/mcp/internal/tools/registry.go`
- `services/mcp/internal/tools/collab_linear.go`
- `services/mcp/internal/tools/collab_slack.go`
- `services/mcp/README.md`
- `services/mcp/.env.example`

## Plan Of Work

1. Add generic external-ref helper methods in the collaboration store so named providers can resolve or attach channels without duplicating logic.
2. Register named MCP tools for the requested providers, including both channel/link tools and upstream passthrough wrappers.
3. Add an env-configured upstream MCP client for proxied provider calls, plus focused tests using a fake local upstream server.
4. Update docs and env examples, then run focused validation and a post-change follow-up scan.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Update `services/mcp/internal/tools/collab.go` with generic provider-backed channel/link helpers.
2. Add provider proxy client/helpers under `services/mcp/internal/tools/`.
3. Register new tool definitions and dispatch paths in `services/mcp/internal/tools/registry.go`.
4. Add focused tests under `services/mcp/internal/tools/`.
5. Update `services/mcp/README.md` and `.env.example`.
6. Run:
   - `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`
   - `cd services/mcp && GOCACHE=/tmp/go-cache go build ./...`

## Validation And Acceptance

Acceptance criteria:

- Collaboration channels can be resolved or linked against the requested external providers without breaking existing Slack/Linear flows.
- Named provider proxy tools can forward to a configured upstream MCP HTTP endpoint using centralized auth env vars.
- Focused tests cover both external-ref storage behavior and upstream proxy request/response handling.

Validation to run:

- `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`
- `cd services/mcp && GOCACHE=/tmp/go-cache go build ./...`

Validation executed:

- `cd services/mcp && GOCACHE=/tmp/go-cache go test ./internal/tools`
- `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`
- `cd services/mcp && GOCACHE=/tmp/go-cache go build ./...`

## Idempotence And Recovery

External-ref writes must remain idempotent so repeated link or get-or-create requests do not duplicate channel metadata. Upstream proxy failures must return actionable tool errors without mutating mailbox state.

## Artifacts And Notes

- Active implementation ticket: `ANM-212`
- Validation unblocker ticket: `ANM-217`
- Session UUID: `019cffb4-e861-7a31-b41f-5e3eeb6c35a9`

## Interfaces And Dependencies

- Existing mailbox channel and external-ref persistence in `services/mcp`
- Environment-driven auth and endpoint config for upstream MCP proxies
- Existing Slack and Linear bridge behavior must remain backward compatible
