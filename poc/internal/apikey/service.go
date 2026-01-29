package apikey

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(name, description string, tenantID *string, expiresAt *time.Time, permissionIDs []string) (*APIKey, string, error) {
	prefix, _, fullKey := generateAPIKey()

	hash, err := bcrypt.GenerateFromPassword([]byte(fullKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash api key: %w", err)
	}

	key, err := s.repo.Create(name, description, prefix, string(hash), tenantID, expiresAt)
	if err != nil {
		return nil, "", err
	}

	for _, permID := range permissionIDs {
		if err := s.repo.AddPermission(key.ID, permID); err != nil {
			return nil, "", fmt.Errorf("add permission: %w", err)
		}
	}

	return key, fullKey, nil
}

func (s *Service) List(tenantID *string) ([]*APIKey, error) {
	return s.repo.List(tenantID)
}

func (s *Service) GetByID(id string) (*APIKey, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *Service) Authenticate(fullKey string) (*APIKey, error) {
	prefix, err := extractPrefix(fullKey)
	if err != nil {
		return nil, fmt.Errorf("invalid api key format")
	}

	key, err := s.repo.GetByPrefix(prefix)
	if err != nil {
		return nil, fmt.Errorf("api key not found")
	}

	if !key.Enabled {
		return nil, fmt.Errorf("api key disabled")
	}

	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("api key expired")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(key.KeyHash), []byte(fullKey)); err != nil {
		return nil, fmt.Errorf("invalid api key")
	}

	if err := s.repo.UpdateLastUsed(key.ID); err != nil {
		return nil, fmt.Errorf("update last used: %w", err)
	}

	return key, nil
}

func (s *Service) CheckPermission(apiKeyID, resourceType, action string) (bool, error) {
	permissions, err := s.repo.GetPermissions(apiKeyID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm.ResourceType == resourceType && perm.Action == action {
			return true, nil
		}
	}

	return false, nil
}

func (s *Service) GetPermissions(apiKeyID string) ([]*Permission, error) {
	return s.repo.GetPermissions(apiKeyID)
}

func (s *Service) AddPermission(apiKeyID, permissionID string) error {
	return s.repo.AddPermission(apiKeyID, permissionID)
}

func (s *Service) RemovePermission(apiKeyID, permissionID string) error {
	return s.repo.RemovePermission(apiKeyID, permissionID)
}

func generateAPIKey() (prefix, secret, fullKey string) {
	prefixBytes := make([]byte, 6)
	rand.Read(prefixBytes)
	prefix = "bst_" + base64.RawURLEncoding.EncodeToString(prefixBytes)[:8]

	secretBytes := make([]byte, 24)
	rand.Read(secretBytes)
	secret = base64.RawURLEncoding.EncodeToString(secretBytes)[:32]

	fullKey = prefix + "." + secret
	return prefix, secret, fullKey
}

func extractPrefix(fullKey string) (string, error) {
	if len(fullKey) < 13 {
		return "", fmt.Errorf("key too short")
	}

	dotIndex := -1
	for i := 0; i < len(fullKey); i++ {
		if fullKey[i] == '.' {
			dotIndex = i
			break
		}
	}

	if dotIndex == -1 {
		return "", fmt.Errorf("invalid key format")
	}

	return fullKey[:dotIndex], nil
}
