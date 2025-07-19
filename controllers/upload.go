package controllers

import (
	"net/http"
	"shared-charge/config"
	"shared-charge/models"
	"shared-charge/service"

	"github.com/gin-gonic/gin"
)

func UploadImage(c *gin.Context) {
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

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "获取文件失败", "error": err.Error()})
		return
	}

	// 添加文件大小预检查
	maxFileSize := config.GetConfig().App.MaxFileSize
	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "文件大小超过限制"})
		return
	}

	result, err := service.SaveUploadImage(userModel.ID, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "上传图片失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "图片上传成功", "data": result})
}
