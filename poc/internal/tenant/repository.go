package tenant

import (
	"database/sql"
	"fmt"
	"time"
)

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Settings  string    `json:"settings,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(name, slug string) (*Tenant, error) {
	var tenant Tenant
	err := r.db.QueryRow(
		`INSERT INTO tenants (name, slug)
		 VALUES ($1, $2)
		 RETURNING id, name, slug, created_at, updated_at`,
		name, slug,
	).Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.CreatedAt, &tenant.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("insert tenant: %w", err)
	}

	return &tenant, nil
}

func (r *Repository) GetByID(id string) (*Tenant, error) {
	var tenant Tenant
	err := r.db.QueryRow(
		`SELECT id, name, slug, created_at, updated_at
		 FROM tenants WHERE id = $1`,
		id,
	).Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.CreatedAt, &tenant.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query tenant: %w", err)
	}

	return &tenant, nil
}

func (r *Repository) GetBySlug(slug string) (*Tenant, error) {
	var tenant Tenant
	err := r.db.QueryRow(
		`SELECT id, name, slug, created_at, updated_at
		 FROM tenants WHERE slug = $1`,
		slug,
	).Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.CreatedAt, &tenant.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query tenant: %w", err)
	}

	return &tenant, nil
}

func (r *Repository) List() ([]*Tenant, error) {
	rows, err := r.db.Query(
		`SELECT id, name, slug, created_at, updated_at
		 FROM tenants ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*Tenant
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan tenant: %w", err)
		}
		tenants = append(tenants, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tenants: %w", err)
	}

	return tenants, nil
}
