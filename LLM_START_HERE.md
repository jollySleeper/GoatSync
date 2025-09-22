# GoatSync: LLM Migration Guide

> **READ THIS FIRST** - Give this document to any LLM to start productive migration work

---

## TL;DR

**Goal:** Port EteSync server from Python to Go with 100% API compatibility.

**Status:** ~8-10% complete (scaffolding only, crypto is BROKEN)

**Priority Order:**
1. Fix crypto (BLOCKING - everything else depends on this)
2. Add PostgreSQL + GORM
3. Create GORM models
4. Implement stoken system
5. Wire handlers to database

---

## Quick Context

| What | Where |
|------|-------|
| Python source | `../eteSync-server/etebase_server/` |
| Go target | `./` (this repo) |
| Key Python files | `fastapi/routers/*.py`, `django/models.py` |
| Architecture doc | `./ARCHITECTURE.md` |
| Detailed status | `./MIGRATION_STATUS.md` |
| Cursor rules | `./.cursor/rules/` |

---

## The Three BLOCKING Issues

### 1. Crypto is WRONG (Must Fix First)

**File:** `internal/handlers/user.go`

**Problem:** Uses bcrypt. EteSync uses Ed25519 + BLAKE2b + SecretBox.

**Fix:** Create `internal/crypto/etebase.go` with:
- `GetEncryptionKey(secretKey string, salt []byte) []byte` - BLAKE2b with key/salt/person
- `Encrypt(key, plaintext []byte) []byte` - NaCl SecretBox
- `Decrypt(key, ciphertext []byte) ([]byte, error)` - NaCl SecretBox
- `VerifySignature(pubkey, message, signature []byte) bool` - Ed25519

**Reference:** `../eteSync-server/etebase_server/fastapi/routers/authentication.py` lines 129-173

### 2. No Database (Must Add)

**Problem:** Using in-memory maps that lose data on restart.

**Fix:**
1. Add GORM: `go get gorm.io/gorm gorm.io/driver/postgres`
2. Create `internal/database/postgres.go`
3. Create models in `internal/model/`

### 3. No Stoken System (Must Implement)

**Problem:** The sync token system doesn't exist.

**Fix:** See `06-stoken.mdc` in cursor rules.

---

## File Creation Order

```
Phase 1: Foundation
├── internal/crypto/etebase.go        # BLAKE2b, SecretBox, Ed25519
├── internal/database/postgres.go      # GORM connection
└── internal/model/stoken.go          # Stoken model (others depend on it)

Phase 2: Models
├── internal/model/user.go            # Update existing with GORM tags
├── internal/model/collection.go      # Collection, CollectionType
├── internal/model/item.go            # CollectionItem
├── internal/model/revision.go        # CollectionItemRevision
├── internal/model/chunk.go           # CollectionItemChunk, RevisionChunkRelation
├── internal/model/member.go          # CollectionMember, CollectionMemberRemoved
├── internal/model/invitation.go      # CollectionInvitation
└── internal/model/token.go           # Update AuthToken

Phase 3: Repository
├── internal/repository/interfaces.go  # All interfaces
├── internal/repository/user.go        # Rewrite with GORM
├── internal/repository/stoken.go      # Critical for sync
├── internal/repository/collection.go
├── internal/repository/item.go
├── internal/repository/member.go
└── internal/repository/invitation.go

Phase 4: Service
├── internal/service/auth.go          # Auth business logic
├── internal/service/collection.go
├── internal/service/item.go
├── internal/service/member.go
└── internal/service/invitation.go

Phase 5: Handler Updates
├── internal/handler/auth.go          # Fix to use service
├── internal/handler/collection.go
├── internal/handler/item.go
├── internal/handler/member.go
├── internal/handler/invitation.go    # New file
└── internal/handler/websocket.go
```

---

## API Endpoints to Implement

**Fully working:** None (crypto is broken)

**Scaffolded but broken:**
- `GET /api/v1/authentication/is_etebase/` ✅ Works
- `POST /api/v1/authentication/login_challenge/` ❌ Wrong crypto
- `POST /api/v1/authentication/login/` ❌ Wrong crypto
- `POST /api/v1/authentication/signup/` ❌ No DB
- All collection/item/member endpoints ❌ No DB

