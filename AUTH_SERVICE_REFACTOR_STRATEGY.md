# Auth Service Refactoring Strategy

## Overview

The `auth_service.go` file (1594 lines, 52 functions) is monolithic and needs to be split into maintainable pieces for production readiness.

## Current State

**File:** `apps/api/internal/application/auth/auth_service.go`
**Lines:** 1,594
**Functions:** 52
**Dependencies:** 9 (operatorRepo, sessionRepo, emailVerifyRepo, passwordResetRepo, passwordHasher, refreshTokenRepo, jwtManager, sessionManager, sessionTTL, refreshTokenExpiry)

## Target File Structure

```
apps/api/internal/application/auth/
├── auth_helpers.go                     # generateID, hashTokenSha256, etc.
├── auth_types.go                       # Type definitions
├── auth_constructors.go                # NewAuthService, constructors
├── auth_login_session.go               # Login, Register, Session management
├── auth_totp_mfa.go                    # TOTP-based MFA (Google Authenticator)
├── auth_email_verification.go          # Email verification flow
├── auth_password_reset.go              # Password reset with rate limiting
├── auth_google_oauth.go                # Google OAuth integration
├── auth_github_oauth.go                # GitHub OAuth integration
├── auth_operator_settings.go            # Operator preferences/settings
├── auth_operator_admin.go               # Admin operator management
├── auth_refresh_token.go               # JWT refresh token rotation
└── auth_service.go                     # REMOVE (replaced by above)
```

## Issues with Current Structure

1. **Single Responsibility Violation** - One file handles: login, register, MFA, email verification, password reset, OAuth, settings, admin operations, refresh tokens
2. **Code Navigation** - Finding specific functions in 1594 lines is difficult
3. **Testing** - Large files are harder to test in isolation
4. **Code Review** - PRs become unwieldy
5. **Onboarding** - New developers struggle to understand the auth flow

## Proposed Split Strategy

### Phase 1: Extract Helper Functions (Low Risk)

Create `auth_helpers.go` with utility functions used across multiple areas:

```go
// auth_helpers.go
func generateID() string
func generateEmailToken() string
func generateResetToken() string
func hashTokenSha256(token string) string
func hashEmailForTracker(email string) string
func sha256Hash(s string) string
```

### Phase 2: Extract Types (Low Risk)

Create `auth_types.go` with all type definitions:

```go
// auth_types.go
type UpdateOperatorRequest struct { ... }
type MFAEnrollResult struct { ... }
type VerifyEmailResult struct { ... }
type OAuthResult struct { ... }
type ResendRateLimitResult struct { ... }
type RefreshTokenResult struct { ... }
type RefreshTokenRepository interface { ... }
type RefreshToken = refresh_token.RefreshToken
```

### Phase 3: Extract Constructors (Low Risk)

Create `auth_constructors.go`:

```go
// auth_constructors.go
type PasswordPolicy = operator.PasswordPolicy
var DefaultPasswordPolicy = operator.DefaultPasswordPolicy
var ValidatePassword = operator.ValidatePassword
var PasswordStrength = operator.Strength
type PasswordError = operator.PasswordError
func NewAuthService(...) *AuthService
func NewAuthServiceWithRefresh(...) *AuthService
func (s *AuthService) SetJWTManager(...)
func (s *AuthService) SetSessionManager(...)
func (s *AuthService) GetSessionManager() ...
```

### Phase 4: Extract Domain Services (Medium Risk - Requires Interface Alignment)

Before splitting, need to verify/correct domain repository interfaces:

#### auth_login_session.go - Core Authentication
```go
// auth_login_session.go
func (s *AuthService) Login(...)
func (s *AuthService) Register(...)
func (s *AuthService) RegisterAsSuperAdmin(...)
func (s *AuthService) CreateSession(...)
func (s *AuthService) Logout(...)
func (s *AuthService) LogoutAll(...)
func (s *AuthService) GetOperator(...)
func (s *AuthService) GetOperatorByEmail(...)
func (s *AuthService) GetSession(...)
func (s *AuthService) ValidateSession(...)
func (s *AuthService) ChangePassword(...)
```

