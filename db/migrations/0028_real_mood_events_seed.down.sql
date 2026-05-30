BEGIN;

DELETE FROM event_tag et
USING event e
WHERE et.event_id = e.id
  AND e.slug IN (
      'piknik-afishi-2026-moscow',
      'nautilus-pompilius-butusov-vdnh-2026',
      'alexander-frolov-vdnh-2026',
      'tsvety-simvol-krasoty-tretyakovka-2026',
      'moskva-ot-zakata-do-rassveta-odeon-2026'
  );

DELETE FROM event_category ec
USING event e
WHERE ec.event_id = e.id
  AND e.slug IN (
      'piknik-afishi-2026-moscow',
      'nautilus-pompilius-butusov-vdnh-2026',
      'alexander-frolov-vdnh-2026',
      'tsvety-simvol-krasoty-tretyakovka-2026',
      'moskva-ot-zakata-do-rassveta-odeon-2026'
  );

DELETE FROM event_image ei
USING event e
WHERE ei.event_id = e.id
  AND e.slug IN (
      'piknik-afishi-2026-moscow',
      'nautilus-pompilius-butusov-vdnh-2026',
      'alexander-frolov-vdnh-2026',
      'tsvety-simvol-krasoty-tretyakovka-2026',
      'moskva-ot-zakata-do-rassveta-odeon-2026'
  );

DELETE FROM event_session es
USING event e
WHERE es.event_id = e.id
  AND e.slug IN (
      'piknik-afishi-2026-moscow',
      'nautilus-pompilius-butusov-vdnh-2026',
      'alexander-frolov-vdnh-2026',
      'tsvety-simvol-krasoty-tretyakovka-2026',
      'moskva-ot-zakata-do-rassveta-odeon-2026'
  );

DELETE FROM event_place ep
USING event e
WHERE ep.event_id = e.id
  AND e.slug IN (
      'piknik-afishi-2026-moscow',
      'nautilus-pompilius-butusov-vdnh-2026',
      'alexander-frolov-vdnh-2026',
      'tsvety-simvol-krasoty-tretyakovka-2026',
      'moskva-ot-zakata-do-rassveta-odeon-2026'
  );

DELETE FROM event
WHERE slug IN (
    'piknik-afishi-2026-moscow',
    'nautilus-pompilius-butusov-vdnh-2026',
    'alexander-frolov-vdnh-2026',
    'tsvety-simvol-krasoty-tretyakovka-2026',
    'moskva-ot-zakata-do-rassveta-odeon-2026'
);

COMMIT;
