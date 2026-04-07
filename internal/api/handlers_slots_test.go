package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"call-booking/internal/models"
)

func TestGetSlots(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := NewRouter(pool)

	// Create a rule for Monday (day 1)
	ruleBody := models.CreateAvailabilityRule{
		Type:      "recurring",
		DayOfWeek: ptrInt32(1),
		TimeRanges: []models.TimeRange{
			{StartTime: "10:00", EndTime: "11:00"},
		},
	}
	data, _ := json.Marshal(ruleBody)
	req := httptest.NewRequest(http.MethodPost, "/api/availability-rules", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Get slots for a Monday (2026-04-06 is Monday)
	req = httptest.NewRequest(http.MethodGet, "/api/slots?date=2026-04-06", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var slots []models.Slot
	if err := json.Unmarshal(w.Body.Bytes(), &slots); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(slots) != 2 {
		t.Fatalf("expected 2 slots, got %d", len(slots))
	}
}

func TestGetSlotsOnBlockedDay(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := NewRouter(pool)

	// Create a rule
	ruleBody := models.CreateAvailabilityRule{
		Type:      "recurring",
		DayOfWeek: ptrInt32(1),
		TimeRanges: []models.TimeRange{
			{StartTime: "10:00", EndTime: "11:00"},
		},
	}
	data, _ := json.Marshal(ruleBody)
	req := httptest.NewRequest(http.MethodPost, "/api/availability-rules", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Block the day
	blockBody := models.CreateBlockedDay{Date: "2026-04-06"}
	data, _ = json.Marshal(blockBody)
	req = httptest.NewRequest(http.MethodPost, "/api/blocked-days", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Get slots — should be empty
	req = httptest.NewRequest(http.MethodGet, "/api/slots?date=2026-04-06", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var slots []models.Slot
	json.Unmarshal(w.Body.Bytes(), &slots)
	if len(slots) != 0 {
		t.Fatalf("expected 0 slots on blocked day, got %d", len(slots))
	}
}
