package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"github.com/vituu69/tiketis/internal/db"
	"github.com/vituu69/tiketis/internal/middleware"
	"github.com/vituu69/tiketis/internal/service"
	"github.com/vituu69/tiketis/postgres/sqlc"
)

type HandlerRequest struct {
	ingress *service.IngressService
	store   *db.Store
}

func NewHandler(ingress *service.IngressService, store *db.Store) *HandlerRequest {
	return &HandlerRequest{ingress: ingress, store: store}
}

func SetupRouter(hctx HandlerRequest) *gin.Engine {
	// Cria um engine do Gin em branco (sem os middlewares padrão Logger e Recovery)
	// Já que você fez um Logger próprio, usar o gin.New() evita logs duplicados.
	r := gin.New()

	// 1. Injeta o middleware de Recovery padrão (evita que a API caia se der um panic)
	r.Use(gin.Recovery())

	// 2. Injeta o middleware personalizado de log globalmente
	r.Use(middleware.GinZeroLoggerPersonalizado())

	// 3. Definição das suas rotas
	r.POST("/register", hctx.Register)
	r.POST("/login", hctx.Login)
	r.POST("/accounts", hctx.CreateAccount)
	r.GET("/accounts", hctx.ListAccounts)
	r.GET("/accounts/:id", hctx.GetAccount)
	r.GET("/events/:event_id/ticket-types", hctx.ListTicketTypesByEvent)
	r.GET("/ticket-types/:ticket_type_id/tickets", hctx.ListAvailableTickets)
	r.POST("/ticket-types/:ticket_type_id/reservations", hctx.ReserveTicketByType)
	r.POST("/ticket-types/:ticket_type_id/buy", hctx.ReserveTicketByType)
	r.GET("/me/reservations", hctx.ListMyReservations)
	r.POST("/reservations/:reservation_id/pay", hctx.PayReservation)
	r.GET("/me/orders", hctx.ListPurchaseHistory)

	return r
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user with email and hashed password, returns user details and JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body    body      object{email=string,password=string}  true  "User registration details"
// @Success      201     {object}  RegisterResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      409     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /register [post]

func (hctx *HandlerRequest) Register(g *gin.Context) {
	//step 1: Decoder registration payload.
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := g.ShouldBindJSON(&input); err != nil {
		log.Warn().Err(err).Msg("Failed to decode register request")
		g.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if input.Email == "" || input.Password == "" {
		g.JSON(http.StatusBadRequest, gin.H{"error": "failed to hash password"})
		return
	}

	// step 3: Hash password before persisting user credential
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Warn().Err(err).Msg("failed to hash password")
		g.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// step 3: persist user record and then mint JWT for immediate login
	user, err := hctx.store.CreateUser(g.Request.Context(), input.Email, string(hashed))

	if err != nil {
		log.Error().Err(err).Str("email", input.Email).Msg("Failed to create user")
		g.JSON(http.StatusConflict, "user already exists or failed")
		return
	}

	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		log.Error().Err(err).Str("email", user.ID.String()).Msg("falhor na geração de token")
		g.JSON(http.StatusInternalServerError, gin.H{"error": "falhou na geração de erro"})
		return
	}

	log.Info().Str("user_id", user.ID.String()).Str("email", user.Email).Msg("User registraddo com ssucesso")
	g.JSON(http.StatusCreated, RegisterResponse{
		UserID: user.ID.String(),
		Email:  user.Email,
		Token:  token,
	})
}

// Login godoc
// @Summary      Login user
// @Description  Authenticates user with email/password and returns JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body    body      object{email=string,password=string}  true  "User login details"
// @Success      200     {object}  TokenResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      401     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /login [post]

func (hctx *HandlerRequest) Login(g *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := g.ShouldBindJSON(&input); err != nil {
		log.Warn().Err(err).Msg("Failed to decoder login request")
		g.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	user, err := hctx.store.GetUserByEmail(g.Request.Context(), input.Email)
	if err != nil {
		log.Warn().Err(err).Str("email", input.Email).Msg("Falha no logion - usario não encontrado")
		g.JSON(http.StatusUnauthorized, "invalid credentiais")
		return
	}

	// compara o hash da senha
	if compareErr := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(input.Password)); compareErr != nil {
		g.JSON(http.StatusUnauthorized, "invalid credential")
		return
	}

	// Retorna um refresh JWT para sucesso automatico
	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg(("Falhou em gerar o token"))
		g.JSON(http.StatusInternalServerError, "Falhou em gerar o token")
		return
	}

	log.Info().Str("user_id", user.ID.String()).Str("email", user.Email).Msg("Usuario logado com sucesso")
	g.JSON(http.StatusOK, TokenResponse{Token: token})
}

