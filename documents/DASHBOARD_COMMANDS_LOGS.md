# Dashboard, Commands & Logs - Enterprise Requirements Specification

> **Version:** 1.2  
> **Status:** Draft  
> **Created:** 2026-06-21  
> **Updated:** 2026-06-24  
> **Target:** Production MVP  
> **Architecture:** Layered (Following `FRONTEND_ARCHITECTURE.md`)  

---

>  **Architecture Alignment Note (v1.2)**
> 
> This document has been updated to align with the **Layered Architecture** defined in `FRONTEND_ARCHITECTURE.md`. The file structure below follows the **4-layer architecture**:
> - **UI Layer** (`src/components/`) - Pure UI rendering, imports only from hooks
> - **Presentation Layer** (`src/hooks/`) - UI logic, state management, imports from domain & data
> - **Domain Layer** (`packages/API_Client/src/domain/`) - Types, transforms, validation (NO external imports)
> - **Data Layer** (`packages/API_Client/src/vyzorServer/`) - API clients (GraphQL/REST), imports only domain types
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

---

## 1. Overview

### 1.1 Purpose

Redesign the Dashboard page with tabs for better organization, create a shared Commands page accessible from both Dashboard and Device, and maintain the Logs page.

### 1.2 Key Principles

- **Commands are shared** - Accessible from Dashboard tabs AND Device page
- **Logs are standalone** - Separate `/logs` route, accessible from Dashboard tabs
- **No duplication** - Same component used in multiple places via routing
- **Device context** - All pages use current device from config
- **Layered Architecture** - Follow FRONTEND_ARCHITECTURE.md dependency rules

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

                        FRONTEND ARCHITECTURE                        

                                                                     
     
                        UI LAYER                                  
                     (src/components/)                           
                                                                  
      Pages, Components, Shared UI                               
      ONLY renders UI. Uses hooks for everything.                 
      NEVER imports from Data or Domain.                          
     
                                                                     
                               uses                                  
                                                                     
     
                     PRESENTATION LAYER                           
                        (src/hooks/)                             
                                                                  
      Custom hooks that:                                         
      - Handle UI logic                                           
      - Transform data for UI                                     
      - Manage state                                              
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
      - Make HTTP requests                                       
      - Handle authentication                                     
      - Parse responses                                          
      NEVER imports from UI or Presentation.                     
     
                                                                     

```

### 2.2 Dependency Rule

```
UI Layer  Presentation Layer  Domain  Data Layer
(components/)          (hooks/)              (domain/)        (lib/api/)
                                                                 
                                                                 
      IMPORTS                                        
          ONLY FROM                                               
          HOOKS                                                   
                                                              
                                              IMPORTS 
                                                  ONLY FROM
                                                  DOMAIN TYPES
