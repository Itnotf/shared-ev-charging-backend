package service

import (
	"errors"
	"shared-charge/models"
	"time"
)

// 获取用户预约列表（可按日期筛选）
func GetReservationsByUser(userID uint, date string) ([]models.Reservation, error) {
	var reservations []models.Reservation
	query := models.DB.Where("user_id = ?", userID)
	if date != "" {
		if len(date) == 10 {
			query = query.Where("date = ?", date)
		} else if len(date) == 7 {
			query = query.Where("to_char(date, 'YYYY-MM') = ?", date)
		}
	}
	err := query.Order("date DESC, created_at DESC").Find(&reservations).Error
	return reservations, err
}

// 创建预约并做业务校验
func CreateReservationWithCheck(userID uint, date time.Time, timeslot, remark string) (models.Reservation, error) {
	// 检查是否有未完成预约
	var ongoing models.Reservation
	err := models.DB.Where("user_id = ? AND status = ? AND date >= ?", userID, "pending", time.Now().Format("2006-01-02")).First(&ongoing).Error
	if err == nil {
		return models.Reservation{}, errors.New("您有未结束的预约，不能重复预约")
	}

	// 检查上一次预约是否未上传充电记录
	var lastReservation models.Reservation
	errLast := models.DB.Where("user_id = ? AND status != ?", userID, "cancelled").Order("date DESC").First(&lastReservation).Error
	if errLast == nil {
		var endTime time.Time
		switch lastReservation.Timeslot {
		case "day":
			endTime = time.Date(lastReservation.Date.Year(), lastReservation.Date.Month(), lastReservation.Date.Day(), 20, 0, 0, 0, time.Local)
		case "night":
			nextDay := lastReservation.Date.AddDate(0, 0, 1)
			endTime = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 8, 0, 0, 0, time.Local)
		}
		if time.Now().After(endTime) {
			var count int64
			models.DB.Model(&models.Record{}).Where("user_id = ? AND reservation_id = ?", userID, lastReservation.ID).Count(&count)
			if count == 0 {
				return models.Reservation{}, errors.New("上一次预约已结束但未上传充电记录，请先上传记录")
			}
		}
	}

	reservation := models.Reservation{
		UserID:   userID,
		Date:     date,
		Timeslot: timeslot,
		Status:   "pending",
		Remark:   remark,
	}
	if err := models.DB.Create(&reservation).Error; err != nil {
		return models.Reservation{}, err
	}
	models.DB.Preload("User").First(&reservation, reservation.ID)
	return reservation, nil
}

// 取消预约
func CancelReservation(id, userID uint) error {
	var reservation models.Reservation
	if err := models.DB.Where("id = ? AND user_id = ?", id, userID).First(&reservation).Error; err != nil {
		return err
	}
	reservation.Status = "cancelled"
	return models.DB.Save(&reservation).Error
}

// 获取当前预约及充电记录状态
func GetCurrentReservationStatus(userID uint) (map[string]interface{}, error) {
	var currentRes *models.Reservation
	var lastRes *models.Reservation
	needUploadRecord := false

	var lastReservation models.Reservation
	err := models.DB.
		Where("user_id = ? AND status = ?", userID, "pending").
		Preload("User").
		Order("date DESC, id DESC").
		First(&lastReservation).Error
	if err == nil {
		var endTime time.Time
		switch lastReservation.Timeslot {
		case "day":
			endTime = time.Date(lastReservation.Date.Year(), lastReservation.Date.Month(), lastReservation.Date.Day(), 20, 0, 0, 0, time.Local)
		case "night":
			nextDay := lastReservation.Date.AddDate(0, 0, 1)
			endTime = time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 8, 0, 0, 0, time.Local)
		}
		if time.Now().Before(endTime) {
			currentRes = &lastReservation
		} else {
			var count int64
			models.DB.Model(&models.Record{}).Where("user_id = ? AND reservation_id = ?", userID, lastReservation.ID).Count(&count)
			if count == 0 {
				needUploadRecord = true
				lastRes = &lastReservation
			}
		}
	}

	data := map[string]interface{}{
		"currentReservation": FormatReservationDate(currentRes),
		"needUploadRecord":   needUploadRecord,
	}
	if needUploadRecord {
		data["lastReservation"] = FormatReservationDate(lastRes)
	}
	return data, nil
}

// FormatReservationDate 格式化预约数据，仅返回日期字符串
func FormatReservationDate(res *models.Reservation) map[string]interface{} {
	if res == nil {
		return nil
	}

	// 使用公共的用户信息格式化函数
	userInfo := res.User.FormatUserInfo()

	return map[string]interface{}{
		"id":          res.ID,
		"user_id":     res.UserID,
		"date":        res.Date.Format("2006-01-02"),
		"timeslot":    res.Timeslot,
		"status":      res.Status,
		"remark":      res.Remark,
		"created_at":  res.CreatedAt,
		"updated_at":  res.UpdatedAt,
		"user_name":   userInfo["user_name"],
		"user_avatar": userInfo["user_avatar"],
	}
}

// 创建预约
func CreateReservation(userID uint, date string, timeslot string) (*models.Reservation, error) {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}

	reservation, err := CreateReservationWithCheck(userID, parsedDate, timeslot, "")
	if err != nil {
		return nil, err
	}

	return &reservation, nil
}

// 删除预约
func DeleteReservation(id, userID uint) error {
	return models.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Reservation{}).Error
}

// 获取当前预约
func GetCurrentReservation(userID uint) (models.Reservation, error) {
	var reservation models.Reservation
	err := models.DB.Where("user_id = ? AND status != ? AND date >= ?", userID, "cancelled", time.Now().Format("2006-01-02")).Order("date ASC").First(&reservation).Error
	return reservation, err
}

// 获取所有未取消的预约（可按日期筛选）
func GetReservations(date string) ([]models.Reservation, error) {
	var reservations []models.Reservation
	query := models.DB.Where("status != ?", "cancelled")
	if date != "" {
		if len(date) == 10 {
			query = query.Where("date = ?", date)
		} else if len(date) == 7 {
			query = query.Where("to_char(date, 'YYYY-MM') = ?", date)
		}
	}
	err := query.Preload("User").Order("date DESC, created_at DESC").Find(&reservations).Error
	return reservations, err
}

// GetReservationsByDate 根据日期获取预约列表
func GetReservationsByDate(date string) ([]models.Reservation, error) {
	return GetReservations(date)
}

// GetCurrentStatus 获取当前状态
func GetCurrentStatus(userID uint) (map[string]interface{}, error) {
	return GetCurrentReservationStatus(userID)
}
