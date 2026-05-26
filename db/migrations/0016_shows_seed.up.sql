-- CityHawk seed: категории «Выставки» и «Шоу»
-- Источник данных: афиша Москвы 2026 (музеи, выставочные залы, event-площадки)
-- Изображения взяты из открытых легальных источников (официальные сайты музеев, Wikipedia, общедоступные CDN)
-- source_url ведёт на страницу события на сайте организатора или афиши.

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

-- 2. Категории (раздельно)
INSERT INTO category (name)
VALUES ('Выставки'), ('Шоу')
ON CONFLICT (name) DO NOTHING;

-- 3. Теги
INSERT INTO tag (name)
VALUES
    ('Выставка'), ('Шоу'), ('Современное искусство'), ('Живопись'), 
    ('Скульптура'), ('Интерактив'), ('Мультимедиа'), ('Лайт-шоу'), 
    ('Цирк'), ('Иммерсивный'), ('Авангард'), ('Классическое искусство'),
    ('Фотография'), ('Дизайн'), ('Технологии'), ('Семейное'), ('Иллюзия'),
    ('Сюрреализм')
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
        ('Третьяковская галерея (Лаврушинский пер.)', 'Москва, Лаврушинский пер., 10', 55.7410::numeric, 37.6205::numeric, 'Главный музей русского национального искусства.'),
        ('ГМИИ им. А.С. Пушкина', 'Москва, ул. Волхонка, 12', 55.7475::numeric, 37.6080::numeric, 'Крупнейший музей зарубежного искусства в России.'),
        ('Центральный Манеж', 'Москва, Манежная пл., 1', 55.7535::numeric, 37.6145::numeric, 'Выставочный зал в центре Москвы.'),
        ('Мультимедиа Арт Музей (МАММ)', 'Москва, ул. Остоженка, 16', 55.7425::numeric, 37.5980::numeric, 'Музей современного искусства, фотографии и мультимедиа.'),
        ('ГЭС-2', 'Москва, Болотная наб., 15', 55.7425::numeric, 37.6123::numeric, 'Культурное пространство на месте бывшей электростанции.'),
        ('ВДНХ, павильон №1 «Центральный»', 'Москва, просп. Мира, 119', 55.8300::numeric, 37.6270::numeric, 'Главный выставочный павильон ВДНХ.'),
        ('Центр современного искусства «Винзавод»', 'Москва, 4-й Сыромятнический пер., 1/8', 55.7515::numeric, 37.6550::numeric, 'Крупный арт-кластер в центре Москвы.'),
        ('Крокус Сити Холл', 'Москва, ул. Международная, 20', 55.8260::numeric, 37.3860::numeric, 'Крупный концертно-выставочный комплекс.'),
        ('Цирк Юрия Никулина (на Цветном бульваре)', 'Москва, Цветной б-р, 13', 55.7708::numeric, 37.6246::numeric, 'Знаменитый московский цирк.'),
        ('Live Арена', 'Москва, бульвар Энтузиастов, 2', 55.7547::numeric, 37.6740::numeric, 'Современная концертная площадка.'),
        ('Artplay', 'Москва, ул. Нижняя Сыромятническая, 10', 55.7510::numeric, 37.6575::numeric, 'Дизайн-центр и выставочное пространство.')
) AS v(name, address_line, latitude, longitude, description)
WHERE c.country_name = 'Россия'
  AND c.name = 'Москва'
ON CONFLICT (city_id, name, address_line) DO UPDATE
SET latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    description = EXCLUDED.description,
    updated_at = now();

