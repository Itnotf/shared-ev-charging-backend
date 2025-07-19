package middleware

import (
	"shared-charge/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceMiddleware 性能监控中间件
func PerformanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		if duration > 100*time.Millisecond {
			utils.WarnCtx(c, "慢请求: %s %s, 耗时: %v", c.Request.Method, c.Request.URL.Path, duration)
		}
		utils.DebugCtx(c, "请求性能: %s %s, 耗时: %v, 状态码: %d", c.Request.Method, c.Request.URL.Path, duration, c.Writer.Status())
	}
}
