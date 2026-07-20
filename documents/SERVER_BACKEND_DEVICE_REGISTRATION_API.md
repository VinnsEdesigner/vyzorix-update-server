# Server Backend - Device Registration API Specification

> **Version:** 1.0
> **Status:** Draft
> **Created:** 2026-06-24
> **Target:** Production MVP
> **Go Package:** `github.com/VinnsEdesigner/vyzorix/apps/api`

---

## Table of Contents

1. Overview
2. Current State Analysis
3. Registration Flow
4. Required API Endpoints
5. Database Schema
6. Backend File Structure
7. Handler Specifications
8. Service Layer
9. GraphQL Schema
10. Error Handling
11. Rate Limiting & Security
12. File Changes Summary
13. Implementation Order

---

## 1. Overview

### 1.1 Purpose

This document maps out the server-side requirements to support the Device Registration system as specified in DEVICE_REGISTRATION_SYSTEM.md.

### 1.2 Frontend Requirements Summary

| Feature | Description | Required Endpoints |
|---------|-------------|-------------------|
| Device Inbox | Pending registration requests | GET /v1/device/inbox |
| Inbox Entry | Single request details | GET /v1/device/inbox/:imei |
| Acknowledge | Approve/reject registration | POST /v1/device/inbox/:imei/ack |
| Deregister | Remove device | DELETE /v1/devices/:imei |
| Register | Register new device | POST /v1/device/inbox (DEPRECATED: was /v1/device/register) |
| Confirm | Device confirms registration | POST /v1/device/confirm |
| Devices List | All registered devices | GET /v1/devices |
| Device Detail | Single device info | GET /v1/devices/:imei |

---

## 2. Current State Analysis

### 2.1 Existing Endpoints

| Endpoint | Status | Handler | Notes |
|----------|--------|---------|-------|
| POST /v1/device/inbox | EXISTS | InboxHandler.HandleInboxRequest | Device sends registration request |
| POST /v1/device/confirm | EXISTS | ConfirmHandler.Handle | Device confirms registration |
| GET /v1/device/:imei/status | EXISTS | StatusHandler.Handle | Get device status (public) |
| GET /v1/devices | EXISTS | DevicesHandler.GetDevices | List all registered devices |
| GET /v1/devices/:imei | EXISTS | DevicesHandler.GetDeviceDetail | Get single device info |
| DELETE /v1/devices/:imei | EXISTS | DevicesHandler.DeregisterDevice | Deregister device |
| GET /v1/device/inbox | EXISTS | InboxHandler.GetInbox | List pending requests |
| GET /v1/device/inbox/:imei | EXISTS | InboxHandler.GetInboxEntry | Get single request |
| POST /v1/device/inbox/:imei/ack | EXISTS | InboxHandler.AckInbox | Approve/reject request |
| POST /v1/device/inbox (DEPRECATED: was /v1/device/register) | EXISTS | RegisterHandler.Handle | Register new device |

### 2.3 Existing Domain Entities

| Entity | Location | Status |
|--------|----------|--------|
| device.Device | domain/device/entity.go | EXISTS |
| device.Status | domain/device/status.go | EXISTS |
| command.Command | domain/command/entity.go | EXISTS |

### 2.4 Database Tables

| Table | Status |
|-------|--------|
| devices | EXISTS |
| inbox_requests | MISSING |

---

## 3. Registration Flow

### 3.1 Full Registration Flow

