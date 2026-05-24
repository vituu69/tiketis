-- name: CreateReservation :one
INSERT INTO ticket_reservations (
    ticket_id,
    user_id,
    reserved_at,
    expires_at,
    status
) VALUES (
    $1, $2, NOW(), $3, 'pending'
)
RETURNING *;

-- name: GetReservation :one
SELECT * FROM ticket_reservations
WHERE id = $1
LIMIT 1;

-- name: GetReservationForUpdate :one
SELECT * FROM ticket_reservations
WHERE id = $1 
LIMIT 1
FOR UPDATE;

-- name: ListActiveReservationsByUser :many
SELECT * FROM ticket_reservations
WHERE user_id = $1
    AND status = 'pending'
    AND expires_at > NOW()
ORDER BY reserved_at DESC;

-- name: CompleteReservation :exec
-- Chamado quando o gateway de pagamento confirma o sucesso da compra
UPDATE ticket_reservations
SET status = 'completed'
WHERE id = $1;

-- name: CancelReservation :exec
-- Chamado se o usuário desistir explicitamente ou o pagamento falhar direto
UPDATE ticket_reservations
SET status = 'expired'
WHERE id = $1;

-- name: ListExpiredReservations :many
-- Usado pelo worker em Go para descobrir quem perdeu o prazo
SELECT * FROM ticket_reservations
WHERE status = 'pending' 
  AND expires_at < NOW();

-- name ExpirePastReservations :exec
-- Atualiza em lote o status das reservas que passaram do tempo
UPDATE ticket_reservations
SET status = 'expired'
WHERE status = 'pending'
    AND expires_at < NOW();
