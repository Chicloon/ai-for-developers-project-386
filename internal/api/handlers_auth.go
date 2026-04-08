package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"call-booking/internal/auth"
	"call-booking/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type authHandler struct {
	pool *pgxpool.Pool
}

func authRouter(pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	h := &authHandler{pool: pool}

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
	err := h.pool.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)",
		req.Email).Scan(&exists)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
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

	// Create user
	var user models.User
	err = h.pool.QueryRow(r.Context(),
		"INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3) RETURNING id, email, name, created_at",
		req.Email, hash, req.Name).
		Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
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
	// #region agent log
	debugLog := func(msg string, data map[string]interface{}) {
		f, _ := os.OpenFile("/home/user/git/ai-for-developers-project-386/.cursor/debug-eb49d8.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			entry := map[string]interface{}{"id": "log_"+fmt.Sprint(time.Now().UnixNano()), "timestamp": time.Now().UnixMilli(), "location": "handlers_auth.go:login", "message": msg, "data": data, "runId": "debug1", "sessionId": "eb49d8"}
			json.NewEncoder(f).Encode(entry)
			f.Close()
		}
	}
	debugLog("login_handler_entry", map[string]interface{}{"pool_is_nil": h.pool == nil, "hypothesisId": "H1"})
	// #endregion
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	// #region agent log
	debugLog("json_decode_error", map[string]interface{}{"error": err.Error(), "hypothesisId": "H1"})
	// #endregion
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
	// #region agent log
	debugLog("before_db_query", map[string]interface{}{"email": req.Email, "pool_is_nil": h.pool == nil, "hypothesisId": "H1"})
	// #endregion
	err := h.pool.QueryRow(r.Context(),
		"SELECT id, email, name, password_hash, created_at FROM users WHERE email = $1",
		req.Email).
		Scan(&user.ID, &user.Email, &user.Name, &passwordHash, &user.CreatedAt)
	// #region agent log
	debugLog("after_db_query", map[string]interface{}{"error": fmt.Sprint(err), "is_no_rows": err == pgx.ErrNoRows, "user_found": err == nil, "hypothesisId": "H1"})
	// #endregion
	if err != nil {
		if err == pgx.ErrNoRows {
			// Run dummy bcrypt to prevent timing attacks
			auth.CheckPassword("dummy", "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy")
			jsonError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		// #region agent log
		debugLog("db_error_not_no_rows", map[string]interface{}{"error": err.Error(), "hypothesisId": "H1"})
		// #endregion
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}

	// Check password
	// #region agent log
	debugLog("before_check_password", map[string]interface{}{"password_hash_length": len(passwordHash), "hypothesisId": "H1"})
	// #endregion
	if !auth.CheckPassword(req.Password, passwordHash) {
		// #region agent log
		debugLog("password_check_failed", map[string]interface{}{"hypothesisId": "H1"})
		// #endregion
		jsonError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	// Generate token
	// #region agent log
	debugLog("before_generate_token", map[string]interface{}{"user_id": user.ID, "email": user.Email, "hypothesisId": "H2"})
	// #endregion
	token, err := auth.GenerateToken(user.ID, user.Email)
	// #region agent log
	debugLog("after_generate_token", map[string]interface{}{"error": fmt.Sprint(err), "token_empty": token == "", "token_length": len(token), "hypothesisId": "H2"})
	// #endregion
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// #region agent log
	debugLog("login_success", map[string]interface{}{"user_id": user.ID, "token_length": len(token), "hypothesisId": "H2"})
	// #endregion
	jsonResponse(w, http.StatusOK, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *authHandler) me(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	var user models.User
	err := h.pool.QueryRow(r.Context(),
		"SELECT id, email, name, created_at FROM users WHERE id = $1",
		userID).
		Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
	if err != nil {
		jsonError(w, http.StatusNotFound, "user not found")
		return
	}

	jsonResponse(w, http.StatusOK, user)
}
