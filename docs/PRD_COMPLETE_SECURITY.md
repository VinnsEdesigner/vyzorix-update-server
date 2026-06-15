# PRD: Vyzorix Update Server - Complete Security Implementation

> **Feature Name:** End-to-End Security Hardening  
> **Version:** 1.0  
> **Status:** Draft  
> **Created:** 2026-06-15  
> **Target Release:** Production MVP  

---

## 1. Introduction

### Problem Statement

The Vyzorix Update Server currently lacks comprehensive security protections required for production deployment. While some features exist (rate limiting, Argon2id, UUIDv7), critical security layers are missing, leaving the system vulnerable to:

- **Unauthorized API access** - No request signing or origin verification
- **Cross-site request forgery** - No CSRF protection on mutative endpoints
- **Bot automation** - No Turnstile verification on auth flows
- **Session hijacking** - No session revocation mechanism
- **User enumeration** - Timing attacks can reveal existing users
- **Replay attacks** - Signed requests can be recorded and replayed
- **Information leakage** - No comprehensive error sanitization
- **Audit gaps** - No logging of security events

### Solution

Implement a **comprehensive, layered security architecture** following enterprise best practices. This PRD covers all security features needed for production deployment, organized into logical phases.

---

## 2. Goals

### Primary Goals

- **G1:** Prevent unauthorized API access via cryptographic request signing
- **G2:** Block automated bot attacks using Cloudflare Turnstile
- **G3:** Prevent CSRF attacks on all mutative endpoints
- **G4:** Enable session revocation to handle compromised sessions
- **G5:** Implement user enumeration prevention with constant-time responses
- **G6:** Add comprehensive audit logging for all security events
- **G7:** Enforce MFA/TOTP for sensitive operations
- **G8:** Harden container deployment with security best practices

### Secondary Goals

- **G9:** Add structured logging with PII redaction
- **G10:** Implement health and metrics endpoints
- **G11:** Add configuration validation with fail-fast startup
- **G12:** Create comprehensive test coverage for all security features
- **G13:** Document all security mechanisms and configurations

---

## 3. Implementation Phases

### Phase 1: Core Security (P0 - Must Have)

| Feature | Files | Est. Time | Priority |
|---------|-------|-----------|----------|
| Request Signing & Encryption | 15 | 18h | P0 |
| Security Headers Middleware | 2 | 1h | P0 |
| Cloudflare Turnstile | 3 | 2h | P0 |
| Panic Recovery + Error Handler | 2 | 1h | P0 |
| Config Validator | 2 | 1h | P0 |
| **Phase 1 Total** | **24** | **23h** | |

### Phase 2: Authentication & Authorization (P1 - Should Have)

| Feature | Files | Est. Time | Priority |
|---------|-------|-----------|----------|
| CSRF Protection | 3 | 2h | P1 |
| Session Revocation | 4 | 3h | P1 |
| User Enumeration Prevention | 2 | 1h | P1 |
| Account Lockout | 4 | 3h | P1 |
| **Phase 2 Total** | **13** | **9h** | |

### Phase 3: Monitoring & Compliance (P2 - Good to Have)

| Feature | Files | Est. Time | Priority |
|---------|-------|-----------|----------|
| Audit Logging | 4 | 4h | P2 |
| MFA/TOTP | 6 | 8h | P2 |
| Health & Metrics | 4 | 3h | P2 |
| Structured Logging | 4 | 2h | P2 |
| **Phase 3 Total** | **18** | **17h** | |

### Phase 4: Infrastructure & Testing (P3 - Future)

| Feature | Files | Est. Time | Priority |
|---------|-------|-----------|----------|
| Container Hardening | 3 | 2h | P3 |
| CI/CD Security | 3 | 3h | P3 |
| Documentation | 4 | 4h | P3 |
| Testing (all new code) | ~25 | 8h | P3 |
| **Phase 4 Total** | **35** | **17h** | |

### Grand Total

| Category | Files | Est. Time |
|----------|-------|-----------|
| **All Phases** | **86** | **66h** |

---

## 4. Detailed File Specifications

### 4.1 Request Signing & Encryption (15 files)

**Purpose:** Cryptographic proof of request origin + body encryption

