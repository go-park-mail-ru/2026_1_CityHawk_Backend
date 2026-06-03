#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BASE_URL="${BASE_URL:-http://localhost:8080}"
EMAIL="${PERF_EMAIL:-perf-author@cityhawk.local}"
PASSWORD="${PERF_PASSWORD:-PerfPassword123}"
USERNAME="${PERF_USERNAME:-perf_author}"
COOKIE_JAR="${COOKIE_JAR:-$ROOT_DIR/perf_test/results/perf_cookies.txt}"
AUTH_ENV="${AUTH_ENV:-$ROOT_DIR/perf_test/results/auth.env}"

mkdir -p "$(dirname "$COOKIE_JAR")"

register_status="$(
  curl -sS -o /tmp/cityhawk-perf-register.json -w "%{http_code}" \
    -c "$COOKIE_JAR" \
    -H "Content-Type: application/json" \
    -X POST "$BASE_URL/api/auth/register" \
    -d "{\"email\":\"$EMAIL\",\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}"
)"

if [[ "$register_status" != "201" && "$register_status" != "409" ]]; then
  echo "register failed with HTTP $register_status"
  cat /tmp/cityhawk-perf-register.json
  exit 1
fi

login_status="$(
  curl -sS -o /tmp/cityhawk-perf-login.json -w "%{http_code}" \
    -c "$COOKIE_JAR" \
    -H "Content-Type: application/json" \
    -X POST "$BASE_URL/api/auth/login" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}"
)"

if [[ "$login_status" != "200" ]]; then
  echo "login failed with HTTP $login_status"
  cat /tmp/cityhawk-perf-login.json
  exit 1
fi

csrf_token="$(awk '$6 == "csrf_token" { value = $7 } END { print value }' "$COOKIE_JAR")"
access_token="$(awk '$6 == "access_token" { value = $7 } END { print value }' "$COOKIE_JAR")"

if [[ -z "$csrf_token" || -z "$access_token" ]]; then
  echo "failed to extract access_token/csrf_token from $COOKIE_JAR"
  exit 1
fi

cat > "$AUTH_ENV" <<EOF
export BASE_URL="$BASE_URL"
export COOKIE_JAR="$COOKIE_JAR"
export CSRF_TOKEN="$csrf_token"
export COOKIE_HEADER="access_token=$access_token; csrf_token=$csrf_token"
EOF

echo "auth data written to $AUTH_ENV"
