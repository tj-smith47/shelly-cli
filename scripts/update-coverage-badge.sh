#!/usr/bin/env bash
# update-coverage-badge.sh - Generate and push a coverage badge to the badges branch.
# Expects coverage.out to exist in the current directory.
set -euo pipefail

# Parse coverage percentage from coverage.out
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

# Determine badge color based on thresholds
if (( $(echo "$COVERAGE >= 90" | bc -l) )); then COLOR="brightgreen"
elif (( $(echo "$COVERAGE >= 80" | bc -l) )); then COLOR="green"
elif (( $(echo "$COVERAGE >= 70" | bc -l) )); then COLOR="yellowgreen"
elif (( $(echo "$COVERAGE >= 60" | bc -l) )); then COLOR="yellow"
else COLOR="red"; fi

# Setup git for github-actions bot
git config user.email "github-actions[bot]@users.noreply.github.com"
git config user.name "github-actions[bot]"

# Fetch or create the badges branch
git fetch origin badges:badges 2>/dev/null || true
if git show-ref --verify --quiet refs/heads/badges; then
  git checkout badges
else
  git checkout --orphan badges
  git rm -rf . > /dev/null 2>&1 || true
fi

# Generate shields.io endpoint badge JSON
BADGE="{\"schemaVersion\":1,\"label\":\"coverage\",\"message\":\"${COVERAGE}%\",\"color\":\"${COLOR}\"}"
echo "$BADGE" > coverage.json

# Commit and force-push to badges branch
git add coverage.json
git diff --cached --quiet || git commit -m "Update coverage to ${COVERAGE}%"
git push origin badges --force
