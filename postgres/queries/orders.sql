-- name: CreateOrder :one
-- Cria o pedido inicialmente com o status 'pending' (o total_amount vira string no Go)
INSERT INTO orders (user_id, total_amount, status, Created_at)
VALUES ($1, $2, 'pending', NOW())
RETURNING *;

-- name: GetOrderForUpdate :one
-- Trava o pedido para evitar que um webhook de pagamento processado duas vezes mude o status simultaneamente
SELECT * FROM orders
WHERE id = $1 LIMIT 1
FOR UPDATE;

-- name: UpdateOrderStatus :exec
-- Atualiza o status do pedido (paid, cancelled, refunded)
UPDATE orders
SET status = $1
WHERE id = $2;

-- name: ListOrdersByUser :many
-- Histórico de compras do cliente no app/site
SELECT * FROM orders
WHERE user_id = $1
ORDER BY created_at DESC;
