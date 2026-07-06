# PRD: Vyzorix API Server Architecture Refactoring

## Introduction

This document outlines a refactoring plan to improve the vyzorix-update-server API architecture. The current codebase follows a Clean Architecture-inspired structure but has several coupling issues, confusing naming, and code organization problems that will cause maintenance pain as the codebase grows.

**Problems to Solve:**
1. Dual `auth` packages causing confusion about which layer things belong to
2. Handlers directly coupling to infrastructure (audit, IP intelligence)
3. Middleware depending on application services creating circular dependency risk
4. `server.go` acting as a "god object" for dependency injection
5. Domain logic leaking into application layer

---

## Goals

- **G-1:** Achieve clear, unambiguous layer boundaries (Domain → Application → Infrastructure → API)
- **G-2:** Reduce coupling between layers, especially handlers to infrastructure
- **G-3:** Make the codebase easier to test through dependency injection
- **G-4:** Improve discoverability — any developer can find where code belongs
- **G-5:** Reduce the cognitive load of understanding `server.go`
- **G-6:** Maintain backward compatibility — all existing APIs must continue to work

---

## Non-Goals

- **NG-1:** Do NOT extract microservices — this is a refactor, not an architecture change
- **NG-2:** Do NOT change business logic — only restructuring, no functional changes
- **NG-3:** Do NOT rewrite existing tests — tests should pass after refactoring without changes
- **NG-4:** Do NOT add new features during this refactoring
- **NG-5:** Do NOT change the database schema

---

## Refactoring Phases

### Phase 1: Package Renaming & Clarification

**Objective:** Fix the confusing dual `auth` package problem.

#### RF-1.1: Rename `internal/auth` → `internal/infrastructure/security`

```
current:  internal/auth/
target:   internal/infrastructure/security/
```

Files to move:
- `jwt.go` → `infrastructure/security/jwt/`
- `password.go` → `infrastructure/security/password/`
- `totp.go` → `infrastructure/security/totp/`
- `ratelimit.go` → `infrastructure/security/ratelimit/`
- `session.go` → `infrastructure/security/session/`
- `google_token.go` → `infrastructure/security/oauth/`
- `lockout.go` → `infrastructure/security/lockout/`
- `request_signer.go` → `infrastructure/security/request_signer/`
- `revocation.go` → `infrastructure/security/revocation/`
- `secretstore/` → `infrastructure/security/secretstore/`

**Rationale:** These are all security infrastructure concerns. Grouping them under `infrastructure/security/` makes it clear they're implementation details.

#### RF-1.2: Rename `internal/application/auth/` → `internal/application/auth/` (keep) + clarify

The `application/auth/` package contains `AuthService` which is appropriate for the application layer. Keep it but ensure imports are consistent.

#### RF-1.3: Update all import paths

```bash
# Find all Go files importing old paths
grep -r "vyzorix/apps/api/internal/auth" --include="*.go" .

# Replace with new paths
sed -i 's|vyzorix/apps/api/internal/auth|vyzorix/apps/api/internal/infrastructure/security|g'
```

**Acceptance Criteria:**
- [ ] `internal/auth/` directory no longer exists
- [ ] All imports updated to `internal/infrastructure/security/`
- [ ] All tests pass without modification
- [ ] No circular dependencies introduced

---

### Phase 2: Extract Handler Output Adapters

**Objective:** Decouple handlers from infrastructure (audit logging, IP intelligence).

#### RF-2.1: Create `internal/api/adapters/response/` package

```go
// internal/api/adapters/response/response.go
package response

// ResponseHandler defines how handlers format responses
type ResponseHandler interface {
    OK(c *gin.Context, data interface{})
    Created(c *gin.Context, data interface{})
    Error(c *gin.Context, status int, errCode, message string)
    Unauthorized(c *gin.Context, message string)
    // ... other standard responses
}

// AuditLogger interface for audit infrastructure
type AuditLogger interface {
    Log(c context.Context, event *AuditEvent) error
}

// IPIntelligence interface 
type IPIntelligence interface {
    RecordAuthSuccess(c *gin.Context)
    RecordAuthFailure(c *gin.Context)
}
```

