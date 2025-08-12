package controllers

import (
	"net/http"
	"shared-charge/service"

	"shared-charge/config"
	"shared-charge/utils"

	"github.com/gin-gonic/gin"
)

// CreateRecordRequest 创建充电记录请求
type CreateRecordRequest struct {
	Date           string  `json:"date" binding:"required"`
	KWH            float64 `json:"kwh" binding:"required,gt=0"`
	ImageURL       string  `json:"image_url"`
	Remark         string  `json:"remark"`
	ReservationID  uint    `json:"reservation_id"`
	Timeslot       string  `json:"timeslot"`
	LicensePlateID *uint   `json:"license_plate_id"`
}

// GetRecords 获取充电记录列表
// @Summary 获取充电记录列表
// @Description 获取充电记录列表，支持按日期筛选
// @Tags 充电记录
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param date query string false "日期筛选 (YYYY-MM-DD)"
// @Success 200 {array} models.Record
// @Router /records [get]
func GetRecords(c *gin.Context) {
	utils.InfoCtx(c, "获取充电记录列表请求")
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		utils.WarnCtx(c, "获取充电记录未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}
	records, err := service.GetRecentRecordsWithTimeslotByUser(c, userModel.ID, 50)
	if err != nil {
		utils.ErrorCtx(c, "获取充电记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取充电记录失败"})
		return
	}
	utils.InfoCtx(c, "获取充电记录成功: user_id=%d, count=%d", userModel.ID, len(records))
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取充电记录成功", "data": records})
}

// CreateRecord 创建充电记录
// @Summary 创建充电记录
// @Description 创建新的充电记录
// @Tags 充电记录
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRecordRequest true "充电记录请求"
// @Success 200 {object} models.Record
// @Router /records [post]
func CreateRecord(c *gin.Context) {
	utils.InfoCtx(c, "创建充电记录请求")
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		utils.WarnCtx(c, "创建充电记录未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}
	var req CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.WarnCtx(c, "创建充电记录参数校验失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}
	if req.ReservationID == 0 {
		utils.WarnCtx(c, "创建充电记录缺少预约ID")
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少预约ID"})
		return
	}
	unitPrice := userModel.UnitPrice
	if unitPrice <= 0 {
		unitPrice = config.GetConfig().App.DefaultUnitPrice
	}
	createReq := service.CreateRecordRequest{
		UserID:         userModel.ID,
		Date:           req.Date,
		KWH:            req.KWH,
		ReservationID:  req.ReservationID,
		Timeslot:       req.Timeslot,
		UnitPrice:      unitPrice,
		ImageURL:       req.ImageURL, // 修复：传递 image_url
		Remark:         req.Remark,   // 修复：传递 remark
		LicensePlateID: req.LicensePlateID,
	}
	if err := service.CreateRecordWithTimeslot(c, createReq); err != nil {
		utils.ErrorCtx(c, "创建充电记录失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "创建充电记录失败"})
		return
	}
	utils.InfoCtx(c, "充电记录创建成功: user_id=%d, reservation_id=%d", userModel.ID, req.ReservationID)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "充电记录创建成功"})
}

// GetUnsubmittedRecords 获取当前用户未提交的充电记录
// @Summary 获取当前用户未提交的充电记录
// @Description 获取当前用户未提交的充电记录
// @Tags Record
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /records/unsubmitted [get]
func GetUnsubmittedRecords(c *gin.Context) {
	utils.InfoCtx(c, "获取未提交充电记录请求")
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		utils.WarnCtx(c, "获取未提交充电记录未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}
	records, err := service.GetUnsubmittedRecords(userModel.ID)
	if err != nil {
		utils.ErrorCtx(c, "获取未提交充电记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取未提交充电记录失败"})
		return
	}
	utils.InfoCtx(c, "获取未提交充电记录成功: user_id=%d, count=%d", userModel.ID, len(records))
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取未提交充电记录成功", "data": records})
}

// GetMonthlyStatistics 月度统计
// @Summary 获取月度累计用电量和费用
// @Description 获取指定月份的累计用电量和累计费用
// @Tags 统计
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param month query string true "月份(YYYY-MM)"
// @Success 200 {object} map[string]interface{}
// @Router /statistics/monthly [get]
func GetMonthlyStatistics(c *gin.Context) {
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		return
	}
	month := c.Query("month")
	if month == "" || len(month) != 7 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数month格式错误，应为YYYY-MM"})
		return
	}
	totalKwh, totalCost, err := service.GetMonthlyStatistics(userModel.ID, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "数据库查询失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": gin.H{"totalKwh": totalKwh, "totalCost": totalCost}})
}

// GetDailyStatistics 日用电量统计
// @Summary 获取指定月份每日用电量
// @Description 获取指定月份每日的累计用电量
// @Tags 统计
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param month query string true "月份(YYYY-MM)"
// @Success 200 {array} map[string]interface{}
// @Router /statistics/daily [get]
func GetDailyStatistics(c *gin.Context) {
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		return
	}
	month := c.Query("month")
	if month == "" || len(month) != 7 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数month格式错误，应为YYYY-MM"})
		return
	}
	resp, err := service.GetDailyStatisticsWithShift(userModel.ID, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "数据库查询失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": resp})
}

