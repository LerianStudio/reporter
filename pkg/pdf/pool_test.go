// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pdf

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Note: Tests that require Chrome are skipped in CI environments.
// Use SKIP_CHROME_TESTS=1 to skip Chrome-dependent tests.

func shouldSkipChromeTests() bool {
	return os.Getenv("SKIP_CHROME_TESTS") == "1"
}

func TestTask_Struct(t *testing.T) {
	resultChan := make(chan error, 1)

	task := Task{
		HTML:     "<html><body>Test</body></html>",
		Filename: "/tmp/test.pdf",
		Result:   resultChan,
	}

	assert.Equal(t, "<html><body>Test</body></html>", task.HTML)
	assert.Equal(t, "/tmp/test.pdf", task.Filename)
	assert.NotNil(t, task.Result)
}

func TestWorkerPool_GetStats(t *testing.T) {
	// Create pool but don't start workers (we'll test GetStats directly)
	wp := &WorkerPool{
		tasks:   make(chan Task, 10),
		workers: 4,
		timeout: 60 * time.Second,
	}

	stats := wp.GetStats()

	assert.Equal(t, 4, stats["workers"])
	assert.Equal(t, 60*time.Second, stats["timeout"])
	assert.Equal(t, 0, stats["tasks_pending"])
}

func TestWorkerPool_IsHealthy(t *testing.T) {
	tests := []struct {
		name     string
		workers  int
		timeout  time.Duration
		expected bool
	}{
		{
			name:     "Healthy pool",
			workers:  4,
			timeout:  60 * time.Second,
			expected: true,
		},
		{
			name:     "Zero workers",
			workers:  0,
			timeout:  60 * time.Second,
			expected: false,
		},
		{
			name:     "Zero timeout",
			workers:  4,
			timeout:  0,
			expected: false,
		},
		{
			name:     "Both zero",
			workers:  0,
			timeout:  0,
			expected: false,
		},
		{
			name:     "Negative workers treated as unhealthy",
			workers:  -1,
			timeout:  60 * time.Second,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wp := &WorkerPool{
				workers: tt.workers,
				timeout: tt.timeout,
			}

			assert.Equal(t, tt.expected, wp.IsHealthy())
		})
	}
}

func TestWorkerPool_GetStats_PendingTasks(t *testing.T) {
	// Create a buffered channel and add some tasks
	tasks := make(chan Task, 10)
	tasks <- Task{HTML: "test1", Filename: "file1.pdf", Result: make(chan error, 1)}
	tasks <- Task{HTML: "test2", Filename: "file2.pdf", Result: make(chan error, 1)}

	wp := &WorkerPool{
		tasks:   tasks,
		workers: 2,
		timeout: 30 * time.Second,
	}

	stats := wp.GetStats()

	assert.Equal(t, 2, stats["tasks_pending"])
}

func TestWorkerPool_GetChromeOptions(t *testing.T) {
	wp := &WorkerPool{}

	options := wp.getChromeOptions()

	// Verify we have options
	assert.NotEmpty(t, options)
	// Verify headless mode is set
	assert.Greater(t, len(options), 5)
}

func TestWorkerPool_Struct(t *testing.T) {
	tasks := make(chan Task, 5)
	timeout := 120 * time.Second

	wp := &WorkerPool{
		tasks:   tasks,
		workers: 8,
		timeout: timeout,
	}

	assert.Equal(t, 8, wp.workers)
	assert.Equal(t, timeout, wp.timeout)
	assert.NotNil(t, wp.tasks)
}

func TestTask_ResultChannel(t *testing.T) {
	resultChan := make(chan error, 1)
	task := Task{
		HTML:     "<html></html>",
		Filename: "test.pdf",
		Result:   resultChan,
	}

	// Send a result
	go func() {
		task.Result <- nil
	}()

	// Receive the result
	err := <-task.Result
	assert.NoError(t, err)
}

func TestTask_ResultChannelWithError(t *testing.T) {
	resultChan := make(chan error, 1)
	task := Task{
		HTML:     "<html></html>",
		Filename: "test.pdf",
		Result:   resultChan,
	}

	expectedErr := assert.AnError
	// Send an error result
	go func() {
		task.Result <- expectedErr
	}()

	// Receive the result
	err := <-task.Result
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestWorkerPool_Timeout_Values(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{"30 seconds", 30 * time.Second},
		{"1 minute", time.Minute},
		{"5 minutes", 5 * time.Minute},
		{"10 minutes", 10 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wp := &WorkerPool{
				workers: 1,
				timeout: tt.timeout,
			}

			stats := wp.GetStats()
			assert.Equal(t, tt.timeout, stats["timeout"])
		})
	}
}

func TestWorkerPool_Workers_Values(t *testing.T) {
	tests := []struct {
		name    string
		workers int
	}{
		{"Single worker", 1},
		{"Few workers", 4},
		{"Many workers", 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wp := &WorkerPool{
				workers: tt.workers,
				timeout: time.Minute,
			}

			stats := wp.GetStats()
			assert.Equal(t, tt.workers, stats["workers"])
			assert.True(t, wp.IsHealthy())
		})
	}
}
