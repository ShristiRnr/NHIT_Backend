package grpc

import (
	"context"

	"github.com/google/uuid"
	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserHandler struct {
	userpb.UnimplementedUserManagementServer
	userService ports.UserService
}

// NewUserHandler creates a new gRPC user handler
func NewUserHandler(userService ports.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.UserResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	user := &domain.User{
		TenantID: tenantID,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	createdUser, err := h.userService.CreateUser(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return toPBUser(createdUser), nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.UserResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	user, err := h.userService.GetUser(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return toPBUser(user), nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UserResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	user := &domain.User{
		UserID:   userID,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	updatedUser, err := h.userService.UpdateUser(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return toPBUser(updatedUser), nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	if err := h.userService.DeleteUser(ctx, userID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	// Extract pagination from PageRequest
	var limit, offset int32 = 10, 0
	if req.Page != nil {
		limit = req.Page.PageSize
		offset = (req.Page.Page - 1) * req.Page.PageSize
	}

	users, err := h.userService.ListUsersByTenant(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	pbUsers := make([]*userpb.User, len(users))
	for i, user := range users {
		pbUsers[i] = &userpb.User{
			UserId: user.UserID.String(),
			Name:   user.Name,
			Email:  user.Email,
		}
	}

	return &userpb.ListUsersResponse{Users: pbUsers}, nil
}

func (h *UserHandler) AssignRolesToUser(ctx context.Context, req *userpb.AssignRolesRequest) (*userpb.UserResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	for _, roleIDStr := range req.Roles {
		roleID, err := uuid.Parse(roleIDStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid role_id: %v", err)
		}

		if err := h.userService.AssignRoleToUser(ctx, userID, roleID); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to assign role: %v", err)
		}
	}

	user, err := h.userService.GetUser(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	roles, err := h.userService.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user roles: %v", err)
	}

	response := toPBUser(user)
	for _, role := range roles {
		response.Roles = append(response.Roles, role.Name)
		response.Permissions = append(response.Permissions, role.Permissions...)
	}

	return response, nil
}

// Helper function to convert domain user to protobuf user
func toPBUser(user *domain.User) *userpb.UserResponse {
	return &userpb.UserResponse{
		UserId: user.UserID.String(),
		Name:   user.Name,
		Email:  user.Email,
	}
}
