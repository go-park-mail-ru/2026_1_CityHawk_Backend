-- CityHawk seed: категория «Кино»
-- Источник данных: afisha.ru
-- Изображения взяты из открытых источников; source_url у каждого события ведет на исходную страницу Афиши.

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
VALUES ('Кино')
ON CONFLICT (name) DO NOTHING;

INSERT INTO tag (name)
VALUES
    ('Кино'), ('Фильм'), ('Премьера'), ('Новинка'),
    ('Фэнтези'), ('Фантастика'), ('Драма'), ('Комедия'), ('Триллер'),
    ('Ужасы'), ('Мелодрама'), ('Боевик'), ('Ромком'), ('Исторический'),
    ('Приключение'), ('Биография'), ('Мультфильм'), ('Документальный'),
    ('Детектив'), ('Трагикомедия'), ('Российское кино'), ('Зарубежное кино'),
    ('Семейный')
ON CONFLICT (name) DO NOTHING;

-- 3. Реальный кинотеатр Москвы
INSERT INTO place (city_id, name, address_line, latitude, longitude, description)
SELECT
    c.id,
    'КАРО 11 Октябрь',
    'Москва, ул. Новый Арбат, 24',
    55.752200::numeric,
    37.586600::numeric,
    'Крупный московский киноцентр на Новом Арбате, премьерная и фестивальная площадка с несколькими залами.'
FROM city c
WHERE c.country_name = 'Россия'
  AND c.name = 'Москва'
ON CONFLICT (city_id, name, address_line) DO UPDATE
SET latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    description = EXCLUDED.description,
    updated_at = now();

