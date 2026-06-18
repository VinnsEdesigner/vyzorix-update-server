# PRD: API Request Signing & Encryption System

> **Feature Name:** Request Signing with End-to-End Encryption  
> **Version:** 1.0  
> **Status:** Draft  
> **Created:** 2026-06-15  
> **Target Release:** Production MVP  

---

## 1. Introduction

### Problem Statement

Currently, the Vyzorix API has **no protection against unauthorized API calls**. An attacker who knows or guesses an endpoint URL can execute sensitive operations (device wipe, command dispatch, settings changes) without any authentication or proof of origin.

**Example Attack:**
```bash
# Attacker sends this directly - NO AUTH REQUIRED
curl -X POST https://api.vinnsedesigner.render.com/v1/device/wipe \
  -H "Content-Type: application/json" \
  -d '{"device_id":"0191abcd-1234-7fff-8aaa-123456789abc"}'
```

This is unacceptable for production where the API controls real device behavior.

### Solution

Implement **cryptographic request signing + encryption** for all API endpoints (except health checks). Every request must carry a verifiable proof of origin and encrypted payload, ensuring:
1. Only registered, authorized clients can call the API
2. Requests cannot be tampered with in transit
3. Request bodies are encrypted and unreadable if intercepted
4. Replay attacks are prevented via timestamp validation

---

## 2. Goals

- **G1:** Prevent unauthorized API access - only clients with valid signed requests can call protected endpoints
- **G2:** Protect request integrity - detect any tampering with request body or headers
- **G3:** Encrypt sensitive payloads - request bodies are encrypted, not just signed
- **G4:** Prevent replay attacks - requests older than 5 minutes are rejected
- **G5:** Provide API key management - admins can create, view, and revoke client credentials
- **G6:** Unified client approach - web frontend and mobile apps use the same signing mechanism
- **G7:** Zero disruption to existing functionality - signing is transparent to authorized clients

---

## 3. User Stories

### US-001: Client Registration & Secret Issuance
**Description:** As an admin, I want to register new API clients and issue them credentials so that legitimate frontends and mobile apps can authenticate.

**Acceptance Criteria:**
- [ ] Admin can create a new client with name, platform (web/ios/android), and allowed origins
- [ ] System generates a unique `client_id` (UUIDv7) and `client_secret` (32-byte random hex)
- [ ] `client_secret` is shown ONCE at creation time and cannot be retrieved again
- [ ] `client_secret` is stored as Argon2id hash in database
- [ ] Client can be activated/deactivated without deletion

---

### US-002: Client Secret Provisioning via API
**Description:** As a frontend client, I want to fetch my signing credentials after login so I can sign subsequent API requests.

**Acceptance Criteria:**
- [ ] Authenticated users can call `POST /v1/auth/client-credentials` to get their client credentials
- [ ] Response includes `client_id` and `client_secret` (encrypted in transit)
- [ ] Credentials are cached client-side for the session duration
- [ ] Each user can have multiple active clients (web, mobile, etc.)

---

### US-003: Request Signing (Client-Side)
**Description:** As a frontend client, I want to sign every API request so the backend can verify my identity and accept the request.

**Acceptance Criteria:**
- [ ] All HTTP methods (GET, POST, PUT, DELETE, PATCH) are signable
- [ ] Signature covers: HTTP method, path, timestamp, encrypted body
- [ ] Signature format: `t={timestamp},v1={hmac_signature}`
- [ ] Body is encrypted using AES-256-GCM before signing
- [ ] Signature headers: `X-Client-ID`, `X-Timestamp`, `X-Signature`, `X-Encrypted-Body`

---

### US-004: Request Verification (Server-Side)
**Description:** As the backend, I want to verify every incoming request's signature so I can reject unauthorized or tampered requests.

