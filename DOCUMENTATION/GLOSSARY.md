# Glossary

Terms used throughout the Vyzorix codebase and documentation. Organized by domain.

## Authentication & Security

**Argon2id** — The password hashing algorithm used for operator passwords and API key storage. Parameters: 64 MB memory, 3 iterations, 4 parallelism, 16-byte salt, 32-byte key. The hash format is `$argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>` with raw base64 encoding.

**API Key** — A service-to-service credential. Format: `vxyz-<prefix>-<secret>`. Stored as an argon2id hash (never plaintext). Scoped to an organization with a role (read, write, admin). Has monthly usage limits and optional expiry.

**ActorContext** — A value type passed to the `RiskEvaluator` carrying the caller's identity: `OperatorID`, `OrgID`, `IsSuperAdmin`, `MFAVerified`, `Confirmed`. The evaluator uses it to decide whether a command is allowed.

**Audit Entry** — A row in the `audit_logs` table recording a security-relevant event. Contains: who (operator, actor type, email), what (action, resource), when (timestamp), where (IP, user agent), result (success/failure/blocked), and context (trace ID, risk tier, old/new values).

**Confirmation Token** — A single-use, TTL-bounded authorization for a risky command. Issued by `POST /v1/devices/:imei/command/confirm`, consumed by the command execution handler. Stored in the `command_confirmations` table.

**CSRF Token** — A signed token used in the double-submit cookie pattern. Set in the `_csrf` cookie, sent back in the `X-CSRF-Token` header. Required on all state-changing requests from browser-authenticated sessions.

**Decision** — The output of the `RiskEvaluator`: `Allow`, `RequireConfirmation`, or `Deny`. Determines whether a command can execute immediately, needs a confirmation token, or is permanently blocked.

**HMAC** — Hash-based Message Authentication Code. Used for request signing (device→server and server→device) and WebSocket authentication. The server and device share a `CommandSecretHash` (SHA-256 of a random secret) and both compute the same HMAC key from it.

**HMAC Window** — The time window (default 30 seconds) during which a signed request is considered valid. Requests with timestamps older than the window are rejected. Prevents replay attacks.

**Idempotency Key** — A client-provided key (`X-Idempotency-Key` header) that makes a request idempotent. If the same key is sent twice, the second request returns the original response instead of creating a duplicate. Stored in the `idempotency_records` table with a 24-hour TTL.

**JWT** — JSON Web Token. Used for token-based auth (`POST /v1/auth/login/tokens`). Signed with `JWT_SECRET`. Contains operator ID, email, and expiration. Short-lived (access token) with a long-lived refresh token.

**MFA** — Multi-Factor Authentication. TOTP-based (RFC 6238). The operator scans a QR code during enrollment (`POST /v1/mfa/enroll`), then provides a 6-digit code on each login. The session's `MFAVerifiedAt` field records when MFA was completed.

**Nonce** — A unique value included in each HMAC-signed request. The `ReplayProtectionMiddleware` tracks nonces and rejects duplicates within the HMAC window. Prevents replay attacks.

**OAuth** — Sign-in with Google or GitHub. The OAuth flow exchanges an authorization code for user info (email, name). If the email matches an existing operator, a session is created. State is stored in the `oauth_states` table to prevent CSRF during the redirect flow.

**Operator** — A person who can log in to the dashboard. Identified by email + password (or OAuth). Global across organizations — one operator can belong to multiple orgs with different roles.

**Session** — An authenticated operator session stored in the `auth_sessions` table. The cookie value (`vyz_session`) is AES-GCM encrypted — the cookie contains the encrypted session ID, not the raw ID. Sessions expire after `JWT_DURATION_HOURS` (default 24).

**SessionSecret** — The AES-256 key used to encrypt session cookie values. Set via `SESSION_SECRET` env var (must be 32+ characters in production).

**Signing Key** — A per-session HMAC key returned on login. The browser client uses it to sign every subsequent request. The server verifies the signature via `SessionSignatureMiddleware`.

**Turnstile** — Cloudflare's CAPTCHA alternative. Required on public endpoints (login, register, password reset) to prevent bot abuse. The token is verified server-side via Cloudflare's API.

## Devices & Commands

**Client** — A WebSocket connection representing either a device or a dashboard client. Has a `Send` channel for outbound messages and a `DeviceID` identifying which device it represents.

**Command** — An instruction sent from the server to a device. Has a type (`CHECK_UPDATE`, `WAKE_UP_UPDATER`, `device.reboot`, etc.), arguments, a dispatch ID (for idempotency), and a status lifecycle (pending → delivered → completed/failed/cancelled).

**CommandFrame** — The JSON wire format for a command sent over WebSocket. Contains: type, command name, dispatch ID, args, timestamp, nonce, and HMAC signature.

**CommandSecretHash** — A SHA-256 hash of a random secret, stored on the device entity. Both the server and the device can compute the same HMAC key from this hash without the server storing the plaintext secret. Used for command signing.

**Device** — An Android phone running the Vyzorix agent. Has a unique ID (IMEI), FCM token, model, OS version, and belongs to an organization.

**Device Lifecycle** — The state machine for device existence: `pending` (awaiting approval) → `registered` (active) → `deregistered` (soft-deleted, 30-day retention). Transitions are validated via `CanTransitionTo`.

**Dispatch ID** — A unique identifier for a command instance. Used for idempotency — if a client sends the same dispatch ID twice, the second request returns the original response.

**Hub** — The central WebSocket connection manager in `internal/ws/hub.go`. Tracks which devices are online, routes messages, manages the message queue, and broadcasts to dashboards.

