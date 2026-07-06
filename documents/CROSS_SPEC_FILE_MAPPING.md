# Cross-Spec File Mapping

> **Version:** 1.0  
> **Created:** 2026-07-06  
> **Purpose:** Map which files are shared vs unique across frontend specs

---

## Overview

This document maps every file mentioned across all frontend specs, showing:
- **SHARED**: Files used by multiple specs (in `_shared/` directories)
- **FEATURE**: Files unique to one spec
- **DUPLICATE**: Files incorrectly named the same in multiple specs (must be renamed)

---

## Shared Files (Cross-Spec)

These files exist once and are shared:

### Domain Layer - `_shared/`

| File | Used By | Purpose |
|------|---------|---------|
| `domain/_shared/domain-pagination.ts` | ALL | Pagination types, `PaginatedResult<T>` |
| `domain/_shared/domain-errors.ts` | ALL | `DomainError`, `ValidationError` base classes |

### Data Layer - `_shared/`

| File | Used By | Purpose |
|------|---------|---------|
| `lib/api/_shared/graphql-client.ts` | ALL | Apollo/GraphQL client setup |
| `lib/api/_shared/rest-client.ts` | ALL | REST client base |

### Hooks - `_shared/`

| File | Used By | Purpose |
|------|---------|---------|
| `hooks/_shared/use-pagination.ts` | ALL | Generic pagination hook |
| `hooks/_shared/use-search.ts` | ALL | Generic search/filter hook |
| `hooks/_shared/use-device-selector.ts` | ALL | Device selection state |
| `hooks/_shared/use-time-range.ts` | ALL | Time range selection |
| `hooks/_shared/use-debounce.ts` | Settings | Debounced save hook |
| `hooks/_shared/use-refresh.ts` | Diagnostics | Refresh trigger |

### UI Components - `shared/`

| File | Used By | Purpose |
|------|---------|---------|
| `components/shared/section.tsx` | ALL | Bordered section wrapper |
| `components/shared/section-header.tsx` | ALL | Section with title/subtitle |
| `components/shared/empty-state.tsx` | ALL | Empty state display |
| `components/shared/loading-skeleton.tsx` | ALL | Loading placeholder |
| `components/shared/data-table.tsx` | Multiple | Table with sort/pagination |
| `components/shared/pagination.tsx` | Multiple | Pagination controls |
| `components/shared/search-input.tsx` | Multiple | Search with clear |
| `components/shared/filter-select.tsx` | Multiple | Dropdown filter |
| `components/shared/status-badge.tsx` | Multiple | Status indicator |
| `components/shared/tab-nav.tsx` | Multiple | Tab navigation |
| `components/shared/refresh-button.tsx` | Diagnostics | Refresh with loading |

---

## Feature Files (Unique per Spec)

### API Keys (`domain/apikey/`, `lib/api/apikey/`, `hooks/apikey/`)

| Layer | File | Purpose |
|-------|------|---------|
| Domain | `apikey-entity.ts` | `ApiKey`, `ApiKeyScope`, `CreateRequest` |
| Domain | `apikey-mappers.ts` | `rawToAPIKey()`, `apiKeyToResponse()` |
| Domain | `apikey-validators.ts` | `validateKeyName()`, `validateScope()` |
| Domain | `apikey-constants.ts` | `MAX_KEYS_PER_MONTH`, `MAX_NAME_LENGTH` |
| Data | `graphql-apikey-queries.ts` | `GET_API_KEYS`, etc. |
| Data | `graphql-apikey-mutations.ts` | `CREATE_API_KEY`, etc. |
| Data | `graphql-apikey-types.ts` | Raw GraphQL response types |
| Data | `rest-apikey-endpoints.ts` | REST fallback endpoints |
| Hooks | `use-apikeys.ts` | List keys hook |
| Hooks | `use-create-apikey.ts` | Create key hook |
| Hooks | `use-revoke-apikey.ts` | Revoke key hook |
| Hooks | `use-rotate-apikey.ts` | Rotate key hook |
| Hooks | `use-update-apikey.ts` | Update key hook |
| Hooks | `use-apikey-stats.ts` | Usage statistics |

### Settings (`domain/settings/`, `lib/api/settings/`, `hooks/settings/`)

| Layer | File | Purpose |
|-------|------|---------|
| Domain | `settings-entity.ts` | `VyzorixConfig`, `ThresholdSettings` |
| Domain | `settings-mappers.ts` | `settingsFromRaw()` |
| Domain | `settings-validators.ts` | `validateSettings()` |
| Data | `graphql-settings-queries.ts` | `GET_SETTINGS`, etc. |
| Data | `graphql-settings-mutations.ts` | `UPDATE_SETTINGS`, etc. |
| Data | `graphql-settings-fragments.ts` | Reusable fragments |
| Data | `graphql-settings-types.ts` | Raw GraphQL response types |
| Data | `rest-settings-endpoints.ts` | REST fallback |
| Hooks | `use-settings.ts` | Get/update settings |
| Hooks | `use-thresholds.ts` | Get/update thresholds |
| Hooks | `use-notifications.ts` | Get/update notifications |

