# Server Backend - Diagnostics API Specification

> **Version:** 1.0
> **Status:** Draft
> **Created:** 2026-06-24
> **Target:** Production MVP
> **Go Package:** `github.com/VinnsEdesigner/vyzorix/apps/api`

---

## Table of Contents

1. Overview
2. Current State Analysis
3. Required API Endpoints
4. Database Schema
5. Backend File Structure
6. Handler Specifications
7. Service Layer
8. GraphQL Schema
9. Error Handling
10. Rate Limiting & Security
11. File Changes Summary
12. Implementation Order

---

## 1. Overview

### 1.1 Purpose

This document maps out the server-side requirements to support the Diagnostics page as specified in DIAGNOSTICS_PAGE.md.

### 1.2 Frontend Requirements Summary

| Feature | Description | Required Endpoints |
|---------|-------------|-------------------|
| Device Inspector | Full device state snapshot | GET /v1/device/:imei/inspect |
| Timeline | Chronological event audit trail | GET /v1/device/:imei/timeline |

### 1.3 Inspector Data Sections

The Diagnostics Inspector requires the following device data:

| Section | Fields |
|---------|--------|
| **Identity** | imei, deviceName, model, manufacturer |
| **Software** | osVersion, appVersion, securityPatch, buildId |
| **Registration** | status, registeredAt, fcmTokenValid, fcmTokenRefreshedAt, commandSecretSet |
| **Connection** | webSocketStatus, connectedAt, fcmStatus, lastSeen, clientIp, protocol |
| **Telemetry** | lastTimestamp, framesToday, avgLatencyMs, totalBytesToday, sessionsToday |

### 1.4 Timeline Event Types

| Event | Description |
|-------|-------------|
| TELEMETRY | Telemetry frame received |
| COMMAND_SENT | Command dispatched |
| COMMAND_ACK | Command acknowledged |
| COMMAND_FAILED | Command failed |
| CONNECTION_OPEN | WebSocket connected |
| CONNECTION_LOST | WebSocket disconnected |
| FCM_FALLBACK | FCM fallback activated |
| RECONNECTED | WebSocket reconnected |
| THRESHOLD_BREACH | Risk/thermal threshold exceeded |
| REGISTERED | Device registered |
| DEREGISTERED | Device deregistered |
| ERROR | Error occurred |

---

## 2. Current State Analysis

### 2.1 Existing Related Endpoints

| Endpoint | Status | Handler | Notes |
|----------|--------|---------|-------|
| GET /v1/device/:imei | EXISTS | DeviceHandler.Get | Basic device info |
| GET /v1/device/:imei/commands/pending | EXISTS | GetPending | Pending commands |
| WebSocket /ws | EXISTS | WebSocketHandler | Real-time telemetry |

### 2.2 Missing Endpoints

| Endpoint | Status |
|----------|--------|
| GET /v1/device/:imei/inspect | MISSING |
| GET /v1/device/:imei/timeline | MISSING |

### 2.3 Data Availability

| Data | Source | Status |
|------|--------|--------|
| Device Identity | devices table | EXISTS |
| Software Info | devices table + telemetry | PARTIAL |
| Registration Status | devices table | EXISTS |
| Connection Status | ws hub + devices table | PARTIAL |
| Telemetry Stats | telemetry aggregation | MISSING |
| Timeline Events | logs/events table | MISSING |

---

## 3. Required API Endpoints

### 3.1 GET /v1/device/:imei/inspect

Get full device inspection data for the Diagnostics Inspector.

**Response (200 OK):**
```json
{
  "identity": {
    "imei": "861234567890123",
    "deviceName": "Pixel 8 Pro",
    "model": "Pixel 8",
    "manufacturer": "Google"
  },
  "software": {
    "osVersion": "Android 14",
    "appVersion": "2.1.0",
    "securityPatch": "2024-03-01",
    "buildId": "UP1A.231005.007"
  },
  "registration": {
    "status": "registered",
    "registeredAt": 1718900300000,
    "fcmTokenValid": true,
    "fcmTokenRefreshedAt": 1718890000000,
    "commandSecretSet": true
  },
  "connection": {
    "webSocketStatus": "connected",
    "connectedAt": 1718900000000,
    "fcmStatus": "valid",
    "lastSeen": 1718900500000,
    "clientIp": "192.168.1.xxx",
    "protocol": "WSS"
  },
  "telemetry": {
    "lastTimestamp": 1718900567000,
    "framesToday": 4521,
    "avgLatencyMs": 45,
    "totalBytesToday": 15728640,
    "sessionsToday": 3
  }
}
```