**Acceptance Criteria:**
- [ ] Middleware intercepts all protected endpoints before handler execution
- [ ] Rejects requests missing any required signature header (401)
- [ ] Rejects requests with timestamp outside ±5 minute window (401)
- [ ] Rejects requests with invalid signature (401)
- [ ] Rejects requests from inactive or unknown clients (401)
- [ ] Rejects replayed requests (same signature used twice within window)
- [ ] Decrypts request body before passing to handler

---

### US-005: Replay Attack Prevention
**Description:** As the backend, I want to detect and reject replayed requests so attackers cannot record and reuse valid signed requests.

**Acceptance Criteria:**
- [ ] Recently seen signatures are cached in memory (Redis-style TTL map)
- [ ] Cache window matches signature timestamp window (5 minutes)
- [ ] Replay attempt returns 401 with code `REPLAY_DETECTED`
- [ ] Cache automatically evicts entries older than window

---

### US-006: Response Encryption (Optional)
**Description:** As a client, I want sensitive responses to be encrypted so even if intercepted, the data is unreadable.

**Acceptance Criteria:**
- [ ] Client can request encrypted response via `X-Encrypt-Response: true` header
- [ ] Encrypted responses include `X-Encryption-Nonce` header
- [ ] Client can decrypt using same client secret

---

### US-007: Client Management UI
**Description:** As an admin, I want a UI to manage API clients so I can issue and revoke access without database manipulation.

**Acceptance Criteria:**
- [ ] Admin dashboard shows list of all clients with status
- [ ] Admin can create new client (name, platform, allowed origins)
- [ ] Admin can view client details (not the secret)
- [ ] Admin can deactivate/reactivate a client
- [ ] Admin can revoke a client (permanently invalidates secret)
- [ ] Admin can view client's last request time and request count

---

### US-008: Signing Key Rotation
**Description:** As an admin, I want to rotate signing keys periodically so we maintain cryptographic hygiene.

**Acceptance Criteria:**
- [ ] Admin can trigger key rotation for a client
- [ ] New requests must use new key
- [ ] Old key remains valid for 24-hour grace period
- [ ] After grace period, old key requests are rejected
- [ ] Rotation is logged in audit trail

---

### US-009: Exclude Non-Critical Endpoints
**Description:** As a developer, I want certain endpoints to be exempt from signing so health checks and public data can still function.

**Acceptance Criteria:**
- [ ] `/health/live` - exempt (returns 200 always)
- [ ] `/health/ready` - exempt (returns 200 if DB connected)
- [ ] `/health/secure` - requires signing (returns 200 if all security features OK)
- [ ] `/metrics` - requires signing (Prometheus metrics)
- [ ] All other endpoints require valid signature

---

### US-010: Graceful Degradation
**Description:** As an operator, I want the system to gracefully handle signing infrastructure failures so we don't have a complete outage.

**Acceptance Criteria:**
- [ ] If signing service is unavailable, log critical error
- [ ] Allow requests without signature only if `ALLOW_UNSIGNED_FALLBACK=true`
- [ ] Fallback mode logs all unsigned requests for audit
- [ ] Automatic alert when falling back to unsigned mode

---

## 4. Functional Requirements

### FR-1: Client Registry
- FR-1.1: Database table `api_clients` stores client records with fields: `id`, `client_secret_hash`, `name`, `platform`, `allowed_origins`, `allowed_paths`, `rate_limit`, `is_active`, `created_at`, `updated_at`
- FR-1.2: Client ID is a UUIDv7 for time-ordering and unguessability
- FR-1.3: Client secret is 32 bytes (64 hex chars), generated with `crypto/rand`
- FR-1.4: Client secret is hashed with Argon2id before storage
- FR-1.5: Allowed origins is a JSON array stored as text

### FR-2: Signing Key Storage
- FR-2.1: Database table `signing_keys` stores key versions with fields: `id`, `key_hash`, `version`, `client_id`, `issued_at`, `expires_at`, `is_active`, `revoked_at`
- FR-2.2: Each client can have multiple active keys (for rotation)
- FR-2.3: Key lookup is by `client_id + is_active + not_expired`

