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
| 1 | ConnectRPC SDK + Nx | Pending | Major; subagent or cloud agent |
| 2 | Turbopack workspace root | **Done** | `next.config.ts` already has `turbopack.root` |
| 3 | Unify Node dependency strategy | **Done** | Single root lockfile; runbook added |
| 4 | Expand cleanup tooling | **Done** | `scripts/clean --full`, `make clean`, `make clean-full` exist |
| 5 | Repo-level .gitignore | **Done** | Root `.gitignore` exists and is comprehensive |
| 6 | Fix docs and scaffolding drift | **Done** | scripts/README already correct; fixed doc paths, added docs lint |
| 7 | Complete Access API OpenAPI schema | **Done** | `/v1/policy/check` response schema added |
| 8 | Add baseline automated tests | Pending | access-api, omnichannel |
| 9 | Replace mock API-key console page | Pending | Real backend integration |

## Progress

- [ ] (2026-03-08) ExecPlan created; subagents launched for tickets 3, 6, 7
- [ ] ConnectRPC SDK + Nx (ticket 1) — cloud agent or dedicated subagent
- [x] (2026-03-08) Unify Node dependency strategy (ticket 3) — single root lockfile, runbook, verify/deploy-preflight updated
- [x] (2026-03-08) Fix docs and scaffolding drift (ticket 6): corrected `docs/top-three-priorities.md` path to `docs/mintlify/services/access-api.mdx`; scripts/README already had `new:mobile-app`; added docs lint in `scripts/verify docs` (broken path grep + mint.json nav page existence)
- [x] (2026-03-08) Complete Access API OpenAPI schema (ticket 7)
- [ ] Add baseline tests (ticket 8)
- [ ] Replace mock API-key page (ticket 9)

## Surprises & Discoveries

- CI previously referenced `omnichannel/frontend/package-lock.json` (invalid path); correct path was `services/omnichannel/frontend/package-lock.json`. Consolidation to root lockfile removed the need for per-workspace cache paths.
- `mcp` uses Bun for its `check` script; npm installs its deps; both work with the unified root lockfile.

## Decision Log

- **Single root lockfile**: Chose npm workspaces with one `package-lock.json` at repo root over a documented multi-lock policy. Rationale: deterministic installs, no workspace-root ambiguity, simpler CI cache.

## Outcomes & Retrospective

(To be filled at completion)
