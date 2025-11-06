package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/core/ports"
)

type departmentService struct {
	repo ports.DepartmentRepository
}

// NewDepartmentService creates a new department service
func NewDepartmentService(repo ports.DepartmentRepository) ports.DepartmentService {
	return &departmentService{
		repo: repo,
	}
}

// CreateDepartment creates a new department with validation
func (s *departmentService) CreateDepartment(ctx context.Context, name, description string) (*domain.Department, error) {
	// Trim and validate input
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	// Create department domain object
	dept := domain.NewDepartment(name, description)

	// Validate
	if err := dept.Validate(); err != nil {
		log.Printf("[DepartmentService] Validation error: %v", err)
		return nil, err
	}

	// Check if department already exists
	exists, err := s.repo.Exists(ctx, name)
	if err != nil {
		log.Printf("[DepartmentService] Error checking existence: %v", err)
		return nil, err
	}
	if exists {
		log.Printf("[DepartmentService] Department already exists: %s", name)
		return nil, domain.ErrDepartmentAlreadyExists
	}

	// Create department
	created, err := s.repo.Create(ctx, dept)
	if err != nil {
		log.Printf("[DepartmentService] Error creating department: %v", err)
		return nil, err
	}

	log.Printf("[DepartmentService] Department created: %s (ID: %s)", created.Name, created.ID)
	return created, nil
}

// GetDepartment retrieves a department by ID
func (s *departmentService) GetDepartment(ctx context.Context, id uuid.UUID) (*domain.Department, error) {
	if id == uuid.Nil {
		return nil, domain.ErrInvalidDepartmentID
	}

	dept, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Printf("[DepartmentService] Error getting department: %v", err)
		return nil, domain.ErrDepartmentNotFound
	}

	return dept, nil
}

// UpdateDepartment updates a department with validation
func (s *departmentService) UpdateDepartment(ctx context.Context, id uuid.UUID, name, description string) (*domain.Department, error) {
	if id == uuid.Nil {
		return nil, domain.ErrInvalidDepartmentID
	}

	// Trim input
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	// Check if department exists
	exists, err := s.repo.ExistsByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrDepartmentNotFound
	}

	// Get existing department
	dept, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrDepartmentNotFound
	}

	// Update fields
	dept.Name = name
	dept.Description = description

	// Validate
	if err := dept.Validate(); err != nil {
		log.Printf("[DepartmentService] Validation error: %v", err)
		return nil, err
	}

	// Check if name is already taken by another department
	existingDept, err := s.repo.GetByName(ctx, name)
	if err == nil && existingDept.ID != id {
		return nil, domain.ErrDepartmentAlreadyExists
	}

	// Update department
	updated, err := s.repo.Update(ctx, dept)
	if err != nil {
		log.Printf("[DepartmentService] Error updating department: %v", err)
		return nil, err
	}

	log.Printf("[DepartmentService] Department updated: %s (ID: %s)", updated.Name, updated.ID)
	return updated, nil
}

// DeleteDepartment deletes a department with validation
func (s *departmentService) DeleteDepartment(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return domain.ErrInvalidDepartmentID
	}

	// Check if department exists
	exists, err := s.repo.ExistsByID(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrDepartmentNotFound
	}

	// Check if department has users
	userCount, err := s.repo.CountUsersByDepartment(ctx, id)
	if err != nil {
		log.Printf("[DepartmentService] Error checking users: %v", err)
		return err
	}
	if userCount > 0 {
		return fmt.Errorf("%w: %d users assigned", domain.ErrDepartmentHasUsers, userCount)
	}

	// Delete department
	if err := s.repo.Delete(ctx, id); err != nil {
		log.Printf("[DepartmentService] Error deleting department: %v", err)
		return err
	}

	log.Printf("[DepartmentService] Department deleted: ID %s", id)
	return nil
}

// ListDepartments retrieves departments with pagination
func (s *departmentService) ListDepartments(ctx context.Context, page, pageSize int32) ([]*domain.Department, int32, error) {
	// Set defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	departments, total, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		log.Printf("[DepartmentService] Error listing departments: %v", err)
		return nil, 0, err
	}

	return departments, total, nil
}
