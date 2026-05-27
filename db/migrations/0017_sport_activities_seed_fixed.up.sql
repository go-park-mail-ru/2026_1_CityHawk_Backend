-- CityHawk seed: категория «Активный отдых»
-- Данные подготовлены для реальных спортивных активностей в Москве.
-- Фото распределены по 4 на каждое событие в порядке пользователя: начиная снизу первое относится к событию «Пейнтбол и лазертаг в ПК «Северный Олень»», далее события идут по списку.

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
VALUES ('Активный отдых')
ON CONFLICT (name) DO NOTHING;

INSERT INTO tag (name)
VALUES
    ('Спорт'),
    ('Активный отдых'),
    ('Командная игра'),
    ('Пейнтбол'),
    ('Лазертаг'),
    ('Картинг'),
    ('Скорость'),
    ('Скалолазание'),
    ('Батуты'),
    ('Вейкбординг'),
    ('Водные активности'),
    ('VR'),
    ('Квест'),
    ('Верёвочный парк'),
    ('Для компании'),
    ('Для детей'),
    ('Адреналин'),
    ('Тренировка')
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
        ('ПК «Северный Олень»', 'Москва, ул. Большая Академическая, 38, территория УСК «Наука»', 55.832963::numeric, 37.535001::numeric, 'Площадка Московского пейнтбольного фронта на территории УСК «Наука» для пейнтбола, лазертага и командных сценарных игр.'),
        ('MIKS Karting', 'Москва, Шарикоподшипниковская ул., 13, стр. 3', 55.719000::numeric, 37.681500::numeric, 'Крытый картинг-клуб в центре Москвы с прокатными заездами, детскими и взрослыми картами, хронометражем и возможностью провести командное мероприятие.'),
        ('Центр скалолазания ЦСКА', 'Москва, 3-я Песчаная ул., 2, стр. 1', 55.792400::numeric, 37.515600::numeric, 'Современный скалодром рядом с метро ЦСКА для тренировок, вводных занятий, самостоятельного лазания и спортивных мероприятий.'),
        ('Батутный центр «Высота»', 'Москва, Дубравная ул., 51, стр. 1', 55.840900::numeric, 37.354400::numeric, 'Батутный центр в Митино с зонами для прыжков, активного отдыха, тренировок и детских спортивных праздников.'),
        ('Вейк-клуб WakeStart', 'Москва, Лётная ул., 99, стр. 1', 55.824850::numeric, 37.415317::numeric, 'Вейк-клуб на воде для катания на вейкборде, обучения новичков и летнего активного отдыха.'),
        ('WARPOINT Columbus', 'Москва, Кировоградская ул., 13а, ТЦ Columbus', 55.612100::numeric, 37.607000::numeric, 'Парк виртуальной реальности в ТЦ Columbus для командных VR-игр, спортивных сценариев, праздников и квестов.'),
        ('Верёвочный парк SkyTown', 'Москва, просп. Мира, 119, стр. 27, ВДНХ', 55.834561::numeric, 37.612908::numeric, 'Верёвочный парк на ВДНХ с высотными маршрутами, страховкой, полосами препятствий и активным отдыхом для взрослых и детей.')
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
            'Пейнтбол и лазертаг в ПК «Северный Олень»',
            'paintball-lazertag-severnyy-olen-2026',
            'ПК «Северный Олень»',
            'Командная игра на открытых площадках ПК «Северный Олень» на территории УСК «Наука». Формат подойдёт для друзей, корпоративных команд и активного выходного: участники делятся на команды, получают экипировку и проходят сценарии с укрытиями, тактическими задачами и динамичными раундами. Можно выбрать пейнтбол или лазертаг, а также адаптировать игру под уровень группы. Площадка подойдёт тем, кто хочет адреналина, движения и живого командного взаимодействия.',
            12,
            'https://xn-----9kcfbnf0angogofe7bi.com/reindeer'
        ),
        (
            'Заезды на картах в MIKS Karting',
            'miks-karting-zaezdy-2026',
            'MIKS Karting',
            'Картинг-заезды на крытой трассе MIKS Karting — активность для тех, кто любит скорость, соревнование и азарт. Гости проходят инструктаж, получают экипировку и выходят на трассу, где можно соревноваться за лучшее время круга. Формат подходит для новичков, опытных любителей, компаний друзей и небольших командных мероприятий. На площадке доступны взрослые и детские заезды, а результаты можно отслеживать по хронометражу.',
            6,
            'https://miks-karting.ru/'
        ),
        (
            'Скалодром ЦСКА',
            'climbing-cska-intro-2026',
            'Центр скалолазания ЦСКА',
            'Вводное занятие по скалолазанию в Центре скалолазания ЦСКА — хороший способ попробовать вертикальный спорт в безопасном формате. Инструктор объясняет базовые правила, показывает технику движения, помогает подобрать уровень трасс и следит за безопасностью. Подходит для новичков, детей и взрослых, а также для тех, кто хочет разнообразить тренировки и попробовать спортивный формат без долгой подготовки.',
            6,
            'https://yandex.ru/maps/org/tsentr_skalolazaniya_tsska/142904419110/'
        ),
        (
            'Активный день в батутном центре «Высота»',
            'batutnyy-tsentr-vysota-2026',
            'Батутный центр «Высота»',
            'Батутный центр «Высота» — пространство для активного отдыха, прыжков и тренировок. Здесь можно провести спортивный час с друзьями, устроить детский праздник или просто выплеснуть энергию после учёбы и работы. Формат подходит для разного уровня подготовки: от простых прыжков до отработки координации, баланса и базовых акробатических элементов под присмотром сотрудников площадки.',
            6,
            'https://vysota-park.ru/kontakty/'
        ),
        (
            'Вейкбординг в WakeStart',
            'wakestart-wakeboarding-2026',
            'Вейк-клуб WakeStart',
            'Вейкбординг в WakeStart — летняя водная активность для тех, кто хочет попробовать катание за лебёдкой, научиться держать баланс и почувствовать скорость на воде. Формат подойдёт новичкам и тем, кто уже катался: можно взять вводное занятие с инструктором, прокатиться самостоятельно или собрать компанию для активного дня у воды. Рекомендуется взять сменную одежду, полотенце и хорошее настроение.',
            12,
            'https://yandex.ru/maps/org/veykstart/1361664121/'
        ),
        (
            'VR-игра в WARPOINT Columbus',
            'warpoint-columbus-vr-arena-2026',
            'WARPOINT Columbus',
            'Командная VR-игра в WARPOINT Columbus — это активный квест в виртуальной реальности, где участники перемещаются по арене, взаимодействуют друг с другом и проходят сценарии в полном погружении. Формат подходит для компании друзей, дня рождения, командного вечера или необычного спортивного досуга. Игры помогают совместить движение, тактику, реакцию и соревновательный азарт без привычного спортивного инвентаря.',
            10,
            'https://columbus.warpoint.ru/'
        ),
        (
            'Маршруты верёвочного парка SkyTown',
            'skytown-rope-park-2026',
            'Верёвочный парк SkyTown',
            'Верёвочный парк SkyTown на ВДНХ — высотная активность с маршрутами, препятствиями и страховкой. Гости проходят инструктаж, надевают снаряжение и отправляются на трассы разной сложности: от более спокойных маршрутов до испытаний на высоте для тех, кто хочет проверить смелость и ловкость. Формат подходит для семей, компаний друзей и активных свиданий, а также для тех, кто хочет провести день на свежем воздухе.',
            6,
            'https://yandex.ru/maps/org/skaytaun/1353121232/'
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
        ('paintball-lazertag-severnyy-olen-2026', 'ПК «Северный Олень»'),
        ('miks-karting-zaezdy-2026', 'MIKS Karting'),
        ('climbing-cska-intro-2026', 'Центр скалолазания ЦСКА'),
        ('batutnyy-tsentr-vysota-2026', 'Батутный центр «Высота»'),
        ('wakestart-wakeboarding-2026', 'Вейк-клуб WakeStart'),
        ('warpoint-columbus-vr-arena-2026', 'WARPOINT Columbus'),
        ('skytown-rope-park-2026', 'Верёвочный парк SkyTown')
) AS v(event_slug, place_name)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id) DO UPDATE
SET place_id = EXCLUDED.place_id,
    updated_at = now();

