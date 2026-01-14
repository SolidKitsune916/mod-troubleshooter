package manifest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/mholt/archiver/v4"
)

// Common errors returned by the extractor.
var (
	ErrNoArchivePath     = errors.New("archive path is required")
	ErrArchiveNotFound   = errors.New("archive file not found")
	ErrUnsupportedFormat = errors.New("unsupported archive format")
	ErrExtractionFailed  = errors.New("extraction failed")
)

// Extractor extracts file manifests from mod archives.
type Extractor struct{}

// NewExtractor creates a new manifest extractor.
func NewExtractor() *Extractor {
	return &Extractor{}
}

// ExtractManifest extracts the file manifest from an archive without extracting file contents.
// This is a lightweight operation that only reads the archive directory.
func (e *Extractor) ExtractManifest(ctx context.Context, archivePath string) (*Manifest, error) {
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

	var entries []FileEntry

	// Walk the archive and collect file information
	err = extractor.Extract(ctx, input, func(ctx context.Context, f archiver.FileInfo) error {
		// Check for context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Skip directories
		if f.IsDir() {
			return nil
		}

		entry := NewFileEntry(f.NameInArchive, f.Size())
		entries = append(entries, entry)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExtractionFailed, err)
	}

	return NewManifest(entries), nil
}

// ExtractManifestWithHashes extracts the file manifest and computes content hashes.
// This is more expensive as it reads file contents to compute hashes.
func (e *Extractor) ExtractManifestWithHashes(ctx context.Context, archivePath string) (*Manifest, error) {
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

	var entries []FileEntry

	// Walk the archive and collect file information with content hashes
	err = extractor.Extract(ctx, input, func(ctx context.Context, f archiver.FileInfo) error {
		// Check for context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Skip directories
		if f.IsDir() {
			return nil
		}

		entry := NewFileEntry(f.NameInArchive, f.Size())

		// Compute content hash
		rc, err := f.Open()
		if err != nil {
			// If we can't open the file, just use path hash
			entries = append(entries, entry)
			return nil
		}
		defer rc.Close()

		hash := sha256.New()
		if _, err := io.Copy(hash, rc); err != nil {
			// If we can't read the file, just use path hash
			entries = append(entries, entry)
			return nil
		}

		entry.Hash = hex.EncodeToString(hash.Sum(nil))
		entries = append(entries, entry)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExtractionFailed, err)
	}

	return NewManifest(entries), nil
}

// ExtractManifestFiltered extracts the manifest only for files matching the filter function.
func (e *Extractor) ExtractManifestFiltered(ctx context.Context, archivePath string, filter func(FileEntry) bool) (*Manifest, error) {
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

	var entries []FileEntry

	// Walk the archive and collect matching file information
	err = extractor.Extract(ctx, input, func(ctx context.Context, f archiver.FileInfo) error {
		// Check for context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Skip directories
		if f.IsDir() {
			return nil
		}

		entry := NewFileEntry(f.NameInArchive, f.Size())

		// Apply filter
		if filter != nil && !filter(entry) {
			return nil
		}

		entries = append(entries, entry)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExtractionFailed, err)
	}

	return NewManifest(entries), nil
}

// FilterByType returns a filter function that matches files of the given type.
func FilterByType(fileType FileType) func(FileEntry) bool {
	return func(entry FileEntry) bool {
		return entry.Type == fileType
	}
}

// FilterByExtension returns a filter function that matches files with the given extension.
func FilterByExtension(extension string) func(FileEntry) bool {
	return func(entry FileEntry) bool {
		return entry.Extension == extension
	}
}

// FilterByDirectory returns a filter function that matches files in the given directory.
func FilterByDirectory(directory string) func(FileEntry) bool {
	normalizedDir := NormalizePath(directory)
	return func(entry FileEntry) bool {
		return entry.Directory == normalizedDir
	}
}

// FilterByPathPrefix returns a filter function that matches files with paths starting with prefix.
func FilterByPathPrefix(prefix string) func(FileEntry) bool {
	normalizedPrefix := NormalizePath(prefix)
	return func(entry FileEntry) bool {
		return len(entry.Path) >= len(normalizedPrefix) &&
			entry.Path[:len(normalizedPrefix)] == normalizedPrefix
	}
}
