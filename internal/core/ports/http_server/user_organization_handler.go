package http_server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
)

type UserOrganizationHandler struct {
	svc *services.UserOrganizationService
}

func NewUserOrganizationHandler(svc *services.UserOrganizationService) *UserOrganizationHandler {
	return &UserOrganizationHandler{svc: svc}
}

// RegisterRoutes registers organization-related endpoints
func (h *UserOrganizationHandler) RegisterRoutes(r chi.Router) {
	r.Post("/organizations/{orgID}/users", h.AddUserToOrganization)
	r.Get("/organizations/{orgID}/users", h.ListUsersByOrganization)
}

// AddUserToOrganization handles POST /organizations/{orgID}/users
func (h *UserOrganizationHandler) AddUserToOrganization(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgID"))
	if err != nil {
		http.Error(w, "invalid orgID", http.StatusBadRequest)
		return
	}

	var req struct {
		UserID uuid.UUID `json:"user_id"`
		RoleID uuid.UUID `json:"role_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.AddUserToOrganization(r.Context(), req.UserID, orgID, req.RoleID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// ListUsersByOrganization handles GET /organizations/{orgID}/users
func (h *UserOrganizationHandler) ListUsersByOrganization(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgID"))
	if err != nil {
		http.Error(w, "invalid orgID", http.StatusBadRequest)
		return
	}

	users, err := h.svc.ListUsersByOrganization(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}
