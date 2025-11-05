package http_server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
)

type RoleHandler struct {
	svc *services.RoleService
}

func NewRoleHandler(svc *services.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

// RegisterRoutes registers role endpoints
func (h *RoleHandler) RegisterRoutes(r chi.Router) {
	r.Post("/roles", h.CreateRole)
	r.Get("/roles/{roleID}", h.GetRole)
	r.Get("/roles", h.ListRoles) // Query param: tenant_id
	r.Put("/roles/{roleID}", h.UpdateRole)
	r.Delete("/roles/{roleID}", h.DeleteRole)

	r.Post("/roles/{roleID}/assign-user/{userID}", h.AssignRoleToUser)
	r.Post("/roles/{roleID}/assign-permission/{permissionID}", h.AssignPermissionToRole)
	r.Get("/users/{userID}/roles", h.ListRolesOfUser)
	r.Get("/users/{userID}/permissions", h.ListPermissionsOfUser)
}

// CreateRole handles POST /roles
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID string `json:"tenant_id"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		http.Error(w, "invalid tenant_id", http.StatusBadRequest)
		return
	}

	role, err := h.svc.CreateRole(r.Context(), tenantID, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(role)
}

// GetRole handles GET /roles/{roleID}
func (h *RoleHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	roleID, err := uuid.Parse(chi.URLParam(r, "roleID"))
	if err != nil {
		http.Error(w, "invalid roleID", http.StatusBadRequest)
		return
	}

	role, err := h.svc.GetRole(r.Context(), roleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(role)
}

// ListRoles handles GET /roles?tenant_id=<tenant_id>
func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.URL.Query().Get("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "invalid tenant_id", http.StatusBadRequest)
		return
	}

	roles, err := h.svc.ListRoles(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(roles)
}

// UpdateRole handles PUT /roles/{roleID}
func (h *RoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	roleID, err := uuid.Parse(chi.URLParam(r, "roleID"))
	if err != nil {
		http.Error(w, "invalid roleID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	role, err := h.svc.UpdateRole(r.Context(), roleID, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(role)
}

// DeleteRole handles DELETE /roles/{roleID}
func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	roleID, err := uuid.Parse(chi.URLParam(r, "roleID"))
	if err != nil {
		http.Error(w, "invalid roleID", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteRole(r.Context(), roleID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AssignRoleToUser handles POST /roles/{roleID}/assign-user/{userID}
func (h *RoleHandler) AssignRoleToUser(w http.ResponseWriter, r *http.Request) {
	roleID, _ := uuid.Parse(chi.URLParam(r, "roleID"))
	userID, _ := uuid.Parse(chi.URLParam(r, "userID"))

	if err := h.svc.AssignRoleToUser(r.Context(), userID, roleID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AssignPermissionToRole handles POST /roles/{roleID}/assign-permission/{permissionID}
func (h *RoleHandler) AssignPermissionToRole(w http.ResponseWriter, r *http.Request) {
	roleID, _ := uuid.Parse(chi.URLParam(r, "roleID"))
	permissionID, _ := uuid.Parse(chi.URLParam(r, "permissionID"))

	if err := h.svc.AssignPermissionToRole(r.Context(), roleID, permissionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListRolesOfUser handles GET /users/{userID}/roles
func (h *RoleHandler) ListRolesOfUser(w http.ResponseWriter, r *http.Request) {
	userID, _ := uuid.Parse(chi.URLParam(r, "userID"))

	roles, err := h.svc.ListRolesOfUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(roles)
}

// ListPermissionsOfUser handles GET /users/{userID}/permissions
func (h *RoleHandler) ListPermissionsOfUser(w http.ResponseWriter, r *http.Request) {
	userID, _ := uuid.Parse(chi.URLParam(r, "userID"))

	perms, err := h.svc.ListPermissionsOfUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(perms)
}
