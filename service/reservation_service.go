package service

import (
	"errors"
	"shared-charge/models"
	"shared-charge/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// 获取用户预约列表（可按日期筛选）
func GetReservationsByUser(c *gin.Context, userID uint, date string) ([]models.Reservation, error) {
	utils.InfoCtx(c, "查询用户预约列表: user_id=%d, date=%s", userID, date)
	var reservations []models.Reservation
	query := models.DB.Where("user_id = ?", userID)
	if date != "" {
		if len(date) == 10 {
			query = query.Where("date = ?", date)
		} else if len(date) == 7 {
			query = query.Where("to_char(date, 'YYYY-MM') = ?", date)
		}
	}
	err := query.Preload("User").Preload("LicensePlate").Order("date DESC, created_at DESC").Find(&reservations).Error
	if err != nil {
		utils.ErrorCtx(c, "查询用户预约列表失败: %v", err)
	}
	return reservations, err
}

// 创建预约并做业务校验
func CreateReservationWithCheck(c *gin.Context, userID uint, date time.Time, timeslot, remark string, licensePlateID *uint) (models.Reservation, error) {
	utils.InfoCtx(c, "创建预约业务校验: user_id=%d, date=%s, timeslot=%s", userID, date.Format("2006-01-02"), timeslot)
	// 检查是否有未完成预约
	var ongoing models.Reservation
	err := models.DB.Where("user_id = ? AND status = ? AND date >= ?", userID, "pending", time.Now().Format("2006-01-02")).Preload("User").Preload("LicensePlate").First(&ongoing).Error
	if err == nil {
		utils.WarnCtx(c, "有未结束预约，不能重复预约: user_id=%d", userID)
		return models.Reservation{}, errors.New("您有未结束的预约，不能重复预约")
	}

	// 检查上一次预约是否未上传充电记录
	var lastReservation models.Reservation
	errLast := models.DB.Where("user_id = ? AND status != ?", userID, "cancelled").Preload("User").Preload("LicensePlate").Order("date DESC").First(&lastReservation).Error
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
				utils.WarnCtx(c, "上次预约未上传充电记录: user_id=%d, last_reservation_id=%d", userID, lastReservation.ID)
				return models.Reservation{}, errors.New("上一次预约已结束但未上传充电记录，请先上传记录")
			}
		}
	}

	// 新增：同一天同一时段只能有一条有效预约（不含cancelled）
	var dupCount int64
	errDup := models.DB.Model(&models.Reservation{}).
		Where("user_id = ? AND date = ? AND timeslot = ? AND status != ?", userID, date, timeslot, "cancelled").
		Count(&dupCount).Error
	if errDup == nil && dupCount > 0 {
		utils.WarnCtx(c, "同一天同一时段已有预约: user_id=%d, date=%s, timeslot=%s", userID, date.Format("2006-01-02"), timeslot)
		return models.Reservation{}, errors.New("同一天同一时段只能有一条有效预约")
	}

	// 验证车牌号是否属于当前用户
	if licensePlateID != nil {
		var licensePlate models.LicensePlate
		err := models.DB.Where("id = ? AND user_id = ?", *licensePlateID, userID).First(&licensePlate).Error
		if err != nil {
			utils.WarnCtx(c, "车牌号不存在或不属于当前用户: user_id=%d, license_plate_id=%d", userID, *licensePlateID)
			return models.Reservation{}, errors.New("车牌号不存在或不属于当前用户")
		}
	}

	reservation := models.Reservation{
		UserID:         userID,
		Date:           date,
		Timeslot:       timeslot,
		Status:         "pending",
		Remark:         remark,
		LicensePlateID: licensePlateID,
	}
	if err := models.DB.Create(&reservation).Error; err != nil {
		utils.ErrorCtx(c, "创建预约入库失败: %v", err)
		return models.Reservation{}, err
	}
	models.DB.Preload("User").Preload("LicensePlate").First(&reservation, reservation.ID)
	utils.InfoCtx(c, "预约创建成功: user_id=%d, reservation_id=%d", userID, reservation.ID)
	return reservation, nil
}

// 取消预约
func CancelReservation(c *gin.Context, id, userID uint) error {
	utils.InfoCtx(c, "取消预约: user_id=%d, reservation_id=%d", userID, id)
	var reservation models.Reservation
	if err := models.DB.Where("id = ? AND user_id = ?", id, userID).Preload("User").Preload("LicensePlate").First(&reservation).Error; err != nil {
		utils.ErrorCtx(c, "取消预约查找失败: %v", err)
		return err
	}
	reservation.Status = "cancelled"
	err := models.DB.Save(&reservation).Error
	if err != nil {
		utils.ErrorCtx(c, "取消预约保存失败: %v", err)
	}
	return err
}

// 获取当前预约及充电记录状态
func GetCurrentReservationStatus(c *gin.Context, userID uint) (map[string]interface{}, error) {
	utils.InfoCtx(c, "查询当前预约状态: user_id=%d", userID)
	var currentRes *models.Reservation
	var lastRes *models.Reservation
	needUploadRecord := false

	var lastReservation models.Reservation
	err := models.DB.
		Where("user_id = ? AND status = ?", userID, "pending").
		Preload("User").
		Preload("LicensePlate").
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

	result := map[string]interface{}{
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

	// 添加车牌号信息
	if res.LicensePlate != nil {
		result["license_plate"] = map[string]interface{}{
			"id":           res.LicensePlate.ID,
			"plate_number": res.LicensePlate.PlateNumber,
		}
	}

	return result
}

// 创建预约
func CreateReservation(userID uint, date string, timeslot string) (*models.Reservation, error) {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}

	reservation, err := CreateReservationWithCheck(nil, userID, parsedDate, timeslot, "", nil) // Pass nil for gin.Context and licensePlateID
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
func GetCurrentReservation(c *gin.Context, userID uint) (models.Reservation, error) {
	utils.InfoCtx(c, "查询当前预约: user_id=%d", userID)
	var reservation models.Reservation
	err := models.DB.Where("user_id = ? AND status != ? AND date >= ?", userID, "cancelled", time.Now().Format("2006-01-02")).Preload("User").Preload("LicensePlate").Order("date ASC").First(&reservation).Error
	if err != nil {
		utils.WarnCtx(c, "查询当前预约失败: %v", err)
	}
	return reservation, err
}

// 获取所有未取消的预约（可按日期筛选）
func GetReservations(c *gin.Context, date string) ([]models.Reservation, error) {
	utils.InfoCtx(c, "查询预约列表: date=%s", date)
	var reservations []models.Reservation
	query := models.DB.Where("status != ?", "cancelled")
	if date != "" {
		if len(date) == 10 {
			query = query.Where("date = ?", date)
		} else if len(date) == 7 {
			query = query.Where("to_char(date, 'YYYY-MM') = ?", date)
		}
	}
	err := query.Preload("User").Preload("LicensePlate").Order("date DESC, created_at DESC").Find(&reservations).Error
	if err != nil {
		utils.ErrorCtx(c, "查询预约列表失败: %v", err)
	}
	return reservations, err
}

// GetReservationsByDate 根据日期获取预约列表
func GetReservationsByDate(date string) ([]models.Reservation, error) {
	return GetReservations(nil, date) // Pass nil for gin.Context
}

// GetCurrentStatus 获取当前状态
func GetCurrentStatus(userID uint) (map[string]interface{}, error) {
	return GetCurrentReservationStatus(nil, userID) // Pass nil for gin.Context
}
