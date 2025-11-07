package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/ports"
)

type organizationRepository struct {
	db *sql.DB
}

// NewOrganizationRepository creates a new PostgreSQL organization repository
func NewOrganizationRepository(db *sql.DB) ports.OrganizationRepository {
	return &organizationRepository{db: db}
}

// Create creates a new organization
func (r *organizationRepository) Create(ctx context.Context, org *domain.Organization) (*domain.Organization, error) {
	query := `
		INSERT INTO organizations (
			org_id, tenant_id, name, code, database_name, 
			description, logo, is_active, created_by, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING org_id, tenant_id, name, code, database_name, 
			description, logo, is_active, created_by, 
			created_at, updated_at
	`
	
	var createdOrg domain.Organization
	err := r.db.QueryRowContext(
		ctx, query,
		org.OrgID, org.TenantID, org.Name, org.Code, org.DatabaseName,
		org.Description, org.Logo, org.IsActive, org.CreatedBy,
		org.CreatedAt, org.UpdatedAt,
	).Scan(
		&createdOrg.OrgID, &createdOrg.TenantID, &createdOrg.Name, 
		&createdOrg.Code, &createdOrg.DatabaseName,
		&createdOrg.Description, &createdOrg.Logo, &createdOrg.IsActive, 
		&createdOrg.CreatedBy, &createdOrg.CreatedAt, &createdOrg.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}
	
	return &createdOrg, nil
}

// GetByID retrieves an organization by ID
func (r *organizationRepository) GetByID(ctx context.Context, orgID uuid.UUID) (*domain.Organization, error) {
	query := `
		SELECT org_id, tenant_id, name, code, database_name, 
			description, logo, is_active, created_by, 
			created_at, updated_at
		FROM organizations
		WHERE org_id = $1
	`
	
	var org domain.Organization
	err := r.db.QueryRowContext(ctx, query, orgID).Scan(
		&org.OrgID, &org.TenantID, &org.Name, &org.Code, &org.DatabaseName,
		&org.Description, &org.Logo, &org.IsActive, &org.CreatedBy,
		&org.CreatedAt, &org.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get organization by ID: %w", err)
	}
	
	return &org, nil
}

// GetByCode retrieves an organization by code
func (r *organizationRepository) GetByCode(ctx context.Context, code string) (*domain.Organization, error) {
	query := `
		SELECT org_id, tenant_id, name, code, database_name, 
			description, logo, is_active, created_by, 
			created_at, updated_at
		FROM organizations
		WHERE code = $1
	`
	
	var org domain.Organization
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&org.OrgID, &org.TenantID, &org.Name, &org.Code, &org.DatabaseName,
		&org.Description, &org.Logo, &org.IsActive, &org.CreatedBy,
		&org.CreatedAt, &org.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get organization by code: %w", err)
	}
	
	return &org, nil
}

// Update updates an existing organization
func (r *organizationRepository) Update(ctx context.Context, org *domain.Organization) (*domain.Organization, error) {
	query := `
		UPDATE organizations
		SET name = $1, code = $2, database_name = $3, description = $4, 
			logo = $5, is_active = $6, updated_at = $7
		WHERE org_id = $8
		RETURNING org_id, tenant_id, name, code, database_name, 
			description, logo, is_active, created_by, 
			created_at, updated_at
	`
	
	var updatedOrg domain.Organization
	err := r.db.QueryRowContext(
		ctx, query,
		org.Name, org.Code, org.DatabaseName, org.Description,
		org.Logo, org.IsActive, org.UpdatedAt,
		org.OrgID,
	).Scan(
		&updatedOrg.OrgID, &updatedOrg.TenantID, &updatedOrg.Name, 
		&updatedOrg.Code, &updatedOrg.DatabaseName,
		&updatedOrg.Description, &updatedOrg.Logo, &updatedOrg.IsActive, 
		&updatedOrg.CreatedBy, &updatedOrg.CreatedAt, &updatedOrg.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}
	
	return &updatedOrg, nil
}

