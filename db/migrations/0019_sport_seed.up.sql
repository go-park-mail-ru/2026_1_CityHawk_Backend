-- CityHawk seed: категория «Спортивные мероприятия»
-- Источник данных: афиша спортивных событий Москвы 2026 (РФС, КХЛ, Единая лига ВТБ, Беговое сообщество и др.)
-- Изображения взяты из открытых легальных источников (Википедия, официальные сайты клубов и организаторов)
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
VALUES ('Спортивные мероприятия')
ON CONFLICT (name) DO NOTHING;

-- 3. Теги
INSERT INTO tag (name)
VALUES
    ('Спортивные мероприятия'), ('Футбол'), ('Хоккей'), ('Баскетбол'), ('Бег'), ('Лыжи'),
    ('Экстремальный спорт'), ('Фестиваль'), ('Семейный'), ('Зрелищное шоу'),
    ('Open air'), ('Зимний спорт'), ('Профессиональный спорт')
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
        ('Стадион «Лужники»', 'Москва, ул. Лужники, 24, стр. 1', 55.7158::numeric, 37.5537::numeric, 'Крупнейшая спортивная арена страны, главная стадионная площадка Москвы.'),
        ('ЛД «Легенды хоккея — Сокольники»', 'Москва, ул. Сокольнический Вал, 1б', 55.7885::numeric, 37.6712::numeric, 'Современный ледовый дворец, домашняя арена молодёжного хоккея.'),
        ('Дворец спорта «Мегаспорт»', 'Москва, Ходынский б-р, 3', 55.7865::numeric, 37.5531::numeric, 'Многофункциональный спортивный комплекс, принимающий матчи Единой лиги ВТБ.'),
        ('Стадион «Искра»', 'Москва, Сельскохозяйственная ул., 26/1', 55.8456::numeric, 37.6439::numeric, 'Спортивный комплекс, площадка для проведения фестиваля «Спортлэнд».'),
        ('ЦСКА Арена', 'Москва, ул. Автозаводская, 23а', 55.7024::numeric, 37.6426::numeric, 'Многофункциональная арена, место проведения фестиваля экстремального спорта «Прорыв».'),
        ('Парк Горького', 'Москва, ул. Крымский Вал, 9', 55.7287::numeric, 37.6018::numeric, 'Центральный парк, одна из площадок фестиваля «Ночь московского спорта».'),
        ('СК «Альфа-Битца»', 'Москва, 36-й км МКАД, зона отдыха «Альфа-Битца»', 55.5782::numeric, 37.5578::numeric, 'Лыжная трасса для проведения Московского лыжного марафона.'),
        ('Улица Косыгина / МГУ — «Лужники»', 'Москва, улица Косыгина, Лужнецкая наб.', 55.7093::numeric, 37.5527::numeric, 'Трасса Московского полумарафона, соединяющая Университетскую площадь и спорткомплекс «Лужники».'),
        ('Арбат', 'Москва, ул. Арбат', 55.7515::numeric, 37.5970::numeric, 'Пешеходная улица, площадка фестиваля «Ночь московского спорта».')
) AS v(name, address_line, latitude, longitude, description)
WHERE c.country_name = 'Россия'
  AND c.name = 'Москва'
ON CONFLICT (city_id, name, address_line) DO UPDATE
SET latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    description = EXCLUDED.description,
    updated_at = now();

