# Build Slack and Linear collaboration bridge for mailbox MCP

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-180` ([issue link](https://linear.app/anmho/issue/ANM-180/medium-add-slack-bridge-between-collaboration-mailbox-events-and)) and `ANM-190` ([issue link](https://linear.app/anmho/issue/ANM-190/medium-add-linear-issue-context-bridge-for-collaboration)).

## Purpose / Big Picture

Add a Slack-facing collaboration bridge around the existing mailbox MCP model so agents and humans can collaborate in Slack threads without making Slack the source of truth. Extend the mailbox model with Linear issue-backed channel metadata so channels can be keyed to `ANM-*` issues and carry issue context for Slack routing and agent handoff flows.

## Progress

- [x] (2026-03-24T07:31:23Z) Moved `ANM-180` to `In Progress`, created `ANM-190` for Linear issue-backed collaboration context, and assigned both tickets to the current session owner.
- [x] (2026-03-24T07:36:39Z) Recorded the active Codex session UUID on both Linear tickets before implementation work.
- [x] (2026-03-24T07:59:04Z) Extended mailbox storage and MCP tools with external refs, channel status, and Linear issue metadata.
- [x] (2026-03-24T07:59:04Z) Added Slack ingress plus outbound thread mirroring in `services/mcp` and registered the Slack Events API endpoint.
- [x] (2026-03-24T07:59:04Z) Added focused tests for Linear-backed channel resolution, Slack event normalization, duplicate-event suppression, and Slack thread mirroring.
- [x] (2026-03-24T07:59:04Z) Updated docs/env examples and completed the post-change audit, including follow-up ticket `ANM-198` for the unrelated malformed root lockfile.

## Surprises & Discoveries

- Observation: the repo already has a mailbox MCP surface with durable local storage, but no existing external reference model for Slack threads or Linear issues.
  Evidence: `services/mcp/internal/tools/collab.go` currently stores participants, metadata, and messages only.

- Observation: `services/agent-tools` exists, but it is currently scoped as a sidecar MCP server for agent-runner rather than a standalone Slack or webhook bridge service.
  Evidence: `services/agent-tools/src/index.js` only exposes `run_shell`, `read_file`, and `write_file` over MCP.

- Observation: there is already a separate browser-side Linear integration surface in the repo, so the bridge should reuse Linear issue identifiers and metadata rather than inventing another task namespace.
  Evidence: `apps/linear-ticket-sidepanel` contains an existing Linear GraphQL client flow and issue-oriented UX.

- Observation: repo-level Nx validation is currently blocked by an unrelated malformed root lockfile.
  Evidence: `package-lock.json:51` contains unresolved merge markers and `npx nx run mcp:build` fails during Nx lockfile parsing before project execution starts.

## Decision Log

- Decision: implement the Slack and Linear bridge inside `services/mcp` for this slice rather than introducing a second backend service.
  Rationale: `services/mcp` already owns the mailbox source of truth, auth middleware, and HTTP server. Keeping the first bridge layer co-located reduces duplication and lets tests exercise mailbox plus bridge behavior together.
  Date/Author: 2026-03-24 / Codex

- Decision: keep mailbox channels as canonical state and treat Slack plus Linear as external references attached to a channel.
  Rationale: this preserves runtime-agnostic collaboration semantics and avoids coupling future non-Slack clients to Slack threading rules or Linear APIs.
  Date/Author: 2026-03-24 / Codex

- Decision: use Linear issue identifiers as the preferred task key when present.
  Rationale: issue-backed collaboration threads are easier to audit, route, and recover than ad hoc free-form task keys.
  Date/Author: 2026-03-24 / Codex

## Outcomes & Retrospective

Completed outcomes:

- Mailbox channels now support first-class status and external references so Slack threads and Linear issues can point to the same collaboration record.
- MCP tools now support Linear-backed channel creation, external-ref lookup, Slack thread linking, and channel status transitions.
- `services/mcp` now exposes `POST /slack/events` for Slack Events API ingress and mirrors mailbox messages back to linked Slack threads when `SLACK_BOT_TOKEN` is configured.
- Service docs and environment examples now document Slack and Linear bridge configuration.

Remaining gaps:

- Nx project-graph validation remains blocked by the unrelated malformed root lockfile tracked in `ANM-198`.
- Repository commit creation is currently blocked by the same unresolved root lockfile conflict because Git will not commit while `package-lock.json` remains unmerged.
- The current Slack bridge handles Events API ingress and thread replies, but does not yet add slash-command flows or richer Slack approval actions.

## Context And Orientation

Relevant files for this task:

- `services/mcp/cmd/server/main.go`
- `services/mcp/internal/tools/collab.go`
- `services/mcp/internal/tools/registry.go`
- `services/mcp/README.md`
- `services/mcp/.env.example`

## Plan Of Work

1. Extend mailbox storage with external refs, channel status, and issue metadata helpers.
2. Add MCP tool support for issue-backed channel resolution and channel status updates.
3. Add Slack event ingress plus outbound mirroring helpers in the MCP server.
4. Validate the new collaboration flows with focused Go tests and repo-level verification.
5. Re-scan for follow-up gaps, reconcile them against Linear, and finalize plan outcomes.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Update `services/mcp/internal/tools/collab.go` and `registry.go` to support:
   - external refs for Slack threads
   - Linear issue metadata
   - channel status transitions
2. Add Slack and Linear bridge modules plus HTTP handlers under `services/mcp/internal/...`.
3. Update `services/mcp/cmd/server/main.go` to register the new handlers.
4. Run:
   - `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`
   - `cd services/mcp && GOCACHE=/tmp/go-cache go build ./...`
5. Update `services/mcp/README.md` and `.env.example` with bridge configuration and usage notes.

## Validation And Acceptance

Acceptance criteria:

- Mailbox channels can be resolved or created from a Linear issue identifier.
- Mailbox channels can store Slack thread references and collaboration status without losing backward compatibility.
- Slack event payloads normalize into mailbox messages and reuse the same channel when a thread is already linked.
- `services/mcp` tests pass after the bridge changes are added.

Validation to run:

- `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`
- `cd services/mcp && GOCACHE=/tmp/go-cache go build ./...`

Validation executed:

- `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`
- `cd services/mcp && GOCACHE=/tmp/go-cache go build ./...`
- `npx nx run mcp:build` failed before project execution because the root `package-lock.json` contains unresolved merge conflict markers; tracked in `ANM-198`

## Idempotence And Recovery

Mailbox persistence changes must remain backward compatible with existing `collab.json` files. Slack and Linear linkage writes must be safe to replay so duplicate webhook deliveries do not create duplicate channels or duplicate mailbox messages.

## Artifacts And Notes

- Active implementation tickets: `ANM-180`, `ANM-190`
- Follow-up ticket: `ANM-198`
- Session UUID: `019cffde-7bd6-7960-936c-a0f1f5b60887`

## Interfaces And Dependencies

- Existing MCP tool registry and HTTP server in `services/mcp`
- Slack Web API and Events API environment configuration
- Linear GraphQL API key for issue metadata enrichment
