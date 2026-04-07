package api

import (
	"encoding/json"
	"net/http"

	"call-booking/internal/models"
	"call-booking/internal/slots"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
		var dateCol pgtype.Date
		var dayOfWeek pgtype.Int4
		if err := rows.Scan(&rule.ID, &rule.Type, &dayOfWeek, &dateCol, &trJSON); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if err := json.Unmarshal(trJSON, &rule.TimeRanges); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if dateCol.Valid {
			s := dateCol.Time.Format("2006-01-02")
			rule.Date = &s
		}
		if dayOfWeek.Valid {
			v := dayOfWeek.Int32
			rule.DayOfWeek = &v
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
		var dateCol pgtype.Date
		if err := blockRows.Scan(&bd.ID, &dateCol); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if dateCol.Valid {
			bd.Date = dateCol.Time.Format("2006-01-02")
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
		var slotDate, endDate pgtype.Date
		var dayOfWeek pgtype.Int4
		if err := bookingRows.Scan(&b.ID, &slotDate, &b.SlotStartTime, &b.Name, &b.Email, &b.Status, &b.Recurrence, &dayOfWeek, &endDate); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if slotDate.Valid {
			b.SlotDate = slotDate.Time.Format("2006-01-02")
		}
		if endDate.Valid {
			s := endDate.Time.Format("2006-01-02")
			b.EndDate = &s
		}
		if dayOfWeek.Valid {
			v := dayOfWeek.Int32
			b.DayOfWeek = &v
		}
		bookings = append(bookings, b)
	}

	s := slots.GenerateSlots(rules, blockedDays, bookings, date)
	if s == nil {
		s = []models.Slot{}
	}

	jsonResponse(w, http.StatusOK, s)
}
