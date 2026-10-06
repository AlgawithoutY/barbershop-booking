package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	iplimiters = make(map[string]*rate.Limiter)
	mu         sync.Mutex
)

func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := iplimiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Every(time.Second/5), 10)
		iplimiters[ip] = limiter
	}

	return limiter
}

func RateLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bypass Rate Limiter jika dalam mode Test atau jika header khusus test dikirim
		if gin.Mode() == gin.TestMode || c.GetHeader("X-Test-Mode") == "true" {
			c.Next()
			return
		}

		ip := c.ClientIP()
		limiter := getLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Terlalu banyak permintaan (Rate Limit Exceeded). Silakan coba beberapa saat lagi.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}