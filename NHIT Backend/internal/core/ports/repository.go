package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
)

// OrganizationRepository defines the interface for Organization operations.
type OrganizationRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, name string) (db.Organization, error)
	Get(ctx context.Context, orgID uuid.UUID) (db.Organization, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]db.Organization, error)
	Update(ctx context.Context, orgID uuid.UUID, name string) (db.Organization, error)
	Delete(ctx context.Context, orgID uuid.UUID) error
}

type PasswordResetRepository interface {
	Create(ctx context.Context, token db.CreatePasswordResetTokenParams) (db.PasswordReset, error)
	Get(ctx context.Context, token uuid.UUID) (db.PasswordReset, error)
	Delete(ctx context.Context, token uuid.UUID) error
}

// RoleRepository defines role-related database operations.
type RoleRepository interface {
	Create(ctx context.Context, arg db.CreateRoleParams) (db.Role, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]db.Role, error)
	Get(ctx context.Context, roleID uuid.UUID) (db.Role, error)
	Update(ctx context.Context, arg db.UpdateRoleParams) (db.Role, error)
	Delete(ctx context.Context, roleID uuid.UUID) error
	AssignRoleToUser(ctx context.Context, arg db.AssignRoleToUserParams) error
	AssignPermissionToRole(ctx context.Context, arg db.AssignPermissionToRoleParams) error
	ListRolesOfUser(ctx context.Context, userID uuid.UUID) ([]db.Role, error)
	ListPermissionsOfUserViaRoles(ctx context.Context, userID uuid.UUID) ([]db.Permission, error)
}

type SessionRepository interface {
    Create(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (db.Session, error)
    Get(ctx context.Context, sessionID uuid.UUID) (db.Session, error)
    Delete(ctx context.Context, sessionID uuid.UUID) error
}

type TenantRepository interface {
	Create(ctx context.Context, tenant db.CreateTenantParams) (db.Tenant, error)
	Get(ctx context.Context, tenantID uuid.UUID) (db.Tenant, error)
}

type UserLoginRepository interface {
	Create(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) (db.UserLoginHistory, error)
	List(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]db.UserLoginHistory, error)
}

type UserOrganizationRepository interface {
	AddUserToOrganization(ctx context.Context, arg db.AddUserToOrganizationParams) error
	ListUsersByOrganization(ctx context.Context, orgID uuid.UUID) ([]db.ListUsersByOrganizationRow, error)
}

type UserRepository interface {
	Create(ctx context.Context, user db.CreateUserParams) (db.User, error)
	Delete(ctx context.Context, userID uuid.UUID) error
	Get(ctx context.Context, userID uuid.UUID) (db.User, error)
	GetRolesAndPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit int32, offset int32) ([]db.User, error)
	Update(ctx context.Context, user db.UpdateUserParams) (db.User, error)
}

type PaginationRepository interface {
	ListPaginated(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]db.User, error)
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	GetUserIDByToken(ctx context.Context, token string) (uuid.UUID, error)
	Delete(ctx context.Context, token string) error
}

type UserRoleRepository interface {
	AssignRole(ctx context.Context, params db.AssignRoleToUserParams) error
	ListRoles(ctx context.Context, userID uuid.UUID) ([]db.ListRolesForUserRow, error)
}


