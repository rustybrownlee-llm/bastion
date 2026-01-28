package user

import (
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	TenantID     *string   `json:"tenant_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(email, passwordHash string, tenantID *string) (*User, error) {
	var user User
	err := r.db.QueryRow(
		`INSERT INTO users (email, password_hash, tenant_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, tenant_id, created_at, updated_at`,
		email, passwordHash, tenantID,
	).Scan(&user.ID, &user.Email, &user.TenantID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return &user, nil
}

func (r *Repository) GetByEmail(email string) (*User, error) {
	var user User
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, tenant_id, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.TenantID, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}

	return &user, nil
}

func (r *Repository) GetByID(id string) (*User, error) {
	var user User
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, tenant_id, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.TenantID, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}

	return &user, nil
}

func (r *Repository) SetTenantID(userID string, tenantID *string) error {
	_, err := r.db.Exec(
		`UPDATE users SET tenant_id = $1, updated_at = NOW() WHERE id = $2`,
		tenantID, userID,
	)
	if err != nil {
		return fmt.Errorf("update user tenant: %w", err)
	}
	return nil
}
