# Make pyenv shell initialization consistent in Cursor terminal

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

`PLANS.md` is checked into this repository, and this document follows the section structure used by current ExecPlans in `plans/`.

Linear ticket linkage for this plan: `ANM-230` ([issue link](https://linear.app/anmho/issue/ANM-230/medium-make-pyenv-shell-initialization-consistent-in-cursor-terminal)).

## Purpose / Big Picture

After this change, `pyenv` should initialize consistently in both Terminal.app and Cursor's integrated terminal, even when the shell starts as interactive non-login `zsh`. The immediate goal is to make `pyenv shell <version>` reliably affect the current shell session regardless of terminal host.

## Progress

- [x] (2026-03-26 08:05Z) Confirmed the active Codex session UUID (`019d292a-308e-7671-9f9a-7f4262014653`) and created `ANM-230`, assigned to `me`, in `In Progress`.
- [x] (2026-03-26 08:05Z) Verified existing shell state: `type pyenv` resolves to a shell function from `~/.zshrc`, while `PYENV_ROOT` and pyenv path entries are currently defined in `~/.zprofile`.
- [x] (2026-03-26 08:11Z) Patched `~/.zshrc` so interactive non-login shells set the minimal pyenv environment before running `pyenv init`.
- [x] (2026-03-26 08:13Z) Logged a wrapper mistake immediately in `docs/agent-mistakes.md` after a first attempt used zsh's readonly `status` parameter and broke `pyenv shell`.
- [x] (2026-03-26 08:26Z) Replaced the failed wrapper with a safe `pyenv` shell-function redefinition that mirrors pyenv's own dispatch and prints `command pyenv version` after successful `pyenv shell ...`.
- [x] (2026-03-26 08:26Z) Validated `pyenv shell system` and `pyenv shell 3.10.6` behavior in a fresh interactive `zsh`, including `python --version`.
- [x] (2026-03-26 08:26Z) Ran the focused audit/reconciliation pass for adjacent shell-config issues and mapped existing follow-up coverage.

## Surprises & Discoveries

- Observation: `pyenv shell 3.10.6` is silent on success, which can look like a no-op even when the shell version actually changes.
  Evidence: an interactive `zsh` session updated `PYENV_VERSION` and `python --version` without emitting command output.

- Observation: the current pyenv bootstrap is split across files by shell mode.
  Evidence: `PYENV_ROOT` plus `$PYENV_ROOT/bin` and `$PYENV_ROOT/shims` are set in `~/.zprofile`, while `eval "$(pyenv init - --no-rehash)"` and `eval "$(pyenv virtualenv-init -)"` live in `~/.zshrc`.

- Observation: the user's global pyenv version is already `3.10.6`, so a prompt-only pyenv segment will not make `pyenv shell 3.10.6` visibly different unless the shell override source itself is surfaced.
  Evidence: `pyenv global` returned `3.10.6` and `pyenv version-name` remained `3.10.6` after `pyenv shell 3.10.6`.

- Observation: pyenv's generated shell function handles `shell` by `eval "$(pyenv "sh-shell" ...)"`, so copying that function and then wrapping it can accidentally recurse back through the wrapper.
  Evidence: inspecting `functions pyenv` showed the dispatch body calling `pyenv "sh-$command"`, which explains the earlier repeated wrapper failures.

## Decision Log

- Decision: keep the main fix minimal by making `~/.zshrc` self-sufficient for pyenv initialization instead of sourcing the whole login profile from interactive shells.
  Rationale: Cursor and Terminal.app may launch different zsh modes, and only the pyenv-specific environment needs to be shared across both.
  Date/Author: 2026-03-26 / Codex

- Decision: make `pyenv shell ...` print the resolved version directly instead of relying on prompt changes.
  Rationale: the user's global version already matches `3.10.6`, so prompt-only changes do not create an obvious visual delta; command-level confirmation is the clearest fix.
  Date/Author: 2026-03-26 / Codex

- Decision: redefine the `pyenv` shell function explicitly using pyenv's own dispatch pattern rather than copying the generated function.
  Rationale: an explicit definition can call `command pyenv "sh-$command"` directly and avoids recursive wrapper behavior.
  Date/Author: 2026-03-26 / Codex

## Outcomes & Retrospective

Completed outcomes:

- `~/.zshrc` now sets `PYENV_ROOT`, `$PYENV_ROOT/bin`, and `$PYENV_ROOT/shims` for interactive non-login shells before `pyenv init`.
- `pyenv shell <version>` now prints the active pyenv version once on success, so Cursor no longer looks like a silent no-op for this command.
- Fresh interactive-shell validation confirmed:
  - `pyenv shell system` prints `system (set by PYENV_VERSION environment variable)`
  - `pyenv shell 3.10.6` prints `3.10.6 (set by PYENV_VERSION environment variable)`
  - `python --version` resolves to `Python 3.10.6`
- Mistake logging is up to date in `docs/agent-mistakes.md`.
- Follow-up ticket reconciliation found existing coverage for adjacent shell-performance work in `ANM-99` and `ANM-112`; no new follow-up ticket was required for this scope.

Remaining gaps:

- The user's currently open Cursor terminal must run `exec zsh` or open a new terminal tab to load the updated `~/.zshrc`.

## Context And Orientation

This task affects local shell startup behavior rather than repository runtime code. The relevant files are:

- `~/.zprofile`: login-shell environment setup, including `PYENV_ROOT` and pyenv path entries.
- `~/.zshrc`: interactive-shell setup, including pyenv shell-function initialization.
- `docs/agent-mistakes.md`: required log of the wrapper mistake discovered during implementation.
- `plans/pyenv-cursor-terminal-init-execplan.md`: execution record for this change.

## Plan Of Work

Make the pyenv bootstrap self-contained for interactive shells by defining `PYENV_ROOT` and the required pyenv path entries before the existing `pyenv init` block in `~/.zshrc`, and make `pyenv shell ...` print the resolved version after a successful shell change so the effect is visible in Cursor.

## Concrete Steps

From repo root (`/Users/andrewho/repos/projects/x`):

1. Edit `~/.zshrc` to set `PYENV_ROOT` if needed and add `$PYENV_ROOT/bin` and `$PYENV_ROOT/shims` before the existing `pyenv init` block.
2. Redefine the interactive `pyenv` shell function using pyenv's native dispatch pattern and print `command pyenv version` after successful `pyenv shell ...`.
3. Keep the current `~/.zprofile` content intact unless validation proves cleanup is necessary.
4. Validate with:
   - `zsh -i -c 'type pyenv'`
   - `zsh -i -c 'pyenv shell 3.10.6'`
   - `zsh -i -c 'pyenv shell system; pyenv shell 3.10.6; python --version; echo $PYENV_VERSION'`
5. Run a focused shell-config audit for adjacent unfinished work or bugs and reconcile follow-up tickets in Linear.

Expected result: both login and non-login interactive shells have enough pyenv environment to make `pyenv shell` work consistently, and the command itself confirms the selected version in Cursor.

## Validation And Acceptance

Acceptance criteria:

- A fresh interactive `zsh` resolves `pyenv` as a shell function.
- `pyenv shell system` and `pyenv shell 3.10.6` change `PYENV_VERSION` in the current shell session.
- `pyenv shell 3.10.6` prints the resolved pyenv version once on success.
- `python --version` matches the pyenv-selected version after switching back to `3.10.6`.
- `ANM-230` remains assigned to the current session owner and tracks the validation outcome.

## Idempotence And Recovery

Reapplying the same `~/.zshrc` additions is safe if duplicate lines are avoided. If the change causes shell startup regressions, remove the added pyenv environment lines and the explicit `pyenv()` redefinition from `~/.zshrc` and rely on the original login-shell-only bootstrap in `~/.zprofile`.

## Artifacts And Notes

Primary artifacts:

- `~/.zshrc`
- `~/.zprofile`
- `docs/agent-mistakes.md`
- `plans/pyenv-cursor-terminal-init-execplan.md`
- Linear issue `ANM-230`

## Interfaces And Dependencies

No application runtime interfaces change. Dependencies are:

- `zsh` shell startup ordering (`.zprofile` for login shells, `.zshrc` for interactive shells)
- Homebrew-installed `pyenv`
- Installed pyenv Python version `3.10.6`

Revision note (2026-03-26): Initial plan created for pyenv bootstrap consistency in Cursor's integrated terminal.
Revision note (2026-03-26): Updated to final state with interactive-shell pyenv bootstrap, direct `pyenv shell` confirmation output, and mistake reconciliation.
