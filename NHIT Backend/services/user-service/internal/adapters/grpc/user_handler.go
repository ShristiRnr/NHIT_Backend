package grpc

import (
	"context"
	"strings"
	"time"

	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	deptpb "github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb"
	desigpb "github.com/ShristiRnr/NHIT_Backend/api/pb/designationpb"
	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHandler struct {
	userpb.UnimplementedUserManagementServer
	userService ports.UserService
	db          *pgxpool.Pool
	authClient  authpb.AuthServiceClient
	deptConn    *grpc.ClientConn
	desigConn   *grpc.ClientConn
}

// NewUserHandler creates a new gRPC user handler
func NewUserHandler(userService ports.UserService, db *pgxpool.Pool, authClient authpb.AuthServiceClient, deptConn *grpc.ClientConn, desigConn *grpc.ClientConn) *UserHandler {
	return &UserHandler{
		userService: userService,
		db:          db,
		authClient:  authClient,
		deptConn:    deptConn,
		desigConn:   desigConn,
	}
}

type authContext struct {
	token *authpb.ValidateTokenResponse
}

// helper to get first non-empty metadata value by keys
func firstMetadataValue(md metadata.MD, keys ...string) string {
	for _, k := range keys {
		if vals := md[strings.ToLower(k)]; len(vals) > 0 && vals[0] != "" {
			return vals[0]
		}
	}
	return ""
}

func (h *UserHandler) requireAuthWithPermissions(ctx context.Context, requiredPerms ...string) (*authContext, error) {
	if h.authClient == nil {
		return nil, status.Error(codes.Internal, "auth client not configured for user handler")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	accessToken := firstMetadataValue(md, "authorization")
	if accessToken == "" {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}
	if strings.HasPrefix(accessToken, "Bearer ") {
		accessToken = strings.TrimPrefix(accessToken, "Bearer ")
	}

	vResp, err := h.authClient.ValidateToken(ctx, &authpb.ValidateTokenRequest{Token: accessToken})
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to validate token: %v", err)
	}
	if !vResp.Valid {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
	}

	isSuperAdmin := false
	for _, r := range vResp.Roles {
		if r == "SUPER_ADMIN" {
			isSuperAdmin = true
			break
		}
	}
	if isSuperAdmin || len(requiredPerms) == 0 {
		return &authContext{token: vResp}, nil
	}

	permSet := make(map[string]struct{}, len(vResp.Permissions))
	for _, p := range vResp.Permissions {
		permSet[p] = struct{}{}
	}
	for _, req := range requiredPerms {
		if _, ok := permSet[req]; ok {
			return &authContext{token: vResp}, nil
		}
	}

	return nil, status.Error(codes.PermissionDenied, "insufficient permissions")
}

// helper to build a simple role description based on permission keys
func buildRoleDescription(perms []string) string {
	if len(perms) == 0 {
		return ""
	}
	return "Permissions: " + strings.Join(perms, ", ")
}

func toPBRole(role *domain.Role) *userpb.RoleResponse {
	return &userpb.RoleResponse{
		RoleId:      role.RoleID.String(),
		TenantId:    role.TenantID.String(),
		Name:        role.Name,
		Permissions: role.Permissions,
	}
}

func toPBPermission(p *domain.Permission) *userpb.PermissionResponse {
	return &userpb.PermissionResponse{
		PermissionId:       p.PermissionID.String(),
		Name:               p.Name,
		Description:        p.Description,
		Module:             p.Module,
		Action:             p.Action,
		IsSystemPermission: p.IsSystemPermission,
	}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.UserResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	// Parse and VALIDATE mandatory department_id
	var departmentID *uuid.UUID
	if req.DepartmentId == "" {
		return nil, status.Error(codes.InvalidArgument, "department_id is required")
	}
	deptID, err := uuid.Parse(req.DepartmentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid department_id: %v", err)
	}
	departmentID = &deptID

	// Parse and VALIDATE mandatory designation_id
	var designationID *uuid.UUID
	if req.DesignationId == "" {
		return nil, status.Error(codes.InvalidArgument, "designation_id is required")
	}
	desigID, err := uuid.Parse(req.DesignationId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid designation_id: %v", err)
	}
	designationID = &desigID

	// VALIDATE mandatory role_id
	if req.RoleId == "" {
		return nil, status.Error(codes.InvalidArgument, "role_id is required")
	}
	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid role_id: %v", err)
	}

	// Parse banking information
	var accountHolderName, bankName, bankAccountNumber, ifscCode *string
	if req.AccountHolderName != "" {
		accountHolderName = &req.AccountHolderName
	}
	if req.BankName != "" {
		bankName = &req.BankName
	}
	if req.BankAccountNumber != "" {
		bankAccountNumber = &req.BankAccountNumber
	}
	if req.IfscCode != "" {
		ifscCode = &req.IfscCode
	}

	user := &domain.User{
		TenantID:          tenantID,
		Name:              req.Name,
		Email:             req.Email,
		Password:          req.Password,
		DepartmentID:      departmentID,
		DesignationID:     designationID,
		AccountHolderName: accountHolderName,
		BankName:          bankName,
		BankAccountNumber: bankAccountNumber,
		IFSCCode:          ifscCode,
	}

	createdUser, err := h.userService.CreateUser(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	// Assign VALIDATED role to user
	if err := h.userService.AssignRoleToUser(ctx, createdUser.UserID, roleID); err != nil {
		// Note: User is created but role assignment failed. Ideally this should be a transaction.
		// For now, we return error so client knows something went wrong.
		return nil, status.Errorf(codes.Internal, "failed to assign role to user: %v", err)
	}

	// Fetch roles/permissions to return in response
	var roleNames []string
	var permissions []string
	
	roles, err := h.userService.GetUserRoles(ctx, createdUser.UserID)
	if err == nil {
		permMap := make(map[string]bool)
		for _, r := range roles {
			roleNames = append(roleNames, r.Name)
			for _, p := range r.Permissions {
				if !permMap[p] {
					permissions = append(permissions, p)
					permMap[p] = true
				}
			}
		}
	}

	return domainUserToProto(createdUser, roleNames, permissions), nil
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

	// Fetch roles/permissions
	roles, err := h.userService.GetUserRoles(ctx, user.UserID)
	var roleNames []string
	var permissions []string
	if err == nil {
		permMap := make(map[string]bool)
		for _, r := range roles {
			roleNames = append(roleNames, r.Name)
			for _, p := range r.Permissions {
				if !permMap[p] {
					permissions = append(permissions, p)
					permMap[p] = true
				}
			}
		}
	}

	return domainUserToProto(user, roleNames, permissions), nil
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

	// Handle role updates if provided
	if len(req.Roles) > 0 {
		// First, list current roles to see if we need to remove or add
		// For simplicity given the requirement "update user roles", we will assign these roles.
		// Since AssignRolesToUser is additive in our service usually, handling "replace" might require a "RemoveAllRoles" 
		// or we assume the UI sends the desired final state and we might need to sync.
		// However, looking at standard implementation, we often iterate and assign.
		// NOTE: Ideally we should probably clear old roles first if this is a "set roles" operation.
		// Let's assume for now we Just Assign. If user wants to remove, they use RemoveUserFromOrganization or distinct API.
		// But wait, the UpdateUser API has `roles` field. If verified UI sends the full list, we should probably ensure consistency.
		// Given current constraints and "AssignRoleToUser" availability, let's just loop and assign.

		for _, roleIDStr := range req.Roles {
			roleID, err := uuid.Parse(roleIDStr)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid role_id: %v", err)
			}
			if err := h.userService.AssignRoleToUser(ctx, userID, roleID); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to assign role: %v", err)
			}
		}
	}

	updatedUser, err := h.userService.UpdateUser(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	// Fetch roles/permissions for response
	roles, err := h.userService.GetUserRoles(ctx, updatedUser.UserID)
	var roleNames []string
	var permissions []string
	if err == nil {
		permMap := make(map[string]bool)
		for _, r := range roles {
			roleNames = append(roleNames, r.Name)
			for _, p := range r.Permissions {
				if !permMap[p] {
					permissions = append(permissions, p)
					permMap[p] = true
				}
			}
		}
	}

	return domainUserToProto(updatedUser, roleNames, permissions), nil
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

// DeactivateUser deactivates a user (soft delete)
func (h *UserHandler) DeactivateUser(ctx context.Context, req *userpb.DeactivateUserRequest) (*userpb.UserResponse, error) {
	// Parse user_id to deactivate
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	// Extract deactivated_by from context (the logged-in admin)
	deactivatedByStr, ok := ctx.Value("user_id").(string)
	if !ok || deactivatedByStr == "" {
		return nil, status.Error(codes.Unauthenticated, "user_id not found in context")
	}

	deactivatedBy, err := uuid.Parse(deactivatedByStr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invalid user_id in context: %v", err)
	}

	// Call service to deactivate user
	deactivatedUser, err := h.userService.DeactivateUser(ctx, userID, deactivatedBy, req.Reason)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to deactivate user: %v", err)
	}

	// Fetch keys for response
	roles, err := h.userService.GetUserRoles(ctx, deactivatedUser.UserID)
	var roleNames []string
	var permissions []string
	if err == nil {
		permMap := make(map[string]bool)
		for _, r := range roles {
			roleNames = append(roleNames, r.Name)
			for _, p := range r.Permissions {
				if !permMap[p] {
					permissions = append(permissions, p)
					permMap[p] = true
				}
			}
		}
	}

	return domainUserToProto(deactivatedUser, roleNames, permissions), nil
}

