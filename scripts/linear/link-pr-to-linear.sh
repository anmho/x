#!/usr/bin/env bash
# Update PR description with Linear backlink.
# Usage: ./scripts/linear/link-pr-to-linear.sh ANM-123 [PR_NUMBER]
# If PR_NUMBER is omitted, uses the PR for the current branch.

set -e

IDENTIFIER="${1:?Usage: $0 ANM-XXX [PR_NUMBER]}"
PR_NUM="${2:-}"

if [[ -z "$PR_NUM" ]]; then
  BRANCH=$(git branch --show-current)
  PR_NUM=$(gh pr list --head "$BRANCH" --json number -q '.[0].number')
  if [[ -z "$PR_NUM" ]]; then
    echo "No PR found for branch $BRANCH"
    exit 1
  fi
fi

LINEAR_LINK="Linear: [${IDENTIFIER}](https://linear.app/anmho/issue/${IDENTIFIER})"
CURRENT_BODY=$(gh pr view "$PR_NUM" --json body -q '.body')

# Prepend Linear link if not already present
if [[ "$CURRENT_BODY" == *"$IDENTIFIER"* ]]; then
  echo "PR #$PR_NUM already references $IDENTIFIER"
  exit 0
fi

NEW_BODY="${LINEAR_LINK}

${CURRENT_BODY}"
gh pr edit "$PR_NUM" --body "$NEW_BODY"
echo "Updated PR #$PR_NUM with $IDENTIFIER"
