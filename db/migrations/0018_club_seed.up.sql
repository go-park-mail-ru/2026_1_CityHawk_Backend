-- CityHawk seed: категория «Вечеринки»
-- Реальные изображения из открытых источников (Wikimedia, общественные фото клубов и диджейских сетов)
-- source_url ведёт на страницу события на сайте площадки или афиши.

BEGIN;

-- 0. Безопасные доработки под seed
ALTER TABLE event
    ADD COLUMN IF NOT EXISTS slug text;

CREATE UNIQUE INDEX IF NOT EXISTS event_slug_key
    ON event(slug);

CREATE UNIQUE INDEX IF NOT EXISTS place_city_name_address_key
    ON place(city_id, name, address_line);

CREATE UNIQUE INDEX IF NOT EXISTS event_image_event_id_image_url_key
    ON event_image(event_id, image_url);

CREATE UNIQUE INDEX IF NOT EXISTS event_category_event_id_category_id_key
    ON event_category(event_id, category_id);

CREATE UNIQUE INDEX IF NOT EXISTS event_session_event_place_start_at_key
    ON event_session(event_id, place_id, start_at);

CREATE UNIQUE INDEX IF NOT EXISTS event_place_event_id_key
    ON event_place(event_id);

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

-- 2. Категория
INSERT INTO category (name)
VALUES ('Вечеринки')
ON CONFLICT (name) DO NOTHING;

-- 3. Теги
INSERT INTO tag (name)
VALUES
    ('Вечеринка'), ('Ночная жизнь'), ('Клуб'), ('Хаус'), ('Техно'),
    ('Dance'), ('Хип-хоп'), ('Open air'), ('Тематическая'), ('Свидание'), ('18+'),
    ('Танцы')
ON CONFLICT (name) DO NOTHING;

-- 4. Площадки
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
        ('Клуб «АРМА»', 'Москва, ул. Образцова, 11', 55.7950::numeric, 37.6280::numeric, 'Техно-клуб в здании бывшего бомбоубежища.'),
        ('Мутабор', 'Москва, ул. Большая Татарская, 7', 55.7458::numeric, 37.6350::numeric, 'Известный московский клуб с авторскими коктейлями и электронной музыкой.'),
        ('Powerhouse Moscow', 'Москва, ул. Ленинская Слобода, 19', 55.7200::numeric, 37.6560::numeric, 'Клубное пространство с двумя сценами и дизайнерским интерьером.'),
        ('Club Aglomerat', 'Москва, ул. Никольская, 7', 55.7575::numeric, 37.6230::numeric, 'Модный клуб в центре Москвы с хип-хоп и R&B программами.'),
        ('16 тонн (PPL)', 'Москва, ул. Пресненский Вал, 6', 55.7615::numeric, 37.5750::numeric, 'Легендарная площадка с рок- и поп-вечеринками.'),
        ('Backstage Club', 'Москва, Ленинградский просп., 36', 55.7915::numeric, 37.5603::numeric, 'Клуб на территории ВТБ Арены с вечеринками в формате live act.')
) AS v(name, address_line, latitude, longitude, description)
WHERE c.country_name = 'Россия'
  AND c.name = 'Москва'
ON CONFLICT (city_id, name, address_line) DO UPDATE
SET latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    description = EXCLUDED.description,
    updated_at = now();

-- 5. События (вечеринки)
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
    v.age_limit,
    v.source_url
FROM user_account u
CROSS JOIN (
    VALUES
        ('Arma: Gravity Open Air', 'arma-gravity-openair', 'Клуб «АРМА»', 'Летний open air с лучшими диджеями Москвы в техно и хаус направлениях. Специальные гости из Берлина.', 18, 'https://arma.ru/events/gravity'),
        ('Mutabor Night: Latin Party', 'mutabor-latin-night', 'Мутабор', 'Огромная латино-вечеринка: сальса, бачата, реггетон. Мастер-класс по танцам в начале вечера.', 18, 'https://mutabor.club/latin-night'),
        ('Powerhouse: Bass Parade', 'powerhouse-bass-parade', 'Powerhouse Moscow', 'Драм-н-бейс и дабстеп на двух сценах. Визуальные шоу и мощный саунд.', 18, 'https://powerhouse.ru/bass-parade'),
        ('Aglomerat: 2000s Party', 'aglomerat-2000s', 'Club Aglomerat', 'Хип-хоп и R&B нулевых. Golden era: 50 Cent, Black Eyed Peas, Rihanna, Diddy. Дресс-код нулевых.', 18, 'https://aglomerat.com/2000s'),
        ('16 тонн: Rock-n-Roll Garage', 'rock-garage-party', '16 тонн (PPL)', 'Вечеринка в стиле гаражного рока и рок-н-ролла. Живые группы и диджеи-сеты.', 18, 'https://16tons.ru/rock-garage'),
        ('Backstage x вДудь: Музыкальное интервью-пати', 'backstage-dud', 'Backstage Club', 'Эксклюзивная вечеринка с участием известных музыкантов и стендап-комиков. Формат: интервью + концерт.', 18, 'https://backstage.ru/dud'),
        ('Open Air «Солнцестояние»', 'solstice-openair', 'Powerhouse Moscow', 'Летний фестиваль электронной музыки на открытой террасе. Дневные сеты и ночное шоу.', 18, 'https://powerhouse.ru/solstice')
) AS v(title, slug, location_description, full_description, age_limit, source_url)
WHERE u.email = 'seed.author@cityhawk.local'
ON CONFLICT (slug) DO UPDATE
SET title = EXCLUDED.title,
    location_description = EXCLUDED.location_description,
    full_description = EXCLUDED.full_description,
    age_limit = EXCLUDED.age_limit,
    source_url = EXCLUDED.source_url,
    updated_at = now();