**Outbox** — A background worker that polls for pending commands and attempts delivery. Runs every 1 second, with exponential backoff retries (max 5). Handles the case where a command was created but the device was offline.

**Risk Tier** — Classification of a command's danger level: `zero` (read), `low` (reversible), `medium` (default), `high` (org-wide), `critical` (irreversible). Determines whether confirmation/MFA is required.

**Telemetry** — Data sent from devices to the server: battery level, CPU usage, memory, temperature, network type, app version. Stored in `device_telemetry` and broadcast to dashboard clients in realtime.

**TelemetryFilter** — A component on the Hub that deduplicates and rate-limits telemetry frames before storage. Prevents a device from flooding the database with redundant readings.

## Architecture & Infrastructure

**App Check** — Google Play Integrity API verification for Android device attestation. Verifies that the device is genuine and the app is unmodified. Configured via `FIREBASE_CREDENTIALS`.

**Config** — The `Config` struct loaded from environment variables in `internal/infrastructure/config/config.go`. All runtime configuration comes from here — no config files.

**DashboardBroadcaster** — An interface on the Hub that forwards device events (telemetry, status changes) to connected dashboard WebSocket clients.

**EventProcessor** — An interface on the Hub that processes device events (telemetry storage, threshold checking, event generation). Set during server initialization.

**FCM** — Firebase Cloud Messaging. Used to send silent push notifications to offline devices, telling them to wake up and connect to receive pending commands. Has a circuit breaker that stops delivery after repeated failures.

**govulncheck** — Go's official vulnerability scanner. Checks the codebase against the Go vulnerability database. Run with `govulncheck ./...`. After fixes, reports 0 affected vulnerabilities.

**IdempotencyRecord** — A stored response for a request with an idempotency key. If the same key is seen again, the stored response is returned. Has a TTL (default 24 hours) and is stored in the `idempotency_records` table.

**MessageQueue** — Two-tier storage for messages to offline devices: in-memory channels for low latency, SQLite for durability. Max 1000 messages per device, 7-day TTL. Replayed when the device reconnects.

**Migration** — A database schema change function. Each migration runs inside a transaction, has a version number, and is recorded in the `schema_migrations` table. 59 migrations as of this writing.

**RateLimiter** (WebSocket) — Per-device message rate limiting on the Hub. Default: 10 messages/second, burst 20. Messages exceeding the rate are dropped.

**RateLimiter** (HTTP) — Per-IP or per-email HTTP rate limiting. Auth routes: `AUTH_RATE_LIMIT_MIN` (default 10/min). General routes: `RATE_LIMIT_PER_MIN` (default 60/min). Returns 429 with `retry_after_seconds`.

**Redactor** — The log sanitization system in `internal/infrastructure/redaction/redact.go`. Masks API keys, JWTs, passwords, DB connection strings, private keys, generic credentials, and email addresses in log output.

**ServerError** — The structured error type in `internal/domain/errors/errors.go`. Carries: code, message, details, trace ID, timestamp, docs URL, and internal error context. Implements the `error` interface.

**SSR** — Server-Side Rendering. The server can render the React frontend server-side (via a Node.js subprocess) and serve the HTML. Configurable via `SSR_ENABLE` and `SSR_AUTO_BUILD`. Falls back to static file serving if SSR is unavailable.

**Trace ID** — A 32-character hex string correlating a single HTTP request across response headers, access logs, error responses, and audit entries. Set by the `Tracing` middleware from `X-Trace-ID` (or `X-Request-ID` as alias, or generated).

**ValidationError** — A structured validation error carrying field-level details (`[]ValidationDetail`). Detected by the error middleware via `errors.As` and rendered as a 400 with the `details` array.

**Wire** — Google's dependency injection code generator. Providers in `internal/api/wire/providers.go` define how each component is constructed. `wire_gen.go` is the generated code that wires everything together at build time (no reflection).

## Organizations

**Organization** — A tenant in the multi-tenant system. Owns devices, updates, API keys. Has members (operators with roles), settings (webhook, password policy), and a member limit.

**Invitation** — The flow for adding an operator to an organization. An admin creates an invitation with the invitee's email and a role. The invitee receives an email and accepts. Invitations expire (default 7 days) and can be revoked.

**Membership** — A row in `organization_members` linking an operator to an organization with a role (`super_admin`, `admin`, `operator`, `viewer`) and a status (`active`, `removed`).

**Role** — The permission level of an operator within an organization. Enforced by the `RBACAuthorize` middleware and the `RequireSuperAdmin` middleware. CHECK-constrained in the database.

**Scope** — The permission level of an API key within an organization: `read`, `write`, `admin`. Enforced by the `ScopeEnforcement` middleware. Maps to HTTP methods.

## Deployment

**Nginx** — The recommended TLS-terminating reverse proxy for VPS deployments. Handles TLS, rate limiting, HTTP/2, and WebSocket proxying. Config in `tooling/nginx/vyzorix.conf`. Not needed on Render (which handles TLS internally).

**Render** — The managed hosting platform configured in `render.yaml`. Handles TLS, health checks, zero-downtime deploys, and persistent disks. No Nginx needed.

**Turso** — A managed libSQL (SQLite-compatible) database service. Used in production when `DATABASE_BACKEND=turso`. Provides distributed SQLite with a primary/replica model. Credentials via `TURSO_DB_URL` and `TURSO_AUTH_TOKEN`.

**WAL Mode** — Write-Ahead Logging mode in SQLite. Enables concurrent readers and a single writer. Required for transactional DDL (the migration fix). Configured via `?_journal_mode=WAL` in the connection string.
