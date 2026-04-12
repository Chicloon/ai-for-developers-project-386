package api

import (
	"context"
	"net/http"
	"testing"
	"time"

	"call-booking/internal/auth"
	"call-booking/internal/models"
	"call-booking/internal/uuid"
)

func TestUsersList_VisibleUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create current user
	currentEmail := "current@example.com"
	currentUserID := createTestUser(t, db, currentEmail, "password123", "Current User")
	currentToken := getAuthToken(currentUserID, currentEmail)

	// Create public user (visible via public group)
	publicUserEmail := "public@example.com"
	publicUserID := createTestUser(t, db, publicUserEmail, "password123", "Public User")
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = 1 WHERE id = ?", publicUserID)

	// Create private user (not visible)
	privateUserEmail := "private@example.com"
	_ = createTestUser(t, db, privateUserEmail, "password123", "Private User")
	// No group created, so not visible

	// Create user with member group (visible via membership)
	memberUserEmail := "member@example.com"
	memberUserID := createTestUser(t, db, memberUserEmail, "password123", "Member User")
	var groupID string
	_ = db.QueryRowContext(ctx, "INSERT INTO visibility_groups (id, owner_id, name, visibility_level) VALUES (?, ?, 'Work', 'work') RETURNING id", uuid.New(), memberUserID).Scan(&groupID)
	_, _ = db.ExecContext(ctx, "INSERT INTO group_members (id, group_id, member_id, added_by) VALUES (?, ?, ?, ?)", uuid.New(), groupID, currentUserID, currentUserID)

	rr := makeRequest(router, "GET", "/api/users", nil, currentToken)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string][]models.User
	parseResponse(t, rr, &resp)

	// Should see public user and member user, but not private user or self
	foundPublic := false
	foundMember := false
	foundPrivate := false
	foundSelf := false

	for _, u := range resp["users"] {
		if u.ID == publicUserID {
			foundPublic = true
		}
		if u.ID == memberUserID {
			foundMember = true
		}
		if u.ID == currentUserID {
			foundSelf = true
		}
		if u.Email == privateUserEmail {
			foundPrivate = true
		}
	}

	if !foundPublic {
		t.Error("expected to find public user")
	}
	if !foundMember {
		t.Error("expected to find member user")
	}
	if foundPrivate {
		t.Error("should not find private user")
	}
	if foundSelf {
		t.Error("should not see self in list")
	}
}

func TestUsersList_EmptyWhenNoVisibleUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)

	// Create current user
	currentEmail := "current2@example.com"
	currentUserID := createTestUser(t, db, currentEmail, "password123", "Current User")
	currentToken := getAuthToken(currentUserID, currentEmail)

	rr := makeRequest(router, "GET", "/api/users", nil, currentToken)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string][]models.User
	parseResponse(t, rr, &resp)

	if resp["users"] == nil {
		t.Error("expected empty array, not nil")
	}
	if len(resp["users"]) != 0 {
		t.Errorf("expected 0 users, got %d", len(resp["users"]))
	}
}

func TestUsersList_Unauthorized(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := NewRouter(db)

	rr := makeRequest(router, "GET", "/api/users", nil, "")

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestUsersGet_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create current user
	currentEmail := "current3@example.com"
	currentUserID := createTestUser(t, db, currentEmail, "password123", "Current User")
	currentToken := getAuthToken(currentUserID, currentEmail)

	// Create public user
	publicUserEmail := "public2@example.com"
	publicUserID := createTestUser(t, db, publicUserEmail, "password123", "Public User")
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = 1 WHERE id = ?", publicUserID)

	rr := makeRequest(router, "GET", "/api/users/"+publicUserID, nil, currentToken)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.User
	parseResponse(t, rr, &resp)

	if resp.ID != publicUserID {
		t.Errorf("expected user ID %s, got %s", publicUserID, resp.ID)
	}
	if resp.Email != publicUserEmail {
		t.Errorf("expected email %s, got %s", publicUserEmail, resp.Email)
	}
}

func TestUsersGet_NotVisible(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)

	// Create current user
	currentEmail := "current4@example.com"
	currentUserID := createTestUser(t, db, currentEmail, "password123", "Current User")
	currentToken := getAuthToken(currentUserID, currentEmail)

	// Create private user (no visibility)
	privateUserEmail := "private2@example.com"
	privateUserID := createTestUser(t, db, privateUserEmail, "password123", "Private User")

	rr := makeRequest(router, "GET", "/api/users/"+privateUserID, nil, currentToken)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}

	errResp := parseErrorResponse(t, rr)
	if errResp.Error != "you don't have access to this user" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

