-- CityHawk seed: категория «Выставки»
-- Данные подготовлены из пользовательского списка событий.
-- Изображения проставлены по порядку снизу вверх из переданного набора: сначала событие 5, затем 4, 3, 2, 1.

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
VALUES ('Выставки')
ON CONFLICT (name) DO NOTHING;

INSERT INTO tag (name)
VALUES
    ('Выставка'),
    ('Иммерсивное'),
    ('Мультимедиа'),
    ('Искусственный интеллект'),
    ('Гастрономия'),
    ('Искусство'),
    ('Ренессанс'),
    ('Аудиогид'),
    ('История'),
    ('Экскурсия'),
    ('Подземелье'),
    ('Современное искусство'),
    ('Стекло'),
    ('Космос'),
    ('Ювелирное искусство'),
    ('Музей'),
    ('Для фото')
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
        ('Арт-пространство Futurione', 'Москва, 2-я Останкинская ул., 3, МФК «Солнце Москвы»', 55.826089::numeric, 37.628936::numeric, 'Интерактивное арт-пространство в МФК «Солнце Москвы», где цифровое искусство, иммерсивные залы и новые технологии объединены в единый маршрут.'),
        ('Музей АртДинамикс', 'Москва, Нижний Сусальный пер., 5, стр. 10, БЦ «Арма»', 55.760661::numeric, 37.664384::numeric, 'Мультимедийный музей в пространстве БЦ «Арма» с иммерсивными цифровыми проектами, проекциями и аудиогидами.'),
        ('Бункер-42 на Таганке', 'Москва, 5-й Котельнический пер., 11', 55.741728::numeric, 37.649292::numeric, 'Музей Холодной войны на глубине 65 метров с подземными тоннелями, гермодверями, экспонатами связи и интерактивными зонами.'),
        ('Heritage', 'Москва, ул. Петровка, 20/1, вход с Петровских Линий', 55.764222::numeric, 37.617321::numeric, 'Галерея современного и декоративно-прикладного искусства в центре Москвы, работающая с collectible design и авторскими выставочными проектами.'),
        ('Исторический музей', 'Москва, Красная пл., 1', 55.755277::numeric, 37.617680::numeric, 'Главное здание Государственного исторического музея на Красной площади с масштабными историческими и художественными экспозициями.')
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
        ('Futurione', 'futurione-2026', 'Арт-пространство Futurione', 'FUTURIONE — пространство, где будущее и настоящее соединяются в иммерсивное приключение. Здесь каждое посещение превращается в источник вдохновения, восторга и ярких впечатлений, а созданные моменты остаются в памяти надолго. Отдельный формат проекта — «Тайная комната»: церемониальное иммерсивное пространство с новым гастрономическим опытом и встречей за одним столом с искусственным интеллектом. Для посещения тайной комнаты нужен отдельный билет, который можно приобрести в кассе на месте проведения. Общее прохождение выставки занимает около 50 минут. Посещение возможно только в указанные в билете день и время.', 0, 'https://vdnh.ru/places/interaktivnoe-prostranstvo-futurione-/'),
        ('Секреты Джузеппе Арчимбольдо', 'sekrety-dzhuzeppe-archimboldo-2026', 'Музей АртДинамикс', 'Иммерсивный мультимедийный проект, предлагающий по-новому взглянуть на аллегористические портреты Джузеппе Арчимбольдо. Объемные 3D-картины, цифровые проекции и звук создают эффект присутствия внутри фантазийных инсталляций, наполненных предметами, животными, рыбами, цветами и фруктами. Работы художника превращаются в анимационные сюжеты и визуальные головоломки, а аудиогид помогает погрузиться в психологический анализ образов, составленный кандидатом психологических наук. Для комфортного посещения рекомендуется взять наушники. Выставка проходит ежедневно с 12:00 до 22:00.', 12, 'https://artdynamics.ru/msk'),
        ('Экскурсия «Бункер Сталина»', 'ekskursiya-bunker-stalina-2026', 'Бункер-42 на Таганке', 'Экскурсия в легендарный подземный объект на глубине 65 метров — музей Холодной войны «Бункер-42» на Таганке. Посетители увидят гермодвери, секретные тоннели, аппаратуру связи и узнают, как было устроено убежище, рассчитанное на работу в условиях ядерной угрозы. В программе — рассказ о быте людей, несших круглосуточное дежурство, макет бомбардировщика Ту-4, полноразмерная копия атомной бомбы РДС-1 и кабинет Иосифа Сталина. В экспозиции есть интерактивные зоны, спецэффекты и экспонаты, которые можно трогать и фотографировать.', 12, 'https://bunker42.com/contacts/'),
        ('Орбиты хрупких тел', 'orbity-khrupkikh-tel-2026', 'Heritage', 'Выставочный проект галереи Heritage о современном художественном стекле и космосе как метафоре бесконечности, преодоления, внутренней свободы и движения вопреки обстоятельствам. Куратор проекта — коллекционер и издатель Марина Добровинская. Экспозиция продолжает разговор о судьбе российского стеклоделия и представляет произведения от второй половины XX века до наших дней, включая работы мастеров экспериментальных лабораторий крупнейших стекольных предприятий СССР и современных авторов. Проект выстраивает диалог между историческим наследием, заводской школой, авторским высказыванием и философской темой выхода за границы возможного.', 0, 'https://heritage-gallery.ru/contacts'),
        ('Очарование красоты', 'ocharovanie-krasoty-2026', 'Исторический музей', 'Выставка «Очарование красоты. Из истории ювелирных украшений в России» — масштабный показ части коллекции ювелирных украшений из собрания Особой кладовой Государственного исторического музея. Экспозиция раскрывает историю украшений в России со второй половины XVII до начала XX века: серьги, кольца, броши, браслеты, ожерелья, колье, подвески, кулоны, тиары и гребни демонстрируют разнообразие форм, материалов, орнаментов и ювелирных техник. Особое внимание уделено украшениям с камеями, мозаикой, бирюзой, горным хрусталем, гранатами и кораллами, а также сентиментальным украшениям, отражающим вкус и эмоциональную культуру эпохи.', 0, 'https://shm.ru/')
) AS v(title, slug, location_description, full_description, age_limit, source_url)
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
        ('futurione-2026', 'Арт-пространство Futurione'),
        ('sekrety-dzhuzeppe-archimboldo-2026', 'Музей АртДинамикс'),
        ('ekskursiya-bunker-stalina-2026', 'Бункер-42 на Таганке'),
        ('orbity-khrupkikh-tel-2026', 'Heritage'),
        ('ocharovanie-krasoty-2026', 'Исторический музей')
) AS v(event_slug, place_name)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id) DO UPDATE
SET place_id = EXCLUDED.place_id,
    updated_at = now();

