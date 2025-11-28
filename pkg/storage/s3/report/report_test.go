package report

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestSimpleRepository_Put_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockS3Client(ctrl)
	mockClient.EXPECT().
		UploadFileWithContentType(gomock.Any(), "reports/test-report.pdf", []byte("test data"), "application/pdf", "1h").
		Return(nil)

	repo := NewSimpleRepository(mockClient)
	ctx := context.Background()

	err := repo.Put(ctx, "test-report.pdf", "application/pdf", []byte("test data"), "1h")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestSimpleRepository_Put_WithoutTTL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockS3Client(ctrl)
	mockClient.EXPECT().
		UploadFileWithContentType(gomock.Any(), "reports/test-report.pdf", []byte("test data"), "application/pdf", "").
		Return(nil)

	repo := NewSimpleRepository(mockClient)
	ctx := context.Background()

	err := repo.Put(ctx, "test-report.pdf", "application/pdf", []byte("test data"), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSimpleRepository_Get_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedData := []byte("report content")
	mockClient := NewMockS3Client(ctrl)
	mockClient.EXPECT().
		DownloadFile(gomock.Any(), "reports/test-report.pdf").
		Return(expectedData, nil)

	repo := NewSimpleRepository(mockClient)
	ctx := context.Background()

	data, err := repo.Get(ctx, "test-report.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(data) != string(expectedData) {
		t.Errorf("expected data '%s', got '%s'", string(expectedData), string(data))
	}
}

func TestSimpleRepository_TTLFormats(t *testing.T) {
	testCases := []string{"3m", "4h", "5d", "6w", "7M", "8y", ""}

	for _, ttl := range testCases {
		t.Run("TTL_"+ttl, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := NewMockS3Client(ctrl)
			mockClient.EXPECT().
				UploadFileWithContentType(gomock.Any(), "reports/test.pdf", []byte("test"), "application/pdf", ttl).
				Return(nil)

			repo := NewSimpleRepository(mockClient)
			ctx := context.Background()

			err := repo.Put(ctx, "test.pdf", "application/pdf", []byte("test"), ttl)
			if err != nil {
				t.Fatalf("unexpected error for TTL '%s': %v", ttl, err)
			}
		})
	}
}
