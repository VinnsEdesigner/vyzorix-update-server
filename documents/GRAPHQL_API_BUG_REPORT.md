# GraphQL API Bug Report

> **Document Status:** Analysis Complete
> **Last Updated:** 2026-07-04
> **Project:** vyzorix-update-server
> **Package:** `github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql`

---

## Executive Summary

This document catalogs all bugs found in the GraphQL API implementation. The GraphQL layer is responsible for:
- Query resolvers (devices, telemetry, commands, settings, updates)
- Mutation resolvers (device management, commands, settings)
- Subscription resolvers (real-time updates via WebSocket)
- Schema definitions (types, enums, scalars)

---

## 🔴 CRITICAL BUGS (Block Production)

### Bug 1: Updates Resolver - `device` Field Returns Wrong Type

**Location:** `internal/api/graphql/resolver/updates_resolver.go` lines 42-46

**Severity:** CRITICAL

**Current Code:**
```go
result := map[string]interface{}{
    // ...
    "device":      status.Device.CurrentVersion,  // ❌ Returns string
    "version":     status.Device.CurrentVersion,
    "apkFilename": nil,
    "sha256":      nil,
}
```

**Problem:** The `device` field returns a plain string (`"2.2.0"`) instead of an object with `currentVersion` and `needsUpdate` fields.

**Expected Response:**
```json
{
  "device": {
    "currentVersion": "2.2.0",
    "needsUpdate": true
  }
}
```

**Fix Required:**
```go
result := map[string]interface{}{
    // ...
    "device": map[string]interface{}{
        "currentVersion": status.Device.CurrentVersion,
        "needsUpdate":   status.Device.NeedsUpdate,
    },
    "version":     status.Device.CurrentVersion,
    "apkFilename": status.Latest.APKFilename,
    "sha256":      status.Latest.SHA256,
}
```

---

### Bug 2: Updates Resolver - Hardcoded `releaseType`

**Location:** `internal/api/graphql/resolver/updates_resolver.go` lines 52, 100

**Severity:** HIGH - All versions show as "patch" type

**Current Code:**
```go
"releaseType": "patch",  // ❌ HARDCODED!
```

**Fix Required:**
```go
"releaseType": strings.ToUpper(string(v.ReleaseType)),
```

---

### Bug 3: GraphQL Schema - `device` Field Wrong Type

**Location:** `internal/api/graphql/schema/objects.go` lines 956-962

**Severity:** CRITICAL

**Current Code:**
```go
"device": &graphql.Field{
    Type:        graphql.NewNonNull(graphql.String),  // ❌ WRONG TYPE
    Description: "Device current version",
},
```

**Fix Required:** Define proper `DeviceUpdateStatus` type with `currentVersion` and `needsUpdate` fields.

---

## 🟡 HIGH PRIORITY ISSUES

### Bug 4: UpdateFCMToken Returns Stale Device Data

**Location:** `internal/api/graphql/resolver/mutation_resolver.go` lines 38-56

**Severity:** MEDIUM

**Problem:** The resolver fetches device data BEFORE updating the FCM token, so the returned device doesn't reflect the new token.

**Current Code:**
```go
dev, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
// ... token update happens here ...
return r.deviceDTOToMap(dev), nil  // ❌ Returns STALE data!
```

**Fix Required:**
```go
// Update FCM token
err = r.DeviceService.UpdateFCMToken(ctx, deviceID, token)
if err != nil {
    return nil, r.Presenter.InternalError("failed to update FCM token")
}

// Fetch updated device to return fresh data
updatedDev, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
if err != nil {
    return nil, r.Presenter.NotFoundError("device not found")
}

return r.deviceDTOToMap(updatedDev), nil
```

---

### Bug 5: Missing `isLatest` in GetUpdatesVersions Response

**Location:** `internal/api/graphql/resolver/updates_resolver.go` lines 95-111

**Severity:** MEDIUM

**Problem:** The GraphQL resolver returns version data but omits `isLatest` field.

**Current Response:**
```json
{
  "id": "2.2.1",
  "version": "2.2.1",
  "releaseType": "patch"  // Also hardcoded!
  // ❌ Missing: isLatest
}
```

