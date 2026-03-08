# Implement `c/domains` With Unified Platform Control Plane DNS

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

`PLANS.md` exists in this repository, and this document follows its structure and execution rules.

## Purpose / Big Picture

After this change, the cloud console Domains surface works end-to-end:

- `/domains` renders a functional DNS management interface.
- `c/domains` resolves to `/domains`.
- DNS records can be listed/created/updated/deleted via a unified Go control plane.
- Control-plane reconciliation supports domains with Cloudflare and Vercel providers.

Observable behavior:

- `./platform control-plane serve --addr :8091` exposes `/v1/domains*` APIs.
- Console domain APIs proxy to that control plane.
- `./platform control-plane plan|apply --project cloud-console` includes domain reconciliation logs.

## Progress

- [x] (2026-03-08T22:12:44Z) Added commit/mistake policy baseline (`AGENTS.md`, `docs/agent-mistakes.md`).
- [x] (2026-03-08T22:14:53Z) Restored frontend routes for `/domains` and `c/domains` alias and protected routing.
- [x] (2026-03-08T22:17:49Z) Added control-plane domain schema/contracts and declarative materialization support.
- [x] (2026-03-08T22:19:53Z) Implemented Cloudflare DNS adapter with unit tests.
- [x] (2026-03-08T22:21:31Z) Implemented Vercel DNS adapter with unit tests.
- [x] (2026-03-08T22:23:32Z) Integrated domain reconcile flow with create/update/delete/no-op drift handling and tests.
- [x] (2026-03-08T22:25:56Z) Added control-plane HTTP server mode and `/v1/domains*` endpoints with handler tests.
- [x] (2026-03-08T22:28:10Z) Wired console `/api/domains/*` proxy routes and live Domains UI with deployment-linked defaults + manual overrides.
- [ ] (2026-03-08T22:30:49Z) Run final cross-repo validation and publish final evidence transcript (completed: `go test ./platform-cli/...`, `python3 scripts/ci/materialize_platform_configs.py`, `./platform build`, `npm run build` in frontend; remaining: live `control-plane serve` smoke blocked by sandbox constraints).

## Surprises & Discoveries

- Observation: `services/omnichannel/frontend` is a nested Git repository, so commits must be executed there directly.
  Evidence: `git rev-parse --show-toplevel` returns `/Users/andrewho/repos/projects/x/services/omnichannel/frontend` when run in that directory.

- Observation: dynamic Next.js route paths require shell quoting in command execution.
  Evidence: unquoted path `app/api/domains/[domain]/...` failed with `zsh: no matches found`.

- Observation: sandbox constraints blocked local runtime smoke test for control-plane HTTP startup.
  Evidence: startup probe command returned `nice(5) failed: operation not permitted` and could not connect to `127.0.0.1:8091`.

## Decision Log

- Decision: keep provisioning centralized in the unified platform control plane, and expose it via `control-plane serve`.
  Rationale: satisfies direct frontend-to-monolith control-plane requirement while preserving one provisioning owner.
  Date/Author: 2026-03-08 / Codex

- Decision: implement provider adapters directly for `cloudflare` and `vercel` in v1.
  Rationale: requested multi-provider day one without generic placeholder-only abstractions.
  Date/Author: 2026-03-08 / Codex

- Decision: include deployment-linked defaults in UI as prefill suggestions, not strict constraints.
  Rationale: enables quick setup while preserving manual DNS flexibility.
  Date/Author: 2026-03-08 / Codex

## Outcomes & Retrospective

Implemented:

- working `/domains` route and `c/domains` alias
- control-plane domain schema + reconciliation
- Cloudflare/Vercel provider adapters with tests
- control-plane HTTP APIs for domains and reconcile
- console API proxy routes and live DNS CRUD UI
- policy additions for granular commits + immediate mistake logging

Remaining gaps:

- end-to-end live-provider verification requires real Cloudflare/Vercel credentials and zones.
- domain APIs currently read desired state from materialized control-plane config and do not write back to declarative source.

