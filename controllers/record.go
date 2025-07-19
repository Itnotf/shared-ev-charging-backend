package controllers

import (
	"net/http"
	"shared-charge/service"

	"shared-charge/utils"

	"github.com/gin-gonic/gin"
)

// CreateRecordRequest 创建充电记录请求
type CreateRecordRequest struct {
	Date          string  `json:"date" binding:"required"`
	KWH           float64 `json:"kwh" binding:"required,gt=0"`
	ImageURL      string  `json:"image_url"`
	Remark        string  `json:"remark"`
	ReservationID uint    `json:"reservation_id"`
	Timeslot      string  `json:"timeslot"`
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
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		return
	}
	records, err := service.GetRecentRecordsWithTimeslotByUser(userModel.ID, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取充电记录失败", "error": err.Error()})
		return
	}
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
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		return
	}
	var req CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}
	if req.ReservationID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少预约ID"})
		return
	}
	createReq := service.CreateRecordRequest{
		UserID:        userModel.ID,
		Date:          req.Date,
		KWH:           req.KWH,
		ReservationID: req.ReservationID,
		Timeslot:      req.Timeslot,
	}
	if err := service.CreateRecordWithTimeslot(createReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "创建充电记录失败", "error": err.Error()})
		return
	}
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
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		return
	}
	records, err := service.GetUnsubmittedRecords(userModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取未提交充电记录失败", "error": err.Error()})
		return
	}
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
