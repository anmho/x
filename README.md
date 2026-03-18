# Project X

Project X is a super monorepo for building and deploying complete products fast.

It is designed to provide reusable building blocks for:

- Notifications
- Authentication and identity
- Agent workflows
- APIs and backend services
- Web frontends and admin panels
- React Native mobile applications
- CLIs
- Shared contracts/protos
- Shared infrastructure and deployment workflows
- Crypto applications and smart contracts

The goal is simple: one repo, many products, consistent standards.

## Vision

Project X should let teams start new products with strong defaults, not from scratch.

Every new app or service should inherit:

- Common architecture patterns
- Shared auth, observability, and deployment rails
- Stable API and schema contracts
- Reusable domain modules
- Scaffolding for AI-agent and human-operated workflows

## What Lives Here

The repository uses one canonical layout:

- `apps/`: deployable user-facing applications.
- `services/`: deployable backend services and workers.
- `agents/`: agent modules, prompts, and orchestration code.
- `packages/`: shared SDKs and libraries.
- `protos/`: protobuf contracts.
- `schemas/`: OpenAPI/JSON/event schemas.
- `infra/`: deployment and infrastructure assets.
- `scripts/`: low-level automation.

- `docs/`: architecture and runbooks.
- `plans/`: executable implementation plans.

Current notable components:

- `services/omnichannel/`: notifications platform stack (backend API/worker/runtime modules).
- `apps/cloud-console/`: dashboard for service access and operations workflows.
- `packages/sdk-omnichannel/`: shared SDK package for omnichannel integrations.
- `apps/x-stream-bot/`: Rust stream bridge for X signals.

## Cloud Console Dashboard (`cloud.anmhela.com`)

Project X will include a dedicated cloud console dashboard at `cloud.anmhela.com`.

Purpose:

- Single place to provision and manage API keys
- Service catalog where developers pick the services they need
- Integrated docs and examples so onboarding happens inside the product

Core capabilities:

1. API Key Management
- Create API keys scoped to an application, environment, and service
- View metadata (name, owner, scopes, created-at, last-used-at, status)
- Rotate keys with overlap windows to avoid downtime
- Revoke compromised keys immediately
- Audit trail for create/rotate/revoke events

2. Service Access Management
- Catalog of available services (`notifications`, `auth`, `agents`, etc.)
- Per-key, per-service scope assignment (`read`, `write`, `admin`)
- Environment-aware access (`dev`, `staging`, `prod`)
- Policy checks before issuance (ownership, quota, org limits)

3. Embedded Docs (Mintlify)
- Mintlify-powered docs surface embedded in the dashboard
- Service pages include endpoint references, auth requirements, and limits
- Language-specific examples (curl, JavaScript/TypeScript, Go, Python)
- Copy-ready snippets generated from selected service + scope context

4. Developer Experience
- Quick-start flow: create app -> choose service -> mint key -> test request
- “Try it” panels for sample requests against sandbox endpoints
- SDK onboarding links and environment setup checklists
- Error troubleshooting guides tied to common API responses

## Trading Bots (Kalshi + Crypto via XAPI)

The cloud console also supports bot operations for event markets and crypto markets, with an `XAPI` integration path for live execution.

Bot capabilities:

1. Kalshi bots
- Create runs for event-market strategies (breakout, mean reversion, news reaction)
- Configure market contracts, position sizing, and risk caps
- Attach alerting/webhook integrations

2. Crypto bots
- Create runs for spot/derivatives strategies (basis arb, market making, momentum)
- Configure trading pairs, notional sizing, and max risk in basis points
- Attach exchange and notification integrations

3. XAPI integration mode
- `mock` mode by default for local development
- `xapi-live` mode when `XAPI_BASE_URL` and `XAPI_API_KEY` are configured
- Server-side route handlers proxy bot run requests to live XAPI

Proposed implementation modules:

- `apps/cloud-console`: cloud console frontend (service picker, key management, docs shell)
- `services/omnichannel`: backend runtime for notification delivery and orchestration
- `packages/sdk-*`: generated clients and helper libraries
- `docs/mintlify`: source content for API docs and usage guides

Initial UX flow:

1. User creates or selects an application
2. User selects one or more platform services
3. User chooses scopes and environment
4. System mints scoped API key and records audit event
5. Dashboard shows usage example and links to Mintlify docs for that exact configuration

## Platform Building Blocks

Project X standardizes foundational capabilities across all products:

1. Auth and Identity
- User auth, API key auth, service-to-service auth
- Role/permission patterns for apps and APIs

2. Notifications
- Channel abstraction (`email`, `sms`, `push`, etc.)
- Template management, delivery tracking, retries, scheduling

