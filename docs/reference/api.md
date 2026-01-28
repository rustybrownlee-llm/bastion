# Bastion API Reference

**Version**: POC (0.1.0)
**Base URL**: `http://localhost:8080/api/v1`
**Last Updated**: 2025-01-28

---

## Overview

The Bastion API provides Zero Trust authentication services. All endpoints return JSON.

### Authentication

Protected endpoints require a Bearer token in the Authorization header:

```
Authorization: Bearer <access_token>
```

Access tokens are JWTs with a 15-minute TTL. Use the refresh endpoint to obtain new access tokens.

---

## Endpoints

### Health Check

#### GET /health

Returns server health status. No authentication required.

**Response (200)**
```json
{
  "status": "ok"
}
```

---

### User Management

#### POST /api/v1/users

Create a new user account. No authentication required (POC only).

**Request**
```json
{
  "email": "user@example.com",
  "password": "secret123"
}
```

**Response (201)**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "created_at": "2025-01-28T10:30:00Z"
}
```

**Errors**
| Status | Error | Description |
|--------|-------|-------------|
| 400 | invalid request | Malformed JSON body |
| 500 | failed to create user | Database error or duplicate email |

---

#### GET /api/v1/users/me

Get the currently authenticated user's profile.

**Headers**
```
Authorization: Bearer <access_token>
```

**Response (200)**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "created_at": "2025-01-28T10:30:00Z"
}
```

**Errors**
| Status | Error | Description |
|--------|-------|-------------|
| 401 | missing authorization header | No token provided |
| 401 | invalid token | Token malformed or expired |
| 404 | user not found | User no longer exists |

---

### Authentication

#### POST /api/v1/auth/login

Authenticate with email and password. Returns access and refresh tokens.

**Request**
```json
{
  "email": "user@example.com",
  "password": "secret123"
}
```

**Response (200)**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "abc123def456...",
  "expires_in": 900
}
```

| Field | Type | Description |
|-------|------|-------------|
| access_token | string | JWT for API authentication (15 min TTL) |
| refresh_token | string | Token for obtaining new access tokens (24h TTL) |
| expires_in | integer | Access token lifetime in seconds |

**Errors**
| Status | Error | Description |
|--------|-------|-------------|
| 400 | invalid request | Malformed JSON body |
| 401 | invalid credentials | Email not found or wrong password |

---

#### POST /api/v1/auth/refresh

Exchange a refresh token for a new access token. Updates session activity.

**Request**
```json
{
  "refresh_token": "abc123def456..."
}
```

**Response (200)**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900
}
```

**Errors**
| Status | Error | Description |
|--------|-------|-------------|
| 400 | invalid request | Malformed JSON body |
| 401 | invalid refresh token | Token invalid, expired, or session revoked |

---

#### POST /api/v1/auth/logout

Revoke the current session. Requires authentication.

**Headers**
```
Authorization: Bearer <access_token>
```

**Response (204)**

No content. Session is revoked and refresh token is invalidated.

**Errors**
| Status | Error | Description |
|--------|-------|-------------|
| 401 | missing authorization header | No token provided |
| 401 | invalid token | Token malformed or expired |
| 500 | logout failed | Database error |

---

## JWT Claims

Access tokens contain the following claims:

```json
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "iat": 1706443800,
  "exp": 1706444700
}
```

| Claim | Description |
|-------|-------------|
| sub | User UUID |
| email | User email address |
| iat | Issued at (Unix timestamp) |
| exp | Expiration (Unix timestamp) |

---

## Error Response Format

All errors return a consistent JSON structure:

```json
{
  "error": "error message here"
}
```

---

## Rate Limiting

Not implemented in POC. Production will include rate limiting on authentication endpoints.
