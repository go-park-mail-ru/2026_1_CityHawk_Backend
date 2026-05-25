BEGIN;

DELETE FROM event_tag et
USING event e
WHERE et.event_id = e.id
  AND e.slug IN (
      'nurlan-saburov-adult-talk', 'ilya-sobolev-here-and-now', 'stas-starovoitov-private',
      'new-voices-standup', 'zoya-yarovitsyna-character', 'comedians-vs-audience',
      'dmitry-kozoma-new-material'
  );

DELETE FROM event_category ec
USING event e
WHERE ec.event_id = e.id
  AND e.slug IN (
      'nurlan-saburov-adult-talk', 'ilya-sobolev-here-and-now', 'stas-starovoitov-private',
      'new-voices-standup', 'zoya-yarovitsyna-character', 'comedians-vs-audience',
      'dmitry-kozoma-new-material'
  );

DELETE FROM event_image ei
USING event e
WHERE ei.event_id = e.id
  AND e.slug IN (
      'nurlan-saburov-adult-talk', 'ilya-sobolev-here-and-now', 'stas-starovoitov-private',
      'new-voices-standup', 'zoya-yarovitsyna-character', 'comedians-vs-audience',
      'dmitry-kozoma-new-material'
  );

DELETE FROM event_session es
USING event e
WHERE es.event_id = e.id
  AND e.slug IN (
      'nurlan-saburov-adult-talk', 'ilya-sobolev-here-and-now', 'stas-starovoitov-private',
      'new-voices-standup', 'zoya-yarovitsyna-character', 'comedians-vs-audience',
      'dmitry-kozoma-new-material'
  );

DELETE FROM event_place ep
USING event e
WHERE ep.event_id = e.id
  AND e.slug IN (
      'nurlan-saburov-adult-talk', 'ilya-sobolev-here-and-now', 'stas-starovoitov-private',
      'new-voices-standup', 'zoya-yarovitsyna-character', 'comedians-vs-audience',
      'dmitry-kozoma-new-material'
  );

DELETE FROM event
WHERE slug IN (
    'nurlan-saburov-adult-talk', 'ilya-sobolev-here-and-now', 'stas-starovoitov-private',
    'new-voices-standup', 'zoya-yarovitsyna-character', 'comedians-vs-audience',
    'dmitry-kozoma-new-material'
);

COMMIT;
