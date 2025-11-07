package services

import (
	"context"
	"fmt"

	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/ports"
	"github.com/google/uuid"
)

// CreateVendorAccount creates a new vendor account (equivalent to PHP VendorService::createVendorAccount)
func (s *vendorService) CreateVendorAccount(ctx context.Context, params domain.CreateVendorAccountParams) (*domain.VendorAccount, error) {
	s.logger.Info(ctx, "Creating vendor account", map[string]interface{}{
		"vendor_id":      params.VendorID.String(),
		"account_name":   params.AccountName,
		"account_number": params.AccountNumber,
		"is_primary":     params.IsPrimary,
	})

	var account *domain.VendorAccount

	err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Verify vendor exists
		_, err := s.repo.GetVendorByID(txCtx, uuid.Nil, params.VendorID) // We'll get tenant from vendor
		if err != nil {
			return fmt.Errorf("vendor not found: %w", err)
		}

		// Create account domain entity
		account, err = domain.NewVendorAccount(params)
		if err != nil {
			return fmt.Errorf("failed to create account entity: %w", err)
		}

		// If setting as primary, unset other primary accounts (equivalent to PHP logic)
		if params.IsPrimary {
			err = s.repo.UnsetPrimaryVendorAccounts(txCtx, params.VendorID, nil)
			if err != nil {
				return fmt.Errorf("failed to unset primary accounts: %w", err)
			}
		}

		// Save account to repository
		err = s.repo.CreateVendorAccount(txCtx, account)
		if err != nil {
			return fmt.Errorf("failed to save account: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "Failed to create vendor account", err, map[string]interface{}{
			"vendor_id":    params.VendorID.String(),
			"account_name": params.AccountName,
		})
		return nil, err
	}

	// Publish domain event
	if s.publisher != nil {
		if err := s.publisher.PublishAccountCreated(ctx, account); err != nil {
			s.logger.Warn(ctx, "Failed to publish account created event", map[string]interface{}{
				"account_id": account.ID.String(),
				"error":      err.Error(),
			})
		}
	}

	s.logger.Info(ctx, "Vendor account created successfully", map[string]interface{}{
		"account_id":   account.ID.String(),
		"vendor_id":    account.VendorID.String(),
		"is_primary":   account.IsPrimary,
	})

	return account, nil
}

// GetVendorAccounts gets all accounts for a vendor (equivalent to PHP VendorService::getVendorAccounts)
func (s *vendorService) GetVendorAccounts(ctx context.Context, vendorID uuid.UUID) ([]*domain.VendorAccount, error) {
	accounts, err := s.repo.GetVendorAccountsByVendorID(ctx, vendorID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get vendor accounts", err, map[string]interface{}{
			"vendor_id": vendorID.String(),
		})
		return nil, err
	}

	return accounts, nil
}

// GetVendorBankingDetails gets banking details for payment processing (equivalent to PHP VendorService::getVendorBankingDetails)
func (s *vendorService) GetVendorBankingDetails(ctx context.Context, vendorID uuid.UUID, accountID *uuid.UUID) (*ports.BankingDetails, error) {
	var account *domain.VendorAccount
	var err error

	if accountID != nil {
		// Get specific account
		account, err = s.repo.GetVendorAccountByID(ctx, *accountID)
		if err != nil {
			return nil, fmt.Errorf("failed to get account: %w", err)
		}

		// Verify account belongs to vendor
		if account.VendorID != vendorID {
			return nil, domain.ErrVendorAccountNotFound
		}
	} else {
		// Get primary account
		account, err = s.repo.GetPrimaryVendorAccount(ctx, vendorID)
		if err != nil {
			// Fallback to vendor's direct banking details (backward compatibility)
			return s.getVendorDirectBankingDetails(ctx, vendorID)
		}
	}

	if account == nil {
		return s.getVendorDirectBankingDetails(ctx, vendorID)
	}

	return &ports.BankingDetails{
		AccountName:   account.AccountName,
		AccountNumber: account.AccountNumber,
		AccountType:   account.AccountType,
		NameOfBank:    account.NameOfBank,
		BranchName:    account.BranchName,
		IFSCCode:      account.IFSCCode,
		SwiftCode:     account.SwiftCode,
		Remarks:       account.Remarks,
	}, nil
}

