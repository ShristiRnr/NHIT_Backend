package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/ports"
	"github.com/google/uuid"
)

// vendorService implements the VendorService interface (Hexagonal Architecture - Business Service)
type vendorService struct {
	repo      ports.VendorRepository
	logger    ports.Logger
	publisher ports.EventPublisher
	config    ports.VendorServiceConfig
}

// NewVendorService creates a new vendor service instance
func NewVendorService(
	repo ports.VendorRepository,
	logger ports.Logger,
	publisher ports.EventPublisher,
	config ports.VendorServiceConfig,
) ports.VendorService {
	return &vendorService{
		repo:      repo,
		logger:    logger,
		publisher: publisher,
		config:    config,
	}
}

// CreateVendor creates a new vendor with complete business logic (equivalent to PHP VendorService::createVendor)
func (s *vendorService) CreateVendor(ctx context.Context, params domain.CreateVendorParams) (*domain.Vendor, error) {
	s.logger.Info(ctx, "Creating vendor", map[string]interface{}{
		"vendor_name":  params.VendorName,
		"vendor_email": params.VendorEmail,
		"tenant_id":    params.TenantID.String(),
	})

	// Start transaction for atomic operation
	var vendor *domain.Vendor
	var primaryAccount *domain.VendorAccount

	err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Check if vendor email already exists
		exists, err := s.repo.IsVendorEmailExists(txCtx, params.TenantID, params.VendorEmail, nil)
		if err != nil {
			return fmt.Errorf("failed to check email existence: %w", err)
		}
		if exists {
			return domain.ErrVendorEmailExists
		}

		// Create vendor domain entity
		vendor, err = domain.NewVendor(params)
		if err != nil {
			return fmt.Errorf("failed to create vendor entity: %w", err)
		}

		// Ensure vendor code is unique
		err = s.ensureUniqueVendorCode(txCtx, vendor)
		if err != nil {
			return fmt.Errorf("failed to ensure unique vendor code: %w", err)
		}

		// Save vendor to repository
		err = s.repo.CreateVendor(txCtx, vendor)
		if err != nil {
			return fmt.Errorf("failed to save vendor: %w", err)
		}

		// Create primary account if banking details provided (equivalent to PHP logic)
		if params.AccountNumber != nil && params.NameOfBank != nil && params.IFSCCode != nil {
			accountParams := domain.CreateVendorAccountParams{
				VendorID:      vendor.ID,
				AccountName:   vendor.VendorName,
				AccountNumber: *params.AccountNumber,
				NameOfBank:    *params.NameOfBank,
				IFSCCode:      *params.IFSCCode,
				IsPrimary:     true,
				CreatedBy:     params.CreatedBy,
			}

			primaryAccount, err = domain.NewVendorAccount(accountParams)
			if err != nil {
				return fmt.Errorf("failed to create primary account: %w", err)
			}

			err = s.repo.CreateVendorAccount(txCtx, primaryAccount)
			if err != nil {
				return fmt.Errorf("failed to save primary account: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "Failed to create vendor", err, map[string]interface{}{
			"vendor_name":  params.VendorName,
			"vendor_email": params.VendorEmail,
		})
		return nil, err
	}

	// Publish domain event
	if s.publisher != nil {
		if err := s.publisher.PublishVendorCreated(ctx, vendor); err != nil {
			s.logger.Warn(ctx, "Failed to publish vendor created event", map[string]interface{}{
				"vendor_id": vendor.ID.String(),
				"error":     err.Error(),
			})
		}
	}

	s.logger.Info(ctx, "Vendor created successfully", map[string]interface{}{
		"vendor_id":   vendor.ID.String(),
		"vendor_code": vendor.VendorCode,
	})

	return vendor, nil
}

// GetVendorByID retrieves a vendor by ID
func (s *vendorService) GetVendorByID(ctx context.Context, tenantID, vendorID uuid.UUID) (*domain.Vendor, error) {
	vendor, err := s.repo.GetVendorByID(ctx, tenantID, vendorID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get vendor by ID", err, map[string]interface{}{
			"vendor_id": vendorID.String(),
			"tenant_id": tenantID.String(),
		})
		return nil, err
	}

	return vendor, nil
}

