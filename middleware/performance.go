package middleware

import (
	"fmt"
	"shared-charge/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceMiddleware 性能监控中间件
func PerformanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lg := utils.CtxLogger(c)
		start := time.Now()

		// 处理请求
		c.Next()

		duration := time.Since(start)

		// 记录慢请求（超过100ms）
		if duration > 100*time.Millisecond {
			lg.Warn(fmt.Sprintf("慢请求: %s %s, 耗时: %v", c.Request.Method, c.Request.URL.Path, duration))
		}

		// 记录所有请求的性能指标
		lg.Debug(fmt.Sprintf("请求性能: %s %s, 耗时: %v, 状态码: %d",
			c.Request.Method, c.Request.URL.Path, duration, c.Writer.Status()))
	}
}
