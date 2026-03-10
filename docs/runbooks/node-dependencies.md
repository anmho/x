# Node Dependencies Runbook

This runbook documents the Node.js dependency strategy for Project X.

## Strategy: Single Root Lockfile

The monorepo uses **npm workspaces** with a **single root lockfile** (`package-lock.json` at repo root).

- All workspaces share one `package-lock.json` for deterministic installs.
- No per-workspace lockfiles; they were removed to avoid ambiguous workspace-root warnings and install drift.

## Workspaces

Defined in root `package.json`:

- `apps/cloud-console`
- `mcp`
- `packages/*`

## Install Instructions

From repository root:

```bash
nvm install
nvm use
npm install
```

For CI or reproducible installs:

```bash
npm ci
```

Never run `npm install` from a workspace directory; always install from root so the workspace tree stays consistent.

## Build and Verify

After installing from root:

```bash
npm run verify
```

Or build a specific app:

```bash
npm run build --workspace=@x/cloud-console
```

## MCP Package Note

The `mcp` workspace uses **Bun** for its `check` script (`bun build`). npm installs its dependencies; Bun is only required when running `npm run mcp:check` or the MCP servers. Ensure Bun is installed if you need to run those commands.
