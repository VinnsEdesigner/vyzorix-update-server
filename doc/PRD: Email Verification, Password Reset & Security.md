# PRD: Email Verification, Password Reset & Security Hardening

## 1. Introduction

This document outlines the implementation plan for securing the Vyzorix Update Server authentication system. The focus is on **email-based verification** for new accounts, **password reset functionality** for account recovery, and **hardened security measures** to protect against common attacks.

### Current State
- ✅ Google OAuth integration (complete with JWKS signature verification)
- ✅ JWT-based session management
- ✅ Password signup with bcrypt hashing
- ❌ No email verification for new accounts
- ❌ No password reset / forgot password flow
- ❌ No rate limiting on auth endpoints
- ❌ No password complexity requirements

### Goals of This Feature
1. Verify email ownership before activating new operator accounts
2. Provide a secure password reset flow for account recovery
3. Prevent brute force and credential stuffing attacks
4. Enforce stronger password requirements

---

## 2. User Stories

### US-001: Email Verification on Signup
**Description:** As a new operator, I want to verify my email address so that the system knows I can receive password reset links and important notifications.

**Acceptance Criteria:**
- [ ] After registration, operator account is created with `email_verified = false`
- [ ] System sends verification email with a unique link containing a secure token
- [ ] Token expires after 24 hours
- [ ] Clicking the link marks the account as verified and redirects to dashboard
- [ ] Expired or invalid tokens show an appropriate error message
- [ ] Verified users can access all features; unverified users have limited access

---

### US-002: Resend Verification Email
**Description:** As a new operator who hasn't verified my email, I want to request a new verification email so that I can complete registration if the first email expired or was lost.

**Acceptance Criteria:**
- [ ] "Resend verification email" link available on login page for unverified accounts
- [ ] New token is generated and old one is invalidated
- [ ] Rate limited to 1 resend per 5 minutes
- [ ] Success message confirms email was sent

---

### US-003: Password Reset Flow
**Description:** As an operator who forgot my password, I want to reset it using my verified email so that I can regain access to my account.

**Acceptance Criteria:**
- [ ] "Forgot password" link available on login page
- [ ] User enters their email address
- [ ] If email exists, system sends password reset email with secure token
- [ ] Token expires after 1 hour
- [ ] User clicks link and is presented with new password form
- [ ] Password must meet complexity requirements
- [ ] On successful reset, all existing sessions are invalidated
- [ ] Confirmation email sent to notify user of password change

---

### US-004: Rate Limiting on Auth Endpoints
**Description:** As a system administrator, I want to limit failed login attempts so that the system is protected against brute force attacks.

**Acceptance Criteria:**
- [ ] Login endpoint limited to 5 attempts per minute per IP address
- [ ] Login endpoint limited to 20 attempts per hour per IP address
- [ ] Registration endpoint limited to 3 attempts per minute per IP
- [ ] Password reset request limited to 3 per hour per email address
- [ ] Rate limit exceeded returns HTTP 429 with retry-after header

---

### US-005: Password Complexity Requirements
**Description:** As a system administrator, I want to enforce strong passwords so that user accounts are less vulnerable to compromise.

