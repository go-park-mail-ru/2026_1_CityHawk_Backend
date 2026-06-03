CREATE INDEX IF NOT EXISTS idx_event_created_at_id
ON event (created_at DESC, id ASC);

CREATE INDEX IF NOT EXISTS idx_event_title_id
ON event (title ASC, id ASC);

CREATE INDEX IF NOT EXISTS idx_event_author_created_at
ON event (author_user_id, created_at DESC, id ASC);

CREATE INDEX IF NOT EXISTS idx_event_session_event_end_start
ON event_session (event_id, end_at, start_at, id);

CREATE INDEX IF NOT EXISTS idx_event_session_start_at_id
ON event_session (start_at, id);

CREATE INDEX IF NOT EXISTS idx_event_session_end_at_event
ON event_session (end_at, event_id);

CREATE INDEX IF NOT EXISTS idx_event_image_event_created_id
ON event_image (event_id, created_at ASC, id ASC);

CREATE INDEX IF NOT EXISTS idx_event_place_place_event
ON event_place (place_id, event_id);

CREATE INDEX IF NOT EXISTS idx_place_city_id
ON place (city_id, id);

CREATE INDEX IF NOT EXISTS idx_city_lower_name
ON city (lower(name));

CREATE INDEX IF NOT EXISTS idx_event_lower_title_trgm
ON event
USING gin (lower(title) gin_trgm_ops);
