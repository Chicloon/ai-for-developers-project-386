package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"call-booking/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://postgres:postgres@localhost:5432/call_booking_test?sslmode=disable")
	if err != nil {
		t.Skipf("Database not available: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Database not available: %v", err)
	}

	_, _ = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS availability_rules (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			type VARCHAR(20) NOT NULL CHECK (type IN ('recurring', 'one-time')),
			day_of_week INT,
			date DATE,
			time_ranges JSONB NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS blocked_days (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			date DATE NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS bookings (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			slot_date DATE NOT NULL,
			slot_start_time TIME NOT NULL,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled')),
			recurrence VARCHAR(20) NOT NULL DEFAULT 'none' CHECK (recurrence IN ('none', 'daily', 'weekly', 'yearly')),
			day_of_week INT,
			end_date DATE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`)
	_, _ = pool.Exec(ctx, "TRUNCATE availability_rules, blocked_days, bookings CASCADE")
	return pool
}

func TestCreateAvailabilityRule(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := NewRouter(pool)

	body := models.CreateAvailabilityRule{
		Type:      "recurring",
		DayOfWeek: ptrInt32(1),
		TimeRanges: []models.TimeRange{
			{StartTime: "10:00", EndTime: "12:00"},
		},
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/availability-rules", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var rule models.AvailabilityRule
	if err := json.Unmarshal(w.Body.Bytes(), &rule); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if rule.Type != "recurring" {
		t.Fatalf("expected type recurring, got %s", rule.Type)
	}
	if *rule.DayOfWeek != 1 {
		t.Fatalf("expected dayOfWeek 1, got %v", rule.DayOfWeek)
	}
}

func TestListAvailabilityRules(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := NewRouter(pool)

	// Create a rule first
	body := models.CreateAvailabilityRule{
		Type:       "one-time",
		Date:       strPtr("2026-04-10"),
		TimeRanges: []models.TimeRange{{StartTime: "14:00", EndTime: "16:00"}},
	}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/availability-rules", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// List rules
	req = httptest.NewRequest(http.MethodGet, "/api/availability-rules", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var rules []models.AvailabilityRule
	if err := json.Unmarshal(w.Body.Bytes(), &rules); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
}

func TestDeleteAvailabilityRule(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	router := NewRouter(pool)

	// Create a rule first
	body := models.CreateAvailabilityRule{
		Type:       "recurring",
		DayOfWeek:  ptrInt32(3),
		TimeRanges: []models.TimeRange{{StartTime: "09:00", EndTime: "11:00"}},
	}
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/availability-rules", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var created models.AvailabilityRule
	json.Unmarshal(w.Body.Bytes(), &created)

	// Delete the rule
	req = httptest.NewRequest(http.MethodDelete, "/api/availability-rules/"+created.ID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}

	// Verify it's deleted
	req = httptest.NewRequest(http.MethodGet, "/api/availability-rules", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var rules []models.AvailabilityRule
	json.Unmarshal(w.Body.Bytes(), &rules)
	if len(rules) != 0 {
		t.Fatalf("expected 0 rules after delete, got %d", len(rules))
	}
}

func ptrInt32(v int32) *int32 { return &v }
func strPtr(v string) *string  { return &v }
