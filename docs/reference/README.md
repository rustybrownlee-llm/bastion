# Bastion Reference Documentation

This directory contains up-to-date reference documentation for the Bastion management plane.

---

## Documents

| Document | Description |
|----------|-------------|
| [API Reference](api.md) | HTTP endpoints, request/response formats, authentication |
| [Database Schema](database-schema.md) | Tables, columns, indexes, relationships |
| [Configuration](configuration.md) | YAML configuration options and defaults |
| [Validation Guide](validation.md) | Step-by-step testing instructions |

---

## Current Implementation Status

### POC Phase

| SOW | Description | Status |
|-----|-------------|--------|
| POC-001.0 | Basic Go structure | Complete |
| POC-002.0 | Core authentication | Complete |
| POC-003 | Basic RBAC | Not started |
| POC-004 | Service accounts / API keys | Not started |

### Implemented Features (POC-002.0)

- User registration (email/password)
- Login with JWT access tokens (15 min TTL)
- Refresh tokens for session continuity (24h TTL)
- Token validation middleware
- Session management with revocation
- Audit logging for all auth events

### Not Yet Implemented

- Multi-factor authentication (MFA)
- Single sign-on (SSO) / Federation
- Service accounts
- API keys
- Role-Based Access Control (RBAC)
- Tenant isolation
- Password reset
- Email verification
- Rate limiting

---

## Quick Start

```bash
# 1. Start PostgreSQL
docker run -d --name bastion-db \
  -e POSTGRES_USER=bastion \
  -e POSTGRES_PASSWORD=bastion_dev \
  -e POSTGRES_DB=bastion_poc \
  -p 5432:5432 \
  postgres:15

# 2. Apply migrations
psql postgresql://bastion:bastion_dev@localhost/bastion_poc \
  < poc/migrations/001_initial_schema.sql

# 3. Run the server
cd poc && go run ./cmd/bastion -config config.yaml

# 4. Test
curl http://localhost:8080/health
```

See [Validation Guide](validation.md) for complete testing instructions.

---

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| github.com/go-chi/chi/v5 | v5.2.4 | HTTP routing |
| github.com/golang-jwt/jwt/v5 | v5.3.1 | JWT tokens |
| github.com/lib/pq | v1.10.9 | PostgreSQL driver |
| golang.org/x/crypto | v0.47.0 | bcrypt password hashing |
| gopkg.in/yaml.v3 | v3.0.1 | Configuration parsing |

---

## Keeping Documentation Updated

When implementing new SOWs:

1. Update [API Reference](api.md) with any new endpoints
2. Update [Database Schema](database-schema.md) with schema changes
3. Update [Configuration](configuration.md) with new config options
4. Update [Validation Guide](validation.md) with new test cases
5. Update this README with implementation status
