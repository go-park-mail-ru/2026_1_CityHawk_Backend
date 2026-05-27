-- CityHawk seed: категория «Фестивали»
-- Данные подготовлены из пользовательского списка событий.
-- Изображения проставлены по порядку пользователя: начиная снизу первое фото относится к событию «Random Fest», далее «Фестиваль Солома», «Доброфест» и «DDX Fitness Fest».

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
VALUES ('Фестивали')
ON CONFLICT (name) DO NOTHING;

INSERT INTO tag (name)
VALUES
    ('Фестиваль'),
    ('Музыка'),
    ('Open air'),
    ('Летний фестиваль'),
    ('Хип-хоп'),
    ('Семейный отдых'),
    ('Кинопарк'),
    ('Рок'),
    ('Фитнес'),
    ('Спорт'),
    ('Для детей'),
    ('Хедлайнеры'),
    ('Для компании')
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
        ('Площадь у ЦСКА Арены', 'Москва, ул. Автозаводская, 23а, площадь у ЦСКА Арены', 55.702400::numeric, 37.642600::numeric, 'Городская open-air площадка у ЦСКА Арены для крупных фестивалей, концертов и массовых мероприятий.'),
        ('Кинопарк Москино', 'Москва, Троицкий административный округ, Краснопахорский район, квартал № 107, ул. Лиозновой', 55.421041::numeric, 37.261651::numeric, 'Крупнейшая фестивальная и кинолокация Москвы с декорациями, площадками под открытым небом, экскурсионными маршрутами и инфраструктурой для семейного отдыха.'),
        ('Музей-заповедник «Коломенское»', 'Москва, просп. Андропова, 39', 55.672016::numeric, 37.665010::numeric, 'Исторический парк и музей-заповедник, где проходят городские фестивали, культурные программы и спортивные мероприятия на открытом воздухе.')
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
            'Random Fest',
            'random-fest-2026-moscow',
            'Площадь у ЦСКА Арены',
            'Random Fest станет самым масштабным за всю историю фестиваля: в 2026 году он объединит сразу 8 городов. Большие лайвы, любимые артисты и фирменная атмосфера Random Fest снова соберут тысячи зрителей по всей стране. Московский фестивальный день на площадке у ЦСКА Арены объединит хип-хоп- и поп-артистов новой волны. В лайн-апе 2026 года заявлены OG Buda, MAYOT, Toxi$, 163ONMYNECK, Instrinna и Джон Гарик. Это событие для тех, кто хочет увидеть громкие лайвы, провести день на open-air площадке и погрузиться в атмосферу большого летнего фестиваля.',
            16,
            'https://cska-arena.ru/'
        ),
        (
            'Фестиваль «Солома»',
            'festival-soloma-2026',
            'Кинопарк Москино',
            'Семейный фестиваль «Солома» объединяет кино, творчество и живую музыку на закате на одной территории. Парк открыт для посетителей с 11:00 до 22:00, а дневная программа стартует с 12:00: постановочные съёмки по сценариям любимых фильмов, мастер-класс по аквагриму, диджитал-аттракционы, управление виртуальным персонажем и познавательные экскурсии по территории кинопарка. Кульминация каждого дня — вечерний концерт с 18:00 до 22:00. 27 июня на сцене — Маша Фэй и Zivert, 28 июня — Антоха МС, IOWA и Uma2rman. Гостям рекомендуют взять плед и удобную одежду по погоде. На территории доступны дома-апартаменты с террасами и гостиница, проживание и парковка оплачиваются отдельно. Доехать можно на автомобиле или на бесплатных автобусах от метро «Тёплый Стан». Состав артистов может пополняться, возможны изменения.',
            0,
            'https://kinopark.moskino.ru/'
        ),
        (
            'Доброфест',
            'dobrofest-2026-moskino',
            'Кинопарк Москино',
            '4 и 5 июля Кинопарк Москино зажжёт вместе с главными рокерами страны на фестивале Dobrofest. В программе — только лучшие рок-группы, десятки часов любимой музыки, драйв и энергия, которые зарядят надолго. На главной сцене заявлены «ПилОт», «Алиса», «Слот», «АнимациЯ» и другие легендарные артисты. Среди участников фестиваля: «Алиса», «Пилот», «Слот», «Северный Флот», «Биртман», Jane Air, F.P.G., «Потомучто», «Бригадный Подряд», Sellout, «Операция Пластилин», «Костя Кулясов | АнимациЯ», «Йорш», «Включай Микрофон!», El Mashe, Cardio Killer, The Translators и другие. Помимо концертов гостей ждут кинопоказы под открытым небом, лектории, гастрономическая программа, палаточный лагерь и экскурсии по декорациям Кинопарка. Абонемент на 2 дня даёт неограниченный доступ на территорию фестиваля в течение двух дней, парковочный билет приобретается отдельно.',
            12,
            'https://kinopark.moskino.ru/'
        ),
        (
            'DDX Fitness Fest',
            'ddx-fitness-fest-2026',
            'Музей-заповедник «Коломенское»',
            'DDX Fitness Fest — это фитнес-город под открытым небом в одном из самых знаковых мест Москвы. На фестивале запланировано 150 и более групповых тренировок за 2 дня — от танцевальных и функциональных до медитаций и авторских программ DDX Fitness, более 20 известных блогеров и 3 концерта хедлайнеров. В программе: групповые тренировки, спортивные турниры, зона духовных практик, ниндзя-гонка, призы и подарки, зоны развлечений, детская зона отдыха и фудкорт. 25 июля заявлен секретный хедлайнер, 26 июля — Мари Краймбрери со специальной фестивальной программой и хитами «Океан», «Нравится жить», «Пряталась в ванной», «Мне так хорошо», «Amore» и другими. Для детей от 6 до 15 лет вход бесплатный, для них будут работать детские зоны, аниматоры, аквагрим и другие развлечения. Гостям рекомендуют взять форму, кроссовки, хорошее настроение и карту или наличные для шопинга.',
            6,
            'https://ddxfitness.ru/fest'
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
        ('random-fest-2026-moscow', 'Площадь у ЦСКА Арены'),
        ('festival-soloma-2026', 'Кинопарк Москино'),
        ('dobrofest-2026-moskino', 'Кинопарк Москино'),
        ('ddx-fitness-fest-2026', 'Музей-заповедник «Коломенское»')
) AS v(event_slug, place_name)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id) DO UPDATE
SET place_id = EXCLUDED.place_id,
    updated_at = now();

