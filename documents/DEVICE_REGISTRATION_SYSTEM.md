# Device Registration System - Enterprise Implementation Specification

> **Version:** 2.0  
> **Status:** Aligned with Organization Model  
> **Created:** 2026-06-21  
> **Updated:** 2026-08-15
> **Target:** Production MVP  
> **Architecture:** Layered (Following `FRONTEND_ARCHITECTURE.md`)
> **Source of Truth:** API Client (`packages/API_Client/`) — implemented & authoritative

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

>  **Architecture & Organization Alignment Note (v2.0)**
> 
> This document has been realigned to the **Organization-scoped model** and the **implemented API Client**.
> The original v1.x specs were written *before* the organization model was designed, so they described
> a single-tenant, flat device registry. All registration data is now **organization-scoped**:
> 
> - Every inbox entry, registered device, and registration action belongs to an **organization**.
> - The API Client injects the active organization on every call via the `X-Organization-ID` header
>   (REST) and the `organizationId` GraphQL variable, sourced from `useCurrentOrganizationId()`.
> - Devices carry an `organization_id` field; inbox requests are created/approved within an org context.
> 
> **Layered Architecture** (unchanged from `FRONTEND_ARCHITECTURE.md`):
> - **UI Layer** (`apps/VyzoriX_web/src/ui/`) - Pure UI rendering, imports only from hooks
> - **Presentation Layer** (`apps/VyzoriX_web/src/hooks/`) - UI logic, state management
> - **Domain Layer** (`packages/API_Client/src/domain/`) - Types, mappers, validation (NO external imports)
> - **Data Layer** (`packages/API_Client/src/vyzorServer/`) - API clients (REST primary / GraphQL), imports only domain types
> 
> **Dependency Rule:** UI → Hooks → Domain → Data (flow inward only)
> 
> **Data layer policy:** REST is the **primary** data layer for registration (fully implemented in
> `vyzorServer/rest/registration/registration-endpoints.ts`). GraphQL registration coverage is partial
> (inbox queries + ack/deregister mutations only); use REST as the source of truth for the API contract.


## 1. Overview

### 1.1 Problem Statement

Current device registration requires:
- Manual form entry on dashboard
- User manually types Firebase IDs, device class
- Single device focus
- No visibility into pending registrations
- No machine-to-machine handshake

### 1.2 Solution

Zero-friction, **organization-scoped** device registration with:
- Device auto-reports on first boot (creates an inbox request)
- Operator reviews pending requests in the organization's inbox
- Machine-to-machine confirmation handshake via `commandSecret`
- Multi-device, multi-organization support
- Full audit trail (`operator_id`, `notes`, status timestamps)

### 1.3 Flow Summary

```

  DEVICE (APK)                    SERVER                       DASHBOARD (operator)

  1. APK starts, fetches device info (IMEI, OS, app version, fcmToken, firebaseInstallId)
  2. POST /v1/device/inbox          (device sends registration request)
     → scoped to the device's target organization (X-Organization-ID)
  3. Server stores in INBOX         (status: pending)
  4. Operator views inbox           GET /v1/device/inbox  (organization-scoped)
     → clicks "Approve"
  5. POST /v1/inbox/:imei/ack {action:"approve"}
     Server validates, generates commandSecret, sends FCM push to device
     (status: pending → acknowledged → approving → approved)
  6. Device receives FCM push, stores commandSecret
     POST /v1/device/confirm {imei, commandSecret}
  7. Server validates commandSecret, moves entry to DEVICES table
     (status: registered)

```

> **Organization scope:** Steps 2-5 are performed within an organization context. The operator's
> active organization (from `useCurrentOrganizationId()`) determines which inbox entries and
> devices are visible. A device always registers into exactly one organization.

---

## 2. System Architecture

### 2.1 Components

```
                         FRONTEND (React) — organization-scoped

    DevicePage     InboxView     CommandsTab
       │               │              │
       └────────┬──────┴──────────────┘
                │  hooks (use-*.ts — one per hook) — inject organizationId
                │
       ┌────────┴────────┐
       │  REST Client     │  (primary)   ← X-Organization-ID header
       │  (registration)  │
       └────────┬────────┘
                │  (GraphQL — partial: inbox queries + ack/deregister)
                │
                ▼
                       SERVER (Go) — Gin
                │
       ┌────────┴─────────────────────────┐
       │  InboxHandler   DevicesHandler    │
       │  (ack/approve)  (list/detail)     │
       └────────┬─────────────────────────┘
                │
       ┌────────┴────────┐
       │  Inbox Store     │  Device Store
       └────────┬────────┘
                ▼
            Database (org-partitioned: inbox_requests, devices, registration_logs)
```

> **Organization boundary:** The active organization flows from `authStore.organizationId`
> → `useCurrentOrganizationId()` → REST `X-Organization-ID` header / GraphQL `organizationId`
> variable → server middleware → org-scoped queries. A device registers into exactly one org.

### 2.2 Directory Structure (high-level — see §8 for the full tree)

```
packages/API_Client/src/                  # shared domain + data layers (web + mobile)
  domain/registration/                    # InboxEntry, InboxStatus, mappers, validators
  domain/device/                          # Device entity (detail page)
  domain/organization/                    # Organization model
  vyzorServer/rest/registration/          # registration REST client (PRIMARY)
  vyzorServer/graphql/registration/       # GraphQL queries/mutations (PARTIAL)

apps/VyzoriX_web/src/
  hooks/registration/use-*.ts             # 12 registration hooks (one per file, REST-primary + GraphQL fallback)
  hooks/_shared/use-current-context.ts    # useCurrentOrganizationId()
  hooks/organization/                     # active org selection
  ui/pages/device/                        # device page (TODO)
  routes/                                 # route files (TODO)
  stores/                                 # authStore.organizationId

apps/api/internal/                        # Go server (see SERVER_BACKEND_DEVICE_REGISTRATION_API.md)
  api/handlers/{inbox,device}/
  application/{inbox,device}/
  domain/{inbox,device}/
  infrastructure/storage/
```

> Server-side file structure is authoritative in `SERVER_BACKEND_DEVICE_REGISTRATION_API.md` §6.

---

## 3. State Machine

### 3.1 Inbox States

