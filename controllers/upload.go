package controllers

import (
	"fmt"
	"net/http"
	"shared-charge/config"
	"shared-charge/models"
	"shared-charge/service"
	"shared-charge/utils"

	"github.com/gin-gonic/gin"
)

func UploadImage(c *gin.Context) {
	lg := utils.CtxLogger(c)
	user, exists := c.Get("user")
	if !exists {
		lg.Error("上传图片失败: 用户未认证")
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户未认证"})
		return
	}

	userModel, ok := user.(models.User)
	if !ok {
		lg.Error("上传图片失败: 用户信息类型错误")
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户信息类型错误"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		lg.Error(fmt.Sprintf("上传图片失败: 获取文件失败: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "获取文件失败", "error": err.Error()})
		return
	}

	// 添加文件大小预检查
	maxFileSize := config.GetConfig().App.MaxFileSize
	if file.Size > maxFileSize {
		lg.Error(fmt.Sprintf("文件大小超过限制: %d > %d", file.Size, maxFileSize))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "文件大小超过限制"})
		return
	}

	lg.Info(fmt.Sprintf("上传图片: UserID=%d, FileName=%s, FileSize=%d", userModel.ID, file.Filename, file.Size))

	result, err := service.SaveUploadImage(userModel.ID, file)
	if err != nil {
		lg.Error(fmt.Sprintf("上传图片失败: UserID=%d, Error=%v", userModel.ID, err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "上传图片失败", "error": err.Error()})
		return
	}

	lg.Info(fmt.Sprintf("图片上传成功: UserID=%d", userModel.ID))
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "图片上传成功", "data": result})
}
