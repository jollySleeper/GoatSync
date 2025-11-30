# GoatSync - LLM Quick Reference

> **Status: ✅ Migration Complete (v1.0.0)**
>
> GoatSync is a fully functional Go implementation of the EteSync server with 100% API compatibility.

## Project Overview

GoatSync is a **Go implementation of the EteSync server** (originally Python/FastAPI/Django).

| Aspect | Value |
|--------|-------|
| Language | Go 1.25+ |
| Framework | Gin |
| ORM | GORM |
| Database | PostgreSQL |
| Cache | Redis (optional) |
| Serialization | MessagePack |

## Quick Commands

```bash
# Start dependencies
docker compose up -d

# Build
go build -o goatsync ./cmd/server

# Run
DATABASE_URL="postgres://goatsync:goatsync@localhost:5432/goatsync?sslmode=disable" \
SECRET_KEY="your-secret-key-32-chars" \
./goatsync

# Test
go test ./... -v

# Stop
pkill goatsync && docker compose down
```

## Project Structure

```
goatSync/
├── cmd/server/main.go          # Entry point
├── internal/
│   ├── config/                 # Environment configuration
│   ├── crypto/                 # BLAKE2b, SecretBox, Ed25519
│   ├── database/               # GORM PostgreSQL connection
│   ├── model/                  # 9 GORM models
│   ├── repository/             # Data access layer
│   ├── service/                # Business logic
│   ├── handler/                # HTTP handlers
│   ├── middleware/             # Auth middleware
│   ├── server/                 # HTTP server setup
│   ├── storage/                # Chunk file storage
│   ├── redis/                  # Redis client
│   └── integration/            # Integration tests
├── pkg/
│   ├── errors/                 # EtebaseError types
│   └── utils/                  # Utilities
├── docker-compose.yml          # PostgreSQL + Redis
├── RUNNING.md                  # How to run everything
└── CHANGELOG.md                # Version history
```

## What's Implemented

### ✅ All API Endpoints (30+)
- Authentication (signup, login, logout, password change)
- Collections (list, create, get, list_multi)
- Items (list, get, batch, transaction, fetch_updates, revisions)
- Chunks (upload, download)
- Members (list, modify, remove, leave)
- Invitations (incoming/outgoing management)
- WebSocket (with Redis pub/sub)
- Health endpoints

### ✅ Crypto (Matching Python Exactly)
- BLAKE2b-256 with key, salt, personalization
- XSalsa20-Poly1305 (NaCl SecretBox)
- Ed25519 signature verification

### ✅ Stoken System
- Incremental sync with pagination
- Proper stoken generation and filtering

### ✅ Production Features
- Graceful shutdown
- Health/ready/live endpoints
- Debug mode with test reset

## Reference Files

| Topic | GoatSync File | Python Reference |
|-------|---------------|------------------|
| Auth | `internal/service/auth.go` | `eteSync-server/etebase_server/fastapi/routers/authentication.py` |
| Crypto | `internal/crypto/etebase.go` | (uses NaCl/libsodium) |
| Models | `internal/model/*.go` | `eteSync-server/etebase_server/django/models.py` |
| Stoken | `internal/repository/stoken_gorm.go` | `eteSync-server/etebase_server/fastapi/stoken_handler.py` |

## Cursor Rules

The `.cursor/rules/` directory contains AI-assistance rules:
- `01-project.mdc` - Project context
- `02-architecture.mdc` - Layer structure
- `03-api-routes.mdc` - All endpoints
- `04-crypto.mdc` - Crypto algorithms
- `05-models.mdc` - GORM models
- `06-stoken.mdc` - Sync system
- `07-errors.mdc` - Error codes
- `08-commits.mdc` - Commit conventions
- `09-commit-strategy.mdc` - Commit strategy

## Testing

```bash
# All tests
go test ./... -v

# Integration tests (requires Docker)
docker compose up -d
go test ./internal/integration/... -v
```

## Future Enhancements

The migration is complete. Potential improvements:
1. More comprehensive integration tests
2. Client compatibility testing (etesync-dav, mobile apps)
3. Performance optimization
4. Prometheus metrics
5. OpenTelemetry tracing
