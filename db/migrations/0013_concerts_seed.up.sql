-- CityHawk seed: категория «Концерты»
-- Основной источник данных: страницы Яндекс Афиши из запроса.
-- Важно: прямые media-URL Яндекс Афиши не всегда отдаются в статическом HTML,
-- поэтому для части событий оставлены прямые media-URL, которые удалось получить из открытых источников.
-- source_url у каждого события ведет на исходную страницу Яндекс Афиши.

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
    user_surname,
    password_hash,
    city_id
)
SELECT
    'seed.author@cityhawk.local',
    'CityHawk',
    'Seed',
    'seed-password-hash',
    c.id
FROM city c
WHERE c.country_name = 'Россия'
  AND c.name = 'Москва'
ON CONFLICT (email) DO UPDATE
SET city_id = EXCLUDED.city_id,
    updated_at = now();

-- 2. Категория и теги
INSERT INTO category (name)
VALUES ('Концерты')
ON CONFLICT (name) DO NOTHING;

INSERT INTO tag (name)
VALUES
    ('Концерт'),
    ('Музыка'),
    ('Поп'),
    ('Хип-хоп'),
    ('Рэп'),
    ('На воздухе'),
    ('Стадион'),
    ('Большое шоу'),
    ('Летний концерт'),
    ('Тур')
ON CONFLICT (name) DO NOTHING;

-- 3. Площадки
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
        ('Большая спортивная арена «Лужники»', 'Москва, ул. Лужники, 24, стр. 1', 55.715800::numeric, 37.553700::numeric, 'Крупнейшая стадионная площадка Москвы для масштабных концертов и спортивных событий.'),
        ('VK Музыка Летом', 'Москва, Шмитовский пр-д, 32а, клуб Atmosphere', 55.754400::numeric, 37.534600::numeric, 'Летняя концертная площадка под открытым небом рядом с пространством Atmosphere.'),
        ('TAU', 'Москва, Рязанский просп., 8а, стр. 10', 55.727000::numeric, 37.726000::numeric, 'Концертная площадка для клубных шоу, рэп-концертов и городских музыкальных событий.'),
        ('Red Summer', 'Москва, ул. Автозаводская, 23а, площадь у ЦСКА Арены', 55.702000::numeric, 37.642800::numeric, 'Летняя площадка у ЦСКА Арены для концертов на открытом воздухе.'),
        ('ВТБ Арена', 'Москва, Ленинградский просп., 36', 55.791500::numeric, 37.560300::numeric, 'Крупная московская арена для концертов, спортивных событий и больших шоу.'),
        ('ЦСКА Арена', 'Москва, ул. Автозаводская, 23а', 55.702400::numeric, 37.642600::numeric, 'Многофункциональная арена у метро ЗИЛ и Автозаводская для крупных концертов и шоу.')
) AS v(name, address_line, latitude, longitude, description)
WHERE c.country_name = 'Россия'
  AND c.name = 'Москва'
ON CONFLICT (city_id, name, address_line) DO UPDATE
SET latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    description = EXCLUDED.description,
    updated_at = now();