// ReactivateUser reactivates a previously deactivated user
func (h *UserHandler) ReactivateUser(ctx context.Context, req *userpb.ReactivateUserRequest) (*userpb.UserResponse, error) {
	// Parse user_id to reactivate
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	// Extract reactivated_by from context (the logged-in admin)
	reactivatedByStr, ok := ctx.Value("user_id").(string)
	if !ok || reactivatedByStr == "" {
		return nil, status.Error(codes.Unauthenticated, "user_id not found in context")
	}

	reactivatedBy, err := uuid.Parse(reactivatedByStr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invalid user_id in context: %v", err)
	}

	// Call service to reactivate user
	reactivatedUser, err := h.userService.ReactivateUser(ctx, userID, reactivatedBy)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reactivate user: %v", err)
	}

	// Fetch keys for response
	roles, err := h.userService.GetUserRoles(ctx, reactivatedUser.UserID)
	var roleNames []string
	var permissions []string
	if err == nil {
		permMap := make(map[string]bool)
		for _, r := range roles {
			roleNames = append(roleNames, r.Name)
			for _, p := range r.Permissions {
				if !permMap[p] {
					permissions = append(permissions, p)
					permMap[p] = true
				}
			}
		}
	}

	return domainUserToProto(reactivatedUser, roleNames, permissions), nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	// Extract tenant_id from JWT metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	tenantIDs := md.Get("tenant_id")
	if len(tenantIDs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "tenant_id not found in token")
	}

	tenantID, err := uuid.Parse(tenantIDs[0])
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id in token: %v", err)
	}

	// Extract pagination from flat fields
	var limit int32 = 10
	if req.PageSize > 0 {
		limit = req.PageSize
	}
	
	var offset int32 = 0
	if req.Page > 0 {
		offset = (req.Page - 1) * limit
	}

	users, err := h.userService.ListUsersByTenant(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	pbUsers := make([]*userpb.User, len(users))
	for i, user := range users {
		// Fetch Roles
		roles, err := h.userService.GetUserRoles(ctx, user.UserID)
		var roleNames []string
		var permissions []string
		permMap := make(map[string]bool)

		if err == nil {
			for _, r := range roles {
				roleNames = append(roleNames, r.Name)
				for _, p := range r.Permissions {
					if !permMap[p] {
						permissions = append(permissions, p)
						permMap[p] = true
					}
				}
			}
		}


		// Fetch Department Name
		var deptName string
		if user.DepartmentID != nil {
			log.Printf("Fetching department for user %s, deptID: %s", user.Name, user.DepartmentID.String())
			client := deptpb.NewDepartmentServiceClient(h.deptConn)
			// Forward metadata for auth
			outCtx := metadata.NewOutgoingContext(ctx, md)
			resp, err := client.GetDepartment(outCtx, &deptpb.GetDepartmentRequest{Id: user.DepartmentID.String()})
			if err != nil {
				log.Printf("Failed to get department for user %s: %v", user.Name, err)
			} else if resp.Department != nil {
				deptName = resp.Department.Name
				log.Printf("Found department: %s", deptName)
			}
		} else {
			log.Printf("User %s has no department ID", user.Name)
		}

		// Fetch Designation Name
		var desigName string
		if user.DesignationID != nil {
			log.Printf("Fetching designation for user %s, desigID: %s", user.Name, user.DesignationID.String())
			client := desigpb.NewDesignationServiceClient(h.desigConn)
			// Forward metadata for auth
			outCtx := metadata.NewOutgoingContext(ctx, md)
			resp, err := client.GetDesignation(outCtx, &desigpb.GetDesignationRequest{Id: user.DesignationID.String()})
			if err != nil {
				log.Printf("Failed to get designation for user %s: %v", user.Name, err)
			} else if resp.Designation != nil {
				desigName = resp.Designation.Name
				log.Printf("Found designation: %s", desigName)
			}
		} else {
			log.Printf("User %s has no designation ID", user.Name)
		}

		pbUsers[i] = toPBUserMessage(user, deptName, desigName, roleNames, permissions)
	}

	return &userpb.ListUsersResponse{
		Users: pbUsers,
		Pagination: &userpb.PageResponse{
			Page:     req.Page,
			PageSize: limit,
			// Total counts would require a separate count query
		},
	}, nil
}

