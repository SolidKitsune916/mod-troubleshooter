package archive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v4"
)

// Common errors returned by the extractor.
var (
	ErrNoArchivePath     = errors.New("archive path is required")
	ErrArchiveNotFound   = errors.New("archive file not found")
	ErrUnsupportedFormat = errors.New("unsupported archive format")
	ErrExtractionFailed  = errors.New("extraction failed")
	ErrPathNotFound      = errors.New("requested path not found in archive")
)

// ExtractorConfig holds configuration for the Extractor.
type ExtractorConfig struct {
	// TempDir is the directory for storing extracted files.
	// If empty, os.TempDir() is used.
	TempDir string

	// MaxFileSize is the maximum allowed size for a single extracted file in bytes.
	// Zero or negative means no limit.
	MaxFileSize int64

	// MaxTotalSize is the maximum allowed total size of all extracted files in bytes.
	// Zero or negative means no limit.
	MaxTotalSize int64
}

// Extractor handles extracting files from archive formats.
type Extractor struct {
	tempDir      string
	maxFileSize  int64
	maxTotalSize int64
}

// NewExtractor creates a new archive extractor with the given configuration.
func NewExtractor(cfg ExtractorConfig) (*Extractor, error) {
	tempDir := cfg.TempDir
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	return &Extractor{
		tempDir:      tempDir,
		maxFileSize:  cfg.MaxFileSize,
		maxTotalSize: cfg.MaxTotalSize,
	}, nil
}

// ExtractResult contains information about a completed extraction.
type ExtractResult struct {
	// OutputDir is the directory containing extracted files.
	OutputDir string

	// Files is a list of extracted file paths relative to OutputDir.
	Files []string

	// TotalSize is the total size of all extracted files in bytes.
	TotalSize int64
}

// Extract extracts all files from the archive to a temporary directory.
func (e *Extractor) Extract(ctx context.Context, archivePath string) (*ExtractResult, error) {
	return e.ExtractPaths(ctx, archivePath, nil)
}

// ExtractPaths extracts only files matching the given path prefixes from the archive.
// If pathPrefixes is nil or empty, all files are extracted.
// Path matching is case-insensitive to handle Windows-style archives.
func (e *Extractor) ExtractPaths(ctx context.Context, archivePath string, pathPrefixes []string) (*ExtractResult, error) {
	if archivePath == "" {
		return nil, ErrNoArchivePath
	}

	// Check if archive exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: %s", ErrArchiveNotFound, archivePath)
	}

	// Open the archive file
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("open archive: %w", err)
	}
	defer file.Close()

	// Identify the archive format
	format, input, err := archiver.Identify(ctx, archivePath, file)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedFormat, err)
	}

	// Ensure we have an extractor format
	extractor, ok := format.(archiver.Extractor)
	if !ok {
		return nil, fmt.Errorf("%w: format does not support extraction", ErrUnsupportedFormat)
	}

	// Create temp directory for extraction
	outputDir, err := os.MkdirTemp(e.tempDir, "mod-extract-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	// Normalize path prefixes for case-insensitive matching
	normalizedPrefixes := make([]string, len(pathPrefixes))
	for i, prefix := range pathPrefixes {
		normalizedPrefixes[i] = strings.ToLower(filepath.ToSlash(prefix))
	}

	var extractedFiles []string
	var totalSize int64

	// Extract files
	err = extractor.Extract(ctx, input, func(ctx context.Context, f archiver.FileInfo) error {
		// Check for context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Skip directories
		if f.IsDir() {
			return nil
		}

		// Get the file path within the archive
		filePath := f.NameInArchive
		normalizedPath := strings.ToLower(filepath.ToSlash(filePath))

		// Check if file matches any path prefix
		if len(normalizedPrefixes) > 0 {
			matched := false
			for _, prefix := range normalizedPrefixes {
				if strings.HasPrefix(normalizedPath, prefix) {
					matched = true
					break
				}
			}
			if !matched {
				return nil // Skip this file
			}
		}

		// Check file size limit
		if e.maxFileSize > 0 && f.Size() > e.maxFileSize {
			return fmt.Errorf("file %s exceeds max file size (%d > %d)", filePath, f.Size(), e.maxFileSize)
		}

		// Check total size limit
		if e.maxTotalSize > 0 && totalSize+f.Size() > e.maxTotalSize {
			return fmt.Errorf("extraction would exceed max total size (%d)", e.maxTotalSize)
		}

		// Create the destination path
		destPath := filepath.Join(outputDir, filePath)

		// Ensure the path is within the output directory (prevent zip slip)
		if !strings.HasPrefix(filepath.Clean(destPath), filepath.Clean(outputDir)) {
			return fmt.Errorf("invalid file path: %s", filePath)
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("create directory for %s: %w", filePath, err)
		}

		// Open the file from the archive
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open file %s in archive: %w", filePath, err)
		}
		defer rc.Close()

		// Create the destination file
		destFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("create file %s: %w", destPath, err)
		}
		defer destFile.Close()

		// Copy the file contents
		written, err := io.Copy(destFile, rc)
		if err != nil {
			return fmt.Errorf("extract file %s: %w", filePath, err)
		}

		extractedFiles = append(extractedFiles, filePath)
		totalSize += written

		return nil
	})

	if err != nil {
		// Clean up on error
		os.RemoveAll(outputDir)
		return nil, fmt.Errorf("%w: %v", ErrExtractionFailed, err)
	}

	return &ExtractResult{
		OutputDir: outputDir,
		Files:     extractedFiles,
		TotalSize: totalSize,
	}, nil
}

// ExtractFomod extracts only the fomod directory from the archive.
// This is a convenience method for FOMOD analysis.
func (e *Extractor) ExtractFomod(ctx context.Context, archivePath string) (*ExtractResult, error) {
	return e.ExtractPaths(ctx, archivePath, []string{"fomod/", "fomod\\"})
}

// ListFiles returns a list of all files in the archive without extracting.
func (e *Extractor) ListFiles(ctx context.Context, archivePath string) ([]string, error) {
	if archivePath == "" {
		return nil, ErrNoArchivePath
	}

	// Check if archive exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: %s", ErrArchiveNotFound, archivePath)
	}

	// Open the archive file
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("open archive: %w", err)
	}
	defer file.Close()

	// Identify the archive format
	format, input, err := archiver.Identify(ctx, archivePath, file)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedFormat, err)
	}

	// Ensure we have an extractor format
	extractor, ok := format.(archiver.Extractor)
	if !ok {
		return nil, fmt.Errorf("%w: format does not support extraction", ErrUnsupportedFormat)
	}

	var files []string

	// Walk the archive without extracting
	err = extractor.Extract(ctx, input, func(ctx context.Context, f archiver.FileInfo) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !f.IsDir() {
			files = append(files, f.NameInArchive)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("list archive: %w", err)
	}

	return files, nil
}

// HasFomod checks if the archive contains a fomod directory.
func (e *Extractor) HasFomod(ctx context.Context, archivePath string) (bool, error) {
	files, err := e.ListFiles(ctx, archivePath)
	if err != nil {
		return false, err
	}

	for _, file := range files {
		normalized := strings.ToLower(filepath.ToSlash(file))
		if strings.HasPrefix(normalized, "fomod/") {
			return true, nil
		}
	}

	return false, nil
}

// Cleanup removes an extraction output directory.
func (e *Extractor) Cleanup(outputDir string) error {
	if outputDir == "" {
		return nil
	}
	return os.RemoveAll(outputDir)
}
