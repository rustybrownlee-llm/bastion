package auth

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rustybrownlee-llm/bastion/poc/internal/config"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db  *sql.DB
	cfg *config.AuthConfig
}

func NewService(db *sql.DB, cfg *config.AuthConfig) *Service {
	return &Service{db: db, cfg: cfg}
}

func (s *Service) Login(email, password string) (string, string, error) {
	var userID, passwordHash string
	var tenantID *string
	err := s.db.QueryRow(
		"SELECT id, password_hash, tenant_id FROM users WHERE email = $1",
		email,
	).Scan(&userID, &passwordHash, &tenantID)

	if err == sql.ErrNoRows {
		return "", "", fmt.Errorf("invalid credentials")
	}
	if err != nil {
		return "", "", fmt.Errorf("query user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return "", "", fmt.Errorf("invalid credentials")
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	refreshTokenHash, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("hash refresh token: %w", err)
	}

	expiresAt := time.Now().Add(s.cfg.RefreshTokenTTL)
	_, err = s.db.Exec(
		`INSERT INTO sessions (user_id, refresh_token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, string(refreshTokenHash), expiresAt,
	)
	if err != nil {
		return "", "", fmt.Errorf("create session: %w", err)
	}

	accessToken, err := GenerateAccessToken(s.cfg, userID, email, tenantID)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *Service) Refresh(refreshToken string) (string, error) {
	rows, err := s.db.Query(
		`SELECT s.id, s.user_id, s.refresh_token_hash, u.email, u.tenant_id
		 FROM sessions s
		 JOIN users u ON u.id = s.user_id
		 WHERE s.revoked = FALSE AND s.expires_at > NOW()`,
	)
	if err != nil {
		return "", fmt.Errorf("query sessions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sessionID, userID, hash, email string
		var tenantID *string
		if err := rows.Scan(&sessionID, &userID, &hash, &email, &tenantID); err != nil {
			continue
		}

		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(refreshToken)) == nil {
			_, err := s.db.Exec(
				"UPDATE sessions SET last_activity = NOW() WHERE id = $1",
				sessionID,
			)
			if err != nil {
				return "", fmt.Errorf("update session: %w", err)
			}

			accessToken, err := GenerateAccessToken(s.cfg, userID, email, tenantID)
			if err != nil {
				return "", fmt.Errorf("generate access token: %w", err)
			}

			return accessToken, nil
		}
	}

	return "", fmt.Errorf("invalid refresh token")
}

func (s *Service) Logout(userID string) error {
	_, err := s.db.Exec(
		"UPDATE sessions SET revoked = TRUE WHERE user_id = $1 AND revoked = FALSE",
		userID,
	)
	if err != nil {
		return fmt.Errorf("revoke sessions: %w", err)
	}
	return nil
}
