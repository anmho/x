# Agent Operating Guide

This repository uses two planning layers:

- `plans/*.md`: concrete, living ExecPlan documents for specific initiatives.
- Existing ExecPlans are also the canonical template reference (section order and required living updates).

If you are creating or updating an ExecPlan, read one or more recent plans in `plans/` first and follow the same section structure.

## Repository Orientation

Project X is both a platform and a monorepo. High-level directories:

- `apps/`: deployable user-facing apps.
- `services/`: deployable APIs, workers, and backend services.
- `agents/`: agent modules and prompts.
- `packages/`: shared libraries and SDKs.
- `scripts/`: local automation and verification entrypoints.
- `platform-cli/`: source for the platform CLI.

## Platform CLI Focus (Required)

Use platform-cli for repository scaffolding and console workflow operations.

1. Prefer platform-cli workflow commands for scaffolding and service/resource operations (for example `platform create ...`, `platform config ...`, `platform project ...`, `platform tokens ...`, `platform notifications ...`, `platform control-plane ...`, `platform deploy ...`).
2. For validation evidence, prefer `scripts/verify` and `scripts/deploy-preflight` (or their npm wrappers) instead of local runtime lifecycle checks.
3. Treat local lifecycle helpers (`platform start|status|stop|logs`, `platform stack ...`, `scripts/dev-stack ...`) as optional troubleshooting paths, not default workflow requirements.
4. In status updates and final summaries, explicitly name the scaffolding/workflow or verification command that was run.

Reference runbook: `docs/runbooks/platform-cli-workflow.md`.

## Required Workflow For Planning Work

1. Start by checking whether a relevant plan already exists in `plans/`.
2. If no plan exists, create one using the section structure used by the current ExecPlans in `plans/`.
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
6. Use `X Platform` as the default Linear project for repository work; if the project is missing, create or restore it before starting implementation.
7. For every active ExecPlan, create or link a corresponding Linear ticket and add the issue ID/link in the plan before executing milestones.
8. Before executing milestones, verify the Linear ticket context and status match the current plan state.
9. When taking a ticket off `Backlog`, immediately set it to `In Progress` before starting implementation work.
10. Synchronize ticket status throughout execution (for example: backlog, in progress, blocked, done) whenever plan progress changes.
11. At completion, ensure the Linear ticket status, plan `Progress`, and `Outcomes & Retrospective` are all aligned.
11. Link each new ticket to the canonical repo tracking project listed in `Linear Project Reference`.

### Linear Project Reference

- Project name: `Project X Agent Execution`
- Project URL: `https://linear.app/anmho/team/ANM/projects/all`
- Rule: include this project link in ticket context and attach the ticket to this project when creating/updating work items.

### Linear CLI Fallback (Ralph Loop Friendly)

When Linear MCP is unavailable, create tickets via the CLI script. **Set `LINEAR_API_KEY` in the agent environment** (e.g. Cursor Cloud secrets, `.env` in project root, or shell export) so the agent can run:

```bash
Use Linear MCP to create issues from <payload>
```

Payloads: `docs/backlog/devex-completed-tickets.json`, `docs/backlog/ci-console-ux-tickets.json`, `docs/backlog/folder-by-feature-tickets.json`, etc. After creating tickets, run `./scripts/linear/link-pr-to-linear.sh ANM-XXX` to add the backlink to the PR. Without `LINEAR_API_KEY`, the agent cannot create tickets; document the manual command for the user.

## Required Workflow For Implementation Work

1. Choose the active ExecPlan in `plans/`.
2. Execute milestones in order; do not skip validation.
3. Record design changes and rationale directly in the active ExecPlan.
4. Keep plan and code in sync in the same change set when possible.

## Dirty Worktree Handling (Required)

1. Do not halt execution solely because the git worktree already contains unrelated modified or untracked files.
2. Continue implementation by scoping edits to requested files and leave unrelated changes untouched.
3. Stop and ask for guidance only when pre-existing changes directly conflict with files required for the current task or make safe progress impossible.

## Bug/Unfinished Audit And Ticket Reconciliation (Required)

1. Before closing implementation, run a focused audit for bugs and unfinished modules (for example: TODO/FIXME markers, incomplete plan milestones, missing modules referenced by plans/docs).
2. For each audit finding, check existing `Anmho` Linear tickets first (Backlog/Todo/In Progress/In Review) to avoid duplicates.
3. Only create a new ticket when no existing issue already covers the finding; keep one ticket per distinct item.
4. Include category (`major`/`medium`/`minor`) and execution context in each new ticket: observed issue, affected files/areas, expected outcome, validation notes, and dependencies.
5. Record reconciliation results in the active ExecPlan (existing ticket links and newly created ticket IDs) before task completion.

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
4. Keep `apps/cloud-console` theme behavior consistent across all routes unless a plan explicitly documents divergence.
5. Quick settings UX should expose `Light`, `Dark`, and `System` in the avatar/account panel (Vercel-style segmented control).
6. When editing theme behavior, validate with frontend builds and visually verify both system light and dark modes.

## Suggested Component Libraries

For frontend component work, use this recommendation order unless a plan documents a different choice:

1. Use `Tailwind CSS` as the default styling foundation and prefer semantic/token-backed utilities.
2. Prefer `shadcn/ui` primitives for common app UI (dialogs, forms, menus, popovers, selects) when they fit the interaction.
3. Use `Headless UI` when you need fully custom visual treatment with accessible behavior and `shadcn/ui` does not fit.
4. Avoid mixing multiple component systems in one surface unless required; document exceptions in the active ExecPlan.
5. Keep component-library choices consistent across `apps/cloud-console` unless divergence is explicitly planned.

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

## Subagent Usage (Required When Applicable)

