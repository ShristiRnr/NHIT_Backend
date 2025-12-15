package grpc

import (
	"context"
	"fmt"

	"strings"

	vendorpb "github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb"
	projectpb "github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/ports"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// VendorGRPCHandler implements the gRPC handler for vendor service
type VendorGRPCHandler struct {
	vendorpb.UnimplementedVendorServiceServer
	vendorService ports.VendorService
}

// NewVendorGRPCHandler creates a new gRPC handler
func NewVendorGRPCHandler(vendorService ports.VendorService) vendorpb.VendorServiceServer {
	return &VendorGRPCHandler{
		vendorService: vendorService,
	}
}

// CreateVendor creates a new vendor
func (h *VendorGRPCHandler) CreateVendor(ctx context.Context, req *vendorpb.CreateVendorRequest) (*vendorpb.VendorResponse, error) {
	// Extract tenant_id and user_id from JWT context (ALWAYS from JWT)
	var tenantID, userID string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if tenantIDs := md.Get("tenant_id"); len(tenantIDs) > 0 {
			tenantID = tenantIDs[0]
		}
		if userIDs := md.Get("user_id"); len(userIDs) > 0 {
			userID = userIDs[0]
		}
	}

	// tenant_id and user_id MUST come from JWT
	if tenantID == "" {
		return nil, status.Error(codes.Unauthenticated, "tenant_id not found in JWT")
	}
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user_id not found in JWT")
	}

	// Parse tenant ID
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tenant_id in JWT")
	}

	// Parse created by (from JWT user_id)
	createdByUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id in JWT")
	}

	// Build create params
	params := domain.CreateVendorParams{
		TenantID:               tenantUUID,
		VendorName:             req.VendorName,
		VendorEmail:            req.VendorEmail,
		VendorMobile:           req.VendorMobile,
		AccountType:            convertAccountType(req.AccountType),
		VendorNickName:         req.VendorNickName,
		ActivityType:           req.ActivityType,
		Email:                  req.Email,
		Mobile:                 req.Mobile,
		GSTIN:                  req.Gstin,
		PAN:                    req.Pan,
		PIN:                    req.Pin,
		CountryID:              req.CountryId,
		StateID:                req.StateId,
		CityID:                 req.CityId,
		CountryName:            req.CountryName,
		StateName:              req.StateName,
		CityName:               req.CityName,
		MSMEClassification:     convertMSMEClassification(req.MsmeClassification),
		MSME:                   req.Msme,
		MSMERegistrationNumber: req.MsmeRegistrationNumber,
		MSMEStartDate:          nil, // TODO: convert if needed
		MSMEEndDate:            nil, // TODO: convert if needed
		MaterialNature:         req.MaterialNature,
		GSTDefaulted:           req.GstDefaulted,
		Section206ABVerified:   req.Section_206AbVerified,
		BeneficiaryName:        req.BeneficiaryName,
		RemarksAddress:         req.RemarksAddress,
		CommonBankDetails:      req.CommonBankDetails,
		IncomeTaxType:          req.IncomeTaxType,
		ProjectID:              req.Project,
		Status:                 convertVendorStatus(req.Status),
		FromAccountType:        req.FromAccountType,
		AccountName:            req.AccountName,
		ShortName:              req.ShortName,
		Parent:                 req.Parent,
		FilePaths:              req.FilePaths,
		CreatedBy:              createdByUUID,
		AccountNumber:          req.AccountNumber,
		NameOfBank:             req.NameOfBank,
		IFSCCode:               req.IfscCode,
		IFSCCodeID:             req.IfscCodeId,
	}

	vendor, err := h.vendorService.CreateVendor(ctx, params)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &vendorpb.VendorResponse{
		Vendor: toProtoVendor(vendor),
	}, nil
}

