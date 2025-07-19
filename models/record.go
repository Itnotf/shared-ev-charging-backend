package models

import (
	"math"
	"time"

	"gorm.io/gorm"
)

// Record 充电记录表
type Record struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"not null;comment:用户ID"`
	Date          time.Time      `json:"date" gorm:"type:date;not null;comment:充电日期(无时区)"`
	KWH           float64        `json:"kwh" gorm:"not null;comment:充电度数(kWh)"`
	Amount        int64          `json:"amount" gorm:"not null;comment:费用金额(分)"`
	UnitPrice     float64        `json:"unit_price" gorm:"not null;comment:单价"`
	ImageURL      string         `json:"image_url" gorm:"column:image_url;size:255;comment:电量截图URL"`
	Remark        string         `json:"remark" gorm:"size:255;comment:备注"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index" swaggerignore:"true"`
	ReservationID uint           `json:"reservation_id"`
	Timeslot      string         `json:"timeslot" gorm:"size:20;comment:班次:day,night"`

	// 关联关系
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (Record) TableName() string {
	return "records"
}

// CalculateAmount 计算费用
func (r *Record) CalculateAmount() {
	r.Amount = int64(math.Round(r.KWH * r.UnitPrice * 100)) // 单位为分
}