```

                         INBOX STATE MACHINE                         


UNREGISTERED 
     
       POST /v1/device/inbox (device sends request)
     
     FCM sent to device, awaiting acknowledgment
  PENDING  

      
        POST /v1/device/inbox/:imei/ack (action=acknowledge)
      
   Device has seen the request
 ACKNOWLEDGED

       
         POST /v1/device/inbox/:imei/ack (action=approve)
       
    Server validating, generating commandSecret, FCM push
 APPROVING

       
         (Automatic transition after commandSecret generation)
       

 APPROVED     Device can now call POST /v1/device/confirm

      
        POST /v1/device/confirm (device confirms)
      

 REGISTERED  

       
         (on rejection at any step)
       

 REJECTED    Operator clicks Dismiss

       
         (auto-cleanup after 30 days)
       

 EXPIRED     No action taken

```

> **Note:** This implements the full 6-state model from the API Client domain
> (`domain/registration/registration-entity.ts`). `commandSecret` is generated
> during the APPROVING state (intermediate) and sent to the device via FCM push.
>
> **Server alignment:** `SERVER_BACKEND_DEVICE_REGISTRATION_API.md` currently documents a simpler
> 3-state model (`PENDING → APPROVED → REJECTED`). The API Client and this frontend spec use the
> full 6-state model. The server is expected to converge on the 6-state model; until then the
> `acknowledge`/`approving` intermediate states are surfaced by the client but may be collapsed
> server-side. All states are **organization-scoped**.

### 3.2 Device States

```

                         DEVICE STATE MACHINE                        


REGISTERED 
     
       First telemetry OR WebSocket connect
     

 ONLINE   

          WebSocket disconnect OR no telemetry for 30s
     

 OFFLINE 

          WebSocket reconnect OR telemetry received
     

 ONLINE  

     
       DELETE /v1/device/:imei (operator deregisters)
     

 DEREGISTERED   Terminal state

```

---

## 4. REST API Specification

> **Source of truth:** `packages/API_Client/src/vyzorServer/rest/registration/registration-endpoints.ts`.
> REST is the **primary** data layer. All endpoints are **organization-scoped**: the active org is
> sent as the `X-Organization-ID` header (injected by the REST client interceptor) and/or an
> `organization_id` query param by the API Client. Operators obtain the org id from
> `useCurrentOrganizationId()`; devices receive it at provisioning time.

### 4.1 Inbox Endpoints

#### `POST /v1/device/inbox`
**Purpose:** Device sends registration request (creates an inbox entry).

