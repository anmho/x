# Normalize repo command surfaces around Nx build/lint/test

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan:

- `ANM-232` ([issue link](https://linear.app/anmho/issue/ANM-232/major-make-app-test-targets-consistent-with-lint-build-and-unit-checks))
- `ANM-234` ([issue link](https://linear.app/anmho/issue/ANM-234/major-narrow-anm-232-into-main-safe-wrapper-and-docs-command-cleanup))
- `ANM-235` ([issue link](https://linear.app/anmho/issue/ANM-235/major-stack-app-target-normalization-on-branch-that-introduces-missing))

## Purpose / Big Picture

Reduce repo command surface bloat by making Nx `build`, `lint`, and `test` the canonical interfaces for local validation and CI. Remove `verify`/`preflight` duplication where it is only naming drift, and keep deploy-preflight as a thin deploy-oriented entrypoint instead of a parallel validation system.

## Progress

- [x] (2026-03-26 08:16Z) Confirmed root drift: `make test` delegates to `./scripts/verify all`, and CI still resolves/runs Nx `verify` targets directly.
- [x] (2026-03-26 08:27Z) Audited current app/service target coverage across `cloud-console`, `linear-ticket-sidepanel`, `agent-runner`, `omnichannel-frontend`, `docs`, `platform`, `mcp`, and omnichannel services.
- [x] (2026-03-26 08:56Z) Patched root scripts, Makefile, CI, and active project targets to prefer Nx `build`/`lint`/`test` over duplicated `verify`/`preflight` surfaces.
- [x] (2026-03-26 09:00Z) Logged two implementation misses immediately in `docs/agent-mistakes.md` and corrected them in the same slice (`sdk:verify` stale dependency and missing sidepanel path in CI change detection).
- [x] (2026-03-26 09:10Z) Restored a valid root `package-lock.json` by regenerating it from the current workspace manifest set; Nx project-graph commands now run again.
- [x] (2026-03-26 09:12Z) Validated passing targets under the new contract: `linear-ticket-sidepanel:lint`, `linear-ticket-sidepanel:test`, `docs:test`, `agent-runner:test`.
- [x] (2026-03-26 09:13Z) Captured remaining real app failures under `ANM-236` after `cloud-console:test` and `omnichannel-frontend:test` surfaced frontend lint debt.

## Surprises & Discoveries

- Observation: the main excess is not only shell wrappers; GitHub Actions also hard-codes `verify` as a first-class target for both platform and apps.
  Evidence: `.github/workflows/ci.yml` resolves and runs `withTarget=verify` for `platform`, `sdk`, `mcp`, `omnichannel-api`, `omnichannel-worker`, and `cloud-console`.

- Observation: some projects already have the right primitives and some only have a split legacy contract.
  Evidence:
  - `apps/cloud-console/project.json`: `test` = lint, `verify` = build
  - `apps/linear-ticket-sidepanel/project.json`: `build` + `verify` + `preflight`, no `lint` or `test`
  - `apps/omnichannel/frontend/project.json`: `build` only
  - `services/agent-control-api/project.json`: already has clean `build` + `test`

- Observation: deploy-preflight still matters as a deploy-oriented entrypoint, but its implementation can delegate to Nx instead of duplicating repo validation logic.
  Evidence: `scripts/deploy-preflight` shells into `scripts/verify` and custom per-target commands today.

- Observation: local Nx validation is blocked by a pre-existing unresolved `package-lock.json` merge conflict in the active branch.
  Evidence: `package-lock.json` contains `<<<<<<< Updated upstream` / `>>>>>>> Stashed changes` markers and `npx nx run ...` fails before execution with `Expected double-quoted property name`; existing ticket `ANM-198` already tracks this blocker.

- Observation: once the lockfile conflict was removed, Nx targets behaved as expected and exposed real project-specific issues instead of command-surface problems.
  Evidence:
  - `npx nx run linear-ticket-sidepanel:test --outputStyle=static` succeeded
  - `npm run test:docs` succeeded
  - `npx nx run agent-runner:test --outputStyle=static` succeeded
  - `npx nx run cloud-console:test --outputStyle=static` failed on actual ESLint findings after adding a flat config
  - `npx nx run omnichannel-frontend:test --outputStyle=static` failed on actual ESLint findings

- Observation: a clean worktree from `HEAD` still needed the untracked app/project directories copied in before the full target contract could validate there.
  Evidence: `/tmp/x-nx-command-surface-cleanup` initially could not resolve `apps/linear-ticket-sidepanel/scripts/build.mjs` or the `agent-runner` project until those branch-local directories were transplanted, confirming this work remains stacked on the active branch shape rather than the last committed baseline.

## Decision Log

- Decision: keep `scripts/verify` only as a thin compatibility bridge and move the real contract to Nx targets plus root npm scripts.
  Rationale: the user wants direct Nx semantics, but CI, docs, and existing operator habits still reference the script path. A thin bridge is acceptable; duplicate logic is not.
  Date/Author: 2026-03-26 / Codex

- Decision: remove `verify` and `preflight` from app-level Nx targets when they are synonyms for `build`/`test`.
  Rationale: project targets should advertise one clear contract, not three names for the same operation.
  Date/Author: 2026-03-26 / Codex

- Decision: keep deploy-preflight as a deploy-oriented script, but make it call Nx targets instead of custom shell validation flows.
  Rationale: deploy-specific invocation is still useful; duplicating command logic is not.
  Date/Author: 2026-03-26 / Codex

## Outcomes & Retrospective

Completed outcomes:

- Root cause is isolated: duplicated semantics live in project target names, root wrappers, and CI job definitions.
- The repo now expresses the desired command contract in code: root `lint` / `build` / `test` scripts, CI jobs keyed off `lint` / `build` / `test`, and duplicate app-level `verify` / `preflight` targets removed from the active scope.
- The branch no longer carries an invalid lockfile merge artifact; Nx commands execute again.
- Active user-facing docs no longer tell contributors to run removed `verify` / `preflight` root commands.

Remaining gaps:

- Some projects still lack full `lint` coverage and may expose real code issues once explicit lint targets are added.
- Docs and runbooks still contain historical `verify`/`preflight` references and should be cleaned after the code path is stable.
- `cloud-console:test` and `omnichannel-frontend:test` still fail on real frontend lint debt; tracked in `ANM-236`.

## Context And Orientation

Primary files in scope:

- `package.json`
- `Makefile`
- `scripts/verify`
- `scripts/deploy-preflight`
- `.github/workflows/ci.yml`
- `apps/cloud-console/project.json`
- `apps/cloud-console/package.json`
- `apps/linear-ticket-sidepanel/project.json`
- `apps/linear-ticket-sidepanel/scripts/verify.mjs`
- `apps/omnichannel/frontend/project.json`
- `apps/omnichannel/frontend/package.json`
- `docs/mintlify/project.json`
- `agents/project.json`
- `platform-config/project.json`
- `services/omnichannel/project.json`
- `services/mcp/project.json`
- `services/omnichannel/backend-api/project.json`
- `services/omnichannel/backend-worker/project.json`

## Plan Of Work

1. Give active app/document/agent/platform projects explicit `build`, `lint`, and/or `test` targets where those concepts exist today.
2. Remove project-level `verify`/`preflight` duplication where it is only an alias for `build` or `test`.
3. Move root commands and CI to `build`/`lint`/`test`.
4. Leave deploy-preflight as a thin deploy-oriented runner over Nx targets.

## Validation And Acceptance

Acceptance criteria:

- Root `npm run build`, `npm run lint`, and `npm run test` exist and use Nx directly.
- `make test` no longer shells into custom validation logic.
- CI jobs resolve and run `build`/`lint`/`test`, not `verify`.
- Active app targets no longer use `verify`/`preflight` as parallel quality gates.

Validation commands to run and record:

- `npx nx run cloud-console:test`
- `npx nx run linear-ticket-sidepanel:test`
- `npx nx run omnichannel-frontend:test`
- `npm run test:apps`
- `npm run build:apps`
- `npm run lint:apps`

Observed validation:

- `node apps/linear-ticket-sidepanel/scripts/verify.mjs --source-only`
- `node apps/linear-ticket-sidepanel/scripts/build.mjs && node apps/linear-ticket-sidepanel/scripts/verify.mjs`
- `node scripts/ci/verify_docs_config.mjs`
- `npx nx run linear-ticket-sidepanel:lint --outputStyle=static`
- `npx nx run linear-ticket-sidepanel:test --outputStyle=static`
- `npm run test:docs`
- `npx nx run agent-runner:test --outputStyle=static`
- `npx nx run cloud-console:test --outputStyle=static` (fails on real frontend lint findings)
- `npx nx run omnichannel-frontend:test --outputStyle=static` (fails on real frontend lint findings)

## Idempotence And Recovery

These changes are target-surface cleanup only. If a project exposes real lint/build/test failures after normalization, keep the target contract and log the failure as a blocker instead of reintroducing duplicate `verify`/`preflight` semantics.