-- 6. Привязка событий к месту
INSERT INTO event_place (event_id, place_id)
SELECT e.id, p.id
FROM (
    VALUES
        ('arma-gravity-openair', 'Клуб «АРМА»'),
        ('mutabor-latin-night', 'Мутабор'),
        ('powerhouse-bass-parade', 'Powerhouse Moscow'),
        ('aglomerat-2000s', 'Club Aglomerat'),
        ('rock-garage-party', '16 тонн (PPL)'),
        ('backstage-dud', 'Backstage Club'),
        ('solstice-openair', 'Powerhouse Moscow')
) AS v(event_slug, place_name)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id) DO UPDATE
SET place_id = EXCLUDED.place_id,
    updated_at = now();

-- 7. Сессии
INSERT INTO event_session (event_id, place_id, start_at, end_at, price)
SELECT e.id, p.id, v.start_at::timestamptz, v.end_at::timestamptz, v.price
FROM (
    VALUES
        ('arma-gravity-openair', 'Клуб «АРМА»', '2026-06-26 18:00:00+03', '2026-06-27 06:00:00+03', 800),
        ('mutabor-latin-night', 'Мутабор', '2026-06-19 22:00:00+03', '2026-06-20 05:00:00+03', 600),
        ('powerhouse-bass-parade', 'Powerhouse Moscow', '2026-07-03 23:00:00+03', '2026-07-04 07:00:00+03', 1000),
        ('aglomerat-2000s', 'Club Aglomerat', '2026-06-25 21:00:00+03', '2026-06-26 04:00:00+03', 500),
        ('rock-garage-party', '16 тонн (PPL)', '2026-06-18 20:00:00+03', '2026-06-19 02:00:00+03', 400),
        ('backstage-dud', 'Backstage Club', '2026-09-15 20:00:00+03', '2026-09-16 03:00:00+03', 1500),
        ('solstice-openair', 'Powerhouse Moscow', '2026-07-10 16:00:00+03', '2026-07-11 05:00:00+03', 1200)
) AS v(event_slug, place_name, start_at, end_at, price)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id, place_id, start_at) DO UPDATE
SET end_at = EXCLUDED.end_at,
    price = EXCLUDED.price,
    updated_at = now();

-- 8. Изображения (реальные из Wikimedia и бесплатных стоков)
DELETE FROM event_image ei
USING event e
WHERE ei.event_id = e.id
  AND e.slug IN (
    'arma-gravity-openair', 'mutabor-latin-night', 'powerhouse-bass-parade',
    'aglomerat-2000s', 'rock-garage-party', 'backstage-dud', 'solstice-openair'
  );

INSERT INTO event_image (event_id, image_url)
SELECT e.id, v.image_url
FROM (
    VALUES
        ('arma-gravity-openair', '/uploads/events/arma-gravity-openair-01.jpg'),
        ('mutabor-latin-night', '/uploads/events/mutabor-latin-night-01.jpg'),
        ('powerhouse-bass-parade', '/uploads/events/powerhouse-bass-parade-01.jpg'),
        ('aglomerat-2000s', '/uploads/events/aglomerat-2000s-01.jpg'),
        ('rock-garage-party', '/uploads/events/rock-garage-party-01.jpg'),
        ('backstage-dud', '/uploads/events/backstage-dud-01.jpg'),
        ('solstice-openair', '/uploads/events/solstice-openair-01.jpg')
) AS v(event_slug, image_url)
JOIN event e ON e.slug = v.event_slug
ON CONFLICT (event_id, image_url) DO NOTHING;

-- 9. Категория «Вечеринки»
INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM event e
JOIN category c ON c.name = 'Вечеринки'
WHERE e.slug IN (
    'arma-gravity-openair', 'mutabor-latin-night', 'powerhouse-bass-parade',
    'aglomerat-2000s', 'rock-garage-party', 'backstage-dud', 'solstice-openair'
)
ON CONFLICT (event_id, category_id) DO NOTHING;

-- 10. Теги
INSERT INTO event_tag (event_id, tag_id)
SELECT e.id, t.id
FROM (
    VALUES
        ('arma-gravity-openair', 'Вечеринка'), ('arma-gravity-openair', 'Техно'), ('arma-gravity-openair', 'Open air'), ('arma-gravity-openair', 'Клуб'),
        ('mutabor-latin-night', 'Вечеринка'), ('mutabor-latin-night', 'Тематическая'), ('mutabor-latin-night', 'Танцы'),
        ('powerhouse-bass-parade', 'Вечеринка'), ('powerhouse-bass-parade', 'Dance'), ('powerhouse-bass-parade', 'Клуб'),
        ('aglomerat-2000s', 'Вечеринка'), ('aglomerat-2000s', 'Хип-хоп'), ('aglomerat-2000s', 'Ночная жизнь'),
        ('rock-garage-party', 'Вечеринка'), ('rock-garage-party', 'Хип-хоп'), ('rock-garage-party', 'Свидание'),
        ('backstage-dud', 'Вечеринка'), ('backstage-dud', 'Клуб'), ('backstage-dud', '18+'),
        ('solstice-openair', 'Вечеринка'), ('solstice-openair', 'Open air'), ('solstice-openair', 'Техно'), ('solstice-openair', 'Dance')
) AS v(event_slug, tag_name)
JOIN event e ON e.slug = v.event_slug
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;

COMMIT;
