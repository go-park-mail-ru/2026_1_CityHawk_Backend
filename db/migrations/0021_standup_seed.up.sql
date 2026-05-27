-- CityHawk seed: категория «Стендап»
-- Данные подготовлены из пользовательского списка событий.
-- Изображения проставлены по обновлённому порядку пользователя: начиная снизу первое фото относится к событию «Женский Стендап», далее по порядку вверх.

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

-- 1. Города и seed-автор
INSERT INTO city (name, country_name, timezone)
VALUES
    ('Москва', 'Россия', 'Europe/Moscow'),
    ('Дмитров', 'Россия', 'Europe/Moscow'),
    ('Одинцово', 'Россия', 'Europe/Moscow')
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
VALUES ('Стендап')
ON CONFLICT (name) DO NOTHING;

INSERT INTO tag (name)
VALUES
    ('Стендап'),
    ('Комедия'),
    ('ТНТ'),
    ('Женский юмор'),
    ('Открытая площадка'),
    ('Сольный концерт'),
    ('Тур'),
    ('Большой зал'),
    ('Бар'),
    ('Бесплатно'),
    ('Подборка комиков'),
    ('Для компании'),
    ('Новая программа')
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
FROM (
    VALUES
        ('Москва', 'Зелёный театр ВДНХ', 'Москва, просп. Мира, 119, стр. 545', 55.831540::numeric, 37.617282::numeric, 'Открытая концертная площадка на территории ВДНХ для летних концертов, шоу и стендап-мероприятий.'),
        ('Дмитров', 'ЦДК «Созвездие»', 'Московская область, Дмитров, Загорская ул., 64', 56.341029::numeric, 37.526016::numeric, 'Центральный дворец культуры Дмитрова, где проходят концерты, спектакли и гастрольные комедийные программы.'),
        ('Одинцово', 'КСЦ «Мечта»', 'Московская область, Одинцово, ул. Маршала Жукова, 38', 55.681143::numeric, 37.269933::numeric, 'Культурно-спортивный центр в Одинцово с концертным залом для гастрольных шоу, спектаклей и стендапа.'),
        ('Москва', 'Бар «Настоишная»', 'Москва, Нижний Сусальный пер., 4, стр. 3', 55.759400::numeric, 37.666000::numeric, 'Бар рядом с Курской, где регулярно проходят камерные стендап-вечера с комиками и свободной атмосферой.')
) AS v(city_name, name, address_line, latitude, longitude, description)
JOIN city c
  ON c.name = v.city_name
 AND c.country_name = 'Россия'
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
            'Женский Стендап',
            'zhenskiy-standup-vdnh-2026-05-29',
            'Зелёный театр ВДНХ',
            '«Женский стендап» — первое и единственное шоу на ТНТ и российском телевидении, в котором женщины честно и откровенно шутят обо всём, что их волнует, интересует и веселит. Без стереотипов и стеснения — только жёсткая, хлёсткая правда и узнаваемые жизненные темы. Для героинь проекта нет запретных тем: отношения, семья, стандарты, повседневные сложности и честный взгляд на то, каково быть хорошей женой, матерью и просто собой. Проект давно обрёл огромную аудиторию, а его резидентки стали суперзвёздами российского юмора. В составе заявлены Карина Мейханаджян, Надя Джабраилова, Белла Малу, Елизавета Варвара Аранова, Найка Казиева, Саша Муратова, Мария Маркова, Ирина Мягкова, Варвара Щербакова и Маргарита Родина. Организатор оставляет за собой право вносить изменения в состав артистов.',
            18,
            'https://vdnh.ru/places/zelenyy-teatr-vdnkh/'
        ),
        (
            'Иван Абрамов. Всё из детства',
            'ivan-abramov-vsyo-iz-detstva-2026-10-16',
            'ЦДК «Созвездие»',
            'Новая программа Ивана Абрамова «Всё из детства» — это не просто stand up, а концерт-терапия для осознанных зрителей, готовых к лёгкому погружению в себя. По словам комика, реакции взрослого человека на жизненные события часто определяются механизмами, заложенными ещё в раннем детстве, поэтому чтобы стать лучше сегодня, нужно вернуться к истокам и проработать свои травмы. В программе — как всегда жизненно, харизматично, музыкально, психологически близко и по-настоящему смешно. Концерт наполнен фирменным абрамовским юмором, который отличает Ивана от других комиков.',
            18,
            'https://xn----dtbecebkckn9b9a2d.xn--p1ai/'
        ),
        (
            'Сергей Орлов. Программа «Звезда»',
            'sergey-orlov-zvezda-2026-05-31',
            'ЦДК «Созвездие»',
            'Сергей Орлов привозит новую программу 2026 года «Звезда». Один из самых узнаваемых российских стендап-комиков и настоящий народный любимец выходит на сцену с честным, остроумным и по-настоящему жизненным юмором. Это гастрольное выступление для зрителей, которым близок современный разговорный стендап без лишней мишуры.',
            18,
            'https://xn----dtbecebkckn9b9a2d.xn--p1ai/'
        ),
        (
            'Андрей Бебуришвили «Удобный»',
            'andrey-beburishvili-udobnyy-2026-06-14',
            'КСЦ «Мечта»',
            'Сольный стендап-концерт Андрея Бебуришвили «Удобный» — это смех, дерзость и немного ностальгии. Программа построена на узнаваемых наблюдениях, остром юморе и живой подаче, благодаря которой вечер обещает быть ярким и эмоциональным. Отличный вариант для тех, кто любит современный стендап с характером.',
            18,
            'https://www.kscmechta.ru/'
        ),
        (
            'Бесплатный Stand Up от комиков с ТНТ',
            'besplatnyy-stand-up-ot-komikov-s-tnt-2026-06',
            'Бар «Настоишная»',
            'Во вторники в баре «Настоишная» проходят бесплатные стендап-вечера с комиками с ТНТ. Тем, кому не хватает смелости, страсти и ярких эмоций в середине недели, предлагают отправиться в бар на вечер шуток, напитков и живого общения. В программе — выступления топовых комиков, большой выбор напитков и еда по демократичным ценам. Вход свободный, действует депозит на еду и напитки — 700 рублей с человека.',
            18,
            'https://nastoishnaya.ru/'
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
        ('zhenskiy-standup-vdnh-2026-05-29', 'Зелёный театр ВДНХ'),
        ('ivan-abramov-vsyo-iz-detstva-2026-10-16', 'ЦДК «Созвездие»'),
        ('sergey-orlov-zvezda-2026-05-31', 'ЦДК «Созвездие»'),
        ('andrey-beburishvili-udobnyy-2026-06-14', 'КСЦ «Мечта»'),
        ('besplatnyy-stand-up-ot-komikov-s-tnt-2026-06', 'Бар «Настоишная»')
) AS v(event_slug, place_name)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id) DO UPDATE
SET place_id = EXCLUDED.place_id,
    updated_at = now();