**Fix Required:** Add `isLatest: v.IsLatest` to the response map.

---

### Bug 6: GetVersions GraphQL Missing Pagination

**Location:** `internal/api/graphql/resolver/updates_resolver.go` lines 95-111

**Severity:** MEDIUM

**Problem:** The GraphQL query returns `[]map[string]interface{}` instead of an object with pagination metadata.

**Current:**
```go
return result, nil  // Just array, no pagination
```

**Expected (per spec):**
```go
return map[string]interface{}{
    "versions": result,
    "pagination": map[string]interface{}{
        "total": resp.Pagination.Total,
        "limit": limit,
        "offset": offset,
        "hasMore": offset+len(result) < resp.Pagination.Total,
    },
}, nil
```

---

## 🟢 MEDIUM PRIORITY ISSUES

### Issue 7: `RequireAuth` Returns Pointer to Context

**Location:** `internal/api/graphql/resolver/resolver.go` lines 102-110

**Severity:** LOW (design issue)

**Current Code:**
```go
func (r *Resolver) RequireAuth(ctx context.Context) (*context.Context, error) {
    op, ok := gqlcontext.GetOperator(ctx)
    if !ok || op == nil {
        return nil, r.Presenter.UnauthorizedError()
    }
    return &ctx, nil  // Returns pointer to context
}
```

**Issue:** This is unusual - callers don't typically need a pointer to context. This pattern is not used elsewhere.

---

### Issue 8: Unused `updateHistory` Query Field

**Location:** `internal/api/graphql/schema/schema.go` lines 247-255

**Severity:** LOW (dead code)

**Problem:** The `updateHistory` field is an alias for `updatesHistoryDetail` but returns `UpdateHistoryType` instead of `UpdatePushType`.

```go
"updateHistory": &graphql.Field{
    Type:        UpdateHistoryType,  // ❌ Different type!
    Resolve:     res.GetUpdatesHistoryDetail,
},
```

**Fix:** Either remove the duplicate or ensure both return the same type.

---

### Issue 9: Schema Field Indentation Inconsistency

**Location:** `internal/api/graphql/schema/objects.go` (InboxEntryType, InboxListResponseType, etc.)

**Severity:** LOW (code quality)

**Problem:** Some GraphQL type definitions have inconsistent indentation (appears to use spaces instead of tabs).

---

### Issue 10: TelemetryHistory Query Filters by `endTime` Instead of `startTime`

**Location:** `internal/api/graphql/resolver/query_resolver.go` lines 289-291

**Severity:** MEDIUM

**Current Code:**
```go
for _, entry := range entries {
    if entry.ReceivedAt.UnixMilli() <= endTime {  // ❌ Only checks endTime
        result = append(result, ...)
    }
}
```

**Problem:** The loop only checks `endTime` but ignores `startTime` for filtering. This could return all entries up to `endTime` instead of entries between `startTime` and `endTime`.

**Fix Required:**
```go
for _, entry := range entries {
    ts := entry.ReceivedAt.UnixMilli()
    if ts >= startTime && ts <= endTime {
        result = append(result, ...)
    }
}
```

---

### Bug 11: Missing `push` Field Parsing in UpdateMyNotifications

**Location:** `internal/api/graphql/resolver/mutation_resolver.go` lines 343-410

**Severity:** MEDIUM

**Problem:** The `UpdateMyNotifications` resolver parses `email` and `webhook` fields from the input but completely ignores the `push` field. The `NotificationInput` struct has a `Push *PushNotificationInput` field that is never populated.

**Current Code (lines 355-408):**
```go
if email, ok := input["email"].(map[string]interface{}); ok {
    // ... parses email
    notifInput.Email = emailInput
}

if webhook, ok := input["webhook"].(map[string]interface{}); ok {
    // ... parses webhook
    notifInput.Webhook = webhookInput
}
// ❌ MISSING: No parsing for "push" field!
```

