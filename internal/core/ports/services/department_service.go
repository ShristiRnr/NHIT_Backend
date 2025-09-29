package services

import (
	"context"

	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

type DepartmentService struct {
	repo ports.DepartmentRepository
}

func NewDepartmentService(repo ports.DepartmentRepository) *DepartmentService {
	return &DepartmentService{repo: repo}
}

func (s *DepartmentService) Create(ctx context.Context, name, description string) (db.Department, error) {
	return s.repo.Create(ctx, name, description)
}

func (s *DepartmentService) Get(ctx context.Context, id string) (db.Department, error) {
	return s.repo.Get(ctx, id)
}

func (s *DepartmentService) List(ctx context.Context, page, pageSize int32) ([]db.Department, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, pageSize, offset)
}

func (s *DepartmentService) Update(ctx context.Context, id, name, description string) (db.Department, error) {
	return s.repo.Update(ctx, id, name, description)
}

func (s *DepartmentService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
