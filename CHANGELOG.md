# GoatSync Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Migration Progress
- **Overall:** ~8-10% complete
- **Crypto:** ❌ Broken (uses bcrypt instead of Ed25519)
- **Database:** ❌ Using in-memory maps
- **Stoken System:** ❌ Not implemented

### Added
- Project structure (`cmd/`, `internal/`, `pkg/`, `api/`)
- Gin framework setup with basic routing
- MessagePack codec for request/response serialization
- Environment configuration system (`internal/config/`)
- Handler stubs for auth, collection, member endpoints
- WebSocket handler stub
- Cursor rules for AI-assisted development (`.cursor/rules/`)
- Migration documentation (`ARCHITECTURE.md`, `MIGRATION_STATUS.md`, `LLM_START_HERE.md`)

### Fixed
- Directory typo: `internal/respository` → `internal/repository`
- Package name mismatch in `api/routes/collection.go`
- Undefined variable `user` in `api/routes/users.go:12`
- Unused variable `apiEngine` in `api/routes/routes.go`

### Changed
- Go version: 1.23.7 → 1.25.1

### Dependencies
Current dependencies in go.mod:
- github.com/gin-gonic/gin v1.11.0
- github.com/google/uuid v1.6.0
- github.com/gorilla/websocket v1.5.3
- github.com/vmihailenco/msgpack/v5 v5.4.1
- golang.org/x/crypto v0.42.0

### Known Issues
1. **Crypto is fundamentally wrong** - Uses bcrypt instead of NaCl (Ed25519, SecretBox, BLAKE2b)
2. **No database** - All data stored in memory, lost on restart
3. **No stoken system** - Core sync mechanism not implemented
4. **Invitation handlers missing** - Not even scaffolded yet

---

## Migration Milestones

### Phase 1: Foundation (Not Started)
- [ ] Fix crypto implementation (`internal/crypto/etebase.go`)
- [ ] Add PostgreSQL + GORM (`internal/database/postgres.go`)
- [ ] Create Stoken model (`internal/model/stoken.go`)
- [ ] Create all GORM models

### Phase 2: Authentication (Not Started)
- [ ] Implement login challenge correctly
- [ ] Implement login with Ed25519 verification
- [ ] Implement signup with proper user creation
- [ ] Implement password change
- [ ] Implement logout

### Phase 3: Collections (Not Started)
- [ ] List collections with stoken pagination
- [ ] Create collection
- [ ] Get collection
- [ ] List items with stoken pagination

### Phase 4: Items (Not Started)
- [ ] Get item
- [ ] Item transaction (atomic with etag)
- [ ] Item batch (without etag)
- [ ] Chunk upload/download

### Phase 5: Members (Not Started)
- [ ] List members
- [ ] Update member access
- [ ] Remove member
- [ ] Leave collection

### Phase 6: Invitations (Not Started)
- [ ] List incoming invitations
- [ ] Accept/reject invitation
- [ ] Create outgoing invitation
- [ ] Delete outgoing invitation

### Phase 7: WebSocket (Not Started)
- [ ] Redis integration
- [ ] Ticket generation
- [ ] WebSocket handler
- [ ] Pub/sub for real-time updates

---

## Version History

### [0.0.1] - 2025-03-20
Initial scaffolding release.

#### Added
- Basic project structure
- Gin HTTP server
- Route definitions
- Handler stubs (non-functional)
- In-memory data storage (temporary)

---

## Commit History Reference

See git log for detailed commit history:
```bash
git log --oneline --since="2025-02-15"
```

Key commits:
- `2b3d0e1` - Added: README
- `9ba1626` - Added: Sample Env File
- `ecb19d3` - Added: WebSocket
- `f723594` - Added: Collection APIs
- `bd31510` - Added: Member APIs
- `2c122fd` - Implemented: APIs for User
- `5bb77a0` - Added: Codec (MsgPack abstraction)
- `554ecdc` - Greeted World (initial commit)
