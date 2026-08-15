# Dashboard, Commands & Logs - Enterprise Requirements Specification

> **Version:** 2.0
> **Status:** Draft
> **Created:** 2026-06-21
> **Updated:** 2026-08-15
> **Target:** Production MVP
> **Architecture:** Layered (Following `FRONTEND_ARCHITECTURE.md`)

---

>  **Architecture Alignment Note (v2.0 — Organization Model Realignment)**
>
> This document has been realigned to the **organization-scoped model** introduced by the org
> context middleware (`apps/api/internal/api/middleware/org_context.go`) and the API Client's
> `getOrganizationContext()` singleton. v1.x assumed a single-tenant "current device from config"
> model; v2.0 makes **organization context explicit and mandatory** for every read and write path.
>
> **Authoritative sources for this revision:** the **server backend** (`SERVER_BACKEND_DASHBOARD_COMMANDS_API.md`
> + actual handlers) and the **API Client package** (`packages/API_Client/`). The web application
> layer (`apps/VyzoriX_web/`) is **scaffold-only at this time** — UI pages/routes are empty and are
> treated as TODO. Do not infer the target design from the current web scaffolding; infer it from
> the server + API Client contracts below.
>
> **Key v1.x → v2.0 deltas:**
> - Every request carries an `organizationId` (resolved server-side from `X-Organization-ID`
>   header / query param / session, and client-side from `authContext.organizationId`).
> - REST is the **primary data layer**; GraphQL is a partial fallback (commands/logs queries only;
>   command mutations and metrics/telemetry/dashboard have **no** GraphQL module yet).
> - Org scoping is **device-anchored**: only `devices` has an `organization_id` column; `commands`
>   and `device_logs` inherit org membership transitively through their owning device (verified
>   via `FindByIDAndOrganization`), not via a SQL-level `WHERE organization_id = ?` on those rows.
>   See the **server scoping caveat** in §14.4.
> - File paths corrected from the v1.x `apps/web/src/` + `lib/api/` layout to the actual
>   monorepo layout: domain + data in `packages/API_Client/src/`, presentation + UI in
>   `apps/VyzoriX_web/src/`.
> - A new **State Stores** section (§23) maps the Zustand stores required to support the
>   dashboard/commands/logs UX (realtime streams, command dispatch tracking, log buffers, etc.).
>
> This document follows the **4-layer architecture** defined in `FRONTEND_ARCHITECTURE.md`:
> - **UI Layer** (`apps/VyzoriX_web/src/ui/`) - Pure UI rendering, imports only from hooks
> - **Presentation Layer** (`apps/VyzoriX_web/src/hooks/`) - UI logic, state management, imports from domain & data
> - **Domain Layer** (`packages/API_Client/src/domain/`) - Types, transforms, validation (NO external imports)
> - **Data Layer** (`packages/API_Client/src/vyzorServer/`) - API clients (REST primary / GraphQL fallback), imports only domain types
>
> **Dependency Rule:** UI → Hooks → Domain → API (flow inward only)

---

## Table of Contents

