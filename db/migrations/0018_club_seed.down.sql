BEGIN;

DELETE FROM event_tag et
USING event e
WHERE et.event_id = e.id
  AND e.slug IN (
      'arma-gravity-openair', 'mutabor-latin-night', 'powerhouse-bass-parade',
      'aglomerat-2000s', 'rock-garage-party', 'backstage-dud', 'solstice-openair'
  );

DELETE FROM event_category ec
USING event e
WHERE ec.event_id = e.id
  AND e.slug IN (
      'arma-gravity-openair', 'mutabor-latin-night', 'powerhouse-bass-parade',
      'aglomerat-2000s', 'rock-garage-party', 'backstage-dud', 'solstice-openair'
  );

DELETE FROM event_image ei
USING event e
WHERE ei.event_id = e.id
  AND e.slug IN (
      'arma-gravity-openair', 'mutabor-latin-night', 'powerhouse-bass-parade',
      'aglomerat-2000s', 'rock-garage-party', 'backstage-dud', 'solstice-openair'
  );

DELETE FROM event_session es
USING event e
WHERE es.event_id = e.id
  AND e.slug IN (
      'arma-gravity-openair', 'mutabor-latin-night', 'powerhouse-bass-parade',
      'aglomerat-2000s', 'rock-garage-party', 'backstage-dud', 'solstice-openair'
  );

DELETE FROM event_place ep
USING event e
WHERE ep.event_id = e.id
  AND e.slug IN (
      'arma-gravity-openair', 'mutabor-latin-night', 'powerhouse-bass-parade',
      'aglomerat-2000s', 'rock-garage-party', 'backstage-dud', 'solstice-openair'
  );

DELETE FROM event
WHERE slug IN (
    'arma-gravity-openair', 'mutabor-latin-night', 'powerhouse-bass-parade',
    'aglomerat-2000s', 'rock-garage-party', 'backstage-dud', 'solstice-openair'
);

COMMIT;
