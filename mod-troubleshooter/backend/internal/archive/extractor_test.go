package archive

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewExtractor(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ExtractorConfig
		wantErr bool
	}{
		{
			name:    "default config",
			cfg:     ExtractorConfig{},
			wantErr: false,
		},
		{
			name: "custom temp dir",
			cfg: ExtractorConfig{
				TempDir: os.TempDir(),
			},
			wantErr: false,
		},
		{
			name: "with size limits",
			cfg: ExtractorConfig{
				MaxFileSize:  1024 * 1024,
				MaxTotalSize: 10 * 1024 * 1024,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, err := NewExtractor(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewExtractor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && ext == nil {
				t.Error("NewExtractor() returned nil extractor")
			}
		})
	}
}

func TestExtractor_Extract(t *testing.T) {
	// Create a test zip file
	zipPath := createTestZip(t, map[string]string{
		"file1.txt":           "content1",
		"subdir/file2.txt":    "content2",
		"fomod/info.xml":      "<fomod><Name>Test</Name></fomod>",
		"fomod/ModuleConfig.xml": "<config/>",
	})
	defer os.Remove(zipPath)

	ext, err := NewExtractor(ExtractorConfig{})
	if err != nil {
		t.Fatalf("NewExtractor() error = %v", err)
	}

	ctx := context.Background()
	result, err := ext.Extract(ctx, zipPath)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	defer ext.Cleanup(result.OutputDir)

	// Verify extracted files
	if len(result.Files) != 4 {
		t.Errorf("Extract() got %d files, want 4", len(result.Files))
	}

	// Verify all files exist
	for _, file := range result.Files {
		fullPath := filepath.Join(result.OutputDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected file %s does not exist", file)
		}
	}
}

func TestExtractor_ExtractPaths(t *testing.T) {
	// Create a test zip file
	zipPath := createTestZip(t, map[string]string{
		"file1.txt":              "content1",
		"subdir/file2.txt":       "content2",
		"fomod/info.xml":         "<fomod><Name>Test</Name></fomod>",
		"fomod/ModuleConfig.xml": "<config/>",
		"fomod/images/test.png":  "fake image data",
	})
	defer os.Remove(zipPath)

	ext, err := NewExtractor(ExtractorConfig{})
	if err != nil {
		t.Fatalf("NewExtractor() error = %v", err)
	}

	ctx := context.Background()
	result, err := ext.ExtractPaths(ctx, zipPath, []string{"fomod/"})
	if err != nil {
		t.Fatalf("ExtractPaths() error = %v", err)
	}
	defer ext.Cleanup(result.OutputDir)

	// Should only extract fomod files
	if len(result.Files) != 3 {
		t.Errorf("ExtractPaths() got %d files, want 3", len(result.Files))
	}

	// Verify only fomod files were extracted
	for _, file := range result.Files {
		if !strings.HasPrefix(strings.ToLower(file), "fomod/") {
			t.Errorf("Unexpected file extracted: %s", file)
		}
	}
}

func TestExtractor_ExtractFomod(t *testing.T) {
	// Create a test zip file
	zipPath := createTestZip(t, map[string]string{
		"data/meshes/test.nif":   "mesh data",
		"fomod/info.xml":         "<fomod><Name>Test</Name></fomod>",
		"fomod/ModuleConfig.xml": "<config/>",
	})
	defer os.Remove(zipPath)

	ext, err := NewExtractor(ExtractorConfig{})
	if err != nil {
		t.Fatalf("NewExtractor() error = %v", err)
	}

	ctx := context.Background()
	result, err := ext.ExtractFomod(ctx, zipPath)
	if err != nil {
		t.Fatalf("ExtractFomod() error = %v", err)
	}
	defer ext.Cleanup(result.OutputDir)

	// Should only extract fomod files
	if len(result.Files) != 2 {
		t.Errorf("ExtractFomod() got %d files, want 2", len(result.Files))
	}
}

func TestExtractor_Extract_Errors(t *testing.T) {
	ext, err := NewExtractor(ExtractorConfig{})
	if err != nil {
		t.Fatalf("NewExtractor() error = %v", err)
	}

	ctx := context.Background()

	t.Run("empty path", func(t *testing.T) {
		_, err := ext.Extract(ctx, "")
		if err != ErrNoArchivePath {
			t.Errorf("Extract() error = %v, want ErrNoArchivePath", err)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := ext.Extract(ctx, "/nonexistent/archive.zip")
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("Extract() error = %v, want error containing 'not found'", err)
		}
	})

	t.Run("invalid archive", func(t *testing.T) {
		// Create a file that's not a valid archive
		tmpFile, err := os.CreateTemp("", "not-an-archive-*.txt")
		if err != nil {
			t.Fatal(err)
		}
		tmpFile.WriteString("this is not an archive")
		tmpFile.Close()
		defer os.Remove(tmpFile.Name())

		_, err = ext.Extract(ctx, tmpFile.Name())
		if err == nil || !strings.Contains(err.Error(), "unsupported") {
			t.Errorf("Extract() error = %v, want error containing 'unsupported'", err)
		}
	})
}

