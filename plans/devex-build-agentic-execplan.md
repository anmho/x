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
| 3 | Unify Node dependency strategy | Pending | Multiple lockfiles; document or consolidate |
| 4 | Expand cleanup tooling | **Done** | `scripts/clean --full`, `make clean`, `make clean-full` exist |
| 5 | Repo-level .gitignore | **Done** | Root `.gitignore` exists and is comprehensive |
| 6 | Fix docs and scaffolding drift | Pending | `scripts/README` lists `new:ios-app` (should be `mobile-app`) |
| 7 | Complete Access API OpenAPI schema | **Done** | `/v1/policy/check` response schema added |
| 8 | Add baseline automated tests | Pending | access-api, omnichannel |
| 9 | Replace mock API-key console page | Pending | Real backend integration |

## Progress

- [ ] (2026-03-08) ExecPlan created; subagents launched for tickets 3, 6, 7
- [ ] ConnectRPC SDK + Nx (ticket 1) — cloud agent or dedicated subagent
- [ ] Unify Node dependency strategy (ticket 3)
- [ ] Fix docs and scaffolding drift (ticket 6)
- [x] (2026-03-08) Complete Access API OpenAPI schema (ticket 7)
- [ ] Add baseline tests (ticket 8)
- [ ] Replace mock API-key page (ticket 9)

## Surprises & Discoveries

(To be filled as work proceeds)

## Decision Log

(To be filled as work proceeds)

## Outcomes & Retrospective

(To be filled at completion)
