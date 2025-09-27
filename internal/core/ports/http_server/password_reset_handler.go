package http_server

import (
	"encoding/json"
	"net/http"

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

func (h *PasswordResetHandler) RegisterRoutes(r chi.Router) {
	r.Post("/password-reset", h.ForgotPassword)
	r.Post("/password-reset/{token}", h.ResetPassword)
}

// ForgotPassword endpoint
func (h *PasswordResetHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.svc.CreateToken(r.Context(), req.Email)
	if err != nil {
		// Do not reveal if email exists
		http.Error(w, "If this email is registered, you will receive a password reset link", http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password reset link sent",
		"token":   token.String(), // optional: usually sent via email
	})
}

// ResetPassword endpoint
func (h *PasswordResetHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	tokenStr := chi.URLParam(r, "token")
	token, err := uuid.Parse(tokenStr)
	if err != nil {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}

	var req struct {
		Password              string `json:"password"`
		PasswordConfirmation  string `json:"password_confirmation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Password != req.PasswordConfirmation {
		http.Error(w, "passwords do not match", http.StatusBadRequest)
		return
	}

	if err := h.svc.ResetPassword(r.Context(), token, req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "password successfully reset"})
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
