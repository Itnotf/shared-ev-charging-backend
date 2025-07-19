package controllers

import (
	"net/http"
	"shared-charge/config"
	"shared-charge/service"
	"shared-charge/utils"

	"github.com/gin-gonic/gin"
)

// UploadImage 上传图片
// @Summary 上传图片
// @Description 用户上传图片文件
// @Tags 文件
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "图片文件"
// @Success 200 {object} gin.H{"code":200,"message":"图片上传成功","data":{}}
// @Router /upload/image [post]
func UploadImage(c *gin.Context) {
	userModel, ok := utils.GetUserFromContext(c)
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
