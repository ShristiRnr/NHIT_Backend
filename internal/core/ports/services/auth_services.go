package services

import (
	"context"
	"errors"
	"time"
	"database/sql"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
)

type AuthService struct {
	userRepo    *db.Queries
	roleRepo    *db.Queries
	sessionRepo *db.Queries
}

func NewAuthService(u *db.Queries, r *db.Queries, s *db.Queries) *AuthService {
	return &AuthService{
		userRepo:    u,
		roleRepo:    r,
		sessionRepo: s,
	}
}

// Register a new user with default role assignment
func (s *AuthService) Register(ctx context.Context, tenantID uuid.UUID, name, email, password, roleName string) (db.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return db.User{}, err
	}

	// Create user
	user, err := s.userRepo.CreateUser(ctx, db.CreateUserParams{
		TenantID: tenantID,
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	})
	if err != nil {
		return db.User{}, err
	}

	// Get role by name (optional: create if doesn't exist)
	roles, err := s.roleRepo.ListRolesByTenant(ctx, tenantID)
	if err != nil {
		return db.User{}, err
	}

	var roleID uuid.UUID
	for _, r := range roles {
		if r.Name == roleName {
			roleID = r.RoleID
			break
		}
	}
	if roleID == uuid.Nil {
		newRole, err := s.roleRepo.CreateRole(ctx, db.CreateRoleParams{
			TenantID: tenantID,
			Name:     roleName,
		})
		if err != nil {
			return db.User{}, err
		}
		roleID = newRole.RoleID
	}

	// Assign role to user
	err = s.roleRepo.AssignRoleToUser(ctx, db.AssignRoleToUserParams{
		UserID: user.UserID,
		RoleID: roleID,
	})
	if err != nil {
		return db.User{}, err
	}

	return user, nil
}


func (s *AuthService) Login(ctx context.Context, email, password string, sessionDuration time.Duration) (string, error) {
	// Get user by email
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // <-- use sql.ErrNoRows
			return "", errors.New("invalid credentials")
		}
		return "", err
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate session token
	token := uuid.New().String()
	expireAt := time.Now().Add(sessionDuration)

	_, err = s.sessionRepo.CreateSession(ctx, db.CreateSessionParams{
		UserID:    uuid.NullUUID{UUID: user.UserID, Valid: true},
		Token:     token,
		ExpiresAt: sql.NullTime{Time: expireAt, Valid: true},
	})
	if err != nil {
		return "", err
	}

	return token, nil
}