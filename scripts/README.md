# Scaffolding Commands

Use `scripts/new` to scaffold common project types.

## Reliability Commands

Use these to harden local development and release safety:

```bash
scripts/doctor
python3 scripts/ci/materialize_platform_configs.py --check
scripts/verify all
scripts/deploy-preflight platform
scripts/dev-stack start
scripts/clean --dry-run --full
```

## Examples

```bash
scripts/new app cloud-console
scripts/new service access-api
scripts/new package sdk-notifications
scripts/new proto notifications
scripts/new agent support-triage
scripts/new mobile-app storefront
scripts/new smart-contract treasury
scripts/new cloud-console
```

## Why This Exists

These commands provide a fast path for the roadmap generators listed in the root README:

- `new:app`
- `new:cloud-console`
- `new:service`
- `new:agent`
- `new:package`
- `new:proto`
- `new:mobile-app`
- `new:smart-contract`

## Run Everything Live

Use `scripts/dev-stack` to run the local stack in one command:

```bash
scripts/dev-stack start
scripts/dev-stack status
scripts/dev-stack temporal-ui
scripts/dev-stack logs access-api
scripts/dev-stack stop
```

This starts/stops:

- Supabase local stack
- Temporal dev server
- `services/access-api`
- `services/omnichannel/backend` API + worker
- `apps/cloud-console` (includes omnichannel routes at `/omnichannel`) on `:3000`

For Go live reload, run with `wgo` enabled:

```bash
WATCH_GO_SERVICES=1 scripts/dev-stack start
# or
WATCH_GO_SERVICES=1 ./platform stack start
```

If `wgo` is missing, the stack falls back to `go run`.

Go wrapper CLI:

```bash
./platform stack start
./platform stack status
./platform stack temporal-ui
./platform stack stop
```

Token and scaffold helpers through Go CLI:

```bash
./platform tokens mint --application cloud-console --owner andrew --env dev --scope notifications:write
./platform tokens list
./platform new service billing-api
```

## Create PRs for Open Branches

Create PRs for branches pushed to origin that don't yet have one. Requires `gh auth login`.

```bash
./scripts/create-prs-for-branches.sh --dry-run   # preview
./scripts/create-prs-for-branches.sh             # create
```

## Linear Issue Creation

Use `scripts/linear/create-issues.mjs` to create Linear tickets from JSON payloads.

```bash
node scripts/linear/create-issues.mjs --dry-run
LINEAR_API_KEY=<your-api-key> LINEAR_TEAM_KEY=ENG node scripts/linear/create-issues.mjs
```

## Web Scraper

Use `scripts/web-scraper/scrape.mjs` for quick metadata extraction from one or more URLs.

```bash
node scripts/web-scraper/scrape.mjs --url https://example.com
node scripts/web-scraper/scrape.mjs --urls-file urls.txt --out scrape-output.json
```

## PM Agent

Use `agents/pm-agent/run.mjs` to generate a daily PM report and optional SMS digest.

```bash
node agents/pm-agent/run.mjs --dry-run
```

Clawdbot webhook delivery is also supported via `config.clawdbot` (see `agents/pm-agent/README.md`).

## Code Agents (Linear)

Use `agents/code-agents/run.mjs` to pull Linear issues and route them to `code-feature-agent`, `code-bugfix-agent`, or `code-refactor-agent`.

```bash
node agents/code-agents/run.mjs --issues-json agents/code-agents/sample-issues.json --dry-run
LINEAR_API_KEY=<your-api-key> LINEAR_TEAM_KEY=ENG node agents/code-agents/run.mjs --config agents/code-agents/config.example.json
```

Outputs are written to `agents/code-agents/out/`:

- `queue.json` with selected issues and assigned agent profile
- `briefs/<ISSUE-ID>.md` per-issue execution brief
