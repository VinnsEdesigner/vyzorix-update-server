# Audit System

Every security-relevant action gets an audit entry. The entries live in the `audit_logs` SQLite table and carry enough context to answer "who did what, when, from where, and what changed."

## The audit Entry

Defined in `internal/audit/audit_repository.go`:

| Field | Type | Purpose |
|------|------|---------|
| `ID` | string | UUID, generated on write if empty |
| `CreatedAt` | time.Time | When the event happened |
| `OperatorID` | string | Who triggered it (operator ID) |
| `Action` | Action | What happened (login_success, command_executed, etc.) |
| `ResourceType` | string | What kind of thing was affected (device, api_key, session) |
| `ResourceID` | string | The specific resource ID |
| `IPAddress` | string | Where the request came from |
| `UserAgent` | string | The client's user agent |
| `Result` | Result | success, failure, blocked, pending |
| `TraceID` | string | Correlates with request logs |
| `RiskTier` | string | For command execution — the risk classification |
| `ActorType` | string | "operator", "api_key", or "system" |
| `ActorEmail` | string | Human-readable actor email (for audit queries) |
| `OldValue` | string | Before-state for change tracking |
| `NewValue` | string | After-state for change tracking |
| `Metadata` | map | Additional key-value pairs (command name, device ID, etc.) |

## Actions that get audited

The full list is in `internal/audit/audit_repository.go` as `Action` constants. The notable ones:

- `login_success`, `login_failed`, `logout`, `register`
- `command_executed` — every command execution attempt
- `api_key_created`, `api_key_revoked`, `api_key_rotated`
- `session_revoked`, `account_locked`
- `csrf_failure`, `signing_failure`, `rate_limit_exceeded`
- `update_pushed`, `update_cancelled`
- `settings_changed`

## How it writes

The `Logger` in `internal/audit/audit_logger.go` uses a buffered channel and a single writer goroutine. When a handler calls `l.LogEvent(ctx, &Entry{...})`, the entry goes into the channel. The writer goroutine drains the channel and writes to the SQLite `audit_logs` table.

If the channel buffer is full (10,000 events), it logs synchronously with a warning. On shutdown, `l.Shutdown(ctx)` flushes remaining events.

The audit repository writes to both the co-located database (same SQLite file as the app) and optionally a separate audit database (if `SeparateDB` is configured in `LoggerConfig`).

## Command execution audit

The most detailed audit trail is for command execution. The `CommandExecutedEvent` struct carries:

- `OperatorID`, `DeviceID`, `Command`, `DispatchID`
- `IPAddress`, `UserAgent`, `TraceID`
- `RiskTier` (zero through critical)
- `Result` (success, blocked, failure)
- `Reason` (if blocked — "confirmation required", "denied")
- `ActorType` ("operator"), `ActorEmail`
- `OldValue` / `NewValue` (state transition — e.g., "pending" → "delivered")

The `Logger.CommandExecuted` method writes this as an `audit_logs` row with `action = "command_executed"`.

## Login audit

The `Presenter.LoginSuccess` method in `internal/api/adapters/response/resp_presenter.go` calls `LogEvent` with an entry that includes the trace ID from the gin context and `ActorType: "operator"`. This runs in a goroutine so it doesn't block the response.

## Database columns

The `audit_logs` table has been extended across three migrations:

- Migration #18 — original table (id, operator_id, action, ip_address, user_agent, created_at)
- Migration #53 — resource_type, resource_id, metadata, result columns
- Migration #57 — trace_id, risk_tier columns
- Migration #59 — actor_type, actor_email, old_value, new_value columns

All column additions use idempotent `ALTER TABLE ADD COLUMN` (checking for "duplicate column" errors).

## PII in audit entries

The `ActorEmail` field stores the operator's email for human-readable audit queries. This is intentional — audit logs need to be readable by humans for compliance. The `OperatorID` is an opaque ID; the email makes it scannable.

Emails are NOT stored in the `Metadata` map — only in the dedicated `ActorEmail` column. The redaction system masks emails in log output (`slog` calls) but audit entries are persisted to the database unredacted, because they're security records that need full fidelity.

## What was fixed

1. The login handler wasn't calling the audit `LoginSuccess` method at all. Added `h.presenter.LoginSuccess(c, result.OperatorID)` before the response.

2. The `LoginSuccess` method used `context.Background()` and didn't pass the trace ID. Rewrote it to pull `trace_id` from the gin context and call `LogEvent` directly with a full `Entry` including `TraceID` and `ActorType`.
