package apikey

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rustybrownlee-llm/bastion/poc/internal/audit"
	"github.com/rustybrownlee-llm/bastion/poc/internal/auth"
)

type Handler struct {
	service     *Service
	auditLogger *audit.Logger
}

func NewHandler(service *Service, auditLogger *audit.Logger) *Handler {
	return &Handler{service: service, auditLogger: auditLogger}
}

func (h *Handler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name          string   `json:"name"`
		Description   string   `json:"description"`
		TenantID      *string  `json:"tenant_id"`
		ExpiresAt     *string  `json:"expires_at"`
		PermissionIDs []string `json:"permission_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		writeError(w, "name is required", http.StatusBadRequest)
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			writeError(w, "invalid expires_at format", http.StatusBadRequest)
			return
		}
		expiresAt = &parsed
	}

	key, fullKey, err := h.service.Create(req.Name, req.Description, req.TenantID, expiresAt, req.PermissionIDs)
	if err != nil {
		writeError(w, "failed to create api key", http.StatusInternalServerError)
		return
	}

	claims := r.Context().Value("claims").(*auth.Claims)
	h.auditLogger.Log("api_key.created", claims.UserID, map[string]interface{}{
		"api_key_id": key.ID,
		"name":       key.Name,
	}, r.RemoteAddr)

	resp := map[string]interface{}{
		"id":          key.ID,
		"name":        key.Name,
		"description": key.Description,
		"api_key":     fullKey,
		"key_prefix":  key.KeyPrefix,
		"tenant_id":   key.TenantID,
		"expires_at":  key.ExpiresAt,
		"enabled":     key.Enabled,
		"created_at":  key.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*auth.Claims)

	keys, err := h.service.List(claims.TenantID)
	if err != nil {
		writeError(w, "failed to list api keys", http.StatusInternalServerError)
		return
	}

	resp := make([]map[string]interface{}, len(keys))
	for i, key := range keys {
		resp[i] = map[string]interface{}{
			"id":           key.ID,
			"name":         key.Name,
			"description":  key.Description,
			"key_prefix":   key.KeyPrefix,
			"tenant_id":    key.TenantID,
			"expires_at":   key.ExpiresAt,
			"enabled":      key.Enabled,
			"created_at":   key.CreatedAt,
			"last_used_at": key.LastUsedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"api_keys": resp})
}

func (h *Handler) GetAPIKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	key, err := h.service.GetByID(id)
	if err != nil {
		writeError(w, "api key not found", http.StatusNotFound)
		return
	}

	permissions, _ := h.service.GetPermissions(key.ID)

	resp := map[string]interface{}{
		"id":           key.ID,
		"name":         key.Name,
		"description":  key.Description,
		"key_prefix":   key.KeyPrefix,
		"tenant_id":    key.TenantID,
		"expires_at":   key.ExpiresAt,
		"enabled":      key.Enabled,
		"created_at":   key.CreatedAt,
		"last_used_at": key.LastUsedAt,
		"permissions":  permissions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(id); err != nil {
		writeError(w, "failed to delete api key", http.StatusInternalServerError)
		return
	}

	claims := r.Context().Value("claims").(*auth.Claims)
	h.auditLogger.Log("api_key.deleted", claims.UserID, map[string]interface{}{
		"api_key_id": id,
	}, r.RemoteAddr)

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AddPermission(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		PermissionID string `json:"permission_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.service.AddPermission(id, req.PermissionID); err != nil {
		writeError(w, "failed to add permission", http.StatusInternalServerError)
		return
	}

	claims := r.Context().Value("claims").(*auth.Claims)
	h.auditLogger.Log("api_key.permission_added", claims.UserID, map[string]interface{}{
		"api_key_id":    id,
		"permission_id": req.PermissionID,
	}, r.RemoteAddr)

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemovePermission(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	permissionID := chi.URLParam(r, "permId")

	if err := h.service.RemovePermission(id, permissionID); err != nil {
		writeError(w, "failed to remove permission", http.StatusInternalServerError)
		return
	}

	claims := r.Context().Value("claims").(*auth.Claims)
	h.auditLogger.Log("api_key.permission_removed", claims.UserID, map[string]interface{}{
		"api_key_id":    id,
		"permission_id": permissionID,
	}, r.RemoteAddr)

	w.WriteHeader(http.StatusNoContent)
}

func writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
