BEGIN;

ALTER TABLE event
    ADD COLUMN IF NOT EXISTS external_source text,
    ADD COLUMN IF NOT EXISTS external_id text;

CREATE UNIQUE INDEX IF NOT EXISTS event_external_source_external_id_uniq
    ON event (external_source, external_id);

COMMIT;
