CREATE TABLE ticket_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL,
    nome VARCHAR(175) NOT NULL,
    price NUMERIC(12,2) NOT NULL,
    total_quantity INT NOT NULL,
    available_quantity INT NOT NULL
);

-- Alterar ticket_types para regras de Meia-Entrada (Foco Brasil)
ALTER TABLE ticket_types 
ADD COLUMN is_half_price_eligible BOOLEAN NOT NULL DEFAULT TRUE,
ADD COLUMN max_half_price_quantity INT NOT NULL DEFAULT 0, -- Lei dos 40% de cota de meia
ADD COLUMN sold_half_price_quantity INT NOT NULL DEFAULT 0;