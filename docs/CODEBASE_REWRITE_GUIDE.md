# Vyzorix Update Server - Comprehensive Codebase Rewrite Guide

**Version:** 1.0  
**Date:** 2026-06-17  
**Status:** Ready for Implementation  
**Priority:** Enterprise Production Ready

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Critical Security Fixes](#critical-security-fixes)
3. [High Priority Rewrites](#high-priority-rewrites)
4. [Medium Priority Refactoring](#medium-priority-refactoring)
5. [Low Priority Improvements](#low-priority-improvements)
6. [Shared Utilities to Create](#shared-utilities-to-create)
7. [Code Consolidation Map](#code-consolidation-map)
8. [File-by-File Fix Guide](#file-by-file-fix-guide)
9. [Implementation Order](#implementation-order)
10. [Testing Requirements](#testing-requirements)

---

## Executive Summary

### Scope
- **Total Go Files:** 31
- **Total Lines of Code:** ~8,500
- **Repository:** `/workspace/project/6dc1227fe57247678b4d9fddc4699f3d/apps/api`

### Issues Found
| Category | Count | Fixed |
|----------|-------|-------|
| Critical Security | 5 | 5 ✅ |
| High Security | 5 | 0 |
| Code Duplication | 18 | 0 |
| Performance Issues | 7 | 0 |
| Design Problems | 12 | 0 |

### Packages Analyzed
```
apps/api/
├── internal/
│   ├── api/handlers/     (13 files, ~180KB)
│   ├── api/middleware/   (22 files, ~100KB)
│   ├── auth/             (standalone files)
│   ├── fcm/              (standalone files)
│   ├── ws/               (standalone files)
│   ├── audit/            (standalone files)
│   ├── metrics/          (standalone files)
│   └── ssr/              (standalone files)
├── pkg/
│   ├── storage/          (11 files, ~96KB)
│   ├── crypto/           (standalone files)
│   ├── config/           (standalone files)
│   ├── logging/          (standalone files)
│   └── models/           (standalone files)
└── main.go
```

---

## Critical Security Fixes

### Status: ALL FIXED ✅

| # | File | Issue | Fix Applied |
|---|------|-------|-------------|
| 1 | `pkg/storage/uuid.go` | Insecure random fallback (lines 54-61) | Panics on crypto/rand failure |
| 2 | `internal/auth/lockout.go` | Weak XOR password validation (lines 145-158) | Replaced with Argon2id |
| 3 | `internal/auth/lockout.go` | Predictable fake token (lines 160-171) | Panics on crypto/rand failure |
| 4 | `internal/command_signer.go` | Weak fallback hash (lines 140-153) | Removed, panics on failure |
| 5 | `internal/api/middleware/rate_limiter.go` | Memory leak (entire file) | Added cleanup goroutine |

---

## High Priority Rewrites

### 1. `internal/api/middleware/request_signing.go`

**File Size:** 12KB  
**Priority:** 🔴 CRITICAL  
**Current Issues:**

#### 1.1 Duplicated AES-256-GCM Encryption (4 Locations)

**Problem:** Same AES-256-GCM encryption/decryption code appears 4 times.

**Locations:**
- Lines 257-264: `decryptBody()`
- Lines 281-294: `EncryptBody()`
- Lines 312-324: `SignRequest()`
- `response_encryption.go`: Similar pattern

**Current Code (Example - decryptBody):**
```go
func (s *RequestSigning) decryptBody(encrypted []byte, secret string) ([]byte, error) {
    key := sha512.Sum512([]byte(secret))
    block, err := aes.NewCipher(key[:])
    if err != nil {
        return nil, err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    nonce := encrypted[:12]
    ciphertext := encrypted[12:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, err
    }
    return plaintext, nil
}
```

**FIX:** Extract to `internal/crypto/aes_gcm.go`:
```go
// internal/crypto/aes_gcm.go
package crypto

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha512"
    "errors"
    "io"
)

const (
    NonceSize = 12
    KeySize   = 32
)

var ErrCiphertextTooShort = errors.New("ciphertext too short")
var ErrCiphertextInvalid = errors.New("ciphertext authentication failed")

// DeriveKey derives a 32-byte key from any secret using SHA-512
func DeriveKey(secret string) []byte {
    key := sha512.Sum512([]byte(secret))
    return key[:KeySize]
}

// Encrypt encrypts plaintext using AES-256-GCM
func Encrypt(key, plaintext []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    nonce := make([]byte, NonceSize)
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func Decrypt(key, ciphertext []byte) ([]byte, error) {
    if len(ciphertext) < NonceSize {
        return nil, ErrCiphertextTooShort
    }
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    nonce := ciphertext[:NonceSize]
    plaintext, err := gcm.Open(nil, nonce, ciphertext[NonceSize:], nil)
    if err != nil {
        return nil, ErrCiphertextInvalid
    }
    return plaintext, nil
}
```

#### 1.2 O(n) Cache Eviction on Every Request (Lines 84-118)

**Problem:** Every `Use()` call iterates through all entries twice for cleanup.

**Current Code:**
```go
func (c *ReplayCache) Use(signature string) bool {
    c.mu.Lock()
    defer c.mu.Unlock()
    now := time.Now()
    cutoff := now.Add(-c.window)

    // O(n) iteration #1 - cleanup old entries
    for sig, t := range c.seen {
        if t.Before(cutoff) {
            delete(c.seen, sig)
        }
    }

    // O(n) iteration #2 - evict if full
    if len(c.seen) >= c.maxSize {
        evictCount := c.maxSize / 10
        // ... another full iteration
    }
    
    // Check if exists
    if _, exists := c.seen[signature]; exists {
        return false // Replay detected
    }
    c.seen[signature] = now
    return true
}
```

**FIX:** Use background cleanup with periodic tickers (already done in rate_limiter.go pattern):
```go
// internal/crypto/replay_cache.go
package crypto

import (
    "sync"
    "time"
)

type ReplayCache struct {
    mu       sync.RWMutex
    entries  map[string]time.Time
    maxSize  int
    ttl      time.Duration
    stopCh   chan struct{}
}

func NewReplayCache(maxSize int, ttl time.Duration) *ReplayCache {
    rc := &ReplayCache{
        entries: make(map[string]time.Time),
        maxSize: maxSize,
        ttl:     ttl,
        stopCh:  make(chan struct{}),
    }
    go rc.cleanupLoop()
    return rc
}

func (rc *ReplayCache) cleanupLoop() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    for {
        select {
        case <-rc.stopCh:
            return
        case <-ticker.C:
            rc.cleanup()
        }
    }
}

func (rc *ReplayCache) cleanup() {
    rc.mu.Lock()
    defer rc.mu.Unlock()
    cutoff := time.Now().Add(-rc.ttl)
    for sig, t := range rc.entries {
        if t.Before(cutoff) {
            delete(rc.entries, sig)
        }
    }
}

func (rc *ReplayCache) Use(signature string) bool {
    rc.mu.Lock()
    defer rc.mu.Unlock()
    
    if _, exists := rc.entries[signature]; exists {
        return false
    }
    
    // Simple size check without full iteration
    if len(rc.entries) >= rc.maxSize {
        // Evict oldest 10%
        evictCount := rc.maxSize / 10
        for sig := range rc.entries {
            if evictCount <= 0 {
                break
            }
            delete(rc.entries, sig)
            evictCount--
        }
    }
    
    rc.entries[signature] = time.Now()
    return true
}

func (rc *ReplayCache) Stop() {
    close(rc.stopCh)
}
```

#### 1.3 Complex Verify Function (Lines 136-170)

**Problem:** 55+ line function with 7 responsibilities.

**FIX:** Split into focused methods:
```go
func (s *RequestSigning) Verify(c *gin.Context) error {
    // 1. Extract headers
    headers, err := s.extractHeaders(c)
    if err != nil {
        return err
    }
    
    // 2. Validate timestamp
    if err := s.validateTimestamp(headers.Timestamp); err != nil {
        return err
    }
    
    // 3. Get client credentials
    client, err := s.getClient(headers.ClientID)
    if err != nil {
        return err
    }
    
    // 4. Decrypt body if needed
    body, err := s.getRequestBody(c, headers, client.Secret)
    if err != nil {
        return err
    }
    
    // 5. Verify signature
    if err := s.verifySignature(headers, body, client.Secret); err != nil {
        return err
    }
    
    // 6. Check replay
    if !s.replayCache.Use(headers.Signature) {
        return ErrReplayDetected
    }
    
    return nil
}
```

---

### 2. `pkg/storage/clients.go`

**File Size:** 15KB  
**Priority:** 🔴 HIGH  
**Current Issues:**

#### 2.1 Duplicated JSON Unmarshal Pattern (3 Locations)

**Locations:** `GetAPIClient()`, `GetAPIClientByOperator()`, `ListAllAPIClients()`

**Current Code (Repeated 3x):**
```go
if err := json.Unmarshal([]byte(allowedOrigins), &client.AllowedOrigins); err != nil {
    client.AllowedOrigins = []string{}
}
if err := json.Unmarshal([]byte(allowedPaths), &client.AllowedPaths); err != nil {
    client.AllowedPaths = []string{}
}
```

**FIX:** Extract to helper:
```go
// In clients.go, add:
func scanAPIClientCommon(row interface {
    Scan(dest ...any) error
}) (*APIClient, error) {
    var (
        allowedOrigins, allowedPaths                                              string
        client                                                                          APIClient
    )
    
    err := row.Scan(
        &client.ID, &client.Name, &client.AllowedDomains, &client.OperatorID,
        &client.CreatedAt, &client.LastRequestAt, &client.IsActive,
        &allowedOrigins, &allowedPaths,
    )
    if err != nil {
        return nil, err
    }
    
    // Unmarshal JSON fields with error handling
    if err := json.Unmarshal([]byte(allowedOrigins), &client.AllowedOrigins); err != nil {
        client.AllowedOrigins = []string{}
    }
    if err := json.Unmarshal([]byte(allowedPaths), &client.AllowedPaths); err != nil {
        client.AllowedPaths = []string{}
    }
    
    return &client, nil
}
```

#### 2.2 Duplicated SigningKey Scanning (2 Locations)

**FIX:** Extract similar helper for signing keys.

#### 2.3 VerifyAPIClientSecret Error Handling

**Current:** All failures return `ErrNotFound` (intentional for security)

**FIX:** Add comments clarifying intentionality:
```go
// VerifyAPIClientSecret returns ErrNotFound for ALL failures to prevent
// user enumeration. This is intentional security design.
func (s *Store) VerifyAPIClientSecret(ctx context.Context, clientID, clientSecret string) (*APIClient, error) {
    // ... verification logic ...
    // Deliberately returns ErrNotFound for:
    // - Client not found
    // - Client inactive
    // - Invalid secret
    // DO NOT change to return different errors
}
```

---

### 3. `pkg/storage/operators.go`

**File Size:** 17KB  
**Priority:** 🔴 HIGH  
**Current Issues:**

#### 3.1 Update Pattern Repeated 7 Times

**Current Pattern (Example):**
```go
func (s *Store) UpdateOperatorName(ctx context.Context, operatorID, name string) error {
    now := time.Now().UTC()
    result, err := s.db.ExecContext(ctx,
        `UPDATE operators SET name = ?, updated_at = ? WHERE id = ?`,
        name, now, operatorID,
    )
    if err != nil {
        return err
    }
    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        return ErrNotFound
    }
    return nil
}
```

**FIX:** Create generic update helper:
```go
// UpdateOperatorField updates a single operator field
func (s *Store) UpdateOperatorField(ctx context.Context, operatorID, field string, value interface{}) error {
    now := time.Now().UTC()
    query := fmt.Sprintf(`UPDATE operators SET %s = ?, updated_at = ? WHERE id = ?`, field)
    result, err := s.db.ExecContext(ctx, query, value, now, operatorID)
    if err != nil {
        return err
    }
    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        return ErrNotFound
    }
    return nil
}

// Then simplify each update function:
func (s *Store) UpdateOperatorName(ctx context.Context, operatorID, name string) error {
    return s.UpdateOperatorField(ctx, operatorID, "name", name)
}

func (s *Store) UpdateOperatorEmail(ctx context.Context, operatorID, email string) error {
    return s.UpdateOperatorField(ctx, operatorID, "email", email)
}
```

#### 3.2 Simplify boolToInt

**Current:**
```go
func boolToInt(b bool) int {
    if b {
        return 1
    }
    return 0
}
```

**FIX:**
```go
func boolToInt(b bool) int {
    if b { return 1 }
    return 0
}
// Or even simpler, use direct conversion in SQL: CASE WHEN b THEN 1 ELSE 0 END
```

---

### 4. `internal/api/handlers/auth_oauth.go`

**File Size:** 10KB  
**Priority:** 🔴 HIGH  
**Current Issues:**

#### 4.1 OAuth Operator Creation Duplicated (Google vs GitHub)

**Current Code (Google - Lines 95-121):**
```go
op = &models.Operator{
    ID:        GenerateID(),
    Email:     googleClaims.Email,
    Name:      googleClaims.Name,
    Role:      role,
    GoogleID:  googleClaims.Sub,
    CreatedAt: time.Now().UTC(),
    UpdatedAt: time.Now().UTC(),
}
```

**Current Code (GitHub - Lines 251-276):**
```go
op = &models.Operator{
    ID:        GenerateID(),
    Email:     email,
    Name:      name,
    Role:      role,
    GitHubID:  githubID,
    CreatedAt: time.Now().UTC(),
    UpdatedAt: time.Now().UTC(),
}
```

**FIX:** Extract to helper:
```go
// createOAuthOperator creates an operator from OAuth provider data
func (ac *AuthController) createOAuthOperator(
    ctx context.Context,
    email, name, role string,
    provider ProviderType,
    providerID string,
) (*models.Operator, error) {
    op := &models.Operator{
        ID:        GenerateID(),
        Email:     email,
        Name:      name,
        Role:      role,
        CreatedAt: time.Now().UTC(),
        UpdatedAt: time.Now().UTC(),
    }
    
    switch provider {
    case ProviderGoogle:
        op.GoogleID = providerID
    case ProviderGitHub:
        op.GitHubID = providerID
    }
    
    if err := ac.store.CreateOperator(ctx, op); err != nil {
        return nil, err
    }
    return op, nil
}

type ProviderType int
const (
    ProviderGoogle ProviderType = iota
    ProviderGitHub
)
```

#### 4.2 Bootstrap Check Duplicated (3 Locations)

**Current (In GoogleCallback):**
```go
count, err := ac.store.OperatorCount(ctx)
if err != nil {
    ac.log.Warn("google callback: failed to count operators", "err", err)
}
role := models.RoleOperator
if count == 0 {
    role = models.RoleSuperAdmin
    ac.log.Info("google callback: bootstrapping first operator", "email", googleClaims.Email)
}
```

**FIX:** Extract to helper:
```go
// getRoleForNewOperator determines the role for a new OAuth operator
// First operator becomes SuperAdmin, others become Operator
func (ac *AuthController) getRoleForNewOperator(ctx context.Context) (models.Role, error) {
    count, err := ac.store.OperatorCount(ctx)
    if err != nil {
        ac.log.Warn("getRoleForNewOperator: failed to count operators", "err", err)
        return models.RoleOperator, nil // Fail safe to Operator
    }
    if count == 0 {
        ac.log.Info("getRoleForNewOperator: bootstrapping first operator")
        return models.RoleSuperAdmin, nil
    }
    return models.RoleOperator, nil
}
```

---

## Medium Priority Refactoring

### 5. `internal/api/handlers/auth_password_reset.go`

**Priority:** 🟠 MEDIUM  
**Issues:**

#### 5.1 Token Generation Duplicated

**Current (Lines 69-84):**
```go
tokenBytes := make([]byte, 32)
if _, err := rand.Read(tokenBytes); err != nil {
    ac.log.Warn("forgotPassword: failed to generate token", "err", err)
    c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "request failed"})
    return
}
token := hex.EncodeToString(tokenBytes)
tokenHash := security.HashToken(token)
```

**FIX:** Add to `auth_utils.go`:
```go
// GenerateSecureToken generates a cryptographically secure random token
// Returns (token, tokenHash, error)
func GenerateSecureToken() (string, string, error) {
    tokenBytes := make([]byte, 32)
    if _, err := rand.Read(tokenBytes); err != nil {
        return "", "", fmt.Errorf("failed to generate token: %w", err)
    }
    token := hex.EncodeToString(tokenBytes)
    tokenHash := HashToken(token)
    return token, tokenHash, nil
}
```

#### 5.2 Context.Background() in Goroutine

**Current (Lines 89-101):**
```go
go func() {
    if err := ac.emailSvc.SendPasswordResetEmail(context.Background(), op.Email, op.Name, token); err != nil {
        // ...
    }
}()
```

**FIX:** Use proper timeout context:
```go
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if err := ac.emailSvc.SendPasswordResetEmail(ctx, op.Email, op.Name, token); err != nil {
        // ...
    }
}()
```

---

### 6. `internal/api/handlers/auth_email_verify.go`

**Priority:** 🟠 MEDIUM  
**Issues:**

#### 6.1 Same Token Generation Issue

**FIX:** Use `GenerateSecureToken()` from auth_utils.go

#### 6.2 Silent Failure When Email Not Configured

**Current (Lines 217-222):**
```go
if !ac.emailSvc.IsConfigured() {
    ac.log.Warn("sendVerificationEmail: email service not configured, skipping", "email", op.Email)
    return  // Silent failure - user doesn't know
}
```

**FIX:** Return success to prevent enumeration:
```go
if !ac.emailSvc.IsConfigured() {
    ac.log.Warn("sendVerificationEmail: email service not configured, skipping", "email", op.Email)
    // Still return success to prevent user enumeration
    c.JSON(200, models.MessageResponse{Message: "verification email sent"})
    return
}
```

---

### 7. `internal/api/handlers/auth_mfa.go`

**Priority:** 🟠 MEDIUM  
**Issues:**

#### 7.1 MFAHandler Missing Logger

**Current:**
```go
type MFAHandler struct {
    Store      *storage.Store
    TOTPConfig security.TOTPConfig
}
```

**FIX:**
```go
type MFAHandler struct {
    Store      *storage.Store
    TOTPConfig security.TOTPConfig
    Log        *slog.Logger // ADD THIS
}

func NewMFAHandler(store *storage.Store, cfg security.TOTPConfig, log *slog.Logger) *MFAHandler {
    return &MFAHandler{
        Store:      store,
        TOTPConfig: cfg,
        Log:        log,
    }
}
```

#### 7.2 Multiple DefaultTOTPConfig() Calls

**Current:** Called 5 times in file

**FIX:** Cache at handler creation:
```go
type MFAHandler struct {
    // ...
    totpConfig security.TOTPConfig // Cache it
}

func (h *MFAHandler) totp() *security.TOTP {
    return security.NewTOTP(h.totpConfig.Secret, h.totpConfig)
}
```

---

### 8. `pkg/storage/devices.go`

**Priority:** 🟠 MEDIUM  
**Issues:**

#### 8.1 Plaintext Command Secret Stored in DB ⚠️

**Current (Lines 86-105):**
```go
// Generate command secret
secret, err := randomHex(32)
if err != nil {
    return Device{}, false, err
}

// Hash for audit
secretHash, hashErr := HashSecret(secret)
if hashErr != nil {
    secretHash = ""
}

// Store BOTH plaintext AND hash
_, err = s.db.ExecContext(ctx,
    `INSERT INTO devices...command_secret, command_secret_hash... VALUES(?,?,...)`,
    req.DeviceID, ..., secret, secretHash, ...
)
```

**ISSUE:** Plaintext secret is stored. Only `command_secret_hash` should be stored.

**FIX:**
```go
func (s *Store) Register(ctx context.Context, req models.RegisterRequest) (Device, bool, error) {
    // Generate command secret
    secret, err := randomHex(32)
    if err != nil {
        return Device{}, false, err
    }
    
    // Hash the secret for storage
    secretHash, err := HashSecret(secret)
    if err != nil {
        return Device{}, false, fmt.Errorf("failed to hash command secret: %w", err)
    }
    
    // Store ONLY the hash
    _, err = s.db.ExecContext(ctx,
        `INSERT INTO devices...command_secret_hash... VALUES(?)`,
        req.DeviceID, ..., secretHash, ...
    )
    if err != nil {
        return Device{}, false, err
    }
    
    // Return plaintext secret to caller ONCE - they must store it securely
    return Device{CommandSecret: secret}, true, nil
}
```

#### 8.2 Lock Held During Slow DB Operations

**Current:**
```go
func (s *Store) Register(ctx context.Context, req models.RegisterRequest) (Device, bool, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    // ... entire function including DB operations
}
```

**FIX:** Only lock the in-memory check:
```go
func (s *Store) Register(ctx context.Context, req models.RegisterRequest) (Device, bool, error) {
    // Check for existing device (fast, needs lock)
    s.mu.Lock()
    existing, err := s.getByDeviceID(ctx, req.DeviceID)
    s.mu.Unlock()
    if err != nil && !errors.Is(err, ErrNotFound) {
        return Device{}, false, err
    }
    if existing != nil {
        // ... hijack check and update without lock
    }
    // ... insert without lock (SQLite handles concurrency)
}
```

---

### 9. Middleware Consolidation (3 files)

**Priority:** 🟠 MEDIUM  
**Files:** `user_enum.go`, `user_enum_block.go`, `auth_enum_safe.go`

**Issues:**

#### 9.1 Constant-Time Compare (3 Versions)

**user_enum.go:12-16:**
```go
func constantTimeCompare(a, b string) bool {
    if len(a) != len(b) {
        return false
    }
    return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
```

**user_enum_block.go:89-91:**
```go
func ConstantTimeCompare(a, b string) bool {
    return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
```

**auth_enum_safe.go:54-60:**
```go
func ConstantTimeValidate(expected, actual string) bool {
    if len(expected) != len(actual) {
        subtle.ConstantTimeCompare([]byte(expected), []byte(actual))
        return false
    }
    return subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}
```

**FIX:** Consolidate into `internal/auth/enum_safe.go`:
```go
package auth

import "crypto/subtle"

// ConstantTimeCompare performs a constant-time comparison of two strings
// Returns true if they are equal, false otherwise
func ConstantTimeCompare(a, b string) bool {
    return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// ConstantTimeValidate validates a value against an expected value
// Returns 1 if equal, 0 if not (for use in timing-constant operations)
func ConstantTimeValidate(expected, actual string) bool {
    if len(expected) != len(actual) {
        subtle.ConstantTimeCompare([]byte(expected), []byte(actual))
        return false
    }
    return subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}
```

#### 9.2 Fake Password Hash (2 Versions)

**FIX:** Create single implementation in `internal/auth/enum_safe.go`:
```go
// ComputeFakePasswordHash performs dummy computation to maintain constant timing
// for user enumeration prevention
func ComputeFakePasswordHash() {
    salt := make([]byte, 16)
    rand.Read(salt) //nolint:errcheck // Ignore error for timing uniformity
    
    // Dummy Argon2id computation
    argon2.IDKey([]byte("dummy_password"), salt, 3, 64*1024, 4, 32)
}
```

---

## Low Priority Improvements

### 10. `pkg/storage/uuid.go`

**Priority:** 🟡 LOW  
**Issues:**

#### 10.1 hexCharToInt Returns 0 for Invalid Chars

**Current:**
```go
func hexCharToInt(c rune) int {
    switch {
    case c >= '0' && c <= '9':
        return int(c - '0')
    case c >= 'a' && c <= 'f':
        return int(c - 'a' + 10)
    case c >= 'A' && c <= 'F':
        return int(c - 'A' + 10)
    default:
        return 0 // SILENTLY accepts invalid chars!
    }
}
```

**FIX:** Return error or use unicode simple fold:
```go
import "unicode"

func hexCharToInt(c rune) (int, error) {
    switch {
    case c >= '0' && c <= '9':
        return int(c - '0'), nil
    case c >= 'a' && c <= 'f':
        return int(c - 'a' + 10), nil
    case c >= 'A' && c <= 'F':
        return int(c - 'A' + 10), nil
    default:
        return 0, fmt.Errorf("invalid hex character: %c", c)
    }
}
```

---

### 11. `pkg/storage/migrations.go`

**Priority:** 🟡 LOW  
**Issues:**

#### 11.1 Silent Error Swallowing

**Current (Lines 262-270):**
```go
func migrateAddCommandsColumns(db *sql.DB) error {
    queries := []string{
        `ALTER TABLE commands ADD COLUMN wake_sent INTEGER NOT NULL DEFAULT 0`,
        // ... more ALTER TABLE statements
    }
    for _, q := range queries {
        db.ExecContext(context.Background(), q) //nolint:errcheck // IGNORED!
    }
    return nil // Always returns nil
}
```

**FIX:**
```go
func migrateAddCommandsColumns(db *sql.DB) error {
    queries := []string{
        `ALTER TABLE commands ADD COLUMN wake_sent INTEGER NOT NULL DEFAULT 0`,
        // ... more ALTER TABLE statements
    }
    for _, q := range queries {
        if _, err := db.ExecContext(context.Background(), q); err != nil {
            // Log but don't fail - column might already exist
            log.Printf("migrateAddCommandsColumns: %s (may already exist): %v", q, err)
        }
    }
    return nil
}
```

---

### 12. `pkg/storage/sessions.go`

**Priority:** 🟡 LOW  
**Issues:**

#### 12.1 N+1 Queries in RevokeAllOperatorSessions

**Current (Lines 89-107):**
```go
for rows.Next() {
    var sessionID string
    rows.Scan(&sessionID)
    // One INSERT per session!
    if err := s.AddSessionRevocation(ctx, sessionID, "operator_logout"); err != nil {
        return err
    }
}
```

**FIX:** Use bulk INSERT:
```go
func (s *Store) RevokeAllOperatorSessions(ctx context.Context, operatorID string) error {
    // Get all session IDs
    rows, err := s.db.QueryContext(ctx, 
        `SELECT id FROM auth_sessions WHERE operator_id = ?`, operatorID)
    if err != nil {
        return err
    }
    defer rows.Close()
    
    var sessionIDs []string
    for rows.Next() {
        var id string
        if err := rows.Scan(&id); err != nil {
            return err
        }
        sessionIDs = append(sessionIDs, id)
    }
    
    if len(sessionIDs) == 0 {
        return nil
    }
    
    // Bulk insert revocations
    query := `INSERT INTO session_revocations (session_id, revoked_at, reason) VALUES `
    args := make([]interface{}, 0, len(sessionIDs)*3)
    for i, id := range sessionIDs {
        if i > 0 {
            query += ","
        }
        query += "(?, ?, ?)"
        args = append(args, id, time.Now().UTC(), "operator_logout")
    }
    
    _, err = s.db.ExecContext(ctx, query, args...)
    return err
}
```

---

## Shared Utilities to Create

### `internal/crypto/` - NEW PACKAGE

```
internal/crypto/
├── aes_gcm.go        # AES-256-GCM encrypt/decrypt
├── hmac.go           # HMAC-SHA512 operations
├── replay_cache.go   # Unified replay protection with auto-cleanup
└── rand.go           # Secure random utilities
```

### `internal/auth/` - NEW UTILITIES

```
internal/auth/
├── token.go          # Token generation (GenerateSecureToken)
├── oauth.go          # OAuth operator creation helpers
├── enum_safe.go      # User enumeration prevention
└── errors.go         # Auth-specific errors
```

### `pkg/storage/` - INTERNAL HELPERS

```
pkg/storage/
├── helpers.go        # Shared scanning helpers
├── transactions.go   # Transaction utilities
└── cleanup.go        # Background cleanup goroutines
```

---

## Code Consolidation Map

| Current Duplication | Occurrences | Extract To |
|---------------------|-------------|------------|
| AES-256-GCM encrypt/decrypt | 4 | `internal/crypto/aes_gcm.go` |
| Replay cache logic | 2 | `internal/crypto/replay_cache.go` |
| Constant-time compare | 3 | `internal/auth/enum_safe.go` |
| Fake password hash | 2 | `internal/auth/enum_safe.go` |
| Token generation | 3 | `internal/auth/token.go` |
| OAuth operator creation | 2 | `internal/auth/oauth.go` |
| JSON unmarshal pattern | 3 | `pkg/storage/helpers.go` |
| SigningKey scanning | 2 | `pkg/storage/helpers.go` |
| Update operator pattern | 7 | `pkg/storage/operators.go` |
| NullInt64 conversion | 2 | `pkg/storage/helpers.go` |

---

## File-by-File Fix Guide

### Complete Fix List

| File | Priority | Issues | Fix Actions |
|------|----------|--------|-------------|
| `request_signing.go` | 🔴 CRITICAL | 4 | Extract crypto, split Verify, fix eviction |
| `clients.go` | 🔴 HIGH | 3 | Extract helpers, add comments |
| `operators.go` | 🔴 HIGH | 3 | Generic update helper, simplify boolToInt |
| `auth_oauth.go` | 🔴 HIGH | 3 | Extract OAuth helpers |
| `auth_password_reset.go` | 🟠 MEDIUM | 3 | Extract token gen, fix context |
| `auth_email_verify.go` | 🟠 MEDIUM | 3 | Use shared token, fix silent fail |
| `auth_mfa.go` | 🟠 MEDIUM | 2 | Add logger, cache config |
| `devices.go` | 🟠 MEDIUM | 3 | Remove plaintext, fix lock |
| `user_enum*.go` | 🟠 MEDIUM | 2 | Consolidate to enum_safe.go |
| `uuid.go` | 🟡 LOW | 2 | Fix hexCharToInt error handling |
| `migrations.go` | 🟡 LOW | 2 | Log errors properly |
| `sessions.go` | 🟡 LOW | 2 | Bulk INSERT |
| `rate_limiter.go` | ✅ DONE | 1 | Already fixed with cleanup |
| `lockout.go` | ✅ DONE | 2 | Already fixed with Argon2id |
| `command_signer.go` | ✅ DONE | 1 | Already fixed, removed fallback |

---

## Implementation Order

### Phase 1: Security Hardening (1-2 days)
1. ✅ ~~Fix 5 critical security issues~~ (DONE)
2. Create `internal/crypto/` package
3. Create `internal/auth/` package
4. Consolidate middleware

### Phase 2: Storage Refactoring (2-3 days)
1. Extract storage helpers
2. Fix plaintext secret storage
3. Add transaction support
4. Implement bulk operations

### Phase 3: Handler Consolidation (2 days)
1. Extract token generation
2. Extract OAuth helpers
3. Add missing loggers
4. Fix goroutine contexts

### Phase 4: Testing & Polish (1-2 days)
1. Add comprehensive tests
2. Run golangci-lint
3. Performance testing
4. Documentation

**Total Estimated Effort:** 6-9 days

---

## Testing Requirements

### Unit Tests
- All extracted utility functions
- Token generation
- Encryption/decryption roundtrip
- Constant-time comparison

### Integration Tests
- OAuth flow (Google + GitHub)
- Password reset flow
- MFA enable/disable/verify
- Replay protection
- Rate limiting

### Security Tests
- Timing attack resistance
- User enumeration prevention
- Input validation
- SQL injection prevention

### Performance Tests
- Cache eviction under load
- Concurrent registration handling
- Session revocation bulk operations

---

## Appendix: Error Codes

### Request Signing Errors
| Code | Error | HTTP Status |
|------|-------|-------------|
| SIGN_001 | Missing required headers | 401 |
| SIGN_002 | Invalid timestamp | 401 |
| SIGN_003 | Timestamp outside window | 401 |
| SIGN_004 | Client not found | 401 |
| SIGN_005 | Invalid signature | 401 |
| SIGN_006 | Replay detected | 401 |
| SIGN_007 | Decryption failed | 401 |

### Auth Errors
| Code | Error | HTTP Status |
|------|-------|-------------|
| AUTH_001 | Invalid credentials | 401 |
| AUTH_002 | Account locked | 403 |
| AUTH_003 | MFA required | 403 |
| AUTH_004 | Invalid MFA code | 401 |
| AUTH_005 | Token expired | 400 |
| AUTH_006 | Token already used | 400 |

---

## Document History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-06-17 | Initial comprehensive guide |

---

*This document serves as the single source of truth for all codebase rewrite and refactoring efforts.*