```
apps/api/internal/api/middleware/
├── request_signing.go           # Core signature verification middleware
├── request_signing_test.go     # Unit tests
├── replay_protection.go        # Replay attack prevention cache
├── replay_protection_test.go  # Unit tests
├── response_encryption.go      # Optional response encryption
└── signed_handlers.go           # Route registration

apps/api/internal/api/handlers/
├── admin_clients.go            # Admin CRUD for clients
├── admin_clients_test.go
├── client_credentials.go       # Get credentials endpoint
└── client_credentials_test.go

apps/api/internal/auth/
├── request_signer.go           # Client-side signing reference
├── request_signer_test.go
├── client_registry.go          # Client management logic
└── client_registry_test.go

apps/api/pkg/config/
└── signing.go                  # Signing configuration

apps/api/pkg/storage/
├── clients.go                  # Client storage operations
├── clients_test.go
├── signing_keys.go             # Key rotation storage
└── signing_keys_test.go
```

**Each file must:**
- `request_signing.go`: Verify HMAC-SHA256 signatures, validate timestamp (±5min), check replay cache, decrypt body
- `replay_protection.go`: Thread-safe in-memory cache with TTL eviction
- `response_encryption.go`: Encrypt responses using AES-256-GCM when requested
- `admin_clients.go`: CRUD endpoints for client management
- `client_credentials.go`: Issue client credentials to authenticated users

---

### 4.2 Security Headers Middleware (2 files)

**Purpose:** Add security headers to all responses

```
apps/api/internal/api/middleware/
├── security_headers.go        # Main middleware
└── security_headers_test.go
```

**Each file must:**
- `security_headers.go`: Set CSP, HSTS, X-Frame-Options, X-Content-Type-Options, Referrer-Policy, Permissions-Policy, COOP, COEP headers

---

### 4.3 Cloudflare Turnstile (3 files)

**Purpose:** Bot detection for auth flows

```
apps/api/internal/api/middleware/
├── turnstile.go               # Verification middleware
└── turnstile_test.go

apps/api/pkg/config/
└── turnstile.go               # Config loading
```

**Each file must:**
- `turnstile.go`: Verify Turnstile token with Cloudflare API, cache results, handle failures
- `turnstile.go`: Apply to signup, login, password reset, sensitive actions

---

### 4.4 Panic Recovery & Error Handler (2 files)

**Purpose:** Prevent information leakage and crashes

```
apps/api/internal/api/middleware/
├── recovery.go                # Panic recovery
└── error.go                   # Uniform error responses
```

**Each file must:**
- `recovery.go`: Catch panics, log securely, return sanitized 500 response
- `error.go`: Standardize error codes, prevent stack traces in production

---

### 4.5 Config Validator (2 files)

**Purpose:** Validate environment variables at startup

```
apps/api/pkg/config/
├── validator.go               # Validation logic
└── validator_test.go
```

**Each file must:**
- `validator.go`: Validate all required env vars, fail fast with clear errors

---

### 4.6 CSRF Protection (3 files)

**Purpose:** Prevent cross-site request forgery

```
apps/api/internal/api/middleware/
├── csrf.go                    # Token generation/validation
└── csrf_test.go

apps/api/internal/api/handlers/
└── auth_csrf.go               # Get CSRF token endpoint
```

**Each file must:**
- `csrf.go`: Generate tokens, validate via header, bind to session
- `auth_csrf.go`: Endpoint to get CSRF token for frontend

---

### 4.7 Session Revocation (4 files)

**Purpose:** Invalidate compromised sessions

```
apps/api/internal/auth/
├── revocation.go              # Revocation list management
└── revocation_test.go

apps/api/internal/api/middleware/
└── auth_revocation.go       # Check revocation

apps/api/pkg/storage/
└── sessions.go               # Session storage with revocation
```

**Each file must:**
- `revocation.go`: Add/remove tokens from revocation list
- `auth_revocation.go`: Middleware to check revocation status
- `sessions.go`: Store session hashes, check against revocation list

---

### 4.8 User Enumeration Prevention (2 files)

**Purpose:** Return identical responses for existing/non-existing users

```
apps/api/internal/api/middleware/
└── user_enum_block.go         # Fake hash timing

apps/api/internal/api/handlers/
└── auth_enum_safe.go          # Constant-time response handlers
```

**Each file must:**
- `user_enum_block.go`: Add fake hash computation for non-existing users
- `auth_enum_safe.go`: Return identical 201 response for all signup/login attempts

---

### 4.9 Account Lockout (4 files)

**Purpose:** Prevent brute force attacks

```
apps/api/internal/auth/
├── lockout.go                 # Lockout logic
└── lockout_test.go

apps/api/internal/api/middleware/
└── auth_lockout.go            # Check lockout on auth

apps/api/internal/api/handlers/
└── auth_lockout.go            # Lockout status endpoints
```

