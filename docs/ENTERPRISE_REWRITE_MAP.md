# Enterprise Rewrite Map

This document maps all files with basic/simple/insecure implementations that need enterprise-grade rewrites.

---

## 🔴 CRITICAL - Security Vulnerabilities

### 1. `pkg/storage/crypto.go` - Insecure Random Fallback

**Lines:** 176-182

**Current Issue:**
```go
func randomBytes(b []byte) {
    if _, err := rand.Read(b); err != nil {
        // Fallback - should never happen in practice.
        for i := range b {
            b[i] = byte(time.Now().UnixNano() & 0xff)  // ⚠️ PREDICTABLE!
        }
    }
}
```

**Risk:** In the extremely rare case crypto/rand fails, the fallback uses time-based predictable values. This could compromise salts/hashes.

**Priority:** HIGH

**Rewrite Strategy:**
- Remove fallback entirely
- If crypto/rand fails, panic or return error
- Add logging for the rare failure case

---

### 2. `pkg/storage/uuid.go` - Insecure UUID Random Fallback

**Lines:** 54-61

**Current Issue:**
```go
if _, err := uuidRand.Read(randBytes[:]); err != nil {
    // Fallback to less secure random on error
    for i := range randBytes {
        randBytes[i] = byte(time.Now().UnixNano() & 0xff)  // ⚠️ PREDICTABLE!
    }
}
```

**Risk:** UUIDs become predictable if crypto/rand fails, compromising uniqueness guarantees.

**Priority:** HIGH

**Rewrite Strategy:**
- Remove fallback entirely
- Fail explicitly if crypto/rand fails
- Consider using `github.com/google/uuid` library instead of custom implementation

---

### 3. `internal/auth/lockout.go` - Weak Fake Password Validation

**Lines:** 145-158

**Current Issue:**
```go
func IsValidPassword(password string) bool {
    if len(password) == 0 {
        return false
    }
    // Constant-time validation using XOR (avoids actual hash computation)
    result := 1
    for i := 0; i < len(password); i++ {
        result ^= int(password[i])  // ⚠️ NOT CRYPTOGRAPHIC!
    }
    return result != 0
}
```

**Risk:** XOR-based check is NOT cryptographically sound. While this is for timing uniformity, the implementation is flawed.

**Priority:** HIGH

**Rewrite Strategy:**
- Replace with proper dummy Argon2id computation
- Use consistent timing with actual password verification
- Store a pre-computed fake hash for comparison

---

### 4. `internal/auth/lockout.go` - Predictable Fake Token Fallback

**Lines:** 160-171

**Current Issue:**
```go
if _, err := rand.Read(b); err != nil {
    // Fallback - should never happen in practice
    for i := range b {
        b[i] = byte(i % 256)  // ⚠️ SEQUENTIAL, NOT RANDOM!
    }
}
```

**Risk:** Fallback generates sequential values, not random.

**Priority:** MEDIUM

**Rewrite Strategy:**
- Remove fallback entirely
- If crypto/rand fails, return empty string or error
- Real crypto/rand is virtually always available

---

### 5. `internal/command_signer.go` - Weak Fallback Hash

**Lines:** 140-153

**Current Issue:**
```go
func (s *CommandSigner) fallbackHash(secret string) string {
    salt := make([]byte, 16)
    if _, err := rand.Read(salt); err != nil {
        salt = []byte(strconv.FormatInt(time.Now().UnixNano(), 10))  // ⚠️ WEAK!
    }
    // ⚠️ Uses plain SHA512 without proper key derivation!
    mac := sha512.New()
    mac.Write(salt)
    mac.Write([]byte(secret))
    hash := mac.Sum(nil)
    return saltHex + ":" + hex.EncodeToString(hash)
}
```

**Risk:** 
1. Weak fallback salt generation
2. Plain SHA512 is not a proper key derivation function (should use PBKDF2/Argon2)

**Priority:** HIGH

**Rewrite Strategy:**
- Remove fallback entirely
- Use `golang.org/x/crypto/pbkdf2` as a fallback KDF if Argon2 is unavailable

---

## 🟡 HIGH - Production Risks

### 6. `internal/api/middleware/rate_limiter.go` - Memory Leak

**Lines:** 11-24

**Current Issue:**
```go
type RateLimiter struct {
    buckets  map[string]*bucket  // ⚠️ GROWS UNBOUNDED!
    Capacity int
    Refill   time.Duration
    mu       sync.Mutex
}
```

**Risk:** The bucket map never cleans up expired entries. In production with many unique IPs, this could lead to memory exhaustion.

**Priority:** HIGH

**Rewrite Strategy:**
- Add background cleanup goroutine with periodic expiration
- Use time-based bucket expiration
- Consider using `github.com/RussellLuo/slidingwindow` for sliding window rate limiting

