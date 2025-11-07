package grpc

import (
	"context"
	"log"

	"github.com/google/uuid"
	organizationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrganizationHandler struct {
	organizationpb.UnimplementedOrganizationServiceServer
	orgService     ports.OrganizationService
	userOrgService ports.UserOrganizationService
}

// NewOrganizationHandler creates a new gRPC organization handler
func NewOrganizationHandler(
	orgService ports.OrganizationService,
	userOrgService ports.UserOrganizationService,
) *OrganizationHandler {
	return &OrganizationHandler{
		orgService:     orgService,
		userOrgService: userOrgService,
	}
}

// CreateOrganization creates a new organization
func (h *OrganizationHandler) CreateOrganization(
	ctx context.Context,
	req *organizationpb.CreateOrganizationRequest,
) (*organizationpb.OrganizationResponse, error) {
	log.Printf("gRPC CreateOrganization: name=%s, code=%s", req.GetName(), req.GetCode())
	
	// Parse UUIDs
	tenantID, err := uuid.Parse(req.GetTenantId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant ID: %v", err)
	}
	
	createdBy, err := uuid.Parse(req.GetCreatedBy())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid creator ID: %v", err)
	}
	
	// Create organization
	org, err := h.orgService.CreateOrganization(
		ctx,
		tenantID,
		req.GetName(),
		req.GetCode(),
		req.GetDescription(),
		req.GetLogo(),
		createdBy,
	)
	
	if err != nil {
		if err == domain.ErrDuplicateOrganizationCode {
			return nil, status.Errorf(codes.AlreadyExists, "organization code already exists")
		}
		if err == domain.ErrInvalidOrganizationName || err == domain.ErrInvalidOrganizationCode {
			return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
		}
		log.Printf("Failed to create organization: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create organization: %v", err)
	}
	
	return &organizationpb.OrganizationResponse{
		Organization: toProtoOrganization(org),
		Message:      "Organization created successfully",
	}, nil
}

// GetOrganization retrieves an organization by ID
func (h *OrganizationHandler) GetOrganization(
	ctx context.Context,
	req *organizationpb.GetOrganizationRequest,
) (*organizationpb.OrganizationResponse, error) {
	orgID, err := uuid.Parse(req.GetOrgId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid organization ID: %v", err)
	}
	
	org, err := h.orgService.GetOrganization(ctx, orgID)
	if err != nil {
		log.Printf("Failed to get organization: %v", err)
		return nil, status.Errorf(codes.NotFound, "organization not found")
	}
	
	return &organizationpb.OrganizationResponse{
		Organization: toProtoOrganization(org),
		Message:      "Organization retrieved successfully",
	}, nil
}

// GetOrganizationByCode retrieves an organization by code
func (h *OrganizationHandler) GetOrganizationByCode(
	ctx context.Context,
	req *organizationpb.GetOrganizationByCodeRequest,
) (*organizationpb.OrganizationResponse, error) {
	org, err := h.orgService.GetOrganizationByCode(ctx, req.GetCode())
	if err != nil {
		log.Printf("Failed to get organization by code: %v", err)
		return nil, status.Errorf(codes.NotFound, "organization not found")
	}
	
	return &organizationpb.OrganizationResponse{
		Organization: toProtoOrganization(org),
		Message:      "Organization retrieved successfully",
	}, nil
}

// UpdateOrganization updates an existing organization
func (h *OrganizationHandler) UpdateOrganization(
	ctx context.Context,
	req *organizationpb.UpdateOrganizationRequest,
) (*organizationpb.OrganizationResponse, error) {
	log.Printf("gRPC UpdateOrganization: id=%s, name=%s", req.GetOrgId(), req.GetName())
	
	orgID, err := uuid.Parse(req.GetOrgId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid organization ID: %v", err)
	}
	
	org, err := h.orgService.UpdateOrganization(
		ctx,
		orgID,
		req.GetName(),
		req.GetCode(),
		req.GetDescription(),
		req.GetLogo(),
		req.GetIsActive(),
	)
	
	if err != nil {
		if err == domain.ErrDuplicateOrganizationCode {
			return nil, status.Errorf(codes.AlreadyExists, "organization code already exists")
		}
		log.Printf("Failed to update organization: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update organization: %v", err)
	}
	
	return &organizationpb.OrganizationResponse{
		Organization: toProtoOrganization(org),
		Message:      "Organization updated successfully",
	}, nil
}