**Each file must:**
- `lockout.go`: Track failed attempts, lock after 5 attempts, exponential backoff
- `auth_lockout.go`: Middleware to check lockout status
- `auth_lockout.go`: Endpoints to check status, request unlock

---

### 4.10 Audit Logging (4 files)

**Purpose:** Track all security events

```
apps/api/internal/audit/
├── logger.go                  # Audit logger
├── logger_test.go
├── repository.go             # SQLite storage
└── repository_test.go
```

**Each file must:**
- `logger.go`: Log security events (login, logout, failed attempts, etc.)
- `repository.go`: Store audit logs in SQLite with indexes

---

### 4.11 MFA/TOTP (6 files)

**Purpose:** Second factor authentication

```
apps/api/internal/auth/
├── totp.go                    # TOTP generation/verification
├── totp_test.go
├── totp_qr.go                 # QR code generation
└── totp_qr_test.go

apps/api/internal/api/handlers/
├── auth_mfa.go                # MFA enrollment/verification endpoints
└── auth_mfa_test.go

apps/api/pkg/config/
└── totp.go                    # TOTP configuration
```

**Each file must:**
- `totp.go`: Generate/verify TOTP codes using RFC 6238
- `totp_qr.go`: Generate QR codes for mobile apps
- `auth_mfa.go`: Enrollment, verification, backup codes

---

### 4.12 Health & Metrics (4 files)

**Purpose:** Observability and monitoring

```
apps/api/internal/api/handlers/
├── health.go                  # Health check endpoints
└── health_test.go

apps/api/internal/metrics/
├── prometheus.go             # Prometheus metrics endpoint
└── middleware.go             # Request metrics
```

**Each file must:**
- `health.go`: `/health/live`, `/health/ready`, `/health/secure` endpoints
- `prometheus.go`: `/metrics` endpoint with security metrics

---

### 4.13 Structured Logging (4 files)

**Purpose:** Better logging with PII redaction

```
apps/api/internal/api/middleware/
└── log_context.go             # Request ID, user ID in logs

apps/api/pkg/logging/
├── structured.go             # Structured JSON logging
├── levels.go                 # Log level management
└── redactor.go               # PII redaction in logs
```

**Each file must:**
- `log_context.go`: Add request_id, user_id to log context
- `structured.go`: JSON logging with levels
- `redactor.go`: Redact PII from logs

---

### 4.14 Container Hardening (3 files)

**Purpose:** Secure container deployment

```
tooling/docker/
├── Dockerfile.prod            # Hardened multi-stage build
├── .dockerignore
└── security-check.sh          # Container security scan
```

**Each file must:**
- `Dockerfile.prod`: Minimal base, non-root user, read-only filesystem
- `security-check.sh`: Run Trivy scan, check for vulnerabilities

---

### 4.15 CI/CD Security (3 files)

**Purpose:** Security in CI/CD pipeline

```
.github/workflows/
├── security.yml               # Security checks in CI
├── test.yml                  # Test pipeline
└── lint.yml                  # Lint pipeline
```

**Each file must:**
- `security.yml`: Run govulncheck, npm audit, dependency review
- `test.yml`: Run all tests with coverage
- `lint.yml`: Run golangci-lint, ESLint

---

### 4.16 Documentation (4 files)

**Purpose:** Complete documentation

```
docs/
├── API.md                     # Complete API documentation
├── SECURITY.md                # Security posture document
├── DEPLOYMENT.md              # Deployment guide
└── ENVIRONMENT.md             # All environment variables
```

**Each file must:**
- `API.md`: All endpoints, request/response formats
- `SECURITY.md`: Security architecture, threat model
- `DEPLOYMENT.md`: Deployment instructions, scaling
- `ENVIRONMENT.md`: All env vars with descriptions

---

### 4.17 Testing (25 files)

**Purpose:** Comprehensive test coverage

```
**/ *_test.go for every new file
**/integration_tests/
├── api_test.go                # API integration tests
├── auth_test.go               # Auth flow tests
└── security_test.go          # Security-specific tests

**/fuzz_tests/
├── parser_fuzz.go             # Fuzzing tests for parsers
└── crypto_fuzz.go             # Fuzzing tests for crypto
```

**Each file must:**
- Unit tests for all functions
- Integration tests for API flows
- Fuzzing tests for critical parsers
- 80%+ code coverage

---

## 5. Database Schema Changes

