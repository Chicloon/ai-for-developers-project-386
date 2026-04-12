package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"call-booking/internal/auth"
	"call-booking/internal/models"

	"github.com/go-chi/chi/v5"
)

type usersHandler struct {
	db *sql.DB
}

func usersRouter(db *sql.DB) http.Handler {
	r := chi.NewRouter()
	h := &usersHandler{db: db}

	r.Use(auth.Middleware)
	r.Get("/", h.list)
	r.Get("/{id}", h.get)
	r.Get("/{id}/slots", h.slots)
	r.Get("/{id}/slots-range", h.slotsRange)
	r.Get("/{id}/available-dates", h.availableDates)
	r.Get("/{id}/available-dates-range", h.availableDatesRange)

	return r
}

// list returns users visible to current user (public or group members)
func (h *usersHandler) list(w http.ResponseWriter, r *http.Request) {
	currentUserID := auth.GetUserID(r.Context())

	query := `
		SELECT DISTINCT u.id, u.email, u.name, u.is_public FROM users u
		LEFT JOIN visibility_groups vg ON vg.owner_id = u.id
		LEFT JOIN group_members gm ON gm.group_id = vg.id AND gm.member_id = ?
		WHERE u.id != ?
		  AND (
			u.is_public != 0
			OR gm.member_id IS NOT NULL
		  )
		ORDER BY u.name
	`

	rows, err := h.db.QueryContext(r.Context(), query, currentUserID, currentUserID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.IsPublic); err != nil {
			jsonError(w, http.StatusInternalServerError, "database error")
			return
		}
		users = append(users, user)
	}

	if users == nil {
		users = []models.User{}
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{"users": users})
}

// availableUsers returns all registered users except current user
// Used for adding members to groups (no visibility restrictions)
func (h *usersHandler) availableUsers(w http.ResponseWriter, r *http.Request) {
	currentUserID := auth.GetUserID(r.Context())

	query := `
		SELECT id, email, name, is_public FROM users
		WHERE id != ?
		ORDER BY name
	`

	rows, err := h.db.QueryContext(r.Context(), query, currentUserID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.IsPublic); err != nil {
			jsonError(w, http.StatusInternalServerError, "database error")
			return
		}
		users = append(users, user)
	}

	if users == nil {
		users = []models.User{}
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{"users": users})
}

// get returns user profile
func (h *usersHandler) get(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	currentUserID := auth.GetUserID(r.Context())

	var exists bool
	err := h.db.QueryRowContext(r.Context(), "SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", userID).Scan(&exists)
	if err != nil || !exists {
		jsonError(w, http.StatusNotFound, "user not found")
		return
	}

	if !h.canSeeUser(r.Context(), currentUserID, userID) {
		jsonError(w, http.StatusForbidden, "you don't have access to this user")
		return
	}

	var user models.User
	err = h.db.QueryRowContext(r.Context(),
		"SELECT id, email, name FROM users WHERE id = ?",
		userID).
		Scan(&user.ID, &user.Email, &user.Name)
	if err != nil {
		jsonError(w, http.StatusNotFound, "user not found")
		return
	}

	jsonResponse(w, http.StatusOK, user)
}

// slots returns available slots for a user on a specific date
func (h *usersHandler) slots(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	currentUserID := auth.GetUserID(r.Context())
	date := r.URL.Query().Get("date")

	if date == "" {
		jsonError(w, http.StatusBadRequest, "date parameter is required")
		return
	}

	// Check visibility
	if !h.canSeeUser(r.Context(), currentUserID, userID) {
		jsonError(w, http.StatusForbidden, "you don't have access to this user")
		return
	}

	// Get schedules for the date with visibility filtering
	slots, err := h.getSlotsForDate(r.Context(), currentUserID, userID, date)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{"slots": slots})
}

