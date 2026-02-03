// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package report

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

// mockObjectStorage is a test double for storage.ObjectStorage
type mockObjectStorage struct {
	uploadFunc   func(ctx context.Context, key string, reader io.Reader, contentType string) (string, error)
	downloadFunc func(ctx context.Context, key string) (io.ReadCloser, error)
	uploadErr    error
	downloadErr  error
	uploadedKey  string
	uploadedData []byte
	downloadData []byte
}

func (m *mockObjectStorage) Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error) {
	if m.uploadFunc != nil {
		return m.uploadFunc(ctx, key, reader, contentType)
	}
	if m.uploadErr != nil {
		return "", m.uploadErr
	}
	m.uploadedKey = key
	data, _ := io.ReadAll(reader)
	m.uploadedData = data
	return key, nil
}

func (m *mockObjectStorage) UploadWithTTL(ctx context.Context, key string, reader io.Reader, contentType string, ttl string) (string, error) {
	if m.uploadErr != nil {
		return "", m.uploadErr
	}
	m.uploadedKey = key
	data, _ := io.ReadAll(reader)
	m.uploadedData = data
	return key, nil
}

func (m *mockObjectStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.downloadFunc != nil {
		return m.downloadFunc(ctx, key)
	}
	if m.downloadErr != nil {
		return nil, m.downloadErr
	}
	return io.NopCloser(bytes.NewReader(m.downloadData)), nil
}

func (m *mockObjectStorage) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockObjectStorage) Exists(ctx context.Context, key string) (bool, error) {
	return true, nil
}

func (m *mockObjectStorage) GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "", nil
}

func TestStorageRepository_Put_Success(t *testing.T) {
	mock := &mockObjectStorage{}
	repo := NewStorageRepository(mock)

	err := repo.Put(context.Background(), "obj.txt", "text/plain", []byte("hello"), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.uploadedKey != "reports/obj.txt" {
		t.Fatalf("expected key reports/obj.txt, got %s", mock.uploadedKey)
	}
	if string(mock.uploadedData) != "hello" {
		t.Fatalf("unexpected data: %q", string(mock.uploadedData))
	}
}

func TestStorageRepository_Put_Error(t *testing.T) {
	mock := &mockObjectStorage{
		uploadErr: errors.New("upload failed"),
	}
	repo := NewStorageRepository(mock)

	err := repo.Put(context.Background(), "obj.txt", "text/plain", []byte("hello"), "")
	if err == nil {
		t.Fatalf("expected put error, got nil")
	}
}

func TestStorageRepository_Put_WithTTL_Success(t *testing.T) {
	mock := &mockObjectStorage{}
	repo := NewStorageRepository(mock)

	err := repo.Put(context.Background(), "temp.txt", "text/plain", []byte("temporary"), "1m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.uploadedKey != "reports/temp.txt" {
		t.Fatalf("expected key reports/temp.txt, got %s", mock.uploadedKey)
	}
	if string(mock.uploadedData) != "temporary" {
		t.Fatalf("unexpected data: %q", string(mock.uploadedData))
	}
}

func TestStorageRepository_Get_Success(t *testing.T) {
	mock := &mockObjectStorage{
		downloadData: []byte("world"),
	}
	repo := NewStorageRepository(mock)

	data, err := repo.Get(context.Background(), "obj.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "world" {
		t.Fatalf("unexpected data: %q", string(data))
	}
}

func TestStorageRepository_Get_Error(t *testing.T) {
	mock := &mockObjectStorage{
		downloadErr: errors.New("download failed"),
	}
	repo := NewStorageRepository(mock)

	_, err := repo.Get(context.Background(), "obj.txt")
	if err == nil {
		t.Fatalf("expected get error, got nil")
	}
}