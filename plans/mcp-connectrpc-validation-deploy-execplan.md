# Validate and align MCP ConnectRPC service deployment

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-163` ([issue link](https://linear.app/anmho/issue/ANM-163/medium-validate-mcp-service-end-to-end-and-align-deployment-with-repo)).

## Purpose / Big Picture

Verify that the new `services/mcp` ConnectRPC gateway and CLI actually work end to end, repair any broken Nx/local wiring, and align deployment with the repository's existing declarative platform workflow. The goal is to avoid leaving MCP as a partially wired local prototype or introducing a second deployment control plane when the repo already has a preferred pattern.

## Progress

- [x] (2026-03-18 06:50Z) Created `ANM-163`, set it to `In Progress`, assigned it to `me`, and recorded the active Codex session UUID.
- [x] (2026-03-18 06:51Z) Inspected the current `services/mcp` server, CLI, Dockerfile, Nx project, platform CLI entrypoints, and deploy-preflight script.
- [x] (2026-03-18 06:51Z) Confirmed `services/mcp/project.json` still writes binaries to `services/bin`, while repo-local tooling expects root-level `bin/`.
- [x] (2026-03-18 07:04Z) Reproduced the MCP runtime and CLI flows with `npx nx run mcp:generate-proto`, `npx nx run mcp:test`, `npx nx run mcp:build`, `npx nx run mcp:build-cli`, `./scripts/deploy-preflight mcp`, `./platform control-plane plan --project mcp`, and `./platform deploy --project mcp --dry-run`.
- [x] (2026-03-18 07:04Z) Patched `services/mcp/project.json` so Nx build outputs land in repo-root `bin/` and fixed the `dev` target's `MCP_ROOT` to point at the actual repo root.
- [x] (2026-03-18 07:05Z) Added `services/mcp/README.md`, copied `platform.controlplane.json` into the Docker image, and routed `./platform mcp ...` through the repo wrapper so the root entrypoint works without rebuilding the broken full `platform-cli`.
- [x] (2026-03-18 07:05Z) Validated authenticated `./platform mcp ... tools list` and `./platform mcp ... tools call gcp_configured_projects` against a live local server using an auto-generated key.
- [x] (2026-03-18 07:04Z) Re-scanned for follow-up issues, reconciled ticket coverage, and created `ANM-167` for the deployment-safe auth gap.
- [x] (2026-03-18 07:06Z) Reconciled the existing broader platform CLI source-drift blocker with `ANM-153` instead of opening a duplicate ticket.
- [x] (2026-03-18 07:21Z) Revalidated the clean PR branch directly from `services/mcp` with `buf generate`, `go test ./...`, `go build -o ../../bin/mcp-server ./cmd/server`, `go build -o ../../bin/mcp ./cmd/mcp`, `make build-mcp`, `GET /health`, authenticated `./bin/mcp ... tools list`, authenticated `./platform mcp ... tools list`, and authenticated JSON-RPC `POST /mcp`.
- [x] (2026-03-18 07:21Z) Confirmed the clean branch still inherits two pre-existing base-branch gaps: `./scripts/deploy-preflight mcp` is absent and is reconciled to existing deploy-workflow ticket `ANM-158`, while `./platform control-plane plan --project mcp` / `./platform deploy --project mcp --dry-run` fail under the broader `platform-cli` compile drift already tracked in `ANM-153`.
- [x] (2026-03-24 09:55Z) Rebased the surviving MCP branch onto current `main` after the CI stack merged and folded in the outstanding PR review blockers.
- [x] (2026-03-24 09:58Z) Replaced secret-printing bootstrap with explicit local key generation, constant-time key comparison, fail-closed auth, repo-root marker fixes, and subprocess exit-code propagation.
- [x] (2026-03-25 00:22Z) Removed the vestigial TypeScript MCP workspace and server stubs after the audit showed they imported non-existent files and were not part of the real Go service path.
- [x] (2026-03-25 00:29Z) Revalidated the hardened Go path with `go test ./...`, `./platform mcp keys generate --local`, and a no-keys startup failure check for `cmd/server`.

## Surprises & Discoveries

- Observation: the current Nx `mcp:build` and `mcp:build-cli` targets build successfully but emit into `services/bin`, not the repo-root `bin/` directory used by `platform-cli/mcp.go`.
  Evidence: `services/mcp/project.json` uses `go build -o ../bin/...`; `platform-cli/mcp.go` resolves `bin/mcp` relative to the platform executable.

- Observation: the earlier failed smoke test was caused by launching `go run ./services/mcp/cmd/server` from the repo root instead of running inside the `services/mcp` module.
  Evidence: the previous terminal transcript showed GOPATH-style import failures for `connectrpc.com/connect` and `github.com/anmhela/x/mcp/...`, which occur when the module root is not active.

- Observation: this repo already has a declarative deployment/control-plane system for Cloud Run-backed services.
  Evidence: `infra/platform/declarative_spec.py`, generated `platform.controlplane.json`, and `platform-cli/control_plane_ops.go` declare and reconcile `gcp-cloud-run` deployments.

- Observation: the MCP server does not honor `MCP_API_KEYS=testkey`; it only authenticates keys present in the local key store and auto-generates one on first startup.
  Evidence: the local smoke test failed with `invalid API key` until the generated key from `~/.x-mcp/keys.json` was used; `services/mcp/internal/keys/store.go` only validates keys loaded from disk.

- Observation: sandboxed local processes could not complete the MCP smoke test because binding and connecting to the localhost server required unsandboxed execution.
  Evidence: sandboxed `go run ./cmd/server` failed with `bind: operation not permitted`, and sandboxed client calls failed with `connect: operation not permitted`; the same commands worked once run outside the sandbox.

- Observation: Docker validation is currently blocked by the local environment because the Docker daemon is not running.
  Evidence: both `docker build -f services/mcp/Dockerfile -t x-mcp .` and `docker ps` failed with `Cannot connect to the Docker daemon at unix:///Users/andrewho/.docker/run/docker.sock`.

