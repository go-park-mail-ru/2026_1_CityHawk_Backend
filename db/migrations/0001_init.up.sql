CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$$;

CREATE TABLE IF NOT EXISTS city (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    country_name text NOT NULL,
    timezone text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT city_name_valid CHECK (char_length(btrim(name)) BETWEEN 1 AND 100),
    CONSTRAINT city_country_name_valid CHECK (char_length(btrim(country_name)) BETWEEN 1 AND 100),
    CONSTRAINT city_timezone_valid CHECK (char_length(btrim(timezone)) BETWEEN 1 AND 64),
    CONSTRAINT city_country_name_name_key UNIQUE (country_name, name)
);

COMMENT ON TABLE city IS 'Справочник городов. Значения по умолчанию заданы только для технических полей id, created_at и updated_at; бизнес-атрибуты должны приходить из пользовательских или импортируемых данных.';

CREATE TABLE IF NOT EXISTS user_account (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL,
    username text NOT NULL,
    user_surname text NOT NULL,
    password_hash text NOT NULL,
    birthday date,
    city_id uuid,
    avatar_url text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_account_email_key UNIQUE (email),
    CONSTRAINT user_account_email_format CHECK (position('@' in email) > 1 AND position(' ' in email) = 0),
    CONSTRAINT user_account_email_valid CHECK (char_length(btrim(email)) BETWEEN 1 AND 254),
    CONSTRAINT user_account_username_valid CHECK (char_length(btrim(username)) BETWEEN 3 AND 64),
    CONSTRAINT user_account_user_surname_valid CHECK (char_length(btrim(user_surname)) BETWEEN 1 AND 64),
    CONSTRAINT user_account_password_hash_valid CHECK (char_length(btrim(password_hash)) BETWEEN 1 AND 255),
    CONSTRAINT user_account_avatar_url_format CHECK (avatar_url IS NULL OR avatar_url ~ '^(https?://|/uploads/)'),
    CONSTRAINT user_account_avatar_url_length CHECK (avatar_url IS NULL OR char_length(avatar_url) <= 2048),
    CONSTRAINT user_account_city_id_fkey
        FOREIGN KEY (city_id)
        REFERENCES city(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL
);

COMMENT ON TABLE user_account IS 'Учетные записи пользователей. У birthday и avatar_url нет default, потому что это необязательные пользовательские данные; city_id может отсутствовать до выбора города.';

CREATE TABLE IF NOT EXISTS refresh_session (
    token_hash text PRIMARY KEY,
    user_id uuid NOT NULL,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT refresh_session_token_hash_valid CHECK (char_length(btrim(token_hash)) = 64),
    CONSTRAINT refresh_session_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES user_account(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

CREATE INDEX idx_refresh_session_user_id ON refresh_session(user_id);
CREATE INDEX idx_refresh_session_expires_at ON refresh_session(expires_at);

COMMENT ON TABLE refresh_session IS 'Сессии refresh-токенов. Хранится только хеш токена, связанный с пользователем и временем истечения, что соответствует серверной логике ротации и отзыва токенов.';

CREATE TABLE IF NOT EXISTS place (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    city_id uuid NOT NULL,
    name text NOT NULL,
    address_line text NOT NULL,
    latitude numeric(9, 6) NOT NULL,
    longitude numeric(9, 6) NOT NULL,
    description text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT place_name_valid CHECK (char_length(btrim(name)) BETWEEN 1 AND 200),
    CONSTRAINT place_address_line_valid CHECK (char_length(btrim(address_line)) BETWEEN 1 AND 300),
    CONSTRAINT place_description_length CHECK (description IS NULL OR char_length(description) <= 5000),
    CONSTRAINT place_latitude_range CHECK (latitude BETWEEN -90 AND 90),
    CONSTRAINT place_longitude_range CHECK (longitude BETWEEN -180 AND 180),
    CONSTRAINT place_city_id_fkey
        FOREIGN KEY (city_id)
        REFERENCES city(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);

COMMENT ON TABLE place IS 'Места проведения. У description нет default, потому что описание может быть неизвестно на момент создания записи.';

CREATE TABLE IF NOT EXISTS category (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT category_name_key UNIQUE (name),
    CONSTRAINT category_name_valid CHECK (char_length(btrim(name)) BETWEEN 1 AND 64)
);

COMMENT ON TABLE category IS 'Справочник категорий событий. Имя категории должно быть уникальным; иных business-default значений нет.';

CREATE TABLE IF NOT EXISTS tag (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT tag_name_key UNIQUE (name),
    CONSTRAINT tag_name_valid CHECK (char_length(btrim(name)) BETWEEN 1 AND 64)
);

COMMENT ON TABLE tag IS 'Справочник тегов событий. Имя тега задается явно и не получает значение по умолчанию.';

CREATE TABLE IF NOT EXISTS event (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    author_user_id uuid NOT NULL,
    title text NOT NULL,
    location_description text NOT NULL,
    full_description text,
    age_limit integer NOT NULL DEFAULT 0,
    source_url text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT event_title_valid CHECK (char_length(btrim(title)) BETWEEN 1 AND 200),
    CONSTRAINT event_location_description_valid CHECK (char_length(btrim(location_description)) BETWEEN 1 AND 500),
    CONSTRAINT event_full_description_length CHECK (full_description IS NULL OR char_length(full_description) <= 5000),
    CONSTRAINT event_age_limit_range CHECK (age_limit BETWEEN 0 AND 21),
    CONSTRAINT event_source_url_format CHECK (source_url IS NULL OR source_url ~ '^https?://'),
    CONSTRAINT event_source_url_length CHECK (source_url IS NULL OR char_length(source_url) <= 2048),
    CONSTRAINT event_author_user_id_fkey
        FOREIGN KEY (author_user_id)
        REFERENCES user_account(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);

COMMENT ON TABLE event IS 'Карточки событий. age_limit по умолчанию равен 0 как безопасное значение; full_description и source_url остаются без default, так как они необязательны и зависят от редакторского ввода.';

CREATE TABLE IF NOT EXISTS event_session (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL,
    place_id uuid NOT NULL,
    start_at timestamptz NOT NULL,
    end_at timestamptz NOT NULL,
    price integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT event_session_time_order CHECK (start_at < end_at),
    CONSTRAINT event_session_price_non_negative CHECK (price >= 0),
    CONSTRAINT event_session_price_reasonable CHECK (price <= 10000000),
    CONSTRAINT event_session_event_id_fkey
        FOREIGN KEY (event_id)
        REFERENCES event(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT event_session_place_id_fkey
        FOREIGN KEY (place_id)
        REFERENCES place(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,
    CONSTRAINT event_session_place_time_excl
        EXCLUDE USING gist (
            place_id WITH =,
            tstzrange(start_at, end_at, '[)') WITH &&
        )
);

COMMENT ON TABLE event_session IS 'Конкретные сеансы событий. price по умолчанию равен 0 для бесплатных событий; EXCLUDE запрещает пересечение интервалов в одном месте проведения.';

CREATE TABLE IF NOT EXISTS event_image (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id uuid NOT NULL,
    image_url text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT event_image_image_url_format CHECK (image_url ~ '^(https?://|/uploads/)'),
    CONSTRAINT event_image_image_url_length CHECK (char_length(image_url) <= 2048),
    CONSTRAINT event_image_event_id_fkey
        FOREIGN KEY (event_id)
        REFERENCES event(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

COMMENT ON TABLE event_image IS 'Изображения событий. URL обязателен и не имеет default, потому что ссылка на файл должна задаваться явно.';

CREATE TABLE IF NOT EXISTS event_category (
    event_id uuid NOT NULL,
    category_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT event_category_pkey PRIMARY KEY (event_id, category_id),
    CONSTRAINT event_category_event_id_fkey
        FOREIGN KEY (event_id)
        REFERENCES event(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT event_category_category_id_fkey
        FOREIGN KEY (category_id)
        REFERENCES category(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);

COMMENT ON TABLE event_category IS 'Связь событий с категориями. created_at и updated_at нужны для аудита назначения категории.';

CREATE TABLE IF NOT EXISTS event_tag (
    event_id uuid NOT NULL,
    tag_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT event_tag_pkey PRIMARY KEY (event_id, tag_id),
    CONSTRAINT event_tag_event_id_fkey
        FOREIGN KEY (event_id)
        REFERENCES event(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT event_tag_tag_id_fkey
        FOREIGN KEY (tag_id)
        REFERENCES tag(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);

COMMENT ON TABLE event_tag IS 'Связь событий с тегами. Значения по умолчанию нужны только для технических временных полей.';

CREATE TABLE IF NOT EXISTS collection (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    author_user_id uuid NOT NULL,
    title text NOT NULL,
    description text,
    is_public boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT collection_title_valid CHECK (char_length(btrim(title)) BETWEEN 1 AND 200),
    CONSTRAINT collection_description_length CHECK (description IS NULL OR char_length(description) <= 5000),
    CONSTRAINT collection_author_user_id_fkey
        FOREIGN KEY (author_user_id)
        REFERENCES user_account(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);

COMMENT ON TABLE collection IS 'Пользовательские подборки событий. is_public по умолчанию false, чтобы исключить случайную публикацию.';

CREATE TABLE IF NOT EXISTS collection_image (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    collection_id uuid NOT NULL,
    image_url text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT collection_image_image_url_format CHECK (image_url ~ '^(https?://|/uploads/)'),
    CONSTRAINT collection_image_image_url_length CHECK (char_length(image_url) <= 2048),
    CONSTRAINT collection_image_collection_id_fkey
        FOREIGN KEY (collection_id)
        REFERENCES collection(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

COMMENT ON TABLE collection_image IS 'Изображения подборок. URL обязателен и задается явно без default.';

CREATE TABLE IF NOT EXISTS collection_event (
    collection_id uuid NOT NULL,
    event_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT collection_event_pkey PRIMARY KEY (collection_id, event_id),
    CONSTRAINT collection_event_collection_id_fkey
        FOREIGN KEY (collection_id)
        REFERENCES collection(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT collection_event_event_id_fkey
        FOREIGN KEY (event_id)
        REFERENCES event(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

COMMENT ON TABLE collection_event IS 'Связь подборок с событиями. Составной первичный ключ не допускает дубликатов одной и той же пары.';

CREATE TABLE IF NOT EXISTS favorite_event (
    user_id uuid NOT NULL,
    event_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT favorite_event_pkey PRIMARY KEY (user_id, event_id),
    CONSTRAINT favorite_event_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES user_account(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT favorite_event_event_id_fkey
        FOREIGN KEY (event_id)
        REFERENCES event(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

COMMENT ON TABLE favorite_event IS 'Избранные события пользователя. Технические даты позволяют хранить историю добавления в избранное.';

CREATE TABLE IF NOT EXISTS user_follow (
    follower_user_id uuid NOT NULL,
    followed_user_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_follow_pkey PRIMARY KEY (follower_user_id, followed_user_id),
    CONSTRAINT user_follow_no_self CHECK (follower_user_id <> followed_user_id),
    CONSTRAINT user_follow_follower_user_id_fkey
        FOREIGN KEY (follower_user_id)
        REFERENCES user_account(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT user_follow_followed_user_id_fkey
        FOREIGN KEY (followed_user_id)
        REFERENCES user_account(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

COMMENT ON TABLE user_follow IS 'Подписки пользователей друг на друга. CHECK запрещает самоподписку.';

CREATE TABLE IF NOT EXISTS event_invitation (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_user_id uuid NOT NULL,
    recipient_user_id uuid NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    message_text text,
    responded_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT event_invitation_sender_recipient_diff CHECK (sender_user_id <> recipient_user_id),
    CONSTRAINT event_invitation_status_valid CHECK (status IN ('pending', 'accepted', 'declined', 'cancelled')),
    CONSTRAINT event_invitation_message_text_length CHECK (message_text IS NULL OR char_length(message_text) <= 500),
    CONSTRAINT event_invitation_sender_user_id_fkey
        FOREIGN KEY (sender_user_id)
        REFERENCES user_account(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,
    CONSTRAINT event_invitation_recipient_user_id_fkey
        FOREIGN KEY (recipient_user_id)
        REFERENCES user_account(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);

COMMENT ON TABLE event_invitation IS 'Базовая сущность приглашения. У message_text и responded_at нет default, потому что сообщение необязательно, а момент ответа появляется только после реакции получателя.';

CREATE TABLE IF NOT EXISTS event_invitation_event (
    invitation_id uuid PRIMARY KEY,
    event_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT event_invitation_event_invitation_id_fkey
        FOREIGN KEY (invitation_id)
        REFERENCES event_invitation(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT event_invitation_event_event_id_fkey
        FOREIGN KEY (event_id)
        REFERENCES event(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

COMMENT ON TABLE event_invitation_event IS 'Привязка приглашения ко всему событию.';

CREATE TABLE IF NOT EXISTS event_invitation_session (
    invitation_id uuid PRIMARY KEY,
    event_session_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT event_invitation_session_invitation_id_fkey
        FOREIGN KEY (invitation_id)
        REFERENCES event_invitation(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT event_invitation_session_event_session_id_fkey
        FOREIGN KEY (event_session_id)
        REFERENCES event_session(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

COMMENT ON TABLE event_invitation_session IS 'Привязка приглашения к конкретному сеансу.';

CREATE TABLE IF NOT EXISTS share_link (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_user_id uuid,
    share_token text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT share_link_share_token_key UNIQUE (share_token),
    CONSTRAINT share_link_share_token_valid CHECK (char_length(btrim(share_token)) BETWEEN 6 AND 128),
    CONSTRAINT share_link_creator_user_id_fkey
        FOREIGN KEY (creator_user_id)
        REFERENCES user_account(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);

COMMENT ON TABLE share_link IS 'Базовая сущность публичной ссылки. Токен уникален и не получает default, потому что должен создаваться приложением или безопасной функцией генерации.';

CREATE TABLE IF NOT EXISTS share_link_event (
    share_link_id uuid PRIMARY KEY,
    event_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT share_link_event_share_link_id_fkey
        FOREIGN KEY (share_link_id)
        REFERENCES share_link(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT share_link_event_event_id_fkey
        FOREIGN KEY (event_id)
        REFERENCES event(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

COMMENT ON TABLE share_link_event IS 'Привязка публичной ссылки к событию.';

CREATE TABLE IF NOT EXISTS share_link_collection (
    share_link_id uuid PRIMARY KEY,
    collection_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT share_link_collection_share_link_id_fkey
        FOREIGN KEY (share_link_id)
        REFERENCES share_link(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT share_link_collection_collection_id_fkey
        FOREIGN KEY (collection_id)
        REFERENCES collection(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

COMMENT ON TABLE share_link_collection IS 'Привязка публичной ссылки к подборке.';

CREATE TABLE IF NOT EXISTS notification (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_user_id uuid NOT NULL,
    notification_type text NOT NULL,
    is_read boolean NOT NULL DEFAULT false,
    read_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT notification_type_valid CHECK (char_length(btrim(notification_type)) BETWEEN 1 AND 64),
    CONSTRAINT notification_read_state CHECK (
        (is_read = false AND read_at IS NULL) OR
        (is_read = true AND read_at IS NOT NULL)
    ),
    CONSTRAINT notification_recipient_user_id_fkey
        FOREIGN KEY (recipient_user_id)
        REFERENCES user_account(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

COMMENT ON TABLE notification IS 'Базовая сущность уведомления. is_read по умолчанию false, а read_at появляется только после фактического прочтения.';

CREATE TABLE IF NOT EXISTS notification_actor (
    notification_id uuid PRIMARY KEY,
    author_user_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT notification_actor_notification_id_fkey
        FOREIGN KEY (notification_id)
        REFERENCES notification(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT notification_actor_author_user_id_fkey
        FOREIGN KEY (author_user_id)
        REFERENCES user_account(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);

COMMENT ON TABLE notification_actor IS 'Связь уведомления с пользователем-инициатором.';

CREATE TABLE IF NOT EXISTS notification_event (
    notification_id uuid PRIMARY KEY,
    event_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT notification_event_notification_id_fkey
        FOREIGN KEY (notification_id)
        REFERENCES notification(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT notification_event_event_id_fkey
        FOREIGN KEY (event_id)
        REFERENCES event(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

COMMENT ON TABLE notification_event IS 'Связь уведомления с событием.';

CREATE TABLE IF NOT EXISTS notification_event_session (
    notification_id uuid PRIMARY KEY,
    event_session_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT notification_event_session_notification_id_fkey
        FOREIGN KEY (notification_id)
        REFERENCES notification(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT notification_event_session_event_session_id_fkey
        FOREIGN KEY (event_session_id)
        REFERENCES event_session(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

COMMENT ON TABLE notification_event_session IS 'Связь уведомления с конкретным сеансом события.';

CREATE TABLE IF NOT EXISTS notification_invitation (
    notification_id uuid PRIMARY KEY,
    invitation_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT notification_invitation_notification_id_fkey
        FOREIGN KEY (notification_id)
        REFERENCES notification(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT notification_invitation_invitation_id_fkey
        FOREIGN KEY (invitation_id)
        REFERENCES event_invitation(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

COMMENT ON TABLE notification_invitation IS 'Связь уведомления с приглашением.';

CREATE TABLE IF NOT EXISTS notification_collection (
    notification_id uuid PRIMARY KEY,
    collection_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT notification_collection_notification_id_fkey
        FOREIGN KEY (notification_id)
        REFERENCES notification(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT notification_collection_collection_id_fkey
        FOREIGN KEY (collection_id)
        REFERENCES collection(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

COMMENT ON TABLE notification_collection IS 'Связь уведомления с подборкой.';

CREATE TRIGGER set_user_account_updated_at
BEFORE UPDATE ON user_account
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_city_updated_at
BEFORE UPDATE ON city
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_place_updated_at
BEFORE UPDATE ON place
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_category_updated_at
BEFORE UPDATE ON category
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_tag_updated_at
BEFORE UPDATE ON tag
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_event_updated_at
BEFORE UPDATE ON event
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_event_session_updated_at
BEFORE UPDATE ON event_session
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_event_image_updated_at
BEFORE UPDATE ON event_image
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_event_category_updated_at
BEFORE UPDATE ON event_category
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_event_tag_updated_at
BEFORE UPDATE ON event_tag
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_collection_updated_at
BEFORE UPDATE ON collection
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_collection_image_updated_at
BEFORE UPDATE ON collection_image
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_collection_event_updated_at
BEFORE UPDATE ON collection_event
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_favorite_event_updated_at
BEFORE UPDATE ON favorite_event
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_user_follow_updated_at
BEFORE UPDATE ON user_follow
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_event_invitation_updated_at
BEFORE UPDATE ON event_invitation
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_share_link_updated_at
BEFORE UPDATE ON share_link
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_event_invitation_event_updated_at
BEFORE UPDATE ON event_invitation_event
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_event_invitation_session_updated_at
BEFORE UPDATE ON event_invitation_session
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_share_link_event_updated_at
BEFORE UPDATE ON share_link_event
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_share_link_collection_updated_at
BEFORE UPDATE ON share_link_collection
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_notification_updated_at
BEFORE UPDATE ON notification
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_notification_actor_updated_at
BEFORE UPDATE ON notification_actor
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_notification_event_updated_at
BEFORE UPDATE ON notification_event
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_notification_event_session_updated_at
BEFORE UPDATE ON notification_event_session
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_notification_invitation_updated_at
BEFORE UPDATE ON notification_invitation
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_notification_collection_updated_at
BEFORE UPDATE ON notification_collection
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