**Response Fields:**

| Section | Field | Type | Source |
|---------|-------|------|--------|
| identity | imei | string | devices.id |
| identity | deviceName | string | devices.device_name |
| identity | model | string | devices.model |
| identity | manufacturer | string | devices.manufacturer |
| software | osVersion | string | devices.os_version |
| software | appVersion | string | devices.app_version |
| software | securityPatch | string | devices.security_patch |
| software | buildId | string | devices.build_id |
| registration | status | string | derived: online/offline/deregistered |
| registration | registeredAt | int64 | devices.registered_at |
| registration | fcmTokenValid | bool | derived: fcm token check |
| registration | fcmTokenRefreshedAt | int64 | devices.fcm_token_refreshed_at |
| registration | commandSecretSet | bool | devices.command_secret_hash IS NOT NULL |
| connection | webSocketStatus | string | ws hub lookup |
| connection | connectedAt | int64 | ws hub connection time |
| connection | fcmStatus | string | derived: fcm validation |
| connection | lastSeen | int64 | devices.last_seen |
| connection | clientIp | string | ws hub / request |
| connection | protocol | string | "WSS" |
| telemetry | lastTimestamp | int64 | latest telemetry frame |
| telemetry | framesToday | int | count of frames today |
| telemetry | avgLatencyMs | int | average WS latency |
| telemetry | totalBytesToday | int64 | bytes transferred today |
| telemetry | sessionsToday | int | session count today |

---

### 3.2 GET /v1/device/:imei/timeline

Get chronological event timeline for the Diagnostics Timeline.

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| eventType | string | all | Filter: telemetry, command, connection, error |
| startTime | int64 | -24h | Start timestamp (ms) |
| endTime | int64 | now | End timestamp (ms) |
| limit | int | 50 | Max results (max 200) |
| cursor | string | null | Pagination cursor |

**Response (200 OK):**
```json
{
  "events": [
    {
      "id": "evt_uuid_001",
      "type": "TELEMETRY",
      "timestamp": 1718900567000,
      "data": {
        "riskScore": 45,
        "thermalTemp": 38.5,
        "bufferLevel": 67,
        "latencyMs": 23
      }
    },
    {
      "id": "evt_uuid_002",
      "type": "COMMAND_SENT",
      "timestamp": 1718900500000,
      "data": {
        "command": "WAKE_UP_UPDATER",
        "dispatchId": "disp_xyz789",
        "status": "delivered"
      }
    },
    {
      "id": "evt_uuid_003",
      "type": "CONNECTION_OPEN",
      "timestamp": 1718900450000,
      "data": {
        "protocol": "websocket",
        "ip": "192.168.1.xxx"
      }
    }
  ],
  "pagination": {
    "limit": 50,
    "hasMore": true,
    "nextCursor": "eyJ0IjoxNzE4OTAwNDYwMDAwMCwidCI6ImV2dF91dWlkXzAwMiJ9"
  }
}
```

**Event Type Mapping:**

| Frontend Type | Backend Event Types |
|---------------|------------------|
| telemetry | TELEMETRY |
| command | COMMAND_SENT, COMMAND_ACK, COMMAND_FAILED |
| connection | CONNECTION_OPEN, CONNECTION_LOST, FCM_FALLBACK, RECONNECTED |
| error | ERROR, THRESHOLD_BREACH |

---

## 4. Database Schema

### 4.1 Timeline Events Table (NEW)

```sql
CREATE TABLE device_events (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    device_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    data JSONB,
    
    CONSTRAINT fk_device FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE
);

CREATE INDEX idx_device_events_device_timestamp ON device_events(device_id, timestamp DESC);
CREATE INDEX idx_device_events_cursor ON device_events(device_id, timestamp DESC, id);
CREATE INDEX idx_device_events_type ON device_events(device_id, event_type, timestamp DESC);
```

### 4.2 Telemetry Stats View (NEW)

