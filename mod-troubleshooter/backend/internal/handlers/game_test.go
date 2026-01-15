package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGameHandler_GetGames(t *testing.T) {
	handler := NewGameHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/games", nil)
	w := httptest.NewRecorder()

	handler.GetGames(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Data == nil {
		t.Fatal("expected response to have data field")
	}

	// Parse the data as []GameDomain
	dataBytes, _ := json.Marshal(resp.Data)
	var games []GameDomain
	if err := json.Unmarshal(dataBytes, &games); err != nil {
		t.Fatalf("failed to parse games data: %v", err)
	}

	// Check that we have the expected number of games
	if len(games) != 3 {
		t.Errorf("expected 3 games, got %d", len(games))
	}

	// Check that games are in the expected order
	expectedOrder := []string{"skyrim", "stardew", "cyberpunk"}
	for i, id := range expectedOrder {
		if i >= len(games) {
			t.Errorf("missing game at index %d", i)
			continue
		}
		if games[i].ID != id {
			t.Errorf("game at index %d: expected ID %q, got %q", i, id, games[i].ID)
		}
	}

	// Verify first game has all required fields
	if games[0].ID != "skyrim" {
		t.Errorf("expected first game ID to be 'skyrim', got %q", games[0].ID)
	}
	if games[0].Label != "Skyrim Special Edition" {
		t.Errorf("expected first game Label to be 'Skyrim Special Edition', got %q", games[0].Label)
	}
	if games[0].DomainName != "skyrimspecialedition" {
		t.Errorf("expected first game DomainName to be 'skyrimspecialedition', got %q", games[0].DomainName)
	}
}

func TestGameHandler_GetGames_MethodNotAllowed(t *testing.T) {
	handler := NewGameHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/games", nil)
	w := httptest.NewRecorder()

	handler.GetGames(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestGetNexusDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"skyrim", "skyrimspecialedition"},
		{"stardew", "stardewvalley"},
		{"cyberpunk", "cyberpunk2077"},
		{"unknown", "unknown"},                      // Falls back to input
		{"skyrimspecialedition", "skyrimspecialedition"}, // Already a domain name
	}

	for _, tt := range tests {
		result := GetNexusDomain(tt.input)
		if result != tt.expected {
			t.Errorf("GetNexusDomain(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestIsValidGameID(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"skyrim", true},
		{"stardew", true},
		{"cyberpunk", true},
		{"unknown", false},
		{"skyrimspecialedition", false}, // Domain name, not ID
		{"", false},
	}

	for _, tt := range tests {
		result := IsValidGameID(tt.input)
		if result != tt.expected {
			t.Errorf("IsValidGameID(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
