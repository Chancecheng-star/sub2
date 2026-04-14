package handler

import (
	"net/http"
	"runtime"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查 Handler
type HealthHandler struct {
	cfg       *config.Config
	startTime time.Time
}

// NewHealthHandler 创建健康检查 Handler
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		cfg:       cfg,
		startTime: time.Now(),
	}
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	// Status 状态：healthy/unhealthy/degraded
	Status string `json:"status"`
	// Version 服务版本号
	Version string `json:"version"`
	// Uptime 运行时长（秒）
	Uptime int64 `json:"uptime"`
	// Timestamp 当前时间戳
	Timestamp time.Time `json:"timestamp"`
	// Checks 详细检查项
	Checks HealthChecks `json:"checks"`
}

// HealthChecks 健康检查项
type HealthChecks struct {
	// Memory 内存状态
	Memory *MemoryHealthItem `json:"memory,omitempty"`
	// Goroutine Goroutine 数量
	Goroutine *GoroutineHealthItem `json:"goroutine,omitempty"`
}

// MemoryHealthItem 内存健康项
type MemoryHealthItem struct {
	// AllocMB 已分配内存（MB）
	AllocMB uint64 `json:"alloc_mb"`
	// SysMB 系统内存（MB）
	SysMB uint64 `json:"sys_mb"`
	// UsagePercent 使用率（%）
	UsagePercent float64 `json:"usage_percent"`
	// Status 状态：healthy/warning/critical
	Status string `json:"status"`
}

// GoroutineHealthItem Goroutine 健康项
type GoroutineHealthItem struct {
	// Count 数量
	Count int `json:"count"`
	// Status 状态：healthy/warning/critical
	Status string `json:"status"`
}

// Health 健康检查接口
// GET /healthz
func (h *HealthHandler) Health(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Version:   "0.1.117",
		Uptime:    int64(time.Since(h.startTime).Seconds()),
		Timestamp: time.Now(),
		Checks:    HealthChecks{},
	}

	// 内存检查
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	allocMB := memStats.Alloc / 1024 / 1024
	sysMB := memStats.Sys / 1024 / 1024
	usagePercent := 0.0
	if sysMB > 0 {
		usagePercent = float64(memStats.Alloc) / float64(memStats.Sys) * 100
	}

	memoryStatus := "healthy"
	if usagePercent > 90 {
		memoryStatus = "critical"
		response.Status = "degraded"
	} else if usagePercent > 70 {
		memoryStatus = "warning"
	}

	response.Checks.Memory = &MemoryHealthItem{
		AllocMB:      allocMB,
		SysMB:        sysMB,
		UsagePercent: usagePercent,
		Status:       memoryStatus,
	}

	// Goroutine 检查
	goroutineCount := runtime.NumGoroutine()
	goroutineStatus := "healthy"
	if goroutineCount > 1000 {
		goroutineStatus = "critical"
		response.Status = "degraded"
	} else if goroutineCount > 500 {
		goroutineStatus = "warning"
	}

	response.Checks.Goroutine = &GoroutineHealthItem{
		Count:  goroutineCount,
		Status: goroutineStatus,
	}

	c.JSON(http.StatusOK, response)
}

// Live 存活检查（简化版）
// GET /live
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now(),
	})
}

// Ready 就绪检查
// GET /ready
func (h *HealthHandler) Ready(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}
