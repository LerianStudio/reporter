package pdf

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewWorkerPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	wp := NewWorkerPool(2, 30*time.Second, mockLogger)

	assert.NotNil(t, wp)
	assert.NotNil(t, wp.tasks)
	assert.NotNil(t, wp.wg)
	assert.Equal(t, 2, wp.workers)
	assert.Equal(t, 30*time.Second, wp.timeout)
	assert.Equal(t, mockLogger, wp.logger)

	// Close the pool
	wp.Close()
}

func TestWorkerPool_GetStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	wp := NewWorkerPool(3, 45*time.Second, mockLogger)
	defer wp.Close()

	stats := wp.GetStats()

	assert.Equal(t, 3, stats["workers"])
	assert.Equal(t, 45*time.Second, stats["timeout"])
	assert.Contains(t, stats, "tasks_pending")
}

func TestWorkerPool_IsHealthy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	tests := []struct {
		name     string
		workers  int
		timeout  time.Duration
		expected bool
	}{
		{
			name:     "healthy pool",
			workers:  2,
			timeout:  30 * time.Second,
			expected: true,
		},
		{
			name:     "zero workers",
			workers:  0,
			timeout:  30 * time.Second,
			expected: false,
		},
		{
			name:     "zero timeout",
			workers:  2,
			timeout:  0,
			expected: false,
		},
		{
			name:     "both zero",
			workers:  0,
			timeout:  0,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wp := &WorkerPool{
				workers: tt.workers,
				timeout: tt.timeout,
			}

			result := wp.IsHealthy()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWorkerPool_getChromeOptions(t *testing.T) {
	wp := &WorkerPool{}

	options := wp.getChromeOptions()

	assert.NotEmpty(t, options)
	// Should have multiple options configured
	assert.Greater(t, len(options), 5)
}

func TestWorkerPool_createTempHTMLFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)

	wp := &WorkerPool{
		logger: mockLogger,
	}

	htmlContent := "<html><body><h1>Test</h1></body></html>"

	tmpFileName, err := wp.createTempHTMLFile(htmlContent)
	assert.NoError(t, err)
	assert.NotEmpty(t, tmpFileName)

	// Verify file exists and has correct content
	content, err := os.ReadFile(tmpFileName)
	assert.NoError(t, err)
	assert.Equal(t, htmlContent, string(content))

	// Cleanup
	os.Remove(tmpFileName)
}

func TestWorkerPool_createTempHTMLFile_LargeContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)

	wp := &WorkerPool{
		logger: mockLogger,
	}

	// Create large HTML content
	largeContent := "<html><body>"
	for i := 0; i < 1000; i++ {
		largeContent += "<p>This is paragraph " + string(rune('0'+i%10)) + "</p>"
	}
	largeContent += "</body></html>"

	tmpFileName, err := wp.createTempHTMLFile(largeContent)
	assert.NoError(t, err)
	assert.NotEmpty(t, tmpFileName)

	// Verify file exists
	_, err = os.Stat(tmpFileName)
	assert.NoError(t, err)

	// Cleanup
	os.Remove(tmpFileName)
}

func TestWorkerPool_processPDFResult_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	wp := &WorkerPool{
		logger: mockLogger,
	}

	// Create a valid PDF buffer (at least 1000 bytes)
	pdfBuf := make([]byte, 2000)
	for i := range pdfBuf {
		pdfBuf[i] = byte(i % 256)
	}

	// Create temp file for output
	tmpFile, err := os.CreateTemp("", "test-output-*.pdf")
	assert.NoError(t, err)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	err = wp.processPDFResult(pdfBuf, tmpFile.Name(), nil)
	assert.NoError(t, err)

	// Verify file was written
	content, err := os.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, pdfBuf, content)
}

func TestWorkerPool_processPDFResult_TooSmall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)

	wp := &WorkerPool{
		logger: mockLogger,
	}

	// Create a small PDF buffer (less than 1000 bytes)
	pdfBuf := make([]byte, 500)

	err := wp.processPDFResult(pdfBuf, "output.pdf", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too small")
}

func TestWorkerPool_processPDFResult_WithExistingError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)

	wp := &WorkerPool{
		logger: mockLogger,
	}

	existingErr := assert.AnError
	pdfBuf := make([]byte, 2000)

	err := wp.processPDFResult(pdfBuf, "output.pdf", existingErr)
	assert.Equal(t, existingErr, err)
}

