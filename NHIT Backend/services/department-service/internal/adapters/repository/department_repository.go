package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/core/ports"
)

type departmentRepository struct {
	queries *sqlc.Queries
}

// NewDepartmentRepository creates a new department repository
func NewDepartmentRepository(queries *sqlc.Queries) ports.DepartmentRepository {
	return &departmentRepository{
		queries: queries,
	}
}

// Create creates a new department
func (r *departmentRepository) Create(ctx context.Context, department *domain.Department) (*domain.Department, error) {
	dbDept, err := r.queries.CreateDepartment(ctx, sqlc.CreateDepartmentParams{
		Name:        department.Name,
		Description: department.Description,
	})
	if err != nil {
		return nil, err
	}

	return dbToDomain(dbDept), nil
}

// GetByID retrieves a department by ID
func (r *departmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Department, error) {
	dbDept, err := r.queries.GetDepartmentByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrDepartmentNotFound
		}
		return nil, err
	}

	return dbToDomain(dbDept), nil
}

// GetByName retrieves a department by name
func (r *departmentRepository) GetByName(ctx context.Context, name string) (*domain.Department, error) {
	dbDept, err := r.queries.GetDepartmentByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrDepartmentNotFound
		}
		return nil, err
	}

	return dbToDomain(dbDept), nil
}

// Update updates a department
func (r *departmentRepository) Update(ctx context.Context, department *domain.Department) (*domain.Department, error) {
	dbDept, err := r.queries.UpdateDepartment(ctx, sqlc.UpdateDepartmentParams{
		ID:          department.ID,
		Name:        department.Name,
		Description: department.Description,
	})
	if err != nil {
		return nil, err
	}

	return dbToDomain(dbDept), nil
}

// Delete deletes a department
func (r *departmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteDepartment(ctx, id)
}

// List retrieves departments with pagination
func (r *departmentRepository) List(ctx context.Context, page, pageSize int32) ([]*domain.Department, int32, error) {
	offset := (page - 1) * pageSize

	// Get departments
	dbDepts, err := r.queries.ListDepartments(ctx, sqlc.ListDepartmentsParams{
		Limit:  pageSize,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	total, err := r.queries.CountDepartments(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Convert to domain
	departments := make([]*domain.Department, len(dbDepts))
	for i, dbDept := range dbDepts {
		departments[i] = dbToDomain(dbDept)
	}

	return departments, int32(total), nil
}

// Exists checks if a department exists by name
func (r *departmentRepository) Exists(ctx context.Context, name string) (bool, error) {
	return r.queries.DepartmentExists(ctx, name)
}

// ExistsByID checks if a department exists by ID
func (r *departmentRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	return r.queries.DepartmentExistsByID(ctx, id)
}

// CountUsersByDepartment counts users in a department
func (r *departmentRepository) CountUsersByDepartment(ctx context.Context, departmentID uuid.UUID) (int32, error) {
	// TODO: Add CountUsersByDepartment query to SQLC
	// This requires a join with the users table which is in a different service
	// For now, return 0
	return 0, nil
}

// dbToDomain converts database model to domain model
func dbToDomain(dbDept *sqlc.Department) *domain.Department {
	return &domain.Department{
		ID:          dbDept.ID,
		Name:        dbDept.Name,
		Description: dbDept.Description,
		CreatedAt:   dbDept.CreatedAt.Time,
		UpdatedAt:   dbDept.UpdatedAt.Time,
	}
}

