package http_server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
)

type PasswordResetHandler struct {
	svc *services.PasswordResetService
}

func NewPasswordResetHandler(svc *services.PasswordResetService) *PasswordResetHandler {
	return &PasswordResetHandler{svc: svc}
}

// RegisterRoutes registers the password reset endpoints
func (h *PasswordResetHandler) RegisterRoutes(r chi.Router) {
	r.Post("/password-reset", h.CreateToken)
	r.Get("/password-reset/{token}", h.GetToken)
	r.Delete("/password-reset/{token}", h.DeleteToken)
}

// CreateToken handles POST /password-reset
func (h *PasswordResetHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID    string `json:"user_id"`
		ExpiresIn int64  `json:"expires_in"` // seconds until expiration
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	token := uuid.New()
	expiresAt := time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)

	resetToken, err := h.svc.CreateToken(r.Context(), userID, token, expiresAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resetToken)
}

// GetToken handles GET /password-reset/{token}
func (h *PasswordResetHandler) GetToken(w http.ResponseWriter, r *http.Request) {
	tokenStr := chi.URLParam(r, "token")
	token, err := uuid.Parse(tokenStr)
	if err != nil {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}

	resetToken, err := h.svc.GetToken(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resetToken)
}

// DeleteToken handles DELETE /password-reset/{token}
func (h *PasswordResetHandler) DeleteToken(w http.ResponseWriter, r *http.Request) {
	tokenStr := chi.URLParam(r, "token")
	token, err := uuid.Parse(tokenStr)
	if err != nil {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteToken(r.Context(), token); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
