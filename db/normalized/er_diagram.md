erDiagram

    USER_ACCOUNT {
        uuid id PK
        text email UK
        text username
        text user_surname
        text password_hash
        smallint status
        text avatar_url
        datetime created_at
        datetime updated_at
    }

    CITY {
        uuid id PK
        text name UK
        text country_name
        text timezone
        datetime created_at
        datetime updated_at
    }

    PLACE {
        uuid id PK
        uuid city_id FK
        text name
        text address_line
        text latitude
        text longitude
        text description
        datetime created_at
        datetime updated_at
    }

    EVENT {
        uuid id PK
        uuid author_user_id FK
        uuid place_id FK
        text title
        text short_description
        text full_description
        int age_limit
        text status
        text source_url
        datetime created_at
        datetime updated_at
    }

    EVENT_SESSION {
        uuid id PK
        uuid event_id FK
        uuid place_id FK
        datetime start_at
        datetime end_at
        int price
        text status
        datetime created_at
        datetime updated_at
    }

    EVENT_IMAGE {
        uuid id PK
        uuid event_id FK
        text image_url
        datetime created_at
        datetime updated_at
    }

    CATEGORY {
        uuid id PK
        text name UK
        datetime created_at
        datetime updated_at
    }

    TAG {
        uuid id PK
        text name UK
        datetime created_at
        datetime updated_at
    }

    EVENT_CATEGORY {
        uuid event_id PK FK
        uuid category_id PK FK
        datetime created_at
        datetime updated_at
    }

    EVENT_TAG {
        uuid event_id PK FK
        uuid tag_id PK FK
        datetime created_at
        datetime updated_at
    }

    COLLECTION {
        uuid id PK
        uuid author_user_id FK
        text title
        text description
        bool is_public
        datetime created_at
        datetime updated_at
    }

    COLLECTION_IMAGE {
        uuid id PK
        uuid collection_id FK
        text image_url
        datetime created_at
        datetime updated_at
    }

    COLLECTION_EVENT {
        uuid collection_id PK FK
        uuid event_id PK FK
        datetime created_at
        datetime updated_at
    }

    FAVORITE_EVENT {
        uuid user_id PK FK
        uuid event_id PK FK
        datetime created_at
        datetime updated_at
    }

    USER_FOLLOW {
        uuid follower_user_id PK FK
        uuid followed_user_id PK FK
        datetime created_at
        datetime updated_at
    }

    EVENT_INVITATION {
        uuid id PK
        uuid event_id FK
        uuid event_session_id FK
        uuid sender_user_id FK
        uuid recipient_user_id FK
        text message_text
        smallint status
        datetime responded_at
        datetime created_at
        datetime updated_at
    }

    SHARE_LINK {
        uuid id PK
        uuid creator_user_id FK
        uuid event_id FK
        uuid collection_id FK
        text share_token
        datetime created_at
        datetime updated_at
    }

    NOTIFICATION {
        uuid id PK
        uuid recipient_user_id FK
        uuid author_user_id FK
        uuid event_id FK
        uuid event_session_id FK
        uuid invitation_id FK
        uuid collection_id FK
        smallint notification_type
        text text
        bool is_read
        datetime read_at
        datetime created_at
        datetime updated_at
    }

    PLACE }o--|| CITY : "city_id FK ON UPDATE CASCADE ON DELETE RESTRICT"
    EVENT }o--|| USER_ACCOUNT : "author_user_id FK ON UPDATE CASCADE ON DELETE RESTRICT"
    EVENT }o--|| PLACE : "place_id FK ON UPDATE CASCADE ON DELETE RESTRICT"
    EVENT_SESSION }o--|| EVENT : "event_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    EVENT_SESSION }o--|| PLACE : "place_id FK ON UPDATE CASCADE ON DELETE RESTRICT"
    EVENT_IMAGE }o--|| EVENT : "event_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    EVENT_CATEGORY }o--|| EVENT : "event_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    EVENT_CATEGORY }o--|| CATEGORY : "category_id FK ON UPDATE CASCADE ON DELETE RESTRICT"
    EVENT_TAG }o--|| EVENT : "event_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    EVENT_TAG }o--|| TAG : "tag_id FK ON UPDATE CASCADE ON DELETE RESTRICT"
    COLLECTION }o--|| USER_ACCOUNT : "author_user_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    COLLECTION_IMAGE }o--|| COLLECTION : "collection_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    COLLECTION_EVENT }o--|| COLLECTION : "collection_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    COLLECTION_EVENT }o--|| EVENT : "event_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    FAVORITE_EVENT }o--|| USER_ACCOUNT : "user_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    FAVORITE_EVENT }o--|| EVENT : "event_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    USER_FOLLOW }o--|| USER_ACCOUNT : "follower_user_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    USER_FOLLOW }o--|| USER_ACCOUNT : "followed_user_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    EVENT_INVITATION }o--|| EVENT : "event_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    EVENT_INVITATION }o--o| EVENT_SESSION : "event_session_id FK ON UPDATE CASCADE ON DELETE SET NULL"
    EVENT_INVITATION }o--|| USER_ACCOUNT : "sender_user_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    EVENT_INVITATION }o--|| USER_ACCOUNT : "recipient_user_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    SHARE_LINK }o--|| USER_ACCOUNT : "creator_user_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    SHARE_LINK }o--o| EVENT : "event_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    SHARE_LINK }o--o| COLLECTION : "collection_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    NOTIFICATION }o--|| USER_ACCOUNT : "recipient_user_id FK ON UPDATE CASCADE ON DELETE CASCADE"
    NOTIFICATION }o--o| USER_ACCOUNT : "author_user_id FK ON UPDATE CASCADE ON DELETE SET NULL"
    NOTIFICATION }o--o| EVENT : "event_id FK ON UPDATE CASCADE ON DELETE SET NULL"
    NOTIFICATION }o--o| EVENT_SESSION : "event_session_id FK ON UPDATE CASCADE ON DELETE SET NULL"
    NOTIFICATION }o--o| EVENT_INVITATION : "invitation_id FK ON UPDATE CASCADE ON DELETE SET NULL"
    NOTIFICATION }o--o| COLLECTION : "collection_id FK ON UPDATE CASCADE ON DELETE SET NULL"

    %% All PK fields are NOT NULL explicitly in DDL
    %% UNIQUE:
    %% USER_ACCOUNT.email
    %% CITY (name country_name)
    %% CATEGORY.name
    %% TAG.name
    %% SHARE_LINK.share_token
    %% DEFAULT:
    %% is_public default false in COLLECTION
    %% is_read default false in NOTIFICATION
    %% created_at default now() and updated_at default now() in all tables
    %% CHECK:
    %% latitude and longitude ranges in PLACE
    %% age_limit range in EVENT
    %% start_at < end_at in EVENT_SESSION
    %% price >= 0 in EVENT_SESSION
    %% follower_user_id <> followed_user_id in USER_FOLLOW
    %% sender_user_id <> recipient_user_id in EVENT_INVITATION
    %% read state consistency in NOTIFICATION: is_read with read_at
    %% Trigger usage:
    %% BEFORE UPDATE trigger set_updated_at() updates updated_at automatically
