# Linear Backlog Completion Runbook

When `LINEAR_API_KEY` is available, create tickets from backlog payloads and work through them until the backlog is empty.

## 1. Create Tickets from Payloads

Batch create all tickets:

```bash
LINEAR_API_KEY=xxx ./scripts/linear/create-all-backlog-tickets.sh
```

Or create individually:

```bash
LINEAR_API_KEY=xxx node scripts/linear/create-issues.mjs --input docs/backlog/<file>.json --team-key ANM
```

Payloads (in priority order):

| File | Tickets | Focus |
|------|---------|-------|
| `session-tickets.json` | 7 | CI fixes, .nvmrc, go-version |
| `ci-console-ux-tickets.json` | 5 | CI/CD, Cloud Console UX |
| `x-platform-repo-hygiene-tickets.json` | 6 | .gitignore, cleanup, workspaces |
| `devex-completed-tickets.json` | 6 | DevEx build system |
| `folder-by-feature-tickets.json` | 1 | Repo restructure |
| `verify-nx-targets-tickets.json` | 1 | Nx migration |
| `local-github-workflow-tickets.json` | 1 | Run CI locally |

## 2. Link PRs to Tickets

After creating tickets:

```bash
./scripts/linear/link-pr-to-linear.sh ANM-XXX <PR_NUMBER>
```

Get PR numbers: `gh pr list --json number,headRefName -q '.[] | "\(.number) \(.headRefName)"'`

## 3. Work Loop

1. Fetch backlog from Linear (or use the backlog JSON as source of truth).
2. Pick highest-priority ticket; set to In Progress.
3. Implement; validate; commit.
4. Mark ticket Done; link PR if applicable.
5. Repeat until backlog empty.

## 4. Completed in This Session (No Linear API)

- Removed `stack.sh` references from docs (script doesn't exist).
- Updated docs: platform-cli-workflow, repository-architecture-deep-dive, AGENTS.md, services/omnichannel/README.
- Consolidated omnichannel frontend into cloud-console (earlier session).
- Removed omnichannel-frontend from workspaces and build config.

## 5. Without LINEAR_API_KEY

The agent cannot create or update Linear tickets. Document manual steps in the active ExecPlan; run the create-issues commands when the key is available.
