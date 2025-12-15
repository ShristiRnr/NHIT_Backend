package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo           ports.UserRepository
	tenantRepo         ports.TenantRepository
	userRoleRepo       ports.UserRoleRepository
	roleRepo           ports.RoleRepository
	permissionRepo     ports.PermissionRepository
	loginHistoryRepo   ports.LoginHistoryRepository
	activityLogRepo    ports.ActivityLogRepository
}

// NewUserService creates a new user service instance

func NewUserService(userRepo ports.UserRepository, tenantRepo ports.TenantRepository, userRoleRepo ports.UserRoleRepository, roleRepo ports.RoleRepository, permissionRepo ports.PermissionRepository, loginHistoryRepo ports.LoginHistoryRepository, activityLogRepo ports.ActivityLogRepository) ports.UserService {
	return &userService{
		userRepo:         userRepo,
		tenantRepo:       tenantRepo,
		userRoleRepo:     userRoleRepo,
		roleRepo:         roleRepo,
		permissionRepo:   permissionRepo,
		loginHistoryRepo: loginHistoryRepo,
		activityLogRepo:  activityLogRepo,
	}
}

func (s *userService) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	// TODO: Hash password before storing (add bcrypt after module setup)
	// For now, store password as-is (NOT PRODUCTION READY)

	// Default to active if not specified (or boolean default is false so force true)
	// If caller explicitly wanted false, they would need a different flow or we check if it was set (but bool makes it hard).
	// Requirement: "newly created user are active by default"
	user.IsActive = true
	
	// Auto-verify email for admin-created users
	now := time.Now()
	if user.EmailVerifiedAt == nil {
		user.EmailVerifiedAt = &now
	}

	// Create user in repository
	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

func (s *userService) GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	// TODO: Hash password if being updated (add bcrypt after module setup)

	updatedUser, err := s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updatedUser, nil
}

func (s *userService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (s *userService) ListUsersByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]*domain.User, error) {
	users, err := s.userRepo.ListByTenant(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (s *userService) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if err := s.userRoleRepo.AssignRole(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}
	return nil
}

func (s *userService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	roles, err := s.userRoleRepo.ListRolesByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return roles, nil
}

// ensureSuperAdminRoleForTenant ensures that a SUPER_ADMIN system role with full permissions
// exists for the given tenant. If it already exists, it is returned. Otherwise it is created
// using the global permission catalog. This helper is internal and does not expose any new
// endpoints; it is used by CreateTenant and can be reused by other internal flows if needed.
func (s *userService) ensureSuperAdminRoleForTenant(ctx context.Context, tenantID uuid.UUID, createdBy string) (*domain.Role, error) {
	// List existing roles for this tenant and try to find SUPER_ADMIN
	existingRoles, err := s.roleRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles for tenant %s: %w", tenantID, err)
	}

	for _, r := range existingRoles {
		if r != nil && r.Name == "SUPER_ADMIN" {
			return r, nil
		}
	}

	// No SUPER_ADMIN role found – create a new one with all available permissions
	perms, err := s.permissionRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions for super admin role: %w", err)
	}
	if len(perms) == 0 {
		return nil, fmt.Errorf("no permissions defined; cannot create SUPER_ADMIN role for tenant %s", tenantID)
	}

	permKeys := make([]string, 0, len(perms))
	for _, p := range perms {
		if p == nil || p.Name == "" {
			return nil, fmt.Errorf("encountered permission with empty name while creating SUPER_ADMIN role for tenant %s", tenantID)
		}
		permKeys = append(permKeys, p.Name)
	}

	now := time.Now()
	roleModel := &domain.Role{
		TenantID:     tenantID,
		OrgID:        nil, // tenant-wide role
		Name:         "SUPER_ADMIN",
		Description:  "Super Administrator with full access to all resources",
		Permissions:  permKeys,
		IsSystemRole: true,
		CreatedBy:    createdBy,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	createdRole, err := s.roleRepo.Create(ctx, roleModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create super admin role: %w", err)
	}

	return createdRole, nil
}

// CreateRole creates a new role with name + permissions for a tenant/org
func (s *userService) CreateRole(ctx context.Context, role *domain.Role) (*domain.Role, error) {
	created, err := s.roleRepo.Create(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}
	return created, nil
}

