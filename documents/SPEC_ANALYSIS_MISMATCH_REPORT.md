# Spec Analysis: Frontend vs Backend Mismatch Report

> **Date:** 2026-07-05  
> **Analysis:** End-to-end flow verification between frontend specs and backend implementations

---

## Executive Summary

Analyzed 6 pairs of frontend/backend specification documents to identify flow mismatches, wrong logic, and missing implementations.

| Document Pair | Status | Issues Found |
|---------------|--------|--------------|
| DEVICE_REGISTRATION_SYSTEM.md vs SERVER_BACKEND_DEVICE_REGISTRATION_API.md | ✅ ALIGNED | None |
| DASHBOARD_COMMANDS_LOGS.md vs SERVER_BACKEND_DASHBOARD_COMMANDS_API.md | ✅ ALIGNED | None |
| UPDATES_PAGE.md vs SERVER_BACKEND_UPDATES_API.md | ⚠️ FIXED | Missing device callback endpoint |
| SETTINGS_PAGE.md vs SERVER_BACKEND_SETTINGS_API.md | ✅ ALIGNED | None |
| DIAGNOSTICS_PAGE.md vs SERVER_BACKEND_DIAGNOSTICS_API.md | ✅ ALIGNED | None |
| AUTHENTICATION_SYSTEM.md vs AUTHENTICATION_SYSTEM_SERVER.md | ✅ ALIGNED | None |

---

## Detailed Analysis

### 1. Device Registration (✅ ALIGNED)

**Flow:** Device → Inbox → Operator Ack → FCM → Device Confirm → Registered

**Findings:**
- Backend implements 5-state inbox model: `PENDING → ACKNOWLEDGED → APPROVING → APPROVED/REJECTED`
- This matches FE spec exactly
- State machine transitions are correctly implemented
- commandSecret generation and validation flow is correct

### 2. Dashboard Commands & Logs (✅ ALIGNED)

**Flow:** Command → Device → Status Update

**Findings:**
- Filter values match: `all`, `pending`, `delivered`, `completed`, `failed`
- Status defaults to `all`
- Pagination implemented correctly
- Log filtering by type works: `connection`, `command`, `telemetry`, `error`

### 3. Updates Page (⚠️ ISSUES FIXED)

**Flow:** Create Push → Send to Devices → Track Progress → Completion

**Issues Found:**

| Issue | Severity | Status |
|-------|----------|--------|
| Missing device callback endpoint `/v1/updates/device-status` | HIGH | ✅ FIXED |
| APK download URL not included in FCM payload | MEDIUM | ✅ FIXED |
| Device status values not aligned with FE expectations | MEDIUM | ✅ FIXED |
| Push completion logic needed refinement | LOW | ✅ FIXED |

**Fixes Applied:**

1. **Device Callback Endpoint** (FIXED)
   - Added `POST /v1/updates/device-status` handler
   - Accepts: `{dispatchId, deviceId, status, error}`
   - Status values: `in_progress`, `completed`, `failed`

2. **APK Download URL** (FIXED)
   - Added `DownloadURL` field to `SilentWake` struct
   - FCM payload now includes `downloadUrl` field
   - Device can construct full URL or use provided URL

3. **Device Status Tracking** (FIXED)
   - Added new push device statuses: `DOWNLOADING`, `INSTALLING`, `COMPLETED`
   - `IsTerminal()` method added to detect completion
   - Push status auto-updates when all devices complete

4. **Error Handling** (FIXED)
   - `checkPushCompletion` errors don't fail device callbacks
   - Device status updates succeed independently
   - Logging for push completion failures

### 4. Settings Page (✅ ALIGNED)

**Endpoints Verified:**
| Endpoint | Status |
|----------|--------|
| GET /v1/auth/me/settings | ✅ Exists |
| GET /v1/auth/me/thresholds | ✅ Exists |
| GET /v1/auth/me/notifications | ✅ Exists |
| POST /v1/auth/me/notifications/webhook/test | ✅ Exists |
| POST /v1/auth/me/notifications/webhook/rotate | ✅ Exists |

### 5. Diagnostics Page (✅ ALIGNED)

**Endpoints Verified:**
| Endpoint | Status |
|----------|--------|
| GET /v1/device/:imei/inspect | ✅ Exists |
| GET /v1/device/:imei/timeline | ✅ Exists |

### 6. Authentication (✅ ALIGNED)

**Flows Verified:**
- Login with credentials
- OAuth (Google)
- MFA support
- Token management
- Session handling

---

## Files Modified During Fixes

### New Files
```
apps/api/internal/api/handlers/updates/updates_device_status_handler.go  # Device callback handler
```

### Modified Files
```
apps/api/internal/api/api_server.go                    # Added PushService to ServerConfig
apps/api/internal/api/handlers/updates/updates_handler.go  # Added deviceStatusHandler
apps/api/internal/api/wire/providers.go                # Updated ProvideUpdatesHandler
apps/api/internal/api/wire/wire_handlers.go           # Updated NewUpdatesHandler call
apps/api/internal/application/updates/updates_push_service.go  # Added UpdateDeviceStatusByDispatch, checkPushCompletion
apps/api/internal/domain/updates/updates_entity.go     # Added new DevicePushStatus values
apps/api/internal/domain/updates/updates_repository.go # Added GetPushDeviceStatus method
apps/api/internal/infrastructure/fcm/notifier.go        # Added DownloadURL, device_id in FCM data
apps/api/internal/infrastructure/storage/updates_storage.go  # Added GetPushDeviceByPushAndDevice, UpdatePushDeviceStatusByDispatch
```

---

## Remaining Considerations

### 1. Documentation Update Needed

The `SERVER_BACKEND_UPDATES_API.md` document does NOT include the device callback endpoint specification:
- `POST /v1/updates/device-status`

This should be added to document the expected behavior.

### 2. Frontend Integration Required

The FE spec (`UPDATES_PAGE.md`) expects device callbacks but:
- The endpoint URL should be documented
- The device must implement the callback logic
- Device needs to send `dispatchId`, `deviceId`, and `status`

### 3. Test Coverage

Consider adding integration tests for:
- Device callback flow end-to-end
- Push completion detection
- Error handling during device callbacks

---

## Conclusion

**5 of 6 document pairs are fully aligned.**

**1 pair (Updates) had issues that were fixed during this session:**

1. Missing device callback endpoint - FIXED
2. APK download URL missing - FIXED  
3. Push completion logic refined - FIXED

The codebase now correctly implements the end-to-end update flow as specified in the frontend documentation.
