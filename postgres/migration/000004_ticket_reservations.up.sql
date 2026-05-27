CREATE TABLE ticket_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reserved_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending'
);

CREATE INDEX idx_ticket_reservations_ticket_id ON ticket_reservations(ticket_id);
CREATE INDEX idx_ticket_reservations_user_id ON ticket_reservations(user_id);
CREATE INDEX idx_ticket_reservations_status_expires_at
    ON ticket_reservations(status, expires_at);