#### auth_totp_mfa.go - Multi-Factor Authentication
```go
// auth_totp_mfa.go
func (s *AuthService) VerifyMFACode(...)
func (s *AuthService) EnrollMFA(...)
func (s *AuthService) EnableMFA(...)
func (s *AuthService) DisableMFA(...)
func (s *AuthService) GetMFAStatus(...)
func (s *AuthService) RegenerateBackupCodes(...)
func (s *AuthService) VerifyBackupCode(...)
```

#### auth_email_verification.go - Email Verification
```go
// auth_email_verification.go
// Dependencies to verify:
// - email_verification.Repository methods: FindByToken, MarkUsed, DeleteByOperatorID
// - Error types: ErrTokenExpired, ErrTokenUsed, ErrTokenNotReady
func (s *AuthService) VerifyEmail(...)
func (s *AuthService) ResendVerification(...)
func (s *AuthService) CreateEmailVerification(...)
func (s *AuthService) PollVerification(...)
func (s *AuthService) CancelVerification(...)
```

#### auth_password_reset.go - Password Reset
```go
// auth_password_reset.go
// Dependencies to verify:
// - password_reset.Repository methods: FindByTokenHash, Create, GetResendTracker, UpdateResendTracker, DeleteResendTracker, MarkUsed
// - Error types: ErrTrackerNotFound
func (s *AuthService) GeneratePasswordResetToken(...)
func (s *AuthService) ValidatePasswordResetToken(...)
func (s *AuthService) ResetPassword(...)
func (s *AuthService) GetResendTracker(...)
func (s *AuthService) CheckResendRateLimit(...)
func (s *AuthService) UpdateResendTracker(...)
func (s *AuthService) calculateResendCount(...)
func (s *AuthService) DeleteResendTracker(...)
```

#### auth_google_oauth.go - Google OAuth
```go
// auth_google_oauth.go
// Dependencies to verify:
// - operator.AuthProvider enum values
// - operator.NormalizeEmail function
func (s *AuthService) FindOrCreateGoogleOperator(...)
func (s *AuthService) createOAuthOperator(...)
```

#### auth_github_oauth.go - GitHub OAuth
```go
// auth_github_oauth.go
// Dependencies to verify:
// - operator.AuthProvider enum values
// - operator.NormalizeEmail function
func (s *AuthService) FindOrCreateGitHubOperator(...)
func (s *AuthService) createOAuthOperator(...)
```

#### auth_operator_settings.go - Operator Settings
```go
// auth_operator_settings.go
type UpdateSettingsRequest struct { ... }
func (s *AuthService) UpdateOperatorName(...)
func (s *AuthService) UpdateSettings(...)
func (s *AuthService) ResetSettings(...)
```

#### auth_operator_admin.go - Admin Operations
```go
// auth_operator_admin.go
// Dependencies to verify:
// - operator.Repository.ListAll method
// - operator.Role type and IsValid method
// - infraauth.JWTManager.VerifyOperatorToken method
func (s *AuthService) ListAllOperators(...)
func (s *AuthService) CreateOperator(...)
func (s *AuthService) GetOperatorByID(...)
func (s *AuthService) VerifyJWT(...)
func (s *AuthService) UpdateOperator(...)
func (s *AuthService) DeleteOperator(...)
```

#### auth_refresh_token.go - Refresh Token Management
```go
// auth_refresh_token.go
func (s *AuthService) RotateRefreshToken(...)
func (s *AuthService) IssueRefreshToken(...)
func (s *AuthService) RevokeAllRefreshTokens(...)
```

## Domain Interface Verification Checklist

Before Phase 4, verify these domain packages have correct interfaces:

### email_verification.Repository
- [ ] FindByToken(ctx, token) (*EmailVerification, error)
- [ ] Create(ctx, ev *EmailVerification) error
- [ ] MarkUsed(ctx, token) error
- [ ] Delete(ctx, token) error
- [ ] DeleteByOperatorID(ctx, operatorID string) error