-- 6. Сессии фестивалей
INSERT INTO event_session (event_id, place_id, start_at, end_at, price)
SELECT e.id, p.id, v.start_at::timestamptz, v.end_at::timestamptz, v.price
FROM (
    VALUES
        -- Random Fest: дата не была указана пользователем, добавлена одна сессия для московского фестивального дня
        ('random-fest-2026-moscow', 'Площадь у ЦСКА Арены', '2026-06-20 16:00:00+03', '2026-06-20 23:00:00+03', 0),

        -- Фестиваль «Солома»
        ('festival-soloma-2026', 'Кинопарк Москино', '2026-06-27 11:00:00+03', '2026-06-27 22:00:00+03', 0),
        ('festival-soloma-2026', 'Кинопарк Москино', '2026-06-28 11:00:00+03', '2026-06-28 22:00:00+03', 0),

        -- Доброфест
        ('dobrofest-2026-moskino', 'Кинопарк Москино', '2026-07-04 12:00:00+03', '2026-07-04 23:00:00+03', 0),
        ('dobrofest-2026-moskino', 'Кинопарк Москино', '2026-07-05 12:00:00+03', '2026-07-05 23:00:00+03', 0),

        -- DDX Fitness Fest
        ('ddx-fitness-fest-2026', 'Музей-заповедник «Коломенское»', '2026-07-25 10:00:00+03', '2026-07-25 22:00:00+03', 0),
        ('ddx-fitness-fest-2026', 'Музей-заповедник «Коломенское»', '2026-07-26 10:00:00+03', '2026-07-26 22:00:00+03', 0)
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
        -- Порядок снизу вверх: Random Fest, Фестиваль «Солома», Доброфест, DDX Fitness Fest
        ('random-fest-2026-moscow', '/uploads/events/orig.jpeg'),
        ('festival-soloma-2026', '/uploads/events/GL_Foto_press-sluzhba_Departamenta_kul_turi_goroda_Moskvi.jpg'),
        ('dobrofest-2026-moskino', '/uploads/events/8b47c8531b1a44b6855aa4b74ca8128a.jpg'),
        ('ddx-fitness-fest-2026', '/uploads/events/orig-2.jpeg')
) AS v(event_slug, image_url)
JOIN event e ON e.slug = v.event_slug
ON CONFLICT (event_id, image_url) DO NOTHING;

-- 8. Категория «Фестивали»
INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM event e
JOIN category c ON c.name = 'Фестивали'
WHERE e.slug IN (
    'random-fest-2026-moscow',
    'festival-soloma-2026',
    'dobrofest-2026-moskino',
    'ddx-fitness-fest-2026'
)
ON CONFLICT (event_id, category_id) DO NOTHING;

-- 9. Теги событий
INSERT INTO event_tag (event_id, tag_id)
SELECT e.id, t.id
FROM (
    VALUES
        ('random-fest-2026-moscow', 'Фестиваль'),
        ('random-fest-2026-moscow', 'Музыка'),
        ('random-fest-2026-moscow', 'Open air'),
        ('random-fest-2026-moscow', 'Летний фестиваль'),
        ('random-fest-2026-moscow', 'Хип-хоп'),
        ('random-fest-2026-moscow', 'Хедлайнеры'),
        ('random-fest-2026-moscow', 'Для компании'),

        ('festival-soloma-2026', 'Фестиваль'),
        ('festival-soloma-2026', 'Музыка'),
        ('festival-soloma-2026', 'Open air'),
        ('festival-soloma-2026', 'Летний фестиваль'),
        ('festival-soloma-2026', 'Семейный отдых'),
        ('festival-soloma-2026', 'Кинопарк'),
        ('festival-soloma-2026', 'Для детей'),

        ('dobrofest-2026-moskino', 'Фестиваль'),
        ('dobrofest-2026-moskino', 'Музыка'),
        ('dobrofest-2026-moskino', 'Open air'),
        ('dobrofest-2026-moskino', 'Летний фестиваль'),
        ('dobrofest-2026-moskino', 'Рок'),
        ('dobrofest-2026-moskino', 'Кинопарк'),
        ('dobrofest-2026-moskino', 'Для компании'),

        ('ddx-fitness-fest-2026', 'Фестиваль'),
        ('ddx-fitness-fest-2026', 'Open air'),
        ('ddx-fitness-fest-2026', 'Летний фестиваль'),
        ('ddx-fitness-fest-2026', 'Фитнес'),
        ('ddx-fitness-fest-2026', 'Спорт'),
        ('ddx-fitness-fest-2026', 'Для детей'),
        ('ddx-fitness-fest-2026', 'Хедлайнеры')
) AS v(event_slug, tag_name)
JOIN event e ON e.slug = v.event_slug
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;

COMMIT;
