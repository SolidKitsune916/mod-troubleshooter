package manifest

import (
	"archive/zip"
	"context"
	"os"
	"strings"
	"testing"
)

func TestNewExtractor(t *testing.T) {
	ext := NewExtractor()
	if ext == nil {
		t.Error("NewExtractor() returned nil")
	}
}

func TestExtractor_ExtractManifest(t *testing.T) {
	// Create a test zip file
	zipPath := createTestZip(t, map[string]string{
		"test.esp":                   "plugin data",
		"meshes/test.nif":            "mesh data",
		"textures/test.dds":          "texture data",
		"Data/Meshes/Actor/test.nif": "another mesh",
	})
	defer os.Remove(zipPath)

	ext := NewExtractor()
	ctx := context.Background()

	manifest, err := ext.ExtractManifest(ctx, zipPath)
	if err != nil {
		t.Fatalf("ExtractManifest() error = %v", err)
	}

	if manifest.TotalCount != 4 {
		t.Errorf("TotalCount = %d, want 4", manifest.TotalCount)
	}

	// Verify file types are correctly identified
	if manifest.ByType[FileTypePlugin] != 1 {
		t.Errorf("Plugin count = %d, want 1", manifest.ByType[FileTypePlugin])
	}
	if manifest.ByType[FileTypeMesh] != 2 {
		t.Errorf("Mesh count = %d, want 2", manifest.ByType[FileTypeMesh])
	}
	if manifest.ByType[FileTypeTexture] != 1 {
		t.Errorf("Texture count = %d, want 1", manifest.ByType[FileTypeTexture])
	}

	// Verify paths are normalized
	if !manifest.HasFile("test.esp") {
		t.Error("Missing test.esp")
	}
	if !manifest.HasFile("meshes/test.nif") {
		t.Error("Missing meshes/test.nif")
	}
	if !manifest.HasFile("data/meshes/actor/test.nif") {
		t.Error("Missing data/meshes/actor/test.nif (should be lowercase)")
	}
}

func TestExtractor_ExtractManifest_Errors(t *testing.T) {
	ext := NewExtractor()
	ctx := context.Background()

	t.Run("empty path", func(t *testing.T) {
		_, err := ext.ExtractManifest(ctx, "")
		if err != ErrNoArchivePath {
			t.Errorf("ExtractManifest() error = %v, want ErrNoArchivePath", err)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := ext.ExtractManifest(ctx, "/nonexistent/archive.zip")
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("ExtractManifest() error = %v, want error containing 'not found'", err)
		}
	})

	t.Run("invalid archive", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "not-an-archive-*.txt")
		if err != nil {
			t.Fatal(err)
		}
		tmpFile.WriteString("this is not an archive")
		tmpFile.Close()
		defer os.Remove(tmpFile.Name())

		_, err = ext.ExtractManifest(ctx, tmpFile.Name())
		if err == nil || !strings.Contains(err.Error(), "unsupported") {
			t.Errorf("ExtractManifest() error = %v, want error containing 'unsupported'", err)
		}
	})
}

func TestExtractor_ExtractManifest_ContextCancellation(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"test.esp": "plugin data",
	})
	defer os.Remove(zipPath)

	ext := NewExtractor()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := ext.ExtractManifest(ctx, zipPath)
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("ExtractManifest() with cancelled context should fail, got error = %v", err)
	}
}

func TestExtractor_ExtractManifestWithHashes(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"test1.esp": "content1",
		"test2.esp": "content2",
		"same.esp":  "content1", // Same content as test1.esp
	})
	defer os.Remove(zipPath)

	ext := NewExtractor()
	ctx := context.Background()

	manifest, err := ext.ExtractManifestWithHashes(ctx, zipPath)
	if err != nil {
		t.Fatalf("ExtractManifestWithHashes() error = %v", err)
	}

	if manifest.TotalCount != 3 {
		t.Errorf("TotalCount = %d, want 3", manifest.TotalCount)
	}

	// Files with same content should have same hash
	test1 := manifest.GetFile("test1.esp")
	same := manifest.GetFile("same.esp")
	if test1 == nil || same == nil {
		t.Fatal("Expected files not found")
	}

	if test1.Hash != same.Hash {
		t.Error("Files with same content should have same hash")
	}

	// Files with different content should have different hashes
	test2 := manifest.GetFile("test2.esp")
	if test2 == nil {
		t.Fatal("test2.esp not found")
	}

	if test1.Hash == test2.Hash {
		t.Error("Files with different content should have different hashes")
	}
}