// ===== Role & Permission Management =====

func (h *UserHandler) CreateRole(ctx context.Context, req *userpb.CreateRoleRequest) (*userpb.RoleResponse, error) {
	authCtx, err := h.requireAuthWithPermissions(ctx, "create-role")
	if err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(authCtx.token.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invalid tenant_id in token: %v", err)
	}

	// Determine organization context: prefer explicit header (current dashboard), fallback to token org_id
	var orgID *uuid.UUID
	md, _ := metadata.FromIncomingContext(ctx)
	if orgHeader := firstMetadataValue(md, "x-org-id", "org-id"); orgHeader != "" {
		id, err := uuid.Parse(orgHeader)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid org_id header: %v", err)
		}
		orgID = &id
	} else if authCtx.token.OrgId != "" {
		id, err := uuid.Parse(authCtx.token.OrgId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "invalid org_id in token: %v", err)
		}
		orgID = &id
	}

	desc := buildRoleDescription(req.Permissions)

	role := &domain.Role{
		TenantID:    tenantID,
		OrgID:       orgID,
		Name:        req.Name,
		Description: desc,
		Permissions: req.Permissions,
		CreatedBy:   authCtx.token.Name,
	}

	created, err := h.userService.CreateRole(ctx, role)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create role: %v", err)
	}

	return toPBRole(created), nil
}

