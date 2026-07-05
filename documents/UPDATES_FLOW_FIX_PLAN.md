# Updates Flow Fix Plan

**Date:** 2026-07-05  
**Based on:** UPDATES_FLOW_MISMATCH_REPORT.md  
**Status:** Ready for Implementation  

---

## Overview

This document outlines the implementation plan to fix the 5 issues identified in the Updates Flow Mismatch Report.

---

## Issue 1: FCM SilentWake Missing APK Payload

### Problem
FCM notifications don't include APK download information.

### Fix Implementation

#### Step 1: Update SilentWake struct
**File:** `internal/infrastructure/fcm/notifier.go`

```go
type SilentWake struct {
    Token          string
    Command        string
    CommandSecret  string
    DispatchID     string
    DeviceID       string
    Priority       string
    // NEW: APK info for updates
    APKFilename    string
    SHA256         string
    APKSize        int64
}
```

#### Step 2: Update FCM message builder
**File:** `internal/infrastructure/fcm/notifier.go`

```go
// In SendSilentWake(), add to Data map:
Data: map[string]string{
    "action":      "WAKE_DAEMON",
    "command":     wake.Command,
    "dispatch_id": wake.DispatchID,
    // NEW:
    "apkFilename": wake.APKFilename,
    "sha256":      wake.SHA256,
    "apkSize":     strconv.FormatInt(wake.APKSize, 10),
}
```

#### Step 3: Update dispatchUpdateCommand
**File:** `internal/application/updates/updates_push_service.go`

```go
wake := fcm.SilentWake{
    Token:       dev.FCMToken,
    Command:     UpdateCommandType,
    DispatchID:  cmdResp.DispatchID,
    DeviceID:    deviceID,
    Priority:    "high",
    // NEW:
    APKFilename: version.APKFilename,
    SHA256:      version.SHA256,
    APKSize:     version.APKSize,
}
```

---

## Issue 2: No Update Completion Endpoint

### Problem
No endpoint for device to report update progress/completion.

### Fix Implementation

#### Step 1: Create DTO
**File:** `internal/application/updates/updates_dto.go`

```go
// DeviceUpdateStatusRequest represents device callback for update status
type DeviceUpdateStatusRequest struct {
    DispatchID string `json:"dispatchId" binding:"required"`
    Status     string `json:"status" binding:"required"` // in_progress, completed, failed
    Error      string `json:"error,omitempty"`
}

// DeviceUpdateStatusResponse represents response to device callback
type DeviceUpdateStatusResponse struct {
    Acknowledged bool   `json:"acknowledged"`
    Message     string `json:"message,omitempty"`
}
```

#### Step 2: Add to updates entity
**File:** `internal/domain/updates/updates_entity.go`

```go
// Add to DevicePushStatus enum:
DevicePushStatusInProgress DevicePushStatus = "in_progress"
DevicePushStatusCompleted  DevicePushStatus = "completed"
```

#### Step 3: Create handler
**File:** `internal/api/handlers/updates/updates_device_status_handler.go`

```go
func (h *Handler) DeviceUpdateStatus(c *gin.Context) {
    var req updates.DeviceUpdateStatusRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
        return
    }
    
    // Find push device by dispatch_id
    // Update status in database
    // Check if all devices complete → update UpdatePush status
    
    c.JSON(http.StatusOK, updates.DeviceUpdateStatusResponse{
        Acknowledged: true,
    })
}
```

#### Step 4: Register route
**File:** `internal/api/handlers/updates/updates_routes.go`

```go
// Device callback endpoint (no auth - uses device auth)
public.POST("/v1/updates/device-status", h.DeviceUpdateStatus)
```

#### Step 5: Add service method
**File:** `internal/application/updates/updates_service.go`

```go
func (s *Service) UpdateDeviceStatus(ctx context.Context, req *DeviceUpdateStatusRequest) error {
    // Find UpdatePushDevice by dispatch_id (stored as PushDevice.ID)
    // Update status and timestamps
    // Check push completion
    return nil
}
```

#### Step 6: Add repository method
**File:** `internal/domain/updates/updates_repository.go`

```go
UpdateDeviceStatus(ctx context.Context, deviceID string, status DevicePushStatus, errorMsg string) error
```

---

## Issue 3: UpdatePush Status Never Transitions to COMPLETED

### Problem
Push stays "pending" forever, never becomes "completed".

### Fix Implementation

#### Option A: Synchronous completion check (simpler)

**File:** `internal/application/updates/updates_push_service.go`

After device status is updated, check if all devices are terminal:

