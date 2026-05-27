ALTER TABLE tag
    ADD COLUMN IF NOT EXISTS tag_group text;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'tag_group_valid'
    ) THEN
        ALTER TABLE tag
            ADD CONSTRAINT tag_group_valid
            CHECK (
                tag_group IS NULL
                OR tag_group IN ('format', 'genre', 'mood', 'audience', 'price')
            );
    END IF;
END $$;

UPDATE tag
SET tag_group = v.tag_group,
    updated_at = now()
FROM (
    VALUES
        ('Кино', 'format'),
        ('Концерт', 'format'),
        ('Театр', 'format'),
        ('Выставка', 'format'),
        ('Экскурсия', 'format'),
        ('Стендап', 'format'),
        ('Мастер-класс', 'format'),
        ('Танцы', 'format'),
        ('Боулинг', 'format'),
        ('Прогулка', 'format'),
        ('Фотоместо', 'format'),
        ('Галерея', 'format'),
        ('Арт-пространство', 'format'),
        ('Клуб', 'format'),
        ('Вечеринка', 'format'),
        ('Спорт', 'format'),
        ('Парк', 'format'),
        ('Драма', 'genre'),
        ('Комедия', 'genre'),
        ('Детектив', 'genre'),
        ('Классика', 'genre'),
        ('Балет', 'genre'),
        ('История', 'genre'),
        ('Архитектура', 'genre'),
        ('Электронная музыка', 'genre'),
        ('Музыка', 'genre'),
        ('Творчество', 'genre'),
        ('Керамика', 'genre'),
        ('Свидание', 'mood'),
        ('Спокойный вечер', 'mood'),
        ('Ночная жизнь', 'mood'),
        ('Для фото', 'mood'),
        ('Иммерсивный', 'mood'),
        ('Движение', 'mood'),
        ('Для детей', 'audience'),
        ('Семейный', 'audience'),
        ('Для начинающих', 'audience'),
        ('Для двоих', 'audience'),
        ('Для компании', 'audience'),
        ('Бесплатно', 'price')
) AS v(name, tag_group)
WHERE tag.name = v.name;

UPDATE tag
SET tag_group = NULL,
    updated_at = now()
WHERE name = 'Премьера';
