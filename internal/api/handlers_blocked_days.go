package api

import (
	"encoding/json"
	"net/http"

	"call-booking/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type blockedDaysHandler struct {
	pool *pgxpool.Pool
}

func blockedDaysRouter(pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	h := &blockedDaysHandler{pool: pool}
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Delete("/{id}", h.delete)
	return r
}

func (h *blockedDaysHandler) list(w http.ResponseWriter, r *http.Request) {
	rows, err := h.pool.Query(r.Context(), "SELECT id, date FROM blocked_days ORDER BY date")
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to fetch blocked days")
		return
	}
	defer rows.Close()

	days := make([]models.BlockedDay, 0)
	for rows.Next() {
		var day models.BlockedDay
		var date pgtype.Date
		if err := rows.Scan(&day.ID, &date); err != nil {
			continue
		}
		if date.Valid {
			day.Date = date.Time.Format("2006-01-02")
		}
		days = append(days, day)
	}

	jsonResponse(w, http.StatusOK, days)
}

func (h *blockedDaysHandler) create(w http.ResponseWriter, r *http.Request) {
	var input models.CreateBlockedDay
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var day models.BlockedDay
	var date pgtype.Date
	err := h.pool.QueryRow(r.Context(),
		"INSERT INTO blocked_days (date) VALUES ($1) RETURNING id, date",
		input.Date).Scan(&day.ID, &date)
	if err != nil {
		jsonError(w, http.StatusConflict, "day already blocked")
		return
	}

	if date.Valid {
		day.Date = date.Time.Format("2006-01-02")
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