- Observation: the clean branch cut from `origin/main` does not currently provide `scripts/deploy-preflight`, and the root `./platform` build path still hits the separately tracked `platform-cli` compile drift.
  Evidence: `./scripts/deploy-preflight mcp` returned `no such file or directory`, while both `./platform control-plane plan --project mcp` and `./platform deploy --project mcp --dry-run` failed with `undefined: runStack` and related symbol errors already reconciled to `ANM-153`.

- Observation: the branch still carried a vestigial TypeScript MCP workspace and Bun server path that imported `../../mcp/common/*`, but those files do not exist anywhere in the repository.
  Evidence: `services/mcp/tools.ts` imported `../../mcp/common/platform-projects.ts` and `../../mcp/common/shell.ts`, while the root `mcp/` directory only contained a broken `package.json`.

## Decision Log

- Decision: start by validating and repairing the existing declarative platform path instead of introducing ArgoCD immediately.
  Rationale: the repo already materializes deployment intent from `infra/platform/declarative_spec.py` into `platform.controlplane.json`, and `./platform deploy --project <name>` is documented as the standard workflow. Adding ArgoCD before proving the current path insufficient would create an extra control plane.
  Date/Author: 2026-03-18 / Codex

- Decision: split the deployment-safe auth/bootstrap problem into follow-up ticket `ANM-167` instead of inflating this validation slice into a full secret-management implementation.
  Rationale: `ANM-163` successfully validated the existing local runtime and declarative deploy path, but the remaining gap is a broader production auth design question that deserves its own implementation ticket and reviewable change set.
  Date/Author: 2026-03-18 / Codex

