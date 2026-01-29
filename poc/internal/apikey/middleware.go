package apikey

import (
	"context"
	"encoding/json"
	"net/http"
)

type APIKeyContext struct {
	APIKeyID string
	TenantID *string
}

func AuthenticateAPIKey(service *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKeyHeader := r.Header.Get("X-API-Key")

			if apiKeyHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			key, err := service.Authenticate(apiKeyHeader)
			if err != nil {
				writeAuthError(w, "invalid api key", http.StatusUnauthorized)
				return
			}

			keyCtx := &APIKeyContext{
				APIKeyID: key.ID,
				TenantID: key.TenantID,
			}

			ctx := context.WithValue(r.Context(), "apikey", keyCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeAuthError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
