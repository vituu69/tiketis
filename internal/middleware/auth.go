package middleware

import (
	"errors"
	"os"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

var (
	TokenAuth *jwtauth.JWTAuth
)

// InitTokenAuthFromEnv inicializa a autenticação JWT usando a variável de ambiente JWT_SECRET.
func InitTokenAuthFromEnv() error {
	secret := os.Getenv("JWT_SECRET")
	return InitTokenAuth(secret)
}

// InitTokenAuth inicializa a autenticação JWT com o segredo fornecido.
func InitTokenAuth(secret string) error {
	if secret == "" {
		return errors.New("JWT_SECRET environment variable is required")
	}

	if len(secret) < 32 {
		return errors.New("JWT_SECRET tem que ter 32 caracter")
	}

	TokenAuth = jwtauth.New(
		"HS256", []byte(secret), nil,
	)
	return nil
}

// GenerateToken cria um JWT assinado para o ID de usuário fornecido.
func GenerateToken(userID uuid.UUID) (string, error) {
	if TokenAuth == nil {
		return "", errors.New("token auth is not initialized")
	}

	// Incluir a identidade do usuário e a data de expiração nas declarações JWT assinadas.
	claims := map[string]interface{}{
		"user_id": userID.String(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	_, tokenString, err := TokenAuth.Encode(claims)
	return tokenString, err
}
