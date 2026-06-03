package api

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vituu69/tiketis/postgres/sqlc"
)

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func uuidString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	return uuid.UUID(value.Bytes).String()
}

func textString(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func toAccountResponse(account *sqlc.Account) AccountResponse {
	return AccountResponse{
		ID:          account.ID.String(),
		OwnerUserID: uuidString(account.OwnerUserID),
		Balance:     account.Balance,
		CreatedAt:   account.CreatedAt.Time,
	}
}

func toTicketTypeResponse(ticketType *sqlc.TicketType) TicketTypeResponse {
	return TicketTypeResponse{
		ID:                    ticketType.ID.String(),
		EventID:               ticketType.EventID.String(),
		Nome:                  ticketType.Nome,
		Price:                 ticketType.Price,
		TotalQuantity:         ticketType.TotalQuantity,
		AvailableQuantity:     ticketType.AvailableQuantity,
		IsHalfPriceEligible:   ticketType.IsHalfPriceEligible,
		MaxHalfPriceQuantity:  ticketType.MaxHalfPriceQuantity,
		SoldHalfPriceQuantity: ticketType.SoldHalfPriceQuantity,
		ProducerAccountID:     uuidString(ticketType.ProducerAccountID),
		EventDate:             ticketType.EventDate,
	}
}

func toTicketResponse(ticket *sqlc.Ticket) TicketResponse {
	return TicketResponse{
		ID:           ticket.ID.String(),
		TicketTypeID: ticket.TicketTypeID.String(),
		OrderID:      uuidString(ticket.OrderID),
		SeatNumber:   textString(ticket.SeatNumber),
		Status:       ticket.Status,
		QRCodeToken:  textString(ticket.QrCodeToken),
	}
}

func toTicketReservationResponse(reservation *sqlc.TicketReservation) TicketReservationResponse {
	return TicketReservationResponse{
		ID:         reservation.ID.String(),
		TicketID:   reservation.TicketID.String(),
		UserID:     reservation.UserID.String(),
		ReservedAt: reservation.ReservedAt.Time,
		ExpiresAt:  reservation.ExpiresAt.Time,
		Status:     reservation.Status,
	}
}

func toOrderResponse(order *sqlc.Order) OrderResponse {
	return OrderResponse{
		ID:             order.ID.String(),
		UserID:         uuidString(order.UserID),
		TotalAmount:    order.TotalAmount,
		Status:         order.Status,
		CreatedAt:      order.CreatedAt.Time,
		ConvenienceFee: order.ConvenienceFee,
		PlatformType:   order.PlatformType,
		EventDate:      order.EventDate,
	}
}
