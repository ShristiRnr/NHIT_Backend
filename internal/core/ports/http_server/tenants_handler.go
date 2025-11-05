package http_server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
)

type TenantHandler struct {
	svc *services.TenantService
}

func NewTenantHandler(svc *services.TenantService) *TenantHandler {
	return &TenantHandler{svc: svc}
}

// RegisterRoutes registers tenant endpoints
func (h *TenantHandler) RegisterRoutes(r chi.Router) {
	r.Post("/tenants", h.CreateTenant)
	r.Get("/tenants/{tenantID}", h.GetTenant)
}

// CreateTenant handles POST /tenants
func (h *TenantHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name            string `json:"name"`
		SuperAdminUserID string `json:"super_admin_user_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var superAdminID *uuid.UUID
	if req.SuperAdminUserID != "" {
		uid, err := uuid.Parse(req.SuperAdminUserID)
		if err != nil {
			http.Error(w, "invalid super_admin_user_id", http.StatusBadRequest)
			return
		}
		superAdminID = &uid
	}

	tenant, err := h.svc.CreateTenant(r.Context(), req.Name, superAdminID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tenant)
}

// GetTenant handles GET /tenants/{tenantID}
func (h *TenantHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "tenantID"))
	if err != nil {
		http.Error(w, "invalid tenantID", http.StatusBadRequest)
		return
	}

	tenant, err := h.svc.GetTenant(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(tenant)
}