#### RF-2.2: Refactor `LoginHandler`

**Before:**
```go
type LoginHandler struct {
    authService    *auth.AuthService
    auditLogger   *audit.Logger      // ← infrastructure leak
    ipIntelligence *middleware.IPIntelligence  // ← infrastructure leak
}
```

**After:**
```go
type LoginHandler struct {
    authService    *auth.AuthService
    presenter     *response.Presenter
}

func NewLoginHandler(
    authService *auth.AuthService,
    presenter *response.Presenter,
) *LoginHandler {
    return &LoginHandler{
        authService: authService,
        presenter:   presenter,
    }
}
```

The presenter handles audit logging and IP intelligence internally:
```go
func (h *LoginHandler) Handle(c *gin.Context) {
    // ... validation ...
    
    result, session, err := h.authService.Login(ctx, &req)
    if err != nil {
        // Presenter handles audit and IP tracking
        h.presenter.AuthFailure(c, err, c.ClientIP())
        return
    }
    
    h.presenter.AuthSuccess(c, session)
    c.JSON(http.StatusOK, result)
}
```

#### RF-2.3: Apply to all handlers

| Handler | Infrastructure Dependencies to Extract |
|---------|----------------------------------------|
| `LoginHandler` | `audit.Logger`, `IPIntelligence` |
| `RegisterHandler` | `audit.Logger`, `IPIntelligence` |
| `LogoutHandler` | `audit.Logger` |
| `MFAHandler` | `audit.Logger` |
| All others | Review for leaks |

**Acceptance Criteria:**
- [ ] Handlers only depend on application services and presenter interface
- [ ] Audit logging moved to presenter layer
- [ ] IP intelligence moved to presenter layer
- [ ] Handler tests only mock `AuthService`, not infrastructure

---

### Phase 3: Middleware Refactoring

**Objective:** Reduce coupling between middleware and application services.

#### RF-3.1: Analyze middleware dependencies

```go
// Current state - middleware might depend on:
type MiddlewareConfig struct {
    AuthService      *auth.AuthService  // ← application layer
    SessionManager   *security.SessionManager  // ← infrastructure
    RateLimiter      *middleware.RateLimiter
    Lockout          *middleware.Lockout
    // etc.
}
```

#### RF-3.2: Extract middleware creation to factory

Create `internal/api/middleware/factory.go`:

```go
package middleware

type MiddlewareFactory struct {
    authService    *auth.AuthService
    sessionManager *security.SessionManager
    config        Config
}

func NewMiddlewareFactory(
    authService *auth.AuthService,
    sessionManager *security.SessionManager,
    cfg Config,
) *MiddlewareFactory {
    return &MiddlewareFactory{
        authService:    authService,
        sessionManager: sessionManager,
        config:         cfg,
    }
}

func (f *MiddlewareFactory) CookieAuth() gin.HandlerFunc {
    return NewCookieAuth(f.sessionManager, f.authService)
}

func (f *MiddlewareFactory) RateLimiter() gin.HandlerFunc {
    return NewRateLimiter(f.config.RateLimitPerMinute, f.config.RateLimitWindow)
}

func (f *MiddlewareFactory) AccountLockout() gin.HandlerFunc {
    return NewLockout(LoadLockoutConfig())
}

// etc.
```

#### RF-3.3: Simplify server.go middleware registration

**Before:**
```go
// server.go - 50+ lines of middleware setup
cookieAuth := middleware.NewCookieAuth(cfg.SessionManager, cfg.AuthService)
lockoutConfig := middleware.LoadLockoutConfig()
lockout := middleware.NewLockout(lockoutConfig)
csrfConfig := middleware.DefaultCSRFConfig()
csrfProtector := middleware.NewCSRFProtector(csrfConfig)
turnstileCfg := middleware.LoadTurnstileConfig()
turnstileVerifier := middleware.NewTurnstileVerifier(turnstileCfg)
// ... 100 more lines
```

