package models

import (
	"time"

	"gorm.io/gorm"
)

// LicensePlate 车牌号模型
type LicensePlate struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      uint           `json:"user_id" gorm:"not null;comment:用户ID"`
	PlateNumber string         `json:"plate_number" gorm:"size:20;not null;comment:车牌号"`
	IsDefault   bool           `json:"is_default" gorm:"default:false;comment:是否为默认车牌号"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index" swaggerignore:"true"`

	// 关联关系
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (LicensePlate) TableName() string {
	return "license_plates"
}

// BeforeCreate 创建前钩子
func (lp *LicensePlate) BeforeCreate(tx *gorm.DB) error {
	// 如果设置为默认车牌号，需要将该用户的其他车牌号设为非默认
	if lp.IsDefault {
		return tx.Model(&LicensePlate{}).
			Where("user_id = ? AND deleted_at IS NULL", lp.UserID).
			Update("is_default", false).Error
	}
	return nil
}

// BeforeUpdate 更新前钩子
func (lp *LicensePlate) BeforeUpdate(tx *gorm.DB) error {
	// 如果设置为默认车牌号，需要将该用户的其他车牌号设为非默认
	if lp.IsDefault {
		return tx.Model(&LicensePlate{}).
			Where("user_id = ? AND id != ? AND deleted_at IS NULL", lp.UserID, lp.ID).
			Update("is_default", false).Error
	}
	return nil
}

// FormatLicensePlateInfo 格式化车牌号信息
func (lp *LicensePlate) FormatLicensePlateInfo() map[string]interface{} {
	return map[string]interface{}{
		"id":           lp.ID,
		"plate_number": lp.PlateNumber,
		"is_default":   lp.IsDefault,
		"created_at":   lp.CreatedAt,
	}
}


