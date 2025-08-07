package middleware

import (
	"net/http"
	"shared-charge/utils"

	"github.com/gin-gonic/gin"
)

// AdminRequired 仅允许管理员访问
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := utils.GetUserFromContext(c)
		if !ok {
			c.Abort()
			return
		}
		if user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "无权限，仅管理员可操作",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
