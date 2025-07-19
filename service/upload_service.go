package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"shared-charge/config"
	"time"

	"shared-charge/utils"

	"github.com/minio/minio-go/v7"
)

// 当前上传图片仅涉及文件存储，若后续需记录上传日志可在此扩展数据库操作。

func SaveUploadImage(userID uint, file *multipart.FileHeader) (map[string]interface{}, error) {
	cfg := config.GetConfig()
	// 检查文件大小
	if file.Size > cfg.App.MaxFileSize {
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
		return nil, fmt.Errorf("不支持的文件类型")
	}
	// 生成对象名称
	objectName := fmt.Sprintf("%d_%s_%s", userID, time.Now().Format("20060102150405"), file.Filename)

	// 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer src.Close()

	// 上传到MinIO
	minioClient := utils.GetMinioClient()
	_, err = minioClient.PutObject(context.Background(), cfg.MinIO.BucketName, objectName, src, file.Size, minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")})
	if err != nil {
		return nil, fmt.Errorf("上传到MinIO失败: %v", err)
	}

	// 生成公共URL
	fileURL := fmt.Sprintf("http://%s/%s/%s", cfg.MinIO.Endpoint, cfg.MinIO.BucketName, objectName)

	return map[string]interface{}{
		"url":      fileURL,
		"filename": objectName,
		"size":     file.Size,
	}, nil
}
