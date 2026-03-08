# Platform CLI Workflow Runbook

This runbook is the canonical quick guide for local stack operations in Project X.

## Default Entry Point

From repository root (`/Users/andrewho/repos/projects/x`), use:

```bash
./platform start
./platform status
./platform logs <service>
./platform stop
```

Use this path by default for validation and user-facing status reporting.

## When To Use `scripts/dev-stack`

Use `scripts/dev-stack` when you explicitly need its supervisor semantics:

```bash
scripts/dev-stack start
scripts/dev-stack status
scripts/dev-stack logs cloud-console
scripts/dev-stack stop
```

## When To Use Service-Local Stack Scripts

Use service-local scripts only for scoped service work:

- Omnichannel stack: `services/omnichannel/scripts/stack.sh`

Example:

```bash
cd services/omnichannel
./scripts/stack.sh up
./scripts/stack.sh status
./scripts/stack.sh logs api
```

## Validation Rules

1. `npm run build` is compile validation only.
2. Stack/runtime validation must come from `./platform ...`, `scripts/dev-stack ...`, or an explicitly scoped service stack script.
3. In summaries, always cite the exact command used for runtime checks.
