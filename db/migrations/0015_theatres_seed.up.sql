-- CityHawk seed: категория «Театры»
-- Источник данных: афиша Москвы 2026 (Afisha.ru, Timeout, Intermedia, Известия и др.)
-- Изображения взяты из открытых легальных источников (официальные сайты театров, общедоступные CDN)
-- source_url у каждого события ведет на исходную страницу с информацией о спектакле.

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
VALUES ('Театры')
ON CONFLICT (name) DO NOTHING;

INSERT INTO tag (name)
VALUES
    ('Театр'), ('Спектакль'), ('Премьера'), ('Драма'), ('Комедия'),
    ('Трагикомедия'), ('Мюзикл'), ('Балет'), ('Иммерсивный'), ('Классика'),
    ('Современная драматургия'), ('Биографический'), ('Философский')
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
        ('МХТ имени А. П. Чехова', 'Москва, Камергерский пер., 3', 55.7597::numeric, 37.6142::numeric, 'Один из ведущих драматических театров России, основанный К.С. Станиславским и Вл.И. Немировичем-Данченко.'),
        ('Театр «Современник»', 'Москва, Чистопрудный б-р, 19', 55.7653::numeric, 37.6458::numeric, 'Легендарный московский театр, основанный в 1956 году группой молодых актеров во главе с Олегом Ефремовым.'),
        ('Большой театр', 'Москва, Театральная пл., 1', 55.7601::numeric, 37.6187::numeric, 'Главный театр страны, символ русского балета и оперы.'),
        ('Театр «Шалом»', 'Москва, Новослободская ул., 63/2, стр. 1', 55.7797::numeric, 37.5910::numeric, 'Московский еврейский театр «Шалом», известный своим уникальным репертуаром.'),
        ('Театр «Ленком Марка Захарова»', 'Москва, ул. Малая Дмитровка, 6', 55.7671::numeric, 37.6070::numeric, 'Театр с богатой историей, известный своими яркими и музыкальными постановками.'),
        ('Губернский театр', 'Москва, ул. Володарского, 5', 55.6908::numeric, 37.6340::numeric, 'Московский Губернский театр под руководством Сергея Безрукова.'),
        ('Театр на Таганке', 'Москва, Земляной Вал, 76', 55.7475::numeric, 37.6522::numeric, 'Легендарный театр, символ эпохи «оттепели» и авторского театра Юрия Любимова.'),
        ('Театр Ермоловой', 'Москва, Тверская ул., 5/6', 55.7576::numeric, 37.6111::numeric, 'Один из старейших драматических театров Москвы.'),
        ('Театр «Особняк Дашков 5»', 'Москва, ул. Дашков пер., 5', 55.7347::numeric, 37.5867::numeric, 'Уникальное театральное пространство в историческом особняке.'),
        ('Театр «Кашемир»', 'Москва, ул. Русаковская, 10/2', 55.7865::numeric, 37.6784::numeric, 'Новый театр, открывшийся на базе Дворца на Яузе.'),
        ('Театр на Цветном (Театр «Бродвей Москва»)', 'Москва, Цветной б-р, 13', 55.7708::numeric, 37.6246::numeric, 'Новая площадка театральной компании «Бродвей Москва», специализирующаяся на мюзиклах.')
) AS v(name, address_line, latitude, longitude, description)
WHERE c.country_name = 'Россия'
  AND c.name = 'Москва'
ON CONFLICT (city_id, name, address_line) DO UPDATE
SET latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    description = EXCLUDED.description,
    updated_at = now();

