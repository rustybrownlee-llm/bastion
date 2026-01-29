package serviceaccount

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rustybrownlee-llm/bastion/poc/internal/audit"
	"github.com/rustybrownlee-llm/bastion/poc/internal/auth"
)

type Handler struct {
	service      *Service
	auditLogger  *audit.Logger
}

func NewHandler(service *Service, auditLogger *audit.Logger) *Handler {
	return &Handler{service: service, auditLogger: auditLogger}
}

func (h *Handler) CreateServiceAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		TenantID    *string  `json:"tenant_id"`
		RoleIDs     []string `json:"role_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		writeError(w, "name is required", http.StatusBadRequest)
		return
	}

	sa, secret, err := h.service.Create(req.Name, req.Description, req.TenantID, req.RoleIDs)
	if err != nil {
		writeError(w, "failed to create service account", http.StatusInternalServerError)
		return
	}

	claims := r.Context().Value("claims").(*auth.Claims)
	h.auditLogger.Log("service_account.created", claims.UserID, map[string]interface{}{
		"service_account_id": sa.ID,
		"name":               sa.Name,
	}, r.RemoteAddr)

	resp := map[string]interface{}{
		"id":            sa.ID,
		"name":          sa.Name,
		"description":   sa.Description,
		"client_id":     sa.ClientID,
		"client_secret": secret,
		"tenant_id":     sa.TenantID,
		"enabled":       sa.Enabled,
		"created_at":    sa.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) ListServiceAccounts(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*auth.Claims)

	accounts, err := h.service.List(claims.TenantID)
	if err != nil {
		writeError(w, "failed to list service accounts", http.StatusInternalServerError)
		return
	}

	resp := make([]map[string]interface{}, len(accounts))
	for i, sa := range accounts {
		resp[i] = map[string]interface{}{
			"id":           sa.ID,
			"name":         sa.Name,
			"description":  sa.Description,
			"client_id":    sa.ClientID,
			"tenant_id":    sa.TenantID,
			"enabled":      sa.Enabled,
			"created_at":   sa.CreatedAt,
			"last_used_at": sa.LastUsedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"service_accounts": resp})
}

func (h *Handler) GetServiceAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	sa, err := h.service.GetByID(id)
	if err != nil {
		writeError(w, "service account not found", http.StatusNotFound)
		return
	}

	resp := map[string]interface{}{
		"id":           sa.ID,
		"name":         sa.Name,
		"description":  sa.Description,
		"client_id":    sa.ClientID,
		"tenant_id":    sa.TenantID,
		"enabled":      sa.Enabled,
		"created_at":   sa.CreatedAt,
		"updated_at":   sa.UpdatedAt,
		"last_used_at": sa.LastUsedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) UpdateServiceAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Enabled     *bool  `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	if err := h.service.Update(id, req.Name, req.Description, enabled); err != nil {
		writeError(w, "failed to update service account", http.StatusInternalServerError)
		return
	}

	claims := r.Context().Value("claims").(*auth.Claims)
	h.auditLogger.Log("service_account.updated", claims.UserID, map[string]interface{}{
		"service_account_id": id,
	}, r.RemoteAddr)

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteServiceAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(id); err != nil {
		writeError(w, "failed to delete service account", http.StatusInternalServerError)
		return
	}

	claims := r.Context().Value("claims").(*auth.Claims)
	h.auditLogger.Log("service_account.deleted", claims.UserID, map[string]interface{}{
		"service_account_id": id,
	}, r.RemoteAddr)

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RegenerateSecret(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	sa, err := h.service.GetByID(id)
	if err != nil {
		writeError(w, "service account not found", http.StatusNotFound)
		return
	}

	newSecret, err := h.service.RegenerateSecret(id)
	if err != nil {
		writeError(w, "failed to regenerate secret", http.StatusInternalServerError)
		return
	}

	claims := r.Context().Value("claims").(*auth.Claims)
	h.auditLogger.Log("service_account.secret_regenerated", claims.UserID, map[string]interface{}{
		"service_account_id": id,
	}, r.RemoteAddr)

	resp := map[string]interface{}{
		"client_id":     sa.ClientID,
		"client_secret": newSecret,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) ClientCredentialsToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeOAuthError(w, "invalid_request", "failed to parse form")
		return
	}

	grantType := r.FormValue("grant_type")
	if grantType != "client_credentials" {
		writeOAuthError(w, "unsupported_grant_type", "only client_credentials supported")
		return
	}

	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	if clientID == "" || clientSecret == "" {
		writeOAuthError(w, "invalid_request", "client_id and client_secret required")
		return
	}

	accessToken, err := h.service.Authenticate(clientID, clientSecret)
	if err != nil {
		h.auditLogger.LogError("service_account.auth_failed", err, r.RemoteAddr)
		writeOAuthError(w, "invalid_client", "invalid client credentials")
		return
	}

	sa, _ := h.service.repo.GetByClientID(clientID)
	if sa != nil {
		h.auditLogger.Log("service_account.authenticated", sa.ID, nil, r.RemoteAddr)
	}

	resp := map[string]interface{}{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   900,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func writeOAuthError(w http.ResponseWriter, code, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"error":             code,
		"error_description": description,
	})
}
