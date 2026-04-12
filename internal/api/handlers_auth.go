package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"call-booking/internal/auth"
	"call-booking/internal/models"
	"call-booking/internal/uuid"

	"github.com/go-chi/chi/v5"
)

// mapAuthDBError turns DB errors into a short client message; details go to logs.
func mapAuthDBError(err error) string {
	if err == nil {
		return "database error"
	}
	s := err.Error()
	switch {
	case strings.Contains(s, "no such table"):
		return "Схема БД не готова: выполните миграции или проверьте DATABASE_URL."
	case strings.Contains(s, "no such column"):
		return "Схема БД не совпадает с приложением: примените миграции из каталога migrations."
	case strings.Contains(s, "password authentication failed"):
		return "Ошибка подключения к БД: неверные учётные данные."
	case strings.Contains(s, "timeout") || strings.Contains(s, "connection refused") || strings.Contains(s, "no such host"):
		return "База данных недоступна: проверьте DATABASE_URL и сеть."
	default:
		return "database error"
	}
}

type authHandler struct {
	db *sql.DB
}

func authRouter(db *sql.DB) http.Handler {
	r := chi.NewRouter()
	h := &authHandler{db: db}

	r.Post("/register", h.register)
	r.Post("/login", h.login)
	r.With(auth.Middleware).Get("/me", h.me)

	return r
}

func (h *authHandler) register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.Name == "" {
		jsonError(w, http.StatusBadRequest, "email, password and name are required")
		return
	}

	// Check if user exists
	var exists bool
	err := h.db.QueryRowContext(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)",
		req.Email).Scan(&exists)
	if err != nil {
		log.Printf("register: check email exists: %v", err)
		jsonError(w, http.StatusInternalServerError, mapAuthDBError(err))
		return
	}
	if exists {
		jsonError(w, http.StatusConflict, "user with this email already exists")
		return
	}

	// Hash password
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	userID := uuid.New()

	// Create user
	var user models.User
	err = h.db.QueryRowContext(r.Context(),
		`INSERT INTO users (id, email, password_hash, name) VALUES (?, ?, ?, ?)
		 RETURNING id, email, name, is_public, strftime('%Y-%m-%dT%H:%M:%SZ', created_at)`,
		userID, req.Email, hash, req.Name).
		Scan(&user.ID, &user.Email, &user.Name, &user.IsPublic, &user.CreatedAt)
	if err != nil {
		log.Printf("register: insert user: %v", err)
		jsonError(w, http.StatusInternalServerError, mapAuthDBError(err))
		return
	}

	// Create fixed visibility groups for the new user
	groupNames := map[string]string{
		"family":  "Семья",
		"work":    "Работа",
		"friends": "Друзья",
	}
	for level, name := range groupNames {
		_, err := h.db.ExecContext(r.Context(),
			`INSERT INTO visibility_groups (id, owner_id, name, visibility_level) VALUES (?, ?, ?, ?)`,
			uuid.New(), user.ID, name, level)
		if err != nil {
			// Log error but don't fail registration
			log.Printf("Failed to create %s group for user %s: %v", level, user.ID, err)
		}
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	jsonResponse(w, http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		jsonError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	// Find user
	var user models.User
	var passwordHash string
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, email, name, is_public, password_hash,
			strftime('%Y-%m-%dT%H:%M:%SZ', created_at) FROM users WHERE email = ?`,
		req.Email).
		Scan(&user.ID, &user.Email, &user.Name, &user.IsPublic, &passwordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Run dummy bcrypt to prevent timing attacks
			auth.CheckPassword("dummy", "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy")
			jsonError(w, http.StatusUnauthorized, "Неверный email или пароль")
			return
		}
		log.Printf("login: select user: %v", err)
		jsonError(w, http.StatusInternalServerError, mapAuthDBError(err))
		return
	}

	// Check password
	if !auth.CheckPassword(req.Password, passwordHash) {
		jsonError(w, http.StatusUnauthorized, "Неверный email или пароль")
		return
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	jsonResponse(w, http.StatusOK, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *authHandler) me(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	var user models.User
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, email, name, is_public, strftime('%Y-%m-%dT%H:%M:%SZ', created_at) FROM users WHERE id = ?`,
		userID).
		Scan(&user.ID, &user.Email, &user.Name, &user.IsPublic, &user.CreatedAt)
	if err != nil {
		jsonError(w, http.StatusNotFound, "user not found")
		return
	}

	jsonResponse(w, http.StatusOK, user)
}
