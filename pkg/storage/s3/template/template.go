package template

import (
	"context"
	"fmt"
	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/constant"
)

// Repository provides an interface for S3 template storage operations
// This interface matches exactly with pkg/seaweedfs/template.Repository
//
//go:generate mockgen --destination=template.mock.go --package=template . Repository
type Repository interface {
	Get(ctx context.Context, objectName string) ([]byte, error)
	Put(ctx context.Context, objectName string, contentType string, data []byte) error
}

// S3Client interface defines the methods needed from the S3 client
type S3Client interface {
	UploadFile(ctx context.Context, path string, data []byte) error
	DownloadFile(ctx context.Context, path string) ([]byte, error)
}

// SimpleRepository provides S3 storage operations for templates using AWS SDK
type SimpleRepository struct {
	client S3Client
}

// NewSimpleRepository creates a new S3 template repository instance
func NewSimpleRepository(client S3Client) *SimpleRepository {
	return &SimpleRepository{
		client: client,
	}
}

// Get retrieves a template file from S3 storage
// Automatically adds .tpl extension for templates (same behavior as SeaweedFS)
func (repo *SimpleRepository) Get(ctx context.Context, objectName string) ([]byte, error) {
	// Add .tpl extension for templates (maintaining compatibility with SeaweedFS behavior)
	path := fmt.Sprintf("%s.tpl", objectName)

	data, err := repo.client.DownloadFile(ctx, path)
	if err != nil {
		return nil, pkg.ValidateBusinessError(constant.ErrCommunicateSeaweedFS, "")
	}

	return data, nil
}

// Put uploads a template file to S3 storage
func (repo *SimpleRepository) Put(ctx context.Context, objectName string, contentType string, data []byte) error {
	logger := pkg.NewLoggerFromContext(ctx)

	err := repo.client.UploadFile(ctx, objectName, data)
	if err != nil {
		logger.Errorf("Error communicating with S3: %v", err)
		return pkg.ValidateBusinessError(constant.ErrCommunicateSeaweedFS, "")
	}

	return nil
}
