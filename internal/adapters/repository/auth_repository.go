package repository

import (
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

// AuthRepo holds references to existing repos
type AuthRepo struct {
	UserRepo    ports.UserRepository
	RoleRepo    ports.RoleRepository
	SessionRepo ports.SessionRepository
	RefreshRepo ports.RefreshTokenRepository
}

// NewAuthRepo initializes AuthRepo with existing repos
func NewAuthRepo(
	userRepo ports.UserRepository,
	roleRepo ports.RoleRepository,
	sessionRepo ports.SessionRepository,
	refreshRepo ports.RefreshTokenRepository,
) *AuthRepo {
	return &AuthRepo{
		UserRepo:    userRepo,
		RoleRepo:    roleRepo,
		SessionRepo: sessionRepo,
		RefreshRepo: refreshRepo,
	}
}
