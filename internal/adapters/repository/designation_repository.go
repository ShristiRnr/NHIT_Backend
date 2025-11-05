package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
)

// DepartmentRepository implements domain.DepartmentRepository
type DesignationRepository struct {
	q *db.Queries
}

func NewDesignationRepository(q *db.Queries) *DesignationRepository {
	return &DesignationRepository{q: q}
}

func (r *DesignationRepository) Create(ctx context.Context, d db.Designation) (db.Designation, error) {
	created, err := r.q.CreateDesignation(ctx, db.CreateDesignationParams{
		Name:        d.Name,
		Description: d.Description,
	})
	if err != nil {
		return db.Designation{}, err
	}
	return dbToDomain(created), nil
}

func (r *DesignationRepository) Get(ctx context.Context, id uuid.UUID) (db.Designation, error) {
	got, err := r.q.GetDesignation(ctx, id)
	if err != nil {
		return db.Designation{}, err
	}
	return dbToDomain(got), nil
}

func (r *DesignationRepository) Update(ctx context.Context, d db.Designation) (db.Designation, error) {
	updated, err := r.q.UpdateDesignation(ctx, db.UpdateDesignationParams{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
	})
	if err != nil {
		return db.Designation{}, err
	}
	return dbToDomain(updated), nil
}

func (r *DesignationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteDesignation(ctx, id)
}

func (r *DesignationRepository) List(ctx context.Context, limit, offset int32) ([]db.Designation, error) {
	items, err := r.q.ListDesignations(ctx, db.ListDesignationsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	var result []db.Designation
	for _, d := range items {
		result = append(result, dbToDomain(d))
	}
	return result, nil
}

// Helper to map db.Designation → domain.Designation
func dbToDomain(d db.Designation) db.Designation {
	return db.Designation{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}
