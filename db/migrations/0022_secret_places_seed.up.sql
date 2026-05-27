-- CityHawk seed: коллекция «Тайные места»
-- Скрытые дворы, арт-точки и небанальные городские места Москвы.

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

-- 2. Категория и теги
INSERT INTO category (name)
VALUES ('Фото')
ON CONFLICT (name) DO NOTHING;

INSERT INTO tag (name)
VALUES
    ('Фото'),
    ('Фотоместо'),
    ('Прогулка'),
    ('Архитектура'),
    ('Историческое место'),
    ('Необычное место'),
    ('Арт-пространство'),
    ('Урбанистика'),
    ('Конструктивизм'),
    ('Стрит-арт'),
    ('Галерея'),
    ('Современное искусство'),
    ('Скрытый двор'),
    ('Бесплатно'),
    ('Для фото'),
    ('Одному'),
    ('Тихое место')
ON CONFLICT (name) DO NOTHING;

-- 3. Места на карте
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
            'Саввинское подворье',
            'Москва, Тверская ул., 6, стр. 1',
            55.760278::numeric,
            37.611667::numeric,
            'Красивое здание во дворе на Тверской, которое легко пропустить: оно спрятано за сталинскими домами. В 1930-е годы его не снесли, а передвинули вглубь двора. Это скрытая городская точка с необычной историей и выразительной архитектурой для фотографий.'
        ),
        (
            'Дом Моссельпрома',
            'Москва, Калашный пер., 2/10',
            55.753889::numeric,
            37.602500::numeric,
            'Конструктивистский дом с яркой советской рекламой Родченко, Степановой и слоганом Маяковского. Самое интересное видно не сразу: здание стоит обойти и заглянуть во двор, чтобы почувствовать атмосферу места.'
        ),
        (
            'Уличная галерея НЕТСТЕН',
            'Москва, стена вдоль ж/д путей между Курским вокзалом и тоннелем на Верхней Сыромятнической ул.',
            55.756900::numeric,
            37.661900::numeric,
            'Первая легальная уличная галерея, связанная с арт-кварталом между Винзаводом и Artplay. Подходит для маршрута со стрит-артом, бесплатной прогулки и неформального знакомства с креативным районом.'
        ),
        (
            'Винзавод',
            'Москва, 4-й Сыромятнический пер., 1/8, стр. 6',
            55.755569::numeric,
            37.664589::numeric,
            'Большой центр современного искусства с галереями, выставками и событиями рядом с метро Курская и Чкаловская. Здесь можно зайти в несколько галерей за раз и собрать спонтанный культурный маршрут.'
        ),
        (
            'Fine Art Gallery',
            'Москва, 4-й Сыромятнический пер., 1/8, стр. 9',
            55.755900::numeric,
            37.665000::numeric,
            'Одна из старейших московских галерей современного искусства, работающая с 1992 года. Камерное пространство внутри арт-кластера, где экспозиции меняются примерно раз в пару месяцев.'
        ),
        (
            'Библиотека Музея «Гараж»',
            'Москва, ул. Крымский Вал, 9, стр. 32',
            55.729900::numeric,
            37.601200::numeric,
            'Не самое очевидное место внутри известной институции: тихая библиотека Музея «Гараж», которая работает ежедневно с 11:00 до 22:00. Подходит для спокойного интеллектуального маршрута одному или одной.'
        ),
        (
            'Двор усадьбы Охотниковых',
            'Москва, ул. Пречистенка, 32/1',
            55.740900::numeric,
            37.593900::numeric,
            'Спрятанная городская точка из подборок небанальной Москвы. Место менее очевидное, чем главные туристические маршруты, и интересно именно ощущением скрытого двора и старого городского слоя.'
        )
) AS v(name, address_line, latitude, longitude, description)
WHERE c.country_name = 'Россия'
  AND c.name = 'Москва'
ON CONFLICT (city_id, name, address_line) DO UPDATE
SET latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    description = EXCLUDED.description,
    updated_at = now();

