-- CityHawk seed: реальные события с вайб-тегами
-- События из актуальной афиши Москвы. Каждое событие связано хотя бы с одним тегом из tag_group = 'mood'.

BEGIN;

-- 0. Безопасные доработки под seed
ALTER TABLE event
    ADD COLUMN IF NOT EXISTS slug text;

CREATE UNIQUE INDEX IF NOT EXISTS event_slug_key
    ON event(slug);

ALTER TABLE tag
    ADD COLUMN IF NOT EXISTS tag_group text;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'tag_group_valid'
    ) THEN
        ALTER TABLE tag
            ADD CONSTRAINT tag_group_valid
            CHECK (
                tag_group IS NULL
                OR tag_group IN ('format', 'genre', 'mood', 'audience', 'price')
            );
    END IF;
END $$;

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

CREATE UNIQUE INDEX IF NOT EXISTS event_tag_event_id_tag_id_key
    ON event_tag(event_id, tag_id);

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
    ('Фестивали'),
    ('Концерты'),
    ('Выставки'),
    ('Мюзиклы')
ON CONFLICT (name) DO NOTHING;

INSERT INTO tag (name, tag_group)
VALUES
    ('Движение', 'mood'),
    ('Ночная жизнь', 'mood'),
    ('Для фото', 'mood'),
    ('Спокойный вечер', 'mood'),
    ('Иммерсивный', 'mood'),
    ('Свидание', 'mood'),
    ('Фестиваль', 'format'),
    ('Концерт', 'format'),
    ('Выставка', 'format'),
    ('Мюзикл', 'format'),
    ('Open air', 'format'),
    ('Музыка', 'genre'),
    ('Рок', 'genre'),
    ('Живая музыка', 'genre'),
    ('Цветы', 'genre'),
    ('Искусство', 'genre'),
    ('Гастрономия', 'genre'),
    ('Для компании', 'audience'),
    ('Для двоих', 'audience'),
    ('Семейный', 'audience')
ON CONFLICT (name) DO UPDATE
SET tag_group = EXCLUDED.tag_group,
    updated_at = now();

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
            'Музей-заповедник «Коломенское»',
            'Москва, просп. Андропова, 39',
            55.672016::numeric,
            37.665010::numeric,
            'Исторический музей-заповедник и большая зеленая территория на юге Москвы, где проходят городские фестивали, концерты и культурные события под открытым небом.'
        ),
        (
            'Зеленый театр ВДНХ',
            'Москва, просп. Мира, 119, стр. 545',
            55.832700::numeric,
            37.629300::numeric,
            'Летняя концертная площадка на территории ВДНХ для крупных концертов, фестивальных программ и вечерних выступлений под открытым небом.'
        ),
        (
            'Мраморный зал Павильона Республики Казахстан на ВДНХ',
            'Москва, просп. Мира, 119, стр. 11',
            55.832000::numeric,
            37.622900::numeric,
            'Камерный зал в павильоне Казахстан на ВДНХ для концертов, культурных встреч и национальных программ.'
        ),
        (
            'Корпус на Кадашёвской набережной',
            'Москва, Кадашёвская наб., 12',
            55.742600::numeric,
            37.624500::numeric,
            'Выставочный корпус Третьяковской галереи рядом с Лаврушинским переулком, где проходят временные художественные проекты.'
        ),
        (
            'Театр «Одеон»',
            'Москва, ул. 3-я Ямского Поля, 15',
            55.783200::numeric,
            37.588800::numeric,
            'Площадка гастрономических музыкальных шоу, где театральное действие совмещается с ужином и вечерним форматом отдыха.'
        )
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
        (
            'Пикник Афиши 2026',
            'piknik-afishi-2026-moscow',
            'Музей-заповедник «Коломенское»',
            'Летний музыкальный фестиваль в Коломенском: концерты, арт-инсталляции, гастрономия, маркет и большой городской open-air формат.',
            0,
            'https://piknikafishi.ru/'
        ),
        (
            'NAUTILUS POMPILIUS. Вячеслав Бутусов. Летний концерт',
            'nautilus-pompilius-butusov-vdnh-2026',
            'Зеленый театр ВДНХ',
            'Легендарные песни Nautilus Pompilius в летнем формате под открытым небом на исторической сцене Зеленого театра ВДНХ.',
            12,
            'https://redkassa.ru/events/bilety_na_letniy_concert_naytilus_pompiliys_vyacheslav_butusov'
        ),
        (
            'Концерт Александра Фролова',
            'alexander-frolov-vdnh-2026',
            'Мраморный зал Павильона Республики Казахстан на ВДНХ',
            'Камерный сольный концерт Александра Фролова с песнями на русском и казахском языке, живым звучанием и специальными гостями.',
            12,
            'https://vdnh.ru/events/3318/'
        ),
        (
            'Выставка «Цветы. Символ красоты»',
            'tsvety-simvol-krasoty-tretyakovka-2026',
            'Корпус на Кадашёвской набережной',
            'Выставка Третьяковской галереи о флоральных образах в русском искусстве: натюрморты, пейзажи и цветочные мотивы конца XIX — начала XX века.',
            0,
            'https://www.tretyakovgallery.ru/exhibitions/o/tsvety-simvol-krasoty/'
        ),
        (
            'Гастрономический мюзикл «Москва от заката до рассвета»',
            'moskva-ot-zakata-do-rassveta-odeon-2026',
            'Театр «Одеон»',
            'Вечерний гастрономический мюзикл о Москве, где музыкальное шоу, ужин и мистическая история собираются в один иммерсивный формат.',
            12,
            'https://theatreodeon.ru/places/543'
        )
) AS v(title, slug, location_description, full_description, age_limit, source_url)
WHERE u.email = 'seed.author@cityhawk.local'
ON CONFLICT (slug) DO UPDATE
SET title = EXCLUDED.title,
    location_description = EXCLUDED.location_description,
    full_description = EXCLUDED.full_description,
    age_limit = EXCLUDED.age_limit,
    source_url = EXCLUDED.source_url,
    updated_at = now();

