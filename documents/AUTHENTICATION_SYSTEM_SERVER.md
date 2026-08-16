# Authentication System - Server Backend Requirements

> **Version:** 1.0  
> **Status:** Draft  
> **Created:** 2026-06-25  
> **Target:** Production MVP  
> **Architecture:** Domain-Driven (Following existing `apps/api` structure)

---

## Table of Contents

1. [Overview](#1-overview)
2. [Current Server Structure](#2-current-server-structure)
3. [Required Endpoints](#3-required-endpoints)
4. [Domain Layer](#4-domain-layer)
5. [Application Layer](#5-application-layer)
6. [Handler Specifications](#6-handler-specifications)
7. [Infrastructure Requirements](#7-infrastructure-requirements)
8. [Database Schema](#8-database-schema)
9. [File Structure](#9-file-structure)
10. [Implementation Order](#10-implementation-order)

---

>  **Architecture Alignment Note**
> 
> This document outlines the server-side requirements for authentication, aligned with the existing Go backend structure in `apps/api/internal/`. It complements `AUTHENTICATION_SYSTEM.md` (frontend) and `FRONTEND_ARCHITECTURE.md`.
>
> **Dependency Flow:** Handler → Application Service → Domain → Infrastructure (Storage)

---

## 1. Overview

### 1.1 Purpose

This document specifies the server-side authentication requirements to support:
- **Credential Authentication** - Email/password login, registration
- **OAuth** - Google OAuth 2.0 integration
- **MFA (TOTP)** - Time-based one-time passwords with backup codes
- **Session Management** - JWT tokens, refresh token rotation
- **Password Reset** - Forgot password, email-based reset flow
- **Email Verification** - Account verification links

### 1.2 Frontend Requirements Summary

| Feature | Frontend Hook | Backend Endpoint |
|---------|---------------|----------------|
| Login | `use-login.ts` | `POST /v1/auth/login` |
| Register | `use-register.ts` | `POST /v1/auth/register` |
| Logout | `use-logout.ts` | `POST /v1/auth/logout` |
| Get Current User | `use-session.ts` | `GET /v1/auth/me` |
| MFA Verify | `use-mfa.ts` | `POST /v1/auth/mfa/verify` |
| MFA Status | - | `GET /v1/auth/mfa/status` |
| MFA Enroll | - | `POST /v1/auth/mfa/enroll` |
| Forgot Password | `use-password-reset.ts` | `POST /v1/auth/forgot-password` |
| Reset Password | `use-password-reset.ts` | `POST /v1/auth/reset-password` |
| OAuth Google | `use-auth-callback.ts` | `GET /v1/auth/google` |
| Refresh Token | - | `POST /v1/auth/refresh` |

---

## 2. Current Server Structure

### 2.1 Existing Auth Structure

```
apps/api/internal/
 api/
    handlers/auth/           # EXISTING
       auth_login.go        # EXISTS
       auth_register.go    # EXISTS
       auth_logout.go      # EXISTS
       auth_me.go          # EXISTS
       auth_mfa.go         # EXISTS
       auth_oauth.go       # EXISTS
       auth_password_reset.go    # EXISTS
       auth_email_verify.go      # EXISTS
       auth_settings.go    # EXISTS
       auth_admin.go       # EXISTS
       auth_routes.go      # EXISTS
       auth_lockout.go     # EXISTS
    middleware/
 application/auth/           # EXISTS
    auth_service.go        # EXISTS (AuthService)
    auth_password.go       # EXISTS
 domain/
    operator/               # EXISTS
       operator_entity.go           # EXISTS (Operator struct)
       operator_repository.go       # EXISTS
       operator_errors.go           # EXISTS
       operator_password.go         # EXISTS
       operator_role.go            # EXISTS
       operator_requests.go         # EXISTS
       operator_responses.go       # EXISTS
    session/                # EXISTS
       session_entity.go          # EXISTS
       session_repository.go      # EXISTS
    refresh_token/          # EXISTS
       refresh_token_entity.go          # EXISTS
       refresh_token_repository.go      # EXISTS
    email_verification/     # EXISTS
       email_verification_entity.go          # EXISTS
       email_verification_repository.go      # EXISTS
       email_verification_requests.go       # EXISTS
       email_verification_responses.go      # EXISTS
    password_reset/         # EXISTS
       password_reset_entity.go          # EXISTS
       password_reset_repository.go      # EXISTS
       password_reset_requests.go        # EXISTS
       password_reset_responses.go       # EXISTS
    oauth/                  # EXISTS
        oauth_entity.go          # EXISTS
        oauth_errors.go         # EXISTS
 infrastructure/
     security/              # EXISTS
     email/                  # EXISTS
     storage/                # EXISTS
```

### 2.2 Existing Handler Methods

| Handler | File | Status | Methods |
|----------|------|--------|---------|
| LoginHandler | `handlers/auth/auth_login.go` | EXISTS | Handle |
| RegisterHandler | `handlers/auth/auth_register.go` | EXISTS | Handle |
| LogoutHandler | `handlers/auth/auth_logout.go` | EXISTS | Handle |
| MeHandler | `handlers/auth/auth_me.go` | EXISTS | Handle |
| MFAHandler | `handlers/auth/auth_mfa.go` | EXISTS | GetMFAStatus, EnrollMFA, VerifySetupMFA, EnableMFA, DisableMFA, VerifyBackupCode, RegenerateBackupCodes |
| OAuthHandler | `handlers/auth/auth_oauth.go` | EXISTS | GoogleLogin, GoogleCallback, GitHubLogin, GitHubCallback |
| PasswordResetHandler | `handlers/auth/password_reset.go` | EXISTS | ForgotPassword, ResetPassword, ResendPasswordReset |
| EmailVerifyHandler | `handlers/auth/email_verify.go` | EXISTS | VerifyEmail, ResendVerification, CancelVerification, PollVerification |
| SettingsHandler | `handlers/auth/auth_settings.go` | EXISTS | UpdateName, UpdateSettings |
| AdminHandler | `handlers/auth/auth_admin.go` | EXISTS | ListOperators, CreateOperator, GetOperator, UpdateOperator, DeleteOperator |
| LockoutHandler | `handlers/auth/auth_lockout.go` | EXISTS | GetLockoutStatus, UnlockAccount |

---

## 3. Required Endpoints

### 3.1 Authentication Endpoints

| Method | Endpoint | Auth | Description | Status |
|--------|----------|------|-------------|--------|
| POST | `/v1/auth/login` | None | Credential login (browser — sets session cookie) | **EXISTS** |
| POST | `/v1/auth/login/tokens` | None | Credential login (API clients — returns JWT + refresh token) | **EXISTS** |
| POST | `/v1/auth/register` | None | Register new operator | **EXISTS** |
| POST | `/v1/auth/logout` | Cookie | Logout current session | **EXISTS** |
| GET | `/v1/auth/csrf-token` | None | Get CSRF token for mutating requests | **EXISTS** |
| GET | `/v1/auth/me` | Cookie | Get current operator | **EXISTS** |
| PATCH | `/v1/auth/me` | Cookie | Update operator name | **EXISTS** |
| GET | `/v1/auth/me/settings` | Cookie | Get operator settings (thresholds + client) | **EXISTS** |
| PATCH | `/v1/auth/me/settings` | Cookie | Update operator settings | **EXISTS** |
| GET | `/v1/auth/me/thresholds` | Cookie | Get operator thresholds | **EXISTS** |
| PATCH | `/v1/auth/me/thresholds` | Cookie | Update operator thresholds | **EXISTS** |
| GET | `/v1/auth/me/notifications` | Cookie | Get notification settings | **EXISTS** |
| PATCH | `/v1/auth/me/notifications` | Cookie | Update notification settings | **EXISTS** |
| POST | `/v1/auth/me/notifications/webhook/test` | Cookie | Test webhook delivery | **EXISTS** |
| POST | `/v1/auth/me/notifications/webhook/rotate` | Cookie | Rotate webhook secret | **EXISTS** |
| POST | `/v1/auth/refresh` | None | Refresh access token (requires `refresh_token` in body) | **EXISTS** |
| GET | `/v1/auth/lockout/status` | Cookie | Get current operator's lockout status | **EXISTS** |
| GET | `/v1/auth/organizations` | Cookie | List operator's organizations | **EXISTS** |
| POST | `/v1/auth/organizations/select` | Cookie | Select active organization | **EXISTS** |

### 3.2 MFA Endpoints

| Method | Endpoint | Auth | Description | Status |
|--------|----------|------|-------------|--------|
| GET | `/v1/auth/mfa/status` | Cookie | Get MFA enrollment status | **EXISTS** |
| POST | `/v1/auth/mfa/enroll` | Cookie | Start MFA enrollment | **EXISTS** |
| POST | `/v1/auth/mfa/verify-setup` | Cookie | Verify MFA setup (expects `{ code, token }`) | **EXISTS** |
| POST | `/v1/auth/mfa/enable` | Cookie | Enable MFA (expects `{ code, token }`, returns backup codes) | **EXISTS** |
| POST | `/v1/auth/mfa/disable` | Cookie | Disable MFA (expects `{ code }`) | **EXISTS** |
| POST | `/v1/auth/mfa/verify-backup` | Cookie | Verify backup code | **EXISTS** |
| POST | `/v1/auth/mfa/regenerate-backup-codes` | Cookie | Regenerate backup codes | **EXISTS** |
| POST | `/v1/auth/mfa/verify` | None | Verify MFA code during login (expects `{ operator_id, code }`) | **EXISTS** |

### 3.3 Password Reset Endpoints

| Method | Endpoint | Auth | Description | Status |
|--------|----------|------|-------------|--------|
| POST | `/v1/auth/forgot-password` | None | Request password reset | **EXISTS** |
| POST | `/v1/auth/reset-password` | None | Reset with token (expects `{ token, newPassword }`) | **EXISTS** |
| POST | `/v1/auth/resend-password-reset` | None | Resend reset email (rate-limited) | **EXISTS** |

### 3.4 Email Verification Endpoints

| Method | Endpoint | Auth | Description | Status |
|--------|----------|------|-------------|--------|
| POST | `/v1/auth/verify-email` | None | Verify email token | **EXISTS** |
| GET | `/v1/auth/verify-email` | None | Verify email token (GET variant) | **EXISTS** |
| POST | `/v1/auth/resend-verification` | None | Resend verification | **EXISTS** |
| GET | `/v1/auth/resend-verification` | None | Resend verification (GET variant) | **EXISTS** |
| POST | `/v1/auth/cancel-verification` | None | Cancel verification | **EXISTS** |
| GET | `/v1/auth/poll-verification` | None | Poll verification status | **EXISTS** |

### 3.5 OAuth Endpoints

| Method | Endpoint | Auth | Description | Status |
|--------|----------|------|-------------|--------|
| GET | `/v1/auth/google` | None | Initiate Google OAuth | **EXISTS** |
| GET | `/v1/auth/google/callback` | None | Google OAuth callback | **EXISTS** |
| GET | `/v1/auth/github` | None | Initiate GitHub OAuth | **EXISTS** |
| GET | `/v1/auth/github/callback` | None | GitHub OAuth callback | **EXISTS** |

### 3.6 Session Management Endpoints

| Method | Endpoint | Auth | Description | Status |
|--------|----------|------|-------------|--------|
| GET | `/v1/auth/sessions` | Cookie | List all active sessions for the operator | **EXISTS** |
| GET | `/v1/auth/sessions/:id` | Cookie | Get a specific session by ID | **EXISTS** |
| GET | `/v1/auth/sessions/concurrent` | Cookie | Check concurrent login count and limit | **EXISTS** |
| DELETE | `/v1/auth/sessions/:id` | Cookie | Revoke a specific session | **EXISTS** |
| DELETE | `/v1/auth/sessions` | Cookie | Revoke all sessions except current | **EXISTS** |
| POST | `/v1/auth/sessions/revoke-all` | Cookie | Logout from all devices | **EXISTS** |

### 3.7 Client Credentials Endpoints

| Method | Endpoint | Auth | Description | Status |
|--------|----------|------|-------------|--------|
| POST | `/v1/auth/client-credentials` | Cookie + Org | Create client credentials (API key) | **EXISTS** |
| GET | `/v1/auth/client-credentials` | Cookie + Org | List client credentials | **EXISTS** |
| GET | `/v1/auth/client-credentials/:clientId` | Cookie + Org | Get a specific client credential | **EXISTS** |
| PATCH | `/v1/auth/client-credentials/:clientId` | Cookie + Org | Update client credentials | **EXISTS** |
| DELETE | `/v1/auth/client-credentials/:clientId` | Cookie + Org | Delete client credentials | **EXISTS** |
| POST | `/v1/auth/client-credentials/:clientId/rotate-secret` | Cookie + Org | Rotate client secret | **EXISTS** |

### 3.8 Admin Endpoints

| Method | Endpoint | Auth | Description | Status |
|--------|----------|------|-------------|--------|
| GET | `/v1/auth/admin/operators` | Admin | List all operators | **EXISTS** |
| POST | `/v1/auth/admin/operators` | Admin | Create operator | **EXISTS** |
| GET | `/v1/auth/admin/operators/:id` | Admin | Get operator | **EXISTS** |
| PATCH | `/v1/auth/admin/operators/:id` | Admin | Update operator | **EXISTS** |
| DELETE | `/v1/auth/admin/operators/:id` | Admin | Delete operator | **EXISTS** |

---

## 4. Domain Layer

### 4.1 Operator Entity (EXISTING)

```go
// apps/api/internal/domain/operator/operator_entity.go

type Operator struct {
    CreatedAt      time.Time
    UpdatedAt      time.Time
    GitHubID       string
    PasswordHash   string
    Role           OperatorRole
    GoogleID       string
    ID             string
    MFASecret      string
    Name           string
    Email          string
    BackupCodes    []string
    Thresholds     Thresholds
    ClientSettings ClientSettings
    MFAEnabled     bool
    EmailVerified  bool
}

// Methods (EXISTING)
func (o *Operator) IsSuperAdmin() bool
func (o *Operator) IsAdmin() bool
func (o *Operator) HasMFA() bool
func (o *Operator) UsesOAuth() bool
func (o *Operator) HasPassword() bool
```

### 4.2 Required Domain Entities

| Entity | File | Status | Purpose |
|--------|------|--------|---------|
| Operator | `domain/operator/operator_entity.go` | **EXISTS** | User/operator entity |
| Session | `domain/session/session_entity.go` | **EXISTS** | Session management |
| RefreshToken | `domain/refresh_token/refresh_token_entity.go` | **EXISTS** | Refresh token storage |
| EmailVerification | `domain/email_verification/email_verification_entity.go` | **EXISTS** | Email verification tokens |
| PasswordReset | `domain/password_reset/password_reset_entity.go` | **EXISTS** | Password reset tokens |
| OAuth | `domain/oauth/oauth_entity.go` | **EXISTS** | OAuth state management |

### 4.3 Domain Repository Interfaces

```go
// domain/operator/operator_repository.go (EXISTING)

type Repository interface {
    Create(ctx context.Context, op *Operator) error
    FindByID(ctx context.Context, id string) (*Operator, error)
    FindByEmail(ctx context.Context, email string) (*Operator, error)
    FindByGoogleID(ctx context.Context, googleID string) (*Operator, error)
    FindByGitHubID(ctx context.Context, githubID string) (*Operator, error)
    Update(ctx context.Context, op *Operator) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, limit, offset int) ([]*Operator, int, error)
    Count(ctx context.Context) (int, error)
}
```

### 4.4 Session Repository (EXISTING)

```go
// domain/session/session_repository.go (EXISTING)

type Repository interface {
    Create(ctx context.Context, sess *Session) error
    FindByID(ctx context.Context, id string) (*Session, error)
    FindByOperatorID(ctx context.Context, operatorID string) ([]*Session, error)
    Delete(ctx context.Context, id string) error
    DeleteByOperatorID(ctx context.Context, operatorID string) error
    CountByOperatorID(ctx context.Context, operatorID string) (int, error)
}
```

---

## 5. Application Layer

### 5.1 Auth Service (EXISTING)

```go
// application/auth/auth_service.go (EXISTS - major methods)

type AuthService struct {
    operatorRepo       operator.Repository
    sessionRepo        session.Repository
    emailVerifyRepo    email_verification.Repository
    passwordResetRepo  password_reset.Repository
    passwordHasher     PasswordHasher
    refreshTokenRepo   RefreshTokenRepository
    jwtManager         *infraauth.JWTManager
    sessionManager     *infraSession.Manager
    sessionTTL         time.Duration
    refreshTokenExpiry time.Duration
}

// Existing Methods
func (s *AuthService) Login(ctx, req) (*dto.LoginResponse, *session.Session, error)
func (s *AuthService) Register(ctx, req, validatePassword) (*dto.RegisterResponse, error)
func (s *AuthService) Logout(ctx, operatorID string, allDevices bool) error
func (s *AuthService) CreateSession(ctx, operatorID string) (*session.Session, error)
func (s *AuthService) VerifyPassword(password, hash string) error
func (s *AuthService) HashPassword(password string) (string, error)

// MFA Methods (EXISTING)
func (s *AuthService) GetMFAStatus(ctx, operatorID) (*dto.MFAStatusResponse, error)
func (s *AuthService) EnrollMFA(ctx, operatorID) (string, string, error) // Returns secret, QRCodeURL
func (s *AuthService) VerifyMFAEnrollment(ctx, operatorID, code string) error
func (s *AuthService) EnableMFA(ctx, operatorID, code string) error
func (s *AuthService) DisableMFA(ctx, operatorID, code string) error
func (s *AuthService) VerifyMFACode(ctx, operatorID, code string) error
func (s *AuthService) RegenerateBackupCodes(ctx, operatorID string) ([]string, error)

// Password Methods (EXISTING)
func (s *AuthService) ChangePassword(ctx, operatorID, oldPassword, newPassword string) error
func (s *AuthService) ResetPassword(ctx, token, newPassword string) error
func (s *AuthService) InitiatePasswordReset(ctx, email string) error

// OAuth Methods (EXISTING)
func (s *AuthService) HandleGoogleCallback(ctx, code string) (*dto.LoginResponse, error)
func (s *AuthService) HandleGitHubCallback(ctx, code string) (*dto.LoginResponse, error)

// Session Management
func (s *AuthService) RotateRefreshToken(ctx, oldRefreshToken string) (*RefreshTokenResult, error)
func (s *AuthService) IssueRefreshToken(ctx, operatorID, sessionID string) (string, error)
func (s *AuthService) RevokeAllRefreshTokens(ctx, operatorID string) error
```

### 5.2 Application DTOs

```go
// application/dto/auth.go (EXISTS)

package dto

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginResponse struct {
    SelectedOrganization *OrganizationInfo  `json:"selected_organization,omitempty"`
    OperatorID           string             `json:"operator_id"`
    Email                string             `json:"email"`
    Name                 string             `json:"name"`
    LastOrganizationID   string             `json:"last_organization_id,omitempty"`
    Organizations        []OrganizationInfo `json:"organizations,omitempty"`
    MFAEnabled           bool               `json:"mfa_enabled"`
    NeedsOrganization    bool               `json:"needs_organization"`
}

type LoginWithTokensResponse struct {
    SelectedOrganization *OrganizationInfo  `json:"selected_organization,omitempty"`
    OperatorID           string             `json:"operator_id"`
    Email                string             `json:"email"`
    Name                 string             `json:"name"`
    LastOrganizationID   string             `json:"last_organization_id,omitempty"`
    AccessToken          string             `json:"access_token"`
    RefreshToken         string             `json:"refresh_token"`
    SessionID            string             `json:"session_id"`
    Organizations        []OrganizationInfo `json:"organizations,omitempty"`
    ExpiresAt            int64              `json:"expires_at"`
    MFAEnabled           bool               `json:"mfa_enabled"`
    NeedsOrganization    bool               `json:"needs_organization"`
}

type RegisterRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    Name     string `json:"name"`
    Role     string `json:"role,omitempty"`
}

type RegisterResponse struct {
    OperatorID string `json:"operator_id"`
    Email      string `json:"email"`
    Name       string `json:"name"`
}

type ForgotPasswordRequest struct {
    Email string `json:"email"`
}

type ResetPasswordRequest struct {
    Token       string `json:"token"`
    NewPassword string `json:"new_password"`
}

type ChangePasswordRequest struct {
    OldPassword string `json:"old_password"`
    NewPassword string `json:"new_password"`
}

// MFA Verification (post-login) — EXISTS
type MFAVerifyRequest struct {
    OperatorID string `json:"operator_id"`
    Code       string `json:"code"`
}

// MFAVerifyResponse is returned by POST /v1/auth/mfa/verify during login.
// The handler returns a flat gin.H with success, session_id, access_token,
// refresh_token, expires_at, and a nested operator object.
type MFAVerifyResponse struct {
    Success      bool   `json:"success"`
    SessionID    string `json:"session_id"`
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresAt    int64  `json:"expires_at"`
    Operator     struct {
        ID         string `json:"id"`
        Email      string `json:"email"`
        Name       string `json:"name"`
        Role       string `json:"role"`
        MFAEnabled bool   `json:"mfa_enabled"`
    } `json:"operator"`
}

// Token Refresh — EXISTS
type RefreshTokenRequest struct {
    RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    SessionID    string `json:"session_id"`
    ExpiresAt    int64  `json:"expires_at"`
}

// MFAStatusResponse — EXISTS (handler returns { "mfa_enabled": bool })
type MFAStatusResponse struct {
    BackupCodes []string `json:"backup_codes,omitempty"`
    Enabled     bool     `json:"enabled"`
}

// OperatorResponse — returned by GET /v1/auth/me
type OperatorResponse struct {
    Thresholds           *Thresholds        `json:"thresholds,omitempty"`
    Client               *ClientSettings    `json:"client,omitempty"`
    SelectedOrganization *OrganizationInfo  `json:"selected_organization,omitempty"`
    ID                   string             `json:"id"`
    Email                string             `json:"email"`
    Name                 string             `json:"name"`
    LastOrganizationID   string             `json:"last_organization_id,omitempty"`
    CreatedAt            string             `json:"created_at"`
    Organizations        []OrganizationInfo `json:"organizations,omitempty"`
    NeedsOrganization    bool               `json:"needs_organization"`
    MFAEnabled           bool               `json:"mfa_enabled"`
    EmailVerified        bool               `json:"email_verified"`
}
```

---

## 6. Handler Specifications

### 6.1 Login Handler (EXISTS)

```go
// handlers/auth/auth_login.go

type LoginHandler struct {
    authService *auth.AuthService
    presenter  *response.Presenter
}

// POST /v1/auth/login — browser login (sets session cookie)
func (h *LoginHandler) Handle(c *gin.Context) {
    // 1. Parse LoginRequest
    // 2. Call authService.Login()
    // 3. If ErrMFARequired, return { mfa_required: true, operator_id }
    // 4. Set session cookie (AES-GCM encrypted vyz_session)
    // 5. Return LoginResponse (no tokens — cookie-based)
}

// POST /v1/auth/login/tokens — API client login (returns JWT + refresh token)
func (h *LoginHandler) HandleWithTokens(c *gin.Context) {
    // 1. Parse LoginRequest
    // 2. Call authService.LoginWithTokens()
    // 3. If ErrMFARequired, return { mfa_required: true, operator_id, email, name, mfa_enabled }
    // 4. Return LoginWithTokensResponse (access_token, refresh_token, expires_at, session_id, operator fields)
}
```

### 6.2 MFA Verification Handler (EXISTS)

```go
// handlers/auth/auth_mfa.go — MFAHandler.VerifyMFA

// POST /v1/auth/mfa/verify
// Rate-limited by operator_id (or IP fallback), protected by Lockout middleware.
func (h *MFAHandler) VerifyMFA(c *gin.Context) {
    // 1. Parse { operator_id, code } (both required)
    // 2. Audit-log MFA verify attempt
    // 3. Call authService.VerifyMFACode(operatorID, code) → returns session
    // 4. On failure: audit-log, return 401 "Invalid MFA code"
    // 5. Re-validate operator (not deleted, MFA still enabled, role unchanged)
    // 6. Create session cookie via SessionManager.CreateCookie(session.ID)
    // 7. Issue refresh token + generate JWT access token
    // 8. Audit-log success
    // 9. Return { success, session_id, access_token, refresh_token, expires_at, operator: { id, email, name, role, mfa_enabled } }
}
```

### 6.3 Refresh Token Handler (EXISTS)

```go
// handlers/auth/auth_refresh.go — RefreshHandler

// POST /v1/auth/refresh
// Expects { refresh_token } in the request body (not cookie-only).
func (h *RefreshHandler) Handle(c *gin.Context) {
    // 1. Parse RefreshTokenRequest { refresh_token }
    // 2. Call authService.RotateRefreshToken(oldRefreshToken)
    //    - Validates the old token
    //    - Revokes old token (sets replaced_by_id for theft detection)
    //    - Issues new refresh token + access token
    // 3. Return RefreshTokenResponse { access_token, refresh_token, session_id, expires_at }
    // 4. On failure (invalid/reused token): clear auth, return 401
}
```

### 6.4 Settings Handler Extensions (EXISTING - verify coverage)

```go
// handlers/auth/settings.go (EXISTS - verify all methods)

type SettingsHandler struct {
    authService *auth.AuthService
    presenter  *response.Presenter
}

// PATCH /v1/auth/me
func (h *SettingsHandler) UpdateName(c *gin.Context) {
    // Update operator name
}

// PATCH /v1/auth/me/settings
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
    // Update thresholds, client settings
}
```

---

## 7. Infrastructure Requirements

### 7.1 Security Infrastructure (EXISTING)

| Component | Location | Purpose |
|----------|----------|---------|
| JWTManager | `infrastructure/security/jwt.go` | JWT generation/verification |
| SessionManager | `infrastructure/security/session/` | Session management |
| PasswordHasher | `infrastructure/security/` | Argon2id hashing |
| GoogleVerifier | `infrastructure/security/` | Google OAuth token verification |
| RateLimiter | `infrastructure/security/` | Rate limiting |
| Lockout | `middleware/lockout.go` | Account lockout |

### 7.2 Email Infrastructure (EXISTING)

| Component | Location | Purpose |
|----------|----------|---------|
| EmailService | `infrastructure/email/` | Email sending |
| Templates | `infrastructure/email/` | Email templates |

### 7.3 Required Security Features

| Feature | Status | Implementation |
|---------|--------|----------------|
| JWT Access Tokens | **EXISTS** | `infrastructure/security/jwt.go` |
| Refresh Token Rotation | **EXISTS** | `domain/refresh_token/` |
| Password Hashing (Argon2id) | **EXISTS** | `infrastructure/security/` |
| Session Cookies | **EXISTS** | Cookie-based auth middleware |
| Rate Limiting | **EXISTS** | `middleware/rate_limit.go` |
| Account Lockout | **EXISTS** | `middleware/lockout.go` |
| MFA/TOTP | **EXISTS** | `application/auth/mfa.go` |
| OAuth (Google) | **EXISTS** | `handlers/auth/oauth.go` |
| OAuth (GitHub) | **EXISTS** | `handlers/auth/oauth.go` |

---

## 8. Database Schema

### 8.1 Existing Tables

```sql
-- operators table (EXISTS)
CREATE TABLE operators (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    password_hash TEXT,
    google_id TEXT UNIQUE,
    github_id TEXT UNIQUE,
    mfa_secret TEXT,
    backup_codes TEXT, -- JSON array
    mfa_enabled BOOLEAN DEFAULT FALSE,
    email_verified BOOLEAN DEFAULT FALSE,
    role TEXT NOT NULL DEFAULT 'operator',
    thresholds JSONB DEFAULT '{}',
    client_settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- sessions table (EXISTS)
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    operator_id TEXT NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    ip_address TEXT,
    user_agent TEXT
);

-- refresh_tokens table (EXISTS)
CREATE TABLE refresh_tokens (
    id TEXT PRIMARY KEY,
    token_hash TEXT UNIQUE NOT NULL,
    operator_id TEXT NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    session_id TEXT REFERENCES sessions(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    replaced_by_id TEXT,
    revoked BOOLEAN DEFAULT FALSE
);

-- email_verifications table (EXISTS)
CREATE TABLE email_verifications (
    id TEXT PRIMARY KEY,
    operator_id TEXT NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    token_hash TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- password_resets table (EXISTS)
CREATE TABLE password_resets (
    id TEXT PRIMARY KEY,
    operator_id TEXT NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    token_hash TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### 8.2 Indexes (EXISTING)

```sql
CREATE INDEX idx_operators_email ON operators(email);
CREATE INDEX idx_operators_google_id ON operators(google_id);
CREATE INDEX idx_operators_github_id ON operators(github_id);
CREATE INDEX idx_sessions_operator_id ON sessions(operator_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_operator_id ON refresh_tokens(operator_id);
CREATE INDEX idx_email_verifications_operator_id ON email_verifications(operator_id);
CREATE INDEX idx_email_verifications_token_hash ON email_verifications(token_hash);
CREATE INDEX idx_password_resets_operator_id ON password_resets(operator_id);
CREATE INDEX idx_password_resets_token_hash ON password_resets(token_hash);
```

---

## 9. File Structure

### 9.1 Current Auth Structure (apps/api/internal/)

```
apps/api/internal/
 api/
    handlers/auth/
       auth_login.go            # EXISTS — login + login/tokens
       auth_register.go         # EXISTS
       auth_logout.go           # EXISTS
       auth_me.go               # EXISTS
       auth_mfa.go              # EXISTS — all MFA endpoints incl. verify
       auth_refresh.go          # EXISTS — token refresh
       auth_oauth.go            # EXISTS — Google + GitHub
       auth_password_reset.go   # EXISTS — forgot/reset/resend
       auth_email_verify.go     # EXISTS — verify/resend/cancel/poll
       auth_settings.go         # EXISTS — name/settings/thresholds/notifications
       auth_admin.go            # EXISTS — operator CRUD
       auth_lockout.go          # EXISTS — lockout status + admin unlock
       auth_sessions.go         # EXISTS — session list/revoke/concurrent
       auth_client_credentials.go # EXISTS — API key CRUD + rotate
       auth_organization.go     # EXISTS — org list + select
       auth_routes.go           # EXISTS — route registration

    middleware/
       cookie_auth.go       # EXISTS — AES-GCM session cookie middleware
       lockout.go           # EXISTS
       rate_limit.go        # EXISTS
       validation.go        # EXISTS — request schema validation

    adapters/response/
        presenter.go        # EXISTS

 application/auth/
    auth_service.go              # EXISTS
    auth_password.go             # EXISTS
    auth_login_session.go        # EXISTS
    auth_device_recognition.go   # EXISTS

 domain/
    operator/
       operator_entity.go           # EXISTS
       operator_repository.go       # EXISTS
       operator_errors.go           # EXISTS
       operator_password.go         # EXISTS
       operator_role.go             # EXISTS

    session/
       session_entity.go          # EXISTS
       session_repository.go      # EXISTS

    refresh_token/
       refresh_token_entity.go          # EXISTS
       refresh_token_repository.go      # EXISTS

    email_verification/
       email_verification_entity.go          # EXISTS
       email_verification_repository.go      # EXISTS

    password_reset/
       password_reset_entity.go          # EXISTS
       password_reset_repository.go      # EXISTS

    oauth/
        oauth_entity.go          # EXISTS

 infrastructure/
     security/
        jwt.go             # EXISTS
        password.go        # EXISTS — Argon2id
        google.go          # EXISTS
        session/           # EXISTS — AES-256-GCM encrypted cookies

     email/
        email_service.go   # EXISTS

     storage/
         auth_storage.go           # EXISTS
         session_storage.go        # EXISTS
         operator_storage.go       # EXISTS
```

### 9.2 Files (Actual Implementation)

| Category | File | Handler/Service |
|----------|------|----------------|
| Handler | `internal/api/handlers/auth/auth_mfa.go` | AuthMfaHandler |
| Handler | `internal/api/handlers/auth/auth_refresh.go` | RefreshHandler |
| Handler | `internal/api/handlers/auth/auth_routes.go` | Route registration |
| Service | `internal/application/auth/auth_service.go` | AuthService |

---

## 10. Implementation Status

All auth endpoints are fully implemented. The phases below describe the completed work.

### Phase 1: Core Auth (COMPLETE)
- Login (browser + tokens), register, logout, refresh, /me
- Session cookies (AES-256-GCM encrypted `vyz_session`)
- JWT access tokens (HS256, 15-minute expiry)
- Refresh token rotation with theft detection (`replaced_by_id` chain)

### Phase 2: MFA (COMPLETE)
- TOTP enrollment, verify-setup, enable, disable
- Backup codes (generate, verify, regenerate)
- MFA verification during login (`POST /v1/auth/mfa/verify`)
- Rate-limited by operator_id, protected by lockout middleware

### Phase 3: Password Reset + Email Verification (COMPLETE)
- Forgot/reset/resend password flows
- Email verification (POST + GET variants, cancel, poll)
- Argon2id password hashing
- Token lifecycles: password-reset 15min, email-verify 24h

### Phase 4: OAuth (COMPLETE)
- Google OAuth 2.0 (login + callback)
- GitHub OAuth (login + callback)
- State-parameter CSRF protection

### Phase 5: Session Management (COMPLETE)
- Session list, get, revoke (single + all-except-current + all-devices)
- Concurrent session limit (default 5, configurable, oldest revoked)

### Phase 6: Admin + Client Credentials (COMPLETE)
- Operator CRUD (SuperAdmin only)
- Account unlock (SuperAdmin)
- Client credentials (API key) CRUD + secret rotation

### Phase 7: Security Hardening (COMPLETE)
- Account lockout (configurable thresholds)
- Rate limiting (login, MFA verify, password reset, settings)
- IP intelligence middleware
- User enumeration prevention
- CSRF token middleware for mutating requests
- Audit logging for auth events

---

## Appendix: API Contract Reference

### Login (Browser)
```
POST /v1/auth/login
Request: { "email": "...", "password": "..." }
Response: {
  "operator_id": "...",
  "email": "...",
  "name": "...",
  "mfa_enabled": true/false,
  "needs_organization": true/false,
  "organizations": [...],
  "selected_organization": { "id": "...", "name": "...", "role": "..." },
  "last_organization_id": "..."
}
MFA Response: { "mfa_required": true, "operator_id": "..." }
Errors: 400, 401, 429
```

### Login (API Clients — Tokens)
```
POST /v1/auth/login/tokens
Request: { "email": "...", "password": "..." }
Response: {
  "operator_id": "...",
  "email": "...",
  "name": "...",
  "mfa_enabled": true/false,
  "needs_organization": true/false,
  "organizations": [...],
  "access_token": "...",
  "refresh_token": "...",
  "expires_at": 1234567890,
  "session_id": "..."
}
MFA Response: { "mfa_required": true, "operator_id": "...", "email": "...", "name": "...", "mfa_enabled": true }
Errors: 400, 401, 429
```

### MFA Verification (Login)
```
POST /v1/auth/mfa/verify
Request: { "operator_id": "...", "code": "123456" }
Response: {
  "success": true,
  "session_id": "...",
  "access_token": "...",
  "refresh_token": "...",
  "expires_at": 1234567890,
  "operator": { "id": "...", "email": "...", "name": "...", "role": "...", "mfa_enabled": true }
}
Errors: 400, 401, 429
```

### MFA Status
```
GET /v1/auth/mfa/status
Response: { "mfa_enabled": true/false }
```

### MFA Enroll
```
POST /v1/auth/mfa/enroll
Response: { "secret": "...", "uri": "otpauth://..." }
```

### MFA Verify Setup
```
POST /v1/auth/mfa/verify-setup
Request: { "code": "123456", "token": "123456" }
Response: { "verified": true }
```

### MFA Enable
```
POST /v1/auth/mfa/enable
Request: { "code": "123456", "token": "123456" }
Response: { "success": true, "backup_codes": ["...", "...", ...] }
```

### MFA Disable
```
POST /v1/auth/mfa/disable
Request: { "code": "123456" }
Response: { "success": true }
```

### MFA Verify Backup Code
```
POST /v1/auth/mfa/verify-backup
Request: { "code": "..." }
Response: { "valid": true/false }
```

### MFA Regenerate Backup Codes
```
POST /v1/auth/mfa/regenerate-backup-codes
Response: { "backup_codes": ["...", "...", ...] }
```

### Refresh Token
```
POST /v1/auth/refresh
Request: { "refresh_token": "..." }
Response: { "access_token": "...", "refresh_token": "...", "expires_at": 1234567890, "session_id": "..." }
Errors: 400, 401
```

### Get Current User
```
GET /v1/auth/me
Response: {
  "id": "...",
  "email": "...",
  "name": "...",
  "mfa_enabled": true,
  "email_verified": true,
  "needs_organization": false,
  "organizations": [...],
  "selected_organization": { "id": "...", "name": "...", "role": "..." },
  "last_organization_id": "...",
  "created_at": "2026-01-01T00:00:00Z",
  "thresholds": { ... },
  "client": { ... }
}
Errors: 401
```

### Logout
```
POST /v1/auth/logout
Response: { "success": true }
Errors: 401
```

### Forgot Password
```
POST /v1/auth/forgot-password
Request: { "email": "..." }
Response: { "message": "If that email exists, a password reset link has been sent." }
Errors: 400, 429
```

### Reset Password
```
POST /v1/auth/reset-password
Request: { "token": "...", "newPassword": "..." }
Response: { "success": true, "message": "Password has been reset successfully." }
Errors: 400, 401
```

### Resend Password Reset
```
POST /v1/auth/resend-password-reset
Request: { "email": "..." }
Response (success): { "success": true, "message": "Password reset link sent." }
Response (rate-limited): { "error": "rate_limited", "message": "...", "retryAfter": N, "lockedUntil": N }
Errors: 400
```

---

## Appendix: Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `invalid_credentials` | 401 | Email or password incorrect |
| `mfa_required` | 401 | MFA verification needed (returned in login response body, not as an error) |
| `mfa_invalid` | 401 | Invalid MFA code |
| `token_expired` | 401 | Reset/refresh token expired |
| `token_invalid` | 400/401 | Invalid token format |
| `email_exists` | 409 | Email already registered |
| `rate_limited` | 200 | Too many attempts (returned in body with retryAfter/lockedUntil) |
| `account_locked` | 423 | Account temporarily locked |

---

*Document Version: 2.0*  
*Status: Aligned with Go server implementation (apps/api/internal/*  
*Architecture: Domain-Driven (apps/api/internal/)*
