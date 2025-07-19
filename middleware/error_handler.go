package middleware

import (
	"net/http"
	"shared-charge/utils"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 全局错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors[0].Err
			// 记录日志
			utils.ErrorCtx(c, "全局捕获到错误: %v", err)
			// 如果未响应，统一返回
			if !c.IsAborted() && !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "服务器内部错误",
					"error":   err.Error(),
				})
			}
		}
	}
}
