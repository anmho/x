# Add interruptible push-event routing for agent control and MCP collaboration

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-188` ([issue link](https://linear.app/anmho/issue/ANM-188/major-add-interruptible-push-event-routing-for-agent-control-and-mcp)).

## Purpose / Big Picture

Add a durable push-event path for agent execution so the system can broadcast or route a message into active work, including interrupt-like delivery for in-progress runs. The agent-facing abstraction should stay as MCP tools in `services/mcp`, while worker delivery moves through a dispatcher service that consumes mailbox deliveries and forwards them into the control plane. The watcher endpoint remains an observer/debug/UI surface, not the worker transport.

## Progress

- [x] (2026-03-18 07:47Z) Created `ANM-188`, assigned it to `me`, moved it to `In Progress`, and recorded Codex session UUID `019cffb4-bcfb-7eb0-a4b4-09047c0282d5`.
- [x] (2026-03-18 07:48Z) Confirmed existing mailbox plans cover synchronous read/write collaboration only and do not yet cover routed push delivery or interrupt semantics.
- [x] (2026-03-24 08:04Z) Updated the execution model: watcher stays observer-only, dispatcher service becomes the worker delivery path for mailbox-driven routing.
- [x] (2026-03-24 08:18Z) Wired the first-slice dispatcher into `services/agent-control-api`, reusing the `PushRunEvent` delivery path and adding dispatcher/env tests.
- [x] (2026-03-24 08:18Z) Exposed MCP agent registration and routing tools plus mailbox metadata coverage in `services/mcp` tests so dispatcher forwarding has stable target hints.
- [x] (2026-03-24 08:18Z) Validated the local code path with `cd services/agent-control-api && GOCACHE=/tmp/go-cache go test ./... && go build ./...` and `cd services/mcp && GOCACHE=/tmp/go-cache go test ./... && go build ./...`.
- [x] (2026-03-24 08:16Z) Synced `ANM-188` with the implementation summary, recorded session UUID `019cffb4-e861-7a31-b41f-5e3eeb6c35a9`, and moved the ticket to `Done`.
- [x] (2026-03-24 08:46Z) Ran a live local smoke for mailbox -> dispatcher -> agent-control under `ANM-200` and captured two runtime failures instead of a clean pass.
- [x] (2026-03-24 09:04Z) Fixed the MCP lock bugs and the Codex CLI invocation mismatch, then reran the live smoke successfully through mailbox -> dispatcher -> `agent_run_events`.

## Surprises & Discoveries

- Observation: The current mailbox MCP work intentionally stopped at synchronous storage-backed operations and deferred watcher/broadcast semantics.
  Evidence: `plans/mcp-collab-mailbox-execplan.md` and `plans/mailbox-mcp-mvp-execplan.md` both explicitly defer watcher/broadcast work.

- Observation: The repo already treats Temporal as part of the local dev stack, but `services/agent-control-api` does not yet have any Temporal integration surface.
  Evidence: `scripts/dev-stack` starts `temporal`, while `rg -n "temporal" services/agent-control-api -S` returns no matches.

- Observation: A mailbox dispatcher loop and routing-oriented MCP store methods already existed in the worktree, but they were not yet wired into the running service or covered by focused tests.
  Evidence: `services/agent-control-api/internal/dispatch/dispatcher.go` existed before this slice, while `cmd/api/main.go` did not start it and `services/agent-control-api/internal/dispatch/` had no tests.

## Decision Log

- Decision: Keep agent collaboration as MCP tools and add routed push delivery as a control-plane capability behind those tools, not instead of them.
  Rationale: The user explicitly wants agents to collaborate through shared tools while still allowing the control plane to inject or route messages into active work.
  Date/Author: 2026-03-18 / Codex

- Decision: Treat `/mailbox/events` as an observer stream only and introduce a dispatcher service for worker delivery.
  Rationale: Workers should not consume the watcher stream directly. The dispatcher can own ordering, resume, and push semantics while UIs/Slack/debugging continue to use the watcher as a read-only surface.
  Date/Author: 2026-03-24 / Codex

- Decision: Model durable interrupt/broadcast delivery around a workflow-compatible signal abstraction.
  Rationale: Push events should be deterministic and durable; Temporal-style signaling is the right execution boundary even if the first local slice uses a simpler in-process implementation underneath.
  Date/Author: 2026-03-18 / Codex

- Decision: The first implementation slice will target active local runs and mailbox-backed agent routing, using explicit agent registration metadata plus dispatcher replay state rather than a full external queue/backbone.
  Rationale: This keeps the slice testable and shippable without blocking on Temporal, Kafka, or a separate service registry while still establishing the dispatcher boundary.
  Date/Author: 2026-03-24 / Codex

## Outcomes & Retrospective

Completed outcomes so far:

- `ANM-188` now tracks the routed push-event slice separately from the earlier control-plane bootstrap work.
- This plan captures the durable-signal architecture shift instead of folding it awkwardly into the earlier REST-to-Connect migration plan.
- `services/agent-control-api` now starts a mailbox dispatcher when `MCP_MAILBOX_EVENTS_URL` is configured and forwards routed mailbox deliveries through the same event path as `PushRunEvent`.
- `services/mcp` now exposes `collab_set_agent_focus`, `collab_list_agents`, and `collab_route_message`, with tests covering routing metadata needed by the dispatcher.
- Live smoke findings:
  - Initial rerun exposed two runtime defects: leaked mutexes in `services/mcp/internal/tools/collab.go` and a bad Codex invocation order in `services/agent-control-api/internal/runner/runtime.go`.
  - After fixing those issues, the live rerun succeeded:
    - created run `bb7415d7-c6d0-4bcc-92f6-e977b355ad78`
    - persisted mailbox channel `03e1b190783e0c1b44fd2f67d464a6dd` and message `d99f345b782bda9d32a1dc55b0c68786`
    - dispatcher logged `dispatched mailbox message`
    - `agent_run_events` contains sequence `1` with `INTERRUPT_AND_REPLAN` and mailbox metadata for the run

Remaining gaps:

- `services/agent-control-api` still lacks any Temporal integration or workflow signal surface.
- Live watcher/broadcast fan-out beyond direct MCP pulls is still separate follow-up work (`ANM-195` remains the slow-subscriber/backpressure ticket).

## Context And Orientation

Primary files and modules in this scope:

- `services/agent-control-api/**`
- `services/mcp/internal/tools/**`
- `plans/mcp-collab-mailbox-execplan.md`
- `plans/mailbox-mcp-mvp-execplan.md`
- `scripts/dev-stack`

## Plan Of Work

1. Add dispatcher state and delivery logic to `services/agent-control-api`.
2. Reuse the existing push-event contract to forward mailbox deliveries into active runs.
3. Extend mailbox MCP tools with agent registration plus route/broadcast operations.
4. Validate the local mailbox-to-worker slice and ticket the remaining Temporal/backbone work explicitly.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Audit the current mailbox and control-plane models:
   - `rg -n "mail_|collab_|WatchRun|CreateRun|ListRuns" services/mcp services/agent-control-api -S`
2. Run focused Go tests after each slice:
   - `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`
   - `cd services/agent-control-api && GOCACHE=/tmp/go-cache go test ./...`
3. Smoke-check dispatcher delivery once the background loop is wired:
   - start `services/mcp`
   - start `services/agent-control-api`
   - register agent/run routing metadata
   - route a mailbox message and confirm the worker receives a pushed control event
3. Re-scan for follow-up gaps:
   - `rg -n "TODO|FIXME|interrupt|broadcast|route|fuzzy|signal" services/mcp services/agent-control-api plans -S`

## Validation And Acceptance

Acceptance criteria:

- Active runs can receive a mailbox-originated routed push event through a stable dispatcher path.
- MCP exposes enough registration/routing tools to target an agent or a fuzzy-matched set of agents without requiring callers to know run IDs directly.
- The first slice is usage-focused and does not require callers to consume the watcher stream as a worker protocol.

Validation completed:

- `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`
- `cd services/mcp && GOCACHE=/tmp/go-cache go build ./...`
- `cd services/agent-control-api && GOCACHE=/tmp/go-cache go test ./...`
- `cd services/agent-control-api && GOCACHE=/tmp/go-cache go build ./...`
- Live local smoke across `services/mcp` and `services/agent-control-api` with dispatcher delivery confirmed in `agent_run_events`

## Idempotence And Recovery

The first slice should be safe to rerun locally: mailbox state is persisted separately from control-plane storage, and control-plane schema changes should be additive. If the event model changes during implementation, update the store migration path rather than replacing existing data structures destructively.

## Artifacts And Notes

- Active implementation ticket: `ANM-188`
- Earlier mailbox plans already cover the synchronous collaboration baseline and should not be duplicated here.
- Ticket reconciliation: the live smoke reused existing runtime ticket `ANM-193` for the Codex CLI mismatch, closed `ANM-203` after fixing the MCP lock/runtime hang, kept `ANM-204` open for the flat `mcp` CLI structured-args gap, and kept existing follow-ups in place (`ANM-180`, `ANM-185`, and `ANM-195`).

## Interfaces And Dependencies

- Existing mailbox MCP store and registry in `services/mcp`
- Existing run persistence and local runner plumbing in `services/agent-control-api`
- New dispatcher loop between mailbox deliveries and `PushRunEvent`
- Future Temporal signal integration for durable workflow delivery
