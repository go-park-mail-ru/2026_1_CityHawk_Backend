BEGIN;

INSERT INTO city (name, country_name, timezone)
VALUES
    ('Москва', 'Россия', 'Europe/Moscow')
ON CONFLICT (country_name, name) DO NOTHING;

INSERT INTO user_account (email, username, user_surname, password_hash, city_id, avatar_url)
SELECT
    'seed.author@cityhawk.local',
    'seed_author',
    'CityHawk',
    '$2a$10$wJv1PLbF6XJz5lG1fV64VeGQFQf5d3M3K2Y6DzG6Q8rB3Yj8wYF1W',
    c.id,
    'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?auto=format&fit=crop&w=400&q=80'
FROM city c
WHERE c.country_name = 'Россия' AND c.name = 'Москва'
ON CONFLICT (email) DO NOTHING;

INSERT INTO category (name)
VALUES
    ('Парк'),
    ('Выставка'),
    ('Семья'),
    ('Шоу'),
    ('Стендап'),
    ('Театр'),
    ('Музей'),
    ('Музыка'),
    ('Квест'),
    ('Фото'),
    ('Лекция'),
    ('Фестиваль')
ON CONFLICT (name) DO NOTHING;

INSERT INTO tag (name)
VALUES
    ('Иммерсивное'),
    ('Семейное'),
    ('Ледовое шоу'),
    ('Стендап'),
    ('Балет'),
    ('Искусство'),
    ('Рок'),
    ('Ретро'),
    ('Квест'),
    ('На воздухе'),
    ('Ночное'),
    ('Образовательное')
ON CONFLICT (name) DO NOTHING;

INSERT INTO place (city_id, name, address_line, latitude, longitude, description)
SELECT
    c.id,
    v.name,
    v.address_line,
    v.latitude,
    v.longitude,
    v.description
FROM city c
CROSS JOIN (
    VALUES
        ('ВДНХ', 'Москва, проспект Мира, 119', 55.829834::numeric, 37.633096::numeric, 'Крупный выставочный и прогулочный комплекс в Москве.'),
        ('Navka Arena', 'Москва, улица Автозаводская, 23А', 55.704987::numeric, 37.644521::numeric, 'Современная площадка для шоу и спортивно-развлекательных событий.'),
        ('Live Арена', 'Московская область, Новоивановское, Западная улица, 145', 55.728246::numeric, 37.378809::numeric, 'Большая концертная площадка для массовых мероприятий.'),
        ('Большой театр', 'Москва, Театральная площадь, 1', 55.760186::numeric, 37.618711::numeric, 'Историческая театральная сцена в центре Москвы.'),
        ('Третьяковская галерея', 'Москва, Лаврушинский переулок, 10', 55.741389::numeric, 37.620556::numeric, 'Одна из ключевых музейных площадок столицы.'),
        ('Атмосфера', 'Москва, Шмитовский проезд, 32А', 55.751244::numeric, 37.618423::numeric, 'Концертная площадка с вечерними музыкальными программами.'),
        ('Лужники', 'Москва, улица Лужники, 24', 55.715765::numeric, 37.553322::numeric, 'Крупный спортивно-концертный кластер.'),
        ('Квест на Пруд-Ключики', 'Москва, улица Пруд-Ключики, 5', 55.747161::numeric, 37.736518::numeric, 'Локация для иммерсивных и командных квестов.'),
        ('Дизайн завод', 'Москва, улица Большая Новодмитровская, 36', 55.804441::numeric, 37.585773::numeric, 'Креативное пространство для фестивалей, лекций и выставок.'),
        ('Сад Эрмитаж', 'Москва, улица Каретный Ряд, 3', 55.770779::numeric, 37.608984::numeric, 'Городской сад для семейных и открытых мероприятий.')
) AS v(name, address_line, latitude, longitude, description)
WHERE c.country_name = 'Россия'
  AND c.name = 'Москва'
  AND NOT EXISTS (
      SELECT 1
      FROM place p
      WHERE p.name = v.name AND p.address_line = v.address_line
  );

