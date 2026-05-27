-- CityHawk seed: категория «Развлечения»
-- Данные подготовлены для реальных развлекательных мест в Москве.
-- Фото распределены по пользовательскому порядку снизу вверх: 1 — 3 фото, 2 — 4 фото, 3 — 2 фото, 4 — 4 фото, 5 — 4 фото, 6 удалено, 7 — 1 фото.

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

-- 2. Категория и теги
INSERT INTO category (name)
VALUES ('Развлечения')
ON CONFLICT (name) DO NOTHING;

INSERT INTO tag (name)
VALUES
    ('Развлечения'),
    ('Парк развлечений'),
    ('Аттракционы'),
    ('Семейный отдых'),
    ('Для детей'),
    ('Океанариум'),
    ('Шоу'),
    ('Колесо обозрения'),
    ('Панорама'),
    ('Профессии'),
    ('Музей'),
    ('Магия'),
    ('Боулинг'),
    ('Игровые автоматы'),
    ('Квест'),
    ('Для компании'),
    ('Выходные')
ON CONFLICT (name) DO NOTHING;

-- 2.1. Удаление события, которое было исключено из пользовательского списка
DELETE FROM event_image ei
USING event e
WHERE ei.event_id = e.id
  AND e.slug = 'cosmik-evropeyskiy-bowling-2026';

DELETE FROM event_category ec
USING event e
WHERE ec.event_id = e.id
  AND e.slug = 'cosmik-evropeyskiy-bowling-2026';

DELETE FROM event_tag et
USING event e
WHERE et.event_id = e.id
  AND e.slug = 'cosmik-evropeyskiy-bowling-2026';

DELETE FROM event_session es
USING event e
WHERE es.event_id = e.id
  AND e.slug = 'cosmik-evropeyskiy-bowling-2026';

DELETE FROM event_place ep
USING event e
WHERE ep.event_id = e.id
  AND e.slug = 'cosmik-evropeyskiy-bowling-2026';

