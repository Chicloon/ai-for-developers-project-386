package api

import (
	"context"
	"net/http"
	"testing"
	"time"

	"call-booking/internal/models"
	"call-booking/internal/uuid"
)

func TestBookingsList_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)

	// Create user
	email := "bookings@example.com"
	userID := createTestUser(t, db, email, "password123", "Bookings Test User")
	token := getAuthToken(userID, email)

	rr := makeRequest(router, "GET", "/api/my/bookings", nil, token)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string][]models.Booking
	parseResponse(t, rr, &resp)

	if resp["bookings"] == nil {
		t.Error("expected empty array, not nil")
	}
	if len(resp["bookings"]) != 0 {
		t.Errorf("expected 0 bookings, got %d", len(resp["bookings"]))
	}
}

func TestBookingsList_AsBooker(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create booker user
	bookerEmail := "booker@example.com"
	bookerID := createTestUser(t, db, bookerEmail, "password123", "Booker User")
	bookerToken := getAuthToken(bookerID, bookerEmail)

	// Create owner user with public visibility
	ownerEmail := "ownerbooked@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Owner User")
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = 1 WHERE id = ?", ownerID)

	slotDay := time.Now().Add(48 * time.Hour)
	dayOfWeek := int32(slotDay.Weekday())
	var scheduleID string
	err := db.QueryRowContext(ctx, "INSERT INTO schedules (id, user_id, type, day_of_week, start_time, end_time, is_blocked) VALUES (?, ?, 'recurring', ?, '09:00:00', '17:00:00', 0) RETURNING id", uuid.New(), ownerID, dayOfWeek).Scan(&scheduleID)
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	_, err = db.ExecContext(ctx, "INSERT INTO bookings (id, schedule_id, booker_id, owner_id, status, slot_date, slot_start_time) VALUES (?, ?, ?, ?, 'active', ?, '10:00:00')", uuid.New(), scheduleID, bookerID, ownerID, slotDay.Format("2006-01-02"))
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	rr := makeRequest(router, "GET", "/api/my/bookings", nil, bookerToken)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string][]models.Booking
	parseResponse(t, rr, &resp)

	if len(resp["bookings"]) != 1 {
		t.Errorf("expected 1 booking, got %d", len(resp["bookings"]))
	}

	if len(resp["bookings"]) > 0 {
		booking := resp["bookings"][0]
		if booking.Booker.ID != bookerID {
			t.Errorf("expected booker ID %s, got %s", bookerID, booking.Booker.ID)
		}
		if booking.Owner.ID != ownerID {
			t.Errorf("expected owner ID %s, got %s", ownerID, booking.Owner.ID)
		}
	}
}

func TestBookingsList_AsOwner(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create booker user
	bookerEmail := "booker2@example.com"
	bookerID := createTestUser(t, db, bookerEmail, "password123", "Booker User")

	// Create owner user
	ownerEmail := "ownerbooked2@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Owner User")
	ownerToken := getAuthToken(ownerID, ownerEmail)
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = 1 WHERE id = ?", ownerID)

	slotDay := time.Now().Add(48 * time.Hour)
	dayOfWeek := int32(slotDay.Weekday())
	var scheduleID string
	err := db.QueryRowContext(ctx, "INSERT INTO schedules (id, user_id, type, day_of_week, start_time, end_time, is_blocked) VALUES (?, ?, 'recurring', ?, '09:00:00', '17:00:00', 0) RETURNING id", uuid.New(), ownerID, dayOfWeek).Scan(&scheduleID)
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	_, err = db.ExecContext(ctx, "INSERT INTO bookings (id, schedule_id, booker_id, owner_id, status, slot_date, slot_start_time) VALUES (?, ?, ?, ?, 'active', ?, '10:00:00')", uuid.New(), scheduleID, bookerID, ownerID, slotDay.Format("2006-01-02"))
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	rr := makeRequest(router, "GET", "/api/my/bookings", nil, ownerToken)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string][]models.Booking
	parseResponse(t, rr, &resp)

	if len(resp["bookings"]) != 1 {
		t.Errorf("expected 1 booking, got %d", len(resp["bookings"]))
	}

	if len(resp["bookings"]) > 0 {
		booking := resp["bookings"][0]
		if booking.Owner.ID != ownerID {
			t.Errorf("expected owner ID %s, got %s", ownerID, booking.Owner.ID)
		}
	}
}

