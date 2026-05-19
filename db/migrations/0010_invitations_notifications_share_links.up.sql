ALTER TABLE event_invitation
ADD COLUMN IF NOT EXISTS status text NOT NULL DEFAULT 'pending';

ALTER TABLE event_invitation
DROP CONSTRAINT IF EXISTS event_invitation_status_valid;

ALTER TABLE event_invitation
ADD CONSTRAINT event_invitation_status_valid
CHECK (status IN ('pending', 'accepted', 'declined', 'cancelled'));

ALTER TABLE event_invitation
DROP CONSTRAINT IF EXISTS event_invitation_message_text_length;

ALTER TABLE event_invitation
ADD CONSTRAINT event_invitation_message_text_length
CHECK (message_text IS NULL OR char_length(message_text) <= 500);

ALTER TABLE share_link
ALTER COLUMN creator_user_id DROP NOT NULL;

ALTER TABLE share_link
DROP CONSTRAINT IF EXISTS share_link_share_token_valid;

ALTER TABLE share_link
ADD CONSTRAINT share_link_share_token_valid
CHECK (char_length(btrim(share_token)) BETWEEN 6 AND 128);

CREATE INDEX IF NOT EXISTS idx_event_invitation_event_event_id
ON event_invitation_event(event_id);

CREATE INDEX IF NOT EXISTS idx_event_invitation_recipient_status_created_at
ON event_invitation(recipient_user_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notification_recipient_read_created_at
ON notification(recipient_user_id, is_read, created_at DESC);
