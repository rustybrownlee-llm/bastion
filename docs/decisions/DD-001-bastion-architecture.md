# DD-001: Bastion Management Plane Architecture

**Status**: Draft
**Created**: 2025-01-28
**Last Updated**: 2025-01-28
**Author**: Architecture Team

---

## 1. Overview

### 1.1 Purpose

Bastion is a reusable management plane providing Zero Trust authentication and configurable Role-Based Access Control (RBAC) for multiple applications. It serves as the central authority for identity, authentication, authorization, and audit across all consuming applications.

### 1.2 Problem Statement

Authentication and authorization are frequently treated as afterthoughts in application development, leading to:
- Inconsistent security models across applications
- Difficult-to-maintain permission logic scattered through codebases
- Painful retrofitting when security requirements increase
- Duplicate effort implementing auth in each application

Bastion solves this by providing a single, well-designed management plane that any application can consume.

### 1.3 Design Goals

1. **Zero Trust by Default**: Short-lived credentials, continuous verification, assume breach
2. **Application Agnostic**: Any application can register and consume Bastion
3. **Multi-Tenant**: Tenant isolation is foundational, not bolted on
4. **Configurable RBAC**: Tenants can define custom roles composing app-provided permissions
5. **Auditable**: Every authentication and authorization decision is logged
6. **Performant**: Authorization checks must be fast (caching, local evaluation where possible)

---

## 2. Core Concepts

### 2.1 Identity Model

Bastion manages three types of identities:

| Identity Type | Description | Authentication Method |
|---------------|-------------|----------------------|
| **User** | Human operator | Password + optional MFA, SSO |
| **Service Account** | Application or service | Client credentials, mTLS |
| **API Key** | External integration | Key + secret with scoping |

All identities exist within a tenant context (except platform-level service accounts).

### 2.2 Tenant Model

```
Platform (Bastion)
    │
    ├── Tenant: Acme Corp
    │   ├── Users
    │   ├── Service Accounts
    │   ├── API Keys
    │   ├── Custom Roles
    │   └── Tenant Settings (idle timeout, etc.)
    │
    ├── Tenant: Beta Inc
    │   └── ...
    │
    └── Platform Service Accounts (cross-tenant)
```

Tenants are isolated by default. Cross-tenant access requires explicit platform-level grants.

### 2.3 Authentication Model

#### Token Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        SESSION                               │
│  Absolute max: configurable (default 12 hours)              │
│  Idle timeout: configurable (default 30 minutes)            │
│                                                              │
│    ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐            │
│    │Token1│ │Token2│ │Token3│ │Token4│ │Token5│ ...        │
│    │ 15m  │ │ 15m  │ │ 15m  │ │ 15m  │ │ 15m  │            │
│    └──────┘ └──────┘ └──────┘ └──────┘ └──────┘            │
│         └───────┴───────┴───────┴───────┴── Silent refresh  │
└─────────────────────────────────────────────────────────────┘
```

**Key Principles**:
- Access tokens: Short-lived (10-15 minutes)
- Refresh tokens: Longer-lived, sliding window with activity
- Session: Container with absolute maximum lifetime
- Activity resets idle timer, not absolute expiry

#### Token Lifecycle

| State | Behavior |
|-------|----------|
| Active user | Silent token refresh in background |
| Idle (approaching timeout) | Warning displayed to user |
| Idle (timeout exceeded) | Quick re-auth required (PIN, click to continue) |
| Absolute session max | Full re-authentication required |

#### Three Timeout Model

| Timeout | Default | Purpose | Configurable By |
|---------|---------|---------|-----------------|
| Token TTL | 15 min | Limit exposure of compromised token | Platform |
| Idle Timeout | 30 min | Lock inactive sessions | Tenant Admin |
| Absolute Max | 12 hours | Force periodic full re-auth | Platform |

---

## 3. Authorization Model

### 3.1 RBAC Hierarchy

Roles exist at three levels:

```
┌─────────────────────────────────────────────────────────────┐
│ Platform Roles (built-in, immutable)                        │
│   platform:superadmin    - Full platform access             │
│   platform:tenant-admin  - Manage tenant settings           │
│   platform:auditor       - Read-only audit access           │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Application Roles (defined by app, templated per tenant)    │
│   {app}:admin            - Full access to app resources     │
│   {app}:reader           - Read-only access                 │
│   {app}:writer           - Read + write access              │
│   {app}:{custom}         - App-specific roles               │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ Tenant Custom Roles (tenant-defined, composable)            │
│   {tenant}:{role}        - Inherits from app roles          │
│                          - Can combine multiple apps        │
│                          - Can add conditions               │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 Resource Model

