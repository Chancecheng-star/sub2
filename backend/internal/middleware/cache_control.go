package middleware

import (
	"github.com/gin-gonic/gin"
)

// CacheControlConfig 缓存控制配置
type CacheControlConfig struct {
	// NoCache 禁止缓存
	NoCache bool
	// NoStore 禁止存储
	NoStore bool
	// MaxAge 最大缓存时间（秒）
	MaxAge int
	// Private 私有缓存（仅浏览器）
	Private bool
	// Public 公共缓存（CDN 可缓存）
	Public bool
}

// DefaultCacheControlConfig 默认配置（禁止缓存敏感数据）
var DefaultCacheControlConfig = CacheControlConfig{
	NoCache: true,
	NoStore: true,
	MaxAge:  0,
	Private: true,
}

// CacheControl 缓存控制中间件
// 防止敏感数据被缓存
func CacheControl() gin.HandlerFunc {
	return CacheControlWithConfig(DefaultCacheControlConfig)
}

// CacheControlWithConfig 自定义配置的缓存控制中间件
func CacheControlWithConfig(config CacheControlConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 构建 Cache-Control 头
		cacheControl := buildCacheControl(config)
		if cacheControl != "" {
			c.Header("Cache-Control", cacheControl)
		}

		// 禁止缓存敏感数据
		if config.NoStore {
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
	}
}

// buildCacheControl 构建 Cache-Control 头
func buildCacheControl(config CacheControlConfig) string {
	directives := []string{}

	if config.NoCache {
		directives = append(directives, "no-cache")
	}
	if config.NoStore {
		directives = append(directives, "no-store")
	}
	if config.MaxAge > 0 {
		directives = append(directives, "max-age="+itoa(config.MaxAge))
	}
	if config.Private {
		directives = append(directives, "private")
	}
	if config.Public {
		directives = append(directives, "public")
	}

	return join(directives, ", ")
}

// 简单实现 itoa
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	negative := false
	if n < 0 {
		negative = true
		n = -n
	}

	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}

// 简单实现 join
func join(elems []string, sep string) string {
	if len(elems) == 0 {
		return ""
	}
	if len(elems) == 1 {
		return elems[0]
	}

	totalLen := len(elems) - 1
	for _, elem := range elems {
		totalLen += len(elem)
	}

	b := make([]byte, 0, totalLen+len(sep))
	b = append(b, elems[0]...)
	for _, elem := range elems[1:] {
		b = append(b, sep...)
		b = append(b, elem...)
	}

	return string(b)
}

// NoCache 快速设置禁止缓存
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		c.Header("Cache-Control", "no-cache, no-store, max-age=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
	}
}

// CacheFor 设置缓存时间（秒）
func CacheFor(seconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		c.Header("Cache-Control", "public, max-age="+itoa(seconds))
	}
}
