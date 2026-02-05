// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package template

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/LerianStudio/reporter/v4/pkg"
	"github.com/LerianStudio/reporter/v4/pkg/constant"
	"github.com/LerianStudio/reporter/v4/pkg/storage"
)

// Repository provides an interface for storage operations
//
//go:generate mockgen --destination=template.mock.go --package=template . Repository
type Repository interface {
	Get(ctx context.Context, objectName string) ([]byte, error)
	Put(ctx context.Context, objectName string, contentType string, data []byte) error
}

// StorageRepository provides access to object storage for template operations.
type StorageRepository struct {
	storage storage.ObjectStorage
}

// NewStorageRepository creates a new instance of StorageRepository with the given storage client.
func NewStorageRepository(storageClient storage.ObjectStorage) *StorageRepository {
	return &StorageRepository{
		storage: storageClient,
	}
}

// Get the content of a .tpl file from the storage.
// objectName can be passed with or without .tpl extension - it will be normalized.
func (repo *StorageRepository) Get(ctx context.Context, objectName string) ([]byte, error) {
	// Normalize: ensure .tpl extension (handles both "uuid" and "uuid.tpl")
	objectName = strings.TrimSuffix(objectName, ".tpl")
	key := fmt.Sprintf("templates/%s.tpl", objectName)

	reader, err := repo.storage.Download(ctx, key)
	if err != nil {
		return nil, pkg.ValidateBusinessError(constant.ErrCommunicateSeaweedFS, "")
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, pkg.ValidateBusinessError(constant.ErrCommunicateSeaweedFS, "")
	}

	return data, nil
}

// Put uploads data to the storage with the given object name and content type.
// objectName can be passed with or without .tpl extension - it will be normalized.
func (repo *StorageRepository) Put(ctx context.Context, objectName string, contentType string, data []byte) error {
	logger := pkg.NewLoggerFromContext(ctx)

	// Normalize: ensure .tpl extension (handles both "uuid" and "uuid.tpl")
	objectName = strings.TrimSuffix(objectName, ".tpl")
	key := fmt.Sprintf("templates/%s.tpl", objectName)

	_, err := repo.storage.Upload(ctx, key, bytes.NewReader(data), contentType)
	if err != nil {
		logger.Errorf("Error communicating with storage: %v", err)
		return pkg.ValidateBusinessError(constant.ErrCommunicateSeaweedFS, "")
	}

	return nil
}
