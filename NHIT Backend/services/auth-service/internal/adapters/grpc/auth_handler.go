package grpc

import (
	"context"
	"fmt"
	"strings"

	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/ports"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	authpb.UnimplementedAuthServiceServer
	authService ports.AuthService
	orgClient   ports.OrganizationServiceClient
}

func NewAuthHandler(authService ports.AuthService, orgClient ports.OrganizationServiceClient) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		orgClient:   orgClient,
	}
}

func toProtoRoles(roleNames []string) []authpb.UserRole {
	protoRoles := make([]authpb.UserRole, len(roleNames))
	for i, name := range roleNames {
		switch name {
		case "SUPER_ADMIN":
			protoRoles[i] = authpb.UserRole_SUPER_ADMIN
		default:
			protoRoles[i] = authpb.UserRole_USER_ROLE_UNSPECIFIED
		}
	}
	return protoRoles
}

// RegisterUser registers a new user
func (h *AuthHandler) RegisterUser(ctx context.Context, req *authpb.RegisterUserRequest) (*authpb.RegisterUserResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	// Convert roles from proto enum to strings
	roles := make([]string, len(req.Roles))
	for i, role := range req.Roles {
		roles[i] = role.String()
	}

	var orgID *uuid.UUID // No organization provided in this request path yet
	response, err := h.authService.Register(ctx, tenantID, orgID, req.Name, req.Email, req.Password, roles)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	// Convert roles back to proto enum
	protoRoles := toProtoRoles(response.Roles)

	return &authpb.RegisterUserResponse{
		UserId:      response.UserID.String(),
		Name:        response.Name,
		Email:       response.Email,
		Roles:       protoRoles,
		Permissions: response.Permissions,
	}, nil
}

// VerifyEmail verifies a user's email
func (h *AuthHandler) VerifyEmail(ctx context.Context, req *authpb.VerifyEmailRequest) (*authpb.VerifyEmailResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	err = h.authService.VerifyEmail(ctx, userID, req.VerificationToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify email: %v", err)
	}

	return &authpb.VerifyEmailResponse{Success: true}, nil
}

// ForgotPassword initiates password reset
func (h *AuthHandler) ForgotPassword(ctx context.Context, req *authpb.ForgotPasswordRequest) (*authpb.ForgotPasswordResponse, error) {
	err := h.authService.ForgotPassword(ctx, req.Email)
	if err != nil {
		// Don't reveal if email exists
		return &authpb.ForgotPasswordResponse{Success: true}, nil
	}

	return &authpb.ForgotPasswordResponse{Success: true}, nil
}

// ResetPasswordByToken resets password using token
func (h *AuthHandler) ResetPasswordByToken(ctx context.Context, req *authpb.ResetPasswordByTokenRequest) (*authpb.ResetPasswordByTokenResponse, error) {
	err := h.authService.ResetPasswordByToken(ctx, req.Otp, req.NewPassword)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reset password: %v", err)
	}

	return &authpb.ResetPasswordByTokenResponse{Success: true}, nil
}

// Login authenticates a user
func (h *AuthHandler) Login(ctx context.Context, req *authpb.UserLoginRequest) (*authpb.UserLoginResponse, error) {
	var tenantID uuid.UUID
	var err error

	// If tenant_id is not provided, auto-detect it from email
	if req.TenantId == "" {
		// Auto-detect tenant by looking up user globally
		response, err := h.authService.LoginGlobal(ctx, req.Login, req.Password, nil)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
		}

		// Convert roles to proto enum
		protoRoles := toProtoRoles(response.Roles)

		orgIDStr := ""
		if response.OrgID != nil {
			orgIDStr = response.OrgID.String()
		}

		return &authpb.UserLoginResponse{
			Token:            response.Token,
			RefreshToken:     response.RefreshToken,
			UserId:           response.UserID.String(),
			Email:            response.Email,
			Name:             response.Name,
			Roles:            protoRoles,
			Permissions:      response.Permissions,
			LastLoginAt:      response.LastLoginAt.Format("2006-01-02T15:04:05Z"),
			LastLoginIp:      response.LastLoginIP,
			TenantId:         response.TenantID.String(),
			OrgId:            orgIDStr,
			TokenExpiresAt:   response.TokenExpiresAt,
			RefreshExpiresAt: response.RefreshExpiresAt,
		}, nil
	}

	// Original tenant-specific flow (when tenant_id is provided)
	tenantID, err = uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	var orgID *uuid.UUID
	if req.OrgId != "" {
		oid, err := uuid.Parse(req.OrgId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid org_id: %v", err)
		}
		orgID = &oid
	}

	response, err := h.authService.Login(ctx, req.Login, req.Password, tenantID, orgID)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}

	// Convert roles to proto enum
	protoRoles := toProtoRoles(response.Roles)

	orgIDStr := ""
	if response.OrgID != nil {
		orgIDStr = response.OrgID.String()
	}

	return &authpb.UserLoginResponse{
		Token:            response.Token,
		RefreshToken:     response.RefreshToken,
		UserId:           response.UserID.String(),
		Email:            response.Email,
		Name:             response.Name,
		Roles:            protoRoles,
		Permissions:      response.Permissions,
		LastLoginAt:      response.LastLoginAt.Format("2006-01-02T15:04:05Z"),
		LastLoginIp:      response.LastLoginIP,
		TenantId:         response.TenantID.String(),
		OrgId:            orgIDStr,
		TokenExpiresAt:   response.TokenExpiresAt,
		RefreshExpiresAt: response.RefreshExpiresAt,
	}, nil
}

