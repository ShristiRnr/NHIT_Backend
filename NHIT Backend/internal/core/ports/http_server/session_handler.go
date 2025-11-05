package http_server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
)

type SessionHandler struct {
	svc *services.SessionService
}

func NewSessionHandler(svc *services.SessionService) *SessionHandler {
	return &SessionHandler{svc: svc}
}

// RegisterRoutes registers session endpoints
func (h *SessionHandler) RegisterRoutes(r chi.Router) {
	r.Post("/sessions", h.CreateSession)
	r.Get("/sessions/{sessionID}", h.GetSession)
	r.Delete("/sessions/{sessionID}", h.DeleteSession)
}

// CreateSession handles POST /sessions
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID    string `json:"user_id"`
		Token     string `json:"token"`
		ExpiresAt string `json:"expires_at"` // ISO8601 format
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

	expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil {
		http.Error(w, "invalid expires_at format", http.StatusBadRequest)
		return
	}

	session, err := h.svc.CreateSession(r.Context(), userID, req.Token, expiresAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(session)
}

// GetSession handles GET /sessions/{sessionID}
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionID"))
	if err != nil {
		http.Error(w, "invalid sessionID", http.StatusBadRequest)
		return
	}

	session, err := h.svc.GetSession(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(session)
}

// DeleteSession handles DELETE /sessions/{sessionID}
func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionID"))
	if err != nil {
		http.Error(w, "invalid sessionID", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteSession(r.Context(), sessionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
