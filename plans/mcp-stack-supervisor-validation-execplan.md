# Validate MCP under the repo stack supervisor

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

Linear ticket linkage for this plan: `ANM-173` ([issue link](https://linear.app/anmho/issue/ANM-173/medium-validate-mcp-service-under-the-repo-stack-supervisor)).

## Purpose / Big Picture

Verify that `services/mcp` runs correctly under the repository's stack supervisor, not only as a standalone server or Docker container. The goal is to prove the stack-managed path works and fix any supervisor bug uncovered along the way.

## Progress

- [x] (2026-03-18 07:15Z) Created `ANM-173`, set it to `In Progress`, assigned it to `me`, and recorded the session UUID.
- [x] (2026-03-18 07:16Z) Confirmed `mcp` is discoverable from `scripts/dev-stack-discover.py` and resolves to `cd services/mcp && go run ./cmd/server` with `MCP_ROOT` set to the repo root.
- [x] (2026-03-18 07:16Z) Reproduced the failure: `scripts/dev-stack start mcp` reported success, but `scripts/dev-stack status` immediately showed `mcp` as not running.
- [x] (2026-03-18 07:16Z) Confirmed the recorded PID was already dead and `mcp.log` remained empty, indicating a supervisor launch bug rather than a server runtime error.
- [ ] Patch the stack supervisor launch path for discovered services and retest `mcp` under stack management.
- [ ] Re-scan for follow-up issues, sync plan notes, and align `ANM-173`.

## Surprises & Discoveries

- Observation: `scripts/dev-stack` currently backgrounds discovered services inside a subshell and then exits that subshell immediately.
  Evidence: `run_discovered()` in `scripts/dev-stack` uses `( ... eval "$run" >> log 2>&1 & echo $! > pid )`; the recorded PID no longer existed immediately after launch.

## Decision Log

- Decision: fix the supervisor launch path instead of treating this as an `mcp`-specific problem.
  Rationale: discovery and service config for `mcp` were correct; the failure happened before the service-specific runtime could even stay attached to the supervisor.
  Date/Author: 2026-03-18 / Codex

## Outcomes & Retrospective

Completed outcomes:

- Pending.

Remaining gaps:

- Stack-managed `mcp` has not been revalidated yet after the supervisor fix.

## Context And Orientation

Relevant files for this task:

- `scripts/dev-stack`
- `scripts/dev-stack-discover.py`
- `services/mcp/stack.json`
- `.logs/dev-stack/mcp.log`

## Plan Of Work

1. Patch `scripts/dev-stack` so discovered services survive supervisor launch.
2. Start `mcp` through the stack supervisor and verify status/logs.
3. Confirm live MCP access against the stack-managed process.
4. Reconcile any follow-up ticketing and finalize `ANM-173`.

## Validation And Acceptance

Acceptance criteria:

- `scripts/dev-stack start mcp` leaves the process running.
- `scripts/dev-stack status` reports `mcp` as running.
- `./platform mcp --server http://127.0.0.1:8765 --key <generated-key> tools list` succeeds against the stack-managed server.

Validation commands:

- `scripts/dev-stack start mcp`
- `scripts/dev-stack status`
- `tail -n 80 .logs/dev-stack/mcp.log`
- `./platform mcp --server http://127.0.0.1:8765 --key <generated-key> tools list`

## Artifacts And Notes

- Linear issue: `ANM-173`
- Session UUID: `019cffb4-e861-7a31-b41f-5e3eeb6c35a9`
