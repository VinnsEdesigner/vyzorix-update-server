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

> ⚠️ **Architecture Alignment Note**
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
├── api/
│   ├── handlers/auth/           # EXISTING
│   │   ├── auth_login.go        # EXISTS
│   │   ├── auth_register.go    # EXISTS
│   │   ├── auth_logout.go      # EXISTS
│   │   ├── auth_me.go          # EXISTS
│   │   ├── auth_mfa.go         # EXISTS
│   │   ├── auth_oauth.go       # EXISTS
│   │   ├── auth_password_reset.go    # EXISTS
│   │   ├── auth_email_verify.go      # EXISTS
│   │   ├── auth_settings.go    # EXISTS
│   │   ├── auth_admin.go       # EXISTS
│   │   ├── auth_routes.go      # EXISTS
│   │   └── auth_lockout.go     # EXISTS
│   └── middleware/
├── application/auth/           # EXISTS
│   ├── auth_service.go        # EXISTS (AuthService)
│   └── auth_password.go       # EXISTS
├── domain/
│   ├── operator/               # EXISTS
│   │   ├── operator_entity.go           # EXISTS (Operator struct)
│   │   ├── operator_repository.go       # EXISTS
│   │   ├── operator_errors.go           # EXISTS
│   │   ├── operator_password.go         # EXISTS
│   │   ├── operator_role.go            # EXISTS
│   │   ├── operator_requests.go         # EXISTS
│   │   └── operator_responses.go       # EXISTS
│   ├── session/                # EXISTS
│   │   ├── session_entity.go          # EXISTS
│   │   └── session_repository.go      # EXISTS
│   ├── refresh_token/          # EXISTS
│   │   ├── refresh_token_entity.go          # EXISTS
│   │   └── refresh_token_repository.go      # EXISTS
│   ├── email_verification/     # EXISTS
│   │   ├── email_verification_entity.go          # EXISTS
│   │   ├── email_verification_repository.go      # EXISTS
│   │   ├── email_verification_requests.go       # EXISTS
│   │   └── email_verification_responses.go      # EXISTS
│   ├── password_reset/         # EXISTS
│   │   ├── password_reset_entity.go          # EXISTS
│   │   ├── password_reset_repository.go      # EXISTS
│   │   ├── password_reset_requests.go        # EXISTS
│   │   └── password_reset_responses.go       # EXISTS
│   └── oauth/                  # EXISTS
│       ├── oauth_entity.go          # EXISTS
│       └── oauth_errors.go         # EXISTS
└── infrastructure/
    ├── security/              # EXISTS
    ├── email/                  # EXISTS
    └── storage/                # EXISTS
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
| POST | `/v1/auth/login` | None | Credential login | **EXISTS** |
| POST | `/v1/auth/register` | None | Register new operator | **EXISTS** |
| POST | `/v1/auth/logout` | Cookie | Logout current session | **EXISTS** |
| GET | `/v1/auth/me` | Cookie | Get current operator | **EXISTS** |
| PATCH | `/v1/auth/me` | Cookie | Update operator name | **EXISTS** |
| POST | `/v1/auth/refresh` | None | Refresh access token | **EXISTS** |

### 3.2 MFA Endpoints

| Method | Endpoint | Auth | Description | Status |
|--------|----------|------|-------------|--------|
| GET | `/v1/auth/mfa/status` | Cookie | Get MFA enrollment status | **EXISTS** |
| POST | `/v1/auth/mfa/enroll` | Cookie | Start MFA enrollment | **EXISTS** |
| POST | `/v1/auth/mfa/verify-setup` | Cookie | Verify MFA setup | **EXISTS** |
| POST | `/v1/auth/mfa/enable` | Cookie | Enable MFA | **EXISTS** |
| POST | `/v1/auth/mfa/disable` | Cookie | Disable MFA | **EXISTS** |
| POST | `/v1/auth/mfa/verify-backup` | Cookie | Verify backup code | **EXISTS** |
| POST | `/v1/auth/mfa/regenerate-backup-codes` | Cookie | Regenerate backup codes | **EXISTS** |
| POST | `/v1/auth/mfa/verify` | None | Verify MFA code (post-login) | **MISSING** |

### 3.3 Password Reset Endpoints

| Method | Endpoint | Auth | Description | Status |
|--------|----------|------|-------------|--------|
| POST | `/v1/auth/forgot-password` | None | Request password reset | **EXISTS** |
| POST | `/v1/auth/reset-password` | None | Reset with token | **EXISTS** |
| POST | `/v1/auth/resend-password-reset` | None | Resend reset email | **EXISTS** |

### 3.4 Email Verification Endpoints

| Method | Endpoint | Auth | Description | Status |
|--------|----------|------|-------------|--------|
| POST | `/v1/auth/verify-email` | None | Verify email token | **EXISTS** |
| POST | `/v1/auth/resend-verification` | None | Resend verification | **EXISTS** |
| POST | `/v1/auth/cancel-verification` | None | Cancel verification | **EXISTS** |
| GET | `/v1/auth/poll-verification` | None | Poll verification status | **EXISTS** |

