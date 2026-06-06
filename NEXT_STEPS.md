# Vyzorix Update Server — Next Steps Planning

> **Last Updated:** June 2026  
> **Status:** PR #6 pending merge, testing in progress

---

## 🎯 Priority Recommendations

### Why These Tasks?

1. **Rate Limit Auth Endpoints** — Prevent brute force attacks on login/register
2. **Dark Mode Toggle** — User-facing feature users expect
3. **OpenAPI Docs** — Developer experience and API discoverability  
4. **Database Migrations** — Foundation for safe schema evolution

---

## Task 1: Wire Up Protected Routes (DEFERRED) 🛡️

> **Status:** Paused - UI testing in progress, not deployed

### Why Paused?
- Auth pages exist but `_app.tsx` has auth disabled for local exploration
- Need deployed server before wiring up
- Prevents dashboard access without login

### Implementation Plan
```go
// In _app.tsx, add beforeLoad guard:
beforeLoad(({ context, location }) => {
  const { isAuthenticated } = useAuth();
  if (!isAuthenticated && !isPublicRoute(location.pathname)) {
    throw redirect({ to: '/login' });
  }
});
```

### Files to Modify
- `src/routes/_app.tsx` — Add auth guard
- `src/routes/login.tsx` — Redirect if already authenticated
- `src/hooks/use-auth.ts` — Enhance with route protection

---

## Task 2: Rate Limit Auth Endpoints 🔒

### Why Important?
- Prevent brute force password attacks
- Mitigate credential stuffing
- Protect against denial of service on auth endpoints

### Implementation

#### Current State
- Rate limiter exists in `middleware/rate_limiter.go`
- Currently applies to all public endpoints equally

#### Proposed Changes
```go
// middleware/rate_limiter.go

// Auth-specific rate limiter with stricter limits
type AuthRateLimiter struct {
    LoginLimit       RateLimitRule // 5 attempts/minute per IP
    RegisterLimit   RateLimitRule // 3 attempts/minute per IP
    ForgotPasswordLimit RateLimitRule // 2 attempts/minute per IP
}
```

#### Endpoints to Protect
| Endpoint | Current Limit | Proposed Limit |
|----------|--------------|----------------|
| `/v1/auth/login` | 60/min | **10/min** |
| `/v1/auth/register` | 60/min | **5/min** |
| `/v1/auth/forgot-password` | 60/min | **3/min** |
| `/v1/auth/resend-verification` | 60/min | **3/min** |

#### Files to Modify
- `middleware/rate_limiter.go` — Add auth-specific limits
- `controllers/server.go` — Apply stricter limits to auth routes
- `middleware/rate_limiter_test.go` — Add auth rate limit tests

---

## Task 4: Dark Mode Toggle 🌙

### Why Important?
- Users expect dark mode in modern apps
- Settings page has appearance section but it's empty
- Reduces eye strain in low-light environments

### Implementation

#### Current State
- Settings page at `/settings/appearance` exists
- Tailwind dark mode configured in `tailwind.config.js`
- No toggle implementation

#### Proposed Changes
```tsx
// src/routes/_app.settings.appearance.tsx

// Add dark mode toggle component
function AppearanceSettings() {
  const [theme, setTheme] = useState<'light' | 'dark' | 'system'>('system');
  
  // Persist to localStorage
  useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark');
  }, [theme]);
  
  return (
    <div className="space-y-4">
      <h2>Appearance</h2>
      <RadioGroup value={theme} onChange={setTheme}>
        <RadioGroupOption value="light">Light</RadioGroupOption>
        <RadioGroupOption value="dark">Dark</RadioGroupOption>
        <RadioGroupOption value="system">System</RadioGroupOption>
      </RadioGroup>
    </div>
  );
}
```

#### Files to Modify
- `src/routes/_app.settings.appearance.tsx` — Implement toggle
- `src/lib/vyzorix-config.tsx` — Add theme persistence
- `tailwind.config.js` — Ensure dark mode class strategy

#### Dependencies
- Already using shadcn/ui RadioGroup component
- Tailwind dark mode already configured

---

## Task 5: OpenAPI Documentation 📚

### Why Important?
- API discoverability for developers
- Generate client SDKs automatically
- Interactive documentation with Swagger UI

