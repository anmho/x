# Fix open PR merge blockers and recommend merge order

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-189` ([issue link](https://linear.app/anmho/issue/ANM-189/major-triage-open-pr-failures-fix-merge-blockers-and-order-merge)).

## Purpose / Big Picture

Inspect the currently open GitHub pull requests in `anmho/x`, fix failing required checks where practical, review the PRs in dependency order, and recommend the merge sequence based on live GitHub state.

## Progress

- [x] (2026-03-24 07:24Z) Created and started Linear ticket `ANM-189` before execution.
- [x] (2026-03-24 07:24Z) Enumerated the current open PRs with `gh pr list` and confirmed there are seven open PRs split between an older stacked chain (`#1 -> #3 -> #4 -> #5`) and newer `main`-targeting PRs (`#6`, `#7`).
- [x] (2026-03-24 07:29Z) Collected live failure logs and diff summaries for the open PR set; confirmed the old stack still failed historical `Platform Checks` / `App Checks` / `Docs Checks` while PR `#6` failed the newer `Affected Apps` / `Affected Platform` jobs.
- [x] (2026-03-24 07:31Z) Determined that PR `#6` is the first merge blocker because `main` currently requires `Affected Platform`, `Affected Apps`, `Affected Agents`, and `Affected Docs`, but `origin/main` does not yet contain `.github/workflows/ci.yml`.
- [x] (2026-03-24 07:38Z) Landed targeted fixes on PR `#6`: moved `npm ci` ahead of Nx scope resolution, refreshed the stale root `package-lock.json`, and revalidated the branch via GitHub Actions run `23478336338` (all required checks green).
- [x] (2026-03-24 07:45Z) Reviewed PRs `#6` and `#7`, left explicit security findings on both, and tested whether `#7` can be stacked on `#6`.
- [ ] Summarize final merge recommendation, including which PRs should merge, which should be restacked or closed, and why.

## Surprises & Discoveries

- Observation: the live open PR set now contains both historical stacked PRs on non-`main` bases and newer `main`-targeting PRs, so merge order is constrained by branch topology before code quality.
  Evidence: `gh pr list --state open` shows `#3` based on `fix/top-10-audit-subagent-changes`, `#4` based on `pr/2-agent-tooling`, and `#5` based on `pr/3-agents-policy`, while `#6` and `#7` both target `main`.

- Observation: PR `#6` is the hard gate for the repository because branch protection requires the `Affected *` contexts even though `origin/main` does not yet contain the workflow that defines them.
  Evidence: `gh api repos/anmho/x/branches/main/protection` reports required contexts `Affected Platform`, `Affected Apps`, `Affected Agents`, and `Affected Docs`, while `git show origin/main:.github/workflows/ci.yml` fails because the workflow file is absent from `main`.

- Observation: PR `#6`'s initial failures were workflow-ordering and lockfile drift issues, not application-target failures.
  Evidence: GitHub Actions run `23478103977` first failed because `npx nx show projects ...` executed before dependencies were installed; after moving `npm ci` earlier, the next failure was `npm ci` rejecting an out-of-sync `package-lock.json`; after regenerating the lockfile, run `23478336338` passed all required checks.

- Observation: both PR `#6` and PR `#7` currently contain a security-sensitive startup behavior in the MCP server.
  Evidence: `services/mcp/cmd/server/main.go` auto-generates a default API key when the key store is empty and prints that secret with `fmt.Printf`, which leaks an admin credential into logs; I left review comments on both PRs referencing this finding.

- Observation: PR `#7` does not cleanly stack on top of PR `#6`; the two branches implement overlapping MCP files with conflicting histories.
  Evidence: after changing PR `#7`'s base to the PR `#6` branch, GitHub marked it `DIRTY`, and an isolated `git rebase origin/andyminhtuanho/anm-96-restore-missing-mcp-observability-workspace-and-re-enable` test hit add/add conflicts in `services/mcp/.env.example`, `services/mcp/Dockerfile`, `services/mcp/internal/tools/registry.go`, and `services/mcp/project.json`.

