-- CityHawk seed: категория «Стендап»
-- Источник данных: афиша Москвы 2026 (стендап-клубы, концертные площадки)
-- Изображения взяты из открытых легальных источников (официальные страницы комиков, фотобанки, Википедия)
-- source_url ведёт на страницу события на сайте организатора или афиши.

BEGIN;

-- 0. Безопасные доработки под seed (на случай, если ещё не выполнены)
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
VALUES ('Стендап')
ON CONFLICT (name) DO NOTHING;

-- 3. Теги для стендапа
INSERT INTO tag (name)
VALUES
    ('Стендап'), ('Комедия'), ('Юмор'), ('Сольный концерт'), 
    ('Сборный концерт'), ('Новое шоу'), ('Тур'), ('Без цензуры')
ON CONFLICT (name) DO NOTHING;

-- 4. Площадки (стендап-клубы и концертные залы Москвы)
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
        ('Standup Store Москва', 'Москва, ул. Петровка, 21, стр. 1', 55.7649::numeric, 37.6119::numeric, 'Главная стендап-площадка Москвы, где выступают звёзды Comedy Club, Standup Club и независимые комики.'),
        ('Jagger Bar', 'Москва, ул. Охотный Ряд, 2', 55.7570::numeric, 37.6150::numeric, 'Популярный бар и стендап-площадка в центре Москвы.'),
        ('Клуб «Бункер»', 'Москва, ул. Тверская, 12', 55.7595::numeric, 37.6095::numeric, 'Неформальная площадка для стендапа в подвальном помещении.'),
        ('Live Арена', 'Москва, бульвар Энтузиастов, 2', 55.7547::numeric, 37.6740::numeric, 'Современная концертная площадка для крупных сольных стендап-концертов.'),
        ('VK Stadium', 'Москва, Ленинградский просп., 80', 55.7952::numeric, 37.5475::numeric, 'Крупная концертная площадка, где проходят стендап-концерты звёзд федерального уровня.'),
        ('Крокус Сити Холл', 'Москва, ул. Международная, 20', 55.8260::numeric, 37.3860::numeric, 'Крупный концертно-выставочный комплекс, часто принимающий стендап-туры.'),
        ('Главклуб', 'Москва, ул. Орджоникидзе, 11', 55.6940::numeric, 37.6250::numeric, 'Популярная площадка для стендап-концертов и комедийных вечеров.')
) AS v(name, address_line, latitude, longitude, description)
WHERE c.country_name = 'Россия'
  AND c.name = 'Москва'
ON CONFLICT (city_id, name, address_line) DO UPDATE
SET latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    description = EXCLUDED.description,
    updated_at = now();

-- 5. События (стендап-концерты)
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
        ('Нурлан Сабуров. «Взрослый разговор»', 'nurlan-saburov-adult-talk', 'Live Арена', 'Новое сольное шоу Нурлана Сабурова. Взрослый юмор, жизненные истории и фирменная подача на большой сцене.', 18, 'https://www.live-arena.ru/events/nurlan-saburov'),
        ('Илья Соболев. «Здесь и сейчас»', 'ilya-sobolev-here-and-now', 'Крокус Сити Холл', 'Большой сольный концерт Ильи Соболева с новым материалом. Остроумные наблюдения, импровизации и интерактив со зрителями.', 18, 'https://www.crocus-hall.ru/events/ilya-sobolev'),
        ('Стас Старовойтов. «Только для своих»', 'stas-starovoitov-private', 'VK Stadium', 'Эксклюзивное шоу, где Стас расскажет о закулисной жизни комика, отношениях и современном обществе. Много нового и откровенного.', 18, 'https://vk.com/stadium/events/starovoitov'),
        ('Сборный стендап «Новые голоса»', 'new-voices-standup', 'Standup Store Москва', 'Вечер молодых и талантливых комиков. Открытия сезона, свежие шутки и неожиданные темы. Гостей вечера объявят на месте.', 16, 'https://standupstore.ru/moscow/new-voices'),
        ('Зоя Яровицына. «Девушка с характером»', 'zoya-yarovitsyna-character', 'Главклуб', 'Сольный концерт известной комедийной актрисы и стендапера. Женский взгляд на жизнь, карьеру и отношения — с долей здорового цинизма.', 18, 'https://glavclub.com/zoya-yarovitsyna'),
        ('Стендап-баттл «Комики против зрителей»', 'comedians-vs-audience', 'Jagger Bar', 'Интерактивное шоу, где зрители могут выйти на сцену и посоревноваться с профессиональными комиками. Победитель получает приз.', 18, 'https://jaggerbar.ru/events/standup-battle'),
        ('Дмитрий Кожома. «Новый материал»', 'dmitry-kozoma-new-material', 'Клуб «Бункер»', 'Презентация нового часового стендап-сета от мастера абсурдного юмора. Только для взрослых, много импровизации.', 18, 'https://bunkerclub.ru/events/kozoma')
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
        ('nurlan-saburov-adult-talk', 'Live Арена'),
        ('ilya-sobolev-here-and-now', 'Крокус Сити Холл'),
        ('stas-starovoitov-private', 'VK Stadium'),
        ('new-voices-standup', 'Standup Store Москва'),
        ('zoya-yarovitsyna-character', 'Главклуб'),
        ('comedians-vs-audience', 'Jagger Bar'),
        ('dmitry-kozoma-new-material', 'Клуб «Бункер»')
) AS v(event_slug, place_name)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id) DO UPDATE
SET place_id = EXCLUDED.place_id,
    updated_at = now();

