# Call Booking App Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a call booking service where owners publish availability rules and clients book 30-minute slots.

**Architecture:** Monorepo with TypeSpec API contract, Go backend (chi + pgx), Next.js frontend, PostgreSQL database.

**Tech Stack:** TypeSpec, Go, chi, pgx, Next.js, Mantine, PostgreSQL

---

## File Structure

```
typespec/
  tspconfig.yaml              # TypeSpec config
  package.json                # TypeSpec dependencies
  main.tsp                    # Entry point
  models.tsp                  # All model definitions
  availability-rules.tsp      # Rules endpoints
  blocked-days.tsp            # Blocked days endpoints
  slots.tsp                   # Slots endpoint
  bookings.tsp                # Bookings endpoints

cmd/server/
  main.go                     # Go entry point

internal/
  api/
    router.go                 # Chi router setup
    handlers_rules.go         # Availability rule handlers
    handlers_blocked_days.go  # Blocked day handlers
    handlers_slots.go         # Slot handler
    handlers_bookings.go      # Booking handlers
  models/
    models.go                 # Go structs matching DB
  db/
    db.go                     # DB connection
    migrate.go                # Migration runner
  slots/
    generator.go              # Slot generation logic

migrations/
  001_initial.up.sql          # Initial schema
  001_initial.down.sql        # Rollback

web/
  app/
    layout.tsx
    page.tsx
    globals.css
  components/
    ...
```

---

### Task 1: Initialize TypeSpec API Contract

**Files:**
- Create: `typespec/tspconfig.yaml`
- Create: `typespec/package.json`
- Create: `typespec/main.tsp`
- Create: `typespec/models.tsp`
- Create: `typespec/availability-rules.tsp`
- Create: `typespec/blocked-days.tsp`
- Create: `typespec/slots.tsp`
- Create: `typespec/bookings.tsp`

- [ ] **Step 1: Create TypeSpec project config**

Create `typespec/package.json`:
```json
{
  "name": "call-booking-api",
  "version": "0.1.0",
  "type": "module",
  "dependencies": {
    "@typespec/compiler": "latest",
    "@typespec/http": "latest",
    "@typespec/rest": "latest",
    "@typespec/openapi": "latest",
    "@typespec/openapi3": "latest"
  }
}
```

Create `typespec/tspconfig.yaml`:
```yaml
emit:
  - "@typespec/openapi3"
options:
  "@typespec/openapi3":
    emitter-output-dir: "{output-dir}/openapi3"
    new-line: lf
```

- [ ] **Step 2: Install TypeSpec dependencies**

Run:
```bash
cd typespec && npm install
```

- [ ] **Step 3: Create main.tsp entry point**

Create `typespec/main.tsp`:
```typespec
import "@typespec/http";
import "@typespec/rest";
import "@typespec/openapi";
import "@typespec/openapi3";

using Http;
using Rest;

@service(#{
  title: "Call Booking API",
  version: "0.1.0"
})
@server("http://localhost:8080", "Local development server")
namespace CallBooking;

include "./models.tsp";
include "./availability-rules.tsp";
include "./blocked-days.tsp";
include "./slots.tsp";
include "./bookings.tsp";
```

- [ ] **Step 4: Create models.tsp**

Create `typespec/models.tsp`:
```typespec
model TimeRange {
  startTime: string;
  endTime: string;
}

model AvailabilityRule {
  id: string;
  type: "recurring" | "one-time";
  dayOfWeek?: int32;
  date?: string;
  timeRanges: TimeRange[];
}

model CreateAvailabilityRule {
  type: "recurring" | "one-time";
  dayOfWeek?: int32;
  date?: string;
  timeRanges: TimeRange[];
}

model BlockedDay {
  id: string;
  date: string;
}

model CreateBlockedDay {
  date: string;
}

model Slot {
  id: string;
  date: string;
  startTime: string;
  endTime: string;
  isBooked: boolean;
}

model Booking {
  id: string;
  slotDate: string;
  slotStartTime: string;
  name: string;
  email: string;
  status: "active" | "cancelled";
  recurrence?: "none" | "daily" | "weekly" | "yearly";
  dayOfWeek?: int32;
  endDate?: string;
}

model CreateBooking {
  slotDate: string;
  slotStartTime: string;
  name: string;
  email: string;
  recurrence?: "none" | "daily" | "weekly" | "yearly";
  dayOfWeek?: int32;
  endDate?: string;
}

model ErrorResponse {
  error: string;
}
```

- [ ] **Step 5: Create availability-rules.tsp**

Create `typespec/availability-rules.tsp`:
```typespec
@route("/api/availability-rules")
interface AvailabilityRules {
  @get list(): AvailabilityRule[] | ErrorResponse;
  @post create(@body rule: CreateAvailabilityRule): AvailabilityRule | ErrorResponse;
  @put @route("/{id}") update(@path id: string, @body rule: CreateAvailabilityRule): AvailabilityRule | ErrorResponse;
  @delete @route("/{id}") delete(@path id: string): void | ErrorResponse;
}
```

- [ ] **Step 6: Create blocked-days.tsp**

Create `typespec/blocked-days.tsp`:
```typespec
@route("/api/blocked-days")
interface BlockedDays {
  @post create(@body day: CreateBlockedDay): BlockedDay | ErrorResponse;
  @delete @route("/{id}") delete(@path id: string): void | ErrorResponse;
}
```

