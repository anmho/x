# Top Three Priorities (Current State)

This file tracks the highest-impact repository foundations that are currently present in code.

## 1) Canonical Monorepo Layout

Delivered:

- Top-level directories are established and active: `apps/`, `services/`, `agents/`, `packages/`, `infra/`, `scripts/`, `docs/`, `plans/`.
- Structure and ownership guidance lives in `docs/repo-structure.md`.

## 2) Platform CLI Scaffolding + Workflow Path

Delivered:

- Platform CLI surfaces are positioned for scaffolding and service/resource workflows (`create`, `config`, `project`, `tokens`, `notifications`, `control-plane`, `deploy`, `docs`).
- Validation guidance is centered on `scripts/verify` and `scripts/deploy-preflight`.
- Command focus is documented in:
  - `AGENTS.md`
  - `docs/runbooks/platform-cli-workflow.md`
  - `scripts/README.md`

## 3) Omnichannel + SDK Foundation

Delivered:

- Omnichannel backend surfaces are present under `services/omnichannel/` (`backend`, `backend-api`, `backend-worker`).
- Cloud Console app is present at `apps/cloud-console`.
- Shared SDK package is present at `packages/sdk-omnichannel`.
- SDK generation/publish entrypoint is `scripts/sdk.sh` (Nx-backed).