-- 5. События (спортивные мероприятия)
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
        ('Суперфинал FONBET Кубка России: «Спартак» — «Краснодар»', 'russian-cup-final-spartak-krasnodar-2026', 'Стадион «Лужники»', 'Главный футбольный матч сезона. Московский «Спартак» встретится с «Краснодаром» в решающем поединке за обладание Кубком России. В программе гала-матч легенд футбола и концерт.', 6, 'https://superfinal.rfs.ru/'),
        ('Финал Кубка Харламова МХЛ: МХК «Спартак» — «Локо» (Матч 6)', 'kharlamov-cup-final-spartak-loko-2026', 'ЛД «Легенды хоккея — Сокольники»', 'Шестой матч финальной серии Кубка Харламова. Молодёжный хоккей: столичный «Спартак» принимает ярославский «Локо». Действующий обладатель трофея «Спартак» борется за защиту титула.', 0, 'https://www.hockey-world.net/junior/324026-raspisanie-final-noj-serii-kubka-harlamova-s-ucastiem-mhk-loko-i-mhk-spartak'),
        ('Единая лига ВТБ. Плей-офф: ЦСКА — «Локомотив-Кубань»', 'vtb-united-league-cska-lokomotiv-2026', 'Дворец спорта «Мегаспорт»', 'Пятый матч полуфинальной серии плей-офф Единой лиги ВТБ. Московский ЦСКА борется за выход в финал с краснодарским «Локомотивом-Кубанью».', 0, 'https://cskabasket.ru/en/game/4658/comments/'),
        ('Семейный фестиваль «Спортлэнд»', 'sportland-festival-2026', 'Стадион «Искра»', 'Более 100 видов спорта, игровые маршруты, встречи с олимпийскими чемпионами. Два дня бесплатных активностей для всей семьи. Вход свободный по регистрации.', 0, 'https://www.afisha.ru/article/semeynyy-festival-sportlend-v-moskve/'),
        ('Московский полумарафон', 'moscow-half-marathon-2026', 'Улица Косыгина / МГУ — «Лужники»', 'Один из крупнейших весенних беговых стартов. Участники из 33 стран и 80 регионов России. Дистанции 5 км, 21,1 км, детский забег и корпоративные эстафеты.', 6, 'https://www.kp.ru/afisha/msk/sportivnye-meropriyatiya/moskovskij-polumarafon/'),
        ('Московский лыжный марафон', 'moscow-ski-marathon-2026', 'СК «Альфа-Битца»', 'Масштабное зимнее спортивное событие. Любители и профессионалы выходят на дистанции 25 и 50 километров. Старт в 11:00, награждение победителей в 14:00.', 12, 'https://www.kp.ru/afisha/msk/sportivnye-meropriyatiya/moskovskij-lyzhnyj-marafon/'),
        ('Фестиваль «Ночь московского спорта»', 'night-of-moscow-sport-2026', 'Парк Горького, Арбат', 'Более 100 бесплатных тренировок: падел-теннис, сайклинг под диджея, танцы, воркаут, паркур, гидрофлай-шоу, а также встречи со спортсменами. Мероприятие проходит с 19:00 до 22:00.', 0, 'https://www.kp.ru/afisha/msk/festivali/festival-noch-moskovskogo-sporta-v-moskve/'),
        ('Фестиваль экстремального спорта «Прорыв»', 'proryv-extreme-sports-festival-2026', 'ЦСКА Арена', 'Мотофристайл (FMX), дрифт, стантрайдинг, шоу каскадеров и силовой экстрим. В фестивале участвуют райдеры из 10 стран мира. Зрелищное шоу на грани возможного.', 6, 'https://www.msk.kp.ru/daily/27767.5/5225437/')
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
        ('russian-cup-final-spartak-krasnodar-2026', 'Стадион «Лужники»'),
        ('kharlamov-cup-final-spartak-loko-2026', 'ЛД «Легенды хоккея — Сокольники»'),
        ('vtb-united-league-cska-lokomotiv-2026', 'Дворец спорта «Мегаспорт»'),
        ('sportland-festival-2026', 'Стадион «Искра»'),
        ('moscow-half-marathon-2026', 'Улица Косыгина / МГУ — «Лужники»'),
        ('moscow-ski-marathon-2026', 'СК «Альфа-Битца»'),
        ('night-of-moscow-sport-2026', 'Парк Горького, Арбат'),
        ('proryv-extreme-sports-festival-2026', 'ЦСКА Арена')
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
        ('russian-cup-final-spartak-krasnodar-2026', 'Стадион «Лужники»', '2026-05-24 18:00:00+03', '2026-05-24 22:00:00+03', 0),
        ('kharlamov-cup-final-spartak-loko-2026', 'ЛД «Легенды хоккея — Сокольники»', '2026-05-24 13:00:00+03', '2026-05-24 16:00:00+03', 0),
        ('vtb-united-league-cska-lokomotiv-2026', 'Дворец спорта «Мегаспорт»', '2026-05-23 17:00:00+03', '2026-05-23 20:00:00+03', 0),
        ('sportland-festival-2026', 'Стадион «Искра»', '2026-06-06 11:00:00+03', '2026-06-06 20:00:00+03', 0),
        ('sportland-festival-2026', 'Стадион «Искра»', '2026-06-07 11:00:00+03', '2026-06-07 20:00:00+03', 0),
        ('moscow-half-marathon-2026', 'Улица Косыгина / МГУ — «Лужники»', '2026-04-25 09:00:00+03', '2026-04-25 17:00:00+03', 0),
        ('moscow-half-marathon-2026', 'Улица Косыгина / МГУ — «Лужники»', '2026-04-26 09:00:00+03', '2026-04-26 17:00:00+03', 0),
        ('moscow-ski-marathon-2026', 'СК «Альфа-Битца»', '2026-02-08 11:00:00+03', '2026-02-08 16:00:00+03', 0),
        ('night-of-moscow-sport-2026', 'Парк Горького, Арбат', '2026-05-30 19:00:00+03', '2026-05-30 22:00:00+03', 0),
        ('proryv-extreme-sports-festival-2026', 'ЦСКА Арена', '2026-03-21 17:00:00+03', '2026-03-21 22:00:00+03', 0)
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
    'russian-cup-final-spartak-krasnodar-2026',
    'kharlamov-cup-final-spartak-loko-2026',
    'vtb-united-league-cska-lokomotiv-2026',
    'sportland-festival-2026',
    'moscow-half-marathon-2026',
    'moscow-ski-marathon-2026',
    'night-of-moscow-sport-2026',
    'proryv-extreme-sports-festival-2026'
  );