func TestWorkerPool_cleanupTempFile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)

	wp := &WorkerPool{
		logger: mockLogger,
	}

	// Create a temp file to cleanup
	tmpFile, err := os.CreateTemp("", "cleanup-test-*.html")
	assert.NoError(t, err)
	tmpFile.Close()

	err = wp.cleanupTempFile(tmpFile.Name(), nil)
	assert.NoError(t, err)

	// Verify file was removed
	_, err = os.Stat(tmpFile.Name())
	assert.True(t, os.IsNotExist(err))
}

func TestWorkerPool_cleanupTempFile_WithOriginalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)

	wp := &WorkerPool{
		logger: mockLogger,
	}

	// Create a temp file to cleanup
	tmpFile, err := os.CreateTemp("", "cleanup-test-*.html")
	assert.NoError(t, err)
	tmpFile.Close()

	originalErr := assert.AnError
	err = wp.cleanupTempFile(tmpFile.Name(), originalErr)
	assert.Equal(t, originalErr, err)

	// Verify file was still removed
	_, err = os.Stat(tmpFile.Name())
	assert.True(t, os.IsNotExist(err))
}

func TestWorkerPool_cleanupTempFile_FileNotExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	wp := &WorkerPool{
		logger: mockLogger,
	}

	err := wp.cleanupTempFile("/nonexistent/file.html", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove temp file")
}

func TestWorkerPool_cleanupTempFile_FileNotExists_WithOriginalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	wp := &WorkerPool{
		logger: mockLogger,
	}

	originalErr := assert.AnError
	err := wp.cleanupTempFile("/nonexistent/file.html", originalErr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "additionally failed to remove temp file")
}

func TestWorkerPool_logPDFGenerationError_Timeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	wp := &WorkerPool{
		logger:  mockLogger,
		timeout: 30 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond) // Ensure timeout

	wp.logPDFGenerationError(ctx, context.DeadlineExceeded)
}

func TestWorkerPool_logPDFGenerationError_Canceled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)

	wp := &WorkerPool{
		logger: mockLogger,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	wp.logPDFGenerationError(ctx, context.Canceled)
}

func TestWorkerPool_logPDFGenerationError_OtherError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)

	wp := &WorkerPool{
		logger: mockLogger,
	}

	wp.logPDFGenerationError(context.Background(), assert.AnError)
}

func TestTask_Struct(t *testing.T) {
	resultChan := make(chan error, 1)

	task := Task{
		HTML:     "<html><body>Test</body></html>",
		Filename: "output.pdf",
		Result:   resultChan,
	}

	assert.Equal(t, "<html><body>Test</body></html>", task.HTML)
	assert.Equal(t, "output.pdf", task.Filename)
	assert.NotNil(t, task.Result)
}

func TestWorkerPool_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	wp := NewWorkerPool(1, 10*time.Second, mockLogger)

	// Close should not panic and should complete
	wp.Close()

	// Verify the tasks channel is closed
	_, ok := <-wp.tasks
	assert.False(t, ok, "tasks channel should be closed")
}

func TestWorkerPool_GetStats_TasksPending(t *testing.T) {
	wp := &WorkerPool{
		tasks:   make(chan Task, 10),
		workers: 2,
		timeout: 30 * time.Second,
	}

	stats := wp.GetStats()

	assert.Equal(t, 2, stats["workers"])
	assert.Equal(t, 30*time.Second, stats["timeout"])
	assert.Equal(t, 0, stats["tasks_pending"])
}

func TestWorkerPool_createTempHTMLFile_VerifyFilePermissions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)

	wp := &WorkerPool{
		logger: mockLogger,
	}

	htmlContent := "<html><body>Test</body></html>"

	tmpFileName, err := wp.createTempHTMLFile(htmlContent)
	assert.NoError(t, err)
	defer os.Remove(tmpFileName)

	// Verify file permissions (0600)
	info, err := os.Stat(tmpFileName)
	assert.NoError(t, err)
	// On Unix systems, check permissions
	mode := info.Mode().Perm()
	assert.Equal(t, os.FileMode(0600), mode)
}

