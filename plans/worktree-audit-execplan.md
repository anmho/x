# Worktree Audit — Logical Bucket Groupings

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-131` ([issue link](https://linear.app/anmho/issue/ANM-131/audit-dirty-worktree-map-changes-to-associated-work-and-tighten-pr)).

## Purpose / Big Picture

Document every modified and untracked path in the dirty worktree of `/Users/andrewho/repos/projects/x`, group them into actionable buckets, and associate each bucket with the most likely Linear ticket and in-flight plan. No source files are modified by this audit; it is documentation only.

---

## AGENTS.md Squash-Merge Policy Check

**Satisfied.** `AGENTS.md` line 216 already reads:

> Default merge strategy is **Squash and Merge** into `main`. Do not use merge commits or rebase merges unless the user explicitly asks for an exception.

This satisfies the ANM-131 requirement without further changes.

---

## Progress

- [x] (2026-03-17) Created `plans/worktree-audit-execplan.md` with full bucket groupings derived from git status snapshot, plan files, and commit history.
- [x] (2026-03-18 06:53Z) Reconciled the audit buckets against current Linear coverage and created missing follow-up tickets `ANM-164`, `ANM-165`, and `ANM-166`.
- [ ] Owner to decide which buckets to: (a) commit as-is, (b) split onto new branches, or (c) clean up.
- [ ] Decide whether ANM-131 should land via a clean standalone branch/PR or remain documented as a ticket-reconciliation/audit step on top of the existing stacked branch history.

---

## Surprises & Discoveries

- `cli/` contains its own `.git` directory — it is a nested Rust repository with a full `target/` build cache. It appears entirely as an untracked root to the outer repo.
- `apps/omnichannel/` shows as a top-level untracked directory but contains no tracked files — it was empty or its contents are not surfaced under the outer git index.
- `agents/` is untracked and contains `pm-agent` and `code-agents` subdirectories with JSON output queues and briefs, suggesting active agent loop work not yet committed.
- Several new `plans/*.md` files appear as untracked (e.g. `plans/ci-go-bootstrap-fix-execplan.md`, `plans/docs-stack-command-sync-execplan.md`); these are already associated with done/completed Linear tickets.
- `docs/mintlify/` is fully untracked but the workflow-truth plan (ANM-125) explicitly created these files as deliverables; they should be committed.
- `services/omnichannel/backend-api/` and `services/omnichannel/backend-worker/` are untracked and are the split-service scaffolds referenced by the DevEx plan (ANM-87 area).
- `scripts/ci/` is untracked and was created by ANM-125 (workflow truth audit); it contains `materialize_platform_configs.py` and `verify_docs_config.mjs`.
- Most high-signal buckets were already covered by existing plans or tickets; the true coverage gaps were narrower than the raw dirty tree suggested and resolved to three explicit follow-up tickets: `ANM-164` (access-api retirement decision), `ANM-165` (nested `cli/` repo handling), and `ANM-166` (agent output artifact policy).

---

## Decision Log

- Decision: classify all paths audit-only; do not modify, delete, or revert any file.
  Rationale: ANM-131 and AGENTS.md both prohibit reverting unrelated changes without explicit instruction.
  Date/Author: 2026-03-17 / Claude

- Decision: use git status snapshot from the session context plus plan-file evidence as the source of truth, since Bash is not available in this agent context.
  Rationale: all relevant information is available through plan files and the session's git status snapshot.
  Date/Author: 2026-03-17 / Claude

- Decision: create one follow-up ticket per uncovered bucket instead of expanding ANM-131 to own implementation cleanup for unrelated areas.
  Rationale: the audit's job is to classify and reconcile the dirty tree, not to silently absorb access-api retirement, nested-repo policy, or agent-output ownership decisions into one broad umbrella ticket.
  Date/Author: 2026-03-18 / Codex

---

## Outcomes & Retrospective

Completed outcomes:

- Full bucket table produced below covering all 53+ tracked changes and all major untracked directories.
- AGENTS.md squash-merge policy confirmed already in place.
- Bucket-to-ticket reconciliation completed; uncovered areas now have explicit follow-up tickets: `ANM-164`, `ANM-165`, and `ANM-166`.

Remaining gaps:

- Bucket decisions (commit / branch / clean) are recommendations only; owner must approve before any git operation.
- ANM-131 still needs a final decision on whether to split a clean branch/PR from the stacked local history or close as a documentation-only reconciliation outcome tied to existing commits.
- `cli/` nested `.git` may need a `.gitignore` entry or `git submodule` treatment; tracked in `ANM-165`.
- `agents/` output queue files may contain ephemeral agent artifacts that should not be committed; tracked in `ANM-166`.

---

## Bucket Groupings Table

> Legend — **Status** column:
> - `ready-to-commit` — work is done and files are correct; commit is low risk
> - `needs-review` — files exist but require human judgment before committing
> - `should-ignore` — local-only artifacts that should be added to `.gitignore` and not committed
> - `in-progress` — related plan milestone is still open
> - `stale/orphaned` — no clear owning plan; needs owner decision

---

### Bucket 1 — CI / GitHub Actions Fixes

**Associated ticket:** ANM-133
**Associated plan:** `plans/ci-go-bootstrap-fix-execplan.md`
**Status:** `ready-to-commit`
**Notes:** ANM-133 is marked done. Both workflow files were patched to replace the dead `go-version-file` path with `go-version: stable`. Also includes the new publish-connectrpc-sdks workflow added by the DevEx plan.

| File | Change type |
|------|-------------|
| `.github/workflows/ci.yml` | Modified — Go version bootstrap fix (ANM-133) |
| `.github/workflows/deploy-preflight.yml` | Modified — Go version bootstrap fix (ANM-133) |
| `.github/workflows/publish-connectrpc-sdks.yml` | Untracked — new workflow for ConnectRPC SDK publish (DevEx / ANM ticket 1 area) |

---

### Bucket 2 — Cloud Console UI / Access-API Refactor

**Associated ticket:** Likely ANM-84 area (settings parity) or access-api retirement tracking
**Associated plan:** `plans/devex-build-agentic-execplan.md` (ticket 9 pending), `plans/domains-fallback-oauth-service-clients-execplan.md` (ANM-82)
**Status:** `needs-review`
**Notes:** Three files deleted (access-api route files and lib helpers) and three UI files modified. Deletions appear intentional as part of removing the mock API-key page and migrating to a real backend (devex ticket 9 is still Pending). Confirm that the deleted routes are no longer referenced before committing.

| File | Change type |
|------|-------------|
| `apps/cloud-console/app/_components/app-nav.tsx` | Modified |
| `apps/cloud-console/app/deployments/page.tsx` | Modified |
| `apps/cloud-console/app/settings/page.tsx` | Modified |
| `apps/cloud-console/app/api/keys/[id]/revoke/route.ts` | Deleted |
| `apps/cloud-console/app/api/keys/route.ts` | Deleted |
| `apps/cloud-console/lib/access-api.ts` | Deleted |
| `apps/cloud-console/lib/access-keys.ts` | Deleted |
| `apps/cloud-console/package.json` | Modified — likely SDK/dependency wiring |
| `apps/cloud-console/project.json` | Modified — Nx project config |
| `apps/cloud-console/.env.example` | Untracked — environment variable template |
| `apps/cloud-console/next-env.d.ts` | Untracked — Next.js generated type declaration (consider `.gitignore`) |
| `apps/cloud-console/stack.json` | Untracked — stack service definition (Bucket 5 overlap) |

---

### Bucket 3 — ConnectRPC SDK / Nx / Proto Wiring

**Associated ticket:** DevEx plan ticket 1 (ConnectRPC + Nx), ANM-87 (sdk:generate-es blocker)
**Associated plan:** `plans/devex-build-agentic-execplan.md`
**Status:** `ready-to-commit`
**Notes:** The DevEx plan marks these items as done (2026-03-09 and 2026-03-10). All files are deliverables of completed plan milestones.

| File | Change type |
|------|-------------|
| `nx.json` | Modified — Nx task graph additions |
| `services/omnichannel/project.json` | Modified — Nx SDK target wiring |
| `services/omnichannel/scripts/sdk.sh` | Modified |
| `services/omnichannel/backend/proto/README.md` | Modified |
| `services/omnichannel/backend/proto/buf.gen.server.yaml` | Untracked — buf server generation template |
| `services/omnichannel/backend/proto/buf.yaml` | Untracked — buf module config |
| `services/omnichannel/backend/proto/temporal/` | Untracked — proto sources for temporal workflow |
| `services/omnichannel/backend/internal/rpc/` | Untracked — generated or hand-written RPC stubs |
| `services/omnichannel/backend-api/` | Untracked — split backend-api service scaffold |
| `services/omnichannel/backend-worker/` | Untracked — split backend-worker service scaffold |
| `scripts/sdk-core.sh` | Untracked — Buf/BSR core SDK operations script |
| `scripts/sdk.sh` | Untracked — root Nx wrapper entrypoint for SDK commands |

---

### Bucket 4 — Mintlify Docs Setup

**Associated ticket:** ANM-98 (Mintlify docs dev setup), ANM-104 (CLI version pin), ANM-125 (workflow truth audit)
**Associated plan:** `plans/mintlify-docs-dev-setup-execplan.md`, `plans/mintlify-cli-version-pin-execplan.md`, `plans/workflow-truth-audit-and-cleanup-execplan.md`
**Status:** `ready-to-commit`
**Notes:** All three tickets are marked Done. Files are deliverables of completed milestones.

| File | Change type |
|------|-------------|
| `docs/mintlify/` (whole directory) | Untracked — new Mintlify workspace including `docs.json`, `index.mdx`, `stack.json`, `project.json`, `workflow-status.mdx`, `workflows-public.mdx`, `workflows-internal.mdx` |

---

### Bucket 5 — Platform CLI / Control-Plane / Dev-Stack Scaffolding

**Associated ticket:** ANM-126 (platform-cli refocus), ANM-82 (domains fallback), control-plane domain schema commits (`6003d8f`, `d089e8f`)
**Associated plan:** `plans/platform-cli-scaffolding-console-workflow-execplan.md`, `plans/domains-fallback-oauth-service-clients-execplan.md`
**Status:** `needs-review`
**Notes:** `platform` (the compiled binary wrapper) and `platform-config/` are untracked and may be local-only build outputs. `project.json` (root) is a new Nx project for stack targets using `dev-stack`. `scripts/dev-stack`, `scripts/dev-stack-discover.py`, and `scripts/setup.sh` are new untracked scripts. `infra/platform/declarative_spec.py` and `platform-cli/main.go` are modified source. The binary `platform` itself should likely be `.gitignore`d rather than committed.

| File | Change type |
|------|-------------|
| `platform-cli/main.go` | Modified — platform CLI Go source |
| `infra/platform/declarative_spec.py` | Modified — declarative spec for infra |
| `platform.controlplane.json` | Modified — control-plane config |
| `platform.controlplane.example.json` | Modified — example/template config |
| `platform` | Untracked — compiled binary (should be `.gitignore`d) |
| `platform-config/` | Untracked — platform configuration directory; `project.json` inside |
| `project.json` (repo root) | Untracked — root Nx stack project (postgres/supabase/temporal targets) |
| `scripts/dev-stack` | Untracked — dev stack supervisor script |
| `scripts/dev-stack-discover.py` | Untracked — service discovery helper |
| `scripts/setup.sh` | Untracked — environment setup script |
| `services/omnichannel/stack.json` | Untracked — stack service definition |
| `apps/cloud-console/stack.json` | Untracked — stack service definition (see Bucket 2) |
| `platform.projects.json` | Untracked — platform projects config |
| `platform.projects.example.json` | Untracked — example platform projects config |
| `infra/platform/deployments.generated.json` | Untracked — generated deployments manifest (should be `.gitignore`d or committed as generated artifact) |

---

### Bucket 6 — Access-API Retirement / Hygiene

**Associated ticket:** `ANM-164`
**Associated plan:** `plans/repo-hygiene-bloat-execplan.md`, `plans/devex-build-agentic-execplan.md`
**Status:** `needs-review`
**Notes:** `services/access-api/.gitignore` and `services/access-api/cmd/api/main_test.go` were deleted, and `schemas/access-api/openapi.yaml` was also deleted. The hygiene plan explicitly added `services/access-api/.gitignore` at commit `92c8659`; its deletion here conflicts with that intent. Follow-up ticket `ANM-164` now tracks the retirement-versus-restoration decision before this bucket can be committed safely.

| File | Change type |
|------|-------------|
| `services/access-api/.gitignore` | Deleted — previously added by `92c8659` |
| `services/access-api/cmd/api/main_test.go` | Deleted — tests removed |
| `schemas/access-api/openapi.yaml` | Deleted — OpenAPI schema removed |

---

### Bucket 7 — Repository-Wide Docs, Policy, and Workflow Cleanup

**Associated tickets:** ANM-97 (stack command sync), ANM-125 (workflow truth audit), ANM-126 (platform-cli focus), ANM-128 (remove retired verify docs), ANM-129 (PR triage policy)
**Associated plans:** `plans/docs-stack-command-sync-execplan.md`, `plans/workflow-truth-audit-and-cleanup-execplan.md`, `plans/platform-cli-scaffolding-console-workflow-execplan.md`, `plans/remove-verify-docs-command-references-execplan.md`, `plans/open-pr-test-failure-triage-execplan.md`
**Status:** `ready-to-commit`
**Notes:** All associated tickets are marked Done. These are documentation, policy, and script surface changes accumulated across multiple completed milestones.

| File | Change type |
|------|-------------|
| `AGENTS.md` | Modified — multiple policy updates (PR linkage, squash-merge, platform-cli focus, GitHub MCP preference) |
| `README.md` | Modified — doc sync updates |
| `Makefile` | Modified — clean/workspace targets added |
| `.gitignore` | Modified — ignore rules for generated binaries |
| `.nvmrc` | Modified — Node version pin |
| `docs/agent-mistakes.md` | Modified — mistake log entries |
| `docs/backlog/ci-console-ux-tickets.json` | Modified — backlog payload |
| `docs/backlog/devex-completed-tickets.json` | Modified — backlog payload |
| `docs/backlog/folder-by-feature-tickets.json` | Modified — backlog payload |
| `docs/backlog/local-github-workflow-tickets.json` | Modified — backlog payload |
| `docs/backlog/session-tickets.json` | Modified — backlog payload |
| `docs/backlog/verify-nx-targets-tickets.json` | Modified — backlog payload |
| `docs/repository-architecture-deep-dive.md` | Modified — doc sync |
| `docs/runbooks/build-and-release-gates.md` | Modified — runbook updates |
| `docs/runbooks/linear-backlog-completion.md` | Modified — runbook updates |
| `docs/runbooks/linear-pr-linkage.md` | Modified — new runbook |
| `docs/top-three-priorities.md` | Modified — priority sync |
| `scripts/README.md` | Modified — scripts doc sync |
| `scripts/clean` | Modified — `--full` mode added |
| `scripts/deploy-preflight` | Modified — cleanup/updates |
| `scripts/verify` | Modified — docs mode removed, Nx targets added |
| `package.json` | Modified — npm workspace/scripts updates |
| `package-lock.json` | Modified — dependency lockfile |

---

### Bucket 8 — New Untracked Plans (Completed Milestones)

**Associated tickets:** See per-plan ticket linkage in each file
**Status:** `ready-to-commit`
**Notes:** These plans were created as deliverables of now-Done tickets. They are not in git history yet because they were never committed. All should be committed alongside their associated bucket changes.

| File | Associated ticket |
|------|-------------------|
| `plans/ci-go-bootstrap-fix-execplan.md` | ANM-133 (Done) |
| `plans/docs-stack-command-sync-execplan.md` | ANM-97 (Done) |
| `plans/domains-fallback-oauth-service-clients-execplan.md` | ANM-82 (Done) |
| `plans/mintlify-cli-version-pin-execplan.md` | ANM-104 (Done) |
| `plans/mintlify-docs-dev-setup-execplan.md` | ANM-98 (Done) |
| `plans/open-pr-test-failure-triage-execplan.md` | ANM-129 (Done) |
| `plans/platform-cli-scaffolding-console-workflow-execplan.md` | ANM-126 (Done) |
| `plans/remove-verify-docs-command-references-execplan.md` | ANM-128 (Done) |
| `plans/workflow-truth-audit-and-cleanup-execplan.md` | ANM-125 (Done) |

---

### Bucket 9 — New Untracked Plans (In Progress or Pending)

**Associated tickets:** See per-plan ticket linkage
**Status:** `in-progress`
**Notes:** These plans have open milestones. They should be committed when their milestones land.

| File | Associated ticket | Plan status |
|------|-------------------|-------------|
| `plans/worktree-audit-execplan.md` (this file) | ANM-131 | In Progress |
| `plans/devex-build-agentic-execplan.md` | DevEx plan (ticket 9 pending) | Partially done |

---

### Bucket 10 — Agent Output Artifacts

**Associated ticket:** `ANM-166`
**Status:** `needs-review`
**Notes:** `agents/` contains `pm-agent/out/` and `code-agents/out/` with JSON queue files and agent briefs. These look like runtime outputs of an autonomous agent loop, not source files. Follow-up ticket `ANM-166` now tracks the policy decision on whether to commit them as snapshots or ignore them as ephemeral outputs.

| File | Change type |
|------|-------------|
| `agents/` (whole directory) | Untracked — agent output queue and briefs |

---

### Bucket 11 — New Backlog Payload Files

**Associated tickets:** Related to Linear backlog management
**Status:** `needs-review`
**Notes:** Two new JSON backlog files are untracked. These appear to be ticket payload definitions for new backlog areas.

| File | Change type |
|------|-------------|
| `docs/backlog/agent-dx-improvements-tickets.json` | Untracked — agent DX improvements backlog |
| `docs/backlog/toolchain-devex-tickets.json` | Untracked — toolchain devex backlog |

---

### Bucket 12 — SKILLS.md and Cursor Skills

**Associated ticket:** `ANM-142`
**Status:** `needs-review`
**Notes:** `SKILLS.md` describes optional CLI skill tooling. This overlaps with `ANM-142`, which already tracks missing repo-local skills and Cursor config promised by `AGENTS.md`.

| File | Change type |
|------|-------------|
| `SKILLS.md` | Untracked — optional skills documentation |

---

### Bucket 13 — CLI (Rust Nested Repo)

**Associated ticket:** `ANM-165`
**Status:** `should-ignore` (or treat as git submodule)
**Notes:** `cli/` has its own `.git` repository and a `target/` Rust build cache. The outer git index sees it as a single untracked path. Follow-up ticket `ANM-165` now tracks the decision between ignoring it, promoting it to a documented submodule/subtree, or relocating it. The compiled build artifacts in `cli/target/` must not be committed regardless of approach.

| File | Change type |
|------|-------------|
| `cli/` | Untracked — nested Rust repo with its own `.git` and `target/` cache |

---

### Bucket 14 — Scripts CI (workflow truth audit deliverable)

**Associated ticket:** ANM-125 (Done)
**Status:** `ready-to-commit`
**Notes:** Created by the workflow truth audit plan as verification helpers.

| File | Change type |
|------|-------------|
| `scripts/ci/materialize_platform_configs.py` | Untracked — CI config materialization helper |
| `scripts/ci/verify_docs_config.mjs` | Untracked — deterministic docs config verification |

---

## Recommended Cleanup Sequence

1. **Commit Bucket 7 + Bucket 8 first** — these are all Done-ticket doc/policy/script changes with no runtime risk.
2. **Commit Bucket 1** — CI fixes are Done and unblocking open PRs.
3. **Commit Bucket 3 + Bucket 14** — ConnectRPC/SDK and scripts/ci deliverables from Done milestones.
4. **Commit Bucket 4** — Mintlify docs from Done milestones.
5. **Decide Bucket 2 (cloud-console)** — confirm access-api route deletions are intentional before committing.
6. **Decide Bucket 6 (access-api retirement)** — confirm `.gitignore` and test deletions are intentional retirement.
7. **Decide Bucket 5 (platform/control-plane)** — add `platform` binary and `infra/platform/deployments.generated.json` to `.gitignore` first; then commit config/source changes.
8. **Decide Bucket 10 (agents output)** — determine if ephemeral or intentional snapshots.
9. **Decide Bucket 13 (cli/)** — choose submodule vs `.gitignore` treatment; do not commit `target/` artifacts.
10. **Commit Buckets 9, 11, 12** — commit new plans, backlog payloads, and SKILLS.md when ready.

---

## Validation And Acceptance

Acceptance criteria (per ANM-131):

- [x] Dirty tree grouped into actionable buckets with associated tickets where possible.
- [x] No unrelated changes reverted.
- [x] `AGENTS.md` squash-merge policy confirmed already present (line 216).
- [x] Plan file created at `plans/worktree-audit-execplan.md`.

---

## Context And Orientation

Sources used for this audit:

- Git status snapshot provided in session context (branch `andyminhtuanho/anm-96-restore-missing-mcp-observability-workspace-and-re-enable`)
- All plan files in `plans/*.md` read to map files to owning tickets
- `AGENTS.md` read to confirm policy state
- `cli/`, `agents/`, `platform-config/`, `scripts/ci/`, `docs/mintlify/` directory contents sampled via Glob
- Prior dirty-worktree audit notes in `plans/dirty-worktree-audit-pr-merge-policy-execplan.md` (ANM-131 predecessor work from 2026-03-10)

## Interfaces And Dependencies

- Depends on owner review of `needs-review` buckets before any commit or cleanup operation.
- `cli/` bucket depends on a `.gitignore` or submodule decision.
- Access-api retirement (Bucket 6) depends on confirming whether the service is being decommissioned.
