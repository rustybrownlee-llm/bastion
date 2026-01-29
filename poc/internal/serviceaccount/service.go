package serviceaccount

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rustybrownlee-llm/bastion/poc/internal/config"
)

type Service struct {
	repo *Repository
	cfg  *config.AuthConfig
}

func NewService(repo *Repository, cfg *config.AuthConfig) *Service {
	return &Service{repo: repo, cfg: cfg}
}

func (s *Service) Create(name, description string, tenantID *string, roleIDs []string) (*ServiceAccount, string, error) {
	clientID := generateClientID()
	clientSecret := generateClientSecret()

	hash, err := bcrypt.GenerateFromPassword([]byte(clientSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash client secret: %w", err)
	}

	sa, err := s.repo.Create(name, description, clientID, string(hash), tenantID)
	if err != nil {
		return nil, "", err
	}

	for _, roleID := range roleIDs {
		if err := s.repo.AssignRole(sa.ID, roleID); err != nil {
			return nil, "", fmt.Errorf("assign role: %w", err)
		}
	}

	return sa, clientSecret, nil
}

func (s *Service) Authenticate(clientID, clientSecret string) (string, error) {
	sa, err := s.repo.GetByClientID(clientID)
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	if !sa.Enabled {
		return "", fmt.Errorf("service account disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(sa.ClientSecretHash), []byte(clientSecret)); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	if err := s.repo.UpdateLastUsed(sa.ID); err != nil {
		return "", fmt.Errorf("update last used: %w", err)
	}

	token, err := s.generateAccessToken(sa)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}

func (s *Service) List(tenantID *string) ([]*ServiceAccount, error) {
	return s.repo.List(tenantID)
}

func (s *Service) GetByID(id string) (*ServiceAccount, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Update(id, name, description string, enabled bool) error {
	return s.repo.Update(id, name, description, enabled)
}

func (s *Service) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *Service) RegenerateSecret(id string) (string, error) {
	newSecret := generateClientSecret()
	hash, err := bcrypt.GenerateFromPassword([]byte(newSecret), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash client secret: %w", err)
	}

	if err := s.repo.RegenerateSecret(id, string(hash)); err != nil {
		return "", err
	}

	return newSecret, nil
}

func (s *Service) generateAccessToken(sa *ServiceAccount) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":           sa.ID,
		"identity_type": "service_account",
		"name":          sa.Name,
		"tenant_id":     sa.TenantID,
		"iat":           now.Unix(),
		"exp":           now.Add(s.cfg.AccessTokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

func generateClientID() string {
	return "sa_" + generateRandomString(20)
}

func generateClientSecret() string {
	return generateRandomString(40)
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.RawURLEncoding.EncodeToString(bytes)[:length]
}
