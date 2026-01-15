package seaweedfs

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSeaweedFSClient(t *testing.T) {
	client := NewSeaweedFSClient("http://localhost:8080")

	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8080", client.baseURL)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
}

func TestUploadFile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/test/file.txt", r.URL.Path)
		assert.Equal(t, "application/octet-stream", r.Header.Get("Content-Type"))

		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, "test content", string(body))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	err := client.UploadFile(context.Background(), "/test/file.txt", []byte("test content"))

	assert.NoError(t, err)
}

func TestUploadFile_Created(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	err := client.UploadFile(context.Background(), "/test/file.txt", []byte("test content"))

	assert.NoError(t, err)
}

func TestUploadFile_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	err := client.UploadFile(context.Background(), "/test/file.txt", []byte("test content"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upload failed with status 500")
}

func TestUploadFileWithTTL_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "1h", r.URL.Query().Get("ttl"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	err := client.UploadFileWithTTL(context.Background(), "/test/file.txt", []byte("test content"), "1h")

	assert.NoError(t, err)
}

func TestUploadFileWithTTL_NoTTL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "", r.URL.Query().Get("ttl"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	err := client.UploadFileWithTTL(context.Background(), "/test/file.txt", []byte("test content"), "")

	assert.NoError(t, err)
}

func TestUploadFileWithTTL_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := client.UploadFileWithTTL(ctx, "/test/file.txt", []byte("test content"), "")

	assert.Error(t, err)
}

func TestDownloadFile_Success(t *testing.T) {
	expectedContent := []byte("downloaded content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/test/file.txt", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write(expectedContent)
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	data, err := client.DownloadFile(context.Background(), "/test/file.txt")

	assert.NoError(t, err)
	assert.Equal(t, expectedContent, data)
}

func TestDownloadFile_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("file not found"))
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	data, err := client.DownloadFile(context.Background(), "/test/nonexistent.txt")

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "download failed with status 404")
}

func TestDownloadFile_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	data, err := client.DownloadFile(ctx, "/test/file.txt")

	assert.Error(t, err)
	assert.Nil(t, data)
}

func TestDeleteFile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/test/file.txt", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	err := client.DeleteFile(context.Background(), "/test/file.txt")

	assert.NoError(t, err)
}

func TestDeleteFile_NoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	err := client.DeleteFile(context.Background(), "/test/file.txt")

	assert.NoError(t, err)
}

func TestDeleteFile_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("file not found"))
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	err := client.DeleteFile(context.Background(), "/test/nonexistent.txt")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed with status 404")
}

func TestDeleteFile_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.DeleteFile(ctx, "/test/file.txt")

	assert.Error(t, err)
}

func TestHealthCheck_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/status", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	err := client.HealthCheck(context.Background())

	assert.NoError(t, err)
}

func TestHealthCheck_Unhealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	err := client.HealthCheck(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "health check failed with status 503")
}

func TestHealthCheck_ServerDown(t *testing.T) {
	client := NewSeaweedFSClient("http://localhost:99999")
	err := client.HealthCheck(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "health check failed")
}

func TestHealthCheck_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.HealthCheck(ctx)

	assert.Error(t, err)
}

// Test with invalid URL
func TestUploadFile_InvalidURL(t *testing.T) {
	client := NewSeaweedFSClient("://invalid-url")
	err := client.UploadFile(context.Background(), "/test/file.txt", []byte("test"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create request")
}

func TestDownloadFile_InvalidURL(t *testing.T) {
	client := NewSeaweedFSClient("://invalid-url")
	data, err := client.DownloadFile(context.Background(), "/test/file.txt")

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "failed to create request")
}

func TestDeleteFile_InvalidURL(t *testing.T) {
	client := NewSeaweedFSClient("://invalid-url")
	err := client.DeleteFile(context.Background(), "/test/file.txt")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create request")
}

func TestHealthCheck_InvalidURL(t *testing.T) {
	client := NewSeaweedFSClient("://invalid-url")
	err := client.HealthCheck(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create health check request")
}

// Test empty content
func TestUploadFile_EmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.Empty(t, body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	err := client.UploadFile(context.Background(), "/test/empty.txt", []byte{})

	assert.NoError(t, err)
}

// Test large content
func TestUploadFile_LargeContent(t *testing.T) {
	largeContent := make([]byte, 1024*1024) // 1MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.Len(t, body, 1024*1024)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSeaweedFSClient(server.URL)
	err := client.UploadFile(context.Background(), "/test/large.bin", largeContent)

	assert.NoError(t, err)
}
