package controllers

import (
	"net/http"
	"shared-charge/service"
	"strconv"
	"time"

	"shared-charge/utils"

	"github.com/gin-gonic/gin"
)

// CreateReservationRequest 创建预约请求
type CreateReservationRequest struct {
	Date     string `json:"date" binding:"required"`
	Timeslot string `json:"timeslot" binding:"required,oneof=day night"`
	Remark   string `json:"remark"`
}

// GetReservations 获取预约列表
// @Summary 获取预约列表
// @Description 获取指定日期的所有预约（不传date则为当天）
// @Tags 预约
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param date query string false "预约日期(YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Router /reservations [get]
func GetReservations(c *gin.Context) {
	utils.InfoCtx(c, "获取预约列表请求: date=%s", c.Query("date"))
	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	reservations, err := service.GetReservations(c, date)
	if err != nil {
		utils.ErrorCtx(c, "获取预约列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取预约列表失败"})
		return
	}
	utils.InfoCtx(c, "获取预约列表成功: count=%d", len(reservations))
	formatted := make([]map[string]interface{}, len(reservations))
	for i, res := range reservations {
		formatted[i] = res.FormatReservationInfo()
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取预约列表成功", "data": formatted})
}

// CreateReservation 创建预约
// @Summary 创建预约
// @Description 创建新的预约
// @Tags 预约
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateReservationRequest true "预约请求体"
// @Success 200 {object} map[string]interface{}
// @Router /reservations [post]
func CreateReservation(c *gin.Context) {
	utils.InfoCtx(c, "创建预约请求")
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		utils.WarnCtx(c, "创建预约未认证")
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}
	var req CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.WarnCtx(c, "创建预约参数校验失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}
	date, err := utils.ParseDate(req.Date)
	if err != nil {
		utils.WarnCtx(c, "创建预约日期格式错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "日期格式错误", "error": err.Error()})
		return
	}
	reservation, err := service.CreateReservationWithCheck(c, userModel.ID, date, req.Timeslot, req.Remark)
	if err != nil {
		utils.ErrorCtx(c, "创建预约失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "创建预约失败"})
		return
	}
	utils.InfoCtx(c, "预约创建成功: user_id=%d, reservation_id=%d", userModel.ID, reservation.ID)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "预约创建成功", "data": reservation.FormatReservationInfo()})
}

// DeleteReservation 取消预约
// @Summary 取消预约
// @Description 取消指定ID的预约
// @Tags 预约
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "预约ID"
// @Success 200 {object} map[string]interface{}
// @Router /reservations/{id} [delete]
func DeleteReservation(c *gin.Context) {
	utils.InfoCtx(c, "取消预约请求: id=%s", c.Param("id"))
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		utils.WarnCtx(c, "取消预约未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WarnCtx(c, "取消预约ID格式错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "ID格式错误", "error": err.Error()})
		return
	}
	if err := service.CancelReservation(c, uint(id), userModel.ID); err != nil {
		utils.ErrorCtx(c, "取消预约失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "取消预约失败"})
		return
	}
	utils.InfoCtx(c, "预约取消成功: user_id=%d, reservation_id=%d", userModel.ID, id)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "预约取消成功"})
}

// GetCurrentReservation 获取当前用户的最近未过期预约
// @Summary 获取当前用户的最近未过期预约
// @Description 获取当前用户的最近未过期预约信息
// @Tags 预约
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /reservations/current [get]
func GetCurrentReservation(c *gin.Context) {
	utils.InfoCtx(c, "获取当前预约请求")
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		utils.WarnCtx(c, "获取当前预约未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}
	reservation, err := service.GetCurrentReservation(c, userModel.ID)
	if err != nil {
		utils.InfoCtx(c, "用户无当前预约: user_id=%d", userModel.ID)
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "无当前预约", "data": nil})
		return
	}
	utils.InfoCtx(c, "获取当前预约成功: user_id=%d, reservation_id=%d", userModel.ID, reservation.ID)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取当前预约成功", "data": reservation.FormatReservationInfo()})
}

// GetCurrentStatus 获取当前预约及充电记录状态
// @Summary 获取当前预约及充电记录状态
// @Description 获取当前用户的预约和充电记录状态
// @Tags 预约
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /reservations/current-status [get]
func GetCurrentStatus(c *gin.Context) {
	utils.InfoCtx(c, "获取当前状态请求")
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		utils.WarnCtx(c, "获取当前状态未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}
	status, err := service.GetCurrentReservationStatus(c, userModel.ID)
	if err != nil {
		utils.ErrorCtx(c, "获取当前状态失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取当前状态失败"})
		return
	}
	utils.InfoCtx(c, "获取当前状态成功: user_id=%d", userModel.ID)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取当前状态成功", "data": status})
}
