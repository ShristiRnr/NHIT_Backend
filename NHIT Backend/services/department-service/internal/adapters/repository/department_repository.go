package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/core/ports"
)

type departmentRepository struct {
	queries *db.Queries
}

// NewDepartmentRepository creates a new department repository
func NewDepartmentRepository(queries *db.Queries) ports.DepartmentRepository {
	return &departmentRepository{
		queries: queries,
	}
}

// Create creates a new department
func (r *departmentRepository) Create(ctx context.Context, department *domain.Department) (*domain.Department, error) {
	dbDept, err := r.queries.CreateDepartment(ctx, db.CreateDepartmentParams{
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
	dbDept, err := r.queries.UpdateDepartment(ctx, db.UpdateDepartmentParams{
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
	dbDepts, err := r.queries.ListDepartments(ctx, db.ListDepartmentsParams{
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
	count, err := r.queries.CountUsersByDepartment(ctx, uuid.NullUUID{
		UUID:  departmentID,
		Valid: true,
	})
	if err != nil {
		return 0, err
	}
	return int32(count), nil
}

// dbToDomain converts database model to domain model
func dbToDomain(dbDept db.Department) *domain.Department {
	return &domain.Department{
		ID:          dbDept.ID,
		Name:        dbDept.Name,
		Description: dbDept.Description,
		CreatedAt:   dbDept.CreatedAt.Time,
		UpdatedAt:   dbDept.UpdatedAt.Time,
	}
}
