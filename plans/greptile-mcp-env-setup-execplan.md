# Restore Greptile MCP startup with persistent GREPTILE_API_KEY

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

`PLANS.md` is checked into this repository, and this document follows the current ExecPlan structure used in `plans/`.

Linear ticket linkage for this plan: `ANM-127` ([issue link](https://linear.app/anmho/issue/ANM-127/minor-restore-greptile-mcp-startup-by-documenting-and-validating)).

## Purpose / Big Picture

Restore Greptile MCP startup for this repository by ensuring the missing `GREPTILE_API_KEY` is injected into the shell environment that launches Codex/Claude/Cursor-compatible MCP clients. The repository already contains correct Greptile MCP config; the missing piece is persistent secret setup and a clearer recovery path in documentation.

## Progress

- [x] (2026-03-10 09:25Z) Confirmed `Anmho` team and `X Platform` / `Project X Agent Execution` project availability in Linear.
- [x] (2026-03-10 09:25Z) Created Linear ticket `ANM-127` and moved it to `In Progress` before implementation.
- [x] (2026-03-10 09:26Z) Verified `.mcp.json` and `.codex/config.toml` already expect `GREPTILE_API_KEY`; startup failure was caused by a missing environment variable, not broken repo config.
- [x] (2026-03-10 09:27Z) Updated Greptile setup docs to document persistent shell export and the exact missing-env failure mode.
- [x] (2026-03-10 09:28Z) Persisted `GREPTILE_API_KEY` in local shell startup and verified a fresh login shell reads it.
- [x] (2026-03-10 09:28Z) Ran a focused follow-up audit for adjacent setup gaps; no additional repo follow-up tickets were needed for this scope.

## Surprises & Discoveries

- Observation: The repository already had both Claude-compatible and Codex-compatible Greptile MCP config checked in.
  Evidence: `.mcp.json` uses `Authorization: Bearer ${GREPTILE_API_KEY}` and `.codex/config.toml` sets `bearer_token_env_var = "GREPTILE_API_KEY"`.

- Observation: The current shell environment did not define `GREPTILE_API_KEY`.
  Evidence: `printenv GREPTILE_API_KEY` returned exit code `1` with no output.

- Observation: A fresh login shell loads the new export successfully, but startup also prints an unrelated `gitstatus` warning from the user's existing prompt configuration.
  Evidence: `zsh -lic 'printenv GREPTILE_API_KEY'` printed the key value after prompt initialization warnings.

## Decision Log

- Decision: Fix the issue by persisting the secret in local shell startup instead of embedding the raw key in tracked project config.
  Rationale: The project config is already correct and secret values must not be committed to version control.
  Date/Author: 2026-03-10 / Codex

- Decision: Add a narrow documentation update for the exact failure and recovery path.
  Rationale: The runbooks mention the secret requirement, but not the concrete startup symptom or the need to restart from a shell that already exports the variable.
  Date/Author: 2026-03-10 / Codex

## Outcomes & Retrospective

Completed outcomes:

- `GREPTILE_API_KEY` is now persisted in the local shell startup file used for new terminal sessions.
- A fresh login shell successfully resolves `GREPTILE_API_KEY`.
- Greptile runbooks now document the exact missing-env startup symptom and the required restart path.

Remaining gaps:

- The current Codex session was started before the export existed, so the user still needs to restart Codex or launch a new session to clear the already-failed MCP startup state.

## Context And Orientation

Relevant artifacts for this task:

- `.mcp.json`: project-level MCP config for Claude-compatible clients.
- `.codex/config.toml`: project-level MCP config for Codex CLI.
- `docs/runbooks/greptile-mcp-setup.md`: Greptile setup runbook.
- `docs/runbooks/cursor-cloud-parity.md`: secret and MCP parity checklist for cloud agents.
- `~/.zshrc`: local shell startup file expected to export the missing secret for future sessions.

The user-visible symptom is MCP startup failing with the message that `GREPTILE_API_KEY` is not set.

## Plan Of Work

Document the startup failure and recovery path in the Greptile runbooks, then persist the provided key in the local shell startup file outside version control. Validate with a fresh non-interactive login shell so the environment seen by newly launched tools matches expectations.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Update `docs/runbooks/greptile-mcp-setup.md` to explain the missing-env startup symptom and persistent export approach.
2. Update `docs/runbooks/cursor-cloud-parity.md` to call out the same secret dependency and restart requirement.
3. Persist `GREPTILE_API_KEY` in `~/.zshrc` using a local export block.
4. Validate:
   - `printenv GREPTILE_API_KEY` in a fresh shell
   - targeted `rg` checks for the new documentation text
5. Perform a focused audit for adjacent gaps and reconcile any follow-up tickets.

Expected result: Greptile config remains secret-free in git, a fresh shell resolves `GREPTILE_API_KEY`, and the docs explicitly cover this startup failure mode.

## Validation And Acceptance

Acceptance criteria:

- A fresh shell session prints a non-empty `GREPTILE_API_KEY`.
- Greptile runbooks document that missing `GREPTILE_API_KEY` causes MCP startup failure and that the launching shell must export it first.
- `ANM-127` status, this plan, and the final implementation state remain synchronized.

## Idempotence And Recovery

Reapplying the docs changes is safe if duplicate guidance is avoided. Reapplying the shell export should replace or preserve a single `GREPTILE_API_KEY` export line. If key rotation is required later, update only the local shell export and restart the Codex session; tracked repo config should remain unchanged.

## Artifacts And Notes

Validation targets:

- `rg -n "missing GREPTILE_API_KEY|restart Codex|fresh shell" docs/runbooks/greptile-mcp-setup.md docs/runbooks/cursor-cloud-parity.md`
- `zsh -lc 'printenv GREPTILE_API_KEY'`

Validation results:

- `zsh -lic 'printenv GREPTILE_API_KEY'` printed a non-empty value from a fresh login shell.
- `rg -n 'missing `GREPTILE_API_KEY`|restart Codex|fresh shell' docs/runbooks/greptile-mcp-setup.md docs/runbooks/cursor-cloud-parity.md` matched the new guidance text.

## Interfaces And Dependencies

No runtime service interfaces change. Dependencies for successful startup are:

- Valid Greptile API key
- Shell startup that exports `GREPTILE_API_KEY` before launching MCP clients
- Existing project MCP config in `.mcp.json` / `.codex/config.toml`

Revision note (2026-03-10): Initial plan created to resolve Greptile MCP startup failure caused by missing `GREPTILE_API_KEY`.
