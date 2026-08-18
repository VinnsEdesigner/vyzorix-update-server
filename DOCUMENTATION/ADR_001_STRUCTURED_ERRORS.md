# ADR-001: Structured Errors via Middleware, not Handler-level JSON

**Status**: Accepted  
**Date**: 2026-08-18  

## Context

The original error handling in Vyzorix had several problems:

1. **No consistency.** Every handler called `c.JSON(status, gin.H{"error": "bad_request", "message": "..."})` directly. The field names varied. Some used `"error"`, some used `"code"`. Some included a `"message"`, some didn't. Some added `"errors"` arrays for validation. There was no single shape.

2. **No trace IDs.** Error responses didn't include a correlation ID. When a client reported an error, the support team had no way to find the matching log line. The access log had `request_id` and the error log had nothing — they couldn't be joined.

3. **No error codes.** The `"error"` field was a free-form string like `"bad_request"` or `"unauthorized"` or `"internal_error"`. Clients had to string-match. There was no enum, no stability guarantee, no docs URL.

4. **Information leakage.** Some handlers logged `err.Error()` directly, which could contain database paths, internal state, or stack-trace-like details. The response messages were sometimes internal (e.g., "failed to marshal args") rather than client-safe.

5. **No validation details.** Validation errors returned a flat `"message"` string. The client couldn't tell which field failed or why. They'd have to parse the message text.

## Decision

We implemented a three-layer error system:

### Layer 1: Domain errors (`internal/domain/errors/`)

A typed `ServerError` struct with 55 canonical error codes across 10 categories. Each code knows its HTTP status, whether it's retryable, and its docs URL. Factory functions (`ErrNotFound`, `ErrForbidden`, `ErrValidationFailed`, etc.) make construction ergonomic. A `ValidationError` type carries field-level details for 400s.

### Layer 2: Error middleware (`internal/api/middleware/error.go`)

A single Gin middleware that runs after all handlers. It checks `c.Errors` — if a handler recorded an error and didn't write a response, the middleware formats and writes the structured envelope. It detects `*ValidationError` (renders 400 with details) and `*ServerError` (renders the error's own code/status/message). For plain errors, it derives the code from the HTTP status.

### Layer 3: Response helpers (`internal/api/responses/structured_error.go`)

For handlers and middleware that need to write the response immediately (gates, auth middleware, validation middleware), helper functions (`RespondStructured`, `RespondStructuredAbort`, `RespondValidationError`) produce the same envelope.

### Why middleware, not per-handler?

The alternative was to have every handler call `responses.RespondStructured(...)` directly. We tried that — it worked, but it meant 550+ call sites that all had to be consistent. If the envelope shape changed, every handler needed updating. And handlers that forgot to include a trace ID or docs URL would produce inconsistent responses.

With the middleware approach, a handler that does `c.Error(serverError); return` gets the same structured response as one that calls `RespondStructured` directly. The middleware is the single source of truth for the response shape. Handlers that use `c.Error` get the trace ID, docs URL, and code derivation for free.

The trade-off: handlers that use `c.Error` can't write a custom response body. They're limited to the `ServerError` fields (code, message, details, docs URL). This is intentional — consistency is more valuable than per-handler customization.

### Why not a global response wrapper?

We considered wrapping every response (success and error) in a `{"data": ...}` / `{"error": ...}` envelope. This is common in enterprise APIs. We rejected it because:

1. It would break every existing client integration — all success responses would change shape.
2. It adds no value for error responses (the `{"error": {...}}` shape is already standard).
3. It complicates streaming responses (WebSocket, Server-Sent Events) that don't fit the envelope.

The current approach gives structured errors without changing success response shapes.

## Consequences

- **All error responses have the same shape**: `{"error":{"code","message","trace_id","docs_url","details"?}}`.
- **Trace IDs are automatic** — the middleware stamps them from the gin context.
- **Error codes are stable** — clients can switch on them. They're documented in `ERROR_CODES.md`.
- **No `gin.H{"error":...}` calls remain** in the codebase. The AST migration tool converted all 550+ legacy sites.
- **Handlers can use either path**: `c.Error(serverError)` (middleware renders) or `RespondStructured(...)` (immediate write). Both produce the same envelope.
- **Adding a new error code** requires adding it to `codes.go` and using it in a handler. The middleware picks it up automatically.
