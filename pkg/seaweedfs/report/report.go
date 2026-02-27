// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package report

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/LerianStudio/reporter/pkg"
	"github.com/LerianStudio/reporter/pkg/constant"
	"github.com/LerianStudio/reporter/pkg/storage"

	"github.com/LerianStudio/lib-commons/v3/commons"
	libOpentelemetry "github.com/LerianStudio/lib-commons/v3/commons/opentelemetry"
	tmS3 "github.com/LerianStudio/lib-commons/v3/commons/tenant-manager/s3"
	"go.opentelemetry.io/otel/attribute"
)

// Repository provides an interface for storage operations
//
//go:generate mockgen --destination=report.mock.go --package=report --copyright_file=../../../COPYRIGHT . Repository
type Repository interface {
	Put(ctx context.Context, objectName string, contentType string, data []byte, ttl string) error
	Get(ctx context.Context, objectName string) ([]byte, error)
}

// StorageRepository provides access to object storage for report operations.
type StorageRepository struct {
	storage storage.ObjectStorage
}

// Compile-time interface satisfaction check.
var _ Repository = (*StorageRepository)(nil)

// NewStorageRepository creates a new instance of StorageRepository with the given storage client.
func NewStorageRepository(storageClient storage.ObjectStorage) *StorageRepository {
	return &StorageRepository{
		storage: storageClient,
	}
}

// Put uploads data to the storage with the given object name, content type, and optional TTL.
// TTL format: 3m (3 minutes), 4h (4 hours), 5d (5 days), 6w (6 weeks), 7M (7 months), 8y (8 years)
// If ttl is empty string, no TTL is applied and the file will be stored permanently
func (repo *StorageRepository) Put(ctx context.Context, objectName string, contentType string, data []byte, ttl string) error {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "repository.report_storage.put")
	defer span.End()

	span.SetAttributes(attribute.String("app.request.request_id", reqId))

	// Add reports prefix, then apply tenant prefix if in multi-tenant mode
	baseKey := fmt.Sprintf("reports/%s", objectName)
	key := tmS3.GetObjectStorageKeyForTenant(ctx, baseKey)

	logger.Infof("Putting report to storage: %s", key)

	_, err := repo.storage.UploadWithTTL(ctx, key, bytes.NewReader(data), contentType, ttl)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to upload report to storage", err)

		return pkg.ValidateBusinessError(constant.ErrCommunicateSeaweedFS, "")
	}

	return nil
}

// Get download data from storage with the given object name
func (repo *StorageRepository) Get(ctx context.Context, objectName string) ([]byte, error) {
	logger, tracer, reqId, _ := commons.NewTrackingFromContext(ctx)

	ctx, span := tracer.Start(ctx, "repository.report_storage.get")
	defer span.End()

	span.SetAttributes(attribute.String("app.request.request_id", reqId))

	// Add reports prefix, then apply tenant prefix if in multi-tenant mode
	baseKey := fmt.Sprintf("reports/%s", objectName)
	key := tmS3.GetObjectStorageKeyForTenant(ctx, baseKey)

	logger.Infof("Getting report from storage: %s", key)

	reader, err := repo.storage.Download(ctx, key)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to download report from storage", err)

		return nil, pkg.ValidateBusinessError(constant.ErrCommunicateSeaweedFS, "")
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		libOpentelemetry.HandleSpanError(&span, "Failed to read report data", err)

		return nil, pkg.ValidateBusinessError(constant.ErrCommunicateSeaweedFS, "")
	}

	return data, nil
}
