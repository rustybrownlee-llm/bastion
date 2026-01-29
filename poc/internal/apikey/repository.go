package apikey

import (
	"database/sql"
	"fmt"
	"time"
)

type APIKey struct {
	ID          string
	Name        string
	Description string
	KeyPrefix   string
	KeyHash     string
	TenantID    *string
	ExpiresAt   *time.Time
	Enabled     bool
	CreatedAt   time.Time
	LastUsedAt  *time.Time
}

type Permission struct {
	ID           string
	ResourceType string
	Action       string
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(name, description, keyPrefix, keyHash string, tenantID *string, expiresAt *time.Time) (*APIKey, error) {
	key := &APIKey{}
	err := r.db.QueryRow(
		`INSERT INTO api_keys (name, description, key_prefix, key_hash, tenant_id, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, name, description, key_prefix, key_hash, tenant_id, expires_at, enabled, created_at, last_used_at`,
		name, description, keyPrefix, keyHash, tenantID, expiresAt,
	).Scan(&key.ID, &key.Name, &key.Description, &key.KeyPrefix, &key.KeyHash, &key.TenantID, &key.ExpiresAt, &key.Enabled, &key.CreatedAt, &key.LastUsedAt)

	if err != nil {
		return nil, fmt.Errorf("create api key: %w", err)
	}

	return key, nil
}

func (r *Repository) GetByID(id string) (*APIKey, error) {
	key := &APIKey{}
	err := r.db.QueryRow(
		`SELECT id, name, description, key_prefix, key_hash, tenant_id, expires_at, enabled, created_at, last_used_at
		 FROM api_keys WHERE id = $1`,
		id,
	).Scan(&key.ID, &key.Name, &key.Description, &key.KeyPrefix, &key.KeyHash, &key.TenantID, &key.ExpiresAt, &key.Enabled, &key.CreatedAt, &key.LastUsedAt)

	if err != nil {
		return nil, fmt.Errorf("get api key: %w", err)
	}

	return key, nil
}

func (r *Repository) GetByPrefix(prefix string) (*APIKey, error) {
	key := &APIKey{}
	err := r.db.QueryRow(
		`SELECT id, name, description, key_prefix, key_hash, tenant_id, expires_at, enabled, created_at, last_used_at
		 FROM api_keys WHERE key_prefix = $1`,
		prefix,
	).Scan(&key.ID, &key.Name, &key.Description, &key.KeyPrefix, &key.KeyHash, &key.TenantID, &key.ExpiresAt, &key.Enabled, &key.CreatedAt, &key.LastUsedAt)

	if err != nil {
		return nil, fmt.Errorf("get api key by prefix: %w", err)
	}

	return key, nil
}

func (r *Repository) List(tenantID *string) ([]*APIKey, error) {
	var rows *sql.Rows
	var err error

	if tenantID == nil {
		rows, err = r.db.Query(
			`SELECT id, name, description, key_prefix, key_hash, tenant_id, expires_at, enabled, created_at, last_used_at
			 FROM api_keys WHERE tenant_id IS NULL`,
		)
	} else {
		rows, err = r.db.Query(
			`SELECT id, name, description, key_prefix, key_hash, tenant_id, expires_at, enabled, created_at, last_used_at
			 FROM api_keys WHERE tenant_id = $1 OR tenant_id IS NULL`,
			*tenantID,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		key := &APIKey{}
		if err := rows.Scan(&key.ID, &key.Name, &key.Description, &key.KeyPrefix, &key.KeyHash, &key.TenantID, &key.ExpiresAt, &key.Enabled, &key.CreatedAt, &key.LastUsedAt); err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		keys = append(keys, key)
	}

	return keys, nil
}

func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM api_keys WHERE id = $1`, id)
	return err
}

func (r *Repository) UpdateLastUsed(id string) error {
	_, err := r.db.Exec(
		`UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`,
		id,
	)
	return err
}

func (r *Repository) AddPermission(apiKeyID, permissionID string) error {
	_, err := r.db.Exec(
		`INSERT INTO api_key_permissions (api_key_id, permission_id)
		 VALUES ($1, $2)
		 ON CONFLICT (api_key_id, permission_id) DO NOTHING`,
		apiKeyID, permissionID,
	)
	return err
}

func (r *Repository) RemovePermission(apiKeyID, permissionID string) error {
	_, err := r.db.Exec(
		`DELETE FROM api_key_permissions WHERE api_key_id = $1 AND permission_id = $2`,
		apiKeyID, permissionID,
	)
	return err
}

func (r *Repository) GetPermissions(apiKeyID string) ([]*Permission, error) {
	rows, err := r.db.Query(
		`SELECT p.id, p.resource_type, p.action
		 FROM permissions p
		 JOIN api_key_permissions akp ON p.id = akp.permission_id
		 WHERE akp.api_key_id = $1`,
		apiKeyID,
	)
	if err != nil {
		return nil, fmt.Errorf("get api key permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*Permission
	for rows.Next() {
		perm := &Permission{}
		if err := rows.Scan(&perm.ID, &perm.ResourceType, &perm.Action); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}
