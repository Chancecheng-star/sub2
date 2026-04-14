package middleware

import (
	"bufio"
	"log/slog"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
)

// LogSamplingConfig 日志采样配置
type LogSamplingConfig struct {
	// Enabled 是否启用日志采样
	Enabled bool
	// SampleRate 采样率 (0.0-1.0)，0.1 表示 10% 的请求会被记录详细日志
	SampleRate float64
	// SlowRequestRate 慢请求采样率，慢请求的采样率 (通常更高)
	SlowRequestRate float64
	// SlowThresholdMs 慢请求阈值（毫秒）
	SlowThresholdMs int
}

// DefaultLogSamplingConfig 默认配置
var DefaultLogSamplingConfig = LogSamplingConfig{
	Enabled:         false,
	SampleRate:      0.1,      // 10% 采样率
	SlowRequestRate: 1.0,      // 慢请求 100% 记录
	SlowThresholdMs: 1000,     // 1 秒算慢请求
}

// LogSampling 日志采样中间件
// 在高流量场景下减少日志量，同时保留慢请求和错误请求的详细日志
func LogSampling() gin.HandlerFunc {
	return LogSamplingWithConfig(DefaultLogSamplingConfig)
}

// LogSamplingWithConfig 自定义配置的日志采样中间件
func LogSamplingWithConfig(config LogSamplingConfig) gin.HandlerFunc {
	if !config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		startTime := time.Now()
		
		// 决定是否采样此请求
		shouldSample := shouldSampleRequest(config.SampleRate)
		
		// 记录请求开始（仅采样请求）
		if shouldSample {
			slog.Info("request started",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"sampled", true,
			)
		}

		// 继续处理
		c.Next()

		// 计算延迟
		latency := time.Since(startTime)
		isSlow := latency.Milliseconds() > int64(config.SlowThresholdMs)
		hasError := c.Writer.Status() >= 400

		// 决定是否记录此请求的详细日志
		shouldLog := shouldSample || isSlow || hasError

		if shouldLog {
			// 慢请求或错误请求 100% 记录
			if isSlow || hasError {
				slog.Warn("request completed",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"status", c.Writer.Status(),
					"latency_ms", latency.Milliseconds(),
					"is_slow", isSlow,
					"has_error", hasError,
					"sampled", shouldSample,
				)
			} else {
				// 普通采样请求
				slog.Info("request completed",
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"status", c.Writer.Status(),
					"latency_ms", latency.Milliseconds(),
					"sampled", true,
				)
			}
		}
	}
}

// shouldSampleRequest 决定是否采样此请求
func shouldSampleRequest(rate float64) bool {
	if rate >= 1.0 {
		return true
	}
	if rate <= 0.0 {
		return false
	}
	return rand.Float64() < rate
}

// LogSamplingLevel 日志采样级别
type LogSamplingLevel int

const (
	// LogSamplingOff 关闭采样
	LogSamplingOff LogSamplingLevel = iota
	// LogSamplingLow 低采样率 (1%)
	LogSamplingLow
	// LogSamplingMedium 中等采样率 (10%)
	LogSamplingMedium
	// LogSamplingHigh 高采样率 (50%)
	LogSamplingHigh
	// LogSamplingAll 全部记录 (100%)
	LogSamplingAll
)

// GetSampleRate 获取采样率
func GetSampleRate(level LogSamplingLevel) float64 {
	switch level {
	case LogSamplingOff:
		return 0.0
	case LogSamplingLow:
		return 0.01
	case LogSamplingMedium:
		return 0.1
	case LogSamplingHigh:
		return 0.5
	case LogSamplingAll:
		return 1.0
	default:
		return 0.1
	}
}

// SamplingWriter 采样日志写入器
type SamplingWriter struct {
	writer     *bufio.Writer
	sampleRate float64
}

// NewSamplingWriter 创建采样日志写入器
func NewSamplingWriter(writer *bufio.Writer, sampleRate float64) *SamplingWriter {
	return &SamplingWriter{
		writer:     writer,
		sampleRate: sampleRate,
	}
}

// Write 写入日志（采样）
func (s *SamplingWriter) Write(p []byte) (n int, err error) {
	if shouldSampleRequest(s.sampleRate) {
		return s.writer.Write(p)
	}
	return len(p), nil
}
