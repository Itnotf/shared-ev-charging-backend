package controllers

import (
	"errors"
	"net/http"
	"shared-charge/config"
	"shared-charge/models"
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

func WechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误", "error": err.Error()})
		return
	}
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
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "微信登录失败", "error": err.Error()})
		return
	}
	if authResult.ErrCode != 0 {
		c.Error(errors.New(authResult.ErrMsg))
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
		userToCreate := service.UserCreateInput{
			OpenID: authResult.OpenID,
			Name:   "用户" + authResult.OpenID[len(authResult.OpenID)-6:],
			Phone:  phone,
			Role:   "user",
			Status: "active",
		}
		user, err = service.CreateUserWithInput(userToCreate)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建用户失败", "error": err.Error()})
			return
		}
	} else if phone != "" && user.Phone == "" {
		// 如果用户已存在但没有手机号，且本次获取到了手机号，则更新手机号
		user.Phone = phone
		models.DB.Save(&user)
	}
	formattedUser := user.FormatUserInfo()
	token, err := utils.GenerateToken(formattedUser)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成令牌失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "登录成功", "data": WechatLoginResponse{Token: token, UserInfo: formattedUser}})
}

func RefreshToken(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}
	userModel, ok := user.(models.User)
	if !ok {
		c.Error(errors.New("用户信息类型错误"))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}
	userData, err := service.GetUserByID(userModel.ID)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户信息失败", "error": err.Error()})
		return
	}
	formattedUser := userData.FormatUserInfo()
	token, err := utils.GenerateToken(formattedUser)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成令牌失败", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "令牌刷新成功", "data": gin.H{"token": token, "user_info": formattedUser}})
}
