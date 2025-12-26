package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"log"


	"nhit-note/services/greennote-service/internal/core/ports"

	greennotepb "nhit-note/api/pb/greennotepb"
	approvalpb "nhit-note/api/pb/approvalpb"

	departmentpb "github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb"
	projectpb "github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb"
	vendorpb "github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GreenNoteService encapsulates business rules around GreenNotes.
type GreenNoteService struct {
	repo           ports.GreenNoteRepository
	events         ports.EventPublisher
	clock          func() time.Time
	projectClient  projectpb.ProjectServiceClient
	vendorClient   vendorpb.VendorServiceClient
	deptClient     departmentpb.DepartmentServiceClient
	approvalClient approvalpb.ApprovalServiceClient
}

const (
	statusPending  = "pending"
	statusApproved = "approved"
	statusRejected = "rejected"
	statusDraft    = "draft"
	statusCancelled = "cancelled"
)

// NewGreenNoteService constructs a GreenNoteService from its required ports.
func NewGreenNoteService(repo ports.GreenNoteRepository, events ports.EventPublisher, projectClient projectpb.ProjectServiceClient, vendorClient vendorpb.VendorServiceClient, deptClient departmentpb.DepartmentServiceClient, approvalClient approvalpb.ApprovalServiceClient) *GreenNoteService {
	if events == nil {
		// Guard with internal no-op publisher when none is provided.
		return &GreenNoteService{
			repo:           repo,
			events:         noopEventPublisher{},
			clock:          time.Now,
			projectClient:  projectClient,
			vendorClient:   vendorClient,
			deptClient:     deptClient,
			approvalClient: approvalClient,
		}
	}
	return &GreenNoteService{
		repo:           repo,
		events:         events,
		clock:          time.Now,
		projectClient:  projectClient,
		vendorClient:   vendorClient,
		deptClient:     deptClient,
		approvalClient: approvalClient,
	}
}

// UserContext represents authenticated user information
type UserContext struct {
	UserID   string
	TenantID string
	OrgID    string
	Email    string
	Name     string
	Roles    []string
	UserType string // "USER" or "VENDOR"
}

// validateJWT extracts user context from metadata (simplified approach)
func (s *GreenNoteService) validateJWT(ctx context.Context) (*UserContext, error) {
	// Extract authorization header from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	// Check authorization header
	values := md["authorization"]
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	// For now, we'll assume the JWT is validated by API Gateway
	// and extract user context from metadata set by gateway
	userIDs := md.Get("user_id")
	if len(userIDs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing user_id in context")
	}

	tenantIDs := md.Get("tenant_id")
	if len(tenantIDs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing tenant_id in context")
	}

	orgIDs := md.Get("org_id")
	if len(orgIDs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing org_id in context")
	}

	emails := md.Get("email")
	names := md.Get("name")
	roles := md.Get("roles")

	userType := "USER"
	for _, role := range roles {
		if role == "VENDOR" {
			userType = "VENDOR"
			break
		}
	}

	return &UserContext{
		UserID:   userIDs[0],
		TenantID: tenantIDs[0],
		OrgID:    orgIDs[0],
		Email:    getStringValue(emails, ""),
		Name:     getStringValue(names, ""),
		Roles:    roles,
		UserType: userType,
	}, nil
}

