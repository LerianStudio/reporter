package report

import (
	"context"
	"fmt"
	"plugin-smart-templates/v2/pkg/seaweedfs"
)

// Repository provides an interface for SeaweedFS storage operations
//
//go:generate mockgen --destination=report.mock.go --package=report . Repository
type Repository interface {
	Put(ctx context.Context, objectName string, contentType string, data []byte) error
	Get(ctx context.Context, objectName string) ([]byte, error)
}

// SimpleRepository provides access to SeaweedFS storage for file operations using direct HTTP.
type SimpleRepository struct {
	client *seaweedfs.SeaweedFSClient
	bucket string
}

// NewSimpleRepository creates a new instance of SimpleRepository with the given HTTP client and bucket name.
func NewSimpleRepository(client *seaweedfs.SeaweedFSClient, bucket string) *SimpleRepository {
	return &SimpleRepository{
		client: client,
		bucket: bucket,
	}
}

// Put uploads data to the SeaweedFS storage with the given object name and content type.
func (repo *SimpleRepository) Put(ctx context.Context, objectName string, contentType string, data []byte) error {
	path := fmt.Sprintf("/%s/%s", repo.bucket, objectName)

	err := repo.client.UploadFile(ctx, path, data)
	if err != nil {
		return fmt.Errorf("failed to put report to SeaweedFS: %w", err)
	}

	return nil
}

// Get download data from SeaweedFS storage with the given object name
func (repo *SimpleRepository) Get(ctx context.Context, objectName string) ([]byte, error) {
	path := fmt.Sprintf("/%s/%s", repo.bucket, objectName)

	data, err := repo.client.DownloadFile(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get report from SeaweedFS: %w", err)
	}

	return data, nil
}
