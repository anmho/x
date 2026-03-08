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

## Required Workflow For Implementation Work

1. Choose the active ExecPlan in `plans/`.
2. Execute milestones in order; do not skip validation.
3. Record design changes and rationale directly in the active ExecPlan.
4. Keep plan and code in sync in the same change set when possible.

## Definition Of Done For Any ExecPlan

An ExecPlan is complete only when:

- The promised user-visible behavior can be demonstrated end-to-end.
- Validation commands are documented and pass.
- The plan's living sections reflect final reality.
- Remaining gaps are explicitly listed in `Outcomes & Retrospective`.

## Commit Discipline (Required)

All implementation work must be committed in logical slices, not as one monolithic commit.

1. Create one commit per logical unit of change (for example: policy/docs, routing, schema/types, provider adapter, reconciler, API handlers, UI wiring).
2. Keep each commit independently reviewable and testable.
3. Never combine unrelated subsystems in the same commit.
4. Include validation evidence in the active ExecPlan as each slice lands.

## Mistake Logging (Required)

Every mistake discovered during implementation must be recorded immediately in `docs/agent-mistakes.md`.

For each entry include:

1. Timestamp (UTC)
2. What happened
3. Root cause
4. Preventive rule/check added
5. Verification that the prevention is in place

Do not defer logging mistakes until the end of a task.
