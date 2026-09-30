#!/usr/bin/env bash
# One read-only org sweep: run-one.sh <base|fixed> <ecosystem:label>
# Needs UPDATECLI_GITHUB_TOKEN, or UPDATECLI_GITHUB_APP_CLIENT_ID, UPDATECLI_GITHUB_APP_INSTALLATION_ID
# and UPDATECLI_GITHUB_APP_PRIVATE_KEY_PATH (or UPDATECLI_GITHUB_APP_PRIVATE_KEY).
set -euo pipefail

VARIANT=$1
LABEL=$2
BENCH=$(cd "$(dirname "$0")" && pwd)
COMPOSE_DIR=${COMPOSE_DIR:?path to the checkout holding updatecli-compose.yaml}
BIN=$BENCH/bin/updatecli-$VARIANT
OUT=$BENCH/runs/${LABEL#ecosystem:}-$VARIANT
# Each run starts on a full budget, so neither variant inherits the other's rate-limit pauses.
MIN_BUDGET=${MIN_BUDGET:-7000}

: "${UPDATECLI_GITHUB_TOKEN:-${UPDATECLI_GITHUB_APP_CLIENT_ID:?set UPDATECLI_GITHUB_TOKEN or the UPDATECLI_GITHUB_APP_* variables}}"
# Only the credentials the budget is measured for may authenticate, and nothing may reach Udash.
unset GITHUB_TOKEN GH_TOKEN UPDATECLI_UDASH_API_URL UPDATECLI_UDASH_URL UPDATECLI_UDASH_ACCESS_TOKEN

rm -rf "$OUT"
mkdir -p "$OUT/tmp"
uv run -q --with 'pyjwt[crypto]' "$BENCH/ratelimit.py" wait "$MIN_BUDGET" >"$OUT/ratelimit-before.json"

export COLLECTOR_CONFIG="receivers:
  otlp:
    protocols:
      http:
        endpoint: 127.0.0.1:4318
exporters:
  file:
    path: $OUT/spans.jsonl
service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [file]"
otelcol-contrib --config=env:COLLECTOR_CONFIG >"$OUT/otelcol.log" 2>&1 &
COLLECTOR=$!
trap 'kill $COLLECTOR 2>/dev/null || true' EXIT
sleep 4

started=$(date -u +%FT%TZ)
start=$(date +%s)
set +e
(
  cd "$COMPOSE_DIR"
  # Private clone dir per run, so both variants start from nothing like a CI job.
  TMPDIR="$OUT/tmp" OTEL_TRACES_EXPORTER=otlphttp OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4318 \
    OTEL_BSP_MAX_QUEUE_SIZE=262144 \
    "$BIN" compose apply --push=false --commit=false --labels "$LABEL" 2>&1
) | python3 -u -c 'import sys, datetime
for line in sys.stdin:
    sys.stdout.write(datetime.datetime.now(datetime.UTC).isoformat() + " " + line)' >"$OUT/updatecli.log"
status=${PIPESTATUS[0]}
set -e
wall=$(($(date +%s) - start))

sleep 6
uv run -q --with 'pyjwt[crypto]' "$BENCH/ratelimit.py" show >"$OUT/ratelimit-after.json"
rm -rf "$OUT/tmp"
printf '{"variant":"%s","label":"%s","exit":%d,"wall_s":%d,"started":"%s"}\n' \
  "$VARIANT" "$LABEL" "$status" "$wall" "$started" >"$OUT/summary.json"
cat "$OUT/summary.json"