```

---

## 3. Target File Structure

### 3.1 Complete Directory Tree

```
apps/web/src/

 domain/                          # DOMAIN LAYER (follows FRONTEND_ARCHITECTURE.md)
    _shared/                   # SHARED domain types
       domain-pagination.ts  # Pagination types & helpers
       domain-errors.ts      # Domain error types
   
    commands/
       command-entity.ts     # Command, CommandStatus, PresetCommand
       command-mappers.ts    # commandFromRaw(), commandToApi()
       command-validators.ts # validateCommand(), validateStatus()
       command-constants.ts  # Preset command definitions
   
    logs/
       log-entity.ts         # LogEntry, LogLevel, LogSource
       log-mappers.ts        # logFromRaw()
       log-filters.ts        # validateLogEntry(), filter functions
   
    devices/                  # Shared device domain types
        device-entity.ts      # Device basic types
        device-mappers.ts     # deviceBasicFromRaw()

 lib/
    api/
        graphql/
           _shared/
              graphql-client.ts  # GraphQL client setup
          
           commands/
              graphql-commands-queries.ts     # GET_COMMANDS, GET_PENDING_COMMANDS
              graphql-commands-mutations.ts  # SEND_COMMAND, CANCEL_COMMAND, RETRY_COMMAND
              graphql-commands-fragments.ts  # Command fragments
              graphql-commands-types.ts     # Raw GraphQL response types
          
           logs/
              graphql-logs-queries.ts        # GET_LOG_ENTRIES, GET_LOG_STATS
              graphql-logs-subscriptions.ts # Real-time log subscription
              graphql-logs-fragments.ts     # Log entry fragments
              graphql-logs-types.ts         # Raw GraphQL response types
          
           devices/
               graphql-devices-queries.ts     # GET_DEVICES, GET_DEVICE, GET_DEVICE_COUNT
               graphql-devices-fragments.ts  # Device fragments
               graphql-devices-types.ts      # Raw GraphQL response types
       
        rest/
            _shared/
               rest-client.ts     # Base REST client
            commands/
               rest-commands-endpoints.ts    # REST endpoints for commands
            logs/
                rest-logs-endpoints.ts        # REST endpoints for logs

 hooks/                           # PRESENTATION LAYER
   
    commands/                    # Commands presentation logic
       use-commands.ts          # Send commands, get presets
       use-command-history.ts   # Command history with pagination
       use-pending-commands.ts  # Pending commands for device
       index.ts                 # Barrel export
   
    logs/                        # Logs presentation logic
       use-logs.ts              # (EXISTING - refactor to follow architecture)
       use-log-stream.ts        # Real-time log streaming
       index.ts                 # Barrel export
   
    devices/                     # Device presentation logic (shared)
       use-devices.ts           # Device list with filters
       use-device-selected.ts   # Current selected device context
       index.ts
   
    shared/                      # Shared presentation utilities
        use-pagination.ts        # Generic pagination hook
        use-search.ts            # Generic search/filter hook
        index.ts

 components/                      # UI LAYER
   
    shared/                      # Shared UI components (NOT feature-specific)
       section.tsx              # Bordered section component
       section-header.tsx       # Section header with title/subtitle
       empty-state.tsx          # Empty state component
       loading-skeleton.tsx     # Loading skeleton variants
       data-table.tsx           # Table wrapper with sorting/pagination
       pagination.tsx           # Pagination controls
       search-input.tsx         # Search input with clear
       filter-select.tsx        # Dropdown filter select
       status-badge.tsx         # (EXISTING - move here)
       connection-badge.tsx     # (EXISTING - move here)
   
    commands/                    # Commands feature components
       commands-send.tsx        # Send commands grid (PRESET_COMMANDS)
       commands-pending.tsx     # Pending queue list
       commands-history.tsx     # Full command history table
       commands-recent.tsx      # Recent commands list
       command-button.tsx       # Single command button
       command-status-badge.tsx # Command status badge
       command-row.tsx          # Single command table row
       index.ts                 # Barrel export
   
    logs/                        # Logs feature components
       logs-stream.tsx          # Real-time log display
       log-entry.tsx            # Single log entry row
       log-filters.tsx          # Log filtering controls
       log-stats.tsx            # Log statistics summary
       index.ts                 # Barrel export
   
    dashboard/                   # Dashboard feature components
       dashboard-page.tsx       # Dashboard page wrapper with tabs
       dashboard-overview.tsx   # Overview tab content
       dashboard-metrics.tsx    # Metrics tab content
       dashboard-commands.tsx   # Commands tab (redirects to /commands)
       dashboard-logs.tsx       # Logs tab (redirects to /logs)
       device-stats-grid.tsx    # Device statistics grid
       activity-feed.tsx        # Recent activity feed
       index.ts                 # Barrel export
   
    layout/                      # (EXISTING)
       app-layout.tsx
       auth-layout.tsx
   
    auth/                        # (EXISTING)
       ... (existing auth components)
   
    ui/                          # (EXISTING - base UI primitives)
        button.tsx
        badge.tsx
        card.tsx
        ... (shadcn/ui components)

 routes/                          # PAGE LAYER (Routes)
    
     __root.tsx                   # (EXISTING)
     router.tsx                   # (EXISTING)
    
     dashboard.tsx                # MODIFIED - redirect to /dashboard/overview
    
     dashboard.commands.tsx       # NEW - /dashboard/commands
     dashboard.commands.pending.tsx  # NEW - /dashboard/commands/pending
     dashboard.commands.history.tsx  # NEW - /dashboard/commands/history
     dashboard.logs.tsx           # NEW - /dashboard/logs
     dashboard.metrics.tsx        # NEW - /dashboard/metrics
    
     commands-page.tsx            # NEW - /commands (standalone)
     commands.pending.tsx         # NEW - /commands/pending
     commands.history.tsx         # NEW - /commands/history
    
     logs-page.tsx                # MODIFIED - standalone logs page
    
     device.tsx                   # MODIFIED - add commands tab
     device.$imei.commands.tsx    # NEW - /device/:imei/commands
     device.$imei.commands.pending.tsx  # NEW
     device.$imei.commands.history.tsx  # NEW
    
     ... (other existing routes)
