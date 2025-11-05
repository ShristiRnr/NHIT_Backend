package http_server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
)

// PaginationHandler handles HTTP requests for paginated users
type PaginationHandler struct {
	svc *services.PaginationService
}

// NewPaginationHandler creates a new PaginationHandler
func NewPaginationHandler(svc *services.PaginationService) *PaginationHandler {
	return &PaginationHandler{svc: svc}
}

// RegisterRoutes registers pagination routes
func (h *PaginationHandler) RegisterRoutes(r chi.Router) {
	r.Get("/tenants/{tenantID}/users", h.ListUsers)
}

// ListUsers handles GET /tenants/{tenantID}/users?limit=10&offset=0
func (h *PaginationHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "tenantID")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "invalid tenantID", http.StatusBadRequest)
		return
	}

	// Parse limit and offset query params
	limit := int32(10)
	offset := int32(0)

	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = int32(v)
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = int32(v)
		}
	}

	users, err := h.svc.ListUsers(r.Context(), tenantID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Optionally return total count
	total, err := h.svc.CountUsers(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Total int64      `json:"total"`
		Users []db.User  `json:"users"`
	}{
		Total: total,
		Users: users,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
