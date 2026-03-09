package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RateLimiterConfig holds rate limiter configuration
type RateLimiterConfig struct {
	MaxRequests  int
	WindowTime   time.Duration
	ExcludePaths []string
}

// RateLimiter implements rate limiting
type RateLimiter struct {
	config   *RateLimiterConfig
	requests map[string][]time.Time
	mu       sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimiterConfig) *RateLimiter {
	if config == nil {
		config = &RateLimiterConfig{
			MaxRequests: 100,
			WindowTime:  time.Minute,
		}
	}

	limiter := &RateLimiter{
		config:   config,
		requests: make(map[string][]time.Time),
	}

	// Start cleanup routine
	go limiter.cleanup()

	return limiter
}

// Limit is the rate limiting middleware
func (rl *RateLimiter) Limit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip rate limiting for excluded paths
		for _, path := range rl.config.ExcludePaths {
			if c.Path() == path {
				return c.Next()
			}
		}

		// Get identifier (user ID if authenticated, otherwise IP)
		identifier := rl.getIdentifier(c)

		rl.mu.Lock()
		defer rl.mu.Unlock()

		now := time.Now()
		windowStart := now.Add(-rl.config.WindowTime)

		// Filter requests within window
		var validRequests []time.Time
		for _, reqTime := range rl.requests[identifier] {
			if reqTime.After(windowStart) {
				validRequests = append(validRequests, reqTime)
			}
		}

		// Check if rate limit exceeded
		if len(validRequests) >= rl.config.MaxRequests {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":        "RATE_LIMIT_EXCEEDED",
					"message":     "Too many requests, please try again later",
					"retry_after": rl.config.WindowTime.Seconds(),
				},
			})
		}

		// Add current request
		rl.requests[identifier] = append(validRequests, now)

		return c.Next()
	}
}

// getIdentifier returns a unique identifier for the client
func (rl *RateLimiter) getIdentifier(c *fiber.Ctx) string {
	// Try to get user ID from context
	if userId, err := GetUserIDFromContext(c); err == nil {
		return "user:" + userId.Hex()
	}

	// Fallback to IP address
	return "ip:" + c.IP()
}

// cleanup removes old requests periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-2 * rl.config.WindowTime)

		for key, requests := range rl.requests {
			var valid []time.Time
			for _, req := range requests {
				if req.After(cutoff) {
					valid = append(valid, req)
				}
			}
			if len(valid) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = valid
			}
		}
		rl.mu.Unlock()
	}
}
