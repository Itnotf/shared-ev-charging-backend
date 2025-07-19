package controllers

import (
	"net/http"
	"shared-charge/models"
	"shared-charge/service"

	"github.com/gin-gonic/gin"
)

// GetUserProfile 获取用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Router /users/profile [get]
func GetUserProfile(c *gin.Context) {
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

	userData, err := service.GetUserByID(userModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取用户信息成功", "data": userData.FormatUserInfo()})
}

// GetUserPrice 获取当前用户专属电价
// @Summary 获取当前用户专属电价
// @Description 获取当前登录用户的专属电价（如无则返回全局默认）
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /users/price [get]
func GetUserPrice(c *gin.Context) {
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

	userData, err := service.GetUserByID(userModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户电价失败", "error": err.Error()})
		return
	}

	price := userData.UnitPrice
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取用户电价成功", "data": gin.H{"unit_price": price}})
}

// UpdateUserProfile 更新用户信息
// @Summary 更新用户信息
// @Description 更新用户的头像和昵称
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateUserProfileRequest true "用户信息更新请求"
// @Success 200 {object} gin.H{"code": 200, "message": "更新成功"}
// @Router /users/profile [post]
func UpdateUserProfile(c *gin.Context) {
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

	var req struct {
		NickName string `json:"nick_name"`
		Avatar   string `json:"avatar"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}

	if err := service.UpdateUserProfile(userModel.ID, req.Avatar, req.NickName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新用户信息失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "用户信息更新成功"})
}

// UpdateUserProfileRequest 用户信息更新请求
type UpdateUserProfileRequest struct {
	AvatarUrl string `json:"avatarUrl" binding:"required"`
	NickName  string `json:"nickName" binding:"required"`
}
