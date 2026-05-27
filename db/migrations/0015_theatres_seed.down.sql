BEGIN;

DELETE FROM event_tag et
USING event e
WHERE et.event_id = e.id
  AND e.slug IN (
      'kabala-svyatosh', 'intuicziya', 'zhil-byl-dom', 'lebedinoe-ozero',
      'romeo-i-dzhuletta-lyubov-vne-vremeni', 'dve-anny', 'pesn-lyubvi-pesn-skorbi',
      'lir', 'gamlet', 'prizrak-myuzikla', 'moskva-petushki', 'idioty',
      'revizor-komediya-v-stihah', 'maskarad'
  );

DELETE FROM event_category ec
USING event e
WHERE ec.event_id = e.id
  AND e.slug IN (
      'kabala-svyatosh', 'intuicziya', 'zhil-byl-dom', 'lebedinoe-ozero',
      'romeo-i-dzhuletta-lyubov-vne-vremeni', 'dve-anny', 'pesn-lyubvi-pesn-skorbi',
      'lir', 'gamlet', 'prizrak-myuzikla', 'moskva-petushki', 'idioty',
      'revizor-komediya-v-stihah', 'maskarad'
  );

DELETE FROM event_image ei
USING event e
WHERE ei.event_id = e.id
  AND e.slug IN (
      'kabala-svyatosh', 'intuicziya', 'zhil-byl-dom', 'lebedinoe-ozero',
      'romeo-i-dzhuletta-lyubov-vne-vremeni', 'dve-anny', 'pesn-lyubvi-pesn-skorbi',
      'lir', 'gamlet', 'prizrak-myuzikla', 'moskva-petushki', 'idioty',
      'revizor-komediya-v-stihah', 'maskarad'
  );

DELETE FROM event_session es
USING event e
WHERE es.event_id = e.id
  AND e.slug IN (
      'kabala-svyatosh', 'intuicziya', 'zhil-byl-dom', 'lebedinoe-ozero',
      'romeo-i-dzhuletta-lyubov-vne-vremeni', 'dve-anny', 'pesn-lyubvi-pesn-skorbi',
      'lir', 'gamlet', 'prizrak-myuzikla', 'moskva-petushki', 'idioty',
      'revizor-komediya-v-stihah', 'maskarad'
  );

DELETE FROM event_place ep
USING event e
WHERE ep.event_id = e.id
  AND e.slug IN (
      'kabala-svyatosh', 'intuicziya', 'zhil-byl-dom', 'lebedinoe-ozero',
      'romeo-i-dzhuletta-lyubov-vne-vremeni', 'dve-anny', 'pesn-lyubvi-pesn-skorbi',
      'lir', 'gamlet', 'prizrak-myuzikla', 'moskva-petushki', 'idioty',
      'revizor-komediya-v-stihah', 'maskarad'
  );

DELETE FROM event
WHERE slug IN (
    'kabala-svyatosh', 'intuicziya', 'zhil-byl-dom', 'lebedinoe-ozero',
    'romeo-i-dzhuletta-lyubov-vne-vremeni', 'dve-anny', 'pesn-lyubvi-pesn-skorbi',
    'lir', 'gamlet', 'prizrak-myuzikla', 'moskva-petushki', 'idioty',
    'revizor-komediya-v-stihah', 'maskarad'
);

COMMIT;
