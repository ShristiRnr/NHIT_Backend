package http_server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
)

// OrganizationHandler handles HTTP requests for organizations
type OrganizationHandler struct {
	svc *services.OrganizationService
}

// NewOrganizationHandler creates a new OrganizationHandler
func NewOrganizationHandler(svc *services.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{svc: svc}
}

// RegisterRoutes registers organization routes to the router
func (h *OrganizationHandler) RegisterRoutes(r chi.Router) {
	r.Post("/organizations", h.Create)
	r.Get("/organizations/{id}", h.Get)
	r.Get("/tenants/{tenantID}/organizations", h.List)
	r.Put("/organizations/{id}", h.Update)
	r.Delete("/organizations/{id}", h.Delete)
}

// Create organization
func (h *OrganizationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID string `json:"tenant_id"`
		Name     string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		http.Error(w, "invalid tenant_id", http.StatusBadRequest)
		return
	}

	org, err := h.svc.CreateOrganization(r.Context(), tenantID, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(org)
}

// Get organization by ID
func (h *OrganizationHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid org id", http.StatusBadRequest)
		return
	}

	org, err := h.svc.GetOrganization(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(org)
}

// List organizations by tenant
func (h *OrganizationHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "tenantID")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "invalid tenant id", http.StatusBadRequest)
		return
	}

	orgs, err := h.svc.ListOrganizations(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(orgs)
}

// Update organization
func (h *OrganizationHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid org id", http.StatusBadRequest)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	org, err := h.svc.UpdateOrganization(r.Context(), orgID, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(org)
}

// Delete organization
func (h *OrganizationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid org id", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteOrganization(r.Context(), orgID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
