package controllers

import (
	"net/http"
	"shared-charge/config"
	"shared-charge/service"
	"shared-charge/utils"

	"github.com/gin-gonic/gin"
)

// GetUserProfile 获取用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /users/profile [get]
func GetUserProfile(c *gin.Context) {
	utils.InfoCtx(c, "获取用户信息请求")
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		utils.WarnCtx(c, "获取用户信息未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	userData, err := service.GetUserByID(c, userModel.ID)
	if err != nil {
		utils.ErrorCtx(c, "获取用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败"})
		return
	}
	utils.InfoCtx(c, "获取用户信息成功: user_id=%d", userModel.ID)
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
	utils.InfoCtx(c, "获取用户电价请求")
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		utils.WarnCtx(c, "获取用户电价未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}
	unitPrice := userModel.UnitPrice
	if unitPrice <= 0 {
		unitPrice = config.GetConfig().App.DefaultUnitPrice
	}
	utils.InfoCtx(c, "获取用户电价成功: user_id=%d, price=%v", userModel.ID, unitPrice)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取用户电价成功", "data": gin.H{"unit_price": unitPrice}})
}

// UpdateUserProfile 更新用户信息
// @Summary 更新用户信息
// @Description 更新用户的头像和昵称
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateUserProfileRequest true "用户信息更新请求"
// @Success 200 {object} map[string]interface{}
// @Router /users/profile [post]
func UpdateUserProfile(c *gin.Context) {
	utils.InfoCtx(c, "更新用户信息请求")
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		utils.WarnCtx(c, "更新用户信息未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}
	var req struct {
		NickName  string `json:"nickName"`
		AvatarUrl string `json:"avatarUrl"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.WarnCtx(c, "更新用户信息参数校验失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
		return
	}
	if req.NickName == "" && req.AvatarUrl == "" {
		utils.WarnCtx(c, "更新用户信息参数为空")
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "昵称和头像不能都为空"})
		return
	}
	if err := service.UpdateUserProfile(c, userModel.ID, req.AvatarUrl, req.NickName); err != nil {
		utils.ErrorCtx(c, "更新用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新用户信息失败"})
		return
	}
	utils.InfoCtx(c, "用户信息更新成功: user_id=%d", userModel.ID)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "用户信息更新成功"})
}

// UpdateUserProfileRequest 用户信息更新请求
type UpdateUserProfileRequest struct {
	AvatarUrl string `json:"avatarUrl" binding:"required"`
	NickName  string `json:"nickName" binding:"required"`
}