**Acceptance Criteria:**
- [ ] Minimum 8 characters (already enforced)
- [ ] At least 1 uppercase letter (A-Z)
- [ ] At least 1 lowercase letter (a-z)
- [ ] At least 1 number (0-9)
- [ ] At least 1 special character (!@#$%^&*()_+-=)
- [ ] Maximum 128 characters
- [ ] Error message clearly states which requirements are missing
- [ ] Passwords checked against HaveIBeenPwned database (optional enhancement)

---

## 3. Architecture Overview

### System Components

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend (React)                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │ Login Page  │  │ Signup Page │  │ Password Reset Flow     │ │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬─────────────┘ │
└─────────┼────────────────┼────────────────────┼────────────────┘
          │                │                    │
          └────────────────┼────────────────────┘
                          │ HTTP/HTTPS
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Backend (Go + Gin)                          │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Auth Controller                        │  │
│  │  ┌──────────┐ ┌──────────┐ ┌───────────┐ ┌───────────┐  │  │
│  │  │ Register  │ │ Login    │ │ Reset Pwd │ │ Verify    │  │  │
│  │  └─────┬────┘ └────┬────┘ └─────┬─────┘ └─────┬─────┘  │  │
│  └────────┼───────────┼───────────┼─────────────┼────────┘  │
│           │           │           │             │            │
│           ▼           ▼           ▼             ▼            │
│  ┌────────────────────────────────────────────────────────┐   │
│  │                   Rate Limiter                          │   │
│  │            (In-memory or Redis-based)                  │   │
│  └────────────────────────────────────────────────────────┘   │
│                              │                                 │
│                              ▼                                 │
│  ┌────────────────────────────────────────────────────────┐   │
│  │                     Store (SQLite)                     │   │
│  │  ┌──────────┐ ┌───────────┐ ┌────────────────────┐   │   │
│  │  │operators │ │email_tokens│ │password_reset_tokens│   │   │
│  │  └──────────┘ └───────────┘ └────────────────────┘   │   │
│  └────────────────────────────────────────────────────────┘   │
│                              │                                 │
│                              ▼                                 │
│  ┌────────────────────────────────────────────────────────┐   │
│  │              Email Service (Resend API)                │   │
│  └────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### Email Flow Diagram

```
SIGNUP FLOW:
─────────────────────────────────────────────────────────────
User submits form → Create operator (unverified) → Generate token 
→ Store token (expires 24h) → Send verification email 
→ User clicks link → Validate token → Mark verified 
→ Redirect to dashboard

PASSWORD RESET FLOW:
─────────────────────────────────────────────────────────────
User clicks "Forgot Password" → Enter email → Check if exists 
→ Generate reset token → Store (expires 1h) → Send reset email 
→ User clicks link → Validate token → Show new password form 
→ Validate password → Update password → Invalidate all sessions 
→ Send confirmation email → Redirect to login
```

---

## 4. Functional Requirements

### 4.1 Database Schema Changes

#### New Table: `email_verifications`
```sql
CREATE TABLE email_verifications (
    id TEXT PRIMARY KEY,
    operator_id TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (operator_id) REFERENCES operators(id) ON DELETE CASCADE
);
CREATE INDEX idx_email_verifications_operator ON email_verifications(operator_id);
CREATE INDEX idx_email_verifications_token ON email_verifications(token_hash);
```

#### New Table: `password_reset_tokens`
```sql
CREATE TABLE password_reset_tokens (
    id TEXT PRIMARY KEY,
    operator_id TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at INTEGER NOT NULL,
    used_at INTEGER,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (operator_id) REFERENCES operators(id) ON DELETE CASCADE
);
CREATE INDEX idx_password_reset_operator ON password_reset_tokens(operator_id);
CREATE INDEX idx_password_reset_token ON password_reset_tokens(token_hash);
```

#### Modified Table: `operators`
```sql
-- Add column (if not exists via migration)
ALTER TABLE operators ADD COLUMN email_verified INTEGER DEFAULT 0;
ALTER TABLE operators ADD COLUMN verification_sent_at INTEGER;
```

### 4.2 New Environment Variables

```bash
# Email Configuration (Resend)
RESEND_API_KEY=re_xxxxxxxxxxxx

# Email Settings
EMAIL_FROM=noreply@vyzorix.example.com
EMAIL_FROM_NAME=Vyzorix Update Server
BASE_URL=https://your-domain.com  # For generating email links

# Security
RATE_LIMIT_WINDOW=60  # seconds
RATE_LIMIT_MAX_REQUESTS=5
PASSWORD_RESET_TOKEN_EXPIRY=3600  # 1 hour in seconds
EMAIL_VERIFY_TOKEN_EXPIRY=86400  # 24 hours in seconds
```

### 4.3 API Endpoints

#### POST /v1/auth/register (Modified)
- **Input:** `{ email, password, name }`
- **Behavior:** Create operator with `email_verified = false`, send verification email
- **Output:** `{ message: "Registration successful. Please check your email to verify your account." }`

#### POST /v1/auth/verify-email
- **Input:** `{ token: string }`
- **Behavior:** Validate token, mark email as verified
- **Output:** `{ verified: true }` or error

#### POST /v1/auth/resend-verification
- **Input:** `{ email: string }`
- **Behavior:** Generate new token, send new verification email (rate limited)
- **Output:** `{ message: "Verification email sent" }`

#### POST /v1/auth/forgot-password
- **Input:** `{ email: string }`
- **Behavior:** Check if email exists, generate reset token, send email
- **Output:** `{ message: "If that email exists, a reset link has been sent" }` (always success for security)

#### POST /v1/auth/reset-password
- **Input:** `{ token: string, new_password: string }`
- **Behavior:** Validate token, check password complexity, update password, invalidate sessions, send confirmation
- **Output:** `{ message: "Password reset successful" }`

### 4.4 Password Complexity Validation

```go
// Password must contain:
// - At least 8 characters
// - At least 1 uppercase letter
// - At least 1 lowercase letter
// - At least 1 digit
// - At least 1 special character (!@#$%^&*()_+-=)
// - Maximum 128 characters
```

### 4.5 Rate Limiting Configuration

| Endpoint | Limit | Window | Key |
|----------|-------|--------|-----|
| POST /v1/auth/login | 5 | 1 minute | IP address |
| POST /v1/auth/login | 20 | 1 hour | IP address |
| POST /v1/auth/register | 3 | 1 minute | IP address |
| POST /v1/auth/forgot-password | 3 | 1 hour | Email |
| POST /v1/auth/resend-verification | 1 | 5 minutes | Email |

---

## 5. File Structure

### New Files to Create

```
vyzorix-update-server/
├── controllers/
│   └── auth.go                    # MODIFIED: Add email verification endpoints
├── models/
│   └── auth.go                    # MODIFIED: Add request/response types for new flows
├── security/
│   ├── password.go                # NEW: Password validation and complexity checking
│   ├── password_test.go           # NEW: Tests for password validation
│   ├── ratelimit.go               # NEW: Rate limiting middleware
│   └── ratelimit_test.go          # NEW: Tests for rate limiting
├── services/
│   └── email.go                   # NEW: Email service using Resend API
│   └── email_test.go              # NEW: Tests for email service
├── storage/
│   ├── sqlite.go                  # MODIFIED: Add verification/reset token methods
│   └── sqlite_test.go             # MODIFIED: Add tests for new methods
├── config/
│   └── config.go                  # MODIFIED: Add email and security config
├── templates/                     # NEW: Email HTML templates
│   ├── verification.html
│   └── password-reset.html
└── .env.example                   # MODIFIED: Add new environment variables
```

### Existing Files to Modify

| File | Changes |
|------|---------|
| `controllers/auth.go` | Add endpoints: verify-email, resend-verification, forgot-password, reset-password |
| `models/auth.go` | Add request/response types for new endpoints |
| `storage/sqlite.go` | Add methods for email tokens and password reset tokens |
| `config/config.go` | Add email configuration fields |
| `middleware/auth.go` | Add email verification check middleware |
| `routes/` or `controllers/server.go` | Register new auth routes |

---

## 6. Detailed Implementation Guide

### 6.1 Password Validation (`security/password.go`)

```go
package security

import (
    "regexp"
    "unicode"
)

type PasswordPolicy struct {
    MinLength     int
    MaxLength     int
    RequireUpper  bool
    RequireLower  bool
    RequireDigit  bool
    RequireSpecial bool
}

var DefaultPolicy = PasswordPolicy{
    MinLength:      8,
    MaxLength:      128,
    RequireUpper:   true,
    RequireLower:   true,
    RequireDigit:   true,
    RequireSpecial: true,
}

type PasswordError struct {
    Missing []string
}

func (e *PasswordError) Error() string {
    return "password does not meet requirements: " + join(e.Missing, ", ")
}

func ValidatePassword(password string, policy PasswordPolicy) error {
    var missing []string
    
    if len(password) < policy.MinLength {
        missing = append(missing, "minimum 8 characters")
    }
    if len(password) > policy.MaxLength {
        missing = append(missing, "maximum 128 characters")
    }
    if policy.RequireUpper && !containsUpper(password) {
        missing = append(missing, "at least 1 uppercase letter")
    }
    if policy.RequireLower && !containsLower(password) {
        missing = append(missing, "at least 1 lowercase letter")
    }
    if policy.RequireDigit && !containsDigit(password) {
        missing = append(missing, "at least 1 number")
    }
    if policy.RequireSpecial && !containsSpecial(password) {
        missing = append(missing, "at least 1 special character (!@#$%^&*()_+-=)")
    }
    
    if len(missing) > 0 {
        return &PasswordError{Missing: missing}
    }
    return nil
}

// Helper functions...
```

### 6.2 Rate Limiter (`security/ratelimit.go`)

```go
package security

import (
    "net/http"
    "sync"
    "time"
)

type RateLimiter struct {
    requests map[string]*window
    mu       sync.RWMutex
    window   time.Duration
    max      int
}

type window struct {
    count     int
    resetTime time.Time
}

func NewRateLimiter(window time.Duration, max int) *RateLimiter {
    limiter := &RateLimiter{
        requests: make(map[string]*window),
        window:   window,
        max:      max,
    }
    // Cleanup goroutine
    go limiter.cleanup()
    return limiter
}

func (rl *RateLimiter) Allow(key string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    now := time.Now()
    w, exists := rl.requests[key]
    
    if !exists || now.After(w.resetTime) {
        rl.requests[key] = &window{
            count:     1,
            resetTime: now.Add(rl.window),
        }
        return true
    }
    
    if w.count >= rl.max {
        return false
    }
    
    w.count++
    return true
}

func (rl *RateLimiter) cleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        rl.mu.Lock()
        now := time.Now()
        for key, w := range rl.requests {
            if now.After(w.resetTime) {
                delete(rl.requests, key)
            }
        }
        rl.mu.Unlock()
    }
}
```

### 6.3 Email Service (`services/email.go`)

```go
package services

import (
    "context"
    "fmt"
    "os"
    "text/template"
)

type EmailService struct {
    apiKey    string
    fromEmail string
    fromName  string
    baseURL   string
}

func NewEmailService() *EmailService {
    return &EmailService{
        apiKey:    os.Getenv("RESEND_API_KEY"),
        fromEmail: os.Getenv("EMAIL_FROM"),
        fromName:  os.Getenv("EMAIL_FROM_NAME"),
        baseURL:   os.Getenv("BASE_URL"),
    }
}

type EmailData struct {
    To      string
    Subject string
    HTML    string
}

func (es *EmailService) SendVerificationEmail(ctx context.Context, email, name, token string) error {
    verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", es.baseURL, token)
    
    html, err := es.parseTemplate("templates/verification.html", map[string]any{
        "name":      name,
        "verifyURL": verifyURL,
        "expiryHours": 24,
    })
    if err != nil {
        return err
    }
    
    return es.send(ctx, email, "Verify your Vyzorix account", html)
}

func (es *EmailService) SendPasswordResetEmail(ctx context.Context, email, name, token string) error {
    resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", es.baseURL, token)
    
    html, err := es.parseTemplate("templates/password-reset.html", map[string]any{
        "name":     name,
        "resetURL": resetURL,
        "expiryMinutes": 60,
    })
    if err != nil {
        return err
    }
    
    return es.send(ctx, email, "Reset your Vyzorix password", html)
}

func (es *EmailService) send(ctx context.Context, to, subject, html string) error {
    // Use Resend API
    payload := map[string]any{
        "from":    fmt.Sprintf("%s <%s>", es.fromName, es.fromEmail),
        "to":      []string{to},
        "subject": subject,
        "html":    html,
    }
    
    // HTTP POST to Resend API...
    return nil
}
```

### 6.4 Database Methods (`storage/sqlite.go`)

```go
// Email Verification Tokens

func (s *Store) CreateEmailVerification(ctx context.Context, token *EmailVerification) error {
    _, err := s.db.ExecContext(ctx,
        `INSERT INTO email_verifications(id, operator_id, token_hash, expires_at, created_at)
         VALUES(?, ?, ?, ?, ?)`,
        token.ID, token.OperatorID, token.TokenHash, token.ExpiresAt.UnixMilli(), token.CreatedAt.UnixMilli(),
    )
    return err
}

func (s *Store) GetEmailVerificationByToken(ctx context.Context, tokenHash string) (*EmailVerification, error) {
    var v EmailVerification
    err := s.db.QueryRowContext(ctx,
        `SELECT id, operator_id, token_hash, expires_at, created_at
         FROM email_verifications WHERE token_hash = ?`,
        tokenHash,
    ).Scan(&v.ID, &v.OperatorID, &v.TokenHash, &v.ExpiresAt, &v.CreatedAt)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    }
    return &v, err
}

func (s *Store) DeleteEmailVerification(ctx context.Context, id string) error {
    _, err := s.db.ExecContext(ctx, `DELETE FROM email_verifications WHERE id = ?`, id)
    return err
}

func (s *Store) DeleteEmailVerificationsByOperator(ctx context.Context, operatorID string) error {
    _, err := s.db.ExecContext(ctx, `DELETE FROM email_verifications WHERE operator_id = ?`, operatorID)
    return err
}

// Password Reset Tokens (similar methods)
```

---

## 7. Email Templates

### verification.html

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .button { display: inline-block; background: #0066cc; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; }
        .footer { color: #666; font-size: 12px; margin-top: 30px; }
    </style>
</head>
<body>
    <div class="container">
        <h2>Verify your email address</h2>
        <p>Hi {{.name}},</p>
        <p>Thanks for signing up! Please verify your email address by clicking the button below:</p>
        <p style="text-align: center;">
            <a href="{{.verifyURL}}" class="button">Verify Email Address</a>
        </p>
        <p>This link expires in {{.expiryHours}} hours.</p>
        <p>If you didn't create an account with Vyzorix, you can safely ignore this email.</p>
        <div class="footer">
            <p>Vyzorix Update Server<br>
            This email was sent automatically. Please do not reply.</p>
        </div>
    </div>
</body>
</html>
```

### password-reset.html

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .button { display: inline-block; background: #cc3300; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; }
        .warning { background: #fff3cd; padding: 15px; border-radius: 4px; margin: 20px 0; }
        .footer { color: #666; font-size: 12px; margin-top: 30px; }
    </style>
</head>
<body>
    <div class="container">
        <h2>Reset your password</h2>
        <p>Hi {{.name}},</p>
        <p>We received a request to reset your password. Click the button below to create a new password:</p>
        <p style="text-align: center;">
            <a href="{{.resetURL}}" class="button">Reset Password</a>
        </p>
        <div class="warning">
            <strong>⚠️ Security Notice:</strong> This link expires in {{.expiryMinutes}} minutes and can only be used once.
        </div>
        <p>If you didn't request a password reset, please ignore this email. Your password won't be changed.</p>
        <div class="footer">
            <p>Vyzorix Update Server<br>
            If you need help, contact support.</p>
        </div>
    </div>
</body>
</html>
```

---

## 8. Frontend Changes

### Login Page (`src/routes/login.tsx`)
- Add "Forgot Password?" link below password field
- Show message for unverified accounts: "Please check your email and verify your account before logging in"

### Signup Page (new or integrated)
- Show success message after registration: "Check your email to verify your account"
- Link to resend verification email

### Forgot Password Page (new route)
- Simple form: email input + "Send Reset Link" button
- Success state: "If that email exists, we've sent a password reset link"

### Reset Password Page (new route)
- Form: new password + confirm password
- Password requirements displayed
- Success: redirect to login with success message

### Auth Callback Page (`src/routes/auth.callback.tsx`)
- Handle email verification token (similar to login token handling)
- Show success or error state

---

## 9. Security Considerations

### 9.1 Token Generation
- Use cryptographically secure random bytes (16+ bytes)
- Hash tokens before storage (SHA-256)
- Display partial token to user (last 4 chars) for verification

### 9.2 Timing Attacks
- Use constant-time comparison for token validation
- Return same error message whether email exists or not (for forgot password)

### 9.3 Email Address Enumeration
- For forgot password: always show "If that email exists, a link has been sent"
- For resend verification: same approach

### 9.4 Session Invalidation
- On password reset, delete all auth_sessions for that operator
- On email verification, no session invalidation needed

### 9.5 Logging
- Log email verification requests (without token)
- Log password reset requests (without token)
- Log failed verification attempts
- Alert on suspicious patterns (multiple requests from same IP)

---

## 10. Configuration Reference

### Environment Variables (.env)

```bash
# Server
PORT=3000
BASE_URL=https://vyzorix.example.com

# Database
DATABASE_URL=./data/vyzorix.db

# JWT
JWT_SECRET=your-secret-key-min-32-chars
JWT_DURATION=24h

# Google OAuth
GOOGLE_OAUTH_CLIENT_ID=your-client-id
GOOGLE_OAUTH_CLIENT_SECRET=your-client-secret

# Email (Resend)
RESEND_API_KEY=re_xxxxxxxxxxxx
EMAIL_FROM=noreply@vyzorix.example.com
EMAIL_FROM_NAME=Vyzorix Update Server

# Security
RATE_LIMIT_WINDOW_SECONDS=60
RATE_LIMIT_MAX_REQUESTS=5
PASSWORD_RESET_TOKEN_EXPIRY_SECONDS=3600
EMAIL_VERIFY_TOKEN_EXPIRY_SECONDS=86400

# Frontend
FRONTEND_URL=https://vyzorix.example.com
```

---

## 11. Testing Requirements

### Unit Tests
- [ ] Password validation: all complexity rules
- [ ] Password validation: edge cases (empty, Unicode, very long)
- [ ] Rate limiter: allow/deny logic
- [ ] Token generation: uniqueness, length
- [ ] Email service: template rendering

### Integration Tests
- [ ] Registration flow: create operator, send email, verify token
- [ ] Resend verification: rate limiting, token regeneration
- [ ] Forgot password: token generation, email sent
- [ ] Reset password: token validation, password update, session invalidation
- [ ] Rate limiting: 429 response after exceeded

### Security Tests
- [ ] Token cannot be reused after verification
- [ ] Token cannot be reused after password reset
- [ ] Expired tokens are rejected
- [ ] Invalid tokens return appropriate error
- [ ] Password complexity enforced on reset

---

## 12. Success Metrics

| Metric | Target |
|--------|--------|
| Email verification rate | > 80% within 24 hours |
| Password reset completion rate | > 70% of reset requests |
| Failed login attempts blocked | 100% of rate-limited attempts |
| Passwords meeting complexity | > 90% of new passwords |

---

## 13. Future Enhancements (Out of Scope)

- [ ] WhatsApp/SMS verification (requires Twilio integration)
- [ ] Two-factor authentication (TOTP)
- [ ] Password breach checking (HaveIBeenPwned API)
- [ ] Login history and device tracking
- [ ] Account lockout after multiple failures
- [ ] Login notifications email

---

## 14. Open Questions

1. Should unverified accounts be allowed to login at all, or only after verification?
2. Should there be an account deletion policy for unverified accounts (e.g., delete after 30 days)?
3. Should password reset tokens be single-use or allow multiple uses until expiry?
4. Should we track verification email delivery status and retry failed deliveries?