// slotsRange returns slots for a user in a date range (inclusive)
func (h *usersHandler) slotsRange(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	currentUserID := auth.GetUserID(r.Context())
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	if start == "" || end == "" {
		jsonError(w, http.StatusBadRequest, "start and end parameters are required (format: YYYY-MM-DD)")
		return
	}

	startDate, err := time.Parse("2006-01-02", start)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "invalid start date format, expected YYYY-MM-DD")
		return
	}

	endDate, err := time.Parse("2006-01-02", end)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "invalid end date format, expected YYYY-MM-DD")
		return
	}

	if endDate.Before(startDate) {
		jsonError(w, http.StatusBadRequest, "end date must be greater than or equal to start date")
		return
	}

	// Check visibility
	if !h.canSeeUser(r.Context(), currentUserID, userID) {
		jsonError(w, http.StatusForbidden, "you don't have access to this user")
		return
	}

	var allSlots []models.Slot
	for current := startDate; !current.After(endDate); current = current.AddDate(0, 0, 1) {
		dateStr := current.Format("2006-01-02")
		slots, err := h.getSlotsForDate(r.Context(), currentUserID, userID, dateStr)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "database error")
			return
		}
		allSlots = append(allSlots, slots...)
	}

	if allSlots == nil {
		allSlots = []models.Slot{}
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{"slots": allSlots})
}

// availableDates returns dates with available slots for a user in a given month
type availableDate struct {
	Date           string `json:"date"`
	AvailableSlots int    `json:"availableSlots"`
}

func (h *usersHandler) availableDates(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	currentUserID := auth.GetUserID(r.Context())
	month := r.URL.Query().Get("month")

	if month == "" {
		jsonError(w, http.StatusBadRequest, "month parameter is required (format: YYYY-MM)")
		return
	}

	// Check visibility
	if !h.canSeeUser(r.Context(), currentUserID, userID) {
		jsonError(w, http.StatusForbidden, "you don't have access to this user")
		return
	}

	// Parse month to get start and end dates
	monthDate, err := time.Parse("2006-01", month)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "invalid month format, expected YYYY-MM")
		return
	}

	startDate := monthDate.Format("2006-01-02")
	// Get last day of month
	nextMonth := monthDate.AddDate(0, 1, 0)
	endDate := nextMonth.AddDate(0, 0, -1).Format("2006-01-02")

	// Get available dates
	dates, err := h.getAvailableDatesForMonth(r.Context(), currentUserID, userID, startDate, endDate)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}

	if dates == nil {
		dates = []availableDate{}
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{"dates": dates})
}

// availableDatesRange returns dates with available slots for a user in a given date range
func (h *usersHandler) availableDatesRange(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	currentUserID := auth.GetUserID(r.Context())
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	if start == "" || end == "" {
		jsonError(w, http.StatusBadRequest, "start and end parameters are required (format: YYYY-MM-DD)")
		return
	}

	// Validate date formats
	_, err := time.Parse("2006-01-02", start)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "invalid start date format, expected YYYY-MM-DD")
		return
	}
	_, err = time.Parse("2006-01-02", end)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "invalid end date format, expected YYYY-MM-DD")
		return
	}

	// Check visibility
	if !h.canSeeUser(r.Context(), currentUserID, userID) {
		jsonError(w, http.StatusForbidden, "you don't have access to this user")
		return
	}

	// Get available dates for the range
	dates, err := h.getAvailableDatesForRange(r.Context(), currentUserID, userID, start, end)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}

	if dates == nil {
		dates = []availableDate{}
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{"dates": dates})
}

