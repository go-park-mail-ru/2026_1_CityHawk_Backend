CREATE TABLE IF NOT EXISTS event_place (
    event_id uuid PRIMARY KEY,
    place_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT event_place_event_id_fkey
        FOREIGN KEY (event_id)
        REFERENCES event(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT event_place_place_id_fkey
        FOREIGN KEY (place_id)
        REFERENCES place(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);

COMMENT ON TABLE event_place IS 'Основное место события для отображения на карте независимо от расписания и сеансов.';

DROP TRIGGER IF EXISTS set_event_place_updated_at ON event_place;

CREATE TRIGGER set_event_place_updated_at
BEFORE UPDATE ON event_place
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
