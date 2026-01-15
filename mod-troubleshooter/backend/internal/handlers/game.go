package handlers

// GameDomain maps frontend game IDs to Nexus Mods domain names.
type GameDomain struct {
	ID         string // Frontend ID (skyrim, stardew, cyberpunk)
	DomainName string // Nexus domain (skyrimspecialedition, stardewvalley, cyberpunk2077)
	Label      string // Display name
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
