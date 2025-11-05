package http_server

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const currentUserKey = contextKey("currentUser")

// AuthMiddleware handles authentication and permission checks
type AuthMiddleware struct {
	userSvc UserService
}

// NewAuthMiddleware creates a new AuthMiddleware with the given UserService
func NewAuthMiddleware(svc UserService) *AuthMiddleware {
	return &AuthMiddleware{
		userSvc: svc,
	}
}

// UserService defines the interface needed by AuthMiddleware
type UserService interface {
	GetUserFromToken(ctx context.Context, token string) (*User, error)
	UserHasPermission(ctx context.Context, userID uuid.UUID, permission string) bool
	LogActivity(ctx context.Context, action string, entityID uuid.UUID, status string)
	NotifySuperAdminsIfNeeded(ctx context.Context, entity interface{})
}

// User represents a logged-in user
type User struct {
	ID    uuid.UUID
	Name  string
	Email string
	Roles []string
}

// RequirePermission returns middleware that ensures the user has at least one of the required permissions
func (a *AuthMiddleware) RequirePermission(perms ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := a.getCurrentUser(r)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if !a.hasAnyPermission(r.Context(), user.ID, perms) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			// Store user in context using custom type key
			ctx := context.WithValue(r.Context(), currentUserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// getCurrentUser extracts the user from the Authorization header
func (a *AuthMiddleware) getCurrentUser(r *http.Request) (*User, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("no authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, errors.New("invalid authorization header")
	}

	token := parts[1]
	user, err := a.userSvc.GetUserFromToken(r.Context(), token)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CurrentUser retrieves the current logged-in user from context
func CurrentUser(ctx context.Context) *User {
	u, _ := ctx.Value(currentUserKey).(*User)
	return u
}

// LogActivity logs an action performed by the user
func (a *AuthMiddleware) LogActivity(ctx context.Context, action string, entityID uuid.UUID, status string) {
	if a.userSvc != nil {
		a.userSvc.LogActivity(ctx, action, entityID, status)
	}
}

// NotifySuperAdminsIfNeeded triggers notifications for super admins
func (a *AuthMiddleware) NotifySuperAdminsIfNeeded(ctx context.Context, entity interface{}) {
	if a.userSvc != nil {
		a.userSvc.NotifySuperAdminsIfNeeded(ctx, entity)
	}
}

// IsCurrentUserInRole checks if the current user has a specific role
func (a *AuthMiddleware) IsCurrentUserInRole(roleName string, ctx context.Context) bool {
	user := CurrentUser(ctx)
	if user == nil {
		return false
	}

	for _, r := range user.Roles {
		if r == roleName {
			return true
		}
	}

	return false
}

// hasAnyPermission checks if the user has at least one permission
func (a *AuthMiddleware) hasAnyPermission(ctx context.Context, userID uuid.UUID, perms []string) bool {
	for _, p := range perms {
		if a.userSvc.UserHasPermission(ctx, userID, p) {
			return true
		}
	}
	return false
}


