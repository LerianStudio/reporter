package report

import (
	"context"
	"testing"
)

// MockS3Client implements S3Client interface for testing
type MockS3Client struct {
	uploadWithTTLFunc func(ctx context.Context, path string, data []byte, ttl string) error
	downloadFunc      func(ctx context.Context, path string) ([]byte, error)
}

func (m *MockS3Client) UploadFileWithTTL(ctx context.Context, path string, data []byte, ttl string) error {
	if m.uploadWithTTLFunc != nil {
		return m.uploadWithTTLFunc(ctx, path, data, ttl)
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
	var capturedTTL string

	mockClient := &MockS3Client{
		uploadWithTTLFunc: func(ctx context.Context, path string, data []byte, ttl string) error {
			capturedPath = path
			capturedData = data
			capturedTTL = ttl
			return nil
		},
	}

	repo := NewSimpleRepository(mockClient)
	ctx := context.Background()
	testData := []byte("test report content")

	err := repo.Put(ctx, "test-report.pdf", "application/pdf", testData, "1h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath != "test-report.pdf" {
		t.Errorf("expected path 'test-report.pdf', got '%s'", capturedPath)
	}

	if string(capturedData) != string(testData) {
		t.Errorf("expected data '%s', got '%s'", string(testData), string(capturedData))
	}

	if capturedTTL != "1h" {
		t.Errorf("expected TTL '1h', got '%s'", capturedTTL)
	}
}

func TestSimpleRepository_Put_WithoutTTL(t *testing.T) {
	var capturedTTL string

	mockClient := &MockS3Client{
		uploadWithTTLFunc: func(ctx context.Context, path string, data []byte, ttl string) error {
			capturedTTL = ttl
			return nil
		},
	}

	repo := NewSimpleRepository(mockClient)
	ctx := context.Background()
	testData := []byte("test report content")

	err := repo.Put(ctx, "test-report.pdf", "application/pdf", testData, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedTTL != "" {
		t.Errorf("expected empty TTL, got '%s'", capturedTTL)
	}
}

func TestSimpleRepository_Get_Success(t *testing.T) {
	expectedData := []byte("report content")
	var capturedPath string

	mockClient := &MockS3Client{
		downloadFunc: func(ctx context.Context, path string) ([]byte, error) {
			capturedPath = path
			return expectedData, nil
		},
	}

	repo := NewSimpleRepository(mockClient)
	ctx := context.Background()

	data, err := repo.Get(ctx, "test-report.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath != "test-report.pdf" {
		t.Errorf("expected path 'test-report.pdf', got '%s'", capturedPath)
	}

	if string(data) != string(expectedData) {
		t.Errorf("expected data '%s', got '%s'", string(expectedData), string(data))
	}
}

func TestSimpleRepository_TTLFormats(t *testing.T) {
	testCases := []string{"3m", "4h", "5d", "6w", "7M", "8y", ""}

	for _, ttl := range testCases {
		t.Run("TTL_"+ttl, func(t *testing.T) {
			var capturedTTL string

			mockClient := &MockS3Client{
				uploadWithTTLFunc: func(ctx context.Context, path string, data []byte, ttl string) error {
					capturedTTL = ttl
					return nil
				},
			}

			repo := NewSimpleRepository(mockClient)
			ctx := context.Background()

			err := repo.Put(ctx, "test.pdf", "application/pdf", []byte("test"), ttl)
			if err != nil {
				t.Fatalf("unexpected error for TTL '%s': %v", ttl, err)
			}

			if capturedTTL != ttl {
				t.Errorf("expected TTL '%s', got '%s'", ttl, capturedTTL)
			}
		})
	}
}