**Headers:** `X-Organization-ID: <org-id>` (device's target organization)
**Request:**
```json
{
  "imei": "861234567890123",
  "deviceName": "Pixel 8 Pro",
  "deviceClass": "phone",
  "model": "Pixel 8",
  "manufacturer": "Google",
  "osVersion": "Android 14",
  "appVersion": "2.1.0",
  "fcmToken": "dGhpcyBpcyBhIGZjIGtleTohZGkh",
  "firebaseInstallId": "firebase_install_id",
  "idempotencyKey": "uuid-v4"
}
```

**Response (201 Created):**
```json
{
  "id": "inb_abc123",
  "imei": "861234567890123",
  "status": "pending",
  "createdAt": 1718900000000
}
```

**Error Responses:**
- `400 Bad Request` - Missing required fields (`imei`, `fcmToken`, `firebaseInstallId`)
- `409 Conflict` - IMEI already in inbox or registered

---

#### `GET /v1/device/inbox`
**Purpose:** List all inbox entries for the active organization (operator).

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | string | all | Filter: pending, acknowledged, approving, approved, rejected, expired, all |
| `page` | int | 1 | Page number |
| `limit` | int | 20 | Items per page (max 100) |
| `organization_id` | string | — | Organization scope (also sent via `X-Organization-ID` header) |

**Response (200 OK):**
```json
{
  "requests": [
    {
      "id": "inb_abc123",
      "imei": "861234567890123",
      "deviceName": "Pixel 8 Pro",
      "deviceClass": "phone",
      "model": "Pixel 8",
      "manufacturer": "Google",
      "osVersion": "Android 14",
      "appVersion": "2.1.0",
      "fcmToken": "firebase_token_here",
      "firebaseInstallId": "firebase_install_id",
      "status": "pending",
      "acknowledgedAt": null,
      "approvingAt": null,
      "approvedAt": null,
      "rejectedAt": null,
      "notes": null,
      "operatorId": null,
      "createdAt": 1718900000000
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

#### `GET /v1/inbox/:imei`
**Purpose:** Get single inbox entry details (organization-scoped).

> **Path note:** The API Client resolves this to `/v1/inbox/${imei}` (not `/v1/device/inbox/:imei`).

**Response (200 OK):**
```json
{
  "id": "inb_abc123",
  "imei": "861234567890123",
  "deviceName": "Pixel 8 Pro",
  "deviceClass": "phone",
  "model": "Pixel 8",
  "manufacturer": "Google",
  "osVersion": "Android 14",
  "appVersion": "2.1.0",
  "fcmToken": "firebase_token_here",
  "firebaseInstallId": "firebase_install_id",
  "status": "pending",
  "acknowledgedAt": null,
  "approvingAt": null,
  "approvedAt": null,
  "rejectedAt": null,
  "notes": null,
  "operatorId": null,
  "createdAt": 1718900000000
}
```

**Error Responses:**
- `404 Not Found` - Entry doesn't exist (or not in the active organization)

---

#### `POST /v1/inbox/:imei/ack`
**Purpose:** Acknowledge / approve / reject an inbox entry (5-action model).

> **Path note:** The API Client resolves this to `/v1/inbox/${imei}/ack`.

**Request:**
```json
{
  "action": "acknowledge",  // PENDING -> ACKNOWLEDGED  (device acknowledges)
  "notes": "Optional notes"
}
```
OR
```json
{
  "action": "approve",      // ACKNOWLEDGED -> APPROVING -> APPROVED (operator approves)
  "notes": "Optional operator notes"
}
```
OR
```json
{
  "action": "reject",       // PENDING/ACKNOWLEDGED -> REJECTED (operator rejects)
  "notes": "Reason for rejection"
}
```

**Response (200 OK):**
```json
{
  "id": "inb_abc123",
  "imei": "861234567890123",
  "status": "approved",          // or "acknowledged", "rejected"
  "acknowledgedAt": 1718900100000,  // present when acknowledged
  "approvingAt": 1718900150000,     // present when approving
  "approvedAt": 1718900200000,      // present when approved
  "rejectedAt": null,               // present when rejected
  "commandSecret": "abc123...",     // only on approve
  "fcmPushSent": true,              // whether FCM notification was sent
  "notes": "Optional operator notes"
}
```

---

#### `POST /v1/inbox/:imei/resend`
**Purpose:** Re-send the FCM approval push for an approved entry (e.g. device missed the first push).

> **Path note:** The API Client resolves this to `/v1/inbox/${imei}/resend`.
> Not present in the original v1.x spec; added to match the implemented client.

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Approval push re-sent"
}
```

---

#### `DELETE /v1/inbox/:imei`
**Purpose:** Operator dismisses an inbox entry.

> **Path note:** The API Client resolves this to `/v1/inbox/${imei}`.

**Response (200 OK):**
```json
{
  "status": "rejected"
}
```

---

### 4.2 Device Endpoints

#### `POST /v1/device/register` (device-initiated, alternative entry point)
**Purpose:** Device-initiated registration (alternative to the inbox flow). Used by the
`DeviceClient` (`vyzorServer/device/device-client.ts`) for direct registration when an operator
is not involved.

**Request:**
```json
{
  "imei": "861234567890123",
  "name": "Pixel 8 Pro"
}
```

**Response (200 OK):**
```json
{
  "commandSecret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}
```

> **Note:** The operator-driven flow uses the inbox `ack` endpoint (approve action) to generate
> `commandSecret` instead. See `SERVER_BACKEND_DEVICE_REGISTRATION_API.md` §4.3.

---

#### `POST /v1/device/confirm`
**Purpose:** Device confirms registration by presenting the `commandSecret` it received via FCM.
The server validates the secret and moves the entry into the `devices` table.

**Request:**
```json
{
  "imei": "861234567890123",
  "commandSecret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}
```

**Response (200 OK):**
```json
{
  "device_id": "uuid-v4",
  "imei": "861234567890123",
  "confirmed": true,
  "online": true,
  "registered_at": 1718900300000,
  "server_time": 1718900300123
}
```

> **Field note:** The raw server response uses snake_case (`device_id`, `registered_at`,
> `server_time`); the API Client maps it to `ConfirmDeviceResult` (`deviceId`, `registeredAt`,
> `serverTime`, `confirmed`, `online`). There is **no** `commandSecret` in the confirm response —
> the secret is supplied by the device in the request.

**Error Responses:**
- `400 Bad Request` - Invalid `commandSecret` or invalid state transition
- `404 Not Found` - No approved inbox entry found for this IMEI

---

#### `GET /v1/devices`
**Purpose:** List all registered devices for the active organization.

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | string | all | Filter: online, offline, all |
| `page` | int | 1 | Page number |
| `limit` | int | 20 | Items per page (max 100) |
| `organization_id` | string | — | Organization scope (also via `X-Organization-ID` header) |

**Response (200 OK):**
```json
{
  "devices": [
    {
      "id": "dev_abc123",
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
**Purpose:** Get single registered device details (organization-scoped).

**Response (200 OK):**
```json
{
  "id": "dev_abc123",
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
```

> **Full Device entity:** The richer `Device` type (`domain/device/device-entity.ts`) — used by the
> device detail page, not the registration list — additionally exposes `organization_id`,
> `security_patch`, `fcm_token_valid`, `command_secret_set`, `connection` (WebSocket status,
> protocol, client IP), `created_at`, and `updated_at`. See `SERVER_BACKEND_DEVICE_REGISTRATION_API.md` §4.2.

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
**Purpose:** Get telemetry history (organization-scoped).

> **Ownership note:** Telemetry is owned by the **telemetry context**
> (`packages/API_Client/src/domain/telemetry/`, `vyzorServer/rest/telemetry/`), not the
> registration context. It is listed here because the device page's Telemetry tab consumes it.
> See the telemetry spec for authoritative details.

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `startTime` | int64 | -24h | Start timestamp (ms) |
| `endTime` | int64 | now | End timestamp (ms) |
| `limit` | int | 100 | Max results |
| `organization_id` | string | — | Organization scope (via `X-Organization-ID`) |

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

> **Status:** Partial. GraphQL registration coverage in the API Client is limited to inbox queries
> (`GET_INBOX_ENTRIES`, `GET_INBOX_ENTRY`) and two mutations (`ACK_INBOX`, `DEREGISTER_DEVICE`).
> The REST client is the source of truth; the schema below reflects the implemented queries plus
> the intended server schema from `SERVER_BACKEND_DEVICE_REGISTRATION_API.md` §9, now
> **organization-scoped**. All root queries/mutations take `organizationId` as the first argument.

### 5.1 Types

```graphql
enum InboxStatus {
  PENDING
  ACKNOWLEDGED
  APPROVING
  APPROVED
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
  deviceClass: String
  model: String
  manufacturer: String
  osVersion: String
  appVersion: String
  fcmToken: String
  firebaseInstallId: String
  status: InboxStatus!
  acknowledgedAt: DateTime
  approvingAt: DateTime
  approvedAt: DateTime
  rejectedAt: DateTime
  notes: String
  operatorId: String
  createdAt: DateTime!
}

type Device {
  id: ID!
  imei: String!
  organizationId: ID!
  deviceName: String
  model: String
  manufacturer: String
  osVersion: String
  appVersion: String
  securityPatch: String
  status: DeviceStatus!
  registeredAt: DateTime
  lastSeen: DateTime
  fcmTokenValid: Boolean!
  commandSecretSet: Boolean!
  connection: DeviceConnection
}

type DeviceConnection {
  webSocketStatus: String!
  connectedAt: DateTime
  protocol: String
  clientIp: String
}

type TelemetryFrame {
  timestamp: DateTime!
  riskScore: Int
  thermalTemp: Float
  bufferLevel: Int
  uptime: Int
}

type Pagination {
  page: Int!
  limit: Int!
  total: Int!
  totalPages: Int!
}
```

### 5.2 Queries

```graphql
type Query {
  # Inbox (organization-scoped)
  inbox(
    organizationId: ID!
    status: InboxStatus
    page: Int = 1
    limit: Int = 20
  ): InboxConnection!

  inboxEntry(
    organizationId: ID!
    imei: String!
  ): InboxEntry

  # Devices (organization-scoped)
  devices(
    organizationId: ID!
    status: DeviceStatus
    search: String
    page: Int = 1
    limit: Int = 20
  ): DeviceListConnection!

  device(
    organizationId: ID!
    id: ID!
  ): Device

  # Telemetry (owned by the telemetry context, shown for reference)
  telemetryHistory(
    organizationId: ID!
    deviceId: ID!
    limit: Int = 100
  ): TelemetryConnection!
}

type InboxConnection {
  requests: [InboxEntry!]!
  pagination: Pagination!
}

type DeviceListConnection {
  devices: [Device!]!
  pagination: Pagination!
}

type TelemetryConnection {
  frames: [TelemetryFrame!]!
  pagination: Pagination!
}
```

### 5.3 Mutations

```graphql
enum AckAction {
  APPROVE
  REJECT
}

type Mutation {
  # Operator inbox actions (organization-scoped)
  ackInbox(
    imei: String!
    action: AckAction!
    notes: String
  ): AckResult!

  # Device management
  deregisterDevice(
    imei: String!
    hard: Boolean
  ): DeregisterResult!
}

type AckResult {
  id: ID!
  imei: String!
  status: InboxStatus!
  acknowledgedAt: DateTime
  approvingAt: DateTime
  approvedAt: DateTime
  rejectedAt: DateTime
  commandSecret: String
  fcmPushSent: Boolean
  notes: String
}

type DeregisterResult {
  imei: String!
  status: String!
  deregisteredAt: DateTime!
  retentionUntil: DateTime!
}
```

> **Implemented mutations** (API Client `graphql-registration-mutations.ts`):
> `ACK_INBOX` (`ackInbox(imei, action: APPROVE|REJECT, notes)`) and `DEREGISTER_DEVICE`
> (`deregisterDevice(imei, hard)`). The `organizationId` is supplied as a variable alongside the
> mutation arguments (see `graphql-registration-queries.ts`).

### 5.4 Subscriptions

```graphql
type Subscription {
  inboxUpdated(organizationId: ID!): InboxEntry!
  deviceStatusChanged(organizationId: ID!): Device!
  telemetryReceived(imei: String!): TelemetryFrame!
}
```

> **Note:** Registration subscriptions are not yet implemented in the API Client; the web app uses
> React Query cache invalidation on mutations instead.

---

## 6. Database Schema

> **Server authority:** `SERVER_BACKEND_DEVICE_REGISTRATION_API.md` §5 defines the canonical
> schema (`inbox_requests`, `devices`, `registration_logs`). The schema below mirrors it with the
> 6-state status model and the **organization_id** partitioning column added. All registration
> tables are partitioned by organization; queries filter on `organization_id` (sourced from the
> `X-Organization-ID` header).

### 6.1 Tables

#### `inbox_requests` (NEW)

```sql
CREATE TABLE inbox_requests (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    organization_id TEXT NOT NULL,        -- Organization scope (NEW)
    imei TEXT NOT NULL,
    device_name TEXT,
    device_class TEXT,                    -- NEW: phone, tablet, etc.
    model TEXT,
    manufacturer TEXT,
    os_version TEXT,
    app_version TEXT,
    fcm_token TEXT NOT NULL,
    firebase_install_id TEXT NOT NULL,    -- NEW: required by API Client
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'acknowledged', 'approving', 'approved', 'rejected', 'expired')),
    acknowledged_at TIMESTAMPTZ,          -- When device acknowledged
    approving_at TIMESTAMPTZ,             -- When operator started approving
    approved_at TIMESTAMPTZ,              -- When fully approved
    rejected_at TIMESTAMPTZ,              -- When rejected
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    command_secret TEXT,                  -- Generated during APPROVING state
    operator_id TEXT,                     -- Operator who approved/rejected
    notes TEXT,

    CONSTRAINT uq_org_imei UNIQUE (organization_id, imei),
    CONSTRAINT fk_organization FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE INDEX idx_inbox_pending ON inbox_requests(organization_id, status, created_at DESC);
CREATE INDEX idx_inbox_imei ON inbox_requests(organization_id, imei);
```

> **Naming note:** The server spec calls this table `inbox_requests`; the v1.x frontend spec called
> it `inbox_entries`. The server name is authoritative. The `firmware`, `security_patch`,
> `build_id`, `received_at`, `updated_at`, and `expires_at` columns from v1.x are **dropped** —
> the API Client domain model does not surface them for registration.

#### `devices` (EXISTING - needs update)

```sql
-- Add new columns if not exist
ALTER TABLE devices ADD COLUMN IF NOT EXISTS organization_id TEXT;   -- NEW: org partitioning
ALTER TABLE devices ADD COLUMN IF NOT EXISTS imei TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS device_name TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS model TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS manufacturer TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS os_version TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS app_version TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS security_patch TEXT;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS command_secret_hash TEXT;  -- hashed, not plaintext
ALTER TABLE devices ADD COLUMN IF NOT EXISTS deregistered_at TIMESTAMPTZ;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS deletion_scheduled_at TIMESTAMPTZ;

-- Status (online/offline/deregistered)
ALTER TABLE devices ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'offline'
    CHECK (status IN ('online', 'offline', 'deregistered'));

-- Organization-scoped uniqueness + search index
CREATE UNIQUE INDEX IF NOT EXISTS uq_devices_org_imei ON devices(organization_id, imei);
CREATE INDEX IF NOT EXISTS idx_devices_name ON devices(organization_id, device_name);

ALTER TABLE devices ADD CONSTRAINT fk_device_org
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;
```

#### `registration_logs` (NEW — audit trail)

```sql
CREATE TABLE registration_logs (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    organization_id TEXT NOT NULL,
    device_id TEXT,
    action TEXT NOT NULL,                 -- submit, acknowledge, approve, reject, confirm, deregister
    operator_id TEXT,
    details JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_device FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE SET NULL,
    CONSTRAINT fk_org FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE INDEX idx_registration_logs ON registration_logs(organization_id, device_id, created_at DESC);
```

---

## 7. Frontend Architecture

### 7.1 Layered Architecture Overview

```

                        FRONTEND ARCHITECTURE                        

                                                                     
     
                        UI LAYER
                     (apps/VyzoriX_web/src/ui/)
                                                                  
      Pages, Components, Shared UI                               
      ONLY renders UI. Uses hooks for everything.                 
      NEVER imports from Data or Domain.                          
     
                                                                     
                               uses                                  
                                                                     
     
                     PRESENTATION LAYER
                        (apps/VyzoriX_web/src/hooks/)
                                                                  
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

### 7.2 Dependency Rule

```
UI Layer  Presentation Layer  Domain  Data Layer
(ui/)                (hooks/)             (domain/)        (vyzorServer/)
                                                                 
                                                                 
      IMPORTS                                        
          ONLY FROM                                               
          HOOKS                                                   
                                                              
                                              IMPORTS 
                                                  ONLY FROM
                                                  DOMAIN TYPES
```

---

## 8. Target File Structure

> **Reflects the implemented layout.** Domain + Data layers live in the `packages/API_Client`
> package (shared by web & mobile). Presentation (hooks) and UI live in `apps/VyzoriX_web/src/`.
> Paths marked **(DONE)** exist in the codebase today.

### 8.1 Complete Directory Tree

```
packages/API_Client/src/                       # DOMAIN + DATA LAYERS (shared package)

 domain/                                       # DOMAIN LAYER (DONE)
    _shared/                                   # (DONE)
       domain-pagination.ts                    # Pagination, RawPagination, paginationFromRaw()
       domain-errors.ts                        # DomainError, ValidationError, NotFoundError...
       domain-shared.ts                        # DeviceStatus ("online"|"offline"|"deregistered")
       domain-validation.ts                    # ValidationResult, validateIMEI()
       index.ts                                # Barrel

    registration/                              # (DONE)
       registration-entity.ts                  # InboxEntry, InboxStatus, CreateInboxRequest,
                                               #   CreateInboxResult, ConfirmDeviceResult,
                                               #   RegisteredDevice, AckResult, DeregisterResult
       registration-mappers.ts                 # inboxEntryFromRaw(), registrationDeviceFromRaw(),
                                               #   createInboxRequestToRaw(), createInboxResultFromRaw(),
                                               #   confirmDeviceResultFromRaw()
       registration-validators.ts              # validateIMEI(), validateDeviceName(),
                                               #   validateVersion(), validateRegistrationRequest(),
                                               #   validateAcknowledgeAction(), validateStatusTransition()
       index.ts                                # Barrel

    device/                                    # (DONE) — richer Device entity for detail page
       device-entity.ts                        # Device, DeviceListItem, DeviceConnection, DeviceStats
       device-mappers.ts
       device-validators.ts
       index.ts

    organization/                              # (DONE) — the organization model
       organization-entity.ts                  # Organization, OrganizationMember, OrganizationRole
       organization-mappers.ts
       organization-validators.ts
       index.ts

 vyzorServer/                                  # DATA LAYER (DONE)

    rest/                                      # REST — PRIMARY data layer
       _shared/rest-client.ts                  # axios instance, X-Organization-ID header,
                                               #   CSRF, circuit breaker, retry, idempotency
       registration/
          registration-endpoints.ts            # registration.* (DONE) — see §10.4
          index.ts

       device/
          device-endpoints.ts                  # device list/detail/deregister (REST)
          index.ts

    graphql/                                   # GraphQL — PARTIAL (secondary)
       _shared/graphql-client.ts               # Apollo client setup
       registration/                           # (DONE)
          graphql-registration-queries.ts      # GET_INBOX_ENTRIES, GET_INBOX_ENTRY (+ query fns)
          graphql-registration-mutations.ts    # ACK_INBOX, DEREGISTER_DEVICE
          graphql-registration-fragments.ts    # INBOX_ENTRY_FRAGMENT, DEVICE_FRAGMENT
          graphql-registration-types.ts        # Raw response types
          index.ts
       device/
          graphql-device-queries.ts            # GET_DEVICES, GET_DEVICE (+ query fns)
          index.ts
       fragments/                              # shared fragments
          inbox-entry.fragment.ts
          device.fragment.ts
          telemetry.fragment.ts
          index.ts
       queries.ts                              # barrel re-export
       mutations.ts                            # barrel re-export

apps/VyzoriX_web/src/

 hooks/                                        # PRESENTATION LAYER
    _shared/                                   # (DONE)
       use-current-context.ts                  # useCurrentOrganizationId(), useSelectedImei()
       use-pagination.ts
       use-debounce.ts
       use-time-range.ts
       index.ts

    registration/                              # (DONE)
       use-inbox.ts                          # useInbox
       use-inbox-entry.ts                    # useInboxEntry
       use-acknowledge.ts                    # useAcknowledgeInbox
       use-create-inbox.ts                   # useCreateInboxRequest
       use-confirm.ts                        # useConfirmDevice
       use-dismiss.ts                        # useDismissInbox
       use-resend.ts                         # useResendInboxApproval
       use-registered-devices.ts             # useRegisteredDevices
       use-registered-device.ts              # useRegisteredDevice
       use-deregister.ts                     # useDeregisterRegisteredDevice
       use-registration-status.ts           # useRegistrationStatus
       use-register-device.ts               # useRegisterDevice
       _graphql-fallback.ts                 # GraphQL fallback normalization
       index.ts

    device/                                    # (DONE)
       use-devices.ts                          # Device list/detail (rich Device entity)
       index.ts

    organization/                              # (DONE)
       use-organizations.ts                    # Active org selection (drives X-Organization-ID)
       index.ts

 ui/                                           # UI LAYER
    pages/device/                              # Device page (TODO — currently .gitkeep)
       device.tsx                              # Tab shell: Inbox | Overview | Telemetry | History | Commands
       device-inbox.tsx                        # Inbox list + approve/reject actions
       device-overview.tsx
       device-telemetry.tsx
       device-commands.tsx
       device-history.tsx

    components/
       shared/                                 # (DONE/partial)
          section/                             # section-container, section-header
          connection-status/
          device-selector/
          command-button/, command-row/, log-entry/, export-menu/, metric-card/
       ui/                                     # (DONE) shadcn/ui primitives

 routes/
    __root.tsx                                 # (EXISTING)
    index.tsx                                  # (EXISTING)
    # Device routes to be added under /device/* (see §12.3)

 stores/                                       # (DONE)
    # authStore.organizationId, deviceSelectorStore.selectedDevice
```

---

## 9. Domain Layer

> **Status:** DONE. Implemented in `packages/API_Client/src/domain/registration/`.

### 9.1 Implemented Files

| File | Exports |
|------|---------|
| `registration-entity.ts` | `InboxStatus` (`pending\|acknowledged\|approving\|approved\|rejected\|expired`), `AcknowledgeAction` (`acknowledge\|approve\|reject`), `CreateInboxRequest`, `CreateInboxResult`, `ConfirmDeviceResult`, `InboxEntry`, `RegisteredDevice`, `InboxListResult`, `RegisteredDeviceListResult`, `AckResult`, `DeregisterResult` |
| `registration-mappers.ts` | `RawInboxEntry`, `RawRegisteredDevice`, `RawCreateInboxRequest`, `RawCreateInboxResponse`, `RawConfirmDeviceResponse`, `createInboxRequestToRaw()`, `createInboxResultFromRaw()`, `confirmDeviceResultFromRaw()`, `inboxEntryFromRaw()`, `registrationDeviceFromRaw()` |
| `registration-validators.ts` | `validateDeviceName()`, `validateVersion()`, `validateRegistrationRequest()`, `validateAcknowledgeAction()`, `isValidStatusTransition()`, `validateStatusTransition()` (re-exports `validateIMEI` from `_shared`) |
| `index.ts` | Barrel (re-exports entities, mappers, `paginationFromRaw` from `_shared`) |

### 9.2 Key Type Shapes (authoritative)

```ts
interface InboxEntry {
  id: string; imei: string; deviceName: string; deviceClass: string;
  model: string; manufacturer: string; osVersion: string; appVersion: string;
  fcmToken: string; firebaseInstallId: string;
  status: InboxStatus;
  acknowledgedAt: Date | null; approvingAt: Date | null;
  approvedAt: Date | null; rejectedAt: Date | null;
  notes: string | null; operatorId: string | null; createdAt: Date;
}

interface ConfirmDeviceResult {
  deviceId: string; imei: string; confirmed: boolean; online: boolean;
  registeredAt: Date; serverTime: Date;
}
```

> **v1.x -> v2.0 deltas:** Added `deviceClass`, `firebaseInstallId` (required), `acknowledgedAt`,
> `approvingAt`, `operatorId`. Removed `firmware`, `securityPatch`, `buildId`, `receivedAt`,
> `updatedAt` (not surfaced by the registration domain). `ConfirmDeviceResult` no longer carries
> `commandSecret`/`status`; it now carries `confirmed`, `online`, `serverTime`.

---

## 10. Data Layer

> **Status:** DONE. REST is primary; GraphQL is partial.

### 10.1 REST — `registration` object (PRIMARY, fully implemented)

**File:** `vyzorServer/rest/registration/registration-endpoints.ts` → exports `registration`

| Method | Signature | HTTP |
|--------|-----------|------|
| `createInboxRequest(request, organizationId?)` | `CreateInboxRequest → CreateInboxResult` | `POST /v1/device/inbox` |
| `confirmDevice(imei, commandSecret, organizationId?)` | `→ ConfirmDeviceResult` | `POST /v1/device/confirm` |
| `getInbox({status?, page?, limit?, organizationId?})` | `→ InboxListResult` | `GET /v1/device/inbox` |
| `getInboxEntry(imei, organizationId?)` | `→ InboxEntry \| null` | `GET /v1/inbox/:imei` |
| `acknowledgeInbox(imei, action, notes?, organizationId?)` | `→ AckResult` | `POST /v1/inbox/:imei/ack` |
| `resendInboxApproval(imei, organizationId?)` | `→ {success, message}` | `POST /v1/inbox/:imei/resend` |
| `dismissInbox(imei, organizationId?)` | `→ {status}` | `DELETE /v1/inbox/:imei` |
| `getDevices({status?, page?, limit?, organizationId?})` | `→ RegisteredDeviceListResult` | `GET /v1/devices` |
| `getDevice(imei, organizationId?)` | `→ RegisteredDevice \| null` | `GET /v1/devices/:imei` |
| `deregisterDevice(imei, organizationId?)` | `→ DeregisterResult` | `DELETE /v1/devices/:imei` |

> Every method injects `organizationId` (from the explicit arg or `getOrganizationContext()`).
> CSRF token is fetched lazily via `ensureCSRFToken()` before mutations.

### 10.2 GraphQL — Queries (implemented)

**File:** `vyzorServer/graphql/registration/graphql-registration-queries.ts`

| Constant | Operation |
|----------|-----------|
| `GET_INBOX_ENTRIES` | `inbox(organizationId, status, page, limit)` |
| `GET_INBOX_ENTRY` | `inboxEntry(organizationId, imei)` |
| `queryInboxEntries()` / `queryInboxEntry()` | Apollo query wrappers |

**File:** `vyzorServer/graphql/queries/registration-queries.ts` (legacy barrel) — also exports
`GET_DEVICES`, `GET_DEVICE`, `GET_DEVICE_TELEMETRY` (+ wrappers).

### 10.3 GraphQL — Mutations (implemented)

**File:** `vyzorServer/graphql/registration/graphql-registration-mutations.ts`

| Constant | Operation |
|----------|-----------|
| `ACK_INBOX` | `ackInbox(imei, action: APPROVE\|REJECT, notes)` |
| `DEREGISTER_DEVICE` | `deregisterDevice(imei, hard)` |

### 10.4 GraphQL — Fragments & Types (implemented)

| File | Contents |
|------|----------|
| `graphql-registration-fragments.ts` | `INBOX_ENTRY_FRAGMENT`, `DEVICE_FRAGMENT`, `TELEMETRY_FRAME_FRAGMENT` |
| `graphql-registration-types.ts` | `RawInboxEntry`, `RawInboxConnection`, `RegistrationRequestInput`, raw response types |

---

## 11. Presentation Layer - Hooks

> **Status:** DONE. All 12 registration hooks are implemented, one per file, under
> `apps/VyzoriX_web/src/hooks/registration/`. Data layer is **REST-primary with GraphQL fallback**:
> each read/mutation hook calls the `registration` REST client first and, on REST rejection,
> falls back to the GraphQL query/mutation (where one exists). Every hook reads the active org via
> `useCurrentOrganizationId()` and passes it to the data layer.

### 11.1 Implemented Hooks (one file each)

| Hook | File | Backed by (primary) | GraphQL fallback |
|------|------|---------------------|------------------|
| `useInbox(params?, opts?)` | `use-inbox.ts` | `registration.getInbox` | `queryInboxEntries` |
| `useInboxEntry(imei?, opts?)` | `use-inbox-entry.ts` | `registration.getInboxEntry` | `queryInboxEntry` |
| `useAcknowledgeInbox()` | `use-acknowledge.ts` | `registration.acknowledgeInbox` | `mutateAckInbox` |
| `useCreateInboxRequest()` | `use-create-inbox.ts` | `registration.createInboxRequest` | -- |
| `useConfirmDevice()` | `use-confirm.ts` | `registration.confirmDevice` | -- |
| `useDismissInbox()` | `use-dismiss.ts` | `registration.dismissInbox` | -- |
| `useResendInboxApproval()` | `use-resend.ts` | `registration.resendInboxApproval` | -- |
| `useRegisteredDevices(params?, opts?)` | `use-registered-devices.ts` | `registration.getDevices` | none (no GQL device list) |
| `useRegisteredDevice(imei?, opts?)` | `use-registered-device.ts` | `registration.getDevice` | none (no GQL single device) |
| `useDeregisterRegisteredDevice()` | `use-deregister.ts` | `registration.deregisterDevice` | `mutateDeregisterDevice` |
| `useRegistrationStatus(imei?, opts?)` | `use-registration-status.ts` | `registration.getInboxEntry` (status projection) | `queryInboxEntry` |
| `useRegisterDevice()` | `use-register-device.ts` | `registration.createInboxRequest` | -- |

> **GraphQL fallback helper:** `apps/VyzoriX_web/src/hooks/registration/_graphql-fallback.ts`
> normalizes raw GraphQL `unknown` responses (snake_case / seconds-based timestamps) into the
> domain types (`InboxEntry`, `AckResult`, `DeregisterResult`) so the hooks return a consistent
> shape regardless of which transport served the request.
>
> **Design rules:** Query hooks are `enabled` only when `organizationId !== null` (and, for
> entry/device/status hooks, when `imei` is defined). Mutation `onSuccess` invalidates the
> `['registration', ...]` and `['devices']` query keys so the UI stays consistent. When no org is
> selected, a REST rejection is rethrown (no GraphQL fallback) -- GraphQL fallback only runs when
> an org context exists.
>
> **v1.x -> v2.0 deltas:** The v1.x spec proposed a single `use-registration.ts`. Hooks are split
> one-per-file for cohesion. The v1.x spec lacked `useCreateInboxRequest`, `useConfirmDevice`,
> `useResendInboxApproval`, `useRegistrationStatus`, and `useRegisterDevice`; all are now present.

---

## 12. UI Layer - Components

> **Status:** TODO. The device page (`ui/pages/device/`) is currently a `.gitkeep` placeholder.
> Hooks and data layer are done; the UI is the remaining work. Components below are the target
> set, organization-aware (they consume the hooks which already inject the org id).

### 12.1 Registration Components to CREATE

| File (under `ui/pages/device/`) | Component | Purpose |
|------|-----------|---------|
| `device.tsx` | DevicePage | Tab shell: Inbox · Overview · Telemetry · History · Commands |
| `device-inbox.tsx` | DeviceInbox | Inbox list + filters (status), uses `useRegistrationInbox` |
| `device-overview.tsx` | DeviceOverview | Device overview, uses `useRegisteredDevice` |
| `device-telemetry.tsx` | DeviceTelemetry | Real-time + charts (telemetry context) |
| `device-commands.tsx` | DeviceCommands | Send commands (commands context) |
| `device-history.tsx` | DeviceHistory | Historical data |

> Inbox row / actions / filters may be inlined or split into `components/registration/` sub-files
> (inbox-entry-row, inbox-filters, registration-actions) per `FRONTEND_ARCHITECTURE.md` §5.

### 12.2 Shared Components (EXISTING / partial)

Shared UI primitives live under `ui/components/ui/` (shadcn/ui) and `ui/components/shared/`
(section, connection-status, device-selector, command-button, metric-card, ...). These are reused
by the device page — no new shared components are required for registration.

### 12.3 Routes

| File | Purpose |
|------|---------|
| `routes/device.tsx` (or `routes/device/index.tsx`) | `/device` — tab shell |
| `routes/device/inbox.tsx` | `/device/inbox` — inbox view |
| `routes/device/$imei.tsx` | `/device/:imei` — overview (default) |
| `routes/device/$imei/telemetry.tsx` | `/device/:imei/telemetry` |
| `routes/device/$imei/history.tsx` | `/device/:imei/history` |
| `routes/device/$imei/commands.tsx` | `/device/:imei/commands` |

> **Note:** The web app currently uses a single `routes/index.tsx` + `__root.tsx`. The device
> routes above are the target once the device page is implemented. All routes render within the
> organization context set by `useOrganizations` / `authStore.organizationId`.

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

> **Reality check:** The v1.x summary counted 51 new files. The actual implementation
> consolidated heavily. Counts below reflect what **exists** (DONE) vs. what **remains** (TODO).

### 13.1 Actual File Count

| Category | DONE | TODO |
|----------|------|------|
| Domain Layer (`packages/API_Client/src/domain/registration/`) | 4 | 0 |
| Domain Layer (`domain/device/`, `domain/organization/`, `domain/_shared/`) | done | 0 |
| Data Layer -- REST (`vyzorServer/rest/registration/`) | 2 | 0 |
| Data Layer -- GraphQL (`vyzorServer/graphql/registration/`) | 5 | 0 |
| Presentation Layer (`hooks/registration/` 12 files + `_graphql-fallback.ts`) | 13 | 0 |
| UI Layer -- Device page (`ui/pages/device/`) | 0 | ~6 |
| Routes (`routes/device/*`) | 0 | ~6 |
| **TOTAL** | **~24 (+shared)** | **~12** |

### 13.2 DONE Files (authoritative -- API Client only)

| File | Purpose |
|------|---------|
| `packages/API_Client/src/domain/registration/registration-entity.ts` | Registration domain types |
| `packages/API_Client/src/domain/registration/registration-mappers.ts` | Raw <-> domain mappers |
| `packages/API_Client/src/domain/registration/registration-validators.ts` | Validation + state-transition rules |
| `packages/API_Client/src/domain/registration/index.ts` | Barrel |
| `packages/API_Client/src/vyzorServer/rest/registration/registration-endpoints.ts` | `registration` REST client (primary) |
| `packages/API_Client/src/vyzorServer/rest/registration/index.ts` | Barrel |
| `packages/API_Client/src/vyzorServer/graphql/registration/graphql-registration-queries.ts` | Inbox queries |
| `packages/API_Client/src/vyzorServer/graphql/registration/graphql-registration-mutations.ts` | ack/deregister mutations |
| `packages/API_Client/src/vyzorServer/graphql/registration/graphql-registration-fragments.ts` | Fragments |
| `packages/API_Client/src/vyzorServer/graphql/registration/graphql-registration-types.ts` | Raw types |
| `packages/API_Client/src/vyzorServer/graphql/registration/index.ts` | Barrel |

### 13.3 TODO Files (UI remaining)

| File | Purpose |
|------|---------|
| `apps/VyzoriX_web/src/ui/pages/device/device.tsx` | Device page tab shell |
| `apps/VyzoriX_web/src/ui/pages/device/device-inbox.tsx` | Inbox list + actions |
| `apps/VyzoriX_web/src/ui/pages/device/device-overview.tsx` | Overview tab |
| `apps/VyzoriX_web/src/ui/pages/device/device-telemetry.tsx` | Telemetry tab |
| `apps/VyzoriX_web/src/ui/pages/device/device-commands.tsx` | Commands tab |
| `apps/VyzoriX_web/src/ui/pages/device/device-history.tsx` | History tab |
| `apps/VyzoriX_web/src/routes/device/*` | Route files (6) |

---

## 14. Implementation Order

> **Status:** Phases 1-6 are DONE (server + API Client domain/data + web hooks). Phase 7 (UI) is
> the only remaining work. Server-side work is tracked in
> `SERVER_BACKEND_DEVICE_REGISTRATION_API.md`.

### Phase 1: Database & Storage -- DONE (server-side)
Create `inbox_requests` (with `organization_id`), update `devices`, add `registration_logs`.

### Phase 2: Domain & Application -- DONE (server-side)
`domain/inbox/`, `application/inbox/`, `application/device/` (see server spec §6-8).

### Phase 3: REST API Handlers -- DONE (server-side)
Inbox + device endpoints, organization-scoped via `X-Organization-ID`.

### Phase 4: GraphQL -- PARTIAL (server-side)
Inbox queries + ack/deregister mutations implemented; REST remains primary.

### Phase 5: Frontend API Client (domain + data) -- DONE
- `packages/API_Client` domain/registration, vyzorServer/rest/registration, graphql/registration

### Phase 6: Frontend Hooks -- DONE (12 files, one per hook, REST-primary + GraphQL fallback)
- `apps/VyzoriX_web/src/hooks/registration/use-*.ts` (12 hooks) + `_graphql-fallback.ts`
- Tests: `apps/VyzoriX_web/src/test/hooks/use-registration.test.ts` (23 tests)
- eslint + tsc clean

### Phase 7: Frontend UI -- TODO
1. Build `ui/pages/device/device.tsx` tab shell
2. Build `device-inbox.tsx` (uses `useRegistrationInbox`, `useAcknowledgeInbox`, `useDismissInbox`, `useResendInboxApproval`)
3. Build `device-overview.tsx` (uses `useRegisteredDevice`)
4. Build `device-telemetry.tsx` (telemetry context hooks)
5. Build `device-commands.tsx` (commands context hooks)
6. Build `device-history.tsx`
7. Add `routes/device/*` route files

### Phase 8: Integration & Polish -- TODO
1. Wire hooks to components (all already org-scoped)
2. Add loading/error/empty states
3. Add optimistic updates on approve/reject/dismiss
4. Test end-to-end flow across an organization
5. Add tests

---

## 15. Testing Strategy

### 15.1 Unit Tests
- Registration mappers (`inboxEntryFromRaw`, `confirmDeviceResultFromRaw`)
- Validators (`validateRegistrationRequest`, `validateStatusTransition`)
- State-transition matrix (6-state model)

### 15.2 Integration Tests
- Full registration flow within an organization
- Inbox state machine (pending → acknowledged → approving → approved → registered)
- REST endpoints (organization-scoped: cross-org isolation)

### 15.3 E2E Tests
- Device sends request → Operator sees in inbox (same org)
- Operator approves → Device confirms via `commandSecret`
- Operator rejects → Entry removed from active list
- Cross-org isolation: org A operator cannot see org B devices/inbox

---

## 16. Rollout Checklist

- [ ] Migration run on production DB (`inbox_requests`, `devices` org columns, `registration_logs`)
- [ ] `organizations` table + `X-Organization-ID` middleware enforced server-side
- [ ] FCM credentials configured
- [ ] REST endpoints deployed (primary)
- [ ] GraphQL schema deployed (partial: inbox + ack/deregister)
- [x] Web registration hooks finished & committed (`hooks/registration/use-*.ts`, 12 files)
- [ ] Frontend device page deployed
- [ ] APK updated with `commandSecret` confirm flow + `X-Organization-ID`
- [ ] Monitoring alerts configured
- [ ] Runbook updated

---

*Document Version: 2.0*
*Status: Aligned with Organization Model & implemented API Client*
*Source of truth: `packages/API_Client/` + `SERVER_BACKEND_DEVICE_REGISTRATION_API.md`*
