
package adapters

import (
	"context"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/http_server"
)

// userServiceAdapter wraps services.UserService to implement http_server.UserService
type userServiceAdapter struct {
	svc *services.UserService
}

// NewUserServiceAdapter creates a new adapter
func NewUserServiceAdapter(svc *services.UserService) http_server.UserService {
	return &userServiceAdapter{svc: svc}
}

// GetUserFromToken converts db.User to http_server.User
func (a *userServiceAdapter) GetUserFromToken(ctx context.Context, token string) (*http_server.User, error) {
    authUser, err := a.svc.GetUserFromToken(ctx, token) // returns *services.AuthUser
    if err != nil {
        return nil, err
    }

    return &http_server.User{
        ID:    authUser.ID,
        Name:  authUser.Name,
        Email: authUser.Email,
        Roles: authUser.Roles,
    }, nil
}

// UserHasPermission delegates to services.UserService
func (a *userServiceAdapter) UserHasPermission(ctx context.Context, userID uuid.UUID, permission string) bool {
	return a.svc.UserHasPermission(ctx, userID, permission)
}

// LogActivity stub
func (a *userServiceAdapter) LogActivity(ctx context.Context, action string, entityID uuid.UUID, status string) {
	//in future
}

// NotifySuperAdminsIfNeeded stub
func (a *userServiceAdapter) NotifySuperAdminsIfNeeded(ctx context.Context, entity interface{}) {
	//in future
}
