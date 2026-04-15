DROP INDEX IF EXISTS idx_collection_title_trgm;
DROP INDEX IF EXISTS idx_tag_name_trgm;
DROP INDEX IF EXISTS idx_category_name_trgm;
DROP INDEX IF EXISTS idx_event_title_trgm;
DROP EXTENSION IF EXISTS pg_trgm;