-- 6. Сессии выставок
-- Для выставочного формата start_at/end_at используются как период проведения.
INSERT INTO event_session (event_id, place_id, start_at, end_at, price)
SELECT e.id, p.id, v.start_at::timestamptz, v.end_at::timestamptz, v.price
FROM (
    VALUES
        ('futurione-2026', 'Арт-пространство Futurione', '2026-05-28 10:00:00+03', '2026-06-28 22:00:00+03', 0),
        ('sekrety-dzhuzeppe-archimboldo-2026', 'Музей АртДинамикс', '2026-05-29 12:00:00+03', '2026-06-29 22:00:00+03', 950),
        ('ekskursiya-bunker-stalina-2026', 'Бункер-42 на Таганке', '2026-05-30 10:00:00+03', '2026-06-30 21:00:00+03', 0),
        ('orbity-khrupkikh-tel-2026', 'Heritage', '2026-05-31 12:00:00+03', '2026-06-30 20:00:00+03', 350),
        ('ocharovanie-krasoty-2026', 'Исторический музей', '2026-05-28 10:00:00+03', '2026-06-30 19:00:00+03', 0)
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
        -- Порядок снизу вверх: 5. «Очарование красоты» — 4 фото
        ('ocharovanie-krasoty-2026', '/uploads/events/68f7e1c378b340038eaeb047cb95.png'),
        ('ocharovanie-krasoty-2026', '/uploads/events/08dfc8a87432401382482ee558e9.jpg'),
        ('ocharovanie-krasoty-2026', '/uploads/events/ca37e96c9238424caa066d95e063.jpg'),
        ('ocharovanie-krasoty-2026', '/uploads/events/0213eb18d7ac4fc3a03b94665511.jpg'),

        -- 4. «Орбиты хрупких тел» — 4 фото
        ('orbity-khrupkikh-tel-2026', '/uploads/events/afdd7c05a865468e831545fbc038.jpg'),
        ('orbity-khrupkikh-tel-2026', '/uploads/events/ac038986004f4efdbf615361bcc2.jpg'),
        ('orbity-khrupkikh-tel-2026', '/uploads/events/1c8c168783b9430aa243b228bf81.jpg'),
        ('orbity-khrupkikh-tel-2026', '/uploads/events/168f814668f24e568601afd6aa31.jpg'),

        -- 3. «Экскурсия “Бункер Сталина”» — 1 фото
        ('ekskursiya-bunker-stalina-2026', '/uploads/events/ec0341f54e394256b34f1b6a650f.jpg'),

        -- 2. «Секреты Джузеппе Арчимбольдо» — 4 фото
        ('sekrety-dzhuzeppe-archimboldo-2026', '/uploads/events/c4fa333bbcb143b58d4288b78e5a.jpg'),
        ('sekrety-dzhuzeppe-archimboldo-2026', '/uploads/events/a324058ac59e469a9524ebc32d25.jpg'),
        ('sekrety-dzhuzeppe-archimboldo-2026', '/uploads/events/7c02bdd93397478da1c3f75060a3.jpg'),
        ('sekrety-dzhuzeppe-archimboldo-2026', '/uploads/events/2473b0e296d2466095d54dd58c77.jpg'),

        -- 1. Futurione — 4 фото
        ('futurione-2026', '/uploads/events/03ff3d64d73a4fab9f26a6fa24d1-2.jpg'),
        ('futurione-2026', '/uploads/events/18f287dba1a248cbaa6ae7946250.jpg'),
        ('futurione-2026', '/uploads/events/62ef72fad8774c0e97c0015e4909.jpg'),
        ('futurione-2026', '/uploads/events/0a939959293048ceace8436f7ed5.jpg')
) AS v(event_slug, image_url)
JOIN event e ON e.slug = v.event_slug
ON CONFLICT (event_id, image_url) DO NOTHING;

