package service

import (
	"shared-charge/models"
	"sort"
	"time"
)

const defaultLimit = 50

// 日期处理辅助函数
func getMonthDateRange(month string) (string, string, error) {
	// 解析月份字符串 "2025-07"
	parsedTime, err := time.Parse("2006-01", month)
	if err != nil {
		return "", "", err
	}

	// 计算月份开始日期
	startDate := parsedTime.Format("2006-01-02")

	// 计算月份结束日期（下个月第一天减一天）
	endDate := parsedTime.AddDate(0, 1, 0).AddDate(0, 0, -1).Format("2006-01-02")

	return startDate, endDate, nil
}

// 获取用户最近N条充电记录
func GetRecentRecordsByUser(userID uint, limit int) ([]models.Record, error) {
	if limit <= 0 {
		limit = defaultLimit
	}
	var records []models.Record
	err := models.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&records).Error
	return records, err
}

// 创建充电记录（自动查预约表获取 timeslot）
type CreateRecordRequest struct {
	Date          string
	KWH           float64
	ImageURL      string
	Remark        string
	ReservationID uint
	UnitPrice     float64
	UserID        uint
	Timeslot      string
}

func CreateRecordWithTimeslot(req CreateRecordRequest) error {
	date, err := time.ParseInLocation("2006-01-02", req.Date, time.Local)
	if err != nil {
		return err
	}
	timeslot := req.Timeslot
	if req.ReservationID != 0 && timeslot == "" {
		var reservation models.Reservation
		err := models.DB.Select("timeslot").First(&reservation, req.ReservationID).Error
		if err == nil {
			timeslot = reservation.Timeslot
		}
	}
	record := &models.Record{
		UserID:        req.UserID,
		Date:          date,
		KWH:           req.KWH,
		UnitPrice:     req.UnitPrice,
		ImageURL:      req.ImageURL,
		Remark:        req.Remark,
		ReservationID: req.ReservationID,
		Timeslot:      timeslot,
	}
	record.CalculateAmount()
	return models.DB.Create(record).Error
}

// 获取未提交的充电记录
func GetUnsubmittedRecords(userID uint) ([]models.Record, error) {
	var records []models.Record
	err := models.DB.Where("user_id = ? AND reservation_id = 0", userID).Find(&records).Error
	return records, err
}

// 统计相关方法略，可根据需要补充

// 获取月度累计用电量和费用
func GetMonthlyStatistics(userID uint, month string) (float64, float64, error) {
	var totalKwh, totalCost float64

	// 获取月份日期范围
	startDate, endDate, err := getMonthDateRange(month)
	if err != nil {
		return 0, 0, err
	}

	err = models.DB.Model(&models.Record{}).
		Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).
		Select("COALESCE(SUM(kwh),0), COALESCE(SUM(amount),0)").
		Row().Scan(&totalKwh, &totalCost)
	return totalKwh, totalCost, err
}

// 获取指定月份每日用电量
func GetDailyStatistics(userID uint, month string) ([]map[string]interface{}, error) {
	var results []struct {
		Date     string  `json:"date"`
		TotalKwh float64 `json:"totalKwh"`
	}

	// 获取月份日期范围
	startDate, endDate, err := getMonthDateRange(month)
	if err != nil {
		return nil, err
	}

	err = models.DB.Model(&models.Record{}).
		Select("to_char(date, 'YYYY-MM-DD') as date, COALESCE(SUM(kwh),0) as total_kwh").
		Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).
		Group("date").
		Order("date").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	var resp []map[string]interface{}
	for _, r := range results {
		resp = append(resp, map[string]interface{}{
			"date":     r.Date,
			"totalKwh": r.TotalKwh,
		})
	}
	return resp, nil
}

