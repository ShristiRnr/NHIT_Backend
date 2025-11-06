package ports

import (
	"context"

	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/domain"
	"github.com/google/uuid"
)

// DesignationRepository defines the interface for designation data access
type DesignationRepository interface {
	// Create creates a new designation
	Create(ctx context.Context, designation *domain.Designation) error

	// GetByID retrieves a designation by ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Designation, error)

	// GetBySlug retrieves a designation by slug
	GetBySlug(ctx context.Context, slug string) (*domain.Designation, error)

	// GetByName retrieves a designation by name (case-insensitive)
	GetByName(ctx context.Context, name string) (*domain.Designation, error)

	// Update updates an existing designation
	Update(ctx context.Context, designation *domain.Designation) error

	// Delete deletes a designation by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves designations with pagination and filters
	List(ctx context.Context, page, pageSize int32, activeOnly bool, parentID *uuid.UUID, search string) ([]*domain.Designation, error)

	// Count returns the total count of designations with filters
	Count(ctx context.Context, activeOnly bool, parentID *uuid.UUID, search string) (int64, error)

	// Exists checks if a designation with the given name exists (case-insensitive)
	Exists(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error)

	// SlugExists checks if a designation with the given slug exists
	SlugExists(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error)

	// GetChildren retrieves all child designations of a parent
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*domain.Designation, error)

	// GetUsersCount returns the count of users assigned to a designation
	GetUsersCount(ctx context.Context, designationID uuid.UUID) (int32, error)

	// UpdateUserCount updates the cached user count for a designation
	UpdateUserCount(ctx context.Context, designationID uuid.UUID, count int32) error

	// GetLevel gets the hierarchy level of a designation
	GetLevel(ctx context.Context, designationID uuid.UUID) (int32, error)

	// CalculateLevel calculates the level based on parent hierarchy
	CalculateLevel(ctx context.Context, parentID *uuid.UUID) (int32, error)
}
