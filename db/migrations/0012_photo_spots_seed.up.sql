-- CityHawk seed: коллекция «Места для фото»
-- Источник локаций и фотографий: https://bagmutskiy.com/blog/100-luchshih-mest-dlya-fotosessii-v-moskve/
-- Важно: описания ниже написаны своими словами для карточек CityHawk, а не скопированы из статьи.
-- Фото оставлены внешними ссылками с сайта-источника/CDN; для продакшена лучше заменить на свои/разрешенные изображения.

BEGIN;

-- 0. Небольшие доработки под MVP, если их еще нет
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
    ('Парк'),
    ('У воды'),
    ('Арт-пространство'),
    ('Историческое место'),
    ('Необычное место'),
    ('Урбанистика'),
    ('На свежем воздухе'),
    ('Индустриальный стиль'),
    ('Для свидания'),
    ('Зелень'),
    ('Красивый вид')
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
            'Завод «Кристалл»',
            'Москва, Самокатная ул., 4, стр. 1',
            55.754600::numeric,
            37.677900::numeric,
            'Бывшая промышленная территория с краснокирпичными корпусами, трубами, внутренними проездами и атмосферой старой московской промзоны. Локация хорошо подходит для индустриальных и урбанистичных кадров: здесь можно снимать на фоне кирпичных стен, фактурных деталей, дворов и необычных архитектурных элементов.'
        ),
        (
            '«Пресня Сити» и «Зотов»',
            'Москва, Ходынская ул., 2, стр. 1',
            55.768600::numeric,
            37.562900::numeric,
            'Современный городской квартал рядом с культурным пространством «Зотов». В одной точке сочетаются высотные башни, конструктивистская эстетика бывшего хлебозавода, геометрия фасадов и выразительные линии. Подходит для стильных городских кадров, архитектурных деталей и съемки в духе новой Москвы.'
        ),
        (
            'Центр дизайна ARTPLAY',
            'Москва, Нижняя Сыромятническая ул., 10, стр. 2',
            55.754100::numeric,
            37.668600::numeric,
            'Креативный кластер рядом с Курским вокзалом. На территории есть яркие фасады, граффити, лестницы, внутренние дворы, дизайнерские пространства и индустриальные элементы. Локация подходит для прогулочной съемки, урбанистичных портретов и кадров с творческой атмосферой.'
        ),
        (
            'ГЭС-2',
            'Москва, Болотная наб., 15',
            55.741800::numeric,
            37.611800::numeric,
            'Культурное пространство на Болотной набережной с белой архитектурой, большими окнами, синими трубами и видом на воду. Локация сочетает историческое здание и современную реконструкцию, поэтому хорошо работает для минималистичных, светлых и экспериментальных кадров.'
        ),
        (
            'Центр «Графит»',
            'Москва, Электродная ул., 2, стр. 32',
            55.750900::numeric,
            37.742500::numeric,
            'Современное культурное пространство на территории бывшего промышленного района. Здесь много муралов, граффити, ярких стен, арт-объектов и индустриальных фактур. Подходит для творческих, смелых и нестандартных съемок, где важны цвет, характер и выразительный фон.'
        ),
        (
            'Оранжерея Главного ботанического сада',
            'Москва, Ботаническая ул., 4, стр. 2',
            55.844900::numeric,
            37.604300::numeric,
            'Оранжерея с тропическими растениями, пальмами, цветами, кактусами и разными микроклиматами. Это удобная локация для съемки в холодное время года, когда хочется зеленого фона и природной атмосферы. Подходит для романтических, семейных и портретных фотографий.'
        ),
        (
            'Улица Школьная',
            'Москва, ул. Школьная',
            55.744900::numeric,
            37.679300::numeric,
            'Атмосферная улица в Таганском районе рядом с метро «Римская» и «Площадь Ильича». Невысокие цветные дома, мощеные участки, старые фонари и спокойная городская среда создают ощущение старой Москвы. Подходит для ретро-кадров, прогулок и романтичных фотосессий.'
        ),
        (
            'Аптекарский огород',
            'Москва, просп. Мира, 26, стр. 1',
            55.781400::numeric,
            37.636700::numeric,
            'Старейший ботанический сад с оранжереями, цветниками, прудами и сезонными растениями. Весной и летом здесь много ярких природных фонов, а зимой можно снимать в теплицах и среди тропической зелени. Хороший вариант для портретов, семейных прогулок и романтичных кадров.'
        ),
        (
            'Депо Лесная',
            'Москва, Лесная ул., 20',
            55.779800::numeric,
            37.588700::numeric,
            'Фудмолл и городское пространство на территории бывшего троллейбусного парка. Здесь сочетаются кирпичная промышленная архитектура, современная гастрономическая среда, световые вывески и фактурные дворы. Локация подходит для урбанистичных и lifestyle-кадров.'
        ),
        (
            'Парк Горького',
            'Москва, ул. Крымский Вал, 9',
            55.729800::numeric,
            37.601100::numeric,
            'Большой городской парк в центре Москвы с аллеями, набережной, фонтанами, павильонами и сезонными активностями. Локация универсальна: здесь можно снять спокойную прогулку, динамичные кадры, семейную историю или романтичную прогулку у воды.'
        ),
        (
            'Парк Сокольники',
            'Москва, ул. Сокольнический Вал, 1, стр. 1',
            55.794300::numeric,
            37.676200::numeric,
            'Один из крупных московских парков с аллеями, зеленью, розарием, аттракционами и сезонными площадками. Весной и летом здесь много природных фонов, осенью — яркая листва, зимой — снежные и динамичные кадры. Подходит для семейных, парных и прогулочных съемок.'
        ),
        (
            'Бизнес-центр «Аквамарин»',
            'Москва, Озерковская наб., 24, стр. 2',
            55.736600::numeric,
            37.637500::numeric,
            'Современный деловой комплекс с белыми фасадами, панорамными окнами, изогнутыми линиями и минималистичной архитектурой. Локация хорошо подходит для деловых портретов, урбанистичных кадров, съемки с отражениями и фотографий в стиле современной Москвы.'
        ),
        (
            'Нескучный сад',
            'Москва, Нескучный сад',
            55.717800::numeric,
            37.590400::numeric,
            'Один из старейших московских парков, соединенный с Парком Горького, но более спокойный и камерный. Здесь есть лесные участки, дорожки, старинные элементы, мостики, водоемы и ощущение уединения. Подходит для романтичных, семейных и прогулочных фотографий на природе.'
        ),
        (
            'Китай-город',
            'Москва, метро Китай-город',
            55.753600::numeric,
            37.633600::numeric,
            'Исторический район в центре Москвы с узкими улицами, старинными зданиями, тихими переулками, кафе и неоклассической архитектурой. Локация подходит для прогулочных маршрутов, городских портретов и кадров с атмосферой старой Москвы.'
        ),
        (
            'Новодевичий монастырь',
            'Москва, Новодевичий пр., 1',
            55.726500::numeric,
            37.555200::numeric,
            'Исторический монастырский комплекс рядом с прудом и зелеными прогулочными зонами. Старинные стены, башни, купола, отражения в воде и спокойная атмосфера делают место подходящим для романтичных, свадебных, прогулочных и архитектурных кадров.'
        )
) AS v(name, address_line, latitude, longitude, description)
WHERE c.country_name = 'Россия'
  AND c.name = 'Москва'