func (h *UserHandler) GetRole(ctx context.Context, req *userpb.GetRoleRequest) (*userpb.RoleResponse, error) {
	if _, err := h.requireAuthWithPermissions(ctx, "view-role"); err != nil {
		return nil, err
	}

	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid role_id: %v", err)
	}

	role, err := h.userService.GetRole(ctx, roleID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "role not found: %v", err)
	}

	return toPBRole(role), nil
}

func (h *UserHandler) ListRoles(ctx context.Context, req *userpb.ListRolesRequest) (*userpb.ListRolesResponse, error) {
	authCtx, err := h.requireAuthWithPermissions(ctx, "view-role")
	if err != nil {
		return nil, err
	}

	if authCtx.token.TenantId == "" {
		return nil, status.Error(codes.Internal, "tenant_id missing in auth token")
	}

	tenantID, err := uuid.Parse(authCtx.token.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invalid tenant_id in token: %v", err)
	}

	roles, err := h.userService.ListRolesByTenant(ctx, tenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list roles: %v", err)
	}

	resp := &userpb.ListRolesResponse{Roles: make([]*userpb.RoleResponse, len(roles))}
	for i, r := range roles {
		resp.Roles[i] = toPBRole(r)
	}
	return resp, nil
}

func (h *UserHandler) ListRolesByOrganization(ctx context.Context, req *userpb.ListRolesByOrganizationRequest) (*userpb.ListRolesResponse, error) {
	if _, err := h.requireAuthWithPermissions(ctx, "view-role"); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	var orgID *uuid.UUID
	if req.OrgId != "" {
		id, err := uuid.Parse(req.OrgId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid org_id: %v", err)
		}
		orgID = &id
	}

	roles, err := h.userService.ListRolesByOrganization(ctx, tenantID, orgID, req.IncludeSystemRoles)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list roles: %v", err)
	}

	resp := &userpb.ListRolesResponse{Roles: make([]*userpb.RoleResponse, len(roles))}
	for i, r := range roles {
		resp.Roles[i] = toPBRole(r)
	}
	return resp, nil
}

