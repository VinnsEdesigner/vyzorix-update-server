# Device Registration System - Enterprise Implementation Specification

> **Version:** 1.1  
> **Status:** Draft  
> **Created:** 2026-06-21  
> **Updated:** 2026-06-24
> **Target:** Production MVP  
> **Architecture:** Layered (Following `FRONTEND_ARCHITECTURE.md`)

---

## Table of Contents

1. [Overview](#1-overview)
2. [Architecture](#2-architecture)
3. [State Machine](#3-state-machine)
4. [REST API Specification](#4-rest-api-specification)
5. [GraphQL Schema](#5-graphql-schema)
6. [Database Schema](#6-database-schema)
7. [Frontend Architecture](#7-frontend-architecture)
8. [Target File Structure](#8-target-file-structure)
9. [Domain Layer](#9-domain-layer)
10. [Data Layer](#10-data-layer)
11. [Presentation Layer - Hooks](#11-presentation-layer---hooks)
12. [UI Layer - Components](#12-ui-layer---components)
13. [File Changes Summary](#13-file-changes-summary)
14. [Implementation Order](#14-implementation-order)
15. [Testing Strategy](#15-testing-strategy)
16. [Rollout Checklist](#16-rollout-checklist)

---

---

> ⚠️ **Architecture Alignment Note (v1.1)**
> 
> This document has been updated to align with the **Layered Architecture** defined in `FRONTEND_ARCHITECTURE.md`. The file structure below follows the **4-layer architecture**:
> - **UI Layer** (`src/components/`) - Pure UI rendering, imports only from hooks
> - **Presentation Layer** (`src/hooks/`) - UI logic, state management, imports from domain & data
> - **Domain Layer** (`src/domain/` - NEW) - Types, transforms, validation (NO external imports)
> - **Data Layer** (`src/lib/api/`) - API clients (GraphQL/REST), imports only domain types
>
> **Dependency Rule:** UI → Hooks → Domain → API (flow inward only)


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
│       │   │   ├── device_entity.go               # Device entity
│       │   │   ├── device_inbox.go               # NEW: InboxEntry entity
│       │   │   ├── device_errors.go
│       │   │   └── device_repository.go          # Interfaces
│       │   │
│       │   └── command/
│       │
│       ├── infrastructure/
│       │   ├── storage/
│       │   │   ├── device_storage.go            # Existing
│       │   │   ├── inbox_storage.go             # NEW: Inbox storage
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
        │   └── device-page.tsx             # Update with tabs
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
      │  POST /v1/device/inbox/:imei/ack (action=acknowledge)
      ▼
┌──────────────┐   Device has seen the request
│ ACKNOWLEDGED│
└──────┬───────┘
       │
       │  POST /v1/device/inbox/:imei/ack (action=approve)
       ▼
┌──────────┐    Server validating, generating commandSecret, FCM push
│ APPROVING│
└──────┬───┘
       │
       │  (Automatic transition after commandSecret generation)
       ▼
┌────────────┐
│ APPROVED   │ ◄── Device can now call POST /v1/device/confirm
└─────┬──────┘
      │
      │  POST /v1/device/confirm (device confirms)
      ▼
┌────────────┐
│ REGISTERED │ ◄─────────────────────────────────────────────
└────────────┘
       │
       │  (on rejection at any step)
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

> **Note:** This implements the full 5-state model from SPEC. commandSecret is generated
> during the APPROVING state (intermediate) and sent to device via FCM push.

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
| `status` | string | all | Filter: pending, acknowledged, approved, rejected |
| `page` | int | 1 | Page number |
| `limit` | int | 20 | Items per page |

**Response (200 OK):**
```json
{
  "requests": [
    {
      "id": "inb_abc123",
      "imei": "861234567890123",
      "deviceName": "Pixel 8 Pro",
      "model": "Pixel 8",
      "manufacturer": "Google",
      "osVersion": "Android 14",
      "appVersion": "2.1.0",
      "fcmToken": "firebase_token_here",
      "firebaseInstallId": "firebase_install_id",
      "status": "pending",
      "createdAt": 1718900000000,
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
  "id": "inb_abc123",
  "imei": "861234567890123",
  "deviceName": "Pixel 8 Pro",
  "model": "Pixel 8",
  "manufacturer": "Google",
  "osVersion": "Android 14",
  "appVersion": "2.1.0",
  "fcmToken": "firebase_token_here",
  "firebaseInstallId": "firebase_install_id",
  "status": "pending",
  "createdAt": 1718900000000,
  "updatedAt": 1718900000000,
  "notes": null
}
```

**Error Responses:**
- `404 Not Found` - Entry doesn't exist

---

#### `POST /v1/device/inbox/:imei/ack`
**Purpose:** Handles inbox acknowledgement based on action type (5-state model)

**Request:**
```json
{
  "action": "acknowledge",  // Device acknowledges (PENDING -> ACKNOWLEDGED)
  "notes": "Optional notes"
}
```
OR
```json
{
  "action": "approve",  // Operator approves (ACKNOWLEDGED -> APPROVING -> APPROVED)
  "notes": "Optional operator notes"
}
```
OR
```json
{
  "action": "reject",  // Operator rejects (PENDING/ACKNOWLEDGED -> REJECTED)
  "notes": "Reason for rejection"
}
```

**Response (200 OK):**
```json
{
  "id": "entry_123",
  "imei": "861234567890123",
  "status": "acknowledged",  // or "approved", "rejected"
  "acknowledgedAt": 1718900100000,  // Only on acknowledge
  "approvedAt": 1718900200000,  // Only on approval
  "commandSecret": "abc123...",  // Only on approval
  "fcmPushSent": true,  // Whether FCM notification was sent
  "notes": "Optional operator notes"
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

#### `GET /v1/devices`
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
      "lastSeen": 1718900500000,
      "online": true
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

#### `GET /v1/devices/:imei`
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

#### `DELETE /v1/devices/:imei`
**Purpose:** Deregister device

**Response (200 OK):**
```json
{
  "imei": "861234567890123",
  "status": "deregistered",
  "deregisteredAt": 1718900500000,
  "retentionUntil": 1719505300000
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
        CHECK (status IN ('pending', 'acknowledged', 'approving', 'approved', 'rejected', 'expired')),
    acknowledged_at INTEGER,  -- When device acknowledged
    approving_at INTEGER,     -- When operator started approving
    approved_at INTEGER,      -- When fully approved
    rejected_at INTEGER,       -- When rejected
    received_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch() * 1000),
    expires_at INTEGER GENERATED ALWAYS AS (received_at + 30 * 24 * 60 * 60 * 1000) STORED,
    command_secret TEXT,      -- Generated during APPROVING state
    operator_id TEXT,         -- Operator who approved/rejected
    notes TEXT,
    
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

## 7. Frontend Architecture

### 7.1 Layered Architecture Overview

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
│  │    Custom hooks that:                                       │   │
│  │    - Handle UI logic                                        │   │
│  │    - Transform data for UI                                  │   │
│  │    - Manage state                                           │   │
│  │    NEVER renders UI. NEVER imports from UI layer.          │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              │ uses                                  │
│                              ▼                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                      DOMAIN LAYER                          │   │
│  │                     (src/domain/)                          │   │
│  │                                                             │   │
│  │    Pure functions that:                                     │   │
│  │    - Define types and interfaces                           │   │
│  │    - Transform data (no side effects)                     │   │
│  │    - Validate input                                         │   │
│  │    NEVER imports from UI, Presentation, or Data.            │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              │ uses                                  │
│                              ▼                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                       DATA LAYER                           │   │
│  │                   (src/lib/api/)                           │   │
│  │                                                             │   │
│  │    API clients that:                                       │   │
│  │    - Make HTTP requests                                    │   │
│  │    - Handle authentication                                  │   │
│  │    - Parse responses                                       │   │
│  │    NEVER imports from UI or Presentation.                  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 7.2 Dependency Rule

```
UI Layer ────────► Presentation Layer ────────► Domain ────────► Data Layer
(components/)          (hooks/)              (domain/)        (lib/api/)
     │                     │                      │                 │
     │                     │                      │                 │
     └── IMPORTS ─────────┘                      │                 │
          ONLY FROM                              │                 │
          HOOKS                                  │                 │
                                             │                 │
                                             └── IMPORTS ─────┘
                                                  ONLY FROM
                                                  DOMAIN TYPES
```

---

## 8. Target File Structure

### 8.1 Complete Directory Tree

```
apps/web/src/
│
├── domain/                              # DOMAIN LAYER (follows FRONTEND_ARCHITECTURE.md)
│   ├── _shared/                      # SHARED domain types
│   │   ├── domain-pagination.ts    # Pagination types & helpers
│   │   └── domain-errors.ts        # Domain error types
│   │
│   ├── registration/
│   │   ├── registration-entity.ts   # InboxEntry, InboxStatus
│   │   ├── registration-mappers.ts # inboxFromRaw(), deviceFromRaw()
│   │   └── registration-validators.ts # validateIMEI(), validateInboxEntry()
│   │
│   └── devices/
│       ├── device-entity.ts         # Device, DeviceStatus, ConnectionState
│       └── device-mappers.ts        # deviceFromRaw(), connectionFromRaw()
│
├── lib/
│   └── api/
│       ├── graphql/
│       │   ├── _shared/
│       │   │   └── graphql-client.ts  # GraphQL client setup
│       │   │
│       │   ├── registration/
│       │   │   ├── graphql-registration-queries.ts   # GET_INBOX, GET_INBOX_ENTRY
│       │   │   ├── graphql-registration-mutations.ts # ACK_INBOX, REGISTER, DEREGISTER
│       │   │   ├── graphql-registration-fragments.ts # InboxEntry & device fragments
│       │   │   └── graphql-registration-types.ts     # Raw GraphQL response types
│       │   │
│       │   └── devices/
│       │       ├── graphql-devices-queries.ts   # GET_DEVICES, GET_DEVICE
│       │       ├── graphql-devices-fragments.ts # Device fragments
│       │       └── graphql-devices-types.ts     # Raw GraphQL response types
│       │
│       └── rest/
│           ├── _shared/
│           │   └── rest-client.ts     # Base REST client
│           └── registration/
│               └── rest-registration-endpoints.ts  # REST fallback
│
├── hooks/                               # PRESENTATION LAYER
│   │
│   ├── registration/
│   │   ├── use-inbox.ts               # Get inbox entries
│   │   ├── use-inbox-entry.ts         # Single inbox entry
│   │   ├── use-ack-inbox.ts           # Acknowledge (approve/reject)
│   │   ├── use-devices.ts             # Device list
│   │   ├── use-device.ts              # Single device
│   │   ├── use-device-status.ts       # Device status + connection
│   │   ├── use-register-device.ts     # Register device
│   │   ├── use-deregister-device.ts   # Deregister device
│   │   └── index.ts                  # Barrel export
│   │
│   ├── devices/
│   │   ├── use-device-stream.ts       # (EXISTING - WebSocket)
│   │   ├── use-device-selected.ts     # Current selected device
│   │   └── index.ts
│   │
│   └── shared/
│       ├── use-pagination.ts           # Generic pagination
│       ├── use-search.ts              # Generic search
│       └── index.ts
│
├── components/
│   │
│   ├── shared/                        # SHARED UI COMPONENTS
│   │   ├── section.tsx               # Bordered section
│   │   ├── section-header.tsx         # Section header
│   │   ├── empty-state.tsx            # Empty state
│   │   ├── loading-skeleton.tsx       # Loading skeleton
│   │   ├── data-table.tsx            # Table wrapper
│   │   ├── pagination.tsx            # Pagination controls
│   │   ├── search-input.tsx           # Search input
│   │   ├── filter-select.tsx          # Filter dropdown
│   │   ├── status-badge.tsx           # (EXISTING - move here)
│   │   ├── connection-badge.tsx       # (EXISTING - move here)
│   │   └── index.ts                  # Barrel export
│   │
│   ├── registration/                  # REGISTRATION FEATURE
│   │   ├── device-inbox.tsx          # Main inbox list
│   │   ├── inbox-entry-row.tsx       # Single inbox row
│   │   ├── inbox-filters.tsx         # Status filter controls
│   │   ├── registration-actions.tsx   # Approve/Reject buttons
│   │   ├── device-list.tsx           # All registered devices
│   │   ├── device-card.tsx           # Device card for list
│   │   ├── device-overview.tsx        # Device overview tab
│   │   ├── device-telemetry.tsx       # Real-time + charts
│   │   ├── device-commands.tsx        # Send commands tab
│   │   ├── device-history.tsx         # Historical data tab
│   │   ├── connection-status.tsx       # Connection status panel
│   │   ├── device-tabs.tsx            # Tab navigation
│   │   └── index.ts                  # Barrel export
│   │
│   ├── layout/                       # (EXISTING)
│   │   ├── app-layout.tsx
│   │   └── auth-layout.tsx
│   │
│   ├── auth/                         # (EXISTING)
│   │   └── ... (existing auth components)
│   │
│   └── ui/                           # (EXISTING - base UI primitives)
│       ├── button.tsx
│       ├── badge.tsx
│       ├── card.tsx
│       └── ... (shadcn/ui components)
│
└── routes/                            # PAGE LAYER (Routes)
    │
    ├── __root.tsx                    # (EXISTING)
    ├── router.tsx                    # (EXISTING)
    │
    ├── device-page.tsx             # MODIFIED - add tabs structure
    ├── device.inbox.tsx             # NEW - /device/inbox (inbox view)
    ├── device.$imei.tsx             # NEW - /device/:imei (Overview tab)
    ├── device.$imei.telemetry.tsx  # NEW - /device/:imei/telemetry
    ├── device.$imei.history.tsx     # NEW - /device/:imei/history
    ├── device.$imei.commands.tsx   # NEW - /device/:imei/commands
    │
    ├── settings.tsx                 # (EXISTING)
    ├── dashboard.tsx                # (EXISTING)
    ├── diagnostics.tsx              # (EXISTING)
    ├── updates.tsx                  # (EXISTING)
    └── ... (other existing routes)
```

---

## 9. Domain Layer

### 9.1 Files to CREATE

| File | Purpose |
|------|---------|
| `domain/shared/pagination.ts` | Pagination types: `PaginationInfo`, `PaginatedResult<T>` |
| `domain/common/error.ts` | Domain errors: `DomainError`, `ValidationError` |
| `domain/shared/types.ts` | Shared types: `DateRange`, `FilterOptions` |
| `domain/registration/registration-types.ts` | InboxEntry, InboxStatus, RegistrationRequest, AckRequest |
| `domain/registration/registration-transforms.ts` | inboxFromRaw(), toRegistrationRequest() |
| `domain/registration/registration-validation.ts` | validateIMEI(), validateInboxEntry() |
| `domain/devices/device-types.ts` | Device, DeviceStatus, ConnectionState |
| `domain/devices/device-transforms.ts` | deviceFromRaw(), connectionFromRaw() |

### 9.2 Files to MODIFY

| File | Changes |
|------|---------|
| `domain/shared/types.ts` | Add if shared types already exist |

---

## 10. Data Layer

### 10.1 GraphQL - Queries to CREATE

| File | Queries |
|------|---------|
| `lib/api/graphql/queries/registration-queries.ts` | GET_INBOX, GET_INBOX_ENTRY, GET_INBOX_COUNT |
| `lib/api/graphql/queries/devices.ts` | GET_DEVICES, GET_DEVICE, GET_DEVICE_STATUS |

### 10.2 GraphQL - Mutations to CREATE

| File | Mutations |
|------|-----------|
| `lib/api/graphql/mutations/registration-mutations.ts` | ACK_INBOX, REGISTER_DEVICE, DEREGISTER_DEVICE |

### 10.3 GraphQL - Fragments to CREATE

| File | Fragment |
|------|----------|
| `lib/api/graphql/fragments/inbox-entry.fragment.ts` | InboxEntry fields |
| `lib/api/graphql/fragments/device.fragment.ts` | Device basic fields |

### 10.4 REST to CREATE

| File | Endpoints |
|------|-----------|
| `lib/api/rest/registration-rest.ts` | POST /inbox, GET /inbox, POST /inbox/:imei/ack, DELETE /inbox/:imei |

---

## 11. Presentation Layer - Hooks

### 11.1 Hooks to CREATE

| File | Hook | Purpose |
|------|------|---------|
| `hooks/registration/use-inbox.ts` | useInbox | Get inbox entries with filters |
| `hooks/registration/use-inbox-entry.ts` | useInboxEntry | Get single inbox entry |
| `hooks/registration/use-ack-inbox.ts` | useAckInbox | Acknowledge (approve/reject) |
| `hooks/registration/use-devices.ts` | useDevices | Get device list |
| `hooks/registration/use-device.ts` | useDevice | Get single device |
| `hooks/registration/use-device-status.ts` | useDeviceStatus | Get device status + connection |
| `hooks/registration/use-register-device.ts` | useRegisterDevice | Register device |
| `hooks/registration/use-deregister-device.ts` | useDeregisterDevice | Deregister device |
| `hooks/registration/index.ts` | | Barrel export |
| `hooks/devices/use-device-selected.ts` | useDeviceSelected | Get current selected device |
| `hooks/devices/index.ts` | | Barrel export |
| `hooks/shared/use-pagination.ts` | usePagination | Generic pagination |
| `hooks/shared/use-search.ts` | useSearch | Generic search |
| `hooks/shared/index.ts` | | Barrel export |

---

## 12. UI Layer - Components

### 12.1 Registration Components to CREATE

| File | Component | Purpose |
|------|-----------|---------|
| `components/registration/device-inbox.tsx` | DeviceInbox | Main inbox list |
| `components/registration/inbox-entry-row.tsx` | InboxEntryRow | Single inbox row |
| `components/registration/inbox-filters.tsx` | InboxFilters | Status filter controls |
| `components/registration/registration-actions.tsx` | RegistrationActions | Approve/Reject buttons |
| `components/registration/device-list.tsx` | DeviceList | All registered devices |
| `components/registration/device-card.tsx` | DeviceCard | Device card for list |
| `components/registration/device-overview.tsx` | DeviceOverview | Device overview tab |
| `components/registration/device-telemetry.tsx` | DeviceTelemetry | Real-time + charts |
| `components/registration/device-commands.tsx` | DeviceCommands | Send commands tab |
| `components/registration/device-history.tsx` | DeviceHistory | Historical data tab |
| `components/registration/connection-status.tsx` | ConnectionStatus | Status panel |
| `components/registration/device-tabs.tsx` | DeviceTabs | Tab navigation |
| `components/registration/index.ts` | | Barrel export |

### 12.2 Shared Components to CREATE

| File | Component | Purpose |
|------|-----------|---------|
| `components/shared/section.tsx` | Section | Bordered section |
| `components/shared/section-header.tsx` | SectionHeader | Section header |
| `components/shared/empty-state.tsx` | EmptyState | Empty state |
| `components/shared/loading-skeleton.tsx` | LoadingSkeleton | Loading placeholder |
| `components/shared/data-table.tsx` | DataTable | Table wrapper |
| `components/shared/pagination.tsx` | Pagination | Pagination controls |
| `components/shared/search-input.tsx` | SearchInput | Search input |
| `components/shared/filter-select.tsx` | FilterSelect | Filter dropdown |

### 12.3 Routes

| File | Purpose |
|------|---------|
| `routes/device-page.tsx` | **MODIFIED** - Add tabs: Inbox, Overview, Telemetry, History, Commands |
| `routes/device.inbox.tsx` | **NEW** - `/device/inbox` standalone inbox view |
| `routes/device.$imei.tsx` | **NEW** - `/device/:imei` device detail (Overview default) |
| `routes/device.$imei.telemetry.tsx` | **NEW** - `/device/:imei/telemetry` real-time charts |
| `routes/device.$imei.history.tsx` | **NEW** - `/device/:imei/history` historical data |
| `routes/device.$imei.commands.tsx` | **NEW** - `/device/:imei/commands` send commands |

**Total: 1 MODIFIED, 5 NEW (6 routes)**

### 12.4 Device Commands Reference

The following commands are supported by the device APK and sent via the Commands tab:

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

> **Note:** All commands are HMAC-SHA256 signed. See `COMMAND_SECURITY.md` for signing specification.

---

## 13. File Changes Summary

### 13.1 Total File Count

| Category | New Files | Modified Files |
|----------|-----------|----------------|
| Domain Layer | 8 | 1 |
| Data Layer (GraphQL) | 5 | 1 |
| Data Layer (REST) | 1 | 1 |
| Presentation Layer | 14 | 0 |
| UI Layer (Shared) | 8 | 0 |
| UI Layer (Registration) | 13 | 0 |
| Routes | 5 | 1 |
| **TOTAL** | **51** | **4** |

### 13.2 All Files Listed

#### Domain Layer (8 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| `domain/shared/pagination.ts` | **NEW** | Pagination types |
| `domain/common/error.ts` | **NEW** | Domain error types |
| `domain/shared/types.ts` | **NEW** | Shared types |
| `domain/registration/registration-types.ts` | **NEW** | Registration types |
| `domain/registration/registration-transforms.ts` | **NEW** | Transforms |
| `domain/registration/registration-validation.ts` | **NEW** | Validation |
| `domain/devices/device-types.ts` | **NEW** | Device types |
| `domain/devices/device-transforms.ts` | **NEW** | Device transforms |
| `domain/shared/types.ts` | MODIFIED | Add shared types |

#### Data Layer - GraphQL (5 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| `lib/api/graphql/queries/registration-queries.ts` | **NEW** | Registration queries |
| `lib/api/graphql/queries/devices.ts` | **NEW** | Device queries |
| `lib/api/graphql/mutations/registration-mutations.ts` | **NEW** | Registration mutations |
| `lib/api/graphql/fragments/inbox-entry.fragment.ts` | **NEW** | Inbox fragment |
| `lib/api/graphql/fragments/device.fragment.ts` | **NEW** | Device fragment |
| `lib/api/graphql/queries.ts` | MODIFIED | Add registration queries |
| `lib/api/graphql/mutations.ts` | MODIFIED | Add registration mutations |

#### Data Layer - REST (1 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| `lib/api/rest/registration-rest.ts` | **NEW** | REST fallback |
| `lib/api/rest/endpoints.ts` | MODIFIED | Add registration endpoints |

#### Presentation Layer (14 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `hooks/registration/use-inbox.ts` | **NEW** | Inbox query |
| `hooks/registration/use-inbox-entry.ts` | **NEW** | Single entry query |
| `hooks/registration/use-ack-inbox.ts` | **NEW** | Acknowledge |
| `hooks/registration/use-devices.ts` | **NEW** | Device list |
| `hooks/registration/use-device.ts` | **NEW** | Single device |
| `hooks/registration/use-device-status.ts` | **NEW** | Device status |
| `hooks/registration/use-register-device.ts` | **NEW** | Register |
| `hooks/registration/use-deregister-device.ts` | **NEW** | Deregister |
| `hooks/registration/index.ts` | **NEW** | Barrel export |
| `hooks/devices/use-device-selected.ts` | **NEW** | Selected device |
| `hooks/devices/index.ts` | **NEW** | Barrel export |
| `hooks/shared/use-pagination.ts` | **NEW** | Pagination |
| `hooks/shared/use-search.ts` | **NEW** | Search |
| `hooks/shared/index.ts` | **NEW** | Barrel export |

#### UI Layer - Shared (8 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `components/shared/section.tsx` | **NEW** | Section |
| `components/shared/section-header.tsx` | **NEW** | Section header |
| `components/shared/empty-state.tsx` | **NEW** | Empty state |
| `components/shared/loading-skeleton.tsx` | **NEW** | Skeleton |
| `components/shared/data-table.tsx` | **NEW** | Table |
| `components/shared/pagination.tsx` | **NEW** | Pagination |
| `components/shared/search-input.tsx` | **NEW** | Search input |
| `components/shared/filter-select.tsx` | **NEW** | Filter select |

#### UI Layer - Registration (13 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `components/registration/device-inbox.tsx` | **NEW** | Inbox list |
| `components/registration/inbox-entry-row.tsx` | **NEW** | Inbox row |
| `components/registration/inbox-filters.tsx` | **NEW** | Filters |
| `components/registration/registration-actions.tsx` | **NEW** | Approve/Reject |
| `components/registration/device-list.tsx` | **NEW** | Device list |
| `components/registration/device-card.tsx` | **NEW** | Device card |
| `components/registration/device-overview.tsx` | **NEW** | Overview |
| `components/registration/device-telemetry.tsx` | **NEW** | Telemetry |
| `components/registration/device-commands.tsx` | **NEW** | Commands |
| `components/registration/device-history.tsx` | **NEW** | History |
| `components/registration/connection-status.tsx` | **NEW** | Connection |
| `components/registration/device-tabs.tsx` | **NEW** | Tabs |
| `components/registration/index.ts` | **NEW** | Barrel export |

#### Routes (2 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| `routes/device-page.tsx` | **MODIFIED** | Add 5 tabs: Inbox, Overview, Telemetry, History, Commands |
| `routes/device.inbox.tsx` | **NEW** | `/device/inbox` - Inbox view |
| `routes/device.$imei.tsx` | **NEW** | `/device/:imei` - Device detail |
| `routes/device.$imei.telemetry.tsx` | **NEW** | `/device/:imei/telemetry` - Telemetry |
| `routes/device.$imei.history.tsx` | **NEW** | `/device/:imei/history` - History |
| `routes/device.$imei.commands.tsx` | **NEW** | `/device/:imei/commands` - Commands |

**Total: 6 routes (1 MODIFIED, 5 NEW)** |
## 14. Implementation Order

### Phase 1: Database & Storage (Day 1)
1. Create migration `001_device_inbox.sql`
2. Update `devices` table schema
3. Implement `internal/infrastructure/storage/inbox_storage.go`
4. Implement `internal/infrastructure/storage/device_storage.go` updates

### Phase 2: Domain & Application (Day 1-2)
1. Create `internal/domain/device/device_inbox.go`
2. Create `internal/domain/device/device_repository.go`
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
1. Update `device-page.tsx` with tabs
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

## 15. Testing Strategy

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

## 16. Rollout Checklist

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
