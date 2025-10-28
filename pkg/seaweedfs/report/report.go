package report

import (
	"context"
	"fmt"

	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/constant"
	"github.com/LerianStudio/reporter/v4/pkg/seaweedfs"
)

// Repository provides an interface for SeaweedFS storage operations
//
//go:generate mockgen --destination=report.mock.go --package=report . Repository
type Repository interface {
	Put(ctx context.Context, objectName string, contentType string, data []byte, ttl string) error
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

// Put uploads data to the SeaweedFS storage with the given object name, content type, and optional TTL.
// TTL format: 3m (3 minutes), 4h (4 hours), 5d (5 days), 6w (6 weeks), 7M (7 months), 8y (8 years)
// If ttl is empty string, no TTL is applied and the file will be stored permanently
func (repo *SimpleRepository) Put(ctx context.Context, objectName string, contentType string, data []byte, ttl string) error {
	path := fmt.Sprintf("/%s/%s", repo.bucket, objectName)

	err := repo.client.UploadFileWithTTL(ctx, path, data, ttl)
	if err != nil {
		return pkg.ValidateBusinessError(constant.ErrCommunicateSeaweedFS, "")
	}

	return nil
}

// Get download data from SeaweedFS storage with the given object name
func (repo *SimpleRepository) Get(ctx context.Context, objectName string) ([]byte, error) {
	path := fmt.Sprintf("/%s/%s", repo.bucket, objectName)

	data, err := repo.client.DownloadFile(ctx, path)
	if err != nil {
		return nil, pkg.ValidateBusinessError(constant.ErrCommunicateSeaweedFS, "")
	}

	return data, nil
}
