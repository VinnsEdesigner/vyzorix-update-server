# API Key System

API keys are for service-to-service authentication — CLI tools, automation scripts, and integrations that need to call the API without a browser session.

## Key structure

Each API key has:

- **Name** — human-readable label (e.g., "CI Pipeline", "Monitoring Agent")
- **Prefix** — the first 8 characters of the key, stored in plaintext for display (`vxyz-` by default)
- **Secret** — the full key, stored as an argon2id hash (never plaintext)
- **Scope** — what the key can do: `read`, `write`, `admin`
- **Expiry** — optional expiration date
- **Monthly limit** — max API calls per month
- **Organization** — which org the key belongs to

## Creating a key

`POST /v1/api-keys` (requires operator auth + CSRF):

```json
{
  "name": "CI Pipeline",
  "scope": "write",
  "expires_in_days": 90
}
```

Response includes the full key string once — it's never shown again:

```json
{
  "id": "key-abc123",
  "name": "CI Pipeline",
  "api_key": "vxyz-prod-key-abcdef0123456789",
  "prefix": "vxyz-prod",
  "scope": "write",
  "expires_at": "2026-11-17T00:00:00Z"
}
```

The key is hashed with argon2id before storage. Only the prefix is stored in plaintext (for display in the dashboard).

## Using a key

Send the key in the `X-API-Key` header:

```
GET /v1/devices
X-API-Key: vxyz-prod-key-abcdef0123456789
```

The `TenantAPIKeyAuth` middleware in `internal/api/middleware/tenant_api_key.go`:

1. Extracts the key from the header
2. Extracts the prefix (first 8 chars)
3. Looks up keys by prefix in the database
4. For each match, verifies the full key against the stored argon2id hash
5. Checks expiry and monthly usage limit
6. Sets `operator_id` and `organization_id` in the gin context
7. Enforces scope (read keys can't POST, write keys can't DELETE, etc.)

## Scope enforcement

The `ScopeEnforcement` middleware checks the key's scope against the HTTP method:

| Scope | Allowed methods |
|-------|----------------|
| `read` | GET, HEAD |
| `write` | GET, POST, PATCH, HEAD |
| `admin` | All (GET, POST, PATCH, DELETE, HEAD) |

If the scope doesn't allow the method, it returns 403.

## Rate limiting

API keys have per-month usage limits. Each authenticated request increments a counter. When the limit is reached, subsequent requests return 429.

The `APIKeyRateLimitMiddleware` in `internal/api/middleware/api_key_rate_limiter.go` handles this.

## Key management

- `GET /v1/api-keys` — list keys for the current operator (shows prefix, not full key)
- `PATCH /v1/api-keys/:id` — update name or scope
- `DELETE /v1/api-keys/:id` — revoke (soft delete, key stops working immediately)
- `POST /v1/api-keys/:id/rotate` — generate a new secret, invalidate the old one

## SuperAdmin key management

SuperAdmins can manage keys across all operators:

- `GET /v1/admin/api-keys` — list all keys globally
- `GET /v1/admin/api-keys/operator/:operatorId` — list keys for a specific operator
- `DELETE /v1/admin/api-keys/:keyId` — force-revoke any key
- `GET /v1/admin/api-keys/stats` — global key usage statistics

## Audit

All key operations are audited:
- `api_key_created` — with key name, prefix, scope
- `api_key_updated` — with changes
- `api_key_revoked` — with key name
- `api_key_rotated` — with key name
- `api_key_failed` — failed authentication attempt (with prefix and reason)
