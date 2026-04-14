package middleware

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
)

// PerformanceMonitorConfig 性能监控配置
type PerformanceMonitorConfig struct {
	Enabled              bool
	SlowRequestThreshold int // 慢请求阈值（毫秒）
	CPUThreshold         int // CPU 使用率阈值（%）
	MemoryThreshold      int // 内存使用率阈值（%）
	DiskThreshold        int // 磁盘使用率阈值（%）
	EnablePoolStats      bool // 启用连接池统计
	EnableMemoryStats    bool // 启用内存统计
}

var performanceMonitorConfig PerformanceMonitorConfig

// InitPerformanceMonitor 初始化性能监控（从配置加载）
func InitPerformanceMonitor(cfg *config.Config) {
	performanceMonitorConfig = PerformanceMonitorConfig{
		Enabled:              true, // 默认启用
		SlowRequestThreshold: cfg.Server.SlowRequestThreshold,
		CPUThreshold:         90,
		MemoryThreshold:      90,
		DiskThreshold:        90,
		EnablePoolStats:      false,
		EnableMemoryStats:    true,
	}
}

// GetPerformanceMonitorConfig 获取性能监控配置
func GetPerformanceMonitorConfig() PerformanceMonitorConfig {
	return performanceMonitorConfig
}

// SystemStatus 系统状态
type SystemStatus struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
}

// GetSystemStatus 获取系统状态（简化版本）
func GetSystemStatus() SystemStatus {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// 简化版本：只返回内存使用率
	// 完整版本需要使用 github.com/shirou/gopsutil
	return SystemStatus{
		CPUUsage:    0, // 需要 gopsupport
		MemoryUsage: float64(m.Alloc) / float64(m.Sys) * 100,
		DiskUsage:   0, // 需要 gopsupport
	}
}

// PerformanceMonitor 性能监控中间件
func PerformanceMonitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !performanceMonitorConfig.Enabled {
			c.Next()
			return
		}

		// 记录请求开始时间
		startTime := time.Now()
		
		// 执行请求
		c.Next()
		
		// 计算请求耗时
		latency := time.Since(startTime)
		latencyMs := latency.Milliseconds()
		
		// 检查是否为慢请求
		if latencyMs > int64(performanceMonitorConfig.SlowRequestThreshold) {
			// 记录慢请求日志
			slog := c.Request.URL.Path
			if c.Request.URL.RawQuery != "" {
				slog += "?" + c.Request.URL.RawQuery
			}
			fmt.Printf("[SLOW REQUEST] %s %s %dms (threshold: %dms)\n",
				c.Request.Method, slog, latencyMs, performanceMonitorConfig.SlowRequestThreshold)
		}
		
		// 记录请求指标（可选：集成 Prometheus）
		if performanceMonitorConfig.EnableMemoryStats {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			// 可以在此处导出指标
			_ = m
		}
	}
}

// SystemPerformanceCheck 系统性能检查中间件（用于关键接口）
func SystemPerformanceCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !performanceMonitorConfig.Enabled {
			c.Next()
			return
		}

		// 检查系统性能
		status := GetSystemStatus()
		
		// 检查内存
		if performanceMonitorConfig.MemoryThreshold > 0 && status.MemoryUsage > float64(performanceMonitorConfig.MemoryThreshold) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": fmt.Sprintf("system memory overloaded (current: %.1f%%, threshold: %d%%)", 
					status.MemoryUsage, performanceMonitorConfig.MemoryThreshold),
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}
