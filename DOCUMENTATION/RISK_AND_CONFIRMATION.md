# Risk Classification & Confirmation Flow

Not all commands are equal. Rebooting a device is riskier than checking its status. Factory resetting one is catastrophic. The risk system classifies commands into tiers and gates the dangerous ones behind a confirmation flow.

## Risk tiers

Five tiers, lowest to highest:

| Tier | What it means | Example commands |
|------|--------------|-----------------|
| `zero` | Pure read, no side effects | `device.status` |
| `low` | Minor, reversible | `WAKE_UP_UPDATER`, `CHECK_UPDATE`, `device.wake_up` |
| `medium` | Mutates user data (default for unknown commands) | (any unclassified command) |
| `high` | System config, org-wide impact | `device.reboot` |
| `critical` | Destructive, irreversible | `device.factory_reset` |

The catalog lives in `internal/domain/command/risk_catalog.go`. Each command has a `CommandRiskProfile` with its tier, whether confirmation is required, and how long a confirmation token stays valid.

## How authorization works

The `RiskEvaluator` takes a command name and an `ActorContext` and returns a `Decision`:

- **Allow** — the command can execute. This is what happens for tier `zero`, `low`, and `medium` commands.
- **RequireConfirmation** — the command needs either a confirmation token (high tier) or MFA + a token (critical tier). Without it, the handler returns 425 Too Early.
- **Deny** — not currently used by any command profile, but available if a command should be permanently blocked.

The decision logic, in `internal/domain/command/risk_evaluator.go`:

1. If the command is critical tier and the actor hasn't verified MFA → RequireConfirmation. A token alone isn't enough for irreversible operations.
2. If the command's profile has `RequiresConfirmation: true` and the actor hasn't confirmed → RequireConfirmation.
3. Otherwise → Allow.

The actor context carries: `OperatorID`, `OrgID`, `IsSuperAdmin`, `MFAVerified`, and `Confirmed`.

## The confirmation token flow

When a handler returns 425, the client needs to get a confirmation token and retry.

### Step 1: Request a token

```
POST /v1/devices/:imei/command/confirm
Content-Type: application/json
X-CSRF-Token: ...

{"command": "device.reboot"}
```

The handler checks if the command actually requires confirmation (low-risk commands get `confirmation_required: false`). If it does, it creates a `PendingConfirmation` with:
- A random token (UUID)
- Scoped to the operator, command, and device
- A TTL from the command's `ConfirmationTTL` (5 minutes for reboot, 2 minutes for factory reset)

Response:
```json
{
  "confirmation_token": "abc-123-def",
  "confirmation_required": true,
  "risk_tier": "high",
  "expires_at": 1724400000,
  "ttl_seconds": 300
}
```

### Step 2: Execute with the token

```
POST /v1/device/:imei/command
Content-Type: application/json

{"command": "device.reboot", "confirmation_token": "abc-123-def"}
```

The handler's `authorizeCommand` method calls `consumeConfirmation`, which calls the confirmation handler's `ConsumeForCommand`. That checks:
- Token exists in the database
- Not expired
- Not already consumed
- Matches the operator, command, and device

If all checks pass, the token is atomically marked consumed (via `UPDATE ... WHERE consumed_at IS NULL AND expires_at > ?`) and the command proceeds. If any check fails, the handler returns 425 with a specific message (expired, already used, mismatched, not found).

### Token properties

- **Single-use** — once consumed, it can never be used again
- **Scoped** — a token for `device.reboot` on device A won't work for `device.factory_reset` on device B
- **TTL-bounded** — expires after the command's risk profile TTL
- **Atomic consumption** — the SQL UPDATE prevents race conditions where two requests try to use the same token simultaneously

## MFA verification

The `MFAVerified` flag on `ActorContext` comes from the authenticated session. The cookie auth middleware sets `*session.Session` in the gin context. The handler checks `sess.MFAVerifiedAt != nil` to determine if MFA was completed during this session.

For critical-tier commands (`device.factory_reset`), the evaluator requires both MFA verification AND a valid confirmation token. Neither alone is sufficient.

## Audit trail

Every command execution attempt is audited. The `Logger.CommandExecuted` method writes an audit entry with:

- The operator who initiated it
- The device, command, and dispatch ID
- The risk tier
- The result (success, blocked, failure)
- The reason (if blocked — "confirmation required", "denied")
- The actor type ("operator") and email
- Old/new state (for change tracking — e.g., "pending" → "delivered")
- The trace ID from the request

These go into the `audit_logs` table with the `trace_id`, `risk_tier`, `actor_type`, `actor_email`, `old_value`, and `new_value` columns (migrations #57 and #59).

## Database schema

The `command_confirmations` table (migration #58):

```sql
CREATE TABLE command_confirmations (
  token       TEXT PRIMARY KEY,
  operator_id TEXT NOT NULL,
  org_id      TEXT,
  command     TEXT NOT NULL,
  device_id   TEXT,
  risk_tier   TEXT,
  created_at  DATETIME NOT NULL,
  expires_at  DATETIME NOT NULL,
  consumed_at DATETIME
);
```

Indexes on `operator_id, command` (for lookups) and `expires_at` (for cleanup).

## What was removed

The `domain/threat/` package is deleted. It had threat detection types (`ThreatType`, `Severity`, `ThreatAction`) that overlapped with the risk system's `RiskTier` and `Decision` but had zero callers anywhere in the codebase. The risk catalog and evaluator replace it with a concrete, wired-in system that actually runs on every command.