func TestBookingsList_Unauthorized(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := NewRouter(db)

	rr := makeRequest(router, "GET", "/api/my/bookings", nil, "")

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestBookingsCreate_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create booker user
	bookerEmail := "booker3@example.com"
	bookerID := createTestUser(t, db, bookerEmail, "password123", "Booker User")
	bookerToken := getAuthToken(bookerID, bookerEmail)

	// Create owner user with public visibility
	ownerEmail := "ownerbooked3@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Owner User")
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = 1 WHERE id = ?", ownerID)

	// Create a schedule for owner (recurring on same weekday as booking date)
	tomorrow := time.Now().Add(24 * time.Hour)
	dayOfWeek := int32(tomorrow.Weekday())
	var scheduleID string
	err := db.QueryRowContext(ctx, "INSERT INTO schedules (id, user_id, type, day_of_week, start_time, end_time, is_blocked) VALUES (?, ?, 'recurring', ?, '09:00:00', '17:00:00', 0) RETURNING id", uuid.New(), ownerID, dayOfWeek).Scan(&scheduleID)
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	req := models.CreateBookingRequest{
		OwnerID:       ownerID,
		ScheduleID:    scheduleID,
		SlotDate:      tomorrow.Format("2006-01-02"),
		SlotStartTime: "10:00",
	}

	rr := makeRequest(router, "POST", "/api/my/bookings", req, bookerToken)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.Booking
	parseResponse(t, rr, &resp)

	if resp.Booker.ID != bookerID {
		t.Errorf("expected booker ID %s, got %s", bookerID, resp.Booker.ID)
	}
	if resp.Owner.ID != ownerID {
		t.Errorf("expected owner ID %s, got %s", ownerID, resp.Owner.ID)
	}
	if resp.Status != "active" {
		t.Errorf("expected status active, got %s", resp.Status)
	}
}

func TestBookingsCreate_MissingOwnerID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)

	// Create booker user
	bookerEmail := "booker4@example.com"
	bookerID := createTestUser(t, db, bookerEmail, "password123", "Booker User")
	bookerToken := getAuthToken(bookerID, bookerEmail)

	req := models.CreateBookingRequest{
		ScheduleID: "some-id",
	}

	rr := makeRequest(router, "POST", "/api/my/bookings", req, bookerToken)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}

	errResp := parseErrorResponse(t, rr)
	if errResp.Error != "owner_id, schedule_id and slot_date are required" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

func TestBookingsCreate_NotVisibleOwner(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create booker user
	bookerEmail := "booker5@example.com"
	bookerID := createTestUser(t, db, bookerEmail, "password123", "Booker User")
	bookerToken := getAuthToken(bookerID, bookerEmail)

	// Create owner user without public visibility
	ownerEmail := "privateowner@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Private Owner")
	// No visibility group created, so not visible

	tomorrow := time.Now().Add(24 * time.Hour)
	dayOfWeek := int32(tomorrow.Weekday())
	var scheduleID string
	err := db.QueryRowContext(ctx, "INSERT INTO schedules (id, user_id, type, day_of_week, start_time, end_time, is_blocked) VALUES (?, ?, 'recurring', ?, '09:00:00', '17:00:00', 0) RETURNING id", uuid.New(), ownerID, dayOfWeek).Scan(&scheduleID)
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	req := models.CreateBookingRequest{
		OwnerID:       ownerID,
		ScheduleID:    scheduleID,
		SlotDate:      tomorrow.Format("2006-01-02"),
		SlotStartTime: "10:00",
	}

	rr := makeRequest(router, "POST", "/api/my/bookings", req, bookerToken)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}

	errResp := parseErrorResponse(t, rr)
	if errResp.Error != "you don't have access to book this schedule" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

