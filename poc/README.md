# Bastion POC

**Status**: Proof of Concept (Throwaway Code)

This directory contains the Bastion proof-of-concept implementation. The code here validates architectural assumptions for Zero Trust authentication and RBAC before building the production platform.

## Purpose

- Prove Zero Trust auth and RBAC concepts work as designed
- Validate token lifecycle management
- Test RBAC permission enforcement
- Establish patterns for production implementation

## What This Is NOT

- **Not production code** - This will be archived after Phase 1
- **Not optimized** - Clarity over performance
- **Not complete** - Only implements what's needed to validate architecture

## Build and Run

```bash
cd poc
go mod tidy
go build ./cmd/bastion
./bastion --port 8080
```

## Endpoints

### Health Check

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "ok",
  "timestamp": "2025-01-28T10:30:00Z"
}
```

## Shutdown

Send SIGINT (Ctrl+C) or SIGTERM for graceful shutdown.

## Related Documentation

- `/docs/decisions/DD-001-bastion-architecture.md` - Architecture decision document
- `/docs/sows/POC-001.0-basic-go-structure.md` - This implementation's SOW
- `/CLAUDE.md` - Project guidelines
