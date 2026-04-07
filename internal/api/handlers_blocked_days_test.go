package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"call-booking/internal/models"
)

func TestCreateBlockedDay(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := NewRouter(pool)

	body := models.CreateBlockedDay{Date: "2026-04-10"}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/blocked-days", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var day models.BlockedDay
	if err := json.Unmarshal(w.Body.Bytes(), &day); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if day.Date != "2026-04-10" {
		t.Fatalf("expected date 2026-04-10, got %s", day.Date)
	}
}

func TestDeleteBlockedDay(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := NewRouter(pool)

	// Create first
	body := models.CreateBlockedDay{Date: "2026-05-01"}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/blocked-days", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var created models.BlockedDay
	json.Unmarshal(w.Body.Bytes(), &created)

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/api/blocked-days/"+created.ID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestListBlockedDays(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := NewRouter(pool)

	// Get empty list
	req := httptest.NewRequest(http.MethodGet, "/api/blocked-days", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var days []models.BlockedDay
	if err := json.Unmarshal(w.Body.Bytes(), &days); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if days == nil {
		t.Fatal("expected empty array, got null")
	}
	if len(days) != 0 {
		t.Fatalf("expected 0 days, got %d", len(days))
	}
}
