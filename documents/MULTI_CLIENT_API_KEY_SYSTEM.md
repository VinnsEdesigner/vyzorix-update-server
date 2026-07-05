# Multi-Client API Key System - Implementation Document

**Version:** 2.0  
**Date:** 2026-07-05  
**Status:** Complete - Implementation Required  
**Frontend Reference:** [FRONTEND_API_KEYS_REQUIREMENTS.md](./FRONTEND_API_KEYS_REQUIREMENTS.md)

---

## Table of Contents

1. [Overview](#1-overview)
2. [Endpoint Authentication Architecture](#2-endpoint-authentication-architecture)
3. [Requirements Summary](#3-requirements-summary)
4. [Architecture](#4-architecture)
5. [Database Schema](#5-database-schema)
6. [Files Structure](#6-files-structure)
7. [API Endpoints](#7-api-endpoints)
8. [Middleware Implementation](#8-middleware-implementation)
9. [Service Layer](#9-service-layer)
10. [Security Implementation](#10-security-implementation)
11. [Rate Limiting](#11-rate-limiting)
12. [Audit Logging](#12-audit-logging)
13. [Testing Strategy](#13-testing-strategy)
14. [Migration Plan](#14-migration-plan)

---

## 1. Overview

### 1.1 Feature Summary

This document describes the implementation of a multi-client API key system that allows operators to generate, manage, and use multiple API keys for accessing the Vyzorix Update Server from different client applications.

**Key Characteristics:**
- Format: `vxyz_<random>` (prefix configurable via `API_KEY_PREFIX` env var)
- Keys are shown once at creation (copy-paste only)
- Immediate rotation (old key stops working immediately)
- Maximum 20 keys per operator per month
- Expiration settings per key (optional)
- Key scopes: read, write, admin
- Rate limiting: 100 req/min per key

**Architecture Principle:**
Different endpoint types use different authentication methods. API keys (database) never overlap with infrastructure keys (env vars).

### 1.2 Key Scopes

| Scope | HTTP Methods Allowed | Description |
|-------|---------------------|-------------|
| `read` | GET, HEAD, OPTIONS | Read-only access |
| `write` | GET, POST, PUT, PATCH, HEAD, OPTIONS | Read + write access |
| `admin` | ALL (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS) | Full access |

### 1.3 Answers to Design Questions

| Question | Answer |
|----------|--------|
| Rate limiting | Global limit: 100 req/min per key |
| Super admin capabilities | View ALL keys, revoke ANY key, see per-operator stats |
| Audit logging | Management events only (create, revoke, rotate, failed auth) |
| Scope enforcement | Middleware (automatic) + Handler-level (override capability) |

---

## 2. Endpoint Authentication Architecture

### 2.1 Authentication Matrix

| Endpoint Type | Path Pattern | Auth Method | Purpose |
|---------------|--------------|-------------|---------|
| **PUBLIC** | `/health` | None | Liveness probe |
| **PUBLIC** | `/v1/auth/*` | None | Login, register, password reset |
| **PUBLIC** | `/v1/device/register` | None | Device self-registration |
| **PUBLIC** | `/v1/device/:imei/status` | None | Public device status |
| **PUBLIC** | `/v1/device/inbox` | None | Device inbox submission |
| **PUBLIC** | `/v1/device/confirm` | None | Device confirmation |
| **INFRASTRUCTURE** | `/healthz` | Env API Key | Readiness probe |
| **INFRASTRUCTURE** | `/metrics` | Env API Key | Prometheus metrics |
| **INFRASTRUCTURE** | `/admin/*` | Env API Key | Server admin operations |
| **INFRASTRUCTURE** | `/internal/*` | Env API Key | Migration, backup ops |
| **SESSION ONLY** | `/bin/*` | Session Cookie | Binary downloads |
| **SESSION ONLY** | `/v1/dashboard/*` | Session Cookie | Dashboard data |
| **SESSION ONLY** | `/v1/api-keys/*` | Session Cookie | API key management |
| **SESSION ONLY** | `/api/v1/apk/*` | Session Cookie | APK downloads |
| **DEVICE AUTH** | `/v1/device/:imei/command` | HMAC Signature | Device receives commands |
| **DEVICE AUTH** | `/v1/device/:imei/fcm-token` | HMAC Signature | Device updates FCM |
| **TENANT** | `/v1/devices/*` | Session OR API Key + Scope | Device listing, deregistration |
| **TENANT** | `/v1/device/:imei/*` | Session OR API Key + Scope | Device management |
| **TENANT** | `/v1/command/*` | Session OR API Key + Scope | Command dispatch |
| **TENANT** | `/v1/telemetry/*` | Session OR API Key + Scope | Telemetry queries |
| **TENANT** | `/v1/updates/*` | Session OR API Key + Scope | Update pushes, sync |
| **TENANT** | `/v1/device/diagnostics/*` | Session OR API Key + Scope | Diagnostics |

### 2.2 Request Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    REQUEST INCOMING                          │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│              IS PATH /health OR /v1/auth/* OR               │
│              /v1/device/public/* ?                          │
│  if YES → Allow (PUBLIC)                                   │
└───────────────────────┬─────────────────────────────────────┘
                        │ No
                        ▼
┌─────────────────────────────────────────────────────────────┐
│              IS PATH /healthz OR /metrics OR                │
│              /admin/* OR /internal/* ?                     │
│  if YES → Check Env API Key                                │
│           if valid → Allow                                 │
│           if invalid → 401                                │
└───────────────────────┬─────────────────────────────────────┘
                        │ No
                        ▼
┌─────────────────────────────────────────────────────────────┐
│              IS PATH /bin/* OR /v1/dashboard/* OR          │
│              /v1/api-keys/* OR /api/v1/apk/* ?            │
│  if YES → Check Session Cookie                             │
│           if valid → Allow                                 │
│           if invalid → 401                                │
└───────────────────────┬─────────────────────────────────────┘
                        │ No
                        ▼
┌─────────────────────────────────────────────────────────────┐
│              IS PATH /v1/device/:imei/command OR             │
│              /v1/device/:imei/fcm-token ?                    │
│  if YES → Check HMAC Signature                            │
│           if valid → Allow                                 │
│           if invalid → 401                                │
└───────────────────────┬─────────────────────────────────────┘
                        │ No
                        ▼
┌─────────────────────────────────────────────────────────────┐
│              DEFAULT: TENANT OPERATIONS                     │
│              Check Session Cookie OR API Key                │
│              if valid → Check Scope (GET/POST/DELETE)      │
│              if valid_session → Allow                      │
│              if valid_api_key → ScopeEnforcementMiddleware  │
│                  → Allow if scope sufficient                │
│                  → 403 if scope insufficient                │
│              else → 401                                   │
└─────────────────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│              RATE LIMIT CHECK (for API key requests)         │
│              100 req/min per key                            │
│              if exceeded → 429 Too Many Requests             │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. Requirements Summary

| ID | Requirement | Priority |
|----|-------------|----------|
| REQ-01 | Users can generate one or more API keys from frontend | Must |
| REQ-02 | Each API key is unique per client application | Must |
| REQ-03 | Keys work for all tenant endpoint features | Must |
| REQ-04 | Keys stored hashed in database (Argon2id) | Must |
| REQ-05 | Keys validated against database | Must |
| REQ-06 | Full key shown only once at creation | Must |
| REQ-07 | API key format: `vxyz_<random>` | Must |
| REQ-08 | Copy-paste key display | Must |
| REQ-09 | Immediate rotation (old key stops) | Must |
| REQ-10 | 20 keys per operator per month limit | Must |
| REQ-11 | Env-var admin keys continue working for infrastructure | Must |
| REQ-12 | API key expiration settings | Must |
| REQ-13 | Key metadata: name, created, last used, request count | Must |
| REQ-14 | Key scopes: read, write, admin | Must |
| REQ-15 | Scope enforcement: middleware checks method vs scope | Must |
| REQ-16 | Scope enforcement: handler-level override capability | Must |
| REQ-17 | Rate limiting: 100 req/min per key | Must |
| REQ-18 | Audit logging: management events only | Must |
| REQ-19 | Super admin can view ALL keys from ALL operators | Must |
| REQ-20 | Super admin can revoke ANY operator's key | Must |
| REQ-21 | Super admin can see per-operator usage stats | Must |
| REQ-22 | PATCH endpoint for rename/scope change | Must |
| REQ-23 | Pagination on list endpoints (20 per page) | Must |
| REQ-24 | DB-level monthly limit enforcement | Must |
| REQ-25 | Synchronous request counting | Must |

---

## 4. Architecture

### 4.1 System Components

```
┌─────────────────────────────────────────────────────────────┐
│                        Client Apps                          │
│   (Web App, iOS App, Android App, 3rd Party, etc.)      │
└─────────────────────┬───────────────────────────────────────┘
                      │ X-API-Key: vxyz_xxx...
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                   API Gateway Layer                          │
│  ┌─────────────────────────────────────────────────────┐   │
│  │           CombinedAPIKeyMiddleware                  │   │
│  │  ┌─────────────┐  ┌─────────────────────────────┐  │   │
│  │  │ EnvVarKeys  │  │ DatabaseKeys (DBLookup)     │  │   │
│  │  │ (Infra)     │  │ (Tenant)                    │  │   │
│  │  └─────────────┘  └─────────────────────────────┘  │   │
│  │                                                      │   │
│  │  ┌─────────────────────────────────────────────┐  │   │
│  │  │       ScopeEnforcementMiddleware            │  │   │
│  │  │  GET → read scope                          │  │   │
│  │  │  POST/PUT/PATCH → write scope             │  │   │
│  │  │  DELETE → admin scope                      │  │   │
│  │  └─────────────────────────────────────────────┘  │   │
│  │                                                      │   │
│  │  ┌─────────────────────────────────────────────┐  │   │
│  │  │       RateLimitMiddleware                   │  │   │
│  │  │  100 req/min per key                       │  │   │
│  │  └─────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                      Handlers                               │
│  ┌─────────────────────────────────────────────────────┐   │
│  │           APIKeyManagementHandler                   │   │
│  │  - CreateKey, ListKeys, UpdateKey, RevokeKey      │   │
│  │  - RotateKey, GetKey                               │   │
│  │  - Requires authenticated session (cookie)         │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │           SuperAdminAPIKeyHandler                   │   │
│  │  - ListAllKeys, GetOperatorKeys, ForceRevoke       │   │
│  │  - GetGlobalStats, GetOperatorStats                 │   │
│  │  - Requires super_admin role                        │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 4.2 Middleware Stack (Order)

```
Request
   │
   ▼
┌──────────────────────┐
│ Global Middleware    │ (CORS, Logger, etc.)
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ API Key Auth         │ ←── CombinedAPIKeyAuth
│ (Env or DB lookup)  │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ Scope Enforcement    │ ←── Only for tenant API key auth
│ (read/write/admin)   │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ Rate Limit           │ ←── 100 req/min per key
│ (Redis or memory)    │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ Handler              │
└──────────────────────┘
```

---

## 5. Database Schema

### 5.1 New Table: `api_keys`

```sql
CREATE TABLE IF NOT EXISTS api_keys (
    id TEXT PRIMARY KEY,                     -- UUIDv7
    operator_id TEXT NOT NULL,              -- FK to operators
    name TEXT NOT NULL,                     -- User-defined label (max 64 chars)
    key_prefix TEXT NOT NULL,               -- First 8 chars of key for display
    key_hash TEXT NOT NULL,                 -- Argon2id hash of full key
    scope TEXT NOT NULL DEFAULT 'read',     -- read, write, or admin
    expires_at INTEGER,                     -- Unix ms, NULL = never
    is_active INTEGER NOT NULL DEFAULT 1,  -- Boolean
    request_count INTEGER NOT NULL DEFAULT 0,
    last_request_at INTEGER,                -- Unix ms
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    revoked_at INTEGER,                     -- Unix ms when revoked
    FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
);

CREATE INDEX idx_api_keys_operator_id ON api_keys(operator_id);
CREATE INDEX idx_api_keys_key_prefix ON api_keys(key_prefix);
CREATE INDEX idx_api_keys_expires_at ON api_keys(expires_at);
CREATE INDEX idx_api_keys_is_active ON api_keys(is_active);
CREATE INDEX idx_api_keys_created_at ON api_keys(created_at);
```

### 5.2 New Table: `api_key_usage_stats`

```sql
CREATE TABLE IF NOT EXISTS api_key_usage_stats (
    id TEXT PRIMARY KEY,
    operator_id TEXT NOT NULL,
    year_month TEXT NOT NULL,                -- "2026-07"
    keys_created INTEGER NOT NULL DEFAULT 0,
    total_requests INTEGER NOT NULL DEFAULT 0,
    UNIQUE(operator_id, year_month)
);

CREATE INDEX idx_api_key_usage_operator ON api_key_usage_stats(operator_id);
CREATE INDEX idx_api_key_usage_year_month ON api_key_usage_stats(year_month);
```

### 5.3 Table: `operators` (Add Columns)

```sql
ALTER TABLE operators ADD COLUMN api_key_count_this_month INTEGER NOT NULL DEFAULT 0;
ALTER TABLE operators ADD COLUMN api_key_count_reset_at INTEGER NOT NULL;
```

---

## 6. Files Structure

### 6.1 New Files to Create

| File | Purpose |
|------|---------|
| `apps/api/internal/domain/api_key.go` | APIKey entity, ApiKeyScope type, scope methods |
| `apps/api/internal/domain/api_key_repository.go` | Repository interface for API keys |
| `apps/api/internal/application/api_key_service.go` | Business logic for API key operations |
| `apps/api/internal/infrastructure/storage/039_api_keys.go` | Migration + APIKeyRepository implementation |
| `apps/api/internal/api/handlers/auth/api_key_handler.go` | HTTP handlers for operator API key CRUD |
| `apps/api/internal/api/handlers/admin/api_key_admin_handler.go` | HTTP handlers for super admin |
| `apps/api/internal/api/middleware/api_key.go` | Refactored: path-based auth, scope enforcement |
| `apps/api/internal/api/middleware/super_admin.go` | Super admin authorization middleware |
| `apps/api/internal/api/middleware/rate_limit.go` | Rate limiting middleware |

### 6.2 Files to Modify

| File | Changes |
|------|---------|
| `apps/api/internal/api/api_server.go` | Register new routes, inject dependencies |
| `apps/api/internal/infrastructure/config/config.go` | Add `APIKeyConfig` struct |
| `apps/api/internal/infrastructure/storage/sqlite.go` | Register new migration |

---

## 7. API Endpoints

### 7.1 Operator API Key Endpoints

| Method | Endpoint | Auth | Purpose |
|--------|----------|------|---------|
| POST | `/v1/auth/api-keys` | Session | Create new API key |
| GET | `/v1/auth/api-keys` | Session | List operator's keys (paginated) |
| GET | `/v1/auth/api-keys/:keyId` | Session | Get single key details |
| PATCH | `/v1/auth/api-keys/:keyId` | Session | Update key (rename, change scope) |
| DELETE | `/v1/auth/api-keys/:keyId` | Session | Revoke API key |
| POST | `/v1/auth/api-keys/:keyId/rotate` | Session | Rotate API key (new key) |

### 7.2 Super Admin API Key Endpoints

| Method | Endpoint | Auth | Purpose |
|--------|----------|------|---------|
| GET | `/v1/admin/api-keys` | Super Admin | List ALL keys (all operators) |
| GET | `/v1/admin/api-keys/stats` | Super Admin | Global API key statistics |
| GET | `/v1/admin/operators/:operatorId/api-keys` | Super Admin | Get all keys for specific operator |
| DELETE | `/v1/admin/api-keys/:keyId` | Super Admin | Force revoke any key |
| GET | `/v1/admin/api-keys/:keyId` | Super Admin | Get any key details |

### 7.3 Endpoint Specifications

All API responses use `snake_case` for field names (JSON API convention).

#### POST /v1/auth/api-keys - Create API Key

**Request:**
```json
{
    "name": "Production iOS App",
    "scope": "read",
    "expires_in_days": 90
}
```

**Response (201):**
```json
{
    "id": "01HX...",
    "name": "Production iOS App",
    "api_key": "vxyz_a1b2c3d4e5f6...",  // FULL KEY - only shown once!
    "key_prefix": "vxyz_a1b2",
    "scope": "read",
    "expires_at": "2026-10-03T00:00:00Z",
    "created_at": "2026-07-05T16:00:00Z"
}
```

**Errors:**
- 400: Invalid request (validation errors)
- 403: Monthly limit exceeded (20 keys)

**Validation:**
- `name`: required, 1-64 chars, alphanumeric + spaces/hyphens/underscores
- `scope`: required, one of "read", "write", "admin"
- `expires_in_days`: optional, 1-365 or null (no expiration)

#### GET /v1/auth/api-keys - List API Keys (Paginated)

**Query Parameters:**
- `page` (int, default: 1)
- `limit` (int, default: 20, max: 100)

**Response (200):**
```json
{
    "keys": [
        {
            "id": "01HX...",
            "name": "Production iOS App",
            "key_prefix": "vxyz_a1b2",
            "scope": "read",
            "expires_at": "2026-10-03T00:00:00Z",
            "is_active": true,
            "request_count": 1542,
            "last_request_at": "2026-07-05T14:30:00Z",
            "created_at": "2026-07-05T16:00:00Z"
        }
    ],
    "pagination": {
        "page": 1,
        "limit": 20,
        "total": 5,
        "total_pages": 1
    },
    "monthly_limit": 20,
    "keys_created_this_month": 3
}
```

#### GET /v1/auth/api-keys/:keyId - Get Single Key

**Response (200):**
```json
{
    "id": "01HX...",
    "name": "Production iOS App",
    "key_prefix": "vxyz_a1b2",
    "scope": "read",
    "expires_at": "2026-10-03T00:00:00Z",
    "is_active": true,
    "request_count": 1542,
    "last_request_at": "2026-07-05T14:30:00Z",
    "created_at": "2026-07-05T16:00:00Z",
    "updated_at": "2026-07-05T16:00:00Z",
    "revoked_at": null
}
```

#### PATCH /v1/auth/api-keys/:keyId - Update API Key

**Request:**
```json
{
    "name": "New Name",
    "scope": "write",
    "expires_in_days": 60
}
```

All fields are optional. Only provided fields are updated.

**Response (200):**
```json
{
    "id": "01HX...",
    "name": "New Name",
    "key_prefix": "vxyz_a1b2",
    "scope": "write",
    "expires_at": "2026-09-03T00:00:00Z",
    "is_active": true,
    "request_count": 1542,
    "last_request_at": "2026-07-05T14:30:00Z",
    "created_at": "2026-07-05T16:00:00Z",
    "updated_at": "2026-07-05T18:00:00Z",
    "revoked_at": null
}
```

**Validation:**
- `name`: 1-64 chars, alphanumeric + spaces/hyphens/underscores
- `scope`: one of "read", "write", "admin"
- `expires_in_days`: 1-365 or null (no expiration)

#### DELETE /v1/auth/api-keys/:keyId - Revoke API Key

**Response (204):** No content

#### POST /v1/auth/api-keys/:keyId/rotate - Rotate API Key

**Response (201):**
```json
{
    "id": "01HY...",
    "name": "Production iOS App",
    "api_key": "vxyz_z9y8x7w6v5u4...",  // NEW full key
    "key_prefix": "vxyz_z9y8",
    "scope": "read",
    "expires_at": "2026-10-03T00:00:00Z",
    "created_at": "2026-07-05T16:30:00Z"
}
```

Note: Old key is immediately invalidated.

#### GET /v1/admin/api-keys - Super Admin: List ALL Keys

**Query Parameters:**
- `page` (int, default: 1)
- `limit` (int, default: 20, max: 100)
- `operator_id` (string, optional, filter by operator)
- `search` (string, optional, search by key name)

**Response (200):**
```json
{
    "keys": [
        {
            "id": "01HX...",
            "operator_id": "op_xxx",
            "operator_name": "Acme Corp",
            "name": "Production iOS App",
            "key_prefix": "vxyz_a1b2",
            "scope": "read",
            "is_active": true,
            "request_count": 1542,
            "created_at": "2026-07-05T16:00:00Z"
        }
    ],
    "pagination": {
        "page": 1,
        "limit": 20,
        "total": 150,
        "total_pages": 8
    }
}
```

#### GET /v1/admin/api-keys/:keyId - Super Admin: Get Any Key

**Response (200):** Same as single key response above.

#### DELETE /v1/admin/api-keys/:keyId - Super Admin: Force Revoke

**Response (204):** No content

**Note:** Super admin can revoke any operator's key.

#### GET /v1/admin/operators/:operatorId/api-keys - Super Admin: Get Operator Keys

**Response (200):** Same as List Keys response but filtered by operator.

#### GET /v1/admin/api-keys/stats - Super Admin: Global Stats

**Response (200):**
```json
{
    "total_keys": 150,
    "active_keys": 145,
    "revoked_keys": 5,
    "total_requests_today": 50000,
    "total_requests_this_month": 1500000,
    "top_operators": [
        {
            "operator_id": "op_xxx",
            "operator_name": "Acme Corp",
            "active_keys": 10,
            "total_requests": 500000
        }
    ],
    "requests_by_scope": {
        "read": 1000000,
        "write": 400000,
        "admin": 100000
    }
}
```

---

## 8. Middleware Implementation

### 8.1 API Key Auth Middleware

```go
// apps/api/internal/api/middleware/api_key.go

type APIKeyAuth struct {
    repo         *storage.APIKeyRepository
    envKeys      map[string]string // loaded from env
    config       *config.APIKeyConfig
}

func (m *APIKeyAuth) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        path := c.Request.URL.Path
        
        // INFRASTRUCTURE: Only env API keys
        if isInfrastructurePath(path) {
            if !m.checkEnvKey(c) {
                c.AbortWithStatusJSON(401, gin.H{"error": "invalid_infrastructure_api_key"})
                return
            }
            c.Set("auth_type", "infrastructure")
            c.Next()
            return
        }
        
        // SESSION ONLY: Reject API keys, let session middleware handle
        if isSessionOnlyPath(path) {
            c.Next()
            return
        }
        
        // DEVICE AUTH: Let HMAC middleware handle
        if isDeviceAuthPath(path) {
            c.Next()
            return
        }
        
        // TENANT: Database API keys with scope enforcement
        apiKey := extractAPIKey(c)
        if apiKey == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "api_key_required"})
            return
        }
        
        key, err := m.repo.FindByKey(apiKey)
        if err != nil || key == nil || !key.IsActive() {
            // Log failed attempt
            audit.Log(audit.ActionAPIKeyFailed, ...)
            c.AbortWithStatusJSON(401, gin.H{"error": "invalid_api_key"})
            return
        }
        
        // Check expiration
        if key.IsExpired() {
            c.AbortWithStatusJSON(401, gin.H{"error": "expired_api_key"})
            return
        }
        
        // Store key in context for scope enforcement middleware
        c.Set("api_key", key)
        c.Set("api_key_id", key.ID)
        c.Set("operator_id", key.OperatorID)
        c.Set("auth_type", "tenant")
        c.Next()
    }
}
```

### 8.2 Scope Enforcement Middleware

```go
// ScopeEnforcement ensures API key has required scope for HTTP method
func ScopeEnforcement() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Only enforce for tenant API key auth (not session)
        keyVal, exists := c.Get("api_key")
        if !exists || keyVal == nil {
            // No API key, let next middleware handle (session or no auth)
            c.Next()
            return
        }
        
        apiKey := keyVal.(*domain.APIKey)
        requiredScope := methodToScope(c.Request.Method)
        
        if !apiKey.HasScope(requiredScope) {
            c.AbortWithStatusJSON(403, gin.H{
                "error": "insufficient_scope",
                "message": fmt.Sprintf("This key requires '%s' scope, but has '%s' scope", 
                    requiredScope, apiKey.Scope),
                "required_scope": string(requiredScope),
                "current_scope": string(apiKey.Scope),
            })
            return
        }
        
        c.Next()
    }
}

// methodToScope returns required scope for HTTP method
func methodToScope(method string) domain.ApiKeyScope {
    switch method {
    case "GET", "HEAD", "OPTIONS":
        return domain.ScopeRead
    case "POST", "PUT", "PATCH":
        return domain.ScopeWrite
    case "DELETE":
        return domain.ScopeAdmin
    default:
        return domain.ScopeAdmin
    }
}
```

### 8.3 Rate Limit Middleware

```go
// RateLimitMiddleware enforces 100 req/min per API key
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Only rate limit requests with API keys (not sessions)
        keyVal, exists := c.Get("api_key")
        if !exists || keyVal == nil {
            c.Next()
            return
        }
        
        apiKey := keyVal.(*domain.APIKey)
        
        result, err := limiter.Allow(c.Request.Context(), apiKey.ID)
        if err != nil {
            // On error, allow request but log
            log.Printf("Rate limiter error: %v", err)
            c.Next()
            return
        }
        
        // Set rate limit headers
        c.Header("X-RateLimit-Limit", "100")
        c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
        c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))
        
        if !result.Allowed {
            c.Header("Retry-After", strconv.Itoa(result.RetryAfter))
            c.AbortWithStatusJSON(429, gin.H{
                "error": "rate_limit_exceeded",
                "message": "Too many requests. Please retry later.",
                "retry_after": result.RetryAfter,
                "limit": 100,
            })
            return
        }
        
        c.Next()
    }
}
```

### 8.4 Super Admin Middleware

```go
// RequireSuperAdmin ensures user has super_admin role
func RequireSuperAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Check session-based auth for super_admin role
        operator, exists := c.Get("operator")
        if !exists {
            c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
            return
        }
        
        if !operator.IsSuperAdmin() {
            c.AbortWithStatusJSON(403, gin.H{"error": "super_admin_required"})
            return
        }
        
        c.Next()
    }
}
```

---

## 9. Service Layer

### 9.1 API Key Service

```go
// apps/api/internal/application/api_key_service.go

type APIKeyService struct {
    repo     *storage.APIKeyRepository
    password *password.Service
    config   *config.APIKeyConfig
}

type CreateAPIKeyInput struct {
    OperatorID   string
    Name         string
    Scope        domain.ApiKeyScope
    ExpiresInDays *int
}

type CreateAPIKeyOutput struct {
    Key   *domain.APIKey
    Secret string // Full key (only time this is returned)
}

func (s *APIKeyService) CreateAPIKey(ctx context.Context, input CreateAPIKeyInput) (*CreateAPIKeyOutput, error) {
    // 1. Check monthly limit (DB-level)
    canCreate, err := s.canCreateKey(ctx, input.OperatorID)
    if err != nil {
        return nil, err
    }
    if !canCreate {
        return nil, ErrMonthlyLimitExceeded
    }
    
    // 2. Generate key
    rawKey := generateAPIKey(s.config.Prefix)
    
    // 3. Hash the key
    hash, err := s.password.HashPassword(rawKey)
    if err != nil {
        return nil, err
    }
    
    // 4. Extract prefix
    prefix := rawKey[:8]
    
    // 5. Calculate expiry
    var expiresAt *int64
    if input.ExpiresInDays != nil && *input.ExpiresInDays > 0 {
        t := time.Now().AddDate(0, 0, *input.ExpiresInDays).UnixMilli()
        expiresAt = &t
    }
    
    // 6. Create key
    key := &domain.APIKey{
        ID:        generateUUID(),
        OperatorID: input.OperatorID,
        Name:      input.Name,
        KeyPrefix: prefix,
        KeyHash:   hash,
        Scope:     input.Scope,
        ExpiresAt: expiresAt,
        IsActive:  true,
        CreatedAt: time.Now().UnixMilli(),
        UpdatedAt: time.Now().UnixMilli(),
    }
    
    err = s.repo.Create(ctx, key)
    if err != nil {
        return nil, err
    }
    
    // 7. Update monthly stats
    s.incrementMonthlyStats(ctx, input.OperatorID)
    
    // 8. Log audit event
    audit.Log(audit.ActionAPIKeyCreated, input.OperatorID, key.ID, map[string]any{
        "name":      key.Name,
        "prefix":    key.KeyPrefix,
        "scope":     key.Scope,
    })
    
    return &CreateAPIKeyOutput{
        Key:   key,
        Secret: rawKey, // Only time we return this
    }, nil
}

func (s *APIKeyService) canCreateKey(ctx context.Context, operatorID string) (bool, error) {
    // Get operator's current month stats
    stats, err := s.repo.GetMonthlyStats(ctx, operatorID)
    if err != nil {
        return false, err
    }
    
    // Count active keys created this month
    count, err := s.repo.CountActiveKeysThisMonth(ctx, operatorID)
    if err != nil {
        return false, err
    }
    
    return count < s.config.MaxPerMonth, nil
}

func (s *APIKeyService) RecordRequest(ctx context.Context, keyID string) error {
    // Synchronous update for reliability
    return s.repo.IncrementRequestCount(ctx, keyID)
}
```

### 9.2 Key Validation

```go
func (s *APIKeyService) ValidateKey(ctx context.Context, rawKey string) (*domain.APIKey, error) {
    // Extract prefix from raw key
    if len(rawKey) < 8 {
        return nil, ErrInvalidKey
    }
    prefix := rawKey[:8]
    
    // Find potential matches by prefix
    candidates, err := s.repo.FindByPrefix(ctx, prefix)
    if err != nil {
        return nil, err
    }
    
    // Verify hash (constant-time)
    for _, key := range candidates {
        if s.password.VerifyPassword(rawKey, key.KeyHash) == nil {
            // Valid key found
            return key, nil
        }
    }
    
    return nil, ErrInvalidKey
}
```

### 9.3 Rotate Key

```go
func (s *APIKeyService) RotateKey(ctx context.Context, keyID, operatorID string) (*CreateAPIKeyOutput, error) {
    // 1. Get existing key
    existing, err := s.repo.GetByID(ctx, keyID)
    if err != nil {
        return nil, err
    }
    
    // 2. Verify ownership
    if existing.OperatorID != operatorID {
        return nil, ErrNotFound
    }
    
    // 3. Revoke old key
    existing.IsActive = false
    existing.RevokedAt = time.Now().UnixMilli()
    err = s.repo.Update(ctx, existing)
    if err != nil {
        return nil, err
    }
    
    // 4. Create new key with same properties
    return s.CreateAPIKey(ctx, CreateAPIKeyInput{
        OperatorID:   operatorID,
        Name:         existing.Name,
        Scope:        existing.Scope,
        ExpiresInDays: nil, // Reset expiry
    })
}
```

### 9.4 Input Validation

```go
import (
    "errors"
    "regexp"
    "strings"
)

var (
    ErrInvalidName    = errors.New("invalid name")
    ErrInvalidScope   = errors.New("invalid scope")
    ErrInvalidExpiry  = errors.New("invalid expiry days")
    ErrNameTooLong    = errors.New("name must be 64 characters or less")
    ErrNameInvalid    = errors.New("name contains invalid characters")
)

var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`)

// ValidateCreateInput validates API key creation input
func ValidateCreateInput(input CreateAPIKeyInput) error {
    // Validate name
    if err := ValidateName(input.Name); err != nil {
        return err
    }
    
    // Validate scope
    if err := ValidateScope(input.Scope); err != nil {
        return err
    }
    
    // Validate expiry
    if err := ValidateExpiry(input.ExpiresInDays); err != nil {
        return err
    }
    
    return nil
}

// ValidateName validates key name
func ValidateName(name string) error {
    if len(name) == 0 {
        return ErrInvalidName
    }
    if len(name) > 64 {
        return ErrNameTooLong
    }
    if !nameRegex.MatchString(name) {
        return ErrNameInvalid
    }
    return nil
}

// ValidateScope validates API key scope
func ValidateScope(scope ApiKeyScope) error {
    switch scope {
    case ScopeRead, ScopeWrite, ScopeAdmin:
        return nil
    default:
        return ErrInvalidScope
    }
}

// ValidateExpiry validates expiry days
func ValidateExpiry(days *int) error {
    if days == nil {
        return nil // No expiry is valid
    }
    if *days < 1 || *days > 365 {
        return ErrInvalidExpiry
    }
    return nil
}

// ParseScopeFromString parses scope from string (case-insensitive)
func ParseScopeFromString(s string) (ApiKeyScope, error) {
    switch strings.ToLower(s) {
    case "read":
        return ScopeRead, nil
    case "write":
        return ScopeWrite, nil
    case "admin":
        return ScopeAdmin, nil
    default:
        return "", ErrInvalidScope
    }
}
```

---

## 10. Security Implementation

### 10.1 Key Generation

```go
const (
    KeyPrefixLength = 8
    KeyRandomLength = 32 // bytes
    KeyFormat       = "vxyz_%s" // prefix + base64url random
)

func generateAPIKey(prefix string) string {
    randomBytes := make([]byte, KeyRandomLength)
    if _, err := rand.Read(randomBytes); err != nil {
        panic(err) // Should not happen
    }
    return fmt.Sprintf(KeyFormat, base64.URLEncoding.EncodeToString(randomBytes))
}
```

### 10.2 Key Hashing (Argon2id)

```go
// Using existing password service with Argon2id
func (s *PasswordService) HashPassword(password string) (string, error) {
    hash, err := argon2id.HashPassword(
        []byte(password),
        &argon2id.Params{
            Memory:      64 * 1024, // 64 MB
            Iterations:  3,
            Parallelism: 4,
            SaltLen:     16,
            KeyLen:      32,
        },
    )
    if err != nil {
        return "", err
    }
    return string(hash), nil
}

func (s *PasswordService) VerifyPassword(password, hash string) error {
    return argon2id.VerifyPassword(password, hash)
}
```

### 10.3 Prefix Extraction

```go
func extractPrefix(rawKey string) string {
    if len(rawKey) < KeyPrefixLength {
        return ""
    }
    return rawKey[:KeyPrefixLength]
}
```

---

## 11. Rate Limiting

### 11.1 Rate Limiter Implementation

```go
type RateLimiter struct {
    store  *redis.Client
    limit  int // 100
    window time.Duration // 1 minute
}

type RateLimitResult struct {
    Allowed    bool
    Remaining  int
    ResetAt    time.Time
    RetryAfter int // seconds
}

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
    return &RateLimiter{
        store:  redisClient,
        limit:  100,
        window: time.Minute,
    }
}

func (r *RateLimiter) Allow(ctx context.Context, keyID string) (*RateLimitResult, error) {
    now := time.Now()
    windowStart := now.Truncate(r.window)
    windowKey := fmt.Sprintf("ratelimit:apikey:%s:%d", keyID, windowStart.Unix())
    
    // Increment counter atomically
    count, err := r.store.Incr(ctx, windowKey).Result()
    if err != nil {
        return nil, err
    }
    
    // Set expiry on first request in window
    if count == 1 {
        r.store.Expire(ctx, windowKey, r.window)
    }
    
    resetAt := windowStart.Add(r.window)
    remaining := r.limit - int(count)
    
    if count > int64(r.limit) {
        retryAfter := int(resetAt.Sub(now).Seconds())
        return &RateLimitResult{
            Allowed:    false,
            Remaining:  0,
            ResetAt:    resetAt,
            RetryAfter: retryAfter,
        }, nil
    }
    
    return &RateLimitResult{
        Allowed:    true,
        Remaining:  remaining,
        ResetAt:    resetAt,
        RetryAfter: 0,
    }, nil
}
```

### 11.2 In-Memory Fallback

```go
// For deployments without Redis
type InMemoryRateLimiter struct {
    mu    sync.RWMutex
    data  map[string]*windowedCounter
    limit int
    window time.Duration
}

type windowedCounter struct {
    count     int64
    windowEnd time.Time
}

func (r *InMemoryRateLimiter) Allow(ctx context.Context, keyID string) (*RateLimitResult, error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    now := time.Now()
    windowEnd := now.Add(r.window).Truncate(r.window).Add(r.window)
    
    counter, exists := r.data[keyID]
    if !exists || now.After(counter.windowEnd) {
        counter = &windowedCounter{windowEnd: windowEnd}
        r.data[keyID] = counter
    }
    
    counter.count++
    
    if counter.count > int64(r.limit) {
        retryAfter := int(counter.windowEnd.Sub(now).Seconds())
        return &RateLimitResult{
            Allowed:    false,
            Remaining:  0,
            ResetAt:    counter.windowEnd,
            RetryAfter: retryAfter,
        }, nil
    }
    
    return &RateLimitResult{
        Allowed:    true,
        Remaining:  r.limit - int(counter.count),
        ResetAt:    counter.windowEnd,
        RetryAfter: 0,
    }, nil
}
```

---

## 12. Audit Logging

### 12.1 Audit Actions (Management Events Only)

```go
const (
    ActionAPIKeyCreated   Action = "api_key_created"
    ActionAPIKeyUpdated   Action = "api_key_updated"
    ActionAPIKeyRevoked   Action = "api_key_revoked"
    ActionAPIKeyRotated   Action = "api_key_rotated"
    ActionAPIKeyFailed    Action = "api_key_failed"
)
```

### 12.2 Audit Log Entry Format

```json
{
    "id": "01HX...",
    "operator_id": "op_xxx",
    "action": "api_key_created",
    "resource_type": "api_key",
    "resource_id": "key_xxx",
    "metadata": {
        "key_id": "key_xxx",
        "key_name": "Production iOS App",
        "key_prefix": "vxyz_a1b2",
        "scope": "read",
        "expires_at": "2026-10-03T00:00:00Z"
    },
    "ip_address": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
    "created_at": "2026-07-05T16:00:00Z"
}
```

### 12.3 Audit Logging Service

```go
func (s *AuditService) Log(action Action, operatorID string, resourceID string, metadata map[string]any) {
    entry := &AuditEntry{
        ID:          generateUUID(),
        OperatorID:  operatorID,
        Action:      action,
        ResourceType: "api_key",
        ResourceID:  resourceID,
        Metadata:    metadata,
        IPAddress:   getClientIP(),
        UserAgent:   getUserAgent(),
        CreatedAt:   time.Now().UnixMilli(),
    }
    
    // Async write to not block request
    go s.writeEntry(entry)
}
```

**Note:** We do NOT log `ActionAPIKeyUsed` for every request. Only management events are logged to avoid log noise and storage bloat.

---

## 13. Testing Strategy

### 13.1 Unit Tests

| Component | Test Cases |
|-----------|------------|
| `GenerateAPIKey()` | Correct format, uniqueness, prefix |
| `HashPassword()` / `VerifyPassword()` | Hash verification works |
| `ValidateKey()` | Valid key passes, invalid fails, expired fails |
| `ScopeEnforcement()` | GET passes for all scopes, DELETE fails for read scope |
| `RateLimiter.Allow()` | Under limit passes, over limit blocked |
| `MonthlyLimitService` | Limit enforced at 20 |

### 13.2 Integration Tests

| Scenario | Expected Result |
|----------|-----------------|
| Valid API key request (GET) | 200 OK |
| Valid API key request (DELETE) with read scope | 403 Forbidden |
| Invalid API key | 401 Unauthorized |
| Expired API key | 401 Unauthorized |
| Revoked API key | 401 Unauthorized |
| Monthly limit exceeded | 403 Forbidden |
| Create key → immediate use | 200 OK |
| Rotate key → old key fails | 401 on old, 200 on new |
| Rate limit exceeded | 429 Too Many Requests |

---

## 14. Migration Plan

### 14.1 Phase 1: Database Migration

- Run migration 039_api_keys
- Add `scope` column to `api_keys` table
- Add `api_key_usage_stats` table
- Add columns to `operators` table

### 14.2 Phase 2: Deploy New Middleware

- Deploy path-based validation middleware
- Existing env-var keys continue for infrastructure endpoints
- Tenant endpoints now check database keys
- Add scope enforcement middleware
- Add rate limiting middleware (100 req/min)

### 14.3 Phase 3: Deploy Key Management + Super Admin

- Add `/v1/auth/api-keys` endpoints (CRUD + rotate)
- Add `/v1/admin/api-keys` endpoints (super admin)
- Frontend integration
- Operators can create/revoke their own keys
- Super admin can view/revoke any key

### 14.4 Phase 4: Deprecation (Optional)

- After 6 months, warn about env-var keys
- Eventually only DB keys supported (major version bump)

---

## Appendix A: Configuration Options

| Env Variable | Default | Description |
|--------------|---------|-------------|
| `API_KEY_PREFIX` | `vxyz` | Key prefix |
| `API_KEY_MAX_PER_MONTH` | `20` | Max keys per operator per month |
| `API_KEY_MAX_NAME_LENGTH` | `64` | Max key name length |
| `API_KEY_DEFAULT_EXPIRY_DAYS` | `0` | Default expiry (0 = never) |
| `API_KEY_EXPIRY_MAX_DAYS` | `365` | Max expiry days |
| `API_KEY_RATE_LIMIT` | `100` | Rate limit per key (req/min) |
| `API_KEY_PREFIX_LENGTH` | `8` | Prefix length for display |

---

## Appendix B: Error Codes

| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `invalid_request` | Malformed request |
| 400 | `validation_error` | Field validation errors |
| 401 | `api_key_required` | No API key provided |
| 401 | `invalid_api_key` | Key not found or invalid |
| 401 | `expired_api_key` | Key has expired |
| 403 | `insufficient_scope` | Key doesn't have required scope |
| 403 | `monthly_limit_exceeded` | Too many keys created this month |
| 403 | `super_admin_required` | Requires super admin role |
| 404 | `key_not_found` | Key ID doesn't exist |
| 409 | `key_name_conflict` | Operator already has key with this name |
| 429 | `rate_limit_exceeded` | Too many requests |

---

## Appendix C: Security Considerations

1. **Never log full API keys** - Only `key_prefix` (first 8 chars)
2. **Rate limiting** - Prevent brute-force attacks on key validation
3. **Argon2id** - Memory-hard hashing prevents GPU attacks
4. **Constant-time comparison** - Prevents timing attacks
5. **Audit logging** - Track management events only (not every request)
6. **Immediate revocation** - Rotated/revoked keys fail immediately
7. **HTTPS only** - Keys should only be transmitted over TLS
8. **Scope enforcement** - Least privilege principle

---

*Document Version: 2.0*  
*Status: Complete - All issues addressed as mandatory requirements*