**Fix Required:**
```go
if push, ok := input["push"].(map[string]interface{}); ok {
    pushInput := &domainoperator.PushNotificationInput{}
    if v, ok := push["thresholdBreach"].(bool); ok {
        pushInput.ThresholdBreach = &v
    }
    if v, ok := push["deviceOffline"].(bool); ok {
        pushInput.DeviceOffline = &v
    }
    if v, ok := push["deviceOnline"].(bool); ok {
        pushInput.DeviceOnline = &v
    }
    if v, ok := push["updateAvailable"].(bool); ok {
        pushInput.UpdateAvailable = &v
    }
    if v, ok := push["commandFailed"].(bool); ok {
        pushInput.CommandFailed = &v
    }
    if v, ok := push["registrationRequest"].(bool); ok {
        pushInput.RegistrationRequest = &v
    }
    notifInput.Push = pushInput
}
```

---

### Bug 12: `isLatest` Field Missing from UpdateVersionType Schema

**Location:** `internal/api/graphql/schema/objects.go` lines 847-890

**Severity:** MEDIUM

**Problem:** The `UpdateVersionType` GraphQL schema is missing the `isLatest` field. The spec requires this field to distinguish the latest version from previous versions.

**Current UpdateVersionType (excerpt):**
```go
Fields: graphql.Fields{
    "id": &graphql.Field{...},
    "version": &graphql.Field{...},
    "releaseType": &graphql.Field{...},
    "releaseNotes": &graphql.Field{...},
    "apkFilename": &graphql.Field{...},
    "apkSize": &graphql.Field{...},
    "sha256": &graphql.Field{...},
    "releasedAt": &graphql.Field{...},
    "createdAt": &graphql.Field{...},
    // ❌ Missing: "isLatest": &graphql.Field{Type: graphql.Boolean}
},
```

**Fix Required:** Add `isLatest` field to `UpdateVersionType`:
```go
"isLatest": &graphql.Field{
    Type:        graphql.NewNonNull(graphql.Boolean),
    Description: "Whether this is the latest version",
},
```

**Note:** The resolver at line 90 also needs to return `isLatest: v.IsLatest` in the response map.

---

### Bug 13: Missing `sent` Field in GetUpdatesHistory Response

**Location:** `internal/api/graphql/resolver/updates_resolver.go` lines 185-190

**Severity:** LOW

**Problem:** The `GetUpdatesHistory` resolver returns `pending`, `acknowledged`, and `failed` device counts but omits the `sent` count. The `HistoryDeviceCounts` DTO has all four fields including `Sent`.

**Current Code (lines 185-190):**
```go
entry := map[string]interface{}{
    // ...
    "pending":      push.Devices.Pending,
    "acknowledged": push.Devices.Acknowledged,
    "failed":       push.Devices.Failed,
    // ❌ MISSING: "sent": push.Devices.Sent
}
```

**Fix Required:**
```go
entry := map[string]interface{}{
    // ...
    "pending":      push.Devices.Pending,
    "sent":        push.Devices.Sent,
    "acknowledged": push.Devices.Acknowledged,
    "failed":       push.Devices.Failed,
}
```

**Note:** The GraphQL schema `PushHistoryEntryType` at line ~1085 also needs a `sent` field added to be consistent.

---

### Bug 14: `retryCommand` Returns Wrong Schema Type

**Location:** `internal/api/graphql/schema/schema.go` lines 372-380

**Severity:** MEDIUM

**Problem:** The `retryCommand` mutation is defined to return `CommandType` but the resolver returns `dispatchId`, `commandId`, and `status` which matches `CommandResultType`. This inconsistency could cause type mismatches.

**Current Code (schema):**
```go
"retryCommand": &graphql.Field{
    Type:        CommandType,  // ❌ Should be CommandResultType
    ...
}
```

**Fix Required:** Change to `CommandResultType` to match the resolver output.

---

### Bug 15: `cancelCommand` Returns Boolean Instead of Proper Response

**Location:** `internal/api/graphql/schema/schema.go` lines 382-391

**Severity:** MEDIUM

**Problem:** The `cancelCommand` mutation returns just `graphql.Boolean` (true/false) which provides no useful information about which command was cancelled. Clients have no way to know the cancellation timestamp or status.

**Current Code (schema):**
```go
"cancelCommand": &graphql.Field{
    Type:        graphql.Boolean,  // ❌ Should be a proper response type
    Description: "Cancel a pending command",
    ...
}
```