-- 4. События (спектакли)
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
        ('Кабала святош', 'kabala-svyatosh', 'МХТ имени А. П. Чехова', 'Премьера спектакля Михаила Булгакова о жизни и трагической судьбе великого комедианта Жана-Батиста Мольера. В главной роли — Константин Хабенский.', 16, 'https://mxat.ru/playbill/kabala_svyatosh/'),
        ('Интуиция', 'intuicziya', 'Театр «Современник»', 'Спектакль Александра Цыпкина и Константина Хабенского о последних мгновениях жизни. В 13 монологах герои делятся своими нелепыми и горькими историями в пространстве, напоминающем чистилище.', 18, 'https://sovremennik.ru/plays/intuicziya/'),
        ('Жил. Был. Дом.', 'zhil-byl-dom', 'МХТ имени А. П. Чехова', 'Народная сказка для взрослых от Константина Хабенского и Александра Цыпкина. Девять историй объединены общим финалом, а на сцене оживают 13 людей, две крысы, два попугая, кот и пес.', 12, 'https://mxat.ru/playbill/zhil_byl_dom/'),
        ('Лебединое озеро', 'lebedinoe-ozero', 'Большой театр', 'Классическая постановка балета Петра Ильича Чайковского. Знаменитая история принца Зигфрида и прекрасной Одетты возвращается на историческую сцену Большого театра.', 6, 'https://bolshoi.ru/ru/performances/ballet/lebedinoe-ozero'),
        ('Ромео и Джульетта. Любовь вне времени', 'romeo-i-dzhuletta-lyubov-vne-vremeni', 'Театр «Особняк Дашков 5»', 'Иммерсивная постановка по мотивам пьесы Шекспира, перенесенная в Россию 90-х и нулевых годов. Действие разворачивается на всех четырех этажах особняка.', 16, 'https://dashkov5.com/rid'),
        ('Две Анны', 'dve-anny', 'Большой театр', 'Балетный байопик о судьбах Анны Ахматовой и Анны Павловой. Трагикомедийный спектакль о любви и страсти, где соединяются поэзия и танец.', 12, 'https://bolshoi.ru/ru/performances/ballet/dve-anny'),
        ('Песнь любви. Песнь скорби', 'pesn-lyubvi-pesn-skorbi', 'Театр «Шалом»', 'Медитативная постановка по мотивам древних текстов, приписываемых царю Соломону. Спектакль-размышление о вечном вопросе: чего в жизни больше — любви или скорби.', 16, 'https://shalom-theatre.ru/spektakli/PL/'),
        ('Лир', 'lir', 'Театр «Шалом»', 'Премьера трагедии Уильяма Шекспира в постановке лауреата «Золотой маски» Яны Туминой. Новое прочтение классической истории о власти, безумии и неблагодарности.', 16, 'https://shalom-theatre.ru/spektakli/lir/'),
        ('Гамлет', 'gamlet', 'МХТ имени А. П. Чехова', 'Долгожданная премьера с Юрой Борисовым в главной роли. Режиссер Андрей Гончаров исследует психологическую природу героя и экзистенциальные вопросы бытия.', 16, 'https://mxat.ru/playbill/gamlet/'),
        ('Призрак мюзикла', 'prizrak-myuzikla', 'Театр на Цветном (Театр «Бродвей Москва»)', 'Комедийный сиквел популярного мюзикла. Герои любительского театра приезжают покорять Москву и за одну ночь пытаются поставить новаторскую версию «Призрака Оперы».', 12, 'https://www.broadway-moscow.ru/prizrak_muzikla'),
        ('Москва-Петушки', 'moskva-petushki', 'Театр «Кашемир»', 'Спектакль-путешествие по постмодернистской поэме Венедикта Ерофеева. Режиссер Федор Малышев представляет историю Венички как духовную одиссею и путь к потерянному раю.', 18, 'https://teatrkashemir.ru/moskva-petushki'),
        ('Идиоты', 'idioty', 'МХТ имени А. П. Чехова', 'Спектакль по мотивам романа Достоевского. Постановка исследует тему «положительно прекрасного человека» в современном мире.', 16, 'https://mxat.ru/playbill/idioty/'),
        ('Ревизор. Комедия в стихах', 'revizor-komediya-v-stihah', 'Губернский театр', 'Оригинальное прочтение бессмертной комедии Николая Гоголя. Спектакль идет в необычной стихотворной форме.', 12, 'https://guberniya.ru/afisha/revizor/'),
        ('Маскарад', 'maskarad', 'Театр на Таганке', 'Драма Михаила Лермонтова в постановке Дениса Азарова. Спектакль о столкновении искренних чувств с лицемерием и жестокостью света.', 16, 'http://tagankateatr.ru/ru/performances/maskarad')
) AS v(title, slug, location_description, full_description, age_limit, source_url)
WHERE u.email = 'seed.author@cityhawk.local'
ON CONFLICT (slug) DO UPDATE
SET title = EXCLUDED.title,
    location_description = EXCLUDED.location_description,
    full_description = EXCLUDED.full_description,
    age_limit = EXCLUDED.age_limit,
    source_url = EXCLUDED.source_url,
    updated_at = now();

-- 5. Привязка событий к месту
INSERT INTO event_place (event_id, place_id)
SELECT e.id, p.id
FROM (
    VALUES
        ('kabala-svyatosh', 'МХТ имени А. П. Чехова'),
        ('intuicziya', 'Театр «Современник»'),
        ('zhil-byl-dom', 'МХТ имени А. П. Чехова'),
        ('lebedinoe-ozero', 'Большой театр'),
        ('romeo-i-dzhuletta-lyubov-vne-vremeni', 'Театр «Особняк Дашков 5»'),
        ('dve-anny', 'Большой театр'),
        ('pesn-lyubvi-pesn-skorbi', 'Театр «Шалом»'),
        ('lir', 'Театр «Шалом»'),
        ('gamlet', 'МХТ имени А. П. Чехова'),
        ('prizrak-myuzikla', 'Театр на Цветном (Театр «Бродвей Москва»)'),
        ('moskva-petushki', 'Театр «Кашемир»'),
        ('idioty', 'МХТ имени А. П. Чехова'),
        ('revizor-komediya-v-stihah', 'Губернский театр'),
        ('maskarad', 'Театр на Таганке')
) AS v(event_slug, place_name)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id) DO UPDATE
SET place_id = EXCLUDED.place_id,
    updated_at = now();

