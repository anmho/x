#!/usr/bin/env bash
# Create PRs for open branches (excluding main).
# Requires: gh auth login
# Usage: ./scripts/create-prs-for-branches.sh [--dry-run]

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DRY_RUN=0

[[ "${1:-}" == "--dry-run" ]] && DRY_RUN=1

cd "$ROOT_DIR"

# Branches that need PRs (pushed to origin, not main)
BRANCHES=()
for ref in $(git for-each-ref --format='%(refname:short)' refs/remotes/origin 2>/dev/null); do
  [[ "$ref" != origin/* ]] && continue
  branch="${ref#origin/}"
  [[ "$branch" == "main" || "$branch" == "HEAD" ]] && continue
  # Skip if PR already exists
  if gh pr list --head "$branch" --state all -q . 2>/dev/null | grep -q .; then
    echo "skip: $branch (PR exists)"
    continue
  fi
  BRANCHES+=("$branch")
done

if [[ ${#BRANCHES[@]} -eq 0 ]]; then
  echo "No branches need PRs."
  exit 0
fi

echo "Branches to create PRs for: ${BRANCHES[*]}"

for branch in "${BRANCHES[@]}"; do
  title=""
  body_file=""
  case "$branch" in
    fix/top-10-audit-subagent-changes)
      title="feat: DevEx build system and agentic improvements"
      body_file="$ROOT_DIR/.github/PULL_REQUEST_TEMPLATE.md"
      ;;
    *)
      title="$branch"
      body_file=""
      ;;
  esac

  if [[ "$DRY_RUN" -eq 1 ]]; then
    echo "would create: gh pr create --base main --head $branch --title \"$title\" ${body_file:+--body-file \"$body_file\"}"
    continue
  fi

  if [[ -n "$body_file" && -f "$body_file" ]]; then
    gh pr create --base main --head "$branch" --title "$title" --body-file "$body_file"
  else
    gh pr create --base main --head "$branch" --title "$title"
  fi
  echo "created PR for $branch"
done
