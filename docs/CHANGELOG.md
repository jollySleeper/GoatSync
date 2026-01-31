# GoatSync Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.1] - 2026-02-01

### Fixed
- Environment variable renamed from `SECRET_KEY` to `ENCRYPTION_SECRET` to match actual code
- Default PORT changed from 8080 to 3735 in config.go
- Default DEBUG changed to false and GIN_MODE to release for production defaults
- Updated all documentation to reference `.env.example` for configuration

### Changed
- `.env.example` now contains complete list of all environment variables with descriptions
- Docker compose files now use `ENCRYPTION_SECRET` instead of `SECRET_KEY`

---

## [0.1.0] - 2026-01-31

### 🚀 First Public Release

First stable release with full EteSync API compatibility.

### Added
- Docker image published to GitHub Container Registry (`ghcr.io/jollysleeper/goatsync`)
- Multi-platform support (linux/amd64, linux/arm64)
- `docker-compose.yml` for standard deployment
- `docker-compose-full.yml` with EteSync-DAV for CalDAV/CardDAV support
- `.env.example` with documented configuration options
- `docs/DEPLOYMENT.md` - comprehensive deployment guide

### Fixed
- Authentication token parsing now correctly strips "Token " prefix
- All `errcheck` lint violations resolved
- Updated to golangci-lint v2.6 for Go 1.25 compatibility

### Changed
- Default port changed from 8080 to 3735 in Dockerfile
- Docker builds now trigger only on version tags (not branch pushes)
- Docker compose files use `env_file` for cleaner configuration

---

## [0.0.1] - 2025-12-01

### 🎉 Migration Complete!

**GoatSync is now a fully functional Go implementation of the EteSync server with 100% API compatibility.**

### Summary
- **135 commits** over the migration period
- **66 Go source files**
- **6 test suites** with comprehensive tests
- **~10,000 lines of code**
- **100% API endpoint coverage**

---

### Added

#### Core Infrastructure
- **PostgreSQL + GORM** database integration (`internal/database/postgres.go`)
- **Redis** client with pub/sub support (`internal/redis/redis.go`)
- **Docker Compose** for PostgreSQL and Redis (`docker-compose.yml`)
- **Graceful shutdown** with signal handling (SIGTERM/SIGINT)
- **Health endpoints** (`/health`, `/ready`, `/live`)

#### Cryptography (Matching Python Exactly)
- **BLAKE2b-256** key derivation with key, salt, personalization (`internal/crypto/etebase.go`)
- **XSalsa20-Poly1305** (NaCl SecretBox) for challenge encryption
- **Ed25519** signature verification for login
- **Secure UID generation** for stokens, items, collections

#### GORM Models (9 Models)
- `Stoken` - Sync token for incremental updates
- `User` + `UserInfo` - User accounts with encrypted data
- `AuthToken` - Session tokens
- `CollectionType` - Collection categories
- `Collection` - User collections (calendars, contacts, etc.)
- `CollectionItem` - Items within collections
- `CollectionItemRevision` - Item version history
- `CollectionItemChunk` + `RevisionChunkRelation` - Binary data chunks
- `CollectionMember` + `CollectionMemberRemoved` - Sharing
- `CollectionInvitation` - Pending invitations

#### Repository Layer
- `UserRepository` - User CRUD operations
- `TokenRepository` - Auth token management
- `StokenRepository` - Sync token queries with filtering
- `CollectionRepository` - Collection queries with stoken pagination
- `ItemRepository` - Item queries with stoken pagination
- `MemberRepository` - Member management
- `InvitationRepository` - Invitation management
- `ChunkRepository` - Chunk metadata
- `RevisionRepository` - Revision history
- `CollectionTypeRepository` - Collection type management

#### Service Layer
- `AuthService` - Login challenge, login, signup, logout, password change
- `CollectionService` - List, create, get collections
- `ItemService` - List, get, batch, transaction, fetch updates, revisions
- `MemberService` - List, modify, remove, leave
- `InvitationService` - Incoming/outgoing invitations
- `ChunkService` - Upload/download chunks

#### HTTP Handlers (30+ Endpoints)
- **Authentication** (7 endpoints)
  - `GET /api/v1/authentication/is_etebase/`
  - `POST /api/v1/authentication/login_challenge/`
  - `POST /api/v1/authentication/login/`
  - `POST /api/v1/authentication/logout/`
  - `POST /api/v1/authentication/signup/`
  - `POST /api/v1/authentication/change_password/`
  - `POST /api/v1/authentication/dashboard_url/`