func TestExtractor_Extract_ContextCancellation(t *testing.T) {
	// Create a test zip file
	zipPath := createTestZip(t, map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
	})
	defer os.Remove(zipPath)

	ext, err := NewExtractor(ExtractorConfig{})
	if err != nil {
		t.Fatalf("NewExtractor() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = ext.Extract(ctx, zipPath)
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Extract() with cancelled context should fail, got error = %v", err)
	}
}

func TestExtractor_Extract_MaxFileSize(t *testing.T) {
	// Create a test zip file with a large file
	zipPath := createTestZip(t, map[string]string{
		"small.txt": "small",
		"large.txt": strings.Repeat("x", 1000),
	})
	defer os.Remove(zipPath)

	ext, err := NewExtractor(ExtractorConfig{
		MaxFileSize: 100, // Very small limit
	})
	if err != nil {
		t.Fatalf("NewExtractor() error = %v", err)
	}

	ctx := context.Background()
	_, err = ext.Extract(ctx, zipPath)
	if err == nil || !strings.Contains(err.Error(), "exceeds max file size") {
		t.Errorf("Extract() with file exceeding limit should fail, got error = %v", err)
	}
}

func TestExtractor_ListFiles(t *testing.T) {
	// Create a test zip file
	zipPath := createTestZip(t, map[string]string{
		"file1.txt":        "content1",
		"subdir/file2.txt": "content2",
		"fomod/info.xml":   "<fomod/>",
	})
	defer os.Remove(zipPath)

	ext, err := NewExtractor(ExtractorConfig{})
	if err != nil {
		t.Fatalf("NewExtractor() error = %v", err)
	}

	ctx := context.Background()
	files, err := ext.ListFiles(ctx, zipPath)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	if len(files) != 3 {
		t.Errorf("ListFiles() got %d files, want 3", len(files))
	}

	// Verify expected files are present
	fileSet := make(map[string]bool)
	for _, f := range files {
		fileSet[f] = true
	}

	expected := []string{"file1.txt", "subdir/file2.txt", "fomod/info.xml"}
	for _, exp := range expected {
		if !fileSet[exp] {
			t.Errorf("ListFiles() missing expected file: %s", exp)
		}
	}
}

func TestExtractor_HasFomod(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		wantHas  bool
	}{
		{
			name: "has fomod",
			files: map[string]string{
				"data/test.esp":          "plugin data",
				"fomod/info.xml":         "<fomod/>",
				"fomod/ModuleConfig.xml": "<config/>",
			},
			wantHas: true,
		},
		{
			name: "no fomod",
			files: map[string]string{
				"data/test.esp": "plugin data",
				"readme.txt":    "readme",
			},
			wantHas: false,
		},
		{
			name: "fomod in filename but not directory",
			files: map[string]string{
				"fomod-readme.txt": "readme about fomod",
			},
			wantHas: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zipPath := createTestZip(t, tt.files)
			defer os.Remove(zipPath)

			ext, err := NewExtractor(ExtractorConfig{})
			if err != nil {
				t.Fatalf("NewExtractor() error = %v", err)
			}

			ctx := context.Background()
			hasFomod, err := ext.HasFomod(ctx, zipPath)
			if err != nil {
				t.Fatalf("HasFomod() error = %v", err)
			}

			if hasFomod != tt.wantHas {
				t.Errorf("HasFomod() = %v, want %v", hasFomod, tt.wantHas)
			}
		})
	}
}

func TestExtractor_Cleanup(t *testing.T) {
	ext, err := NewExtractor(ExtractorConfig{})
	if err != nil {
		t.Fatalf("NewExtractor() error = %v", err)
	}

	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "test-cleanup-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a file inside
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Cleanup
	if err := ext.Cleanup(tmpDir); err != nil {
		t.Errorf("Cleanup() error = %v", err)
	}

	// Verify directory is removed
	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		t.Error("Cleanup() did not remove directory")
	}
}

func TestExtractor_Cleanup_EmptyPath(t *testing.T) {
	ext, err := NewExtractor(ExtractorConfig{})
	if err != nil {
		t.Fatalf("NewExtractor() error = %v", err)
	}

	// Should not error on empty path
	if err := ext.Cleanup(""); err != nil {
		t.Errorf("Cleanup(\"\") error = %v, want nil", err)
	}
}

// createTestZip creates a temporary zip file with the given files.
// Returns the path to the created zip file.
func createTestZip(t *testing.T, files map[string]string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test-archive-*.zip")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	zipWriter := zip.NewWriter(tmpFile)

	for name, content := range files {
		w, err := zipWriter.Create(name)
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			t.Fatalf("Failed to create file in zip: %v", err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			t.Fatalf("Failed to write file content: %v", err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to close temp file: %v", err)
	}

	return tmpFile.Name()
}
