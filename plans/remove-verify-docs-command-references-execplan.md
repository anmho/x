# Remove retired docs verify command references

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-128` ([issue link](https://linear.app/anmho/issue/ANM-128/minor-remove-stale-scriptsverify-docs-references-after-docs-verify)).

## Purpose / Big Picture

The repository has retired the docs-mode verify command, but stale references and command surfaces remain. This plan removes the retired command from active script entrypoints and cleans remaining references so contributors stop invoking it.

## Progress

- [x] (2026-03-10 09:27Z) Created `ANM-128` with scope/validation/risk context and moved it to `In Progress`.
- [x] (2026-03-10 09:27Z) Audited current references to the retired docs-mode verify command and npm alias across scripts/docs/plans.
- [x] (2026-03-10 09:29Z) Removed retired docs verify command surface from active scripts.
- [x] (2026-03-10 09:29Z) Removed stale references in docs/plans/backlog text where appropriate.
- [x] (2026-03-10 09:31Z) Ran post-change rescan and reconciled follow-up findings (no new ticket required).
- [x] (2026-03-10 09:32Z) Moved `ANM-128` to `Done`.

## Surprises & Discoveries

- Observation: The retired docs-mode verify command still existed as an executable mode and was included in `scripts/verify all`.
  Evidence: pre-change `scripts/verify` usage and `case` dispatch included `docs`.

## Decision Log

- Decision: Remove the `docs` mode from `scripts/verify` and drop the retired npm docs-verify alias from root scripts.
  Rationale: The command is retired and continuing to expose it encourages accidental usage.
  Date/Author: 2026-03-10 / Codex

## Outcomes & Retrospective

Completed outcomes:

- Removed `docs` mode from `scripts/verify` and from the `all` execution path.
- Removed root npm alias for the retired docs verify mode.
- Updated docs/plans/backlog references to avoid recommending the retired command.

Remaining gaps:

- Historical references to old runtime patterns outside this command cleanup may still exist and are tracked separately.

## Context And Orientation

Relevant files:

- `scripts/verify`
- `package.json`
- `scripts/README.md`
- `docs/repository-architecture-deep-dive.md`
- `plans/*.md` entries that still mention the retired command
- `docs/backlog/*.json` entries that still cite retired command usage

## Plan Of Work

1. Remove retired command surface from scripts.
2. Clean stale references from policy/docs/plans/backlog artifacts.
3. Rescan and confirm no references to the retired docs-mode verify command or alias remain.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Edit `scripts/verify` to remove `docs` mode and associated usage text.
2. Edit `package.json` to remove retired docs-verify npm alias.
3. Update references in docs/plans/backlog text.
4. Run:
   - `rg -n "<retired docs-verify command patterns>" -S`
   - `rg -n "TODO|FIXME|TBD|XXX" scripts/verify scripts/README.md plans/remove-verify-docs-command-references-execplan.md -S`

## Validation And Acceptance

Acceptance criteria:

- `scripts/verify` no longer accepts `docs` mode.
- `package.json` no longer exposes the retired docs-verify npm alias.
- No stale references to the retired docs verify command remain.

Validation executed:

- Initial audit grep for retired docs-verify command/alias references.
- Refined post-change grep for retired docs-verify command/alias references (no matches after cleanup).
- `rg -n "TODO|FIXME|TBD|XXX" scripts/verify scripts/README.md plans/remove-verify-docs-command-references-execplan.md -S` (no unresolved markers in scoped files)

## Idempotence And Recovery

If references regress, rerun the grep patterns in this plan and remove newly introduced mentions and command surfaces.

## Artifacts And Notes

- Linear issue: `ANM-128`
- Plan file: `plans/remove-verify-docs-command-references-execplan.md`

## Interfaces And Dependencies

- Root script command surfaces (`scripts/verify`, `package.json` scripts).
- Contributor docs and historical plan/backlog records.