### FR-3: Request Signing Algorithm
- FR-3.1: **String to sign format:** `{METHOD}\n{PATH}\n{TIMESTAMP}\n{BODY_SHA512}`
- FR-3.2: **Signature format:** `t={TIMESTAMP},v1={HMAC-SHA512(secret, string_to_sign)}`
- FR-3.3: Body is AES-256-GCM encrypted with random 12-byte nonce before signing
- FR-3.4: Nonce is prepended to ciphertext
- FR-3.5: Encrypted body is base64-encoded for transport
- FR-3.6: AES key derived from client secret using SHA-512 (first 32 bytes)

### FR-4: Request Verification Middleware
- FR-4.1: Middleware runs on all routes except excluded ones
- FR-4.2: Extract and validate all required headers
- FR-4.3: Verify timestamp within ±300 seconds (5 minutes)
- FR-4.4: Check client exists and is active
- FR-4.5: Verify signature using client's current active secret
- FR-4.6: Check signature is not in replay cache
- FR-4.7: Add signature to replay cache with TTL
- FR-4.8: Decrypt and replace request body

### FR-5: Replay Prevention Cache
- FR-5.1: In-memory cache with sync.Map for thread safety
- FR-5.2: Key: full signature string, Value: timestamp
- FR-5.3: Auto-cleanup of entries older than 5 minutes
- FR-5.4: Maximum cache size: 100,000 entries (configurable)

### FR-6: Admin API Endpoints
- FR-6.1: `GET /v1/admin/clients` - list all clients (paginated)
- FR-6.2: `POST /v1/admin/clients` - create new client
- FR-6.3: `GET /v1/admin/clients/{id}` - get client details
- FR-6.4: `PUT /v1/admin/clients/{id}` - update client
- FR-6.5: `DELETE /v1/admin/clients/{id}` - revoke client (soft delete)
- FR-6.6: `POST /v1/admin/clients/{id}/rotate` - rotate signing key
- FR-6.7: `GET /v1/admin/clients/{id}/usage` - get client's request statistics

### FR-7: Client Credentials Endpoint
- FR-7.1: `POST /v1/auth/client-credentials` - get client's signing credentials
- FR-7.2: Requires authenticated session (HttpOnly cookie)
- FR-7.3: Creates client record if first time, retrieves if exists
- FR-7.4: Returns `{client_id, client_secret, expires_at}`

### FR-8: Response Encryption
- FR-8.1: If `X-Encrypt-Response: true` header present, encrypt response body
- FR-8.2: Use AES-256-GCM with client secret as key
- FR-8.3: Include `X-Encryption-Nonce` header with nonce
- FR-8.4: Return `Content-Type: application/octet-stream` for encrypted responses

---

## 5. Non-Goals

- **NG-1:** This PRD does NOT cover OAuth/OIDC integration (handled separately)
- **NG-2:** This PRD does NOT cover mobile-specific certificate pinning (handled separately)
- **NG-3:** This PRD does NOT cover API usage billing/quotas beyond basic rate limiting
- **NG-4:** This PRD does NOT cover WebSocket connection signing (separate mechanism)
- **NG-5:** This PRD does NOT cover third-party API integrations (they use their own auth)

---

## 6. Design Considerations

### 6.1 Database Schema

