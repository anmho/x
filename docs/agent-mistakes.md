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

## 2026-03-08T22:14:08Z - Committed in wrong repository
- What happened: I tried to commit frontend route changes from the top-level repository, but the files live in a nested Git repository at `services/omnichannel/frontend`.
- Root cause: I did not verify repository boundaries before staging and committing.
- Preventive rule/check added: Before every commit, run `git rev-parse --show-toplevel` in the working directory that contains the edited files.
- Verification: Confirmed separate top-level paths for `/Users/andrewho/repos/projects/x` and `/Users/andrewho/repos/projects/x/services/omnichannel/frontend`.

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