-- 5. Привязка событий к местам
INSERT INTO event_place (event_id, place_id)
SELECT e.id, p.id
FROM (
    VALUES
        ('piknik-afishi-2026-moscow', 'Музей-заповедник «Коломенское»'),
        ('nautilus-pompilius-butusov-vdnh-2026', 'Зеленый театр ВДНХ'),
        ('alexander-frolov-vdnh-2026', 'Мраморный зал Павильона Республики Казахстан на ВДНХ'),
        ('tsvety-simvol-krasoty-tretyakovka-2026', 'Корпус на Кадашёвской набережной'),
        ('moskva-ot-zakata-do-rassveta-odeon-2026', 'Театр «Одеон»')
) AS v(event_slug, place_name)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id) DO UPDATE
SET place_id = EXCLUDED.place_id,
    updated_at = now();

-- 6. Сессии
INSERT INTO event_session (event_id, place_id, start_at, end_at, price)
SELECT e.id, p.id, v.start_at::timestamptz, v.end_at::timestamptz, v.price
FROM (
    VALUES
        ('piknik-afishi-2026-moscow', 'Музей-заповедник «Коломенское»', '2026-06-20 14:00:00+03', '2026-06-20 23:00:00+03', 0),
        ('nautilus-pompilius-butusov-vdnh-2026', 'Зеленый театр ВДНХ', '2026-06-26 20:00:00+03', '2026-06-26 22:30:00+03', 4800),
        ('alexander-frolov-vdnh-2026', 'Мраморный зал Павильона Республики Казахстан на ВДНХ', '2026-06-12 19:00:00+03', '2026-06-12 21:30:00+03', 0),
        ('tsvety-simvol-krasoty-tretyakovka-2026', 'Корпус на Кадашёвской набережной', '2026-06-02 10:00:00+03', '2026-06-02 18:00:00+03', 900),
        ('moskva-ot-zakata-do-rassveta-odeon-2026', 'Театр «Одеон»', '2026-06-30 20:00:00+03', '2026-06-30 23:00:00+03', 0)
) AS v(event_slug, place_name, start_at, end_at, price)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id, place_id, start_at) DO UPDATE
SET end_at = EXCLUDED.end_at,
    price = EXCLUDED.price,
    updated_at = now();

-- 7. Картинки событий. Используются существующие локальные ассеты из app/uploads.
DELETE FROM event_image ei
USING event e
WHERE ei.event_id = e.id
  AND e.slug IN (
      'piknik-afishi-2026-moscow',
      'nautilus-pompilius-butusov-vdnh-2026',
      'alexander-frolov-vdnh-2026',
      'tsvety-simvol-krasoty-tretyakovka-2026',
      'moskva-ot-zakata-do-rassveta-odeon-2026'
  );

