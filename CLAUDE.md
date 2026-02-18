Inherits shared standards from ~/Development/CLAUDE.md

# CLAUDE.md - Bastion Project Guidelines

## Project Overview

**Bastion** is a reusable management plane providing Zero Trust authentication and Role-Based Access Control (RBAC) for multiple applications. Designed to be application-agnostic, Bastion enables consistent identity management, authentication, and authorization across Signal Smith, Vektera, and future platforms.

## Project-Specific Stack Additions

Beyond the shared approved stack, this project also uses:

- **Password Hashing**: golang.org/x/crypto/bcrypt (or argon2)
- **Token Signing**: Standard library crypto + PASETO or JWT (minimal library)

## Repository Structure

```
rbac-zta-poc/
├── poc/                    # POC codebase (Phase 1 - throwaway)
│   ├── cmd/               # Simple binaries
│   ├── internal/          # Inline implementations
│   └── README.md          # "THIS IS THROWAWAY CODE"
│
├── platform/              # Production codebase (Phase 2+ - permanent)
│   ├── cmd/               # Production services
│   ├── pkg/               # Reusable packages
│   └── README.md          # "Production implementation"
│
├── docs/
│   ├── decisions/         # Decision documents (DD-XXX)
│   └── sows/              # Statements of Work
│
└── CLAUDE.md              # This file
```

### POC Codebase (poc/)

**Purpose**: Prove Zero Trust auth and RBAC concepts
**Status**: Throwaway code, archived after Phase 1
**Standards**: Just make it work

**Characteristics**:
- Hardcoded test data acceptable
- Single binary preferred
- Inline SQL acceptable
- Crash on errors is acceptable for POC
- Printf debugging is fine
- Manual steps acceptable

**Success Criteria**: Authentication works, RBAC enforces permissions, tokens refresh correctly

### Production Codebase (platform/)

**Purpose**: Production-grade management plane
**Status**: Starts empty, built after POC succeeds
**Standards**: Full production quality

**Built From**: POC learnings + DDs + SOWs written with POC experience

## Core Architectural Principles

### Zero Trust Foundations
- Never trust, always verify
- Assume breach - design for compromised credentials
- Verify every request, not just at session start
- Least privilege - tokens grant minimum necessary access
- Short-lived credentials with continuous refresh

### Multi-Tenancy First
- Tenant isolation is foundational, not bolted on
- Every query scoped to tenant
- Cross-tenant access explicitly denied by default

### Application-Agnostic Design
- Applications register their resource types
- Management plane doesn't know application internals
- Same auth/authz patterns for all consumers

### Separation of Concerns
- **Identity**: Who (users, services, API keys)
- **Authentication**: Proving identity (tokens, credentials)
- **Authorization**: What can identity do (RBAC, policies)
- **Audit**: Recording all decisions

## REMEMBER

This project exists to build a reusable management plane. Avoid application-specific logic. Design for Signal Smith, Vektera, and applications not yet conceived.