---

### 7. `internal/api/handlers/auth_utils.go` - Connection Pool Not Reused

**Current Issue:**
```go
client := &http.Client{Timeout: 10 * time.Second}  // ⚠️ NEW CLIENT PER REQUEST!
httpResp, err := client.Do(req)
```

**Risk:** Creating a new HTTP client for each request prevents connection reuse and DNS caching.

**Priority:** MEDIUM

**Rewrite Strategy:**
- Create a package-level HTTP client with proper configuration
- Use `http.DefaultTransport` settings with custom timeouts
- Add connection pooling configuration

---

### 8. `internal/auth/password.go` - Limited Special Character Set

**Current Issue:**
```go
specialChars := "!@#$%^&*()_+-="  // ⚠️ INCOMPLETE SET
```

**Risk:** Many commonly allowed special characters are excluded (quotes, backticks, slashes, etc.)

**Priority:** LOW

**Rewrite Strategy:**
- Expand character set to OWASP recommendations
- Or document why it's intentionally limited for UX

---

## 🟢 MEDIUM - Code Quality Improvements

### 9. `pkg/crypto/hmac.go` - Non-Standard HMAC Algorithm

**Current Issue:**
```go
mac := hmac.New(sha512.New, []byte(secret))  // ⚠️ SHA512 vs SHA256
```

**Risk:** HMAC-SHA256 is more standard and efficient for most use cases.

**Priority:** LOW

**Rewrite Strategy:**
- Consider standardizing on HMAC-SHA256
- Document if SHA512 is required for specific security requirements

---

### 10. `internal/auth/session.go` - SHA512 Overkill for Lookups

**Current Issue:**
```go
func HashOperatorID(operatorID string) string {
    h := sha512.Sum512([]byte(operatorID))  // ⚠️ SHA512 for simple lookup
    return hex.EncodeToString(h[:])
}
```

**Risk:** SHA-256 would be sufficient and faster for non-security-critical lookups.

**Priority:** LOW

**Rewrite Strategy:**
- Consider using SHA-256 for database lookups
- Or document why SHA-512 is required

---

## 📋 Complete File List with Priority

| File | Lines | Issue | Priority | Status |
|------|-------|-------|----------|--------|
| `pkg/storage/crypto.go` | 176-182 | Insecure random fallback | 🔴 CRITICAL | TODO |
| `pkg/storage/uuid.go` | 54-61 | Insecure UUID random fallback | 🔴 CRITICAL | TODO |
| `internal/auth/lockout.go` | 145-158 | Weak fake password validation | 🔴 CRITICAL | TODO |
| `internal/command_signer.go` | 140-153 | Weak fallback hash | 🔴 CRITICAL | TODO |
| `internal/auth/lockout.go` | 160-171 | Predictable fake token fallback | 🟡 HIGH | TODO |
| `internal/api/middleware/rate_limiter.go` | 11-24 | Memory leak - no cleanup | 🟡 HIGH | TODO |
| `internal/api/handlers/auth_utils.go` | - | Connection pool not reused | 🟡 MEDIUM | TODO |
| `internal/auth/password.go` | - | Limited special char set | 🟢 LOW | TODO |
| `pkg/crypto/hmac.go` | - | Non-standard HMAC algorithm | 🟢 LOW | TODO |
| `internal/auth/session.go` | - | SHA512 overkill for lookups | 🟢 LOW | TODO |

---

## 📊 Summary Statistics

- **Total Files Needing Attention:** 10
- **Critical Security Issues:** 4
- **High Production Risks:** 2
- **Medium Code Quality:** 2
- **Low Improvements:** 2

---

## ✅ Already Rewritten (Recent Fixes)

| File | Issue | Fixed |
|------|-------|-------|
| `internal/auth/totp_qr.go` | Simplified QR encoding | ✅ v2.0 |
| `internal/auth/validate.go` | Basic email regex | ✅ v2.0 |
| `internal/api/middleware/replay_protection.go` | Fake sorting | ✅ v2.0 |
| `internal/api/handlers/auth_csrf.go` | Insecure CSRF storage | ✅ v2.0 |
| `internal/auth/origin.go` | HTTPS enforcement bug | ✅ v2.0 |
| `internal/api/handlers/server.go` | Placeholder client lookup | ✅ v2.0 |

---

## 🔧 Rewrite Guidelines

### For Critical Security Issues:
1. Remove fallback mechanisms entirely
2. Fail explicitly if cryptographic operations fail
3. Document why fallback was removed
4. Add logging for failure cases

### For High Production Risks:
1. Implement proper cleanup mechanisms
2. Use established libraries where available
3. Add monitoring/observability

### For Code Quality:
1. Document intentional design decisions
2. Consider performance vs. security tradeoffs
3. Add comprehensive tests