**After:**
```go
// server.go
mwFactory := middleware.NewMiddlewareFactory(
    authService,
    sessionManager,
    middleware.Config{...},
)

engine.Use(mwFactory.Recovery())
engine.Use(mwFactory.Logger())
engine.Use(mwFactory.RateLimiter())
engine.Use(mwFactory.CookieAuth())
// etc.
```

**Acceptance Criteria:**
- [ ] Middleware creation isolated to factory
- [ ] Server.go middleware section reduced to < 30 lines
- [ ] Middleware dependencies clearly documented

---

### Phase 4: Extract Domain Logic

**Objective:** Move domain logic from application to domain layer.

#### RF-4.1: Move password policy to domain

```
current:  application/auth/service.go (ValidatePassword, PasswordPolicy)
target:   domain/operator/password.go
```

```go
// domain/operator/password.go
package operator

type PasswordPolicy struct {
    MinLength      int
    MaxLength      int
    RequireUpper   bool
    RequireLower   bool
    RequireDigit   bool
    RequireSpecial bool
}

var DefaultPasswordPolicy = PasswordPolicy{
    MinLength:      8,
    MaxLength:      128,
    RequireUpper:   true,
    RequireLower:   true,
    RequireDigit:   true,
    RequireSpecial: true,
}

func (p PasswordPolicy) Validate(password string) error {
    // ... implementation
}
```

#### RF-4.2: Move email validation to domain

```
current:  infrastructure/security/validate.go
target:   domain/operator/email.go
```

```go
// domain/operator/email.go
package operator

func ValidateEmail(email string) error {
    // ... existing implementation
}
```

#### RF-4.3: Review for other domain leaks

Check `application/` packages for:
- Logic that should be in domain entities (e.g., `HasMFA()`, role checking)
- Validation that belongs in domain
- Error definitions that belong in domain

**Acceptance Criteria:**
- [ ] Password policy in domain/operator
- [ ] Email validation in domain/operator
- [ ] Domain entities contain all entity-specific methods
- [ ] Application services only orchestrate, don't contain domain rules

---

### Phase 5: Server.go Decomposition

**Objective:** Break `server.go` (currently ~800 lines) into manageable pieces.

#### RF-5.1: Create `wire/` package

```
internal/api/
 server.go          ← Main server, minimal (~100 lines)
 wire/
    wire_handlers.go    ← Handler instantiation
    wire_middleware.go  ← Middleware factory usage
    wire_services.go    ← Service instantiation
    wire.go             ← Main wiring function
```

#### RF-5.2: `wire_handlers.go`

```go
package api

import (
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/auth"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/handlers/device"
    // ... other handlers
)

type HandlerConfig struct {
    AuthService    *auth.AuthService
    DeviceService  *device.Service
    Presenter     *response.Presenter
    // ... dependencies
}

func WireHandlers(cfg HandlerConfig) *HandlerSet {
    return &HandlerSet{
        Login:    auth.NewLoginHandler(cfg.AuthService, cfg.Presenter),
        Register: auth.NewRegisterHandler(cfg.AuthService, cfg.Presenter),
        Device:   device.NewRegisterHandler(cfg.DeviceService),
        // ...
    }
}

type HandlerSet struct {
    Login    *auth.LoginHandler
    Register *auth.RegisterHandler
    Device   *device.RegisterHandler
    // ...
}
```

#### RF-5.3: `wire_middleware.go`

```go
package api

import (
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
)

type MiddlewareConfig struct {
    AuthService    *auth.AuthService
    SessionManager *security.SessionManager
    RateLimit      RateLimitConfig
}

func WireMiddleware(cfg MiddlewareConfig) *MiddlewareSet {
    factory := middleware.NewFactory(cfg.AuthService, cfg.SessionManager)
    
    return &MiddlewareSet{
        Recovery:   factory.Recovery(),
        Logger:     factory.Logger(),
        RateLimit:  factory.RateLimiter(cfg.RateLimit.PerMinute),
        CookieAuth: factory.CookieAuth(),
        Lockout:    factory.Lockout(),
        // ...
    }
}

type MiddlewareSet struct {
    Recovery   gin.HandlerFunc
    Logger     gin.HandlerFunc
    RateLimit  gin.HandlerFunc
    CookieAuth gin.HandlerFunc
    Lockout    gin.HandlerFunc
}
```

