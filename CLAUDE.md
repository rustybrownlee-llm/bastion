Inherits shared standards from ~/Development/CLAUDE.md

# Bastion Project Guidelines

## Project Overview

Bastion is a reusable management plane providing Zero Trust authentication and Role-Based Access Control (RBAC) for multiple applications. It exists because auth and RBAC are frequently afterthoughts that become painful to retrofit. Bastion solves this by building identity management, authentication, authorization, and audit as a standalone platform that any application can consume. The repo contains two codebases: `poc/` (throwaway proof-of-concept, now complete) and `platform/` (production implementation, not yet started).

## Current State

- **Phase**: POC complete (Phase 1 finished). Production phase not yet started.
- **Last completed work**: POC-004.0 Service Accounts and API Keys (2025-01-28).
- **POC coverage**: ~3,250 lines across 36 files. Auth, RBAC, service accounts, API keys all validated.
- **Next work**: Production planning (SOW-100+ series). Evaluate POC lessons learned, design production-grade implementation in `platform/`.
- **Blockers**: None. Awaiting decision to begin production phase.
- **Architecture doc**: `docs/decisions/DD-001-bastion-architecture.md`
- **SOW history**: `docs/sows/` contains POC-001 through POC-004 with validation results.

## Project Relationships

- **Signal Smith** (`~/Development/signal-smith`): Security signal detection platform. Will consume Bastion for all auth/authz. Same tech stack (Go, PostgreSQL, chi, YAML).
- **Vektera**: Proprietary data store being built separately. Will consume Bastion for all auth/authz.
- **Future applications**: Bastion is designed to serve applications not yet conceived. Tenants can create custom roles that span multiple apps (e.g., "acme:soc-analyst" combining "signal-smith:analyst" + "vektera:reader").

## Project-Specific Stack

| Component | Technology | Phase |
|-----------|-----------|-------|
| Password hashing | golang.org/x/crypto/bcrypt | POC, Production |
| Token signing | github.com/golang-jwt/jwt/v5 (PASETO for production TBD) | POC |
| PostgreSQL driver | github.com/lib/pq | POC, Production |
| HTTP routing | github.com/go-chi/chi/v5 | POC, Production |
| Configuration | gopkg.in/yaml.v3 | POC, Production |

## Configuration

- **Config file**: `poc/config.yaml` (YAML format, per shared standards)
- **Server port**: 8081 (8080 was occupied during POC testing)
- **Database**: PostgreSQL 15+ container `bastion-db`, database `bastion_poc`, user `bastion`
- **Migrations**: `poc/migrations/` (applied via psql, numbered sequentially: 001, 002, 003)
- **JWT signing**: HMAC secret in config (production must use asymmetric keys)
- **GitHub**: https://github.com/rustybrownlee-llm/bastion

## Key Principles

- **Zero Trust**: Never trust, always verify. Short-lived credentials (15 min access tokens, 24h refresh tokens), continuous verification on every request, assume breach, least privilege.
- **Activity-based sessions**: Three-tier timeout: token TTL (15 min), idle timeout (30 min), absolute max (12 hours). Active users rotate tokens silently; idle users re-auth.
- **Multi-tenancy first**: Tenant isolation is foundational. Every query scoped to tenant. Cross-tenant access denied by default. Platform roles (superadmin, admin, auditor) operate cross-tenant via nullable tenant_id.
- **Application-agnostic**: Applications register their resource types. Bastion does not know application internals. Same auth/authz patterns for all consumers.
- **Hierarchical roles**: Three levels -- platform roles (immutable), application roles (app-defined defaults), tenant custom roles (composable across apps).
- **Separation of concerns**: Identity (who), Authentication (proving identity), Authorization (what they can do via RBAC), Audit (recording all decisions) are distinct subsystems.
- **Two permission models**: Role-based RBAC for users and service accounts. Direct permission grants for API keys.
- **Three identity types**: Human users (password + JWT), service accounts (client credentials + JWT), API keys (X-API-Key header, direct permission grants).

## Lessons and Anti-Patterns

- JWT chosen for POC debuggability (jwt.io). Evaluate PASETO for production hardening.
- No ORM. Direct database/sql gives needed control for tenant-scoped queries.
- Service accounts use OAuth2 client credentials flow. API keys use X-API-Key header. Do not conflate these identity types.
- All secrets (passwords, client secrets, API keys) are bcrypt hashed and shown only once at creation. No plaintext storage.
- POC used inline SQL and hardcoded test data. Production must use connection pooling, caching, and proper config management.
- Anti-pattern: Do not add application-specific resource types into Bastion code. Applications register their own resource types at runtime.

## REMEMBER

- This project builds a reusable management plane. Avoid application-specific logic.
- Design for Signal Smith, Vektera, and applications not yet conceived.
- The POC (`poc/`) is complete and archived. Do not modify POC code.
- Production work begins in `platform/` under the SOW-100+ series.
- "Bastion" is the working name. Domain research is pending.
- Security that breaks workflow gets bypassed. Balance security with usability.