### Updates (`domain/updates/`, `lib/api/updates/`, `hooks/updates/`)

| Layer | File | Purpose |
|-------|------|---------|
| Domain | `updates-entity.ts` | `Version`, `Changelog`, `UpdatePush` |
| Domain | `updates-mappers.ts` | `versionFromAPI()` |
| Domain | `updates-validators.ts` | `validateVersion()` |
| Data | `graphql-updates-queries.ts` | `GET_VERSIONS`, `GET_CHANGELOG` |
| Data | `graphql-updates-mutations.ts` | `PUSH_UPDATE`, `CANCEL_UPDATE` |
| Data | `graphql-updates-fragments.ts` | Version fragments |
| Data | `graphql-updates-types.ts` | Raw GraphQL response types |
| Hooks | `use-versions.ts` | List versions |
| Hooks | `use-changelog.ts` | Get changelog |
| Hooks | `use-push-update.ts` | Push update |
| Hooks | `use-update-history.ts` | Update history |
| Hooks | `use-sync-status.ts` | GitHub sync status |

### Diagnostics (`domain/diagnostics/`, `lib/api/diagnostics/`, `hooks/diagnostics/`)

| Layer | File | Purpose |
|-------|------|---------|
| Domain | `diagnostics-entity.ts` | `DeviceInspection`, `TimelineEvent` |
| Domain | `diagnostics-mappers.ts` | `inspectionFromRaw()`, `eventFromRaw()` |
| Domain | `diagnostics-validators.ts` | `validateInspection()` |
| Data | `graphql-diagnostics-queries.ts` | `GET_DEVICE_INSPECTION` |
| Data | `graphql-diagnostics-fragments.ts` | Inspection fragments |
| Data | `graphql-diagnostics-types.ts` | Raw response types |
| Data | `rest-diagnostics-endpoints.ts` | REST fallback |
| Hooks | `use-device-inspection.ts` | Fetch inspection |
| Hooks | `use-device-timeline.ts` | Fetch timeline |
| Hooks | `use-timeline-filter.ts` | Filter timeline |

### Commands (`domain/commands/`, `lib/api/commands/`, `hooks/commands/`)

| Layer | File | Purpose |
|-------|------|---------|
| Domain | `command-entity.ts` | `Command`, `CommandStatus`, `PresetCommand` |
| Domain | `command-mappers.ts` | `commandFromRaw()`, `commandToApi()` |
| Domain | `command-validators.ts` | `validateCommand()` |
| Domain | `command-constants.ts` | `PRESET_COMMANDS` array |
| Data | `graphql-commands-queries.ts` | `GET_COMMANDS`, `GET_PENDING_COMMANDS` |
| Data | `graphql-commands-mutations.ts` | `SEND_COMMAND`, `CANCEL_COMMAND` |
| Data | `graphql-commands-fragments.ts` | Command fragments |
| Data | `graphql-commands-types.ts` | Raw response types |
| Data | `rest-commands-endpoints.ts` | REST fallback |
| Hooks | `use-commands.ts` | Send commands, get presets |
| Hooks | `use-command-history.ts` | Command history |
| Hooks | `use-pending-commands.ts` | Pending queue |

### Logs (`domain/logs/`, `lib/api/logs/`, `hooks/logs/`)

| Layer | File | Purpose |
|-------|------|---------|
| Domain | `log-entity.ts` | `LogEntry`, `LogLevel`, `LogSource` |
| Domain | `log-mappers.ts` | `logFromRaw()` |
| Domain | `log-filters.ts` | `filterByType()`, `validateLogEntry()` |
| Data | `graphql-logs-queries.ts` | `GET_LOG_ENTRIES`, `GET_LOG_STATS` |
| Data | `graphql-logs-subscriptions.ts` | Real-time log subscription |
| Data | `graphql-logs-fragments.ts` | Log entry fragments |
| Data | `graphql-logs-types.ts` | Raw response types |
| Data | `rest-logs-endpoints.ts` | REST fallback |
| Hooks | `use-logs.ts` | Get logs |
| Hooks | `use-log-stream.ts` | Real-time streaming |
| Hooks | `use-log-filter.ts` | Filter logs |

### Device Registration (`domain/registration/`, `lib/api/registration/`, `hooks/registration/`)

