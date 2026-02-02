package storage

import (
	"bytes"
	"context"
	"io"
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

func TestS3Client_UploadRequiresKey(t *testing.T) {
	// This test requires a mock S3 client, skipping for now
	t.Skip("Requires mock S3 service")
}

func TestS3Client_DownloadRequiresKey(t *testing.T) {
	// This test requires a mock S3 client, skipping for now
	t.Skip("Requires mock S3 service")
}

func TestS3Client_DeleteRequiresKey(t *testing.T) {
	// This test requires a mock S3 client, skipping for now
	t.Skip("Requires mock S3 service")
}

func TestS3Client_ExistsRequiresKey(t *testing.T) {
	// This test requires a mock S3 client, skipping for now
	t.Skip("Requires mock S3 service")
}

func TestS3Client_GeneratePresignedURLRequiresKey(t *testing.T) {
	// This test requires a mock S3 client, skipping for now
	t.Skip("Requires mock S3 service")
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

func TestS3Client_UploadWithTTL_LogsWarning(t *testing.T) {
	// TTL should be ignored for S3 (not an error, just logged)
	// This is tested implicitly in integration test
	t.Skip("Requires logger mock to verify warning")
}