// 获取指定月份白班、夜班和总用电量
func GetMonthlyShiftStatistics(userID uint, month string) (float64, float64, float64, error) {
	var result struct {
		DayKwh   float64 `json:"day_kwh"`
		NightKwh float64 `json:"night_kwh"`
	}

	// 获取月份日期范围
	startDate, endDate, err := getMonthDateRange(month)
	if err != nil {
		return 0, 0, 0, err
	}

	// 使用单次查询替代两次独立查询
	err = models.DB.Model(&models.Record{}).
		Select(`
			COALESCE(SUM(CASE WHEN r.timeslot = 'day' THEN records.kwh ELSE 0 END), 0) as day_kwh,
			COALESCE(SUM(CASE WHEN r.timeslot = 'night' THEN records.kwh ELSE 0 END), 0) as night_kwh
		`).
		Joins("LEFT JOIN reservations r ON records.reservation_id = r.id").
		Where("records.user_id = ? AND records.date >= ? AND records.date <= ?", userID, startDate, endDate).
		Scan(&result).Error

	if err != nil {
		return 0, 0, 0, err
	}

	totalKwh := result.DayKwh + result.NightKwh
	return result.DayKwh, result.NightKwh, totalKwh, nil
}

// 获取用户最近N条充电记录（带timeslot）
func GetRecentRecordsWithTimeslotByUser(userID uint, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = defaultLimit
	}
	var records []models.Record
	err := models.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&records).Error
	if err != nil {
		return nil, err
	}
	// 批量查找所有涉及的预约ID
	reservationIDs := make(map[uint]struct{})
	for _, r := range records {
		if r.ReservationID != 0 {
			reservationIDs[r.ReservationID] = struct{}{}
		}
	}
	// 批量查预约
	reservations := make(map[uint]string)
	if len(reservationIDs) > 0 {
		var resList []models.Reservation
		ids := make([]uint, 0, len(reservationIDs))
		for id := range reservationIDs {
			ids = append(ids, id)
		}
		_ = models.DB.Select("id, timeslot").Where("id IN ?", ids).Find(&resList).Error
		for _, res := range resList {
			reservations[res.ID] = res.Timeslot
		}
	}
	// 组装结果
	var resp []map[string]interface{}
	for _, r := range records {
		timeslot := ""
		if r.ReservationID != 0 {
			timeslot = reservations[r.ReservationID]
		}
		resp = append(resp, map[string]interface{}{
			"date":     r.Date.Format("2006-01-02"),
			"kwh":      r.KWH,
			"amount":   r.Amount,
			"timeslot": timeslot,
		})
	}
	return resp, nil
}

// 设置预约状态为 completed
func SetReservationCompleted(reservationID, userID uint) error {
	return models.DB.Model(&models.Reservation{}).Where("id = ? AND user_id = ?", reservationID, userID).Update("status", "completed").Error
}

// 获取指定月份每日白班、夜班、总用电量，按日期倒序（直接用冗余字段timeslot）
func GetDailyStatisticsWithShift(userID uint, month string) ([]map[string]interface{}, error) {
	var results []struct {
		Date     string
		Timeslot string
		TotalKwh float64
	}

	// 获取月份日期范围
	startDate, endDate, err := getMonthDateRange(month)
	if err != nil {
		return nil, err
	}

	err = models.DB.Model(&models.Record{}).
		Select("to_char(date, 'YYYY-MM-DD') as date, timeslot, COALESCE(SUM(kwh),0) as total_kwh").
		Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).
		Group("date, timeslot").
		Order("date DESC").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	// 组装每天的 dayKwh/nightKwh/totalKwh
	statMap := make(map[string]map[string]float64)
	for _, r := range results {
		if statMap[r.Date] == nil {
			statMap[r.Date] = map[string]float64{"dayKwh": 0, "nightKwh": 0}
		}
		if r.Timeslot == "day" {
			statMap[r.Date]["dayKwh"] = r.TotalKwh
		} else if r.Timeslot == "night" {
			statMap[r.Date]["nightKwh"] = r.TotalKwh
		}
	}
	var resp []map[string]interface{}
	for day, stat := range statMap {
		total := stat["dayKwh"] + stat["nightKwh"]
		resp = append(resp, map[string]interface{}{
			"date":     day,
			"dayKwh":   stat["dayKwh"],
			"nightKwh": stat["nightKwh"],
			"totalKwh": total,
		})
	}
	sort.Slice(resp, func(i, j int) bool {
		return resp[i]["date"].(string) > resp[j]["date"].(string)
	})
	return resp, nil
}
