# Sign-Up Integration Plan: Library → vyzorix-update-server

> **Document Version:** 1.0  
> **Date:** 2026-06-13  
> **Purpose:** End-to-end integration of Library's auth components into vyzorix-update-server

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [4 Routes to Create](#2-4-routes-to-create)
3. [Data Flow Diagrams](#3-data-flow-diagrams)
4. [File Modifications](#4-file-modifications)
5. [API Contract Mapping](#5-api-contract-mapping)
6. [Component Integration](#6-component-integration)
7. [SSR Hydration Flow](#7-ssr-hydration-flow)
8. [Auth Layout with Wolf Background](#8-auth-layout-with-wolf-background)
9. [Missing Go Endpoints](#9-missing-go-endpoints)
10. [Implementation Order](#10-implementation-order)

---

## 1. Architecture Overview

### Current State

```
vyzorix-update-server (TanStack Start)
┌─────────────────────────────────────────────────────────┐
│  routes/                                                │
│  ├── login.tsx              ← Basic card UI           │
│  ├── forgot-password.tsx     ← Basic form               │
│  └── verify-email.tsx      ← Basic form               │
│                                                         │
│  lib/                                                │
│  ├── vyzorix-auth.ts       ← localStorage + JWT      │
│  └── vyzorix-config.tsx    ← serverUrl config        │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│  Go Backend (apps/api/)                               │
│  ├── internal/api/handlers/auth.go   ← JWT validation │
│  └── pkg/storage/sqlite.go           ← Database       │
└─────────────────────────────────────────────────────────┘
```

### Target State

```
vyzorix-update-server (TanStack Start)
┌─────────────────────────────────────────────────────────┐
│  routes/                                                │
│  ├── login.tsx              ← Library's LoginForm     │
│  ├── signup.tsx             ← Library's SignUpForm    │
│  ├── forgot-password.tsx    ← Library's ForgotForm    │
│  └── verify-email.tsx      ← Library's WaitingVerif  │
│                                                         │
│  components/auth/            ← COPY from Library      │
│  ├── LoginForm.tsx                                  │
│  ├── SignUpForm.tsx                                 │
│  ├── ForgotPasswordForm.tsx                          │
│  ├── WaitingVerification.tsx                          │
│  ├── SuccessView.tsx                                │
│  └── SpinningBlocksLoader.tsx                        │
│                                                         │
│  lib/clients/               ← 3 AUTH CLIENTS          │
│  ├── auth.ts                ← register, login, etc.  │
│  ├── sso.ts                 ← Google OAuth init    │
│  └── verification.ts         ← Poll, resend          │
│                                                         │
│  lib/server/                ← SSR (DONE)             │
│  ├── cookie-reader.ts                               │
│  └── state-injector.tsx                             │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│  Go Backend (apps/api/) - COOKIE AUTH (DONE)          │
│  ├── SessionManager        ← AES-256-GCM cookies     │
│  ├── CookieAuth            ← Middleware              │
│  └── handlers/auth.go      ← Sets/clears cookies     │
└─────────────────────────────────────────────────────────┘
```

---

## 2. 4 Routes to Create

| Route | Component | Parent Route | Purpose |
|-------|-----------|--------------|---------|
| `/login` | `LoginForm.tsx` | None | Email/password + SSO login |
| `/signup` | `SignUpForm.tsx` | None | Registration with email verification |
| `/forgot-password` | `ForgotPasswordForm.tsx` | None | Request password reset email |
| `/verify-email` | `WaitingVerification.tsx` | None | Poll for email verification |

---

## 3. Data Flow Diagrams

### 3.1 Sign-Up Flow

```
┌──────────────────────────────────────────────────────────────────────────┐
│ USER ACTION: Submit SignUpForm                                            │
└────────────────────────────┬───────────────────────────────────────────────┘
                             ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ /signup.tsx                                                             │
│ Route component:                                                          │
│   const handleSignUp = async (data) => {                               │
│     await register({ email, password, name });  ← auth.register()     │
│     navigate('/verify-email');                                           │
│   }                                                                     │
└────────────────────────────┬───────────────────────────────────────────────┘
                             ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ /lib/clients/auth.ts                                                     │
│ export const register = async (payload) => {                            │
│   const res = await fetch(`${API}/register`, {                         │
│     method: 'POST',                                                    │
│     headers: { 'Content-Type': 'application/json' },                   │
│     credentials: 'include',        ← NEW: Send cookies                 │
│     body: JSON.stringify(payload)                                       │
│   });                                                                   │
│   if (!res.ok) throw new Error(...);                                   │
│   return res.json();                   ← { message: "..." }             │
│ }                                                                       │
└────────────────────────────┬───────────────────────────────────────────────┘
                             ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ Go Backend: POST /v1/auth/register                                       │
│                                                                          │
│ auth.go → Register():                                                    │
│   1. Validate email, password, name                                      │
│   2. bcrypt hash password                                               │
│   3. Create operator in SQLite                                          │
│   4. Send verification email                                             │
│   5. Return: { message: "Check your email..." }                         │
└────────────────────────────┬───────────────────────────────────────────────┘
                             ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ /verify-email.tsx                                                        │
│ Shows: "Check your email" + poll /resend button                        │
│                                                                          │
│ Polls: GET /v1/auth/poll-verification?token=X                          │
│        (Need to ADD this endpoint to Go backend)                        │
└──────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Login Flow

```
┌──────────────────────────────────────────────────────────────────────────┐
│ USER ACTION: Submit LoginForm                                             │
└────────────────────────────┬───────────────────────────────────────────────┘
                             ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ /lib/clients/auth.ts                                                     │
│ export const login = async (ident, pass) => {                           │
│   const res = await fetch(`${API}/login`, {                            │
│     method: 'POST',                                                     │
│     headers: { 'Content-Type': 'application/json' },                    │
│     credentials: 'include',           ← Cookie will be SET by backend   │
│     body: JSON.stringify({ email: ident, password: pass })              │
│   });                                                                    │
│   if (!res.ok) throw new Error(...);                                   │
│   return res.json();                  ← Operator object                  │
│ }                                                                        │
└────────────────────────────┬───────────────────────────────────────────────┘
                             ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ Go Backend: POST /v1/auth/login                                         │
│                                                                          │
│ auth.go → Login():                                                       │
│   1. Find operator by email                                              │
│   2. bcrypt compare password                                             │
│   3. CreateSessionCookie(operatorID)  ← AES-256-GCM encrypted           │
│   4. http.SetCookie(writer, cookie)  ← Set-Cookie: vyz_session=...     │
│   5. Return: operator JSON              (NO TOKEN IN RESPONSE)          │
└────────────────────────────┬───────────────────────────────────────────────┘
                             ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ /login.tsx                                                              │
│   const operator = await login(...);                                    │
│   // Cookie is now set in browser automatically                          │
│   navigate('/dashboard');                                                │
└──────────────────────────────────────────────────────────────────────────┘
```

### 3.3 SSO (Google) Flow

```
┌──────────────────────────────────────────────────────────────────────────┐
│ USER ACTION: Click "Login with Google"                                   │
└────────────────────────────┬───────────────────────────────────────────────┘
                             ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ /lib/clients/sso.ts                                                     │
│ export const initiateGoogleOAuth = () => {                              │
│   window.location.href = `${API}/google?state=${callbackPath}`;        │
│ }                                                                        │
└────────────────────────────┬───────────────────────────────────────────────┘
                             ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ Go Backend: GET /v1/auth/google                                        │
│                                                                          │
│ auth.go → GoogleLoginRedirect():                                         │
│   Redirects to Google OAuth consent screen                              │
└────────────────────────────┬───────────────────────────────────────────────┘
                             ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ Google OAuth Consent → Callback                                           │
└────────────────────────────┬───────────────────────────────────────────────┘
                             ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ Go Backend: GET /v1/auth/google/callback                                │
│                                                                          │
│ auth.go → GoogleCallback():                                               │
│   1. Exchange code for tokens                                           │
│   2. Verify ID token                                                    │
│   3. Find/create operator                                               │
│   4. CreateSessionCookie(operatorID)   ← Set cookie                     │
│   5. Redirect: /dashboard?oauth=success&new=true                       │
│      (NO TOKEN IN URL ANYMORE - cookie is set!)                         │
└────────────────────────────┬───────────────────────────────────────────────┘
                             ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ /dashboard.tsx                                                           │
│   Checks cookie via SSR or /me endpoint                                  │
│   Shows authenticated dashboard                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## 4. File Modifications

### 4.1 CREATE: Route Files

| File | Purpose | Source |
|------|---------|--------|
| `apps/web/src/routes/signup.tsx` | Sign-up page wrapper | Create new |
| `apps/web/src/routes/login.tsx` | Login page wrapper (UPDATE existing) | Create new |
| `apps/web/src/routes/forgot-password.tsx` | Forgot password wrapper | Update existing |
| `apps/web/src/routes/verify-email.tsx` | Email verification wrapper | Update existing |

### 4.2 CREATE: Auth Components

| File | Source | Change |
|------|--------|--------|
| `apps/web/src/components/auth/SignUpForm.tsx` | Library | Copy as-is |
| `apps/web/src/components/auth/LoginForm.tsx` | Library | Copy as-is |
| `apps/web/src/components/auth/ForgotPasswordForm.tsx` | Library | Copy as-is |
| `apps/web/src/components/auth/WaitingVerification.tsx` | Library | Copy as-is |
| `apps/web/src/components/auth/SuccessView.tsx` | Library | Copy as-is |
| `apps/web/src/components/auth/SpinningBlocksLoader.tsx` | Library | Copy as-is |

### 4.3 CREATE: Auth Clients

| File | Source | Change |
|------|--------|--------|
| `apps/web/src/lib/clients/auth.ts` | Library authClient.ts | Update API endpoints |
| `apps/web/src/lib/clients/sso.ts` | Library ssoClient.ts | Update API endpoints |
| `apps/web/src/lib/clients/verification.ts` | Library verificationClient.ts | Update API endpoints |

### 4.4 MODIFY: Existing Files

| File | Changes |
|------|---------|
| `apps/web/src/routes/login.tsx` | Replace content with LoginForm import |
| `apps/web/src/routes/forgot-password.tsx` | Replace content with ForgotPasswordForm import |
| `apps/web/src/routes/verify-email.tsx` | Replace content with WaitingVerification import |
| `apps/web/src/lib/vyzorix-auth.ts` | REMOVE - replaced by clients/ |
| `apps/web/src/styles.css` | MERGE Library's index.css styles |

### 4.5 DELETE: Library Copy

| File/Folder | Reason |
|-------------|--------|
| `apps/web/src/library-auth/` | Temp copy, no longer needed |

---

## 5. API Contract Mapping

### Library Endpoints → vyzorix-update-server Endpoints

| Library Endpoint | vyzorix-update-server Endpoint | Status |
|-----------------|-------------------------------|--------|
| `POST /api/auth/register` | `POST /v1/auth/register` | ✅ Exists |
| `POST /api/auth/login` | `POST /v1/auth/login` | ✅ Exists (cookie) |
| `POST /api/auth/logout` | `POST /v1/auth/logout` | ✅ Exists (cookie) |
| `GET /api/auth/me` | `GET /v1/auth/me` | ✅ Exists (cookie) |
| `POST /api/auth/forgot-password` | `POST /v1/auth/forgot-password` | ✅ Exists |
| `GET /api/auth/sso/google` | `GET /v1/auth/google` | ✅ Exists |
| `GET /api/auth/sso/google/callback` | `GET /v1/auth/google/callback` | ✅ Exists (cookie) |
| `GET /api/auth/poll-verification` | ❌ MISSING | **NEED TO ADD** |

### Missing Endpoint: Poll Verification

The Library expects a `GET /v1/auth/poll-verification?token=X` endpoint that doesn't exist in vyzorix-update-server.

**Current vyzorix-update-server has:**
- `POST /v1/auth/verify-email` - Verify email with token
- `POST /v1/auth/resend-verification` - Resend verification email

**Need to ADD:**
```go
// handlers/auth.go
func (ac *AuthController) PollVerification(c *gin.Context) {
    token := c.Query("token")
    // Check if token exists and is verified
    // Return: { verified: true/false, email: "...", operator: {...} }
}
```

---

## 6. Component Integration

### 6.1 SignUpForm Props

```typescript
// Library's SignUpForm expects:
interface SignUpFormProps {
  onSignUp: (data: { fullName: string; email: string; username: string }) => void;
  onSSO: (provider: 'GitHub' | 'Google') => void;
  isSubmitting: boolean;
  triggerToast: (msg: string, type?: 'success' | 'alert') => void;
}
```

**Adaptation for vyzorix-update-server:**
- `onSignUp` → call `register()` from `lib/clients/auth.ts`
- `onSSO` → call `initiateGoogleOAuth()` from `lib/clients/sso.ts`
- `isSubmitting` → use state
- `triggerToast` → use `sonner` (already in project)

### 6.2 LoginForm Props

```typescript
// Library's LoginForm expects:
interface LoginFormProps {
  onLogin: (ident: string, pass: string) => void;
  onSSO: (provider: 'GitHub' | 'Google') => void;
  onForgotPassword: () => void;
  isSubmitting: boolean;
}
```

**Adaptation:**
- `onLogin` → call `login()` from `lib/clients/auth.ts`
- `onForgotPassword` → `navigate('/forgot-password')`

### 6.3 Data Payload Differences

| Field | Library | vyzorix-update-server | Action |
|-------|---------|----------------------|--------|
| Register | `fullName, email, username, password` | `email, password, name` | Map `fullName` → `name`, drop `username` |
| Login | `identity (email), password` | `email, password` | Same |

---

## 7. SSR Hydration Flow

```
┌──────────────────────────────────────────────────────────────────────────┐
│ 1. Browser requests /login                                               │
└────────────────────────────┬───────────────────────────────────────────────┘
                             ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ 2. server.ts (TanStack Start SSR)                                       │
│                                                                          │
│    // Read cookie from request                                           │
│    const cookieHeader = request.headers.get('cookie');                  │
│    if (cookieHeader?.includes('vyz_session=')) {                       │
│      // Validate session via Go API                                      │
│      const res = await fetch('http://localhost:3000/v1/auth/me', {    │
│        headers: { Cookie: cookieHeader }                               │
│      });                                                                │
│      if (res.ok) {                                                      │
│        authState = { isAuthenticated: true, operator: await res.json() }│
│      }                                                                  │
│    }                                                                    │
│                                                                          │
│    // Inject state into HTML                                             │
│    const html = injectStateIntoHtml(rawHtml, authState);               │
│    // Returns: <script>window.__VYZORIX_PREFETCHED_STATE__ = {...}</>  │
└────────────────────────────┬───────────────────────────────────────────────┘
                             ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ 3. Browser receives HTML with state                                      │
│                                                                          │
│    <div id="app"><!--@tanstack/start-entry--></div>                    │
│    <script>window.__VYZORIX_PREFETCHED_STATE__ = {                     │
│      isAuthenticated: true,                                             │
│      operator: { id: "...", email: "...", name: "..." }               │
│    };</script>                                                          │
└────────────────────────────┬───────────────────────────────────────────────┘
                             ↓
┌──────────────────────────────────────────────────────────────────────────┐
│ 4. React hydrateRoot() attaches                                          │
│                                                                          │
│    // Route component checks auth                                        │
│    const { operator } = useOperator();  // From SSR state               │
│    if (operator) {                                                       │
│      navigate('/dashboard');  // Redirect if already logged in             │
│    }                                                                     │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## 8. Auth Layout with Wolf Background

### Problem
Library's App.tsx manages wolf image background as part of its state machine. With route-per-page approach, each route needs the same background treatment.

### Solution: Shared Layout Component

Create a parent route layout that wraps all auth pages with the wolf background.

### File Structure

```
routes/
├── _auth-layout.tsx              ← NEW: Shared wolf background layout
├── login.tsx                     ← Uses _auth-layout
├── signup.tsx                    ← Uses _auth-layout
├── forgot-password.tsx           ← Uses _auth-layout
└── verify-email.tsx             ← Uses _auth-layout
```

### Layout Component Implementation

```tsx
// apps/web/src/routes/_auth-layout.tsx
import { Outlet, createFileRoute } from "@tanstack/react-router";
import { StrictMode } from "react";

// Import wolf image from copied Library assets
import wolfImage from "@/library-auth/assets/images/black_wolf_evening_1781264516831.jpg";

/**
 * Shared authentication layout with wolf background
 * 
 * All auth routes (/login, /signup, /forgot-password, /verify-email)
 * inherit this layout via TanStack's parent route pattern.
 */
export const Route = createFileRoute("/_auth")({
  component: AuthLayoutComponent,
});

function AuthLayoutComponent() {
  return (
    <StrictMode>
      <div className="relative min-h-screen w-full overflow-hidden">
        {/* Wolf background image - fixed, covers viewport */}
        <div 
          className="fixed inset-0 bg-cover bg-center bg-no-repeat"
          style={{ backgroundImage: `url(${wolfImage})` }}
          aria-hidden="true"
        />
        
        {/* Dark gradient overlay for readability */}
        <div 
          className="fixed inset-0 bg-gradient-to-br from-slate-950/80 via-slate-950/70 to-slate-950/85"
          aria-hidden="true"
        />
        
        {/* Content container - centered */}
        <div className="relative z-10 flex min-h-screen items-center justify-center px-4 py-8">
          <div className="w-full max-w-md">
            <Outlet />
          </div>
        </div>
      </div>
    </StrictMode>
  );
}
```

### Route Configuration

Each auth route inherits the parent layout:

```tsx
// apps/web/src/routes/signup.tsx
import { createFileRoute } from "@tanstack/react-router";
import SignUpForm from "@/components/auth/SignUpForm";
// ... imports

// Parent route is /_auth which provides the wolf background layout
export const Route = createFileRoute("/_auth/signup")({
  beforeLoad: () => {
    // Redirect if already authenticated
    const state = getFullHydratedState();
    if (state?.isAuthenticated) {
      throw redirect({ to: "/dashboard" });
    }
  },
  component: SignUpPage,
});

function SignUpPage() {
  const navigate = useNavigate();
  
  const handleSignUp = async (data: SignUpData) => {
    try {
      await register(data.email, data.password, data.name);
      navigate({ to: "/verify-email", search: { email: data.email } });
    } catch (error) {
      toast.error(error.message);
    }
  };

  const handleSSO = (provider: "GitHub" | "Google") => {
    if (provider === "Google") {
      initiateGoogleOAuth();
    }
    // GitHub handling...
  };

  return (
    <div className="w-full">
      <SignUpForm
        onSignUp={handleSignUp}
        onSSO={handleSSO}
        isSubmitting={isLoading}
        triggerToast={(msg, type) => toast(msg)}
      />
    </div>
  );
}
```

### Tailwind Classes Reference (from Library's index.css)

```css
/* Key styles from Library's index.css that need to be merged */
.bg-slate-950 { background-color: #020617; }
.from-slate-950/80 { --tw-gradient-from: rgb(2 6 23 / 0.8); }
.bg-cover { background-size: cover; }
.bg-center { background-position: center; }
.bg-no-repeat { background-repeat: no-repeat; }
.text-slate-400 { color: #94a3b8; }
.text-slate-350 { color: #cbd5e1; }
.bg-rose-600 { background-color: #e11d48; }
.bg-rose-505 { background-color: #f43f5e; }
.text-rose-400 { color: #fb7185; }
.text-rose-450 { color: #f472b6; }
.border-rose-500 { border-color: #f43f5e; }
.shadow-rose-955/30 { --tw-shadow-color: rgb(69 10 30 / 0.3); }
```

---

## 9. Missing Go Endpoints

### Required Endpoints

Library's verification client expects 3 endpoints that don't exist in vyzorix-update-server:

| Endpoint | Method | Library Name | Purpose |
|----------|--------|--------------|---------|
| `/v1/auth/poll-verification` | GET | `pollVerificationStatus` | Poll for email verification status |
| `/v1/auth/resend-token` | POST | `triggerTokenResend` | Resend verification email |
| `/v1/auth/cancel-verification` | POST | `cancelVerificationSession` | Cancel pending verification |

### 9.1 Poll Verification Endpoint

**Purpose:** Allows the frontend to poll until the user clicks the verification link in their email.

**Request:**
```
GET /v1/auth/poll-verification?token=<verification_token>
```

**Response (waiting):**
```json
{
  "status": "waiting",
  "email": "user@example.com"
}
```

**Response (verified):**
```json
{
  "status": "success",
  "email": "user@example.com"
}
```

**Response (expired/invalid):**
```json
{
  "status": "expired" | "invalid",
  "message": "Verification token has expired or is invalid"
}
```

**Implementation:**

```go
// handlers/auth.go

// PollVerification checks the status of an email verification token.
// GET /v1/auth/poll-verification?token=<token>
func (ac *AuthController) PollVerification(c *gin.Context) {
    token := c.Query("token")
    if token == "" {
        c.JSON(400, models.ErrorResponse{
            Error:   "bad_request",
            Message: "token is required",
        })
        return
    }

    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    // Hash the token for lookup
    tokenHash := security.HashToken(token)
    ev, err := ac.store.GetEmailVerificationByTokenHash(ctx, tokenHash)
    if err != nil {
        ac.log.Warn("pollVerification: db error", "err", err)
        c.JSON(500, models.ErrorResponse{
            Error:   "internal_error",
            Message: "verification check failed",
        })
        return
    }

    if ev == nil {
        c.JSON(200, models.VerificationPollResponse{
            Status: "invalid",
            Email:  "",
        })
        return
    }

    // Check if expired
    if time.Now().UTC().After(ev.ExpiresAt) {
        // Delete expired token
        ac.store.DeleteEmailVerification(ctx, ev.ID) //nolint:errcheck
        c.JSON(200, models.VerificationPollResponse{
            Status: "expired",
            Email:  ev.Email,
        })
        return
    }

    // Check if already verified (operator exists and verified)
    op, err := ac.store.GetOperatorByID(ctx, ev.OperatorID)
    if err != nil || op == nil {
        // Operator not yet created, still waiting
        c.JSON(200, models.VerificationPollResponse{
            Status: "waiting",
            Email:  ev.Email,
        })
        return
    }

    if op.EmailVerified {
        // Already verified!
        c.JSON(200, models.VerificationPollResponse{
            Status: "success",
            Email:  ev.Email,
        })
        return
    }

    // Still waiting for user to click
    c.JSON(200, models.VerificationPollResponse{
        Status: "waiting",
        Email:  ev.Email,
    })
}
```

### 9.2 Resend Token Endpoint

**Purpose:** Allows user to request a new verification email.

**Request:**
```
POST /v1/auth/resend-token
Content-Type: application/json

{
  "email": "user@example.com"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Verification link sent"
}
```

**Implementation:** This already exists as `ResendVerification`. We can either:
- Add an alias route in server.go: `POST /resend-token` → `ResendVerification`
- Or update the frontend client to call `/resend-verification` instead

**Recommendation:** Add alias route for cleaner integration:

```go
// server.go Engine() function
auth.POST("/resend-token", s.jwtCtrl.ResendVerification)  // Alias
auth.POST("/resend-verification", s.jwtCtrl.ResendVerification)  // Original
```

### 9.3 Cancel Verification Endpoint

**Purpose:** Allows user to cancel a pending verification (useful if they used wrong email).

**Request:**
```
POST /v1/auth/cancel-verification
Content-Type: application/json

{
  "email": "user@example.com"
}
```

**Response:**
```json
{
  "success": true
}
```

**Implementation:**

```go
// handlers/auth.go

// CancelVerification removes pending verification tokens for an email.
// POST /v1/auth/cancel-verification
func (ac *AuthController) CancelVerification(c *gin.Context) {
    var req models.CancelVerificationRequest
    if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
        c.JSON(400, models.ErrorResponse{
            Error:   "bad_request",
            Message: "invalid request body",
        })
        return
    }

    email := strings.TrimSpace(strings.ToLower(req.Email))
    if email == "" {
        c.JSON(400, models.ErrorResponse{
            Error:   "bad_request",
            Message: "email is required",
        })
        return
    }

    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    // Find operator by email
    op, err := ac.store.GetOperatorByEmail(ctx, email)
    if err != nil {
        ac.log.Warn("cancelVerification: db error", "err", err)
        c.JSON(500, models.ErrorResponse{
            Error:   "internal_error",
            Message: "cancellation failed",
        })
        return
    }

    // Delete any pending verification tokens
    if op != nil {
        if err := ac.store.DeleteEmailVerificationsByOperator(ctx, op.ID); err != nil {
            ac.log.Warn("cancelVerification: failed to delete verifications", "err", err)
        }
    }

    // Always return success for security
    c.JSON(200, map[string]bool{"success": true})
}
```

### 9.4 New Model Types

Add to `pkg/models/auth.go`:

```go
// VerificationPollResponse is the response for polling verification status.
type VerificationPollResponse struct {
    Status string `json:"status"` // "waiting", "success", "expired", "invalid"
    Email  string `json:"email,omitempty"`
}

// CancelVerificationRequest is the payload for canceling pending verification.
type CancelVerificationRequest struct {
    Email string `json:"email"`
}
```

### 9.5 New Storage Methods

Add to `pkg/storage/sqlite.go` or storage interface:

```go
// GetEmailVerificationByTokenHash retrieves an email verification by its token hash.
func (s *Store) GetEmailVerificationByTokenHash(ctx context.Context, tokenHash string) (*EmailVerification, error) {
    // Implementation - query verification_tokens table
}

// DeleteEmailVerification deletes a single email verification.
func (s *Store) DeleteEmailVerification(ctx context.Context, id string) error {
    // Implementation
}

// DeleteEmailVerificationsByOperator deletes all verifications for an operator.
func (s *Store) DeleteEmailVerificationsByOperator(ctx context.Context, operatorID string) error {
    // Implementation
}
```

### 9.6 Route Registration

Update `server.go` to register new endpoints:

```go
// Auth routes
auth := public.Group("/v1/auth")

// Existing endpoints
auth.POST("/register", s.jwtCtrl.Register)
auth.POST("/login", s.jwtCtrl.Login)
auth.POST("/logout", CookieAuth(s.jwtCtrl.session, s.Store), s.jwtCtrl.Logout)
auth.GET("/me", CookieAuth(s.jwtCtrl.session, s.Store), s.jwtCtrl.Me)

// Verification endpoints
auth.POST("/verify-email", s.jwtCtrl.VerifyEmail)
auth.POST("/resend-verification", s.jwtCtrl.ResendVerification)
auth.POST("/resend-token", s.jwtCtrl.ResendVerification)        // Alias
auth.POST("/cancel-verification", s.jwtCtrl.CancelVerification)   // NEW
auth.GET("/poll-verification", s.jwtCtrl.PollVerification)       // NEW
```

---

## 10. Implementation Order

### Phase 2A: Setup (Do First)

1. **Copy wolf image** from `library-auth/` to `assets/images/`
2. **Create auth components directory** `components/auth/`
3. **Copy components** from `library-auth/components/` to `components/auth/`
4. **Create `_auth-layout.tsx`** with wolf background
5. **Create clients/ directory** with 3 client files
6. **Update client API endpoints** to point to vyzorix-update-server
7. **Delete `library-auth/`** temp folder

### Phase 2B: Go Backend Endpoints

1. **Add `VerificationPollResponse` model** to `pkg/models/auth.go`
2. **Add `CancelVerificationRequest` model** to `pkg/models/auth.go`
3. **Add storage methods** for email verification queries
4. **Implement `PollVerification`** handler in `handlers/auth.go`
5. **Implement `CancelVerification`** handler in `handlers/auth.go`
6. **Add route registrations** in `server.go`

### Phase 2C: Routes (One by One)

1. **Create `/signup` route**
   - Uses `_auth` layout (wolf background)
   - Wire up `SignUpForm` → `auth.register()`
   - Wire up `onSSO` → `sso.initiateGoogleOAuth()`
   - Test registration flow

2. **Create `/login` route**
   - Uses `_auth` layout
   - Wire up `LoginForm` → `auth.login()`
   - Wire up `onSSO` → `sso.initiateGoogleOAuth()`
   - Test login + cookie setting

3. **Create `/forgot-password` route**
   - Uses `_auth` layout
   - Wire up `ForgotPasswordForm` → `auth.requestPasswordReset()`
   - Test password reset flow

4. **Create `/verify-email` route**
   - Uses `_auth` layout
   - Wire up `WaitingVerification` → `verification.pollVerification()`
   - Wire up resend → `verification.triggerTokenResend()`
   - Test polling flow

### Phase 2D: Cleanup

1. **Merge styles** from Library's `index.css` into `styles.css`
2. **Remove old `vyzorix-auth.ts`** logic
3. **Test complete flows:**
   - Register → Verify Email → Login
   - Login → Cookie Set → Dashboard
   - Logout → Cookie Cleared

### Phase 2E: Final Testing

1. **SSR Hydration**: Verify auth state injects correctly
2. **Cookie Flow**: Verify HttpOnly cookies work
3. **All 4 Routes**: Test each page renders with wolf background
4. **Email Flow**: Test verification email sent, link clicked, status polled

---

## Summary: Complete File Checklist

### CREATE (18 files)

```
Frontend:
├── routes/
│   ├── _auth-layout.tsx              ← Shared wolf background
│   └── signup.tsx                    ← Sign-up page
├── components/auth/
│   ├── SignUpForm.tsx               ← Copy from Library
│   ├── LoginForm.tsx                ← Copy from Library
│   ├── ForgotPasswordForm.tsx       ← Copy from Library
│   ├── WaitingVerification.tsx       ← Copy from Library
│   ├── SuccessView.tsx              ← Copy from Library
│   └── SpinningBlocksLoader.tsx     ← Copy from Library
└── lib/clients/
    ├── auth.ts                      ← Register, login, logout
    ├── sso.ts                       ← Google OAuth initiation
    └── verification.ts              ← Poll, resend, cancel

Go Backend:
└── pkg/models/auth.go               ← Add new model types

MODIFY (5 files):
├── routes/login.tsx                ← Replace with LoginForm
├── routes/forgot-password.tsx      ← Replace with ForgotPasswordForm
├── routes/verify-email.tsx         ← Replace with WaitingVerification
├── handlers/auth.go                 ← Add Poll, Cancel endpoints
├── server.go                        ← Register new endpoints
└── styles.css                      ← Merge Library styles

DELETE (1 folder):
└── library-auth/                   ← Temp copy

ASSETS:
└── assets/images/wolf.jpg          ← Copy from Library
```

---

*Document Version: 1.1*
*Last Updated: 2026-06-13*
*Status: Implementation Ready*
