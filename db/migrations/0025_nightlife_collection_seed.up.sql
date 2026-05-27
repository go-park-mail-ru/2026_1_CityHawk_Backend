-- CityHawk seed: коллекция «Ночная жизнь»
-- Ночные события, клубы, стендап и активный вечер для компании.

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
    ('Ночная жизнь'),
    ('Клубы'),
    ('Бары')
ON CONFLICT (name) DO NOTHING;

INSERT INTO tag (name)
VALUES
    ('Ночная жизнь'),
    ('Клуб'),
    ('Бар'),
    ('Вечеринка'),
    ('Танцы'),
    ('Музыка'),
    ('Электронная музыка'),
    ('Коктейли'),
    ('Для компании'),
    ('После полуночи'),
    ('DJ-сет'),
    ('Боулинг'),
    ('Активный вечер'),
    ('Стендап')
ON CONFLICT (name) DO NOTHING;

-- 3. Места
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
            'StandUp Store Moscow',
            'Москва, ул. Петровка, д. 21, стр. 1',
            55.761900::numeric,
            37.612400::numeric,
            'Стендап-площадка в центре Москвы для комедийных вечеров, встреч с друзьями и лёгкого старта ночного маршрута.'
        ),
        (
            'ANIMA',
            'Москва, ул. Сущёвская, д. 21',
            55.782200::numeric,
            37.601500::numeric,
            'Ночное клубное пространство на Менделеевской с вечеринками, концертами, шоу и выразительной визуальной атмосферой.'
        ),
        (
            'PIPL Bar',
            'Москва, Комсомольская площадь, д. 6, цокольный этаж',
            55.775700::numeric,
            37.655500::numeric,
            'Бар и клубное пространство у Комсомольской площади для ночных выходов, коктейлей, танцев и встреч компанией.'
        ),
        (
            'Mutabor',
            'Москва, ул. Шарикоподшипниковская, д. 13, стр. 32',
            55.720500::numeric,
            37.676400::numeric,
            'Пространство для электронной музыки, вечеринок и ночных событий с полноценной клубной программой.'
        ),
        (
            'Боулинг-клуб «Самокат»',
            'Москва, ул. Самокатная, д. 2, корп. 1',
            55.746500::numeric,
            37.681800::numeric,
            'Боулинг-клуб для активного вечера с компанией, едой, напитками и азартом без большого клубного шума.'
        ),
        (
            'Powerhouse Moscow',
            'Москва, ул. Гончарная, д. 7/4',
            55.762200::numeric,
            37.651600::numeric,
            'Бар и музыкальное пространство с диджей-сетами, коктейлями и камерной клубной атмосферой.'
        )
) AS v(name, address_line, latitude, longitude, description)
WHERE c.country_name = 'Россия'
  AND c.name = 'Москва'
ON CONFLICT (city_id, name, address_line) DO UPDATE
SET latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    description = EXCLUDED.description,
    updated_at = now();

-- 4. Новые события. Сессии не создаются: расписание ночных программ меняется.
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
    18,
    NULL