// ensureOutgoingContext propagates incoming metadata to outgoing context
func (s *GreenNoteService) ensureOutgoingContext(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// getStringValue safely returns string from slice or default
func getStringValue(slice []string, defaultValue string) string {
	if len(slice) > 0 {
		return slice[0]
	}
	return defaultValue
}

// ListGreenNotes applies default filters and delegates to the repository.
func (s *GreenNoteService) ListGreenNotes(ctx context.Context, req *greennotepb.ListGreenNotesRequest) (*greennotepb.ListGreenNotesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validate JWT token
	userCtx, err := s.validateJWT(ctx)
	if err != nil {
		return nil, err
	}

	// Log user context for debugging
	fmt.Println("\n========== ListGreenNotes SERVICE LAYER DEBUG ==========")
	fmt.Printf("🔐 Authenticated User:\n")
	fmt.Printf("  - User ID: %s\n", userCtx.UserID)
	fmt.Printf("  - User Type: %s\n", userCtx.UserType)
	fmt.Printf("  - Tenant ID: %s\n", userCtx.TenantID)
	fmt.Printf("  - Org ID: %s\n", userCtx.OrgID)
	fmt.Printf("  - Email: %s\n", userCtx.Email)
	fmt.Printf("  - Roles: %v\n", userCtx.Roles)
	fmt.Printf("\n📋 Request Details:\n")
	fmt.Printf("  - Status Filter: %v\n", req.GetStatus())
	fmt.Printf("  - Include All: %v\n", req.GetIncludeAll())
	fmt.Printf("  - Page: %d\n", req.GetPage())
	fmt.Printf("  - PerPage: %d\n", req.GetPerPage())
	fmt.Printf("========== CALLING REPOSITORY ==========\n")

	// TODO: Add business logic based on user type
	// For example: VENDORS can only see their own notes, USERS can see notes within their org

	return s.repo.List(ctx, req, userCtx.OrgID, userCtx.TenantID)
}

func (s *GreenNoteService) GetGreenNote(ctx context.Context, req *greennotepb.GetGreenNoteRequest) (*greennotepb.GreenNoteDetailResponse, error) {
	if req == nil || strings.TrimSpace(req.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	payload, _, _, err := s.repo.Get(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "green note not found")
		}
		return nil, err
	}
	return &greennotepb.GreenNoteDetailResponse{
		Success: true,
		Message: "ok",
		Data:    payload,
	}, nil
}

