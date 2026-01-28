package rbac

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rustybrownlee-llm/bastion/poc/internal/auth"
)

type Handler struct {
	service *Service
}

type AssignRoleRequest struct {
	UserID   string  `json:"user_id"`
	TenantID *string `json:"tenant_id"`
}

type RevokeRoleRequest struct {
	UserID   string  `json:"user_id"`
	TenantID *string `json:"tenant_id"`
}

type AuthzCheckRequest struct {
	UserID       string  `json:"user_id"`
	TenantID     *string `json:"tenant_id"`
	ResourceType string  `json:"resource_type"`
	Action       string  `json:"action"`
}

type AuthzCheckResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

type PermissionResponse struct {
	ResourceType string `json:"resource_type"`
	Action       string `json:"action"`
}

type UserPermissionsResponse struct {
	UserID      string               `json:"user_id"`
	TenantID    *string              `json:"tenant_id"`
	Permissions []PermissionResponse `json:"permissions"`
}

type RoleResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	GrantedAt string  `json:"granted_at"`
	TenantID  *string `json:"tenant_id,omitempty"`
}

type UserRolesResponse struct {
	UserID   string         `json:"user_id"`
	TenantID *string        `json:"tenant_id"`
	Roles    []RoleResponse `json:"roles"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) AssignRole(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "roleId")
	if roleID == "" {
		writeError(w, "role ID required", http.StatusBadRequest)
		return
	}

	var req AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	claims := r.Context().Value("claims").(*auth.Claims)
	grantedBy := &claims.UserID

	if err := h.service.AssignRole(req.UserID, roleID, req.TenantID, grantedBy); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "role assigned",
		"role":      roleID,
		"user_id":   req.UserID,
		"tenant_id": req.TenantID,
	})
}

func (h *Handler) RevokeRole(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "roleId")
	if roleID == "" {
		writeError(w, "role ID required", http.StatusBadRequest)
		return
	}

	var req RevokeRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.RevokeRole(req.UserID, roleID, req.TenantID); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		writeError(w, "user ID required", http.StatusBadRequest)
		return
	}

	tenantIDStr := r.URL.Query().Get("tenant_id")
	var tenantID *string
	if tenantIDStr != "" {
		tenantID = &tenantIDStr
	}

	roles, err := h.service.GetUserRoles(userID, tenantID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := UserRolesResponse{
		UserID:   userID,
		TenantID: tenantID,
		Roles:    make([]RoleResponse, 0, len(roles)),
	}

	for _, role := range roles {
		resp.Roles = append(resp.Roles, RoleResponse{
			ID:        role.RoleID,
			Name:      role.RoleName,
			GrantedAt: role.GrantedAt.Format("2006-01-02T15:04:05Z07:00"),
			TenantID:  role.TenantID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetUserPermissions(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		writeError(w, "user ID required", http.StatusBadRequest)
		return
	}

	tenantIDStr := r.URL.Query().Get("tenant_id")
	var tenantID *string
	if tenantIDStr != "" {
		tenantID = &tenantIDStr
	}

	perms, err := h.service.GetUserPermissions(userID, tenantID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := UserPermissionsResponse{
		UserID:      userID,
		TenantID:    tenantID,
		Permissions: make([]PermissionResponse, 0, len(perms)),
	}

	for _, perm := range perms {
		resp.Permissions = append(resp.Permissions, PermissionResponse{
			ResourceType: perm.ResourceType,
			Action:       perm.Action,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) CheckAuthorization(w http.ResponseWriter, r *http.Request) {
	var req AuthzCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	allowed, reason, err := h.service.CheckPermission(req.UserID, req.TenantID, req.ResourceType, req.Action)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := AuthzCheckResponse{
		Allowed: allowed,
		Reason:  reason,
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
