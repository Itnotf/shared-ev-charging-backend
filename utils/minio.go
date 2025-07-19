package utils

import (
	"context"
	"errors"
	"shared-charge/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client

// 初始化MinIO客户端
func InitMinioClient() error {
	cfg := config.GetConfig()

	// 创建MinIO客户端
	var err error
	minioClient, err = minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		return err
	}

	// 检查bucket是否存在
	exists, err := minioClient.BucketExists(context.Background(), cfg.MinIO.BucketName)
	if err != nil {
		return err
	}

	// 如果bucket不存在，创建它
	if !exists {
		err = minioClient.MakeBucket(context.Background(), cfg.MinIO.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}

		// 设置bucket策略为公共读取
		policy := `{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": "*"},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::` + cfg.MinIO.BucketName + `/*"]
				}
			]
		}`
		err = minioClient.SetBucketPolicy(context.Background(), cfg.MinIO.BucketName, policy)
		if err != nil {
			return err
		}
	}

	return nil
}

// 获取MinIO客户端实例
func GetMinioClient() *minio.Client {
	return minioClient
}

// TestMinIOConnection 测试MinIO连接
func TestMinIOConnection() error {
	if minioClient == nil {
		return errors.New("MinIO客户端未初始化")
	}

	// 尝试列出buckets来测试连接
	_, err := minioClient.ListBuckets(context.Background())
	return err
}