-- 6. Сессии стендапов
INSERT INTO event_session (event_id, place_id, start_at, end_at, price)
SELECT e.id, p.id, v.start_at::timestamptz, v.end_at::timestamptz, v.price
FROM (
    VALUES
        ('zhenskiy-standup-vdnh-2026-05-29', 'Зелёный театр ВДНХ', '2026-05-29 20:00:00+03', '2026-05-29 22:00:00+03', 0),
        ('ivan-abramov-vsyo-iz-detstva-2026-10-16', 'ЦДК «Созвездие»', '2026-10-16 18:00:00+03', '2026-10-16 20:00:00+03', 0),
        ('sergey-orlov-zvezda-2026-05-31', 'ЦДК «Созвездие»', '2026-05-31 19:00:00+03', '2026-05-31 21:00:00+03', 0),
        ('andrey-beburishvili-udobnyy-2026-06-14', 'КСЦ «Мечта»', '2026-06-14 19:00:00+03', '2026-06-14 21:00:00+03', 0),
        ('besplatnyy-stand-up-ot-komikov-s-tnt-2026-06', 'Бар «Настоишная»', '2026-06-02 20:00:00+03', '2026-06-02 22:00:00+03', 0),
        ('besplatnyy-stand-up-ot-komikov-s-tnt-2026-06', 'Бар «Настоишная»', '2026-06-09 20:00:00+03', '2026-06-09 22:00:00+03', 0),
        ('besplatnyy-stand-up-ot-komikov-s-tnt-2026-06', 'Бар «Настоишная»', '2026-06-16 20:00:00+03', '2026-06-16 22:00:00+03', 0),
        ('besplatnyy-stand-up-ot-komikov-s-tnt-2026-06', 'Бар «Настоишная»', '2026-06-23 20:00:00+03', '2026-06-23 22:00:00+03', 0),
        ('besplatnyy-stand-up-ot-komikov-s-tnt-2026-06', 'Бар «Настоишная»', '2026-06-30 20:00:00+03', '2026-06-30 22:00:00+03', 0)
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
        -- Порядок снизу вверх: Женский Стендап, Иван Абрамов, Сергей Орлов, Андрей Бебуришвили, Бесплатный Stand Up
        ('zhenskiy-standup-vdnh-2026-05-29', '/uploads/events/42184151cb744d869ad3cc85e4ef.jpg'),
        ('ivan-abramov-vsyo-iz-detstva-2026-10-16', '/uploads/events/9627ad2881b24ad780838b4dd60d.jpg'),
        ('sergey-orlov-zvezda-2026-05-31', '/uploads/events/e9f95662bd334155840a3d94407a.jpg'),
        ('andrey-beburishvili-udobnyy-2026-06-14', '/uploads/events/4b5e48abbd5240e28ea3e171c0ec.jpg'),
        ('besplatnyy-stand-up-ot-komikov-s-tnt-2026-06', '/uploads/events/64ace53f927b49d6972f6b677329.jpg')
) AS v(event_slug, image_url)
JOIN event e ON e.slug = v.event_slug
ON CONFLICT (event_id, image_url) DO NOTHING;