// GetRole retrieves a role by ID
func (s *userService) GetRole(ctx context.Context, roleID uuid.UUID) (*domain.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return role, nil
}

// ListRolesByTenant lists all roles for a tenant
func (s *userService) ListRolesByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error) {
	roles, err := s.roleRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	return roles, nil
}

// ListRolesByOrganization lists roles for a tenant+org, optionally including system roles
func (s *userService) ListRolesByOrganization(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID, includeSystem bool) ([]*domain.Role, error) {
	if includeSystem {
		roles, err := s.roleRepo.ListByTenantAndOrgIncludingSystem(ctx, tenantID, orgID)
		if err != nil {
			return nil, fmt.Errorf("failed to list roles: %w", err)
		}
		return roles, nil
	}

	// Without system roles: filter out system/global roles from combined result
	roles, err := s.roleRepo.ListByTenantAndOrgIncludingSystem(ctx, tenantID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	filtered := make([]*domain.Role, 0, len(roles))
	for _, r := range roles {
		if r.IsSystemRole {
			continue
		}
		// Ensure role is bound to this org only
		if orgID != nil && r.OrgID != nil && r.OrgID.String() == orgID.String() {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

// UpdateRole updates role name/description/permissions
func (s *userService) UpdateRole(ctx context.Context, role *domain.Role) (*domain.Role, error) {
	updated, err := s.roleRepo.Update(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}
	return updated, nil
}

// DeleteRole deletes a role by ID
func (s *userService) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	// Load role first to enforce domain rules around system roles
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Prevent deletion of system roles, especially SUPER_ADMIN
	if role.IsSystemRole || role.Name == "SUPER_ADMIN" {
		return fmt.Errorf("cannot delete system role %s", role.Name)
	}

	if err := s.roleRepo.Delete(ctx, roleID); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

// ListPermissions returns the fixed permission catalog, optionally filtered by module
func (s *userService) ListPermissions(ctx context.Context, module *string) ([]*domain.Permission, error) {
	if module != nil && *module != "" {
		perms, err := s.permissionRepo.ListByModule(ctx, module)
		if err != nil {
			return nil, fmt.Errorf("failed to list permissions by module: %w", err)
		}
		return perms, nil
	}

	perms, err := s.permissionRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	return perms, nil
}

// DeactivateUser implements soft delete for users
func (s *userService) DeactivateUser(ctx context.Context, userID, deactivatedBy uuid.UUID, reason string) (*domain.User, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Fetch the admin user who is deactivating
	adminUser, err := s.userRepo.GetByID(ctx, deactivatedBy)
	var adminName *string
	if err == nil && adminUser != nil {
		name := adminUser.Name
		adminName = &name
	} else {
		// Log warning but proceed
		fmt.Printf("⚠️ Failed to fetch info for deactivator %s: %v\n", deactivatedBy, err)
	}

	// Mark as inactive
	user.IsActive = false
	now := time.Now()
	user.DeactivatedAt = &now
	user.DeactivatedBy = &deactivatedBy
	user.DeactivatedByName = adminName
	

	// Update user
	updatedUser, err := s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate user: %w", err)
	}

	// TODO: Create notification for super admin
	// This would be a call to notification service

	// TODO: Log activity
	// This would be a call to activity log service

	return updatedUser, nil
}

// ReactivateUser reactivates a deactivated user
func (s *userService) ReactivateUser(ctx context.Context, userID, reactivatedBy uuid.UUID) (*domain.User, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Mark as active
	user.IsActive = true
	user.DeactivatedAt = nil
	user.DeactivatedBy = nil

	// Update user
	updatedUser, err := s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to reactivate user: %w", err)
	}

	// TODO: Log activity

	return updatedUser, nil
}

// CreateActivityLog creates an activity log entry
func (s *userService) CreateActivityLog(ctx context.Context, log *domain.ActivityLog) (*domain.ActivityLog, error) {
	created, err := s.activityLogRepo.Create(ctx, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create activity log: %w", err)
	}
	fmt.Printf("Activity Log: %s - %s\n", created.Name, created.Description)
	return created, nil
}

// ListActivityLogs lists activity logs
func (s *userService) ListActivityLogs(ctx context.Context, limit, offset int32) ([]*domain.ActivityLog, error) {
	logs, err := s.activityLogRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list activity logs: %w", err)
	}
	return logs, nil
}

// CreateNotification creates a notification
func (s *userService) CreateNotification(ctx context.Context, notification *domain.Notification) (*domain.Notification, error) {
	// TODO: Implement notification repository
	notification.NotificationID = uuid.New()
	notification.CreatedAt = time.Now()
	notification.IsRead = false

	// Mock implementation for now
	fmt.Printf("Notification: %s - %s\n", notification.Title, notification.Message)

	return notification, nil
}

// ListNotifications lists notifications for a user
func (s *userService) ListNotifications(ctx context.Context, userID uuid.UUID, unreadOnly bool, limit, offset int32) ([]*domain.Notification, error) {
	// TODO: Implement notification repository
	// Mock implementation
	return []*domain.Notification{}, nil
}

// MarkNotificationAsRead marks a notification as read
func (s *userService) MarkNotificationAsRead(ctx context.Context, notificationID uuid.UUID) (*domain.Notification, error) {
	// TODO: Implement notification repository
	// Mock implementation
	now := time.Now()
	return &domain.Notification{
		NotificationID: notificationID,
		IsRead:         true,
		ReadAt:         &now,
	}, nil
}

// CreateLoginHistory creates a login history entry
func (s *userService) CreateLoginHistory(ctx context.Context, history *domain.UserLoginHistory) (*domain.UserLoginHistory, error) {
	created, err := s.loginHistoryRepo.Create(ctx, history)
	if err != nil {
		return nil, fmt.Errorf("failed to create login history: %w", err)
	}
	fmt.Printf("Login History: User %s logged in from %s\n", history.UserID, history.IPAddress)

	// Also update the user's last_login_at and last_login_ip in the users table
	ip := ""
	if history.IPAddress != nil {
		ip = *history.IPAddress
	}
	ua := ""
	if history.UserAgent != nil {
		ua = *history.UserAgent
	}
	
	if err := s.userRepo.UpdateLastLogin(ctx, history.UserID, ip, ua); err != nil {
		// Log error but don't fail the history creation
		fmt.Printf("⚠️ Failed to update user last_login: %v\n", err)
	}

	return created, nil
}

// ListLoginHistory lists login history for a user
func (s *userService) ListLoginHistory(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*domain.UserLoginHistory, error) {
	histories, err := s.loginHistoryRepo.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list login history: %w", err)
	}
	return histories, nil
}

