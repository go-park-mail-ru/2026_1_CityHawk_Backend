BEGIN;

DELETE FROM event_tag et
USING event e
WHERE et.event_id = e.id
  AND e.slug IN (
      'ayvazovsky-ocean', 'dali-surrealism', 'russian-avant-garde',
      'body-worlds-pulse', 'van-gogh-alive', 'see-by-hands', 'picasso-and-muses',
      'ocean-secrets-show', 'magicians-21', 'shining-moscow',
      'men-in-black-invasion', 'lioness-empire'
  );

DELETE FROM event_category ec
USING event e
WHERE ec.event_id = e.id
  AND e.slug IN (
      'ayvazovsky-ocean', 'dali-surrealism', 'russian-avant-garde',
      'body-worlds-pulse', 'van-gogh-alive', 'see-by-hands', 'picasso-and-muses',
      'ocean-secrets-show', 'magicians-21', 'shining-moscow',
      'men-in-black-invasion', 'lioness-empire'
  );

DELETE FROM event_image ei
USING event e
WHERE ei.event_id = e.id
  AND e.slug IN (
      'ayvazovsky-ocean', 'dali-surrealism', 'russian-avant-garde',
      'body-worlds-pulse', 'van-gogh-alive', 'see-by-hands', 'picasso-and-muses',
      'ocean-secrets-show', 'magicians-21', 'shining-moscow',
      'men-in-black-invasion', 'lioness-empire'
  );

DELETE FROM event_session es
USING event e
WHERE es.event_id = e.id
  AND e.slug IN (
      'ayvazovsky-ocean', 'dali-surrealism', 'russian-avant-garde',
      'body-worlds-pulse', 'van-gogh-alive', 'see-by-hands', 'picasso-and-muses',
      'ocean-secrets-show', 'magicians-21', 'shining-moscow',
      'men-in-black-invasion', 'lioness-empire'
  );

DELETE FROM event_place ep
USING event e
WHERE ep.event_id = e.id
  AND e.slug IN (
      'ayvazovsky-ocean', 'dali-surrealism', 'russian-avant-garde',
      'body-worlds-pulse', 'van-gogh-alive', 'see-by-hands', 'picasso-and-muses',
      'ocean-secrets-show', 'magicians-21', 'shining-moscow',
      'men-in-black-invasion', 'lioness-empire'
  );

DELETE FROM event
WHERE slug IN (
    'ayvazovsky-ocean', 'dali-surrealism', 'russian-avant-garde',
    'body-worlds-pulse', 'van-gogh-alive', 'see-by-hands', 'picasso-and-muses',
    'ocean-secrets-show', 'magicians-21', 'shining-moscow',
    'men-in-black-invasion', 'lioness-empire'
);

COMMIT;
