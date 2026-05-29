-- Тестовые пользователи для проверки поиска друзей, подписок и приглашений.
-- Пароль для всех аккаунтов: password

INSERT INTO city (name, country_name, timezone)
VALUES ('Москва', 'Россия', 'Europe/Moscow')
ON CONFLICT (country_name, name) DO UPDATE
SET timezone = EXCLUDED.timezone,
    updated_at = now();

INSERT INTO user_account (
    email,
    username,
    password_hash,
    city_id
)
SELECT
    v.email,
    v.username,
    '$2a$10$lwUrEa2Iv3VJZeNT94SZK.soSzsTnmxcScFBI9H58MM8KYSQwTHOG',
    c.id
FROM city c
CROSS JOIN (
    VALUES
        ('anna.friend@cityhawk.local', 'anna_friend'),
        ('boris.friend@cityhawk.local', 'boris_friend'),
        ('katya.friend@cityhawk.local', 'katya_friend'),
        ('misha.friend@cityhawk.local', 'misha_friend')
) AS v(email, username)
WHERE c.country_name = 'Россия'
  AND c.name = 'Москва'
ON CONFLICT (email) DO UPDATE
SET username = EXCLUDED.username,
    password_hash = EXCLUDED.password_hash,
    city_id = EXCLUDED.city_id,
    updated_at = now();

WITH seed_users AS (
    SELECT id, email
    FROM user_account
    WHERE email IN (
        'seed.author@cityhawk.local',
        'anna.friend@cityhawk.local',
        'boris.friend@cityhawk.local',
        'katya.friend@cityhawk.local',
        'misha.friend@cityhawk.local'
    )
),
follow_pairs AS (
    SELECT follower.id AS follower_user_id, followed.id AS followed_user_id
    FROM seed_users follower
    CROSS JOIN seed_users followed
    WHERE follower.id <> followed.id
)
INSERT INTO user_follow (follower_user_id, followed_user_id)
SELECT follower_user_id, followed_user_id
FROM follow_pairs
ON CONFLICT (follower_user_id, followed_user_id) DO NOTHING;
