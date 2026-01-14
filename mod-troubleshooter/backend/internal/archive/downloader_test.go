package archive

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewDownloader(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		d, err := NewDownloader(DownloaderConfig{})
		if err != nil {
			t.Fatalf("NewDownloader() error = %v", err)
		}
		if d == nil {
			t.Fatal("NewDownloader() returned nil")
		}
		defer d.Cleanup()

		if d.tempDir == "" {
			t.Error("tempDir should not be empty with default config")
		}
		if d.httpClient == nil {
			t.Error("httpClient should not be nil with default config")
		}
		if d.userAgent != "ModTroubleshooter/1.0" {
			t.Errorf("userAgent = %q, want %q", d.userAgent, "ModTroubleshooter/1.0")
		}
	})

	t.Run("custom config", func(t *testing.T) {
		customDir := t.TempDir()
		customClient := &http.Client{Timeout: 5 * time.Minute}

		d, err := NewDownloader(DownloaderConfig{
			TempDir:     customDir,
			HTTPClient:  customClient,
			MaxFileSize: 1024 * 1024,
			UserAgent:   "TestAgent/1.0",
		})
		if err != nil {
			t.Fatalf("NewDownloader() error = %v", err)
		}
		defer d.Cleanup()

		if d.tempDir != customDir {
			t.Errorf("tempDir = %q, want %q", d.tempDir, customDir)
		}
		if d.httpClient != customClient {
			t.Error("httpClient should be the custom client")
		}
		if d.maxFileSize != 1024*1024 {
			t.Errorf("maxFileSize = %d, want %d", d.maxFileSize, 1024*1024)
		}
		if d.userAgent != "TestAgent/1.0" {
			t.Errorf("userAgent = %q, want %q", d.userAgent, "TestAgent/1.0")
		}
	})
}

func TestDownloader_Download(t *testing.T) {
	t.Run("successful download", func(t *testing.T) {
		// Set up test server
		content := "Hello, World! This is test content."
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(content))
		}))
		defer server.Close()

		d, err := NewDownloader(DownloaderConfig{TempDir: t.TempDir()})
		if err != nil {
			t.Fatalf("NewDownloader() error = %v", err)
		}
		defer d.Cleanup()

		result, err := d.Download(context.Background(), server.URL+"/test-file.zip", nil)
		if err != nil {
			t.Fatalf("Download() error = %v", err)
		}

		if result.Size != int64(len(content)) {
			t.Errorf("Size = %d, want %d", result.Size, len(content))
		}
		if result.ContentType != "application/octet-stream" {
			t.Errorf("ContentType = %q, want %q", result.ContentType, "application/octet-stream")
		}
		if !strings.HasSuffix(result.FilePath, "test-file.zip") {
			t.Errorf("FilePath = %q, should end with test-file.zip", result.FilePath)
		}

		// Verify file contents
		data, err := os.ReadFile(result.FilePath)
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}
		if string(data) != content {
			t.Errorf("file content = %q, want %q", string(data), content)
		}
	})

	t.Run("empty URL", func(t *testing.T) {
		d, err := NewDownloader(DownloaderConfig{TempDir: t.TempDir()})
		if err != nil {
			t.Fatalf("NewDownloader() error = %v", err)
		}
		defer d.Cleanup()

		_, err = d.Download(context.Background(), "", nil)
		if err != ErrNoURL {
			t.Errorf("Download() error = %v, want %v", err, ErrNoURL)
		}
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		d, err := NewDownloader(DownloaderConfig{TempDir: t.TempDir()})
		if err != nil {
			t.Fatalf("NewDownloader() error = %v", err)
		}
		defer d.Cleanup()

		_, err = d.Download(context.Background(), server.URL+"/file.zip", nil)
		if err == nil {
			t.Error("Download() should return error for 500 status")
		}
		if !strings.Contains(err.Error(), "500") {
			t.Errorf("error should contain status code, got: %v", err)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			time.Sleep(2 * time.Second)
			w.Write([]byte("content"))
		}))
		defer server.Close()

		d, err := NewDownloader(DownloaderConfig{TempDir: t.TempDir()})
		if err != nil {
			t.Fatalf("NewDownloader() error = %v", err)
		}
		defer d.Cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err = d.Download(ctx, server.URL+"/file.zip", nil)
		if err == nil {
			t.Error("Download() should return error when context is cancelled")
		}
	})

	t.Run("progress callback", func(t *testing.T) {
		content := strings.Repeat("x", 10000)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "10000")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(content))
		}))
		defer server.Close()

		d, err := NewDownloader(DownloaderConfig{TempDir: t.TempDir()})
		if err != nil {
			t.Fatalf("NewDownloader() error = %v", err)
		}
		defer d.Cleanup()

		var callCount int32
		var lastDownloaded int64

		_, err = d.Download(context.Background(), server.URL+"/file.zip", func(downloaded, total int64) {
			atomic.AddInt32(&callCount, 1)
			lastDownloaded = downloaded
		})
		if err != nil {
			t.Fatalf("Download() error = %v", err)
		}

		if callCount == 0 {
			t.Error("progress callback was never called")
		}
		if lastDownloaded != 10000 {
			t.Errorf("final downloaded = %d, want 10000", lastDownloaded)
		}
	})

	t.Run("file size limit - content-length", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "10000")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		d, err := NewDownloader(DownloaderConfig{
			TempDir:     t.TempDir(),
			MaxFileSize: 1000,
		})
		if err != nil {
			t.Fatalf("NewDownloader() error = %v", err)
		}
		defer d.Cleanup()

		_, err = d.Download(context.Background(), server.URL+"/file.zip", nil)
		if err == nil {
			t.Error("Download() should return error when file exceeds size limit")
		}
		if !strings.Contains(err.Error(), "exceeds") {
			t.Errorf("error should mention size limit, got: %v", err)
		}
	})

	t.Run("file size limit - streaming", func(t *testing.T) {
		content := strings.Repeat("x", 5000)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Don't set Content-Length to test streaming limit
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(content))
		}))
		defer server.Close()

		d, err := NewDownloader(DownloaderConfig{
			TempDir:     t.TempDir(),
			MaxFileSize: 1000,
		})
		if err != nil {
			t.Fatalf("NewDownloader() error = %v", err)
		}
		defer d.Cleanup()

		_, err = d.Download(context.Background(), server.URL+"/file.zip", nil)
		if err == nil {
			t.Error("Download() should return error when file exceeds size limit during streaming")
		}
	})
}

