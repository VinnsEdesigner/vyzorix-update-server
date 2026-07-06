# Settings Page - Enterprise Requirements Specification

> **Version:** 1.0
> **Status:** Draft
> **Created:** 2026-06-24
> **Target:** Production MVP
> **Architecture:** Layered (Following `FRONTEND_ARCHITECTURE.md`)

---

## Table of Contents

1. Overview
2. Page Structure
3. Architecture
4. Target File Structure
5. Tab 1: Overview
6. Tab 2: Connection
7. Tab 3: Operator
8. Tab 4: Thresholds
9. Tab 5: Notifications
10. Tab 6: Advanced
11. Domain Layer
12. Data Layer
13. Presentation Layer - Hooks
14. UI Layer - Components
15. File Changes Summary
16. Implementation Order

---

## 1. Overview

### 1.1 Purpose

The Settings page allows operators to configure their account, connection preferences, notification settings, thresholds, and advanced options.

### 1.2 Current Routes

| Route | File | Status |
|-------|------|--------|
| `/settings` | `settings.tsx` | EXISTS |
| `/settings/connection` | `settings.connection.tsx` | EXISTS |
| `/settings/operator` | `settings.operator.tsx` | EXISTS |
| `/settings/thresholds` | `settings.thresholds.tsx` | EXISTS |
| `/settings/notifications` | `settings.notifications.tsx` | EXISTS |
| `/settings/advanced` | `settings.advanced.tsx` | EXISTS |

### 1.3 Existing Configuration Structure

```typescript
interface VyzorixConfig {
  // Connection
  serverUrl: string;
  deviceId: string;
  dashboardToken: string;
  requestTimeoutMs: number;
  autoReconnect: boolean;
  strictHmac: boolean;
  
  // Thresholds
  thresholds: {
    riskWarn: number;
    riskCrit: number;
    thermalWarn: number;
    thermalCrit: number;
    bufferWarn: number;
    bufferCrit: number;
  };
  
  // Client
  notificationsEnabled: boolean;
  logBufferLimit: number;
  signalHistoryLimit: number;
}
```

---

## 2. Page Structure

### 2.1 Routes (TanStack Start File-Based Routing)

| Route | File | Purpose |
|-------|------|---------|
| `/settings` | `settings.tsx` | Layout with tab navigation |
| `/settings/connection` | `settings.connection.tsx` | Connection settings |
| `/settings/operator` | `settings.operator.tsx` | Account settings |
| `/settings/thresholds` | `settings.thresholds.tsx` | Alert thresholds |
| `/settings/notifications` | `settings.notifications.tsx` | Notification preferences |
| `/settings/advanced` | `settings.advanced.tsx` | Advanced settings |

**Total: 6 routes (1 MODIFIED, 5 EXISTS)**

### 2.2 Navigation Structure

```
Settings
├── Overview (/)          → Redirects to /settings/connection
├── Connection           → Server URL, device ID, timeout
├── Operator            → Account info, display name
├── Thresholds           → Risk, thermal, buffer thresholds
├── Notifications        → Email, push, webhook settings
└── Advanced            → Buffers, danger zone
```

---

## 3. Architecture

### 3.1 Layered Architecture Overview

```
UI Layer (components/) -> Presentation Layer (hooks/) -> Domain Layer (domain/) -> Data Layer (lib/api/)
```

### 3.2 Dependency Rule

- UI Layer can ONLY import from Presentation Layer (hooks)
- Presentation Layer can import from Domain Layer and Data Layer
- Domain Layer can NOT import from any other layer
- Data Layer can import from Domain Layer only

---

## 4. Target File Structure

