package middleware

import (
	"log/slog"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
)

// DebugConfig 调试模式配置
type DebugConfig struct {
	// Enabled 是否启用调试模式
	Enabled bool
	// VerboseHeaders 是否记录详细请求头
	VerboseHeaders bool
	// VerboseBody 是否记录详细请求体
	VerboseBody bool
	// SlowQueryThresholdMs 慢查询阈值（毫秒）
	SlowQueryThresholdMs int
}

// DefaultDebugConfig 默认配置
var DefaultDebugConfig = DebugConfig{
	Enabled:              false,
	VerboseHeaders:       false,
	VerboseBody:          false,
	SlowQueryThresholdMs: 1000,
}

// Debug 调试模式中间件
// 在调试模式下记录详细的请求/响应信息
func Debug() gin.HandlerFunc {
	return DebugWithConfig(DefaultDebugConfig)
}

// DebugWithConfig 自定义配置的调试模式中间件
func DebugWithConfig(config DebugConfig) gin.HandlerFunc {
	if !config.Enabled {
		// 调试模式未启用，直接返回空中间件
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// 记录请求开始
		startTime := time.Now()

		// 记录请求详情
		if config.VerboseHeaders {
			// 记录所有请求头
			for key, values := range c.Request.Header {
				for _, value := range values {
					slog.Debug("request header",
						"key", key,
						"value", value,
						"path", c.Request.URL.Path,
					)
				}
			}
		}

		// 继续处理
		c.Next()

		// 记录响应详情
		latency := time.Since(startTime)

		slog.Debug("request completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", latency.Milliseconds(),
		)

		// 慢请求告警
		if latency.Milliseconds() > int64(config.SlowQueryThresholdMs) {
			slog.Warn("slow request detected",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", c.Writer.Status(),
				"latency_ms", latency.Milliseconds(),
				"threshold_ms", config.SlowQueryThresholdMs,
			)
		}
	}
}

// IsDebugMode 检查是否处于调试模式
func IsDebugMode(cfg *config.Config) bool {
	return cfg.Server.Mode == "debug"
}
