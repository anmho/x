# Add watcher proxy streaming to the MCP mailbox service

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-179` ([issue link](https://linear.app/anmho/issue/ANM-179/medium-add-collaboration-watcher-proxy-for-mailbox-broadcast-and-fan)).

## Purpose / Big Picture

Add a lightweight watcher proxy behind the mailbox collaboration model in `services/mcp` so live consumers can replay ordered mailbox events and then live-tail new ones without changing the synchronous MCP tool contract. This slice keeps storage and tool semantics in `mail_*` as the source of truth, adds a streaming surface for listeners/UIs, and explicitly defers Slack plus external event backbones.

## Progress

- [x] (2026-03-24 07:29Z) Moved `ANM-179` to `In Progress`, confirmed ownership is `me`, and attached the active Codex session UUID in a Linear comment.
- [x] (2026-03-24 07:31Z) Audited existing streaming patterns in the repo and selected the agent-control `replay then tail` model plus mailbox sequence cursors as the watcher contract baseline.
- [x] (2026-03-24 07:38Z) Added in-process mailbox subscriptions and ordered replay helpers to `services/mcp/internal/tools/collab.go`.
- [x] (2026-03-24 07:39Z) Added authenticated SSE watcher handling at `/mailbox/events` in `services/mcp/cmd/server/main.go` and `services/mcp/internal/watch/mailbox.go`.
- [x] (2026-03-24 07:40Z) Added focused tests covering mailbox fan-out/order plus replay/live-tail watcher behavior.
- [x] (2026-03-24 07:41Z) Validated with `npx nx run mcp:test`, `npx nx run mcp:build`, and `./scripts/deploy-preflight mcp`.
- [x] (2026-03-24 07:44Z) Completed a live smoke test against `./bin/mcp-server`, verifying `ready`, replayed `message`, and live-tailed `message` events from `/mailbox/events`.
- [x] (2026-03-24 07:46Z) Completed the post-change audit and reconciled remaining gaps to existing follow-ups plus new ticket `ANM-195` for watcher backpressure handling.

## Surprises & Discoveries

- Observation: `services/mcp` did not already contain any SSE or watcher HTTP surface, so the watcher slice had to introduce a new authenticated endpoint rather than extending ConnectRPC unary calls or JSON-RPC tool responses.
  Evidence: `services/mcp/cmd/server/main.go` only mounted ConnectRPC admin/tool handlers plus `/mcp` and `/health` before this change.

- Observation: The mailbox store already had global monotonic `Sequence` values, which made replay cursors straightforward without adding a new event table or transport-specific cursor type.
  Evidence: `services/mcp/internal/tools/collab.go` increments `NextSequence` for every appended mailbox message.

- Observation: A small in-process subscriber bus is enough for this slice because mailbox writes already pass through one process, and replay from persisted storage closes the recovery gap after reconnects/restarts.
  Evidence: `mail_send` and related writes go through the in-process `Registry` store instance used by the server.

## Decision Log

- Decision: Use Server-Sent Events on a plain authenticated HTTP endpoint instead of introducing ConnectRPC server streaming for mailbox watchers right now.
  Rationale: SSE works for browsers and simple local consumers immediately, while the mailbox tools remain the canonical request/response surface.
  Date/Author: 2026-03-24 / Codex

- Decision: Keep the watcher source of truth in the existing mailbox store and add a small in-memory subscriber bus only for live wake-ups.
  Rationale: Persisted mailbox messages already provide ordered replay; the bus only removes the need to poll the file for new writes.
  Date/Author: 2026-03-24 / Codex

- Decision: Support both channel-scoped streams and an all-channel ordered feed.
  Rationale: Per-channel agent mailboxes are the main use case, but UIs and notification adapters need an aggregate feed without extra plumbing.
  Date/Author: 2026-03-24 / Codex

## Outcomes & Retrospective

Completed outcomes:

- `services/mcp` now serves `GET /mailbox/events` with ordered replay (`after_sequence` or `Last-Event-ID`) and live-tail semantics.
- Multiple listeners can subscribe concurrently through the in-process mailbox bus, and replay remains available after reconnects or restarts.
- The watcher endpoint stays separate from `mail_*` MCP tools, preserving the mailbox tool contract as the synchronous source of truth.

Remaining gaps:

- Slow subscribers are disconnected rather than backpressuring writers; reconnect-and-replay is the recovery path for now.
- Slack bridging and any external backbone choice remain out of scope and continue to live in follow-up work (`ANM-180` and related items).
- Mailbox access control is still a separate follow-up tracked in `ANM-185`.
- Backpressure/overflow handling improvements for slow watcher clients are now tracked in `ANM-195`.

## Context And Orientation

Relevant files for this task:

- `services/mcp/internal/tools/collab.go`
- `services/mcp/internal/watch/mailbox.go`
- `services/mcp/cmd/server/main.go`
- `services/mcp/README.md`

## Plan Of Work

1. Extend mailbox storage with replay helpers and a live notification primitive.
2. Add an authenticated watcher endpoint that replays first and then tails new mailbox messages.
3. Test the watcher behavior and document the endpoint contract for local consumers.
4. Reconcile remaining notification/access-control follow-ups after validation.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Implement mailbox subscriber and replay support under `services/mcp/internal/tools`.
2. Add `GET /mailbox/events` in the MCP server with the existing API key auth middleware.
3. Run:
   - `npx nx run mcp:test`
   - `npx nx run mcp:build`
   - `./scripts/deploy-preflight mcp`
4. Smoke-check with:
   - `./bin/mcp-server`
   - `curl -N -H "Authorization: Bearer <api-key>" "http://127.0.0.1:8765/mailbox/events?..."`

## Validation And Acceptance

Acceptance criteria:

- A listener can reconnect with `after_sequence` or `Last-Event-ID` and receive ordered replayed messages.
- A live listener receives new mailbox messages without polling the mailbox file directly.
- Multiple listeners can subscribe concurrently without changing the `mail_*` MCP tools.

Validation executed:

- `npx nx run mcp:test`
- `npx nx run mcp:build`
- `./scripts/deploy-preflight mcp`
- Live smoke test using `./bin/mcp-server`, `./bin/mcp`, and `curl -N` against `/mailbox/events`

## Idempotence And Recovery

Watcher clients should treat mailbox sequence values as resumable cursors. If a stream disconnects, reconnect with the last processed sequence via `after_sequence` or `Last-Event-ID` and continue consuming from persisted mailbox state.

## Artifacts And Notes

- Active implementation ticket: `ANM-179`
- Follow-up tickets: `ANM-180`, `ANM-185`, `ANM-195`
- Session UUID: `019cffb4-e861-7a31-b41f-5e3eeb6c35a9`

## Interfaces And Dependencies

- Existing API key auth middleware in `services/mcp/cmd/server/main.go`
- Existing mailbox storage and `mail_*` tools in `services/mcp/internal/tools`
- SSE clients that can reconnect with `Last-Event-ID` or an explicit `after_sequence`
