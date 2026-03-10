#!/usr/bin/env bash
# Create Linear tickets from all backlog payloads.
# Requires: LINEAR_API_KEY in environment
# Usage: LINEAR_API_KEY=xxx ./scripts/linear/create-all-backlog-tickets.sh

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

if [[ -z "${LINEAR_API_KEY:-}" ]]; then
  echo "Error: LINEAR_API_KEY is required"
  echo "Usage: LINEAR_API_KEY=xxx $0"
  exit 1
fi

PAYLOADS=(
  docs/backlog/session-tickets.json
  docs/backlog/ci-console-ux-tickets.json
  docs/backlog/x-platform-repo-hygiene-tickets.json
  docs/backlog/devex-completed-tickets.json
  docs/backlog/folder-by-feature-tickets.json
  docs/backlog/verify-nx-targets-tickets.json
  docs/backlog/local-github-workflow-tickets.json
)

for f in "${PAYLOADS[@]}"; do
  if [[ -f "$f" ]]; then
    echo "=== Creating from $f ==="
    node scripts/linear/create-issues.mjs --input "$f" --team-key ANM
    echo ""
  else
    echo "skip: $f (not found)"
  fi
done

echo "Done. Link PRs: ./scripts/linear/link-pr-to-linear.sh ANM-XXX <PR_NUMBER>"