```

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

At-a-glance status of current device with key metrics and quick actions.

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
| "View Device " | Dropdown to switch devices |
| Metrics | Click to navigate to Metrics tab |
| "Send Command" | Dropdown with preset commands |
| "View Logs" | Navigate to Logs tab |

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

### 13.1 Structure

```
Device Page (tabs):
 Inbox        → Pending registration requests
 Overview     → Device info, health, connection
 Telemetry    → Real-time charts, metrics
 History      → Historical data, export
 Commands     → CommandsPanel (shared component)
```

---

## 14. REST API Specification

### 14.1 Commands Endpoint

#### `POST /v1/device/:imei/command`

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

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | string | all | Filter: pending, delivered, completed, failed |
| `page` | int | 1 | Page number |
| `limit` | int | 20 | Items per page |

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
  "pagination": { "page": 1, "limit": 20, "total": 45, "totalPages": 3 }
}
```

---

#### `DELETE /v1/command/:dispatchId`

Cancel a pending command. Note: This endpoint is at the command root level, not nested under device.

**Response:**
```json
{ "dispatchId": "abc123def456", "status": "cancelled", "cancelled": true, "serverTime": 1718900600 }
```

---

### 14.2 Logs Endpoint

#### `GET /v1/device/:imei/logs`

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | string | all | Filter: connection, command, telemetry, error |
| `startTime` | int64 | -24h | Start timestamp (ms) |
| `endTime` | int64 | now | End timestamp (ms) |
| `limit` | int | 100 | Max results |
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

### 14.3 Telemetry Endpoint

#### `GET /v1/device/:imei/telemetry`

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `startTime` | int64 | -6h | Start timestamp (ms) |
| `endTime` | int64 | now | End timestamp (ms) |
| `limit` | int | 500 | Max results |

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

