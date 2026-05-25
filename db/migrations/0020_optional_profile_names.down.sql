BEGIN;

UPDATE user_account
SET username = 'user_' || substr(id::text, 1, 8)
WHERE char_length(btrim(username)) = 0;

UPDATE user_account
SET user_surname = '-'
WHERE char_length(btrim(user_surname)) = 0;

ALTER TABLE user_account
    ALTER COLUMN username DROP DEFAULT,
    ALTER COLUMN user_surname DROP DEFAULT;

ALTER TABLE user_account
    DROP CONSTRAINT IF EXISTS user_account_username_valid,
    DROP CONSTRAINT IF EXISTS user_account_user_surname_valid;

ALTER TABLE user_account
    ADD CONSTRAINT user_account_username_valid
        CHECK (char_length(btrim(username)) BETWEEN 3 AND 64),
    ADD CONSTRAINT user_account_user_surname_valid
        CHECK (char_length(btrim(user_surname)) BETWEEN 1 AND 64);

COMMIT;