-- 6. Сессии активностей
INSERT INTO event_session (event_id, place_id, start_at, end_at, price)
SELECT e.id, p.id, v.start_at::timestamptz, v.end_at::timestamptz, v.price
FROM (
    VALUES
        ('paintball-lazertag-severnyy-olen-2026', 'ПК «Северный Олень»', '2026-06-06 12:00:00+03', '2026-06-06 15:00:00+03', 1800),
        ('paintball-lazertag-severnyy-olen-2026', 'ПК «Северный Олень»', '2026-06-20 12:00:00+03', '2026-06-20 15:00:00+03', 1800),

        ('miks-karting-zaezdy-2026', 'MIKS Karting', '2026-06-07 15:00:00+03', '2026-06-07 17:00:00+03', 1800),
        ('miks-karting-zaezdy-2026', 'MIKS Karting', '2026-06-21 15:00:00+03', '2026-06-21 17:00:00+03', 1800),

        ('climbing-cska-intro-2026', 'Центр скалолазания ЦСКА', '2026-06-10 19:00:00+03', '2026-06-10 21:00:00+03', 1500),
        ('climbing-cska-intro-2026', 'Центр скалолазания ЦСКА', '2026-06-24 19:00:00+03', '2026-06-24 21:00:00+03', 1500),

        ('batutnyy-tsentr-vysota-2026', 'Батутный центр «Высота»', '2026-06-13 13:00:00+03', '2026-06-13 15:00:00+03', 1000),
        ('batutnyy-tsentr-vysota-2026', 'Батутный центр «Высота»', '2026-06-27 13:00:00+03', '2026-06-27 15:00:00+03', 1000),

        ('wakestart-wakeboarding-2026', 'Вейк-клуб WakeStart', '2026-07-04 11:00:00+03', '2026-07-04 13:00:00+03', 2500),
        ('wakestart-wakeboarding-2026', 'Вейк-клуб WakeStart', '2026-07-18 11:00:00+03', '2026-07-18 13:00:00+03', 2500),

        ('warpoint-columbus-vr-arena-2026', 'WARPOINT Columbus', '2026-06-14 16:00:00+03', '2026-06-14 18:00:00+03', 1600),
        ('warpoint-columbus-vr-arena-2026', 'WARPOINT Columbus', '2026-06-28 16:00:00+03', '2026-06-28 18:00:00+03', 1600),

        ('skytown-rope-park-2026', 'Верёвочный парк SkyTown', '2026-07-05 12:00:00+03', '2026-07-05 14:00:00+03', 1800),
        ('skytown-rope-park-2026', 'Верёвочный парк SkyTown', '2026-07-19 12:00:00+03', '2026-07-19 14:00:00+03', 1800)
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
    'paintball-lazertag-severnyy-olen-2026',
    'miks-karting-zaezdy-2026',
    'climbing-cska-intro-2026',
    'batutnyy-tsentr-vysota-2026',
    'wakestart-wakeboarding-2026',
    'warpoint-columbus-vr-arena-2026',
    'skytown-rope-park-2026'
  );

