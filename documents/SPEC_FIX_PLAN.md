# Spec Fix Plan: Resolving Frontend vs Backend Mismatches

> **Date:** 2026-07-05
> **Based on:** SPEC_MISMATCH_REPORT.md
> **Status:** UPDATED - Most endpoints are IMPLEMENTED, fixes are primarily documentation

---

## Overview

After thorough code review, **most endpoints are already implemented** in the Go code. The issues identified are primarily:
1. **Documentation mismatches** between FE and BE specs
2. **One actual implementation gap**: Metrics export endpoint

This document outlines what needs to be fixed in docs and the one real code gap.

---

## Critical Fixes (Documentation Only - Code Already Implemented)

> **Note:** The code already implements these endpoints correctly. The fixes below are to UPDATE THE DOCUMENTATION to match reality.

### 1. Settings Path Fix (SP-001) - DOCUMENTATION ONLY

**Problem:** Frontend doc shows REST calls to `/v1/settings`, but backend uses `/v1/auth/me/settings`

**Status:** ✅ IMPLEMENTED in code (auth_settings.go line 142-149)

**Files to Update:**
- `documents/SETTINGS_PAGE.md` - Change BASE from `/v1/settings` to `/v1/auth/me`

---

### 2. Dashboard Command Cancel Path (DCL-001) - DOCUMENTATION ONLY

**Problem:** Frontend doc shows `DELETE /v1/device/:imei/command/:dispatchId`, but actual path is `DELETE /v1/command/:dispatchId`

**Status:** ✅ IMPLEMENTED in code (server_routes.go line 211, command_execute.go line 271-312)

**Files to Update:**
- `documents/DASHBOARD_COMMANDS_LOGS.md` - Correct the cancel endpoint path

---

### 3. Authentication MFA Verify (AU-001) - DOCUMENTATION ONLY

**Problem:** Frontend doesn't document MFA verify endpoint properly

**Status:** ✅ IMPLEMENTED in code (auth_routes.go line 183)

**Files to Update:**
- `documents/AUTHENTICATION_SYSTEM.md` - Document `POST /v1/auth/mfa/verify` for post-login

---

### 4. Authentication Refresh Token (AU-002) - DOCUMENTATION ONLY

**Problem:** Frontend doesn't document refresh endpoint

**Status:** ✅ IMPLEMENTED in code (auth_routes.go line 126, auth_refresh.go)

**Files to Update:**
- `documents/AUTHENTICATION_SYSTEM.md` - Document `POST /v1/auth/refresh` endpoint

---

## High Priority Fixes

### 5. Parameter Naming Standardization (DCL-002) - ✅ FIXED

**Problem:** Backend uses `:id` for some device endpoints, frontend uses `:imei`

**Status:** ✅ FIXED - All endpoints now use `:imei`

**Files Updated:**
- `apps/api/internal/api/server_routes.go` - Changed all `:id` to `:imei`
- `apps/api/internal/api/handlers/device/dev_status.go` - Uses `:imei`
- `apps/api/internal/api/handlers/device/device_list.go` - Uses `:imei`
- `apps/api/internal/api/handlers/device/device_updater.go` - Uses `:imei`
- `apps/api/internal/api/handlers/command/command_execute.go` - Uses `:imei`
- `documents/SERVER_BACKEND_DASHBOARD_COMMANDS_API.md` - Fixed `:id` → `:imei`
- `documents/SERVER_BACKEND_DIAGNOSTICS_API.md` - Fixed `:id` → `:imei`
- `documents/SERVER_BACKEND_UPDATES_API.md` - Fixed `:id` → `:imei`
- `documents/MULTI_CLIENT_API_KEY_SYSTEM.md` - Fixed `:id` → `:imei`

**Canonical Convention:** Use `:imei` for all device-specific REST endpoints ✅

---

### 6. Settings Webhook Rotate (SP-002) - DOCUMENTATION ONLY

**Problem:** Backend has webhook rotate endpoint, frontend doesn't document it.

**Status:** ✅ IMPLEMENTED in code (auth_settings.go line 537-570)

**Files to Update:**
- `documents/SETTINGS_PAGE.md` - Add `POST /v1/auth/me/notifications/webhook/rotate`

---

### 7. Settings Reset (SP-003) - DOCUMENTATION ONLY

**Problem:** Backend has settings reset endpoint, frontend doesn't document it.

**Status:** ✅ IMPLEMENTED in code (auth_settings.go line 221-222)

**Files to Update:**
- `documents/SETTINGS_PAGE.md` - Add `POST /v1/auth/me/settings/reset`

---

### 8. MFA Flow Clarification (AU-004) - DOCUMENTATION ONLY