// TODO: GlobalLogin - Enable after regenerating protobuf Go code
// Run: protoc --go_out=. --go-grpc_out=. api/proto/auth.proto
/*
// GlobalLogin authenticates a user without requiring tenant_id (tenant-agnostic)
func (h *AuthHandler) GlobalLogin(ctx context.Context, req *authpb.GlobalLoginRequest) (*authpb.UserLoginResponse, error) {
	var orgID *uuid.UUID
	if req.OrgId != "" {
		oid, err := uuid.Parse(req.OrgId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid org_id: %v", err)
		}
		orgID = &oid
	}

	response, err := h.authService.LoginGlobal(ctx, req.Login, req.Password, orgID)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}

	// Convert roles to proto enum
	protoRoles := make([]authpb.UserRole, len(response.Roles))
	for i := range response.Roles {
		protoRoles[i] = authpb.UserRole_USER_ROLE_UNSPECIFIED
	}

	orgIDStr := ""
	if response.OrgID != nil {
		orgIDStr = response.OrgID.String()
	}

	return &authpb.UserLoginResponse{
		Token:            response.Token,
		RefreshToken:     response.RefreshToken,
		UserId:           response.UserID.String(),
		Email:            response.Email,
		Name:             response.Name,
		Roles:            protoRoles,
		Permissions:      response.Permissions,
		LastLoginAt:      response.LastLoginAt.Format("2006-01-02T15:04:05Z"),
		LastLoginIp:      response.LastLoginIP,
		TenantId:         response.TenantID.String(),
		OrgId:            orgIDStr,
		TokenExpiresAt:   response.TokenExpiresAt,
		RefreshExpiresAt: response.RefreshExpiresAt,
	}, nil
}
*/

// Logout logs out a user
func (h *AuthHandler) Logout(ctx context.Context, req *authpb.UserLogoutRequest) (*authpb.UserLogoutResponse, error) {
	var userID uuid.UUID
	if req.UserId != "" {
		parsedID, err := uuid.Parse(req.UserId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
		}
		userID = parsedID
	}

	err := h.authService.Logout(ctx, userID, req.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to logout: %v", err)
	}

	return &authpb.UserLogoutResponse{Success: true}, nil
}

// RefreshToken refreshes access token
func (h *AuthHandler) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	var orgID *uuid.UUID
	if req.OrgId != "" {
		oid, err := uuid.Parse(req.OrgId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid org_id: %v", err)
		}
		orgID = &oid
	}

	response, err := h.authService.RefreshToken(ctx, req.RefreshToken, tenantID, orgID)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to refresh token: %v", err)
	}

	logMsg := fmt.Sprintf("Session refreshed with ID: %s", response.SessionID)
	fmt.Println(logMsg)

	return &authpb.RefreshTokenResponse{
		Token:            response.Token,
		RefreshToken:     response.RefreshToken,
		TokenExpiresAt:   response.TokenExpiresAt,
		RefreshExpiresAt: response.RefreshExpiresAt,
	}, nil
}

// ValidateToken validates an access token and returns user/tenant/org context
func (h *AuthHandler) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Errorf(codes.InvalidArgument, "token is required")
	}

	validation, err := h.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}
	if validation == nil || !validation.Valid {
		return &authpb.ValidateTokenResponse{Valid: false}, nil
	}

	resp := &authpb.ValidateTokenResponse{
		Valid:       true,
		UserId:      validation.UserID.String(),
		Email:       validation.Email,
		Name:        validation.Name,
		TenantId:    validation.TenantID.String(),
		Roles:       validation.Roles,
		Permissions: validation.Permissions,
		ExpiresAt:   validation.ExpiresAt.Unix(),
	}
	if validation.OrgID != nil {
		resp.OrgId = validation.OrgID.String()
	}

	return resp, nil
}

// InitiateSSO initiates SSO login (placeholder)
func (h *AuthHandler) InitiateSSO(ctx context.Context, req *authpb.InitiateSSORequest) (*authpb.InitiateSSOResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "SSO not implemented yet")
}