// updateMe updates the current user's profile
func (h *usersHandler) updateMe(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Build dynamic update query
	var setFields []string
	var args []interface{}

	if req.Name != nil {
		setFields = append(setFields, "name = ?")
		args = append(args, *req.Name)
	}

	if req.IsPublic != nil {
		var v int
		if *req.IsPublic {
			v = 1
		}
		setFields = append(setFields, "is_public = ?")
		args = append(args, v)
	}

	if len(setFields) == 0 {
		jsonError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	// Add user ID as the last argument
	args = append(args, userID)

	query := fmt.Sprintf(
		`UPDATE users SET %s, updated_at = strftime('%%Y-%%m-%%dT%%H:%%M:%%SZ', 'now') WHERE id = ? RETURNING id, email, name, is_public, strftime('%%Y-%%m-%%dT%%H:%%M:%%SZ', created_at)`,
		joinStrings(setFields, ", "))

	var user models.User
	err := h.db.QueryRowContext(r.Context(), query, args...).
		Scan(&user.ID, &user.Email, &user.Name, &user.IsPublic, &user.CreatedAt)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}

	jsonResponse(w, http.StatusOK, user)
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// canSeeUser checks if current user can see target user
func (h *usersHandler) canSeeUser(ctx context.Context, currentUserID, targetUserID string) bool {
	if currentUserID == targetUserID {
		return true
	}

	var visible bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM users u
			LEFT JOIN visibility_groups vg ON vg.owner_id = u.id
			LEFT JOIN group_members gm ON gm.group_id = vg.id AND gm.member_id = ?
			WHERE u.id = ?
			  AND (u.is_public != 0 OR gm.member_id IS NOT NULL)
		)
	`
	err := h.db.QueryRowContext(ctx, query, currentUserID, targetUserID).Scan(&visible)
	if err != nil {
		return false
	}
	return visible
}

// getSlotsForDate generates 30-min slots from schedules with visibility filtering
func (h *usersHandler) getSlotsForDate(ctx context.Context, currentUserID, ownerID, date string) ([]models.Slot, error) {
	// Check if owner is public
	var isOwnerPublic bool
	err := h.db.QueryRowContext(ctx, "SELECT is_public != 0 FROM users WHERE id = ?", ownerID).Scan(&isOwnerPublic)
	if err != nil {
		return nil, err
	}

	// Get user's group memberships with the owner
	groupRows, err := h.db.QueryContext(ctx, `
		SELECT vg.id 
		FROM visibility_groups vg
		JOIN group_members gm ON gm.group_id = vg.id
		WHERE vg.owner_id = ? AND gm.member_id = ?
	`, ownerID, currentUserID)
	if err != nil {
		return nil, err
	}
	defer groupRows.Close()

	memberGroupIDs := make(map[string]bool)
	for groupRows.Next() {
		var gid string
		if err := groupRows.Scan(&gid); err != nil {
			continue
		}
		memberGroupIDs[gid] = true
	}

	// Get schedules for the date with visibility filtering
	rows, err := h.db.QueryContext(ctx, `
		SELECT s.id, s.start_time, s.end_time, s.is_blocked,
			COALESCE(GROUP_CONCAT(svg.group_id), '') AS group_ids
		FROM schedules s
		LEFT JOIN schedule_visibility_groups svg ON svg.schedule_id = s.id
		WHERE s.user_id = ?
		  AND (
			  (s.type = 'one-time' AND s.date = ?)
			  OR
			  (s.type = 'recurring' AND s.day_of_week = CAST(strftime('%w', ?) AS INTEGER))
		  )
		GROUP BY s.id, s.start_time, s.end_time, s.is_blocked
		ORDER BY s.start_time
	`, ownerID, date, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type schedule struct {
		id        string
		startTime string
		endTime   string
		isBlocked bool
		groupIDs  []string
	}

	var allSchedules []schedule
	var groupSchedules []schedule
	var generalSchedules []schedule

	for rows.Next() {
		var s schedule
		var groupIDsRaw sql.NullString
		if err := rows.Scan(&s.id, &s.startTime, &s.endTime, &s.isBlocked, &groupIDsRaw); err != nil {
			return nil, err
		}
		s.groupIDs = parseGroupConcat(groupIDsRaw)
		allSchedules = append(allSchedules, s)

		// Separate into group and general schedules
		if len(s.groupIDs) > 0 {
			groupSchedules = append(groupSchedules, s)
		} else {
			generalSchedules = append(generalSchedules, s)
		}
	}

	// Build time ranges covered by ALL group schedules (for exclusion of general slots)
	type timeRange struct {
		start time.Time
		end   time.Time
	}
	var groupTimeRanges []timeRange
	for _, s := range groupSchedules {
		if s.isBlocked {
			continue
		}
		start, _ := time.Parse("15:04:05", s.startTime)
		end, _ := time.Parse("15:04:05", s.endTime)
		groupTimeRanges = append(groupTimeRanges, timeRange{start, end})
	}

	// Determine which schedules are visible to user
	var visibleSchedules []schedule

	// Add group schedules that user has access to
	for _, s := range groupSchedules {
		if s.isBlocked {
			continue
		}
		// Check if user is in any of this schedule's groups
		hasAccess := false
		for _, gid := range s.groupIDs {
			if memberGroupIDs[gid] {
				hasAccess = true
				break
			}
		}
		if hasAccess {
			visibleSchedules = append(visibleSchedules, s)
		}
	}

	// Add general schedules (slots outside group time ranges)
	// General schedules only visible if isPublic=true
	if isOwnerPublic {
		for _, s := range generalSchedules {
			if s.isBlocked {
				continue
			}
			visibleSchedules = append(visibleSchedules, s)
		}
	}

	// Get booked slots - use slot_date to properly track bookings on recurring schedules.
	bookedRows, err := h.db.QueryContext(ctx, `
		SELECT substr(slot_start_time, 1, 5)
		FROM bookings
		WHERE owner_id = ?
		  AND slot_date = ?
		  AND status = 'active'
	`, ownerID, date)
	if err != nil {
		return nil, err
	}
	defer bookedRows.Close()

	bookedTimes := make(map[string]bool)
	for bookedRows.Next() {
		var startTime string
		if err := bookedRows.Scan(&startTime); err != nil {
			return nil, err
		}
		bookedTimes[startTime] = true
	}

	// Generate 30-min slots
	var slots []models.Slot
	slotDuration := 30 * time.Minute

	for _, s := range visibleSchedules {
		if s.isBlocked {
			continue
		}

		start, err := time.Parse("15:04:05", s.startTime)
		if err != nil {
			return nil, fmt.Errorf("invalid start time format: %w", err)
		}
		end, err := time.Parse("15:04:05", s.endTime)
		if err != nil {
			return nil, fmt.Errorf("invalid end time format: %w", err)
		}

		for current := start; current.Before(end); current = current.Add(slotDuration) {
			slotEnd := current.Add(slotDuration)
			if slotEnd.After(end) {
				break
			}

			slotStartStr := current.Format("15:04")

			// For general schedules (no groups), skip slots that fall within any group schedule time range
			// This ensures group schedules have higher priority
			isGeneralSchedule := len(s.groupIDs) == 0
			if isGeneralSchedule {
				skipSlot := false
				for _, tr := range groupTimeRanges {
					if current.Equal(tr.start) || current.After(tr.start) && current.Before(tr.end) {
						skipSlot = true
						break
					}
				}
				if skipSlot {
					continue
				}
			}

			slots = append(slots, models.Slot{
				ID:        s.id + "_" + slotStartStr,
				Date:      date,
				StartTime: slotStartStr,
				EndTime:   slotEnd.Format("15:04"),
				IsBooked:  bookedTimes[slotStartStr],
			})
		}
	}

	// В списке для бронирования только свободные слоты
	var available []models.Slot
	for _, s := range slots {
		if !s.IsBooked {
			available = append(available, s)
		}
	}
	return available, nil
}

// getAvailableDatesForMonth returns all dates in a range that have available slots
func (h *usersHandler) getAvailableDatesForMonth(ctx context.Context, currentUserID, ownerID, startDate, endDate string) ([]availableDate, error) {
	return h.getAvailableDatesForRange(ctx, currentUserID, ownerID, startDate, endDate)
}

// getAvailableDatesForRange returns all dates in a range that have available slots
func (h *usersHandler) getAvailableDatesForRange(ctx context.Context, currentUserID, ownerID, startDate, endDate string) ([]availableDate, error) {
	// Iterate through each day in the range
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	var availableDates []availableDate

	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		dateStr := current.Format("2006-01-02")

		// Get slots for this date
		slots, err := h.getSlotsForDate(ctx, currentUserID, ownerID, dateStr)
		if err != nil {
			return nil, err
		}

		// Count available (not booked) slots
		availableCount := 0
		for _, slot := range slots {
			if !slot.IsBooked {
				availableCount++
			}
		}

		if availableCount > 0 {
			availableDates = append(availableDates, availableDate{
				Date:           dateStr,
				AvailableSlots: availableCount,
			})
		}
	}

	return availableDates, nil
}