INSERT INTO event_image (event_id, image_url)
SELECT e.id, v.image_url
FROM (
    VALUES
        ('piknik-afishi-2026-moscow', '/uploads/piknik-afishi-1.png'),
        ('nautilus-pompilius-butusov-vdnh-2026', '/uploads/concert-10.jpg'),
        ('alexander-frolov-vdnh-2026', '/uploads/concert-11.jpg'),
        ('tsvety-simvol-krasoty-tretyakovka-2026', '/uploads/tsvety-simvol-1.jpg'),
        ('moskva-ot-zakata-do-rassveta-odeon-2026', '/uploads/moskva-ot-zakata-do-rassveta-1.webp')
) AS v(event_slug, image_url)
JOIN event e ON e.slug = v.event_slug
ON CONFLICT (event_id, image_url) DO NOTHING;

-- 8. Категории событий
INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM (
    VALUES
        ('piknik-afishi-2026-moscow', 'Фестивали'),
        ('nautilus-pompilius-butusov-vdnh-2026', 'Концерты'),
        ('alexander-frolov-vdnh-2026', 'Концерты'),
        ('tsvety-simvol-krasoty-tretyakovka-2026', 'Выставки'),
        ('moskva-ot-zakata-do-rassveta-odeon-2026', 'Мюзиклы')
) AS v(event_slug, category_name)
JOIN event e ON e.slug = v.event_slug
JOIN category c ON c.name = v.category_name
ON CONFLICT (event_id, category_id) DO NOTHING;

-- 9. Теги событий, включая mood-теги для выдачи по вайбу
INSERT INTO event_tag (event_id, tag_id)
SELECT e.id, t.id
FROM (
    VALUES
        ('piknik-afishi-2026-moscow', 'Движение'),
        ('piknik-afishi-2026-moscow', 'Фестиваль'),
        ('piknik-afishi-2026-moscow', 'Open air'),
        ('piknik-afishi-2026-moscow', 'Музыка'),
        ('piknik-afishi-2026-moscow', 'Для компании'),
        ('piknik-afishi-2026-moscow', 'Семейный'),

        ('nautilus-pompilius-butusov-vdnh-2026', 'Движение'),
        ('nautilus-pompilius-butusov-vdnh-2026', 'Концерт'),
        ('nautilus-pompilius-butusov-vdnh-2026', 'Open air'),
        ('nautilus-pompilius-butusov-vdnh-2026', 'Рок'),
        ('nautilus-pompilius-butusov-vdnh-2026', 'Музыка'),
        ('nautilus-pompilius-butusov-vdnh-2026', 'Для компании'),

        ('alexander-frolov-vdnh-2026', 'Спокойный вечер'),
        ('alexander-frolov-vdnh-2026', 'Концерт'),
        ('alexander-frolov-vdnh-2026', 'Живая музыка'),
        ('alexander-frolov-vdnh-2026', 'Музыка'),

        ('tsvety-simvol-krasoty-tretyakovka-2026', 'Для фото'),
        ('tsvety-simvol-krasoty-tretyakovka-2026', 'Спокойный вечер'),
        ('tsvety-simvol-krasoty-tretyakovka-2026', 'Выставка'),
        ('tsvety-simvol-krasoty-tretyakovka-2026', 'Цветы'),
        ('tsvety-simvol-krasoty-tretyakovka-2026', 'Искусство'),

        ('moskva-ot-zakata-do-rassveta-odeon-2026', 'Иммерсивный'),
        ('moskva-ot-zakata-do-rassveta-odeon-2026', 'Свидание'),
        ('moskva-ot-zakata-do-rassveta-odeon-2026', 'Ночная жизнь'),
        ('moskva-ot-zakata-do-rassveta-odeon-2026', 'Мюзикл'),
        ('moskva-ot-zakata-do-rassveta-odeon-2026', 'Гастрономия'),
        ('moskva-ot-zakata-do-rassveta-odeon-2026', 'Для двоих')
) AS v(event_slug, tag_name)
JOIN event e ON e.slug = v.event_slug
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;

COMMIT;
