package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	OpenID    string         `json:"openid" gorm:"column:openid;uniqueIndex;not null"`
	Name      string         `json:"name" gorm:"not null"`
	Phone     string         `json:"phone"`
	Avatar    string         `json:"avatar"`
	Role      string         `json:"role" gorm:"default:'user'"`
	Status    string         `json:"status" gorm:"default:'active'"`
	UnitPrice float64        `json:"unit_price" gorm:"default:0.7"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index" swaggerignore:"true"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
	return u.Status == "active"
}

// FormatUserInfo 格式化用户信息，保持API一致性
func (u *User) FormatUserInfo() map[string]interface{} {
	return map[string]interface{}{
		"id":          u.ID,
		"openid":      u.OpenID,
		"user_name":   u.Name,
		"phone":       u.Phone,
		"user_avatar": u.Avatar,
		"role":        u.Role,
		"status":      u.Status,
		"created_at":  u.CreatedAt,
		"updated_at":  u.UpdatedAt,
		"deleted_at":  u.DeletedAt,
		"unit_price":  u.UnitPrice,
	}
}