// ListVendors lists vendors by tenant (from JWT)
func (h *VendorGRPCHandler) ListVendors(ctx context.Context, req *vendorpb.ListVendorsRequest) (*vendorpb.ListVendorsResponse, error) {
	// Extract tenant_id and org_id from JWT context
	var tenantID, orgID string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if tenantIDs := md.Get("tenant_id"); len(tenantIDs) > 0 {
			tenantID = tenantIDs[0]
		}
		if orgIDs := md.Get("org_id"); len(orgIDs) > 0 {
			orgID = orgIDs[0]
		}
	}

	// tenant_id MUST come from JWT
	if tenantID == "" {
		return nil, status.Error(codes.Unauthenticated, "tenant_id not found in JWT")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tenant_id in JWT")
	}

	// Build filters with org_id from JWT
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	filters := ports.VendorListFilters{
		IsActive: req.IsActive,
		OrganizationID: &orgID, // Filter by logged-in user's org
		Project:  req.Project,
		Search:   req.Search,
		Limit:    limit,
		Offset:   int(req.Offset),
	}

	// Filter by Organization ID (if present in JWT) and Org has projects
	// If Org is present, we MUST fetch projects for this org to filter vendors
	if orgID != "" {
		fmt.Printf("🔍 Filtering vendors by Organization ID: %s (fetching projects...)\n", orgID)

		// 1. Connect to project-service (Port 50057)
		// Note: Ideally use a shared client, but for now we dial locally to ensure isolation
		conn, err := grpc.Dial("localhost:50057", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Printf("⚠️ Failed to connect to project service for filtering: %v\n", err)
			return nil, status.Error(codes.Internal, "failed to connect to project service for filtering")
		}
		defer conn.Close()
		projectClient := projectpb.NewProjectServiceClient(conn)

		// 2. Fetch projects
		// Forward metadata for auth
		outCtx := metadata.NewOutgoingContext(ctx, md) 
		projectResp, err := projectClient.ListProjectsByOrganization(outCtx, &projectpb.ListProjectsByOrganizationRequest{
			OrgId: orgID,
		})
		
		if err != nil {
			fmt.Printf("⚠️ Failed to fetch projects for filtering: %v\n", err)
			return nil, status.Error(codes.Internal, "failed to fetch organization projects")
		}

		// 3. Extract IDs
		var projectIDs []string
		if projectResp != nil && len(projectResp.Projects) > 0 {
			for _, p := range projectResp.Projects {
				projectIDs = append(projectIDs, p.ProjectId)
			}
			fmt.Printf("✅ Found %d projects for Org: %v\n", len(projectIDs), projectIDs)
			filters.ProjectIDs = projectIDs
		} else {
			fmt.Printf("⚠️ No projects found for Org: %s. Proceeding without project filter (showing unassigned/all).\n", orgID)
			// Do NOT return empty. Allow repo to handle it (it will show unassigned vendors or all vendors)
		}
	}

	vendors, total, err := h.vendorService.ListVendors(ctx, tenantUUID, filters)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoVendors := make([]*vendorpb.Vendor, len(vendors))
	for i, v := range vendors {
		protoVendors[i] = toProtoVendor(v)
	}

	return &vendorpb.ListVendorsResponse{
		Vendors:    protoVendors,
		TotalCount: total,
	}, nil
}

// GenerateVendorCode generates a vendor code
func (h *VendorGRPCHandler) GenerateVendorCode(ctx context.Context, req *vendorpb.GenerateVendorCodeRequest) (*vendorpb.GenerateVendorCodeResponse, error) {
	if req.VendorName == "" {
		return nil, status.Error(codes.InvalidArgument, "vendor_name is required")
	}

	accountType := convertAccountType(req.AccountType)

	code, err := h.vendorService.GenerateVendorCode(ctx, req.VendorName, &accountType)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &vendorpb.GenerateVendorCodeResponse{
		VendorCode: code,
	}, nil
}

