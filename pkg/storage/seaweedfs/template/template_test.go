package template

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LerianStudio/reporter/v4/pkg/storage/seaweedfs"
)

func TestSimpleRepository_Get_AppendsTplExtension(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("tpl-data"))
	}))
	defer srv.Close()

	client := seaweedfs.NewSeaweedFSClient(srv.URL)
	repo := NewSimpleRepository(client, "templates")

	data, err := repo.Get(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/templates/abc123.tpl" {
		t.Fatalf("expected path /templates/abc123.tpl, got %s", gotPath)
	}
	if string(data) != "tpl-data" {
		t.Fatalf("unexpected data: %q", string(data))
	}
}

func TestSimpleRepository_Get_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "missing", http.StatusNotFound)
	}))
	defer srv.Close()

	client := seaweedfs.NewSeaweedFSClient(srv.URL)
	repo := NewSimpleRepository(client, "templates")

	_, err := repo.Get(context.Background(), "abc123")
	if err == nil {
		t.Fatalf("expected get error, got nil")
	}
}

func TestSimpleRepository_Put_Success(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = b
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	client := seaweedfs.NewSeaweedFSClient(srv.URL)
	repo := NewSimpleRepository(client, "templates")

	err := repo.Put(context.Background(), "folder/file.tpl", "text/plain", []byte("content"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Fatalf("expected PUT, got %s", gotMethod)
	}
	if gotPath != "/templates/folder/file.tpl" {
		t.Fatalf("expected path /templates/folder/file.tpl, got %s", gotPath)
	}
	if string(gotBody) != "content" {
		t.Fatalf("unexpected body: %q", string(gotBody))
	}
}

func TestSimpleRepository_Put_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := seaweedfs.NewSeaweedFSClient(srv.URL)
	repo := NewSimpleRepository(client, "templates")

	err := repo.Put(context.Background(), "file.tpl", "text/plain", []byte("x"))
	if err == nil {
		t.Fatalf("expected put error, got nil")
	}
}