## Context And Orientation

Key files for this initiative:

- `platform-cli/control_plane_ops.go`: control-plane CLI reconciliation entrypoints.
- `platform-cli/control_plane_server.go`: control-plane HTTP server and `/v1/domains*` handlers.
- `platform-cli/domains_provider_cloudflare.go`: Cloudflare DNS adapter.
- `platform-cli/domains_provider_vercel.go`: Vercel DNS adapter.
- `services/omnichannel/frontend/app/domains/page.tsx`: Domains UI.
- `services/omnichannel/frontend/app/api/domains/*`: frontend server-side proxies to control plane.
- `infra/platform/declarative_spec.py`: declarative config source used to materialize `platform.controlplane.json`.

Non-obvious terms:

- "Unified platform control plane": the single Go provisioning monolith accessed by CLI and console.
- "Deployment-linked default": suggested DNS record prefill derived from known deployment identity, still editable by users.

## Plan Of Work

1. Enforce policy baseline for commit slicing and mistake capture.
2. Restore broken route (`/domains` and `c/domains` compatibility).
3. Add domain schemas/contracts and materialization support.
4. Implement provider adapters (Cloudflare and Vercel).
5. Add reconciliation engine support for domain records.
6. Expose domain HTTP APIs from the control plane.
7. Wire frontend API proxy and live UI workflows.
8. Run tests/builds and document operational usage.

## Concrete Steps

From `/Users/andrewho/repos/projects/x`:

    go test ./platform-cli/...
    python3 scripts/ci/materialize_platform_configs.py

From `/Users/andrewho/repos/projects/x/services/omnichannel/frontend`:

    npm run build

Control-plane API smoke:

    ./platform control-plane serve --addr :8091
    curl -sS http://127.0.0.1:8091/health
    curl -sS http://127.0.0.1:8091/v1/domains

## Validation And Acceptance

Acceptance criteria:

- `c/domains` lands on a working Domains UI.
- UI can list/create/update/delete records through `/api/domains/*`.
- `./platform control-plane plan|apply --project cloud-console` includes domain reconciliation output.
- provider adapters and reconciliation logic are covered by `go test ./platform-cli/...`.
- frontend production build passes.

## Idempotence And Recovery

- Re-running control-plane reconciliation is safe; no-op paths are explicit when records already match desired state.
- If reconciliation is configured with `desired_state: absent`, deletions only occur with `--prune` (or `destroy`).
- Recovery path for bad DNS changes: update desired config and rerun reconcile, or delete incorrect records via Domains UI/API.

## Artifacts And Notes

Validation outputs captured during implementation:

- `go test ./platform-cli/...` passed after adapter/reconciler/server additions.
- `npm run build` in `services/omnichannel/frontend` passed and emitted routes including:
  - `/domains`
  - `/c/domains`
  - `/api/domains`
  - `/api/domains/[domain]/records`
  - `/api/domains/[domain]/records/[recordId]`
  - `/api/domains/reconcile`

## Interfaces And Dependencies

Control-plane domain schema (`platform.controlplane.json` project entry):

- `domains[]` with `name`, `provider`, optional `zone_id`, optional `desired_state`, optional `records[]`.
- `records[]` with `type`, `name`, `content`, optional `ttl`, optional `proxied`, optional `desired_state`, optional `deployment_link`.

Control-plane HTTP surface:

- `GET /health`
- `GET /v1/domains`
- `GET /v1/domains/{domain}/records?provider=<provider>`
- `POST /v1/domains/{domain}/records?provider=<provider>`
- `PATCH /v1/domains/{domain}/records/{recordId}?provider=<provider>`
- `DELETE /v1/domains/{domain}/records/{recordId}?provider=<provider>`
- `POST /v1/domains/reconcile`

Dependencies:

- Cloudflare API token (`CLOUDFLARE_API_TOKEN`)
- Vercel API token (`VERCEL_API_TOKEN`)