// getVendorDirectBankingDetails gets banking details from vendor record (backward compatibility)
func (s *vendorService) getVendorDirectBankingDetails(ctx context.Context, vendorID uuid.UUID) (*ports.BankingDetails, error) {
	vendor, err := s.repo.GetVendorByID(ctx, uuid.Nil, vendorID) // We'll handle tenant lookup
	if err != nil {
		return nil, fmt.Errorf("failed to get vendor: %w", err)
	}

	if vendor.AccountNumber == nil || vendor.NameOfBank == nil || vendor.IFSCCode == nil {
		return nil, domain.ErrNoPrimaryAccount
	}

	return &ports.BankingDetails{
		AccountName:   vendor.VendorName,
		AccountNumber: *vendor.AccountNumber,
		NameOfBank:    *vendor.NameOfBank,
		IFSCCode:      *vendor.IFSCCode,
	}, nil
}

// UpdateVendorAccount updates a vendor account (equivalent to PHP VendorService::updateVendorAccount)
func (s *vendorService) UpdateVendorAccount(ctx context.Context, accountID uuid.UUID, params domain.UpdateVendorAccountParams) (*domain.VendorAccount, error) {
	s.logger.Info(ctx, "Updating vendor account", map[string]interface{}{
		"account_id": accountID.String(),
	})

	var account *domain.VendorAccount

	err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get existing account
		var err error
		account, err = s.repo.GetVendorAccountByID(txCtx, accountID)
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}

		// If setting as primary and not already primary, unset other primary accounts
		if params.IsPrimary != nil && *params.IsPrimary && !account.IsPrimary {
			err = s.repo.UnsetPrimaryVendorAccounts(txCtx, account.VendorID, &accountID)
			if err != nil {
				return fmt.Errorf("failed to unset primary accounts: %w", err)
			}
		}

		// Update account entity
		err = account.Update(params)
		if err != nil {
			return fmt.Errorf("failed to update account entity: %w", err)
		}

		// Save updated account
		err = s.repo.UpdateVendorAccount(txCtx, account)
		if err != nil {
			return fmt.Errorf("failed to save updated account: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "Failed to update vendor account", err, map[string]interface{}{
			"account_id": accountID.String(),
		})
		return nil, err
	}

	// Publish domain event
	if s.publisher != nil {
		if err := s.publisher.PublishAccountUpdated(ctx, account); err != nil {
			s.logger.Warn(ctx, "Failed to publish account updated event", map[string]interface{}{
				"account_id": account.ID.String(),
				"error":      err.Error(),
			})
		}
	}

	s.logger.Info(ctx, "Vendor account updated successfully", map[string]interface{}{
		"account_id": account.ID.String(),
	})

	return account, nil
}

// DeleteVendorAccount deletes a vendor account (equivalent to PHP VendorService::deleteVendorAccount)
func (s *vendorService) DeleteVendorAccount(ctx context.Context, accountID uuid.UUID) error {
	s.logger.Info(ctx, "Deleting vendor account", map[string]interface{}{
		"account_id": accountID.String(),
	})

	var vendorID uuid.UUID
	var wasPrimary bool

	err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get account to check if it's primary
		account, err := s.repo.GetVendorAccountByID(txCtx, accountID)
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}

		vendorID = account.VendorID
		wasPrimary = account.IsPrimary

		// Delete account
		err = s.repo.DeleteVendorAccount(txCtx, accountID)
		if err != nil {
			return fmt.Errorf("failed to delete account: %w", err)
		}

		// If deleted account was primary, set another account as primary (equivalent to PHP logic)
		if wasPrimary {
			accounts, err := s.repo.GetVendorAccountsByVendorID(txCtx, vendorID)
			if err != nil {
				return fmt.Errorf("failed to get remaining accounts: %w", err)
			}

			// Find first active account to set as primary
			for _, acc := range accounts {
				if acc.IsActive {
					acc.SetPrimary()
					err = s.repo.UpdateVendorAccount(txCtx, acc)
					if err != nil {
						return fmt.Errorf("failed to set new primary account: %w", err)
					}
					break
				}
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "Failed to delete vendor account", err, map[string]interface{}{
			"account_id": accountID.String(),
		})
		return err
	}

	// Publish domain event
	if s.publisher != nil {
		if err := s.publisher.PublishAccountDeleted(ctx, accountID); err != nil {
			s.logger.Warn(ctx, "Failed to publish account deleted event", map[string]interface{}{
				"account_id": accountID.String(),
				"error":      err.Error(),
			})
		}
	}

	s.logger.Info(ctx, "Vendor account deleted successfully", map[string]interface{}{
		"account_id": accountID.String(),
		"was_primary": wasPrimary,
	})

	return nil
}

