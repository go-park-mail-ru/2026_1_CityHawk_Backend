-- CityHawk seed: коллекция «Для свидания»
-- Идеи для свидания: мастер-классы, театр, выставки и прогулки.

BEGIN;

-- 0. Безопасные доработки под seed
ALTER TABLE event
    ADD COLUMN IF NOT EXISTS slug text;

CREATE UNIQUE INDEX IF NOT EXISTS event_slug_key
    ON event(slug);

ALTER TABLE collection
    ADD COLUMN IF NOT EXISTS city_id uuid,
    ADD COLUMN IF NOT EXISTS slug text;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'collection_city_id_fkey'
    ) THEN
        ALTER TABLE collection
            ADD CONSTRAINT collection_city_id_fkey
            FOREIGN KEY (city_id)
            REFERENCES city(id)
            ON UPDATE CASCADE
            ON DELETE RESTRICT;
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS collection_slug_key
    ON collection(slug);

CREATE UNIQUE INDEX IF NOT EXISTS place_city_name_address_key
    ON place(city_id, name, address_line);

CREATE UNIQUE INDEX IF NOT EXISTS event_image_event_id_image_url_key
    ON event_image(event_id, image_url);

CREATE UNIQUE INDEX IF NOT EXISTS event_category_event_id_category_id_key
    ON event_category(event_id, category_id);

CREATE UNIQUE INDEX IF NOT EXISTS event_place_event_id_key
    ON event_place(event_id);

CREATE UNIQUE INDEX IF NOT EXISTS event_tag_event_id_tag_id_key
    ON event_tag(event_id, tag_id);

CREATE UNIQUE INDEX IF NOT EXISTS collection_image_collection_id_image_url_key
    ON collection_image(collection_id, image_url);

-- 1. Город и seed-автор
INSERT INTO city (name, country_name, timezone)
VALUES ('Москва', 'Россия', 'Europe/Moscow')
ON CONFLICT (country_name, name) DO UPDATE
SET timezone = EXCLUDED.timezone,
    updated_at = now();

INSERT INTO user_account (
    email,
    username,
    password_hash,
    city_id
)
SELECT
    'seed.author@cityhawk.local',
    'CityHawk',
    'seed-password-hash',
    c.id
FROM city c
WHERE c.country_name = 'Россия'
  AND c.name = 'Москва'
ON CONFLICT (email) DO UPDATE
SET city_id = EXCLUDED.city_id,
    updated_at = now();

-- 2. Категории и теги
INSERT INTO category (name)
VALUES
    ('Мастер-классы'),
    ('Танцы')
ON CONFLICT (name) DO NOTHING;

INSERT INTO tag (name)
VALUES
    ('Свидание'),
    ('Для двоих'),
    ('Мастер-класс'),
    ('Творчество'),
    ('Керамика'),
    ('Ручная работа'),
    ('Танцы'),
    ('Парные танцы'),
    ('Для начинающих'),
    ('Движение'),
    ('Спокойный вечер'),
    ('Для компании')
ON CONFLICT (name) DO NOTHING;

-- 3. Новые места
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
        (
            'Chalet Studio, Красный Октябрь',
            'Москва, Берсеневский пер., 5, стр. 1',
            55.741900::numeric,
            37.609600::numeric,
            'Творческая студия на Красном Октябре для спокойных мастер-классов, камерных встреч и вечеров с новым опытом.'
        ),
        (
            'Ivara, Трёхгорный вал',
            'Москва, ул. Трёхгорный Вал, 3',
            55.762800::numeric,
            37.563400::numeric,
            'Студия парных танцев на Трёхгорном Валу, где можно попробовать танцевальный формат без опыта и провести живой вечер вдвоём или с друзьями.'
        )
) AS v(name, address_line, latitude, longitude, description)
WHERE c.country_name = 'Россия'
  AND c.name = 'Москва'
ON CONFLICT (city_id, name, address_line) DO UPDATE
SET latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    description = EXCLUDED.description,
    updated_at = now();

-- 4. Новые события для свидания. Сессии не создаются: точного расписания нет.
INSERT INTO event (
    author_user_id,
    title,
    slug,
    location_description,
    full_description,
    age_limit,
    source_url
)
SELECT
    u.id,
    v.title,
    v.slug,
    v.location_description,
    v.full_description,
    0,
    NULL
FROM user_account u
CROSS JOIN (
    VALUES
        (
            'Гончарный мастер-класс в Chalet Studio',
            'goncharnyy-master-klass-chalet-studio',
            'Chalet Studio, Берсеневский пер., 5, стр. 1',
            'Творческий мастер-класс, где можно сделать кружку, вазу или тарелку своими руками. Подходит для пары, друзей или спокойного вечера с новым опытом.'
        ),
        (
            'Парные танцы в Ivara',
            'parnye-tantsy-ivara',
            'Ivara, ул. Трёхгорный Вал, 3',
            'Мастер-класс по парным танцам для начинающих: можно попробовать хастл, свинг или танго без опыта. Формат хорошо подходит для двоих или компании друзей — много движения, общения и лёгкой неловкости, которая быстро превращается в веселье.'
        )
) AS v(title, slug, location_description, full_description)
WHERE u.email = 'seed.author@cityhawk.local'
ON CONFLICT (slug) DO UPDATE
SET title = EXCLUDED.title,
    location_description = EXCLUDED.location_description,
    full_description = EXCLUDED.full_description,
    age_limit = EXCLUDED.age_limit,
    source_url = EXCLUDED.source_url,
    updated_at = now();

