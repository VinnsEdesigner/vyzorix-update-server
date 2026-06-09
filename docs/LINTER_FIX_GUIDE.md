# Vyzorix Linter Issues Mapping - 100% Pass Strategy

> **Document Version:** 1.0  
> **Status:** Active  
> **Purpose:** Map all linter issues to specific files/lines for targeted fixes

---

## Summary of Issues by Linter

| Linter | Count | Issue Type |
|--------|-------|------------|
| stylecheck | 35 | Package comments, error formatting |
| godot | 32 | Comments not ending in period |
| gofumpt | 12 | File formatting issues |
| gci | 8 | Import ordering issues |
| perfsprint | 10 | fmt.Errorf instead of errors.New |
| nilnil | 7 | Returns nil error with invalid value |
| nestif | 3 | Complex nested if blocks |
| nilerr | 2 | Error not nil but returns nil |
| noctx | 1 | HTTP call without context |
| predeclared | 1 | Variable name shadowing |

**Total Issues: 111**

---

## File-by-File Fix Guide

### 1. internal/auth/google_token.go

| Line | Issue | Fix |
|------|-------|-----|
| 26 | godot - Comment should end in period | Change `// Google JWKS endpoint` to `// Google JWKS endpoint.` |
| 29 | godot - Comment should end in period | Change `// Google valid issuers` to `// Google valid issuers.` |
| 153 | noctx - (*net/http.Client).Get must not be called | Use `http.NewRequestWithContext` instead |
| 21 | stylecheck - ST1005: error strings should not be capitalized | `ErrGoogleTokenExpired` to `errGoogleTokenExpired` |
| 22 | stylecheck - error strings should not be capitalized | `ErrGoogleTokenBadIssuer` to `errGoogleTokenBadIssuer` |
| 23 | stylecheck - error strings should not be capitalized | `ErrGoogleTokenBadAudience` to `errGoogleTokenBadAudience` |
| 1 | stylecheck - ST1000: package needs comment | Add `// Package security provides authentication utilities.` |

### 2. internal/auth/validate.go

| Line | Issue | Fix |
|------|-------|-----|
| 9 | godot | Change `// Max lengths for common fields` to `// Max lengths for common fields.` |
| 19 | godot | Change `// Email regex pattern (RFC 5322 simplified)` to `// Email regex pattern (RFC 5322 simplified).` |
| 22 | godot | Change `// ValidationError represents a validation error` to `// ValidationError represents a validation error.` |
| 32 | godot | Change `// ValidateEmail validates and sanitizes an email address` to `// ValidateEmail validates and sanitizes an email address.` |
| 49 | godot | Change `// ValidateName validates and sanitizes a name` to `// ValidateName validates and sanitizes a name.` |
| 63 | godot | Change `// ValidatePasswordLength validates password length constraints` to `// ValidatePasswordLength validates password length constraints.` |
| 74 | godot | Change `// ValidateDeviceID validates and sanitizes a device ID` to `// ValidateDeviceID validates and sanitizes a device ID.` |
| 94 | godot | Change `// ValidateCommand validates a command string` to `// ValidateCommand validates a command string.` |
| 108 | godot | Change `// ValidateToken validates a token string` to `// ValidateToken validates a token string.` |
| 122 | godot | Change `// SanitizeString removes potentially dangerous characters` to `// SanitizeString removes potentially dangerous characters.` |

### 3. internal/auth/ratelimit.go

| Line | Issue | Fix |
|------|-------|-----|
| 172 | gofumpt | Move `})` to proper position |
| 41 | predeclared | Rename `max` to `maxRequests` |

### 4-7. internal/auth/*.go (jwt, origin, password, validate)

Add package comment: `// Package security provides authentication utilities.`

### 8. internal/auth/secretstore/secretstore.go

| Line | Issue | Fix |
|------|-------|-----|
| 157 | godot | Change `// Errors` to `// Errors.` |
| 45 | gofumpt | Format file |
| 80 | gofumpt | Format file |
| 1 | stylecheck | Add package comment |

### 9-10. internal/fcm/*.go (fcm, notifier)

Add package comment: `// Package fcm provides Firebase Cloud Messaging integration.`

### 11. pkg/crypto/hmac.go

| Line | Issue | Fix |
|------|-------|-----|
| 27, 66 | gofumpt | Format file |
| 74,79,82,86,93 | perfsprint | Change `fmt.Errorf("...")` to `errors.New("...")` |
| 1 | stylecheck | Add package comment |

