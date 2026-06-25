# Server Architecture Refactoring Plan

> **Version:** 1.0
> **Status:** Draft
> **Created:** 2026-06-25
> **Target:** Production MVP

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Current Architecture Analysis](#2-current-architecture-analysis)
3. [Target Architecture](#3-target-architecture)
4. [Violations Inventory](#4-violations-inventory)
5. [Refactoring Phases](#5-refactoring-phases)
6. [File-by-File Refactoring Guide](#6-file-by-file-refactoring-guide)
7. [Testing Strategy](#7-testing-strategy)
8. [Rollback Plan](#8-rollback-plan)

---

## 1. Executive Summary

### Problem

The current server architecture has **layering violations** where handlers directly import infrastructure packages, bypassing the application service layer. This creates tight coupling, makes testing difficult, and violates clean architecture principles.

### Impact

- **Testability:** Handlers cannot be unit tested without mocking infrastructure
- **Maintainability:** Changes to infrastructure require modifying handlers
- **Reusability:** Business logic is tied to HTTP handlers
- **Consistency:** No clear separation of concerns

### Solution

Refactor handlers to be thin wrappers that delegate to application services, which in turn use domain interfaces implemented by infrastructure.

### Effort Estimate

| Phase | Files | Complexity | Estimated Time |
|-------|-------|------------|----------------|
| Phase 1 | 1 | Medium | 2-4 hours |
| Phase 2 | 7 | High | 8-12 hours |
| Phase 3 | 2 | Medium | 4-6 hours |
| Phase 4 | 2 | Low | 2-3 hours |
| **Total** | **12** | - | **16-25 hours** |

---

## 2. Current Architecture Analysis

### 2.1 Intended Architecture (Clean)

```
┌─────────────────────────────────────────────────────────────────┐
│                         HANDLERS                                 │
│              (HTTP routing, request parsing)                    │
│                                                                 │
│  handlers/auth/register.go                                      │
│  handlers/device/list.go                                        │
│  handlers/command/execute.go                                   │
└─────────────────────────┬───────────────────────────────────────┘
                          │ calls
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                     APPLICATION LAYER                            │
│              (Business logic, orchestration)                    │
│                                                                 │
│  application/auth/service.go                                    │
│  application/device/service.go                                  │
│  application/command/service.go                                 │
└─────────────────────────┬───────────────────────────────────────┘
                          │ uses interfaces
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                       DOMAIN LAYER                              │
│               (Entities, interface definitions)                 │
│                                                                 │
│  domain/device/repository.go  (interface only)                   │
│  domain/command/repository.go  (interface only)                 │
└─────────────────────────┬───────────────────────────────────────┘
                          │ implemented by
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                   INFRASTRUCTURE LAYER                         │
│              (Database, external services)                      │
│                                                                 │
│  infrastructure/storage/device.go  (implements Repository)      │
│  infrastructure/fcm/notifier.go                                │
│  infrastructure/email/service.go                                │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Current Violations

```
┌─────────────────────────────────────────────────────────────────┐
│                         HANDLERS                                 │
│  ❌ handlers/auth/register.go ──────► infrastructure/email      │
│  ❌ handlers/auth/login.go ────────► infrastructure/security    │
│  ❌ handlers/telemetry_history.go ─► infrastructure/storage      │
│  ❌ handlers/websocket/handler.go ─► infrastructure/config      │
└─────────────────────────────────────────────────────────────────┘
```

---

## 3. Target Architecture

### 3.1 Handler Responsibilities (After Refactoring)

Handlers should ONLY:
- Parse HTTP request parameters
- Validate input format
- Call appropriate application service method
- Transform service response to HTTP response
- Return appropriate HTTP status codes

Handlers should NOT:
- Import infrastructure packages
- Access database directly
- Send emails or push notifications
- Perform business logic

### 3.2 Application Service Responsibilities

Application services should:
- Contain all business logic
- Coordinate between domain entities
- Call infrastructure through interfaces
- Handle transactions
- Emit domain events

### 3.3 Target File Structure

**Naming Convention:** Files are named descriptively to indicate their purpose:
- `auth_handler.go` (not `handler.go`)
- `auth_service.go` (not `service.go`)  
- `auth_repository.go` (not `repository.go`)
- `auth_dto.go` (not `dto.go`)

```
apps/api/internal/
├── api/
│   └── handlers/
│       ├── auth/
│       │   ├── auth_handler.go      # Thin handler - HTTP routing only
│       │   ├── auth_routes.go       # Route registration
│       │   └── auth_middleware.go    # Auth-specific middleware
│       ├── device/
│       │   ├── device_handler.go    # Thin handler
│       │   └── device_routes.go    # Route registration
│       ├── command/
│       │   ├── command_handler.go   # Thin handler
│       │   └── command_routes.go   # Route registration
│       └── websocket/
│           ├── websocket_handler.go # Thin handler
│           └── websocket_routes.go  # Route registration
├── application/
│   ├── auth/
│   │   ├── auth_service.go         # All auth business logic
│   │   ├── auth_dto.go             # Auth request/response DTOs
│   │   └── auth_errors.go          # Auth-specific errors
│   ├── device/
│   │   ├── device_service.go       # All device business logic
│   │   ├── device_dto.go           # Device request/response DTOs
│   │   └── device_errors.go        # Device-specific errors
│   ├── command/
│   │   ├── command_service.go      # All command business logic
│   │   ├── command_dto.go          # Command request/response DTOs
│   │   └── command_errors.go       # Command-specific errors
│   └── telemetry/
│       ├── telemetry_service.go    # All telemetry business logic
│       ├── telemetry_dto.go        # Telemetry request/response DTOs
│       └── telemetry_errors.go     # Telemetry-specific errors
├── domain/
│   ├── auth/
│   │   └── auth_domain.go          # Auth domain service (if needed)
│   ├── device/
│   │   ├── device_entity.go        # Device entity
│   │   ├── device_repository.go    # Repository interface
│   │   └── device_status.go       # Device status enum
│   ├── command/
│   │   ├── command_entity.go       # Command entity
│   │   └── command_repository.go   # Repository interface
│   └── telemetry/
│       ├── telemetry_entity.go     # Telemetry entity
│       └── telemetry_repository.go  # Repository interface
└── infrastructure/
    ├── storage/
    │   ├── device_storage.go       # Implements device.Repository
    │   ├── command_storage.go      # Implements command.Repository
    │   └── telemetry_storage.go    # Implements telemetry.Repository
    ├── fcm/
    │   └── fcm_notifier.go         # FCM notification implementation
    ├── email/
    │   └── email_service.go         # Email sending implementation
    └── security/
        └── security_service.go     # Security implementation
```

---

## 4. Violations Inventory

### 4.1 Summary Table

| File | Type | Infrastructure Import | Severity |
|------|------|----------------------|----------|
| `handlers/telemetry_history.go` | Direct DB | `storage` | 🔴 Critical |
| `handlers/auth/register.go` | Email/Security | `email`, `security` | 🟠 High |
| `handlers/auth/login.go` | Security | `security` | 🟠 High |
| `handlers/auth/mfa.go` | Security | `security` | 🟠 High |
| `handlers/auth/oauth.go` | Config/Security | `config`, `security` | 🟠 High |
| `handlers/auth/password_reset.go` | Email | `email` | 🟠 High |
| `handlers/auth/email_verify.go` | Email | `email` | 🟠 High |
| `handlers/auth/auth_routes.go` | Config/Email/Security | `config`, `email`, `security` | 🟡 Medium |
| `handlers/command/command_execute.go` | FCM | `fcm` | 🟡 Medium |
| `handlers/websocket/websocket_handler.go` | Config/Crypto | `config`, `crypto` | 🟡 Medium |
| `handlers/websocket/websocket_stream.go` | Config/Crypto/Security | `config`, `crypto`, `security` | 🟡 Medium |
| `handlers/updater/update_check.go` | Config | `config` | 🟢 Low |

### 4.2 Detailed Violation Analysis

#### `handlers/telemetry_history.go` 🔴 CRITICAL

**Current State:**
```go
import (
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
)

func (h *TelemetryHandler) Query(c *gin.Context) {
    // Directly uses storage.DB
    rows, err := storage.DB.Query(...)
}
```

**Target State:**
```go
import (
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/telemetry"
)

func (h *TelemetryHandler) Query(c *gin.Context) {
    frames, err := h.telemetryService.GetHistory(ctx, query)
}
```

---

#### Auth Handlers 🟠 HIGH

**Current Pattern:**
```
handlers/auth/register.go ──► infrastructure/email
                        └──► infrastructure/security
```

**Target Pattern:**
```
handlers/auth/register.go ──► application/auth/auth_service.go ──► infrastructure/email
                                                                         └──► infrastructure/security
```

---

## 5. Refactoring Phases

### Phase 1: Telemetry Handler (2-4 hours)

**Goal:** Eliminate direct storage access in telemetry handler

**Files:**
| Action | Current File | New File |
|--------|-------------|----------|
| MODIFY | `handlers/telemetry_history.go` | `handlers/telemetry_handler.go` |
| CREATE | - | `application/telemetry/telemetry_service.go` |
| CREATE | - | `domain/telemetry/telemetry_repository.go` (interface) |
| CREATE | - | `domain/telemetry/telemetry_entity.go` |
| MODIFY | `infrastructure/storage/telemetry.go` | `infrastructure/storage/telemetry_storage.go` |

**Steps:**
1. Define `telemetry.Repository` interface in `domain/telemetry/telemetry_repository.go`
2. Create `application/telemetry/telemetry_service.go`
3. Move storage logic from handler to service
4. Rename handler to `telemetry_handler.go` and update to call service
5. Update wire/wire_handlers.go

---

### Phase 2: Auth Handlers (8-12 hours)

**Goal:** Consolidate auth infrastructure calls in application service

**Files:**
| Action | Current File | New File |
|--------|-------------|----------|
| RENAME | `handlers/auth/register.go` | `handlers/auth/auth_register.go` |
| RENAME | `handlers/auth/login.go` | `handlers/auth/auth_login.go` |
| RENAME | `handlers/auth/mfa.go` | `handlers/auth/auth_mfa.go` |
| RENAME | `handlers/auth/oauth.go` | `handlers/auth/auth_oauth.go` |
| RENAME | `handlers/auth/password_reset.go` | `handlers/auth/password_reset.go` |
| RENAME | `handlers/auth/email_verify.go` | `handlers/auth/email_verify.go` |
| RENAME | `handlers/auth/routes.go` | `handlers/auth/auth_routes.go` |
| CREATE | - | `domain/auth/email_sender.go` (interface) |
| CREATE | - | `domain/auth/password_hasher.go` (interface) |
| CREATE | - | `domain/auth/token_generator.go` (interface) |
| MODIFY | `application/auth/service.go` | `application/auth/auth_service.go` |

**Steps:**
1. Extend `application/auth/auth_service.go` with methods for each handler's needs
2. Define interfaces for `EmailSender`, `PasswordHasher`, `TokenGenerator` in domain
3. Update handlers to call service methods
4. Remove direct infrastructure imports from handlers
5. Rename files to use descriptive naming

---

### Phase 3: WebSocket Handlers (4-6 hours)

**Goal:** Clean up WebSocket infrastructure coupling

**Files:**
| Action | Current File | New File |
|--------|-------------|----------|
| RENAME | `handlers/websocket/handler.go` | `handlers/websocket/websocket_handler.go` |
| RENAME | `handlers/websocket/stream_upgrade.go` | `handlers/websocket/websocket_stream.go` |
| CREATE | - | `application/websocket/websocket_service.go` |

**Steps:**
1. Create `application/websocket/websocket_service.go`
2. Move crypto/config logic to initialization (wire package)
3. Rename handlers to descriptive names
4. Keep handlers thin

---

### Phase 4: Remaining Handlers (2-3 hours)

**Goal:** Fix remaining violations

**Files:**
| Action | Current File | New File |
|--------|-------------|----------|
| RENAME | `handlers/command/execute.go` | `handlers/command/command_execute.go` |
| RENAME | `handlers/updater/updater.go` | `handlers/updater/update_check.go` |
| MODIFY | - | Extend `application/command/command_service.go` |
| MODIFY | - | Extend `application/updater/update_service.go` |

**Steps:**
1. Extend `application/command/command_service.go` with FCM notification logic
2. Extend `application/updater/update_service.go` with config logic
3. Rename handlers to descriptive names
4. Update handlers to use service methods

---

## 6. File-by-File Refactoring Guide

### 6.1 `handlers/telemetry_history.go` → `handlers/telemetry_handler.go`

**Current imports:**
```go
import (
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
)
```

**Refactored imports:**
```go
import (
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/telemetry"
)
```

**Handler struct before:**
```go
type TelemetryHandler struct{}
```

**Handler struct after:**
```go
type TelemetryHandler struct {
    telemetryService *telemetry.Service
}
```

---

### 6.2 `handlers/auth/register.go` → `handlers/auth/auth_register.go`

**Current:**
```go
import (
    emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
    infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
)

func Register(c *gin.Context) {
    hash, _ := infraauth.HashPassword(req.Password)
    emailService.Send(c.Request.Context(), to, subject, body)
}
```

**Target:**
```go
import (
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
)

func (h *AuthHandler) Register(c *gin.Context) {
    err := h.authService.Register(c.Request.Context(), &req)
    // Handler just returns response
}
```

---

### 6.3 `handlers/auth/auth_routes.go`

**Current:**
```go
import (
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/config"
    emailService "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/email"
    infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
)

func RegisterRoutes(rg *gin.RouterGroup, ...) {
    cfg := config.Load()
    email := emailService.New(cfg.Email)
    security := infraauth.New(cfg.Security)
    // ... create handlers with infrastructure
}
```

**Target:**
Move infrastructure initialization to wire package:
```go
// wire/wire_auth.go
func InitializeAuthHandler(cfg *config.Config) *AuthHandler {
    emailService := email.NewEmailService(cfg.Email)
    security := security.NewSecurityService(cfg.Security)
    authService := auth.NewAuthService(emailService, security, ...)
    return auth.NewAuthHandler(authService)
}
```

---

## 7. Testing Strategy

### 7.1 Before Refactoring (Current State)

**Handler tests require mocking infrastructure:**
```go
func TestRegisterHandler(t *testing.T) {
    // Must mock infrastructure/email
    // Must mock infrastructure/security
    mockEmail := email.NewMock()
    mockSecurity := security.NewMock()
    handler := NewRegisterHandler(mockEmail, mockSecurity)
}
```

### 7.2 After Refactoring (Target State)

**Handler tests only mock application service:**
```go
func TestRegisterHandler(t *testing.T) {
    // Only mock application service
    mockService := auth.NewMockService()
    handler := NewAuthHandler(mockService)
}
```

**Application service tests mock infrastructure:**
```go
func TestAuthService(t *testing.T) {
    mockEmail := email.NewMock()
    mockSecurity := security.NewMock()
    mockRepo := operator.NewMockRepo()
    service := auth.NewService(mockEmail, mockSecurity, mockRepo)
}
```

### 7.3 Test Coverage Goals

| Layer | Before | After |
|-------|--------|-------|
| Handlers | ~40% | ~80% |
| Application | ~60% | ~90% |
| Infrastructure | ~50% | ~70% |

---

## 8. Rollback Plan

### 8.1 Per-Phase Rollback

Each phase can be rolled back independently by:
1. Reverting handler file changes
2. Keeping new application service files (they don't break anything)
3. Re-adding infrastructure imports to handlers

### 8.2 Git Branch Strategy

```bash
# Main refactoring branch
git checkout -b refactor/layered-architecture

# Phase branches
git checkout -b refactor/phase1-telemetry
git checkout -b refactor/phase2-auth
git checkout -b refactor/phase3-websocket
git checkout -b refactor/phase4-misc
```

### 8.3 Rollback Commands

```bash
# Phase 1 rollback
git checkout HEAD~1 -- handlers/telemetry_history.go
git checkout HEAD~1 -- application/telemetry/
git checkout HEAD~1 -- domain/telemetry/

# Full rollback
git checkout main -- handlers/
```

---

## Appendix A: Domain Interfaces Required

```go
// domain/telemetry/telemetry_repository.go
package telemetry

type Repository interface {
    FindByDeviceID(ctx context.Context, deviceID string, query *HistoryQuery) ([]*TelemetryFrame, error)
    GetStats(ctx context.Context, deviceID string, start, end time.Time) (*Stats, error)
}

// domain/auth/email_sender.go  
package auth

type EmailSender interface {
    Send(ctx context.Context, to, subject, body string) error
}

// domain/auth/password_hasher.go
package auth

type PasswordHasher interface {
    Hash(password string) (string, error)
    Compare(hash, password string) bool
}

// domain/auth/token_generator.go
package auth

type TokenGenerator interface {
    GenerateToken(op *Operator) (string, error)
}
```

---

## Appendix B: Affected File List

| Phase | Current File | New File | Action |
|-------|-------------|----------|--------|
| 1 | `handlers/telemetry_history.go` | `handlers/telemetry_handler.go` | RENAME - remove storage import |
| 1 | - | `application/telemetry/telemetry_service.go` | NEW - create |
| 1 | - | `domain/telemetry/telemetry_repository.go` | NEW - create interface |
| 1 | - | `domain/telemetry/telemetry_entity.go` | NEW - create entity |
| 1 | `infrastructure/storage/telemetry.go` | `infrastructure/storage/telemetry_storage.go` | RENAME - implement interface |
| 2 | `handlers/auth/register.go` | `handlers/auth/auth_register.go` | RENAME |
| 2 | `handlers/auth/login.go` | `handlers/auth/auth_login.go` | RENAME |
| 2 | `handlers/auth/mfa.go` | `handlers/auth/auth_mfa.go` | RENAME |
| 2 | `handlers/auth/oauth.go` | `handlers/auth/auth_oauth.go` | RENAME |
| 2 | `handlers/auth/password_reset.go` | `handlers/auth/password_reset.go` | RENAME (keep) |
| 2 | `handlers/auth/email_verify.go` | `handlers/auth/email_verify.go` | RENAME (keep) |
| 2 | `handlers/auth/routes.go` | `handlers/auth/auth_routes.go` | RENAME |
| 2 | `application/auth/service.go` | `application/auth/auth_service.go` | RENAME |
| 2 | - | `domain/auth/email_sender.go` | NEW - interface |
| 2 | - | `domain/auth/password_hasher.go` | NEW - interface |
| 2 | - | `domain/auth/token_generator.go` | NEW - interface |
| 3 | `handlers/websocket/handler.go` | `handlers/websocket/websocket_handler.go` | RENAME |
| 3 | `handlers/websocket/stream_upgrade.go` | `handlers/websocket/websocket_stream.go` | RENAME |
| 3 | - | `application/websocket/websocket_service.go` | NEW - create service |
| 4 | `handlers/command/execute.go` | `handlers/command/command_execute.go` | RENAME |
| 4 | `handlers/updater/updater.go` | `handlers/updater/update_check.go` | RENAME |
| 4 | `application/command/service.go` | `application/command/command_service.go` | RENAME - extend |
| 4 | - | `application/updater/update_service.go` | NEW - extend |

---

*Document Version: 1.0*
*Status: Ready for Implementation*
