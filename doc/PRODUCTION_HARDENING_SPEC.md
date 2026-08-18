
# Vyzorix Production Hardening Specification

> **Purpose:** Document the refactoring plan to add production-grade security, error handling, risk management, and hardening to the Vyzorix Update Server.

---

## Table of Contents

1. [Current State Analysis](#1-current-state-analysis)
2. [New Architecture Patterns](#2-new-architecture-patterns)
3. [File Refactoring Map](#3-file-refactoring-map)
4. [New File Registry](#4-new-file-registry)
5. [Naming Conventions](#5-naming-conventions)
6. [Implementation Phases](#6-implementation-phases)
7. [Migration Strategy](#7-migration-strategy)

---

## 1. Current State Analysis

### 1.1 Error Handling (Current)

**Location:** `apps/api/internal/domain/domain_errors.go`

**Current Pattern:**
```go
var (
    ErrNotFound = errors.New("entity not found")
    ErrAlreadyExists = errors.New("entity already exists")
    ErrInvalidInput = errors.New("invalid input")
    // ... flat, non-hierarchical
)
```

**Problems:**
- No error codes for client consumption
- No structured metadata
- No trace IDs
- No validation context
- Cannot distinguish between sub-types programmatically

**Impact:** ~45 handlers need refactoring for consistent error responses

---

### 1.2 Security Infrastructure (Existing вњ…)

| Component | Location | Status |
|-----------|----------|--------|
| JWT/Session | `security/jwt/`, `security/session/` | вњ… Complete |
| Password Hashing | `security/password/argon2_hasher.go` | вњ… Complete |
| Rate Limiting | `security/ratelimit/` | вњ… Complete |
| MFA (TOTP) | `security/totp/` | вњ… Complete |
| Account Lockout | `security/lockout/` | вњ… Complete |
| Request Signing | `security/request_signer/` | вњ… Complete |
| Threat Detection | `domain/threat/` | вњ… Complete |

---

### 1.3 Logging (Current)

**Location:** Various handlers using `slog`

**Problems:**
- No secret redaction
- No correlation/trace IDs
- No structured audit trail
- Inconsistent field naming

---

### 1.4 Command Execution (Current)

**Location:** `domain/command/`

**Problems:**
- No risk classification
- No confirmation flow for dangerous commands
- No audit trail for executed commands

---

## 2. New Architecture Patterns

### 2.1 Hierarchical Error System

**Pattern:**
```go
// Error codes are constants, not strings
type ErrorCode string

const (
    // Authentication
    CodeAuthInvalidCredentials ErrorCode = "AUTH_INVALID_CREDENTIALS"
    CodeAuthAccountLocked     ErrorCode = "AUTH_ACCOUNT_LOCKED"
    CodeAuthMFARequired      ErrorCode = "AUTH_MFA_REQUIRED"
    CodeAuthSessionExpired    ErrorCode = "AUTH_SESSION_EXPIRED"
    
    // Authorization
    CodeAuthzInsufficientPermissions ErrorCode = "AUTHZ_INSUFFICIENT_PERMISSIONS"
    CodeAuthzOrgMembershipRequired  ErrorCode = "AUTHZ_ORG_MEMBERSHIP_REQUIRED"
    
    // Resource
    CodeResourceNotFound       ErrorCode = "RESOURCE_NOT_FOUND"
    CodeResourceAlreadyExists  ErrorCode = "RESOURCE_ALREADY_EXISTS"
    CodeResourceConflict       ErrorCode = "RESOURCE_CONFLICT"
    
    // Validation
    CodeValidationFailed      ErrorCode = "VALIDATION_FAILED"
    CodeValidationRequired    ErrorCode = "VALIDATION_REQUIRED_FIELD"
    
    // Rate Limiting
    CodeRateLimitExceeded     ErrorCode = "RATE_LIMIT_EXCEEDED"
    
    // Security
    CodeSecurityThreatDetected ErrorCode = "SECURITY_THREAT_DETECTED"
    CodeSecurityRiskUnconfirmed ErrorCode = "SECURITY_RISK_UNCONFIRMED"
    
    // Internal
    CodeInternalServerError  ErrorCode = "INTERNAL_SERVER_ERROR"
)

// Structured error with context
type ServerError struct {
    Code       ErrorCode              `json:"code"`
    Message    string                `json:"message"`    // Safe for client
    Details    interface{}           `json:"details,omitempty"`  // Validation errors, etc.
    TraceID    string               `json:"trace_id"`
    Timestamp  time.Time            `json:"timestamp"`
    DocsURL    string               `json:"docs_url,omitempty"`
}

// Validation errors have field-level detail
type ValidationDetail struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Code    string `json:"code,omitempty"`
}
```

---

### 2.2 Risk Classification System

**Pattern:**
```go
// Risk levels for operations
type RiskTier string

const (
    RiskTierZero  RiskTier = "zero"   // Read operations, no side effects
    RiskTierLow   RiskTier = "low"    // Minor side effects, reversible
    RiskTierMedium RiskTier = "medium" // User data affected
    RiskTierHigh  RiskTier = "high"   // System config, org-wide
    RiskTierCritical RiskTier = "critical" // Destructive, irreversible
)

// Command risk profile
type CommandRiskProfile struct {
    Tier            RiskTier
    RequiresMFA     bool
    RequiresOrgAdmin bool
    AuditRequired   bool
    ConfirmationTTL time.Duration // How long confirmation lasts
}

// Predefined profiles
var CommandRiskRegistry = map[string]CommandRiskProfile{
    "device.status": {
        Tier: RiskTierZero,
        AuditRequired: false,
    },
    "device.reboot": {
        Tier: RiskTierHigh,
        RequiresOrgAdmin: true,
        AuditRequired: true,
        ConfirmationTTL: 5 * time.Minute,
    },
    "device.factory_reset": {
        Tier: RiskTierCritical,
        RequiresMFA: true,
        RequiresOrgAdmin: true,
        AuditRequired: true,
        ConfirmationTTL: 2 * time.Minute,
    },
}
```

---

### 2.3 Secret Redaction System

**Pattern:**
```go
// Redaction patterns for log sanitization
type SecretPattern struct {
    Name     string
    Pattern  *regexp.Regexp
    RedactWith string
}

var RedactionPatterns = []SecretPattern{
    {
        Name:    "api_key",
        Pattern: regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[=:]\s*["']?([\w-]{8,})["']?`),
        RedactWith: "$1=[REDACTED]",
    },
    {
        Name:    "bearer_token",
        Pattern: regexp.MustCompile(`(?i)bearer\s+([\w-]{20,})`),
        RedactWith: "Bearer [REDACTED]",
    },
    {
        Name:    "authorization",
        Pattern: regexp.MustCompile(`(?i)authorization\s*:\s*([^\s]+)`),
        RedactWith: "Authorization: [REDACTED]",
    },
}

// SanitizedLogEntry - what goes into logs
type SanitizedLogEntry struct {
    TraceID    string                 `json:"trace_id"`
    Timestamp  time.Time              `json:"timestamp"`
    Level      string                 `json:"level"`
    Message    string                 `json:"message"`
    ActorID    string                 `json:"actor_id,omitempty"`
    OrgID      string                 `json:"org_id,omitempty"`
    Action     string                 `json:"action"`
    DurationMs int64                  `json:"duration_ms,omitempty"`
    Metadata   map[string]interface{} `json:"metadata,omitempty"`
}
```

---

### 2.4 Audit Event Schema

**Pattern:**
```go
// Audit events for compliance
type AuditEvent struct {
    ID          string                 `json:"id"`
    Timestamp   time.Time             `json:"timestamp"`
    TraceID     string                `json:"trace_id"`
    
    // Who
    ActorID     string                `json:"actor_id"`
    ActorType   ActorType            `json:"actor_type"`  // operator, api_key, system
    ActorEmail  string               `json:"actor_email,omitempty"`
    
    // Where
    OrgID       string                `json:"org_id,omitempty"`
    IPAddress   string               `json:"ip_address"`
    UserAgent   string              `json:"user_agent,omitempty"`
    
    // What
    Action      string               `json:"action"`
    Resource    string               `json:"resource"`
    ResourceID  string              `json:"resource_id,omitempty"`
    RiskTier    RiskTier            `json:"risk_tier,omitempty"`
    
    // Result
    Result      AuditResult         `json:"result"`  // success, failure, blocked, skipped
    Reason      string              `json:"reason,omitempty"`
    
    // Change tracking
    OldValue    interface{}         `json:"old_value,omitempty"`
    NewValue    interface{}         `json:"new_value,omitempty"`
}

type ActorType string
const (
    ActorOperator ActorType = "operator"
    ActorAPIKey   ActorType = "api_key"
    ActorSystem   ActorType = "system"
)

type AuditResult string
const (
    AuditSuccess AuditResult = "success"
    AuditFailure AuditResult = "failure"
    AuditBlocked AuditResult = "blocked"
    AuditSkipped AuditResult = "skipped"
)

// Predefined action constants
const (
    AuditActionDeviceRegister     = "device.register"
    AuditActionDeviceDeregister  = "device.deregister"
    AuditActionDeviceCommand      = "device.command"
    AuditActionCommandConfirmed   = "device.command.confirmed"
    AuditActionCommandBlocked     = "device.command.blocked"
    AuditActionDeviceSettings     = "device.settings.update"
    AuditActionLoginSuccess       = "auth.login.success"
    AuditActionLoginFailed        = "auth.login.failed"
    AuditActionOrgMemberInvite    = "org.member.invite"
    AuditActionOrgMemberRemove    = "org.member.remove"
)
```

---

### 2.5 Confirmation Flow

**Pattern:**
```go
// Pending confirmations for risky operations
type PendingConfirmation struct {
    ID          string           `json:"id"`
    TraceID     string           `json:"trace_id"`
    
    ActorID     string           `json:"actor_id"`
    ActorType   ActorType       `json:"actor_type"`
    OrgID       string           `json:"org_id"`
    
    Operation   string           `json:"operation"`
    TargetType  string           `json:"target_type"`  // device, org, etc.
    TargetID    string           `json:"target_id"`
    Parameters  interface{}     `json:"parameters,omitempty"`
    
    RiskTier    RiskTier        `json:"risk_tier"`
    CreatedAt   time.Time       `json:"created_at"`
    ExpiresAt   time.Time       `json:"expires_at"`
    Status      ConfirmationStatus `json:"status"`
}

type ConfirmationStatus string
const (
    ConfirmationPending   ConfirmationStatus = "pending"
    ConfirmationConfirmed ConfirmationStatus = "confirmed"
    ConfirmationExpired   ConfirmationStatus = "expired"
    ConfirmationRejected ConfirmationStatus = "rejected"
    ConfirmationUsed     ConfirmationStatus = "used"
)
```

---

## 3. File Refactoring Map

### 3.1 Refactor Existing Files

| Current File | New File | Changes |
|--------------|----------|---------|
| `domain/domain_errors.go` | `domain/errors/codes.go` | Add error code constants, replace flat errors |
| `domain/domain_errors.go` | `domain/errors/errors.go` | Add ServerError struct, structured error creation |
| `api/middleware/error.go` | `api/middleware/error_handler.go` | Use new error system, add trace ID |
| `api/responses/api_errors.go` | `api/responses/errors.go` | Use new structured error format |
| `domain/command/command_entity.go` | `domain/command/risk.go` | Add risk profiles, merge into entity |
| `domain/threat/threat_types.go` | `domain/security/threat.go` | Move to security folder, enhance |
| `infrastructure/security/security.go` | `infrastructure/security/security.go` | Keep, add new security components |

### 3.2 Merge Duplicate Files

| Files To Merge | Target |
|----------------|--------|
| `domain/operator/operator_errors.go` | `domain/errors/codes.go` |
| `domain/organization/organization_errors.go` | `domain/errors/codes.go` |
| `domain/device/device_errors.go` (if exists) | `domain/errors/codes.go` |

---

## 4. New File Registry

### 4.1 Error System (3 files)

| Filename | Purpose |
|----------|---------|
| `domain/errors/codes.go` | All ErrorCode constants, organized by category |
| `domain/errors/errors.go` | ServerError struct, constructor functions |
| `domain/errors/validation.go` | Validation-specific error helpers |

### 4.2 Security/Risk System (4 files)

| Filename | Purpose |
|----------|---------|
| `domain/security/threat_profile.go` | Threat evaluation, moved from domain/threat |
| `domain/security/risk_catalog.go` | Command risk profiles, risk tier definitions |
| `domain/security/risk_evaluator.go` | Risk evaluation logic |
| `domain/audit/audit_event.go` | Audit event struct and constants |

### 4.3 Confirmation System (2 files)

| Filename | Purpose |
|----------|---------|
| `domain/confirmation/pending_confirmation.go` | Pending confirmation model |
| `domain/confirmation/confirmation_store.go` | Confirmation repository interface |

### 4.4 Infrastructure (3 files)

| Filename | Purpose |
|----------|---------|
| `infrastructure/redaction/redact.go` | Secret redaction logic |
| `infrastructure/tracing/trace_context.go` | Trace ID middleware and context |
| `infrastructure/audit/audit_logger.go` | Audit event persistence |

### 4.5 Middleware Enhancement (2 files)

| Filename | Purpose |
|----------|---------|
| `api/middleware/tracing.go` | Add trace ID to all requests |
| `api/middleware/audit_trail.go` | Automatic audit logging |

### 4.6 API Responses (1 file)

| Filename | Purpose |
|----------|---------|
| `api/responses/structured_error.go` | Structured error response formatting |

---

## 5. Naming Conventions

### 5.1 Error Code Naming

**Format:** `CATEGORY_SPECIFIC`

| Category | Prefix | Example |
|----------|--------|---------|
| Authentication | `AUTH_` | `AUTH_INVALID_CREDENTIALS` |
| Authorization | `AUTHZ_` | `AUTHZ_INSUFFICIENT_PERMISSIONS` |
| Resource | `RESOURCE_` | `RESOURCE_NOT_FOUND` |
| Validation | `VALIDATION_` | `VALIDATION_REQUIRED_FIELD` |
| Rate Limiting | `RATE_LIMIT_` | `RATE_LIMIT_EXCEEDED` |
| Security | `SECURITY_` | `SECURITY_THREAT_DETECTED` |
| Internal | `INTERNAL_` | `INTERNAL_SERVER_ERROR` |

### 5.2 Audit Action Naming

**Format:** `resource.verb` (lowercase, dot-separated)

| Resource | Verbs |
|----------|-------|
| device | `register`, `deregister`, `command`, `settings.update` |
| auth | `login.success`, `login.failed`, `logout`, `mfa.attempt`, `mfa.success` |
| org | `member.invite`, `member.remove`, `settings.update` |
| api_key | `create`, `revoke`, `rotate` |

### 5.3 Risk Tier Naming

**Format:** `RiskTier` + Tier name (camelCase values)

| Value | Usage |
|-------|-------|
| `RiskTierZero` | Safe read operations |
| `RiskTierLow` | Minor side effects |
| `RiskTierMedium` | User data affected |
| `RiskTierHigh` | System/org config |
| `RiskTierCritical` | Destructive operations |

### 5.4 File Naming Rules

1. **No generic names** like `errors.go`, `security.go`, `helper.go`
2. **Specific purpose** in filename: `codes.go` is OK if folder is `errors/`
3. **Use underscores** for multi-word: `risk_catalog.go`, not `riskcatalog.go`
4. **Match domain**: `threat_profile.go` not `threat.go`

---

## 6. Implementation Phases

### Phase 1: Foundation (Critical Path)

**Goal:** Error system, tracing, redacation - these underpin everything else

| # | Task | Files | Priority |
|---|------|-------|----------|
| 1 | Create error code constants | `domain/errors/codes.go` | P0 |
| 2 | Create ServerError struct | `domain/errors/errors.go` | P0 |
| 3 | Add trace ID middleware | `api/middleware/tracing.go` | P0 |
| 4 | Create redaction logic | `infrastructure/redaction/redact.go` | P0 |
| 5 | Update error responses | `api/responses/structured_error.go` | P0 |
| 6 | Update error middleware | `api/middleware/error_handler.go` | P0 |

### Phase 2: Risk & Audit

**Goal:** Command risk classification and audit trail

| # | Task | Files | Priority |
|---|------|-------|----------|
| 1 | Create risk catalog | `domain/security/risk_catalog.go` | P1 |
| 2 | Create risk evaluator | `domain/security/risk_evaluator.go` | P1 |
| 3 | Create audit event schema | `domain/audit/audit_event.go` | P1 |
| 4 | Create audit logger | `infrastructure/audit/audit_logger.go` | P1 |
| 5 | Add audit middleware | `api/middleware/audit_trail.go` | P1 |

### Phase 3: Confirmation Flow

**Goal:** Dangerous operations require confirmation

| # | Task | Files | Priority |
|---|------|-------|----------|
| 1 | Create confirmation model | `domain/confirmation/pending_confirmation.go` | P2 |
| 2 | Create confirmation store | `domain/confirmation/confirmation_store.go` | P2 |
| 3 | Add confirmation endpoints | New handlers in `api/handlers/confirmation/` | P2 |

### Phase 4: Validation Errors

**Goal:** Better validation error messages

| # | Task | Files | Priority |
|---|------|-------|----------|
| 1 | Create validation errors | `domain/errors/validation.go` | P3 |
| 2 | Update validators | Existing validator files | P3 |

---

## 7. Migration Strategy

### 7.1 Backward Compatibility

1. **New errors wrap old errors** - `ServerError` can contain original error
2. **Old handlers still work** - Middleware translates automatically
3. **Gradual migration** - One module at a time

### 7.2 Migration Steps

**Step 1:** Create new error system files (Phase 1)
- Creates `domain/errors/` with new structure
- Existing code still uses old `domain_errors.go`

**Step 2:** Add middleware
- New tracing middleware wraps all requests
- New error middleware translates old errors

**Step 3:** Migrate handlers one by one
- Pick one handler, update to return `ServerError`
- Test, verify, commit
- Repeat

**Step 4:** Delete old error file
- When all handlers migrated, remove `domain_errors.go`
- Update imports

### 7.3 Testing Strategy

1. **Unit tests** for error creation functions
2. **Integration tests** for error middleware
3. **Contract tests** for error response format
4. **Redaction tests** for secret patterns

---

## Appendix A: File Tree (After Refactoring)

```
apps/api/internal/
в”њв”Ђв”Ђ domain/
в”‚   в”њв”Ђв”Ђ errors/                          # NEW
в”‚   в”‚   в”њв”Ђв”Ђ codes.go                     # ErrorCode constants
в”‚   в”‚   в”њв”Ђв”Ђ errors.go                    # ServerError struct
в”‚   в”‚   в””в”Ђв”Ђ validation.go                # Validation helpers
в”‚   в”њв”Ђв”Ђ command/
в”‚   в”‚   в”њв”Ђв”Ђ command_entity.go            # Keep, enhance with risk
в”‚   в”‚   в”њв”Ђв”Ђ risk_catalog.go              # NEW: Command risk profiles
в”‚   в”‚   в””в”Ђв”Ђ risk_evaluator.go           # NEW: Risk evaluation
в”‚   в”њв”Ђв”Ђ confirmation/                    # NEW
в”‚   в”‚   в”њв”Ђв”Ђ pending_confirmation.go      # Confirmation model
в”‚   в”‚   в””в”Ђв”Ђ confirmation_store.go       # Confirmation repo interface
в”‚   в”њв”Ђв”Ђ audit/                          # NEW
в”‚   в”‚   в””в”Ђв”Ђ audit_event.go              # Audit event schema
в”‚   в”њв”Ђв”Ђ security/                        # ENHANCE
в”‚   в”‚   в”њв”Ђв”Ђ threat_profile.go           # Moved from domain/threat
в”‚   в”‚   в””в”Ђв”Ђ ...
в”‚   в”њв”Ђв”Ђ operator/
в”‚   в”‚   в””в”Ђв”Ђ operator_errors.go          # DELETE - merged to errors/
в”‚   в”њв”Ђв”Ђ organization/
в”‚   в”‚   в””в”Ђв”Ђ organization_errors.go       # DELETE - merged to errors/
в”‚   в””в”Ђв”Ђ ...
в”‚
в”њв”Ђв”Ђ application/
в”‚   в”њв”Ђв”Ђ confirmation/                    # NEW
в”‚   в”‚   в””в”Ђв”Ђ confirmation_service.go     # Confirmation business logic
в”‚   в”њв”Ђв”Ђ risk/                          # NEW
в”‚   в”‚   в””в”Ђв”Ђ risk_service.go            # Risk evaluation service
в”‚   в””в”Ђв”Ђ audit/                         # NEW
в”‚       в””в”Ђв”Ђ audit_service.go            # Audit logging service
в”‚
в”њв”Ђв”Ђ infrastructure/
в”‚   в”њв”Ђв”Ђ redaction/                      # NEW
в”‚   в”‚   в””в”Ђв”Ђ redact.go                  # Secret redaction
в”‚   в”њв”Ђв”Ђ tracing/                        # NEW
в”‚   в”‚   в””в”Ђв”Ђ trace_context.go           # Trace ID handling
в”‚   в”њв”Ђв”Ђ audit/                          # NEW
в”‚   в”‚   в””в”Ђв”Ђ audit_logger.go            # Audit persistence
в”‚   в”њв”Ђв”Ђ security/
в”‚   в”‚   в””в”Ђв”Ђ ...                        # Keep existing
в”‚   в””в”Ђв”Ђ storage/
в”‚       в””в”Ђв”Ђ ...                        # Keep existing
в”‚
в””в”Ђв”Ђ api/
    в”њв”Ђв”Ђ middleware/
    в”‚   в”њв”Ђв”Ђ tracing.go                 # NEW
    в”‚   в”њв”Ђв”Ђ audit_trail.go             # NEW
    в”‚   в”њв”Ђв”Ђ error_handler.go            # REFACTORED
    в”‚   в””в”Ђв”Ђ ...                        # Keep existing
    в”‚
    в”њв”Ђв”Ђ responses/
    в”‚   в”њв”Ђв”Ђ structured_error.go        # NEW
    в”‚   в””в”Ђв”Ђ ...                        # Keep existing
    в”‚
    в””в”Ђв”Ђ handlers/
        в”њв”Ђв”Ђ confirmation/               # NEW
        в”‚   в””в”Ђв”Ђ confirmation_handler.go
        в””в”Ђв”Ђ ...
```

---

## Appendix B: Import Paths Reference

After refactoring, new import paths:

```go
// Errors
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"

// Security/Risk
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/security"

// Confirmation
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/confirmation"

// Audit
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/audit"

// Infrastructure
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/redaction"
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/tracing"
"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/audit"
```

---

## Appendix C: Status

- [x] Phase 1: Foundation вњ…
- [x] Phase 2: Risk & Audit вњ…
- [x] Phase 3: Confirmation Flow вњ…
- [x] Phase 4: Validation Errors вњ…

---

## Appendix D: Phase 1 Implementation Summary

### Files Created

| File | Purpose | Lines |
|------|---------|-------|
| `domain/errors/codes.go` | ErrorCode constants (50+ codes) | ~300 |
| `domain/errors/codes_test.go` | Tests for error codes | ~120 |
| `domain/errors/errors.go` | ServerError struct + factory functions | ~280 |
| `domain/errors/errors_test.go` | Tests for error creation | ~150 |
| `infrastructure/tracing/trace_context.go` | Trace ID generation & validation | ~150 |
| `infrastructure/tracing/trace_context_test.go` | Tests for tracing | ~120 |
| `infrastructure/tracing/docs_url.go` | Dynamic docs URL builder | ~50 |
| `infrastructure/redaction/redact.go` | Secret redaction for logs | ~280 |
| `infrastructure/redaction/redact_test.go` | Tests for redaction | ~100 |
| `api/middleware/tracing.go` | Trace ID middleware | ~50 |
| `api/middleware/error.go` | Refactored error handler | ~175 |
| `api/responses/structured_error.go` | Structured error responses | ~277 |

### Key Features Implemented

1. **Error Code System**: 50+ categorized error codes (AUTH_*, AUTHZ_*, RESOURCE_*, etc.)
2. **Structured ServerError**: Trace ID, timestamp, details, docs URL (dynamic)
3. **Trace ID Middleware**: Auto-generates/extracts X-Trace-ID header
4. **Secret Redaction**: API keys, JWTs, passwords, DB connections sanitized from logs
5. **Factory Functions**: `ErrNotFound()`, `ErrInvalidCredentials()`, `ErrRateLimitExceeded()`, etc.
6. **Dynamic Docs URLs**: Built from server context, not hardcoded

**Created:** 2026-08-17
**Author:** VinnsEdesigner / OpenHands Analysis

---

## Appendix E: Phase 2 Implementation Summary

### Goal

Command risk classification and an audit trail for every command execution
attempt вЂ” allowed, confirmation-gated, or blocked вЂ” so risky operations are
traceable and dangerous commands cannot fire silently.

### Files Created

| File | Purpose |
|------|---------|
| `domain/command/risk_catalog.go` | `RiskTier`, `CommandRiskProfile`, and the catalog of known commands + their profiles |
| `domain/command/risk_evaluator.go` | `RiskEvaluator` combining a command's profile with `ActorContext` into a `Decision` |
| `domain/command/risk_catalog_test.go` | Tier ordering, profile lookup, default fallback, TTL ordering |
| `domain/command/risk_evaluator_test.go` | Allow/require-confirmation/allow-when-confirmed/MFA-gating/zero-value |
| `api/handlers/command/command_execute_test.go` | Handler-level risk gate + audit emission (fake audit logger) |
| `internal/audit/risk_audit_test.go` | `Entry` risk fields, `ActionCommandExecuted`, `NoOpLogger` interface conformance |

### Files Modified

| File | Change |
|------|--------|
| `internal/audit/audit_repository.go` | Extended `Entry` with `TraceID` + `RiskTier`; added `ActionCommandExecuted`; INSERT now persists both columns |
| `internal/audit/audit_separate_store.go` | Separate-DB schema + INSERT extended with `trace_id`/`risk_tier`; tightened dir perms to 0750 |
| `internal/audit/audit_logger.go` | Added `Logger.CommandExecuted` + `CommandExecutedEvent`; `NoOpLogger` now satisfies the command-audit interface; added `NewNoOpLogger()` |
| `internal/infrastructure/storage/sqlite.go` | Idempotent migration #57 adds `trace_id`/`risk_tier` columns to `audit_logs` |
| `api/handlers/command/command_execute.go` | Injects `RiskEvaluator` + `AuditLogger`; `Handle` now gates on risk before dispatch and emits an audit event on every attempt; added `confirm` request flag |
| `api/wire/wire_handlers.go`, `wire.go`, `providers.go` | Thread `RiskEvaluator` + audit fallback through the wiring |
| `api/api_server.go` | Construct `ExecuteHandler` with the new dependencies |

### How it is wired (no dead code)

- `RiskEvaluator` is constructed in the wire layer and passed into `ExecuteHandler`.
- `ExecuteHandler.Handle` calls `authorizeCommand` **after** validation/org checks and **before** dispatch, so a bad request never reaches the evaluator and a dangerous command never reaches the device.
- `DecisionRequireConfirmation` в†’ HTTP 425 (Too Early) with `risk_tier` + `trace_id`, and a `blocked` audit entry with `reason: "confirmation required"`.
- `DecisionDeny` в†’ HTTP 403 + `blocked` audit entry.
- On success в†’ HTTP 202 + a `success` audit entry carrying `trace_id`, `risk_tier`, `dispatch_id`.
- The audit `Entry` carries `TraceID` (joins to request logs) and `RiskTier` (classifies the event), persisted by both the co-located and separate-DB repositories.

### Design note: why no blanket request-level audit middleware

The spec listed an "audit middleware" item. A middleware that fires on every
HTTP request would duplicate the access log and drown the security signal in
noise. Instead the audit trail is emitted at the one place that knows the
command name and its risk tier вЂ” the command execution handler вЂ” which is
exactly the risky-operation path the spec targets. This keeps every Phase 2
artifact used and avoids the dead-code pattern that bit Phase 1.

### Forward hooks for Phase 3

- `ActorContext.MFAVerified` is plumbed through the evaluator but not yet read
  from the gin context (no MFA-context middleware yet). Critical-tier commands
  are therefore gated behind MFA via the evaluator contract; Phase 3 will wire
  the live MFA-verified flag.
- `CommandRiskProfile.ConfirmationTTL` is defined for the future confirmation
  store; today a confirmation is a per-request `confirm` flag. Phase 3 will
  replace this with a short-lived confirmation token.

---

## Appendix F: Phase 3 Implementation Summary

### Goal

A single-use, short-lived confirmation token gates risky device commands. An
operator requests a token for a specific (operator, command, device) triple,
then presents it on command execution; the token is consumed exactly once,
scoped, and bounded by the command's risk TTL. The live MFA-verified flag is
sourced from the authenticated session, completing the critical-tier gate.

### Files Created

| File | Purpose |
|------|---------|
| `domain/confirmation/pending_confirmation.go` | `PendingConfirmation` model + `Repository` interface (Create/Get/Consume/DeleteExpired) |
| `domain/confirmation/pending_confirmation_test.go` | Expiry, consumption, matching (incl. unscoped device) |
| `application/confirmation/service.go` | `Service.RequestConfirmation` (issues TTL'd tokens) + `ConsumeForCommand` (validates ownership/match/expiry/single-use) |
| `application/confirmation/service_test.go` | Issue/consume/mismatch/unknown/expired/single-use (in-memory fake repo) |
| `infrastructure/storage/confirmation_store.go` | SQLite `ConfirmationRepository` (atomic `Consume` via UPDATEвЂ¦WHERE consumed_at IS NULL AND expires_at > ?) |
| `infrastructure/storage/confirmation_store_test.go` | Create/Get/Consume success/already-consumed/expired/DeleteExpired against a migrated temp DB |
| `infrastructure/storage/058_command_confirmations.go` | Migration #58: `command_confirmations` table + indexes |
| `api/handlers/confirmation/confirmation_handler.go` | `RequestConfirmation` endpoint (issues tokens, returns `confirmation_required:false` for low-risk) |
| `api/handlers/confirmation/confirmation_handler_test.go` | Issues token for risky, not-required for low-risk, rejects missing command/operator |

### Files Modified

| File | Change |
|------|--------|
| `api/handlers/command/command_execute.go` | Replaced the per-request `confirm` flag with a `confirmation_token`; `authorizeCommand` now sources `MFAVerified` from the session and consumes the token via a `ConfirmationConsumer` interface; critical-tier enforces MFA before token acceptance |
| `internal/infrastructure/storage/sqlite.go` | Registered migration #58 |
| `api/wire/wire_handlers.go`, `wire.go`, `providers.go` | Thread `ConfirmationService` through the wiring; the confirmation handler doubles as the command handler's confirmation consumer |
| `api/api_server.go` | `ServerConfig.ConfirmationService` + `Server.confirmationHandler`; construct the confirmation handler and pass it to `NewExecuteHandler` |
| `api/server_routes.go` | Register `POST /v1/devices/:imei/command/confirm` under the operator-authenticated, org-scoped devices group |

### How it is wired (no dead code)

- The confirmation `Service` is the only writer to the store; `RequestConfirmation` issues tokens, `ConsumeForCommand` consumes them.
- `ExecuteHandler.authorizeCommand` в†’ on `DecisionRequireConfirmation`, calls `consumeConfirmation`: for critical-tier it first requires an MFA-verified session, then requires a non-empty token, then calls the `ConfirmationConsumer` (the confirmation handler) which atomically consumes the token. On any failure it returns 425 with a specific message and a `blocked` audit entry.
- The confirmation `Handler` implements `ConfirmationConsumer`, so the command handler depends on an interface, not the confirmation service internals вЂ” and the handler is constructed once and reused for both the HTTP endpoint and the consumer role.
- When no confirmation service is configured, `confirmations` is nil and risky commands are blocked with 425 ("Confirmations are not enabled") вЂ” safe by default.

### MFA wiring (completes the Phase 2 hook)

- `authorizeCommand` now reads `middleware.GetSession(c).MFAVerifiedAt != nil` to set `ActorContext.MFAVerified`. Critical-tier commands (`device.factory_reset`) require this AND a valid confirmation token; high-tier (`device.reboot`) require only the token.

### Forward hook for Phase 4

Validation errors remain to be migrated to the new `domain/errors/validation.go` helpers; the structured error response plumbing from Phase 1 is ready to consume them.

---

## Appendix G: Phase 4 Implementation Summary

### Goal

Field-level validation errors flow through the structured error system: every
failed validation returns a `VALIDATION_FAILED` response carrying per-field
`details`, a `trace_id`, and a `docs_url` — replacing the legacy
`{error:"bad_request", message, errors}` envelope. The Phase 1 validation
helpers (`ValidationDetail`, `NewValidationDetail`, `ErrValidationErrors`) that
were previously dead code are now consumed by production paths.

### Files Created

| File | Purpose |
|------|---------|
| `domain/errors/validation.go` | `ValidationError` type + `NewValidationError` + `ValidationDetailsOf` (errors.As extraction) |
| `domain/errors/validation_test.go` | Error formatting, construction, direct + wrapped extraction, non-validation rejection |
| `api/middleware/error_test.go` | Error middleware renders structured 400 with details + trace_id for a recorded `ValidationError` |

### Files Modified

| File | Change |
|------|--------|
| `api/middleware/validation.go` | `ValidationMiddleware` + `ValidationMiddlewareFunc` now convert `ValidationErrors` → `[]domainerrors.ValidationDetail` and call `responses.RespondValidationError` (structured envelope + `c.Abort()`) instead of the legacy `gin.H` body |
| `api/middleware/error.go` | `ErrorHandler` detects a recorded `*errors.ValidationError` via `ValidationDetailsOf` and renders a structured 400 (`VALIDATION_FAILED` + details + trace_id + docs_url) before the generic status→code path |
| `api/handlers/command/command_execute.go` | `validateCommandRequest` returns a `*domainerrors.ValidationError` with field-level details (removing the dead `command == ""` branch); `Handle` records it via `c.Error` so the error middleware renders the structured 400 |

### How it is wired (no dead code)

- Every `ValidationMiddleware`/`ValidationMiddlewareFunc` user (login, register, command execute, FCM token, settings, etc.) now emits the structured validation envelope automatically — no per-handler changes required.
- `ExecuteHandler.Handle` validates → `c.Error(*ValidationError)` → `ErrorHandler` extracts details via `ValidationDetailsOf` → structured 400 with `details`, `trace_id`, `docs_url`. A single migration point proves the handler-side path; other handlers can adopt `c.Error(NewValidationError(...))` incrementally without touching the response shape.
- The Phase 1 `responses.RespondValidationError` and `domainerrors.NewValidationDetail` helpers — previously unused — are now the production rendering path, retiring their dead-code status.

### Verification

- `go build ./...`, `go vet ./...`, `go test ./...` all pass.
- `golangci-lint v2.12.2` (project config): 0 issues in `domain/errors`, `handlers/command`, and all Phase 4 additions; pre-existing `ifElseChain` nits in legacy schema validators and pre-existing `noctx` in untouched test files remain (out of scope).