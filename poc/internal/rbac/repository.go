package rbac

import (
	"database/sql"
	"fmt"
	"time"
)

type Role struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	RoleType        string    `json:"role_type"`
	ApplicationName string    `json:"application_name,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type Permission struct {
	ID           string `json:"id"`
	ResourceType string `json:"resource_type"`
	Action       string `json:"action"`
	Description  string `json:"description,omitempty"`
}

type UserRole struct {
	UserID    string    `json:"user_id"`
	RoleID    string    `json:"role_id"`
	RoleName  string    `json:"role_name"`
	TenantID  *string   `json:"tenant_id"`
	GrantedAt time.Time `json:"granted_at"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetRoleByID(id string) (*Role, error) {
	var role Role
	err := r.db.QueryRow(
		`SELECT id, name, description, role_type, COALESCE(application_name, ''), created_at
		 FROM roles WHERE id = $1`,
		id,
	).Scan(&role.ID, &role.Name, &role.Description, &role.RoleType, &role.ApplicationName, &role.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query role: %w", err)
	}

	return &role, nil
}

func (r *Repository) GetRoleByName(name string) (*Role, error) {
	var role Role
	err := r.db.QueryRow(
		`SELECT id, name, description, role_type, COALESCE(application_name, ''), created_at
		 FROM roles WHERE name = $1`,
		name,
	).Scan(&role.ID, &role.Name, &role.Description, &role.RoleType, &role.ApplicationName, &role.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query role: %w", err)
	}

	return &role, nil
}

func (r *Repository) ListRoles() ([]*Role, error) {
	rows, err := r.db.Query(
		`SELECT id, name, description, role_type, COALESCE(application_name, ''), created_at
		 FROM roles ORDER BY role_type, name`,
	)
	if err != nil {
		return nil, fmt.Errorf("query roles: %w", err)
	}
	defer rows.Close()

	var roles []*Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.RoleType, &role.ApplicationName, &role.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, &role)
	}

	return roles, rows.Err()
}

func (r *Repository) GetPermissionByID(id string) (*Permission, error) {
	var perm Permission
	err := r.db.QueryRow(
		`SELECT id, resource_type, action, COALESCE(description, '')
		 FROM permissions WHERE id = $1`,
		id,
	).Scan(&perm.ID, &perm.ResourceType, &perm.Action, &perm.Description)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("permission not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query permission: %w", err)
	}

	return &perm, nil
}

func (r *Repository) ListPermissions() ([]*Permission, error) {
	rows, err := r.db.Query(
		`SELECT id, resource_type, action, COALESCE(description, '')
		 FROM permissions ORDER BY resource_type, action`,
	)
	if err != nil {
		return nil, fmt.Errorf("query permissions: %w", err)
	}
	defer rows.Close()

	var perms []*Permission
	for rows.Next() {
		var perm Permission
		if err := rows.Scan(&perm.ID, &perm.ResourceType, &perm.Action, &perm.Description); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		perms = append(perms, &perm)
	}

	return perms, rows.Err()
}

func (r *Repository) GetRolePermissions(roleID string) ([]*Permission, error) {
	rows, err := r.db.Query(
		`SELECT p.id, p.resource_type, p.action, COALESCE(p.description, '')
		 FROM permissions p
		 JOIN role_permissions rp ON p.id = rp.permission_id
		 WHERE rp.role_id = $1
		 ORDER BY p.resource_type, p.action`,
		roleID,
	)
	if err != nil {
		return nil, fmt.Errorf("query role permissions: %w", err)
	}
	defer rows.Close()

	var perms []*Permission
	for rows.Next() {
		var perm Permission
		if err := rows.Scan(&perm.ID, &perm.ResourceType, &perm.Action, &perm.Description); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		perms = append(perms, &perm)
	}

	return perms, rows.Err()
}

func (r *Repository) AssignRoleToUser(userID, roleID string, tenantID *string, grantedBy *string) error {
	_, err := r.db.Exec(
		`INSERT INTO user_roles (user_id, role_id, tenant_id, granted_by)
		 VALUES ($1, $2, $3, $4)`,
		userID, roleID, tenantID, grantedBy,
	)
	if err != nil {
		return fmt.Errorf("assign role: %w", err)
	}
	return nil
}

func (r *Repository) GetUserRoles(userID string, tenantID *string) ([]*UserRole, error) {
	query := `
		SELECT ur.user_id, ur.role_id, r.name, ur.tenant_id, ur.granted_at
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND (ur.tenant_id = $2 OR ur.tenant_id IS NULL)
		ORDER BY r.role_type, r.name`

	rows, err := r.db.Query(query, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query user roles: %w", err)
	}
	defer rows.Close()

	var userRoles []*UserRole
	for rows.Next() {
		var ur UserRole
		if err := rows.Scan(&ur.UserID, &ur.RoleID, &ur.RoleName, &ur.TenantID, &ur.GrantedAt); err != nil {
			return nil, fmt.Errorf("scan user role: %w", err)
		}
		userRoles = append(userRoles, &ur)
	}

	return userRoles, rows.Err()
}

func (r *Repository) GetUserPermissions(userID string, tenantID *string) ([]*Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.resource_type, p.action, COALESCE(p.description, '')
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1 AND (ur.tenant_id = $2 OR ur.tenant_id IS NULL)
		ORDER BY p.resource_type, p.action`

	rows, err := r.db.Query(query, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query user permissions: %w", err)
	}
	defer rows.Close()

	var perms []*Permission
	for rows.Next() {
		var perm Permission
		if err := rows.Scan(&perm.ID, &perm.ResourceType, &perm.Action, &perm.Description); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		perms = append(perms, &perm)
	}

	return perms, rows.Err()
}

func (r *Repository) RevokeRoleFromUser(userID, roleID string, tenantID *string) error {
	_, err := r.db.Exec(
		`DELETE FROM user_roles
		 WHERE user_id = $1 AND role_id = $2 AND tenant_id = $3`,
		userID, roleID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("revoke role: %w", err)
	}
	return nil
}

func (r *Repository) HasPermission(userID string, tenantID *string, resourceType, action string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
		AND (ur.tenant_id = $2 OR ur.tenant_id IS NULL)
		AND p.resource_type = $3
		AND p.action = $4`

	err := r.db.QueryRow(query, userID, tenantID, resourceType, action).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check permission: %w", err)
	}

	return count > 0, nil
}