// GetProjectsDropdown fetches projects from project-service
func (h *VendorGRPCHandler) GetProjectsDropdown(ctx context.Context, req *vendorpb.GetProjectsDropdownRequest) (*vendorpb.GetProjectsDropdownResponse, error) {
	// Extract org_id from context metadata
	var orgID string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if orgIDs := md.Get("org_id"); len(orgIDs) > 0 {
			orgID = orgIDs[0]
		}
	}
	
	if orgID == "" {
		fmt.Printf("⚠️ No org_id found in metadata for GetProjectsDropdown\n")
		return &vendorpb.GetProjectsDropdownResponse{
			Projects: []*vendorpb.ProjectDropdownItem{},
		}, nil
	}
	
	// Connect to project-service at port 50057
	conn, err := grpc.Dial("localhost:50057", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Failed to connect to project service: %v\n", err)
		return &vendorpb.GetProjectsDropdownResponse{
			Projects: []*vendorpb.ProjectDropdownItem{},
		}, nil
	}
	defer conn.Close()
	
	// Create project service client
	projectClient := projectpb.NewProjectServiceClient(conn)
	
	// Forward metadata
	outCtx := metadata.NewOutgoingContext(ctx, md)
	
	// Fetch projects by organization ID
	projectResp, err := projectClient.ListProjectsByOrganization(outCtx, &projectpb.ListProjectsByOrganizationRequest{
		OrgId: orgID,
	})
	if err != nil {
		fmt.Printf("Failed to fetch projects from project-service: %v\n", err)
		return &vendorpb.GetProjectsDropdownResponse{
			Projects: []*vendorpb.ProjectDropdownItem{},
		}, nil
	}
	
	// Convert projects to dropdown items
	projects := make([]*vendorpb.ProjectDropdownItem, 0)
	if projectResp != nil && len(projectResp.Projects) > 0 {
		fmt.Printf("✅ Found %d projects from project-service for org_id: %s\n", len(projectResp.Projects), orgID)
		for _, project := range projectResp.Projects {
			projects = append(projects, &vendorpb.ProjectDropdownItem{
				Id:   project.ProjectId,
				Name: project.ProjectName,
			})
		}
	} else {
		fmt.Printf("⚠️ No projects found in project-service for org_id: %s\n", orgID)
	}

	return &vendorpb.GetProjectsDropdownResponse{
		Projects: projects,
	}, nil
}

// GetVendor retrieves a vendor by ID
func (h *VendorGRPCHandler) GetVendor(ctx context.Context, req *vendorpb.GetVendorRequest) (*vendorpb.VendorResponse, error) {
	// Extract tenant_id from JWT
	var tenantID string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if tenantIDs := md.Get("tenant_id"); len(tenantIDs) > 0 {
			tenantID = tenantIDs[0]
		}
	}
	
	if tenantID == "" {
		return nil, status.Error(codes.Unauthenticated, "tenant_id not found in JWT")
	}
	
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tenant_id")
	}
	
	vendorUUID, err := uuid.Parse(req.VendorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vendor_id")
	}
	
	vendor, err := h.vendorService.GetVendorByID(ctx, tenantUUID, vendorUUID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "vendor not found")
	}
	
	return &vendorpb.VendorResponse{
		Vendor: toProtoVendor(vendor),
	}, nil
}

func (h *VendorGRPCHandler) GetVendorByCode(ctx context.Context, req *vendorpb.GetVendorByCodeRequest) (*vendorpb.VendorResponse, error) {
	// Extract tenant_id from JWT
	var tenantID string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if tenantIDs := md.Get("tenant_id"); len(tenantIDs) > 0 {
			tenantID = tenantIDs[0]
		}
	}
	
	if tenantID == "" {
		return nil, status.Error(codes.Unauthenticated, "tenant_id not found in JWT")
	}
	
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tenant_id")
	}
	
	if req.VendorCode == "" {
		return nil, status.Error(codes.InvalidArgument, "vendor_code is required")
	}
	
	vendor, err := h.vendorService.GetVendorByCode(ctx, tenantUUID, req.VendorCode)
	if err != nil {
		return nil, status.Error(codes.NotFound, "vendor not found")
	}
	
	return &vendorpb.VendorResponse{
		Vendor: toProtoVendor(vendor),
	}, nil
}

func (h *VendorGRPCHandler) UpdateVendor(ctx context.Context, req *vendorpb.UpdateVendorRequest) (*vendorpb.VendorResponse, error) {
	// Extract tenant_id from JWT
	var tenantID string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if tenantIDs := md.Get("tenant_id"); len(tenantIDs) > 0 {
			tenantID = tenantIDs[0]
		}
	}
	
	if tenantID == "" {
		return nil, status.Error(codes.Unauthenticated, "tenant_id not found in JWT")
	}
	
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tenant_id")
	}
	
	vendorUUID, err := uuid.Parse(req.VendorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vendor_id")
	}
	
	params := domain.UpdateVendorParams{
		VendorName:             req.VendorName,
		VendorEmail:            req.VendorEmail,
		VendorMobile:           req.VendorMobile,
		PAN:                    req.Pan,
		BeneficiaryName:        req.BeneficiaryName,
		// Add new fields
		MSME:                   req.Msme,
		MSMERegistrationNumber: req.MsmeRegistrationNumber,
		MaterialNature:         req.MaterialNature,
		GSTDefaulted:           req.GstDefaulted,
		Section206ABVerified:   req.Section_206AbVerified,
		IncomeTaxType:          req.IncomeTaxType,
		ProjectID:              req.Project,
		AccountName:            req.AccountName,
		RemarksAddress:         req.RemarksAddress,
	}

	// Handle MSME Classification
	if req.MsmeClassification != nil {
		val := convertMSMEClassification(*req.MsmeClassification)
		params.MSMEClassification = &val
	}
	// Handle Account Type from request if needed
	if req.FromAccountType != nil {
		val := convertAccountType(*req.FromAccountType)
		params.AccountType = &val
	}
	// Handle Status
	if req.Status != nil {
		val := convertVendorStatus(*req.Status)
		params.Status = &val
	}
	
	vendor, err := h.vendorService.UpdateVendor(ctx, tenantUUID, vendorUUID, params)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &vendorpb.VendorResponse{
		Vendor: toProtoVendor(vendor),
	}, nil
}

