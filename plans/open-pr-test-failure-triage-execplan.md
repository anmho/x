# Triage open PR test failures and codify GitHub PR inspection workflow

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

`PLANS.md` is checked into this repository, and this document follows its required structure.

Linear ticket linkage for this plan: `ANM-129` ([issue link](https://linear.app/anmho/issue/ANM-129/medium-triage-open-pr-test-failures-and-codify-github-workflow)).

## Purpose / Big Picture

Determine why the currently open GitHub pull requests are failing CI, capture the actual root causes from GitHub checks, and update repository agent policy so PR/check evaluation prefers GitHub MCP tools or `gh` CLI instead of indirect local guesswork.

## Progress

- [x] (2026-03-10 09:32Z) Created and started Linear ticket `ANM-129` before execution.
- [x] (2026-03-10 09:33Z) Read recent ExecPlans to match required section structure.
- [x] (2026-03-10 09:34Z) Enumerated open PRs with `gh pr list`; found open PRs `#1` through `#5`.
- [x] (2026-03-10 09:35Z) Collected current check summaries with `gh pr checks` for each open PR.
- [x] (2026-03-10 09:38Z) Inspected failed GitHub Actions logs with `gh run view --log-failed` for the active PR runs.
- [x] (2026-03-10 09:41Z) Confirmed repo-side path/config mismatches causing the repeated failures.
- [x] (2026-03-10 09:37Z) Updated `AGENTS.md` to prefer GitHub MCP tools or `gh` CLI for PR/check inspection work.
- [x] (2026-03-10 09:38Z) Ran focused post-change rescan and reconciled follow-up issues/tickets within this scope.
- [x] (2026-03-10 09:39Z) Aligned `ANM-129` status, final outcomes, and validation evidence.

## Surprises & Discoveries

- Observation: All five open PRs fail with the same three job families rather than branch-specific test regressions.
  Evidence: `gh pr checks 1`, `gh pr checks 2`, `gh pr checks 3`, `gh pr checks 4`, and `gh pr checks 5` all reported `App Checks`, `Docs Checks`, and `Platform Checks` as failed while `Greptile Review` passed.

- Observation: `Platform Checks` fails before tests run because GitHub Actions still points `actions/setup-go` at `services/omnichannel/backend/go.mod`, which no longer exists.
  Evidence: `gh run view 22889320776 --log-failed` and sibling PR runs show `The specified go version file at: services/omnichannel/backend/go.mod does not exist`; `.github/workflows/ci.yml` contains `go-version-file: services/omnichannel/backend/go.mod`.

- Observation: `Docs Checks` fails because `scripts/verify docs` still reads `docs/mintlify/mint.json`, but the docs workspace now contains `docs/mintlify/docs.json` instead.
  Evidence: GitHub logs show `ENOENT: no such file or directory, open '/home/runner/work/x/x/docs/mintlify/mint.json'`; local file listing shows `docs/mintlify/docs.json` and no `docs/mintlify/mint.json`.

- Observation: `App Checks` fails because `sdk:generate-es` runs `buf generate` from `backend/proto`, which contains no `.proto` files relative to repo root.
  Evidence: GitHub logs show `cd backend/proto && buf generate --template buf.gen.client.yaml` followed by `Module "path: "." had no .proto files`; the actual proto path in the repo is `services/omnichannel/backend/proto`.

- Observation: Older PR runs also included a bootstrap failure where CI could not find `package-lock.json` / `./scripts/verify`, indicating the PR stack spans multiple historical workflow states.
  Evidence: `gh run view 22834784419 --log-failed` shows `Some specified paths were not resolved, unable to cache dependencies` and `./scripts/verify: No such file or directory`.

- Observation: Existing Linear coverage already matched the app-side `sdk:generate-es` blocker, but not the docs-config or stale workflow Go-path failures.
  Evidence: `ANM-87` covers the proto-generation blocker; new tickets `ANM-132` and `ANM-133` were required for the other two CI failures.

## Decision Log

- Decision: Use live GitHub check/log inspection rather than local repro first.
  Rationale: The user explicitly asked to evaluate open PR failures and preferred GitHub tooling; CI failures can differ from local state because these PRs target different base branches.
  Date/Author: 2026-03-10 / Codex

- Decision: Treat the current failures as shared CI/workflow drift across the PR stack, not five independent application bugs.
  Rationale: The same failing job families and path errors recur across all open PRs, with only older PRs showing additional historical bootstrap issues.
  Date/Author: 2026-03-10 / Codex

- Decision: Limit code changes in this task to policy/process documentation unless the user asks to fix CI.
  Rationale: The request was to evaluate failures and update `AGENTS.md`, not to land workflow repairs across a dirty worktree.
  Date/Author: 2026-03-10 / Codex

## Outcomes & Retrospective

Completed outcomes:

- Open PR failure causes were identified from GitHub checks/logs rather than inferred locally.
- Shared CI path/config mismatches were mapped to concrete repository files.
- `AGENTS.md` now explicitly requires GitHub MCP / `gh`-first PR and CI inspection workflow.
- Follow-up ticket coverage now exists for each distinct blocker:
  - `ANM-87` for the app-side `sdk:generate-es` failure
  - `ANM-132` for the docs `mint.json` / `docs.json` mismatch
  - `ANM-133` for the stale `go-version-file` workflow path

Remaining gaps:

- The CI failures themselves are diagnosed but not fixed in this task.

## Context And Orientation

Key artifacts for this task:

- `.github/workflows/ci.yml`
- `.github/workflows/deploy-preflight.yml`
- `scripts/verify`
- `scripts/sdk-core.sh`
- `docs/mintlify/project.json`
- `plans/open-pr-test-failure-triage-execplan.md`
- `AGENTS.md`

GitHub PRs inspected:

- `#1` `feat: DevEx build system and agentic improvements`
- `#2` `fix(ci): add .nvmrc and correct go-version-file path`
- `#3` `feat: Greptile MCP, Linear tooling, PR template, tech debt removal`
- `#4` `docs(agents): Ralph loop, Linear CLI fallback, mistake logging`
- `#5` `refactor(verify): migrate verify_apps to Nx targets`

## Plan Of Work

1. Use GitHub tooling to collect open PR and failing check details.
2. Cross-check the failing paths/commands against the current repository layout.
3. Update `AGENTS.md` with explicit GitHub MCP / `gh` preference for PR triage and check inspection.
4. Run a scoped follow-up audit and reconcile any new issues through Linear.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Enumerate open PRs:
   - `gh pr list --state open --limit 20 --json number,title,headRefName,author,baseRefName,url`
2. Review failing checks:
   - `gh pr checks 1`
   - `gh pr checks 2`
   - `gh pr checks 3`
   - `gh pr checks 4`
   - `gh pr checks 5`
3. Inspect failing logs:
   - `gh run view 22889320776 --log-failed`
   - `gh run view 22834784419 --log-failed`
   - `gh run view 22889057784 --log-failed`
   - `gh run view 22889057437 --log-failed`
   - `gh run view 22834801724 --log-failed`
4. Confirm offending paths/commands locally:
   - `sed -n '1,240p' .github/workflows/ci.yml`
   - `sed -n '1,260p' scripts/verify`
   - `sed -n '120,180p' scripts/sdk-core.sh`
   - `rg --files docs/mintlify services/omnichannel`
5. Patch `AGENTS.md` with GitHub workflow guidance and validate with targeted `rg`.

## Validation And Acceptance

Acceptance criteria:

- Open PR failures are explained with concrete GitHub check evidence.
- Root causes are mapped to specific repo paths/commands.
- `AGENTS.md` explicitly says to prefer GitHub MCP tools or `gh` CLI when evaluating PRs/checks.
- Plan progress and Linear ticket state remain synchronized.

Validation commands:

- `gh pr list --state open --limit 20 --json number,title,headRefName,author,baseRefName,url`
- `gh pr checks 1`
- `gh pr checks 2`
- `gh pr checks 3`
- `gh pr checks 4`
- `gh pr checks 5`
- `rg -n "GitHub MCP|gh CLI|gh pr|gh run|PR/check" AGENTS.md`

## Idempotence And Recovery

Re-running the GitHub inspection commands is safe and should reflect current CI state. The `AGENTS.md` update is additive documentation; if the wording duplicates an existing rule, collapse it into one explicit instruction and rerun the targeted `rg` validation.

## Artifacts And Notes

- Linear issue: `ANM-129`
- GitHub check inspection used `gh pr list`, `gh pr checks`, and `gh run view --log-failed`
- Current shared CI failures:
  - App Checks: `sdk:generate-es` runs from the wrong proto directory.
  - Docs Checks: verify script expects `docs/mintlify/mint.json` instead of the present config file.
  - Platform Checks: `actions/setup-go` points at removed `services/omnichannel/backend/go.mod`.
- Ticket reconciliation:
  - Existing: `ANM-87` covers the app failure.
  - New: `ANM-132` covers the docs failure.
  - New: `ANM-133` covers the platform failure.

## Interfaces And Dependencies

- GitHub CLI authentication and GitHub API reachability
- GitHub Actions workflow definitions in `.github/workflows/`
- Verification scripts in `scripts/`
- Current repo layout under `docs/mintlify/` and `services/omnichannel/`

Revision note (2026-03-10): Initial plan created for GitHub PR failure triage and `AGENTS.md` workflow policy update.
Revision note (2026-03-10): Updated with AGENTS validation and follow-up ticket reconciliation (`ANM-87`, `ANM-132`, `ANM-133`).