```sql
-- api_clients: Registered API clients
CREATE TABLE api_clients (
    id TEXT PRIMARY KEY,
    operator_id TEXT NOT NULL,              -- Owner (operator who created it)
    name TEXT NOT NULL,
    platform TEXT NOT NULL,                  -- "web" | "ios" | "android"
    client_secret_hash TEXT NOT NULL,
    allowed_origins TEXT,                    -- JSON array
    allowed_paths TEXT,                      -- JSON array
    rate_limit INTEGER NOT NULL DEFAULT 100,
    is_active INTEGER NOT NULL DEFAULT 1,
    request_count INTEGER NOT NULL DEFAULT 0,
    last_request_at INTEGER,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (operator_id) REFERENCES operators(id) ON DELETE CASCADE
);

-- signing_keys: Rotatable signing keys per client
CREATE TABLE signing_keys (
    id TEXT PRIMARY KEY,
    client_id TEXT NOT NULL,
    key_hash TEXT NOT NULL,                  -- Hash of actual key
    version INTEGER NOT NULL,
    issued_at INTEGER NOT NULL,
    expires_at INTEGER,                      -- NULL = never
    is_active INTEGER NOT NULL DEFAULT 1,
    revoked_at INTEGER,
    FOREIGN KEY (client_id) REFERENCES api_clients(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX idx_api_clients_operator ON api_clients(operator_id);
CREATE INDEX idx_api_clients_active ON api_clients(is_active);
CREATE INDEX idx_signing_keys_client ON signing_keys(client_id);
CREATE INDEX idx_signing_keys_active ON signing_keys(client_id, is_active);
```

### 6.2 Request Flow Diagram

```

                        REQUEST SIGNING FLOW                                  

                                                                              
  CLIENT (Web/Mobile)                    SERVER                               
                                                                              
  1. GET /v1/auth/client-credentials                                           
     (with session cookie)                                                    
                                            
                                      2. Validate session                     
                                      3. Create/fetch client record           
                                      4. Generate client_secret               
                                      5. Store hash in db                     
                                            
     { client_id, client_secret }         (encrypted response)                
                                                                              
  6. Cache credentials locally                                                
                                                                              
  7. BUILD API REQUEST                                                        
               
      Method: POST                                                           
      Path: /v1/device/wipe                                                  
      Body: {"device_id": "..."}                                             
      Timestamp: 1720000000                                                  
               
                                                                              
  8. ENCRYPT BODY                                                             
     key = SHA-512(client_secret)[0:32]  # Derive 256-bit key from secret
     nonce = random(12 bytes)
     AES-256-GCM(key, nonce, body) → encrypted_body (base64)               
                                                                              
  9. CREATE SIGNATURE                                                         
     body_hash = SHA-512(encrypted_body)
     string_to_sign = "POST\n/v1/device/wipe\n1720000000\n<body_hash_hex>" 
     signature = HMAC-SHA512(client_secret, string_to_sign)                   
     signature_string = "t=1720000000,v1={hex_signature}"                     
                                                                              
  10. SEND REQUEST                                                            
      POST /v1/device/wipe                                                    
      Headers:                                                                
        X-Client-ID: 0191abcd-...                                             
        X-Timestamp: 1720000000                                               
        X-Signature: t=1720000000,v1=abc123...                                
        X-Encrypted-Body: base64_encrypted_body                               
       
                                                                              
                                    11. EXTRACT HEADERS                        
                                    12. LOOKUP CLIENT                          
                                    13. CHECK TIMESTAMP (±5min)                
                                    14. CHECK REPLAY CACHE                     
                                    15. DECRYPT BODY                           
                                    16. VERIFY SIGNATURE                       
                                    17. PROCESS HANDLER                        
                                    18. ADD TO REPLAY CACHE                    
                                    19. LOG AUDIT                              
                                                                              
        
      { success: true }                                                       
                                                                              

```

### 6.3 Security Properties

| Property | Mechanism | Protection Against |
|----------|-----------|-------------------|
| **Origin Verification** | HMAC-SHA512 signature with client secret | Unauthorized clients |
| **Integrity** | HMAC-SHA512 over method+path+timestamp+SHA512(body) | Request tampering |
| **Confidentiality** | AES-256-GCM body encryption (key derived via SHA-512) | Eavesdropping |
| **Freshness** | Timestamp validation (±5min) | Replay attacks |
| **Replay Prevention** | Signature cache | Replay attacks |

### 6.4 Error Codes

| Code | HTTP Status | Meaning |
|------|-------------|---------|
| `SIGN_001` | 401 | Missing required signature headers |
| `SIGN_002` | 401 | Invalid timestamp format |
| `SIGN_003` | 401 | Request timestamp outside window |
| `SIGN_004` | 401 | Unknown or inactive client |
| `SIGN_005` | 401 | Invalid signature |
| `SIGN_006` | 401 | Replay detected |
| `SIGN_007` | 400 | Invalid encrypted body |
| `SIGN_008` | 403 | Client not allowed for this path |
| `SIGN_009` | 429 | Client rate limit exceeded |

