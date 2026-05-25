BEGIN;

ALTER TABLE user_account
    ALTER COLUMN username SET DEFAULT '',
    ALTER COLUMN user_surname SET DEFAULT '';

ALTER TABLE user_account
    DROP CONSTRAINT IF EXISTS user_account_username_valid,
    DROP CONSTRAINT IF EXISTS user_account_user_surname_valid;

ALTER TABLE user_account
    ADD CONSTRAINT user_account_username_valid
        CHECK (char_length(btrim(username)) = 0 OR char_length(btrim(username)) BETWEEN 3 AND 64),
    ADD CONSTRAINT user_account_user_surname_valid
        CHECK (char_length(btrim(user_surname)) = 0 OR char_length(btrim(user_surname)) BETWEEN 1 AND 64);

COMMIT;
