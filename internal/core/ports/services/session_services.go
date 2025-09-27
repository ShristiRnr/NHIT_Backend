package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

type SessionService struct {
	repo ports.SessionRepository
}

func NewSessionService(repo ports.SessionRepository) *SessionService {
	return &SessionService{repo: repo}
}

// CreateSession creates a new session for a user
func (s *SessionService) CreateSession(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (db.Session, error) {
	return s.repo.Create(ctx, userID, token, expiresAt)
}

// GetSession retrieves a session by ID
func (s *SessionService) GetSession(ctx context.Context, sessionID uuid.UUID) (db.Session, error) {
	return s.repo.Get(ctx, sessionID)
}

// DeleteSession deletes a session by ID
func (s *SessionService) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	return s.repo.Delete(ctx, sessionID)
}
