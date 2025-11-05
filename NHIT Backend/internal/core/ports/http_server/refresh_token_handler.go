package http_server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
)

type RefreshTokenHandler struct {
	svc *services.RefreshTokenService
}

func NewRefreshTokenHandler(svc *services.RefreshTokenService) *RefreshTokenHandler {
	return &RefreshTokenHandler{svc: svc}
}

// writeJSON is a helper for sending JSON responses
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
	}
}

// RegisterRoutes sets up the refresh token endpoints
func (h *RefreshTokenHandler) RegisterRoutes(r chi.Router) {
	r.Post("/refresh-token", h.CreateToken)
	r.Get("/refresh-token/{token}", h.GetUserID)
	r.Delete("/refresh-token/{token}", h.DeleteToken)
}

// CreateToken handles POST /refresh-token
func (h *RefreshTokenHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID    string `json:"user_id"`
		Token     string `json:"token"`
		ExpiresIn int64  `json:"expires_in"` // seconds until expiration
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user_id"})
		return
	}

	expiresAt := time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)

	if err := h.svc.CreateToken(r.Context(), userID, req.Token, expiresAt); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "refresh token created"})
}

// GetUserID handles GET /refresh-token/{token}
func (h *RefreshTokenHandler) GetUserID(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	userID, err := h.svc.GetUserIDByToken(r.Context(), token)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"user_id": userID.String()})
}

// DeleteToken handles DELETE /refresh-token/{token}
func (h *RefreshTokenHandler) DeleteToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if err := h.svc.DeleteToken(r.Context(), token); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