```
apps/web/src/
|
├── domain/                          # DOMAIN LAYER (follows FRONTEND_ARCHITECTURE.md)
|   ├── _shared/                   # SHARED domain types
|   │   ├── domain-pagination.ts  # Pagination types
|   │   └── domain-errors.ts      # Domain error types
|   │
|   └── settings/
|       ├── settings-entity.ts      # Settings types (settings-specific)
|       ├── settings-mappers.ts    # settingsFromRaw() transformations
|       └── settings-validators.ts # validateSettings()
|
├── lib/
│   └── api/
|       ├── graphql/
|       │   ├── settings/
|       │   │   ├── graphql-settings-queries.ts    # GET_SETTINGS, GET_THRESHOLDS
|       │   │   ├── graphql-settings-mutations.ts # UPDATE_SETTINGS, UPDATE_THRESHOLDS
|       │   │   ├── graphql-settings-fragments.ts # Reusable fragments
|       │   │   └── graphql-settings-types.ts     # Raw GraphQL response types
|       │   └── _shared/
|       │       └── graphql-client.ts    # GraphQL client setup
|       └── rest/
|           ├── settings/
|           │   └── rest-settings-endpoints.ts  # REST endpoints for settings
|           └── _shared/
|               └── rest-client.ts     # REST client setup
|
├── hooks/                           # PRESENTATION LAYER
|   ├── settings/
|   │   ├── use-settings.ts        # Get/update settings
|   │   ├── use-thresholds.ts     # Get/update thresholds
|   │   └── use-notifications.ts   # Get/update notifications
|   └── _shared/
|       └── use-debounce.ts        # Debounced save hook
|
├── components/                      # UI LAYER
|   ├── shared/                    # Shared UI components
|   │   ├── section.tsx            # Bordered section
|   │   ├── section-header.tsx     # Section header
|   │   ├── empty-state.tsx        # Empty state
|   │   ├── loading-skeleton.tsx   # Loading skeleton
|   │   ├── slider-input.tsx       # Slider + input combo
|   │   ├── status-badge.tsx       # Status badge
|   │   └── danger-zone.tsx        # Danger zone wrapper
|   │
|   └── settings/
|       ├── connection-settings.tsx  # Connection tab
|       ├── operator-settings.tsx    # Operator tab
|       ├── threshold-settings.tsx   # Thresholds tab
|       ├── notification-settings.tsx # Notifications tab
|       ├── advanced-settings.tsx    # Advanced tab
|       ├── threshold-input.tsx      # Threshold input component
|       └── notification-row.tsx     # Notification toggle row
|
└── routes/                         # PAGE LAYER (Routes)
    ├── settings.tsx                # Layout with tabs
    ├── settings.connection.tsx
    ├── settings.operator.tsx
    ├── settings.thresholds.tsx
    ├── settings.notifications.tsx
    └── settings.advanced.tsx
```

---

## 5. Tab 1: Overview

### 5.1 Purpose

Landing page for settings, showing summary of current configuration and quick links to sections.

### 5.2 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  SETTINGS                                                        │
├─────────────────────────────────────────────────────────────────────┤
│  ┌─ QUICK STATUS ──────────────────────────────────────────────┐   │
│  │  Server: https://api.example.com       [Edit]              │   │
│  │  Device: Pixel 8 Pro                  [Edit]              │   │
│  │  Notifications: Enabled                [Toggle]             │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ SECTIONS ─────────────────────────────────────────────────┐   │
│  │  [Connection]  [Operator]  [Thresholds]  [Notifications] │   │
│  │  [Advanced]                                            [Advanced] │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

### 5.3 Data Displayed

| Field | Source | Description |
|-------|--------|-------------|
| Server URL | config.serverUrl | Current API server |
| Device ID | config.deviceId | Current device |
| Notifications | config.notificationsEnabled | Notification status |
| Last Sync | server | Last GitHub sync time |

---

## 6. Tab 2: Connection

### 6.1 Purpose

Configure server connection and device settings.

### 6.2 Current Fields

| Field | Type | Description |
|-------|------|-------------|
| Server URL | input (URL) | API server endpoint |
| Device ID | input (text) | Device identifier |
| Dashboard Token | input (password) | Dashboard auth token |
| Request Timeout | input (number) | Request timeout in ms |
| Auto Reconnect | toggle | Auto-reconnect on disconnect |
| Strict HMAC | toggle | Enforce HMAC validation |

