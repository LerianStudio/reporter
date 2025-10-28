package template

import (
	"context"
	"fmt"

	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/constant"
	"github.com/LerianStudio/reporter/v4/pkg/seaweedfs"
)

// Repository provides an interface for SeaweedFS storage operations
//
//go:generate mockgen --destination=template.mock.go --package=template . Repository
type Repository interface {
	Get(ctx context.Context, objectName string) ([]byte, error)
	Put(ctx context.Context, objectName string, contentType string, data []byte) error
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

// Get the content of a .tpl file from the SeaweedFS storage.
func (repo *SimpleRepository) Get(ctx context.Context, objectName string) ([]byte, error) {
	// Add .tpl extension for templates
	path := fmt.Sprintf("/%s/%s.tpl", repo.bucket, objectName)

	data, err := repo.client.DownloadFile(ctx, path)
	if err != nil {
		return nil, pkg.ValidateBusinessError(constant.ErrCommunicateSeaweedFS, "")
	}

	return data, nil
}

// Put uploads data to the SeaweedFS storage with the given object name and content type.
func (repo *SimpleRepository) Put(ctx context.Context, objectName string, contentType string, data []byte) error {
	logger := pkg.NewLoggerFromContext(ctx)

	path := fmt.Sprintf("/%s/%s", repo.bucket, objectName)

	err := repo.client.UploadFile(ctx, path, data)
	if err != nil {
		logger.Errorf("Error communicating with SeaweedFS: %v", err)
		return pkg.ValidateBusinessError(constant.ErrCommunicateSeaweedFS, "")
	}

	return nil
}