```

                        DEVICE REGISTRATION FLOW                        


  DEVICE                          SERVER                         OPERATOR
                                                                   
      1. POST /v1/device/inbox                                   
      {imei, model, manufacturer, fcmToken, firebaseInstallId}    
                                    
                                                                   
                                     Store in INBOX               
                                     Status: pending              
                                                                   
                                                                     2. GET /v1/device/inbox
                                   
                                     [pending requests]            
                                                                   
                                                                     3. POST /v1/device/inbox/:imei/ack
                                   
                                     {action: "approve"}          
                                                                   
                                     Generate commandSecret       
                                     Send FCM push with secret    
                                     Update status: approved      
                                                                   
      4. FCM push received                                        
      {commandSecret, approved}                                   
                                    
                                                                   
      Store commandSecret locally                                  
                                                                   
      5. POST /v1/device/confirm                                  
      {imei, commandSecret}                                        
                                    
                                                                   
                                     Validate commandSecret        
                                     Move to DEVICES table        
                                     Status: registered           
                                                                   
      6. 200 OK                                                   
                                    
                                                                   
```

### 3.2 Deregistration Flow

```
  DEVICE                          SERVER                         OPERATOR
                                                                   
                                                                     1. DELETE /v1/device/:imei
                                   
                                                                   
                                     Mark as deregistered          
                                     (soft delete - 30 day retention)
                                                                   
                                                                   
      2. Next command fails                                       
      (FCM returns error)                                        
                                    
                                                                   
      3. Device re-registers                                       
      (if needed)                                                 
                                    
                                                                   
```

---

## 4. Required API Endpoints

### 4.1 Inbox Endpoints

#### GET /v1/device/inbox

Get all pending registration requests.

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| status | string | pending | Filter: all, pending, approved, rejected |
| page | int | 1 | Page number |
| limit | int | 20 | Items per page (max 100) |

**Response (200 OK):**
```json
{
  "requests": [
    {
      "id": "inb_abc123",
      "imei": "861234567890123",
      "model": "Pixel 8",
      "manufacturer": "Google",
      "fcmToken": "firebase_token_here",
      "firebaseInstallId": "firebase_install_id",
      "status": "pending",
      "createdAt": 1718900000000,
      "approvedAt": null,
      "rejectedAt": null
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 5,
    "totalPages": 1
  }
}
```

---

#### GET /v1/device/inbox/:imei

Get a single pending request by IMEI.

**Response (200 OK):**
```json
{
  "id": "inb_abc123",
  "imei": "861234567890123",
  "model": "Pixel 8",
  "manufacturer": "Google",
  "fcmToken": "firebase_token_here",
  "firebaseInstallId": "firebase_install_id",
  "osVersion": "Android 14",
  "appVersion": "2.1.0",
  "status": "pending",
  "createdAt": 1718900000000,
  "notes": null
}
```

**Errors:**
- 404 - Request not found

---

#### POST /v1/device/inbox/:imei/ack

Acknowledge (approve or reject) a pending request.

**Request:**
```json
{
  "action": "approve",
  "notes": "Optional operator notes"
}
```

**Response (200 OK) - Approved:**
```json
{
  "id": "inb_abc123",
  "imei": "861234567890123",
  "status": "approved",
  "approvedAt": 1718900500000,
  "commandSecret": "generated_secret_here",
  "fcmPushSent": true
}
```

**Response (200 OK) - Rejected:**
```json
{
  "id": "inb_abc123",
  "imei": "861234567890123",
  "status": "rejected",
  "rejectedAt": 1718900500000,
  "notes": "Device not authorized"
}
```

**Errors:**
- 400 - Invalid action (must be approve or reject)
- 404 - Request not found
- 409 - Request already acknowledged

---

### 4.2 Device Endpoints

#### GET /v1/devices

Get all registered devices.

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| status | string | all | Filter: all, online, offline |
| search | string | null | Search by IMEI or device name |
| page | int | 1 | Page number |
| limit | int | 20 | Items per page (max 100) |

**Response (200 OK):**
```json
{
  "devices": [
    {
      "id": "dev_abc123",
      "imei": "861234567890123",
      "deviceName": "Pixel 8 Pro",
      "model": "Pixel 8",
      "manufacturer": "Google",
      "status": "online",
      "registeredAt": 1718900500000,
      "lastSeen": 1718901000000,
      "appVersion": "2.1.0"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "totalPages": 8
  }
}
```