### 6.3 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  CONNECTION                                                    │
├─────────────────────────────────────────────────────────────────────┤
│  ┌─ SERVER ────────────────────────────────────────────────────┐   │
│  │  Server URL                                                │   │
│  │  ┌─────────────────────────────────────────────────┐     │   │
│  │  │ https://api.example.com                         │     │   │
│  │  └─────────────────────────────────────────────────┘     │   │
│  │  Health: ● ok                                            │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ DEVICE ───────────────────────────────────────────────────┐   │
│  │  Device ID                                               │   │
│  │  ┌─────────────────────────────────────────────────┐     │   │
│  │  │ 861234567890123                                 │     │   │
│  │  └─────────────────────────────────────────────────┘     │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ AUTHENTICATION ──────────────────────────────────────────┐   │
│  │  Dashboard Token                                        │   │
│  │  ┌─────────────────────────────────────────────────┐     │   │
│  │  │ ••••••••••••••••                               │     │   │
│  │  └─────────────────────────────────────────────────┘     │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ ADVANCED ───────────────────────────────────────────────┐   │
│  │  Request Timeout (ms)                    [5000        ]  │   │
│  │  Auto Reconnect                         [Toggle: ON]    │   │
│  │  Strict HMAC Validation                 [Toggle: OFF]   │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  [Save Changes]                                                    │
└─────────────────────────────────────────────────────────────────────┘
```

### 6.4 Validation Rules

| Field | Rule |
|-------|------|
| Server URL | Required, valid HTTP/HTTPS URL |
| Device ID | Optional, alphanumeric |
| Dashboard Token | Optional |
| Request Timeout | 500-60000ms |
| Auto Reconnect | Boolean |
| Strict HMAC | Boolean |

---

## 7. Tab 3: Operator

### 7.1 Purpose

Manage operator account information and permissions.

### 7.2 Current Fields

| Field | Type | Description |
|-------|------|-------------|
| Display Name | input (text) | Operator display name |
| Email | display | Operator email (read-only) |
| Role | display | admin/operator/super_admin |
| Permissions | list | Feature permissions |

### 7.3 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  OPERATOR                                                    │
├─────────────────────────────────────────────────────────────────────┤
│  ┌─ ACCOUNT ─────────────────────────────────────────────────┐   │
│  │  Display Name                                           │   │
│  │  ┌─────────────────────────────────────────────────┐     │   │
│  │  │ John Doe                                         │     │   │
│  │  └─────────────────────────────────────────────────┘     │   │
│  │  Auto-saves on change                                    │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ IDENTITY ─────────────────────────────────────────────────┐   │
│  │  Email         operator@example.com                        │   │
│  │  Role          admin                                     │   │
│  │  Member Since  2024-01-15                               │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ PERMISSIONS ──────────────────────────────────────────────┐   │
│  │  ● Register Devices      ● Deregister Devices             │   │
│  │  ● Send Commands         ● View Telemetry                 │   │
│  │  ● Push Updates         ● Manage Settings                │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 8. Tab 4: Thresholds

### 8.1 Purpose

Configure alert thresholds for risk, thermal, and buffer levels.

### 8.2 Current Fields

| Field | Type | Range | Default |
|-------|------|-------|---------|
| Risk Warning | slider | 0-100 | 70 |
| Risk Critical | slider | 0-100 | 85 |
| Thermal Warning | slider | 0-100 | 45 |
| Thermal Critical | slider | 0-100 | 50 |
| Buffer Warning | slider | 0-100 | 30 |
| Buffer Critical | slider | 0-100 | 15 |

### 8.3 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  THRESHOLDS                                                   │
├─────────────────────────────────────────────────────────────────────┤
│  ┌─ RISK SCORE ───────────────────────────────────────────────┐   │
│  │                                                          │   │
│  │  Warning                              Critical            │   │
│  │  ○────────●─────────────────────────────○───────────────│   │
│  │  70                                  85                  │   │
│  │                                                          │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ THERMAL TEMPERATURE ─────────────────────────────────────┐   │
│  │                                                          │   │
│  │  Warning                              Critical            │   │
│  │  ○─────────────●─────────────────────────○──────────────│   │
│  │  45                                  50                  │   │
│  │                                                          │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ BUFFER LEVEL ───────────────────────────────────────────┐   │
│  │                                                          │   │
│  │  Critical                             Warning             │   │
│  │  ○─────●─────────────────────────────○──────────────────│   │
│  │  15                                 30                  │   │
│  │  (Inverted: low buffer is bad)                          │   │
│  │                                                          │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  [Reset to Defaults]                        [Save Changes]           │
└─────────────────────────────────────────────────────────────────────┘
```

