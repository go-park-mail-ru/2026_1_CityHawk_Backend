DROP INDEX IF EXISTS idx_user_account_role;

ALTER TABLE user_account
DROP CONSTRAINT IF EXISTS user_account_role_valid;

ALTER TABLE user_account
DROP COLUMN IF EXISTS role;
