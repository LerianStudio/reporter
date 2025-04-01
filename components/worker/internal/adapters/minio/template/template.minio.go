package template

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
	"io"
)

// Repository provides an interface for MinIO storage operations
//
//go:generate mockgen --destination=template.mock.go --package=template . Repository
type Repository interface {
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

// Get retrieves the content of a .txt file from the MinIO bucket.
func (repo *MinioRepository) Get(ctx context.Context, objectName string) ([]byte, error) {
	file, err := repo.minioClient.GetObject(ctx, repo.BucketName, objectName+".txt", minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