### 8.4 Threshold Relationships

| Metric | Warning > Critical | Notes |
|--------|-------------------|-------|
| Risk Score | Warning > Critical | Higher is worse |
| Thermal | Warning > Critical | Higher is worse |
| Buffer | Warning > Critical | Lower is worse (inverted) |

---

## 9. Tab 5: Notifications

### 9.1 Purpose

Configure notification preferences (email, push, webhook).

### 9.2 Notification Types

| Type | Description |
|------|-------------|
| threshold_breach | Risk/thermal threshold exceeded |
| device_offline | Device went offline |
| device_online | Device came online |
| update_available | New APK version available |
| command_failed | Command delivery failed |
| registration_request | New device pending registration |

### 9.3 Notification Channels

| Channel | Description |
|---------|-------------|
| email | Email notifications |
| push | Browser push notifications |
| webhook | HTTP webhook calls |

### 9.4 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  NOTIFICATIONS                                                │
├─────────────────────────────────────────────────────────────────────┤
│  ┌─ EMAIL ───────────────────────────────────────────────────┐   │
│  │  [Enable Email Notifications]                             │   │
│  │                                                          │   │
│  │  ● threshold_breach                                     │   │
│  │  ● device_offline                                       │   │
│  │  ● device_online                                       │   │
│  │  ○ update_available                                    │   │
│  │  ○ command_failed                                      │   │
│  │  ○ registration_request                                │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ PUSH ─────────────────────────────────────────────────┐   │
│  │  [Enable Push Notifications]                            │   │
│  │                                                          │   │
│  │  ● threshold_breach                                   │   │
│  │  ○ device_offline                                     │   │
│  │  ○ device_online                                       │   │
│  │  ○ update_available                                    │   │
│  │  ○ command_failed                                      │   │
│  │  ○ registration_request                                │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ WEBHOOK ────────────────────────────────────────────────┐   │
│  │  [Enable Webhook]                                      │   │
│  │                                                          │   │
│  │  Webhook URL                                           │   │
│  │  ┌─────────────────────────────────────────────────┐     │   │
│  │  │ https://hooks.example.com/vyzorix              │     │   │
│  │  └─────────────────────────────────────────────────┘     │   │
│  │                                                          │   │
│  │  [Test Webhook]                                         │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 10. Tab 6: Advanced

### 10.1 Purpose

Configure advanced client settings and access danger zone.

### 10.2 Current Fields

| Field | Type | Range | Default |
|-------|------|-------|---------|
| Log Buffer Limit | input | 50-5000 | 500 |
| Signal History Limit | input | 30-2000 | 240 |

### 10.3 Danger Zone

| Action | Description | Permission |
|--------|-------------|------------|
| Reset to Defaults | Reset all settings to server defaults | super_admin |
| Delete Account | Permanently delete operator account | super_admin |
| Export Settings | Export current configuration | admin |

### 10.4 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  ADVANCED                                                     │
├─────────────────────────────────────────────────────────────────────┤
│  ┌─ BUFFERS ────────────────────────────────────────────────┐   │
│  │                                                          │   │
│  │  Log Retention (entries)     [500              ]       │   │
│  │  Signal History (entries)    [240              ]       │   │
│  │                                                          │   │
│  │  Memory limits for in-browser signal and log buffers    │   │
│  │                                                          │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ DANGER ZONE ──────────────────────────────────────────────┐   │
│  │                                                          │   │
│  │  ⚠️  These actions are irreversible                     │   │
│  │                                                          │   │
│  │  [Reset All Settings to Defaults]                       │   │
│  │  [Export Configuration]                                 │   │
│  │                                                          │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 11. Domain Layer

