# Audit dirty worktree, map changes to associated work, and tighten PR merge policy

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-131` ([issue link](https://linear.app/anmho/issue/ANM-131/audit-dirty-worktree-map-changes-to-associated-work-and-tighten-pr)).

## Purpose / Big Picture

The repository currently has a large dirty worktree spanning docs, SDK generation, platform/control-plane changes, frontend refactors, and untracked scaffolding. This task does not blindly clean the tree. It identifies likely associated workstreams so the tree can be split safely later, and it updates repository policy so future PRs always carry explicit ticket linkage and use squash merge into `main`.

## Progress

- [x] (2026-03-10 09:34Z) Created `ANM-131` in Linear with scope covering dirty-worktree audit plus PR/merge policy updates.
- [x] (2026-03-10 09:35Z) Moved `ANM-131` to `In Progress`.
- [x] (2026-03-10 09:36Z) Captured current dirty worktree snapshot with `git status --short` and `git diff --stat`.
- [x] (2026-03-10 09:37Z) Audited recent commit history to find likely parent workstreams for tracked dirty files.
- [x] (2026-03-10 09:39Z) Added explicit `AGENTS.md` rules requiring a Linear ticket reference in every PR and squash-and-merge into `main`.
- [ ] Decide whether to convert the grouped workstream findings into follow-up cleanup tickets or branch-by-branch recovery steps.

## Surprises & Discoveries

- Observation: the worktree is not just "a few edits"; it spans at least four distinct workstreams plus a large untracked scaffold set.
  Evidence: `git diff --stat` showed 53 tracked files changed, with large deltas in `package-lock.json`, docs, SDK wiring, and deleted access-api files.

- Observation: several tracked dirty files point back to different historical commits, which strongly suggests the current branch contains mixed leftover work rather than one coherent change set.
  Evidence: last-touch audit mapped files to commits such as `9fcb3e8` (SDK/ConnectRPC), `a960e1f` (docs/verify cleanup), `98a15a0` (cloud-console access-api refactor), `6003d8f` (control-plane domains), and `f8c6c3a`/`92c8659` (access-api tests/ignore files).

- Observation: many untracked directories look like major feature scaffolds rather than throwaway local files.
  Evidence: untracked roots include `apps/omnichannel/`, `cli/`, `platform-config/`, `docs/mintlify/`, `scripts/dev-stack`, `scripts/sdk.sh`, `services/omnichannel/backend-api/`, and `services/omnichannel/backend-worker/`.

## Decision Log

- Decision: do not attempt destructive cleanup or automatic restoration from guessed commits.
  Rationale: the repository policy forbids reverting unrelated changes without explicit instruction, and the current tree clearly mixes multiple workstreams.
  Date/Author: 2026-03-10 / Codex

- Decision: classify the tree into likely work buckets using last-touch commit history plus directory intent.
  Rationale: grouping by workstream gives the user a safe path to split/commit/recover later without fabricating certainty about every file.
  Date/Author: 2026-03-10 / Codex

- Decision: express PR linkage and merge-strategy policy directly in `AGENTS.md`.
  Rationale: the repo already has PR discipline rules; adding explicit "ticket required in every PR" and "squash merge to main" guidance makes the expected workflow unambiguous.
  Date/Author: 2026-03-10 / Codex

## Outcomes & Retrospective

Completed outcomes:

- Dirty worktree was audited and grouped into likely workstreams.
- `AGENTS.md` now requires each PR to reference its Linear ticket and standardizes squash-and-merge into `main`.

Remaining gaps:

- The tree is not yet "clean"; it is only classified.
- Some untracked directories cannot be tied to a specific commit because they are not in git history yet.
- A safe cleanup still requires choosing whether to:
  - commit grouped workstreams separately,
  - move groups onto new branches,
  - or discard specific buckets with explicit user approval.

## Context And Orientation

Relevant files for this task:

- `AGENTS.md`
- `plans/dirty-worktree-audit-pr-merge-policy-execplan.md`

Audit commands used:

- `git status --short`
- `git diff --stat`
- `git log --oneline --decorate -n 25`
- per-path `git log --oneline -n 1 -- <path>`

## Plan Of Work

1. Snapshot the current dirty tree.
2. Map tracked paths to likely parent commits and workstreams.
3. Record grouped findings in this plan.
4. Update PR/merge policy in `AGENTS.md`.
5. Recommend the safest cleanup order without reverting anything automatically.

## Concrete Steps

### Dirty tree workstream buckets

1. Docs and verification policy cleanup
   Likely associated commits:
   - `a960e1f` `docs: remove stack.sh refs, add linear backlog runbook, fix create-prs gh check`
   - `9047557` `fix(ci): add .nvmrc and correct go-version-file path`
   - `3b33e19` `feat: Greptile MCP, Linear tooling, PR template, tech debt removal`
   - `b1c6816` `docs(agents): Ralph loop, Linear CLI fallback, mistake logging`

   Representative files:
   - `AGENTS.md`
   - `Makefile`
   - `docs/runbooks/*`
   - `scripts/README.md`
   - `scripts/verify`
   - backlog JSON payloads

2. ConnectRPC / SDK / Nx wiring
   Likely associated commit:
   - `9fcb3e8` `feat(sdk): add Nx task graph and local ConnectRPC ES generation`

   Representative tracked files:
   - `nx.json`
   - `apps/cloud-console/package.json`
   - `apps/cloud-console/project.json`
   - `services/omnichannel/project.json`
   - `services/omnichannel/scripts/sdk.sh`
   - `services/omnichannel/backend/proto/README.md`

   Representative untracked files:
   - `.github/workflows/publish-connectrpc-sdks.yml`
   - `scripts/sdk-core.sh`
   - `scripts/sdk.sh`
   - `services/omnichannel/backend/proto/buf.yaml`
   - `services/omnichannel/backend/proto/temporal/`
   - `services/omnichannel/backend/internal/rpc/`
   - `services/omnichannel/backend-api/`
   - `services/omnichannel/backend-worker/`

3. Cloud-console / access-api refactor drift
   Likely associated commits:
   - `98a15a0` `feat(cloud-console): extract access-api helper, remove hardcoded fallback`
   - `ab622ce` `fix(cloud-console): fix CreateNotificationRequest build blocker`
   - `f9c0010` `perf(cloud-console): memoize HeadlessSelect options to fix React DevTools warning`
   - `bb79c18` `fix(cloud-console): port omnichannel runtime pages from service frontend`

   Representative files:
   - deleted `apps/cloud-console/app/api/keys/*`
   - deleted `apps/cloud-console/lib/access-api.ts`
   - deleted `apps/cloud-console/lib/access-keys.ts`
   - modified `apps/cloud-console/app/deployments/page.tsx`
   - modified `apps/cloud-console/app/settings/page.tsx`
   - modified `apps/cloud-console/app/_components/app-nav.tsx`

4. Control-plane / domains / platform scaffolding
   Likely associated commits:
   - `6003d8f` `control-plane: add domain schema and provider contracts`
   - `d089e8f` `control-plane: add domains HTTP API and serve mode`

   Representative tracked files:
   - `infra/platform/declarative_spec.py`
   - `platform-cli/main.go`
   - `platform.controlplane.json`
   - `platform.controlplane.example.json`

   Representative untracked files:
   - `platform`
   - `platform-config/`
   - `project.json`
   - `scripts/dev-stack`
   - `scripts/dev-stack-discover.py`
   - `scripts/setup.sh`
   - `apps/cloud-console/stack.json`
   - `services/omnichannel/stack.json`

5. Access-api retirement or churn
   Likely associated commits:
   - `f8c6c3a` `test(access-api): add auth and key lifecycle tests`
   - `92c8659` `hygiene: add access-api local ignore policy`
   - `2f5df15` `fix(schema): add response schema for POST /v1/policy/check`

   Representative files:
   - deleted `services/access-api/.gitignore`
   - deleted `services/access-api/cmd/api/main_test.go`
   - deleted `schemas/access-api/openapi.yaml`

### Recommended cleanup sequence

1. Separate tracked changes from untracked scaffolding.
2. Decide whether the SDK/ConnectRPC bucket is meant to land as one coherent branch.
3. Decide whether access-api deletions are intentional retirement or accidental leftovers.
4. Split docs/policy cleanup from product/runtime code changes.
5. Only after grouping, create or recover dedicated branches/PRs per workstream.

## Validation And Acceptance

Acceptance criteria:

- `AGENTS.md` explicitly requires ticket reference in every PR.
- `AGENTS.md` explicitly standardizes squash-and-merge into `main`.
- The dirty tree is grouped into plausible workstreams with commit evidence where available.

Validation executed:

- `git status --short`
- `git diff --stat`
- `git log --oneline --decorate -n 25`
- per-path last-touch audit via `git log --oneline -n 1 -- <path>`

## Idempotence And Recovery

This task is documentation/policy plus audit only. Re-running it is safe and will only refresh grouping evidence. No destructive cleanup was performed.

## Artifacts And Notes

- Linear issue: `ANM-131`
- Plan file: `plans/dirty-worktree-audit-pr-merge-policy-execplan.md`

## Interfaces And Dependencies

- Depends on current git history being available locally.
- Depends on repo policy in `AGENTS.md`.
