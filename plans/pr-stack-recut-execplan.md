# Recut oversized PR branches into a topological stack

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-197` ([issue link](https://linear.app/anmho/issue/ANM-197/major-split-overstuffed-pr-branches-into-logical-mergeable-prs-in-x)), follow-up correction `ANM-202` ([issue link](https://linear.app/anmho/issue/ANM-202/medium-remove-generated-omnichannel-sdk-artifacts-from-stack)).

## Purpose / Big Picture

Replace the overloaded open PR branches with a smaller, topologically ordered stack that is easier to review and merge. The first branch establishes the missing Node workspace, root lockfile, docs verifier, and stable GitHub check names on top of the old `main` branch without inheriting unrelated MCP or settings-page work.

## Progress

- [x] (2026-03-24 07:50Z) Created and started `ANM-197` before beginning the split analysis.
- [x] (2026-03-24 07:50Z) Audited `origin/main` and confirmed it is much older than the current open PR branches, with no `.github/workflows/ci.yml`, no `.nvmrc`, no `nx.json`, and no MCP workspace.
- [x] (2026-03-24 07:51Z) Collected the full commit range for PR `#6` and PR `#7` to identify stack candidates and overlap.
- [x] (2026-03-24 07:52Z) Identified the first required split boundary: repo/tooling foundation must precede the required-check CI rollout, which must precede MCP or UI feature work.
- [x] (2026-03-24 08:22Z) Created `stack/1-ci-foundation` from `origin/main` and recut branch 1 as a pre-Nx foundation layer.
- [x] (2026-03-24 08:30Z) Replaced the copied Nx-based CI workflow with stable `Affected Platform`, `Affected Apps`, `Affected Docs`, and `Affected Agents` checks that only validate files present in this branch.
- [x] (2026-03-24 08:34Z) Validated branch 1 locally with `./scripts/verify platform`, `./scripts/verify apps`, `./scripts/verify docs`, and `./scripts/verify agents`.
- [x] (2026-03-24 08:37Z) Created `ANM-202` after confirming branch 1 still incorrectly carried generated omnichannel SDK artifacts and a local workspace dependency only added to support them.
- [x] (2026-03-24 08:16Z) Created `stack/2-nx-affected` on top of branch 1 and added `nx.json`, app/docs project metadata, and cloud-console TypeScript config.
- [x] (2026-03-24 08:19Z) Validated branch 2 locally with `./scripts/verify platform`, `./scripts/verify apps`, `./scripts/verify docs`, and `./scripts/verify agents`.
- [x] (2026-03-24 08:41Z) Patched branch 1 to remove `packages/sdk-omnichannel`, pushed the fix to PR `#8`, and started restacking branch 2 on top of the cleaned foundation.
- [x] (2026-03-24 08:47Z) Completed the branch-2 rebase onto the cleaned foundation, rebuilt the lockfile from scratch, and revalidated the Nx-based app/docs checks.
- [ ] Recut MCP work onto a later stack branch and address the open Greptile/Cursor security findings there.
- [ ] Decide which legacy PRs should be replaced or closed after the new stack exists.

## Surprises & Discoveries

- Observation: `origin/main` is only at commit `4ad2bb9` and lacks most of the infrastructure assumed by the newer PR branches.
  Evidence: `git ls-tree -r --name-only origin/main` shows no `.github/workflows/ci.yml`, no `.nvmrc`, no `nx.json`, no `docs/mintlify`, and no `services/mcp`.

- Observation: the current required-check rollout depends on infrastructure that is not present on `main`, so it cannot be the first stack layer in its current form.
  Evidence: the workflow on PR `#6` assumes `.nvmrc`, `package-lock.json`, `nx.json`, and multiple project metadata files that do not exist on `origin/main`.

- Observation: `origin/main` also lacked `apps/cloud-console/package.json`, so the first stack layer had to import a minimal app workspace manifest rather than only CI files.
  Evidence: `git show origin/main:apps/cloud-console/package.json` fails while the app source tree exists.

- Observation: Greptile/Cursor comments were useful boundary markers for the split.
  Evidence: PR `#6` comments target the settings/Auth tab work and misleading `project:registry` change, while PR `#7` comments target MCP auth, repo-root detection, exit-code propagation, and secret leakage. None of those changes belong in branch 1.

- Observation: branch 1 still incorrectly included `packages/sdk-omnichannel` and generated Temporal client artifacts, even though the repo's proto policy says client SDKs should be published from BSR and consumed as versioned dependencies.
  Evidence: `services/omnichannel/backend/proto/README.md` explicitly says "SDKs are published from BSR and consumed as versioned dependencies" and "Do not manually version checked-in generated client SDKs in this repo," while PR `#8` contains `packages/sdk-omnichannel/*`.

