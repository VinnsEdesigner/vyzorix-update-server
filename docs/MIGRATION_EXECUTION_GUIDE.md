# Vyzorix Update Server - Complete Migration Execution Guide

> **Document Version:** 1.0  
> **Date:** 2026-06-17  
> **Status:** Ready for Execution  
> **Approach:** Strangler Fig Pattern - Domain-by-Domain Migration

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current State Analysis](#current-state-analysis)
3. [Migration Strategy](#migration-strategy)
4. [Phase 0: Foundation Completion](#phase-0-foundation-completion)
5. [Phase 1: Commands Domain Migration](#phase-1-commands-domain-migration)
6. [Phase 2: Devices Domain Migration](#phase-2-devices-domain-migration)
7. [Phase 3: Auth Domain Migration](#phase-3-auth-domain-migration)
8. [Phase 4: Cleanup](#phase-4-cleanup)
9. [File Inventory](#file-inventory)
10. [Success Criteria](#success-criteria)

---

## Executive Summary

### The Problem

The new clean architecture structure was **designed correctly** but **implemented incompletely**:

| Component | Status | Issue |
|-----------|--------|-------|
| Directory Structure |  Complete | - |
| Repository Interfaces |  Complete | - |
| Domain Entities |  Complete | - |
| Infrastructure Layer |  Partial | Incomplete implementations |
| Application Services |  Partial | Missing methods, not wired up |
| API Handlers |  Partial | New handlers exist but old handlers still have working logic |
| Main.go Wiring |  Incomplete | CommandService never instantiated |

### The Solution

**Strangler Fig Pattern**: Migrate domain-by-domain, switching routes only when new implementation is 100% verified working.

---

## Current State Analysis

### Directory Structure (NEW - Correct)

```
apps/api/
 cmd/api/main.go                           Entry point
 internal/
    api/
       server.go                        Router
       health_handler.go                Health
       handlers/
          auth/                        NEW handlers (partial)
             login.go
             register.go
             logout.go
             me.go
             settings.go
             admin.go
             mfa.go
             oauth.go
             email_verify.go
             password_reset.go
             client_credentials.go
             routes.go
          device/                      NEW handlers (partial)
             register.go
             status.go
             updater.go
          command/                    NEW handlers (partial)
             execute.go
          websocket/
             handler.go
          admin/
             clients.go
          admin_clients.go            OLD - needs migration
          auth_core.go                OLD - needs migration
          auth_email_verify.go         OLD - needs migration
          auth_mfa.go                 OLD - needs migration
          auth_oauth.go                OLD - needs migration
          auth_password_reset.go       OLD - needs migration
          auth_settings.go            OLD - needs migration
          auth_admin.go               OLD - needs migration
          auth_csrf.go                OLD - needs migration
          auth_rate_limit.go          OLD - needs migration
          auth_utils.go               OLD - needs migration
          auth_test.go                OLD - needs migration
          auth_mfa_test.go           OLD - needs migration
          client_credentials.go        OLD - needs migration
          command.go                   OLD - needs migration
          command_test.go             OLD - needs migration
          device.go                   OLD - needs migration
          device_test.go             OLD - needs migration
          health.go                  OLD - needs migration
          health_handler_test.go     OLD - needs migration
          health_test.go             OLD - needs migration
          lockout.go                  OLD - needs migration
          lockout_test.go            OLD - needs migration
          rate_limit_test.go         OLD - needs migration
          server.go                   OLD - needs migration
          updater.go                 OLD - needs migration
          websocket_handler.go        OLD - needs migration
          websocket_handler_test.go  OLD - needs migration
       middleware/                    Middleware (complete)
    application/                       Services (partial)
       auth/
          service.go                Missing session management
          password.go
       device/
          service.go                Missing LastSeen, Count, pagination
       command/
          service.go                NOT WIRED in main.go!
       client/
          service.go
       dto/
          auth.go
          device.go
       shared/
          token.go
       errors.go
    domain/                            Domain (complete)
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
       client/
          entity.go
          repository.go
       email_verification/
          entity.go
       password_reset/
          entity.go
       auth/
           enum_safe.go
    infrastructure/                    Infrastructure (partial)
       storage/
          sqlite.go
          operator.go
          device.go
          session.go
          command.go
          client.go
          email_verification.go
          password_reset.go
       crypto/
          aes_gcm.go
          replay_cache.go
       auth/
          argon2_hasher.go
       email/
           service.go
    ws/
       hub.go
       client.go
    fcm/
       fcm.go
       notifier.go
    auth/
       jwt.go
       session.go
       lockout.go
       password.go
       totp.go
       origin.go
       ...
    email.go
    email_test.go
    command_signer.go
    command_signer_test.go
    audit/
    metrics/
    ssr/
 pkg/                                    Legacy (to be cleaned)
    storage/                           DUPLICATE - old storage
       store.go
       operators.go
       devices.go
       sessions.go
       commands.go
       clients.go
       settings.go
       migrations.go
       crypto.go
       uuid.go
       telemetry.go
    models/                           DUPLICATE - old models
       auth.go
       device.go
       command.go
       response.go
       telemetry.go
       updater.go
    crypto/
       hmac.go
    config/
       ...
    logging/
        ...
 public/                                Static assets
```

---

## Migration Strategy

### Strangler Fig Pattern

1. **Phase 0**: Complete Foundation - Wire what's started, add missing pieces
2. **Phase 1**: Commands Domain - Full migration, switch routes, verify
3. **Phase 2**: Devices Domain - Full migration, switch routes, verify
4. **Phase 3**: Auth Domain - Full migration, switch routes, verify
5. **Phase 4**: Cleanup - Delete old handlers, remove duplicates

### Key Principles

| Principle | Description |
|-----------|-------------|
| **No Delete Until Verified** | Old handlers remain until new is 100% working |
| **Domain-by-Domain** | Complete one domain before starting next |
| **Test Thoroughly** | Each domain switch requires verification |
| **Working Reference** | Previous migrations guide next ones |
| **Rollback Capable** | Can revert to old handler if issues found |

---

## Phase 0: Foundation Completion

### Goal
Make the new structure actually usable. Wire up what's been created but not connected.

### Tasks

#### Task 0.1: Wire CommandService in main.go

**File**: `cmd/api/main.go`

**Problem**: `CommandService` exists in `internal/application/command/service.go` but is **NEVER instantiated** in `main.go`.

**Current main.go creates**:
```go
// Created:
authService        
deviceService      
clientService     
emailSvc          
fcmNotifier       
wsHub             

// NOT Created:
commandService     MISSING
```

**Action Required**:
1. Add `commandRepo` instantiation
2. Create `CommandService` with dependencies
3. Pass to `NewServer()`

**New code to add in main.go**:
```go
// After device service creation (~line 168)
// Command Repository (for command service)
commandRepo := storage.NewCommandRepository(db.DB())
printStatus("CommandRepository", "Initialized", false)

// After device service (~line 168)
commandService := command.NewService(commandRepo, deviceRepo)
printStatus("CommandService", "Initialized", false)
```

**Files to modify**: `cmd/api/main.go`

---

#### Task 0.2: Complete DeviceService Methods

**File**: `internal/application/device/service.go`

**Problem**: Domain repository has methods that service doesn't expose.

**Missing methods to add**:

```go
// Add to DeviceService:

// GetDevice retrieves a device by ID (should already exist, verify)
func (s *Service) GetDevice(ctx context.Context, deviceID string) (*device.Device, error) {
    return s.deviceRepo.FindByID(ctx, deviceID)
}

// UpdateLastSeen updates the last seen timestamp
func (s *Service) UpdateLastSeen(ctx context.Context, deviceID string) error {
    return s.deviceRepo.UpdateLastSeen(ctx, deviceID)
}

// Count returns total device count
func (s *Service) Count(ctx context.Context) (int, error) {
    return s.deviceRepo.Count(ctx)
}

// CountByOperator returns device count for operator
func (s *Service) CountByOperator(ctx context.Context, operatorID string) (int, error) {
    return s.deviceRepo.CountByOperator(ctx, operatorID)
}
```

**Files to modify**: `internal/application/device/service.go`

---

#### Task 0.3: Complete AuthService Session Management

**File**: `internal/application/auth/service.go`

**Problem**: Session creation happens in handler, not service.

**Missing methods to add**:

```go
// Add to AuthService:

// CreateSession creates a new session for an operator
func (s *AuthService) CreateSession(ctx context.Context, operatorID string) (*session.Session, error) {
    id, err := shared.GenerateID()
    if err != nil {
        return nil, err
    }
    
    now := time.Now()
    sess := &session.Session{
        ID:         id,
        OperatorID: operatorID,
        CreatedAt:  now,
        ExpiresAt: now.Add(s.sessionTTL),
    }
    
    if err := s.sessionRepo.Create(ctx, sess); err != nil {
        return nil, err
    }
    
    return sess, nil
}

// DeleteSession deletes a session
func (s *AuthService) DeleteSession(ctx context.Context, sessionID string) error {
    return s.sessionRepo.Delete(ctx, sessionID)
}

// DeleteAllOperatorSessions deletes all sessions for an operator
func (s *AuthService) DeleteAllOperatorSessions(ctx context.Context, operatorID string) error {
    return s.sessionRepo.DeleteByOperatorID(ctx, operatorID)
}
```

**Files to modify**: `internal/application/auth/service.go`

---

#### Task 0.4: Add Missing Command Endpoints

**File**: `internal/api/handlers/command/execute.go`

**Problem**: Only `ExecuteCommand` implemented. Missing status, retry, pending, cancel.

**Add to `execute.go`**:

```go
// Add after Handle method (~line 113):

// GetStatus handles GET /v1/command/:dispatchId/status
func (h *ExecuteHandler) GetStatus(c *gin.Context) {
    dispatchID := c.Param("dispatchId")
    if dispatchID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "dispatchId required"})
        return
    }
    
    // Use CommandService if wired, otherwise return not implemented
    // For now, return placeholder until service is wired
    c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented", "message": "use CommandService"})
}

// Retry handles POST /v1/command/:dispatchId/retry  
func (h *ExecuteHandler) Retry(c *gin.Context) {
    dispatchID := c.Param("dispatchId")
    if dispatchID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "dispatchId required"})
        return
    }
    
    c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented", "message": "use CommandService"})
}

// GetPending handles GET /v1/device/:id/commands/pending
func (h *ExecuteHandler) GetPending(c *gin.Context) {
    deviceID := c.Param("id")
    if deviceID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "deviceId required"})
        return
    }
    
    c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented", "message": "use CommandService"})
}

// Cancel handles DELETE /v1/command/:dispatchId
func (h *ExecuteHandler) Cancel(c *gin.Context) {
    dispatchID := c.Param("dispatchId")
    if dispatchID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "dispatchId required"})
        return
    }
    
    c.JSON(http.StatusNotImplemented, gin.H{"error": "not_implemented", "message": "use CommandService"})
}
```

**Files to modify**: `internal/api/handlers/command/execute.go`

---

#### Task 0.5: Add FCM Integration to CommandService

**File**: `internal/application/command/service.go`

**Problem**: `CommandService` doesn't integrate with WebSocket hub or FCM notifier.

**Required modification**:

```go
// Service needs hub and notifier dependencies:

type Service struct {
    commandRepo command.Repository
    deviceRepo device.Repository
    hub        *hub.Hub          // ADD
    fcmNotifier fcm.Notifier    // ADD
}

func NewService(commandRepo command.Repository, deviceRepo device.Repository, hub *hub.Hub, fcmNotifier fcm.Notifier) *Service {
    return &Service{
        commandRepo: commandRepo,
        deviceRepo: deviceRepo,
        hub: hub,
        fcmNotifier: fcmNotifier,
    }
}
```

Then modify `SendCommand` to:
1. Check if device online via hub
2. Send via WebSocket if online
3. Queue via FCM if offline
4. Persist command record

**Files to modify**: `internal/application/command/service.go`

---

### Phase 0 Deliverables

| Task | File | Status |
|------|------|--------|
| 0.1 | Wire CommandService in main.go |  TODO |
| 0.2 | Complete DeviceService methods |  TODO |
| 0.3 | Complete AuthService session |  TODO |
| 0.4 | Add missing command endpoints |  TODO |
| 0.5 | Add FCM/Hub to CommandService |  TODO |

**Phase 0 is complete when**:
- `CommandService` is wired in main.go
- All service methods exist (even if returning NotImplemented)
- Code compiles without errors

---

## Phase 1: Commands Domain Migration

### Goal
Full command functionality in new structure. Switch routes when verified.

### Current Commands Implementation

**OLD handler**: `internal/api/handlers/command.go` - 221 lines
**NEW handler**: `internal/api/handlers/command/execute.go` - 114 lines (incomplete)

**OLD has endpoints**:
- `POST /v1/device/:id/command` - SendCommand 
- `GET /v1/command/:dispatchId/status` - GetCommandStatus 
- `POST /v1/command/:dispatchId/retry` - RetryCommand 
- `GET /v1/device/:id/commands/pending` - GetPendingCommands 
- `DELETE /v1/command/:dispatchId` - CancelCommand 

### Tasks

#### Task 1.1: Complete CommandService WebSocket/FCM Integration

**File**: `internal/application/command/service.go`

Implement full `SendCommand` method:

```go
// SendCommand sends a command to a device via WebSocket or FCM
func (s *Service) SendCommand(ctx context.Context, req *dto.SendCommandRequest) (*dto.SendCommandResponse, error) {
    // 1. Check device exists
    device, err := s.deviceRepo.FindByID(ctx, req.DeviceID)
    if err != nil {
        if err == device.ErrNotFound {
            return nil, application.ErrDeviceNotFound
        }
        return nil, err
    }
    
    // 2. Generate dispatch ID
    dispatchID := generateDispatchID()
    
    // 3. Build command frame
    frame := buildCommandFrame(dispatchID, req)
    
    // 4. Determine delivery method
    delivery := "queued"
    
    // 5. Try WebSocket first
    if s.hub != nil && s.hub.Online(req.DeviceID) {
        if s.hub.Send(req.DeviceID, frame) {
            delivery = "sent"
        }
    }
    
    // 6. If offline and has FCM, send wake
    if delivery == "queued" && s.fcmNotifier != nil && device.FCMToken != "" {
        wake := fcm.SilentWake{
            Token:      device.FCMToken,
            Command:    req.Command,
            DispatchID: dispatchID,
            DeviceID:   req.DeviceID,
        }
        if err := s.fcmNotifier.SendSilentWake(ctx, wake); err != nil {
            // Log but don't fail
        } else {
            delivery = "queued_fcm"
        }
    }
    
    // 7. Persist command
    cmd := &command.Command{
        ID:         generateID(),
        DeviceID:   req.DeviceID,
        DispatchID: dispatchID,
        Command:    req.Command,
        Status:     command.StatusPending,
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
    }
    if req.Args != nil {
        cmd.SetArgs(req.Args)
    }
    
    if err := s.commandRepo.Create(ctx, cmd); err != nil {
        return nil, err
    }
    
    return &dto.SendCommandResponse{
        DispatchID: dispatchID,
        Status:     delivery,
        DeviceID:   req.DeviceID,
        CreatedAt:  cmd.CreatedAt,
    }, nil
}
```

**Files to modify**: `internal/application/command/service.go`

---

#### Task 1.2: Implement Command Status Endpoint

**File**: `internal/api/handlers/command/execute.go`

Add to `ExecuteHandler`:

```go
// GetStatus handles GET /v1/command/:dispatchId/status
func (h *ExecuteHandler) GetStatus(c *gin.Context) {
    dispatchID := c.Param("dispatchId")
    deviceID := c.Query("deviceId") // Required for idempotency check
    
    if dispatchID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "dispatchId required"})
        return
    }
    
    if deviceID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "deviceId required"})
        return
    }
    
    // Use CommandService
    status, err := h.commandService.GetStatus(c.Request.Context(), deviceID, dispatchID)
    if err != nil {
        if err == command.ErrNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "command not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, status)
}
```

**Note**: `CommandService` needs `GetStatus` method added.

**Files to modify**: 
- `internal/application/command/service.go`
- `internal/api/handlers/command/execute.go`

---

#### Task 1.3: Implement Command Retry Endpoint

**File**: `internal/api/handlers/command/execute.go`

```go
// Retry handles POST /v1/command/:dispatchId/retry
func (h *ExecuteHandler) Retry(c *gin.Context) {
    dispatchID := c.Param("dispatchId")
    
    if dispatchID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "dispatchId required"})
        return
    }
    
    status, err := h.commandService.RetryCommand(c.Request.Context(), dispatchID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, status)
}
```

**Files to modify**: 
- `internal/application/command/service.go`
- `internal/api/handlers/command/execute.go`

---

#### Task 1.4: Implement Pending Commands Endpoint

**File**: `internal/api/handlers/command/execute.go`

```go
// GetPending handles GET /v1/device/:id/commands/pending
func (h *ExecuteHandler) GetPending(c *gin.Context) {
    deviceID := c.Param("id")
    
    if deviceID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "deviceId required"})
        return
    }
    
    commands, err := h.commandService.GetPendingCommands(c.Request.Context(), deviceID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"commands": commands})
}
```

**Files to modify**: 
- `internal/application/command/service.go`
- `internal/api/handlers/command/execute.go`

---

#### Task 1.5: Implement Command Cancel Endpoint

**File**: `internal/api/handlers/command/execute.go`

```go
// Cancel handles DELETE /v1/command/:dispatchId
func (h *ExecuteHandler) Cancel(c *gin.Context) {
    dispatchID := c.Param("dispatchId")
    
    if dispatchID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "dispatchId required"})
        return
    }
    
    err := h.commandService.CancelCommand(c.Request.Context(), dispatchID)
    if err != nil {
        if err == command.ErrNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "command not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"cancelled": true, "dispatchId": dispatchID})
}
```

**Files to modify**: 
- `internal/application/command/service.go`
- `internal/api/handlers/command/execute.go`

---

#### Task 1.6: Wire CommandService to Handler

**File**: `internal/api/server.go`

Update `NewServer` and route registration:

```go
// In ServerConfig, add:
type ServerConfig struct {
    // ... existing fields ...
    CommandService *command.Service  // ADD
}

// In NewServer, create handler with service:
commandHandler := command.NewExecuteHandler(cfg.CommandService, cfg.Hub, cfg.FCMNotifier)

// In setupRoutes, add new routes:
deviceMgmt.POST("/:id/command", s.commandHandler.Handle)
deviceMgmt.GET("/command/:dispatchId/status", s.commandHandler.GetStatus)
deviceMgmt.POST("/command/:dispatchId/retry", s.commandHandler.Retry)
deviceMgmt.GET("/device/:id/commands/pending", s.commandHandler.GetPending)
deviceMgmt.DELETE("/command/:dispatchId", s.commandHandler.Cancel)
```

**Files to modify**: `internal/api/server.go`

---

#### Task 1.7: Update main.go CommandService Wiring

**File**: `cmd/api/main.go`

```go
// Add CommandService creation after deviceService:

commandService := command.NewService(
    storage.NewCommandRepository(db.DB()),
    deviceRepo,
    wsHub,           // Pass hub for WebSocket send
    fcmNotifier,    // Pass FCM notifier
)
printStatus("CommandService", "Initialized", false)

// Pass to api.Server:
apiServer := api.NewServer(&api.ServerConfig{
    // ... existing fields ...
    CommandService: commandService,
})
```

**Files to modify**: `cmd/api/main.go`

---

#### Task 1.8: Test Commands Domain

**Verification Steps**:

1. Start server with new code
2. Test each endpoint manually:
   - `POST /v1/device/:id/command` - Send command
   - `GET /v1/command/:dispatchId/status` - Check status
   - `POST /v1/command/:dispatchId/retry` - Retry command
   - `GET /v1/device/:id/commands/pending` - List pending
   - `DELETE /v1/command/:dispatchId` - Cancel command
3. Verify WebSocket delivery works
4. Verify FCM fallback works
5. Verify persistence works

**Phase 1 is complete when**: All 5 command endpoints work via new handler.

---

## Phase 2: Devices Domain Migration

### Goal
Full device functionality in new structure. Switch routes when verified.

### Current Devices Implementation

**OLD handler**: `internal/api/handlers/device.go` - 252 lines
**NEW handler directory**: `internal/api/handlers/device/` - 3 files (incomplete)

**OLD has endpoints**:
- `POST /v1/device/register` - Register  (new exists)
- `GET /v1/device/:id/status` - Status  (new exists)
- `PATCH /v1/device/:id/fcm-token` - UpdateFCMToken  (new exists)
- `DELETE /v1/device/:id` - Delete  (new exists)
- `GET /v1/dashboard/devices` - List with pagination  MISSING in new
- `GET /v1/device/:id` - GetDevice  MISSING in new
- `GET /v1/device/count` - Count  MISSING in new

### Tasks

#### Task 2.1: Add DeviceService.GetDevice

**File**: `internal/application/device/service.go`

```go
// GetDevice retrieves a device by ID
func (s *Service) GetDevice(ctx context.Context, deviceID string) (*device.Device, error) {
    return s.deviceRepo.FindByID(ctx, deviceID)
}
```

**Files to modify**: `internal/application/device/service.go`

---

#### Task 2.2: Add DeviceService.List with Pagination

**File**: `internal/application/device/service.go`

```go
// ListRequest holds pagination params
type ListRequest struct {
    Limit   int
    Cursor  int64  // LastSeen timestamp for cursor
    Online  *bool  // Filter by online status
}

// ListResponse holds paginated response
type ListResponse struct {
    Devices    []DeviceResponse
    NextCursor int64
    Total      int
}

// List returns devices with cursor-based pagination
func (s *Service) List(ctx context.Context, req *ListRequest) (*ListResponse, error) {
    if req.Limit <= 0 {
        req.Limit = 50
    }
    if req.Limit > 100 {
        req.Limit = 100
    }
    
    // Query one extra to check if there are more
    devices, err := s.deviceRepo.List(ctx, req.Limit+1, 0)
    if err != nil {
        return nil, err
    }
    
    response := &ListResponse{
        Devices: make([]DeviceResponse, 0),
    }
    
    hasMore := len(devices) > req.Limit
    if hasMore {
        devices = devices[:req.Limit]
    }
    
    for _, d := range devices {
        // Apply cursor filter if provided
        if req.Cursor > 0 && d.LastSeen >= req.Cursor {
            continue
        }
        
        // Apply online filter if provided
        if req.Online != nil && d.Online != *req.Online {
            continue
        }
        
        response.Devices = append(response.Devices, DeviceResponse{
            ID:                d.ID,
            FirebaseInstallID: d.FirebaseInstallID,
            AppVersion:       d.AppVersion,
            DeviceClass:      d.DeviceClass,
            Online:           d.Online,
            LastSeen:         d.LastSeen,
        })
    }
    
    // Set next cursor
    if hasMore && len(response.Devices) > 0 {
        response.NextCursor = response.Devices[len(response.Devices)-1].LastSeen
    }
    
    // Get total count
    total, err := s.deviceRepo.Count(ctx)
    if err != nil {
        return nil, err
    }
    response.Total = total
    
    return response, nil
}
```

**Files to modify**: `internal/application/device/service.go`

---

#### Task 2.3: Add DeviceService.Count

**File**: `internal/application/device/service.go`

```go
// Count returns total device count
func (s *Service) Count(ctx context.Context) (int, error) {
    return s.deviceRepo.Count(ctx)
}
```

**Files to modify**: `internal/application/device/service.go`

---

#### Task 2.4: Add List Handler Endpoint

**File**: `internal/api/handlers/device/list.go` (CREATE NEW)

```go
package device

import (
    "net/http"
    "strconv"
    
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
    "github.com/gin-gonic/gin"
)

// ListHandler handles device listing with pagination
type ListHandler struct {
    deviceService *device.Service
}

// NewListHandler creates a new ListHandler
func NewListHandler(deviceService *device.Service) *ListHandler {
    return &ListHandler{deviceService: deviceService}
}

// Handle handles GET /v1/dashboard/devices
func (h *ListHandler) Handle(c *gin.Context) {
    // Parse limit
    limit := 50
    if l := c.Query("limit"); l != "" {
        if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
            limit = parsed
        }
    }
    
    // Parse cursor
    var cursor int64
    if cur := c.Query("cursor"); cur != "" {
        if parsed, err := strconv.ParseInt(cur, 10, 64); err == nil {
            cursor = parsed
        }
    }
    
    // Parse online filter
    var online *bool
    if o := c.Query("online"); o != "" {
        v := o == "true"
        online = &v
    }
    
    req := &device.ListRequest{
        Limit:  limit,
        Cursor: cursor,
        Online: online,
    }
    
    result, err := h.deviceService.List(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
        return
    }
    
    response := gin.H{"devices": result.Devices}
    if result.NextCursor > 0 {
        response["nextCursor"] = result.NextCursor
    }
    response["total"] = result.Total
    
    c.JSON(http.StatusOK, response)
}
```

**Files to create**: `internal/api/handlers/device/list.go`

---

#### Task 2.5: Wire DeviceService to All Handlers

**File**: `internal/api/server.go`

```go
// In setupRoutes, update device routes:
deviceGroup := r.Group("/device")
{
    deviceGroup.GET("/dashboard/devices", s.deviceListHandler.Handle)
    // ... existing routes
}

// Add new handler to Server struct
deviceListHandler *device.ListHandler

// In NewServer, create handler:
deviceListHandler := device.NewListHandler(cfg.DeviceService)
```

**Files to modify**: `internal/api/server.go`

---

#### Task 2.6: Test Devices Domain

**Verification Steps**:

1. Test all device endpoints:
   - `POST /v1/device/register` - Register device
   - `GET /v1/device/:id/status` - Get status
   - `GET /v1/device/:id` - Get device (new)
   - `PATCH /v1/device/:id/fcm-token` - Update FCM
   - `DELETE /v1/device/:id` - Delete device
   - `GET /v1/dashboard/devices` - List with pagination (new)
   - `GET /v1/device/count` - Count (new)

**Phase 2 is complete when**: All device endpoints work via new structure.

---

## Phase 3: Auth Domain Migration

### Goal
Full auth functionality in new structure. This is the most complex domain.

### Current Auth Implementation

**OLD handlers** (to be migrated):
- `auth_core.go` - Login, Register, Me, Logout
- `auth_email_verify.go` - Email verification
- `auth_password_reset.go` - Password reset
- `auth_mfa.go` - MFA
- `auth_oauth.go` - OAuth
- `auth_settings.go` - Settings
- `auth_admin.go` - Admin operations

**NEW handlers** (partially complete):
- `auth/login.go` 
- `auth/register.go` 
- `auth/logout.go` 
- `auth/me.go` 
- `auth/email_verify.go`  (needs completion)
- `auth/password_reset.go`  (needs completion)
- `auth/mfa.go`  (needs completion)
- `auth/oauth.go`  (needs completion)
- `auth/settings.go`  (needs completion)
- `auth/admin.go`  (ListOperators only)

### Tasks

#### Task 3.1: Complete AuthService Session Methods

**File**: `internal/application/auth/service.go`

Already partially done in Phase 0.3. Ensure these exist:

```go
// VerifyPassword verifies operator password
func (s *AuthService) VerifyPassword(ctx context.Context, email, password string) (*operator.Operator, error) {
    op, err := s.operatorRepo.FindByEmail(ctx, email)
    if err != nil {
        if err == operator.ErrNotFound {
            // Run dummy verification to prevent timing attacks
            _ = s.passwordHasher.Verify(password, "$argon2id$v=19$m=65536,t=3,p=4$YWRkcmVzc2FsdA$ZmFrZWhhc2g=")
            return nil, application.ErrInvalidCredentials
        }
        return nil, err
    }
    
    if err := s.passwordHasher.Verify(password, op.PasswordHash); err != nil {
        return nil, application.ErrInvalidCredentials
    }
    
    return op, nil
}

// CreateSessionWithExpiry creates session and returns cookie value
func (s *AuthService) CreateSessionWithExpiry(ctx context.Context, operatorID string) (string, time.Time, error) {
    sess, err := s.CreateSession(ctx, operatorID)
    if err != nil {
        return "", time.Time{}, err
    }
    return sess.ID, sess.ExpiresAt, nil
}
```

**Files to modify**: `internal/application/auth/service.go`

---

#### Task 3.2: Complete Email Verification Flow

**File**: `internal/api/handlers/auth/email_verify.go`

Ensure all endpoints implemented:
- `POST /v1/auth/verify-email` - Verify email
- `POST /v1/auth/resend-verification` - Resend verification
- `POST /v1/auth/cancel-verification` - Cancel verification
- `GET /v1/auth/poll-verification` - Poll status

**Files to modify**: `internal/api/handlers/auth/email_verify.go`
**Files to add**: Methods to `internal/application/auth/service.go`

---

#### Task 3.3: Complete Password Reset Flow

**File**: `internal/api/handlers/auth/password_reset.go`

Ensure all endpoints implemented:
- `POST /v1/auth/forgot-password` - Request reset
- `POST /v1/auth/reset-password` - Reset password
- `POST /v1/auth/resend-password-reset` - Resend reset

**Files to modify**: `internal/api/handlers/auth/password_reset.go`
**Files to add**: Methods to `internal/application/auth/service.go`

---

#### Task 3.4: Complete MFA Flow

**File**: `internal/api/handlers/auth/mfa.go`

Ensure all endpoints implemented:
- `GET /v1/auth/mfa/status` - Get MFA status
- `POST /v1/auth/mfa/enroll` - Start enrollment
- `POST /v1/auth/mfa/verify-setup` - Verify setup
- `POST /v1/auth/mfa/enable` - Enable MFA
- `POST /v1/auth/mfa/disable` - Disable MFA
- `POST /v1/auth/mfa/verify-backup` - Verify backup code
- `POST /v1/auth/mfa/regenerate-backup-codes` - Regenerate backup codes

**Files to modify**: `internal/api/handlers/auth/mfa.go`
**Files to add**: Methods to `internal/application/auth/service.go`

---

#### Task 3.5: Complete OAuth Flow

**File**: `internal/api/handlers/auth/oauth.go`

Ensure all endpoints implemented:
- `GET /v1/auth/google` - Google login redirect
- `GET /v1/auth/google/callback` - Google callback
- `GET /v1/auth/github` - GitHub login redirect
- `GET /v1/auth/github/callback` - GitHub callback

**Files to modify**: `internal/api/handlers/auth/oauth.go`
**Files to add**: Methods to `internal/application/auth/service.go`

---

#### Task 3.6: Complete Settings Flow

**File**: `internal/api/handlers/auth/settings.go`

Ensure all endpoints implemented:
- `PATCH /v1/auth/me` - Update name
- `PATCH /v1/auth/me/settings` - Update settings

Add full operator update to `AuthService`:

```go
// UpdateOperator updates operator fields
func (s *AuthService) UpdateOperator(ctx context.Context, operatorID string, req *UpdateOperatorRequest) (*operator.Operator, error) {
    op, err := s.operatorRepo.FindByID(ctx, operatorID)
    if err != nil {
        return nil, err
    }
    
    if req.Name != nil {
        op.Name = *req.Name
    }
    if req.Thresholds != nil {
        op.Thresholds = *req.Thresholds
    }
    if req.ClientSettings != nil {
        op.ClientSettings = *req.ClientSettings
    }
    
    op.UpdatedAt = time.Now()
    
    if err := s.operatorRepo.Update(ctx, op); err != nil {
        return nil, err
    }
    
    return op, nil
}
```

**Files to modify**: 
- `internal/api/handlers/auth/settings.go`
- `internal/application/auth/service.go`

---

#### Task 3.7: Complete Admin Operations

**File**: `internal/api/handlers/auth/admin.go`

Currently has `ListOperators`. Add:
- `POST /v1/auth/admin/operators` - Create operator
- `PATCH /v1/auth/admin/operators/:id` - Update operator
- `DELETE /v1/auth/admin/operators/:id` - Delete operator

**Files to modify**: `internal/api/handlers/auth/admin.go`
**Files to add**: Methods to `internal/application/auth/service.go`

---

#### Task 3.8: Test Auth Domain

**Verification Steps**:

1. Test registration flow
2. Test login flow
3. Test email verification flow
4. Test password reset flow
5. Test MFA enrollment and usage
6. Test OAuth (Google, GitHub)
7. Test settings update
8. Test admin operations

**Phase 3 is complete when**: All auth endpoints work via new structure.

---

## Phase 4: Cleanup

### Goal
Remove duplicate/old code once everything is verified working.

### Tasks

#### Task 4.1: Delete Old Handler Files

**Files to DELETE** (after verifying new works):

```
internal/api/handlers/
 admin_clients.go           # Replaced by handlers/admin/clients.go
 auth_core.go              # Replaced by handlers/auth/
 auth_email_verify.go      # Replaced by handlers/auth/email_verify.go
 auth_mfa.go               # Replaced by handlers/auth/mfa.go
 auth_oauth.go             # Replaced by handlers/auth/oauth.go
 auth_password_reset.go    # Replaced by handlers/auth/password_reset.go
 auth_settings.go          # Replaced by handlers/auth/settings.go
 auth_admin.go             # Replaced by handlers/auth/admin.go
 auth_csrf.go              # Logic moved to middleware
 auth_rate_limit.go        # Logic moved to middleware
 auth_utils.go             # Utility functions distributed
 auth_test.go              # Tests moved to new locations
 auth_mfa_test.go          # Tests moved to new locations
 client_credentials.go     # Replaced by handlers/auth/client_credentials.go
 command.go                # Replaced by handlers/command/
 command_test.go           # Tests moved to new location
 device.go                 # Replaced by handlers/device/
 device_test.go           # Tests moved to new location
 health.go                 # Replaced by health_handler.go
 health_handler_test.go   # Tests moved to new location
 health_test.go          # Tests moved to new location
 lockout.go               # Logic moved to middleware/auth/lockout.go
 lockout_test.go         # Tests moved
 rate_limit_test.go      # Tests moved
 server.go                # Logic moved to server.go (api package)
 updater.go              # Logic moved to handlers/device/updater.go
 websocket_handler.go     # Replaced by handlers/websocket/handler.go
 websocket_handler_test.go # Tests moved
```

---

#### Task 4.2: Delete Old Storage Files

**Files to DELETE** (after verifying new works):

```
pkg/storage/
 clients.go              # Replaced by internal/infrastructure/storage/client.go
 commands.go             # Replaced by internal/infrastructure/storage/command.go
 crypto.go              # Replaced by internal/infrastructure/crypto/
 devices.go              # Replaced by internal/infrastructure/storage/device.go
 migrations.go           # Keep or move to internal/infrastructure/storage/migrations/
 operators.go            # Replaced by internal/infrastructure/storage/operator.go
 sessions.go             # Replaced by internal/infrastructure/storage/session.go
 settings.go             # Replaced by domain repositories
 store.go               # Replaced by internal/infrastructure/storage/sqlite.go
 telemetry.go            # May not be needed
 uuid.go                # Replaced by internal/application/shared/

# ALSO DELETE pkg/models/ (all files) - replaced by internal/domain/
```

---

#### Task 4.3: Update Imports

After deleting old files, ensure all imports are updated:

```bash
# Run to find broken imports:
go build ./...

# Fix any import errors
```

---

#### Task 4.4: Run Full Test Suite

```bash
go test ./...
go build ./...
```

---

## File Inventory

### Files to CREATE (New)

| File | Purpose | Phase |
|------|---------|-------|
| `internal/api/handlers/device/list.go` | Device list with pagination | 2 |
| `internal/application/command/service.go` methods | Complete command methods | 1 |

### Files to MODIFY

| File | Changes | Phase |
|------|---------|-------|
| `cmd/api/main.go` | Wire CommandService | 0, 1 |
| `internal/application/device/service.go` | Add methods | 0, 2 |
| `internal/application/auth/service.go` | Add session methods | 0, 3 |
| `internal/api/handlers/command/execute.go` | Add endpoints | 0, 1 |
| `internal/api/handlers/auth/email_verify.go` | Complete implementation | 3 |
| `internal/api/handlers/auth/password_reset.go` | Complete implementation | 3 |
| `internal/api/handlers/auth/mfa.go` | Complete implementation | 3 |
| `internal/api/handlers/auth/oauth.go` | Complete implementation | 3 |
| `internal/api/handlers/auth/settings.go` | Complete implementation | 3 |
| `internal/api/handlers/auth/admin.go` | Complete admin ops | 3 |
| `internal/api/server.go` | Wire all services | 0, 1, 2, 3 |

### Files to DELETE (After Verification)

| File | Replaced By | Phase |
|------|-------------|--------|
| All `internal/api/handlers/*.go` (old) | `internal/api/handlers/{auth,device,command}/` | 4 |
| `pkg/storage/*.go` (old) | `internal/infrastructure/storage/` | 4 |
| `pkg/models/*.go` | `internal/domain/` | 4 |

---

## Success Criteria

### Phase 0 Complete When:
- [ ] Code compiles without errors
- [ ] CommandService is wired in main.go
- [ ] All service methods exist (even if returning NotImplemented)

### Phase 1 Complete When:
- [ ] `POST /v1/device/:id/command` works via new handler
- [ ] `GET /v1/command/:dispatchId/status` works via new handler
- [ ] `POST /v1/command/:dispatchId/retry` works via new handler
- [ ] `GET /v1/device/:id/commands/pending` works via new handler
- [ ] `DELETE /v1/command/:dispatchId` works via new handler
- [ ] WebSocket delivery verified
- [ ] FCM fallback verified

### Phase 2 Complete When:
- [ ] `POST /v1/device/register` works via new handler
- [ ] `GET /v1/device/:id/status` works via new handler
- [ ] `GET /v1/device/:id` works via new handler
- [ ] `GET /v1/device/count` works via new handler
- [ ] `PATCH /v1/device/:id/fcm-token` works via new handler
- [ ] `DELETE /v1/device/:id` works via new handler
- [ ] `GET /v1/dashboard/devices` with pagination works via new handler

### Phase 3 Complete When:
- [ ] Login works via new handler
- [ ] Register works via new handler
- [ ] Email verification works via new handler
- [ ] Password reset works via new handler
- [ ] MFA works via new handler
- [ ] OAuth works via new handler
- [ ] Settings update works via new handler
- [ ] Admin operations work via new handler

### Phase 4 Complete When:
- [ ] All old handler files deleted
- [ ] All old storage files deleted
- [ ] All old model files deleted
- [ ] Code compiles without errors
- [ ] All tests pass
- [ ] No duplicate functionality remains

---

## Summary

| Phase | Duration | Focus |
|-------|----------|-------|
| 0 | 1-2 days | Foundation completion |
| 1 | 2-3 days | Commands domain |
| 2 | 2-3 days | Devices domain |
| 3 | 4-5 days | Auth domain |
| 4 | 1-2 days | Cleanup |
| **Total** | **10-15 days** | Full migration |

**Key Principle**: Don't rush. Verify each domain before moving to next. Old handlers remain as fallback until new is verified.

---

*Document prepared for step-by-step execution.*
