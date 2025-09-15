# GoatSync Architecture

> **Version:** 2.0  
> **Go Version:** 1.25+  
> **Last Updated:** November 2025

---

## Overview

GoatSync is a Go implementation of the EteSync server, designed for **100% API compatibility** with existing EteSync clients (web, iOS, Android, etesync-dav).

### Design Principles

1. **Clean Architecture** - Separation of concerns with clear layer boundaries
2. **Dependency Injection** - Constructor injection, no global state
3. **Interface-Based Design** - Define interfaces where consumed, not where implemented
4. **Context Propagation** - Pass context through all layers for cancellation and timeouts
5. **Explicit Error Handling** - No panic-driven flow, structured error types

---

## Layer Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              PRESENTATION LAYER                              │
│                                                                              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │   Handler   │ │  Middleware │ │   Server    │ │  WebSocket  │           │
│  │  (auth.go)  │ │  (auth.go)  │ │ (server.go) │ │(websocket.go│           │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘           │
│         │               │               │               │                   │
│         └───────────────┴───────────────┴───────────────┘                   │
│                                   │                                          │
│                                   ▼                                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                              BUSINESS LAYER                                  │
│                                                                              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │    Auth     │ │ Collection  │ │   Member    │ │ Invitation  │           │
│  │  Service    │ │  Service    │ │  Service    │ │  Service    │           │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘           │
│         │               │               │               │                   │
│         └───────────────┴───────────────┴───────────────┘                   │
│                                   │                                          │
│                                   ▼                                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                              DATA ACCESS LAYER                               │
│                                                                              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │    User     │ │ Collection  │ │   Stoken    │ │   Redis     │           │
│  │    Repo     │ │    Repo     │ │    Repo     │ │   Client    │           │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘           │
│         │               │               │               │                   │
│         └───────────────┴───────────────┴───────────────┘                   │
│                                   │                                          │
│                                   ▼                                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                              INFRASTRUCTURE                                  │
│                                                                              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │ PostgreSQL  │ │    Redis    │ │ Filesystem  │ │   Config    │           │
│  │   (GORM)    │ │  (go-redis) │ │  (chunks)   │ │   (env)     │           │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Directory Structure

```
goatSync/
├── cmd/
│   └── goatsync/
│       └── main.go                 # Entry point, wire dependencies
│
├── internal/                       # Private application code
│   │
│   ├── config/                     # Configuration management
│   │   └── config.go               # Load from environment
│   │
│   ├── crypto/                     # Etebase cryptographic operations
│   │   └── etebase.go              # BLAKE2b, SecretBox, Ed25519
│   │
│   ├── database/                   # Database connection
│   │   └── postgres.go             # GORM setup and migrations
│   │
│   ├── model/                      # Domain models (GORM structs)
│   │   ├── user.go                 # User, UserInfo
│   │   ├── collection.go           # Collection, CollectionType
│   │   ├── item.go                 # CollectionItem
│   │   ├── revision.go             # CollectionItemRevision
│   │   ├── chunk.go                # CollectionItemChunk, RevisionChunkRelation
│   │   ├── member.go               # CollectionMember, CollectionMemberRemoved
│   │   ├── invitation.go           # CollectionInvitation
│   │   ├── stoken.go               # Stoken
│   │   └── token.go                # AuthToken
│   │
│   ├── repository/                 # Data access layer
│   │   ├── interfaces.go           # Repository interface definitions
│   │   ├── user.go                 # UserRepository implementation
│   │   ├── collection.go           # CollectionRepository implementation
│   │   ├── item.go                 # ItemRepository implementation
│   │   ├── member.go               # MemberRepository implementation
│   │   ├── invitation.go           # InvitationRepository implementation
│   │   ├── stoken.go               # StokenRepository (critical for sync)
│   │   └── token.go                # TokenRepository implementation
│   │
│   ├── service/                    # Business logic layer
│   │   ├── auth.go                 # AuthService
│   │   ├── collection.go           # CollectionService
│   │   ├── item.go                 # ItemService
│   │   ├── member.go               # MemberService
│   │   └── invitation.go           # InvitationService
│   │
│   ├── handler/                    # HTTP handlers (presentation)
│   │   ├── handler.go              # Base handler, shared utilities
│   │   ├── auth.go                 # AuthHandler
│   │   ├── collection.go           # CollectionHandler
│   │   ├── item.go                 # ItemHandler
│   │   ├── member.go               # MemberHandler
│   │   ├── invitation.go           # InvitationHandler
│   │   └── websocket.go            # WebSocketHandler
│   │
│   ├── middleware/                 # HTTP middleware
│   │   ├── auth.go                 # Token authentication
│   │   ├── cors.go                 # CORS handling
│   │   ├── logging.go              # Request/response logging
│   │   └── recovery.go             # Panic recovery
│   │
│   ├── server/                     # HTTP server setup
│   │   └── server.go               # Server struct, route registration
│   │
│   └── storage/                    # File storage
│       └── filesystem.go           # Chunk file operations
│
├── pkg/                            # Public packages (can be imported)
│   ├── msgpack/
│   │   └── msgpack.go              # MessagePack helpers
│   └── errors/
│       └── etebase.go              # Etebase error types
│
├── migrations/                     # SQL migrations
│   ├── 001_initial_schema.up.sql
│   └── 001_initial_schema.down.sql
│
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
│
├── .cursor/
│   └── rules/
│       └── goatsync.mdc            # Cursor IDE rules
│
├── .env.example                    # Environment variable template
├── go.mod
├── go.sum
├── README.md
├── ARCHITECTURE.md                 # This file
└── CHANGELOG.md
```

---

## Component Details

### 1. Entry Point (`cmd/goatsync/main.go`)

Responsibilities:
- Load configuration
- Initialize database connection
- Wire dependencies (dependency injection)
- Start HTTP server
- Handle graceful shutdown

```go
func main() {
    // 1. Load config
    cfg := config.Load()
    
    // 2. Initialize database
    db, err := database.Connect(cfg.DatabaseURL)
    if err != nil {
        log.Fatal(err)
    }
    
    // 3. Initialize Redis (optional)
    var redisClient *redis.Client
    if cfg.RedisURL != "" {
        redisClient = redis.NewClient(...)
    }
    
    // 4. Wire repositories
    userRepo := repository.NewUserRepository(db)
    tokenRepo := repository.NewTokenRepository(db)
    collectionRepo := repository.NewCollectionRepository(db)
    // ... more repos
    
    // 5. Wire services
    crypto := crypto.NewEtebaseCrypto(cfg.EncryptionSecret)
    authService := service.NewAuthService(userRepo, tokenRepo, crypto, cfg)
    collectionService := service.NewCollectionService(collectionRepo, ...)
    // ... more services
    
    // 6. Wire handlers
    authHandler := handler.NewAuthHandler(authService)
    collectionHandler := handler.NewCollectionHandler(collectionService)
    // ... more handlers
    
    // 7. Create and start server
    srv := server.New(cfg, authHandler, collectionHandler, ...)
    srv.Run()
}
```

### 2. Configuration (`internal/config/`)

Single source of truth for all configuration:

```go
type Config struct {
    // Server
    Port    string `env:"PORT" envDefault:"8080"`
    Debug   bool   `env:"DEBUG" envDefault:"false"`
    GinMode string `env:"GIN_MODE" envDefault:"release"`
    
    // Security
    EncryptionSecret string   `env:"ENCRYPTION_SECRET,required"`
    AllowedOrigins   []string `env:"ALLOWED_ORIGINS" envDefault:"*"`
    AllowedHosts     []string `env:"ALLOWED_HOSTS" envDefault:"*"`
    
    // Database
    DatabaseURL string `env:"DATABASE_URL,required"`
    
    // Redis (optional, for WebSocket)
    RedisURL string `env:"REDIS_URL"`
    
    // Storage
    ChunkStoragePath string `env:"CHUNK_STORAGE_PATH" envDefault:"./data/chunks"`
    
    // Challenge
    ChallengeValidSeconds int `env:"CHALLENGE_VALID_SECONDS" envDefault:"300"`
}
```

### 3. Models (`internal/model/`)

GORM models matching Django schema exactly:

```go
// user.go
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Username  string    `gorm:"uniqueIndex;size:150;not null"`
    Email     string    `gorm:"uniqueIndex;size:254;not null"`
    FirstName string    `gorm:"size:150"` // Used to store original username casing
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
    UserInfo  *UserInfo `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE"`
}