func TestUsersGet_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := NewRouter(db)

	// Create current user
	currentEmail := "current5@example.com"
	currentUserID := createTestUser(t, db, currentEmail, "password123", "Current User")
	currentToken := getAuthToken(currentUserID, currentEmail)

	missingID := uuid.New()
	rr := makeRequest(router, "GET", "/api/users/"+missingID, nil, currentToken)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestUsersGet_Unauthorized(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := NewRouter(db)

	rr := makeRequest(router, "GET", "/api/users/some-id", nil, "")

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestUsersSlots_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create current user
	currentEmail := "current6@example.com"
	currentUserID := createTestUser(t, db, currentEmail, "password123", "Current User")
	currentToken := getAuthToken(currentUserID, currentEmail)

	// Create owner user with public visibility
	ownerEmail := "owner@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Owner User")
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = 1 WHERE id = ?", ownerID)

	// Create a recurring schedule for the owner
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	dayOfWeek := int32(time.Now().Add(24 * time.Hour).Weekday())

	_, err := db.ExecContext(ctx,
		"INSERT INTO schedules (id, user_id, type, day_of_week, start_time, end_time, is_blocked) VALUES (?, ?, 'recurring', ?, '09:00:00', '17:00:00', 0)",
		uuid.New(), ownerID, dayOfWeek)
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	rr := makeRequest(router, "GET", "/api/users/"+ownerID+"/slots?date="+tomorrow, nil, currentToken)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string][]models.Slot
	parseResponse(t, rr, &resp)

	// Should have 16 slots (9:00-17:00 = 8 hours = 16 x 30-min slots)
	if len(resp["slots"]) != 16 {
		t.Errorf("expected 16 slots, got %d", len(resp["slots"]))
	}

	// Verify first slot
	if len(resp["slots"]) > 0 {
		firstSlot := resp["slots"][0]
		if firstSlot.StartTime != "09:00" {
			t.Errorf("expected first slot at 09:00, got %s", firstSlot.StartTime)
		}
		if firstSlot.EndTime != "09:30" {
			t.Errorf("expected first slot to end at 09:30, got %s", firstSlot.EndTime)
		}
	}
}

func TestUsersSlots_MissingDate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create current user
	currentEmail := "current7@example.com"
	currentUserID := createTestUser(t, db, currentEmail, "password123", "Current User")
	currentToken := getAuthToken(currentUserID, currentEmail)

	// Create owner user with public visibility
	ownerEmail := "owner2@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Owner User")
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = 1 WHERE id = ?", ownerID)

	rr := makeRequest(router, "GET", "/api/users/"+ownerID+"/slots", nil, currentToken)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}

	errResp := parseErrorResponse(t, rr)
	if errResp.Error != "date parameter is required" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

func TestUsersSlots_NotVisible(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)

	// Create current user
	currentEmail := "current8@example.com"
	currentUserID := createTestUser(t, db, currentEmail, "password123", "Current User")
	currentToken := getAuthToken(currentUserID, currentEmail)

	// Create private user
	privateEmail := "private3@example.com"
	privateID := createTestUser(t, db, privateEmail, "password123", "Private User")

	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	rr := makeRequest(router, "GET", "/api/users/"+privateID+"/slots?date="+tomorrow, nil, currentToken)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}
}

func TestUsersSlots_Unauthorized(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := NewRouter(db)

	rr := makeRequest(router, "GET", "/api/users/some-id/slots?date=2024-01-01", nil, "")

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestUsersSlots_BlockedSchedule(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create current user
	currentEmail := "current9@example.com"
	currentUserID := createTestUser(t, db, currentEmail, "password123", "Current User")
	currentToken := getAuthToken(currentUserID, currentEmail)

	// Create owner user with public visibility
	ownerEmail := "owner3@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Owner User")
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = 1 WHERE id = ?", ownerID)

	// Create a blocked schedule
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	dayOfWeek := int32(time.Now().Add(24 * time.Hour).Weekday())

	_, err := db.ExecContext(ctx,
		"INSERT INTO schedules (id, user_id, type, day_of_week, start_time, end_time, is_blocked) VALUES (?, ?, 'recurring', ?, '00:00:00', '23:59:00', 1)",
		uuid.New(), ownerID, dayOfWeek)
	if err != nil {
		t.Fatalf("failed to create blocked schedule: %v", err)
	}

	rr := makeRequest(router, "GET", "/api/users/"+ownerID+"/slots?date="+tomorrow, nil, currentToken)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string][]models.Slot
	parseResponse(t, rr, &resp)

	// Blocked schedule should result in no slots
	if len(resp["slots"]) != 0 {
		t.Errorf("expected 0 slots for blocked day, got %d", len(resp["slots"]))
	}
}