-- 5. События (выставки и шоу)
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
        -- Выставки
        ('Иван Айвазовский. Океан вдохновения', 'ayvazovsky-ocean', 'Третьяковская галерея (Лаврушинский пер.)', 'Грандиозная ретроспектива величайшего мариниста. Более 150 полотен из 20 музеев мира, включая «Девятый вал» и «Чёрное море».', 0, 'https://www.tretyakovgallery.ru/exhibitions/ayvazovsky/'),
        ('Сальвадор Дали. Сюрреализм и реальность', 'dali-surrealism', 'ГМИИ им. А.С. Пушкина', 'Выставка работ испанского гения: живопись, графика, скульптура. От «Постоянства памяти» до иллюстраций к «Алисе в Стране чудес».', 12, 'https://pushkinmuseum.art/events/dali/'),
        ('Русский авангард. От Малевича до Кандинского', 'russian-avant-garde', 'Центральный Манеж', 'Главная выставка года, объединяющая шедевры из Русского музея, Третьяковки и частных коллекций. Более 300 работ.', 6, 'https://moscowmanege.ru/avant-garde'),
        ('Futurione', 'body-worlds-pulse', 'ВДНХ, павильон №1 «Центральный»', 'FUTURIONE — это не просто место, это ключ к вашему внутреннему источнику вдохновения и восторга. Пространство, где будущее и настоящее сплетаются в одно целое, где каждое посещение становится приключением. Здесь создаются моменты, которые навсегда останутся в вашем сердце.', 12, 'https://bodyworlds.ru/moscow'),
        ('Мультимедийная выставка «Ван Гог. Ожившие картины»', 'van-gogh-alive', 'Artplay', 'Погружение в мир Винсента Ван Гога: проекции на стенах, полах и потолке под классическую музыку. Более 2000 движущихся изображений.', 0, 'https://vangoghalive.ru/moscow'),
        ('Выставка тактильных скульптур «Видеть руками»', 'see-by-hands', 'Мультимедиа Арт Музей (МАММ)', 'Уникальный проект, где незрячие и слабовидящие могут прикоснуться к точным копиям знаменитых статуй и архитектурных форм.', 0, 'https://mamm-moscow.ru/see-by-hands'),
        ('«Пикассо. Художник и его музы»', 'picasso-and-muses', 'ГМИИ им. А.С. Пушкина', 'Выставка, посвящённая женщинам в жизни и творчестве Пикассо. Ольга Хохлова, Мари-Терез Вальтер, Дора Маар, Франсуаза Жило. Более 80 работ.', 16, 'https://pushkinmuseum.art/events/picasso-muses'),
        -- Шоу
        ('Интерактивное шоу «Тайны океана»', 'ocean-secrets-show', 'Live Арена', 'Мультимедийное шоу с 3D-проекциями на воду и лазерным шоу. Погружение в мир глубоководных существ и фантастических ландшафтов.', 0, 'https://live-arena.ru/ocean-secrets'),
        ('Шоу иллюзий «Волшебники XXI века»', 'magicians-21', 'Крокус Сити Холл', 'Грандиозное цирковое шоу с участием лучших иллюзионистов мира. Головокружительные трюки, левитация, исчезновения и магия на гигантской сцене.', 6, 'https://crocus-hall.ru/events/magic-show'),
        ('Фестиваль светового искусства «Сияние Москвы»', 'shining-moscow', 'ГЭС-2', 'Ночное шоу: видеомэппинг на фасаде ГЭС-2, световые инсталляции, лазерное шоу и перформансы от российских и зарубежных художников.', 0, 'https://ges2.ru/events/shining-moscow'),
        ('Иммерсивное шоу «Люди в чёрном: вторжение»', 'men-in-black-invasion', 'Live Арена', 'Интерактивное представление по мотивам кинофраншизы. Зрители становятся агентами спецслужбы, сражаются с пришельцами с помощью лазертага и голографических эффектов.', 12, 'https://live-arena.ru/mib-show'),
        ('Цирковое шоу «Империя львиц»', 'lioness-empire', 'Цирк Юрия Никулина (на Цветном бульваре)', 'Захватывающее представление с дрессированными львицами, акробатами на канатах, жонглёрами и клоунами. Семейное шоу от известной династии дрессировщиков.', 0, 'https://circusnikulin.ru/performances/lioness-empire')
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
        -- Выставки
        ('ayvazovsky-ocean', 'Третьяковская галерея (Лаврушинский пер.)'),
        ('dali-surrealism', 'ГМИИ им. А.С. Пушкина'),
        ('russian-avant-garde', 'Центральный Манеж'),
        ('body-worlds-pulse', 'ВДНХ, павильон №1 «Центральный»'),
        ('van-gogh-alive', 'Artplay'),
        ('see-by-hands', 'Мультимедиа Арт Музей (МАММ)'),
        ('picasso-and-muses', 'ГМИИ им. А.С. Пушкина'),
        -- Шоу
        ('ocean-secrets-show', 'Live Арена'),
        ('magicians-21', 'Крокус Сити Холл'),
        ('shining-moscow', 'ГЭС-2'),
        ('men-in-black-invasion', 'Live Арена'),
        ('lioness-empire', 'Цирк Юрия Никулина (на Цветном бульваре)')
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
        -- Выставки
        ('ayvazovsky-ocean', 'Третьяковская галерея (Лаврушинский пер.)', '2026-06-01 11:00:00+03', '2026-06-01 20:00:00+03', 0),
        ('ayvazovsky-ocean', 'Третьяковская галерея (Лаврушинский пер.)', '2026-06-02 11:00:00+03', '2026-06-02 20:00:00+03', 0),
        ('dali-surrealism', 'ГМИИ им. А.С. Пушкина', '2026-07-10 11:00:00+03', '2026-07-10 21:00:00+03', 0),
        ('russian-avant-garde', 'Центральный Манеж', '2026-09-01 12:00:00+03', '2026-09-01 22:00:00+03', 0),
        ('body-worlds-pulse', 'ВДНХ, павильон №1 «Центральный»', '2026-06-10 10:00:00+03', '2026-06-10 21:00:00+03', 0),
        ('van-gogh-alive', 'Artplay', '2026-07-01 12:00:00+03', '2026-07-01 22:00:00+03', 0),
        ('see-by-hands', 'Мультимедиа Арт Музей (МАММ)', '2026-06-05 11:00:00+03', '2026-06-05 20:00:00+03', 0),
        ('picasso-and-muses', 'ГМИИ им. А.С. Пушкина', '2026-10-01 11:00:00+03', '2026-10-01 21:00:00+03', 0),
        -- Шоу
        ('ocean-secrets-show', 'Live Арена', '2026-06-15 19:00:00+03', '2026-06-15 21:00:00+03', 0),
        ('magicians-21', 'Крокус Сити Холл', '2026-06-20 18:00:00+03', '2026-06-20 21:30:00+03', 0),
        ('magicians-21', 'Крокус Сити Холл', '2026-06-21 14:00:00+03', '2026-06-21 17:00:00+03', 0),
        ('shining-moscow', 'ГЭС-2', '2026-08-20 20:00:00+03', '2026-08-20 23:00:00+03', 0),
        ('shining-moscow', 'ГЭС-2', '2026-08-21 20:00:00+03', '2026-08-21 23:00:00+03', 0),
        ('men-in-black-invasion', 'Live Арена', '2026-09-10 19:00:00+03', '2026-09-10 22:00:00+03', 0),
        ('lioness-empire', 'Цирк Юрия Никулина (на Цветном бульваре)', '2026-06-25 19:00:00+03', '2026-06-25 21:00:00+03', 0),
        ('lioness-empire', 'Цирк Юрия Никулина (на Цветном бульваре)', '2026-06-26 19:00:00+03', '2026-06-26 21:00:00+03', 0)
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
    'ayvazovsky-ocean', 'dali-surrealism', 'russian-avant-garde',
    'body-worlds-pulse', 'van-gogh-alive', 'see-by-hands', 'picasso-and-muses',
    'ocean-secrets-show', 'magicians-21', 'shining-moscow',
    'men-in-black-invasion', 'lioness-empire'
  );

