package tenant

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rustybrownlee-llm/bastion/poc/internal/audit"
)

type Handler struct {
	service     *Service
	auditLogger *audit.Logger
}

type CreateTenantRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type TenantResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"created_at"`
}

type ListTenantsResponse struct {
	Tenants []TenantResponse `json:"tenants"`
}

func NewHandler(service *Service, auditLogger *audit.Logger) *Handler {
	return &Handler{
		service:     service,
		auditLogger: auditLogger,
	}
}

func (h *Handler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tenant, err := h.service.CreateTenant(req.Name, req.Slug)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.auditLogger.Log("tenant.created", "", map[string]interface{}{
		"tenant_id": tenant.ID,
		"name":      tenant.Name,
		"slug":      tenant.Slug,
	}, "")

	resp := TenantResponse{
		ID:        tenant.ID,
		Name:      tenant.Name,
		Slug:      tenant.Slug,
		CreatedAt: tenant.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) ListTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.service.ListTenants()
	if err != nil {
		writeError(w, "failed to list tenants", http.StatusInternalServerError)
		return
	}

	resp := ListTenantsResponse{
		Tenants: make([]TenantResponse, 0, len(tenants)),
	}

	for _, t := range tenants {
		resp.Tenants = append(resp.Tenants, TenantResponse{
			ID:        t.ID,
			Name:      t.Name,
			Slug:      t.Slug,
			CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "id")
	if tenantID == "" {
		writeError(w, "tenant ID required", http.StatusBadRequest)
		return
	}

	tenant, err := h.service.GetTenant(tenantID)
	if err != nil {
		writeError(w, "tenant not found", http.StatusNotFound)
		return
	}

	resp := TenantResponse{
		ID:        tenant.ID,
		Name:      tenant.Name,
		Slug:      tenant.Slug,
		CreatedAt: tenant.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
