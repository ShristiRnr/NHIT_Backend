package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/ports"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/utils"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

// authService implements the AuthService interface
type authService struct {
	userRepo              ports.UserRepository
	sessionRepo           ports.SessionRepository
	refreshTokenRepo      ports.RefreshTokenRepository
	passwordResetRepo     ports.PasswordResetRepository
	emailVerificationRepo ports.EmailVerificationRepository
	jwtManager            *utils.JWTManager
	emailService          utils.EmailService
	kafkaPublisher        ports.KafkaPublisher
	notificationClient    ports.NotificationClient
	userServiceClient     userpb.UserManagementClient     // gRPC client for User Service
	orgClient             ports.OrganizationServiceClient // gRPC client for Organization Service
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo ports.UserRepository,
	sessionRepo ports.SessionRepository,
	refreshTokenRepo ports.RefreshTokenRepository,
	passwordResetRepo ports.PasswordResetRepository,
	emailVerificationRepo ports.EmailVerificationRepository,
	jwtManager *utils.JWTManager,
	emailService utils.EmailService,
	kafkaPublisher ports.KafkaPublisher,
	notificationClient ports.NotificationClient,
	userServiceClient userpb.UserManagementClient,
	orgClient ports.OrganizationServiceClient,
) ports.AuthService {
	return &authService{
		userRepo:              userRepo,
		sessionRepo:           sessionRepo,
		refreshTokenRepo:      refreshTokenRepo,
		passwordResetRepo:     passwordResetRepo,
		emailVerificationRepo: emailVerificationRepo,
		jwtManager:            jwtManager,
		emailService:          emailService,
		kafkaPublisher:        kafkaPublisher,
		notificationClient:    notificationClient,
		userServiceClient:     userServiceClient,
		orgClient:             orgClient,
	}
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

// Register registers a new user
func (s *authService) Register(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID, name, email, password string, roles []string) (*domain.LoginResponse, error) {
	// Validate password strength
	if err := utils.ValidatePasswordStrength(password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, tenantID, email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	// ✅ OPTION 2: Create user via User Service gRPC call (Industry Standard)
	// Match the ACTUAL generated proto structure
	createUserReq := &userpb.CreateUserRequest{
		TenantId: tenantID.String(), // Field 1
		Email:    email,             // Field 2
		Name:     name,              // Field 3
		Password: hashedPassword,    // Field 4 - hashed password
		// Roles will be assigned separately after user creation
	}

	userResp, err := s.userServiceClient.CreateUser(ctx, createUserReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create user via User Service: %w", err)
	}

	// Parse user ID from response
	userID, err := uuid.Parse(userResp.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID returned from User Service: %w", err)
	}

	log.Printf("✅ User created via User Service: %s (%s)", userResp.UserId, email)

	// Send verification email
	verificationToken, err := s.emailVerificationRepo.Create(ctx, userID, time.Now().Add(24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("failed to create verification token: %w", err)
	}

	// Send email with error handling
	if err := s.emailService.SendVerificationEmail(email, name, verificationToken.Token.String()); err != nil {
		// If email fails, notify user to update email
		fmt.Printf("⚠️  Failed to send verification email to %s: %v\n", email, err)
		if err := s.emailService.SendEmailUpdateNotification(email, name); err != nil {
			fmt.Printf("⚠️  Failed to send email update notification: %v\n", err)
		}
	}

	// Generate tokens
	accessToken, accessExpiresAt, err := s.jwtManager.GenerateAccessToken(
		userID.String(), email, name, tenantID.String(), "", roles, []string{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(userID.String(), tenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	if err := s.refreshTokenRepo.Create(ctx, &domain.RefreshToken{
		Token:     refreshToken,
		UserID:    userID,
		ExpiresAt: time.Unix(refreshExpiresAt, 0),
		CreatedAt: time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Create session
	session := &domain.Session{
		SessionID:    uuid.New(),
		UserID:       userID,
		SessionToken: accessToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Unix(accessExpiresAt, 0),
	}
	if _, err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &domain.LoginResponse{
		Token:            accessToken,
		RefreshToken:     refreshToken,
		UserID:           userID,
		Email:            email,
		Name:             name,
		Roles:            roles,
		Permissions:      []string{},
		LastLoginAt:      time.Now(),
		LastLoginIP:      "",
		TenantID:         tenantID,
		OrgID:            orgID,
		TokenExpiresAt:   accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// LoginGlobal authenticates a user without requiring tenant_id
func (s *authService) LoginGlobal(ctx context.Context, email, password string, orgID *uuid.UUID) (*domain.LoginResponse, error) {
	// Get user by email globally (across all tenants)
	user, err := s.userRepo.GetByEmailGlobal(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Verify password
	if err := utils.VerifyPassword(user.Password, password); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check if email is verified unless user is super admin
	if user.EmailVerifiedAt == nil {
		roles, err := s.userServiceClient.ListRolesOfUser(ctx, &userpb.GetUserRequest{
			UserId: user.UserID.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch user roles: %w", err)
		}

		isSuperAdmin := false
		for _, role := range roles.Roles {
			if role.Name == "SUPER_ADMIN" {
				isSuperAdmin = true
				break
			}
		}

		if !isSuperAdmin {
			return nil, fmt.Errorf("email not verified. Please verify your email before logging in")
		}
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.UserID, "", ""); err != nil {
		fmt.Printf("⚠️  Failed to update last login: %v\n", err)
	}

	// Auto-detect organization if not provided
	if orgID == nil {
		orgsResp, err := s.userServiceClient.ListUserOrganizations(ctx, &userpb.ListUserOrganizationsRequest{
			UserId: user.UserID.String(),
		})
		if err != nil {
			fmt.Printf("⚠️  Failed to list user organizations for user %s: %v\n", user.UserID, err)
		} else if len(orgsResp.Organizations) > 0 {
			var selected *userpb.UserOrganizationInfo
			for _, o := range orgsResp.Organizations {
				if o.IsCurrentContext {
					selected = o
					break
				}
			}
			if selected == nil {
				selected = orgsResp.Organizations[0]
			}
			if selected.OrgId != "" {
				if parsedOrgID, err := uuid.Parse(selected.OrgId); err != nil {
					fmt.Printf("⚠️  Invalid org_id %q for user %s: %v\n", selected.OrgId, user.UserID, err)
				} else {
					orgID = &parsedOrgID
				}
			}
		}
	}

	if orgID == nil {
		return nil, fmt.Errorf("no organization found for this user; please create an organization before logging in")
	}

	orgIDStr := orgID.String()

	// Fetch roles and permissions for JWT claims
	rolesResp, err := s.userServiceClient.ListRolesOfUser(ctx, &userpb.GetUserRequest{
		UserId: user.UserID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user roles for token: %w", err)
	}

	roleNames := make([]string, 0, len(rolesResp.Roles))
	permSet := make(map[string]struct{})
	for _, r := range rolesResp.Roles {
		if r.Name != "" {
			roleNames = append(roleNames, r.Name)
		}
		for _, p := range r.Permissions {
			permSet[p] = struct{}{}
		}
	}
	permissions := make([]string, 0, len(permSet))
	for p := range permSet {
		permissions = append(permissions, p)
	}

	accessToken, accessExpiresAt, err := s.jwtManager.GenerateAccessToken(
		user.UserID.String(), user.Email, user.Name, user.TenantID.String(), orgIDStr, roleNames, permissions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(user.UserID.String(), user.TenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	if err := s.refreshTokenRepo.Create(ctx, &domain.RefreshToken{
		Token:     refreshToken,
		UserID:    user.UserID,
		ExpiresAt: time.Unix(refreshExpiresAt, 0),
		CreatedAt: time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Create session
	session := &domain.Session{
		SessionID:    uuid.New(),
		UserID:       user.UserID,
		SessionToken: accessToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Unix(accessExpiresAt, 0),
	}
	createdSession, err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log session creation
	fmt.Printf("Global login: Created new session %s for user %s (tenant: %s) with expiry at %s\n",
		createdSession.SessionID, user.UserID, user.TenantID, createdSession.ExpiresAt.Format(time.RFC3339))

	// Create login history entry
	if s.userServiceClient != nil {
		// Extract IP and User-Agent from context metadata
		md, _ := metadata.FromIncomingContext(ctx)
		ipAddress := firstMetadataValue(md, "x-forwarded-for", "x-real-ip", "remote-addr")
		userAgent := firstMetadataValue(md, "user-agent")
		_, err = s.userServiceClient.CreateUserLoginHistory(ctx, &userpb.CreateUserLoginHistoryRequest{
			UserId:    user.UserID.String(),
			IpAddress: ipAddress,
			UserAgent: userAgent,
		})
		if err != nil {
			fmt.Printf("⚠️  Failed to create login history: %v\n", err)
			// Don't fail login if history creation fails
		}
	}

	return &domain.LoginResponse{
		Token:            accessToken,
		RefreshToken:     refreshToken,
		UserID:           user.UserID,
		Email:            user.Email,
		Name:             user.Name,
		Roles:            roleNames,
		Permissions:      permissions,
		LastLoginAt:      time.Now(),
		LastLoginIP:      "",
		TenantID:         user.TenantID,
		OrgID:            orgID,
		TokenExpiresAt:   accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		SessionID:        createdSession.SessionID,
	}, nil
}

// Login authenticates a user and returns tokens (tenant-specific)
func (s *authService) Login(ctx context.Context, email, password string, tenantID uuid.UUID, orgID *uuid.UUID) (*domain.LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Verify password
	if err := utils.VerifyPassword(user.Password, password); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check if email is verified unless user is super admin
	if user.EmailVerifiedAt == nil {
		roles, err := s.userServiceClient.ListRolesOfUser(ctx, &userpb.GetUserRequest{
			UserId: user.UserID.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch user roles: %w", err)
		}

		isSuperAdmin := false
		for _, role := range roles.Roles {
			if role.Name == "SUPER_ADMIN" {
				isSuperAdmin = true
				break
			}
		}

		if !isSuperAdmin {
			return nil, fmt.Errorf("email not verified. Please verify your email before logging in")
		}
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.UserID, "", ""); err != nil {
		fmt.Printf("⚠️  Failed to update last login: %v\n", err)
	}

	// Auto-detect organization if not provided
	if orgID == nil {
		orgsResp, err := s.userServiceClient.ListUserOrganizations(ctx, &userpb.ListUserOrganizationsRequest{
			UserId: user.UserID.String(),
		})
		if err != nil {
			fmt.Printf("⚠️  Failed to list user organizations for user %s: %v\n", user.UserID, err)
		} else if len(orgsResp.Organizations) > 0 {
			var selected *userpb.UserOrganizationInfo
			for _, o := range orgsResp.Organizations {
				if o.IsCurrentContext {
					selected = o
					break
				}
			}
			if selected == nil {
				selected = orgsResp.Organizations[0]
			}
			if selected.OrgId != "" {
				if parsedOrgID, err := uuid.Parse(selected.OrgId); err != nil {
					fmt.Printf("⚠️  Invalid org_id %q for user %s: %v\n", selected.OrgId, user.UserID, err)
				} else {
					orgID = &parsedOrgID
				}
			}
		}
	}

	if orgID == nil {
		return nil, fmt.Errorf("no organization found for this user; please create an organization before logging in")
	}

	orgIDStr := orgID.String()

	// Fetch roles and permissions for JWT claims
	rolesResp, err := s.userServiceClient.ListRolesOfUser(ctx, &userpb.GetUserRequest{
		UserId: user.UserID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user roles for token: %w", err)
	}

	roleNames := make([]string, 0, len(rolesResp.Roles))
	permSet := make(map[string]struct{})
	for _, r := range rolesResp.Roles {
		if r.Name != "" {
			roleNames = append(roleNames, r.Name)
		}
		for _, p := range r.Permissions {
			permSet[p] = struct{}{}
		}
	}
	permissions := make([]string, 0, len(permSet))
	for p := range permSet {
		permissions = append(permissions, p)
	}

	accessToken, accessExpiresAt, err := s.jwtManager.GenerateAccessToken(
		user.UserID.String(), user.Email, user.Name, user.TenantID.String(), orgIDStr, roleNames, permissions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(user.UserID.String(), user.TenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	if err := s.refreshTokenRepo.Create(ctx, &domain.RefreshToken{
		Token:     refreshToken,
		UserID:    user.UserID,
		ExpiresAt: time.Unix(refreshExpiresAt, 0),
		CreatedAt: time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Create session
	session := &domain.Session{
		SessionID:    uuid.New(),
		UserID:       user.UserID,
		SessionToken: accessToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Unix(accessExpiresAt, 0),
	}
	createdSession, err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log session creation
	fmt.Printf("Created new session %s for user %s with expiry at %s\n",
		createdSession.SessionID, user.UserID, createdSession.ExpiresAt.Format(time.RFC3339))

	// Create login history entry
	if s.userServiceClient != nil {
		// Extract IP and User-Agent from context metadata
		md, _ := metadata.FromIncomingContext(ctx)
		ipAddress := firstMetadataValue(md, "x-forwarded-for", "x-real-ip", "remote-addr")
		userAgent := firstMetadataValue(md, "user-agent")
		_, err = s.userServiceClient.CreateUserLoginHistory(ctx, &userpb.CreateUserLoginHistoryRequest{
			UserId:    user.UserID.String(),
			IpAddress: ipAddress,
			UserAgent: userAgent,
		})
		if err != nil {
			fmt.Printf("⚠️  Failed to create login history: %v\n", err)
			// Don't fail login if history creation fails
		}
	}

	return &domain.LoginResponse{
		Token:            accessToken,
		RefreshToken:     refreshToken,
		UserID:           user.UserID,
		Email:            user.Email,
		Name:             user.Name,
		Roles:            roleNames,
		Permissions:      permissions,
		LastLoginAt:      time.Now(),
		LastLoginIP:      "",
		TenantID:         user.TenantID,
		OrgID:            orgID,
		TokenExpiresAt:   accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		SessionID:        createdSession.SessionID,
	}, nil
}

// Logout logs out a user by invalidating their refresh token and session
func (s *authService) Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	// Delete refresh token
	if err := s.refreshTokenRepo.Delete(ctx, refreshToken); err != nil {
		fmt.Printf("⚠️  Failed to delete refresh token: %v\n", err)
	}

	// Find token associated with session to invalidate the specific session
	// Default to invalidating all sessions for safety
	invalidatedSessions := 0

	// Get all sessions for the user
	sessions, err := s.sessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		fmt.Printf("⚠️  Failed to get sessions for user %s: %v\n", userID, err)
		// Continue despite error - we'll try to delete sessions anyway
	}

	// Try to find a session with this refresh token
	sessionFound := false
	for _, session := range sessions {
		// Since we don't store the refresh token in the session, we delete all sessions
		if err := s.sessionRepo.Delete(ctx, session.SessionID); err != nil {
			fmt.Printf("⚠️  Failed to delete session %s: %v\n", session.SessionID, err)
		} else {
			invalidatedSessions++
			sessionFound = true
		}
	}

	if !sessionFound && len(sessions) > 0 {
		fmt.Printf("⚠️  No sessions found to invalidate for user %s\n", userID)
	}

	fmt.Printf("Logout complete for user %s: Invalidated %d session(s)\n", userID, invalidatedSessions)
	return nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string, tenantID uuid.UUID, orgID *uuid.UUID) (*domain.LoginResponse, error) {
	// Validate refresh token
	userIDStr, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	// Check if refresh token exists in database
	storedUserID, err := s.refreshTokenRepo.GetUserIDByToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh token not found or expired")
	}

	if storedUserID != userID {
		return nil, fmt.Errorf("refresh token user ID mismatch")
	}

	// Get user details by ID from the refresh token
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Generate new tokens
	orgIDStr := ""
	if orgID != nil {
		orgIDStr = orgID.String()
	}

	// Fetch roles and permissions for JWT claims
	rolesResp, err := s.userServiceClient.ListRolesOfUser(ctx, &userpb.GetUserRequest{
		UserId: user.UserID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user roles for token: %w", err)
	}

	roleNames := make([]string, 0, len(rolesResp.Roles))
	permSet := make(map[string]struct{})
	for _, r := range rolesResp.Roles {
		if r.Name != "" {
			roleNames = append(roleNames, r.Name)
		}
		for _, p := range r.Permissions {
			permSet[p] = struct{}{}
		}
	}
	permissions := make([]string, 0, len(permSet))
	for p := range permSet {
		permissions = append(permissions, p)
	}

	accessToken, accessExpiresAt, err := s.jwtManager.GenerateAccessToken(
		user.UserID.String(), user.Email, user.Name, user.TenantID.String(), orgIDStr, roleNames, permissions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(user.UserID.String(), user.TenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Delete old refresh token
	if err := s.refreshTokenRepo.Delete(ctx, refreshToken); err != nil {
		fmt.Printf("⚠️  Failed to delete old refresh token: %v\n", err)
	}

	// Store new refresh token
	if err := s.refreshTokenRepo.Create(ctx, &domain.RefreshToken{
		Token:     newRefreshToken,
		UserID:    user.UserID,
		ExpiresAt: time.Unix(refreshExpiresAt, 0),
		CreatedAt: time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	// Handle existing sessions
	// First, check if there are any active sessions
	existingSessions, err := s.sessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		fmt.Printf("⚠️  Error checking existing sessions: %v\n", err)
		// Continue even if there's an error checking sessions
	}

	// Create new session regardless of existing ones
	session := &domain.Session{
		SessionID:    uuid.New(),
		UserID:       user.UserID,
		SessionToken: accessToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Unix(accessExpiresAt, 0),
	}
	createdSession, err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log the session refresh activity for debugging purposes
	expiredSessionsCount := 0
	for _, existingSession := range existingSessions {
		if existingSession.ExpiresAt.Before(time.Now()) {
			expiredSessionsCount++
		}
	}

	if len(existingSessions) > 0 {
		fmt.Printf("Refreshed session for user %s: Found %d existing sessions (%d expired)\n",
			userID, len(existingSessions), expiredSessionsCount)
	}

	// Update last login time
	if err := s.userRepo.UpdateLastLogin(ctx, user.UserID, "", ""); err != nil {
		fmt.Printf("⚠️  Failed to update last login during refresh: %v\n", err)
		// Continue despite error
	}

	return &domain.LoginResponse{
		Token:            accessToken,
		RefreshToken:     newRefreshToken,
		UserID:           user.UserID,
		Email:            user.Email,
		Name:             user.Name,
		Roles:            roleNames,
		Permissions:      permissions,
		LastLoginAt:      time.Now(),
		LastLoginIP:      "",
		TenantID:         user.TenantID,
		OrgID:            orgID,
		TokenExpiresAt:   accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		SessionID:        createdSession.SessionID,
	}, nil
}

// ValidateToken validates an access token
func (s *authService) ValidateToken(ctx context.Context, token string) (*domain.TokenValidation, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return &domain.TokenValidation{Valid: false}, fmt.Errorf("invalid token: %w", err)
	}

	// Check if session exists and is valid
	session, err := s.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return &domain.TokenValidation{Valid: false}, fmt.Errorf("session not found")
	}

	if session.ExpiresAt.Before(time.Now()) {
		return &domain.TokenValidation{Valid: false}, fmt.Errorf("session expired")
	}

	userID, _ := uuid.Parse(claims.UserID)
	tenantID, _ := uuid.Parse(claims.TenantID)
	var orgID *uuid.UUID
	if claims.OrgID != "" {
		oid, _ := uuid.Parse(claims.OrgID)
		orgID = &oid
	}

	expiresAt := time.Now()
	if claims.RegisteredClaims.ExpiresAt != nil {
		expiresAt = claims.RegisteredClaims.ExpiresAt.Time
	}

	return &domain.TokenValidation{
		Valid:       true,
		UserID:      userID,
		Email:       claims.Email,
		Name:        claims.Name,
		TenantID:    tenantID,
		OrgID:       orgID,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
		ExpiresAt:   expiresAt,
	}, nil
}

// SendVerificationEmail sends a verification email to the user
func (s *authService) SendVerificationEmail(ctx context.Context, userID uuid.UUID) error {
	// Get user details (in production, call User Service)
	// For now, we'll use mock data
	email := "user@example.com"
	name := "User"

	// Create verification token
	verificationToken, err := s.emailVerificationRepo.Create(ctx, userID, time.Now().Add(24*time.Hour))
	if err != nil {
		return fmt.Errorf("failed to create verification token: %w", err)
	}

	// Send email with error handling
	if err := s.emailService.SendVerificationEmail(email, name, verificationToken.Token.String()); err != nil {
		fmt.Printf("⚠️  Failed to send verification email: %v\n", err)
		// Send notification about email failure
		if err := s.emailService.SendEmailUpdateNotification(email, name); err != nil {
			fmt.Printf("⚠️  Failed to send email update notification: %v\n", err)
		}
		return fmt.Errorf("failed to send verification email. Please update your email address")
	}

	return nil
}

// VerifyEmail verifies a user's email address
func (s *authService) VerifyEmail(ctx context.Context, userID uuid.UUID, token string) error {
	tokenUUID, err := uuid.Parse(token)
	if err != nil {
		return fmt.Errorf("invalid verification token format")
	}

	// Verify token
	valid, err := s.emailVerificationRepo.Verify(ctx, userID, tokenUUID)
	if err != nil {
		return fmt.Errorf("failed to verify token: %w", err)
	}

	if !valid {
		return fmt.Errorf("invalid or expired verification token")
	}

	// Update user email verified status
	if err := s.userRepo.VerifyEmail(ctx, userID); err != nil {
		return fmt.Errorf("failed to update email verification status: %w", err)
	}

	// Delete verification token
	if err := s.emailVerificationRepo.Delete(ctx, userID); err != nil {
		fmt.Printf("⚠️  Failed to delete verification token: %v\n", err)
	}

	return nil
}

// ForgotPassword initiates a password reset flow
func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	// Get user by email (need tenant ID in production)
	// For now, using a mock tenant ID
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	user, err := s.userRepo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		// Don't reveal if user exists or not for security
		return nil
	}

	// Create password reset token
	resetToken := uuid.New()
	expiresAt := time.Now().Add(1 * time.Hour)
	_, err = s.passwordResetRepo.Create(ctx, user.UserID, resetToken, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	// Use notification service for sending emails
	if s.notificationClient != nil {
		if err := s.notificationClient.SendPasswordResetEmail(ctx, email, user.Name, resetToken.String()); err != nil {
			fmt.Printf("⚠️  Failed to send password reset email via notification service: %v\n", err)
			// Fallback to legacy email service
			return s.sendPasswordResetEmailLegacy(email, user.Name, resetToken.String())
		}
	} else {
		// Fallback to legacy email service if notification client is not configured
		return s.sendPasswordResetEmailLegacy(email, user.Name, resetToken.String())
	}

	if s.kafkaPublisher != nil {
		eventTimestamp := time.Now()
		eventData := map[string]interface{}{
			"user_id":      user.UserID.String(),
			"tenant_id":    user.TenantID.String(),
			"email":        user.Email,
			"name":         user.Name,
			"reset_token":  resetToken.String(),
			"expires_at":   expiresAt.Format(time.RFC3339),
			"requested_at": eventTimestamp.Format(time.RFC3339),
		}
		if err := s.kafkaPublisher.Publish(ctx, "user.password_reset_requested", eventData); err != nil {
			log.Printf("⚠️  Failed to publish password reset event: %v\n", err)
		}
	}

	return nil
}

// ResetPasswordByToken resets a user's password using a reset token
func (s *authService) ResetPasswordByToken(ctx context.Context, token, newPassword string) error {
	// Validate password strength
	if err := utils.ValidatePasswordStrength(newPassword); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	tokenUUID, err := uuid.Parse(token)
	if err != nil {
		return fmt.Errorf("invalid reset token format")
	}

	// Get password reset record
	resetRecord, err := s.passwordResetRepo.GetByToken(ctx, tokenUUID)
	if err != nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	// Check if token is expired
	if resetRecord.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("reset token has expired")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, resetRecord.UserID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Delete reset token
	if err := s.passwordResetRepo.Delete(ctx, tokenUUID); err != nil {
		fmt.Printf("⚠️  Failed to delete reset token: %v\n", err)
	}

	// Invalidate all sessions for security
	if err := s.InvalidateAllSessions(ctx, resetRecord.UserID); err != nil {
		fmt.Printf("⚠️  Failed to invalidate sessions: %v\n", err)
	}

	return nil
}

// InvalidateAllSessions invalidates all sessions for a user
func (s *authService) InvalidateAllSessions(ctx context.Context, userID uuid.UUID) error {
	// Get all sessions for user
	sessions, err := s.GetActiveSessions(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get sessions: %w", err)
	}

	// Delete each session
	for _, session := range sessions {
		if err := s.sessionRepo.Delete(ctx, session.SessionID); err != nil {
			fmt.Printf("⚠️  Failed to delete session %s: %v\n", session.SessionID, err)
		}
	}

	return nil
}

// GetActiveSessions gets all active sessions for a user
func (s *authService) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	// Use sessionRepo's GetByUserID method
	return s.sessionRepo.GetByUserID(ctx, userID)
}

// SwitchOrganization switches a user to a different organization within the same tenant
func (s *authService) SwitchOrganization(ctx context.Context, userID uuid.UUID, newOrgID uuid.UUID, tenantID uuid.UUID) (*domain.LoginResponse, error) {
	// Verify user exists and belongs to the tenant
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.TenantID != tenantID {
		return nil, fmt.Errorf("user does not belong to the specified tenant")
	}

	// Verify the organization exists and belongs to the SAME tenant (critical validation)
	org, err := s.orgClient.GetOrganization(ctx, newOrgID)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	if org.TenantID != tenantID {
		return nil, fmt.Errorf("organization does not belong to the specified tenant - cross-tenant switching not allowed")
	}

	// Additional validation: ensure user has access to organizations within the same tenant
	// Get user's current organizations to validate switching rights
	userOrgs, err := s.orgClient.ListUserOrganizations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	// Check if user has access to the target organization
	// User can switch to any organization they belong to within the same tenant
	hasAccess := false
	for _, userOrg := range userOrgs {
		if userOrg.OrgID == newOrgID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return nil, fmt.Errorf("user does not have access to switch to this organization")
	}

	// Invalidate all existing sessions for the user
	if err := s.InvalidateAllSessions(ctx, userID); err != nil {
		log.Printf("⚠️  Failed to invalidate sessions for user %s: %v", userID, err)
	}

	// Create new session with the new organization context
	sessionID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hour session

	session := &domain.Session{
		SessionID:    sessionID,
		UserID:       userID,
		SessionToken: "", // Will be set by repository
		CreatedAt:    time.Now(),
		ExpiresAt:    expiresAt,
	}

	createdSession, err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Generate new tokens for the new organization context
	// Note: UserData doesn't have Roles/Permissions fields, so we'll use empty arrays for now
	token, err := s.jwtManager.GenerateToken(userID, user.Email, user.Name, tenantID, &newOrgID, []string{}, []string{})
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(userID.String(), tenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	refreshTokenRecord := &domain.RefreshToken{
		Token:     refreshToken,
		UserID:    userID,
		ExpiresAt: time.Unix(refreshExpiresAt, 0),
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshTokenRecord); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	log.Printf("🔄 User %s switched to organization %s within tenant %s", userID, newOrgID, tenantID)

	return &domain.LoginResponse{
		UserID:           userID,
		Email:            user.Email,
		Name:             user.Name,
		Roles:            []string{}, // UserData doesn't have roles, using empty array
		Permissions:      []string{}, // UserData doesn't have permissions, using empty array
		TenantID:         tenantID,
		OrgID:            &newOrgID,
		Token:            token,
		RefreshToken:     refreshToken,
		TokenExpiresAt:   expiresAt.Unix(),
		RefreshExpiresAt: refreshExpiresAt,
		SessionID:        createdSession.SessionID,
		LastLoginAt:      time.Now(),
		LastLoginIP:      "", // Will be set by middleware
	}, nil
}
