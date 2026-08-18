# Error Code Reference

Every error code the server can return, what it means, when it fires, and what the HTTP status is. Clients should switch on the `code` string, not the HTTP status or the message text.

## Authentication codes (AUTH_*)

All return HTTP 401.

### AUTH_INVALID_CREDENTIALS

The email or password is wrong. The response doesn't say which one — that's intentional, to prevent user enumeration. The login handler runs a fake argon2id hash verification even when the email doesn't exist, so the response time is indistinguishable from a wrong-password attempt.

### AUTH_ACCOUNT_LOCKED

Too many failed login attempts. The account is temporarily locked. The `details` field includes `retry_after_minutes` — how long until the lock expires. The lockout threshold is 5 failed attempts (configurable). Lockout state is tracked in the `account_lockouts` table.

### AUTH_MFA_REQUIRED

The operator has MFA enabled, and the login request didn't include a TOTP code. The response includes `operator_id` so the client knows which account is being challenged. The client should prompt for a TOTP code and call `POST /v1/mfa/verify`.

### AUTH_MFA_INVALID

The TOTP code is wrong or expired. TOTP codes are time-based (30-second windows) and the server allows a small clock skew. If the code doesn't match the current or adjacent windows, this error is returned.

### AUTH_SESSION_EXPIRED

The session cookie exists but the session has passed its expiration time. The client should redirect to login. Default session lifetime is 24 hours (configurable via `JWT_DURATION_HOURS`).

### AUTH_SESSION_REVOKED

The session was explicitly revoked — either the operator logged out from another device, or an admin revoked it. The `auth_sessions` table tracks revocation. The `AuthRevocationMiddleware` checks a revocation list on every authenticated request.

### AUTH_TOKEN_INVALID

The most common auth error. It covers several cases:

