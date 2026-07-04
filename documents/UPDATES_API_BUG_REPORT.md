# Updates API Bug Report

> **Document Status:** Analysis Complete
> **Last Updated:** 2026-07-04
> **Project:** vyzorix-update-server
> **Severity:** Production Blocking

---

## Executive Summary

This document catalogs all bugs and issues found in the Updates API implementation that must be fixed before production release.

---

## 🔴 CRITICAL BUGS (Block Production)

### Bug 1: GraphQL `device` Field Returns Wrong Type

**Location:** `internal/api/graphql/resolver/updates_resolver.go` lines 42-46

**Severity:** CRITICAL - Frontend cannot parse response

**Current Code:**
```go
result := map[string]interface{}{
    // ...
    "device":      status.Device.CurrentVersion,  // ❌ BUG: Returns string
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

**Actual Response:**
```json
{
  "device": "2.2.0"
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

**Frontend Impact:** The Updates page cannot display device-specific update status.

---

### Bug 2: Hardcoded `releaseType` in GraphQL Resolver

**Location:** `internal/api/graphql/resolver/updates_resolver.go` lines 52, 100

**Severity:** HIGH - All versions show as "patch" type

**Current Code:**
```go
// Line 52:
"releaseType": "patch",  // ❌ HARDCODED!

// Line 100:
"releaseType": "patch",  // ❌ HARDCODED!
```

**Problem:** The `releaseType` field is hardcoded as `"patch"` for all versions instead of using the actual release type from the database.

**Fix Required:**
```go
// Use actual release type
"releaseType": strings.ToUpper(string(v.ReleaseType)),
```

**Note:** Need to add `"strings"` to imports if not present.

---

### Bug 3: GraphQL Schema Type Mismatch

**Location:** `internal/api/graphql/schema/objects.go` lines 956-962

**Severity:** CRITICAL - Schema doesn't match intended behavior

**Current Code:**
```go
"device": &graphql.Field{
    Type:        graphql.NewNonNull(graphql.String),  // ❌ WRONG TYPE
    Description: "Device current version",
},
"version": &graphql.Field{
    Type:        graphql.NewNonNull(graphql.String),
    Description: "Device current version (alias for device)",
},
```

**Problem:** The `device` field is defined as `String` type instead of an object type with `currentVersion` and `needsUpdate` fields.

**Fix Required:**
```go
// Add a new type or use inline object
"device": &graphql.Field{
    Type: graphql.NewObject(graphql.ObjectConfig{
        Name: "DeviceUpdateStatus",
        Fields: graphql.Fields{
            "currentVersion": &graphql.Field{Type: graphql.String},
            "needsUpdate":   &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
        },
    }),
    Description: "Device current version and update status",
},
```

---

## 🟡 MEDIUM PRIORITY ISSUES

### Issue 4: Missing `isLatest` Field in GraphQL Response

**Location:** `internal/api/graphql/resolver/updates_resolver.go`

**Problem:** The `UpdateVersion` type has an `isLatest` field, but the resolver doesn't return it.

**Current:**
```go
entry := map[string]interface{}{
    "id":           v.Version,
    "version":      v.Version,
    "releaseType":  "patch",  // Also hardcoded!
    // ❌ Missing: "isLatest": v.IsLatest,
    // ❌ Missing: "releaseDate": formatDateTimeInt64(v.ReleasedAt),
}
```

**Fix:** Add `isLatest` and `releaseDate` to the response map.

---

### Issue 5: Device Status Not Supported in GetStatus

**Location:** `internal/application/updates/updates_versions_status_service.go`

**Problem:** The `GetStatus` endpoint doesn't accept a `deviceId` parameter to return device-specific update status (`needsUpdate`).

**Current Signature:**
```go
func (s *VersionsStatusService) GetStatus(ctx context.Context) (*GetStatusResponse, error)
```

**Expected:** Should accept optional `deviceId` to compute `needsUpdate`:
```go
func (s *VersionsStatusService) GetStatus(ctx context.Context, deviceID string) (*GetStatusResponse, error)
```

**Impact:** `needsUpdate` in the GraphQL response will always be `false` or `nil`.

---

### Issue 6: SHA256 Fetch Failure is Silent

**Location:** `internal/infrastructure/github/github_sync.go` lines 105-114

**Severity:** MEDIUM - Security concern

**Current Code:**
```go
if !s.skipSHA256 {
    sha256, err := s.client.FetchAssetChecksum(ctx, release.TagName, asset.Name)
    if err != nil {
        // Log warning but don't fail - checksum is optional
        s.logger.Warn("Failed to fetch SHA256 for asset", ...)
        // ❌ Continues with empty SHA256!
    } else {
        versionAsset.SHA256 = sha256
    }
}
```

**Problem:** If SHA256 checksum fetch fails, the sync continues with an empty SHA256. This is a security concern because devices cannot verify APK integrity.

**Recommendation:** 
- Either fail the sync if SHA256 cannot be obtained
- Or at minimum, set a flag indicating checksum is unavailable
- Consider making SHA256 required for production

---

### Issue 7: Release Notes Fallback Logic

**Location:** `internal/infrastructure/github/github_sync.go` lines 162-167

**Problem:** If both `Body` (release notes) and `Name` are empty, `ReleaseNotes` will be empty.

```go
notes := release.Body
if notes == "" {
    notes = release.Name  // Fallback to name
}
// If both empty, notes == ""
```

**Impact:** Frontend may display empty release notes.

---

## 🟢 LOW PRIORITY / NOTES

### Note 8: GraphQL Schema Has Unused Fields

**Location:** `internal/api/graphql/schema/objects.go` lines 956-970

**Issue:** The `UpdateStatusType` has `version`, `apkFilename`, `sha256` fields that duplicate information from `latest`. These should be deprecated or removed.

---

### Note 9: Pagination Not Returned in GetVersions

**Location:** `internal/api/graphql/resolver/updates_resolver.go` lines 95-111

**Issue:** The GraphQL resolver returns `[]map[string]interface{}` instead of an object with pagination metadata.

**Current:**
```go
return result, nil  // Just array
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

## 📋 FIX PRIORITY

### Phase 1: Critical Fixes (Must Fix Before Release)
1. **Bug 1:** Fix `device` field in `GetUpdatesStatus` resolver - returns wrong type
2. **Bug 2:** Fix hardcoded `releaseType` - all versions show as "patch"
3. **Bug 3:** Fix GraphQL schema type for `device` field - should be object, not string

### Phase 2: High Priority
4. **Issue 4:** Add `isLatest` to resolver response
5. **Issue 5:** Add device ID support to GetStatus (for `needsUpdate` calculation)
6. **Bug 10:** Fix UpdateFCMToken to return fresh device data, not stale

### Phase 3: Medium Priority  
7. **Issue 6:** Handle SHA256 fetch failure (consider making required)
8. **Issue 7:** Improve release notes fallback

### Phase 4: Nice to Have
9. **Note 8:** Clean up unused schema fields
10. **Note 9:** Add pagination to GetVersions GraphQL response

---

### Bug 10: UpdateFCMToken Returns Stale Device Data

**Location:** `internal/api/graphql/resolver/mutation_resolver.go` lines 38-56

**Severity:** MEDIUM

**Current Code:**
```go
// Verify device ownership - returns *dto.DeviceResponse
dev, err := r.DeviceService.GetDeviceByOperator(ctx, deviceID, op.ID)
if err != nil {
    return nil, r.Presenter.NotFoundError("device not found")
}

// Update FCM token
err = r.DeviceService.UpdateFCMToken(ctx, deviceID, token)
if err != nil {
    return nil, r.Presenter.InternalError("failed to update FCM token")
}

// Return the updated device as a map
return r.deviceDTOToMap(dev), nil  // ❌ Returns STALE data!
```

**Problems:**
1. The resolver fetches the device BEFORE updating the token
2. The returned device doesn't reflect the new token
3. For proper confirmation, should return fresh data or at least indicate success

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

## 🔧 TEST CASES

### TC1: GetUpdatesStatus Returns Correct Device Object
```graphql
query {
  updatesStatus {
    sync { status lastSyncAt }
    latest { version apkFilename sha256 }
    device { currentVersion needsUpdate }
  }
}
```
**Expected:** `device` should be an object with `currentVersion` and `needsUpdate`

### TC2: GetUpdatesVersions Returns Correct ReleaseType
```graphql
query {
  updatesVersions {
    version
    releaseType  # Should be MAJOR, MINOR, or PATCH
    isLatest
  }
}
```
**Expected:** `releaseType` should vary based on actual release type

### TC3: PushUpdate Works with InstallType
```graphql
mutation {
  pushUpdate(input: {
    version: "2.2.1"
    deviceIds: ["device123"]
    installType: IMMEDIATE
  }) {
    id
    status
    installType
  }
}
```
**Expected:** `installType` should be stored and returned correctly

---

## 📝 REMAINING IMPLEMENTATION ITEMS

Based on `SERVER_BACKEND_UPDATES_API.md`, these features need verification:

| Feature | Status | Notes |
|---------|--------|-------|
| GET /v1/updates/status | ✅ Implemented | Has bugs (Bug 1, 2, 3) |
| GET /v1/updates/versions | ✅ Implemented | Missing pagination |
| GET /v1/updates/changelog | ✅ Implemented | |
| POST /v1/updates/push | ✅ Implemented | |
| GET /v1/updates/history | ✅ Implemented | |
| GET /v1/updates/export | ✅ Implemented | |
| POST /v1/updates/sync | ✅ Implemented | |
| FCM Push Notifications | ✅ Implemented | |
| Rate Limiting | ❓ Needs verification | |
| GitHub Webhook | ❓ Needs verification | |
| APK Binary Sync | ⚠️ Fetches metadata | SHA256 fetch can fail silently |

---

*Document Version: 1.0*
*Status: Bug Analysis Complete - Fixes Needed*
