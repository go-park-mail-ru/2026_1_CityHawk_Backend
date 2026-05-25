BEGIN;

DELETE FROM collection_event ce
USING collection col
WHERE ce.collection_id = col.id
  AND col.slug = 'photo-spots-moscow';

DELETE FROM collection_image ci
USING collection col
WHERE ci.collection_id = col.id
  AND col.slug = 'photo-spots-moscow';

DELETE FROM collection
WHERE slug = 'photo-spots-moscow';

DELETE FROM event_tag et
USING event e
WHERE et.event_id = e.id
  AND e.slug IN (
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
  );

DELETE FROM event_category ec
USING event e
WHERE ec.event_id = e.id
  AND e.slug IN (
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
  );

DELETE FROM event_image ei
USING event e
WHERE ei.event_id = e.id
  AND e.slug IN (
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
  );

DELETE FROM event_place ep
USING event e
WHERE ep.event_id = e.id
  AND e.slug IN (
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
  );

DELETE FROM event
WHERE slug IN (
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
);

DELETE FROM place
WHERE (name, address_line) IN (
    ('Завод «Кристалл»', 'Москва, Самокатная ул., 4, стр. 1'),
    ('«Пресня Сити» и «Зотов»', 'Москва, Ходынская ул., 2, стр. 1'),
    ('Центр дизайна ARTPLAY', 'Москва, Нижняя Сыромятническая ул., 10, стр. 2'),
    ('ГЭС-2', 'Москва, Болотная наб., 15'),
    ('Центр «Графит»', 'Москва, Электродная ул., 2, стр. 32'),
    ('Оранжерея Главного ботанического сада', 'Москва, Ботаническая ул., 4, стр. 2'),
    ('Улица Школьная', 'Москва, ул. Школьная'),
    ('Аптекарский огород', 'Москва, просп. Мира, 26, стр. 1'),
    ('Депо Лесная', 'Москва, Лесная ул., 20'),
    ('Парк Горького', 'Москва, ул. Крымский Вал, 9'),
    ('Парк Сокольники', 'Москва, ул. Сокольнический Вал, 1, стр. 1'),
    ('Бизнес-центр «Аквамарин»', 'Москва, Озерковская наб., 24, стр. 2'),
    ('Нескучный сад', 'Москва, Нескучный сад'),
    ('Китай-город', 'Москва, метро Китай-город'),
    ('Новодевичий монастырь', 'Москва, Новодевичий пр., 1')
);

COMMIT;