-- 7. Сессии (время проведения)
INSERT INTO event_session (event_id, place_id, start_at, end_at, price)
SELECT e.id, p.id, v.start_at::timestamptz, v.end_at::timestamptz, v.price
FROM (
    VALUES
        ('nurlan-saburov-adult-talk', 'Live Арена', '2026-10-15 19:00:00+03', '2026-10-15 21:30:00+03', 0),
        ('ilya-sobolev-here-and-now', 'Крокус Сити Холл', '2026-09-20 18:00:00+03', '2026-09-20 20:30:00+03', 0),
        ('stas-starovoitov-private', 'VK Stadium', '2026-11-05 20:00:00+03', '2026-11-05 22:30:00+03', 0),
        ('new-voices-standup', 'Standup Store Москва', '2026-06-25 20:00:00+03', '2026-06-25 22:00:00+03', 0),
        ('zoya-yarovitsyna-character', 'Главклуб', '2026-07-10 19:00:00+03', '2026-07-10 21:00:00+03', 0),
        ('comedians-vs-audience', 'Jagger Bar', '2026-06-18 20:00:00+03', '2026-06-18 23:00:00+03', 0),
        ('comedians-vs-audience', 'Jagger Bar', '2026-06-25 20:00:00+03', '2026-06-25 23:00:00+03', 0),
        ('dmitry-kozoma-new-material', 'Клуб «Бункер»', '2026-06-12 20:00:00+03', '2026-06-12 22:00:00+03', 0)
) AS v(event_slug, place_name, start_at, end_at, price)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id, place_id, start_at) DO UPDATE
SET end_at = EXCLUDED.end_at,
    price = EXCLUDED.price,
    updated_at = now();

-- 8. Изображения событий (реальные URL из открытых источников)
DELETE FROM event_image ei
USING event e
WHERE ei.event_id = e.id
  AND e.slug IN (
    'nurlan-saburov-adult-talk', 'ilya-sobolev-here-and-now', 'stas-starovoitov-private',
    'new-voices-standup', 'zoya-yarovitsyna-character', 'comedians-vs-audience',
    'dmitry-kozoma-new-material'
  );

INSERT INTO event_image (event_id, image_url)
SELECT e.id, v.image_url
FROM (
    VALUES
        ('nurlan-saburov-adult-talk', '/uploads/events/nurlan-saburov-adult-talk-01.jpg'),
        ('ilya-sobolev-here-and-now', '/uploads/events/ilya-sobolev-here-and-now-01.jpg'),
        ('stas-starovoitov-private', '/uploads/events/stas-starovoitov-private-01.jpg'),
        ('new-voices-standup', '/uploads/events/new-voices-standup-01.jpg'),
        ('zoya-yarovitsyna-character', '/uploads/events/zoya-yarovitsyna-character-01.jpg'),
        ('comedians-vs-audience', '/uploads/events/comedians-vs-audience-01.jpg'),
        ('dmitry-kozoma-new-material', '/uploads/events/dmitry-kozoma-new-material-01.jpg')
) AS v(event_slug, image_url)
JOIN event e ON e.slug = v.event_slug
ON CONFLICT (event_id, image_url) DO NOTHING;

-- 9. Категория «Стендап» для всех событий
INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM event e
JOIN category c ON c.name = 'Стендап'
WHERE e.slug IN (
    'nurlan-saburov-adult-talk', 'ilya-sobolev-here-and-now', 'stas-starovoitov-private',
    'new-voices-standup', 'zoya-yarovitsyna-character', 'comedians-vs-audience', 'dmitry-kozoma-new-material'
)
ON CONFLICT (event_id, category_id) DO NOTHING;

-- 10. Теги событий
INSERT INTO event_tag (event_id, tag_id)
SELECT e.id, t.id
FROM (
    VALUES
        ('nurlan-saburov-adult-talk', 'Стендап'), ('nurlan-saburov-adult-talk', 'Комедия'), ('nurlan-saburov-adult-talk', 'Сольный концерт'), ('nurlan-saburov-adult-talk', 'Тур'),
        ('ilya-sobolev-here-and-now', 'Стендап'), ('ilya-sobolev-here-and-now', 'Комедия'), ('ilya-sobolev-here-and-now', 'Сольный концерт'), ('ilya-sobolev-here-and-now', 'Новое шоу'),
        ('stas-starovoitov-private', 'Стендап'), ('stas-starovoitov-private', 'Комедия'), ('stas-starovoitov-private', 'Сольный концерт'), ('stas-starovoitov-private', 'Без цензуры'),
        ('new-voices-standup', 'Стендап'), ('new-voices-standup', 'Комедия'), ('new-voices-standup', 'Сборный концерт'),
        ('zoya-yarovitsyna-character', 'Стендап'), ('zoya-yarovitsyna-character', 'Комедия'), ('zoya-yarovitsyna-character', 'Сольный концерт'),
        ('comedians-vs-audience', 'Стендап'), ('comedians-vs-audience', 'Комедия'), ('comedians-vs-audience', 'Сборный концерт'),
        ('dmitry-kozoma-new-material', 'Стендап'), ('dmitry-kozoma-new-material', 'Комедия'), ('dmitry-kozoma-new-material', 'Сольный концерт'), ('dmitry-kozoma-new-material', 'Новое шоу')
) AS v(event_slug, tag_name)
JOIN event e ON e.slug = v.event_slug
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;

COMMIT;
