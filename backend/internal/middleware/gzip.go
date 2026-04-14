package middleware

import (
	"compress/gzip"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

type GzipConfig struct {
	Level                int
	MinLength            int
	ExcludedContentTypes []string
}

var DefaultGzipConfig = GzipConfig{
	Level:                6,
	MinLength:            1024,
	ExcludedContentTypes: []string{"image/", "video/", "audio/"},
}

var gzipPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

func Gzip() gin.HandlerFunc {
	return GzipWithConfig(DefaultGzipConfig)
}

func GzipWithConfig(config GzipConfig) gin.HandlerFunc {
	if config.Level <= 0 {
		config.Level = gzip.DefaultCompression
	}
	if config.MinLength <= 0 {
		config.MinLength = 1024
	}

	return func(c *gin.Context) {
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		gz := gzipPool.Get().(*gzip.Writer) //nolint:errcheck // sync.Pool.Get() does not return error
		defer gzipPool.Put(gz)

		gz.Reset(c.Writer)
		defer func() {
			_ = gz.Close() // ignore error on close
		}()

		c.Writer = &gzipResponseWriter{
			ResponseWriter: c.Writer,
			Writer:         gz,
			minLength:      config.MinLength,
			excludedTypes:  config.ExcludedContentTypes,
		}

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		c.Next()
	}
}

type gzipResponseWriter struct {
	gin.ResponseWriter
	Writer        *gzip.Writer
	minLength     int
	excludedTypes []string
	written       bool
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.written = true

		contentType := w.Header().Get("Content-Type")
		for _, excluded := range w.excludedTypes {
			if strings.HasPrefix(contentType, excluded) {
				return w.ResponseWriter.Write(b)
			}
		}

		if len(b) < w.minLength {
			return w.ResponseWriter.Write(b)
		}
	}

	return w.Writer.Write(b)
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) WriteString(s string) (int, error) {
	return w.Writer.Write([]byte(s))
}
