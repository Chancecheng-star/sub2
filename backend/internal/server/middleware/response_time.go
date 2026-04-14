package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

func ResponseTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		
		c.Header("X-Response-Time", latency.String())
	}
}
