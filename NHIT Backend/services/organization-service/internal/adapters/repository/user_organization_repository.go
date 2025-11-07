package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/ports"
)

type userOrganizationRepository struct {
	db *sql.DB
}

// NewUserOrganizationRepository creates a new PostgreSQL user-organization repository
func NewUserOrganizationRepository(db *sql.DB) ports.UserOrganizationRepository {
	return &userOrganizationRepository{db: db}
}

// AddUserToOrganization adds a user to an organization with a specific role
func (r *userOrganizationRepository) AddUserToOrganization(
	ctx context.Context,
	userOrg *domain.UserOrganization,
) error {
	query := `
		INSERT INTO user_organizations (
			user_id, org_id, role_id, is_current_context, 
			joined_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`
	
	_, err := r.db.ExecContext(
		ctx, query,
		userOrg.UserID, userOrg.OrgID, userOrg.RoleID,
		userOrg.IsCurrentContext, userOrg.JoinedAt, userOrg.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to add user to organization: %w", err)
	}
	
	return nil
}

// RemoveUserFromOrganization removes a user from an organization
func (r *userOrganizationRepository) RemoveUserFromOrganization(
	ctx context.Context,
	userID, orgID uuid.UUID,
) error {
	query := `DELETE FROM user_organizations WHERE user_id = $1 AND org_id = $2`
	
	result, err := r.db.ExecContext(ctx, query, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to remove user from organization: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user-organization relationship not found")
	}
	
	return nil
}

// ListUsersByOrganization retrieves all user IDs in an organization
func (r *userOrganizationRepository) ListUsersByOrganization(
	ctx context.Context,
	orgID uuid.UUID,
) ([]uuid.UUID, error) {
	query := `SELECT user_id FROM user_organizations WHERE org_id = $1`
	
	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list users by organization: %w", err)
	}
	defer rows.Close()
	
	var userIDs []uuid.UUID
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}
	
	return userIDs, nil
}

// ListOrganizationsByUser retrieves all organizations for a user
func (r *userOrganizationRepository) ListOrganizationsByUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]*domain.Organization, error) {
	query := `
		SELECT o.org_id, o.tenant_id, o.name, o.code, o.database_name, 
			o.description, o.logo, o.is_active, o.created_by, 
			o.created_at, o.updated_at
		FROM organizations o
		INNER JOIN user_organizations uo ON o.org_id = uo.org_id
		WHERE uo.user_id = $1
		ORDER BY o.name
	`
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations by user: %w", err)
	}
	defer rows.Close()
	
	var orgs []*domain.Organization
	for rows.Next() {
		var org domain.Organization
		err := rows.Scan(
			&org.OrgID, &org.TenantID, &org.Name, &org.Code, &org.DatabaseName,
			&org.Description, &org.Logo, &org.IsActive, &org.CreatedBy,
			&org.CreatedAt, &org.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		orgs = append(orgs, &org)
	}
	
	return orgs, nil
}

// GetUserOrganization retrieves a specific user-organization relationship
func (r *userOrganizationRepository) GetUserOrganization(
	ctx context.Context,
	userID, orgID uuid.UUID,
) (*domain.UserOrganization, error) {
	query := `
		SELECT user_id, org_id, role_id, is_current_context, 
			joined_at, updated_at
		FROM user_organizations
		WHERE user_id = $1 AND org_id = $2
	`
	
	var userOrg domain.UserOrganization
	err := r.db.QueryRowContext(ctx, query, userID, orgID).Scan(
		&userOrg.UserID, &userOrg.OrgID, &userOrg.RoleID,
		&userOrg.IsCurrentContext, &userOrg.JoinedAt, &userOrg.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user-organization: %w", err)
	}
	
	return &userOrg, nil
}

// UpdateUserOrganization updates a user-organization relationship
func (r *userOrganizationRepository) UpdateUserOrganization(
	ctx context.Context,
	userOrg *domain.UserOrganization,
) error {
	query := `
		UPDATE user_organizations
		SET role_id = $1, is_current_context = $2, updated_at = $3
		WHERE user_id = $4 AND org_id = $5
	`
	
	result, err := r.db.ExecContext(
		ctx, query,
		userOrg.RoleID, userOrg.IsCurrentContext, userOrg.UpdatedAt,
		userOrg.UserID, userOrg.OrgID,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update user-organization: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user-organization relationship not found")
	}
	
	return nil
}

// SetCurrentOrganization sets an organization as the current context for a user
func (r *userOrganizationRepository) SetCurrentOrganization(
	ctx context.Context,
	userID, orgID uuid.UUID,
) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// First, unset all current contexts for the user
	query1 := `
		UPDATE user_organizations
		SET is_current_context = false, updated_at = NOW()
		WHERE user_id = $1
	`
	
	_, err = tx.ExecContext(ctx, query1, userID)
	if err != nil {
		return fmt.Errorf("failed to unset current contexts: %w", err)
	}
	
	// Then, set the new current context
	query2 := `
		UPDATE user_organizations
		SET is_current_context = true, updated_at = NOW()
		WHERE user_id = $1 AND org_id = $2
	`
	
	result, err := tx.ExecContext(ctx, query2, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to set current context: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user-organization relationship not found")
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// GetCurrentOrganization retrieves the current organization for a user
func (r *userOrganizationRepository) GetCurrentOrganization(
	ctx context.Context,
	userID uuid.UUID,
) (*domain.Organization, error) {
	query := `
		SELECT o.org_id, o.tenant_id, o.name, o.code, o.database_name, 
			o.description, o.logo, o.is_active, o.created_by, 
			o.created_at, o.updated_at
		FROM organizations o
		INNER JOIN user_organizations uo ON o.org_id = uo.org_id
		WHERE uo.user_id = $1 AND uo.is_current_context = true
	`
	
	var org domain.Organization
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&org.OrgID, &org.TenantID, &org.Name, &org.Code, &org.DatabaseName,
		&org.Description, &org.Logo, &org.IsActive, &org.CreatedBy,
		&org.CreatedAt, &org.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get current organization: %w", err)
	}
	
	return &org, nil
}

// UserHasAccessToOrganization checks if a user has access to an organization
func (r *userOrganizationRepository) UserHasAccessToOrganization(
	ctx context.Context,
	userID, orgID uuid.UUID,
) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_organizations 
			WHERE user_id = $1 AND org_id = $2
		)
	`
	
	var hasAccess bool
	err := r.db.QueryRowContext(ctx, query, userID, orgID).Scan(&hasAccess)
	if err != nil {
		return false, fmt.Errorf("failed to check user access: %w", err)
	}
	
	return hasAccess, nil
}

// CountOrganizationsByUser counts organizations for a user
func (r *userOrganizationRepository) CountOrganizationsByUser(
	ctx context.Context,
	userID uuid.UUID,
) (int32, error) {
	query := `SELECT COUNT(*) FROM user_organizations WHERE user_id = $1`
	
	var count int32
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count organizations by user: %w", err)
	}
	
	return count, nil
}
