# Signing Architecture — Three Distinct Domains

This document reconciles the multiple HMAC signing systems in the codebase.
They are **not duplicates** — each serves a different trust boundary with
different headers, canonical formats, and key sources. Do not merge them.

---

## Overview

| Domain | Direction | Who signs | Who verifies | Headers | Canonical format | Key source |
|--------|-----------|-----------|--------------|---------|------------------|------------|
| **A. Client request signing** | Web app → Server | API Client (browser) | `session_signing.go` middleware | `X-Vyzorix-Timestamp`, `X-Vyzorix-Nonce`, `X-Vyzorix-Signature` | `METHOD\nPATH\nNONCE\nTIMESTAMP\nBODY` | Per-session `SigningKey` |
| **B. Device REST API signing** | Android device → Server | Device APK | `request_signing.go` middleware | `X-Client-ID`, `X-Timestamp`, `X-Signature` | `t={ts},v1={sig}` + SHA-512 body hash | Device `CommandSecretHash` |
| **C. Command frame signing** | Server → Android device | Server (`CommandSigner`) | Device APK | Frame fields: `nonce`, `signature`, `timestamp` | `dispatchId\|deviceId\|command\|ts_ms\|nonce\|args` | Device `CommandSecretHash` |

---

## Domain A — Client Request Signing (web app ↔ server)

**Scope:** ALL traffic from the web app to the server — REST, GraphQL
queries/mutations, and WebSocket subscription handshakes.

**Files:**
- Client: `packages/API_Client/src/vyzorServer/crypto/browser-sign.ts`
  - `signHttpRequestBrowser(method, path, body, signingKey)` → returns
    `X-Vyzorix-Timestamp`, `X-Vyzorix-Nonce`, `X-Vyzorix-Signature` headers
  - HMAC-SHA512, **base64** output, nonce-based replay protection
- Client interceptor: `packages/API_Client/src/vyzorServer/rest/_shared/rest-client.ts`
  - Axios request interceptor signs every outgoing request with the session
    signing key (skipped for auth endpoints where no key exists yet)
- Server verify: `apps/api/internal/infrastructure/crypto/verifier.go`
  - `Verifier.ReadAndVerify` / `ReadAndVerifyHTTP` validates the `X-Vyzorix-*`
    headers, checks timestamp window + nonce replay cache
- Server middleware: `apps/api/internal/api/middleware/session_signing.go`
  - `SessionSignatureMiddleware` — runs AFTER cookie/API-key auth, reads the
    session from the gin context, uses `session.SigningKey` as the HMAC secret
  - Wired into the `tenantGroup` route group in `server_routes.go`

**Key lifecycle:** The `SigningKey` is issued at login/MFA-verify time and
returned in the `LoginWithTokensResponse` / MFA verify response. The client
stores it in `authContext.signingKey` and the server stores it on the session
record.

---

## Domain B — Device REST API Signing (device → server)

**Scope:** REST API calls from the Android device APK to the server (device
registration, telemetry upload, command ack, etc.).

**Files:**
- Server middleware: `apps/api/internal/api/middleware/request_signing.go`
  - `RequestSigningMiddleware` — resolves the device by `X-Client-ID`,
    retrieves the device's `CommandSecretHash`, verifies the signature
  - Header format: `X-Client-ID`, `X-Timestamp`, `X-Signature`
    (signature is `t={unix_timestamp},v1={hmac_hex}`)
  - Includes AES-256-GCM body decryption (the device encrypts request bodies)
  - Replay cache keyed by client ID + timestamp
- Wired via `requireHMAC` / `requireStrictHMAC` in `server_handlers.go`

**Why not merge with Domain A?** Different client (device vs browser),
different header convention (`X-Client-ID` vs `X-Vyzorix-*`), different
canonical format (comma-delimited `t=,v1=` vs newline-delimited), and device
requests include body encryption that browser requests don't. Merging would
break the device API contract.

---

## Domain C — Command Frame Signing (server → device)

**Scope:** Command frames sent over WebSocket from the server to the Android
device. The web app is **not** involved in this domain — it dispatches via
REST (Domain A) and the server signs the outbound frame (Domain C).

**Files:**
- Signer: `apps/api/internal/infrastructure/crypto/command_signer.go`
  - `CommandSigner.SignCommand(frame, deviceID, secret)` → nonce + HMAC hex
  - `CommandSigner.ValidateCommandHMAC(frame, deviceID, secret)` → bool
    (device-side verification, called by the APK)
  - `SignCommandFrame(signer, frame, deviceID, secret)` — shared helper that
    signs a frame in place; called by all dispatch paths
- Dispatch paths (all three now sign before delivery):
  - `apps/api/internal/api/handlers/command/command_execute.go` —
    `signCommandFrame()` signs before `hub.Send`
  - `apps/api/internal/application/command/command_outbox.go` —
    `deliverViaWebSocket()` signs before `hub.SendWithDeliveryConfirmation`
  - `apps/api/internal/application/updates/updates_push_service.go` —
    `dispatchUpdateCommand()` signs before `hub.Send`

**Key source:** The device's `CommandSecretHash` — a deterministic SHA-256
derivation of the plaintext command secret issued at registration. Both the
server (stored hash) and the device (computes `SHA-256(plaintext)`) derive
the same HMAC key without the server ever storing the plaintext.

**Frame fields:** The `CommandFrame` struct carries `nonce`, `signature`,
and `timestamp` JSON fields. The server generates all three; client-provided
values are discarded (the server is the signing authority).

---

## Why three systems instead of one?

1. **Different trust boundaries:** The web app authenticates via session
   cookies/API keys; devices authenticate via HMAC headers. A single system
   would conflate these trust models.

2. **Different directions:** Domain A and B are inbound (client/device →
   server); Domain C is outbound (server → device). Inbound needs middleware;
   outbound needs a signer called at dispatch time.

3. **Different key sources:** Sessions have a `SigningKey`; devices have a
   `CommandSecretHash`; API keys need their own signing secret (planned).
   A single key store doesn't fit all three.

4. **Different canonical formats:** Browser uses newline-delimited canonical
   with nonce; devices use comma-delimited `t=,v1=` with body hash; command
   frames use pipe-delimited with args. Each format is optimized for its
   transport and validation requirements.

---

## API Key Signing (planned)

API keys currently authenticate without request signing. To enable signed
API-key requests, the `APIKey` entity will gain a `signing_secret` field
(derived from the API key value at creation time). This extends Domain A to
API-key-authenticated clients — same `X-Vyzorix-*` headers, same verifier,
but the key source is the API key's `signing_secret` instead of a session
`SigningKey`.
