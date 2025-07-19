package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"shared-charge/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TraceMiddleware 为每个请求生成trace_id并注入logger
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成trace_id
		traceID := generateTraceID()

		// 创建带 trace_id 字段的 logger
		logger := utils.GetLogger().Desugar().With(zap.String("trace_id", traceID)).Sugar()

		// 将 logger 存入 gin.Context
		c.Set("logger", logger)

		// 设置响应头，便于链路排查
		c.Header("X-Trace-ID", traceID)

		c.Next()
	}
}

// generateTraceID 生成唯一的trace_id
func generateTraceID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
