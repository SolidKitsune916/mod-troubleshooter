package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mod-troubleshooter/backend/internal/nexus"
)

// mockNexusClientGetter implements NexusClientGetter for testing.
type mockNexusClientGetter struct {
	client *nexus.Client
}

func (m *mockNexusClientGetter) Get() *nexus.Client {
	return m.client
}

func TestQuotaHandler_GetQuota_NoClient(t *testing.T) {
	handler := NewQuotaHandler(&mockNexusClientGetter{client: nil})

	req := httptest.NewRequest(http.MethodGet, "/api/quota", nil)
	w := httptest.NewRecorder()

	handler.GetQuota(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Check that the response has a data field
	if resp.Data == nil {
		t.Fatal("expected response to have data field")
	}

	// Parse the data as QuotaResponse
	dataBytes, _ := json.Marshal(resp.Data)
	var quota QuotaResponse
	if err := json.Unmarshal(dataBytes, &quota); err != nil {
		t.Fatalf("failed to parse quota data: %v", err)
	}

	if quota.Available {
		t.Error("expected Available to be false when no client configured")
	}
}

func TestQuotaHandler_GetQuota_MethodNotAllowed(t *testing.T) {
	handler := NewQuotaHandler(&mockNexusClientGetter{client: nil})

	req := httptest.NewRequest(http.MethodPost, "/api/quota", nil)
	w := httptest.NewRecorder()

	handler.GetQuota(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}
