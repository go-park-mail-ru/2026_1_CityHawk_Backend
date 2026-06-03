#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BASE_URL="${BASE_URL:-http://localhost:8080}"
COUNT="${COUNT:-100000}"
RATE="${RATE:-200}"
DURATION_SECONDS="${DURATION_SECONDS:-$((COUNT / RATE))}"
RESULT_PREFIX="${RESULT_PREFIX:-baseline-create}"
AUTH_ENV="${AUTH_ENV:-$ROOT_DIR/perf_test/results/auth.env}"
TARGETS="$ROOT_DIR/perf_test/results/${RESULT_PREFIX}-targets.jsonl"
BIN="$ROOT_DIR/perf_test/results/${RESULT_PREFIX}.bin"
TXT="$ROOT_DIR/perf_test/results/${RESULT_PREFIX}.txt"
JSON_REPORT="$ROOT_DIR/perf_test/results/${RESULT_PREFIX}.json"

source "$AUTH_ENV"

python3 "$ROOT_DIR/perf_test/scripts/generate_create_targets.py" \
  --base-url "$BASE_URL" \
  --count "$COUNT" \
  --csrf-token "$CSRF_TOKEN" \
  --cookie-header "$COOKIE_HEADER" \
  > "$TARGETS"

vegeta attack \
  -format=json \
  -targets="$TARGETS" \
  -rate="$RATE" \
  -duration="${DURATION_SECONDS}s" \
  -name="$RESULT_PREFIX" \
  | tee "$BIN" \
  | vegeta report > "$TXT"

vegeta report -type=json "$BIN" > "$JSON_REPORT"
vegeta plot "$BIN" > "$ROOT_DIR/perf_test/reports/${RESULT_PREFIX}.html"

cat "$TXT"
