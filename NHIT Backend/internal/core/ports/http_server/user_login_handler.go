package http_server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
)

type UserLoginHandler struct {
	svc *services.UserLoginService
}

func NewUserLoginHandler(svc *services.UserLoginService) *UserLoginHandler {
	return &UserLoginHandler{svc: svc}
}

// RegisterRoutes registers user login endpoints
func (h *UserLoginHandler) RegisterRoutes(r chi.Router) {
	r.Post("/users/{userID}/login", h.RecordLogin)
	r.Get("/users/{userID}/login-history", h.ListLoginHistory)
}

// RecordLogin handles POST /users/{userID}/login
func (h *UserLoginHandler) RecordLogin(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "invalid userID", http.StatusBadRequest)
		return
	}

	var req struct {
		IpAddress string `json:"ip_address"`
		UserAgent string `json:"user_agent"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	history, err := h.svc.RecordLogin(r.Context(), userID, req.IpAddress, req.UserAgent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(history)
}

// ListLoginHistory handles GET /users/{userID}/login-history
func (h *UserLoginHandler) ListLoginHistory(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "invalid userID", http.StatusBadRequest)
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

	histories, err := h.svc.GetLoginHistory(r.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(histories)
}