// GetMonthlyShiftStatistics 月度分时段统计
// @Summary 获取月度分时段统计
// @Description 获取指定月份的分时段统计信息
// @Tags 统计
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param month query string true "月份(YYYY-MM)"
// @Success 200 {object} map[string]interface{}
// @Router /statistics/monthly-shift [get]
func GetMonthlyShiftStatistics(c *gin.Context) {
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		return
	}
	month := c.Query("month")
	if month == "" || len(month) != 7 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数month格式错误，应为YYYY-MM"})
		return
	}
	dayKwh, nightKwh, totalKwh, err := service.GetMonthlyShiftStatistics(userModel.ID, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "数据库查询失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": gin.H{"dayKwh": dayKwh, "nightKwh": nightKwh, "totalKwh": totalKwh}})
}

// UpdateRecordRequest 更新充电记录请求
type UpdateRecordRequest struct {
	KWH            float64 `json:"kwh" binding:"required,gt=0"`
	ImageURL       string  `json:"image_url"`
	Remark         string  `json:"remark"`
	LicensePlateID *uint   `json:"license_plate_id"`
}

// GetRecordsList 获取充电记录列表（按月筛选）
// @Summary 获取充电记录列表
// @Description 获取指定月份的充电记录列表
// @Tags 充电记录
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param month query string true "月份 (YYYY-MM)"
// @Success 200 {object} map[string]interface{}
// @Router /records/list [get]
func GetRecordsList(c *gin.Context) {
	utils.InfoCtx(c, "获取充电记录列表请求")
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		utils.WarnCtx(c, "获取充电记录列表未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	month := c.Query("month")
	if month == "" {
		utils.WarnCtx(c, "缺少月份参数")
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少月份参数"})
		return
	}

	records, err := service.GetRecordsByMonth(userModel.ID, month)
	if err != nil {
		utils.ErrorCtx(c, "获取充电记录列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取充电记录列表失败"})
		return
	}

	utils.InfoCtx(c, "获取充电记录列表成功: user_id=%d, month=%s, count=%d", userModel.ID, month, len(records))
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": records})
}

// GetRecordDetail 获取充电记录详情
// @Summary 获取充电记录详情
// @Description 根据记录ID获取充电记录详情
// @Tags 充电记录
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "记录ID"
// @Success 200 {object} map[string]interface{}
// @Router /records/{id} [get]
func GetRecordDetail(c *gin.Context) {
	utils.InfoCtx(c, "获取充电记录详情请求")
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		utils.WarnCtx(c, "获取充电记录详情未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	recordID := c.Param("id")
	if recordID == "" {
		utils.WarnCtx(c, "缺少记录ID参数")
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少记录ID参数"})
		return
	}

	record, err := service.GetRecordByID(userModel.ID, recordID)
	if err != nil {
		utils.ErrorCtx(c, "获取充电记录详情失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取充电记录详情失败"})
		return
	}

	if record == nil {
		utils.WarnCtx(c, "充电记录不存在: record_id=%s, user_id=%d", recordID, userModel.ID)
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "充电记录不存在"})
		return
	}

	utils.InfoCtx(c, "获取充电记录详情成功: record_id=%s, user_id=%d", recordID, userModel.ID)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": record})
}

// UpdateRecord 更新充电记录
// @Summary 更新充电记录
// @Description 根据记录ID更新充电记录信息
// @Tags 充电记录
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "记录ID"
// @Param request body UpdateRecordRequest true "更新充电记录请求"
// @Success 200 {object} map[string]interface{}
// @Router /records/{id} [put]
func UpdateRecord(c *gin.Context) {
	utils.InfoCtx(c, "更新充电记录请求")
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		utils.WarnCtx(c, "更新充电记录未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	recordID := c.Param("id")
	if recordID == "" {
		utils.WarnCtx(c, "缺少记录ID参数")
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少记录ID参数"})
		return
	}

	var req UpdateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.WarnCtx(c, "更新充电记录参数校验失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}

	updatedRecord, err := service.UpdateRecordByID(userModel.ID, recordID, service.UpdateRecordRequest{
		KWH:            req.KWH,
		ImageURL:       req.ImageURL,
		Remark:         req.Remark,
		LicensePlateID: req.LicensePlateID,
	})
	if err != nil {
		utils.ErrorCtx(c, "更新充电记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新充电记录失败"})
		return
	}

	if updatedRecord == nil {
		utils.WarnCtx(c, "充电记录不存在: record_id=%s, user_id=%d", recordID, userModel.ID)
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "充电记录不存在"})
		return
	}

	utils.InfoCtx(c, "更新充电记录成功: record_id=%s, user_id=%d", recordID, userModel.ID)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": updatedRecord})
}
