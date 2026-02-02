package storage

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS3Config_DefaultSeaweedS3Config(t *testing.T) {
	bucket := "test-bucket"
	cfg := DefaultSeaweedS3Config(bucket)

	assert.Equal(t, "http://localhost:8333", cfg.Endpoint)
	assert.Equal(t, "us-east-1", cfg.Region)
	assert.Equal(t, bucket, cfg.Bucket)
	assert.True(t, cfg.UsePathStyle)
	assert.True(t, cfg.DisableSSL)
}

func TestNewS3Client_RequiresBucket(t *testing.T) {
	ctx := context.Background()
	cfg := S3Config{
		Endpoint: "http://localhost:9000",
	}

	client, err := NewS3Client(ctx, cfg)

	assert.Nil(t, client)
	assert.Equal(t, ErrBucketRequired, err)
}

// createTestClient creates a S3Client for testing parameter validation.
// The client won't be able to connect to S3, but we can test input validation.
func createTestClient(t *testing.T) *S3Client {
	t.Helper()

	ctx := context.Background()
	cfg := S3Config{
		Endpoint:        "http://localhost:9999", // Non-existent endpoint
		Region:          "us-east-1",
		Bucket:          "test-bucket",
		AccessKeyID:     "test-key",
		SecretAccessKey: "test-secret",
		UsePathStyle:    true,
		DisableSSL:      true,
	}

	client, err := NewS3Client(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	return client
}

func TestS3Client_UploadRequiresKey(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	// Test with empty key
	_, err := client.Upload(ctx, "", strings.NewReader("test data"), "text/plain")

	assert.Equal(t, ErrKeyRequired, err)
}

func TestS3Client_UploadWithTTLRequiresKey(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	// Test with empty key
	_, err := client.UploadWithTTL(ctx, "", strings.NewReader("test data"), "text/plain", "1h")

	assert.Equal(t, ErrKeyRequired, err)
}

func TestS3Client_DownloadRequiresKey(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	// Test with empty key
	_, err := client.Download(ctx, "")

	assert.Equal(t, ErrKeyRequired, err)
}

func TestS3Client_DeleteRequiresKey(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	// Test with empty key
	err := client.Delete(ctx, "")

	assert.Equal(t, ErrKeyRequired, err)
}

func TestS3Client_ExistsRequiresKey(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	// Test with empty key
	exists, err := client.Exists(ctx, "")

	assert.False(t, exists)
	assert.Equal(t, ErrKeyRequired, err)
}

func TestS3Client_GeneratePresignedURLRequiresKey(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	// Test with empty key
	url, err := client.GeneratePresignedURL(ctx, "", 1*time.Hour)

	assert.Empty(t, url)
	assert.Equal(t, ErrKeyRequired, err)
}

func TestS3Client_GeneratePresignedURL_Success(t *testing.T) {
	client := createTestClient(t)
	ctx := context.Background()

	// GeneratePresignedURL doesn't need actual S3 connection to generate a URL
	url, err := client.GeneratePresignedURL(ctx, "test-key.txt", 1*time.Hour)

	// Should succeed even without S3 connection (it just generates the URL locally)
	assert.NoError(t, err)
	assert.NotEmpty(t, url)
	assert.Contains(t, url, "test-key.txt")
	assert.Contains(t, url, "test-bucket")
}

func TestS3Config_WithAllOptions(t *testing.T) {
	ctx := context.Background()
	cfg := S3Config{
		Endpoint:        "http://custom-endpoint:9000",
		Region:          "eu-west-1",
		Bucket:          "my-bucket",
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		UsePathStyle:    true,
		DisableSSL:      true,
	}

	client, err := NewS3Client(ctx, cfg)

	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestS3Config_MinimalConfig(t *testing.T) {
	ctx := context.Background()
	cfg := S3Config{
		Bucket: "minimal-bucket",
	}

	client, err := NewS3Client(ctx, cfg)

	// Should succeed with minimal config (will use default AWS config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

// Integration test - requires actual S3 service running
func TestS3ClientIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()

	// Configure for local MinIO or SeaweedFS S3
	cfg := S3Config{
		Endpoint:        "http://localhost:9000",
		Region:          "us-east-1",
		Bucket:          "test-bucket",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		UsePathStyle:    true,
		DisableSSL:      true,
	}

	client, err := NewS3Client(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Test upload
	testKey := "test-file.txt"
	testContent := []byte("Hello S3!")
	testContentType := "text/plain"

	uploadedKey, err := client.Upload(ctx, testKey, bytes.NewReader(testContent), testContentType)
	if err != nil {
		t.Logf("Upload failed (S3 service may not be available): %v", err)
		t.Skip("S3 service not available")
	}
	assert.Equal(t, testKey, uploadedKey)

	// Test exists
	exists, err := client.Exists(ctx, testKey)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test download
	reader, err := client.Download(ctx, testKey)
	require.NoError(t, err)
	defer reader.Close()

	downloadedContent, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, testContent, downloadedContent)

	// Test presigned URL
	url, err := client.GeneratePresignedURL(ctx, testKey, 1*time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, url)

	// Test delete
	err = client.Delete(ctx, testKey)
	require.NoError(t, err)

	// Verify deletion
	exists, err = client.Exists(ctx, testKey)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestS3Client_UploadWithTTL_IgnoresTTL(t *testing.T) {
	// TTL should be ignored for S3 (not an error, just logged)
	// We verify the function doesn't error with a TTL value
	client := createTestClient(t)
	ctx := context.Background()

	// This will fail to connect but should pass validation with TTL
	_, err := client.UploadWithTTL(ctx, "test-key.txt", strings.NewReader("data"), "text/plain", "1h")

	// Should fail with connection error, NOT with TTL error
	// The TTL warning is logged but the function continues
	assert.Error(t, err)
	assert.NotEqual(t, ErrTTLNotSupported, err)
	assert.NotEqual(t, ErrKeyRequired, err)
	// Error should be about uploading/connection, not TTL
	assert.Contains(t, err.Error(), "uploading object")
}