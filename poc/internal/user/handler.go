package user

import (
	"encoding/json"
	"net/http"

	"github.com/rustybrownlee-llm/bastion/poc/internal/audit"
	"github.com/rustybrownlee-llm/bastion/poc/internal/auth"
)

type Handler struct {
	service *Service
	audit   *audit.Logger
}

func NewHandler(service *Service, audit *audit.Logger) *Handler {
	return &Handler{
		service: service,
		audit:   audit,
	}
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.service.CreateUser(req.Email, req.Password)
	if err != nil {
		h.audit.Log("user_creation_failure", "", map[string]interface{}{
			"email": req.Email,
			"error": err.Error(),
		}, getIP(r))
		writeError(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	h.audit.Log("user_created", user.ID, map[string]interface{}{
		"email": user.Email,
	}, getIP(r))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*auth.Claims)

	user, err := h.service.GetByID(claims.UserID)
	if err != nil {
		writeError(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func getIP(r *http.Request) string {
	return r.RemoteAddr
}