### 11.1 Files to CREATE

| File | Status | Purpose |
|------|--------|---------|
| `domain/shared/types.ts` | NEW | Shared types |
| `domain/common/error.ts` | NEW | Domain errors |
| `domain/settings/settings-types.ts` | NEW | Settings types |
| `domain/settings/settings-transforms.ts` | NEW | settingsFromRaw() |
| `domain/settings/settings-validation.ts` | NEW | validateSettings() |

### 11.2 Types

```typescript
// domain/settings/settings-types.ts

// Threshold levels
export interface Thresholds {
  riskWarn: number;
  riskCrit: number;
  thermalWarn: number;
  thermalCrit: number;
  bufferWarn: number;
  bufferCrit: number;
}

// Notification settings
export enum NotificationType {
  THRESHOLD_BREACH = "threshold_breach",
  DEVICE_OFFLINE = "device_offline",
  DEVICE_ONLINE = "device_online",
  UPDATE_AVAILABLE = "update_available",
  COMMAND_FAILED = "command_failed",
  REGISTRATION_REQUEST = "registration_request",
}

export enum NotificationChannel {
  EMAIL = "email",
  PUSH = "push",
  WEBHOOK = "webhook",
}

export interface NotificationSettings {
  enabled: boolean;
  channels: NotificationChannel[];
  types: NotificationType[];
  webhookUrl?: string;
}

// Connection settings
export interface ConnectionSettings {
  serverUrl: string;
  deviceId: string;
  dashboardToken?: string;
  requestTimeoutMs: number;
  autoReconnect: boolean;
  strictHmac: boolean;
}

// Advanced settings
export interface AdvancedSettings {
  logBufferLimit: number;
  signalHistoryLimit: number;
}

// Full settings
export interface Settings {
  thresholds: Thresholds;
  notifications: NotificationSettings;
  connection: ConnectionSettings;
  advanced: AdvancedSettings;
}
```

---

## 12. Data Layer

### 12.1 REST Endpoints to CREATE

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/auth/me/settings` | GET | Get current settings |
| `/v1/auth/me/settings` | PATCH | Update settings |
| `/v1/auth/me/settings` | POST | Reset settings to defaults |
| `/v1/auth/me/thresholds` | GET | Get thresholds |
| `/v1/auth/me/thresholds` | PATCH | Update thresholds |
| `/v1/auth/me/notifications` | GET | Get notification settings |
| `/v1/auth/me/notifications` | PATCH | Update notification settings |
| `/v1/auth/me/notifications/webhook/test` | POST | Test webhook endpoint |
| `/v1/auth/me/notifications/webhook/rotate` | POST | Rotate webhook secret |

### 12.2 REST Implementation

```typescript
// lib/api/rest/settings-rest.ts

const BASE = "/v1/auth/me";

export async function fetchSettings(serverUrl: string): Promise<Settings> {
  const res = await fetch(join(serverUrl, `${BASE}/settings`), {
    credentials: "include",
  });
  if (!res.ok) throw new Error(`Settings fetch failed: ${res.status}`);
  return res.json();
}

export async function updateSettings(
  serverUrl: string,
  settings: Partial<Settings>
): Promise<Settings> {
  const res = await fetch(join(serverUrl, `${BASE}/settings`), {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(settings),
  });
  if (!res.ok) throw new Error(`Settings update failed: ${res.status}`);
  return res.json();
}

export async function updateThresholds(
  serverUrl: string,
  thresholds: Thresholds
): Promise<Thresholds> {
  const res = await fetch(join(serverUrl, `${BASE}/thresholds`), {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(thresholds),
  });
  if (!res.ok) throw new Error(`Thresholds update failed: ${res.status}`);
  return res.json();
}

