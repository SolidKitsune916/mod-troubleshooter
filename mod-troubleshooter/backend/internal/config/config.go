package config

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Config holds all configuration for the application.
type Config struct {
	// Port is the HTTP server port (default: 8080)
	Port string

	// NexusAPIKey is the API key for Nexus Mods API
	NexusAPIKey string

	// DataDir is the directory for storing cached data (default: ./data)
	DataDir string

	// CacheTTLHours is how long to cache data in hours (default: 168 = 1 week)
	CacheTTLHours int

	// Environment is the running environment (development, production)
	Environment string

	// CORSOrigins are the allowed origins for CORS
	CORSOrigins []string
}

// Load reads configuration from environment variables and optional .env file.
// The .env file is loaded first, then environment variables override.
func Load() (*Config, error) {
	// Try to load .env file from current directory and parent directories
	loadEnvFile()

	cfg := &Config{
		Port:          getEnv("PORT", "8080"),
		NexusAPIKey:   getEnv("NEXUS_API_KEY", ""),
		DataDir:       getEnv("DATA_DIR", "./data"),
		CacheTTLHours: getEnvInt("CACHE_TTL_HOURS", 168),
		Environment:   getEnv("ENVIRONMENT", "development"),
	}

	// Parse CORS origins
	origins := getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:3000")
	cfg.CORSOrigins = parseCSV(origins)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that required configuration is present.
func (c *Config) Validate() error {
	// NexusAPIKey is only required in production
	if c.Environment == "production" && c.NexusAPIKey == "" {
		return errors.New("NEXUS_API_KEY is required in production")
	}

	return nil
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// loadEnvFile attempts to load a .env file from the current directory
// or parent directories.
func loadEnvFile() {
	// Try current directory first
	paths := []string{".env", "../.env", "../../.env"}

	for _, path := range paths {
		if err := loadEnvFromPath(path); err == nil {
			return
		}
	}
}

// loadEnvFromPath loads environment variables from a file.
func loadEnvFromPath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if present
		value = trimQuotes(value)

		// Only set if not already set in environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// trimQuotes removes surrounding quotes from a string.
func trimQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') ||
			(s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// getEnv returns the environment variable value or the default.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt returns the environment variable as an int or the default.
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var result int
	for _, c := range value {
		if c < '0' || c > '9' {
			return defaultValue
		}
		result = result*10 + int(c-'0')
	}
	return result
}

// parseCSV splits a comma-separated string into a slice.
func parseCSV(s string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
