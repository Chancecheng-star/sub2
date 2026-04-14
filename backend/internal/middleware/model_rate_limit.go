package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ModelRateLimitConfig struct {
	Enabled      bool
	Limits       map[string]ModelLimit
	DefaultRPS   float64
	DefaultBurst int
}

type ModelLimit struct {
	RPS     float64
	Burst   int
	Enabled bool
}

var DefaultModelRateLimitConfig = ModelRateLimitConfig{
	Enabled:      false,
	DefaultRPS:   10.0,
	DefaultBurst: 20,
	Limits: map[string]ModelLimit{
		"claude-sonnet-4-20250514": {RPS: 20.0, Burst: 40, Enabled: true},
		"claude-opus-4-20250514":   {RPS: 5.0, Burst: 10, Enabled: true},
		"gpt-4":                    {RPS: 10.0, Burst: 20, Enabled: true},
	},
}

type ModelRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	config   ModelRateLimitConfig
}

func NewModelRateLimiter(config ModelRateLimitConfig) *ModelRateLimiter {
	limiter := &ModelRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		config:   config,
	}

	for model, limit := range config.Limits {
		if limit.Enabled {
			limiter.limiters[model] = rate.NewLimiter(rate.Limit(limit.RPS), limit.Burst)
		}
	}

	return limiter
}

func (l *ModelRateLimiter) getLimiter(model string) *rate.Limiter {
	l.mu.RLock()
	if limiter, exists := l.limiters[model]; exists {
		l.mu.RUnlock()
		return limiter
	}
	l.mu.RUnlock()

	l.mu.Lock()
	defer l.mu.Unlock()

	if limiter, exists := l.limiters[model]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(rate.Limit(l.config.DefaultRPS), l.config.DefaultBurst)
	l.limiters[model] = limiter
	return limiter
}

func (l *ModelRateLimiter) Middleware() gin.HandlerFunc {
	if !l.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		model := extractModelFromRequest(c)
		if model == "" {
			model = "default"
		}

		limiter := l.getLimiter(model)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "model rate limit exceeded",
				"model": model,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func extractModelFromRequest(c *gin.Context) string {
	type RequestBody struct {
		Model string `json:"model"`
	}

	var body RequestBody
	if c.Request.Body != nil {
		_ = c.ShouldBindJSON(&body)
	}

	if body.Model != "" {
		return body.Model
	}

	if model := c.Query("model"); model != "" {
		return model
	}

	if model := c.GetHeader("X-Model"); model != "" {
		return model
	}

	return ""
}

func ModelRateLimit(limits map[string]float64, defaultRPS float64) gin.HandlerFunc {
	config := DefaultModelRateLimitConfig
	config.Enabled = true
	config.DefaultRPS = defaultRPS

	if limits != nil {
		config.Limits = make(map[string]ModelLimit)
		for model, rps := range limits {
			config.Limits[model] = ModelLimit{
				RPS:     rps,
				Burst:   int(rps * 2),
				Enabled: true,
			}
		}
	}

	limiter := NewModelRateLimiter(config)
	return limiter.Middleware()
}