WITH seed_author AS (
    SELECT u.id
    FROM user_account u
    WHERE u.email = 'seed.author@cityhawk.local'
),
generated_events AS (
    SELECT
        gs AS n,
        format('Seed Event %s', lpad(gs::text, 3, '0')) AS title,
        format(
            'Подборка событий CityHawk: карточка %s для наполнения каталога, тестов фильтрации и проверки списков.',
            lpad(gs::text, 3, '0')
        ) AS location_description,
        format(
            'Сидовое событие номер %s. Создано для наполнения базы 100 полноценными карточками с категориями, тегами, картинкой, сессией, коллекциями и избранным.',
            lpad(gs::text, 3, '0')
        ) AS full_description,
        (ARRAY[0, 6, 12, 16, 18])[((gs - 1) % 5) + 1] AS age_limit,
        format('https://seed.cityhawk.local/events/%s', lpad(gs::text, 3, '0')) AS source_url
    FROM generate_series(1, 100) AS gs
)
INSERT INTO event (author_user_id, title, location_description, full_description, age_limit, source_url)
SELECT
    sa.id,
    ge.title,
    ge.location_description,
    ge.full_description,
    ge.age_limit,
    ge.source_url
FROM generated_events ge
CROSS JOIN seed_author sa
WHERE NOT EXISTS (
    SELECT 1
    FROM event e
    WHERE e.source_url = ge.source_url
);

WITH place_map AS (
    SELECT
        ROW_NUMBER() OVER (ORDER BY p.name, p.id) AS idx,
        p.id,
        p.name
    FROM place p
    JOIN city c ON c.id = p.city_id
    WHERE c.country_name = 'Россия'
      AND c.name = 'Москва'
),
generated_sessions AS (
    SELECT
        gs AS n,
        format('https://seed.cityhawk.local/events/%s', lpad(gs::text, 3, '0')) AS source_url,
        ((gs - 1) % 10) + 1 AS place_idx,
        ('2026-05-01 10:00:00+03'::timestamptz + ((gs - 1) * interval '4 hours')) AS start_at,
        ('2026-05-01 10:00:00+03'::timestamptz + ((gs - 1) * interval '4 hours') + interval '2 hours') AS end_at,
        (500 + (((gs - 1) % 9) * 250))::integer AS price
    FROM generate_series(1, 100) AS gs
)
INSERT INTO event_session (event_id, place_id, start_at, end_at, price)
SELECT
    e.id,
    pm.id,
    gs.start_at,
    gs.end_at,
    gs.price
FROM generated_sessions gs
JOIN event e ON e.source_url = gs.source_url
JOIN place_map pm ON pm.idx = gs.place_idx
WHERE NOT EXISTS (
    SELECT 1
    FROM event_session es
    WHERE es.event_id = e.id
      AND es.place_id = pm.id
      AND es.start_at = gs.start_at
      AND es.end_at = gs.end_at
);

WITH generated_images AS (
    SELECT
        gs AS n,
        format('https://seed.cityhawk.local/events/%s', lpad(gs::text, 3, '0')) AS source_url,
        format('https://picsum.photos/seed/cityhawk-event-%s/1200/800', lpad(gs::text, 3, '0')) AS image_url
    FROM generate_series(1, 100) AS gs
)
INSERT INTO event_image (event_id, image_url)
SELECT
    e.id,
    gi.image_url
FROM generated_images gi
JOIN event e ON e.source_url = gi.source_url
WHERE NOT EXISTS (
    SELECT 1
    FROM event_image ei
    WHERE ei.event_id = e.id
      AND ei.image_url = gi.image_url
);

WITH category_map AS (
    SELECT
        ROW_NUMBER() OVER (ORDER BY c.name, c.id) AS idx,
        c.id
    FROM category c
),
generated_event_categories AS (
    SELECT
        format('https://seed.cityhawk.local/events/%s', lpad(gs::text, 3, '0')) AS source_url,
        ((gs - 1) % 12) + 1 AS category_idx
    FROM generate_series(1, 100) AS gs
    UNION
    SELECT
        format('https://seed.cityhawk.local/events/%s', lpad(gs::text, 3, '0')) AS source_url,
        ((gs + 4) % 12) + 1 AS category_idx
    FROM generate_series(1, 100) AS gs
)
INSERT INTO event_category (event_id, category_id)
SELECT
    e.id,
    cm.id