- [ ] **Step 7: Create slots.tsp**

Create `typespec/slots.tsp`:
```typespec
@route("/api/slots")
interface Slots {
  @get list(@query date: string): Slot[] | ErrorResponse;
}
```

- [ ] **Step 8: Create bookings.tsp**

Create `typespec/bookings.tsp`:
```typespec
@route("/api/bookings")
interface Bookings {
  @post create(@body booking: CreateBooking): Booking | ErrorResponse;
  @get list(): Booking[] | ErrorResponse;
  @delete @route("/{id}") delete(@path id: string): void | ErrorResponse;
}
```

- [ ] **Step 9: Compile and verify**

Run:
```bash
cd typespec && npx tsp compile .
```

Expected: `tsp-output/openapi3/openapi.yaml` generated successfully.

---

### Task 2: Initialize Go Project and Database

**Files:**
- Create: `go.mod`
- Create: `go.sum`
- Create: `cmd/server/main.go`
- Create: `internal/db/db.go`
- Create: `internal/db/migrate.go`
- Create: `internal/models/models.go`
- Create: `migrations/001_initial.up.sql`
- Create: `migrations/001_initial.down.sql`

- [ ] **Step 1: Initialize Go module**

Run:
```bash
go mod init call-booking
go get github.com/go-chi/chi/v5
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool
```

- [ ] **Step 2: Create database migration up**

Create `migrations/001_initial.up.sql`:
```sql
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
```

- [ ] **Step 3: Create database migration down**

Create `migrations/001_initial.down.sql`:
```sql
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS blocked_days;
DROP TABLE IF EXISTS availability_rules;
```

- [ ] **Step 4: Create DB connection**

Create `internal/db/db.go`:
```go
package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/call_booking?sslmode=disable"
	}
	return pgxpool.New(ctx, dsn)
}
```

- [ ] **Step 5: Create migration runner**

Create `internal/db/migrate.go`:
```go
package db

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed ../../migrations/*.sql
var migrationsFS embed.FS

func Migrate(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var upFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	for _, f := range upFiles {
		content, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("exec %s: %w", f, err)
		}
		fmt.Printf("Applied migration: %s\n", f)
	}
	return nil
}
```

- [ ] **Step 6: Create Go models**

Create `internal/models/models.go`:
```go
package models

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type TimeRange struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type AvailabilityRule struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	DayOfWeek  *int32      `json:"dayOfWeek,omitempty"`
	Date       *string     `json:"date,omitempty"`
	TimeRanges []TimeRange `json:"timeRanges"`
}

type CreateAvailabilityRule struct {
	Type       string      `json:"type"`
	DayOfWeek  *int32      `json:"dayOfWeek,omitempty"`
	Date       *string     `json:"date,omitempty"`
	TimeRanges []TimeRange `json:"timeRanges"`
}

type BlockedDay struct {
	ID   string `json:"id"`
	Date string `json:"date"`
}

type CreateBlockedDay struct {
	Date string `json:"date"`
}

type Slot struct {
	ID         string `json:"id"`
	Date       string `json:"date"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
	IsBooked   bool   `json:"isBooked"`
}

type Booking struct {
	ID            string  `json:"id"`
	SlotDate      string  `json:"slotDate"`
	SlotStartTime string  `json:"slotStartTime"`
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	Status        string  `json:"status"`
	Recurrence    *string `json:"recurrence,omitempty"`
	DayOfWeek     *int32  `json:"dayOfWeek,omitempty"`
	EndDate       *string `json:"endDate,omitempty"`
}

