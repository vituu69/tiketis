package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vituu69/tiketis/internal/like-redis/infrastructure/domain/repository"
)

func RateLimit(memory repository.KeyValueRepository, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if memory == nil || limit <= 0 || window <= 0 {
			c.Next()
			return
		}

		key := fmt.Sprintf("rate_limit:%s:%s", c.ClientIP(), c.FullPath())
		count, ttl, ok := memory.Incr(c.Request.Context(), key, window)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "rate limit failed"})
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, limit-count)))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", ttl))

		if count > limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}

		c.Next()
	}
}
