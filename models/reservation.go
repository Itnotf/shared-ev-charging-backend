package models

import (
	"time"

	"gorm.io/gorm"
)

// Reservation 预约表
type Reservation struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null;comment:用户ID"`
	Date      time.Time      `json:"date" gorm:"type:date;not null;comment:预约日期(无时区)"`
	Timeslot  string         `json:"timeslot" gorm:"size:20;not null;comment:时段:day,night"`
	Status    string         `json:"status" gorm:"size:20;default:'pending';comment:状态:pending,confirmed,cancelled,completed"`
	Remark    string         `json:"remark" gorm:"size:255;comment:备注"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index" swaggerignore:"true"`

	// 关联关系
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (Reservation) TableName() string {
	return "reservations"
}

// TimeslotText 获取时段文本
func (r *Reservation) TimeslotText() string {
	switch r.Timeslot {
	case "day":
		return "白班 (08:00-20:00)"
	case "night":
		return "夜班 (20:00-08:00)"
	default:
		return r.Timeslot
	}
}

// IsConfirmed 检查是否已确认
func (r *Reservation) IsConfirmed() bool {
	return r.Status == "confirmed"
}

// IsCancelled 检查是否已取消
func (r *Reservation) IsCancelled() bool {
	return r.Status == "cancelled"
}

// IsCompleted 检查是否已完成
func (r *Reservation) IsCompleted() bool {
	return r.Status == "completed"
}

// FormatReservationInfo 格式化预约信息
func (r *Reservation) FormatReservationInfo() map[string]interface{} {
	return map[string]interface{}{
		"id":            r.ID,
		"user_id":       r.UserID,
		"date":          r.Date.Format("2006-01-02"),
		"timeslot":      r.Timeslot,
		"timeslot_text": r.TimeslotText(),
		"status":        r.Status,
		"remark":        r.Remark,
		"created_at":    r.CreatedAt,
		"updated_at":    r.UpdatedAt,
		"user_name":     r.User.Name,
		"user_avatar":   r.User.Avatar,
	}
}