---

#### GET /v1/devices/:imei

Get single device details.

**Response (200 OK):**
```json
{
  "id": "dev_abc123",
  "imei": "861234567890123",
  "deviceName": "Pixel 8 Pro",
  "model": "Pixel 8",
  "manufacturer": "Google",
  "osVersion": "Android 14",
  "appVersion": "2.1.0",
  "securityPatch": "2024-03-01",
  "status": "online",
  "registeredAt": 1718900500000,
  "lastSeen": 1718901000000,
  "fcmTokenValid": true,
  "commandSecretSet": true,
  "connection": {
    "webSocketStatus": "connected",
    "connectedAt": 1718900900000,
    "protocol": "WSS",
    "clientIp": "192.168.1.x"
  }
}
```

---

#### DELETE /v1/devices/:imei

Deregister a device (soft delete).

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| hard | bool | false | If true, permanently delete |

**Response (200 OK):**
```json
{
  "imei": "861234567890123",
  "status": "deregistered",
  "deregisteredAt": 1718901500000,
  "retentionUntil": 1721493500000
}
```

**Notes:**
- Soft delete by default (30 day retention)
- Hard delete requires confirmation

---

### 4.3 Registration Endpoints (Update)

#### POST /v1/device/inbox (DEPRECATED: was /v1/device/register)

Operator-initiated registration (alternative to inbox flow).

**Request:**
```json
{
  "imei": "861234567890123",
  "deviceName": "Pixel 8 Pro",
  "fcmToken": "firebase_token_here",
  "firebaseInstallId": "firebase_install_id"
}
```

**Response (201 Created):**
```json
{
  "deviceId": "dev_abc123",
  "imei": "861234567890123",
  "status": "pending_command_secret",
  "commandSecret": "generated_secret_here"
}
```

---

#### POST /v1/device/confirm

Device confirms registration (already exists, may need updates).

**Request:**
```json
{
  "imei": "861234567890123",
  "commandSecret": "device_provided_secret"
}
```

**Response (200 OK):**
```json
{
  "imei": "861234567890123",
  "status": "registered",
  "registeredAt": 1718900500000
}
```

---

## 5. Database Schema

### 5.1 Inbox Table (NEW)

```sql
CREATE TABLE inbox_requests (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    imei TEXT NOT NULL UNIQUE,
    model TEXT,
    manufacturer TEXT,
    os_version TEXT,
    app_version TEXT,
    fcm_token TEXT NOT NULL,
    firebase_install_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    rejected_at TIMESTAMPTZ,
    command_secret TEXT,
    notes TEXT,
    operator_id TEXT,
    
    CONSTRAINT idx_imei UNIQUE (imei),
    CONSTRAINT idx_status (status),
    CONSTRAINT idx_created_at (created_at DESC)
);

CREATE INDEX idx_inbox_pending ON inbox_requests(status, created_at DESC);
```

### 5.2 Devices Table Updates

```sql
-- Add new columns to existing devices table
ALTER TABLE devices ADD COLUMN IF NOT EXISTS device_name TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS os_version TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS security_patch TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS command_secret_hash TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS deregistered_at TIMESTAMPTZ;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS deletion_scheduled_at TIMESTAMPTZ;

-- Create index for device name search
CREATE INDEX IF NOT EXISTS idx_devices_name ON devices(device_name);
```

### 5.3 Registration Logs Table (NEW)

```sql
CREATE TABLE registration_logs (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    device_id TEXT NOT NULL,
    action TEXT NOT NULL,
    operator_id TEXT,
    details JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_device FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE SET NULL
);

CREATE INDEX idx_registration_logs_device ON registration_logs(device_id, created_at DESC);
```

---

## 6. Backend File Structure