-- 4. События (фильмы)
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
        ('Ангелы Ладоги', 'angely-ladogi', 'КАРО 11 Октябрь', 'В ноябре, в самое начало страшной зимы 1941 года, когда вокруг Ленинграда смыкается кольцо блокады, отряд молодых спортсменов-буеристов выходит на тонкий, еще не вставший лед. Миссия по доставке боеприпасов превращается в операцию по спасению сирот из детского дома, которых не успели эвакуировать. Бывшие соперники в спорте по регатам теперь должны превратиться в настоящую единую команду, чтобы выжить самим и подарить надежду всем остальным.', 12, 'https://www.afisha.ru/movie/angely-ladogi-308861/'),
        ('Вот это драма!', 'vot-eto-drama', 'КАРО 11 Октябрь', 'Чарли и Эмма влюблены и вот-вот поженятся. Место свадьбы выбрано, приглашения разосланы, до долгожданной церемонии остаются считанные дни. Но одно внезапное откровение переворачивает всё — Чарли узнаёт об Эмме то, что предпочёл бы не знать. Теперь на пути к «долго и счастливо» их отношениям предстоит пройти через неожиданную драму.', 16, 'https://www.afisha.ru/movie/vot-eto-drama-1002618/'),
        ('Ждун 2', 'zhdun-2', 'КАРО 11 Октябрь', 'ДЖдун возвращается на Землю, чтобы предупредить об угрозе своего друга, мальчика Никиту, отдыхающего в летнем лагере. Злодеи из космоса планируют визит на Землю, и этот визит грозит землянам настоящей катастрофой. Никите предстоит собрать команду и дать отпор непрошенным гостям.', 6, 'https://www.afisha.ru/movie/zhdun-2-1002129/'),
        ('Домовенок Кузя 2', 'domovenok-kuzya-2', 'КАРО 11 Октябрь', 'Кузя продолжает жить с Наташей и её семьёй. Баба Яга окончательно перебралась в мир людей и старается быть полезной, став неожиданной помощницей для окружающих. Всё кажется спокойным, но в их дом приходит загадочная Тихоня, с которой начинается новая волна волшебных и опасных событий. Всем предстоит снова объединиться, чтобы защитить мир от Кощея и сохранить мир сказок таким же волшебным и светлым.', 0, 'https://www.afisha.ru/movie/domovenok-kuzya-2-1000990/'),
        ('Дьявол носит Прада 2', 'dyavol-nosit-prada-2', 'КАРО 11 Октябрь', 'Мечтающая стать журналисткой провинциальная девушка Энди по окончании университета получает должность помощницы всесильной Миранды Пристли, деспотичного редактора одного из крупнейших нью-йоркских журналов мод. Энди всегда мечтала о такой работе, не зная, с каким нервным напряжением это будет связано…', 18, 'https://www.afisha.ru/movie/dyavol-nosit-prada-2-1002692/'),
        ('Коммерсант', 'kommersant', 'КАРО 11 Октябрь', 'Москва, 1996 год. У молодого и предприимчивого банкира Андрея Рубанова есть жена Ирма и маленький ребенок. А ещё — больше миллиона долларов на двоих с другом Мишей. Видя себя будущей элитой нового государства, олигархами, они вписываются в крупную сделку, но оказывается, что друзей втянули в масштабную аферу по хищению миллиардов из казны. После ареста Андрей попадает в переполненную общую камеру «Матросской тишины» — чистилище, где ему предстоит стать другим человеком, чтобы выжить и вернуться в привычный мир.', 18, 'https://www.afisha.ru/movie/kommersant-1000204/'),
        ('Семь верст до рассвета', 'sem-verst-do-rassveta', 'КАРО 11 Октябрь', 'Зима 1942 года, деревня Куракино. В суровых условиях оккупированной Псковской области, Матвей Кузьмин повторяет подвиг Ивана Сусанина и предотвращает диверсию в советском тылу. Получив приказ провести вражеский отряд в тыл прорвавшихся красноармейцев, Кузьмин подчиняется. Всю ночь старик водит немцев сквозь метель по бездорожью, но враги не подозревают, что он задумал.', 16, 'https://www.afisha.ru/movie/sem-verst-do-rassveta-1003075/'),
        ('Проект Конец света', 'proekt-konec-sveta', 'КАРО 11 Октябрь', 'Микробиолог Райленд Грейс просыпается на космическом корабле, не помня ни себя, ни свою миссию. Мужчина выясняет, что он — единственный выживший из экипажа, отправленного к звезде Тау Кита в поисках спасения от катастрофы на Земле. Чтобы найти решение, Райленду предстоит положиться на свои обширные научные знания и изобретательность, но, возможно, ему не придётся искать в одиночку.', 16, 'https://www.afisha.ru/movie/proekt-konec-sveta-1002688/'),
        ('Литвяк', 'litvyak', 'КАРО 11 Октябрь', 'Великая Отечественная война. 21-летняя Лидия Литвяк служит лётчицей-истребителем в авиационном полку. Вместе с боевыми товарищами она выполняет задания на самых напряжённых участках фронта, участвуя в ожесточенных воздушных боях и показывая себя одним из самых результативных асов. В полку у неё завязываются отношения с лётчиком Алексеем Соломатиным. Боевые победы, потери и личные переживания не меняют твёрдый характер Лидии.', 12, 'https://www.afisha.ru/movie/litvyak-308882/'),
        ('Моя собака — космонавт', 'moya-sobaka-kosmonavt', 'КАРО 11 Октябрь', '1960 год. В городке у космодрома Байконур, где каждый запуск ракеты озаряет небо мечтами, живёт десятилетний Миша — мальчик со светлой головой, полной космических фантазий. В свой день рождения Миша находит собаку Белку, которая предопределит не только его судьбу, но и напишет новую страницу в истории освоения космоса.', 6, 'https://www.afisha.ru/movie/moya-sobaka-kosmonavt-308030/'),
        ('Грузовички', 'gruzovichki', 'КАРО 11 Октябрь', 'Грузовичок-подросток Саня, всю жизнь проработавший в карьере с родителями, случайно знакомится с городскими жителями: еще одним грузовичком и поливальной машиной. Новые знакомства открывают перед Саней огромный удивительный мир за пределами карьера, на который неожиданно надвигается серьезная угроза — проснувшийся вулкан в горах. Теперь Сане и его друзьям предстоит придумать, как победить стихию и убедить всех грузовичков объединиться, пока не стало слишком поздно.', 0, 'https://www.afisha.ru/movie/gruzovichki-307471/'),
        ('Полутон', 'poluton', 'КАРО 11 Октябрь', 'Ведущая популярного подкаста о паранормальных явлениях начинает получать загадочные аудиозаписи от неизвестного отправителя', 18, 'https://www.afisha.ru/movie/poluton-1003089/'),
        ('Грация', 'graciya', 'КАРО 11 Октябрь', 'Мариано Де Сантис погружается в воспоминания о главной любви своей жизни. Чувства, которые когда-то определили его судьбу, теперь могут изменить её вновь. Мариано — президент Италии и готовится покинуть пост, но напоследок он должен принять ряд крайне сложных решений. Общество и окружение замерли в напряжённом ожидании.', 16, 'https://www.afisha.ru/movie/graciya-1001140/'),
        ('Отец', 'otec', 'КАРО 11 Октябрь', '1942 год. Сибирский охотник Гавриил Собинов узнаёт, что его сын Семен числится пропавшим без вести. Гавриил уходит добровольцем на фронт в надежде найти сына живым. Оказавшись на передовой, он возглавляет группу снайперов, и теперь ему предстоит научить молодых бойцов сражаться.', 18, 'https://www.afisha.ru/movie/otec-1003161/'),
        ('Наследник', 'naslednik', 'КАРО 11 Октябрь', 'Роскошные особняки, частные самолёты и даже собственные острова — состояние клана Редфеллоу исчисляется миллиардами. Всё это мечтает унаследовать Беккет, хоть глава рода и отрёкся от него ещё до его рождения. Однако на пути к заветной жизни стоят семь избалованных богачей. Чтобы ускорить свою очередь на наследство, Беккету придётся «спилить» несколько ветвей семейного древа. И сделать это так дерзко и изобретательно, чтобы избежать наказания за убийства.', 18, 'https://www.afisha.ru/movie/naslednik-1000452/'),
        ('Толстяк Юзи', 'tolstyak-yuzi', 'КАРО 11 Октябрь', 'Юзи — редчайший представитель семейства кошачьих. Только обитает он не в тропических лесах, а в Нью-Йорке, где он живет в шикарной квартире с джакузи и спортзалом. А ещё он умеет водить машину и любит тортики со сладкой газировкой. Всё меняет телепередача о диких леопардах. Вместе с бездомным котом Бобби Юзи решает найти своих родных', 12, 'https://www.afisha.ru/movie/tolstyak-yuzi-1003105/'),
        ('Не одна дома 3. Выпускной', 'ne-odna-doma-3', 'КАРО 11 Октябрь', 'Маша опять сталкивается с коварной Няней. Но теперь у злодейки появился новый союзник — обаятельный аферист Антон. На кону — школьный выпускной! Вместе с другом Егором Маше предстоит остановить злоумышленников и спасти самый важный вечер года.', 12, 'https://www.afisha.ru/movie/ne-odna-doma-3-vypuskniy-1001393/'),
        ('Красный призрак 1812', 'krasnyy-prizrak-1812', 'КАРО 11 Октябрь', 'Имение семьи юного князя Михаила Романовского безжалостно грабят фуражиры отступающей армии Наполеона. Молодого дворянина спасает от гибели партизан-одиночка — Призрак. Решив отомстить командиру фуражного отряда, Михаил увязывается за партизаном и проходит путь взросления от романтика, мечтающего о военной славе, до героя Отечественной войны 1812 года.', 18, 'https://www.afisha.ru/movie/krasniy-prizrak-1812-1003169'),
        ('Песни джиннов', 'pesni-dzhinnov', 'КАРО 11 Октябрь', 'Две актрисы блуждают по улицам Варанаси в новогоднюю ночь. Тем временем герои из Гоа готовят музыкальное мероприятие с участием звезд эстрады. Однажды они окажутся на корабле, где будут по очереди рассказывать свои истории. Корабль поплывет по тихой реке, а на берегу все-все-все будут прислушиваться: и летучие мыши в разрушенных дворцах, и бродячие фокусники, и гастролирующие диджеи. Им надо расслышать пение джиннов.', 18, 'https://www.afisha.ru/movie/pesni-dzhinnov-1003170'),
        ('Королек моей любви', 'korolek-moei-lyubvi', 'КАРО 11 Октябрь', 'Ироничная комедия о любовных перипетиях в современном мегаполисе.', 18, 'https://www.afisha.ru/movie/korolek-moei-lyubvi-1003240/'),
        ('На деревню дедушке', 'na-derevnyu-dedushke', 'КАРО 11 Октябрь', '10-летний Владик — гений нейросетей и мастер по избавлению от нянь. Когда маме срочно нужно уехать в командировку, ему приходится отправиться в деревню к дедушке, которого он ни разу в жизни не видел. Владик мечтает вернуться в город и попасть на важный IT-конкурс, а дед вовсе не намерен идти у него на поводу. Каждый из них уверен, что легко перехитрит другого, но эта встреча изменит их обоих.', 12, 'https://www.afisha.ru/movie/na-derevnyu-dedushke-1000160/'),
        ('Не одна дома 2', 'ne-odna-doma-2', 'КАРО 11 Октябрь', 'Маша отправляется с папой в летнее путешествие в автодоме. Три года назад мама и папа пообещали дочке, что будут проводить больше времени вместе и теперь поездка на природу — семейная традиция. Но Маша уже выросла, и ей стало скучно проводить столько времени с родителями. Душа подростка рвется к новым приключениям, однако взрослые этого совсем не понимают. Одной из остановок на пути семьи становится кемпинг. Там же останавливается автобус, который везет подростков в летний лагерь. Маша знакомится с одной из девочек. По счастливой случайности ее тоже зовут Маша, и она совсем не хочет ехать в лагерь. Недолго думая, девочки меняются местами: Маша отправляется в лагерь, а «другая Маша» спокойно возвращается домой. В лагере Маша встречает своих «старых» знакомых — жуликов Хахаля и Няню, которые снова замышляют что-то нехорошее. Маша со своей новой подружкой Викой решают дать им отпор и разрабатывают хитроумный план с изощренными ловушками, чтобы помешать злодеям осуществить задуманное.', 12, 'https://www.afisha.ru/movie/ne-odna-doma-2-1001297/')
) AS v(title, slug, location_description, full_description, age_limit, source_url)
WHERE u.email = 'seed.author@cityhawk.local'
ON CONFLICT (slug) DO UPDATE
SET title = EXCLUDED.title,
    location_description = EXCLUDED.location_description,
    full_description = EXCLUDED.full_description,
    age_limit = EXCLUDED.age_limit,
    source_url = EXCLUDED.source_url,
    updated_at = now();

