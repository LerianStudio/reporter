package report

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
)

// Repository provides an interface for MinIO storage operations
//
//go:generate mockgen --destination=template.mock.go --package=template . Repository
type Repository interface {
	Put(ctx context.Context, objectName string, contentType string, data []byte) error
}

// MinioRepository provides access to a MinIO bucket for file operations.
type MinioRepository struct {
	minioClient *minio.Client
	BucketName  string
}

// NewMinioRepository creates a new instance of MinioRepository with the given client and bucket name.
func NewMinioRepository(minioClient *minio.Client, bucketName string) *MinioRepository {
	return &MinioRepository{
		minioClient: minioClient,
		BucketName:  bucketName,
	}
}

// Put uploads data to the MinIO bucket with the given object name and content type.
func (repo *MinioRepository) Put(ctx context.Context, objectName string, contentType string, data []byte) error {
	fileReader := bytes.NewReader(data)
	fileSize := int64(len(data))

	_, err := repo.minioClient.PutObject(ctx, repo.BucketName, objectName, fileReader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return err
	}

	return nil
}
