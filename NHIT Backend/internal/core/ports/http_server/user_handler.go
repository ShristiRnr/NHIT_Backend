package http_server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
)

type UserHandler struct {
	svc *services.UserService
}

func NewUserHandler(svc *services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// RegisterRoutes registers user endpoints
func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Post("/users", h.CreateUser)
	r.Get("/users/{userID}", h.GetUser)
	r.Put("/users/{userID}", h.UpdateUser)
	r.Delete("/users/{userID}", h.DeleteUser)
	r.Get("/tenants/{tenantID}/users", h.ListUsers)
	r.Get("/users/{userID}/permissions", h.GetPermissions)
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req db.CreateUserParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.svc.CreateUser(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// GetUser handles GET /users/{userID}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "invalid userID", http.StatusBadRequest)
		return
	}

	user, err := h.svc.GetUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}

// UpdateUser handles PUT /users/{userID}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "invalid userID", http.StatusBadRequest)
		return
	}

	var req db.UpdateUserParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.UserID = userID

	user, err := h.svc.UpdateUser(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}

// DeleteUser handles DELETE /users/{userID}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "invalid userID", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteUser(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListUsers handles GET /tenants/{tenantID}/users?limit=10&offset=0
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "tenantID"))
	if err != nil {
		http.Error(w, "invalid tenantID", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	users, err := h.svc.ListUsers(r.Context(), tenantID, int32(limit), int32(offset))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

// GetPermissions handles GET /users/{userID}/permissions
func (h *UserHandler) GetPermissions(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "invalid userID", http.StatusBadRequest)
		return
	}

	perms, err := h.svc.GetUserPermissions(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(perms)
}
