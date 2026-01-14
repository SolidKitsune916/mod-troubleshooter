package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/mod-troubleshooter/backend/internal/nexus"
)

// SettingsStore manages runtime settings with thread-safe access.
type SettingsStore struct {
	mu        sync.RWMutex
	nexusKey  string
	onKeyChange func(string) // Callback when API key changes
}

// NewSettingsStore creates a new settings store with initial API key.
func NewSettingsStore(initialKey string) *SettingsStore {
	return &SettingsStore{
		nexusKey: initialKey,
	}
}

// SetOnKeyChange sets the callback for when the API key changes.
func (s *SettingsStore) SetOnKeyChange(fn func(string)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onKeyChange = fn
}

// GetNexusAPIKey returns the current Nexus API key.
func (s *SettingsStore) GetNexusAPIKey() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nexusKey
}

// SetNexusAPIKey updates the Nexus API key.
func (s *SettingsStore) SetNexusAPIKey(key string) {
	s.mu.Lock()
	s.nexusKey = key
	callback := s.onKeyChange
	s.mu.Unlock()

	if callback != nil {
		callback(key)
	}
}

// Settings represents the user-configurable settings.
type Settings struct {
	NexusAPIKey   string `json:"nexusApiKey"`
	HasNexusKey   bool   `json:"hasNexusKey"`
	KeyConfigured bool   `json:"keyConfigured"`
}

// UpdateSettingsRequest is the request body for updating settings.
type UpdateSettingsRequest struct {
	NexusAPIKey string `json:"nexusApiKey"`
}

// SettingsHandler handles settings-related HTTP requests.
type SettingsHandler struct {
	store *SettingsStore
}

// NewSettingsHandler creates a new settings handler.
func NewSettingsHandler(store *SettingsStore) *SettingsHandler {
	return &SettingsHandler{store: store}
}

// GetSettings handles GET /api/settings
// Returns current settings (API key is masked for security).
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	key := h.store.GetNexusAPIKey()

	settings := Settings{
		NexusAPIKey:   maskAPIKey(key),
		HasNexusKey:   key != "",
		KeyConfigured: key != "",
	}

	WriteJSON(w, http.StatusOK, settings)
}

// UpdateSettings handles POST /api/settings
// Updates the Nexus API key.
func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Trim whitespace from API key
	apiKey := strings.TrimSpace(req.NexusAPIKey)

	// Validate API key format (basic validation)
	if apiKey != "" && len(apiKey) < 10 {
		WriteError(w, http.StatusBadRequest, "API key appears to be invalid (too short)")
		return
	}

	h.store.SetNexusAPIKey(apiKey)

	WriteSuccess(w, "Settings updated successfully")
}

// ValidateAPIKey handles POST /api/settings/validate
// Validates the API key by making a test request to Nexus API.
func (h *SettingsHandler) ValidateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	apiKey := strings.TrimSpace(req.NexusAPIKey)
	if apiKey == "" {
		WriteError(w, http.StatusBadRequest, "API key is required")
		return
	}

	// Create a temporary client to test the API key
	client, err := nexus.NewClient(nexus.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create API client")
		return
	}

	// Test the API key by fetching user info
	valid, err := client.ValidateAPIKey(r.Context())
	if err != nil {
		WriteError(w, http.StatusBadGateway, "Failed to validate API key: "+err.Error())
		return
	}

	if !valid {
		WriteError(w, http.StatusUnauthorized, "API key is invalid")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]bool{"valid": true})
}

// maskAPIKey masks all but the last 4 characters of an API key.
func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}

	if len(key) <= 4 {
		return strings.Repeat("*", len(key))
	}

	return strings.Repeat("*", len(key)-4) + key[len(key)-4:]
}
