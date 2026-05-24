-- name: GetTicketType :one
SELECT * FROM ticket_types
WHERE id = $1
LIMIT 1;

-- name: GetTicketTypeForUpdate :one
-- Trava o lote para atualizar a quantidade disponível com segurança financeira
SELECT * FROM ticket_types
WHERE id = $1
LIMIT 1
FOR UPDATE;

-- name: ListTicketTypesByEvent :many
-- Lista todos os setores/lotes de um evento para mostrar na página de vendas
SELECT * FROM ticket_types
WHERE event_id = $1
ORDER BY price ASC;

-- name: ReserveTicketsStock :exec
-- Deduz a quantidade disponível quando uma reserva é criada
UPDATE ticket_types
SET available_quantity = available_quantity - $1
WHERE id = $2 AND available_quantity >= $1;

-- name: ReleaseTicketsStock :exec
-- Devolve os ingressos ao estoque caso a reserva expire ou seja cancelada
UPDATE ticket_types
SET available_quantity = available_quantity + $1
WHERE id = $2;

-- name: UpdateTicketTypePrice :one
-- Caso precise virar o lote ou alterar o valor (o preço vira string no Go)
UPDATE ticket_types
SET price = $1
WHERE id = $2
RETURNING *;