# Fix stale Go bootstrap in CI workflows

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-133` ([issue link](https://linear.app/anmho/issue/ANM-133/major-fix-ci-workflow-stale-go-version-file-path-for-omnichannel)).

## Purpose / Big Picture

Restore GitHub Actions `Platform Checks` and deploy preflight workflow bootstrapping so they no longer fail immediately on a nonexistent `go-version-file`. The goal for this task is to get CI past the dead path and into real validation, while documenting the remaining Go module-layout drift separately if needed.

## Progress

- [x] (2026-03-10 11:28Z) Updated `ANM-133` with the exact affected PR links and moved the ticket to `In Progress`.
- [x] (2026-03-10 11:30Z) Recorded the active Codex session UUID on `ANM-133`.
- [x] (2026-03-10 11:31Z) Inspected `.github/workflows/ci.yml`, `.github/workflows/deploy-preflight.yml`, and `scripts/verify` to confirm the stale bootstrap path.
- [x] (2026-03-10 11:32Z) Confirmed no tracked `go.mod` exists anywhere in the current repo state, so `go-version-file` cannot be repaired by pointing at another checked-in module file.
- [x] (2026-03-10 11:33Z) Patched `.github/workflows/ci.yml` and `.github/workflows/deploy-preflight.yml` to use explicit `go-version: stable`.
- [x] (2026-03-10 11:33Z) Ran targeted validation on the modified workflow files with `rg` and `git diff`.
- [x] (2026-03-10 11:34Z) Re-scanned for follow-up gaps and created `ANM-146` for the missing Go module manifest issue exposed by the bootstrap fix.
- [ ] Align `ANM-133` status, plan outcomes, and validation notes after remote CI rerun or handoff.

## Surprises & Discoveries

- Observation: the stale `go-version-file` is not just wrong for `services/omnichannel/backend`; there is no tracked `go.mod` anywhere in the repository right now.
  Evidence: `rg --files -g 'go.mod'` returned no matches; `git ls-tree -r --name-only HEAD | rg 'go\.mod$|go\.sum$|go\.work$|go\.work\.sum$'` also returned no matches.

- Observation: `scripts/verify platform` still expects Go testable directories under `services/omnichannel/backend` and `platform-cli`, so fixing CI bootstrap will likely expose a later validation failure unrelated to the dead path.
  Evidence: `scripts/verify` runs `go test ./...` in both `services/omnichannel/backend` and `platform-cli`.

- Observation: the current omnichannel backend tree contains untracked Go source and tests, but still no module manifest.
  Evidence: `find services/omnichannel/backend -maxdepth 4 -type f` showed Go source/test files under `internal/`, with no `go.mod`.

## Decision Log

- Decision: use an explicit `go-version` in GitHub Actions rather than another `go-version-file`.
  Rationale: there is no checked-in Go module or workspace file to reference, so retaining `go-version-file` would keep CI coupled to a nonexistent source of truth.
  Date/Author: 2026-03-10 / Codex

- Decision: scope `ANM-133` to restoring CI bootstrap, not inventing or reconstructing missing Go module manifests.
  Rationale: the user asked to continue implementing `ANM-133`, which specifically tracks the stale workflow path. Missing module manifests are a distinct repository-state issue and should be tracked separately if uncovered by this fix.
  Date/Author: 2026-03-10 / Codex

## Outcomes & Retrospective

Completed outcomes:

- `.github/workflows/ci.yml` no longer references the dead `services/omnichannel/backend/go.mod` file during `actions/setup-go`.
- `.github/workflows/deploy-preflight.yml` no longer references the dead `services/omnichannel/backend/go.mod` file during `actions/setup-go`.
- Both workflows now use `go-version: stable`, which removes the immediate bootstrap failure caused by a nonexistent file path.
- Follow-up ticket `ANM-146` was created to track the separate missing-Go-manifests problem that remains after bootstrap succeeds.

Remaining gaps:

- Remote GitHub Actions reruns have not been executed from this local-only change, so live PR status has not yet been revalidated.
- `scripts/verify platform` still expects Go module manifests that are not currently tracked in the repo; that work is explicitly deferred to `ANM-146`.

## Context And Orientation

Relevant files for this task:

- `.github/workflows/ci.yml`
- `.github/workflows/deploy-preflight.yml`
- `scripts/verify`
- `plans/ci-go-bootstrap-fix-execplan.md`

Affected PRs confirmed from GitHub:

- `#1` `feat: DevEx build system and agentic improvements`
- `#2` `fix(ci): add .nvmrc and correct go-version-file path`
- `#3` `feat: Greptile MCP, Linear tooling, PR template, tech debt removal`
- `#4` `docs(agents): Ralph loop, Linear CLI fallback, mistake logging`
- `#5` `refactor(verify): migrate verify_apps to Nx targets`

## Plan Of Work

1. Replace the dead `go-version-file` usage in CI and deploy preflight with an explicit Go toolchain version.
2. Validate the workflow diffs locally with targeted inspection commands.
3. Record the bootstrap-vs-module-layout distinction in the plan and ticket state.
4. Reconcile follow-up ticket coverage for the remaining Go module/verify gap if no existing issue already covers it.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Inspect current workflow definitions:
   - `sed -n '1,120p' .github/workflows/ci.yml`
   - `sed -n '1,120p' .github/workflows/deploy-preflight.yml`
2. Patch both workflow files to replace `go-version-file` with explicit `go-version`.
3. Validate the updated workflow text:
   - `rg -n "setup-go|go-version|go-version-file" .github/workflows/ci.yml .github/workflows/deploy-preflight.yml`
4. Re-scan for remaining Go module/toolchain gaps and reconcile ticket coverage.

## Validation And Acceptance

Acceptance criteria:

- Neither workflow references `services/omnichannel/backend/go.mod` anymore.
- GitHub Actions Go setup no longer depends on a missing file path.
- `ANM-133` and this plan explicitly record that the bootstrap fix may reveal separate downstream validation issues.

Validation commands:

- `rg -n "setup-go|go-version|go-version-file" .github/workflows/ci.yml .github/workflows/deploy-preflight.yml`
- `git diff -- .github/workflows/ci.yml .github/workflows/deploy-preflight.yml plans/ci-go-bootstrap-fix-execplan.md`

## Idempotence And Recovery

Re-running this task should produce the same explicit Go setup in both workflows. If a canonical toolchain manifest or checked-in `go.mod` later appears, the workflows can be updated again to use that single source of truth.

## Artifacts And Notes

- Linear issue: `ANM-133`
- Session UUID comment added on the issue: `019cd77b-78a6-7141-88aa-27faf2bb18f9`
- Follow-up ticket created from post-change rescan: `ANM-146`

## Interfaces And Dependencies

- GitHub Actions workflow definitions in `.github/workflows/`
- Current repository toolchain policy, which presently lacks a canonical Go version source
- `scripts/verify`, which still defines downstream platform Go test expectations
