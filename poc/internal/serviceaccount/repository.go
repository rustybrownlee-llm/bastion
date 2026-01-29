package serviceaccount

import (
	"database/sql"
	"fmt"
	"time"
)

type ServiceAccount struct {
	ID                 string
	Name               string
	Description        string
	ClientID           string
	ClientSecretHash   string
	TenantID           *string
	Enabled            bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	LastUsedAt         *time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(name, description, clientID, clientSecretHash string, tenantID *string) (*ServiceAccount, error) {
	sa := &ServiceAccount{}
	err := r.db.QueryRow(
		`INSERT INTO service_accounts (name, description, client_id, client_secret_hash, tenant_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, name, description, client_id, client_secret_hash, tenant_id, enabled, created_at, updated_at, last_used_at`,
		name, description, clientID, clientSecretHash, tenantID,
	).Scan(&sa.ID, &sa.Name, &sa.Description, &sa.ClientID, &sa.ClientSecretHash, &sa.TenantID, &sa.Enabled, &sa.CreatedAt, &sa.UpdatedAt, &sa.LastUsedAt)

	if err != nil {
		return nil, fmt.Errorf("create service account: %w", err)
	}

	return sa, nil
}

func (r *Repository) GetByID(id string) (*ServiceAccount, error) {
	sa := &ServiceAccount{}
	err := r.db.QueryRow(
		`SELECT id, name, description, client_id, client_secret_hash, tenant_id, enabled, created_at, updated_at, last_used_at
		 FROM service_accounts WHERE id = $1`,
		id,
	).Scan(&sa.ID, &sa.Name, &sa.Description, &sa.ClientID, &sa.ClientSecretHash, &sa.TenantID, &sa.Enabled, &sa.CreatedAt, &sa.UpdatedAt, &sa.LastUsedAt)

	if err != nil {
		return nil, fmt.Errorf("get service account: %w", err)
	}

	return sa, nil
}

func (r *Repository) GetByClientID(clientID string) (*ServiceAccount, error) {
	sa := &ServiceAccount{}
	err := r.db.QueryRow(
		`SELECT id, name, description, client_id, client_secret_hash, tenant_id, enabled, created_at, updated_at, last_used_at
		 FROM service_accounts WHERE client_id = $1`,
		clientID,
	).Scan(&sa.ID, &sa.Name, &sa.Description, &sa.ClientID, &sa.ClientSecretHash, &sa.TenantID, &sa.Enabled, &sa.CreatedAt, &sa.UpdatedAt, &sa.LastUsedAt)

	if err != nil {
		return nil, fmt.Errorf("get service account by client_id: %w", err)
	}

	return sa, nil
}

func (r *Repository) List(tenantID *string) ([]*ServiceAccount, error) {
	var rows *sql.Rows
	var err error

	if tenantID == nil {
		rows, err = r.db.Query(
			`SELECT id, name, description, client_id, client_secret_hash, tenant_id, enabled, created_at, updated_at, last_used_at
			 FROM service_accounts WHERE tenant_id IS NULL`,
		)
	} else {
		rows, err = r.db.Query(
			`SELECT id, name, description, client_id, client_secret_hash, tenant_id, enabled, created_at, updated_at, last_used_at
			 FROM service_accounts WHERE tenant_id = $1 OR tenant_id IS NULL`,
			*tenantID,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("list service accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*ServiceAccount
	for rows.Next() {
		sa := &ServiceAccount{}
		if err := rows.Scan(&sa.ID, &sa.Name, &sa.Description, &sa.ClientID, &sa.ClientSecretHash, &sa.TenantID, &sa.Enabled, &sa.CreatedAt, &sa.UpdatedAt, &sa.LastUsedAt); err != nil {
			return nil, fmt.Errorf("scan service account: %w", err)
		}
		accounts = append(accounts, sa)
	}

	return accounts, nil
}

func (r *Repository) Update(id, name, description string, enabled bool) error {
	_, err := r.db.Exec(
		`UPDATE service_accounts SET name = $1, description = $2, enabled = $3, updated_at = NOW()
		 WHERE id = $4`,
		name, description, enabled, id,
	)
	return err
}

func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM service_accounts WHERE id = $1`, id)
	return err
}

func (r *Repository) UpdateLastUsed(id string) error {
	_, err := r.db.Exec(
		`UPDATE service_accounts SET last_used_at = NOW() WHERE id = $1`,
		id,
	)
	return err
}

func (r *Repository) RegenerateSecret(id, newSecretHash string) error {
	_, err := r.db.Exec(
		`UPDATE service_accounts SET client_secret_hash = $1, updated_at = NOW()
		 WHERE id = $2`,
		newSecretHash, id,
	)
	return err
}

func (r *Repository) AssignRole(serviceAccountID, roleID string) error {
	_, err := r.db.Exec(
		`INSERT INTO service_account_roles (service_account_id, role_id)
		 VALUES ($1, $2)
		 ON CONFLICT (service_account_id, role_id) DO NOTHING`,
		serviceAccountID, roleID,
	)
	return err
}

func (r *Repository) GetRoles(serviceAccountID string) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT role_id FROM service_account_roles WHERE service_account_id = $1`,
		serviceAccountID,
	)
	if err != nil {
		return nil, fmt.Errorf("get service account roles: %w", err)
	}
	defer rows.Close()

	var roleIDs []string
	for rows.Next() {
		var roleID string
		if err := rows.Scan(&roleID); err != nil {
			return nil, fmt.Errorf("scan role_id: %w", err)
		}
		roleIDs = append(roleIDs, roleID)
	}

	return roleIDs, nil
}