### 6.5 Configuration

```bash
# Environment Variables
REQUEST_SIGNING_ENABLED=true
SIGNING_TIMESTAMP_WINDOW=300          # 5 minutes in seconds
SIGNING_MAX_CACHE_SIZE=100000         # Max replay cache entries
SIGNING_GRACE_PERIOD=86400            # 24 hours for key rotation
ALLOW_UNSIGNED_FALLBACK=false         # Emergency mode only
```

---

## 7. Technical Considerations

### 7.1 Performance

- **Latency overhead:** ~1-2ms per request (HMAC + AES operations are fast)
- **Memory overhead:** Replay cache ~100KB per 10,000 entries
- **No database lookups on hot path:** Client secret verified in-memory after first lookup
- **Connection pooling:** Use sync.Pool for cipher instances

### 7.2 Dependencies

- `golang.org/x/crypto/argon2` - Already in use
- `golang.org/x/crypto/hkdf` - Not needed (HMAC is sufficient)
- No new external dependencies required

### 7.3 Testing Strategy

- Unit tests for signing/verification logic
- Integration tests with actual HTTP requests
- Replay attack detection tests
- Key rotation tests
- Performance benchmarks (target: <5ms overhead)

---

## 8. Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Unauthenticated requests blocked | 100% | All requests without valid signature rejected |
| Replay attacks detected | 100% | Replayed signatures within window rejected |
| Signing latency overhead | <5ms p99 | Performance benchmark |
| Client creation success rate | >99.9% | Admin client creation succeeds |
| Key rotation success rate | >99.9% | Rotation completes without downtime |
| Zero unauthorized access incidents | 0 per month | Security audit |

---

## 9. Open Questions

1. **Should mobile apps cache credentials indefinitely?** Or expire after 30 days requiring re-auth?

2. **Should we support Ed25519 instead of HMAC-SHA512?** Current implementation uses HMAC-SHA512 which is NIST approved, FIPS 140-2 compliant, and widely supported across platforms.

3. **Do we need API versioning for the signing scheme itself?** (e.g., `v1` in signature indicates algorithm version)

4. **Should we log all signing failures to a separate table?** For security monitoring.

5. **Should admins be able to set per-client rate limits?** Yes, included in FR-1.

---

## 10. File Structure

```
apps/api/
 internal/
    api/
       middleware/
          request_signing.go           # Core signature verification
          request_signing_test.go
          replay_protection.go         # Replay attack prevention
          replay_protection_test.go
          response_encryption.go       # Encrypt responses
          signed_handlers.go           # Route registration
       handlers/
           admin_clients.go             # Admin CRUD for clients
           admin_clients_test.go
           client_credentials.go        # Get credentials endpoint
    auth/
        request_signer.go                # Client-side signing (for reference)
        request_signer_test.go
        client_registry.go               # Client management logic
        client_registry_test.go
 pkg/
    config/
       signing.go                       # Signing configuration
    storage/
        clients.go                       # Client storage operations
        clients_test.go
        signing_keys.go                  # Key rotation storage
        signing_keys_test.go
```

---

## 11. Dependencies on Other PRDs

- **Session Management** (if not already complete): Needed for `GET /v1/auth/client-credentials`
- **Security Headers**: Should be applied before signing middleware
- **Audit Logging**: Should log all signed request attempts (success and failure)

---

## 12. Out of Scope for This PRD (Future PRDs)

- Mobile-specific certificate pinning
- OAuth/OIDC integration
- WebSocket connection signing
- API usage billing/quotas
- Third-party API integrations
- GraphQL API signing (if added later)

---

*Document Version: 1.0*  
*Status: Ready for Review*  
*Next Steps: Review with team, then implementation*