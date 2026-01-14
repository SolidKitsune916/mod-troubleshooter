package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				DBPath: filepath.Join(tempDir, "test.db"),
				TTL:    time.Hour,
			},
			wantErr: false,
		},
		{
			name: "default TTL",
			cfg: Config{
				DBPath: filepath.Join(tempDir, "test2.db"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := New(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if cache != nil {
				cache.Close()
			}
		})
	}
}

func TestCacheKey(t *testing.T) {
	key := CacheKey("skyrimspecialedition", 12345, 67890)
	expected := "fomod:skyrimspecialedition:12345:67890"
	if key != expected {
		t.Errorf("CacheKey() = %q, want %q", key, expected)
	}
}

func TestCache_SetGet(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := New(Config{
		DBPath: filepath.Join(tempDir, "test.db"),
		TTL:    time.Hour,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	type testData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	// Test Set and Get
	t.Run("set and get", func(t *testing.T) {
		data := testData{Name: "test", Value: 42}
		err := cache.Set(ctx, "key1", data)
		if err != nil {
			t.Errorf("Set() error = %v", err)
		}

		var result testData
		err = cache.Get(ctx, "key1", &result)
		if err != nil {
			t.Errorf("Get() error = %v", err)
		}
		if result.Name != data.Name || result.Value != data.Value {
			t.Errorf("Get() = %+v, want %+v", result, data)
		}
	})

	// Test Get non-existent key
	t.Run("get non-existent", func(t *testing.T) {
		var result testData
		err := cache.Get(ctx, "nonexistent", &result)
		if err != ErrNotFound {
			t.Errorf("Get() error = %v, want %v", err, ErrNotFound)
		}
	})

	// Test Update existing key
	t.Run("update existing", func(t *testing.T) {
		data := testData{Name: "updated", Value: 100}
		err := cache.Set(ctx, "key1", data)
		if err != nil {
			t.Errorf("Set() error = %v", err)
		}

		var result testData
		err = cache.Get(ctx, "key1", &result)
		if err != nil {
			t.Errorf("Get() error = %v", err)
		}
		if result.Name != data.Name || result.Value != data.Value {
			t.Errorf("Get() = %+v, want %+v", result, data)
		}
	})
}

func TestCache_Expiration(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := New(Config{
		DBPath: filepath.Join(tempDir, "test.db"),
		TTL:    50 * time.Millisecond, // Very short TTL for testing
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	data := map[string]string{"key": "value"}

	// Set data
	err = cache.Set(ctx, "expiring", data)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Should be retrievable immediately
	var result map[string]string
	err = cache.Get(ctx, "expiring", &result)
	if err != nil {
		t.Errorf("Get() immediate error = %v", err)
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired now
	err = cache.Get(ctx, "expiring", &result)
	if err != ErrExpired {
		t.Errorf("Get() after expiration error = %v, want %v", err, ErrExpired)
	}
}

func TestCache_SetWithTTL(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := New(Config{
		DBPath: filepath.Join(tempDir, "test.db"),
		TTL:    time.Hour, // Long default TTL
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	data := map[string]string{"key": "value"}

	// Set with short custom TTL
	err = cache.SetWithTTL(ctx, "custom_ttl", data, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("SetWithTTL() error = %v", err)
	}

	// Should be retrievable immediately
	var result map[string]string
	err = cache.Get(ctx, "custom_ttl", &result)
	if err != nil {
		t.Errorf("Get() immediate error = %v", err)
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired now
	err = cache.Get(ctx, "custom_ttl", &result)
	if err != ErrExpired {
		t.Errorf("Get() after expiration error = %v, want %v", err, ErrExpired)
	}
}

func TestCache_Delete(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := New(Config{
		DBPath: filepath.Join(tempDir, "test.db"),
		TTL:    time.Hour,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	data := map[string]string{"key": "value"}

	// Set data
	err = cache.Set(ctx, "to_delete", data)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Delete
	err = cache.Delete(ctx, "to_delete")
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Should not be found
	var result map[string]string
	err = cache.Get(ctx, "to_delete", &result)
	if err != ErrNotFound {
		t.Errorf("Get() after delete error = %v, want %v", err, ErrNotFound)
	}
}

func TestCache_Cleanup(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := New(Config{
		DBPath: filepath.Join(tempDir, "test.db"),
		TTL:    50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	data := map[string]string{"key": "value"}

	// Set multiple entries
	err = cache.Set(ctx, "entry1", data)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	err = cache.Set(ctx, "entry2", data)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Cleanup
	err = cache.Cleanup(ctx)
	if err != nil {
		t.Errorf("Cleanup() error = %v", err)
	}

	// Entries should be gone (not just expired)
	var result map[string]string
	err = cache.Get(ctx, "entry1", &result)
	if err != ErrNotFound {
		t.Errorf("Get() after cleanup error = %v, want %v", err, ErrNotFound)
	}
}

func TestCache_CreateDirectory(t *testing.T) {
	tempDir := t.TempDir()
	nestedPath := filepath.Join(tempDir, "nested", "deep", "cache.db")

	cache, err := New(Config{
		DBPath: nestedPath,
		TTL:    time.Hour,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer cache.Close()

	// Verify directory was created
	dir := filepath.Dir(nestedPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Directory %s was not created", dir)
	}
}