```
apps/api/internal/
 api/
    server_routes.go                # Route registration
    api_server.go                   # Server setup
    wire/
       providers.go               # Dependency providers
       wire_handlers.go           # Handler wiring
    handlers/
        inbox/
           inbox_handler.go        # InboxHandler
           inbox_routes.go        # (included in handler)
        device/
           devices_handler.go      # DevicesHandler (GetDevices, GetDeviceDetail)
           device_list.go          # ListHandler (Count)
           device_updater.go       # UpdaterHandler (Delete, Update)
           device_confirm.go       # ConfirmHandler
           device_register.go      # RegisterHandler
           dev_status.go          # StatusHandler
           device_logs_handler.go  # LogsHandler
           device_metrics_handler.go # MetricsHandler
           device_telemetry_handler.go # TelemetryHandler
        websocket/
           websocket_handler.go    # StreamHandler
           stream_message.go       # MessageRouter
           websocket_presenter.go  # WebSocket presenter
           websocket_stream.go    # HTTP→WS upgrade
        ...
 application/
    inbox/
       inbox_service.go           # InboxService
    device/
        device_service.go           # DeviceService
 domain/
    inbox/
       inbox_entity.go             # InboxEntry, InboxListResponse
       inbox_repository.go        # Repository interface
    device/
        device_entity.go           # Device entity
        device_repository.go        # Repository interface
 infrastructure/
     storage/
         device_storage.go           # Storage implementation
```


---

## 7. Handler Specifications

### 7.1 Inbox Handler

**File:** `internal/api/handlers/inbox/inbox_handler.go`

```go
package inbox

import (
    "net/http"
    "strconv"

    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/inbox"
    "github.com/gin-gonic/gin"
)

type Handler struct {
    service *inbox.Service
}

func NewHandler(service *inbox.Service) *Handler {
    return &Handler{service: service}
}

// GetInbox handles GET /v1/device/inbox
func (h *Handler) GetInbox(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    status := c.Query("status")

    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 20
    }

    result, err := h.service.GetInbox(c.Request.Context(), &inbox.Query{
        Status: status,
        Page:   page,
        Limit:  limit,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get inbox"})
        return
    }

    c.JSON(http.StatusOK, result)
}

// GetInboxEntry handles GET /v1/device/inbox/:imei
func (h *Handler) GetInboxEntry(c *gin.Context) {
    imei := c.Param("imei")
    
    entry, err := h.service.GetInboxEntry(c.Request.Context(), imei)
    if err != nil {
        if err == inbox.ErrNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "inbox entry not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get inbox entry"})
        return
    }

    c.JSON(http.StatusOK, entry)
}

// AckInbox handles POST /v1/device/inbox/:imei/ack
func (h *Handler) AckInbox(c *gin.Context) {
    imei := c.Param("imei")
    
    var req struct {
        Action string `json:"action" binding:"required"`
        Notes  string `json:"notes"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "invalid request"})
        return
    }

    if req.Action != "approve" && req.Action != "reject" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "action must be approve or reject"})
        return
    }

    result, err := h.service.AckInbox(c.Request.Context(), imei, req.Action, req.Notes)
    if err != nil {
        if err == inbox.ErrNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "inbox entry not found"})
            return
        }
        if err == inbox.ErrAlreadyAcknowledged {
            c.JSON(http.StatusConflict, gin.H{"error": "conflict", "message": "entry already acknowledged"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to acknowledge"})
        return
    }

    c.JSON(http.StatusOK, result)
}
```

---

### 7.2 Devices Handler

**File:** `internal/api/handlers/device/devices_handler.go`

```go
package device

import (
    "net/http"
    "strconv"

    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
    "github.com/gin-gonic/gin"
)

type DevicesHandler struct {
    service *device.Service
}

func NewDevicesHandler(service *device.Service) *DevicesHandler {
    return &DevicesHandler{service: service}
}

// GetDevices handles GET /v1/devices
func (h *DevicesHandler) GetDevices(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    status := c.Query("status")
    search := c.Query("search")

    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 20
    }

    result, err := h.service.GetDevices(c.Request.Context(), &device.ListQuery{
        Status: status,
        Search: search,
        Page:   page,
        Limit:  limit,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get devices"})
        return
    }

    c.JSON(http.StatusOK, result)
}

// GetDevice handles GET /v1/devices/:imei
func (h *DevicesHandler) GetDevice(c *gin.Context) {
    imei := c.Param("imei")
    
    device, err := h.service.GetDeviceByIMEI(c.Request.Context(), imei)
    if err != nil {
        if err == device.ErrNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "device not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get device"})
        return
    }

    c.JSON(http.StatusOK, device)
}

