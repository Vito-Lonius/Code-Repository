package storage

import (
	"context"
	"log"

	"code-repo/pkg/utils"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var MinioClient *minio.Client

// InitMinio 初始化对象存储客户端
func InitMinio(cfg utils.MinioConfig) {
	var err error
	MinioClient, err = minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Fatalf("初始化 MinIO 客户端失败: %v", err)
	}

	log.Println("MinIO 客户端连接成功!")

	// 确保存储桶存在
	ctx := context.Background()
	ensureBucketExists(ctx, cfg.BucketName)
	ensureBucketExists(ctx, cfg.TempBucket)
}

// ensureBucketExists 检查桶是否存在，不存在则创建
func ensureBucketExists(ctx context.Context, bucketName string) {
	exists, err := MinioClient.BucketExists(ctx, bucketName)
	if err != nil {
		log.Fatalf("检查 Bucket [%s] 状态失败: %v", bucketName, err)
	}
	if !exists {
		log.Printf("Bucket [%s] 不存在，正在创建...", bucketName)
		err = MinioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalf("创建 Bucket [%s] 失败: %v", bucketName, err)
		}
		log.Printf("Bucket [%s] 创建成功!", bucketName)
	}
}