func TestUsersSlots_BookedSlot(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create current user (as booker)
	currentEmail := "booker@example.com"
	currentUserID := createTestUser(t, db, currentEmail, "password123", "Booker User")
	currentToken := getAuthToken(currentUserID, currentEmail)

	// Create owner user with public visibility
	ownerEmail := "owner4@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Owner User")
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = 1 WHERE id = ?", ownerID)

	// Create a schedule and a booking
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	dayOfWeek := int32(time.Now().Add(24 * time.Hour).Weekday())

	var scheduleID string
	err := db.QueryRowContext(ctx,
		"INSERT INTO schedules (id, user_id, type, day_of_week, start_time, end_time, is_blocked) VALUES (?, ?, 'recurring', ?, '09:00:00', '10:00:00', 0) RETURNING id",
		uuid.New(), ownerID, dayOfWeek).Scan(&scheduleID)
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	// Create a booking (slot_date обязателен для фильтра getSlotsForDate)
	_, err = db.ExecContext(ctx,
		"INSERT INTO bookings (id, schedule_id, booker_id, owner_id, status, slot_start_time, slot_date) VALUES (?, ?, ?, ?, 'active', '09:00:00', ?)",
		uuid.New(), scheduleID, currentUserID, ownerID, tomorrow)
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	rr := makeRequest(router, "GET", "/api/users/"+ownerID+"/slots?date="+tomorrow, nil, currentToken)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string][]models.Slot
	parseResponse(t, rr, &resp)

	// Забронированный слот 09:00 не должен попадать в список доступных
	for _, slot := range resp["slots"] {
		if slot.StartTime == "09:00" {
			t.Error("booked 09:00–09:30 slot must not appear in available slots list")
		}
	}
	// Остаётся свободный слот 09:30–10:00
	var has0930 bool
	for _, slot := range resp["slots"] {
		if slot.StartTime == "09:30" {
			has0930 = true
			break
		}
	}
	if !has0930 {
		t.Error("expected free 09:30 slot to still be available")
	}
}

func TestAvailableUsers_ReturnsAllUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create current user
	currentEmail := "current@example.com"
	currentUserID := createTestUser(t, db, currentEmail, "password123", "Current User")
	currentToken := getAuthToken(currentUserID, currentEmail)

	// Create public user
	publicUserEmail := "public@example.com"
	publicUserID := createTestUser(t, db, publicUserEmail, "password123", "Public User")
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = true WHERE id = ?", publicUserID)

	// Create private user (not public, not in any group)
	privateUserEmail := "private@example.com"
	_ = createTestUser(t, db, privateUserEmail, "password123", "Private User")

	// Create another private user
	privateUserEmail2 := "private2@example.com"
	_ = createTestUser(t, db, privateUserEmail2, "password123", "Private User 2")

	rr := makeRequest(router, "GET", "/api/my/available-users", nil, currentToken)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string][]models.User
	parseResponse(t, rr, &resp)

	// Should see all users except self (public and private)
	foundPublic := false
	foundPrivate := false
	foundPrivate2 := false
	foundSelf := false

	for _, u := range resp["users"] {
		if u.ID == publicUserID {
			foundPublic = true
		}
		if u.Email == privateUserEmail {
			foundPrivate = true
		}
		if u.Email == privateUserEmail2 {
			foundPrivate2 = true
		}
		if u.ID == currentUserID {
			foundSelf = true
		}
	}

	if !foundPublic {
		t.Error("expected to find public user")
	}
	if !foundPrivate {
		t.Error("expected to find private user")
	}
	if !foundPrivate2 {
		t.Error("expected to find private user 2")
	}
	if foundSelf {
		t.Error("should not see self in list")
	}

	// Should have exactly 3 users (all except current)
	if len(resp["users"]) != 3 {
		t.Errorf("expected 3 users, got %d", len(resp["users"]))
	}
}

func TestAvailableUsers_Unauthorized(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := NewRouter(db)

	rr := makeRequest(router, "GET", "/api/my/available-users", nil, "")

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

// SetSecret is needed for tests
func init() {
	auth.SetSecret("test-secret-key-minimum-32-characters-long-for-testing-only")
}