- Decision: route `./platform mcp ...` through the repo wrapper now instead of trying to rebuild the full `platform-cli` package in this slice.
  Rationale: the broader `platform-cli` source tree is already broken and tracked in `ANM-153`; the wrapper route restores the MCP user path without broadening this task into a repo-wide CLI reconstruction.
  Date/Author: 2026-03-18 / Codex

- Decision: replace implicit server-side key generation with explicit bootstrap via `MCP_API_KEYS` or `./bin/mcp keys generate --local`.
  Rationale: the PR must not leak admin credentials into logs, and the CLI can provide a reviewable bootstrap path without relying on a live unauthenticated server.
  Date/Author: 2026-03-24 / Codex

- Decision: remove the vestigial TypeScript MCP workspace and server helpers from this PR instead of trying to rehabilitate them alongside the Go control-plane path.
  Rationale: the actual productized path in this branch is the Go service plus Go CLI. Keeping an unreferenced Bun/TypeScript implementation that already imports missing modules would make the branch look broader than it really is and leave dead, broken code in the tree.
  Date/Author: 2026-03-25 / Codex

## Outcomes & Retrospective

Completed outcomes:

- `services/mcp/project.json` now writes `mcp` and `mcp-server` into repo-root `bin/`, matching the rest of the repo tooling.
- The rebased PR branch now requires explicit auth bootstrap, supports env-backed or file-backed key validation, and preserves constant-time key comparison in the live Go auth path.
- `./bin/mcp keys generate --local` now provides a local bootstrap path without emitting credentials into service logs.
- The rebased PR no longer carries the dead TypeScript MCP workspace; the branch now reflects a single real implementation path through the Go server, Go CLI, and platform wiring.
- Additional hardening validation passed for:
  - `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`
  - `./platform mcp keys generate --local --store /tmp/x-pr7-rebase-mcp-keys.json`
  - `cd services/mcp && MCP_KEYS_FILE=/tmp/x-pr7-rebase-no-keys.json GOCACHE=/tmp/go-cache go run ./cmd/server` (expected fail-fast with no configured keys)
- Local smoke tests passed for:
  - `GET /health`
  - authenticated ConnectRPC via `./bin/mcp ... tools list`
  - authenticated ConnectRPC via `./platform mcp ... tools list`
  - authenticated ConnectRPC via `./platform mcp ... tools call gcp_configured_projects`
  - authenticated JSON-RPC via `POST /mcp` with `tools/list`
- Clean PR branch validation also passed for:
  - `cd services/mcp/proto && PATH="$PATH:$HOME/go/bin" buf generate`
  - `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`
  - `cd services/mcp && GOCACHE=/tmp/go-cache go build -o ../../bin/mcp-server ./cmd/server`
  - `cd services/mcp && GOCACHE=/tmp/go-cache go build -o ../../bin/mcp ./cmd/mcp`
  - `make build-mcp`
- Declarative deployment validation passed for:
  - `./scripts/deploy-preflight mcp`
  - `./platform control-plane plan --project mcp`
  - `./platform deploy --project mcp --dry-run`
- Docker image wiring now includes `platform.controlplane.json` and defaults `MCP_ROOT=/app`, making config-backed tools available in containerized environments.
- Follow-up deployment auth gap captured in `ANM-167`.

Remaining gaps:

- The current MCP auth model is still file-local and auto-generated, which is not yet a deployment-safe Cloud Run operator story; tracked in `ANM-167`.
- The immediate PR review blockers around timing-safe auth, repo-root detection, and exit-code propagation are addressed on the rebased branch, but broader Cloud Run operator secret management is still tracked in `ANM-167`.
- The clean PR branch cut from `origin/main` still lacks `scripts/deploy-preflight`, so repo-standard preflight validation cannot run there until that base-branch tooling is restored.
- The full `platform-cli` Go source tree still does not compile cleanly; that broader source-drift issue is already tracked separately in `ANM-153`.
- Docker build/run was not validated because the local Docker daemon is unavailable in this environment.
- Remote deployment execution has not been attempted; this slice validated the repo-aligned dry-run path only.

