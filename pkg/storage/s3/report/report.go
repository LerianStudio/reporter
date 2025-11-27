package report

import (
	"context"
	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/constant"
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

// Put uploads a report file to S3 storage with optional TTL
// TTL is implemented using S3 object expiration
// TTL format: 3m (3 minutes), 4h (4 hours), 5d (5 days), 6w (6 weeks), 7M (7 months), 8y (8 years)
// If ttl is empty string, no TTL is applied and the file will be stored permanently
func (repo *SimpleRepository) Put(ctx context.Context, objectName string, contentType string, data []byte, ttl string) error {
	err := repo.client.UploadFileWithTTL(ctx, objectName, data, ttl)
	if err != nil {
		return pkg.ValidateBusinessError(constant.ErrCommunicateSeaweedFS, "")
	}

	return nil
}

// Get downloads a report file from S3 storage
func (repo *SimpleRepository) Get(ctx context.Context, objectName string) ([]byte, error) {
	data, err := repo.client.DownloadFile(ctx, objectName)
	if err != nil {
		return nil, pkg.ValidateBusinessError(constant.ErrCommunicateSeaweedFS, "")
	}

	return data, nil
}
