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
| POC-003.0 | Basic RBAC | Complete |
| POC-004.0 | Service accounts / API keys | Complete |

### Implemented Features

**Authentication (POC-002.0)**:
- User registration (email/password)
- Login with JWT access tokens (15 min TTL)
- Refresh tokens for session continuity (24h TTL)
- Token validation middleware
- Session management with revocation
- Audit logging for all auth events

**RBAC (POC-003.0)**:
- Multi-tenancy with tenant isolation
- Platform roles (superadmin, admin, auditor)
- Application roles (tenant-admin, user-admin, viewer)
- Permission-based authorization
- Role assignment per tenant
- Authorization check endpoint
- RequirePermission middleware
- JWT includes tenant context

**Non-Human Identities (POC-004.0)**:
- Service accounts with OAuth2 client credentials
- API keys with X-API-Key header auth
- Service accounts use role-based RBAC
- API keys use direct permission grants
- Secret regeneration for service accounts
- Optional expiration for API keys

### Not Yet Implemented

- Multi-factor authentication (MFA)
- Single sign-on (SSO) / Federation
- Role inheritance
- Permission conditions (time, IP)
- Password reset
- Email verification
- Rate limiting
- API key IP restrictions

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
docker exec -i bastion-db psql -U bastion -d bastion_poc \
  < poc/migrations/001_initial_schema.sql
docker exec -i bastion-db psql -U bastion -d bastion_poc \
  < poc/migrations/002_rbac_schema.sql
docker exec -i bastion-db psql -U bastion -d bastion_poc \
  < poc/migrations/003_service_accounts_api_keys.sql

# 3. Run the server (port 8081)
cd poc && go run ./cmd/bastion -config config.yaml

# 4. Test
curl http://localhost:8081/health
```

**Default Admin**: `admin@bastion.local` / `BastionAdmin2025` (platform:superadmin)

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
