package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mod-troubleshooter/backend/internal/nexus"
)

// QuotaResponse represents the API quota information returned to the frontend.
type QuotaResponse struct {
	HourlyLimit     int  `json:"hourlyLimit"`
	HourlyRemaining int  `json:"hourlyRemaining"`
	DailyLimit      int  `json:"dailyLimit"`
	DailyRemaining  int  `json:"dailyRemaining"`
	Available       bool `json:"available"`
}

// QuotaHandler handles quota-related endpoints.
type QuotaHandler struct {
	getClient NexusClientGetter
}

// NewQuotaHandler creates a new QuotaHandler.
func NewQuotaHandler(getClient NexusClientGetter) *QuotaHandler {
	return &QuotaHandler{
		getClient: getClient,
	}
}

// GetQuota returns the current Nexus API rate limit information.
// GET /api/quota
func (h *QuotaHandler) GetQuota(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	client := h.getClient.Get()
	if client == nil {
		// No client configured - return unavailable quota info
		resp := QuotaResponse{
			Available: false,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{Data: resp})
		return
	}

	info := client.GetRateLimitInfo()
	if info == nil {
		// No rate limit info available yet (no requests made)
		resp := QuotaResponse{
			Available: false,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{Data: resp})
		return
	}

	resp := QuotaResponse{
		HourlyLimit:     info.HourlyLimit,
		HourlyRemaining: info.HourlyRemaining,
		DailyLimit:      info.DailyLimit,
		DailyRemaining:  info.DailyRemaining,
		Available:       true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Data: resp})
}

// NexusClientInterface defines the interface for the Nexus client used by quota handler.
type NexusClientInterface interface {
	GetRateLimitInfo() *nexus.RateLimitInfo
}
