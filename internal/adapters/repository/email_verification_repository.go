package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// Ensure interface compliance
var _ ports.EmailVerificationRepository = &EmailVerificationRepo{}

type EmailVerificationRepo struct {
	q *db.Queries
}

func NewEmailVerificationRepo(q *db.Queries) ports.EmailVerificationRepository {
	return &EmailVerificationRepo{q: q}
}

func (r *EmailVerificationRepo) Insert(ctx context.Context, userID uuid.UUID, token uuid.UUID) (db.EmailVerification, error) {
	arg := db.InsertEmailVerificationParams{
		UserID: uuid.NullUUID{UUID: userID, Valid: true},
		Token:  token,
	}
	return r.q.InsertEmailVerification(ctx, arg)
}

func (r *EmailVerificationRepo) GetByToken(ctx context.Context, token uuid.UUID) (db.EmailVerification, error) {
	return r.q.GetEmailVerificationByToken(ctx, token)
}

func (r *EmailVerificationRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.q.DeleteEmailVerificationByUserID(ctx, uuid.NullUUID{UUID: userID, Valid: true})
}