| Layer | File | Purpose |
|-------|------|---------|
| Domain | `registration-entity.ts` | `InboxEntry`, `InboxStatus` |
| Domain | `registration-mappers.ts` | `inboxFromRaw()` |
| Domain | `registration-validators.ts` | `validateIMEI()` |
| Data | `graphql-registration-queries.ts` | `GET_INBOX`, `GET_INBOX_ENTRY` |
| Data | `graphql-registration-mutations.ts` | `ACK_INBOX`, `REGISTER_DEVICE` |
| Data | `graphql-registration-fragments.ts` | Inbox fragments |
| Data | `graphql-registration-types.ts` | Raw response types |
| Data | `rest-registration-endpoints.ts` | REST fallback |
| Hooks | `use-inbox.ts` | Get inbox |
| Hooks | `use-inbox-entry.ts` | Single entry |
| Hooks | `use-ack-inbox.ts` | Acknowledge |
| Hooks | `use-register-device.ts` | Register device |
| Hooks | `use-deregister-device.ts` | Deregister |

### Real-time WebSocket (`domain/realtime/`, `lib/api/websocket/`, `hooks/realtime/`)

| Layer | File | Purpose |
|-------|------|---------|
| Domain | `realtime-entity.ts` | `WSTelemetry`, `WSEvent`, `WSCommand` |
| Domain | `realtime-mappers.ts` | `telemetryFromRaw()` |
| Domain | `realtime-validators.ts` | `validateTelemetry()` |
| Data | `websocket-client.ts` | WebSocket wrapper |
| Data | `websocket-connection.ts` | Connection state machine |
| Data | `websocket-heartbeat.ts` | Heartbeat manager |
| Data | `websocket-reconnect.ts` | Reconnection logic |
| Data | `websocket-messages.ts` | Message parsers |
| Data | `graphql-realtime-subscriptions.ts` | GraphQL subscriptions |
| Data | `graphql-realtime-mutations.ts` | Command mutations |
| Hooks | `use-websocket-connection.ts` | Connection hook |
| Hooks | `use-device-telemetry.ts` | Telemetry stream |
| Hooks | `use-dashboard-events.ts` | Dashboard events |
| Hooks | `use-command-dispatch.ts` | Dispatch commands |

### Devices (Shared - used by multiple features)

| Layer | File | Purpose |
|-------|------|---------|
| Domain | `device-entity.ts` | `Device`, `DeviceStatus`, `ConnectionState` |
| Domain | `device-mappers.ts` | `deviceFromRaw()` |
| Domain | `device-validators.ts` | `validateDeviceId()` |
| Domain | `device-constants.ts` | Device status constants |
| Data | `graphql-devices-queries.ts` | `GET_DEVICES`, `GET_DEVICE` |
| Data | `graphql-devices-mutations.ts` | Device mutations |
| Data | `graphql-devices-fragments.ts` | Device fragments |
| Data | `graphql-devices-types.ts` | Raw response types |
| Hooks | `use-devices.ts` | Device list |
| Hooks | `use-device.ts` | Single device |
| Hooks | `use-device-stream.ts` | Real-time updates |
| Hooks | `use-device-telemetry.ts` | Telemetry data |
| Hooks | `use-device-inspection.ts` | Device inspection |
| Hooks | `use-device-timeline.ts` | Device timeline |

---

## Resolved Filename Conflicts

| Original (Conflicting) | Resolved | Notes |
|------------------------|----------|-------|
| `types.ts` in every feature | `{feature}-entity.ts` | API Keys: `apikey-entity.ts` |
| `transforms.ts` in every feature | `{feature}-mappers.ts` | API Keys: `apikey-mappers.ts` |
| `validation.ts` in every feature | `{feature}-validators.ts` | API Keys: `apikey-validators.ts` |
| `pagination.ts` | `domain-pagination.ts` | Moved to `_shared/` |
| `error.ts` | `domain-errors.ts` | Moved to `_shared/` |
| `api-response-types.ts` | `graphql-{feature}-types.ts` | Per feature |

---

## Spec to File Mapping

| Spec Document | Features | Domain Dir | Data Dir |
|--------------|----------|------------|----------|
| `FRONTEND_ARCHITECTURE.md` | ALL | `domain/{feature}/` | `lib/api/{feature}/` |
| `FRONTEND_API_KEYS_REQUIREMENTS.md` | API Keys | `domain/apikey/` | `lib/api/apikey/` |
| `SETTINGS_PAGE.md` | Settings | `domain/settings/` | `lib/api/settings/` |
| `UPDATES_PAGE.md` | Updates | `domain/updates/` | `lib/api/updates/` |
| `DIAGNOSTICS_PAGE.md` | Diagnostics | `domain/diagnostics/` | `lib/api/diagnostics/` |
| `DASHBOARD_COMMANDS_LOGS.md` | Commands, Logs | `domain/commands/`, `domain/logs/` | `lib/api/commands/`, `lib/api/logs/` |
| `DEVICE_REGISTRATION_SYSTEM.md` | Registration | `domain/registration/` | `lib/api/registration/` |
| `REALTIME_WEBSOCKET_ARCHITECTURE.md` | Real-time | `domain/realtime/` | `lib/api/websocket/` |

---

## Implementation Priority

1. **Shared files first** - Create `_shared/` directories in domain, data, hooks, components
2. **Device domain** - Used by almost all features
3. **Per-feature** - Implement one feature at a time following the mapping
