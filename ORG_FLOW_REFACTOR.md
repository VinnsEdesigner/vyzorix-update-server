# Organization Flow Refactor - Complete Specification

> **Created:** 2026-07-16
> **Updated:** 2026-07-17
> **Status:** In Progress
> **Target:** Complete multi-tenant organization model with settings hierarchy
> **Model:** Multi-Tenant Organization (per REFACTOR_PLAN.md)

---

## 🎯 MODEL OVERVIEW

### Core Principles
- **Operators** are global identities (email, password, OAuth) - **NO global role**
- **Organizations** are tenants that own resources
- **Memberships** link operators to organizations with scoped roles
- **Role is ONLY determined when an organization is selected**
- **MFA/Backup Codes remain per-operator (global)**
- **Settings are hierarchical**: Operator → Organization → Device

### About Devices
**Devices are Android phones running the Vyzorix APK**, not industrial sensors:
- They connect via WebSocket (WSS) to the server
- They use HMAC-SHA256 signing for command verification
- They receive FCM push notifications for commands
- Device settings control the Android app behavior, not hardware sensors

### Data Model

```
OPERATOR (Global Identity)
├── id, email, name, password_hash
├── mfa_enabled, mfa_secret, backup_codes
├── last_organization_id (for auto-select)
└── client_settings (Android app behavior)

ORGANIZATION (Tenant - owns all resources)
├── id, name, description, created_by
├── max_members, is_active
└── organization_settings
     ├── default_thresholds
     ├── timezone, date_format
     └── alert_cooldown_minutes

DEVICE (Android Phone with Vyzorix APK)
├── imei, organization_id, registered_by
├── fcm_token, firebase_install_id
├── command_secret (HMAC key)
├── state (REGISTERED, ONLINE, OFFLINE, DEREGISTERED)
└── device_settings
     ├── thresholds (NULL = use org defaults)
     └── custom_name, location
```

---

## 🔐 AUTHENTICATION & SESSION FLOW

### Login Flow
```
POST /v1/auth/login
├── Validate email/password
├── Create Session
└── Auto-resolve organization:
    ├── 0 memberships → needs_organization: true
    ├── 1 membership → auto-select it
    ├── 2+ with LastOrganizationID → auto-select that (if valid)
    └── 2+ without LastOrganizationID → needs_organization: true
```

### Login Response
```json
{
  "operator_id": "op_xxx",
  "email": "user@example.com",
  "name": "John Doe",
  "mfa_enabled": false,
  "needs_organization": true,
  "organizations": [{"id": "org_xxx", "name": "My Company", "role": "super_admin"}],
  "last_organization_id": "org_xxx",
  "selected_organization": {"id": "org_xxx", "name": "My Company", "role": "super_admin"}
}
```

---

## 🏢 ORGANIZATION CREATION

### POST /v1/organizations

**Request:**
```json
{
  "name": "My Company",
  "description": "We build IoT devices",
  "maxMembers": 50,
  "role": "super_admin"
}
```

**Validation:**
| Field | Rules |
|-------|-------|
| name | Min 2 chars, max 255. Default: "personal" |
| description | Required, non-empty string |
| maxMembers | Min 1, default 100 |
| role | Must be exactly "super_admin" or "admin" |

---

## 🛡️ ROUTE PROTECTION

### Middleware Chain (in order)
1. **CookieAuth** - Validates session, sets operator + session in context
2. **OrganizationContext** - Extracts org_id from session
3. **OrganizationMembership** - Validates operator is member

### Protected Routes (Require Organization)
- `/dashboard/*` - Dashboard, devices, stats
- `/devices/*` - Device management
- `/v1/organizations/:id/*` - Org management

### Public Routes (No Organization Required)
- `/v1/auth/register`, `/v1/auth/login`
- `/v1/auth/me/settings/client` - Android app settings
- `/v1/auth/me/notifications` - Personal notification prefs

---

## ⚙️ SETTINGS HIERARCHY

### Level 1: Operator Settings (Global)

```
GET/PATCH /v1/auth/me
GET/PATCH /v1/auth/me/settings/client
GET/PATCH /v1/auth/me/notifications
```

---

## 📱 CLIENT SETTINGS (Operator-Level)

These settings control the **Android app behavior** on the device.

### GET/PATCH /v1/auth/me/settings/client

**Settings:**
| Setting | Type | Description |
|---------|------|-------------|
| serverUrl | string | WebSocket server URL |
| deviceId | string | Device identifier (UUID) |
| requestTimeoutMs | int | HTTP request timeout (500-60000) |
| autoReconnect | boolean | Automatically reconnect if WebSocket disconnects |
| strictHmac | boolean | Strictly validate HMAC signatures on all commands |
| logBufferLimit | int | Maximum log entries to buffer (50-5000) |
| signalHistoryLimit | int | Signal history entries to retain (30-2000) |

**Example:**
```json
{
  "serverUrl": "wss://updates.vyzorix.com/v1/ws",
  "deviceId": "550e8400-e29b-41d4-a716-446655440000",
  "requestTimeoutMs": 5000,
  "autoReconnect": true,
  "strictHmac": true,
  "logBufferLimit": 1000,
  "signalHistoryLimit": 500
}
```

---

### Level 2: Organization Settings

```
GET/PATCH /v1/organizations/:id/settings
GET/PATCH /v1/organizations/:id/settings/thresholds
```

**Organization Settings:**
```json
{
  "timezone": "UTC",
  "dateFormat": "YYYY-MM-DD",
  "alertCooldownMinutes": 15,
  "defaultThresholds": {
    "riskWarn": 70, "riskCrit": 90,
    "thermalWarn": 75, "thermalCrit": 85,
    "bufferWarn": 30, "bufferCrit": 10
  }
}
```

