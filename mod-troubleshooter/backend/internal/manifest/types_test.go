package manifest

import (
	"testing"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "forward slashes",
			input:    "data/meshes/test.nif",
			expected: "data/meshes/test.nif",
		},
		{
			name:     "backslashes",
			input:    "data\\meshes\\test.nif",
			expected: "data/meshes/test.nif",
		},
		{
			name:     "mixed slashes",
			input:    "data\\meshes/test.nif",
			expected: "data/meshes/test.nif",
		},
		{
			name:     "uppercase",
			input:    "Data/Meshes/Test.NIF",
			expected: "data/meshes/test.nif",
		},
		{
			name:     "leading slash",
			input:    "/data/meshes/test.nif",
			expected: "data/meshes/test.nif",
		},
		{
			name:     "trailing slash",
			input:    "data/meshes/",
			expected: "data/meshes",
		},
		{
			name:     "dots in path",
			input:    "./data/../data/meshes/./test.nif",
			expected: "data/meshes/test.nif",
		},
		{
			name:     "empty string",
			input:    "",
			expected: ".",
		},
		{
			name:     "root only",
			input:    "/",
			expected: "",
		},
		{
			name:     "file in root",
			input:    "test.esp",
			expected: "test.esp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestComputePathHash(t *testing.T) {
	t.Run("consistent hashing", func(t *testing.T) {
		path := "data/meshes/test.nif"
		hash1 := ComputePathHash(path)
		hash2 := ComputePathHash(path)
		if hash1 != hash2 {
			t.Errorf("ComputePathHash() not consistent: %s != %s", hash1, hash2)
		}
	})

	t.Run("different paths different hashes", func(t *testing.T) {
		hash1 := ComputePathHash("data/meshes/test1.nif")
		hash2 := ComputePathHash("data/meshes/test2.nif")
		if hash1 == hash2 {
			t.Error("ComputePathHash() should produce different hashes for different paths")
		}
	})

	t.Run("hash format", func(t *testing.T) {
		hash := ComputePathHash("test.nif")
		if len(hash) != 64 { // SHA-256 produces 64 hex characters
			t.Errorf("ComputePathHash() hash length = %d, want 64", len(hash))
		}
	})
}

func TestDetermineFileType(t *testing.T) {
	tests := []struct {
		extension string
		expected  FileType
	}{
		{".esp", FileTypePlugin},
		{".esm", FileTypePlugin},
		{".esl", FileTypePlugin},
		{".ESP", FileTypePlugin},
		{".nif", FileTypeMesh},
		{".NIF", FileTypeMesh},
		{".dds", FileTypeTexture},
		{".png", FileTypeTexture},
		{".tga", FileTypeTexture},
		{".bmp", FileTypeTexture},
		{".jpg", FileTypeTexture},
		{".jpeg", FileTypeTexture},
		{".wav", FileTypeSound},
		{".xwm", FileTypeSound},
		{".fuz", FileTypeSound},
		{".lip", FileTypeSound},
		{".pex", FileTypeScript},
		{".psc", FileTypeScript},
		{".swf", FileTypeInterface},
		{".seq", FileTypeSEQ},
		{".bsa", FileTypeBSA},
		{".ba2", FileTypeBSA},
		{".txt", FileTypeOther},
		{".xml", FileTypeOther},
		{"", FileTypeOther},
	}

	for _, tt := range tests {
		t.Run(tt.extension, func(t *testing.T) {
			result := DetermineFileType(tt.extension)
			if result != tt.expected {
				t.Errorf("DetermineFileType(%q) = %v, want %v", tt.extension, result, tt.expected)
			}
		})
	}
}