### Implementation

#### Tools
- `swaggo/swag` for Go (generates from annotations)
- `swagger-ui` for frontend

#### Go Annotations Example
```go
// @Summary Login operator
// @Description Authenticate operator with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /v1/auth/login [post]
func (ac *AuthController) Login(c *gin.Context) { ... }
```

#### Files to Create/Modify
- All controller files — Add Swagger annotations
- `docs/swagger.yaml` — Generated spec
- `swagger-ui/` — Frontend documentation

#### Endpoints to Document
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/auth/login` | Operator login |
| POST | `/v1/auth/register` | Operator registration |
| POST | `/v1/auth/logout` | Operator logout |
| GET | `/v1/auth/me` | Get current operator |
| POST | `/v1/auth/forgot-password` | Request password reset |
| POST | `/v1/auth/reset-password` | Reset password |
| POST | `/v1/auth/verify-email` | Verify email |
| GET | `/v1/auth/google` | Google OAuth redirect |
| GET | `/v1/auth/google/callback` | Google OAuth callback |
| POST | `/v1/device/register` | Register device |
| GET | `/v1/device/:id/status` | Get device status |
| PATCH | `/v1/device/:id/fcm-token` | Update FCM token |
| DELETE | `/v1/device/:id` | Deregister device |
| GET | `/v1/dashboard/devices` | List all devices |
| GET | `/healthz` | Health check |

---

## Task 6: Database Migrations 📦

### Why Important?
- Safe schema evolution without data loss
- Rollback capability for failed migrations
- Version control for database schema

### Implementation

#### Approach
- Use `golang-migrate/migrate` library
- SQL-based migration files
- Versioned migration directory

#### Migration File Structure
```
migrations/
├── 000001_create_operators_table.up.sql
├── 000001_create_operators_table.down.sql
├── 000002_add_email_verified.up.sql
├── 000002_add_email_verified.down.sql
└── ...
```

#### Current Schema (from sqlite.go)
```sql
-- operators table
CREATE TABLE operators (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    password_hash TEXT,
    role TEXT NOT NULL DEFAULT 'operator',
    email_verified INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- auth_sessions table
CREATE TABLE auth_sessions (
    id TEXT PRIMARY KEY,
    operator_id TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (operator_id) REFERENCES operators(id) ON DELETE CASCADE
);

-- devices table
CREATE TABLE devices (
    id TEXT PRIMARY KEY,
    firebase_install_id TEXT NOT NULL,
    fcm_token TEXT,
    app_version TEXT,
    device_class TEXT,
    command_secret TEXT NOT NULL,
    command_secret_hash TEXT,
    online INTEGER NOT NULL DEFAULT 0,
    registered_at INTEGER NOT NULL,
    last_seen INTEGER NOT NULL
);

-- commands table
CREATE TABLE commands (
    dispatch_id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL,
    command TEXT NOT NULL,
    args TEXT,
    delivery TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    delivered_at INTEGER,
    status TEXT NOT NULL DEFAULT 'pending'
);
```

#### Files to Create/Modify
- `migrations/` — Migration files directory
- `storage/migrate.go` — Migration runner
- `main.go` — Run migrations on startup

#### Future Migration Ideas
| Version | Migration | Description |
|---------|-----------|-------------|
| 2.0 | `000003_add_refresh_tokens` | JWT refresh token support |
| 2.0 | `000004_add_audit_log` | Operator action audit trail |
| 2.1 | `000005_add_device_tags` | Custom device labels |
| 2.1 | `000006_add_command_history` | Full command execution history |

---

## Summary

| Task | Priority | Effort | Status |
|------|----------|--------|--------|
| Protected Routes | HIGH | Medium | Deferred |
| Rate Limit Auth | HIGH | Low | Pending |
| Dark Mode | MEDIUM | Low | Pending |
| OpenAPI Docs | MEDIUM | Medium | Pending |
| Database Migrations | MEDIUM | Medium | Pending |
| Playwright E2E | HIGH | High | Planned |

---

## Notes

- **Deferred tasks** will be revisited after server deployment
- **Playwright E2E** will use internal desktop PC tool for testing
- All changes should be backward compatible
- Tests required for all new functionality