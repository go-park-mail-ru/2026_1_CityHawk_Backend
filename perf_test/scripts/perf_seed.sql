-- Deterministic reference data for load tests.
-- It creates one city, reusable categories/tags and many places.
-- Event rows themselves are created through POST /api/events.

INSERT INTO city (id, name, country_name, timezone)
VALUES (
    '10000000-0000-0000-0000-000000000001',
    'Perf City',
    'Perf Country',
    'Europe/Moscow'
)
ON CONFLICT (country_name, name)
DO UPDATE SET timezone = EXCLUDED.timezone, updated_at = now();

INSERT INTO category (id, name)
VALUES
    ('20000000-0000-0000-0000-000000000001', 'Perf Concerts'),
    ('20000000-0000-0000-0000-000000000002', 'Perf Movies'),
    ('20000000-0000-0000-0000-000000000003', 'Perf Exhibitions'),
    ('20000000-0000-0000-0000-000000000004', 'Perf Sport')
ON CONFLICT (name)
DO UPDATE SET updated_at = now();

INSERT INTO tag (id, name, tag_group)
VALUES
    ('30000000-0000-0000-0000-000000000001', 'Perf Live', 'format'),
    ('30000000-0000-0000-0000-000000000002', 'Perf Outdoor', 'format'),
    ('30000000-0000-0000-0000-000000000003', 'Perf Family', 'audience'),
    ('30000000-0000-0000-0000-000000000004', 'Perf Evening', 'time')
ON CONFLICT (name)
DO UPDATE SET tag_group = EXCLUDED.tag_group, updated_at = now();

INSERT INTO place (id, city_id, name, address_line, latitude, longitude, description)
SELECT
    ('40000000-0000-0000-0000-' || lpad(i::text, 12, '0'))::uuid,
    '10000000-0000-0000-0000-000000000001'::uuid,
    'Perf Place ' || i,
    'Perf Street ' || i,
    55.000000 + ((i % 900)::numeric / 10000),
    37.000000 + ((i % 900)::numeric / 10000),
    'Synthetic place for HW 4 performance tests'
FROM generate_series(1, 2048) AS s(i)
ON CONFLICT DO NOTHING;

ANALYZE city;
ANALYZE category;
ANALYZE tag;
ANALYZE place;
