# HW 4: DB Performance Optimization

## Goal

This homework documents a reproducible load-testing and optimization cycle for
the CityHawk backend:

1. Load test the main write endpoint.
2. Load test the main read endpoint after the database contains 100k main
   entities.
3. Inspect PostgreSQL metrics and query plans.
4. Apply a focused optimization.
5. Repeat the read test and compare results.

The main business entity is `event`: it is the core object shown on the feed,
map, collections and event details pages.

## Tested API

Write endpoint:

```text
POST /api/events
```

Read endpoint:

```text
GET /api/events?limit=12&offset=0&sort=dateAsc
```

The read test also mixes representative variants of the same endpoint:

```text
GET /api/events?sort=popular
GET /api/events?query=Perf+Event+09
GET /api/events?cityId=10000000-0000-0000-0000-000000000001
GET /api/events?categoryId=20000000-0000-0000-0000-000000000001
GET /api/events?tag=30000000-0000-0000-0000-000000000001
```

## Tooling

Load generator: `vegeta`.

Database analysis:

- `pg_stat_statements`
- `EXPLAIN (ANALYZE, BUFFERS)`
- PostgreSQL slow logs and `auto_explain`
- Prometheus, Grafana and `postgres_exporter`

The project already enables the PostgreSQL observability pieces in
`db/postgres/postgresql.conf` and `docker-compose.yml`.

## Files

```text
perf_test/
  README.md                         this report
  init.sql                          baseline DDL before optimization migration 0030
  scripts/
    perf_seed.sql                   deterministic dictionaries for the test
    prepare_db.sh                   loads dictionaries
    auth.sh                         creates/logs in the test user and stores cookies
    generate_create_targets.py      generates vegeta JSON targets for POST /api/events
    generate_read_targets.py        generates vegeta JSON targets for GET /api/events
    run_create_test.sh              runs create load test
    run_read_test.sh                runs read load test
    collect_db_stats.sql            pg_stat_statements and DB counters
    explain_read_events.sql         representative read query plan
    reset_pg_stat_statements.sql    clears query stats between iterations
  results/                          raw vegeta and DB outputs
  reports/                          vegeta plots and screenshots
```

Optimization migration:

```text
db/migrations/0030_perf_event_indexes.up.sql
db/migrations/0030_perf_event_indexes.down.sql
```

## Environment

The test is intended to run on a dedicated VM, as required by the homework.

Record the actual VM parameters before running:

```text
Date:
CPU:
RAM:
Disk:
OS:
Docker:
PostgreSQL:
Go:
Vegeta:
```

Useful commands:

```bash
uname -a
docker version
docker compose version
vegeta -version
```

## Baseline Setup

Start the application stack on the VM:

```bash
docker compose up -d postgres photon cityhawk-auth-service cityhawk-profile-service cityhawk-events-service cityhawk-support-service cityhawk-social-service cityhawk-backend
```

Optional monitoring stack:

```bash
docker compose --profile monitoring up -d
```

Prepare deterministic dictionaries:

```bash
export DATABASE_URL='postgres://cityhawk_migrator:cityhawk_migrator@localhost:5432/cityhawk?sslmode=disable'
./perf_test/scripts/prepare_db.sh
```

Create/login the load-test user:

```bash
export BASE_URL='http://localhost:8080'
./perf_test/scripts/auth.sh
```

Reset query statistics before each iteration:

```bash
psql "$DATABASE_URL" -f perf_test/scripts/reset_pg_stat_statements.sql
```

## Iteration 1: Create 100k Events

Command:

```bash
COUNT=100000 RATE=200 RESULT_PREFIX=baseline-create ./perf_test/scripts/run_create_test.sh
```

Expected behavior:

- `POST /api/events` creates 100k `event` rows through the public API.
- Each event has one category, one tag, one image URL, one primary place and one
  session.
- Generated sessions are spread across 2048 places and non-overlapping time
  slots to satisfy the `event_session_place_time_excl` constraint.

Artifacts:

```text
perf_test/results/baseline-create.bin
perf_test/results/baseline-create.txt
perf_test/results/baseline-create.json
perf_test/reports/baseline-create.html
```

Result:

```text
Requests:
Duration:
RPS:
Success ratio:
p50:
p95:
p99:
Errors:
```

## Iteration 1: Read Baseline

Command:

```bash
COUNT=20000 RATE=500 RESULT_PREFIX=baseline-read ./perf_test/scripts/run_read_test.sh
```

Collect DB stats:

```bash
psql "$DATABASE_URL" -f perf_test/scripts/collect_db_stats.sql > perf_test/results/baseline-db-stats.txt
psql "$DATABASE_URL" -f perf_test/scripts/explain_read_events.sql > perf_test/results/baseline-explain-read-events.txt
```