func TestNewFileEntry(t *testing.T) {
	tests := []struct {
		name         string
		originalPath string
		size         int64
		wantPath     string
		wantDir      string
		wantFilename string
		wantExt      string
		wantType     FileType
	}{
		{
			name:         "plugin in data folder",
			originalPath: "Data\\Test.esp",
			size:         1024,
			wantPath:     "data/test.esp",
			wantDir:      "data",
			wantFilename: "test.esp",
			wantExt:      ".esp",
			wantType:     FileTypePlugin,
		},
		{
			name:         "mesh in subfolder",
			originalPath: "meshes/actors/character/test.nif",
			size:         2048,
			wantPath:     "meshes/actors/character/test.nif",
			wantDir:      "meshes/actors/character",
			wantFilename: "test.nif",
			wantExt:      ".nif",
			wantType:     FileTypeMesh,
		},
		{
			name:         "file in root",
			originalPath: "readme.txt",
			size:         100,
			wantPath:     "readme.txt",
			wantDir:      "",
			wantFilename: "readme.txt",
			wantExt:      ".txt",
			wantType:     FileTypeOther,
		},
		{
			name:         "texture with uppercase",
			originalPath: "Textures\\Test.DDS",
			size:         4096,
			wantPath:     "textures/test.dds",
			wantDir:      "textures",
			wantFilename: "test.dds",
			wantExt:      ".dds",
			wantType:     FileTypeTexture,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := NewFileEntry(tt.originalPath, tt.size)

			if entry.Path != tt.wantPath {
				t.Errorf("Path = %q, want %q", entry.Path, tt.wantPath)
			}
			if entry.OriginalPath != tt.originalPath {
				t.Errorf("OriginalPath = %q, want %q", entry.OriginalPath, tt.originalPath)
			}
			if entry.Size != tt.size {
				t.Errorf("Size = %d, want %d", entry.Size, tt.size)
			}
			if entry.Directory != tt.wantDir {
				t.Errorf("Directory = %q, want %q", entry.Directory, tt.wantDir)
			}
			if entry.Filename != tt.wantFilename {
				t.Errorf("Filename = %q, want %q", entry.Filename, tt.wantFilename)
			}
			if entry.Extension != tt.wantExt {
				t.Errorf("Extension = %q, want %q", entry.Extension, tt.wantExt)
			}
			if entry.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", entry.Type, tt.wantType)
			}
			if entry.Hash == "" {
				t.Error("Hash should not be empty")
			}
		})
	}
}

func TestNewManifest(t *testing.T) {
	entries := []FileEntry{
		NewFileEntry("data/test.esp", 1000),
		NewFileEntry("meshes/test.nif", 2000),
		NewFileEntry("textures/test.dds", 3000),
		NewFileEntry("textures/test2.dds", 4000),
	}

	manifest := NewManifest(entries)

	t.Run("total count", func(t *testing.T) {
		if manifest.TotalCount != 4 {
			t.Errorf("TotalCount = %d, want 4", manifest.TotalCount)
		}
	})

	t.Run("total size", func(t *testing.T) {
		if manifest.TotalSize != 10000 {
			t.Errorf("TotalSize = %d, want 10000", manifest.TotalSize)
		}
	})

	t.Run("by type counts", func(t *testing.T) {
		if manifest.ByType[FileTypePlugin] != 1 {
			t.Errorf("ByType[Plugin] = %d, want 1", manifest.ByType[FileTypePlugin])
		}
		if manifest.ByType[FileTypeMesh] != 1 {
			t.Errorf("ByType[Mesh] = %d, want 1", manifest.ByType[FileTypeMesh])
		}
		if manifest.ByType[FileTypeTexture] != 2 {
			t.Errorf("ByType[Texture] = %d, want 2", manifest.ByType[FileTypeTexture])
		}
	})

	t.Run("by extension counts", func(t *testing.T) {
		if manifest.ByExtension[".esp"] != 1 {
			t.Errorf("ByExtension[.esp] = %d, want 1", manifest.ByExtension[".esp"])
		}
		if manifest.ByExtension[".dds"] != 2 {
			t.Errorf("ByExtension[.dds] = %d, want 2", manifest.ByExtension[".dds"])
		}
	})
}

