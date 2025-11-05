
package adapters

import (
	"context"
    "log"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/http_server"
)

// userServiceAdapter wraps services.UserService to implement http_server.UserService
type userServiceAdapter struct {
	svc *services.AuthService
}

// NewUserServiceAdapter creates a new adapter
func NewUserServiceAdapter(svc *services.AuthService) http_server.UserService {
	return &userServiceAdapter{svc: svc}
}

// GetUserFromToken converts db.User to http_server.User
func (a *userServiceAdapter) GetUserFromToken(ctx context.Context, token string) (*http_server.User, error) {
	user, err := a.svc.GetUserBySessionToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return &http_server.User{
		ID:    user.UserID,
		Name:  user.Name,
		Email: user.Email,
		// Roles can be fetched from RoleRepo if needed
		Roles: []string{},
	}, nil
}

func (a *userServiceAdapter) UserHasPermission(ctx context.Context, userID uuid.UUID, permission string) bool {
	return a.svc.UserHasPermission(ctx, userID, permission)
}

func (a *userServiceAdapter) LogActivity(ctx context.Context, action string, entityID uuid.UUID, status string) {
	if err := a.svc.LogUserActivity(ctx, action, entityID, status); err != nil {
		log.Printf("[LogActivity] error: %v\n", err)
	}
}

func (a *userServiceAdapter) NotifySuperAdminsIfNeeded(ctx context.Context, entity interface{}) {
	if err := a.svc.NotifySuperAdmins(ctx, entity); err != nil {
		log.Printf("[NotifySuperAdmins] error: %v\n", err)
	}
}