```go
func (s *PushService) checkPushCompletion(ctx context.Context, pushID string) error {
    devices, err := s.repo.GetPushDevices(ctx, pushID)
    if err != nil {
        return err
    }
    
    allTerminal := true
    allSucceeded := true
    for _, d := range devices {
        if !d.IsTerminal() {
            allTerminal = false
        }
        if d.Status == DevicePushStatusFailed {
            allSucceeded = false
        }
    }
    
    if allTerminal {
        var newStatus updates.UpdateStatus
        if allSucceeded {
            newStatus = updates.UpdateStatusCompleted
        } else {
            newStatus = updates.UpdateStatusFailed
        }
        return s.repo.UpdatePushStatus(ctx, pushID, newStatus)
    }
    
    return nil
}
```

#### Option B: Background worker (more robust)

For production, consider a background worker that periodically checks:
- Pending pushes older than X minutes
- Mark as failed if devices don't respond

---

## Issue 4: DevicePushStatus Mismatch

### Problem
Frontend expects `IN_PROGRESS`, backend has `ACKNOWLEDGED`.

### Fix Implementation

#### Decision: Add both statuses to backend

**File:** `internal/domain/updates/updates_entity.go`

```go
// DevicePushStatus represents the status of a device in an update push.
type DevicePushStatus string

const (
    DevicePushStatusPending      DevicePushStatus = "pending"
    DevicePushStatusSent         DevicePushStatus = "sent"
    DevicePushStatusInProgress   DevicePushStatus = "in_progress"   // NEW
    DevicePushStatusAcknowledged DevicePushStatus = "acknowledged"
    DevicePushStatusCompleted    DevicePushStatus = "completed"     // NEW
    DevicePushStatusFailed       DevicePushStatus = "failed"
)
```

#### Update GraphQL schema
**File:** `api/graphql/schema/objects.go`

```graphql
enum DevicePushStatus {
  PENDING
  SENT
  IN_PROGRESS
  ACKNOWLEDGED
  COMPLETED
  FAILED
}
```

---

## Issue 5: SilentWake CommandSecret Unused

### Problem
`CommandSecret` field exists but is never populated.

### Fix Implementation

#### Clarification Required

The `CommandSecret` in SilentWake appears to be a leftover from device registration flow. For updates:

**Option A:** Remove it (not needed for updates)
- Updates use dispatch_id for tracking, not authorization
- Device is already authenticated when it receives FCM

**Option B:** Keep for future security enhancement
- If updates need additional authorization
- Device must prove identity to download APK

#### Recommendation
Remove `CommandSecret` from SilentWake for now to avoid confusion. If security is needed later, add a proper mechanism.

```go
type SilentWake struct {
    Token          string
    Command        string
    // CommandSecret  string  // REMOVE
    DispatchID     string
    DeviceID       string
    Priority       string
    APKFilename    string
    SHA256         string
    APKSize        int64
}
```

---

## Implementation Order

### Phase 1: Critical Fixes (Day 1)
1. Issue 1: Add APK fields to SilentWake + FCM
2. Issue 2: Create device status endpoint

### Phase 2: Status Completion (Day 2)
3. Issue 3: Add push completion logic
4. Issue 4: Align status enums

### Phase 3: Cleanup (Day 3)
5. Issue 5: Remove or clarify CommandSecret

---

## Testing Plan

### Unit Tests
- `SilentWake` JSON marshaling includes APK fields
- Device status endpoint validates input
- Push completion logic correctly transitions states

### Integration Tests
1. Push update → device receives FCM with APK info
2. Device calls status endpoint → status updates
3. All devices complete → push status becomes "completed"

### E2E Tests
1. Operator pushes update
2. Device receives, downloads, installs
3. Device reports completion
4. Operator sees "completed" status

---

## Files to Modify

| File | Changes |
|------|---------|
| `infrastructure/fcm/notifier.go` | Add APK fields to SilentWake |
| `application/updates/updates_push_service.go` | Populate APK fields in FCM |
| `domain/updates/updates_entity.go` | Add IN_PROGRESS, COMPLETED statuses |
| `application/updates/updates_dto.go` | Add device status DTOs |
| `application/updates/updates_service.go` | Add UpdateDeviceStatus method |
| `api/handlers/updates/` | New handler for device status |
| `api/handlers/updates/updates_routes.go` | Register device status route |
| `api/graphql/schema/objects.go` | Update DevicePushStatus enum |

---

## Rollback Plan

If issues arise:
1. Revert FCM changes first (most disruptive)
2. Keep device status endpoint but return mock responses
3. Monitor UpdatePush table for stuck records

---

*Document Version: 1.0*
*Status: Ready for Implementation*
