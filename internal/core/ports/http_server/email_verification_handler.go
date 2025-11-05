package http_server

import (
	"encoding/json"
	"net/http"
	"time"
	"fmt"

	"github.com/google/uuid"
	"github.com/go-chi/chi/v5"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
	"github.com/ShristiRnr/NHIT_Backend/helpers"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/email"
)

type EmailVerificationHandler struct {
	svc    *services.EmailVerificationService
	sender *email.SMTPTLSender
}

func NewEmailVerificationHandler(svc *services.EmailVerificationService, sender *email.SMTPTLSender) *EmailVerificationHandler {
	return &EmailVerificationHandler{
		svc:    svc,
		sender: sender,
	}
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
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Generate verification token
	token, err := h.svc.SendVerificationEmail(r.Context(), userID, req.Email)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Construct verification link
	link := "https://yourfrontend.com/verify-email?token=" + token.String()
	expiry := time.Now().Add(24 * time.Hour) // or get from service

	// Send verification email asynchronously
	h.sender.SendAsync(r.Context(), req.Email, 
		fmt.Sprintf("[%s] Verify Your Email", h.sender.AppName),
		fmt.Sprintf("Hello,<br><br>Please verify your email by clicking the link below:<br><a href=\"%s\">Verify Email</a><br><br>This link expires at %s.<br><br>Thanks,<br>%s Team", 
			link, expiry.Format(time.RFC1123), h.sender.AppName), "text/html")

	helpers.WriteJSON(w, http.StatusCreated, map[string]string{"message": "Verification email sent"})
}

// GET /verify-email?token=...
func (h *EmailVerificationHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	token, err := uuid.Parse(tokenStr)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid token"})
		return
	}

	if err := h.svc.VerifyEmail(r.Context(), token); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, map[string]string{"message": "Email verified successfully"})
}
