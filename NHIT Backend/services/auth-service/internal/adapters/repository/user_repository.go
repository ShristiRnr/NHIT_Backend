package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/ports"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*ports.UserData, error) {
	query := `
		SELECT user_id, tenant_id, email, name, password, email_verified_at
		FROM users
		WHERE tenant_id = $1 AND email = $2
	`

	user := &ports.UserData{}
	err := r.db.QueryRowContext(ctx, query, tenantID, email).Scan(
		&user.UserID,
		&user.TenantID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.EmailVerifiedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	query := `
		UPDATE users
		SET password = $1, updated_at = $2
		WHERE user_id = $3
	`

	result, err := r.db.ExecContext(ctx, query, hashedPassword, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) error {
	query := `
		UPDATE users
		SET last_login_at = $1, last_login_ip = $2, updated_at = $3
		WHERE user_id = $4
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), ipAddress, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

func (r *userRepository) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET email_verified_at = $1, updated_at = $2
		WHERE user_id = $3
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
