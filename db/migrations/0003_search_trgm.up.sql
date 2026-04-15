CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_event_title_trgm
    ON event
    USING gin (title gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_category_name_trgm
    ON category
    USING gin (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_tag_name_trgm
    ON tag
    USING gin (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_collection_title_trgm
    ON collection
    USING gin (title gin_trgm_ops);