func (h *UserHandler) UpdateRole(ctx context.Context, req *userpb.UpdateRoleRequest) (*userpb.RoleResponse, error) {
	if _, err := h.requireAuthWithPermissions(ctx, "edit-role"); err != nil {
		return nil, err
	}

	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid role_id: %v", err)
	}

	existing, err := h.userService.GetRole(ctx, roleID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "role not found: %v", err)
	}

	existing.Name = req.Name
	existing.Permissions = req.Permissions
	existing.Description = buildRoleDescription(req.Permissions)

	updated, err := h.userService.UpdateRole(ctx, existing)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update role: %v", err)
	}

	return toPBRole(updated), nil
}

func (h *UserHandler) DeleteRole(ctx context.Context, req *userpb.DeleteRoleRequest) (*emptypb.Empty, error) {
	if _, err := h.requireAuthWithPermissions(ctx, "delete-role"); err != nil {
		return nil, err
	}

	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid role_id: %v", err)
	}

	if err := h.userService.DeleteRole(ctx, roleID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete role: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (h *UserHandler) ListPermissions(ctx context.Context, req *userpb.ListPermissionsRequest) (*userpb.ListPermissionsResponse, error) {
	// Only require a valid authenticated token; no specific permission is needed
	if _, err := h.requireAuthWithPermissions(ctx); err != nil {
		return nil, err
	}

	var module *string
	if req.Module != "" {
		m := req.Module
		module = &m
	}

	perms, err := h.userService.ListPermissions(ctx, module)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list permissions: %v", err)
	}

	resp := &userpb.ListPermissionsResponse{Permissions: make([]*userpb.PermissionResponse, len(perms))}
	for i, p := range perms {
		resp.Permissions[i] = toPBPermission(p)
	}
	return resp, nil
}

func (h *UserHandler) GetPermissionsByModule(ctx context.Context, req *userpb.GetPermissionsByModuleRequest) (*userpb.ListPermissionsResponse, error) {
	// Only require authentication for fetching permission catalog by module
	if _, err := h.requireAuthWithPermissions(ctx); err != nil {
		return nil, err
	}

	module := req.Module
	perms, err := h.userService.ListPermissions(ctx, &module)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list permissions by module: %v", err)
	}

	resp := &userpb.ListPermissionsResponse{Permissions: make([]*userpb.PermissionResponse, len(perms))}
	for i, p := range perms {
		resp.Permissions[i] = toPBPermission(p)
	}
	return resp, nil
}