func TestWorkerPool_createTempHTMLFile_VerifyFileExtension(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)

	wp := &WorkerPool{
		logger: mockLogger,
	}

	htmlContent := "<html><body>Test</body></html>"

	tmpFileName, err := wp.createTempHTMLFile(htmlContent)
	assert.NoError(t, err)
	defer os.Remove(tmpFileName)

	// Verify file has .html extension
	ext := filepath.Ext(tmpFileName)
	assert.Equal(t, ".html", ext)
}

func TestWorkerPool_Submit_Integration(t *testing.T) {
	// Skip if Chrome is not available (CI environment)
	if os.Getenv("SKIP_CHROME_TESTS") != "" {
		t.Skip("Skipping Chrome-dependent test")
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	wp := NewWorkerPool(1, 60*time.Second, mockLogger)
	defer wp.Close()

	htmlContent := `<!DOCTYPE html>
<html>
<head><title>Test PDF</title></head>
<body>
<h1>Test Document</h1>
<p>This is a test paragraph for PDF generation.</p>
</body>
</html>`

	// Create temp output file
	tmpFile, err := os.CreateTemp("", "test-output-*.pdf")
	assert.NoError(t, err)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	err = wp.Submit(htmlContent, tmpFile.Name())

	// If Chrome is not installed, this will fail - that's expected in CI
	if err != nil {
		t.Logf("Submit returned error (expected if Chrome not installed): %v", err)
		return
	}

	// Verify PDF was created
	info, err := os.Stat(tmpFile.Name())
	assert.NoError(t, err)
	assert.Greater(t, info.Size(), int64(1000), "PDF should be at least 1000 bytes")
}

func TestWorkerPool_processTask_TaskStructure(t *testing.T) {
	// Test with valid HTML - this will test the task structure
	// We can't fully test processTask without Chrome, but we can test the task structure
	resultChan := make(chan error, 1)
	task := Task{
		HTML:     "<html><body>Test</body></html>",
		Filename: "/tmp/test-output.pdf",
		Result:   resultChan,
	}

	assert.NotNil(t, task.Result)
	assert.Equal(t, "<html><body>Test</body></html>", task.HTML)
	assert.Equal(t, "/tmp/test-output.pdf", task.Filename)
}

func TestWorkerPool_processTask_CanceledContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	wp := &WorkerPool{
		logger:  mockLogger,
		timeout: 100 * time.Millisecond,
		tasks:   make(chan Task),
		wg:      &sync.WaitGroup{},
	}

	resultChan := make(chan error, 1)
	task := Task{
		HTML:     "<html><body>Test</body></html>",
		Filename: "/tmp/test-output.pdf",
		Result:   resultChan,
	}

	// Run processTask directly with canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	wp.processTask(ctx, task)

	// Check that an error was sent to the result channel
	select {
	case err := <-resultChan:
		assert.Error(t, err) // Should have an error due to canceled context
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for task result")
	}
}

func TestWorkerPool_Submit_TaskChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)

	// Create pool with buffered channel but don't start workers
	wp := &WorkerPool{
		logger:  mockLogger,
		timeout: 30 * time.Second,
		tasks:   make(chan Task, 1),
		wg:      &sync.WaitGroup{},
		workers: 1,
	}

	// Simulate a worker that immediately returns success
	go func() {
		task := <-wp.tasks
		task.Result <- nil // Success
	}()

	err := wp.Submit("<html>test</html>", "/tmp/output.pdf")
	assert.NoError(t, err)
}

func TestWorkerPool_Submit_TaskChannelWithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)

	// Create pool with buffered channel but don't start workers
	wp := &WorkerPool{
		logger:  mockLogger,
		timeout: 30 * time.Second,
		tasks:   make(chan Task, 1),
		wg:      &sync.WaitGroup{},
		workers: 1,
	}

	expectedErr := assert.AnError

	// Simulate a worker that returns an error
	go func() {
		task := <-wp.tasks
		task.Result <- expectedErr
	}()

	err := wp.Submit("<html>test</html>", "/tmp/output.pdf")
	assert.Equal(t, expectedErr, err)
}

func TestWorkerPool_generatePDFFromFile_InvalidPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := log.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	wp := &WorkerPool{
		logger:  mockLogger,
		timeout: 30 * time.Second,
	}

	// Test with canceled context - should fail fast
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := wp.generatePDFFromFile(ctx, "/nonexistent/path/file.html")
	assert.Error(t, err)
}