// DeregisterDevice handles DELETE /v1/devices/:imei
func (h *DevicesHandler) DeregisterDevice(c *gin.Context) {
    imei := c.Param("imei")
    hard := c.Query("hard") == "true"

    result, err := h.service.DeregisterDevice(c.Request.Context(), imei, hard)
    if err != nil {
        if err == device.ErrNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "device not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deregister device"})
        return
    }

    c.JSON(http.StatusOK, result)
}
```

---

## 8. Service Layer

### 8.1 Inbox Service

**File:** `application/inbox/service.go`

```go
package inbox

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "errors"
    "time"

    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/inbox"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/fcm"
)

var (
    ErrNotFound             = errors.New("inbox entry not found")
    ErrAlreadyAcknowledged  = errors.New("entry already acknowledged")
)

type Service struct {
    repo         *inbox.Repository
    fcmNotifier  fcm.Notifier
}

type Query struct {
    Status string
    Page   int
    Limit  int
}

func NewService(repo *inbox.Repository, fcmNotifier fcm.Notifier) *Service {
    return &Service{repo: repo, fcmNotifier: fcmNotifier}
}

func (s *Service) GetInbox(ctx context.Context, q *Query) (*InboxResult, error) {
    offset := (q.Page - 1) * q.Limit
    
    entries, total, err := s.repo.List(ctx, q.Status, q.Limit, offset)
    if err != nil {
        return nil, err
    }

    totalPages := (total + q.Limit - 1) / q.Limit

    return &InboxResult{
        Requests:  entries,
        Pagination: Pagination{
            Page:       q.Page,
            Limit:      q.Limit,
            Total:      total,
            TotalPages: totalPages,
        },
    }, nil
}

func (s *Service) GetInboxEntry(ctx context.Context, imei string) (*inbox.Entry, error) {
    entry, err := s.repo.GetByIMEI(ctx, imei)
    if err != nil {
        if err == inbox.ErrNotFound {
            return nil, ErrNotFound
        }
        return nil, err
    }
    return entry, nil
}

func (s *Service) AckInbox(ctx context.Context, imei, action, notes string) (*AckResult, error) {
    entry, err := s.repo.GetByIMEI(ctx, imei)
    if err != nil {
        if err == inbox.ErrNotFound {
            return nil, ErrNotFound
        }
        return nil, err
    }

    if entry.Status != inbox.StatusPending {
        return nil, ErrAlreadyAcknowledged
    }

    if action == "approve" {
        // Generate command secret
        secret := generateSecret(32)
        
        // Update entry
        entry.Status = inbox.StatusApproved
        entry.ApprovedAt = time.Now()
        entry.CommandSecret = secret
        entry.OperatorNotes = notes

        // Send FCM push
        if s.fcmNotifier != nil {
            err := s.fcmNotifier.SendRegistrationApproved(ctx, fcm.RegistrationApproved{
                Token:        entry.FCMToken,
                CommandSecret: secret,
            })
            if err != nil {
                // Log error but don't fail
                // Could implement retry logic here
            }
        }

        if err := s.repo.Update(ctx, entry); err != nil {
            return nil, err
        }

        return &AckResult{
            ID:            entry.ID,
            IMEI:          entry.IMEI,
            Status:        entry.Status,
            ApprovedAt:    entry.ApprovedAt,
            CommandSecret: secret,
            FCMPushSent:   true,
        }, nil

    } else {
        // Reject
        entry.Status = inbox.StatusRejected
        entry.RejectedAt = time.Now()
        entry.OperatorNotes = notes

        if err := s.repo.Update(ctx, entry); err != nil {
            return nil, err
        }

        return &AckResult{
            ID:         entry.ID,
            IMEI:       entry.IMEI,
            Status:     entry.Status,
            RejectedAt: entry.RejectedAt,
            Notes:      notes,
        }, nil
    }
}

