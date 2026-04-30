package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"

	"code-repo/pkg/utils"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var MinioClient *minio.Client

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

	ctx := context.Background()
	ensureBucketExists(ctx, cfg.BucketName)
	ensureBucketExists(ctx, cfg.TempBucket)
}

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

func UploadChunk(ctx context.Context, uploadID string, chunkIndex int, data []byte) error {
	objectKey := fmt.Sprintf("chunks/%s/%d", uploadID, chunkIndex)
	_, err := MinioClient.PutObject(ctx, utils.Config.Minio.TempBucket, objectKey,
		bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
	return err
}

func UploadFile(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	_, err := MinioClient.PutObject(ctx, utils.Config.Minio.BucketName, objectKey,
		reader, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

func UploadFileFromMultipart(ctx context.Context, objectKey string, file multipart.File, header *multipart.FileHeader) error {
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err := MinioClient.PutObject(ctx, utils.Config.Minio.BucketName, objectKey,
		file, header.Size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

func DownloadFile(ctx context.Context, objectKey string) (*minio.Object, error) {
	return MinioClient.GetObject(ctx, utils.Config.Minio.BucketName, objectKey, minio.GetObjectOptions{})
}

func GetFileInfo(ctx context.Context, objectKey string) (minio.ObjectInfo, error) {
	return MinioClient.StatObject(ctx, utils.Config.Minio.BucketName, objectKey, minio.StatObjectOptions{})
}

func DeleteFile(ctx context.Context, objectKey string) error {
	return MinioClient.RemoveObject(ctx, utils.Config.Minio.BucketName, objectKey, minio.RemoveObjectOptions{})
}

func DeleteChunks(ctx context.Context, uploadID string, totalChunks int) error {
	for i := 0; i < totalChunks; i++ {
		objectKey := fmt.Sprintf("chunks/%s/%d", uploadID, i)
		_ = MinioClient.RemoveObject(ctx, utils.Config.Minio.TempBucket, objectKey, minio.RemoveObjectOptions{})
	}
	return nil
}

func MergeChunks(ctx context.Context, uploadID string, totalChunks int, destObjectKey string) error {
	parts := make([]minio.CopySrcOptions, 0, totalChunks)
	for i := 0; i < totalChunks; i++ {
		srcKey := fmt.Sprintf("chunks/%s/%d", uploadID, i)
		parts = append(parts, minio.CopySrcOptions{
			Bucket: utils.Config.Minio.TempBucket,
			Object: srcKey,
		})
	}

	dstOpts := minio.CopyDestOptions{
		Bucket: utils.Config.Minio.BucketName,
		Object: destObjectKey,
	}

	_, err := MinioClient.ComposeObject(ctx, dstOpts, parts...)
	return err
}

func BuildObjectKey(repoID uint64, filePath string) string {
	cleanPath := strings.TrimPrefix(filepath.Clean(filePath), "/")
	return fmt.Sprintf("repos/%d/%s", repoID, cleanPath)
}
