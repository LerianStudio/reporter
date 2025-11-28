package template

import (
	"context"
	"testing"
)

// MockS3Client implements S3Client interface for testing
type MockS3Client struct {
	uploadFunc   func(ctx context.Context, path string, data []byte, contentType, ttl string) error
	downloadFunc func(ctx context.Context, path string) ([]byte, error)
}

func (m *MockS3Client) UploadFileWithContentType(ctx context.Context, path string, data []byte, contentType, ttl string) error {
	if m.uploadFunc != nil {
		return m.uploadFunc(ctx, path, data, contentType, ttl)
	}
	return nil
}

func (m *MockS3Client) DownloadFile(ctx context.Context, path string) ([]byte, error) {
	if m.downloadFunc != nil {
		return m.downloadFunc(ctx, path)
	}
	return []byte("test data"), nil
}

func TestSimpleRepository_Put_Success(t *testing.T) {
	var capturedPath string
	var capturedData []byte
	var capturedContentType string

	mockClient := &MockS3Client{
		uploadFunc: func(ctx context.Context, path string, data []byte, contentType, ttl string) error {
			capturedPath = path
			capturedData = data
			capturedContentType = contentType
			return nil
		},
	}

	repo := NewSimpleRepository(mockClient)
	ctx := context.Background()
	testData := []byte("test template content")

	err := repo.Put(ctx, "test-template", "text/plain", testData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should automatically add templates/ prefix and .tpl extension
	expectedPath := "templates/test-template.tpl"
	if capturedPath != expectedPath {
		t.Errorf("expected path '%s', got '%s'", expectedPath, capturedPath)
	}

	if string(capturedData) != string(testData) {
		t.Errorf("expected data '%s', got '%s'", string(testData), string(capturedData))
	}

	if capturedContentType != "text/plain" {
		t.Errorf("expected contentType 'text/plain', got '%s'", capturedContentType)
	}
}

func TestSimpleRepository_Get_Success(t *testing.T) {
	expectedData := []byte("template content")
	var capturedPath string

	mockClient := &MockS3Client{
		downloadFunc: func(ctx context.Context, path string) ([]byte, error) {
			capturedPath = path
			return expectedData, nil
		},
	}

	repo := NewSimpleRepository(mockClient)
	ctx := context.Background()

	data, err := repo.Get(ctx, "test-template")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should automatically add templates/ prefix and .tpl extension
	expectedPath := "templates/test-template.tpl"
	if capturedPath != expectedPath {
		t.Errorf("expected path '%s', got '%s'", expectedPath, capturedPath)
	}

	if string(data) != string(expectedData) {
		t.Errorf("expected data '%s', got '%s'", string(expectedData), string(data))
	}
}

func TestSimpleRepository_Get_AutoAddsTplExtension(t *testing.T) {
	var capturedPath string

	mockClient := &MockS3Client{
		downloadFunc: func(ctx context.Context, path string) ([]byte, error) {
			capturedPath = path
			return []byte("test"), nil
		},
	}

	repo := NewSimpleRepository(mockClient)
	ctx := context.Background()

	_, err := repo.Get(ctx, "my-template")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPath := "templates/my-template.tpl"
	if capturedPath != expectedPath {
		t.Errorf("expected path '%s', got '%s'", expectedPath, capturedPath)
	}
}