INSERT INTO event_image (event_id, image_url)
SELECT e.id, v.image_url
FROM (
    VALUES
        ('russian-cup-final-spartak-krasnodar-2026', '/uploads/events/russian-cup-final-spartak-krasnodar-2026-01.jpg'),
        ('kharlamov-cup-final-spartak-loko-2026', '/uploads/events/kharlamov-cup-final-spartak-loko-2026-01.jpg'),
        ('vtb-united-league-cska-lokomotiv-2026', '/uploads/events/vtb-united-league-cska-lokomotiv-2026-01.jpg'),
        ('sportland-festival-2026', '/uploads/events/sportland-festival-2026-01.jpg'),
        ('moscow-half-marathon-2026', '/uploads/events/moscow-half-marathon-2026-01.jpg'),
        ('moscow-ski-marathon-2026', '/uploads/events/moscow-ski-marathon-2026-01.jpg'),
        ('night-of-moscow-sport-2026', '/uploads/events/night-of-moscow-sport-2026-01.jpg'),
        ('proryv-extreme-sports-festival-2026', '/uploads/events/proryv-extreme-sports-festival-2026-01.jpg')
) AS v(event_slug, image_url)
JOIN event e ON e.slug = v.event_slug
ON CONFLICT (event_id, image_url) DO NOTHING;

-- 9. Категория «Спортивные мероприятия»
INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM event e
JOIN category c ON c.name = 'Спортивные мероприятия'
WHERE e.slug IN (
    'russian-cup-final-spartak-krasnodar-2026',
    'kharlamov-cup-final-spartak-loko-2026',
    'vtb-united-league-cska-lokomotiv-2026',
    'sportland-festival-2026',
    'moscow-half-marathon-2026',
    'moscow-ski-marathon-2026',
    'night-of-moscow-sport-2026',
    'proryv-extreme-sports-festival-2026'
)
ON CONFLICT (event_id, category_id) DO NOTHING;

-- 10. Теги событий
INSERT INTO event_tag (event_id, tag_id)
SELECT e.id, t.id
FROM (
    VALUES
        ('russian-cup-final-spartak-krasnodar-2026', 'Спортивные мероприятия'), ('russian-cup-final-spartak-krasnodar-2026', 'Футбол'), ('russian-cup-final-spartak-krasnodar-2026', 'Профессиональный спорт'), ('russian-cup-final-spartak-krasnodar-2026', 'Зрелищное шоу'),
        ('kharlamov-cup-final-spartak-loko-2026', 'Спортивные мероприятия'), ('kharlamov-cup-final-spartak-loko-2026', 'Хоккей'), ('kharlamov-cup-final-spartak-loko-2026', 'Профессиональный спорт'),
        ('vtb-united-league-cska-lokomotiv-2026', 'Спортивные мероприятия'), ('vtb-united-league-cska-lokomotiv-2026', 'Баскетбол'), ('vtb-united-league-cska-lokomotiv-2026', 'Профессиональный спорт'),
        ('sportland-festival-2026', 'Спортивные мероприятия'), ('sportland-festival-2026', 'Фестиваль'), ('sportland-festival-2026', 'Семейный'), ('sportland-festival-2026', 'Open air'),
        ('moscow-half-marathon-2026', 'Спортивные мероприятия'), ('moscow-half-marathon-2026', 'Бег'), ('moscow-half-marathon-2026', 'Фестиваль'),
        ('moscow-ski-marathon-2026', 'Спортивные мероприятия'), ('moscow-ski-marathon-2026', 'Лыжи'), ('moscow-ski-marathon-2026', 'Зимний спорт'), ('moscow-ski-marathon-2026', 'Фестиваль'),
        ('night-of-moscow-sport-2026', 'Спортивные мероприятия'), ('night-of-moscow-sport-2026', 'Фестиваль'), ('night-of-moscow-sport-2026', 'Семейный'), ('night-of-moscow-sport-2026', 'Open air'),
        ('proryv-extreme-sports-festival-2026', 'Спортивные мероприятия'), ('proryv-extreme-sports-festival-2026', 'Экстремальный спорт'), ('proryv-extreme-sports-festival-2026', 'Фестиваль'), ('proryv-extreme-sports-festival-2026', 'Зрелищное шоу')
) AS v(event_slug, tag_name)
JOIN event e ON e.slug = v.event_slug
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;

COMMIT;
