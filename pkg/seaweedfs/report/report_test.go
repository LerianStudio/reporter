// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package report

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LerianStudio/reporter/v4/pkg/seaweedfs"
)

func TestSimpleRepository_Put_Success(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		data, _ := io.ReadAll(r.Body)
		gotBody = data
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	client := seaweedfs.NewSeaweedFSClient(srv.URL)
	repo := NewSimpleRepository(client, "reports")

	err := repo.Put(context.Background(), "obj.txt", "text/plain", []byte("hello"), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Fatalf("expected PUT, got %s", gotMethod)
	}
	if gotPath != "/reports/obj.txt" {
		t.Fatalf("expected path /reports/obj.txt, got %s", gotPath)
	}
	if string(gotBody) != "hello" {
		t.Fatalf("unexpected body: %q", string(gotBody))
	}
}

func TestSimpleRepository_Put_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := seaweedfs.NewSeaweedFSClient(srv.URL)
	repo := NewSimpleRepository(client, "reports")

	err := repo.Put(context.Background(), "obj.txt", "text/plain", []byte("hello"), "")
	if err == nil {
		t.Fatalf("expected put error, got nil")
	}
}

func TestSimpleRepository_Put_WithTTL_Success(t *testing.T) {
	var gotMethod, gotPath, gotQuery string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		data, _ := io.ReadAll(r.Body)
		gotBody = data
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	client := seaweedfs.NewSeaweedFSClient(srv.URL)
	repo := NewSimpleRepository(client, "reports")

	err := repo.Put(context.Background(), "temp.txt", "text/plain", []byte("temporary"), "1m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Fatalf("expected PUT, got %s", gotMethod)
	}
	if gotPath != "/reports/temp.txt" {
		t.Fatalf("expected path /reports/temp.txt, got %s", gotPath)
	}
	if gotQuery != "ttl=1m" {
		t.Fatalf("expected query ttl=1m, got %s", gotQuery)
	}
	if string(gotBody) != "temporary" {
		t.Fatalf("unexpected body: %q", string(gotBody))
	}
}

func TestSimpleRepository_Get_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/reports/obj.txt" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("world"))
	}))
	defer srv.Close()

	client := seaweedfs.NewSeaweedFSClient(srv.URL)
	repo := NewSimpleRepository(client, "reports")

	data, err := repo.Get(context.Background(), "obj.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "world" {
		t.Fatalf("unexpected data: %q", string(data))
	}
}

func TestSimpleRepository_Get_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "missing", http.StatusNotFound)
	}))
	defer srv.Close()

	client := seaweedfs.NewSeaweedFSClient(srv.URL)
	repo := NewSimpleRepository(client, "reports")

	_, err := repo.Get(context.Background(), "obj.txt")
	if err == nil {
		t.Fatalf("expected get error, got nil")
	}
}
