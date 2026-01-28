# Bastion Project Handoff

**Created**: 2025-01-28
**Last Updated**: 2025-01-28
**Purpose**: Context for session continuity

---

## Origin Story

Bastion emerged from a discussion about Signal Smith's architecture. The original question was whether to put authentication and RBAC into Signal Smith's control plane. The conclusion: **design a separate management plane that can serve multiple applications**.

The insight: Auth and RBAC are frequently afterthoughts in application development (user has a decade of Splunk admin experience confirming this pattern). Retrofitting security is painful because identity/authorization touches everything. Building it right once, as a reusable platform, pays dividends.

**Bastion serves**: Signal Smith, Vektera (a proprietary data store being built separately), and future applications.

---

## Key Design Decisions and Reasoning

### Zero Trust Architecture
Commercial software increasingly requires Zero Trust. Design for it from day one rather than retrofitting. Key principles:
- Short-lived credentials (tokens expire in 10-15 minutes)
- Continuous verification (every request, not just at login)
- Assume breach (limit blast radius of compromised credentials)

### Activity-Based Session Management
The user asked: "How do we handle someone working in the platform all day?"

The solution uses the **laptop lock analogy**:
- Active users: Tokens rotate silently in background, session extends
- Idle users: Session locks after timeout, quick re-auth required
- Long idle: Full re-authentication

Three separate timeouts:
1. Token TTL (15 min) - security, users never notice
2. Idle timeout (30 min) - locks inactive sessions
3. Absolute max (12 hours) - forces periodic full re-auth

This balances security with workflow - security that breaks workflow gets bypassed.

### Application-Agnostic RBAC
Applications register their resource types with Bastion. Bastion doesn't know application internals - it just knows "signal-smith:story" is a thing that has read/write/delete actions.

This enables:
- Consistent auth/authz patterns across all apps
- Tenant-defined custom roles that span multiple apps
- Single audit trail for all access decisions

### Hierarchical Roles
Three levels: Platform roles (immutable) -> Application roles (app-defined defaults) -> Tenant custom roles (composable from app roles)

Tenants can create roles like "acme:soc-analyst" that combines "signal-smith:analyst" + "vektera:reader".

### Token Format Decision (Resolved)
JWT selected for POC. Simpler, well-understood, tokens can be debugged at jwt.io. PASETO can be evaluated for production hardening later.

---

## Relationship to Other Projects

```
                    Bastion
        (Identity, Auth, RBAC, Audit)
                      |
        +-------------+-------------+
        |             |             |
        v             v             v
   Signal Smith    Vektera      Future Apps
```

- **Signal Smith**: Security signal detection platform (separate repo, uses same tech stack)
- **Vektera**: Proprietary data store being built (separate from Signal Smith)
- **Bastion**: Management plane serving both

All three share the same technology stack: Go, PostgreSQL, chi router, YAML config.

---

## Repository

**GitHub**: https://github.com/rustybrownlee-llm/bastion
**Local**: /Users/rustybrownlee/Development/rbac-zta-poc

---

## Current State

### Completed

| Item | Description |
|------|-------------|
| CLAUDE.md | Project guidelines with SOW workflow constraints |
| DD-001 | Comprehensive architecture decision document |
| POC-001.0 | Basic Go structure - **APPROVED AND IMPLEMENTED** |
| POC-002.0 | Core Authentication - **APPROVED AND IMPLEMENTED** |
| POC-003.0 | Basic RBAC - **APPROVED AND IMPLEMENTED** |
| SOW Agent | `.claude/agents/sow-implementation-agent.md` created |
| Reference Docs | `docs/reference/` with API, schema, config, validation guides |

### POC-001.0 Implementation (Complete)
- Go module at `github.com/rustybrownlee-llm/bastion/poc`
- chi router with `/health` endpoint
- Graceful shutdown on SIGINT/SIGTERM
- All success criteria validated

### POC-002.0 Implementation (Complete)
- PostgreSQL database connection with lib/pq driver
- User creation with bcrypt password hashing
- Login endpoint returning JWT access + refresh tokens
- Token refresh and logout endpoints
- Session tracking with activity timestamps
- Basic audit logging to database
- Token validation middleware
- All 8 success criteria validated

### POC-003.0 Implementation (Complete)
- Multi-tenancy with tenant table and user assignment
- Platform roles (superadmin, admin, auditor)
- Application roles (tenant-admin, user-admin, viewer)
- Permissions table (resource_type + action)
- Role-permission mappings
- User-role assignments per tenant
- Authorization check endpoint (`/api/v1/authz/check`)
- RequirePermission middleware
- JWT claims include tenant_id
- Bootstrap: admin@bastion.local is platform:superadmin
- All 12 success criteria validated

### SOW Implementation Agent

A custom agent was created at `.claude/agents/sow-implementation-agent.md` for implementing approved SOWs.

The agent enforces:
- No implementation without SOW approval
- POC vs Production standards
- Approved technology stack only
- Validation and lessons learned documentation

---

## Next Steps

1. **Draft POC-004** - Service accounts and API keys
2. **Production planning** - Evaluate POC lessons learned, plan SOW-100+ series
3. **Optional**: Bootstrap UI for testing (separate SOW)

---

## User Preferences

1. **SOW discipline is critical** - No code without explicit approval. This prevents scope creep and wasted effort. Previous iterations of Signal Smith failed when implementation ran ahead of architecture.

2. **Professional language** - No emojis, enterprise-grade terminology throughout.

3. **Start simple** - Prove concepts before adding complexity. POC is throwaway code; production comes later with lessons learned.

4. **Multi-tenancy in scope** - Even for POC, tenant isolation should be considered from the start.

5. **Same stack as Signal Smith** - Go 1.24+, PostgreSQL, chi router, YAML config. No unapproved dependencies.

6. **UI for testing** - User mentioned wanting a Bootstrap UI for testing later. This would be a separate SOW and decision document, not part of the core POC.

---

## Technical Notes

### Database
PostgreSQL, same as Signal Smith. No ORM - use database/sql directly.

### Dependencies

| Package | Version | Added In | Purpose |
|---------|---------|----------|---------|
| github.com/go-chi/chi/v5 | v5.2.4 | POC-001.0 | HTTP routing |
| github.com/golang-jwt/jwt/v5 | v5.3.1 | POC-002.0 | JWT tokens |
| github.com/lib/pq | v1.10.9 | POC-002.0 | PostgreSQL driver |
| golang.org/x/crypto | v0.47.0 | POC-002.0 | bcrypt password hashing |
| gopkg.in/yaml.v3 | v3.0.1 | POC-002.0 | YAML configuration |

### Naming
"Bastion" is the working name. User is researching domains.

---

## Session Continuity

To continue this work:
1. Read CLAUDE.md for project rules
2. Read this HANDOFF.md for current state
3. Review completed SOWs in `docs/sows/` for context
4. Get explicit approval before implementing any new SOW
5. Use the SOW implementation agent for execution

### Running the POC

```bash
# Start PostgreSQL
docker start bastion-db  # or create new container

# Apply all migrations
docker exec -i bastion-db psql -U bastion -d bastion_poc < poc/migrations/001_initial_schema.sql
docker exec -i bastion-db psql -U bastion -d bastion_poc < poc/migrations/002_rbac_schema.sql

# Run server (port 8081)
cd poc && go run ./cmd/bastion -config config.yaml
```

### Test Admin User
- Email: `admin@bastion.local`
- Password: `BastionAdmin2025`
- Role: `platform:superadmin`
