CREATE TABLE IF NOT EXISTS support_ticket (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    category text NOT NULL,
    status text NOT NULL DEFAULT 'open',
    title text NOT NULL,
    message text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    closed_at timestamptz,

    CONSTRAINT support_ticket_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES user_account(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,

    CONSTRAINT support_ticket_category_valid
        CHECK (category IN ('bug', 'suggestion', 'product_complaint', 'other')),

    CONSTRAINT support_ticket_status_valid
        CHECK (status IN ('open', 'in_progress', 'closed')),

    CONSTRAINT support_ticket_title_valid
        CHECK (char_length(btrim(title)) BETWEEN 3 AND 200),

    CONSTRAINT support_ticket_message_valid
        CHECK (char_length(btrim(message)) BETWEEN 10 AND 5000),

    CONSTRAINT support_ticket_closed_at_valid
        CHECK (
            (status = 'closed' AND closed_at IS NOT NULL)
            OR (status <> 'closed' AND closed_at IS NULL)
        )
);

CREATE INDEX IF NOT EXISTS idx_support_ticket_user_id ON support_ticket(user_id);
CREATE INDEX IF NOT EXISTS idx_support_ticket_status ON support_ticket(status);
CREATE INDEX IF NOT EXISTS idx_support_ticket_category ON support_ticket(category);
CREATE INDEX IF NOT EXISTS idx_support_ticket_created_at ON support_ticket(created_at DESC);

CREATE TRIGGER set_support_ticket_updated_at
BEFORE UPDATE ON support_ticket
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