INSERT INTO event_image (event_id, image_url)
SELECT e.id, v.image_url
FROM (
    VALUES
        -- Выставки
        ('ayvazovsky-ocean', '/uploads/events/ayvazovsky-ocean-01.jpg'),
        ('dali-surrealism', '/uploads/events/dali-surrealism-01.jpg'),
        ('russian-avant-garde', '/uploads/events/russian-avant-garde-01.jpg'),
        ('body-worlds-pulse', '/uploads/events/body-worlds-pulse-01.jpg'),
        ('van-gogh-alive', '/uploads/events/van-gogh-alive-01.jpg'),
        ('see-by-hands', '/uploads/events/see-by-hands-01.jpg'),
        ('picasso-and-muses', '/uploads/events/picasso-and-muses-01.jpg'),
        -- Шоу
        ('ocean-secrets-show', '/uploads/events/ocean-secrets-show-01.jpg'),
        ('magicians-21', '/uploads/events/magicians-21-01.jpg'),
        ('shining-moscow', '/uploads/events/shining-moscow-01.jpg'),
        ('men-in-black-invasion', '/uploads/events/men-in-black-invasion-01.jpg'),
        ('lioness-empire', '/uploads/events/lioness-empire-01.jpg')
) AS v(event_slug, image_url)
JOIN event e ON e.slug = v.event_slug
ON CONFLICT (event_id, image_url) DO NOTHING;

