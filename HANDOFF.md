# Bastion Project Handoff

**Created**: 2025-01-28
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
Three levels: Platform roles (immutable) → Application roles (app-defined defaults) → Tenant custom roles (composable from app roles)

Tenants can create roles like "acme:soc-analyst" that combines "signal-smith:analyst" + "vektera:reader".

---

## Relationship to Other Projects

```
┌─────────────────────────────────────────┐
│              Bastion                     │
│  (Identity, Auth, RBAC, Audit)          │
└─────────────────┬───────────────────────┘
                  │
      ┌───────────┼───────────┐
      │           │           │
      ▼           ▼           ▼
┌──────────┐ ┌──────────┐ ┌──────────┐
│  Signal  │ │  Vektera │ │  Future  │
│  Smith   │ │          │ │   Apps   │
└──────────┘ └──────────┘ └──────────┘
```

- **Signal Smith**: Security signal detection platform (separate repo, uses same tech stack)
- **Vektera**: Proprietary data store being built (separate from Signal Smith)
- **Bastion**: Management plane serving both

All three share the same technology stack: Go, PostgreSQL, chi router, YAML config.

---

## Current State

### Completed
- `CLAUDE.md` - Project guidelines with SOW workflow constraints
- `docs/decisions/DD-001-bastion-architecture.md` - Comprehensive architecture decision document
- `docs/sows/POC-001.0-basic-go-structure.md` - First SOW, awaiting approval

### Awaiting Approval
**POC-001.0**: Basic Go project structure
- Module init, directory layout, health endpoint
- Minimal foundation for subsequent SOWs

### Next After POC-001.0
- POC-002: Core authentication (user identity, sessions, tokens)
- POC-003: Basic RBAC (roles, permissions, check endpoint)

---

## User Preferences

1. **SOW discipline is critical** - No code without explicit approval. This prevents scope creep and wasted effort. Previous iterations of Signal Smith failed when implementation ran ahead of architecture.

2. **Professional language** - No emojis, enterprise-grade terminology throughout.

3. **Start simple** - Prove concepts before adding complexity. POC is throwaway code; production comes later with lessons learned.

4. **Multi-tenancy in scope** - Even for POC, tenant isolation should be considered from the start.

5. **Same stack as Signal Smith** - Go 1.24+, PostgreSQL, chi router, YAML config. No unapproved dependencies.

---

## Technical Notes

### Token Format Decision (Open)
DD-001 notes PASETO vs JWT as an open question. PASETO has better cryptographic defaults; JWT has broader tooling. Decision can be made during POC-002.

### Database
PostgreSQL, same as Signal Smith. No ORM - use database/sql directly.

### Naming
"Bastion" is the working name. User is researching domains. The joke name "FUPA" (Fortress of Ultimate Power and Authority) should not appear in documentation.

---

## Session Continuity

To continue this work:
1. Review DD-001 for full architecture context
2. Review POC-001.0 for immediate next step
3. Get explicit approval before any implementation
4. Follow CLAUDE.md guidelines strictly
