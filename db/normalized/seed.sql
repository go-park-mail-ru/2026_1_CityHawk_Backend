BEGIN;

INSERT INTO city (name, country_name, timezone)
VALUES
    ('Москва', 'Россия', 'Europe/Moscow')
ON CONFLICT (country_name, name) DO NOTHING;

INSERT INTO user_account (email, username, user_surname, password_hash, city_id, avatar_url)
SELECT
    'seed.author@cityhawk.local',
    'seed_author',
    'CityHawk',
    '$2a$10$wJv1PLbF6XJz5lG1fV64VeGQFQf5d3M3K2Y6DzG6Q8rB3Yj8wYF1W',
    c.id,
    'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?auto=format&fit=crop&w=400&q=80'
FROM city c
WHERE c.country_name = 'Россия' AND c.name = 'Москва'
ON CONFLICT (email) DO NOTHING;

INSERT INTO category (name)
VALUES
    ('Парк'),
    ('Выставка'),
    ('Семья'),
    ('Шоу'),
    ('Стендап'),
    ('Театр'),
    ('Музей'),
    ('Музыка'),
    ('Квест'),
    ('Фото')
ON CONFLICT (name) DO NOTHING;

INSERT INTO tag (name)
VALUES
    ('Иммерсивное'),
    ('Семейное'),
    ('Ледовое шоу'),
    ('Стендап'),
    ('Балет'),
    ('Искусство'),
    ('Рок'),
    ('Ретро'),
    ('Квест')
ON CONFLICT (name) DO NOTHING;

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
        ('ВДНХ', 'Москва, проспект Мира, 119', 55.829834::numeric, 37.633096::numeric, 'Крупный выставочный и прогулочный комплекс в Москве.'),
        ('Navka Arena', 'Москва, Navka arena', 55.804196::numeric, 37.531361::numeric, 'Современная площадка для шоу и спортивно-развлекательных событий.'),
        ('Live Арена', 'Москва, Live Арена', 55.728246::numeric, 37.378809::numeric, 'Большая концертная площадка для массовых мероприятий.'),
        ('Большой театр', 'Москва, Большой театр', 55.760186::numeric, 37.618711::numeric, 'Историческая театральная сцена в центре Москвы.'),
        ('Третьяковская галерея', 'Москва, Третьяковская галерея', 55.741389::numeric, 37.620556::numeric, 'Одна из ключевых музейных площадок столицы.'),
        ('Атмосфера', 'Москва, Атмосфера', 55.751244::numeric, 37.618423::numeric, 'Концертная площадка с вечерними музыкальными программами.'),
        ('Лужники', 'Москва, Лужники', 55.715765::numeric, 37.553322::numeric, 'Крупный спортивно-концертный кластер.'),
        ('Квест на Пруд-Ключики', 'Москва, улица Пруд-Ключики, 5', 55.747161::numeric, 37.736518::numeric, 'Локация для иммерсивных и командных квестов.')
) AS v(name, address_line, latitude, longitude, description)
WHERE c.country_name = 'Россия'
  AND c.name = 'Москва'
  AND NOT EXISTS (
      SELECT 1
      FROM place p
      WHERE p.name = v.name AND p.address_line = v.address_line
  );

INSERT INTO event (author_user_id, title, location_description, full_description, age_limit, source_url)
SELECT
    u.id,
    v.title,
    v.location_description,
    v.full_description,
    v.age_limit,
    v.source_url