// DeleteOrganization deletes an organization
func (h *OrganizationHandler) DeleteOrganization(
	ctx context.Context,
	req *organizationpb.DeleteOrganizationRequest,
) (*organizationpb.DeleteOrganizationResponse, error) {
	log.Printf("gRPC DeleteOrganization: id=%s", req.GetOrgId())
	
	orgID, err := uuid.Parse(req.GetOrgId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid organization ID: %v", err)
	}
	
	// For now, use a nil UUID for requestedBy (in production, this should come from auth context)
	err = h.orgService.DeleteOrganization(ctx, orgID, uuid.Nil)
	if err != nil {
		log.Printf("Failed to delete organization: %v", err)
		return &organizationpb.DeleteOrganizationResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	
	return &organizationpb.DeleteOrganizationResponse{
		Success: true,
		Message: "Organization deleted successfully",
	}, nil
}

// ListOrganizationsByTenant lists organizations by tenant
func (h *OrganizationHandler) ListOrganizationsByTenant(
	ctx context.Context,
	req *organizationpb.ListOrganizationsByTenantRequest,
) (*organizationpb.ListOrganizationsResponse, error) {
	tenantID, err := uuid.Parse(req.GetTenantId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant ID: %v", err)
	}
	
	pagination := ports.PaginationParams{
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}
	
	orgs, paginationResult, err := h.orgService.ListOrganizationsByTenant(ctx, tenantID, pagination)
	if err != nil {
		log.Printf("Failed to list organizations by tenant: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list organizations")
	}
	
	protoOrgs := make([]*organizationpb.Organization, len(orgs))
	for i, org := range orgs {
		protoOrgs[i] = toProtoOrganization(org)
	}
	
	return &organizationpb.ListOrganizationsResponse{
		Organizations: protoOrgs,
		TotalCount:    paginationResult.TotalItems,
		Pagination:    toProtoPagination(paginationResult),
	}, nil
}

// ListAccessibleOrganizations lists organizations accessible by a user
func (h *OrganizationHandler) ListAccessibleOrganizations(
	ctx context.Context,
	req *organizationpb.ListAccessibleOrganizationsRequest,
) (*organizationpb.ListOrganizationsResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}
	
	pagination := ports.PaginationParams{
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}
	
	orgs, paginationResult, err := h.orgService.ListAccessibleOrganizations(ctx, userID, pagination)
	if err != nil {
		log.Printf("Failed to list accessible organizations: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to list accessible organizations")
	}
	
	protoOrgs := make([]*organizationpb.Organization, len(orgs))
	for i, org := range orgs {
		protoOrgs[i] = toProtoOrganization(org)
	}
	
	return &organizationpb.ListOrganizationsResponse{
		Organizations: protoOrgs,
		TotalCount:    paginationResult.TotalItems,
		Pagination:    toProtoPagination(paginationResult),
	}, nil
}

// ToggleOrganizationStatus toggles organization active status
func (h *OrganizationHandler) ToggleOrganizationStatus(
	ctx context.Context,
	req *organizationpb.ToggleOrganizationStatusRequest,
) (*organizationpb.ToggleOrganizationStatusResponse, error) {
	orgID, err := uuid.Parse(req.GetOrgId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid organization ID: %v", err)
	}
	
	requestedBy, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}
	
	org, err := h.orgService.ToggleOrganizationStatus(ctx, orgID, requestedBy)
	if err != nil {
		log.Printf("Failed to toggle organization status: %v", err)
		return &organizationpb.ToggleOrganizationStatusResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	
	statusMsg := "deactivated"
	if org.IsActive {
		statusMsg = "activated"
	}
	
	return &organizationpb.ToggleOrganizationStatusResponse{
		Success:      true,
		Message:      "Organization " + statusMsg + " successfully",
		Organization: toProtoOrganization(org),
	}, nil
}

// CheckOrganizationCode checks if organization code is available
func (h *OrganizationHandler) CheckOrganizationCode(
	ctx context.Context,
	req *organizationpb.CheckOrganizationCodeRequest,
) (*organizationpb.CheckOrganizationCodeResponse, error) {
	var excludeOrgID *uuid.UUID
	if req.GetExcludeOrgId() != "" {
		id, err := uuid.Parse(req.GetExcludeOrgId())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid exclude organization ID: %v", err)
		}
		excludeOrgID = &id
	}
	
	isAvailable, err := h.orgService.CheckOrganizationCode(ctx, req.GetCode(), excludeOrgID)
	if err != nil {
		log.Printf("Failed to check organization code: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to check organization code")
	}
	
	message := "Code is available"
	if !isAvailable {
		message = "Code is already taken"
	}
	
	return &organizationpb.CheckOrganizationCodeResponse{
		IsAvailable: isAvailable,
		Message:     message,
	}, nil
}

// Helper functions for converting domain to proto

func toProtoOrganization(org *domain.Organization) *organizationpb.Organization {
	return &organizationpb.Organization{
		OrgId:        org.OrgID.String(),
		TenantId:     org.TenantID.String(),
		Name:         org.Name,
		Code:         org.Code,
		DatabaseName: org.DatabaseName,
		Description:  org.Description,
		Logo:         org.Logo,
		IsActive:     org.IsActive,
		CreatedBy:    org.CreatedBy.String(),
		CreatedAt:    timestamppb.New(org.CreatedAt),
		UpdatedAt:    timestamppb.New(org.UpdatedAt),
	}
}

func toProtoPagination(p *ports.PaginationResult) *organizationpb.PaginationMetadata {
	return &organizationpb.PaginationMetadata{
		CurrentPage: p.CurrentPage,
		PageSize:    p.PageSize,
		TotalItems:  p.TotalItems,
		TotalPages:  p.TotalPages,
	}
}
