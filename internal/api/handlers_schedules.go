package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"call-booking/internal/auth"
	"call-booking/internal/models"
	"call-booking/internal/uuid"

	"github.com/go-chi/chi/v5"
)

type schedulesHandler struct {
	db *sql.DB
}

func schedulesRouter(db *sql.DB) http.Handler {
	r := chi.NewRouter()
	h := &schedulesHandler{db: db}

	r.Use(auth.Middleware)
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)

	return r
}

func (h *schedulesHandler) list(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT s.id, s.user_id, s.type, s.day_of_week, s.date, s.start_time, s.end_time, s.is_blocked,
			s.created_at,
			COALESCE(GROUP_CONCAT(svg.group_id), '') AS group_ids
		 FROM schedules s
		 LEFT JOIN schedule_visibility_groups svg ON svg.schedule_id = s.id
		 WHERE s.user_id = ?
		 GROUP BY s.id, s.user_id, s.type, s.day_of_week, s.date, s.start_time, s.end_time, s.is_blocked, s.created_at
		 ORDER BY s.created_at DESC`,
		userID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	var schedules []models.Schedule
	for rows.Next() {
		var s models.Schedule
		var dayOfWeek *int32
		var date sql.NullString
		var groupIDsRaw sql.NullString
		if err := rows.Scan(&s.ID, &s.UserID, &s.Type, &dayOfWeek, &date, &s.StartTime, &s.EndTime, &s.IsBlocked, &s.CreatedAt, &groupIDsRaw); err != nil {
			jsonError(w, http.StatusInternalServerError, "database error: "+err.Error())
			return
		}
		s.DayOfWeek = dayOfWeek
		if date.Valid {
			s.Date = &date.String
		}
		s.GroupIDs = parseGroupConcat(groupIDsRaw)
		schedules = append(schedules, s)
	}

	if schedules == nil {
		schedules = []models.Schedule{}
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{"schedules": schedules})
}

// filterEmptyUUIDs removes empty/zero UUIDs from the slice
func filterEmptyUUIDs(ids []string) []string {
	var result []string
	emptyUUID := "00000000-0000-0000-0000-000000000000"
	for _, id := range ids {
		if id != "" && id != emptyUUID {
			result = append(result, id)
		}
	}
	return result
}

func (h *schedulesHandler) create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	var req models.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate
	if req.Type != "recurring" && req.Type != "one-time" {
		jsonError(w, http.StatusBadRequest, "type must be 'recurring' or 'one-time'")
		return
	}
	if req.StartTime == "" || req.EndTime == "" {
		jsonError(w, http.StatusBadRequest, "start_time and end_time are required")
		return
	}

	// Format time to HH:MM:SS
	startTime := req.StartTime
	if len(startTime) == 5 {
		startTime = startTime + ":00"
	}
	endTime := req.EndTime
	if len(endTime) == 5 {
		endTime = endTime + ":00"
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error: "+err.Error())
		return
	}
	defer func() { _ = tx.Rollback() }()

	var s models.Schedule
	var dayOfWeek *int32
	var date sql.NullString
	scheduleID := uuid.New()

	var dow interface{}
	if req.DayOfWeek != nil {
		dow = int(*req.DayOfWeek)
	}
	var dateVal interface{}
	if req.Date != nil {
		dateVal = *req.Date
	}

	err = tx.QueryRowContext(r.Context(),
		`INSERT INTO schedules (id, user_id, type, day_of_week, date, start_time, end_time, is_blocked)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 RETURNING id, user_id, type, day_of_week, date, start_time, end_time, is_blocked, created_at`,
		scheduleID, userID, req.Type, dow, dateVal, startTime, endTime, req.IsBlocked).
		Scan(&s.ID, &s.UserID, &s.Type, &dayOfWeek, &date, &s.StartTime, &s.EndTime, &s.IsBlocked, &s.CreatedAt)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error: "+err.Error())
		return
	}
	s.DayOfWeek = dayOfWeek
	if date.Valid {
		s.Date = &date.String
	}

	// Insert group associations if provided
	if len(req.GroupIDs) > 0 {
		for _, groupID := range req.GroupIDs {
			var ownerID string
			err := tx.QueryRowContext(r.Context(),
				"SELECT owner_id FROM visibility_groups WHERE id = ?",
				groupID).Scan(&ownerID)
			if err != nil {
				continue
			}
			if ownerID != userID {
				continue
			}

			_, err = tx.ExecContext(r.Context(),
				`INSERT INTO schedule_visibility_groups (id, schedule_id, group_id) VALUES (?, ?, ?)
				 ON CONFLICT(schedule_id, group_id) DO NOTHING`,
				uuid.New(), s.ID, groupID)
			if err != nil {
				continue
			}
		}
	}

	if err := tx.Commit(); err != nil {
		jsonError(w, http.StatusInternalServerError, "database error: "+err.Error())
		return
	}

	rows, err := h.db.QueryContext(r.Context(),
		"SELECT group_id FROM schedule_visibility_groups WHERE schedule_id = ?",
		s.ID)
	if err == nil {
		defer rows.Close()
		var groupIDs []string
		for rows.Next() {
			var gid string
			if err := rows.Scan(&gid); err == nil {
				groupIDs = append(groupIDs, gid)
			}
		}
		s.GroupIDs = groupIDs
	}

	jsonResponse(w, http.StatusCreated, s)
}

func (h *schedulesHandler) update(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	scheduleID := chi.URLParam(r, "id")

	var req models.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	startTime := req.StartTime
	if len(startTime) == 5 {
		startTime = startTime + ":00"
	}
	endTime := req.EndTime
	if len(endTime) == 5 {
		endTime = endTime + ":00"
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error: "+err.Error())
		return
	}
	defer func() { _ = tx.Rollback() }()

	var s models.Schedule
	var dayOfWeek *int32
	var date sql.NullString

	var updDow interface{}
	if req.DayOfWeek != nil {
		updDow = int(*req.DayOfWeek)
	}
	var updDate interface{}
	if req.Date != nil {
		updDate = *req.Date
	}

	err = tx.QueryRowContext(r.Context(),
		`UPDATE schedules SET type=?, day_of_week=?, date=?, start_time=?, end_time=?, is_blocked=?
		 WHERE id=? AND user_id=?
		 RETURNING id, user_id, type, day_of_week, date, start_time, end_time, is_blocked, created_at`,
		req.Type, updDow, updDate, startTime, endTime, req.IsBlocked, scheduleID, userID).
		Scan(&s.ID, &s.UserID, &s.Type, &dayOfWeek, &date, &s.StartTime, &s.EndTime, &s.IsBlocked, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonError(w, http.StatusNotFound, "schedule not found")
			return
		}
		jsonError(w, http.StatusInternalServerError, "database error: "+err.Error())
		return
	}
	s.DayOfWeek = dayOfWeek
	if date.Valid {
		s.Date = &date.String
	}

	_, err = tx.ExecContext(r.Context(),
		"DELETE FROM schedule_visibility_groups WHERE schedule_id = ?",
		scheduleID)
	if err != nil {
		// Continue even if delete fails
	}

	if len(req.GroupIDs) > 0 {
		for _, groupID := range req.GroupIDs {
			var ownerID string
			err := tx.QueryRowContext(r.Context(),
				"SELECT owner_id FROM visibility_groups WHERE id = ?",
				groupID).Scan(&ownerID)
			if err != nil {
				continue
			}
			if ownerID != userID {
				continue
			}

			_, err = tx.ExecContext(r.Context(),
				`INSERT INTO schedule_visibility_groups (id, schedule_id, group_id) VALUES (?, ?, ?)
				 ON CONFLICT(schedule_id, group_id) DO NOTHING`,
				uuid.New(), scheduleID, groupID)
			if err != nil {
				continue
			}
		}
	}

	if err := tx.Commit(); err != nil {
		jsonError(w, http.StatusInternalServerError, "database error: "+err.Error())
		return
	}

	rows, err := h.db.QueryContext(r.Context(),
		"SELECT group_id FROM schedule_visibility_groups WHERE schedule_id = ?",
		s.ID)
	if err == nil {
		defer rows.Close()
		var groupIDs []string
		for rows.Next() {
			var gid string
			if err := rows.Scan(&gid); err == nil {
				groupIDs = append(groupIDs, gid)
			}
		}
		s.GroupIDs = groupIDs
	}

	jsonResponse(w, http.StatusOK, s)
}

func (h *schedulesHandler) delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	scheduleID := chi.URLParam(r, "id")

	result, err := h.db.ExecContext(r.Context(),
		"DELETE FROM schedules WHERE id = ? AND user_id = ?",
		scheduleID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			_, _ = h.db.ExecContext(r.Context(),
				"DELETE FROM schedule_visibility_groups WHERE schedule_id = ?",
				scheduleID)
			result, err = h.db.ExecContext(r.Context(),
				"DELETE FROM schedules WHERE id = ? AND user_id = ?",
				scheduleID, userID)
		}
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "database error")
			return
		}
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		jsonError(w, http.StatusNotFound, "schedule not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// isValidUUID checks if a string is a valid UUID format
func isValidUUID(u string) bool {
	if len(u) != 36 {
		return false
	}
	parts := strings.Split(u, "-")
	if len(parts) != 5 {
		return false
	}
	return true
}
