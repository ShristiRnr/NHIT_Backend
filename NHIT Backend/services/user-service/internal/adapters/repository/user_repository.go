package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
)

type userRepository struct {
	queries *sqlc.Queries
}

// NewUserRepository creates a new user repository
func NewUserRepository(queries *sqlc.Queries) ports.UserRepository {
	return &userRepository{queries: queries}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	// Helper for NullUUID
	toNullUUID := func(u *uuid.UUID) uuid.NullUUID {
		if u == nil {
			return uuid.NullUUID{}
		}
		return uuid.NullUUID{UUID: *u, Valid: true}
	}

	if user.EmailVerifiedAt != nil {
		params := sqlc.CreateUserWithVerificationParams{
			TenantID:          user.TenantID,
			Name:              user.Name,
			Email:             user.Email,
			Password:          user.Password,
			EmailVerifiedAt:   pgtype.Timestamptz{Time: *user.EmailVerifiedAt, Valid: true},
			DepartmentID:      toNullUUID(user.DepartmentID),
			DesignationID:     toNullUUID(user.DesignationID),
			AccountHolderName: user.AccountHolderName,
			BankName:          user.BankName,
			BankAccountNumber: user.BankAccountNumber,
			IfscCode:          user.IFSCCode,
			SignatureUrl:      user.SignatureURL,
			IsActive:          user.IsActive,
		}

		dbUser, err := r.queries.CreateUserWithVerification(ctx, params)
		if err != nil {
			return nil, err
		}

		return toDomainUser(dbUser), nil
	}

	params := sqlc.CreateUserParams{
		TenantID:          user.TenantID,
		Name:              user.Name,
		Email:             user.Email,
		Password:          user.Password,
		DepartmentID:      toNullUUID(user.DepartmentID),
		DesignationID:     toNullUUID(user.DesignationID),
		AccountHolderName: user.AccountHolderName,
		BankName:          user.BankName,
		BankAccountNumber: user.BankAccountNumber,
		IfscCode:          user.IFSCCode,
		SignatureUrl:      user.SignatureURL,
		IsActive:          user.IsActive,
	}

	dbUser, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainUser(dbUser), nil
}

func (r *userRepository) GetByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	dbUser, err := r.queries.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return toDomainUser(dbUser), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error) {
	params := sqlc.GetUserByEmailAndTenantParams{
		TenantID: tenantID,
		Email:    email,
	}

	dbUser, err := r.queries.GetUserByEmailAndTenant(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainUser(dbUser), nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	// Helper for NullUUID
	toNullUUID := func(u *uuid.UUID) uuid.NullUUID {
		if u == nil {
			return uuid.NullUUID{}
		}
		return uuid.NullUUID{UUID: *u, Valid: true}
	}

	var deactivatedAt pgtype.Timestamptz
	if user.DeactivatedAt != nil {
		deactivatedAt = pgtype.Timestamptz{Time: *user.DeactivatedAt, Valid: true}
	}

	var deactivatedBy pgtype.UUID
	if user.DeactivatedBy != nil {
		deactivatedBy = pgtype.UUID{Bytes: *user.DeactivatedBy, Valid: true}
	}

	params := sqlc.UpdateUserParams{
		UserID:            user.UserID,
		Name:              user.Name,
		Email:             user.Email,
		Password:          user.Password,
		DepartmentID:      toNullUUID(user.DepartmentID),
		DesignationID:     toNullUUID(user.DesignationID),
		AccountHolderName: user.AccountHolderName,
		BankName:          user.BankName,
		BankAccountNumber: user.BankAccountNumber,
		IfscCode:          user.IFSCCode,
		SignatureUrl:      user.SignatureURL,
		IsActive:          user.IsActive,
		DeactivatedAt:     deactivatedAt,
		DeactivatedBy:     deactivatedBy,
	}

	dbUser, err := r.queries.UpdateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainUser(dbUser), nil
}

func (r *userRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	return r.queries.DeleteUser(ctx, userID)
}

func (r *userRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]*domain.User, error) {
	params := sqlc.ListUsersByTenantParams{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	}

	dbUsers, err := r.queries.ListUsersByTenant(ctx, params)
	if err != nil {
		return nil, err
	}

	users := make([]*domain.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = toDomainUser(dbUser)
	}

	return users, nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, lastLoginIP, userAgent string) error {
	var ip *string
	if lastLoginIP != "" {
		ip = &lastLoginIP
	}
	var ua *string
	if userAgent != "" {
		ua = &userAgent
	}
	
	params := sqlc.UpdateUserLastLoginParams{
		UserID:      userID,
		LastLoginIp: ip,
		UserAgent:   ua,
	}

	_, err := r.queries.UpdateUserLastLogin(ctx, params)
	return err
}

// Helper functions to convert between db and domain models
func toDomainUser(dbUser *sqlc.User) *domain.User {
	var deptID *uuid.UUID
	if dbUser.DepartmentID.Valid {
		id := dbUser.DepartmentID.UUID
		deptID = &id
	}

	var desigID *uuid.UUID
	if dbUser.DesignationID.Valid {
		id := dbUser.DesignationID.UUID
		desigID = &id
	}
	
	var deactivatedBy *uuid.UUID
	if dbUser.DeactivatedBy.Valid {
		id := uuid.UUID(dbUser.DeactivatedBy.Bytes)
		deactivatedBy = &id
	}

	return &domain.User{
		UserID:            dbUser.UserID,
		TenantID:          dbUser.TenantID,
		Name:              dbUser.Name,
		Email:             dbUser.Email,
		Password:          dbUser.Password,
		DepartmentID:      deptID,
		DesignationID:     desigID,
		AccountHolderName: dbUser.AccountHolderName,
		BankName:          dbUser.BankName,
		BankAccountNumber: dbUser.BankAccountNumber,
		IFSCCode:          dbUser.IfscCode,
		SignatureURL:      dbUser.SignatureUrl,
		IsActive:          dbUser.IsActive,
		DeactivatedAt:     fromPgTimestamptz(dbUser.DeactivatedAt),
		DeactivatedBy:     deactivatedBy,
		EmailVerifiedAt:   fromPgTimestamptz(dbUser.EmailVerifiedAt),
		LastLoginAt:       fromPgTimestamptz(dbUser.LastLoginAt),
		LastLogoutAt:      fromPgTimestamptz(dbUser.LastLogoutAt),
		LastLoginIP:       fromStringPtr(dbUser.LastLoginIp),
		UserAgent:         fromStringPtr(dbUser.UserAgent),
		CreatedAt:         dbUser.CreatedAt.Time,
		UpdatedAt:         dbUser.UpdatedAt.Time,
	}
}

func fromPgTimestamptz(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}

func fromStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