func generateSecret(length int) string {
    bytes := make([]byte, length)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}
```

---

## 9. GraphQL Schema

### 9.1 Types

```graphql
enum InboxStatus {
  PENDING
  APPROVED
  REJECTED
}

type InboxEntry {
  id: ID!
  imei: String!
  model: String
  manufacturer: String
  osVersion: String
  appVersion: String
  fcmToken: String!
  firebaseInstallId: String!
  status: InboxStatus!
  createdAt: DateTime!
  approvedAt: DateTime
  rejectedAt: DateTime
  commandSecret: String
  notes: String
  operatorId: String
}

type InboxConnection {
  requests: [InboxEntry!]!
  pagination: Pagination!
}

type Device {
  id: ID!
  imei: String!
  deviceName: String
  model: String
  manufacturer: String
  osVersion: String
  appVersion: String
  securityPatch: String
  status: DeviceStatus!
  registeredAt: DateTime
  lastSeen: DateTime
  fcmTokenValid: Boolean!
  commandSecretSet: Boolean!
  connection: DeviceConnection
}

type DeviceConnection {
  webSocketStatus: String!
  connectedAt: DateTime
  protocol: String
  clientIp: String
}

type DeviceListConnection {
  devices: [Device!]!
  pagination: Pagination!
}

type AckResult {
  id: ID!
  imei: String!
  status: InboxStatus!
  approvedAt: DateTime
  rejectedAt: DateTime
  commandSecret: String
  fcmPushSent: Boolean
  notes: String
}