Applications register their resource types with Bastion:

```yaml
application: signal-smith
resources:
  - name: story
    actions: [read, write, delete, admin]

  - name: episode
    actions: [read, write, delete]
    parent: story

  - name: investigation
    actions: [read, write, assign, close]
```

**Standard Actions** (available to all resource types):
- `read` - View resource
- `write` - Create or modify
- `delete` - Remove
- `admin` - Manage permissions on resource

**Custom Actions**: Applications can define additional actions specific to their domain.

### 3.3 Permission Structure

A permission grants an action on a resource type with a scope:

```
Permission = (resource_type, action, scope, conditions)
```

**Scopes**:
| Scope | Meaning |
|-------|---------|
| `tenant` | All resources of type within tenant |
| `owned` | Resources created by this identity |
| `assigned` | Resources explicitly assigned to identity |
| `team` | Resources belonging to identity's team(s) |
| `specific:{id}` | One specific resource instance |

**Conditions** (optional):
- Time windows (business hours only)
- IP ranges (internal network only)
- Attribute matching (classification level, etc.)

### 3.4 Role Inheritance

Roles can inherit from other roles:

```yaml
role: acme:security-analyst
inherits:
  - signal-smith:analyst
  - vektera:reader
additional_permissions:
  - resource: signal-smith:investigation
    action: assign
    scope: team
```

Inheritance is additive. A role receives all permissions from inherited roles plus its own.

---

## 4. Application Integration

### 4.1 Registration Flow

```
┌─────────────┐                      ┌─────────────────┐
│    App      │                      │     Bastion     │
└──────┬──────┘                      └────────┬────────┘
       │                                      │
       │ 1. Register application              │
       │ ────────────────────────────────────▶│
       │    {name, resources, default_roles}  │
       │                                      │
       │◀──────────────────────────────────── │
       │    {app_id, service_credentials}     │
       │                                      │
```

### 4.2 Authorization Check API

Applications call Bastion to check authorization:

```
POST /api/v1/authz/check

Request:
{
  "identity": {
    "type": "user",
    "id": "user-uuid",
    "tenant_id": "tenant-uuid"
  },
  "action": "write",
  "resource": {
    "type": "signal-smith:story",
    "id": "story-uuid",
    "attributes": {}
  },
  "context": {
    "ip": "10.0.1.50",
    "timestamp": "2025-01-28T10:30:00Z"
  }
}

Response:
{
  "allowed": true,
  "reason": "role:signal-smith:author grants write on story",
  "audit_id": "audit-uuid"
}
```

### 4.3 Caching Strategy

To avoid per-request latency, applications cache policy data:

| Data | Cache TTL | Invalidation |
|------|-----------|--------------|
| User role assignments | 5 min | Push on change |
| Role permission mappings | 15 min | Push on change |
| Resource ownership | 1 min | App responsibility |

**Hybrid Approach**:
- Cache hit + simple check = decide locally
- Cache miss or complex condition = call Bastion

---

## 5. Audit Model

### 5.1 Audit Events

Every security-relevant action is logged:

| Event Type | Data Captured |
|------------|---------------|
| Authentication | identity, method, success/failure, IP, user agent |
| Token Refresh | identity, session_id, new_expiry |
| Authorization Check | identity, resource, action, decision, reason |
| Role Assignment | granter, grantee, role, tenant |
| Permission Change | actor, role, old_permissions, new_permissions |

### 5.2 Audit Storage

Audit logs are:
- Append-only (immutable)
- Retained per compliance requirements (configurable)
- Queryable by tenant admins (their tenant only)
- Fully queryable by platform auditors

---

## 6. Data Model (Logical)

### 6.1 Core Entities

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Tenant    │────<│  Identity   │────<│   Session   │
└─────────────┘     └─────────────┘     └─────────────┘
                          │
                          │
                    ┌─────┴─────┐
                    │           │
              ┌─────┴─────┐ ┌───┴───────┐
              │IdentityRole│ │  API Key  │
              └─────┬─────┘ └───────────┘
                    │
              ┌─────┴─────┐
              │   Role    │────< RolePermission
              └───────────┘
                    │
              ┌─────┴─────┐
              │Permission │
              └───────────┘

