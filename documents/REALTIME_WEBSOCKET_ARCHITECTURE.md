# Real-Time WebSocket Architecture - Enterprise Implementation Specification

> **Version:** 1.0
> **Status:** Draft
> **Created:** 2026-06-25
> **Target:** Production MVP
> **Architecture:** Layered (Following `FRONTEND_ARCHITECTURE.md`)

---

## Table of Contents

1. [Overview](#1-overview)
2. [Architecture](#2-architecture)
3. [Target File Structure](#3-target-file-structure)
4. [Backend Architecture](#4-backend-architecture)
5. [Frontend Architecture](#5-frontend-architecture)
6. [Domain Layer](#6-domain-layer)
7. [Data Layer](#7-data-layer)
8. [Presentation Layer - Hooks](#8-presentation-layer---hooks)
9. [UI Layer - Components](#9-ui-layer---components)
10. [File Changes Summary](#10-file-changes-summary)
11. [Implementation Order](#11-implementation-order)
12. [Testing Strategy](#12-testing-strategy)
13. [Rollout Checklist](#13-rollout-checklist)

---

> ⚠️ **Architecture Alignment Note (v1.0)**
>
> This document follows the **4-layer architecture** defined in `FRONTEND_ARCHITECTURE.md`:
> - **UI Layer** (`src/components/`) - Pure UI rendering, imports only from hooks
> - **Presentation Layer** (`src/hooks/`) - UI logic, state management, imports from domain & data
> - **Domain Layer** (`src/domain/` - NEW) - Types, transforms, validation (NO external imports)
> - **Data Layer** (`src/lib/api/`) - API clients (GraphQL/REST), imports only domain types
>
> **Dependency Rule:** UI → Hooks → Domain → API (flow inward only)

---

## 1. Overview

### 1.1 Purpose

The Real-Time WebSocket system enables bidirectional communication between registered Android devices and the operator dashboard:

- **Telemetry Ingestion:** Devices push real-time telemetry (risk score, thermal temp, buffer level, latency)
- **Command Dispatch:** Operators send commands to devices via the dashboard
- **Event Broadcasting:** Server pushes device events (connection status, alerts, threshold breaches) to dashboard clients
- **Fallback Communication:** FCM push notifications when WebSocket is unavailable

### 1.2 Communication Channels

| Channel | Direction | Protocol | Purpose |
|---------|-----------|----------|---------|
| **Telemetry Channel** | Device → Server | WebSocket (WSS) | Real-time telemetry push |
| **Command Channel** | Server → Device | WebSocket (primary), FCM (fallback) | Command dispatch |
| **Event Channel** | Server → Dashboard | WebSocket (WSS) | Live event broadcast |
| **Status Channel** | Bidirectional | WebSocket (WSS) | Connection health, heartbeat |

### 1.3 Data Flow Summary

```
┌─────────────────────────────────────────────────────────────────────┐
│                         DATA FLOW                                   │
└─────────────────────────────────────────────────────────────────────┘

DEVICE                         SERVER                        DASHBOARD
  │                               │                              │
  │  1. Connect WSS              │                              │
  │──────────────────────────────►│                              │
  │                               │                              │
  │  2. Authenticate             │                              │
  │  (HMAC signed token)         │                              │
  │──────────────────────────────►│                              │
  │                               │                              │
  │  3. Push Telemetry Frame     │                              │
  │  {riskScore, temp, buffer}   │                              │
  │──────────────────────────────►│──────┐                       │
  │                               │      │ Broadcast to           │
  │                               │      │ connected dashboard    │
  │                               │◄─────┘ clients                │
  │                               │                              │
  │                               │      ┌──────────────────────►│
  │                               │      │ Real-time update      │
  │                               │◄─────┼────── Dashboard      │
  │                               │      │ operator sends        │
  │                               │      │ command               │
  │  4. Receive Command          │      │                      │
  │◄──────────────────────────────│      │                      │
  │                               │      │                      │
  │  5. Execute & Ack            │      │                      │
  │  {dispatchId, status}        │      │                      │
  │──────────────────────────────►│──────┴──────────────────────►│
```

### 1.4 Message Protocols

#### Device → Server Messages

| Type | Purpose | Payload |
|------|---------|---------|
| `AUTH` | Device authentication | `{deviceToken}` |
| `TELEMETRY` | Periodic telemetry push | `{riskScore, temp, buffer, latency}` |
| `PONG` | Heartbeat response | `{}` |
| `CMD_ACK` | Command acknowledgment | `{dispatchId, status, result}` |

#### Server → Device Messages

| Type | Purpose | Payload |
|------|---------|---------|
| `CMD` | Command dispatch | `{dispatchId, command, parameters}` |
| `PING` | Heartbeat request | `{timestamp}` |
| `ACK` | Authentication response | `{success, error?}` |

#### Dashboard ↔ Server Messages

| Type | Direction | Purpose |
|------|-----------|---------|
| `AUTH` | → | Dashboard authentication |
| `SUBSCRIBE` | → | Subscribe to device events |
| `UNSUBSCRIBE` | → | Unsubscribe from device |
| `COMMAND` | → | Send command to device |
| `TELEMETRY` | ← | Real-time device telemetry |
| `EVENT` | ← | Device connection/disconnection, alerts |

#### GraphQL Subscriptions (Implemented)

The server implements GraphQL subscriptions for dashboard real-time updates:

| Subscription | Purpose | Payload |
|--------------|---------|---------|
| `deviceUpdated(deviceId: ID)` | Subscribe to device update events | `Device` |
| `telemetryReceived(deviceId: ID)` | Subscribe to real-time telemetry | `TelemetryEntry` |
| `commandStatusChanged(dispatchId: ID)` | Subscribe to command status changes | `Command` |

**Note on Device Authentication:** Device authentication uses HMAC middleware (HTTP header-based), not a message type.

**Note on Command Acknowledgments:** Command acknowledgments from devices go through the REST API (`POST /v1/device/:imei/command/:dispatchId/ack`), not WebSocket messages.

---

## 2. Architecture

### 2.1 Layered Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        FRONTEND ARCHITECTURE                        │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                      UI LAYER                               │   │
│  │                   (src/components/)                        │   │
│  │                                                             │   │
│  │    Pages, Components, Shared UI                            │   │
│  │    ONLY renders UI. Uses hooks for everything.              │   │
│  │    NEVER imports from Data or Domain.                       │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              │ uses                                  │
│                              ▼                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                   PRESENTATION LAYER                        │   │
│  │                      (src/hooks/)                          │   │
│  │                                                             │   │
│  │    UI Logic, State Management, Data Transformation           │   │
│  │    Imports from Domain and Data layers.                      │   │
│  │    NEVER imports UI components.                              │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              │ uses                                  │
│                              ▼                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                      DOMAIN LAYER                           │   │
│  │                       (src/domain/)                         │   │
│  │                                                             │   │
│  │    Types, Transforms, Validation (Pure TypeScript)          │   │
│  │    NO external imports (no React, no API, no i18n)          │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              │ uses                                  │
│                              ▼                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                       DATA LAYER                            │   │
│  │                    (src/lib/api/)                          │   │
│  │                                                             │   │
│  │    GraphQL Queries/Mutations, REST Endpoints, WS Client     │   │
│  │    Imports Domain types only.                               │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 Backend Architecture (Go)

```
┌─────────────────────────────────────────────────────────────────────┐
│                    BACKEND WEBSOCKET ARCHITECTURE                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │                      WebSocket Hub                          │   │
│   │                   (in-memory, single instance)               │   │
│   │                                                             │   │
│   │   ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │   │
│   │   │ Device Clients│  │Dashboard Clients│  │ Event Queue │        │   │
│   │   │  (IMEI keyed) │  │  (user keyed)  │  │  (in-memory) │        │   │
│   │   └───────┬──────┘  └───────┬──────┘  └───────┬──────┘        │   │
│   │           │                 │                 │                 │   │
│   │           └─────────────────┼─────────────────┘                 │   │
│   │                             │                                   │   │
│   │                      ┌──────▼──────┐                           │   │
│   │                      │  Broadcaster │                           │   │
│   │                      │  (fan-out)   │                           │   │
│   │                      └──────┬──────┘                           │   │
│   └────────────────────────────┼───────────────────────────────────┘   │
│                                │                                       │
│         ┌──────────────────────┼──────────────────────┐              │
│         │                      │                      │              │
│    ┌────▼────┐           ┌─────▼─────┐          ┌────▼────┐        │
│    │ Telemetry│           │  Command  │          │  Event  │        │
│    │ Processor│           │ Dispatcher│          │ Emitter │        │
│    └────┬────┘           └─────┬─────┘          └────┬────┘        │
│         │                      │                      │              │
│    ┌────▼────┐           ┌─────▼─────┐          ┌────▼────┐        │
│    │ SQLite  │           │    FCM   │          │ SQLite  │        │
│    │(telemetry)│          │(fallback)│          │(events) │        │
│    └─────────┘           └──────────┘          └─────────┘        │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.3 Connection Topology

```
┌─────────────────────────────────────────────────────────────────────┐
│                    SINGLE SERVER INSTANCE                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │                      WebSocket Hub                          │   │
│   │                                                             │   │
│   │   Connected Devices Map (IMEI → Client)                    │   │
│   │   ┌─────────────────────────────────────────────────────┐   │   │
│   │   │ IMEI_001 ──► Client{socket, lastPing, telemetry}   │   │   │
│   │   │ IMEI_002 ──► Client{socket, lastPing, telemetry}   │   │   │
│   │   │ IMEI_003 ──► Client{socket, lastPing, telemetry}   │   │   │
│   │   └─────────────────────────────────────────────────────┘   │   │
│   │                                                             │   │
│   │   Connected Dashboard Clients Map (UserID → Client)        │   │
│   │   ┌─────────────────────────────────────────────────────┐   │   │
│   │   │ User_001 ──► DashboardClient{socket, subscribed}   │   │   │
│   │   │ User_002 ──► DashboardClient{socket, subscribed}   │   │   │
│   │   └─────────────────────────────────────────────────────┘   │   │
│   │                                                             │   │
│   │   Message Routing                                           │   │
│   │   ┌─────────────────────────────────────────────────────┐   │   │
│   │   │ Device Telemetry ──► Broadcaster ──► Dashboards   │   │   │
│   │   │ Dashboard Command ──► Device (by IMEI)             │   │   │
│   │   │ System Event ──► Broadcaster ──► Dashboards        │   │   │
│   │   └─────────────────────────────────────────────────────┘   │   │
│   └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.4 Fallback Mechanisms

```
┌─────────────────────────────────────────────────────────────────────┐
│                         FALLBACK FLOW                               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1. WebSocket Available                                             │
│     └── Device ←──WSS──► Hub ←──WSS──► Dashboard                   │
│                                                                     │
│  2. WebSocket Disconnected (Device)                                 │
│     └── Hub ──FCM Push──► Device                                   │
│         └── FCM payload includes command + dispatchId             │
│                                                                     │
│  3. WebSocket Disconnected (Dashboard)                             │
│     └── Client auto-reconnects with exponential backoff           │
│                                                                     │
│  4. FCM Fallback Trigger                                            │
│     └── After 3 failed reconnection attempts                       │
│         └── Device marked as "WS_UNAVAILABLE"                      │
│         └── Commands queued in memory + FCM pushed                 │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 3. Target File Structure

### 3.1 Backend (Go) - EXISTING Structure

```
apps/api/internal/
├── domain/
│   ├── command/                   # EXISTING
│   │   ├── command_entity.go             # Command entity
│   │   └── command_repository.go         # Command repository interface
│   ├── device/                   # EXISTING
│   │   ├── device_entity.go             # Device entity
│   │   └── device_repository.go         # Device repository interface
│   └── telemetry/                # EXISTING
│       ├── telemetry_entity.go             # TelemetryFrame entity
│       └── telemetry_repository.go         # Telemetry repository interface
│
├── ws/                           # EXISTING - WebSocket core
│   ├── hub.go                    # Hub manages connections, broadcasts
│   ├── client.go                 # Client read/write pumps
│   ├── compression.go            # Message compression
│   ├── message_queue.go          # Offline message queuing
│   ├── rate_limiter.go           # Rate limiting
│   ├── subscriptions.go           # Telemetry subscriptions
│   └── telemetry_filter.go        # Telemetry filtering
│
├── api/
│   └── handlers/
│       └── websocket/             # EXISTING
│           ├── websocket_handler.go         # StreamHandler (WS upgrade entry)
│           ├── websocket_stream.go  # HTTP→WS upgrade logic
│           ├── websocket_message.go  # Message parsing
│           └── websocket_presenter.go       # Audit logging
│
├── application/
│   ├── command/                  # EXISTING - Command dispatch
│   ├── device/                  # EXISTING - Device management
│   ├── dto/                     # EXISTING - Data transfer objects
│   └── shared/                  # EXISTING
│
└── infrastructure/
    ├── fcm/                     # EXISTING - FCM notifications
    ├── storage/                  # EXISTING - SQLite storage
    └── metrics/                 # EXISTING - Prometheus metrics
```

### 3.2 Backend (Go) - NEW Files Required

```
apps/api/internal/
├── domain/
│   └── event/                   # NEW - Event entity
│       ├── event_entity.go            # Event types (DEVICE_CONNECTED, etc.)
│       └── event_repository.go        # Event repository interface
│
├── ws/                          # MODIFY existing
│   ├── hub.go                   # MODIFY - Add event broadcasting
│   └── client.go                # MODIFY - Add event emission on connect/disconnect
│
├── application/
│   └── event/                  # NEW - Event use cases
│       ├── event_broadcaster.go       # Broadcast events to dashboards
│       └── event_processor.go        # Process and emit events
│
└── infrastructure/
    └── storage/
        └── event_storage.go            # NEW - Event SQLite storage
```

### 3.3 Frontend (React)

```
apps/web/src/
│
├── domain/                          # DOMAIN LAYER
│   ├── shared/
│   │   ├── pagination.ts
│   │   ├── errors.ts
│   │   └── types.ts
│   │
│   └── realtime/
│       ├── types.ts                 # WSTelemetry, WSEvent, WSCommand
│       ├── transforms.ts            # telemetryFromRaw(), eventFromRaw()
│       └── validation.ts            # validateTelemetry(), validateEvent()
│
├── lib/
│   └── api/
│       ├── graphql/
│       │   ├── client.ts            # (EXISTING)
│       │   ├── queries/
│       │   │   └── realtime.ts     # Subscription queries (if needed)
│       │   └── mutations/
│       │       └── realtime.ts      # Command mutations
│       │
│       └── websocket/
│           ├── client.ts            # WebSocket client wrapper
│           ├── connection.ts        # Connection state machine
│           ├── heartbeat.ts        # Heartbeat manager
│           ├── reconnect.ts         # Reconnection logic
│           └── messages.ts          # Message types/parsers
│
├── hooks/
│   └── realtime/                    # PRESENTATION LAYER
│       ├── use-websocket-connection.ts
│       ├── use-device-telemetry.ts
│       ├── use-dashboard-events.ts
│       ├── use-command-dispatch.ts
│       └── index.ts
│
└── components/
    ├── shared/                      # SHARED UI LAYER
    │   ├── connection-status.tsx
    │   ├── reconnecting-indicator.tsx
    │   ├── offline-banner.tsx
    │   └── index.ts
    │
    └── realtime/                    # REALTIME UI LAYER
        ├── telemetry-feed.tsx
        ├── event-feed.tsx
        ├── command-status.tsx
        └── index.ts
```

---

## 4. Backend Architecture

### 4.1 Existing WebSocket Hub (EXISTING)

The hub at `internal/ws/hub.go` manages all WebSocket connections in-memory:

| Operation | Description |
|-----------|-------------|
| `Register(client)` | Add device client to hub |
| `Unregister(client)` | Remove device client |
| `Online(deviceID)` | Check if device is connected |
| `Clients()` | Get all connected clients |
| `Send(deviceID, frame)` | Send command to device |
| `BroadcastTelemetry(raw)` | Broadcast telemetry to all |
| `BroadcastTelemetryToFiltered(sender, raw)` | Filtered telemetry broadcast |

### 4.2 New Event System (NEEDED)

Event broadcasting to be added:

| Operation | Description |
|-----------|-------------|
| `BroadcastEvent(event)` | Emit event to all dashboards |
| `EmitDeviceConnected(deviceID)` | Device established WS connection |
| `EmitDeviceDisconnected(deviceID)` | Device connection closed |
| `EmitThresholdBreach(deviceID, metric)` | Telemetry threshold exceeded |

### 4.3 Message Processing

#### Telemetry Flow
1. Device sends `TELEMETRY` message
2. Hub validates IMEI matches authenticated device
3. Broadcaster fans out to all dashboard clients
4. Telemetry processor stores frame in SQLite

#### Command Flow
1. Dashboard sends `COMMAND` message via WSS
2. Hub routes to target device by IMEI
3. If device connected: immediate delivery via WSS
4. If device disconnected: queue + FCM fallback

### 4.3 Event Types

| Event | Description |
|-------|-------------|
| `DEVICE_CONNECTED` | Device established WebSocket connection |
| `DEVICE_DISCONNECTED` | Device connection closed (clean or timeout) |
| `THRESHOLD_BREACH` | Telemetry exceeded configured threshold |
| `COMMAND_DELIVERED` | Command was successfully delivered |
| `COMMAND_FAILED` | Command delivery failed |

### 4.4 Heartbeat Protocol (EXISTING)

| Step | Description |
|------|-------------|
| 1 | Server sends `PING` every 30 seconds |
| 2 | Client responds with `PONG` within 10 seconds |
| 3 | Missing 3 consecutive PONGs = connection dead |
| 4 | Server closes connection, emits `DEVICE_DISCONNECTED` |

---

## 5. Frontend Architecture

### 5.1 WebSocket Client

The WebSocket client manages the connection lifecycle:

```
┌─────────────────────────────────────────────────────────────────────┐
│                    WEBSOCKET CLIENT STATE MACHINE                    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ┌─────────┐     connect()      ┌─────────────┐                   │
│   │ CLOSED  │ ─────────────────► │ CONNECTING  │                   │
│   └─────────┘                    └──────┬──────┘                   │
│        ▲                               │                           │
│        │                         onOpen │                           │
│        │                        success ▼                           │
│        │                      ┌─────────────┐                       │
│        ├───────────────────── │   OPEN     │                       │
│        │                      └──────┬──────┘                       │
│        │                             │                              │
│        │                      ping/pong                           │
│        │                             │                              │
│        │                       onClose │                           │
│        │                             ▼                              │
│        │                      ┌─────────────┐                       │
│        ├───────────────────── │ CLOSING    │                       │
│        │                      └──────┬──────┘                       │
│        │                             │                              │
│        │                    onClosed │                              │
│        │                             ▼                              │
│        │                      ┌─────────────┐                       │
│        └───────────────────── │   RECONNECT │ (exponential backoff)│
│                               └─────────────┘                       │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 5.2 Reconnection Strategy

| Attempt | Delay | Max Delay |
|---------|-------|-----------|
| 1 | 1s | 30s |
| 2 | 2s | 30s |
| 3 | 4s | 30s |
| 4 | 8s | 30s |
| 5+ | 30s | 30s |

After 5 failed attempts, trigger FCM fallback for commands.

---

## 6. Domain Layer

### 6.1 Types

```typescript
// domain/realtime-entity.ts

export interface WSTelemetry {
  imei: string;
  riskScore: number;
  thermalTemp: number;
  bufferLevel: number;
  latencyMs: number;
  timestamp: Date;
}

export interface WSEvent {
  id: string;
  eventType: WSEventType;
  deviceImei: string;
  deviceName?: string;
  timestamp: Date;
  data: Record<string, unknown>;
}

export type WSEventType =
  | 'DEVICE_CONNECTED'
  | 'DEVICE_DISCONNECTED'
  | 'THRESHOLD_BREACH'
  | 'COMMAND_DELIVERED'
  | 'COMMAND_FAILED';

export interface WSCommand {
  dispatchId: string;
  deviceImei: string;
  command: WSCommandType;
  parameters: Record<string, unknown>;
  priority: 'high' | 'normal' | 'low';
}

export type WSCommandType =
  | 'FORCE_SPEAKER'
  | 'RESET_AUDIO_HAL'
  | 'TOGGLE_CAPTURE'
  | 'REINIT_PROJECTION'
  | 'DUMP_FLIGHT_DATA'
  | 'UPLOAD_CRASH_ZIP'
  | 'SET_LOG_LEVEL'
  | 'WAKE_UP_UPDATER';
```

### 6.2 Transforms

```typescript
// domain/realtime-mappers.ts

import type { WSTelemetry, WSEvent } from './types';

interface RawTelemetry {
  imei: string;
  riskScore: number;
  thermalTemp: number;
  bufferLevel: number;
  latencyMs: number;
  timestamp: string;
}

export const telemetryFromRaw = (raw: RawTelemetry): WSTelemetry => ({
  imei: raw.imei,
  riskScore: raw.riskScore,
  thermalTemp: raw.thermalTemp,
  bufferLevel: raw.bufferLevel,
  latencyMs: raw.latencyMs,
  timestamp: new Date(raw.timestamp),
});

interface RawEvent {
  id: string;
  eventType: string;
  deviceImei: string;
  deviceName?: string;
  timestamp: string;
  data: Record<string, unknown>;
}

export const eventFromRaw = (raw: RawEvent): WSEvent => ({
  id: raw.id,
  eventType: raw.eventType as WSEvent['eventType'],
  deviceImei: raw.deviceImei,
  deviceName: raw.deviceName,
  timestamp: new Date(raw.timestamp),
  data: raw.data,
});
```

### 6.3 Validation

```typescript
// domain/realtime-validators.ts

import type { WSTelemetry, WSCommand } from './types';

export const validateTelemetry = (data: unknown): data is WSTelemetry => {
  if (!data || typeof data !== 'object') return false;
  const t = data as Record<string, unknown>;
  return (
    typeof t.imei === 'string' &&
    typeof t.riskScore === 'number' &&
    t.riskScore >= 0 &&
    t.riskScore <= 100 &&
    typeof t.thermalTemp === 'number' &&
    typeof t.bufferLevel === 'number' &&
    t.bufferLevel >= 0 &&
    t.bufferLevel <= 100
  );
};

export const validateCommand = (data: unknown): data is WSCommand => {
  if (!data || typeof data !== 'object') return false;
  const c = data as Record<string, unknown>;
  return (
    typeof c.dispatchId === 'string' &&
    typeof c.deviceImei === 'string' &&
    typeof c.command === 'string'
  );
};
```

---

## 7. Data Layer

### 7.1 WebSocket Client

```typescript
// lib/api/websocket/websocket-client.ts

import type { WSTelemetry, WSEvent, WSCommand } from '@/domain/realtime';

export type WSMessageType =
  | 'AUTH'
  | 'TELEMETRY'
  | 'COMMAND'
  | 'EVENT'
  | 'PING'
  | 'PONG'
  | 'CMD_ACK'
  | 'SUBSCRIBE'
  | 'UNSUBSCRIBE';

export interface WSMessage<T = unknown> {
  type: WSMessageType;
  payload: T;
}

export interface WebSocketClient {
  connect(): Promise<void>;
  disconnect(): void;
  send(message: WSMessage): void;
  onTelemetry(handler: (telemetry: WSTelemetry) => void): void;
  onEvent(handler: (event: WSEvent) => void): void;
  onCommand(handler: (command: WSCommand) => void): void;
  onConnectionChange(handler: (connected: boolean) => void): void;
}
```

### 7.2 Connection State Machine

```typescript
// lib/api/websocket/websocket-connection.ts

export type ConnectionState =
  | 'CLOSED'
  | 'CONNECTING'
  | 'OPEN'
  | 'CLOSING'
  | 'RECONNECTING';

export interface ConnectionStateMachine {
  state: ConnectionState;
  transition(newState: ConnectionState): void;
  onStateChange(handler: (state: ConnectionState) => void): void;
}
```

### 7.3 Heartbeat Manager

```typescript
// lib/api/websocket/websocket-heartbeat.ts

export interface HeartbeatManager {
  start(): void;
  stop(): void;
  onMissedHeartbeat(handler: () => void): void;
  getLastPingTime(): Date | null;
  getRTT(): number | null;
}
```

---

## 8. Presentation Layer - Hooks

### 8.1 useWebSocketConnection

```typescript
// hooks/realtime/use-websocket-connection.ts

import { useWebSocket } from '@/lib/api/websocket/client';

interface UseWebSocketConnectionResult {
  isConnected: boolean;
  isReconnecting: boolean;
  connect: () => Promise<void>;
  disconnect: () => void;
  lastConnectedAt: Date | null;
  connectionError: Error | null;
}

export const useWebSocketConnection = (): UseWebSocketConnectionResult => {
  // Implementation
};
```

### 8.2 useDeviceTelemetry

```typescript
// hooks/realtime/use-device-telemetry.ts

import type { WSTelemetry } from '@/domain/realtime';

interface UseDeviceTelemetryOptions {
  imei?: string;
}

interface UseDeviceTelemetryResult {
  telemetry: WSTelemetry | null;
  telemetryHistory: WSTelemetry[];
  isLoading: boolean;
  error: Error | null;
}

export const useDeviceTelemetry = (
  options: UseDeviceTelemetryOptions = {}
): UseDeviceTelemetryResult => {
  // Implementation
};
```

### 8.3 useDashboardEvents

```typescript
// hooks/realtime/use-dashboard-events.ts

import type { WSEvent, WSEventType } from '@/domain/realtime';

interface UseDashboardEventsOptions {
  eventTypes?: WSEventType[];
  deviceImei?: string;
}

interface UseDashboardEventsResult {
  events: WSEvent[];
  unreadCount: number;
  markAsRead: (eventId: string) => void;
  clearEvents: () => void;
}

export const useDashboardEvents = (
  options: UseDashboardEventsOptions = {}
): UseDashboardEventsResult => {
  // Implementation
};
```

### 8.4 useCommandDispatch

```typescript
// hooks/realtime/use-command-dispatch.ts

import type { WSCommand, WSCommandType } from '@/domain/realtime';

interface UseCommandDispatchOptions {
  imei: string;
}

interface UseCommandDispatchResult {
  sendCommand: (command: WSCommandType, parameters?: Record<string, unknown>) => Promise<string>;
  pendingCommands: WSCommand[];
  commandStatus: Map<string, 'pending' | 'delivered' | 'failed'>;
}

export const useCommandDispatch = (
  options: UseCommandDispatchOptions
): UseCommandDispatchResult => {
  // Implementation
};
```

---

## 9. UI Layer - Components

### 9.1 ConnectionStatus

```typescript
// components/shared/connection-status.tsx

import { useWebSocketConnection } from '@/hooks/realtime/use-websocket-connection';

interface ConnectionStatusProps {
  deviceImei?: string;
}

export const ConnectionStatus = ({ deviceImei }: ConnectionStatusProps) => {
  const { isConnected, isReconnecting } = useWebSocketConnection();

  if (isConnected) {
    return <Badge variant="success">Connected</Badge>;
  }

  if (isReconnecting) {
    return <Badge variant="warning">Reconnecting...</Badge>;
  }

  return <Badge variant="destructive">Disconnected</Badge>;
};
```

### 9.2 ReconnectingIndicator

```typescript
// components/shared/reconnecting-indicator.tsx

import { useWebSocketConnection } from '@/hooks/realtime/use-websocket-connection';

export const ReconnectingIndicator = () => {
  const { isReconnecting } = useWebSocketConnection();

  if (!isReconnecting) return null;

  return (
    <div className="flex items-center gap-2 text-amber-600">
      <RefreshCw className="h-4 w-4 animate-spin" />
      <span>Reconnecting...</span>
    </div>
  );
};
```

### 9.3 OfflineBanner

```typescript
// components/shared/offline-banner.tsx

import { useWebSocketConnection } from '@/hooks/realtime/use-websocket-connection';

export const OfflineBanner = () => {
  const { isConnected } = useWebSocketConnection();

  if (isConnected) return null;

  return (
    <div className="bg-destructive text-destructive-foreground px-4 py-2 text-center">
      Real-time connection lost. Some data may be outdated.
    </div>
  );
};
```

### 9.4 TelemetryFeed

```typescript
// components/realtime/telemetry-feed.tsx

import { useDeviceTelemetry } from '@/hooks/realtime/use-device-telemetry';

interface TelemetryFeedProps {
  imei: string;
}

export const TelemetryFeed = ({ imei }: TelemetryFeedProps) => {
  const { telemetry, telemetryHistory } = useDeviceTelemetry({ imei });

  if (!telemetry) {
    return <Skeleton className="h-32" />;
  }

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <MetricCard label="Risk Score" value={telemetry.riskScore} />
        <MetricCard label="Temperature" value={`${telemetry.thermalTemp}°C`} />
        <MetricCard label="Buffer" value={`${telemetry.bufferLevel}%`} />
        <MetricCard label="Latency" value={`${telemetry.latencyMs}ms`} />
      </div>
    </div>
  );
};
```

### 9.5 EventFeed

```typescript
// components/realtime/event-feed.tsx

import { useDashboardEvents } from '@/hooks/realtime/use-dashboard-events';

interface EventFeedProps {
  deviceImei?: string;
}

export const EventFeed = ({ deviceImei }: EventFeedProps) => {
  const { events, unreadCount } = useDashboardEvents({ deviceImei });

  return (
    <div className="space-y-2">
      {unreadCount > 0 && (
        <Badge variant="secondary">{unreadCount} new events</Badge>
      )}
      {events.map((event) => (
        <EventCard key={event.id} event={event} />
      ))}
    </div>
  );
};
```

### 9.6 CommandStatus

```typescript
// components/realtime/command-status.tsx

import { useCommandDispatch } from '@/hooks/realtime/use-command-dispatch';

interface CommandStatusProps {
  imei: string;
}

export const CommandStatus = ({ imei }: CommandStatusProps) => {
  const { pendingCommands, commandStatus } = useCommandDispatch({ imei });

  if (pendingCommands.length === 0) {
    return null;
  }

  return (
    <div className="space-y-2">
      <h3 className="font-semibold">Pending Commands</h3>
      {pendingCommands.map((cmd) => (
        <CommandCard
          key={cmd.dispatchId}
          command={cmd}
          status={commandStatus.get(cmd.dispatchId)}
        />
      ))}
    </div>
  );
};
```

---

## 10. File Changes Summary

### 10.1 Total File Count

| Side | Category | New Files | Modified Files |
|------|----------|-----------|----------------|
| **Backend (Go)** | Domain | 1 | 0 |
| **Backend (Go)** | Application | 2 | 0 |
| **Backend (Go)** | Infrastructure | 1 | 0 |
| **Backend (Go)** | WS (existing) | 0 | 2 |
| **Frontend (React)** | Domain | 3 | 0 |
| **Frontend (React)** | Data Layer | 5 | 0 |
| **Frontend (React)** | Presentation Layer | 4 | 0 |
| **Frontend (React)** | UI Layer | 6 | 0 |
| | **TOTAL** | **22** | **2** |

### 10.2 Backend Files (Go)

#### Domain Layer (1 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `internal/domain/event/event_entity.go` | **NEW** | Event entity (DEVICE_CONNECTED, THRESHOLD_BREACH, etc.) |
| `internal/domain/event/event_repository.go` | **NEW** | Event repository interface |

#### Application Layer (2 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `internal/application/event/event_broadcaster.go` | **NEW** | Broadcast events to dashboard clients |
| `internal/application/event/event_processor.go` | **NEW** | Process and emit events |

#### Infrastructure Layer (1 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `internal/infrastructure/storage/event_storage.go` | **NEW** | Event SQLite storage |

#### WS Layer - MODIFIED (2 files)

| File | Status | Purpose |
|------|--------|---------|
| `internal/ws/hub.go` | **MODIFY** | Add event broadcasting methods |
| `internal/ws/client.go` | **MODIFY** | Emit connect/disconnect events |

---

### 10.3 Frontend Files (React)

#### Domain Layer (3 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `domain/realtime-entity.ts` | **NEW** | WSTelemetry, WSEvent, WSCommand types |
| `domain/realtime-mappers.ts` | **NEW** | telemetryFromRaw(), eventFromRaw() |
| `domain/realtime-validators.ts` | **NEW** | validateTelemetry(), validateCommand() |

#### Data Layer - WebSocket (5 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `lib/api/websocket/websocket-client.ts` | **NEW** | WebSocket client wrapper |
| `lib/api/websocket/websocket-connection.ts` | **NEW** | Connection state machine |
| `lib/api/websocket/websocket-heartbeat.ts` | **NEW** | Heartbeat manager |
| `lib/api/websocket/websocket-reconnect.ts` | **NEW** | Reconnection logic |
| `lib/api/websocket/websocket-messages.ts` | **NEW** | Message types/parsers |

#### Presentation Layer (4 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `hooks/realtime/use-websocket-connection.ts` | **NEW** | Connection state hook |
| `hooks/realtime/use-device-telemetry.ts` | **NEW** | Telemetry subscription hook |
| `hooks/realtime/use-dashboard-events.ts` | **NEW** | Event subscription hook |
| `hooks/realtime/use-command-dispatch.ts` | **NEW** | Command dispatch hook |
| `hooks/realtime/index.ts` | **NEW** | Barrel export |

#### UI Layer - Shared (3 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `components/shared/connection-status.tsx` | **NEW** | Connection status badge |
| `components/shared/reconnecting-indicator.tsx` | **NEW** | Reconnecting spinner |
| `components/shared/offline-banner.tsx` | **NEW** | Offline warning banner |
| `components/shared/index.ts` | **NEW** | Barrel export (update) |

#### UI Layer - Realtime (3 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `components/realtime/telemetry-feed.tsx` | **NEW** | Telemetry display |
| `components/realtime/event-feed.tsx` | **NEW** | Event feed |
| `components/realtime/command-status.tsx` | **NEW** | Command status display |
| `components/realtime/index.ts` | **NEW** | Barrel export |

---

## 11. Implementation Order

### Phase 1: Backend - Domain & Infrastructure (Day 1)
1. Create `internal/domain/event/event_entity.go`
2. Create `internal/domain/event/event_repository.go`
3. Create `internal/infrastructure/storage/event_storage.go`

### Phase 2: Backend - Event Broadcasting (Day 1-2)
1. Create `internal/application/event/event_broadcaster.go`
2. Create `internal/application/event/event_processor.go`
3. Modify `internal/ws/hub.go` - add event broadcasting methods
4. Modify `internal/ws/client.go` - emit connect/disconnect events

### Phase 3: Frontend - Domain Layer (Day 2)
1. Create `domain/realtime-entity.ts`
2. Create `domain/realtime-mappers.ts`
3. Create `domain/realtime-validators.ts`

### Phase 4: Frontend - Data Layer (Day 2)
1. Create `lib/api/websocket/websocket-messages.ts`
2. Create `lib/api/websocket/websocket-connection.ts`
3. Create `lib/api/websocket/websocket-heartbeat.ts`
4. Create `lib/api/websocket/websocket-reconnect.ts`
5. Create `lib/api/websocket/websocket-client.ts`

### Phase 5: Frontend - Presentation Layer (Day 2-3)
1. Create `hooks/realtime/use-websocket-connection.ts`
2. Create `hooks/realtime/use-device-telemetry.ts`
3. Create `hooks/realtime/use-dashboard-events.ts`
4. Create `hooks/realtime/use-command-dispatch.ts`

### Phase 6: Frontend - UI Layer (Day 3)
1. Create `components/shared/connection-status.tsx`
2. Create `components/shared/reconnecting-indicator.tsx`
3. Create `components/shared/offline-banner.tsx`
4. Create `components/realtime/telemetry-feed.tsx`
5. Create `components/realtime/event-feed.tsx`
6. Create `components/realtime/command-status.tsx`

### Phase 7: Integration (Day 3-4)
1. Wire hooks to components
2. Add to dashboard pages
3. Test real-time updates

---

## 12. Testing Strategy

### Unit Tests
- Domain transforms (`domain/realtime/transforms.test.ts`)
- Domain validation (`domain/realtime/validation.test.ts`)
- WebSocket client state machine
- Heartbeat manager

### Integration Tests
- WebSocket connection lifecycle
- Reconnection logic with mock server
- Command dispatch flow

### E2E Tests
- Device connects and sends telemetry
- Dashboard receives telemetry in real-time
- Command sent from dashboard reaches device
- Reconnection after network loss

---

## 13. Rollout Checklist

### Pre-Launch
- [ ] All hooks have loading states
- [ ] All hooks have error states
- [ ] Reconnection works correctly
- [ ] FCM fallback tested
- [ ] Memory leak check (event history cleanup)

### Post-Launch
- [ ] Monitor WebSocket connection stability
- [ ] Monitor FCM fallback rates
- [ ] Check for memory leaks in long sessions

---

*Document Version: 1.0*
*Status: Ready for Implementation*
*Architecture: Layered (Following FRONTEND_ARCHITECTURE.md)*
