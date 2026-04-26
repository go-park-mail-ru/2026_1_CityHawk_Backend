ALTER TABLE user_account
ADD COLUMN IF NOT EXISTS role text NOT NULL DEFAULT 'user';

ALTER TABLE user_account
DROP CONSTRAINT IF EXISTS user_account_role_valid;

ALTER TABLE user_account
ADD CONSTRAINT user_account_role_valid
CHECK (role IN ('user', 'admin'));

CREATE INDEX IF NOT EXISTS idx_user_account_role ON user_account(role);
