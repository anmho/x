# Platform CLI Workflow Runbook

This runbook is the canonical quick guide for platform-cli workflows in Project X.

## Scaffolding, Project, and Deploy Commands

Use `./platform` for scaffolding and operational workflows:

```bash
./platform create service billing-api
./platform create integration add vercel --project cloud-console
./platform config init
./platform project list
./platform project keys mint --project cloud-console --owner andrew --env dev
./platform tokens list --project cloud-console
./platform notifications list
./platform control-plane plan --project cloud-console
./platform deploy --project cloud-console --dry-run
./platform docs dev --port 3002
```

## Repo Validation

Use root commands for repo validation:

```bash
npm run lint
npm run build
npm run test
```

For targeted work, prefer Nx directly:

```bash
npx nx run <project>:lint
npx nx run <project>:build
npx nx run <project>:test
```

## Runtime and Stack (Troubleshooting)

Use `./platform` for stack lifecycle troubleshooting when you need the local stack running:

```bash
./platform start
./platform status
./platform logs <service>
./platform stop
./platform stack temporal-ui
```

Use `scripts/dev-stack` for supervisor-level troubleshooting:

```bash
scripts/dev-stack list
scripts/dev-stack status
scripts/dev-stack logs <service>
```

Current known blocker:

- `./platform start` can fail when local port `54322` is already bound by `supabase_db_omnichannel`.

## Validation Rules

1. Treat `npm run lint`, `npm run build`, and `npm run test` as the primary repo validation gates.
2. Use `./scripts/deploy-preflight <target>` only for deploy-oriented checks.
3. Cite the exact command used for runtime evidence in status updates.
4. Check the audited state table in [workflow-status-matrix.md](./workflow-status-matrix.md) before claiming a flow is green.
