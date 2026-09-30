#!/usr/bin/env bash
# Baseline then fixed for each org-wide sweep, one at a time: they share one GraphQL budget.
set -uo pipefail
BENCH=$(cd "$(dirname "$0")" && pwd)
for label in ecosystem:docker ecosystem:packages ecosystem:githubactions; do
  for variant in base fixed; do
    echo "$(date -u +%T) starting $variant $label"
    "$BENCH/run-one.sh" "$variant" "$label" || echo "$variant $label exited non-zero, continuing"
  done
done
python3 "$BENCH/analyze.py"
