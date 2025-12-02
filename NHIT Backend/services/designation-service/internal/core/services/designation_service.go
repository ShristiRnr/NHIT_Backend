package services

import (
	"context"

	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/ports"
	"github.com/google/uuid"
)

type designationService struct {
	repo ports.DesignationRepository
}

func NewDesignationService(repo ports.DesignationRepository) ports.DesignationService {
	return &designationService{repo: repo}
}

func (s *designationService) CreateDesignation(ctx context.Context, name, description string) (*domain.Designation, error) {
	// domain.NewDesignation sets ID, CreatedAt, UpdatedAt
	d, err := domain.NewDesignation(name, description)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *designationService) GetDesignation(ctx context.Context, id uuid.UUID) (*domain.Designation, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *designationService) UpdateDesignation(ctx context.Context, id uuid.UUID, name, description string) (*domain.Designation, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := d.Update(name, description); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *designationService) DeleteDesignation(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *designationService) ListDesignations(ctx context.Context, page, pageSize int32) ([]*domain.Designation, error) {
    list, err := s.repo.List(ctx, page, pageSize)
    if err != nil {
        return nil, err
    }
    return list, nil
}
