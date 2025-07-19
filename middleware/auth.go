package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"shared-charge/config"
	"shared-charge/models"
	"strings"
	"sync"
	"time"

	"shared-charge/utils"

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

// getUserFromCache 从 Redis 获取用户缓存
func getUserFromCache(userID uint) (*models.User, bool) {
	key := fmt.Sprintf("user_cache:%d", userID)
	val, err := utils.GetRedis().Get(utils.RedisCtx(), key).Result()
	if err != nil {
		return nil, false
	}
	var user models.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, false
	}
	return &user, true
}

// setUserToCache 设置用户缓存到 Redis
func setUserToCache(user *models.User) {
	key := fmt.Sprintf("user_cache:%d", user.ID)
	data, _ := json.Marshal(user)
	utils.GetRedis().Set(utils.RedisCtx(), key, data, time.Hour) // 1小时过期
}

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
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
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证令牌格式错误",
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// 解析JWT令牌
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.GetConfig().JWT.Secret), nil
		})

		if err != nil || !token.Valid {
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
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证令牌解析失败",
			})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户ID无效",
			})
			c.Abort()
			return
		}

		// 先从缓存获取用户信息
		userIDUint := uint(userID)
		if user, exists := getUserFromCache(userIDUint); exists {
			// 检查用户状态
			if !user.IsActive() {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "用户已被禁用",
				})
				c.Abort()
				return
			}
			c.Set("user", *user)
			c.Next()
			return
		}

		// 缓存未命中，查询数据库
		var user models.User
		if err := models.DB.First(&user, userIDUint).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户不存在",
			})
			c.Abort()
			return
		}

		// 检查用户状态
		if !user.IsActive() {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "用户已被禁用",
			})
			c.Abort()
			return
		}

		// 更新缓存
		setUserToCache(&user)

		// 将用户信息存储到上下文中
		c.Set("user", user)
		c.Next()
	}
}