INSERT INTO event_image (event_id, image_url)
SELECT e.id, v.image_url
FROM (
    VALUES
        -- Порядок снизу вверх: по 4 фото на каждое событие
        ('paintball-lazertag-severnyy-olen-2026', '/uploads/events/orig-3.jpeg'),
        ('paintball-lazertag-severnyy-olen-2026', '/uploads/events/L_height.webp'),
        ('paintball-lazertag-severnyy-olen-2026', '/uploads/events/M_height.jpeg'),
        ('paintball-lazertag-severnyy-olen-2026', '/uploads/events/M_height-2.jpeg'),

        ('miks-karting-zaezdy-2026', '/uploads/events/L_height-2.webp'),
        ('miks-karting-zaezdy-2026', '/uploads/events/ijtw5qsnuo7unpfiw0k0e9nzirikdf7n.webp'),
        ('miks-karting-zaezdy-2026', '/uploads/events/010esprqb0ot4m12ya4pmrufxiqayl7o.webp'),
        ('miks-karting-zaezdy-2026', '/uploads/events/2022-01-27-14.03.52-1024x768.jpg'),

        ('climbing-cska-intro-2026', '/uploads/events/z1tw7jbo7bhnavv37q9btr0kuic0bbrd.jpg'),
        ('climbing-cska-intro-2026', '/uploads/events/2368sqw0r7zi08ounyfsdqixu83d5hqp.jpg'),
        ('climbing-cska-intro-2026', '/uploads/events/L_height-3.webp'),
        ('climbing-cska-intro-2026', '/uploads/events/ytxul2pyezc2s8wx1cwd3n3s9hlhnwsg.jpg'),

        ('batutnyy-tsentr-vysota-2026', '/uploads/events/vysota-2.jpg'),
        ('batutnyy-tsentr-vysota-2026', '/uploads/events/L_height-4.webp'),
        ('batutnyy-tsentr-vysota-2026', '/uploads/events/f4316fc727b45137683298086641b8aec6e487dd-e1747959556295.webp'),
        ('batutnyy-tsentr-vysota-2026', '/uploads/events/img20230314131038.jpg'),

        ('wakestart-wakeboarding-2026', '/uploads/events/M_height-3.jpeg'),
        ('wakestart-wakeboarding-2026', '/uploads/events/WakeStart-21-07-2023.jpg'),
        ('wakestart-wakeboarding-2026', '/uploads/events/WS_01.jpg'),
        ('wakestart-wakeboarding-2026', '/uploads/events/M_height-4.jpeg'),

        ('warpoint-columbus-vr-arena-2026', '/uploads/events/_2.jpg'),
        ('warpoint-columbus-vr-arena-2026', '/uploads/events/__3.jpeg'),
        ('warpoint-columbus-vr-arena-2026', '/uploads/events/_WARPOINT_1312_2.jpg'),
        ('warpoint-columbus-vr-arena-2026', '/uploads/events/4_.jpg'),

        ('skytown-rope-park-2026', '/uploads/events/2187e9132bab56c7119f9d9be1529a56.webp'),
        ('skytown-rope-park-2026', '/uploads/events/caa0844393f682e6babd8eb40fefffc3.webp'),
        ('skytown-rope-park-2026', '/uploads/events/4.png'),
        ('skytown-rope-park-2026', '/uploads/events/64bdf1947055ad0478de463eca1e9254.jpg')
) AS v(event_slug, image_url)
JOIN event e ON e.slug = v.event_slug
ON CONFLICT (event_id, image_url) DO NOTHING;