- No session cookie present (the `CookieAuth` middleware didn't find `vyz_session`)
- The session cookie exists but can't be decrypted (AES-GCM decryption failed — tampered or wrong key)
- The session ID was found but the session doesn't exist in the database (deleted)
- An API key was provided but is malformed or doesn't match any stored key
- The request is missing required HMAC signature headers on device-facing endpoints

This is the code returned when you hit any authenticated endpoint without a valid session.

### AUTH_TOKEN_EXPIRED

A JWT token (from `POST /v1/auth/login/tokens`) has passed its `exp` claim. Different from `AUTH_SESSION_EXPIRED` which is about cookie sessions, not JWT tokens.

### AUTH_ORIGIN_FORBIDDEN

The request's `Origin` header doesn't match any configured `ALLOWED_ORIGINS`. This is the CORS preflight rejection — the server refuses to send `Access-Control-Allow-Origin` for untrusted origins.

### AUTH_CSRF_INVALID

A state-changing request (POST, PATCH, DELETE) was made without a valid CSRF token. The `X-CSRF-Token` header is either missing or doesn't match the `_csrf` cookie. API key-authenticated requests bypass CSRF (they're not vulnerable since API keys aren't sent automatically by browsers).

## Authorization codes (AUTHZ_*)

All return HTTP 403.

### AUTHZ_INSUFFICIENT_PERMISSIONS

The operator is authenticated but doesn't have the required role for the operation. For example, an `operator`-role user trying to create an API key (requires `admin`), or a `viewer` trying to execute a command (requires `operator`).

### AUTHZ_ORG_MEMBERSHIP_REQUIRED

The operator isn't a member of the organization they're trying to access. The `NewOrganizationMembership` middleware checks the `organization_members` table and returns this if no active membership exists.

### AUTHZ_ORG_OWNERSHIP_REQUIRED

The operation requires being the organization owner (super_admin). Examples: deleting the organization, transferring devices between orgs.

### AUTHZ_SCOPE_FORBIDDEN

An API key was used but its scope doesn't allow the operation. A `read`-scope key trying to POST, or a `write`-scope key trying to DELETE. The scope-to-method mapping: read → GET/HEAD, write → GET/POST/PATCH/HEAD, admin → all methods.

### AUTHZ_DEVICE_ACCESS_DENIED

The device exists but belongs to a different organization. The operator is authenticated and is an org member, but the device they're trying to access isn't in their org. The `verifyDeviceInOrganization` check in the command handler catches this.

## Resource codes (RESOURCE_*)

### RESOURCE_NOT_FOUND (404)

The requested resource doesn't exist. Could be a device, command, API key, update version, invitation — anything identified by an ID that wasn't found in the database.

### RESOURCE_ALREADY_EXISTS (409)

A create operation found an existing resource with the same unique identifier. For example, registering a device with an IMEI that's already registered, or creating an API key with a duplicate name.

### RESOURCE_CONFLICT (409)

A state conflict prevents the operation. The resource exists but its current state doesn't allow the action. For example, trying to cancel a command that's already completed, or pushing an update when another push is in progress.

### RESOURCE_DELETED (404)

The resource was previously soft-deleted. Different from `RESOURCE_NOT_FOUND` (never existed) — this means it existed but was deleted. Returns 404 (not 410 Gone) because the client should treat it the same way: the resource is gone.

### RESOURCE_LIMIT_EXCEEDED (429)

A quota or limit was reached. For example, an organization has hit its `max_members` limit, or the pending commands per device limit (50) is reached. Returns 429 because it's a rate/quota issue, not a resource existence issue.

## Validation codes (VALIDATION_*)

All return HTTP 400.

### VALIDATION_FAILED

General validation failure. The `details` array contains field-level information. This is the code the `ValidationMiddleware` returns when a request body fails schema validation (e.g., missing required fields, wrong format).

### VALIDATION_REQUIRED_FIELD

A specific required field is missing or empty. The `details` array includes the field name. Used by factory functions like `ErrRequiredField("email")`.

### VALIDATION_INVALID_FORMAT

A field value doesn't match the expected format. For example, an email that doesn't match the RFC 5322 simplified pattern, or a dispatch ID that doesn't match the expected pattern.

### VALIDATION_INVALID_EMAIL

The email address is malformed. Separate from `VALIDATION_INVALID_FORMAT` because email validation is common enough to warrant its own code, and clients may want to show a specific "invalid email" message.

### VALIDATION_INVALID_PASSWORD

The password doesn't meet the policy requirements. The password policy checks: minimum length (8), at least one uppercase, one lowercase, one digit. The response includes which requirement wasn't met.

### VALIDATION_PASSWORD_BREACHED

The password was found in a known data breach database. The server checks passwords against a breach list (HIBP-style). If the password has been seen in a breach, registration is rejected.

### VALIDATION_TOO_LONG

A field value exceeds the maximum allowed length. For example, a command name longer than 64 characters, or a device name longer than 255 characters.

### VALIDATION_TOO_SHORT

A field value is below the minimum allowed length. For example, a password shorter than 8 characters.

### VALIDATION_INVALID_ENUM

A field value is not one of the allowed options. For example, an install type that isn't `immediate`, `scheduled`, or `deferred`. Or an organization role that isn't `super_admin`, `admin`, `operator`, or `viewer`.

## Rate limiting codes (RATE_LIMIT_*)

All return HTTP 429. These are retryable.

### RATE_LIMIT_EXCEEDED

Too many requests in the time window. The rate limiter tracks requests per IP (general) or per email (auth). The `details` field includes `retry_after_seconds`. Rate limits are configurable via `RATE_LIMIT_PER_MIN` (general) and `AUTH_RATE_LIMIT_MIN` (auth).

### RATE_LIMIT_RETRY_AFTER

Similar to `RATE_LIMIT_EXCEEDED` but explicitly tells the client to retry after a specific duration. Used by the account lockout system — the account is locked and the client should wait before retrying.

## Security codes (SECURITY_*)

### SECURITY_THREAT_DETECTED (403)

Suspicious activity was detected and blocked. This could be impossible travel (login from two distant IPs within 30 minutes), brute force (5+ failed attempts), or a suspicious device fingerprint.

### SECURITY_RISK_UNCONFIRMED (449)

A high-risk or critical command was attempted without a valid confirmation token. The HTTP status 449 ("Retry With") signals that the client should get a confirmation token and retry. The `details` field includes the `risk_tier` and `operation` name. See [Risk & Confirmation](RISK_AND_CONFIRMATION.md).

### SECURITY_IP_BLOCKED (403)

The requesting IP address is on the blocked list. IPs can be blocked by the threat detection system or by an admin. The `IPIntelligence` middleware checks the block list.

### SECURITY_DEVICE_FINGERPRINT_BLOCKED (403)

The device fingerprint (computed from IP, user agent, accept headers) is on the blocked list. This is a more persistent block than IP — it survives IP changes.

### SECURITY_REPLAY_DETECTED (403)

A replay attack was detected. The `ReplayProtectionMiddleware` tracks nonce+timestamp pairs and rejects duplicate nonces. Each request must have a unique nonce within the HMAC window.

### SECURITY_SIGNATURE_INVALID (403)

The HMAC request signature verification failed. The request was signed with the wrong key, or the signature doesn't match the payload. This means either the device's `CommandSecretHash` is wrong, or the request was tampered with in transit.

### SECURITY_TURNSTILE_FAILED (403)

Cloudflare Turnstile (CAPTCHA) verification failed. Required on public endpoints (login, register, password reset) to prevent bot abuse. The Turnstile token is verified server-side via the Cloudflare API.

## Device codes (DEVICE_*)

### DEVICE_NOT_ONLINE (400)

The device is not connected via WebSocket and can't receive commands. The command was created (pending status) but couldn't be delivered immediately. The outbox or FCM will attempt delivery later.

### DEVICE_COMMAND_PENDING (400)

A command is already pending on this device. The server limits pending commands per device to 50 (`MaxPendingCommandsPerDevice`). If the limit is reached, new commands are rejected until existing ones complete or expire.

### DEVICE_COMMAND_TIMEOUT (504)

The device didn't respond to the command within the timeout. The command's `ExpiresAt` field determines the timeout. The outbox marks the command as failed after max retries are exhausted. This is retryable.

### DEVICE_COMMAND_FAILED (400)

The device received and executed the command but reported a failure. The `FailureReason` field on the command record contains the device's error message. Not retryable automatically — the operator must decide whether to retry.

### DEVICE_NOT_REGISTERED (400)

The device ID wasn't found in the registry. The device needs to register first via `POST /v1/device/register`. Different from `RESOURCE_NOT_FOUND` — this is specific to the device domain.

### DEVICE_ALREADY_REGISTERED (400)

A device with this ID is already registered. Registration is idempotent if the same operator re-registers, but a different operator trying to register an existing device gets this error.

## Organization codes (ORG_*)

### ORG_NOT_FOUND (404)

The organization ID wasn't found. The org was deleted or never existed.

### ORG_MEMBER_LIMIT (400)

The organization has reached its `max_members` limit. The default is 2 (configurable per org). New invitations are rejected until the limit is increased or members are removed.

### ORG_INVITATION_EXPIRED (400)

The invitation has passed its TTL. The invitee needs to request a new invitation. Default invitation TTL is 7 days.

### ORG_INVITATION_INVALID (400)

The invitation code is invalid — either it was already used, revoked, or doesn't exist.

## Update codes (UPDATE_*)

### UPDATE_NOT_FOUND (404)

The update version ID wasn't found. The version was deleted or never existed.

### UPDATE_IN_PROGRESS (409)

Another update push is currently in progress for this device or organization. Only one push can be active at a time. The client should wait for the current push to complete before initiating a new one.

### UPDATE_DEVICE_INCOMPATIBLE (409)

The device is not compatible with this update version. The compatibility check may consider OS version, device model, or app version. The `details` field includes the incompatibility reason.

## Internal codes (INTERNAL_*)

### INTERNAL_SERVER_ERROR (500)

An unexpected error occurred. This is the catch-all for unhandled errors — panics, nil pointer dereferences, unexpected type assertions. The response message is generic ("An unexpected error occurred") and the stack trace is logged internally with the trace ID. Retryable.

### INTERNAL_DATABASE_ERROR (500)

A database operation failed. This could be a connection error, a constraint violation that wasn't expected, or a query syntax error. The specific error is logged with the trace ID. Retryable.

### INTERNAL_EXTERNAL_SERVICE_ERROR (500)

An external service call failed — FCM (Firebase), Resend (email), Google OAuth, GitHub API, Turso. The error is logged with the service name and trace ID. Retryable.

### INTERNAL_TIMEOUT (504)

An operation timed out. This could be a database query timeout, an HTTP client timeout, or a command execution timeout. Different from `DEVICE_COMMAND_TIMEOUT` which is specifically about device response — this is about the server's own operations. Retryable.

## Retryability

The `IsRetryable()` method on `ErrorCode` returns `true` for:

- `RATE_LIMIT_EXCEEDED`, `RATE_LIMIT_RETRY_AFTER`
- `INTERNAL_SERVER_ERROR`
- `INTERNAL_TIMEOUT`
- `INTERNAL_EXTERNAL_SERVICE_ERROR`
- `DEVICE_COMMAND_TIMEOUT`

Clients should implement exponential backoff when retrying these codes. All other codes are not retryable — retrying will produce the same result.