export async function testWebhook(
  serverUrl: string,
  url: string
): Promise<{ success: boolean; statusCode?: number; responseTime?: number; error?: string }> {
  const res = await fetch(join(serverUrl, `${BASE}/notifications/webhook/test`), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ url }),
  });
  if (!res.ok) throw new Error(`Webhook test failed: ${res.status}`);
  return res.json();
}

export async function rotateWebhookSecret(serverUrl: string): Promise<{ secret: string }> {
  const res = await fetch(join(serverUrl, `${BASE}/notifications/webhook/rotate`), {
    method: "POST",
    credentials: "include",
  });
  if (!res.ok) throw new Error(`Webhook secret rotation failed: ${res.status}`);
  return res.json();
}
```

---

## 13. Presentation Layer - Hooks

### 13.1 Hooks to CREATE

| File | Status | Purpose |
|------|--------|---------|
| `hooks/settings/use-settings.ts` | NEW | Get/update settings |
| `hooks/settings/use-thresholds.ts` | NEW | Get/update thresholds |
| `hooks/settings/use-notifications.ts` | NEW | Get/update notifications |
| `hooks/settings/use-debounce.ts` | NEW | Debounced save |
| `hooks/settings/index.ts` | NEW | Barrel export |

### 13.2 Hook Implementations

```typescript
// hooks/settings/use-settings.ts
export function useSettings() {
  const query = useQuery({
    queryKey: ["settings"],
    queryFn: () => fetchSettings(),
  });

  const mutation = useMutation({
    mutationFn: (settings: Partial<Settings>) => updateSettings(settings),
    onSuccess: () => {
      query.invalidate();
    },
  });

  return {
    settings: query.data,
    isLoading: query.isLoading,
    isError: query.isError,
    updateSettings: mutation.mutate,
    isUpdating: mutation.isPending,
  };
}

// hooks/settings/use-thresholds.ts
export function useThresholds() {
  const query = useQuery({
    queryKey: ["settings", "thresholds"],
    queryFn: () => fetchThresholds(),
  });

  const mutation = useMutation({
    mutationFn: (thresholds: Thresholds) => updateThresholds(thresholds),
    onSuccess: () => {
      query.invalidate();
    },
  });

  return {
    thresholds: query.data,
    isLoading: query.isLoading,
    updateThresholds: mutation.mutate,
    isUpdating: mutation.isPending,
  };
}

// hooks/settings/use-debounce.ts
export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedValue(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);

  return debouncedValue;
}
```

---

## 14. UI Layer - Components

### 14.1 Components to CREATE

| File | Status | Purpose |
|------|--------|---------|
| `components/shared/slider-input.tsx` | NEW | Slider + input combo |
| `components/shared/danger-zone.tsx` | NEW | Danger zone wrapper |
| `components/settings/threshold-input.tsx` | NEW | Threshold slider component |
| `components/settings/notification-row.tsx` | NEW | Notification toggle |
| `components/settings/notification-settings.tsx` | NEW | Notifications tab |

### 14.2 Component Implementations

```typescript
// components/shared/slider-input.tsx
interface SliderInputProps {
  label: string;
  value: number;
  onChange: (value: number) => void;
  min: number;
  max: number;
  step?: number;
  unit?: string;
}

export function SliderInput({
  label,
  value,
  onChange,
  min,
  max,
  step = 1,
  unit = "",
}: SliderInputProps) {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium">{label}</label>
        <Input
          type="number"
          value={value}
          onChange={(e) => onChange(Number(e.target.value))}
          className="w-20"
        />
      </div>
      <Slider
        value={[value]}
        onValueChange={([v]) => onChange(v)}
        min={min}
        max={max}
        step={step}
      />
    </div>
  );
}

// components/settings/threshold-input.tsx
interface ThresholdInputProps {
  label: string;
  warningValue: number;
  criticalValue: number;
  onWarningChange: (value: number) => void;
  onCriticalChange: (value: number) => void;
  min?: number;
  max?: number;
  inverted?: boolean;
}

