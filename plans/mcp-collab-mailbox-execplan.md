# Add mailbox collaboration tools to the MCP service

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-178` ([issue link](https://linear.app/anmho/issue/ANM-178/major-add-mailbox-collaboration-mcp-tools-to-servicesmcp)), `ANM-179` ([issue link](https://linear.app/anmho/issue/ANM-179/medium-add-collaboration-watcher-proxy-for-mailbox-broadcast-and-fan)), `ANM-180` ([issue link](https://linear.app/anmho/issue/ANM-180/medium-add-slack-bridge-between-collaboration-mailbox-events-and)), and `ANM-185` ([issue link](https://linear.app/anmho/issue/ANM-185/medium-add-channel-level-access-control-for-mailbox-mcp-collaboration)).

## Purpose / Big Picture

Add runtime-agnostic mailbox collaboration tools to `services/mcp` so agents collaborate through standard MCP tool calls instead of a control-plane-specific contract. The first slice focuses on synchronous collaboration workflows only: channel discovery/selection, finding channels by agent, posting messages, and reading recent messages. Broadcast/watcher mechanics and Slack bridging stay as explicit follow-up work.

## Progress

- [x] (2026-03-18 07:34Z) Created `ANM-178` for mailbox MCP core and moved it to `In Progress`; created follow-up tickets `ANM-179` (watcher proxy) and `ANM-180` (Slack bridge) in `Backlog`.
- [x] (2026-03-18 07:35Z) Audited the current `services/mcp` registry and confirmed tool definitions are hand-maintained in `internal/tools/registry.go`, making core mailbox tools implementable without changing the Connect wrapper first.
- [x] (2026-03-18 07:37Z) Added lightweight persistent collaboration storage and mailbox domain model under `services/mcp/internal/tools`.
- [x] (2026-03-18 07:37Z) Exposed mailbox/collab MCP tools in the registry and wired tool execution paths.
- [x] (2026-03-18 07:38Z) Added focused tests for channel creation/listing and message post/read flows.
- [x] (2026-03-18 07:38Z) Validated the mailbox slice with `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`.
- [x] (2026-03-18 07:39Z) Updated `services/mcp/README.md` and `.env.example` with mailbox tool usage and store configuration.
- [x] (2026-03-18 07:39Z) Performed post-change audit, discovered overlapping collaboration tickets already existed in Linear, and recorded the overlap in this plan plus `docs/agent-mistakes.md`.

## Surprises & Discoveries

- Observation: `services/mcp` currently exposes tools solely through the in-process registry, not through a separate tool plugin system.
  Evidence: `services/mcp/internal/tools/registry.go` defines both tool metadata and execution in one file.

- Observation: There is no existing collaboration/mailbox store in the MCP service, so the first slice needs its own persistence model.
  Evidence: `find services/mcp -maxdepth 4 -type f` showed no mailbox/collab modules before this work.

- Observation: Linear already contained overlapping collaboration/mailbox tickets by the time the post-change reconciliation scan ran.
  Evidence: focused Linear issue search returned overlapping collaboration work items after `ANM-178` through `ANM-180` had already been created.

## Decision Log

- Decision: Implement the initial collaboration abstraction as MCP tools in `services/mcp`, not as part of agent-control.
  Rationale: The user explicitly wants collaboration to be an additional tool call shared across runtimes.
  Date/Author: 2026-03-18 / Codex

- Decision: Scope the initial slice to synchronous mailbox operations and defer broadcast/watchers to a separate ticket.
  Rationale: Current agent harnesses are better at explicit tool calls than async context injection; this keeps the first contract stable and broadly usable.
  Date/Author: 2026-03-18 / Codex

- Decision: Use lightweight persistence first instead of choosing Kafka now.
  Rationale: Kafka/proxy is only justified once the mailbox event model is stable and the watcher/fan-out requirements are being implemented.
  Date/Author: 2026-03-18 / Codex

## Outcomes & Retrospective

Completed outcomes:

- `services/mcp` now exposes mailbox collaboration tools for channel create/select, channel listing/filtering, agent-centric lookup, message post, and message read flows.
- Mailbox state is persisted locally and documented via `services/mcp/README.md` and `services/mcp/.env.example`.
- Focused `services/mcp` tests pass with the new collaboration storage and tool logic.

Remaining gaps:

- Watcher proxy and Slack bridge remain intentionally out of scope for this slice and are tracked in `ANM-179` and `ANM-180`.
- Collaboration mailbox access control/authz is still absent from the initial slice and is now tracked in `ANM-185`.
- There is overlap between the newly created collaboration tickets and pre-existing collaboration issues discovered during the reconciliation scan.

## Context And Orientation

Relevant files for this task:

- `services/mcp/internal/tools/registry.go`
- `services/mcp/internal/handler/mcp.go`
- `services/mcp/internal/jsonrpc/passthrough.go`
- `services/mcp/README.md`
- `services/mcp/project.json`

## Plan Of Work

1. Add mailbox persistence and domain helpers to `services/mcp`.
2. Add collaboration tools for channel create/select, list/find, post, and read.
3. Test and validate the new tools through the existing server/tool registry.
4. Rescan for remaining collaboration gaps and keep them ticketed separately.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Implement persistent mailbox store under `services/mcp/internal/tools`.
2. Add tool definitions and call handlers in `services/mcp/internal/tools/registry.go`.
3. Run:
   - `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`
4. Optionally smoke-check with:
   - `cd services/mcp && GOCACHE=/tmp/go-cache go run ./cmd/server`
   - query the tool registry through the existing MCP CLI or JSON-RPC path.

## Validation And Acceptance

Acceptance criteria:

- Agents can create/select channels, list/find channels by agent, post messages, and read recent messages through MCP tools.
- The mailbox tool contract is runtime-agnostic and does not depend on Slack or watcher infrastructure.
- `services/mcp` tests pass after the mailbox tools are added.

Validation executed:

- `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`

## Idempotence And Recovery

The mailbox store format should be safe to re-read and rewrite. If the implementation changes shape during this slice, preserve compatibility or add a straightforward migration for the local store file.

## Artifacts And Notes

- Active implementation ticket: `ANM-178`
- Follow-up tickets: `ANM-179`, `ANM-180`
- Session UUID: `019cffb4-bcfb-7eb0-a4b4-09047c0282d5`

## Interfaces And Dependencies

- Existing MCP tool registry and JSON-RPC/Connect wrappers in `services/mcp`
- Local filesystem persistence for the first mailbox slice
- Future watcher/broadcast infrastructure remains decoupled from the initial tool contract