-- 5. Привязка новых событий к местам
INSERT INTO event_place (event_id, place_id)
SELECT e.id, p.id
FROM (
    VALUES
        ('goncharnyy-master-klass-chalet-studio', 'Chalet Studio, Красный Октябрь'),
        ('parnye-tantsy-ivara', 'Ivara, Трёхгорный вал')
) AS v(event_slug, place_name)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id) DO UPDATE
SET place_id = EXCLUDED.place_id,
    updated_at = now();

-- 6. Картинки новых событий
DELETE FROM event_image ei
USING event e
WHERE ei.event_id = e.id
  AND e.slug IN (
    'goncharnyy-master-klass-chalet-studio',
    'parnye-tantsy-ivara'
  );

INSERT INTO event_image (event_id, image_url)
SELECT e.id, v.image_url
FROM (
    VALUES
        ('goncharnyy-master-klass-chalet-studio', '/uploads/events/foto_26_04_2024_13_54_30.jpg'),
        ('goncharnyy-master-klass-chalet-studio', '/uploads/events/123.png'),
        ('goncharnyy-master-klass-chalet-studio', '/uploads/events/15.jpg'),
        ('parnye-tantsy-ivara', '/uploads/events/lalala.png')
) AS v(event_slug, image_url)
JOIN event e ON e.slug = v.event_slug
ON CONFLICT (event_id, image_url) DO NOTHING;

-- 7. Категории новых событий
INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM (
    VALUES
        ('goncharnyy-master-klass-chalet-studio', 'Мастер-классы'),
        ('parnye-tantsy-ivara', 'Танцы')
) AS v(event_slug, category_name)
JOIN event e ON e.slug = v.event_slug
JOIN category c ON c.name = v.category_name
ON CONFLICT (event_id, category_id) DO NOTHING;

-- 8. Теги новых событий
INSERT INTO event_tag (event_id, tag_id)
SELECT e.id, t.id
FROM (
    VALUES
        ('goncharnyy-master-klass-chalet-studio', 'Свидание'),
        ('goncharnyy-master-klass-chalet-studio', 'Для двоих'),
        ('goncharnyy-master-klass-chalet-studio', 'Мастер-класс'),
        ('goncharnyy-master-klass-chalet-studio', 'Творчество'),
        ('goncharnyy-master-klass-chalet-studio', 'Керамика'),
        ('goncharnyy-master-klass-chalet-studio', 'Ручная работа'),
        ('goncharnyy-master-klass-chalet-studio', 'Спокойный вечер'),

        ('parnye-tantsy-ivara', 'Свидание'),
        ('parnye-tantsy-ivara', 'Для двоих'),
        ('parnye-tantsy-ivara', 'Мастер-класс'),
        ('parnye-tantsy-ivara', 'Танцы'),
        ('parnye-tantsy-ivara', 'Парные танцы'),
        ('parnye-tantsy-ivara', 'Для начинающих'),
        ('parnye-tantsy-ivara', 'Движение'),
        ('parnye-tantsy-ivara', 'Для компании')
) AS v(event_slug, tag_name)
JOIN event e ON e.slug = v.event_slug
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;

-- 9. Коллекция «Для свидания»
INSERT INTO collection (author_user_id, city_id, title, slug, description, is_public)
SELECT
    u.id,
    c.id,
    'Для свидания',
    'date-ideas-moscow',
    'Идеи для свидания в Москве: творческие мастер-классы, танцы, театр, выставки и красивые прогулочные места.',
    TRUE
FROM user_account u
JOIN city c ON c.country_name = 'Россия'
           AND c.name = 'Москва'
WHERE u.email = 'seed.author@cityhawk.local'
ON CONFLICT (slug) DO UPDATE
SET title = EXCLUDED.title,
    city_id = EXCLUDED.city_id,
    description = EXCLUDED.description,
    is_public = EXCLUDED.is_public,
    updated_at = now();

-- 10. Обложка коллекции
DELETE FROM collection_image ci
USING collection col
WHERE ci.collection_id = col.id
  AND col.slug = 'date-ideas-moscow';

INSERT INTO collection_image (collection_id, image_url)
SELECT
    col.id,
    '/uploads/events/foto_26_04_2024_13_54_30.jpg'
FROM collection col
WHERE col.slug = 'date-ideas-moscow'
ON CONFLICT (collection_id, image_url) DO NOTHING;

-- 11. Состав коллекции
DELETE FROM collection_event ce
USING collection col
WHERE ce.collection_id = col.id
  AND col.slug = 'date-ideas-moscow';

INSERT INTO collection_event (collection_id, event_id)
SELECT
    col.id,
    e.id
FROM collection col
JOIN event e ON e.slug IN (
    -- Новые события
    'goncharnyy-master-klass-chalet-studio',
    'parnye-tantsy-ivara',

    -- Театры
    'intuicziya',
    'romeo-i-dzhuletta-lyubov-vne-vremeni',
    'dve-anny',
    'pesn-lyubvi-pesn-skorbi',
    'prizrak-myuzikla',

    -- Выставки
    'futurione-2026',
    'sekrety-dzhuzeppe-archimboldo-2026',
    'orbity-khrupkikh-tel-2026',
    'ocharovanie-krasoty-2026',

    -- Прогулки и красивые места
    'gorky-park',
    'neskuchny-garden',
    'kitay-gorod',
    'novodevichy-monastery',
    'savvinskoe-podvorie',
    'vinzavod'
)
WHERE col.slug = 'date-ideas-moscow'
ON CONFLICT (collection_id, event_id) DO NOTHING;

COMMIT;