### email_verification errors
- [ ] ErrTokenExpired
- [ ] ErrTokenUsed
- [ ] ErrTokenNotReady

### password_reset.Repository
- [ ] FindByTokenHash(ctx, tokenHash string) (*PasswordReset, error)
- [ ] Create(ctx, pr *PasswordReset) error
- [ ] MarkUsed(ctx, token string) error
- [ ] GetResendTracker(ctx, emailHash string) (*ResendTracker, error)
- [ ] UpdateResendTracker(ctx, emailHash string) error
- [ ] DeleteResendTracker(ctx, emailHash string) error

### password_reset errors
- [ ] ErrTrackerNotFound
- [ ] ErrNotFound

### operator.Repository
- [ ] FindByEmail(ctx, email string) (*Operator, error)
- [ ] FindByID(ctx, id string) (*Operator, error)
- [ ] Create(ctx, op *Operator) error
- [ ] Update(ctx, op *Operator) error
- [ ] Delete(ctx, id string) error
- [ ] ListAll(ctx, limit, offset int) ([]*Operator, int, error)

### operator types
- [ ] AuthProvider enum (AuthProviderGoogle, AuthProviderGitHub, etc.)
- [ ] NormalizeEmail(email string) string
- [ ] Role type with IsValid() method

### infraauth.JWTManager
- [ ] VerifyOperatorToken(token string) (*OperatorClaims, error)
- [ ] GenerateOperatorToken(op *Operator) (string, error)

## Implementation Order

1. **auth_helpers.go** - Extract utilities first (no dependencies)
2. **auth_types.go** - Extract types (no dependencies)
3. **auth_constructors.go** - Extract constructors and exports
4. **auth_login_session.go** - Login/Register/Session (mostly standalone)
5. **auth_totp_mfa.go** - TOTP MFA (standalone)
6. **auth_email_verification.go** - After verifying email_verification domain
7. **auth_password_reset.go** - After verifying password_reset domain
8. **auth_google_oauth.go** - Google OAuth (after operator domain verified)
9. **auth_github_oauth.go** - GitHub OAuth (after operator domain verified)
10. **auth_operator_settings.go** - Operator settings (standalone)
11. **auth_operator_admin.go** - Admin operations (after ListAll/JWT verified)
12. **auth_refresh_token.go** - Refresh tokens (standalone)
13. **auth_service.go** - Delete original file

## Testing Strategy

After each phase:
1. Run `go build ./...` to verify compilation
2. Run `go vet ./...` for static analysis
3. Run existing tests to ensure no regressions

## Risk Mitigation

1. **Backup**: Keep original file until all new files compile
2. **Incremental**: Add one file at a time, not everything at once
3. **Verify interfaces**: Check domain interfaces before extracting functions that depend on them
4. **Diff comparison**: Compare compiled output before/after split

## Files to Create

```
apps/api/internal/application/auth/
├── auth_helpers.go                     # Utility functions
├── auth_types.go                       # Type definitions
├── auth_constructors.go                # Constructors and exports
├── auth_login_session.go               # Login, Register, Session management
├── auth_totp_mfa.go                    # TOTP-based MFA (Google Authenticator)
├── auth_email_verification.go          # Email verification flow
├── auth_password_reset.go              # Password reset with rate limiting
├── auth_google_oauth.go                # Google OAuth integration
├── auth_github_oauth.go                # GitHub OAuth integration
├── auth_operator_settings.go            # Operator preferences/settings
├── auth_operator_admin.go               # Admin operator management
├── auth_refresh_token.go               # JWT refresh token rotation
└── auth_service.go                     # REMOVE (replaced by above)
```

## Success Criteria

- [ ] All code compiles without errors
- [ ] All existing tests pass
- [ ] No functionality changes (only file organization)
- [ ] Each file has clear single responsibility
- [ ] Code is more navigable and maintainable
