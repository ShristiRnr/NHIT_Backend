package handlers

import (
	"context"
	"regexp"
	"strings"

	vendorpb "github.com/ShristiRnr/NHIT_Backend/api/pb/vendorpb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// VendorAccountController handles vendor account operations - equivalent to PHP VendorAccountController
type VendorAccountController struct {
	vendorpb.UnimplementedVendorServiceServer
}

// NewVendorAccountController creates a new vendor account controller
func NewVendorAccountController() *VendorAccountController {
	return &VendorAccountController{}
}

// CreateVendorAccount creates a new vendor account (equivalent to PHP store method)
func (c *VendorAccountController) CreateVendorAccount(ctx context.Context, req *vendorpb.CreateVendorAccountRequest) (*vendorpb.VendorAccountResponse, error) {
	// Validate request (equivalent to PHP validation rules)
	if err := c.validateCreateAccountRequest(req); err != nil {
		return nil, err
	}

	// Check if vendor exists
	_, exists := vendors[req.VendorId]
	if !exists {
		return nil, status.Error(codes.NotFound, "vendor not found")
	}

	accountID := uuid.New().String()

	// Handle primary account logic (equivalent to PHP VendorService logic)
	if req.IsPrimary {
		if err := c.unsetOtherPrimaryAccounts(req.VendorId); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update primary accounts: %v", err)
		}
	}

	// Create account (equivalent to PHP model creation)
	account := &vendorpb.VendorAccount{
		Id:            accountID,
		VendorId:      req.VendorId,
		AccountName:   req.AccountName,
		AccountNumber: req.AccountNumber,
		AccountType:   req.AccountType,
		NameOfBank:    req.NameOfBank,
		BranchName:    req.BranchName,
		IfscCode:      req.IfscCode,
		SwiftCode:     req.SwiftCode,
		IsPrimary:     req.IsPrimary,
		IsActive:      true,
		Remarks:       req.Remarks,
		CreatedBy:     req.CreatedBy,
		CreatedAt:     timestamppb.Now(),
		UpdatedAt:     timestamppb.Now(),
	}

	// Store account
	accounts[accountID] = account
	vendorAccounts[req.VendorId] = append(vendorAccounts[req.VendorId], account)

	return &vendorpb.VendorAccountResponse{Account: account}, nil
}

// GetVendorAccounts gets all accounts for a vendor (equivalent to PHP index method)
func (c *VendorAccountController) GetVendorAccounts(ctx context.Context, req *vendorpb.GetVendorAccountsRequest) (*vendorpb.GetVendorAccountsResponse, error) {
	if req.VendorId == "" {
		return nil, status.Error(codes.InvalidArgument, "vendor_id is required")
	}

	// Check if vendor exists
	_, exists := vendors[req.VendorId]
	if !exists {
		return nil, status.Error(codes.NotFound, "vendor not found")
	}

	vendorAccountsList, exists := vendorAccounts[req.VendorId]
	if !exists {
		vendorAccountsList = []*vendorpb.VendorAccount{}
	}

	return &vendorpb.GetVendorAccountsResponse{
		Accounts: vendorAccountsList,
	}, nil
}

// GetVendorAccount gets a specific vendor account (equivalent to PHP show method)
func (c *VendorAccountController) GetVendorAccount(ctx context.Context, req *vendorpb.GetVendorAccountRequest) (*vendorpb.VendorAccountResponse, error) {
	if req.AccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	account, exists := accounts[req.AccountId]
	if !exists {
		return nil, status.Error(codes.NotFound, "account not found")
	}

	// Verify account belongs to the specified vendor
	if req.VendorId != nil && *req.VendorId != "" && account.VendorId != *req.VendorId {
		return nil, status.Error(codes.NotFound, "account not found for this vendor")
	}

	return &vendorpb.VendorAccountResponse{Account: account}, nil
}

