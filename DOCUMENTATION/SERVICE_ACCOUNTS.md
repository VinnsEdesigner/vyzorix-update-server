# Service Accounts

A service account is an identity for automation. It isn't a person — it's a
name your scripts, CI pipelines, and integrations authenticate as. It lives in
one organization, holds its own tokens, and carries the same read/write/admin
scope model as an operator's API keys, but it's deliberately *not* attached to
any human.

That separation is the point. When a key leaks, you rotate the service account's
token without touching a person's credentials. When someone leaves, their
operator account goes away and your automations keep running.

---

## The shape of it

Two levels: the account, and its tokens.

**Service account** — the identity.

| Field | Meaning |
|:------|:--------|
| `name` | What the account is for — "nightly report job", "deploy bot". |
| `org_id` | The organization it belongs to. Required. |
| `enabled` | A disabled account's tokens stop authenticating. |

**Token** — a credential belonging to an account. An account can have several,
so you can give each integration its own and revoke them independently.

| Field | Meaning |
|:------|:--------|
| `name` | Label for this token. |
| `scopes` | What the token may do — a list drawn from `read`, `write`, `admin`. |
| `expires_at` | Optional expiry. No expiry means it lives until revoked. |
| `key_prefix` | The visible, non-secret prefix. |
| `key_hash` | The stored hash. The full key is never stored. |

---

## The key is shown once

When you create or rotate a token, the response contains the **full key exactly
once**. After that, only the prefix and hash exist. If you lose it, you rotate —
you don't recover.

A token is usable when it is valid, not expired, and not revoked. Any one of
those failing means it authenticates nothing.

---

## Rotation

```
POST /v1/service-accounts/:id/tokens/:token_id/rotate
```

Rotation issues a fresh key for the token and returns it once. The old key stops
working. This is the routine you reach for when a key is exposed, or on a
schedule.

---

## Managing service accounts over REST

All routes require organization context.

| Method | Path | Does |
|:-------|:-----|:-----|
| `GET` | `/v1/service-accounts` | List the org's accounts. |
| `POST` | `/v1/service-accounts` | Create an account. |
| `GET` | `/v1/service-accounts/:id/tokens` | List an account's tokens. |
| `POST` | `/v1/service-accounts/:id/tokens` | Mint a token — returns the key once. |
| `DELETE` | `/v1/service-accounts/:id/tokens/:token_id` | Revoke a token. |
| `POST` | `/v1/service-accounts/:id/tokens/:token_id/rotate` | Rotate a token — returns a new key once. |

Reading needs `service_accounts.read`; writing needs `service_accounts.write`,
scoped to `org:*`.

---

## How a token authenticates a request

A service-account token is presented like an API key. On the way in, the server
resolves the token to its account, checks that the account is enabled and the
token is usable, and attaches the token's scopes to the request. From that point
the request is evaluated exactly like any other scoped request — the scoped RBAC
engine checks the action against the token's scopes.

Because service accounts have no operator, the auth middleware handles them
without a human identity: scope evaluation reads the token's own scope list
rather than looking up a membership. An account can be a member of exactly the
organization it belongs to, and no other.

---

## Why this exists next to operator API keys

Operator API keys hang off a person. That's the right default for a human
driving the dashboard. But automation shouldn't borrow a person's identity — it
makes revocation messy, audit confusing, and offboarding dangerous.

Service accounts give automation its own lane: separate credentials, separate
rotation, separate audit. A token's `last_used_at` tells you when the automation
last ran; revoking it tells you it won't run again.