export function ThresholdInput({
  label,
  warningValue,
  criticalValue,
  onWarningChange,
  onCriticalChange,
  min = 0,
  max = 100,
  inverted = false,
}: ThresholdInputProps) {
  return (
    <div className="space-y-3">
      <div className="text-sm font-medium">{label}</div>
      <div className="flex items-center gap-4">
        <div className="flex-1">
          <div className="text-xs text-muted-foreground mb-1">
            {inverted ? "Critical" : "Warning"}
          </div>
          <SliderInput
            value={criticalValue}
            onChange={onCriticalChange}
            min={min}
            max={max}
          />
        </div>
        <span className="text-muted-foreground">→</span>
        <div className="flex-1">
          <div className="text-xs text-muted-foreground mb-1">
            {inverted ? "Warning" : "Critical"}
          </div>
          <SliderInput
            value={warningValue}
            onChange={onWarningChange}
            min={min}
            max={max}
          />
        </div>
      </div>
    </div>
  );
}
```

---

## 15. File Changes Summary

### 15.1 Total File Count

| Category | New | Modified | Total |
|----------|-----|----------|-------|
| Domain Layer | 4 | 0 | 4 |
| Data Layer (REST) | 1 | 0 | 1 |
| Presentation Layer | 4 | 0 | 4 |
| UI Layer (Shared) | 2 | 0 | 2 |
| UI Layer (Settings) | 2 | 0 | 2 |
| Routes | 0 | 1 | 1 |
| **TOTAL** | **13** | **1** | **14** |

### 15.2 All Files Listed

#### Domain Layer (4 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `domain/shared/types.ts` | NEW | Shared types |
| `domain/common/error.ts` | NEW | Domain errors |
| `domain/settings/settings-types.ts` | NEW | Settings types |
| `domain/settings/settings-transforms.ts` | NEW | settingsFromRaw() |
| `domain/settings/settings-validation.ts` | NEW | validateSettings() |

#### Data Layer - REST (1 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `lib/api/rest/settings-rest.ts` | NEW | REST endpoints |

#### Presentation Layer (4 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `hooks/settings/use-settings.ts` | NEW | Settings hook |
| `hooks/settings/use-thresholds.ts` | NEW | Thresholds hook |
| `hooks/settings/use-notifications.ts` | NEW | Notifications hook |
| `hooks/settings/use-debounce.ts` | NEW | Debounce utility |
| `hooks/settings/index.ts` | NEW | Barrel export |

#### UI Layer - Shared (2 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `components/shared/slider-input.tsx` | NEW | Slider + input |
| `components/shared/danger-zone.tsx` | NEW | Danger zone |

#### UI Layer - Settings (2 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `components/settings/threshold-input.tsx` | NEW | Threshold component |
| `components/settings/notification-row.tsx` | NEW | Notification toggle |

#### Routes (1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| `routes/settings-page.tsx` | MODIFIED | Update layout if needed |

---

## 16. Implementation Order

### Phase 1: Domain Layer (Day 1)
1. Create `domain/shared/` types
2. Create `domain/settings/settings-types.ts`
3. Create transforms and validation

### Phase 2: Data Layer (Day 1)
1. Create REST client `lib/api/rest/settings-rest.ts`
2. Test endpoints

### Phase 3: Presentation Layer (Day 1-2)
1. Create `use-settings` hook
2. Create `use-thresholds` hook
3. Create `use-notifications` hook
4. Create `use-debounce` utility

### Phase 4: UI Layer - Shared (Day 2)
1. Create `SliderInput` component
2. Create `DangerZone` component

### Phase 5: UI Layer - Settings (Day 2)
1. Create `ThresholdInput` component
2. Create `NotificationRow` component
3. Update existing route components to use hooks

### Phase 6: Integration (Day 2-3)
1. Wire hooks to components
2. Add loading/error states
3. Test auto-save functionality

---

*Document Version: 1.0*
*Status: Ready for Implementation*
*Architecture: Layered (Following FRONTEND_ARCHITECTURE.md)*