---

### Level 3: Device Settings

```
GET/PATCH /v1/devices/:imei/settings
GET/PATCH /v1/devices/:imei/thresholds
```

---

## 📊 THRESHOLDS SPECIFICATION

### Default Values
```json
{
  "riskWarn": 70, "riskCrit": 90,
  "thermalWarn": 75, "thermalCrit": 85,
  "bufferWarn": 30, "bufferCrit": 10
}
```

### Threshold Resolution: device → org → default

---

## 🔄 API ENDPOINT SUMMARY

### Authentication (No Org Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /v1/auth/register | Register new operator |
| POST | /v1/auth/login | Login, auto-resolve org |
| POST | /v1/auth/login/tokens | Login for API clients |
| POST | /v1/auth/organizations/select | Switch organization |
| GET | /v1/auth/me | Get current operator |
| PATCH | /v1/auth/me | Update name |

### Operator Settings (No Org Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/PATCH | /v1/auth/me/settings/client | Android app settings |
| GET/PATCH | /v1/auth/me/notifications | Notification prefs |

### Organizations (Org Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /v1/organizations | List operator's orgs |
| POST | /v1/organizations | Create organization |
| GET/PATCH/DELETE | /v1/organizations/:id | Manage organization |

### Organization Settings (Org Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/PATCH | /v1/organizations/:id/settings | Org settings |
| GET/PATCH | /v1/organizations/:id/settings/thresholds | Org defaults |

### Devices (Org Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /v1/devices | List devices |
| GET/DELETE | /v1/devices/:imei | Device operations |
| GET/PATCH | /v1/devices/:imei/settings | Device settings |
| GET/PATCH | /v1/devices/:imei/thresholds | Device thresholds |

---

## 🗄️ DATABASE SCHEMA

### organization_settings (NEW)
```sql
CREATE TABLE organization_settings (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL UNIQUE,
    timezone TEXT DEFAULT 'UTC',
    date_format TEXT DEFAULT 'YYYY-MM-DD',
    alert_cooldown_minutes INTEGER DEFAULT 15,
    default_thresholds JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);
```

### device_settings (NEW)
```sql
CREATE TABLE device_settings (
    id TEXT PRIMARY KEY,
    device_imei TEXT NOT NULL UNIQUE,
    custom_name TEXT,
    location TEXT,
    metadata JSONB,
    thresholds JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (device_imei) REFERENCES devices(imei) ON DELETE CASCADE
);
```

---

## 📋 IMPLEMENTATION CHECKLIST

### Phase 1: Core Organization Flow ✅ DONE
- [x] Operator struct (no global role)
- [x] Organization entity with description
- [x] Membership linking operators to orgs
- [x] CreateOrganization endpoint with role validation
- [x] Login with auto-org resolution
- [x] Session tracking of selected organization
- [x] Middleware protection for dashboard routes
- [x] Organization selection endpoint
- [x] Description required validation

### Phase 2: Settings Foundation ✅ DONE
- [x] Operator client settings (Android app behavior)
- [x] Operator notification preferences
- [x] organization_settings table
- [x] organization_settings CRUD endpoints
- [x] Move thresholds to org level defaults
- [x] device_settings table
- [x] device_settings CRUD endpoints
- [x] Threshold resolution: device → org → default

### Phase 3: Cleanup (Deferred)
- [ ] Remove operator.thresholds column
- [ ] Update all API paths to new structure

### Phase 4: Role Management (Deferred)
- [ ] Member role update endpoints
- [ ] Admin can promote/demote members
- [ ] Super_admin protection

---

## 🔑 ROLE SYSTEM

| Role | Level | Permissions |
|------|-------|-------------|
| super_admin | 4 | Full access, manage admins, delete org |
| admin | 3 | Manage members, devices, settings |
| operator | 2 | View devices, send commands |
| viewer | 1 | Read-only access |

**Role is determined by:** Active session's SelectedOrganizationID + OrganizationMember lookup

---

## 📱 DEVICE LIFECYCLE

### Device States
| State | Description |
|-------|-------------|
| REGISTERED | Initial state after first registration |
| ONLINE | WebSocket connected or telemetry received |
| OFFLINE | WebSocket disconnected |
| DEREGISTERED | Terminal state after DELETE |

### Registration Flow
```
1. Device: POST /v1/device/inbox (registration request)
2. Server: Stores in INBOX (pending)
3. Operator: Views inbox, clicks Register
4. Server: Generates commandSecret, FCM push
5. Device: POST /v1/device/confirm
6. Server: Marks as REGISTERED
```

---

## ✅ VERIFIED IMPLEMENTATIONS

### Create Organization
- ✅ Role validation: only "super_admin" or "admin" accepted
- ✅ Description required
- ✅ Name defaults to "personal"
- ✅ MaxMembers defaults to 100
- ✅ Operator added as member
- ✅ LastOrganizationID updated

### Login Auto-Selection
- ✅ 0 memberships → needs_organization: true
- ✅ 1 membership → auto-selected
- ✅ 2+ with valid LastOrganizationID → auto-selected
- ✅ 2+ without LastOrganizationID → needs_organization: true

### Route Protection
- ✅ Dashboard routes require organization context
- ✅ Device routes require organization context
- ✅ Organization routes require membership

### Session Tracking
- ✅ Session.SelectedOrganizationID field exists
- ✅ Updated on login
- ✅ Updated on organization switch