// CreateAccount godoc
// @Summary      Create a new account
// @Description  Creates a new user-owned financial account
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Success      201     {object}  AccountResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      401     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /accounts [post]
// @Security     Bearer

func (hctx *HandlerRequest) CreateAccount(g *gin.Context) {
	userID, ok := authenticatedUserID(g)
	if !ok {
		return
	}

	account, err := hctx.store.CreateAccount(g.Request.Context(), pgtype.UUID{
		Bytes: userID,
		Valid: true,
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID.String()).
			Msg("falhou na criação da conta")

		g.JSON(http.StatusInternalServerError, gin.H{"error": "falhou na criação da conta"})
		return
	}

	g.JSON(http.StatusCreated, toAccountResponse(account))
}

func (h *HandlerRequest) ListAccounts(g *gin.Context) {
	userID, ok := authenticatedUserID(g)
	if !ok {
		return
	}

	accounts, err := h.store.ListAccountsByOwner(g.Request.Context(), pgUUID(userID))
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to list accounts")
		g.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list accounts"})
		return
	}

	response := make([]AccountResponse, len(accounts))
	for i, acc := range accounts {
		response[i] = toAccountResponse(acc)
	}

	g.JSON(http.StatusOK, response)
}

func (h *HandlerRequest) GetAccount(g *gin.Context) {
	userID, ok := authenticatedUserID(g)
	if !ok {
		return
	}

	accountID, err := uuid.Parse(g.Param("id"))
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	acc, err := h.store.GetAccount(g.Request.Context(), accountID)
	if err != nil {
		log.Warn().Err(err).Str("account_id", accountID.String()).Msg("Account not found")
		g.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	if acc.OwnerUserID.Valid && uuid.UUID(acc.OwnerUserID.Bytes) != userID {
		log.Warn().
			Str("account_id", accountID.String()).
			Str("user_id", userID.String()).
			Str("owner_id", uuid.UUID(acc.OwnerUserID.Bytes).String()).
			Msg("Access denied to account")

		g.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	g.JSON(http.StatusOK, toAccountResponse(acc))
}

func (h *HandlerRequest) ListTicketTypesByEvent(g *gin.Context) {
	eventID, err := uuid.Parse(g.Param("event_id"))
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	ticketTypes, err := h.store.ListTicketTypesByEvent(g.Request.Context(), eventID)
	if err != nil {
		log.Error().Err(err).Str("event_id", eventID.String()).Msg("failed to list ticket types")
		g.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list ticket types"})
		return
	}

	response := make([]TicketTypeResponse, len(ticketTypes))
	for i, ticketType := range ticketTypes {
		response[i] = toTicketTypeResponse(ticketType)
	}

	g.JSON(http.StatusOK, response)
}

func (h *HandlerRequest) ListAvailableTickets(g *gin.Context) {
	ticketTypeID, err := uuid.Parse(g.Param("ticket_type_id"))
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket type ID"})
		return
	}

	tickets, err := h.store.ListAvailableTickets(g.Request.Context(), ticketTypeID)
	if err != nil {
		log.Error().Err(err).Str("ticket_type_id", ticketTypeID.String()).Msg("failed to list available tickets")
		g.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list available tickets"})
		return
	}

	response := make([]TicketResponse, len(tickets))
	for i, ticket := range tickets {
		response[i] = toTicketResponse(ticket)
	}

	g.JSON(http.StatusOK, response)
}

func (h *HandlerRequest) ReserveTicketByType(g *gin.Context) {
	userID, ok := authenticatedUserID(g)
	if !ok {
		return
	}

	ticketTypeID, err := uuid.Parse(g.Param("ticket_type_id"))
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket type ID"})
		return
	}

	var input struct {
		IsHalfPrice bool `json:"is_half_price"`
	}
	if err := g.ShouldBindJSON(&input); err != nil && err.Error() != "EOF" {
		g.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if err := h.ingress.ReserveTicket(g.Request.Context(), service.ReserveTicketParams{
		UserID:       userID,
		TicketTypeID: ticketTypeID,
		IsHalfPrice:  input.IsHalfPrice,
	}); err != nil {
		switch {
		case errors.Is(err, service.ErrTicketStockExhausted):
			g.JSON(http.StatusConflict, gin.H{"error": "ticket type sold out"})
		case errors.Is(err, service.ErrHalfPriceQuotaExceeded):
			g.JSON(http.StatusConflict, gin.H{"error": "half-price quota exceeded"})
		case errors.Is(err, service.ErrTicketTypeLocked):
			g.JSON(http.StatusTooManyRequests, gin.H{"error": "ticket type is temporarily locked"})
		default:
			log.Error().Err(err).Str("user_id", userID.String()).Str("ticket_type_id", ticketTypeID.String()).Msg("failed to reserve ticket")
			g.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reserve ticket"})
		}
		return
	}

	reservations, err := h.store.ListActiveReservationsByUser(g.Request.Context(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to list user reservations")
		g.JSON(http.StatusCreated, gin.H{"message": "ticket reserved"})
		return
	}

	for _, reservation := range reservations {
		ticket, err := h.store.GetTicket(g.Request.Context(), reservation.TicketID)
		if err == nil && ticket.TicketTypeID == ticketTypeID {
			g.JSON(http.StatusCreated, toTicketReservationResponse(reservation))
			return
		}
	}

	g.JSON(http.StatusCreated, gin.H{"message": "ticket reserved"})
}

func (h *HandlerRequest) ListMyReservations(g *gin.Context) {
	userID, ok := authenticatedUserID(g)
	if !ok {
		return
	}

	reservations, err := h.store.ListActiveReservationsByUser(g.Request.Context(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to list reservations")
		g.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reservations"})
		return
	}

	response := make([]TicketReservationResponse, len(reservations))
	for i, reservation := range reservations {
		response[i] = toTicketReservationResponse(reservation)
	}

	g.JSON(http.StatusOK, response)
}

func (h *HandlerRequest) PayReservation(g *gin.Context) {
	userID, ok := authenticatedUserID(g)
	if !ok {
		return
	}

	reservationID, err := uuid.Parse(g.Param("reservation_id"))
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": "invalid reservation ID"})
		return
	}

	var order *sqlc.Order
	var soldTicket *sqlc.Ticket

	err = h.store.ExecTx(g.Request.Context(), func(q *sqlc.Queries) error {
		reservation, err := q.GetReservationForUpdate(g.Request.Context(), reservationID)
		if err != nil {
			return err
		}

		if reservation.UserID != userID {
			return service.ErrInvalidReservationUser
		}
		if reservation.Status != "pending" {
			return service.ErrTicketAlreadySold
		}
		if reservation.ExpiresAt.Valid && time.Now().After(reservation.ExpiresAt.Time) {
			return service.ErrReservationExpired
		}

		ticket, err := q.GetTicket(g.Request.Context(), reservation.TicketID)
		if err != nil {
			return err
		}

		ticketType, err := q.GetTicketType(g.Request.Context(), ticket.TicketTypeID)
		if err != nil {
			return err
		}

		order, err = q.CreateOrder(g.Request.Context(), pgUUID(userID), ticketType.Price)
		if err != nil {
			return err
		}

		qrCodeToken := uuid.NewString()
		if err := q.AttachTicketToOrder(g.Request.Context(), ticket.ID, pgUUID(order.ID), pgtype.Text{
			String: qrCodeToken,
			Valid:  true,
		}); err != nil {
			return err
		}

		if err := q.CompleteReservation(g.Request.Context(), reservation.ID); err != nil {
			return err
		}

		if err := q.UpdateOrderStatus(g.Request.Context(), "paid", order.ID); err != nil {
			return err
		}

		order.Status = "paid"
		ticket.Status = "sold"
		ticket.OrderID = pgUUID(order.ID)
		ticket.QrCodeToken = pgtype.Text{String: qrCodeToken, Valid: true}
		soldTicket = ticket

		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidReservationUser):
			g.JSON(http.StatusForbidden, gin.H{"error": "reservation does not belong to this user"})
		case errors.Is(err, service.ErrReservationExpired):
			_ = h.ingress.ReturnReservationToStock(g.Request.Context(), reservationID)
			g.JSON(http.StatusConflict, gin.H{"error": "reservation expired"})
		case errors.Is(err, service.ErrTicketAlreadySold):
			g.JSON(http.StatusConflict, gin.H{"error": "reservation is not pending"})
		default:
			log.Error().Err(err).Str("reservation_id", reservationID.String()).Msg("failed to pay reservation")
			g.JSON(http.StatusInternalServerError, gin.H{"error": "failed to pay reservation"})
		}
		return
	}

	g.JSON(http.StatusOK, gin.H{
		"order":  toOrderResponse(order),
		"ticket": toTicketResponse(soldTicket),
	})
}

func (h *HandlerRequest) ListPurchaseHistory(g *gin.Context) {
	userID, ok := authenticatedUserID(g)
	if !ok {
		return
	}

	orders, err := h.store.ListOrdersByUser(g.Request.Context(), pgUUID(userID))
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to list purchase history")
		g.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list purchase history"})
		return
	}

	response := make([]OrderResponse, len(orders))
	for i, order := range orders {
		response[i] = toOrderResponse(order)
	}

	g.JSON(http.StatusOK, response)
}

