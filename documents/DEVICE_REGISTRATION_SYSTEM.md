# Device Registration System - Enterprise Implementation Specification

> **Version:** 1.0  
> **Status:** Draft  
> **Created:** 2026-06-21  
> **Target:** Production MVP  

---

## Table of Contents

1. [Overview](#1-overview)
2. [System Architecture](#2-system-architecture)
3. [State Machine](#3-state-machine)
4. [REST API Specification](#4-rest-api-specification)
5. [GraphQL Schema](#5-graphql-schema)
6. [Database Schema](#6-database-schema)
7. [Frontend Components](#7-frontend-components)
8. [File Changes Summary](#8-file-changes-summary)
9. [Implementation Order](#9-implementation-order)

---

## 1. Overview

### 1.1 Problem Statement

Current device registration requires:
- Manual form entry on dashboard
- User manually types Firebase IDs, device class
- Single device focus
- No visibility into pending registrations
- No machine-to-machine handshake

### 1.2 Solution

Zero-friction device registration with:
- Device auto-reports on first boot
- Operator reviews in inbox
- Machine-to-machine confirmation handshake
- Multi-device support
- Full audit trail

### 1.3 Flow Summary

```
┌─────────────────────────────────────────────────────────────────────┐
│  DEVICE (APK)                    SERVER                       DASHBOARD
│
│  1. APK starts, fetches device info (IMEI, firmware, OS, etc.)
│  2. POST /v1/device/inbox (device sends registration request)
│  3. Server stores in INBOX (status: pending)
│  4. Operator views inbox, clicks "Register"
│  5. Server validates, generates commandSecret, FCM push to device
│  6. Device receives push, stores commandSecret, POST /v1/device/confirm
│  7. Server marks device as REGISTERED
└─────────────────────────────────────────────────────────────────────┘
```

---

## 2. System Architecture

### 2.1 Components

```
┌─────────────────────────────────────────────────────────────────────┐
│                         FRONTEND (React)                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐               │
│  │  DevicePage │  │  InboxView  │  │ CommandsTab │               │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘               │
│         │                │                │                        │
│         └────────────────┼────────────────┘                        │
│                          │                                         │
│                    ┌─────▼─────┐                                  │
│                    │  GraphQL  │ (primary)                         │
│                    │  Client   │                                   │
│                    └─────┬─────┘                                  │
│                          │                                         │
│                    ┌─────▼─────┐                                  │
│                    │   REST    │ (fallback)                        │
│                    │   Client  │                                   │
│                    └─────┬─────┘                                  │
└──────────────────────────┼─────────────────────────────────────────┘
                           │
                    ┌──────▼──────┐
                    │   SERVER    │
                    │   (Go)     │
                    └──────┬──────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
    ┌────▼────┐      ┌─────▼─────┐     ┌────▼────┐
    │ Inbox   │      │  Device   │     │ Command │
    │ Handler │      │  Handler  │     │ Handler │
    └────┬────┘      └─────┬─────┘     └────┬────┘
         │                 │                 │
    ┌────▼────┐      ┌─────▼─────┐     ┌────▼────┐
    │ Inbox   │      │  Device   │     │ Command │
    │ Store   │      │  Store    │     │ Store   │
    └────┬────┘      └─────┬─────┘     └────┬────┘
         │                 │                 │
         └─────────────────┼─────────────────┘
                           │
                    ┌──────▼──────┐
                    │   SQLite    │
                    │   Database  │
                    └─────────────┘
```

### 2.2 Directory Structure

```
apps/
├── api/
│   └── internal/
│       ├── api/
│       │   ├── handlers/
│       │   │   ├── inbox/
│       │   │   │   ├── inbox.go           # Inbox handlers
│       │   │   │   ├── inbox_test.go
│       │   │   │   └── routes.go          # Inbox routes
│       │   │   └── device/
│       │   │       ├── register.go         # Existing - update
│       │   │       ├── confirm.go          # New - confirm endpoint
│       │   │       ├── status.go
│       │   │       ├── list.go
│       │   │       └── deregister.go
│       │   │
│       │   ├── middleware/
│       │   │   └── auth.go
│       │   │
│       │   └── router.go
│       │
│       ├── domain/
│       │   ├── device/
│       │   │   ├── device.go               # Device entity
│       │   │   ├── inbox.go                # NEW: InboxEntry entity
│       │   │   ├── errors.go
│       │   │   └── repository.go           # Interfaces
│       │   │
│       │   └── command/
│       │
│       ├── infrastructure/
│       │   ├── storage/
│       │   │   ├── device.go               # Existing
│       │   │   ├── inbox.go                # NEW: Inbox storage
│       │   │   └── migrations/
│       │   │       └── 001_device_inbox.sql
│       │   │
│       │   ├── fcm/
│       │   │   └── notifier.go
│       │   │
│       │   └── crypto/
│       │
│       └── application/
│           ├── device/
│           │   ├── register_usecase.go
│           │   ├── inbox_usecase.go
│           │   └── command_usecase.go
│           │
│           └── dto/
│               ├── inbox.go
│               └── device.go
│
└── web/
    └── src/
        ├── routes/
        │   └── device.tsx                  # Update with tabs
        │
        ├── components/
        │   └── device/
        │       ├── DeviceInbox.tsx          # NEW: Inbox component
        │       ├── DeviceOverview.tsx       # NEW: Overview tab
        │       ├── DeviceTelemetry.tsx      # NEW: Telemetry tab
        │       ├── DeviceCommands.tsx       # NEW: Commands tab
        │       └── DeviceHistory.tsx        # NEW: History tab
        │
        ├── lib/
        │   └── api/
        │       ├── graphql/
        │       │   ├── queries.ts          # Update
        │       │   ├── mutations.ts        # Update
        │       │   └── types.ts           # Update
        │       │
        │       └── rest/
        │           └── device-client.ts    # NEW: REST fallback
```

---

## 3. State Machine

### 3.1 Inbox States

```
┌─────────────────────────────────────────────────────────────────────┐
│                         INBOX STATE MACHINE                         │
└─────────────────────────────────────────────────────────────────────┘

UNREGISTERED ─────────────────────────────────────────────────────
     │
     │  POST /v1/device/inbox (device sends request)
     ▼
┌───────────┐     FCM sent to device, awaiting acknowledgment
│  PENDING  │
└─────┬─────┘
      │
      │  POST /v1/device/inbox/:imei/ack (device acknowledges)
      ▼
┌──────────────┐   Device has seen the request
│ ACKNOWLEDGED│
└──────┬───────┘
       │
       │  POST /v1/device/register (operator clicks Register)
       ▼
┌──────────┐    Server validating, generating commandSecret, FCM push
│ APPROVING│
└──────┬───┘
       │
       │  POST /v1/device/confirm (device confirms)
       ▼
┌────────────┐
│ REGISTERED │ ◄─────────────────────────────────────────────
└────────────┘
       │
       │  (on failure/rejection)
       ▼
┌───────────┐
│ REJECTED  │ ◄── Operator clicks Dismiss
└───────────┘
       │
       │  (auto-cleanup after 30 days)
       ▼
┌───────────┐
│ EXPIRED   │ ◄── No action taken
└───────────┘
```

### 3.2 Device States

```
┌─────────────────────────────────────────────────────────────────────┐
│                         DEVICE STATE MACHINE                        │
└─────────────────────────────────────────────────────────────────────┘

REGISTERED ─────────────────────────────────────────────────────────
     │
     │  First telemetry OR WebSocket connect
     ▼
┌─────────┐
│ ONLINE  │ ◄──────────────────────────────────────────────────
└────┬────┘
     │     WebSocket disconnect OR no telemetry for 30s
     ▼
┌──────────┐
│ OFFLINE │
└────┬─────┘
     │     WebSocket reconnect OR telemetry received
     ▼
┌─────────┐
│ ONLINE  │
└────┬────┘
     │
     │  DELETE /v1/device/:imei (operator deregisters)
     ▼
┌──────────────┐
│ DEREGISTERED │ ◄── Terminal state
└──────────────┘
```

---

## 4. REST API Specification

### 4.1 Inbox Endpoints

#### `POST /v1/device/inbox`
**Purpose:** Device sends registration request

**Request:**
```json
{
  "imei": "861234567890123",
  "deviceName": "Pixel 8 Pro",
  "model": "Pixel 8",
  "manufacturer": "Google",
  "osVersion": "Android 14",
  "appVersion": "2.1.0",
  "fcmToken": "dGhpcyBpcyBhIGZjIGtleTohZGkh", 
  "firmware": "oriole-user 14 UP1A.231005.007 11095021 release-keys",
  "securityPatch": "2024-03-01",
  "buildId": "UP1A.231005.007"
}
```

**Response (201 Created):**
```json
{
  "status": "pending",
  "messageId": "uuid-v4",
  "receivedAt": 1718900000000
}
```

**Error Responses:**
- `400 Bad Request` - Missing required fields
- `409 Conflict` - IMEI already in inbox or registered

---

#### `GET /v1/device/inbox`
**Purpose:** List all inbox entries (operator)

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | string | all | Filter: pending, acknowledged, approving, rejected |
| `page` | int | 1 | Page number |
| `limit` | int | 20 | Items per page |

**Response (200 OK):**
```json
{
  "entries": [
    {
      "id": "uuid-v4",
      "imei": "861234567890123",
      "deviceName": "Pixel 8 Pro",
      "model": "Pixel 8",
      "manufacturer": "Google",
      "osVersion": "Android 14",
      "appVersion": "2.1.0",
      "status": "pending",
      "receivedAt": 1718900000000,
      "updatedAt": 1718900000000
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

#### `GET /v1/device/inbox/:imei`
**Purpose:** Get single inbox entry details

**Response (200 OK):**
```json
{
  "id": "uuid-v4",
  "imei": "861234567890123",
  "deviceName": "Pixel 8 Pro",
  "model": "Pixel 8",
  "manufacturer": "Google",
  "osVersion": "Android 14",
  "appVersion": "2.1.0",
  "fcmToken": "dGhpcyBpcyBhIGZjIGtleTohZGkh",
  "firmware": "oriole-user 14 UP1A.231005.007...",
  "securityPatch": "2024-03-01",
  "buildId": "UP1A.231005.007",
  "status": "pending",
  "receivedAt": 1718900000000,
  "updatedAt": 1718900000000
}
```

**Error Responses:**
- `404 Not Found` - Entry doesn't exist

---

#### `POST /v1/device/inbox/:imei/ack`
**Purpose:** Device acknowledges receipt of registration request

**Request:**
```json
{
  "imei": "861234567890123",
  "status": "seen"
}
```

**Response (200 OK):**
```json
{
  "status": "acknowledged",
  "updatedAt": 1718900100000
}
```

---

#### `DELETE /v1/device/inbox/:imei`
**Purpose:** Operator dismisses inbox entry

**Response (200 OK):**
```json
{
  "status": "rejected",
  "updatedAt": 1718900200000
}
```

---

### 4.2 Device Endpoints

#### `POST /v1/device/register`
**Purpose:** Operator initiates device registration

**Request:**
```json
{
  "imei": "861234567890123"
}
```

**Response (200 OK):**
```json
{
  "status": "approving",
  "deviceId": "uuid-v4",
  "message": "Registration request sent to device"
}
```

**Error Responses:**
- `404 Not Found` - No pending inbox entry
- `409 Conflict` - Device already registered

---

#### `POST /v1/device/confirm`
**Purpose:** Device confirms registration, server saves

**Request:**
```json
{
  "imei": "861234567890123",
  "confirmed": true
}
```

**Response (200 OK):**
```json
{
  "status": "registered",
  "deviceId": "uuid-v4",
  "commandSecret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "registeredAt": 1718900300000
}
```

**Error Responses:**
- `400 Bad Request` - Invalid state transition
- `404 Not Found` - No approving entry found

---

#### `GET /v1/device`
**Purpose:** List all registered devices

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | string | all | Filter: online, offline |
| `page` | int | 1 | Page number |
| `limit` | int | 20 | Items per page |

**Response (200 OK):**
```json
{
  "devices": [
    {
      "id": "uuid-v4",
      "imei": "861234567890123",
      "deviceName": "Pixel 8 Pro",
      "model": "Pixel 8",
      "manufacturer": "Google",
      "osVersion": "Android 14",
      "appVersion": "2.1.0",
      "status": "online",
      "registeredAt": 1718900300000,
      "lastSeen": 1718900500000
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 12,
    "totalPages": 1
  }
}
```

---

#### `GET /v1/device/:imei`
**Purpose:** Get single device details

**Response (200 OK):**
```json
{
  "id": "uuid-v4",
  "imei": "861234567890123",
  "deviceName": "Pixel 8 Pro",
  "model": "Pixel 8",
  "manufacturer": "Google",
  "osVersion": "Android 14",
  "appVersion": "2.1.0",
  "fcmToken": "dGhpcyBpcyBhIGZjIGtleTohZGkh",
  "status": "online",
  "commandSecret": null,
  "registeredAt": 1718900300000,
  "lastSeen": 1718900500000,
  "createdAt": 1718900000000,
  "updatedAt": 1718900500000
}
```

---

#### `DELETE /v1/device/:imei`
**Purpose:** Deregister device

**Response (200 OK):**
```json
{
  "status": "deregistered",
  "message": "Device has been deregistered"
}
```

---

#### `GET /v1/device/:imei/telemetry`
**Purpose:** Get telemetry history

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `startTime` | int64 | -24h | Start timestamp (ms) |
| `endTime` | int64 | now | End timestamp (ms) |
| `limit` | int | 100 | Max results |

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
  "pagination": {
    "limit": 100,
    "hasMore": false
  }
}
```

---

## 5. GraphQL Schema

### 5.1 Types

```graphql
enum InboxStatus {
  PENDING
  ACKNOWLEDGED
  APPROVING
  REJECTED
  EXPIRED
}

enum DeviceStatus {
  ONLINE
  OFFLINE
  DEREGISTERED
}

type InboxEntry {
  id: ID!
  imei: String!
  deviceName: String!
  model: String
  manufacturer: String
  osVersion: String
  appVersion: String
  firmware: String
  securityPatch: String
  buildId: String
  status: InboxStatus!
  receivedAt: DateTime!
  updatedAt: DateTime!
}

type Device {
  id: ID!
  imei: String!
  deviceName: String!
  model: String
  manufacturer: String
  osVersion: String
  appVersion: String
  fcmToken: String
  status: DeviceStatus!
  registeredAt: DateTime
  lastSeen: DateTime
}

type TelemetryFrame {
  timestamp: DateTime!
  riskScore: Int
  thermalTemp: Float
  bufferLevel: Int
  uptime: Int
}

type PaginationInfo {
  page: Int!
  limit: Int!
  total: Int!
  totalPages: Int!
}
```

### 5.2 Queries

```graphql
type Query {
  # Inbox
  inboxEntries(
    status: InboxStatus
    page: Int = 1
    limit: Int = 20
  ): InboxConnection!
  
  inboxEntry(imei: String!): InboxEntry

  # Devices
  devices(
    status: DeviceStatus
    page: Int = 1
    limit: Int = 20
  ): DeviceConnection!
  
  device(imei: String!): Device

  # Telemetry
  deviceTelemetry(
    imei: String!
    startTime: Int
    endTime: Int
    limit: Int = 100
  ): TelemetryConnection!
}

type InboxConnection {
  entries: [InboxEntry!]!
  pagination: PaginationInfo!
}

type DeviceConnection {
  devices: [Device!]!
  pagination: PaginationInfo!
}

type TelemetryConnection {
  frames: [TelemetryFrame!]!
  pagination: PaginationInfo!
}
```

### 5.3 Mutations

```graphql
type Mutation {
  # Inbox mutations (device-facing)
  submitRegistrationRequest(
    input: RegistrationRequestInput!
  ): RegistrationRequestResponse!
  
  acknowledgeRequest(
    imei: String!
  ): AcknowledgeResponse!

  # Operator mutations
  registerDevice(
    imei: String!
  ): RegisterDeviceResponse!
  
  dismissInboxEntry(
    imei: String!
  ): DismissResponse!

  # Device mutations
  confirmRegistration(
    imei: String!
    confirmed: Boolean!
  ): ConfirmRegistrationResponse!

  # Device management
  deregisterDevice(
    imei: String!
  ): DeregisterResponse!
}

input RegistrationRequestInput {
  imei: String!
  deviceName: String!
  model: String
  manufacturer: String
  osVersion: String!
  appVersion: String!
  fcmToken: String!
  firmware: String
  securityPatch: String
  buildId: String
}

type RegistrationRequestResponse {
  success: Boolean!
  status: InboxStatus!
  messageId: String
  error: String
}

type AcknowledgeResponse {
  success: Boolean!
  status: InboxStatus!
  error: String
}

type RegisterDeviceResponse {
  success: Boolean!
  status: InboxStatus!
  deviceId: String
  message: String
  error: String
}

type DismissResponse {
  success: Boolean!
  status: InboxStatus!
  error: String
}

type ConfirmRegistrationResponse {
  success: Boolean!
  status: DeviceStatus!
  deviceId: String
  commandSecret: String
  registeredAt: DateTime
  error: String
}

type DeregisterResponse {
  success: Boolean!
  status: DeviceStatus!
  message: String
  error: String
}
```

### 5.4 Subscriptions

```graphql
type Subscription {
  inboxUpdated: InboxEntry!
  deviceStatusChanged: Device!
  telemetryReceived(imei: String!): TelemetryFrame!
}
```

---

## 6. Database Schema

### 6.1 Tables

#### `inbox_entries` (NEW)

```sql
CREATE TABLE inbox_entries (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    imei TEXT NOT NULL UNIQUE,
    device_name TEXT NOT NULL,
    model TEXT,
    manufacturer TEXT,
    os_version TEXT NOT NULL,
    app_version TEXT NOT NULL,
    fcm_token TEXT,
    firmware TEXT,
    security_patch TEXT,
    build_id TEXT,
    status TEXT NOT NULL DEFAULT 'pending' 
        CHECK (status IN ('pending', 'acknowledged', 'approving', 'rejected', 'expired')),
    received_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    expires_at INTEGER GENERATED ALWAYS AS (received_at + 30 * 24 * 60 * 60 * 1000) STORED,
    
    CONSTRAINT fk_device FOREIGN KEY (imei) REFERENCES devices(imei) ON DELETE CASCADE
);

CREATE INDEX idx_inbox_status ON inbox_entries(status);
CREATE INDEX idx_inbox_imei ON inbox_entries(imei);
CREATE INDEX idx_inbox_expires ON inbox_entries(expires_at) WHERE status = 'pending';
```

#### `devices` (EXISTING - needs update)

```sql
-- Add new columns if not exist
ALTER TABLE devices ADD COLUMN IF NOT EXISTS imei TEXT UNIQUE;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS device_name TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS model TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS manufacturer TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS os_version TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS firmware TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS security_patch TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS build_id TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS command_secret TEXT;

-- Update existing registration flow
ALTER TABLE devices ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'offline' 
    CHECK (status IN ('online', 'offline', 'deregistered'));
```

---

## 7. Frontend Components

### 7.1 Page Structure

```
/device
├── [Tabs]
│   ├── Overview     → DeviceOverview.tsx
│   ├── Telemetry    → DeviceTelemetry.tsx  
│   ├── Commands     → DeviceCommands.tsx
│   └── History      → DeviceHistory.tsx
│
├── [Sub-sections]
│   ├── InboxView    → DeviceInbox.tsx (inbox section at top)
│   └── RegisteredList → DeviceList.tsx (sidebar or modal)
```

### 7.2 Component List

| Component | File | Purpose | Lines Est |
|-----------|------|---------|----------|
| DeviceInbox | `components/device/DeviceInbox.tsx` | Inbox list with actions | ~300 |
| DeviceOverview | `components/device/DeviceOverview.tsx` | Device info + health | ~250 |
| DeviceTelemetry | `components/device/DeviceTelemetry.tsx` | Real-time + charts | ~400 |
| DeviceCommands | `components/device/DeviceCommands.tsx` | Send + pending + history | ~350 |
| DeviceHistory | `components/device/DeviceHistory.tsx` | Historical data + export | ~300 |
| DeviceList | `components/device/DeviceList.tsx` | All registered devices | ~250 |
| ConnectionStatus | `components/device/ConnectionStatus.tsx` | Enhanced status panel | ~150 |

### 7.3 Hooks

| Hook | File | Purpose |
|------|------|---------|
| `useInbox` | `hooks/use-inbox.ts` | Inbox entries query/mutations |
| `useDevices` | `hooks/use-devices.ts` | Device list query |
| `useDevice` | `hooks/use-device.ts` | Single device query |

---

## 8. File Changes Summary

### 8.1 NEW Files (Go Backend)

| File | Purpose |
|------|---------|
| `internal/api/handlers/inbox/inbox.go` | Inbox CRUD handlers |
| `internal/api/handlers/inbox/inbox_test.go` | Tests |
| `internal/api/handlers/inbox/routes.go` | Route registration |
| `internal/api/handlers/device/confirm.go` | Confirm endpoint |
| `internal/api/handlers/device/deregister.go` | Deregister endpoint |
| `internal/domain/device/inbox.go` | InboxEntry entity |
| `internal/domain/device/repository.go` | Repository interfaces |
| `internal/infrastructure/storage/inbox.go` | Inbox SQLite storage |
| `internal/infrastructure/storage/migrations/001_device_inbox.sql` | Migration |
| `internal/application/device/inbox_usecase.go` | Business logic |
| `internal/application/device/register_usecase.go` | Registration logic |
| `internal/application/dto/inbox.go` | DTOs |

### 8.2 MODIFIED Files (Go Backend)

| File | Changes |
|------|---------|
| `internal/api/handlers/device/register.go` | Update to support IMEI flow |
| `internal/api/handlers/device/status.go` | Add IMEI to response |
| `internal/api/handlers/device/list.go` | Add filters, IMEI |
| `internal/infrastructure/storage/device.go` | Add IMEI, status columns |
| `internal/domain/device/device.go` | Add IMEI fields |
| `internal/api/router.go` | Add inbox routes |
| `internal/api/middleware/auth.go` | Update auth checks |

### 8.3 NEW Files (Frontend)

| File | Purpose |
|------|---------|
| `src/components/device/DeviceInbox.tsx` | Inbox UI |
| `src/components/device/DeviceOverview.tsx` | Overview tab |
| `src/components/device/DeviceTelemetry.tsx` | Telemetry tab |
| `src/components/device/DeviceCommands.tsx` | Commands tab |
| `src/components/device/DeviceHistory.tsx` | History tab |
| `src/components/device/DeviceList.tsx` | Device list |
| `src/components/device/ConnectionStatus.tsx` | Status panel |
| `src/hooks/use-inbox.ts` | Inbox hook |
| `src/hooks/use-devices.ts` | Devices hook |
| `src/hooks/use-device.ts` | Single device hook |
| `src/lib/api/graphql/queries.ts` | Add inbox/device queries |
| `src/lib/api/graphql/mutations.ts` | Add mutations |
| `src/lib/api/graphql/types.ts` | Add types |
| `src/lib/api/rest/device-client.ts` | REST fallback |

### 8.4 MODIFIED Files (Frontend)

| File | Changes |
|------|---------|
| `src/routes/device.tsx` | Add tabs, inbox section |
| `src/lib/api/graphql/queries.ts` | Add device/inbox queries |
| `src/lib/api/graphql/mutations.ts` | Add mutations |
| `src/lib/device-stream-context.tsx` | Add IMEI support |
| `src/hooks/use-device-stream.ts` | Add metrics exposure |

---

## 9. Implementation Order

### Phase 1: Database & Storage (Day 1)
1. Create migration `001_device_inbox.sql`
2. Update `devices` table schema
3. Implement `internal/infrastructure/storage/inbox.go`
4. Implement `internal/infrastructure/storage/device.go` updates

### Phase 2: Domain & Application (Day 1-2)
1. Create `internal/domain/device/inbox.go`
2. Create `internal/domain/device/repository.go`
3. Implement `internal/application/device/inbox_usecase.go`
4. Implement `internal/application/device/register_usecase.go`

### Phase 3: REST API Handlers (Day 2-3)
1. Implement `POST /v1/device/inbox`
2. Implement `GET /v1/device/inbox`
3. Implement `GET /v1/device/inbox/:imei`
4. Implement `POST /v1/device/inbox/:imei/ack`
5. Implement `DELETE /v1/device/inbox/:imei`
6. Implement `POST /v1/device/register` (update)
7. Implement `POST /v1/device/confirm`
8. Implement `DELETE /v1/device/:imei`
9. Wire routes in `router.go`

### Phase 4: GraphQL (Day 3-4)
1. Define types in GraphQL schema
2. Implement resolver for inbox queries
3. Implement resolver for device queries
4. Implement mutations
5. Add subscriptions (optional)

### Phase 5: Frontend Core (Day 4-5)
1. Create `use-inbox` hook
2. Create `use-devices` hook
3. Create REST fallback client
4. Create `DeviceInbox` component

### Phase 6: Frontend Device Page (Day 5-6)
1. Update `device.tsx` with tabs
2. Create `DeviceOverview` component
3. Create `DeviceTelemetry` component
4. Create `DeviceCommands` component
5. Create `DeviceHistory` component

### Phase 7: Integration & Polish (Day 6-7)
1. Wire GraphQL queries to components
2. Add loading/error states
3. Add optimistic updates
4. Test end-to-end flow
5. Add to existing tests

---

## 10. Testing Strategy

### 10.1 Unit Tests
- Inbox use cases
- Device registration flow
- State transitions

### 10.2 Integration Tests
- Full registration flow
- Inbox state machine
- REST endpoints

### 10.3 E2E Tests
- Device sends request → Operator sees in inbox
- Operator registers → Device confirms
- Device deregisters → Removed from list

---

## 11. Rollout Checklist

- [ ] Migration run on production DB
- [ ] FCM credentials configured
- [ ] GraphQL schema deployed
- [ ] Frontend deployed
- [ ] APK updated with new endpoints
- [ ] Monitoring alerts configured
- [ ] Runbook updated

---

*Document Version: 1.0*  
*Status: Ready for Implementation*
