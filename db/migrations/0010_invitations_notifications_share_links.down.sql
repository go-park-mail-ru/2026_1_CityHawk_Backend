DROP INDEX IF EXISTS idx_notification_recipient_read_created_at;
DROP INDEX IF EXISTS idx_event_invitation_recipient_status_created_at;
DROP INDEX IF EXISTS idx_event_invitation_event_event_id;

ALTER TABLE share_link
DROP CONSTRAINT IF EXISTS share_link_share_token_valid;

ALTER TABLE share_link
ADD CONSTRAINT share_link_share_token_valid
CHECK (char_length(btrim(share_token)) BETWEEN 16 AND 128);

ALTER TABLE share_link
ALTER COLUMN creator_user_id SET NOT NULL;

ALTER TABLE event_invitation
DROP CONSTRAINT IF EXISTS event_invitation_message_text_length;

ALTER TABLE event_invitation
ADD CONSTRAINT event_invitation_message_text_length
CHECK (message_text IS NULL OR char_length(message_text) <= 2000);

ALTER TABLE event_invitation
DROP CONSTRAINT IF EXISTS event_invitation_status_valid;

ALTER TABLE event_invitation
DROP COLUMN IF EXISTS status;