- Observation: cloud-console can build successfully on top of branch 1 once Nx metadata and the app `tsconfig.json` are added, without importing the settings/Auth tab work from PR `#6`.
  Evidence: `./scripts/verify apps` passes on `stack/2-nx-affected` and builds the existing routes from `origin/main`.

## Decision Log

- Decision: recut the work as a topological stack instead of trying to keep the existing PR numbers alive.
  Rationale: the current PR branches overlap too heavily, have misleading names, and are not cleanly stackable.
  Date/Author: 2026-03-24 / Codex

- Decision: branch 1 should be a pre-Nx foundation layer with stable required-check names, not a direct transplant of PR `#6`'s `nx affected` workflow.
  Rationale: `main` lacks the project graph and service files needed for the later CI shape. A lighter first layer gets stable checks onto `main` without dragging in unrelated feature work.
  Date/Author: 2026-03-24 / Codex

- Decision: exclude settings-page/Auth-tab changes and all MCP service code from branch 1.
  Rationale: those areas are the source of the live Greptile/Cursor review findings and are logically later stack layers.
  Date/Author: 2026-03-24 / Codex

- Decision: remove `packages/sdk-omnichannel` from the foundation stack layers.
  Rationale: generated omnichannel client artifacts are not CI/bootstrap infrastructure and conflict with the repository's publish-first SDK policy.

- Decision: make branch 2 app/docs-focused Nx rollout instead of importing the omnichannel SDK-generation slice now.
  Rationale: the proto/service inputs for that slice are still absent on top of branch 1, but app/docs `affected` checks can already be made real with much smaller scope.
  Date/Author: 2026-03-24 / Codex

## Outcomes & Retrospective

Completed outcomes:

- Cut a new first stack branch from `main` that introduces the missing root Node workspace, root lockfile, docs verifier, and stable CI check names.
- Kept branch 1 scoped away from the known PR `#6` settings-page review issues and the PR `#7` MCP security issues.

Remaining gaps:

- Branch 3 still needs to absorb MCP wiring while fixing constant-time key comparison, fail-closed auth, repo-root detection, exit-code propagation, and secret logging.
- Nx may still need a small follow-up cleanup around inferred outputs or graph metadata after the SDK artifact removal settles.

## Context And Orientation

Key artifacts for this task:

- `plans/pr-stack-recut-execplan.md`
- `origin/main`
- `origin/andyminhtuanho/anm-96-restore-missing-mcp-observability-workspace-and-re-enable`
- `origin/andyminhtuanho/anm-163-mcp-validate-pr`
- `/tmp/x-stack-1-ci-foundation`

## Plan Of Work

1. Land a small foundation branch from `main` that adds the missing workspace manifests, root lockfile, docs verifier, and stable CI check names.
2. Build a second branch on top containing `nx.json`, project metadata, and the `nx affected` required-check rollout.
3. Recut MCP work onto a later branch after the CI base exists, fixing the open auth/security findings as part of that slice.
4. Re-evaluate which old PRs can be closed as superseded.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Create the stack worktree from `origin/main`.
2. Copy only the minimum foundation files needed from the oversized branch.
3. Trim root workspaces and scripts to match what branch 1 actually contains.
4. Validate the branch locally.
5. Commit and push branch 1, then cut branch 2 from it.

## Validation And Acceptance

Acceptance criteria:

- Branch 1 is independently understandable and materially smaller than PR `#6`.
- Branch 1 emits the stable required check names already enforced on `main`.
- Branch 1 avoids carrying known review-problematic UI and MCP changes.

Validation commands:

- `npm install --package-lock-only --ignore-scripts`
- `./scripts/verify platform`
- `./scripts/verify apps`
- `./scripts/verify docs`
- `./scripts/verify agents`

## Idempotence And Recovery

If the branch-1 boundary proves wrong, recreate `stack/1-ci-foundation` from `origin/main` and replay only the minimal workspace, docs, and CI foundation files. Do not mutate the dirty main checkout while recutting; use isolated worktrees for each stack layer.

## Artifacts And Notes

- Linear issue: `ANM-197`
- Branch 1 worktree: `/tmp/x-stack-1-ci-foundation`
- Live review findings informing branch boundaries:
  - PR `#6` Cursor/Greptile comments on settings page, external link handling, URL state sync, duplicated Google Cloud card, and `project:registry`
  - PR `#7` Greptile comments on timing-safe key comparison, fail-open auth, repo-root discovery, exit-code propagation, and secret handling

## Interfaces And Dependencies

- Git branch topology and worktree isolation
- GitHub branch protection requirements
- Root Node workspace and package-lock integrity
- Mintlify docs navigation verification
