package s3storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client provides S3 storage operations
type S3Client struct {
	client *s3.Client
	bucket string
}

// S3Config holds AWS S3 specific configuration
type S3Config struct {
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string
	ForcePathStyle  bool
}

// NewS3Client creates a new S3 client with the given configuration
func NewS3Client(cfg *S3Config, bucket string) (*S3Client, error) {
	// Create AWS config
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with custom endpoint if provided (for MinIO, LocalStack, etc.)
	var s3Client *s3.Client
	if cfg.Endpoint != "" {
		s3Client = s3.NewFromConfig(awsConfig, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = cfg.ForcePathStyle
		})
	} else {
		s3Client = s3.NewFromConfig(awsConfig)
	}

	return &S3Client{
		client: s3Client,
		bucket: bucket,
	}, nil
}

// UploadFile uploads a file to S3
func (c *S3Client) UploadFile(ctx context.Context, path string, data []byte) error {
	return c.UploadFileWithTTL(ctx, path, data, "")
}

// UploadFileWithContentType uploads a file to S3 with specified content type and TTL
func (c *S3Client) UploadFileWithContentType(ctx context.Context, path string, data []byte, contentType, ttl string) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(path),
		Body:   bytes.NewReader(data),
	}

	// Use provided content type or detect from file extension
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	} else {
		detectedType := getContentType(path)
		if detectedType != "" {
			input.ContentType = aws.String(detectedType)
		}
	}

	// Handle TTL by setting expiration date
	if ttl != "" {
		expiration, err := parseTTL(ttl)
		if err != nil {
			return fmt.Errorf("invalid TTL format: %w", err)
		}

		input.Expires = aws.Time(expiration)
	}

	_, err := c.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return nil
}

// UploadFileWithTTL uploads a file to S3 with optional TTL
// Note: TTL sets HTTP cache expiration metadata only, not automatic object deletion.
// For automatic deletion, configure S3 Lifecycle Rules on the bucket.
func (c *S3Client) UploadFileWithTTL(ctx context.Context, path string, data []byte, ttl string) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(path),
		Body:   bytes.NewReader(data),
	}

	// Set content type based on file extension
	contentType := getContentType(path)
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	// Handle TTL by setting expiration date
	if ttl != "" {
		expiration, err := parseTTL(ttl)
		if err != nil {
			return fmt.Errorf("invalid TTL format: %w", err)
		}

		input.Expires = aws.Time(expiration)
	}

	_, err := c.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return nil
}

// DownloadFile downloads a file from S3
func (c *S3Client) DownloadFile(ctx context.Context, path string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(path),
	}

	result, err := c.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %w", err)
	}

	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object body: %w", err)
	}

	return data, nil
}

// DeleteFile deletes a file from S3
func (c *S3Client) DeleteFile(ctx context.Context, path string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(path),
	}

	_, err := c.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}

// HealthCheck verifies if S3 is accessible
func (c *S3Client) HealthCheck(ctx context.Context) error {
	input := &s3.HeadBucketInput{
		Bucket: aws.String(c.bucket),
	}

	_, err := c.client.HeadBucket(ctx, input)
	if err != nil {
		return fmt.Errorf("S3 health check failed: %w", err)
	}

	return nil
}

// parseTTL converts TTL string to time.Time
func parseTTL(ttl string) (time.Time, error) {
	if len(ttl) < 2 {
		return time.Time{}, fmt.Errorf("invalid TTL format: %s", ttl)
	}

	unit := ttl[len(ttl)-1:]
	valueStr := ttl[:len(ttl)-1]

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid TTL value: %s", valueStr)
	}

	now := time.Now()

	switch unit {
	case "m":
		return now.Add(time.Duration(value) * time.Minute), nil
	case "h":
		return now.Add(time.Duration(value) * time.Hour), nil
	case "d":
		return now.Add(time.Duration(value) * 24 * time.Hour), nil
	case "w":
		return now.Add(time.Duration(value) * 7 * 24 * time.Hour), nil
	case "M":
		return now.AddDate(0, value, 0), nil
	case "y":
		return now.AddDate(value, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported TTL unit: %s", unit)
	}
}

// getContentType returns appropriate content type based on file extension
func getContentType(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".tpl":
		return "text/plain"
	case ".html":
		return "text/html"
	case ".htm":
		return "text/html"
	case ".pdf":
		return "application/pdf"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