-- 8. Категория «Активный отдых»
DELETE FROM event_category ec
USING event e, category c
WHERE ec.event_id = e.id
  AND ec.category_id = c.id
  AND e.slug IN (
    'paintball-lazertag-severnyy-olen-2026',
    'miks-karting-zaezdy-2026',
    'climbing-cska-intro-2026',
    'batutnyy-tsentr-vysota-2026',
    'wakestart-wakeboarding-2026',
    'warpoint-columbus-vr-arena-2026',
    'skytown-rope-park-2026'
  )
  AND c.name IN ('Спорт и активный отдых', 'Активный отдых');

INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM event e
JOIN category c ON c.name = 'Активный отдых'
WHERE e.slug IN (
    'paintball-lazertag-severnyy-olen-2026',
    'miks-karting-zaezdy-2026',
    'climbing-cska-intro-2026',
    'batutnyy-tsentr-vysota-2026',
    'wakestart-wakeboarding-2026',
    'warpoint-columbus-vr-arena-2026',
    'skytown-rope-park-2026'
)
ON CONFLICT (event_id, category_id) DO NOTHING;

-- 9. Теги событий
INSERT INTO event_tag (event_id, tag_id)
SELECT e.id, t.id
FROM (
    VALUES
        ('paintball-lazertag-severnyy-olen-2026', 'Спорт'),
        ('paintball-lazertag-severnyy-olen-2026', 'Активный отдых'),
        ('paintball-lazertag-severnyy-olen-2026', 'Командная игра'),
        ('paintball-lazertag-severnyy-olen-2026', 'Пейнтбол'),
        ('paintball-lazertag-severnyy-olen-2026', 'Лазертаг'),
        ('paintball-lazertag-severnyy-olen-2026', 'Для компании'),
        ('paintball-lazertag-severnyy-olen-2026', 'Адреналин'),

        ('miks-karting-zaezdy-2026', 'Спорт'),
        ('miks-karting-zaezdy-2026', 'Активный отдых'),
        ('miks-karting-zaezdy-2026', 'Картинг'),
        ('miks-karting-zaezdy-2026', 'Скорость'),
        ('miks-karting-zaezdy-2026', 'Для компании'),
        ('miks-karting-zaezdy-2026', 'Адреналин'),

        ('climbing-cska-intro-2026', 'Спорт'),
        ('climbing-cska-intro-2026', 'Активный отдых'),
        ('climbing-cska-intro-2026', 'Скалолазание'),
        ('climbing-cska-intro-2026', 'Тренировка'),
        ('climbing-cska-intro-2026', 'Для детей'),

        ('batutnyy-tsentr-vysota-2026', 'Спорт'),
        ('batutnyy-tsentr-vysota-2026', 'Активный отдых'),
        ('batutnyy-tsentr-vysota-2026', 'Батуты'),
        ('batutnyy-tsentr-vysota-2026', 'Для детей'),
        ('batutnyy-tsentr-vysota-2026', 'Для компании'),

        ('wakestart-wakeboarding-2026', 'Спорт'),
        ('wakestart-wakeboarding-2026', 'Активный отдых'),
        ('wakestart-wakeboarding-2026', 'Вейкбординг'),
        ('wakestart-wakeboarding-2026', 'Водные активности'),
        ('wakestart-wakeboarding-2026', 'Адреналин'),

        ('warpoint-columbus-vr-arena-2026', 'Спорт'),
        ('warpoint-columbus-vr-arena-2026', 'Активный отдых'),
        ('warpoint-columbus-vr-arena-2026', 'VR'),
        ('warpoint-columbus-vr-arena-2026', 'Квест'),
        ('warpoint-columbus-vr-arena-2026', 'Командная игра'),
        ('warpoint-columbus-vr-arena-2026', 'Для компании'),

        ('skytown-rope-park-2026', 'Спорт'),
        ('skytown-rope-park-2026', 'Активный отдых'),
        ('skytown-rope-park-2026', 'Верёвочный парк'),
        ('skytown-rope-park-2026', 'Для детей'),
        ('skytown-rope-park-2026', 'Адреналин'),
        ('skytown-rope-park-2026', 'Для компании')
) AS v(event_slug, tag_name)
JOIN event e ON e.slug = v.event_slug
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;

COMMIT;
