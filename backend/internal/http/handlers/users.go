package handlers

import (
	"encoding/json"
	"net/http"

	"go-nmos/backend/internal/models"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.ListUsers(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list users failed"})
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username and password are required"})
		return
	}
	if req.Role == "" {
		req.Role = "viewer"
	}
	if req.Role != "admin" && req.Role != "editor" && req.Role != "viewer" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "role must be one of admin, editor, viewer"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "password hash failed"})
		return
	}

	if err := h.repo.CreateUser(r.Context(), models.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		Role:         req.Role,
	}); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]bool{"ok": true})
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username is required"})
		return
	}

	// Check if user exists
	_, err := h.repo.GetUserByUsername(r.Context(), username)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	var req struct {
		Password string `json:"password,omitempty"`
		Role     string `json:"role,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	updates := map[string]any{}
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "password hash failed"})
			return
		}
		updates["password"] = string(hash)
	}
	if req.Role != "" {
		if req.Role != "admin" && req.Role != "editor" && req.Role != "viewer" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "role must be one of admin, editor, viewer"})
			return
		}
		updates["role"] = req.Role
	}

	if len(updates) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no updates provided"})
		return
	}

	if err := h.repo.UpdateUser(r.Context(), username, updates); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username is required"})
		return
	}

	// Prevent deleting the last admin user
	users, err := h.repo.ListUsers(r.Context())
	if err == nil {
		adminCount := 0
		for _, u := range users {
			if u.Role == "admin" {
				adminCount++
			}
		}
		user, _ := h.repo.GetUserByUsername(r.Context(), username)
		if user != nil && user.Role == "admin" && adminCount == 1 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot delete the last admin user"})
			return
		}
	}

	if err := h.repo.DeleteUser(r.Context(), username); err != nil {
		if err.Error() == "user not found" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