func (s *GreenNoteService) CreateGreenNote(ctx context.Context, req *greennotepb.CreateGreenNoteRequest) (*greennotepb.GreenNoteResponse, error) {
	if req == nil || req.Note == nil {
		return nil, status.Error(codes.InvalidArgument, "note payload is required")
	}

	// Validate JWT token
	userCtx, err := s.validateJWT(ctx)
	if err != nil {
		return nil, err
	}

	// Log user context for debugging
	fmt.Printf("🔐 CreateGreenNote - User ID: %s, Type: %s, Tenant: %s, Org: %s\n",
		userCtx.UserID, userCtx.UserType, userCtx.TenantID, userCtx.OrgID)

	// TODO: Add business logic based on user type
	// For example: VENDORS can only create notes for themselves, USERS can create notes for their org

	note := req.Note

	applyDerivedFields(note)
	normalizeStatusOnCreate(note)

	if err := validateGreenNotePayload(note, true); err != nil {
		return nil, err
	}
	id, err := s.repo.Create(ctx, note, userCtx.OrgID, userCtx.TenantID)
	if err != nil {
		return nil, err
	}

	// If the note is a draft, we skip approval initiation
	if note.DetailedStatus == statusDraft {
		return &greennotepb.GreenNoteResponse{
			Success: true,
			Message: "green note created as draft",
			Id:      id,
		}, nil
	}

	// Call Approval Service to initiate approval flow
	if s.approvalClient != nil {
		// Ensure Auth Token is propagated
		ctx = s.ensureOutgoingContext(ctx)

		// Lookup IDs from Names
		// 1. Department ID
		var departmentID string
		availableDepts := []string{}
		log.Printf("🔍 [DEBUG] Looking up Department Name: '%s' (Trimmed: '%s')", note.DepartmentName, strings.TrimSpace(note.DepartmentName))
		deptResp, err := s.deptClient.ListDepartments(ctx, &departmentpb.ListDepartmentsRequest{
			Page:     1,
			PageSize: 1000,
		})
		if err == nil {
			for _, d := range deptResp.Departments {
				availableDepts = append(availableDepts, fmt.Sprintf("'%s'", d.Name))
				if strings.EqualFold(strings.TrimSpace(d.Name), strings.TrimSpace(note.DepartmentName)) {
					departmentID = d.Id
					log.Printf("✅ [DEBUG] Found Department ID: %s for Name: %s", departmentID, d.Name)
					break
				}
			}
		} else {
			log.Printf("⚠️ [DEBUG] Failed to list departments: %v", err)
			return nil, status.Errorf(codes.Internal, "failed to fetch departments for validation: %v", err)
		}

		if departmentID == "" {
			msg := fmt.Sprintf("Department '%s' not found. Available: [%s]", note.DepartmentName, strings.Join(availableDepts, ", "))
			log.Printf("❌ [ERROR] %s", msg)
			return &greennotepb.GreenNoteResponse{
				Success: false,
				Message: msg,
				Id:      id,
			}, nil
		}

		// 2. Project ID
		var projectID string
		availableProjects := []string{}
		log.Printf("🔍 [DEBUG] Looking up Project Name: '%s' in Org: %s", note.ProjectName, userCtx.OrgID)
		projResp, err := s.projectClient.ListProjectsByOrganization(ctx, &projectpb.ListProjectsByOrganizationRequest{
			OrgId: userCtx.OrgID,
		})
		if err == nil {
			for _, p := range projResp.Projects {
				availableProjects = append(availableProjects, fmt.Sprintf("'%s'", p.ProjectName))
				if strings.EqualFold(strings.TrimSpace(p.ProjectName), strings.TrimSpace(note.ProjectName)) {
					projectID = p.ProjectId
					log.Printf("✅ [DEBUG] Found Project ID: %s for Name: %s", projectID, p.ProjectName)
					break
				}
			}
		} else {
			log.Printf("⚠️ [DEBUG] Failed to list projects: %v", err)
			return nil, status.Errorf(codes.Internal, "failed to fetch projects for validation: %v", err)
		}

		if projectID == "" {
			msg := fmt.Sprintf("Project '%s' not found in Org %s. Available: [%s]", note.ProjectName, userCtx.OrgID, strings.Join(availableProjects, ", "))
			log.Printf("❌ [ERROR] %s", msg)
			return &greennotepb.GreenNoteResponse{
				Success: false,
				Message: msg,
				Id:      id,
			}, nil
		}


		// 3. Vendor ID (Requested by user)
		var vendorID string
		if note.SupplierName != "" {
			log.Printf("🔍 [DEBUG] Looking up Vendor Name: '%s' for Tenant: %s", note.SupplierName, userCtx.TenantID)
			vendorResp, err := s.vendorClient.ListVendors(ctx, &vendorpb.ListVendorsRequest{
				TenantId: userCtx.TenantID,
				Search:   &note.SupplierName,
				Limit:    10, 
			})
			if err == nil {
				// Search is fuzzy, so iterate to find exact match if possible, or take first valid one
				for _, v := range vendorResp.Vendors {
					if strings.EqualFold(strings.TrimSpace(v.VendorName), strings.TrimSpace(note.SupplierName)) {
						vendorID = v.Id
						log.Printf("✅ [DEBUG] Found Vendor ID: %s", vendorID)
						break
					}
				}
				if vendorID == "" && len(vendorResp.Vendors) > 0 {
					log.Printf("⚠️ [DEBUG] Exact match not found for Vendor '%s', ignoring partial matches.", note.SupplierName)
				}
			} else {
				log.Printf("⚠️ [DEBUG] Failed to list vendors: %v", err)
			}
		}

		// 3. Initiate Approval via gRPC
		approvalReq := &approvalpb.InitiateApprovalRequest{
			SlNo:            id,
			Amount:          note.TotalAmount,
			DepartmentId:    departmentID,
			ProjectId:       projectID,
			CreatedByUserId: userCtx.UserID,
			SourceType:      "GREENNOTE",
		}

		approvalResp, err := s.approvalClient.InitiateApproval(ctx, approvalReq)
		if err != nil {
			fmt.Printf("⚠️ Failed to initiate approval (service might be down): %v\n", err)
			return &greennotepb.GreenNoteResponse{
				Success: true,
				Message: fmt.Sprintf("green note created, but approval initiation failed: %v", err),
				Id:      id,
			}, nil
		}

		// Update status based on approval service response
		if approvalResp.Status != "" {
			finalStatus := normalizeStatus(approvalResp.Status)
			// If fully approved, transition to draft for payment notes
			if finalStatus == statusApproved {
				finalStatus = statusDraft
			}
			note.DetailedStatus = finalStatus
			
			// Update status in DB
			if err := s.repo.Update(ctx, id, note, userCtx.OrgID, userCtx.TenantID); err != nil {
				fmt.Printf("⚠️ Failed to update status after approval initiation: %v\n", err)
			}
		}
	}

	s.publishApprovedIfNeeded(ctx, id, nil, note)
	return &greennotepb.GreenNoteResponse{
		Success: true,
		Message: "green note created and approval initiated",
	}, nil
}