-- 8. Категория «Выставки»
INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM event e
JOIN category c ON c.name = 'Выставки'
WHERE e.slug IN (
    'futurione-2026',
    'sekrety-dzhuzeppe-archimboldo-2026',
    'ekskursiya-bunker-stalina-2026',
    'orbity-khrupkikh-tel-2026',
    'ocharovanie-krasoty-2026'
)
ON CONFLICT (event_id, category_id) DO NOTHING;

-- 9. Теги событий
INSERT INTO event_tag (event_id, tag_id)
SELECT e.id, t.id
FROM (
    VALUES
        ('futurione-2026', 'Выставка'),
        ('futurione-2026', 'Иммерсивное'),
        ('futurione-2026', 'Искусственный интеллект'),
        ('futurione-2026', 'Гастрономия'),
        ('futurione-2026', 'Для фото'),

        ('sekrety-dzhuzeppe-archimboldo-2026', 'Выставка'),
        ('sekrety-dzhuzeppe-archimboldo-2026', 'Иммерсивное'),
        ('sekrety-dzhuzeppe-archimboldo-2026', 'Мультимедиа'),
        ('sekrety-dzhuzeppe-archimboldo-2026', 'Ренессанс'),
        ('sekrety-dzhuzeppe-archimboldo-2026', 'Аудиогид'),

        ('ekskursiya-bunker-stalina-2026', 'Экскурсия'),
        ('ekskursiya-bunker-stalina-2026', 'История'),
        ('ekskursiya-bunker-stalina-2026', 'Подземелье'),
        ('ekskursiya-bunker-stalina-2026', 'Музей'),

        ('orbity-khrupkikh-tel-2026', 'Выставка'),
        ('orbity-khrupkikh-tel-2026', 'Современное искусство'),
        ('orbity-khrupkikh-tel-2026', 'Стекло'),
        ('orbity-khrupkikh-tel-2026', 'Космос'),

        ('ocharovanie-krasoty-2026', 'Выставка'),
        ('ocharovanie-krasoty-2026', 'Ювелирное искусство'),
        ('ocharovanie-krasoty-2026', 'История'),
        ('ocharovanie-krasoty-2026', 'Музей')
) AS v(event_slug, tag_name)
JOIN event e ON e.slug = v.event_slug
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;

COMMIT;
