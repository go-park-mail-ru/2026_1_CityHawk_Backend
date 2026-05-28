DELETE FROM user_follow uf
USING user_account follower, user_account followed
WHERE uf.follower_user_id = follower.id
  AND uf.followed_user_id = followed.id
  AND follower.email IN (
    'seed.author@cityhawk.local',
    'anna.friend@cityhawk.local',
    'boris.friend@cityhawk.local',
    'katya.friend@cityhawk.local',
    'misha.friend@cityhawk.local'
  )
  AND followed.email IN (
    'seed.author@cityhawk.local',
    'anna.friend@cityhawk.local',
    'boris.friend@cityhawk.local',
    'katya.friend@cityhawk.local',
    'misha.friend@cityhawk.local'
  );

DELETE FROM user_account
WHERE email IN (
    'anna.friend@cityhawk.local',
    'boris.friend@cityhawk.local',
    'katya.friend@cityhawk.local',
    'misha.friend@cityhawk.local'
);
