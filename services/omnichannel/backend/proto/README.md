# ConnectRPC SDK Publishing

This proto module is published to Buf Schema Registry (BSR):

- Module: `buf.build/anmhela/omnichannel`

## Policy

- Source of truth is `.proto` files in this folder.
- SDKs are published from BSR and consumed as versioned dependencies.
- Do not manually version checked-in generated client SDKs in this repo.

## Publish flow

From `services/omnichannel`:

```bash
# 1) Validate schema
./scripts/sdk.sh lint

# 2) Push proto schema
./scripts/sdk.sh push

# 3) Publish SDK versions
SDK_VERSION=v1.0.0 ./scripts/sdk.sh publish-all
```

## Server code generation

Backend runtime stubs can still be generated for server builds:

```bash
./scripts/sdk.sh generate-server
# or: nx run sdk:generate-go
```

This updates `backend/internal/rpc/gen` from `backend/proto` using `buf.gen.server.yaml`.

## Client (ES/TypeScript) generation

For local frontend development, generate the TypeScript client:

```bash
nx run sdk:generate-es
# or: ./scripts/sdk.sh generate-es
```

Output: `packages/sdk-omnichannel/src/gen`. Cloud console and omnichannel frontend consume this via workspace dependency `@x/sdk-omnichannel`.

## Discover SDK versions

```bash
./scripts/sdk.sh list
```

Use BSR-listed versions in downstream services/apps rather than generating client SDK code in those repos.
