package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// UserOrganizationRepo implements UserOrganizationRepository using sqlc-generated db.
type UserOrganizationRepo struct {
	q *db.Queries
}

// NewUserOrganizationRepo creates a new repository instance.
func NewUserOrganizationRepo(q *db.Queries) ports.UserOrganizationRepository {
	return &UserOrganizationRepo{q: q}
}

// AddUserToOrganization adds a user to an organization with a role using SQLC params.
func (r *UserOrganizationRepo) AddUserToOrganization(ctx context.Context, arg db.AddUserToOrganizationParams) error {
	return r.q.AddUserToOrganization(ctx, arg)
}

// ListUsersByOrganization lists all users assigned to a specific organization.
func (r *UserOrganizationRepo) ListUsersByOrganization(ctx context.Context, orgID uuid.UUID) ([]db.ListUsersByOrganizationRow, error) {
	return r.q.ListUsersByOrganization(ctx, orgID)
}
