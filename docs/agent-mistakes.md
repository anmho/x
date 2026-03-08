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