type CreateBooking struct {
	SlotDate      string  `json:"slotDate"`
	SlotStartTime string  `json:"slotStartTime"`
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	Recurrence    *string `json:"recurrence,omitempty"`
	DayOfWeek     *int32  `json:"dayOfWeek,omitempty"`
	EndDate       *string `json:"endDate,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
```

- [ ] **Step 7: Create server entry point**

Create `cmd/server/main.go`:
```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"call-booking/internal/api"
	"call-booking/internal/db"
)

func main() {
	ctx := context.Background()

	pool, err := db.NewPool(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	if err := db.Migrate(ctx, pool, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	router := api.NewRouter(pool)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		fmt.Printf("Server starting on port %s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	fmt.Println("Server stopped")
}
```

- [ ] **Step 8: Verify Go builds**

Run:
```bash
go build ./cmd/server
```

Expected: Binary compiles successfully.

- [ ] **Step 9: Commit**

```bash
git add typespec/ go.mod go.sum cmd/ internal/ migrations/
git commit -m "init: TypeSpec API contract, Go project structure, DB migrations"
```

---

### Task 3: Implement Router and Availability Rules API

**Files:**
- Create: `internal/api/router.go`
- Create: `internal/api/handlers_rules.go`

- [ ] **Step 1: Write tests for availability rules handlers**

Create `internal/api/handlers_rules_test.go`:
```go
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
		Type:      "one-time",
		Date:      strPtr("2026-04-10"),
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/api/... -v -run TestCreateAvailabilityRule
```
Expected: FAIL — handlers don't exist yet.

- [ ] **Step 3: Create router**

Create `internal/api/router.go`:
```go
package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(pool *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Mount("/availability-rules", availabilityRulesRouter(pool))
		r.Mount("/blocked-days", blockedDaysRouter(pool))
		r.Mount("/slots", slotsRouter(pool))
		r.Mount("/bookings", bookingsRouter(pool))
	})

	return r
}
```

- [ ] **Step 4: Implement availability rules handlers**

Create `internal/api/handlers_rules.go`:
```go
package api

import (
	"encoding/json"
	"net/http"

	"call-booking/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type rulesHandler struct {
	pool *pgxpool.Pool
}

func availabilityRulesRouter(pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	h := &rulesHandler{pool: pool}
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	return r
}

func (h *rulesHandler) list(w http.ResponseWriter, r *http.Request) {
	rows, err := h.pool.Query(r.Context(),
		"SELECT id, type, day_of_week, date, time_ranges FROM availability_rules ORDER BY created_at DESC")
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var rules []models.AvailabilityRule
	for rows.Next() {
		var rule models.AvailabilityRule
		var timeRangesJSON []byte
		if err := rows.Scan(&rule.ID, &rule.Type, &rule.DayOfWeek, &rule.Date, &timeRangesJSON); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if err := json.Unmarshal(timeRangesJSON, &rule.TimeRanges); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		rules = append(rules, rule)
	}

	if rules == nil {
		rules = []models.AvailabilityRule{}
	}

	jsonResponse(w, http.StatusOK, rules)
}

func (h *rulesHandler) create(w http.ResponseWriter, r *http.Request) {
	var input models.CreateAvailabilityRule
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	timeRangesJSON, err := json.Marshal(input.TimeRanges)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "invalid time ranges")
		return
	}

	var rule models.AvailabilityRule
	err = h.pool.QueryRow(r.Context(),
		"INSERT INTO availability_rules (type, day_of_week, date, time_ranges) VALUES ($1, $2, $3, $4) RETURNING id, type, day_of_week, date, time_ranges",
		input.Type, input.DayOfWeek, input.Date, timeRangesJSON).
		Scan(&rule.ID, &rule.Type, &rule.DayOfWeek, &rule.Date, &timeRangesJSON)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rule.TimeRanges = input.TimeRanges
	jsonResponse(w, http.StatusCreated, rule)
}

func (h *rulesHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var input models.CreateAvailabilityRule
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	timeRangesJSON, err := json.Marshal(input.TimeRanges)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "invalid time ranges")
		return
	}

	var rule models.AvailabilityRule
	err = h.pool.QueryRow(r.Context(),
		"UPDATE availability_rules SET type=$1, day_of_week=$2, date=$3, time_ranges=$4 WHERE id=$5 RETURNING id, type, day_of_week, date, time_ranges",
		input.Type, input.DayOfWeek, input.Date, timeRangesJSON, id).
		Scan(&rule.ID, &rule.Type, &rule.DayOfWeek, &rule.Date, &timeRangesJSON)
	if err != nil {
		jsonError(w, http.StatusNotFound, "rule not found")
		return
	}

	rule.TimeRanges = input.TimeRanges
	jsonResponse(w, http.StatusOK, rule)
}

func (h *rulesHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, err := h.pool.Exec(r.Context(), "DELETE FROM availability_rules WHERE id = $1", id)
	if err != nil {
		jsonError(w, http.StatusNotFound, "rule not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ErrorResponse{Error: msg})
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run:
```bash
go test ./internal/api/... -v -run "TestCreateAvailabilityRule|TestListAvailabilityRules|TestDeleteAvailabilityRule"
```
Expected: All PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/api/router.go internal/api/handlers_rules.go internal/api/handlers_rules_test.go
git commit -m "feat: implement availability rules CRUD API"
```

---

### Task 4: Implement Blocked Days API

**Files:**
- Create: `internal/api/handlers_blocked_days.go`
- Create: `internal/api/handlers_blocked_days_test.go`

- [ ] **Step 1: Write tests for blocked days handlers**

Create `internal/api/handlers_blocked_days_test.go`:
```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/api/... -v -run "TestCreateBlockedDay|TestDeleteBlockedDay"
```
Expected: FAIL.

- [ ] **Step 3: Implement blocked days handlers**

Create `internal/api/handlers_blocked_days.go`:
```go
package api

import (
	"encoding/json"
	"net/http"

	"call-booking/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type blockedDaysHandler struct {
	pool *pgxpool.Pool
}

func blockedDaysRouter(pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	h := &blockedDaysHandler{pool: pool}
	r.Post("/", h.create)
	r.Delete("/{id}", h.delete)
	return r
}

func (h *blockedDaysHandler) create(w http.ResponseWriter, r *http.Request) {
	var input models.CreateBlockedDay
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var day models.BlockedDay
	err := h.pool.QueryRow(r.Context(),
		"INSERT INTO blocked_days (date) VALUES ($1) RETURNING id, date",
		input.Date).Scan(&day.ID, &day.Date)
	if err != nil {
		jsonError(w, http.StatusConflict, "day already blocked")
		return
	}

	jsonResponse(w, http.StatusCreated, day)
}

func (h *blockedDaysHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, err := h.pool.Exec(r.Context(), "DELETE FROM blocked_days WHERE id = $1", id)
	if err != nil {
		jsonError(w, http.StatusNotFound, "blocked day not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:
```bash
go test ./internal/api/... -v -run "TestCreateBlockedDay|TestDeleteBlockedDay"
```
Expected: All PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/api/handlers_blocked_days.go internal/api/handlers_blocked_days_test.go
git commit -m "feat: implement blocked days API"
```

---

### Task 5: Implement Slot Generation Logic

**Files:**
- Create: `internal/slots/generator.go`
- Create: `internal/slots/generator_test.go`

- [ ] **Step 1: Write tests for slot generator**

Create `internal/slots/generator_test.go`:
```go
package slots

import (
	"testing"
	"time"

	"call-booking/internal/models"
)

func TestGenerateSlotsFromRule(t *testing.T) {
	rule := models.AvailabilityRule{
		Type:      "recurring",
		DayOfWeek: ptrInt32(1), // Monday
		TimeRanges: []models.TimeRange{
			{StartTime: "10:00", EndTime: "11:00"},
		},
	}

	slots := generateSlotsFromRule(rule, "2026-04-06") // Monday

	if len(slots) != 2 {
		t.Fatalf("expected 2 slots, got %d", len(slots))
	}
	if slots[0].StartTime != "10:00" || slots[0].EndTime != "10:30" {
		t.Fatalf("unexpected first slot: %+v", slots[0])
	}
	if slots[1].StartTime != "10:30" || slots[1].EndTime != "11:00" {
		t.Fatalf("unexpected second slot: %+v", slots[1])
	}
}

func TestGenerateSlotsFromRuleMultipleRanges(t *testing.T) {
	rule := models.AvailabilityRule{
		Type: "one-time",
		Date: strPtr("2026-04-10"),
		TimeRanges: []models.TimeRange{
			{StartTime: "10:00", EndTime: "11:00"},
			{StartTime: "14:00", EndTime: "15:00"},
		},
	}

	slots := generateSlotsFromRule(rule, "2026-04-10")

	if len(slots) != 4 {
		t.Fatalf("expected 4 slots, got %d", len(slots))
	}
}

func TestGenerateSlotsEmptyWhenNoRules(t *testing.T) {
	slots := GenerateSlots(nil, nil, nil, "2026-04-06")
	if len(slots) != 0 {
		t.Fatalf("expected 0 slots, got %d", len(slots))
	}
}

func TestGenerateSlotsExcludesBlockedDay(t *testing.T) {
	rule := models.AvailabilityRule{
		Type:      "recurring",
		DayOfWeek: ptrInt32(1),
		TimeRanges: []models.TimeRange{
			{StartTime: "10:00", EndTime: "11:00"},
		},
	}
	blocked := []models.BlockedDay{{Date: "2026-04-06"}}

	slots := GenerateSlots([]models.AvailabilityRule{rule}, blocked, nil, "2026-04-06")
	if len(slots) != 0 {
		t.Fatalf("expected 0 slots on blocked day, got %d", len(slots))
	}
}

func TestAddMinutes(t *testing.T) {
	result := addMinutes("10:00", 30)
	if result != "10:30" {
		t.Fatalf("expected 10:30, got %s", result)
	}

	result = addMinutes("10:30", 30)
	if result != "11:00" {
		t.Fatalf("expected 11:00, got %s", result)
	}

	result = addMinutes("23:30", 30)
	if result != "00:00" {
		t.Fatalf("expected 00:00, got %s", result)
	}
}

func ptrInt32(v int32) *int32 { return &v }
func strPtr(v string) *string  { return &v }
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/slots/... -v
```
Expected: FAIL.

- [ ] **Step 3: Implement slot generator**

Create `internal/slots/generator.go`:
```go
package slots

import (
	"fmt"
	"time"

	"call-booking/internal/models"
)

func GenerateSlots(rules []models.AvailabilityRule, blockedDays []models.BlockedDay, bookings []models.Booking, date string) []models.Slot {
	// Check if day is blocked
	for _, bd := range blockedDays {
		if bd.Date == date {
			return []models.Slot{}
		}
	}

	// Parse the date to get day of week
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return []models.Slot{}
	}
	dayOfWeek := int32(t.Weekday()) // 0=Sunday

	// Collect matching rules
	var matchingRules []models.AvailabilityRule
	for _, rule := range rules {
		if rule.Type == "recurring" && rule.DayOfWeek != nil && *rule.DayOfWeek == dayOfWeek {
			matchingRules = append(matchingRules, rule)
		}
		if rule.Type == "one-time" && rule.Date != nil && *rule.Date == date {
			matchingRules = append(matchingRules, rule)
		}
	}

	// Generate slots from rules
	var slots []models.Slot
	for _, rule := range matchingRules {
		slots = append(slots, generateSlotsFromRule(rule, date)...)
	}

	// Mark booked slots
	for i := range slots {
		for _, booking := range bookings {
			if booking.Status != "active" {
				continue
			}
			if isSlotBooked(slots[i], booking, date) {
				slots[i].IsBooked = true
				break
			}
		}
	}

	return slots
}

func generateSlotsFromRule(rule models.AvailabilityRule, date string) []models.Slot {
	var slots []models.Slot

	for _, tr := range rule.TimeRanges {
		current := tr.StartTime
		for current < tr.EndTime {
			end := addMinutes(current, 30)
			if end > tr.EndTime {
				break
			}
			slot := models.Slot{
				ID:        fmt.Sprintf("%s_%s", date, current),
				Date:      date,
				StartTime: current,
				EndTime:   end,
				IsBooked:  false,
			}
			slots = append(slots, slot)
			current = end
		}
	}

	return slots
}

func addMinutes(t string, minutes int) string {
	parsed, _ := time.Parse("15:04", t)
	result := parsed.Add(time.Duration(minutes) * time.Minute)
	return result.Format("15:04")
}

func isSlotBooked(slot models.Slot, booking models.Booking, date string) bool {
	if booking.SlotDate != date || booking.SlotStartTime != slot.StartTime {
		return false
	}

	rec := "none"
	if booking.Recurrence != nil {
		rec = *booking.Recurrence
	}

	switch rec {
	case "none":
		return true
	case "daily":
		if booking.EndDate == nil {
			return true
		}
		return date <= *booking.EndDate
	case "weekly":
		if booking.DayOfWeek == nil {
			return false
		}
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			return false
		}
		if int32(t.Weekday()) != *booking.DayOfWeek {
			return false
		}
		if booking.EndDate != nil && date > *booking.EndDate {
			return false
		}
		return true
	case "yearly":
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			return false
		}
		bookingDate, _ := time.Parse("2006-01-02", booking.SlotDate)
		if t.Month() != bookingDate.Month() || t.Day() != bookingDate.Day() {
			return false
		}
		if booking.EndDate != nil && date > *booking.EndDate {
			return false
		}
		return true
	default:
		return false
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run:
```bash
go test ./internal/slots/... -v
```
Expected: All PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/slots/
git commit -m "feat: implement slot generation logic with recurrence support"
```

---

### Task 6: Implement Slots and Bookings API

**Files:**
- Create: `internal/api/handlers_slots.go`
- Create: `internal/api/handlers_slots_test.go`
- Create: `internal/api/handlers_bookings.go`
- Create: `internal/api/handlers_bookings_test.go`

- [ ] **Step 1: Write tests for slots and bookings**

Create `internal/api/handlers_slots_test.go`:
```go
package api

import (
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
```

Create `internal/api/handlers_bookings_test.go`:
```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/api/... -v -run "TestGetSlots|TestCreateBooking|TestListBookings|TestDeleteBooking"
```
Expected: FAIL.

- [ ] **Step 3: Implement slots handler**

Create `internal/api/handlers_slots.go`:
```go
package api

import (
	"net/http"

	"call-booking/internal/models"
	"call-booking/internal/slots"

	"github.com/jackc/pgx/v5/pgxpool"
)

type slotsHandler struct {
	pool *pgxpool.Pool
}

func slotsRouter(pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	h := &slotsHandler{pool: pool}
	r.Get("/", h.list)
	return r
}

func (h *slotsHandler) list(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		jsonError(w, http.StatusBadRequest, "date query parameter is required")
		return
	}

	// Fetch rules
	rows, err := h.pool.Query(r.Context(),
		"SELECT id, type, day_of_week, date, time_ranges FROM availability_rules")
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var rules []models.AvailabilityRule
	for rows.Next() {
		var rule models.AvailabilityRule
		var trJSON []byte
		if err := rows.Scan(&rule.ID, &rule.Type, &rule.DayOfWeek, &rule.Date, &trJSON); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if err := json.Unmarshal(trJSON, &rule.TimeRanges); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		rules = append(rules, rule)
	}

	// Fetch blocked days
	blockRows, err := h.pool.Query(r.Context(), "SELECT id, date FROM blocked_days")
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer blockRows.Close()

	var blockedDays []models.BlockedDay
	for blockRows.Next() {
		var bd models.BlockedDay
		if err := blockRows.Scan(&bd.ID, &bd.Date); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		blockedDays = append(blockedDays, bd)
	}

	// Fetch active bookings
	bookingRows, err := h.pool.Query(r.Context(),
		"SELECT id, slot_date, slot_start_time, name, email, status, recurrence, day_of_week, end_date FROM bookings WHERE status = 'active'")
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer bookingRows.Close()

	var bookings []models.Booking
	for bookingRows.Next() {
		var b models.Booking
		if err := bookingRows.Scan(&b.ID, &b.SlotDate, &b.SlotStartTime, &b.Name, &b.Email, &b.Status, &b.Recurrence, &b.DayOfWeek, &b.EndDate); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		bookings = append(bookings, b)
	}

	s := slots.GenerateSlots(rules, blockedDays, bookings, date)
	if s == nil {
		s = []models.Slot{}
	}

	jsonResponse(w, http.StatusOK, s)
}
```

- [ ] **Step 4: Implement bookings handler**

Create `internal/api/handlers_bookings.go`:
```go
package api

import (
	"encoding/json"
	"net/http"

	"call-booking/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type bookingsHandler struct {
	pool *pgxpool.Pool
}

func bookingsRouter(pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	h := &bookingsHandler{pool: pool}
	r.Post("/", h.create)
	r.Get("/", h.list)
	r.Delete("/{id}", h.delete)
	return r
}

func (h *bookingsHandler) create(w http.ResponseWriter, r *http.Request) {
	var input models.CreateBooking
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	recurrence := "none"
	if input.Recurrence != nil {
		recurrence = *input.Recurrence
	}

	var booking models.Booking
	err := h.pool.QueryRow(r.Context(),
		`INSERT INTO bookings (slot_date, slot_start_time, name, email, status, recurrence, day_of_week, end_date)
		 VALUES ($1, $2, $3, $4, 'active', $5, $6, $7)
		 RETURNING id, slot_date, slot_start_time, name, email, status, recurrence, day_of_week, end_date`,
		input.SlotDate, input.SlotStartTime, input.Name, input.Email,
		recurrence, input.DayOfWeek, input.EndDate).
		Scan(&booking.ID, &booking.SlotDate, &booking.SlotStartTime, &booking.Name, &booking.Email,
			&booking.Status, &booking.Recurrence, &booking.DayOfWeek, &booking.EndDate)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, booking)
}

func (h *bookingsHandler) list(w http.ResponseWriter, r *http.Request) {
	rows, err := h.pool.Query(r.Context(),
		"SELECT id, slot_date, slot_start_time, name, email, status, recurrence, day_of_week, end_date FROM bookings ORDER BY created_at DESC")
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var bookings []models.Booking
	for rows.Next() {
		var b models.Booking
		if err := rows.Scan(&b.ID, &b.SlotDate, &b.SlotStartTime, &b.Name, &b.Email, &b.Status, &b.Recurrence, &b.DayOfWeek, &b.EndDate); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		bookings = append(bookings, b)
	}

	if bookings == nil {
		bookings = []models.Booking{}
	}

	jsonResponse(w, http.StatusOK, bookings)
}

func (h *bookingsHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, err := h.pool.Exec(r.Context(),
		"UPDATE bookings SET status = 'cancelled' WHERE id = $1", id)
	if err != nil {
		jsonError(w, http.StatusNotFound, "booking not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 5: Run all tests**

Run:
```bash
go test ./... -v
```
Expected: All PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/api/handlers_slots.go internal/api/handlers_slots_test.go internal/api/handlers_bookings.go internal/api/handlers_bookings_test.go internal/slots/
git commit -m "feat: implement slots and bookings API with full test coverage"
```

---

### Task 7: Initialize Next.js Frontend

**Files:**
- Create: `web/package.json`
- Create: `web/next.config.ts`
- Create: `web/tsconfig.json`
- Create: `web/app/layout.tsx`
- Create: `web/app/page.tsx`
- Create: `web/app/globals.css`

- [ ] **Step 1: Create Next.js project files**

Create `web/package.json`:
```json
{
  "name": "call-booking-web",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start"
  },
  "dependencies": {
    "@mantine/core": "latest",
    "@mantine/dates": "latest",
    "@mantine/hooks": "latest",
    "@tabler/icons-react": "latest",
    "dayjs": "latest",
    "next": "latest",
    "react": "latest",
    "react-dom": "latest"
  },
  "devDependencies": {
    "@types/node": "latest",
    "@types/react": "latest",
    "@types/react-dom": "latest",
    "postcss": "latest",
    "postcss-preset-mantine": "latest",
    "postcss-simple-vars": "latest",
    "typescript": "latest"
  }
}
```

Create `web/next.config.ts`:
```typescript
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  env: {
    API_URL: process.env.API_URL || "http://localhost:8080",
  },
};

export default nextConfig;
```

Create `web/tsconfig.json`:
```json
{
  "compilerOptions": {
    "target": "ES2017",
    "lib": ["dom", "dom.iterable", "esnext"],
    "allowJs": true,
    "skipLibCheck": true,
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "incremental": true,
    "plugins": [{ "name": "next" }],
    "paths": {
      "@/*": ["./*"]
    }
  },
  "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
  "exclude": ["node_modules"]
}
```

Create `web/postcss.config.mjs`:
```javascript
const config = {
  plugins: {
    'postcss-preset-mantine': {},
    'postcss-simple-vars': {
      variables: {
        'mantine-breakpoint-xs': '36em',
        'mantine-breakpoint-sm': '48em',
        'mantine-breakpoint-md': '62em',
        'mantine-breakpoint-lg': '75em',
        'mantine-breakpoint-xl': '88em',
      },
    },
  },
};
export default config;
```

- [ ] **Step 2: Create layout and globals**

Create `web/app/globals.css`:
```css
@import '@mantine/core/styles.css';
@import '@mantine/dates/styles.css';

body {
  margin: 0;
  padding: 0;
}
```

Create `web/app/layout.tsx`:
```typescript
import '@mantine/core/styles.css';
import '@mantine/dates/styles.css';
import { MantineProvider, createTheme } from '@mantine/core';

const theme = createTheme({
  primaryColor: 'blue',
  fontFamily: 'system-ui, -apple-system, sans-serif',
});

export const metadata = {
  title: 'Запись на звонок',
  description: 'Выберите удобное время для звонка',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ru">
      <body>
        <MantineProvider theme={theme}>
          <main style={{ maxWidth: 960, margin: '0 auto', padding: '32px 16px' }}>
            {children}
          </main>
        </MantineProvider>
      </body>
    </html>
  );
}
```

- [ ] **Step 3: Create main page**

Create `web/app/page.tsx`:
```typescript
"use client";

import { useState } from 'react';
import { Paper, Text, Stack, Title } from '@mantine/core';

export default function Home() {
  return (
    <Stack gap="xl">
      <Paper p="md" withBorder>
        <Title order={2}>Запись на звонок</Title>
        <Text c="dimmed" mt="sm">
          Выберите доступное время и запишитесь на звонок.
        </Text>
      </Paper>

      <Paper p="md" withBorder>
        <Title order={4} mb="md">Доступное время</Title>
        <Text c="dimmed">Загрузка слотов...</Text>
      </Paper>

      <Paper p="md" withBorder>
        <Title order={4} mb="md">Предстоящие встречи</Title>
        <Text c="dimmed">Загрузка встреч...</Text>
      </Paper>
    </Stack>
  );
}
```

- [ ] **Step 4: Install dependencies and verify**

Run:
```bash
cd web && npm install
```

Expected: All Mantine packages (`@mantine/core`, `@mantine/dates`, `@mantine/hooks`), `@tabler/icons-react`, `dayjs`, `postcss-preset-mantine`, and `postcss-simple-vars` installed successfully.

- [ ] **Step 5: Commit**

```bash
git add web/
git commit -m "feat: initialize Next.js frontend with Mantine UI"
```

---

### Task 8: Build Frontend Components

**Files:**
- Create: `web/components/SlotPicker.tsx`
- Create: `web/components/BookingForm.tsx`
- Create: `web/components/BookingList.tsx`
- Create: `web/lib/api.ts`
- Modify: `web/app/page.tsx`

- [ ] **Step 1: Create API client**

Create `web/lib/api.ts`:
```typescript
const API_URL = process.env.API_URL || "http://localhost:8080";

export interface Slot {
  id: string;
  date: string;
  startTime: string;
  endTime: string;
  isBooked: boolean;
}

export interface Booking {
  id: string;
  slotDate: string;
  slotStartTime: string;
  name: string;
  email: string;
  status: string;
  recurrence?: string;
  dayOfWeek?: number;
  endDate?: string;
}

export async function getSlots(date: string): Promise<Slot[]> {
  const res = await fetch(`${API_URL}/api/slots?date=${date}`);
  if (!res.ok) throw new Error("Failed to fetch slots");
  return res.json();
}

export async function createBooking(data: {
  slotDate: string;
  slotStartTime: string;
  name: string;
  email: string;
}): Promise<Booking> {
  const res = await fetch(`${API_URL}/api/bookings`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error("Failed to create booking");
  return res.json();
}

export async function getBookings(): Promise<Booking[]> {
  const res = await fetch(`${API_URL}/api/bookings`);
  if (!res.ok) throw new Error("Failed to fetch bookings");
  return res.json();
}

export async function cancelBooking(id: string): Promise<void> {
  const res = await fetch(`${API_URL}/api/bookings/${id}`, { method: "DELETE" });
  if (!res.ok) throw new Error("Failed to cancel booking");
}
```

- [ ] **Step 2: Create SlotPicker component**

Create `web/components/SlotPicker.tsx`:
```typescript
"use client";

import { Group, Button, Text } from '@mantine/core';
import { Slot } from '@/lib/api';

interface SlotPickerProps {
  slots: Slot[];
  selectedSlot: Slot | null;
  onSelect: (slot: Slot) => void;
}

export default function SlotPicker({ slots, selectedSlot, onSelect }: SlotPickerProps) {
  if (slots.length === 0) {
    return <Text c="dimmed">Нет доступных слотов на эту дату</Text>;
  }

  return (
    <Group gap="xs" wrap="wrap">
      {slots.map((slot) => (
        <Button
          key={slot.id}
          variant={selectedSlot?.id === slot.id ? 'filled' : 'outline'}
          disabled={slot.isBooked}
          onClick={() => !slot.isBooked && onSelect(slot)}
          size="sm"
        >
          {slot.startTime}
        </Button>
      ))}
    </Group>
  );
}
```

- [ ] **Step 3: Create BookingForm component**

Create `web/components/BookingForm.tsx`:
```typescript
"use client";

import { useState } from 'react';
import { TextInput, Button, Group, Stack, Text } from '@mantine/core';
import { Slot } from '@/lib/api';

interface BookingFormProps {
  slot: Slot;
  onSubmit: (name: string, email: string) => void;
  onCancel: () => void;
}

export default function BookingForm({ slot, onSubmit, onCancel }: BookingFormProps) {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (name && email) {
      onSubmit(name, email);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <Stack>
        <Text fw={500}>
          Запись на {slot.date} в {slot.startTime}
        </Text>
        <TextInput
          label="Имя"
          value={name}
          onChange={(e) => setName(e.currentTarget.value)}
          required
        />
        <TextInput
          label="Email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.currentTarget.value)}
          required
        />
        <Group>
          <Button type="submit">Записаться</Button>
          <Button variant="default" onClick={onCancel}>Отмена</Button>
        </Group>
      </Stack>
    </form>
  );
}
```

- [ ] **Step 4: Create BookingList component**

Create `web/components/BookingList.tsx`:
```typescript
"use client";

import { Paper, Text, Group, Button, Stack } from '@mantine/core';
import { Booking } from '@/lib/api';

interface BookingListProps {
  bookings: Booking[];
  onCancel: (id: string) => void;
}

export default function BookingList({ bookings, onCancel }: BookingListProps) {
  const active = bookings.filter((b) => b.status === 'active');

  if (active.length === 0) {
    return <Text c="dimmed">Нет предстоящих встреч</Text>;
  }

  return (
    <Stack gap="sm">
      {active.map((booking) => (
        <Paper key={booking.id} p="md" withBorder>
          <Group justify="space-between">
            <div>
              <Text fw={500}>{booking.name}</Text>
              <Text size="sm" c="dimmed">
                {booking.slotDate} в {booking.slotStartTime}
                {booking.recurrence && booking.recurrence !== 'none'
                  ? ` (${booking.recurrence})`
                  : ''}
              </Text>
              <Text size="sm" c="dimmed">{booking.email}</Text>
            </div>
            <Button variant="light" color="red" size="xs" onClick={() => onCancel(booking.id)}>
              Отменить
            </Button>
          </Group>
        </Paper>
      ))}
    </Stack>
  );
}
```

- [ ] **Step 5: Update main page to use components**

Update `web/app/page.tsx`:
```typescript
"use client";

import { useState, useEffect } from 'react';
import { Paper, Text, Stack, Title, Loader, Center } from '@mantine/core';
import { DatePickerInput } from '@mantine/dates';
import { getSlots, createBooking, getBookings, cancelBooking, Slot, Booking } from '@/lib/api';
import SlotPicker from '@/components/SlotPicker';
import BookingForm from '@/components/BookingForm';
import BookingList from '@/components/BookingList';

export default function Home() {
  const [selectedDate, setSelectedDate] = useState<Date | null>(new Date());
  const [slots, setSlots] = useState<Slot[]>([]);
  const [selectedSlot, setSelectedSlot] = useState<Slot | null>(null);
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(true);

  const dateStr = selectedDate?.toISOString().split('T')[0] || '';

  useEffect(() => {
    if (!dateStr) return;
    loadSlots();
    loadBookings();
  }, [dateStr]);

  const loadSlots = async () => {
    try {
      setLoading(true);
      const data = await getSlots(dateStr);
      setSlots(data);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const loadBookings = async () => {
    try {
      const data = await getBookings();
      setBookings(data);
    } catch (e) {
      console.error(e);
    }
  };

  const handleBooking = async (name: string, email: string) => {
    if (!selectedSlot) return;
    await createBooking({
      slotDate: selectedSlot.date,
      slotStartTime: selectedSlot.startTime,
      name,
      email,
    });
    setSelectedSlot(null);
    await loadSlots();
    await loadBookings();
  };

  const handleCancel = async (id: string) => {
    await cancelBooking(id);
    await loadBookings();
    await loadSlots();
  };

  return (
    <Stack gap="xl">
      <Paper p="md" withBorder>
        <Title order={4} mb="md">Выберите дату</Title>
        <DatePickerInput
          value={selectedDate}
          onChange={(val) => {
            setSelectedDate(val);
            setSelectedSlot(null);
          }}
          locale="ru"
          weekendDays={[]}
          minDate={new Date()}
        />
      </Paper>

      <Paper p="md" withBorder>
        <Title order={4} mb="md">
          Доступное время на {dateStr}
        </Title>
        {loading ? (
          <Center>
            <Loader />
          </Center>
        ) : selectedSlot ? (
          <BookingForm
            slot={selectedSlot}
            onSubmit={handleBooking}
            onCancel={() => setSelectedSlot(null)}
          />
        ) : (
          <SlotPicker
            slots={slots}
            selectedSlot={selectedSlot}
            onSelect={setSelectedSlot}
          />
        )}
      </Paper>

      <Paper p="md" withBorder>
        <Title order={4} mb="md">Предстоящие встречи</Title>
        <BookingList bookings={bookings} onCancel={handleCancel} />
      </Paper>
    </Stack>
  );
}
```

- [ ] **Step 6: Verify frontend builds**

Run:
```bash
cd web && npm run build
```
Expected: Build succeeds.

- [ ] **Step 7: Commit**

```bash
git add web/
git commit -m "feat: build frontend components for slot booking"
```

---

### Task 9: Add Docker and Deployment Config

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`
- Create: `.env.example`
- Create: `.gitignore`

- [ ] **Step 1: Create Dockerfile for Go backend**

Create `Dockerfile`:
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /call-booking ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /call-booking .
COPY migrations ./migrations
EXPOSE 8080
CMD ["./call-booking"]
```

- [ ] **Step 2: Create docker-compose.yml**

Create `docker-compose.yml`:
```yaml
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: call_booking
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  api:
    build: .
    environment:
      DATABASE_URL: postgres://postgres:postgres@db:5432/call_booking?sslmode=disable
      PORT: "8080"
    ports:
      - "8080:8080"
    depends_on:
      - db

volumes:
  pgdata:
```

- [ ] **Step 3: Create .env.example**

Create `.env.example`:
```
DATABASE_URL=postgres://postgres:postgres@localhost:5432/call_booking?sslmode=disable
PORT=8080
```

- [ ] **Step 4: Create .gitignore**

Create `.gitignore`:
```
node_modules/
.next/
tsp-output/
*.exe
*.test
.env
vendor/
```

- [ ] **Step 5: Commit**

```bash
git add Dockerfile docker-compose.yml .env.example .gitignore
git commit -m "feat: add Docker and deployment configuration"
```

---

## Self-Review

**1. Spec coverage check:**
- ✅ Availability Rules CRUD — Task 3
- ✅ Blocked Days CRUD — Task 4
- ✅ Slot generation from rules — Task 5
- ✅ Slots API (GET with date query) — Task 6
- ✅ Bookings CRUD with recurrence — Task 6
- ✅ Recurrence logic (daily, weekly, yearly, endDate) — Task 5
- ✅ TimeRanges as array in rules — All tasks
- ✅ TypeSpec API contract — Task 1
- ✅ Go backend with chi + pgx — Tasks 2-6
- ✅ Next.js frontend — Tasks 7-8
- ✅ PostgreSQL with migrations — Task 2
- ✅ Docker setup — Task 9

**2. Placeholder scan:** No TBDs, TODOs, or incomplete sections found.

**3. Type consistency:** All models use consistent types across tasks. `ptrInt32` and `strPtr` helpers are defined where needed. API request/response types match between TypeSpec, Go models, and TypeScript interfaces.
