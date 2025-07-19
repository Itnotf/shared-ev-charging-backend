package utils

import (
	"shared-charge/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken 生成JWT令牌
func GenerateToken(user interface{}) (string, error) {
	cfg := config.GetConfig()

	// 类型断言获取用户信息
	userMap, ok := user.(map[string]interface{})
	if !ok {
		return "", jwt.ErrSignatureInvalid
	}

	claims := jwt.MapClaims{
		"user_id": userMap["id"],
		"openid":  userMap["openid"],
		"name":    userMap["name"],
		"role":    userMap["role"],
		"exp":     time.Now().Add(time.Duration(cfg.JWT.ExpireHours) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetConfig().JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
