package storage

import (
	"fmt"

	"github.com/LerianStudio/reporter/v4/pkg/constant"
	s3storage "github.com/LerianStudio/reporter/v4/pkg/storage/s3"
	reportS3 "github.com/LerianStudio/reporter/v4/pkg/storage/s3/report"
	templateS3 "github.com/LerianStudio/reporter/v4/pkg/storage/s3/template"
	"github.com/LerianStudio/reporter/v4/pkg/storage/seaweedfs"
	reportSeaweedFS "github.com/LerianStudio/reporter/v4/pkg/storage/seaweedfs/report"
	templateSeaweedFS "github.com/LerianStudio/reporter/v4/pkg/storage/seaweedfs/template"
)

// CreateTemplateRepository creates a template repository based on the configured storage provider
// Returns a generic TemplateRepository interface that works with any storage backend
func CreateTemplateRepository(config *Config) (TemplateRepository, error) {
	// Default to SeaweedFS if no provider specified (backward compatibility)
	if config.Provider == "" || config.Provider == "seaweedfs" {
		return createSeaweedFSTemplateRepository(config)
	}

	if config.Provider == "s3" {
		return createS3TemplateRepository(config)
	}

	return nil, fmt.Errorf("unsupported storage provider: %s", config.Provider)
}

// CreateReportRepository creates a report repository based on the configured storage provider
// Returns a generic ReportRepository interface that works with any storage backend
func CreateReportRepository(config *Config) (ReportRepository, error) {
	// Default to SeaweedFS if no provider specified (backward compatibility)
	if config.Provider == "" || config.Provider == "seaweedfs" {
		return createSeaweedFSReportRepository(config)
	}

	if config.Provider == "s3" {
		return createS3ReportRepository(config)
	}

	return nil, fmt.Errorf("unsupported storage provider: %s", config.Provider)
}

// createSeaweedFSTemplateRepository creates a SeaweedFS template repository
func createSeaweedFSTemplateRepository(config *Config) (TemplateRepository, error) {
	seaweedFSEndpoint := fmt.Sprintf("http://%s:%s", config.SeaweedFSHost, config.SeaweedFSFilerPort)
	seaweedFSClient := seaweedfs.NewSeaweedFSClient(seaweedFSEndpoint)

	return templateSeaweedFS.NewSimpleRepository(seaweedFSClient, constant.TemplateBucketName), nil
}

// createS3TemplateRepository creates an S3 template repository
func createS3TemplateRepository(config *Config) (TemplateRepository, error) {
	s3Config := &s3storage.S3Config{
		Region:          config.S3Region,
		Bucket:          config.S3Bucket,
		AccessKeyID:     config.S3AccessKeyID,
		SecretAccessKey: config.S3SecretAccessKey,
		Endpoint:        config.S3Endpoint,
		ForcePathStyle:  config.S3ForcePathStyle,
	}

	bucket := config.S3TemplateBucket
	if bucket == "" {
		bucket = config.S3Bucket
	}

	s3Client, err := s3storage.NewS3Client(s3Config, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	return templateS3.NewSimpleRepository(s3Client), nil
}

// createSeaweedFSReportRepository creates a SeaweedFS report repository
func createSeaweedFSReportRepository(config *Config) (ReportRepository, error) {
	seaweedFSEndpoint := fmt.Sprintf("http://%s:%s", config.SeaweedFSHost, config.SeaweedFSFilerPort)
	seaweedFSClient := seaweedfs.NewSeaweedFSClient(seaweedFSEndpoint)

	return reportSeaweedFS.NewSimpleRepository(seaweedFSClient, constant.ReportBucketName), nil
}

// createS3ReportRepository creates an S3 report repository
func createS3ReportRepository(config *Config) (ReportRepository, error) {
	s3Config := &s3storage.S3Config{
		Region:          config.S3Region,
		Bucket:          config.S3Bucket,
		AccessKeyID:     config.S3AccessKeyID,
		SecretAccessKey: config.S3SecretAccessKey,
		Endpoint:        config.S3Endpoint,
		ForcePathStyle:  config.S3ForcePathStyle,
	}

	bucket := config.S3ReportBucket
	if bucket == "" {
		bucket = config.S3Bucket
	}

	s3Client, err := s3storage.NewS3Client(s3Config, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	return reportS3.NewSimpleRepository(s3Client), nil
}
