package handlers

import (
	"encoding/json"
	"net/http"
)

// GameDomain maps frontend game IDs to Nexus Mods domain names.
type GameDomain struct {
	ID         string `json:"id"`         // Frontend ID (skyrim, stardew, cyberpunk)
	DomainName string `json:"domainName"` // Nexus domain (skyrimspecialedition, stardewvalley, cyberpunk2077)
	Label      string `json:"label"`      // Display name
}

// GameDomains is a map of game IDs to their Nexus domain info.
var GameDomains = map[string]GameDomain{
	"skyrim": {
		ID:         "skyrim",
		DomainName: "skyrimspecialedition",
		Label:      "Skyrim Special Edition",
	},
	"stardew": {
		ID:         "stardew",
		DomainName: "stardewvalley",
		Label:      "Stardew Valley",
	},
	"cyberpunk": {
		ID:         "cyberpunk",
		DomainName: "cyberpunk2077",
		Label:      "Cyberpunk 2077",
	},
}

// orderedGameIDs defines the display order of games.
var orderedGameIDs = []string{"skyrim", "stardew", "cyberpunk"}

// GetNexusDomain returns the Nexus domain name for a given game ID.
// Falls back to the input if not found (for backwards compatibility).
func GetNexusDomain(gameID string) string {
	if domain, ok := GameDomains[gameID]; ok {
		return domain.DomainName
	}
	// Return input as-is (might already be a domain name)
	return gameID
}

// IsValidGameID returns true if the given game ID is supported.
func IsValidGameID(gameID string) bool {
	_, ok := GameDomains[gameID]
	return ok
}

// GameHandler handles game-related endpoints.
type GameHandler struct{}

// NewGameHandler creates a new GameHandler.
func NewGameHandler() *GameHandler {
	return &GameHandler{}
}

// GetGames returns the list of supported games.
// GET /api/games
func (h *GameHandler) GetGames(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Build ordered list of games
	games := make([]GameDomain, 0, len(orderedGameIDs))
	for _, id := range orderedGameIDs {
		if game, ok := GameDomains[id]; ok {
			games = append(games, game)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Data: games})
}
