package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"call-booking/internal/auth"
	"call-booking/internal/db"
	"call-booking/internal/models"
	"call-booking/internal/uuid"
)

func migrationsDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for _, rel := range []string{"migrations", filepath.Join("..", "..", "migrations")} {
		p := filepath.Join(wd, rel)
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p
		}
	}
	t.Fatalf("migrations directory not found (cwd=%s)", wd)
	return ""
}

// setupTestDB creates a SQLite database for testing (temp file by default).
// Skips the test if the database is unavailable.
func setupTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		p := filepath.Join(t.TempDir(), "test.db")
		dsn = "file:" + p + "?_foreign_keys=on"
	}

	ctx := context.Background()
	sqldb, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Skipf("Database not available: %v", err)
	}

	if err := sqldb.PingContext(ctx); err != nil {
		_ = sqldb.Close()
		t.Skipf("Database not available: %v", err)
	}

	if err := db.Migrate(ctx, sqldb, migrationsDir(t)); err != nil {
		_ = sqldb.Close()
		t.Fatalf("migrations failed: %v", err)
	}

	return sqldb
}

// cleanupTestData removes test data from the database.
func cleanupTestData(t *testing.T, sqldb *sql.DB) {
	ctx := context.Background()
	_, _ = sqldb.ExecContext(ctx, "DELETE FROM bookings")
	_, _ = sqldb.ExecContext(ctx, "DELETE FROM schedule_visibility_groups")
	_, _ = sqldb.ExecContext(ctx, "DELETE FROM group_members")
	_, _ = sqldb.ExecContext(ctx, "DELETE FROM visibility_groups")
	_, _ = sqldb.ExecContext(ctx, "DELETE FROM schedules")
	_, _ = sqldb.ExecContext(ctx, "DELETE FROM users")
}

// createTestUser creates a test user (with default visibility groups, same as registration) and returns the user ID.
func createTestUser(t *testing.T, sqldb *sql.DB, email, password, name string) string {
	ctx := context.Background()

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	userID := uuid.New()
	_, err = sqldb.ExecContext(ctx,
		`INSERT INTO users (id, email, password_hash, name) VALUES (?, ?, ?, ?)`,
		userID, email, hash, name)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	groupNames := map[string]string{
		"family":  "Семья",
		"work":    "Работа",
		"friends": "Друзья",
	}
	for level, gname := range groupNames {
		_, err := sqldb.ExecContext(ctx,
			`INSERT INTO visibility_groups (id, owner_id, name, visibility_level) VALUES (?, ?, ?, ?)`,
			uuid.New(), userID, gname, level)
		if err != nil {
			t.Fatalf("failed to create group %s: %v", level, err)
		}
	}

	return userID
}

// getAuthToken generates a JWT token for testing.
func getAuthToken(userID, email string) string {
	auth.SetSecret("test-secret-key-minimum-32-characters-long-for-testing-only")

	token, err := auth.GenerateToken(userID, email)
	if err != nil {
		panic("failed to generate test token: " + err.Error())
	}
	return token
}

// makeRequest creates and executes an HTTP request for testing.
func makeRequest(router http.Handler, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// parseResponse parses the response body into the target struct.
func parseResponse(t *testing.T, rr *httptest.ResponseRecorder, target interface{}) {
	if err := json.NewDecoder(rr.Body).Decode(target); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
}

// parseErrorResponse parses an error response.
func parseErrorResponse(t *testing.T, rr *httptest.ResponseRecorder) models.ErrorResponse {
	var errResp models.ErrorResponse
	parseResponse(t, rr, &errResp)
	return errResp
}

// ptrInt32 returns a pointer to an int32 value.
func ptrInt32(v int32) *int32 {
	return &v
}

// strPtr returns a pointer to a string value.
func strPtr(v string) *string {
	return &v
}