**Current Code (resolver):**
```go
func (r *Resolver) CancelCommand(p graphql.ResolveParams) (interface{}, error) {
    // ... cancellation logic ...
    return true, nil  // ❌ Returns boolean, no useful info
}
```

**Fix Required:** Define a `CancelCommandResponseType` and update both schema and resolver:
```go
// Schema
"cancelCommand": &graphql.Field{
    Type:        CancelCommandResponseType,
    Description: "Cancel a pending command",
    ...
}

// Resolver should return:
return map[string]interface{}{
    "dispatchId":  dispatchID,
    "cancelledAt": time.Now().UnixMilli(),
    "status":      "cancelled",
}, nil
```

---

### Bug 16: Missing `deviceOnline` in SendCommand Response Schema

**Location:** `internal/api/graphql/schema/objects.go` lines 356-378 (`CommandResultType`)

**Severity:** LOW

**Problem:** The `SendCommand` resolver returns `deviceOnline` field in its response (line 176), but `CommandResultType` schema doesn't define this field. This means the response is missing information the resolver provides.

**Current SendCommand resolver return (line 174-177):**
```go
return map[string]interface{}{
    "dispatchId":   cmdResp.DispatchID,
    "commandId":    cmdResp.CommandID,
    "status":       delivery,
    "deviceOnline": delivery == "sent",  // ❌ Not in schema
}, nil
```

**Fix Required:** Add `deviceOnline` field to `CommandResultType`:
```go
"deviceOnline": &graphql.Field{
    Type:        graphql.NewNonNull(graphql.Boolean),
    Description: "Whether device was online when command was sent",
},
```

---

### Bug 17: FCM Fallback Missing Device Ownership Check

**Location:** `internal/api/graphql/resolver/mutation_resolver.go` lines 153

**Severity:** MEDIUM

**Problem:** When sending a command via FCM fallback, the code uses `GetDevice` without verifying device ownership. This could allow operators to send commands to devices they don't own.

**Current Code:**
```go
// If not sent via WebSocket, try FCM
if delivery == "queued" && r.FCMNotifier != nil {
    dev, _ := r.DeviceService.GetDevice(ctx, deviceID)  // ❌ No ownership check
    if dev != nil && dev.FCMToken != "" {
        // ... send FCM
    }
}
```

**Fix Required:**
```go
// If not sent via WebSocket, try FCM
if delivery == "queued" && r.FCMNotifier != nil {
    // First verify ownership
    _, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
    if err != nil {
        // Device not owned by operator, don't try FCM
        return nil, r.Presenter.NotFoundError("device not found")
    }
    dev, _ := r.DeviceService.GetDevice(ctx, deviceID)
    if dev != nil && dev.FCMToken != "" {
        // ... send FCM
    }
}
```

---

### Bug 18: DeregisterDeviceGraphQL Missing Operator ID in Service Call

**Location:** `internal/api/graphql/resolver/inbox_resolver.go` lines 208-214

**Severity:** MEDIUM

**Problem:** The `DeregisterDeviceGraphQL` resolver passes `op.ID` to `DeregisterDeviceByOperator`, but the `hard` parameter is used to determine if it's a hard or soft delete. However, looking at the method signature, this might not properly verify ownership.

**Current Code:**
```go
result, err := r.DeviceService.DeregisterDeviceByOperator(ctx, imei, op.ID, hard)
if err != nil {
    if err == device.ErrNotFound {
        return nil, r.Presenter.NotFoundError("device not found or not owned by operator")
    }
    return nil, r.Presenter.InternalError("failed to deregister device")
}
```

**Issue:** The error message says "device not found or not owned by operator" but the underlying service might not actually verify operator ownership in all code paths. Need to verify `DeregisterDeviceByOperator` properly checks ownership.

---

## 📋 RESOLVER FILE ANALYSIS

### Files Checked