```sql
-- Create a view for aggregated telemetry stats
CREATE VIEW device_telemetry_stats AS
SELECT 
    d.id as device_id,
    MAX(t.timestamp) as last_timestamp,
    COUNT(t.id) as frames_today,
    AVG(EXTRACT(EPOCH FROM (t.timestamp - LAG(t.timestamp) OVER w)) * 1000) as avg_latency_ms,
    SUM(t.data->>'bytes') as total_bytes_today,
    COUNT(DISTINCT DATE(t.timestamp)) as sessions_today
FROM devices d
LEFT JOIN telemetry_frames t ON t.device_id = d.id
    AND t.timestamp >= NOW() - INTERVAL '24 hours'
WINDOW w AS (PARTITION BY t.device_id ORDER BY t.timestamp)
GROUP BY d.id;
```

### 4.3 Device Table Updates

```sql
-- Add columns for inspection data
ALTER TABLE devices ADD COLUMN IF NOT EXISTS device_name TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS os_version TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS security_patch TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS build_id TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS fcm_token_refreshed_at TIMESTAMPTZ;
```

---

## 5. Backend File Structure

```
apps/api/internal/
├── api/
│   ├── handlers/
│   │   ├── diagnostics/
│   │   │   ├── diagnostics_inspect_handler.go  # NEW - inspection handler
│   │   │   ├── diagnostics_timeline_handler.go  # NEW - timeline handler
│   │   │   └── diagnostics_routes.go           # NEW - diagnostics routes
│   │   └── router.go                          # MODIFIED - add diagnostics
│   └── middleware/
│       └── ...
├── application/
│   └── diagnostics/
│       ├── diagnostics_service.go              # NEW - inspection + timeline logic
│       └── diagnostics_dto.go                  # NEW - request/response DTOs
├── domain/
│   ├── diagnostics/
│   │   ├── diagnostics_types.go              # NEW - diagnostic types
│   │   ├── diagnostics_repository.go         # NEW - repository interface
│   │   └── diagnostics_errors.go            # NEW - domain errors
│   └── device/
│       └── device_repository.go              # EXISTS - may need new methods
├── infrastructure/
│   └── storage/
│       ├── diagnostics_storage.go            # NEW - diagnostics queries
│       └── device_storage.go                 # EXISTS - add stats query
└── ws/
    └── hub.go                                # EXISTS - may expose connection info
```

---

## 6. Handler Specifications

### 6.1 Inspect Handler

**File:** `api/handlers/diagnostics/diagnostics_inspect_handler.go`

```go
package diagnostics

import (
    "net/http"

    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/diagnostics"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
    "github.com/gin-gonic/gin"
)

type InspectHandler struct {
    service        *diagnostics.Service
    deviceService  *device.Service
}

func NewInspectHandler(service *diagnostics.Service, deviceService *device.Service) *InspectHandler {
    return &InspectHandler{
        service:       service,
        deviceService: deviceService,
    }
}

// GetInspection handles GET /v1/device/:imei/inspect
func (h *InspectHandler) GetInspection(c *gin.Context) {
    imei := c.Param("imei")
    if imei == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "imei required"})
        return
    }

    // Get device first to verify it exists and get basic info
    dev, err := h.deviceService.GetDeviceByIMEI(c.Request.Context(), imei)
    if err != nil {
        if err == device.ErrNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get device"})
        return
    }

    // Get full inspection data
    inspection, err := h.service.GetDeviceInspection(c.Request.Context(), imei)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get inspection"})
        return
    }

    c.JSON(http.StatusOK, inspection)
}
```

---

### 6.2 Timeline Handler

**File:** `api/handlers/diagnostics/timeline.go`

```go
package diagnostics

import (
    "net/http"
    "strconv"
    "time"

    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/diagnostics"
    "github.com/gin-gonic/gin"
)

type TimelineHandler struct {
    service *diagnostics.Service
}

func NewTimelineHandler(service *diagnostics.Service) *TimelineHandler {
    return &TimelineHandler{service: service}
}

// GetTimeline handles GET /v1/device/:imei/timeline
func (h *TimelineHandler) GetTimeline(c *gin.Context) {
    imei := c.Param("imei")
    if imei == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "imei required"})
        return
    }

    // Parse query params
    eventType := c.Query("eventType")
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
    cursor := c.Query("cursor")
    
    startTime, _ := strconv.ParseInt(c.DefaultQuery("startTime", "-1"), 10, 64)
    endTime, _ := strconv.ParseInt(c.DefaultQuery("endTime", "0"), 10, 64)

    // Validate
    if limit < 1 || limit > 200 {
        limit = 50
    }

    // Default to 24h if not set
    if startTime < 0 {
        startTime = time.Now().Add(-24 * time.Hour).UnixMilli()
    }
    if endTime == 0 {
        endTime = time.Now().UnixMilli()
    }

    // Get timeline
    result, err := h.service.GetDeviceTimeline(c.Request.Context(), &diagnostics.TimelineQuery{
        DeviceIMEI: imei,
        EventType:  eventType,
        Limit:      limit,
        Cursor:     cursor,
        StartTime:  startTime,
        EndTime:    endTime,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get timeline"})
        return
    }

    c.JSON(http.StatusOK, result)
}
```