## Context And Orientation

Relevant files for this task:

- `services/mcp/project.json`
- `services/mcp/cmd/server/main.go`
- `services/mcp/cmd/mcp/main.go`
- `services/mcp/Dockerfile`
- `platform-cli/main.go`
- `platform-cli/mcp.go`
- `platform-cli/README.md`
- `infra/platform/declarative_spec.py`
- `platform.controlplane.json`
- `scripts/deploy-preflight`
- `scripts/verify`

## Plan Of Work

1. Reproduce the current Nx, local CLI, and server flows to identify real failures instead of inferred ones.
2. Patch pathing or runtime defects in `services/mcp`, `Makefile`, and `platform-cli` until local smoke tests pass.
3. Align MCP deployment with the declarative platform workflow and update any missing docs or validation hooks.
4. Run targeted validation, rescan for follow-up work, and sync the plan plus `ANM-163`.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Re-run MCP validation:
   - `npx nx run mcp:generate-proto`
   - `npx nx run mcp:test`
   - `npx nx run mcp:build`
   - `npx nx run mcp:build-cli`
   - `scripts/deploy-preflight mcp`
2. Smoke-test the server and CLI from the module root:
   - `./bin/mcp keys generate --local`
   - `cd services/mcp && GOCACHE=/tmp/go-cache go run ./cmd/server`
   - read the configured key from `~/.x-mcp/keys.json` (or set `MCP_API_KEYS`)
   - `../../bin/mcp --server http://127.0.0.1:18765 --key <generated-key> tools list`
   - `curl -H "Authorization: Bearer <generated-key>" -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' http://127.0.0.1:18765/mcp`
3. Inspect the deployment path:
   - `./platform control-plane plan --project mcp`
   - `./platform deploy --project mcp --dry-run`
4. Patch documentation and wiring based on the validated path.

## Validation And Acceptance

Acceptance criteria:

- `mcp` Nx targets build and test successfully without writing binaries into the wrong directory.
- The local MCP server responds on `/health`, serves authenticated ConnectRPC calls, and serves authenticated `/mcp` JSON-RPC requests.
- `platform mcp ...` and/or the standalone `mcp` CLI can successfully talk to the local server.
- The repo has one clear deployment path for `mcp`, documented and wired through the existing platform workflow.

Validation commands:

- `npx nx run mcp:generate-proto`
- `npx nx run mcp:test`
- `npx nx run mcp:build`
- `npx nx run mcp:build-cli`
- `scripts/deploy-preflight mcp`
- `./platform control-plane plan --project mcp`
- `./platform deploy --project mcp --dry-run`

## Idempotence And Recovery

Re-running the validation commands should be safe. Rebuilding should overwrite the same `bin/mcp` and `bin/mcp-server` outputs once paths are corrected. If remote deploy commands require unavailable credentials, dry-run evidence and local smoke-test evidence will still keep the deployment path reviewable.

## Artifacts And Notes

- Linear issue: `ANM-163`
- Session UUID: `019cffb4-e861-7a31-b41f-5e3eeb6c35a9`
- Validation continuation session UUID: `019cffb5-3acf-7193-ad5c-8e7965ec70e9`
- Related earlier issues: `ANM-149`, `ANM-150`, `ANM-151`, `ANM-152`
- Follow-up created during validation: `ANM-167`
- Existing broader blocker confirmed during validation: `ANM-153`
- Existing deploy-workflow follow-up reconciled during PR hardening: `ANM-158`

## Interfaces And Dependencies

- Go module in `services/mcp`
- Nx target wiring in `services/mcp/project.json` and `nx.json`
- Platform CLI `mcp`, `deploy`, and `control-plane` commands
- Generated platform configs from `infra/platform/declarative_spec.py`
