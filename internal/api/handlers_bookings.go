package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"call-booking/internal/auth"
	"call-booking/internal/models"
	"call-booking/internal/uuid"

	"github.com/go-chi/chi/v5"
)

type bookingsHandler struct {
	db *sql.DB
}

func bookingsRouter(db *sql.DB) http.Handler {
	r := chi.NewRouter()
	h := &bookingsHandler{db: db}

	r.Use(auth.Middleware)
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Delete("/{id}", h.cancel)

	return r
}

func (h *bookingsHandler) list(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	query := `
		SELECT b.id, b.schedule_id,
		       bu.id, bu.email, bu.name,
		       ou.id, ou.email, ou.name,
		       b.slot_date, b.slot_start_time,
		       strftime('%H:%M:%S', datetime(b.slot_date || ' ' || b.slot_start_time, '+30 minutes')),
		       b.status, strftime('%Y-%m-%dT%H:%M:%SZ', b.created_at), b.cancelled_at,
		       COALESCE(GROUP_CONCAT(svg.group_id), '') AS group_ids
		FROM bookings b
		JOIN users bu ON bu.id = b.booker_id
		JOIN users ou ON ou.id = b.owner_id
		JOIN schedules s ON s.id = b.schedule_id
		LEFT JOIN schedule_visibility_groups svg ON svg.schedule_id = b.schedule_id
		WHERE b.booker_id = ? OR b.owner_id = ?
		GROUP BY b.id, b.schedule_id, bu.id, bu.email, bu.name, ou.id, ou.email, ou.name,
		         b.slot_date, b.slot_start_time, b.status, b.created_at, b.cancelled_at
		ORDER BY b.created_at DESC
	`

	rows, err := h.db.QueryContext(r.Context(), query, userID, userID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var bookings []models.Booking
	bookingGroupIDs := make(map[string][]string)

	for rows.Next() {
		var b models.Booking
		var cancelledAt *string
		var groupIDsRaw sql.NullString
		var slotDate, slotStart, endSlot sql.NullString
		if err := rows.Scan(
			&b.ID, &b.ScheduleID,
			&b.Booker.ID, &b.Booker.Email, &b.Booker.Name,
			&b.Owner.ID, &b.Owner.Email, &b.Owner.Name,
			&slotDate, &slotStart, &endSlot,
			&b.Status, &b.CreatedAt, &cancelledAt, &groupIDsRaw); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		b.CancelledAt = cancelledAt
		if slotDate.Valid {
			b.Date = slotDate.String
		}
		if slotStart.Valid {
			b.StartTime = slotStart.String
		}
		if endSlot.Valid {
			b.EndTime = endSlot.String
		}
		b.Groups = []models.VisibilityGroup{}
		cleanGroupIDs := parseGroupConcat(groupIDsRaw)
		if len(cleanGroupIDs) > 0 {
			bookingGroupIDs[b.ID] = cleanGroupIDs
		}
		bookings = append(bookings, b)
	}

	if bookings == nil {
		bookings = []models.Booking{}
	}

	if len(bookingGroupIDs) > 0 {
		groupIDSet := make(map[string]bool)
		for _, ids := range bookingGroupIDs {
			for _, id := range ids {
				groupIDSet[id] = true
			}
		}

		uniqueGroupIDs := make([]string, 0, len(groupIDSet))
		for id := range groupIDSet {
			uniqueGroupIDs = append(uniqueGroupIDs, id)
		}

		ph := placeholders(len(uniqueGroupIDs))
		groupQuery := fmt.Sprintf(`
			SELECT id, owner_id, name, visibility_level
			FROM visibility_groups
			WHERE id IN (%s)
		`, ph)
		args := make([]interface{}, len(uniqueGroupIDs))
		for i, id := range uniqueGroupIDs {
			args[i] = id
		}

		groupRows, err := h.db.QueryContext(r.Context(), groupQuery, args...)
		if err == nil {
			defer groupRows.Close()
			allGroupMap := make(map[string]models.VisibilityGroup)
			for groupRows.Next() {
				var g models.VisibilityGroup
				if err := groupRows.Scan(&g.ID, &g.OwnerID, &g.Name, &g.VisibilityLevel); err == nil {
					allGroupMap[g.ID] = g
				}
			}

			mph := placeholders(len(uniqueGroupIDs))
			memberQuery := fmt.Sprintf(`
				SELECT group_id FROM group_members
				WHERE group_id IN (%s) AND member_id = ?
			`, mph)
			margs := make([]interface{}, 0, len(uniqueGroupIDs)+1)
			for _, id := range uniqueGroupIDs {
				margs = append(margs, id)
			}
			margs = append(margs, userID)

			memberRows, err := h.db.QueryContext(r.Context(), memberQuery, margs...)
			userMemberGroups := make(map[string]bool)
			if err == nil {
				defer memberRows.Close()
				for memberRows.Next() {
					var groupID string
					if err := memberRows.Scan(&groupID); err == nil {
						userMemberGroups[groupID] = true
					}
				}
			}

			for i := range bookings {
				if ids, ok := bookingGroupIDs[bookings[i].ID]; ok {
					isOwner := bookings[i].Owner.ID == userID
					for _, id := range ids {
						if g, found := allGroupMap[id]; found {
							if isOwner || userMemberGroups[id] {
								bookings[i].Groups = append(bookings[i].Groups, g)
							}
						}
					}
				}
			}
		}
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{"bookings": bookings})
}

func (h *bookingsHandler) create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	var req models.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.OwnerID == "" || req.ScheduleID == "" || req.SlotDate == "" {
		jsonError(w, http.StatusBadRequest, "owner_id, schedule_id and slot_date are required")
		return
	}

	slotDateTime, err := parseSlotDateTime(req.SlotDate, req.SlotStartTime)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "invalid slot date or time format")
		return
	}
	if !slotDateTime.After(time.Now()) {
		jsonError(w, http.StatusBadRequest, "cannot book a slot in the past")
		return
	}

	canBook, err := h.canBookSchedule(r.Context(), userID, req.OwnerID, req.ScheduleID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !canBook {
		jsonError(w, http.StatusForbidden, "you don't have access to book this schedule")
		return
	}

	normalizedSlotTime := req.SlotStartTime
	if len(normalizedSlotTime) == 5 {
		normalizedSlotTime = normalizedSlotTime + ":00"
	}

	var exists bool
	err = h.db.QueryRowContext(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM bookings WHERE schedule_id = ? AND slot_date = ? AND slot_start_time = ? AND status = 'active')`,
		req.ScheduleID, req.SlotDate, normalizedSlotTime).Scan(&exists)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if exists {
		jsonError(w, http.StatusConflict, "this slot is already booked")
		return
	}

	var b models.Booking
	bid := uuid.New()
	err = h.db.QueryRowContext(r.Context(),
		`INSERT INTO bookings (id, schedule_id, booker_id, owner_id, slot_start_time, slot_date)
		 VALUES (?, ?, ?, ?, ?, ?)
		 RETURNING id, schedule_id, booker_id, owner_id, status, strftime('%Y-%m-%dT%H:%M:%SZ', created_at)`,
		bid, req.ScheduleID, userID, req.OwnerID, normalizedSlotTime, req.SlotDate).
		Scan(&b.ID, &b.ScheduleID, &b.Booker.ID, &b.Owner.ID, &b.Status, &b.CreatedAt)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = h.db.QueryRowContext(r.Context(),
		"SELECT email, name FROM users WHERE id = ?", userID).Scan(&b.Booker.Email, &b.Booker.Name)
	_ = h.db.QueryRowContext(r.Context(),
		"SELECT email, name FROM users WHERE id = ?", req.OwnerID).Scan(&b.Owner.Email, &b.Owner.Name)

	_ = h.db.QueryRowContext(r.Context(),
		"SELECT date, start_time, end_time FROM schedules WHERE id = ?", req.ScheduleID).
		Scan(&b.Date, &b.StartTime, &b.EndTime)

	jsonResponse(w, http.StatusCreated, b)
}

func parseSlotDateTime(slotDate, slotStartTime string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
	}
	raw := slotDate + " " + slotStartTime
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid slot date or time format")
}

func (h *bookingsHandler) cancel(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	bookingID := chi.URLParam(r, "id")

	var bookerID, ownerID, status string
	var isPast bool
	err := h.db.QueryRowContext(r.Context(), `
		SELECT booker_id, owner_id, status,
			(slot_date IS NOT NULL AND slot_start_time IS NOT NULL AND
			 datetime(slot_date || ' ' || slot_start_time, '+30 minutes') < datetime('now')
			) AS is_past
		FROM bookings WHERE id = ?`,
		bookingID).Scan(&bookerID, &ownerID, &status, &isPast)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonError(w, http.StatusNotFound, "booking not found")
			return
		}
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if bookerID != userID && ownerID != userID {
		jsonError(w, http.StatusForbidden, "you don't have permission to cancel this booking")
		return
	}

	if status != "active" {
		jsonError(w, http.StatusBadRequest, "booking is not active")
		return
	}

	if isPast {
		jsonError(w, http.StatusBadRequest, "cannot cancel a past booking")
		return
	}

	_, err = h.db.ExecContext(r.Context(),
		`UPDATE bookings SET status = 'cancelled',
			cancelled_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now'),
			cancelled_by = ? WHERE id = ?`,
		userID, bookingID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *bookingsHandler) canBookSchedule(ctx context.Context, currentUserID, ownerID, scheduleID string) (bool, error) {
	if currentUserID == ownerID {
		return true, nil
	}

	rows, err := h.db.QueryContext(ctx,
		"SELECT group_id FROM schedule_visibility_groups WHERE schedule_id = ?",
		scheduleID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var groupIDs []string
	for rows.Next() {
		var gid string
		if err := rows.Scan(&gid); err != nil {
			continue
		}
		groupIDs = append(groupIDs, gid)
	}

	if len(groupIDs) == 0 {
		var isPublic bool
		err := h.db.QueryRowContext(ctx,
			"SELECT is_public != 0 FROM users WHERE id = ?",
			ownerID).Scan(&isPublic)
		if err != nil {
			return false, err
		}
		return isPublic, nil
	}

	ph := placeholders(len(groupIDs))
	query := fmt.Sprintf(`
		SELECT EXISTS (
			SELECT 1 FROM group_members gm
			JOIN visibility_groups vg ON vg.id = gm.group_id
			WHERE vg.owner_id = ?
			  AND gm.member_id = ?
			  AND gm.group_id IN (%s)
		)
	`, ph)
	args := []interface{}{ownerID, currentUserID}
	for _, g := range groupIDs {
		args = append(args, g)
	}
	var isMember bool
	err = h.db.QueryRowContext(ctx, query, args...).Scan(&isMember)
	return isMember, err
}