type UserInfo struct {
    OwnerID          uint   `gorm:"primaryKey"`
    Version          int    `gorm:"default:1"`
    LoginPubkey      []byte `gorm:"type:bytea;not null"`
    Pubkey           []byte `gorm:"type:bytea;not null"`
    EncryptedContent []byte `gorm:"type:bytea;not null"`
    Salt             []byte `gorm:"type:bytea;not null"`
}

// stoken.go
type Stoken struct {
    ID  uint   `gorm:"primaryKey"`
    UID string `gorm:"uniqueIndex;size:43;not null"`
}

// Before create hook to generate UID
func (s *Stoken) BeforeCreate(tx *gorm.DB) error {
    if s.UID == "" {
        s.UID = generateStokenUID()
    }
    return nil
}
```

### 4. Repository Layer (`internal/repository/`)

Interface definitions + implementations:

```go
// interfaces.go
type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    GetByUsername(ctx context.Context, username string) (*model.User, error)
    GetByID(ctx context.Context, id uint) (*model.User, error)
    Update(ctx context.Context, user *model.User) error
}

type StokenRepository interface {
    Create(ctx context.Context) (*model.Stoken, error)
    GetByUID(ctx context.Context, uid string) (*model.Stoken, error)
}

type CollectionRepository interface {
    Create(ctx context.Context, collection *model.Collection) error
    GetByUID(ctx context.Context, uid string) (*model.Collection, error)
    ListForUser(ctx context.Context, userID uint, stoken string, limit int) (
        collections []model.Collection,
        newStoken *model.Stoken,
        done bool,
        err error,
    )
    // ... more methods
}