FROM user_account u
CROSS JOIN (
    VALUES
        (
            'Вечер стендапа в StandUp Store Moscow',
            'vecher-stendapa-standup-store-moscow',
            'StandUp Store Moscow, ул. Петровка, д. 21, стр. 1',
            'Стендап-вечер в центре города: можно прийти компанией, посмеяться и продолжить вечер где-то рядом. Подходит тем, кто хочет лёгкий и понятный план на вечер.'
        ),
        (
            'Вечеринка в ANIMA',
            'vecherinka-anima',
            'ANIMA, ул. Сущёвская, д. 21',
            'Ночное клубное пространство на Менделеевской с вечеринками, концертами, шоу и сильной визуальной атмосферой. Подходит для тех, кто хочет полноценный ночной выход с музыкой, танцами и ощущением большого события.'
        ),
        (
            'Ночная вечеринка в PIPL',
            'nochnaya-vecherinka-pipl',
            'PIPL Bar, Комсомольская площадь, д. 6',
            'Клубный вечер с музыкой, танцами, коктейлями и атмосферой ночной Москвы. Подойдёт для компании друзей, дня рождения или спонтанного выхода после полуночи.'
        ),
        (
            'Электронная ночь в Mutabor',
            'elektronnaya-noch-mutabor',
            'Mutabor, ул. Шарикоподшипниковская, д. 13, стр. 32',
            'Пространство для электронной музыки, вечеринок и ночных событий. Подойдёт для тех, кто хочет полноценную ночную программу и клубную атмосферу.'
        ),
        (
            'Ночной боулинг в «Самокат»',
            'nochnoy-bouling-samokat',
            'Боулинг-клуб «Самокат», ул. Самокатная, д. 2, корп. 1',
            'Активный вечер для компании: боулинг, еда, напитки и азарт без клубного шума. Хороший вариант, если хочется не просто сидеть за столом.'
        ),
        (
            'Танцы и коктейли в Powerhouse',
            'tantsy-i-kokteyli-powerhouse',
            'Powerhouse Moscow, ул. Гончарная, д. 7/4',
            'Бар, музыка, диджей-сеты и камерная клубная атмосфера. Формат для тех, кто хочет вечер с движением, но без ощущения огромного ночного клуба.'
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
        ('vecher-stendapa-standup-store-moscow', 'StandUp Store Moscow'),
        ('vecherinka-anima', 'ANIMA'),
        ('nochnaya-vecherinka-pipl', 'PIPL Bar'),
        ('elektronnaya-noch-mutabor', 'Mutabor'),
        ('nochnoy-bouling-samokat', 'Боулинг-клуб «Самокат»'),
        ('tantsy-i-kokteyli-powerhouse', 'Powerhouse Moscow')
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
    'vecher-stendapa-standup-store-moscow',
    'vecherinka-anima',
    'nochnaya-vecherinka-pipl',
    'elektronnaya-noch-mutabor',
    'nochnoy-bouling-samokat',
    'tantsy-i-kokteyli-powerhouse'
  );

INSERT INTO event_image (event_id, image_url)
SELECT e.id, v.image_url
FROM (
    VALUES
        ('vecher-stendapa-standup-store-moscow', '/uploads/events/image-1.jpg'),
        ('vecherinka-anima', '/uploads/events/85d82880aa112535cddc5567e5bf7532_w828_h552--big.jpg'),
        ('nochnaya-vecherinka-pipl', '/uploads/events/orig-5.jpeg'),
        ('elektronnaya-noch-mutabor', '/uploads/events/b2a78d6088416ec3e8dac1ef761797fc.jpg'),
        ('nochnoy-bouling-samokat', '/uploads/events/luchshiye_bouling_kluby_moskvy_17.jpg'),
        ('tantsy-i-kokteyli-powerhouse', '/uploads/events/dewar-s-powerhouse.jpg')
) AS v(event_slug, image_url)
JOIN event e ON e.slug = v.event_slug
ON CONFLICT (event_id, image_url) DO NOTHING;

-- 7. Категории новых событий
INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM (
    VALUES
        ('vecher-stendapa-standup-store-moscow', 'Стендап'),
        ('vecherinka-anima', 'Клубы'),
        ('nochnaya-vecherinka-pipl', 'Бары'),
        ('elektronnaya-noch-mutabor', 'Клубы'),
        ('nochnoy-bouling-samokat', 'Развлечения'),
        ('tantsy-i-kokteyli-powerhouse', 'Бары')
) AS v(event_slug, category_name)
JOIN event e ON e.slug = v.event_slug
JOIN category c ON c.name = v.category_name
ON CONFLICT (event_id, category_id) DO NOTHING;

-- 8. Теги новых событий
INSERT INTO event_tag (event_id, tag_id)
SELECT e.id, t.id
FROM (
    VALUES
        ('vecher-stendapa-standup-store-moscow', 'Ночная жизнь'),
        ('vecher-stendapa-standup-store-moscow', 'Стендап'),
        ('vecher-stendapa-standup-store-moscow', 'Для компании'),

        ('vecherinka-anima', 'Ночная жизнь'),
        ('vecherinka-anima', 'Клуб'),
        ('vecherinka-anima', 'Вечеринка'),
        ('vecherinka-anima', 'Танцы'),
        ('vecherinka-anima', 'Музыка'),

        ('nochnaya-vecherinka-pipl', 'Ночная жизнь'),
        ('nochnaya-vecherinka-pipl', 'Бар'),
        ('nochnaya-vecherinka-pipl', 'Вечеринка'),
        ('nochnaya-vecherinka-pipl', 'Коктейли'),
        ('nochnaya-vecherinka-pipl', 'После полуночи'),

        ('elektronnaya-noch-mutabor', 'Ночная жизнь'),
        ('elektronnaya-noch-mutabor', 'Клуб'),
        ('elektronnaya-noch-mutabor', 'Электронная музыка'),
        ('elektronnaya-noch-mutabor', 'DJ-сет'),
        ('elektronnaya-noch-mutabor', 'Танцы'),

        ('nochnoy-bouling-samokat', 'Ночная жизнь'),
        ('nochnoy-bouling-samokat', 'Боулинг'),
        ('nochnoy-bouling-samokat', 'Активный вечер'),
        ('nochnoy-bouling-samokat', 'Для компании'),

        ('tantsy-i-kokteyli-powerhouse', 'Ночная жизнь'),
        ('tantsy-i-kokteyli-powerhouse', 'Бар'),
        ('tantsy-i-kokteyli-powerhouse', 'Коктейли'),
        ('tantsy-i-kokteyli-powerhouse', 'DJ-сет'),
        ('tantsy-i-kokteyli-powerhouse', 'Танцы')
) AS v(event_slug, tag_name)
JOIN event e ON e.slug = v.event_slug
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;

-- 9. Коллекция «Ночная жизнь»
INSERT INTO collection (author_user_id, city_id, title, slug, description, is_public)
SELECT
    u.id,
    c.id,
    'Ночная жизнь',
    'nightlife-moscow',
    'Клубы, бары, стендап, боулинг и ночные события Москвы для компании и спонтанного выхода после заката.',
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
  AND col.slug = 'nightlife-moscow';

INSERT INTO collection_image (collection_id, image_url)
SELECT
    col.id,
    '/uploads/events/85d82880aa112535cddc5567e5bf7532_w828_h552--big.jpg'
FROM collection col
WHERE col.slug = 'nightlife-moscow'
ON CONFLICT (collection_id, image_url) DO NOTHING;

-- 11. Состав коллекции
DELETE FROM collection_event ce
USING collection col
WHERE ce.collection_id = col.id
  AND col.slug = 'nightlife-moscow';

INSERT INTO collection_event (collection_id, event_id)
SELECT
    col.id,
    e.id
FROM collection col
JOIN event e ON e.slug IN (
    -- Новые ночные события
    'vecher-stendapa-standup-store-moscow',
    'vecherinka-anima',
    'nochnaya-vecherinka-pipl',
    'elektronnaya-noch-mutabor',
    'nochnoy-bouling-samokat',
    'tantsy-i-kokteyli-powerhouse',

    -- Стендап
    'zhenskiy-standup-vdnh-2026-05-29',
    'ivan-abramov-vsyo-iz-detstva-2026-10-16',
    'sergey-orlov-zvezda-2026-05-31',
    'andrey-beburishvili-udobnyy-2026-06-14',
    'besplatnyy-stand-up-ot-komikov-s-tnt-2026-06'
)
WHERE col.slug = 'nightlife-moscow'
ON CONFLICT (collection_id, event_id) DO NOTHING;

COMMIT;