### 12. pkg/models/auth.go

| Line | Issue | Fix |
|------|-------|-----|
| 92 | stylecheck - ST1021 | Change to `// OperatorRegisterRequest is the payload for operator self-registration.` |

### 13. pkg/storage/sqlite.go

| Line | Issue | Fix |
|------|-------|-----|
| 11, 17 | gci/goimports | Move imports to proper group |
| 216, 379, 645, 687 | gofumpt | Format file |
| 1318-1320, 1323-1325 | nilerr | Check err before returning nil |
| 595, 780, 818, 872, 1042, 1114, 1178 | nilnil | Return (nil, errSomething) instead of (nil, nil) |
| 1332 | perfsprint | Use `strconv.Itoa(seconds)` instead of `fmt.Sprintf("%d", seconds)` |
| 1 | stylecheck | Add package comment |

### 14-18. internal/api/handlers/*.go

Add package comment: `// Package controllers provides HTTP handlers.`

gci fixes at lines: auth.go:15, command.go:8, device.go:9, server.go:13, updater.go:11, websocket_handler.go:8

godot fixes: 7 comments in auth.go, 4 in command.go, 4 in device.go, 7 in updater.go, 1 in websocket_handler.go

perfsprint: auth.go:869

nestif: command.go:112, server.go:328

### 19-24. internal/api/middleware/*.go

Add package comment: `// Package middleware provides HTTP middleware.`

godot fixes: body_size.go:27,30, request_id.go:16

### 25-26. internal/ws/*.go (client, hub)

Add package comment: `// Package hub provides WebSocket functionality.`

gci fix: client.go:9, nestif: client.go:83

### 27. pkg/config/config.go

| Line | Issue | Fix |
|------|-------|-----|
| 112, 123 | gofumpt | Format file |
| 98, 101 | perfsprint | Change to `errors.New("...")` |

### 28. internal/command_signer.go

| Line | Issue | Fix |
|------|-------|-----|
| 12 | gci | Move import to proper group |
| 25, 49, 108 | godot | Change to end with period |
| 155 | perfsprint | Use `strconv.FormatInt` instead of fmt.Sprintf |
| 1 | stylecheck | Add package comment |

### 29. internal/email.go

| Line | Issue | Fix |
|------|-------|-----|
| 47, 68, 89 | perfsprint | Change to `errors.New("RESEND_API_KEY not configured")` |
| 44, 65, 86 | stylecheck - ST1020 | Change comment format |
| 1 | stylecheck | Add package comment |

---

## Quick Fix Commands

### Auto-fix formatting (gofumpt, goimports, gci)
```bash
cd apps/api
golangci-lint run --fix ./...
```

### Manual fixes required
- godot (comments ending with period)
- perfsprint (fmt.Errorf to errors.New)
- nilnil (proper error returns)
- nilerr (check errors before nil return)
- nestif (simplify nested conditions)
- noctx (add context to HTTP calls)
- predeclared (rename variable)
- stylecheck (add package comments, fix error names)

---

## Auto-Fixable (run golangci-lint --fix)

| Linter | Can Auto-Fix |
|--------|--------------|
| gofmt | Yes |
| gofumpt | Yes |
| goimports | Yes |
| gci | Yes |

---

## Files Needing Package Comments (27 total)

1. internal/auth/google_token.go
2. internal/auth/jwt.go
3. internal/auth/origin.go
4. internal/auth/password.go
5. internal/auth/ratelimit.go
6. internal/auth/validate.go
7. internal/auth/secretstore/secretstore.go
8. internal/fcm/fcm.go
9. internal/fcm/notifier.go
10. pkg/crypto/hmac.go
11. pkg/storage/sqlite.go
12. internal/api/handlers/auth.go
13. internal/api/handlers/command.go
14. internal/api/handlers/device.go
15. internal/api/handlers/server.go
16. internal/api/handlers/updater.go
17. internal/api/handlers/websocket_handler.go
18. internal/api/middleware/auth.go
19. internal/api/middleware/body_size.go
20. internal/api/middleware/cors.go
21. internal/api/middleware/logger.go
22. internal/api/middleware/rate_limiter.go
23. internal/api/middleware/request_id.go
24. internal/ws/client.go
25. internal/ws/hub.go
26. internal/command_signer.go
27. internal/email.go

---

## End of Document