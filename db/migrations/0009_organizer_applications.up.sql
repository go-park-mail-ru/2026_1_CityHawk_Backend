CREATE TABLE IF NOT EXISTS organizer_application (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    name text NOT NULL,
    email text NOT NULL,
    phone text NOT NULL,
    city text NOT NULL,
    project_name text NOT NULL,
    categories text NOT NULL,
    links text,
    about text NOT NULL,
    review_comment text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT organizer_application_status_valid CHECK (status IN ('pending', 'needs_info', 'approved', 'rejected')),
    CONSTRAINT organizer_application_name_valid CHECK (char_length(btrim(name)) BETWEEN 1 AND 200),
    CONSTRAINT organizer_application_email_valid CHECK (position('@' in email) > 1 AND position(' ' in email) = 0 AND char_length(btrim(email)) BETWEEN 1 AND 254),
    CONSTRAINT organizer_application_phone_valid CHECK (char_length(btrim(phone)) BETWEEN 1 AND 64),
    CONSTRAINT organizer_application_city_valid CHECK (char_length(btrim(city)) BETWEEN 1 AND 100),
    CONSTRAINT organizer_application_project_name_valid CHECK (char_length(btrim(project_name)) BETWEEN 1 AND 200),
    CONSTRAINT organizer_application_categories_valid CHECK (char_length(btrim(categories)) BETWEEN 1 AND 500),
    CONSTRAINT organizer_application_links_length CHECK (links IS NULL OR char_length(links) <= 2000),
    CONSTRAINT organizer_application_about_length CHECK (char_length(btrim(about)) BETWEEN 1 AND 5000),
    CONSTRAINT organizer_application_review_comment_length CHECK (review_comment IS NULL OR char_length(review_comment) <= 2000),
    CONSTRAINT organizer_application_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES user_account(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS organizer_application_active_user_key
ON organizer_application(user_id)
WHERE status IN ('pending', 'needs_info');

CREATE INDEX IF NOT EXISTS idx_organizer_application_status_created_at
ON organizer_application(status, created_at DESC);

DROP TRIGGER IF EXISTS set_organizer_application_updated_at ON organizer_application;

CREATE TRIGGER set_organizer_application_updated_at
BEFORE UPDATE ON organizer_application
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