// user.go - implementation
type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).
        Preload("UserInfo").
        Where("LOWER(username) = LOWER(?)", username).
        First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}
```

### 5. Service Layer (`internal/service/`)

Business logic, orchestration, crypto operations:

```go
// auth.go
type AuthService struct {
    userRepo   repository.UserRepository
    tokenRepo  repository.TokenRepository
    crypto     *crypto.EtebaseCrypto
    cfg        *config.Config
}

func NewAuthService(
    userRepo repository.UserRepository,
    tokenRepo repository.TokenRepository,
    crypto *crypto.EtebaseCrypto,
    cfg *config.Config,
) *AuthService {
    return &AuthService{
        userRepo:  userRepo,
        tokenRepo: tokenRepo,
        crypto:    crypto,
        cfg:       cfg,
    }
}

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
    challengeBytes, _ := msgpack.Marshal(challengeData)
    encryptedChallenge := s.crypto.Encrypt(encKey, challengeBytes)
    
    return &LoginChallengeResponse{
        Salt:      salt,
        Challenge: encryptedChallenge,
        Version:   user.UserInfo.Version,
    }, nil
}
```

### 6. Handler Layer (`internal/handler/`)

HTTP request/response handling:

```go
// handler.go - base
type Handler struct {
    // Shared utilities
}

func (h *Handler) respond(c *gin.Context, status int, data interface{}) {
    packed, err := msgpack.Marshal(data)
    if err != nil {
        c.AbortWithStatus(http.StatusInternalServerError)
        return
    }
    c.Data(status, "application/msgpack", packed)
}

func (h *Handler) parse(c *gin.Context, v interface{}) error {
    body, err := io.ReadAll(c.Request.Body)
    if err != nil {
        return err
    }
    return msgpack.Unmarshal(body, v)
}

func (h *Handler) handleError(c *gin.Context, err error) {
    var eteErr *errors.EtebaseError
    if errors.As(err, &eteErr) {
        h.respond(c, eteErr.StatusCode, eteErr)
        return
    }
    c.AbortWithStatus(http.StatusInternalServerError)
}

// auth.go
type AuthHandler struct {
    Handler
    authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
    return &AuthHandler{authService: authService}
}

func (h *AuthHandler) LoginChallenge(c *gin.Context) {
    var req LoginChallengeRequest
    if err := h.parse(c, &req); err != nil {
        h.handleError(c, err)
        return
    }
    
    ctx := c.Request.Context()
    resp, err := h.authService.LoginChallenge(ctx, req.Username)
    if err != nil {
        h.handleError(c, err)
        return
    }
    
    h.respond(c, http.StatusOK, resp)
}
```

### 7. Server (`internal/server/`)

HTTP server setup and route registration:

```go
type Server struct {
    cfg               *config.Config
    authHandler       *handler.AuthHandler
    collectionHandler *handler.CollectionHandler
    memberHandler     *handler.MemberHandler
    invitationHandler *handler.InvitationHandler
    websocketHandler  *handler.WebSocketHandler
    engine            *gin.Engine
}

