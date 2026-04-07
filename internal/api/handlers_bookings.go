package api

import (
	"encoding/json"
	"net/http"

	"call-booking/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
	var slotDate, endDate pgtype.Date
	var dayOfWeek pgtype.Int4
	err := h.pool.QueryRow(r.Context(),
		`INSERT INTO bookings (slot_date, slot_start_time, name, email, status, recurrence, day_of_week, end_date)
		 VALUES ($1, $2, $3, $4, 'active', $5, $6, $7)
		 RETURNING id, slot_date, slot_start_time, name, email, status, recurrence, day_of_week, end_date`,
		input.SlotDate, input.SlotStartTime, input.Name, input.Email,
		recurrence, input.DayOfWeek, input.EndDate).
		Scan(&booking.ID, &slotDate, &booking.SlotStartTime, &booking.Name, &booking.Email,
			&booking.Status, &booking.Recurrence, &dayOfWeek, &endDate)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if slotDate.Valid {
		booking.SlotDate = slotDate.Time.Format("2006-01-02")
	}
	if endDate.Valid {
		s := endDate.Time.Format("2006-01-02")
		booking.EndDate = &s
	}
	if dayOfWeek.Valid {
		v := dayOfWeek.Int32
		booking.DayOfWeek = &v
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
		var slotDate, endDate pgtype.Date
		var dayOfWeek pgtype.Int4
		if err := rows.Scan(&b.ID, &slotDate, &b.SlotStartTime, &b.Name, &b.Email, &b.Status, &b.Recurrence, &dayOfWeek, &endDate); err != nil {
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
