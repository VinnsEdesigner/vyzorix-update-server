# Spec Mismatch Report: Frontend vs Backend Documentation vs Implementation Analysis

> **Date:** 2026-07-05
> **Project:** Vyzorix Update Server
> **Purpose:** Identify mismatches, wrong logics, and inconsistencies between frontend specs, backend specs, and actual implementation

---

## Executive Summary

This report analyzes 6 pairs of Frontend (FE) and Backend (BE) specification documents AND cross-references them with the actual Go implementation to identify:
- Endpoint mismatches between FE and BE specs
- Implementation gaps (spec says should exist but code doesn't)
- Implementation extras (code exists but spec doesn't mention)
- Parameter naming inconsistencies
- Logic discrepancies

---

## IMPLEMENTATION STATUS OVERVIEW

### Authentication (AUTHENTICATION_SYSTEM.md vs AUTHENTICATION_SYSTEM_SERVER.md)
| Endpoint | Spec Status | Implementation Status |
|----------|-------------|----------------------|
| POST /v1/auth/login | Specified | ✅ IMPLEMENTED |
| POST /v1/auth/register | Specified | ✅ IMPLEMENTED |
| POST /v1/auth/logout | Specified | ✅ IMPLEMENTED |
| GET /v1/auth/me | Specified | ✅ IMPLEMENTED |
| PATCH /v1/auth/me | Specified | ✅ IMPLEMENTED |
| POST /v1/auth/refresh | Specified | ✅ IMPLEMENTED (RefreshHandler) |
| GET /v1/auth/mfa/status | Specified | ✅ IMPLEMENTED |
| POST /v1/auth/mfa/enroll | Specified | ✅ IMPLEMENTED |
| POST /v1/auth/mfa/verify-setup | Specified | ✅ IMPLEMENTED |
| POST /v1/auth/mfa/enable | Specified | ✅ IMPLEMENTED |
| POST /v1/auth/mfa/disable | Specified | ✅ IMPLEMENTED |
| POST /v1/auth/mfa/verify-backup | Specified | ✅ IMPLEMENTED |
| POST /v1/auth/mfa/regenerate-backup-codes | Specified | ✅ IMPLEMENTED |
| POST /v1/auth/mfa/verify | Specified | ✅ IMPLEMENTED (line 183 in auth_routes.go) |
| POST /v1/auth/forgot-password | Specified | ✅ IMPLEMENTED |
| POST /v1/auth/reset-password | Specified | ✅ IMPLEMENTED |
| GET /v1/auth/google | Specified | ✅ IMPLEMENTED |
| GET /v1/auth/google/callback | Specified | ✅ IMPLEMENTED |
| GET /v1/auth/github | Specified | ✅ IMPLEMENTED |
| GET /v1/auth/github/callback | Specified | ✅ IMPLEMENTED |

### Device Management (DEVICE_REGISTRATION_SYSTEM.md vs SERVER_BACKEND_DEVICE_REGISTRATION_API.md)
| Endpoint | Spec Status | Implementation Status |
|----------|-------------|----------------------|
| POST /v1/device/inbox | Specified | ✅ IMPLEMENTED (public route) |
| GET /v1/device/inbox | Specified | ✅ IMPLEMENTED |
| GET /v1/device/inbox/:imei | Specified | ✅ IMPLEMENTED |
| POST /v1/device/inbox/:imei/ack | Specified | ✅ IMPLEMENTED |
| DELETE /v1/device/inbox/:imei | Specified | ⚠️ Uses PATCH (UpdateInboxEntry) |
| GET /v1/devices | Specified | ✅ IMPLEMENTED |
| GET /v1/devices/:imei | Specified | ✅ IMPLEMENTED |
| DELETE /v1/devices/:imei | Specified | ✅ IMPLEMENTED |

### Dashboard/Commands (DASHBOARD_COMMANDS_LOGS.md vs SERVER_BACKEND_DASHBOARD_COMMANDS_API.md)
| Endpoint | Spec Status | Implementation Status |
|----------|-------------|----------------------|
| POST /v1/device/:id/command | Specified | ✅ IMPLEMENTED (uses :id) |
| GET /v1/device/:id/commands/pending | Specified | ✅ IMPLEMENTED (uses :id) |
| GET /v1/device/:imei/commands | Specified | ✅ IMPLEMENTED (command_history_handler.go) |
| GET /v1/device/:imei/logs | Specified | ✅ IMPLEMENTED (device_logs_handler.go) |
| GET /v1/device/:imei/metrics | Specified | ✅ IMPLEMENTED (device_metrics_handler.go) |
| GET /v1/device/:imei/telemetry | Specified | ✅ IMPLEMENTED (device_telemetry_handler.go) |
| GET /v1/device/:imei/metrics/export | Specified | ⚠️ MISSING from code |
| GET /v1/dashboard/stats | Specified | ✅ IMPLEMENTED |
| DELETE /v1/command/:dispatchId | Specified | ✅ IMPLEMENTED |

### Updates (UPDATES_PAGE.md vs SERVER_BACKEND_UPDATES_API.md)
| Endpoint | Spec Status | Implementation Status |
|----------|-------------|----------------------|
| GET /v1/updates/status | Specified | ✅ IMPLEMENTED |
| GET /v1/updates/versions | Specified | ✅ IMPLEMENTED |
| GET /v1/updates/changelog | Specified | ✅ IMPLEMENTED |
| POST /v1/updates/push | Specified | ✅ IMPLEMENTED |
| GET /v1/updates/history | Specified | ✅ IMPLEMENTED |
| GET /v1/updates/export | Specified | ✅ IMPLEMENTED |
| POST /v1/updates/sync | Specified | ✅ IMPLEMENTED |
| POST /v1/updates/history/:pushId/cancel | Specified | ✅ IMPLEMENTED |

### Diagnostics (DIAGNOSTICS_PAGE.md vs SERVER_BACKEND_DIAGNOSTICS_API.md)
| Endpoint | Spec Status | Implementation Status |
|----------|-------------|----------------------|
| GET /v1/device/:imei/inspect | Specified | ✅ IMPLEMENTED |
| GET /v1/device/:imei/timeline | Specified | ✅ IMPLEMENTED |

### Settings (SETTINGS_PAGE.md vs SERVER_BACKEND_SETTINGS_API.md)
| Endpoint | Spec Status | Implementation Status |
|----------|-------------|----------------------|
| GET /v1/auth/me/settings | Specified | ✅ IMPLEMENTED |
| PATCH /v1/auth/me/settings | Specified | ✅ IMPLEMENTED |
| POST /v1/auth/me/settings/reset | Specified | ✅ IMPLEMENTED (via UpdateSettings with reset flag) |
| GET /v1/auth/me/thresholds | Specified | ✅ IMPLEMENTED |
| PATCH /v1/auth/me/thresholds | Specified | ✅ IMPLEMENTED |
| GET /v1/auth/me/notifications | Specified | ✅ IMPLEMENTED |
| PATCH /v1/auth/me/notifications | Specified | ✅ IMPLEMENTED |
| POST /v1/auth/me/notifications/webhook/test | Specified | ✅ IMPLEMENTED |
| POST /v1/auth/me/notifications/webhook/rotate | Specified | ✅ IMPLEMENTED |

---

## 1. DEVICE_REGISTRATION_SYSTEM.md vs SERVER_BACKEND_DEVICE_REGISTRATION_API.md

### 1.1 Status: ✅ ALIGNED (Minor Issues)

### 1.2 Endpoints Comparison

| Frontend Expects | Backend Provides | Status |
|-----------------|------------------|--------|
| POST /v1/device/inbox | POST /v1/device/inbox | ✅ Match |
| GET /v1/device/inbox | GET /v1/device/inbox | ✅ Match |
| GET /v1/device/inbox/:imei | GET /v1/device/inbox/:imei | ✅ Match |
| POST /v1/device/inbox/:imei/ack | POST /v1/device/inbox/:imei/ack | ✅ Match |
| DELETE /v1/device/inbox/:imei | DELETE /v1/device/inbox/:imei | ✅ Match |
| POST /v1/device/register | POST /v1/device/register | ✅ Match |
| POST /v1/device/confirm | POST /v1/device/confirm | ✅ Match |
| DELETE /v1/device/:imei | DELETE /v1/device/:imei | ✅ Match |
| GET /v1/devices | GET /v1/devices | ✅ Match |
| GET /v1/devices/:imei | GET /v1/devices/:imei | ✅ Match |

### 1.3 Mismatches Found

| ID | Issue | Severity | Description |
|----|-------|----------|-------------|
| DR-001 | **Inconsistent ID Field** | Medium | FE doc shows `id` for inbox entries (line 181), BE doc uses `ID` (line 182). Go uses uppercase in struct tags. Need consistent naming convention. |
| DR-002 | **GraphQL vs REST Naming** | Low | FE uses `device_imei` in GraphQL fragments, BE uses `imei`. Minor inconsistency in naming. |
| DR-003 | **Missing Endpoint in FE** | Low | BE doc mentions `POST /v1/device/register` as "PARTIAL" needing updates, FE doesn't document this endpoint clearly. |

### 1.4 Logic Issues
- None identified - flow is consistent between documents

---

## 2. DASHBOARD_COMMANDS_LOGS.md vs SERVER_BACKEND_DASHBOARD_COMMANDS_API.md

### 2.1 Status: ⚠️ MISMATCHES FOUND

### 2.2 Endpoints Comparison

| Frontend Expects | Backend Provides | Status |
|-----------------|------------------|--------|
| POST /v1/device/:imei/command | POST /v1/device/:id/command | ⚠️ **Parameter name mismatch (imei vs id)** |
| GET /v1/device/:imei/commands | GET /v1/device/:imei/commands | ✅ Match |
| DELETE /v1/device/:imei/command/:dispatchId | DELETE /v1/command/:dispatchId | ⚠️ **Path mismatch** |
| GET /v1/device/:imei/commands/pending | GET /v1/device/:id/commands/pending | ⚠️ **Parameter name mismatch** |
| GET /v1/device/:imei/logs | GET /v1/device/:imei/logs | ✅ Match |
| GET /v1/device/:imei/metrics | GET /v1/device/:imei/metrics | ✅ Match |
| GET /v1/device/:imei/telemetry | GET /v1/device/:imei/telemetry | ✅ Match |
| GET /v1/device/:imei/metrics/export | GET /v1/device/:imei/metrics/export | ✅ Match |
| GET /v1/dashboard/stats | GET /v1/dashboard/stats | ✅ Match |

### 2.3 Mismatches Found

| ID | Issue | Severity | Description |
|----|-------|----------|-------------|
| DCL-001 | **Critical: Command Cancel Path Mismatch** | Critical | FE expects `DELETE /v1/device/:imei/command/:dispatchId`, BE provides `DELETE /v1/command/:dispatchId`. The cancel endpoint is at root level, not nested under device. |
| DCL-002 | **Parameter Naming Inconsistency** | High | FE consistently uses `:imei`, BE uses `:id` in existing endpoints. Should standardize on `:imei` for device-specific endpoints. |
| DCL-003 | **Pending Commands Path** | Medium | FE shows `GET /v1/device/:imei/commands/pending`, BE has existing `GET /v1/device/:id/commands/pending`. Inconsistent with new `:imei` convention. |
| DCL-004 | **Missing Cancel Endpoint** | High | BE doc doesn't mention how to cancel a pending command via REST. Only mentions GraphQL `cancelCommand` mutation. Need REST fallback. |

### 2.4 Logic Issues

| ID | Issue | Severity | Description |
|----|-------|----------|-------------|
| DCL-005 | **Command Status Flow** | Medium | BE mentions 4 command statuses: pending, delivered, failed, completed. FE uses same set but doesn't document the transition logic. |
| DCL-006 | **Metrics Time Range** | Low | FE defaults to `-24h` for logs but `-30d` for commands. BE defaults `-24h` for logs. Inconsistent defaults. |

---

## 3. UPDATES_PAGE.md vs SERVER_BACKEND_UPDATES_API.md

### 3.1 Status: ✅ WELL ALIGNED

### 3.2 Endpoints Comparison

| Frontend Expects | Backend Provides | Status |
|-----------------|------------------|--------|
| GET /v1/updates/status | GET /v1/updates/status | ✅ Match |
| GET /v1/updates/versions | GET /v1/updates/versions | ✅ Match |
| GET /v1/updates/changelog | GET /v1/updates/changelog | ✅ Match |
| POST /v1/updates/push | POST /v1/updates/push | ✅ Match |
| GET /v1/updates/history | GET /v1/updates/history | ✅ Match |
| GET /v1/updates/export | GET /v1/updates/export | ✅ Match |
| POST /v1/updates/sync | POST /v1/updates/sync | ✅ Match |
| - | POST /v1/updates/history/:id/cancel | ⚠️ **FE doesn't document cancel** |

### 3.3 Mismatches Found

| ID | Issue | Severity | Description |
|----|-------|----------|-------------|
| UP-001 | **Missing Cancel in FE** | Medium | BE has `POST /v1/updates/history/:id/cancel` for cancelling scheduled updates. FE only mentions cancel in History tab description but doesn't define the API call. |
| UP-002 | **Version Status Values** | Low | BE uses `all`, `latest`, `previous`. FE doc mentions `all`, `latest`, `previous` but in a table it's labeled differently. |
| UP-003 | **Sync Status Values** | Low | BE defines: idle, syncing, synced, error. FE doesn't document these values explicitly. |

### 3.4 Logic Issues
- None identified - update push flow is consistent

---

## 4. SETTINGS_PAGE.md vs SERVER_BACKEND_SETTINGS_API.md

### 4.1 Status: ⚠️ MISMATCHES FOUND

### 4.2 Endpoints Comparison

| Frontend Expects | Backend Provides | Status |
|-----------------|------------------|--------|
| GET/PATCH /v1/settings | - | ⚠️ **FE expects /v1/settings** |
| GET/PATCH /v1/auth/me/settings | GET/PATCH /v1/auth/me/settings | ✅ Match |
| GET/PATCH /v1/auth/me/thresholds | GET/PATCH /v1/auth/me/thresholds | ✅ Match |
| GET/PATCH /v1/auth/me/notifications | GET/PATCH /v1/auth/me/notifications | ✅ Match |
| POST /v1/auth/me/notifications/webhook/test | POST /v1/auth/me/notifications/webhook/test | ✅ Match |
| POST /v1/auth/me/notifications/webhook/rotate | POST /v1/auth/me/notifications/webhook/rotate | ✅ Match |

### 4.3 Mismatches Found

| ID | Issue | Severity | Description |
|----|-------|----------|-------------|
| SP-001 | **Critical: Settings Base Path** | Critical | FE doc (line 246-249) shows REST calls to `join(serverUrl, \`\${BASE}/thresholds\`)` where BASE is `/v1/settings`. But backend uses `/v1/auth/me/settings`. Path mismatch! |
| SP-002 | **Missing Webhook Rotate** | Medium | BE has `POST /v1/auth/me/notifications/webhook/rotate`, FE doesn't document this endpoint. |
| SP-003 | **Reset Endpoint Missing** | Medium | BE has `POST /v1/auth/me/settings/reset` but FE doesn't document it. |

### 4.4 Logic Issues

| ID | Issue | Severity | Description |
|----|-------|----------|-------------|
| SP-004 | **Threshold Validation Logic** | Medium | BE defines validation rules: riskWarn < riskCrit, thermalWarn < thermalCrit, bufferCrit < bufferWarn (inverted). FE doesn't document these validation rules. User could set invalid values. |
| SP-005 | **Webhook Secret Handling** | Low | BE returns `secret` in response when updating webhook, FE doesn't mention this. Should clarify if secret is returned or must be fetched separately. |

---

## 5. DIAGNOSTICS_PAGE.md vs SERVER_BACKEND_DIAGNOSTICS_API.md

### 5.1 Status: ✅ WELL ALIGNED

### 5.2 Endpoints Comparison

| Frontend Expects | Backend Provides | Status |
|-----------------|------------------|--------|
| GET /v1/device/:imei/inspect | GET /v1/device/:imei/inspect | ✅ Match |
| GET /v1/device/:imei/timeline | GET /v1/device/:imei/timeline | ✅ Match |

### 5.3 Mismatches Found

| ID | Issue | Severity | Description |
|----|-------|----------|-------------|
| DI-001 | **Minor: Timeline Event Types** | Low | BE defines 13 event types, FE documents 12 (missing `RECONNECTED`). Should add to FE doc. |
| DI-002 | **Inspector Section Field** | Low | BE includes `buildId` in software section, FE mentions it in types but not in the UI section breakdown. |

### 5.4 Logic Issues
- None identified - diagnostic data flow is consistent

---

## 6. AUTHENTICATION_SYSTEM.md vs AUTHENTICATION_SYSTEM_SERVER.md

### 6.1 Status: ⚠️ MISMATCHES FOUND

### 6.2 Endpoints Comparison

| Frontend Expects | Backend Provides | Status |
|-----------------|------------------|--------|
| POST /v1/auth/login | POST /v1/auth/login | ✅ Match |
| POST /v1/auth/register | POST /v1/auth/register | ✅ Match |
| POST /v1/auth/logout | POST /v1/auth/logout | ✅ Match |
| GET /v1/auth/me | GET /v1/auth/me | ✅ Match |
| POST /v1/auth/mfa/verify | POST /v1/auth/mfa/verify | ⚠️ **BE uses POST /v1/auth/mfa/verify-setup** |
| GET /v1/auth/mfa/status | GET /v1/auth/mfa/status | ✅ Match |
| POST /v1/auth/mfa/enroll | POST /v1/auth/mfa/enroll | ✅ Match |
| POST /v1/auth/forgot-password | POST /v1/auth/forgot-password | ✅ Match |
| POST /v1/auth/reset-password | POST /v1/auth/reset-password | ✅ Match |
| GET /v1/auth/google | GET /v1/auth/google | ✅ Match |
| POST /v1/auth/refresh | POST /v1/auth/refresh | ⚠️ **BE says MISSING** |

### 6.3 Mismatches Found

| ID | Issue | Severity | Description |
|----|-------|----------|-------------|
| AU-001 | **Critical: MFA Verify Endpoint** | High | FE expects `POST /v1/auth/mfa/verify` (for post-login MFA verification). BE has `POST /v1/auth/mfa/verify-setup` (for setup verification). Post-login verification endpoint is MISSING. |
| AU-002 | **Critical: Refresh Token** | High | FE expects token refresh mechanism. BE says `POST /v1/auth/refresh` is MISSING. This is essential for session management. |
| AU-003 | **OAuth Callback Paths** | Medium | FE mentions `GET /v1/auth/google` and callback. BE also shows `/v1/auth/google/callback`. FE doesn't document the callback path explicitly. |

### 6.4 Logic Issues

| ID | Issue | Severity | Description |
|----|-------|----------|-------------|
| AU-004 | **MFA Flow Step Mismatch** | High | FE describes MFA as: User submits credentials → Server returns MFA required → User submits TOTP → Server returns JWT. But BE has separate endpoints for MFA status, enroll, verify-setup, enable. The flow doesn't clearly support the simple post-login MFA verification. |
| AU-005 | **Token Storage** | Medium | FE mentions "HttpOnly cookies or secure storage". BE doesn't specify token storage mechanism. Should align on HttpOnly cookie approach. |

---

## Summary of Critical Issues

### UPDATED: Issues vs Actual Implementation Status

> **IMPORTANT:** After code review, many "missing" endpoints are actually IMPLEMENTED in the Go code. The issues below are primarily DOCUMENTATION MISMATCHES between FE and BE specs, not implementation gaps.

### Critical Issues (Doc vs Doc, NOT Implementation)

| ID | Doc Pair | Issue | Actual Status |
|----|----------|-------|---------------|
| DCL-001 | Dashboard/Commands | Command cancel path mismatch | ✅ IMPLEMENTED (code exists) - doc mismatch |
| SP-001 | Settings | Settings base path mismatch | ✅ IMPLEMENTED (code exists) - FE doc wrong path |
| AU-001 | Authentication | MFA verify endpoint | ✅ IMPLEMENTED (line 183 auth_routes.go) |
| AU-002 | Authentication | Refresh token endpoint | ✅ IMPLEMENTED (RefreshHandler exists) |

### High Priority (Documentation Fixes)

| ID | Doc Pair | Issue | Action Needed |
|----|----------|-------|---------------|
| DCL-002 | Dashboard/Commands | Parameter naming inconsistency (id vs imei) | Standardize docs to use `:imei` |
| DCL-004 | Dashboard/Commands | REST cancel endpoint not documented | Add to BE spec |
| SP-002 | Settings | Webhook rotate endpoint missing from FE doc | Add to FE doc |
| SP-003 | Settings | Settings reset endpoint missing from FE doc | Add to FE doc |
| AU-004 | Authentication | MFA flow step mismatch | Clarify setup vs verify flow in docs |

### Medium Priority (Documentation Fixes)

| ID | Doc Pair | Issue | Action Needed |
|----|----------|-------|---------------|
| DR-001 | Device Registration | Inbox entry ID field naming | Standardize to lowercase `id` |
| UP-001 | Updates | Cancel update endpoint missing from FE doc | Add to FE doc |
| SP-004 | Settings | Threshold validation rules not documented | Add validation rules to FE doc |

### Low Priority (Minor)

| ID | Doc Pair | Issue |
|----|----------|-------|
| DR-002 | Device Registration | GraphQL vs REST naming |
| UP-002 | Updates | Version status values |
| UP-003 | Updates | Sync status values |
| DI-001 | Diagnostics | Missing RECONNECTED event type in FE |
| DI-002 | Diagnostics | buildId field in UI |
| AU-003 | Authentication | OAuth callback path |

---

## ACTUAL IMPLEMENTATION GAPS (Real Issues Found)

### 1. Metrics Export Missing
| Issue | Status | Location |
|-------|--------|----------|
| GET /v1/device/:imei/metrics/export | ❌ MISSING | Specified in BE doc but NOT in code |

**File:** `internal/api/handlers/device/device_metrics_handler.go`
**Fix:** Add ExportMetrics handler method

### 2. Inbox DELETE vs PATCH
| Issue | Status | Location |
|-------|--------|----------|
| DELETE /v1/device/inbox/:imei | ⚠️ Uses PATCH | Code uses UpdateInboxEntry (PATCH) instead of DELETE |

**File:** `internal/api/handlers/inbox/inbox_routes.go` line 12
**Fix:** Change to proper DELETE or clarify the route

---

## Fix Plan

### Phase 1: Documentation Fixes Only (No Code Changes Needed)

1. **Settings Path Fix (SP-001)**
   - Update FE doc `SETTINGS_PAGE.md` to use `/v1/auth/me/*` paths
   - Implementation already correct

2. **Dashboard Command Cancel (DCL-001)**
   - Update FE doc to reflect correct cancel endpoint path `/v1/command/:dispatchId`
   - Implementation already correct

3. **Update Auth Flow Documentation (AU-001, AU-002, AU-004)**
   - Both endpoints exist in code
   - Update FE doc to document MFA verify flow and refresh endpoint

### Phase 2: Documentation Additions

4. **Add Missing Endpoints to FE Docs**
   - Webhook rotate → Add to SETTINGS_PAGE.md
   - Settings reset → Add to SETTINGS_PAGE.md  
   - Updates cancel → Add to UPDATES_PAGE.md

5. **Standardize Parameter Naming**
   - Change all `:id` to `:imei` in FE and BE docs for device endpoints

### Phase 3: Code Implementation Gaps

6. **Add Metrics Export (REAL GAP)**
   - File: `internal/api/handlers/device/device_metrics_handler.go`
   - Add `ExportMetrics` method for `GET /v1/device/:imei/metrics/export`

7. **Clarify Inbox DELETE**
   - Option A: Implement actual DELETE functionality
   - Option B: Update route to use PATCH and update docs

---

## Recommendations

1. **Establish Canonical API Path Convention**
   - Device-specific: `/v1/device/:imei/*`
   - User-specific: `/v1/auth/me/*`
   - Command-specific: `/v1/command/*`

2. **Create Shared API Contract Document**
   - Single source of truth for all endpoints
   - Both FE and BE teams reference same document
   - Avoids drift between documents

3. **Add Field-Level Consistency Check**
   - Document should list all response fields
   - Include data types and constraints
   - Validate both docs match before implementation

4. **Review Authentication Flow**
   - Current MFA flow is unclear
   - Need to clarify post-login MFA vs setup MFA
   - Token refresh mechanism needs to be added

---

*Report Generated: 2026-07-05*
*Next Action: Review and approve fix plan with team*
