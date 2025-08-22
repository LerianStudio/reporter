package template

import (
	"bytes"
	"context"
	"io"
	"plugin-smart-templates/v2/pkg"

	"github.com/minio/minio-go/v7"
)

// Repository provides an interface for MinIO storage operations
//
//go:generate mockgen --destination=template.minio.mock.go --package=template . Repository
type Repository interface {
	Get(ctx context.Context, objectName string) ([]byte, error)
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

// Get the content of a .tpl file from the MinIO bucket.
func (repo *MinioRepository) Get(ctx context.Context, objectName string) ([]byte, error) {
	file, err := repo.minioClient.GetObject(ctx, repo.BucketName, objectName+".tpl", minio.GetObjectOptions{})
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

// Put uploads data to the MinIO bucket with the given object name and content type.
func (repo *MinioRepository) Put(ctx context.Context, objectName string, contentType string, data []byte) error {
	fileReader := bytes.NewReader(data)
	fileSize := int64(len(data))
	logger := pkg.NewLoggerFromContext(ctx)

	_, err := repo.minioClient.PutObject(ctx, repo.BucketName, objectName, fileReader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		logger.Errorf("Erro to comunicate with minio, Err: %v", err)
		return err
	}

	return nil
}
