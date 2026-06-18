# Vyzorix Update Server - Complete File Mapping Guide

**Date:** 2026-06-17  
**Status:** Complete File-by-File Analysis  
**Purpose:** Maps every Go file to new architecture location and required changes

---

## Table of Contents

1. [Summary Statistics](#summary-statistics)
2. [Current → Proposed Location Mapping](#current--proposed-location-mapping)
3. [File-by-File Detailed Analysis](#file-by-file-detailed-analysis)
4. [Dependency Matrix](#dependency-matrix)
5. [Files to DELETE](#files-to-delete)
6. [Files to CREATE (New)](#files-to-create-new)
7. [Files Requiring Significant Rewrite](#files-requiring-significant-rewrite)

---

## Summary Statistics

| Category | Count |
|----------|-------|
| **Total Go Files** | 140 |
| **Handler Files** | 27 (including tests) |
| **Middleware Files** | 24 (including tests) |
| **Auth Package Files** | 21 |
| **Storage Files** | 17 (including tests) |
| **Model Files** | 13 |
| **Config Files** | 10 |
| **Other Infrastructure** | 28 |
| **Files to DELETE** | 5 |
| **Files to CREATE** | 25+ |
| **Files to MOVE** | 45+ |
| **Files to REWRITE** | 20+ |

---

## Current → Proposed Location Mapping

### HANDLERS (Presentation Layer)

| Current Location | Proposed Location | Action | Changes Needed |
|-----------------|------------------|--------|---------------|
| `internal/api/handlers/server.go` | `internal/api/router.go` | **RENAME + REWRITE** | Extract Server struct to router, split by feature |
| `internal/api/handlers/auth_core.go` | `internal/api/handlers/auth/login.go` + `register.go` + `logout.go` | **SPLIT + REWRITE** | Split by endpoint, thin handler (HTTP only) |
| `internal/api/handlers/auth_oauth.go` | `internal/api/handlers/auth/oauth.go` | **MOVE** | Keep as-is structure, refactor to call application layer |
| `internal/api/handlers/auth_mfa.go` | `internal/api/handlers/auth/mfa.go` | **MOVE** | Add logger, cache config |
| `internal/api/handlers/auth_password_reset.go` | `internal/api/handlers/auth/password_reset.go` | **MOVE** | Use shared token generation |
| `internal/api/handlers/auth_email_verify.go` | `internal/api/handlers/auth/email_verify.go` | **MOVE** | Use shared token generation |
| `internal/api/handlers/auth_settings.go` | `internal/api/handlers/auth/settings.go` | **MOVE** | Minor refactor |
| `internal/api/handlers/auth_admin.go` | `internal/api/handlers/auth/admin.go` | **MOVE** | Minor refactor |
| `internal/api/handlers/auth_csrf.go` | **DELETE** | **DELETE** | Logic moved to middleware |
| `internal/api/handlers/auth_rate_limit.go` | **DELETE** | **DELETE** | Logic moved to middleware |
| `internal/api/handlers/auth_utils.go` | `internal/application/shared/` | **MOVE** | Extract shared utilities |
| `internal/api/handlers/auth_test.go` | `internal/api/handlers/auth/auth_test.go` | **MOVE** | Keep with handler tests |
| `internal/api/handlers/auth_mfa_test.go` | `internal/api/handlers/auth/mfa_test.go` | **MOVE** | Keep with handler tests |
| `internal/api/handlers/device.go` | `internal/api/handlers/device/register.go` + `status.go` | **SPLIT** | Split by endpoint |
| `internal/api/handlers/device_test.go` | `internal/api/handlers/device/device_test.go` | **MOVE** | Keep with handler |
| `internal/api/handlers/command.go` | `internal/api/handlers/command/execute.go` | **MOVE** | Minor refactor |
| `internal/api/handlers/command_test.go` | `internal/api/handlers/command/command_test.go` | **MOVE** | Keep with handler |
| `internal/api/handlers/updater.go` | `internal/api/handlers/device/updater.go` | **MOVE** | Move to device folder |
| `internal/api/handlers/websocket_handler.go` | `internal/api/handlers/websocket/handler.go` | **MOVE** | Minor refactor |
| `internal/api/handlers/websocket_handler_test.go` | `internal/api/handlers/websocket/handler_test.go` | **MOVE** | Keep with handler |
| `internal/api/handlers/health.go` | `internal/api/handlers/health.go` | **MOVE** | Keep as-is |
| `internal/api/handlers/health_test.go` | `internal/api/handlers/health_test.go` | **MOVE** | Keep with handler |
| `internal/api/handlers/health_handler_test.go` | `internal/api/handlers/health_handler_test.go` | **MOVE** | Keep with handler |
| `internal/api/handlers/lockout.go` | **DELETE** | **DELETE** | Logic in `internal/auth/lockout.go` |
| `internal/api/handlers/lockout_test.go` | **DELETE** | **DELETE** | Logic in `internal/auth/lockout.go` |
| `internal/api/handlers/admin_clients.go` | `internal/api/handlers/admin/clients.go` | **MOVE** | Move to admin folder |
| `internal/api/handlers/client_credentials.go` | `internal/api/handlers/auth/client_credentials.go` | **MOVE** | Move to auth folder |
| `internal/api/handlers/rate_limit_test.go` | `internal/api/middleware/rate_limiter_test.go` | **MOVE** | Move to middleware tests |

---

### MIDDLEWARE (Presentation Layer)

| Current Location | Proposed Location | Action | Changes Needed |
|-----------------|------------------|--------|---------------|
| `internal/api/middleware/auth.go` | `internal/api/middleware/auth.go` | **KEEP** | Minor refactor |
| `internal/api/middleware/auth_enum_safe.go` | `internal/domain/auth/enum_safe.go` | **MOVE** | Consolidate with other enum utils |
| `internal/api/middleware/auth_revocation.go` | `internal/auth/revocation.go` | **MOVE** | Move to auth package |
| `internal/api/middleware/auth_test.go` | `internal/api/middleware/auth_test.go` | **MOVE** | Keep with middleware |
| `internal/api/middleware/body_size.go` | `internal/api/middleware/body_size.go` | **KEEP** | Keep as-is |
| `internal/api/middleware/body_size_test.go` | `internal/api/middleware/body_size_test.go` | **MOVE** | Keep with middleware |
| `internal/api/middleware/cors.go` | `internal/api/middleware/cors.go` | **KEEP** | Keep as-is |
| `internal/api/middleware/cors_test.go` | `internal/api/middleware/cors_test.go` | **MOVE** | Keep with middleware |
| `internal/api/middleware/csrf.go` | `internal/api/middleware/csrf.go` | **KEEP** | Fix empty secret issue |
| `internal/api/middleware/error.go` | `internal/api/middleware/error.go` | **KEEP** | Consolidate error responses |
| `internal/api/middleware/lockout.go` | `internal/api/middleware/lockout.go` | **KEEP** | Fix email case sensitivity |
| `internal/api/middleware/logger.go` | `internal/api/middleware/logging.go` | **RENAME** | Minor refactor |
| `internal/api/middleware/logger_test.go` | `internal/api/middleware/logging_test.go` | **MOVE** | Keep with middleware |
| `internal/api/middleware/rate_limiter.go` | `internal/api/middleware/ratelimit.go` | **REWRITE** | Already has cleanup, minor fixes |
| `internal/api/middleware/rate_limiter_test.go` | `internal/api/middleware/ratelimit_test.go` | **MOVE** | Keep with middleware |
| `internal/api/middleware/recovery.go` | `internal/api/middleware/recovery.go` | **KEEP** | Extract stack filtering util |
| `internal/api/middleware/replay_protection.go` | `internal/infrastructure/crypto/replay_cache.go` | **MOVE** | Consolidate replay protection |
| `internal/api/middleware/request_id.go` | `internal/api/middleware/request_id.go` | **KEEP** | Keep as-is |
| `internal/api/middleware/request_id_test.go` | `internal/api/middleware/request_id_test.go` | **MOVE** | Keep with middleware |
| `internal/api/middleware/request_signing.go` | `internal/api/middleware/request_signing.go` + `internal/infrastructure/crypto/` | **SPLIT** | Extract crypto to infrastructure |
| `internal/api/middleware/response_encryption.go` | `internal/infrastructure/crypto/response_encryption.go` | **MOVE** | Move to crypto package |
| `internal/api/middleware/security_config_test.go` | `internal/api/middleware/security_config_test.go` | **MOVE** | Keep with middleware |
| `internal/api/middleware/security_headers.go` | `internal/api/middleware/security_headers.go` | **KEEP** | Keep as-is |
| `internal/api/middleware/ssr-proxy.go` | `internal/api/middleware/ssr_proxy.go` | **RENAME** | Keep as-is |
| `internal/api/middleware/turnstile.go` | `internal/api/middleware/turnstile.go` | **KEEP** | Add rate limiting |
| `internal/api/middleware/user_enum.go` | `internal/domain/auth/user_enum.go` | **MOVE** | Consolidate with enum_safe |
| `internal/api/middleware/user_enum_block.go` | `internal/domain/auth/user_enum_block.go` | **MOVE** | Consolidate with enum_safe |

---

### AUTH PACKAGE

| Current Location | Proposed Location | Action | Changes Needed |
|-----------------|------------------|--------|---------------|
| `internal/auth/github.go` | `internal/infrastructure/auth/github.go` | **MOVE** | OAuth implementation |
| `internal/auth/google_token.go` | `internal/infrastructure/auth/google_token.go` | **MOVE** | OAuth implementation |
| `internal/auth/google_token_test.go` | `internal/infrastructure/auth/google_token_test.go` | **MOVE** | Keep with auth |
| `internal/auth/jwt.go` | `internal/infrastructure/auth/jwt.go` | **MOVE** | JWT implementation |
| `internal/auth/jwt_test.go` | `internal/infrastructure/auth/jwt_test.go` | **MOVE** | Keep with auth |
| `internal/auth/lockout.go` | `internal/domain/auth/lockout.go` + `internal/infrastructure/auth/lockout.go` | **SPLIT** | Domain interface + Infrastructure impl |
| `internal/auth/origin.go` | `internal/infrastructure/auth/origin.go` | **MOVE** | Origin validation |
| `internal/auth/origin_test.go` | `internal/infrastructure/auth/origin_test.go` | **MOVE** | Keep with auth |
| `internal/auth/password.go` | `internal/infrastructure/crypto/password.go` | **MOVE** | Password hashing |
| `internal/auth/password_test.go` | `internal/infrastructure/crypto/password_test.go` | **MOVE** | Keep with crypto |
| `internal/auth/ratelimit.go` | `internal/api/middleware/auth_ratelimit.go` | **MOVE** | Move to middleware |
| `internal/auth/ratelimit_test.go` | `internal/api/middleware/auth_ratelimit_test.go` | **MOVE** | Keep with middleware |
| `internal/auth/request_signer.go` | `internal/infrastructure/crypto/request_signer.go` | **MOVE** | Move to crypto |
| `internal/auth/revocation.go` | `internal/infrastructure/auth/revocation.go` | **MOVE** | Session revocation |
| `internal/auth/secretstore/secretstore.go` | `internal/infrastructure/auth/secretstore.go` | **MOVE** | Move up one level |
| `internal/auth/secretstore/secretstore_test.go` | `internal/infrastructure/auth/secretstore_test.go` | **MOVE** | Keep with auth |
| `internal/auth/session.go` | `internal/infrastructure/auth/session.go` | **MOVE** | Session management |
| `internal/auth/session_test.go` | `internal/infrastructure/auth/session_test.go` | **MOVE** | Keep with auth |
| `internal/auth/totp.go` | `internal/infrastructure/auth/totp.go` | **MOVE** | TOTP generation |
| `internal/auth/totp_qr.go` | `internal/infrastructure/auth/totp_qr.go` | **MOVE** | QR code generation |
| `internal/auth/validate.go` | `internal/domain/auth/validate.go` | **MOVE** | Validation logic |
| `internal/auth/validate_test.go` | `internal/domain/auth/validate_test.go` | **MOVE** | Keep with domain |

---

### STANDALONE INTERNAL FILES

| Current Location | Proposed Location | Action | Changes Needed |
|-----------------|------------------|--------|---------------|
| `internal/command_signer.go` | `internal/infrastructure/crypto/command_signer.go` | **MOVE** | Was orphaned file! |
| `internal/command_signer_test.go` | `internal/infrastructure/crypto/command_signer_test.go` | **MOVE** | Keep with crypto |
| `internal/email.go` | `internal/infrastructure/external/email.go` | **MOVE** | Email service |
| `internal/email_test.go` | `internal/infrastructure/external/email_test.go` | **MOVE** | Keep with external |

---

### STORAGE PACKAGE

| Current Location | Proposed Location | Action | Changes Needed |
|-----------------|------------------|--------|---------------|
| `pkg/storage/store.go` | `internal/infrastructure/storage/sqlite.go` | **MOVE + REWRITE** | Extract connection management |
| `pkg/storage/operators.go` | `internal/infrastructure/storage/operator.go` | **MOVE** | Implement `domain.OperatorRepository` |
| `pkg/storage/devices.go` | `internal/infrastructure/storage/device.go` | **MOVE + REWRITE** | Fix plaintext secret, implement interface |
| `pkg/storage/clients.go` | `internal/infrastructure/storage/client.go` | **MOVE + REWRITE** | Extract helpers |
| `pkg/storage/commands.go` | `internal/infrastructure/storage/command.go` | **MOVE** | Implement `domain.CommandRepository` |
| `pkg/storage/sessions.go` | `internal/infrastructure/storage/session.go` | **MOVE + REWRITE** | Bulk INSERT for revocations |
| `pkg/storage/settings.go` | `internal/infrastructure/storage/settings.go` | **MOVE** | Keep as-is |
| `pkg/storage/migrations.go` | `internal/infrastructure/storage/migrations/` | **MOVE + SPLIT** | Split into SQL files |
| `pkg/storage/telemetry.go` | `internal/infrastructure/storage/telemetry.go` | **MOVE** | Keep as-is |
| `pkg/storage/crypto.go` | `internal/infrastructure/crypto/storage_crypto.go` | **MOVE** | Password hashing for storage |
| `pkg/storage/uuid.go` | `internal/infrastructure/crypto/uuid.go` | **MOVE** | UUID generation (already fixed) |
| `pkg/storage/client_settings_test.go` | `internal/infrastructure/storage/client_settings_test.go` | **MOVE** | Keep with storage |
| `pkg/storage/pagination_test.go` | `internal/infrastructure/storage/pagination_test.go` | **MOVE** | Keep with storage |
| `pkg/storage/secret_hash_test.go` | `internal/infrastructure/storage/secret_hash_test.go` | **MOVE** | Keep with storage |
| `pkg/storage/sqlite_test.go` | `internal/infrastructure/storage/sqlite_test.go` | **MOVE** | Keep with storage |
| `pkg/storage/uuid_test.go` | `internal/infrastructure/crypto/uuid_test.go` | **MOVE** | Keep with crypto |

---

### OTHER PACKAGES

| Current Location | Proposed Location | Action | Changes Needed |
|-----------------|------------------|--------|---------------|
| `pkg/config/config.go` | `pkg/config/config.go` | **KEEP** | Configuration loading |
| `pkg/config/signing.go` | `pkg/config/signing.go` | **KEEP** | Signing config |
| `pkg/config/ssr.go` | `pkg/config/ssr.go` | **KEEP** | SSR config |
| `pkg/config/totp.go` | `pkg/config/totp.go` | **KEEP** | TOTP config |
| `pkg/config/turnstile.go` | `pkg/config/turnstile.go` | **KEEP** | Turnstile config |
| `pkg/config/validator.go` | `pkg/config/validator.go` | **KEEP** | Config validation |
| `pkg/config/config_test.go` | `pkg/config/config_test.go` | **KEEP** | Keep with config |
| `pkg/config/ssr_test.go` | `pkg/config/ssr_test.go` | **KEEP** | Keep with config |
| `pkg/config/validator_test.go` | `pkg/config/validator_test.go` | **KEEP** | Keep with config |
| `pkg/crypto/hmac.go` | `internal/infrastructure/crypto/hmac.go` | **MOVE** | HMAC signing |
| `pkg/crypto/hmac_test.go` | `internal/infrastructure/crypto/hmac_test.go` | **MOVE** | Keep with crypto |
| `pkg/logging/redactor.go` | `pkg/logging/redactor.go` | **KEEP** | Keep as-is |
| `pkg/logging/structured.go` | `pkg/logging/structured.go` | **KEEP** | Keep as-is |
| `pkg/logging/structured_test.go` | `pkg/logging/structured_test.go` | **KEEP** | Keep with logging |
| `pkg/models/auth.go` | `internal/application/dto/auth.go` | **MOVE** | Auth DTOs |
| `pkg/models/command.go` | `internal/domain/command/entity.go` | **MOVE** | Command entity |
| `pkg/models/device.go` | `internal/domain/device/entity.go` | **MOVE** | Device entity |
| `pkg/models/models.go` | `internal/domain/operator/entity.go` | **MOVE** | Operator entity + shared models |
| `pkg/models/response.go` | `internal/application/dto/response.go` | **MOVE** | Response DTOs |
| `pkg/models/telemetry.go` | `internal/domain/telemetry/entity.go` | **MOVE** | Telemetry entity |
| `pkg/models/updater.go` | `internal/domain/updater/entity.go` | **MOVE** | Updater entity |
| `pkg/models/*_test.go` | `internal/domain/*/` | **MOVE** | Move test files with source |

---

### FCM, WS, AUDIT, METRICS, SSR

| Current Location | Proposed Location | Action | Changes Needed |
|-----------------|------------------|--------|---------------|
| `internal/fcm/fcm.go` | `internal/infrastructure/external/fcm.go` | **MOVE** | FCM implementation |
| `internal/fcm/fcm_test.go` | `internal/infrastructure/external/fcm_test.go` | **MOVE** | Keep with external |
| `internal/fcm/notifier.go` | `internal/infrastructure/external/fcm_notifier.go` | **MOVE** | Notification service |
| `internal/fcm/safe_notifier_test.go` | `internal/infrastructure/external/fcm_notifier_test.go` | **MOVE** | Keep with external |
| `internal/ws/client.go` | `internal/infrastructure/websocket/client.go` | **MOVE** | WebSocket client |
| `internal/ws/hub.go` | `internal/infrastructure/websocket/hub.go` | **MOVE** | WebSocket hub |
| `internal/ws/hub_test.go` | `internal/infrastructure/websocket/hub_test.go` | **MOVE** | Keep with websocket |
| `internal/audit/logger.go` | `internal/infrastructure/audit/logger.go` | **MOVE** | Audit logging |
| `internal/audit/logger_test.go` | `internal/infrastructure/audit/logger_test.go` | **MOVE** | Keep with audit |
| `internal/audit/repository.go` | `internal/infrastructure/audit/repository.go` | **MOVE** | Audit storage |
| `internal/metrics/middleware.go` | `internal/api/middleware/metrics.go` | **MOVE** | Metrics middleware |
| `internal/metrics/prometheus.go` | `internal/infrastructure/metrics/prometheus.go` | **MOVE** | Prometheus metrics |
| `internal/ssr/builder.go` | `internal/infrastructure/ssr/builder.go` | **MOVE** | SSR builder |
| `internal/ssr/config.go` | `internal/infrastructure/ssr/config.go` | **MOVE** | SSR config |
| `internal/ssr/manager.go` | `internal/infrastructure/ssr/manager.go` | **MOVE** | SSR manager |
| `internal/ssr/monitor.go` | `internal/infrastructure/ssr/monitor.go` | **MOVE** | SSR monitoring |
| `internal/ssr/process.go` | `internal/infrastructure/ssr/process.go` | **MOVE** | SSR process |

---

### MAIN

| Current Location | Proposed Location | Action | Changes Needed |
|-----------------|------------------|--------|---------------|
| `main.go` | `cmd/api/main.go` | **MOVE** | Dependency injection wiring |

---

## File-by-File Detailed Analysis

### 🔴 HIGH PRIORITY FILES (Require Significant Rewrite)

---

#### 1. `internal/api/handlers/server.go` → `internal/api/router.go`

**Current Problems:**
- Server struct holds 17 dependencies
- Mixed concerns (routing + middleware setup + handler instantiation)
- Hard to trace dependencies

**Required Changes:**
```go
// BEFORE (server.go - 500+ lines)
type Server struct {
    Notifier        fcm.Notifier
    Store           *storage.Store
    // ... 15 more dependencies
}

func (s *Server) Routes() *gin.Engine {
    // All routing + middleware in one file
}

func (s *Server) Login(c *gin.Context) { /* handler */ }
func (s *Server) Logout(c *gin.Context) { /* handler */ }
// ... 30 more handlers
```

```go
// AFTER (router.go + split handlers)
type Router struct {
    engine *gin.Engine
    // Minimal deps - most go into application layer
}

func NewRouter(cfg *config.Config) *Router {
    // Wire dependencies here
}

// Handlers become thin wrappers
type AuthHandler struct {
    authService *application.AuthService
}
func (h *AuthHandler) Login(c *gin.Context) {
    // Parse request
    // Call authService.Login()
    // Return response
}
```

**Steps:**
1. Create `internal/api/router.go` with minimal router struct
2. Extract all handlers to separate files
3. Create `internal/application/` layer for business logic
4. Create `internal/domain/` layer for interfaces
5. Create `internal/infrastructure/` layer for implementations

---

#### 2. `internal/api/middleware/request_signing.go`

**Current Problems:**
- 4 locations with duplicated AES-256-GCM code
- O(n) cache eviction
- 55+ line Verify function

**Required Changes:**
```go
// BEFORE: All in one file
func (s *RequestSigning) decryptBody(...) { /* AES-GCM */ }
func (s *RequestSigning) EncryptBody(...) { /* AES-GCM */ }
func (s *RequestSigning) SignRequest(...) { /* AES-GCM */ }
func (s *RequestSigning) Verify(...) { /* 55 lines */ }
```

```go
// AFTER: Split into
// internal/infrastructure/crypto/aes_gcm.go
func Encrypt(key, plaintext []byte) ([]byte, error)
func Decrypt(key, ciphertext []byte) ([]byte, error)

// internal/infrastructure/crypto/replay_cache.go
type ReplayCache struct { /* with background cleanup */ }

// internal/api/middleware/request_signing.go (thin middleware)
type RequestSigningMiddleware struct {
    verifier *RequestVerifier  // From application layer
}
```

**Steps:**
1. Create `internal/infrastructure/crypto/aes_gcm.go`
2. Create `internal/infrastructure/crypto/replay_cache.go`
3. Rewrite request_signing.go to use extracted utilities
4. Split Verify into smaller focused methods

---

#### 3. `pkg/storage/operators.go`

**Current Problems:**
- 7+ similar update functions
- Missing domain interface
- Direct SQL in handler layer

**Required Changes:**
```go
// BEFORE: Direct storage access
func (ac *AuthController) Login(...) {
    op, err := ac.store.GetOperatorByEmail(ctx, req.Email)
    // handler does business logic
}
```

```go
// AFTER: Domain interface + application layer
// internal/domain/operator/repository.go
interface OperatorRepository {
    FindByEmail(ctx context.Context, email string) (*Operator, error)
    // ...
}

// internal/application/auth/login.go
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
    op, err := s.operatorRepo.FindByEmail(ctx, email)
    // application layer does business logic
}

// internal/infrastructure/storage/operator.go
type SQLiteOperatorRepository struct { db *sql.DB }
func (r *SQLiteOperatorRepository) FindByEmail(...) { /* implements interface */ }
```

**Steps:**
1. Create `internal/domain/operator/entity.go`
2. Create `internal/domain/operator/repository.go`
3. Create `internal/infrastructure/storage/operator.go` implementing interface
4. Rewrite handlers to use application layer
5. Delete old `pkg/storage/operators.go`

---

#### 4. `internal/api/handlers/auth_core.go`

**Current Problems:**
- Mixed login/register/logout in one file
- Handler does HTTP + business logic + storage
- Duplicated token generation

**Required Changes:**
```go
// BEFORE: auth_core.go (500+ lines, mixed)
func (ac *AuthController) Login(c *gin.Context) {
    // Parse request
    // Lookup operator
    // Verify password
    // Create session
    // Return response
}

func (ac *AuthController) Register(c *gin.Context) {
    // Similar structure
}
```

```go
// AFTER: Split into
// internal/api/handlers/auth/login.go
type LoginHandler struct { authService *application.AuthService }
func (h *LoginHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil { /* error */ }
    result, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
    if err != nil { /* handle error */ }
    setSessionCookie(c, result.Session)
    c.JSON(200, result.Operator)
}

// internal/application/auth/login.go
type AuthService struct {
    operatorRepo domain.OperatorRepository
    passwordHasher domain.PasswordHasher
    sessionManager domain.SessionManager
}
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
    // Business logic here
}
```

**Steps:**
1. Split into `login.go`, `register.go`, `logout.go`
2. Extract business logic to `internal/application/auth/`
3. Create domain interfaces in `internal/domain/`
4. Implement in `internal/infrastructure/`

---

#### 5. `pkg/storage/devices.go`

**Current Problems:**
- Plaintext command secret stored in DB ⚠️
- Lock held during slow DB operations
- UPDATE followed by SELECT (N+1)

**Required Changes:**
```go
// BEFORE: Line 86-105
secret, _ := randomHex(32)
secretHash, _ := HashSecret(secret)
// Stores BOTH plaintext AND hash!
db.ExecContext(ctx, `INSERT...command_secret, command_secret_hash... VALUES(?,?...)`, secret, secretHash)
```

```go
// AFTER: Store ONLY hash
secret, err := randomHex(32)
if err != nil { return err }
secretHash, err := HashSecret(secret)
if err != nil { return err }
// Store ONLY hash
db.ExecContext(ctx, `INSERT...command_secret_hash... VALUES(?)`, secretHash)
// Return plaintext secret ONCE to caller
return Device{CommandSecret: secret}, true, nil
```

**Steps:**
1. Remove plaintext storage line
2. Change INSERT to not include `command_secret` column
3. Verify callers only need hash for verification
4. Add transaction for atomicity

---

#### 6. `internal/api/handlers/auth_oauth.go`

**Current Problems:**
- Google and GitHub operator creation 90% identical
- Bootstrap check duplicated
- Frontend URL default duplicated 3x

**Required Changes:**
```go
// BEFORE: 2 almost identical blocks
// Google (lines 95-121)
op = &models.Operator{
    ID: GenerateID(),
    Email: googleClaims.Email,
    Name: googleClaims.Name,
    Role: role,
    GoogleID: googleClaims.Sub,
}

// GitHub (lines 251-276)
op = &models.Operator{
    ID: GenerateID(),
    Email: email,
    Name: name,
    Role: role,
    GitHubID: githubID,
}
```

```go
// AFTER: Single helper
// internal/application/auth/oauth.go
func (s *OAuthService) CreateOperator(ctx context.Context, email, name, provider, providerID string) (*Operator, error) {
    role, err := s.getRoleForNewOperator(ctx)
    if err != nil { return nil, err }
    
    op := &Operator{
        ID: GenerateID(),
        Email: email,
        Name: name,
        Role: role,
    }
    
    switch provider {
    case "google": op.GoogleID = providerID
    case "github": op.GitHubID = providerID
    }
    
    return s.operatorRepo.Create(ctx, op)
}
```

**Steps:**
1. Extract `createOAuthOperator()` helper
2. Extract `getRoleForNewOperator()` helper
3. Extract `getFrontendURL()` helper
4. Move to `internal/application/auth/oauth.go`

---

### 🟠 MEDIUM PRIORITY FILES

---

#### 7. `internal/api/handlers/auth_password_reset.go` + `auth_email_verify.go`

**Current Problems:**
- Token generation duplicated
- `context.Background()` in goroutines
- Silent failure when email not configured

**Required Changes:**
```go
// BEFORE: Duplicated token gen (3 locations)
tokenBytes := make([]byte, 32)
rand.Read(tokenBytes)
token := hex.EncodeToString(tokenBytes)
tokenHash := security.HashToken(token)
```

```go
// AFTER: Shared utility
// internal/application/shared/token.go
func GenerateSecureToken() (token, hash string, err error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", "", err
    }
    return hex.EncodeToString(b), HashToken(hex.EncodeToString(b)), nil
}
```

**Steps:**
1. Create `internal/application/shared/token.go`
2. Update both handlers to use shared utility
3. Fix goroutine context (use timeout context)
4. Fix silent failure (return success for enum prevention)

---

#### 8. `pkg/storage/clients.go`

**Current Problems:**
- JSON unmarshal pattern repeated 3x
- SigningKey scanning duplicated 2x
- `VerifyAPIClientSecret` returns same error for all failures

**Required Changes:**
```go
// BEFORE: Repeated 3x
if err := json.Unmarshal([]byte(allowedOrigins), &client.AllowedOrigins); err != nil {
    client.AllowedOrigins = []string{}
}
```

```go
// AFTER: Single helper
// internal/infrastructure/storage/client.go
func scanAPIClient(row Scannable) (*APIClient, error) {
    // Common scanning logic
    if err := json.Unmarshal([]byte(allowedOrigins), &client.AllowedOrigins); err != nil {
        client.AllowedOrigins = []string{}
    }
    return client, nil
}
```

---

#### 9. `pkg/storage/sessions.go`

**Current Problems:**
- N+1 queries in `RevokeAllOperatorSessions`
- No automatic cleanup for revocation list

**Required Changes:**
```go
// BEFORE: N+1 queries
for rows.Next() {
    var sessionID string
    rows.Scan(&sessionID)
    s.AddSessionRevocation(ctx, sessionID, "operator_logout") // INSERT per session
}
```

```go
// AFTER: Bulk INSERT
query := `INSERT INTO session_revocations (session_id, revoked_at, reason) VALUES `
for i, id := range sessionIDs {
    if i > 0 { query += "," }
    query += "(?, ?, ?)"
}
s.db.ExecContext(ctx, query, args...) // Single query
```

---

#### 10. `internal/api/middleware/user_enum.go` + `user_enum_block.go` + `auth_enum_safe.go`

**Current Problems:**
- Constant-time compare has 3 versions
- Fake password hash has 2 implementations

**Required Changes:**
```go
// AFTER: Consolidated
// internal/domain/auth/enum_safe.go
package auth

import "crypto/subtle"

func ConstantTimeCompare(a, b string) bool {
    return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func ComputeFakePasswordHash() {
    // Single implementation
}
```

---

### 🟡 LOW PRIORITY FILES

---

#### 11. `pkg/storage/uuid.go`

**Required Changes:**
```go
// FIX: hexCharToInt returns error for invalid chars
func hexCharToInt(c rune) (int, error) {
    switch {
    case c >= '0' && c <= '9':
        return int(c - '0'), nil
    // ...
    default:
        return 0, fmt.Errorf("invalid hex char: %c", c)
    }
}
```

---

#### 12. `pkg/storage/migrations.go`

**Required Changes:**
```go
// FIX: Log errors instead of swallowing
for _, q := range queries {
    if _, err := db.ExecContext(ctx, q); err != nil {
        log.Printf("migrate: %s failed: %v", q, err) // Log it!
    }
}
```

---

## Dependency Matrix

### Who Depends on What

```
main.go
  └─► pkg/config/config.go
  └─► pkg/storage/store.go (currently)
       └─► Will become internal/infrastructure/storage/

main.go
  └─► internal/api/handlers/server.go (currently)
       ├─► pkg/storage/operators.go
       ├─► pkg/storage/devices.go
       ├─► pkg/storage/clients.go
       ├─► internal/auth/session.go
       ├─► internal/auth/password.go
       ├─► internal/api/middleware/* (many)
       └─► Will become internal/api/router.go
```

### New Dependency Graph

```
cmd/api/main.go
  │
  ├─► pkg/config/
  │
  ├─► internal/api/router.go
  │    ├─► internal/api/middleware/* (presentation)
  │    └─► internal/api/handlers/* (presentation)
  │
  ├─► internal/application/
  │    ├─► internal/domain/* (interfaces only)
  │    └─► internal/infrastructure/* (implementations)
  │
  └─► internal/infrastructure/
       ├─► internal/infrastructure/storage/* (implements domain interfaces)
       ├─► internal/infrastructure/crypto/* (implements domain interfaces)
       ├─► internal/infrastructure/auth/* (implements domain interfaces)
       └─► internal/infrastructure/external/* (external services)
```

---

## Files to DELETE

| File | Reason |
|------|--------|
| `internal/api/handlers/auth_csrf.go` | Logic moved to middleware |
| `internal/api/handlers/auth_rate_limit.go` | Logic moved to middleware |
| `internal/api/handlers/lockout.go` | Logic moved to `internal/auth/lockout.go` |
| `internal/api/handlers/lockout_test.go` | Logic moved to `internal/auth/` |
| `internal/api/handlers/rate_limit_test.go` | Duplicated, move to middleware |

---

## Files to CREATE (New)

### Domain Layer (New Files)

| File | Purpose |
|------|---------|
| `internal/domain/errors.go` | Domain-level errors (ErrNotFound, etc.) |
| `internal/domain/operator/entity.go` | Operator struct |
| `internal/domain/operator/repository.go` | Repository interface |
| `internal/domain/device/entity.go` | Device struct |
| `internal/domain/device/repository.go` | Repository interface |
| `internal/domain/session/entity.go` | Session struct |
| `internal/domain/session/repository.go` | Repository interface |
| `internal/domain/command/entity.go` | Command struct |
| `internal/domain/command/repository.go` | Repository interface |
| `internal/domain/auth/enum_safe.go` | Consolidated enum protection |
| `internal/domain/auth/lockout.go` | Lockout interface |

### Application Layer (New Files)

| File | Purpose |
|------|---------|
| `internal/application/auth/login.go` | Login use case |
| `internal/application/auth/register.go` | Register use case |
| `internal/application/auth/logout.go` | Logout use case |
| `internal/application/auth/oauth.go` | OAuth use cases |
| `internal/application/auth/mfa.go` | MFA use cases |
| `internal/application/auth/password.go` | Password reset/use cases |
| `internal/application/device/register.go` | Device registration use case |
| `internal/application/device/command.go` | Command execution use case |
| `internal/application/shared/token.go` | Shared token generation |
| `internal/application/dto/` | Request/Response DTOs |
| `internal/application/errors.go` | Application errors |

### Infrastructure Layer (New Files)

| File | Purpose |
|------|---------|
| `internal/infrastructure/storage/sqlite.go` | DB connection management |
| `internal/infrastructure/storage/migrations/` | SQL migration files |
| `internal/infrastructure/crypto/aes_gcm.go` | AES-256-GCM utilities |
| `internal/infrastructure/crypto/replay_cache.go` | Replay protection |
| `internal/infrastructure/crypto/uuid.go` | UUID generation |
| `internal/infrastructure/auth/session.go` | Session management |
| `internal/infrastructure/auth/totp.go` | TOTP generation |
| `internal/infrastructure/auth/jwt.go` | JWT handling |
| `internal/infrastructure/external/email.go` | Email service |
| `internal/infrastructure/external/fcm.go` | FCM notifications |
| `internal/infrastructure/external/turnstile.go` | Turnstile verification |
| `internal/infrastructure/audit/logger.go` | Audit logging |
| `internal/infrastructure/metrics/prometheus.go` | Metrics |
| `internal/infrastructure/websocket/hub.go` | WebSocket hub |

---

## Files Requiring Significant Rewrite

| File | Changes | Effort |
|------|---------|--------|
| `server.go` → `router.go` | Extract handlers, simplify deps | High |
| `request_signing.go` | Split crypto to infrastructure | High |
| `auth_core.go` | Split into login/register/logout | High |
| `auth_oauth.go` | Extract OAuth helpers | Medium |
| `auth_password_reset.go` | Use shared token gen | Medium |
| `auth_email_verify.go` | Use shared token gen | Medium |
| `devices.go` | Remove plaintext, fix lock | Medium |
| `operators.go` | Implement domain interface | Medium |
| `clients.go` | Extract scanning helpers | Medium |
| `sessions.go` | Bulk INSERT | Medium |
| `user_enum*.go` | Consolidate to single file | Low |
| `uuid.go` | Fix hexCharToInt error | Low |
| `migrations.go` | Log errors properly | Low |
| `rate_limiter.go` | Minor cleanup | Low |

---

## Migration Sequence

### Phase 1: Create Domain Layer (Day 1)
1. Create `internal/domain/` structure
2. Extract entities: `Operator`, `Device`, `Session`, `Command`
3. Extract repository interfaces
4. **Result:** Code still works, domain exists

### Phase 2: Create Infrastructure (Day 2-3)
1. Move `pkg/storage/` → `internal/infrastructure/storage/`
2. Move `internal/auth/` → `internal/infrastructure/auth/`
3. Move `pkg/crypto/` → `internal/infrastructure/crypto/`
4. Implement domain interfaces
5. **Result:** Dependencies point inward

### Phase 3: Create Application Layer (Day 4-5)
1. Create `internal/application/` structure
2. Extract business logic from handlers
3. Handlers become thin wrappers
4. **Result:** Clean separation of concerns

### Phase 4: Refactor Router (Day 6)
1. Create `internal/api/router.go`
2. Split `server.go` into router + handlers
3. Simplify dependency injection
4. **Result:** Single entry point for routing

### Phase 5: Consolidate & Test (Day 7-8)
1. Remove all duplication
2. Run tests
3. Run linter
4. **Result:** Clean, working codebase

---

## Summary Checklist

```
PHASE 1 - Domain Layer
□ Create internal/domain/operator/entity.go
□ Create internal/domain/operator/repository.go
□ Create internal/domain/device/entity.go
□ Create internal/domain/device/repository.go
□ Create internal/domain/session/entity.go
□ Create internal/domain/session/repository.go
□ Create internal/domain/command/entity.go
□ Create internal/domain/command/repository.go
□ Create internal/domain/errors.go
□ Create internal/domain/auth/enum_safe.go

PHASE 2 - Infrastructure Layer
□ Move pkg/storage/ → internal/infrastructure/storage/
□ Move internal/auth/ → internal/infrastructure/auth/
□ Move pkg/crypto/ → internal/infrastructure/crypto/
□ Create internal/infrastructure/crypto/aes_gcm.go
□ Create internal/infrastructure/crypto/replay_cache.go
□ Implement all domain repository interfaces

PHASE 3 - Application Layer
□ Create internal/application/auth/
□ Create internal/application/device/
□ Create internal/application/shared/token.go
□ Extract business logic from handlers

PHASE 4 - Router Refactor
□ Create internal/api/router.go
□ Split server.go handlers into separate files
□ Simplify Server struct dependencies

PHASE 5 - Cleanup
□ Delete orphaned/dup files
□ Run golangci-lint
□ Run all tests
□ Verify build
```

---

*This document maps every Go file to its new location and required changes.*
