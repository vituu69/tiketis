-- name: GetPrimaryTicketTypeForUpdate :one
SELECT
    id,
    event_id,
    nome,
    price,
    total_quantity,
    available_quantity,
    is_half_price_eligible,
    max_half_price_quantity,
    sold_half_price_quantity,
    producer_account_id,
    event_date
FROM ticket_types
WHERE id = $1
LIMIT 1
FOR UPDATE;

-- name: IncrementHalfPriceCount :exec
UPDATE ticket_types
SET sold_half_price_quantity = sold_half_price_quantity + 1
WHERE id = $1
    AND is_half_price_eligible = TRUE
    AND sold_half_price_quantity < max_half_price_quantity;

-- name: CreatePrimaryOrder :one
INSERT INTO orders (
    user_id,
    total_amount,
    convenience_fee,
    platform_type,
    status,
    event_date
) VALUES (
    $1, $2, $3, 'PRIMARY', 'pending', $4
)
RETURNING id, user_id, total_amount, status, created_at;

-- name: CreateEscrowSettlement :one
INSERT INTO escrow_settlements (
    order_id,
    seller_account_id,
    amount_to_release,
    status,
    release_at
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id, order_id, seller_account_id, amount_to_release, status, release_at, released_at, created_at;

-- name: ListMaturedSettlements :many
SELECT id, order_id, seller_account_id, amount_to_release, status, release_at, released_at, created_at
FROM escrow_settlements
WHERE status = 'HELD'
    AND release_at <= $1
ORDER BY release_at ASC;

-- name: UpdateAccountBalance :exec
UPDATE accounts
SET balance = balance + $1
WHERE id = $2;

-- name: UpdateSettlementStatus :exec
UPDATE escrow_settlements
SET status = $1,
    released_at = CASE WHEN $1 = 'RELEASED' THEN NOW() ELSE released_at END
WHERE id = $2;