// CreateCustomPermission is not supported in this service (fixed permission catalog)
func (h *UserHandler) CreateCustomPermission(ctx context.Context, req *userpb.CreateCustomPermissionRequest) (*userpb.PermissionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "custom permissions are not supported; use fixed catalog")
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



	response := domainUserToProto(user, nil, nil) // Base structure
	
	// Fetch fresh roles to ensure accuracy
	freshRoles, err := h.userService.GetUserRoles(ctx, userID)
	if err == nil {
		var roleNames []string
		var permissions []string
		permMap := make(map[string]bool)
		for _, r := range freshRoles {
			roleNames = append(roleNames, r.Name)
			for _, p := range r.Permissions {
				if !permMap[p] {
					permissions = append(permissions, p)
					permMap[p] = true
				}
			}
		}
		response.Roles = roleNames
		response.Permissions = permissions
	}

	return response, nil
}

func (h *UserHandler) ListRolesOfUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.ListRolesResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	roles, err := h.userService.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user roles: %v", err)
	}

	pbRoles := make([]*userpb.RoleResponse, len(roles))
	for i, role := range roles {
		pbRoles[i] = &userpb.RoleResponse{
			RoleId:      role.RoleID.String(),
			TenantId:    role.TenantID.String(),
			Name:        role.Name,
			Permissions: role.Permissions,
		}
	}

	return &userpb.ListRolesResponse{Roles: pbRoles}, nil
}

// CreateTenant creates a new tenant with super admin
func (h *UserHandler) CreateTenant(ctx context.Context, req *userpb.CreateTenantRequest) (*userpb.TenantResponse, error) {
	// Create tenant and super admin
	tenant, err := h.userService.CreateTenant(ctx, req.Name, req.Email, req.Password, "SUPER_ADMIN")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create tenant: %v", err)
	}

	return &userpb.TenantResponse{
		TenantId: tenant.TenantID.String(),
		Name:     tenant.Name,
		Email:    tenant.Email,
		Password: "", // Never expose hashed password
	}, nil
}

// GetTenant retrieves tenant information
func (h *UserHandler) GetTenant(ctx context.Context, req *userpb.GetTenantRequest) (*userpb.TenantResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	tenant, err := h.userService.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "tenant not found: %v", err)
	}

	return &userpb.TenantResponse{
		TenantId: tenant.TenantID.String(),
		Name:     tenant.Name,
		Email:    tenant.Email,
		Password: "",
	}, nil
}