func authenticatedUserID(g *gin.Context) (uuid.UUID, bool) {
	if middleware.TokenAuth == nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": "token auth is not initialized"})
		return uuid.Nil, false
	}

	tokenString := strings.TrimPrefix(g.GetHeader("Authorization"), "Bearer ")
	if tokenString == "" || tokenString == g.GetHeader("Authorization") {
		g.JSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
		return uuid.Nil, false
	}

	token, err := middleware.TokenAuth.Decode(tokenString)
	if err != nil {
		log.Warn().Err(err).Msg("invalid JWT")
		g.JSON(http.StatusUnauthorized, gin.H{"error": "token invalido"})
		return uuid.Nil, false
	}

	if exp, ok := token.Expiration(); ok && time.Now().After(exp) {
		g.JSON(http.StatusUnauthorized, gin.H{"error": "token expirado"})
		return uuid.Nil, false
	}

	var userIDStr string
	if err := token.Get("user_id", &userIDStr); err != nil || userIDStr == "" {
		log.Warn().Err(err).Msg("user_id claim missing or invalid in JWT")
		g.JSON(http.StatusUnauthorized, gin.H{"error": "token invalido"})
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userIDStr).Msg("invalid user_id in JWT")
		g.JSON(http.StatusUnauthorized, gin.H{"error": "token invalido"})
		return uuid.Nil, false
	}

	return userID, true
}
