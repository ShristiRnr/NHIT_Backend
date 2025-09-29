package http_server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ShristiRnr/NHIT_Backend/internal/core/ports"
)

type DepartmentHandler struct {
	svc ports.DepartmentService
}

func NewDepartmentHandler(svc ports.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{svc: svc}
}

// Routes registers all department endpoints
func (h *DepartmentHandler) Routes(r chi.Router) {
	r.Get("/", h.List)         // GET /departments
	r.Post("/", h.Create)      // POST /departments
	r.Get("/{id}", h.Get)      // GET /departments/{id}
	r.Put("/{id}", h.Update)   // PUT /departments/{id}
	r.Delete("/{id}", h.Delete) // DELETE /departments/{id}
}

// ---------------- Handlers ----------------

func (h *DepartmentHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize == 0 {
		pageSize = 10
	}

	depts, err := h.svc.List(r.Context(), int32(page), int32(pageSize))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(depts)
}

func (h *DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dept, err := h.svc.Create(r.Context(), body.Name, body.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dept)
}

func (h *DepartmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	dept, err := h.svc.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dept)
}

func (h *DepartmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dept, err := h.svc.Update(r.Context(), id, body.Name, body.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dept)
}

func (h *DepartmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