---

## 7. Service Layer

### 7.1 Diagnostics Service

**File:** `application/diagnostics/service.go`

```go
package diagnostics

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "time"

    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/diagnostics"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"
)

type Service struct {
    diagnosticsRepo *diagnostics.Repository
    deviceService  *device.Service
    wsHub         *ws.Hub
}

func NewService(repo *diagnostics.Repository, deviceService *device.Service, wsHub *ws.Hub) *Service {
    return &Service{
        diagnosticsRepo: repo,
        deviceService:  deviceService,
        wsHub:          wsHub,
    }
}

// GetDeviceInspection returns full inspection data for a device
func (s *Service) GetDeviceInspection(ctx context.Context, imei string) (*diagnostics.Inspection, error) {
    // Get base device info
    dev, err := s.deviceService.GetDeviceByIMEI(ctx, imei)
    if err != nil {
        return nil, err
    }

    // Get WebSocket connection status
    wsStatus := "disconnected"
    connectedAt := int64(0)
    clientIp := ""
    
    if s.wsHub != nil && s.wsHub.Online(imei) {
        wsStatus = "connected"
        connectedAt = s.wsHub.ConnectionTime(imei) // Would need to expose this from hub
        clientIp = s.wsHub.ClientIP(imei)
    }

    // Get telemetry stats
    stats, err := s.diagnosticsRepo.GetTelemetryStats(ctx, imei)
    if err != nil {
        // Log but don't fail
        stats = &diagnostics.TelemetryStats{}
    }

    // Build inspection response
    return &diagnostics.Inspection{
        Identity: diagnostics.IdentityInfo{
            IMEI:          dev.ID,
            DeviceName:    dev.DeviceName,
            Model:         dev.Model,
            Manufacturer:  dev.Manufacturer,
        },
        Software: diagnostics.SoftwareInfo{
            OSVersion:     dev.OSVersion,
            AppVersion:   dev.AppVersion,
            SecurityPatch: dev.SecurityPatch,
            BuildID:       dev.BuildID,
        },
        Registration: diagnostics.RegistrationInfo{
            Status:              s.determineDeviceStatus(dev),
            RegisteredAt:        dev.RegisteredAt,
            FCMTokenValid:        dev.FCMToken != "",
            FCMTokenRefreshedAt:  dev.FCMTokenRefreshedAt,
            CommandSecretSet:     dev.CommandSecretHash != "",
        },
        Connection: diagnostics.ConnectionInfo{
            WebSocketStatus: wsStatus,
            ConnectedAt:     connectedAt,
            FCMStatus:       s.determineFCMStatus(dev),
            LastSeen:         dev.LastSeen,
            ClientIP:         clientIp,
            Protocol:         "WSS",
        },
        Telemetry: diagnostics.TelemetryInfo{
            LastTimestamp:    stats.LastTimestamp,
            FramesToday:      stats.FramesToday,
            AvgLatencyMs:    stats.AvgLatencyMs,
            TotalBytesToday:  stats.TotalBytesToday,
            SessionsToday:    stats.SessionsToday,
        },
    }, nil
}

// GetDeviceTimeline returns paginated timeline events
func (s *Service) GetDeviceTimeline(ctx context.Context, query *TimelineQuery) (*TimelineResult, error) {
    // Decode cursor if provided
    var cursorTime time.Time
    var cursorID string
    if query.Cursor != "" {
        cursorTime, cursorID = s.decodeCursor(query.Cursor)
    }

    // Fetch events (fetch limit+1 to check for more)
    events, err := s.diagnosticsRepo.GetTimelineEvents(ctx, query.DeviceIMEI, query.EventType, query.Limit+1, query.StartTime, query.EndTime, cursorTime, cursorID)
    if err != nil {
        return nil, err
    }

    // Check if there are more results
    hasMore := len(events) > query.Limit
    if hasMore {
        events = events[:query.Limit]
    }

    // Generate next cursor
    var nextCursor string
    if hasMore && len(events) > 0 {
        last := events[len(events)-1]
        nextCursor = s.encodeCursor(last.Timestamp, last.ID)
    }

    return &TimelineResult{
        Events:     events,
        HasMore:    hasMore,
        NextCursor: nextCursor,
    }, nil
}

func (s *Service) decodeCursor(encoded string) (time.Time, string) {
    data, err := base64.StdEncoding.DecodeString(encoded)
    if err != nil {
        return time.Time{}, ""
    }
    var cursor struct {
        T string `json:"t"`
        I string `json:"i"`
    }
    if err := json.Unmarshal(data, &cursor); err != nil {
        return time.Time{}, ""
    }
    t, _ := time.Parse(time.RFC3339Nano, cursor.T)
    return t, cursor.I
}

func (s *Service) encodeCursor(t time.Time, id string) string {
    cursor := struct {
        T string `json:"t"`
        I string `json:"i"`
    }{t.Format(time.RFC3339Nano), id}
    data, _ := json.Marshal(cursor)
    return base64.StdEncoding.EncodeToString(data)
}

func (s *Service) determineDeviceStatus(dev *device.Device) string {
    if dev.DeregisteredAt != nil {
        return "deregistered"
    }
    if dev.LastSeen > 0 && time.Since(time.UnixMilli(dev.LastSeen)) > 5*time.Minute {
        return "offline"
    }
    return "online"
}

func (s *Service) determineFCMStatus(dev *device.Device) string {
    if dev.FCMToken == "" {
        return "not_set"
    }
    if dev.FCMTokenRefreshedAt > 0 && time.Since(time.UnixMilli(dev.FCMTokenRefreshedAt)) > 30*24*time.Hour {
        return "invalid"
    }
    return "valid"
}
```

