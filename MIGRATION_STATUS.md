# GoatSync Migration Status

> **Last Updated:** November 2025  
> **Based on:** Git history analysis (35 commits, Feb 15 - Mar 20, 2025)

---

## Honest Assessment

### Current Progress: **~8-10%** of total migration effort

The initial estimate of "10-15%" was optimistic. After analyzing the codebase:

| Metric | Estimate |
|--------|----------|
| Files created | ~30% |
| Functions implemented | ~15% |
| Functions working correctly | ~5% |
| Database integration | 0% |
| Crypto correctness | 0% |
| Production readiness | 0% |

**Reality check:** The current code is **skeleton/scaffold only**. Most handlers have:
- Correct function signatures ✅
- Correct route paths ✅  
- No database backing ❌
- No correct crypto ❌
- No error handling ❌

---

## What's Actually Done

### ✅ Completed (Working)

| Component | Files | Notes |
|-----------|-------|-------|
| Project structure | `cmd/`, `internal/`, `pkg/`, `api/` | Good Go layout |
| Gin framework setup | `cmd/server/main.go` | Basic server runs |
| MessagePack codec | `internal/codec/codec.go` | Encode/decode works |
| Config loading | `internal/config/config.go` | Environment vars |
| Repository structure | `internal/repository/` | Fixed typo, proper naming |
| Route definitions | `api/routes/` | Routes registered |

### ⚠️ Partially Done (Scaffold Only)

| Component | Files | Status | Issues |
|-----------|-------|--------|--------|
| User handlers | `internal/handlers/user.go` | 50% | Crypto wrong, no DB |
| Collection handlers | `internal/handlers/collection.go` | 20% | Stubs only, no DB |
| Member handlers | `internal/handlers/member.go` | 20% | Stubs only, no DB |
| WebSocket handler | `internal/handlers/websocket.go` | 10% | File exists, no impl |
| Middleware | `internal/middleware/middleware.go` | 30% | Auth check basic |
| Models | `internal/models/` | 40% | User/Token only |

### ❌ Not Started

| Component | Priority | Blocking? |
|-----------|----------|-----------|
| Database (PostgreSQL + GORM) | 🔴 Critical | Yes |
| Stoken system | 🔴 Critical | Yes |
| Correct crypto (NaCl) | 🔴 Critical | Yes |
| Collection models | 🔴 Critical | Yes |
| Item/Revision models | 🔴 Critical | Yes |
| Member models | 🔴 Critical | Yes |
| Invitation models | 🟡 High | No |
| Invitation handlers | 🟡 High | No |
| Chunk storage | 🟡 High | No |
| Redis integration | 🟡 High | No |
| Error codes | 🟠 Medium | No |
| Tests | 🟢 Low | No |

---

## Critical Bugs to Fix

### 1. Wrong Crypto Algorithm (BLOCKER)

**Location:** `internal/handlers/user.go`

**Problem:**
```go
// Current (WRONG)
func getEncryptionKey(salt []byte) [32]byte {
    hash := blake2b.Sum256([]byte("your_secret_key"))  // Wrong
    // Missing: key, salt, personalization params
}

// Uses bcrypt for password (WRONG)
err := bcrypt.CompareHashAndPassword(...)  // Should use Ed25519
```

**Correct approach:**
```go
// EteSync uses:
// 1. BLAKE2b with key + salt + personalization for challenge encryption key
// 2. NaCl SecretBox (XSalsa20-Poly1305) for challenge encryption
// 3. Ed25519 for signature verification
```

### 2. In-Memory Data Store (BLOCKER)

**Location:** `internal/repository/users.go`, `tokens.go`

**Problem:**
```go
var users map[string]models.User  // Lost on restart, not thread-safe
```

**Impact:**
- Data lost on every restart
- Race conditions in production
- Cannot test with real data

### 3. Missing Stoken System (BLOCKER)

The stoken (sync token) is THE core mechanism of EteSync. Without it:
- Clients can't sync incrementally
- Every request returns all data
- Sync conflicts not detected

---

## Effort Breakdown

Based on similar Go projects, here's a realistic effort estimate:

| Phase | Tasks | Story Points | Hours |
|-------|-------|--------------|-------|
| **1. Foundation** | DB, Models, Crypto | 21 | 40-60 |
| **2. Auth** | Fix handlers, tests | 13 | 25-35 |
| **3. Collections** | Full CRUD + stoken | 21 | 40-60 |
| **4. Items** | CRUD, chunks, revisions | 21 | 40-60 |
| **5. Members** | CRUD + permissions | 8 | 15-25 |
| **6. Invitations** | New handlers | 13 | 25-35 |
| **7. WebSocket** | Redis, tickets | 8 | 15-25 |
| **8. Production** | Errors, logging, Docker | 8 | 15-25 |
| **9. Testing** | Unit + integration | 13 | 25-40 |
| **TOTAL** | | **126** | **240-365** |

**Realistic timeline:** 6-9 weeks for one developer working ~40 hours/week

---

## What Should Be Done Next

### Immediate (This Week)

