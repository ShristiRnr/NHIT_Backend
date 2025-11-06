package services

import (
	"context"
	"fmt"
	"log"

	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/ports"
	"github.com/google/uuid"
)

const (
	MaxHierarchyDepth = 5 // Maximum allowed hierarchy depth
)

type designationService struct {
	repo ports.DesignationRepository
}

// NewDesignationService creates a new designation service
func NewDesignationService(repo ports.DesignationRepository) ports.DesignationService {
	return &designationService{
		repo: repo,
	}
}

// CreateDesignation creates a new designation with strong validation
func (s *designationService) CreateDesignation(ctx context.Context, name, description string, isActive bool, parentID *uuid.UUID) (*domain.Designation, error) {
	log.Printf("Creating designation: name=%s, isActive=%v, hasParent=%v", name, isActive, parentID != nil)

	// Validate parent if provided
	if parentID != nil {
		if err := s.validateParent(ctx, *parentID); err != nil {
			log.Printf("Parent validation failed: %v", err)
			return nil, err
		}

		// Check hierarchy depth
		level, err := s.repo.CalculateLevel(ctx, parentID)
		if err != nil {
			log.Printf("Failed to calculate level: %v", err)
			return nil, err
		}

		if level >= MaxHierarchyDepth {
			log.Printf("Maximum hierarchy depth exceeded: level=%d", level)
			return nil, domain.ErrMaxHierarchyDepth
		}
	}

	// Create designation domain object (includes validation)
	designation, err := domain.NewDesignation(name, description, isActive, parentID)
	if err != nil {
		log.Printf("Designation creation failed: %v", err)
		return nil, err
	}

	// Check for duplicate name (case-insensitive)
	exists, err := s.repo.Exists(ctx, designation.Name, nil)
	if err != nil {
		log.Printf("Failed to check designation existence: %v", err)
		return nil, err
	}
	if exists {
		log.Printf("Designation already exists: name=%s", designation.Name)
		return nil, domain.ErrDesignationAlreadyExists
	}

	// Check for duplicate slug
	slugExists, err := s.repo.SlugExists(ctx, designation.Slug, nil)
	if err != nil {
		log.Printf("Failed to check slug existence: %v", err)
		return nil, err
	}
	if slugExists {
		// Generate unique slug by appending UUID suffix
		designation.Slug = fmt.Sprintf("%s-%s", designation.Slug, uuid.New().String()[:8])
		log.Printf("Slug collision detected, using unique slug: %s", designation.Slug)
	}

	// Calculate and set hierarchy level
	level, err := s.repo.CalculateLevel(ctx, parentID)
	if err != nil {
		log.Printf("Failed to calculate level: %v", err)
		return nil, err
	}
	designation.SetLevel(level)

	// Save to repository
	if err := s.repo.Create(ctx, designation); err != nil {
		log.Printf("Failed to create designation in repository: %v", err)
		return nil, err
	}

	log.Printf("Designation created successfully: id=%s, name=%s, slug=%s, level=%d", designation.ID, designation.Name, designation.Slug, designation.Level)
	return designation, nil
}

// GetDesignation retrieves a designation by ID
func (s *designationService) GetDesignation(ctx context.Context, id uuid.UUID) (*domain.Designation, error) {
	log.Printf("Getting designation: id=%s", id)

	designation, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Printf("Failed to get designation: %v", err)
		return nil, domain.ErrDesignationNotFound
	}

	// Update user count from database
	userCount, err := s.repo.GetUsersCount(ctx, id)
	if err != nil {
		log.Printf("Warning: Failed to get user count: %v", err)
		// Don't fail the request, just use cached count
	} else {
		designation.UpdateUserCount(userCount)
	}

	log.Printf("Designation retrieved: id=%s, name=%s, userCount=%d", designation.ID, designation.Name, designation.UserCount)
	return designation, nil
}

// GetDesignationBySlug retrieves a designation by slug
func (s *designationService) GetDesignationBySlug(ctx context.Context, slug string) (*domain.Designation, error) {
	log.Printf("Getting designation by slug: slug=%s", slug)

	designation, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		log.Printf("Failed to get designation by slug: %v", err)
		return nil, domain.ErrDesignationNotFound
	}

	// Update user count
	userCount, err := s.repo.GetUsersCount(ctx, designation.ID)
	if err == nil {
		designation.UpdateUserCount(userCount)
	}

	log.Printf("Designation retrieved by slug: id=%s, name=%s", designation.ID, designation.Name)
	return designation, nil
}

