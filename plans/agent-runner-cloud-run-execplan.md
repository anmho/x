# Agent Runner / Agent Control Execution Plan

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-41` ([issue link](https://linear.app/anmho/issue/ANM-41/major-implement-missing-servicesagent-control-api-runapproval-api)) and `ANM-171` ([issue link](https://linear.app/anmho/issue/ANM-171/medium-sync-mintlify-docs-with-agent-control-connectrpc-control-plane)).

## Purpose / Big Picture

Build an agent execution control plane that is protobuf-first and ConnectRPC-based, persists run state in Postgres, supports local and Cloud Run-backed execution, and gives the Cloud Console a generated SDK path instead of a hand-written REST client. The same initiative also needs Mintlify docs that explain the contract, local endpoint configuration, and the default-remote-plus-local-override SDK workflow.

## Progress

- [x] (2026-03-18 06:50Z) Moved `ANM-41` to `In Progress`, assigned it to `me`, and recorded Codex session UUID `019cffb4-bcfb-7eb0-a4b4-09047c0282d5`.
- [x] (2026-03-18 06:58Z) Added runtime/provider modeling to `services/agent-control-api` so runs can target `claude` or `codex` with sandbox and approval policy config.
- [x] (2026-03-18 07:04Z) Added `services/agent-control-api/proto/agentcontrol/v1/agent_control.proto` and generated Go/TypeScript Connect stubs.
- [x] (2026-03-18 07:08Z) Began migrating `services/agent-control-api` from ad hoc HTTP routes to `agentcontrol.v1.AgentControlService`.
- [x] (2026-03-18 07:14Z) Created `ANM-171` for Mintlify docs sync and moved it to `In Progress`.
- [x] (2026-03-18 07:15Z) Added Mintlify `agent-control` page and synced public/internal workflow docs plus workflow status with the ConnectRPC/local-endpoint model.
- [x] (2026-03-18 07:15Z) Validated Mintlify config with `node scripts/ci/verify_docs_config.mjs`.
- [ ] Finish the ConnectRPC server migration cleanly enough to compile and run `go test ./...` in `services/agent-control-api`.
- [ ] Replace the Cloud Console fetch helpers with the generated Connect client while preserving the local endpoint env contract.
- [ ] Reconcile SDK consumption so published remote BSR SDKs are the default and local generated output is development-only.
- [ ] Run focused post-change audit and create follow-up tickets for any distinct remaining gaps.

## Surprises & Discoveries

- Observation: The existing repo proto workflow is only formalized for omnichannel; `scripts/sdk.sh` and `scripts/sdk-core.sh` are hard-coded to `services/omnichannel/backend/proto`.
  Evidence: `scripts/sdk-core.sh` sets `PROTO_DIR="$ROOT_DIR/services/omnichannel/backend/proto"`.

- Observation: The Mintlify docs surface is small and currently workflow-oriented, so syncing agent-control docs is best done by adding one focused page plus updates to the existing public/internal workflow pages instead of creating a parallel docs section.
  Evidence: `docs/mintlify/docs.json` contains only `index`, `workflow-status`, `workflows-public`, and `workflows-internal` before this update.

- Observation: The original agent-control implementation drifted toward hand-written HTTP/JSON and frontend `fetch` helpers before reconciling with the repo’s ConnectRPC contract pattern.
  Evidence: `docs/agent-mistakes.md` entry added on 2026-03-18 for the REST-first mistake.

## Decision Log

- Decision: Treat `services/agent-control-api` as a protobuf-first ConnectRPC control plane, not a REST service.
  Rationale: This repo already uses ConnectRPC generation patterns, and the user explicitly clarified that this should be an RPC control plane.
  Date/Author: 2026-03-18 / Codex

- Decision: Model live output as a server-streaming RPC (`WatchRun`) instead of an SSE-only endpoint.
  Rationale: Server-streaming is the idiomatic protobuf/Connect shape for incremental run events and avoids parallel transport contracts.
  Date/Author: 2026-03-18 / Codex

- Decision: Default SDK consumption should be published remote BSR packages with local-only alias/override for unpublished schema work.
  Rationale: This avoids committing generated SDK code as canonical source while preserving fast local iteration.
  Date/Author: 2026-03-18 / Codex

- Decision: Keep local control-plane endpoint selection runtime-configurable via `NEXT_PUBLIC_AGENT_CONTROL_URL`.
  Rationale: SDK source selection and service endpoint selection are different concerns and should not be conflated.
  Date/Author: 2026-03-18 / Codex

## Outcomes & Retrospective

Completed outcomes so far:

- `services/agent-control-api` now has an explicit proto module at `services/agent-control-api/proto`.
- Go Connect stubs and TypeScript Connect stubs now generate from the agent-control proto contract.
- Mintlify docs now include an agent-control page and local endpoint guidance for the Cloud Console.
- The execution plan now reflects the actual ConnectRPC direction instead of the stale REST-only narrative.

Remaining gaps:

- The ConnectRPC server migration is in progress and still needs a clean compile/test pass.
- The Cloud Console client still needs final alignment to the generated SDK and the remote-by-default/local-override SDK consumption model.
- Service-local Buf publish/consume workflow is not yet integrated into the repo’s broader SDK tooling the way omnichannel is.
- Post-change audit and follow-up ticket reconciliation are still pending.

## Context And Orientation

Primary files and modules in this scope:

- `services/agent-control-api/proto/**`
- `services/agent-control-api/internal/**`
- `apps/cloud-console/app/agents/page.tsx`
- `apps/cloud-console/lib/agent-runs.ts`
- `apps/agent-runner/**`
- `docs/mintlify/**`

## Plan Of Work

1. Finalize the `agentcontrol.v1` protobuf contract and generated stub usage.
2. Complete the Go ConnectRPC service implementation and keep local/cloud execution behavior intact.
3. Move the Cloud Console agent page to the generated Connect client.
4. Align docs and SDK consumption guidance to remote-by-default BSR usage with local-only overrides.
5. Validate, rescan, and reconcile follow-up work into Linear before closing.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Generate agent-control stubs:
   - `cd services/agent-control-api/proto && buf generate --template buf.gen.server.yaml`
   - `cd services/agent-control-api/proto && buf generate --template buf.gen.client.yaml`
2. Compile and test the Go service:
   - `cd services/agent-control-api && go test ./...`
3. Validate docs config:
   - `node scripts/ci/verify_docs_config.mjs`
4. Re-scan for stale REST wording and missing agent-control docs:
   - `rg -n 'agentcontrol\\.v1|NEXT_PUBLIC_AGENT_CONTROL_URL|REST|agent-control-api/proto|ConnectRPC' docs/mintlify -S`

## Validation And Acceptance

Acceptance criteria:

- `agentcontrol.v1.AgentControlService` is the public contract for agent control.
- The Cloud Console consumes the control plane through generated Connect SDK bindings or a thin wrapper over them, not a hand-written REST contract.
- Mintlify docs describe the ConnectRPC contract, local endpoint env, and remote-default/local-override SDK model.

Validation executed so far:

- `buf generate --template buf.gen.server.yaml` from `services/agent-control-api/proto`
- `buf generate --template buf.gen.client.yaml` from `services/agent-control-api/proto`
- `node scripts/ci/verify_docs_config.mjs`
- `rg -n 'agentcontrol\\.v1|NEXT_PUBLIC_AGENT_CONTROL_URL|REST|agent-control-api/proto|local-only alias|Buf Registry|ConnectRPC' docs/mintlify -S`

Validation still required:

- `cd services/agent-control-api && go test ./...`
- Focused Cloud Console build/lint validation once the generated client wiring is complete

## Idempotence And Recovery

The proto generation steps are safe to rerun. If generated stubs drift, regenerate from `services/agent-control-api/proto` and revalidate the Connect service compile/test path plus Mintlify config verification.

## Artifacts And Notes

- Active implementation ticket: `ANM-41`
- Docs sync ticket: `ANM-171`
- Mistake log entry recorded: `docs/agent-mistakes.md` at `2026-03-18T07:16:12Z`

## Interfaces And Dependencies

- ConnectRPC Go runtime and generated handlers for `agentcontrol.v1.AgentControlService`
- Buf generation pipeline for service-local proto stubs
- Cloud Console runtime config via `NEXT_PUBLIC_AGENT_CONTROL_URL`
- Local runtime providers (`claude`, `codex`) and Cloud Run job dispatch path