Artifacts:

```text
perf_test/results/baseline-read.bin
perf_test/results/baseline-read.txt
perf_test/results/baseline-read.json
perf_test/reports/baseline-read.html
perf_test/results/baseline-db-stats.txt
perf_test/results/baseline-explain-read-events.txt
```

Result:

```text
Requests:
Duration:
RPS:
Success ratio:
p50:
p95:
p99:
Errors:
Top pg_stat_statements query:
Main EXPLAIN bottleneck:
```

## Bottleneck Analysis

The main read query is built in `internal/place/repository/postgres.go` by
`buildListEventsQuery`.

The expected hot areas after 100k events are:

- scanning/sorting `event_session` to find the next active session;
- `DISTINCT ON (event_id)` over `event_image` to find the first image;
- repeated `EXISTS` checks by `event_id` in `event_session`;
- sorting `event` by `created_at`, `title` or next session time;
- city filtering through `event_place`, `place` and `event_session`;
- trigram search over `lower(title)`.

The baseline plan should be inspected for:

- sequential scans over large tables;
- external sorts or large in-memory sorts;
- high shared block reads;
- nested loops that multiply work by event count;
- high `total_exec_time` in `pg_stat_statements`.

## Optimization

The optimization is isolated in:

```text
db/migrations/0030_perf_event_indexes.up.sql
```

It adds indexes for the observed hot paths:

```sql
idx_event_created_at_id
idx_event_title_id
idx_event_author_created_at
idx_event_session_event_end_start
idx_event_session_start_at_id
idx_event_session_end_at_event
idx_event_image_event_created_id
idx_event_place_place_event
idx_place_city_id
idx_city_lower_name
idx_event_lower_title_trgm
```

Apply it after saving baseline results:

```bash
psql "$DATABASE_URL" -f db/migrations/0030_perf_event_indexes.up.sql
psql "$DATABASE_URL" -c 'ANALYZE;'
psql "$DATABASE_URL" -f perf_test/scripts/reset_pg_stat_statements.sql
```

## Iteration 2: Read After Optimization

Command:

```bash
COUNT=20000 RATE=500 RESULT_PREFIX=optimized-read ./perf_test/scripts/run_read_test.sh
```

Collect DB stats:

```bash
psql "$DATABASE_URL" -f perf_test/scripts/collect_db_stats.sql > perf_test/results/optimized-db-stats.txt
psql "$DATABASE_URL" -f perf_test/scripts/explain_read_events.sql > perf_test/results/optimized-explain-read-events.txt
```

Artifacts:

```text
perf_test/results/optimized-read.bin
perf_test/results/optimized-read.txt
perf_test/results/optimized-read.json
perf_test/reports/optimized-read.html
perf_test/results/optimized-db-stats.txt
perf_test/results/optimized-explain-read-events.txt
```

Result:

```text
Requests:
Duration:
RPS:
Success ratio:
p50:
p95:
p99:
Errors:
Top pg_stat_statements query:
Main EXPLAIN improvement:
```

## Comparison

Fill this table after running the commands on the VM.

| Scenario | RPS | p50 | p95 | p99 | Success | Notes |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Create baseline, 100k events | TBD | TBD | TBD | TBD | TBD | Initial data population |
| Read baseline | TBD | TBD | TBD | TBD | TBD | Before migration 0030 |
| Read optimized | TBD | TBD | TBD | TBD | TBD | After migration 0030 |

Database comparison:

| Metric | Baseline | Optimized | Change |
| --- | ---: | ---: | ---: |
| Main query mean_exec_time | TBD | TBD | TBD |
| Main query total_exec_time | TBD | TBD | TBD |
| Shared blocks read | TBD | TBD | TBD |
| Shared blocks hit | TBD | TBD | TBD |
| Slow queries in PostgreSQL log | TBD | TBD | TBD |

## Expected Conclusion

The expected result is that read latency improves after adding indexes that match
the event list query shape. The largest improvement should appear in:

- lower p95/p99 latency for `GET /api/events`;
- lower mean execution time in `pg_stat_statements`;
- fewer sequential scans and sorts in `EXPLAIN`;
- increased usage of indexes on `event_session`, `event_image`, `event` and
  `place`.

Write performance can become slightly slower because additional indexes must be
maintained on insert. This is acceptable if the main product path is read-heavy,
which is true for CityHawk: users browse events much more often than they create
them.

## Reproducibility Checklist

Before submitting, make sure the repository contains:

- `perf_test/init.sql`
- all scripts under `perf_test/scripts`
- raw vegeta outputs in `perf_test/results`
- DB stats and explain outputs in `perf_test/results`
- HTML plots or screenshots in `perf_test/reports`
- filled comparison tables in this README
- the optimization migration `0030_perf_event_indexes`
