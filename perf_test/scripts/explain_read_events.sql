-- Representative read endpoint query plan.
-- The SQL mirrors the heavy GET /api/events code path closely enough to compare plans
-- before and after db/migrations/0030_perf_event_indexes.up.sql.

EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
WITH first_image AS (
    SELECT DISTINCT ON (ei.event_id)
        ei.event_id,
        ei.image_url
    FROM event_image ei
    ORDER BY ei.event_id, ei.created_at ASC, ei.id ASC
),
ranked_next_session AS (
    SELECT
        es.event_id,
        es.start_at,
        p.id::text AS place_id,
        p.name AS place_name,
        p.address_line,
        p.latitude::float8,
        p.longitude::float8,
        c.id::text AS city_id,
        c.name AS city_name,
        c.country_name,
        c.timezone,
        ROW_NUMBER() OVER (PARTITION BY es.event_id ORDER BY es.start_at ASC, es.id ASC) AS rn
    FROM event_session es
    JOIN place p ON p.id = es.place_id
    JOIN city c ON c.id = p.city_id
    WHERE es.end_at >= now()
),
next_session AS (
    SELECT event_id, start_at, place_id, place_name, address_line, latitude, longitude, city_id, city_name, country_name, timezone
    FROM ranked_next_session
    WHERE rn = 1
),
tags AS (
    SELECT
        et.event_id,
        ARRAY_AGG(t.id::text ORDER BY t.name) AS tag_ids,
        ARRAY_AGG(t.name ORDER BY t.name) AS tag_names
    FROM event_tag et
    JOIN tag t ON t.id = et.tag_id
    GROUP BY et.event_id
),
favorite_counts AS (
    SELECT
        fe.event_id,
        COUNT(*)::int AS favorite_count
    FROM favorite_event fe
    GROUP BY fe.event_id
),
filtered_events AS (
    SELECT e.id
    FROM event e
    WHERE (
        NOT EXISTS (
            SELECT 1
            FROM event_session es_active
            WHERE es_active.event_id = e.id
        )
        OR EXISTS (
            SELECT 1
            FROM event_session es_active
            WHERE es_active.event_id = e.id
              AND es_active.end_at >= now()
        )
    )
    GROUP BY e.id
),
base AS (
    SELECT
        e.id,
        e.title,
        e.location_description,
        COALESCE(fi.image_url, '') AS cover_image_url,
        COALESCE(tags.tag_ids, '{}'::text[]) AS tag_ids,
        COALESCE(tags.tag_names, '{}'::text[]) AS tag_names,
        ns.start_at,
        COUNT(*) OVER()::int AS total_count,
        COALESCE(fc.favorite_count, 0)::int AS popularity
    FROM filtered_events filtered
    JOIN event e ON e.id = filtered.id
    LEFT JOIN first_image fi ON fi.event_id = e.id
    LEFT JOIN next_session ns ON ns.event_id = e.id
    LEFT JOIN tags ON tags.event_id = e.id
    LEFT JOIN favorite_counts fc ON fc.event_id = e.id
    ORDER BY ns.start_at ASC NULLS LAST, e.id ASC
    LIMIT 12 OFFSET 0
)
SELECT *
FROM base;
