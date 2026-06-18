# API Security Documentation
# Implements PRD Section 4.16 - Documentation

## Overview

This document describes the security mechanisms implemented in the Vyzorix Update Server API.

## Authentication

### Session-Based Authentication

The API uses session-based authentication with secure cookies.

**Login Endpoint:** `POST /v1/auth/login`
```json
{
  "email": "user@example.com",
  "password": "secure_password"
}
```

**Response:**
```json
{
  "success": true,
  "operator": {
    "id": "0191abcd-1234-7fff-8aaa-123456789abc",
    "email": "user@example.com"
  }
}
```

### JWT Tokens

For API access, JWT tokens are issued with configurable expiration.

**Headers Required:**
```
Authorization: Bearer <jwt_token>
```

## Request Signing (HMAC-SHA512)

All API requests (except health checks) must be signed using HMAC-SHA512.

### Required Headers

| Header | Description |
|--------|-------------|
| `X-Client-ID` | The client's unique identifier |
| `X-Timestamp` | Unix timestamp of the request |
| `X-Signature` | HMAC-SHA512 signature |
| `X-Encrypted-Body` | Base64-encoded encrypted body (optional) |

### Signature Format

```
t={timestamp},v1={hmac_signature}
```

### Signing Algorithm

1. Create SHA-512 hash of body
2. Build string to sign: `{method}\n{path}\n{timestamp}\n{sha512_body_hash_hex}`
3. Compute HMAC-SHA512 using client secret
4. Format as `t={timestamp},v1={hex_signature}`

### Code Example (JavaScript)

```javascript
async function signRequest(method, path, body, clientSecret) {
  const timestamp = Math.floor(Date.now() / 1000);
  
  // Step 1: Compute SHA-512 hash of body
  const bodyBytes = new TextEncoder().encode(body || '');
  const bodyHashBuffer = await crypto.subtle.digest('SHA-512', bodyBytes);
  const bodyHashHex = Array.from(new Uint8Array(bodyHashBuffer))
    .map(b => b.toString(16).padStart(2, '0')).join('');
  
  // Step 2: Build string to sign
  const stringToSign = `${method}\n${path}\n${timestamp}\n${bodyHashHex}`;
  
  // Step 3: Import key for HMAC-SHA512
  const key = await crypto.subtle.importKey(
    'raw',
    new TextEncoder().encode(clientSecret),
    { name: 'HMAC', hash: 'SHA-512' },
    false,
    ['sign']
  );
  
  // Step 4: Compute HMAC-SHA512
  const signature = await crypto.subtle.sign('HMAC', key,
    new TextEncoder().encode(stringToSign));
  const signatureHex = Array.from(new Uint8Array(signature))
    .map(b => b.toString(16).padStart(2, '0')).join('');
  
  return `t=${timestamp},v1=${signatureHex}`;
}
```

### Body Encryption (AES-256-GCM)

Request bodies can be encrypted using AES-256-GCM. The key is derived from the client secret using SHA-512 (first 32 bytes).

**Encryption Process:**
1. Derive 32-byte key from client secret using SHA-512
2. Generate random 12-byte nonce
3. Encrypt using AES-256-GCM
4. Prepend nonce to ciphertext
5. Base64 encode the result

**Encrypted Body Format:**
```
base64(nonce[12 bytes] || ciphertext)
```

## Rate Limiting

Rate limiting is enforced per-client and per-IP.

**Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1720000000
```

## Security Headers

All responses include security headers:

| Header | Value |
|--------|-------|
| `X-Content-Type-Options` | `nosniff` |
| `X-Frame-Options` | `DENY` |
| `X-XSS-Protection` | `1; mode=block` |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` |
| `Content-Security-Policy` | Configured per environment |

## CSRF Protection

CSRF tokens are required for mutative operations. Tokens are signed using HMAC-SHA512.

**Flow:**
1. GET `/v1/auth/csrf-token` - Retrieve CSRF token (signed token)
2. Include token in `X-CSRF-Token` header for POST/PUT/DELETE requests

**Token Format:**
```
{csrf_token}.{base64_hmac_signature}
```

**Signing Algorithm:**
1. Generate random CSRF token (32 bytes, base64 encoded)
2. Compute HMAC-SHA512 of token using CSRF secret
3. Format as `token.base64(signature)`

## Cloudflare Turnstile

Bot protection via Cloudflare Turnstile on auth endpoints.

**Protected Endpoints:**
- `/v1/auth/register`
- `/v1/auth/login`
- `/v1/auth/forgot-password`
- `/v1/auth/resend-password-reset`

**Token Header:** `X-Turnstile-Token`

## Account Lockout

After 5 failed login attempts, accounts are locked for 1 hour.

**Response when locked:**
```json
{
  "error": "account_locked",
  "message": "Too many failed attempts, please try again later",
  "retry_after": 3600
}
```

## Error Responses

All error responses follow a consistent format:

```json
{
  "error": "error_code",
  "message": "Human-readable message"
}
```

**Common Error Codes:**

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `unauthorized` | 401 | Missing or invalid authentication |
| `forbidden` | 403 | Authenticated but not authorized |
| `not_found` | 404 | Resource not found |
| `rate_limited` | 429 | Too many requests |
| `internal_server_error` | 500 | Server error (no details exposed) |

## Audit Logging

Security-relevant events are logged:

- Login attempts (success/failure)
- Password changes
- MFA enrollments
- Session revocations
- API client management
- Admin actions

## Related Documents

- [Container Hardening](./CONTAINER_HARDING.md)
- [Security Architecture](./COMPREHENSIVE_SECURITY_ARCHITECTURE.md)
- [Deployment Guide](./DEPLOYMENT.md)
