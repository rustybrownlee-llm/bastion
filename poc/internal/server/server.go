package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rustybrownlee-llm/bastion/poc/internal/apikey"
	"github.com/rustybrownlee-llm/bastion/poc/internal/audit"
	"github.com/rustybrownlee-llm/bastion/poc/internal/auth"
	"github.com/rustybrownlee-llm/bastion/poc/internal/config"
	"github.com/rustybrownlee-llm/bastion/poc/internal/rbac"
	"github.com/rustybrownlee-llm/bastion/poc/internal/serviceaccount"
	"github.com/rustybrownlee-llm/bastion/poc/internal/tenant"
	"github.com/rustybrownlee-llm/bastion/poc/internal/user"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

func New(db *sql.DB, cfg *config.Config, auditLogger *audit.Logger) http.Handler {
	r := chi.NewRouter()

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService, auditLogger)

	authService := auth.NewService(db, &cfg.Auth)
	authHandler := auth.NewHandler(authService, auditLogger, &cfg.Auth)

	tenantRepo := tenant.NewRepository(db)
	tenantService := tenant.NewService(tenantRepo)
	tenantHandler := tenant.NewHandler(tenantService, auditLogger)

	rbacRepo := rbac.NewRepository(db)
	rbacService := rbac.NewService(rbacRepo, auditLogger)
	rbacHandler := rbac.NewHandler(rbacService)

	serviceAccountRepo := serviceaccount.NewRepository(db)
	serviceAccountService := serviceaccount.NewService(serviceAccountRepo, &cfg.Auth)
	serviceAccountHandler := serviceaccount.NewHandler(serviceAccountService, auditLogger)

	apiKeyRepo := apikey.NewRepository(db)
	apiKeyService := apikey.NewService(apiKeyRepo)
	apiKeyHandler := apikey.NewHandler(apiKeyService, auditLogger)

	r.Get("/health", handleHealth)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/users", userHandler.CreateUser)

		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/refresh", authHandler.Refresh)
		r.Post("/auth/token", serviceAccountHandler.ClientCredentialsToken)

		r.Group(func(r chi.Router) {
			r.Use(apikey.AuthenticateAPIKey(apiKeyService))
			r.Use(auth.RequireAuth(&cfg.Auth))
			r.Post("/auth/logout", authHandler.Logout)
			r.Get("/users/me", userHandler.GetMe)

			r.Route("/tenants", func(r chi.Router) {
				r.Use(rbac.RequirePermission(rbacService, "bastion:tenant", "create"))
				r.Post("/", tenantHandler.CreateTenant)
			})

			r.Get("/tenants", tenantHandler.ListTenants)
			r.Get("/tenants/{id}", tenantHandler.GetTenant)

			r.Post("/roles/{roleId}/assign", rbacHandler.AssignRole)
			r.Delete("/roles/{roleId}/assign", rbacHandler.RevokeRole)

			r.Get("/users/{userId}/roles", rbacHandler.GetUserRoles)
			r.Get("/users/{userId}/permissions", rbacHandler.GetUserPermissions)

			r.Post("/authz/check", rbacHandler.CheckAuthorization)

			r.Group(func(r chi.Router) {
				r.Use(rbac.RequirePermission(rbacService, "bastion:service-account", "create"))
				r.Post("/service-accounts", serviceAccountHandler.CreateServiceAccount)
			})

			r.Group(func(r chi.Router) {
				r.Use(rbac.RequirePermission(rbacService, "bastion:service-account", "read"))
				r.Get("/service-accounts", serviceAccountHandler.ListServiceAccounts)
				r.Get("/service-accounts/{id}", serviceAccountHandler.GetServiceAccount)
			})

			r.Group(func(r chi.Router) {
				r.Use(rbac.RequirePermission(rbacService, "bastion:service-account", "update"))
				r.Put("/service-accounts/{id}", serviceAccountHandler.UpdateServiceAccount)
				r.Post("/service-accounts/{id}/regenerate-secret", serviceAccountHandler.RegenerateSecret)
			})

			r.Group(func(r chi.Router) {
				r.Use(rbac.RequirePermission(rbacService, "bastion:service-account", "delete"))
				r.Delete("/service-accounts/{id}", serviceAccountHandler.DeleteServiceAccount)
			})

			r.Group(func(r chi.Router) {
				r.Use(rbac.RequirePermission(rbacService, "bastion:api-key", "create"))
				r.Post("/api-keys", apiKeyHandler.CreateAPIKey)
			})

			r.Group(func(r chi.Router) {
				r.Use(rbac.RequirePermission(rbacService, "bastion:api-key", "read"))
				r.Get("/api-keys", apiKeyHandler.ListAPIKeys)
				r.Get("/api-keys/{id}", apiKeyHandler.GetAPIKey)
			})

			r.Group(func(r chi.Router) {
				r.Use(rbac.RequirePermission(rbacService, "bastion:api-key", "delete"))
				r.Delete("/api-keys/{id}", apiKeyHandler.DeleteAPIKey)
			})

			r.Post("/api-keys/{id}/permissions", apiKeyHandler.AddPermission)
			r.Delete("/api-keys/{id}/permissions/{permId}", apiKeyHandler.RemovePermission)
		})
	})

	return r
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
