package tenant

import (
	"fmt"
	"regexp"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTenant(name, slug string) (*Tenant, error) {
	if name == "" {
		return nil, fmt.Errorf("tenant name required")
	}
	if slug == "" {
		return nil, fmt.Errorf("tenant slug required")
	}

	if !isValidSlug(slug) {
		return nil, fmt.Errorf("invalid slug: must be lowercase alphanumeric with hyphens")
	}

	return s.repo.Create(name, slug)
}

func (s *Service) GetTenant(id string) (*Tenant, error) {
	if id == "" {
		return nil, fmt.Errorf("tenant ID required")
	}
	return s.repo.GetByID(id)
}

func (s *Service) GetTenantBySlug(slug string) (*Tenant, error) {
	if slug == "" {
		return nil, fmt.Errorf("tenant slug required")
	}
	return s.repo.GetBySlug(slug)
}

func (s *Service) ListTenants() ([]*Tenant, error) {
	return s.repo.List()
}

func isValidSlug(slug string) bool {
	slug = strings.TrimSpace(slug)
	if len(slug) < 2 || len(slug) > 50 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-z0-9]+(?:-[a-z0-9]+)*$`, slug)
	return matched
}
