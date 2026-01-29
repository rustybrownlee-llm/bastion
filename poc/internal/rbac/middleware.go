package rbac

import (
	"encoding/json"
	"net/http"

	"github.com/rustybrownlee-llm/bastion/poc/internal/auth"
)

type APIKeyContext struct {
	APIKeyID string
	TenantID *string
}

func RequirePermission(service *Service, resourceType, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKeyCtx := r.Context().Value("apikey")
			if apiKeyCtx != nil {
				keyInfo := apiKeyCtx.(*APIKeyContext)
				allowed, err := service.CheckAPIKeyPermission(keyInfo.APIKeyID, resourceType, action)
				if err != nil {
					writeAuthError(w, "authorization check failed", http.StatusInternalServerError)
					return
				}
				if !allowed {
					writeAuthError(w, "permission denied", http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			claims := r.Context().Value("claims").(*auth.Claims)
			if claims == nil {
				writeAuthError(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			allowed, reason, err := service.CheckPermission(
				claims.UserID,
				claims.TenantID,
				resourceType,
				action,
			)

			if err != nil {
				writeAuthError(w, "authorization check failed", http.StatusInternalServerError)
				return
			}

			if !allowed {
				writeAuthError(w, reason, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeAuthError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