-- 5. Привязка событий к кинотеатру для карты
INSERT INTO event_place (event_id, place_id)
SELECT e.id, p.id
FROM (
    VALUES
        ('angely-ladogi'), ('vot-eto-drama'), ('zhdun-2'), ('domovenok-kuzya-2'),
        ('dyavol-nosit-prada-2'), ('kommersant'), ('sem-verst-do-rassveta'), ('proekt-konec-sveta'),
        ('litvyak'), ('moya-sobaka-kosmonavt'), ('gruzovichki'), ('poluton'),
        ('graciya'), ('otec'), ('naslednik'), ('tolstyak-yuzi'),
        ('ne-odna-doma-3'), ('krasnyy-prizrak-1812'), ('pesni-dzhinnov'),
        ('korolek-moei-lyubvi'), ('na-derevnyu-dedushke'), ('ne-odna-doma-2')
) AS v(event_slug)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = 'КАРО 11 Октябрь'
            AND p.address_line = 'Москва, ул. Новый Арбат, 24'
JOIN city c ON c.id = p.city_id
           AND c.country_name = 'Россия'
           AND c.name = 'Москва'
ON CONFLICT (event_id) DO UPDATE
SET place_id = EXCLUDED.place_id,
    updated_at = now();

-- 6. Сеансы в КАРО 11 Октябрь
INSERT INTO event_session (event_id, place_id, start_at, end_at, price)
SELECT e.id, p.id, v.start_at::timestamptz, v.end_at::timestamptz, v.price
FROM (
    VALUES
        ('angely-ladogi', '2026-06-01 10:00:00+03', '2026-06-01 11:55:00+03', 650),
        ('vot-eto-drama', '2026-06-01 12:20:00+03', '2026-06-01 14:05:00+03', 620),
        ('zhdun-2', '2026-06-01 14:30:00+03', '2026-06-01 16:05:00+03', 550),
        ('domovenok-kuzya-2', '2026-06-01 16:30:00+03', '2026-06-01 18:00:00+03', 500),
        ('dyavol-nosit-prada-2', '2026-06-01 18:30:00+03', '2026-06-01 20:40:00+03', 850),
        ('kommersant', '2026-06-01 21:10:00+03', '2026-06-01 23:05:00+03', 700),
        ('sem-verst-do-rassveta', '2026-06-02 10:00:00+03', '2026-06-02 11:55:00+03', 600),
        ('proekt-konec-sveta', '2026-06-02 12:20:00+03', '2026-06-02 14:25:00+03', 750),
        ('litvyak', '2026-06-02 14:50:00+03', '2026-06-02 16:45:00+03', 650),
        ('moya-sobaka-kosmonavt', '2026-06-02 17:10:00+03', '2026-06-02 18:45:00+03', 520),
        ('gruzovichki', '2026-06-02 19:10:00+03', '2026-06-02 20:35:00+03', 480),
        ('poluton', '2026-06-02 21:00:00+03', '2026-06-02 22:45:00+03', 680),
        ('graciya', '2026-06-03 10:00:00+03', '2026-06-03 11:50:00+03', 620),
        ('otec', '2026-06-03 12:15:00+03', '2026-06-03 14:05:00+03', 650),
        ('naslednik', '2026-06-03 14:30:00+03', '2026-06-03 16:25:00+03', 700),
        ('tolstyak-yuzi', '2026-06-03 16:50:00+03', '2026-06-03 18:20:00+03', 520),
        ('ne-odna-doma-3', '2026-06-03 18:45:00+03', '2026-06-03 20:25:00+03', 600),
        ('krasnyy-prizrak-1812', '2026-06-03 20:50:00+03', '2026-06-03 22:55:00+03', 720),
        ('pesni-dzhinnov', '2026-06-04 10:00:00+03', '2026-06-04 11:50:00+03', 620),
        ('korolek-moei-lyubvi', '2026-06-04 12:15:00+03', '2026-06-04 13:55:00+03', 580),
        ('na-derevnyu-dedushke', '2026-06-04 14:20:00+03', '2026-06-04 16:00:00+03', 560),
        ('ne-odna-doma-2', '2026-06-04 16:25:00+03', '2026-06-04 18:05:00+03', 560)
) AS v(event_slug, start_at, end_at, price)
JOIN event e ON e.slug = v.event_slug
JOIN place p ON p.name = 'КАРО 11 Октябрь'
            AND p.address_line = 'Москва, ул. Новый Арбат, 24'