-- 4. События-точки для карты. Сессии для этой подборки не создаются.
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
            'Саввинское подворье',
            'savvinskoe-podvorie',
            'Тверская ул., 6, стр. 1',
            'Скрытый двор на Тверской с красивым зданием, которое в 1930-е не снесли, а передвинули вглубь квартала. Место интересно сочетанием парадной архитектуры, необычной городской истории и ощущения точки, которую легко пройти мимо.'
        ),
        (
            'Дом Моссельпрома',
            'dom-mosselproma',
            'Калашный пер., 2/10',
            'Конструктивистский дом с яркой советской рекламой, графикой Родченко и Степановой и слоганом Маяковского. Чтобы раскрыть место, стоит обойти здание и зайти во двор: так лучше считываются фасад, детали и локальный характер.'
        ),
        (
            'Уличная галерея НЕТСТЕН',
            'ulichnaya-galereya-netsten',
            'стена вдоль ж/д путей между Курским вокзалом и тоннелем на Верхней Сыромятнической ул.',
            'Легальная уличная галерея в креативном районе между Винзаводом и Artplay. Бесплатная точка для прогулки со стрит-артом, фотографиями и неформальной атмосферой московского арт-квартала.'
        ),
        (
            'Винзавод',
            'vinzavod',
            '4-й Сыромятнический пер., 1/8, стр. 6',
            'Центр современного искусства с галереями, выставками и событиями рядом с Курской и Чкаловской. Хороший вариант для спонтанного плана: можно зайти в несколько пространств за один маршрут и совместить выставки с прогулкой.'
        ),
        (
            'Fine Art Gallery',
            'fine-art-gallery',
            '4-й Сыромятнический пер., 1/8, стр. 9',
            'Камерная галерея современного искусства внутри арт-кластера, работающая с 1992 года. Подойдёт тем, кто ищет не массовое место, а небольшую экспозицию и более спокойный контакт с искусством.'
        ),
        (
            'Библиотека Музея «Гараж»',
            'biblioteka-muzeya-garazh',
            'ул. Крымский Вал, 9, стр. 32',
            'Тихий интеллектуальный спот внутри Музея «Гараж»: библиотека работает ежедневно с 11:00 до 22:00, а фонд редких книг доступен в рамках специальных событий. Хороший вариант для самостоятельного визита без туристической суеты.'
        ),
        (
            'Двор усадьбы Охотниковых',
            'dvor-usadby-okhotnikovykh',
            'ул. Пречистенка, 32/1',
            'Небанальный двор старой Москвы из маршрутов по скрытым городским местам. Подходит для тех, кто хочет уйти от очевидных туристических локаций и найти более тихую, атмосферную точку для прогулки и фото.'
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

-- 5. Привязка событий к местам для карты
INSERT INTO event_place (event_id, place_id)
SELECT e.id, p.id
FROM (
    VALUES
        ('savvinskoe-podvorie', 'Саввинское подворье'),
        ('dom-mosselproma', 'Дом Моссельпрома'),
        ('ulichnaya-galereya-netsten', 'Уличная галерея НЕТСТЕН'),
        ('vinzavod', 'Винзавод'),
        ('fine-art-gallery', 'Fine Art Gallery'),
        ('biblioteka-muzeya-garazh', 'Библиотека Музея «Гараж»'),
        ('dvor-usadby-okhotnikovykh', 'Двор усадьбы Охотниковых')
) AS v(event_slug, place_name)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id) DO UPDATE
SET place_id = EXCLUDED.place_id,
    updated_at = now();

-- 6. Изображения событий
DELETE FROM event_image ei
USING event e
WHERE ei.event_id = e.id
  AND e.slug IN (
    'savvinskoe-podvorie',
    'dom-mosselproma',
    'ulichnaya-galereya-netsten',
    'vinzavod',
    'fine-art-gallery',
    'biblioteka-muzeya-garazh',
    'dvor-usadby-okhotnikovykh'
  );

INSERT INTO event_image (event_id, image_url)
SELECT e.id, v.image_url
FROM (
    VALUES
        ('savvinskoe-podvorie', '/uploads/events/Savvinskoye_Podvorie_Moscow_Main_Entrance.jpg'),
        ('savvinskoe-podvorie', '/uploads/events/savvinspod-01.jpg'),
        ('savvinskoe-podvorie', '/uploads/events/Саввинское подворье - 022-1.jpg'),
        ('savvinskoe-podvorie', '/uploads/events/moskva-luchshe__img-3.8z0e4tan0q28.webp'),

        ('dom-mosselproma', '/uploads/events/Moscow_Mosselprom_0864.jpg'),
        ('dom-mosselproma', '/uploads/events/IMG_5671.jpg'),

        ('ulichnaya-galereya-netsten', '/uploads/events/wr-960.webp'),
        ('ulichnaya-galereya-netsten', '/uploads/events/wr-960-2.webp'),
        ('ulichnaya-galereya-netsten', '/uploads/events/IMG_0011.jpg'),

        ('vinzavod', '/uploads/events/L_height-7.webp'),
        ('vinzavod', '/uploads/events/L_height-8.webp'),
        ('vinzavod', '/uploads/events/vinzavod_msk86.jpeg'),
        ('vinzavod', '/uploads/events/ba4c8400e9384aac8733e27c5b96f9d1.webp'),

        ('fine-art-gallery', '/uploads/events/landscape.jpg'),
        ('fine-art-gallery', '/uploads/events/b2617a1d98fbd810c1ef2f4afa1a8b61.jpeg'),
        ('fine-art-gallery', '/uploads/events/Kate_Moss,_Sun_and_Henna_9_copy.jpg'),

        ('biblioteka-muzeya-garazh', '/uploads/events/eden-fine-art-london.jpg'),
        ('biblioteka-muzeya-garazh', '/uploads/events/large_image-1338ccfb-1d87-4607-a545-45aad5822faa.webp'),
        ('biblioteka-muzeya-garazh', '/uploads/events/Снимок-экрана-2022-08-27-в-14.39.50-678x381.png'),
        ('biblioteka-muzeya-garazh', '/uploads/events/dsc07420.jpg'),

        ('dvor-usadby-okhotnikovykh', '/uploads/events/Dom-Okhotnikovykh-3-800x599.jpg'),
        ('dvor-usadby-okhotnikovykh', '/uploads/events/4.jpeg'),
        ('dvor-usadby-okhotnikovykh', '/uploads/events/Dom-Okhotnikovykh-9-800x599.jpg'),
        ('dvor-usadby-okhotnikovykh', '/uploads/events/L_height-9.webp')
) AS v(event_slug, image_url)
JOIN event e ON e.slug = v.event_slug
ON CONFLICT (event_id, image_url) DO NOTHING;

