package controllers

import (
	"net/http"
	"shared-charge/config"
	"shared-charge/service"
	"shared-charge/utils"

	"github.com/gin-gonic/gin"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	miniConfig "github.com/silenceper/wechat/v2/miniprogram/config"
)

type WechatLoginRequest struct {
	Code      string `json:"code" binding:"required"`
	PhoneCode string `json:"phoneCode"`
}

type WechatLoginResponse struct {
	Token    string      `json:"token"`
	UserInfo interface{} `json:"user_info"`
}

// WechatLogin 微信登录
// @Summary 微信小程序登录
// @Description 微信小程序登录，获取token和用户信息
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body WechatLoginRequest true "微信登录请求体"
// @Success 200 {object} controllers.WechatLoginResponse
// @Router /auth/login [post]
func WechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.WarnCtx(c, "微信登录参数校验失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}
	utils.InfoCtx(c, "微信登录请求: code=%s, hasPhoneCode=%t", req.Code, req.PhoneCode != "")
	cfg := config.GetConfig()
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	miniprogram := wc.GetMiniProgram(&miniConfig.Config{
		AppID:     cfg.Wechat.AppID,
		AppSecret: cfg.Wechat.Secret,
		Cache:     memory,
	})
	authResult, err := miniprogram.GetAuth().Code2Session(req.Code)
	if err != nil {
		utils.ErrorCtx(c, "微信登录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "微信登录失败"})
		return
	}
	if authResult.ErrCode != 0 {
		utils.WarnCtx(c, "微信登录API错误: %s", authResult.ErrMsg)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "微信登录失败", "error": authResult.ErrMsg})
		return
	}

	// 获取手机号
	var phone string
	if req.PhoneCode != "" {
		phoneResult, err := miniprogram.GetAuth().GetPhoneNumber(req.PhoneCode)
		if err != nil {
		} else if phoneResult.ErrCode == 0 {
			phone = phoneResult.PhoneInfo.PhoneNumber
		} else {
		}
	}
	user, err := service.GetUserByOpenID(authResult.OpenID)
	if err != nil {
		utils.InfoCtx(c, "创建新用户: openid=%s", authResult.OpenID)
		userToCreate := service.UserCreateInput{
			OpenID: authResult.OpenID,
			Name:   "",
			Phone:  phone,
			Role:   "user",
			Status: "active",
		}
		user, err = service.CreateUserWithInput(userToCreate)
		if err != nil {
			utils.ErrorCtx(c, "创建用户失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建用户失败"})
			return
		}
		utils.InfoCtx(c, "新用户创建成功: user_id=%d", user.ID)
	} else if phone != "" && user.Phone == "" {
		utils.InfoCtx(c, "更新用户手机号: user_id=%d", user.ID)
		// 如果用户已存在但没有手机号，且本次获取到了手机号，则更新手机号
		err := service.UpdateUserPhoneByID(c, user.ID, phone)
		if err != nil {
			utils.ErrorCtx(c, "更新用户手机号失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新用户手机号失败", "error": err.Error()})
			return
		}
	}
	formattedUser := user.FormatUserInfo()
	token, err := utils.GenerateToken(formattedUser)
	if err != nil {
		utils.ErrorCtx(c, "生成令牌失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成令牌失败", "error": err.Error()})
		return
	}
	utils.InfoCtx(c, "用户登录成功: user_id=%d", user.ID)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "登录成功", "data": WechatLoginResponse{Token: token, UserInfo: formattedUser}})
}

// RefreshToken 刷新令牌
// @Summary 刷新令牌
// @Description 刷新用户token
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /auth/refresh [post]
func RefreshToken(c *gin.Context) {
	utils.InfoCtx(c, "刷新令牌请求")
	userModel, ok := utils.GetUserFromContext(c)
	if !ok {
		utils.WarnCtx(c, "刷新令牌未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}
	userData, err := service.GetUserByID(c, userModel.ID)
	if err != nil {
		utils.ErrorCtx(c, "刷新令牌获取用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败", "error": err.Error()})
		return
	}
	formattedUser := userData.FormatUserInfo()
	token, err := utils.GenerateToken(formattedUser)
	if err != nil {
		utils.ErrorCtx(c, "刷新令牌生成失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成令牌失败", "error": err.Error()})
		return
	}
	utils.InfoCtx(c, "令牌刷新成功: user_id=%d", userData.ID)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "令牌刷新成功", "data": gin.H{"token": token, "user_info": formattedUser}})
}
