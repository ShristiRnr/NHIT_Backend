package services

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

type AuthService struct {
	userRepo    ports.UserRepository
	roleRepo    ports.RoleRepository
	sessionRepo ports.SessionRepository
	refreshRepo ports.RefreshTokenRepository
}

// AuthUser represents a logged-in user
type AuthUser struct {
	ID    uuid.UUID
	Name  string
	Email string
	Roles []string
}

// Constructor
func NewAuthService(
	u ports.UserRepository,
	r ports.RoleRepository,
	s ports.SessionRepository,
	refresh ports.RefreshTokenRepository,
) *AuthService {
	return &AuthService{
		userRepo:    u,
		roleRepo:    r,
		sessionRepo: s,
		refreshRepo: refresh,
	}
}

// Register a new user with role assignment
func (s *AuthService) Register(ctx context.Context, tenantID uuid.UUID, name, email, password, roleName string) (db.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return db.User{}, err
	}

	user, err := s.userRepo.Create(ctx, db.CreateUserParams{
		TenantID: tenantID,
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	})
	if err != nil {
		return db.User{}, err
	}

	roles, err := s.roleRepo.List(ctx, tenantID)
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
		newRole, err := s.roleRepo.Create(ctx, db.CreateRoleParams{
			TenantID: tenantID,
			Name:     roleName,
		})
		if err != nil {
			return db.User{}, err
		}
		roleID = newRole.RoleID
	}

	err = s.roleRepo.AssignRoleToUser(ctx, db.AssignRoleToUserParams{
		UserID: user.UserID,
		RoleID: roleID,
	})
	if err != nil {
		return db.User{}, err
	}

	return user, nil
}

// Login authenticates a user and creates session + refresh token
func (s *AuthService) Login(ctx context.Context, email, password string, sessionDuration time.Duration, refreshDuration time.Duration) (sessionToken string, refreshToken string, err error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", errors.New("invalid credentials")
		}
		return "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", errors.New("invalid credentials")
	}

	// Create session token
	sessionToken = uuid.New().String()
	sessionExpire := time.Now().Add(sessionDuration)

	_, err = s.sessionRepo.Create(ctx, user.UserID, sessionToken, sessionExpire)
	if err != nil {
		return "", "", err
	}

	// Create refresh token
	refreshToken = uuid.New().String()
	refreshExpire := time.Now().Add(refreshDuration)

	err = s.refreshRepo.Create(ctx, user.UserID, refreshToken, refreshExpire)
	if err != nil {
		return "", "", err
	}

	return sessionToken, refreshToken, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionToken, refreshToken string) error {
    // Delete session by token
    if err := s.sessionRepo.Delete(ctx, sessionToken); err != nil {
        return err
    }

    // Delete refresh token
    if err := s.refreshRepo.Delete(ctx, refreshToken); err != nil {
        return err
    }

    return nil
}

// UserHasPermission checks if a user has a specific permission
func (s *AuthService) UserHasPermission(ctx context.Context, userID uuid.UUID, permission string) bool {
	// TODO: implement proper permission check
	// For now, allow everything for demonstration
	return true
}

// LogUserActivity logs a user action
func (s *AuthService) LogUserActivity(ctx context.Context, action string, entityID uuid.UUID, status string) error {
	// You could write to a table `user_activities` here
	log.Printf("[UserActivity] Action=%s, EntityID=%s, Status=%s\n", action, entityID, status)
	return nil
}

// NotifySuperAdmins sends notifications to all super admins
func (s *AuthService) NotifySuperAdmins(ctx context.Context, entity interface{}) error {
    superAdmins, err := s.roleRepo.ListSuperAdmins(ctx)
    if err != nil {
        return err
    }

    for _, admin := range superAdmins {
        // for now, just log; later you can send email
        log.Printf("[NotifySuperAdmins] Notify %s about %v\n", admin.Email, entity)
    }

    return nil
}


// GetUserBySessionToken retrieves user info from a session token
func (s *AuthService) GetUserBySessionToken(ctx context.Context, token string) (*db.User, error) {
	// Fetch session
	session, err := s.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, errors.New("invalid or expired session")
	}

	// Check if UserID is valid
	if !session.UserID.Valid {
		return nil, errors.New("invalid session user")
	}
	userID := session.UserID.UUID

	// Get user by ID
	user, err := s.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Optional: check if session is expired
	if !session.ExpiresAt.Valid || time.Now().After(session.ExpiresAt.Time) {
		return nil, errors.New("session expired")
	}

	return &user, nil
}