func (h *VendorGRPCHandler) DeleteVendor(ctx context.Context, req *vendorpb.DeleteVendorRequest) (*emptypb.Empty, error) {
	// Extract tenant_id from JWT
	var tenantID string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if tenantIDs := md.Get("tenant_id"); len(tenantIDs) > 0 {
			tenantID = tenantIDs[0]
		}
	}
	
	if tenantID == "" {
		return nil, status.Error(codes.Unauthenticated, "tenant_id not found in JWT")
	}
	
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tenant_id")
	}
	
	vendorUUID, err := uuid.Parse(req.VendorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vendor_id")
	}
	
	if err := h.vendorService.DeleteVendor(ctx, tenantUUID, vendorUUID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &emptypb. Empty{}, nil
}

func (h *VendorGRPCHandler) UpdateVendorCode(ctx context.Context, req *vendorpb.UpdateVendorCodeRequest) (*vendorpb.VendorResponse, error) {
	// Extract tenant_id from JWT
	var tenantID string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if tenantIDs := md.Get("tenant_id"); len(tenantIDs) > 0 {
			tenantID = tenantIDs[0]
		}
	}
	
	if tenantID == "" {
		return nil, status.Error(codes.Unauthenticated, "tenant_id not found in JWT")
	}
	
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tenant_id")
	}
	
	vendorUUID, err := uuid.Parse(req.VendorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vendor_id")
	}
	
	if req.VendorCode == "" {
		return nil, status.Error(codes.InvalidArgument, "vendor_code is required")
	}
	
	vendor, err := h.vendorService.UpdateVendorCode(ctx, tenantUUID, vendorUUID, req.VendorCode)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &vendorpb.VendorResponse{
		Vendor: toProtoVendor(vendor),
	}, nil
}

func (h *VendorGRPCHandler) RegenerateVendorCode(ctx context.Context, req *vendorpb.RegenerateVendorCodeRequest) (*vendorpb.VendorResponse, error) {
	// Extract tenant_id from JWT
	var tenantID string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if tenantIDs := md.Get("tenant_id"); len(tenantIDs) > 0 {
			tenantID = tenantIDs[0]
		}
	}
	
	if tenantID == "" {
		return nil, status.Error(codes.Unauthenticated, "tenant_id not found in JWT")
	}
	
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tenant_id")
	}
	
	vendorUUID, err := uuid.Parse(req.VendorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vendor_id")
	}
	
	vendor, err := h.vendorService.RegenerateVendorCode(ctx, tenantUUID, vendorUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &vendorpb.VendorResponse{
		Vendor: toProtoVendor(vendor),
	}, nil
}