func TestDownloader_Cleanup(t *testing.T) {
	t.Run("cleanup removes temp directories", func(t *testing.T) {
		content := "test content"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(content))
		}))
		defer server.Close()

		tempDir := t.TempDir()
		d, err := NewDownloader(DownloaderConfig{TempDir: tempDir})
		if err != nil {
			t.Fatalf("NewDownloader() error = %v", err)
		}

		// Download a file
		result, err := d.Download(context.Background(), server.URL+"/file.zip", nil)
		if err != nil {
			t.Fatalf("Download() error = %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(result.FilePath); os.IsNotExist(err) {
			t.Fatal("downloaded file should exist")
		}

		downloadDir := filepath.Dir(result.FilePath)

		// Cleanup
		err = d.Cleanup()
		if err != nil {
			t.Fatalf("Cleanup() error = %v", err)
		}

		// Verify file is removed
		if _, err := os.Stat(downloadDir); !os.IsNotExist(err) {
			t.Error("temp directory should be removed after cleanup")
		}
	})
}

func TestDownloader_CleanupPath(t *testing.T) {
	t.Run("cleanup specific path", func(t *testing.T) {
		content := "test content"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(content))
		}))
		defer server.Close()

		d, err := NewDownloader(DownloaderConfig{TempDir: t.TempDir()})
		if err != nil {
			t.Fatalf("NewDownloader() error = %v", err)
		}
		defer d.Cleanup()

		// Download two files
		result1, err := d.Download(context.Background(), server.URL+"/file1.zip", nil)
		if err != nil {
			t.Fatalf("Download() error = %v", err)
		}

		result2, err := d.Download(context.Background(), server.URL+"/file2.zip", nil)
		if err != nil {
			t.Fatalf("Download() error = %v", err)
		}

		// Clean up first file
		err = d.CleanupPath(result1.FilePath)
		if err != nil {
			t.Fatalf("CleanupPath() error = %v", err)
		}

		// First file should be gone
		if _, err := os.Stat(result1.FilePath); !os.IsNotExist(err) {
			t.Error("first file should be removed")
		}

		// Second file should still exist
		if _, err := os.Stat(result2.FilePath); os.IsNotExist(err) {
			t.Error("second file should still exist")
		}
	})
}

func TestExtractFilename(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "simple filename",
			url:      "https://example.com/downloads/file.zip",
			expected: "file.zip",
		},
		{
			name:     "with query string",
			url:      "https://example.com/file.zip?token=abc123",
			expected: "file.zip",
		},
		{
			name:     "complex path",
			url:      "https://example.com/path/to/mods/archive.7z",
			expected: "archive.7z",
		},
		{
			name:     "no filename",
			url:      "https://example.com/",
			expected: "",
		},
		{
			name:     "empty url",
			url:      "",
			expected: "",
		},
		{
			name:     "just domain",
			url:      "https://example.com",
			expected: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFilename(tt.url)
			if result != tt.expected {
				t.Errorf("extractFilename(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}