JOIN city c ON c.id = p.city_id
           AND c.country_name = 'Россия'
           AND c.name = 'Москва'
ON CONFLICT (event_id, place_id, start_at) DO UPDATE
SET end_at = EXCLUDED.end_at,
    price = EXCLUDED.price,
    updated_at = now();

-- 7. Изображения событий
DELETE FROM event_image ei
USING event e
WHERE ei.event_id = e.id
  AND e.slug IN (
    'angely-ladogi', 'vot-eto-drama', 'zhdun-2', 'domovenok-kuzya-2',
    'dyavol-nosit-prada-2', 'kommersant', 'sem-verst-do-rassveta', 'proekt-konec-sveta',
    'litvyak', 'moya-sobaka-kosmonavt', 'gruzovichki', 'poluton',
    'graciya', 'otec', 'naslednik', 'tolstyak-yuzi',
    'ne-odna-doma-3', 'krasnyy-prizrak-1812', 'pesni-dzhinnov',
    'korolek-moei-lyubvi', 'na-derevnyu-dedushke', 'ne-odna-doma-2'
  );

INSERT INTO event_image (event_id, image_url)
SELECT e.id, v.image_url
FROM (
    VALUES
        ('angely-ladogi', '/uploads/events/1.webp'),
        ('vot-eto-drama', '/uploads/events/2.webp'),
        ('zhdun-2', '/uploads/events/3.webp'),
        ('domovenok-kuzya-2', '/uploads/events/4.webp'),
        ('dyavol-nosit-prada-2', '/uploads/events/5.webp'),
        ('kommersant', '/uploads/events/6.webp'),
        ('sem-verst-do-rassveta', '/uploads/events/7.webp'),
        ('proekt-konec-sveta', '/uploads/events/8.webp'),
        ('litvyak', '/uploads/events/9.webp'),
        ('moya-sobaka-kosmonavt', '/uploads/events/10.webp'),
        ('gruzovichki', '/uploads/events/11.webp'),
        ('poluton', '/uploads/events/12.webp'),
        ('graciya', '/uploads/events/13.webp'),
        ('otec', '/uploads/events/14.webp'),
        ('naslednik', '/uploads/events/15.webp'),
        ('tolstyak-yuzi', '/uploads/events/16.webp'),
        ('ne-odna-doma-3', '/uploads/events/17.webp'),
        ('krasnyy-prizrak-1812', '/uploads/events/18.webp'),
        ('pesni-dzhinnov', '/uploads/events/19.webp'),
        ('korolek-moei-lyubvi', '/uploads/events/22.webp'),
        ('na-derevnyu-dedushke', '/uploads/events/20.webp'),
        ('ne-odna-doma-2', '/uploads/events/21.webp')
) AS v(event_slug, image_url)
JOIN event e ON e.slug = v.event_slug
ON CONFLICT (event_id, image_url) DO NOTHING;