DELETE FROM event
WHERE slug = 'cosmik-evropeyskiy-bowling-2026';

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
        ('Остров Мечты', 'Москва, просп. Андропова, 1', 55.694246::numeric, 37.674705::numeric, 'Крупный крытый тематический парк развлечений с аттракционами, городским променадом, ресторанами и зонами для семейного отдыха.'),
        ('Москвариум', 'Москва, ВДНХ, просп. Мира, 119, стр. 23', 55.832720::numeric, 37.619650::numeric, 'Центр океанографии и морской биологии на ВДНХ с аквариумом, водными шоу, образовательными программами и семейными маршрутами.'),
        ('Солнце Москвы', 'Москва, 2-я Останкинская ул., 3', 55.826089::numeric, 37.628936::numeric, 'Многофункциональный комплекс на ВДНХ с одним из самых высоких колёс обозрения, кафе, прогулочными зонами и панорамными кабинами.'),
        ('Кидзания', 'Москва, Ходынский бул., 4, ТЦ «Авиапарк», 4 этаж', 55.790231::numeric, 37.531289::numeric, 'Детский парк профессий, где ребёнок может попробовать разные роли, пройти игровые сценарии и получить опыт городских профессий.'),
        ('Музей Магии', 'Москва, ул. Новый Арбат, 7, стр. 1', 55.752572::numeric, 37.597549::numeric, 'Интерактивный музей о фокусах, иллюзиях и истории магического искусства с экспозицией и развлекательными программами.'),
        ('Квесты «Клаустрофобия»', 'Москва, Дмитровское ш., 29, корп. 1', 55.828051::numeric, 37.571965::numeric, 'Квест-пространство для командных игр, загадок, сюжетных сценариев и необычного досуга с друзьями.')
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
            'День в парке развлечений «Остров Мечты»',
            'ostrov-mechty-park-razvlecheniy-2026',
            'Остров Мечты',
            '«Остров Мечты» — один из самых известных крытых парков развлечений Москвы. Здесь можно провести целый день: пройти тематические зоны, покататься на аттракционах, прогуляться по городскому променаду, заглянуть в кафе и магазины. Формат подходит для семей, компаний друзей и тех, кто хочет устроить яркий выходной без привязки к погоде. Для гостей доступны разные зоны развлечений, фотолокации и маршруты для детей и взрослых.',
            0,
            'https://dreamisland.ru/'
        ),
        (
            'Прогулка по океанариуму «Москвариум»',
            'moskvarium-oceanarium-2026',
            'Москвариум',
            '«Москвариум» на ВДНХ — большой океанариум и центр морской биологии, где можно увидеть водных обитателей, пройти по экспозиции, посетить шоу или выбрать образовательную программу. Это спокойный, но впечатляющий формат досуга для семейного выходного, свидания или прогулки с друзьями. Маршрут подходит тем, кто любит животных, атмосферные пространства и познавательные развлечения.',
            0,
            'https://moskvarium.ru/'
        ),
        (
            'Панорамный подъём на колесе «Солнце Москвы»',
            'solntse-moskvy-wheel-2026',
            'Солнце Москвы',
            '«Солнце Москвы» — современное колесо обозрения на ВДНХ и всесезонная точка для прогулки, свидания или семейного досуга. Во время подъёма открывается панорама Москвы, а после можно провести время в городском пространстве комплекса, зайти в кафе или сделать фотографии у колеса. Формат подойдёт для короткого, но яркого развлечения в центре ВДНХ.',
            0,
            'https://moscow-sun.ru/'
        ),
        (
            'Детский парк профессий «Кидзания»',
            'kidzania-park-professiy-2026',
            'Кидзания',
            '«Кидзания» — интерактивный детский город профессий, где дети пробуют себя в разных ролях, выполняют игровые задания, знакомятся с устройством города и получают новый опыт через игру. Парк подойдёт для семейного дня, праздника или образовательного досуга. Здесь ребёнок может почувствовать себя пилотом, врачом, журналистом, спасателем и представителем других профессий.',
            4,
            'https://kidzaniamoscow.ru/'
        ),
        (
            'Интерактивный визит в Музей Магии',
            'muzey-magii-2026',
            'Музей Магии',
            'Музей Магии на Новом Арбате — развлекательное пространство о фокусах, иллюзиях и магическом искусстве. Гости знакомятся с экспозицией, необычными предметами, историей трюков и интерактивными элементами. Формат хорошо подходит для семей, туристов и компаний, которые хотят добавить в прогулку по центру Москвы немного загадочности и вау-эффекта.',
            6,
            'https://magicmuseum.ru/'
        ),
        (
            'Командный квест «Клаустрофобия»',
            'claustrophobia-quest-2026',
            'Квесты «Клаустрофобия»',
            'Квесты «Клаустрофобия» — командный формат развлечения, где участники погружаются в сюжет, ищут подсказки, решают загадки и проходят сценарий за ограниченное время. Такой досуг подходит для компании друзей, свидания, дня рождения или тимбилдинга. Главное в квесте — внимание к деталям, коммуникация и готовность действовать вместе.',
            12,
            'https://claustrophobia.com/moscow'
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
        ('ostrov-mechty-park-razvlecheniy-2026', 'Остров Мечты'),
        ('moskvarium-oceanarium-2026', 'Москвариум'),
        ('solntse-moskvy-wheel-2026', 'Солнце Москвы'),
        ('kidzania-park-professiy-2026', 'Кидзания'),
        ('muzey-magii-2026', 'Музей Магии'),
        ('claustrophobia-quest-2026', 'Квесты «Клаустрофобия»')
) AS v(event_slug, place_name)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id) DO UPDATE
SET place_id = EXCLUDED.place_id,
    updated_at = now();