---

## 8. GraphQL Schema

### 8.1 Types

```graphql
type IdentityInfo {
  imei: String!
  deviceName: String
  model: String
  manufacturer: String
}

type SoftwareInfo {
  osVersion: String!
  appVersion: String!
  securityPatch: String
  buildId: String
}

type RegistrationInfo {
  status: DeviceStatus!
  registeredAt: DateTime
  fcmTokenValid: Boolean!
  fcmTokenRefreshedAt: DateTime
  commandSecretSet: Boolean!
}

type ConnectionInfo {
  webSocketStatus: String!
  connectedAt: DateTime
  fcmStatus: String!
  lastSeen: DateTime
  clientIp: String
  protocol: String
}

type TelemetryInfo {
  lastTimestamp: DateTime!
  framesToday: Int!
  avgLatencyMs: Int
  totalBytesToday: BigInt
  sessionsToday: Int!
}

type DeviceInspection {
  identity: IdentityInfo!
  software: SoftwareInfo!
  registration: RegistrationInfo!
  connection: ConnectionInfo!
  telemetry: TelemetryInfo!
}

enum TimelineEventType {
  TELEMETRY
  COMMAND_SENT
  COMMAND_ACK
  COMMAND_FAILED
  CONNECTION_OPEN
  CONNECTION_LOST
  FCM_FALLBACK
  RECONNECTED
  THRESHOLD_BREACH
  REGISTERED
  DEREGISTERED
  ERROR
}

type TimelineEvent {
  id: ID!
  type: TimelineEventType!
  timestamp: DateTime!
  data: JSON
}

type TimelineConnection {
  events: [TimelineEvent!]!
  hasMore: Boolean!
  nextCursor: String
}
```

### 8.2 Queries

```graphql
type Query {
  deviceInspection(imei: String!): DeviceInspection!
  
  deviceTimeline(
    imei: String!
    eventType: TimelineEventType
    startTime: Int
    endTime: Int
    limit: Int = 50
    cursor: String
  ): TimelineConnection!
}
```

---

## 9. Error Handling

### 9.1 Error Response Format

```json
{
  "error": "error_code",
  "message": "Human readable message",
  "details": {}
}
```

### 9.2 Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| bad_request | 400 | Invalid request parameters |
| unauthorized | 401 | Authentication required |
| forbidden | 403 | Access denied |
| not_found | 404 | Device not found |
| rate_limited | 429 | Too many requests |
| internal_error | 500 | Server error |

