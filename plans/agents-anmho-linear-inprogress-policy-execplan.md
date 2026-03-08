# Align AGENTS.md Linear policy to Anmho and enforce backlog-to-in-progress transitions

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

`PLANS.md` is checked into this repository, and this document follows its required structure.

Linear ticket linkage for this plan: `ANM-32` ([issue link](https://linear.app/anmho/issue/ANM-32/docspolicy-update-agentsmd-linear-team-guidance-and-in-progress)).

## Purpose / Big Picture

After this change, repository policy points agents to the `Anmho` Linear team for new tickets and explicitly requires moving tickets from `Backlog` to `In Progress` when work starts. This is observable by reading `AGENTS.md` and confirming updated rules in the `Linear MCP Ticketing (Required)` section.

## Progress

- [x] (2026-03-08 22:35Z) Confirmed active Linear teams in MCP and identified `Anmho` as available primary team for this workspace.
- [x] (2026-03-08 22:36Z) Created Linear ticket `ANM-32` with required request/scope/validation/risk context and set status to `In Progress`.
- [x] (2026-03-08 22:38Z) Updated `AGENTS.md` Linear policy rules to use `Anmho` and added explicit backlog-to-in-progress transition requirement.
- [x] (2026-03-08 22:38Z) Created this ExecPlan and linked it to `ANM-32` before execution closeout.
- [x] (2026-03-08 22:38Z) Validated final policy text with targeted grep checks and confirmed plan file path discovery.
- [x] (2026-03-08 22:38Z) Performed post-change rescan of policy scope, identified no additional follow-up items requiring new tickets, and moved `ANM-32` to `Done`.
- [x] (2026-03-08 22:37Z) Logged an execution mistake immediately in `docs/agent-mistakes.md` per repository policy.

## Surprises & Discoveries

- Observation: The previously referenced `XPLAT` team key was not available in the connected Linear workspace.
  Evidence: Linear MCP team listing returned only `Anmho`.

- Observation: Shell command substitution errors still occur if `rg` patterns containing backticks are double-quoted.
  Evidence: zsh attempted to execute `Anmho`, `Backlog`, and `In` during initial validation commands.

## Decision Log

- Decision: Treat `Anmho` as the primary Linear team for this repository policy.
  Rationale: The user explicitly approved `Anmho` as the working team and MCP confirms availability.
  Date/Author: 2026-03-08 / Codex

- Decision: Add a dedicated numbered rule that requires immediate `Backlog` to `In Progress` transition when work begins.
  Rationale: The prior synchronization rule was broad; explicit transition language removes ambiguity at execution start.
  Date/Author: 2026-03-08 / Codex

## Outcomes & Retrospective

Completed outcomes:

- `AGENTS.md` now names `Anmho` as the primary Linear team for repository work.
- `AGENTS.md` now explicitly requires moving a ticket from `Backlog` to `In Progress` when implementation starts.
- Ticket lifecycle stayed synchronized (`Backlog` -> `In Progress` -> `Done`) for `ANM-32`.
- Required mistake logging was completed in `docs/agent-mistakes.md` immediately after detection.

Remaining gaps: none for this scope.

## Context And Orientation

This change is policy-only and affects:

- `AGENTS.md`: operating rules for coding agents in this repository.
- `plans/agents-anmho-linear-inprogress-policy-execplan.md`: living execution record for this specific policy change.
- `docs/agent-mistakes.md`: mandatory mistake log updated for an execution-time shell quoting error.

No runtime services, packages, or application behavior are modified.

## Plan Of Work

Apply a focused edit to the existing `Linear MCP Ticketing (Required)` rules in `AGENTS.md` to replace team guidance and add explicit start-of-work status behavior. Keep this plan updated with ticket linkage and progress timestamps, then validate the resulting text using targeted grep checks.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Edit `AGENTS.md` `Linear MCP Ticketing (Required)` rule list:
   - Replace `XPLAT` guidance with `Anmho` and `ANM-<number>`.
   - Add explicit `Backlog` to `In Progress` transition rule when implementation starts.
2. Create/update this ExecPlan with progress, decisions, and outcomes linked to `ANM-32`.
3. Validate with:
   - `rg -n "Use the \`Anmho\` Linear team/key as the primary team" AGENTS.md`
   - `rg -n "When taking a ticket off \`Backlog\`, immediately set it to \`In Progress\`" AGENTS.md`
   - `rg --files plans | rg "agents-anmho-linear-inprogress-policy-execplan.md"`

Expected result: each command returns one match.

## Validation And Acceptance

Acceptance criteria:

- `AGENTS.md` references `Anmho` as primary team in Linear ticketing policy.
- `AGENTS.md` includes explicit rule for moving a ticket to `In Progress` when taking it off `Backlog`.
- This plan file exists under `plans/` and links to the associated Linear issue.
- Linear issue state reflects execution lifecycle (`In Progress` during work, `Done` at completion).

## Idempotence And Recovery

Reapplying these text edits is safe. If duplicate wording appears, remove duplicate lines and rerun the grep validations. If this policy must be rolled back, revert only the `AGENTS.md` and plan-file changes together so policy and execution record stay aligned.

## Artifacts And Notes

Primary artifacts:

- `AGENTS.md` policy update.
- `plans/agents-anmho-linear-inprogress-policy-execplan.md` living execution record.
- `docs/agent-mistakes.md` appended mistake entry for this task.
- Linear issue: `ANM-32`.

## Interfaces And Dependencies

No code interfaces changed. Dependencies are operational process dependencies only:

- Linear MCP availability for team `Anmho`.
- Repository policy compliance via `AGENTS.md` and `PLANS.md`.

Revision note (2026-03-08): Initial plan created for user-requested policy change switching Linear primary team to `Anmho` and adding explicit backlog-to-in-progress transition guidance.
Revision note (2026-03-08): Updated completion status, added validation evidence, recorded post-change rescan result, and documented mistake-log compliance.