- **Collections** (4 endpoints)
  - `GET /api/v1/collection/`
  - `POST /api/v1/collection/`
  - `POST /api/v1/collection/list_multi/`
  - `GET /api/v1/collection/:collection_uid/`

- **Items** (8 endpoints)
  - `GET /api/v1/collection/:uid/item/`
  - `GET /api/v1/collection/:uid/item/:item_uid/`
  - `GET /api/v1/collection/:uid/item/:item_uid/revision/`
  - `POST /api/v1/collection/:uid/item/batch/`
  - `POST /api/v1/collection/:uid/item/transaction/`
  - `POST /api/v1/collection/:uid/item/fetch_updates/`
  - `PUT /api/v1/collection/:uid/item/:item_uid/chunk/:chunk_uid/`
  - `GET /api/v1/collection/:uid/item/:item_uid/chunk/:chunk_uid/download/`

- **Members** (4 endpoints)
  - `GET /api/v1/collection/:uid/member/`
  - `DELETE /api/v1/collection/:uid/member/:username/`
  - `PATCH /api/v1/collection/:uid/member/:username/`
  - `POST /api/v1/collection/:uid/member/leave/`

- **Invitations** (8 endpoints)
  - `GET /api/v1/invitation/incoming/`
  - `GET /api/v1/invitation/incoming/:invitation_uid/`
  - `DELETE /api/v1/invitation/incoming/:invitation_uid/`
  - `POST /api/v1/invitation/incoming/:invitation_uid/accept/`
  - `GET /api/v1/invitation/outgoing/`
  - `DELETE /api/v1/invitation/outgoing/:invitation_uid/`
  - `POST /api/v1/invitation/outgoing/fetch_user_profile/`

- **WebSocket** (1 endpoint)
  - `GET /api/v1/ws/:ticket/`

- **Debug** (1 endpoint, DEBUG mode only)
  - `POST /api/v1/test/authentication/reset/`

#### Storage
- **Filesystem chunk storage** (`internal/storage/filesystem.go`)
- Hierarchical path structure matching Python implementation

#### Testing
- **Unit tests** for crypto, errors, models, storage, services
- **Integration tests** for full API flow
- **Benchmarks** for key derivation and encryption

#### Middleware
- **Authentication middleware** with token validation
- **CORS middleware** with configurable origins

### Fixed
- Crypto implementation (was using bcrypt, now uses Ed25519)
- Database layer (was in-memory, now PostgreSQL)
- Stoken system (was missing, now fully implemented)
- Foreign key constraints during migration (circular dependency handling)

### Changed
- Complete rewrite of handler layer with proper service injection
- MessagePack responses for all endpoints
- Error responses match EteSync format exactly

### Dependencies Added
```
gorm.io/gorm v1.25.12
gorm.io/driver/postgres v1.5.11
github.com/go-redis/redis/v8 v8.11.5
github.com/dchest/blake2b (for full BLAKE2b config support)
```

---

## [0.0.1] - 2025-03-20

Initial scaffolding release.

### Added
- Basic project structure
- Gin HTTP server
- Route definitions
- Handler stubs (non-functional)
- In-memory data storage (temporary)

### Known Issues (All Fixed in 1.0.0)
- ❌ Crypto was fundamentally wrong (bcrypt instead of Ed25519)
- ❌ No database (in-memory maps)
- ❌ No stoken system
- ❌ Invitation handlers missing

---

## Migration Comparison

### Python EteSync Server vs GoatSync

| Feature | Python | GoatSync | Status |
|---------|--------|----------|--------|
| Framework | FastAPI | Gin | ✅ |
| ORM | Django | GORM | ✅ |
| Database | PostgreSQL | PostgreSQL | ✅ |
| Serialization | MessagePack | MessagePack | ✅ |
| Crypto - BLAKE2b | ✅ | ✅ | ✅ |
| Crypto - SecretBox | ✅ | ✅ | ✅ |
| Crypto - Ed25519 | ✅ | ✅ | ✅ |
| Stoken System | ✅ | ✅ | ✅ |
| WebSocket | ✅ | ✅ | ✅ |
| Redis Pub/Sub | ✅ | ✅ | ✅ |
| All API Endpoints | 30+ | 30+ | ✅ |
| Error Codes | ✅ | ✅ | ✅ |

---

## How to Verify 1:1 Compatibility

1. Run both servers side-by-side
2. Use the same database
3. Test with real EteSync clients (etesync-dav, web, iOS, Android)
4. Compare response bodies byte-for-byte
5. Run integration tests against both

See `RUNNING.md` for detailed instructions.
