# Settings Page — Architecture & Implementation Reference

> **Version:** 2.0
> **Status:** Updated to reflect actual server & API client
> **Last Updated:** 2026-08-15
> **Architecture:** Layered (see `FRONTEND_ARCHITECTURE.md`)

---

## Table of Contents

1. [Overview](#1-overview)
2. [Three-Tier Settings Model](#2-three-tier-settings-model)
3. [Organization Model](#3-organization-model)
4. [Server Endpoints](#4-server-endpoints)
5. [API Client Endpoints](#5-api-client-endpoints)
6. [Domain Types](#6-domain-types)
7. [Frontend Current State](#7-frontend-current-state)
8. [Settings Page Design](#8-settings-page-design)
9. [Implementation Plan](#9-implementation-plan)
10. [Default Values Reference](#10-default-values-reference)
11. [Known Issues & Schema Mismatches](#11-known-issues--schema-mismatches)

---

## 1. Overview

The Settings page allows operators to configure their client preferences, alert
thresholds, and notification channels. Settings are organized in a **three-tier
hierarchy**: operator-level (per user), organization-level (per tenant), and
device-level (per device). The Settings page primarily manages operator-level
settings via the `/v1/auth/me/*` endpoints.

---

## 2. Three-Tier Settings Model

The server resolves settings using a cascading fallback: **device → organization → operator defaults**.

| Tier | Scope | Route Prefix | Handler File |
|------|-------|-------------|--------------|
| **Operator** | Per authenticated user | `/v1/auth/me/` | `apps/api/internal/api/handlers/auth/auth_settings.go` |
| **Organization** | Per tenant (org) | `/v1/organizations/:id/` | `apps/api/internal/api/handlers/organization/organization_settings_handler.go` |
| **Device** | Per device (IMEI) | `/v1/devices/:imei/` | `apps/api/internal/api/handlers/device/device_settings_handler.go` |

Each tier exposes the same sub-resources: `settings`, `thresholds`,
`notifications` (operator only), and `webhook` (operator only).

---

## 3. Organization Model

### 3.1 Organization Entity

**File:** `apps/api/internal/domain/organization/organization_entity.go`

```go
type Organization struct {
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   *time.Time
    Lifecycle   OrganizationLifecycle  // active | inactive | archived
    ID          string
    Name        string
    Description string
    CreatedBy   string
    MaxMembers  int
    MemberCount int
}
```

### 3.2 Roles & Permissions

**File:** `apps/api/internal/domain/organization/organization_entity.go`

| Role | Level | Key Permissions |
|------|-------|----------------|
| `super_admin` | 4 | Delete org, manage API keys, manage members |
| `admin` | 3 | Manage org settings, manage members, manage API keys |
| `operator` | 2 | Register/manage devices |
| `viewer` | 1 | View devices (read-only) |

### 3.3 OrganizationMember

```go
type OrganizationMember struct {
    Lifecycle      MemberLifecycle  // invited | active | suspended | removed
    ID             string
    OrganizationID string
    OperatorID     string
    Role           OrganizationRole
    InvitedBy      *string
    JoinedAt       time.Time
    RemovedAt      *time.Time
    SuspendedAt    *time.Time
    OperatorName   string  // joined
    OperatorEmail  string  // joined
}
```

### 3.4 OrganizationSettings

**File:** `apps/api/internal/domain/organization/organization_settings_entity.go`

```go
type OrganizationSettings struct {
    CreatedAt            time.Time   `json:"createdAt"`
    UpdatedAt            time.Time   `json:"updatedAt"`
    DefaultThresholds    *Thresholds `json:"defaultThresholds,omitempty"`
    ID                   string      `json:"id"`
    OrganizationID       string      `json:"organizationId"`
    Timezone             string      `json:"timezone"`
    DateFormat           string      `json:"dateFormat"`
    AlertCooldownMinutes int         `json:"alertCooldownMinutes"`
}

type Thresholds struct {
    RiskWarn    int `json:"riskWarn"`
    RiskCrit    int `json:"riskCrit"`
    ThermalWarn int `json:"thermalWarn"`
    ThermalCrit int `json:"thermalCrit"`
    BufferWarn  int `json:"bufferWarn"`
    BufferCrit  int `json:"bufferCrit"`
}
```

---

## 4. Server Endpoints

### 4.1 Operator-Level Settings (auth/me)

**Route registration:** `apps/api/internal/api/handlers/auth/auth_routes.go:152-159`
**Handler:** `apps/api/internal/api/handlers/auth/auth_settings.go`

| Method | Endpoint | Handler | Purpose |
|--------|----------|---------|---------|
| GET | `/v1/auth/me/settings` | `GetSettings` | Returns full `OperatorSettings` (client + thresholds + notifications) |
| PATCH | `/v1/auth/me/settings` | `UpdateSettings` | Update client settings or reset (`{reset: true}`) — returns full `OperatorSettings` |
| GET | `/v1/auth/me/thresholds` | `GetThresholds` | Returns `Thresholds` |
| PATCH | `/v1/auth/me/thresholds` | `UpdateThresholds` | Partial update of threshold fields — returns `Thresholds` |
| GET | `/v1/auth/me/notifications` | `GetNotifications` | Returns `NotificationSettings` |
| PATCH | `/v1/auth/me/notifications` | `UpdateNotifications` | Partial update — returns `NotificationSettings` |
| POST | `/v1/auth/me/notifications/webhook/test` | `TestWebhook` | Tests webhook delivery |
| POST | `/v1/auth/me/notifications/webhook/rotate` | `RotateWebhookSecret` | Rotates webhook signing secret |

> **Note:** `/v1/me/settings` (server_routes.go:546-551) is a duplicate alias for the same handler. The API client uses `/v1/auth/me/*`.

### 4.2 Organization-Level Settings

**Route registration:** `apps/api/internal/api/server_routes.go:507-510`
**Handler:** `apps/api/internal/api/handlers/organization/organization_settings_handler.go`

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/v1/organizations/:id/settings` | Returns `OrganizationSettings` |
| PATCH | `/v1/organizations/:id/settings` | Update timezone, dateFormat, alertCooldownMinutes |
| GET | `/v1/organizations/:id/settings/thresholds` | Returns org-level `Thresholds` |
| PATCH | `/v1/organizations/:id/settings/thresholds` | Update org-level thresholds |

### 4.3 Device-Level Settings

**Route registration:** `apps/api/internal/api/server_routes.go:376-379`
**Handler:** `apps/api/internal/api/handlers/device/device_settings_handler.go`

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/v1/devices/:imei/settings` | Returns `DeviceSettings` (customName, location, metadata, thresholds) |
| PATCH | `/v1/devices/:imei/settings` | Update device-specific settings |
| GET | `/v1/devices/:imei/settings/thresholds` | Returns device-level `Thresholds` |
| PATCH | `/v1/devices/:imei/settings/thresholds` | Update device-specific thresholds |

---

## 5. API Client Endpoints

### 5.1 Operator Settings (settings)

**File:** `packages/API_Client/src/vyzorServer/rest/settings/settings-endpoints.ts`

```typescript
const settings = {
  getSettings(): Promise<{ client: ClientSettings; thresholds: Thresholds }>;
  updateSettings(data: { client?: Partial<ClientSettings> }): Promise<{ client: ClientSettings; thresholds: Thresholds }>;
  getThresholds(): Promise<Thresholds>;
  updateThresholds(data: Partial<Thresholds>): Promise<Thresholds>;
  getNotifications(): Promise<NotificationSettings>;
  updateNotifications(data: Partial<NotificationSettings>): Promise<NotificationSettings>;
  testWebhook(url: string): Promise<{ success: boolean; statusCode?: number; responseTime?: number; error?: string }>;
  rotateWebhookSecret(): Promise<{ secret: string }>;
};
```

### 5.2 Organization Settings

**File:** `packages/API_Client/src/vyzorServer/rest/organization/settings-endpoints.ts`

```typescript
const settings = {
  get(orgId: string): Promise<OrganizationSettings>;
  update(orgId: string, request: SettingsUpdateRequest): Promise<OrganizationSettings>;
  getThresholds(orgId: string): Promise<{ thresholds: ThresholdUpdateRequest }>;
  updateThresholds(orgId: string, request: ThresholdUpdateRequest): Promise<{ thresholds: ThresholdUpdateRequest }>;
};
```

> ⚠️ **Schema mismatch:** The API client `OrganizationSettings` interface uses
> `temp_min`/`temp_max`/`battery_min`/`battery_max` field names, while the Go
> server returns `riskWarn`/`riskCrit`/`thermalWarn`/`thermalCrit`/`bufferWarn`/
> `bufferCrit`. See [§11 Known Issues](#11-known-issues--schema-mismatches).

### 5.3 Device Settings

**File:** `packages/API_Client/src/vyzorServer/rest/device/device-endpoints.ts`

```typescript
const devices = {
  getSettings(imei: string, organizationId?: string): Promise<DeviceSettings>;
  updateSettings(imei: string, settings: Partial<DeviceSettings>, organizationId?: string): Promise<DeviceSettings>;
};
```

> ⚠️ **Schema mismatch:** The API client `DeviceSettings` interface uses
> `fcmEnabled`/`fcmToken`/`tempMin`/`batteryMin` field names, while the Go
> server returns `customName`/`location`/`metadata`/`riskWarn`/`thermalWarn`/
> `bufferWarn`. See [§11 Known Issues](#11-known-issues--schema-mismatches).

---

## 6. Domain Types

**File:** `packages/API_Client/src/domain/settings/settings-entity.ts`

### 6.1 Thresholds

```typescript
interface Thresholds {
  riskWarn: number;
  riskCrit: number;
  thermalWarn: number;
  thermalCrit: number;
  bufferWarn: number;
  bufferCrit: number;
}
```

### 6.2 ClientSettings

```typescript
interface ClientSettings {
  serverUrl: string;
  deviceId: string;
  requestTimeoutMs: number;   // valid: 500–60000
  logBufferLimit: number;     // valid: 50–5000
  signalHistoryLimit: number; // valid: 30–2000
  autoReconnect: boolean;
  strictHmac: boolean;
  notificationsEnabled: boolean;
}
```

### 6.3 NotificationSettings

```typescript
type NotificationEvent =
  | "threshold_breach" | "device_offline" | "device_online"
  | "update_available" | "command_failed" | "registration_request";

interface NotificationSettings {
  enabled: boolean;
  channels: string[];              // ["email", "push", "webhook"]
  email: EmailNotifications;       // per-event boolean flags
  push: PushNotifications;         // per-event boolean flags
  webhook: WebhookNotifications;  // {enabled, url, secret, types}
}
```

### 6.4 Go Server Response Shape (OperatorSettings)

**File:** `apps/api/internal/domain/operator/operator_entity.go`

```go
type OperatorSettings struct {
    Notifications *NotificationSettings `json:"notifications"`
    Client        ClientSettings        `json:"client"`
    Thresholds    Thresholds            `json:"thresholds"`
}
```

Both `GET /v1/auth/me/settings` and `PATCH /v1/auth/me/settings` return this shape.

---

## 7. Frontend Current State

### 7.1 Routes

**Directory:** `apps/VyzoriX_web/src/routes/`

| File | Status |
|------|--------|
| `__root.tsx` | EXISTS — root layout (wired with analytics provider + consent banner) |
| `index.tsx` | EXISTS — landing page |

> **No settings routes exist yet.** The following routes need to be created:
> `/settings`, `/settings/connection`, `/settings/operator`,
> `/settings/thresholds`, `/settings/notifications`, `/settings/advanced`.

### 7.2 Hooks

**File:** `apps/VyzoriX_web/src/hooks/settings/use-settings.ts` — EXISTS

| Hook | Purpose | API Client Method |
|------|---------|-------------------|
| `useSettings` | Fetch operator settings | `settings.getSettings()` |
| `useThresholds` | Fetch thresholds | `settings.getThresholds()` |
| `useNotifications` | Fetch notifications | `settings.getNotifications()` |
| `useUpdateSettings` | Update client settings | `settings.updateSettings()` |
| `useUpdateThresholds` | Update thresholds | `settings.updateThresholds()` |
| `useUpdateNotifications` | Update notifications | `settings.updateNotifications()` |
| `useTestWebhook` | Test webhook | `settings.testWebhook()` |
| `useRotateWebhookSecret` | Rotate webhook secret | `settings.rotateWebhookSecret()` |

### 7.3 Query Keys

**File:** `apps/VyzoriX_web/src/lib/query-keys.ts`

```typescript
settings: ['settings'] as const,
thresholds: ['settings', 'thresholds'] as const,
notifications: ['settings', 'notifications'] as const,
deviceSettings: (imei: string) => ['devices', imei, 'settings'] as const,
```

### 7.4 Stores

**Directory:** `apps/VyzoriX_web/src/stores/`

No settings-specific Zustand store exists. Settings use TanStack Query for
server state (no local persistence needed). Available related stores:
`auth-store.ts`, `theme-store.ts`, `vyzor-consent-store.ts`, `vyzor-i18n-store.ts`.

### 7.5 UI Components

No settings UI components exist yet. Placeholder directories:
- `apps/VyzoriX_web/src/ui/pages/settings/` (empty, `.gitkeep`)
- `apps/VyzoriX_web/src/ui/components/shared/` (empty, `.gitkeep`)

---

## 8. Settings Page Design

### 8.1 Target Routes (TanStack Start file-based routing)

| Route | File | Purpose |
|-------|------|---------|
| `/settings` | `settings.tsx` | Layout with tab navigation |
| `/settings/connection` | `settings.connection.tsx` | Server URL, device ID, timeout |
| `/settings/operator` | `settings.operator.tsx` | Account info, display name |
| `/settings/thresholds` | `settings.thresholds.tsx` | Risk, thermal, buffer thresholds |
| `/settings/notifications` | `settings.notifications.tsx` | Email, push, webhook preferences |
| `/settings/advanced` | `settings.advanced.tsx` | Danger zone (reset settings) |

### 8.2 Navigation Structure

```
Settings
├── Connection    (/settings/connection)    — Server URL, device ID, timeout
├── Operator      (/settings/operator)      — Display name, role display
├── Thresholds    (/settings/thresholds)    — Risk, thermal, buffer sliders
├── Notifications (/settings/notifications) — Email/push/webhook toggles
└── Advanced      (/settings/advanced)      — Reset settings to defaults
```

### 8.3 Tab: Connection

Uses `useSettings()` + `useUpdateSettings()`.

| Field | Type | Validation | Default |
|-------|------|------------|---------|
| Server URL | input (URL) | Must be valid HTTP/HTTPS URL | `""` |
| Device ID | input (text) | — | `""` |
| Request Timeout | input (number) | 500–60000 ms | 8000 |
| Auto Reconnect | toggle | — | true |
| Strict HMAC | toggle | — | false |

### 8.4 Tab: Thresholds

Uses `useThresholds()` + `useUpdateThresholds()`.

| Field | Type | Description |
|-------|------|-------------|
| Risk Warning | slider (0–100) | Risk score above which to warn |
| Risk Critical | slider (0–100) | Risk score above which to alert critical |
| Thermal Warning | slider (0–200) | Temperature warn threshold (°C) |
| Thermal Critical | slider (0–200) | Temperature critical threshold (°C) |
| Buffer Warning | slider (0–100) | Buffer level below which to warn |
| Buffer Critical | slider (0–100) | Buffer level below which to alert critical |

### 8.5 Tab: Notifications

Uses `useNotifications()` + `useUpdateNotifications()` + `useTestWebhook()` + `useRotateWebhookSecret()`.

| Field | Type | Description |
|-------|------|-------------|
| Enabled | toggle | Master notification switch |
| Channels | multi-select | `["email", "push", "webhook"]` |
| Email events | 6 toggles | Per-event email notification toggles |
| Push events | 6 toggles | Per-event push notification toggles |
| Webhook URL | input (URL) | Webhook endpoint URL |
| Webhook Enabled | toggle | Enable webhook delivery |
| Webhook Types | multi-select | Event types to send to webhook |
| Test Webhook | button | Tests webhook delivery (10s timeout) |
| Rotate Secret | button | Rotates webhook signing secret |

### 8.6 Tab: Advanced (Danger Zone)

Uses `useUpdateSettings({ reset: true })`.

| Action | Type | Description |
|--------|------|-------------|
| Reset to Defaults | button (danger) | Resets all operator settings to server defaults |

---

## 9. Implementation Plan

### Phase 1: Routes & Layout
1. Create `src/routes/settings.tsx` — tab layout with navigation
2. Create 5 sub-route files (`settings.connection.tsx`, etc.)
3. Add `/settings` → redirect to `/settings/connection`

### Phase 2: UI Components
1. Create `src/ui/components/settings/connection-settings.tsx`
2. Create `src/ui/components/settings/threshold-settings.tsx`
3. Create `src/ui/components/settings/notification-settings.tsx`
4. Create `src/ui/components/settings/operator-settings.tsx`
5. Create `src/ui/components/settings/advanced-settings.tsx`
6. Create shared components: `slider-input.tsx`, `threshold-input.tsx`, `notification-row.tsx`, `danger-zone.tsx`

### Phase 3: Wire Hooks
The hooks already exist (`use-settings.ts`). Wire them to the new components:
- Connection tab → `useSettings()` + `useUpdateSettings()`
- Thresholds tab → `useThresholds()` + `useUpdateThresholds()`
- Notifications tab → `useNotifications()` + `useUpdateNotifications()` + `useTestWebhook()` + `useRotateWebhookSecret()`
- Advanced tab → `useUpdateSettings({ reset: true })`

### Phase 4: Auto-Save & UX
1. Add debounced auto-save for threshold sliders
2. Add loading skeletons and error states
3. Add toast notifications on successful saves

---

## 10. Default Values Reference

### 10.1 Threshold Defaults

| Field | Operator (Go) | Organization (Go) | API Client (TS) |
|-------|---------------|-------------------|------------------|
| riskWarn | 70 | 70 | 70 |
| riskCrit | 90 | 90 | 85 |
| thermalWarn | 75 | 75 | 45 |
| thermalCrit | 85 | 85 | 50 |
| bufferWarn | 80 | 30 | 30 |
| bufferCrit | 95 | 10 | 15 |

> ⚠️ Default values are **inconsistent** across sources. The API client
> `DEFAULT_THRESHOLDS` does not match the Go server operator defaults. See
> [§11 Known Issues](#11-known-issues--schema-mismatches).

### 10.2 ClientSettings Defaults

| Field | Go Server | API Client (TS) |
|-------|-----------|-----------------|
| requestTimeoutMs | 8000 | 8000 |
| logBufferLimit | 500 | 500 |
| signalHistoryLimit | 100 | 240 |
| autoReconnect | true | true |
| strictHmac | false | false |
| notificationsEnabled | true | true |

> ⚠️ `signalHistoryLimit` default differs: Go=100, API client=240.

### 10.3 Validation Ranges

| Field | Min | Max |
|-------|-----|-----|
| requestTimeoutMs | 500 | 60000 |
| logBufferLimit | 50 | 5000 |
| signalHistoryLimit | 30 | 2000 |
| riskWarn/riskCrit | 0 | 100 |
| thermalWarn/thermalCrit (device) | 0 | 200 |
| bufferWarn/bufferCrit | 0 | 100 |

---

## 11. API Client Alignment Status

All three tiers of the API client are now fully aligned with the Go server's
field names and default values.

### 11.1 Organization Settings (aligned)

**File:** `packages/API_Client/src/vyzorServer/rest/organization/settings-endpoints.ts`

```typescript
export interface OrganizationSettings {
  id: string;
  organizationId: string;
  timezone: string;
  dateFormat: string;
  alertCooldownMinutes: number;
  defaultThresholds?: Thresholds;
  createdAt: string;
  updatedAt: string;
}

export type ThresholdUpdateRequest = Partial<Thresholds>;

export interface SettingsUpdateRequest {
  timezone?: string;
  dateFormat?: string;
  alertCooldownMinutes?: number;
  defaultThresholds?: Thresholds;
}
```

All threshold fields use the shared `Thresholds` type (`riskWarn`/`riskCrit`/
`thermalWarn`/`thermalCrit`/`bufferWarn`/`bufferCrit`) imported from
`domain/settings`, matching the Go server's `organization.Thresholds` struct.

### 11.2 Device Settings (aligned)

**File:** `packages/API_Client/src/vyzorServer/rest/device/device-endpoints.ts`

```typescript
export interface DeviceSettings {
  id: string;
  deviceImei: string;
  customName?: string;
  location?: string;
  metadata?: Record<string, string>;
  thresholds?: Thresholds;
  createdAt: string;
  updatedAt: string;
}

export type DeviceThresholdUpdateRequest = Partial<Thresholds>;

export interface DeviceSettingsUpdateRequest {
  customName?: string;
  location?: string;
  metadata?: Record<string, string>;
  thresholds?: Thresholds;
}
```

Matches the Go server's `device.DeviceSettings` struct exactly. Threshold
endpoints (`getSettingsThresholds`/`updateSettingsThresholds`) return and accept
the shared `Thresholds` type.

### 11.3 Default Values (aligned)

All defaults in `packages/API_Client/src/domain/settings/settings-entity.ts`
now match the Go server's `operator.DefaultThresholds()` and
`operator.DefaultClientSettings()`:

```typescript
export const DEFAULT_THRESHOLDS: Thresholds = {
  riskWarn: 70, riskCrit: 90,
  thermalWarn: 75, thermalCrit: 85,
  bufferWarn: 80, bufferCrit: 95,
};

export const DEFAULT_CLIENT_SETTINGS: ClientSettings = {
  serverUrl: "", deviceId: "",
  requestTimeoutMs: 8000,
  logBufferLimit: 500,
  signalHistoryLimit: 100,
  autoReconnect: true,
  strictHmac: false,
  notificationsEnabled: true,
};
```

---

## 12. Frontend Settings Hooks

### 12.1 Operator Settings Hooks (exist)

**File:** `apps/VyzoriX_web/src/hooks/settings/use-settings.ts`

| Hook | API Method | Purpose |
|------|------------|---------|
| `useSettings` | `settings.getSettings()` | Fetch operator client + thresholds |
| `useThresholds` | `settings.getThresholds()` | Fetch operator thresholds |
| `useNotifications` | `settings.getNotifications()` | Fetch notification settings |
| `useUpdateSettings` | `settings.updateSettings()` | Update client settings |
| `useUpdateThresholds` | `settings.updateThresholds()` | Partial threshold update |
| `useUpdateNotifications` | `settings.updateNotifications()` | Partial notification update |
| `useTestWebhook` | `settings.testWebhook()` | Test webhook delivery |
| `useRotateWebhookSecret` | `settings.rotateWebhookSecret()` | Rotate webhook secret |

### 12.2 Organization Settings Hooks

**File:** `apps/VyzoriX_web/src/hooks/organization/use-org-settings.ts`

| Hook | API Method | Purpose |
|------|------------|---------|
| `useOrgSettings(orgId)` | `orgSettings.get(orgId)` | Fetch org settings |
| `useUpdateOrgSettings` | `orgSettings.update(orgId, req)` | Update org settings |
| `useOrgThresholds(orgId)` | `orgSettings.getThresholds(orgId)` | Fetch org thresholds |
| `useUpdateOrgThresholds` | `orgSettings.updateThresholds(orgId, req)` | Update org thresholds |

### 12.3 Device Settings Hooks

**File:** `apps/VyzoriX_web/src/hooks/device/use-devices.ts`

| Hook | API Method | Purpose |
|------|------------|---------|
| `useDeviceSettings(imei)` | `devices.getSettings(imei)` | Fetch device settings |
| `useUpdateDeviceSettings(imei)` | `devices.updateSettings(imei, req)` | Update device settings |
| `useDeviceThresholds(imei)` | `devices.getSettingsThresholds(imei)` | Fetch effective thresholds |
| `useUpdateDeviceThresholds(imei)` | `devices.updateSettingsThresholds(imei, req)` | Update device thresholds |

### 12.4 Query Keys

```typescript
settings: ['settings'] as const,
thresholds: ['settings', 'thresholds'] as const,
notifications: ['settings', 'notifications'] as const,
deviceSettings: (imei) => ['devices', imei, 'settings'] as const,
deviceThresholds: (imei) => ['devices', imei, 'thresholds'] as const,
orgSettings: (orgId) => ['organizations', orgId, 'settings'] as const,
orgThresholds: (orgId) => ['organizations', orgId, 'thresholds'] as const,
```

### 12.5 Stores

No Zustand store for settings. All settings state is server state managed
via TanStack Query (fetch + mutate + cache invalidation). This follows the
existing codebase pattern: hooks + query cache for server state, Zustand
stores only for realtime/client state (websocket, command queue, etc.).

---

## 13. MSW Test Infrastructure

### 13.1 Settings Handlers

**File:** `apps/VyzoriX_web/src/test/msw/vyzor-msw-handlers-settings.ts`

Covers all three tiers:
- `GET/PATCH /v1/auth/me/settings` — operator settings
- `GET/PATCH /v1/auth/me/thresholds` — operator thresholds
- `GET/PATCH /v1/auth/me/notifications` — operator notifications
- `POST /v1/auth/me/notifications/webhook/{test,rotate}` — webhook ops
- `GET/PATCH /v1/organizations/:id/settings` — org settings
- `GET/PATCH /v1/organizations/:id/settings/thresholds` — org thresholds
- `GET/PATCH /v1/devices/:imei/settings` — device settings
- `GET/PATCH /v1/devices/:imei/settings/thresholds` — device thresholds

### 13.2 Settings Fixtures

**File:** `apps/VyzoriX_web/src/test/fixtures/vyzor-test-fixtures.ts`

Builders added: `buildThresholds`, `buildClientSettings`,
`buildNotificationSettings`, `buildOperatorSettings`, `buildOrgSettings`,
`buildDeviceSettings`.

### 13.3 Settings Hook Tests

**File:** `apps/VyzoriX_web/src/test/hooks/use-settings-msw.test.ts`

Integration tests exercising all hooks against MSW handlers across all 3 tiers.

---

*Document Version: 2.1*
*Status: API client aligned, hooks + MSW tests implemented*
*Last Updated: 2026-08-15*
*Architecture: Layered (Following `FRONTEND_ARCHITECTURE.md`)*