// CreateTenant creates a new tenant with super admin user
func (s *userService) CreateTenant(ctx context.Context, name, email, password, role string) (*domain.Tenant, error) {
	// Generate tenant ID
	tenantID := uuid.New()
	now := time.Now()

	// Hash password for super admin
	hashedPassword, err := s.hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create tenant
	tenant := &domain.Tenant{
		TenantID:  tenantID,
		Name:      name,
		Email:     email,
		Password:  hashedPassword,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if _, err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant record: %w", err)
	}

	// Create super admin user for this tenant
	superAdmin := &domain.User{
		UserID:          uuid.New(),
		TenantID:        tenantID,
		Name:            name,
		Email:           email,
		Password:        hashedPassword,
		EmailVerifiedAt: &now,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Store super admin user (tenant creation creates the first user)
	createdSuperAdmin, err := s.userRepo.Create(ctx, superAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to create super admin user: %w", err)
	}

	// Ensure SUPER_ADMIN role with full permissions exists for this tenant
	createdRole, err := s.ensureSuperAdminRoleForTenant(ctx, tenantID, name)
	if err != nil {
		return nil, err
	}

	// Assign SUPER_ADMIN role to the initial super admin user
	if err := s.userRoleRepo.AssignRole(ctx, createdSuperAdmin.UserID, createdRole.RoleID); err != nil {
		return nil, fmt.Errorf("failed to assign super admin role to user: %w", err)
	}

	fmt.Printf("Created tenant %s with super admin user %s and SUPER_ADMIN role %s\n", tenantID, createdSuperAdmin.UserID, createdRole.RoleID)

	return tenant, nil
}

// GetTenant retrieves tenant information by tenant ID
func (s *userService) GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// hashPassword hashes a password using bcrypt
func (s *userService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
