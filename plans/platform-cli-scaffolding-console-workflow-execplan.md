# Refocus platform-cli on scaffolding and console workflows

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-126` ([issue link](https://linear.app/anmho/issue/ANM-126/medium-deprecate-platform-local-runtime-workflow-and-refocus-platform)).

## Purpose / Big Picture

Repository guidance currently pushes `./platform` as the default local runtime stack control path. The requested direction is different: platform-cli should be documented primarily as a scaffolding and console-workflow access tool, not as a local runtime manager.

## Progress

- [x] (2026-03-10 09:22Z) Created `ANM-126` with scope, risks, and validation context; moved issue to `In Progress`.
- [x] (2026-03-10 09:22Z) Audited existing `./platform`-first guidance across policy and runbook docs.
- [x] (2026-03-10 09:25Z) Updated policy/runbook docs to remove `./platform`-first local runtime guidance and reframe platform-cli focus.
- [x] (2026-03-10 09:25Z) Ran post-change rescan for stale local-runtime guidance; fixed remaining stale language in `docs/agent-mistakes.md`; no additional follow-up tickets required.
- [x] (2026-03-10 09:25Z) Moved `ANM-126` to `Done` after docs updates and rescan reconciliation.

## Surprises & Discoveries

- Observation: The `platform` command surface currently still exposes local lifecycle operations (`start|status|stop|logs`) alongside scaffolding and operational commands.
  Evidence: `./platform --help`.

## Decision Log

- Decision: Keep command-surface references factual while removing policy language that treats `./platform` local runtime control as required/default.
  Rationale: User request is about workflow focus and guidance, not immediate removal of command implementations.
  Date/Author: 2026-03-10 / Codex

## Outcomes & Retrospective

Completed outcomes:

- `AGENTS.md` now positions platform-cli around scaffolding and console workflow access instead of local runtime-first validation.
- `docs/runbooks/platform-cli-workflow.md` now documents scaffolding/workflow commands and validation guidance without local stack lifecycle defaults.
- `scripts/README.md`, `README.md`, and `docs/top-three-priorities.md` now align with the same platform-cli focus.
- `docs/agent-mistakes.md` stale runtime-first preventive note was updated to task-scoped validation guidance.

Remaining gaps:

- Historical plan files still contain `./platform` runtime references as point-in-time records; these were not rewritten in this scoped docs policy update.

## Context And Orientation

Relevant files:

- `AGENTS.md`
- `docs/runbooks/platform-cli-workflow.md`
- `scripts/README.md`
- `README.md`
- `docs/top-three-priorities.md`

## Plan Of Work

1. Rewrite policy language to emphasize scaffolding and console workflow operations.
2. Remove `./platform` local stack control as default guidance.
3. Validate consistency by rescanning docs for stale `./platform start/status/stop/logs` policy language.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Edit policy/runbook docs to remove local runtime-first framing.
2. Ensure replacement guidance calls out scaffolding and console workflow use cases.
3. Run:
   - `rg -n "\\./platform start|\\./platform status|\\./platform stop|\\./platform logs|default runtime|runtime validation" AGENTS.md README.md docs scripts -S`
4. Reconcile findings with existing Linear tickets and create follow-up only when needed.

## Validation And Acceptance

Acceptance criteria:

- `AGENTS.md` no longer mandates `./platform` local runtime validation workflow.
- Platform CLI guidance is centered on scaffolding and console workflows.

Validation executed:

- `./platform --help`
- `rg -n "\\./platform|platform-cli|dev-stack|stack:start|stack:status|stack:stop|platform status|platform start|platform stop|platform logs" AGENTS.md README.md docs plans package.json scripts -S`
- `rg -n "\\./platform start|\\./platform status|\\./platform stop|\\./platform logs|default runtime|runtime validation|Platform CLI First Workflow|local stack operations" AGENTS.md README.md docs/runbooks/platform-cli-workflow.md scripts/README.md docs/top-three-priorities.md docs/agent-mistakes.md -S`

## Idempotence And Recovery

If guidance regresses, restore these docs to remove local-runtime-first policy language and re-establish platform-cli positioning around scaffolding and console workflows.

## Artifacts And Notes

- Linear issue: `ANM-126`
- Plan file: `plans/platform-cli-scaffolding-console-workflow-execplan.md`

## Interfaces And Dependencies

- Platform CLI documentation and runbook language (`AGENTS.md`, runbooks, script docs, README status sections).