// Delete deletes an organization by ID
func (r *organizationRepository) Delete(ctx context.Context, orgID uuid.UUID) error {
	query := `DELETE FROM organizations WHERE org_id = $1`
	
	result, err := r.db.ExecContext(ctx, query, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("organization not found")
	}
	
	return nil
}

// ListByTenant retrieves all organizations for a tenant with pagination
func (r *organizationRepository) ListByTenant(
	ctx context.Context,
	tenantID uuid.UUID,
	pagination ports.PaginationParams,
) ([]*domain.Organization, *ports.PaginationResult, error) {
	// Calculate offset
	offset := (pagination.Page - 1) * pagination.PageSize
	
	// Query for organizations
	query := `
		SELECT org_id, tenant_id, name, code, database_name, 
			description, logo, is_active, created_by, 
			created_at, updated_at
		FROM organizations
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.QueryContext(ctx, query, tenantID, pagination.PageSize, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list organizations by tenant: %w", err)
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
			return nil, nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		orgs = append(orgs, &org)
	}
	
	// Count total items
	totalCount, err := r.CountByTenant(ctx, tenantID)
	if err != nil {
		return nil, nil, err
	}
	
	// Calculate pagination metadata
	totalPages := int32(math.Ceil(float64(totalCount) / float64(pagination.PageSize)))
	
	paginationResult := &ports.PaginationResult{
		CurrentPage: pagination.Page,
		PageSize:    pagination.PageSize,
		TotalItems:  totalCount,
		TotalPages:  totalPages,
	}
	
	return orgs, paginationResult, nil
}

// ListAccessibleByUser retrieves all organizations accessible by a user with pagination
func (r *organizationRepository) ListAccessibleByUser(
	ctx context.Context,
	userID uuid.UUID,
	pagination ports.PaginationParams,
) ([]*domain.Organization, *ports.PaginationResult, error) {
	// Calculate offset
	offset := (pagination.Page - 1) * pagination.PageSize
	
	// Query for organizations through user_organizations junction table
	query := `
		SELECT o.org_id, o.tenant_id, o.name, o.code, o.database_name, 
			o.description, o.logo, o.is_active, o.created_by, 
			o.created_at, o.updated_at
		FROM organizations o
		INNER JOIN user_organizations uo ON o.org_id = uo.org_id
		WHERE uo.user_id = $1 AND o.is_active = true
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.QueryContext(ctx, query, userID, pagination.PageSize, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list accessible organizations: %w", err)
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
			return nil, nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		orgs = append(orgs, &org)
	}
	
	// Count total items
	countQuery := `
		SELECT COUNT(*)
		FROM organizations o
		INNER JOIN user_organizations uo ON o.org_id = uo.org_id
		WHERE uo.user_id = $1 AND o.is_active = true
	`
	
	var totalCount int32
	err = r.db.QueryRowContext(ctx, countQuery, userID).Scan(&totalCount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count accessible organizations: %w", err)
	}
	
	// Calculate pagination metadata
	totalPages := int32(math.Ceil(float64(totalCount) / float64(pagination.PageSize)))
	
	paginationResult := &ports.PaginationResult{
		CurrentPage: pagination.Page,
		PageSize:    pagination.PageSize,
		TotalItems:  totalCount,
		TotalPages:  totalPages,
	}
	
	return orgs, paginationResult, nil
}

// CodeExists checks if an organization code already exists
func (r *organizationRepository) CodeExists(
	ctx context.Context,
	code string,
	excludeOrgID *uuid.UUID,
) (bool, error) {
	var query string
	var args []interface{}
	
	if excludeOrgID != nil {
		query = `SELECT EXISTS(SELECT 1 FROM organizations WHERE code = $1 AND org_id != $2)`
		args = []interface{}{code, *excludeOrgID}
	} else {
		query = `SELECT EXISTS(SELECT 1 FROM organizations WHERE code = $1)`
		args = []interface{}{code}
	}
	
	var exists bool
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check code existence: %w", err)
	}
	
	return exists, nil
}

// CountByTenant counts organizations for a tenant
func (r *organizationRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int32, error) {
	query := `SELECT COUNT(*) FROM organizations WHERE tenant_id = $1`
	
	var count int32
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count organizations by tenant: %w", err)
	}
	
	return count, nil
}
