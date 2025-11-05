package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents the core user domain model
type User struct {
	UserID          uuid.UUID
	TenantID        uuid.UUID
	Name            string
	Email           string
	Password        string
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	LastLogoutAt    *time.Time
	LastLoginIP     string
	UserAgent       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// UserRole represents user-role association
type UserRole struct {
	UserID uuid.UUID
	RoleID uuid.UUID
}

// Role represents a role in the system
type Role struct {
	RoleID      uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Permissions []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