// ListUserOrganizations returns organizations linked to a user.
// It also lazily creates a mapping for super admin users by
// linking any organizations where organizations.super_admin_email
// matches the user's email, so that super-admin logins can resolve orgId.
func (h *UserHandler) ListUserOrganizations(ctx context.Context, req *userpb.ListUserOrganizationsRequest) (*userpb.ListUserOrganizationsResponse, error) {
	if h.db == nil {
		return nil, status.Error(codes.Internal, "database not configured for user organizations")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	// Helper to fetch existing mappings
	fetch := func() ([]*userpb.UserOrganizationInfo, error) {
		rows, err := h.db.Query(ctx, `
			SELECT uo.org_id, o.name, uo.is_current_context, uo.joined_at
			FROM user_organizations uo
			JOIN organizations o ON o.org_id = uo.org_id
			WHERE uo.user_id = $1
			ORDER BY uo.joined_at ASC`, userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var result []*userpb.UserOrganizationInfo
		for rows.Next() {
			var (
				orgID     uuid.UUID
				orgName   string
				isCurrent bool
				joinedAt  time.Time
			)
			if err := rows.Scan(&orgID, &orgName, &isCurrent, &joinedAt); err != nil {
				return nil, err
			}
			result = append(result, &userpb.UserOrganizationInfo{
				OrgId:            orgID.String(),
				OrgName:          orgName,
				RoleName:         "",
				DepartmentName:   "",
				DesignationName:  "",
				ProjectNames:     nil,
				IsCurrentContext: isCurrent,
				JoinedAt:         timestamppb.New(joinedAt),
			})
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return result, nil
	}

	orgInfos, err := fetch()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list user organizations: %v", err)
	}
	if len(orgInfos) > 0 {
		return &userpb.ListUserOrganizationsResponse{Organizations: orgInfos}, nil
	}

	// No existing mapping: lazily link super admin user to organizations
	// whose super_admin_email matches this user's email.
	var email string
	if err := h.db.QueryRow(ctx, `SELECT email FROM users WHERE user_id = $1`, userID).Scan(&email); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user for organization lookup: %v", err)
	}

	orgRows, err := h.db.Query(ctx, `
		SELECT org_id, name
		FROM organizations
		WHERE super_admin_email = $1
		ORDER BY created_at ASC`, email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find organizations for user: %v", err)
	}
	defer orgRows.Close()

	type orgRecord struct {
		id   uuid.UUID
		name string
	}
	var orgs []orgRecord
	for orgRows.Next() {
		var (
			orgID uuid.UUID
			name  string
		)
		if err := orgRows.Scan(&orgID, &name); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan organization: %v", err)
		}
		orgs = append(orgs, orgRecord{id: orgID, name: name})
	}
	if err := orgRows.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to iterate organizations: %v", err)
	}

	if len(orgs) == 0 {
		// No organizations associated with this user yet
		return &userpb.ListUserOrganizationsResponse{Organizations: nil}, nil
	}

	now := time.Now()
	for i, o := range orgs {
		isCurrent := i == 0
		// role_id is required but there is no role table in this service; use zero UUID as placeholder
		_, err := h.db.Exec(ctx, `
			INSERT INTO user_organizations (user_id, org_id, role_id, is_current_context, joined_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (user_id, org_id) DO NOTHING`,
			userID, o.id, uuid.Nil, isCurrent, now, now,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to link user to organization: %v", err)
		}
	}

	// Fetch again with the newly created mappings
	orgInfos, err = fetch()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list user organizations after linking: %v", err)
	}

	return &userpb.ListUserOrganizationsResponse{Organizations: orgInfos}, nil
}

// CreateUserLoginHistory creates a login history entry
func (h *UserHandler) CreateUserLoginHistory(ctx context.Context, req *userpb.CreateUserLoginHistoryRequest) (*userpb.UserLoginHistoryResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	history := &domain.UserLoginHistory{
		UserID:    userID,
		IPAddress: &req.IpAddress,
		UserAgent: &req.UserAgent,
		LoginTime: time.Now(),
	}

	created, err := h.userService.CreateLoginHistory(ctx, history)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create login history: %v", err)
	}

	return &userpb.UserLoginHistoryResponse{
		HistoryId: created.HistoryID.String(),
		UserId:    created.UserID.String(),
		IpAddress: *created.IPAddress,
		UserAgent: *created.UserAgent,
		LoginTime: timestamppb.New(created.LoginTime),
	}, nil
}

// ListUserLoginHistories lists login history for a user
func (h *UserHandler) ListUserLoginHistories(ctx context.Context, req *userpb.ListUserLoginHistoriesRequest) (*userpb.ListUserLoginHistoriesResponse, error) {
	var userID uuid.UUID
	var err error

	// If user_id not provided, extract from JWT metadata
	if req.UserId == "" {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		userIDs := md.Get("user_id")
		if len(userIDs) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "missing user_id in token")
		}

		userID, err = uuid.Parse(userIDs[0])
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid user_id in token: %v", err)
		}
	} else {
		userID, err = uuid.Parse(req.UserId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
		}
	}

	var limit, offset int32 = 10, 0
	if req.Page != nil {
		limit = req.Page.PageSize
		offset = (req.Page.Page - 1) * req.Page.PageSize
	}

	histories, err := h.userService.ListLoginHistory(ctx, userID, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list login histories: %v", err)
	}

	pbHistories := make([]*userpb.UserLoginHistoryResponse, len(histories))
	for i, hist := range histories {
		pbHistories[i] = &userpb.UserLoginHistoryResponse{
			HistoryId: hist.HistoryID.String(),
			UserId:    hist.UserID.String(),
			IpAddress: *hist.IPAddress,
			UserAgent: *hist.UserAgent,
			LoginTime: timestamppb.New(hist.LoginTime),
		}
	}

	return &userpb.ListUserLoginHistoriesResponse{
		Histories: pbHistories,
	}, nil
}

