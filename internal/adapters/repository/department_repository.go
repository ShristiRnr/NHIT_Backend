package repository

import (
	"context"

	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
	"github.com/google/uuid"
)

type DepartmentRepository struct {
	q *db.Queries
}

func NewDepartmentRepository(q *db.Queries) ports.DepartmentRepository {
	return &DepartmentRepository{q: q}
}

func (r *DepartmentRepository) Create(ctx context.Context, name, description string) (db.Department, error) {
	dept, err := r.q.CreateDepartment(ctx, db.CreateDepartmentParams{
		Name:        name,
		Description: description,
	})
	if err != nil {
		return db.Department{}, err
	}
	return dept, nil
}

func (r *DepartmentRepository) Get(ctx context.Context, id string) (db.Department, error) {
	return r.q.GetDepartment(ctx, uuid.MustParse(id))
}

func (r *DepartmentRepository) List(ctx context.Context, limit, offset int32) ([]db.Department, error) {
	return r.q.ListDepartments(ctx, db.ListDepartmentsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *DepartmentRepository) Update(ctx context.Context, id, name, description string) (db.Department, error) {
	dept, err := r.q.UpdateDepartment(ctx, db.UpdateDepartmentParams{
		ID:          uuid.MustParse(id),
		Name:        name,
		Description: description,
	})
	if err != nil {
		return db.Department{}, err
	}
	return dept, nil
}

func (r *DepartmentRepository) Delete(ctx context.Context, id string) error {
	return r.q.DeleteDepartment(ctx, uuid.MustParse(id))
}
