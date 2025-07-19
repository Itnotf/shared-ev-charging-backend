package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"shared-charge/config"
	"shared-charge/utils"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/minio/minio-go/v7"
)

// 当前上传图片仅涉及文件存储，若后续需记录上传日志可在此扩展数据库操作。

func SaveUploadImage(c *gin.Context, userID uint, file *multipart.FileHeader) (map[string]interface{}, error) {
	cfg := config.GetConfig()
	utils.InfoCtx(c, "用户上传图片: user_id=%d, filename=%s, size=%d", userID, file.Filename, file.Size)
	// 检查文件大小
	if file.Size > cfg.App.MaxFileSize {
		utils.WarnCtx(c, "文件大小超过限制: %d > %d", file.Size, cfg.App.MaxFileSize)
		return nil, fmt.Errorf("文件大小超过限制")
	}
	// 检查文件类型
	ext := filepath.Ext(file.Filename)
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif"}
	isAllowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		utils.WarnCtx(c, "不支持的文件类型: %s", ext)
		return nil, fmt.Errorf("不支持的文件类型")
	}
	// 生成对象名称
	objectName := fmt.Sprintf("%d_%s_%s", userID, time.Now().Format("20060102150405"), file.Filename)
	// 打开文件
	src, err := file.Open()
	if err != nil {
		utils.ErrorCtx(c, "打开文件失败: %v", err)
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer src.Close()
	// 上传到MinIO
	minioClient := utils.GetMinioClient()
	_, err = minioClient.PutObject(context.Background(), cfg.MinIO.BucketName, objectName, src, file.Size, minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")})
	if err != nil {
		utils.ErrorCtx(c, "上传到MinIO失败: %v", err)
		return nil, fmt.Errorf("上传到MinIO失败: %v", err)
	}
	// 生成公共URL
	fileURL := fmt.Sprintf("http://%s/%s/%s", cfg.MinIO.Endpoint, cfg.MinIO.BucketName, objectName)
	utils.InfoCtx(c, "图片上传成功: user_id=%d, filename=%s", userID, objectName)
	return map[string]interface{}{
		"url":      fileURL,
		"filename": objectName,
		"size":     file.Size,
	}, nil
}
