CREATE TABLE ticket_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL,
    nome VARCHAR(175) NOT NULL,
    price NUMERIC(12,2) NOT NULL,
    total_quantity INT NOT NULL,
    available_quantity INT NOT NULL
);