// GetVendorByCode retrieves a vendor by code
func (s *vendorService) GetVendorByCode(ctx context.Context, tenantID uuid.UUID, vendorCode string) (*domain.Vendor, error) {
	vendor, err := s.repo.GetVendorByCode(ctx, tenantID, vendorCode)
	if err != nil {
		s.logger.Error(ctx, "Failed to get vendor by code", err, map[string]interface{}{
			"vendor_code": vendorCode,
			"tenant_id":   tenantID.String(),
		})
		return nil, err
	}

	return vendor, nil
}

// UpdateVendor updates an existing vendor (equivalent to PHP VendorService::updateVendor)
func (s *vendorService) UpdateVendor(ctx context.Context, tenantID, vendorID uuid.UUID, params domain.UpdateVendorParams) (*domain.Vendor, error) {
	s.logger.Info(ctx, "Updating vendor", map[string]interface{}{
		"vendor_id": vendorID.String(),
		"tenant_id": tenantID.String(),
	})

	var vendor *domain.Vendor

	err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get existing vendor
		var err error
		vendor, err = s.repo.GetVendorByID(txCtx, tenantID, vendorID)
		if err != nil {
			return fmt.Errorf("failed to get vendor: %w", err)
		}

		// Check email uniqueness if updating email
		if params.VendorEmail != nil {
			exists, err := s.repo.IsVendorEmailExists(txCtx, tenantID, *params.VendorEmail, &vendorID)
			if err != nil {
				return fmt.Errorf("failed to check email existence: %w", err)
			}
			if exists {
				return domain.ErrVendorEmailExists
			}
		}

		// Update vendor entity
		err = vendor.Update(params)
		if err != nil {
			return fmt.Errorf("failed to update vendor entity: %w", err)
		}

		// Save updated vendor
		err = s.repo.UpdateVendor(txCtx, vendor)
		if err != nil {
			return fmt.Errorf("failed to save updated vendor: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "Failed to update vendor", err, map[string]interface{}{
			"vendor_id": vendorID.String(),
		})
		return nil, err
	}

	// Publish domain event
	if s.publisher != nil {
		if err := s.publisher.PublishVendorUpdated(ctx, vendor); err != nil {
			s.logger.Warn(ctx, "Failed to publish vendor updated event", map[string]interface{}{
				"vendor_id": vendor.ID.String(),
				"error":     err.Error(),
			})
		}
	}

	s.logger.Info(ctx, "Vendor updated successfully", map[string]interface{}{
		"vendor_id": vendor.ID.String(),
	})

	return vendor, nil
}

// DeleteVendor deletes a vendor (equivalent to PHP VendorService::deleteVendor)
func (s *vendorService) DeleteVendor(ctx context.Context, tenantID, vendorID uuid.UUID) error {
	s.logger.Info(ctx, "Deleting vendor", map[string]interface{}{
		"vendor_id": vendorID.String(),
		"tenant_id": tenantID.String(),
	})

	err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Verify vendor exists
		_, err := s.repo.GetVendorByID(txCtx, tenantID, vendorID)
		if err != nil {
			return fmt.Errorf("failed to get vendor: %w", err)
		}

		// Delete vendor (cascade will handle accounts)
		err = s.repo.DeleteVendor(txCtx, tenantID, vendorID)
		if err != nil {
			return fmt.Errorf("failed to delete vendor: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "Failed to delete vendor", err, map[string]interface{}{
			"vendor_id": vendorID.String(),
		})
		return err
	}

	// Publish domain event
	if s.publisher != nil {
		if err := s.publisher.PublishVendorDeleted(ctx, tenantID, vendorID); err != nil {
			s.logger.Warn(ctx, "Failed to publish vendor deleted event", map[string]interface{}{
				"vendor_id": vendorID.String(),
				"error":     err.Error(),
			})
		}
	}

	s.logger.Info(ctx, "Vendor deleted successfully", map[string]interface{}{
		"vendor_id": vendorID.String(),
	})

	return nil
}

// ListVendors lists vendors with filtering and pagination
func (s *vendorService) ListVendors(ctx context.Context, tenantID uuid.UUID, filters ports.VendorListFilters) ([]*domain.Vendor, int64, error) {
	vendors, total, err := s.repo.ListVendors(ctx, tenantID, filters)
	if err != nil {
		s.logger.Error(ctx, "Failed to list vendors", err, map[string]interface{}{
			"tenant_id": tenantID.String(),
			"filters":   filters,
		})
		return nil, 0, err
	}

	return vendors, total, nil
}

