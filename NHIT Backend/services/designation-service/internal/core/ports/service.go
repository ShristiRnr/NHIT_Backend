package ports

import (
	"context"

	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/domain"
	"github.com/google/uuid"
)

// DesignationService defines the interface for designation business logic
type DesignationService interface {
	// CreateDesignation creates a new designation with validation
	CreateDesignation(ctx context.Context, name, description string, isActive bool, parentID *uuid.UUID) (*domain.Designation, error)

	// GetDesignation retrieves a designation by ID
	GetDesignation(ctx context.Context, id uuid.UUID) (*domain.Designation, error)

	// GetDesignationBySlug retrieves a designation by slug
	GetDesignationBySlug(ctx context.Context, slug string) (*domain.Designation, error)

	// UpdateDesignation updates an existing designation
	UpdateDesignation(ctx context.Context, id uuid.UUID, name, description string, isActive bool, parentID *uuid.UUID) (*domain.Designation, error)

	// DeleteDesignation deletes a designation
	DeleteDesignation(ctx context.Context, id uuid.UUID, force bool) error

	// ListDesignations retrieves designations with pagination and filters
	ListDesignations(ctx context.Context, page, pageSize int32, activeOnly bool, parentID *uuid.UUID, search string) ([]*domain.Designation, int64, error)

	// GetDesignationHierarchy retrieves designation with parent and children
	GetDesignationHierarchy(ctx context.Context, id uuid.UUID) (*domain.Designation, *domain.Designation, []*domain.Designation, error)

	// ToggleDesignationStatus activates or deactivates a designation
	ToggleDesignationStatus(ctx context.Context, id uuid.UUID, isActive bool) (*domain.Designation, error)

	// CheckDesignationExists checks if a designation name exists
	CheckDesignationExists(ctx context.Context, name string, excludeID *uuid.UUID) (bool, *uuid.UUID, error)

	// GetUsersCount returns the count of users assigned to a designation
	GetUsersCount(ctx context.Context, designationID uuid.UUID) (int32, error)
}
