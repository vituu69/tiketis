ALTER TABLE ticket_types
    ADD COLUMN IF NOT EXISTS is_half_price_eligible BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS max_half_price_quantity INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS sold_half_price_quantity INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS producer_account_id UUID,
    ADD COLUMN IF NOT EXISTS event_date TIMESTAMP;

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS convenience_fee NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    ADD COLUMN IF NOT EXISTS platform_type VARCHAR(50) NOT NULL DEFAULT 'PRIMARY',
    ADD COLUMN IF NOT EXISTS secondary_seller_account_id UUID,
    ADD COLUMN IF NOT EXISTS net_seller_amount NUMERIC(12,2),
    ADD COLUMN IF NOT EXISTS event_date TIMESTAMP;

CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    balance NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS escrow_settlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    seller_account_id UUID NOT NULL REFERENCES accounts(id),
    amount_to_release NUMERIC(12,2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'HELD',
    release_at TIMESTAMP NOT NULL,
    released_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_escrow_settlements_status_release_at
    ON escrow_settlements(status, release_at);

CREATE INDEX IF NOT EXISTS idx_escrow_settlements_seller_account_id
    ON escrow_settlements(seller_account_id);
