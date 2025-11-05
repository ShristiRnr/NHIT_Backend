package domain

import (
	"time"

	"github.com/google/uuid"
)

// Organization represents an organization in the system
type Organization struct {
	OrgID     uuid.UUID
	TenantID  uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserOrganization represents the association between users and organizations
type UserOrganization struct {
	UserID uuid.UUID
	OrgID  uuid.UUID
	RoleID uuid.UUID
}

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	TenantID         uuid.UUID
	Name             string
	SuperAdminUserID *uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
