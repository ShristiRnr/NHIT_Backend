package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

type DesignationService struct {
	repo ports.DesignationRepository
}

func NewDesignationService(repo ports.DesignationRepository) *DesignationService {
	return &DesignationService{repo: repo}
}

func (s *DesignationService) Create(ctx context.Context, name, description string) (db.Designation, error) {
	return s.repo.Create(ctx, db.Designation{
		Name:        name,
		Description: description,
	})
}

func (s *DesignationService) Get(ctx context.Context, id string) (db.Designation, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return db.Designation{}, err
	}
	return s.repo.Get(ctx, uid)
}

func (s *DesignationService) Update(ctx context.Context, id, name, description string) (db.Designation, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return db.Designation{}, err
	}
	return s.repo.Update(ctx, db.Designation{
		ID:          uid,
		Name:        name,
		Description: description,
	})
}

func (s *DesignationService) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, uid)
}

func (s *DesignationService) List(ctx context.Context, page, pageSize int32) ([]db.Designation, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, pageSize, offset)
}