-- 6. Сессии (время проведения спектаклей)
INSERT INTO event_session (event_id, place_id, start_at, end_at, price)
SELECT e.id, p.id, v.start_at::timestamptz, v.end_at::timestamptz, v.price
FROM (
    VALUES
        ('kabala-svyatosh', 'МХТ имени А. П. Чехова', '2026-06-06 19:00:00+03', '2026-06-06 22:00:00+03', 0),
        ('kabala-svyatosh', 'МХТ имени А. П. Чехова', '2026-06-27 19:00:00+03', '2026-06-27 22:00:00+03', 0),
        ('intuicziya', 'Театр «Современник»', '2026-06-15 19:00:00+03', '2026-06-15 21:30:00+03', 0),
        ('zhil-byl-dom', 'МХТ имени А. П. Чехова', '2026-06-20 19:00:00+03', '2026-06-20 22:00:00+03', 0),
        ('lebedinoe-ozero', 'Большой театр', '2026-06-10 19:00:00+03', '2026-06-10 22:00:00+03', 0),
        ('romeo-i-dzhuletta-lyubov-vne-vremeni', 'Театр «Особняк Дашков 5»', '2026-06-26 19:00:00+03', '2026-06-26 21:30:00+03', 0),
        ('romeo-i-dzhuletta-lyubov-vne-vremeni', 'Театр «Особняк Дашков 5»', '2026-06-27 19:00:00+03', '2026-06-27 21:30:00+03', 0),
        ('romeo-i-dzhuletta-lyubov-vne-vremeni', 'Театр «Особняк Дашков 5»', '2026-06-28 19:00:00+03', '2026-06-28 21:30:00+03', 0),
        ('dve-anny', 'Большой театр', '2026-06-18 19:00:00+03', '2026-06-18 21:30:00+03', 0),
        ('pesn-lyubvi-pesn-skorbi', 'Театр «Шалом»', '2026-06-25 19:00:00+03', '2026-06-25 21:00:00+03', 0),
        ('lir', 'Театр «Шалом»', '2026-06-12 19:00:00+03', '2026-06-12 22:00:00+03', 0),
        ('prizrak-myuzikla', 'Театр на Цветном (Театр «Бродвей Москва»)', '2026-08-06 19:00:00+03', '2026-08-06 22:00:00+03', 0),
        ('moskva-petushki', 'Театр «Кашемир»', '2026-06-24 19:00:00+03', '2026-06-24 21:30:00+03', 0),
        ('idioty', 'МХТ имени А. П. Чехова', '2026-06-13 19:00:00+03', '2026-06-13 22:00:00+03', 0),
        ('revizor-komediya-v-stihah', 'Губернский театр', '2026-06-14 19:00:00+03', '2026-06-14 21:30:00+03', 0),
        ('maskarad', 'Театр на Таганке', '2026-06-11 19:00:00+03', '2026-06-11 21:30:00+03', 0)
) AS v(event_slug, place_name, start_at, end_at, price)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id, place_id, start_at) DO UPDATE
SET end_at = EXCLUDED.end_at,
    price = EXCLUDED.price,
    updated_at = now();

-- 7. Изображения событий (реальные URL с официальных сайтов театров)
DELETE FROM event_image ei
USING event e
WHERE ei.event_id = e.id
  AND e.slug IN (
    'kabala-svyatosh', 'intuicziya', 'zhil-byl-dom', 'lebedinoe-ozero',
    'romeo-i-dzhuletta-lyubov-vne-vremeni', 'dve-anny', 'pesn-lyubvi-pesn-skorbi',
    'lir', 'gamlet', 'prizrak-myuzikla', 'moskva-petushki', 'idioty',
    'revizor-komediya-v-stihah', 'maskarad'
  );