-- 7. Категория «Фото»
INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM event e
JOIN category c ON c.name = 'Фото'
WHERE e.slug IN (
    'savvinskoe-podvorie',
    'dom-mosselproma',
    'ulichnaya-galereya-netsten',
    'vinzavod',
    'fine-art-gallery',
    'biblioteka-muzeya-garazh',
    'dvor-usadby-okhotnikovykh'
)
ON CONFLICT (event_id, category_id) DO NOTHING;

-- 8. Теги событий
INSERT INTO event_tag (event_id, tag_id)
SELECT e.id, t.id
FROM (
    VALUES
        ('savvinskoe-podvorie', 'Фото'),
        ('savvinskoe-podvorie', 'Фотоместо'),
        ('savvinskoe-podvorie', 'Архитектура'),
        ('savvinskoe-podvorie', 'Историческое место'),
        ('savvinskoe-podvorie', 'Скрытый двор'),
        ('savvinskoe-podvorie', 'Необычное место'),

        ('dom-mosselproma', 'Фото'),
        ('dom-mosselproma', 'Архитектура'),
        ('dom-mosselproma', 'Конструктивизм'),
        ('dom-mosselproma', 'Историческое место'),
        ('dom-mosselproma', 'Скрытый двор'),

        ('ulichnaya-galereya-netsten', 'Прогулка'),
        ('ulichnaya-galereya-netsten', 'Стрит-арт'),
        ('ulichnaya-galereya-netsten', 'Арт-пространство'),
        ('ulichnaya-galereya-netsten', 'Бесплатно'),
        ('ulichnaya-galereya-netsten', 'Урбанистика'),

        ('vinzavod', 'Арт-пространство'),
        ('vinzavod', 'Галерея'),
        ('vinzavod', 'Современное искусство'),
        ('vinzavod', 'Прогулка'),
        ('vinzavod', 'Для фото'),

        ('fine-art-gallery', 'Галерея'),
        ('fine-art-gallery', 'Современное искусство'),
        ('fine-art-gallery', 'Арт-пространство'),
        ('fine-art-gallery', 'Тихое место'),

        ('biblioteka-muzeya-garazh', 'Одному'),
        ('biblioteka-muzeya-garazh', 'Тихое место'),
        ('biblioteka-muzeya-garazh', 'Современное искусство'),
        ('biblioteka-muzeya-garazh', 'Необычное место'),

        ('dvor-usadby-okhotnikovykh', 'Скрытый двор'),
        ('dvor-usadby-okhotnikovykh', 'Историческое место'),
        ('dvor-usadby-okhotnikovykh', 'Необычное место'),
        ('dvor-usadby-okhotnikovykh', 'Фотоместо')
) AS v(event_slug, tag_name)
JOIN event e ON e.slug = v.event_slug
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;

-- 9. Коллекция «Тайные места»
INSERT INTO collection (author_user_id, city_id, title, slug, description, is_public)
SELECT
    u.id,
    c.id,
    'Тайные места',
    'secret-places-moscow',
    'Скрытые дворы, арт-точки, камерные галереи и небанальные места Москвы для прогулок, фото и спокойного городского открытия.',
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

-- 10. Картинка коллекции
DELETE FROM collection_image ci
USING collection col
WHERE ci.collection_id = col.id
  AND col.slug = 'secret-places-moscow';

INSERT INTO collection_image (collection_id, image_url)
SELECT
    col.id,
    '/uploads/events/Savvinskoye_Podvorie_Moscow_Main_Entrance.jpg'
FROM collection col
WHERE col.slug = 'secret-places-moscow'
ON CONFLICT (collection_id, image_url) DO NOTHING;

-- 11. Связь коллекции с событиями
INSERT INTO collection_event (collection_id, event_id)
SELECT
    col.id,
    e.id
FROM collection col
JOIN event e ON e.slug IN (
    'savvinskoe-podvorie',
    'dom-mosselproma',
    'ulichnaya-galereya-netsten',
    'vinzavod',
    'fine-art-gallery',
    'biblioteka-muzeya-garazh',
    'dvor-usadby-okhotnikovykh'
)
WHERE col.slug = 'secret-places-moscow'
ON CONFLICT (collection_id, event_id) DO NOTHING;

COMMIT;
