BEGIN;

DELETE FROM event_tag et
USING event e
WHERE et.event_id = e.id
  AND e.slug IN (
      'angely-ladogi', 'vot-eto-drama', 'zhdun-2', 'domovenok-kuzya-2',
      'dyavol-nosit-prada-2', 'kommersant', 'sem-verst-do-rassveta', 'proekt-konec-sveta',
      'litvyak', 'moya-sobaka-kosmonavt', 'gruzovichki', 'poluton',
      'graciya', 'otec', 'naslednik', 'tolstyak-yuzi',
      'ne-odna-doma-3', 'krasnyy-prizrak-1812', 'pesni-dzhinnov',
      'korolek-moei-lyubvi', 'na-derevnyu-dedushke', 'ne-odna-doma-2'
  );

DELETE FROM event_category ec
USING event e
WHERE ec.event_id = e.id
  AND e.slug IN (
      'angely-ladogi', 'vot-eto-drama', 'zhdun-2', 'domovenok-kuzya-2',
      'dyavol-nosit-prada-2', 'kommersant', 'sem-verst-do-rassveta', 'proekt-konec-sveta',
      'litvyak', 'moya-sobaka-kosmonavt', 'gruzovichki', 'poluton',
      'graciya', 'otec', 'naslednik', 'tolstyak-yuzi',
      'ne-odna-doma-3', 'krasnyy-prizrak-1812', 'pesni-dzhinnov',
      'korolek-moei-lyubvi', 'na-derevnyu-dedushke', 'ne-odna-doma-2'
  );

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

DELETE FROM event_session es
USING event e
WHERE es.event_id = e.id
  AND e.slug IN (
      'angely-ladogi', 'vot-eto-drama', 'zhdun-2', 'domovenok-kuzya-2',
      'dyavol-nosit-prada-2', 'kommersant', 'sem-verst-do-rassveta', 'proekt-konec-sveta',
      'litvyak', 'moya-sobaka-kosmonavt', 'gruzovichki', 'poluton',
      'graciya', 'otec', 'naslednik', 'tolstyak-yuzi',
      'ne-odna-doma-3', 'krasnyy-prizrak-1812', 'pesni-dzhinnov',
      'korolek-moei-lyubvi', 'na-derevnyu-dedushke', 'ne-odna-doma-2'
  );

DELETE FROM event_place ep
USING event e
WHERE ep.event_id = e.id
  AND e.slug IN (
      'angely-ladogi', 'vot-eto-drama', 'zhdun-2', 'domovenok-kuzya-2',
      'dyavol-nosit-prada-2', 'kommersant', 'sem-verst-do-rassveta', 'proekt-konec-sveta',
      'litvyak', 'moya-sobaka-kosmonavt', 'gruzovichki', 'poluton',
      'graciya', 'otec', 'naslednik', 'tolstyak-yuzi',
      'ne-odna-doma-3', 'krasnyy-prizrak-1812', 'pesni-dzhinnov',
      'korolek-moei-lyubvi', 'na-derevnyu-dedushke', 'ne-odna-doma-2'
  );

DELETE FROM event
WHERE slug IN (
    'angely-ladogi', 'vot-eto-drama', 'zhdun-2', 'domovenok-kuzya-2',
    'dyavol-nosit-prada-2', 'kommersant', 'sem-verst-do-rassveta', 'proekt-konec-sveta',
    'litvyak', 'moya-sobaka-kosmonavt', 'gruzovichki', 'poluton',
    'graciya', 'otec', 'naslednik', 'tolstyak-yuzi',
    'ne-odna-doma-3', 'krasnyy-prizrak-1812', 'pesni-dzhinnov',
    'korolek-moei-lyubvi', 'na-derevnyu-dedushke', 'ne-odna-doma-2'
);

DELETE FROM place
WHERE name = 'КАРО 11 Октябрь'
  AND address_line = 'Москва, ул. Новый Арбат, 24'
  AND NOT EXISTS (
      SELECT 1
      FROM event_place ep
      WHERE ep.place_id = place.id
  )
  AND NOT EXISTS (
      SELECT 1
      FROM event_session es
      WHERE es.place_id = place.id
  );

COMMIT;
