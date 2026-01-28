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
| SOW Agent | `.claude/agents/sow-implementation-agent.md` created |

### POC-001.0 Implementation (Complete)
- Go module at `github.com/rustybrownlee-llm/bastion/poc`
- chi router with `/health` endpoint
- Graceful shutdown on SIGINT/SIGTERM
- All success criteria validated

### Awaiting Approval

**POC-002.0**: Core Authentication (`docs/sows/POC-002.0-core-authentication.md`)
- PostgreSQL database connection
- User creation with bcrypt password hashing
- Login endpoint returning JWT access + refresh tokens
- Token refresh and logout endpoints
- Session tracking with activity timestamps
- Basic audit logging
- **NOT YET COMMITTED** - file exists locally but not in git

### SOW Implementation Agent

A custom agent was created at `.claude/agents/sow-implementation-agent.md` for implementing approved SOWs. Use this agent to execute POC-002.0 after approval.

The agent enforces:
- No implementation without SOW approval
- POC vs Production standards
- Approved technology stack only
- Validation and lessons learned documentation

---

## Next Steps

1. **Review POC-002.0** at `docs/sows/POC-002.0-core-authentication.md`
2. **Approve or request changes** to the SOW
3. **Use SOW implementation agent** to execute the approved SOW
4. After POC-002: Draft POC-003 (Basic RBAC)

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

### Dependencies Added (POC-001.0)
- `github.com/go-chi/chi/v5 v5.2.4`

### Dependencies for POC-002.0 (pending approval)
- `github.com/golang-jwt/jwt/v5` - JWT tokens
- `github.com/lib/pq` - PostgreSQL driver
- `golang.org/x/crypto` - bcrypt
- `gopkg.in/yaml.v3` - YAML config

### Naming
"Bastion" is the working name. User is researching domains.

---

## Session Continuity

To continue this work:
1. Read CLAUDE.md for project rules
2. Read this HANDOFF.md for current state
3. Review POC-002.0 SOW (not yet committed, exists locally)
4. Get explicit approval before implementing
5. Use the SOW implementation agent for execution
