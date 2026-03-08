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
- Preventive rule/check added:
- Verification:
```

## Entries

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
