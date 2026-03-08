# Require plan file commits for every task

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

`PLANS.md` is checked into this repo, and this document follows its required structure and workflow.

## Purpose / Big Picture

After this change, policy guidance explicitly requires that every task has a tracked plan file in `plans/<plan>.md` and that the plan is committed with related changes. Success is observable by reading `AGENTS.md` and finding the new Commit Discipline rule.

## Progress

- [x] (2026-03-08 22:24Z) Reviewed `PLANS.md` and `AGENTS.md`, confirmed no existing ExecPlan covered this policy tweak.
- [x] (2026-03-08 22:25Z) Added a new Commit Discipline rule in `AGENTS.md` requiring `plans/<plan>.md` to be added and committed for each task.
- [x] (2026-03-08 22:25Z) Created this ExecPlan to document scope, validation, and completion state for the policy change.
- [x] (2026-03-08 22:27Z) Validated the new rule text with targeted grep checks and committed the policy/docs slice.

## Surprises & Discoveries

- Observation: The required Linear team `XPLAT` was unavailable in the connected workspace at execution time.
  Evidence: Linear MCP `get_team("XPLAT")` returned “Entity not found: Team”.

## Decision Log

- Decision: Add the new requirement as a numbered item in `## Commit Discipline (Required)` instead of creating a new top-level section.
  Rationale: The request is explicitly about adding and committing plan files, which is commit behavior and fits the existing section semantics.
  Date/Author: 2026-03-08 / Codex

- Decision: Use `plans/agents-plan-commit-every-time-execplan.md` as the plan path for this change.
  Rationale: Keeps the filename action-oriented and discoverable for future policy audits.
  Date/Author: 2026-03-08 / Codex

## Outcomes & Retrospective

Completed: `AGENTS.md` now includes an explicit rule requiring an ExecPlan at `plans/<plan>.md` to be added and committed for every task.

Remaining gaps: Linear ticket linkage remains blocked by missing `XPLAT` team availability in the current MCP workspace.

## Context And Orientation

This change is limited to repository policy documentation:

- `AGENTS.md` is the operator policy used by coding agents in this repo.
- `plans/*.md` files are task-specific living ExecPlans governed by `PLANS.md`.

No runtime code paths, product behavior, or service interfaces were modified.

## Plan Of Work

Apply a focused docs change in `AGENTS.md` by appending one new rule in the Commit Discipline list. Create a dedicated ExecPlan under `plans/` capturing intent, evidence, and completion details. Validate via direct content checks to confirm the new policy text exists and is scoped correctly.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Edit `AGENTS.md` and add a Commit Discipline item requiring `plans/<plan>.md` to be added and committed every task.
2. Create `plans/agents-plan-commit-every-time-execplan.md` with all required `PLANS.md` sections.
3. Run:
   - `rg -n "For every task, ensure an ExecPlan exists at \`plans/<plan>.md\`" AGENTS.md`
   - `rg --files plans | rg "agents-plan-commit-every-time-execplan.md"`
4. Commit both files in one logical docs/policy commit using a Conventional Commit message.

Expected result: both grep checks return one matching line/path and the commit contains only the policy file plus this plan.

## Validation And Acceptance

Acceptance criteria:

- `AGENTS.md` includes an explicit rule requiring plan-file add/commit behavior for every task.
- This ExecPlan exists in `plans/` and is committed in the same logical change.
- `git show --name-only` for the commit lists only the two expected files.

## Idempotence And Recovery

Re-applying the same text change is safe; duplicate lines can be removed with a single-line revert in `AGENTS.md`. If this policy needs rollback, revert the commit containing both files so the policy and its execution record are removed together.

## Artifacts And Notes

Validation commands executed:

- `rg -n "For every task, ensure an ExecPlan exists at \`plans/<plan>.md\`" AGENTS.md`
- `rg --files plans | rg "agents-plan-commit-every-time-execplan.md"`
- `git show --name-only --oneline --no-patch HEAD`

## Interfaces And Dependencies

No code interfaces or runtime dependencies changed. Affected artifacts are documentation-only:

- `AGENTS.md` policy text
- `plans/agents-plan-commit-every-time-execplan.md` execution record

Revision note: Created to implement and document the requested policy that each task must add and commit a plan file at `plans/<plan>.md`.