FROM user_account u
CROSS JOIN (
    VALUES
        ('Futurione', 'ВДНХ, Москва', 'Иммерсивная выставка Futurione на территории ВДНХ с мультимедийными инсталляциями и интерактивными зонами. Подходит для посещения с друзьями и семьей.', 0, 'https://cityhawk.local/events/futurione'),
        ('Ледовое шоу Татьяны Навки', '10 февраля - 21 марта, Navka arena, Москва', 'Большое ледовое шоу с постановочными номерами и театральной драматургией. Формат подходит для вечернего досуга и семейного похода.', 6, 'https://cityhawk.local/events/navka-show'),
        ('Женский стендап', '27 марта, Live Арена, Москва', 'Концертный стендап-формат с выступлениями резидентов и актуальными монологами. Подойдет для компании друзей и насыщенного вечернего досуга.', 18, 'https://cityhawk.local/events/womens-standup'),
        ('Балет Щелкунчик', '14 марта, Большой театр, Москва', 'Классическая постановка балета на исторической сцене Большого театра. Подходит для романтического вечера и культурного маршрута.', 6, 'https://cityhawk.local/events/nutcracker'),
        ('Искусство XX века', '10–24 марта, Третьяковская галерея, Москва', 'Выставочный проект об искусстве XX века с работами ключевых авторов и тематическими залами. Хороший выбор для вдумчивого культурного посещения.', 0, 'https://cityhawk.local/events/art-xx'),
        ('Рок-концерт', '20 марта, Атмосфера, Москва', 'Большой рок-концерт с живым звуком и вечерней программой. Подходит для любителей концертного формата и активного отдыха.', 12, 'https://cityhawk.local/events/rock-concert'),
        ('Руки Вверх', '17 марта, Лужники, Москва', 'Концерт группы Руки Вверх на большой площадке с ретро-хитами и масштабным шоу. Формат рассчитан на вечерний досуг и большую компанию.', 12, 'https://cityhawk.local/events/ruki-vverh'),
        ('Квест Искупление', 'улица Пруд-Ключики, 5, Москва', 'Атмосферный квест с сюжетными загадками и командным прохождением. Подходит для небольших групп и вечернего досуга.', 16, 'https://cityhawk.local/events/quest-iskuplenie')
) AS v(title, location_description, full_description, age_limit, source_url)
WHERE u.email = 'seed.author@cityhawk.local'
  AND NOT EXISTS (
      SELECT 1
      FROM event e
      WHERE e.title = v.title AND e.source_url = v.source_url
  );

INSERT INTO event_session (event_id, place_id, start_at, end_at, price)
SELECT
    e.id,
    p.id,
    v.start_at,
    v.end_at,
    v.price
FROM (
    VALUES
        ('Futurione', 'ВДНХ', '2026-04-02 10:00:00+03'::timestamptz, '2026-04-02 22:00:00+03'::timestamptz, 1200),
        ('Ледовое шоу Татьяны Навки', 'Navka Arena', '2026-04-03 19:00:00+03'::timestamptz, '2026-04-03 21:30:00+03'::timestamptz, 2500),
        ('Женский стендап', 'Live Арена', '2026-04-04 20:00:00+03'::timestamptz, '2026-04-04 22:00:00+03'::timestamptz, 1800),
        ('Балет Щелкунчик', 'Большой театр', '2026-04-05 18:00:00+03'::timestamptz, '2026-04-05 20:30:00+03'::timestamptz, 3000),
        ('Искусство XX века', 'Третьяковская галерея', '2026-04-06 11:00:00+03'::timestamptz, '2026-04-06 20:00:00+03'::timestamptz, 900),
        ('Рок-концерт', 'Атмосфера', '2026-04-07 19:30:00+03'::timestamptz, '2026-04-07 22:30:00+03'::timestamptz, 2200),
        ('Руки Вверх', 'Лужники', '2026-04-08 19:00:00+03'::timestamptz, '2026-04-08 22:00:00+03'::timestamptz, 3500),
        ('Квест Искупление', 'Квест на Пруд-Ключики', '2026-04-09 18:30:00+03'::timestamptz, '2026-04-09 20:00:00+03'::timestamptz, 1500)
) AS v(event_title, place_name, start_at, end_at, price)
JOIN event e ON e.title = v.event_title
JOIN place p ON p.name = v.place_name
WHERE NOT EXISTS (
    SELECT 1
    FROM event_session es
    WHERE es.event_id = e.id
      AND es.place_id = p.id
      AND es.start_at = v.start_at
      AND es.end_at = v.end_at
);

INSERT INTO event_image (event_id, image_url)
SELECT
    e.id,
    v.image_url
