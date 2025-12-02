package handler

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/ports"
)

type projectHandler struct {
	pb.UnimplementedProjectServiceServer
	service ports.ProjectService
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(service ports.ProjectService) pb.ProjectServiceServer {
	return &projectHandler{
		service: service,
	}
}

// GetProject retrieves a project by ID
func (h *projectHandler) GetProject(ctx context.Context, req *pb.GetProjectRequest) (*pb.GetProjectResponse, error) {
	projectID, err := uuid.Parse(req.ProjectId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid project ID: %v", err)
	}

	project, err := h.service.GetProject(ctx, projectID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "project not found: %v", err)
	}

	return &pb.GetProjectResponse{
		Project: toProtoProject(project),
	}, nil
}

// ListProjectsByOrganization lists all projects for an organization
func (h *projectHandler) ListProjectsByOrganization(ctx context.Context, req *pb.ListProjectsByOrganizationRequest) (*pb.ListProjectsByOrganizationResponse, error) {
	// TODO: Implement this method by calling the service
	// For now, return empty list
	projects := []*pb.Project{}

	return &pb.ListProjectsByOrganizationResponse{
		Projects: projects,
	}, nil
}

// Helper function to convert domain project to proto project
func toProtoProject(p *domain.Project) *pb.Project {
	project := &pb.Project{
		ProjectId:   p.ProjectID.String(),
		TenantId:    p.TenantID.String(),
		OrgId:       p.OrgID.String(),
		ProjectName: p.ProjectName,
		CreatedBy:   p.CreatedBy,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}

	return project
}
