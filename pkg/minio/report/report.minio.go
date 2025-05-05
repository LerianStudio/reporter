package report

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
)

// Repository provides an interface for MinIO storage operations
//
//go:generate mockgen --destination=report.minio.mock.go --package=report . Repository
type Repository interface {
	Put(ctx context.Context, objectName string, contentType string, data []byte) error
	Get(ctx context.Context, objectName string) ([]byte, error)
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

// Get download data of MinIO bucket with the given object name
func (repo *MinioRepository) Get(ctx context.Context, objectName string) ([]byte, error) {
	obj, err := repo.minioClient.GetObject(ctx, repo.BucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(obj); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