func TestExtractor_ExtractManifestFiltered(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"test.esp":          "plugin",
		"meshes/test.nif":   "mesh",
		"textures/test.dds": "texture",
		"scripts/test.pex":  "script",
	})
	defer os.Remove(zipPath)

	ext := NewExtractor()
	ctx := context.Background()

	t.Run("filter by type", func(t *testing.T) {
		manifest, err := ext.ExtractManifestFiltered(ctx, zipPath, FilterByType(FileTypeMesh))
		if err != nil {
			t.Fatalf("ExtractManifestFiltered() error = %v", err)
		}

		if manifest.TotalCount != 1 {
			t.Errorf("TotalCount = %d, want 1", manifest.TotalCount)
		}

		if !manifest.HasFile("meshes/test.nif") {
			t.Error("Expected meshes/test.nif")
		}
	})

	t.Run("filter by extension", func(t *testing.T) {
		manifest, err := ext.ExtractManifestFiltered(ctx, zipPath, FilterByExtension(".esp"))
		if err != nil {
			t.Fatalf("ExtractManifestFiltered() error = %v", err)
		}

		if manifest.TotalCount != 1 {
			t.Errorf("TotalCount = %d, want 1", manifest.TotalCount)
		}
	})

	t.Run("filter by directory", func(t *testing.T) {
		manifest, err := ext.ExtractManifestFiltered(ctx, zipPath, FilterByDirectory("meshes"))
		if err != nil {
			t.Fatalf("ExtractManifestFiltered() error = %v", err)
		}

		if manifest.TotalCount != 1 {
			t.Errorf("TotalCount = %d, want 1", manifest.TotalCount)
		}
	})

	t.Run("filter by path prefix", func(t *testing.T) {
		manifest, err := ext.ExtractManifestFiltered(ctx, zipPath, FilterByPathPrefix("textures"))
		if err != nil {
			t.Fatalf("ExtractManifestFiltered() error = %v", err)
		}

		if manifest.TotalCount != 1 {
			t.Errorf("TotalCount = %d, want 1", manifest.TotalCount)
		}
	})

	t.Run("nil filter extracts all", func(t *testing.T) {
		manifest, err := ext.ExtractManifestFiltered(ctx, zipPath, nil)
		if err != nil {
			t.Fatalf("ExtractManifestFiltered() error = %v", err)
		}

		if manifest.TotalCount != 4 {
			t.Errorf("TotalCount = %d, want 4", manifest.TotalCount)
		}
	})
}

func TestExtractor_LargeArchive(t *testing.T) {
	// Create archive with many files
	files := make(map[string]string)
	for i := 0; i < 100; i++ {
		files["meshes/test"+string(rune('a'+i%26))+".nif"] = "mesh data"
		files["textures/test"+string(rune('a'+i%26))+".dds"] = "texture data"
	}

	zipPath := createTestZip(t, files)
	defer os.Remove(zipPath)

	ext := NewExtractor()
	ctx := context.Background()

	manifest, err := ext.ExtractManifest(ctx, zipPath)
	if err != nil {
		t.Fatalf("ExtractManifest() error = %v", err)
	}

	// Due to key collisions in the map, we expect 52 unique files (26 meshes + 26 textures)
	if manifest.TotalCount < 52 {
		t.Errorf("TotalCount = %d, want at least 52", manifest.TotalCount)
	}
}

func TestExtractor_SpecialPaths(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"normal.esp":              "normal",
		"Data/with spaces.esp":   "spaces",
		"Data/special-chars.esp": "special",
		"深层/test.esp":            "unicode dir",
	})
	defer os.Remove(zipPath)

	ext := NewExtractor()
	ctx := context.Background()

	manifest, err := ext.ExtractManifest(ctx, zipPath)
	if err != nil {
		t.Fatalf("ExtractManifest() error = %v", err)
	}

	if manifest.TotalCount != 4 {
		t.Errorf("TotalCount = %d, want 4", manifest.TotalCount)
	}

	// All paths should be normalized
	for _, entry := range manifest.Files {
		if strings.Contains(entry.Path, "\\") {
			t.Errorf("Path %q contains backslash", entry.Path)
		}
	}
}

func TestFilterByType(t *testing.T) {
	filter := FilterByType(FileTypePlugin)

	pluginEntry := NewFileEntry("test.esp", 100)
	meshEntry := NewFileEntry("test.nif", 100)

	if !filter(pluginEntry) {
		t.Error("FilterByType(Plugin) should match .esp files")
	}

	if filter(meshEntry) {
		t.Error("FilterByType(Plugin) should not match .nif files")
	}
}

func TestFilterByExtension(t *testing.T) {
	filter := FilterByExtension(".esp")

	espEntry := NewFileEntry("test.esp", 100)
	esmEntry := NewFileEntry("test.esm", 100)

	if !filter(espEntry) {
		t.Error("FilterByExtension(.esp) should match .esp files")
	}

	if filter(esmEntry) {
		t.Error("FilterByExtension(.esp) should not match .esm files")
	}
}

func TestFilterByDirectory(t *testing.T) {
	filter := FilterByDirectory("meshes")

	meshEntry := NewFileEntry("meshes/test.nif", 100)
	textureEntry := NewFileEntry("textures/test.dds", 100)
	rootEntry := NewFileEntry("test.esp", 100)

	if !filter(meshEntry) {
		t.Error("FilterByDirectory(meshes) should match files in meshes/")
	}

	if filter(textureEntry) {
		t.Error("FilterByDirectory(meshes) should not match files in textures/")
	}

	if filter(rootEntry) {
		t.Error("FilterByDirectory(meshes) should not match files in root")
	}
}

func TestFilterByPathPrefix(t *testing.T) {
	filter := FilterByPathPrefix("meshes/actors")

	matchEntry := NewFileEntry("meshes/actors/test.nif", 100)
	partialEntry := NewFileEntry("meshes/weapons/test.nif", 100)

	if !filter(matchEntry) {
		t.Error("FilterByPathPrefix(meshes/actors) should match meshes/actors/test.nif")
	}

	if filter(partialEntry) {
		t.Error("FilterByPathPrefix(meshes/actors) should not match meshes/weapons/test.nif")
	}
}

// createTestZip creates a temporary zip file with the given files.
func createTestZip(t *testing.T, files map[string]string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test-manifest-*.zip")
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