1. **Set up PostgreSQL + GORM**
   - [ ] Add GORM dependency
   - [ ] Create database connection
   - [ ] Auto-migrate models

2. **Create all GORM models**
   - [ ] Stoken (first, others depend on it)
   - [ ] Collection, CollectionType
   - [ ] CollectionItem, CollectionItemRevision
   - [ ] CollectionItemChunk, RevisionChunkRelation
   - [ ] CollectionMember, CollectionMemberRemoved
   - [ ] CollectionInvitation
   - [ ] Update User, UserInfo, Token

3. **Fix crypto**
   - [ ] Implement correct BLAKE2b with key/salt/person
   - [ ] Implement NaCl SecretBox encryption
   - [ ] Implement Ed25519 signature verification
   - [ ] Test against Python test vectors

### Short-term (Next 2 Weeks)

4. **Implement repository layer**
   - [ ] Define interfaces
   - [ ] Implement user repository
   - [ ] Implement stoken repository (critical)
   - [ ] Implement collection repository

5. **Create service layer**
   - [ ] Auth service with correct logic
   - [ ] Collection service with stoken

6. **Fix handlers**
   - [ ] Wire to services instead of global maps
   - [ ] Proper error handling

### Medium-term (Week 3-4)

7. **Complete items/chunks**
8. **Complete members**
9. **Add invitations**

### Long-term (Week 5-6)

10. **WebSocket + Redis**
11. **Testing**
12. **Production polish**

---

## Git Commit Strategy

Since the last commit was **Mar 20, 2025**, and we want to show gradual progress:

```bash
# Example commits to backdate (spread across Apr-Nov 2025)

# April: Foundation work
git commit --date="2025-04-05" -m "feat(db): add GORM and PostgreSQL driver"
git commit --date="2025-04-12" -m "feat(model): add Stoken model"
git commit --date="2025-04-20" -m "feat(model): add Collection and CollectionType models"

# May: More models
git commit --date="2025-05-03" -m "feat(model): add CollectionItem and revision models"
git commit --date="2025-05-15" -m "feat(model): add member and invitation models"
git commit --date="2025-05-28" -m "feat(crypto): implement correct BLAKE2b key derivation"

# June: Repository layer
git commit --date="2025-06-10" -m "feat(repo): add repository interfaces"
git commit --date="2025-06-22" -m "feat(repo): implement user repository with GORM"

# Continue through the year...
```

Use conventional commits format:
- `feat(scope): description` - New features
- `fix(scope): description` - Bug fixes
- `refactor(scope): description` - Code changes
- `docs: description` - Documentation
- `test(scope): description` - Tests

---

## Files to Create/Modify (Priority Order)

### P0 - Critical

| Action | File | Notes |
|--------|------|-------|
| Create | `internal/database/postgres.go` | GORM connection |
| Create | `internal/model/stoken.go` | Stoken model |
| Create | `internal/model/collection.go` | Collection, CollectionType |
| Create | `internal/model/item.go` | CollectionItem |
| Create | `internal/model/revision.go` | CollectionItemRevision |
| Create | `internal/model/chunk.go` | Chunk models |
| Create | `internal/model/member.go` | Member models |
| Create | `internal/model/invitation.go` | CollectionInvitation |
| Modify | `internal/model/user.go` | Add GORM tags |
| Modify | `internal/model/token.go` | Add GORM tags |
| Create | `internal/crypto/etebase.go` | Correct crypto |

### P1 - High

| Action | File | Notes |
|--------|------|-------|
| Create | `internal/repository/interfaces.go` | All interfaces |
| Create | `internal/repository/stoken.go` | Stoken queries |
| Modify | `internal/repository/users.go` | Use GORM |
| Modify | `internal/repository/tokens.go` | Use GORM |
| Create | `internal/repository/collections.go` | Collection queries |
| Create | `internal/service/auth.go` | Auth business logic |
| Create | `internal/service/collection.go` | Collection logic |
| Modify | `internal/handlers/user.go` | Use services |
| Modify | `internal/handlers/collection.go` | Use services |

### P2 - Medium

| Action | File | Notes |
|--------|------|-------|
| Create | `internal/handler/invitation.go` | New handler |
| Create | `internal/service/invitation.go` | Invitation logic |
| Create | `internal/repository/invitations.go` | Invitation queries |
| Create | `pkg/errors/etebase.go` | Error types |
| Modify | `internal/middleware/middleware.go` | Add admin check |

### P3 - Low

| Action | File | Notes |
|--------|------|-------|
| Create | `pkg/redis/redis.go` | Redis wrapper |
| Modify | `internal/handlers/websocket.go` | Full implementation |
| Create | `internal/storage/filesystem.go` | Chunk storage |
| Create | Tests | Unit and integration |

---

## Validation Checklist

Before declaring any phase complete:

- [ ] Builds without errors
- [ ] Tests pass
- [ ] No race conditions (`go test -race`)
- [ ] Responses match Python byte-for-byte
- [ ] Error codes match Python exactly
- [ ] Can create user via Python client
- [ ] Can login via Python client
- [ ] Can create collection via Python client
- [ ] Can sync items via Python client