Use Claude Code subagents to parallelize independent work and protect the main context window from large result sets. Subagents are especially valuable for audit, search, and research tasks that would otherwise flood context.

### When to spawn a subagent

1. **Codebase exploration / audit**: any open-ended scan across multiple directories or services (use `subagent_type: Explore`).
2. **Multi-file parallel searches**: when more than ~3 independent Glob/Grep queries are needed before you can proceed.
3. **Independent implementation slices**: when two or more milestones have no shared state and can safely run concurrently (use `subagent_type: general-purpose` with `run_in_background: true`).
4. **Linear ticket batching**: when creating or syncing more than ~5 tickets, delegate bulk API calls to a background agent rather than blocking the main thread.
5. **Heavy research**: doc fetches, web searches, or reading large dependency trees (use `subagent_type: Explore` or `general-purpose`).

### When NOT to spawn a subagent

- Single targeted file reads or Grep calls — use tools directly.
- Sequential tasks where step N depends on step N-1 output.
- Tasks that require interactive user decisions mid-execution.

### Subagent guidelines

- **Launch subagents when necessary.** For multi-ticket execution, ExecPlan milestones, or parallelizable work, spawn subagents rather than executing sequentially. Do not block on single-threaded execution when the task can be parallelized.
- Prefer `run_in_background: true` for work that does not block the next step.
- Pass complete, self-contained prompts; subagents do not share main-thread context.
- Resume a subagent (via `resume` parameter) for follow-up queries instead of spawning a new one.
- After all background agents complete, consolidate their outputs before updating Linear or plans.
- **Subagents must commit their changes.** Each subagent must create atomic commits before returning; do not leave uncommitted changes in the working tree.

## Commit and PR Discipline (Required)

### Atomic commits

1. Each commit must be atomic: one logical unit of change, independently reviewable and testable.
2. Use detailed commit messages: Conventional Commits format with a descriptive body explaining the problem, approach, and validation (for example: `fix(api): remove hardcoded auth fallback — ACCESS_API_ADMIN_KEY now required; fail fast when unset`).
3. Never combine unrelated subsystems in the same commit.

### PR inspection and triage

1. When evaluating open PRs, statuses, checks, or CI failures, prefer GitHub MCP tools first and `gh` CLI second instead of inferring state from local branches alone.
2. Use GitHub surfaces that expose live PR/check state when possible (for example `gh pr list`, `gh pr checks`, `gh pr view`, and `gh run view --log-failed`).
3. In status updates and final summaries, distinguish GitHub-observed failures from local reproduction attempts.

### PR creation and description

1. When creating or managing PRs, use the **gh-pr-mcp** skill (`.cursor/skills/gh-pr-mcp/SKILL.md`): prefer GitHub MCP tools when available, fall back to `gh` CLI.
2. Use `gh pr create` (or `gh pr create --fill`) when GitHub MCP is unavailable.
3. PR description must include:
   - **Linear backlink:** `Linear: [ANM-XXX](https://linear.app/anmho/issue/ANM-XXX)` at the top.
   - **Summary:** What changed and why.
   - **Context:** Affected areas, validation steps, risks.
   - **Screenshots:** When UI or visual behavior changes, attach screenshots (before/after if relevant).
4. Ensure the PR is linked to the Linear issue (via description or GitHub–Linear integration). Runbook: [docs/runbooks/linear-pr-linkage.md](docs/runbooks/linear-pr-linkage.md). After creating tickets: `./scripts/linear/link-pr-to-linear.sh ANM-XXX`.
5. Every PR must reference the owning Linear ticket. The PR body backlink is mandatory even if the branch name or title already includes the ticket ID; once the PR exists, add its link back into the Linear ticket so each issue records its review URL.
6. Default merge strategy is **Squash and Merge** into `main`. Do not use merge commits or rebase merges unless the user explicitly asks for an exception.
7. Treat `main` as the only default integration branch unless a plan explicitly documents a different long-lived branch.

## Ralph Loop Friendly (Required)

This repo is designed for autonomous agent loops (e.g. Ralph). Agents must:

1. **Commit and push after each logical change.** Do not leave work uncommitted; push to the branch so the loop can resume or another agent can continue.
2. **Create Linear tickets when instructed.** Use the Linear MCP tool (create_issue or equivalent). Do not defer ticket creation when Linear MCP is available.
3. **Link PRs to Linear.** After creating a parent ticket, run `./scripts/linear/link-pr-to-linear.sh ANM-XXX` to update the PR description.
4. **Complete the full loop.** Ticket creation → implementation → commit → push → PR update. Do not stop mid-loop with "run this command yourself" when the required env vars are available.

## Cursor Cloud Specific Instructions

When running as a Cursor Cloud Agent:

1. **Environment**: `.cursor/environment.json` defines `install: npm install`. The snapshot should already include Go, Rust, and Python from agent-driven setup at [cursor.com/onboard](https://cursor.com/onboard).
2. **Validation**: Run `npm run verify all` for full checks, or `npm run verify platform` / `npm run verify apps` for specific checks.
3. **Secrets and MCPs**: For parity with local Cursor, add secrets and MCP servers via the dashboard. See [docs/runbooks/cursor-cloud-parity.md](docs/runbooks/cursor-cloud-parity.md) for the full checklist (GitHub MCP, Linear MCP, Greptile MCP, `GITHUB_PERSONAL_ACCESS_TOKEN`, `GREPTILE_API_KEY`, `LINEAR_API_KEY`, etc.). **Ralph loop:** Add `LINEAR_API_KEY` so the agent can create tickets via CLI when Linear MCP is unavailable. Greptile setup: [docs/runbooks/greptile-mcp-setup.md](docs/runbooks/greptile-mcp-setup.md).
