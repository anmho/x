# Enforce MCP-only runtime tools for agent execution

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-209` ([issue link](https://linear.app/anmho/issue/ANM-209/major-enforce-mcp-only-tool-registration-in-claudecodex-runtime)), with scoped follow-ons in `ANM-215` (workspace/git MCP tools), `ANM-216` (durable tool-call events), and `ANM-210` (dockerized Bun worker path).

## Purpose / Big Picture

Shift the Claude-first data plane from vendor-native workspace tools to Project X MCP tools. The runtime worker should only use capabilities exposed through `services/mcp`, so workspace and git operations are policy-controlled, auditable, and consistent across Claude and future Codex support.

This slice focuses on the minimum practical foundation:

- add MCP-backed workspace and git tools needed for local repo work
- reconfigure the Bun Claude worker to allow only MCP tools
- validate the local MCP tool path and the updated worker configuration

Tool-call durability in the control plane and the final dockerized worker path remain separate tracked follow-ons so this slice can land a coherent enforcement boundary first.

## Progress

- [x] (2026-03-25 23:55Z) Confirmed `ANM-209` is the active integration ticket, moved it to `In Progress`, and split follow-on scope into `ANM-215` (workspace/git MCP tools) and `ANM-216` (durable tool-call events).
- [x] (2026-03-25 23:58Z) Audited the current MCP and Bun worker foundations. Verified that `services/mcp/internal/tools/registry.go` is collaboration-only today and `apps/agent-runner/src/claude.ts` still explicitly allows native Claude tools (`Read`, `Write`, `Edit`, `Bash`, `Glob`, `Grep`).
- [x] (2026-03-26 07:00Z) Implemented repo-scoped MCP workspace/git tools (`workspace_read_file`, `workspace_list_files`, `workspace_search`, `workspace_apply_patch`, `git_status`, `git_diff`, `git_checkout`) with focused Go tests.
- [x] (2026-03-26 07:04Z) Reconfigured the Claude worker to fail closed on MCP-only tools by defaulting an MCP config path, whitelisting only `mcp__tools__...` names, and removing native Claude workspace tools from `allowedTools`.
- [x] (2026-03-26 07:15Z) Verified the live MCP server path on an isolated port: `tools/list` includes the new workspace/git tools, and live `workspace_read_file` / `git_status` calls succeeded over JSON-RPC with auth.
- [ ] Complete a full live Bun worker invocation through the normal `src/index.ts` path.

## Surprises & Discoveries

- Observation: The Go MCP server already exposes one centralized tool registry and JSON-RPC passthrough, so the new workspace/git tools can land without introducing a second MCP service.
  Evidence: `services/mcp/internal/tools/registry.go`, `services/mcp/internal/jsonrpc/passthrough.go`, and `services/mcp/cmd/server/main.go` already handle `tools/list` and `tools/call` for all registered tools.

- Observation: The Bun worker is already configured to load external MCP server definitions from `apps/agent-runner/mcp-config.json`, but it simultaneously enables native Claude workspace tools.
  Evidence: `apps/agent-runner/src/claude.ts` passes both `mcpServers` and `allowedTools: ["Read", "Write", "Edit", "Bash", "Glob", "Grep"]` to `query(...)`.

- Observation: The current MCP registry implementation is monolithic but straightforward, so adding a new helper module for workspace/git behavior is lower risk than expanding the big switch inline forever.
  Evidence: `Registry.Call` in `services/mcp/internal/tools/registry.go` already contains a large name switch for collaboration and platform tools.

- Observation: The checked-in `apps/agent-runner/mcp-config.json` was not usable for real local runs as-is because it targeted the wrong default port and lacked an auth path.
  Evidence: Live MCP testing required fixing the URL to `http://localhost:8765/mcp` and adding header expansion for `Authorization: Bearer ${MCP_API_KEY}`.

- Observation: The full Bun worker entrypoint is still blocked before Claude execution by the already-tracked local package resolution issue in `@x/sdk-agent-control`.
  Evidence: `bun src/index.ts` failed with `Cannot find module '@bufbuild/protobuf/codegenv2' from .../packages/sdk-agent-control/src/gen/...`, which is tracked in `ANM-206`.

- Observation: A direct Claude SDK invocation that bypassed `@x/sdk-agent-control` got past the MCP config layer but still exited with a generic non-zero Claude process error and no surfaced stderr context.
  Evidence: `bun -e 'import { runWithClaude } ...'` exited from `@anthropic-ai/claude-agent-sdk/sdk.mjs` with `Claude Code process exited with code 1`.

