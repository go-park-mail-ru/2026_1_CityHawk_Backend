BEGIN;

DROP INDEX IF EXISTS event_external_source_external_id_uniq;

ALTER TABLE event
    DROP COLUMN IF EXISTS external_source,
    DROP COLUMN IF EXISTS external_id;

COMMIT;