## 15. GraphQL Schema

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
  commands(imei: String!, status: CommandStatus, page: Int, limit: Int): CommandConnection!
  logs(imei: String!, type: LogEventType, startTime: Int, endTime: Int, limit: Int, cursor: String): LogConnection!
  telemetry(imei: String!, startTime: Int, endTime: Int, limit: Int): TelemetryResult!
}
```

### 15.3 Mutations

```graphql
type Mutation {
  sendCommand(imei: String!, command: String!): SendCommandResponse!
  cancelCommand(imei: String!, dispatchId: ID!): CancelCommandResponse!
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

### 18.1 Commands Hooks (`hooks/commands/`)

**use-commands.ts:**
```typescript
import { useMutation, useQuery } from "@tanstack/react-query";
import { graphqlClient } from "@vyzorix/api-client/vyzorServer/graphql/client";
import { GET_COMMANDS, SEND_COMMAND } from "@vyzorix/api-client/vyzorServer/graphql/queries/command-queries";
import { commandFromRaw } from "@vyzorix/api-client/domain/commands";
import type { Command, CommandStatus } from "@vyzorix/api-client/domain/commands";

interface UseCommandsOptions {
  imei: string;
  status?: CommandStatus;
  page?: number;
  limit?: number;
}

export const useCommands = (options: UseCommandsOptions) => {
  const { imei, status, page = 1, limit = 20 } = options;
  const { data, isLoading, error } = useQuery({
    queryKey: ["commands", imei, status, page, limit],
    queryFn: async () => {
      const response = await graphqlClient.query({
        query: GET_COMMANDS,
        variables: { imei, status, page, limit },
      });
      return response.data.commands;
    },
  });
  const commands = data?.commands.map(commandFromRaw) ?? [];
  const pagination = data?.pagination ?? null;
  return { commands, pagination, isLoading, error: error as Error | null };
};

export const useSendCommand = () => {
  return useMutation({
    mutationFn: async ({ imei, command }: { imei: string; command: string }) => {
      const response = await graphqlClient.mutation({
        mutation: SEND_COMMAND,
        variables: { imei, command },
      });
      return response.data.sendCommand;
    },
  });
};
```

**index.ts:**
```typescript
export { useCommands, useSendCommand, useCancelCommand } from "./use-commands";
export { useCommandHistory } from "./use-command-history";
export { usePendingCommands } from "./use-pending-commands";
```

---

### 18.2 Logs Hooks (`hooks/logs/`)

**use-logs.ts:**
```typescript
import { useQuery } from "@tanstack/react-query";
import { graphqlClient } from "@vyzorix/api-client/vyzorServer/graphql/client";
import { GET_LOGS } from "@vyzorix/api-client/vyzorServer/graphql/queries/log-queries";
import { logFromRaw } from "@vyzorix/api-client/domain/logs";
import type { LogEntry, LogEventType } from "@vyzorix/api-client/domain/logs";

interface UseLogsOptions {
  imei: string;
  type?: LogEventType;
  startTime?: number;
  endTime?: number;
  limit?: number;
}

export const useLogs = (options: UseLogsOptions) => {
  const { imei, type, startTime, endTime, limit = 100 } = options;
  const { data, isLoading, error } = useQuery({
    queryKey: ["logs", imei, type, startTime, endTime, limit],
    queryFn: async () => {
      const response = await graphqlClient.query({
        query: GET_LOGS,
        variables: { imei, type, startTime, endTime, limit },
      });
      return response.data.logs;
    },
  });
  const logs = data?.events.map(logFromRaw) ?? [];
  const hasMore = data?.hasMore ?? false;
  const nextCursor = data?.nextCursor ?? null;
  return { logs, hasMore, nextCursor, isLoading, error: error as Error | null };
};
```

**index.ts:**
```typescript
export { useLogs } from "./use-logs";
export { useLogStream } from "./use-log-stream";
```

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

### 20.1 NEW Files by Layer

#### Domain Layer (12 new files)

| File | Purpose |
|------|---------|
| `domain/shared/pagination.ts` | Pagination types |
| `domain/common/error.ts` | Domain error types |
| `domain/shared/types.ts` | Shared types |
| `domain/commands/command-types.ts` | Command types |
| `domain/commands/transforms.ts` | commandFromRaw() |
| `domain/commands/validation.ts` | validateCommand() |
| `domain/commands/presets.ts` | Preset definitions |
| `domain/logs/log-types.ts` | Log types |
| `domain/logs/transforms.ts` | logFromRaw() |
| `domain/logs/validation.ts` | validateLogEntry() |
| `domain/devices/device-types.ts` | Device types |
| `domain/devices/transforms.ts` | deviceFromRaw() |

#### Data Layer - GraphQL (6 new files)

| File | Purpose |
|------|---------|
| `lib/api/graphql/queries/command-queries.ts` | Command queries |
| `lib/api/graphql/queries/log-queries.ts` | Log queries |
| `lib/api/graphql/queries/devices.ts` | Device queries |
| `lib/api/graphql/mutations/command-mutations.ts` | Command mutations |
| `lib/api/graphql/fragments/command.fragment.ts` | Command fragment |
| `lib/api/graphql/fragments/log-entry.fragment.ts` | Log fragment |

#### Data Layer - REST (2 new files)

| File | Purpose |
|------|---------|
| `lib/api/rest/command-rest.ts` | REST endpoints |
| `lib/api/rest/log-rest.ts` | REST endpoints |

#### Presentation Layer - Hooks (12 new files)

| File | Purpose |
|------|---------|
| `hooks/commands/use-commands.ts` | Send/cancel commands |
| `hooks/commands/use-command-history.ts` | History with pagination |
| `hooks/commands/use-pending-commands.ts` | Pending commands |
| `hooks/commands/index.ts` | Barrel export |
| `hooks/logs/use-logs.ts` | Log queries |
| `hooks/logs/use-log-stream.ts` | Real-time streaming |
| `hooks/logs/index.ts` | Barrel export |
| `hooks/devices/use-devices.ts` | Device list |
| `hooks/devices/use-device-selected.ts` | Selected device |
| `hooks/devices/index.ts` | Barrel export |
| `hooks/shared/use-pagination.ts` | Generic pagination |
| `hooks/shared/use-search.ts` | Generic search |

#### UI Layer - Shared (8 new files)

| File | Purpose |
|------|---------|
| `components/shared/section.tsx` | Bordered section |
| `components/shared/section-header.tsx` | Section header |
| `components/shared/empty-state.tsx` | Empty state |
| `components/shared/loading-skeleton.tsx` | Loading skeleton |
| `components/shared/data-table.tsx` | Table wrapper |
| `components/shared/pagination.tsx` | Pagination |
| `components/shared/search-input.tsx` | Search input |
| `components/shared/filter-select.tsx` | Filter dropdown |

#### UI Layer - Commands (8 new files)

| File | Purpose |
|------|---------|
| `components/commands/commands-send.tsx` | Send grid |
| `components/commands/commands-pending.tsx` | Pending queue |
| `components/commands/commands-history.tsx` | History table |
| `components/commands/commands-recent.tsx` | Recent list |
| `components/commands/command-button.tsx` | Single button |
| `components/commands/command-status-badge.tsx` | Status badge |
| `components/commands/command-row.tsx` | Table row |
| `components/commands/index.ts` | Barrel export |

#### UI Layer - Logs (4 new files)

| File | Purpose |
|------|---------|
| `components/logs/logs-stream.tsx` | Log display |
| `components/logs/log-entry.tsx` | Single entry |
| `components/logs/log-filters.tsx` | Filters |
| `components/logs/log-stats.tsx` | Stats |
| `components/logs/index.ts` | Barrel export |

#### UI Layer - Dashboard (8 new files)

| File | Purpose |
|------|---------|
| `components/dashboard/dashboard-page.tsx` | Page wrapper |
| `components/dashboard/dashboard-overview.tsx` | Overview tab |
| `components/dashboard/dashboard-metrics.tsx` | Metrics tab |
| `components/dashboard/dashboard-commands.tsx` | Commands tab |
| `components/dashboard/dashboard-logs.tsx` | Logs tab |
| `components/dashboard/device-stats-grid.tsx` | Stats grid |
| `components/dashboard/activity-feed.tsx` | Activity feed |
| `components/dashboard/index.ts` | Barrel export |

#### Routes (14 new/modified files)

| File | Status | Purpose |
|------|--------|---------|
| `routes/dashboard.tsx` | **MODIFIED** | Redirect to sub-route |
| `routes/dashboard.overview.tsx` | **NEW** | Overview tab |
| `routes/dashboard.metrics.tsx` | **NEW** | Metrics tab |
| `routes/dashboard.commands.tsx` | **NEW** | Commands tab |
| `routes/dashboard.commands.pending.tsx` | **NEW** | Pending queue |
| `routes/dashboard.commands.history.tsx` | **NEW** | History |
| `routes/dashboard.logs.tsx` | **NEW** | Logs tab |
| `routes/commands-page.tsx` | **NEW** | Send commands |
| `routes/commands.pending.tsx` | **NEW** | Pending |
| `routes/commands.history.tsx` | **NEW** | History |
| `routes/logs-page.tsx` | **MODIFIED** | Standalone logs |
| `routes/device.tsx` | **MODIFIED** | Add Commands tab |
| `routes/device.$imei.commands.tsx` | **NEW** | Device commands |
| `routes/device.$imei.commands.pending.tsx` | **NEW** | Device pending |
| `routes/device.$imei.commands.history.tsx` | **NEW** | Device history |

### 20.2 Total File Count

| Category | New Files | Modified Files |
|----------|-----------|----------------|
| Domain Layer | 12 | 0 |
| Data Layer (GraphQL) | 6 | 2 |
| Data Layer (REST) | 2 | 1 |
| Presentation Layer | 12 | 0 |
| UI Layer (Shared) | 8 | 0 |
| UI Layer (Commands) | 8 | 0 |
| UI Layer (Logs) | 5 | 0 |
| UI Layer (Dashboard) | 8 | 0 |
| Routes | 13 | 3 |
| **TOTAL** | **74** | **6** |

---

## 21. Implementation Order

### Phase 1: Backend APIs (Day 1)
1. Implement `GET /v1/device/:imei/commands`
2. Implement `GET /v1/device/:imei/logs`
3. Implement `GET /v1/device/:imei/telemetry`
4. Update GraphQL resolvers

### Phase 2: Domain & Data Layer (Day 1-2)
1. Create `domain/commands/` files
2. Create `domain/logs/` files
3. Create `lib/api/graphql/queries/command-queries.ts`
4. Create `lib/api/graphql/queries/log-queries.ts`
5. Create GraphQL fragments

### Phase 3: Presentation Layer - Hooks (Day 2)
1. Create `hooks/commands/use-commands.ts`
2. Create `hooks/commands/use-command-history.ts`
3. Create `hooks/commands/use-pending-commands.ts`
4. Create `hooks/logs/use-logs.ts`
5. Create shared hooks

### Phase 4: UI Layer - Shared Components (Day 2-3)
1. Create shared components
2. Create commands components
3. Create logs components

### Phase 5: Dashboard Components (Day 3)
1. Create `DashboardPage`
2. Create tab components

### Phase 6: Routes (Day 3-4)
1. Create `/commands` routes
2. Create `/dashboard/*` routes
3. Create `/logs` route
4. Update existing routes

### Phase 7: Polish (Day 4)
1. Time range selector
2. Export functionality
3. Loading states
4. Error handling

---

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

*Document Version: 1.2*  
*Status: Ready for Implementation*  
*Architecture: Layered (Following FRONTEND_ARCHITECTURE.md)*
