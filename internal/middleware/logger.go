package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func GinZeroLoggerPersonalizado() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// processo da requisição
		c.Next()

		// Estatisticas Pos-requisição
		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		// estrutura do log
		if raw != "" {
			path = path + "?" + raw
		}

		// cria log estruturado
		logger := log.Info()

		if len(c.Errors) > 0 {
			logger = log.Error().Strs("errors", c.Errors.Errors())
		} else if statusCode >= 400 && statusCode < 500 {
			logger = log.Warn()
		} else if statusCode >= 500 {
			logger = log.Error()
		}

		logger.
			Int("status", statusCode).
			Str("method", method).
			Str("path", path).
			Str("ip", clientIP).
			Dur("latency", latency).
			Msg("HTTP Request")
	}
}