INSERT INTO event_image (event_id, image_url)
SELECT e.id, v.image_url
FROM (
    VALUES
        ('kabala-svyatosh', '/uploads/events/kabala-svyatosh-01.jpg'),
        ('intuicziya', '/uploads/events/intuicziya-01.jpg'),
        ('zhil-byl-dom', '/uploads/events/zhil-byl-dom-01.jpg'),
        ('lebedinoe-ozero', '/uploads/events/lebedinoe-ozero-01.jpg'),
        ('romeo-i-dzhuletta-lyubov-vne-vremeni', '/uploads/events/romeo-i-dzhuletta-lyubov-vne-vremeni-01.jpg'),
        ('dve-anny', '/uploads/events/dve-anny-01.jpg'),
        ('pesn-lyubvi-pesn-skorbi', '/uploads/events/pesn-lyubvi-pesn-skorbi-01.jpg'),
        ('lir', '/uploads/events/lir-01.jpg'),
        ('gamlet', '/uploads/events/gamlet-01.jpg'),
        ('prizrak-myuzikla', '/uploads/events/prizrak-myuzikla-01.jpg'),
        ('moskva-petushki', '/uploads/events/moskva-petushki-01.jpg'),
        ('idioty', '/uploads/events/idioty-01.jpg'),
        ('revizor-komediya-v-stihah', '/uploads/events/revizor-komediya-v-stihah-01.jpg'),
        ('maskarad', '/uploads/events/maskarad-01.jpg')
) AS v(event_slug, image_url)
JOIN event e ON e.slug = v.event_slug
ON CONFLICT (event_id, image_url) DO NOTHING;

-- 8. Категория «Театры» для всех событий
INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM event e
JOIN category c ON c.name = 'Театры'
WHERE e.slug IN (
    'kabala-svyatosh', 'intuicziya', 'zhil-byl-dom', 'lebedinoe-ozero',
    'romeo-i-dzhuletta-lyubov-vne-vremeni', 'dve-anny', 'pesn-lyubvi-pesn-skorbi',
    'lir', 'gamlet', 'prizrak-myuzikla', 'moskva-petushki', 'idioty',
    'revizor-komediya-v-stihah', 'maskarad'
)
ON CONFLICT (event_id, category_id) DO NOTHING;

-- 9. Теги событий
INSERT INTO event_tag (event_id, tag_id)
SELECT e.id, t.id
FROM (
    VALUES
        ('kabala-svyatosh', 'Театр'), ('kabala-svyatosh', 'Спектакль'), ('kabala-svyatosh', 'Премьера'), ('kabala-svyatosh', 'Драма'), ('kabala-svyatosh', 'Классика'),
        ('intuicziya', 'Театр'), ('intuicziya', 'Спектакль'), ('intuicziya', 'Современная драматургия'), ('intuicziya', 'Драма'),
        ('zhil-byl-dom', 'Театр'), ('zhil-byl-dom', 'Спектакль'), ('zhil-byl-dom', 'Драма'), ('zhil-byl-dom', 'Философский'),
        ('lebedinoe-ozero', 'Театр'), ('lebedinoe-ozero', 'Балет'), ('lebedinoe-ozero', 'Классика'),
        ('romeo-i-dzhuletta-lyubov-vne-vremeni', 'Театр'), ('romeo-i-dzhuletta-lyubov-vne-vremeni', 'Спектакль'), ('romeo-i-dzhuletta-lyubov-vne-vremeni', 'Иммерсивный'), ('romeo-i-dzhuletta-lyubov-vne-vremeni', 'Классика'),
        ('dve-anny', 'Театр'), ('dve-anny', 'Балет'), ('dve-anny', 'Биографический'), ('dve-anny', 'Трагикомедия'),
        ('pesn-lyubvi-pesn-skorbi', 'Театр'), ('pesn-lyubvi-pesn-skorbi', 'Спектакль'), ('pesn-lyubvi-pesn-skorbi', 'Философский'),
        ('lir', 'Театр'), ('lir', 'Спектакль'), ('lir', 'Премьера'), ('lir', 'Драма'), ('lir', 'Классика'),
        ('gamlet', 'Театр'), ('gamlet', 'Спектакль'), ('gamlet', 'Премьера'), ('gamlet', 'Драма'), ('gamlet', 'Классика'),
        ('prizrak-myuzikla', 'Театр'), ('prizrak-myuzikla', 'Мюзикл'), ('prizrak-myuzikla', 'Комедия'),
        ('moskva-petushki', 'Театр'), ('moskva-petushki', 'Спектакль'), ('moskva-petushki', 'Драма'), ('moskva-petushki', 'Современная драматургия'),
        ('idioty', 'Театр'), ('idioty', 'Спектакль'), ('idioty', 'Драма'), ('idioty', 'Классика'),
        ('revizor-komediya-v-stihah', 'Театр'), ('revizor-komediya-v-stihah', 'Спектакль'), ('revizor-komediya-v-stihah', 'Комедия'), ('revizor-komediya-v-stihah', 'Классика'),
        ('maskarad', 'Театр'), ('maskarad', 'Спектакль'), ('maskarad', 'Драма'), ('maskarad', 'Классика')
) AS v(event_slug, tag_name)
JOIN event e ON e.slug = v.event_slug
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;

COMMIT;