### 3.5 OAuth Endpoints

| Method | Endpoint | Auth | Description | Status |
|--------|----------|------|-------------|--------|
| GET | `/v1/auth/google` | None | Initiate Google OAuth | **EXISTS** |
| GET | `/v1/auth/google/callback` | None | Google OAuth callback | **EXISTS** |
| GET | `/v1/auth/github` | None | Initiate GitHub OAuth | **EXISTS** |
| GET | `/v1/auth/github/callback` | None | GitHub OAuth callback | **EXISTS** |

### 3.6 Admin Endpoints

| Method | Endpoint | Auth | Description | Status |
|--------|----------|------|-------------|--------|
| GET | `/v1/auth/admin/operators` | Admin | List all operators | **EXISTS** |
| POST | `/v1/auth/admin/operators` | Admin | Create operator | **EXISTS** |
| GET | `/v1/auth/admin/operators/:id` | Admin | Get operator | **EXISTS** |
| PATCH | `/v1/auth/admin/operators/:id` | Admin | Update operator | **EXISTS** |
| DELETE | `/v1/auth/admin/operators/:id` | Admin | Delete operator | **EXISTS** |
| POST | `/v1/auth/admin/lockout/unlock/:operator_id` | Admin | Unlock account | **EXISTS** |

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

### 5.2 Required Application DTOs

```go
// application/dto/auth.go (EXISTING + NEW)

package dto

// EXISTING
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginResponse struct {
    OperatorID string `json:"operator_id"`
    Email      string `json:"email"`
    Name       string `json:"name"`
    Role       string `json:"role"`
    MFAEnabled bool   `json:"mfa_enabled"`
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

// NEW - MFA Verification (post-login)
type MFAVerifyRequest struct {
    OperatorID string `json:"operator_id"`
    Code       string `json:"code"`
}

type MFAVerifyResponse struct {
    Success bool   `json:"success"`
    Token   string `json:"token,omitempty"`
    Session string `json:"session,omitempty"`
}

// NEW - Token Refresh
type RefreshTokenRequest struct {
    RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresAt    int64  `json:"expires_at"`
    SessionID    string `json:"session_id"`
}
```

---

## 6. Handler Specifications

### 6.1 Login Handler (EXISTING)

```go
// handlers/auth/login.go

type LoginHandler struct {
    authService *auth.AuthService
    presenter  *response.Presenter
}

func (h *LoginHandler) Handle(c *gin.Context) {
    // 1. Parse LoginRequest
    // 2. Call authService.Login()
    // 3. If ErrMFARequired, return with mfa_required=true
    // 4. Set session cookie
    // 5. Return LoginResponse
}
```

### 6.2 MFA Verification Handler (NEW)

```go
// handlers/auth/mfa_verify.go (NEW)

type MFAVerifyHandler struct {
    authService *auth.AuthService
    presenter  *response.Presenter
}

// POST /v1/auth/mfa/verify
func (h *MFAVerifyHandler) Handle(c *gin.Context) {
    // 1. Parse MFAVerifyRequest { operator_id, code }
    // 2. Validate code format (6 digits)
    // 3. Call authService.VerifyMFACode()
    // 4. If valid, create session
    // 5. Set session cookie
    // 6. Return MFAVerifyResponse with tokens
}
```

### 6.3 Refresh Token Handler (NEW)