-- 4. События
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
        ('Руки Вверх!', 'ruki-vverkh-30-nam-uzhe', 'Большая спортивная арена «Лужники»', 'Юбилейный стадионный концерт «30 нам уже!» в Лужниках. Большое шоу к 30-летию группы: хиты, которые знают несколько поколений, масштабная сцена и атмосфера летнего музыкального праздника.', 6, 'https://afisha.yandex.ru/moscow/concert/ruki-vverkh-30-nam-uzhe?city=moscow&rubricCode=concert&refererSurface=top-events'),
        ('Toxi$', 'toxi-tour', 'VK Музыка Летом', 'Летний концерт Toxi$ на площадке VK Музыка Летом. Живое выступление молодого артиста, новый материал, любимые треки и формат open-air, где можно прийти заранее и комфортно провести вечер.', 16, 'https://afisha.yandex.ru/moscow/concert/toxi-tour?source=selection-events&rubricCode=concert&refererSelection=concert-hiphop&refererSurface=selection'),
        ('Леонид Агутин', 'leonid-agutin-tour-2026', 'Большая спортивная арена «Лужники»', 'Большой стадионный концерт Леонида Агутина с лучшими хитами. В программе — поп-музыка, фанк, поп-рок и фирменная концертная энергетика артиста на одной из главных московских арен.', 6, 'https://afisha.yandex.ru/moscow/concert/leonid-agutin-tour?city=moscow&rubricCode=concert&refererSurface=top-events'),
        ('Bushido Zho', 'bushido-zho-2026-06-27', 'TAU', 'Специальное летнее шоу Bushido Zho в Москве. Концерт для поклонников актуальной рэп-сцены: плотный звук, энергичная подача и вечер с треками одного из главных артистов новой волны.', 16, 'https://afisha.yandex.ru/moscow/concert/bushido-zho-2026-06-27?source=selection-events&rubricCode=concert&refererSelection=concert-hiphop&refererSurface=selection'),
        ('OG Buda', 'og-buda-tour-2026', 'Red Summer', 'Летний концерт OG Buda с музыкантами на площадке Red Summer. Open-air формат, живое выступление и атмосфера большого городского события для поклонников хип-хопа.', 16, 'https://afisha.yandex.ru/moscow/concert/og-buda-tour?source=selection-events&selectionPage=concert-hiphop&refererSelection=main_page_landing&refererSurface=selection'),
        ('SALUKI', 'saluki-2026-06-28', 'VK Музыка Летом', 'Специальное шоу SALUKI под открытым небом в рамках проекта VK Музыка Летом. Концерт с авторским звучанием, атмосферой летнего фестиваля и удобным open-air форматом.', 16, 'https://afisha.yandex.ru/moscow/concert/saluki-2026-06-27?source=selection-events&selectionPage=concert-hiphop&refererSelection=main_page_landing&refererSurface=selection'),
        ('Big Baby Tape', 'big-baby-tape-tour-2026', 'ВТБ Арена', 'Концерт Big Baby Tape в Москве. Большое рэп-шоу на арене с узнаваемыми треками, плотным саундом и энергией крупной концертной площадки.', 16, 'https://afisha.yandex.ru/moscow/concert/big-baby-tape-tour?source=selection-events&selectionPage=concert-hiphop&refererSelection=main_page_landing&refererSurface=selection'),
        ('Friendly Thug 52 Ngg', 'friendly-thug-52-ngg-tour-2026', 'ЦСКА Арена', 'Концерт Friendly Thug 52 Ngg в рамках Mille Grazie Tour. Сольный тур артиста новой волны русской рэп-культуры с честным саундом и южным хип-хоп настроением.', 16, 'https://afisha.yandex.ru/moscow/concert/friendly-thug-52-ngg-tour?source=selection-events&selectionPage=concert-hiphop&refererSelection=main_page_landing&refererSurface=selection'),
        ('Джейсон Деруло', 'jason-derulo-2026', 'ЦСКА Арена', 'Единственный концерт Джейсона Деруло в Москве: большое поп-шоу, мировые хиты, танцевальная постановка и яркая сцена в рамках европейского тура The Last Dance World Tour.', 16, 'https://afisha.yandex.ru/moscow/concert/jason-derulo-2026?city=moscow&rubricCode=concert&refererSurface=top-events')
) AS v(title, slug, location_description, full_description, age_limit, source_url)
WHERE u.email = 'seed.author@cityhawk.local'
ON CONFLICT (slug) DO UPDATE
SET title = EXCLUDED.title,
    location_description = EXCLUDED.location_description,
    full_description = EXCLUDED.full_description,
    age_limit = EXCLUDED.age_limit,
    source_url = EXCLUDED.source_url,
    updated_at = now();

-- 5. Привязка события к месту для карты
INSERT INTO event_place (event_id, place_id)
SELECT e.id, p.id
FROM (
    VALUES
        ('ruki-vverkh-30-nam-uzhe', 'Большая спортивная арена «Лужники»'),
        ('toxi-tour', 'VK Музыка Летом'),
        ('leonid-agutin-tour-2026', 'Большая спортивная арена «Лужники»'),
        ('bushido-zho-2026-06-27', 'TAU'),
        ('og-buda-tour-2026', 'Red Summer'),
        ('saluki-2026-06-28', 'VK Музыка Летом'),
        ('big-baby-tape-tour-2026', 'ВТБ Арена'),
        ('friendly-thug-52-ngg-tour-2026', 'ЦСКА Арена'),
        ('jason-derulo-2026', 'ЦСКА Арена')
) AS v(event_slug, place_name)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id) DO UPDATE
SET place_id = EXCLUDED.place_id,
    updated_at = now();

-- 6. Сессии концертов
INSERT INTO event_session (event_id, place_id, start_at, end_at, price)
SELECT e.id, p.id, v.start_at::timestamptz, v.end_at::timestamptz, v.price
FROM (
    VALUES
        ('ruki-vverkh-30-nam-uzhe', 'Большая спортивная арена «Лужники»', '2026-07-25 20:00:00+03', '2026-07-25 23:00:00+03', 0),
        ('ruki-vverkh-30-nam-uzhe', 'Большая спортивная арена «Лужники»', '2026-07-26 20:00:00+03', '2026-07-26 23:00:00+03', 0),
        ('toxi-tour', 'VK Музыка Летом', '2026-06-11 17:00:00+03', '2026-06-11 22:00:00+03', 0),
        ('leonid-agutin-tour-2026', 'Большая спортивная арена «Лужники»', '2026-07-18 19:00:00+03', '2026-07-18 22:00:00+03', 0),
        ('bushido-zho-2026-06-27', 'TAU', '2026-06-27 18:00:00+03', '2026-06-27 21:30:00+03', 0),
        ('og-buda-tour-2026', 'Red Summer', '2026-07-16 20:00:00+03', '2026-07-16 23:00:00+03', 0),
        ('saluki-2026-06-28', 'VK Музыка Летом', '2026-06-28 17:00:00+03', '2026-06-28 22:00:00+03', 0),
        ('big-baby-tape-tour-2026', 'ВТБ Арена', '2026-05-30 17:00:00+03', '2026-05-30 21:00:00+03', 0),
        ('friendly-thug-52-ngg-tour-2026', 'ЦСКА Арена', '2026-12-11 19:00:00+03', '2026-12-11 22:00:00+03', 0),
        ('jason-derulo-2026', 'ЦСКА Арена', '2026-06-14 18:00:00+03', '2026-06-14 21:00:00+03', 0)
) AS v(event_slug, place_name, start_at, end_at, price)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id, place_id, start_at) DO UPDATE
SET end_at = EXCLUDED.end_at,
    price = EXCLUDED.price,
    updated_at = now();

