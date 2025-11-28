package report

import (
	"context"
	"fmt"

	"github.com/LerianStudio/reporter/v4/pkg"
)

// Repository provides an interface for S3 report storage operations
// This interface matches exactly with pkg/seaweedfs/report.Repository
//
//go:generate mockgen --destination=report.mock.go --package=report . Repository
type Repository interface {
	Put(ctx context.Context, objectName string, contentType string, data []byte, ttl string) error
	Get(ctx context.Context, objectName string) ([]byte, error)
}

// S3Client interface defines the methods needed from the S3 client
type S3Client interface {
	UploadFileWithTTL(ctx context.Context, path string, data []byte, ttl string) error
	UploadFileWithContentType(ctx context.Context, path string, data []byte, contentType, ttl string) error
	DownloadFile(ctx context.Context, path string) ([]byte, error)
}

// SimpleRepository provides S3 storage operations for reports using AWS SDK
type SimpleRepository struct {
	client S3Client
}

// NewSimpleRepository creates a new S3 report repository instance
func NewSimpleRepository(client S3Client) *SimpleRepository {
	return &SimpleRepository{
		client: client,
	}
}

// Put uploads a report file to S3 storage with optional time-to-live
// TTL is passed as HTTP cache metadata to the S3 client
// Format: 3m (3 minutes), 4h (4 hours), 5d (5 days), 6w (6 weeks), 7M (7 months), 8y (8 years)
// If ttl is empty string, no TTL metadata is applied
func (repo *SimpleRepository) Put(ctx context.Context, objectName string, contentType string, data []byte, ttl string) error {
	logger := pkg.NewLoggerFromContext(ctx)

	// Add reports/ prefix to maintain compatibility with SeaweedFS structure
	path := fmt.Sprintf("reports/%s", objectName)
	err := repo.client.UploadFileWithContentType(ctx, path, data, contentType, ttl)

	if err != nil {
		logger.Errorf("Error communicating with S3: %v", err)
		return fmt.Errorf("upload report to S3: %w", err)
	}

	return nil
}

// Get downloads a report file from S3 storage
func (repo *SimpleRepository) Get(ctx context.Context, objectName string) ([]byte, error) {
	logger := pkg.NewLoggerFromContext(ctx)

	// Add reports/ prefix to maintain compatibility with SeaweedFS structure
	path := fmt.Sprintf("reports/%s", objectName)
	data, err := repo.client.DownloadFile(ctx, path)

	if err != nil {
		logger.Errorf("Error communicating with S3: %v", err)
		return nil, fmt.Errorf("download report from S3: %w", err)
	}

	return data, nil
}
