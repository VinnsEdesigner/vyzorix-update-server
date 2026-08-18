# Error System

Every API response that represents an error uses the same JSON shape. There are no exceptions to this. If you're writing a handler and you need to return an error, you have three options, and all three produce the identical response body.

## The response shape

```json
{
  "error": {
    "code": "AUTH_TOKEN_INVALID",
    "message": "Authentication required",
    "trace_id": "aabbccddeeff00112233445566778899",
    "docs_url": "https://docs.vyzorix.com/errors/AUTH_TOKEN_INVALID"
  }
}
```

The `code` field is a machine-readable string. Clients should switch on it, not the HTTP status or the message text. The `trace_id` matches the `X-Trace-ID` response header and the `trace_id` field in the server's structured logs. The `docs_url` is built dynamically from `VYZORIX_DOCS_BASE_URL` (defaults to `https://docs.vyzorix.com/errors`).

For validation errors, there's an extra `details` array:

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "One or more fields have validation errors",
    "details": [
      { "field": "deviceId", "message": "deviceId is required" },
      { "field": "command", "message": "command is required" }
    ],
    "trace_id": "...",
    "docs_url": "https://docs.vyzorix.com/errors/VALIDATION_FAILED"
  }
}
```

## The three ways to return an error

### 1. Record a ServerError (preferred for handlers)

```go
c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "Device not found"))
return
```

The error middleware picks this up after the handler returns and renders it. The `ServerError` carries its own code, message, and docs URL, so the middleware just formats it into the response. The HTTP status is derived from `Code.HTTPStatusCode()` — for `CodeResourceNotFound` that's 404.

This is the path most handlers use. You construct the error, record it, and return. The middleware does the rest.

### 2. Call RespondStructured directly (for gates and middleware)

```go
responses.RespondStructured(c, http.StatusForbidden, "You don't have permission to do this")
return
```

This writes the response immediately. Use it when you're inside a middleware or a gate function that needs to write and return right away — not at the top level of a handler. The code is derived from the HTTP status via `statusCodeToErrorCode` (403 → `AUTHZ_INSUFFICIENT_PERMISSIONS`, 404 → `RESOURCE_NOT_FOUND`, etc.).

For middleware that needs to abort the chain:

```go
responses.RespondStructuredAbort(c, http.StatusUnauthorized, "session cookie required")
```

This calls `RespondStructured` then `c.Abort()`.

### 3. Record a ValidationError (for field-level validation)

```go
details := []apperrors.ValidationDetail{
    apperrors.NewValidationDetail("email", "invalid format"),
    apperrors.NewValidationDetail("password", "too short"),
}
c.Error(apperrors.NewValidationError(details))
return
```

The error middleware detects `*ValidationError` via `errors.As` and renders a 400 with the `details` array filled in.

## Error codes

There are 55 error codes across 10 categories. They're all defined in `internal/domain/errors/codes.go` as `ErrorCode` string constants. They never change once published — clients depend on them.

| Category | Prefix | Example codes |
|----------|--------|---------------|
| Authentication | `AUTH_` | `AUTH_TOKEN_INVALID`, `AUTH_MFA_REQUIRED`, `AUTH_ACCOUNT_LOCKED` |
| Authorization | `AUTHZ_` | `AUTHZ_INSUFFICIENT_PERMISSIONS`, `AUTHZ_SCOPE_FORBIDDEN` |
| Resource | `RESOURCE_` | `RESOURCE_NOT_FOUND`, `RESOURCE_ALREADY_EXISTS`, `RESOURCE_CONFLICT` |
| Validation | `VALIDATION_` | `VALIDATION_FAILED`, `VALIDATION_REQUIRED_FIELD`, `VALIDATION_INVALID_EMAIL` |
| Rate limiting | `RATE_LIMIT_` | `RATE_LIMIT_EXCEEDED` |
| Security | `SECURITY_` | `SECURITY_THREAT_DETECTED`, `SECURITY_RISK_UNCONFIRMED` |
| Device | `DEVICE_` | `DEVICE_NOT_ONLINE`, `DEVICE_COMMAND_TIMEOUT` |
| Organization | `ORG_` | `ORG_NOT_FOUND`, `ORG_INVITATION_EXPIRED` |
| Update | `UPDATE_` | `UPDATE_NOT_FOUND`, `UPDATE_IN_PROGRESS` |
| Internal | `INTERNAL_` | `INTERNAL_SERVER_ERROR`, `INTERNAL_DATABASE_ERROR` |

Each code knows its HTTP status. `CodeResourceNotFound` returns 404. `CodeResourceAlreadyExists` returns 409. `CodeResourceLimitExceeded` returns 429. `CodeInternalTimeout` returns 504. This mapping lives in `HTTPStatusCode()` on the `ErrorCode` type.

## The error middleware

Registered globally in `setupGlobalMiddleware`. It runs after all handlers and middleware. If any `c.Error()` was recorded and the response hasn't been written yet, the middleware formats and writes the structured response.

It checks in order:
1. Is it a `*ValidationError`? → 400 with field details
2. Is it a `*ServerError`? → use the error's own code/status/message
3. Fall through to status-based mapping (derive code from `c.Writer.Status()`)

## Factory functions

For common error patterns, there are constructors in `internal/domain/errors/errors.go`:

```go
apperrors.ErrNotFound("Device", "device-123")           // RESOURCE_NOT_FOUND
apperrors.ErrAlreadyExists("API key")                    // RESOURCE_ALREADY_EXISTS
apperrors.ErrInvalidCredentials()                        // AUTH_INVALID_CREDENTIALS
apperrors.ErrForbidden()                                 // AUTHZ_INSUFFICIENT_PERMISSIONS
apperrors.ErrRateLimitExceeded(60)                        // RATE_LIMIT_EXCEEDED with retry_after
apperrors.ErrValidationFailed("email is required")        // VALIDATION_FAILED
apperrors.ErrInternal("database connection failed")       // INTERNAL_SERVER_ERROR
```

Each one sets the trace ID, timestamp, and docs URL automatically.

## What was removed

The old system had three things that are gone:

- `domain/domain_errors.go` — flat `errors.New("entity not found")` sentinels with no codes, no trace IDs. Deleted.
- `domain/threat/` — speculative threat-detection package with zero callers. Deleted.
- `api/responses/api_errors.go`, `api_responses.go`, `telemetry.go` — the old `{error, message}` response shape. Deleted.

Every error response in the server now goes through one of the three paths above. There are no `gin.H{"error": ...}` calls left anywhere in the codebase.
