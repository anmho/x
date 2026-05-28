# Build platform mcp setup wizard for portable Docker deployment

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-227` ([issue link](https://linear.app/anmho/issue/ANM-227/medium-add-platform-mcp-setup-wizard-for-portable-docker-deployment)).

## Purpose / Big Picture

Add a `platform mcp setup` wizard that makes the MCP service portable as a Docker-hosted or centrally hosted gateway. The wizard should collect the runtime settings the container actually needs, write a reusable env/config file, and print the exact commands needed to run the container with centralized auth and provider passthroughs.

## Progress

- [x] (2026-03-26T07:58:54Z) Created `ANM-227`, moved it to `In Progress`, assigned it to `me`, and recorded the active Codex session UUID in the ticket description.
- [x] (2026-03-26T07:58:54Z) Reviewed `platform-cli`, `services/mcp/Dockerfile`, and current MCP env docs to confirm the container contract and CLI integration point.
- [x] (2026-03-26T08:03:38Z) Added `platform mcp setup` wizard plus Docker env writer in `platform-cli`.
- [x] (2026-03-26T08:03:38Z) Kept existing `platform mcp tools|keys` passthrough behavior intact while adding setup/run guidance.
- [x] (2026-03-26T08:03:38Z) Validated the setup flow in isolated file-level tests, updated docs, and reconciled follow-up audit findings against existing ticket `ANM-153`.

## Surprises & Discoveries

- Observation: `platform mcp` is currently only a thin wrapper around the standalone `mcp` binary, so the wizard belongs in `platform-cli` rather than `services/mcp/cmd/mcp`.
  Evidence: `platform-cli/mcp.go` currently just discovers the `mcp` binary and `exec`s it.

- Observation: the portable container already has a stable runtime contract: env vars plus the persisted `/root/.x-mcp` volume.
  Evidence: `services/mcp/Dockerfile` sets `MCP_PORT` and `MCP_ROOT=/app`, exposes `8765`, and declares `VOLUME ["/root/.x-mcp"]`.

- Observation: full-package `platform-cli` validation is already blocked by a pre-existing source gap unrelated to this wizard slice.
  Evidence: `cd platform-cli && GOCACHE=/tmp/go-cache go test ./...` and `go build .` fail on unresolved handlers like `runStack`; existing ticket `ANM-153` already tracks the broken `platform-cli` source state.

## Decision Log

- Decision: implement a text-based terminal wizard using the standard library instead of introducing a new TUI dependency.
  Rationale: `platform-cli` is already a small flag/switch based Go CLI, and the immediate need is a portable setup workflow rather than a full-screen UI framework.
  Date/Author: 2026-03-26 / Codex

- Decision: have the wizard write a Docker-ready env file and supporting config under the user’s `~/.x-mcp` directory by default.
  Rationale: that matches the service’s existing persisted volume path and keeps secrets/config outside the repo worktree.
  Date/Author: 2026-03-26 / Codex

## Outcomes & Retrospective

Completed outcomes:

- `platform mcp setup` now runs an interactive terminal wizard that writes a Docker-ready env file under `~/.x-mcp/mcp.env` by default.
- The wizard supports gateway API key generation, optional Linear and Slack credentials, and optional upstream MCP proxy settings for providers such as Google Docs and Perplexity.
- The setup flow prints exact follow-up commands for `docker run` and `platform mcp` client usage so the same container config can be used locally or handed off for centralized hosting.

Remaining gaps:

- Full-package `platform-cli` build/test validation remains blocked by the pre-existing missing source tracked in `ANM-153`.
- The wizard currently uses a standard terminal prompt flow rather than hidden-input password entry or a richer full-screen TUI.

## Context And Orientation

Relevant files for this task:

- `platform-cli/main.go`
- `platform-cli/mcp.go`
- `services/mcp/Dockerfile`
- `services/mcp/README.md`
- `services/mcp/.env.example`

## Plan Of Work

1. Add a `platform mcp setup` command that runs an interactive terminal wizard.
2. Write the selected MCP runtime configuration to a reusable env file plus any supporting local metadata.
3. Print exact follow-up commands for Docker-hosted local usage and centralized deployment handoff.
4. Update docs and validate the wizard with a non-destructive temp-path run plus MCP build/test checks.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Update `platform-cli/mcp.go` and any helper files needed for setup/config writing.
2. Add safe default output paths under `~/.x-mcp`, with override flags for tests.
3. Update `services/mcp/README.md` and `.env.example`.
4. Run:
   - `go build ./platform-cli`
   - a non-destructive setup flow pointed at `/tmp/...`
   - `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`
   - `cd services/mcp && GOCACHE=/tmp/go-cache go build ./...`

## Validation And Acceptance

Acceptance criteria:

- `platform mcp setup` can collect MCP gateway/provider auth settings interactively.
- The wizard writes a Docker-usable env file without requiring repo edits.
- The wizard prints the exact `docker run` and `platform mcp ...` follow-up commands needed to use the configured container.
- Existing `platform mcp tools|keys` passthrough commands still work unchanged.

Validation to run:

- `go build ./platform-cli`
- non-destructive wizard run against temp output
- `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`
- `cd services/mcp && GOCACHE=/tmp/go-cache go build ./...`

Validation executed:

- `cd platform-cli && GOCACHE=/tmp/go-cache go test mcp.go mcp_setup_test.go`
- `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`
- `cd services/mcp && GOCACHE=/tmp/go-cache go build ./...`

Validation blocked:

- `cd platform-cli && GOCACHE=/tmp/go-cache go test ./...`
- `cd platform-cli && GOCACHE=/tmp/go-cache go build .`
  Blocked by pre-existing unresolved handlers tracked in `ANM-153`.

## Idempotence And Recovery

Repeated setup runs should safely overwrite or refresh the generated env file without mutating existing mailbox data. The wizard should avoid printing secrets back to the terminal after input is captured.

## Artifacts And Notes

- Active implementation ticket: `ANM-227`
- Existing follow-up/blocker ticket reused during audit: `ANM-153`
- Session UUID: `019cffb4-e861-7a31-b41f-5e3eeb6c35a9`

## Interfaces And Dependencies

- Existing `platform-cli` command dispatch
- `services/mcp` Docker runtime contract
- Existing env-driven provider proxy configuration in `services/mcp`
