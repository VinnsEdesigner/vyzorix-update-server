# Spec Fix Plan: Resolving Frontend vs Backend Mismatches

> **Date:** 2026-07-05
> **Based on:** SPEC_MISMATCH_REPORT.md
> **Priority:** Critical issues must be fixed before implementation

---

## Overview

This document outlines the specific fixes required to align frontend and backend specification documents. Issues are grouped by priority and provide concrete action items.

---

## Critical Fixes (Must Fix Before Implementation)

### 1. Settings Path Fix (SP-001)

**Problem:** Frontend doc shows REST calls to `/v1/settings`, but backend uses `/v1/auth/me/settings`

**Files Affected:**
- `documents/SETTINGS_PAGE.md`
- `documents/SERVER_BACKEND_SETTINGS_API.md`

**Fix Required:**
Update `SETTINGS_PAGE.md` line ~246-249:
```typescript
// CURRENT (WRONG):
const BASE = "/v1/settings";  // ❌ Wrong

// FIXED:
const BASE = "/v1/auth/me";   // ✅ Correct
```

**API Endpoints:**
- `GET/PATCH /v1/auth/me/settings` (was `/v1/settings`)
- `GET/PATCH /v1/auth/me/thresholds` (was `/v1/settings/thresholds`)
- `GET/PATCH /v1/auth/me/notifications` (was `/v1/settings/notifications`)

---

### 2. Dashboard Command Cancel Path (DCL-001)

**Problem:** Frontend expects `DELETE /v1/device/:imei/command/:dispatchId`, backend provides `DELETE /v1/command/:dispatchId`

**Files Affected:**
- `documents/DASHBOARD_COMMANDS_LOGS.md`
- `documents/SERVER_BACKEND_DASHBOARD_COMMANDS_API.md`

**Fix Required:**
Update `DASHBOARD_COMMANDS_LOGS.md` REST API Specification section:
- Change cancel command endpoint from `DELETE /v1/device/:imei/command/:dispatchId` to `DELETE /v1/command/:dispatchId`
- Update the flow diagram to reflect correct path

**Backend Path (Canonical):**
```
DELETE /v1/command/:dispatchId
```

---

### 3. Authentication MFA Verify (AU-001)

**Problem:** Frontend expects `POST /v1/auth/mfa/verify` for post-login MFA, but backend has only `POST /v1/auth/mfa/verify-setup` for setup verification.

**Files Affected:**
- `documents/AUTHENTICATION_SYSTEM.md`
- `documents/AUTHENTICATION_SYSTEM_SERVER.md`

**Fix Required:**
Add missing endpoint to backend:
```
POST /v1/auth/mfa/verify
```
- Accepts: `{ temporaryToken: string, code: string }`
- Returns: `{ accessToken: string, refreshToken: string }` or error

Update frontend flow documentation to clarify:
- Setup MFA: Use `POST /v1/auth/mfa/verify-setup`
- Login MFA: Use `POST /v1/auth/mfa/verify`

---

### 4. Authentication Refresh Token (AU-002)

**Problem:** Frontend expects token refresh, backend has no refresh endpoint.

**Files Affected:**
- `documents/AUTHENTICATION_SYSTEM.md`
- `documents/AUTHENTICATION_SYSTEM_SERVER.md`

**Fix Required:**
Add missing endpoint to backend:
```
POST /v1/auth/refresh
```
- Accepts: `{ refreshToken: string }` in body or cookie
- Returns: `{ accessToken: string, refreshToken: string }`

---

## High Priority Fixes

### 5. Parameter Naming Standardization (DCL-002)

**Problem:** Backend uses `:id` for some device endpoints, frontend uses `:imei`

**Files Affected:**
- `documents/SERVER_BACKEND_DASHBOARD_COMMANDS_API.md`

**Fix Required:**
Change all device-specific endpoints from `:id` to `:imei`:
- `GET /v1/device/:id/commands/pending` → `GET /v1/device/:imei/commands/pending`
- `POST /v1/device/:id/command` → `POST /v1/device/:imei/command`

**Canonical Convention:** Use `:imei` for all device-specific REST endpoints

---

### 6. Settings Webhook Rotate (SP-002)

**Problem:** Backend has webhook rotate endpoint, frontend doesn't document it.

**Files Affected:**
- `documents/SETTINGS_PAGE.md`

**Fix Required:**
Add to Settings API documentation:
```
POST /v1/auth/me/notifications/webhook/rotate
```
- Auth: Cookie required
- Returns: `{ secret: string }`

---

### 7. Settings Reset (SP-003)

**Problem:** Backend has settings reset endpoint, frontend doesn't document it.

**Files Affected:**
- `documents/SETTINGS_PAGE.md`

**Fix Required:**
Add to Settings API documentation:
```
POST /v1/auth/me/settings/reset
```
- Auth: Cookie, requires super_admin role
- Returns: Default settings object

---

### 8. MFA Flow Clarification (AU-004)

**Problem:** MFA flow is unclear - setup vs login verification mixed up.

