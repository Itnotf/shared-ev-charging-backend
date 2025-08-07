package service

import (
	"shared-charge/models"
	"time"

	"github.com/gin-gonic/gin"
)

// GetAllUsers 获取所有用户列表
func GetAllUsers(c *gin.Context) ([]map[string]interface{}, error) {
	var users []models.User
	if err := models.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	for _, user := range users {
		result = append(result, map[string]interface{}{
			"id":          user.ID,
			"user_name":   user.Name,
			"avatar":      user.Avatar,
			"can_reserve": user.CanReserve,
			"unit_price":  user.UnitPrice,
			"role":        user.Role,
		})
	}
	return result, nil
}

// UpdateUserCanReserve 更新用户可预约状态
func UpdateUserCanReserve(c *gin.Context, userID uint, canReserve bool) error {
	return models.DB.Model(&models.User{}).Where("id = ?", userID).Update("can_reserve", canReserve).Error
}

// UpdateUserUnitPrice 更新用户电价
func UpdateUserUnitPrice(c *gin.Context, userID uint, unitPrice float64) error {
	return models.DB.Model(&models.User{}).Where("id = ?", userID).Update("unit_price", unitPrice).Error
}

// GetMonthlyReport 获取月度对账数据
func GetMonthlyReport(c *gin.Context, month string) (map[string]interface{}, error) {
	startDate, _ := time.Parse("2006-01", month)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	// 只查 can_reserve = true 的用户
	var users []models.User
	if err := models.DB.Where("can_reserve = ?", true).Find(&users).Error; err != nil {
		return nil, err
	}

	// 统计所有用户本月预约总数
	var reservationStats []struct {
		UserID uint
		Total  int64
	}
	models.DB.Model(&models.Reservation{}).
		Select("user_id, COUNT(*) as total").
		Where("date >= ? AND date <= ? AND status != ?", startDate, endDate, "cancelled").
		Group("user_id").
		Scan(&reservationStats)

	// 统计所有用户本月已上传预约数
	var uploadedStats []struct {
		UserID   uint
		Uploaded int64
	}
	models.DB.Table("reservations").
		Select("reservations.user_id, COUNT(DISTINCT reservations.id) as uploaded").
		Joins("JOIN records ON reservations.id = records.reservation_id").
		Where("reservations.date >= ? AND reservations.date <= ? AND reservations.status != ?", startDate, endDate, "cancelled").
		Group("reservations.user_id").
		Scan(&uploadedStats)

	// 构建 userId -> 预约数/上传数 map
	reservationMap := make(map[uint]int64)
	uploadedMap := make(map[uint]int64)
	for _, stat := range reservationStats {
		reservationMap[stat.UserID] = stat.Total
	}
	for _, stat := range uploadedStats {
		uploadedMap[stat.UserID] = stat.Uploaded
	}

	var result []map[string]interface{}
	for _, user := range users {
		// 修正聚合查询
		type aggResult struct {
			TotalAmount int64
		}
		var agg aggResult
		models.DB.Model(&models.Record{}).
			Select("COALESCE(SUM(amount), 0) as total_amount").
			Where("user_id = ? AND date >= ? AND date <= ?", user.ID, startDate, endDate).
			Scan(&agg)

		// 计算 has_uploaded
		totalReservations := reservationMap[user.ID]
		uploadedReservations := uploadedMap[user.ID]
		hasUploaded := true
		if totalReservations > 0 && uploadedReservations < totalReservations {
			hasUploaded = false
		}

		userData := map[string]interface{}{
			"id":           user.ID,
			"user_name":    user.Name,
			"avatar":       user.Avatar,
			"total_amount": float64(agg.TotalAmount) / 100.0,
			"has_uploaded": hasUploaded,
		}
		result = append(result, userData)
	}

	return map[string]interface{}{
		"month": month,
		"users": result,
	}, nil
}