// Vendor Account methods
func (h *VendorGRPCHandler) CreateVendorAccount(ctx context.Context, req *vendorpb.CreateVendorAccountRequest) (*vendorpb.VendorAccountResponse, error) {
	// Extract user_id from JWT for created_by
	var userID string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if userIDs := md.Get("user_id"); len(userIDs) > 0 {
			userID = userIDs[0]
		}
	}
	
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user_id not found in JWT")
	}
	
	vendorUUID, err := uuid.Parse(req.VendorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vendor_id")
	}
	
	createdByUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id in JWT")
	}
	
	params := domain.CreateVendorAccountParams{
		VendorID:      vendorUUID,
		AccountName:   req.AccountName,
		AccountNumber: req.AccountNumber,
		NameOfBank:    req.NameOfBank,
		IFSCCode:      req.IfscCode,
		IsPrimary:     req.IsPrimary,
		CreatedBy:     createdByUUID,
	}
	
	// Add optional fields
	if req.AccountType != nil {
		params.AccountType = req.AccountType
	}
	if req.BranchName != nil {
		params.BranchName = req.BranchName
	}
	if req.SwiftCode != nil {
		params.SwiftCode = req.SwiftCode
	}
	if req.Remarks != nil {
		params.Remarks = req.Remarks
	}
	
	account, err := h.vendorService.CreateVendorAccount(ctx, params)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &vendorpb.VendorAccountResponse{
		Account: toProtoVendorAccount(account),
	}, nil
}

func (h *VendorGRPCHandler) GetVendorAccounts(ctx context.Context, req *vendorpb.GetVendorAccountsRequest) (*vendorpb.GetVendorAccountsResponse, error) {
	vendorUUID, err := uuid.Parse(req.VendorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vendor_id")
	}
	
	accounts, err := h.vendorService.GetVendorAccounts(ctx, vendorUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	protoAccounts := make([]*vendorpb.VendorAccount, len(accounts))
	for i, acc := range accounts {
		protoAccounts[i] = toProtoVendorAccount(acc)
	}
	
	return &vendorpb.GetVendorAccountsResponse{
		Accounts: protoAccounts,
	}, nil
}

func (h *VendorGRPCHandler) GetVendorBankingDetails(ctx context.Context, req *vendorpb.GetVendorBankingDetailsRequest) (*vendorpb.BankingDetailsResponse, error) {
	vendorUUID, err := uuid.Parse(req.VendorId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vendor_id")
	}
	
	var accountUUID *uuid.UUID
	if req.AccountId != nil && *req.AccountId != "" {
		accUUID, err := uuid.Parse(*req.AccountId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid account_id")
		}
		accountUUID = &accUUID
	}
	
	bankingDetails, err := h.vendorService.GetVendorBankingDetails(ctx, vendorUUID, accountUUID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "banking details not found")
	}
	
	return &vendorpb.BankingDetailsResponse{
		BankingDetails: &vendorpb.BankingDetails{
			AccountName:   bankingDetails.AccountName,
			AccountNumber: bankingDetails.AccountNumber,
			AccountType:   bankingDetails.AccountType,
			NameOfBank:    bankingDetails.NameOfBank,
			BranchName:    bankingDetails.BranchName,
			IfscCode:      bankingDetails.IFSCCode,
			SwiftCode:     bankingDetails.SwiftCode,
			Remarks:       bankingDetails.Remarks,
		},
	}, nil
}

func (h *VendorGRPCHandler) UpdateVendorAccount(ctx context.Context, req *vendorpb.UpdateVendorAccountRequest) (*vendorpb.VendorAccountResponse, error) {
	// Extract user_id from JWT for updated_by
	var userID string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if userIDs := md.Get("user_id"); len(userIDs) > 0 {
			userID = userIDs[0]
		}
	}
	
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user_id not found in JWT")
	}
	
	accountUUID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid account_id")
	}
	
	params := domain.UpdateVendorAccountParams{
		AccountName:   req.AccountName,
		AccountNumber: req.AccountNumber,
		AccountType:   req.AccountType,
		NameOfBank:    req.NameOfBank,
		BranchName:    req.BranchName,
		IFSCCode:      req.IfscCode,
		SwiftCode:     req.SwiftCode,
		IsPrimary:     req.IsPrimary,
		IsActive:      req.IsActive,
		Remarks:       req.Remarks,
	}
	
	account, err := h.vendorService.UpdateVendorAccount(ctx, accountUUID, params)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &vendorpb.VendorAccountResponse{
		Account: toProtoVendorAccount(account),
	}, nil
}

func (h *VendorGRPCHandler) DeleteVendorAccount(ctx context.Context, req *vendorpb.DeleteVendorAccountRequest) (*emptypb.Empty, error) {
	accountUUID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid account_id")
	}
	
	if err := h.vendorService.DeleteVendorAccount(ctx, accountUUID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &emptypb.Empty{}, nil
}