func New(cfg *config.Config, handlers ...) *Server {
    if cfg.Debug {
        gin.SetMode(gin.DebugMode)
    } else {
        gin.SetMode(gin.ReleaseMode)
    }
    
    engine := gin.New()
    
    // Global middleware
    engine.Use(gin.Recovery())
    engine.Use(middleware.Logger())
    engine.Use(middleware.CORS(cfg.AllowedOrigins))
    
    srv := &Server{
        cfg:    cfg,
        engine: engine,
        // ... handlers
    }
    
    srv.registerRoutes()
    return srv
}

func (s *Server) registerRoutes() {
    api := s.engine.Group("/api/v1")
    
    // Authentication routes
    auth := api.Group("/authentication")
    {
        auth.GET("/is_etebase/", s.authHandler.IsEtebase)
        auth.POST("/login_challenge/", s.authHandler.LoginChallenge)
        auth.POST("/login/", s.authHandler.Login)
        auth.POST("/logout/", middleware.RequireAuth(), s.authHandler.Logout)
        auth.POST("/change_password/", middleware.RequireAuth(), s.authHandler.ChangePassword)
        auth.POST("/signup/", s.authHandler.Signup)
        auth.POST("/dashboard_url/", middleware.RequireAuth(), s.authHandler.DashboardURL)
    }
    
    // Collection routes
    collection := api.Group("/collection", middleware.RequireAuth())
    {
        collection.GET("/", s.collectionHandler.List)
        collection.POST("/", s.collectionHandler.Create)
        collection.POST("/list_multi/", s.collectionHandler.ListMulti)
        collection.GET("/:collection_uid/", s.collectionHandler.Get)
        
        // Item routes (nested under collection)
        items := collection.Group("/:collection_uid")
        {
            items.GET("/item/", s.collectionHandler.ListItems)
            // ... more item routes
        }
        
        // Member routes
        members := collection.Group("/:collection_uid", middleware.RequireCollectionAccess())
        {
            members.GET("/member/", middleware.RequireAdmin(), s.memberHandler.List)
            // ... more member routes
        }
    }
    
    // Invitation routes
    invitation := api.Group("/invitation", middleware.RequireAuth())
    {
        incoming := invitation.Group("/incoming")
        {
            incoming.GET("/", s.invitationHandler.ListIncoming)
            // ... more incoming routes
        }
        outgoing := invitation.Group("/outgoing")
        {
            outgoing.GET("/", s.invitationHandler.ListOutgoing)
            // ... more outgoing routes
        }
    }
    
    // WebSocket route
    ws := api.Group("/ws")
    {
        ws.GET("/:ticket/", s.websocketHandler.Handle)
    }
}

func (s *Server) Run() error {
    addr := ":" + s.cfg.Port
    log.Printf("Server starting on %s", addr)
    return s.engine.Run(addr)
}
```

---

## Data Flow Example: Login

```
1. Client sends POST /api/v1/authentication/login_challenge/
   Body: { username: "user@example.com" }

2. Router → middleware.Logger() → AuthHandler.LoginChallenge()

3. AuthHandler parses MessagePack body into LoginChallengeRequest

4. AuthHandler calls authService.LoginChallenge(ctx, "user@example.com")

5. AuthService:
   a. userRepo.GetByUsername(ctx, "user@example.com") → User with UserInfo
   b. crypto.GetEncryptionKey(userInfo.Salt) → encryption key
   c. Create challenge data: { timestamp, userId }
   d. crypto.Encrypt(key, msgpack(challengeData)) → encrypted challenge
   e. Return LoginChallengeResponse

6. AuthHandler serializes response to MessagePack

7. Client receives: { salt, challenge, version }
```

---

## Error Handling Strategy

```go
// pkg/errors/etebase.go
type EtebaseError struct {
    Code       string            `json:"code" msgpack:"code"`
    Detail     string            `json:"detail" msgpack:"detail"`
    Field      string            `json:"field,omitempty" msgpack:"field,omitempty"`
    Errors     []*EtebaseError   `json:"errors,omitempty" msgpack:"errors,omitempty"`
    StatusCode int               `json:"-" msgpack:"-"`
}

