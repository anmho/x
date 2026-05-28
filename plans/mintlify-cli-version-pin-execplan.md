# Pin Mintlify CLI version for docs workflows

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-104` ([issue link](https://linear.app/anmho/issue/ANM-104/medium-pin-mintlify-cli-version-for-docs-commands)).

## Purpose / Big Picture

Mintlify commands currently run through unpinned `npx mint ...` invocations. This plan pins the CLI version across docs development and verification paths so command behavior remains deterministic.

## Progress

- [x] (2026-03-10 09:10Z) Confirmed `ANM-104` scope and moved ticket state to `In Progress`.
- [x] (2026-03-10 09:13Z) Audited docs command paths in `package.json`, `docs/mintlify/project.json`, and `scripts/verify`.
- [x] (2026-03-10 09:13Z) Updated docs tooling commands to use pinned `mint@4.2.420`.
- [x] (2026-03-10 09:18Z) Ran validation commands and captured environment constraints/results.
- [x] (2026-03-10 09:18Z) Completed post-change bug/unfinished audit, fixed one remaining unpinned docs command (`docs/mintlify/stack.json`), and confirmed no additional follow-up items.
- [x] (2026-03-10 09:18Z) Moved `ANM-104` to `Done` after implementation, validation capture, and audit reconciliation.

## Surprises & Discoveries

- Observation: The local shell can execute `npx mint` from cache, but direct npm registry resolution is blocked in this environment.
  Evidence: `npx --yes mint@latest version` fails with `ENOTFOUND registry.npmjs.org`.

- Observation: Cached CLI package metadata is available under `~/.npm/_npx/...`, including a concrete installed version.
  Evidence: `~/.npm/_npx/6d81244766d363b7/node_modules/mint/package.json` contains `\"version\": \"4.2.420\"`.

- Observation: The docs stack launcher had an additional unpinned Mintlify invocation outside the initially listed files.
  Evidence: `docs/mintlify/stack.json` originally contained `\"run\": \"npx mint dev\"`.

## Decision Log

- Decision: Pin all docs-facing Mintlify commands to `mint@4.2.420` via explicit versioned `npx` calls.
  Rationale: This keeps invocation deterministic without adding another workspace/package-manager install path for `docs/mintlify`.
  Date/Author: 2026-03-10 / Codex

## Outcomes & Retrospective

Completed outcomes:

- Root docs dev script now uses versioned Mintlify CLI invocation.
- Docs Nx project targets now use versioned Mintlify CLI invocation.
- Docs stack launcher now uses versioned Mintlify CLI invocation.
- Docs verification now checks pinned CLI version with the same deterministic package reference.

Remaining gaps:

- Full docs verification in this shell remains constrained by repository Node version requirement (`>=24.14.0`) when current runtime is lower.
- npm registry access is blocked in this sandbox, so explicit version fetch checks cannot be fully validated against a live registry.

## Context And Orientation

Relevant files:

- `package.json`
- `docs/mintlify/project.json`
- `docs/mintlify/stack.json`
- `scripts/verify`

## Plan Of Work

1. Pin Mintlify CLI usage in all docs tooling entrypoints.
2. Validate `docs:dev`/docs verify command paths use the same pinned CLI version.
3. Audit for additional docs tooling gaps and reconcile with existing/new Linear tickets.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Replace unpinned `npx mint ...` calls with `npx --yes mint@4.2.420 ...`.
2. Run:
   - `npm run docs:dev` (startup smoke check)
   - docs verification command active at execution time (environment permitting)
3. Audit with:
   - `rg -n \"npx mint|mint@|TODO|FIXME|TBD|XXX\" package.json scripts/verify docs/mintlify -S`
4. Reconcile findings with existing Linear issues before creating any new follow-ups.

## Validation And Acceptance

Acceptance criteria:

- Docs command entrypoints no longer invoke unpinned `npx mint`.
- Root docs script and verify script reference the same pinned Mintlify CLI version.

Validation executed:

- `timeout 70 npm run docs:dev` (startup smoke check timed out without reintroducing prior missing-config failure)
- docs verification command active at execution time (failed early due Node version guard requiring `>=24.14.0`)
- `cd docs/mintlify && npx --yes mint@4.2.420 version` (fails in this sandbox with `ENOTFOUND registry.npmjs.org`)
- `rg -n "npx mint" package.json scripts/verify docs/mintlify -S` (no unpinned invocations remain)
- `rg -n "TODO|FIXME|TBD|XXX" package.json scripts/verify docs/mintlify -S` (no unfinished markers in scoped files)

## Idempotence And Recovery

If docs tooling drifts, reapply this plan by restoring consistent pinned `mint@4.2.420` invocations across root scripts, Nx targets, and docs verification checks.

## Artifacts And Notes

- Linear issue: `ANM-104`
- Plan file: `plans/mintlify-cli-version-pin-execplan.md`

## Interfaces And Dependencies

- Mintlify CLI (`mint`) invoked through `npx` with an explicit version.
- Node runtime/version gate enforced by `scripts/verify`.
