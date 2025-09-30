package http_server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
)

// RoleHandler handles role-related HTTP requests
type RoleHandler struct {
	svc  *services.RoleService
	auth *AuthMiddleware // middleware to check user and permissions
}

func NewRoleHandler(svc *services.RoleService, auth *AuthMiddleware) *RoleHandler {
	return &RoleHandler{svc: svc, auth: auth}
}

// RegisterRoutes registers role endpoints with middleware
func (h *RoleHandler) RegisterRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(h.auth.RequirePermission("create-role", "edit-role", "delete-role"))
		r.Get("/roles", h.ListRoles)
		r.Get("/roles/{roleID}", h.GetRole)
	})

	r.Group(func(r chi.Router) {
		r.Use(h.auth.RequirePermission("create-role"))
		r.Post("/roles", h.CreateRole)
	})

	r.Group(func(r chi.Router) {
		r.Use(h.auth.RequirePermission("edit-role"))
		r.Put("/roles/{roleID}", h.UpdateRole)
	})

	r.Group(func(r chi.Router) {
		r.Use(h.auth.RequirePermission("delete-role"))
		r.Delete("/roles/{roleID}", h.DeleteRole)
	})

	// Role assignment and permissions
	r.Post("/roles/{roleID}/assign-user/{userID}", h.AssignRoleToUser)
	r.Post("/roles/{roleID}/assign-permission/{permissionID}", h.AssignPermissionToRole)
	r.Get("/users/{userID}/roles", h.ListRolesOfUser)
	r.Get("/users/{userID}/permissions", h.ListPermissionsOfUser)
}

// CreateRole creates a new role
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID    string      `json:"tenant_id"`
		Name        string      `json:"name"`
		Permissions []uuid.UUID `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
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

	// Assign permissions
	for _, pid := range req.Permissions {
		h.svc.AssignPermissionToRole(r.Context(), role.RoleID, pid)
	}

	// Log activity
	h.auth.LogActivity(r.Context(), "Role Created", role.RoleID, "created")

	// Notify super admins
	h.auth.NotifySuperAdminsIfNeeded(r.Context(), role)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(role)
}

// UpdateRole updates role info
func (h *RoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	roleID, _ := uuid.Parse(chi.URLParam(r, "roleID"))

	var req struct {
		Name        string      `json:"name"`
		Permissions []uuid.UUID `json:"permissions"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	role, err := h.svc.UpdateRole(r.Context(), roleID, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Sync permissions
	for _, pid := range req.Permissions {
		h.svc.AssignPermissionToRole(r.Context(), role.RoleID, pid)
	}

	// Log activity
	h.auth.LogActivity(r.Context(), "Role Updated", role.RoleID, "updated")

	// Notify super admins
	h.auth.NotifySuperAdminsIfNeeded(r.Context(), role)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(role)
}

// DeleteRole deletes a role
func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	roleID, _ := uuid.Parse(chi.URLParam(r, "roleID"))

	role, _ := h.svc.GetRole(r.Context(), roleID)

	// Prevent deleting "Super Admin" role
	if role.Name == "Super Admin" || h.auth.IsCurrentUserInRole(role.Name, r.Context()) {
		http.Error(w, "cannot delete this role", http.StatusForbidden)
		return
	}

	if err := h.svc.DeleteRole(r.Context(), roleID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log activity
	h.auth.LogActivity(r.Context(), "Role Deleted", role.RoleID, "deleted")

	// Notify super admins
	h.auth.NotifySuperAdminsIfNeeded(r.Context(), role)

	w.WriteHeader(http.StatusNoContent)
}

// GetRole returns a role
func (h *RoleHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	roleID, _ := uuid.Parse(chi.URLParam(r, "roleID"))
	role, err := h.svc.GetRole(r.Context(), roleID)
	if err != nil {
		http.Error(w, "role not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(role)
}

// ListRoles returns roles for a tenant
func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := uuid.Parse(r.URL.Query().Get("tenant_id"))
	roles, err := h.svc.ListRoles(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(roles)
}

// AssignRoleToUser assigns a role to a user
func (h *RoleHandler) AssignRoleToUser(w http.ResponseWriter, r *http.Request) {
	roleID, _ := uuid.Parse(chi.URLParam(r, "roleID"))
	userID, _ := uuid.Parse(chi.URLParam(r, "userID"))

	if err := h.svc.AssignRoleToUser(r.Context(), userID, roleID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log activity
	h.auth.LogActivity(r.Context(), "Assigned Role To User", roleID, "assigned")

	w.WriteHeader(http.StatusNoContent)
}

// AssignPermissionToRole assigns permission to role
func (h *RoleHandler) AssignPermissionToRole(w http.ResponseWriter, r *http.Request) {
	roleID, _ := uuid.Parse(chi.URLParam(r, "roleID"))
	permissionID, _ := uuid.Parse(chi.URLParam(r, "permissionID"))

	if err := h.svc.AssignPermissionToRole(r.Context(), roleID, permissionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log activity
	h.auth.LogActivity(r.Context(), "Assigned Permission To Role", roleID, "assigned")

	w.WriteHeader(http.StatusNoContent)
}

// ListRolesOfUser lists roles of a user
func (h *RoleHandler) ListRolesOfUser(w http.ResponseWriter, r *http.Request) {
	userID, _ := uuid.Parse(chi.URLParam(r, "userID"))
	roles, _ := h.svc.ListRolesOfUser(r.Context(), userID)
	json.NewEncoder(w).Encode(roles)
}

// ListPermissionsOfUser lists permissions of a user
func (h *RoleHandler) ListPermissionsOfUser(w http.ResponseWriter, r *http.Request) {
	userID, _ := uuid.Parse(chi.URLParam(r, "userID"))
	perms, _ := h.svc.ListPermissionsOfUser(r.Context(), userID)
	json.NewEncoder(w).Encode(perms)
}