---

## 10. Rate Limiting & Security

### 10.1 Rate Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| GET /v1/device/:imei/inspect | 30 | 1 minute |
| GET /v1/device/:imei/timeline | 30 | 1 minute |

### 10.2 Caching

| Endpoint | Cache TTL | Notes |
|----------|----------|-------|
| GET /v1/device/:imei/inspect | 10s | Short TTL due to real-time nature |
| GET /v1/device/:imei/timeline | No cache | Real-time events |

### 10.3 Security

1. **Authentication** - All endpoints require authenticated operator
2. **Device Authorization** - DOA check on all device operations
3. **Data Filtering** - Only return events user has access to

---

## 11. File Changes Summary

### 11.1 Total File Count

| Category | New | Modified | Total |
|----------|-----|---------|-------|
| Domain Layer | 3 | 1 | 4 |
| Application Layer | 2 | 0 | 2 |
| Handler Layer | 3 | 1 | 4 |
| Infrastructure | 2 | 1 | 3 |
| GraphQL | 2 | 2 | 4 |
| Router | 0 | 1 | 1 |
| **TOTAL** | **12** | **6** | **18** |

### 11.2 All Files Listed

#### Domain Layer (3 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| domain/diagnostics/diagnostics_types.go | NEW | Inspection, Timeline types |
| domain/diagnostics/diagnostics_repository.go | NEW | Repository interface |
| domain/diagnostics/diagnostics_errors.go | NEW | Domain errors |
| domain/device/device_repository.go | MODIFIED | Add stats query |

#### Application Layer (2 NEW)

| File | Status | Purpose |
|------|--------|---------|
| application/diagnostics/diagnostics_service.go | NEW | Inspection + timeline logic |
| application/diagnostics/diagnostics_dto.go | NEW | Request/response DTOs |

#### Handler Layer (3 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| api/handlers/diagnostics/diagnostics_inspect_handler.go | NEW | Inspection handler |
| api/handlers/diagnostics/diagnostics_timeline_handler.go | NEW | Timeline handler |
| api/handlers/diagnostics/diagnostics_routes.go | NEW | Route registration |
| api/handlers/router.go | MODIFIED | Wire diagnostics routes |

#### Infrastructure (2 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| infrastructure/storage/diagnostics_storage.go | NEW | Diagnostics queries |
| infrastructure/storage/migrations/ | NEW | SQL migrations |
| infrastructure/storage/device_storage.go | MODIFIED | Add stats query |

#### GraphQL (2 NEW, 2 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| api/graphql/schema/objects.go | MODIFIED | Add diagnostics types |
| api/graphql/schema/resolver.go | MODIFIED | Add resolvers |
| api/graphql/schema/schema.go | MODIFIED | Add queries |

---

## 12. Implementation Order

### Phase 1: Database (Day 1)
1. Create `device_events` table
2. Create telemetry stats view
3. Add columns to `devices` table
4. Test migrations

### Phase 2: Domain Layer (Day 1)
1. Create `domain/diagnostics/diagnostics_types.go`
2. Create `domain/diagnostics/diagnostics_repository.go`
3. Create `domain/diagnostics/diagnostics_errors.go`
4. Update `domain/device/device_repository.go`

### Phase 3: Infrastructure (Day 1-2)
1. Create `infrastructure/storage/diagnostics_storage.go`
2. Implement timeline events queries
3. Implement telemetry stats aggregation
4. Update `infrastructure/storage/device_storage.go`

### Phase 4: Application Layer (Day 2)
1. Create `application/diagnostics/diagnostics_service.go`
2. Implement `GetDeviceInspection`
3. Implement `GetDeviceTimeline`
4. Add cursor encoding/decoding

### Phase 5: Handlers (Day 2)
1. Create `handlers/diagnostics/diagnostics_inspect_handler.go`
2. Create `handlers/diagnostics/diagnostics_timeline_handler.go`
3. Wire routes
4. Test endpoints

### Phase 6: GraphQL (Day 2-3)
1. Add schema types
2. Add resolvers
3. Test queries

### Phase 7: Testing (Day 3)
1. Unit tests for service
2. Integration tests for handlers
3. E2E tests for diagnostics flow

---

*Document Version: 1.0*
*Status: Ready for Implementation*
