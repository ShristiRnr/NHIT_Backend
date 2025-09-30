package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// VendorServiceImpl implements VendorService
type VendorServiceImpl struct {
	repo ports.VendorRepository
}

// NewVendorService creates a new VendorService
func NewVendorService(repo ports.VendorRepository) *VendorServiceImpl {
	return &VendorServiceImpl{repo: repo}
}

// List returns all vendors, optionally only active ones
func (s *VendorServiceImpl) List(ctx context.Context, onlyActive bool) ([]db.Vendor, error) {
	const defaultLimit = 100
	const defaultOffset = 0

	vendors, err := s.repo.List(ctx, defaultLimit, defaultOffset)
	if err != nil {
		return nil, err
	}

	if onlyActive {
		activeVendors := make([]db.Vendor, 0, len(vendors))
		for _, v := range vendors {
			if v.Active == "Y" {
				activeVendors = append(activeVendors, v)
			}
		}
		return activeVendors, nil
	}

	return vendors, nil
}

// Get returns a single vendor by ID
func (s *VendorServiceImpl) Get(ctx context.Context, id uuid.UUID) (*db.Vendor, error) {
	vendor, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}

// Create adds a new vendor
func (s *VendorServiceImpl) Create(ctx context.Context, v db.Vendor) (*db.Vendor, error) {
	arg := db.CreateVendorParams{
		SNo:                    v.SNo,
		FromAccountType:        v.FromAccountType,
		Status:                 v.Status,
		Project:                v.Project,
		AccountName:            v.AccountName,
		ShortName:              v.ShortName,
		Parent:                 v.Parent,
		AccountNumber:          v.AccountNumber,
		NameOfBank:             v.NameOfBank,
		IfscCodeID:             v.IfscCodeID,
		IfscCode:               v.IfscCode,
		VendorType:             v.VendorType,
		VendorCode:             v.VendorCode,
		VendorName:             v.VendorName,
		VendorEmail:            v.VendorEmail,
		VendorMobile:           v.VendorMobile,
		ActivityType:           v.ActivityType,
		VendorNickName:         v.VendorNickName,
		Email:                  v.Email,
		Mobile:                 v.Mobile,
		Gstin:                  v.Gstin,
		Pan:                    v.Pan,
		Pin:                    v.Pin,
		CountryID:              v.CountryID,
		StateID:                v.StateID,
		CityID:                 v.CityID,
		CountryName:            v.CountryName,
		StateName:              v.StateName,
		CityName:               v.CityName,
		MsmeClassification:     v.MsmeClassification,
		Msme:                   v.Msme,
		MsmeRegistrationNumber: v.MsmeRegistrationNumber,
		MsmeStartDate:          v.MsmeStartDate,
		MsmeEndDate:            v.MsmeEndDate,
		MaterialNature:         v.MaterialNature,
		GstDefaulted:           v.GstDefaulted,
		Section206abVerified:   v.Section206abVerified,
		BenificiaryName:        v.BenificiaryName,
		RemarksAddress:         v.RemarksAddress,
		CommonBankDetails:      v.CommonBankDetails,
		IncomeTaxType:          v.IncomeTaxType,
		FilePath:               v.FilePath,
		Active:                 v.Active,
	}

	vendor, err := s.repo.Create(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}

// Update modifies an existing vendor
func (s *VendorServiceImpl) Update(ctx context.Context, id uuid.UUID, v db.Vendor) (*db.Vendor, error) {
	// ensure vendor exists
	_, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, errors.New("vendor not found")
	}

	arg := db.UpdateVendorParams{
		SNo:                    v.SNo,
		FromAccountType:        v.FromAccountType,
		Status:                 v.Status,
		Project:                v.Project,
		AccountName:            v.AccountName,
		ShortName:              v.ShortName,
		Parent:                 v.Parent,
		AccountNumber:          v.AccountNumber,
		NameOfBank:             v.NameOfBank,
		IfscCodeID:             v.IfscCodeID,
		IfscCode:               v.IfscCode,
		VendorType:             v.VendorType,
		VendorCode:             v.VendorCode,
		VendorName:             v.VendorName,
		VendorEmail:            v.VendorEmail,
		VendorMobile:           v.VendorMobile,
		ActivityType:           v.ActivityType,
		VendorNickName:         v.VendorNickName,
		Email:                  v.Email,
		Mobile:                 v.Mobile,
		Gstin:                  v.Gstin,
		Pan:                    v.Pan,
		Pin:                    v.Pin,
		CountryID:              v.CountryID,
		StateID:                v.StateID,
		CityID:                 v.CityID,
		CountryName:            v.CountryName,
		StateName:              v.StateName,
		CityName:               v.CityName,
		MsmeClassification:     v.MsmeClassification,
		Msme:                   v.Msme,
		MsmeRegistrationNumber: v.MsmeRegistrationNumber,
		MsmeStartDate:          v.MsmeStartDate,
		MsmeEndDate:            v.MsmeEndDate,
		MaterialNature:         v.MaterialNature,
		GstDefaulted:           v.GstDefaulted,
		Section206abVerified:   v.Section206abVerified,
		BenificiaryName:        v.BenificiaryName,
		RemarksAddress:         v.RemarksAddress,
		CommonBankDetails:      v.CommonBankDetails,
		IncomeTaxType:          v.IncomeTaxType,
		FilePath:               v.FilePath,
		Active:                 v.Active,
		ID:                     id,
	}

	vendor, err := s.repo.Update(ctx, arg)
	if err != nil {
		return nil, err
	}

	return &vendor, nil
}

// Delete removes a vendor by ID
func (s *VendorServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	// optional: check if vendor exists first
	_, err := s.repo.Get(ctx, id)
	if err != nil {
		return errors.New("vendor not found")
	}

	return s.repo.Delete(ctx, id)
}

// Search finds vendors by query
func (s *VendorServiceImpl) Search(ctx context.Context, query string, limit, offset int32) ([]db.Vendor, error) {
	vendors, err := s.repo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return vendors, nil
}