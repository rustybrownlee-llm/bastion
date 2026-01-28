# Bastion Database Schema Reference

**Version**: POC (0.1.0)
**Database**: PostgreSQL 15+
**Last Updated**: 2025-01-28

---

## Overview

The Bastion POC uses three core tables for authentication and auditing.

---

## Tables

### users

Stores user accounts and credentials.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK, auto-generated | Unique user identifier |
| email | VARCHAR(255) | UNIQUE, NOT NULL | User email (login identifier) |
| password_hash | TEXT | NOT NULL | bcrypt hash (cost 10) |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | Account creation time |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | Last modification time |

**Indexes**
- `idx_users_email` - Fast email lookups for login

---

### sessions

Tracks active user sessions and refresh tokens.

```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_activity TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_revoked ON sessions(revoked) WHERE NOT revoked;
```

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK, auto-generated | Session identifier |
| user_id | UUID | FK -> users.id, CASCADE | Owning user |
| refresh_token_hash | TEXT | NOT NULL | SHA-256 hash of refresh token |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | Session start time |
| last_activity | TIMESTAMP | NOT NULL, DEFAULT NOW() | Last token refresh time |
| expires_at | TIMESTAMP | NOT NULL | Absolute session expiration |
| revoked | BOOLEAN | NOT NULL, DEFAULT FALSE | True if logged out |

**Indexes**
- `idx_sessions_user_id` - Find sessions by user
- `idx_sessions_revoked` - Partial index for active sessions only

**Session Lifecycle**
1. Created on login with 24-hour expiration
2. `last_activity` updated on each token refresh
3. `revoked` set to TRUE on logout
4. Expired sessions remain for audit (future: cleanup job)

---

### audit_log

Records authentication events for security auditing.

```sql
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    details JSONB,
    ip_address VARCHAR(45),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_event_type ON audit_log(event_type);
CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);
```

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK, auto-generated | Event identifier |
| event_type | VARCHAR(100) | NOT NULL | Event category |
| user_id | UUID | FK -> users.id, SET NULL | Associated user (nullable) |
| details | JSONB | nullable | Event-specific metadata |
| ip_address | VARCHAR(45) | nullable | Client IP (IPv4 or IPv6) |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | Event timestamp |

**Event Types**
| Event | Description | Details |
|-------|-------------|---------|
| user_created | New user registered | email |
| user_creation_failure | Registration failed | email, error |
| login_success | User authenticated | email |
| login_failure | Authentication failed | email, error |
| token_refresh | Access token refreshed | - |
| token_refresh_failure | Refresh failed | error |
| logout | User logged out | - |

**Indexes**
- `idx_audit_log_event_type` - Filter by event type
- `idx_audit_log_user_id` - Find events for a user
- `idx_audit_log_created_at` - Time-based queries

---

## Entity Relationship Diagram

```
+------------+       +------------+       +------------+
|   users    |       |  sessions  |       | audit_log  |
+------------+       +------------+       +------------+
| id (PK)    |<------| user_id    |   +-->| user_id    |
| email      |   1:N | id (PK)    |   |   | id (PK)    |
| password   |       | token_hash |   |   | event_type |
| created_at |       | expires_at |   |   | details    |
| updated_at |       | revoked    |   |   | ip_address |
+------------+       +------------+   |   | created_at |
      |                               |   +------------+
      +-------------------------------+
                    1:N (nullable)
```

---

## Migrations

Migrations are stored in `poc/migrations/` and applied manually for POC:

```bash
psql postgresql://bastion:bastion_dev@localhost/bastion_poc \
  < poc/migrations/001_initial_schema.sql
```

| Migration | Description |
|-----------|-------------|
| 001_initial_schema.sql | Creates users, sessions, audit_log tables |

---

## Connection Details (POC Defaults)

| Setting | Value |
|---------|-------|
| Host | localhost |
| Port | 5432 |
| Database | bastion_poc |
| User | bastion |
| Password | bastion_dev |
| SSL Mode | disable |

Connection string:
```
postgresql://bastion:bastion_dev@localhost:5432/bastion_poc?sslmode=disable
```