**Files Affected:**
- `documents/AUTHENTICATION_SYSTEM.md`
- `documents/AUTHENTICATION_SYSTEM_SERVER.md`

**Fix Required:**
Document clear MFA flows:

**Flow A: MFA Setup (First Time)**
1. User enables MFA → `POST /v1/auth/mfa/enroll`
2. User scans QR → `POST /v1/auth/mfa/verify-setup` with TOTP code
3. Server enables MFA → `POST /v1/auth/mfa/enable`
4. User gets backup codes → `POST /v1/auth/mfa/regenerate-backup-codes`

**Flow B: MFA Login**
1. User submits credentials → `POST /v1/auth/login`
2. Server returns `{ mfaRequired: true, temporaryToken: "..." }`
3. User submits TOTP → `POST /v1/auth/mfa/verify` with temporaryToken + code
4. Server returns `{ accessToken, refreshToken }`

---

### 9. Command Cancel REST Endpoint (DCL-004)

**Problem:** Backend doesn't document REST endpoint for canceling commands.

**Files Affected:**
- `documents/SERVER_BACKEND_DASHBOARD_COMMANDS_API.md`

**Fix Required:**
Add to REST API documentation:
```
DELETE /v1/command/:dispatchId
```
- Cancels a pending command
- Returns: `{ dispatchId, status: "cancelled" }`

---

## Medium Priority Fixes

### 10. Device Registration ID Naming (DR-001)

**Problem:** Inbox entry uses `id` in frontend, `ID` in backend.

**Files Affected:**
- `documents/DEVICE_REGISTRATION_SYSTEM.md`
- `documents/SERVER_BACKEND_DEVICE_REGISTRATION_API.md`

**Fix Required:**
Standardize on `id` (lowercase) for JSON responses in both docs.

---

### 11. Updates Cancel Endpoint (UP-001)

**Problem:** Frontend mentions cancel for scheduled updates but doesn't define the API.

**Files Affected:**
- `documents/UPDATES_PAGE.md`

**Fix Required:**
Add to API documentation:
```
POST /v1/updates/history/:id/cancel
```
- Cancels a pending/scheduled update push
- Returns: Updated push object with status "cancelled"

---

### 12. Settings Threshold Validation (SP-004)

**Problem:** Backend has validation rules, frontend doesn't document them.

**Files Affected:**
- `documents/SETTINGS_PAGE.md`

**Fix Required:**
Add Validation Rules section:
| Field | Range | Rule |
|-------|-------|------|
| riskWarn | 0-100 | Must be < riskCrit |
| riskCrit | 0-100 | Must be > riskWarn |
| thermalWarn | 0-100 | Must be < thermalCrit |
| thermalCrit | 0-100 | Must be > thermalWarn |
| bufferWarn | 0-100 | Must be > bufferCrit (inverted) |
| bufferCrit | 0-100 | Must be < bufferWarn (inverted) |

---

## Low Priority Fixes

### 13. Diagnostics RECONNECTED Event (DI-001)

**Problem:** Backend defines 13 event types, frontend documents 12.

**Files Affected:**
- `documents/DIAGNOSTICS_PAGE.md`

**Fix Required:**
Add `RECONNECTED` to Timeline Event Types table.

---

### 14. Updates Sync Status Values (UP-003)

**Problem:** Backend defines sync status values, frontend doesn't document.

**Files Affected:**
- `documents/UPDATES_PAGE.md`

**Fix Required:**
Add Sync Status Values table:
| Status | Description |
|--------|-------------|
| idle | No sync in progress |
| syncing | Currently syncing from GitHub |
| synced | Last sync successful |
| error | Last sync failed |

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

### Sprint 1: Critical Fixes (Week 1)
1. Fix Settings path mismatch (SP-001)
2. Fix Command cancel path (DCL-001)
3. Add MFA verify endpoint (AU-001)
4. Add refresh token endpoint (AU-002)

### Sprint 2: High Priority (Week 2)
5. Standardize parameter naming (DCL-002)
6. Add Settings webhook rotate (SP-002)
7. Add Settings reset (SP-003)
8. Clarify MFA flow (AU-004)
9. Add Command cancel REST (DCL-004)

### Sprint 3: Documentation Cleanup (Week 3)
10. Fix ID naming (DR-001)
11. Add Updates cancel (UP-001)
12. Add Threshold validation (SP-004)
13. Add Sync status (UP-003)
14. Add RECONNECTED event (DI-001)

---

## Verification Checklist

After making fixes, verify:

- [ ] Settings page FE doc uses `/v1/auth/me/*` paths
- [ ] Command cancel uses `/v1/command/:dispatchId`
- [ ] All device endpoints use `:imei` parameter
- [ ] MFA has both setup and verify endpoints documented
- [ ] Refresh token endpoint is documented
- [ ] All API paths match between FE and BE docs
- [ ] All field names are consistent
- [ ] Validation rules are documented where applicable

---

*Fix Plan Generated: 2026-07-05*
*Review with: Frontend Team, Backend Team*
