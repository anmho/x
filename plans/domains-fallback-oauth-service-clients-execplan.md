# Fix Domains Offline Failure And Add OAuth Service Client Management

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear linkage:

- Primary issue: `ANM-82` (`https://linear.app/anmho/issue/ANM-82/medium-fix-domains-fetch-failure-and-add-oauth-client-credentials`)
- Default project: `X Platform` (`https://linear.app/anmho/project/x-platform-de7a3e703d58`)
- Canonical tracking project reference: `Project X Agent Execution` (`https://linear.app/anmho/team/ANM/projects/all`)

## Purpose / Big Picture

After this change:

- `/domains` no longer hard-fails with `fetch failed` when the control-plane service is offline in local development.
- Domain listing/records/reconcile API routes transparently fall back to local config-backed data when control-plane fetch is unavailable.
- Settings include an OAuth service-to-service management section for `client_id` + `client_credentials` style clients.

Observable behavior:

- Visiting `/domains` shows domains/records from local fallback data (seeded from `platform.controlplane.json`) if `http://127.0.0.1:8091` is unreachable.
- DNS CRUD operations work in fallback mode via in-memory route handlers.
- `/settings?tab=security` includes OAuth client creation, rotation, revocation, and scope display for `client_credentials` grant flow.

## Progress

- [x] (2026-03-10T06:04:09Z) Created and started Linear issue `ANM-82`.
- [x] (2026-03-10T06:10:00Z) Implemented control-plane offline fallback for `/api/domains*` routes.
- [x] (2026-03-10T06:11:00Z) Added OAuth service-to-service section with `client_id` and `client_credentials` management in settings security tab.
- [x] (2026-03-10T06:12:21Z) Ran TypeScript validation for omnichannel frontend (`npm exec tsc -- --noEmit`).
- [x] (2026-03-10T06:13:52Z) Completed post-change audit/rescan and filed follow-up tickets `ANM-83` and `ANM-84`.
- [ ] (2026-03-10T06:12:21Z) Full production build validation is blocked by pre-existing missing dependencies in workspace (`axios`, `picocolors`, `tslib`).

## Surprises & Discoveries

- Observation: `./platform status` cannot run from this repository state because the local wrapper expects `scripts/dev-stack` root shape that is not present.
  Evidence: command returned `platform: could not find repo root from /Users/andrewho/repos/projects/x (expected scripts/dev-stack)`.

- Observation: `PLANS.md` is not present in the repository even though multiple plans reference it.
  Evidence: `rg --files -g 'PLANS.md'` returned no matches.

- Observation: `apps/omnichannel/frontend` build currently fails due existing dependency issues unrelated to this patch.
  Evidence: `npm run build` errors include unresolved `axios`, `picocolors`, and `tslib`.

## Decision Log

- Decision: keep control-plane as primary source and only use local fallback when the network fetch to control-plane fails.
  Rationale: preserves real control-plane behavior in live mode while making local UI resilient during service downtime.
  Date/Author: 2026-03-10 / Codex

- Decision: seed fallback domains/records from `platform.controlplane.json` and maintain in-memory mutations for local iteration.
  Rationale: avoids introducing new persistence infrastructure while enabling working CRUD behavior in dev.
  Date/Author: 2026-03-10 / Codex

- Decision: place OAuth service-client management in Settings > Security and persist records to browser localStorage.
  Rationale: matches user request for section management while keeping implementation scoped to frontend UX without backend contract changes.
  Date/Author: 2026-03-10 / Codex

## Outcomes & Retrospective

Implemented:

- Added `domains-local-store` server-only fallback store with config seeding and in-memory record CRUD.
- Updated `/api/domains`, `/api/domains/[domain]/records`, `/api/domains/[domain]/records/[recordId]`, and `/api/domains/reconcile` to fail over on control-plane unavailability.
- Added explicit `ControlPlaneUnavailableError` to improve error semantics.
- Added OAuth service-to-service client management UI in settings security tab (`client_id`, `client_credentials`, scopes, rotate, revoke).

Remaining gaps:

- Full `npm run build` for `apps/omnichannel/frontend` remains blocked by existing dependency resolution issues not introduced by this change.
- OAuth service-client management is frontend-local (localStorage) and not yet backed by secure server-side storage.
- Cloud-console settings parity for OAuth service-client section is tracked as `ANM-84`.
- Omnichannel frontend dependency restore to unblock full build validation is tracked as `ANM-83`.

## Context And Orientation

Primary files:

- `apps/omnichannel/frontend/lib/domains-control-plane.ts`
- `apps/omnichannel/frontend/lib/domains-local-store.ts`
- `apps/omnichannel/frontend/app/api/domains/route.ts`
- `apps/omnichannel/frontend/app/api/domains/[domain]/records/route.ts`
- `apps/omnichannel/frontend/app/api/domains/[domain]/records/[recordId]/route.ts`
- `apps/omnichannel/frontend/app/api/domains/reconcile/route.ts`
- `apps/omnichannel/frontend/app/settings/page.tsx`

## Plan Of Work

1. Create/link Linear ticket and move to `In Progress`.
2. Patch domains API routes to support local fallback when control-plane fetch is unavailable.
3. Add OAuth service-to-service section with `client_id` / `client_credentials` management in settings.
4. Validate via TypeScript/build checks and document blockers.
5. Reconcile outcomes with Linear ticket status and follow-up findings.

## Concrete Steps

From repo root:

    rg -n "domains|control-plane|oauth|client_credentials" apps/omnichannel/frontend -S

From `apps/omnichannel/frontend`:

    npm exec tsc -- --noEmit
    npm run build

## Validation And Acceptance

Acceptance criteria:

- `/domains` no longer shows immediate `fetch failed` when control-plane process is down.
- DNS listing and record API handlers return successful JSON in local fallback mode.
- Settings security tab includes OAuth service-to-service management with visible `client_id` and masked credentials handling.

Validation results:

- `npm exec tsc -- --noEmit` passed.
- `npm run build` failed due pre-existing missing dependencies (`axios`, `picocolors`, `tslib`) unrelated to touched files.

## Idempotence And Recovery

- Re-running this change with control-plane online keeps control-plane responses authoritative.
- Fallback store only activates when control-plane network call fails and keeps data in-memory for the process lifetime.
- Removing fallback behavior is isolated to route handlers and `domains-local-store.ts`.

## Artifacts And Notes

- Linear issue: `ANM-82`
- Screenshot context: user observed `/domains` error banner with message `fetch failed` while control-plane was offline.

## Interfaces And Dependencies

- No new external APIs added.
- Fallback mode depends on presence of `platform.controlplane.json` when available.
- OAuth service-client management currently depends on browser `localStorage` key `console:oauth:service-clients:v1`.
