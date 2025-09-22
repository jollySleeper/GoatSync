# GoatSync Dependencies

## Current Dependencies (in go.mod)

```
github.com/gin-gonic/gin v1.11.0        # HTTP framework
github.com/google/uuid v1.6.0            # UUID generation
github.com/gorilla/websocket v1.5.3      # WebSocket support
github.com/vmihailenco/msgpack/v5 v5.4.1 # MessagePack serialization
golang.org/x/crypto v0.42.0              # Crypto primitives
```

## Required Dependencies to Add

Run these commands to add the missing dependencies:

```bash
# Database (GORM + PostgreSQL)
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres

# Redis (for WebSocket tickets and pub/sub)
go get -u github.com/redis/go-redis/v9

# Environment configuration
go get -u github.com/joho/godotenv

# Structured logging
go get -u go.uber.org/zap

# Testing
go get -u github.com/stretchr/testify

# Then tidy up
go mod tidy
```

## Updated go.mod (after adding dependencies)

```go
module goatsync

go 1.25

require (
    // HTTP
    github.com/gin-gonic/gin v1.11.0
    github.com/gorilla/websocket v1.5.3
    
    // Database
    gorm.io/gorm v1.25.5
    gorm.io/driver/postgres v1.5.4
    
    // Cache/PubSub
    github.com/redis/go-redis/v9 v9.3.0
    
    // Serialization
    github.com/vmihailenco/msgpack/v5 v5.4.1
    
    // Crypto
    golang.org/x/crypto v0.42.0
    
    // Utils
    github.com/google/uuid v1.6.0
    github.com/joho/godotenv v1.5.1
    
    // Logging
    go.uber.org/zap v1.26.0
    
    // Testing
    github.com/stretchr/testify v1.8.4
)
```

## Dependency Purposes

| Package | Purpose | Why Needed |
|---------|---------|------------|
| `gorm.io/gorm` | ORM | Database access, model definitions |
| `gorm.io/driver/postgres` | PostgreSQL driver | Database connectivity |
| `github.com/redis/go-redis/v9` | Redis client | WebSocket tickets, pub/sub |
| `github.com/joho/godotenv` | Env loading | Development configuration |
| `go.uber.org/zap` | Logging | Structured, performant logging |
| `github.com/stretchr/testify` | Testing | Assertions, mocking |

## NaCl Crypto Note

The EteSync protocol requires NaCl-compatible cryptography:

- **BLAKE2b** with key, salt, and personalization: `golang.org/x/crypto/blake2b`
- **SecretBox** (XSalsa20-Poly1305): `golang.org/x/crypto/nacl/secretbox`
- **Ed25519** signatures: `crypto/ed25519` (standard library)

All of these are available in `golang.org/x/crypto` which is already in go.mod.

## Installation Script

Create `scripts/setup.sh`:

```bash
#!/bin/bash
set -e

echo "Installing Go dependencies..."

# Core dependencies
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
go get -u github.com/redis/go-redis/v9
go get -u github.com/joho/godotenv
go get -u go.uber.org/zap
go get -u github.com/stretchr/testify

# Tidy up
go mod tidy

echo "Dependencies installed successfully!"
echo ""
echo "Next steps:"
echo "1. Copy .env.example to .env"
echo "2. Configure DATABASE_URL and ENCRYPTION_SECRET"
echo "3. Run: go run ./cmd/goatsync"
```

Make it executable:
```bash
chmod +x scripts/setup.sh
```

