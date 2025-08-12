package service

import (
	"errors"
	"shared-charge/models"
	"shared-charge/utils"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
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
	Date           string
	KWH            float64
	ImageURL       string
	Remark         string
	ReservationID  uint
	UnitPrice      float64
	UserID         uint
	Timeslot       string
	LicensePlateID *uint
}

func CreateRecordWithTimeslot(c *gin.Context, req CreateRecordRequest) error {
	utils.InfoCtx(c, "创建充电记录: user_id=%d, date=%s, kwh=%v, reservation_id=%d, image_url=%s", req.UserID, req.Date, req.KWH, req.ReservationID, req.ImageURL)
	date, err := time.ParseInLocation("2006-01-02", req.Date, time.Local)
	if err != nil {
		utils.WarnCtx(c, "创建充电记录日期格式错误: %v", err)
		return err
	}
	timeslot := req.Timeslot
	if req.ReservationID != 0 && timeslot == "" {
		var reservation models.Reservation
		err := models.DB.Select("timeslot").First(&reservation, req.ReservationID).Error
		if err == nil {
			timeslot = reservation.Timeslot
		} else {
			utils.WarnCtx(c, "查找预约时段失败: reservation_id=%d, err=%v", req.ReservationID, err)
		}
	}
	// 新增：校验预约必须为pending状态，且一个预约只能有一条record
	if req.ReservationID != 0 {
		var reservation models.Reservation
		errRes := models.DB.First(&reservation, req.ReservationID).Error
		if errRes != nil {
			utils.WarnCtx(c, "预约不存在: reservation_id=%d", req.ReservationID)
			return errRes
		}
		if reservation.Status != "pending" {
			utils.WarnCtx(c, "预约状态不是pending: reservation_id=%d, status=%s", req.ReservationID, reservation.Status)
			return errors.New("预约状态必须为pending")
		}
		var count int64
		models.DB.Model(&models.Record{}).Where("reservation_id = ?", req.ReservationID).Count(&count)
		if count > 0 {
			utils.WarnCtx(c, "该预约已上传过充电记录: reservation_id=%d", req.ReservationID)
			return errors.New("一个预约只能上传一条充电记录")
		}
	}
	// 验证车牌号是否属于当前用户
	if req.LicensePlateID != nil {
		var licensePlate models.LicensePlate
		err := models.DB.Where("id = ? AND user_id = ?", *req.LicensePlateID, req.UserID).First(&licensePlate).Error
		if err != nil {
			utils.WarnCtx(c, "车牌号不存在或不属于当前用户: user_id=%d, license_plate_id=%d", req.UserID, *req.LicensePlateID)
			return errors.New("车牌号不存在或不属于当前用户")
		}
	}

	record := &models.Record{
		UserID:         req.UserID,
		Date:           date,
		KWH:            req.KWH,
		UnitPrice:      req.UnitPrice,
		ImageURL:       req.ImageURL,
		Remark:         req.Remark,
		ReservationID:  req.ReservationID,
		Timeslot:       timeslot,
		LicensePlateID: req.LicensePlateID,
	}
	utils.InfoCtx(c, "即将写入数据库的 record.ImageURL=%s", record.ImageURL)
	record.CalculateAmount()
	errCreate := models.DB.Create(record).Error
	if errCreate != nil {
		utils.ErrorCtx(c, "充电记录入库失败: %v", errCreate)
		return errCreate
	}
	utils.InfoCtx(c, "充电记录创建成功: user_id=%d, record_id=%d, image_url=%s", req.UserID, record.ID, record.ImageURL)
	// 新增：自动将预约状态设为 completed
	if req.ReservationID != 0 {
		if err := SetReservationCompleted(req.ReservationID, req.UserID); err != nil {
			utils.WarnCtx(c, "设置预约完成状态失败: reservation_id=%d, user_id=%d, err=%v", req.ReservationID, req.UserID, err)
		} else {
			utils.InfoCtx(c, "预约状态已设为 completed: reservation_id=%d, user_id=%d", req.ReservationID, req.UserID)
		}
	}
	return nil
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
func GetRecentRecordsWithTimeslotByUser(c *gin.Context, userID uint, limit int) ([]map[string]interface{}, error) {
	utils.InfoCtx(c, "查询用户充电记录: user_id=%d, limit=%d", userID, limit)
	if limit <= 0 {
		limit = defaultLimit
	}
	var records []models.Record
	err := models.DB.Where("user_id = ?", userID).Preload("LicensePlate").Order("created_at DESC").Limit(limit).Find(&records).Error
	if err != nil {
		utils.ErrorCtx(c, "查询用户充电记录失败: %v", err)
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
		result := map[string]interface{}{
			"date":     r.Date.Format("2006-01-02"),
			"kwh":      r.KWH,
			"amount":   r.Amount,
			"timeslot": timeslot,
		}

		// 添加车牌号信息
		if r.LicensePlate != nil {
			result["license_plate"] = map[string]interface{}{
				"id":           r.LicensePlate.ID,
				"plate_number": r.LicensePlate.PlateNumber,
			}
		}

		resp = append(resp, result)
	}
	utils.InfoCtx(c, "查询用户充电记录成功: user_id=%d, count=%d", userID, len(records))
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

// GetRecordsByMonth 获取指定月份的充电记录列表
func GetRecordsByMonth(userID uint, month string) ([]map[string]interface{}, error) {
	// 获取月份日期范围
	startDate, endDate, err := getMonthDateRange(month)
	if err != nil {
		return nil, err
	}

	var records []models.Record
	err = models.DB.Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).
		Preload("LicensePlate").
		Order("date DESC, created_at DESC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	// 组装返回数据
	var resp []map[string]interface{}
	for _, record := range records {
		result := map[string]interface{}{
			"id":         record.ID,
			"date":       record.Date.Format("2006-01-02"),
			"timeslot":   record.Timeslot,
			"kwh":        record.KWH,
			"amount":     record.Amount,
			"remark":     record.Remark,
			"image_url":  record.ImageURL,
			"created_at": record.CreatedAt,
			"updated_at": record.UpdatedAt,
		}

		// 添加车牌号信息
		if record.LicensePlate != nil {
			result["license_plate"] = map[string]interface{}{
				"id":           record.LicensePlate.ID,
				"plate_number": record.LicensePlate.PlateNumber,
			}
		}

		resp = append(resp, result)
	}

	return resp, nil
}

// GetRecordByID 根据ID获取充电记录详情
func GetRecordByID(userID uint, recordID string) (map[string]interface{}, error) {
	var record models.Record
	err := models.DB.Where("id = ? AND user_id = ?", recordID, userID).Preload("LicensePlate").First(&record).Error
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":         record.ID,
		"date":       record.Date.Format("2006-01-02"),
		"timeslot":   record.Timeslot,
		"kwh":        record.KWH,
		"amount":     record.Amount,
		"remark":     record.Remark,
		"image_url":  record.ImageURL,
		"created_at": record.CreatedAt,
		"updated_at": record.UpdatedAt,
	}

	// 添加车牌号信息
	if record.LicensePlate != nil {
		result["license_plate"] = map[string]interface{}{
			"id":           record.LicensePlate.ID,
			"plate_number": record.LicensePlate.PlateNumber,
		}
	}

	return result, nil
}

// UpdateRecordByID 根据ID更新充电记录
func UpdateRecordByID(userID uint, recordID string, req UpdateRecordRequest) (map[string]interface{}, error) {
	var record models.Record
	err := models.DB.Where("id = ? AND user_id = ?", recordID, userID).First(&record).Error
	if err != nil {
		return nil, err
	}

	// 验证车牌号是否属于当前用户
	if req.LicensePlateID != nil {
		var licensePlate models.LicensePlate
		err := models.DB.Where("id = ? AND user_id = ?", *req.LicensePlateID, userID).First(&licensePlate).Error
		if err != nil {
			return nil, errors.New("车牌号不存在或不属于当前用户")
		}
	}

	// 更新记录
	updates := map[string]interface{}{
		"kwh":        req.KWH,
		"remark":     req.Remark,
		"updated_at": time.Now(),
	}

	// 如果提供了新的图片URL，则更新
	if req.ImageURL != "" {
		updates["image_url"] = req.ImageURL
	}

	// 更新车牌号
	if req.LicensePlateID != nil {
		updates["license_plate_id"] = req.LicensePlateID
	}

	// 重新计算费用
	record.KWH = req.KWH
	record.CalculateAmount()
	updates["amount"] = record.Amount

	err = models.DB.Model(&record).Updates(updates).Error
	if err != nil {
		return nil, err
	}

	// 返回更新后的记录
	return map[string]interface{}{
		"id":         record.ID,
		"date":       record.Date.Format("2006-01-02"),
		"timeslot":   record.Timeslot,
		"kwh":        record.KWH,
		"amount":     record.Amount,
		"remark":     record.Remark,
		"image_url":  record.ImageURL,
		"created_at": record.CreatedAt,
		"updated_at": record.UpdatedAt,
	}, nil
}

// UpdateRecordRequest 更新充电记录请求结构
type UpdateRecordRequest struct {
	KWH            float64 `json:"kwh"`
	ImageURL       string  `json:"image_url"`
	Remark         string  `json:"remark"`
	LicensePlateID *uint   `json:"license_plate_id"`
}