// CompleteSSO completes SSO login (placeholder)
func (h *AuthHandler) CompleteSSO(ctx context.Context, req *authpb.CompleteSSORequest) (*authpb.UserLoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "SSO not implemented yet")
}

// InitiateSSOLogout initiates SSO logout (placeholder)
func (h *AuthHandler) InitiateSSOLogout(ctx context.Context, req *authpb.InitiateSSOLogoutRequest) (*authpb.InitiateSSOLogoutResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "SSO logout not implemented yet")
}

// CompleteSSOLogout completes SSO logout (placeholder)
func (h *AuthHandler) CompleteSSOLogout(ctx context.Context, req *authpb.CompleteSSOLogoutRequest) (*authpb.UserLogoutResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "SSO logout not implemented yet")
}

// SendVerificationEmail sends verification email
func (h *AuthHandler) SendVerificationEmail(ctx context.Context, req *authpb.SendVerificationEmailRequest) (*authpb.SendVerificationEmailResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	err = h.authService.SendVerificationEmail(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send verification email: %v", err)
	}

	return &authpb.SendVerificationEmailResponse{Success: true}, nil
}

// SendPasswordResetEmail sends password reset email
func (h *AuthHandler) SendPasswordResetEmail(ctx context.Context, req *authpb.SendPasswordResetEmailRequest) (*authpb.SendPasswordResetEmailResponse, error) {
	err := h.authService.ForgotPassword(ctx, req.Email)
	if err != nil {
		// Don't reveal if email exists
		return &authpb.SendPasswordResetEmailResponse{Success: true}, nil
	}

	return &authpb.SendPasswordResetEmailResponse{Success: true}, nil
}

// SwitchOrganization switches a user to a different organization within the same tenant
func (h *AuthHandler) SwitchOrganization(ctx context.Context, req *authpb.SwitchOrganizationRequest) (*authpb.UserLoginResponse, error) {
	// Extract user information from JWT context (set by middleware)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	// Get user ID from context
	userIDs := md.Get("user_id")
	if len(userIDs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing user_id in context")
	}
	userID, err := uuid.Parse(userIDs[0])
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id in context: %v", err)
	}

	// Get tenant ID from context
	tenantIDs := md.Get("tenant_id")
	if len(tenantIDs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing tenant_id in context")
	}
	tenantID, err := uuid.Parse(tenantIDs[0])
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id in context: %v", err)
	}

	// Determine target organization ID
	var orgID uuid.UUID
	
	if req.OrgName != "" {
		// Switch by organization name
		// Get user's organizations and find by name
		userOrgs, err := h.orgClient.ListUserOrganizations(ctx, userID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get user organizations: %v", err)
		}

		// Find organization by name (case-insensitive)
		found := false
		for _, org := range userOrgs {
			if strings.EqualFold(org.Name, req.OrgName) {
				orgID = org.OrgID
				found = true
				break
			}
		}

		if !found {
			return nil, status.Errorf(codes.NotFound, "organization with name '%s' not found or you don't have access", req.OrgName)
		}
	} else if req.OrgId != "" {
		// Switch by organization ID
		orgID, err = uuid.Parse(req.OrgId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid org_id: %v", err)
		}
	} else {
		return nil, status.Error(codes.InvalidArgument, "either org_id or org_name must be provided")
	}

	// Additional validation: ensure target org is not the same as current org
	currentOrgIDs := md.Get("org_id")
	if len(currentOrgIDs) > 0 {
		currentOrgID, err := uuid.Parse(currentOrgIDs[0])
		if err == nil && currentOrgID == orgID {
			return nil, status.Error(codes.InvalidArgument, "already in this organization")
		}
	}

	response, err := h.authService.SwitchOrganization(ctx, userID, orgID, tenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to switch organization: %v", err)
	}

	// Convert roles to proto enum
	protoRoles := toProtoRoles(response.Roles)

	orgIDStr := ""
	if response.OrgID != nil {
		orgIDStr = response.OrgID.String()
	}

	return &authpb.UserLoginResponse{
		Token:            response.Token,
		RefreshToken:     response.RefreshToken,
		UserId:           response.UserID.String(),
		Email:            response.Email,
		Name:             response.Name,
		Roles:            protoRoles,
		Permissions:      response.Permissions,
		LastLoginAt:      response.LastLoginAt.Format("2006-01-02T15:04:05Z"),
		LastLoginIp:      response.LastLoginIP,
		TenantId:         response.TenantID.String(),
		OrgId:            orgIDStr,
		TokenExpiresAt:   response.TokenExpiresAt,
		RefreshExpiresAt: response.RefreshExpiresAt,
	}, nil
}
