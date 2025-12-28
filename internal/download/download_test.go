package download_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/download"
)

func TestResult_Fields(t *testing.T) {
	t.Parallel()

	result := &download.Result{
		LocalPath: "/tmp/test-file",
		Cleanup:   func() {},
	}

	if result.LocalPath != "/tmp/test-file" {
		t.Errorf("LocalPath = %q, want %q", result.LocalPath, "/tmp/test-file")
	}

	if result.Cleanup == nil {
		t.Error("Cleanup should not be nil")
	}
}

func TestFromURL_Success(t *testing.T) {
	t.Parallel()

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "shelly-cli" {
			t.Errorf("User-Agent = %q, want %q", r.Header.Get("User-Agent"), "shelly-cli")
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test content")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	result, err := download.FromURL(ctx, server.URL)
	if err != nil {
		t.Fatalf("FromURL() error = %v", err)
	}
	defer result.Cleanup()

	if result.LocalPath == "" {
		t.Error("LocalPath should not be empty")
	}

	// Verify file exists and has correct content
	content, err := os.ReadFile(result.LocalPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}

	if string(content) != "test content" {
		t.Errorf("content = %q, want %q", string(content), "test content")
	}

	// Verify file is executable
	info, err := os.Stat(result.LocalPath)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	mode := info.Mode()
	if mode&0o100 == 0 {
		t.Error("downloaded file should be executable")
	}
}

func TestFromURL_HTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	ctx := context.Background()
	result, err := download.FromURL(ctx, server.URL)
	if err == nil {
		result.Cleanup()
		t.Error("FromURL() should return error for 404")
	}
}

func TestFromURL_ServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx := context.Background()
	result, err := download.FromURL(ctx, server.URL)
	if err == nil {
		result.Cleanup()
		t.Error("FromURL() should return error for 500")
	}
}

func TestFromURL_InvalidURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	result, err := download.FromURL(ctx, "not-a-valid-url")
	if err == nil {
		result.Cleanup()
		t.Error("FromURL() should return error for invalid URL")
	}
}

func TestFromURL_ContextCanceled(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("content")); err != nil {
			t.Logf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := download.FromURL(ctx, server.URL)
	if err == nil {
		result.Cleanup()
		t.Error("FromURL() should return error when context is canceled")
	}
}

func TestFromURL_Cleanup(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("test")); err != nil {
			t.Logf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	result, err := download.FromURL(ctx, server.URL)
	if err != nil {
		t.Fatalf("FromURL() error = %v", err)
	}

	localPath := result.LocalPath

	// File should exist before cleanup
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		t.Error("file should exist before cleanup")
	}

	// Call cleanup
	result.Cleanup()

	// File should be removed after cleanup
	if _, err := os.Stat(localPath); !os.IsNotExist(err) {
		t.Error("file should be removed after cleanup")
	}
}

func TestFromURL_NetworkError(t *testing.T) {
	t.Parallel()

	// Use a URL that will definitely fail to connect
	ctx := context.Background()
	result, err := download.FromURL(ctx, "http://localhost:1") // Port 1 is unlikely to be listening
	if err == nil {
		result.Cleanup()
		t.Error("FromURL() should return error for network error")
	}
}
