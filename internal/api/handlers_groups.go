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

type groupsHandler struct {
	db *sql.DB
}

func groupsRouter(db *sql.DB) http.Handler {
	r := chi.NewRouter()
	h := &groupsHandler{db: db}

	r.Use(auth.Middleware)
	r.Get("/", h.list)

	// Member management routes
	r.Get("/{id}/members", h.listMembers)
	r.Post("/{id}/members", h.addMember)
	r.Delete("/{id}/members/{memberId}", h.removeMember)

	return r
}

// list returns all fixed groups owned by the current user
func (h *groupsHandler) list(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, owner_id, name, visibility_level, strftime('%Y-%m-%dT%H:%M:%SZ', created_at)
		 FROM visibility_groups
		 WHERE owner_id = ?
		 ORDER BY
		   CASE visibility_level
		     WHEN 'family' THEN 1
		     WHEN 'friends' THEN 2
		     WHEN 'work' THEN 3
		   END`,
		userID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	var groups []models.VisibilityGroup
	for rows.Next() {
		var g models.VisibilityGroup
		if err := rows.Scan(&g.ID, &g.OwnerID, &g.Name, &g.VisibilityLevel, &g.CreatedAt); err != nil {
			jsonError(w, http.StatusInternalServerError, "database error")
			return
		}
		groups = append(groups, g)
	}

	if len(groups) == 0 {
		groupNames := map[string]string{
			"family":  "Семья",
			"work":    "Работа",
			"friends": "Друзья",
		}
		for level, name := range groupNames {
			_, err := h.db.ExecContext(r.Context(),
				`INSERT INTO visibility_groups (id, owner_id, name, visibility_level) VALUES (?, ?, ?, ?)`,
				uuid.New(), userID, name, level)
			if err != nil {
				log.Printf("Failed to create %s group for user %s: %v", level, userID, err)
			}
		}

		rows, err = h.db.QueryContext(r.Context(),
			`SELECT id, owner_id, name, visibility_level, strftime('%Y-%m-%dT%H:%M:%SZ', created_at)
			 FROM visibility_groups
			 WHERE owner_id = ?
			 ORDER BY
			   CASE visibility_level
			     WHEN 'family' THEN 1
			     WHEN 'friends' THEN 2
			     WHEN 'work' THEN 3
			   END`,
			userID)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "database error")
			return
		}
		defer rows.Close()

		for rows.Next() {
			var g models.VisibilityGroup
			if err := rows.Scan(&g.ID, &g.OwnerID, &g.Name, &g.VisibilityLevel, &g.CreatedAt); err != nil {
				jsonError(w, http.StatusInternalServerError, "database error")
				return
			}
			groups = append(groups, g)
		}
	}

	if groups == nil {
		groups = []models.VisibilityGroup{}
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{"groups": groups})
}

// listMembers returns all members of a group (only owner can view)
func (h *groupsHandler) listMembers(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	groupID := chi.URLParam(r, "id")

	var ownerID string
	err := h.db.QueryRowContext(r.Context(),
		"SELECT owner_id FROM visibility_groups WHERE id = ?",
		groupID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonError(w, http.StatusNotFound, "group not found")
			return
		}
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}

	if ownerID != userID {
		jsonError(w, http.StatusForbidden, "you don't have access to this group")
		return
	}

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT gm.id, gm.group_id, gm.added_by, strftime('%Y-%m-%dT%H:%M:%SZ', gm.added_at),
			 u.id, u.email, u.name, u.is_public, strftime('%Y-%m-%dT%H:%M:%SZ', u.created_at)
		 FROM group_members gm
		 JOIN users u ON gm.member_id = u.id
		 WHERE gm.group_id = ?
		 ORDER BY gm.added_at DESC`,
		groupID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	var members []models.GroupMember
	for rows.Next() {
		var m models.GroupMember
		var user models.User
		if err := rows.Scan(&m.ID, &m.GroupID, &m.AddedBy, &m.AddedAt,
			&user.ID, &user.Email, &user.Name, &user.IsPublic, &user.CreatedAt); err != nil {
			jsonError(w, http.StatusInternalServerError, "database error")
			return
		}
		m.Member = user
		members = append(members, m)
	}

	if members == nil {
		members = []models.GroupMember{}
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{"members": members})
}

// addMember adds a member to a group by email or userId (only owner can add)
func (h *groupsHandler) addMember(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	groupID := chi.URLParam(r, "id")

	var req models.AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var ownerID string
	err := h.db.QueryRowContext(r.Context(),
		"SELECT owner_id FROM visibility_groups WHERE id = ?",
		groupID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonError(w, http.StatusNotFound, "group not found")
			return
		}
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}

	if ownerID != userID {
		jsonError(w, http.StatusForbidden, "you don't have access to this group")
		return
	}

	var memberID string
	if req.UserID != nil && *req.UserID != "" {
		err = h.db.QueryRowContext(r.Context(),
			"SELECT id FROM users WHERE id = ?",
			*req.UserID).Scan(&memberID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				jsonError(w, http.StatusNotFound, "user not found")
				return
			}
			jsonError(w, http.StatusInternalServerError, "database error")
			return
		}
	} else if req.Email != nil && *req.Email != "" {
		err = h.db.QueryRowContext(r.Context(),
			"SELECT id FROM users WHERE email = ?",
			*req.Email).Scan(&memberID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				jsonError(w, http.StatusNotFound, "user not found")
				return
			}
			jsonError(w, http.StatusInternalServerError, "database error")
			return
		}
	} else {
		jsonError(w, http.StatusBadRequest, "either email or userId is required")
		return
	}

	if memberID == ownerID {
		jsonError(w, http.StatusBadRequest, "owner cannot be added as a member")
		return
	}

	var m models.GroupMember
	var user models.User
	err = h.db.QueryRowContext(r.Context(),
		`INSERT INTO group_members (id, group_id, member_id, added_by) VALUES (?, ?, ?, ?)
		 RETURNING id, group_id, added_by, strftime('%Y-%m-%dT%H:%M:%SZ', added_at)`,
		uuid.New(), groupID, memberID, userID).
		Scan(&m.ID, &m.GroupID, &m.AddedBy, &m.AddedAt)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			jsonError(w, http.StatusConflict, "user is already a member of this group")
			return
		}
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}

	err = h.db.QueryRowContext(r.Context(),
		`SELECT id, email, name, is_public, strftime('%Y-%m-%dT%H:%M:%SZ', created_at) FROM users WHERE id = ?`,
		memberID).Scan(&user.ID, &user.Email, &user.Name, &user.IsPublic, &user.CreatedAt)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}
	m.Member = user

	jsonResponse(w, http.StatusCreated, m)
}

// removeMember removes a member from a group (only owner can remove)
func (h *groupsHandler) removeMember(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	groupID := chi.URLParam(r, "id")
	memberID := chi.URLParam(r, "memberId")

	var ownerID string
	err := h.db.QueryRowContext(r.Context(),
		"SELECT owner_id FROM visibility_groups WHERE id = ?",
		groupID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonError(w, http.StatusNotFound, "group not found")
			return
		}
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}

	if ownerID != userID {
		jsonError(w, http.StatusForbidden, "you don't have access to this group")
		return
	}

	result, err := h.db.ExecContext(r.Context(),
		"DELETE FROM group_members WHERE id = ? AND group_id = ?",
		memberID, groupID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "database error")
		return
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		jsonError(w, http.StatusNotFound, "member not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
