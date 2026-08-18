# Architecture Overview

## Stack

- **Language**: Go 1.26.6
- **Web framework**: Gin (v1.9.1)
- **Database**: SQLite (local) or Turso/libSQL (production)
- **GraphQL**: graphql-go
- **Realtime**: Gorilla WebSocket
- **Auth**: argon2id passwords, JWT tokens, AES-GCM session cookies, TOTP MFA
- **Push**: Firebase Cloud Messaging (FCM)
- **Dependency injection**: Google Wire (codegen, not reflection)
- **Frontend**: React + Tamagui (separate, served via SSR or static)

## Directory layout

```
apps/api/
  cmd/api/          -- entry point (api_main.go)
  internal/
    api/
      middleware/    -- Gin middleware (auth, CSRF, CORS, rate limit, tracing, etc.)
      handlers/      -- HTTP handlers (REST endpoints)
        command/     -- command execution
        auth/        -- login, register, MFA, API keys
        device/      -- device management
        confirmation/-- confirmation token endpoint
        updates/     -- update push, sync, history
        websocket/   -- WebSocket stream handler
        inbox/       -- device inbox
        dashboard/   -- dashboard stats
        diagnostics/ -- device diagnostics
        admin/       -- admin-only endpoints
        operator/    -- operator settings, notifications
        organization/-- org management, invitations
      responses/     -- structured error response helpers
      adapters/      -- response presenter, GraphQL adapters
      graphql/       -- GraphQL schema, resolvers, subscriptions
      wire/          -- dependency injection (Wire-generated)
      server_routes.go  -- route registration
      server_handlers.go -- HMAC handler, health checks
      api_server.go  -- Server struct, constructor, lifecycle
    application/
      auth/          -- AuthService (login, register, MFA, OAuth, sessions)
      command/       -- CommandService + Outbox worker
      device/        -- DeviceService (register, deregister, settings)
      confirmation/  -- ConfirmationService (issue/consume tokens)
      updates/       -- UpdateService (push, sync, versions)
      inbox/         -- Inbox service
      dashboard/     -- Dashboard stats service
      diagnostics/   -- Diagnostics service
      dto/           -- Data transfer objects (request/response types)
      shared/        -- ID generation, tokens
    domain/
      errors/        -- ErrorCode constants, ServerError, ValidationError
      command/       -- Command entity, risk catalog, risk evaluator
      confirmation/  -- PendingConfirmation model, Repository interface
      device/        -- Device entity, lifecycle, repository interface
      operator/      -- Operator entity, membership
      organization/  -- Organization entity, invitations
      session/       -- Session entity, revocation
      event/         -- Device event entity
      telemetry/     -- Telemetry frame types
      idempotency/   -- Idempotency record
    infrastructure/
      storage/       -- SQLite repositories + migrations (0XX_*.go)
      security/      -- JWT, sessions, passwords, rate limiting, TOTP, CSRF
      crypto/        -- HMAC command signing, token generation
      redaction/     -- Log secret redaction
      tracing/       -- Trace ID generation, docs URL builder
      fcm/           -- Firebase Cloud Messaging client
      email/         -- Resend email service
      config/        -- Environment config loading
      uuid/          -- UUID generation
      appcheck/      -- Google Play Integrity
      webhook/       -- Webhook client (SSRF protection)
      ssr/           -- Server-side rendering manager
      worker/        -- Background workers (device deletion, FCM retry)
    audit/           -- Audit logger, repository, entry types
    ws/              -- WebSocket hub, message queue, rate limiter
```

## Request lifecycle

```
HTTP request
  → Nginx (TLS, rate limit) [if VPS deployment]
  → Gin engine
  → PanicRecovery middleware
  → Tracing middleware (sets X-Trace-ID)
  → Logger middleware (logs request with trace_id)
  → CORS middleware
  → SecurityHeaders middleware
  → BodySizeLimit middleware
  → DisableTrace/DisableConnect
  → ErrorHandler middleware (deferred — runs after handler)
  → Idempotency middleware (on tenant routes)
  → SessionSignature middleware (HMAC verification)
  → CookieAuth / TenantAPIKeyAuth middleware
  → OrgContext + OrgMembership middleware
  → Route-specific middleware (CSRF, rate limiter, RBAC, etc.)
  → Handler
  ← Response (structured error if any c.Error was recorded)
  ← Logger middleware logs status + duration + trace_id
```

## Error flow

A handler that encounters an error does one of:

1. `c.Error(apperrors.NewServerError(code, msg))` then `return` — the ErrorHandler middleware renders it
2. `responses.RespondStructured(c, status, msg)` then `return` — writes immediately
3. `c.Error(apperrors.NewValidationError(details))` then `return` — renders 400 with field details

The ErrorHandler middleware (registered last, runs first in the deferred phase) checks `c.Errors` after all handlers return. If there's a recorded error and the response hasn't been written, it renders the structured response.

## Audit flow

Audit entries are written asynchronously via a buffered channel:

```
Handler calls Logger.CommandExecuted(ctx, event)
  → Entry queued to channel (buffer: 10,000)
  → Writer goroutine drains channel
  → Writes to audit_logs table (SQLite)
  → Optionally writes to separate audit DB
```

If the channel is full, the write falls back to synchronous (blocking) with a warning log.

## Dependency injection

Google Wire generates the dependency graph at build time. The providers in `internal/api/wire/providers.go` define how each component is constructed. `wire_gen.go` is the generated code that wires everything together.

The `Injector` function in `wire_gen.go` takes a `config.Config` and returns a fully wired `*Server` with all dependencies injected.

## Config

All configuration comes from environment variables, loaded in `internal/infrastructure/config/config.go`. The `Load()` function:

1. Reads env vars
2. Parses API keys from `API_KEY_*` vars
3. Validates required secrets (JWT_SECRET, SESSION_SECRET, API keys)
4. In production: validates secret lengths (32+ chars) and SERVER_API_TOKEN
5. Returns a `Config` struct

The `.env` / `.env.example` file is loaded by `godotenv` at startup (for local dev).