FROM generated_event_categories gec
JOIN event e ON e.source_url = gec.source_url
JOIN category_map cm ON cm.idx = gec.category_idx
ON CONFLICT (event_id, category_id) DO NOTHING;

WITH tag_map AS (
    SELECT
        ROW_NUMBER() OVER (ORDER BY t.name, t.id) AS idx,
        t.id
    FROM tag t
),
generated_event_tags AS (
    SELECT
        format('https://seed.cityhawk.local/events/%s', lpad(gs::text, 3, '0')) AS source_url,
        ((gs - 1) % 12) + 1 AS tag_idx
    FROM generate_series(1, 100) AS gs
    UNION
    SELECT
        format('https://seed.cityhawk.local/events/%s', lpad(gs::text, 3, '0')) AS source_url,
        ((gs + 6) % 12) + 1 AS tag_idx
    FROM generate_series(1, 100) AS gs
)
INSERT INTO event_tag (event_id, tag_id)
SELECT
    e.id,
    tm.id
FROM generated_event_tags getg
JOIN event e ON e.source_url = getg.source_url
JOIN tag_map tm ON tm.idx = getg.tag_idx
ON CONFLICT (event_id, tag_id) DO NOTHING;

WITH seed_author AS (
    SELECT u.id
    FROM user_account u
    WHERE u.email = 'seed.author@cityhawk.local'
)
INSERT INTO collection (author_user_id, title, description, is_public)
SELECT
    sa.id,
    v.title,
    v.description,
    TRUE
FROM seed_author sa
CROSS JOIN (
    VALUES
        ('Seed Collection Weekend', 'Подборка сидовых событий для главной страницы и списков.'),
        ('Seed Collection Family', 'Подборка семейных и дневных сидовых событий.'),
        ('Seed Collection Music', 'Подборка музыкальных и вечерних сидовых событий.')
) AS v(title, description)
WHERE NOT EXISTS (
    SELECT 1
    FROM collection c
    WHERE c.title = v.title
      AND c.author_user_id = sa.id
);

INSERT INTO collection_image (collection_id, image_url)
SELECT
    c.id,
    format('https://picsum.photos/seed/%s/1200/800', replace(lower(c.title), ' ', '-'))
FROM collection c
WHERE c.title IN ('Seed Collection Weekend', 'Seed Collection Family', 'Seed Collection Music')
  AND NOT EXISTS (
      SELECT 1
      FROM collection_image ci
      WHERE ci.collection_id = c.id
  );

WITH generated_collection_events AS (
    SELECT 'Seed Collection Weekend' AS collection_title, format('https://seed.cityhawk.local/events/%s', lpad(gs::text, 3, '0')) AS source_url
    FROM generate_series(1, 12) AS gs
    UNION ALL
    SELECT 'Seed Collection Family', format('https://seed.cityhawk.local/events/%s', lpad(gs::text, 3, '0'))
    FROM generate_series(13, 24) AS gs
    UNION ALL
    SELECT 'Seed Collection Music', format('https://seed.cityhawk.local/events/%s', lpad(gs::text, 3, '0'))
    FROM generate_series(25, 36) AS gs
)
INSERT INTO collection_event (collection_id, event_id)
SELECT
    c.id,
    e.id
FROM generated_collection_events gce
JOIN collection c ON c.title = gce.collection_title
JOIN event e ON e.source_url = gce.source_url
ON CONFLICT (collection_id, event_id) DO NOTHING;

WITH generated_favorites AS (
    SELECT format('https://seed.cityhawk.local/events/%s', lpad(gs::text, 3, '0')) AS source_url
    FROM generate_series(1, 10) AS gs
)
INSERT INTO favorite_event (user_id, event_id)
SELECT
    u.id,
    e.id
FROM generated_favorites gf
JOIN user_account u ON u.email = 'seed.author@cityhawk.local'
JOIN event e ON e.source_url = gf.source_url
ON CONFLICT (user_id, event_id) DO NOTHING;

COMMIT;
