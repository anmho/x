# Build and Release Gates

This runbook defines the minimum gates before merging or deploying changes.

## Local Merge Gate

From repo root:

```bash
nvm install
nvm use
npm install
npm run lint
npm run build
npm run test
```

See `docs/runbooks/node-dependencies.md` for the Node dependency strategy.

Expected outcome:

- `lint`, `build`, and `test` should pass in a healthy local setup.
- If `test:platform` fails, use the status matrix to identify the current platform blockers.

## Target Deploy Preflight

Run one of:

```bash
./scripts/deploy-preflight platform
./scripts/deploy-preflight cloud-console
./scripts/deploy-preflight omnichannel
```

These checks validate each deploy target can build/test with current code.

Observed state (March 10, 2026):

- `cloud-console`: currently blocked by active frontend lint errors
- `platform`: broken when platform test targets fail
- `omnichannel`: broken when omnichannel Go checks fail

## CI Gate

The repository enforces:

- `.github/workflows/ci.yml`
- `.github/workflows/deploy-preflight.yml`
- `.github/workflows/publish-connectrpc-sdks.yml` (manual publish gate for BSR SDK versions)

`CI` should pass before merge. `Deploy Preflight` should be executed before production release windows.

## Current Known Risks / Blockers

- `test:platform` fails in omnichannel Go checks due missing module/dependency wiring and undefined symbols.
- `test:agents` depends on `agents/blueprints/verify`, which is still a compatibility path.
- `scripts/deploy-preflight omnichannel` fails in omnichannel Go checks (module/dependency/symbol issues).
- SDK publish commands require a valid `BUF_TOKEN` (`buf registry login`).
- Deploy dry-run requires provider credentials (for example `CLOUDFLARE_API_TOKEN`).

For the complete audited command list, use [workflow-status-matrix.md](./workflow-status-matrix.md).
