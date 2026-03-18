# Switch main branch required checks to Nx affected CI

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-168` ([issue link](https://linear.app/anmho/issue/ANM-168/medium-switch-main-branch-required-checks-to-nx-affected-ci-in-x)).

## Purpose / Big Picture

Replace coarse GitHub Actions verification buckets on `anmho/x` with stable required checks that always appear on PRs but scope their actual work through `nx affected`. The intent is to keep `main` protected while avoiding monorepo-wide CI for unrelated changes.

## Progress

- [x] (2026-03-18 07:05Z) Created `ANM-168`, assigned it to `me`, moved it to `In Progress`, and recorded the active Codex session UUID.
- [x] (2026-03-18 07:06Z) Audited the current `.github/workflows/ci.yml`, `scripts/verify`, and Nx project graph to confirm existing CI is grouped by coarse shell buckets instead of graph-aware targets.
- [x] (2026-03-18 07:11Z) Patched `.github/workflows/ci.yml` to introduce stable `Affected Platform`, `Affected Apps`, `Affected Docs`, and `Affected Agents` jobs driven by Nx base/head SHAs.
- [x] (2026-03-18 07:12Z) Corrected the initial implementation mistake where `nx affected --projects=...` forwarded extra args into `nx:run-commands` targets; switched the workflow to `nx show projects --affected ...` plus `nx run-many`.
- [x] (2026-03-18 07:12Z) Ran local spot checks proving docs-only changes trigger only docs verification, cloud-console no-ops for docs changes, and `services/mcp` changes scope cleanly to `mcp:verify`.
- [ ] Inspect and record the unrelated local `agent-control-api:test` failure exposed during platform-spot validation so it is not conflated with the CI scoping change.
- [ ] Update GitHub branch protection on `main` to require the new stable check contexts.
- [ ] Push a throwaway validation branch and inspect the resulting GitHub Actions/check status.
- [ ] Reconcile the plan outcomes and `ANM-168` with the final validation evidence.

## Surprises & Discoveries

- Observation: the repository already has Nx projects spanning apps, docs, agents, and multiple services, but the checked-in CI workflow still shells through `./scripts/verify platform|apps|docs`.
  Evidence: `.github/workflows/ci.yml` currently defines `Platform Checks`, `App Checks`, and `Docs Checks` jobs; `scripts/verify` runs grouped shell logic rather than `nx affected`.

- Observation: target support is not uniform across all Nx projects.
  Evidence: projects such as `cloud-console`, `docs`, `mcp`, `omnichannel-api`, and `omnichannel-worker` expose `verify`; `agent-control-api` exposes `build` and `test` but not `verify`; `agents` exposes `verify`; `agent-runner` exposes only `build`.

- Observation: `nx affected --projects=...` is not a safe project filter for this repo’s `nx:run-commands` targets because Nx v22 forwards the flag as an extra task argument.
  Evidence: local validation attempted `npx nx affected -t verify --projects=mcp --files=services/mcp/project.json`, and the underlying command became `go build ./... --projects=mcp`, which failed with `malformed import path "--projects=mcp": leading dash`.

- Observation: `agent-control-api:test` currently fails locally for reasons unrelated to this CI-scoping refactor.
  Evidence: `npx nx run-many -t test --projects=agent-control-api --outputStyle=static` failed with both a network resolution error for `proxy.golang.org` and build-time API mismatches around `runner.NewLocalRunner`.

## Decision Log

- Decision: use a small number of stable GitHub job names and let Nx scope the internal work with `affected`.
  Rationale: GitHub branch protection requires always-present check names, while path-filtered workflows or conditionally absent jobs can deadlock merges.
  Date/Author: 2026-03-18 / Codex

- Decision: group required checks by supported Nx targets rather than forcing one monolithic `affected` invocation across every project.
  Rationale: the current graph does not expose one uniform target surface, so stable groupings are the narrowest safe path without inventing or backfilling unrelated project targets in this task.
  Date/Author: 2026-03-18 / Codex

## Outcomes & Retrospective

Completed outcomes:

- `.github/workflows/ci.yml` now exposes stable GitHub Actions job names intended for branch protection while computing actual work via Nx affected project resolution.
- The workflow no longer relies on coarse `./scripts/verify platform|apps|docs` buckets for required checks.

Remaining gaps:

- Branch-protection updates and live GitHub validation are still in progress.
- `agent-control-api:test` remains broken locally; dummy validation should avoid touching that project so the scoping test isolates the workflow behavior.

## Context And Orientation

Relevant files and systems:

- `.github/workflows/ci.yml`
- `scripts/verify`
- `nx.json`
- `apps/*/project.json`
- `docs/mintlify/project.json`
- `services/*/project.json`
- GitHub branch protection for `anmho/x`

## Plan Of Work

1. Convert CI to stable jobs that always run and compute affected scope from the PR/base diff.
2. Use `nx show projects --affected` plus `nx run-many` groupings that match the current project graph instead of broad shell verify buckets.
3. Update `main` branch protection to require only the stable Nx-aware job contexts.
4. Validate the resulting workflow with a throwaway remote branch and inspect the reported check status names and conclusions.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Inspect Nx projects and targets:
   - `npx nx show projects --json`
   - `rg -n '"verify"|"test"|"build"' **/project.json`
2. Patch `.github/workflows/ci.yml` to add stable affected jobs and shared base/head resolution.
3. Run targeted local validation commands against the modified workflow and representative Nx invocations.
4. Update branch protection on `main` to require the new stable contexts.
5. Push a throwaway validation branch and inspect GitHub Actions runs/checks with `gh`.

## Validation And Acceptance

Acceptance criteria:

- GitHub Actions exposes a stable set of Nx-aware check names on PRs.
- Those jobs always appear, even when no relevant projects are affected.
- The jobs scope internal work through `nx affected` rather than coarse `scripts/verify` buckets.
- `main` branch protection requires the new stable contexts.
- A throwaway remote validation branch demonstrates the expected check names/statuses.

Validation commands:

- `sed -n '1,240p' .github/workflows/ci.yml`
- `npx nx show projects --json`
- `git diff -- .github/workflows/ci.yml plans/nx-affected-required-checks-execplan.md`
- `gh run list --workflow CI --limit 5`
- `gh pr checks <pr-number>` or `gh run view <run-id>`

## Idempotence And Recovery

Re-running this task should converge on the same stable required-check names and branch-protection contexts. If validation shows a target group is too broad or missing a project, adjust the grouping without renaming already-required checks unless protection settings are updated in the same change.

## Artifacts And Notes

- Linear issue: `ANM-168`
- Session UUID: `019cffb5-6166-7b41-9c23-db12149403ea`

## Interfaces And Dependencies

- GitHub Actions workflow/job naming behavior
- GitHub branch protection required status check contexts
- Nx project graph and target availability
- Existing repo verification commands that may remain as fallback implementation details
