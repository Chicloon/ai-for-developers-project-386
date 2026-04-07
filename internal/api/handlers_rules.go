package api

import (
	"encoding/json"
	"net/http"

	"call-booking/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
		var date pgtype.Date
		var dayOfWeek pgtype.Int4
		if err := rows.Scan(&rule.ID, &rule.Type, &dayOfWeek, &date, &timeRangesJSON); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if err := json.Unmarshal(timeRangesJSON, &rule.TimeRanges); err != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if date.Valid {
			s := date.Time.Format("2006-01-02")
			rule.Date = &s
		}
		if dayOfWeek.Valid {
			v := dayOfWeek.Int32
			rule.DayOfWeek = &v
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
	var date pgtype.Date
	var dayOfWeek pgtype.Int4
	err = h.pool.QueryRow(r.Context(),
		"INSERT INTO availability_rules (type, day_of_week, date, time_ranges) VALUES ($1, $2, $3, $4) RETURNING id, type, day_of_week, date, time_ranges",
		input.Type, input.DayOfWeek, input.Date, timeRangesJSON).
		Scan(&rule.ID, &rule.Type, &dayOfWeek, &date, &timeRangesJSON)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rule.TimeRanges = input.TimeRanges
	if date.Valid {
		s := date.Time.Format("2006-01-02")
		rule.Date = &s
	}
	if dayOfWeek.Valid {
		v := dayOfWeek.Int32
		rule.DayOfWeek = &v
	}
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
	var date pgtype.Date
	var dayOfWeek pgtype.Int4
	err = h.pool.QueryRow(r.Context(),
		"UPDATE availability_rules SET type=$1, day_of_week=$2, date=$3, time_ranges=$4 WHERE id=$5 RETURNING id, type, day_of_week, date, time_ranges",
		input.Type, input.DayOfWeek, input.Date, timeRangesJSON, id).
		Scan(&rule.ID, &rule.Type, &dayOfWeek, &date, &timeRangesJSON)
	if err != nil {
		jsonError(w, http.StatusNotFound, "rule not found")
		return
	}

	rule.TimeRanges = input.TimeRanges
	if date.Valid {
		s := date.Time.Format("2006-01-02")
		rule.Date = &s
	}
	if dayOfWeek.Valid {
		v := dayOfWeek.Int32
		rule.DayOfWeek = &v
	}
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