// Helper function to convert domain user to protobuf user with full details
func domainUserToProto(user *domain.User, roleNames []string, permissions []string) *userpb.UserResponse {
	resp := &userpb.UserResponse{
		UserId:      user.UserID.String(),
		Name:        user.Name,
		Email:       user.Email,
		Roles:       roleNames,
		Permissions: permissions,
	}
	return resp
}

// Helper to convert domain user to PB User message (for ListUsers)
func toPBUserMessage(user *domain.User, deptName, desigName string, roleNames, permissions []string) *userpb.User {
	return &userpb.User{
		UserId:          user.UserID.String(),
		Name:            user.Name,
		Email:           user.Email,
		DepartmentName:  deptName,
		DesignationName: desigName,
		Roles:           roleNames,
		Permissions:     permissions,
		IsActive:          user.IsActive,
		DeactivatedAt:     timestamppb.New(safeTime(user.DeactivatedAt)),
		DeactivatedBy:     safeUUIDStr(user.DeactivatedBy),
		DeactivatedByName: stringPtrValue(user.DeactivatedByName),
		CreatedAt:         timestamppb.New(user.CreatedAt),
		UpdatedAt:         timestamppb.New(user.UpdatedAt),
		EmailVerifiedAt:   timestamppb.New(safeTime(user.EmailVerifiedAt)),
		LastLoginAt:       timestamppb.New(safeTime(user.LastLoginAt)),
		LastLogoutAt:      timestamppb.New(safeTime(user.LastLogoutAt)),
		LastLoginIp:       user.LastLoginIP,
		UserAgent:         user.UserAgent,
	}
}

func safeTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func safeUUIDStr(u *uuid.UUID) string {
	if u == nil {
		return ""
	}
	return u.String()
}

// CreateActivityLog creates a new activity log entry
func (h *UserHandler) CreateActivityLog(ctx context.Context, req *userpb.CreateActivityLogRequest) (*userpb.ActivityLogResponse, error) {
	log := &domain.ActivityLog{
		Name:        req.Name,
		Description: req.Description,
	}

	created, err := h.userService.CreateActivityLog(ctx, log)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create activity log: %v", err)
	}

	return &userpb.ActivityLogResponse{
		Id:          created.ID,
		Name:        created.Name,
		Description: created.Description,
		CreatedAt:   timestamppb.New(created.CreatedAt),
	}, nil
}

// ListActivityLogs lists activity logs
func (h *UserHandler) ListActivityLogs(ctx context.Context, req *userpb.ListActivityLogsRequest) (*userpb.ListActivityLogsResponse, error) {
	var limit, offset int32 = 10, 0
	if req.Page != nil {
		limit = req.Page.PageSize
		offset = (req.Page.Page - 1) * req.Page.PageSize
	}

	logs, err := h.userService.ListActivityLogs(ctx, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list activity logs: %v", err)
	}

	pbLogs := make([]*userpb.ActivityLog, len(logs))
	for i, log := range logs {
		pbLogs[i] = &userpb.ActivityLog{
			Id:          log.ID,
			Name:        log.Name,
			Description: log.Description,
			CreatedAt:   timestamppb.New(log.CreatedAt),
		}
	}

	return &userpb.ListActivityLogsResponse{
		Logs: pbLogs,
		Pagination: &userpb.PageResponse{
			Page:       req.Page.Page,
			PageSize:   limit,
			TotalPages: 1,
			TotalItems: int32(len(pbLogs)),
		},
	}, nil
}

// Helper function to safely get string value from pointer
func stringPtrValue(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}
