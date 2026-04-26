ALTER TABLE support_ticket
DROP COLUMN IF EXISTS support_comment;

CREATE TABLE IF NOT EXISTS support_ticket_message (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id uuid NOT NULL,
    author_user_id uuid NOT NULL,
    author_role text NOT NULL DEFAULT 'user',
    body text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT support_ticket_message_ticket_id_fkey
        FOREIGN KEY (ticket_id)
        REFERENCES support_ticket(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,

    CONSTRAINT support_ticket_message_author_user_id_fkey
        FOREIGN KEY (author_user_id)
        REFERENCES user_account(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,

    CONSTRAINT support_ticket_message_author_role_valid
        CHECK (author_role IN ('user', 'support', 'admin')),

    CONSTRAINT support_ticket_message_body_valid
        CHECK (char_length(btrim(body)) BETWEEN 1 AND 5000)
);

CREATE INDEX IF NOT EXISTS idx_support_ticket_message_ticket_id_created_at
ON support_ticket_message(ticket_id, created_at ASC, id ASC);

CREATE INDEX IF NOT EXISTS idx_support_ticket_message_author_user_id
ON support_ticket_message(author_user_id);