func (h *VendorGRPCHandler) ToggleAccountStatus(ctx context.Context, req *vendorpb.ToggleAccountStatusRequest) (*vendorpb.VendorAccountResponse, error) {
	accountUUID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid account_id")
	}
	
	account, err := h.vendorService.ToggleAccountStatus(ctx, accountUUID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &vendorpb.VendorAccountResponse{
		Account: toProtoVendorAccount(account),
	}, nil
}

// Helper functions
func toProtoVendor(v *domain.Vendor) *vendorpb.Vendor {
	vendor := &vendorpb.Vendor{
		Id:                    v.ID.String(),
		TenantId:              v.TenantID.String(),
		VendorCode:            v.VendorCode,
		VendorName:            v.VendorName,
		VendorEmail:           v.VendorEmail,
		VendorMobile:          v.VendorMobile,
		AccountType:           convertAccountType(v.AccountType),
		VendorNickName:        v.VendorNickName,
		ActivityType:          v.ActivityType,
		Email:                 v.Email,
		Mobile:                v.Mobile,
		Gstin:                 v.GSTIN,
		Pan:                   v.PAN,
		Pin:                   v.PIN,
		CountryId:             v.CountryID,
		StateId:               v.StateID,
		CityId:                v.CityID,
		CountryName:           v.CountryName,
		StateName:             v.StateName,
		CityName:              v.CityName,
		MsmeClassification:    v.MSMEClassification, // Already string in domain, just assign directly if normalized. Or use helper if needing normalization.
		Msme:                  v.MSME,
		MsmeRegistrationNumber: v.MSMERegistrationNumber,
		MaterialNature:        v.MaterialNature,
		GstDefaulted:          v.GSTDefaulted,
		Section_206AbVerified: v.Section206ABVerified,
		BeneficiaryName:       v.BeneficiaryName,
		RemarksAddress:        v.RemarksAddress,
		CommonBankDetails:     v.CommonBankDetails,
		IncomeTaxType:         v.IncomeTaxType,
		Project:               v.ProjectID,
		FromAccountType:       v.FromAccountType,
		AccountName:           v.AccountName,
		ShortName:             v.ShortName,
		Parent:                v.Parent,
		FilePaths:             v.FilePaths,
		CodeAutoGenerated:     v.CodeAutoGenerated,
		CreatedBy:             v.CreatedBy.String(),
		CreatedAt:             timestamppb.New(v.CreatedAt),
		UpdatedAt:             timestamppb.New(v.UpdatedAt),
		AccountNumber:         v.AccountNumber,
		NameOfBank:            v.NameOfBank,
		IfscCode:              v.IFSCCode,
		IfscCodeId:            v.IFSCCodeID,
	}
	
	return vendor
}

func convertAccountType(at string) string {
	// Normalize to uppercase
	at = strings.ToUpper(at)
	
	switch at {
	case "INTERNAL":
		return "INTERNAL"
	case "EXTERNAL":
		return "EXTERNAL"
	default:
		// Default to EXTERNAL if unspecified or invalid, matching basic default logic
		// Or return empty if strict validation is needed. 
		// Given proto default is 0 (UNSPECIFIED), but DB usually needs a value.
		// If the user sends "Internal" we want "INTERNAL".
		// If they send "Foo", we probably default to "EXTERNAL" or just return "EXTERNAL" as fallthrough
		// Check against proto map if we want to be strict, but for now simple normalization:
		return "EXTERNAL"
	}
}

func convertMSMEClassification(mc string) string {
	mc = strings.ToUpper(mc)
	switch mc {
	case "MICRO":
		return "MICRO"
	case "SMALL":
		return "SMALL"
	case "MEDIUM":
		return "MEDIUM"
	default:
		// If empty or invalid, return implicit default or empty string
		return "MSME_CLASSIFICATION_UNSPECIFIED"
	}
}

func convertVendorStatus(vs string) string {
	vs = strings.ToUpper(vs)
	switch vs {
	case "ACTIVE":
		return "ACTIVE"
	case "INACTIVE":
		return "INACTIVE"
	default:
		return "VENDOR_STATUS_UNSPECIFIED"
	}
}

func convertStringToMSMEClassification(s string) string {
	s = strings.ToUpper(s)
	switch s {
	case "MICRO":
		return "MICRO"
	case "SMALL":
		return "SMALL"
	case "MEDIUM":
		return "MEDIUM"
	default:
		return "MSME_CLASSIFICATION_UNSPECIFIED"
	}
}
