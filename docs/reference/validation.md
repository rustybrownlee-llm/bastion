# Bastion Validation Guide

**Version**: POC (0.1.0)
**Last Updated**: 2025-01-28

---

## Overview

This guide provides step-by-step instructions for validating the Bastion POC implementation against the success criteria defined in POC-002.0.

---

## Prerequisites

- Go 1.24+
- PostgreSQL 15+ (via Docker or local installation)
- curl (for API testing)
- jq (optional, for JSON formatting)

---

## Setup

### 1. Start PostgreSQL

**Using Docker:**
```bash
docker run -d --name bastion-db \
  -e POSTGRES_USER=bastion \
  -e POSTGRES_PASSWORD=bastion_dev \
  -e POSTGRES_DB=bastion_poc \
  -p 5432:5432 \
  postgres:15
```

**Using local PostgreSQL:**
```bash
createdb bastion_poc
createuser bastion
psql -c "ALTER USER bastion WITH PASSWORD 'bastion_dev';"
psql -c "GRANT ALL PRIVILEGES ON DATABASE bastion_poc TO bastion;"
```

### 2. Apply Migrations

```bash
psql postgresql://bastion:bastion_dev@localhost/bastion_poc \
  < poc/migrations/001_initial_schema.sql
```

### 3. Start the Server

```bash
cd poc
go run ./cmd/bastion -config config.yaml
```

Expected output:
```
Starting Bastion POC server on :8080
```

---

## Validation Tests

### Test 1: Health Check

**Criteria**: Server is running and responds to health checks.

```bash
curl -s http://localhost:8080/health | jq .
```

**Expected Response:**
```json
{
  "status": "ok"
}
```

---

### Test 2: User Creation

**Criteria**: User can be created via API.

```bash
curl -s -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' | jq .
```

**Expected Response (201):**
```json
{
  "id": "<uuid>",
  "email": "test@example.com",
  "created_at": "<timestamp>"
}
```

---

### Test 3: Login

**Criteria**: User can login and receive access + refresh tokens.

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' | jq .
```

**Expected Response (200):**
```json
{
  "access_token": "eyJhbG...",
  "refresh_token": "<random-string>",
  "expires_in": 900
}
```

Save the tokens for subsequent tests:
```bash
export ACCESS_TOKEN="<access_token_from_response>"
export REFRESH_TOKEN="<refresh_token_from_response>"
```

---

### Test 4: Token Validation (Get Current User)

**Criteria**: Access token validates correctly in middleware.

```bash
curl -s http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Expected Response (200):**
```json
{
  "id": "<uuid>",
  "email": "test@example.com",
  "created_at": "<timestamp>"
}
```

---

### Test 5: Invalid Token Rejected

**Criteria**: Invalid credentials return 401.

```bash
curl -s http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer invalid-token" | jq .
```

**Expected Response (401):**
```json
{
  "error": "invalid token"
}
```

---

### Test 6: Token Refresh

**Criteria**: Refresh token issues new access token.

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}" | jq .
```

**Expected Response (200):**
```json
{
  "access_token": "eyJhbG...",
  "expires_in": 900
}
```

Update the access token:
```bash
export ACCESS_TOKEN="<new_access_token_from_response>"
```

---

### Test 7: Logout

**Criteria**: Logout revokes session.

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN" -w "\nHTTP Status: %{http_code}\n"
```

**Expected Response:**
```
HTTP Status: 204
```

---

### Test 8: Refresh After Logout Fails

**Criteria**: Subsequent refresh with revoked session fails.

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}" | jq .
```

**Expected Response (401):**
```json
{
  "error": "invalid refresh token"
}
```

---

### Test 9: Invalid Login Credentials

**Criteria**: Invalid credentials return 401.

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"wrongpassword"}' | jq .
```

**Expected Response (401):**
```json
{
  "error": "invalid credentials"
}
```

---

### Test 10: Verify Audit Log

**Criteria**: Audit log contains login, refresh, logout events.

```bash
psql postgresql://bastion:bastion_dev@localhost/bastion_poc \
  -c "SELECT event_type, user_id, details, ip_address, created_at FROM audit_log ORDER BY created_at;"
```

**Expected Events:**
- user_created
- login_success
- token_refresh
- logout
- login_failure (if tested with wrong password)

---

## Success Criteria Checklist

| # | Criteria | Test | Status |
|---|----------|------|--------|
| 1 | User can be created via API | Test 2 | |
| 2 | User can login and receive access + refresh tokens | Test 3 | |
| 3 | Access token validates correctly in middleware | Test 4 | |
| 4 | Refresh token issues new access token | Test 6 | |
| 5 | Logout revokes session | Test 7 | |
| 6 | Subsequent refresh with revoked session fails | Test 8 | |
| 7 | Audit log contains login, refresh, logout events | Test 10 | |
| 8 | Invalid credentials return 401 | Test 5, 9 | |

---

## Cleanup

### Stop the Server

Press `Ctrl+C` in the terminal running the server.

### Stop and Remove Docker Container

```bash
docker stop bastion-db
docker rm bastion-db
```

### Reset Database (if using local PostgreSQL)

```bash
dropdb bastion_poc
createdb bastion_poc
psql postgresql://bastion:bastion_dev@localhost/bastion_poc \
  < poc/migrations/001_initial_schema.sql
```

---

## Troubleshooting

### Connection Refused

Ensure PostgreSQL is running:
```bash
docker ps | grep bastion-db
```

Or check local PostgreSQL:
```bash
pg_isready -h localhost -p 5432
```

### Authentication Failed

Check database credentials in `config.yaml` match your PostgreSQL setup.

### Token Validation Fails

Ensure you're using the most recent access token. Tokens expire after 15 minutes.

### Missing Tables

Re-run migrations:
```bash
psql postgresql://bastion:bastion_dev@localhost/bastion_poc \
  < poc/migrations/001_initial_schema.sql
```
