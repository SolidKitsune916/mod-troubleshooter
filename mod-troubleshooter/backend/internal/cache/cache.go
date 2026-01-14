package cache

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Common errors returned by the cache.
var (
	ErrNotFound = errors.New("cache entry not found")
	ErrExpired  = errors.New("cache entry expired")
)

// Config holds configuration for the cache.
type Config struct {
	// DBPath is the path to the SQLite database file.
	DBPath string

	// TTL is the default time-to-live for cache entries.
	TTL time.Duration
}

// Cache provides SQLite-backed caching for FOMOD analysis results.
type Cache struct {
	db  *sql.DB
	ttl time.Duration
}

// New creates a new cache with the given configuration.
func New(cfg Config) (*Cache, error) {
	// Ensure the directory exists
	dir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create cache directory: %w", err)
	}

	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Initialize schema
	if err := initSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize schema: %w", err)
	}

	ttl := cfg.TTL
	if ttl == 0 {
		ttl = 7 * 24 * time.Hour // Default 1 week
	}

	return &Cache{
		db:  db,
		ttl: ttl,
	}, nil
}

// initSchema creates the necessary tables.
func initSchema(db *sql.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS fomod_cache (
			cache_key TEXT PRIMARY KEY,
			data TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			expires_at INTEGER NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_fomod_cache_expires ON fomod_cache(expires_at);
	`
	_, err := db.Exec(schema)
	return err
}

// CacheKey generates a cache key from game domain, mod ID, and file ID.
func CacheKey(game string, modID, fileID int) string {
	return fmt.Sprintf("fomod:%s:%d:%d", game, modID, fileID)
}

// Get retrieves a cached entry.
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	var data string
	var expiresAt int64

	err := c.db.QueryRowContext(ctx, `
		SELECT data, expires_at FROM fomod_cache WHERE cache_key = ?
	`, key).Scan(&data, &expiresAt)

	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("query cache: %w", err)
	}

	// Check expiration (using milliseconds for precision)
	if time.Now().UnixMilli() > expiresAt {
		// Clean up expired entry
		c.db.ExecContext(ctx, "DELETE FROM fomod_cache WHERE cache_key = ?", key)
		return ErrExpired
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("unmarshal cache data: %w", err)
	}

	return nil
}

// Set stores an entry in the cache.
func (c *Cache) Set(ctx context.Context, key string, value interface{}) error {
	return c.SetWithTTL(ctx, key, value, c.ttl)
}

// SetWithTTL stores an entry in the cache with a custom TTL.
func (c *Cache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal cache data: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(ttl)

	// Use milliseconds for precision
	_, err = c.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO fomod_cache (cache_key, data, created_at, expires_at)
		VALUES (?, ?, ?, ?)
	`, key, string(data), now.UnixMilli(), expiresAt.UnixMilli())

	if err != nil {
		return fmt.Errorf("insert cache entry: %w", err)
	}

	return nil
}

// Delete removes an entry from the cache.
func (c *Cache) Delete(ctx context.Context, key string) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM fomod_cache WHERE cache_key = ?", key)
	return err
}

// Cleanup removes expired entries from the cache.
func (c *Cache) Cleanup(ctx context.Context) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM fomod_cache WHERE expires_at < ?", time.Now().UnixMilli())
	return err
}

// Close closes the database connection.
func (c *Cache) Close() error {
	return c.db.Close()
}
