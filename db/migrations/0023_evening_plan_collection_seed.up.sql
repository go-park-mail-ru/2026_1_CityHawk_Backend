-- CityHawk seed: коллекция «План на вечер»
-- Готовый микс из стендапа, концертов, театра и прогулочных мест.

BEGIN;

-- 0. Безопасные доработки под seed
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

-- 2. Коллекция «План на вечер»
INSERT INTO collection (author_user_id, city_id, title, slug, description, is_public)
SELECT
    u.id,
    c.id,
    'План на вечер',
    'evening-plan-moscow',
    'Подборка для вечера в Москве: стендап, концерты, театр и прогулочные места, куда удобно зайти до или после события.',
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

-- 3. Обложка коллекции
DELETE FROM collection_image ci
USING collection col
WHERE ci.collection_id = col.id
  AND col.slug = 'evening-plan-moscow';

INSERT INTO collection_image (collection_id, image_url)
SELECT
    col.id,
    '/uploads/events/vinzavod_msk86.jpeg'
FROM collection col
WHERE col.slug = 'evening-plan-moscow'
ON CONFLICT (collection_id, image_url) DO NOTHING;

-- 4. Состав коллекции
DELETE FROM collection_event ce
USING collection col
WHERE ce.collection_id = col.id
  AND col.slug = 'evening-plan-moscow';

INSERT INTO collection_event (collection_id, event_id)
SELECT
    col.id,
    e.id
FROM collection col
JOIN event e ON e.slug IN (
    -- Стендап
    'zhenskiy-standup-vdnh-2026-05-29',
    'sergey-orlov-zvezda-2026-05-31',
    'besplatnyy-stand-up-ot-komikov-s-tnt-2026-06',

    -- Концерты
    'toxi-tour',
    'bushido-zho-2026-06-27',
    'saluki-2026-06-28',
    'jason-derulo-2026',

    -- Театры
    'intuicziya',
    'romeo-i-dzhuletta-lyubov-vne-vremeni',
    'prizrak-myuzikla',
    'moskva-petushki',

    -- Прогулочные места
    'gorky-park',
    'kitay-gorod',
    'vinzavod',
    'savvinskoe-podvorie'
)
WHERE col.slug = 'evening-plan-moscow'
ON CONFLICT (collection_id, event_id) DO NOTHING;

COMMIT;
