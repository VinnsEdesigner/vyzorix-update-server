# Diagnostics Page - Enterprise Requirements Specification

> **Version:** 1.0  
> **Status:** Draft  
> **Created:** 2026-06-21  
> **Target:** Production MVP  

---

## Table of Contents

1. [Overview](#1-overview)
2. [Page Structure](#2-page-structure)
3. [Tab 1: Inspector](#3-tab-1-inspector)
4. [Tab 2: Timeline](#4-tab-2-timeline)
5. [REST API Specification](#5-rest-api-specification)
6. [GraphQL Schema](#6-graphql-schema)
7. [Frontend Components](#7-frontend-components)
8. [File Changes Summary](#8-file-changes-summary)
9. [Implementation Order](#9-implementation-order)

---

## 1. Overview

### 1.1 Purpose

The Diagnostics page provides operators with deep visibility into:
1. **Current State** - What the server knows about the device right now
2. **Audit Trail** - What events have occurred over time

### 1.2 Design Principles

- **No fake tests** - Only show real data from the server
- **No filler content** - Every data point is real and actionable
- **Two focused views** - Inspector for now, Timeline for history
- **Premium aesthetic** - Clean, dense, command-center feel

### 1.3 Relation to Other Pages

| Page | Responsibility |
|------|---------------|
| **Dashboard** | Overview + Metrics + Commands + Logs + Alerts (tabs) |
| **Device** | Inbox + Overview + Telemetry + Commands + History (tabs) |
| **Diagnostics** | Inspector + Timeline (tabs) - Deep dive only |
| **Updates** | Version status + Changelog + Push updates |
| **Settings** | Configuration |

---

## 2. Page Structure

### 2.1 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  DIAGNOSTICS                                              [Refresh] │
├─────────────────────────────────────────────────────────────────────┤
│  [Inspector]  [Timeline]                                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  TAB CONTENT                                                       │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 Navigation

- Two tabs: **Inspector** (default) and **Timeline**
- Refresh button in header
- Tab state persists in URL

---

## 3. Tab 1: Inspector

### 3.1 Purpose

Shows what the server currently knows about the device and its connection state.

### 3.2 Sections

#### Section: Identity
| Field | Source | Description |
|-------|--------|-------------|
| IMEI | `device.imei` | 15-digit device identifier |
| Device Name | `device.deviceName` | User-defined friendly name |
| Model | `device.model` | Device model (from APK) |
| Manufacturer | `device.manufacturer` | Device manufacturer (from APK) |

#### Section: Software
| Field | Source | Description |
|-------|--------|-------------|
| OS Version | `device.osVersion` | Android version |
| App Version | `device.appVersion` | VyzorixAudioRouter version |
| Security Patch | `device.securityPatch` | Android security patch date |
| Build ID | `device.buildId` | Build fingerprint |

#### Section: Registration
| Field | Source | Description |
|-------|--------|-------------|
| Status | `device.status` | online / offline / deregistered |
| Registered At | `device.registeredAt` | Registration timestamp |
| FCM Token | `fcm.valid` | Token validity + last refresh |
| Command Secret | `commandSecret.set` | Whether secret is stored on device |

#### Section: Connection
| Field | Source | Description |
|-------|--------|-------------|
| WebSocket | `ws.status` | connected / disconnected |
| Connected Since | `ws.connectedAt` | When WS was established |
| FCM Status | `fcm.status` | Valid / Invalid / Not Set |
| Last Seen | `device.lastSeen` | Last telemetry or activity |
| IP Address | `ws.clientIp` | Client IP (if available) |
| Protocol | `ws.protocol` | WSS (TLS 1.3) |

#### Section: Telemetry
| Field | Source | Description |
|-------|--------|-------------|
| Last Frame | `telemetry.lastTimestamp` | Time of last telemetry |
| Frames Today | `telemetry.framesToday` | Count of frames in last 24h |
| Avg Latency | `telemetry.avgLatency` | Average WS round-trip |

---

## 4. Tab 2: Timeline

### 4.1 Purpose

Shows a chronological audit trail of events related to the device and connection.

### 4.2 Event Types

| Event | Icon | Description |
|-------|------|-------------|
| `TELEMETRY` | ● | Telemetry frame received |
| `COMMAND_SENT` | ○ | Command dispatched to device |
| `COMMAND_ACK` | ● | Command delivery confirmed |
| `COMMAND_FAILED` | ● | Command delivery failed |
| `CONNECTION_OPEN` | ● | WebSocket connected |
| `CONNECTION_LOST` | ● | WebSocket disconnected |
| `FCM_FALLBACK` | ● | Switched to FCM delivery |
| `RECONNECTED` | ● | WebSocket reconnected |
| `THRESHOLD_BREACH` | ● | Risk/Thermal exceeded threshold |
| `REGISTERED` | ● | Device registered |
| `DEREGISTERED` | ● | Device deregistered |
| `ERROR` | ● | Any error occurred |

### 4.3 Entry Format

```
┌─ TIMESTAMP ─────────────────────────────────────────────────────┐
│  ● EVENT_TYPE                                                    │
│  Description line 1                                               │
│  Description line 2 (if applicable)                              │
│  Metadata: key=value, key=value                                   │
└─────────────────────────────────────────────────────────────────┘
```

### 4.4 Example Entry

```
┌─ 12:34:56 ─────────────────────────────────────────────────────┐
│  ● TELEMETRY RECEIVED                                            │
│  Risk: 45  Thermal: 38.5°C  Buffer: 67%                       │
│  Latency: 23ms                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### 4.5 Controls

| Control | Type | Description |
|---------|------|-------------|
| Filter | Dropdown | Filter by event type |
| Auto-scroll | Toggle | Auto-scroll to newest |
| Clear | Button | Clear current timeline |
| Load More | Button | Load older events |

### 4.6 Pagination

- Load 50 events initially
- "Load More" button fetches next 50
- Infinite scroll option

---

## 5. REST API Specification

### 5.1 Inspector Endpoints

#### `GET /v1/device/:imei/inspect`
**Purpose:** Get full device inspection data

**Response (200 OK):**
```json
{
  "identity": {
    "imei": "861234567890123",
    "deviceName": "Pixel 8 Pro",
    "model": "Pixel 8",
    "manufacturer": "Google"
  },
  "software": {
    "osVersion": "Android 14",
    "appVersion": "2.1.0",
    "securityPatch": "2024-03-01",
    "buildId": "UP1A.231005.007"
  },
  "registration": {
    "status": "registered",
    "registeredAt": 1718900300000,
    "fcmTokenValid": true,
    "fcmTokenRefreshedAt": 1718890000000,
    "commandSecretSet": true
  },
  "connection": {
    "webSocketStatus": "connected",
    "connectedAt": 1718900000000,
    "fcmStatus": "valid",
    "lastSeen": 1718900500000,
    "clientIp": "192.168.1.xxx",
    "protocol": "WSS"
  },
  "telemetry": {
    "lastTimestamp": 1718900567000,
    "framesToday": 4521,
    "avgLatencyMs": 45
  }
}
```

---

#### `GET /v1/device/:imei/timeline`
**Purpose:** Get event timeline

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `eventType` | string | all | Filter: telemetry, command, connection, error |
| `startTime` | int64 | -24h | Start timestamp (ms) |
| `endTime` | int64 | now | End timestamp (ms) |
| `limit` | int | 50 | Max results |
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
        "riskScore": 45,
        "thermalTemp": 38.5,
        "bufferLevel": 67,
        "latencyMs": 23
      }
    },
    {
      "id": "uuid-v4",
      "type": "COMMAND_SENT",
      "timestamp": 1718900500000,
      "data": {
        "command": "WAKE_UP_UPDATER",
        "dispatchId": "abc123..."
      }
    }
  ],
  "pagination": {
    "limit": 50,
    "hasMore": true,
    "nextCursor": "base64-encoded-cursor"
  }
}
```

---

### 5.2 Event Types Enum

```json
{
  "TELEMETRY": "Device sent telemetry frame",
  "COMMAND_SENT": "Command dispatched to device",
  "COMMAND_ACK": "Command delivery confirmed",
  "COMMAND_FAILED": "Command delivery failed",
  "CONNECTION_OPEN": "WebSocket connected",
  "CONNECTION_LOST": "WebSocket disconnected",
  "FCM_FALLBACK": "Switched to FCM delivery",
  "RECONNECTED": "WebSocket reconnected",
  "THRESHOLD_BREACH": "Risk/Thermal exceeded threshold",
  "REGISTERED": "Device registered",
  "DEREGISTERED": "Device deregistered",
  "ERROR": "Error occurred"
}
```

---

## 6. GraphQL Schema

### 6.1 Types

```graphql
enum EventType {
  TELEMETRY
  COMMAND_SENT
  COMMAND_ACK
  COMMAND_FAILED
  CONNECTION_OPEN
  CONNECTION_LOST
  FCM_FALLBACK
  RECONNECTED
  THRESHOLD_BREACH
  REGISTERED
  DEREGISTERED
  ERROR
}

type Identity {
  imei: String!
  deviceName: String
  model: String
  manufacturer: String
}

type Software {
  osVersion: String!
  appVersion: String!
  securityPatch: String
  buildId: String
}

type Registration {
  status: DeviceStatus!
  registeredAt: DateTime
  fcmTokenValid: Boolean!
  fcmTokenRefreshedAt: DateTime
  commandSecretSet: Boolean!
}

type Connection {
  webSocketStatus: String!
  connectedAt: DateTime
  fcmStatus: String!
  lastSeen: DateTime
  clientIp: String
  protocol: String
}

type TelemetryStats {
  lastTimestamp: DateTime!
  framesToday: Int!
  avgLatencyMs: Int
}

type DeviceInspection {
  identity: Identity!
  software: Software!
  registration: Registration!
  connection: Connection!
  telemetry: TelemetryStats!
}

type TimelineEvent {
  id: ID!
  type: EventType!
  timestamp: DateTime!
  data: JSON
}

type TimelineConnection {
  events: [TimelineEvent!]!
  hasMore: Boolean!
  nextCursor: String
}
```

### 6.2 Queries

```graphql
type Query {
  deviceInspection(imei: String!): DeviceInspection!
  
  deviceTimeline(
    imei: String!
    eventType: EventType
    startTime: Int
    endTime: Int
    limit: Int = 50
    cursor: String
  ): TimelineConnection!
}
```

---

## 7. Frontend Components

### 7.1 Page Structure

```
/diagnostics
├── [Tabs]
│   ├── Inspector  → DiagnosticsInspector.tsx
│   └── Timeline  → DiagnosticsTimeline.tsx
```

### 7.2 Component List

| Component | File | Purpose |
|-----------|------|---------|
| DiagnosticsPage | `routes/diagnostics.tsx` | Page wrapper with tabs |
| DiagnosticsInspector | `components/diagnostics/DiagnosticsInspector.tsx` | Inspector tab |
| DiagnosticsTimeline | `components/diagnostics/DiagnosticsTimeline.tsx` | Timeline tab |
| TimelineEvent | `components/diagnostics/TimelineEvent.tsx` | Single event row |
| InspectorSection | `components/diagnostics/InspectorSection.tsx` | Collapsible section |

### 7.3 Hooks

| Hook | File | Purpose |
|------|------|---------|
| `useDeviceInspection` | `hooks/use-device-inspection.ts` | Fetch inspection data |
| `useDeviceTimeline` | `hooks/use-device-timeline.ts` | Fetch timeline events |

---

## 8. File Changes Summary

### 8.1 MODIFIED Files (Go Backend)

| File | Changes |
|------|---------|
| `internal/api/handlers/device/inspect.go` | NEW - inspection endpoint |
| `internal/api/handlers/device/timeline.go` | NEW - timeline endpoint |
| `internal/domain/device/device.go` | Add inspection fields |
| `internal/infrastructure/storage/device.go` | Add inspection queries |
| `internal/api/router.go` | Add routes |

### 8.2 NEW Files (Go Backend)

| File | Purpose |
|------|---------|
| `internal/api/handlers/device/inspect.go` | GET /v1/device/:imei/inspect |
| `internal/api/handlers/device/timeline.go` | GET /v1/device/:imei/timeline |

### 8.3 NEW Files (Frontend)

| File | Purpose |
|------|---------|
| `src/components/diagnostics/DiagnosticsInspector.tsx` | Inspector tab |
| `src/components/diagnostics/DiagnosticsTimeline.tsx` | Timeline tab |
| `src/components/diagnostics/TimelineEvent.tsx` | Event row component |
| `src/components/diagnostics/InspectorSection.tsx` | Collapsible section |
| `src/hooks/use-device-inspection.ts` | Inspection hook |
| `src/hooks/use-device-timeline.ts` | Timeline hook |

### 8.4 MODIFIED Files (Frontend)

| File | Changes |
|------|---------|
| `src/routes/diagnostics.tsx` | Replace with tab-based layout |
| `src/lib/api/graphql/queries.ts` | Add inspection/timeline queries |
| `src/lib/api/graphql/types.ts` | Add types |

---

## 9. Implementation Order

### Phase 1: Backend (Day 1)
1. Create `GET /v1/device/:imei/inspect` endpoint
2. Create `GET /v1/device/:imei/timeline` endpoint
3. Add timeline event logging to existing handlers

### Phase 2: Frontend Core (Day 1-2)
1. Create `useDeviceInspection` hook
2. Create `useDeviceTimeline` hook
3. Create `DiagnosticsInspector` component
4. Create `DiagnosticsTimeline` component

### Phase 3: Page Assembly (Day 2)
1. Update `diagnostics.tsx` with tab layout
2. Wire hooks to components
3. Add loading/error states
4. Add refresh functionality

### Phase 4: Polish (Day 2)
1. Timeline auto-scroll
2. Event filtering
3. Load more pagination
4. Animations

---

## 10. Design Notes

### 10.1 Visual Style

- Command center aesthetic
- Dense information display
- Rose-500 for highlights/accents only
- Minimal borders, use spacing
- Monospace for IDs and technical data

### 10.2 Interactions

| Element | Interaction |
|---------|-------------|
| Section headers | Click to collapse/expand |
| Timeline events | Hover to highlight |
| Refresh button | Spin icon while loading |
| Load More | Show spinner, disable button |

### 10.3 Empty States

| State | Message |
|-------|---------|
| No device selected | "Select a device to inspect" |
| No timeline events | "No events in the selected time range" |
| Connection error | "Failed to load diagnostics" |

---

*Document Version: 1.0*  
*Status: Ready for Implementation*
