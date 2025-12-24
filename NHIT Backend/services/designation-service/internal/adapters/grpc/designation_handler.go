package grpc

import (
	"context"
	"log"

	"strings"

	designationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/designationpb"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/designation-service/internal/core/ports"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DesignationHandler struct {
	designationpb.UnimplementedDesignationServiceServer
	service ports.DesignationService
}

func NewDesignationHandler(service ports.DesignationService) *DesignationHandler {
	return &DesignationHandler{service: service}
}

// helper to get first non-empty metadata value by keys
func firstMetadataValue(md metadata.MD, keys ...string) string {
	for _, k := range keys {
		if vals := md[strings.ToLower(k)]; len(vals) > 0 && vals[0] != "" {
			return vals[0]
		}
	}
	return ""
}

// getOrgIDFromContext extracts org_id from metadata
func getOrgIDFromContext(ctx context.Context) *uuid.UUID {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	// Check headers first (if context switching)
	orgIDStr := firstMetadataValue(md, "x-org-id", "org-id", "orgId", "org_id")
	if orgIDStr == "" {
		return nil
	}

	id, err := uuid.Parse(orgIDStr)
	if err != nil {
		return nil
	}
	return &id
}

func (h *DesignationHandler) CreateDesignation(ctx context.Context, req *designationpb.CreateDesignationRequest) (*designationpb.DesignationResponse, error) {
	log.Printf("gRPC CreateDesignation: name=%s", req.Name)

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Description == "" {
		return nil, status.Error(codes.InvalidArgument, "description is required")
	}

	orgID := getOrgIDFromContext(ctx)

	designation, err := h.service.CreateDesignation(ctx, req.Name, req.Description, orgID)
	if err != nil {
		return nil, handleError(err)
	}

	return &designationpb.DesignationResponse{
		Designation: toProtoDesignation(designation),
	}, nil
}

func (h *DesignationHandler) GetDesignation(ctx context.Context, req *designationpb.GetDesignationRequest) (*designationpb.DesignationResponse, error) {
	log.Printf("gRPC GetDesignation: id=%s", req.Id)

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid designation ID")
	}

	designation, err := h.service.GetDesignation(ctx, id)
	if err != nil {
		return nil, handleError(err)
	}

	return &designationpb.DesignationResponse{Designation: toProtoDesignation(designation)}, nil
}

func (h *DesignationHandler) UpdateDesignation(ctx context.Context, req *designationpb.UpdateDesignationRequest) (*designationpb.DesignationResponse, error) {
	log.Printf("gRPC UpdateDesignation: id=%s", req.Id)

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid designation ID")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Description == "" {
		return nil, status.Error(codes.InvalidArgument, "description is required")
	}

	designation, err := h.service.UpdateDesignation(ctx, id, req.Name, req.Description)
	if err != nil {
		return nil, handleError(err)
	}

	return &designationpb.DesignationResponse{Designation: toProtoDesignation(designation)}, nil
}

func (h *DesignationHandler) DeleteDesignation(ctx context.Context, req *designationpb.DeleteDesignationRequest) (*designationpb.DeleteDesignationResponse, error) {
	log.Printf("gRPC DeleteDesignation: id=%s", req.Id)

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid designation ID")
	}

	if err := h.service.DeleteDesignation(ctx, id); err != nil {
		return nil, handleError(err)
	}

	return &designationpb.DeleteDesignationResponse{
		Success: true,
		Message: "Designation deleted successfully",
	}, nil
}
func (h *DesignationHandler) ListDesignations(
    ctx context.Context,
    req *designationpb.ListDesignationsRequest,
) (*designationpb.ListDesignationsResponse, error) {

    if req.Page < 1 {
        req.Page = 1
    }
    if req.PageSize < 1 || req.PageSize > 100 {
        req.PageSize = 10
    }

    orgID := getOrgIDFromContext(ctx)

    list, total, err := h.service.ListDesignations(ctx, orgID, req.Page, req.PageSize)
    if err != nil {
        return nil, handleError(err)
    }

    result := make([]*designationpb.Designation, len(list))
    for i, d := range list {
        result[i] = &designationpb.Designation{
            Id:          d.ID.String(),
            Name:        d.Name,
            Description: d.Description,
            CreatedAt:   timestamppb.New(d.CreatedAt),
            UpdatedAt:   timestamppb.New(d.UpdatedAt),
        }
    }

	totalPages := (int32(total) + req.PageSize - 1) / req.PageSize

    return &designationpb.ListDesignationsResponse{
        Designations: result,
		Pagination: &designationpb.PaginationMetadata{
			CurrentPage: req.Page,
			PageSize:    req.PageSize,
			TotalItems:  total,
			TotalPages:  totalPages,
		},
    }, nil
}


func toProtoDesignation(d *domain.Designation) *designationpb.Designation {
	return &designationpb.Designation{
		Id:          d.ID.String(),
		Name:        d.Name,
		Description: d.Description,
		CreatedAt:   timestamppb.New(d.CreatedAt),
		UpdatedAt:   timestamppb.New(d.UpdatedAt),
	}
}

func handleError(err error) error {
	switch err {
	case domain.ErrDesignationNotFound:
		return status.Error(codes.NotFound, err.Error())
	case domain.ErrDesignationAlreadyExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case domain.ErrDesignationHasUsers:
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
