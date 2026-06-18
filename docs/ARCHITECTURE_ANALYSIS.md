# Vyzorix Update Server - Architecture Analysis & Restructuring Guide

**Date:** 2026-06-17  
**Status:** Analysis Complete - Restructuring Recommended

---

## Table of Contents

1. [Current Architecture Analysis](#current-architecture-analysis)
2. [Current Problems](#current-problems)
3. [Proposed Clean Architecture](#proposed-clean-architecture)
4. [Layer Responsibilities](#layer-responsibilities)
5. [Dependency Flow](#dependency-flow)
6. [File Organization](#file-organization)
7. [Where to Find Things](#where-to-find-things)
8. [Migration Path](#migration-path)

---

## Current Architecture Analysis

### How the Backend Currently Works

```
┌─────────────────────────────────────────────────────────────────────┐
│                           main.go                                    │
│                   (Entry point, wires dependencies)                   │
└─────────────────────────────────┬───────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      handlers/server.go                               │
│              (Server struct, holds ALL dependencies)                 │
│         ┌──────────────────────────────────────────────────┐       │
│         │ Notifier, Store, Hub, Limiter, AuthLimiter,       │       │
│         │ jwtCtrl, mfaCtrl, HMAC, Config, CSRFProtector,   │       │
│         │ Lockout, Turnstile, Signer, RevocationList...    │       │
│         └──────────────────────────────────────────────────┘       │
└─────────────────────────────────┬───────────────────────────────────┘
                                  │
                    ┌─────────────┴─────────────┐
                    ▼                           ▼
         ┌──────────────────┐        ┌──────────────────┐
         │   middleware/     │        │   handlers/       │
         │   - ratelimit    │        │   - auth_core    │
         │   - turnstile    │        │   - auth_oauth    │
         │   - lockout      │        │   - auth_mfa      │
         │   - csrf         │        │   - device        │
         │   - cors         │        │   - command       │
         └──────────────────┘        └────────┬─────────┘
                                              │
                                              ▼
                              ┌───────────────────────────┐
                              │     pkg/storage/           │
                              │  (direct DB access)        │
                              │  operators.go, devices.go, │
                              │  clients.go, crypto.go...  │
                              └───────────────────────────┘
```

### Current Request Flow: POST /api/v1/auth/login

```
Step 1: main.go
  ├─ Creates storage.Store
  ├─ Creates handlers.Server (passes store, notifier, hub)
  └─ Starts HTTP server

Step 2: handlers/server.go
  ├─ Server.New() initializes:
  │   ├─ AuthController (holds session manager, email service)
  │   ├─ MFAHandler
  │   ├─ RateLimiters (2 instances)
  │   ├─ TurnstileVerifier
  │   ├─ Lockout middleware
  │   ├─ CSRFProtector
  │   ├─ SignatureVerifier
  │   └─ HMAC verifier
  └─ Server.Routes() registers routes with middleware

Step 3: Middleware Chain (per-request)
  ├─ GinPanicRecovery
  ├─ RequestIDMiddleware
  ├─ Logger
  ├─ CORSHandler
  ├─ SecurityHeaders
  ├─ BodySizeLimit
  ├─ AuthLimiter (5 req/min) ──────────────────┐
  ├─ TurnstileMiddleware ───────────────────────┤ Per-route
  └─ LockoutMiddleware ────────────────────────┘

Step 4: handlers/auth_core.go - Login()
  ├─ Parse JSON request
  ├─ storage.GetOperatorByEmail()
  ├─ storage.VerifyPassword() (Argon2id)
  ├─ auth.CreateSessionCookieWithExpiry() (AES-256-GCM)
  └─ Return response

Step 5: pkg/storage/operators.go
  ├─ SQL query: SELECT * FROM operators WHERE email = ?
  └─ Argon2id comparison in crypto.go
```

### Current Directory Structure

```
apps/api/
├── main.go                           # Entry point
├── internal/
│   ├── api/
│   │   ├── handlers/                # ALL handlers mixed together
│   │   │   ├── server.go           # Route setup + Server struct
│   │   │   ├── auth_core.go       # Login, Register, Logout, Me
│   │   │   ├── auth_oauth.go      # Google, GitHub OAuth
│   │   │   ├── auth_mfa.go        # TOTP, Backup codes
│   │   │   ├── auth_password_reset.go
│   │   │   ├── auth_email_verify.go
│   │   │   ├── auth_settings.go
│   │   │   ├── auth_admin.go
│   │   │   ├── auth_rate_limit.go
│   │   │   ├── auth_utils.go
│   │   │   ├── auth_csrf.go
│   │   │   ├── device.go
│   │   │   ├── command.go
│   │   │   ├── updater.go
│   │   │   ├── websocket_handler.go
│   │   │   ├── health.go
│   │   │   └── lockout.go
│   │   │
│   │   └── middleware/              # All middleware
│   │       ├── rate_limiter.go
│   │       ├── turnstile.go
│   │       ├── lockout.go
│   │       ├── csrf.go
│   │       ├── cors.go
│   │       ├── security_headers.go
│   │       ├── request_signing.go
│   │       ├── replay_protection.go
│   │       ├── user_enum.go
│   │       ├── user_enum_block.go
│   │       ├── auth_enum_safe.go
│   │       ├── response_encryption.go
│   │       ├── recovery.go
│   │       └── ...
│   │
│   ├── auth/                        # Auth utilities
│   │   ├── session.go              # Session cookies
│   │   ├── password.go             # Password validation
│   │   ├── totp.go                # TOTP generation
│   │   └── ...
│   │
│   ├── command_signer.go           # Command signing (standalone file!)
│   ├── email.go                    # Email service
│   ├── fcm/                        # FCM integration
│   ├── ws/                         # WebSocket hub
│   ├── audit/                      # Audit logging
│   ├── metrics/                    # Metrics
│   └── ssr/                        # SSR server
│
└── pkg/
    ├── storage/                     # Database layer
    │   ├── store.go
    │   ├── operators.go
    │   ├── devices.go
    │   ├── clients.go
    │   ├── commands.go
    │   ├── sessions.go
    │   ├── settings.go
    │   ├── migrations.go
    │   ├── crypto.go
    │   ├── uuid.go
    │   └── telemetry.go
    ├── crypto/                     # Crypto utilities
    ├── config/                     # Configuration
    ├── logging/                    # Logging
    └── models/                     # Data models
```

---

## Current Problems

### 1. No Clear Layering

```
Problem: Everything talks to everything
─────────────────────────────────────
handlers/auth_core.go
  ├─ imports storage (pkg)
  ├─ imports auth/session (internal)
  ├─ imports auth/password (internal)
  └─ imports models (pkg)

handlers/device.go
  ├─ imports storage (pkg)
  ├─ imports crypto (pkg)
  └─ imports models (pkg)

middleware/rate_limiter.go
  └─ standalone, no dependencies

internal/command_signer.go
  └─ standalone file in internal/, not in any package!
```

### 2. Server Struct Knows Everything

```go
// Current: Server holds ALL dependencies
type Server struct {
    Notifier        fcm.Notifier
    Log             *slog.Logger
    Store           *storage.Store      // Database
    Hub             *hub.Hub            // WebSocket
    Limiter         *middleware.RateLimiter
    AuthLimiter     *middleware.RateLimiter
    jwtCtrl         *AuthController
    mfaCtrl         *MFAHandler
    originValidator *security.OriginValidator
    HMAC            hmac.Verifier
    Config          config.Config
    CSRFProtector   *middleware.CSRFProtector
    Lockout         *middleware.Lockout
    RevocationList  *security.RevocationList
    Turnstile       *middleware.TurnstileVerifier
    Signer          *middleware.SignatureVerifier
    // 17 dependencies!
}
```

### 3. Handlers Do Too Much

```go
// Current: Handler does HTTP + business logic + storage access
func (ac *AuthController) Login(c *gin.Context) {
    // 1. HTTP parsing
    var req models.LoginRequest
    json.NewDecoder(c.Request.Body).Decode(&req)
    
    // 2. Business logic (should be in service layer)
    if req.Email == "" { /* validation */ }
    
    // 3. Database access (should be in repository)
    op, err := ac.store.GetOperatorByEmail(ctx, req.Email)
    
    // 4. Password verification (should be in domain)
    err = storage.VerifyPassword(req.Password, op.PasswordHash)
    
    // 5. Session creation (should be in auth service)
    cookie, err := ac.session.CreateSessionCookieWithExpiry(...)
    
    // 6. HTTP response (this is OK in handler)
    http.SetCookie(c.Writer, cookie)
    c.JSON(200, op.ToResponse())
}
```

### 4. Storage is God Object

```go
// Current: Store has 50+ methods, does everything
type Store struct {
    db   *sql.DB
    path string
    mu   sync.Mutex
}

func (s *Store) GetOperatorByEmail(...)
func (s *Store) GetOperatorByID(...)
func (s *Store) CreateOperator(...)
func (s *Store) UpdateOperatorName(...)
func (s *Store) UpdateOperatorEmail(...)
func (s *Store) UpdateOperatorPassword(...)
func (s *Store) GetDeviceByID(...)
func (s *Store) RegisterDevice(...)
func (s *Store) UpdateDeviceFCMToken(...)
// ... 50+ more methods
```

### 5. Duplicated Patterns Everywhere

| Pattern | Occurrences | Locations |
|---------|-------------|-----------|
| Token generation | 3 | auth_core, auth_email_verify, auth_password_reset |
| OAuth operator creation | 2 | auth_oauth.go (Google, GitHub) |
| JSON unmarshal | 3 | clients.go (3 functions) |
| Argon2id password hash | 2 | crypto.go, storage/crypto.go |
| AES-256-GCM encrypt | 4 | request_signing.go (3), response_encryption.go |
| Constant-time compare | 3 | user_enum.go, user_enum_block.go, auth_enum_safe.go |

---

## Proposed Clean Architecture

### The Hexagonal / Layered Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          ENTRY POINT                                     │
│                         cmd/api/main.go                                   │
└─────────────────────────────────┬───────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                      PRESENTATION LAYER                                   │
│                   internal/api/handlers/                                  │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐        │
│  │  auth.go   │ │  device.go │ │ command.go │ │  health.go │        │
│  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘        │
│         │               │               │               │               │
│         └───────────────┴───────────────┴───────────────┘               │
│                               │                                           │
│                    Responsibility: HTTP ONLY                             │
│                    - Parse request                                       │
│                    - Validate input                                      │
│                    - Call application service                            │
│                    - Format response                                     │
└───────────────────────────────┬─────────────────────────────────────────┘
                                │
                                │ depends on
                                ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                      APPLICATION LAYER                                   │
│                    internal/application/                                  │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐           │
│  │ auth/           │ │ device/         │ │ command/        │           │
│  │ ├── login.go   │ │ ├── register.go │ │ ├── execute.go │           │
│  │ │ logout.go    │ │ │ update.go     │ │ ├── status.go  │           │
│  │ ├── register.go│ │ └── usecases.go│ │ └── usecases.go│           │
│  │ └── usecases.go│ │                 │ │                 │           │
│  └────────┬────────┘ └────────┬────────┘ └────────┬────────┘           │
│           │                  │                  │                     │
│           └──────────────────┴──────────────────┘                     │
│                               │                                         │
│                    Responsibility: USE CASES                            │
│                    - Orchestrate domain services                        │
│                    - Transaction management                             │
│                    - Input validation                                   │
└───────────────────────────────┬─────────────────────────────────────────┘
                                │
                                │ depends on
                                ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        DOMAIN LAYER                                      │
│                      internal/domain/                                     │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐        │
│  │  operator/  │ │   device/   │ │  session/   │ │  command/   │        │
│  │ ├── entity.go│ │ ├── entity.go│ │ ├── entity.go│ │ ├── entity.go│        │
│  │ └── repo.go │ │ └── repo.go │ │ └── repo.go │ │ └── repo.go │        │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘        │
│                               │                                           │
│                    Responsibility: PURE BUSINESS LOGIC                    │
│                    - No external dependencies                            │
│                    - Entity definitions                                  │
│                    - Repository interfaces                               │
│                    - Domain services (interfaces)                        │
└───────────────────────────────┬─────────────────────────────────────────┘
                                │
                                │ implemented by
                                ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                     INFRASTRUCTURE LAYER                                │
│                   internal/infrastructure/                               │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐           │
│  │    storage/     │ │    crypto/      │ │   external/     │           │
│  │ ├── sqlite.go   │ │ ├── argon2.go   │ │ ├── fcm.go      │           │
│  │ ├── operator.go │ │ ├── aes_gcm.go  │ │ ├── email.go    │           │
│  │ ├── device.go   │ │ └── hmac.go     │ │ └── turnstile.go│           │
│  │ └── session.go  │ │                 │ │                 │           │
│  └─────────────────┘ └─────────────────┘ └─────────────────┘           │
│                               │                                           │
│                    Responsibility: EXTERNAL INTEGRATIONS                 │
│                    - Database implementation                             │
│                    - Cryptographic operations                           │
│                    - Third-party services                               │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Layer Responsibilities

### Layer 1: Presentation (HTTP Handlers)

**Location:** `internal/api/handlers/`

**What it does:**
- Receives HTTP requests
- Parses JSON/body
- Validates request format
- Calls application layer
- Formats and returns HTTP responses

**What it DOES NOT do:**
- Business logic
- Database access
- Password hashing
- Session management directly

**Example:**
```go
// internal/api/handlers/auth/login.go
type LoginHandler struct {
    authService *application.AuthService  // Depends on application layer
}

func (h *LoginHandler) Login(c *gin.Context) {
    // 1. Parse request (HTTP concern)
    var reqDTO LoginRequest
    if err := c.ShouldBindJSON(&reqDTO); err != nil {
        c.JSON(400, ErrorResponse{Error: "invalid_request"})
        return
    }
    
    // 2. Call application layer (orchestration)
    result, err := h.authService.Login(c.Request.Context(), reqDTO.Email, reqDTO.Password)
    if err != nil {
        // 3. Handle errors and return response
        handleAuthError(c, err)
        return
    }
    
    // 4. Return response
    setSessionCookie(c, result.SessionToken)
    c.JSON(200, result.Operator)
}
```

### Layer 2: Application (Use Cases)

**Location:** `internal/application/`

**What it does:**
- Implements use cases (login, register, etc.)
- Orchestrates domain services
- Handles transactions
- Input validation beyond format

**What it DOES NOT do:**
- HTTP handling
- SQL queries
- Encryption details

**Example:**
```go
// internal/application/auth/login.go
type AuthService struct {
    operatorRepo domain.OperatorRepository    // Interface, not implementation
    passwordHasher crypto.PasswordHasher       // Interface
    sessionManager auth.SessionManager         // Interface
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
    // 1. Validation
    if err := validateEmail(email); err != nil {
        return nil, ErrInvalidEmail
    }
    
    // 2. Find operator (via repository interface)
    operator, err := s.operatorRepo.FindByEmail(ctx, email)
    if err != nil {
        if errors.Is(err, domain.ErrNotFound) {
            // Return same error for security (no enumeration)
            return nil, ErrInvalidCredentials
        }
        return nil, err
    }
    
    // 3. Verify password (domain logic via interface)
    if err := s.passwordHasher.Verify(password, operator.PasswordHash); err != nil {
        return nil, ErrInvalidCredentials
    }
    
    // 4. Create session
    session, err := s.sessionManager.Create(operator.ID)
    if err != nil {
        return nil, err
    }
    
    return &LoginResult{
        Operator: operator,
        Session:  session,
    }, nil
}
```

### Layer 3: Domain (Business Logic)

**Location:** `internal/domain/`

**What it does:**
- Defines entities (Operator, Device, Session)
- Defines repository interfaces
- Pure business rules
- No external dependencies

**What it DOES NOT do:**
- HTTP handling
- Database queries
- Encryption

**Example:**
```go
// internal/domain/operator/entity.go
type Operator struct {
    ID           string
    Email        string
    Name         string
    PasswordHash string
    Role         Role
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// internal/domain/operator/repository.go
type Repository interface {
    FindByID(ctx context.Context, id string) (*Operator, error)
    FindByEmail(ctx context.Context, email string) (*Operator, error)
    Create(ctx context.Context, op *Operator) error
    Update(ctx context.Context, op *Operator) error
    Delete(ctx context.Context, id string) error
}
```

### Layer 4: Infrastructure (Implementations)

**Location:** `internal/infrastructure/`

**What it does:**
- Implements domain interfaces
- Database access (SQLite)
- Cryptographic operations
- External service clients

**What it DOES NOT do:**
- Business logic
- HTTP handling

**Example:**
```go
// internal/infrastructure/storage/sqlite/operator.go
type SQLiteOperatorRepository struct {
    db *sql.DB
}

func (r *SQLiteOperatorRepository) FindByEmail(ctx context.Context, email string) (*domain.Operator, error) {
    query := `SELECT id, email, name, password_hash, role, created_at, updated_at 
              FROM operators WHERE email = ?`
    
    var op domain.Operator
    err := r.db.QueryRowContext(ctx, query, email).Scan(
        &op.ID, &op.Email, &op.Name, &op.PasswordHash, &op.Role,
        &op.CreatedAt, &op.UpdatedAt,
    )
    if errors.Is(err, sql.ErrNoRows) {
        return nil, domain.ErrNotFound
    }
    return &op, err
}

// Satisfies domain.OperatorRepository interface
var _ domain.OperatorRepository = (*SQLiteOperatorRepository)(nil)
```

---

## Dependency Flow

### Current (Problematic)

```
┌─────────┐     ┌──────────┐     ┌─────────┐
│ Handler │ ──► │ Service  │ ──► │ Storage │
└─────────┘     └──────────┘     └─────────┘
    │               │
    └───────────────┴─────── All mixed together
```

### Proposed (Clean)

```
┌─────────────────────────────────────────────────────────────┐
│                     DEPENDENCIES FLOW                        │
│                                                             │
│   ┌─────────┐      ┌─────────────┐      ┌─────────────┐   │
│   │Handler  │ ──► │ Application │ ──► │   Domain    │   │
│   └─────────┘      │   Service   │      │ (Interfaces)│   │
│                   └─────────────┘      └──────┬──────┘   │
│                                                │           │
│                   ┌───────────────────────────┘           │
│                   │ implements                            │
│                   ▼                                      │
│            ┌─────────────┐                               │
│            │Infrastructure│                               │
│            │ (Storage,    │                               │
│            │  Crypto,     │                               │
│            │  External)   │                               │
│            └─────────────┘                               │
│                                                             │
│   RULE: Dependencies point INWARD                          │
│         Inner layers know nothing about outer layers        │
└─────────────────────────────────────────────────────────────┘
```

### Login Flow Comparison

**Current:**
```
main.go
  │
  ├─► storage.Open() ─────────────────────────┐
  │                                          │
  ├─► handlers.NewServer(store, ...) ────────┤
  │    │                                      │
  │    ├─► AuthController{store} ─────────────┤
  │    │    │                                 │
  │    │    └─► LoginHandler                  │
  │    │         │                            │
  │    │         ├─► store.GetOperatorByEmail() ─┐
  │    │         ├─► store.VerifyPassword()  ─────┤
  │    │         └─► session.CreateCookie()   ────┤
  │    │                                           │
  │    ├─► RateLimiter{}                         │
  │    ├─► TurnstileVerifier{}                   │
  │    ├─► LockoutMiddleware{}                    │
  │    └─► CSRFProtector{}                       │
  │                                          │
  └─► server.Routes() ◄──────────────────────┘
```

**Proposed:**
```
main.go (wires everything)
  │
  ├─► infrastructure.storage.NewSQLite(cfg.DatabasePath)
  │       │
  │       └─► *sql.DB
  │
  ├─► infrastructure.crypto.NewArgon2Hasher()
  │       │
  │       └─► PasswordHasher (interface)
  │
  ├─► infrastructure.auth.NewSessionManager(cfg.SessionSecret)
  │       │
  │       └─► SessionManager (interface)
  │
  ├─► application.NewAuthService(
  │         operatorRepo,      // domain.OperatorRepository
  │         passwordHasher,    // domain.PasswordHasher
  │         sessionManager,    // domain.SessionManager
  │     )
  │       │
  │       └─► *application.AuthService
  │
  ├─► handlers.NewAuthHandler(authService)
  │       │
  │       └─► *handlers.AuthHandler
  │
  └─► router.Setup(authHandler, middleware...)
       │
       └─► gin.Engine
```

---

## File Organization

### Proposed Directory Structure

```
apps/api/
│
├── cmd/
│   └── api/
│       └── main.go                      # Entry point, DI wiring
│
├── internal/
│   │
│   ├── api/                            # PRESENTATION LAYER
│   │   │
│   │   ├── handlers/                   # HTTP handlers (ONE FILE PER ENDPOINT)
│   │   │   ├── auth/
│   │   │   │   ├── login.go
│   │   │   │   ├── logout.go
│   │   │   │   ├── register.go
│   │   │   │   ├── refresh.go
│   │   │   │   ├── oauth.go            # OAuth flow
│   │   │   │   ├── mfa.go              # MFA operations
│   │   │   │   └── password.go         # Reset, change
│   │   │   │
│   │   │   ├── device/
│   │   │   │   ├── register.go
│   │   │   │   ├── status.go
│   │   │   │   ├── fcm_token.go
│   │   │   │   └── delete.go
│   │   │   │
│   │   │   ├── command/
│   │   │   │   ├── execute.go
│   │   │   │   └── status.go
│   │   │   │
│   │   │   ├── health.go
│   │   │   ├── version.go
│   │   │   └── error.go                # Shared error handling
│   │   │
│   │   ├── middleware/                 # HTTP middleware
│   │   │   ├── ratelimit.go
│   │   │   ├── cors.go
│   │   │   ├── security.go
│   │   │   ├── recovery.go
│   │   │   ├── request_id.go
│   │   │   └── logging.go
│   │   │
│   │   └── router.go                   # Route definitions
│   │
│   ├── application/                     # APPLICATION LAYER
│   │   │
│   │   ├── auth/
│   │   │   ├── login.go
│   │   │   ├── logout.go
│   │   │   ├── register.go
│   │   │   ├── oauth.go
│   │   │   ├── mfa.go
│   │   │   └── dto/                    # Request/Response DTOs
│   │   │       ├── login.go
│   │   │       └── register.go
│   │   │
│   │   ├── device/
│   │   │   ├── register.go
│   │   │   └── dto/
│   │   │
│   │   └── errors.go                   # Application-level errors
│   │
│   ├── domain/                         # DOMAIN LAYER
│   │   │
│   │   ├── operator/
│   │   │   ├── entity.go               # Operator struct
│   │   │   ├── repository.go           # Interface
│   │   │   └── service.go              # Interface (optional)
│   │   │
│   │   ├── device/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   │
│   │   ├── session/
│   │   │   ├── entity.go
│   │   │   └── repository.go
│   │   │
│   │   ├── command/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   │
│   │   └── errors.go                   # Domain errors (ErrNotFound, etc.)
│   │
│   └── infrastructure/                 # INFRASTRUCTURE LAYER
│       │
│       ├── storage/                     # Database implementation
│       │   ├── sqlite.go               # Connection management
│       │   ├── operator.go             # Implements domain.OperatorRepository
│       │   ├── device.go               # Implements domain.DeviceRepository
│       │   ├── session.go              # Implements domain.SessionRepository
│       │   ├── command.go              # Implements domain.CommandRepository
│       │   └── migrations/             # Schema migrations
│       │       ├── 001_init.sql
│       │       ├── 002_add_mfa.sql
│       │       └── ...
│       │
│       ├── crypto/                      # Cryptographic implementations
│       │   ├── argon2.go               # Password hashing (implements domain.PasswordHasher)
│       │   ├── aes_gcm.go             # Encryption (implements domain.Encrypter)
│       │   ├── hmac.go                # HMAC signing
│       │   └── uuid.go                # UUID generation
│       │
│       ├── auth/                        # Auth infrastructure
│       │   ├── session.go              # Session management (implements domain.SessionManager)
│       │   └── cookies.go             # Cookie handling
│       │
│       └── external/                    # External services
│           ├── fcm.go                  # Firebase Cloud Messaging
│           ├── email.go                # Email sending
│           └── turnstile.go           # Cloudflare Turnstile
│
├── pkg/
│   ├── config/
│   │   └── config.go                   # Configuration loading
│   │
│   ├── models/                         # Shared models (only if truly shared across layers)
│   │   └── errors.go                   # Error types
│   │
│   └── logging/
│       └── logger.go                   # Logger setup
│
└── go.mod
```

---

## Where to Find Things

### Current vs Proposed

| What You're Looking For | CURRENT Location | PROPOSED Location |
|------------------------|-----------------|-------------------|
| **HTTP Routing** | `handlers/server.go` | `internal/api/router.go` |
| **Login Handler** | `handlers/auth_core.go` (mixed) | `internal/api/handlers/auth/login.go` |
| **OAuth Handler** | `handlers/auth_oauth.go` (mixed) | `internal/api/handlers/auth/oauth.go` |
| **Device Registration** | `handlers/device.go` (mixed) | `internal/api/handlers/device/register.go` |
| **Rate Limiting** | `middleware/rate_limiter.go` | `internal/api/middleware/ratelimit.go` |
| **CSRF Protection** | `middleware/csrf.go` | `internal/api/middleware/security.go` |
| **Login Business Logic** | `handlers/auth_core.go` | `internal/application/auth/login.go` |
| **Password Hashing** | `pkg/storage/crypto.go` | `internal/infrastructure/crypto/argon2.go` |
| **Session Management** | `internal/auth/session.go` | `internal/infrastructure/auth/session.go` |
| **Database Queries** | `pkg/storage/operators.go` | `internal/infrastructure/storage/operator.go` |
| **Operator Entity** | `pkg/models/` | `internal/domain/operator/entity.go` |
| **Repository Interface** | (doesn't exist) | `internal/domain/operator/repository.go` |
| **Error Types** | scattered | `internal/domain/errors.go` |

### Issue-to-File Mapping (Proposed)

| Issue Type | Where to Look | Why |
|-----------|---------------|-----|
| Handler returns wrong status code | `internal/api/handlers/<feature>/` | HTTP concern |
| Auth logic is wrong | `internal/application/auth/` | Business logic |
| Password not hashing correctly | `internal/infrastructure/crypto/` | Implementation |
| Database query is slow | `internal/infrastructure/storage/` | Database |
| Rate limit too aggressive | `internal/api/middleware/ratelimit.go` | HTTP middleware |
| Session not persisting | `internal/infrastructure/auth/session.go` | Session infra |
| Email not sending | `internal/infrastructure/external/email.go` | External service |

---

## Migration Path

### Phase 1: Extract Domain Layer (1-2 days)

1. Create `internal/domain/` directory
2. Extract entities: `Operator`, `Device`, `Session`, `Command`
3. Extract repository interfaces
4. No implementation changes yet

**Result:** Code compiles, same behavior, clearer structure

### Phase 2: Create Infrastructure Layer (2 days)

1. Move `pkg/storage/` → `internal/infrastructure/storage/`
2. Move `pkg/crypto/` → `internal/infrastructure/crypto/`
3. Move `internal/auth/` → `internal/infrastructure/auth/`
4. Make implementations satisfy domain interfaces

**Result:** Dependencies point inward correctly

### Phase 3: Extract Application Layer (2 days)

1. Create `internal/application/` directory
2. Move business logic from handlers to use cases
3. Handlers become thin HTTP wrappers

**Result:** Handlers do HTTP only, business logic is testable

### Phase 4: Split Handlers (1 day)

1. Split `handlers/auth_core.go` → `handlers/auth/login.go`, `logout.go`, etc.
2. Split `handlers/device.go` → `handlers/device/`

**Result:** One file per endpoint, easy to find

### Phase 5: Consolidate Utilities (1 day)

1. Create `internal/crypto/` package for AES-GCM, HMAC utilities
2. Create `internal/auth/token.go` for token generation
3. Remove all duplication

**Result:** DRY code, single place to fix issues

---

## Benefits of Clean Architecture

### For Finding Issues

| Problem | Old Way | New Way |
|---------|---------|---------|
| "Login returns 500" | Search 10 files | `internal/api/handlers/auth/login.go` |
| "Password verification fails" | Search 5 files | `internal/infrastructure/crypto/argon2.go` |
| "Session expires early" | Search 8 files | `internal/infrastructure/auth/session.go` |
| "Rate limit too strict" | One file | `internal/api/middleware/ratelimit.go` |
| "DB query is slow" | `storage/operators.go` | `internal/infrastructure/storage/operator.go` |

### For Testing

```
DOMAIN LAYER: Test pure business logic (no mocks needed)
  └─► Unit test AuthService.Login()

INFRASTRUCTURE: Test implementations against interfaces
  └─► Integration test SQLiteOperatorRepository

APPLICATION: Test use cases with mocked dependencies
  └─► Unit test AuthService with mock OperatorRepository

PRESENTATION: Test HTTP handling with mocked application
  └─► Integration test LoginHandler with mock AuthService
```

### For Onboarding

New developer asks: "Where is the login code?"

**Old answer:** "It's in `handlers/auth_core.go`, but also `storage/operators.go` for the DB query, and `internal/auth/session.go` for sessions, and `pkg/crypto/` for password hashing..."

**New answer:** "Start in `internal/api/handlers/auth/login.go` - that's where the HTTP handling is. Business logic is in `internal/application/auth/login.go`. Database is in `internal/infrastructure/storage/operator.go`."

---

## Summary

### Current State
- **No clear layers** - Everything talks to everything
- **Handlers do too much** - HTTP + business + DB
- **Storage is a god object** - 50+ methods
- **Duplication everywhere** - Same patterns in 10+ places
- **Hard to find things** - Issue in 5 different files

### Target State
- **Clean layers** - Presentation → Application → Domain ← Infrastructure
- **Handlers are thin** - HTTP only, delegate to application
- **Domain has interfaces** - Infrastructure implements them
- **Single responsibility** - One file per feature
- **Dependencies point inward** - Domain center, infrastructure outside

### Effort
| Phase | Time | Deliverable |
|-------|------|-------------|
| 1. Domain Layer | 1-2 days | Entities + Interfaces |
| 2. Infrastructure | 2 days | Move storage/crypto/auth |
| 3. Application | 2 days | Use cases extracted |
| 4. Split Handlers | 1 day | One file per endpoint |
| 5. Consolidate | 1 day | Remove duplication |

**Total: 7-8 days**

---

*This document defines the target architecture for the Vyzorix Update Server backend.*
