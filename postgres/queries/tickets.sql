-- name: GetTicket :one
SELECT * FROM tickets
WHERE id = $1 LIMIT 1;

-- name: ListAvailableTickets :many
-- Lista os ingressos livres para o usuário escolher na tela (sem travar o banco)
SELECT t.* FROM tickets t
LEFT JOIN ticket_reservations tr ON t.id = tr.ticket_id
    AND tr.status = 'pending'
    AND tr.expires_at > NOW()
WHERE t.status = 'available'
    AND t.ticket_type_id = $1
    AND tr.id IS NULL;

-- name: BookAvailableTicket :one
-- Trava e reserva o PRIMEIRO ingresso disponível do lote selecionado.
-- O SKIP LOCKED evita que duas requisições fiquem presas tentando pegar o mesmo ingresso.
UPDATE tickets
SET status = 'reserved'
WHERE id = (
    SELECT t.id
    FROM tickets t
    LEFT JOIN ticket_reservations tr ON t.id = tr.ticket_id
        AND tr.status = 'pending'
        AND tr.expires_at > NOW()
    WHERE t.status = 'available'
        AND t.ticket_type_id = $1
        AND tr.id IS NULL LIMIT 1
        FOR UPDATE SKIP LOCKED
)
RETURNING *;

-- name: AttachTicketToOrder :exec
-- Vincula o ingresso ao pedido final e gera o token do QR Code (pós-pagamento)
UPDATE tickets
SET status = 'available',
    order_id = NULL
WHERE id = $1;

-- name: ListTicketsByOrder :many
-- Lista todos os ingressos de um pedido específico (para gerar os PDFs/QR Codes)
SELECT * FROM tickets
WHERE order_id = $1;

