package service

import (
	"shared-charge/config"
	"shared-charge/models"
)

// 通过openid查找用户
func GetUserByOpenID(openid string) (models.User, error) {
	var user models.User
	err := models.DB.Where("openid = ?", openid).First(&user).Error
	return user, err
}

// 创建新用户
func CreateUser(user *models.User) error {
	return models.DB.Create(user).Error
}

// 新增：用于封装新建用户输入参数
type UserCreateInput struct {
	OpenID string
	Name   string
	Phone  string
	Role   string
	Status string
}

// 新增：根据 UserCreateInput 创建用户
func CreateUserWithInput(input UserCreateInput) (models.User, error) {
	cfg := config.GetConfig()
	user := models.User{
		OpenID:    input.OpenID,
		Name:      input.Name,
		Phone:     input.Phone,
		Role:      input.Role,
		Status:    input.Status,
		UnitPrice: cfg.App.DefaultUnitPrice, // 设置默认电价
	}
	err := models.DB.Create(&user).Error
	return user, err
}
