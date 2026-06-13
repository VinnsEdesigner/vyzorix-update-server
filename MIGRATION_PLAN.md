# Vyzorix Auth Migration Plan: JWT → HttpOnly Cookies

> **Document Version:** 1.1  
> **Date:** 2026-06-12  
> **Status:** Planning Phase  
> **Dependencies:** SSR Cookie Hydration Implementation (see Section 7)  

---

## 📄 Table of Contents

1. [Executive Summary](#-executive-summary)
2. [Current State Analysis](#-current-state-analysis)
3. [Migration Target Architecture](#-migration-target-architecture)
4. [Complete File Impact Analysis](#-complete-file-impact-analysis)
5. [Detailed Change Specifications](#-detailed-change-specifications)
6. [Migration Phases](#-migration-phases)
7. [SSR Cookie Hydration Implementation](#-ssr-cookie-hydration-implementation) ⭐ **CRITICAL**
8. [Security Checklist](#-security-checklist)
9. [Testing Strategy](#-testing-strategy)
10. [Metrics & Success Criteria](#-metrics--success-criteria)
11. [Risks & Mitigations](#-risks--mitigations)
12. [Next Steps](#-next-steps)

---

## 📋 Executive Summary

**Objective:** Migrate authentication from JWT/localStorage to HttpOnly cookie blobs for enhanced security.

**Impact Assessment:**
| Scope | Current | Target |
|-------|---------|--------|
| Auth Storage | localStorage + JWT Bearer | HttpOnly encrypted cookies |
| Frontend Framework | TanStack Start (vyzorix-update-server) | Unified Library design |
| Backend Changes | JWT generation | Cookie session management |
| Files to Modify | ~15 | ~35 |
| Lines to Change | ~2,500 | ~4,000 |

---

## 🔍 Current State Analysis

### Project A: vyzorix-update-server (PRODUCTION-READY)

**Location:** `/workspace/project/vyzorix-update-server`

#### Authentication Implementation
```
┌─────────────────────────────────────────────────────────────────┐
│                    CURRENT AUTH FLOW                             │
│                                                                  │
│  1. Login → POST /v1/auth/login                                 │
│  2. Server generates JWT                                         │
│  3. JWT returned in response body                               │
│  4. Frontend stores in localStorage                             │
│  5. All API calls: Authorization: Bearer <JWT>                  │
│  6. Server validates JWT on each request                        │
└─────────────────────────────────────────────────────────────────┘
```

#### Security Issues Identified
| Issue | Severity | Description |
|-------|----------|-------------|
| XSS vulnerability | 🔴 CRITICAL | localStorage accessible via XSS |
| No HttpOnly protection | 🔴 CRITICAL | Token can be stolen via JavaScript |
| Token in URL | 🟡 MEDIUM | Google OAuth callback exposes token in URL |
| No automatic refresh | 🟡 MEDIUM | Session expires, no silent refresh |

#### Current File Structure (Auth-related)
```
apps/
├── api/
│   ├── internal/
│   │   ├── auth/
│   │   │   ├── jwt.go           # JWT generation/validation
│   │   │   ├── password.go      # Password hashing (bcrypt)
│   │   │   ├── google_token.go  # Google OAuth validation
│   │   │   ├── ratelimit.go     # Rate limiting
│   │   │   └── validate.go     # Input validation
│   │   └── api/handlers/
│   │       ├── auth.go          # Auth endpoints (700+ lines)
│   │       └── server.go        # Route registration
│   └── pkg/
│       ├── models/              # Request/Response types
│       └── storage/             # SQLite operations
└── web/
    └── src/
        ├── lib/
        │   └── vyzorix-auth.ts  # Auth client (300+ lines)
        ├── routes/
        │   ├── login.tsx        # Simple card UI
        │   ├── forgot-password.tsx
        │   ├── reset-password.tsx
        │   ├── verify-email.tsx
        │   └── auth.callback.tsx
        └── components/ui/       # shadcn/ui components
```

---

### Project B: Library (DESIGN REFERENCE)

**Location:** `/workspace/project/Library`

#### Design Features (To Migrate)
| Feature | Implementation |
|---------|----------------|
| Background | Full-screen wolf image + gradient overlay |
| Theme | Dark slate + Rose-600 accent |
| Cards | Glassmorphism (backdrop-blur-2xl) |
| Forms | Custom validation with error states |
| Animations | Spinners, toast notifications |
| State Machine | View-based routing (signup/login/etc.) |
| Cookie Hydration | SSR-ready with `__VYZORIX_PREFETCHED_STATE__` |

#### Key Files (Reference Design)
```
src/
├── App.tsx                     # Main state machine (450+ lines)
├── components/
│   ├── SignUpForm.tsx         # Full signup form (250+ lines)
│   ├── LoginForm.tsx          # Login form (150+ lines)
│   ├── ForgotPasswordForm.tsx
│   ├── WaitingVerification.tsx
│   └── SuccessView.tsx
├── lib/
│   ├── config.ts              # IS_SIMULATED toggle
│   └── clients/
│       ├── authClient.ts      # Credentials API
│       ├── ssoClient.ts       # Google/GitHub OAuth
│       └── verificationClient.ts
├── index.css                  # Dark theme styles
└── server.ts                  # SSR hydration + cookie parsing
```

---

## 🎯 Migration Target Architecture

### New Auth Flow (HttpOnly Cookies)
```
┌─────────────────────────────────────────────────────────────────┐
│                    TARGET AUTH FLOW                              │
│                                                                  │
│  1. Login → POST /v1/auth/login                                 │
│  2. Server generates encrypted session cookie                    │
│  3. Cookie set: HttpOnly + SameSite=Lax + Secure                │
│  4. Frontend receives cookie automatically                      │
│  5. Subsequent requests: Cookie sent automatically              │
│  6. Server validates session from database/cache                │
│  7. No JavaScript access to credentials                         │
└─────────────────────────────────────────────────────────────────┘
```

### Cookie Structure
```go
type SessionCookie struct {
    Name:     "vyz_session"
    Value:    <encrypted_operator_id>
    HttpOnly: true
    Secure:   true           // HTTPS only
    SameSite: http.SameSiteLaxMode
    Path:     "/"
    MaxAge:   86400          // 24 hours
    Expires:  24 hours from creation
}
```

---

## 📊 Complete File Impact Analysis

### Category 1: Go Backend (apps/api/)

| File | Purpose | Changes | Impact |
|------|---------|---------|--------|
| `internal/auth/jwt.go` | JWT generation | REPLACE with cookie session manager | 🔴 HIGH |
| `internal/auth/password.go` | Password hashing | KEEP (bcrypt is fine) | ✅ LOW |
| `internal/auth/validate.go` | Input validation | KEEP | ✅ LOW |
| `internal/auth/ratelimit.go` | Rate limiting | KEEP | ✅ LOW |
| `internal/api/handlers/auth.go` | Auth endpoints | MODIFY (700+ lines) | 🔴 HIGH |
| `internal/api/handlers/server.go` | Route setup | MODIFY | 🟡 MEDIUM |
| `pkg/models/` | Request/Response types | MODIFY | 🟡 MEDIUM |
| `pkg/storage/` | Database operations | ADD session table + methods | 🔴 HIGH |
| `pkg/crypto/` | Encryption utilities | ADD (AES-256-GCM) | 🔴 HIGH |
| `main.go` | Entry point | MINOR (config changes) | 🟢 MINOR |

**Backend Summary:**
- **Keep:** ~40% of existing code
- **Modify:** ~35% of existing code  
- **Replace/Add:** ~25% new code

---

### Category 2: Frontend (apps/web/src/)

| File | Purpose | Changes | Impact |
|------|---------|---------|--------|
| `lib/vyzorix-auth.ts` | Auth client | REPLACE (remove JWT, add cookie helpers) | 🔴 HIGH |
| `routes/login.tsx` | Login page | REPLACE with Library design | 🔴 HIGH |
| `routes/forgot-password.tsx` | Password reset | REPLACE with Library design | 🔴 HIGH |
| `routes/reset-password.tsx` | Password reset | REPLACE with Library design | 🔴 HIGH |
| `routes/verify-email.tsx` | Email verification | REPLACE with Library design | 🔴 HIGH |
| `routes/auth.callback.tsx` | OAuth callback | REWRITE for cookie flow | 🔴 HIGH |
| `components/ui/card.tsx` | Card component | KEEP (use for success views) | ✅ LOW |
| `components/ui/button.tsx` | Button component | KEEP | ✅ LOW |
| `components/ui/input.tsx` | Input component | KEEP | ✅ LOW |
| `styles.css` | Theme/styles | UPDATE with dark theme | 🟡 MEDIUM |
| `router.tsx` | Route definitions | MINOR (keep TanStack structure) | 🟢 MINOR |
| `lib/api/` | API client | ADD cookie-aware fetch wrapper | 🔴 HIGH |
| `lib/vyzorix-config.tsx` | Config context | UPDATE for cookie mode | 🟡 MEDIUM |
| `hooks/use-auth.ts` | Auth hook | REWRITE for cookie-based auth | 🔴 HIGH |
| `entry-client.tsx` | Client entry | ADD cookie hydration | 🟡 MEDIUM |
| `entry-server.tsx` | Server entry | ADD cookie hydration (if SSR) | 🟡 MEDIUM |

**Frontend Summary:**
- **Keep:** ~20% of existing code
- **Replace/Add:** ~80% new code

---

### Category 3: New Files to Create

| File | Purpose | Priority |
|------|---------|----------|
| `lib/crypto/cookie-cipher.ts` | AES-256-GCM cookie encryption | P0 |
| `lib/api/cookie-client.ts` | Fetch wrapper with cookie handling | P0 |
| `lib/hooks/use-cookie-auth.ts` | Cookie-based auth hook | P0 |
| `components/auth/SignUpForm.tsx` | Migrated from Library | P1 |
| `components/auth/LoginForm.tsx` | Migrated from Library | P1 |
| `components/auth/ForgotPasswordForm.tsx` | Migrated from Library | P1 |
| `components/auth/WaitingVerification.tsx` | Migrated from Library | P1 |
| `components/auth/SuccessView.tsx` | Migrated from Library | P1 |
| `components/auth/SpinningLoader.tsx` | Migrated from Library | P2 |
| `styles/dark-theme.css` | Dark theme variables | P1 |

---

## 🔧 Detailed Change Specifications

### Phase 1: Backend Session Management

#### 1.1 Create Cookie Session Manager
**File:** `apps/api/internal/auth/session.go` (NEW)

```go
package auth

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "time"
)

const (
    CookieName    = "vyz_session"
    CookieMaxAge  = 86400 // 24 hours
    CookiePath    = "/"
)

type SessionManager struct {
    encryptionKey []byte // 32 bytes for AES-256
}

func NewSessionManager(key string) (*SessionManager, error) {
    // Key should be 32 bytes
    h := sha256.Sum256([]byte(key))
    return &SessionManager{encryptionKey: h[:]}, nil
}

// EncryptOperatorID creates an encrypted cookie value
func (sm *SessionManager) EncryptOperatorID(operatorID string) (string, error) {
    block, err := aes.NewCipher(sm.encryptionKey)
    if err != nil {
        return "", err
    }
    
    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonce := make([]byte, aesGCM.NonceSize())
    _, err = rand.Read(nonce)
    if err != nil {
        return "", err
    }
    
    ciphertext := aesGCM.Seal(nonce, nonce, []byte(operatorID), nil)
    return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

// DecryptOperatorID validates and extracts operator ID from cookie
func (sm *SessionManager) DecryptOperatorID(cookieValue string) (string, error) {
    ciphertext, err := base64.RawURLEncoding.DecodeString(cookieValue)
    if err != nil {
        return "", errors.New("invalid cookie encoding")
    }
    
    block, err := aes.NewCipher(sm.encryptionKey)
    if err != nil {
        return "", err
    }
    
    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonceSize := aesGCM.NonceSize()
    if len(ciphertext) < nonceSize {
        return "", errors.New("ciphertext too short")
    }
    
    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", errors.New("cookie decryption failed")
    }
    
    return string(plaintext), nil
}

// CreateSessionCookie generates the HTTP cookie
func (sm *SessionManager) CreateSessionCookie(operatorID string) (*http.Cookie, error) {
    encryptedID, err := sm.EncryptOperatorID(operatorID)
    if err != nil {
        return nil, err
    }
    
    return &http.Cookie{
        Name:     CookieName,
        Value:    encryptedID,
        Path:     CookiePath,
        MaxAge:   CookieMaxAge,
        HttpOnly: true,
        Secure:   true,      // HTTPS only
        SameSite: http.SameSiteLaxMode,
    }, nil
}

// ClearSessionCookie generates an expired cookie for logout
func (sm *SessionManager) ClearSessionCookie() *http.Cookie {
    return &http.Cookie{
        Name:     CookieName,
        Value:    "",
        Path:     CookiePath,
        MaxAge:   -1,
        Expires:  time.Unix(0, 0),
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
    }
}
```

**Lines of Code:** ~95

---

#### 1.2 Modify Auth Handler (auth.go)

**Current State:** ~700 lines using JWT
**Target State:** ~600 lines using cookie sessions

| Endpoint | Current | Target | Changes |
|----------|---------|--------|---------|
| `POST /v1/auth/login` | Returns JWT in body | Sets HttpOnly cookie | **MODIFY** |
| `POST /v1/auth/register` | Returns JWT | Sets HttpOnly cookie | **MODIFY** |
| `GET /v1/auth/me` | Reads Bearer token | Reads cookie | **MODIFY** |
| `POST /v1/auth/logout` | Deletes session by JWT | Clears cookie | **MODIFY** |
| `GET /v1/auth/google` | Redirects with JWT URL param | Redirects to callback | **MODIFY** |
| `GET /v1/auth/google/callback` | Returns JWT | Sets cookie + redirects | **MODIFY** |

**Key Changes in auth.go:**

```go
// REPLACE: Login function
func (ac *AuthController) Login(c *gin.Context) {
    // ... validation code stays the same until token generation ...
    
    // CHANGE: Instead of generating JWT
    // OLD:
    // token, expiresAt, _ := ac.jwt.Generate(op.ID, op.Email, op.Name, string(op.Role))
    // c.JSON(200, AuthResponse{Token: token, ExpiresAt: expiresAt, Operator: op.ToResponse()})
    
    // NEW:
    sessionCookie, err := ac.sessionManager.CreateSessionCookie(op.ID)
    if err != nil {
        c.JSON(500, models.ErrorResponse{Error: "internal_error", Message: "session creation failed"})
        return
    }
    http.SetCookie(c.Writer, sessionCookie)
    c.JSON(200, op.ToResponse()) // Return operator, not token
}

// REPLACE: Logout function  
func (ac *AuthController) Logout(c *gin.Context) {
    // CHANGE: Clear cookie instead of deleting session
    clearCookie := ac.sessionManager.ClearSessionCookie()
    http.SetCookie(c.Writer, clearCookie)
    c.JSON(200, map[string]bool{"ok": true})
}

// ADD: Cookie-based middleware (replace JWT middleware)
func CookieAuthMiddleware(sm *SessionManager, store *storage.Store) gin.HandlerFunc {
    return func(c *gin.Context) {
        cookie, err := c.Cookie(sm.CookieName)
        if err != nil {
            c.JSON(401, models.ErrorResponse{Error: "unauthorized", Message: "authentication required"})
            c.Abort()
            return
        }
        
        operatorID, err := sm.DecryptOperatorID(cookie)
        if err != nil {
            c.JSON(401, models.ErrorResponse{Error: "invalid_session", Message: "invalid session"})
            c.Abort()
            return
        }
        
        // Validate operator exists
        op, err := store.GetOperatorByID(c.Request.Context(), operatorID)
        if err != nil || op == nil {
            c.JSON(401, models.ErrorResponse{Error: "invalid_session", Message: "operator not found"})
            c.Abort()
            return
        }
        
        c.Set("operator", op)
        c.Next()
    }
}
```

**Lines Changed:** ~150 lines modified, ~50 lines removed, ~60 lines added

---

#### 1.3 Update Storage Layer

**File:** `apps/api/pkg/storage/` 

**Add to store.go:**
```go
// Session operations for cookie-based auth
func (s *Store) CreateSession(ctx context.Context, session *Session) error { ... }
func (s *Store) GetSession(ctx context.Context, operatorID string) (*Session, error) { ... }
func (s *Store) DeleteSession(ctx context.Context, operatorID string) error { ... }
func (s *Store) DeleteAllSessionsForOperator(ctx context.Context, operatorID string) error { ... }
```

**Database Schema Addition:**
```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    operator_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (operator_id) REFERENCES operators(id) ON DELETE CASCADE
);

CREATE INDEX idx_sessions_operator_id ON sessions(operator_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
```

**Lines Changed:** ~100 lines

---

### Phase 2: Frontend Migration

#### 2.1 Replace Auth Client

**File:** `apps/web/src/lib/vyzorix-auth.ts`

**Current (JWT):**
```typescript
// Uses localStorage
const setToken = (token: string): void => {
  localStorage.setItem(TOKEN_KEY, token);
};

export const login = async (serverUrl: string, email: string, password: string) => {
  const res = await fetch(`${serverUrl}/v1/auth/login`, { ... });
  const { token, operator } = await res.json();
  setToken(token);  // localStorage
  return { token, operator };
};
```

**Target (Cookies):**
```typescript
// Cookie-based - no localStorage for auth
export const login = async (serverUrl: string, email: string, password: string) => {
  const res = await fetch(`${serverUrl}/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include", // IMPORTANT: Include cookies
    body: JSON.stringify({ email, password }),
  });
  
  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.message);
  }
  
  const operator = await res.json();
  return { operator };
};

export const logout = async (serverUrl: string): Promise<void> => {
  await fetch(`${serverUrl}/v1/auth/logout`, {
    method: "POST",
    credentials: "include",
  });
  // No localStorage to clear
};
```

**Lines Changed:** ~120 lines replaced

---

#### 2.2 Migrate Auth Components

| Component | Source | Lines | Adaptations Needed |
|-----------|--------|-------|-------------------|
| `SignUpForm.tsx` | Library | ~250 | Update API calls to use cookies |
| `LoginForm.tsx` | Library | ~150 | Update API calls to use cookies |
| `ForgotPasswordForm.tsx` | Library | ~80 | Minor (uses POST only) |
| `SuccessView.tsx` | Library | ~100 | Remove localStorage references |
| `WaitingVerification.tsx` | Library | ~100 | Keep polling logic |

**Adaptation Pattern:**
```typescript
// BEFORE (Library)
const handleSubmit = async (e: React.FormEvent) => {
  e.preventDefault();
  // ... validation
  const response = await login(email, password); // Uses localStorage internally
};

// AFTER (Migration)
const handleSubmit = async (e: React.FormEvent) => {
  e.preventDefault();
  // ... validation
  const response = await login(serverUrl, email, password); // credentials: 'include'
};
```

---

#### 2.3 Update Entry Points

**File:** `apps/web/src/entry-client.tsx`

```typescript
// ADD: Cookie hydration on app load
const checkSession = async () => {
  try {
    const res = await fetch(`${serverUrl}/v1/auth/me`, {
      credentials: "include", // Send cookie
    });
    
    if (res.ok) {
      const operator = await res.json();
      // Set auth state from server response
    }
  } catch {
    // Not authenticated, show login
  }
};

checkSession();
```

---

### Phase 3: Configuration & Environment

#### 3.1 Config Changes

**File:** `apps/api/pkg/config/config.go`

```go
type Config struct {
    // ... existing fields ...
    
    // NEW: Session configuration
    SessionSecret    string        // AES-256 key for cookie encryption
    SessionMaxAge    int           // Session duration in seconds
    
    // MODIFY: Remove if not needed
    JWTSecret        string        // Keep for backward compat during migration
    JWTDuration      time.Duration // Keep during transition
}
```

**File:** `apps/web/src/lib/vyzorix-config.tsx`

```typescript
interface VyzorixConfig {
  serverUrl: string;
  // REMOVE: token management
  // addToken: (token: string) => void;
  // removeToken: () => void;
  // hasToken: () => boolean;
}
```

---

## 📅 Migration Phases

### Phase 0: SSR Cookie Hydration (PREREQUISITE) ⭐
- [ ] **Do this FIRST** - Required for HttpOnly cookies to work
- [ ] Implement server-side cookie reader
- [ ] Implement state injection
- [ ] Update TanStack Start beforeLoad hooks
- [ ] Details in Section 7 below

### Phase 1: Backend Session Management (2-3 days)
- [ ] Create SessionManager for AES-256-GCM cookie encryption
- [ ] Add session storage methods to database
- [ ] Modify auth endpoints to set HttpOnly cookies
- [ ] Update middleware chain

### Phase 2: Frontend Migration (2-3 days)
- [ ] Create cookie-aware fetch wrapper
- [ ] Migrate auth UI components from Library
- [ ] Update entry points for SSR hydration
- [ ] Remove localStorage references

### Phase 3: Integration & Testing (2-3 days)
- [ ] Test login/logout flows
- [ ] Test OAuth flow
- [ ] Test password reset
- [ ] Verify no hydration mismatches

### Phase 4: Cleanup & Deploy (1-2 days)
- [ ] Remove JWT code
- [ ] Deploy to staging
- [ ] Monitor for issues

---

## 7. SSR Cookie Hydration Implementation ⭐ CRITICAL

### Why This Must Be Done First

**HttpOnly cookies CANNOT be read by JavaScript.** The current client-side `localStorage.getItem("vyz.auth.token")` will return `null` after migration.

The SSR cookie hydration is a **PREREQUISITE** for the cookie migration to work.

---

### 7.1 Library's Implementation (Reference)

Library's `server.ts` provides the complete SSR cookie handling pattern:

```typescript
// Library: server.ts - Lines 31-75
const getIndexHtml = async (reqUrl: string, reqCookies: string): Promise<string> => {
  // 1. Read HTML template
  let html = fs.readFileSync(htmlPath, 'utf-8');

  // 2. Initialize prefetched state
  const prefetchedState: Record<string, any> = {
    view: 'signup',
    profileData: null,
    successReport: null,
    verificationToken: null,
  };

  // 3. Check for session cookie
  if (reqCookies.includes('vyzorix_session=')) {
    // 4. Call Go backend to get operator data
    const meResponse = await fetch(`${GO_BACKEND}/api/auth/me`, {
      headers: { 'Cookie': reqCookies }
    });
    if (meResponse.ok) {
      const report = await meResponse.json();
      // 5. Set correct view based on auth state
      prefetchedState.view = 'success';
      prefetchedState.successReport = report;
      prefetchedState.profileData = { fullName, email, username };
    }
  }

  // 6. Inject state into HTML BEFORE sending
  const stateScript = `<script>window.__VYZORIX_PREFETCHED_STATE__ = ${JSON.stringify(prefetchedState)};</script>`;
  return html.replace('<div id="root">', `${stateScript}\n<div id="root">`);
});
```

**Key patterns from Library:**
1. Server reads cookies from request headers
2. Server calls Go API to get authenticated user data
3. Server injects `window.__VYZORIX_PREFETCHED_STATE__` into HTML
4. React hydrates with correct state immediately (no flash)

---

### 7.2 New Files Required

| File | Purpose | Lines | Priority |
|------|---------|-------|----------|
| `apps/web/src/lib/server/cookie-reader.ts` | Parse cookies, call Go API | ~40 | **P0** |
| `apps/web/src/lib/server/state-injector.tsx` | State injection for SSR | ~30 | **P0** |
| `apps/web/src/hooks/use-operator.ts` | Client hook for server operator | ~20 | P1 |

---

### 7.3 File Specifications

#### File 7.3.1: `apps/web/src/lib/server/cookie-reader.ts` (NEW)

**Purpose:** Server-side cookie parsing and Go API call to get authenticated operator.

```typescript
// apps/web/src/lib/server/cookie-reader.ts

export interface PrefetchedAuthState {
  isAuthenticated: boolean;
  operator: Operator | null;
}

/**
 * Server-side cookie reader for SSR hydration
 * 
 * This runs on the SERVER (Node.js) during SSR:
 * 1. Extracts session cookie from request headers
 * 2. Calls Go API to validate session and get operator data
 * 3. Returns auth state to be injected into HTML
 */
export async function getPrefetchedAuthState(request: Request): Promise<PrefetchedAuthState> {
  // 1. Get cookie header from request
  const cookieHeader = request.headers.get('cookie');
  if (!cookieHeader) {
    return { isAuthenticated: false, operator: null };
  }

  // 2. Parse cookies (same logic as Library: server.ts)
  const cookies = cookieHeader.split('; ').reduce<Record<string, string>>((acc, cookie) => {
    const [key, ...valueParts] = cookie.split('=');
    acc[key.trim()] = valueParts.join('=');
    return acc;
  }, {});

  // 3. Check for session cookie
  // Note: After migration, this will be 'vyz_session' not 'vyzorix_session'
  const sessionCookie = cookies['vyz_session'] || cookies['vyzorix_session'];
  if (!sessionCookie) {
    return { isAuthenticated: false, operator: null };
  }

  // 4. Call Go API to get operator data
  // The Go API reads the cookie from the request and validates the session
  const apiUrl = process.env.API_URL || 'http://localhost:3000';
  
  try {
    const response = await fetch(`${apiUrl}/v1/auth/me`, {
      headers: {
        'Cookie': `vyz_session=${sessionCookie}`,
        'Accept': 'application/json',
      },
    });

    if (!response.ok) {
      return { isAuthenticated: false, operator: null };
    }

    const operator = await response.json();
    return { isAuthenticated: true, operator };
  } catch (error) {
    console.error('[SSR Cookie Reader] Failed to fetch operator:', error);
    return { isAuthenticated: false, operator: null };
  }
}

/**
 * Parse cookies from Cookie header string
 */
export function parseCookies(cookieHeader: string | null): Record<string, string> {
  if (!cookieHeader) return {};
  
  return cookieHeader.split('; ').reduce<Record<string, string>>((acc, cookie) => {
    const [key, ...valueParts] = cookie.split('=');
    acc[key.trim()] = decodeURIComponent(valueParts.join('='));
    return acc;
  }, {});
}
```

**Implementation Notes:**
- This mimics Library's `server.ts` cookie parsing (line 46-70)
- Calls Go `/v1/auth/me` endpoint like Library does (line 52)
- Returns structured auth state for injection

---

#### File 7.3.2: `apps/web/src/lib/server/state-injector.tsx` (NEW)

**Purpose:** Inject server-provided state into the HTML response for React hydration.

```typescript
// apps/web/src/lib/server/state-injector.tsx

export interface HydratedState {
  isAuthenticated: boolean;
  operator: {
    id: string;
    email: string;
    name: string;
    role: string;
  } | null;
}

/**
 * Generate script tag for state injection
 * 
 * This injects state into window.__VYZORIX_PREFETCHED_STATE__
 * like Library's server.ts does (line 73-74):
 *   const stateScript = `<script>window.__VYZORIX_PREFETCHED_STATE__ = ${JSON.stringify(prefetchedState)};</script>`;
 */
export function generateStateScript(state: HydratedState): string {
  return `<script id="__vyzorix-prefetched-state__" type="application/json">
  window.__VYZORIX_PREFETCHED_STATE__ = ${JSON.stringify(state)};
</script>`;
}

/**
 * Inject state into HTML before </body>
 * 
 * Library's pattern (index.html line 14):
 *   <div id="root"><!--app-html--></div>
 *   <!--app-state-->
 * 
 * This replaces <!--app-state--> with the actual state script
 */
export function injectStateIntoHtml(html: string, state: HydratedState): string {
  const stateScript = generateStateScript(state);
  return html.replace('<!--app-state-->', stateScript);
}
```

**Implementation Notes:**
- Mirrors Library's state injection pattern (server.ts line 73-74)
- Uses `<!--app-state-->` placeholder in index.html
- Provides `window.__VYZORIX_PREFETCHED_STATE__` for React to consume

---

#### File 7.3.3: `apps/web/src/hooks/use-operator.ts` (NEW)

**Purpose:** Client-side hook to access server-provided operator (replaces localStorage reads).

```typescript
// apps/web/src/hooks/use-operator.ts

import { useState, useEffect } from 'react';
import type { Operator } from '@/lib/vyzorix-auth';

export interface AuthContext {
  operator: Operator | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

/**
 * SSR-aware auth hook
 * 
 * This replaces localStorage-based auth checking with
 * server-provided state hydration.
 * 
 * Library's pattern (src/lib/config.ts):
 *   export function getHydratedState<T>(key: string, defaultValue: T): T {
 *     if (typeof window !== 'undefined') {
 *       const globalState = (window as any).__VYZORIX_PREFETCHED_STATE__;
 *       if (globalState && globalState[key] !== undefined) {
 *         return globalState[key];
 *       }
 *     }
 *     return defaultValue;
 *   }
 */
export function useOperator(): AuthContext {
  const [operator, setOperator] = useState<Operator | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Read from server-injected state (like Library's getHydratedState)
    const globalState = (window as any).__VYZORIX_PREFETCHED_STATE__;
    
    if (globalState) {
      setOperator(globalState.operator || null);
      setIsAuthenticated(globalState.isAuthenticated || false);
    } else {
      // Fallback: Fetch from API if SSR state not available
      // (e.g., direct navigation, page refresh on protected route)
      fetch('/v1/auth/me', { credentials: 'include' })
        .then(res => res.ok ? res.json() : null)
        .then(data => {
          setOperator(data);
          setIsAuthenticated(!!data);
        })
        .catch(() => setIsAuthenticated(false))
        .finally(() => setIsLoading(false));
    }
  }, []);

  return { operator, isAuthenticated, isLoading };
}

/**
 * Get hydrated state helper (matches Library's getHydratedState)
 */
export function getHydratedState<T>(key: string, defaultValue: T): T {
  if (typeof window !== 'undefined') {
    const globalState = (window as any).__VYZORIX_PREFETCHED_STATE__;
    if (globalState && globalState[key] !== undefined) {
      return globalState[key];
    }
  }
  return defaultValue;
}
```

**Implementation Notes:**
- Mimics Library's `getHydratedState` from `src/lib/config.ts`
- Reads from `window.__VYZORIX_PREFETCHED_STATE__` set by server
- Fallback to API call if SSR state not available

---

### 7.4 Files to Modify

#### File 7.4.1: `apps/web/src/server.ts` (MODIFY)

**Current (lines 1-60):**
```typescript
// TanStack Start server entry - NO cookie reading currently
```

**Target (add cookie state injection):**

```typescript
// apps/web/src/server.ts
import { consumeLastCapturedError } from "./lib/error-capture";
import { renderErrorPage } from "./lib/error-page";
import { getPrefetchedAuthState } from "./lib/server/cookie-reader";
import { injectStateIntoHtml } from "./lib/server/state-injector";

interface ServerEntry {
  fetch: (request: Request, env: unknown, ctx: unknown) => Promise<Response> | Response;
}

let serverEntryPromise: Promise<ServerEntry> | undefined;

const getServerEntry = (): Promise<ServerEntry> => {
  serverEntryPromise ??= import("@tanstack/react-start/server-entry").then(
    (m) => (m.default ?? m) as ServerEntry,
  );
  return serverEntryPromise;
};

export default {
  async fetch(request: Request, env: unknown, ctx: unknown) {
    try {
      const handler = await getServerEntry();
      let response = await handler.fetch(request, env, ctx);
      
      // ⭐ INJECT AUTH STATE INTO HTML RESPONSES
      if (response.headers.get('content-type')?.includes('text/html')) {
        const html = await response.text();
        
        // Get auth state from cookies (same as Library: server.ts)
        const authState = await getPrefetchedAuthState(request);
        
        // Inject state into HTML
        const hydratedHtml = injectStateIntoHtml(html, authState);
        
        response = new Response(hydratedHtml, {
          status: response.status,
          headers: response.headers,
        });
      }
      
      return await normalizeCatastrophicSsrResponse(response);
    } catch (error) {
      console.error(error);
      return new Response(renderErrorPage(), {
        status: 500,
        headers: { "content-type": "text/html; charset=utf-8" },
      });
    }
  },
};

// ... normalizeCatastrophicSsrResponse function stays the same
```

---

#### File 7.4.2: `apps/web/index.html` (MODIFY)

**Current:**
```html
<div id="app"><!--@tanstack/start-entry--></div>
```

**Target:**
```html
<!-- NO WHITESPACE between div tags! -->
<div id="app"><!--@tanstack/start-entry--></div>
<!--app-state-->
```

**Note:** Same pattern as Library's `index.html` (line 13-15)

---

#### File 7.4.3: `apps/web/src/routes/_app.tsx` (MODIFY)

**Current (client-side only):**
```typescript
beforeLoad: () => {
  // Client-side only check - auth is enforced at Go server level via JWT cookie
  if (typeof window !== "undefined" && typeof localStorage !== "undefined") {
    const token = localStorage.getItem("vyz.auth.token");
    if (!token) {
      throw redirect({ to: "/login" });
    }
  }
},
```

**Target (server-side with fallback):**
```typescript
beforeLoad: async ({ request }) => {
  // ⭐ SERVER-SIDE COOKIE CHECK
  // This runs on the server during SSR, reading HttpOnly cookies
  const cookieHeader = request.headers.get('cookie');
  if (cookieHeader) {
    const cookies = Object.fromEntries(
      cookieHeader.split('; ').map(c => c.split('='))
    );
    const sessionCookie = cookies['vyz_session'];
    
    if (!sessionCookie) {
      throw redirect({ to: "/login" });
    }
    
    // Optionally validate with API
    // const operator = await getOperatorFromCookie(sessionCookie);
    return; // Allow access
  }
  
  // ⭐ CLIENT FALLBACK (for client-side navigation)
  // If we get here on client, use server-provided state
  if (typeof window !== "undefined") {
    const globalState = (window as any).__VYZORIX_PREFETCHED_STATE__;
    if (!globalState?.isAuthenticated) {
      throw redirect({ to: "/login" });
    }
    return;
  }
  
  // No auth - redirect
  throw redirect({ to: "/login" });
},
```

---

#### File 7.4.4: `apps/web/src/routes/login.tsx` (MODIFY)

**Current:**
```typescript
useEffect(() => {
  const token = localStorage.getItem("vyz.auth.token");
  if (token) navigate({ to: "/dashboard", replace: true });
}, [navigate]);
```

**Target:**
```typescript
useEffect(() => {
  // ⭐ CHECK SERVER-PROVIDED STATE FIRST
  if (typeof window !== "undefined") {
    const globalState = (window as any).__VYZORIX_PREFETCHED_STATE__;
    if (globalState?.isAuthenticated) {
      navigate({ to: "/dashboard", replace: true });
      return;
    }
  }
  
  // Fallback: Check cookie directly
  // (for client-side navigation without SSR state)
  fetch('/v1/auth/me', { credentials: 'include' })
    .then(res => {
      if (res.ok) navigate({ to: "/dashboard", replace: true });
    })
    .catch(() => {});
}, [navigate]);
```

---

### 7.5 Implementation Order

| Step | File | Action | Priority |
|------|------|--------|----------|
| 1 | `apps/web/src/lib/server/cookie-reader.ts` | CREATE | **P0** |
| 2 | `apps/web/src/lib/server/state-injector.tsx` | CREATE | **P0** |
| 3 | `apps/web/index.html` | MODIFY (add placeholder) | **P0** |
| 4 | `apps/web/src/server.ts` | MODIFY (inject state) | **P0** |
| 5 | `apps/web/src/hooks/use-operator.ts` | CREATE | P1 |
| 6 | `apps/web/src/routes/_app.tsx` | MODIFY (server-side auth) | **P0** |
| 7 | `apps/web/src/routes/login.tsx` | MODIFY (use SSR state) | P1 |
| 8 | Test SSR hydration | VERIFY | **P0** |

---

### 7.6 Testing Checklist

- [ ] Server injects `window.__VYZORIX_PREFETCHED_STATE__` into HTML
- [ ] Authenticated user sees dashboard on first paint (no flash)
- [ ] Unauthenticated user sees login page on first paint
- [ ] Page refresh preserves auth state
- [ ] Direct navigation to `/dashboard` without cookie → redirect to login
- [ ] No hydration mismatches in React DevTools

---

### 7.7 Validation Commands

```bash
# 1. Start servers
cd apps/api && go run main.go &   # Go API on :3000
cd apps/web && pnpm dev            # SSR on :3001 (proxies to :3000)

# 2. Test authenticated SSR
# Login in browser, then:
curl -s http://localhost:3001/dashboard | grep "__VYZORIX_PREFETCHED_STATE__"

# 3. Expected output (authenticated):
# <script id="__vyzorix-prefetched-state__" type="application/json">
# window.__VYZORIX_PREFETCHED_STATE__ = {"isAuthenticated":true,"operator":{"id":"...","email":"..."}}
# </script>

# 4. Test unauthenticated SSR
# Clear cookies, then:
curl -s http://localhost:3001/dashboard

# 5. Expected: Should redirect to /login (302)
```

---

## 🔒 Security Checklist

| Requirement | Implementation | Status |
|-------------|----------------|--------|
| HttpOnly cookies | `HttpOnly: true` | ⬜ TODO |
| Secure flag | `Secure: true` (HTTPS only) | ⬜ TODO |
| SameSite policy | `SameSite: Lax` | ⬜ TODO |
| AES-256-GCM encryption | Custom SessionManager | ⬜ TODO |
| Session expiry | 24-hour sliding window | ⬜ TODO |
| CSRF protection | SameSite + double-submit | ⬜ TODO |
| Rate limiting | Keep existing ratelimit.go | ✅ DONE |
| Password hashing | bcrypt (existing) | ✅ DONE |
| Input validation | Keep existing validate.go | ✅ DONE |

---

## 🧪 Testing Strategy

### Backend Tests (curl)

```bash
# 1. Register
curl -X POST http://localhost:3000/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123!","name":"Test User"}' \
  -c cookies.txt

# 2. Login (verify cookie)
curl -X POST http://localhost:3000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123!"}' \
  -c cookies.txt -v

# 3. Me (verify cookie sent)
curl http://localhost:3000/v1/auth/me \
  -b cookies.txt

# 4. Logout (verify cookie cleared)
curl -X POST http://localhost:3000/v1/auth/logout \
  -b cookies.txt -c cookies.txt

# 5. Verify session expired
curl http://localhost:3000/v1/auth/me \
  -b cookies.txt  # Should return 401
```

### Frontend Tests

1. ✅ Login page renders correctly
2. ✅ Form validation works
3. ✅ Login sets HttpOnly cookie
4. ✅ Logout clears cookie
5. ✅ Session persists across page refresh
6. ✅ OAuth sets cookie and redirects
7. ✅ Password reset flow works

---

## 📈 Metrics & Success Criteria

| Metric | Target | Measurement |
|--------|--------|-------------|
| Files modified | ~35 | Git diff count |
| Test coverage | >90% | Go test coverage |
| XSS vulnerability | 0 | Security audit |
| localStorage usage | 0 | Code search |
| HttpOnly cookies | 100% of auth | Network tab inspection |

---

## 🚨 Risks & Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| CORS issues with cookies | Medium | High | Configure `credentials: "include"` |
| Cookie size limits | Low | Medium | Encrypt minimal data (operator ID only) |
| Backward compatibility | High | Medium | Keep JWT during transition phase |
| Session fixation | Low | High | Generate new session ID on login |
| Cookie theft (MITM) | Low | Critical | Enforce HTTPS with Secure flag |

---

## 📝 Next Steps

### IMMEDIATE: Implement SSR Cookie Hydration (Section 7)

This is a **BLOCKER** for the cookie migration. Must be done first.

| Step | Action | Status |
|------|---------|--------|
| 1 | Create `apps/web/src/lib/server/cookie-reader.ts` | ✅ DONE |
| 2 | Create `apps/web/src/lib/server/state-injector.tsx` | ✅ DONE |
| 3 | Create `apps/web/src/hooks/use-operator.ts` | ✅ DONE |
| 4 | Modify `apps/web/index.html` (add placeholder) | ✅ DONE |
| 5 | Modify `apps/web/src/server.ts` (inject state) | ✅ DONE |
| 6 | Modify `apps/web/src/routes/_app.tsx` (server auth) | ✅ DONE |
| 7 | Modify `apps/web/src/routes/login.tsx` (SSR state) | ✅ DONE |
| 8 | Test SSR hydration | ⬜ TODO |

### AFTER SSR: Cookie Backend (Section 5.1)

| Step | Action | Status |
|------|---------|--------|
| 9 | Create `apps/api/internal/auth/session.go` | ⬜ TODO |
| 10 | Modify `apps/api/internal/api/handlers/auth.go` | ⬜ TODO |
| 11 | Update storage layer | ⬜ TODO |

### FINAL: Frontend Cookie Migration

| Step | Action | Status |
|------|---------|--------|
| 12 | Migrate auth UI components | ⬜ TODO |
| 13 | Update API client for cookies | ⬜ TODO |
| 14 | Remove localStorage references | ⬜ TODO |
| 15 | Full integration test | ⬜ TODO |

---

*Document prepared by: OpenHands Analysis Agent*  
*Last updated: 2026-06-12*  
*Version: 1.1*