// ToggleAccountStatus toggles account status (equivalent to PHP VendorService::toggleAccountStatus)
func (s *vendorService) ToggleAccountStatus(ctx context.Context, accountID uuid.UUID) (*domain.VendorAccount, error) {
	s.logger.Info(ctx, "Toggling account status", map[string]interface{}{
		"account_id": accountID.String(),
	})

	var account *domain.VendorAccount

	err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get existing account
		var err error
		account, err = s.repo.GetVendorAccountByID(txCtx, accountID)
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}

		wasPrimary := account.IsPrimary
		
		// Toggle status
		account.ToggleStatus()

		// If deactivating a primary account, set another account as primary (equivalent to PHP logic)
		if !account.IsActive && wasPrimary {
			accounts, err := s.repo.GetVendorAccountsByVendorID(txCtx, account.VendorID)
			if err != nil {
				return fmt.Errorf("failed to get vendor accounts: %w", err)
			}

			// Find first active account to set as primary
			for _, acc := range accounts {
				if acc.ID != accountID && acc.IsActive {
					acc.SetPrimary()
					err = s.repo.UpdateVendorAccount(txCtx, acc)
					if err != nil {
						return fmt.Errorf("failed to set new primary account: %w", err)
					}
					break
				}
			}

			// Unset primary status from current account
			account.UnsetPrimary()
		}

		// Save updated account
		err = s.repo.UpdateVendorAccount(txCtx, account)
		if err != nil {
			return fmt.Errorf("failed to save updated account: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "Failed to toggle account status", err, map[string]interface{}{
			"account_id": accountID.String(),
		})
		return nil, err
	}

	// Publish domain event
	if s.publisher != nil {
		if err := s.publisher.PublishAccountUpdated(ctx, account); err != nil {
			s.logger.Warn(ctx, "Failed to publish account updated event", map[string]interface{}{
				"account_id": account.ID.String(),
				"error":      err.Error(),
			})
		}
	}

	s.logger.Info(ctx, "Account status toggled successfully", map[string]interface{}{
		"account_id": account.ID.String(),
		"new_status": account.IsActive,
	})

	return account, nil
}

// SetPrimaryAccount sets an account as primary (equivalent to PHP VendorService::updateVendorAccount with is_primary=true)
func (s *vendorService) SetPrimaryAccount(ctx context.Context, accountID uuid.UUID) (*domain.VendorAccount, error) {
	s.logger.Info(ctx, "Setting account as primary", map[string]interface{}{
		"account_id": accountID.String(),
	})

	var account *domain.VendorAccount

	err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get existing account
		var err error
		account, err = s.repo.GetVendorAccountByID(txCtx, accountID)
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}

		// Only proceed if not already primary
		if !account.IsPrimary {
			// Unset other primary accounts
			err = s.repo.UnsetPrimaryVendorAccounts(txCtx, account.VendorID, &accountID)
			if err != nil {
				return fmt.Errorf("failed to unset primary accounts: %w", err)
			}

			// Set as primary
			account.SetPrimary()

			// Save updated account
			err = s.repo.UpdateVendorAccount(txCtx, account)
			if err != nil {
				return fmt.Errorf("failed to save updated account: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "Failed to set account as primary", err, map[string]interface{}{
			"account_id": accountID.String(),
		})
		return nil, err
	}

	// Publish domain event
	if s.publisher != nil {
		if err := s.publisher.PublishAccountUpdated(ctx, account); err != nil {
			s.logger.Warn(ctx, "Failed to publish account updated event", map[string]interface{}{
				"account_id": account.ID.String(),
				"error":      err.Error(),
			})
		}
	}

	s.logger.Info(ctx, "Account set as primary successfully", map[string]interface{}{
		"account_id": account.ID.String(),
	})

	return account, nil
}