-- 8. Категория «Кино» для всех событий
INSERT INTO event_category (event_id, category_id)
SELECT e.id, c.id
FROM event e
JOIN category c ON c.name = 'Кино'
WHERE e.slug IN (
    'angely-ladogi', 'vot-eto-drama', 'zhdun-2', 'domovenok-kuzya-2',
    'dyavol-nosit-prada-2', 'kommersant', 'sem-verst-do-rassveta', 'proekt-konec-sveta',
    'litvyak', 'moya-sobaka-kosmonavt', 'gruzovichki', 'poluton',
    'graciya', 'otec', 'naslednik', 'tolstyak-yuzi',
    'ne-odna-doma-3', 'krasnyy-prizrak-1812', 'pesni-dzhinnov',
    'korolek-moei-lyubvi', 'na-derevnyu-dedushke', 'ne-odna-doma-2'
)
ON CONFLICT (event_id, category_id) DO NOTHING;

-- 9. Теги событий
INSERT INTO event_tag (event_id, tag_id)
SELECT e.id, t.id
FROM (
    VALUES
        ('angely-ladogi', 'Кино'), ('angely-ladogi', 'Исторический'), ('angely-ladogi', 'Драма'), ('angely-ladogi', 'Российское кино'),
        ('vot-eto-drama', 'Кино'), ('vot-eto-drama', 'Ромком'), ('vot-eto-drama', 'Комедия'), ('vot-eto-drama', 'Премьера'),
        ('zhdun-2', 'Кино'), ('zhdun-2', 'Фэнтези'), ('zhdun-2', 'Комедия'), ('zhdun-2', 'Российское кино'),
        ('domovenok-kuzya-2', 'Кино'), ('domovenok-kuzya-2', 'Фэнтези'), ('domovenok-kuzya-2', 'Семейный'), ('domovenok-kuzya-2', 'Российское кино'),
        ('dyavol-nosit-prada-2', 'Кино'), ('dyavol-nosit-prada-2', 'Трагикомедия'), ('dyavol-nosit-prada-2', 'Зарубежное кино'), ('dyavol-nosit-prada-2', 'Новинка'),
        ('kommersant', 'Кино'), ('kommersant', 'Драма'), ('kommersant', 'Триллер'), ('kommersant', 'Российское кино'),
        ('sem-verst-do-rassveta', 'Кино'), ('sem-verst-do-rassveta', 'Исторический'), ('sem-verst-do-rassveta', 'Драма'), ('sem-verst-do-rassveta', 'Российское кино'),
        ('proekt-konec-sveta', 'Кино'), ('proekt-konec-sveta', 'Фантастика'), ('proekt-konec-sveta', 'Боевик'), ('proekt-konec-sveta', 'Зарубежное кино'),
        ('litvyak', 'Кино'), ('litvyak', 'Биография'), ('litvyak', 'Драма'), ('litvyak', 'Исторический'),
        ('moya-sobaka-kosmonavt', 'Кино'), ('moya-sobaka-kosmonavt', 'Приключение'), ('moya-sobaka-kosmonavt', 'Семейный'), ('moya-sobaka-kosmonavt', 'Российское кино'),
        ('gruzovichki', 'Кино'), ('gruzovichki', 'Мультфильм'), ('gruzovichki', 'Комедия'), ('gruzovichki', 'Зарубежное кино'),
        ('poluton', 'Кино'), ('poluton', 'Триллер'), ('poluton', 'Ужасы'), ('poluton', 'Зарубежное кино'),
        ('graciya', 'Кино'), ('graciya', 'Драма'), ('graciya', 'Биография'), ('graciya', 'Зарубежное кино'),
        ('otec', 'Кино'), ('otec', 'Драма'), ('otec', 'Российское кино'), ('otec', 'Премьера'),
        ('naslednik', 'Кино'), ('naslednik', 'Триллер'), ('naslednik', 'Детектив'), ('naslednik', 'Российское кино'),
        ('tolstyak-yuzi', 'Кино'), ('tolstyak-yuzi', 'Мультфильм'), ('tolstyak-yuzi', 'Комедия'), ('tolstyak-yuzi', 'Зарубежное кино'),
        ('ne-odna-doma-3', 'Кино'), ('ne-odna-doma-3', 'Комедия'), ('ne-odna-doma-3', 'Семейный'), ('ne-odna-doma-3', 'Российское кино'),
        ('krasnyy-prizrak-1812', 'Кино'), ('krasnyy-prizrak-1812', 'Боевик'), ('krasnyy-prizrak-1812', 'Исторический'), ('krasnyy-prizrak-1812', 'Российское кино'),
        ('pesni-dzhinnov', 'Кино'), ('pesni-dzhinnov', 'Фэнтези'), ('pesni-dzhinnov', 'Драма'), ('pesni-dzhinnov', 'Российское кино'),
        ('korolek-moei-lyubvi', 'Кино'), ('korolek-moei-lyubvi', 'Комедия'), ('korolek-moei-lyubvi', 'Мелодрама'), ('korolek-moei-lyubvi', 'Российское кино'),
        ('na-derevnyu-dedushke', 'Кино'), ('na-derevnyu-dedushke', 'Комедия'), ('na-derevnyu-dedushke', 'Российское кино'),
        ('ne-odna-doma-2', 'Кино'), ('ne-odna-doma-2', 'Комедия'), ('ne-odna-doma-2', 'Семейный'), ('ne-odna-doma-2', 'Российское кино')
) AS v(event_slug, tag_name)
JOIN event e ON e.slug = v.event_slug
JOIN tag t ON t.name = v.tag_name
ON CONFLICT (event_id, tag_id) DO NOTHING;

COMMIT;
