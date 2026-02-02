package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/LerianStudio/reporter/v4/pkg/seaweedfs"
)

// StorageType defines the type of storage backend
type StorageType string

const (
	// StorageTypeSeaweedFS uses SeaweedFS HTTP API
	StorageTypeSeaweedFS StorageType = "seaweedfs"
	// StorageTypeS3 uses S3-compatible API (AWS S3, MinIO, SeaweedFS S3)
	StorageTypeS3 StorageType = "s3"
)

// Config contains configuration for creating a storage client
type Config struct {
	// Type specifies which storage backend to use (seaweedfs or s3)
	Type StorageType

	// Bucket name for the storage
	Bucket string

	// SeaweedFS specific config (used when Type=seaweedfs)
	SeaweedFSEndpoint string

	// S3 specific config (used when Type=s3)
	S3Endpoint        string
	S3Region          string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3UsePathStyle    bool
	S3DisableSSL      bool
}

// NewStorageClient creates a storage client based on the provided configuration
func NewStorageClient(ctx context.Context, cfg Config) (ObjectStorage, error) {
	switch cfg.Type {
	case StorageTypeSeaweedFS:
		if cfg.SeaweedFSEndpoint == "" {
			return nil, fmt.Errorf("SeaweedFS endpoint is required when storage type is seaweedfs")
		}
		if cfg.Bucket == "" {
			return nil, fmt.Errorf("bucket name is required")
		}

		client := seaweedfs.NewSeaweedFSClient(cfg.SeaweedFSEndpoint)
		return NewSeaweedFSAdapter(client, cfg.Bucket), nil

	case StorageTypeS3:
		if cfg.Bucket == "" {
			return nil, fmt.Errorf("bucket name is required")
		}

		s3Config := S3Config{
			Endpoint:        cfg.S3Endpoint,
			Region:          cfg.S3Region,
			Bucket:          cfg.Bucket,
			AccessKeyID:     cfg.S3AccessKeyID,
			SecretAccessKey: cfg.S3SecretAccessKey,
			UsePathStyle:    cfg.S3UsePathStyle,
			DisableSSL:      cfg.S3DisableSSL,
		}

		return NewS3Client(ctx, s3Config)

	default:
		return nil, fmt.Errorf("unsupported storage type: %s (supported: seaweedfs, s3)", cfg.Type)
	}
}

// ParseStorageType converts a string to StorageType
func ParseStorageType(s string) (StorageType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "seaweedfs", "seaweed", "":
		return StorageTypeSeaweedFS, nil
	case "s3":
		return StorageTypeS3, nil
	default:
		return "", fmt.Errorf("invalid storage type: %s (valid options: seaweedfs, s3)", s)
	}
}
