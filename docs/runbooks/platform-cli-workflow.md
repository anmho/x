# Platform CLI Workflow Runbook

This runbook is the canonical quick guide for platform-cli scaffolding and console workflows in Project X.

## Default Focus

Use platform-cli to scaffold repository components and run service/resource workflows:

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

This runbook intentionally avoids local stack lifecycle playbooks as a default developer workflow.

## Validation Rules

1. `npm run build` is compile validation only.
2. Prefer `scripts/verify` (targeted `docs`, `apps`, `platform`, `agents`) when verifying docs, frontends, or agent tooling; `scripts/deploy-preflight` remains the go/no-go step before deploying.
3. When summaries require evidence, cite the exact scaffolding/workflow or validation command executed (`platform stack`, `scripts/verify`, `scripts/deploy-preflight`, etc.).
