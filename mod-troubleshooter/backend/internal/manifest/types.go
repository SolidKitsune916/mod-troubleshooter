package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"strings"
)

// FileType represents the category of a file based on extension.
type FileType string

const (
	// FileTypePlugin represents plugin files (.esp, .esm, .esl).
	FileTypePlugin FileType = "plugin"
	// FileTypeMesh represents 3D mesh files (.nif).
	FileTypeMesh FileType = "mesh"
	// FileTypeTexture represents texture files (.dds, .png, etc.).
	FileTypeTexture FileType = "texture"
	// FileTypeSound represents sound/audio files (.wav, .xwm, etc.).
	FileTypeSound FileType = "sound"
	// FileTypeScript represents script files (.pex, .psc).
	FileTypeScript FileType = "script"
	// FileTypeInterface represents interface files (.swf).
	FileTypeInterface FileType = "interface"
	// FileTypeSEQ represents sequence files (.seq).
	FileTypeSEQ FileType = "seq"
	// FileTypeBSA represents archive files (.bsa, .ba2).
	FileTypeBSA FileType = "bsa"
	// FileTypeOther represents all other file types.
	FileTypeOther FileType = "other"
)

// FileEntry represents a single file in a mod archive.
type FileEntry struct {
	// Path is the normalized path within the archive (forward slashes, lowercase).
	Path string `json:"path"`
	// OriginalPath is the original path as it appears in the archive.
	OriginalPath string `json:"originalPath"`
	// Size is the uncompressed file size in bytes.
	Size int64 `json:"size"`
	// Hash is the SHA-256 hash of the file path (used for dedup, not content).
	// For content hashing, use ComputeContentHash.
	Hash string `json:"hash"`
	// Type is the file category based on extension.
	Type FileType `json:"type"`
	// Extension is the lowercase file extension including the dot.
	Extension string `json:"extension"`
	// Directory is the normalized directory path.
	Directory string `json:"directory"`
	// Filename is the filename without directory.
	Filename string `json:"filename"`
}

// Manifest represents the complete file listing from a mod archive.
type Manifest struct {
	// Files is the list of all files in the archive.
	Files []FileEntry `json:"files"`
	// TotalSize is the sum of all file sizes.
	TotalSize int64 `json:"totalSize"`
	// TotalCount is the number of files.
	TotalCount int `json:"totalCount"`
	// ByType contains counts grouped by file type.
	ByType map[FileType]int `json:"byType"`
	// ByExtension contains counts grouped by extension.
	ByExtension map[string]int `json:"byExtension"`
}

// NormalizePath converts a path to a canonical form for comparison.
// - Converts backslashes to forward slashes.
// - Converts to lowercase.
// - Removes leading/trailing slashes.
// - Cleans the path (removes . and ..).
func NormalizePath(path string) string {
	// Convert backslashes to forward slashes
	normalized := strings.ReplaceAll(path, "\\", "/")
	// Convert to lowercase for case-insensitive comparison
	normalized = strings.ToLower(normalized)
	// Clean the path
	normalized = filepath.ToSlash(filepath.Clean(normalized))
	// Remove leading slash if present
	normalized = strings.TrimPrefix(normalized, "/")
	// Remove trailing slash if present
	normalized = strings.TrimSuffix(normalized, "/")
	return normalized
}

// ComputePathHash computes a SHA-256 hash of the normalized path.
// This is used for deduplication detection based on file path.
func ComputePathHash(normalizedPath string) string {
	hash := sha256.Sum256([]byte(normalizedPath))
	return hex.EncodeToString(hash[:])
}

// DetermineFileType determines the file type based on extension.
func DetermineFileType(extension string) FileType {
	ext := strings.ToLower(extension)
	switch ext {
	case ".esp", ".esm", ".esl":
		return FileTypePlugin
	case ".nif":
		return FileTypeMesh
	case ".dds", ".png", ".tga", ".bmp", ".jpg", ".jpeg":
		return FileTypeTexture
	case ".wav", ".xwm", ".fuz", ".lip":
		return FileTypeSound
	case ".pex", ".psc":
		return FileTypeScript
	case ".swf":
		return FileTypeInterface
	case ".seq":
		return FileTypeSEQ
	case ".bsa", ".ba2":
		return FileTypeBSA
	default:
		return FileTypeOther
	}
}

// NewFileEntry creates a new FileEntry with computed fields.
func NewFileEntry(originalPath string, size int64) FileEntry {
	normalized := NormalizePath(originalPath)
	ext := strings.ToLower(filepath.Ext(originalPath))
	dir := filepath.ToSlash(filepath.Dir(normalized))
	if dir == "." {
		dir = ""
	}
	filename := filepath.Base(normalized)

	return FileEntry{
		Path:         normalized,
		OriginalPath: originalPath,
		Size:         size,
		Hash:         ComputePathHash(normalized),
		Type:         DetermineFileType(ext),
		Extension:    ext,
		Directory:    dir,
		Filename:     filename,
	}
}

// NewManifest creates a new Manifest from a list of file entries.
func NewManifest(entries []FileEntry) *Manifest {
	m := &Manifest{
		Files:       entries,
		TotalCount:  len(entries),
		ByType:      make(map[FileType]int),
		ByExtension: make(map[string]int),
	}

	for _, entry := range entries {
		m.TotalSize += entry.Size
		m.ByType[entry.Type]++
		if entry.Extension != "" {
			m.ByExtension[entry.Extension]++
		}
	}

	return m
}

// GetFilesByType returns all files of a specific type.
func (m *Manifest) GetFilesByType(fileType FileType) []FileEntry {
	var result []FileEntry
	for _, entry := range m.Files {
		if entry.Type == fileType {
			result = append(result, entry)
		}
	}
	return result
}

// GetFilesByDirectory returns all files in a specific directory.
// Use empty string "" to get files in the root directory.
func (m *Manifest) GetFilesByDirectory(directory string) []FileEntry {
	normalizedDir := NormalizePath(directory)
	// Handle root directory case: NormalizePath("") returns "." but entry.Directory is ""
	if normalizedDir == "." {
		normalizedDir = ""
	}
	var result []FileEntry
	for _, entry := range m.Files {
		if entry.Directory == normalizedDir {
			result = append(result, entry)
		}
	}
	return result
}

// GetFilesByExtension returns all files with a specific extension.
func (m *Manifest) GetFilesByExtension(extension string) []FileEntry {
	ext := strings.ToLower(extension)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	var result []FileEntry
	for _, entry := range m.Files {
		if entry.Extension == ext {
			result = append(result, entry)
		}
	}
	return result
}

// HasFile checks if a file exists in the manifest by normalized path.
func (m *Manifest) HasFile(path string) bool {
	normalizedPath := NormalizePath(path)
	for _, entry := range m.Files {
		if entry.Path == normalizedPath {
			return true
		}
	}
	return false
}

// GetFile returns a file entry by normalized path, or nil if not found.
func (m *Manifest) GetFile(path string) *FileEntry {
	normalizedPath := NormalizePath(path)
	for i := range m.Files {
		if m.Files[i].Path == normalizedPath {
			return &m.Files[i]
		}
	}
	return nil
}