-- 6. Сессии развлечений
INSERT INTO event_session (event_id, place_id, start_at, end_at, price)
SELECT e.id, p.id, v.start_at::timestamptz, v.end_at::timestamptz, v.price
FROM (
    VALUES
        ('ostrov-mechty-park-razvlecheniy-2026', 'Остров Мечты', '2026-06-06 12:00:00+03', '2026-06-06 22:00:00+03', 0),
        ('ostrov-mechty-park-razvlecheniy-2026', 'Остров Мечты', '2026-06-20 12:00:00+03', '2026-06-20 22:00:00+03', 0),

        ('moskvarium-oceanarium-2026', 'Москвариум', '2026-06-07 10:00:00+03', '2026-06-07 21:00:00+03', 0),
        ('moskvarium-oceanarium-2026', 'Москвариум', '2026-06-21 10:00:00+03', '2026-06-21 21:00:00+03', 0),

        ('solntse-moskvy-wheel-2026', 'Солнце Москвы', '2026-06-13 11:00:00+03', '2026-06-13 22:00:00+03', 0),
        ('solntse-moskvy-wheel-2026', 'Солнце Москвы', '2026-06-27 11:00:00+03', '2026-06-27 22:00:00+03', 0),

        ('kidzania-park-professiy-2026', 'Кидзания', '2026-06-14 12:00:00+03', '2026-06-14 20:00:00+03', 0),
        ('kidzania-park-professiy-2026', 'Кидзания', '2026-06-28 12:00:00+03', '2026-06-28 20:00:00+03', 0),

        ('muzey-magii-2026', 'Музей Магии', '2026-07-04 12:00:00+03', '2026-07-04 20:00:00+03', 0),
        ('muzey-magii-2026', 'Музей Магии', '2026-07-18 12:00:00+03', '2026-07-18 20:00:00+03', 0),


        ('claustrophobia-quest-2026', 'Квесты «Клаустрофобия»', '2026-07-11 15:00:00+03', '2026-07-11 17:00:00+03', 0),
        ('claustrophobia-quest-2026', 'Квесты «Клаустрофобия»', '2026-07-25 15:00:00+03', '2026-07-25 17:00:00+03', 0)
) AS v(event_slug, place_name, start_at, end_at, price)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id, place_id, start_at) DO UPDATE
SET end_at = EXCLUDED.end_at,
    price = EXCLUDED.price,
    updated_at = now();

-- 7. Изображения событий
DELETE FROM event_image ei
USING event e
WHERE ei.event_id = e.id
  AND e.slug IN (
    'ostrov-mechty-park-razvlecheniy-2026',
    'moskvarium-oceanarium-2026',
    'solntse-moskvy-wheel-2026',
    'kidzania-park-professiy-2026',
    'muzey-magii-2026',
    'claustrophobia-quest-2026'
  );

INSERT INTO event_image (event_id, image_url)
SELECT e.id, v.image_url
FROM (
    VALUES
        -- Порядок снизу вверх: 1. Остров Мечты — 3 фото
        ('ostrov-mechty-park-razvlecheniy-2026', '/uploads/events/2by0ra2shlrgx7gnwzvxa06qbi86qzwz.jpg.webp'),
        ('ostrov-mechty-park-razvlecheniy-2026', '/uploads/events/5fohv4n1xnsocy78et8ah6qekbxbd5bt.jpg.webp'),
        ('ostrov-mechty-park-razvlecheniy-2026', '/uploads/events/cm4nub6u9id20x3i6es7rsxdja4tw8cw.jpg.webp'),

        -- 2. Москвариум — 4 фото
        ('moskvarium-oceanarium-2026', '/uploads/events/moskvarium-165777big.jpg'),
        ('moskvarium-oceanarium-2026', '/uploads/events/unnamed.jpg'),
        ('moskvarium-oceanarium-2026', '/uploads/events/5xg2io3wocwqle3x1ksmi11i16ncz3r4.jpg'),
        ('moskvarium-oceanarium-2026', '/uploads/events/vip3_1.jpg'),

        -- 3. Солнце Москвы — 2 фото
        ('solntse-moskvy-wheel-2026', '/uploads/events/0015ca24a0cf20114417a61cf159d5e0.jpg'),
        ('solntse-moskvy-wheel-2026', '/uploads/events/88h90cc92z63yeeq98h51rb7hbjz1ba.jpg'),

        -- 4. Кидзания — 4 фото
        ('kidzania-park-professiy-2026', '/uploads/events/L_height-5.webp'),
        ('kidzania-park-professiy-2026', '/uploads/events/orig-4.jpeg'),
        ('kidzania-park-professiy-2026', '/uploads/events/kidzania-street.jpg'),
        ('kidzania-park-professiy-2026', '/uploads/events/6d2fc47b9dc79205ede39f847185bbf6.jpg'),

        -- 5. Музей Магии — 4 фото
        ('muzey-magii-2026', '/uploads/events/697352682d68a793f896c55e1d5107f2.jpg'),
        ('muzey-magii-2026', '/uploads/events/c76eccf46a6a978fa89855a3d7446d11.jpg'),
        ('muzey-magii-2026', '/uploads/events/57b662b70ed1238f09109fd4d9eee14e.jpg'),
        ('muzey-magii-2026', '/uploads/events/L_height-6.webp'),

        -- 6. Событие «Космик» удалено

        -- 7. Клаустрофобия — 1 фото
        ('claustrophobia-quest-2026', '/uploads/events/vu1121-2.jpeg')
) AS v(event_slug, image_url)
JOIN event e ON e.slug = v.event_slug
ON CONFLICT (event_id, image_url) DO NOTHING;

