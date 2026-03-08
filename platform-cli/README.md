# platform CLI (Go)

Supabase-style local CLI for Project X:

- stack lifecycle
- token lifecycle
- project config + deploy
- project-scoped key management
- scaffolding
- local diagnostics

## Usage

From repo root:

```bash
./platform start
./platform status
./platform logs access-api
./platform logs docs
./platform stack temporal-ui
./platform docs
./platform stop
```

Build a binary into `bin/platform`:

```bash
./platform build
```

Install into your PATH (defaults to `~/.local/bin/platform`):

```bash
./platform install
platform --help
```

Install to a custom directory:

```bash
./platform install /usr/local/bin
```

Equivalent long form:

```bash
./platform stack start
```

Token management:

```bash
./platform tokens mint \
  --application cloud-console \
  --owner andrew \
  --env dev \
  --scope notifications:write \
  --scope notifications:read

./platform tokens list
./platform tokens rotate --id <key-id>
./platform tokens revoke --id <key-id>
./platform tokens check --token <issued-token> --service notifications --scope write --env dev
./platform tokens audit
./platform tokens helpers --token <issued-token>
```

Project config and deploy:

```bash
./platform config init
./platform config show

./platform project list
./platform project show --project cloud-console
./platform project keys mint --project cloud-console --owner andrew --env dev
./platform project keys mint --project cloud-console --owner andrew --env dev --dry-run
./platform project keys helpers --project cloud-console --token <issued-token>

./platform deploy --project cloud-console
./platform deploy --project cloud-console --dry-run

./platform control-plane init
./platform control-plane show
./platform control-plane plan --project cloud-console
./platform control-plane apply --project cloud-console
./platform control-plane apply --prune
./platform control-plane destroy --project cloud-console
./platform control-plane serve --addr :8091
```

`platform.projects.json` defines:

- project dependencies (`stack`, `http`, `env`)
- deploy command/cwd/health check
- deployment declarations (`deployments[]`)
- integration declarations (`integrations[]`)
- platform targets (`gcp`, `endpoints`, `api_keys`)
- managed key scopes and env exports

`platform.controlplane.json` defines desired infra state:

- project-level desired state (`present` or `absent`)
- GCP project declaration (`project_id`, `region`, APIs)
- deployment declarations (metadata for provider/service/region)
- domain declarations (`projects[].domains[]`) for provider-managed DNS records
- account-level secret declarations (`accounts[].secrets[]`) with share targets
- optional project-level secret declarations (`projects[].secrets[]`) for compatibility
- optional `source_env` for secret value syncs during apply

For `provider: gcp-secret-manager`:

- `desired_state: present` ensures the secret exists.
- if `source_env` is set and present in the local shell, a new secret version is added on apply.
- `desired_state: absent` deletes the secret when using `--prune` (or `destroy`).

For `shares[]`:

- `platform: gcp` shares to Google Secret Manager in `project_id` (or project default).
- `platform: vercel` shares to Vercel project envs in listed `environments` (defaults: `development`, `preview`, `production`).
- `target_type` + `target_id` can annotate share ownership (`project`, `application`, `deployment`).
- `project` can be set on application/deployment shares so `--project <name>` only reconciles relevant account shares.
- this lets one account secret fan out to multiple projects/apps/deployments without per-app duplication.

When `targets.gcp` is configured, `./platform deploy --project <name>` now reconciles Google Cloud automatically before running the deploy command:

- verifies `gcloud` is installed
- creates missing GCP projects (`gcloud projects create`)
- links active gcloud project context (`gcloud config set project`)
- enables configured Google APIs (entries in `targets.gcp[].services` ending with `.googleapis.com`)
- writes sync metadata back to `platform.projects.json` (`project_number`, `linked`, `last_synced_at`) to reduce drift

If `platform.controlplane.json` exists, `platform deploy` first runs control-plane reconciliation for that project so CLI deploys stay aligned with declared infra state.

Domains support in the unified control plane:

- providers: `cloudflare`, `vercel`
- zone declaration: `name`, `provider`, optional `zone_id`, optional `desired_state`
- record declaration: `type`, `name`, `content`, optional `ttl`, optional `proxied`, optional `desired_state`
- reconcile behavior:
  - `desired_state: present` creates/updates drifted records
  - `desired_state: absent` deletes declared records only when `--prune` (or `destroy`) is used

Start the control-plane HTTP API for console/domain calls:

```bash
./platform control-plane serve --addr :8091
curl -sS http://127.0.0.1:8091/health
curl -sS http://127.0.0.1:8091/v1/domains
```

Config source-of-truth workflow:

- edit `infra/platform/declarative_spec.py`
- optional mutable integration overrides: `infra/platform/integrations.overrides.json`
- materialize JSON configs: `python3 scripts/ci/materialize_platform_configs.py`
- verify no drift: `python3 scripts/ci/materialize_platform_configs.py --check`

Start with:

- `./platform config init`
- or copy [platform.projects.example.json](/Users/andrewho/repos/projects/x/platform.projects.example.json) to `platform.projects.json`

Scaffolding and diagnostics:

```bash
./platform create service billing-api
./platform create cloud-console
./platform create integration add vercel --project cloud-console --provider vercel --set project_id=cloud-console
./platform create integration list --project cloud-console
./platform create integration remove vercel --project cloud-console
./platform docs dev --port 3002
./platform notifications list
./platform doctor
./platform verify all
./platform preflight platform
```

Environment variables:

- `ACCESS_API_URL` (default: `http://127.0.0.1:8090`)
- `ACCESS_API_ADMIN_KEY` (default: `x-admin-dev-key`)
- `OMNICHANNEL_API_URL` (default: `http://127.0.0.1:8080/api/v1`)
- `OMNICHANNEL_API_KEY` (default: `test-api-key-123`)
- `DOCS_PORT` (default: `3002`, used by `platform start` for Mintlify docs)
- `PLATFORM_CONTROL_PLANE_ADDR` (default: `:8091`, listen addr for `control-plane serve`)
- `CLOUDFLARE_API_TOKEN` (required for Cloudflare DNS operations)
- `CLOUDFLARE_API_BASE_URL` (optional, default: `https://api.cloudflare.com/client/v4`)
- `VERCEL_API_TOKEN` (required for Vercel DNS operations)
- `VERCEL_API_BASE_URL` (optional, default: `https://api.vercel.com`)

## SDK + Protocol Notes

- `platform tokens ...` and `platform project keys ...` use Access API REST endpoints (`/v1/...`) with JSON payloads.
- `@x/sdk-access` (`packages/sdk-access`) uses the same REST contract.
- ConnectRPC is currently used in omnichannel for Temporal workflow procedures, not Access API key management.
