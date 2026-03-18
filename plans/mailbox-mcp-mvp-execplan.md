# Build mailbox MCP MVP in services/mcp

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-181` ([issue link](https://linear.app/anmho/issue/ANM-181/featmailbox-build-mailbox-mcp-for-agent-collaboration-workflows)) and `ANM-182` ([issue link](https://linear.app/anmho/issue/ANM-182/server-add-mailbox-mcp-mvp-tools-to-servicesmcp)).

## Purpose / Big Picture

Add a mailbox-style collaboration surface to the existing Go MCP server so agents can perform synchronous collaboration actions today without overloading MCP itself with notification fan-out. The MVP covers ordered, replayable mailbox operations only; live notification plumbing and Slack bridging are explicitly deferred.

## Progress

- [x] (2026-03-18 07:35Z) Confirmed no existing mailbox/collab ExecPlan or implementation exists in the repo.
- [x] (2026-03-18 07:35Z) Created `ANM-181` (parent), `ANM-182` (MVP implementation), `ANM-183` (notification layer follow-up), and `ANM-184` (Slack bridge follow-up).
- [x] (2026-03-18 07:36Z) Corrected the Linear child-parent links after the initial wrong `parentId` attachment and logged the mistake.
- [x] (2026-03-18 07:44Z) Adopted the existing uncommitted collaboration store in `services/mcp/internal/tools/collab.go` as the base instead of duplicating it with a second mailbox implementation.
- [x] (2026-03-18 07:44Z) Added stable `mail_*` mailbox aliases, stable per-agent channel lookup, sequence-based replay reads, and test coverage in `services/mcp/internal/tools/`.
- [x] (2026-03-18 07:44Z) Validated mailbox operations through the local MCP server with `mail_get_channel_for_agent`, `mail_send`, `mail_read`, and `mail_find_channels`.
- [x] (2026-03-18 07:45Z) Re-scanned for follow-up issues and confirmed the existing deferred tickets `ANM-183` and `ANM-184` already cover the remaining major layers.

## Surprises & Discoveries

- Observation: the official MCP steering-group reference server set does not include a mailbox/collaboration server.
  Evidence: `modelcontextprotocol/servers` currently lists reference servers like Filesystem, Git, Memory, Sequential Thinking, and Time, but no mailbox/collaboration server.

- Observation: official/production MCP options exist for adjacent integrations, not mailbox semantics.
  Evidence: the MCP Registry and official integrations list include providers such as Slack-related integrations and multi-app connectors, but these are adapters rather than a mailbox source of truth.

- Observation: the working tree already contained an untracked collaboration store and partial tool registration under `services/mcp/internal/tools/`.
  Evidence: `services/mcp/internal/tools/collab.go`, `services/mcp/internal/tools/collab_test.go`, and modified `registry.go` were present locally before the mailbox-specific replay and alias changes in this slice.

## Decision Log

- Decision: implement the mailbox MVP inside the existing `services/mcp` server first.
  Rationale: the server already exposes generic MCP tools and auth, so adding mailbox tools there is the shortest path to a usable collaboration primitive without introducing another service before the semantics are proven.
  Date/Author: 2026-03-18 / Codex

- Decision: defer notification fan-out and Slack bridging into separate follow-up tickets.
  Rationale: synchronous mailbox operations are the source-of-truth surface; watch streams and external adapters are secondary layers that should not complicate the MVP storage and API semantics.
  Date/Author: 2026-03-18 / Codex

## Outcomes & Retrospective

Completed outcomes:

- `services/mcp` now exposes preferred mailbox tool names: `mail_find_channels`, `mail_get_channel_for_agent`, `mail_send`, and `mail_read`.
- Stable agent mailbox resolution is available via the existing collaboration store by keying channels as `agent:<agent-id>`.
- Message reads now support ordered replay via `after_sequence`.
- Local validation passed with:
  - `npx nx run mcp:test`
  - `npx nx run mcp:build`
  - live MCP smoke test for channel creation, send, ordered read, replay read, and channel listing
- The previous `collab_*` tool names remain available as compatibility aliases.

Remaining gaps:

- Notification proxy and Slack bridge remain explicitly deferred to `ANM-183` and `ANM-184`.
- There is still no live watcher/broadcast layer; the current slice is intentionally synchronous and storage-backed only.

## Context And Orientation

Relevant files for this task:

- `services/mcp/cmd/server/main.go`
- `services/mcp/internal/tools/registry.go`
- `services/mcp/internal/handler/mcp.go`
- `services/mcp/README.md`
- `plans/mailbox-mcp-mvp-execplan.md`

## Plan Of Work

1. Add a durable mailbox store to `services/mcp` with ordered message IDs and replay-friendly reads.
2. Register mailbox tools on the existing MCP registry.
3. Document local usage and validate the new tool flow through the running server.
4. Reconcile follow-up issues and finalize the plan.

## Validation And Acceptance

Acceptance criteria:

- Agents can create or resolve a mailbox channel, send a message, and read ordered messages back through MCP tools.
- Reads support replay by channel and ordered message position.
- The new mailbox tools build and run inside the existing `services/mcp` server without breaking current tools.

Validation commands:

- `npx nx run mcp:test`
- `npx nx run mcp:build`
- Local MCP smoke test for `mail_get_channel_for_agent`, `mail_send`, `mail_read`, and `mail_find_channels`

## Artifacts And Notes

- Parent Linear issue: `ANM-181`
- Active implementation issue: `ANM-182`
- Follow-up issues: `ANM-183`, `ANM-184`
- Session UUID: `019cffb4-e861-7a31-b41f-5e3eeb6c35a9`
