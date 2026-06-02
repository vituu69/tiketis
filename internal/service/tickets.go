package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vituu69/tiketis/internal/db"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/domain/repository"
	"github.com/vituu69/tiketis/postgres/sqlc"
)

var (
	// ErrTicketStockExhausted é retornado quando o lote geral de ingressos zerou.
	ErrTicketStockExhausted = errors.New("ticket pool or batch is sold out")
	// ErrReservationExpired é retornado quando o usuário tenta pagar por uma reserva que já passou do tempo.
	ErrReservationExpired = errors.New("ticket reservation has expired")
	// ErrHalfPriceQuotaExceeded é retornado quando a cota de 40% reservada para meia-entrada por lei já acabou.
	ErrHalfPriceQuotaExceeded = errors.New("half-price ticket quota for this sector has been exceeded")
	// ErrTicketAlreadySold é retornado se houver tentativa de modificar ou comprar um ingresso com status 'sold'.
	ErrTicketAlreadySold = errors.New("ticket has already been sold")
	// ErrInvalidReservationUser acontece se um usuário tentar pagar ou manipular a reserva de outra pessoa.
	ErrInvalidReservationUser = errors.New("reservation does not belong to this user")
	// ErrTicketTypeLocked evita corrida local enquanto outro processo reserva o mesmo lote/setor.
	ErrTicketTypeLocked = errors.New("ticket type is temporarily locked")
)

type IngressService struct {
	store  *db.Store
	memory repository.KeyValueRepository
}

func NewIngressService(store *db.Store, memory ...repository.KeyValueRepository) *IngressService {
	var mem repository.KeyValueRepository
	if len(memory) > 0 {
		mem = memory[0]
	}

	return &IngressService{
		store:  store,
		memory: mem,
	}
}

type ReserveTicketParams struct {
	UserID       uuid.UUID
	TicketTypeID uuid.UUID
	IsHalfPrice  bool
}

func (s *IngressService) ReserveTicket(ctx context.Context, arg ReserveTicketParams) error {
	const reservationTTL = 5 * time.Minute

	lockOwner := uuid.NewString()
	lockKey := fmt.Sprintf("lock:ticket_type:%s", arg.TicketTypeID)
	if s.memory != nil {
		if ok := s.memory.AcquireLock(ctx, lockKey, lockOwner, 10*time.Second); !ok {
			return ErrTicketTypeLocked
		}
		defer s.memory.ReleaseLock(context.Background(), lockKey, lockOwner)
	}

	var reservation *sqlc.TicketReservation
	err := s.store.ExecTx(ctx, func(q *sqlc.Queries) error {
		ticketType, err := q.GetTicketTypeForUpdate(ctx, arg.TicketTypeID)
		if err != nil {
			return err
		}

		//estoque geral
		if ticketType.AvailableQuantity <= 0 {
			return ErrTicketStockExhausted
		}

		// meia entrada
		if arg.IsHalfPrice {
			if !ticketType.IsHalfPriceEligible {
				return ErrHalfPriceQuotaExceeded
			}

			if ticketType.SoldHalfPriceQuantity >= ticketType.MaxHalfPriceQuantity {
				return ErrHalfPriceQuotaExceeded
			}
		}

		if err := q.ReserveTicketsStock(ctx, 1, arg.TicketTypeID); err != nil {
			return err
		}

		ticket, err := q.BookAvailableTicket(ctx, arg.TicketTypeID)
		if err != nil {
			return err
		}

		// incrementa meia entrada
		if arg.IsHalfPrice {
			err = q.IncrementHalfPriceCount(
				ctx,
				arg.TicketTypeID,
			)

			if err != nil {
				return err
			}
		}

		// cria reserva
		reservation, err = q.CreateReservation(ctx, ticket.ID, arg.UserID, pgtype.Timestamp{
			Time:  time.Now().Add(reservationTTL),
			Valid: true,
		})

		return err

	})
	if err != nil {
		return err
	}

	if s.memory != nil && reservation != nil {
		reservationKey := fmt.Sprintf("reservation:%s", reservation.ID)
		userReservationKey := fmt.Sprintf(
			"user:%s:ticket_type:%s:reservation",
			arg.UserID,
			arg.TicketTypeID,
		)
		value := fmt.Sprintf("%s:%s", reservation.UserID, reservation.TicketID)

		s.memory.SetEX(ctx, reservationKey, value, reservationTTL)
		s.memory.SetEX(ctx, userReservationKey, reservation.ID.String(), reservationTTL)
	}

	return nil

}

// devolve +1 ao estoque caso o timeout do carrinho espire ou cancelamneto
func (s *IngressService) ReturnReservationToStock(ctx context.Context, reservationID uuid.UUID) error {
	return s.store.ExecTx(ctx, func(q *sqlc.Queries) error {
		reservation, err := q.GetReservationForUpdate(ctx, reservationID)
		if err != nil {
			return err
		}

		if reservation.Status != "pending" {
			return nil
		}

		ticket, err := q.GetTicket(ctx, reservation.TicketID)
		if err != nil {
			return err
		}

		if err := q.ReleaseTicketsStock(ctx, 1, ticket.TicketTypeID); err != nil {
			return err
		}

		if err := q.ReturnTicketToAvailable(ctx, ticket.ID); err != nil {
			return err
		}

		err = q.CancelReservation(ctx, reservation.ID)
		if err == nil && s.memory != nil {
			s.memory.Del(ctx, fmt.Sprintf("reservation:%s", reservation.ID))
			s.memory.Del(ctx, fmt.Sprintf(
				"user:%s:ticket_type:%s:reservation",
				reservation.UserID,
				ticket.TicketTypeID,
			))
		}
		return err
	})
}
