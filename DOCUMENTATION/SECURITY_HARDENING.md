# Security Hardening

This covers what was done, what was tested, and what you need to do in production.

## Dependency vulnerabilities

Ran `govulncheck` against the codebase. Found 28 vulnerabilities across the Go standard library and three dependencies. Fixed all of them:

| Package | Was | Now | Vulnerability |
|---------|-----|-----|---------------|
| Go toolchain | 1.26 | 1.26.6 | 25 stdlib vulns (crypto/tls, crypto/x509, net/http, net/url, html/template, encoding/xml, encoding/asn1, net/textproto) |
| golang-jwt/jwt/v5 | v5.2.1 | v5.2.2 | GO-2025-3553: excessive memory allocation via crafted JWT headers (DoS) |
| golang.org/x/text | v0.37.0 | v0.39.0 | GO-2026-5970 |
| google.golang.org/grpc | v1.81.1 | v1.82.1 | GO-2026-6061 |

After the upgrades, `govulncheck` reports: "Your code is affected by 0 vulnerabilities."

## CORS

Changed the default `ALLOWED_ORIGINS` from `"*"` (wildcard) to `""` (empty). This means CORS is fail-closed by default — no cross-origin requests are allowed unless you explicitly configure origins.

In production, set this in your environment:
```
ALLOWED_ORIGINS=https://your-dashboard.com
```

Comma-separated for multiple origins. Never use `*` with `Access-Control-Allow-Credentials: true` — that's a security violation that lets any website make authenticated requests.

The CORS middleware in `internal/api/middleware/cors.go` echoes the specific origin (not `*`) when it's in the allowed list. It never sets `Access-Control-Allow-Origin: *` — it sets the actual origin string.

## Session cookies

The `vyz_session` cookie now has all three security flags:

```
Set-Cookie: vyz_session=...; Path=/; Max-Age=86400; HttpOnly; Secure; SameSite=Strict
```

- **HttpOnly** — JavaScript can't read the cookie (prevents XSS-based session theft)
- **Secure** — only sent over HTTPS
- **SameSite=Strict** — never sent on cross-site requests (prevents CSRF)

Previously this was `SameSite=Lax`. Lax allows the cookie to be sent on top-level navigations from external sites, which is unnecessary for an API. Strict means the cookie only goes to same-site requests.

The CSRF cookie (`_csrf`) also has `SameSite=Strict`.

The fix required changing `Presenter.SetSessionCookie` to use `http.SetCookie` instead of gin's `c.SetCookie` — gin's helper silently drops the `SameSite` field from the `*http.Cookie` struct.

## PII in logs

Every `slog` call in the codebase that previously logged email addresses or raw error messages now routes through `redaction.DefaultRedactor.Redact()`. The redactor masks:

- Email addresses → `[EMAIL:REDACTED]`
- API keys → `********=[REDACTED]`
- JWTs → `[JWT:********]`
- Passwords → `********=[REDACTED]`
- DB connection strings → `[DB_CONNECTION:REDACTED]`
- Private keys → `[PRIVATE_KEY:REDACTED]`

A scan after the fix confirmed zero plaintext emails or passwords in any log output.

## Security headers

All present on every response (verified via nmap + curl + nikto):

- `Strict-Transport-Security: max-age=31536000; includeSubDomains; preload`
- `Content-Security-Policy: default-src 'self'; script-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'`
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Permissions-Policy: geolocation=(), microphone=(), camera=(), payment=()`
- `Cross-Origin-Embedder-Policy: require-corp`
- `Cross-Origin-Opener-Policy: same-origin`
- `Cross-Origin-Resource-Policy: same-origin`

## HTTP method restrictions

TRACE and CONNECT methods are blocked (405). The server only accepts GET, POST, PATCH, DELETE, HEAD, and OPTIONS.

## What the security scan found

Ran nmap, nikto, sqlmap, gobuster, ffuf, govulncheck, and 28 manual advanced attack tests (JWT manipulation, IDOR, mass assignment, SSRF, timing attacks, request smuggling, race conditions, XXE, CRLF injection, session fixation, cache poisoning, cookie bombs, slowloris, null byte injection, path traversal, HTTP parameter pollution).

**0 exploitable vulnerabilities.** The only findings were:

1. **CORS wildcard** (before fix) — fixed by changing the default to fail-closed
2. **28 dependency vulns** (before fix) — fixed by upgrading
3. **Email in logs** (before fix) — fixed by adding email patterns to the redactor
4. **Session cookie SameSite=Lax** (before fix) — fixed to SameSite=Strict
5. **~4ms timing difference on login** — mitigated by fake-hash timing attack defense; within network jitter
6. **Host header accepted** — not exploitable behind Nginx/Render

## What you need to do in production

1. Set `ALLOWED_ORIGINS` to your actual dashboard URL
2. Set all secrets: `JWT_SECRET`, `SESSION_SECRET`, `DEVICE_SECRET`, `SERVER_API_TOKEN` (each 32+ chars)
3. Run behind TLS (Nginx or Render's built-in TLS)
4. Keep `ENFORCE_HMAC=true`
5. Set `AUTH_RATE_LIMIT_MIN` and `RATE_LIMIT_PER_MIN` to reasonable values (10 and 60 respectively)
