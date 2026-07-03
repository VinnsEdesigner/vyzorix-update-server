# Vyzorix Update Server - Bug Analysis & Enterprise Requirements

> **Document Version:** 1.0  
> **Status:** Implementation Required  
> **Created:** 2026-07-02  
> **Priority:** CRITICAL - Security & Enterprise Readiness  

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [CRITICAL Security Bugs](#critical-security-bugs)
3. [HIGH Priority Issues](#high-priority-issues)
4. [MEDIUM Priority Issues](#medium-priority-issues)
5. [Missing Enterprise Features](#missing-enterprise-features)
6. [Implementation Checklist](#implementation-checklist)

---

## Executive Summary

This document outlines all identified bugs and missing enterprise features in the Vyzorix Update Server codebase. Issues are categorized by severity and include specific file locations, code analysis, and remediation requirements.

**Total Issues Identified:**
- CRITICAL: 8 issues
- HIGH: 10 issues
- MEDIUM: 12 issues
- Missing Enterprise Features: 15+ items

---

## CRITICAL Security Bugs

### CRITICAL-1: No MFA Rate Limiting on /mfa/verify

**Severity:** CRITICAL  
**Status:** MUST FIX  
**File:** `apps/api/internal/api/handlers/auth/auth_routes.go`

**Problem:**
The `/mfa/verify` endpoint is completely unprotected against brute force attacks. An attacker can try all 1,000,000 possible 6-digit TOTP codes against any operator_id.

**Current Code:**
```go
// auth_routes.go - NO RATE LIMITING APPLIED
mfa := rg.Group("/mfa")
mfa.Use(middleware.NoCache())
{
    mfa.POST("/verify", h.MFA.VerifyMFA) // No cookie auth, NO RATE LIMIT
}
```

**Attack Scenario:**
1. Attacker obtains operator_id (often predictable)
2. Attacker brute forces 6-digit TOTP (1M combinations)
3. No rate limiting, no account lockout, no attempt counter
4. Account compromised

**Required Fix:**
```go
mfa := rg.Group("/mfa")
mfa.Use(middleware.NoCache())
mfa.Use(rateLimiter.MFAVerifyLimit()) // ADDED
{
    mfa.POST("/verify", h.MFA.VerifyMFA)
}
```

**Rate Limit Recommendation:**
- 5 attempts per minute per operator_id
- Progressive lockout after failures
- IP-based blocking for distributed attacks

---

### CRITICAL-2: Race Condition in MFA Verify

**Severity:** CRITICAL  
**Status:** MUST FIX  
**File:** `apps/api/internal/api/handlers/auth/auth_mfa.go`

**Problem:**
Between `VerifyMFACode` and `CreateCookie`, the operator could be deleted, MFA disabled, or role changed. No atomic operation.

**Current Code:**
```go
// auth_mfa.go - VerifyMFA handler
sess, err := h.authService.VerifyMFACode(c.Request.Context(), req.OperatorID, req.Code)
if err != nil {
    // error handling
}
// ... TIME GAP - operator_id could be invalidated ...

// Create session cookie
if h.authService.GetSessionManager() != nil {
    cookie, cookieErr := h.authService.GetSessionManager().CreateCookie(req.OperatorID)
    // ...
}
```

**Race Condition Window:**
Between VerifyMFACode and CreateCookie:
1. Operator could be DELETED
2. MFA could be DISABLED
3. Role could be CHANGED
4. Account could be LOCKED

**Required Fix:**
Atomic operation combining verification and session creation, or re-validate operator state before session creation.

---

### CRITICAL-3: Cookie Error Silently Ignored

**Severity:** CRITICAL  
**Status:** MUST FIX  
**File:** `apps/api/internal/api/handlers/auth/auth_mfa.go`

**Problem:**
If `CreateCookie` fails, the handler returns 200 with `success=true` but no session cookie. Client thinks MFA worked but has no session.

**Current Code:**
```go
// auth_mfa.go
cookie, cookieErr := h.authService.GetSessionManager().CreateCookie(req.OperatorID)
if cookieErr == nil { // Only sets cookie if successful
    h.presenter.SetSessionCookie(c, cookie)
}
return h.presenter.OK(c, gin.H{
    "success": true,  // RETURNS SUCCESS EVEN IF COOKIE FAILED
    "session_id": sess.ID,
    "operator": {...},
})
```

**Impact:**
- User sees "Login Successful"
- No session cookie set
- Every subsequent request fails
- Confusing UX / security issue

**Required Fix:**
```go
cookie, cookieErr := h.authService.GetSessionManager().CreateCookie(req.OperatorID)
if cookieErr != nil {
    return h.presenter.Error(c, http.StatusInternalServerError, "session_error", "Failed to create session")
}
h.presenter.SetSessionCookie(c, cookie)
```

---

### CRITICAL-4: No JWT Returned from MFA Verify

**Severity:** CRITICAL  
**Status:** MUST FIX  
**File:** `apps/api/internal/api/handlers/auth/auth_mfa.go`

**Problem:**
After MFA verification, the client needs tokens but only receives `session_id`.

**Current Response:**
```go
h.presenter.OK(c, gin.H{
    "success": true,
    "session_id": sess.ID,
    "operator": {...},
    // MISSING: access_token
    // MISSING: refresh_token
    // MISSING: expires_at
})
```

**Impact:**
- Client cannot maintain session after MFA verification
- Client cannot refresh expired tokens
- Session expires immediately after MFA

**Required Fix:**
```go
h.presenter.OK(c, gin.H{
    "success": true,
    "session_id": sess.ID,
    "access_token": accessToken,      // ADD
    "refresh_token": refreshToken,    // ADD
    "expires_at": expiresAt.Unix(),   // ADD
    "operator": {...},
})
```

---

### CRITICAL-5: MFA Token Not Cryptographically Bound

**Severity:** CRITICAL  
**Status:** MUST FIX  
**File:** `apps/api/internal/domain/auth/mfa_token_entity.go`

**Problem:**
MFA tokens are not cryptographically bound to the operator. The binding is just a database field, not a cryptographic signature.

**Current Implementation:**
```go
// mfa_token_entity.go
type MFAToken struct {
    Token     string
    OperatorID string  // Just a database field
    ExpiresAt time.Time
    // NO cryptographic binding
}
```

**Required Fix:**
```go
type MFAToken struct {
    Token     string
    OperatorID string
    ExpiresAt time.Time
    // ADD: Cryptographic signature
    Signature []byte
}

// Verify checks that the token signature is valid
func (t *MFAToken) Verify(operatorSecret []byte) bool {
    mac := hmac.New(sha256.New, operatorSecret)
    mac.Write([]byte(t.Token))
    mac.Write([]byte(t.OperatorID))
    return hmac.Equal(t.Signature, mac.Sum(nil))
}
```

---

### CRITICAL-6: OAuth State Not Persisted (Race Condition)

**Severity:** CRITICAL  
**Status:** MUST FIX  
**File:** `apps/api/internal/api/handlers/auth/auth_oauth.go`

**Problem:**
State is generated but never persisted to DB - only encoded in URL. Vulnerable to race conditions and state substitution attacks.

**Current Code (lines 62-78):**
```go
// auth_oauth.go
state := generateState()  // Generated
// Stored in cookie ONLY - not in DB
setStateCookie(state)
// Sent in URL - attacker can modify
redirectURL := fmt.Sprintf("%s?state=%s", authURL, state)
```

**Attack Vector:**
1. Attacker initiates OAuth flow, gets state=A
2. Attacker sends state=B to victim
3. Victim authenticates with state=B
4. Attacker uses state=A to complete flow

**Required Fix:**
```go
// Store state in database with expiry
state := &OAuthState{
    State:     generateState(),
    OperatorID: operatorID,  // Bind to operator
    RedirectURL: redirectURL,
    ExpiresAt: time.Now().Add(10 * time.Minute),
    CreatedAt: time.Now(),
}
if err := s.oauthRepo.CreateState(ctx, state); err != nil {
    return err
}
```

---

### CRITICAL-7: No Password Breach Checking

**Severity:** CRITICAL  
**Status:** MUST FIX  
**File:** `apps/api/internal/api/handlers/auth/auth_register.go`

**Problem:**
Passwords aren't checked against HaveIBeenPwned - basic enterprise requirement.

**Current Implementation:**
```go
// auth_register.go
func (h *RegisterHandler) Handle(c *gin.Context) {
    // Password validated for format only
    // NO breach checking
}
```

**Required Fix:**
```go
// Check password against HIBP
func (s *AuthService) CheckPasswordBreach(password string) (bool, error) {
    // k-Anonymity model: hash password, send first 5 chars to HIBP API
    hash := sha1.Sum([]byte(password))
    hashStr := strings.ToUpper(hex.EncodeToString(hash[:]))
    
    resp, err := http.Get("https://api.pwnedpasswords.com/range/" + hashStr[:5])
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()
    
    // Check if hash suffix is in response
    body, _ := io.ReadAll(resp.Body)
    suffix := strings.Split(hashStr[5:], "\r\n")[0]
    if strings.Contains(string(body), suffix) {
        return true, nil // Password found in breach
    }
    return false, nil
}
```

---

### CRITICAL-8: No JWT Secret Validation on Startup

**Severity:** CRITICAL  
**Status:** MUST FIX  
**File:** `apps/api/internal/application/auth/auth_service.go`

**Problem:**
If JWT manager is nil, tokens are still issued with empty secret.

**Current Code:**
```go
// auth_service.go
if h.jwtManager == nil {
    // Should return error, but continues
    return nil, ErrJWTMisconfigured
}
// ... later ...
token, _ := h.jwtManager.Generate(claims) // Uses empty secret!
```

**Required Fix:**
```go
// In AuthService initialization
if cfg.JWTSecret == "" || len(cfg.JWTSecret) < 32 {
    return nil, errors.New("JWT secret must be at least 32 characters")
}

h.jwtManager = NewJWTManager(cfg.JWTSecret)
```

---

## HIGH Priority Issues

### HIGH-1: No MFA Token Binding Between Login and MFA

**Severity:** HIGH  
**Status:** MUST FIX  
**File:** `apps/api/internal/api/handlers/auth/auth_mfa.go`

**Problem:**
No token binding between login step and MFA step. We fixed the code but the binding is weak (just operator_id in MFA token).

**Current Flow:**
1. Login returns `operator_id`
2. MFA verify takes `operator_id + code`
3. No single auth token spanning entire flow

**Required Fix:**
Bind MFA token to login session:
```go
// In login response
"mfa_token": bindMFAToSession(operatorID, sessionID)

// In MFA verify
func VerifyMFA(req *MFAVerifyRequest) error {
    // Verify mfa_token was issued for this session
    sessionID := extractSessionID(req.MFAToken)
    if sessionID != currentSession {
        return ErrTokenMismatch
    }
}
```

---

### HIGH-2: Token Rotation Creates Sessions Without Revoking Old

**Severity:** HIGH  
**Status:** MUST FIX  
**File:** `apps/api/internal/application/auth/refresh_token.go`

**Problem:**
Refresh token rotation creates new session but old session NOT revoked (concurrent sessions). Old refresh token NOT immediately revoked.

**Impact:**
- Token theft = persistent access
- Legitimate user doesn't know someone has their token
- Creates session accumulation per operator

**Required Fix:**
```go
func (s *RefreshTokenService) RotateRefreshToken(ctx context.Context, operatorID, oldToken string) (*TokenPair, error) {
    // 1. Validate old token is not revoked
    if s.isRevoked(oldToken) {
        return nil, ErrTokenReused
    }
    
    // 2. Revoke old token (keep in DB for 90 days)
    if err := s.revokeToken(oldToken, 90*24*time.Hour); err != nil {
        return nil, err
    }
    
    // 3. Create new tokens
    return s.createTokenPair(operatorID)
}
```

---

### HIGH-3: Logout Doesn't Revoke Refresh Tokens

**Severity:** HIGH  
**Status:** MUST FIX  
**File:** `apps/api/internal/api/handlers/auth/auth_logout.go`

**Problem:**
Logout endpoint doesn't revoke refresh tokens - user can still use old refresh token.

**Required Fix:**
```go
func (h *LogoutHandler) Handle(c *gin.Context) {
    // Revoke all refresh tokens for operator
    if err := h.authService.RevokeAllRefreshTokens(ctx, operatorID); err != nil {
        return err
    }
    // Revoke session
    if err := h.sessionManager.RevokeSession(ctx, sessionID); err != nil {
        return err
    }
}
```

---

### HIGH-4: No Session Enumeration/Management for Users

**Severity:** HIGH  
**Status:** MISSING FEATURE  
**Files:** `apps/api/internal/api/handlers/auth/sessions.go`

**Missing Endpoints:**
```
GET  /v1/auth/sessions              - List all sessions
GET  /v1/auth/sessions/concurrent   - Check concurrent logins
DELETE /v1/auth/sessions/:id        - Revoke specific session
DELETE /v1/auth/sessions            - Revoke all except current
POST /v1/auth/sessions/revoke-all   - Logout all devices
```

**Required Implementation:**
```go
// ListSessions returns all active sessions for operator
func (h *SessionsHandler) ListSessions(c *gin.Context) {
    sessions, err := h.sessionManager.ListSessions(ctx, operatorID)
    if err != nil {
        return err
    }
    return c.JSON(200, sessions)
}

// RevokeSession revokes a specific session
func (h *SessionsHandler) RevokeSession(c *gin.Context) {
    sessionID := c.Param("id")
    if sessionID == currentSessionID {
        return c.JSON(400, gin.H{"error": "Cannot revoke current session"})
    }
    return h.sessionManager.RevokeSession(ctx, sessionID)
}
```

---

### HIGH-5: Refresh Token Reuse Detection Incomplete

**Severity:** HIGH  
**Status:** MUST FIX  
**File:** `apps/api/internal/infrastructure/security/refresh_token.go`

**Problem:**
Refresh token reuse detection only works for same operator. If token is stolen and used from different operator account, it won't detect it.

**Required Fix:**
```go
func (s *RefreshTokenService) ValidateToken(ctx context.Context, token string) (*Claims, error) {
    claims, err := s.parseToken(token)
    if err != nil {
        return nil, err
    }
    
    // Check if token is in revocation list (from ANY operator)
    if s.isRevokedGlobally(token) {
        // Token was used after being revoked - POSSIBLE THEFT
        // Revoke ALL active tokens for this operator
        s.RevokeAllTokensForOperator(ctx, claims.OperatorID)
        return nil, ErrTokenCompromised
    }
    
    return claims, nil
}
```

---

### HIGH-6: Revoked Tokens Must Persist 90 Days

**Severity:** HIGH  
**Status:** MUST FIX  
**Database:** `apps/api/internal/infrastructure/storage/session_storage.go`

**Problem:**
When tokens are revoked, they must be kept in DB for 90 days to prevent reuse.

**Required Implementation:**
```sql
CREATE TABLE revoked_tokens (
    token_hash TEXT PRIMARY KEY,
    operator_id TEXT NOT NULL,
    revoked_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,  -- 90 days after revoked_at
    reason TEXT
);

-- Cleanup job removes tokens older than 90 days
DELETE FROM revoked_tokens WHERE expires_at < NOW();
```

---

### HIGH-7: MFA Not Enforced (Opt-in Only)

**Severity:** HIGH  
**Status:** SPEC VIOLATION  
**File:** `apps/api/internal/application/auth/auth_service.go`

**Problem:**
MFA is opt-in only, but enterprise requirements mandate MFA for all operators.

**Required Fix:**
```go
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
    // After password verification, check if MFA is required
    if operator.MFAEnabled && !operator.MFAVerified {
        // Force MFA enrollment before login completes
        mfaToken, _ := s.CreateMFAToken(ctx, operator.ID)
        return &LoginResponse{
            MFARequired: true,
            MFAToken:    mfaToken,
        }, nil
    }
}
```

---

### HIGH-8: Double Operator Fetch Inefficiency

**Severity:** MEDIUM  
**Status:** MUST FIX  
**File:** `apps/api/internal/application/settings/settings_service.go`

**Problem:**
Settings service duplicates auth service methods - fetches operator separately instead of using AuthService.

**Current Code:**
```go
// SettingsService has its own operator fetch logic
op, err := s.operatorRepo.FindByID(ctx, operatorID)
```

**Required Fix:**
```go
// Use AuthService method for consistency
op, err := s.authService.GetOperator(ctx, operatorID)
```

---

### HIGH-9: Concurrent Login Detection Missing

**Severity:** HIGH  
**Status:** MISSING FEATURE  
**File:** `apps/api/internal/api/handlers/auth/concurrent_login.go`

**Problem:**
No notification when same account logs in from different IP/device.

**Missing Features:**
1. No concurrent session detection
2. No notification to user when concurrent login detected
3. No "kick existing session" option

**Required Implementation:**
```go
// Detect concurrent login
func (s *SessionService) DetectConcurrentLogin(ctx context.Context, operatorID, newSessionID, newIP string) error {
    existingSessions := s.ListActiveSessions(ctx, operatorID)
    
    if len(existingSessions) > 0 {
        // Notify existing sessions
        for _, sess := range existingSessions {
            s.NotifySessionReplaced(ctx, sess.ID, newSessionID)
        }
        
        // Optional: revoke existing sessions
        if s.config.RevokeOnConcurrentLogin {
            for _, sess := range existingSessions {
                s.RevokeSession(ctx, sess.ID)
            }
        }
    }
    
    return nil
}
```

---

### HIGH-10: No "Remember Me" / Persistent Sessions

**Severity:** HIGH  
**Status:** MISSING FEATURE  
**File:** `apps/api/internal/api/handlers/auth/auth_login.go`

**Problem:**
All sessions expire at sessionTTL (default ~24h). No extended session option for trusted devices.

**Missing Features:**
1. "Remember Me" checkbox on login
2. Extended session duration (30 days vs 24 hours)
3. Persistent session storage

**Required Implementation:**
```go
// Login with remember me
func (h *LoginHandler) Handle(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        return err
    }
    
    // If remember_me is true, use extended TTL
    ttl := time.Hour * 24  // 24 hours
    if req.RememberMe {
        ttl = time.Hour * 24 * 30  // 30 days
    }
    
    session, err := h.authService.LoginWithTTL(ctx, req, ttl)
}
```

---

## MEDIUM Priority Issues

### MEDIUM-1: ID Exposure - operator_id in MFA Response

**Severity:** MEDIUM  
**Status:** MUST FIX  
**File:** `apps/api/internal/api/handlers/auth/auth_mfa.go`

**Problem:**
operator_id exposed in MFA response - information disclosure.

**Required Fix:**
Remove operator_id from response or use internal ID mapping.

---

### MEDIUM-2: No Audit Logging on MFA Verify

**Severity:** MEDIUM  
**Status:** MISSING  
**File:** `apps/api/internal/api/handlers/auth/auth_mfa.go`

**Problem:**
MFA verify failures not logged - security blind spot.

**Required Fix:**
```go
func (h *MFAHandler) VerifyMFA(c *gin.Context) {
    // Log attempt
    h.auditLogger.Log(&AuditEvent{
        Action:   "mfa_verify_attempt",
        OperatorID: req.OperatorID,
        IP:       c.ClientIP(),
        Success:  false,
    })
    
    if err := h.authService.VerifyMFACode(...); err != nil {
        h.auditLogger.Log(&AuditEvent{
            Action:   "mfa_verify_failure",
            OperatorID: req.OperatorID,
            Reason:   err.Error(),
        })
    }
}
```

---

### MEDIUM-3: Lockout MFA Verify Has No Lockout Protection

**Severity:** MEDIUM  
**Status:** MISSING  
**File:** `apps/api/internal/api/handlers/auth/auth_lockout.go`

**Problem:**
MFA verify has no lockout protection - brute force possible even with rate limiting.

**Required Fix:**
Apply account lockout to MFA verify endpoint.

---

### MEDIUM-4: Cookie Error Silently Ignored

**Severity:** MEDIUM  
**Status:** MUST FIX  
**File:** `apps/api/internal/api/handlers/auth/auth_mfa.go`

**Already covered in CRITICAL-3.**

---

### MEDIUM-5: No JWT Returned from VerifyMFACode

**Severity:** MEDIUM  
**Status:** MUST FIX  
**File:** `apps/api/internal/api/handlers/auth/auth_mfa.go`

**Already covered in CRITICAL-4.**

---

### MEDIUM-6: Audit MFA Verify Failures Not Logged

**Severity:** MEDIUM  
**Status:** MISSING  
**File:** `apps/api/internal/api/handlers/auth/auth_mfa.go`

**Already covered in MEDIUM-2.**

---

### MEDIUM-7: Double Operator Fetch

**Severity:** MEDIUM  
**Status:** MUST FIX  
**File:** `apps/api/internal/application/settings/settings_service.go`

**Already covered in HIGH-8.**

---

### MEDIUM-8: Session Fingerprinting Missing

**Severity:** MEDIUM  
**Status:** MISSING  
**File:** `apps/api/internal/infrastructure/security/session.go`

**Problem:**
Device fingerprint stored but never validated on requests.

**Required Implementation:**
```go
type SessionFingerprint struct {
    UserAgent string
    IPAddress string
    DeviceID  string
}

func (s *SessionManager) ValidateFingerprint(ctx context.Context, sessionID string, fp *SessionFingerprint) bool {
    stored := s.GetFingerprint(ctx, sessionID)
    return stored.UserAgent == fp.UserAgent && 
           stored.IPAddress == fp.IPAddress
}
```

---

### MEDIUM-9: No IP Anomaly Detection

**Severity:** MEDIUM  
**Status:** MISSING  
**File:** `apps/api/internal/api/middleware/ip_anomaly.go`

**Problem:**
Concurrent login detection exists but doesn't alert/block based on IP change.

**Required Implementation:**
```go
func (s *IPIntelligence) DetectAnomaly(ctx context.Context, operatorID, newIP string) (*AnomalyAlert, error) {
    lastLogin := s.GetLastLogin(ctx, operatorID)
    
    if lastLogin.IP != newIP {
        // New IP detected
        if s.IsSuspiciousIP(newIP) {
            return &AnomalyAlert{
                Type:    IP_CHANGE_SUSPICIOUS,
                OldIP:   lastLogin.IP,
                NewIP:   newIP,
                Action:  ALERT_AND_BLOCK,
            }
        }
        
        return &AnomalyAlert{
            Type:   IP_CHANGE_DETECTED,
            OldIP:  lastLogin.IP,
            NewIP:  newIP,
            Action: ALERT_ONLY,
        }
    }
    return nil, nil
}
```

---

### MEDIUM-10: Settings Service Duplicates Auth Service Methods

**Severity:** MEDIUM  
**Status:** MUST FIX  
**File:** `apps/api/internal/application/settings/settings_service.go`

**Already covered in HIGH-8.**

---

## Missing Enterprise Features

### Missing-1: No RBAC/Permissions System

**Severity:** HIGH  
**Status:** MISSING  
**Impact:**
Just roles (admin/user), no resource-level permissions.

**Required Implementation:**
```go
type Permission string

const (
    PermissionRegisterDevices   Permission = "register_devices"
    PermissionDeregisterDevices Permission = "deregister_devices"
    PermissionSendCommands     Permission = "send_commands"
    PermissionViewTelemetry   Permission = "view_telemetry"
    PermissionPushUpdates     Permission = "push_updates"
    PermissionManageOperators  Permission = "manage_operators"
)

type Role struct {
    ID          string
    Name        string
    Permissions []Permission
}

func (s *AuthService) HasPermission(ctx context.Context, operatorID string, perm Permission) bool {
    role := s.GetOperatorRole(ctx, operatorID)
    for _, p := range role.Permissions {
        if p == perm {
            return true
        }
    }
    return false
}
```

---

### Missing-2: No Password Policy Enforcement Per Tenant

**Severity:** HIGH  
**Status:** MISSING  
**Impact:**
Enterprise needs configurable complexity rules.

**Required Implementation:**
```go
type PasswordPolicy struct {
    MinLength           int
    RequireUppercase    bool
    RequireLowercase    bool
    RequireNumbers      bool
    RequireSpecial      bool
    MaxAge             time.Duration
    PreventReuse       int  // Last N passwords
    CheckBreached      bool // HIBP check
}

func (s *AuthService) ValidatePassword(password string, policy *PasswordPolicy) error {
    if len(password) < policy.MinLength {
        return ErrPasswordTooShort
    }
    if policy.CheckBreached {
        breached, _ := s.CheckPasswordBreach(password)
        if breached {
            return ErrPasswordBreached
        }
    }
    // ... other validations
}
```

---

### Missing-3: No Session Pinning

**Severity:** MEDIUM  
**Status:** MISSING  
**Impact:**
Bind session to IP/device fingerprint.

**Required Implementation:**
```go
func (s *SessionManager) CreateSession(ctx context.Context, req *CreateSessionRequest) (*Session, error) {
    session := &Session{
        ID:           uuid.New(),
        OperatorID:   req.OperatorID,
        IPAddress:    req.IP,
        DeviceID:     req.DeviceID,
        Fingerprint:  hashFingerprint(req),
        CreatedAt:   time.Now(),
        ExpiresAt:   time.Now().Add(req.TTL),
    }
    
    if err := s.repo.Create(ctx, session); err != nil {
        return nil, err
    }
    return session, nil
}

func (s *SessionManager) ValidateSession(ctx context.Context, sessionID string, fp *Fingerprint) error {
    session := s.Get(ctx, sessionID)
    
    // Strict mode: exact match required
    if session.Fingerprint != fp.Hash {
        return ErrSessionFingerprintMismatch
    }
    
    // Lenient mode: allow IP change with alert
    if session.IPAddress != fp.IP {
        s.AlertIPChange(ctx, session.ID, session.IPAddress, fp.IP)
    }
}
```

---

### Missing-4: No Emergency Lockout by Admin

**Severity:** HIGH  
**Status:** MISSING  
**Impact:**
Ability to lock account without knowing password.

**Required Implementation:**
```go
func (h *AdminHandler) EmergencyLockout(c *gin.Context) {
    operatorID := c.Param("operator_id")
    reason := c.Query("reason")
    
    // Immediately invalidate all sessions
    h.sessionManager.RevokeAllSessions(ctx, operatorID)
    
    // Set lockout flag
    h.operatorRepo.SetLockout(ctx, operatorID, true)
    
    // Notify operator via email
    h.emailService.SendLockoutNotification(ctx, operatorID, reason)
    
    // Log action
    h.auditLogger.Log(&AuditEvent{
        Action:      "emergency_lockout",
        TargetID:    operatorID,
        PerformedBy: currentOperatorID,
        Reason:     reason,
    })
}
```

---

### Missing-5: No Login Notification to User

**Severity:** MEDIUM  
**Status:** MISSING  
**Impact:**
Email/SMS when new login occurs.

**Required Implementation:**
```go
func (s *AuthService) NotifyNewLogin(ctx context.Context, operatorID string, login *LoginEvent) error {
    operator, _ := s.operatorRepo.FindByID(ctx, operatorID)
    
    template := &EmailTemplate{
        Subject: "New login to your account",
        Body: `
            New login detected:
            Time: {{.Time}}
            IP: {{.IP}}
            Location: {{.Location}}
            Device: {{.Device}}
            
            If this wasn't you, secure your account immediately.
        `,
        Data: login,
    }
    
    return s.emailService.Send(operator.Email, template)
}
```

---

### Missing-6: No Limited Session Count Per User

**Severity:** MEDIUM  
**Status:** MISSING  
**Impact:**
"Max 3 devices" enforcement.

**Required Implementation:**
```go
func (s *SessionManager) CreateSession(ctx context.Context, req *CreateSessionRequest) (*Session, error) {
    maxSessions := s.config.MaxConcurrentSessions  // Default: 3
    
    // Count existing sessions
    activeSessions := s.ListActiveSessions(ctx, req.OperatorID)
    
    if len(activeSessions) >= maxSessions {
        // Revoke oldest session
        oldest := activeSessions[0]
        s.RevokeSession(ctx, oldest.ID)
        
        h.auditLogger.Log(&AuditEvent{
            Action:    "session_revoked_max_reached",
            SessionID: oldest.ID,
        })
    }
    
    return s.createSession(ctx, req)
}
```

---

### Missing-7: No LDAP/AD Integration

**Severity:** MEDIUM  
**Status:** MISSING  
**Impact:**
Enterprise identity providers not supported.

**Required Implementation:**
```go
type LDAPConfig struct {
    Server     string
    Port       int
    BaseDN     string
    BindDN     string
    BindPass   string
    UserFilter string
}

func (s *AuthService) LDAPLogin(ctx context.Context, username, password string) (*LoginResponse, error) {
    // Connect to LDAP
    conn, err := ldap.DialURL(s.config.LDAP.Server)
    if err != nil {
        return nil, err
    }
    defer conn.Close()
    
    // Bind with user credentials
    if err := conn.Bind(username, password); err != nil {
        return nil, ErrInvalidCredentials
    }
    
    // Search for user attributes
    user, _ := s.searchLDAPUser(conn, username)
    
    // Create or update local operator
    return s.SyncOperatorFromLDAP(ctx, user)
}
```

---

### Missing-8: No SCIM Provisioning

**Severity:** MEDIUM  
**Status:** MISSING  
**Impact:**
Automated user lifecycle management.

**Required Implementation:**
```go
// SCIM endpoints
v1.POST("/scim/v2/Users", h.scim.CreateUser)
v1.GET("/scim/v2/Users/:id", h.scim.GetUser)
v1.PUT("/scim/v2/Users/:id", h.scim.UpdateUser)
v1.DELETE("/scim/v2/Users/:id", h.scim.DeleteUser)
v1.GET("/scim/v2/Users", h.scim.ListUsers)

// User provisioning from IdP
func (s *SCIMService) CreateUser(ctx context.Context, user *SCIMUser) (*Operator, error) {
    operator := &Operator{
        ExternalID: user.ID,
        Email:      user.Emails[0].Value,
        Name:       user.Name.GivenName + " " + user.Name.FamilyName,
        Active:     user.Active,
    }
    return s.operatorRepo.Create(ctx, operator)
}
```

---

### Missing-9: No WebAuthn/FIDO2 Support

**Severity:** MEDIUM  
**Status:** MISSING  
**Impact:**
Enterprise customers expect passwordless MFA options.

**Required Implementation:**
```go
type WebAuthnService struct {
    relyingParty *webauthn.WebAuthn
}

func (s *WebAuthnService) BeginRegistration(ctx context.Context, operatorID string) (*webauthn.CredentialCreation, error) {
    operator, _ := s.operatorRepo.FindByID(ctx, operatorID)
    
    return s.relyingParty.BeginRegistration(
        credential.NewUserEntity(
            operator.Name,
            operator.ID,
            operator.Email,
        ),
    )
}

func (s *WebAuthnService) FinishLogin(ctx context.Context, sessionData *webauthn.SessionData, response []byte) (*Operator, error) {
    return s.relyingParty.FinishLogin(sessionData, response)
}
```

---

### Missing-10: No Threat Detection/Response

**Severity:** HIGH  
**Status:** MISSING  
**Impact:**
No automated security response to threats.

**Required Implementation:**
```go
type ThreatDetector struct {
    rules []ThreatRule
}

type ThreatRule struct {
    Name        string
    Condition   func(*LoginEvent) bool
    Action      ThreatAction
    Severity    Severity
}

const (
    ThreatActionBlock         ThreatAction = "block"
    ThreatActionAlert          ThreatAction = "alert"
    ThreatActionMFAChallenge   ThreatAction = "mfa_challenge"
    ThreatActionLockout       ThreatAction = "lockout"
)

func (t *ThreatDetector) Evaluate(ctx context.Context, event *LoginEvent) (*ThreatResponse, error) {
    for _, rule := range t.rules {
        if rule.Condition(event) {
            return &ThreatResponse{
                Threat:   rule.Name,
                Action:   rule.Action,
                Severity: rule.Severity,
            }, nil
        }
    }
    return nil, nil
}

// Example rules
var ThreatRules = []ThreatRule{
    {
        Name: "impossible_travel",
        Condition: func(e *LoginEvent) bool {
            last := e.LastLogin
            distance := geo.Distance(last.Location, e.Location)
            timeDiff := e.Timestamp.Sub(last.Timestamp)
            speed := distance / timeDiff.Hours()
            return speed > 900 // km/h - faster than plane
        },
        Action:   ThreatActionMFAChallenge,
        Severity: SeverityHigh,
    },
    {
        Name: "new_geography",
        Condition: func(e *LoginEvent) bool {
            return !e.KnownLocations.Contains(e.Location)
        },
        Action:   ThreatActionAlert,
        Severity: SeverityMedium,
    },
}
```

---

## Implementation Checklist

### Phase 1: Critical Security Fixes (Week 1)

- [ ] CRITICAL-1: MFA Rate Limiting
- [ ] CRITICAL-2: Race Condition Fix
- [ ] CRITICAL-3: Cookie Error Handling
- [ ] CRITICAL-4: JWT in MFA Response
- [ ] CRITICAL-5: MFA Token Cryptographic Binding
- [ ] CRITICAL-6: OAuth State Persistence
- [ ] CRITICAL-7: Password Breach Checking
- [ ] CRITICAL-8: JWT Secret Validation

### Phase 2: High Priority Fixes (Week 2)

- [ ] HIGH-1: MFA Token Binding
- [ ] HIGH-2: Token Rotation Proper Revocation
- [ ] HIGH-3: Logout Revokes Refresh Tokens
- [ ] HIGH-4: Session Enumeration Endpoints
- [ ] HIGH-5: Global Token Reuse Detection
- [ ] HIGH-6: 90-Day Token Retention
- [ ] HIGH-7: MFA Enforcement
- [ ] HIGH-9: Concurrent Login Detection
- [ ] HIGH-10: Remember Me Feature

### Phase 3: Medium Priority (Week 3)

- [ ] MEDIUM-1: Operator ID Exposure Fix
- [ ] MEDIUM-2: MFA Audit Logging
- [ ] MEDIUM-3: MFA Lockout Protection
- [ ] MEDIUM-8: Session Fingerprinting
- [ ] MEDIUM-9: IP Anomaly Detection
- [ ] MEDIUM-10: Settings Service Deduplication

### Phase 4: Enterprise Features (Week 4-6)

- [ ] RBAC/Permissions System
- [ ] Password Policy Per Tenant
- [ ] Session Pinning
- [ ] Emergency Lockout
- [ ] Login Notifications
- [ ] Session Count Limits
- [ ] LDAP/AD Integration
- [ ] SCIM Provisioning
- [ ] WebAuthn/FIDO2
- [ ] Threat Detection

---

## File Locations Reference

| Component | Location |
|-----------|----------|
| Auth Handlers | `apps/api/internal/api/handlers/auth/` |
| Auth Service | `apps/api/internal/application/auth/` |
| Domain Entities | `apps/api/internal/domain/` |
| Session Storage | `apps/api/internal/infrastructure/storage/session_storage.go` |
| Refresh Token | `apps/api/internal/infrastructure/security/refresh_token.go` |
| MFA Entity | `apps/api/internal/domain/auth/mfa_token_entity.go` |
| OAuth | `apps/api/internal/api/handlers/auth/auth_oauth.go` |
| Routes | `apps/api/internal/api/server_routes.go` |
| Middleware | `apps/api/internal/api/middleware/` |

---

*Document Version: 1.0*  
*Status: Ready for Implementation*  
*Priority: CRITICAL - Security & Enterprise Readiness*
