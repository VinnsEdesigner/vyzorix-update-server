# Comprehensive Enterprise Security Architecture

> **Purpose:** Complete mapping of all security layers  system
> **Status:** Discussion Draft - Not Yet Implemented
> **Priority:** Critical

---

## Table of Contents

1. [Network Security Layers](#1-network-security-layers)
2. [Application Security Layers](#2-application-security-layers)  
3. [Authentication & Authorization Layers](#3-authentication--authorization-layers)
4. [Data Security Layers](#4-data-security-layers)
5. [Infrastructure Security Layers](#5-infrastructure-security-layers)
6. [Operational Security Layers](#6-operational-security-layers)
7. [Device/Mobile Security Layers](#7-devicemobile-security-layers)
8. [Implementation Priority Matrix](#8-implementation-priority-matrix)

---

## 0. Cloudflare Services (Already Implemented - Free Tier)

### What Cloudflare Provides for Free

| Service | Feature | Your Status |
|---------|---------|-------------|
| **DDoS Protection** | Unlimited L3/L4/L7 protection |  Active |
| **WAF** | OWASP ModSecurity Core Rule Set |  Active |
| **CDN** | 200+ global PoPs, smart caching |  Active |
| **TLS 1.3** | Latest encryption, automatic certs |  Active |
| **Bot Management** | JavaScript detection, challenge pages |  Active |
| **Turnstile** | CAPTCHA alternative (free tier) |  Active |
| **Argo Smart Routing** | Intelligent traffic routing |  Active |
| **Rate Limiting** | 10,000 free requests/month |  Partial |
| **Firewall Rules** | 5 free rules |  Partial |

### Cloudflare Configuration

**DNS Settings:**
- Proxy status:  Proxied (orange cloud)
- SSL/TLS:  Full (strict)
- Always Use HTTPS:  Enabled
- Automatic HTTPS Rewrites:  Enabled

**Security Settings:**
- Security Level:  Medium (recommended)
- Challenge Passage:  30 minutes
- Browser Integrity Check:  Enabled
- Privacy Pass Support:  Enabled

**WAF Settings:**
- Managed Rules:  OWASP ModSecurity Core Rule Set
- Sensitivity Level:  Medium
- Action Mode:  Block

**Bot Management:**
- Bot Fight Mode:  Enabled
- Super Bot Fight Mode:  Disabled (paid feature)
- JavaScript Detections:  Enabled

---

## 1. Network Security Layers

### 1.1 Edge Protection (Perimeter)

| Layer | Technology | Purpose | Your Current Status |
|-------|-----------|---------|-------------------|
| **WAF** | Cloudflare WAF (Free) | Filter malicious traffic, SQL injection, XSS at edge |  **Implemented via Cloudflare** |
| **DDoS Protection** | Cloudflare DDoS (Free) | Absorb volumetric attacks (L3/L4/L7) |  **Implemented via Cloudflare** |
| **API Gateway** | Kong, AWS API Gateway, NGINX | Rate limiting, auth, routing |  Not Implemented (Render handles basic routing) |
| **Load Balancer** | AWS ALB, Cloudflare | SSL termination, health checks |  Using Render default |
| **CDN** | Cloudflare CDN (Free) | Static asset caching, DDoS mitigation, 200+ PoPs |  **Implemented via Cloudflare** |

### 1.2 Transport Security

| Layer | Implementation | Purpose | Your Current Status |
|-------|---------------|---------|-------------------|
| **TLS 1.3** | Server config, CDN settings | Modern encryption only |  Render handles |
| **Certificate Management** | Let's Encrypt, AWS ACM | Automated renewal |  Render handles |
| **HSTS** | Response header | Force HTTPS |  Need to add header |
| **OCSP Stapling** | Server config | Faster certificate validation |  Render handles |
| **Certificate Pinning** | Mobile apps | Prevent MITM |  Not Implemented |

### 1.3 Network Segmentation

| Layer | Implementation | Purpose | Your Current Status |
|-------|---------------|---------|-------------------|
| **VPC/Private Subnet** | AWS VPC | Isolate database from public |  Render handles |
| **Security Groups** | AWS SG | Firewall rules |  Render handles |
| **Private Endpoints** | AWS PrivateLink | No public DB access |  Not Implemented |
| **Jump Server** | Bastion host | Admin access only |  Not Implemented |

---

## 2. Application Security Layers

### 2.1 Request Validation Funnel (Documented in Funnel.md)

```
[Incoming Request]
        │
        ▼
┌───────────────────────────────┐
│ L1: Ingress Filtering        │ ← Bots, missing User-Agent, known IPs
└───────────────────────────────┘
        │
        ▼
┌───────────────────────────────┐
│ L2: Memory Ceiling (1MB)     │ ← http.MaxBytesReader
└───────────────────────────────┘
        │
        ▼
┌───────────────────────────────┐
│ L3: Attestation (Turnstile)   │ ← Cloudflare verification
└───────────────────────────────┘
        │
        ▼
┌───────────────────────────────┐
│ L4: Rate Limiting            │ ← Per-session throttle
└───────────────────────────────┘
        │
        ▼
┌───────────────────────────────┐
│ L5: DOA Ownership Check     │ ← Query includes user_id
└───────────────────────────────┘
        │
        ▼
   [Database Execution]
```

### 2.2 Headers & Response Security

| Header | Purpose | Your Current Status |
|--------|---------|-------------------|
| `X-Frame-Options: DENY` | Clickjacking prevention |  Not Implemented |
| `X-Content-Type-Options: nosniff` | MIME sniffing prevention |  In golangci-lint |
| `X-XSS-Protection: 1; mode=block` | XSS filter (legacy browsers) |  Not Implemented |
| `Content-Security-Policy` | Whitelist scripts, styles, frames |  Not Implemented |
| `Strict-Transport-Security` | Force HTTPS for 6+ months |  Not Implemented |
| `Referrer-Policy` | Control referrer info |  Not Implemented |
| `Permissions-Policy` | Disable browser features |  Not Implemented |
| `Cross-Origin-Embedder-Policy` | Prevent Spectre attacks |  Not Implemented |
| `Cross-Origin-Opener-Policy` | Isolate browsing context |  Not Implemented |

### 2.3 Input Validation & Sanitization

| Layer | Implementation | Your Current Status |
|-------|---------------|-------------------|
| **Schema Validation** | Zod, Valibot, Yup |  Partial (env only) |
| **SQL Parameterization** | All queries use `?` placeholders |  Should verify |
| **Output Encoding** | HTML escape, JSON encode |  Need audit |
| **Filename Sanitization** | Strip path traversal `../` |  Not Implemented |
| **Email Validation** | Regex + disposable domain block |  Not Implemented |
| **UUIDv7 IDs** | Replace auto-increment integers |  Not Implemented |
| **Cloudflare Turnstile** | Bot detection, CAPTCHA alternative |  **Implemented** |

### 2.4 Error Handling & Logging

| Layer | Implementation | Your Current Status |
|-------|---------------|-------------------|
| **Panic Recovery** | Global middleware catching all panics |  In golangci-lint |
| **Structured Errors** | JSON `{code, message}` only |  Should verify |
| **No Stack Traces** | Never expose in production |  Need to verify |
| **Audit Logging** | Every mutation logged |  Not Implemented |
| **Log Redaction** | PII/sensitive data masked |  Not Implemented |

---

## 3. Authentication & Authorization Layers

### 3.1 Documented in AUTH.md - 5-Layer Auth Pipeline

```
[Registration/Login Request]
        │
        ▼
┌───────────────────────────────┐
│ L1: Rate Limiting            │ ← Token bucket per IP
└───────────────────────────────┘
        │
        ▼
┌───────────────────────────────┐
│ L2: Input Validation (1MB)     │ ← MaxBytesReader + Zod
└───────────────────────────────┘
        │
        ▼
┌───────────────────────────────┐
│ L3: User Enum Prevention     │ ← Constant-time 201 response
└───────────────────────────────┘
        │
        ▼
┌───────────────────────────────┐
│ L4: Argon2id Hashing         │ ← 64MB, 1 iter, 4 parallel
└───────────────────────────────┘
        │
        ▼
┌───────────────────────────────┐
│ L5: Async Email Queue        │ ← Background SMTP worker
└───────────────────────────────┘
        │
        ▼
   [Database Commit]
```

### 3.2 Session Management

| Layer | Implementation | Your Current Status |
|-------|---------------|-------------------|
| **HttpOnly Cookies** | `http.SetCookie(w, &http.Cookie{HttpOnly: true})` |  Should verify |
| **Secure Flag** | `Secure: true` (HTTPS only) |  Need to verify |
| **SameSite** | `SameSite: Strict` or `Lax` |  Need to verify |
| **Session Expiry** | Short-lived tokens |  JWT has expiry |
| **Session Rotation** | Refresh token rotation |  Not Implemented |
| **Concurrent Sessions** | Limit per account |  Not Implemented |
| **Logout Propagation** | Server-side revocation |  Not Implemented |

### 3.3 Token Security

| Layer | Implementation | Your Current Status |
|-------|---------------|-------------------|
| **JWT Signing** | Ed25519 or RS256 |  Using Ed25519 |
| **Token Revocation** | Redis/DB lookup |  Not Implemented |
| **Refresh Token** | Separate long-lived token |  Not Implemented |
| **Token Binding** | Bind to device/fingerprint |  Not Implemented |
| **Silent Refresh** | Background token renewal |  Not Implemented |

### 3.4 Authorization Models

| Layer | Implementation | Your Current Status |
|-------|---------------|-------------------|
| **RBAC** | Role-based access control |  Basic roles exist |
| **Ownership Check** | User owns resource |  DOA should verify |
| **Permission Matrix** | Capability-based |  Not Implemented |
| **Resource Scoping** | Users see only their data |  DOA should verify |

### 3.5 Advanced Auth (Enterprise)

| Layer | Purpose | Your Current Status |
|-------|---------|-------------------|
| **MFA/TOTP** | Time-based one-time passwords |  Not Implemented |
| **Passkeys/WebAuthn** | Passwordless authentication |  Not Implemented |
| **OAuth 2.0/OIDC** | Google, GitHub login |  Partial (Google exists) |
| **SAML** | Enterprise SSO |  Not Implemented |
| **LDAP** | Corporate directory |  Not Implemented |
| **mTLS** | Client certificates |  Not Implemented |
| **Cloudflare Turnstile** | Bot detection for auth flows |  **Implemented** |

---

## 4. Data Security Layers

### 4.1 Encryption at Rest

| Layer | Implementation | Your Current Status |
|-------|---------------|-------------------|
| **Database Encryption** | SQLite + OS-level (Render) |  Render handles |
| **Disk Encryption** | LUKS, AWS EBS encryption |  Render handles |
| **Backup Encryption** | AES-256 for backups |  Not Implemented |
| **Field-Level Encryption** | Sensitive fields encrypted |  Not Implemented |

### 4.2 Secrets Management

| Layer | Implementation | Your Current Status |
|-------|---------------|-------------------|
| **Vault** | HashiCorp Vault, AWS Secrets Manager |  Not Implemented |
| **Secret Rotation** | Automated credential rotation |  Not Implemented |
| **Env Var Cleanup** | "Read-and-burn" pattern |  Not Implemented |
| **No Hardcoding** | All secrets in vault/env |  Need audit |

### 4.3 Database Security

| Layer | Implementation | Your Current Status |
|-------|---------------|-------------------|
| **WAL Mode** | `PRAGMA journal_mode=WAL` |  Should verify |
| **Connection Limits** | `SetMaxOpenConns(1)` |  In DEFENSE.md |
| **Parameterized Queries** | All queries use `?` |  Should verify |
| **UUIDv7 IDs** | Replace auto-increment |  Not Implemented |
| **Audit Logging** | Track all mutations |  Not Implemented |

---

## 5. Infrastructure Security Layers

### 5.1 Container & Deployment

| Layer | Implementation | Your Current Status |
|-------|---------------|-------------------|
| **Minimal Base Image** | `distroless`, `alpine` |  Using default |
| **Non-Root User** | Run as unprivileged user |  Not Configured |
| **Read-Only Root FS** | Container restriction |  Not Configured |
| **No Privileged Mode** | Container capabilities |  Not Configured |
| **Image Scanning** | Trivy, Snyk on CI |  Not Implemented |
| **SBOM** | Software Bill of Materials |  Not Implemented |

### 5.2 Kubernetes/Orchestration

| Layer | Implementation | Your Current Status |
|-------|---------------|-------------------|
| **Network Policies** | Pod-to-pod restrictions |  Not on K8s (Render) |
| **Resource Limits** | CPU/memory caps |  Render handles |
| **Pod Security** | PSP/PSA policies |  Not on K8s |
| **Secret Encryption** | Kubernetes secrets encrypted |  Not on K8s |

### 5.3 CI/CD Security

| Layer | Implementation | Your Current Status |
|-------|---------------|-------------------|
| **SBOM Generation** | Track dependencies |  Not Implemented |
| **Dependency Scanning** | `govulncheck`, npm audit |  golangci-lint has some |
| **SAST** | Static analysis in pipeline |  golangci-lint |
| **DAST** | Dynamic scanning (OWASP ZAP) |  Not Implemented |
| **Container Scanning** | Trivy in pipeline |  Not Implemented |
| **Signed Commits** | Require signed commits |  Not Enforced |
| **SBOM in Images** | Embed in container |  Not Implemented |

---

## 6. Operational Security Layers

### 6.1 Monitoring & Detection

| Layer | Implementation | Your Current Status |
|-------|---------------|-------------------|
| **SIEM** | Splunk, Datadog, Elastic |  Not Implemented |
| **APM** | Datadog, New Relic |  Not Implemented |
| **Uptime Monitoring** | Status page, alerts |  Render basic |
| **Anomaly Detection** | ML-based alerting |  Not Implemented |
| **Real User Monitoring** | Session tracking |  Not Implemented |

### 6.2 Incident Response

| Layer | Implementation | Your Current Status |
|-------|---------------|-------------------|
| **IR Plan** | Documented runbook |  Not Implemented |
| **Breach Notification** | SLA-defined process |  Not Implemented |
| **Forensics** | Log retention, snapshots |  Not Implemented |
| **Communication Plan** | Stakeholder notification |  Not Implemented |

### 6.3 Compliance

| Framework | Requirement | Your Current Status |
|-----------|-------------|-------------------|
| **GDPR** | EU data protection |  Not Assessed |
| **SOC 2** | Security controls |  Not Assessed |
| **HIPAA** | Healthcare data |  Not Applicable |
| **PCI-DSS** | Payment card data |  Not Applicable |

### 6.4 Backup & Recovery

| Layer | Implementation | Your Current Status |
|-------|---------------|-------------------|
| **Automated Backups** | Daily snapshots |  Render handles |
| **Backup Encryption** | AES-256 at rest |  Not Implemented |
| **Backup Testing** | Quarterly restoration test |  Not Implemented |
| **Recovery Time Objective** | Defined RTO |  Not Defined |
| **Recovery Point Objective** | Defined RPO |  Not Defined |

---

## 7. Device/Mobile Security Layers

*(From Funnel.md PART 3)*

### 7.1 Android Configuration

| Layer | Implementation | Your Current Status |
|-------|---------------|-------------------|
| **Certificate Pinning** | network-security-config.xml |  Not Implemented |
| **No Debug Flags** | `android:debuggable="false"` |  Not Assured |
| **Obfuscation** | ProGuard/R8 |  Not Implemented |
| **Root Detection** | Detect compromised devices |  Not Implemented |
| **Emulator Detection** | Block rooted devices |  Not Implemented |

### 7.2 FCM/Notification Security

| Layer | Implementation | Your Current Status |
|-------|---------------|-------------------|
| **Tickle Pattern** | Empty push → mTLS pull |  Not Implemented |
| **Signed Payloads** | Ed25519 command signing |  Implemented |
| **Anti-Replay Nonce** | Timestamp + sequence |  Not Implemented |
| **Intent Sanitization** | No Runtime.exec() |  Not Implemented |

---

## 8. Implementation Priority Matrix

### Critical (Must Have - Block External Attacks)

| Priority | Layer | Est. Effort | Status |
|----------|-------|-------------|--------|
| P0 | Security Headers Middleware | 2h |  |
| P0 | Input Validation (Zod schema) | 4h |  Partial |
| P0 | Rate Limiting | 4h |  Partial |
| P0 | Panic Recovery + Error Sanitization | 2h |  In golangci |
| P0 | HTTPS Enforcement (HSTS) | 1h |  |

### High (Should Have - Internal Security)

| Priority | Layer | Est. Effort | Status |
|----------|-------|-------------|--------|
| P1 | Argon2id Password Hashing | 4h |  Using bcrypt? |
| P1 | DOA (Ownership Queries) | 8h |  Need audit |
| P1 | Session Revocation | 4h |  |
| P1 | UUIDv7 Migration | 8h |  |
| P1 | Token Bucket Rate Limit | 4h |  Partial |

### Medium (Good Security Practice)

| Priority | Layer | Est. Effort | Status |
|----------|-------|-------------|--------|
| P2 | CSP Header | 2h |  |
| P2 | Audit Logging | 8h |  |
| P2 | Backup Encryption | 4h |  |
| P2 | User Enum Prevention | 2h |  |
| P2 | MFA/TOTP | 16h |  |

### Lower (Enterprise Features)

| Priority | Layer | Est. Effort | Status |
|----------|-------|-------------|--------|
| P3 | OAuth/OIDC Integration | 24h |  Partial |
| P3 | Secrets Vault (HashiCorp) | 16h |  |
| P3 | Container Hardening | 8h |  |
| P3 | SIEM Integration | 24h |  |

---

## Summary Checklist

### Immediate (This Sprint)

- [ ] Add security headers middleware to all endpoints
- [ ] Verify http.MaxBytesReader on all handlers
- [ ] Implement Turnstile verification
- [ ] Audit DOA implementation in all queries
- [ ] Add HSTS header
- [ ] Verify password hashing algorithm (Argon2id vs bcrypt)
- [ ] Check all queries use parameterization

### Next Sprint

- [ ] Implement user enumeration prevention (constant-time responses)
- [ ] Add UUIDv7 migration strategy
- [ ] Implement session revocation list
- [ ] Add comprehensive audit logging
- [ ] Verify rate limiting per session (not just IP)

### Future

- [ ] MFA/TOTP implementation
- [ ] OAuth/OIDC for Google/GitHub
- [ ] Secrets vault integration
- [ ] Container security hardening
- [ ] SIEM integration
- [ ] Penetration testing

---

## Open Questions for Discussion

1. **Hosting Platform**: Is Render providing enough security for your threat model?
2. **Compliance**: Do you need GDPR/SOC2 compliance?
3. **User Base**: Consumer app vs enterprise customers?
4. **Data Sensitivity**: What level of PII are you handling?
5. **Device Fleet**: Will you support Android with sensitive commands?
6. **Threat Model**: What specific attackers are you defending against?

---

*Document Version: 1.0*
*Last Updated: 2026-06-10*
*Next Review: After architecture discussion*
