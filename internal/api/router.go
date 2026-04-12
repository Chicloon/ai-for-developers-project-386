package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"call-booking/internal/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func NewRouter(sqldb *sql.DB) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware - ALL middleware must be defined before routes
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, http.StatusOK, map[string]string{"status": "ok", "service": "api"})
	})

	// Root: browsers opening http://localhost:8080/ get a useful response (API has no HTML UI)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Call Booking API — use /health and /api/*\n"))
	})

	// Public routes (no JWT required)
	r.Mount("/api/auth", authRouter(sqldb))

	// Protected routes (JWT required)
	r.Route("/api", func(r chi.Router) {
		r.Use(auth.Middleware)
		r.Mount("/users", usersRouter(sqldb))
		r.Mount("/my/schedules", schedulesRouter(sqldb))
		r.Mount("/my/groups", groupsRouter(sqldb))
		r.Mount("/my/bookings", bookingsRouter(sqldb))

		// Available users for group member addition (all users except current)
		r.Get("/my/available-users", func(w http.ResponseWriter, r *http.Request) {
			h := &usersHandler{db: sqldb}
			h.availableUsers(w, r)
		})
	})

	// Route for current user profile updates
	r.With(auth.Middleware).Put("/api/users/me", func(w http.ResponseWriter, r *http.Request) {
		h := &usersHandler{db: sqldb}
		h.updateMe(w, r)
	})

	return r
}
