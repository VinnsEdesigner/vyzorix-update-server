# Log Redaction

Logs should never contain passwords, API keys, JWTs, or email addresses. The redaction system makes sure of that.

## How it works

There's a `Redactor` in `internal/infrastructure/redaction/redact.go`. It takes a string and returns a sanitized version. The default instance is `redaction.DefaultRedactor`, which is configured with all pattern types enabled.

The `Redact` method runs seven pattern groups in order of sensitivity:

1. **Private keys** — `-----BEGIN PRIVATE KEY-----` markers
2. **DB connection strings** — `postgres://user:pass@host` patterns
3. **JWTs** — `eyJ...eyJ...` patterns
4. **API keys** — `api_key=...`, `apikey=...`, `secret=...`, `bearer ...` patterns
5. **Passwords** — `password=...`, `passwd=...`, `pwd=...` patterns
6. **Generic credentials** — `credential=...`, `access_key_id=...`, `aws_secret_access_key=...`
7. **Emails** — `user@domain.com` patterns

Each pattern replaces the matched text with a redaction marker (`[REDACTED]`, `[JWT:********]`, `[EMAIL:REDACTED]`, etc.).

## Where it's used

### In the error middleware

When a handler records an error via `c.Error`, the error middleware logs it. The error message goes through `redaction.DefaultRedactor.Redact()` before hitting `slog`. This means if the error message happens to contain a password or email, it gets masked.

### In structured error responses

The `logInternalError` function in `responses/structured_error.go` redacts the original error and the internal context map before logging.

### In handler log calls

Every `slog` call in the auth handlers that previously logged `req.Email` or `err.Error()` now routes through `redaction.DefaultRedactor.Redact()`. Specifically:

- `auth_login.go` — login error log, login notification email log
- `auth_register.go` — verification email failure log, email send error log
- `auth_email_verification.go` — email verification resent log
- `ssr-proxy.go` — SSR access granted log (JWT claims email)

### In `RedactStruct`

The `RedactStruct` method recursively walks a `map[string]any` or `[]any` and redacts every string value. Used for internal context maps that get logged.

## What it catches

```
Input:  "api_key=sk_test_1234567890abcdef"
Output: "********=[REDACTED]"

Input:  "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
Output: "[JWT:********]"

Input:  "password=SuperSecret123!"
Output: "********=[REDACTED]"

Input:  "admin@vyzorix.test"
Output: "[EMAIL:REDACTED]"

Input:  "postgres://user:pass@localhost:5432/db"
Output: "[DB_CONNECTION:REDACTED]"

Input:  "-----BEGIN RSA PRIVATE KEY-----"
Output: "[PRIVATE_KEY:REDACTED]"
```

## What was fixed

The original implementation had a `SanitizeForLog` function that was supposed to sanitize log entries, but it had a bug: it only copied a fixed allowlist of fields (`trace_id`, `level`, `message`, `user_id`, `org_id`, `action`, `duration`, `status`, `metadata`). Fields like `path`, `method`, `error_message`, and `actor_id` were silently dropped. The function was deleted entirely, and all call sites were rewritten to log fields directly via `slog.String(...)` / `slog.Any(...)` with redaction applied only to the error message (the only field that could contain sensitive data).

The email pattern was added after discovering that `admin@vyzorix.test` was appearing in log output. The existing patterns only caught `password=`, `api_key=`, `bearer ` — none of which match a bare email address.
