# Require Codex session UUID traceability for active Linear tickets

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

`PLANS.md` is checked into this repository, and this document follows the section structure used by current ExecPlans in `plans/`.

Linear ticket linkage for this plan: `ANM-136` ([issue link](https://linear.app/anmho/issue/ANM-136/minor-require-session-assignee-ownership-on-active-linear-tickets)).

## Purpose / Big Picture

After this change, repository policy explicitly requires every active Linear work item to record the active Codex session UUID before implementation begins, while still keeping a real Linear user as assignee. This makes execution traceability concrete instead of relying only on project, team, and status fields.

## Progress

- [x] (2026-03-10 10:53Z) Created `ANM-136` in Linear, assigned it to `me`, and moved it to `In Progress` before editing policy.
- [x] (2026-03-10 10:55Z) Recorded the active Codex session UUID (`CODEX_THREAD_ID`) on `ANM-136` in both ticket context and a comment.
- [x] (2026-03-10 11:03Z) Updated `AGENTS.md` to require session UUID traceability on active tickets while retaining real-user assignee ownership.
- [x] (2026-03-10 11:03Z) Created this ExecPlan and linked it to `ANM-136`.
- [x] (2026-03-10 11:04Z) Validated the new rule text with targeted searches and found no additional follow-up items for this policy scope.

## Surprises & Discoveries

- Observation: The current environment exposes a Codex session UUID via `CODEX_THREAD_ID`, which can be recorded for traceability.
  Evidence: `env` returned `CODEX_THREAD_ID=019cd710-2952-7f62-b437-2943ea7d9693`.

- Observation: Linear assignee fields accept users, not arbitrary session UUIDs.
  Evidence: Linear MCP issue operations expose `assignee` as a user field and the current issue remains assigned to `Andrew Ho` while the session UUID is recorded separately in ticket context/comment.

## Decision Log

- Decision: Keep the real issue assignee as the current session owner (`me` in Linear MCP), but enforce Codex session UUID recording as the session identity requirement.
  Rationale: Linear cannot assign an issue to a UUID, but the current environment exposes `CODEX_THREAD_ID`, which can be recorded reliably on each active ticket.
  Date/Author: 2026-03-10 / Codex

- Decision: Place the session UUID rule immediately after the assignee rule and before the existing `Backlog` to `In Progress` transition rule.
  Rationale: The ticket should have both real-user ownership and session traceability before status changes or code edits begin.
  Date/Author: 2026-03-10 / Codex

## Outcomes & Retrospective

Completed outcomes:

- `ANM-136` is assigned to the current Linear user and records the active Codex session UUID in both its description and a comment.
- `AGENTS.md` now requires recording `CODEX_THREAD_ID` on active tickets before implementation begins.
- Targeted `rg` validation confirms the new assignee/session-UUID rules are present and the ExecPlan file is discoverable.

Remaining gaps:

- None for this scope.

## Context And Orientation

This is a policy/docs-only change affecting:

- `AGENTS.md`: repository operating rules for Linear ticket workflow.
- `plans/agents-linear-session-assignee-policy-execplan.md`: living record for this policy addition.

No runtime code, product surfaces, or local tooling behavior are modified.

## Plan Of Work

Add one explicit session-UUID traceability rule to `AGENTS.md` under `Linear MCP Ticketing (Required)`, keep this ExecPlan synchronized with the task state, and validate the resulting policy text with targeted repo searches.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Edit `AGENTS.md` to require recording the active Codex session UUID (`CODEX_THREAD_ID` when available) on every active work item before implementation begins.
2. Create/update this ExecPlan with ticket linkage, decisions, and progress timestamps.
3. Validate:
   - `rg -n 'Record the active Codex session UUID' AGENTS.md`
   - `rg --files plans | rg 'agents-linear-session-assignee-policy-execplan.md'`
4. Run a focused post-change scan for adjacent ticket-workflow policy gaps and reconcile follow-up ticketing if needed.

Expected result: the policy clearly requires recording `CODEX_THREAD_ID` on active tickets and this plan records the linked ticket lifecycle.

## Validation And Acceptance

Acceptance criteria:

- `AGENTS.md` contains an explicit session UUID traceability rule for active tickets.
- This plan exists under `plans/` and links to `ANM-136`.
- `ANM-136` remains assigned to the current session owner, records the active Codex session UUID, and moves from `In Progress` to `Done` with the plan.

## Idempotence And Recovery

Reapplying the same text change is safe if duplicate wording is avoided. If rollback is required, revert `AGENTS.md` and this plan together so the policy and execution record remain aligned.

## Artifacts And Notes

Primary artifacts:

- `AGENTS.md`
- `plans/agents-linear-session-assignee-policy-execplan.md`
- Linear issue `ANM-136`

## Interfaces And Dependencies

No runtime interfaces change. Process dependencies are:

- Linear MCP support for `assignee: me`
- Codex exposing `CODEX_THREAD_ID` in the session environment
- Existing `Anmho` team and `X Platform` project policy already in place

Revision note (2026-03-10): Initial plan created to enforce Codex session UUID traceability for active Linear tickets.
Revision note (2026-03-10): Updated to final validated state with session UUID recorded on `ANM-136` and no remaining follow-up items.