func TestBookingsCreate_DuplicateBooking(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create booker user
	bookerEmail := "booker6@example.com"
	bookerID := createTestUser(t, db, bookerEmail, "password123", "Booker User")
	bookerToken := getAuthToken(bookerID, bookerEmail)

	// Create owner user with public visibility
	ownerEmail := "ownerdup@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Owner User")
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = 1 WHERE id = ?", ownerID)

	tomorrow := time.Now().Add(24 * time.Hour)
	dayOfWeek := int32(tomorrow.Weekday())
	slotDate := tomorrow.Format("2006-01-02")
	var scheduleID string
	err := db.QueryRowContext(ctx, "INSERT INTO schedules (id, user_id, type, day_of_week, start_time, end_time, is_blocked) VALUES (?, ?, 'recurring', ?, '09:00:00', '17:00:00', 0) RETURNING id", uuid.New(), ownerID, dayOfWeek).Scan(&scheduleID)
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	// Create first booking
	_, err = db.ExecContext(ctx, "INSERT INTO bookings (id, schedule_id, booker_id, owner_id, status, slot_start_time, slot_date) VALUES (?, ?, ?, ?, 'active', '10:00:00', ?)", uuid.New(), scheduleID, bookerID, ownerID, slotDate)
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	// Try to create duplicate booking
	req := models.CreateBookingRequest{
		OwnerID:       ownerID,
		ScheduleID:    scheduleID,
		SlotDate:      slotDate,
		SlotStartTime: "10:00",
	}

	rr := makeRequest(router, "POST", "/api/my/bookings", req, bookerToken)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rr.Code)
	}

	errResp := parseErrorResponse(t, rr)
	if errResp.Error != "this slot is already booked" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

func TestBookingsCreate_PastSlot(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create booker user
	bookerEmail := "bookerpast@example.com"
	bookerID := createTestUser(t, db, bookerEmail, "password123", "Booker User")
	bookerToken := getAuthToken(bookerID, bookerEmail)

	// Create owner user with public visibility
	ownerEmail := "ownerpast@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Owner User")
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = 1 WHERE id = ?", ownerID)

	// Create a schedule for owner
	var scheduleID string
	dayOfWeek := int32(1)
	err := db.QueryRowContext(ctx, "INSERT INTO schedules (id, user_id, type, day_of_week, start_time, end_time, is_blocked) VALUES (?, ?, 'recurring', ?, '09:00:00', '17:00:00', 0) RETURNING id", uuid.New(), ownerID, dayOfWeek).Scan(&scheduleID)
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	req := models.CreateBookingRequest{
		OwnerID:       ownerID,
		ScheduleID:    scheduleID,
		SlotDate:      time.Now().Add(-24 * time.Hour).Format("2006-01-02"),
		SlotStartTime: "10:00",
	}

	rr := makeRequest(router, "POST", "/api/my/bookings", req, bookerToken)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}

	errResp := parseErrorResponse(t, rr)
	if errResp.Error != "cannot book a slot in the past" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

func TestBookingsCreate_InvalidSlotTimeFormat(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create booker user
	bookerEmail := "bookerinvalidtime@example.com"
	bookerID := createTestUser(t, db, bookerEmail, "password123", "Booker User")
	bookerToken := getAuthToken(bookerID, bookerEmail)

	// Create owner user with public visibility
	ownerEmail := "ownerinvalidtime@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Owner User")
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = 1 WHERE id = ?", ownerID)

	// Create a schedule for owner
	var scheduleID string
	dayOfWeek := int32(1)
	err := db.QueryRowContext(ctx, "INSERT INTO schedules (id, user_id, type, day_of_week, start_time, end_time, is_blocked) VALUES (?, ?, 'recurring', ?, '09:00:00', '17:00:00', 0) RETURNING id", uuid.New(), ownerID, dayOfWeek).Scan(&scheduleID)
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	req := models.CreateBookingRequest{
		OwnerID:       ownerID,
		ScheduleID:    scheduleID,
		SlotDate:      time.Now().Add(24 * time.Hour).Format("2006-01-02"),
		SlotStartTime: "invalid-time",
	}

	rr := makeRequest(router, "POST", "/api/my/bookings", req, bookerToken)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}

	errResp := parseErrorResponse(t, rr)
	if errResp.Error != "invalid slot date or time format" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

