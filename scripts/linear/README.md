# Linear Integration

Create Linear issues from a local JSON payload.

## Files

- `scripts/linear/create-issues.mjs`: creates issues via Linear GraphQL API.
- `docs/backlog/linear-repo-analysis-tickets.json`: ticket payload generated from repository analysis.

## Usage

Dry run (no API key required):

```bash
node scripts/linear/create-issues.mjs --dry-run
```

Create issues in Linear (requires API key + team):

```bash
LINEAR_API_KEY=<your-api-key> LINEAR_TEAM_KEY=ENG node scripts/linear/create-issues.mjs
```

You can also pass team explicitly:

```bash
LINEAR_API_KEY=<your-api-key> node scripts/linear/create-issues.mjs --team-id <team-id>
```

Optional arguments:

- `--input <path>`: use a different payload file.
- `--dry-run`: print tickets without creating issues.
- `--team-id <id>`: target Linear team ID.
- `--team-key <key>`: target Linear team key (e.g., `ENG`).

## Route Linear Issues to Code Agents

Use the code-agent router to convert existing Linear issues into implementation briefs:

```bash
LINEAR_API_KEY=<your-api-key> LINEAR_TEAM_KEY=ENG node agents/code-agents/run.mjs --config agents/code-agents/config.example.json
```

Dry run against local sample payload:

```bash
node agents/code-agents/run.mjs --issues-json agents/code-agents/sample-issues.json --dry-run
```
