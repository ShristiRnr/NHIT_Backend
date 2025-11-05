package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
)

type userRepository struct {
	queries *db.Queries
}

// NewUserRepository creates a new user repository
func NewUserRepository(queries *db.Queries) ports.UserRepository {
	return &userRepository{queries: queries}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	params := db.CreateUserParams{
		TenantID: user.TenantID,
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
	}

	dbUser, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainUser(&dbUser), nil
}

func (r *userRepository) GetByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	dbUser, err := r.queries.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return toDomainUser(&dbUser), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.User, error) {
	params := db.GetUserByEmailAndTenantParams{
		TenantID: tenantID,
		Email:    email,
	}

	dbUser, err := r.queries.GetUserByEmailAndTenant(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainUser(&dbUser), nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	params := db.UpdateUserParams{
		UserID:   user.UserID,
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
	}

	dbUser, err := r.queries.UpdateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainUser(&dbUser), nil
}

func (r *userRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	return r.queries.DeleteUser(ctx, userID)
}

func (r *userRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]*domain.User, error) {
	params := db.ListUsersByTenantParams{
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
		users[i] = toDomainUser(&dbUser)
	}

	return users, nil
}

// Helper functions to convert between db and domain models
func toDomainUser(dbUser *db.User) *domain.User {
	return &domain.User{
		UserID:          dbUser.UserID,
		TenantID:        dbUser.TenantID,
		Name:            dbUser.Name,
		Email:           dbUser.Email,
		Password:        dbUser.Password,
		EmailVerifiedAt: fromNullTime(dbUser.EmailVerifiedAt),
		LastLoginAt:     fromNullTime(dbUser.LastLoginAt),
		LastLogoutAt:    fromNullTime(dbUser.LastLogoutAt),
		LastLoginIP:     dbUser.LastLoginIp.String,
		UserAgent:       dbUser.UserAgent.String,
		CreatedAt:       dbUser.CreatedAt.Time,
		UpdatedAt:       dbUser.UpdatedAt.Time,
	}
}

func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func fromNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}
