# Scripts Reference

This directory contains low-level operational scripts for local development, validation, and release checks.

Primary repo entrypoints live at the root:

```bash
npm run lint
npm run build
npm run test
make lint
make test
```

Use the scripts in this directory when you specifically need a low-level helper or a deploy-oriented preflight target.

## Core Commands

```bash
# Root validation
npm run lint
npm run build
npm run test
make lint
make test

# Deploy preflight
./scripts/deploy-preflight platform
./scripts/deploy-preflight cloud-console
./scripts/deploy-preflight omnichannel

# Cleanup
scripts/clean
scripts/clean --dry-run
scripts/clean --full
scripts/clean --dry-run --full
```

## Platform CLI Scaffolding and Workflows

Use platform-cli for scaffolding and service/resource workflows:

```bash
platform create service billing-api
platform create integration add vercel --project cloud-console
platform config init
platform project list
platform project keys mint --project cloud-console --owner andrew --env dev
platform tokens list --project cloud-console
platform notifications list
platform control-plane plan --project cloud-console
platform deploy --project cloud-console --dry-run
platform docs
```

See workflow pass/fail status in `docs/runbooks/workflow-status-matrix.md`.

## ConnectRPC SDK Flow (Nx)

Use `scripts/sdk.sh` as the monorepo SDK entrypoint.

```bash
scripts/sdk.sh lint
scripts/sdk.sh generate-es
scripts/sdk.sh generate-server
scripts/sdk.sh push
scripts/sdk.sh publish-all
scripts/sdk.sh list
```

## PR and Linear Helpers

Create PRs for pushed branches:

```bash
./scripts/create-prs-for-branches.sh --dry-run
./scripts/create-prs-for-branches.sh
```

For Linear linkage, use:

```bash
./scripts/linear/link-pr-to-linear.sh ANM-123
```

Issue creation itself should use Linear MCP directly (see `scripts/linear/README.md`).

## Notes

- `scripts/new` is not part of the current repository scripts.
- `scripts/doctor` is a compatibility wrapper around `scripts/setup.sh`.
- `scripts/verify` is a compatibility bridge that delegates to the root `npm run test:*` commands.