```go
// handlers/auth/refresh.go (NEW)

type RefreshHandler struct {
    authService *auth.AuthService
    presenter  *response.Presenter
}

// POST /v1/auth/refresh
func (h *RefreshHandler) Handle(c *gin.Context) {
    // 1. Parse RefreshTokenRequest
    // 2. Validate refresh token
    // 3. Call authService.RotateRefreshToken()
    // 4. Implement refresh token rotation (revoke old, issue new)
    // 5. Return new tokens
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
├── api/
│   ├── handlers/auth/
│   │   ├── login.go            # EXISTS
│   │   ├── register.go         # EXISTS
│   │   ├── logout.go           # EXISTS
│   │   ├── me.go               # EXISTS
│   │   ├── mfa.go              # EXISTS
│   │   ├── oauth.go             # EXISTS
│   │   ├── password_reset.go    # EXISTS
│   │   ├── email_verify.go      # EXISTS
│   │   ├── settings.go          # EXISTS
│   │   ├── admin.go             # EXISTS
│   │   ├── lockout.go           # EXISTS
│   │   ├── routes.go            # EXISTS
│   │   └── client_credentials.go # EXISTS
│   │
│   ├── middleware/
│   │   ├── auth.go              # EXISTS
│   │   ├── cookie_auth.go       # EXISTS
│   │   ├── lockout.go           # EXISTS
│   │   ├── rate_limit.go       # EXISTS
│   │   └── validation.go        # EXISTS
│   │
│   └── responses/
│       └── presenter.go         # EXISTS
│
├── application/auth/
│   ├── auth_service.go              # EXISTS
│   └── auth_password.go              # EXISTS
│
├── domain/
│   ├── operator/
│   │   ├── operator_entity.go           # EXISTS
│   │   ├── operator_repository.go       # EXISTS
│   │   ├── operator_errors.go           # EXISTS
│   │   ├── operator_password.go         # EXISTS
│   │   ├── operator_role.go            # EXISTS
│   │   ├── operator_requests.go       # EXISTS
│   │   ├── operator_responses.go      # EXISTS
│   │   └── operator_email.go          # EXISTS
│   │
│   ├── session/
│   │   ├── session_entity.go          # EXISTS
│   │   └── session_repository.go      # EXISTS
│   │
│   ├── refresh_token/
│   │   ├── refresh_token_entity.go          # EXISTS
│   │   └── refresh_token_repository.go      # EXISTS
│   │
│   ├── email_verification/
│   │   ├── email_verification_entity.go          # EXISTS
│   │   ├── email_verification_repository.go      # EXISTS
│   │   ├── email_verification_requests.go        # EXISTS
│   │   └── email_verification_responses.go     # EXISTS
│   │
│   ├── password_reset/
│   │   ├── password_reset_entity.go          # EXISTS
│   │   ├── password_reset_repository.go      # EXISTS
│   │   ├── password_reset_requests.go        # EXISTS
│   │   └── password_reset_responses.go      # EXISTS
│   │
│   └── oauth/
│       ├── oauth_entity.go          # EXISTS
│       └── oauth_errors.go         # EXISTS
│
└── infrastructure/
    ├── security/
    │   ├── jwt.go             # EXISTS
    │   ├── password.go        # EXISTS
    │   ├── google.go          # EXISTS
    │   └── session/           # EXISTS
    │
    ├── email/
    │   └── email_service.go   # EXISTS
    │
    └── storage/
        ├── auth_storage.go           # EXISTS
        ├── session_storage.go       # EXISTS
        └── operator_storage.go      # EXISTS
```

### 9.2 Files (Actual Implementation)

| Category | File | Handler/Service |
|----------|------|----------------|
| Handler | `internal/api/handlers/auth/auth_mfa.go` | AuthMfaHandler |
| Handler | `internal/api/handlers/auth/auth_refresh.go` | RefreshHandler |
| Handler | `internal/api/handlers/auth/auth_routes.go` | Route registration |
| Service | `internal/application/auth/auth_service.go` | AuthService |

---

## 10. Implementation Order

### Phase 1: Verify Existing Implementation (Day 1)
1. Audit existing auth handlers for completeness
2. Verify MFA enrollment flow
3. Verify OAuth flows (Google, GitHub)
4. Verify password reset flow
5. Identify gaps

### Phase 2: Fill Gaps (Day 1-2)
1. Create `handlers/auth/mfa_verify.go` if needed
2. Create `handlers/auth/refresh.go` if needed
3. Update routes.go with new endpoints
4. Add missing DTOs

### Phase 3: Service Layer (Day 2)
1. Verify `RotateRefreshToken` implementation
2. Add any missing service methods
3. Add token validation helpers

### Phase 4: Testing (Day 2-3)
1. Unit tests for handlers
2. Integration tests for auth flows
3. E2E tests for complete flows:
   - Login → MFA → Session
   - OAuth → Callback → Session
   - Password Reset → New Password → Login
   - Token Refresh → New Tokens

### Phase 5: Documentation (Day 3)
1. Update API documentation
2. Document auth flows
3. Add security considerations

---

## Appendix: API Contract Reference

### Login
```
POST /v1/auth/login
Request: { "email": "...", "password": "..." }
Response: { "operator_id": "...", "email": "...", "name": "...", "role": "...", "mfa_enabled": true/false }
Errors: 400, 401, 429
```

### MFA Verification
```
POST /v1/auth/mfa/verify
Request: { "operator_id": "...", "code": "123456" }
Response: { "success": true, "token": "...", "session": "..." }
Errors: 400, 401, 429
```

### Refresh Token
```
POST /v1/auth/refresh
Request: { "refresh_token": "..." }
Response: { "access_token": "...", "refresh_token": "...", "expires_at": 1234567890, "session_id": "..." }
Errors: 400, 401, 429
```

### Get Current User
```
GET /v1/auth/me
Response: { "id": "...", "email": "...", "name": "...", "role": "...", "mfa_enabled": true, "email_verified": true }
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
Response: { "success": true }
Errors: 400, 429
```

### Reset Password
```
POST /v1/auth/reset-password
Request: { "token": "...", "new_password": "..." }
Response: { "success": true }
Errors: 400, 410 (expired)
```

---

## Appendix: Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `invalid_credentials` | 401 | Email or password incorrect |
| `mfa_required` | 403 | MFA verification needed |
| `mfa_invalid` | 401 | Invalid MFA code |
| `token_expired` | 410 | Reset/token expired |
| `token_invalid` | 400 | Invalid token format |
| `email_exists` | 409 | Email already registered |
| `rate_limited` | 429 | Too many attempts |
| `account_locked` | 423 | Account temporarily locked |

---

*Document Version: 1.0*  
*Status: Aligned with existing server structure*  
*Architecture: Domain-Driven (apps/api/internal/)*
