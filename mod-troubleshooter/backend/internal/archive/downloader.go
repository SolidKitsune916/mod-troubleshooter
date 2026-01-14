package archive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Common errors returned by the downloader.
var (
	ErrNoURL           = errors.New("download URL is required")
	ErrDownloadFailed  = errors.New("download failed")
	ErrInvalidResponse = errors.New("invalid server response")
	ErrFileTooLarge    = errors.New("file exceeds maximum allowed size")
)

// ProgressCallback is called periodically during download with progress information.
// downloaded is the number of bytes downloaded so far.
// total is the total file size (-1 if unknown).
type ProgressCallback func(downloaded, total int64)

// DownloaderConfig holds configuration for the Downloader.
type DownloaderConfig struct {
	// TempDir is the directory for storing downloaded files.
	// If empty, os.TempDir() is used.
	TempDir string

	// HTTPClient is the HTTP client to use for downloads.
	// If nil, a default client with 10-minute timeout is used.
	HTTPClient *http.Client

	// MaxFileSize is the maximum allowed file size in bytes.
	// Zero or negative means no limit.
	MaxFileSize int64

	// UserAgent is the User-Agent header for download requests.
	UserAgent string
}

// Downloader handles downloading mod archives from URLs.
type Downloader struct {
	tempDir     string
	httpClient  *http.Client
	maxFileSize int64
	userAgent   string

	mu       sync.Mutex
	tempDirs []string // Track created temp directories for cleanup
}

// NewDownloader creates a new archive downloader with the given configuration.
func NewDownloader(cfg DownloaderConfig) (*Downloader, error) {
	tempDir := cfg.TempDir
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Minute, // Large files may take a while
		}
	}

	userAgent := cfg.UserAgent
	if userAgent == "" {
		userAgent = "ModTroubleshooter/1.0"
	}

	return &Downloader{
		tempDir:     tempDir,
		httpClient:  httpClient,
		maxFileSize: cfg.MaxFileSize,
		userAgent:   userAgent,
		tempDirs:    make([]string, 0),
	}, nil
}

// DownloadResult contains information about a completed download.
type DownloadResult struct {
	// FilePath is the path to the downloaded file.
	FilePath string

	// Size is the size of the downloaded file in bytes.
	Size int64

	// ContentType is the Content-Type header from the response.
	ContentType string
}

// Download downloads a file from the given URL and returns the path to the downloaded file.
// The file is stored in a temporary directory that should be cleaned up after use.
// If onProgress is not nil, it will be called periodically with download progress.
func (d *Downloader) Download(ctx context.Context, url string, onProgress ProgressCallback) (*DownloadResult, error) {
	if url == "" {
		return nil, ErrNoURL
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", d.userAgent)

	// Execute request
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrInvalidResponse, resp.StatusCode)
	}

	// Check file size against limit
	contentLength := resp.ContentLength
	if d.maxFileSize > 0 && contentLength > 0 && contentLength > d.maxFileSize {
		return nil, fmt.Errorf("%w: %d bytes exceeds limit of %d bytes", ErrFileTooLarge, contentLength, d.maxFileSize)
	}

	// Create temp directory for this download
	downloadDir, err := os.MkdirTemp(d.tempDir, "mod-download-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	// Track the temp directory for cleanup
	d.mu.Lock()
	d.tempDirs = append(d.tempDirs, downloadDir)
	d.mu.Unlock()

	// Extract filename from URL or use default
	filename := extractFilename(url)
	if filename == "" {
		filename = "download"
	}

	filePath := filepath.Join(downloadDir, filename)

	// Create destination file
	file, err := os.Create(filePath)
	if err != nil {
		os.RemoveAll(downloadDir)
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	// Download with progress tracking
	var downloaded int64
	var reader io.Reader = resp.Body

	// If we have a progress callback, wrap the reader
	if onProgress != nil {
		reader = &progressReader{
			reader:     resp.Body,
			total:      contentLength,
			onProgress: onProgress,
		}
	}

	// Also check max file size during download if Content-Length wasn't provided
	if d.maxFileSize > 0 {
		reader = &limitedReader{
			reader:   reader,
			maxSize:  d.maxFileSize,
			readSize: &downloaded,
		}
	}

	// Copy data to file
	written, err := io.Copy(file, reader)
	if err != nil {
		file.Close()
		os.RemoveAll(downloadDir)
		return nil, fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}

	return &DownloadResult{
		FilePath:    filePath,
		Size:        written,
		ContentType: resp.Header.Get("Content-Type"),
	}, nil
}

// Cleanup removes all temporary directories created by this downloader.
func (d *Downloader) Cleanup() error {
	d.mu.Lock()
	dirs := d.tempDirs
	d.tempDirs = make([]string, 0)
	d.mu.Unlock()

	var firstErr error
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// CleanupPath removes a specific download directory.
func (d *Downloader) CleanupPath(downloadDir string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Remove from tracking list
	for i, dir := range d.tempDirs {
		if dir == downloadDir || filepath.Dir(downloadDir) == dir {
			d.tempDirs = append(d.tempDirs[:i], d.tempDirs[i+1:]...)
			break
		}
	}

	// Get the parent temp directory for the file
	parentDir := downloadDir
	if !isDirectory(downloadDir) {
		parentDir = filepath.Dir(downloadDir)
	}

	return os.RemoveAll(parentDir)
}

// progressReader wraps an io.Reader and reports progress.
type progressReader struct {
	reader     io.Reader
	total      int64
	downloaded int64
	onProgress ProgressCallback
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.downloaded += int64(n)
		pr.onProgress(pr.downloaded, pr.total)
	}
	return n, err
}

// limitedReader wraps an io.Reader and enforces a maximum size.
type limitedReader struct {
	reader   io.Reader
	maxSize  int64
	readSize *int64
}

func (lr *limitedReader) Read(p []byte) (int, error) {
	n, err := lr.reader.Read(p)
	if n > 0 {
		*lr.readSize += int64(n)
		if *lr.readSize > lr.maxSize {
			return n, fmt.Errorf("%w: exceeded %d bytes", ErrFileTooLarge, lr.maxSize)
		}
	}
	return n, err
}

// extractFilename extracts the filename from a URL path.
func extractFilename(url string) string {
	// Find the last slash
	lastSlash := -1
	for i := len(url) - 1; i >= 0; i-- {
		if url[i] == '/' {
			lastSlash = i
			break
		}
	}

	if lastSlash == -1 || lastSlash == len(url)-1 {
		return ""
	}

	filename := url[lastSlash+1:]

	// Remove query string if present
	for i, c := range filename {
		if c == '?' {
			filename = filename[:i]
			break
		}
	}

	return filename
}

// isDirectory checks if the given path is a directory.
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
