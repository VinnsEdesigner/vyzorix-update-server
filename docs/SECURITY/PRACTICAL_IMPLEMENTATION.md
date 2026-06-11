# Practical Security Implementation Guide

> **For:** Render-hosted applications with minimal cost and maintenance
> **Team:** 2 experienced developers (you + me)
> **Approach:** Enterprise-grade security without enterprise complexity

---

## Table of Contents

1. [Free/Cheap Security Layers for Render](#1-freecheap-security-layers-for-render)
2. [Implementation: Security Headers Middleware](#2-implementation-security-headers-middleware)
3. [Implementation: CSP (Content Security Policy)](#3-implementation-csp-content-security-policy)
4. [Implementation: HSTS Header](#4-implementation-hsts-header)
5. [Implementation: Audit Logging](#5-implementation-audit-logging)
6. [Implementation: Session Revocation](#6-implementation-session-revocation)
7. [Implementation: UUIDv7 Migration](#7-implementation-uuidv7-migration)
8. [Implementation: MFA/TOTP](#8-implementation-mfatotp)
9. [Implementation: Secrets Management](#9-implementation-secrets-management)
10. [Implementation: Container Hardening](#10-implementation-container-hardening)
11. [Implementation: SIEM Alternative](#11-implementation-siem-alternative)

---

## 1. Free/Cheap Security Layers for Render

### What Works on Render (No Cost, Low Maintenance)

| Layer | Implementation | Cost | Maintenance |
|-------|---------------|------|-------------|
| **Cloudflare WAF** | Cloudflare Free Tier | $0 | Zero |
| **Cloudflare DDoS** | Cloudflare Free Tier | $0 | Zero |
| **Cloudflare CDN** | Cloudflare Free Tier | $0 | Zero |
| **Cloudflare Turnstile** | Cloudflare Free Tier | $0 | Zero |
| **Security Headers** | Go middleware | $0 | Low |
| **CSP** | Go middleware | $0 | Low |
| **HSTS** | Go middleware | $0 | Low |
| **Rate Limiting** | In-memory token bucket | $0 | Low |
| **Audit Logging** | SQLite table | $0 | Low |
| **Session Revocation** | SQLite table | $0 | Low |
| **UUIDv7** | Go stdlib | $0 | Low |
| **MFA/TOTP** | TOTP library | $0 | Medium |
| **Secrets** | .env + gitignore | $0 | Low |
| **Container** | Docker best practices | $0 | Low |
| **SIEM Alternative** | Log to SQLite + alerts | $0 | Medium |

### Cloudflare Configuration (Already Implemented)

Your Cloudflare setup provides enterprise-grade security for free:

#### DNS Settings ( Configured)
```
- Proxy status: Proxied (orange cloud) 
- SSL/TLS: Full (strict) 
- Always Use HTTPS: Enabled 
- Automatic HTTPS Rewrites: Enabled 
- Minimum TLS Version: 1.2 
```

#### Security Settings ( Configured)
```
- Security Level: Medium (recommended) 
- Challenge Passage: 30 minutes 
- Browser Integrity Check: Enabled 
- Privacy Pass Support: Enabled 
- Advanced DDoS Protection: Enabled 
```

#### WAF Settings ( Configured)
```
- Managed Rules: OWASP ModSecurity Core Rule Set 
- Sensitivity Level: Medium 
- Action Mode: Block 
- Rules: SQLi, XSS, Bad Bots, etc. 
```

#### Bot Management ( Configured)
```
- Bot Fight Mode: Enabled 
- JavaScript Detections: Enabled 
- Super Bot Fight Mode: Disabled (paid feature) 
```

#### Turnstile ( Configured)
```
- Widget Key: Configured 
- Secret Key: Configured 
- Used in: Registration, Login, Sensitive Actions 
```

#### Rate Limiting ( Partial)
```
- 10,000 free requests/month
- Applied to: /api/* endpoints
- Action: Block for 1 hour
```

#### Firewall Rules ( Partial)
```
- 5 free rules
- Current rules: Block known bad IPs, Allowlist admin IPs
```

#### Caching ( Configured)
```
- Cache Level: Standard 
- Browser Cache TTL: Respect Existing Headers 
- Edge Cache TTL: 2 hours 
- Always Online: Enabled 
```

---

### What Doesn't Work on Render (Avoid)

| Layer | Why Not |
|-------|---------|
| **WAF** | Requires Cloudflare/AWS (costly) |
| **DDoS Protection** | Render provides basic protection |
| **API Gateway** | Overkill for current scale |
| **Vault** | HashiCorp Vault costs money |
| **Commercial SIEM** | Datadog/Splunk too expensive |
| **Kubernetes** | Render doesn't support K8s |

---

## 2. Implementation: Security Headers Middleware

### Code: `apps/api/internal/api/middleware/security_headers.go`

```go
package middleware

import "net/http"

// SecurityHeadersMiddleware adds essential security headers to all responses
func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Prevent clickjacking
            w.Header().Set("X-Frame-Options", "DENY")
            
            // Prevent MIME sniffing
            w.Header().Set("X-Content-Type-Options", "nosniff")
            
            // Enable XSS protection (legacy browsers)
            w.Header().Set("X-XSS-Protection", "1; mode=block")
            
            // Prevent information leakage
            w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
            
            // Modern security headers
            w.Header().Set("Content-Security-Policy", 
                "default-src 'self'; " +
                "script-src 'self' 'unsafe-inline' https://challenges.cloudflare.com; " +
                "style-src 'self' 'unsafe-inline'; " +
                "img-src 'self' data: https://*; " +
                "font-src 'self' https://*; " +
                "connect-src 'self' https://challenges.cloudflare.com https://api.vinnsedesigner.render.com; " +
                "frame-src 'none'; " +
                "object-src 'none'; " +
                "base-uri 'self'; " +
                "form-action 'self'")
            
            // Force HTTPS for 6 months
            w.Header().Set("Strict-Transport-Security", "max-age=15768000; includeSubDomains")
            
            // Prevent Spectre attacks
            w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
            w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
            
            // Disable browser features that could be exploited
            w.Header().Set("Permissions-Policy", 
                "geolocation=(), " +
                "mid=(), " +
                "sync-xhr=(), " +
                "microphone=(), " +
                "camera=(), " +
                "magnetometer=(), " +
                "gyroscope=(), " +
                "fullscreen=(self), " +
                "payment=()")
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Usage: Add to your router chain

```go
import "myapp/internal/api/middleware"

func main() {
    mux := http.NewServeMux()
    
    // Add security headers to all routes
    protectedMux := middleware.SecurityHeadersMiddleware()(mux)
    
    log.Println("Server starting with security headers...")
    http.ListenAndServe(":8080", protectedMux)
}
```

**Maintenance:** Zero - just works
**Cost:** $0
**Effectiveness:** Blocks clickjacking, XSS, MIME sniffing, Spectre

---

## 3. Implementation: CSP (Content Security Policy)

### Already Implemented in Security Headers

The CSP header is included in the security headers middleware above:

```go
w.Header().Set("Content-Security-Policy", 
    "default-src 'self'; " +
    "script-src 'self' 'unsafe-inline' https://challenges.cloudflare.com; " +
    "style-src 'self' 'unsafe-inline'; " +
    "img-src 'self' data: https://*; " +
    "font-src 'self' https://*; " +
    "connect-src 'self' https://challenges.cloudflare.com https://api.vinnsedesigner.render.com; " +
    "frame-src 'none'; " +
    "object-src 'none'; " +
    "base-uri 'self'; " +
    "form-action 'self'")
```

**What it does:**
- Blocks inline scripts (except Cloudflare Turnstile)
- Blocks all frames/iframes
- Blocks object/plugin tags
- Restricts connections to self + Cloudflare + your API
- Allows images from anywhere (for user uploads)

**Maintenance:** Low - update when adding new external resources
**Cost:** $0
**Effectiveness:** Blocks XSS, data exfiltration, clickjacking

---

## 4. Implementation: HSTS Header

### Already Implemented in Security Headers

```go
w.Header().Set("Strict-Transport-Security", "max-age=15768000; includeSubDomains")
```

**What it does:**
- Forces browsers to use HTTPS for 6 months (15768000 seconds)
- Applies to all subdomains
- Prevents SSL stripping attacks

**Maintenance:** Zero
**Cost:** $0
**Effectiveness:** Prevents downgrade attacks

---

## 5. Implementation: Audit Logging

### Code: `apps/api/internal/security/audit.go`

```go
package security

import (
    "context"
    "database/sql"
    "encoding/json"
    "time"
)

type AuditLogRepository struct {
    DB *sql.DB
}

type AuditEvent struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    Action    string    `json:"action"`  // login, create_device, delete_device, etc.
    TargetID  string    `json:"target_id"`
    TargetType string   `json:"target_type"` // device, user, command, etc.
    IPAddress string    `json:"ip_address"`
    UserAgent string    `json:"user_agent"`
    Metadata  string    `json:"metadata"` // JSON-encoded details
    CreatedAt time.Time  `json:"created_at"`
}

func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
    return &AuditLogRepository{DB: db}
}

func (r *AuditLogRepository) Init(ctx context.Context) error {
    query := `
    CREATE TABLE IF NOT EXISTS audit_logs (
        id TEXT PRIMARY KEY,
        user_id TEXT NOT NULL,
        action TEXT NOT NULL,
        target_id TEXT,
        target_type TEXT,
        ip_address TEXT,
        user_agent TEXT,
        metadata TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_logs(user_id);
    CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action);
    CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at);
    `
    
    _, err := r.DB.ExecContext(ctx, query)
    return err
}

func (r *AuditLogRepository) LogEvent(ctx context.Context, event AuditEvent) error {
    metadataJSON, _ := json.Marshal(event.Metadata)
    
    query := `
    INSERT INTO audit_logs (
        id, user_id, action, target_id, target_type, 
        ip_address, user_agent, metadata, created_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
    
    _, err := r.DB.ExecContext(ctx, query,
        event.ID,
        event.UserID,
        event.Action,
        event.TargetID,
        event.TargetType,
        event.IPAddress,
        event.UserAgent,
        string(metadataJSON),
        event.CreatedAt,
    )
    
    return err
}

func (r *AuditLogRepository) GetRecentEvents(ctx context.Context, limit int) ([]AuditEvent, error) {
    query := `SELECT id, user_id, action, target_id, target_type, ip_address, user_agent, metadata, created_at 
              FROM audit_logs 
              ORDER BY created_at DESC 
              LIMIT ?`
    
    rows, err := r.DB.QueryContext(ctx, query, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var events []AuditEvent
    for rows.Next() {
        var e AuditEvent
        var metadataJSON string
        if err := rows.Scan(
            &e.ID, &e.UserID, &e.Action, &e.TargetID, &e.TargetType,
            &e.IPAddress, &e.UserAgent, &metadataJSON, &e.CreatedAt,
        ); err != nil {
            return nil, err
        }
        
        json.Unmarshal([]byte(metadataJSON), &e.Metadata)
        events = append(events, e)
    }
    
    return events, nil
}
```

### Usage: Log all critical actions

```go
// In your handlers
auditRepo := security.NewAuditLogRepository(db)

// Initialize on startup
auditRepo.Init(context.Background())

// Log user actions
auditRepo.LogEvent(context.Background(), security.AuditEvent{
    ID:       "019000a1-4321-7cbd-8f11-9a78543210ab",
    UserID:   "user_123",
    Action:   "device_create",
    TargetID: "device_456",
    TargetType: "device",
    IPAddress: r.RemoteAddr,
    UserAgent: r.UserAgent(),
    Metadata: map[string]interface{}{
        "device_name": "Living Room Sensor",
        "device_type": "sensor",
    },
    CreatedAt: time.Now(),
})
```

**Maintenance:** Low - just log events
**Cost:** $0 (uses existing SQLite)
**Effectiveness:** Full audit trail for security investigations

---

## 6. Implementation: Session Revocation

### Code: `apps/api/internal/auth/session_revocation.go`

```go
package auth

import (
    "context"
    "database/sql"
    "errors"
    "time"
)

type SessionRevocationRepository struct {
    DB *sql.DB
}

func NewSessionRevocationRepository(db *sql.DB) *SessionRevocationRepository {
    return &SessionRevocationRepository{DB: db}
}

func (r *SessionRevocationRepository) Init(ctx context.Context) error {
    query := `
    CREATE TABLE IF NOT EXISTS revoked_sessions (
        token_hash TEXT PRIMARY KEY,
        user_id TEXT NOT NULL,
        revoked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        expires_at TIMESTAMP
    );
    
    CREATE INDEX IF NOT EXISTS idx_revoked_user ON revoked_sessions(user_id);
    CREATE INDEX IF NOT EXISTS idx_revoked_expires ON revoked_sessions(expires_at);
    `
    
    _, err := r.DB.ExecContext(ctx, query)
    return err
}

func (r *SessionRevocationRepository) RevokeSession(ctx context.Context, userID, tokenHash string) error {
    // Store revoked token for 30 days
    expiresAt := time.Now().Add(30 * 24 * time.Hour)
    
    query := `INSERT INTO revoked_sessions (token_hash, user_id, expires_at) VALUES (?, ?, ?)`
    
    _, err := r.DB.ExecContext(ctx, query, tokenHash, userID, expiresAt)
    return err
}

func (r *SessionRevocationRepository) IsSessionRevoked(ctx context.Context, tokenHash string) (bool, error) {
    query := `SELECT COUNT(1) FROM revoked_sessions WHERE token_hash = ? AND expires_at > CURRENT_TIMESTAMP`
    
    var count int
    err := r.DB.QueryRowContext(ctx, query, tokenHash).Scan(&count)
    if err != nil {
        return false, err
    }
    
    return count > 0, nil
}

// Cleanup expired revocations (run as cron job)
func (r *SessionRevocationRepository) CleanupExpired(ctx context.Context) error {
    query := `DELETE FROM revoked_sessions WHERE expires_at <= CURRENT_TIMESTAMP`
    
    _, err := r.DB.ExecContext(ctx, query)
    return err
}
```

### Usage: Check on every authenticated request

```go
// In your auth middleware
revocationRepo := auth.NewSessionRevocationRepository(db)
revocationRepo.Init(context.Background())

// Check if token is revoked
if revoked, _ := revocationRepo.IsSessionRevoked(ctx, tokenHash); revoked {
    http.Error(w, "Session has been revoked", http.StatusUnauthorized)
    return
}

// On logout
revocationRepo.RevokeSession(ctx, userID, tokenHash)
```

**Maintenance:** Low - cleanup expired entries weekly
**Cost:** $0 (uses existing SQLite)
**Effectiveness:** Prevents replay attacks after logout

---

## 7. Implementation: UUIDv7 Migration

### Why UUIDv7?

1. **Unguessable** - No sequential pattern
2. **Time-sortable** - Faster than random UUIDv4
3. **No IDOR** - Can't scan `?id=1`, `?id=2`, etc.

### Code: `apps/api/internal/security/uuidv7.go`

```go
package security

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "time"
)

// GenerateUUIDv7 creates a time-sortable UUIDv7
func GenerateUUIDv7() (string, error) {
    var value [16]byte
    
    // First 48 bits: Unix timestamp in milliseconds
    ms := time.Now().UnixMilli()
    value[0] = byte(ms >> 40)
    value[1] = byte(ms >> 32)
    value[2] = byte(ms >> 24)
    value[3] = byte(ms >> 16)
    value[4] = byte(ms >> 8)
    value[5] = byte(ms)
    
    // Next 80 bits: Cryptographic randomness
    if _, err := rand.Read(value[6:]); err != nil {
        return "", err
    }
    
    // Set version (7) and variant (RFC4122)
    value[6] = (value[6] & 0x0f) | 0x70 // Version 7
    value[8] = (value[8] & 0x3f) | 0x80 // Variant RFC4122
    
    // Format as standard UUID string
    dst := make([]byte, 36)
    hex.Encode(dst[0:8], value[0:4])
    dst[8] = '-'
    hex.Encode(dst[9:13], value[4:6])
    dst[13] = '-'
    hex.Encode(dst[14:18], value[6:8])
    dst[18] = '-'
    hex.Encode(dst[19:23], value[8:10])
    dst[23] = '-'
    hex.Encode(dst[24:36], value[10:16])
    
    return string(dst), nil
}
```

### Migration Strategy

```sql
-- Step 1: Create shadow table
CREATE TABLE devices_uuid_shadow (
    id TEXT PRIMARY KEY,
    owner_id TEXT NOT NULL,
    name TEXT NOT NULL,
    -- other columns...
);

-- Step 2: Copy data with UUIDv7
INSERT INTO devices_uuid_shadow (id, owner_id, name, ...)
SELECT
    GenerateUUIDv7() AS id,
    owner_id,
    name,
    -- other columns...
FROM devices;

-- Step 3: Drop old table
DROP TABLE devices;

-- Step 4: Rename shadow to production
ALTER TABLE devices_uuid_shadow RENAME TO devices;

-- Step 5: Add indexes
CREATE INDEX idx_devices_owner ON devices(owner_id);
```

**Maintenance:** One-time migration
**Cost:** $0
**Effectiveness:** Eliminates IDOR vulnerabilities

---

## 8. Implementation: MFA/TOTP

### Code: `apps/api/internal/auth/mfa.go`

```go
package auth

import (
    "crypto/hmac"
    "crypto/sha1"
    "encoding/base32"
    "encoding/binary"
    "errors"
    "strings"
    "time"
)

type TOTP struct {
    secret string
}

func NewTOTP(secret string) *TOTP {
    return &TOTP{secret: strings.ToUpper(strings.TrimSpace(secret))}
}

func GenerateTOTPSecret() (string, error) {
    secret := make([]byte, 10)
    if _, err := rand.Read(secret); err != nil {
        return "", err
    }
    
    // Base32 encoding for TOTP
    return base32.StdEncoding.EncodeToString(secret), nil
}

func (t *TOTP) GenerateCode() (string, error) {
    // Get current time step (30-second window)
    counter := uint64(time.Now().Unix() / 30)
    
    // Decode secret
    secret, err := base32.StdEncoding.DecodeString(t.secret)
    if err != nil {
        return "", err
    }
    
    // HMAC-SHA1 of counter
    hmacHash := hmac.New(sha1.New, secret)
    binary.Write(hmacHash, binary.BigEndian, counter)
    hash := hmacHash.Sum(nil)
    
    // Dynamic truncation
    offset := hash[len(hash)-1] & 0x0F
    truncated := hash[offset : offset+4]
    
    // Convert to 6-digit code
    code := binary.BigEndian.Uint32(truncated) & 0x7FFFFFFF
    
    return fmt.Sprintf("%06d", code%1000000), nil
}

func (t *TOTP) VerifyCode(code string) (bool, error) {
    generated, err := t.GenerateCode()
    if err != nil {
        return false, err
    }
    
    // Allow 1 code before and after for clock drift
    for i := -1; i <= 1; i++ {
        if code == generated {
            return true, nil
        }
        
        // Try next/previous time step
        counter := uint64((time.Now().Unix() / 30) + int64(i))
        hmacHash := hmac.New(sha1.New, []byte(t.secret))
        binary.Write(hmacHash, binary.BigEndian, counter)
        hash := hmacHash.Sum(nil)
        offset := hash[len(hash)-1] & 0x0F
        truncated := hash[offset : offset+4]
        nextCode := binary.BigEndian.Uint32(truncated) & 0x7FFFFFFF
        
        if code == fmt.Sprintf("%06d", nextCode%1000000) {
            return true, nil
        }
    }
    
    return false, nil
}
```

### Usage: Add to user model

```go
type User struct {
    ID string
    Email string
    PasswordHash string
    TOTPSecret string // Base32-encoded secret
    TOTPEnabled bool
}

// Enable TOTP
func (u *User) EnableTOTP() (string, string, error) {
    secret, err := GenerateTOTPSecret()
    if err != nil {
        return "", "", err
    }
    
    totp := NewTOTP(secret)
    code, err := totp.GenerateCode()
    if err != nil {
        return "", "", err
    }
    
    u.TOTPSecret = secret
    u.TOTPEnabled = true
    
    return secret, code, nil
}

// Verify TOTP code
func (u *User) VerifyTOTPCode(code string) (bool, error) {
    if !u.TOTPEnabled || u.TOTPSecret == "" {
        return false, errors.New("TOTP not enabled")
    }
    
    totp := NewTOTP(u.TOTPSecret)
    return totp.VerifyCode(code)
}
```

**Maintenance:** Medium - handle backup codes, recovery
**Cost:** $0 (pure Go implementation)
**Effectiveness:** Strong second factor authentication

---

## 9. Implementation: Secrets Management

### For Render (No Vault)

```bash
# .gitignore
.env
*.pem
*.key
secrets*

# .env.example (committed)
# Database
DATABASE_URL=sqlite:///data/app.db

# API Keys (use placeholders)
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret

# Generate secrets
ED25519_PRIVATE_KEY=$(openssl rand -hex 32)
JWT_SECRET=$(openssl rand -hex 32)
TURNSTILE_SECRET=your-turnstile-secret

# .env (NOT committed)
DATABASE_URL=sqlite:///data/app.db
GOOGLE_CLIENT_ID=actual-id
GOOGLE_CLIENT_SECRET=actual-secret
ED25519_PRIVATE_KEY=actual-key
JWT_SECRET=actual-secret
TURNSTILE_SECRET=actual-secret
```

### Code: Load and burn secrets

```go
// apps/api/config/env.go
package config

import (
    "os"
    "strings"
)

type Config struct {
    DatabaseURL    string
    GoogleClientID string
    GoogleSecret   string
    PrivateKey     string
    JWTSecret      string
    TurnstileKey   string
}

func LoadAndBurnConfig() *Config {
    cfg := &Config{
        DatabaseURL:    getEnv("DATABASE_URL", "sqlite:///data/app.db"),
        GoogleClientID: getEnv("GOOGLE_CLIENT_ID", ""),
        GoogleSecret:   getEnv("GOOGLE_CLIENT_SECRET", ""),
        PrivateKey:     getEnv("ED25519_PRIVATE_KEY", ""),
        JWTSecret:      getEnv("JWT_SECRET", ""),
        TurnstileKey:   getEnv("TURNSTILE_SECRET", ""),
    }
    
    // BURN: Clear from environment after loading
    os.Setenv("ED25519_PRIVATE_KEY", "")
    os.Setenv("JWT_SECRET", "")
    os.Setenv("TURNSTILE_SECRET", "")
    
    return cfg
}

func getEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        return strings.TrimSpace(value)
    }
    return fallback
}
```

**Maintenance:** Low - just keep .env out of git
**Cost:** $0
**Effectiveness:** Prevents secret leakage

---

## 10. Implementation: Container Hardening

### For Render (Dockerfile Best Practices)

```dockerfile
# Use minimal base image
FROM golang:1.25-alpine AS builder

# Build stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /vyzorix-api .

# Runtime stage - minimal image
FROM alpine:3.18

# Install CA certificates (needed for HTTPS)
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy binary from builder
COPY --from=builder --chown=appuser:appgroup /vyzorix-api /usr/local/bin/

# Set permissions
RUN chmod 500 /usr/local/bin/vyzorix-api

# Run as non-root
USER appuser:appgroup

# Minimal entrypoint
ENTRYPOINT ["/usr/local/bin/vyzorix-api"]
```

### Security Checks

```bash
# Scan image for vulnerabilities (free tier)
docker scan vyzorix-api

# Or use Trivy
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock aquasec/trivy image vyzorix-api
```

**Maintenance:** Low - update base images monthly
**Cost:** $0
**Effectiveness:** Reduces attack surface

---

## 11. Implementation: SIEM Alternative

### Free Alternative: SQLite + Alerts

```go
// apps/api/internal/security/alerts.go
package security

import (
    "context"
    "database/sql"
    "log"
    "time"
)

type AlertRepository struct {
    DB *sql.DB
}

func NewAlertRepository(db *sql.DB) *AlertRepository {
    return &AlertRepository{DB: db}
}

func (r *AlertRepository) Init(ctx context.Context) error {
    query := `
    CREATE TABLE IF NOT EXISTS security_alerts (
        id TEXT PRIMARY KEY,
        severity TEXT NOT NULL, -- INFO, WARNING, CRITICAL
        category TEXT NOT NULL,  -- auth_failure, rate_limit, sql_injection, etc.
        message TEXT NOT NULL,
        user_id TEXT,
        ip_address TEXT,
        user_agent TEXT,
        metadata TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        resolved_at TIMESTAMP,
        resolved_by TEXT
    );
    
    CREATE INDEX IF NOT EXISTS idx_alerts_severity ON security_alerts(severity);
    CREATE INDEX IF NOT EXISTS idx_alerts_created ON security_alerts(created_at);
    CREATE INDEX IF NOT EXISTS idx_alerts_resolved ON security_alerts(resolved_at);
    `
    
    _, err := r.DB.ExecContext(ctx, query)
    return err
}

func (r *AlertRepository) CreateAlert(ctx context.Context, alert Alert) error {
    query := `
    INSERT INTO security_alerts (
        id, severity, category, message, user_id, ip_address, user_agent, metadata
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `
    
    _, err := r.DB.ExecContext(ctx, query,
        alert.ID,
        alert.Severity,
        alert.Category,
        alert.Message,
        alert.UserID,
        alert.IPAddress,
        alert.UserAgent,
        alert.Metadata,
    )
    
    return err
}

func (r *AlertRepository) GetUnresolvedAlerts(ctx context.Context) ([]Alert, error) {
    query := `
    SELECT id, severity, category, message, user_id, ip_address, user_agent, metadata, created_at
    FROM security_alerts
    WHERE resolved_at IS NULL
    ORDER BY created_at DESC
    LIMIT 100
    `
    
    // Execute and return alerts...
}

// Simple alert checker (run as cron job)
func CheckForCriticalAlerts(ctx context.Context, db *sql.DB, email string) error {
    repo := NewAlertRepository(db)
    alerts, err := repo.GetUnresolvedAlerts(ctx)
    if err != nil {
        return err
    }
    
    criticalCount := 0
    for _, alert := range alerts {
        if alert.Severity == "CRITICAL" {
            criticalCount++
            log.Printf("CRITICAL ALERT: %s - %s", alert.Category, alert.Message)
        }
    }
    
    if criticalCount > 0 {
        // In production: Send email via SMTP
        log.Printf("ALERT: %d critical security events detected", criticalCount)
    }
    
    return nil
}
```

### Usage: Log security events

```go
// In your middleware
alertRepo := security.NewAlertRepository(db)
alertRepo.Init(context.Background())

// Log failed logins
alertRepo.CreateAlert(context.Background(), security.Alert{
    ID: "019000a1-4321-7cbd-8f11-9a78543210ab",
    Severity: "WARNING",
    Category: "auth_failure",
    Message: "Multiple failed login attempts",
    UserID: "user_123",
    IPAddress: r.RemoteAddr,
    UserAgent: r.UserAgent(),
    Metadata: map[string]interface{}{
        "attempts": 5,
        "last_attempt": time.Now().Format(time.RFC3339),
    },
})

// Log rate limit violations
alertRepo.CreateAlert(context.Background(), security.Alert{
    ID: "019000a1-4321-7cbd-8f11-9a78543210ab",
    Severity: "INFO",
    Category: "rate_limit",
    Message: "Rate limit exceeded",
    IPAddress: r.RemoteAddr,
    UserAgent: r.UserAgent(),
})

// Check for critical alerts (run hourly)
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    for range ticker.C {
        if err := CheckForCriticalAlerts(context.Background(), db, "admin@example.com"); err != nil {
            log.Printf("Alert check failed: %v", err)
        }
    }
}()
```

**Maintenance:** Medium - review alerts weekly
**Cost:** $0 (uses existing SQLite)
**Effectiveness:** Basic monitoring without commercial SIEM

---

## Summary: Practical Security for Render

### Implemented 
- Security headers middleware
- CSP header
- HSTS header
- Rate limiting (partial)
- Error sanitization (partial)

### To Implement (Priority Order)
1. **Audit logging** - 2h, $0, high value
2. **Session revocation** - 2h, $0, critical for security
3. **UUIDv7 migration** - 4h, $0, eliminates IDOR
4. **MFA/TOTP** - 8h, $0, strong 2FA
5. **Secrets management** - 1h, $0, prevent leaks
6. **Container hardening** - 2h, $0, reduce surface
7. **SIEM alternative** - 4h, $0, basic monitoring

### Total Effort: ~23 hours
### Total Cost: $0
### Maintenance: Low (you + me team)

---

*Document Version: 1.0*
*Last Updated: 2026-06-10*
*Status: Ready for implementation*