// UpdateVendorAccount updates a vendor account (equivalent to PHP update method)
func (c *VendorAccountController) UpdateVendorAccount(ctx context.Context, req *vendorpb.UpdateVendorAccountRequest) (*vendorpb.VendorAccountResponse, error) {
	if req.AccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	account, exists := accounts[req.AccountId]
	if !exists {
		return nil, status.Error(codes.NotFound, "account not found")
	}

	// Validate updated fields (equivalent to PHP validation rules)
	if err := c.validateUpdateAccountRequest(req); err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.AccountName != nil {
		account.AccountName = *req.AccountName
	}
	if req.AccountNumber != nil {
		account.AccountNumber = *req.AccountNumber
	}
	if req.AccountType != nil {
		account.AccountType = req.AccountType
	}
	if req.NameOfBank != nil {
		account.NameOfBank = *req.NameOfBank
	}
	if req.BranchName != nil {
		account.BranchName = req.BranchName
	}
	if req.IfscCode != nil {
		if err := c.validateIFSC(*req.IfscCode); err != nil {
			return nil, err
		}
		account.IfscCode = *req.IfscCode
	}
	if req.SwiftCode != nil {
		account.SwiftCode = req.SwiftCode
	}
	if req.Remarks != nil {
		account.Remarks = req.Remarks
	}

	// Handle primary account logic (equivalent to PHP VendorService logic)
	if req.IsPrimary != nil && *req.IsPrimary && !account.IsPrimary {
		if err := c.unsetOtherPrimaryAccounts(account.VendorId); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update primary accounts: %v", err)
		}
		account.IsPrimary = true
	}

	if req.IsActive != nil {
		account.IsActive = *req.IsActive
	}

	account.UpdatedAt = timestamppb.Now()
	accounts[req.AccountId] = account

	return &vendorpb.VendorAccountResponse{Account: account}, nil
}

// DeleteVendorAccount deletes a vendor account (equivalent to PHP destroy method)
func (c *VendorAccountController) DeleteVendorAccount(ctx context.Context, req *vendorpb.DeleteVendorAccountRequest) (*emptypb.Empty, error) {
	if req.AccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	account, exists := accounts[req.AccountId]
	if !exists {
		return nil, status.Error(codes.NotFound, "account not found")
	}

	vendorId := account.VendorId
	wasPrimary := account.IsPrimary

	// Remove from accounts
	delete(accounts, req.AccountId)

	// Remove from vendor accounts list and handle primary reassignment
	if vendorAccountsList, exists := vendorAccounts[vendorId]; exists {
		var updatedList []*vendorpb.VendorAccount
		for _, acc := range vendorAccountsList {
			if acc.Id != req.AccountId {
				updatedList = append(updatedList, acc)
			}
		}
		vendorAccounts[vendorId] = updatedList

		// If deleted account was primary, set another as primary (equivalent to PHP logic)
		if wasPrimary && len(updatedList) > 0 {
			updatedList[0].IsPrimary = true
			accounts[updatedList[0].Id] = updatedList[0]
		}
	}

	return &emptypb.Empty{}, nil
}

// ToggleAccountStatus toggles account active status (equivalent to PHP toggleStatus method)
func (c *VendorAccountController) ToggleAccountStatus(ctx context.Context, req *vendorpb.ToggleAccountStatusRequest) (*vendorpb.VendorAccountResponse, error) {
	if req.AccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	account, exists := accounts[req.AccountId]
	if !exists {
		return nil, status.Error(codes.NotFound, "account not found")
	}

	wasPrimary := account.IsPrimary
	account.IsActive = !account.IsActive

	// If deactivating a primary account, set another as primary (equivalent to PHP logic)
	if !account.IsActive && wasPrimary {
		if err := c.reassignPrimaryAccount(account.VendorId, account.Id); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to reassign primary account: %v", err)
		}
		account.IsPrimary = false
	}

	account.UpdatedAt = timestamppb.Now()
	accounts[req.AccountId] = account

	return &vendorpb.VendorAccountResponse{Account: account}, nil
}

// SetPrimaryAccount sets an account as primary (equivalent to PHP setPrimary method)
func (c *VendorAccountController) SetPrimaryAccount(ctx context.Context, req *vendorpb.SetPrimaryAccountRequest) (*vendorpb.VendorAccountResponse, error) {
	if req.AccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	account, exists := accounts[req.AccountId]
	if !exists {
		return nil, status.Error(codes.NotFound, "account not found")
	}

	// Only proceed if not already primary
	if !account.IsPrimary {
		if err := c.unsetOtherPrimaryAccounts(account.VendorId); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update primary accounts: %v", err)
		}
		account.IsPrimary = true
		account.UpdatedAt = timestamppb.Now()
		accounts[req.AccountId] = account
	}

	return &vendorpb.VendorAccountResponse{Account: account}, nil
}

