package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/rustybrownlee-llm/bastion/poc/internal/config"
)

func RequireAuth(cfg *config.AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeError(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			token := parts[1]
			claims, err := ValidateAccessToken(cfg, token)
			if err != nil {
				writeError(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "claims", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