func TestManifest_GetFilesByType(t *testing.T) {
	entries := []FileEntry{
		NewFileEntry("test1.esp", 100),
		NewFileEntry("test2.esp", 200),
		NewFileEntry("test.nif", 300),
	}
	manifest := NewManifest(entries)

	plugins := manifest.GetFilesByType(FileTypePlugin)
	if len(plugins) != 2 {
		t.Errorf("GetFilesByType(Plugin) returned %d files, want 2", len(plugins))
	}

	meshes := manifest.GetFilesByType(FileTypeMesh)
	if len(meshes) != 1 {
		t.Errorf("GetFilesByType(Mesh) returned %d files, want 1", len(meshes))
	}

	scripts := manifest.GetFilesByType(FileTypeScript)
	if len(scripts) != 0 {
		t.Errorf("GetFilesByType(Script) returned %d files, want 0", len(scripts))
	}
}

func TestManifest_GetFilesByDirectory(t *testing.T) {
	entries := []FileEntry{
		NewFileEntry("data/test.esp", 100),
		NewFileEntry("data/test2.esp", 200),
		NewFileEntry("meshes/test.nif", 300),
		NewFileEntry("readme.txt", 50),
	}
	manifest := NewManifest(entries)

	dataFiles := manifest.GetFilesByDirectory("data")
	if len(dataFiles) != 2 {
		t.Errorf("GetFilesByDirectory(data) returned %d files, want 2", len(dataFiles))
	}

	// Test with different case
	dataFiles2 := manifest.GetFilesByDirectory("Data")
	if len(dataFiles2) != 2 {
		t.Errorf("GetFilesByDirectory(Data) returned %d files, want 2", len(dataFiles2))
	}

	rootFiles := manifest.GetFilesByDirectory("")
	if len(rootFiles) != 1 {
		t.Errorf("GetFilesByDirectory('') returned %d files, want 1", len(rootFiles))
	}
}

func TestManifest_GetFilesByExtension(t *testing.T) {
	entries := []FileEntry{
		NewFileEntry("test1.esp", 100),
		NewFileEntry("test2.esp", 200),
		NewFileEntry("test.esm", 300),
	}
	manifest := NewManifest(entries)

	espFiles := manifest.GetFilesByExtension(".esp")
	if len(espFiles) != 2 {
		t.Errorf("GetFilesByExtension(.esp) returned %d files, want 2", len(espFiles))
	}

	// Test without leading dot
	espFiles2 := manifest.GetFilesByExtension("esp")
	if len(espFiles2) != 2 {
		t.Errorf("GetFilesByExtension(esp) returned %d files, want 2", len(espFiles2))
	}
}

func TestManifest_HasFile(t *testing.T) {
	entries := []FileEntry{
		NewFileEntry("data/test.esp", 100),
		NewFileEntry("meshes/test.nif", 200),
	}
	manifest := NewManifest(entries)

	tests := []struct {
		path     string
		expected bool
	}{
		{"data/test.esp", true},
		{"Data/Test.esp", true}, // Case insensitive
		{"data\\test.esp", true}, // Backslash normalization
		{"meshes/test.nif", true},
		{"data/missing.esp", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := manifest.HasFile(tt.path)
			if result != tt.expected {
				t.Errorf("HasFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestManifest_GetFile(t *testing.T) {
	entries := []FileEntry{
		NewFileEntry("data/test.esp", 100),
	}
	manifest := NewManifest(entries)

	t.Run("existing file", func(t *testing.T) {
		file := manifest.GetFile("data/test.esp")
		if file == nil {
			t.Fatal("GetFile() returned nil for existing file")
		}
		if file.Size != 100 {
			t.Errorf("GetFile() Size = %d, want 100", file.Size)
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		file := manifest.GetFile("Data/Test.ESP")
		if file == nil {
			t.Fatal("GetFile() returned nil for case-variant path")
		}
	})

	t.Run("non-existing file", func(t *testing.T) {
		file := manifest.GetFile("data/missing.esp")
		if file != nil {
			t.Error("GetFile() should return nil for non-existing file")
		}
	})
}

func TestEmptyManifest(t *testing.T) {
	manifest := NewManifest([]FileEntry{})

	if manifest.TotalCount != 0 {
		t.Errorf("TotalCount = %d, want 0", manifest.TotalCount)
	}
	if manifest.TotalSize != 0 {
		t.Errorf("TotalSize = %d, want 0", manifest.TotalSize)
	}
	if len(manifest.Files) != 0 {
		t.Errorf("Files length = %d, want 0", len(manifest.Files))
	}
}
