# Bastion Configuration Reference

**Version**: POC (0.1.0)
**Last Updated**: 2025-01-28

---

## Overview

Bastion uses YAML configuration files. The default configuration file is `config.yaml` in the POC directory.

---

## Configuration File

### Location

```
poc/config.yaml
```

### Command Line

```bash
go run ./cmd/bastion -config config.yaml
```

If no `-config` flag is provided, the server will fail to start.

---

## Configuration Schema

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

---

## Configuration Options

### Server

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| server.port | integer | 8080 | HTTP server listen port |

### Database

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| database.host | string | localhost | PostgreSQL host |
| database.port | integer | 5432 | PostgreSQL port |
| database.name | string | bastion_poc | Database name |
| database.user | string | bastion | Database user |
| database.password | string | bastion_dev | Database password |
| database.sslmode | string | disable | SSL mode (disable, require, verify-ca, verify-full) |

### Auth

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| auth.access_token_ttl | duration | 15m | Access token lifetime |
| auth.refresh_token_ttl | duration | 24h | Refresh token/session lifetime |
| auth.jwt_secret | string | (required) | HMAC signing key for JWTs |

---

## Duration Format

Duration values use Go's time.ParseDuration format:

| Unit | Suffix | Example |
|------|--------|---------|
| Nanoseconds | ns | 500ns |
| Microseconds | us | 100us |
| Milliseconds | ms | 250ms |
| Seconds | s | 30s |
| Minutes | m | 15m |
| Hours | h | 24h |

Combinations are supported: `1h30m`, `2h45m30s`

---

## Environment-Specific Configurations

### Development (POC)

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

### Production (Future)

Production configuration will include:
- Secrets management integration (not hardcoded passwords)
- SSL/TLS for database connections
- Shorter token TTLs based on security requirements
- Connection pooling settings

---

## Security Notes

### JWT Secret

The `jwt_secret` value is critical for token security:
- Must be at least 32 characters for HMAC-SHA256
- Must be kept secret and never committed to version control
- Changing the secret invalidates all existing tokens

**POC Only**: The secret is hardcoded in config.yaml for convenience. Production will use secrets management.

### Database Password

**POC Only**: Plaintext password in config file. Production will use:
- Environment variables
- Secrets management (HashiCorp Vault, AWS Secrets Manager, etc.)
- Certificate-based authentication
