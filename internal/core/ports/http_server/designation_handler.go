package http_server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports/services"
)

type DesignationHandler struct {
	svc *services.DesignationService
}

func NewDesignationHandler(svc *services.DesignationService) *DesignationHandler {
	return &DesignationHandler{svc: svc}
}

func (h *DesignationHandler) Routes(r chi.Router) {
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
}

func (h *DesignationHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize == 0 {
		pageSize = 10
	}

	designations, err := h.svc.List(r.Context(), int32(page), int32(pageSize))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(designations)
}

func (h *DesignationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	designation, err := h.svc.Create(r.Context(), body.Name, body.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(designation)
}

func (h *DesignationHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	designation, err := h.svc.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(designation)
}

func (h *DesignationHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	designation, err := h.svc.Update(r.Context(), id, body.Name, body.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(designation)
}

func (h *DesignationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
