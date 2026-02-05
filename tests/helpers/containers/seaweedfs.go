// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package containers

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	SeaweedBucket    = "reporter-storage"
	SeaweedAccessKey = "any"
	SeaweedSecretKey = "any"
	SeaweedRegion    = "us-east-1"
)

// SeaweedFSContainer wraps a SeaweedFS testcontainer with S3 endpoint info.
type SeaweedFSContainer struct {
	testcontainers.Container
	S3Endpoint string
	Host       string
	S3Port     string
	AdminPort  string
}

// StartSeaweedFS creates and starts a SeaweedFS container in S3 mode.
func StartSeaweedFS(ctx context.Context, networkName, image string) (*SeaweedFSContainer, error) {
	if image == "" {
		image = "chrislusf/seaweedfs:3.97"
	}

	req := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{"8333/tcp", "9333/tcp"},
		Cmd:          []string{"server", "-s3", "-dir=/data"},
		Networks:     []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"seaweedfs", "reporter-seaweedfs"},
		},
		WaitingFor: wait.ForAll(
			wait.ForHTTP("/cluster/status").WithPort("9333/tcp"),
			wait.ForListeningPort("8333/tcp"),
		).WithDeadline(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("start seaweedfs container: %w", err)
	}

	// Get host and ports
	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get seaweedfs host: %w", err)
	}

	s3Port, err := container.MappedPort(ctx, "8333")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get seaweedfs s3 port: %w", err)
	}

	adminPort, err := container.MappedPort(ctx, "9333")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get seaweedfs admin port: %w", err)
	}

	s3Endpoint := fmt.Sprintf("http://%s:%s", host, s3Port.Port())

	sc := &SeaweedFSContainer{
		Container:  container,
		S3Endpoint: s3Endpoint,
		Host:       host,
		S3Port:     s3Port.Port(),
		AdminPort:  adminPort.Port(),
	}

	// Create bucket
	if err := sc.createBucket(ctx); err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("create bucket: %w", err)
	}

	return sc, nil
}

// createBucket creates the default storage bucket with retry.
func (s *SeaweedFSContainer) createBucket(ctx context.Context) error {
	client, err := s.getS3Client(ctx)
	if err != nil {
		return err
	}

	// Retry bucket creation - S3 API may not be immediately ready
	var lastErr error

	for i := 0; i < 10; i++ {
		_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(SeaweedBucket),
		})
		if err == nil {
			return nil
		}

		lastErr = err

		time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
	}

	return fmt.Errorf("create bucket %s after retries: %w", SeaweedBucket, lastErr)
}

// getS3Client creates an S3 client for the SeaweedFS container.
func (s *SeaweedFSContainer) getS3Client(ctx context.Context) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(SeaweedRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			SeaweedAccessKey,
			SeaweedSecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s.S3Endpoint)
		o.UsePathStyle = true
	})

	return client, nil
}

// Restart stops and starts the SeaweedFS container.
func (s *SeaweedFSContainer) Restart(ctx context.Context, delay time.Duration) error {
	timeout := 10 * time.Second
	if err := s.Stop(ctx, &timeout); err != nil {
		return fmt.Errorf("stop seaweedfs: %w", err)
	}

	if delay > 0 {
		time.Sleep(delay)
	}

	if err := s.Start(ctx); err != nil {
		return fmt.Errorf("start seaweedfs: %w", err)
	}

	// Re-create bucket after restart
	if err := s.createBucket(ctx); err != nil {
		// Bucket might already exist, ignore error
		_ = err
	}

	return nil
}

// GetS3Config returns S3 configuration for connecting to this container.
func (s *SeaweedFSContainer) GetS3Config() map[string]string {
	return map[string]string{
		"OBJECT_STORAGE_ENDPOINT":       s.S3Endpoint,
		"OBJECT_STORAGE_REGION":         SeaweedRegion,
		"OBJECT_STORAGE_ACCESS_KEY_ID":  SeaweedAccessKey,
		"OBJECT_STORAGE_SECRET_KEY":     SeaweedSecretKey,
		"OBJECT_STORAGE_BUCKET":         SeaweedBucket,
		"OBJECT_STORAGE_USE_PATH_STYLE": "true",
		"OBJECT_STORAGE_DISABLE_SSL":    "true",
	}
}
