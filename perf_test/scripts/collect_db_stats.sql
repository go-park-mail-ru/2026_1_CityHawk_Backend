-- Save with:
-- psql "$DATABASE_URL" -f perf_test/scripts/collect_db_stats.sql > perf_test/results/db-stats.txt

\timing on

SELECT now() AS collected_at;

SELECT count(*) AS events_count FROM event;
SELECT count(*) AS event_sessions_count FROM event_session;
SELECT count(*) AS event_images_count FROM event_image;

SELECT
    datname,
    numbackends,
    xact_commit,
    xact_rollback,
    blks_read,
    blks_hit,
    tup_returned,
    tup_fetched,
    tup_inserted,
    tup_updated,
    tup_deleted,
    deadlocks
FROM pg_stat_database
WHERE datname = current_database();

SELECT
    query,
    calls,
    round(total_exec_time::numeric, 2) AS total_exec_time_ms,
    round(mean_exec_time::numeric, 2) AS mean_exec_time_ms,
    rows,
    shared_blks_hit,
    shared_blks_read,
    temp_blks_read,
    temp_blks_written
FROM pg_stat_statements
WHERE dbid = (SELECT oid FROM pg_database WHERE datname = current_database())
ORDER BY total_exec_time DESC
LIMIT 20;

SELECT
    schemaname,
    relname,
    indexrelname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE relname IN ('event', 'event_session', 'event_image', 'event_place', 'place', 'city')
ORDER BY relname, indexrelname;
