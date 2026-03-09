# Linear + PR Linkage Runbook

Per AGENTS.md: PRs must include a Linear backlink and be linked to the Linear issue.

**PR template:** `.github/PULL_REQUEST_TEMPLATE.md` — Linear at top, Summary, Context, Screenshots. Keep descriptions concise; avoid verbose Note/Overview sections.

## 1. Create Linear tickets (parent + subtickets)

```bash
LINEAR_API_KEY=lin_api_xxx node scripts/linear/create-issues.mjs \
  --input docs/backlog/devex-completed-tickets.json \
  --team-key ANM
```

The script creates the parent ticket first, then subtickets with `parentRef: "parent"` linked to it. Output includes the parent identifier (e.g. `ANM-123`).

## 2. Update PR with Linear backlink

Add to the PR description (top):

```
Linear: [ANM-XXX](https://linear.app/anmho/issue/ANM-XXX)
```

Replace `ANM-XXX` with the parent ticket identifier from step 1.

```bash
gh pr edit 1 --body "$(cat <<'BODY'
Linear: [ANM-XXX](https://linear.app/anmho/issue/ANM-XXX)

## Summary
...

## Context
...
BODY
)"
```

Or edit the PR in the GitHub UI and paste the Linear link at the top.

## 3. Link PR in Linear (optional)

If GitHub–Linear integration is configured, the PR may auto-link when the branch name or PR body references the issue. Otherwise, add the PR URL to the Linear issue description or comments.

## Payload format for parent/child tickets

In the JSON payload, the first ticket is the parent. Subtickets use `"parentRef": "parent"`:

```json
{
  "teamKey": "ANM",
  "tickets": [
    { "title": "Parent ticket", "description": "..." },
    { "title": "Child 1", "parentRef": "parent", "description": "..." },
    { "title": "Child 2", "parentRef": "parent", "description": "..." }
  ]
}
```
