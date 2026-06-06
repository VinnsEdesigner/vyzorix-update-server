# Backend Bug Fixes & Implementation Requirements

> **Document Purpose:** This document outlines all identified bugs, misconfigurations, and implementation gaps in the Vyzorix Update Server backend, along with detailed requirements and implementation guidance for fixing each issue.

---

## Table of Contents

1. [WebSocket Origin Validation](#1-websocket-origin-validation)
2. [Device Online Status Integration](#2-device-online-status-integration)
3. [Enforce HMAC Dashboard Setting](#3-enforce-hmac-dashboard-setting)
4. [WebSocket Handler Upgrader Cleanup](#4-websocket-handler-upgrader-cleanup)
5. [Duplicate Device Registration Handlers](#5-duplicate-device-registration-handlers)
6. [Command Secrets Hash Column](#6-command-secrets-hash-column)
7. [HMAC Window Configuration](#7-hmac-window-configuration)
8. [JWT Secret Implementation](#8-jwt-secret-implementation)
9. [Google OAuth Signature Verification](#9-google-oauth-signature-verification)
10. [Unused postJSON Function](#10-unused-postjson-function)
11. [Consistent Error Response Format](#11-consistent-error-response-format)
12. [Rate Limiter Bucket Cleanup](#12-rate-limiter-bucket-cleanup)
13. [Command Status Implementation](#13-command-status-implementation)
14. [Retry and Cancel Command Implementation](#14-retry-and-cancel-command-implementation)
15. [Token ID Generation Safety](#15-token-id-generation-safety)

---

## 1. WebSocket Origin Validation

### Current State

Both WebSocket handlers disable origin checking:

```go
// controllers/websocket_handler.go:37
upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},

// controllers/server.go:48-50
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}
```

### Security Impact

- **Cross-Site WebSocket Hijacking (CSWSH):** Attackers can open WebSocket connections to the server from malicious websites
- **Session Hijacking:** If JWT is sent via query params, attackers can steal sessions
- **Command Injection:** Unauthorized access to device streams and telemetry

### Implementation Requirements

#### 1.1 Create Origin Validator

Create `security/origin.go`:

```go
package security

import (
    "net/url"
    "strings"
)

// OriginValidator validates WebSocket origins against an allowed list.
type OriginValidator struct {
    allowedOrigins map[string]bool
    allowedSchemes map[string]bool
}

// NewOriginValidator creates a validator with allowed origins from config.
func NewOriginValidator(origins []string) *OriginValidator {
    allowed := make(map[string]bool)
    schemes := make(map[string]bool)
    
    for _, origin := range origins {
        origin = strings.TrimSpace(origin)
        if origin == "*" {
            // Wildcard requires special handling
            allowed["*"] = true
            continue
        }
        
        u, err := url.Parse(origin)
        if err != nil {
            continue
        }
        allowed[origin] = true
        allowed[strings.ToLower(origin)] = true
        schemes[u.Scheme] = true
    }
    
    return &OriginValidator{
        allowedOrigins: allowed,
        allowedSchemes: schemes,
    }
}

// Validate checks if the origin is allowed.
// Returns true if:
// - Origin is empty (non-browser client)
// - Origin is "*" wildcard and no credentials required
// - Origin matches an entry in the allowed list
func (v *OriginValidator) Validate(origin string) bool {
    // Empty origin - non-browser client (curl, mobile app, etc.)
    if origin == "" {
        return true
    }
    
    // Wildcard allowed
    if v.allowedOrigins["*"] {
        return true
    }
    
    // Direct match
    if v.allowedOrigins[origin] {
        return true
    }
    
    // Case-insensitive match
    if v.allowedOrigins[strings.ToLower(origin)] {
        return true
    }
    
    // Parse and check scheme + host match
    u, err := url.Parse(origin)
    if err != nil {
        return false
    }
    
    // Only allow wss:// in production
    if u.Scheme == "http" || u.Scheme == "ws" {
        return false
    }
    
    return false
}

// CheckOrigin returns a function compatible with gorilla/websocket Upgrader.
func (v *OriginValidator) CheckOrigin() func(*http.Request) bool {
    return func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        return v.Validate(origin)
    }
}
```

#### 1.2 Update WebSocket Handlers

Update `controllers/server.go`:

```go
func New(log *slog.Logger, cfg config.Config, st *storage.Store, h *hub.Hub, notifier fcm.Notifier) *Server {
    s := &Server{Log: log, Config: cfg, Store: st, Hub: h, Notifier: notifier}
    
    // Create origin validator from config
    s.originValidator = security.NewOriginValidator(cfg.AllowedOrigins)
    
    // ... rest of initialization
}

// Update upgrader initialization
var upgrader = websocket.Upgrader{
    CheckOrigin: s.originValidator.CheckOrigin(),
    // Add security headers
    HandshakeTimeout: 10 * time.Second,
}
```

Update `controllers/websocket_handler.go`:

```go
func NewWebSocketHandler(...) *WebSocketHandler {
    originValidator := security.NewOriginValidator(cfg.AllowedOrigins)
    
    return &WebSocketHandler{
        // ...
        upgrader: websocket.Upgrader{
            CheckOrigin: originValidator.CheckOrigin(),
            HandshakeTimeout: 10 * time.Second,
        },
    }
}
```

#### 1.3 Add Origin Logging

Log rejected origins for security monitoring:

```go
func (v *OriginValidator) CheckOrigin() func(*http.Request) bool {
    return func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        if !v.Validate(origin) && origin != "" {
            // Log security event
            log.Warn("websocket origin rejected", 
                "origin", origin, 
                "remoteAddr", r.RemoteAddr,
                "path", r.URL.Path,
            )
        }
        return v.Validate(origin)
    }
}
```

### Verification Checklist

- [ ] Origin validator created in `security/origin.go`
- [ ] Both WebSocket handlers use the validator
- [ ] Production origins configured in `ALLOWED_ORIGINS` env var
- [ ] Rejected origins are logged
- [ ] Test with curl (no origin) - should succeed
- [ ] Test with browser - should validate against allowed list

---

## 2. Device Online Status Integration

### Current State

`controllers/device.go:176-179`:

```go
func (s *DeviceController) isDeviceOnline(deviceID string) bool {
    // This would be implemented by checking the hub's active connections
    // In a full implementation, this would integrate with the hub
    return false
}
```

### Implementation Requirements

#### 2.1 Hub Integration

The `Hub` already tracks online devices. We need to expose this properly:

```go
// hub/hub.go - Add Online method (already exists at line 100-104)
func (h *Hub) Online(deviceID string) bool {
    h.mu.RLock()
    defer h.mu.RUnlock()
    return h.clients[deviceID] != nil
}
```

#### 2.2 DeviceController Integration

Update `controllers/device.go`:

```go
type DeviceController struct {
    log       *slog.Logger
    config    config.Config
    store     *storage.Store
    hmac      security.Verifier
    hub       *hub.Hub  // Add hub reference
}

func NewDeviceController(
    log *slog.Logger, 
    cfg config.Config, 
    st *storage.Store, 
    hmac security.Verifier,
    hub *hub.Hub,  // Add hub parameter
) *DeviceController {
    return &DeviceController{
        log:    log,
        config: cfg,
        store:  st,
        hmac:   hmac,
        hub:    hub,
    }
}

// isDeviceOnline checks if a device has an active WebSocket connection.
func (s *DeviceController) isDeviceOnline(deviceID string) bool {
    if s.hub == nil {
        return false
    }
    return s.hub.Online(deviceID)
}
```

#### 2.3 Server Initialization

Update `controllers/server.go`:

```go
func (s *Server) Engine() *gin.Engine {
    // ... existing setup ...
    
    deviceCtrl := NewDeviceController(log, cfg, st, s.HMAC, s.Hub)
    
    // ... routes using deviceCtrl ...
}
```

#### 2.4 Device Status Events

Add support for device online/offline notifications:

```go
// hub/hub.go - Add event channel
type Hub struct {
    // ... existing fields ...
    events chan *DeviceEvent  // Add event channel
}

type DeviceEvent struct {
    DeviceID string
    Type     string  // "online", "offline"
    Time     time.Time
}

// Update Run() to emit events
func (h *Hub) Run(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case c := <-h.register:
            // ... existing registration logic ...
            // Emit event
            if h.events != nil {
                select {
                case h.events <- &DeviceEvent{
                    DeviceID: c.DeviceID,
                    Type:     "online",
                    Time:     time.Now(),
                }:
                default:
                }
            }
        case c := <-h.unreg:
            // ... existing unregistration logic ...
            // Emit event
            if h.events != nil {
                select {
                case h.events <- &DeviceEvent{
                    DeviceID: c.DeviceID,
                    Type:     "offline",
                    Time:     time.Now(),
                }:
                default:
                }
            }
        }
    }
}
```

### Database Online Tracking

The `devices` table should track online status. Add migration:

```go
// storage/migrations.go - Add to migrate()
if _, err := s.db.ExecContext(ctx, `ALTER TABLE devices ADD COLUMN is_online INTEGER NOT NULL DEFAULT 0`); err != nil {
    // Ignore if column exists
}
```

Update hub to sync with database:

```go
// hub/hub.go - Update Run()
case c := <-h.register:
    h.mu.Lock()
    h.clients[c.DeviceID] = c
    h.mu.Unlock()
    // Sync to DB
    _ = h.store.SetOnline(context.Background(), c.DeviceID, true)
    // Also sync on unregister
```

### Verification Checklist

- [ ] `DeviceController` has hub reference
- [ ] `isDeviceOnline()` returns actual hub status
- [ ] Device status endpoint returns correct online state
- [ ] Dashboard device list shows correct online/offline status
- [ ] Database `is_online` column stays in sync with WebSocket state

---

## 3. Enforce HMAC Dashboard Setting

### Current State

The `EnforceHMAC` setting exists in config but isn't connected to a UI setting. When `false`, WebSocket connections have no authentication.

### Implementation Requirements

#### 3.1 Database Setting Store

Create a settings table:

```go
// storage/sqlite.go - Add settings migration
func (s *Store) migrateSettings(ctx context.Context) error {
    if _, err := s.db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS settings (
            key TEXT PRIMARY KEY,
            value TEXT NOT NULL,
            updated_at INTEGER NOT NULL
        )
    `); err != nil {
        return err
    }
    
    // Set defaults if not exist
    defaults := map[string]string{
        "enforce_hmac":         "false",
        "rate_limit_capacity":  "100",
        "rate_limit_refill":    "60",
        "hmac_window_seconds": "300",
    }
    
    for k, v := range defaults {
        s.db.ExecContext(ctx, `
            INSERT OR IGNORE INTO settings(key, value, updated_at) VALUES(?, ?, ?)
        `, k, v, time.Now().UTC().UnixMilli())
    }
    
    return nil
}

// Settings methods
func (s *Store) GetSetting(ctx context.Context, key string) (string, error) {
    var value string
    err := s.db.QueryRowContext(ctx, 
        `SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
    if errors.Is(err, sql.ErrNoRows) {
        return "", nil
    }
    return value, err
}

func (s *Store) SetSetting(ctx context.Context, key, value string) error {
    _, err := s.db.ExecContext(ctx,
        `INSERT OR REPLACE INTO settings(key, value, updated_at) VALUES(?, ?, ?)`,
        key, value, time.Now().UTC().UnixMilli())
    return err
}

func (s *Store) GetEnforceHMAC(ctx context.Context) (bool, error) {
    val, err := s.GetSetting(ctx, "enforce_hmac")
    if err != nil || val == "" {
        return false, err
    }
    return val == "true" || val == "1", nil
}
```

#### 3.2 Settings Controller

Create `controllers/settings.go`:

```go
package controllers

import (
    "encoding/json"
    "log/slog"
    "net/http"
    "strconv"
    "time"

    "github.com/VinnsEdesigner/vyzorix-update-server/config"
    "github.com/VinnsEdesigner/vyzorix-update-server/storage"
    "github.com/gin-gonic/gin"
)

// SettingsController handles system settings management.
type SettingsController struct {
    log    *slog.Logger
    store  *storage.Store
    config config.Config
}

func NewSettingsController(log *slog.Logger, cfg config.Config, st *storage.Store) *SettingsController {
    return &SettingsController{log: log, store: st, config: cfg}
}

// GetSettings returns all current settings.
func (sc *SettingsController) GetSettings(c *gin.Context) {
    settings := map[string]interface{}{
        "enforceHMAC":        false,
        "hmacWindowSeconds":  300,
        "rateLimitCapacity":  100,
        "rateLimitRefill":    60,
    }
    
    // Load from database
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()
    
    if val, err := sc.store.GetSetting(ctx, "enforce_hmac"); err == nil {
        settings["enforceHMAC"] = val == "true" || val == "1"
    }
    
    if val, err := sc.store.GetSetting(ctx, "hmac_window_seconds"); err == nil {
        if n, err := strconv.Atoi(val); err == nil {
            settings["hmacWindowSeconds"] = n
        }
    }
    
    if val, err := sc.store.GetSetting(ctx, "rate_limit_capacity"); err == nil {
        if n, err := strconv.Atoi(val); err == nil {
            settings["rateLimitCapacity"] = n
        }
    }
    
    c.JSON(http.StatusOK, settings)
}

// UpdateSettings updates system settings.
func (sc *SettingsController) UpdateSettings(c *gin.Context) {
    var req struct {
        EnforceHMAC        *bool `json:"enforceHMAC"`
        HMACWindowSeconds  *int  `json:"hmacWindowSeconds"`
        RateLimitCapacity  *int  `json:"rateLimitCapacity"`
    }
    
    if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
        return
    }
    
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()
    
    changes := make(map[string]string)
    
    if req.EnforceHMAC != nil {
        val := "false"
        if *req.EnforceHMAC {
            val = "true"
            // Require TOKEN_SECRET when enabling HMAC
            if sc.config.TokenSecret == "" {
                c.JSON(http.StatusBadRequest, gin.H{
                    "error":   "configuration_required",
                    "message": "TOKEN_SECRET must be set to enable HMAC enforcement",
                })
                return
            }
        }
        changes["enforce_hmac"] = val
    }
    
    if req.HMACWindowSeconds != nil {
        if *req.HMACWindowSeconds < 10 || *req.HMACWindowSeconds > 3600 {
            c.JSON(http.StatusBadRequest, gin.H{
                "error":   "invalid_value",
                "message": "hmacWindowSeconds must be between 10 and 3600",
            })
            return
        }
        changes["hmac_window_seconds"] = strconv.Itoa(*req.HMACWindowSeconds)
    }
    
    if req.RateLimitCapacity != nil {
        if *req.RateLimitCapacity < 1 || *req.RateLimitCapacity > 10000 {
            c.JSON(http.StatusBadRequest, gin.H{
                "error":   "invalid_value",
                "message": "rateLimitCapacity must be between 1 and 10000",
            })
            return
        }
        changes["rate_limit_capacity"] = strconv.Itoa(*req.RateLimitCapacity)
    }
    
    // Apply changes
    for key, value := range changes {
        if err := sc.store.SetSetting(ctx, key, value); err != nil {
            sc.log.Error("failed to save setting", "key", key, "err", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "save_failed"})
            return
        }
        sc.log.Info("setting updated", "key", key, "value", value)
    }
    
    c.JSON(http.StatusOK, gin.H{
        "message": "settings updated",
        "changes": changes,
    })
}
```

#### 3.3 Routes

Add to `controllers/server.go`:

```go
func (s *Server) Engine() *gin.Engine {
    // ... existing setup ...
    
    settingsCtrl := NewSettingsController(log, cfg, st)
    
    // Settings routes (JWT protected)
    settings := r.Group("/v1/settings", JWTAuth(s.jwtCtrl.jwt, s.Store))
    settings.GET("", settingsCtrl.GetSettings)
    settings.PUT("", settingsCtrl.UpdateSettings)
    
    // ... rest of routes ...
}
```

#### 3.4 Dynamic HMAC Enforcement

Update the Server to check settings dynamically:

```go
type Server struct {
    // ... existing fields ...
    enforceHMACCache atomic.Bool
}

func (s *Server) ShouldEnforceHMAC(ctx context.Context) bool {
    // Check cache first (1 second TTL)
    if s.enforceHMACCache.Load() {
        return s.Config.EnforceHMAC
    }
    
    // Load from database
    val, err := s.Store.GetEnforceHMAC(ctx)
    if err != nil {
        return s.Config.EnforceHMAC  // Fallback to config
    }
    
    s.enforceHMACCache.Store(val)
    
    // Refresh cache every second
    go func() {
        time.Sleep(time.Second)
        s.enforceHMACCache.Store(false)  // Invalidate
    }()
    
    return val
}
```

Update WebSocket handler to use dynamic check:

```go
func (s *Server) stream(c *gin.Context) {
    enforceHMAC, _ := s.ShouldEnforceHMAC(c.Request.Context())
    
    if enforceHMAC {
        if err := s.HMAC.Verify(...); err != nil {
            c.JSON(401, map[string]string{"error": "bad_hmac", "message": err.Error()})
            return
        }
    }
    // ... rest of WebSocket upgrade ...
}
```

### Verification Checklist

- [ ] Settings table created with defaults
- [ ] Settings API endpoints functional
- [ ] HMAC enforcement reads from database, not just config
- [ ] Frontend Settings page has toggle for HMAC enforcement
- [ ] Changes take effect immediately (no restart required)
- [ ] Token secret required when enabling HMAC

---

## 4. WebSocket Handler Upgrader Cleanup

### Current State

Two separate WebSocket handlers with duplicate upgrader configurations:

1. `controllers/server.go:48-50` - Global `upgrader` variable
2. `controllers/websocket_handler.go:37` - Handler-specific upgrader

### Implementation Requirements

#### 4.1 Create Single Upgrader Factory

Create `hub/upgrader.go`:

```go
package hub

import (
    "net/http"
    "time"

    "github.com/gorilla/websocket"
)

// UpgraderConfig holds WebSocket upgrade configuration.
type UpgraderConfig struct {
    ReadLimit       int64
    WriteLimit      int64
    ReadTimeout     time.Duration
    WriteTimeout    time.Duration
    PongTimeout     time.Duration
    PingPeriod      time.Duration
    CheckOrigin     func(*http.Request) bool
    EnableCompression bool
}

// DefaultUpgraderConfig returns sensible defaults.
func DefaultUpgraderConfig() UpgraderConfig {
    return UpgraderConfig{
        ReadLimit:   1 << 20,  // 1MB
        WriteLimit:  1 << 20,  // 1MB
        ReadTimeout: 70 * time.Second,
        WriteTimeout: 10 * time.Second,
        PongTimeout:  70 * time.Second,
        PingPeriod:  30 * time.Second,
        CheckOrigin: func(r *http.Request) bool { return true },  // Override this!
    }
}

// NewUpgrader creates a configured websocket.Upgrader.
func NewUpgrader(cfg UpgraderConfig) websocket.Upgrader {
    return websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
        CheckOrigin:     cfg.CheckOrigin,
        EnableCompression: cfg.EnableCompression,
    }
}
```

#### 4.2 Update Server to Use Shared Upgrader

```go
type Server struct {
    // ... existing fields ...
    upgrader websocket.Upgrader
}

func New(...) *Server {
    s := &Server{...}
    
    // Create single upgrader with origin validation
    originCfg := hub.DefaultUpgraderConfig()
    originCfg.CheckOrigin = security.NewOriginValidator(cfg.AllowedOrigins).CheckOrigin()
    s.upgrader = hub.NewUpgrader(originCfg)
    
    return s
}

func (s *Server) stream(c *gin.Context) {
    // Use shared upgrader
    conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
    // ...
}
```

#### 4.3 Remove Duplicate Handler

The `WebSocketHandler` in `controllers/websocket_handler.go` appears to be unused. Verify and remove if duplicate:

```bash
grep -r "NewWebSocketHandler" .
```

If unused, delete `controllers/websocket_handler.go`.

### Verification Checklist

- [ ] Single upgrader configuration used across all WebSocket handlers
- [ ] Origin validation applied consistently
- [ ] No duplicate upgrader variables
- [ ] Unused handler code removed

---

## 5. Duplicate Device Registration Handlers

### Current State

Two device registration endpoints exist:

1. `controllers/server.go:89` - `s.register` method
2. `controllers/device.go:32-62` - `DeviceController.Register` method

### Implementation Requirements

#### 5.1 Audit Route Registration

Check `server.go Engine()` for all device routes:

```go
// Current (problematic):
r.POST("/v1/device/register", s.register)  // Defined at line 89

// DeviceController routes need to be checked
// Remove duplicate if found
```

#### 5.2 Consolidate to Single Handler

Keep `controllers/device.go:DeviceController.Register` as the canonical handler and remove the inline `register` method from `Server`.

```go
// controllers/server.go - Remove this method
func (s *Server) register(c *gin.Context) { ... }  // DELETE THIS

// Use DeviceController instead:
func (s *Server) Engine() *gin.Engine {
    deviceCtrl := NewDeviceController(log, cfg, st, s.HMAC, s.Hub)
    
    r.POST("/v1/device/register", deviceCtrl.Register)
    // OR use server.go inline but remove from device.go
}
```

#### 5.3 Standardize Response Format

Both handlers should use the same response structure:

```go
// models/device.go
type RegisterResponse struct {
    DeviceID      string `json:"deviceId"`
    CommandSecret string `json:"commandSecret"`
    RegisteredAt  int64  `json:"registeredAt"`
    ServerTime    int64  `json:"serverTime"`
}
```

### Verification Checklist

- [ ] Only one `/v1/device/register` handler exists
- [ ] Same response format from both success paths
- [ ] Consistent error handling
- [ ] No route conflicts

---

## 6. Command Secrets Hash Column

### Current State

`storage/sqlite.go:59` stores the raw `command_secret`:

```go
command_secret TEXT NOT NULL
```

### Documentation Requirement

Per `DEVICE_REGISTRATION.md` §6:

> "The raw `command_secret` is needed server-side because the server signs commands on behalf of the dashboard... It does NOT live in the SQLite `devices` table... Only a hash should be stored so that: Existence checks are fast, An attacker who exfiltrates data.db cannot recover the raw secret."

### Implementation Requirements

#### 6.1 Add Hash Column

```go
// storage/sqlite.go - Add migration
func (s *Store) migrateSecrets(ctx context.Context) error {
    // Add command_secret_hash column
    if _, err := s.db.ExecContext(ctx, `
        ALTER TABLE devices ADD COLUMN command_secret_hash TEXT
    `); err != nil {
        // Ignore if column exists (SQLite doesn't support IF NOT EXISTS for columns)
        if !strings.Contains(err.Error(), "duplicate column") {
            return err
        }
    }
    return nil
}
```

#### 6.2 Update Registration

```go
// storage/sqlite.go - Update Register method
func (s *Store) Register(ctx context.Context, req models.RegisterRequest) (result struct{...}, isNew bool, err error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Generate command secret
    secretBytes := make([]byte, 32)
    if _, err := rand.Read(secretBytes); err != nil {
        return result, false, err
    }
    commandSecret := hex.EncodeToString(secretBytes)
    
    // Hash the secret for storage (bcrypt as per docs)
    hasher := services.NewCommandSigner()
    secretHash := hasher.HashSecret(commandSecret)
    
    // Store hash in DB, return raw secret only in response
    // ...
}
```

#### 6.3 Create Secret Store

Create `security/secretstore/`:

```go
package secretstore

import (
    "context"
    "os"
    "path/filepath"
    "sync"

    "github.com/VinnsEdesigner/vyzorix-update-server/security"
)

// FileStore implements secret storage using encrypted files.
type FileStore struct {
    dir     string
    hasher  *services.CommandSigner
    cache   map[string]string
    cacheMu sync.RWMutex
}

// NewFileStore creates a new file-based secret store.
func NewFileStore(dataDir string) (*FileStore, error) {
    dir := filepath.Join(dataDir, "secrets")
    if err := os.MkdirAll(dir, 0700); err != nil {
        return nil, err
    }
    return &FileStore{
        dir:    dir,
        hasher: services.NewCommandSigner(),
        cache:  make(map[string]string),
    }, nil
}

// Get retrieves a device's command secret.
func (s *FileStore) Get(ctx context.Context, deviceID string) (string, bool) {
    // Check cache first
    s.cacheMu.RLock()
    if secret, ok := s.cache[deviceID]; ok {
        s.cacheMu.RUnlock()
        return secret, true
    }
    s.cacheMu.RUnlock()
    
    // Load from file
    path := filepath.Join(s.dir, deviceID+".enc")
    data, err := os.ReadFile(path)
    if err != nil {
        return "", false
    }
    
    // Decrypt (implementation depends on encryption strategy)
    secret := string(data)  // Placeholder - implement proper encryption
    
    // Update cache
    s.cacheMu.Lock()
    s.cache[deviceID] = secret
    s.cacheMu.Unlock()
    
    return secret, true
}

// Put stores a device's command secret.
func (s *FileStore) Put(ctx context.Context, deviceID, secret string) error {
    path := filepath.Join(s.dir, deviceID+".enc")
    
    // Encrypt before writing
    // ... encryption logic ...
    
    if err := os.WriteFile(path, []byte(secret), 0600); err != nil {
        return err
    }
    
    // Update cache
    s.cacheMu.Lock()
    s.cache[deviceID] = secret
    s.cacheMu.Unlock()
    
    return nil
}

// Delete removes a device's secret.
func (s *FileStore) Delete(ctx context.Context, deviceID string) error {
    path := filepath.Join(s.dir, deviceID+".enc")
    
    if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
        return err
    }
    
    s.cacheMu.Lock()
    delete(s.cache, deviceID)
    s.cacheMu.Unlock()
    
    return nil
}
```

#### 6.4 Update Device Controller

```go
func (s *DeviceController) Register(c *gin.Context) {
    // ...
    d, isNew, err := s.store.Register(c.Request.Context(), req)
    // ...
    // CommandSecret is returned only in this response
    // Hash is stored in DB
    // Raw secret stored in secretstore
}
```

### Verification Checklist

- [ ] `command_secret_hash` column added
- [ ] Raw secret never returned after registration
- [ ] Hash used for existence checks
- [ ] Secret store implemented for raw secrets
- [ ] Hash verified with bcrypt

---

## 7. HMAC Window Configuration

### Current State

`config/config.go:74`:

```go
HMACWindow: 5 * time.Minute,  // 300 seconds
```

### Documentation Requirement

Per `COMMAND_SECURITY.md` §4:

> "TTL: 5 minutes — matches the maximum timestamp drift window with 2.5x safety margin"

But §3 states:

> "Timestamp window check (±30s)"

The discrepancy: Nonce TTL is 5 minutes, but timestamp window should be 30 seconds.

### Implementation Requirements

#### 7.1 Separate Nonce and Timestamp Windows

```go
// config/config.go
type Config struct {
    // ... existing fields ...
    HMACWindow        time.Duration  // Timestamp window (±30s default)
    NonceCacheTTL     time.Duration  // Nonce deduplication window (5min default)
}

func Load() (Config, error) {
    // ...
    c := Config{
        // ...
        HMACWindow:     30 * time.Second,   // Per docs: ±30s
        NonceCacheTTL:  5 * time.Minute,    // Per docs: 5min TTL
    }
    
    // Allow override
    if v := os.Getenv("HMAC_WINDOW_SECONDS"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 {
            c.HMACWindow = time.Duration(n) * time.Second
        }
    }
    if v := os.Getenv("NONCE_CACHE_TTL_SECONDS"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 {
            c.NonceCacheTTL = time.Duration(n) * time.Second
        }
    }
    
    return c, nil
}
```

#### 7.2 Update Nonce Cache

```go
// security/hmac.go
type Verifier struct {
    Secret       func(deviceID string) (string, bool)
    Nonces       *NonceCache
    Window       time.Duration  // Timestamp window
    NonceWindow  time.Duration  // Nonce TTL window
}

func (v Verifier) Verify(...) error {
    // Check timestamp against Window (30s)
    milli, _ := strconv.ParseInt(ts, 10, 64)
    t := time.UnixMilli(milli)
    if t.Before(now.Add(-v.Window)) || t.After(now.Add(v.Window)) {
        return fmt.Errorf("timestamp outside replay window")
    }
    
    // Check nonce against NonceWindow (5min)
    if v.Nonces != nil && !v.Nonces.Use(deviceID+":"+nonce, now) {
        return fmt.Errorf("replayed nonce")
    }
}
```

### Verification Checklist

- [ ] `HMACWindow` defaults to 30 seconds
- [ ] `NonceCacheTTL` defaults to 5 minutes
- [ ] Both configurable via environment variables
- [ ] Dashboard settings allow separate configuration

---

## 8. JWT Secret Implementation

### Current State

`security/jwt.go:37-40`:

```go
h := sha256.New()
h.Write([]byte(secret))
return &JWTManager{
    secret: h.Sum(nil),  // Only 32 bytes used
```

### Documentation Requirement

Per `UPDATE_SERVER_ARCHITECTURE_SPEC.md`:

> "TokenSecret: secret for validating dashboard API requests"
> "JWTSecret: explicit separate key for JWT signing"

The docs require distinct keys and proper key handling.

### Implementation Requirements

#### 8.1 Key Derivation Function

```go
// security/jwt.go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "fmt"
    
    "golang.org/x/crypto/argon2"
)

// deriveSigningKey creates a proper signing key from a secret.
// Uses Argon2id for key derivation with a fixed salt.
func deriveSigningKey(secret string, purpose string, outputLen int) []byte {
    // Fixed salt per purpose to derive different keys
    salt := []byte("vyzorix-jwt-" + purpose)
    
    // Argon2id parameters (memory-hard function)
    key := argon2.IDKey(
        []byte(secret),
        salt,
        3,      // iterations
        64,     // memory in KB
        4,      // parallelism
        uint32(outputLen),
    )
    
    return key
}

// Or use HKDF for simpler key derivation:
func deriveKeyHKDF(secret, purpose string, outLen int) []byte {
    // Use HMAC-SHA256 as HKDF-like construct
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(purpose))
    
    // Expand to required length
    result := make([]byte, outLen)
    copy(result, h.Sum(nil))
    
    // Further rounds if needed
    for i := 1; len(result) < outLen; i++ {
        h = hmac.New(sha256.New, h.Sum([]byte{byte(i)}))
        h.Write([]byte(purpose))
        copy(result[len(h.Sum(nil)):], h.Sum(nil))
    }
    
    return result[:outLen]
}
```

#### 8.2 Update JWTManager

```go
type JWTManager struct {
    secret     []byte  // Signing key
    verifyKey  []byte  // Verification key (may differ)
    expiry     time.Duration
    issuer     string
}

func NewJWTManager(secret string, expiry time.Duration, issuer string) *JWTManager {
    // Derive separate signing and verification keys
    signingKey := deriveSigningKey(secret, "jwt-signing", 32)
    verifyKey := deriveSigningKey(secret, "jwt-verify", 32)
    
    return &JWTManager{
        secret:    signingKey,
        verifyKey: verifyKey,
        expiry:    expiry,
        issuer:    issuer,
    }
}
```

#### 8.3 Environment Variable Documentation

Add to `.env.example`:

```bash
# JWT Secret - used for signing operator session tokens
# IMPORTANT: Generate with: openssl rand -hex 64
# Must be at least 32 characters
JWT_SECRET=

# Token Secret - used for dashboard API authentication  
# IMPORTANT: Generate with: openssl rand -hex 64
# Must be at least 32 characters
TOKEN_SECRET=
```

### Verification Checklist

- [ ] JWT signing key derived with KDF
- [ ] Separate keys for signing vs verification
- [ ] Minimum key length enforced
- [ ] Clear documentation of secret requirements

---

## 9. Google OAuth Signature Verification

### Current State

`controllers/auth.go:447-458`:

```go
// decodeJWTPayload extracts the payload from a JWT without signature verification.
// In production, fetch Google's public keys and verify the signature.
// For Phase 1.5, we trust the token format and extract the claims directly.
func decodeJWTPayload(token string, out any) error {
    // WARNING: Does not verify signature!
```

### Implementation Requirements

#### 9.1 Remove Unsafe Function

Remove `decodeJWTPayload` entirely. Use `GoogleTokenVerifier.Verify()` instead.

#### 9.2 Update Google Callback

```go
func (ac *AuthController) GoogleCallback(c *gin.Context) {
    code := c.Query("code")
    if code == "" {
        c.JSON(400, gin.H{"error": "missing_code"})
        return
    }
    
    // Exchange code for tokens
    tokens, err := ac.exchangeCode(code)
    if err != nil {
        ac.log.Warn("google callback: token exchange failed", "err", err)
        c.JSON(400, gin.H{"error": "token_exchange_failed"})
        return
    }
    
    // CRITICAL: Verify the ID token signature
    claims, err := ac.googleVer.Verify(tokens.IDToken)
    if err != nil {
        ac.log.Warn("google callback: ID token verification failed", "err", err)
        c.JSON(401, gin.H{"error": "invalid_id_token"})
        return
    }
    
    // Verify audience matches our client ID
    if claims.Aud != ac.config.GoogleOAuthClientID {
        ac.log.Warn("google callback: wrong audience", "got", claims.Aud, "want", ac.config.GoogleOAuthClientID)
        c.JSON(401, gin.H{"error": "wrong_audience"})
        return
    }
    
    // Use verified claims...
}
```

#### 9.3 Ensure GoogleTokenVerifier Is Properly Configured

```go
func NewGoogleTokenVerifier(audience string) *GoogleTokenVerifier {
    return &GoogleTokenVerifier{
        client:   &http.Client{Timeout: 10 * time.Second},
        jwksURL:  googleJWKSURL,
        keys:     make(map[string]*rsa.PublicKey),
        cacheTTL: 1 * time.Hour,
        audience: audience,  // Set to GoogleOAuthClientID
    }
}
```

### Verification Checklist

- [ ] `decodeJWTPayload` function removed
- [ ] All Google token verification uses `GoogleTokenVerifier.Verify()`
- [ ] Audience is validated against client ID
- [ ] Failed verification is logged and rejected

---

## 10. Unused postJSON Function

### Current State

`controllers/auth.go:422-444` - `postJSON` is defined but never used.

### Implementation Requirements

Either:

**Option A: Remove if unused**
```bash
# Verify no usages
grep -n "postJSON" controllers/auth.go
# If only definition, remove the function
```

**Option B: Use it if needed**

If the function is intended for future use (e.g., calling external OAuth endpoints), keep it but add proper error handling:

```go
func postJSON(ctx context.Context, url string, body any, resp any) error {
    reqBody, err := json.Marshal(body)
    if err != nil {
        return fmt.Errorf("marshal error: %w", err)
    }
    
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(reqBody)))
    if err != nil {
        return fmt.Errorf("request error: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{Timeout: 10 * time.Second}
    httpResp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("request failed: %w", err)
    }
    defer httpResp.Body.Close()
    
    rb, err := io.ReadAll(httpResp.Body)
    if err != nil {
        return fmt.Errorf("read error: %w", err)
    }
    
    if httpResp.StatusCode >= 400 {
        return fmt.Errorf("HTTP %d: %s", httpResp.StatusCode, string(rb))
    }
    
    if resp != nil {
        if err := json.Unmarshal(rb, resp); err != nil {
            return fmt.Errorf("unmarshal error: %w", err)
        }
    }
    
    return nil
}
```

### Verification Checklist

- [ ] Function is either removed or documented as intended for future use
- [ ] No compiler warnings about unused functions

---

## 11. Consistent Error Response Format

### Current State

Inconsistent error formats across endpoints:

```go
// Format 1: map[string]string
c.JSON(400, map[string]string{"error": "bad_json", "message": err.Error()})

// Format 2: models.ErrorResponse
c.JSON(400, models.ErrorResponse{Error: "bad_request", Message: "invalid JSON body"})

// Format 3: gin.H
c.JSON(400, gin.H{"error": "bad_request", "message": "invalid JSON body"})
```

### Implementation Requirements

#### 11.1 Define Standard Error Types

```go
// models/response.go
package models

import "time"

// ErrorResponse is the standard error response format.
type ErrorResponse struct {
    Error     string                 `json:"error"`
    Message   string                 `json:"message"`
    Details   map[string]interface{} `json:"details,omitempty"`
    Code      string                 `json:"code,omitempty"`
    Timestamp int64                  `json:"timestamp"`
}

// ErrorWithCode creates an ErrorResponse with an error code.
func ErrorWithCode(error, message, code string) ErrorResponse {
    return ErrorResponse{
        Error:     error,
        Message:   message,
        Code:      code,
        Timestamp: time.Now().UnixMilli(),
    }
}

// BadRequest creates a 400 error response.
func BadRequest(message string) ErrorResponse {
    return ErrorResponse{
        Error:     "bad_request",
        Message:   message,
        Timestamp: time.Now().UnixMilli(),
    }
}

// Unauthorized creates a 401 error response.
func Unauthorized(message string) ErrorResponse {
    return ErrorResponse{
        Error:     "unauthorized",
        Message:   message,
        Timestamp: time.Now().UnixMilli(),
    }
}

// Forbidden creates a 403 error response.
func Forbidden(message string) ErrorResponse {
    return ErrorResponse{
        Error:     "forbidden",
        Message:   message,
        Timestamp: time.Now().UnixMilli(),
    }
}

// NotFound creates a 404 error response.
func NotFound(message string) ErrorResponse {
    return ErrorResponse{
        Error:     "not_found",
        Message:   message,
        Timestamp: time.Now().UnixMilli(),
    }
}

// InternalError creates a 500 error response.
func InternalError(message string) ErrorResponse {
    return ErrorResponse{
        Error:     "internal_error",
        Message:   message,
        Timestamp: time.Now().UnixMilli(),
    }
}

// ValidationError creates a 422 error with field details.
func ValidationError(details map[string]interface{}) ErrorResponse {
    return ErrorResponse{
        Error:     "validation_error",
        Message:   "request validation failed",
        Details:   details,
        Timestamp: time.Now().UnixMilli(),
    }
}
```

#### 11.2 Create Middleware for Error Handling

```go
// middleware/errors.go
package middleware

import (
    "log/slog"
    "net/http"
    
    "github.com/VinnsEdesigner/vyzorix-update-server/models"
    "github.com/gin-gonic/gin"
)

// ErrorHandler provides consistent error responses.
type ErrorHandler struct {
    log *slog.Logger
}

func NewErrorHandler(log *slog.Logger) *ErrorHandler {
    return &ErrorHandler{log: log}
}

// Handle converts errors to consistent responses.
func (e *ErrorHandler) Handle() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            e.log.Warn("request error", "err", err, "path", c.Request.URL.Path)
            
            // Don't override already-written response
            if c.Writer.Written() {
                return
            }
            
            c.JSON(http.StatusInternalServerError, models.InternalError(err.Error()))
        }
    }
}

// AbortWithError sends a standardized error and aborts.
func (e *ErrorHandler) AbortWithError(c *gin.Context, status int, errResp models.ErrorResponse) {
    e.log.Warn("aborting request", 
        "status", status, 
        "error", errResp.Error, 
        "message", errResp.Message,
        "path", c.Request.URL.Path,
    )
    c.AbortWithStatusJSON(status, errResp)
}
```

#### 11.3 Update All Controllers

Replace all error responses with the standard format:

```go
// Before
c.JSON(400, map[string]string{"error": "bad_json", "message": err.Error()})

// After
c.JSON(400, models.BadRequest(err.Error()))
```

### Verification Checklist

- [ ] All endpoints use `models.ErrorResponse` or helpers
- [ ] Error middleware installed
- [ ] No `map[string]string` or `gin.H` error responses
- [ ] Client can rely on consistent format

---

## 12. Rate Limiter Bucket Cleanup

### Current State

`middleware/rate_limiter.go` - Buckets accumulate indefinitely:

```go
type RateLimiter struct {
    mu       sync.Mutex
    buckets  map[string]*bucket
    Capacity int
    Refill   time.Duration
}
```

### Implementation Requirements

#### 12.1 Add Cleanup Goroutine

```go
type RateLimiter struct {
    mu       sync.Mutex
    buckets  map[string]*bucket
    Capacity int
    Refill   time.Duration
    
    // Cleanup configuration
    maxIdleTime  time.Duration
    cleanupTick  time.Duration
    stopCleanup  chan struct{}
}

func NewRateLimiter(capacity int, refill time.Duration) *RateLimiter {
    return &RateLimiter{
        buckets:     map[string]*bucket{},
        Capacity:    capacity,
        Refill:      refill,
        maxIdleTime: 10 * time.Minute,  // Remove buckets idle for 10+ minutes
        cleanupTick: time.Minute,       // Cleanup every minute
        stopCleanup: make(chan struct{}),
    }
}

// StartCleanup begins periodic cleanup of stale buckets.
func (l *RateLimiter) StartCleanup() {
    go func() {
        ticker := time.NewTicker(l.cleanupTick)
        defer ticker.Stop()
        
        for {
            select {
            case <-l.stopCleanup:
                return
            case <-ticker.C:
                l.cleanupStale()
            }
        }
    }()
}

// StopCleanup stops the cleanup goroutine.
func (l *RateLimiter) StopCleanup() {
    close(l.stopCleanup)
}

func (l *RateLimiter) cleanupStale() {
    l.mu.Lock()
    defer l.mu.Unlock()
    
    now := time.Now()
    cutoff := now.Add(-l.maxIdleTime)
    
    for key, b := range l.buckets {
        if b.last.Before(cutoff) {
            delete(l.buckets, key)
        }
    }
}
```

#### 12.2 Update Bucket to Track Activity

```go
type bucket struct {
    tokens int
    last   time.Time
}

func (l *RateLimiter) Allow(key string) bool {
    l.mu.Lock()
    defer l.mu.Unlock()
    
    now := time.Now()
    b := l.buckets[key]
    if b == nil {
        b = &bucket{tokens: l.Capacity, last: now}
        l.buckets[key] = b
    }
    
    // Update last access time
    b.last = now
    
    // ... rest of rate limiting logic ...
}
```

#### 12.3 Start Cleanup on Server Init

```go
func New(log *slog.Logger, cfg config.Config, ...) *Server {
    limiter := middleware.NewRateLimiter(100, time.Minute)
    limiter.StartCleanup()  // Start cleanup goroutine
    
    // ... rest of init ...
}
```

### Verification Checklist

- [ ] Cleanup goroutine runs periodically
- [ ] Idle buckets are removed after timeout
- [ ] Memory usage stays bounded under load
- [ ] Server shutdown properly stops cleanup goroutine

---

## 13. Command Status Implementation

### Current State

`controllers/command.go:148-151`:

```go
func (s *CommandController) GetCommandStatus(c *gin.Context) {
    c.JSON(200, map[string]any{
        "dispatchId": dispatchID,
        "status":     "pending",  // Always returns pending!
    })
}
```

### Implementation Requirements

#### 13.1 Add Status to Commands Table

```go
// storage/sqlite.go - Add status column
if _, err := s.db.ExecContext(ctx, `
    ALTER TABLE commands ADD COLUMN status TEXT NOT NULL DEFAULT 'pending'
`); err != nil {
    if !strings.Contains(err.Error(), "duplicate column") {
        return err
    }
}

// Valid statuses:
// - pending: Command created, awaiting delivery
// - sent: Sent via WebSocket
// - delivered: Device acknowledged receipt
// - executing: Device is processing
// - completed: Command completed successfully
// - failed: Command failed
// - cancelled: Command was cancelled
```

#### 13.2 Add Status Methods

```go
// CommandStatus represents the status of a command.
type CommandStatus struct {
    DispatchID  string    `json:"dispatchId"`
    DeviceID    string    `json:"deviceId"`
    Command     string    `json:"command"`
    Args        string    `json:"args,omitempty"`
    Status      string    `json:"status"`
    Delivery    string    `json:"delivery"`
    CreatedAt   time.Time `json:"createdAt"`
    DeliveredAt *time.Time `json:"deliveredAt,omitempty"`
    CompletedAt *time.Time `json:"completedAt,omitempty"`
    Error       string    `json:"error,omitempty"`
}

func (s *Store) GetCommandStatus(ctx context.Context, dispatchID string) (*CommandStatus, error) {
    var cs CommandStatus
    var deliveredAt, completedAt sql.NullInt64
    
    err := s.db.QueryRowContext(ctx, `
        SELECT dispatch_id, device_id, command, args, status, delivery,
               created_at, delivered_at, completed_at, wake_error
        FROM commands WHERE dispatch_id = ?
    `, dispatchID).Scan(
        &cs.DispatchID, &cs.DeviceID, &cs.Command, &cs.Args,
        &cs.Status, &cs.Delivery, &cs.CreatedAt,
        &deliveredAt, &completedAt, &cs.Error,
    )
    
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    
    if deliveredAt.Valid {
        t := time.UnixMilli(deliveredAt.Int64).UTC()
        cs.DeliveredAt = &t
    }
    if completedAt.Valid {
        t := time.UnixMilli(completedAt.Int64).UTC()
        cs.CompletedAt = &t
    }
    
    return &cs, nil
}

func (s *Store) UpdateCommandStatus(ctx context.Context, dispatchID, status string) error {
    now := time.Now().UnixMilli()
    _, err := s.db.ExecContext(ctx, `
        UPDATE commands SET status = ? WHERE dispatch_id = ?
    `, status, dispatchID)
    return err
}
```

#### 13.3 Update GetCommandStatus Handler

```go
func (s *CommandController) GetCommandStatus(c *gin.Context) {
    dispatchID := c.Param("dispatchId")
    if dispatchID == "" {
        c.JSON(400, models.BadRequest("dispatch id required"))
        return
    }
    
    status, err := s.store.GetCommandStatus(c.Request.Context(), dispatchID)
    if err != nil {
        c.JSON(500, models.InternalError("lookup failed"))
        return
    }
    if status == nil {
        c.JSON(404, models.NotFound("command not found"))
        return
    }
    
    c.JSON(200, status)
}
```

### Verification Checklist

- [ ] Status column added to commands table
- [ ] `GetCommandStatus` returns actual status from database
- [ ] Status updated when command is sent/delivered/completed
- [ ] Dashboard shows accurate command status

---

## 14. Retry and Cancel Command Implementation

### Current State

`controllers/command.go:154-198` - Stub implementations:

```go
func (s *CommandController) RetryCommand(c *gin.Context) {
    // ... just returns success
    c.JSON(200, map[string]any{"retried": true})
}

func (s *CommandController) CancelCommand(c *gin.Context) {
    // ... just returns success
    c.JSON(200, map[string]any{"cancelled": true})
}
```

### Implementation Requirements

#### 14.1 Retry Command

```go
func (s *CommandController) RetryCommand(c *gin.Context) {
    dispatchID := c.Param("dispatchId")
    
    // Get original command
    status, err := s.store.GetCommandStatus(c.Request.Context(), dispatchID)
    if err != nil {
        c.JSON(500, models.InternalError("lookup failed"))
        return
    }
    if status == nil {
        c.JSON(404, models.NotFound("command not found"))
        return
    }
    
    // Only retry failed/pending commands
    if status.Status != "failed" && status.Status != "pending" && status.Status != "cancelled" {
        c.JSON(400, models.ErrorResponse{
            Error:   "invalid_status",
            Message: "can only retry failed, pending, or cancelled commands",
        })
        return
    }
    
    // Reset status to pending
    if err := s.store.UpdateCommandStatus(c.Request.Context(), dispatchID, "pending"); err != nil {
        c.JSON(500, models.InternalError("update failed"))
        return
    }
    
    // Re-send command
    delivery := "queued"
    if s.hub != nil && s.hub.Send(status.DeviceID, models.CommandFrame{
        Type:       "command",
        DispatchID: dispatchID,
        Command:    status.Command,
        Args:        []byte(status.Args),
    }) {
        delivery = "sent"
        _ = s.store.UpdateCommandStatus(c.Request.Context(), dispatchID, "sent")
    }
    
    // Trigger FCM wake if offline
    if delivery == "queued" {
        device, _, err := s.store.Device(c.Request.Context(), status.DeviceID)
        if err == nil && device.FCMToken != "" {
            _ = s.notifier.SendSilentWake(c.Request.Context(), fcm.SilentWake{
                Token:      device.FCMToken,
                Command:    status.Command,
                DispatchID: dispatchID,
                DeviceID:   status.DeviceID,
            })
        }
    }
    
    c.JSON(200, map[string]any{
        "dispatchId": dispatchID,
        "retried":    true,
        "delivery":   delivery,
    })
}
```

#### 14.2 Cancel Command

```go
func (s *CommandController) CancelCommand(c *gin.Context) {
    dispatchID := c.Param("dispatchId")
    
    // Get command status
    status, err := s.store.GetCommandStatus(c.Request.Context(), dispatchID)
    if err != nil {
        c.JSON(500, models.InternalError("lookup failed"))
        return
    }
    if status == nil {
        c.JSON(404, models.NotFound("command not found"))
        return
    }
    
    // Cannot cancel completed commands
    if status.Status == "completed" {
        c.JSON(400, models.ErrorResponse{
            Error:   "cannot_cancel",
            Message: "completed commands cannot be cancelled",
        })
        return
    }
    
    // Update status to cancelled
    if err := s.store.UpdateCommandStatus(c.Request.Context(), dispatchID, "cancelled"); err != nil {
        c.JSON(500, models.InternalError("update failed"))
        return
    }
    
    // If device is online, try to send cancellation
    if s.hub != nil && s.hub.Online(status.DeviceID) {
        // Send cancel frame
        s.hub.Send(status.DeviceID, models.CommandFrame{
            Type:       "cancel",
            DispatchID: dispatchID,
        })
    }
    
    c.JSON(200, map[string]any{
        "dispatchId": dispatchID,
        "cancelled":  true,
    })
}
```

### Verification Checklist

- [ ] `RetryCommand` re-sends failed/pending commands
- [ ] `CancelCommand` marks command as cancelled
- [ ] Cannot cancel completed commands
- [ ] Cannot retry non-existent or active commands
- [ ] FCM wake triggered for offline devices on retry

---

## 15. Token ID Generation Safety

### Current State

`security/jwt.go:96-99`:

```go
func generateTokenID() string {
    b := make([]byte, 16)
    _, _ = rand.Read(b)  // Error ignored!
    return base64.RawURLEncoding.EncodeToString(b)
}
```

### Implementation Requirements

#### 15.1 Handle Random Read Errors

```go
import (
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "log/slog"
)

// Package-level logger (set during init)
var log = slog.Default()

func generateTokenID() string {
    b := make([]byte, 16)
    n, err := rand.Read(b)
    if err != nil {
        // Log the error but don't expose it
        log.Error("crypto/rand read failed", "err", err)
        
        // Fall back to a less secure but non-zero result
        // This should never happen in practice with crypto/rand
        // Use current time + random bytes as fallback
        fallback(b)
    }
    
    if n != len(b) {
        log.Warn("crypto/rand returned unexpected length", "expected", len(b), "got", n)
        if n < len(b) {
            fallback(b)
        }
    }
    
    return base64.RawURLEncoding.EncodeToString(b)
}

func fallback(b []byte) {
    // Fill with timestamp + some randomness from system
    // This is NOT cryptographically secure but better than zeros
    for i := range b {
        b[i] = byte(i * 17 + 42)  // Pseudo-random pattern
    }
}
```

#### 15.2 Use crypto/rand More Consistently

Check all token generation in the codebase:

```bash
grep -rn "rand.Read\|math/rand" --include="*.go" | grep -v "_test.go"
```

Ensure all cryptographic random generation uses `crypto/rand` with proper error handling.

### Verification Checklist

- [ ] `generateTokenID` handles errors properly
- [ ] All random generation uses `crypto/rand`
- [ ] Errors are logged, not silently ignored
- [ ] Fallback produces non-zero values

---

## Implementation Priority

| Priority | Issue | Estimated Effort |
|----------|-------|-----------------|
| 🔴 P0 | WebSocket Origin Validation | Medium |
| 🔴 P0 | Command Status Implementation | Medium |
| 🟠 P1 | Enforce HMAC Dashboard Setting | Medium |
| 🟠 P1 | Device Online Status Integration | Small |
| 🟠 P1 | HMAC Window Configuration | Small |
| 🟡 P2 | Command Secrets Hash Column | Medium |
| 🟡 P2 | JWT Secret Implementation | Medium |
| 🟡 P2 | Google OAuth Signature Verification | Small |
| 🟢 P3 | WebSocket Handler Cleanup | Small |
| 🟢 P3 | Duplicate Registration Handlers | Small |
| 🟢 P3 | Rate Limiter Cleanup | Small |
| 🟢 P3 | Retry/Cancel Command | Medium |
| 🟢 P3 | Token ID Generation | Small |
| 🟢 P3 | Unused postJSON | Trivial |
| 🟢 P3 | Error Response Consistency | Medium |

---

## Testing Requirements

For each fix, add tests covering:

1. **WebSocket Origin**: Test allowed/blocked origins
2. **Device Online**: Test hub integration
3. **HMAC Enforcement**: Test both modes
4. **Command Status**: Test all status transitions
5. **Retry/Cancel**: Test edge cases

Run tests after each fix:
```bash
go test ./... -v -race
```

---

*Document Version: 1.0*
*Last Updated: 2026-06-06*