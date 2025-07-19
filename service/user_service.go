package service

import (
	"shared-charge/config"
	"shared-charge/models"
)

// 获取用户信息
func GetUserByID(id uint) (models.User, error) {
	var user models.User
	err := models.DB.First(&user, id).Error
	return user, err
}

// 获取用户专属电价（无则返回全局默认）
func GetUserUnitPrice(user models.User) float64 {
	if user.UnitPrice > 0 {
		return user.UnitPrice
	}
	return config.GetConfig().App.DefaultUnitPrice
}

// GetUserPrice 获取用户电价
func GetUserPrice(userID uint) (float64, error) {
	var user models.User
	err := models.DB.First(&user, userID).Error
	if err != nil {
		return 0, err
	}
	return GetUserUnitPrice(user), nil
}

// UpdateUserProfile 更新用户信息（修复函数签名）
func UpdateUserProfile(userID uint, avatar string, nickName string) error {
	return models.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"avatar":   avatar,
		"nickname": nickName,
	}).Error
}