**Not started:**
- All invitation endpoints
- WebSocket endpoint

---

## When Implementing a Feature

1. **Read the Python source first** - The behavior must match exactly
2. **Check cursor rules** - `.cursor/rules/` has patterns and requirements
3. **Use MessagePack** - NOT JSON (except for debugging)
4. **Match error codes** - Python error codes must be identical
5. **Test against Python** - Use existing etesync clients to verify

---

## Example: Implementing Login Challenge

### Step 1: Read Python (`../eteSync-server/.../authentication.py`)

```python
@authentication_router.post("/login_challenge/", response_model=LoginChallengeOut)
def login_challenge(user: UserType = Depends(get_login_user)):
    salt = bytes(user.userinfo.salt)
    enc_key = get_encryption_key(salt)
    box = nacl.secret.SecretBox(enc_key)
    challenge_data = {
        "timestamp": int(datetime.now().timestamp()),
        "userId": user.id,
    }
    challenge = bytes(box.encrypt(msgpack_encode(challenge_data), encoder=nacl.encoding.RawEncoder))
    return MsgpackResponse(LoginChallengeOut(salt=salt, challenge=challenge, version=user.userinfo.version))
```

### Step 2: Implement in Go

```go
// internal/service/auth.go
func (s *AuthService) LoginChallenge(ctx context.Context, username string) (*LoginChallengeResponse, error) {
    user, err := s.userRepo.GetByUsername(ctx, username)
    if err != nil {
        return nil, err
    }
    if user == nil || user.UserInfo == nil {
        return nil, errors.ErrUserNotFound
    }
    
    salt := user.UserInfo.Salt
    encKey := s.crypto.GetEncryptionKey(salt)
    
    challengeData := map[string]interface{}{
        "timestamp": time.Now().Unix(),
        "userId":    user.ID,
    }
    encoded, _ := msgpack.Marshal(challengeData)
    encrypted := s.crypto.Encrypt(encKey, encoded)
    
    return &LoginChallengeResponse{
        Salt:      salt,
        Challenge: encrypted,
        Version:   user.UserInfo.Version,
    }, nil
}
```

### Step 3: Wire to Handler

```go
// internal/handler/auth.go
func (h *AuthHandler) LoginChallenge(c *gin.Context) {
    var req LoginChallengeRequest
    if err := h.parseMsgpack(c, &req); err != nil {
        h.handleError(c, err)
        return
    }
    
    resp, err := h.authService.LoginChallenge(c.Request.Context(), req.Username)
    if err != nil {
        h.handleError(c, err)
        return
    }
    
    h.respondMsgpack(c, http.StatusOK, resp)
}
```

---

## Quick Commands

```bash
# Build and verify
go build ./cmd/server

# Run server
go run ./cmd/server

# Test
go test ./...

# Check for race conditions
go test -race ./...

# Add a dependency
go get -u <package>
go mod tidy
```

---

## Reference Documents

| Document | Purpose |
|----------|---------|
| `ARCHITECTURE.md` | Layer structure, component details |
| `MIGRATION_STATUS.md` | Detailed progress, effort estimates |
| `DEPENDENCIES.md` | Required packages |
| `.cursor/rules/01-project.mdc` | Project context |
| `.cursor/rules/04-crypto.mdc` | Crypto implementation details |
| `.cursor/rules/05-models.mdc` | GORM model definitions |
| `.cursor/rules/06-stoken.mdc` | Sync token system |

---

## When You're Stuck

1. **Check the Python source** - It's the authoritative reference
2. **Check cursor rules** - Specific patterns and requirements
3. **Look at existing Go code** - For patterns already used
4. **Search for the error code** - In both Python and Go

---

## Success Criteria

A feature is DONE when:
- [ ] Build passes (`go build ./...`)
- [ ] Tests pass (`go test ./...`)
- [ ] No race conditions (`go test -race ./...`)
- [ ] Response matches Python byte-for-byte (for MessagePack)
- [ ] Error codes match Python exactly
- [ ] Works with existing etesync-dav client

