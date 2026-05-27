CREATE TABLE tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_type_id UUID NOT NULL REFERENCES ticket_types(id),
    order_id UUID REFERENCES orders(id) ON DELETE SET NULL,
    seat_number VARCHAR(50),
    status VARCHAR(50) NOT NULL DEFAULT 'available',
    qr_code_token VARCHAR(255) UNIQUE
);

CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_ticket_type_id ON tickets(ticket_type_id);