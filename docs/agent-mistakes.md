# Agent Mistake Log

This file is the permanent mistake memory for repository agents.

## Rules

1. Add a new entry immediately when a mistake is discovered.
2. Do not rewrite history; append corrections as new entries.
3. Every entry must include a concrete prevention step and verification.
4. Use UTC timestamps in ISO-8601 format.

## Entry Template

```text
## <YYYY-MM-DDTHH:MM:SSZ> - <short title>
- What happened:
- Root cause:
- How to avoid it:
- What to do instead:
- Verification:
```

## Entries

## 2026-03-09T02:00:00Z - Used scripts/verify instead of Nx targets for app checks
- What happened: CI and scripts/verify apps used custom shell logic (nx run sdk:generate-es + npm run build) instead of Nx targets.
- Root cause: Did not treat Nx as the single source of truth for build/verify orchestration.
- How to avoid it: Prefer Nx targets for all build, test, and verify steps. Use nx run-many -t <target> instead of ad-hoc scripts.
- What to do instead: Define verify/build as Nx targets in project.json; CI runs nx run-many directly. scripts/verify becomes a thin nx wrapper or is removed.
- Verification: Migrated verify_apps to nx run-many -t build; CI apps job now runs nx directly. Ticket: docs/backlog/verify-nx-targets-tickets.json for full migration.

## 2026-03-09T02:00:00Z - Did not push and check workflow logs after changes
- What happened: Made multiple changes (CI fix, AGENTS.md, etc.) but did not push until the user asked. Did not run gh run list / gh run view --log-failed to verify CI.
- Root cause: Focused on edits without completing the loop: commit → push → verify.
- How to avoid it: After every logical change, commit, push, and check workflow status. Per AGENTS.md Ralph loop: complete the full loop.
- What to do instead: Immediately after committing: git push; gh run list --limit 3; if failure, gh run view <id> --log-failed. Do not stop with uncommitted/unpushed work.
- Verification: AGENTS.md Ralph Loop Friendly section; this entry.

## 2026-03-08T22:38:06Z - Used project-style ticket prefix instead of team-key issue identifiers
- What happened: I initially tracked execution items as `XPLAT-*` in the ExecPlan and did not publish corresponding Linear issues before execution.
- Root cause: I treated the project label as the issue identifier prefix and deferred external ticket creation.
- Preventive rule/check added: Before editing plan ticket IDs, confirm team key vs project naming and run `node scripts/linear/create-issues.mjs` (dry-run first, then publish) at task start.
- Verification: Updated the plan tracker to `ANM`-scoped placeholders, prepared `docs/backlog/x-platform-repo-hygiene-tickets.json`, and executed the Linear dry-run command with `--team-key ANM`.

## 2026-03-08T22:29:00Z - Used compile build checks instead of platform stack command for runtime validation
- What happened: I validated frontend changes mainly with `npm run build` before checking stack state via `./platform status`.
- Root cause: I optimized for fast compile feedback and skipped the repository-default runtime validation command in initial checks.
- Preventive rule/check added: For runtime/environment claims, always run and report `./platform status` (or an explicitly scoped stack command) before finalizing.
- Verification: Added `Platform CLI First Workflow (Required)` to `AGENTS.md` and created `docs/runbooks/platform-cli-workflow.md`.

## 2026-03-08T22:14:08Z - Committed in wrong repository
- What happened: I tried to commit frontend route changes from the top-level repository, but the files live in a nested Git repository at `services/omnichannel/frontend`.
- Root cause: I did not verify repository boundaries before staging and committing.
- Preventive rule/check added: Before every commit, run `git rev-parse --show-toplevel` in the working directory that contains the edited files.
- Verification: Confirmed separate top-level paths for `/Users/andrewho/repos/projects/x` and `/Users/andrewho/repos/projects/x/services/omnichannel/frontend`.