**Problem:** MFA flow is unclear - setup vs login verification mixed up.

**Status:** Both endpoints exist in code

**Files to Update:**
- `documents/AUTHENTICATION_SYSTEM.md` - Document both flows clearly

---

## Medium Priority Fixes

### 9. Updates Cancel Endpoint (UP-001) - DOCUMENTATION ONLY

**Problem:** Frontend mentions cancel for scheduled updates but doesn't define the API.

**Status:** ✅ IMPLEMENTED in code (updates_history_handler.go line 84-129)

**Files to Update:**
- `documents/UPDATES_PAGE.md` - Add `POST /v1/updates/history/:pushId/cancel`

---

### 10. Settings Threshold Validation (SP-004) - DOCUMENTATION ONLY

**Problem:** Backend has validation rules, frontend doesn't document them.

**Status:** ✅ IMPLEMENTED in code (auth_settings.go line 293-305)

**Files to Update:**
- `documents/SETTINGS_PAGE.md` - Add Validation Rules section

---

## REAL IMPLEMENTATION GAPS

### 11. Metrics Export - ACTUAL CODE GAP

**Problem:** `GET /v1/device/:imei/metrics/export` specified in BE doc but NOT in code

**Status:** ❌ MISSING

**File to Create/Update:**
- `internal/api/handlers/device/device_metrics_handler.go` - Add ExportMetrics method

**Route to Add in server_routes.go:**
```go
deviceMgmt.GET("/:imei/metrics/export", s.deviceMetricsHandler.ExportMetrics)
```

---

### 12. Inbox DELETE Route - CLARIFICATION NEEDED

**Problem:** Code uses PATCH (UpdateInboxEntry) but route might be intended as DELETE

**Status:** ⚠️ Route exists but implementation may be wrong

**File to Review:**
- `internal/api/handlers/inbox/inbox_routes.go` line 12

**Options:**
1. Implement actual DELETE functionality
2. Change route to match implementation (PATCH)

---

## Files Summary for Changes

### Frontend Documents to Update:
1. `SETTINGS_PAGE.md` - Fix paths, add endpoints, add validation rules
2. `DASHBOARD_COMMANDS_LOGS.md` - Fix command cancel path, update flow
3. `AUTHENTICATION_SYSTEM.md` - Clarify MFA flow, add refresh endpoint
4. `UPDATES_PAGE.md` - Add cancel endpoint, add sync status values
5. `DIAGNOSTICS_PAGE.md` - Add RECONNECTED event type

### Backend Documents to Update:
1. `SERVER_BACKEND_SETTINGS_API.md` - Already aligned, no changes needed
2. `SERVER_BACKEND_DASHBOARD_COMMANDS_API.md` - Add REST cancel endpoint, standardize :imei
3. `SERVER_BACKEND_AUTHENTICATION_API.md` (new or existing) - Add mfa/verify, add refresh
4. `SERVER_BACKEND_DIAGNOSTICS_API.md` - Already aligned, minor addition of RECONNECTED

---

## Implementation Order

> **Note:** Most endpoints are already IMPLEMENTED. The work is primarily documentation updates.

### Sprint 1: Documentation Updates Only (Week 1)
1. Update SETTINGS_PAGE.md with correct paths
2. Update DASHBOARD_COMMANDS_LOGS.md with correct cancel path
3. Update AUTHENTICATION_SYSTEM.md with MFA verify and refresh endpoints

### Sprint 2: Documentation Additions (Week 2)
4. Add missing endpoints to FE docs (webhook rotate, settings reset, updates cancel)
5. Add threshold validation rules to Settings doc
6. Add sync status values to Updates doc
7. Add RECONNECTED event to Diagnostics doc
8. Standardize parameter naming across docs (`:imei` vs `:id`)

### Sprint 3: Code Implementation Gap (Week 2-3)
9. **Add Metrics Export Endpoint** (REAL CODE GAP)
   - Create `ExportMetrics` method in device_metrics_handler.go
   - Add route in server_routes.go
   - Document in both BE and FE specs

---

## Verification Checklist

After making fixes, verify:

- [ ] Settings page FE doc uses `/v1/auth/me/*` paths
- [ ] Command cancel uses `/v1/command/:dispatchId`
- [ ] All device endpoints use `:imei` parameter
- [ ] All API paths match between FE and BE docs
- [ ] Metrics export endpoint exists in code
- [ ] All validation rules are documented
- [ ] MFA has both setup and verify endpoints documented
- [ ] Refresh token endpoint is documented
- [ ] All field names are consistent
- [ ] Validation rules are documented where applicable

---

*Fix Plan Generated: 2026-07-05*
*Review with: Frontend Team, Backend Team*
