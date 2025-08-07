package controllers

import (
	"bytes"
	"io"
	"net/http"
	"shared-charge/service"

	"shared-charge/utils"

	"github.com/gin-gonic/gin"
)

// GetAllUsers 管理员获取所有用户列表
func GetAllUsers(c *gin.Context) {
	result, err := service.GetAllUsers(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total": len(result),
		"users": result,
	})
}

// UpdateUserCanReserve 管理员切换用户可预约状态
func UpdateUserCanReserve(c *gin.Context) {
	utils.InfoCtx(c, "管理员切换用户可预约状态请求开始")

	type reqBody struct {
		UserID     uint `json:"user_id" binding:"required"`
		CanReserve bool `json:"can_reserve"`
	}
	var req reqBody

	// 读取请求体
	body, _ := c.GetRawData()
	utils.InfoCtx(c, "请求体内容: %s", string(body))

	// 重新设置请求体，因为 GetRawData 会消费掉
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.WarnCtx(c, "参数校验失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	utils.InfoCtx(c, "参数校验通过: user_id=%d, can_reserve=%t", req.UserID, req.CanReserve)

	if err := service.UpdateUserCanReserve(c, req.UserID, req.CanReserve); err != nil {
		utils.ErrorCtx(c, "更新用户可预约状态失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	utils.InfoCtx(c, "用户可预约状态更新成功: user_id=%d, can_reserve=%t", req.UserID, req.CanReserve)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UpdateUserUnitPrice 管理员修改用户电价
func UpdateUserUnitPrice(c *gin.Context) {
	type reqBody struct {
		UserID    uint    `json:"user_id" binding:"required"`
		UnitPrice float64 `json:"unit_price" binding:"required"`
	}
	var req reqBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if req.UnitPrice <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "电价必须为正数"})
		return
	}
	if err := service.UpdateUserUnitPrice(c, req.UserID, req.UnitPrice); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetMonthlyReport 管理员获取月度对账数据
func GetMonthlyReport(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "月份参数不能为空"})
		return
	}
	result, err := service.GetMonthlyReport(c, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取月度对账失败"})
		return
	}
	c.JSON(http.StatusOK, result)
}