ON CONFLICT (city_id, name, address_line) DO UPDATE
SET latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    description = EXCLUDED.description,
    updated_at = now();


-- 4. События-точки для карты
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
    'https://bagmutskiy.com/blog/100-luchshih-mest-dlya-fotosessii-v-moskve/'
FROM user_account u
CROSS JOIN (
    VALUES
        (
            'Завод «Кристалл»',
            'zavod-kristall',
            'Самокатная ул., 4, стр. 1',
            'Бывший завод «Кристалл» — атмосферная промышленная территория с красным кирпичом, трубами, внутренними дворами и ощущением старой московской промзоны. Здесь хорошо получаются индустриальные портреты, урбанистичные кадры и прогулочные фотографии с выразительной фактурой. Локация интересна тем, что выглядит не как стандартная туристическая точка: в кадре много деталей, глубины, линий и архитектурного характера.'
        ),
        (
            '«Пресня Сити» и «Зотов»',
            'presnya-city-zotov',
            'Ходынская ул., 2, стр. 1',
            '«Пресня Сити» и культурное пространство «Зотов» — точка для съемки в стиле конструктивизма и современной городской архитектуры. Рядом находятся высотные башни, отреставрированное здание бывшего хлебозавода, большие плоскости фасадов и четкая геометрия. Это место подходит для стильных портретов, кадров с архитектурными акцентами и маршрута по новому городскому району.'
        ),
        (
            'Центр дизайна ARTPLAY',
            'artplay',
            'Нижняя Сыромятническая ул., 10, стр. 2',
            'ARTPLAY — креативный кластер с граффити, цветными фасадами, лестницами, дворами, кафе и дизайн-пространствами. Здесь можно снимать как яркие lifestyle-кадры, так и более жесткие урбанистичные портреты. Локация хорошо подходит для тех, кому нужны фоны с характером: бетон, кирпич, металл, вывески, переходы и визуальная плотность творческого пространства.'
        ),
        (
            'ГЭС-2',
            'ges-2',
            'Болотная наб., 15',
            'ГЭС-2 — современное культурное пространство у воды, где историческая промышленная архитектура переосмыслена в светлый минималистичный комплекс. Белые фасады, синие трубы, большие окна, открытые пространства и набережная создают спокойный, чистый и стильный фон. Локация подойдет для городских прогулок, экспериментальных кадров и съемки в современной московской атмосфере.'
        ),
        (
            'Центр «Графит»',
            'grafit',
            'Электродная ул., 2, стр. 32',
            'Центр «Графит» — арт-пространство с муралами, граффити, индустриальными фактурами и яркими визуальными акцентами. Это место подходит для смелых съемок, уличного стиля, творческих портретов и кадров, где фон должен быть активным и запоминающимся. Локация особенно хорошо работает, если хочется не классической красоты, а энергии города и современного искусства.'
        ),
        (
            'Оранжерея Главного ботанического сада',
            'main-botanical-garden-greenhouse',
            'Ботаническая ул., 4, стр. 2',
            'Оранжерея Главного ботанического сада — зеленая локация для съемок в любое время года. Внутри можно найти пальмы, экзотические растения, цветы, стеклянные конструкции и мягкий рассеянный свет. Это хороший выбор для романтических, семейных и портретных фотографий, особенно зимой или в холодную погоду, когда хочется получить живой природный фон.'
        ),
        (
            'Улица Школьная',
            'shkolnaya-street',
            'ул. Школьная, район метро «Римская» и «Площадь Ильича»',
            'Улица Школьная — спокойная и атмосферная городская локация с невысокими цветными домами, мощеными участками и ощущением старой Москвы. Здесь можно сделать романтичные, семейные и ретро-кадры без слишком шумного туристического фона. Место хорошо подходит для прогулочного маршрута, где важны цвет, архитектура и камерная атмосфера.'
        ),
        (
            'Аптекарский огород',
            'aptekarsky-ogorod',
            'просп. Мира, 26, стр. 1',
            'Аптекарский огород — ботанический сад в центре Москвы с оранжереями, цветниками, прудами и сезонными растениями. Весной и летом здесь много ярких цветочных фонов, а зимой можно снимать среди тропической зелени. Локация подходит для портретов, семейных съемок, романтичных прогулок и кадров, где нужен природный фон без выезда за город.'
        ),
        (
            'Депо Лесная',
            'depo-lesnaya',
            'Лесная ул., 20',
            'Депо Лесная — городское пространство на территории бывшего троллейбусного парка. Кирпичные стены, промышленные детали, современные рестораны, световые вывески и внутренние дворы создают интересный контраст старого и нового. Это место удобно для lifestyle-съемки, прогулочных кадров, стрит-фото и короткой фотопрогулки рядом с кафе.'
        ),
        (
            'Парк Горького',
            'gorky-park',
            'ул. Крымский Вал, 9',
            'Парк Горького — универсальная локация для прогулочных, романтичных, семейных и динамичных кадров. Здесь есть аллеи, набережная, фонтаны, павильоны, сезонные активности и много открытого пространства. Парк хорошо подходит для разных сценариев: от спокойной прогулки у воды до более живых кадров с движением, велосипедами, катком или городской активностью.'
        ),
        (
            'Парк Сокольники',
            'sokolniki-park',
            'ул. Сокольнический Вал, 1, стр. 1',
            'Парк Сокольники — большое зеленое пространство с аллеями, розарием, аттракционами и сезонными площадками. Здесь можно снимать прогулки, семейные истории, парные кадры и динамичные фотографии на свежем воздухе. Весной и летом парк дает много зелени, осенью — теплые цвета, зимой — снежные маршруты и атмосферу городского отдыха.'
        ),
        (
            'Бизнес-центр «Аквамарин»',
            'aquamarine-business-center',
            'Озерковская наб., 24, стр. 2',
            'Бизнес-центр «Аквамарин» — современная локация с белыми фасадами, стеклом, строгими линиями и выразительной перспективой. Место подходит для деловых портретов, минималистичных урбанистичных кадров и съемки в стиле современной архитектуры. Особенно интересно работать с отражениями, симметрией и мягким светом между зданиями.'
        ),
        (
            'Нескучный сад',
            'neskuchny-garden',
            'Москва, Нескучный сад',
            'Нескучный сад — спокойная природная локация рядом с Парком Горького, но с более уединенной атмосферой. Здесь есть старинные элементы, мостики, дорожки, густая зелень, водоемы и ощущение старого парка. Место хорошо подходит для романтичных, свадебных, семейных и прогулочных фотографий, особенно если нужен тихий фон в центре города.'
        ),
        (
            'Китай-город',
            'kitay-gorod',
            'метро Китай-город',
            'Китай-город — исторический район с узкими улицами, старинными фасадами, тихими переулками, кафе и неоклассическими зданиями. Это хорошая локация для прогулочного маршрута по центру, городских портретов и кадров с атмосферой старой Москвы. В будни здесь можно найти более спокойные улицы и интересные архитектурные детали.'
        ),
        (
            'Новодевичий монастырь',
            'novodevichy-monastery',
            'Новодевичий пр., 1',
            'Новодевичий монастырь — исторический комплекс рядом с прудом и зелеными прогулочными зонами. В кадре хорошо работают стены, башни, купола, отражения в воде и спокойная атмосфера места. Локация подходит для романтичных прогулок, свадебных кадров, архитектурных фотографий и съемки с историческим московским настроением.'
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


-- 5. Привязка событий к местам на карте без расписания и дат
INSERT INTO event_place (event_id, place_id)
SELECT
    e.id,
    p.id
FROM (
    VALUES
        ('zavod-kristall', 'Завод «Кристалл»'),
        ('presnya-city-zotov', '«Пресня Сити» и «Зотов»'),
        ('artplay', 'Центр дизайна ARTPLAY'),
        ('ges-2', 'ГЭС-2'),
        ('grafit', 'Центр «Графит»'),
        ('main-botanical-garden-greenhouse', 'Оранжерея Главного ботанического сада'),
        ('shkolnaya-street', 'Улица Школьная'),
        ('aptekarsky-ogorod', 'Аптекарский огород'),
        ('depo-lesnaya', 'Депо Лесная'),
        ('gorky-park', 'Парк Горького'),
        ('sokolniki-park', 'Парк Сокольники'),
        ('aquamarine-business-center', 'Бизнес-центр «Аквамарин»'),
        ('neskuchny-garden', 'Нескучный сад'),
        ('kitay-gorod', 'Китай-город'),
        ('novodevichy-monastery', 'Новодевичий монастырь')
) AS v(event_slug, place_name)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = v.place_name
ON CONFLICT (event_id) DO UPDATE
SET place_id = EXCLUDED.place_id,
    updated_at = now();


-- 6. Фото событий: максимум 4 на событие, если в статье меньше — оставлено меньше
INSERT INTO event_image (event_id, image_url)
SELECT
    e.id,
    v.image_url
FROM (
    VALUES
        -- Завод «Кристалл»: 4 фото
        ('zavod-kristall', '/uploads/events/zavod-kristall-01.jpg'),
        ('zavod-kristall', '/uploads/events/zavod-kristall-02.jpg'),
        ('zavod-kristall', '/uploads/events/zavod-kristall-03.jpg'),
        ('zavod-kristall', '/uploads/events/zavod-kristall-04.jpg'),

        -- «Пресня Сити» и «Зотов»: 4 фото
        ('presnya-city-zotov', '/uploads/events/presnya-city-zotov-01.jpg'),
        ('presnya-city-zotov', '/uploads/events/presnya-city-zotov-02.jpg'),
        ('presnya-city-zotov', '/uploads/events/presnya-city-zotov-03.jpg'),
        ('presnya-city-zotov', '/uploads/events/presnya-city-zotov-04.jpg'),

        -- Центр дизайна ARTPLAY: 4 фото
        ('artplay', '/uploads/events/artplay-01.jpg'),
        ('artplay', '/uploads/events/artplay-02.jpg'),
        ('artplay', '/uploads/events/artplay-03.jpg'),
        ('artplay', '/uploads/events/artplay-04.jpg'),

        -- ГЭС-2: 4 фото
        ('ges-2', '/uploads/events/ges-2-01.jpg'),
        ('ges-2', '/uploads/events/ges-2-02.jpg'),
        ('ges-2', '/uploads/events/ges-2-03.jpg'),
        ('ges-2', '/uploads/events/ges-2-04.jpg'),

        -- Центр «Графит»: 4 фото
        ('grafit', '/uploads/events/grafit-01.jpg'),
        ('grafit', '/uploads/events/grafit-02.jpg'),
        ('grafit', '/uploads/events/grafit-03.jpg'),
        ('grafit', '/uploads/events/grafit-04.jpg'),

        -- Оранжерея Главного ботанического сада: 4 фото
        ('main-botanical-garden-greenhouse', '/uploads/events/main-botanical-garden-greenhouse-01.jpg'),
        ('main-botanical-garden-greenhouse', '/uploads/events/main-botanical-garden-greenhouse-02.jpg'),
        ('main-botanical-garden-greenhouse', '/uploads/events/main-botanical-garden-greenhouse-03.jpg'),
        ('main-botanical-garden-greenhouse', '/uploads/events/main-botanical-garden-greenhouse-04.jpg'),

        -- Улица Школьная: 4 фото
        ('shkolnaya-street', '/uploads/events/shkolnaya-street-01.jpg'),
        ('shkolnaya-street', '/uploads/events/shkolnaya-street-02.jpg'),
        ('shkolnaya-street', '/uploads/events/shkolnaya-street-03.jpg'),
        ('shkolnaya-street', '/uploads/events/shkolnaya-street-04.jpg'),

        -- Аптекарский огород: 3 фото
        ('aptekarsky-ogorod', '/uploads/events/aptekarsky-ogorod-01.jpg'),
        ('aptekarsky-ogorod', '/uploads/events/aptekarsky-ogorod-02.jpg'),
        ('aptekarsky-ogorod', '/uploads/events/aptekarsky-ogorod-03.jpg'),

        -- Депо Лесная: 3 фото
        ('depo-lesnaya', '/uploads/events/depo-lesnaya-01.jpg'),
        ('depo-lesnaya', '/uploads/events/depo-lesnaya-02.jpg'),
        ('depo-lesnaya', '/uploads/events/depo-lesnaya-03.jpg'),

        -- Парк Горького: 4 фото
        ('gorky-park', '/uploads/events/gorky-park-01.jpg'),
        ('gorky-park', '/uploads/events/gorky-park-02.jpg'),
        ('gorky-park', '/uploads/events/gorky-park-03.jpg'),
        ('gorky-park', '/uploads/events/gorky-park-04.jpg'),

        -- Парк Сокольники: 3 фото
        ('sokolniki-park', '/uploads/events/sokolniki-park-01.jpg'),
        ('sokolniki-park', '/uploads/events/sokolniki-park-02.jpg'),
        ('sokolniki-park', '/uploads/events/sokolniki-park-03.jpg'),

        -- Бизнес-центр «Аквамарин»: 3 фото
        ('aquamarine-business-center', '/uploads/events/aquamarine-business-center-01.jpg'),
        ('aquamarine-business-center', '/uploads/events/aquamarine-business-center-02.jpg'),
        ('aquamarine-business-center', '/uploads/events/aquamarine-business-center-03.jpg'),

        -- Нескучный сад: 4 фото
        ('neskuchny-garden', '/uploads/events/neskuchny-garden-01.jpg'),
        ('neskuchny-garden', '/uploads/events/neskuchny-garden-02.jpg'),
        ('neskuchny-garden', '/uploads/events/neskuchny-garden-03.jpg'),
        ('neskuchny-garden', '/uploads/events/neskuchny-garden-04.jpg'),

        -- Китай-город: 4 фото
        ('kitay-gorod', '/uploads/events/kitay-gorod-01.jpg'),
        ('kitay-gorod', '/uploads/events/kitay-gorod-02.jpg'),
        ('kitay-gorod', '/uploads/events/kitay-gorod-03.jpg'),
        ('kitay-gorod', '/uploads/events/kitay-gorod-04.jpg'),

        -- Новодевичий монастырь: 3 фото
        ('novodevichy-monastery', '/uploads/events/novodevichy-monastery-01.jpg'),
        ('novodevichy-monastery', '/uploads/events/novodevichy-monastery-02.jpg'),
        ('novodevichy-monastery', '/uploads/events/novodevichy-monastery-03.jpg')
) AS v(event_slug, image_url)
JOIN event e ON e.slug = v.event_slug
ON CONFLICT (event_id, image_url) DO NOTHING;


-- 7. Категория «Фото» для всех 15 событий
INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM event e
JOIN category c ON c.name = 'Фото'
WHERE e.slug IN (
    'zavod-kristall',
    'presnya-city-zotov',
    'artplay',
    'ges-2',
    'grafit',
    'main-botanical-garden-greenhouse',
    'shkolnaya-street',
    'aptekarsky-ogorod',
    'depo-lesnaya',
    'gorky-park',
    'sokolniki-park',
    'aquamarine-business-center',
    'neskuchny-garden',
    'kitay-gorod',
    'novodevichy-monastery'
)
ON CONFLICT (event_id, category_id) DO NOTHING;


-- 8. Теги событий
INSERT INTO event_tag (event_id, tag_id)
SELECT e.id, t.id
FROM (
    VALUES
        ('zavod-kristall', 'Фото'),
        ('zavod-kristall', 'Фотоместо'),
        ('zavod-kristall', 'Архитектура'),
        ('zavod-kristall', 'Индустриальный стиль'),
        ('zavod-kristall', 'Урбанистика'),

        ('presnya-city-zotov', 'Фото'),
        ('presnya-city-zotov', 'Фотоместо'),
        ('presnya-city-zotov', 'Архитектура'),
        ('presnya-city-zotov', 'Урбанистика'),

        ('artplay', 'Фото'),
        ('artplay', 'Фотоместо'),
        ('artplay', 'Арт-пространство'),
        ('artplay', 'Урбанистика'),
        ('artplay', 'Необычное место'),

        ('ges-2', 'Фото'),
        ('ges-2', 'Фотоместо'),
        ('ges-2', 'У воды'),
        ('ges-2', 'Арт-пространство'),
        ('ges-2', 'Архитектура'),

        ('grafit', 'Фото'),
        ('grafit', 'Фотоместо'),
        ('grafit', 'Арт-пространство'),
        ('grafit', 'Необычное место'),
        ('grafit', 'Индустриальный стиль'),

        ('main-botanical-garden-greenhouse', 'Фото'),
        ('main-botanical-garden-greenhouse', 'Фотоместо'),
        ('main-botanical-garden-greenhouse', 'Зелень'),
        ('main-botanical-garden-greenhouse', 'Для свидания'),

        ('shkolnaya-street', 'Фото'),
        ('shkolnaya-street', 'Фотоместо'),
        ('shkolnaya-street', 'Прогулка'),
        ('shkolnaya-street', 'Историческое место'),

        ('aptekarsky-ogorod', 'Фото'),
        ('aptekarsky-ogorod', 'Фотоместо'),
        ('aptekarsky-ogorod', 'Парк'),
        ('aptekarsky-ogorod', 'Зелень'),
        ('aptekarsky-ogorod', 'Для свидания'),

        ('depo-lesnaya', 'Фото'),
        ('depo-lesnaya', 'Фотоместо'),
        ('depo-lesnaya', 'Архитектура'),
        ('depo-lesnaya', 'Урбанистика'),
        ('depo-lesnaya', 'Индустриальный стиль'),

        ('gorky-park', 'Фото'),
        ('gorky-park', 'Фотоместо'),
        ('gorky-park', 'Парк'),
        ('gorky-park', 'У воды'),
        ('gorky-park', 'На свежем воздухе'),

        ('sokolniki-park', 'Фото'),
        ('sokolniki-park', 'Фотоместо'),
        ('sokolniki-park', 'Парк'),
        ('sokolniki-park', 'На свежем воздухе'),

        ('aquamarine-business-center', 'Фото'),
        ('aquamarine-business-center', 'Фотоместо'),
        ('aquamarine-business-center', 'Архитектура'),
        ('aquamarine-business-center', 'Урбанистика'),

        ('neskuchny-garden', 'Фото'),
        ('neskuchny-garden', 'Фотоместо'),
        ('neskuchny-garden', 'Парк'),
        ('neskuchny-garden', 'На свежем воздухе'),
        ('neskuchny-garden', 'Для свидания'),

        ('kitay-gorod', 'Фото'),
        ('kitay-gorod', 'Фотоместо'),
        ('kitay-gorod', 'Прогулка'),
        ('kitay-gorod', 'Историческое место'),
        ('kitay-gorod', 'Архитектура'),

        ('novodevichy-monastery', 'Фото'),
        ('novodevichy-monastery', 'Фотоместо'),
        ('novodevichy-monastery', 'Историческое место'),
        ('novodevichy-monastery', 'У воды'),
        ('novodevichy-monastery', 'Для свидания')
) AS v(event_slug, tag_name)
JOIN event e ON e.slug = v.event_slug
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;


-- 9. Коллекция «Места для фото»
INSERT INTO collection (author_user_id, city_id, title, slug, description, is_public)
SELECT
    u.id,
    c.id,
    'Места для фото',
    'photo-spots-moscow',
    'Фотогеничные места Москвы для прогулок, свиданий, портретов и архитектурных кадров.',
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
INSERT INTO collection_image (collection_id, image_url)
SELECT
    col.id,
    '/uploads/collections/photo-spots-moscow.jpg'
FROM collection col
WHERE col.slug = 'photo-spots-moscow'
ON CONFLICT (collection_id, image_url) DO NOTHING;


-- 11. Связь коллекции с 15 событиями
INSERT INTO collection_event (collection_id, event_id)
SELECT
    col.id,
    e.id
FROM collection col
JOIN event e ON e.slug IN (
    'zavod-kristall',
    'presnya-city-zotov',
    'artplay',
    'ges-2',
    'grafit',
    'main-botanical-garden-greenhouse',
    'shkolnaya-street',
    'aptekarsky-ogorod',
    'depo-lesnaya',
    'gorky-park',
    'sokolniki-park',
    'aquamarine-business-center',
    'neskuchny-garden',
    'kitay-gorod',
    'novodevichy-monastery'
)
WHERE col.slug = 'photo-spots-moscow'
ON CONFLICT (collection_id, event_id) DO NOTHING;

COMMIT;
