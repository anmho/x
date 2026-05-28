# Recut oversized PR branches into a topological stack

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-197` ([issue link](https://linear.app/anmho/issue/ANM-197/major-split-overstuffed-pr-branches-into-logical-mergeable-prs-in-x)).

## Purpose / Big Picture

Replace the overloaded open PR branches with a smaller, topologically ordered stack that is easier to review and merge. The first branch must establish the repository/tooling foundation required for later CI and feature branches to pass.

## Progress

- [x] (2026-03-24 07:50Z) Created and started `ANM-197` before beginning the split analysis.
- [x] (2026-03-24 07:50Z) Audited `origin/main` and confirmed it is much older than the current open PR branches, with no `.github/workflows/ci.yml`, no `.nvmrc`, no `nx.json`, and no MCP workspace.
- [x] (2026-03-24 07:51Z) Collected the full commit range for PR `#6` and PR `#7` to identify stack candidates and overlap.
- [x] (2026-03-24 07:52Z) Identified the first required split boundary: repo/tooling foundation must precede the required-check CI rollout, which must precede MCP or UI feature work.
- [ ] Create a clean branch stack from `main` in topological order.
- [ ] Move the CI-unblock work onto the second branch in the stack and verify it still passes.
- [ ] Recut MCP work onto a later stack branch, excluding security issues or unrelated docs drift where possible.
- [ ] Decide which legacy PRs should be replaced or closed after the new stack exists.

## Surprises & Discoveries

- Observation: `origin/main` is only at commit `4ad2bb9` and lacks most of the infrastructure assumed by the newer PR branches.
  Evidence: `git ls-tree -r --name-only origin/main` shows no `.github/workflows/ci.yml`, no `.nvmrc`, no `nx.json`, no `docs/mintlify`, and no `services/mcp`.

- Observation: the current required-check rollout depends on infrastructure that is not present on `main`, so it cannot be the first stack layer.
  Evidence: the required-check workflow on PR `#6` assumes `.nvmrc`, `package-lock.json`, `nx.json`, and relevant project metadata; these are absent from `origin/main`.

- Observation: the most coherent early “foundation” commits are not UI or MCP commits; they are the dependency/workspace and Nx/task-graph commits.
  Evidence: `498ed5b` adds root lockfile plus repo verification scripts and workflows, while `9fcb3e8` adds `nx.json`, project metadata, SDK generation, and Nx-aware verify paths.

- Observation: PR `#7` is not a simple child of PR `#6`; it is a parallel branch from the same old base that overlaps heavily in MCP files.
  Evidence: an isolated rebase of `origin/andyminhtuanho/anm-163-mcp-validate-pr` onto the PR `#6` branch hit add/add conflicts in core MCP files including `services/mcp/.env.example`, `services/mcp/Dockerfile`, `services/mcp/internal/tools/registry.go`, and `services/mcp/project.json`.

## Decision Log

- Decision: recut the work as a topological stack instead of trying to keep the existing PR numbers alive.
  Rationale: the current PR branches overlap too heavily, have misleading names, and are not cleanly stackable.
  Date/Author: 2026-03-24 / Codex

- Decision: use this order for the new stack:
  1. repo/tooling foundation
  2. required-check CI rollout
  3. feature branches such as MCP service/platform wiring
  4. remaining UI/docs/policy slices that are independent of the feature branches
  Rationale: later layers assume files and tooling introduced by the earlier ones.
  Date/Author: 2026-03-24 / Codex

## Outcomes & Retrospective

Completed outcomes:

- Established the dependency shape for a cleaner replacement stack.

Remaining gaps:

- The new stack branches have not yet been created.
- Legacy PR replacement/closure decisions remain pending.

## Context And Orientation

Key artifacts for this task:

- `plans/pr-stack-recut-execplan.md`
- `plans/open-pr-merge-blockers-execplan.md`
- `origin/main`
- `origin/andyminhtuanho/anm-96-restore-missing-mcp-observability-workspace-and-re-enable`
- `origin/andyminhtuanho/anm-163-mcp-validate-pr`

## Plan Of Work

1. Build a small foundation branch from `main` that introduces the Node/workspace and Nx prerequisites later CI depends on.
2. Build a second branch on top containing the required-check CI rollout and related verification assets.
3. Recut MCP work onto a later branch after the CI base exists.
4. Re-evaluate which old PRs can be closed as superseded.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Audit main vs. candidate commits:
   - `git ls-tree -r --name-only origin/main`
   - `git show --stat --summary <commit>`
2. Create worktrees/branches for the new stack.
3. Cherry-pick or reconstruct coherent slices onto each branch.
4. Validate each branch locally and/or on GitHub.
5. Update PR ordering and supersession notes.

## Validation And Acceptance

Acceptance criteria:

- The new stack is ordered so each branch only depends on earlier branches.
- The first two branches establish a mergeable path for current branch protection.
- Later feature branches avoid reintroducing the full `#6`/`#7` sprawl.

Validation commands:

- `git log --oneline --reverse 4ad2bb9..origin/andyminhtuanho/anm-96-restore-missing-mcp-observability-workspace-and-re-enable`
- `git log --oneline --reverse 4ad2bb9..origin/andyminhtuanho/anm-163-mcp-validate-pr`
- `git diff --stat <base>...<branch>`
- targeted local and GitHub CI validation for each recut branch

## Idempotence And Recovery

If a branch boundary proves wrong, drop the experimental stack branch and recreate it from the prior stack base. Do not mutate the dirty main checkout while recutting; use isolated worktrees for each stack layer.

## Artifacts And Notes

- Linear issue: `ANM-197`
- Candidate base commits:
  - `498ed5b` Node/root-lockfile workflow foundation
  - `9fcb3e8` Nx and SDK task-graph foundation
  - `fe2f381` onward for required-check CI rollout

## Interfaces And Dependencies

- Git branch topology and worktree isolation
- GitHub branch protection requirements
- Root Node workspace, Nx configuration, and CI workflow definitions
