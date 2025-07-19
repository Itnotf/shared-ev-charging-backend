package controllers

import (
	"net/http"
	"shared-charge/models"
	"shared-charge/service"
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
	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	reservations, err := service.GetReservations(date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取预约列表失败", "error": err.Error()})
		return
	}

	formatted := make([]map[string]interface{}, len(reservations))
	for i, res := range reservations {
		formatted[i] = res.FormatReservationInfo()
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取预约列表成功", "data": formatted})
}

// CreateReservation 创建预约
func CreateReservation(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	userModel, ok := user.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}

	var req CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "日期格式错误", "error": err.Error()})
		return
	}

	reservation, err := service.CreateReservationWithCheck(userModel.ID, date, req.Timeslot, req.Remark)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "创建预约失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "预约创建成功", "data": reservation.FormatReservationInfo()})
}

// DeleteReservation 取消预约
func DeleteReservation(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	userModel, ok := user.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "ID格式错误", "error": err.Error()})
		return
	}

	if err := service.CancelReservation(uint(id), userModel.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "取消预约失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "预约取消成功"})
}

// GetCurrentReservation 获取当前用户的最近未过期预约
func GetCurrentReservation(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	userModel, ok := user.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}

	reservation, err := service.GetCurrentReservation(userModel.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "无当前预约", "data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取当前预约成功", "data": reservation.FormatReservationInfo()})
}

// GetCurrentStatus 获取当前预约及充电记录状态
func GetCurrentStatus(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	userModel, ok := user.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}

	status, err := service.GetCurrentReservationStatus(userModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取当前状态失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取当前状态成功", "data": status})
}
