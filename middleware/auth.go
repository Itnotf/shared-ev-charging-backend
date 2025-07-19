package middleware

import (
	"fmt"
	"net/http"
	"shared-charge/config"
	"shared-charge/models"
	"shared-charge/utils"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// 用户缓存
var userCache = make(map[uint]*models.User)
var cacheMutex sync.RWMutex
var cacheCleanupTicker *time.Ticker

// 初始化缓存清理定时器
func init() {
	cacheCleanupTicker = time.NewTicker(30 * time.Minute) // 每30分钟清理一次缓存
	go func() {
		for range cacheCleanupTicker.C {
			cleanupUserCache()
		}
	}()
}

// 清理过期缓存
func cleanupUserCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// 清理超过1小时的缓存
	cutoff := time.Now().Add(-time.Hour)
	for userID, user := range userCache {
		if user.UpdatedAt.Before(cutoff) {
			delete(userCache, userID)
		}
	}
}

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lg := utils.CtxLogger(c)
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			lg.Warn(fmt.Sprintf("未提供认证令牌: %s", c.Request.URL.Path))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			lg.Warn(fmt.Sprintf("认证令牌格式错误: %s, 格式: %v", c.Request.URL.Path, tokenParts))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证令牌格式错误",
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]
		lg.Debug(fmt.Sprintf("解析令牌: %s", c.Request.URL.Path))

		// 解析JWT令牌
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.GetConfig().JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			lg.Warn(fmt.Sprintf("令牌无效: %s, 错误: %v", c.Request.URL.Path, err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证令牌无效",
			})
			c.Abort()
			return
		}

		// 获取用户信息
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			lg.Error(fmt.Sprintf("令牌解析失败: %s", c.Request.URL.Path))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证令牌解析失败",
			})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			lg.Error(fmt.Sprintf("用户ID无效: %s, claims: %v", c.Request.URL.Path, claims))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户ID无效",
			})
			c.Abort()
			return
		}

		lg.Debug(fmt.Sprintf("用户ID: %.0f, 路径: %s", userID, c.Request.URL.Path))

		// 先从缓存获取用户信息
		userIDUint := uint(userID)
		cacheMutex.RLock()
		if user, exists := userCache[userIDUint]; exists {
			cacheMutex.RUnlock()
			// 检查用户状态
			if !user.IsActive() {
				lg.Warn(fmt.Sprintf("用户已被禁用: ID=%d", user.ID))
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "用户已被禁用",
				})
				c.Abort()
				return
			}
			c.Set("user", *user)
			lg.Info(fmt.Sprintf("认证成功(缓存): 用户ID=%d", user.ID))
			c.Next()
			return
		}
		cacheMutex.RUnlock()

		// 缓存未命中，查询数据库
		var user models.User
		if err := models.DB.First(&user, userIDUint).Error; err != nil {
			lg.Error(fmt.Sprintf("用户不存在: ID=%.0f, 错误: %v", userID, err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户不存在",
			})
			c.Abort()
			return
		}

		// 检查用户状态
		if !user.IsActive() {
			lg.Warn(fmt.Sprintf("用户已被禁用: ID=%d", user.ID))
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "用户已被禁用",
			})
			c.Abort()
			return
		}

		// 更新缓存
		cacheMutex.Lock()
		userCache[userIDUint] = &user
		cacheMutex.Unlock()

		lg.Info(fmt.Sprintf("认证成功(数据库): 用户ID=%d", user.ID))

		// 将用户信息存储到上下文中
		c.Set("user", user)
		c.Next()
	}
}
