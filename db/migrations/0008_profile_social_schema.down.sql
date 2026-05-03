ALTER TABLE user_account
ADD COLUMN IF NOT EXISTS role text NOT NULL DEFAULT 'user';

UPDATE user_account ua
SET role = COALESCE((
    SELECT CASE
        WHEN bool_or(ur.role = 'admin') THEN 'admin'
        WHEN bool_or(ur.role = 'organizer') THEN 'organizer'
        ELSE 'user'
    END
    FROM user_role ur
    WHERE ur.user_id = ua.id
), 'user');

ALTER TABLE user_account
DROP CONSTRAINT IF EXISTS user_account_role_valid;

ALTER TABLE user_account
ADD CONSTRAINT user_account_role_valid
CHECK (role IN ('user', 'organizer', 'admin'));

CREATE INDEX IF NOT EXISTS idx_user_account_role ON user_account(role);

DROP TRIGGER IF EXISTS set_organizer_profile_updated_at ON organizer_profile;

DROP TABLE IF EXISTS user_interest_tag;
DROP TABLE IF EXISTS organizer_profile;
DROP TABLE IF EXISTS user_role;

ALTER TABLE user_account
DROP CONSTRAINT IF EXISTS user_account_bio_length;

ALTER TABLE user_account
DROP COLUMN IF EXISTS bio;

