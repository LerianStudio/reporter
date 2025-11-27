package storage

import (
	"testing"
)

func TestCreateTemplateRepository_SeaweedFS(t *testing.T) {
	config := &Config{
		Provider:           "seaweedfs",
		SeaweedFSHost:      "localhost",
		SeaweedFSFilerPort: "8888",
	}

	repo, err := CreateTemplateRepository(config)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo == nil {
		t.Fatal("expected repository, got nil")
	}
}

func TestCreateTemplateRepository_S3(t *testing.T) {
	config := &Config{
		Provider:          "s3",
		S3Region:          "us-east-1",
		S3Bucket:          "test-bucket",
		S3AccessKeyID:     "test-key",
		S3SecretAccessKey: "test-secret",
	}

	repo, err := CreateTemplateRepository(config)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo == nil {
		t.Fatal("expected repository, got nil")
	}
}

func TestCreateTemplateRepository_DefaultToSeaweedFS(t *testing.T) {
	config := &Config{
		Provider:           "", // Empty should default to SeaweedFS
		SeaweedFSHost:      "localhost",
		SeaweedFSFilerPort: "8888",
	}

	repo, err := CreateTemplateRepository(config)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo == nil {
		t.Fatal("expected repository, got nil")
	}
}

func TestCreateTemplateRepository_UnsupportedProvider(t *testing.T) {
	config := &Config{
		Provider: "unsupported",
	}

	repo, err := CreateTemplateRepository(config)
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}

	if repo != nil {
		t.Fatal("expected nil repository for unsupported provider")
	}

	expectedError := "unsupported storage provider: unsupported"
	if err.Error() != expectedError {
		t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestCreateReportRepository_SeaweedFS(t *testing.T) {
	config := &Config{
		Provider:           "seaweedfs",
		SeaweedFSHost:      "localhost",
		SeaweedFSFilerPort: "8888",
	}

	repo, err := CreateReportRepository(config)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo == nil {
		t.Fatal("expected repository, got nil")
	}
}

func TestCreateReportRepository_S3(t *testing.T) {
	config := &Config{
		Provider:          "s3",
		S3Region:          "us-east-1",
		S3Bucket:          "test-bucket",
		S3AccessKeyID:     "test-key",
		S3SecretAccessKey: "test-secret",
	}

	repo, err := CreateReportRepository(config)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo == nil {
		t.Fatal("expected repository, got nil")
	}
}

func TestCreateReportRepository_CaseSensitive(t *testing.T) {
	config := &Config{
		Provider: "S3", // Uppercase should fail
	}

	repo, err := CreateReportRepository(config)
	if err == nil {
		t.Fatal("expected error for uppercase provider")
	}

	if repo != nil {
		t.Fatal("expected nil repository for invalid provider")
	}
}
