# Storage Migration & Modularization PRD

**Version:** 1.0  
**Date:** 2026-06-15  
**Status:** Draft  
**Author:** Vyzorix Development Team

---

## 1. Overview

### 1.1 Purpose

This document specifies the requirements for three interconnected architectural changes to the Vyzorix Update Server:

1. **SQLite Storage Modularization** - Split monolithic `sqlite.go` into focused domain modules
2. **UUIDv7 Migration** - Migrate telemetry IDs from auto-increment integers to time-ordered UUIDs
3. **Argon2id Migration** - Replace bcrypt with Argon2id for all password/secret hashing

### 1.2 Background

The current `sqlite.go` file is a 1,618-line monolith handling all database operations. As the project matures, this creates:
- Maintainability issues (hard to navigate, test, and review)
- No clear domain boundaries
- Difficult to onboard new contributors

Additionally, the current ID strategy and hashing algorithm need modernization for better security and scalability.

### 1.3 Scope

| Item | Scope |
|:-----|:------|
| Storage modularization | `pkg/storage/` package |
| UUID migration | `telemetry` table only |
| Hash migration | Passwords and command secrets |

### 1.4 Out of Scope

- PostgreSQL or other database migrations
- API version changes
- Frontend changes
- FCM notification migration

---

## 2. Technical Requirements

### 2.1 SQLite Storage Modularization

#### 2.1.1 Target Structure

```
pkg/storage/
 store.go           # Store struct, Open/Close, DB connection
 migrations.go      # Migration registry and runner
 devices.go         # Device CRUD operations
 telemetry.go       # Telemetry storage (UUIDv7)
 commands.go        # Command dispatch & status
 operators.go       # Operator CRUD, sessions, verifications
 settings.go        # System settings
 crypto.go          # Argon2id utilities (NEW)
 uuid.go            # UUIDv7 generation utilities (NEW)
```

#### 2.1.2 Module Specifications

| Module | Responsibility | Public API |
|:-------|:---------------|:-----------|
| `store.go` | Connection management, transaction support | `Open()`, `Close()`, `Ping()`, `BeginTx()` |
| `migrations.go` | Schema versioning, migration runner | `migrate()`, `GetVersion()`, `SetVersion()` |
| `devices.go` | Device registration, status, FCM token | `Register()`, `Device()`, `SetOnline()`, `Touch()`, `DeleteDevice()` |
| `telemetry.go` | Telemetry frame storage | `SaveTelemetry()`, `GetTelemetry()` |
| `commands.go` | Command dispatch, delivery tracking | `SaveCommand()`, `GetCommandStatus()`, `MarkDelivered()` |
| `operators.go` | Operator CRUD, sessions, verifications | `CreateOperator()`, `GetOperator()`, `CreateSession()`, etc. |
| `settings.go` | Key-value settings | `GetSetting()`, `SetSetting()`, `GetEnforceHMAC()` |
| `crypto.go` | Argon2id hashing utilities | `HashPassword()`, `VerifyPassword()`, `HashSecret()`, `VerifySecret()` |
| `uuid.go` | UUIDv7 generation | `NewUUIDv7()`, `ParseUUIDv7()` |

#### 2.1.3 Constraints

- All modules must be in the `storage` package
- No circular dependencies between modules
- Each module's public API must have godoc comments
- All public functions must have tests

### 2.2 UUIDv7 Migration

#### 2.2.1 Rationale

| Auto-increment Integer | UUIDv7 |
|:-----------------------|:-------|
| Predictable sequence | Time-ordered, unguessable |
| No context information | Embeds timestamp |
| Integer overflow risk | Unlimited scale |
| Cross-database issues | Database agnostic |

#### 2.2.2 Specification

**UUIDv7 Format:**
```
xxxxxxx-xxxx-7xxx-yxxx-xxxxxxxxxxxx
  timestamp   version  variant+random
```

**Example:**
```
01920a14-7f3c-7abc-b123-4567890abcdef
```

**Requirements:**
- Must be time-sortable (Unix timestamp in most-significant bits)
- Must be generated server-side (not from client)
- Must use cryptographically secure random for remaining bits
- Must be lowercase hex format

#### 2.2.3 Affected Table

| Table | Column | Change |
|:-------|:-------|:-------|
| `telemetry` | `id` | `INTEGER PRIMARY KEY AUTOINCREMENT` → `TEXT PRIMARY KEY` |

#### 2.2.4 Migration Strategy

Since this is a development project with no production data:

1. Drop existing `telemetry` table
2. Create new `telemetry` table with `TEXT PRIMARY KEY`
3. Update `SaveTelemetry()` to generate UUIDv7
4. Update telemetry queries to use string IDs

**No backward compatibility required.**

### 2.3 Argon2id Migration

#### 2.3.1 Rationale

| bcrypt | Argon2id |
|:-------|:--------|
| Memory-hard but not optimal | OWASP-recommended since 2023 |
| Fixed memory cost (4-6 MB) | Configurable memory (64+ MB) |
| Not side-channel resistant | Side-channel resistant |
| GPU-friendly | GPU/ASIC resistant |

#### 2.3.2 OWASP 2023 Parameters

```go
var Argon2Params = argon2.Params{
    Memory:      64 * 1024, // 64 MB
    Iterations:  3,
    Parallelism: 4,
    SaltLen:     32,
    KeyLen:      32,
    HashLen:     32,
}
```

