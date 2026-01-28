package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rustybrownlee-llm/bastion/poc/internal/audit"
	"github.com/rustybrownlee-llm/bastion/poc/internal/config"
)

type Handler struct {
	service *Service
	audit   *audit.Logger
	cfg     *config.AuthConfig
}

func NewHandler(service *Service, audit *audit.Logger, cfg *config.AuthConfig) *Handler {
	return &Handler{
		service: service,
		audit:   audit,
		cfg:     cfg,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		h.audit.Log("login_failure", "", map[string]interface{}{
			"email": req.Email,
			"error": err.Error(),
		}, getIP(r))
		writeError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	claims, _ := ValidateAccessToken(h.cfg, accessToken)
	h.audit.Log("login_success", claims.UserID, map[string]interface{}{
		"email": req.Email,
	}, getIP(r))

	resp := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(h.cfg.AccessTokenTTL.Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}

	accessToken, err := h.service.Refresh(req.RefreshToken)
	if err != nil {
		h.audit.Log("token_refresh_failure", "", map[string]interface{}{
			"error": err.Error(),
		}, getIP(r))
		writeError(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	claims, _ := ValidateAccessToken(h.cfg, accessToken)
	h.audit.Log("token_refresh", claims.UserID, nil, getIP(r))

	resp := RefreshResponse{
		AccessToken: accessToken,
		ExpiresIn:   int(h.cfg.AccessTokenTTL.Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*Claims)

	if err := h.service.Logout(claims.UserID); err != nil {
		writeError(w, "logout failed", http.StatusInternalServerError)
		return
	}

	h.audit.Log("logout", claims.UserID, nil, getIP(r))
	w.WriteHeader(http.StatusNoContent)
}

func writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}
	return r.RemoteAddr
}
