# Tracing & Log Correlation

Every HTTP request gets a trace ID. It shows up in the response headers, in the structured logs, and in audit entries. You can join all three with a single value.

## How the trace ID flows

1. The `Tracing()` middleware runs first (registered in `setupGlobalMiddleware`).
2. It reads the `X-Trace-ID` header from the request. If that's empty, it falls back to `X-Request-ID` (legacy alias). If both are empty, it generates a new 32-character hex ID.
3. It stores the ID in the gin context under `tracing.ContextKeyTraceID` (the string `"trace_id"`).
4. It sets `X-Trace-ID` on the response so the client can see it.
5. It also records the request start time for the `X-Response-Time` header.

## Where it shows up

**Response headers:**
```
X-Trace-Id: aabbccddeeff00112233445566778899
X-Response-Time: 1.2ms
```

**Access log (the `api_logger.go` middleware):**
```
level=INFO msg=http_request method=POST path=/v1/auth/login status=200 duration_ms=104 remote=127.0.0.1 trace_id=aabbccddeeff00112233445566778899
```

**Error responses:**
```json
{"error":{"code":"AUTH_TOKEN_INVALID","message":"...","trace_id":"aabbccddeeff00112233445566778899","docs_url":"..."}}
```

**Audit log entries (the `trace_id` column in `audit_logs`):**
```
action=login_success result=success trace=aabbccddeeff00112233445566778899 risk= actor_type=operator
```

**Error middleware log (when a handler records an error via `c.Error`):**
```
level=WARN msg=request_error trace_id=... error_code=AUTH_TOKEN_INVALID path=/v1/devices method=GET status=401
```

All of these use the same trace ID for the same request. If a client sends `X-Trace-ID: my-correlation-id`, that value flows through every layer.

## Getting the trace ID in handlers

```go
traceID := middleware.GetTraceID(c)
```

Returns the ID from the gin context. If for some reason the tracing middleware didn't run (shouldn't happen in production), it generates a new one.

## What was unified

Before this work, there were two separate correlation IDs:

- `X-Request-ID` / `request_id` — set by `RequestIDMiddleware` (now deleted)
- `X-Trace-ID` / `trace_id` — set by `Tracing` middleware

They were independent. A request's access log line showed `request_id` and its error log line showed `trace_id`, and they didn't match. Now there's one ID. The `X-Request-ID` header is still accepted as an inbound alias (for clients that send it), but only `X-Trace-ID` is emitted on responses.

## The fallback trace ID generator

If `crypto/rand` fails (essentially impossible on supported platforms), there's a fallback generator that mixes `time.Now().UnixNano()` with a process-wide atomic counter and hashes it with FNV-1a. It's tested for uniqueness across 1,000 rapid calls. In practice this path never runs.
