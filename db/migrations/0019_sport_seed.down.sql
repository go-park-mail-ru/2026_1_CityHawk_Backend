BEGIN;

DELETE FROM event_tag et
USING event e
WHERE et.event_id = e.id
  AND e.slug IN (
      'russian-cup-final-spartak-krasnodar-2026',
      'kharlamov-cup-final-spartak-loko-2026',
      'vtb-united-league-cska-lokomotiv-2026',
      'sportland-festival-2026',
      'moscow-half-marathon-2026',
      'moscow-ski-marathon-2026',
      'night-of-moscow-sport-2026',
      'proryv-extreme-sports-festival-2026'
  );

DELETE FROM event_category ec
USING event e
WHERE ec.event_id = e.id
  AND e.slug IN (
      'russian-cup-final-spartak-krasnodar-2026',
      'kharlamov-cup-final-spartak-loko-2026',
      'vtb-united-league-cska-lokomotiv-2026',
      'sportland-festival-2026',
      'moscow-half-marathon-2026',
      'moscow-ski-marathon-2026',
      'night-of-moscow-sport-2026',
      'proryv-extreme-sports-festival-2026'
  );

DELETE FROM event_image ei
USING event e
WHERE ei.event_id = e.id
  AND e.slug IN (
      'russian-cup-final-spartak-krasnodar-2026',
      'kharlamov-cup-final-spartak-loko-2026',
      'vtb-united-league-cska-lokomotiv-2026',
      'sportland-festival-2026',
      'moscow-half-marathon-2026',
      'moscow-ski-marathon-2026',
      'night-of-moscow-sport-2026',
      'proryv-extreme-sports-festival-2026'
  );

DELETE FROM event_session es
USING event e
WHERE es.event_id = e.id
  AND e.slug IN (
      'russian-cup-final-spartak-krasnodar-2026',
      'kharlamov-cup-final-spartak-loko-2026',
      'vtb-united-league-cska-lokomotiv-2026',
      'sportland-festival-2026',
      'moscow-half-marathon-2026',
      'moscow-ski-marathon-2026',
      'night-of-moscow-sport-2026',
      'proryv-extreme-sports-festival-2026'
  );

DELETE FROM event_place ep
USING event e
WHERE ep.event_id = e.id
  AND e.slug IN (
      'russian-cup-final-spartak-krasnodar-2026',
      'kharlamov-cup-final-spartak-loko-2026',
      'vtb-united-league-cska-lokomotiv-2026',
      'sportland-festival-2026',
      'moscow-half-marathon-2026',
      'moscow-ski-marathon-2026',
      'night-of-moscow-sport-2026',
      'proryv-extreme-sports-festival-2026'
  );

DELETE FROM event
WHERE slug IN (
    'russian-cup-final-spartak-krasnodar-2026',
    'kharlamov-cup-final-spartak-loko-2026',
    'vtb-united-league-cska-lokomotiv-2026',
    'sportland-festival-2026',
    'moscow-half-marathon-2026',
    'moscow-ski-marathon-2026',
    'night-of-moscow-sport-2026',
    'proryv-extreme-sports-festival-2026'
);

COMMIT;
