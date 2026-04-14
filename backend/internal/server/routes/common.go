package routes

import (
	"net/http"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/gin-gonic/gin"
)

// RegisterCommonRoutes 注册通用路由（健康检查、状态等）
func RegisterCommonRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 存活检查
	r.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "alive",
			"timestamp": time.Now(),
		})
	})

	// Claude Code 遥测日志（忽略，直接返回 200）
	r.POST("/api/event_logging/batch", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Setup status endpoint
	r.GET("/setup/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"needs_setup": false,
				"step":        "completed",
			},
		})
	})
}

// RegisterHealthRoutes 注册健康检查路由（需要 HealthHandler 实例）
func RegisterHealthRoutes(r *gin.Engine, h *handler.HealthHandler) {
	// 完整健康检查
	r.GET("/healthz", h.Health)
	
	// 就绪检查
	r.GET("/ready", h.Ready)
}
