# Reduce local artifact bloat and harden ignore hygiene

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

`PLANS.md` is checked into this repo, and this document follows its required structure and workflow.

## Purpose / Big Picture

After this change, contributors can keep the monorepo clean with one command, avoid accidentally committing compiled binaries, and use a consistent ignore policy across services. Success is observable when: (1) `scripts/clean --dry-run --full` lists all high-volume artifacts, (2) repo-local binaries like `services/omnichannel/backend/api` are ignored, and (3) `services/access-api` has an explicit `.gitignore` for local build outputs.

## Progress

- [x] (2026-03-08 01:21Z) Baseline audit captured rough edges and measured artifact hotspots.
- [x] (2026-03-08 01:22Z) Added targeted ignore rules for generated Go binaries and replaced broad `/bin/` masking with explicit `bin/platform` ignore.
- [x] (2026-03-08 02:03Z) Added `services/access-api/.gitignore` for local env/build/runtime artifacts.
- [x] (2026-03-08 22:28Z) Extended `scripts/clean` with `--full` coverage for high-volume Node/Next artifacts and added Go service binaries to default cleanup paths.
- [x] (2026-03-08 22:32Z) Added root `packageManager` pin and explicit npm workspace map for active Node packages.
- [x] (2026-03-08 22:33Z) Ran final ignore/cleanup validation and documented outcomes.
- [x] (2026-03-08 22:38Z) Prepared ANM-team Linear payload for the X Platform project and validated tickets via dry-run.
- [ ] Publish ANM tickets once `LINEAR_API_KEY` is available in the execution environment.

## Ticket Tracker (Project: X Platform / Team Key: ANM)

- ANM-TBD-001: Ignore generated service binaries and narrow root bin ignore scope. Status: Implemented in code (`46d5ea0`), ticket payload prepared, awaiting publish.
- ANM-TBD-002: Add `services/access-api/.gitignore`. Status: Implemented in code (`92c8659`), ticket payload prepared, awaiting publish.
- ANM-TBD-003: Extend cleanup tooling and command surfaces. Status: Implemented in code (`f13478e`), ticket payload prepared, awaiting publish.
- ANM-TBD-004: Add root workspace guidance for Node installs. Status: Implemented in code (`86f2f02`), ticket payload prepared, awaiting publish.
- ANM-TBD-005: Run final validation and close outcomes. Status: Implemented in code (`88fc8cc`), ticket payload prepared, awaiting publish.

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

- Decision: Standardize plan ticket namespace as `XPLAT-*` for this repository effort.
  Rationale: Align with the single project naming convention (`X Platform` / `XPLAT`).
  Date/Author: 2026-03-08 / Codex

- Decision: Use `ANM` as the issue identifier key and keep `X Platform` as the project grouping label.
  Rationale: Linear issue prefixes are team-key-based; the user confirmed the active team key is `ANM`.
  Date/Author: 2026-03-08 / Codex

## Outcomes & Retrospective

Completed outcomes:

- Root and service ignore policies now cover generated Go binaries (`services/access-api/api`, `services/omnichannel/backend/api`, `services/omnichannel/backend/worker`) and explicitly ignore `bin/platform` instead of masking all `/bin`.
- `services/access-api/.gitignore` now exists and codifies local env/build/log artifact handling.
- `scripts/clean` now supports `--full` for high-volume Node/Next cleanup while preserving conservative default behavior.
- Root package metadata now includes a pinned npm package manager version and explicit workspace map for active Node packages.

Remaining gaps:

- No full workspace dependency unification was performed; existing per-project lockfiles remain.
- Linear publish requires `LINEAR_API_KEY` at execution time; ANM ticket payload is prepared.

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

Validation snippets captured:

- `./scripts/clean --dry-run` lists default removals including `services/access-api/api` and omnichannel `api`/`worker` binaries.
- `./scripts/clean --dry-run --full` additionally lists `apps/cloud-console/node_modules`, `apps/cloud-console/.next`, `services/omnichannel/frontend/node_modules`, `services/omnichannel/frontend/.next`, and `mcp/node_modules`.

Commit list per fix:

- `46d5ea0` (`XPLAT-001`): ignore generated service binaries and narrow root bin ignore scope.
- `92c8659` (`XPLAT-002`): add `services/access-api/.gitignore`.
- `f13478e` (`XPLAT-003`): expand cleanup tooling and command entrypoints.
- `86f2f02` (`XPLAT-004`): add root npm workspace guidance.

Linear execution notes:

- Prepared payload: `docs/backlog/x-platform-repo-hygiene-tickets.json` (team key `ANM`, project context `X Platform`).
- Dry-run command: `node scripts/linear/create-issues.mjs --input docs/backlog/x-platform-repo-hygiene-tickets.json --team-key ANM --dry-run`.
- Publish attempt command: `node scripts/linear/create-issues.mjs --input docs/backlog/x-platform-repo-hygiene-tickets.json --team-key ANM`.
  Result: blocked with `LINEAR_API_KEY is required when not using --dry-run`.

## Interfaces And Dependencies

Interfaces touched:

- Bash CLI behavior for `scripts/clean` (`--dry-run`, `--full`).
- Root npm scripts in `package.json`.
- Make targets in `Makefile`.

No external network dependencies are required for this work.

Revision note: Added `XPLAT-*` ticket namespace, recorded validation evidence, and documented final outcomes for the hygiene implementation.
