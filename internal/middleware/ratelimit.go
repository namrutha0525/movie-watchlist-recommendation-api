package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/namru/movie-recommend/pkg/response"
)

// RateLimiter implements a simple in-memory token bucket rate limiter per IP.
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
}

type visitor struct {
	tokens    int
	lastSeen  time.Time
}

// NewRateLimiter creates a rate limiter allowing `rate` requests per `window`.
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}

	// Cleanup stale visitors every minute
	go func() {
		for {
			time.Sleep(time.Minute)
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	for ip, v := range rl.visitors {
		if time.Since(v.lastSeen) > rl.window*2 {
			delete(rl.visitors, ip)
		}
	}
}

func (rl *RateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{tokens: rl.rate, lastSeen: time.Now()}
		rl.visitors[ip] = v
		return v
	}

	// Replenish tokens based on elapsed time
	elapsed := time.Since(v.lastSeen)
	if elapsed >= rl.window {
		v.tokens = rl.rate
	}
	v.lastSeen = time.Now()
	return v
}

// Middleware returns the Gin middleware function.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		v := rl.getVisitor(c.ClientIP())

		if v.tokens <= 0 {
			c.Header("Retry-After", rl.window.String())
			c.JSON(http.StatusTooManyRequests, response.APIResponse{
				Success: false,
				Error:   "rate limit exceeded, please try again later",
			})
			c.Abort()
			return
		}

		rl.mu.Lock()
		v.tokens--
		rl.mu.Unlock()

		c.Next()
	}
}