type DeregisterResult {
  imei: String!
  status: String!
  deregisteredAt: DateTime!
  retentionUntil: DateTime!
}
```

### 9.2 Queries

```graphql
type Query {
  inbox(status: InboxStatus, page: Int, limit: Int): InboxConnection!
  inboxEntry(imei: String!): InboxEntry
  devices(status: DeviceStatus, search: String, page: Int, limit: Int): DeviceListConnection!
  device(imei: String!): Device
}
```

### 9.3 Mutations

```graphql
type Mutation {
  ackInbox(imei: String!, action: AckAction!, notes: String): AckResult!
  deregisterDevice(imei: String!, hard: Boolean): DeregisterResult!
}
```

---

## 10. Error Handling

### 10.1 Error Response Format

```json
{
  "error": "error_code",
  "message": "Human readable message",
  "details": {}
}
```

### 10.2 Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| bad_request | 400 | Invalid request parameters |
| unauthorized | 401 | Authentication required |
| forbidden | 403 | Access denied |
| not_found | 404 | Resource not found |
| conflict | 409 | Resource state conflict |
| rate_limited | 429 | Too many requests |
| internal_error | 500 | Server error |

---

## 11. Rate Limiting & Security

### 11.1 Rate Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| GET /v1/device/inbox | 60 | 1 minute |
| POST /v1/device/inbox/:imei/ack | 10 | 1 minute |
| GET /v1/devices | 60 | 1 minute |
| DELETE /v1/devices/:imei | 10 | 1 minute |

### 11.2 Security Requirements

1. **Authentication** - All endpoints require authenticated operator
2. **Device Ownership** - DOA check on all device-specific operations
3. **Audit Logging** - Log all registration/deregistration actions
4. **Secret Generation** - Use crypto/rand for commandSecret
5. **FCM Validation** - Verify FCM token format before storing

---

## 12. File Changes Summary

### 12.1 Total File Count

| Category | New | Modified | Total |
|----------|-----|---------|-------|
| Domain Layer | 4 | 1 | 5 |
| Application Layer | 3 | 1 | 4 |
| Handler Layer | 3 | 2 | 5 |
| Infrastructure | 3 | 1 | 4 |
| GraphQL | 2 | 2 | 4 |
| Router | 0 | 1 | 1 |
| **TOTAL** | **15** | **8** | **23** |

### 12.2 All Files Listed

#### Domain Layer (4 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| domain/inbox/inbox_entity.go | NEW | InboxEntry types |
| domain/inbox/inbox_repository.go | NEW | Repository interface |
| domain/inbox/inbox_errors.go | NEW | Domain errors |
| domain/inbox/inbox_status.go | NEW | Status constants |
| domain/device/device_entity.go | MODIFIED | Add new fields |

#### Application Layer (3 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| application/inbox/inbox_service.go | NEW | Inbox business logic |
| application/inbox/inbox_dto.go | NEW | Inbox DTOs |
| application/inbox/inbox_errors.go | NEW | Service errors |
| application/device/device_service.go | MODIFIED | Add list/filter methods |

#### Handler Layer (EXISTING - Updated)

| File | Handler | Purpose |
|------|---------|---------|
| internal/api/handlers/inbox/inbox_handler.go | InboxHandler | GetInboxRequest, GetInboxEntry, UpdateInboxEntry |
| internal/api/handlers/device/devices_handler.go | DevicesHandler | GetDevices, GetDeviceDetail, DeregisterDevice |
| internal/api/handlers/device/device_list.go | ListHandler | Count |
| internal/api/handlers/device/device_updater.go | UpdaterHandler | Delete, Update |
| internal/api/server_routes.go | (routes) | Route registration |

#### Infrastructure (EXISTING)

| File | Purpose |
|------|---------|
| internal/infrastructure/storage/device_storage.go | Device & inbox storage |
| internal/infrastructure/storage/migrations/ | SQL migrations |

#### GraphQL (EXISTING - Updated)

| File | Purpose |
|------|---------|
| internal/api/graphql/schema/objects.go | Add inbox/device types |
| internal/api/graphql/resolver/ | Add resolvers |
| internal/api/graphql/schema/schema.go | Add root types |

---

## 13. Implementation Order

### Phase 1: Database (Day 1)
1. Create migration `001_create_inbox_requests.sql`
2. Add columns to `devices` table
3. Create `registration_logs` table
4. Test migrations

### Phase 2: Domain Layer (Day 1)
1. Create `domain/inbox/inbox_entity.go`
2. Create `domain/inbox/inbox_status.go`
3. Create `domain/inbox/inbox_repository.go`
4. Create `domain/inbox/inbox_errors.go`
5. Update `domain/device/device_entity.go`

### Phase 3: Infrastructure (Day 1-2)
1. Create `infrastructure/storage/inbox_storage.go`
2. Implement inbox repository
3. Create `infrastructure/storage/registration_log_storage.go`
4. Update `infrastructure/storage/device_storage.go`

### Phase 4: Application Layer (Day 2)
1. Create `application/inbox/inbox_service.go`
2. Implement GetInbox, GetEntry, Ack methods
3. Implement FCM notification
4. Update `application/device/device_service.go`

### Phase 5: Handlers (Day 2-3)
1. Create `handlers/inbox/inbox_handler.go`
2. Create `handlers/device/device_list_handler.go`
3. Update router
4. Test endpoints

### Phase 6: GraphQL (Day 3)
1. Create GraphQL schema files
2. Add resolvers
3. Test queries

### Phase 7: Testing (Day 3-4)
1. Unit tests for services
2. Integration tests for handlers
3. E2E tests for full flow
4. FCM notification tests

---

*Document Version: 1.0*
*Status: Ready for Implementation*
