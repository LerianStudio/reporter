// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

// Package storage provides object storage adapters for templates and reports.
package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v2/commons/opentelemetry"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Config contains configuration for S3-compatible storage.
// Works with AWS S3, MinIO, SeaweedFS S3, and other S3-compatible services.
type S3Config struct {
	Endpoint        string // For SeaweedFS: http://localhost:8333, for MinIO: http://localhost:9000
	Region          string // Default: us-east-1
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	UsePathStyle    bool // Required for SeaweedFS/MinIO
	DisableSSL      bool
}

// DefaultSeaweedS3Config returns a configuration suitable for local SeaweedFS S3 development.
func DefaultSeaweedS3Config(bucket string) S3Config {
	return S3Config{
		Endpoint:     "http://localhost:8333",
		Region:       "us-east-1",
		Bucket:       bucket,
		UsePathStyle: true,
		DisableSSL:   true,
	}
}

// S3Client provides S3-compatible object storage operations.
type S3Client struct {
	s3     *s3.Client
	bucket string
}

var (
	// ErrBucketRequired indicates bucket name is missing.
	ErrBucketRequired = errors.New("bucket name is required")
	// ErrKeyRequired indicates object key is missing.
	ErrKeyRequired = errors.New("object key is required")
	// ErrObjectNotFound indicates the object does not exist.
	ErrObjectNotFound = errors.New("object not found")
	// ErrTTLNotSupported indicates TTL is not supported by S3 (use lifecycle policies instead).
	ErrTTLNotSupported = errors.New("TTL parameter not supported in S3 mode - use bucket lifecycle policies instead")
)

// NewS3Client creates a new S3 client with the given configuration.
func NewS3Client(ctx context.Context, cfg S3Config) (*S3Client, error) {
	if cfg.Bucket == "" {
		return nil, ErrBucketRequired
	}

	var opts []func(*config.LoadOptions) error

	if cfg.Region != "" {
		opts = append(opts, config.WithRegion(cfg.Region))
	}

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("loading aws config: %w", err)
	}

	clientOpts := []func(*s3.Options){}

	if cfg.Endpoint != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
	}

	if cfg.UsePathStyle {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	}

	s3Client := s3.NewFromConfig(awsCfg, clientOpts...)

	return &S3Client{
		s3:     s3Client,
		bucket: cfg.Bucket,
	}, nil
}

// Upload stores content from a reader at the given key.
func (client *S3Client) Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error) {
	return client.UploadWithTTL(ctx, key, reader, contentType, "")
}

// UploadWithTTL stores content with a time-to-live.
// Note: S3 does not support per-object TTL via upload parameters.
// TTL parameter is ignored - use S3 bucket lifecycle policies instead.
// This method exists for interface compatibility with SeaweedFS.
func (client *S3Client) UploadWithTTL(ctx context.Context, key string, reader io.Reader, contentType string, ttl string) (string, error) {
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)
	ctx, span := tracer.Start(ctx, "repository.storage.upload")

	defer span.End()

	if key == "" {
		return "", ErrKeyRequired
	}

	// Log warning if TTL is provided (not supported in S3)
	if ttl != "" && logger != nil {
		logger.Warnf("TTL parameter '%s' ignored for S3 storage - configure bucket lifecycle policies instead", ttl)
	}

	// Read all data into memory (required for S3 SDK)
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("reading data: %w", err)
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(client.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	if _, err := client.s3.PutObject(ctx, input); err != nil {
		libOpentelemetry.HandleSpanError(&span, "failed to upload object", err)

		if logger != nil {
			logger.Errorf("failed to upload object %s: %v", key, err)
		}

		return "", fmt.Errorf("uploading object: %w", err)
	}

	if logger != nil {
		logger.Infof("uploaded object %s to bucket %s", key, client.bucket)
	}

	return key, nil
}

// Download retrieves content from the given key.
func (client *S3Client) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)
	ctx, span := tracer.Start(ctx, "repository.storage.download")

	defer span.End()

	if key == "" {
		return nil, ErrKeyRequired
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(client.bucket),
		Key:    aws.String(key),
	}

	result, err := client.s3.GetObject(ctx, input)
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil, ErrObjectNotFound
		}

		libOpentelemetry.HandleSpanError(&span, "failed to download object", err)

		if logger != nil {
			logger.Errorf("failed to download object %s: %v", key, err)
		}

		return nil, fmt.Errorf("downloading object: %w", err)
	}

	return result.Body, nil
}

// Delete removes an object by key.
func (client *S3Client) Delete(ctx context.Context, key string) error {
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)
	ctx, span := tracer.Start(ctx, "repository.storage.delete")

	defer span.End()

	if key == "" {
		return ErrKeyRequired
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(client.bucket),
		Key:    aws.String(key),
	}

	if _, err := client.s3.DeleteObject(ctx, input); err != nil {
		libOpentelemetry.HandleSpanError(&span, "failed to delete object", err)

		if logger != nil {
			logger.Errorf("failed to delete object %s: %v", key, err)
		}

		return fmt.Errorf("deleting object: %w", err)
	}

	if logger != nil {
		logger.Infof("deleted object %s from bucket %s", key, client.bucket)
	}

	return nil
}

// GeneratePresignedURL creates a time-limited download URL.
func (client *S3Client) GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)
	ctx, span := tracer.Start(ctx, "repository.storage.generate_presigned_url")

	defer span.End()

	if key == "" {
		return "", ErrKeyRequired
	}

	presigner := s3.NewPresignClient(client.s3)

	input := &s3.GetObjectInput{
		Bucket: aws.String(client.bucket),
		Key:    aws.String(key),
	}

	result, err := presigner.PresignGetObject(ctx, input, s3.WithPresignExpires(expiry))
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "failed to generate presigned url", err)

		if logger != nil {
			logger.Errorf("failed to generate presigned url for %s: %v", key, err)
		}

		return "", fmt.Errorf("generating presigned url: %w", err)
	}

	return result.URL, nil
}

// Exists checks if an object exists at the given key.
func (client *S3Client) Exists(ctx context.Context, key string) (bool, error) {
	logger, tracer, _, _ := libCommons.NewTrackingFromContext(ctx)
	ctx, span := tracer.Start(ctx, "repository.storage.exists")

	defer span.End()

	if key == "" {
		return false, ErrKeyRequired
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(client.bucket),
		Key:    aws.String(key),
	}

	if _, err := client.s3.HeadObject(ctx, input); err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return false, nil
		}

		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}

		libOpentelemetry.HandleSpanError(&span, "failed to check object existence", err)

		if logger != nil {
			logger.Errorf("failed to check existence of %s: %v", key, err)
		}

		return false, fmt.Errorf("checking object existence: %w", err)
	}

	return true, nil
}

// Compile-time interface check.
var _ ObjectStorage = (*S3Client)(nil)
