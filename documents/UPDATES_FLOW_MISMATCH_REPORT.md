# Updates Flow Mismatch Report

**Date:** 2026-07-05  
**Status:** Issues Identified  
**Severity:** Critical  

---

## Executive Summary

The update push flow has **5 critical issues** that prevent proper end-to-end functionality:

1. **FCM SilentWake missing APK payload** - Device wakes but doesn't know where to download
2. **No update completion endpoint** - Backend can't track if update succeeded
3. **UpdatePush status never transitions to COMPLETED** - Push stays "in_progress" forever
4. **DevicePushStatus mismatch** - Frontend expects IN_PROGRESS, backend has ACKNOWLEDGED
5. **SilentWake CommandSecret unused** - Authorization field never populated

---

## Issue 1: FCM SilentWake Missing APK Payload

### Severity: CRITICAL

### Description
When `dispatchUpdateCommand` uses FCM fallback, it only sends:
```go
wake := fcm.SilentWake{
    Token:       dev.FCMToken,
    Command:     "CHECK_UPDATE",
    DispatchID:  cmdResp.DispatchID,
    DeviceID:    deviceID,
    Priority:    "high",
}
```

### Problem
Device receives wake notification but does NOT receive:
- `apkFilename` - Which file to download
- `sha256` - For integrity verification
- `apkSize` - For progress tracking

### Expected Behavior (from spec)
FCM data should contain:
```json
{
  "action": "WAKE_DAEMON",
  "command": "CHECK_UPDATE",
  "dispatch_id": "push_abc123",
  "apkFilename": "VyzorixAudioRouter-v2.2.1.apk",
  "sha256": "abc123...",
  "apkSize": "15728640"
}
```

### Fix Required
1. Add `APKFilename`, `SHA256`, `APKSize` to `SilentWake` struct
2. Populate these fields in `dispatchUpdateCommand`
3. Include in FCM message data

---

## Issue 2: No Update Completion Endpoint

### Severity: CRITICAL

### Description
After device receives `CHECK_UPDATE` command, there is **no endpoint** for device to report:
- Download started
- Download completed
- Install completed
- Install failed

### Code Evidence
```bash
# No device callback handler for update status
grep -rn "update.*status\|UpdateStatus" internal/api/handlers/ | grep -v sync
# Returns nothing related to device update callbacks
```

### Expected Behavior (from spec)
Device should call back to report progress:
```
POST /v1/updates/status (or similar)
{
  "dispatch_id": "push_abc123",
  "status": "in_progress|completed|failed",
  "error": "optional error message"
}
```

### Fix Required
1. Create `POST /v1/updates/device-status` endpoint
2. Update `UpdatePushDevice` status in database
3. Check if all devices complete → update `UpdatePush` to COMPLETED

---

## Issue 3: UpdatePush Status Never Transitions to COMPLETED

### Severity: HIGH

### Description
Looking at `updates_push_service.go`, the `UpdatePush.Status` is only set to:
- `UpdateStatusPending` on creation
- `UpdateStatusInProgress` is mentioned but never set

There is **no logic** to check if all devices completed and transition to:
- `UpdateStatusCompleted` when all devices succeed
- `UpdateStatusFailed` when all devices fail

### Code Evidence
```go
// In PushUpdate():
push := &updates.UpdatePush{
    // ...
    Status: updates.UpdateStatusPending,
}

// No status update to in_progress or completed anywhere
```

### Fix Required
1. Add status transition logic after all devices are processed
2. Or add background worker to check completion
3. Update `UpdatePush.Status` based on device outcomes

---

## Issue 4: DevicePushStatus Mismatch

### Severity: MEDIUM

### Frontend Spec (UPDATES_PAGE.md)
```typescript
export enum UpdateStatus {
  PENDING = "pending",
  IN_PROGRESS = "in_progress",
  COMPLETED = "completed",
  FAILED = "failed",
  CANCELLED = "cancelled",
}
```

### Backend GraphQL Schema (SERVER_BACKEND_UPDATES_API.md)
```graphql
enum DevicePushStatus {
  PENDING
  SENT
  ACKNOWLEDGED
  FAILED
}
```

### Problem
Frontend expects `IN_PROGRESS` status but backend only has:
- `PENDING` - Command queued
- `SENT` - Command dispatched
- `ACKNOWLEDGED` - Device received command
- `FAILED` - Command failed

**Missing:** `IN_PROGRESS` for when device is actively downloading/installing

### Fix Required
1. Add `DevicePushStatusInProgress` to enum
2. Update frontend to match backend enum OR vice versa
3. Device needs to report `IN_PROGRESS` status

---

## Issue 5: SilentWake CommandSecret Unused

### Severity: MEDIUM

### Description
`SilentWake` struct has a `CommandSecret` field:
```go
type SilentWake struct {
    Token          string
    Command        string
    CommandSecret  string  // <-- Never set!
    DispatchID     string
    DeviceID       string
    Priority string
}
```

But in `dispatchUpdateCommand`, this field is **never populated**.

### Question
Is `CommandSecret` needed for update authorization? The device registration flow uses `CommandSecret` for authorization. Should updates use a similar mechanism?

### Recommended Action
1. Clarify if updates need device authorization
2. If yes, populate `CommandSecret` 
3. If no, remove the field to avoid confusion

---

## Flow Diagram (Current)

```
Operator                    Backend                      Device
   │                           │                           │
   │  POST /v1/updates/push   │                           │
   │──────────────────────────>│                           │
   │                           │                           │
   │                           │  CHECK_UPDATE command     │
   │                           │  (WS or FCM)              │
   │                           │──────────────────────────>│
   │                           │                           │
   │                           │        ???                │
   │                           │<─────────────────────────│ (NO CALLBACK!)
   │                           │                           │
   │                           │                           │
   │  UpdatePush created      │                           │
   │  (Status: pending)       │                           │
   │                           │                           │
```

---

## Flow Diagram (Expected)

```
Operator                    Backend                      Device
   │                           │                           │
   │  POST /v1/updates/push   │                           │
   │──────────────────────────>│                           │
   │                           │                           │
   │                           │  CHECK_UPDATE + APK info  │
   │                           │  (WS or FCM)              │
   │                           │──────────────────────────>│
   │                           │                           │
   │                           │<─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │ Device reports:
   │                           │  POST /v1/updates/       │ - in_progress
   │                           │  device-status           │ - completed
   │                           │                           │ - failed
   │                           │                           │
   │                           │  UpdatePushDevice         │
   │                           │  updated in DB            │
   │                           │                           │
   │  UpdatePush updated       │                           │
   │  (completed when all     │                           │
   │   devices done)           │                           │
   │                           │                           │
```

---

## Fix Priority

| Priority | Issue | Effort |
|----------|-------|--------|
| P0 | Issue 1: FCM missing APK payload | Medium |
| P0 | Issue 2: No completion endpoint | Medium |
| P1 | Issue 3: Push status never completes | Low |
| P2 | Issue 4: Status enum mismatch | Low |
| P3 | Issue 5: CommandSecret clarification | Low |

---

## Recommended Actions

1. **Add APK fields to SilentWake struct**
2. **Create device status callback endpoint**
3. **Add UpdatePush completion logic**
4. **Align DevicePushStatus enum with frontend**
5. **Clarify CommandSecret usage**
