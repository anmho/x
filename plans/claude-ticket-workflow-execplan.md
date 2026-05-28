# Implement the first long-running Claude bootstrap workflow slice

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-194` ([issue link](https://linear.app/anmho/issue/ANM-194/major-implement-first-long-running-claude-ticket-workflow-slice)), with related epic context in `ANM-191` (control plane) and `ANM-192`/`ANM-193` (data plane).

## Purpose / Big Picture

Build the first narrow end-to-end slice for Project X agent execution: submit a bootstrap-oriented run into the control plane, execute it as a long-running Claude-backed workflow, stream progress/output back through the control plane, and support cancellation. This slice deliberately excludes later epics such as rich live attach, mailbox-driven collaboration routing, and multi-provider support, but it must leave room for them without requiring API churn. The immediate follow-on after this slice should be a dispatcher that can turn tickets or any other work source into rich bootstrap payloads, plus webhook-driven resync when those upstream requirements change.

The architecture assumption is now explicit: the control plane stays in Go and continues to use ConnectRPC, while the runtime/data plane should move to TypeScript so the worker can use vendor SDKs directly instead of relying on CLI wrappers or nonexistent Go agent SDKs.

## Progress

- [x] (2026-03-24 07:46Z) Created `ANM-194`, assigned it to `me`, moved it to `In Progress`, and recorded Codex session UUID `019cffb4-bcfb-7eb0-a4b4-09047c0282d5`.
- [x] (2026-03-24 07:47Z) Confirmed the current worktree already contains follow-on MCP/watch work and decided to keep this slice out of `services/mcp` unless a hard dependency appears.
- [x] (2026-03-24 08:27Z) Simplified the control-plane API to a message-first contract with typed resources and runtime configuration, removed public `env` and top-level metadata bags, and regenerated the ConnectRPC SDKs.
- [x] (2026-03-24 08:27Z) Implemented the Bun/TypeScript `apps/agent-runner` worker foundation with ConnectRPC client wiring and Claude-first runtime adapter scaffolding.
- [ ] Validate run creation, watch/progress streaming, and cancellation end to end with a live local server/database environment.

## Surprises & Discoveries

- Observation: `services/agent-control-api` already has a partial ConnectRPC control-plane surface, including generated stubs and a `WatchRun` streaming path, so the first slice can extend existing foundations instead of starting from zero.
  Evidence: generated SDK files exist under `packages/sdk-agent-control/src/gen`, and tests already cover `WatchRun` replay behavior in `services/agent-control-api/internal/api/handler_test.go`.

- Observation: There is already uncommitted collaboration/watcher work under `services/mcp`, but it is not required to prove the first long-running ticket workflow slice.
  Evidence: modified/untracked `services/mcp/internal/tools/*` and `services/mcp/internal/watch/*` are present in the current worktree before this slice begins.

- Observation: Claude and OpenAI agent runtime SDKs are language-specific, with official SDK surfaces documented for TypeScript/Python rather than Go.
  Evidence: Claude Agent SDK docs are published for TypeScript and Python, and OpenAI Agents SDK docs likewise point to Python/TypeScript instead of Go.

- Observation: A generic `metadata` map and client-controlled `env` map made the create contract harder to reason about and leaked implementation details into the API.
  Evidence: The simplification review for this slice concluded that the right public shape is a canonical `message` plus typed `resources` and `runtime`, with process environment remaining an internal adapter detail.

- Observation: The focused Go and Bun test suites pass under the simplified message-first contract, but a real `go run ./cmd/api` smoke test still fails locally because `agent_control` on `127.0.0.1:54322` is not reachable in this session.
  Evidence: `cd services/agent-control-api && GOCACHE=/tmp/go-cache go test ./...` passed, `cd apps/agent-runner && bun test` passed, and `go run ./cmd/api` still exited with `dial tcp 127.0.0.1:54322: connect: connection refused`.

## Decision Log

- Decision: Limit the first slice to Claude as the only runtime provider, while preserving an internal provider-pluggable contract.
  Rationale: This keeps the runtime path simple enough to land a real end-to-end workflow without prematurely standardizing Codex behavior.
  Date/Author: 2026-03-24 / Codex

- Decision: Keep the control plane as the durable source of truth and do not require live attach/editor semantics in this slice.
  Rationale: The user wants those later, but the first shippable capability is "submit, watch, cancel" for a long-running bootstrap-driven run.
  Date/Author: 2026-03-24 / Codex

- Decision: Design the runtime path to work for both Kubernetes and Cloud Run service by avoiding Kubernetes-only execution assumptions.
  Rationale: Host portability is a core requirement for the data-plane architecture.
  Date/Author: 2026-03-24 / Codex

- Decision: Center the execution contract on a single canonical message plus typed resources rather than a nested bootstrap envelope.
  Rationale: The first slice does not need more abstraction than "send this message and these attachments to an agent." Upstream systems can still render tickets, PRs, incidents, and other work items into that same message-first shape.
  Date/Author: 2026-03-24 / Codex

- Decision: Keep the control plane in Go with ConnectRPC, but move the data-plane worker direction to TypeScript so provider-native SDKs can be used directly.
  Rationale: A TypeScript worker is a better fit for Claude/Codex SDK integration than continuing to model the runtime around Go structs plus subprocess CLI execution.
  Date/Author: 2026-03-24 / Codex

## Outcomes & Retrospective

Completed outcomes so far:

- A dedicated implementation ticket now exists for the narrow first slice, separate from the broader push-routing and data-plane epics.
- This plan captures the intentionally reduced scope so implementation does not drift into later attach/collaboration/provider work.
- `agentcontrol.v1` now uses a message-first create/run contract with typed resources and no public `env` bag.
- Go and TypeScript ConnectRPC SDKs were regenerated from the simplified proto.
- `apps/agent-runner` is now a Bun/TypeScript worker scaffold instead of only a shell-entrypoint container.

Remaining gaps:

- The control plane still launches work locally in-process; the leased long-running worker/data-plane protocol is still a follow-on step.
- End-to-end cancellation and streaming validation still need to be demonstrated against a live local server/database environment.
- The eventual MCP operator surface for launching and controlling agents remains a follow-on epic.

## Context And Orientation

Primary files and modules in this scope:

- `services/agent-control-api/**`
- `apps/agent-runner/**`
- `packages/sdk-agent-control/**`
- `plans/agent-runner-cloud-run-execplan.md`
- `plans/agent-push-routing-execplan.md`

## Plan Of Work

1. Extend the control plane with the minimal message-oriented run model and long-running lifecycle APIs.
2. Implement a Claude-first TypeScript runtime worker contract in `apps/agent-runner`.
3. Connect the control plane and worker with a portable long-running execution path.
4. Validate create/watch/cancel behavior and record follow-up epics for attach, routing, multi-provider support, and upstream bootstrap producers such as Linear dispatch/webhook sync.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Inspect current control-plane and runtime foundations:
   - `rg -n "CreateRun|WatchRun|PushRunEvent|RuntimeProvider|claude|codex" services/agent-control-api apps/agent-runner packages/sdk-agent-control -S`
2. Run focused Go validation during implementation:
   - `cd services/agent-control-api && GOCACHE=/tmp/go-cache go test ./...`
3. Run worker/runtime validation once the Claude path is wired:
   - focused runtime tests or smoke validation for `apps/agent-runner`
4. Re-scan for follow-up gaps:
   - `rg -n "TODO|FIXME|attach|session|collab|codex|temporal" services/agent-control-api apps/agent-runner plans -S`

## Validation And Acceptance

Acceptance criteria:

- A caller can submit a message-oriented run to the control plane.
- The run executes as a long-running Claude-backed workflow.
- Progress/output can be observed through the control plane.
- The run can be canceled through the control plane.
- The design remains compatible with later attach/collaboration/provider epics.
- The persisted message/resources run model is suitable for later upstream producers, including a dispatcher that ingests Linear tickets and a webhook sync path that updates agent requirements over time.

Validation still required:

- `cd services/agent-control-api && GOCACHE=/tmp/go-cache go test ./...`
- `cd apps/agent-runner && bun test`
- End-to-end smoke test for create/watch/cancel once the local `agent_control` database is reachable on `127.0.0.1:54322`

## Idempotence And Recovery

The control-plane schema changes for this slice should be additive. If the runtime worker crashes or is unavailable, the control plane must preserve run/event state and surface the run as failed, canceled, or recoverable without losing history.

## Artifacts And Notes

- Active implementation ticket: `ANM-194`
- Related control-plane epic: `ANM-191`
- Related data-plane epic/work item: `ANM-192`, `ANM-193`
- Planned follow-on extension path: upstream dispatchers such as Linear + webhook-based resync built on the same message/resources run model, plus an MCP operator surface for launching and steering agents from Claude/Codex

## Interfaces And Dependencies

- ConnectRPC Go runtime and generated `agentcontrol.v1` SDKs
- Claude-first TypeScript runtime worker under `apps/agent-runner`
- Existing Postgres-backed run persistence in `services/agent-control-api`
- Later Temporal integration remains a follow-on concern unless a minimal dependency is required to land durable run control in this slice
