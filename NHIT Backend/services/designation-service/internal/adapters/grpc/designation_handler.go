package grpc

import (
	"context"
	"log"

	designationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/designationpb"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/ports"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DesignationHandler implements the gRPC DesignationService
type DesignationHandler struct {
	designationpb.UnimplementedDesignationServiceServer
	service ports.DesignationService
}

// NewDesignationHandler creates a new designation gRPC handler
func NewDesignationHandler(service ports.DesignationService) *DesignationHandler {
	return &DesignationHandler{
		service: service,
	}
}

// CreateDesignation creates a new designation
func (h *DesignationHandler) CreateDesignation(ctx context.Context, req *designationpb.CreateDesignationRequest) (*designationpb.DesignationResponse, error) {
	log.Printf("gRPC CreateDesignation: name=%s", req.Name)

	// Validate request
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Description == "" {
		return nil, status.Error(codes.InvalidArgument, "description is required")
	}

	// Parse parent ID if provided
	var parentID *uuid.UUID
	if req.ParentId != "" {
		pid, err := uuid.Parse(req.ParentId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid parent_id format")
		}
		parentID = &pid
	}

	// Create designation
	designation, err := h.service.CreateDesignation(ctx, req.Name, req.Description, req.IsActive, parentID)
	if err != nil {
		return nil, handleError(err)
	}

	return &designationpb.DesignationResponse{
		Designation: toProtoDesignation(designation),
	}, nil
}

// GetDesignation retrieves a designation by ID
func (h *DesignationHandler) GetDesignation(ctx context.Context, req *designationpb.GetDesignationRequest) (*designationpb.DesignationResponse, error) {
	log.Printf("gRPC GetDesignation: id=%s", req.Id)

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid designation ID format")
	}

	designation, err := h.service.GetDesignation(ctx, id)
	if err != nil {
		return nil, handleError(err)
	}

	return &designationpb.DesignationResponse{
		Designation: toProtoDesignation(designation),
	}, nil
}

// GetDesignationBySlug retrieves a designation by slug
func (h *DesignationHandler) GetDesignationBySlug(ctx context.Context, req *designationpb.GetDesignationBySlugRequest) (*designationpb.DesignationResponse, error) {
	log.Printf("gRPC GetDesignationBySlug: slug=%s", req.Slug)

	if req.Slug == "" {
		return nil, status.Error(codes.InvalidArgument, "slug is required")
	}

	designation, err := h.service.GetDesignationBySlug(ctx, req.Slug)
	if err != nil {
		return nil, handleError(err)
	}

	return &designationpb.DesignationResponse{
		Designation: toProtoDesignation(designation),
	}, nil
}

// UpdateDesignation updates an existing designation
func (h *DesignationHandler) UpdateDesignation(ctx context.Context, req *designationpb.UpdateDesignationRequest) (*designationpb.DesignationResponse, error) {
	log.Printf("gRPC UpdateDesignation: id=%s, name=%s", req.Id, req.Name)

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid designation ID format")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Description == "" {
		return nil, status.Error(codes.InvalidArgument, "description is required")
	}

	// Parse parent ID if provided
	var parentID *uuid.UUID
	if req.ParentId != "" {
		pid, err := uuid.Parse(req.ParentId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid parent_id format")
		}
		parentID = &pid
	}

	designation, err := h.service.UpdateDesignation(ctx, id, req.Name, req.Description, req.IsActive, parentID)
	if err != nil {
		return nil, handleError(err)
	}

	return &designationpb.DesignationResponse{
		Designation: toProtoDesignation(designation),
	}, nil
}

// DeleteDesignation deletes a designation
func (h *DesignationHandler) DeleteDesignation(ctx context.Context, req *designationpb.DeleteDesignationRequest) (*designationpb.DeleteDesignationResponse, error) {
	log.Printf("gRPC DeleteDesignation: id=%s, force=%v", req.Id, req.Force)

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid designation ID format")
	}

	err = h.service.DeleteDesignation(ctx, id, req.Force)
	if err != nil {
		return nil, handleError(err)
	}

	return &designationpb.DeleteDesignationResponse{
		Success: true,
		Message: "Designation deleted successfully",
	}, nil
}

// ListDesignations lists designations with pagination and filters
func (h *DesignationHandler) ListDesignations(ctx context.Context, req *designationpb.ListDesignationsRequest) (*designationpb.ListDesignationsResponse, error) {
	log.Printf("gRPC ListDesignations: page=%d, pageSize=%d, activeOnly=%v", req.Page, req.PageSize, req.ActiveOnly)

	// Parse parent ID if provided
	var parentID *uuid.UUID
	if req.ParentId != "" {
		pid, err := uuid.Parse(req.ParentId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid parent_id format")
		}
		parentID = &pid
	}

	designations, totalCount, err := h.service.ListDesignations(ctx, req.Page, req.PageSize, req.ActiveOnly, parentID, req.Search)
	if err != nil {
		return nil, handleError(err)
	}

	protoDesignations := make([]*designationpb.Designation, len(designations))
	for i, d := range designations {
		protoDesignations[i] = toProtoDesignation(d)
	}

	return &designationpb.ListDesignationsResponse{
		Designations: protoDesignations,
		TotalCount:   totalCount,
		Page:         req.Page,
		PageSize:     req.PageSize,
	}, nil
}

