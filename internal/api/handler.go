package api

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/vituu69/tiketis/internal/db"
	"github.com/vituu69/tiketis/internal/middleware"
	"github.com/vituu69/tiketis/internal/service"
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
		g.JSON(http.StatusConflict, "user already existis or failed")
		return
	}

	g.JSON(http.StatusCreated, gin.H{
		"user_id": user.ID.String(),
		"email":   user.Email,
	})
}
