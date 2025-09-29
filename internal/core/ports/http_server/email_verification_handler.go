package http_server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/go-chi/chi/v5"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
)

type EmailVerificationHandler struct {
	svc *services.EmailVerificationService
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func NewEmailVerificationHandler(svc *services.EmailVerificationService) *EmailVerificationHandler {
	return &EmailVerificationHandler{svc: svc}
}

// Routes registers all email verification endpoints
func (h *EmailVerificationHandler) Routes(r chi.Router) {
	r.Post("/users/{userID}/send-verification", h.SendVerification)
	r.Get("/verify-email", h.VerifyEmail)
}

// POST /users/{userID}/send-verification
func (h *EmailVerificationHandler) SendVerification(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.svc.SendVerificationEmail(r.Context(), userID, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"token": token.String()})
}

// GET /verify-email?token=...
func (h *EmailVerificationHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	token, err := uuid.Parse(tokenStr)
	if err != nil {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}

	if err := h.svc.VerifyEmail(r.Context(), token); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Email verified successfully"})
}
