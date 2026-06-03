#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DATABASE_URL="${DATABASE_URL:-postgres://cityhawk_migrator:cityhawk_migrator@localhost:5432/cityhawk?sslmode=disable}"

psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$ROOT_DIR/perf_test/scripts/perf_seed.sql"
