# CLAUDE.md - Bastion Project Guidelines

## Project Overview

**Bastion** is a reusable management plane providing Zero Trust authentication and Role-Based Access Control (RBAC) for multiple applications. Designed to be application-agnostic, Bastion enables consistent identity management, authentication, and authorization across Signal Smith, Vektera, and future platforms.

## CRITICAL COMMANDMENT: NO CODE WITHOUT SOW APPROVAL

### ABSOLUTELY FORBIDDEN - THIS IS NON-NEGOTIABLE:
- **NO CODE IMPLEMENTATION** without explicit SOW (Statement of Work) approval
- **NO FILE CREATION** beyond documentation without SOW approval
- **NO DIRECTORY CREATION** without SOW approval
- **NO "HELPFUL" IMPLEMENTATION** - wait for explicit approval
- **NO JUMPING AHEAD** - SOW must be reviewed and approved FIRST
- **NO EXTERNAL DEPENDENCIES** without checking go.mod and getting approval
- **NO NEW TECHNOLOGIES** outside the approved stack

### MANDATORY WORKFLOW - NO EXCEPTIONS:
1. **PRE-SOW CHECKLIST** - MUST complete before creating ANY SOW:
   - [ ] Check go.mod for existing dependencies
   - [ ] Review existing code patterns in /poc and /platform
   - [ ] Verify technology stack compliance
   - [ ] Check for similar existing implementations
2. **CREATE SOW DOCUMENT** - Complete specification with scope, structure, deliverables
3. **PRESENT SOW FOR REVIEW** - User must review the complete SOW
4. **WAIT FOR EXPLICIT APPROVAL** - User must say "approved" or "execute"
5. **ONLY THEN IMPLEMENT** - Follow SOW specifications exactly
6. **NO DEVIATIONS** - Any changes require new SOW or amendment

**VIOLATION = IMMEDIATE WORK STOPPAGE AND PROJECT FAILURE**

## APPROVED TECHNOLOGY STACK - NO DEVIATIONS

### MANDATORY: Check Before ANY Recommendation
**BEFORE suggesting ANY technology, library, or pattern:**
1. CHECK go.mod for existing dependencies
2. CHECK existing code for established patterns
3. CHECK this section for approved technologies
4. DEFAULT to standard library solutions

### Approved Stack ONLY:
- **Language**: Go (1.24+) - standard library preferred
- **Database**: PostgreSQL 15+
- **Configuration**: YAML files only (gopkg.in/yaml.v3)
- **HTTP Routing**: chi router (github.com/go-chi/chi/v5) - minimal, stdlib-compatible
- **Password Hashing**: golang.org/x/crypto/bcrypt (or argon2)
- **Token Signing**: Standard library crypto + PASETO or JWT (minimal library)
- **Container**: Docker and Docker Compose
- **Testing**: Standard library testing package

### FORBIDDEN Technologies:
- External CLI frameworks (Cobra, urfave/cli, etc.) - use standard flag package
- ORMs or database abstractions beyond database/sql
- Heavy web frameworks (gin, echo, fiber, etc.) - use chi router
- External logging libraries - use standard log or custom minimal kit
- Configuration libraries beyond yaml.v3
- ANY dependency not explicitly approved

## Repository Structure: Two-Codebase Approach

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

## SOW Requirements

### SOW Types

**POC SOWs** (POC-001 to POC-099):
- Purpose: Validate architectural assumptions
- Codebase: `poc/`
- Standards: Just make it work
- Format: `POC-XXX.Y-description.md`

**Production SOWs** (SOW-100 to SOW-999):
- Purpose: Production implementation
- Codebase: `platform/`
- Standards: Full production quality
- Format: `SOW-XXX.Y-description.md`

### Every SOW MUST Include:
1. **Objective** - Clear statement of purpose
2. **Scope** - What's included and excluded
3. **Deliverables** - Specific files/directories to create
4. **Success Criteria** - Measurable validation points
5. **Dependencies** - Required prior SOWs
6. **Technical Approach** - Implementation details

## Professional Standards

### MANDATORY:
- **NO emojis** in code, comments, scripts, documentation, or UI
- Professional enterprise language only
- Clear technical terminology throughout
- Enterprise-grade quality in all deliverables

## Component Size Limits

### Backend Code:
- Service functions: **60 lines maximum**
- File size: **500 lines maximum** (excluding tests)
- Package complexity: **10 public functions maximum**
- Interface methods: **5 methods maximum**

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

**THE PRIME DIRECTIVE**: No implementation without explicit SOW approval. PERIOD.