// UpdateDesignation updates an existing designation with validation
func (s *designationService) UpdateDesignation(ctx context.Context, id uuid.UUID, name, description string, isActive bool, parentID *uuid.UUID) (*domain.Designation, error) {
	log.Printf("Updating designation: id=%s, name=%s, isActive=%v", id, name, isActive)

	// Get existing designation
	designation, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Printf("Designation not found: %v", err)
		return nil, domain.ErrDesignationNotFound
	}

	// Validate parent if provided
	if parentID != nil {
		// Cannot set self as parent
		if *parentID == id {
			log.Printf("Circular reference detected: designation cannot be its own parent")
			return nil, domain.ErrCircularReference
		}

		if err := s.validateParent(ctx, *parentID); err != nil {
			log.Printf("Parent validation failed: %v", err)
			return nil, err
		}

		// Check hierarchy depth
		level, err := s.repo.CalculateLevel(ctx, parentID)
		if err != nil {
			log.Printf("Failed to calculate level: %v", err)
			return nil, err
		}

		if level >= MaxHierarchyDepth {
			log.Printf("Maximum hierarchy depth exceeded: level=%d", level)
			return nil, domain.ErrMaxHierarchyDepth
		}
	}

	// Check for duplicate name (excluding current designation)
	exists, err := s.repo.Exists(ctx, name, &id)
	if err != nil {
		log.Printf("Failed to check designation existence: %v", err)
		return nil, err
	}
	if exists {
		log.Printf("Designation name already exists: name=%s", name)
		return nil, domain.ErrDesignationAlreadyExists
	}

	// If deactivating, check if users are assigned
	if !isActive && designation.IsActive {
		userCount, err := s.repo.GetUsersCount(ctx, id)
		if err != nil {
			log.Printf("Failed to get user count: %v", err)
			return nil, err
		}
		if userCount > 0 {
			log.Printf("Cannot deactivate designation with users: userCount=%d", userCount)
			return nil, domain.ErrCannotDeactivateWithUsers
		}
	}

	// Update designation
	if err := designation.Update(name, description, isActive, parentID); err != nil {
		log.Printf("Failed to update designation: %v", err)
		return nil, err
	}

	// Recalculate level if parent changed
	level, err := s.repo.CalculateLevel(ctx, parentID)
	if err != nil {
		log.Printf("Failed to calculate level: %v", err)
		return nil, err
	}
	designation.SetLevel(level)

	// Save to repository
	if err := s.repo.Update(ctx, designation); err != nil {
		log.Printf("Failed to update designation in repository: %v", err)
		return nil, err
	}

	log.Printf("Designation updated successfully: id=%s, name=%s, level=%d", designation.ID, designation.Name, designation.Level)
	return designation, nil
}

// DeleteDesignation deletes a designation with business logic validation
func (s *designationService) DeleteDesignation(ctx context.Context, id uuid.UUID, force bool) error {
	log.Printf("Deleting designation: id=%s, force=%v", id, force)

	// Get designation
	designation, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Printf("Designation not found: %v", err)
		return domain.ErrDesignationNotFound
	}

	// Check if users are assigned
	userCount, err := s.repo.GetUsersCount(ctx, id)
	if err != nil {
		log.Printf("Failed to get user count: %v", err)
		return err
	}

	if userCount > 0 && !force {
		log.Printf("Cannot delete designation with users: userCount=%d", userCount)
		return domain.ErrDesignationHasUsers
	}

	// Check if designation has children
	children, err := s.repo.GetChildren(ctx, id)
	if err != nil {
		log.Printf("Failed to get children: %v", err)
		return err
	}

	if len(children) > 0 && !force {
		log.Printf("Cannot delete designation with children: childCount=%d", len(children))
		return fmt.Errorf("cannot delete designation: %d child designations exist", len(children))
	}

	// If force delete, reassign children to parent or null
	if force && len(children) > 0 {
		log.Printf("Force deleting: reassigning %d children", len(children))
		// In a real implementation, you would reassign children here
		// For now, we'll just log it
	}

	// Delete designation
	if err := s.repo.Delete(ctx, id); err != nil {
		log.Printf("Failed to delete designation: %v", err)
		return err
	}

	log.Printf("Designation deleted successfully: id=%s, name=%s", id, designation.Name)
	return nil
}

// ListDesignations retrieves designations with pagination and filters
func (s *designationService) ListDesignations(ctx context.Context, page, pageSize int32, activeOnly bool, parentID *uuid.UUID, search string) ([]*domain.Designation, int64, error) {
	log.Printf("Listing designations: page=%d, pageSize=%d, activeOnly=%v, hasParent=%v, search=%s", page, pageSize, activeOnly, parentID != nil, search)

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Get designations
	designations, err := s.repo.List(ctx, page, pageSize, activeOnly, parentID, search)
	if err != nil {
		log.Printf("Failed to list designations: %v", err)
		return nil, 0, err
	}

	// Get total count
	totalCount, err := s.repo.Count(ctx, activeOnly, parentID, search)
	if err != nil {
		log.Printf("Failed to count designations: %v", err)
		return nil, 0, err
	}

	// Update user counts for all designations
	for _, d := range designations {
		userCount, err := s.repo.GetUsersCount(ctx, d.ID)
		if err == nil {
			d.UpdateUserCount(userCount)
		}
	}

	log.Printf("Designations listed: count=%d, total=%d", len(designations), totalCount)
	return designations, totalCount, nil
}

