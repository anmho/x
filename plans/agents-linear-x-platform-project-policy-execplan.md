# Require X Platform as the default Linear project context

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

`PLANS.md` is checked into this repository, and this document follows its required structure.

Linear ticket linkage for this plan: `ANM-33` ([issue link](https://linear.app/anmho/issue/ANM-33/docspolicy-require-linear-project-context-as-x-platform-in-agentsmd)).

## Purpose / Big Picture

After this change, policy states that repository work should use `X Platform` as the default Linear project context, in addition to existing team and status requirements. This is observable in `AGENTS.md` under `Linear MCP Ticketing (Required)`.

## Progress

- [x] (2026-03-08 22:41Z) Created and started Linear ticket `ANM-33` before execution.
- [x] (2026-03-08 22:42Z) Updated `AGENTS.md` to require `X Platform` as the default Linear project context.
- [x] (2026-03-08 22:42Z) Created this ExecPlan and linked it to `ANM-33`.
- [x] (2026-03-08 22:42Z) Validated final policy text and plan file discovery with targeted `rg` checks.
- [x] (2026-03-08 22:43Z) Created Linear project `X Platform` under team `Anmho` and assigned `ANM-33` to it.
- [x] (2026-03-08 22:44Z) Performed post-change rescan for this policy scope; no additional follow-up tickets were needed.
- [x] (2026-03-08 22:44Z) Moved `ANM-33` to `Done` and confirmed project linkage remained `X Platform`.

## Surprises & Discoveries

- Observation: Linear MCP initially returned no current projects in this workspace.
  Evidence: `list_projects` query for `x platform` and default listing returned empty results.

- Observation: Creating `X Platform` project through Linear MCP succeeded and resolved project-context dependency.
  Evidence: `save_project` returned project URL `https://linear.app/anmho/project/x-platform-de7a3e703d58`.

## Decision Log

- Decision: Add a direct rule in `AGENTS.md` requiring `X Platform` as default project context.
  Rationale: The user explicitly set project context to X Platform, so policy must encode it clearly.
  Date/Author: 2026-03-08 / Codex

- Decision: Include "create or restore it before implementation" fallback when project is missing.
  Rationale: Workspace initially had no visible projects, so policy needed explicit recovery behavior.
  Date/Author: 2026-03-08 / Codex

- Decision: Create the missing `X Platform` project immediately in Linear and attach the active ticket.
  Rationale: This aligns execution with the newly documented policy and the user-provided project context.
  Date/Author: 2026-03-08 / Codex

## Outcomes & Retrospective

Completed outcomes:

- `AGENTS.md` now defines `X Platform` as the default Linear project context.
- Linear project `X Platform` now exists in team `Anmho`.
- Active task ticket `ANM-33` is linked to project `X Platform`.

Remaining gap: none for this scope.

## Context And Orientation

This is a documentation/policy-only change affecting:

- `AGENTS.md`: agent operating policy.
- `plans/agents-linear-x-platform-project-policy-execplan.md`: living execution record for this request.

No runtime code, services, or frontend behavior changed.

## Plan Of Work

Add one explicit project-context rule to the existing Linear workflow section in `AGENTS.md`, then validate with targeted searches. Keep this plan synced with the ticket lifecycle and execution evidence.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Edit `AGENTS.md` under `Linear MCP Ticketing (Required)` to add `X Platform` project guidance.
2. Create/update this ExecPlan with status, decisions, and outcomes linked to `ANM-33`.
3. Validate:
   - `rg -n 'Use `X Platform` as the default Linear project' AGENTS.md`
   - `rg --files plans | rg 'agents-linear-x-platform-project-policy-execplan.md'`

Expected result: one match per command.

## Validation And Acceptance

Acceptance criteria:

- `AGENTS.md` includes explicit `X Platform` default project guidance.
- This plan exists under `plans/` and links to `ANM-33`.
- Linear ticket lifecycle is synchronized from `In Progress` to `Done` at completion.

## Idempotence And Recovery

Re-applying this text change is safe. If duplicate lines appear, remove duplicates and rerun validations. If rollback is needed, revert both `AGENTS.md` and this plan in one docs rollback so policy and record stay aligned.

## Artifacts And Notes

Artifacts:

- `AGENTS.md` policy update.
- `plans/agents-linear-x-platform-project-policy-execplan.md` ExecPlan.
- Linear project `X Platform`: `https://linear.app/anmho/project/x-platform-de7a3e703d58`.
- Linear issue `ANM-33`.

## Interfaces And Dependencies

No runtime interfaces changed. Process dependencies:

- Linear team availability for `Anmho`.
- Linear project availability for `X Platform`.

Revision note (2026-03-08): Initial plan created for user request to set `X Platform` as the repository project context in Linear policy.
Revision note (2026-03-08): Updated with validation evidence, project creation artifact, and completion-state details.
