# DevEx / Build System / Agentic-Friendly Execution Plan

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Follows `PLANS.md` structure.

## Purpose / Big Picture

Improve developer experience, build system consistency, and agent-friendliness across the monorepo. After this work:

- Proto→SDK changes flow locally without BSR publish wait
- Builds are deterministic and workspace-root warnings are resolved
- Docs and scaffolding are aligned with actual paths/commands
- Agents have clear validation signals (tests, OpenAPI contracts)

## Ticket Status (Pre-Execution Audit)

| # | Ticket | Status | Notes |
|---|--------|--------|------|
| 1 | ConnectRPC SDK + Nx | **Done** | Nx task graph, local ES gen, packages/sdk-omnichannel |
| 2 | Turbopack workspace root | **Done** | `next.config.ts` already has `turbopack.root` |
| 3 | Unify Node dependency strategy | **Done** | Single root lockfile; runbook added |
| 4 | Expand cleanup tooling | **Done** | `scripts/clean --full`, `make clean`, `make clean-full` exist |
| 5 | Repo-level .gitignore | **Done** | Root `.gitignore` exists and is comprehensive |
| 6 | Fix docs and scaffolding drift | **Done** | scripts/README already correct; fixed doc paths, added docs lint |
| 7 | Complete Access API OpenAPI schema | **Done** | `/v1/policy/check` response schema added |
| 8 | Add baseline automated tests | **Done** | access-api, omnichannel |
| 9 | Replace mock API-key console page | Pending | Real backend integration |

## Progress

- [x] (2026-03-08) ExecPlan created; subagents launched for tickets 3, 6, 7
- [x] (2026-03-09) ConnectRPC SDK + Nx (ticket 1): Nx task graph (sdk:generate-go, sdk:generate-es, sdk:lint, sdk:publish), buf.gen.client.yaml for local ES generation, packages/sdk-omnichannel, frontends depend on local SDK, sdk.sh generate-es
- [x] (2026-03-08) Unify Node dependency strategy (ticket 3) — single root lockfile, runbook, verify/deploy-preflight updated
- [x] (2026-03-08) Fix docs and scaffolding drift (ticket 6): corrected `docs/top-three-priorities.md` path to `docs/mintlify/services/access-api.mdx`; scripts/README already had `new:mobile-app`; added docs lint in the docs verification path active at the time (broken path grep + mint.json nav page existence)
- [x] (2026-03-08) Complete Access API OpenAPI schema (ticket 7)
- [x] (2026-03-08) Add baseline tests (ticket 8): access-api auth/key lifecycle unit + integration (skip when no DB); omnichannel domain, middleware, handlers, repository tests
- [x] (2026-03-09) Completed proto→publish workflow wiring: added active root workflow at `.github/workflows/publish-connectrpc-sdks.yml`, introduced root `scripts/sdk.sh` Nx entrypoint, and moved Buf operations into `scripts/sdk-core.sh` used by Nx `sdk:*` targets.
- [x] (2026-03-10) Restored missing proto module source (`buf.yaml`, `buf.gen.server.yaml`, `temporal/v1/temporal.proto`) and validated Nx SDK flow (`sdk:lint`, `sdk:generate-es`, `sdk:generate-go`) from root `scripts/sdk.sh`.
- [ ] Replace mock API-key page (ticket 9)

## Surprises & Discoveries

- CI previously referenced `omnichannel/frontend/package-lock.json` (invalid path); correct path was `services/omnichannel/frontend/package-lock.json`. Consolidation to root lockfile removed the need for per-workspace cache paths.
- `mcp` uses Bun for its `check` script; npm installs its deps; both work with the unified root lockfile.
- ConnectRPC publish workflow had been defined under `services/omnichannel/.github/workflows`, which is not the canonical root workflow directory for this monorepo.
- `main` currently has no proto files at `services/omnichannel/backend/proto`, so `buf breaking` must be skipped until a baseline exists on `main`.

## Decision Log

- **Single root lockfile**: Chose npm workspaces with one `package-lock.json` at repo root over a documented multi-lock policy. Rationale: deterministic installs, no workspace-root ambiguity, simpler CI cache.
- **SDK command entrypoint**: Standardize on root `scripts/sdk.sh` for all proto/SDK actions, with Nx handling orchestration and `scripts/sdk-core.sh` handling Buf operations. Rationale: one command surface for humans + CI while preserving low-level control in a single core script.
- **Buf CLI compatibility**: Implement publish/list logic in `scripts/sdk-core.sh` to support current Buf CLI behavior (`registry sdk version`) and only use explicit create/list subcommands when available. Rationale: keep local and CI behavior stable across Buf versions.

## Outcomes & Retrospective

### ConnectRPC SDK + Nx (ticket 1, 2026-03-09)

- **Delivered:** Nx task graph (sdk:generate-go, sdk:generate-es, sdk:lint, sdk:publish), buf.gen.client.yaml for local ES generation, packages/sdk-omnichannel, frontend workspace deps, sdk.sh generate-es, verify:apps runs sdk:generate-es first.
- **Delivered extension (2026-03-09):** root `.github/workflows/publish-connectrpc-sdks.yml` now runs active manual publish via Nx scripts; root `scripts/sdk.sh` is the Nx wrapper (`sdk:*`), and `scripts/sdk-core.sh` centralizes Buf/BSR operations for targets and CI.
- **Delivered extension (2026-03-10):** proto source/config restored under `services/omnichannel/backend/proto`, breaking-check logic fixed for monorepo subdir baseline, and root `scripts/sdk.sh` commands now validate locally (`lint`, `generate-es`, `generate-server`). Registry-facing `list`/`publish-*` now fail only on auth/network in this environment.
- **Note:** services/omnichannel/frontend has nested .git; add @x/sdk-omnichannel and project.json there when integrating. PR template already has Linear section (criterion 3).

### Execution method

All tickets were executed via **Cursor subagents** (mcp_task). Cloud agents were not used; for future runs, launch from [cursor.com/agents](https://cursor.com/agents) or use the Cloud Agents API with `CURSOR_API_KEY` for programmatic execution.