func TestBookingsCreate_Unauthorized(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := NewRouter(db)

	req := models.CreateBookingRequest{
		OwnerID:       "owner-id",
		ScheduleID:    "schedule-id",
		SlotDate:      time.Now().Add(24 * time.Hour).Format("2006-01-02"),
		SlotStartTime: "10:00",
	}

	rr := makeRequest(router, "POST", "/api/my/bookings", req, "")

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestBookingsCancel_AsBooker(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create booker user
	bookerEmail := "booker7@example.com"
	bookerID := createTestUser(t, db, bookerEmail, "password123", "Booker User")
	bookerToken := getAuthToken(bookerID, bookerEmail)

	// Create owner user with public visibility
	ownerEmail := "ownercancel@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Owner User")
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = 1 WHERE id = ?", ownerID)

	// Create a schedule for owner
	var scheduleID string
	dayOfWeek := int32(1)
	err := db.QueryRowContext(ctx, "INSERT INTO schedules (id, user_id, type, day_of_week, start_time, end_time, is_blocked) VALUES (?, ?, 'recurring', ?, '09:00:00', '17:00:00', 0) RETURNING id", uuid.New(), ownerID, dayOfWeek).Scan(&scheduleID)
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	// Create a booking
	var bookingID string
	err = db.QueryRowContext(ctx, "INSERT INTO bookings (id, schedule_id, booker_id, owner_id, status) VALUES (?, ?, ?, ?, 'active') RETURNING id", uuid.New(), scheduleID, bookerID, ownerID).Scan(&bookingID)
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	rr := makeRequest(router, "DELETE", "/api/my/bookings/"+bookingID, nil, bookerToken)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestBookingsCancel_AsOwner(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create booker user
	bookerEmail := "booker8@example.com"
	bookerID := createTestUser(t, db, bookerEmail, "password123", "Booker User")

	// Create owner user
	ownerEmail := "ownercancel2@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Owner User")
	ownerToken := getAuthToken(ownerID, ownerEmail)
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = 1 WHERE id = ?", ownerID)

	// Create a schedule for owner
	var scheduleID string
	dayOfWeek := int32(1)
	err := db.QueryRowContext(ctx, "INSERT INTO schedules (id, user_id, type, day_of_week, start_time, end_time, is_blocked) VALUES (?, ?, 'recurring', ?, '09:00:00', '17:00:00', 0) RETURNING id", uuid.New(), ownerID, dayOfWeek).Scan(&scheduleID)
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	// Create a booking
	var bookingID string
	err = db.QueryRowContext(ctx, "INSERT INTO bookings (id, schedule_id, booker_id, owner_id, status) VALUES (?, ?, ?, ?, 'active') RETURNING id", uuid.New(), scheduleID, bookerID, ownerID).Scan(&bookingID)
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	rr := makeRequest(router, "DELETE", "/api/my/bookings/"+bookingID, nil, ownerToken)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestBookingsCancel_PastSlot(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	bookerEmail := "bookerpast@example.com"
	bookerID := createTestUser(t, db, bookerEmail, "password123", "Booker User")
	bookerToken := getAuthToken(bookerID, bookerEmail)

	ownerEmail := "ownerpast@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Owner User")
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = 1 WHERE id = ?", ownerID)

	var scheduleID string
	dayOfWeek := int32(1)
	err := db.QueryRowContext(ctx, "INSERT INTO schedules (id, user_id, type, day_of_week, start_time, end_time, is_blocked) VALUES (?, ?, 'recurring', ?, '09:00:00', '17:00:00', 0) RETURNING id", uuid.New(), ownerID, dayOfWeek).Scan(&scheduleID)
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	var bookingID string
	err = db.QueryRowContext(ctx,
		`INSERT INTO bookings (id, schedule_id, booker_id, owner_id, status, slot_date, slot_start_time)
		 VALUES (?, ?, ?, ?, 'active', '2000-01-01', '10:00:00') RETURNING id`,
		uuid.New(), scheduleID, bookerID, ownerID).Scan(&bookingID)
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	rr := makeRequest(router, "DELETE", "/api/my/bookings/"+bookingID, nil, bookerToken)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}

	errResp := parseErrorResponse(t, rr)
	if errResp.Error != "cannot cancel a past booking" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

