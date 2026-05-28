# Restore Mintlify local docs startup from repo root

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-98` ([issue link](https://linear.app/anmho/issue/ANM-98/medium-fix-mintlify-docsdev-root-execution-missing-docsjson)).

## Purpose / Big Picture

Running `npm run docs:dev` from the repository root should start Mintlify locally without configuration errors. This plan restores the missing Mintlify config and aligns verification scripts with current CLI expectations.

## Progress

- [x] (2026-03-10 08:38Z) Created `ANM-98` in Linear with scope/validation/risk context and moved it to `In Progress`.
- [x] (2026-03-10 08:40Z) Audited root scripts and `docs/mintlify`; confirmed Mintlify config file is missing.
- [x] (2026-03-10 08:42Z) Added minimal Mintlify docs config and starter page (`docs/mintlify/docs.json`, `docs/mintlify/index.mdx`).
- [x] (2026-03-10 08:43Z) Updated verification script logic to accept `docs.json` and `mint.json`.
- [x] (2026-03-10 08:43Z) Ran `npm run docs:dev` startup smoke check; original `docs.json`-missing error no longer appears.
- [x] (2026-03-10 08:44Z) Completed post-change bug/unfinished scan and created follow-up ticket `ANM-104` for Mintlify version pinning.
- [x] (2026-03-10 08:44Z) Moved `ANM-98` to `Done` after implementation + validation handoff.

## Surprises & Discoveries

- Observation: `docs/mintlify` currently contains only orchestration metadata (`project.json`, `stack.json`) and no Mintlify config (`docs.json`/`mint.json`).
  Evidence: `ls -la docs/mintlify`.

- Observation: Full docs verification remains blocked in this shell because repository checks require Node `>=24.14.0`, but local runtime is `v20.12.2`.
  Evidence: docs verification command output at the time.

## Decision Log

- Decision: Use `docs.json` as the primary config and keep tooling compatible with either `docs.json` or `mint.json`.
  Rationale: Current Mintlify CLI error explicitly requires `docs.json`, while existing verification code still references `mint.json`.
  Date/Author: 2026-03-10 / Codex

## Outcomes & Retrospective

Completed outcomes:

- Added a working baseline Mintlify project config at `docs/mintlify/docs.json`.
- Added a navigable starter page at `docs/mintlify/index.mdx`.
- Updated docs verification logic in `scripts/verify` to support both `docs.json` and `mint.json`.
- `npm run docs:dev` no longer immediately fails with "must be run in a directory where a docs.json file exists."
- Post-change reconciliation created `ANM-104` for deterministic/pinned Mintlify CLI versioning.

Remaining gaps:

- Could not run full docs verification command in use at the time due local Node runtime mismatch (`v20.12.2` vs required `>=24.14.0`).

## Context And Orientation

Relevant files:

- `package.json`
- `scripts/verify`
- `docs/mintlify/project.json`
- `docs/mintlify/stack.json`
- `docs/mintlify/docs.json` (to add)
- `docs/mintlify/index.mdx` (to add)

## Plan Of Work

1. Add required Mintlify config and a minimal landing page.
2. Make docs verification tolerant of either Mintlify config filename.
3. Run startup validation from repo root and record outcomes.
4. Scan for adjacent unfinished docs tooling gaps and create/reconcile tickets.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Create `docs/mintlify/docs.json` with valid minimal navigation.
2. Create `docs/mintlify/index.mdx` referenced by navigation.
3. Update `scripts/verify` node check to load `docs.json` or `mint.json`.
4. Run:
   - `npm run docs:dev` (startup smoke check)
   - docs verification command active at execution time (if environment requirements permit)
5. Re-scan with:
   - `rg -n "mint\\.json|docs\\.json|docs:dev|npx mint" package.json scripts/verify docs/mintlify`

## Validation And Acceptance

Acceptance criteria:

- `npm run docs:dev` from repo root no longer errors about missing `docs.json`.
- Docs verify logic does not hard-fail when only `docs.json` exists.

Validation executed:

- `npm run docs:dev` (startup smoke check; command no longer emits missing `docs.json` error)
- docs verification command active at execution time (failed early due Node version guard)
- `rg -n "TODO|FIXME|TBD|XXX" scripts/verify docs/mintlify plans/mintlify-docs-dev-setup-execplan.md -S`

## Idempotence And Recovery

If setup drifts again, rerun this plan: restore a valid `docs/mintlify/docs.json`, ensure referenced pages exist, and keep verify scripts compatible with active Mintlify config filename.

## Artifacts And Notes

- Linear issue: `ANM-98`
- Plan file: `plans/mintlify-docs-dev-setup-execplan.md`

## Interfaces And Dependencies

- Mintlify CLI invoked through `npx mint`.
- Node runtime requirements enforced by `scripts/verify`.
