# Workflow Status Matrix

Last audited: **March 10, 2026** (`America/Los_Angeles`).

All commands below were executed from repo root: `/Users/andrewho/repos/projects/x`.

## Public Workflows

| Workflow | Command | Status | Notes |
| --- | --- | --- | --- |
| Environment setup | `nvm install && nvm use && npm install` | Working | Required for Node `24.14.0` workflow baseline. |
| Root lint | `npm run lint` | Broken | Active frontend lint debt remains in `cloud-console` and `omnichannel-frontend`. |
| Root build | `npm run build` | Working/Partial | Root build entrypoint now exists; success depends on active project build health. |
| Root test | `npm run test` | Broken | Root test entrypoint now exists; failure currently comes from real project lint/test debt rather than missing scripts. |
| App targets | `npx nx run <app>:lint|build|test` | Working/Partial | Nx command contract is active; project results vary by current app health. |
| Docs test | `npm run test:docs` (Node `24.14.0`) | Working | Validates docs config/page references (`docs/mintlify/docs.json` and page files). |
| Docs local demo | `npm run docs:dev` | Working (manual) | Starts Mintlify dev mode; interactive/long-running. |
| SDK lint | `./scripts/sdk.sh lint` | Working | Buf lint and breaking checks run. |
| SDK local generation | `./scripts/sdk.sh generate-es` / `generate-go` | Working | Local proto-to-SDK generation succeeds. |
| Direct proto generation | `buf generate --template ...` | Working | `buf.gen.client.yaml` and `buf.gen.server.yaml` run locally. |

## Internal Workflows

| Workflow | Command | Status | Notes |
| --- | --- | --- | --- |
| Platform test | `npm run test:platform` (Node `24.14.0`) | Broken | Omnichannel Go checks fail due missing module/dependencies and undefined symbols. |
| Agent test | `npm run test:agents` | Broken | `agents:test` still shells into the missing `agents/blueprints/verify` path. |
| Deploy preflight (platform) | `./scripts/deploy-preflight platform` (Node `24.14.0`) | Broken | Fails when root test/platform targets fail. |
| Deploy preflight (cloud-console) | `./scripts/deploy-preflight cloud-console` (Node `24.14.0`) | Broken | Cloud console lint now runs and surfaces real frontend issues. |
| Deploy preflight (omnichannel) | `./scripts/deploy-preflight omnichannel` | Broken | Omnichannel Go checks fail (missing module/dependency wiring and undefined symbols). |
| Stack status | `./platform status` | Working | Service discovery/status output is returned. |
| Stack start | `./platform start` | Blocked | Postgres bind conflict on `54322` (`supabase_db_omnichannel`). |
| Deploy dry-run | `./platform deploy --project cloud-console --dry-run` | Blocked | Requires provider secrets (for example `CLOUDFLARE_API_TOKEN`). |
| SDK publish | `./scripts/sdk.sh push` / `publish-all` | Blocked | Requires valid Buf token (`buf registry login`). |

## Immediate Fixes Applied In This Audit

- Added root `lint`, `build`, and `test` entrypoints.
- Moved CI app/platform/docs/agent checks onto Nx `lint`, `build`, and `test` targets instead of `verify`.
- Reduced `scripts/verify` to a compatibility bridge over the root test surfaces.

## Follow-Up Work Needed

- Repair omnichannel Go module/test surface used by `test:platform`.
- Restore/replace the `agents:test` compatibility path away from `agents/blueprints/verify`.
- Stabilize `cloud-console:test` and `omnichannel-frontend:test` by fixing active frontend lint debt.
- Resolve local stack Postgres/Supabase port conflict for reliable `./platform start`.