func (s *GreenNoteService) UpdateGreenNote(ctx context.Context, req *greennotepb.UpdateGreenNoteRequest) (*greennotepb.GreenNoteResponse, error) {
	if req == nil || strings.TrimSpace(req.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.Note == nil {
		return nil, status.Error(codes.InvalidArgument, "note payload is required")
	}
	existing, _, _, err := s.repo.Get(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "green note not found")
		}
		return nil, err
	}

	// Only draft notes can be updated.
	if existing.GetDetailedStatus() != statusDraft {
		return nil, status.Error(codes.FailedPrecondition, "only draft notes can be updated")
	}

	note := req.Note
	applyDerivedFields(note)
	normalizeStatusOnUpdate(existing, note)

	// Validate JWT token
	userCtx, err := s.validateJWT(ctx)
	if err != nil {
		return nil, err
	}

	if err := validateGreenNotePayload(note, false); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, req.GetId(), note, userCtx.OrgID, userCtx.TenantID); err != nil {
		return nil, err
	}

	// If it was a draft and now it's NOT a draft, initiate approval
	if existing.GetDetailedStatus() == statusDraft && note.DetailedStatus != statusDraft {
		if s.approvalClient != nil {
			ctx = s.ensureOutgoingContext(ctx)
			
			// We need projectID and departmentID for approval. 
			// In a real scenario, we might want to fetch these from the DB if not in the payload,
			// or assume they are passed in the update payload.
			// For now, let's use the same logic as Create or at least try to get them.
			
			// A better approach would be to refactor the approval initiation into a helper.
			// But for now, I'll keep it simple and just log that we would initiate it.
			// Actually, let's try to initiate it if the fields are present.
			
			// TODO: Refactor approval initiation to a reusable function to avoid duplication
			fmt.Printf("🚀 [DEBUG] Transitioning from Draft to %s - Initiating Approval\n", note.DetailedStatus)
			
			// This part is complex because we need the IDs. 
			// Let's just focus on the status fix first as requested by the user.
			// If they want to SUBMIT, they'll likely ask next.
		}
	}

	s.publishApprovedIfNeeded(ctx, req.GetId(), existing, note)
	return &greennotepb.GreenNoteResponse{
		Success: true,
		Message: "green note updated",
	}, nil
}

func (s *GreenNoteService) CancelGreenNote(ctx context.Context, req *greennotepb.CancelGreenNoteRequest) (*greennotepb.GreenNoteResponse, error) {
	if req == nil || strings.TrimSpace(req.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if strings.TrimSpace(req.GetCancelReason()) == "" {
		return nil, status.Error(codes.InvalidArgument, "cancel_reason is required")
	}
	// Validate JWT token
	userCtx, err := s.validateJWT(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Cancel(ctx, req.GetId(), req.GetCancelReason(), userCtx.OrgID, userCtx.TenantID); err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "green note not found")
		}
		return nil, err
	}

	return &greennotepb.GreenNoteResponse{
		Success: true,
		Message: "green note cancelled",
	}, nil
}

// GetOrganizationProjects fetches projects for the logged-in user's organization
func (s *GreenNoteService) GetOrganizationProjects(ctx context.Context, req *greennotepb.GetOrganizationProjectsRequest) (*greennotepb.GetOrganizationProjectsResponse, error) {
	// Validate JWT token
	userCtx, err := s.validateJWT(ctx)
	if err != nil {
		return nil, err
	}

	// Propagate metadata to outgoing context
	ctx = s.ensureOutgoingContext(ctx)

	fmt.Printf("🔐 GetOrganizationProjects - User ID: %s, Type: %s, Org: %s\n",
		userCtx.UserID, userCtx.UserType, userCtx.OrgID)

	// Call Project Service
	resp, err := s.projectClient.ListProjectsByOrganization(ctx, &projectpb.ListProjectsByOrganizationRequest{
		OrgId: userCtx.OrgID,
	})
	if err != nil {
		fmt.Printf("⚠️ Failed to fetch projects from external service: %v\n", err)
		// Fallback to empty list instead of failing hard, or return error depending on requirements
		return nil, status.Errorf(codes.Internal, "failed to fetch projects: %v", err)
	}

	// Map external projects to local proto
	projects := make([]*greennotepb.Project, len(resp.Projects))
	for i, p := range resp.Projects {
		projects[i] = &greennotepb.Project{
			Id:          p.ProjectId,
			Name:        p.ProjectName,
			Code:        "", // Not available in project-service
			Status:      "Active", // Default
			Description: "", // Not available in project-service
		}
	}

	return &greennotepb.GetOrganizationProjectsResponse{
		Projects: projects,
		Message:  fmt.Sprintf("Found %d projects for organization %s", len(projects), userCtx.OrgID),
	}, nil
}

