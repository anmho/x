# Agent Operating Guide

This repository uses two planning layers:

- `PLANS.md`: the canonical specification for writing and maintaining execution plans ("ExecPlans").
- `plans/*.md`: concrete, living ExecPlan documents for specific initiatives.

If you are creating or updating an ExecPlan, read `PLANS.md` completely before editing any file under `plans/`.

## Repository Orientation

Project X is both a platform and a monorepo. High-level directories:

- `apps/`: deployable user-facing apps.
- `services/`: deployable APIs, workers, and backend services.
- `agents/`: agent modules and prompts.
- `packages/`: shared libraries and SDKs.
- `scripts/`: local automation and verification entrypoints.
- `platform-cli/`: source for the platform CLI.

## Platform CLI First Workflow (Required)

Use the platform entrypoint for runtime validation and stack control.

1. From repo root, prefer `./platform start`, `./platform status`, `./platform stop`, and `./platform logs <service>` for day-to-day environment checks.
2. Use `scripts/dev-stack <start|status|stop|logs>` only when you specifically need the dev-stack supervisor behavior.
3. Use service-local stack scripts (for example, `services/omnichannel/scripts/stack.sh`) only when intentionally working inside that single service stack.
4. Treat `npm run build` and similar package commands as compile validation only; do not present them as stack/runtime validation.
5. In status updates and final summaries, explicitly name which stack command was used (`./platform ...`, `scripts/dev-stack ...`, or service-local stack script).

Reference runbook: `docs/runbooks/platform-cli-workflow.md`.

## Required Workflow For Planning Work

1. Start by checking whether a relevant plan already exists in `plans/`.
2. If no plan exists, create one using the section structure mandated by `PLANS.md`.
3. Keep plan sections current as work progresses, especially:
   - `Progress`
   - `Surprises & Discoveries`
   - `Decision Log`
   - `Outcomes & Retrospective`
4. At each stopping point, update progress checkboxes with timestamps.
5. Ensure every plan is self-contained for a novice contributor.

## Linear MCP Ticketing (Required)

1. For every user-requested change and every autonomous change, create a Linear ticket using the Linear MCP tools before execution begins, including policy/docs changes such as `AGENTS.md` updates.
2. **Prioritize tickets** when selecting work: prefer `major` (production-risk, security, blockers) over `medium` (correctness/reliability gaps) over `minor` (polish, refactors). When multiple tickets have the same category, prefer those that unblock other work or align with the active ExecPlan.
3. Include additional context in each Linear ticket: request summary, scope, affected areas/files, validation plan, and risks/dependencies.
4. Use the `Anmho` Linear team/key as the primary team so issue identifiers follow `ANM-<number>` for new work in this repository.
5. If `Anmho` is unavailable in MCP, stop and request workspace/team-key correction before creating new tickets.
6. For every active ExecPlan, create or link a corresponding Linear ticket and add the issue ID/link in the plan before executing milestones.
7. Before executing milestones, verify the Linear ticket context and status match the current plan state.
8. When taking a ticket off `Backlog`, immediately set it to `In Progress` before starting implementation work.
9. Synchronize ticket status throughout execution (for example: backlog, in progress, blocked, done) whenever plan progress changes.
10. At completion, ensure the Linear ticket status, plan `Progress`, and `Outcomes & Retrospective` are all aligned.

## Required Workflow For Implementation Work

1. Choose the active ExecPlan in `plans/`.
2. Execute milestones in order; do not skip validation.
3. Record design changes and rationale directly in the active ExecPlan.
4. Keep plan and code in sync in the same change set when possible.

## Post-Change Rescan & Follow-Up Ticketing (Required)

1. After completing every change, rescan the codebase for additional bugs, follow-up work, and feature ideas before closing out the task.
2. Create one Linear TODO/follow-up ticket per identified item using Linear MCP, and do not batch multiple unrelated items into one ticket.
3. Categorize every follow-up ticket as `major`, `medium`, or `minor` and include the category in the ticket title/body.
4. Use `major` for production-risk defects, security issues, or blockers; `medium` for important non-blocking correctness/reliability gaps; `minor` for polish, low-risk refactors, and nice-to-have enhancements.
5. Include enough context in each ticket to execute later: observed issue/opportunity, affected files/areas, expected outcome, and validation notes.
6. Treat policy/docs changes (including `AGENTS.md` updates) the same way: complete the post-change rescan and create categorized follow-up tickets for all identified items.

## Definition Of Done For Any ExecPlan

An ExecPlan is complete only when:

- The promised user-visible behavior can be demonstrated end-to-end.
- Validation commands are documented and pass.
- The plan's living sections reflect final reality.
- Remaining gaps are explicitly listed in `Outcomes & Retrospective`.

## UI Theme Mode (Dark/Light)

Frontend work must support both light and dark mode via system preference by default, with optional user override from quick settings.

1. Default theme source is `prefers-color-scheme`; quick settings may override via `data-theme` (`light` or `dark`) and `localStorage` (`console:theme-mode`).
2. Use semantic theme tokens in each frontend `app/globals.css` and shared shell helpers (`app-shell`, `app-shell-bg`, `app-shell-bg-95`) instead of hardcoded hex backgrounds.
3. Prefer token-backed utilities (`bg-zinc-*`, `text-zinc-*`, `border-zinc-*`) over literal dark/light color values in components.
4. Keep `apps/cloud-console` and `services/omnichannel/frontend` theme behavior aligned unless a plan explicitly documents divergence.
5. Quick settings UX should expose `Light`, `Dark`, and `System` in the avatar/account panel (Vercel-style segmented control).
6. When editing theme behavior, validate with frontend builds and visually verify both system light and dark modes.

## Commit Discipline (Required)

All implementation work must be committed in logical slices, not as one monolithic commit.

1. Create one commit per logical unit of change (for example: policy/docs, routing, schema/types, provider adapter, reconciler, API handlers, UI wiring).
2. Keep each commit independently reviewable and testable.
3. Never combine unrelated subsystems in the same commit.
4. Include validation evidence in the active ExecPlan as each slice lands.
5. Whenever making a change, commit all associated files required for that logical unit so code, docs, tests, and plan updates stay in sync, including `AGENTS.md` when policy/instruction changes are made.
6. Use Conventional Commits for every commit message with a descriptive summary (for example: `feat(api): add provider status endpoint`).
7. For every task, ensure an ExecPlan exists at `plans/<plan>.md`, add it to version control, and include that plan file in the relevant logical commit(s).

## Mistake Logging (Required)

Every mistake discovered during implementation must be recorded immediately in `docs/agent-mistakes.md`.

For each entry include:

1. Timestamp (UTC)
2. What happened
3. Root cause
4. Preventive rule/check added
5. Verification that the prevention is in place

Do not defer logging mistakes until the end of a task.
