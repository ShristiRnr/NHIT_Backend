package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/ports"
)

type projectRepository struct {
	db *sql.DB
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *sql.DB) ports.ProjectRepository {
	return &projectRepository{db: db}
}

// Create creates a new project
func (r *projectRepository) Create(ctx context.Context, project *domain.Project) (*domain.Project, error) {
	query := `
		INSERT INTO projects (
			project_id, tenant_id, org_id, project_name, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING project_id, created_at, updated_at
	`
	
	err := r.db.QueryRowContext(ctx, query,
		project.ProjectID, project.TenantID, project.OrgID,
		project.ProjectName, project.CreatedBy, project.CreatedAt, project.UpdatedAt,
	).Scan(&project.ProjectID, &project.CreatedAt, &project.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	
	return project, nil
}

// GetByID retrieves a project by ID
func (r *projectRepository) GetByID(ctx context.Context, projectID uuid.UUID) (*domain.Project, error) {
	query := `
		SELECT project_id, tenant_id, org_id, project_name, created_by,
			created_at, updated_at
		FROM projects
		WHERE project_id = $1
	`
	
	project := &domain.Project{}
	err := r.db.QueryRowContext(ctx, query, projectID).Scan(
		&project.ProjectID, &project.TenantID, &project.OrgID, &project.ProjectName, &project.CreatedBy, &project.CreatedAt, &project.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	return project, nil
}