DROP TABLE IF EXISTS support_ticket_message;

ALTER TABLE support_ticket
ADD COLUMN IF NOT EXISTS support_comment text;

ALTER TABLE support_ticket
DROP CONSTRAINT IF EXISTS support_ticket_support_comment_length;

ALTER TABLE support_ticket
ADD CONSTRAINT support_ticket_support_comment_length
CHECK (support_comment IS NULL OR char_length(support_comment) <= 5000);