FROM (
    VALUES
        ('Futurione', 'https://images.unsplash.com/photo-1511578314322-379afb476865?auto=format&fit=crop&w=1200&q=80'),
        ('Ледовое шоу Татьяны Навки', 'https://images.unsplash.com/photo-1540039155733-5bb30b53aa14?auto=format&fit=crop&w=1200&q=80'),
        ('Женский стендап', 'https://images.unsplash.com/photo-1527224538127-2104bb71c51b?auto=format&fit=crop&w=1200&q=80'),
        ('Балет Щелкунчик', 'https://images.unsplash.com/photo-1503095396549-807759245b35?auto=format&fit=crop&w=1200&q=80'),
        ('Искусство XX века', 'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=1200&q=80'),
        ('Рок-концерт', 'https://images.unsplash.com/photo-1501386761578-eac5c94b800a?auto=format&fit=crop&w=1200&q=80'),
        ('Руки Вверх', 'https://images.unsplash.com/photo-1493225457124-a3eb161ffa5f?auto=format&fit=crop&w=1200&q=80'),
        ('Квест Искупление', 'https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=crop&w=1200&q=80')
) AS v(event_title, image_url)
JOIN event e ON e.title = v.event_title
WHERE NOT EXISTS (
    SELECT 1
    FROM event_image ei
    WHERE ei.event_id = e.id AND ei.image_url = v.image_url
);

INSERT INTO event_category (event_id, category_id)
SELECT
    e.id,
    c.id
FROM (
    VALUES
        ('Futurione', 'Парк'),
        ('Futurione', 'Выставка'),
        ('Futurione', 'Семья'),
        ('Ледовое шоу Татьяны Навки', 'Шоу'),
        ('Ледовое шоу Татьяны Навки', 'Семья'),
        ('Женский стендап', 'Шоу'),
        ('Женский стендап', 'Стендап'),
        ('Балет Щелкунчик', 'Театр'),
        ('Искусство XX века', 'Музей'),
        ('Искусство XX века', 'Фото'),
        ('Рок-концерт', 'Музыка'),
        ('Руки Вверх', 'Музыка'),
        ('Квест Искупление', 'Квест')
) AS v(event_title, category_name)
JOIN event e ON e.title = v.event_title
JOIN category c ON c.name = v.category_name
ON CONFLICT (event_id, category_id) DO NOTHING;

INSERT INTO event_tag (event_id, tag_id)
SELECT
    e.id,
    t.id
FROM (
    VALUES
        ('Futurione', 'Иммерсивное'),
        ('Futurione', 'Семейное'),
        ('Ледовое шоу Татьяны Навки', 'Ледовое шоу'),
        ('Ледовое шоу Татьяны Навки', 'Семейное'),
        ('Женский стендап', 'Стендап'),
        ('Балет Щелкунчик', 'Балет'),
        ('Искусство XX века', 'Искусство'),
        ('Рок-концерт', 'Рок'),
        ('Руки Вверх', 'Ретро'),
        ('Квест Искупление', 'Квест')
) AS v(event_title, tag_name)
JOIN event e ON e.title = v.event_title
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;

INSERT INTO collection (author_user_id, title, description, is_public)
SELECT
    u.id,
    'Выбор на выходные',
    'Подборка лучших событий на выходные',
    TRUE
FROM user_account u
WHERE u.email = 'seed.author@cityhawk.local'
  AND NOT EXISTS (
      SELECT 1
      FROM collection c
      WHERE c.title = 'Выбор на выходные' AND c.author_user_id = u.id
  );

INSERT INTO collection_image (collection_id, image_url)
SELECT
    c.id,
    'https://example.com/collection.jpg'
FROM collection c
WHERE c.title = 'Выбор на выходные'
  AND NOT EXISTS (
      SELECT 1
      FROM collection_image ci
      WHERE ci.collection_id = c.id
        AND ci.image_url = 'https://example.com/collection.jpg'
  );

INSERT INTO collection_event (collection_id, event_id)
SELECT
    c.id,
    e.id
FROM (
    VALUES
        ('Выбор на выходные', 'Futurione'),
        ('Выбор на выходные', 'Рок-концерт'),
        ('Выбор на выходные', 'Ледовое шоу Татьяны Навки')
) AS v(collection_title, event_title)
JOIN collection c ON c.title = v.collection_title
JOIN event e ON e.title = v.event_title
ON CONFLICT (collection_id, event_id) DO NOTHING;

INSERT INTO favorite_event (user_id, event_id)
SELECT
    u.id,
    e.id
FROM (
    VALUES
        ('seed.author@cityhawk.local', 'Futurione'),
        ('seed.author@cityhawk.local', 'Ледовое шоу Татьяны Навки'),
        ('seed.author@cityhawk.local', 'Женский стендап')
) AS v(user_email, event_title)
JOIN user_account u ON u.email = v.user_email
JOIN event e ON e.title = v.event_title
ON CONFLICT (user_id, event_id) DO NOTHING;

COMMIT;
