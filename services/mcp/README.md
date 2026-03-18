# MCP service

`services/mcp` exposes the repository's internal observability tools over ConnectRPC and MCP JSON-RPC.

## Local development

From repo root:

```bash
npx nx run mcp:generate-proto
npx nx run mcp:test
npx nx run mcp:build
npx nx run mcp:build-cli
./scripts/deploy-preflight mcp
```

Run the server:

```bash
cd services/mcp
GOCACHE=/tmp/go-cache go run ./cmd/server
```

On first startup the server generates a key in `~/.x-mcp/keys.json` and prints it once. Use that key with the standalone CLI:

```bash
./bin/mcp --server http://localhost:8765 --key <api-key> tools list
./bin/mcp --server http://localhost:8765 --key <api-key> keys list
```

The repo-level wrapper delegates to the same binary directly:

```bash
./platform mcp tools list --server http://localhost:8765 --key <api-key>
```

## Docker

Build from the repo root:

```bash
docker build -f services/mcp/Dockerfile -t x-mcp .
```

Run locally with the key store persisted:

```bash
docker run --rm -p 8765:8765 \
  -v "$HOME/.x-mcp:/root/.x-mcp" \
  x-mcp
```

The image now includes `platform.controlplane.json` and sets `MCP_ROOT=/app`, so config-backed tools work without an extra bind mount. Tools that shell out to `gcloud` or `vercel` still require those CLIs and their auth to be available in the runtime environment.

## Deployment model

This repo already uses a declarative platform workflow rather than ArgoCD:

1. Edit `infra/platform/declarative_spec.py`.
2. Materialize generated configs with `python3 scripts/ci/materialize_platform_configs.py`.
3. Review with `./platform control-plane plan --project mcp`.
4. Dry-run deployment with `./platform deploy --project mcp --dry-run`.

That keeps `mcp` aligned with the same generated `platform.projects.json` and `platform.controlplane.json` flow used by the other services.