## Decision Log

- Decision: Start with first-party MCP tools for repo-safe workspace and git actions instead of exposing a generic shell immediately.
  Rationale: The runtime specifically needs patch/read/search/checkout/status/diff capabilities. A narrower tool surface is easier to audit and less likely to bypass repo safety rules.
  Date/Author: 2026-03-25 / Codex

- Decision: Enforce MCP-only behavior in the Claude worker by removing native Claude workspace tools from `allowedTools`, not by trying to post-process or forbid tool results after the fact.
  Rationale: The policy boundary should be explicit at runtime configuration time.
  Date/Author: 2026-03-25 / Codex

- Decision: Keep durable control-plane tool-call event persistence out of this slice and track it under `ANM-216`.
  Rationale: The immediate blocker is that the runtime still needs an MCP tool surface to work at all. Event durability can build cleanly on top of that boundary once the tools exist.
  Date/Author: 2026-03-25 / Codex

## Outcomes & Retrospective

Completed outcomes so far:

- `ANM-209` is now the active integration ticket for MCP-only runtime enforcement.
- The work is split into concrete follow-ons: `ANM-215` for the MCP tool surface and `ANM-216` for durable event plumbing.
- The original enforcement gap is now closed in the worker configuration: native Claude workspace tools are no longer whitelisted.
- Project X MCP now exposes the minimum repo-work tools needed by the runtime, backed by focused Go tests.
- The Bun Claude worker now defaults to an MCP config path, expands `${ENV_VAR}` placeholders in that config, and only whitelists MCP tool names for local workspace/repo work.
- Live MCP smoke validation succeeded against an isolated local server with auth.

Remaining gaps:

- The normal Bun worker entrypoint is still blocked by `ANM-206` (`@x/sdk-agent-control` local package resolution under Bun).
- Direct Claude SDK invocation still needs better runtime error surfacing or environment validation before the Claude-backed live path can be considered proven.
- Dockerization remains tracked separately in `ANM-210`.

## Context And Orientation

Primary files and modules in this scope:

- `services/mcp/internal/tools/registry.go`
- `services/mcp/cmd/server/main.go`
- `services/mcp/internal/jsonrpc/passthrough.go`
- `apps/agent-runner/src/claude.ts`
- `apps/agent-runner/mcp-config.json`
- `apps/agent-runner/README.md`

## Plan Of Work

1. Add a focused workspace/git tool module to the MCP server with typed inputs and repo-scoped safety checks.
2. Register those tools in the MCP registry and cover them with Go tests.
3. Reconfigure the Bun Claude worker to allow only Project X MCP tools for workspace/repo operations.
4. Run local validation for the MCP tool surface and the updated worker configuration.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Implement and register workspace/git tools:
   - `read_file`, `list_files`, `search_files`, `apply_patch`
   - `git_status`, `git_diff`, `git_checkout`
2. Add focused tests under `services/mcp/internal/tools`.
3. Update `apps/agent-runner/src/claude.ts` to remove native Claude workspace tools and allow only MCP-prefixed tools.
4. Validate:
   - `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`
   - `cd apps/agent-runner && bun test`
   - local smoke test with the MCP server plus a worker-local Claude invocation

## Validation And Acceptance

Acceptance criteria:

- Project X MCP exposes the minimum workspace/git tools needed by the agent runtime.
- The Claude worker configuration no longer whitelists native Claude workspace tools.
- Local validation shows the MCP server and worker configuration are both functional.
- The implementation leaves room for later durable tool-call event registration and dockerized runtime deployment.

Validation still required:

- `cd services/mcp && GOCACHE=/tmp/go-cache go test ./...`
- `cd apps/agent-runner && bun test`
- local smoke validation for the full Bun worker entrypoint once `ANM-206` is resolved

## Idempotence And Recovery

The MCP server changes should be additive. Repeated local validation should not mutate repo state except where an explicit tool such as `apply_patch` or `git_checkout` is invoked on a test fixture. The runtime reconfiguration should fail closed: if the MCP server is unavailable, the worker should not silently fall back to native Claude workspace tools.

## Artifacts And Notes

- Active integration ticket: `ANM-209`
- Active implementation ticket for this slice: `ANM-215`
- Follow-on durable event ticket: `ANM-216`
- Follow-on dockerization ticket: `ANM-210`

## Interfaces And Dependencies

- Go MCP server under `services/mcp`
- Bun Claude worker under `apps/agent-runner`
- Claude Agent SDK MCP configuration support
- Existing repo safety expectations from `AGENTS.md`
