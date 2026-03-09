# Build and Release Gates

This runbook defines the minimum gates before merging or deploying changes.

## Local Merge Gate

From repo root:

```bash
nvm install
nvm use
npm install
npm run doctor
npm run verify
```

See `docs/runbooks/node-dependencies.md` for the Node dependency strategy.

Expected outcome: all checks pass for platform services, app builds, and docs tooling.

## Target Deploy Preflight

Run one of:

```bash
npm run preflight
npm run preflight:cloud-console
npm run preflight:omnichannel
npm run preflight:access-api
```

These checks validate each deploy target can build/test with current code.

## CI Gate

The repository enforces:

- `.github/workflows/ci.yml`
- `.github/workflows/deploy-preflight.yml`

`CI` should pass before merge. `Deploy Preflight` should be executed before production release windows.

## Current Known Risks

- Mintlify and some web dependencies require Node `>=20.19.0`; older local runtimes will fail docs checks.
- Tests are still sparse in several services; passing checks currently prove build integrity more than behavior coverage.
