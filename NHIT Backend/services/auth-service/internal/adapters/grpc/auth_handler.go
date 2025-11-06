package grpc

import (
	"context"

	"github.com/google/uuid"
	authpb "github.com/ShristiRnr/NHIT_Backend/api/proto"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	authpb.UnimplementedAuthServiceServer
	authService ports.AuthService
}

func NewAuthHandler(authService ports.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
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

	response, err := h.authService.Register(ctx, tenantID, req.Name, req.Email, req.Password, roles)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	// Convert roles back to proto enum
	protoRoles := make([]authpb.UserRole, len(response.Roles))
	for i := range response.Roles {
		protoRoles[i] = authpb.UserRole_USER_ROLE_UNSPECIFIED // Default
	}

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
	err := h.authService.ResetPasswordByToken(ctx, req.Token, req.NewPassword)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reset password: %v", err)
	}

	return &authpb.ResetPasswordByTokenResponse{Success: true}, nil
}

// Login authenticates a user
func (h *AuthHandler) Login(ctx context.Context, req *authpb.UserLoginRequest) (*authpb.UserLoginResponse, error) {
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

	response, err := h.authService.Login(ctx, req.Login, req.Password, tenantID, orgID)
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

// Logout logs out a user
func (h *AuthHandler) Logout(ctx context.Context, req *authpb.UserLogoutRequest) (*authpb.UserLogoutResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	err = h.authService.Logout(ctx, userID, req.RefreshToken)
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

	return &authpb.RefreshTokenResponse{
		Token:            response.Token,
		RefreshToken:     response.RefreshToken,
		TokenExpiresAt:   response.TokenExpiresAt,
		RefreshExpiresAt: response.RefreshExpiresAt,
	}, nil
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
