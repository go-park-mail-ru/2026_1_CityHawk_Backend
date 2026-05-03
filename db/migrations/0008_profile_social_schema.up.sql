ALTER TABLE user_account
ADD COLUMN IF NOT EXISTS bio text;

ALTER TABLE user_account
DROP CONSTRAINT IF EXISTS user_account_bio_length;

ALTER TABLE user_account
ADD CONSTRAINT user_account_bio_length
CHECK (bio IS NULL OR char_length(bio) <= 1000);

CREATE TABLE IF NOT EXISTS user_role (
    user_id uuid NOT NULL,
    role text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_role_pkey PRIMARY KEY (user_id, role),
    CONSTRAINT user_role_allowed_values CHECK (role IN ('user', 'organizer', 'admin')),
    CONSTRAINT user_role_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES user_account(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'user_account'
          AND column_name = 'role'
    ) THEN
        EXECUTE $sql$
            INSERT INTO user_role (user_id, role, created_at)
            SELECT id, COALESCE(NULLIF(role, ''), 'user'), now()
            FROM user_account
            ON CONFLICT (user_id, role) DO NOTHING
        $sql$;
    END IF;
END;
$$;

INSERT INTO user_role (user_id, role, created_at)
SELECT id, 'user', now()
FROM user_account
WHERE NOT EXISTS (
    SELECT 1
    FROM user_role ur
    WHERE ur.user_id = user_account.id
)
ON CONFLICT (user_id, role) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_user_role_role ON user_role(role);

CREATE TABLE IF NOT EXISTS organizer_profile (
    user_id uuid PRIMARY KEY,
    display_name text NOT NULL,
    description text,
    website_url text,
    is_verified boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT organizer_profile_display_name_valid CHECK (char_length(btrim(display_name)) BETWEEN 1 AND 200),
    CONSTRAINT organizer_profile_description_length CHECK (description IS NULL OR char_length(description) <= 5000),
    CONSTRAINT organizer_profile_website_url_format CHECK (website_url IS NULL OR website_url ~ '^https?://'),
    CONSTRAINT organizer_profile_website_url_length CHECK (website_url IS NULL OR char_length(website_url) <= 2048),
    CONSTRAINT organizer_profile_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES user_account(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_interest_tag (
    user_id uuid NOT NULL,
    tag_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_interest_tag_pkey PRIMARY KEY (user_id, tag_id),
    CONSTRAINT user_interest_tag_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES user_account(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT user_interest_tag_tag_id_fkey
        FOREIGN KEY (tag_id)
        REFERENCES tag(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_interest_tag_tag_id ON user_interest_tag(tag_id);

DROP TRIGGER IF EXISTS set_organizer_profile_updated_at ON organizer_profile;

CREATE TRIGGER set_organizer_profile_updated_at
BEFORE UPDATE ON organizer_profile
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

ALTER TABLE user_account
DROP CONSTRAINT IF EXISTS user_account_role_valid;

DROP INDEX IF EXISTS idx_user_account_role;

ALTER TABLE user_account
DROP COLUMN IF EXISTS role;
