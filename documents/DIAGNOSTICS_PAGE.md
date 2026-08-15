# Diagnostics Page - Enterprise Implementation Specification

> **Version:** 2.0
> **Status:** Aligned with Organization Model
> **Created:** 2026-06-21
> **Updated:** 2026-08-15
> **Target:** Production MVP
> **Architecture:** Layered (Following `FRONTEND_ARCHITECTURE.md`)
> **Source of Truth:** API Client (`packages/API_Client/`) + Go server (`apps/api/internal/`) — implemented & authoritative

---

## Table of Contents

1. [Overview](#1-overview)
2. [Page Structure](#2-page-structure)
3. [Architecture](#3-architecture)
4. [Target File Structure](#4-target-file-structure)
5. [Tab 1: Inspector](#5-tab-1-inspector)
6. [Tab 2: Timeline](#6-tab-2-timeline)
7. [Domain Layer](#7-domain-layer)
8. [Data Layer](#8-data-layer)
9. [Presentation Layer - Hooks](#9-presentation-layer---hooks)
10. [UI Layer - Components](#10-ui-layer---components)
11. [State Management - Zustand Stores](#11-state-management---zustand-stores)
12. [File Changes Summary](#12-file-changes-summary)
13. [Implementation Order](#13-implementation-order)
14. [Testing Strategy](#14-testing-strategy)
15. [Rollout Checklist](#15-rollout-checklist)

---

---

> **Architecture & Organization Alignment Note (v2.0)**
>
> This document has been realigned to the **Organization-scoped model** and the
> **implemented API Client + Go server**. The original v1.x specs were written *before*
> the organization model was designed, so they described a single-tenant diagnostics
> surface with flat `apps/web/src/` paths and no org context. All diagnostics data is
> now **organization-scoped**:
>
> - Every device inspection and timeline event belongs to an **organization**. Devices
>   carry an `organization_id`; an operator can only inspect devices in organizations
>   they are a member of.
> - The API Client injects the active organization on every call via the
>   `organization_id` query param (REST) **and** the `X-Organization-ID` header, sourced
>   from `useCurrentOrganizationId()`. The GraphQL queries require a non-null
>   `$organizationId: ID!` variable.
> - The server enforces org isolation through a layered defense:
>   `OrganizationContext` middleware → `OrganizationMembership` middleware →
>   `VerifyDeviceOwnership` DOA check (service) → `FindByIMEI` + `dev.OrganizationID != orgID`
>   guard (service). The `device_events` / `telemetry` tables themselves are filtered by
>   `device_id` only (org isolation is enforced above the SQL layer by resolving the
>   device through its org-scoped ownership first).
>
> **Layered Architecture** (unchanged from `FRONTEND_ARCHITECTURE.md`):
> - **UI Layer** (`apps/VyzoriX_web/src/ui/`) — Pure UI rendering, imports only from hooks
> - **Presentation Layer** (`apps/VyzoriX_web/src/hooks/`) — UI logic, state management
> - **Domain Layer** (`packages/API_Client/src/domain/`) — Types, mappers, validation (NO external imports)
> - **Data Layer** (`packages/API_Client/src/vyzorServer/`) — API clients (REST primary / GraphQL), imports only domain types
>
> **Dependency Rule:** UI → Hooks → Domain → Data (flow inward only)
>
> **Data layer policy:** REST is the **primary** data layer for diagnostics (fully
> implemented in `vyzorServer/rest/diagnostics/diagnostics-endpoints.ts`). GraphQL
> diagnostics coverage is complete (inspection + timeline queries, org-scoped) and is
> used as the **fallback** when REST rejects. REST is the source of truth for the API
> contract (paths, params, response shapes).
>
> **Org context contract (hooks):** Every diagnostics hook MUST:
> 1. Resolve the active org via `useCurrentOrganizationId()` (from `hooks/_shared/use-current-context`).
> 2. Be **disabled (return idle) when `organizationId === null`** — no request leaves the
>    browser without an org context. Use `enabled: organizationId !== null && imei !== undefined`.
> 3. Forward `organizationId ?? undefined` to the REST data-layer call (the REST methods
>    default to `getOrganizationContext()` when omitted, but the hook passes it explicitly).
> 4. Include `organizationId` in every **TanStack Query key** so org switches never serve
>    another org's cached inspection/timeline.

---

## 1. Overview

### 1.1 Purpose

The Diagnostics page provides operators with deep visibility into a single device:
1. **Inspector** — Current State: a full snapshot of what the server knows about the device right now (identity, software, registration, connection, telemetry stats).
2. **Timeline** — Audit Trail: a chronological, cursor-paginated event stream for the device (telemetry, command, connection, error events).

### 1.2 Design Principles

- **No fake data** — every field is backed by a real server source (devices table, ws hub, telemetry table, device_events table).
- **No filler content** — every data point is actionable.
- **Org-scoped by default** — an operator never sees devices outside their active organization.
- **Two focused views** — Inspector for now, Timeline for history.
- **Premium aesthetic** — clean, dense, command-center feel.

### 1.3 Relation to Other Pages

| Page | Responsibility |
|------|----------------|
| **Dashboard** | Overview + Metrics + Commands + Logs + Alerts (org-scoped tabs) |
| **Device** | Inbox + Overview + Telemetry + Commands + History (tabs) |
| **Diagnostics** | Inspector + Timeline (tabs) — deep dive on one device |
| **Updates** | Version status + Changelog + Push updates |
| **Settings** | Configuration |

### 1.4 Server Sources of Truth (implemented)

| Inspector section | Server source |
|-------------------|---------------|
| Identity | `devices` row (id/imei, device_name, model, manufacturer) |
| Software | `devices` row (os_version, app_version, security_patch, build_id) |
| Registration | derived from `device.Lifecycle` + `registered_at` + `fcm_token*` + `command_secret_hash` |
| Connection | `ws.Hub.GetConnectionInfo(deviceID)` + `devices.last_seen` |
| Telemetry stats | `DiagnosticsRepository.GetTelemetryStats(deviceID)` (telemetry table, today) |
| Timeline events | `device_events` table (cursor-paginated, desc by timestamp+id) |

---

## 2. Page Structure

### 2.1 Layout

```
DIAGNOSTICS                                              [Refresh]
[Inspector]  [Timeline]
TAB CONTENT
```

### 2.2 Navigation

- Two tabs: **Inspector** (default) and **Timeline**.
- Refresh button in header (refetches the active tab's query).
- Tab state persists in URL (`?tab=inspector|timeline`).
- The target device IMEI is sourced from `useDeviceSelectorStore` (or the route param `:imei`).

---

## 3. Architecture

### 3.1 Layered Architecture Overview

```
UI LAYER  (apps/VyzoriX_web/src/ui/)
   Pages, Components, Shared UI
   ONLY renders UI. Uses hooks for everything.
   NEVER imports from Data or Domain.
        uses
PRESENTATION LAYER  (apps/VyzoriX_web/src/hooks/)
   UI Logic, State Management (TanStack Query + Zustand)
   Imports from Domain and Data layers.
        uses
DOMAIN LAYER  (packages/API_Client/src/domain/)
   Types, Mappers, Validation (Pure TypeScript)
   NO external imports (no React, no API, no i18n)
        uses
DATA LAYER  (packages/API_Client/src/vyzorServer/)
   REST Endpoints (primary) + GraphQL Queries (fallback)
   Imports Domain types only.
```

### 3.2 Dependency Rule

- UI Layer can ONLY import from Presentation Layer (hooks).
- Presentation Layer can import from Domain Layer and Data Layer.
- Domain Layer can NOT import from any other layer.
- Data Layer can import from Domain Layer only.

### 3.3 Org-Context Flow

```
useAuthStore.organizationId
   -> useCurrentOrganizationId()  (hooks/_shared/use-current-context)
        -> passed into every diagnostics hook
             -> queryKey includes organizationId  (TanStack cache isolation)
             -> queryFn calls diagnostics.inspectDevice(imei, organizationId)
                  -> REST: params.organization_id  +  header X-Organization-ID
                  -> GraphQL fallback: $organizationId: ID! (non-null)
                       -> server OrganizationContext middleware -> Membership -> DOA -> service org guard
```

---

## 4. Target File Structure

> All paths are org-model-aligned (`apps/VyzoriX_web/src/`, `packages/API_Client/src/`).
> The `domain/`, `vyzorServer/rest/diagnostics/`, and `vyzorServer/graphql/diagnostics/`
> directories already exist and are implemented. Hooks exist but need org-key + GraphQL
> fallback hardening. UI + stores are TODO.

```
packages/API_Client/src/
  domain/diagnostics/                         # DOMAIN LAYER (IMPLEMENTED)
    diagnostics-entity.ts                     # DeviceInspection, TimelineEvent, TimelineEventType, ...
    diagnostics-mappers.ts                    # *FromRaw + graphql*FromRaw mappers + Raw* types
    index.ts                                  # barrel

  vyzorServer/
    rest/diagnostics/                         # DATA LAYER — REST (IMPLEMENTED, primary)
      diagnostics-endpoints.ts                # diagnostics.inspectDevice / diagnostics.getTimeline
      index.ts
    graphql/diagnostics/                      # DATA LAYER — GraphQL (IMPLEMENTED, fallback)
      graphql-diagnostics-queries.ts          # GET_DEVICE_INSPECTION, GET_DEVICE_TIMELINE + query wrappers
      graphql-diagnostics-fragments.ts        # DEVICE_INSPECTION_FRAGMENT, TIMELINE_EVENT_FRAGMENT
      graphql-diagnostics-types.ts            # Raw* GraphQL response types
      index.ts

apps/VyzoriX_web/src/
  hooks/
    _shared/use-current-context.ts            # useCurrentOrganizationId() / useRequiredOrganizationId() (IMPLEMENTED)
    diagnostics/                              # PRESENTATION LAYER
      use-diagnostics.ts                      # useDeviceInspection, useDeviceTimeline (EXIST — harden)
      use-timeline-filter.ts                  # NEW — local filter state (type, range, auto-scroll)
      use-diagnostic-stream.ts                # NEW — realtime WS subscription for live connection + events
      _graphql-fallback.ts                    # NEW — GraphQL fallback wrappers (inspection + timeline)
      index.ts

  stores/                                     # STATE MANAGEMENT (TODO — see section 11)
    diagnostics-store.ts                      # NEW — inspector snapshot cache + refresh cadence
    timeline-stream-store.ts                  # NEW — ring-buffered live event stream
    index.ts                                  # MODIFIED — add new exports

  lib/
    query-keys.ts                             # inspection/timeline factories (EXIST — add organizationId)

  ui/
    shared/                                   # Shared UI atoms
      section.tsx, section-header.tsx, empty-state.tsx,
      loading-skeleton.tsx, refresh-button.tsx, tab-nav.tsx, pagination.tsx, index.ts
    diagnostics/                              # UI LAYER (TODO)
      diagnostics-page.tsx                    # Page wrapper with tabs
      diagnostics-inspector.tsx               # Inspector tab content
      diagnostics-timeline.tsx                # Timeline tab content
      inspector-section.tsx                   # Collapsible section
      inspector-field.tsx                     # Key-value display
      inspector-identity.tsx, inspector-software.tsx,
      inspector-registration.tsx, inspector-connection.tsx, inspector-telemetry.tsx
      timeline-event.tsx, timeline-filters.tsx, timeline-controls.tsx
      index.ts

  routes/
    diagnostics-page.tsx                      # MODIFIED — layout with tabs
    diagnostics.inspector.tsx                 # NEW — Inspector tab route
    diagnostics.timeline.tsx                  # NEW — Timeline tab route
```

---

## 5. Tab 1: Inspector

### 5.1 Purpose

A single-shot snapshot of the device's current state. Backed by
`GET /v1/device/:imei/inspect` (REST) with GraphQL `deviceInspection` fallback.

### 5.2 Data Sections

The inspector renders five collapsible sections, mapping 1:1 to the
`DeviceInspection` domain type:

| Section | `DeviceInspection` field | Key fields |
|---------|--------------------------|------------|
| **Identity** | `identity: IdentityInfo` | imei, deviceName, model, manufacturer |
| **Software** | `software: SoftwareInfo` | osVersion, appVersion, securityPatch, buildId |
| **Registration** | `registration: RegistrationInfo` | status, registeredAt, fcmTokenValid, fcmTokenRefreshedAt, commandSecretSet |
| **Connection** | `connection: ConnectionInfo` | webSocketStatus, connectedAt, fcmStatus, lastSeen, clientIp, protocol |
| **Telemetry** | `telemetry: TelemetryInfo` | lastTimestamp, framesToday, avgLatencyMs, totalBytesToday, sessionsToday |

### 5.3 Field Provenance (server to TS)

| Section | Field | TS type | Server source | Notes |
|---------|-------|---------|---------------|-------|
| identity | imei | string | `devices.id` | device IMEI is the primary key |
| identity | deviceName | string? | `devices.device_name` | |
| identity | model | string? | `devices.model` | |
| identity | manufacturer | string? | `devices.manufacturer` | |
| software | osVersion | string? | `devices.os_version` | |
| software | appVersion | string? | `devices.app_version` | |
| software | securityPatch | string? | `devices.security_patch` | |
| software | buildId | string? | `devices.build_id` | currently `""` server-side (TODO) |
| registration | status | DeviceStatus | derived: `deregistered`/`offline`/`registered`/`connected` | from `device.Lifecycle` + `last_seen` + `cfg.OfflineThresholdMinutes` |
| registration | registeredAt | Date? | `devices.registered_at` (int64 ms) | |
| registration | fcmTokenValid | boolean | derived: `FCMToken != "" && IsFCMTokenValid()` | 30-day validity window |
| registration | fcmTokenRefreshedAt | Date? | `devices.fcm_token_refreshed_at` | |
| registration | commandSecretSet | boolean | derived: `CommandSecretHash != ""` | |
| connection | webSocketStatus | "connected"/"disconnected" | `ws.Hub.GetConnectionInfo` + `last_seen` | |
| connection | connectedAt | Date? | `ws.Hub` connection time | nil when offline |
| connection | fcmStatus | "valid"/"invalid"/"not_set" | derived from `FCMToken` + `cfg.FCMTokenExpiryDays` | |
| connection | lastSeen | Date? | `devices.last_seen` (int64 ms) | |
| connection | clientIp | string? | `ws.Hub` `Conn.RemoteAddr` | IP w/o port |
| connection | protocol | string? | hardcoded `"WSS"` | |
| telemetry | lastTimestamp | Date? | latest telemetry frame | |
| telemetry | framesToday | number | `COUNT(*)` telemetry since start-of-day | |
| telemetry | avgLatencyMs | number? | `ws.Hub.GetAverageLatency(deviceID)` | only when WS connected |
| telemetry | totalBytesToday | number? | `SUM(length(frame_data))` telemetry today | |
| telemetry | sessionsToday | number | distinct hour-buckets of telemetry today | |

### 5.4 Caching

- The server caches `GetDeviceInspectionHTTP` per `imei:orgID` for
  `cfg.InspectionCacheTTLSeconds` (default 10s), in-memory, cleaned every 30s.
- The client uses a TanStack Query with a short `staleTime` (e.g. 10s) to align with the
  server cache; the Refresh button issues `refetch()` bypassing stale.

### 5.5 Realtime (optional, live inspector)

- A `useDiagnosticStream` hook may subscribe to the WebSocket to update
  `connection.webSocketStatus`, `connectedAt`, `clientIp`, and `telemetry.avgLatencyMs`
  in realtime (patching the cached inspection via `diagnostics-store`).
- This is additive — the base inspector works purely from the REST snapshot.

---

## 6. Tab 2: Timeline

### 6.1 Purpose

A chronological, cursor-paginated audit trail of device events. Backed by
`GET /v1/device/:imei/timeline` (REST) with GraphQL `deviceTimeline` fallback.

### 6.2 Event Types and Categories

`TimelineEventType` (12 variants) maps to 4 frontend `TimelineEventCategory` filters:

| Category | Event types |
|----------|-------------|
| `telemetry` | `TELEMETRY` |
| `command` | `COMMAND_SENT`, `COMMAND_ACK`, `COMMAND_FAILED` |
| `connection` | `CONNECTION_OPEN`, `CONNECTION_LOST`, `FCM_FALLBACK`, `RECONNECTED`, `REGISTERED`, `DEREGISTERED` |
| `error` | `ERROR`, `THRESHOLD_BREACH` |

> **Category mapping (aligned):** The TS `getEventCategory()` maps `REGISTERED` and
> `DEREGISTERED` to `"connection"` to match the Go `EventCategory` map. Note: the server
> `mapEventTypeCategory` filter does not include `REGISTERED`/`DEREGISTERED` in any
> category filter, so selecting a category that does not include them will hide them; the
> "all" filter (empty eventType) fetches everything.

### 6.3 Query Parameters (REST)

| Param (REST snake_case) | GraphQL var | Type | Default | Description |
|-------------------------|-------------|------|---------|-------------|
| `event_type` | `eventType: TimelineEventType` | string | all | category (`telemetry`/`command`/`connection`/`error`) **or** exact event type |
| `start_time` | `startTime: Int` | int64 ms | now-24h | start timestamp |
| `end_time` | `endTime: Int` | int64 ms | now | end timestamp |
| `limit` | `limit: Int = 50` | int | 50 | max results (server clamps to 1-200, default 50) |
| `cursor` | `cursor: String` | string | null | base64-JSON `{t, i}` pagination cursor |
| `organization_id` | `$organizationId: ID!` | string | required | org context (REST query param + header; GraphQL non-null var) |

### 6.4 Response Shape

REST returns `TimelineResponse`:
```json
{
  "events": [{ "id": "...", "type": "TELEMETRY", "timestamp": 1718900567000, "data": {} }],
  "pagination": { "limit": 50, "hasMore": true, "nextCursor": "eyJ0Ijoi..." }
}
```
GraphQL returns `TimelineConnection` (flat):
```json
{ "events": [...], "hasMore": true, "nextCursor": "..." }
```

> **Pagination flattening (aligned):** The REST mapper `timelineResultFromRaw` reads
> `hasMore`/`nextCursor` from the nested `pagination` object and falls back to the
> top-level fields, so it handles both the REST (`pagination`-nested) and GraphQL (flat)
> response shapes correctly.

### 6.5 Cursor Encoding

Opaque base64-JSON `{ "t": <RFC3339Nano>, "i": <event-id> }`. The server encodes at both
the service layer (`encodeCursor`) and the storage repo (`encodeCursor`); clients must
treat it as opaque and pass it back unchanged.

### 6.6 Realtime (live timeline)

- `useDiagnosticStream` may append incoming WS events into `timeline-stream-store` so new
  events appear without a full refetch.
- The paginated REST/GraphQL query remains the source of truth for historical events;
  the live stream only prepends *new* events while the tab is open.

---

## 7. Domain Layer

> **Status: IMPLEMENTED.** Source of truth: `packages/API_Client/src/domain/diagnostics/`.

### 7.1 Entity Types (`diagnostics-entity.ts`)

```typescript
import type { DeviceStatus } from "../_shared";
export type { DeviceStatus } from "../_shared";

export type FCMStatus = "valid" | "invalid" | "not_set";
export type WebSocketStatus = "connected" | "disconnected";

export type TimelineEventType =
  | "TELEMETRY" | "COMMAND_SENT" | "COMMAND_ACK" | "COMMAND_FAILED"
  | "CONNECTION_OPEN" | "CONNECTION_LOST" | "FCM_FALLBACK" | "RECONNECTED"
  | "THRESHOLD_BREACH" | "REGISTERED" | "DEREGISTERED" | "ERROR";

export type TimelineEventCategory = "telemetry" | "command" | "connection" | "error";

export interface IdentityInfo { imei: string; deviceName?: string; model?: string; manufacturer?: string; }
export interface SoftwareInfo { osVersion?: string; appVersion?: string; securityPatch?: string; buildId?: string; }
export interface RegistrationInfo {
  status: DeviceStatus; registeredAt?: Date; fcmTokenValid: boolean;
  fcmTokenRefreshedAt?: Date; commandSecretSet: boolean;
}
export interface ConnectionInfo {
  webSocketStatus: WebSocketStatus; connectedAt?: Date; fcmStatus: FCMStatus;
  lastSeen?: Date; clientIp?: string; protocol?: string;
}
export interface TelemetryInfo {
  lastTimestamp?: Date; framesToday: number; avgLatencyMs?: number;
  totalBytesToday?: number; sessionsToday: number;
}
export interface DeviceInspection {
  identity: IdentityInfo; software: SoftwareInfo; registration: RegistrationInfo;
  connection: ConnectionInfo; telemetry: TelemetryInfo;
}
export interface TimelineEvent {
  id: string; deviceId: string; type: TimelineEventType; timestamp: Date;
  data: Record<string, unknown>;
}
export interface TimelineResult { events: TimelineEvent[]; hasMore: boolean; nextCursor?: string; }
```

Also exports `getEventCategory(type)` and `timelineEventTypeLabel(type)` helpers.

### 7.2 Mappers (`diagnostics-mappers.ts`)

Two parallel mapper sets — one for REST (int64-ms timestamps) and one for GraphQL (RFC3339 strings):

| REST mapper | GraphQL mapper | Output |
|-------------|----------------|--------|
| `identityFromRaw` | (reuses REST) | `IdentityInfo` |
| `softwareFromRaw` | (reuses REST) | `SoftwareInfo` |
| `registrationFromRaw` | `graphqlRegistrationFromRaw` | `RegistrationInfo` |
| `connectionFromRaw` | `graphqlConnectionFromRaw` | `ConnectionInfo` |
| `diagnosticTelemetryFromRaw` | `graphqlTelemetryFromRaw` | `TelemetryInfo` |
| `deviceInspectionFromRaw` | `graphqlDeviceInspectionFromRaw` | `DeviceInspection` |
| `timelineEventFromRaw` | `graphqlTimelineEventFromRaw` | `TimelineEvent` |
| `timelineResultFromRaw` | `graphqlTimelineResultFromRaw` | `TimelineResult` |

Raw types: `Raw*` (REST, int64 ms) and `RawGraphQL*` (GraphQL, string timestamps).

> **GraphQL `TimelineEvent.deviceId` gap:** The GraphQL schema's `TimelineEvent` does not
> expose `deviceId`; `graphqlTimelineEventFromRaw` hardcodes `deviceId: ""`. The REST
> path populates it from the path param. Hooks should inject the known IMEI into events
> when only the GraphQL fallback was used.

### 7.3 Barrel (`index.ts`)

```typescript
export * from "./diagnostics-entity";
export * from "./diagnostics-mappers";
```

---

## 8. Data Layer

> **Status: IMPLEMENTED.** Source of truth: `packages/API_Client/src/vyzorServer/`.

### 8.1 REST Endpoints (primary) — `rest/diagnostics/diagnostics-endpoints.ts`

```typescript
import { restClient, getOrganizationContext } from "../_shared/rest-client";
import { deviceInspectionFromRaw, timelineResultFromRaw, type RawDeviceInspection, type RawTimelineResult } from "../../../domain/diagnostics";
import type { DeviceInspection, TimelineResult, TimelineEventType } from "../../../domain/diagnostics";

const PATHS = {
  inspect:  (imei: string) => `/v1/device/${imei}/inspect`,
  timeline: (imei: string) => `/v1/device/${imei}/timeline`,
} as const;

export const diagnostics = {
  async inspectDevice(imei: string, organizationId?: string): Promise<DeviceInspection> {
    const response = await restClient.get<RawDeviceInspection>(PATHS.inspect(imei), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return deviceInspectionFromRaw(response);
  },
  async getTimeline(imei: string, params?: {
    eventType?: TimelineEventType; startTime?: number; endTime?: number;
    cursor?: string; limit?: number; organizationId?: string;
  }): Promise<TimelineResult> {
    const response = await restClient.get<RawTimelineResult>(PATHS.timeline(imei), {
      params: {
        event_type: params?.eventType, start_time: params?.startTime,
        end_time: params?.endTime, cursor: params?.cursor, limit: params?.limit,
        organization_id: params?.organizationId || getOrganizationContext(),
      },
    });
    return timelineResultFromRaw(response);
  },
};
```

**Org threading:** Each method takes `organizationId?` and sends it as the
`organization_id` query param, falling back to `getOrganizationContext()` (the global
client singleton). The `restClient` request interceptor *also* injects the
`X-Organization-ID` header from `clientState.organizationId`. Both reach the server's
`OrganizationContext` middleware (query param wins, then header, then session).

### 8.2 GraphQL Queries (fallback) — `graphql/diagnostics/`

```graphql
query GetDeviceInspection($imei: String!, $organizationId: ID!) {
  deviceInspection(imei: $imei, organizationId: $organizationId) { ...DeviceInspection }
}

query GetDeviceTimeline($imei: String!, $organizationId: ID!, $eventType: TimelineEventType, $startTime: Int, $endTime: Int, $limit: Int, $cursor: String) {
  deviceTimeline(imei: $imei, organizationId: $organizationId, eventType: $eventType, startTime: $startTime, endTime: $endTime, limit: $limit, cursor: $cursor) {
    events { ...TimelineEvent }
    hasMore
    nextCursor
  }
}
```

- **Both queries require `$organizationId: ID!`** (non-null) — there is no global
  fallback in the GraphQL wrappers (unlike REST). The hook MUST pass a real org id.
- The `DEVICE_INSPECTION_FRAGMENT` and `TIMELINE_EVENT_FRAGMENT` are inlined into the
  `gql`-wrapped `GET_DEVICE_INSPECTION` / `GET_DEVICE_TIMELINE` so Apollo resolves the
  `...` spreads at runtime.
- Wrapper functions `queryDeviceInspection({ imei, organizationId })` and
  `queryDeviceTimeline({ imei, organizationId, ... })` return `Promise<unknown>`; the
  caller extracts `.data.deviceInspection` / `.data.deviceTimeline` and applies the
  `graphql*FromRaw` mappers. They are re-exported from `vyzorServer/graphql/index.ts`
  (and thus the package `.` entry) alongside the `GET_DEVICE_*` query constants.

### 8.3 Package Exports

`packages/API_Client/package.json` exports: `.`, `./node`, `./domain`, `./vyzorServer`.
The web app imports `diagnostics` + types from `@vyzorix/api-client` (the `.` entry).

---

## 9. Presentation Layer - Hooks

> **Status: IMPLEMENTED.** `use-diagnostics.ts` is org-aware with REST-primary + GraphQL
> fallback (inspection + timeline) and a 10s `staleTime` aligned with the server cache.
> `use-timeline-filter.ts` derives REST/GraphQL query params from a category + time-range
> filter (backed by `timeline-stream-store`). `use-diagnostic-stream.ts` subscribes to
> device-updated + telemetry-received WS events and patches `diagnostics-store` +
> `timeline-stream-store`. Org-gated throughout; clears on org switch.

### 9.1 Org-Context Contract (applies to all diagnostics hooks)

1. Resolve org via `useCurrentOrganizationId()`.
2. `enabled: organizationId !== null && imei !== undefined && imei !== ''`.
3. Pass `organizationId ?? undefined` into the data-layer call.
4. Include `organizationId` in the query key (first-class segment, not buried in a blob).

### 9.2 `useDeviceInspection`

```typescript
import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import { diagnostics, type DeviceInspection } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { fetchInspectionViaGraphQL } from './_graphql-fallback';

export function useDeviceInspection(
  imei: string | undefined,
  options?: Omit<UseQueryOptions<DeviceInspection>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.inspection(organizationId ?? '', imei ?? ''),
    queryFn: async () => {
      try {
        return await diagnostics.inspectDevice(imei!, organizationId ?? undefined);
      } catch (restErr) {
        if (organizationId) {
          return fetchInspectionViaGraphQL(imei!, organizationId);   // GraphQL fallback
        }
        throw restErr;
      }
    },
    enabled: organizationId !== null && imei !== undefined && imei !== '',
    staleTime: 10_000,  // align with server 10s cache
    ...options,
  });
}
```

### 9.3 `useDeviceTimeline`

```typescript
export function useDeviceTimeline(
  imei: string | undefined,
  params?: TimelineParams,
  options?: Omit<UseQueryOptions<TimelineResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.timeline(organizationId ?? '', imei ?? '', params ?? {}),
    queryFn: async () => {
      try {
        return await diagnostics.getTimeline(imei!, { ...params, organizationId: organizationId ?? undefined });
      } catch (restErr) {
        if (organizationId) {
          return fetchTimelineViaGraphQL(imei!, organizationId, params);
        }
        throw restErr;
      }
    },
    enabled: organizationId !== null && imei !== undefined && imei !== '',
    ...options,
  });
}
```

### 9.4 `useTimelineFilter` (NEW — local UI state)

A small hook backed by `use-timeline-filter.ts` (or a `timeline-stream-store` slice) that
holds the selected `TimelineEventCategory`, time range, and auto-scroll flag. Pure local
state; no server call. Drives `useDeviceTimeline`'s `params`.

### 9.5 `useDiagnosticStream` (NEW — realtime)

Subscribes to the WebSocket (via `websocket-store`) for the selected device's events and
telemetry, patching the `diagnostics-store` (live connection status) and appending to the
`timeline-stream-store` (live events). Org-gated: only subscribes when
`organizationId !== null`. See section 11.

### 9.6 GraphQL Fallback (`_graphql-fallback.ts`, NEW)

Mirrors the `hooks/commands/_graphql-fallback.ts` / `hooks/logs/_graphql-fallback.ts`
pattern:

```typescript
import { queryDeviceInspection, queryDeviceTimeline } from '@vyzorix/api-client';
import {
  graphqlDeviceInspectionFromRaw, graphqlTimelineResultFromRaw,
  type RawGraphQLDeviceInspection, type RawGraphQLTimelineConnection,
  type DeviceInspection, type TimelineResult,
} from '@vyzorix/api-client';

export async function fetchInspectionViaGraphQL(imei: string, organizationId: string): Promise<DeviceInspection> {
  const res = await queryDeviceInspection({ imei, organizationId });
  const raw = (res as { data?: { deviceInspection?: RawGraphQLDeviceInspection } })?.data?.deviceInspection;
  if (!raw) throw new Error('GraphQL deviceInspection returned no data');
  return graphqlDeviceInspectionFromRaw(raw);
}

export async function fetchTimelineViaGraphQL(imei: string, organizationId: string, params?: TimelineParams): Promise<TimelineResult> {
  const res = await queryDeviceTimeline({
    imei, organizationId, eventType: params?.eventType,
    startTime: params?.startTime, endTime: params?.endTime,
    limit: params?.limit, cursor: params?.cursor,
  });
  const raw = (res as { data?: { deviceTimeline?: RawGraphQLTimelineConnection } })?.data?.deviceTimeline;
  if (!raw) throw new Error('GraphQL deviceTimeline returned no data');
  const result = graphqlTimelineResultFromRaw(raw);
  // GraphQL TimelineEvent lacks deviceId; inject the known imei.
  result.events = result.events.map((e) => ({ ...e, deviceId: imei }));
  return result;
}
```

### 9.7 Query Keys (`lib/query-keys.ts`)

**Org-isolated (implemented):**
```typescript
inspection: (organizationId: string, imei: string) => ['diagnostics', organizationId, imei, 'inspection'] as const,
timeline: (organizationId: string, imei: string, params?: Record<string, unknown>) => ['diagnostics', organizationId, imei, 'timeline', params ?? {}] as const,
```

Both factories take `organizationId` as an explicit first argument, so an org switch
produces a distinct cache entry and TanStack refetches — no cross-org cache collision.

---

## 10. UI Layer - Components

> **Status: TODO.** Senior-level component breakdown.

### 10.1 Shared Atoms (`ui/shared/`)

| Component | Purpose |
|-----------|---------|
| `section.tsx` | Bordered, collapsible section container |
| `section-header.tsx` | Section header with collapse toggle + optional badge |
| `empty-state.tsx` | Empty-state placeholder (icon + message) |
| `loading-skeleton.tsx` | Skeleton loaders matching section layout |
| `refresh-button.tsx` | Refresh button with loading spinner |
| `tab-nav.tsx` | Tab navigation (Inspector / Timeline) |
| `pagination.tsx` | Cursor "Load more" control (not page numbers) |

### 10.2 Inspector Components (`ui/diagnostics/`)

| Component | Props | Backing |
|-----------|-------|---------|
| `diagnostics-page.tsx` | `imei` | reads selected device + tab; renders tab nav |
| `diagnostics-inspector.tsx` | — | `useDeviceInspection`; renders 5 sections |
| `inspector-section.tsx` | `title, children, defaultOpen` | collapsible |
| `inspector-field.tsx` | `label, value, mono?` | key-value |
| `inspector-identity.tsx` | `identity: IdentityInfo` | |
| `inspector-software.tsx` | `software: SoftwareInfo` | |
| `inspector-registration.tsx` | `registration: RegistrationInfo` | status badge |
| `inspector-connection.tsx` | `connection: ConnectionInfo` | live WS badge |
| `inspector-telemetry.tsx` | `telemetry: TelemetryInfo` | stat tiles |

### 10.3 Timeline Components (`ui/diagnostics/`)

| Component | Props | Backing |
|-----------|-------|---------|
| `diagnostics-timeline.tsx` | — | `useDeviceTimeline` + `timeline-stream-store`; renders filters, list, controls |
| `timeline-event.tsx` | `event: TimelineEvent` | type icon + timestamp + data preview |
| `timeline-filters.tsx` | — | category filter, time range (drives `useTimelineFilter`) |
| `timeline-controls.tsx` | — | auto-scroll toggle, clear, load-more |

### 10.4 States

Every data-driven component handles: **loading** (skeleton), **error** (retry), **empty**
(empty-state), **success** (data). Org-missing state (`organizationId === null`) renders
an "Select an organization" empty-state.

---

## 11. State Management - Zustand Stores

> **Status: IMPLEMENTED.** Two new Zustand stores (`diagnostics-store.ts` +
> `timeline-stream-store.ts`) are created and exported from `stores/index.ts`. Both are
> org-isolated (keyed by `orgId:imei` / `orgId` and clear on org switch). Existing stores
> (`auth-store`, `device-selector-store`, `websocket-store`, `connectivity-store`) are
> reused.

### 11.1 Store Inventory

| Store | Status | Role for Diagnostics |
|-------|--------|----------------------|
| `auth-store` | EXISTS | source of `organizationId` |
| `device-selector-store` | EXISTS | source of the selected device IMEI |
| `websocket-store` | EXISTS | WS connection registry; powers live inspector + timeline |
| `connectivity-store` | EXISTS | online/offline gating for the GraphQL fallback decision |
| `diagnostics-store` | **IMPLEMENTED** | inspector snapshot cache + refresh cadence + live patches |
| `timeline-stream-store` | **IMPLEMENTED** | ring-buffered live event stream + filters + auto-scroll |
| `log-stream-store` | EXISTS (sibling) | pattern reference for `timeline-stream-store` |
| `metrics-realtime-store` | EXISTS (sibling) | pattern reference for live telemetry patches |

### 11.2 `diagnostics-store.ts` (NEW)

**Purpose:** Hold the org-scoped inspector snapshot so live WS patches can update
`connection`/`telemetry` fields without a full refetch, and to coordinate refresh cadence.

```typescript
interface DiagnosticsState {
  // Org-scoped snapshot cache: keyed by `${organizationId}:${imei}`.
  snapshots: Record<string, DeviceInspection | undefined>;
  lastRefreshedAt: Record<string, number | null>;
  isRefreshing: Record<string, boolean>;

  // Refresh cadence.
  refreshIntervalMs: number;       // default 10_000 (aligns with server cache)
  isPolling: boolean;

  // Org isolation.
  activeOrganizationId: string | null;

  // Actions.
  getSnapshot: (organizationId: string, imei: string) => DeviceInspection | undefined;
  setSnapshot: (organizationId: string, imei: string, data: DeviceInspection) => void;
  patchConnection: (organizationId: string, imei: string, patch: Partial<ConnectionInfo>) => void;
  patchTelemetry: (organizationId: string, imei: string, patch: Partial<TelemetryInfo>) => void;
  setRefreshing: (organizationId: string, imei: string, v: boolean) => void;
  setRefreshInterval: (ms: number) => void;
  setActiveOrganization: (orgId: string | null) => void;  // clears snapshots on org switch
  clear: (organizationId: string, imei: string) => void;
}
```

**Design notes:**
- Keyed by `orgId:imei` (not just imei) so org switches never leak another org's snapshot.
- `patchConnection` / `patchTelemetry` are called by `useDiagnosticStream` on WS frames.
- `setActiveOrganization` clears all snapshots when the org changes (defense-in-depth).
- Not persisted (no localStorage) — diagnostics is realtime; stale persisted snapshots are misleading.

### 11.3 `timeline-stream-store.ts` (NEW)

**Purpose:** Ring-buffered, per-device live event stream with filters and auto-scroll —
the streaming companion to the paginated `useDeviceTimeline` query.

```typescript
interface TimelineStreamState {
  byDevice: Record<string, TimelineEvent[]>;   // ring buffer, capped (e.g. 500)
  filters: { category?: TimelineEventCategory; rangeMs?: number };
  autoScroll: boolean;
  activeOrganizationId: string | null;

  append: (imei: string, event: TimelineEvent) => void;
  appendBatch: (imei: string, events: TimelineEvent[]) => void;
  setFilter: (f: Partial<{ category?: TimelineEventCategory; rangeMs?: number }>) => void;
  getEvents: (imei: string) => TimelineEvent[];   // applies filters
  toggleAutoScroll: () => void;
  clear: (imei?: string) => void;
  setActiveOrganization: (orgId: string | null) => void;  // clears on org switch
}
```

**Design notes:**
- Mirrors the proven `log-stream-store` pattern (per-device ring buffer, cap 500, filter,
  auto-scroll, org-switch clear).
- `append` prepends newest-first (to match the server's desc ordering) and trims to cap.
- `getEvents` applies the category filter + time range before returning.
- The paginated REST/GraphQL query populates history; the live stream only adds *new*
  events while the tab is open. The UI merges: live-stream (head) + paginated history (tail).

### 11.4 Reused Stores

- **`websocket-store`**: `useDiagnosticStream` reads the WS client + subscribes to the
  device's channel. No new WS plumbing needed.
- **`device-selector-store`**: the page reads the selected IMEI; the inspector/timeline
  react to selection changes (TanStack query key changes -> automatic refetch).
- **`connectivity-store`**: when offline, `useDeviceInspection`/`useDeviceTimeline` skip
  the REST attempt and go straight to the GraphQL fallback (or show cached data).

### 11.5 Why Not One Big "Diagnostics Store"

Splitting `diagnostics-store` (snapshot + live patches) from `timeline-stream-store`
(ring buffer) follows the single-responsibility principle and matches the existing
`dashboard-store` / `log-stream-store` / `metrics-realtime-store` split. A merged store
would couple refresh cadence with event buffering and make org-clear logic branch across
unrelated slices.

---

## 12. File Changes Summary

### 12.1 Total File Count

| Category | New | Modified | Notes |
|----------|-----|----------|-------|
| Domain Layer | 0 | 0 | implemented; `getEventCategory` + `timelineResultFromRaw` aligned with server |
| Data Layer (REST) | 0 | 0 | implemented; pagination flattening in place |
| Data Layer (GraphQL) | 0 | 1 | fragments inlined; `queryDeviceInspection`/`queryDeviceTimeline` re-exported |
| Presentation Layer | 4 | 2 | new hooks + fallback; `use-diagnostics.ts` + `query-keys.ts` org-scoped |
| State Management | 2 | 1 | 2 new stores + `stores/index.ts` |
| UI Layer (Shared) | 8 | 0 | shared atoms |
| UI Layer (Diagnostics) | 14 | 0 | diagnostics components |
| Routes | 2 | 1 | tab routes + page layout |
| **TOTAL** | **30** | **5** | |

### 12.2 All Files Listed

#### Domain Layer (0 NEW — implemented)

| File | Status | Purpose |
|------|--------|---------|
| `domain/diagnostics/diagnostics-entity.ts` | EXISTS | types + helpers; `getEventCategory` maps REGISTERED/DEREGISTERED → connection |
| `domain/diagnostics/diagnostics-mappers.ts` | EXISTS | REST + GraphQL mappers; `timelineResultFromRaw` flattens `pagination` |
| `domain/diagnostics/index.ts` | EXISTS | barrel |

#### Data Layer — REST (0 NEW — implemented)

| File | Status | Purpose |
|------|--------|---------|
| `vyzorServer/rest/diagnostics/diagnostics-endpoints.ts` | EXISTS | `diagnostics.inspectDevice` / `diagnostics.getTimeline` |

#### Data Layer — GraphQL (0 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| `vyzorServer/graphql/diagnostics/graphql-diagnostics-queries.ts` | EXISTS | `GET_DEVICE_INSPECTION`, `GET_DEVICE_TIMELINE`, `queryDeviceInspection`, `queryDeviceTimeline`; fragments inlined |
| `vyzorServer/graphql/diagnostics/graphql-diagnostics-fragments.ts` | EXISTS | `DEVICE_INSPECTION_FRAGMENT`, `TIMELINE_EVENT_FRAGMENT` |
| `vyzorServer/graphql/diagnostics/graphql-diagnostics-types.ts` | EXISTS | `Raw*` GraphQL types |
| `vyzorServer/graphql/index.ts` | **MODIFIED** | re-exports `queryDeviceInspection` / `queryDeviceTimeline` |

#### Presentation Layer (4 NEW, 2 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| `hooks/diagnostics/use-diagnostics.ts` | **MODIFIED** | org-scoped keys + REST-primary/GraphQL fallback + 10s staleTime |
| `hooks/diagnostics/use-timeline-filter.ts` | **NEW** | category + time-range filter → REST/GraphQL params (backed by timeline-stream-store) |
| `hooks/diagnostics/use-diagnostic-stream.ts` | **NEW** | WS device-updated + telemetry → diagnostics-store + timeline-stream-store |
| `hooks/diagnostics/_graphql-fallback.ts` | **NEW** | `fetchInspectionViaGraphQL` / `fetchTimelineViaGraphQL` |
| `hooks/diagnostics/index.ts` | **MODIFIED** | export new hooks |
| `lib/query-keys.ts` | **MODIFIED** | `inspection`/`timeline` take `organizationId` |

#### State Management (2 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| `stores/diagnostics-store.ts` | **NEW** | org-scoped inspector snapshot + live connection/telemetry patches + org-clear |
| `stores/timeline-stream-store.ts` | **NEW** | ring-buffered (500) live event stream + category/range filters + auto-scroll + org-clear |
| `stores/index.ts` | **MODIFIED** | export new stores |

#### UI Layer - Shared (8 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `ui/shared/section.tsx` | NEW | bordered collapsible section |
| `ui/shared/section-header.tsx` | NEW | section header |
| `ui/shared/empty-state.tsx` | NEW | empty state |
| `ui/shared/loading-skeleton.tsx` | NEW | skeletons |
| `ui/shared/refresh-button.tsx` | NEW | refresh button |
| `ui/shared/tab-nav.tsx` | NEW | tab nav |
| `ui/shared/pagination.tsx` | NEW | load-more control |
| `ui/shared/index.ts` | NEW | barrel |

#### UI Layer - Diagnostics (14 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `ui/diagnostics/diagnostics-page.tsx` | NEW | page wrapper with tabs |
| `ui/diagnostics/diagnostics-inspector.tsx` | NEW | inspector tab |
| `ui/diagnostics/diagnostics-timeline.tsx` | NEW | timeline tab |
| `ui/diagnostics/inspector-section.tsx` | NEW | collapsible section |
| `ui/diagnostics/inspector-field.tsx` | NEW | key-value |
| `ui/diagnostics/inspector-identity.tsx` | NEW | identity |
| `ui/diagnostics/inspector-software.tsx` | NEW | software |
| `ui/diagnostics/inspector-registration.tsx` | NEW | registration |
| `ui/diagnostics/inspector-connection.tsx` | NEW | connection |
| `ui/diagnostics/inspector-telemetry.tsx` | NEW | telemetry |
| `ui/diagnostics/timeline-event.tsx` | NEW | event row |
| `ui/diagnostics/timeline-filters.tsx` | NEW | filters |
| `ui/diagnostics/timeline-controls.tsx` | NEW | controls |
| `ui/diagnostics/index.ts` | NEW | barrel |

#### Routes (2 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| `routes/diagnostics-page.tsx` | **MODIFIED** | layout with tabs |
| `routes/diagnostics.inspector.tsx` | NEW | inspector tab route |
| `routes/diagnostics.timeline.tsx` | NEW | timeline tab route |

---

## 13. Implementation Order

### Phase 1: Data-Layer Fixes (Day 1) — ✅ DONE
1. ✅ `timelineResultFromRaw` reads `hasMore`/`nextCursor` from `pagination` (REST) with top-level fallback (GraphQL).
2. ✅ `getEventCategory` maps `REGISTERED`/`DEREGISTERED` to `"connection"` (matches Go server).
3. ✅ `TIMELINE_EVENT_FRAGMENT` + `DEVICE_INSPECTION_FRAGMENT` inlined into the `gql`-wrapped queries.
4. ✅ `queryDeviceInspection`/`queryDeviceTimeline` re-exported from `graphql/index.ts`.
5. ✅ Server: `GetAverageLatency` per-device (global fallback); `isFCMTokenValid` uses `cfg.FCMTokenExpiryDays`.

### Phase 2: Presentation Layer (Day 1) — ✅ DONE
1. ✅ `query-keys.ts` `inspection`/`timeline` take `organizationId`.
2. ✅ `_graphql-fallback.ts` (inspection + timeline, deviceId injection).
3. ✅ `use-diagnostics.ts` (REST-primary + GraphQL fallback + 10s staleTime).
4. ✅ `use-timeline-filter.ts` (category + range → params).
5. ✅ `use-diagnostic-stream.ts` (WS → stores).

### Phase 3: State Management (Day 1-2) — ✅ DONE
1. ✅ `diagnostics-store.ts` (org-scoped snapshot + patches + org-clear).
2. ✅ `timeline-stream-store.ts` (ring buffer 500 + filters + auto-scroll + org-clear).
3. ✅ Exported from `stores/index.ts`.

### Phase 4: UI Layer - Shared (Day 2)
1. Create shared atoms (section, header, empty-state, skeleton, refresh, tab-nav, pagination).

### Phase 5: UI Layer - Diagnostics (Day 2-3)
1. Inspector components (section, field, 5 sub-sections).
2. Timeline components (event, filters, controls).
3. `diagnostics-inspector.tsx` + `diagnostics-timeline.tsx` + `diagnostics-page.tsx`.

### Phase 6: Realtime (Day 3)
1. Create `use-diagnostic-stream.ts` (WS subscription -> `diagnostics-store` + `timeline-stream-store`).
2. Wire live patches into inspector connection/telemetry.
3. Wire live prepend into timeline.

### Phase 7: Route Assembly (Day 3)
1. Update `routes/diagnostics-page.tsx`.
2. Add inspector + timeline tab routes.
3. Loading/error/empty/org-missing states.

### Phase 8: Polish (Day 4)
1. Auto-scroll behavior.
2. Large-timeline performance (virtualization).
3. Dark mode + mobile responsive.

---

## 14. Testing Strategy

### Unit Tests — ✅ IMPLEMENTED
- Domain mappers: `diagnostics-mappers.test.ts` (REST int64 + GraphQL string timestamps + category mapping).
- `getEventCategory` / `timelineEventTypeLabel`.
- `diagnostics-store.test.ts` (snapshot set, org isolation, patchConnection/patchTelemetry merge, org-clear, clear/clearAll) — 12 tests.
- `timeline-stream-store.test.ts` (prepend, ring buffer trim, appendBatch dedupe, category/range filters, REGISTERED→connection, org-clear) — 12 tests.

### Hook Tests — ✅ IMPLEMENTED
- `use-diagnostics.test.ts` (9 tests): org-gating (disabled when orgId null / imei undefined), REST-primary with orgId, GraphQL fallback on REST rejection (inspection + timeline, deviceId injection), query-key org isolation (org switch refetches).
- `use-timeline-filter.test.ts` (7 tests): defaults, setCategory/setRangeMs, toggleAutoScroll, toQueryParams 24h window + custom range, clear resets filters + stream.
- `use-diagnostic-stream.test.ts` (5 tests): no-subscribe when orgId/imei missing, org propagation, WS next → store patches (connection + telemetry + timeline event), unmount unsubscribes.

### Integration Tests
- REST inspect + timeline with mock server (org-scoped params).
- GraphQL fallback with mock Apollo client.
- Cursor pagination round-trip.

### E2E Tests
- Navigate to Diagnostics (org selected) -> Inspector renders 5 sections.
- Switch org -> inspector refetches (no cross-org cache leak).
- Timeline: filter by category, load-more, auto-scroll, clear.
- No org selected -> "Select an organization" empty-state.

### Visual Regression
- Inspector section layout, status badges, telemetry tiles.
- Timeline event rows, filter bar.

---

## 15. Rollout Checklist

### Pre-Launch
- [x] `timelineResultFromRaw` pagination flattening in place.
- [x] `getEventCategory` REGISTERED/DEREGISTERED → `connection`.
- [x] `TIMELINE_EVENT_FRAGMENT` + `DEVICE_INSPECTION_FRAGMENT` inlined into the queries.
- [x] `queryDeviceInspection`/`queryDeviceTimeline` re-exported from package entry.
- [x] `inspection`/`timeline` query keys include `organizationId`.
- [x] `useDeviceInspection` / `useDeviceTimeline` disabled when `organizationId === null`.
- [x] Server: `GetAverageLatency` per-device; `isFCMTokenValid` uses config.
- [x] GraphQL fallback wired for inspection + timeline.
- [x] `diagnostics-store` + `timeline-stream-store` clear on org switch.
- [x] `useDiagnosticStream` patches stores on WS device-updated + telemetry events.
- [ ] Loading/error/empty/org-missing states on every component.
- [ ] Cursor "load more" works; auto-scroll tested.
- [ ] Mobile responsive + dark mode.

### Post-Launch
- [ ] Monitor inspect query errors + cache hit rate.
- [ ] Monitor timeline query errors.
- [ ] Check for memory leaks in timeline ring buffer.
- [ ] Verify no cross-org data leakage in staging with 2 orgs + shared IMEI (should 404/403).

---

*Document Version: 2.1*
*Status: Aligned with Organization Model — data layer + hooks + stores implemented*
*Architecture: Layered (Following `FRONTEND_ARCHITECTURE.md`)*
*Source of Truth: API Client + Go server (implemented)*
