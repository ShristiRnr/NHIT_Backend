package http_server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/go-chi/chi/v5"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
	"github.com/ShristiRnr/NHIT_Backend/internal/adapters/database/db"
	"github.com/ShristiRnr/NHIT_Backend/helpers"
)

// VendorHandler handles HTTP requests for vendors
type VendorHandler struct {
	svc   ports.VendorService
	auth  *AuthMiddleware
}

// NewVendorHandler creates a new VendorHandler
func NewVendorHandler(svc ports.VendorService, auth *AuthMiddleware) *VendorHandler {
	return &VendorHandler{svc: svc, auth: auth}
}

// Routes sets up vendor routes with middleware
func (h *VendorHandler) Routes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(h.auth.RequirePermission("view-vendors"))
		r.Get("/", h.List)
		r.Get("/search", h.Search)
		r.Get("/{id}", h.Show)
	})

	r.Group(func(r chi.Router) {
		r.Use(h.auth.RequirePermission("create-vendors"))
		r.Post("/", h.Create)
	})

	r.Group(func(r chi.Router) {
		r.Use(h.auth.RequirePermission("edit-vendors"))
		r.Put("/{id}", h.Update)
	})

	r.Group(func(r chi.Router) {
		r.Use(h.auth.RequirePermission("delete-vendors"))
		r.Delete("/{id}", h.Delete)
	})
}

// List returns all vendors (active only for non-admins)
func (h *VendorHandler) List(w http.ResponseWriter, r *http.Request) {
	onlyActive := !h.auth.IsCurrentUserInRole("Admin", r.Context())

	vendors, err := h.svc.List(r.Context(), onlyActive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, vendors)
}



// Show returns a single vendor by ID
func (h *VendorHandler) Show(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid vendor ID", http.StatusBadRequest)
		return
	}

	vendor, err := h.svc.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "vendor not found", http.StatusNotFound)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, vendor)
}

// Create adds a new vendor
func (h *VendorHandler) Create(w http.ResponseWriter, r *http.Request) {
	var v db.Vendor
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.svc.Create(r.Context(), v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.auth.LogActivity(r.Context(), "Vendor Created", created.ID, "success")
	helpers.WriteJSON(w, http.StatusCreated, created)
}

// Update modifies an existing vendor
func (h *VendorHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid vendor ID", http.StatusBadRequest)
		return
	}

	var v db.Vendor
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updated, err := h.svc.Update(r.Context(), id, v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.auth.LogActivity(r.Context(), "Vendor Updated", updated.ID, "success")
	helpers.WriteJSON(w, http.StatusOK, updated)
}

// Delete removes a vendor
func (h *VendorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid vendor ID", http.StatusBadRequest)
		return
	}

	_, err = h.svc.Update(r.Context(), id, db.Vendor{Active: "N"}) // soft delete
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.auth.LogActivity(r.Context(), "Vendor Deleted", id, "success")
	w.WriteHeader(http.StatusNoContent)
}

// Search vendors with query, limit, offset
func (h *VendorHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	results, err := h.svc.Search(r.Context(), query, int32(limit), int32(offset))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, results)
}
