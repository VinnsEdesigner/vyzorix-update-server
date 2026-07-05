# Spec Mismatch Report: Frontend vs Backend Documentation Analysis

> **Date:** 2026-07-05
> **Project:** Vyzorix Update Server
> **Purpose:** Identify mismatches, wrong logics, and inconsistencies between frontend and backend spec documents

---

## Executive Summary

This report analyzes 6 pairs of Frontend (FE) and Backend (BE) specification documents to identify:
- Endpoint mismatches
- Missing/inconsistent field definitions
- Data type discrepancies
- Logic inconsistencies
- Security requirement gaps

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

### Critical (Must Fix Before Implementation)

| ID | Doc Pair | Issue |
|----|----------|-------|
| DCL-001 | Dashboard/Commands | Command cancel path mismatch |
| SP-001 | Settings | Settings base path mismatch |
| AU-001 | Authentication | MFA verify endpoint missing |
| AU-002 | Authentication | Refresh token endpoint missing |

### High Priority

| ID | Doc Pair | Issue |
|----|----------|-------|
| DCL-002 | Dashboard/Commands | Parameter naming inconsistency (id vs imei) |
| DCL-004 | Dashboard/Commands | Missing REST cancel command endpoint |
| SP-002 | Settings | Webhook rotate endpoint missing from FE |
| SP-003 | Settings | Settings reset endpoint missing from FE |
| AU-004 | Authentication | MFA flow step mismatch |

### Medium Priority

| ID | Doc Pair | Issue |
|----|----------|-------|
| DR-001 | Device Registration | Inbox entry ID field naming |
| UP-001 | Updates | Cancel update endpoint missing from FE |
| SP-004 | Settings | Threshold validation rules not documented |

### Low Priority

| ID | Doc Pair | Issue |
|----|----------|-------|
| DR-002 | Device Registration | GraphQL vs REST naming |
| UP-002 | Updates | Version status values |
| UP-003 | Updates | Sync status values |
| DI-001 | Diagnostics | Missing RECONNECTED event type |
| DI-002 | Diagnostics | buildId field in UI |
| AU-003 | Authentication | OAuth callback path |

---

## Fix Plan

### Phase 1: Critical Fixes (Do First)

1. **Settings Path Fix (SP-001)**
   - Decision: Use backend path `/v1/auth/me/*` as canonical
   - Update FE doc to change BASE from `/v1/settings` to `/v1/auth/me`

2. **Dashboard Command Cancel (DCL-001)**
   - Decision: Use backend path `/v1/command/:dispatchId` for cancel
   - Update FE doc to reflect correct cancel endpoint path

3. **Authentication MFA (AU-001, AU-002)**
   - Add `POST /v1/auth/mfa/verify` endpoint to backend
   - Add `POST /v1/auth/refresh` endpoint to backend

### Phase 2: High Priority Fixes

4. **Parameter Naming (DCL-002)**
   - Standardize on `:imei` for all device-specific endpoints
   - Update backend to use `:imei` instead of `:id` for consistency

5. **Settings Endpoints (SP-002, SP-003)**
   - Add webhook rotate and settings reset to FE documentation

6. **MFA Flow (AU-004)**
   - Clarify MFA flow: setup (verify-setup) vs login (verify)
   - Document when each endpoint is used

### Phase 3: Documentation Cleanup

7. **Update FE Docs with Missing Endpoints**
   - Add cancel update endpoint for Updates
   - Add webhook rotate endpoint for Settings
   - Add RECONNECTED event type for Diagnostics

8. **Add Validation Rules**
   - Document threshold validation rules in Settings FE doc
   - Document sync status values in Updates FE doc

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