func TestBookingsCancel_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)

	// Create user
	email := "booker9@example.com"
	userID := createTestUser(t, db, email, "password123", "User")
	token := getAuthToken(userID, email)

	rr := makeRequest(router, "DELETE", "/api/my/bookings/nonexistent-id", nil, token)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}

	errResp := parseErrorResponse(t, rr)
	if errResp.Error != "booking not found" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

func TestBookingsCancel_NotAuthorized(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create booker user
	bookerEmail := "booker10@example.com"
	bookerID := createTestUser(t, db, bookerEmail, "password123", "Booker User")

	// Create owner user
	ownerEmail := "ownercancel3@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Owner User")
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = 1 WHERE id = ?", ownerID)

	// Create third user (unauthorized)
	otherEmail := "otherbooker@example.com"
	otherID := createTestUser(t, db, otherEmail, "password123", "Other User")
	otherToken := getAuthToken(otherID, otherEmail)

	// Create a schedule for owner
	var scheduleID string
	dayOfWeek := int32(1)
	err := db.QueryRowContext(ctx, "INSERT INTO schedules (id, user_id, type, day_of_week, start_time, end_time, is_blocked) VALUES (?, ?, 'recurring', ?, '09:00:00', '17:00:00', 0) RETURNING id", uuid.New(), ownerID, dayOfWeek).Scan(&scheduleID)
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	// Create a booking
	var bookingID string
	err = db.QueryRowContext(ctx, "INSERT INTO bookings (id, schedule_id, booker_id, owner_id, status) VALUES (?, ?, ?, ?, 'active') RETURNING id", uuid.New(), scheduleID, bookerID, ownerID).Scan(&bookingID)
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	rr := makeRequest(router, "DELETE", "/api/my/bookings/"+bookingID, nil, otherToken)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}

	errResp := parseErrorResponse(t, rr)
	if errResp.Error != "you don't have permission to cancel this booking" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

func TestBookingsCancel_Unauthorized(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := NewRouter(db)

	rr := makeRequest(router, "DELETE", "/api/my/bookings/some-id", nil, "")

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestBookingsList_WithCancelled(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create booker user
	bookerEmail := "booker11@example.com"
	bookerID := createTestUser(t, db, bookerEmail, "password123", "Booker User")
	bookerToken := getAuthToken(bookerID, bookerEmail)

	// Create owner user
	ownerEmail := "ownercancelled@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Owner User")
	_, _ = db.ExecContext(ctx, "UPDATE users SET is_public = 1 WHERE id = ?", ownerID)

	// Create a schedule for owner
	var scheduleID string
	dayOfWeek := int32(1)
	err := db.QueryRowContext(ctx, "INSERT INTO schedules (id, user_id, type, day_of_week, start_time, end_time, is_blocked) VALUES (?, ?, 'recurring', ?, '09:00:00', '17:00:00', 0) RETURNING id", uuid.New(), ownerID, dayOfWeek).Scan(&scheduleID)
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	// Create a cancelled booking
	cancelledAt := time.Now()
	_, err = db.ExecContext(ctx, "INSERT INTO bookings (id, schedule_id, booker_id, owner_id, status, cancelled_at, cancelled_by) VALUES (?, ?, ?, ?, 'cancelled', ?, ?)", uuid.New(), scheduleID, bookerID, ownerID, cancelledAt.Format(time.RFC3339), bookerID)
	if err != nil {
		t.Fatalf("failed to create cancelled booking: %v", err)
	}

	rr := makeRequest(router, "GET", "/api/my/bookings", nil, bookerToken)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string][]models.Booking
	parseResponse(t, rr, &resp)

	if len(resp["bookings"]) != 1 {
		t.Errorf("expected 1 booking, got %d", len(resp["bookings"]))
	}

	if len(resp["bookings"]) > 0 {
		booking := resp["bookings"][0]
		if booking.Status != "cancelled" {
			t.Errorf("expected status cancelled, got %s", booking.Status)
		}
		if booking.CancelledAt == nil {
			t.Error("expected cancelled_at to be set")
		}
	}
}
