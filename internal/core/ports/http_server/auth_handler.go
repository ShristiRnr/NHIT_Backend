package http_server

import (
	"encoding/json"
	"net/http"
	"time"
	"log"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
)

type RegisterRequest struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
	RoleName string    `json:"role_name"`
}

type RegisterResponse struct {
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
	Email  string    `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type LogoutRequest struct {
	Token string `json:"token"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}

type AuthHandler struct {
	svc *services.AuthService
}

func NewAuthHandler(svc *services.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// ---------------------- Handlers ----------------------

// POST /register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.svc.Register(r.Context(), req.TenantID, req.Name, req.Email, req.Password, req.RoleName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log activity
	if err := h.svc.LogUserActivity(r.Context(), "User Registered", user.UserID, "success"); err != nil {
		log.Printf("Failed to log activity: %v", err)
	}

	// Notify super admins
	if err := h.svc.NotifySuperAdmins(r.Context(), user); err != nil {
		log.Printf("Failed to notify super admins: %v", err)
	}

	resp := RegisterResponse{
		UserID: user.UserID,
		Name:   user.Name,
		Email:  user.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// POST /login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Example: session lasts 24 hours, refresh token 7 days
	token, _, err := h.svc.Login(r.Context(), req.Email, req.Password, 24*time.Hour, 7*24*time.Hour)
	if err != nil {
		h.svc.LogUserActivity(r.Context(), "User Login Failed", uuid.Nil, "failed")
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	// Log successful login
	if err := h.svc.LogUserActivity(r.Context(), "User Logged In", uuid.Nil, "success"); err != nil {
		log.Printf("Failed to log activity: %v", err)
	}

	resp := LoginResponse{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// POST /logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.svc.GetUserBySessionToken(r.Context(), req.Token)
	if err != nil {
		http.Error(w, "invalid session", http.StatusUnauthorized)
		return
	}

	if err := h.svc.Logout(r.Context(), req.Token, req.Token); err != nil {
		http.Error(w, "failed to logout", http.StatusInternalServerError)
		return
	}

	// Log activity
	if err := h.svc.LogUserActivity(r.Context(), "User Logged Out", user.UserID, "success"); err != nil {
		log.Printf("Failed to log activity: %v", err)
	}

	resp := LogoutResponse{
		Message: "Logged out successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GET /me - returns the current user info (requires session token)
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "missing authorization token", http.StatusUnauthorized)
		return
	}

	user, err := h.svc.GetUserBySessionToken(r.Context(), token)
	if err != nil {
		http.Error(w, "invalid session", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(struct {
		UserID uuid.UUID `json:"user_id"`
		Name   string    `json:"name"`
		Email  string    `json:"email"`
	}{
		UserID: user.UserID,
		Name:   user.Name,
		Email:  user.Email,
	})
}

// Example middleware check for admin-only endpoints
func (h *AuthHandler) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "missing authorization token", http.StatusUnauthorized)
			return
		}

		user, err := h.svc.GetUserBySessionToken(r.Context(), token)
		if err != nil {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		if !h.svc.UserHasPermission(r.Context(), user.UserID, "admin") {
			http.Error(w, "permission denied", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}