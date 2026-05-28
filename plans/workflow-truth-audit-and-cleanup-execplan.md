# Workflow truth audit, documentation sync, and dead-code cleanup

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-125` ([issue link](https://linear.app/anmho/issue/ANM-125/medium-audit-and-update-workflow-docs-remove-dead-workflow-code)).

## Purpose / Big Picture

Ensure repository workflow documentation reflects actual behavior today, identify which core workflows are working vs broken, and remove confirmed dead workflow code/references without breaking core development paths.

## Progress

- [x] (2026-03-10 09:20Z) Created `ANM-125` in Linear with scope, validation plan, and risk context; set status to `In Progress`.
- [x] (2026-03-10 09:21Z) Created this ExecPlan and linked it to `ANM-125`.
- [x] (2026-03-18 05:52Z) Audited workflow command surfaces across `./platform`, `scripts/*`, `package.json`, CI workflows, Mintlify docs config, and SDK/proto paths.
- [x] (2026-03-18 06:12Z) Executed validation sweeps and recorded pass/fail for test/build/deploy/publish/docs/stack/proto+SDK workflows.
- [x] (2026-03-18 06:21Z) Updated repo docs and Mintlify docs with explicit working/broken/blocked status and public/internal workflow separation.
- [x] (2026-03-18 06:23Z) Removed dead workflow references and stale commands (`preflight:access-api`, deprecated Mintlify `build` checks, stale docs verify paths).
- [x] (2026-03-18 06:26Z) Ran post-change rescan, reconciled existing ticket coverage (`ANM-139`, `ANM-146`), and created distinct follow-up tickets `ANM-153` and `ANM-154`.
- [x] (2026-03-18 06:28Z) Moved `ANM-125` to `Done` and aligned plan outcomes with final ticket state.

## Surprises & Discoveries

- Observation: Root scripts `npm run test` and `npm run build` are absent, so these canonical workflow names are currently broken at root.
  Evidence: `npm run test` and `npm run build` each return "Missing script".

- Observation: Platform CLI behavior currently depends on a prebuilt `bin/platform` binary while `platform-cli` source does not compile.
  Evidence: `cd platform-cli && go test ./...` fails with unresolved symbols (`runStack`, `findRepoRoot`, etc.), while `./platform --help` still runs.

- Observation: Stack startup failure is caused by a concrete local port conflict, not only missing tooling.
  Evidence: `./platform start` fails on Postgres; `docker ps` shows `supabase_db_omnichannel` already bound to host port `54322`.

- Observation: SDK local generation flows work, but publish flows are credential-gated.
  Evidence: `./scripts/sdk.sh generate-es` and `generate-go` succeed; `./scripts/sdk.sh push`/`publish-all` fail with invalid Buf token.

## Decision Log

- Decision: Create a new ExecPlan instead of extending prior docs plans.
  Rationale: This request spans docs truth-auditing, runtime validation, and dead-code cleanup beyond earlier narrow docs sync tasks.
  Date/Author: 2026-03-10 / Codex

- Decision: Replace dead Mintlify `build` verification commands with deterministic local docs-config validation.
  Rationale: `mint build` is no longer available in the pinned CLI path and `mint validate` was non-deterministic in this environment; config/page checks are stable and CI-safe.
  Date/Author: 2026-03-18 / Codex

- Decision: Keep root `test`/`build` as documented broken flows (for now) rather than introducing speculative aliases.
  Rationale: Existing verification/preflight entrypoints are already used across docs/CI; creating aliases without broader agreement risks masking unresolved platform issues.
  Date/Author: 2026-03-18 / Codex

## Outcomes & Retrospective

Completed outcomes:

- Added a central audited status matrix at `docs/runbooks/workflow-status-matrix.md`.
- Added Mintlify workflow documentation with internal/public split:
  - `docs/mintlify/workflow-status.mdx`
  - `docs/mintlify/workflows-public.mdx`
  - `docs/mintlify/workflows-internal.mdx`
- Updated Mintlify config/docs navigation and made docs checks deterministic:
  - `docs/mintlify/docs.json`
  - `scripts/ci/verify_docs_config.mjs`
  - `scripts/verify` docs mode
  - `docs/mintlify/project.json` docs targets
- Cleaned dead/stale workflow references:
  - Removed `preflight:access-api` from `package.json`.
  - Updated runbooks/README/scripts docs to match current command behavior.
- Captured unresolved risks with follow-up issues:
  - Existing coverage: `ANM-139`, `ANM-146`
  - New: `ANM-153` (platform-cli source/binary drift), `ANM-154` (stack Postgres port conflict)

Remaining gaps:

- `npm run verify` and `npm run preflight` remain red due unresolved platform/omnichannel Go issues.
- `verify:agents` remains broken due missing `agents/blueprints/verify`.
- SDK publish and deploy dry-runs remain credential-gated (Buf/provider secrets).

## Context And Orientation

Likely high-impact files:

- `README.md`
- `AGENTS.md`
- `package.json`
- `platform`
- `scripts/README.md`
- `scripts/dev-stack`
- `scripts/verify`
- `docs/runbooks/platform-cli-workflow.md`
- `docs/runbooks/build-and-release-gates.md`
- `docs/mintlify/docs.json`
- `docs/mintlify/index.mdx`

## Plan Of Work

1. Inspect current command implementations and script targets to establish ground truth.
2. Run workflow validations and classify each workflow as working, blocked by environment/credentials, or broken.
3. Update docs (repo + Mintlify) to reflect ground truth and to separate internal/public flows where possible.
4. Remove dead code/references confirmed by the audit.
5. Run post-change rescan, reconcile ticket coverage, and create follow-up tickets as required.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Baseline commands and scripts:
   - `./platform --help`
   - `./platform status`
   - `./platform stack --help`
   - `cat package.json`
   - `ls -la scripts`
2. Validate workflows:
   - `npm run test`
   - `npm run build`
   - `npm run docs:dev` (startup smoke)
   - `./platform status` and `./platform start`/`./platform logs <service>` as needed
   - proto/sdk commands discovered from scripts/docs (including `buf generate` path if present)
3. Audit docs drift with `rg` and patch mismatches.
4. Identify dead workflow code via orphaned scripts/references and remove safe targets.
5. Re-scan with `rg -n "TODO|FIXME|TBD|XXX"` and ticket reconciliation against existing Linear issues.

## Validation And Acceptance

Acceptance criteria:

- Documentation explicitly reports which key workflows currently work, are blocked by prerequisites, or are broken.
- Mintlify docs include current workflow guidance and indicate internal/public applicability where feasible.
- Removed code/references are demonstrably dead (unreferenced or superseded) and do not break documented core flows.
- `ANM-125` status and this plan `Progress`/`Outcomes & Retrospective` are aligned.

Validation commands to run and record:

- `./platform --help`
- `./platform status`
- `./platform stack --help`
- `npm run test`
- `npm run build`
- `npm run docs:dev` (smoke)
- workflow-specific proto/sdk/deploy/publish commands discovered during audit

Validation executed (selected):

- `./platform --help` (pass)
- `./platform status` (pass)
- `./platform start` (fails: Postgres container/port conflict)
- `npm run test` (fails: missing root script)
- `npm run build` (fails: missing root script)
- `/bin/zsh -lc 'source ~/.nvm/nvm.sh && nvm use >/dev/null && npm run verify:docs'` (pass)
- `/bin/zsh -lc 'source ~/.nvm/nvm.sh && nvm use >/dev/null && npm run verify:apps'` (pass)
- `/bin/zsh -lc 'source ~/.nvm/nvm.sh && nvm use >/dev/null && npm run verify:platform'` (fails: omnichannel Go module/dependency/symbol issues)
- `npm run verify:agents` (fails: missing `agents/blueprints/verify`)
- `/bin/zsh -lc 'source ~/.nvm/nvm.sh && nvm use >/dev/null && npm run preflight:cloud-console'` (pass)
- `/bin/zsh -lc 'source ~/.nvm/nvm.sh && nvm use >/dev/null && npm run preflight:omnichannel'` (fails: omnichannel Go module/dependency/symbol issues)
- `NX_DAEMON=false ./scripts/sdk.sh generate-es` (pass)
- `NX_DAEMON=false ./scripts/sdk.sh generate-go` (pass)
- `cd services/omnichannel/backend/proto && ../../../../node_modules/.bin/buf generate --template buf.gen.client.yaml` (pass)
- `NX_DAEMON=false ./scripts/sdk.sh push` with network access (fails: invalid Buf API token)

## Idempotence And Recovery

All doc updates are repeatable by rerunning the same workflow checks and updating status tables. Dead-code cleanup should be limited to verified orphaned paths so restoration is possible via git revert of specific files if needed.

## Artifacts And Notes

- Linear issue: `ANM-125`
- Plan file: `plans/workflow-truth-audit-and-cleanup-execplan.md`
- Canonical project tracking URL: `https://linear.app/anmho/team/ANM/projects/all`
- Follow-up ticket reconciliation:
  - Existing: `ANM-139`, `ANM-146`
  - New: `ANM-153`, `ANM-154`

## Interfaces And Dependencies

- Platform CLI wrappers (`platform`, `platform-cli`)
- Stack supervisor helpers (`scripts/dev-stack`)
- Workspace scripts (`package.json`, `scripts/verify`)
- Mintlify docs tooling (`docs/mintlify`)
- Proto/SDK generation and publish scripts (to be discovered during audit)