-- 8. Категория «Стендап»
INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM event e
JOIN category c ON c.name = 'Стендап'
WHERE e.slug IN (
    'zhenskiy-standup-vdnh-2026-05-29',
    'ivan-abramov-vsyo-iz-detstva-2026-10-16',
    'sergey-orlov-zvezda-2026-05-31',
    'andrey-beburishvili-udobnyy-2026-06-14',
    'besplatnyy-stand-up-ot-komikov-s-tnt-2026-06'
)
ON CONFLICT (event_id, category_id) DO NOTHING;

-- 9. Теги событий
INSERT INTO event_tag (event_id, tag_id)
SELECT e.id, t.id
FROM (
    VALUES
        ('zhenskiy-standup-vdnh-2026-05-29', 'Стендап'),
        ('zhenskiy-standup-vdnh-2026-05-29', 'Комедия'),
        ('zhenskiy-standup-vdnh-2026-05-29', 'ТНТ'),
        ('zhenskiy-standup-vdnh-2026-05-29', 'Женский юмор'),
        ('zhenskiy-standup-vdnh-2026-05-29', 'Открытая площадка'),
        ('zhenskiy-standup-vdnh-2026-05-29', 'Для компании'),

        ('ivan-abramov-vsyo-iz-detstva-2026-10-16', 'Стендап'),
        ('ivan-abramov-vsyo-iz-detstva-2026-10-16', 'Комедия'),
        ('ivan-abramov-vsyo-iz-detstva-2026-10-16', 'Сольный концерт'),
        ('ivan-abramov-vsyo-iz-detstva-2026-10-16', 'Новая программа'),
        ('ivan-abramov-vsyo-iz-detstva-2026-10-16', 'Большой зал'),

        ('sergey-orlov-zvezda-2026-05-31', 'Стендап'),
        ('sergey-orlov-zvezda-2026-05-31', 'Комедия'),
        ('sergey-orlov-zvezda-2026-05-31', 'Сольный концерт'),
        ('sergey-orlov-zvezda-2026-05-31', 'Тур'),
        ('sergey-orlov-zvezda-2026-05-31', 'Новая программа'),

        ('andrey-beburishvili-udobnyy-2026-06-14', 'Стендап'),
        ('andrey-beburishvili-udobnyy-2026-06-14', 'Комедия'),
        ('andrey-beburishvili-udobnyy-2026-06-14', 'Сольный концерт'),
        ('andrey-beburishvili-udobnyy-2026-06-14', 'Большой зал'),

        ('besplatnyy-stand-up-ot-komikov-s-tnt-2026-06', 'Стендап'),
        ('besplatnyy-stand-up-ot-komikov-s-tnt-2026-06', 'Комедия'),
        ('besplatnyy-stand-up-ot-komikov-s-tnt-2026-06', 'ТНТ'),
        ('besplatnyy-stand-up-ot-komikov-s-tnt-2026-06', 'Бар'),
        ('besplatnyy-stand-up-ot-komikov-s-tnt-2026-06', 'Бесплатно'),
        ('besplatnyy-stand-up-ot-komikov-s-tnt-2026-06', 'Подборка комиков'),
        ('besplatnyy-stand-up-ot-komikov-s-tnt-2026-06', 'Для компании')
) AS v(event_slug, tag_name)
JOIN event e ON e.slug = v.event_slug
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;

COMMIT;