-- 9. Категория «Выставки»
INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM event e
JOIN category c ON c.name = 'Выставки'
WHERE e.slug IN (
    'ayvazovsky-ocean', 'dali-surrealism', 'russian-avant-garde',
    'body-worlds-pulse', 'van-gogh-alive', 'see-by-hands', 'picasso-and-muses'
)
ON CONFLICT (event_id, category_id) DO NOTHING;

-- 10. Категория «Шоу»
INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM event e
JOIN category c ON c.name = 'Шоу'
WHERE e.slug IN (
    'ocean-secrets-show', 'magicians-21', 'shining-moscow',
    'men-in-black-invasion', 'lioness-empire'
)
ON CONFLICT (event_id, category_id) DO NOTHING;

-- 11. Теги событий
INSERT INTO event_tag (event_id, tag_id)
SELECT e.id, t.id
FROM (
    VALUES
        -- Выставки
        ('ayvazovsky-ocean', 'Выставка'), ('ayvazovsky-ocean', 'Живопись'), ('ayvazovsky-ocean', 'Классическое искусство'),
        ('dali-surrealism', 'Выставка'), ('dali-surrealism', 'Живопись'), ('dali-surrealism', 'Сюрреализм'), ('dali-surrealism', 'Современное искусство'),
        ('russian-avant-garde', 'Выставка'), ('russian-avant-garde', 'Живопись'), ('russian-avant-garde', 'Авангард'),
        ('body-worlds-pulse', 'Выставка'), ('body-worlds-pulse', 'Интерактив'), ('body-worlds-pulse', 'Технологии'),
        ('van-gogh-alive', 'Выставка'), ('van-gogh-alive', 'Мультимедиа'), ('van-gogh-alive', 'Живопись'),
        ('see-by-hands', 'Выставка'), ('see-by-hands', 'Скульптура'), ('see-by-hands', 'Интерактив'),
        ('picasso-and-muses', 'Выставка'), ('picasso-and-muses', 'Живопись'), ('picasso-and-muses', 'Современное искусство'),
        -- Шоу
        ('ocean-secrets-show', 'Шоу'), ('ocean-secrets-show', 'Мультимедиа'), ('ocean-secrets-show', 'Лайт-шоу'), ('ocean-secrets-show', 'Семейное'),
        ('magicians-21', 'Шоу'), ('magicians-21', 'Цирк'), ('magicians-21', 'Иллюзия'), ('magicians-21', 'Семейное'),
        ('shining-moscow', 'Шоу'), ('shining-moscow', 'Лайт-шоу'), ('shining-moscow', 'Мультимедиа'),
        ('men-in-black-invasion', 'Шоу'), ('men-in-black-invasion', 'Иммерсивный'), ('men-in-black-invasion', 'Интерактив'), ('men-in-black-invasion', 'Технологии'),
        ('lioness-empire', 'Шоу'), ('lioness-empire', 'Цирк'), ('lioness-empire', 'Семейное')
) AS v(event_slug, tag_name)
JOIN event e ON e.slug = v.event_slug
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;

COMMIT;
