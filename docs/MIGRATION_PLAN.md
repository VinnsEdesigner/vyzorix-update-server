# Vyzorix Update Server - Complete Migration Plan

**Date:** 2026-06-17  
**Version:** 1.0  
**Total Duration:** 8 Days  
**Status:** Ready to Execute

---

## Table of Contents

1. [Migration Philosophy](#migration-philosophy)
2. [Day 0: Crypto Foundation](#day-0-crypto-foundation)
3. [Day 1: Domain Layer](#day-1-domain-layer)
4. [Day 2: Infrastructure Layer](#day-2-infrastructure-layer)
5. [Day 3-4: Application Layer](#day-34-application-layer)
6. [Day 5: Router & Handlers](#day-5-router--handlers)
7. [Day 6-7: Cleanup & Polish](#day-67-cleanup--polish)
8. [Day 8: Final Verification](#day-8-final-verification)
9. [Rollback Plan](#rollback-plan)

---

## Migration Philosophy

### Rules
1. **Never break the build** - Each step must compile and pass tests
2. **Small commits** - One logical change per commit
3. **Verify at each step** - Run tests before moving on
4. **No big bangs** - Incrementally migrate, don't rewrite everything at once
5. **Keep it working** - External behavior never changes

### Dependency Order
```
CRYPTO (Day 0)
    ↓
DOMAIN (Day 1)
    ↓
INFRASTRUCTURE (Day 2)
    ↓
APPLICATION (Day 3-4)
    ↓
ROUTER + HANDLERS (Day 5)
    ↓
CLEANUP (Day 6-7)
    ↓
VERIFY (Day 8)
```

---

## Day 0: Crypto Foundation

**Goal:** Extract duplicated crypto utilities to shared location  
**Duration:** 0.5 days  
**Risk:** LOW - Self-contained, no business logic changes

### Step 0.1: Create Directory Structure

**Action:**
```bash
mkdir -p /workspace/project/6dc1227fe57247678b4d9fddc4699f3d/apps/api/internal/infrastructure/crypto
```

**Verify:** Directory exists

---

### Step 0.2: Create `aes_gcm.go`

**File:** `internal/infrastructure/crypto/aes_gcm.go`

```go
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
    AES256KeySize = 32
    NonceSize     = 12
)

var (
    ErrCiphertextTooShort   = errors.New("ciphertext too short")
    ErrDecryptionFailed     = errors.New("decryption failed")
    ErrKeyDerivationFailed  = errors.New("key derivation failed")
    ErrEncryptionFailed     = errors.New("encryption failed")
)

// DeriveKey derives a 32-byte key from any secret using SHA-512
func DeriveKey(secret string) []byte {
    key := sha512.Sum512([]byte(secret))
    return key[:AES256KeySize]
}

// EncryptAES256GCM encrypts plaintext using AES-256-GCM
// Returns nonce || ciphertext
func EncryptAES256GCM(secret string, plaintext []byte) ([]byte, error) {
    key := DeriveKey(secret)
    
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, ErrKeyDerivationFailed
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, ErrEncryptionFailed
    }
    
    nonce := make([]byte, NonceSize)
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, ErrEncryptionFailed
    }
    
    return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptAES256GCM decrypts ciphertext using AES-256-GCM
// ciphertext must be nonce || actual_ciphertext
func DecryptAES256GCM(secret string, ciphertext []byte) ([]byte, error) {
    if len(ciphertext) < NonceSize {
        return nil, ErrCiphertextTooShort
    }
    
    key := DeriveKey(secret)
    
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, ErrKeyDerivationFailed
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, ErrDecryptionFailed
    }
    
    nonce := ciphertext[:NonceSize]
    plaintext, err := gcm.Open(nil, nonce, ciphertext[NonceSize:], nil)
    if err != nil {
        return nil, ErrDecryptionFailed
    }
    
    return plaintext, nil
}
```

**Files to Update:** NONE yet - just create this file

**Verify:**
```bash
cd /workspace/project/6dc1227fe57247678b4d9fddc4699f3d/apps/api
export PATH="$PWD/go/bin:$PATH"
go build ./internal/infrastructure/crypto/
```

---

### Step 0.3: Create `aes_gcm_test.go`

**File:** `internal/infrastructure/crypto/aes_gcm_test.go`

```go
package crypto

import (
    "bytes"
    "testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
    secret := "test-secret-key"
    plaintext := []byte("Hello, World!")
    
    ciphertext, err := EncryptAES256GCM(secret, plaintext)
    if err != nil {
        t.Fatalf("Encrypt failed: %v", err)
    }
    
    decrypted, err := DecryptAES256GCM(secret, ciphertext)
    if err != nil {
        t.Fatalf("Decrypt failed: %v", err)
    }
    
    if !bytes.Equal(decrypted, plaintext) {
        t.Errorf("Decrypted != plaintext: got %x, want %x", decrypted, plaintext)
    }
}

func TestDecryptWithWrongKey(t *testing.T) {
    ciphertext, _ := EncryptAES256GCM("secret1", []byte("test"))
    
    _, err := DecryptAES256GCM("secret2", ciphertext)
    if err == nil {
        t.Error("Expected error with wrong key, got nil")
    }
}

func TestDecryptTruncatedCiphertext(t *testing.T) {
    _, err := DecryptAES256GCM("secret", []byte("short"))
    if err != ErrCiphertextTooShort {
        t.Errorf("Expected ErrCiphertextTooShort, got %v", err)
    }
}
```

**Verify:**
```bash
go test ./internal/infrastructure/crypto/ -v
```

---

### Step 0.4: Create `replay_cache.go`

**File:** `internal/infrastructure/crypto/replay_cache.go`

```go
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
    
    if len(rc.entries) >= rc.maxSize {
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

func (rc *ReplayCache) Len() int {
    rc.mu.RLock()
    defer rc.mu.RUnlock()
    return len(rc.entries)
}
```

**Verify:**
```bash
go build ./internal/infrastructure/crypto/
```

---

### Step 0.5: Create `token.go`

**File:** `internal/infrastructure/crypto/token.go`

```go
package crypto

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
)

const TokenSize = 32

// GenerateToken generates a cryptographically secure random token
func GenerateToken() (string, error) {
    b := make([]byte, TokenSize)
    if _, err := rand.Read(b); err != nil {
        return "", fmt.Errorf("failed to generate token: %w", err)
    }
    return hex.EncodeToString(b), nil
}
```

**Verify:**
```bash
go build ./internal/infrastructure/crypto/
```

---

### Step 0.6: Update `request_signing.go` to Use New Crypto

**File:** `internal/api/middleware/request_signing.go`

**Find these functions and update:**

OLD:
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

NEW:
```go
import (
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
)

func (s *RequestSigning) decryptBody(encrypted []byte, secret string) ([]byte, error) {
    return crypto.DecryptAES256GCM(secret, encrypted)
}
```

**Do the same for:**
- `EncryptBody` 
- `SignRequest`

**Add import:**
```go
import (
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
)
```

**Verify:**
```bash
go build ./internal/api/middleware/
```

---

### Step 0.7: Update `response_encryption.go`

**File:** `internal/api/middleware/response_encryption.go`

**Update to use:**
```go
import "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
```

**Replace AES-GCM calls with:**
```go
ciphertext, err := crypto.EncryptAES256GCM(s.secret, plaintext)
```

**Verify:**
```bash
go build ./internal/api/middleware/
```

---

### Step 0.8: Update `replay_protection.go`

**File:** `internal/api/middleware/replay_protection.go`

**Replace ReplayCache struct usage with:**
```go
import "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"

type ReplayProtection struct {
    cache *crypto.ReplayCache
}
```

**Verify:**
```bash
go build ./internal/api/middleware/
```

---

### Day 0 Checklist

- [ ] Created `internal/infrastructure/crypto/aes_gcm.go`
- [ ] Created `internal/infrastructure/crypto/aes_gcm_test.go`
- [ ] Created `internal/infrastructure/crypto/replay_cache.go`
- [ ] Created `internal/infrastructure/crypto/token.go`
- [ ] Updated `request_signing.go` to use new crypto
- [ ] Updated `response_encryption.go` to use new crypto
- [ ] Updated `replay_protection.go` to use new replay cache
- [ ] All tests pass
- [ ] Build succeeds

---

## Day 1: Domain Layer

**Goal:** Create domain entities and repository interfaces  
**Duration:** 1 day  
**Risk:** LOW - Interface definitions only, no implementations

### Step 1.1: Create Domain Directory Structure

**Action:**
```bash
mkdir -p /workspace/project/6dc1227fe57247678b4d9fddc4699f3d/apps/api/internal/domain/{operator,device,session,command,auth}
```

---

### Step 1.2: Create Domain Errors

**File:** `internal/domain/errors.go`

```go
package domain

import "errors"

var (
    ErrNotFound          = errors.New("entity not found")
    ErrAlreadyExists     = errors.New("entity already exists")
    ErrInvalidInput      = errors.New("invalid input")
    ErrUnauthorized      = errors.New("unauthorized")
    ErrForbidden         = errors.New("forbidden")
    ErrInternal          = errors.New("internal error")
    ErrNotImplemented    = errors.New("not implemented")
)
```

**Verify:**
```bash
go build ./internal/domain/
```

---

### Step 1.3: Create Operator Entity

**File:** `internal/domain/operator/entity.go`

```go
package operator

import "time"

type Role string

const (
    RoleSuperAdmin Role = "super_admin"
    RoleAdmin      Role = "admin"
    RoleOperator   Role = "operator"
)

type Operator struct {
    ID           string
    Email        string
    Name         string
    PasswordHash string
    Role         Role
    
    // OAuth fields (optional)
    GoogleID string
    GitHubID string
    
    // MFA fields
    MFASecret   string
    MFAEnabled  bool
    
    // Timestamps
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (o *Operator) IsSuperAdmin() bool {
    return o.Role == RoleSuperAdmin
}

func (o *Operator) IsAdmin() bool {
    return o.Role == RoleSuperAdmin || o.Role == RoleAdmin
}

func (o *Operator) CanManageOperators() bool {
    return o.IsAdmin()
}
```

**Verify:**
```bash
go build ./internal/domain/operator/
```

---

### Step 1.4: Create Operator Repository Interface

**File:** `internal/domain/operator/repository.go`

```go
package operator

import (
    "context"
)

type Repository interface {
    // Queries
    FindByID(ctx context.Context, id string) (*Operator, error)
    FindByEmail(ctx context.Context, email string) (*Operator, error)
    FindByGoogleID(ctx context.Context, googleID string) (*Operator, error)
    FindByGitHubID(ctx context.Context, githubID string) (*Operator, error)
    
    // Mutations
    Create(ctx context.Context, op *Operator) error
    Update(ctx context.Context, op *Operator) error
    Delete(ctx context.Context, id string) error
    
    // Counting
    Count(ctx context.Context) (int, error)
}
```

**Verify:**
```bash
go build ./internal/domain/operator/
```

---

### Step 1.5: Create Device Entity

**File:** `internal/domain/device/entity.go`

```go
package device

import "time"

type Device struct {
    ID                string
    FirebaseInstallID string
    FCMToken          string
    AppVersion        string
    DeviceClass       string
    
    // Command signing
    CommandSecretHash string
    
    // Status
    Online       bool
    RegisteredAt int64
    LastSeen     int64
    
    // Timestamps
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Verify:**
```bash
go build ./internal/domain/device/
```

---

### Step 1.6: Create Device Repository Interface

**File:** `internal/domain/device/repository.go`

```go
package device

import (
    "context"
)

type Repository interface {
    FindByID(ctx context.Context, id string) (*Device, error)
    FindByFirebaseInstallID(ctx context.Context, fid string) (*Device, error)
    
    Create(ctx context.Context, d *Device) error
    Update(ctx context.Context, d *Device) error
    UpdateFCMToken(ctx context.Context, id, fcmToken string) error
    SetOnline(ctx context.Context, id string, online bool) error
    
    List(ctx context.Context, limit, offset int) ([]*Device, int, error)
    Delete(ctx context.Context, id string) error
}
```

**Verify:**
```bash
go build ./internal/domain/device/
```

---

### Step 1.7: Create Session Entity

**File:** `internal/domain/session/entity.go`

```go
package session

import "time"

type Session struct {
    ID         string
    OperatorID string
    ExpiresAt  time.Time
    CreatedAt  time.Time
}

func (s *Session) IsExpired() bool {
    return time.Now().After(s.ExpiresAt)
}
```

**Verify:**
```bash
go build ./internal/domain/session/
```

---

### Step 1.8: Create Session Repository Interface

**File:** `internal/domain/session/repository.go`

```go
package session

import (
    "context"
)

type Repository interface {
    FindByID(ctx context.Context, id string) (*Session, error)
    FindByOperatorID(ctx context.Context, operatorID string) ([]*Session, error)
    
    Create(ctx context.Context, s *Session) error
    Delete(ctx context.Context, id string) error
    DeleteByOperatorID(ctx context.Context, operatorID string) error
    
    DeleteExpired(ctx context.Context) (int, error)
}
```

**Verify:**
```bash
go build ./internal/domain/session/
```

---

### Step 1.9: Create Command Entity

**File:** `internal/domain/command/entity.go`

```go
package command

import "time"

type Status string

const (
    StatusPending    Status = "pending"
    StatusDelivered  Status = "delivered"
    StatusCompleted Status = "completed"
    StatusFailed   Status = "failed"
)

type Command struct {
    ID          string
    DeviceID    string
    DispatchID  string
    Command     string
    Args        []byte
    
    Status      Status
    
    DeliveredAt *int64
    CompletedAt *int64
    
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Verify:**
```bash
go build ./internal/domain/command/
```

---

### Step 1.10: Create Command Repository Interface

**File:** `internal/domain/command/repository.go`

```go
package command

import (
    "context"
)

type Repository interface {
    FindByID(ctx context.Context, id string) (*Command, error)
    FindByDispatchID(ctx context.Context, dispatchID string) (*Command, error)
    FindByDeviceID(ctx context.Context, deviceID string, limit int) ([]*Command, error)
    
    Create(ctx context.Context, c *Command) error
    UpdateStatus(ctx context.Context, id string, status Status) error
    
    Delete(ctx context.Context, id string) error
}
```

**Verify:**
```bash
go build ./internal/domain/command/
```

---

### Step 1.11: Create Auth Domain Utilities

**File:** `internal/domain/auth/enum_safe.go`

```go
package auth

import (
    "crypto/subtle"
)

func ConstantTimeCompare(a, b string) bool {
    return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
```

**Verify:**
```bash
go build ./internal/domain/auth/
```

---

### Day 1 Checklist

- [ ] Created `internal/domain/errors.go`
- [ ] Created `internal/domain/operator/entity.go`
- [ ] Created `internal/domain/operator/repository.go`
- [ ] Created `internal/domain/device/entity.go`
- [ ] Created `internal/domain/device/repository.go`
- [ ] Created `internal/domain/session/entity.go`
- [ ] Created `internal/domain/session/repository.go`
- [ ] Created `internal/domain/command/entity.go`
- [ ] Created `internal/domain/command/repository.go`
- [ ] Created `internal/domain/auth/enum_safe.go`
- [ ] All domain packages build successfully
- [ ] All tests pass

---

## Day 2: Infrastructure Layer

**Goal:** Implement domain interfaces with storage and crypto  
**Duration:** 1 day  
**Risk:** MEDIUM - Implementation changes, but backward compatible

### Step 2.1: Create Infrastructure Directory Structure

```bash
mkdir -p /workspace/project/6dc1227fe57247678b4d9fddc4699f3d/apps/api/internal/infrastructure/{storage,auth}
```

---

### Step 2.2: Create SQLite Storage Base

**File:** `internal/infrastructure/storage/sqlite.go`

```go
package storage

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
    DB *sql.DB
}

func Open(path string) (*SQLite, error) {
    db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_cache_size=-2000&_busy_timeout=5000&_foreign_keys=1")
    if err != nil {
        return nil, err
    }
    
    if err := db.Ping(); err != nil {
        return nil, err
    }
    
    return &SQLite{DB: db}, nil
}

func (s *SQLite) Close() error {
    return s.DB.Close()
}
```

**Verify:**
```bash
go build ./internal/infrastructure/storage/
```

---

### Step 2.3: Create Operator Repository Implementation

**File:** `internal/infrastructure/storage/operator.go`

```go
package storage

import (
    "context"
    "database/sql"
    "errors"
    
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
)

var _ operator.Repository = (*OperatorRepository)(nil)

type OperatorRepository struct {
    db *sql.DB
}

func NewOperatorRepository(db *sql.DB) *OperatorRepository {
    return &OperatorRepository{db: db}
}

func (r *OperatorRepository) FindByID(ctx context.Context, id string) (*operator.Operator, error) {
    query := `SELECT id, email, name, password_hash, role, google_id, github_id, 
              mfa_secret, mfa_enabled, created_at, updated_at 
              FROM operators WHERE id = ?`
    
    var op operator.Operator
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &op.ID, &op.Email, &op.Name, &op.PasswordHash, &op.Role,
        &op.GoogleID, &op.GitHubID, &op.MFASecret, &op.MFAEnabled,
        &op.CreatedAt, &op.UpdatedAt,
    )
    if errors.Is(err, sql.ErrNoRows) {
        return nil, operator.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    return &op, nil
}

func (r *OperatorRepository) FindByEmail(ctx context.Context, email string) (*operator.Operator, error) {
    query := `SELECT id, email, name, password_hash, role, google_id, github_id, 
              mfa_secret, mfa_enabled, created_at, updated_at 
              FROM operators WHERE email = ?`
    
    var op operator.Operator
    err := r.db.QueryRowContext(ctx, query, email).Scan(
        &op.ID, &op.Email, &op.Name, &op.PasswordHash, &op.Role,
        &op.GoogleID, &op.GitHubID, &op.MFASecret, &op.MFAEnabled,
        &op.CreatedAt, &op.UpdatedAt,
    )
    if errors.Is(err, sql.ErrNoRows) {
        return nil, operator.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    return &op, nil
}

func (r *OperatorRepository) FindByGoogleID(ctx context.Context, googleID string) (*operator.Operator, error) {
    query := `SELECT id, email, name, password_hash, role, google_id, github_id, 
              mfa_secret, mfa_enabled, created_at, updated_at 
              FROM operators WHERE google_id = ?`
    
    var op operator.Operator
    err := r.db.QueryRowContext(ctx, query, googleID).Scan(
        &op.ID, &op.Email, &op.Name, &op.PasswordHash, &op.Role,
        &op.GoogleID, &op.GitHubID, &op.MFASecret, &op.MFAEnabled,
        &op.CreatedAt, &op.UpdatedAt,
    )
    if errors.Is(err, sql.ErrNoRows) {
        return nil, operator.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    return &op, nil
}

func (r *OperatorRepository) FindByGitHubID(ctx context.Context, githubID string) (*operator.Operator, error) {
    query := `SELECT id, email, name, password_hash, role, google_id, github_id, 
              mfa_secret, mfa_enabled, created_at, updated_at 
              FROM operators WHERE github_id = ?`
    
    var op operator.Operator
    err := r.db.QueryRowContext(ctx, query, githubID).Scan(
        &op.ID, &op.Email, &op.Name, &op.PasswordHash, &op.Role,
        &op.GoogleID, &op.GitHubID, &op.MFASecret, &op.MFAEnabled,
        &op.CreatedAt, &op.UpdatedAt,
    )
    if errors.Is(err, sql.ErrNoRows) {
        return nil, operator.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    return &op, nil
}

func (r *OperatorRepository) Create(ctx context.Context, op *operator.Operator) error {
    query := `INSERT INTO operators (id, email, name, password_hash, role, google_id, github_id, 
              mfa_secret, mfa_enabled, created_at, updated_at) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
    
    _, err := r.db.ExecContext(ctx, query,
        op.ID, op.Email, op.Name, op.PasswordHash, op.Role,
        op.GoogleID, op.GitHubID, op.MFASecret, op.MFAEnabled,
        op.CreatedAt, op.UpdatedAt,
    )
    return err
}

func (r *OperatorRepository) Update(ctx context.Context, op *operator.Operator) error {
    query := `UPDATE operators SET email = ?, name = ?, password_hash = ?, role = ?,
              google_id = ?, github_id = ?, mfa_secret = ?, mfa_enabled = ?, updated_at = ?
              WHERE id = ?`
    
    _, err := r.db.ExecContext(ctx, query,
        op.Email, op.Name, op.PasswordHash, op.Role,
        op.GoogleID, op.GitHubID, op.MFASecret, op.MFAEnabled,
        op.UpdatedAt, op.ID,
    )
    return err
}

func (r *OperatorRepository) Delete(ctx context.Context, id string) error {
    _, err := r.db.ExecContext(ctx, "DELETE FROM operators WHERE id = ?", id)
    return err
}

func (r *OperatorRepository) Count(ctx context.Context) (int, error) {
    var count int
    err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM operators").Scan(&count)
    return count, err
}
```

**Verify:**
```bash
go build ./internal/infrastructure/storage/
```

---

### Step 2.4: Create Device Repository Implementation

**File:** `internal/infrastructure/storage/device.go`

```go
package storage

import (
    "context"
    "database/sql"
    "errors"
    
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
)

var _ device.Repository = (*DeviceRepository)(nil)

type DeviceRepository struct {
    db *sql.DB
}

func NewDeviceRepository(db *sql.DB) *DeviceRepository {
    return &DeviceRepository{db: db}
}

func (r *DeviceRepository) FindByID(ctx context.Context, id string) (*device.Device, error) {
    query := `SELECT id, firebase_install_id, fcm_token, app_version, device_class,
              command_secret_hash, online, registered_at, last_seen, created_at, updated_at
              FROM devices WHERE id = ?`
    
    var d device.Device
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &d.ID, &d.FirebaseInstallID, &d.FCMToken, &d.AppVersion, &d.DeviceClass,
        &d.CommandSecretHash, &d.Online, &d.RegisteredAt, &d.LastSeen,
        &d.CreatedAt, &d.UpdatedAt,
    )
    if errors.Is(err, sql.ErrNoRows) {
        return nil, device.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    return &d, nil
}

func (r *DeviceRepository) FindByFirebaseInstallID(ctx context.Context, fid string) (*device.Device, error) {
    query := `SELECT id, firebase_install_id, fcm_token, app_version, device_class,
              command_secret_hash, online, registered_at, last_seen, created_at, updated_at
              FROM devices WHERE firebase_install_id = ?`
    
    var d device.Device
    err := r.db.QueryRowContext(ctx, query, fid).Scan(
        &d.ID, &d.FirebaseInstallID, &d.FCMToken, &d.AppVersion, &d.DeviceClass,
        &d.CommandSecretHash, &d.Online, &d.RegisteredAt, &d.LastSeen,
        &d.CreatedAt, &d.UpdatedAt,
    )
    if errors.Is(err, sql.ErrNoRows) {
        return nil, device.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    return &d, nil
}

func (r *DeviceRepository) Create(ctx context.Context, d *device.Device) error {
    query := `INSERT INTO devices (id, firebase_install_id, fcm_token, app_version, device_class,
              command_secret_hash, online, registered_at, last_seen, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
    
    _, err := r.db.ExecContext(ctx, query,
        d.ID, d.FirebaseInstallID, d.FCMToken, d.AppVersion, d.DeviceClass,
        d.CommandSecretHash, d.Online, d.RegisteredAt, d.LastSeen,
        d.CreatedAt, d.UpdatedAt,
    )
    return err
}

func (r *DeviceRepository) Update(ctx context.Context, d *device.Device) error {
    query := `UPDATE devices SET fcm_token = ?, app_version = ?, device_class = ?,
              command_secret_hash = ?, online = ?, registered_at = ?, last_seen = ?, updated_at = ?
              WHERE id = ?`
    
    _, err := r.db.ExecContext(ctx, query,
        d.FCMToken, d.AppVersion, d.DeviceClass,
        d.CommandSecretHash, d.Online, d.RegisteredAt, d.LastSeen,
        d.UpdatedAt, d.ID,
    )
    return err
}

func (r *DeviceRepository) UpdateFCMToken(ctx context.Context, id, fcmToken string) error {
    _, err := r.db.ExecContext(ctx,
        "UPDATE devices SET fcm_token = ?, updated_at = ? WHERE id = ?",
        fcmToken, time.Now(), id,
    )
    return err
}

func (r *DeviceRepository) SetOnline(ctx context.Context, id string, online bool) error {
    _, err := r.db.ExecContext(ctx,
        "UPDATE devices SET online = ?, last_seen = ?, updated_at = ? WHERE id = ?",
        online, time.Now().UnixMilli(), time.Now(), id,
    )
    return err
}

func (r *DeviceRepository) List(ctx context.Context, limit, offset int) ([]*device.Device, int, error) {
    query := `SELECT id, firebase_install_id, fcm_token, app_version, device_class,
              command_secret_hash, online, registered_at, last_seen, created_at, updated_at
              FROM devices ORDER BY created_at DESC LIMIT ? OFFSET ?`
    
    rows, err := r.db.QueryContext(ctx, query, limit, offset)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()
    
    var devices []*device.Device
    for rows.Next() {
        var d device.Device
        if err := rows.Scan(
            &d.ID, &d.FirebaseInstallID, &d.FCMToken, &d.AppVersion, &d.DeviceClass,
            &d.CommandSecretHash, &d.Online, &d.RegisteredAt, &d.LastSeen,
            &d.CreatedAt, &d.UpdatedAt,
        ); err != nil {
            return nil, 0, err
        }
        devices = append(devices, &d)
    }
    
    var total int
    r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM devices").Scan(&total)
    
    return devices, total, nil
}

func (r *DeviceRepository) Delete(ctx context.Context, id string) error {
    _, err := r.db.ExecContext(ctx, "DELETE FROM devices WHERE id = ?", id)
    return err
}
```

**Verify:**
```bash
go build ./internal/infrastructure/storage/
```

---

### Step 2.5: Create Session Repository Implementation

**File:** `internal/infrastructure/storage/session.go`

```go
package storage

import (
    "context"
    "database/sql"
    "errors"
    
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"
)

var _ session.Repository = (*SessionRepository)(nil)

type SessionRepository struct {
    db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
    return &SessionRepository{db: db}
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (*session.Session, error) {
    query := `SELECT id, operator_id, expires_at, created_at FROM auth_sessions WHERE id = ?`
    
    var s session.Session
    err := r.db.QueryRowContext(ctx, query, id).Scan(&s.ID, &s.OperatorID, &s.ExpiresAt, &s.CreatedAt)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, session.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    return &s, nil
}

func (r *SessionRepository) FindByOperatorID(ctx context.Context, operatorID string) ([]*session.Session, error) {
    query := `SELECT id, operator_id, expires_at, created_at FROM auth_sessions WHERE operator_id = ?`
    
    rows, err := r.db.QueryContext(ctx, query, operatorID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var sessions []*session.Session
    for rows.Next() {
        var s session.Session
        if err := rows.Scan(&s.ID, &s.OperatorID, &s.ExpiresAt, &s.CreatedAt); err != nil {
            return nil, err
        }
        sessions = append(sessions, &s)
    }
    return sessions, nil
}

func (r *SessionRepository) Create(ctx context.Context, s *session.Session) error {
    query := `INSERT INTO auth_sessions (id, operator_id, expires_at, created_at) VALUES (?, ?, ?, ?)`
    _, err := r.db.ExecContext(ctx, query, s.ID, s.OperatorID, s.ExpiresAt, s.CreatedAt)
    return err
}

func (r *SessionRepository) Delete(ctx context.Context, id string) error {
    _, err := r.db.ExecContext(ctx, "DELETE FROM auth_sessions WHERE id = ?", id)
    return err
}

func (r *SessionRepository) DeleteByOperatorID(ctx context.Context, operatorID string) error {
    _, err := r.db.ExecContext(ctx, "DELETE FROM auth_sessions WHERE operator_id = ?", operatorID)
    return err
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) (int, error) {
    result, err := r.db.ExecContext(ctx, "DELETE FROM auth_sessions WHERE expires_at < ?", time.Now())
    if err != nil {
        return 0, err
    }
    return result.RowsAffected()
}
```

**Verify:**
```bash
go build ./internal/infrastructure/storage/
```

---

### Step 2.6: Create Command Repository Implementation

**File:** `internal/infrastructure/storage/command.go`

```go
package storage

import (
    "context"
    "database/sql"
    "encoding/json"
    "errors"
    
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
)

var _ command.Repository = (*CommandRepository)(nil)

type CommandRepository struct {
    db *sql.DB
}

func NewCommandRepository(db *sql.DB) *CommandRepository {
    return &CommandRepository{db: db}
}

func (r *CommandRepository) FindByID(ctx context.Context, id string) (*command.Command, error) {
    query := `SELECT id, device_id, dispatch_id, command, args, status, delivered_at, completed_at,
              created_at, updated_at FROM commands WHERE id = ?`
    
    var c command.Command
    var argsJSON []byte
    var deliveredAt, completedAt sql.NullInt64
    
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &c.ID, &c.DeviceID, &c.DispatchID, &c.Command, &argsJSON, &c.Status,
        &deliveredAt, &completedAt, &c.CreatedAt, &c.UpdatedAt,
    )
    if errors.Is(err, sql.ErrNoRows) {
        return nil, command.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    
    if argsJSON != nil {
        json.Unmarshal(argsJSON, &c.Args)
    }
    if deliveredAt.Valid {
        c.DeliveredAt = &deliveredAt.Int64
    }
    if completedAt.Valid {
        c.CompletedAt = &completedAt.Int64
    }
    
    return &c, nil
}

func (r *CommandRepository) FindByDispatchID(ctx context.Context, dispatchID string) (*command.Command, error) {
    query := `SELECT id, device_id, dispatch_id, command, args, status, delivered_at, completed_at,
              created_at, updated_at FROM commands WHERE dispatch_id = ?`
    
    var c command.Command
    var argsJSON []byte
    var deliveredAt, completedAt sql.NullInt64
    
    err := r.db.QueryRowContext(ctx, query, dispatchID).Scan(
        &c.ID, &c.DeviceID, &c.DispatchID, &c.Command, &argsJSON, &c.Status,
        &deliveredAt, &completedAt, &c.CreatedAt, &c.UpdatedAt,
    )
    if errors.Is(err, sql.ErrNoRows) {
        return nil, command.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    
    if argsJSON != nil {
        json.Unmarshal(argsJSON, &c.Args)
    }
    if deliveredAt.Valid {
        c.DeliveredAt = &deliveredAt.Int64
    }
    if completedAt.Valid {
        c.CompletedAt = &completedAt.Int64
    }
    
    return &c, nil
}

func (r *CommandRepository) FindByDeviceID(ctx context.Context, deviceID string, limit int) ([]*command.Command, error) {
    query := `SELECT id, device_id, dispatch_id, command, args, status, delivered_at, completed_at,
              created_at, updated_at FROM commands WHERE device_id = ? ORDER BY created_at DESC LIMIT ?`
    
    rows, err := r.db.QueryContext(ctx, query, deviceID, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var commands []*command.Command
    for rows.Next() {
        var c command.Command
        var argsJSON []byte
        var deliveredAt, completedAt sql.NullInt64
        
        if err := rows.Scan(
            &c.ID, &c.DeviceID, &c.DispatchID, &c.Command, &argsJSON, &c.Status,
            &deliveredAt, &completedAt, &c.CreatedAt, &c.UpdatedAt,
        ); err != nil {
            return nil, err
        }
        
        if argsJSON != nil {
            json.Unmarshal(argsJSON, &c.Args)
        }
        if deliveredAt.Valid {
            c.DeliveredAt = &deliveredAt.Int64
        }
        if completedAt.Valid {
            c.CompletedAt = &completedAt.Int64
        }
        
        commands = append(commands, &c)
    }
    return commands, nil
}

func (r *CommandRepository) Create(ctx context.Context, c *command.Command) error {
    argsJSON, _ := json.Marshal(c.Args)
    
    query := `INSERT INTO commands (id, device_id, dispatch_id, command, args, status, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
    
    _, err := r.db.ExecContext(ctx, query,
        c.ID, c.DeviceID, c.DispatchID, c.Command, argsJSON, c.Status, c.CreatedAt, c.UpdatedAt,
    )
    return err
}

func (r *CommandRepository) UpdateStatus(ctx context.Context, id string, status command.Status) error {
    var query string
    var args []interface{}
    
    switch status {
    case command.StatusDelivered:
        query = "UPDATE commands SET status = ?, delivered_at = ?, updated_at = ? WHERE id = ?"
        args = []interface{}{status, time.Now().UnixMilli(), time.Now(), id}
    case command.StatusCompleted:
        query = "UPDATE commands SET status = ?, completed_at = ?, updated_at = ? WHERE id = ?"
        args = []interface{}{status, time.Now().UnixMilli(), time.Now(), id}
    default:
        query = "UPDATE commands SET status = ?, updated_at = ? WHERE id = ?"
        args = []interface{}{status, time.Now(), id}
    }
    
    _, err := r.db.ExecContext(ctx, query, args...)
    return err
}

func (r *CommandRepository) Delete(ctx context.Context, id string) error {
    _, err := r.db.ExecContext(ctx, "DELETE FROM commands WHERE id = ?", id)
    return err
}
```

**Verify:**
```bash
go build ./internal/infrastructure/storage/
```

---

### Day 2 Checklist

- [ ] Created `internal/infrastructure/storage/sqlite.go`
- [ ] Created `internal/infrastructure/storage/operator.go`
- [ ] Created `internal/infrastructure/storage/device.go`
- [ ] Created `internal/infrastructure/storage/session.go`
- [ ] Created `internal/infrastructure/storage/command.go`
- [ ] All storage implementations satisfy domain interfaces
- [ ] Build succeeds
- [ ] Tests pass

---

## Day 3-4: Application Layer

**Goal:** Create use cases that orchestrate domain logic  
**Duration:** 2 days  
**Risk:** MEDIUM - Business logic extraction

### Step 3.1: Create Application Directory Structure

```bash
mkdir -p /workspace/project/6dc1227fe57247678b4d9fddc4699f3d/apps/api/internal/application/{auth,device,shared}
mkdir -p /workspace/project/6dc1227fe57247678b4d9fddc4699f3d/apps/api/internal/application/dto
```

---

### Step 3.2: Create Application Errors

**File:** `internal/application/errors.go`

```go
package application

import "errors"

var (
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrAccountLocked      = errors.New("account locked")
    ErrMFARequired       = errors.New("mfa required")
    ErrInvalidMFACode     = errors.New("invalid mfa code")
    ErrTokenExpired       = errors.New("token expired")
    ErrTokenUsed          = errors.New("token already used")
    ErrEmailNotVerified    = errors.New("email not verified")
    ErrUserExists         = errors.New("user already exists")
)
```

---

### Step 3.3: Create Auth DTOs

**File:** `internal/application/dto/auth.go`

```go
package dto

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginResponse struct {
    OperatorID string `json:"operator_id"`
    Email      string `json:"email"`
    Name       string `json:"name"`
    Role       string `json:"role"`
}

type RegisterRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    Name     string `json:"name"`
}

type RegisterResponse struct {
    OperatorID string `json:"operator_id"`
    Email      string `json:"email"`
    Name       string `json:"name"`
}
```

---

### Step 3.4: Create Password Hasher Interface

**File:** `internal/application/auth/password.go`

```go
package auth

import "github.com/VinnsEdesigner/vyzorix/apps/api/pkg/crypto"

type PasswordHasher interface {
    Hash(password string) (string, error)
    Verify(password, hash string) error
}

type Argon2PasswordHasher struct{}

func (h *Argon2PasswordHasher) Hash(password string) (string, error) {
    return crypto.HashPassword(password)
}

func (h *Argon2PasswordHasher) Verify(password, hash string) error {
    return crypto.VerifyPassword(password, hash)
}
```

---

### Step 3.5: Create Auth Service

**File:** `internal/application/auth/service.go`

```go
package auth

import (
    "context"
    "time"
    
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
)

type AuthService struct {
    operatorRepo operator.Repository
    sessionRepo  session.Repository
    hasher       PasswordHasher
}

func NewAuthService(
    operatorRepo operator.Repository,
    sessionRepo session.Repository,
    hasher PasswordHasher,
) *AuthService {
    return &AuthService{
        operatorRepo: operatorRepo,
        sessionRepo:  sessionRepo,
        hasher:       hasher,
    }
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
    op, err := s.operatorRepo.FindByEmail(ctx, req.Email)
    if err != nil {
        if err == operator.ErrNotFound {
            // Deliberately run password hash to prevent timing attacks
            s.hasher.Verify(req.Password, "$argon2id$placeholder")
            return nil, application.ErrInvalidCredentials
        }
        return nil, err
    }
    
    if err := s.hasher.Verify(req.Password, op.PasswordHash); err != nil {
        return nil, application.ErrInvalidCredentials
    }
    
    return &dto.LoginResponse{
        OperatorID: op.ID,
        Email:      op.Email,
        Name:       op.Name,
        Role:       string(op.Role),
    }, nil
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
    // Check if user exists
    existing, _ := s.operatorRepo.FindByEmail(ctx, req.Email)
    if existing != nil {
        return nil, application.ErrUserExists
    }
    
    // Hash password
    hash, err := s.hasher.Hash(req.Password)
    if err != nil {
        return nil, err
    }
    
    // Determine role (first operator is super admin)
    count, _ := s.operatorRepo.Count(ctx)
    role := operator.RoleOperator
    if count == 0 {
        role = operator.RoleSuperAdmin
    }
    
    // Generate ID
    id, _ := crypto.GenerateToken()
    
    now := time.Now()
    op := &operator.Operator{
        ID:           id,
        Email:        req.Email,
        Name:         req.Name,
        PasswordHash: hash,
        Role:         role,
        CreatedAt:    now,
        UpdatedAt:    now,
    }
    
    if err := s.operatorRepo.Create(ctx, op); err != nil {
        return nil, err
    }
    
    return &dto.RegisterResponse{
        OperatorID: op.ID,
        Email:      op.Email,
        Name:       op.Name,
    }, nil
}

func (s *AuthService) CreateSession(ctx context.Context, operatorID string, maxAge time.Duration) (*session.Session, error) {
    id, err := crypto.GenerateToken()
    if err != nil {
        return nil, err
    }
    
    sess := &session.Session{
        ID:         id,
        OperatorID: operatorID,
        ExpiresAt:  time.Now().Add(maxAge),
        CreatedAt:  time.Now(),
    }
    
    if err := s.sessionRepo.Create(ctx, sess); err != nil {
        return nil, err
    }
    
    return sess, nil
}
```

**Verify:**
```bash
go build ./internal/application/auth/
```

---

### Step 3.6: Create Device DTOs

**File:** `internal/application/dto/device.go`

```go
package dto

type RegisterDeviceRequest struct {
    DeviceID          string `json:"device_id"`
    FirebaseInstallID string `json:"firebase_install_id"`
    FCMToken         string `json:"fcm_token"`
    AppVersion        string `json:"app_version"`
    DeviceClass       string `json:"device_class"`
}

type RegisterDeviceResponse struct {
    DeviceID       string `json:"device_id"`
    CommandSecret  string `json:"command_secret"`
    RegisteredAt   int64  `json:"registered_at"`
}
```

---

### Step 3.7: Create Device Service

**File:** `internal/application/device/service.go`

```go
package device

import (
    "context"
    "time"
    
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"
)

type DeviceService struct {
    deviceRepo device.Repository
}

func NewDeviceService(deviceRepo device.Repository) *DeviceService {
    return &DeviceService{deviceRepo: deviceRepo}
}

func (s *DeviceService) Register(ctx context.Context, req *dto.RegisterDeviceRequest) (*dto.RegisterDeviceResponse, error) {
    // Generate command secret
    secret, err := crypto.GenerateToken()
    if err != nil {
        return nil, err
    }
    
    now := time.Now()
    d := &device.Device{
        ID:                req.DeviceID,
        FirebaseInstallID: req.FirebaseInstallID,
        FCMToken:         req.FCMToken,
        AppVersion:        req.AppVersion,
        DeviceClass:       req.DeviceClass,
        Online:           true,
        RegisteredAt:      now.UnixMilli(),
        LastSeen:         now.UnixMilli(),
        CreatedAt:        now,
        UpdatedAt:         now,
    }
    
    if err := s.deviceRepo.Create(ctx, d); err != nil {
        return nil, err
    }
    
    return &dto.RegisterDeviceResponse{
        DeviceID:      d.ID,
        CommandSecret: secret, // Return to client - they must store it securely
        RegisteredAt:  d.RegisteredAt,
    }, nil
}

func (s *DeviceService) GetStatus(ctx context.Context, deviceID string) (*device.Device, error) {
    return s.deviceRepo.FindByID(ctx, deviceID)
}
```

**Verify:**
```bash
go build ./internal/application/device/
```

---

### Day 3-4 Checklist

- [ ] Created `internal/application/errors.go`
- [ ] Created `internal/application/dto/auth.go`
- [ ] Created `internal/application/dto/device.go`
- [ ] Created `internal/application/auth/password.go`
- [ ] Created `internal/application/auth/service.go`
- [ ] Created `internal/application/device/service.go`
- [ ] All application packages build successfully
- [ ] Tests pass

---

## Day 5: Router & Handlers

**Goal:** Refactor handlers to use application layer  
**Duration:** 1 day  
**Risk:** MEDIUM - Handler changes

### Step 5.1: Create Router

**File:** `internal/api/router.go`

```go
package api

import (
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
    
    "github.com/gin-gonic/gin"
)

type Router struct {
    engine         *gin.Engine
    authService    *auth.AuthService
    deviceService  *device.DeviceService
    operatorRepo   operator.Repository
    deviceRepo     device.Repository
}

func NewRouter(
    db *sql.DB,
) *Router {
    // Create repositories
    operatorRepo := storage.NewOperatorRepository(db)
    deviceRepo := storage.NewDeviceRepository(db)
    sessionRepo := storage.NewSessionRepository(db)
    
    // Create hasher
    hasher := &auth.Argon2PasswordHasher{}
    
    // Create services
    authService := auth.NewAuthService(operatorRepo, sessionRepo, hasher)
    deviceService := device.NewDeviceService(deviceRepo)
    
    // Create handlers
    authHandler := NewAuthHandler(authService)
    deviceHandler := NewDeviceHandler(deviceService)
    
    // Create router
    r := &Router{
        engine:        gin.Default(),
        authService:   authService,
        deviceService: deviceService,
        operatorRepo:  operatorRepo,
        deviceRepo:    deviceRepo,
    }
    
    // Setup routes
    r.setupRoutes(authHandler, deviceHandler)
    
    return r
}

func (r *Router) setupRoutes(authHandler *AuthHandler, deviceHandler *DeviceHandler) {
    // Public routes
    r.engine.POST("/v1/auth/login", authHandler.Login)
    r.engine.POST("/v1/auth/register", authHandler.Register)
    r.engine.POST("/v1/device/register", deviceHandler.Register)
    
    // Health
    r.engine.GET("/health", healthHandler)
}

func (r *Router) Engine() *gin.Engine {
    return r.engine
}
```

---

### Step 5.2: Create Auth Handler

**File:** `internal/api/auth_handler.go`

```go
package api

import (
    "net/http"
    
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
    
    "github.com/gin-gonic/gin"
)

type AuthHandler struct {
    authService *auth.AuthService
}

func NewAuthHandler(authService *auth.AuthService) *AuthHandler {
    return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req dto.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
        return
    }
    
    result, err := h.authService.Login(c.Request.Context(), &req)
    if err != nil {
        switch err {
        case auth.ErrInvalidCredentials:
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
        }
        return
    }
    
    c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) Register(c *gin.Context) {
    var req dto.RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
        return
    }
    
    result, err := h.authService.Register(c.Request.Context(), &req)
    if err != nil {
        switch err {
        case auth.ErrUserExists:
            c.JSON(http.StatusConflict, gin.H{"error": "user_exists"})
        default:
            c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
        }
        return
    }
    
    c.JSON(http.StatusCreated, result)
}
```

---

### Step 5.3: Create Device Handler

**File:** `internal/api/device_handler.go`

```go
package api

import (
    "net/http"
    
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
    
    "github.com/gin-gonic/gin"
)

type DeviceHandler struct {
    deviceService *device.DeviceService
}

func NewDeviceHandler(deviceService *device.DeviceService) *DeviceHandler {
    return &DeviceHandler{deviceService: deviceService}
}

func (h *DeviceHandler) Register(c *gin.Context) {
    var req dto.RegisterDeviceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
        return
    }
    
    result, err := h.deviceService.Register(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
        return
    }
    
    c.JSON(http.StatusCreated, result)
}
```

---

### Step 5.4: Create Health Handler

**File:** `internal/api/health_handler.go`

```go
package api

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
)

func healthHandler(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "healthy",
    })
}
```

---

### Day 5 Checklist

- [ ] Created `internal/api/router.go`
- [ ] Created `internal/api/auth_handler.go`
- [ ] Created `internal/api/device_handler.go`
- [ ] Created `internal/api/health_handler.go`
- [ ] Build succeeds
- [ ] Basic endpoints work

---

## Day 6-7: Cleanup & Polish

**Goal:** Remove duplicated code, finalize structure  
**Duration:** 2 days  
**Risk:** LOW - Cleanup only

### Step 6.1: Delete Orphaned Files

```bash
# Delete duplicated files that are no longer needed
rm -f internal/api/handlers/auth_csrf.go
rm -f internal/api/handlers/auth_rate_limit.go
rm -f internal/api/handlers/lockout.go
rm -f internal/api/handlers/lockout_test.go
rm -f internal/api/handlers/rate_limit_test.go
```

---

### Step 6.2: Consolidate Constant-Time Compare

**Update `internal/domain/auth/enum_safe.go`:**

```go
package auth

import (
    "crypto/subtle"
    "time"
    
    "golang.org/x/crypto/argon2"
)

func ConstantTimeCompare(a, b string) bool {
    return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// ComputeFakePasswordHash performs dummy computation for timing uniformity
func ComputeFakePasswordHash() {
    salt := make([]byte, 16)
    for i := range salt {
        salt[i] = byte(i)
    }
    
    // Dummy Argon2 computation
    argon2.IDKey([]byte("dummy_password"), salt, 3, 64*1024, 4, 32)
}

// DummyPasswordHashDuration is the target duration for fake hash computation
const DummyPasswordHashDuration = 100 * time.Millisecond
```

---

### Step 6.3: Update Imports Everywhere

For each file that used old paths, update imports:

OLD:
```go
import "github.com/VinnsEdesigner/vyzorix/apps/api/pkg/storage"
```

NEW:
```go
import "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
```

---

### Step 6.4: Fix Remaining Duplications

Update `auth_password_reset.go` to use shared token generation:
```go
import "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/crypto"

// Replace local token generation with:
token, err := crypto.GenerateToken()
```

Update `auth_email_verify.go` similarly.

---

### Day 6-7 Checklist

- [ ] Deleted orphaned files
- [ ] Consolidated enum_safe.go
- [ ] Updated all imports
- [ ] Removed duplicated token generation
- [ ] All tests pass
- [ ] Build succeeds

---

## Day 8: Final Verification

**Goal:** Ensure everything works correctly  
**Duration:** 1 day  
**Risk:** LOW - Verification only

### Step 8.1: Run Full Test Suite

```bash
cd /workspace/project/6dc1227fe57247678b4d9fddc4699f3d/apps/api
export PATH="$PWD/go/bin:$PATH"

# Run all tests
go test ./... -v -count=1

# Run linter
golangci-lint run ./...
```

---

### Step 8.2: Verify Build

```bash
# Clean build
go clean
go build -o bin/api ./...

# Verify binary
./bin/api &
sleep 2
curl -s http://localhost:8080/health
kill %1
```

---

### Step 8.3: Manual API Testing

Test each endpoint:
```bash
# Health
curl http://localhost:8080/health

# Register
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123","name":"Test"}'

# Login
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123"}'

# Device Register
curl -X POST http://localhost:8080/v1/device/register \
  -H "Content-Type: application/json" \
  -d '{"device_id":"dev-123","firebase_install_id":"fid-123","app_version":"1.0.0","device_class":"phone"}'
```

---

### Step 8.4: Final Checklist

- [ ] All tests pass
- [ ] golangci-lint reports no issues
- [ ] Build succeeds
- [ ] Health endpoint works
- [ ] Register endpoint works
- [ ] Login endpoint works
- [ ] Device register endpoint works
- [ ] No panics or crashes

---

## Rollback Plan

If anything goes wrong:

### Immediate Rollback
```bash
cd /workspace/project/6dc1227fe57247678b4d9fddc4699f3d
git checkout -- apps/api/
```

### Step-by-Step Rollback

If only Day 0 crypto extraction failed:
```bash
# Remove crypto package
rm -rf apps/api/internal/infrastructure/crypto

# Restore original middleware files
git checkout apps/api/internal/api/middleware/request_signing.go
git checkout apps/api/internal/api/middleware/response_encryption.go
```

If only Day 1 domain layer failed:
```bash
# Remove domain package
rm -rf apps/api/internal/domain

# Restore storage files
git checkout apps/api/pkg/storage/
```

### Recovery Sequence

1. `git stash` any uncommitted changes
2. Identify failing step
3. Fix the specific step
4. `git stash pop`
5. Continue from where you left off

---

## Final Directory Structure

After ALL migration is complete:

```
apps/api/
 cmd/
    api/
        main.go

 internal/
    api/
       router.go
       auth_handler.go
       device_handler.go
       health_handler.go
   
    application/
       auth/
          service.go
          password.go
       device/
          service.go
       dto/
          auth.go
          device.go
       errors.go
   
    domain/
       errors.go
       operator/
          entity.go
          repository.go
       device/
          entity.go
          repository.go
       session/
          entity.go
          repository.go
       command/
          entity.go
          repository.go
       auth/
           enum_safe.go
   
    infrastructure/
       storage/
          sqlite.go
          operator.go
          device.go
          session.go
          command.go
      
       crypto/
           aes_gcm.go
           replay_cache.go
           token.go
   
    api/
        middleware/
            auth.go
            cors.go
            ratelimit.go
            ...

 pkg/
    config/
    logging/
    models/

 go.mod
 main.go
```

---

## Summary Timeline

| Day | Phase | Deliverable |
|-----|-------|-------------|
| 0 | Crypto Foundation | Shared crypto package |
| 1 | Domain Layer | Entities + Interfaces |
| 2 | Infrastructure | Storage implementations |
| 3-4 | Application | Use cases + DTOs |
| 5 | Router + Handlers | Clean HTTP layer |
| 6-7 | Cleanup | Remove duplication |
| 8 | Verify | Tests + Lint + Manual |

**Total: 8 days**

---

*This document is the single source of truth for the migration. Execute each step in order and verify before proceeding.*