// Standard errors (must match Python exactly)
var (
    // Auth errors
    ErrUserNotFound      = &EtebaseError{Code: "user_not_found", Detail: "User not found", StatusCode: 401}
    ErrUserNotInit       = &EtebaseError{Code: "user_not_init", Detail: "User not properly init", StatusCode: 401}
    ErrBadSignature      = &EtebaseError{Code: "login_bad_signature", Detail: "Wrong password for user.", StatusCode: 401}
    ErrWrongAction       = &EtebaseError{Code: "wrong_action", Detail: "Expected different action", StatusCode: 400}
    ErrChallengeExpired  = &EtebaseError{Code: "challenge_expired", Detail: "Login challenge has expired", StatusCode: 400}
    ErrWrongUser         = &EtebaseError{Code: "wrong_user", Detail: "This challenge is for the wrong user", StatusCode: 400}
    ErrWrongHost         = &EtebaseError{Code: "wrong_host", Detail: "Found wrong host name", StatusCode: 400}
    ErrUserExists        = &EtebaseError{Code: "user_exists", Detail: "User already exists", StatusCode: 409}
    
    // Collection errors
    ErrBadStoken         = &EtebaseError{Code: "bad_stoken", Detail: "Invalid stoken.", StatusCode: 400}
    ErrStaleStoken       = &EtebaseError{Code: "stale_stoken", Detail: "Stoken is too old", StatusCode: 409}
    ErrWrongEtag         = &EtebaseError{Code: "wrong_etag", Detail: "Wrong etag", StatusCode: 409}
    ErrUniqueUID         = &EtebaseError{Code: "unique_uid", Detail: "Collection with this uid already exists", StatusCode: 409}
    
    // Permission errors
    ErrAdminRequired     = &EtebaseError{Code: "admin_access_required", Detail: "Only collection admins can perform this operation.", StatusCode: 403}
    ErrNoWriteAccess     = &EtebaseError{Code: "no_write_access", Detail: "You need write access to write to this collection", StatusCode: 403}
    
    // Chunk errors
    ErrChunkExists       = &EtebaseError{Code: "chunk_exists", Detail: "Chunk already exists.", StatusCode: 409}
    ErrChunkNoContent    = &EtebaseError{Code: "chunk_no_content", Detail: "Tried to create a new chunk without content", StatusCode: 400}
    
    // Invitation errors
    ErrNoSelfInvite      = &EtebaseError{Code: "no_self_invite", Detail: "Inviting yourself is not allowed", StatusCode: 400}
    ErrInvitationExists  = &EtebaseError{Code: "invitation_exists", Detail: "Invitation already exists", StatusCode: 409}
    
    // WebSocket errors
    ErrNotSupported      = &EtebaseError{Code: "not_supported", Detail: "This end-point requires Redis to be configured", StatusCode: 501}
)
```

---

## Testing Strategy

### Unit Tests
- Test services with mocked repositories
- Test crypto functions with known test vectors from Python
- Test handlers with mocked services

### Integration Tests
- Test repositories with test database (use transactions, rollback after each test)
- Test full request/response cycles

### Compatibility Tests
- Record Python server responses
- Compare Go server responses byte-for-byte
- Use existing EteSync clients as test harness

---

## Deployment

### Docker

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /goatsync ./cmd/goatsync

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /goatsync /goatsync
EXPOSE 8080
CMD ["/goatsync"]
```

### Environment Variables

```bash
# Required
DATABASE_URL=postgres://user:pass@localhost:5432/goatsync?sslmode=disable
ENCRYPTION_SECRET=your-secret-key-min-32-chars

# Optional
PORT=8080
DEBUG=false
GIN_MODE=release
REDIS_URL=redis://localhost:6379/0
CHUNK_STORAGE_PATH=/var/lib/goatsync/chunks
ALLOWED_ORIGINS=https://pim.etesync.com,https://notes.etesync.com
ALLOWED_HOSTS=sync.example.com
CHALLENGE_VALID_SECONDS=300
```

---

## References

- [EteSync Protocol](https://docs.etebase.com/)
- [Gin Web Framework](https://gin-gonic.com/)
- [GORM Documentation](https://gorm.io/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

