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
GET /api/events?sort=titleAsc
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
Date: 2026-06-03
Host: 2026-1-cityhawk VM
CPU: fill from VM
RAM: fill from VM
Disk: fill from VM
OS: Ubuntu
Docker: fill from `docker version`
PostgreSQL: 16
Go: fill from `go version`
Vegeta: fill from `vegeta -version`
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
Main create run, COUNT=100000 RATE=200:

Requests      [total, rate, throughput]  100000, 200.00, 181.03
Duration      [total, attack, wait]      8m23.387213159s, 8m19.995433871s, 3.391779288s
Latencies     [mean, 50, 95, 99, max]    1.877587155s, 1.837538605s, 4.620437972s, 5.203054522s, 5.423837861s
Bytes In      [total, mean]              4546682, 45.47
Bytes Out     [total, mean]              59994228, 599.94
Success       [ratio]                    91.13%
Status Codes  [code:count]               0:2  201:91127  500:8871
Error Set:
500 Internal Server Error
Post "http://localhost:8080/api/events": EOF

Fill run, COUNT=2 RATE=1:

Requests      [total, rate, throughput]  2, 2.00, 1.98
Duration      [total, attack, wait]      1.007600486s, 1.000012235s, 7.588251ms
Latencies     [mean, 50, 95, 99, max]    7.683488ms, 7.683488ms, 7.778725ms, 7.778725ms, 7.778725ms
Bytes In      [total, mean]              92, 46.00
Bytes Out     [total, mean]              1200, 600.00
Success       [ratio]                    100.00%
Status Codes  [code:count]               201:2
Error Set:

Final DB check:

SELECT count(*) FROM event WHERE title LIKE 'Perf Event %%';
count = 100000
```

The write endpoint reached the VM's practical limit at `RATE=200`: p99 exceeded
five seconds, and the service returned `8871` HTTP 500 responses. The missing
events were loaded by a low-rate fill run. The homework requirement to create
100k main entities was satisfied by the final DB count.

## Iteration 1: Read Baseline

Command:

```bash
COUNT=5000 RATE=20 RESULT_PREFIX=baseline-read ./perf_test/scripts/run_read_test.sh
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
Requests      [total, rate, throughput]  5000, 20.00, 0.00
Duration      [total, attack, wait]      4m39.950655221s, 4m9.95022971s, 30.000425511s
Latencies     [mean, 50, 95, 99, max]    29.963377006s, 30.000627419s, 30.001316485s, 30.001713638s, 30.002338935s
Bytes In      [total, mean]              80, 0.02
Bytes Out     [total, mean]              0, 0.00
Success       [ratio]                    0.00%
Status Codes  [code:count]               0:4998  500:2
Error Set:
500 Internal Server Error
context deadline exceeded / EOF for GET /api/events variants
```

At 100k generated events, the unoptimized read endpoint did not produce stable
responses even at `RATE=20`. Vegeta timed out after 30 seconds. This became the
main bottleneck for the optimization cycle.

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

The first optimization is isolated in:

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

If the read endpoint still times out after the index-only optimization, apply the
second optimization from `internal/place/repository/postgres.go`: the event list
query now selects the paginated `event.id` set first in `page_events`, and only
then fetches images, tags, nearest sessions, places and favorite counts for that
small page. This avoids aggregating and sorting large related tables for all
100k events before `LIMIT`.

The index-only optimization was applied first and tested as `optimized-read`.
It did not solve the endpoint timeout at the selected load, so the second
optimization changed the query shape to a pagination-first approach and was
tested as `optimized-read-v2`.

## Iteration 2: Read After Optimization

Command:

```bash
COUNT=5000 RATE=20 RESULT_PREFIX=optimized-read ./perf_test/scripts/run_read_test.sh
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
Index-only optimization, COUNT=5000 RATE=20:

Requests      [total, rate, throughput]  5000, 20.00, 0.00
Duration      [total, attack, wait]      4m39.95140021s, 4m9.950580291s, 30.000819919s
Latencies     [mean, 50, 95, 99, max]    29.935932865s, 30.000629951s, 30.001310652s, 30.001711856s, 30.002416664s
Bytes In      [total, mean]              240, 0.05
Bytes Out     [total, mean]              0, 0.00
Success       [ratio]                    0.00%
Status Codes  [code:count]               0:4994  500:6
Error Set:
500 Internal Server Error
context deadline exceeded / EOF for GET /api/events variants

Pagination-first query rewrite, COUNT=5000 RATE=20:

Requests      [total, rate, throughput]  5000, 20.00, 0.01
Duration      [total, attack, wait]      4m39.950555106s, 4m9.95023258s, 30.000322526s
Latencies     [mean, 50, 95, 99, max]    29.966801045s, 30.000631888s, 30.001296391s, 30.001732848s, 30.002455233s
Bytes In      [total, mean]              23618, 4.72
Bytes Out     [total, mean]              0, 0.00
Success       [ratio]                    0.04%
Status Codes  [code:count]               0:4998  200:2
Error Set:
context deadline exceeded / EOF for GET /api/events variants
```

The index-only optimization made the new indexes visible in PostgreSQL stats,
especially on `event_session`, `event_image`, `event_place` and `place`, but it
was insufficient for the selected read workload. The pagination-first rewrite
allowed a small number of successful responses, but the endpoint was still not
stable under `RATE=20`. The conclusion is that the bottleneck is not only index
coverage but also the endpoint/query shape and current service/DB capacity.

## Comparison

| Scenario | RPS | p50 | p95 | p99 | Success | Notes |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Create baseline, 100k events | 181.03 throughput | 1.84s | 4.62s | 5.20s | 91.13% main run; final count 100000 | `RATE=200` overloaded write path; fill run completed missing rows |
| Create fill | 1.98 throughput | 7.68ms | 7.78ms | 7.78ms | 100.00% | Low-rate fill for missing events |
| Read baseline | 0.00 throughput | 30.00s | 30.00s | 30.00s | 0.00% | Before migration 0030 |
| Read optimized, indexes | 0.00 throughput | 30.00s | 30.00s | 30.00s | 0.00% | Migration 0030 was insufficient |
| Read optimized, query rewrite | 0.01 throughput | 30.00s | 30.00s | 30.00s | 0.04% | Pagination-first query shape; still saturated |

Database comparison:

| Metric | Baseline | Optimized indexes | Change |
| --- | ---: | ---: | ---: |
| Top pg_stat_statements total_exec_time | 33233.08ms | 29920.14ms | -3312.94ms |
| Top pg_stat_statements mean_exec_time | 6.65ms | 5.98ms | -0.67ms |
| DB shared blocks read | 9952 | 15057 | +5105 |
| DB shared blocks hit | 42643996 | 74667006 | +32023010 |
| Deadlocks | 0 | 0 | no change |

## Conclusion

The workload successfully created 100k main entities through the public API.
The write path at `RATE=200` was close to the VM's limit: p95 was `4.62s`, p99
was `5.20s`, and the service returned errors. A low-rate fill completed the data
set.

The read path became the main bottleneck. With 100k generated events,
`GET /api/events` timed out at 30 seconds even at `RATE=20`. Indexes from
migration `0030` were used by PostgreSQL, but the index-only optimization did
not restore stable read throughput. A second optimization changed the query to
fetch the paginated event page first and then load related data for that page,
but the endpoint still remained saturated on the VM.

The practical conclusion is that CityHawk needs further read-side work before
this endpoint can serve 100k events under concurrent load: narrower endpoint
queries, removing expensive total counts from hot paths, caching/precomputed
event cards, or separate read models/materialized views for event listings.

## Reproducibility Checklist

Before submitting, make sure the repository contains:

- `perf_test/init.sql`
- all scripts under `perf_test/scripts`
- raw vegeta outputs in `perf_test/results`
- DB stats and explain outputs in `perf_test/results`
- HTML plots or screenshots in `perf_test/reports`
- filled comparison tables in this README
- the optimization migration `0030_perf_event_indexes`
