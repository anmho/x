# Reduce local artifact bloat and harden ignore hygiene

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

`PLANS.md` is checked into this repo, and this document follows its required structure and workflow.

## Purpose / Big Picture

After this change, contributors can keep the monorepo clean with one command, avoid accidentally committing compiled binaries, and use a consistent ignore policy across services. Success is observable when: (1) `scripts/clean --dry-run --full` lists all high-volume artifacts, (2) repo-local binaries like `services/omnichannel/backend/api` are ignored, and (3) `services/access-api` has an explicit `.gitignore` for local build outputs.

## Progress

- [x] (2026-03-08 01:21Z) Baseline audit captured rough edges and measured artifact hotspots.
- [x] (2026-03-08 01:22Z) Added targeted ignore rules for generated Go binaries and replaced broad `/bin/` masking with explicit `bin/platform` ignore.
- [x] (2026-03-08 02:03Z) Added `services/access-api/.gitignore` for local env/build/runtime artifacts.
- [ ] Extend cleanup tooling to cover high-volume Node/Next/Go artifacts via an explicit full-clean mode.
- [ ] Add lightweight monorepo package-manager guidance in root config to reduce per-project install drift.
- [ ] Validate behavior and record final outcomes.

## Surprises & Discoveries

- Observation: The largest non-ignored files are compiled Go binaries in service directories.
  Evidence: `services/omnichannel/backend/api` (~33MB), `services/omnichannel/backend/worker` (~33MB), `services/access-api/api` (~14MB).

- Observation: Existing `scripts/clean` omits the largest current artifact classes (`node_modules`, `.next`, and Go service binaries).
  Evidence: repo artifact totals are dominated by `node_modules` + `target` + `.next`.

- Observation: Root `.gitignore` currently ignores entire `/bin/`, which can hide stale binaries broadly.
  Evidence: `bin/platform` exists (~8.2MB) and the directory is fully masked.

## Decision Log

- Decision: Add targeted ignore rules for known generated binaries instead of broad filename patterns like `**/api`.
  Rationale: Avoid accidentally ignoring legitimate source files named `api` in other paths.
  Date/Author: 2026-03-08 / Codex

- Decision: Keep cleanup defaults conservative and introduce an explicit `--full` mode for heavyweight removals.
  Rationale: Preserves current behavior while enabling predictable deep cleanup when requested.
  Date/Author: 2026-03-08 / Codex

## Outcomes & Retrospective

Pending implementation.

## Context And Orientation

Key files for this effort:

- Root ignore policy: `.gitignore`
- Service ignore policy: `services/omnichannel/.gitignore`
- Missing service ignore policy to add: `services/access-api/.gitignore`
- Cleanup entrypoint: `scripts/clean`
- Developer command surfaces: `package.json`, `Makefile`

Generated artifacts here are local development byproducts (for example, `node_modules`, `.next`, Rust `target`, and compiled Go binaries). They should remain local and not pollute commit surfaces.

## Plan Of Work

First, tighten ignore rules at root and service level so generated Go binaries and selected local executables are not accidentally staged. Second, add an explicit ignore file for `services/access-api` to standardize behavior with the omnichannel service. Third, upgrade `scripts/clean` with a full-clean mode that targets measured hotspots (Node modules, Next.js caches, Go binaries) while preserving the existing default cleaning scope. Finally, expose cleanup and workspace guidance via root command surfaces (`package.json`, `Makefile`) and validate with dry-run outputs.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Edit `.gitignore` and `services/omnichannel/.gitignore` to include path-specific generated binaries and refine `/bin/` behavior.
2. Add `services/access-api/.gitignore` with local env/build/runtime patterns.
3. Update `scripts/clean` to support `--full` plus dry-run compatibility.
4. Update `package.json` and `Makefile` to expose cleanup/workspace commands.
5. Run:
   - `./scripts/clean --dry-run`
   - `./scripts/clean --dry-run --full`
   - `git check-ignore services/omnichannel/backend/api services/omnichannel/backend/worker services/access-api/api`

Expected outputs include ignore matches for all listed binaries and full-clean dry-run lines for Node/Next/Go artifact paths.

## Validation And Acceptance

Acceptance criteria:

- `git check-ignore` confirms Go binary outputs are ignored.
- `scripts/clean --dry-run --full` prints all high-volume artifact paths without failing.
- New files and edits are limited to hygiene/config/cleanup scope.

## Idempotence And Recovery

All edits are configuration-level and safe to re-run. `scripts/clean --dry-run --full` is non-destructive and can be run repeatedly. If any cleanup behavior is too aggressive, revert by editing `paths_full` in `scripts/clean` and re-running dry-run validation before actual removal.

## Artifacts And Notes

Artifacts will be appended during implementation:

- Ignore validation snippets
- Cleanup dry-run snippets
- Commit list per fix

## Interfaces And Dependencies

Interfaces touched:

- Bash CLI behavior for `scripts/clean` (`--dry-run`, `--full`).
- Root npm scripts in `package.json`.
- Make targets in `Makefile`.

No external network dependencies are required for this work.

Revision note: Initial plan created to execute repo hygiene fixes identified from artifact and ignore audit.