#### 2.3.3 Affected Code Locations

| File | Function | Change |
|:-----|:---------|:-------|
| `handlers/auth_core.go` | `Register()`, `Login()` | bcrypt → argon2id |
| `handlers/auth_password_reset.go` | `ResetPassword()` | bcrypt → argon2id |
| `internal/command_signer.go` | `HashSecret()`, `VerifySecretHash()` | bcrypt → argon2id |
| `pkg/storage/secret_hash.go` | `HashSecret()`, `VerifyHash()` | bcrypt → argon2id |

#### 2.3.4 Implementation Requirements

- Use `github.com/tvtháng/kim/argon2` or equivalent pure-Go library
- All new hashes must start with `$argon2id$`
- No backward compatibility with bcrypt (development phase, no existing users)
- All hash operations must use constant-time comparison

---

## 3. Migration Phases

### Phase 1: Storage Modularization

| Step | Task | Files |
|:-----|:-----|:------|
| 1.1 | Create `uuid.go` with UUIDv7 generation | `pkg/storage/uuid.go` |
| 1.2 | Create `crypto.go` with Argon2id (empty, placeholder) | `pkg/storage/crypto.go` |
| 1.3 | Create `migrations.go` with migration registry | `pkg/storage/migrations.go` |
| 1.4 | Extract device operations to `devices.go` | `pkg/storage/devices.go` |
| 1.5 | Extract telemetry operations to `telemetry.go` | `pkg/storage/telemetry.go` |
| 1.6 | Extract command operations to `commands.go` | `pkg/storage/commands.go` |
| 1.7 | Extract operator operations to `operators.go` | `pkg/storage/operators.go` |
| 1.8 | Extract settings operations to `settings.go` | `pkg/storage/settings.go` |
| 1.9 | Simplify `sqlite.go` to connection + orchestration | `pkg/storage/sqlite.go` |
| 1.10 | Run tests and linters | All files |

### Phase 2: UUIDv7 Migration

| Step | Task | Files |
|:-----|:-----|:------|
| 2.1 | Update `migrations.go` to handle telemetry table recreation | `pkg/storage/migrations.go` |
| 2.2 | Update `telemetry.go` to use UUIDv7 | `pkg/storage/telemetry.go` |
| 2.3 | Update all telemetry queries to use string IDs | `pkg/storage/telemetry.go` |
| 2.4 | Run tests and linters | All files |

### Phase 3: Argon2id Migration

| Step | Task | Files |
|:-----|:-----|:------|
| 3.1 | Implement `crypto.go` with Argon2id | `pkg/storage/crypto.go` |
| 3.2 | Update `auth_core.go` to use `crypto.go` | `handlers/auth_core.go` |
| 3.3 | Update `auth_password_reset.go` to use `crypto.go` | `handlers/auth_password_reset.go` |
| 3.4 | Update `command_signer.go` to use Argon2id | `internal/command_signer.go` |
| 3.5 | Update `secret_hash.go` to use Argon2id | `pkg/storage/secret_hash.go` |
| 3.6 | Update test files | All `*_test.go` files |
| 3.7 | Run tests and linters | All files |

---

## 4. Testing Requirements

### 4.1 Unit Tests

Each module must have tests covering:
- Happy path operations
- Error conditions
- Edge cases (empty inputs, boundary values)

### 4.2 Integration Tests

- Full migration sequence must work end-to-end
- All CRUD operations must work after migration
- WebSocket connections must survive migration

### 4.3 Test Coverage

| Metric | Target |
|:-------|:-------|
| Line coverage | ≥ 80% |
| Function coverage | 100% public API |

---

## 5. Verification Checklist

Before considering this migration complete:

- [ ] All storage modules compile without errors
- [ ] All storage modules pass their unit tests
- [ ] `golangci-lint` reports 0 issues
- [ ] `go test ./...` passes all tests
- [ ] Telemetry IDs are valid UUIDv7 format
- [ ] New password hashes start with `$argon2id$`
- [ ] Device registration still works end-to-end
- [ ] WebSocket telemetry streaming still works
- [ ] Command dispatch and status still works
- [ ] Operator auth flow still works (login, register, logout)

---

## 6. Dependencies

### 6.1 New Go Dependencies

| Package | Purpose | Version |
|:--------|:--------|:--------|
| `github.com/google/uuid` | UUID utilities | v1.6.0 |
| `github.com/tvtháng/kim/argon2` | Argon2id hashing | latest |

### 6.2 Existing Dependencies

No changes to existing dependency versions required.

---

## 7. Rollback Plan

Since this is development-phase with no production data:

1. **If issues found during implementation**: Fix and continue
2. **If tests fail after migration**: Revert changes from version control
3. **If production deployment needed before completion**: Do not merge until complete

---

## 8. Timeline

| Phase | Estimated Time |
|:------|:---------------|
| Phase 1: Storage Modularization | 4-6 hours |
| Phase 2: UUIDv7 Migration | 1-2 hours |
| Phase 3: Argon2id Migration | 2-3 hours |
| Testing & Verification | 2 hours |
| **Total** | **9-13 hours** |

---

## 9. Document History

| Version | Date | Changes |
|:--------|:-----|:--------|
| 1.0 | 2026-06-15 | Initial draft |