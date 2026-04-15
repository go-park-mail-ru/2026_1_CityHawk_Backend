ALTER TABLE user_account
    DROP CONSTRAINT IF EXISTS user_account_avatar_url_format;

ALTER TABLE user_account
    ADD CONSTRAINT user_account_avatar_url_format
    CHECK (avatar_url IS NULL OR avatar_url ~ '^https?://');

ALTER TABLE event_image
    DROP CONSTRAINT IF EXISTS event_image_image_url_format;

ALTER TABLE event_image
    ADD CONSTRAINT event_image_image_url_format
    CHECK (image_url ~ '^https?://');

ALTER TABLE collection_image
    DROP CONSTRAINT IF EXISTS collection_image_image_url_format;

ALTER TABLE collection_image
    ADD CONSTRAINT collection_image_image_url_format
    CHECK (image_url ~ '^https?://');
