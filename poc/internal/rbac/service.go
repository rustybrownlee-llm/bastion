package rbac

import (
	"fmt"

	"github.com/rustybrownlee-llm/bastion/poc/internal/audit"
)

type Service struct {
	repo        *Repository
	auditLogger *audit.Logger
}

func NewService(repo *Repository, auditLogger *audit.Logger) *Service {
	return &Service{
		repo:        repo,
		auditLogger: auditLogger,
	}
}

func (s *Service) AssignRole(userID, roleName string, tenantID *string, grantedBy *string) error {
	if userID == "" {
		return fmt.Errorf("user ID required")
	}
	if roleName == "" {
		return fmt.Errorf("role name required")
	}

	role, err := s.repo.GetRoleByName(roleName)
	if err != nil {
		return fmt.Errorf("get role: %w", err)
	}

	if err := s.repo.AssignRoleToUser(userID, role.ID, tenantID, grantedBy); err != nil {
		return fmt.Errorf("assign role: %w", err)
	}

	var grantedByStr string
	if grantedBy != nil {
		grantedByStr = *grantedBy
	}

	s.auditLogger.Log("role.assigned", grantedByStr, map[string]interface{}{
		"user_id":   userID,
		"role_name": roleName,
		"role_id":   role.ID,
		"tenant_id": tenantID,
	}, "")

	return nil
}

func (s *Service) RevokeRole(userID, roleName string, tenantID *string) error {
	if userID == "" {
		return fmt.Errorf("user ID required")
	}
	if roleName == "" {
		return fmt.Errorf("role name required")
	}

	role, err := s.repo.GetRoleByName(roleName)
	if err != nil {
		return fmt.Errorf("get role: %w", err)
	}

	if err := s.repo.RevokeRoleFromUser(userID, role.ID, tenantID); err != nil {
		return fmt.Errorf("revoke role: %w", err)
	}

	s.auditLogger.Log("role.revoked", "", map[string]interface{}{
		"user_id":   userID,
		"role_name": roleName,
		"role_id":   role.ID,
		"tenant_id": tenantID,
	}, "")

	return nil
}

func (s *Service) CheckPermission(userID string, tenantID *string, resourceType, action string) (bool, string, error) {
	if userID == "" {
		return false, "user ID required", fmt.Errorf("user ID required")
	}
	if resourceType == "" {
		return false, "resource type required", fmt.Errorf("resource type required")
	}
	if action == "" {
		return false, "action required", fmt.Errorf("action required")
	}

	allowed, err := s.repo.HasPermission(userID, tenantID, resourceType, action)
	if err != nil {
		return false, "permission check failed", err
	}

	var reason string
	if allowed {
		reason = fmt.Sprintf("user has %s:%s permission", resourceType, action)
	} else {
		reason = fmt.Sprintf("user lacks %s:%s permission", resourceType, action)
	}

	s.auditLogger.Log("authz.check", userID, map[string]interface{}{
		"resource_type": resourceType,
		"action":        action,
		"tenant_id":     tenantID,
		"allowed":       allowed,
		"reason":        reason,
	}, "")

	return allowed, reason, nil
}

func (s *Service) GetUserPermissions(userID string, tenantID *string) ([]*Permission, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID required")
	}

	return s.repo.GetUserPermissions(userID, tenantID)
}

func (s *Service) GetUserRoles(userID string, tenantID *string) ([]*UserRole, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID required")
	}

	return s.repo.GetUserRoles(userID, tenantID)
}
