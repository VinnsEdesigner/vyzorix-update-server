# Authentication & Sessions

The server uses cookie-based session auth for browser clients and API key auth for programmatic access. Both go through the same operator identity system.

## Login flow

`POST /v1/auth/login` with email + password. The handler (`internal/api/handlers/auth/auth_login.go`) does:

1. Validates the request (CSRF token required for browser clients)
2. Calls `AuthService.LoginWithDevice` which:
   - Looks up the operator by email
   - Verifies the password against the argon2id hash
   - Checks account lockout (5 failed attempts locks for a duration)
   - If MFA is enabled, returns `mfa_required: true` instead of a session
3. On success, creates a `Session` (with `MFAVerifiedAt` if MFA was completed)
4. Encrypts the session ID and sets it in the `vyz_session` cookie
5. Audits the login (`login_success` audit entry with trace_id and actor_type)
6. Returns the operator profile, organizations, and signing key

## Password hashing

Passwords are hashed with argon2id (OWASP 2023 recommended parameters):

- Memory: 64 MB
- Iterations: 3
- Parallelism: 4
- Salt: 16 bytes (random)
- Key: 32 bytes

The hash format is `$argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>` with raw base64 encoding (no padding).

The hasher in `internal/infrastructure/security/password/argon2_hasher.go` also implements a fake-hash timing attack defense — if the email doesn't exist, it still runs a hash verification against a dummy hash so the response time is indistinguishable from a wrong-password attempt.

## Session management

Sessions are stored in the `auth_sessions` table. Each session has:

- An encrypted ID (the cookie value is AES-GCM encrypted, not the raw session ID)
- Operator ID
- Expiration time (default 24 hours, configurable via `JWT_DURATION_HOURS`)
- `MFAVerifiedAt` — when MFA was completed during this session
- `SigningKey` — per-session HMAC key for request signing
- `SelectedOrganizationID` — the org the operator is currently working in

The `SessionManager` in `internal/infrastructure/security/session/session.go` handles encryption, decryption, and cookie creation.

## Cookie security

The `vyz_session` cookie has:
- `HttpOnly` — no JavaScript access
- `Secure` — HTTPS only
- `SameSite=Strict` — no cross-site requests

The CSRF cookie (`_csrf`) has the same flags.

## Token-based auth (for API clients)

`POST /v1/auth/login/tokens` returns JWT access + refresh tokens instead of a cookie. This is for non-browser clients (mobile apps, CLI tools, automation). The tokens are:

- Access token: short-lived JWT signed with `JWT_SECRET`
- Refresh token: long-lived, stored in `refresh_tokens` table, revocable

## API key auth

API keys are for service-to-service access. They're managed via `POST /v1/api-keys` (create), `GET /v1/api-keys` (list), `DELETE /v1/api-keys/:id` (revoke).

Each key has:
- A prefix (`vxyz-` by default) + random secret
- A scope (read, write, admin)
- An expiry date
- A per-month usage limit

API keys are stored as argon2id hashes (not plaintext). The `TenantAPIKeyAuth` middleware verifies the key, checks the scope, and sets the operator context.

## CSRF protection

The double-submit cookie pattern. The server generates a CSRF token, signs it with the session secret, and sets it in the `_csrf` cookie. The client reads the token from the cookie and sends it back in the `X-CSRF-Token` header. The `CSRFProtector` in `internal/api/middleware/csrf.go` verifies the header matches the cookie.

CSRF is required on all state-changing requests (POST, PATCH, DELETE) for browser-authenticated sessions. API key auth bypasses CSRF (it's not vulnerable to CSRF since API keys aren't sent automatically by browsers).

## MFA

TOTP-based. Operators enroll via `POST /v1/mfa/enroll` (generates a QR code), verify setup via `POST /v1/mfa/verify-setup`, and then must provide a TOTP code on login.

When MFA is enabled for an operator, login returns `mfa_required: true` with the operator ID. The client then calls `POST /v1/mfa/verify` with the TOTP code. On success, the session's `MFAVerifiedAt` is set.

Critical-tier commands (factory reset) require `MFAVerifiedAt` to be set on the session.

## OAuth

Google and GitHub OAuth are supported for "Sign in with..." flows. The OAuth flow:

1. Client redirects to Google/GitHub
2. OAuth provider redirects back to `/v1/auth/google/callback` or `/v1/auth/github/callback`
3. Server exchanges the auth code for user info
4. If the email matches an existing operator, creates a session
5. If not, creates a new operator (if registration is open)

OAuth state is stored in the `oauth_states` table to prevent CSRF during the OAuth flow.

## Account lockout

After 5 failed login attempts (configurable), the account is locked for a duration. The lockout is tracked in the `account_lockouts` table. The `Lockout` middleware in `internal/api/middleware/api_lockout.go` checks and records lockout state.

During lockout, login returns 429 with `retry_after` indicating how long until the account unlocks.