### 5.1 api_clients
```sql
CREATE TABLE api_clients (
    id TEXT PRIMARY KEY,
    operator_id TEXT NOT NULL,
    name TEXT NOT NULL,
    platform TEXT NOT NULL,
    client_secret_hash TEXT NOT NULL,
    allowed_origins TEXT,
    allowed_paths TEXT,
    rate_limit INTEGER NOT NULL DEFAULT 100,
    is_active INTEGER NOT NULL DEFAULT 1,
    request_count INTEGER NOT NULL DEFAULT 0,
    last_request_at INTEGER,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (operator_id) REFERENCES operators(id) ON DELETE CASCADE
);

CREATE INDEX idx_api_clients_operator ON api_clients(operator_id);
CREATE INDEX idx_api_clients_active ON api_clients(is_active);
```

### 5.2 signing_keys
```sql
CREATE TABLE signing_keys (
    id TEXT PRIMARY KEY,
    client_id TEXT NOT NULL,
    key_hash TEXT NOT NULL,
    version INTEGER NOT NULL,
    issued_at INTEGER NOT NULL,
    expires_at INTEGER,
    is_active INTEGER NOT NULL DEFAULT 1,
    revoked_at INTEGER,
    FOREIGN KEY (client_id) REFERENCES api_clients(id) ON DELETE CASCADE
);

CREATE INDEX idx_signing_keys_client ON signing_keys(client_id);
CREATE INDEX idx_signing_keys_active ON signing_keys(client_id, is_active);
```

### 5.3 session_revocation_list
```sql
CREATE TABLE session_revocation_list (
    token_hash TEXT PRIMARY KEY,
    revoked_at INTEGER NOT NULL,
    reason TEXT
);

CREATE INDEX idx_revocation_token ON session_revocation_list(token_hash);
```

### 5.4 failed_login_attempts
```sql
CREATE TABLE failed_login_attempts (
    id TEXT PRIMARY KEY,
    operator_id TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    attempted_at INTEGER NOT NULL
);

CREATE INDEX idx_failed_attempts_operator ON failed_login_attempts(operator_id);
CREATE INDEX idx_failed_attempts_ip ON failed_login_attempts(ip_address);
```

### 5.5 account_lockouts
```sql
CREATE TABLE account_lockouts (
    operator_id TEXT PRIMARY KEY,
    locked_until INTEGER NOT NULL,
    reason TEXT,
    created_at INTEGER NOT NULL
);
```

### 5.6 audit_logs
```sql
CREATE TABLE audit_logs (
    id TEXT PRIMARY KEY,
    operator_id TEXT,
    action TEXT NOT NULL,
    resource_type TEXT,
    resource_id TEXT,
    ip_address TEXT,
    user_agent TEXT,
    metadata TEXT,
    result TEXT NOT NULL,
    created_at INTEGER NOT NULL
);

CREATE INDEX idx_audit_operator ON audit_logs(operator_id);
CREATE INDEX idx_audit_action ON audit_logs(action);
CREATE INDEX idx_audit_created ON audit_logs(created_at DESC);
```

### 5.7 mfa_secrets
```sql
ALTER TABLE operators ADD COLUMN mfa_secret TEXT;
ALTER TABLE operators ADD COLUMN mfa_enabled INTEGER NOT NULL DEFAULT 0;
ALTER TABLE operators ADD COLUMN mfa_backup_codes TEXT;
```

---

## 6. Configuration

### 6.1 Environment Variables

```bash
# Request Signing
REQUEST_SIGNING_ENABLED=true
SIGNING_TIMESTAMP_WINDOW=300
SIGNING_MAX_CACHE_SIZE=100000
SIGNING_GRACE_PERIOD=86400
ALLOW_UNSIGNED_FALLBACK=false

# Turnstile
TURNSTILE_SECRET=your-secret-key
TURNSTILE_ENABLED=true

# CSRF
CSRF_SECRET=your-secret-key
CSRF_ENABLED=true

# Session Revocation
SESSION_REVOCATION_ENABLED=true

# User Enumeration
USER_ENUM_PREVENTION_ENABLED=true

# Account Lockout
ACCOUNT_LOCKOUT_ENABLED=true
ACCOUNT_LOCKOUT_ATTEMPTS=5
ACCOUNT_LOCKOUT_DURATION=3600

# Audit Logging
AUDIT_LOGGING_ENABLED=true
AUDIT_LOG_RETENTION_DAYS=90

# MFA
MFA_ENABLED=true
MFA_ISSUER="Vyzorix"

# Health & Metrics
METRICS_ENABLED=true
METRICS_PATH="/metrics"

# Logging
LOG_LEVEL="info"
LOG_REDACT_PII=true

# Container
DOCKER_NON_ROOT_USER="appuser"
DOCKER_READONLY_ROOT=true
```

