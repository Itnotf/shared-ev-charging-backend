package controllers

import (
	"fmt"
	"net/http"
	"shared-charge/models"
	"shared-charge/service"
	"shared-charge/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateReservationRequest 创建预约请求
type CreateReservationRequest struct {
	Date     string `json:"date" binding:"required"`
	Timeslot string `json:"timeslot" binding:"required,oneof=day night"`
	Remark   string `json:"remark"`
}

// GetReservations 获取预约列表
func GetReservations(c *gin.Context) {
	lg := utils.CtxLogger(c)
	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	lg.Info(fmt.Sprintf("获取预约列表: date=%s", date))

	reservations, err := service.GetReservations(date)
	if err != nil {
		lg.Error(fmt.Sprintf("获取预约列表失败: date=%s, Error=%v", date, err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取预约列表失败", "error": err.Error()})
		return
	}

	formatted := make([]map[string]interface{}, len(reservations))
	for i, res := range reservations {
		formatted[i] = res.FormatReservationInfo()
	}

	lg.Info(fmt.Sprintf("获取预约列表成功: count=%d", len(formatted)))
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取预约列表成功", "data": formatted})
}

// CreateReservation 创建预约
func CreateReservation(c *gin.Context) {
	lg := utils.CtxLogger(c)
	user, exists := c.Get("user")
	if !exists {
		lg.Error("创建预约失败: 用户未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	userModel, ok := user.(models.User)
	if !ok {
		lg.Error("创建预约失败: 用户信息类型错误")
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}

	var req CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		lg.Error(fmt.Sprintf("创建预约失败: 请求参数错误: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		lg.Error(fmt.Sprintf("创建预约失败: 日期格式错误: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "日期格式错误", "error": err.Error()})
		return
	}

	lg.Info(fmt.Sprintf("创建预约: UserID=%d, Date=%s, Timeslot=%s", userModel.ID, req.Date, req.Timeslot))

	reservation, err := service.CreateReservationWithCheck(userModel.ID, date, req.Timeslot, req.Remark)
	if err != nil {
		lg.Error(fmt.Sprintf("创建预约失败: UserID=%d, Error=%v", userModel.ID, err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "创建预约失败", "error": err.Error()})
		return
	}

	lg.Info(fmt.Sprintf("预约创建成功: UserID=%d, ReservationID=%d", userModel.ID, reservation.ID))
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "预约创建成功", "data": reservation.FormatReservationInfo()})
}

// DeleteReservation 取消预约
func DeleteReservation(c *gin.Context) {
	lg := utils.CtxLogger(c)
	user, exists := c.Get("user")
	if !exists {
		lg.Error("取消预约失败: 用户未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	userModel, ok := user.(models.User)
	if !ok {
		lg.Error("取消预约失败: 用户信息类型错误")
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		lg.Error(fmt.Sprintf("取消预约失败: ID格式错误: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "ID格式错误", "error": err.Error()})
		return
	}

	lg.Info(fmt.Sprintf("取消预约: UserID=%d, ReservationID=%d", userModel.ID, id))

	if err := service.CancelReservation(uint(id), userModel.ID); err != nil {
		lg.Error(fmt.Sprintf("取消预约失败: UserID=%d, ReservationID=%d, Error=%v", userModel.ID, id, err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "取消预约失败", "error": err.Error()})
		return
	}

	lg.Info(fmt.Sprintf("预约取消成功: UserID=%d, ReservationID=%d", userModel.ID, id))
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "预约取消成功"})
}

// GetCurrentReservation 获取当前用户的最近未过期预约
func GetCurrentReservation(c *gin.Context) {
	lg := utils.CtxLogger(c)
	user, exists := c.Get("user")
	if !exists {
		lg.Error("获取当前预约失败: 用户未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	userModel, ok := user.(models.User)
	if !ok {
		lg.Error("获取当前预约失败: 用户信息类型错误")
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}

	lg.Info(fmt.Sprintf("获取当前预约: UserID=%d", userModel.ID))

	reservation, err := service.GetCurrentReservation(userModel.ID)
	if err != nil {
		lg.Info(fmt.Sprintf("用户无当前预约: UserID=%d", userModel.ID))
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "无当前预约", "data": nil})
		return
	}

	lg.Info(fmt.Sprintf("获取当前预约成功: UserID=%d, ReservationID=%d", userModel.ID, reservation.ID))
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取当前预约成功", "data": reservation.FormatReservationInfo()})
}

// GetCurrentStatus 获取当前预约及充电记录状态
func GetCurrentStatus(c *gin.Context) {
	lg := utils.CtxLogger(c)
	user, exists := c.Get("user")
	if !exists {
		lg.Error("获取当前状态失败: 用户未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	userModel, ok := user.(models.User)
	if !ok {
		lg.Error("获取当前状态失败: 用户信息类型错误")
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}

	lg.Info(fmt.Sprintf("获取当前状态: UserID=%d", userModel.ID))

	status, err := service.GetCurrentReservationStatus(userModel.ID)
	if err != nil {
		lg.Error(fmt.Sprintf("获取当前状态失败: UserID=%d, Error=%v", userModel.ID, err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取当前状态失败", "error": err.Error()})
		return
	}

	lg.Info(fmt.Sprintf("获取当前状态成功: UserID=%d", userModel.ID))
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取当前状态成功", "data": status})
}
