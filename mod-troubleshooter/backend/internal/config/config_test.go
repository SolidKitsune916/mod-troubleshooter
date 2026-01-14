package config

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	// Test default value when env var not set
	result := getEnv("TEST_NONEXISTENT_VAR_12345", "default")
	if result != "default" {
		t.Errorf("getEnv() = %q, want %q", result, "default")
	}

	// Test with env var set
	os.Setenv("TEST_VAR_12345", "custom_value")
	defer os.Unsetenv("TEST_VAR_12345")

	result = getEnv("TEST_VAR_12345", "default")
	if result != "custom_value" {
		t.Errorf("getEnv() = %q, want %q", result, "custom_value")
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue int
		want         int
	}{
		{"empty uses default", "", 42, 42},
		{"valid int", "123", 0, 123},
		{"invalid uses default", "abc", 42, 42},
		{"mixed uses default", "12abc", 42, 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("TEST_INT_VAR", tt.envValue)
				defer os.Unsetenv("TEST_INT_VAR")
			} else {
				os.Unsetenv("TEST_INT_VAR")
			}

			result := getEnvInt("TEST_INT_VAR", tt.defaultValue)
			if result != tt.want {
				t.Errorf("getEnvInt() = %d, want %d", result, tt.want)
			}
		})
	}
}

func TestParseCSV(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty string", "", nil},
		{"single value", "http://localhost:5173", []string{"http://localhost:5173"}},
		{"multiple values", "http://localhost:5173,http://localhost:3000", []string{"http://localhost:5173", "http://localhost:3000"}},
		{"with spaces", " http://localhost:5173 , http://localhost:3000 ", []string{"http://localhost:5173", "http://localhost:3000"}},
		{"empty parts", "a,,b", []string{"a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCSV(tt.input)
			if len(result) != len(tt.want) {
				t.Errorf("parseCSV() len = %d, want %d", len(result), len(tt.want))
				return
			}
			for i, v := range result {
				if v != tt.want[i] {
					t.Errorf("parseCSV()[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestTrimQuotes(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"hello"`, "hello"},
		{`'hello'`, "hello"},
		{`hello`, "hello"},
		{`"hello`, `"hello`},
		{`hello"`, `hello"`},
		{`""`, ""},
		{`''`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := trimQuotes(tt.input)
			if result != tt.want {
				t.Errorf("trimQuotes(%q) = %q, want %q", tt.input, result, tt.want)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	// Clean environment
	os.Unsetenv("NEXUS_API_KEY")
	os.Unsetenv("PORT")
	os.Unsetenv("DATA_DIR")
	os.Unsetenv("CACHE_TTL_HOURS")
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("CORS_ORIGINS")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Test defaults
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.DataDir != "./data" {
		t.Errorf("DataDir = %q, want %q", cfg.DataDir, "./data")
	}
	if cfg.CacheTTLHours != 168 {
		t.Errorf("CacheTTLHours = %d, want %d", cfg.CacheTTLHours, 168)
	}
	if cfg.Environment != "development" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "development")
	}
	if len(cfg.CORSOrigins) != 2 {
		t.Errorf("CORSOrigins len = %d, want 2", len(cfg.CORSOrigins))
	}
}

func TestValidate(t *testing.T) {
	// Production without API key should fail
	cfg := &Config{
		Environment: "production",
		NexusAPIKey: "",
	}
	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should fail for production without API key")
	}

	// Production with API key should pass
	cfg.NexusAPIKey = "test-key"
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}

	// Development without API key should pass
	cfg.Environment = "development"
	cfg.NexusAPIKey = ""
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestIsDevelopment(t *testing.T) {
	cfg := &Config{Environment: "development"}
	if !cfg.IsDevelopment() {
		t.Error("IsDevelopment() = false, want true")
	}

	cfg.Environment = "production"
	if cfg.IsDevelopment() {
		t.Error("IsDevelopment() = true, want false")
	}
}
