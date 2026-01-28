# Bastion POC

**Status**: Proof of Concept (Throwaway Code)

This directory contains the Bastion proof-of-concept implementation. The code here validates architectural assumptions for Zero Trust authentication and RBAC before building the production platform.

## Purpose

- Prove Zero Trust auth and RBAC concepts work as designed
- Validate token lifecycle management
- Test RBAC permission enforcement
- Establish patterns for production implementation

## What This Is NOT

- **Not production code** - This will be archived after Phase 1
- **Not optimized** - Clarity over performance
- **Not complete** - Only implements what's needed to validate architecture

## Implemented POCs

- **POC-001.0**: Basic Go structure with HTTP server
- **POC-002.0**: Core authentication (users, login, tokens, sessions, audit)

## Setup

### 1. Start PostgreSQL

```bash
docker run -d --name bastion-db \
  -e POSTGRES_USER=bastion \
  -e POSTGRES_PASSWORD=bastion_dev \
  -e POSTGRES_DB=bastion_poc \
  -p 5432:5432 \
  postgres:15
```

### 2. Apply Migrations

```bash
psql postgresql://bastion:bastion_dev@localhost/bastion_poc \
  < migrations/001_initial_schema.sql
```

### 3. Build and Run

```bash
cd poc
go build ./cmd/bastion
./bastion -config config.yaml
```

Or run directly:

```bash
go run ./cmd/bastion -config config.yaml
```

## Testing Authentication (POC-002.0)

### Automated Test Script

```bash
./test-auth.sh
```

### Manual API Testing

#### 1. Create User

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'
```

#### 2. Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'
```

Response:
```json
{
  "access_token": "eyJhbG...",
  "refresh_token": "abc123...",
  "expires_in": 900
}
```

#### 3. Access Protected Endpoint

```bash
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <access_token>"
```

#### 4. Refresh Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

#### 5. Logout

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <access_token>"
```

## Endpoints

### Health Check

```bash
curl http://localhost:8080/health
```

### Authentication Endpoints

- `POST /api/v1/users` - Create user
- `POST /api/v1/auth/login` - Login (get tokens)
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout (requires auth)
- `GET /api/v1/users/me` - Get current user (requires auth)

## Configuration

Edit `config.yaml`:

```yaml
server:
  port: 8080

database:
  host: localhost
  port: 5432
  name: bastion_poc
  user: bastion
  password: bastion_dev
  sslmode: disable

auth:
  access_token_ttl: 15m
  refresh_token_ttl: 24h
  jwt_secret: change-me-in-production
```

## Architecture

### Token Flow

1. **Login**: Validate credentials, create session, return access + refresh tokens
2. **Refresh**: Validate refresh token, issue new access token, update activity
3. **Logout**: Revoke session

### Token Types

- **Access Token**: JWT, 15 minute TTL, used for API authentication
- **Refresh Token**: Random string, 24 hour TTL, stored hashed in database

### Database Schema

- `users`: User accounts with bcrypt password hashes
- `sessions`: Active sessions with refresh token hashes
- `audit_log`: All authentication events

## Shutdown

Send SIGINT (Ctrl+C) or SIGTERM for graceful shutdown.

## Related Documentation

- `/docs/decisions/DD-001-bastion-architecture.md` - Architecture decision document
- `/docs/sows/POC-001.0-basic-go-structure.md` - Basic structure SOW
- `/docs/sows/POC-002.0-core-authentication.md` - Core authentication SOW
- `/CLAUDE.md` - Project guidelines
