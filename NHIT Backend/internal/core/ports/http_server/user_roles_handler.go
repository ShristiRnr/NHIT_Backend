package http_server

import (
	"context"

	"github.com/google/uuid"
	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
)

// UserRoleHandler implements role-related gRPC methods.
type UserRoleHandler struct {
	userpb.UnimplementedUserManagementServer
	svc *services.UserRoleService
}

// NewUserRoleHandler constructs a new handler.
func NewUserRoleHandler(svc *services.UserRoleService) *UserRoleHandler {
	return &UserRoleHandler{svc: svc}
}

// AssignRolesToUser assigns roles to a user.
func (h *UserRoleHandler) AssignRolesToUser(ctx context.Context, req *userpb.AssignRolesRequest) (*userpb.UserResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	// Convert role strings → UUIDs
	var roleIDs []uuid.UUID
	for _, r := range req.Roles {
		rid, err := uuid.Parse(r)
		if err != nil {
			return nil, err
		}
		roleIDs = append(roleIDs, rid)
	}

	if err := h.svc.AssignRoles(ctx, userID, roleIDs); err != nil {
		return nil, err
	}

	// Fetch roles back
	roles, err := h.svc.ListRolesOfUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var roleNames []string
	for _, r := range roles {
		roleNames = append(roleNames, r.Name)
	}

	return &userpb.UserResponse{
		UserId: req.UserId,
		Roles:  roleNames,
	}, nil
}

// ListRolesOfUser returns roles and permissions of a user.
func (h *UserRoleHandler) ListRolesOfUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.ListRolesResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	roles, err := h.svc.ListRolesOfUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var pbRoles []*userpb.RoleResponse
	for _, r := range roles {
		pbRoles = append(pbRoles, &userpb.RoleResponse{
			RoleId: r.RoleID.String(),
			Name:   r.Name,
		})
	}

	return &userpb.ListRolesResponse{Roles: pbRoles}, nil
}