// GetDesignationHierarchy retrieves designation with parent and children
func (s *designationService) GetDesignationHierarchy(ctx context.Context, id uuid.UUID) (*domain.Designation, *domain.Designation, []*domain.Designation, error) {
	log.Printf("Getting designation hierarchy: id=%s", id)

	// Get designation
	designation, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Printf("Designation not found: %v", err)
		return nil, nil, nil, domain.ErrDesignationNotFound
	}

	// Get parent if exists
	var parent *domain.Designation
	if designation.ParentID != nil {
		parent, err = s.repo.GetByID(ctx, *designation.ParentID)
		if err != nil {
			log.Printf("Warning: Parent not found: %v", err)
			// Don't fail, just set parent to nil
			parent = nil
		}
	}

	// Get children
	children, err := s.repo.GetChildren(ctx, id)
	if err != nil {
		log.Printf("Failed to get children: %v", err)
		return nil, nil, nil, err
	}

	log.Printf("Hierarchy retrieved: designation=%s, hasParent=%v, childCount=%d", designation.Name, parent != nil, len(children))
	return designation, parent, children, nil
}

// ToggleDesignationStatus activates or deactivates a designation
func (s *designationService) ToggleDesignationStatus(ctx context.Context, id uuid.UUID, isActive bool) (*domain.Designation, error) {
	log.Printf("Toggling designation status: id=%s, isActive=%v", id, isActive)

	// Get designation
	designation, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Printf("Designation not found: %v", err)
		return nil, domain.ErrDesignationNotFound
	}

	// If deactivating, check if users are assigned
	if !isActive && designation.IsActive {
		userCount, err := s.repo.GetUsersCount(ctx, id)
		if err != nil {
			log.Printf("Failed to get user count: %v", err)
			return nil, err
		}
		if userCount > 0 {
			log.Printf("Cannot deactivate designation with users: userCount=%d", userCount)
			return nil, domain.ErrCannotDeactivateWithUsers
		}
	}

	// Toggle status
	if isActive {
		designation.Activate()
	} else {
		designation.Deactivate()
	}

	// Save to repository
	if err := s.repo.Update(ctx, designation); err != nil {
		log.Printf("Failed to update designation status: %v", err)
		return nil, err
	}

	log.Printf("Designation status toggled: id=%s, isActive=%v", id, isActive)
	return designation, nil
}

// CheckDesignationExists checks if a designation name exists
func (s *designationService) CheckDesignationExists(ctx context.Context, name string, excludeID *uuid.UUID) (bool, *uuid.UUID, error) {
	log.Printf("Checking designation existence: name=%s, hasExclude=%v", name, excludeID != nil)

	exists, err := s.repo.Exists(ctx, name, excludeID)
	if err != nil {
		log.Printf("Failed to check existence: %v", err)
		return false, nil, err
	}

	if !exists {
		log.Printf("Designation does not exist: name=%s", name)
		return false, nil, nil
	}

	// Get the existing designation ID
	designation, err := s.repo.GetByName(ctx, name)
	if err != nil {
		log.Printf("Failed to get designation by name: %v", err)
		return true, nil, nil
	}

	log.Printf("Designation exists: name=%s, id=%s", name, designation.ID)
	return true, &designation.ID, nil
}

// GetUsersCount returns the count of users assigned to a designation
func (s *designationService) GetUsersCount(ctx context.Context, designationID uuid.UUID) (int32, error) {
	log.Printf("Getting users count: designationID=%s", designationID)

	// Verify designation exists
	_, err := s.repo.GetByID(ctx, designationID)
	if err != nil {
		log.Printf("Designation not found: %v", err)
		return 0, domain.ErrDesignationNotFound
	}

	count, err := s.repo.GetUsersCount(ctx, designationID)
	if err != nil {
		log.Printf("Failed to get users count: %v", err)
		return 0, err
	}

	log.Printf("Users count retrieved: designationID=%s, count=%d", designationID, count)
	return count, nil
}

// validateParent validates that a parent designation exists and is active
func (s *designationService) validateParent(ctx context.Context, parentID uuid.UUID) error {
	parent, err := s.repo.GetByID(ctx, parentID)
	if err != nil {
		return domain.ErrParentNotFound
	}

	if !parent.IsActive {
		return fmt.Errorf("parent designation is not active")
	}

	return nil
}
