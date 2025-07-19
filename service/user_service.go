package service

import (
	"shared-charge/config"
	"shared-charge/models"
	"shared-charge/utils"

	"github.com/gin-gonic/gin"
)

// 获取用户信息
func GetUserByID(c *gin.Context, id uint) (models.User, error) {
	utils.InfoCtx(c, "查询用户信息: user_id=%d", id)
	var user models.User
	err := models.DB.First(&user, id).Error
	if err != nil {
		utils.ErrorCtx(c, "查询用户信息失败: %v", err)
	}
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
func UpdateUserProfile(c *gin.Context, userID uint, avatar string, nickName string) error {
	utils.InfoCtx(c, "更新用户信息: user_id=%d", userID)
	err := models.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"avatar": avatar,
		"name":   nickName, // 修正为 name
	}).Error
	if err != nil {
		utils.ErrorCtx(c, "更新用户信息失败: %v", err)
	}
	return err
}

// UpdateUserPhoneByID 更新用户手机号
func UpdateUserPhoneByID(c *gin.Context, userID uint, phone string) error {
	utils.InfoCtx(c, "更新用户手机号: user_id=%d", userID)
	err := models.DB.Model(&models.User{}).Where("id = ?", userID).Update("phone", phone).Error
	if err != nil {
		utils.ErrorCtx(c, "更新用户手机号失败: %v", err)
	}
	return err
}
