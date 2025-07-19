package utils

import (
	"net/http"
	"shared-charge/models"
	"time"

	"github.com/gin-gonic/gin"
)

// GetUserFromContext 统一用户认证和类型断言
func GetUserFromContext(c *gin.Context) (models.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return models.User{}, false
	}
	userModel, ok := user.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户信息类型错误"})
		return models.User{}, false
	}
	return userModel, true
}

// ParseDate 统一日期格式校验，layout为"2006-01-02"
func ParseDate(dateStr string) (date time.Time, err error) {
	return time.Parse("2006-01-02", dateStr)
}
