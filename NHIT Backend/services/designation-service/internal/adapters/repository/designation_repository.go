package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/ports"
)

type designationRepository struct {
	queries *sqlc.Queries
}

// NewDesignationRepository creates a new designation repository
func NewDesignationRepository(queries *sqlc.Queries) ports.DesignationRepository {
	return &designationRepository{
		queries: queries,
	}
}

// Create creates a new designation
func (r *designationRepository) Create(ctx context.Context, designation *domain.Designation) error {
	_, err := r.queries.CreateDesignation(ctx, sqlc.CreateDesignationParams{
		ID:          designation.ID,
		Name:        designation.Name,
		Description: designation.Description,
		CreatedAt:   pgtype.Timestamptz{Time: designation.CreatedAt, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: designation.UpdatedAt, Valid: true},
	})

	return err
}

// GetByID retrieves a designation by ID
func (r *designationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Designation, error) {
	dbDesignation, err := r.queries.GetDesignationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return toDomainDesignation(dbDesignation), nil
}

// Update updates an existing designation
func (r *designationRepository) Update(ctx context.Context, designation *domain.Designation) error {
	_, err := r.queries.UpdateDesignation(ctx, sqlc.UpdateDesignationParams{
		ID:          designation.ID,
		Name:        designation.Name,
		Description: designation.Description,
		UpdatedAt:   pgtype.Timestamptz{Time: designation.UpdatedAt, Valid: true},
	})

	return err
}

// Delete deletes a designation by ID
func (r *designationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteDesignation(ctx, id)
}

func (r *designationRepository) List(ctx context.Context, page, pageSize int32) ([]*domain.Designation, error) {
    offset := (page - 1) * pageSize

    dbDesignations, err := r.queries.ListDesignations(ctx, sqlc.ListDesignationsParams{
        Limit:  pageSize,
        Offset: offset,
    })
    if err != nil {
        return nil, err
    }

    designations := make([]*domain.Designation, len(dbDesignations))
    for i, d := range dbDesignations {
        designations[i] = toDomainDesignation(d)
    }

    return designations, nil
}

// toDomainDesignation converts a database designation to a domain designation
func toDomainDesignation(dbDesignation *sqlc.Designation) *domain.Designation {
	return &domain.Designation{
		ID:          dbDesignation.ID,
		Name:        dbDesignation.Name,
		Description: dbDesignation.Description,
		CreatedAt:   dbDesignation.CreatedAt.Time,
		UpdatedAt:   dbDesignation.UpdatedAt.Time,
	}
}

