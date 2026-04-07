package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"call-booking/internal/models"
)

func TestCreateBooking(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := NewRouter(pool)

	body := models.CreateBooking{
		SlotDate:      "2026-04-06",
		SlotStartTime: "10:00",
		Name:          "John Doe",
		Email:         "john@example.com",
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/bookings", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var booking models.Booking
	if err := json.Unmarshal(w.Body.Bytes(), &booking); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if booking.Name != "John Doe" {
		t.Fatalf("expected name John Doe, got %s", booking.Name)
	}
	if booking.Status != "active" {
		t.Fatalf("expected status active, got %s", booking.Status)
	}
}

func TestCreateRecurringBooking(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := NewRouter(pool)

	endDate := "2026-05-01"
	body := models.CreateBooking{
		SlotDate:      "2026-04-06",
		SlotStartTime: "10:00",
		Name:          "Jane Doe",
		Email:         "jane@example.com",
		Recurrence:    strPtr("weekly"),
		DayOfWeek:     ptrInt32(1),
		EndDate:       &endDate,
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/bookings", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var booking models.Booking
	json.Unmarshal(w.Body.Bytes(), &booking)
	if booking.Recurrence == nil || *booking.Recurrence != "weekly" {
		t.Fatalf("expected recurrence weekly, got %v", booking.Recurrence)
	}
}

func TestListBookings(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := NewRouter(pool)

	// Create a booking
	body := models.CreateBooking{
		SlotDate:      "2026-04-06",
		SlotStartTime: "10:00",
		Name:          "Test User",
		Email:         "test@example.com",
	}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/bookings", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// List bookings
	req = httptest.NewRequest(http.MethodGet, "/api/bookings", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var bookings []models.Booking
	json.Unmarshal(w.Body.Bytes(), &bookings)
	if len(bookings) != 1 {
		t.Fatalf("expected 1 booking, got %d", len(bookings))
	}
}

func TestDeleteBooking(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := NewRouter(pool)

	// Create a booking
	body := models.CreateBooking{
		SlotDate:      "2026-04-06",
		SlotStartTime: "10:00",
		Name:          "Test User",
		Email:         "test@example.com",
	}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/bookings", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var created models.Booking
	json.Unmarshal(w.Body.Bytes(), &created)

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/api/bookings/"+created.ID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}
