package controllers

import (
	"fmt"
	"net/http"
	"shared-charge/models"
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
	lg := utils.CtxLogger(c)
	user, exists := c.Get("user")
	if !exists {
		lg.Error("获取充电记录失败: 用户未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	userModel, ok := user.(models.User)
	if !ok {
		lg.Error("获取充电记录失败: 用户信息类型错误")
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}

	lg.Info(fmt.Sprintf("获取充电记录: UserID=%d", userModel.ID))

	records, err := service.GetRecentRecordsWithTimeslotByUser(userModel.ID, 50)
	if err != nil {
		lg.Error(fmt.Sprintf("获取充电记录失败: UserID=%d, Error=%v", userModel.ID, err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取充电记录失败", "error": err.Error()})
		return
	}

	lg.Info(fmt.Sprintf("获取充电记录成功: UserID=%d, count=%d", userModel.ID, len(records)))
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
	lg := utils.CtxLogger(c)
	user, exists := c.Get("user")
	if !exists {
		lg.Error("创建充电记录失败: 用户未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	userModel, ok := user.(models.User)
	if !ok {
		lg.Error("创建充电记录失败: 用户信息类型错误")
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}

	var req CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		lg.Error(fmt.Sprintf("创建充电记录失败: 请求参数错误: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}

	if req.ReservationID == 0 {
		lg.Error("创建充电记录失败: 缺少预约ID")
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少预约ID"})
		return
	}

	lg.Info(fmt.Sprintf("创建充电记录: UserID=%d, Date=%s, KWH=%f, ReservationID=%d", userModel.ID, req.Date, req.KWH, req.ReservationID))

	createReq := service.CreateRecordRequest{
		UserID:        userModel.ID,
		Date:          req.Date,
		KWH:           req.KWH,
		ReservationID: req.ReservationID,
		Timeslot:      req.Timeslot,
	}

	if err := service.CreateRecordWithTimeslot(createReq); err != nil {
		lg.Error(fmt.Sprintf("创建充电记录失败: UserID=%d, Error=%v", userModel.ID, err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "创建充电记录失败", "error": err.Error()})
		return
	}

	lg.Info(fmt.Sprintf("充电记录创建成功: UserID=%d, ReservationID=%d", userModel.ID, req.ReservationID))
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
	lg := utils.CtxLogger(c)
	user, exists := c.Get("user")
	if !exists {
		lg.Error("获取未提交充电记录失败: 用户未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}
	userModel, ok := user.(models.User)
	if !ok {
		lg.Error("获取未提交充电记录失败: 用户信息类型错误")
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}
	lg.Info(fmt.Sprintf("获取未提交充电记录: UserID=%d", userModel.ID))
	records, err := service.GetUnsubmittedRecords(userModel.ID)
	if err != nil {
		lg.Error(fmt.Sprintf("获取未提交充电记录失败: UserID=%d, Error=%v", userModel.ID, err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取未提交充电记录失败", "error": err.Error()})
		return
	}
	lg.Info(fmt.Sprintf("获取未提交充电记录成功: UserID=%d, count=%d", userModel.ID, len(records)))
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
	user, exists := c.Get("user")
	if !exists {
		utils.Error(fmt.Sprintf("获取月度统计失败: 用户未认证"))
		c.JSON(401, gin.H{"code": 401, "message": "未认证"})
		return
	}
	userModel, ok := user.(models.User)
	if !ok {
		utils.Error(fmt.Sprintf("获取月度统计失败: 用户信息类型错误"))
		c.JSON(500, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}
	month := c.Query("month")
	if month == "" || len(month) != 7 {
		utils.Error(fmt.Sprintf("获取月度统计失败: 月份格式错误: %s", month))
		c.JSON(400, gin.H{"code": 400, "message": "参数month格式错误，应为YYYY-MM"})
		return
	}
	utils.Info(fmt.Sprintf("获取月度统计: UserID=%d, Month=%s", userModel.ID, month))
	totalKwh, totalCost, err := service.GetMonthlyStatistics(userModel.ID, month)
	if err != nil {
		utils.Error(fmt.Sprintf("获取月度统计失败: UserID=%d, Month=%s, Error=%v", userModel.ID, month, err))
		c.JSON(500, gin.H{"code": 500, "message": "数据库查询失败", "error": err.Error()})
		return
	}
	utils.Info(fmt.Sprintf("获取月度统计成功: UserID=%d, Month=%s, TotalKWH=%f, TotalCost=%f", userModel.ID, month, totalKwh, totalCost))
	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"totalKwh":  totalKwh,
			"totalCost": totalCost,
		},
	})
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
	user, exists := c.Get("user")
	if !exists {
		utils.Error(fmt.Sprintf("获取日统计失败: 用户未认证"))
		c.JSON(401, gin.H{"code": 401, "message": "未认证"})
		return
	}
	userModel, ok := user.(models.User)
	if !ok {
		utils.Error(fmt.Sprintf("获取日统计失败: 用户信息类型错误"))
		c.JSON(500, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}
	month := c.Query("month")
	if month == "" || len(month) != 7 {
		utils.Error(fmt.Sprintf("获取日统计失败: 月份格式错误: %s", month))
		c.JSON(400, gin.H{"code": 400, "message": "参数month格式错误，应为YYYY-MM"})
		return
	}
	utils.Info(fmt.Sprintf("获取日统计: UserID=%d, Month=%s", userModel.ID, month))
	resp, err := service.GetDailyStatisticsWithShift(userModel.ID, month)
	if err != nil {
		utils.Error(fmt.Sprintf("获取日统计失败: UserID=%d, Month=%s, Error=%v", userModel.ID, month, err))
		c.JSON(500, gin.H{"code": 500, "message": "数据库查询失败", "error": err.Error()})
		return
	}
	utils.Info(fmt.Sprintf("获取日统计成功: UserID=%d, Month=%s, count=%d", userModel.ID, month, len(resp)))
	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"data":    resp,
	})
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
	user, exists := c.Get("user")
	if !exists {
		utils.Error(fmt.Sprintf("获取月度分时段统计失败: 用户未认证"))
		c.JSON(401, gin.H{"code": 401, "message": "未认证"})
		return
	}
	userModel, ok := user.(models.User)
	if !ok {
		utils.Error(fmt.Sprintf("获取月度分时段统计失败: 用户信息类型错误"))
		c.JSON(500, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}
	month := c.Query("month")
	if month == "" || len(month) != 7 {
		utils.Error(fmt.Sprintf("获取月度分时段统计失败: 月份格式错误: %s", month))
		c.JSON(400, gin.H{"code": 400, "message": "参数month格式错误，应为YYYY-MM"})
		return
	}
	utils.Info(fmt.Sprintf("获取月度分时段统计: UserID=%d, Month=%s", userModel.ID, month))
	dayKwh, nightKwh, totalKwh, err := service.GetMonthlyShiftStatistics(userModel.ID, month)
	if err != nil {
		utils.Error(fmt.Sprintf("获取月度分时段统计失败: UserID=%d, Month=%s, Error=%v", userModel.ID, month, err))
		c.JSON(500, gin.H{"code": 500, "message": "数据库查询失败", "error": err.Error()})
		return
	}
	utils.Info(fmt.Sprintf("获取月度分时段统计成功: UserID=%d, Month=%s, DayKWH=%f, NightKWH=%f, TotalKWH=%f", userModel.ID, month, dayKwh, nightKwh, totalKwh))
	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"dayKwh":   dayKwh,
			"nightKwh": nightKwh,
			"totalKwh": totalKwh,
		},
	})
}
