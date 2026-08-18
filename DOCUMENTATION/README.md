# Document Index

## Hardening & Security

- [Error System](ERROR_SYSTEM.md) — The structured error response shape, error codes, the three ways handlers return errors, and what was removed.
- [Error Codes](ERROR_CODES.md) — Every error code explained: what it means, when it fires, HTTP status, retryability.
- [Tracing](TRACING.md) — How the X-Trace-ID correlation ID flows through response headers, logs, and audit entries.
- [Log Redaction](LOG_REDACTION.md) — How passwords, API keys, JWTs, and emails get masked in log output.
- [Security Hardening](SECURITY_HARDENING.md) — Dependency vulns fixed, CORS fix, cookie flags, PII redaction, security headers, scan results.

## Risk & Audit

- [Risk & Confirmation](RISK_AND_CONFIRMATION.md) — Risk tiers, the evaluator, confirmation token flow, MFA requirements.
- [Audit System](AUDIT_SYSTEM.md) — Audit entry fields, what gets audited, the buffered channel writer, command execution audit trail.

## Database

- [Database Migrations](DATABASE_MIGRATIONS.md) — How migrations run in transactions, the table-rebuild pattern, the bricking fix.

## Core Features

- [Authentication](AUTHENTICATION.md) — Login, sessions, passwords, MFA, OAuth, CSRF, API keys, account lockout.
- [Organizations](ORGANIZATIONS.md) — Multi-tenancy model, roles, membership, invitations.
- [Device Management](DEVICE_MANAGEMENT.md) — Registration, status, lifecycle, settings, events.
- [Command Execution](COMMAND_EXECUTION.md) — Command lifecycle, signing, delivery (WebSocket/FCM), outbox, retries.
- [Realtime Communication](REALTIME_COMMUNICATION.md) — WebSocket hub, message routing, dashboard broadcasting, telemetry.
- [WebSocket System](WEBSOCKET_SYSTEM.md) — Deep dive: Hub internals, clients, message queue, rate limiting, compression, latency tracking.
- [Update System](UPDATE_SYSTEM.md) — APK versions, GitHub webhook sync, push delivery, status tracking.
- [API Keys](API_KEYS.md) — Key structure, scope enforcement, rate limiting, management endpoints.
- [GraphQL API](GRAPHQL_API.md) — Endpoint, schema, resolvers, subscriptions.
- [Background Workers](BACKGROUND_WORKERS.md) — Device deletion, FCM retry, command outbox.

## Architecture Decisions

- [ADR-001: Structured Errors via Middleware, not Handler-level JSON](ADR_001_STRUCTURED_ERRORS.md) — Why we moved error rendering out of handlers and into a central middleware.
- [ADR-002: Transactional DDL Migrations in SQLite](ADR_002_TRANSACTIONAL_MIGRATIONS.md) — Why we wrap migrations in transactions and why it prevents the bricking bug.

## Operations

- [Deployment](DEPLOYMENT.md) — Render, Docker Compose + Nginx, bare metal. Env vars reference.
- [Architecture](ARCHITECTURE.md) — Stack, directory layout, request lifecycle, error flow, audit flow, DI.

## Reference

- [Glossary](GLOSSARY.md) — Every term used in the codebase and docs, organized by domain.
