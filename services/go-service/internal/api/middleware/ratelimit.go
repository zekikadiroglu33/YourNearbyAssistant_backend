package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	limiters = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

// getLimiter returns (or lazily creates) a per-IP token-bucket limiter at 100 req/min with a burst of 20.
func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	if l, ok := limiters[ip]; ok {
		return l
	}

	// 100 requests per minute, burst of 20
	l := rate.NewLimiter(rate.Limit(100.0/60.0), 20)
	limiters[ip] = l
	return l
}

// RateLimit returns a middleware that enforces per-IP rate limiting; exceeding the limit yields 429.
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter := getLimiter(c.ClientIP())
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