// GetVendorBankingDetails gets banking details for payment processing (equivalent to PHP getBankingDetails method)
func (c *VendorAccountController) GetVendorBankingDetails(ctx context.Context, req *vendorpb.GetVendorBankingDetailsRequest) (*vendorpb.BankingDetailsResponse, error) {
	if req.VendorId == "" {
		return nil, status.Error(codes.InvalidArgument, "vendor_id is required")
	}

	var account *vendorpb.VendorAccount

	if req.AccountId != nil && *req.AccountId != "" {
		// Get specific account
		if acc, exists := accounts[*req.AccountId]; exists && acc.VendorId == req.VendorId {
			account = acc
		}
	} else {
		// Get primary account
		if vendorAccountsList, exists := vendorAccounts[req.VendorId]; exists {
			for _, acc := range vendorAccountsList {
				if acc.IsPrimary && acc.IsActive {
					account = acc
					break
				}
			}
		}
	}

	if account == nil {
		return nil, status.Error(codes.NotFound, "banking details not found")
	}

	return &vendorpb.BankingDetailsResponse{
		BankingDetails: &vendorpb.BankingDetails{
			AccountName:   account.AccountName,
			AccountNumber: account.AccountNumber,
			AccountType:   account.AccountType,
			NameOfBank:    account.NameOfBank,
			BranchName:    account.BranchName,
			IfscCode:      account.IfscCode,
			SwiftCode:     account.SwiftCode,
			Remarks:       account.Remarks,
		},
	}, nil
}

// Helper methods (equivalent to PHP VendorService methods)

func (c *VendorAccountController) validateCreateAccountRequest(req *vendorpb.CreateVendorAccountRequest) error {
	if req.VendorId == "" {
		return status.Error(codes.InvalidArgument, "vendor_id is required")
	}
	if req.AccountName == "" {
		return status.Error(codes.InvalidArgument, "account_name is required")
	}
	if req.AccountNumber == "" {
		return status.Error(codes.InvalidArgument, "account_number is required")
	}
	if req.NameOfBank == "" {
		return status.Error(codes.InvalidArgument, "name_of_bank is required")
	}
	if req.IfscCode == "" {
		return status.Error(codes.InvalidArgument, "ifsc_code is required")
	}
	if req.CreatedBy == "" {
		return status.Error(codes.InvalidArgument, "created_by is required")
	}

	// Validate account number format
	if err := c.validateAccountNumber(req.AccountNumber); err != nil {
		return err
	}

	// Validate IFSC code
	if err := c.validateIFSC(req.IfscCode); err != nil {
		return err
	}

	// Validate account name length (equivalent to PHP max:255)
	if len(strings.TrimSpace(req.AccountName)) > 255 {
		return status.Error(codes.InvalidArgument, "account_name must not exceed 255 characters")
	}

	return nil
}

func (c *VendorAccountController) validateUpdateAccountRequest(req *vendorpb.UpdateVendorAccountRequest) error {
	if req.AccountName != nil && len(strings.TrimSpace(*req.AccountName)) > 255 {
		return status.Error(codes.InvalidArgument, "account_name must not exceed 255 characters")
	}

	if req.AccountNumber != nil {
		if err := c.validateAccountNumber(*req.AccountNumber); err != nil {
			return err
		}
	}

	if req.NameOfBank != nil && len(strings.TrimSpace(*req.NameOfBank)) > 255 {
		return status.Error(codes.InvalidArgument, "name_of_bank must not exceed 255 characters")
	}

	return nil
}

func (c *VendorAccountController) validateAccountNumber(accountNumber string) error {
	// Basic account number validation (9-18 digits)
	accountRegex := regexp.MustCompile(`^[0-9]{9,18}$`)
	if !accountRegex.MatchString(accountNumber) {
		return status.Error(codes.InvalidArgument, "invalid account number format (must be 9-18 digits)")
	}
	return nil
}

func (c *VendorAccountController) validateIFSC(ifsc string) error {
	// IFSC code validation: 4 letters, 1 zero, 6 alphanumeric
	ifscRegex := regexp.MustCompile(`^[A-Z]{4}0[A-Z0-9]{6}$`)
	if !ifscRegex.MatchString(ifsc) {
		return status.Error(codes.InvalidArgument, "invalid IFSC code format")
	}
	return nil
}

func (c *VendorAccountController) unsetOtherPrimaryAccounts(vendorId string) error {
	if vendorAccountsList, exists := vendorAccounts[vendorId]; exists {
		for _, account := range vendorAccountsList {
			if account.IsPrimary {
				account.IsPrimary = false
				account.UpdatedAt = timestamppb.Now()
				accounts[account.Id] = account
			}
		}
	}
	return nil
}

func (c *VendorAccountController) reassignPrimaryAccount(vendorId, excludeAccountId string) error {
	if vendorAccountsList, exists := vendorAccounts[vendorId]; exists {
		for _, acc := range vendorAccountsList {
			if acc.Id != excludeAccountId && acc.IsActive {
				acc.IsPrimary = true
				acc.UpdatedAt = timestamppb.Now()
				accounts[acc.Id] = acc
				return nil
			}
		}
	}
	return nil
}