┌─────────────┐     ┌─────────────┐
│ Application │────<│ResourceType │────< ResourceAction
└─────────────┘     └─────────────┘
```

### 6.2 Key Tables

**tenants**: Multi-tenant container
**identities**: Users, service accounts
**sessions**: Active sessions with activity tracking
**roles**: Platform, application, and custom roles
**role_permissions**: Permission grants to roles
**identity_roles**: Role assignments to identities
**applications**: Registered consuming applications
**resource_types**: App-defined resource types
**audit_log**: Immutable audit trail

---

## 7. API Surface

### 7.1 Authentication Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/auth/login` | POST | Authenticate, receive tokens |
| `/api/v1/auth/refresh` | POST | Refresh access token |
| `/api/v1/auth/logout` | POST | Invalidate session |
| `/api/v1/auth/sessions` | GET | List active sessions |
| `/api/v1/auth/sessions/{id}` | DELETE | Revoke specific session |

### 7.2 Authorization Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/authz/check` | POST | Check if action is allowed |
| `/api/v1/authz/permissions` | GET | List identity's permissions |

### 7.3 Identity Management Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/users` | GET, POST | List, create users |
| `/api/v1/users/{id}` | GET, PUT, DELETE | Manage user |
| `/api/v1/users/{id}/roles` | GET, POST, DELETE | Manage user roles |
| `/api/v1/service-accounts` | GET, POST | List, create service accounts |
| `/api/v1/api-keys` | GET, POST | List, create API keys |

### 7.4 Role Management Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/roles` | GET, POST | List, create roles |
| `/api/v1/roles/{id}` | GET, PUT, DELETE | Manage role |
| `/api/v1/roles/{id}/permissions` | GET, POST, DELETE | Manage role permissions |

### 7.5 Application Registration Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/applications` | GET, POST | List, register applications |
| `/api/v1/applications/{id}/resources` | GET, POST | Manage resource types |

---

## 8. Security Considerations

### 8.1 Zero Trust Implementation

- **No long-lived credentials**: Maximum token lifetime enforced
- **Continuous verification**: Every request validated
- **Least privilege**: Tokens scoped to minimum necessary
- **Assume breach**: Short TTLs limit exposure window

### 8.2 Credential Storage

- Passwords: bcrypt or argon2 hashed, never plaintext
- Refresh tokens: Hashed in database
- API key secrets: Hashed, shown once at creation

### 8.3 Transport Security

- All endpoints require TLS
- Service-to-service: mTLS where possible
- Token binding: Consider DPoP for high-security deployments

---

## 9. POC Scope

### 9.1 Phase 1: Core Authentication

**In Scope**:
- User identity (create, authenticate)
- Session management with activity tracking
- Access token + refresh token flow
- Token validation endpoint
- Basic audit logging

**Out of Scope**:
- MFA, SSO, federation
- Service accounts
- API keys

### 9.2 Phase 2: Basic RBAC

**In Scope**:
- Role definition and assignment
- Permission checking
- Single application (test app)
- Tenant isolation

**Out of Scope**:
- Role inheritance
- Custom conditions
- Multi-app scenarios

### 9.3 Phase 3: Application Integration

**In Scope**:
- Application registration
- Resource type registration
- Multi-application RBAC
- Caching strategy

---

## 10. Success Criteria

### POC Success

1. User can authenticate and receive short-lived token
2. Token refreshes silently while user is active
3. Session locks after idle timeout
4. RBAC correctly enforces permissions
5. Audit log captures all auth/authz events
6. Second test application can consume same Bastion instance

### Production Readiness (Future)

1. Sub-10ms authorization checks (cached)
2. 99.9% availability
3. Horizontal scaling capability
4. Comprehensive audit and compliance reporting
5. SSO/federation support

---

## 11. Open Questions

1. **Token Format**: PASETO vs JWT - PASETO has better defaults, JWT has broader tooling
2. **Policy Language**: Simple permission model vs OPA/Rego for complex policies
3. **Audit Retention**: How long to retain, where to store at scale
4. **Rate Limiting**: Per-tenant, per-identity, or both

---

## 12. References

- Zero Trust Architecture: NIST SP 800-207
- RBAC Standard: NIST RBAC Model
- OAuth 2.0: RFC 6749
- DPoP: RFC 9449
- SPIFFE: spiffe.io

---

## Appendix A: Glossary

| Term | Definition |
|------|------------|
| **Identity** | User, service account, or API key that can authenticate |
| **Session** | Container for tokens with activity tracking |
| **Access Token** | Short-lived credential for API access |
| **Refresh Token** | Longer-lived credential for obtaining new access tokens |
| **Role** | Named collection of permissions |
| **Permission** | Grant to perform action on resource type |
| **Scope** | Modifier limiting permission to subset of resources |
| **Tenant** | Isolated organization within the platform |
| **Resource Type** | Category of thing an application manages (story, collection) |
| **Action** | Operation on a resource (read, write, delete) |

---

## Appendix B: Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-01-28 | Architecture Team | Initial draft |