-- 8. Категория «Развлечения»
INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM event e
JOIN category c ON c.name = 'Развлечения'
WHERE e.slug IN (
    'ostrov-mechty-park-razvlecheniy-2026',
    'moskvarium-oceanarium-2026',
    'solntse-moskvy-wheel-2026',
    'kidzania-park-professiy-2026',
    'muzey-magii-2026',
    'claustrophobia-quest-2026'
)
ON CONFLICT (event_id, category_id) DO NOTHING;

-- 9. Теги событий
INSERT INTO event_tag (event_id, tag_id)
SELECT e.id, t.id
FROM (
    VALUES
        ('ostrov-mechty-park-razvlecheniy-2026', 'Развлечения'),
        ('ostrov-mechty-park-razvlecheniy-2026', 'Парк развлечений'),
        ('ostrov-mechty-park-razvlecheniy-2026', 'Аттракционы'),
        ('ostrov-mechty-park-razvlecheniy-2026', 'Семейный отдых'),
        ('ostrov-mechty-park-razvlecheniy-2026', 'Для детей'),
        ('ostrov-mechty-park-razvlecheniy-2026', 'Выходные'),

        ('moskvarium-oceanarium-2026', 'Развлечения'),
        ('moskvarium-oceanarium-2026', 'Океанариум'),
        ('moskvarium-oceanarium-2026', 'Шоу'),
        ('moskvarium-oceanarium-2026', 'Семейный отдых'),
        ('moskvarium-oceanarium-2026', 'Для детей'),

        ('solntse-moskvy-wheel-2026', 'Развлечения'),
        ('solntse-moskvy-wheel-2026', 'Колесо обозрения'),
        ('solntse-moskvy-wheel-2026', 'Панорама'),
        ('solntse-moskvy-wheel-2026', 'Семейный отдых'),
        ('solntse-moskvy-wheel-2026', 'Выходные'),

        ('kidzania-park-professiy-2026', 'Развлечения'),
        ('kidzania-park-professiy-2026', 'Профессии'),
        ('kidzania-park-professiy-2026', 'Для детей'),
        ('kidzania-park-professiy-2026', 'Семейный отдых'),

        ('muzey-magii-2026', 'Развлечения'),
        ('muzey-magii-2026', 'Музей'),
        ('muzey-magii-2026', 'Магия'),
        ('muzey-magii-2026', 'Для компании'),
        ('muzey-magii-2026', 'Выходные'),


        ('claustrophobia-quest-2026', 'Развлечения'),
        ('claustrophobia-quest-2026', 'Квест'),
        ('claustrophobia-quest-2026', 'Для компании'),
        ('claustrophobia-quest-2026', 'Выходные')
) AS v(event_slug, tag_name)
JOIN event e ON e.slug = v.event_slug
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;

COMMIT;
