package repository

import (
	"context"
	"time"

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
	var parentID uuid.NullUUID
	if designation.ParentID != nil {
		parentID = uuid.NullUUID{UUID: *designation.ParentID, Valid: true}
	}

	_, err := r.queries.CreateDesignation(ctx, sqlc.CreateDesignationParams{
		ID:          designation.ID,
		Name:        designation.Name,
		Description: designation.Description,
		Slug:        designation.Slug,
		IsActive:    &designation.IsActive,
		ParentID:    parentID,
		Level:       &designation.Level,
		UserCount:   &designation.UserCount,
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

// GetBySlug retrieves a designation by slug
func (r *designationRepository) GetBySlug(ctx context.Context, slug string) (*domain.Designation, error) {
	dbDesignation, err := r.queries.GetDesignationBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return toDomainDesignation(dbDesignation), nil
}

// GetByName retrieves a designation by name (case-insensitive)
func (r *designationRepository) GetByName(ctx context.Context, name string) (*domain.Designation, error) {
	dbDesignation, err := r.queries.GetDesignationByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return toDomainDesignation(dbDesignation), nil
}

// Update updates an existing designation
func (r *designationRepository) Update(ctx context.Context, designation *domain.Designation) error {
	var parentID uuid.NullUUID
	if designation.ParentID != nil {
		parentID = uuid.NullUUID{UUID: *designation.ParentID, Valid: true}
	}

	_, err := r.queries.UpdateDesignation(ctx, sqlc.UpdateDesignationParams{
		ID:          designation.ID,
		Name:        designation.Name,
		Description: designation.Description,
		Slug:        designation.Slug,
		IsActive:    &designation.IsActive,
		ParentID:    parentID,
		Level:       &designation.Level,
		UpdatedAt:   pgtype.Timestamptz{Time: designation.UpdatedAt, Valid: true},
	})

	return err
}

// Delete deletes a designation by ID
func (r *designationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteDesignation(ctx, id)
}

// List retrieves designations with pagination and filters
func (r *designationRepository) List(ctx context.Context, page, pageSize int32, activeOnly bool, parentID *uuid.UUID, search string) ([]*domain.Designation, error) {
	offset := (page - 1) * pageSize

	var parentUUID pgtype.UUID
	if parentID != nil {
		parentUUID = pgtype.UUID{Bytes: *parentID, Valid: true}
	} else {
		// Use zero UUID to indicate root level
		parentUUID = pgtype.UUID{Bytes: uuid.Nil, Valid: true}
	}

	dbDesignations, err := r.queries.ListDesignations(ctx, sqlc.ListDesignationsParams{
		Column1: activeOnly,
		Column2: parentUUID,
		Column3: search,
		Limit:   pageSize,
		Offset:  offset,
	})

	if err != nil {
		return nil, err
	}

	designations := make([]*domain.Designation, len(dbDesignations))
	for i, dbDesignation := range dbDesignations {
		designations[i] = toDomainDesignation(dbDesignation)
	}

	return designations, nil
}

// Count returns the total count of designations with filters
func (r *designationRepository) Count(ctx context.Context, activeOnly bool, parentID *uuid.UUID, search string) (int64, error) {
	var parentUUID pgtype.UUID
	if parentID != nil {
		parentUUID = pgtype.UUID{Bytes: *parentID, Valid: true}
	} else {
		parentUUID = pgtype.UUID{Bytes: uuid.Nil, Valid: true}
	}

	count, err := r.queries.CountDesignations(ctx, sqlc.CountDesignationsParams{
		Column1: activeOnly,
		Column2: parentUUID,
		Column3: search,
	})

	return count, err
}

// Exists checks if a designation with the given name exists (case-insensitive)
func (r *designationRepository) Exists(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	var excludeUUID pgtype.UUID
	if excludeID != nil {
		excludeUUID = pgtype.UUID{Bytes: *excludeID, Valid: true}
	}

	exists, err := r.queries.CheckDesignationExists(ctx, sqlc.CheckDesignationExistsParams{
		Lower:   name,
		Column2: excludeUUID,
	})

	return exists, err
}

// SlugExists checks if a designation with the given slug exists
func (r *designationRepository) SlugExists(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error) {
	var excludeUUID pgtype.UUID
	if excludeID != nil {
		excludeUUID = pgtype.UUID{Bytes: *excludeID, Valid: true}
	}

	exists, err := r.queries.CheckSlugExists(ctx, sqlc.CheckSlugExistsParams{
		Slug:    slug,
		Column2: excludeUUID,
	})

	return exists, err
}

// GetChildren retrieves all child designations of a parent
func (r *designationRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*domain.Designation, error) {
	dbDesignations, err := r.queries.GetDesignationChildren(ctx, uuid.NullUUID{UUID: parentID, Valid: true})

	if err != nil {
		return nil, err
	}

	designations := make([]*domain.Designation, len(dbDesignations))
	for i, dbDesignation := range dbDesignations {
		designations[i] = toDomainDesignation(dbDesignation)
	}

	return designations, nil
}

// GetUsersCount returns the count of users assigned to a designation
func (r *designationRepository) GetUsersCount(ctx context.Context, designationID uuid.UUID) (int32, error) {
	// TODO: Add GetDesignationUsersCount query to SQLC
	// This requires a join with the users table which is in a different service
	// For now, return 0
	return 0, nil
}

// UpdateUserCount updates the cached user count for a designation
func (r *designationRepository) UpdateUserCount(ctx context.Context, designationID uuid.UUID, count int32) error {
	now := time.Now()
	return r.queries.UpdateDesignationUserCount(ctx, sqlc.UpdateDesignationUserCountParams{
		ID:        designationID,
		UserCount: &count,
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	})
}

// GetLevel gets the hierarchy level of a designation
func (r *designationRepository) GetLevel(ctx context.Context, designationID uuid.UUID) (int32, error) {
	level, err := r.queries.GetDesignationLevel(ctx, designationID)
	if err != nil {
		return 0, err
	}
	if level == nil {
		return 0, nil
	}
	return *level, nil
}

// CalculateLevel calculates the level based on parent hierarchy
func (r *designationRepository) CalculateLevel(ctx context.Context, parentID *uuid.UUID) (int32, error) {
	if parentID == nil {
		return 0, nil
	}

	level, err := r.queries.CalculateDesignationLevel(ctx, *parentID)
	if err != nil {
		return 0, err
	}

	// CalculateDesignationLevel returns interface{}, need to convert
	if levelInt, ok := level.(int64); ok {
		return int32(levelInt), nil
	}
	return 0, nil
}

// toDomainDesignation converts a database designation to a domain designation
func toDomainDesignation(dbDesignation *sqlc.Designation) *domain.Designation {
	var parentID *uuid.UUID
	if dbDesignation.ParentID.Valid {
		parentID = &dbDesignation.ParentID.UUID
	}

	var isActive bool
	if dbDesignation.IsActive != nil {
		isActive = *dbDesignation.IsActive
	}

	var level int32
	if dbDesignation.Level != nil {
		level = *dbDesignation.Level
	}

	var userCount int32
	if dbDesignation.UserCount != nil {
		userCount = *dbDesignation.UserCount
	}

	return &domain.Designation{
		ID:          dbDesignation.ID,
		Name:        dbDesignation.Name,
		Description: dbDesignation.Description,
		Slug:        dbDesignation.Slug,
		IsActive:    isActive,
		ParentID:    parentID,
		Level:       level,
		UserCount:   userCount,
		CreatedAt:   dbDesignation.CreatedAt.Time,
		UpdatedAt:   dbDesignation.UpdatedAt.Time,
	}
}

