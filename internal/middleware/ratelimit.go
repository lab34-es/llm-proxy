package middleware

import (
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// RateLimiter is an in-memory per-key rate limiter.
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

// getLimiter returns or creates a token-bucket limiter for the given key ID
// with the specified requests-per-minute rate.
func (rl *RateLimiter) getLimiter(keyID string, rpm int) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if l, ok := rl.limiters[keyID]; ok {
		return l
	}

	// rpm requests per minute → rpm/60 per second, burst of rpm (allow short spikes).
	rps := rate.Limit(float64(rpm) / 60.0)
	l := rate.NewLimiter(rps, rpm)
	rl.limiters[keyID] = l
	return l
}

// Middleware returns an Echo middleware that enforces rate limits.
// It must run after ProxyAuth so the API key is present in the context.
func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			k := GetAPIKey(c)
			if k == nil || k.RateLimitRPM <= 0 {
				// No key or unlimited — skip.
				return next(c)
			}

			limiter := rl.getLimiter(k.ID, k.RateLimitRPM)
			if !limiter.Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
			}
			return next(c)
		}
	}
}
