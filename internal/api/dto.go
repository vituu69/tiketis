package api

import "time"

// TicketTypeResponse representa um lote/setor de ingressos retornado pela API.
type TicketTypeResponse struct {
	ID                    string    `json:"id"`
	EventID               string    `json:"event_id"`
	Nome                  string    `json:"nome"`
	Price                 string    `json:"price"`
	TotalQuantity         int32     `json:"total_quantity"`
	AvailableQuantity     int32     `json:"available_quantity"`
	IsHalfPriceEligible   bool      `json:"is_half_price_eligible"`
	MaxHalfPriceQuantity  int32     `json:"max_half_price_quantity"`
	SoldHalfPriceQuantity int32     `json:"sold_half_price_quantity"`
	ProducerAccountID     string    `json:"producer_account_id,omitempty"`
	EventDate             time.Time `json:"event_date,omitempty"`
}

// TicketsTypeResponse mantém compatibilidade com o nome antigo.
type TicketsTypeResponse = TicketTypeResponse

// TicketResponse representa um ingresso retornado pela API.
type TicketResponse struct {
	ID           string `json:"id"`
	TicketTypeID string `json:"ticket_type_id"`
	OrderID      string `json:"order_id,omitempty"`
	SeatNumber   string `json:"seat_number,omitempty"`
	Status       string `json:"status"`
	QRCodeToken  string `json:"qr_code_token,omitempty"`
}

// TicketReservationResponse representa uma reserva de ingresso retornada pela API.
type TicketReservationResponse struct {
	ID         string    `json:"id"`
	TicketID   string    `json:"ticket_id"`
	UserID     string    `json:"user_id"`
	ReservedAt time.Time `json:"reserved_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	Status     string    `json:"status"`
}

// OrderResponse representa um pedido retornado pela API.
type OrderResponse struct {
	ID                       string    `json:"id"`
	UserID                   string    `json:"user_id,omitempty"`
	TotalAmount              string    `json:"total_amount"`
	Status                   string    `json:"status"`
	CreatedAt                time.Time `json:"created_at"`
	ConvenienceFee           string    `json:"convenience_fee"`
	PlatformType             string    `json:"platform_type"`
	SecondarySellerAccountID string    `json:"secondary_seller_account_id,omitempty"`
	NetSellerAmount          string    `json:"net_seller_amount,omitempty"`
	EventDate                time.Time `json:"event_date,omitempty"`
}

// AccountResponse representa uma conta financeira retornada pela API.
type AccountResponse struct {
	ID          string    `json:"id"`
	OwnerUserID string    `json:"owner_user_id,omitempty"`
	Balance     string    `json:"balance"`
	CreatedAt   time.Time `json:"created_at"`
}

// EscrowSettlementResponse representa um repasse financeiro retido/liberado.
type EscrowSettlementResponse struct {
	ID              string    `json:"id"`
	OrderID         string    `json:"order_id"`
	SellerAccountID string    `json:"seller_account_id"`
	AmountToRelease string    `json:"amount_to_release"`
	Status          string    `json:"status"`
	ReleaseAt       time.Time `json:"release_at"`
	ReleasedAt      time.Time `json:"released_at,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// RegisterResponse is returned after successful registration.
type RegisterResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token"`
}

// TokenResponse contains a signed JWT.
type TokenResponse struct {
	Token string `json:"token"`
}

// MessageResponse contains a simple status message.
type MessageResponse struct {
	Message string `json:"message"`
}

// ErrorResponse contains an API error message.
type ErrorResponse struct {
	Error string `json:"error"`
}

// ReconcileResponse reports whether stored and computed balances match.
type ReconcileResponse struct {
	Message string `json:"message"`
	Matched bool   `json:"matched"`
}
