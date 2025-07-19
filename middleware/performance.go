package middleware

import (
	"github.com/gin-gonic/gin"
)

// PerformanceMiddleware 性能监控中间件
func PerformanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
