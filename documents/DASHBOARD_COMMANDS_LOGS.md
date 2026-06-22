# Dashboard, Commands & Logs - Enterprise Requirements Specification

> **Version:** 1.1  
> **Status:** Draft  
> **Created:** 2026-06-21  
> **Updated:** 2026-06-22  
> **Target:** Production MVP  

---

## Table of Contents

1. [Overview](#1-overview)
2. [Page Structure](#2-page-structure)
3. [Shared UI Components](#3-shared-ui-components)
4. [Export System](#4-export-system)
5. [Dashboard: Overview Tab](#5-dashboard-overview-tab)
6. [Dashboard: Metrics Tab](#6-dashboard-metrics-tab)
7. [Dashboard: Commands Tab](#7-dashboard-commands-tab)
8. [Dashboard: Logs Tab](#8-dashboard-logs-tab)
9. [Commands Page](#9-commands-page)
10. [Logs Page](#10-logs-page)
11. [Device Page Updates](#11-device-page-updates)
12. [REST API Specification](#12-rest-api-specification)
13. [GraphQL Schema](#13-graphql-schema)
14. [Frontend Components](#14-frontend-components)
15. [File Changes Summary](#15-file-changes-summary)
16. [Implementation Order](#16-implementation-order)
17. [Preset Commands Reference](#17-preset-commands-reference)

---

## 1. Overview

### 1.1 Purpose

Redesign the Dashboard page with tabs for better organization, create a shared Commands page accessible from both Dashboard and Device, and maintain the Logs page.

### 1.2 Key Principles

- **Commands are shared** - Accessible from Dashboard tabs AND Device page
- **Logs are standalone** - Separate `/logs` route, accessible from Dashboard tabs
- **No duplication** - Same component used in multiple places via routing
- **Device context** - All pages use current device from config

### 1.3 Design Aesthetic

- Command center feel (dense, information-rich)
- Custom sections with borders (not floating cards)
- Rose-500 for accents/highlights only
- Minimal shadows
- Monospace for technical data

---

## 2. Page Structure

### 2.1 Sidebar Navigation

```
┌─────────────────────────────────────────────────────────────────────┐
│  VYZORIX                                                          │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ■ Dashboard                                                       │
│    └─ [Overview] [Metrics] [Commands] [Logs]                       │
│                                                                     │
│  ■ Device                                                          │
│    └─ [Inbox] [Overview] [Telemetry] [History]                   │
│                                                                     │
│  ■ Updates                                                         │
│                                                                     │
│  ■ Diagnostics                                                     │
│    └─ [Inspector] [Timeline]                                       │
│                                                                     │
│  ■ Alerts                                                          │
│    └─ [Active] [Status] [History]                                  │
│                                                                     │
│  ■ Settings                                                        │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 Route Structure

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

### 2.3 Shared Components

| Component | Used By | Purpose |
|-----------|---------|---------|
| `CommandsSend` | All command pages | Send commands grid |
| `CommandsPending` | All pending pages | Pending queue |
| `CommandsHistory` | All history pages | Full history |
| `CommandsRecent` | Send page | Recent commands |
| `LogsStream` | Dashboard/Logs tab, /logs page | Event log display |

---

## 3. Shared UI Components

### 3.1 Design Principles

- **NO CODE DUPLICATION** - All UI components defined once in `components/ui/`
- **Composable** - Small components composed into larger ones
- **Themable** - Uses existing design tokens (--primary, --foreground, etc.)
- **Accessible** - Proper ARIA labels, keyboard navigation

### 3.2 Core UI Components

All components live in `src/components/ui/`:

#### 3.2.1 Button Variants

```typescript
// Button sizes
"sm"   // Small button (h-8 px-3 text-xs)
"md"   // Medium button (h-10 px-4 text-sm) [DEFAULT]
"lg"   // Large button (h-12 px-6 text-base)

// Button variants
"default"    // Primary rose-500 bg
"outline"    // Border only, transparent bg
"ghost"      // No border, transparent bg
"destructive" // Red bg for danger actions

// Button with icon
Button with `leftIcon` or `rightIcon` prop (lucide-react icons)
```

#### 3.2.2 Badge Component

```typescript
Badge variants:
"default"      // Gray bg
"success"      // Green (for: online, delivered, success)
"warning"      // Yellow (for: pending, warning)
"destructive"  // Red (for: error, failed, critical)
"outline"      // Border only
```

#### 3.2.3 Card/Section Component

```typescript
// Replaces floating Card with bordered Section
<Section>
  <Section.Header title="Section Title" subtitle="Optional subtitle" />
  <Section.Content>
    {/* content */}
  </Section.Content>
</Section>
```

#### 3.2.4 Table Components

```typescript
<Table>
  <Table.Header>
    <Table.Row>
      <Table.Head>Column 1</Table.Head>
      <Table.Head>Column 2</Table.Head>
    </Table.Row>
  </Table.Header>
  <Table.Body>
    {items.map(item => (
      <Table.Row key={item.id}>
        <Table.Cell>{item.name}</Table.Cell>
        <Table.Cell>{item.status}</Table.Cell>
      </Table.Row>
    ))}
  </Table.Body>
</Table>
```

#### 3.2.5 Dropdown Menu

```typescript
DropdownMenu
├── DropdownMenu.Trigger (button)
├── DropdownMenu.Content
│   ├── DropdownMenu.Item (onClick handler)
│   ├── DropdownMenu.Separator
│   └── DropdownMenu.Label
└── DropdownMenu
```

#### 3.2.6 Tabs

```typescript
<Tabs defaultValue="overview">
  <Tabs.List>
    <Tabs.Trigger value="overview">Overview</Tabs.Trigger>
    <Tabs.Trigger value="metrics">Metrics</Tabs.Trigger>
  </Tabs.List>
  <Tabs.Content value="overview">
    {/* content */}
  </Tabs.Content>
  <Tabs.Content value="metrics">
    {/* content */}
  </Tabs.Content>
</Tabs>
```

#### 3.2.7 Search Input

```typescript
<SearchInput
  placeholder="Search commands..."
  value={search}
  onChange={setSearch}
  onClear={() => setSearch("")}
/>
```

#### 3.2.8 Select/Dropdown

```typescript
<Select value={value} onValueChange={setValue}>
  <Select.Trigger>
    <Select.Value placeholder="Select..." />
  </Select.Trigger>
  <Select.Content>
    <Select.Item value="all">All</Select.Item>
    <Select.Item value="pending">Pending</Select.Item>
    <Select.Item value="delivered">Delivered</Select.Item>
  </Select.Content>
</Select>
```

#### 3.2.9 Pagination

```typescript
<Pagination
  page={currentPage}
  totalPages={totalPages}
  onPageChange={setPage}
/>
```

#### 3.2.10 Skeleton/Loading

```typescript
<Skeleton variant="text" width="100%" />
<Skeleton variant="rect" width="100%" height={200} />
<Skeleton variant="circle" width={40} height={40} />
```

#### 3.2.11 Empty State

```typescript
<EmptyState
  icon={InboxIcon}
  title="No pending commands"
  description="All commands have been delivered"
  action={{
    label: "Send Command",
    onClick: () => navigate("/commands")
  }}
/>
```

#### 3.2.12 Tooltip

```typescript
<Tooltip content="Cancel this command">
  <Button variant="ghost" size="sm">Cancel</Button>
</Tooltip>
```

#### 3.2.13 Toast Notifications

```typescript
// Use existing sonner integration
toast.success("Command sent", { description: "dispatch abc123" });
toast.error("Command failed", { description: error.message });
toast.warning("Buffer above 80%");
```

### 3.3 Shared Page Layout Components

#### 3.3.1 PageHeader

```typescript
<PageHeader
  title="Commands"
  subtitle="Send commands to your device"
  actions={[
    { label: "Export", onClick: handleExport, icon: Download }
  ]}
/>
```

#### 3.3.2 PageTabs

```typescript
<PageTabs
  tabs={[
    { value: "send", label: "Send", href: "/commands" },
    { value: "pending", label: "Pending (2)", href: "/commands/pending" },
    { value: "history", label: "History", href: "/commands/history" }
  ]}
/>
```

#### 3.3.3 DeviceSelector

```typescript
// Shared device dropdown used across all pages
<DeviceSelector
  value={selectedImei}
  onChange={setSelectedImei}
  devices={devices}
/>
```

### 3.4 File Structure

```
src/components/ui/
├── button.tsx
├── badge.tsx
├── section.tsx        # Bordered panel (replaces Card)
├── table.tsx
├── dropdown-menu.tsx
├── tabs.tsx
├── search-input.tsx
├── select.tsx
├── pagination.tsx
├── skeleton.tsx
├── empty-state.tsx
├── tooltip.tsx
├── page-header.tsx
├── page-tabs.tsx
└── device-selector.tsx

src/components/shared/
├── device-selector.tsx
├── connection-status.tsx
├── metric-card.tsx
└── export-menu.tsx
```

---

## 4. Export System

### 4.1 Overview

The export system provides comprehensive data export capabilities across Metrics, Commands, and Logs pages. All exports use shared components to avoid duplication.

### 4.2 Export Formats

#### 4.2.1 CSV Format

**Structure:**
- Header row with column names
- Data rows with values
- UTF-8 encoding with BOM for Excel compatibility
- Comma-separated values
- Quoted strings where necessary

**Filename pattern:**
```
{type}_{deviceName}_{date}_{time}.csv
Example: telemetry_pixel8_2026-06-22_143052.csv
```

#### 4.2.2 JSON Format

**Structure:**
```json
{
  "exportedAt": "2026-06-22T14:30:52Z",
  "device": {
    "imei": "861234567890123",
    "name": "Pixel 8 Pro"
  },
  "filters": {
    "startTime": 1718900000000,
    "endTime": 1718986400000,
    "type": "all"
  },
  "data": [
    { /* record */ }
  ],
  "metadata": {
    "totalRecords": 1234,
    "format": "csv",
    "version": "1.0"
  }
}
```

**Filename pattern:**
```
{type}_{deviceName}_{date}_{time}.json
Example: telemetry_pixel8_2026-06-22_143052.json
```

#### 4.2.3 XML Format (Optional)

**Structure:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<export>
  <metadata>
    <exportedAt>2026-06-22T14:30:52Z</exportedAt>
    <device>
      <imei>861234567890123</imei>
      <name>Pixel 8 Pro</name>
    </device>
    <totalRecords>1234</totalRecords>
  </metadata>
  <data>
    <record>
      <!-- fields -->
    </record>
  </data>
</export>
```

### 4.3 Export Menu Component

#### 4.3.1 UI Design

```
┌─────────────────────────────────┐
│  ┌───────────────────────────┐  │
│  │  Export                   │  │
│  └───────────────────────────┘  │
└─────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────┐
│  ┌───────────────────────────┐  │
│  │  CSV (.csv)           ↓  │  │
│  │  JSON (.json)            │  │
│  │  XML (.xml)              │  │ (Optional)
│  └───────────────────────────┘  │
│                                  │
│  ─────────────────────────────── │
│                                  │
│  ┌───────────────────────────┐  │
│  │  Current View             │  │  ← Export visible data only
│  │  All Time Range           │  │  ← Export full filtered data
│  └───────────────────────────┘  │
└─────────────────────────────────┘
```

#### 4.3.2 ExportMenu Props

```typescript
interface ExportMenuProps {
  data: ExportableData[];
  filename: string;
  formats: ExportFormat[];
  onExport: (format: ExportFormat, scope: ExportScope) => void;
}

type ExportFormat = "csv" | "json" | "xml";
type ExportScope = "current" | "all";
```

#### 4.3.3 ExportScope Options

| Scope | Description |
|-------|-------------|
| `current` | Export only the data currently displayed (respects pagination) |
| `all` | Export all data matching current filters (requires confirmation for large datasets) |

### 4.4 Large Dataset Handling

#### 4.4.1 Client-Side Export (Small datasets < 10,000 rows)

- Generate export in browser
- Show progress indicator
- Use Blob API for download

#### 4.4.2 Server-Side Export (Large datasets > 10,000 rows)

When user requests "All" scope:
1. Show confirmation dialog:
   ```
   ┌─────────────────────────────────────────────────┐
   │  Large Export Detected                         │
   │                                                 │
   │  This export will include 45,231 records.     │
   │  This may take a few minutes.                 │
   │                                                 │
   │  ┌─────────────────────┐  ┌─────────────────┐  │
   │  │  Export Anyway     │  │     Cancel      │  │
   │  └─────────────────────┘  └─────────────────┘  │
   └─────────────────────────────────────────────────┘
   ```

2. Server streams export:
   - Returns 202 Accepted with job ID
   - Client polls for status
   - When complete, provides download link

#### 4.4.3 Export Job API

```typescript
// POST /v1/export/telemetry
Request:
{
  "type": "telemetry" | "commands" | "logs",
  "imei": "861234567890123",
  "format": "csv" | "json",
  "filters": {
    "startTime": 1718900000000,
    "endTime": 1718986400000,
    "type": "all"
  }
}

Response:
{
  "jobId": "export_abc123",
  "status": "processing",
  "estimatedRecords": 45231
}

// GET /v1/export/:jobId
Response:
{
  "jobId": "export_abc123",
  "status": "complete" | "processing" | "failed",
  "progress": 75,  // percentage
  "downloadUrl": "https://..."  // when complete
}
```

### 4.5 Export Types by Page

#### 4.5.1 Metrics Export

**CSV Columns:**
```csv
timestamp,riskScore,thermalTemp,bufferLevel,uptime
1718900000000,45,38.5,67,86400
1718900100000,46,38.6,68,86410
```

**JSON Structure:**
```json
{
  "exportedAt": "...",
  "device": {...},
  "filters": {...},
  "data": [
    {
      "timestamp": 1718900000000,
      "riskScore": 45,
      "thermalTemp": 38.5,
      "bufferLevel": 67,
      "uptime": 86400
    }
  ]
}
```

#### 4.5.2 Commands Export

**CSV Columns:**
```csv
dispatchId,command,status,sentAt,deliveredAt,latencyMs
abc123def456,GET_STATUS,delivered,1718900000000,1718900000234,234
```

**JSON Structure:**
```json
{
  "data": [
    {
      "dispatchId": "abc123def456",
      "command": "GET_STATUS",
      "status": "delivered",
      "sentAt": 1718900000000,
      "deliveredAt": 1718900000234,
      "latencyMs": 234
    }
  ]
}
```

#### 4.5.3 Logs Export

**CSV Columns:**
```csv
timestamp,type,data
1718900000000,TELEMETRY,"{""riskScore"":45,""thermalTemp"":38.5}"
```

**JSON Structure:**
```json
{
  "data": [
    {
      "id": "uuid",
      "timestamp": 1718900000000,
      "type": "TELEMETRY",
      "data": {
        "riskScore": 45,
        "thermalTemp": 38.5
      }
    }
  ]
}
```

### 4.6 Export Hook

```typescript
// useExport.ts
export function useExport() {
  const exportData = async (
    data: ExportableData[],
    filename: string,
    format: ExportFormat,
    scope: ExportScope
  ) => {
    // Check if server-side export needed
    if (scope === "all" && data.length > 10000) {
      // Start server export job
      return startServerExport(data, filename, format);
    }
    
    // Client-side export
    return generateClientExport(data, filename, format);
  };
  
  const generateClientExport = (data, filename, format) => {
    switch (format) {
      case "csv":
        return generateCSV(data, filename);
      case "json":
        return generateJSON(data, filename);
      case "xml":
        return generateXML(data, filename);
    }
  };
}
```

### 4.7 Export UI States

| State | UI |
|-------|-----|
| Ready | "Export" button enabled |
| Exporting | Spinner + "Exporting..." |
| Success | Toast + Download triggered |
| Error | Toast with error message |
| Large Export | Confirmation dialog |

---

## 5. Dashboard: Overview Tab

### 5.1 Purpose

At-a-glance status of current device with key metrics and quick actions.

### 5.2 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  DASHBOARD                                    ● Connected | 2s ago  │
├─────────────────────────────────────────────────────────────────────┤
│  [Overview] [Metrics] [Commands] [Logs]                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─ CONNECTION ────────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  ● Pixel 8 Pro                         [View Device ▾]   │   │
│  │  IMEI: 861234567890123                                     │   │
│  │  WS: Connected · FCM: Valid · Last: 2s ago              │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ METRICS ─────────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  RISK          THERMAL        UPTIME         BUFFER      │   │
│  │  ████████░ 72  █████░ 45°C  ████ 4d     ████████░ 67% │   │
│  │  Healthy        Normal        Running        Stable         │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ QUICK ACTIONS ───────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  [Send Command ▾]  [Refresh]  [View Logs]  [Alerts: 2] │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ DEVICE INFO ────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  OS: Android 14    App: v2.1.0    Build: UP1A.231005.007 │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 5.3 Sections

| Section | Content |
|---------|---------|
| **Connection** | Device name, IMEI, connection status, last seen |
| **Metrics** | Risk, Thermal, Uptime, Buffer (with progress bars) |
| **Quick Actions** | Send Command, Refresh, View Logs, Alerts count |
| **Device Info** | OS, App version, Build ID |

### 5.4 Interactions

| Element | Action |
|---------|--------|
| "View Device ▾" | Dropdown to switch devices |
| Metrics | Click to navigate to Metrics tab |
| "Send Command" | Dropdown with preset commands |
| "View Logs" | Navigate to Logs tab |
| "Alerts: 2" | Badge, click to Alerts page |

---

## 6. Dashboard: Metrics Tab

### 6.1 Purpose

Deep dive into telemetry data with time range selection and export options.

### 6.2 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  METRICS                                    [1h] [6h] [24h] [7d]   │
│                                                             [Export ▾]│
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─ RISK SCORE ─────────────────────────────────────────────┐   │
│  │         Current: 72  Avg: 45  Min: 32  Max: 78        │   │
│  │  ┌─────────────────────────────────────────────────┐   │   │
│  │  │                                                 │   │   │
│  │  │    100 ─┤                                    ╱╲  │   │   │
│  │  │     50 ─┤                       ╱╲       ╱    ╲ │   │   │
│  │  │      0 ─┴────────────────╯──╲──╱──────╱──────╲───│   │   │
│  │  │                       12:00  12:30  13:00  13:30 │   │   │
│  │  └─────────────────────────────────────────────────┘   │   │
│  │  Thresholds: ─ ─ ─ Warning (50)  ┄┄ Critical (70)     │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ THERMAL ───────────────────────────────────────────────┐   │
│  │         Current: 45°C  Avg: 42°C  Min: 38°C  Max: 52°C│   │
│  │  [Similar chart]                                     │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ BUFFER LEVEL ─────────────────────────────────────────┐   │
│  │         Current: 67%  Avg: 55%  Min: 12%  Max: 89%    │   │
│  │  [Similar chart]                                     │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 6.3 Controls

| Control | Type | Description |
|---------|------|-------------|
| Time Range | Buttons | 1h, 6h, 24h, 7d (default: 6h) |
| Export | Dropdown | CSV, JSON |
| Chart | Line chart | With threshold lines |

### 6.4 Data Points

| Metric | Stats Shown |
|--------|------------|
| Risk Score | Current, Avg, Min, Max |
| Thermal | Current, Avg, Min, Max |
| Buffer Level | Current, Avg, Min, Max |

### 6.5 Export

Uses shared `ExportMenu` component from Section 4.

---

## 7. Dashboard: Commands Tab

### 7.1 Purpose

Quick access to command sending (redirects to `/commands` page).

### 7.2 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  COMMANDS                                     [Pending] [Recent]   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─ SEND COMMAND ────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐       │   │
│  │  │ GET_STATUS  │ │ REBOOT       │ │ CLEAR_BUFFER │       │   │
│  │  │              │ │              │ │              │       │   │
│  │  │ Get device  │ │ Restart      │ │ Clear audio  │       │   │
│  │  │ status      │ │ device       │ │ buffer       │       │   │
│  │  └──────────────┘ └──────────────┘ └──────────────┘       │   │
│  │                                                              │   │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐       │   │
│  │  │ WAKE_UPDATER│ │ SOFT_REBOOT  │ │ STOP_DAEMON │       │   │
│  │  │              │ │              │ │              │       │   │
│  │  │ Wake audio   │ │ Soft restart │ │ Stop audio  │       │   │
│  │  │ updater      │ │ (app only)   │ │ service     │       │   │
│  │  └──────────────┘ └──────────────┘ └──────────────┘       │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ PENDING (2) ───────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  WAKE_UPDATER   abc123def456   Pending   [Cancel]        │   │
│  │  REBOOT         ghi789jkl012   Pending   [Cancel]        │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ RECENT ─────────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  GET_STATUS       mno345pqr678   ✓ Delivered   12:34    │   │
│  │  CLEAR_BUFFER     stu901vwx234   ✓ Delivered   12:30    │   │
│  │  GET_STATUS       yz901234abc    ✓ Delivered   12:00    │   │
│  │                                                              │   │
│  │                                    [View Full History →]  │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 7.3 Preset Commands

Based on APK capabilities (no args):

| Command | Label | Description | Danger |
|---------|-------|-------------|--------|
| `GET_STATUS` | Get Status | Request device status | No |
| `REBOOT` | Reboot | Full device restart | Yes |
| `CLEAR_BUFFER` | Clear Buffer | Clear audio capture buffer | No |
| `WAKE_UPDATER` | Wake Updater | Wake audio updater service | No |
| `SOFT_REBOOT` | Soft Reboot | Restart app only | Yes |
| `STOP_DAEMON` | Stop Daemon | Stop audio service | Yes |

### 7.4 Sections

| Section | Content |
|---------|---------|
| **Send Command** | Grid of preset command buttons |
| **Pending** | Commands awaiting ACK (with Cancel) |
| **Recent** | Last 3-5 commands with status |

### 7.5 Shared Components

Uses shared components from Section 3:
- `CommandsSend` - Send commands grid
- `CommandsPending` - Pending queue
- `CommandsRecent` - Recent commands

---

## 8. Dashboard: Logs Tab

### 8.1 Purpose

Real-time WebSocket event stream for debugging.

### 8.2 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  LOGS                              [All ▾] [Auto-scroll ✓] [Clear]│
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─ EVENT STREAM ─────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  12:34:56.123  ● CONNECTED     WebSocket established        │   │
│  │  12:35:02.456  ● TELEMETRY    Risk: 72, Thermal: 45°C      │   │
│  │  12:35:05.789  ○ COMMAND      Sent: WAKE_UPDATER           │   │
│  │  12:35:06.012  ● ACK          Delivered in 234ms            │   │
│  │  12:35:12.345  ● WARNING      Buffer level above 80%        │   │
│  │  12:35:20.678  ● ERROR        Command timeout: GET_STATUS  │   │
│  │  12:35:22.901  ● RECONNECT   Attempt 2/5 - connection lost │   │
│  │  12:35:24.123  ● CONNECTED   Reconnected successfully      │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 8.3 Event Types

| Type | Icon | Color |
|------|------|-------|
| `CONNECTED` | ● | Rose (primary) |
| `DISCONNECTED` | ○ | Gray |
| `RECONNECT` | ● | Rose (primary) |
| `TELEMETRY` | ● | Rose (primary) |
| `COMMAND` | ○ | Gray |
| `ACK` | ● | Rose (primary) |
| `ERROR` | ● | Rose (destructive) |
| `WARNING` | ● | Rose (warning) |
| `INFO` | ● | Gray |

### 8.4 Controls

| Control | Type | Description |
|---------|------|-------------|
| Filter | Dropdown | All, Connection, Commands, Telemetry, Errors |
| Auto-scroll | Toggle | Auto-scroll to newest |
| Clear | Button | Clear log buffer |
| Export | ExportMenu | Export to CSV/JSON |

### 8.5 Shared Components

Uses shared components from Section 3:
- `LogsStream` - Event log display
- `LogEntry` - Single log row
- `ExportMenu` - Export dropdown

---

## 9. Commands Page

### 9.1 Purpose

Standalone commands page, accessible via:
- `/dashboard/commands` → Send commands
- `/dashboard/commands/pending` → Pending queue
- `/dashboard/commands/history` → Command history
- `/device/:imei/commands` → Send commands (device-specific)
- `/device/:imei/commands/pending` → Pending queue
- `/device/:imei/commands/history` → Command history
- `/commands` → Send commands (standalone)
- `/commands/pending` → Pending queue
- `/commands/history` → Command history

### 7.2 Routes

| Route | Purpose | Component |
|-------|---------|-----------|
| `/commands` | Send commands | `CommandsSend` |
| `/commands/pending` | Pending queue | `CommandsPending` |
| `/commands/history` | Full history | `CommandsHistory` |

### 7.3 Commands Send Page

```
┌─────────────────────────────────────────────────────────────────────┐
│  COMMANDS                                        Device: Pixel 8 ▾│
├─────────────────────────────────────────────────────────────────────┤
│  [Send] [Pending (2)] [History]                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─ SEND COMMAND ────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  Select a command to send:                                 │   │
│  │                                                              │   │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐       │   │
│  │  │ GET_STATUS  │ │ REBOOT       │ │ CLEAR_BUFFER │       │   │
│  │  │              │ │              │ │              │       │   │
│  │  │ Get device  │ │ Restart      │ │ Clear audio  │       │   │
│  │  │ status      │ │ device       │ │ buffer       │       │   │
│  │  └──────────────┘ └──────────────┘ └──────────────┘       │   │
│  │                                                              │   │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐       │   │
│  │  │ WAKE_UPDATER│ │ SOFT_REBOOT  │ │ STOP_DAEMON │       │   │
│  │  │              │ │              │ │              │       │   │
│  │  │ Wake audio   │ │ Soft restart │ │ Stop audio  │       │   │
│  │  │ updater      │ │ (app only)   │ │ service     │       │   │
│  │  └──────────────┘ └──────────────┘ └──────────────┘       │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ RECENT ─────────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  GET_STATUS       mno345pqr678   ✓ Delivered   12:34    │   │
│  │  CLEAR_BUFFER     stu901vwx234   ✓ Delivered   12:30    │   │
│  │  GET_STATUS       yz901234abc    ✓ Delivered   12:00    │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 7.4 Commands Pending Page

```
┌─────────────────────────────────────────────────────────────────────┐
│  COMMANDS                                        Device: Pixel 8 ▾│
├─────────────────────────────────────────────────────────────────────┤
│  [Send] [Pending (2)] [History]                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─ PENDING (2) ───────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  ┌─────────────────────────────────────────────────┐   │   │
│  │  │ WAKE_UPDATER   abc123def456   Pending   [Cancel] │   │   │
│  │  │ Sent: 12:34:56 · ETA: ~5s                        │   │   │
│  │  └─────────────────────────────────────────────────┘   │   │
│  │                                                              │   │
│  │  ┌─────────────────────────────────────────────────┐   │   │
│  │  │ REBOOT         ghi789jkl012   Pending   [Cancel] │   │   │
│  │  │ Sent: 12:34:50 · ETA: ~10s                       │   │   │
│  │  └─────────────────────────────────────────────────┘   │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ RECENT ─────────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  GET_STATUS       mno345pqr678   ✓ Delivered   12:34    │   │
│  │  CLEAR_BUFFER     stu901vwx234   ✓ Delivered   12:30    │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 7.5 Commands History Page

```
┌─────────────────────────────────────────────────────────────────────┐
│  COMMANDS                                        Device: Pixel 8 ▾│
├─────────────────────────────────────────────────────────────────────┤
│  [Send] [Pending (2)] [History]                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─ HISTORY ─────────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  Search: [________________]  Filter: [All ▾]  [Export ▾] │   │
│  │                                                              │   │
│  │  ┌─────────────────────────────────────────────────┐   │   │
│  │  │ GET_STATUS       mno345pqr678   ✓ Delivered    │   │   │
│  │  │ Sent: 12:34:00 · Latency: 234ms                │   │   │
│  │  └─────────────────────────────────────────────────┘   │   │
│  │  ┌─────────────────────────────────────────────────┐   │   │
│  │  │ CLEAR_BUFFER     stu901vwx234   ✓ Delivered    │   │   │
│  │  │ Sent: 12:30:00 · Latency: 189ms                │   │   │
│  │  └─────────────────────────────────────────────────┘   │   │
│  │  ┌─────────────────────────────────────────────────┐   │   │
│  │  │ GET_STATUS       yz901234abc    ✗ Failed       │   │   │
│  │  │ Sent: 12:15:00 · Reason: Timeout              │   │   │
│  │  └─────────────────────────────────────────────────┘   │   │
│  │  ┌─────────────────────────────────────────────────┐   │   │
│  │  │ REBOOT           123456def789   ✓ Delivered    │   │   │
│  │  │ Sent: 12:00:00 · Latency: 567ms                │   │   │
│  │  └─────────────────────────────────────────────────┘   │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  [Load More]                                                       │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 7.6 Shared Components

| Component | Used By | Purpose |
|-----------|---------|---------|
| `CommandsSend` | All send routes | Send commands grid |
| `CommandsPending` | All pending routes | Pending queue |
| `CommandsHistory` | All history routes | Full history |
| `CommandsRecent` | Send page | Recent 3-5 commands |

---

## 10. Logs Page

### 16.1 Purpose

Standalone logs page, accessible via:
- `/dashboard/logs`
- `/logs`

### 16.2 Layout (Same as Dashboard/Logs tab)

```
┌─────────────────────────────────────────────────────────────────────┐
│  LOGS                                              [Export] [Clear]│
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Filter: [All ▾] [Connection] [Commands] [Telemetry] [Errors]    │
│                                                                     │
│  ┌─ EVENT STREAM ─────────────────────────────────────────────┐   │
│  │  [Full log content as shown in Dashboard/Logs]             │   │
│  │                                                              │   │
│  │  [Auto-scroll ✓]                                           │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 16.3 Additional Features

| Feature | Description |
|---------|-------------|
| Export | Download full log as JSON |
| Clear | Clear log buffer |
| Persistent | Logs persist across page refreshes (localStorage) |

### 16.4 Shared Components

Uses shared components from Section 3:
- `LogsStream` - Event log display
- `LogEntry` - Single log row
- `ExportMenu` - Export dropdown

---

## 11. Device Page Updates

### 15.1 Structure

```
Device Page (tabs):
├── Inbox        → Pending registration requests
├── Overview     → Device info, health, connection
├── Telemetry    → Real-time charts, metrics
├── History      → Historical data, export
└── Commands     → CommandsPanel (shared component)
```

### 9.2 Commands Tab in Device

Same `CommandsPanel` component, filtered to specific device by IMEI.

---

## 12. REST API Specification

### 16.1 Commands Endpoint

#### `POST /v1/device/:imei/command`
**Purpose:** Dispatch command to device

**Request:**
```json
{
  "command": "GET_STATUS"
}
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
**Purpose:** Get command history

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | string | all | Filter: pending, delivered, failed |
| `page` | int | 1 | Page number |
| `limit` | int | 20 | Items per page |

**Response (200 OK):**
```json
{
  "commands": [
    {
      "dispatchId": "abc123def456",
      "command": "GET_STATUS",
      "status": "delivered",
      "sentAt": 1718900000000,
      "deliveredAt": 1718900000234,
      "latencyMs": 234
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "totalPages": 3
  }
}
```

---

#### `DELETE /v1/device/:imei/command/:dispatchId`
**Purpose:** Cancel pending command

**Response (200 OK):**
```json
{
  "dispatchId": "abc123def456",
  "status": "cancelled"
}
```

---

### 16.2 Logs Endpoint

#### `GET /v1/device/:imei/logs`
**Purpose:** Get historical log events

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
      "data": {
        "riskScore": 72,
        "thermalTemp": 45.2,
        "bufferLevel": 67
      }
    },
    {
      "id": "uuid-v4",
      "type": "COMMAND_SENT",
      "timestamp": 1718900500000,
      "data": {
        "command": "WAKE_UPDATER",
        "dispatchId": "abc123..."
      }
    }
  ],
  "pagination": {
    "limit": 100,
    "hasMore": true,
    "nextCursor": "base64-cursor"
  }
}
```

---

### 16.3 Telemetry Endpoint (for Metrics)

#### `GET /v1/device/:imei/telemetry`
**Purpose:** Get telemetry history for charts

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
    {
      "timestamp": 1718900000000,
      "riskScore": 45,
      "thermalTemp": 38.5,
      "bufferLevel": 67,
      "uptime": 86400
    }
  ],
  "stats": {
    "riskScore": { "current": 72, "avg": 45, "min": 32, "max": 78 },
    "thermalTemp": { "current": 45.2, "avg": 42, "min": 38, "max": 52 },
    "bufferLevel": { "current": 67, "avg": 55, "min": 12, "max": 89 }
  }
}
```

---

## 13. GraphQL Schema

### 15.1 Types

```graphql
enum CommandStatus {
  PENDING
  DELIVERED
  FAILED
  CANCELLED
}

enum LogEventType {
  CONNECTED
  DISCONNECTED
  RECONNECT
  TELEMETRY
  COMMAND_SENT
  COMMAND_ACK
  COMMAND_FAILED
  ERROR
  WARNING
  INFO
}

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

type TelemetryFrame {
  timestamp: DateTime!
  riskScore: Int
  thermalTemp: Float
  bufferLevel: Int
  uptime: Int
}

type MetricStats {
  current: Float!
  avg: Float!
  min: Float!
  max: Float!
}

type TelemetryStats {
  riskScore: MetricStats!
  thermalTemp: MetricStats!
  bufferLevel: MetricStats!
}

type PaginationInfo {
  page: Int!
  limit: Int!
  total: Int!
  totalPages: Int!
}
```

### 15.2 Queries

```graphql
type Query {
  # Commands
  commands(
    imei: String!
    status: CommandStatus
    page: Int = 1
    limit: Int = 20
  ): CommandConnection!

  # Logs
  logs(
    imei: String!
    type: LogEventType
    startTime: Int
    endTime: Int
    limit: Int = 100
    cursor: String
  ): LogConnection!

  # Telemetry
  telemetry(
    imei: String!
    startTime: Int
    endTime: Int
    limit: Int = 500
  ): TelemetryResult!
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

type TelemetryResult {
  frames: [TelemetryFrame!]!
  stats: TelemetryStats!
}
```

### 15.3 Mutations

```graphql
type Mutation {
  sendCommand(
    imei: String!
    command: String!
  ): SendCommandResponse!

  cancelCommand(
    imei: String!
    dispatchId: ID!
  ): CancelCommandResponse!
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

## 14. Frontend Components

### 16.1 Dashboard Components

| Component | File | Purpose |
|-----------|------|---------|
| DashboardPage | `routes/dashboard.tsx` | Page wrapper with tabs |
| DashboardOverview | `components/dashboard/DashboardOverview.tsx` | Overview tab content |
| DashboardMetrics | `components/dashboard/DashboardMetrics.tsx` | Metrics tab content |
| DashboardCommands | `components/dashboard/DashboardCommands.tsx` | Commands tab (redirects) |
| DashboardLogs | `components/dashboard/DashboardLogs.tsx` | Logs tab (redirects) |

### 16.2 Commands Page Components

| Component | File | Purpose |
|-----------|------|---------|
| CommandsPage | `routes/commands.tsx` | Main commands page |
| CommandsSend | `components/commands/CommandsSend.tsx` | Send commands grid |
| CommandsPending | `components/commands/CommandsPending.tsx` | Pending queue |
| CommandsHistory | `components/commands/CommandsHistory.tsx` | Full history |
| CommandsRecent | `components/commands/CommandsRecent.tsx` | Recent commands |
| CommandButton | `components/commands/CommandButton.tsx` | Single command button |

### 16.3 Logs Page Components

| Component | File | Purpose |
|-----------|------|---------|
| LogsPage | `routes/logs.tsx` | Standalone logs page |
| LogsStream | `components/logs/LogsStream.tsx` | Event log display |
| LogEntry | `components/logs/LogEntry.tsx` | Single log row |

### 16.4 Hooks

| Hook | File | Purpose |
|------|------|---------|
| `useCommands` | `hooks/use-commands.ts` | Command queries/mutations |
| `useCommandHistory` | `hooks/use-command-history.ts` | History queries |
| `useLogs` | `hooks/use-logs.ts` | Log queries |
| `useTelemetry` | `hooks/use-telemetry.ts` | Telemetry for charts |

---

## 15. File Changes Summary

### 15.1 NEW Files (Frontend - Routes)

| File | Purpose |
|------|---------|
| `src/routes/commands.tsx` | /commands page |
| `src/routes/commands.pending.tsx` | /commands/pending |
| `src/routes/commands.history.tsx` | /commands/history |
| `src/routes/logs.tsx` | /logs page |
| `src/routes/dashboard.metrics.tsx` | /dashboard/metrics |
| `src/routes/dashboard.commands.tsx` | /dashboard/commands |
| `src/routes/dashboard.commands.pending.tsx` | /dashboard/commands/pending |
| `src/routes/dashboard.commands.history.tsx` | /dashboard/commands/history |
| `src/routes/dashboard.logs.tsx` | /dashboard/logs |
| `src/routes/device.$imei.commands.tsx` | /device/:imei/commands |
| `src/routes/device.$imei.commands.pending.tsx` | /device/:imei/commands/pending |
| `src/routes/device.$imei.commands.history.tsx` | /device/:imei/commands/history |

### 15.2 NEW Files (Frontend - Components)

| File | Purpose |
|------|---------|
| `src/components/dashboard/DashboardOverview.tsx` | Overview tab content |
| `src/components/dashboard/DashboardMetrics.tsx` | Metrics tab content |
| `src/components/dashboard/DashboardCommands.tsx` | Commands tab wrapper |
| `src/components/dashboard/DashboardLogs.tsx` | Logs tab wrapper |
| `src/components/commands/CommandsSend.tsx` | Send commands grid |
| `src/components/commands/CommandsPending.tsx` | Pending queue |
| `src/components/commands/CommandsRecent.tsx` | Recent commands |
| `src/components/commands/CommandsHistory.tsx` | Full history |
| `src/components/commands/CommandButton.tsx` | Single command button |
| `src/components/logs/LogsStream.tsx` | Event log display |
| `src/components/logs/LogEntry.tsx` | Single log row |

### 15.3 NEW Files (Frontend - Hooks)

| File | Purpose |
|------|---------|
| `src/hooks/use-commands.ts` | Send/cancel commands |
| `src/hooks/use-command-history.ts` | Command history queries |
| `src/hooks/use-logs.ts` | Log queries |
| `src/hooks/use-telemetry.ts` | Telemetry for charts |

### 15.4 MODIFIED Files (Frontend)

| File | Changes |
|------|---------|
| `src/routes/dashboard.tsx` | Update with tabs (redirect to sub-pages) |
| `src/routes/device.tsx` | Update with tabs (add Commands) |
| `src/routes/device.$imei.tsx` | Add Commands tab routes |
| `src/routes/device.$imei.telemetry.tsx` | Telemetry tab |
| `src/routes/device.$imei.history.tsx` | History tab |
| `src/lib/api/graphql/queries.ts` | Add commands/logs/telemetry queries |
| `src/lib/api/graphql/mutations.ts` | Add command mutations |
| `src/lib/api/graphql/types.ts` | Add types |
| `src/lib/api/rest/device-client.ts` | Add REST fallback |

### 15.5 API Changes (Go Backend)

| File | Changes |
|------|---------|
| `internal/api/handlers/device/commands.go` | NEW - commands endpoint |
| `internal/api/handlers/device/logs.go` | NEW - logs endpoint |
| `internal/api/handlers/device/telemetry.go` | NEW - telemetry endpoint |
| `internal/api/router.go` | Add routes |

---

## 16. Implementation Order

### Phase 1: Backend APIs (Day 1)
1. Implement `GET /v1/device/:imei/commands`
2. Implement `GET /v1/device/:imei/logs`
3. Implement `GET /v1/device/:imei/telemetry`
4. Update GraphQL resolvers

### Phase 2: Shared Components & Hooks (Day 1-2)
1. Create `useCommands` hook
2. Create `useCommandHistory` hook
3. Create `useLogs` hook
4. Create `useTelemetry` hook
5. Create `CommandButton` component
6. Create `CommandsSend` component
7. Create `CommandsPending` component
8. Create `CommandsRecent` component
9. Create `CommandsHistory` component
10. Create `LogEntry` component
11. Create `LogsStream` component

### Phase 3: Commands Page Routes (Day 2)
1. Create `/commands` route → CommandsSend
2. Create `/commands/pending` route → CommandsPending
3. Create `/commands/history` route → CommandsHistory
4. Create `/dashboard/commands` route → redirects to /commands
5. Create `/dashboard/commands/pending` route → redirects to /commands/pending
6. Create `/dashboard/commands/history` route → redirects to /commands/history

### Phase 4: Logs Page Route (Day 2)
1. Create `/logs` route → LogsStream
2. Create `/dashboard/logs` route → redirects to /logs

### Phase 5: Dashboard Tabs (Day 2-3)
1. Refactor `dashboard.tsx` with tabs
2. Create `DashboardOverview`
3. Create `DashboardMetrics`
4. Create `DashboardCommands` (redirects to /commands)
5. Create `DashboardLogs` (redirects to /logs)
6. Create `/dashboard/metrics` route → DashboardMetrics

### Phase 6: Device Page Routes (Day 3)
1. Create `/device/:imei/telemetry` route
2. Create `/device/:imei/history` route
3. Create `/device/:imei/commands` route
4. Create `/device/:imei/commands/pending` route
5. Create `/device/:imei/commands/history` route

### Phase 7: Polish (Day 3-4)
1. Time range selector
2. Export functionality (CSV/JSON)
3. Loading states
4. Error handling
5. Animations
6. Device selector in Commands page

---

## 17. Preset Commands Reference

From APK software capabilities:

| ID | Label | Description | Danger |
|----|-------|-------------|--------|
| `GET_STATUS` | Get Status | Request current device status | No |
| `REBOOT` | Reboot | Full device restart | Yes |
| `CLEAR_BUFFER` | Clear Buffer | Clear audio capture buffer | No |
| `WAKE_UPDATER` | Wake Updater | Wake audio updater service | No |
| `SOFT_REBOOT` | Soft Reboot | Restart app without rebooting device | Yes |
| `STOP_DAEMON` | Stop Daemon | Stop audio service completely | Yes |

---

*Document Version: 1.0*  
*Status: Ready for Implementation*