3. Contracts
- Protobuf and OpenAPI-first interfaces
- Versioned APIs and typed client generation

4. Data
- Shared migration strategy and schema ownership rules
- Event and audit trails as first-class data products

5. Agents
- Scaffolding for task-specific agents and tool integrations
- Guardrails, evaluation hooks, and execution logs

6. Frontend Foundation
- Shared design primitives and API client patterns
- Consistent auth/session and error/loading handling
- Dashboard shell pattern for console products (navigation, service catalog, key workflows, embedded docs)

7. Mobile Foundation (React Native)
- React Native app scaffolding with shared architecture defaults
- Typed API clients generated from shared contracts
- Environment config, secrets strategy, and feature flag wiring
- Shared release strategy across iOS and Android

8. Crypto and Smart Contracts
- EVM-compatible smart contract scaffolding and deployment rails
- Contract testing, static analysis, and security audit checkpoints
- Indexer/event ingestion patterns for on-chain data into platform services
- Wallet auth and signing flows for web/mobile apps

9. Deployment Foundation
- Repeatable CI/CD workflows
- Environment promotion (`dev` -> `staging` -> `prod`)
- Health checks, rollbacks, and release metadata

10. Golden Paths
- Opinionated release paths for each surface (web, service, mobile, agents)
- Guardrails for quality, compliance, and predictable publishing
 
11. Developer Console and Docs
- First-class cloud console experience at `cloud.anmhela.com`
- In-product Mintlify docs tightly coupled to real service and key scopes
- Example generation based on selected service, auth mode, and environment

## Monorepo Principles

- Contract-first: define API/proto/schema before implementation
- Reuse-first: prefer shared package before app-local duplication
- Backward-compatibility: additive changes by default
- Operability-first: every service ships with health checks, metrics, logs
- Developer velocity: every component should be scaffoldable in minutes

## Operator Surface

To reduce command sprawl, use these two primary entrypoints:

- `make`: high-level repo tasks (`make setup`, `make test`, `make build`, `make stack-up`).
- `platform`: scaffolding and service/resource workflows (`platform create ...`, `platform deploy --project ...`, `platform tokens ...`, `platform project ...`).

Low-level scripts in `scripts/` still exist, but should be treated as implementation details behind `make` and platform-cli workflow commands.

## Project Registry and MCP Observability

Projects and deployments are declared once in Python and materialized into JSON:

- declarative source: `infra/platform/declarative_spec.py`
- mutable integration overrides: `infra/platform/integrations.overrides.json`
- generated configs:
  - `platform.projects.json`
  - `platform.projects.example.json`
  - `platform.controlplane.json`
  - `platform.controlplane.example.json`
  - `infra/platform/deployments.generated.json`

Definitions include:

- deploy metadata
- deployment declarations
- integration declarations
- dependencies
- Terraform root/workspace
- observability targets (Vercel and/or GCP)
- registered service endpoints
- account-level secret sharing targets for GCP/Vercel (shareable to projects/apps/deployments) in control-plane config

Materialize config JSON with:

- `make materialize-configs`

Validate generated config is current with:

- `python3 scripts/ci/materialize_platform_configs.py --check`

Generate canonical platform artifacts with:

- `make project-registry`

MCP servers:

- `mcp/vercel-observability/server.ts`
- `mcp/gcp-observability/server.ts`

## Top-Level Layout

```text
x/
├── apps/
├── services/
├── agents/
├── packages/
├── protos/
├── schemas/
├── infra/
├── scripts/
├── platform-cli/
├── docs/
├── plans/
└── README.md
```

## Scaffolding Roadmap

Minimum generators to add:

- `new:app` (Next.js or other frontend starter)
- `new:cloud-console` (dashboard starter with auth, key management UI, and docs embedding)
- `new:crypto-app` (web/mobile scaffold with wallet auth and on-chain data hooks)
- `new:mobile-app` (React Native project with shared modules and CI)
- `new:service` (API/worker with health checks, config, logging)
- `new:agent` (agent config + tool interfaces + evaluation harness)
- `new:package` (shared library template with tests and release config)
- `new:proto` (service proto + generation config + client stubs)
- `new:smart-contract` (Foundry/Hardhat template with tests, deployment scripts, and verification)

Each scaffold should include:

- Lint/test config
- CI pipeline defaults
- Environment config template
- Observability hooks
- Deployment manifest
- Mintlify docs stubs when relevant to developer-facing APIs
- Security checks for contract code where relevant (slither/static analysis/audit checklist)

## Deployment Model

Project X should support both independent and coordinated deploys:

- App-level deploys (frontend-only releases)
- Service-level deploys (API/worker/webhook)
- Mobile publishing pipelines (TestFlight + App Store + Google Play)
- Monorepo-wide contract checks before release
- Migration gating for database changes
- Automated rollback strategy per service

Recommended release flow:

1. Validate contracts (proto/OpenAPI/schema compatibility)
2. Run unit/integration tests for affected projects
3. Build deployable artifacts
4. Deploy to `dev`, then `staging`, then `prod`
5. Verify health/SLO gates
6. Promote or rollback

## Golden Paths

Golden Paths are the default, paved workflows for how anything ships in Project X.

Every new app, service, or agent should follow a Golden Path template instead of custom release logic.

### React Native Mobile Golden Path (Publishing)

1. Scaffold app via `new:mobile-app`
2. Generate typed API clients from `protos/` or OpenAPI schemas
3. Run local quality gates (lint, tests, static analysis)
4. Build signed artifacts in CI
5. Publish to TestFlight and Play internal track automatically on release branch
6. Run smoke checks and release checklist for iOS and Android
7. Promote to App Store and Google Play with tracked release metadata and rollback plan

Required mobile pipeline components:

- EAS/Fastlane lane conventions for beta and production
- iOS and Android signing management
- Version/build number automation
- Crash and analytics instrumentation
- Store metadata and screenshot management for both stores
- Post-release monitoring and alerting

### Smart Contract Golden Path

1. Scaffold contract via `new:smart-contract`
2. Define contract interfaces and events with explicit versioning
3. Implement unit/invariant/fuzz tests
4. Run security gates (static analyzers + dependency checks)
5. Deploy to testnet, run smoke and integration checks
6. Verify source and ABI in block explorer tooling
7. Promote to mainnet with signed release approvals and rollback playbook

Required smart contract pipeline components:

- Deterministic builds and pinned compiler/toolchain versions
- Automated ABI artifact publication to `packages/` SDKs
- Contract upgrade governance and timelock/multisig controls where applicable
- Event indexing + monitoring for critical on-chain activity
- Emergency pause/incident procedures for high-risk protocols

## Working Agreements

- New feature work starts with a short design note in `docs/`
- Cross-service changes require explicit contract updates
- Shared package changes include at least one downstream usage update
- Every deployable unit must expose `/health` and structured logs
- Every React Native app must ship through the mobile Golden Path (no manual-only releases)
- Every production smart contract must ship through the Smart Contract Golden Path
- Breaking changes require migration plan and deprecation window

## Near-Term Priorities

1. Define canonical repo structure (`apps/services/packages/protos/infra`)
2. Stand up scaffolding commands in `scripts/` (including `new:mobile-app`, `new:cloud-console`, and `new:smart-contract`)
3. Define Cloud Console MVP at `cloud.anmhela.com` with key issuance + service catalog
4. Add crypto platform primitives (wallet auth, chain config, indexer patterns)
5. Add smart contract security gates and release governance
6. Add shared contract toolchain (`protos/`, `schemas/`, ABI artifacts)
7. Add Mintlify docs pipeline and dashboard embedding strategy
8. Consolidate deployment workflows and Golden Paths for all deployable units
9. Promote existing `omnichannel` components into reusable platform modules
10. Wire `CLI` into real platform APIs for operational workflows

## Build and Deploy Gates

Use these commands from repo root:

```bash
nvm install
nvm use
npm install
npm run verify
npm run verify:docs
npm run preflight
```

Node dependencies use a single root lockfile; see `docs/runbooks/node-dependencies.md`.

Audited working/broken status for every core workflow is tracked in:

- `docs/runbooks/workflow-status-matrix.md`
- `docs/mintlify/workflow-status.mdx`

For target-specific checks:

- `npm run preflight:cloud-console`
- `npm run preflight:omnichannel`

CI workflows:

- `.github/workflows/ci.yml` for mandatory platform/app/docs checks on PRs and pushes
- `.github/workflows/deploy-preflight.yml` for manual release preflight by target

## Status

Project X is in early foundation stage.

Completed in this repo (March 2026):

- Canonical top-level structure and ownership guide: `docs/repo-structure.md`
- Core operator/scaffolding surfaces:
  - `platform-cli` command modules (create/project/tokens/notifications/control-plane/deploy/docs)
  - `scripts/verify` and `scripts/deploy-preflight` for validation gates
- Active product/runtime modules:
  - `apps/cloud-console`
  - `services/omnichannel`
  - `packages/sdk-omnichannel`
- Reliability and release gates:
  - `scripts/verify`
  - `scripts/deploy-preflight`
  - `.github/workflows/ci.yml`
  - `.github/workflows/deploy-preflight.yml`

The existing notification MVP and CLI starter remain seed implementations for the broader platform architecture described above.
