package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
)

type AuthUser struct {
	ID    uuid.UUID
	Name  string
	Email string
	Roles []string
}

// OrganizationRepository defines the interface for Organization operations.
type OrganizationRepository interface {
	Create(ctx context.Context, tenantID uuid.UUID, name string) (db.Organization, error)
	Get(ctx context.Context, orgID uuid.UUID) (db.Organization, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]db.Organization, error)
	Update(ctx context.Context, orgID uuid.UUID, name string) (db.Organization, error)
	Delete(ctx context.Context, orgID uuid.UUID) error
}

type PasswordResetRepository interface {
	Create(ctx context.Context, email string, token uuid.UUID, expiresAt time.Time) (db.PasswordReset, error)
	GetByToken(ctx context.Context, token uuid.UUID) (db.PasswordReset, error)  // <-- add this
	Delete(ctx context.Context, token uuid.UUID) error
}

// RoleRepository defines role-related database operations.
type RoleRepository interface {
    Create(ctx context.Context, arg db.CreateRoleParams) (db.Role, error)
    Get(ctx context.Context, roleID uuid.UUID) (db.Role, error)
    List(ctx context.Context, tenantID uuid.UUID) ([]db.Role, error)
    Update(ctx context.Context, arg db.UpdateRoleParams) (db.Role, error)
    Delete(ctx context.Context, roleID uuid.UUID) error
    AssignRoleToUser(ctx context.Context, arg db.AssignRoleToUserParams) error
    AssignPermissionToRole(ctx context.Context, arg db.AssignPermissionToRoleParams) error
    ListRolesOfUser(ctx context.Context, userID uuid.UUID) ([]db.Role, error)
    ListPermissionsOfUserViaRoles(ctx context.Context, userID uuid.UUID) ([]db.Permission, error)
    GetByEmail(ctx context.Context, email string) (db.User, error) // optional if needed
	ListSuperAdmins(ctx context.Context) ([]db.User, error)
}

type SessionRepository interface {
    Create(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (db.Session, error)
    Get(ctx context.Context, sessionID uuid.UUID) (db.Session, error)
    Delete(ctx context.Context, token string) error
	GetByToken(ctx context.Context, token string) (db.Session, error)
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
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]db.User, error)
	Update(ctx context.Context, user db.UpdateUserParams) (db.User, error)
	GetByEmail(ctx context.Context, email string) (db.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) (db.User, error)
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error
	ConfirmPassword(ctx context.Context, userID uuid.UUID, password string) (bool, error)
	GetUserByToken(ctx context.Context, token string) (db.User, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
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

type EmailVerificationRepository interface {
	Insert(ctx context.Context, userID uuid.UUID, token uuid.UUID) (db.EmailVerification, error)
	GetByToken(ctx context.Context, token uuid.UUID) (db.EmailVerification, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type EmailSender interface {
	SendVerificationEmail(ctx context.Context, to string, link string, expiresAt string) error
	SendResetPasswordEmail(ctx context.Context, to, link, expiresAt string) error
}

type DepartmentRepository interface {
	Create(ctx context.Context, name, description string) (db.Department, error)
	Get(ctx context.Context, id string) (db.Department, error)
	List(ctx context.Context, limit, offset int32) ([]db.Department, error)
	Update(ctx context.Context, id, name, description string) (db.Department, error)
	Delete(ctx context.Context, id string) error
}

type DepartmentService interface {
	Create(ctx context.Context, name, description string) (db.Department, error)
	Get(ctx context.Context, id string) (db.Department, error)
	List(ctx context.Context, page, pageSize int32) ([]db.Department, error)
	Update(ctx context.Context, id, name, description string) (db.Department, error)
	Delete(ctx context.Context, id string) error
}

type DesignationRepository interface {
	Create(ctx context.Context, d db.Designation) (db.Designation, error)
	Get(ctx context.Context, id uuid.UUID) (db.Designation, error)
	Update(ctx context.Context, d db.Designation) (db.Designation, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int32) ([]db.Designation, error)
}