// GetDesignationHierarchy retrieves designation with parent and children
func (h *DesignationHandler) GetDesignationHierarchy(ctx context.Context, req *designationpb.GetDesignationHierarchyRequest) (*designationpb.GetDesignationHierarchyResponse, error) {
	log.Printf("gRPC GetDesignationHierarchy: id=%s", req.Id)

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid designation ID format")
	}

	designation, parent, children, err := h.service.GetDesignationHierarchy(ctx, id)
	if err != nil {
		return nil, handleError(err)
	}

	hierarchy := &designationpb.DesignationHierarchy{
		Designation: toProtoDesignation(designation),
	}

	if parent != nil {
		hierarchy.Parent = toProtoDesignation(parent)
	}

	hierarchy.Children = make([]*designationpb.Designation, len(children))
	for i, child := range children {
		hierarchy.Children[i] = toProtoDesignation(child)
	}

	return &designationpb.GetDesignationHierarchyResponse{
		Hierarchy: hierarchy,
	}, nil
}

// ToggleDesignationStatus activates or deactivates a designation
func (h *DesignationHandler) ToggleDesignationStatus(ctx context.Context, req *designationpb.ToggleDesignationStatusRequest) (*designationpb.DesignationResponse, error) {
	log.Printf("gRPC ToggleDesignationStatus: id=%s, isActive=%v", req.Id, req.IsActive)

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid designation ID format")
	}

	designation, err := h.service.ToggleDesignationStatus(ctx, id, req.IsActive)
	if err != nil {
		return nil, handleError(err)
	}

	return &designationpb.DesignationResponse{
		Designation: toProtoDesignation(designation),
	}, nil
}

// CheckDesignationExists checks if a designation name exists
func (h *DesignationHandler) CheckDesignationExists(ctx context.Context, req *designationpb.CheckDesignationExistsRequest) (*designationpb.CheckDesignationExistsResponse, error) {
	log.Printf("gRPC CheckDesignationExists: name=%s", req.Name)

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	var excludeID *uuid.UUID
	if req.ExcludeId != "" {
		eid, err := uuid.Parse(req.ExcludeId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid exclude_id format")
		}
		excludeID = &eid
	}

	exists, existingID, err := h.service.CheckDesignationExists(ctx, req.Name, excludeID)
	if err != nil {
		return nil, handleError(err)
	}

	response := &designationpb.CheckDesignationExistsResponse{
		Exists: exists,
	}

	if existingID != nil {
		response.ExistingId = existingID.String()
	}

	return response, nil
}

// GetUsersCount returns the count of users assigned to a designation
func (h *DesignationHandler) GetUsersCount(ctx context.Context, req *designationpb.GetUsersCountRequest) (*designationpb.GetUsersCountResponse, error) {
	log.Printf("gRPC GetUsersCount: designationId=%s", req.DesignationId)

	id, err := uuid.Parse(req.DesignationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid designation ID format")
	}

	count, err := h.service.GetUsersCount(ctx, id)
	if err != nil {
		return nil, handleError(err)
	}

	return &designationpb.GetUsersCountResponse{
		Count: count,
	}, nil
}

// toProtoDesignation converts a domain designation to a proto designation
func toProtoDesignation(d *domain.Designation) *designationpb.Designation {
	proto := &designationpb.Designation{
		Id:          d.ID.String(),
		Name:        d.Name,
		Description: d.Description,
		Slug:        d.Slug,
		IsActive:    d.IsActive,
		Level:       d.Level,
		UserCount:   d.UserCount,
		CreatedAt:   timestamppb.New(d.CreatedAt),
		UpdatedAt:   timestamppb.New(d.UpdatedAt),
	}

	if d.ParentID != nil {
		proto.ParentId = d.ParentID.String()
	}

	return proto
}

// handleError converts domain errors to gRPC status errors
func handleError(err error) error {
	switch err {
	case domain.ErrDesignationNotFound:
		return status.Error(codes.NotFound, err.Error())
	case domain.ErrDesignationAlreadyExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case domain.ErrDesignationHasUsers:
		return status.Error(codes.FailedPrecondition, err.Error())
	case domain.ErrCircularReference:
		return status.Error(codes.InvalidArgument, err.Error())
	case domain.ErrInvalidParent:
		return status.Error(codes.InvalidArgument, err.Error())
	case domain.ErrParentNotFound:
		return status.Error(codes.NotFound, err.Error())
	case domain.ErrMaxHierarchyDepth:
		return status.Error(codes.InvalidArgument, err.Error())
	case domain.ErrCannotDeactivateWithUsers:
		return status.Error(codes.FailedPrecondition, err.Error())
	case domain.ErrDesignationNameRequired,
		domain.ErrDesignationNameTooShort,
		domain.ErrDesignationNameTooLong,
		domain.ErrDesignationNameInvalidChars,
		domain.ErrDesignationNameReserved,
		domain.ErrDesignationDescriptionRequired,
		domain.ErrDesignationDescriptionTooShort,
		domain.ErrDesignationDescriptionTooLong:
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		log.Printf("Unhandled error: %v", err)
		return status.Error(codes.Internal, "internal server error")
	}
}
