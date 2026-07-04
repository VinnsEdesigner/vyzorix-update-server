# GraphQL API Bug Report

> **Document Version:** 1.0  
> **Status:** RESOLVED  
> **Created:** 2026-07-02  
> **Last Updated:** 2026-07-04  
> **Project:** Vyzorix Update Server

---

## Executive Summary

This document catalogs all GraphQL and REST API bugs identified in the Vyzorix-update-server codebase, their severity, file locations, and resolution status.

---

## CRITICAL Bugs (Resolved)

### CRITICAL-API-1: MFA Verify Missing Rate Limiting

**Severity:** CRITICAL  
**Status:** RESOLVED  
**File:** `apps/api/internal/api/handlers/auth/auth_routes.go`

**Problem:**  
The `/mfa/verify` endpoint had no rate limiting, allowing brute force attacks on TOTP codes.

**Resolution:**  
Rate limiting middleware applied to MFA verify endpoint. See commit `01a0836f`.

---

### CRITICAL-API-2: Race Condition in MFA Verify Flow

**Severity:** CRITICAL  
**Status:** RESOLVED  
**File:** `apps/api/internal/api/handlers/auth/auth_mfa.go`

**Problem:**  
Between VerifyMFACode and CreateCookie, operator state could change (deleted, MFA disabled, etc.)

**Resolution:**  
Re-validation of operator state before session creation implemented. See commit `01a0836f`.

---

### CRITICAL-API-3: No JWT Tokens Returned from MFA Verify

**Severity:** CRITICAL  
**Status:** RESOLVED  
**File:** `apps/api/internal/api/handlers/auth/auth_mfa.go`

**Problem:**  
After MFA verification, client received session_id but not access_token, refresh_token, or expires_at.

**Resolution:**  
Full token bundle returned from MFA verify. See commit `01a0836f`.

---

### CRITICAL-API-4: Cookie Error Silently Ignored

**Severity:** CRITICAL  
**Status:** RESOLVED  
**File:** `apps/api/internal/api/handlers/auth/auth_login.go`

**Problem:**  
If CreateCookie failed, handler returned 200 with success=true but no session cookie.

**Resolution:**  
Proper error handling for cookie creation failures. See commit `01a0836f`.

---

## HIGH Priority Bugs (Resolved)

### HIGH-API-1: MFA Token Not Bound Between Login and MFA

**Severity:** HIGH  
**Status:** RESOLVED  
**File:** `apps/api/internal/api/handlers/auth/auth_mfa.go`

**Resolution:**  
MFA token binding implemented. See commit `02a3b83b`.

---

### HIGH-API-2: Token Rotation Not Properly Revoking Old Tokens

**Severity:** HIGH  
**Status:** RESOLVED  
**File:** `apps/api/internal/application/auth/auth_service.go`

**Resolution:**  
Proper token revocation on rotation. See commit `41077c4c`.

---

### HIGH-API-3: Logout Not Revoking Refresh Tokens

**Severity:** HIGH  
**Status:** RESOLVED  
**File:** `apps/api/internal/application/auth/auth_logout.go`

**Resolution:**  
Logout now revokes all refresh tokens. See commit `51b2d7e8`.

---

### HIGH-API-4: Session Enumeration Possible

**Severity:** HIGH  
**Status:** RESOLVED  
**File:** `apps/api/internal/api/handlers/auth/auth_sessions.go`

**Resolution:**  
Session listing with proper authorization. See commit `41077c4c`.

---

## MEDIUM Priority Bugs (Resolved)

### MEDIUM-API-1: operator_id Exposed in MFA Response

**Severity:** MEDIUM  
**Status:** RESOLVED  
**File:** `apps/api/internal/api/handlers/auth/auth_mfa.go`

**Resolution:**  
operator_id removed from MFA response. See commit `14e8c2e3`.

---

### MEDIUM-API-2: Missing MFA Audit Logging

**Severity:** MEDIUM  
**Status:** RESOLVED  
**File:** `apps/api/internal/application/auth/auth_mfa.go`

**Resolution:**  
Comprehensive MFA audit logging added. See commit `34d19745`.

---

### MEDIUM-API-3: Lockout Not Applied to MFA Verify

**Severity:** MEDIUM  
**Status:** RESOLVED  
**File:** `apps/api/internal/api/middleware/lockout.go`

**Resolution:**  
Lockout middleware applied to MFA verify endpoint. See commit `60a383ab`.

---

## Bug Report Summary

| Bug ID | Severity | Description | Status |
|--------|----------|-------------|--------|
| CRITICAL-API-1 | CRITICAL | MFA rate limiting missing | RESOLVED |
| CRITICAL-API-2 | CRITICAL | MFA race condition | RESOLVED |
| CRITICAL-API-3 | CRITICAL | Missing JWT tokens | RESOLVED |
| CRITICAL-API-4 | CRITICAL | Cookie error ignored | RESOLVED |
| HIGH-API-1 | HIGH | MFA token binding | RESOLVED |
| HIGH-API-2 | HIGH | Token rotation | RESOLVED |
| HIGH-API-3 | HIGH | Logout refresh revocation | RESOLVED |
| HIGH-API-4 | HIGH | Session enumeration | RESOLVED |
| MEDIUM-API-1 | MEDIUM | operator_id exposure | RESOLVED |
| MEDIUM-API-2 | MEDIUM | Missing MFA audit | RESOLVED |
| MEDIUM-API-3 | MEDIUM | Lockout on MFA | RESOLVED |

---

## Resolution Verification

All critical and high-priority API bugs have been resolved as of commit `f66dbdd3`. The codebase now passes all security checks and is ready for production deployment.

---

*Document Version: 1.0*  
*Status: Complete - All bugs resolved*  
*Vyzorix-update-server*