1. [Overview](#1-overview)
2. [Architecture](#2-architecture)
3. [Target File Structure](#3-target-file-structure)
4. [Page Structure](#4-page-structure)
5. [Shared UI Components](#5-shared-ui-components)
6. [Export System](#6-export-system)
7. [Dashboard: Overview Tab](#7-dashboard-overview-tab)
8. [Dashboard: Metrics Tab](#8-dashboard-metrics-tab)
9. [Dashboard: Commands Tab](#9-dashboard-commands-tab)
10. [Dashboard: Logs Tab](#10-dashboard-logs-tab)
11. [Commands Page](#11-commands-page)
12. [Logs Page](#12-logs-page)
13. [Device Page Updates](#13-device-page-updates)
14. [REST API Specification](#14-rest-api-specification)
15. [GraphQL Schema](#15-graphql-schema)
16. [Domain Types](#16-domain-types)
17. [Data Layer](#17-data-layer)
18. [Presentation Layer - Hooks](#18-presentation-layer---hooks)
19. [UI Layer - Components](#19-ui-layer---components)
20. [File Changes Summary](#20-file-changes-summary)
21. [Implementation Order](#21-implementation-order)
22. [Preset Commands Reference](#22-preset-commands-reference)
23. [State Stores](#23-state-stores)

---

## 1. Overview

### 1.1 Purpose

Redesign the Dashboard page with tabs for better organization, create a shared Commands page accessible from both Dashboard and Device, and maintain the Logs page.

### 1.2 Key Principles

- **Organization context first** - Every read/write resolves the active `organizationId`
  (web: `useCurrentOrganizationId()` → `authContext.organizationId`; server: `X-Organization-ID`
  header resolved by `org_context.go` middleware). No command, log, metric, telemetry, or
  dashboard request is ever made without an org context.
- **Device context is org-scoped** - The "current device" is selected within an organization; a
  device is only addressable if it belongs to the active org (server enforces via
  `FindByIDAndOrganization`).
- **Commands are shared** - Accessible from Dashboard tabs AND Device page, always within the
  active org.
- **Logs are standalone** - Separate `/logs` route, accessible from Dashboard tabs, org-scoped.
- **No duplication** - Same component used in multiple places via routing.
- **REST primary, GraphQL fallback** - REST is the primary data layer; GraphQL exists as a partial
  fallback (commands/logs queries only — see §15). Metrics/telemetry/dashboard are REST-only.
- **Layered Architecture** - Follow FRONTEND_ARCHITECTURE.md dependency rules.

### 1.3 Design Aesthetic

- Command center feel (dense, information-rich)
- Custom sections with borders (not floating cards)
- Rose-500 for accents/highlights only
- Minimal shadows
- Monospace for technical data

---

## 2. Architecture

### 2.1 Layered Architecture Overview

```
                     +---------------------------------------------+
                     |  ORGANIZATION CONTEXT (active org)            |
                     |  useCurrentOrganizationId() -> authContext     |
                     |  server: X-Organization-ID header (middleware) |
                     +---------------------------------------------+
                                     | flows down to every layer
                                     v

                        UI LAYER
                     (apps/VyzoriX_web/src/ui/)

      Pages, Components, Shared UI
      ONLY renders UI. Uses hooks for everything.
      NEVER imports from Data or Domain.


                               uses


                     PRESENTATION LAYER
                  (apps/VyzoriX_web/src/hooks/)

      Custom hooks that:
      - Handle UI logic, transform data for UI, manage state
      - Resolve org context (useCurrentOrganizationId) and forward orgId
      - Pass orgId into every data-layer call
      NEVER renders UI. NEVER imports from UI layer.


                               uses


                        DOMAIN LAYER
                   (packages/API_Client/src/domain/)

      Pure functions that:
      - Define types and interfaces
      - Transform data (no side effects)
      - Validate input
      NEVER imports from UI, Presentation, or Data.


                               uses


                         DATA LAYER
              (packages/API_Client/src/vyzorServer/)

      API clients that:
      - Make HTTP requests, parse responses
      - REST (primary) -> auto-injects X-Organization-ID via axios interceptor
      - GraphQL (partial fallback) -> commands/logs queries only
      NEVER imports from UI or Presentation.
```

### 2.2 Organization Context Flow

```
  useCurrentOrganizationId()         (web - null when no org selected)
            |
            v
  hook forwards organizationId ------> REST endpoint (organizationId?: string)
            |                              |  defaults to getOrganizationContext()
            |                              v
            |                       axios interceptor sets X-Organization-ID header
            |                              |
            v                              v
  query enabled only when orgId !== null   server org_context.go middleware resolves orgId
            |                              |  (query param -> header -> context -> session)
            |                              v
            |                       FindByIDAndOrganization(deviceId, orgId) membership check
            |                              |
            v                              v
  GraphQL fallback (commands/logs queries only, on REST rejection)
```

> **Cache isolation:** TanStack Query keys for commands/logs/telemetry MUST include
> `organizationId` so that switching orgs never serves another org's cached data. Query-key
> factories that currently omit `organizationId` (`dashboardStats`, `deviceMetrics`,
> `latestTelemetry`, `telemetryStats`) must be fixed - see Section 18 + Section 23.

### 2.3 Dependency Rule

```
UI Layer            Presentation Layer        Domain Layer          Data Layer
(ui/)               (hooks/)                   (domain/)             (vyzorServer/)
   |                    |                         |                       |
   +-- IMPORTS ONLY --> hooks -- IMPORTS ONLY --> domain types ---------> (REST/GraphQL)


      UI never imports from Data or Domain directly.
      Hooks never import from UI.
      Domain has zero external imports.
      Data layer imports only domain types.
```

---

## 3. Target File Structure

### 3.1 Complete Directory Tree

> Paths reflect the **actual monorepo layout**. Domain + Data layers live in
> `packages/API_Client/src/` (the `@vyzorix/api-client` workspace package); Presentation + UI
> layers live in `apps/VyzoriX_web/src/`. The v1.x `apps/web/src/` + `lib/api/` paths are
> deprecated.

```
packages/API_Client/src/                         # @vyzorix/api-client (shared data + domain)

  domain/                                        # DOMAIN LAYER (pure, no external imports)
    _shared/
       domain-shared.ts                          # Pagination, DeviceStatus, shared primitives
    commands/
       command-entity.ts                         # Command, CommandStatus, CommandParams, PresetCommandType
       command-mappers.ts                        # commandFromRaw() / commandToRaw()
       command-validators.ts                     # validateCommand(), validateStatus()
       command-constants.ts                      # PRESET_COMMANDS
    logs/
       log-entity.ts                             # LogEntry, LogListResult, LogLevel, LogStats
       log-mappers.ts                            # logFromRaw()
       log-filters.ts                            # validateLogEntry(), filter helpers
    metrics/
       metrics-entity.ts                         # DeviceMetrics, DashboardStats, TimeRange, MetricResolution
    telemetry/
       telemetry-entity.ts                       # TelemetryFrame, TelemetryHistoryResponse, TelemetryStats
    (registration/, device/, organization/ ...) # cross-feature shared domains

  vyzorServer/                                   # DATA LAYER (imports only domain types)
    rest/
       _shared/
          rest-client.ts                        # axios base + X-Organization-ID interceptor
       commands/
          command-endpoints.ts                   # commands REST object (getHistory, getPending, send, cancel, retry, pollCommandStatus)
       logs/
          log-endpoints.ts                       # logs REST object (list, get, stats)
       metrics/
          metrics-endpoints.ts                  # fetchDeviceMetrics, fetchDashboardStats, exportMetrics
       telemetry/
          telemetry-endpoints.ts                 # queryTelemetryHistory, getLatestTelemetry, getTelemetryStats, exportTelemetry
       (registration/, ...)                      # other feature REST groups
    graphql/
       _shared/
          graphql-client.ts                     # GraphQL client setup
       commands/
          graphql-commands-queries.ts            # GET_COMMANDS, GET_PENDING_COMMANDS (org-scoped)
          graphql-commands-mutations.ts         # SEND_COMMAND, CANCEL_COMMAND, RETRY_COMMAND  <-- TODO: add organizationId arg
          graphql-commands-fragments.ts
          graphql-commands-types.ts
       logs/
          graphql-logs-queries.ts               # GET_LOGS (org-scoped)
          graphql-logs-subscriptions.ts         # realtime log subscription (device-scoped)
          graphql-logs-fragments.ts
          graphql-logs-types.ts
       (NO metrics / telemetry / dashboard graphql modules - REST only)


apps/VyzoriX_web/src/                            # VyzoriX_web app (presentation + UI)

  hooks/                                         # PRESENTATION LAYER
     _shared/
        use-current-context.ts                   # useCurrentOrganizationId(), useRequiredOrganizationId(), useSelectedImei()
     commands/
        use-commands.ts                          # useCommandHistory, useCommand, usePendingCommands, useSendCommand,
                                                 #   usePollCommandStatus, useCancelCommand, useRetryCommand  (TODO - org-aware)
        index.ts                                # barrel (TODO)
     logs/
        use-logs.ts                             # useDeviceLogs, useLog, useLogStats  (TODO - org-aware)
        use-log-stream.ts                       # realtime log streaming (TODO)
        index.ts                                # barrel (TODO)
     metrics/
        use-metrics.ts                           # useDeviceMetrics, useDashboardStats, useExportMetrics  (TODO - org-aware)
        index.ts                                # barrel (TODO)
     telemetry/
        use-telemetry.ts                         # useTelemetryHistory, useLatestTelemetry, useTelemetryStats, useExportTelemetry  (TODO)
        index.ts                                # barrel (TODO)
     shared/
        use-pagination.ts                        # generic pagination hook
        use-export.ts                            # CSV/JSON export hook
        index.ts
     (registration/, device/, organization/ ...)

  stores/                                       # ZUSTAND STATE STORES (see Section 23)
     auth-store.ts                               # auth + active organizationId (DONE)
     websocket-store.ts                          # WS connection/subscription state (DONE)
     connectivity-store.ts                       # online/connectivity status (DONE)
     device-selector-store.ts                    # selected device + filters (DONE)
     theme-store.ts                              # theme mode (DONE)
     command-dispatch-store.ts                   # in-memory pending-command tracking (DONE)
     (log-stream-store.ts, metrics-realtime-store.ts, dashboard-store.ts, command-queue-store.ts - TODO, see Section 23)

  lib/
     query-keys.ts                              # TanStack query-key factories (MUST include organizationId)

  ui/                                            # UI LAYER
     components/
        shared/                                  # bordered section, empty-state, data-table, pagination, search-input, filter-select, status-badge
        commands/                                # commands-send, commands-pending, commands-history, command-row, command-status-badge  (TODO)
        logs/                                    # logs-stream, log-entry, log-filters, log-stats  (TODO)
        dashboard/                               # dashboard-page, dashboard-overview, dashboard-metrics, device-stats-grid, activity-feed  (TODO)
     pages/
        dashboard/                               # (empty scaffold - .gitkeep only - TODO)
        commands/                                # (empty scaffold - .gitkeep only - TODO)
        logs/                                    # (empty scaffold - .gitkeep only - TODO)

  routes/                                        # PAGE LAYER (TanStack Router)
     dashboard.tsx, dashboard.overview.tsx, dashboard.metrics.tsx, dashboard.commands.tsx, dashboard.logs.tsx  (TODO)
     commands-page.tsx, commands.pending.tsx, commands.history.tsx  (TODO)
     logs-page.tsx  (TODO)
     device.tsx, device.$imei.commands.tsx, ...  (TODO)
```

> **Status note:** the API Client domain + REST data layers for commands/logs/metrics/telemetry
> are DONE (org-scoped). The web hooks exist but are flagged TODO for org-aware hardening (query
> keys missing `organizationId`, GraphQL fallback wiring). The web UI pages and routes are empty
> scaffolds (`.gitkeep` only). See Section 20 for the authoritative DONE/TODO file list.

### 3.2 Sidebar Navigation

```

  VYZORIX                                                          

                                                                     
   Dashboard                                                       
     [Overview] [Metrics] [Commands] [Logs]                       
                                                                     
   Device                                                          
     [Inbox] [Overview] [Telemetry] [History] [Commands]         
                                                                     
   Updates                                                         
                                                                     
   Diagnostics                                                     
     [Inspector] [Timeline]                                       
                                                                     
   Alerts                                                          
     [Active] [Status] [History]                                  
                                                                     
   Settings                                                        
                                                                     

```

### 3.3 Route Structure

```
DASHBOARD:
/dashboard                     → Overview tab
/dashboard/metrics             → Metrics tab
/dashboard/commands            → Commands (shared)
/dashboard/commands/pending    → Pending queue
/dashboard/commands/history    → Command history
/dashboard/logs                → Logs tab

DEVICE:
/device                        → Overview tab
/device/inbox                  → Inbox tab
/device/:imei                  → Overview tab
/device/:imei/telemetry         → Telemetry tab
/device/:imei/history           → History tab
/device/:imei/commands          → Commands (shared)
/device/:imei/commands/pending → Pending queue
/device/:imei/commands/history → Command history

COMMANDS (STANDALONE):
/commands                      → Send commands
/commands/pending               → Pending queue
/commands/history               → Command history

LOGS:
/logs                          → Logs (standalone page)
```

### 3.4 Shared Components (by Layer)

| Component | Layer | Used By | Purpose |
|-----------|-------|---------|---------|
| `CommandsSend` | UI | All command pages | Send commands grid |
| `CommandsPending` | UI | All pending pages | Pending queue |
| `CommandsHistory` | UI | All history pages | Full history |
| `CommandsRecent` | UI | Send page | Recent commands |
| `LogsStream` | UI | Dashboard/Logs tab, /logs page | Event log display |
| `useCommands` | Presentation | UI components | Send/cancel commands |
| `useCommandHistory` | Presentation | UI components | Command history |
| `useLogs` | Presentation | UI components | Log queries |
| `commandFromRaw` | Domain | Presentation | Transform API data |
| `logFromRaw` | Domain | Presentation | Transform API data |

---

## 4. Page Structure (Routes)

> Routes are in `src/routes/` and follow TanStack Start conventions.

### 4.1 Dashboard Routes

| Route | File | Component | Description |
|-------|------|-----------|-------------|
| `/dashboard` | `dashboard.tsx` | `DashboardPage` | Redirects to `/dashboard/overview` |
| `/dashboard/overview` | `dashboard.overview.tsx` | `DashboardOverview` | Overview tab |
| `/dashboard/metrics` | `dashboard.metrics.tsx` | `DashboardMetrics` | Metrics tab |
| `/dashboard/commands` | `dashboard.commands.tsx` | `DashboardCommands` | Commands tab (→ /commands) |
| `/dashboard/commands/pending` | `dashboard.commands.pending.tsx` | - | → /commands/pending |
| `/dashboard/commands/history` | `dashboard.commands.history.tsx` | - | → /commands/history |
| `/dashboard/logs` | `dashboard.logs.tsx` | `DashboardLogs` | Logs tab (→ /logs) |

### 4.2 Standalone Commands Routes

| Route | File | Component | Description |
|-------|------|-----------|-------------|
| `/commands` | `commands.tsx` | `CommandsSend` | Send commands page |
| `/commands/pending` | `commands.pending.tsx` | `CommandsPending` | Pending queue |
| `/commands/history` | `commands.history.tsx` | `CommandsHistory` | Command history |

### 4.3 Logs Routes

| Route | File | Component | Description |
|-------|------|-----------|-------------|
| `/logs` | `logs.tsx` | `LogsStream` | Standalone logs page |

### 4.4 Device Commands Routes

| Route | File | Component | Description |
|-------|------|-----------|-------------|
| `/device/:imei/commands` | `device.$imei.commands.tsx` | `CommandsSend` | Device commands |
| `/device/:imei/commands/pending` | `device.$imei.commands.pending.tsx` | `CommandsPending` | Pending queue |
| `/device/:imei/commands/history` | `device.$imei.commands.history.tsx` | `CommandsHistory` | Command history |

---

## 5. Shared UI Components (Base)

### 5.1 Design Principles

- **NO CODE DUPLICATION** - All UI components defined once in `components/shared/`
- **Composable** - Small components composed into larger ones
- **Themable** - Uses existing design tokens (--primary, --foreground, etc.)
- **Accessible** - Proper ARIA labels, keyboard navigation

### 5.2 Shared UI Components List

All shared components live in `src/components/shared/`:

| Component | File | Purpose |
|-----------|------|---------|
| `Section` | `section.tsx` | Bordered section with header |
| `SectionHeader` | `section-header.tsx` | Section header with title/subtitle |
| `EmptyState` | `empty-state.tsx` | Empty state with icon/message |
| `LoadingSkeleton` | `loading-skeleton.tsx` | Skeleton loading variants |
| `DataTable` | `data-table.tsx` | Table wrapper with sorting/pagination |
| `Pagination` | `pagination.tsx` | Pagination controls |
| `SearchInput` | `search-input.tsx` | Search input with clear |
| `FilterSelect` | `filter-select.tsx` | Dropdown filter select |
| `StatusBadge` | `status-badge.tsx` | Status indicator badge |
| `ConnectionBadge` | `connection-badge.tsx` | Connection status badge |

---

## 6. Export System

### 6.1 Export Formats

| Format | Extension | Use Case |
|--------|-----------|----------|
| CSV | `.csv` | Spreadsheet analysis |
| JSON | `.json` | Programmatic processing |
| PDF | `.pdf` | Human-readable reports |

### 6.2 Export Data Types

| Data Type | Exported Fields |
|-----------|-----------------|
| Commands | ID, Device, Command, Status, Created, Delivered, Duration |
| Logs | Timestamp, Level, Source, Message, Metadata |
| Telemetry | Timestamp, Device, Risk, Thermal, Buffer, Latency |

### 6.3 Export Hook

```typescript
// hooks/shared/use-export.ts
export const useExport = () => {
  const exportToCSV = (data: ExportableData[], filename: string) => { ... };
  const exportToJSON = (data: ExportableData[], filename: string) => { ... };
  return { exportToCSV, exportToJSON };
};
```

---

## 7. Dashboard: Overview Tab

### 7.1 Purpose

At-a-glance status of the **currently selected device within the active organization**, with key
metrics and quick actions. All data is org-scoped (the device must belong to the active org).

### 7.2 Layout

```

  DASHBOARD                                     Connected | 2s ago  

  [Overview] [Metrics] [Commands] [Logs]                            

                                                                     
   CONNECTION    
     Pixel 8 Pro                         [View Device ]      
    IMEI: 861234567890123                                        
    WS: Connected · FCM: Valid · Last: 2s ago                 
     
                                                                     
   METRICS    
    RISK          THERMAL        UPTIME         BUFFER         
     72   45°C   4d      67%    
    Healthy        Normal        Running        Stable            
     
                                                                     
   QUICK ACTIONS    
    [Send Command ]  [Refresh]  [View Logs]  [Alerts: 2]    
     
                                                                     
   DEVICE INFO    
    OS: Android 14    App: v2.1.0    Build: UP1A.231005.007    
     

```

### 7.3 Sections

| Section | Content |
|---------|---------|
| **Connection** | Device name, IMEI, connection status, last seen |
| **Metrics** | Risk, Thermal, Uptime, Buffer (with progress bars) |
| **Quick Actions** | Send Command, Refresh, View Logs, Alerts count |

### 7.4 Interactions

| Element | Action |
|---------|--------|
| "View Device " | Dropdown to switch devices **within the active org** |
| Metrics | Click to navigate to Metrics tab |
| "Send Command" | Dropdown with preset commands (dispatched with org context) |
| "View Logs" | Navigate to Logs tab (org-scoped) |

---

## 8. Dashboard: Metrics Tab

### 8.1 Purpose

Deep dive into telemetry data with time range selection and export options.

### 8.2 Layout

```

  METRICS                                    [1h] [6h] [24h] [7d]   
                                                             [Export ]

   RISK SCORE    
           Current: 72  Avg: 45  Min: 32  Max: 78           
          
        100                                             
         50                                          
          0       
          
     

```

### 8.3 Controls

| Control | Type | Description |
|---------|------|-------------|
| Time Range | Buttons | 1h, 6h, 24h, 7d (default: 6h) |
| Export | Dropdown | CSV, JSON |

---

## 9. Dashboard: Commands Tab

### 9.1 Purpose

Quick access to command sending (redirects to `/commands` page).

### 9.2 Layout

```

  COMMANDS                                     [Pending] [Recent]   

   SEND COMMAND    
                
     FORCE_SPEAKER   RESET_AUDIO_HAL   TOGGLE_CAPTURE           
                
                
     REINIT_PROJECTION   DUMP_FLIGHT_DATA   UPLOAD_CRASH_ZIP           
                
     

```

---

## 10. Dashboard: Logs Tab

### 10.1 Purpose

Real-time WebSocket event stream for debugging.

### 10.2 Layout

```

  LOGS                              [All ] [Auto-scroll ] [Clear]

   EVENT STREAM    
    12:34:56.123   CONNECTED     WebSocket established           
    12:35:02.456   TELEMETRY    Risk: 72, Thermal: 45°C         
     

```

---

## 11. Commands Page

### 11.1 Routes

| Route | Purpose | Component |
|-------|---------|-----------|
| `/commands` | Send commands | `CommandsSend` |
| `/commands/pending` | Pending queue | `CommandsPending` |
| `/commands/history` | Full history | `CommandsHistory` |

### 11.2 Commands Send Page Layout

```

  COMMANDS                                        Device: Pixel 8 

  [Send] [Pending (2)] [History]                                   

   SEND COMMAND    
    Select a command to send:                                    
                
     FORCE_SPEAKER   RESET_AUDIO_HAL   TOGGLE_CAPTURE           
                
     

```

---

## 12. Logs Page

### 12.1 Purpose

Standalone logs page accessible via `/dashboard/logs` or `/logs`.

### 12.2 Layout

```

  LOGS                                              [Export] [Clear]

  Filter: [All ]                                                
   EVENT STREAM    
    [Same content as Dashboard/Logs tab]                          
     

```

---

## 13. Device Page Updates

> All tabs are org-scoped: the device must belong to the active org (server enforces via
> `FindByIDAndOrganization`). Switching org clears the selected device + scoped stores.

### 13.1 Structure

```
Device Page (tabs):
 Inbox        → Pending registration requests (see DEVICE_REGISTRATION_SYSTEM.md)
 Overview     → Device info, health, connection
 Telemetry    → Real-time charts, metrics
 History      → Historical data, export
 Commands     → CommandsPanel (shared component, org-scoped dispatch)
```

---

## 14. REST API Specification

> **Organization context (v2.0):** All endpoints below are org-scoped. The active
> `organizationId` is sent via the `X-Organization-ID` header (auto-injected by the API Client
> axios interceptor from `getOrganizationContext()`), and may also be echoed as an
> `organization_id` query param. The server resolves it in
> `apps/api/internal/api/middleware/org_context.go` (precedence: query param -> header -> auth
> context -> session `SelectedOrganizationID`). Every device-addressed endpoint verifies the device
> belongs to the active org via `deviceRepo.FindByIDAndOrganization(ctx, deviceID, orgID)` before
> returning command/log/metric/telemetry data.

### 14.1 Commands Endpoints

#### `POST /v1/device/:imei/command`

Send (dispatch) a command to a device within the active org.

**Headers:** `X-Organization-ID: <orgId>`
**Request:**
```json
{ "command": "FORCE_SPEAKER" }
```
**Response (200 OK):**
```json
{
  "dispatchId": "abc123def456",
  "delivery": "sent",
  "serverTime": 1718900000000
}
```

---

#### `GET /v1/device/:imei/commands`

Paginated command history for a device (org-scoped via device membership).

**Headers:** `X-Organization-ID: <orgId>`

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | string | all | Filter: pending, delivered, completed, failed |
| `page` | int | 1 | Page number |
| `limit` | int | 20 | Items per page (max 100) |
| `startTime` | int64 | -30d | Start timestamp (ms) |
| `endTime` | int64 | now | End timestamp (ms) |

**Response (200 OK):**
```json
{
  "commands": [
    {
      "dispatchId": "abc123def456",
      "command": "FORCE_SPEAKER",
      "status": "delivered",
      "sentAt": 1718900000000,
      "deliveredAt": 1718900000234,
      "latencyMs": 234
    }
  ],
  "pagination": { "page": 1, "limit": 20, "total": 45, "totalPages": 3, "hasMore": true }
}
```

---

#### `GET /v1/device/:imei/commands/pending`

Pending commands for a device (org-scoped).

**Headers:** `X-Organization-ID: <orgId>`

---

#### `DELETE /v1/command/:dispatchId`

Cancel a pending command. At the command root level (not nested under device); the dispatch is
verified to belong to a device in the active org.

**Headers:** `X-Organization-ID: <orgId>`
**Response:**
```json
{ "dispatchId": "abc123def456", "status": "cancelled", "cancelled": true, "serverTime": 1718900600 }
```

---

#### `POST /v1/command/:dispatchId/retry`

Retry a failed command (org-scoped via the command's owning device).

---

#### `GET /v1/command/:dispatchId/status`

Poll a command's delivery status (org-scoped).

---

### 14.2 Logs Endpoint

#### `GET /v1/device/:imei/logs`

Event logs for a device (org-scoped via device membership). Cursor-paginated.

**Headers:** `X-Organization-ID: <orgId>`

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | string | all | Filter: connection, command, telemetry, error, warning |
| `startTime` | int64 | -24h | Start timestamp (ms) |
| `endTime` | int64 | now | End timestamp (ms) |
| `limit` | int | 100 | Max results (max 500) |
| `cursor` | string | null | Pagination cursor |

**Response (200 OK):**
```json
{
  "events": [
    {
      "id": "uuid-v4",
      "type": "TELEMETRY",
      "timestamp": 1718900567000,
      "data": { "riskScore": 72, "thermalTemp": 45.2, "bufferLevel": 67 }
    }
  ],
  "pagination": { "limit": 100, "hasMore": true, "nextCursor": "base64-cursor" }
}
```

---

### 14.3 Metrics Endpoints

#### `GET /v1/device/:imei/metrics`

Aggregated chart metrics for a device (org-scoped).

**Headers:** `X-Organization-ID: <orgId>`

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `range` | string | "6h" | "1h", "6h", "24h", "7d" |
| `startTime` | int64 | (from range) | Overrides range |
| `endTime` | int64 | now | End timestamp (ms) |
| `resolution` | string | "auto" | "1m", "5m", "15m", "1h", "auto" |

**Response (200 OK):**
```json
{
  "device": { "imei": "...", "deviceName": "..." },
  "timeRange": { "start": 0, "end": 0, "range": "6h", "resolution": "5m" },
  "metrics": {
    "riskScore": { "current": 72, "avg": 45, "min": 32, "max": 78, "unit": "", "chart": [], "threshold": { "warning": 0, "critical": 0 } },
    "thermalTemp": { "current": 45.2, "avg": 42, "min": 38, "max": 52, "unit": "C", "chart": [], "threshold": { "warning": 0, "critical": 0 } },
    "bufferLevel": { "current": 67, "avg": 55, "min": 12, "max": 89, "unit": "%", "chart": [], "threshold": { "warning": 0, "critical": 0 } },
    "uptime": { "current": 86400, "unit": "s" }
  },
  "events": []
}
```

#### `GET /v1/device/:imei/metrics/export`

Export metrics (file download). `format=json|csv`, `range`, `metrics`.

---

### 14.4 Telemetry Endpoint

#### `GET /v1/device/:imei/telemetry`

Historical raw telemetry frames for a device (org-scoped).

**Headers:** `X-Organization-ID: <orgId>`

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `startTime` | int64 | -6h | Start timestamp (ms) |
| `endTime` | int64 | now | End timestamp (ms) |
| `limit` | int | 500 | Max results (max 10000) |

**Response (200 OK):**
```json
{
  "frames": [
    { "timestamp": 1718900000000, "riskScore": 45, "thermalTemp": 38.5, "bufferLevel": 67, "uptime": 86400 }
  ],
  "stats": {
    "riskScore": { "current": 72, "avg": 45, "min": 32, "max": 78 },
    "thermalTemp": { "current": 45.2, "avg": 42, "min": 38, "max": 52 },
    "bufferLevel": { "current": 67, "avg": 55, "min": 12, "max": 89 }
  }
}
```

---

### 14.5 Dashboard Stats Endpoint

#### `GET /v1/dashboard/stats`

Aggregated dashboard statistics for the active org.

**Headers:** `X-Organization-ID: <orgId>`

**Response (200 OK):**
```json
{
  "devices": { "total": 0, "online": 0, "offline": 0 },
  "commands": { "totalToday": 0, "pending": 0, "failed": 0 },
  "activity": { "last24h": { "commands": 0, "registrations": 0, "deregistrations": 0 } }
}
```

> **Server scoping caveat (org-scoping gap):** The `devices` list in dashboard stats is properly
> org-filtered (`ListByOrganization`). However the pending-command count
> (`commandRepo.CountPending(ctx)`) and the registration/deregistration log counts
> (`logsRepo.CountLogs(ctx, "", ...)`) are currently **global, not org-filtered**. This is a known
> server bug to fix before dashboard stats can be trusted in a multi-org tenant. Tracking issue
> needed in `SERVER_BACKEND_DASHBOARD_COMMANDS_API.md`.

---

## 15. GraphQL Schema

> **GraphQL status (v2.0 - PARTIAL fallback only):** REST is the primary data layer. GraphQL
> exists as a partial fallback for **commands and logs queries only**. The command mutations
> (`sendCommand`, `cancelCommand`, `retryCommand`) currently **omit the `organizationId` argument**
> - this is a gap vs the REST layer and the org model and must be added before they can be used as
> a fallback. There are **no** GraphQL modules for metrics, telemetry, or dashboard stats - those
> features are REST-only. The realtime log subscription is device-scoped (`deviceId`) and does not
> take an orgId (org scoping is enforced by the subscriber's device membership).


### 15.1 Types

```graphql
enum CommandStatus { PENDING, DELIVERED, COMPLETED, FAILED, CANCELLED }
enum LogEventType { CONNECTED, DISCONNECTED, TELEMETRY, COMMAND_SENT, COMMAND_ACK, ERROR, WARNING }

type Command {
  dispatchId: ID!
  command: String!
  status: CommandStatus!
  sentAt: DateTime!
  deliveredAt: DateTime
  latencyMs: Int
}

type LogEvent {
  id: ID!
  type: LogEventType!
  timestamp: DateTime!
  data: JSON
}

type CommandConnection {
  commands: [Command!]!
  pagination: PaginationInfo!
}

type LogConnection {
  events: [LogEvent!]!
  hasMore: Boolean!
  nextCursor: String
}
```

### 15.2 Queries

```graphql
type Query {
  commands(imei: String!, status: CommandStatus, page: Int, limit: Int, organizationId: ID!): CommandConnection!
  logs(imei: String!, type: LogEventType, startTime: Int, endTime: Int, limit: Int, cursor: String, organizationId: ID!): LogConnection!
  # telemetry / metrics / dashboardStats: NO GraphQL resolvers (REST only)
}
```

### 15.3 Mutations

```graphql
type Mutation {
  sendCommand(imei: String!, command: String!, organizationId: ID!): SendCommandResponse!   # <-- organizationId TODO (currently omitted)
  cancelCommand(imei: String!, dispatchId: ID!, organizationId: ID!): CancelCommandResponse!  # <-- organizationId TODO (currently omitted)
  retryCommand(imei: String!, dispatchId: ID!, organizationId: ID!): RetryCommandResponse!     # <-- TODO (currently omitted)
}

type SendCommandResponse {
  success: Boolean!
  dispatchId: String
  delivery: String
  serverTime: Int
  error: String
}

type CancelCommandResponse {
  success: Boolean!
  dispatchId: String!
  status: CommandStatus!
  error: String
}
```

---

## 16. Domain Types

> **Location & status (v2.0):** Domain types live in `packages/API_Client/src/domain/`
> (the `@vyzorix/api-client` package) - NOT in the web app. The commands, logs, metrics, and
> telemetry domain modules are DONE. The code samples below are illustrative of the shape; the
> authoritative definitions are the entity files listed in Section 3.1. `DashboardStats` lives in
> `domain/metrics/` (no separate dashboard domain). Org scoping is enforced at the data layer
> (REST `organizationId` param / `X-Organization-ID` header); domain entities themselves do not
> carry `organizationId` (it is a transport/auth concern, not an entity field).


### 16.1 Commands Domain (`domain/commands/`)

**types.ts:**
```typescript
export enum CommandStatus {
  PENDING = "PENDING",
  DELIVERED = "DELIVERED",
  COMPLETED = "COMPLETED",
  FAILED = "FAILED",
  CANCELLED = "CANCELLED",
}

export interface Command {
  dispatchId: string;
  command: string;
  status: CommandStatus;
  sentAt: Date;
  deliveredAt?: Date;
  latencyMs?: number;
}

export interface PresetCommand {
  id: string;
  label: string;
  description: string;
  danger: "low" | "high";
}

export const PRESET_COMMANDS: PresetCommand[] = [
  { id: "FORCE_SPEAKER", label: "Force Speaker", description: "Force speaker on", danger: "low" },
  { id: "RESET_AUDIO_HAL", label: "Reset Audio HAL", description: "Soft HAL reset", danger: "medium" },
  { id: "TOGGLE_CAPTURE", label: "Toggle Capture", description: "Start/stop capture", danger: "low" },
  { id: "REINIT_PROJECTION", label: "Reinit Projection", description: "Re-initiate projection", danger: "medium" },
  { id: "DUMP_FLIGHT_DATA", label: "Dump Flight Data", description: "Gather metrics", danger: "low" },
  { id: "UPLOAD_CRASH_ZIP", label: "Upload Crash Zip", description: "Upload crash logs", danger: "low" },
  { id: "SET_LOG_LEVEL", label: "Set Log Level", description: "Modify log level", danger: "low" },
  { id: "WAKE_UP_UPDATER", label: "Wake Updater", description: "Run update checker", danger: "low" },
];
```

**transforms.ts:**
```typescript
import type { Command, CommandStatus } from "./types";

interface RawCommand {
  dispatchId: string;
  command: string;
  status: string;
  sentAt: string;
  deliveredAt?: string;
  latencyMs?: number;
}

export const commandFromRaw = (raw: RawCommand): Command => ({
  dispatchId: raw.dispatchId,
  command: raw.command,
  status: raw.status as CommandStatus,
  sentAt: new Date(raw.sentAt),
  deliveredAt: raw.deliveredAt ? new Date(raw.deliveredAt) : undefined,
  latencyMs: raw.latencyMs,
});
```

**validation.ts:**
```typescript
import { PRESET_COMMANDS } from "./presets";

export const isValidCommand = (command: string): boolean => {
  return PRESET_COMMANDS.some(p => p.id === command);
};

export const isDangerousCommand = (command: string): boolean => {
  const preset = PRESET_COMMANDS.find(p => p.id === command);
  return preset?.danger === "high";
};
```

---

### 16.2 Logs Domain (`domain/logs/`)

**types.ts:**
```typescript
export enum LogEventType {
  CONNECTED = "CONNECTED",
  DISCONNECTED = "DISCONNECTED",
  TELEMETRY = "TELEMETRY",
  COMMAND_SENT = "COMMAND_SENT",
  COMMAND_ACK = "COMMAND_ACK",
  ERROR = "ERROR",
  WARNING = "WARNING",
  INFO = "INFO",
}

export interface LogEntry {
  id: string;
  type: LogEventType;
  timestamp: Date;
  data?: Record<string, unknown>;
}
```

**transforms.ts:**
```typescript
import type { LogEntry, LogEventType } from "./types";

interface RawLogEntry {
  id: string;
  type: string;
  timestamp: string;
  data?: Record<string, unknown>;
}

export const logFromRaw = (raw: RawLogEntry): LogEntry => ({
  id: raw.id,
  type: raw.type as LogEventType,
  timestamp: new Date(raw.timestamp),
  data: raw.data,
});
```

---

## 17. Data Layer

> **Data layer status (v2.0):** REST is the **primary** data layer and is fully implemented
> for commands, logs, metrics, and telemetry in `packages/API_Client/src/vyzorServer/rest/`.
> Each REST method takes an optional `organizationId?: string` that defaults to
> `getOrganizationContext()`; the axios interceptor auto-injects `X-Organization-ID`. GraphQL is
> a **partial fallback** (commands/logs queries only; mutations lack `organizationId` - see
> Section 15). The v1.x `lib/api/` paths below are deprecated; actual paths are in
> `packages/API_Client/src/vyzorServer/graphql/{commands,logs}/`. There are no metrics/telemetry
> GraphQL clients. The illustrative snippets below show the intended org-scoped shape.


### 17.1 GraphQL Queries (`lib/api/graphql/queries/`)

**commands.ts:**
```typescript
import { gql } from "graphql-request";
import { COMMAND_FRAGMENT } from "../fragments/command.fragment";

export const GET_COMMANDS = gql`
  ${COMMAND_FRAGMENT}
  query GetCommands($imei: String!, $status: CommandStatus, $page: Int, $limit: Int) {
    commands(imei: $imei, status: $status, page: $page, limit: $limit) {
      commands { ...CommandFields }
      pagination { page limit total totalPages }
    }
  }
`;

export const SEND_COMMAND = gql`
  mutation SendCommand($imei: String!, $command: String!) {
    sendCommand(imei: $imei, command: $command) {
      success dispatchId delivery serverTime error
    }
  }
`;
```

**logs.ts:**
```typescript
import { gql } from "graphql-request";
import { LOG_ENTRY_FRAGMENT } from "../fragments/log-entry.fragment";

export const GET_LOGS = gql`
  ${LOG_ENTRY_FRAGMENT}
  query GetLogs($imei: String!, $type: LogEventType, $startTime: Int, $endTime: Int, $limit: Int) {
    logs(imei: $imei, type: $type, startTime: $startTime, endTime: $endTime, limit: $limit) {
      events { ...LogEntryFields }
      hasMore nextCursor
    }
  }
`;
```

### 17.2 GraphQL Fragments (`lib/api/graphql/fragments/`)

**command.fragment.ts:**
```typescript
import { gql } from "graphql-request";

export const COMMAND_FRAGMENT = gql`
  fragment CommandFields on Command {
    dispatchId command status sentAt deliveredAt latencyMs
  }
`;
```

**log-entry.fragment.ts:**
```typescript
import { gql } from "graphql-request";

export const LOG_ENTRY_FRAGMENT = gql`
  fragment LogEntryFields on LogEvent {
    id type timestamp data
  }
`;
```

---

## 18. Presentation Layer - Hooks

> **Status:** TODO (org-aware hardening). Hook files exist on disk in
> `apps/VyzoriX_web/src/hooks/{commands,logs,metrics,telemetry}/` but the web layer is scaffold-
> only and not yet production-hardened. Per the v2.0 org model, every hook MUST:
>
> 1. Resolve the active org via `useCurrentOrganizationId()` (from `hooks/_shared/use-current-context`).
> 2. Be **disabled (return idle) when `organizationId === null`** - no request leaves the browser
>    without an org context. Use `enabled: organizationId !== null && imei !== undefined`.
> 3. Forward `organizationId ?? undefined` to the REST data-layer call (the REST methods default
>    to `getOrganizationContext()`, but explicit forwarding is required for testability + clarity).
> 4. Use REST as the **primary** data layer, with GraphQL as a **fallback only for commands/logs
>    queries** (on REST rejection, when an org context exists). Metrics/telemetry/dashboard have
>    no GraphQL fallback (REST-only).
> 5. Include `organizationId` in every **TanStack Query key** so org switches never serve another
>    org's cached data. (Current query-key factories for `dashboardStats`, `deviceMetrics`,
>    `latestTelemetry`, `telemetryStats` omit `organizationId` - this is a bug to fix.)
> 6. Invalidate the relevant query keys on mutation success.
>
> The v1.x snippets below used `graphqlClient` directly with no org context - they are
> **deprecated**. The target shape is REST-first, org-aware, one hook per concern.

### 18.1 Commands Hooks (`apps/VyzoriX_web/src/hooks/commands/`)

One hook file `use-commands.ts` (plus barrel). Target exports (all org-aware, REST-primary):

| Hook | Backed by (primary) | GraphQL fallback | Notes |
|------|---------------------|-----------------|-------|
| `useCommandHistory(params, opts)` | `commands.getHistory` | `GET_COMMANDS` query | paginated; key includes orgId |
| `useCommand(dispatchId, opts)` | `commands.getByDispatchId` | - | single command status |
| `usePendingCommands(imei, opts)` | `commands.getPending` | `GET_PENDING_COMMANDS` | enabled when orgId+imei present |
| `useSendCommand()` | `commands.send` | none (mutation lacks orgId) | invalidates command history + pending |
| `useCancelCommand()` | `commands.cancel` | none | invalidates pending + history |
| `useRetryCommand()` | `commands.retry` | none | invalidates history + pending |
| `usePollCommandStatus(dispatchId, opts)` | `commands.pollCommandStatus` | - | polling query |

```typescript
// Target shape (illustrative - REST-primary, org-aware)
import { useQuery } from "@tanstack/react-query";
import { commands, type CommandStatus } from "@vyzorix/api-client";
import { queryKeys } from "@/lib/query-keys";
import { useCurrentOrganizationId } from "@/hooks/_shared/use-current-context";

export function useCommandHistory(params: { imei: string; status?: CommandStatus; page?: number; limit?: number }) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.commands(params.imei, { ...params, organizationId: organizationId ?? undefined }),
    queryFn: () => commands.getHistory({ ...params, organizationId: organizationId ?? undefined }),
    enabled: organizationId !== null && !!params.imei,
  });
}
```

### 18.2 Logs Hooks (`apps/VyzoriX_web/src/hooks/logs/`)

| Hook | Backed by (primary) | GraphQL fallback | Notes |
|------|---------------------|-----------------|-------|
| `useDeviceLogs(params, opts)` | `logs.list` | `GET_LOGS` query | cursor-paginated; key includes orgId |
| `useLog(id, opts)` | `logs.get` | - | single log entry |
| `useLogStats(params, opts)` | `logs.stats` | - | log statistics |
| `useLogStream(imei, opts)` | WebSocket subscription | - | realtime; org scoping via device membership |

### 18.3 Metrics Hooks (`apps/VyzoriX_web/src/hooks/metrics/`)

| Hook | Backed by (primary) | GraphQL fallback | Notes |
|------|---------------------|-----------------|-------|
| `useDeviceMetrics(imei, range, opts)` | `fetchDeviceMetrics` | none (REST-only) | key MUST include orgId (bug to fix) |
| `useDashboardStats(opts)` | `fetchDashboardStats` | none (REST-only) | key MUST include orgId (bug to fix); gated on orgId !== null |
| `useExportMetrics(imei, params)` | `exportMetrics` | none | triggers file download |

### 18.4 Telemetry Hooks (`apps/VyzoriX_web/src/hooks/telemetry/`)

| Hook | Backed by (primary) | GraphQL fallback | Notes |
|------|---------------------|-----------------|-------|
| `useTelemetryHistory(deviceId, params, opts)` | `queryTelemetryHistory` | none (REST-only) | key includes orgId |
| `useLatestTelemetry(deviceId, opts)` | `getLatestTelemetry` | none | key MUST include orgId (bug to fix) |
| `useTelemetryStats(deviceId, params, opts)` | `getTelemetryStats` | none | key MUST include orgId (bug to fix) |
| `useExportTelemetry(deviceId, params)` | `exportTelemetry` | none | file download |

### 18.5 Query Keys (org isolation)

`apps/VyzoriX_web/src/lib/query-keys.ts` factories MUST embed `organizationId`:

```typescript
// Target - every factory takes/threads organizationId
export const queryKeys = {
  commands: (imei, params) => ["commands", organizationId, imei, params] as const,
  command: (dispatchId) => ["command", organizationId, dispatchId] as const,
  pendingCommands: (imei) => ["commands", "pending", organizationId, imei] as const,
  logs: (imei, params) => ["logs", organizationId, imei, params] as const,
  logStats: (imei, params) => ["logs", "stats", organizationId, imei, params] as const,
  deviceMetrics: (imei, range) => ["metrics", organizationId, imei, range] as const,   // <-- add orgId
  dashboardStats: () => ["dashboard", "stats", organizationId] as const,                // <-- add orgId
  telemetryHistory: (deviceId, params) => ["telemetry", "history", organizationId, deviceId, params] as const,
  latestTelemetry: (deviceId) => ["telemetry", "latest", organizationId, deviceId] as const,  // <-- add orgId
  telemetryStats: (deviceId) => ["telemetry", "stats", organizationId, deviceId] as const,    // <-- add orgId
};
```

> **Pattern reference:** see `apps/VyzoriX_web/src/hooks/registration/` (DONE) for the
> REST-primary + GraphQL-fallback + org-gating pattern these hooks should follow.

---

## 19. UI Layer - Components

### 19.1 Commands Components (`components/commands/`)

**commands-send.tsx:**
```typescript
import { useSendCommand } from "@/hooks/commands";
import { PRESET_COMMANDS } from "@vyzorix/api-client/domain/commands";
import { Button } from "@/components/ui/button";
import { AlertTriangle } from "lucide-react";

interface CommandsSendProps {
  imei: string;
}

export const CommandsSend = ({ imei }: CommandsSendProps) => {
  const { mutate: sendCommand, isPending } = useSendCommand();

  const handleSend = (commandId: string) => {
    sendCommand({ imei, command: commandId });
  };

  return (
    <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
      {PRESET_COMMANDS.map((preset) => (
        <Button
          key={preset.id}
          variant={preset.danger === "high" ? "destructive" : "default"}
          onClick={() => handleSend(preset.id)}
          disabled={isPending}
        >
          {preset.danger === "high" && <AlertTriangle className="w-4 h-4 mr-2" />}
          {preset.label}
        </Button>
      ))}
    </div>
  );
};
```

**commands-pending.tsx:**
```typescript
import { usePendingCommands, useCancelCommand } from "@/hooks/commands";
import { Button } from "@/components/ui/button";
import { CommandStatusBadge } from "./command-status-badge";

interface CommandsPendingProps {
  imei: string;
}

export const CommandsPending = ({ imei }: CommandsPendingProps) => {
  const { commands, isLoading } = usePendingCommands({ imei });
  const { mutate: cancelCommand } = useCancelCommand();

  if (isLoading) return <div>Loading...</div>;
  if (commands.length === 0) return <div>No pending commands</div>;

  return (
    <div className="space-y-2">
      {commands.map((cmd) => (
        <div key={cmd.dispatchId} className="flex items-center justify-between p-3 border rounded">
          <div>
            <div className="font-medium">{cmd.command}</div>
            <div className="text-sm text-muted-foreground">
              Sent {formatDistanceToNow(cmd.sentAt)}
            </div>
          </div>
          <div className="flex items-center gap-2">
            <CommandStatusBadge status={cmd.status} />
            <Button variant="outline" size="sm" onClick={() => cancelCommand({ imei, dispatchId: cmd.dispatchId })}>
              Cancel
            </Button>
          </div>
        </div>
      ))}
    </div>
  );
};
```

**commands-history.tsx:**
```typescript
import { useState } from "react";
import { useCommandHistory } from "@/hooks/commands";
import { SearchInput } from "@/components/shared/search-input";
import { FilterSelect } from "@/components/shared/filter-select";
import { DataTable } from "@/components/shared/data-table";
import { CommandRow } from "./command-row";
import type { CommandStatus } from "@vyzorix/api-client/domain/commands";

interface CommandsHistoryProps {
  imei: string;
}

export const CommandsHistory = ({ imei }: CommandsHistoryProps) => {
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState<CommandStatus | "all">("all");
  const [page, setPage] = useState(1);
  const { commands, pagination, isLoading } = useCommandHistory({ imei, status, page });

  return (
    <div className="space-y-4">
      <div className="flex gap-4">
        <SearchInput placeholder="Search commands..." value={search} onChange={setSearch} />
        <FilterSelect
          value={status}
          onValueChange={setStatus}
          options={[
            { value: "all", label: "All" },
            { value: "pending", label: "Pending" },
            { value: "delivered", label: "Delivered" },
            { value: "completed", label: "Completed" },
            { value: "failed", label: "Failed" },
          ]}
        />
      </div>
      <DataTable
        data={commands}
        columns={["Command", "Status", "Sent", "Latency"]}
        renderRow={(cmd) => <CommandRow key={cmd.dispatchId} command={cmd} />}
        pagination={pagination}
        onPageChange={setPage}
      />
    </div>
  );
};
```

**index.ts:**
```typescript
export { CommandsSend } from "./commands-send";
export { CommandsPending } from "./commands-pending";
export { CommandsHistory } from "./commands-history";
export { CommandButton } from "./command-button";
export { CommandStatusBadge } from "./command-status-badge";
export { CommandRow } from "./command-row";
```

---

### 19.2 Logs Components (`components/logs/`)

**logs-stream.tsx:**
```typescript
import { useEffect, useRef } from "react";
import { useLogs } from "@/hooks/logs";
import { LogEntry } from "./log-entry";
import { LogFilters } from "./log-filters";
import { useState } from "react";

interface LogsStreamProps {
  imei: string;
}

export const LogsStream = ({ imei }: LogsStreamProps) => {
  const [filter, setFilter] = useState<string>("all");
  const [autoScroll, setAutoScroll] = useState(true);
  const { logs, isLoading } = useLogs({ imei, type: filter as LogEventType });
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (autoScroll && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  return (
    <div className="space-y-4">
      <LogFilters filter={filter} onFilterChange={setFilter} autoScroll={autoScroll} onAutoScrollChange={setAutoScroll} />
      <div ref={containerRef} className="h-[400px] overflow-y-auto space-y-1 font-mono text-sm">
        {logs.map((log) => (
          <LogEntry key={log.id} log={log} />
        ))}
      </div>
    </div>
  );
};
```

**index.ts:**
```typescript
export { LogsStream } from "./logs-stream";
export { LogEntry } from "./log-entry";
export { LogFilters } from "./log-filters";
```

---

## 20. File Changes Summary

> **Authoritative status (v2.0):** driven by the server + API Client contracts. The API Client
> domain + REST data layers are DONE. GraphQL is partial. The web presentation layer exists but is
> flagged TODO for org-aware hardening; the web UI pages/routes are empty scaffolds.

### 20.1 DONE Files (authoritative)

#### API Client - Domain Layer (`packages/API_Client/src/domain/`) - DONE
| File | Purpose |
|------|---------|
| `domain/commands/command-entity.ts` | Command, CommandStatus, CommandParams, PresetCommandType |
| `domain/commands/command-mappers.ts` | commandFromRaw() |
| `domain/commands/command-validators.ts` | validateCommand() |
| `domain/commands/command-constants.ts` | PRESET_COMMANDS |
| `domain/logs/log-entity.ts` | LogEntry, LogListResult, LogLevel, LogStats |
| `domain/logs/log-mappers.ts` | logFromRaw() |
| `domain/logs/log-filters.ts` | validateLogEntry() |
| `domain/metrics/metrics-entity.ts` | DeviceMetrics, DashboardStats, TimeRange, MetricResolution |
| `domain/telemetry/telemetry-entity.ts` | TelemetryFrame, TelemetryHistoryResponse, TelemetryStats |

#### API Client - Data Layer - REST (`packages/API_Client/src/vyzorServer/rest/`) - DONE (org-scoped)
| File | Purpose |
|------|---------|
| `rest/commands/command-endpoints.ts` | commands REST object (getHistory, getPending, send, cancel, retry, pollCommandStatus) |
| `rest/logs/log-endpoints.ts` | logs REST object (list, get, stats) |
| `rest/metrics/metrics-endpoints.ts` | fetchDeviceMetrics, fetchDashboardStats, exportMetrics |
| `rest/telemetry/telemetry-endpoints.ts` | queryTelemetryHistory, getLatestTelemetry, getTelemetryStats, exportTelemetry |

#### API Client - Data Layer - GraphQL (`packages/API_Client/src/vyzorServer/graphql/`) - PARTIAL
| File | Purpose | Status |
|------|---------|--------|
| `graphql/commands/graphql-commands-queries.ts` | GET_COMMANDS, GET_PENDING_COMMANDS (org-scoped) | DONE |
| `graphql/commands/graphql-commands-mutations.ts` | SEND_COMMAND, CANCEL_COMMAND, RETRY_COMMAND | TODO: add `organizationId` arg |
| `graphql/logs/graphql-logs-queries.ts` | GET_LOGS (org-scoped) | DONE |
| `graphql/logs/graphql-logs-subscriptions.ts` | realtime log subscription (device-scoped) | DONE |
| (metrics / telemetry / dashboard graphql) | - | NOT PLANNED (REST-only) |

#### Server Backend - Handlers (`apps/api/internal/api/handlers/`) - DONE (org-scoped via middleware)
| File | Purpose |
|------|---------|
| `handlers/command/command_history_handler.go` | GET /v1/device/:imei/commands |
| `handlers/command/command_execute.go` | POST/DELETE/retry/status/pending command handlers |
| `handlers/device/device_logs_handler.go` | GET /v1/device/:imei/logs |
| `handlers/device/device_metrics_handler.go` | GET /v1/device/:imei/metrics + export |
| `handlers/device/device_telemetry_handler.go` | GET /v1/device/:imei/telemetry |
| `handlers/dashboard/dashboard_stats_handler.go` | GET /v1/dashboard/stats (partial org bug - see 14.5) |
| `middleware/org_context.go` | resolves X-Organization-ID |

### 20.2 TODO Files (web presentation + UI)

#### Presentation Layer - Hooks (`apps/VyzoriX_web/src/hooks/`) - TODO (org-aware hardening)
| File | Purpose |
|------|---------|
| `hooks/commands/use-commands.ts` | useCommandHistory, useCommand, usePendingCommands, useSendCommand, useCancelCommand, useRetryCommand, usePollCommandStatus |
| `hooks/commands/index.ts` | barrel |
| `hooks/logs/use-logs.ts` | useDeviceLogs, useLog, useLogStats |
| `hooks/logs/use-log-stream.ts` | realtime log streaming |
| `hooks/logs/index.ts` | barrel |
| `hooks/metrics/use-metrics.ts` | useDeviceMetrics, useDashboardStats, useExportMetrics |
| `hooks/metrics/index.ts` | barrel |
| `hooks/telemetry/use-telemetry.ts` | useTelemetryHistory, useLatestTelemetry, useTelemetryStats, useExportTelemetry |
| `hooks/telemetry/index.ts` | barrel |
| `hooks/shared/use-pagination.ts` | generic pagination |
| `hooks/shared/use-export.ts` | CSV/JSON export |
| `lib/query-keys.ts` | FIX: add `organizationId` to dashboardStats/deviceMetrics/latestTelemetry/telemetryStats keys |

#### State Stores (`apps/VyzoriX_web/src/stores/`) - TODO (see Section 23)
| File | Purpose |
|------|---------|
| `stores/log-stream-store.ts` | realtime log ring buffer + filters |
| `stores/metrics-realtime-store.ts` | live metric point accumulation |
| `stores/dashboard-store.ts` | dashboard aggregation + polling cadence |
| `stores/command-queue-store.ts` | offline command queue + retry |

#### UI Layer (`apps/VyzoriX_web/src/ui/`) - TODO (empty scaffolds)
| File | Purpose |
|------|---------|
| `ui/components/shared/*` | section, empty-state, data-table, pagination, search-input, filter-select, status-badge |
| `ui/components/commands/*` | commands-send, commands-pending, commands-history, command-row, command-status-badge |
| `ui/components/logs/*` | logs-stream, log-entry, log-filters, log-stats |
| `ui/components/dashboard/*` | dashboard-page, dashboard-overview, dashboard-metrics, device-stats-grid, activity-feed |
| `ui/pages/dashboard/*` | (empty .gitkeep) |
| `ui/pages/commands/*` | (empty .gitkeep) |
| `ui/pages/logs/*` | (empty .gitkeep) |

#### Routes (`apps/VyzoriX_web/src/routes/`) - TODO
| File | Purpose |
|------|---------|
| `routes/dashboard.tsx` + `dashboard.{overview,metrics,commands,logs}.tsx` | dashboard routes |
| `routes/commands-page.tsx`, `commands.{pending,history}.tsx` | standalone commands |
| `routes/logs-page.tsx` | standalone logs |
| `routes/device.tsx`, `device.$imei.commands*.tsx` | device commands tabs |

### 20.3 Total File Count

| Category | Done | TODO |
|----------|------|------|
| API Client Domain | ~9 | 0 |
| API Client REST Data | 4 | 0 |
| API Client GraphQL Data | 4 | 1 (mutations orgId) |
| Server Handlers/Middleware | ~7 | 0 |
| Web Hooks | 0 (scaffold) | ~12 |
| Web Stores | 6 existing | 4 new |
| Web UI Components | 0 | ~25 |
| Web Pages | 0 | ~6 |
| Web Routes | 0 | ~14 |

---

## 21. Implementation Order

> **Status (v2.0):** Phases 1-2 (backend + API Client domain/data) are DONE. Phase 3 (web hooks)
> is scaffold-only and TODO for org-aware hardening. Phases 4-7 (UI) are TODO. Server-side work is
> tracked in `SERVER_BACKEND_DASHBOARD_COMMANDS_API.md`.

### Phase 1: Backend APIs + API Client Domain/Data - DONE
- Server handlers (org-scoped via `org_context.go` middleware) - DONE
- API Client domain entities (commands/logs/metrics/telemetry) - DONE
- API Client REST endpoints (org-scoped via `getOrganizationContext()` + interceptor) - DONE
- API Client GraphQL queries (commands/logs, org-scoped) - DONE
- API Client GraphQL command mutations - TODO: add `organizationId` arg

### Phase 2: State Stores - TODO (see Section 23)
1. `command-queue-store.ts` (offline dispatch + retry)
2. `log-stream-store.ts` (realtime ring buffer + filters)
3. `metrics-realtime-store.ts` (live point accumulation)
4. `dashboard-store.ts` (aggregation + polling cadence)

### Phase 3: Presentation Layer - Hooks - TODO (org-aware hardening)
1. Harden `hooks/commands/use-commands.ts` (REST-primary, GraphQL fallback for queries, org gating, query keys with orgId)
2. Harden `hooks/logs/use-logs.ts` + `use-log-stream.ts`
3. Harden `hooks/metrics/use-metrics.ts` (REST-only, org gating)
4. Harden `hooks/telemetry/use-telemetry.ts` (REST-only, org gating)
5. FIX `lib/query-keys.ts`: add `organizationId` to `dashboardStats`/`deviceMetrics`/`latestTelemetry`/`telemetryStats`
6. Write tests (vitest + `renderHookWithQueryClient`, mock API client via `vi.hoisted`) - follow registration hook test pattern

### Phase 4: UI Layer - Shared Components - TODO
1. Shared components (section, data-table, pagination, search-input, filter-select, status-badge)

### Phase 5: UI Layer - Feature Components - TODO
1. Commands components (send, pending, history, row, status-badge)
2. Logs components (stream, entry, filters, stats)
3. Dashboard components (page, overview, metrics, device-stats-grid, activity-feed)

### Phase 6: Routes - TODO
1. `/commands` routes
2. `/dashboard/*` routes
3. `/logs` route
4. Update device routes for commands tabs

### Phase 7: Polish - TODO
1. Time range selector
2. Export functionality (CSV/JSON)
3. Loading + error states
4. Realtime WebSocket integration (logs stream + metrics live points)

## 22. Device Commands Reference

All commands are HMAC-SHA256 signed. See `COMMAND_SECURITY.md` for signing specification.

| Command | Parameters | Description |
|---------|------------|-------------|
| `FORCE_SPEAKER` | None | Force speaker on with reassertion loop |
| `RESET_AUDIO_HAL` | None | Soft HAL reset via BT stream cycling |
| `TOGGLE_CAPTURE` | `active` (boolean) | Start/stop AudioRecord read loops |
| `REINIT_PROJECTION` | None | Re-initiate media projection via notification |
| `DUMP_FLIGHT_DATA` | None | Gather local metrics → JSON postback |
| `UPLOAD_CRASH_ZIP` | None | Zip SQLite logs → POST binary |
| `SET_LOG_LEVEL` | `level` (string) | Dynamically modify Logger minLogLevel |
| `WAKE_UP_UPDATER` | None | Override WorkManager delays → run UpdateChecker |

---

*Document Version: 2.0*
*Status: Draft (org-model realigned; web TODO)*
*Architecture: Layered (Following FRONTEND_ARCHITECTURE.md)*


---

## 23. State Stores

> **Purpose (v2.0):** Maps the Zustand stores required to support the dashboard/commands/logs/metrics/telemetry UX. The web app already has 6 stores; this section identifies the **4 new stores** needed and clarifies the boundary between TanStack Query (server cache) and Zustand (ephemeral/UI/realtime state). Driven by the server + API Client contracts; the web is TODO.

### 23.1 Store selection principles (senior)

1. **TanStack Query owns server state** (commands history, logs, metrics, telemetry, dashboard stats). Stores do NOT duplicate server cache. Stores own only what the server does not: realtime stream buffers, UI selection, ephemeral dispatch tracking, offline queues, and aggregation cadence.
2. **Org isolation:** every store that holds org-scoped data MUST key its state by `organizationId` (or clear on org switch) so a user switching orgs never sees another org's pending commands, log stream, or live metrics. Subscribe to `useAuthStore.organizationId` and reset scoped slices on change.
3. **Realtime writes go to stores, not Query cache:** WebSocket log/metric events are high-frequency and append-only; writing each into the Query cache thrashes subscribers. Instead, accumulate in a store and let the UI subscribe; periodically (or on unmount) the store can flush a summary into Query for persistence.
4. **Offline resilience:** commands dispatched while offline are queued in a store and flushed to REST when connectivity returns (the API Client already has a connectivity monitor + request queue; the store is the UI-facing mirror).

### 23.2 Existing stores (DONE - do not recreate)

| Store | File | Owns | Org-aware? |
|-------|------|------|-----------|
| `useAuthStore` | `stores/auth-store.ts` | auth state + active `organizationId` (source of truth) | n/a (holds orgId) |
| `useWebSocketStore` | `stores/websocket-store.ts` | WS connection + subscription state | subscribe scoped to org/device |
| `useConnectivityStore` | `stores/connectivity-store.ts` | online status + effective type | global |
| `useDeviceSelectorStore` | `stores/device-selector-store.ts` | selected device + filters | scoped to active org |
| `useThemeStore` | `stores/theme-store.ts` | theme mode | global |
| `useCommandDispatchStore` | `stores/command-dispatch-store.ts` | in-memory pending-command tracking (`pendingCommands` map, `addPending`/`removePending`/`clearPending`/`pendingCount`) | **MUST scope by orgId** (currently global) |

### 23.3 New stores (TODO)

#### 23.3.1 `command-queue-store.ts` — offline command dispatch queue

**Why:** Commands sent while offline (or while the WS is down) must be queued and replayed in order once connectivity returns, with per-command retry/backoff and user-visible status.

| Aspect | Decision |
|--------|----------|
| State shape | `{ queue: QueuedCommand[]; isFlushing: boolean; lastFlushError: string \| null }` where `QueuedCommand = { id, imei, command, organizationId, queuedAt, attempts, nextRetryAt, status }` |
| Actions | `enqueue(cmd)`, `dequeue(id)`, `markDispatched(id, dispatchId)`, `markFailed(id, error)`, `flush()`, `clear()` |
| Org isolation | `queue` filtered by `organizationId` on read; `flush()` only sends commands for the active org; switching org does NOT drop another org's queued commands |
| Persistence | `persist` middleware (localStorage) so queued commands survive reload |
| Integration | `flush()` calls `commands.send` (REST); on success -> `markDispatched` + `useCommandDispatchStore.addPending(dispatchId)`; on failure -> increment `attempts`, exponential backoff via `nextRetryAt` |
| Subscriptions | Subscribes to `useConnectivityStore.isOnline` to trigger `flush()` when back online |
| Tests | enqueue/dequeue ordering, org filtering, flush-on-reconnect, backoff |

#### 23.3.2 `log-stream-store.ts` — realtime log ring buffer

**Why:** The WS log subscription emits high-frequency events; the UI needs a bounded, filterable, auto-scrolling buffer without re-rendering the whole page per event.

| Aspect | Decision |
|--------|----------|
| State shape | `{ byDevice: Record<string, LogEntry[]>; filters: { type?: LogEventType; search?: string }; autoScroll: boolean; hasMore: boolean; nextCursor: string \| null }` |
| Actions | `append(deviceId, entry)`, `appendBatch(deviceId, entries)`, `setFilter(filter)`, `toggleAutoScroll()`, `clear(deviceId)`, `trim(deviceId, max=500)` |
| Buffer strategy | Ring buffer per device (cap ~500 entries); older entries dropped on overflow; `trim` called on `append` |
| Org isolation | Subscribe/unsubscribe to WS on org change; clear `byDevice` on org switch (a device from org A is not visible in org B) |
| Integration | `useLogStream` hook subscribes to WS (`useWebSocketStore`) and calls `append`; reads from store for render; historical backfill via `logs.list` REST merged at the cursor boundary |
| Persistence | none (ephemeral) |
| Tests | ring-buffer trim, filter application, org-switch clears, append ordering |

#### 23.3.3 `metrics-realtime-store.ts` — live metric point accumulation

**Why:** Live telemetry WS frames must update charts smoothly; accumulating points in a store (with a sliding window) avoids thrashing the Query cache and gives the chart a stable data source.

| Aspect | Decision |
|--------|----------|
| State shape | `{ byDevice: Record<string, MetricPoint[]>; windowMs: number; lastFrame: Record<string, TelemetryFrame \| null> }` where `MetricPoint = { t, riskScore, thermalTemp, bufferLevel, uptime }` |
| Actions | `push(deviceId, frame)`, `setWindow(ms)`, `clear(deviceId)`, `getSeries(deviceId, metric)` |
| Window strategy | Sliding window (default `windowMs = 6h`); `push` drops points older than `now - windowMs` |
| Org isolation | clear `byDevice` on org switch; subscribe to WS scoped to active org's devices |
| Integration | `useDeviceMetrics` hook seeds from REST (`fetchDeviceMetrics`) then live-updates from the store via WS; `useLatestTelemetry` reads `lastFrame` |
| Persistence | none |
| Tests | sliding-window eviction, org-switch clears, series extraction per metric |

#### 23.3.4 `dashboard-store.ts` — dashboard aggregation + polling cadence

**Why:** The dashboard overview combines `dashboardStats` (REST) + device counts + recent activity + (optionally) live metric summaries. A store coordinates the polling cadence, the "last refreshed" timestamp, and the merge of REST snapshot + realtime updates so the UI has one coherent view.

| Aspect | Decision |
|--------|----------|
| State shape | `{ stats: DashboardStats \| null; lastRefreshedAt: number \| null; isRefreshing: boolean; refreshIntervalMs: number; recentActivity: ActivityItem[] }` |
| Actions | `setStats(stats)`, `setRefreshing(bool)`, `setRefreshInterval(ms)`, `pushActivity(item)`, `clear()` |
| Polling | Coowns a `setInterval` (default 30s) that triggers `useDashboardStats` refetch; pauses when tab hidden (`document.visibilityState`) and when orgId is null |
| Org isolation | `clear()` on org switch; polling loop re-reads active org from `useAuthStore` each tick (do not close over a stale orgId) |
| Integration | `useDashboardStats` (REST) writes into the store on success; `useCommandDispatchStore`/`useLogStream` push `ActivityItem`s for "command sent"/"log alert" feed entries |
| Persistence | none |
| Tests | poll cadence, visibility pause, org-switch clears, activity merge ordering |

### 23.4 Store -> Query -> UI wiring

```
  WebSocket events -> useWebSocketStore -> log-stream-store / metrics-realtime-store -> UI (subscribe)
                                                     | (periodic summary flush)
                                                     v
                                              TanStack Query (server cache) -> UI (useQuery)

  User sends command -> command-queue-store.enqueue -> (offline?) persist
                                   | (online)
                                   v
                          commands.send (REST) -> useCommandDispatchStore.addPending -> UI

  Dashboard -> dashboard-store (poll cadence) -> useDashboardStats (REST) -> store.setStats -> UI
```

### 23.5 Open questions to resolve before implementation

1. Should the `command-queue-store` reuse the API Client's existing connectivity monitor + request queue (it already has `subscribe`/`flushQueue`), or remain a UI-only mirror? Recommendation: reuse the API Client queue for transport, and let the store be the UI-facing status mirror.
2. Realtime log/metric WS messages — confirm the server emits them org-scoped (subscription by `deviceId`, org enforced via device membership). If a single WS connection serves multiple orgs, the store must filter by org on receipt.
3. Ring-buffer size (500) and metrics sliding window (6h) — confirm against memory budget for long-running dashboard sessions.
4. Whether `dashboard-store` polling should be replaced by a WS "stats delta" subscription once the server supports it (REST polling is the v2.0 baseline).
