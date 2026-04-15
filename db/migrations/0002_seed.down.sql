BEGIN;

DELETE FROM favorite_event fe
USING event e
WHERE fe.event_id = e.id
  AND e.source_url LIKE 'https://seed.cityhawk.local/events/%';

DELETE FROM collection_event ce
USING collection c
WHERE ce.collection_id = c.id
  AND c.title LIKE 'Seed Collection%';

DELETE FROM collection_image ci
USING collection c
WHERE ci.collection_id = c.id
  AND c.title LIKE 'Seed Collection%';

DELETE FROM collection
WHERE title LIKE 'Seed Collection%';

DELETE FROM event_category ec
USING event e
WHERE ec.event_id = e.id
  AND e.source_url LIKE 'https://seed.cityhawk.local/events/%';

DELETE FROM event_tag et
USING event e
WHERE et.event_id = e.id
  AND e.source_url LIKE 'https://seed.cityhawk.local/events/%';

DELETE FROM event_image ei
USING event e
WHERE ei.event_id = e.id
  AND e.source_url LIKE 'https://seed.cityhawk.local/events/%';

DELETE FROM event_session es
USING event e
WHERE es.event_id = e.id
  AND e.source_url LIKE 'https://seed.cityhawk.local/events/%';

DELETE FROM event
WHERE source_url LIKE 'https://seed.cityhawk.local/events/%';

DELETE FROM place
WHERE name IN (
    'ВДНХ',
    'Navka Arena',
    'Live Арена',
    'Большой театр',
    'Третьяковская галерея',
    'Атмосфера',
    'Лужники',
    'Квест на Пруд-Ключики',
    'Дизайн завод',
    'Сад Эрмитаж'
);

DELETE FROM tag
WHERE name IN (
    'Иммерсивное',
    'Семейное',
    'Ледовое шоу',
    'Стендап',
    'Балет',
    'Искусство',
    'Рок',
    'Ретро',
    'Квест',
    'На воздухе',
    'Ночное',
    'Образовательное'
);

DELETE FROM category
WHERE name IN (
    'Парк',
    'Выставка',
    'Семья',
    'Шоу',
    'Стендап',
    'Театр',
    'Музей',
    'Музыка',
    'Квест',
    'Фото',
    'Лекция',
    'Фестиваль'
);

DELETE FROM user_account
WHERE email = 'seed.author@cityhawk.local';

DELETE FROM city
WHERE country_name = 'Россия'
  AND name = 'Москва';

COMMIT;
