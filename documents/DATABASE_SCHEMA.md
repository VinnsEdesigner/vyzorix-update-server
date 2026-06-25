# Database Schema - Enterprise Implementation Specification

> **Version:** 2.0
> **Status:** Draft
> **Created:** 2026-06-25
> **Updated:** 2026-06-25
> **Target:** Production MVP
> **Database:** SQLite (WAL mode)

---

## Table of Contents

1. [Overview](#1-overview)
2. [Migration Versions](#2-migration-versions)
3. [Existing Schema (v1-v19)](#3-existing-schema-v1-v19)
4. [New Schema (v20+)](#4-new-schema-v20)
5. [Schema Modifications](#5-schema-modifications)
6. [Schema Diagram](#6-schema-diagram)
7. [Entity Relationships](#7-entity-relationships)
8. [Full Index List](#8-full-index-list)
9. [Migration Code](#9-migration-code)

---

## 1. Overview

The Vyzorix Update Server uses **SQLite** with WAL (Write-Ahead Logging) mode for persistence. All migrations are managed via Go code in `internal/infrastructure/storage/sqlite.go`.

### 1.1 Design Decisions

| Decision | Rationale |
|----------|-----------|
| **SQLite over PostgreSQL** | Single-file DB, no external dependencies, sufficient for 10K devices |
| **WAL Mode** | Better concurrency for read-heavy workloads |
| **2GB Cache** | Adequate for telemetry time-series data |
| **Idempotent Migrations** | ALTER TABLE ignores duplicate column additions |

### 1.2 Technology Stack

- **Database:** SQLite 3 with WAL mode
- **ORM:** Raw SQL via `database/sql`
- **Driver:** `github.com/mattn/go-sqlite3`
- **Migration Location:** `internal/infrastructure/storage/sqlite.go`

---

## 2. Migration Versions

| Version | Table/Change | Status |
|---------|--------------|--------|
| 1 | devices | ✅ Existing |
| 2 | telemetry | ✅ Existing |
| 3 | commands | ✅ Existing |
| 4 | operators | ✅ Existing |
| 5 | auth_sessions | ✅ Existing |
| 6 | email_verifications | ✅ Existing |
| 7 | password_reset_tokens | ✅ Existing |
| 8 | settings | ✅ Existing |
| 9 | commands (add columns) | ✅ Existing |
| 10 | devices (command_secret_hash) | ✅ Existing |
| 11 | operators (github_id) | ✅ Existing |
| 12 | password_reset_resend_tracker | ✅ Existing |
| 13 | api_clients | ✅ Existing |
| 14 | signing_keys | ✅ Existing |
| 15 | session_revocations | ✅ Existing |
| 16 | failed_login_attempts | ✅ Existing |
| 17 | account_lockouts | ✅ Existing |
| 18 | audit_logs | ✅ Existing |
| 19 | message_queue | ✅ Existing |
| **20** | **events** | **NEW** |
| **21** | **inbox_requests, registration_logs** | **NEW** |
| **22** | **device_logs, device_events** | **NEW** |
| **23** | **update_versions, update_pushes, update_push_devices** | **NEW** |
| **24** | **operator_settings** | **NEW** |
| **25** | **refresh_tokens** | **NEW** |
| **26** | **notification_audit_log** | **NEW** |
| **27** | **devices (add columns)** | **ALTER** |

---

## 3. Existing Schema (v1-v19)

### 3.1 devices

```sql
CREATE TABLE devices (
    id                      TEXT PRIMARY KEY,
    firebase_install_id      TEXT NOT NULL,
    fcm_token               TEXT,
    app_version              TEXT,
    device_class             TEXT,
    command_secret           TEXT NOT NULL,
    command_secret_hash      TEXT,
    online                   INTEGER NOT NULL DEFAULT 0,
    operator_id              TEXT,
    registered_at           INTEGER NOT NULL,
    last_seen               INTEGER NOT NULL,
    created_at              INTEGER NOT NULL,
    updated_at              INTEGER NOT NULL
);
```

### 3.2 telemetry

```sql
CREATE TABLE telemetry (
    id              TEXT PRIMARY KEY,
    device_id       TEXT NOT NULL,
    received_at     INTEGER NOT NULL,
    payload         TEXT NOT NULL,
    risk_score      INTEGER,
    buffer_level    INTEGER,
    thermal_temp    REAL,
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
);

CREATE INDEX idx_telemetry_device_time ON telemetry(device_id, received_at DESC);
```

### 3.3 commands

```sql
CREATE TABLE commands (
    id              TEXT PRIMARY KEY,
    dispatch_id     TEXT NOT NULL UNIQUE,
    device_id       TEXT NOT NULL,
    type            TEXT NOT NULL,
    args            TEXT,
    created_at      INTEGER NOT NULL,
    expires_at      INTEGER NOT NULL,
    delivered_at    INTEGER,
    completed_at    INTEGER,
    failed_at       INTEGER,
    status          TEXT NOT NULL DEFAULT 'pending',
    failure_reason  TEXT,
    wake_sent       INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
);
```

### 3.4 operators

```sql
CREATE TABLE operators (
    id                  TEXT PRIMARY KEY,
    email               TEXT NOT NULL UNIQUE,
    name                TEXT NOT NULL,
    password_hash       TEXT,
    role                TEXT NOT NULL DEFAULT 'operator',
    mfa_secret          TEXT,
    mfa_enabled         INTEGER NOT NULL DEFAULT 0,
    backup_codes        TEXT,
    email_verified      INTEGER NOT NULL DEFAULT 0,
    google_id           TEXT,
    github_id           TEXT,
    thresholds          TEXT,
    client_settings     TEXT,
    created_at          INTEGER NOT NULL,
    updated_at          INTEGER NOT NULL
);
```

### 3.5 auth_sessions

```sql
CREATE TABLE auth_sessions (
    id              TEXT PRIMARY KEY,
    operator_id     TEXT NOT NULL,
    token_hash      TEXT NOT NULL UNIQUE,
    expires_at      INTEGER NOT NULL,
    created_at      INTEGER NOT NULL,
    user_agent      TEXT,
    ip_address      TEXT,
    FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
);
```

### 3.6 email_verifications

```sql
CREATE TABLE email_verifications (
    id              TEXT PRIMARY KEY,
    operator_id     TEXT NOT NULL,
    token_hash      TEXT NOT NULL UNIQUE,
    expires_at      INTEGER NOT NULL,
    created_at      INTEGER NOT NULL,
    FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
);
```

### 3.7 password_reset_tokens

```sql
CREATE TABLE password_reset_tokens (
    id              TEXT PRIMARY KEY,
    operator_id     TEXT NOT NULL,
    token_hash      TEXT NOT NULL UNIQUE,
    expires_at      INTEGER NOT NULL,
    used_at         INTEGER,
    created_at      INTEGER NOT NULL,
    FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
);
```

### 3.8 settings

```sql
CREATE TABLE settings (
    key             TEXT PRIMARY KEY,
    value           TEXT NOT NULL,
    updated_at      INTEGER NOT NULL
);
```

### 3.9 api_clients

```sql
CREATE TABLE api_clients (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    api_key_hash    TEXT NOT NULL,
    hmac_key_hash   TEXT,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);
```

### 3.10 signing_keys

```sql
CREATE TABLE signing_keys (
    id              TEXT PRIMARY KEY,
    client_id       TEXT NOT NULL,
    key_hash        TEXT NOT NULL,
    created_at      INTEGER NOT NULL,
    expires_at      INTEGER,
    FOREIGN KEY(client_id) REFERENCES api_clients(id) ON DELETE CASCADE
);
```

### 3.11 session_revocations

```sql
CREATE TABLE session_revocations (
    session_id      TEXT PRIMARY KEY,
    revoked_at      INTEGER NOT NULL,
    reason          TEXT
);
```

### 3.12 failed_login_attempts

```sql
CREATE TABLE failed_login_attempts (
    email           TEXT NOT NULL,
    ip_address      TEXT NOT NULL,
    attempted_at    INTEGER NOT NULL,
    success         INTEGER NOT NULL DEFAULT 0
);
```

### 3.13 account_lockouts

```sql
CREATE TABLE account_lockouts (
    email               TEXT PRIMARY KEY,
    locked_until        INTEGER,
    attempt_count       INTEGER NOT NULL DEFAULT 0,
    first_attempt_at    INTEGER NOT NULL
);
```

### 3.14 audit_logs

```sql
CREATE TABLE audit_logs (
    id              TEXT PRIMARY KEY,
    operator_id     TEXT,
    action          TEXT NOT NULL,
    details         TEXT,
    ip_address      TEXT,
    user_agent      TEXT,
    created_at      INTEGER NOT NULL,
    FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
);
```

### 3.15 message_queue

```sql
CREATE TABLE message_queue (
    id              TEXT PRIMARY KEY,
    device_id       TEXT NOT NULL,
    frame_json      TEXT NOT NULL,
    enqueued_at     INTEGER NOT NULL,
    expires_at      INTEGER NOT NULL,
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
);

CREATE INDEX idx_message_queue_device_expires ON message_queue(device_id, expires_at);
```

### 3.16 password_reset_resend_tracker

```sql
CREATE TABLE password_reset_resend_tracker (
    id              TEXT PRIMARY KEY,
    email_hash      TEXT NOT NULL UNIQUE,
    resend_count    INTEGER NOT NULL DEFAULT 1,
    last_resend_at  INTEGER NOT NULL,
    lockout_until   INTEGER,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);
```

---

## 4. New Schema (v20+)

### 4.1 events (Migration 20)

Real-time event logging for WebSocket broadcasts and system alerts.

```sql
CREATE TABLE events (
    id              TEXT PRIMARY KEY,
    device_id       TEXT NOT NULL,
    event_type      TEXT NOT NULL,
    severity        TEXT NOT NULL DEFAULT 'info',
    message         TEXT,
    metadata        TEXT,
    created_at      INTEGER NOT NULL,
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
);

CREATE INDEX idx_events_device_time ON events(device_id, created_at DESC);
CREATE INDEX idx_events_type ON events(event_type, created_at DESC);
```

**Event Types:**

| Event Type | Severity | Description |
|------------|----------|-------------|
| `DEVICE_CONNECTED` | info | Device established WebSocket connection |
| `DEVICE_DISCONNECTED` | info | Device WebSocket connection closed |
| `DEVICE_AUTH_FAILED` | warning | Device authentication failed |
| `THRESHOLD_BREACH_RISK` | warning | Risk score exceeded threshold |
| `THRESHOLD_BREACH_THERMAL` | critical | Temperature exceeded threshold |
| `THRESHOLD_BREACH_BUFFER` | warning | Buffer level exceeded threshold |
| `THRESHOLD_BREACH_LATENCY` | warning | Latency exceeded threshold |
| `COMMAND_DELIVERED` | info | Command successfully delivered |
| `COMMAND_FAILED` | error | Command delivery/execution failed |
| `COMMAND_EXPIRED` | warning | Command expired before delivery |
| `FCM_FALLBACK_ACTIVATED` | warning | Device switched to FCM mode |
| `DEVICE_OFFLINE` | info | Device marked offline |
| `DEVICE_ONLINE` | info | Device marked online |

### 4.2 inbox_requests (Migration 21)

Device registration inbox for operator approval workflow.

```sql
CREATE TABLE inbox_requests (
    id                      TEXT PRIMARY KEY,
    imei                    TEXT NOT NULL UNIQUE,
    model                   TEXT,
    manufacturer            TEXT,
    os_version              TEXT,
    app_version             TEXT,
    fcm_token               TEXT,
    firebase_install_id      TEXT,
    status                  TEXT NOT NULL DEFAULT 'pending',
    created_at              INTEGER NOT NULL,
    approved_at             INTEGER,
    rejected_at              INTEGER,
    command_secret           TEXT,
    notes                   TEXT,
    operator_id              TEXT,
    FOREIGN KEY(operator_id) REFERENCES operators(id)
);

CREATE INDEX idx_inbox_pending ON inbox_requests(status, created_at DESC);
```

**Status Values:** `pending`, `acknowledged`, `approving`, `registered`, `rejected`, `expired`

### 4.3 registration_logs (Migration 21)

Audit trail for device registration actions.

```sql
CREATE TABLE registration_logs (
    id              TEXT PRIMARY KEY,
    device_id       TEXT NOT NULL,
    action          TEXT NOT NULL,
    operator_id     TEXT,
    details         TEXT,
    created_at      INTEGER NOT NULL,
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE,
    FOREIGN KEY(operator_id) REFERENCES operators(id)
);

CREATE INDEX idx_registration_logs_device ON registration_logs(device_id, created_at DESC);
```

### 4.4 device_logs (Migration 22)

Structured device event logs for diagnostics.

```sql
CREATE TABLE device_logs (
    id              TEXT PRIMARY KEY,
    device_id       TEXT NOT NULL,
    event_type      TEXT NOT NULL,
    timestamp       INTEGER NOT NULL,
    data            TEXT,
    created_at      INTEGER NOT NULL,
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
);

CREATE INDEX idx_device_logs ON device_logs(device_id, timestamp DESC);
CREATE INDEX idx_device_logs_cursor ON device_logs(device_id, timestamp DESC, id);
CREATE INDEX idx_device_logs_event_type ON device_logs(event_type);
```

### 4.5 device_events (Migration 22)

Timeline events for diagnostics page.

```sql
CREATE TABLE device_events (
    id              TEXT PRIMARY KEY,
    device_id       TEXT NOT NULL,
    event_type      TEXT NOT NULL,
    timestamp       INTEGER NOT NULL,
    data            TEXT,
    created_at      INTEGER NOT NULL,
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
);

CREATE INDEX idx_device_events_device_timestamp ON device_events(device_id, timestamp DESC);
CREATE INDEX idx_device_events_cursor ON device_events(device_id, timestamp DESC, id);
CREATE INDEX idx_device_events_type ON device_events(event_type, timestamp DESC);
```

### 4.6 update_versions (Migration 23)

APK version storage for update system.

```sql
CREATE TABLE update_versions (
    id              TEXT PRIMARY KEY,
    version         TEXT NOT NULL UNIQUE,
    apk_filename    TEXT NOT NULL,
    apk_size         INTEGER NOT NULL,
    sha256           TEXT NOT NULL,
    release_date    INTEGER NOT NULL,
    release_notes    TEXT,
    release_type    TEXT NOT NULL DEFAULT 'minor',
    is_latest       INTEGER NOT NULL DEFAULT 0,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL
);

CREATE INDEX idx_versions_date ON update_versions(release_date DESC);
```

**Release Types:** `major`, `minor`, `patch`

### 4.7 update_pushes (Migration 23)

Update push tracking for batch operations.

```sql
CREATE TABLE update_pushes (
    id              TEXT PRIMARY KEY,
    version_id      TEXT NOT NULL,
    install_type    TEXT NOT NULL DEFAULT 'immediate',
    scheduled_at    INTEGER,
    status          TEXT NOT NULL DEFAULT 'pending',
    initiated_by   TEXT NOT NULL,
    initiated_at    INTEGER NOT NULL,
    completed_at    INTEGER,
    cancelled_at    INTEGER,
    cancelled_by    TEXT,
    FOREIGN KEY(version_id) REFERENCES update_versions(id),
    FOREIGN KEY(initiated_by) REFERENCES operators(id)
);

CREATE INDEX idx_update_pushes_status ON update_pushes(status);
CREATE INDEX idx_update_pushes_initiated_at ON update_pushes(initiated_at DESC);
```

**Install Types:** `immediate`, `scheduled`
**Status:** `pending`, `in_progress`, `completed`, `failed`, `cancelled`

### 4.8 update_push_devices (Migration 23)

Per-device update status within a push operation.

```sql
CREATE TABLE update_push_devices (
    id              TEXT PRIMARY KEY,
    push_id         TEXT NOT NULL,
    device_id       TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    sent_at         INTEGER,
    acknowledged_at INTEGER,
    error           TEXT,
    retry_count     INTEGER NOT NULL DEFAULT 0,
    created_at      INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL,
    FOREIGN KEY(push_id) REFERENCES update_pushes(id) ON DELETE CASCADE,
    FOREIGN KEY(device_id) REFERENCES devices(id)
);

CREATE UNIQUE INDEX idx_push_device ON update_push_devices(push_id, device_id);
CREATE INDEX idx_update_push_devices_status ON update_push_devices(push_id, status);
```

**Status:** `pending`, `sent`, `acknowledged`, `failed`

### 4.9 operator_settings (Migration 24)

Per-operator configuration and threshold settings.

```sql
CREATE TABLE operator_settings (
    operator_id              TEXT PRIMARY KEY,
    server_url               TEXT,
    device_id                TEXT,
    request_timeout_ms      INTEGER NOT NULL DEFAULT 8000,
    auto_reconnect           INTEGER NOT NULL DEFAULT 1,
    strict_hmac             INTEGER NOT NULL DEFAULT 0,
    log_buffer_limit         INTEGER NOT NULL DEFAULT 500,
    signal_history_limit     INTEGER NOT NULL DEFAULT 240,
    risk_warn                INTEGER NOT NULL DEFAULT 70,
    risk_crit                INTEGER NOT NULL DEFAULT 85,
    thermal_warn            INTEGER NOT NULL DEFAULT 45,
    thermal_crit            INTEGER NOT NULL DEFAULT 50,
    buffer_warn             INTEGER NOT NULL DEFAULT 30,
    buffer_crit             INTEGER NOT NULL DEFAULT 15,
    notifications_enabled    INTEGER NOT NULL DEFAULT 1,
    notify_email             TEXT,
    notify_push              INTEGER NOT NULL DEFAULT 1,
    notify_webhook           INTEGER NOT NULL DEFAULT 0,
    webhook_url             TEXT,
    webhook_secret           TEXT,
    webhook_types           TEXT,
    notify_threshold_breach INTEGER NOT NULL DEFAULT 1,
    notify_device_offline   INTEGER NOT NULL DEFAULT 1,
    notify_device_online     INTEGER NOT NULL DEFAULT 1,
    notify_update_available INTEGER NOT NULL DEFAULT 1,
    notify_command_failed    INTEGER NOT NULL DEFAULT 1,
    notify_registration_request INTEGER NOT NULL DEFAULT 1,
    created_at              INTEGER NOT NULL,
    updated_at              INTEGER NOT NULL,
    FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
);
```

### 4.10 refresh_tokens (Migration 25)

JWT refresh token storage for extended sessions.

```sql
CREATE TABLE refresh_tokens (
    id              TEXT PRIMARY KEY,
    token_hash      TEXT NOT NULL UNIQUE,
    operator_id     TEXT NOT NULL,
    session_id      TEXT NOT NULL,
    expires_at      INTEGER NOT NULL,
    created_at      INTEGER NOT NULL,
    replaced_by_id  TEXT,
    revoked         INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE,
    FOREIGN KEY(session_id) REFERENCES auth_sessions(id) ON DELETE CASCADE,
    FOREIGN KEY(replaced_by_id) REFERENCES refresh_tokens(id)
);

CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_operator_id ON refresh_tokens(operator_id);
```

### 4.11 notification_audit_log (Migration 26)

Audit trail for operator notifications.

```sql
CREATE TABLE notification_audit_log (
    id              TEXT PRIMARY KEY,
    operator_id     TEXT NOT NULL,
    event_type      TEXT NOT NULL,
    channel         TEXT NOT NULL,
    payload         TEXT,
    sent_at         INTEGER NOT NULL,
    FOREIGN KEY(operator_id) REFERENCES operators(id)
);

CREATE INDEX idx_notification_audit_operator ON notification_audit_log(operator_id, sent_at DESC);
CREATE INDEX idx_notification_audit_type ON notification_audit_log(event_type, sent_at DESC);
```

---

## 5. Schema Modifications

### 5.1 Devices Table (Migration 27 - ALTER)

Add new columns to existing devices table.

```sql
ALTER TABLE devices ADD COLUMN device_name TEXT;
ALTER TABLE devices ADD COLUMN os_version TEXT;
ALTER TABLE devices ADD COLUMN security_patch TEXT;
ALTER TABLE devices ADD COLUMN build_id TEXT;
ALTER TABLE devices ADD COLUMN deregistered_at INTEGER;
ALTER TABLE devices ADD COLUMN deletion_scheduled_at INTEGER;
ALTER TABLE devices ADD COLUMN fcm_token_refreshed_at INTEGER;
```

---

## 6. Schema Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           VYZORIX DATABASE SCHEMA                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐               │
│  │   devices    │────►│  telemetry   │     │   commands   │               │
│  │              │     │              │     │              │               │
│  │ • id (PK)   │     │ • id (PK)   │     │ • id (PK)   │               │
│  │ • imei      │     │ • device_id │     │ • device_id │               │
│  │ • online    │     │ • payload   │     │ • type      │               │
│  │ • fcm_token │     │ • risk_score│     │ • status    │               │
│  │ • +14 cols │     │ • buffer    │     │ • expires_at│               │
│  └──────┬───────┘     │ • temp     │     └──────┬───────┘               │
│         │              └──────────────┘            │                       │
│         │                                              │                       │
│         ├──────►┌──────────────┐                                          │
│         │       │    events    │ (v20)                                     │
│         │       │ • event_type │                                          │
│         │       │ • severity   │                                          │
│         │       │ • metadata   │                                          │
│         │       └──────────────┘                                          │
│         │                                                              │
│         ├──────►┌──────────────────┐                                    │
│         │       │  inbox_requests  │ (v21)                                │
│         │       │ • imei (UNIQUE) │                                    │
│         │       │ • status        │                                    │
│         │       │ • command_secret│                                    │
│         │       └──────────────────┘                                    │
│         │                                                              │
│         ├──────►┌──────────────────┐                                    │
│         │       │ registration_logs │ (v21)                                │
│         │       └──────────────────┘                                    │
│         │                                                              │
│         ├──────►┌──────────────┐                                        │
│         │       │ device_logs  │ (v22)                                    │
│         │       └──────────────┘                                        │
│         │                                                              │
│         ├──────►┌──────────────┐                                        │
│         │       │device_events │ (v22)                                     │
│         │       └──────────────┘                                        │
│         │                                                              │
│         └──────►┌──────────────┐                                        │
│                 │message_queue │ (v19)                                     │
│                 └──────────────┘                                        │
│                                                                            │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐               │
│  │  operators   │────►│auth_sessions │────►│refresh_tokens│ (v25)         │
│  │              │     │              │     │              │               │
│  │ • id (PK)   │     │ • id (PK)   │     │ • token_hash│               │
│  │ • email     │     │ • operator_id│     └──────────────┘               │
│  │ • password  │     └──────────────┘                                    │
│  │ • mfa_enab  │                                                          │
│  │ • thresholds│     ┌──────────────┐     ┌──────────────┐               │
│  └──────┬──────┘     │ operator_    │     │notification_ │               │
│         │     │     │ settings     │     │audit_log    │ (v26)         │
│         │     │     │ (v24)       │     │              │               │
│         │     └────►└──────────────┘     └──────────────┘               │
│         │                                                                  │
│         ├──────►┌──────────────┐     ┌──────────────┐               │
│         │       │email_verifs  │     │pwd_reset_   │               │
│         │       └──────────────┘     │tokens       │               │
│         │                              └──────────────┘               │
│         ├──────►┌──────────────┐                                    │
│         │       │audit_logs   │                                    │
│         │       └──────────────┘                                    │
│         │                                                                  │
│  ┌─────▼──────┐     ┌──────────────┐                                    │
│  │api_clients │────►│ signing_keys │                                    │
│  │            │     └──────────────┘                                    │
│  └────────────┘                                                         │
│                                                                            │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐               │
│  │update_      │────►│update_pushes │────►│update_push_ │               │
│  │versions (v23)     │              │     │devices     │ (v23)         │
│  │ • version   │     │ • status    │     │ • status    │               │
│  │ • is_latest │     │ • scheduled  │     └──────────────┘               │
│  └──────────────┘     └──────────────┘                                    │
│                                                                            │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 7. Entity Relationships

```
operators (1) ────── (N) auth_sessions
     │                    │
     ├───── (1) operator_settings (1:1)
     ├───── (N) refresh_tokens
     ├───── (N) email_verifications
     ├───── (N) password_reset_tokens
     ├───── (N) audit_logs
     ├───── (N) notification_audit_log
     ├───── (N) inbox_requests (as operator)
     └───── (N) registration_logs (as operator)

api_clients (1) ─── (N) signing_keys

devices (1) ────── (N) telemetry
     │
     ├───── (N) commands
     ├───── (N) events
     ├───── (N) message_queue
     ├───── (N) device_logs
     ├───── (N) device_events
     ├───── (N) registration_logs
     ├───── (N) update_push_devices
     │
     └──── (0/1) inbox_requests (inbox entry for this device)

inbox_requests ──becomes──> devices (after registration)

update_versions (1) ─── (N) update_pushes
                              │
                              └───── (N) update_push_devices
```

---

## 8. Full Index List

| Table | Index | Columns | Type |
|-------|-------|---------|------|
| telemetry | idx_telemetry_device_time | device_id, received_at DESC | Regular |
| events | idx_events_device_time | device_id, created_at DESC | Regular |
| events | idx_events_type | event_type, created_at DESC | Regular |
| message_queue | idx_message_queue_device_expires | device_id, expires_at | Regular |
| inbox_requests | idx_inbox_pending | status, created_at DESC | Regular |
| registration_logs | idx_registration_logs_device | device_id, created_at DESC | Regular |
| device_logs | idx_device_logs | device_id, timestamp DESC | Regular |
| device_logs | idx_device_logs_cursor | device_id, timestamp DESC, id | Regular |
| device_logs | idx_device_logs_event_type | event_type | Regular |
| device_events | idx_device_events_device_timestamp | device_id, timestamp DESC | Regular |
| device_events | idx_device_events_cursor | device_id, timestamp DESC, id | Regular |
| device_events | idx_device_events_type | event_type, timestamp DESC | Regular |
| update_versions | idx_versions_date | release_date DESC | Regular |
| update_pushes | idx_update_pushes_status | status | Regular |
| update_pushes | idx_update_pushes_initiated_at | initiated_at DESC | Regular |
| update_push_devices | idx_push_device | push_id, device_id | UNIQUE |
| update_push_devices | idx_update_push_devices_status | push_id, status | Regular |
| operator_settings | operator_id (PRIMARY KEY) | operator_id | Primary |
| notification_audit_log | idx_notification_audit_operator | operator_id, sent_at DESC | Regular |
| notification_audit_log | idx_notification_audit_type | event_type, sent_at DESC | Regular |
| refresh_tokens | idx_refresh_tokens_token_hash | token_hash | Regular |
| refresh_tokens | idx_refresh_tokens_operator_id | operator_id | Regular |

---

## 9. Migration Code

### 9.1 Migration 20 - events

```go
func migrateCreateEvents(db *sql.DB) error {
    _, err := db.ExecContext(context.Background(), `
        CREATE TABLE IF NOT EXISTS events (
            id              TEXT PRIMARY KEY,
            device_id       TEXT NOT NULL,
            event_type      TEXT NOT NULL,
            severity        TEXT NOT NULL DEFAULT 'info',
            message         TEXT,
            metadata        TEXT,
            created_at      INTEGER NOT NULL,
            FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
        )
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE INDEX IF NOT EXISTS idx_events_device_time 
        ON events(device_id, created_at DESC)
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE INDEX IF NOT EXISTS idx_events_type 
        ON events(event_type, created_at DESC)
    `)

    return err
}
```

### 9.2 Migration 21 - inbox_requests, registration_logs

```go
func migrateCreateInboxAndRegistrationLogs(db *sql.DB) error {
    _, err := db.ExecContext(context.Background(), `
        CREATE TABLE IF NOT EXISTS inbox_requests (
            id                      TEXT PRIMARY KEY,
            imei                    TEXT NOT NULL UNIQUE,
            model                   TEXT,
            manufacturer            TEXT,
            os_version              TEXT,
            app_version             TEXT,
            fcm_token               TEXT,
            firebase_install_id      TEXT,
            status                  TEXT NOT NULL DEFAULT 'pending',
            created_at              INTEGER NOT NULL,
            approved_at             INTEGER,
            rejected_at             INTEGER,
            command_secret           TEXT,
            notes                   TEXT,
            operator_id              TEXT,
            FOREIGN KEY(operator_id) REFERENCES operators(id)
        )
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE INDEX IF NOT EXISTS idx_inbox_pending 
        ON inbox_requests(status, created_at DESC)
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE TABLE IF NOT EXISTS registration_logs (
            id              TEXT PRIMARY KEY,
            device_id       TEXT NOT NULL,
            action          TEXT NOT NULL,
            operator_id     TEXT,
            details         TEXT,
            created_at      INTEGER NOT NULL,
            FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE,
            FOREIGN KEY(operator_id) REFERENCES operators(id)
        )
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE INDEX IF NOT EXISTS idx_registration_logs_device 
        ON registration_logs(device_id, created_at DESC)
    `)

    return err
}
```

### 9.3 Migration 22 - device_logs, device_events

```go
func migrateCreateDeviceLogsAndEvents(db *sql.DB) error {
    _, err := db.ExecContext(context.Background(), `
        CREATE TABLE IF NOT EXISTS device_logs (
            id              TEXT PRIMARY KEY,
            device_id       TEXT NOT NULL,
            event_type      TEXT NOT NULL,
            timestamp       INTEGER NOT NULL,
            data            TEXT,
            created_at      INTEGER NOT NULL,
            FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
        )
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE INDEX IF NOT EXISTS idx_device_logs 
        ON device_logs(device_id, timestamp DESC)
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE INDEX IF NOT EXISTS idx_device_logs_cursor 
        ON device_logs(device_id, timestamp DESC, id)
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE INDEX IF NOT EXISTS idx_device_logs_event_type 
        ON device_logs(event_type)
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE TABLE IF NOT EXISTS device_events (
            id              TEXT PRIMARY KEY,
            device_id       TEXT NOT NULL,
            event_type      TEXT NOT NULL,
            timestamp       INTEGER NOT NULL,
            data            TEXT,
            created_at      INTEGER NOT NULL,
            FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
        )
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE INDEX IF NOT EXISTS idx_device_events_device_timestamp 
        ON device_events(device_id, timestamp DESC)
    `)

    return err
}
```

### 9.4 Migration 23 - update_versions, update_pushes, update_push_devices

```go
func migrateCreateUpdates(db *sql.DB) error {
    _, err := db.ExecContext(context.Background(), `
        CREATE TABLE IF NOT EXISTS update_versions (
            id              TEXT PRIMARY KEY,
            version         TEXT NOT NULL UNIQUE,
            apk_filename    TEXT NOT NULL,
            apk_size         INTEGER NOT NULL,
            sha256           TEXT NOT NULL,
            release_date    INTEGER NOT NULL,
            release_notes    TEXT,
            release_type    TEXT NOT NULL DEFAULT 'minor',
            is_latest       INTEGER NOT NULL DEFAULT 0,
            created_at      INTEGER NOT NULL,
            updated_at      INTEGER NOT NULL
        )
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE INDEX IF NOT EXISTS idx_versions_date 
        ON update_versions(release_date DESC)
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE TABLE IF NOT EXISTS update_pushes (
            id              TEXT PRIMARY KEY,
            version_id      TEXT NOT NULL,
            install_type    TEXT NOT NULL DEFAULT 'immediate',
            scheduled_at    INTEGER,
            status          TEXT NOT NULL DEFAULT 'pending',
            initiated_by   TEXT NOT NULL,
            initiated_at    INTEGER NOT NULL,
            completed_at    INTEGER,
            cancelled_at    INTEGER,
            cancelled_by    TEXT,
            FOREIGN KEY(version_id) REFERENCES update_versions(id),
            FOREIGN KEY(initiated_by) REFERENCES operators(id)
        )
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE INDEX IF NOT EXISTS idx_update_pushes_status 
        ON update_pushes(status)
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE INDEX IF NOT EXISTS idx_update_pushes_initiated_at 
        ON update_pushes(initiated_at DESC)
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE TABLE IF NOT EXISTS update_push_devices (
            id              TEXT PRIMARY KEY,
            push_id         TEXT NOT NULL,
            device_id       TEXT NOT NULL,
            status          TEXT NOT NULL DEFAULT 'pending',
            sent_at         INTEGER,
            acknowledged_at INTEGER,
            error           TEXT,
            retry_count     INTEGER NOT NULL DEFAULT 0,
            created_at      INTEGER NOT NULL,
            updated_at      INTEGER NOT NULL,
            FOREIGN KEY(push_id) REFERENCES update_pushes(id) ON DELETE CASCADE,
            FOREIGN KEY(device_id) REFERENCES devices(id)
        )
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE UNIQUE INDEX IF NOT EXISTS idx_push_device 
        ON update_push_devices(push_id, device_id)
    `)

    return err
}
```

### 9.5 Migration 24 - operator_settings

```go
func migrateCreateOperatorSettings(db *sql.DB) error {
    _, err := db.ExecContext(context.Background(), `
        CREATE TABLE IF NOT EXISTS operator_settings (
            operator_id              TEXT PRIMARY KEY,
            server_url               TEXT,
            device_id                TEXT,
            request_timeout_ms       INTEGER NOT NULL DEFAULT 8000,
            auto_reconnect           INTEGER NOT NULL DEFAULT 1,
            strict_hmac             INTEGER NOT NULL DEFAULT 0,
            log_buffer_limit         INTEGER NOT NULL DEFAULT 500,
            signal_history_limit     INTEGER NOT NULL DEFAULT 240,
            risk_warn                INTEGER NOT NULL DEFAULT 70,
            risk_crit                INTEGER NOT NULL DEFAULT 85,
            thermal_warn            INTEGER NOT NULL DEFAULT 45,
            thermal_crit            INTEGER NOT NULL DEFAULT 50,
            buffer_warn             INTEGER NOT NULL DEFAULT 30,
            buffer_crit             INTEGER NOT NULL DEFAULT 15,
            notifications_enabled    INTEGER NOT NULL DEFAULT 1,
            notify_email             TEXT,
            notify_push              INTEGER NOT NULL DEFAULT 1,
            notify_webhook           INTEGER NOT NULL DEFAULT 0,
            webhook_url             TEXT,
            webhook_secret           TEXT,
            webhook_types           TEXT,
            notify_threshold_breach INTEGER NOT NULL DEFAULT 1,
            notify_device_offline   INTEGER NOT NULL DEFAULT 1,
            notify_device_online     INTEGER NOT NULL DEFAULT 1,
            notify_update_available INTEGER NOT NULL DEFAULT 1,
            notify_command_failed    INTEGER NOT NULL DEFAULT 1,
            notify_registration_request INTEGER NOT NULL DEFAULT 1,
            created_at              INTEGER NOT NULL,
            updated_at              INTEGER NOT NULL,
            FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE
        )
    `)

    return err
}
```

### 9.6 Migration 25 - refresh_tokens

```go
func migrateCreateRefreshTokens(db *sql.DB) error {
    _, err := db.ExecContext(context.Background(), `
        CREATE TABLE IF NOT EXISTS refresh_tokens (
            id              TEXT PRIMARY KEY,
            token_hash      TEXT NOT NULL UNIQUE,
            operator_id     TEXT NOT NULL,
            session_id      TEXT NOT NULL,
            expires_at      INTEGER NOT NULL,
            created_at      INTEGER NOT NULL,
            replaced_by_id  TEXT,
            revoked         INTEGER NOT NULL DEFAULT 0,
            FOREIGN KEY(operator_id) REFERENCES operators(id) ON DELETE CASCADE,
            FOREIGN KEY(session_id) REFERENCES auth_sessions(id) ON DELETE CASCADE,
            FOREIGN KEY(replaced_by_id) REFERENCES refresh_tokens(id)
        )
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash 
        ON refresh_tokens(token_hash)
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE INDEX IF NOT EXISTS idx_refresh_tokens_operator_id 
        ON refresh_tokens(operator_id)
    `)

    return err
}
```

### 9.7 Migration 26 - notification_audit_log

```go
func migrateCreateNotificationAuditLog(db *sql.DB) error {
    _, err := db.ExecContext(context.Background(), `
        CREATE TABLE IF NOT EXISTS notification_audit_log (
            id              TEXT PRIMARY KEY,
            operator_id     TEXT NOT NULL,
            event_type      TEXT NOT NULL,
            channel         TEXT NOT NULL,
            payload         TEXT,
            sent_at         INTEGER NOT NULL,
            FOREIGN KEY(operator_id) REFERENCES operators(id)
        )
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE INDEX IF NOT EXISTS idx_notification_audit_operator 
        ON notification_audit_log(operator_id, sent_at DESC)
    `)
    if err != nil {
        return err
    }

    _, err = db.ExecContext(context.Background(), `
        CREATE INDEX IF NOT EXISTS idx_notification_audit_type 
        ON notification_audit_log(event_type, sent_at DESC)
    `)

    return err
}
```

### 9.8 Migration 27 - devices columns

```go
func migrateAddDevicesColumns(db *sql.DB) error {
    cols := []string{
        `ALTER TABLE devices ADD COLUMN device_name TEXT`,
        `ALTER TABLE devices ADD COLUMN os_version TEXT`,
        `ALTER TABLE devices ADD COLUMN security_patch TEXT`,
        `ALTER TABLE devices ADD COLUMN build_id TEXT`,
        `ALTER TABLE devices ADD COLUMN deregistered_at INTEGER`,
        `ALTER TABLE devices ADD COLUMN deletion_scheduled_at INTEGER`,
        `ALTER TABLE devices ADD COLUMN fcm_token_refreshed_at INTEGER`,
    }
    
    for _, col := range cols {
        db.ExecContext(context.Background(), col) //nolint:errcheck
    }

    return nil
}
```

---

*Document Version: 2.0*
*Status: Ready for Implementation*
*Database: SQLite (WAL mode)*
*Total Tables: 27 (16 existing + 11 new)*
*Total Migrations: 27*