---

## 7. Testing Strategy

### 7.1 Unit Tests
- Cover all functions with table-driven tests
- Test edge cases (empty inputs, invalid formats, etc.)
- Mock database dependencies

### 7.2 Integration Tests
- Test complete API flows
- Test middleware chains
- Test database interactions

### 7.3 Security Tests
- Test signature verification
- Test replay attack prevention
- Test CSRF protection
- Test Turnstile verification

### 7.4 Performance Tests
- Benchmark signing/verification latency
- Test under load (1000+ requests/sec)
- Memory profiling

### 7.5 Fuzzing Tests
- Fuzz request parsers
- Fuzz crypto operations
- Fuzz signature verification

---

## 8. Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Unauthenticated requests blocked | 100% | All requests without valid signature rejected |
| Replay attacks detected | 100% | Replayed signatures within window rejected |
| Signing latency overhead | <5ms p99 | Performance benchmark |
| Client creation success rate | >99.9% | Admin client creation succeeds |
| Key rotation success rate | >99.9% | Rotation completes without downtime |
| Zero unauthorized access incidents | 0 per month | Security audit |
| CSRF attacks blocked | 100% | All CSRF attempts rejected |
| Session revocation effectiveness | 100% | Revoked sessions cannot access API |
| User enumeration prevention | 100% | Timing attacks cannot distinguish users |
| Account lockout effectiveness | 100% | Locked accounts cannot login |
| Audit log completeness | 100% | All security events logged |
| MFA adoption rate | >90% | Percentage of users with MFA enabled |
| Health check availability | >99.9% | Health endpoints respond within 100ms |
| Container security score | A+ | Trivy scan results |
| Test coverage | >80% | Code coverage percentage |

---

## 9. Open Questions

1. **Should we implement Ed25519 instead of HMAC-SHA256 for request signing?** Ed25519 is more modern but slightly more complex in JavaScript.

2. **Should mobile apps cache credentials indefinitely?** Or expire after 30 days requiring re-auth?

3. **Should we support API versioning for the signing scheme itself?** (e.g., `v1` in signature indicates algorithm version)

4. **Should we log all signing failures to a separate table?** For security monitoring.

5. **Should admins be able to set per-client rate limits?** Yes, included in FR-1.

6. **Should we implement IP-based rate limiting in addition to client-based?** Yes, layered approach.

7. **Should we add WebSocket connection signing?** Separate PRD needed.

8. **Should we implement certificate pinning for mobile apps?** Separate PRD needed.

---

## 10. Dependencies on Other PRDs

- **Request Signing** depends on: Session Management (for client credentials)
- **Turnstile** depends on: None
- **CSRF** depends on: Session Management
- **Session Revocation** depends on: Session Management
- **User Enumeration** depends on: None
- **Account Lockout** depends on: None
- **Audit Logging** depends on: None
- **MFA** depends on: Session Management
- **Health & Metrics** depends on: None
- **Structured Logging** depends on: None
- **Container Hardening** depends on: None
- **CI/CD Security** depends on: None
- **Documentation** depends on: All features

---

## 11. Out of Scope for This PRD

- OAuth/OIDC integration (separate PRD)
- WebSocket connection signing (separate PRD)
- Mobile certificate pinning (separate PRD)
- API usage billing/quotas (separate PRD)
- Third-party API integrations (separate PRD)
- GraphQL API signing (if added later)
- Advanced threat detection (AI-based anomaly detection)
- Hardware security modules (HSM) integration

---

## 12. Implementation Order Recommendation

```
Phase 1: Core Security (P0)
├── Config Validator
├── Panic Recovery + Error Handler
├── Security Headers Middleware
├── Request Signing & Encryption
└── Cloudflare Turnstile

Phase 2: Authentication & Authorization (P1)
├── CSRF Protection
├── Session Revocation
├── User Enumeration Prevention
└── Account Lockout

Phase 3: Monitoring & Compliance (P2)
├── Audit Logging
├── MFA/TOTP
├── Health & Metrics
└── Structured Logging

Phase 4: Infrastructure & Testing (P3)
├── Container Hardening
├── CI/CD Security
├── Documentation
└── Testing
```

---

*Document Version: 1.0*  
*Status: Ready for Review*  
*Next Steps: Review with team, then implementation*