// GenerateVendorCode generates a vendor code (equivalent to PHP VendorService::generateVendorCode)
func (s *vendorService) GenerateVendorCode(ctx context.Context, vendorName string, vendorType *string) (string, error) {
	// Create temporary vendor for code generation
	accountType := ""
	if vendorType != nil {
		accountType = *vendorType
	}
	tempVendor := &domain.Vendor{
		VendorName:  vendorName,
		AccountType: accountType,
	}

	code := tempVendor.GenerateCode()

	s.logger.Debug(ctx, "Generated vendor code", map[string]interface{}{
		"vendor_type": vendorType,
		"generated_code": code,
	})

	return code, nil
}

// UpdateVendorCode updates vendor code (equivalent to PHP VendorService::updateVendorCode)
func (s *vendorService) UpdateVendorCode(ctx context.Context, tenantID, vendorID uuid.UUID, newCode string) (*domain.Vendor, error) {
	s.logger.Info(ctx, "Updating vendor code", map[string]interface{}{
		"vendor_id": vendorID.String(),
		"new_code":  newCode,
	})

	var vendor *domain.Vendor

	err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get existing vendor
		var err error
		vendor, err = s.repo.GetVendorByID(txCtx, tenantID, vendorID)
		if err != nil {
			return fmt.Errorf("failed to get vendor: %w", err)
		}

		// Check code uniqueness
		exists, err := s.repo.IsVendorCodeExists(txCtx, tenantID, newCode, &vendorID)
		if err != nil {
			return fmt.Errorf("failed to check code existence: %w", err)
		}
		if exists {
			return domain.ErrVendorCodeExists
		}

		// Update vendor code
		err = vendor.UpdateCode(newCode)
		if err != nil {
			return fmt.Errorf("failed to update vendor code: %w", err)
		}

		// Save updated vendor
		err = s.repo.UpdateVendor(txCtx, vendor)
		if err != nil {
			return fmt.Errorf("failed to save updated vendor: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "Failed to update vendor code", err, map[string]interface{}{
			"vendor_id": vendorID.String(),
			"new_code":  newCode,
		})
		return nil, err
	}

	s.logger.Info(ctx, "Vendor code updated successfully", map[string]interface{}{
		"vendor_id": vendor.ID.String(),
		"new_code":  vendor.VendorCode,
	})

	return vendor, nil
}

// RegenerateVendorCode regenerates vendor code (equivalent to PHP VendorService::updateVendorCode with null)
func (s *vendorService) RegenerateVendorCode(ctx context.Context, tenantID, vendorID uuid.UUID) (*domain.Vendor, error) {
	s.logger.Info(ctx, "Regenerating vendor code", map[string]interface{}{
		"vendor_id": vendorID.String(),
	})

	var vendor *domain.Vendor

	err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get existing vendor
		var err error
		vendor, err = s.repo.GetVendorByID(txCtx, tenantID, vendorID)
		if err != nil {
			return fmt.Errorf("failed to get vendor: %w", err)
		}

		// Regenerate code
		vendor.RegenerateCode()

		// Ensure new code is unique
		err = s.ensureUniqueVendorCode(txCtx, vendor)
		if err != nil {
			return fmt.Errorf("failed to ensure unique vendor code: %w", err)
		}

		// Save updated vendor
		err = s.repo.UpdateVendor(txCtx, vendor)
		if err != nil {
			return fmt.Errorf("failed to save updated vendor: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "Failed to regenerate vendor code", err, map[string]interface{}{
			"vendor_id": vendorID.String(),
		})
		return nil, err
	}

	s.logger.Info(ctx, "Vendor code regenerated successfully", map[string]interface{}{
		"vendor_id": vendor.ID.String(),
		"new_code":  vendor.VendorCode,
	})

	return vendor, nil
}

// ensureUniqueVendorCode ensures vendor code is unique by adding suffix if needed
func (s *vendorService) ensureUniqueVendorCode(ctx context.Context, vendor *domain.Vendor) error {
	originalCode := vendor.VendorCode
	attempts := 0
	maxAttempts := 10

	for attempts < maxAttempts {
		exists, err := s.repo.IsVendorCodeExists(ctx, vendor.TenantID, vendor.VendorCode, &vendor.ID)
		if err != nil {
			return err
		}

		if !exists {
			return nil
		}

		// Add suffix to make it unique
		attempts++
		vendor.VendorCode = fmt.Sprintf("%s-%d", originalCode, time.Now().Unix()%1000+int64(attempts))
	}

	return fmt.Errorf("failed to generate unique vendor code after %d attempts", maxAttempts)
}