-- 7. Изображения событий
INSERT INTO event_image (event_id, image_url)
SELECT e.id, v.image_url
FROM (
    VALUES
        -- Изображения из прямых media URL сохранены локально в uploads/events.
        ('ruki-vverkh-30-nam-uzhe', '/uploads/events/concert-1.jpg'),
        ('toxi-tour', '/uploads/events/concert-2.jpg'),
        ('jason-derulo-2026', '/uploads/events/jason-derulo-2026-01.jpg'),
        ('bushido-zho-2026-06-27', '/uploads/events/concert-4.jpg'),
        ('bushido-zho-2026-06-27', '/uploads/events/bushido-zho-2026-06-27-02.jpg'),

        -- Локальные изображения, найденные и сохраненные в uploads/events.
        ('leonid-agutin-tour-2026', '/uploads/events/concert-3.jpg'),
        ('og-buda-tour-2026', '/uploads/events/concert-5.jpg'),
        ('saluki-2026-06-28', '/uploads/events/concert-6.jpg'),
        ('big-baby-tape-tour-2026', '/uploads/events/concert-7.jpg'),
        ('friendly-thug-52-ngg-tour-2026', '/uploads/events/concert-8.jpg')
) AS v(event_slug, image_url)
JOIN event e ON e.slug = v.event_slug
ON CONFLICT (event_id, image_url) DO NOTHING;

-- 8. Категория «Концерты»
INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM event e
JOIN category c ON c.name = 'Концерты'
WHERE e.slug IN (
    'ruki-vverkh-30-nam-uzhe',
    'toxi-tour',
    'leonid-agutin-tour-2026',
    'bushido-zho-2026-06-27',
    'og-buda-tour-2026',
    'saluki-2026-06-28',
    'big-baby-tape-tour-2026',
    'friendly-thug-52-ngg-tour-2026',
    'jason-derulo-2026'
)
ON CONFLICT (event_id, category_id) DO NOTHING;

-- 9. Теги
INSERT INTO event_tag (event_id, tag_id)
SELECT e.id, t.id
FROM (
    VALUES
        ('ruki-vverkh-30-nam-uzhe', 'Концерт'), ('ruki-vverkh-30-nam-uzhe', 'Поп'), ('ruki-vverkh-30-nam-uzhe', 'Стадион'), ('ruki-vverkh-30-nam-uzhe', 'Большое шоу'),
        ('toxi-tour', 'Концерт'), ('toxi-tour', 'Хип-хоп'), ('toxi-tour', 'Рэп'), ('toxi-tour', 'На воздухе'), ('toxi-tour', 'Летний концерт'),
        ('leonid-agutin-tour-2026', 'Концерт'), ('leonid-agutin-tour-2026', 'Поп'), ('leonid-agutin-tour-2026', 'Стадион'), ('leonid-agutin-tour-2026', 'Тур'),
        ('bushido-zho-2026-06-27', 'Концерт'), ('bushido-zho-2026-06-27', 'Хип-хоп'), ('bushido-zho-2026-06-27', 'Рэп'), ('bushido-zho-2026-06-27', 'Большое шоу'),
        ('og-buda-tour-2026', 'Концерт'), ('og-buda-tour-2026', 'Хип-хоп'), ('og-buda-tour-2026', 'Рэп'), ('og-buda-tour-2026', 'На воздухе'),
        ('saluki-2026-06-28', 'Концерт'), ('saluki-2026-06-28', 'Хип-хоп'), ('saluki-2026-06-28', 'Рэп'), ('saluki-2026-06-28', 'На воздухе'),
        ('big-baby-tape-tour-2026', 'Концерт'), ('big-baby-tape-tour-2026', 'Хип-хоп'), ('big-baby-tape-tour-2026', 'Рэп'), ('big-baby-tape-tour-2026', 'Большое шоу'),
        ('friendly-thug-52-ngg-tour-2026', 'Концерт'), ('friendly-thug-52-ngg-tour-2026', 'Хип-хоп'), ('friendly-thug-52-ngg-tour-2026', 'Рэп'), ('friendly-thug-52-ngg-tour-2026', 'Тур'),
        ('jason-derulo-2026', 'Концерт'), ('jason-derulo-2026', 'Поп'), ('jason-derulo-2026', 'Большое шоу'), ('jason-derulo-2026', 'Тур')
) AS v(event_slug, tag_name)
JOIN event e ON e.slug = v.event_slug
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;

COMMIT;