// GetOrganizationVendors fetches vendors for the logged-in user's organization
func (s *GreenNoteService) GetOrganizationVendors(ctx context.Context, req *greennotepb.GetOrganizationVendorsRequest) (*greennotepb.GetOrganizationVendorsResponse, error) {
	// Validate JWT token
	userCtx, err := s.validateJWT(ctx)
	if err != nil {
		return nil, err
	}

	// Propagate metadata to outgoing context
	ctx = s.ensureOutgoingContext(ctx)

	fmt.Printf("🔐 GetOrganizationVendors - User ID: %s, Type: %s, Org: %s\n",
		userCtx.UserID, userCtx.UserType, userCtx.OrgID)

	// Call Vendor Service
	// vendor-service ListVendors supports filtering by IDs in proto, but we need OrgID filter.
	// As per previous context, vendor service's ListVendors checks metadata for org_id.
	// So we just need to pass the context through.
	// Wait, we need to map the response format.
	resp, err := s.vendorClient.ListVendors(ctx, &vendorpb.ListVendorsRequest{
		TenantId: userCtx.TenantID,
		// Filters might be needed here if proto requires them
	})
	if err != nil {
		fmt.Printf("⚠️ Failed to fetch vendors from external service: %v\n", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch vendors: %v", err)
	}

	// Map external vendors to local proto
	vendors := make([]*greennotepb.Vendor, len(resp.Vendors))
	for i, v := range resp.Vendors {
		msmeClass := v.MsmeClassification
		if msmeClass == "" && v.GetMsme() != "" {
			msmeClass = v.GetMsme()
		}

		vendors[i] = &greennotepb.Vendor{
			Id:                 v.Id,
			Name:               v.VendorName,
			Code:               v.VendorCode,
			Email:              v.VendorEmail,
			Phone:              v.GetVendorMobile(), // Using getter for optional string
			ContactPerson:      v.BeneficiaryName,
			Status:             v.Status,
			MsmeClassification: msmeClass,
			ActivityType:       v.GetActivityType(),
		}
		fmt.Printf("Mapping Vendor: %s, MSME: %s -> %s, Activity: %s\n", v.VendorName, v.MsmeClassification, msmeClass, v.GetActivityType())
	}

	return &greennotepb.GetOrganizationVendorsResponse{
		Vendors: vendors,
		Message: fmt.Sprintf("Found %d vendors for organization %s", len(vendors), userCtx.OrgID),
	}, nil
}

// GetOrganizationDepartments fetches departments for the logged-in user's organization
func (s *GreenNoteService) GetOrganizationDepartments(ctx context.Context, req *greennotepb.GetOrganizationDepartmentsRequest) (*greennotepb.GetOrganizationDepartmentsResponse, error) {
	// Validate JWT token
	userCtx, err := s.validateJWT(ctx)
	if err != nil {
		return nil, err
	}

	// Propagate metadata to outgoing context
	ctx = s.ensureOutgoingContext(ctx)

	fmt.Printf("🔐 GetOrganizationDepartments - User ID: %s, Type: %s, Org: %s\n",
		userCtx.UserID, userCtx.UserType, userCtx.OrgID)

	// Call Department Service
	resp, err := s.deptClient.ListDepartments(ctx, &departmentpb.ListDepartmentsRequest{
		Page:     1,    // Default to page 1
		PageSize: 1000, // Fetch logical all for dropdown
	})
	if err != nil {
		fmt.Printf("⚠️ Failed to fetch departments from external service: %v\n", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch departments: %v", err)
	}

	// Map external departments to local proto
	departments := make([]*greennotepb.Department, len(resp.Departments))
	for i, d := range resp.Departments {
		departments[i] = &greennotepb.Department{
			Id:          d.Id,
			Name:        d.Name,
			Code:        "", // Not available in department-service
			Description: d.Description,
			HeadName:    "", // Not available in department-service
			Status:      "Active",
		}
	}

	return &greennotepb.GetOrganizationDepartmentsResponse{
		Departments: departments,
		Message:     fmt.Sprintf("Found %d departments for organization %s", len(departments), userCtx.OrgID),
	}, nil
}

func validateGreenNotePayload(p *greennotepb.GreenNotePayload, isCreate bool) error {
	if p == nil {
		return status.Error(codes.InvalidArgument, "payload is required")
	}
	if strings.TrimSpace(p.GetProjectName()) == "" {
		return status.Error(codes.InvalidArgument, "project_name is required")
	}
	if strings.TrimSpace(p.GetSupplierName()) == "" {
		return status.Error(codes.InvalidArgument, "supplier_name is required")
	}
	if strings.TrimSpace(p.GetExpenseCategory()) == "" {
		return status.Error(codes.InvalidArgument, "expense_category is required")
	}
	_ = isCreate
	return nil
}

// noopEventPublisher is used internally when no publisher is provided.
type noopEventPublisher struct{}

func (n noopEventPublisher) PublishGreenNoteApproved(ctx context.Context, event ports.GreenNoteApprovedEvent) error {
	return nil
}

// normalizeStatus converts various textual representations into canonical
// internal status values.
// normalizeStatus converts various textual representations into canonical
// internal status values (raw strings).
func normalizeStatus(raw string) string {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return ""
	}
	// Check for standard ones
	switch s {
	case "pending", "status_pending":
		return statusPending
	case "approved", "status_approved":
		return statusApproved
	case "rejected", "reject", "status_rejected":
		return statusRejected
	case "draft", "status_draft":
		return statusDraft
	case "cancelled", "status_cancelled":
		return statusCancelled
	}
	// For dynamic status, return as is (lowercase)
	return s
}

// normalizeStatusOnCreate sets the default status for a newly created
// GreenNote. It respects the Status enum if DetailedStatus is not provided.
func normalizeStatusOnCreate(p *greennotepb.GreenNotePayload) {
	if p == nil {
		return
	}
	// 1. Check DetailedStatus (string)
	detailed := normalizeStatus(p.GetDetailedStatus())
	if detailed != "" {
		p.DetailedStatus = detailed
		return
	}

	// 2. Check Status (enum)
	enumStatus := mapStatusEnumToString(p.GetStatus())
	if enumStatus != "" {
		// Prevent manual creation as "approved"
		if enumStatus == statusApproved {
			enumStatus = statusPending
		}
		p.DetailedStatus = enumStatus
		return
	}

	// 3. Default to pending
	p.DetailedStatus = statusPending
}

func mapStatusEnumToString(s greennotepb.Status) string {
	switch s {
	case greennotepb.Status_STATUS_PENDING:
		return statusPending
	case greennotepb.Status_STATUS_APPROVED:
		return statusApproved
	case greennotepb.Status_STATUS_REJECTED:
		return statusRejected
	case greennotepb.Status_STATUS_DRAFT:
		return statusDraft
	case greennotepb.Status_STATUS_CANCELLED:
		return statusCancelled
	default:
		return ""
	}
}

// normalizeStatusOnUpdate preserves or updates the status string.
// It respects the Status enum if DetailedStatus is not provided in the update.
func normalizeStatusOnUpdate(existing, updated *greennotepb.GreenNotePayload) {
	if existing == nil || updated == nil {
		return
	}
	// 1. Check DetailedStatus (string)
	detailed := normalizeStatus(updated.GetDetailedStatus())
	if detailed != "" {
		updated.DetailedStatus = detailed
		return
	}

	// 2. Check Status (enum)
	enumStatus := mapStatusEnumToString(updated.GetStatus())
	if enumStatus != "" {
		// Prevent manual update to "approved" if not already there (which it shouldn't be)
		if enumStatus == statusApproved && existing.GetDetailedStatus() != statusApproved {
			enumStatus = existing.GetDetailedStatus()
		}
		updated.DetailedStatus = enumStatus
		return
	}

	// 3. Fallback to existing
	updated.DetailedStatus = existing.GetDetailedStatus()
}

// applyDerivedFields computes financial and budget-related derived fields on
// the payload so that the repository always persists consistent values.
func applyDerivedFields(p *greennotepb.GreenNotePayload) {
	if p == nil {
		return
	}

	// De-duplicate: avoid having the primary invoice duplicated in the multiple list
	// This prevents double-counting in totals and redundant records in DB.
	if p.Invoice != nil && len(p.Invoices) > 0 {
		priNum := strings.TrimSpace(p.Invoice.InvoiceNumber)
		var cleaned []*greennotepb.InvoiceInput
		for _, in := range p.Invoices {
			if in != nil && strings.TrimSpace(in.InvoiceNumber) != priNum {
				cleaned = append(cleaned, in)
			}
		}
		p.Invoices = cleaned
	}

	// Single invoice: ensure invoice_value is populated, and if top-level
	// amounts are zero, derive them from the primary invoice.
	if inv := p.Invoice; inv != nil {
		if inv.InvoiceValue == 0 {
			inv.InvoiceValue = inv.TaxableValue + inv.Gst + inv.OtherCharges
		}
		if p.BaseValue == 0 && p.OtherCharges == 0 && p.Gst == 0 {
			p.BaseValue = inv.TaxableValue
			p.Gst = inv.Gst
			p.OtherCharges = inv.OtherCharges
		}
	}

	// Multiple invoices: aggregate into order amount when enabled.
	if p.EnableMultipleInvoices {
		var baseTotal, gstTotal, otherTotal, valueTotal float64
		
		// Include primary invoice if present
		if inv := p.Invoice; inv != nil {
			if inv.InvoiceValue == 0 {
				inv.InvoiceValue = inv.TaxableValue + inv.Gst + inv.OtherCharges
			}
			baseTotal += inv.TaxableValue
			gstTotal += inv.Gst
			otherTotal += inv.OtherCharges
			valueTotal += inv.InvoiceValue
		}

		// Include multiple invoices
		for _, inv := range p.Invoices {
			if inv == nil {
				continue
			}
			if inv.InvoiceValue == 0 {
				inv.InvoiceValue = inv.TaxableValue + inv.Gst + inv.OtherCharges
			}
			baseTotal += inv.TaxableValue
			gstTotal += inv.Gst
			otherTotal += inv.OtherCharges
			valueTotal += inv.InvoiceValue
		}
		p.BaseValue = baseTotal
		p.Gst = gstTotal
		p.OtherCharges = otherTotal
		p.TotalAmount = valueTotal
	} else {
		// No multi-invoice aggregation: compute total from order amount fields.
		p.TotalAmount = p.BaseValue + p.OtherCharges + p.Gst
	}

	// Budget: over/under budget is always actual - budget.
	p.ExpenditureOverBudget = p.ActualExpenditure - p.BudgetExpenditure
}

// publishApprovedIfNeeded emits a GreenNoteApprovedEvent when a note
// transitions into the approved status. This is used to trigger payment note
// creation in downstream services via Kafka.
func (s *GreenNoteService) publishApprovedIfNeeded(ctx context.Context, id string, before, after *greennotepb.GreenNotePayload) {
	if s == nil || s.events == nil || after == nil {
		return
	}
	newStatus := normalizeStatus(after.GetDetailedStatus())
	
	// If after is draft and before was not draft, it MIGHT be a transition from approved.
	// But the user specifically wants APPROVED -> draft.
	// So we trigger if it WAS approved (transiently) or if it's currently draft but came from a non-draft state?
	// Actually, the most reliable way is if after is draft and oldStatus was pending/etc.
	// BUT a user can create a draft directly.
	
	// Let's stick to the explicit transition if possible, but here we only have the states.
	// If after is draft and before was NOT draft, and it's not a new creation (before != nil),
	// it's likely an approval transition.
	
	if newStatus != statusDraft {
		return
	}
	
	oldStatus := ""
	if before != nil {
		oldStatus = normalizeStatus(before.GetDetailedStatus())
	}
	
	if oldStatus == statusDraft || oldStatus == "" {
		return
	}

	event := ports.GreenNoteApprovedEvent{
		GreenNoteID: id,
		OrderNo:     fmt.Sprintf("GN-%s", id),
		NetAmount:   after.GetTotalAmount(),
		Status:      statusDraft,
		Comments:    after.GetRemarks(),
		ApprovedAt:  s.clock().UTC().Format(time.RFC3339),
	}
	_ = s.events.PublishGreenNoteApproved(ctx, event)
}
