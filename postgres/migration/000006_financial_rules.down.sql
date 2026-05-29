DROP INDEX IF EXISTS idx_escrow_settlements_seller_account_id;
DROP INDEX IF EXISTS idx_escrow_settlements_status_release_at;
DROP TABLE IF EXISTS escrow_settlements;
DROP TABLE IF EXISTS accounts;

ALTER TABLE orders
    DROP COLUMN IF EXISTS event_date,
    DROP COLUMN IF EXISTS net_seller_amount,
    DROP COLUMN IF EXISTS secondary_seller_account_id,
    DROP COLUMN IF EXISTS platform_type,
    DROP COLUMN IF EXISTS convenience_fee;

ALTER TABLE ticket_types
    DROP COLUMN IF EXISTS event_date,
    DROP COLUMN IF EXISTS producer_account_id,
    DROP COLUMN IF EXISTS sold_half_price_quantity,
    DROP COLUMN IF EXISTS max_half_price_quantity,
    DROP COLUMN IF EXISTS is_half_price_eligible;
