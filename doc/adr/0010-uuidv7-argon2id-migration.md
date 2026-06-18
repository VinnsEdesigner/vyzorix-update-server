# ADR-0010: UUIDv7 for Telemetry and Argon2id for Password Hashing

**Status:** Accepted  
**Date:** 2026-06-15

## Context

The Vyzorix Update Server's storage layer has two technical debts that need addressing:

1. **Telemetry IDs use auto-increment integers** - Not scalable across distributed systems, predictable, no time ordering guarantees
2. **Password hashing uses bcrypt** - While still considered secure, Argon2id is the OWASP-recommended successor since 2023

Since this is a development-phase project with no production users, we have the opportunity to make breaking changes without backward compatibility requirements.

## Decision

We will:

1. Migrate telemetry IDs from `INTEGER PRIMARY KEY AUTOINCREMENT` to `TEXT PRIMARY KEY` using UUIDv7
2. Replace bcrypt with Argon2id for all password and secret hashing

## Technical Specifications

### UUIDv7

UUIDv7 embeds a Unix timestamp in the most-significant 48 bits, providing:
- Time-ordered IDs (good for database indexing)
- Globally unique
- Unpredictable (unlike sequential integers)
- Database agnostic (no conflicts across distributed systems)

Format: `xxxxxxxx-xxxx-7xxx-yxxx-xxxxxxxxxxxx` (version 7, variant 1)

### Argon2id Parameters (OWASP 2023)

```go
var Argon2Params = argon2.Params{
    Memory:      64 * 1024, // 64 MB
    Iterations:  3,
    Parallelism: 4,
    SaltLen:     32 bytes,
    KeyLen:      32 bytes,
}
```

**Why these values:**
- Memory: 64 MB provides strong GPU/ASIC resistance
- Iterations: 3 is the OWASP minimum
- Parallelism: 4 matches typical CPU core counts
- Salt: 32 bytes (256 bits) is industry standard

## Alternatives Considered

### UUIDv4 (Random UUID)

- Pros: Simple, widely supported
- Cons: Not time-ordered, poor database locality for inserts, reveals no information
- **Decision: Rejected** - Time-ordering is important for telemetry queries

### ULID (Universally Unique Lexicographically Sortable Identifier)

- Pros: Time-ordered, lexicographically sortable
- Cons: Not standard UUID format, Crockford base32 encoding
- **Decision: Rejected** - Standard UUIDv7 is more universally recognized

### bcrypt

- Pros: Well-tested, widely supported, backward compatible
- Cons: Lower memory cost, not OWASP-recommended for new systems
- **Decision: Rejected** - Argon2id is the modern standard

### scrypt

- Pros: Memory-hard, well-tested
- Cons: Higher latency, not side-channel resistant
- **Decision: Rejected** - Argon2id has better side-channel resistance and OWASP recommendation

## Implementation Plan

### Phase 1: Storage Modularization

Split `sqlite.go` (1,618 lines) into domain-specific modules:

```
pkg/storage/
 store.go       # Store struct, connection
 migrations.go  # Migration registry
 devices.go     # Device CRUD
 telemetry.go    # Telemetry storage (UUIDv7)
 commands.go    # Command dispatch
 operators.go   # Operator CRUD
 settings.go    # System settings
 crypto.go      # Argon2id utilities
 uuid.go        # UUIDv7 utilities
```

### Phase 2: UUIDv7 Migration

1. Create new `telemetry` table with `TEXT PRIMARY KEY`
2. Update `SaveTelemetry()` to generate UUIDv7
3. Update all telemetry queries

### Phase 3: Argon2id Migration

1. Implement `crypto.go` with Argon2id
2. Replace bcrypt in:
   - `handlers/auth_core.go` (operator passwords)
   - `handlers/auth_password_reset.go` (password reset)
   - `internal/command_signer.go` (command secrets)
   - `pkg/storage/secret_hash.go` (device secrets)

## Consequences

### Positive

- Time-ordered telemetry IDs improve query performance
- UUIDs are database-agnostic for future scalability
- Argon2id provides stronger protection against GPU/ASIC attacks
- OWASP compliance for password hashing
- Modular storage code is easier to maintain and test

### Negative

- Slightly longer IDs (36 chars vs 8 bytes)
- Argon2id is slower than bcrypt (by design - more secure)
- Breaking change for any external systems relying on integer telemetry IDs

### Risks

| Risk | Mitigation |
|:-----|:----------|
| Argon2id CPU cost on login | Pre-compute where possible, acceptable latency |
| Longer IDs in database | Minimal storage impact (indexed column) |
| Migration complexity | No backward compat needed (dev phase only) |

## Review History

| Date | Reviewer | Notes |
|:-----|:---------|:------|
| 2026-06-15 | Vyzorix Team | Initial acceptance |