## Decision Log

- Decision: start from live GitHub PR/check state, then fix the most merge-ready blocker rather than patching the oldest branch first by default.
  Rationale: some historical PRs may be obsolete or superseded by newer `main`-targeting work, and branch topology plus current CI state should drive the actual merge order.
  Date/Author: 2026-03-24 / Codex

- Decision: fix PR `#6` before reviewing merge order for the rest of the queue.
  Rationale: until PR `#6` lands, the required-check policy on `main` cannot be satisfied by any other branch because the relevant workflow does not exist on `main`.
  Date/Author: 2026-03-24 / Codex

- Decision: attempt to stack PR `#7` on PR `#6`, but treat merge conflicts there as evidence that `#7` needs restacking or recutting rather than blind sequencing.
  Rationale: a clean stack would have allowed live CI against the new workflow, but the conflict set proves the two branches overlap too heavily to be treated as simple sequential merges.
  Date/Author: 2026-03-24 / Codex

## Outcomes & Retrospective

Completed outcomes:

- Identified the true repository gate: PR `#6`.
- Fixed PR `#6` CI and revalidated it live.
- Left review comments on PRs `#6` and `#7` for the MCP default-key log leak.
- Determined that PR `#7` does not cleanly stack on `#6`.

Remaining gaps:

- The final merge recommendation still needs to be summarized for the user.
- PR `#7` still needs a restack or recut if it is to merge after `#6`.
- The older PR stack (`#1`-`#5`) still appears stale/superseded and likely should not be merged as-is.

## Context And Orientation

Key artifacts for this task:

- `plans/open-pr-merge-blockers-execplan.md`
- `.github/workflows/ci.yml`
- `scripts/verify`
- open PRs `#1` through `#7`

## Plan Of Work

1. Inspect live PR metadata, checks, and failed logs for each open PR.
2. Identify dependency order, superseded branches, and the first practical merge blocker.
3. Apply narrow fixes on the relevant PR branches and validate with GitHub Actions.
4. Review each PR in merge order and recommend the final merge sequence.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Collect live PR metadata:
   - `gh pr list --state open --json number,title,headRefName,baseRefName,mergeStateStatus,reviewDecision,statusCheckRollup,url`
2. Inspect each blocked PR:
   - `gh pr view <number> --json files,commits,mergeable,reviewDecision,statusCheckRollup`
   - `gh pr checks <number>`
   - `gh run view <run-id> --log-failed`
3. Patch only the files needed for the first active blocker.
4. Run focused local validation plus GitHub verification on the updated PR branch.
5. Summarize review findings and merge order.

## Validation And Acceptance

Acceptance criteria:

- Live GitHub blockers are identified for every open PR.
- At least the first actionable merge blocker is fixed and revalidated if practical within this task.
- Open PRs are ordered by actual merge dependency and readiness.
- Review findings are captured against the relevant PRs.

Validation commands:

- `gh pr list --state open --json number,title,headRefName,baseRefName,mergeStateStatus,reviewDecision,statusCheckRollup,url`
- `gh pr checks <number>`
- `gh run view <run-id> --log-failed`
- targeted repo validation commands for the specific fixes landed

## Idempotence And Recovery

Re-running the GitHub inspection steps is safe and should converge on the same open PR topology. If a PR branch is too stale or conflicts heavily with current `main`, document that the merge blocker is rebase debt rather than continuing with broad speculative fixes.

## Artifacts And Notes

- Linear issue: `ANM-189`
- Prior related triage: `plans/open-pr-test-failure-triage-execplan.md`
- Key validation run: `23478336338` (PR `#6` all required checks passing)

## Interfaces And Dependencies

- GitHub PR metadata, checks, and Actions logs
- Repository CI workflow and verification scripts
- Existing open-branch topology and review requirements
