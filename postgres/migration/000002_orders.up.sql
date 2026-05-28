CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    total_amount NUMERIC(12,2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Alterar orders para suportar taxas de conveniência (Mercado Primário)
ALTER TABLE orders 
ADD COLUMN convenience_fee NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
ADD COLUMN platform_type VARCHAR(50) NOT NULL DEFAULT 'PRIMARY'; -- PRIMARY, SECONDARY, SELF_SERVICE

-- Alterar ticket_types para regras de Meia-Entrada (Foco Brasil)
ALTER TABLE ticket_types 
ADD COLUMN is_half_price_eligible BOOLEAN NOT NULL DEFAULT TRUE,
ADD COLUMN max_half_price_quantity INT NOT NULL DEFAULT 0, -- Lei dos 40% de cota de meia
ADD COLUMN sold_half_price_quantity INT NOT NULL DEFAULT 0;