## 2026-03-08T22:27:59Z - Unquoted bracket path in git add
- What happened: I ran `git add` with an unquoted path containing `[domain]`, and zsh expanded it as a glob, causing `no matches found`.
- Root cause: I forgot shell globbing behavior for dynamic-route paths.
- Preventive rule/check added: Always quote any path containing `[` or `]` when running shell commands.
- Verification: Retried staging using quoted dynamic-route paths.

## 2026-03-08T22:30:08Z - Invalid background smoke-check command
- What happened: I attempted a one-liner background startup/smoke check for `control-plane serve` that failed (`nice(5) failed`) and did not produce a valid process id for cleanup.
- Root cause: I used a brittle shell one-liner instead of a deterministic start/poll/stop sequence.
- Preventive rule/check added: Use explicit command steps (`start`, `poll`, `stop`) and verify PID capture before issuing `kill`.
- Verification: Subsequent validation relies on deterministic compile/test checks; runtime smoke is marked as blocked by sandbox.

## 2026-03-08T22:15:06Z - Missed one dark shell class during bulk replacement
- What happened: My initial automated replacement left `services/omnichannel/frontend/app/domains/page.tsx` still using `bg-[#0a0a0a]`.
- Root cause: I validated only a subset of grep output and did not run a final zero-match check for the hardcoded color pattern.
- Preventive rule/check added: After any bulk class migration, run a repository-wide `rg` for the old pattern and require zero matches before moving on.
- Verification: Re-ran `rg -n "bg-\\[#0a0a0a\\]" apps/cloud-console/app services/omnichannel/frontend/app`; zero results returned.

## 2026-03-08T22:25:53Z - Shell interpreted literal `<plan>` during validation
- What happened: I ran `rg -n "For every task, ensure an ExecPlan exists at `plans/<plan>.md`" AGENTS.md`, and zsh interpreted `` `...` `` plus `<plan>` as command/input redirection, producing `no such file or directory: plan`.
- Root cause: I used backticks inside double quotes instead of shell-safe single quoting for a literal Markdown snippet.
- Preventive rule/check added: When searching for literal strings that include backticks or `<...>`, wrap the entire pattern in single quotes.
- Verification: Re-ran validation with `rg -n 'For every task, ensure an ExecPlan exists at `plans/<plan>.md`' AGENTS.md` and received the expected single match.

## 2026-03-08T22:28:07Z - Ran TypeScript check without selecting project
- What happened: I executed `npm --prefix services/omnichannel/frontend exec tsc --noEmit` from the repo root and got the TypeScript CLI help output instead of a project typecheck.
- Root cause: I assumed `npm --prefix ... exec tsc` would always resolve the project context without explicitly running in the package directory.
- Preventive rule/check added: For frontend typechecks, run `./node_modules/.bin/tsc --noEmit` from the package directory (or provide `-p <tsconfig>` explicitly).
- Verification: Re-ran from `services/omnichannel/frontend` with `./node_modules/.bin/tsc --noEmit`; command exited successfully.

## 2026-03-08T22:28:48Z - Repeated shell quoting error while searching for backtick text
- What happened: I ran an `rg` pattern in double quotes that contained `` `ANM-28` ``, and zsh tried to execute `ANM-28` as a command.
- Root cause: I reused unsafe double-quoted search syntax despite already documenting this shell behavior.
- Preventive rule/check added: Treat all `rg` patterns containing backticks as single-quoted literals, no exceptions.
- Verification: Subsequent command output confirmed the shell error source and this entry was added immediately.

## 2026-03-08T22:37:52Z - Repeated backtick pattern quoting failure in rg validation
- What happened: I ran `rg` validations for updated `AGENTS.md` rules using double-quoted patterns with backticks, and zsh attempted to execute `Anmho`, `Backlog`, and `In` as commands.
- Root cause: I failed to apply the existing single-quote rule for literal Markdown text containing backticks.
- Preventive rule/check added: Before running any `rg` command, if the pattern contains backticks, require single-quoted pattern syntax and avoid shell interpolation.
- Verification: Re-ran validations using single-quoted patterns and obtained expected matches without shell command-substitution errors.