#### RF-5.4: Simplified `server.go`

```go
// server.go (target: ~150 lines)
package api

func NewServer(cfg *ServerConfig) *Server {
    // 1. Wire infrastructure
    infra := wireInfrastructure(cfg)
    
    // 2. Wire services
    services := wireServices(infra)
    
    // 3. Wire middleware
    middleware := wireMiddleware(services)
    
    // 4. Wire handlers
    handlers := wireHandlers(services)
    
    // 5. Build engine
    engine := gin.New()
    engine.Use(middleware.Recovery)
    engine.Use(middleware.Logger)
    // ... register routes using handlers
    
    return &Server{
        engine:    engine,
        handlers:  handlers,
        middleware: middleware,
    }
}
```

**Acceptance Criteria:**
- [ ] `server.go` reduced to < 200 lines
- [ ] Each `wire_*.go` file < 100 lines
- [ ] Adding a new handler only requires editing `wire_handlers.go`
- [ ] Adding a new middleware only requires editing `wire_middleware.go`

---

## Implementation Order

| Phase | Description | Estimated Effort | Risk |
|-------|-------------|------------------|------|
| 1 | Package renaming | Medium | Low (mechanical) |
| 2 | Response adapters | Medium | Medium (interface changes) |
| 3 | Middleware factory | Medium | Low (additive) |
| 4 | Domain logic extraction | High | Medium (logic moves) |
| 5 | Server decomposition | Medium | Low (refactoring only) |

**Recommended Order:** Phase 1 → Phase 3 → Phase 2 → Phase 4 → Phase 5

Rationale:
1. Phase 1 (rename) is prerequisite for everything else
2. Phase 3 (middleware factory) is additive, low risk
3. Phase 2 (adapters) depends on Phase 1
4. Phase 4 (domain logic) is the highest risk, do near end
5. Phase 5 (server decomposition) is the payoff for all previous work

---

## Verification Checklist

After each phase:

- [ ] All tests pass: `cd apps/api && make test`
- [ ] Linting passes: `make lint-go`
- [ ] No new import cycles introduced: `go mod tidy && go build ./...`
- [ ] API contracts unchanged (compare with integration tests)
- [ ] No regression in functionality

---

## Open Questions

1. **GraphQL handlers:** Should they follow the same presenter pattern?
   - Current: GraphQL resolvers call services directly
   - Consider: Same adapter pattern for consistency

2. **WebSocket handlers:** These have different patterns (goroutine-based). Should they be refactored similarly?

3. **Error handling:** Should we introduce a unified error type across layers?
   - Current: `application.Err*`, `domain.Err*`, `infraerror*`
   - Consider: Single error type with source tracking

4. **Configuration:** Should middleware config live in `infrastructure/config/` or stay with middleware?

5. **Performance:** Will the presenter pattern add measurable overhead?
   - Benchmark before/after if latency is critical

---

## Appendix: Target Structure

After all refactoring:

```
internal/
 domain/
    operator/
       entity.go
       repository.go
       errors.go
       email.go          ← email validation moved here
       password.go       ← password policy moved here
       role.go
    device/
       entity.go
       repository.go
       errors.go
    client/
    command/
    session/
    ...

 application/
    auth/
       service.go
    device/
       service.go
    client/
       service.go
    command/
       service.go
    dto/
       login.go
       register.go
       ...
    shared/
        id.go

 infrastructure/
    security/              ← renamed from internal/auth
       jwt/
       password/
       totp/
       session/
       lockout/
       ratelimit/
       revocation/
       oauth/
       secretstore/
    storage/
       sqlite.go
       operator.go
       device.go
       ...
    email/
    fcm/
    logging/
    metrics/
    config/
    crypto/

 api/
     server.go             ← ~150 lines
     wire/
        wire.go
        wire_handlers.go
        wire_middleware.go
        wire_services.go
     adapters/
        response/
            presenter.go
     handlers/
        auth/
        device/
        command/
        websocket/
     middleware/
        factory.go        ← new
        ratelimit.go
        cookie_auth.go
        ...
     graphql/
```