| File | Lines | Status |
|------|-------|--------|
| `resolver.go` | 110 | ⚠️ 1 design issue |
| `query_resolver.go` | 991 | 🔴 1 bug (telemetry filtering) |
| `mutation_resolver.go` | 457 | 🟡 1 bug (stale data) |
| `updates_resolver.go` | 421 | 🔴 3 bugs (device type, releaseType, isLatest) |
| `inbox_resolver.go` | 228 | ✅ OK |
| `subscription_resolver.go` | 34 | ✅ OK (intentional stubs) |
| `helpers.go` | 72 | ✅ OK |

### Schema Files

| File | Lines | Status |
|------|-------|--------|
| `schema.go` | 474 | ⚠️ 1 duplicate field issue |
| `objects.go` | 1697 | 🔴 1 critical type issue |
| `enums.go` | 198 | ✅ OK |
| `scalars.go` | 41 | ✅ OK |
| `subscription.go` | 49 | ✅ OK |

### Support Files

| File | Status |
|------|--------|
| `server.go` | ✅ OK |
| `handler/handler.go` | ✅ OK |
| `handler/presenter.go` | ✅ OK |
| `middleware/auth.go` | ✅ OK |
| `context/context.go` | ✅ OK |
| `errors/errors.go` | ✅ OK |
| `validator/validator.go` | ✅ OK |
| `subscription/handler.go` | ✅ OK |
| `subscription/client.go` | ⚠️ Needs review (WebSocket logic) |

---

## 📋 FIX PRIORITY

### Phase 1: Critical Fixes (Must Fix Before Release)
1. **Bug 1:** Fix `device` field in `GetUpdatesStatus` resolver - returns wrong type
2. **Bug 2:** Fix hardcoded `releaseType` - all versions show as "patch"
3. **Bug 3:** Fix GraphQL schema type for `device` field

### Phase 2: High Priority
7. **Bug 11:** Add missing `push` field parsing in UpdateMyNotifications
8. **Bug 12:** Add `isLatest` field to UpdateVersionType schema
4. **Bug 4:** Fix UpdateFCMToken to return fresh device data
5. **Bug 5:** Add `isLatest` to GetUpdatesVersions response
6. **Bug 6:** Add pagination to GetUpdatesVersions GraphQL response

### Phase 3: Medium Priority  
7. **Bug 10:** Fix TelemetryHistory to filter by both startTime AND endTime
8. **Issue 8:** Remove duplicate `updateHistory` field or fix type

### Phase 4: Low Priority (Nice to Have)
9. **Issue 7:** Review `RequireAuth` interface
10. **Issue 9:** Fix schema indentation

---

## 🔧 TEST CASES

### TC1: GetUpdatesStatus Returns Correct Device Object
```graphql
query {
  updatesStatus {
    sync { status lastSyncAt }
    latest { version apkFilename sha256 releaseType isLatest }
    device { currentVersion needsUpdate }
  }
}
```
**Expected:** `device` should be an object with `currentVersion` and `needsUpdate`

### TC2: GetUpdatesVersions Returns Correct ReleaseType and isLatest
```graphql
query {
  updatesVersions {
    version
    releaseType  # Should be MAJOR, MINOR, or PATCH
    isLatest     # Should be true for latest only
  }
}
```

### TC3: TelemetryHistory Filters Correctly
```graphql
query {
  telemetryHistory(deviceId: "123", startTime: 1718890000000, endTime: 1718900000000) {
    id
    receivedAt
  }
}
```
**Expected:** Only entries between startTime and endTime

### TC4: UpdateFCMToken Returns Updated Token
```graphql
mutation {
  updateFCMToken(deviceId: "123", token: "new-token") {
    id
    fcmToken  # Should show new token
  }
}
```

---

## 📝 NOTES

### Subscription Resolvers
The subscription resolvers (`DeviceUpdated`, `TelemetryReceived`, `CommandStatusChanged`) are intentionally stubbed out since subscriptions are handled via WebSocket directly. This is by design.

### Service Availability Checks
All resolvers properly check for nil services before use, returning appropriate errors if services are unavailable.

### Error Handling
The presenter pattern is consistently used across all resolvers for error handling:
- `BadRequestError` for validation errors
- `UnauthorizedError` for auth failures
- `NotFoundError` for missing resources
- `InternalError` for server errors

---

*Document Version: 1.0*
*Status: Bug Analysis Complete*
