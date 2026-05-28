# Reconcile documentation with current stack/runtime command surfaces

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-97` ([issue link](https://linear.app/anmho/issue/ANM-97/medium-reconcile-stack-control-docs-with-current-repo-tooling)).

## Purpose / Big Picture

After this change, core repo docs consistently describe the currently available stack/runtime commands and remove stale references to removed scripts/modules. Success is observable by verifying command guidance in `AGENTS.md`, runbooks, and script documentation aligns with current code surfaces.

## Progress

- [x] (2026-03-10 08:30Z) Created `ANM-97` in Linear with request/scope/validation/risk context and moved it to `In Progress`.
- [x] (2026-03-10 08:33Z) Audited stack/runtime docs and confirmed drift in `AGENTS.md`, `docs/runbooks/platform-cli-workflow.md`, and `scripts/README.md`.
- [x] (2026-03-10 08:42Z) Updated `AGENTS.md` to remove duplicated/conflicting `scripts/dev-stack` guidance and clarified `./platform stack` vs supervisor-only usage.
- [x] (2026-03-10 08:44Z) Updated `docs/runbooks/platform-cli-workflow.md` to remove non-existent npm stack targets and align service-scoped examples with current commands.
- [x] (2026-03-10 08:47Z) Rewrote `scripts/README.md` to document only scripts currently present in `scripts/`.
- [x] (2026-03-10 08:50Z) Updated `README.md`, `docs/runbooks/build-and-release-gates.md`, `docs/top-three-priorities.md`, and `docs/repository-architecture-deep-dive.md` to remove stale references and align command/module descriptions with the current tree.
- [x] (2026-03-10 08:40Z) Completed post-change bug/follow-up rescan, reconciled existing Linear coverage, and created follow-up tickets `ANM-101`, `ANM-102`, and `ANM-103` for newly identified command-surface breakages.
- [x] (2026-03-10 08:41Z) Logged an execution mistake in `docs/agent-mistakes.md` immediately after detection and corrected the associated invalid path reference.
- [ ] Run the then-active docs verification path once Node runtime is upgraded to satisfy repository requirement (`>=24.14.0`).
- [x] (2026-03-10 08:41Z) Moved `ANM-97` to `Done` after documentation reconciliation and follow-up ticket creation.

## Surprises & Discoveries

- Observation: The then-active docs verification path was blocked in this shell because Node is `v20.12.2` while repo verification required `>=24.14.0`.
  Evidence: command output at the time reported `node version too old`.

- Observation: `scripts/README.md` and several runbooks referenced removed scripts (`scripts/new`, `scripts/doctor`) and removed service docs paths.
  Evidence: targeted grep results across `README.md`, `docs/`, and `scripts/README.md`.

## Decision Log

- Decision: Prioritize command-surface correctness first (`./platform`, `scripts/dev-stack`, npm stack scripts) because this was the user-reported inconsistency.
  Rationale: Runtime command drift causes immediate operational confusion and bad validation reports.
  Date/Author: 2026-03-10 / Codex

- Decision: Keep historical backlog payload files unchanged even when they reference retired modules/paths.
  Rationale: Those files are historical ticket artifacts, not current operational runbooks.
  Date/Author: 2026-03-10 / Codex

- Decision: Create separate follow-up tickets per distinct broken command surface discovered during rescan.
  Rationale: Keeps remediation tasks independently actionable and aligns with repository policy requiring one ticket per distinct item.
  Date/Author: 2026-03-10 / Codex

## Outcomes & Retrospective

Completed outcomes:

- Stack/runtime command guidance is now consistent across:
  - `AGENTS.md`
  - `docs/runbooks/platform-cli-workflow.md`
  - `scripts/README.md`
- Core docs no longer claim removed script surfaces in active guidance (`scripts/new`, `scripts/doctor`).
- README/runbook text now reflects current runtime modules and preflight surfaces.

Remaining gaps:

- The then-active docs verification path could not be completed in the current shell due Node version mismatch.
- Historical backlog JSON payloads intentionally retain past references and were not rewritten.
- Follow-up implementation needed for newly logged issues:
  - `ANM-101` (`scripts/doctor` missing while still referenced)
  - `ANM-102` (stale config script file paths in `package.json`)
  - `ANM-103` (broken `linear:create-issues` npm script)

## Context And Orientation

Primary files updated in this scope:

- `AGENTS.md`
- `README.md`
- `docs/agent-mistakes.md`
- `scripts/README.md`
- `docs/runbooks/platform-cli-workflow.md`
- `docs/runbooks/build-and-release-gates.md`
- `docs/top-three-priorities.md`
- `docs/repository-architecture-deep-dive.md`

## Plan Of Work

1. Validate actual command surfaces from source (`platform`, `platform-cli`, `scripts/dev-stack`, `package.json` scripts).
2. Update policy/runbook/script docs to match those command surfaces exactly.
3. Perform follow-up scan for stale command/module references and patch mismatches in high-traffic docs.
4. Record outcomes, residual validation constraints, and ticket state in this plan.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Confirm command behavior with:
   - `./platform --help`
   - `./platform status`
   - `./platform stack`
   - `sed -n '1,220p' scripts/dev-stack`
   - `cat package.json`
2. Search docs for stale references with `rg` patterns for removed scripts/paths and invalid npm stack commands.
3. Edit mismatched docs and re-scan with `rg`.
4. Run docs verification for the active workflow (blocked pending Node upgrade in current environment).

## Validation And Acceptance

Acceptance criteria:

- No stale active guidance for removed script entrypoints (`scripts/new`, `scripts/doctor`).
- Stack command docs only mention command variants currently available.
- `ANM-97` status tracks plan state (`In Progress` during execution, `Done` at closeout).

Validation executed:

- `rg -n "scripts/new|scripts/doctor|services/access-api|stack:postgres|stack:service|npm run doctor|preflight:access-api|PLANS.md" README.md AGENTS.md docs scripts/README.md`
- `./platform --help`
- `./platform status`
- `./platform stack`
- Docs verification command in use at the time (failed due Node version requirement in current shell)

## Idempotence And Recovery

These are documentation-only changes. Reapplying is safe; if wording diverges later, rerun the same `rg` scan patterns and command checks to realign docs quickly.

## Artifacts And Notes

- Linear issue: `ANM-97`
- Plan file: `plans/docs-stack-command-sync-execplan.md`
- Mistake log update: `docs/agent-mistakes.md`
- Validation artifact: failed docs verify due local Node runtime mismatch (`v20.12.2` vs required `>=24.14.0`)
- Post-change follow-up issues: `ANM-101`, `ANM-102`, `ANM-103`

## Interfaces And Dependencies

- Depends on current command implementations in:
  - `platform`
  - `platform-cli/*`
  - `scripts/dev-stack`
  - root `package.json` scripts

Revision note (2026-03-10): Initial plan created during docs reconciliation execution and linked to `ANM-97`.
