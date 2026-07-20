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

### Route Protection Matrix

| Route | Auth Required | Org Required | Description |
|-------|--------------|--------------|-------------|
| **Public Routes** | ❌ | ❌ | |
| `/v1/auth/login` | ❌ | ❌ | Login page |
| `/v1/auth/register` | ❌ | ❌ | Registration |
| `/v1/auth/forgot-password` | ❌ | ❌ | Password reset flow |
| `/v1/auth/reset-password` | ❌ | ❌ | Password reset flow |
| `/v1/auth/refresh` | ❌ | ❌ | Token refresh |
| `/v1/auth/verify-email` | ❌ | ❌ | Email verification |
| `/v1/auth/resend-verification` | ❌ | ❌ | Resend verification |
| `/v1/device/:imei/status` | ❌ | ❌ | Device public status |
| `/v1/device/inbox` | ❌ | ❌ | Device registration inbox |
| `/v1/device/confirm` | ❌ | ❌ | Device confirmation |
| **Authenticated - No Org Required** | ✅ | ❌ | |
| `/v1/auth/me` | ✅ | ❌ | Current operator profile |
| `/v1/auth/me/settings` | ✅ | ❌ | Client settings (Android app behavior) |
| `/v1/auth/me/notifications` | ✅ | ❌ | Personal notification preferences |
| `/v1/auth/mfa/*` | ✅ | ❌ | MFA enrollment, verify, enable/disable, backup codes |
| `/v1/auth/logout` | ✅ | ❌ | Logout |
| **Authenticated - Org Required** | ✅ | ✅ | |
| `/v1/auth/sessions/*` | ✅ | ✅ | Session management (org-scoped) |
| `/v1/auth/client-credentials/*` | ✅ | ✅ | API key management (org-scoped) |
| `/v1/auth/organizations` | ✅ | ❌ | List operator's organizations |
| `/v1/auth/organizations/select` | ✅ | ❌ | Select active organization |
| `/v1/organizations/*` | ✅ | ✅ | All organization routes |
| `/v1/dashboard/*` | ✅ | ✅ | Dashboard, device stats |
| `/v1/devices/*` | ✅ | ✅ | Device management |
| `/v1/command/*` | ✅ | ✅ | Command dispatch |
| `/v1/telemetry/*` | ✅ | ✅ | Telemetry history |
| `/v1/connections/*` | ✅ | ✅ | Connection status |
| `/v1/updates/*` | ✅ | ✅ | Update management |
| `/v1/admin/*` | ✅ | ✅ | SuperAdmin routes |

### Public Routes (No Organization Required)
- `/v1/auth/register`, `/v1/auth/login`
- `/v1/auth/me/settings/client` - Android app settings
- `/v1/auth/me/notifications` - Personal notification prefs

### Settings Subroutes - Detailed Structure

```
/v1/auth/me/settings
├── GET/PATCH                    → Client settings (serverUrl, deviceId, timeouts, HMAC)
└── NO subroutes

/v1/auth/me/notifications
├── GET/PATCH                    → Notification preferences (email, webhook)
├── POST /webhook/test           → Test webhook configuration
├── POST /webhook/rotate         → Rotate webhook secret
└── NO subroutes

/v1/organizations/:id/settings   (REQUIRES ORG)
├── GET/PATCH                    → Org general settings (timezone, date format)
├── GET/PATCH /thresholds        → Default thresholds for new devices
└── NO subroutes
```

### Settings Accessibility by Org Status

**No Organization Created:**
| Settings Area | Accessible | Notes |
|--------------|------------|-------|
| Profile (`/v1/auth/me`) | ✅ | Basic operator info |
| Client Settings (`/v1/auth/me/settings`) | ✅ | Android app behavior |
| Notifications (`/v1/auth/me/notifications`) | ✅ | Personal notification prefs |
| MFA (`/v1/auth/mfa/*`) | ✅ | Auth required, no org |
| Sessions (`/v1/auth/sessions/*`) | 🔒 | Locked - requires org |
| API Keys (`/v1/auth/client-credentials/*`) | 🔒 | Locked - requires org |
| Organization Settings | 🔒 | No org exists yet |
| Device Settings | 🔒 | No org exists yet |

**Organization Exists:**
| Settings Area | Accessible | Notes |
|--------------|------------|-------|
| Profile | ✅ | |
| Client Settings | ✅ | |
| Notifications | ✅ | |
| MFA | ✅ | Auth required, no org |
| Sessions | ✅ | Org-scoped session management |
| API Keys | ✅ | Org-scoped API keys |
| Organization Settings | ✅ | Full access |
| Device Settings | ✅ | Full access |

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

### Public Authentication (No Auth Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /v1/auth/register | Register new operator |
| POST | /v1/auth/login | Login, auto-resolve org |
| POST | /v1/auth/login/tokens | Login for API clients |
| POST | /v1/auth/forgot-password | Request password reset |
| POST | /v1/auth/reset-password | Reset password with token |
| POST | /v1/auth/refresh | Refresh access token |
| GET/POST | /v1/auth/verify-email | Verify email address |
| GET/POST | /v1/auth/resend-verification | Resend verification email |

### Operator Settings (Auth Required - NO Org Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /v1/auth/me | Get current operator |
| PATCH | /v1/auth/me | Update name |
| GET/PATCH | /v1/auth/me/settings | Client settings (Android app behavior) |
| GET/PATCH | /v1/auth/me/notifications | Notification preferences |
| POST | /v1/auth/me/notifications/webhook/test | Test webhook |
| POST | /v1/auth/me/notifications/webhook/rotate | Rotate webhook secret |
| POST | /v1/auth/logout | Logout |

### MFA & Security (Auth Required - NO Org Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /v1/auth/mfa/status | Get MFA status |
| POST | /v1/auth/mfa/enroll | Start MFA enrollment |
| POST | /v1/auth/mfa/verify-setup | Verify MFA setup |
| POST | /v1/auth/mfa/enable | Enable MFA |
| POST | /v1/auth/mfa/disable | Disable MFA |
| POST | /v1/auth/mfa/verify | Verify MFA code |
| POST | /v1/auth/mfa/verify-backup | Verify backup code |
| POST | /v1/auth/mfa/regenerate-backup-codes | Regenerate backup codes |

### Organization Management (Auth + Org Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /v1/auth/organizations | List operator's organizations |
| POST | /v1/auth/organizations/select | Switch active organization |
| GET | /v1/organizations | List operator's orgs |
| POST | /v1/organizations | Create organization |
| GET/PATCH/DELETE | /v1/organizations/:id | Manage organization |
| GET | /v1/organizations/:id/members | List members |
| PATCH/DELETE | /v1/organizations/:id/members/:memberId | Update/remove member |
| POST | /v1/organizations/:id/members/:memberId/transfer | Transfer ownership |
| POST | /v1/organizations/:id/members/:memberId/suspend | Suspend member |
| POST | /v1/organizations/:id/members/:memberId/reinstate | Reinstate member |
| GET | /v1/organizations/:id/invitations | List invitations |
| GET | /v1/invite/:token | Get invitation by token |
| POST | /v1/invite/:token/accept | Accept invitation |
| POST | /v1/invite/:token/reject | Reject invitation |

### Organization Settings (Auth + Org Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET/PATCH | /v1/organizations/:id/settings | Org settings |
| GET/PATCH | /v1/organizations/:id/settings/thresholds | Default thresholds |

### Session & API Key Management (Auth + Org Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /v1/auth/sessions | List sessions |
| GET | /v1/auth/sessions/concurrent | Check concurrent sessions |
| GET/DELETE | /v1/auth/sessions/:id | Get/revoke specific session |
| DELETE | /v1/auth/sessions | Revoke all except current |
| POST | /v1/auth/sessions/revoke-all | Revoke all devices |
| POST | /v1/auth/client-credentials | Create API key |
| GET | /v1/auth/client-credentials | List API keys |
| GET | /v1/auth/client-credentials/:clientId | Get API key details |
| PATCH | /v1/auth/client-credentials/:clientId | Update API key |
| DELETE | /v1/auth/client-credentials/:clientId | Delete API key |
| POST | /v1/auth/client-credentials/:clientId/rotate-secret | Rotate secret |

### Devices (Auth + Org Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /v1/devices | List devices |
| GET/DELETE | /v1/devices/:imei | Device operations |
| GET/PATCH | /v1/devices/:imei/settings | Device settings |
| GET/PATCH | /v1/devices/:imei/thresholds | Device thresholds |
| GET | /v1/dashboard/devices | Dashboard device list |
| GET | /v1/dashboard/devices/operator | Operator's devices |

### Device Lifecycle (Public + Auth + Org)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /v1/device/:imei/status | Public device status |
| POST | /v1/device/inbox | Device registration request (public) |
| GET | /v1/inbox | Auth: List pending registrations |
| GET/PATCH | /v1/inbox/:imei | Auth: Get/update inbox entry |
| POST | /v1/inbox/:imei/ack | Auth: Approve device |
| POST | /v1/inbox/:imei/resend | Auth: Resend approval |
| POST | /v1/device/confirm | Device confirms registration (public) |

### Commands & Telemetry (Auth + Org Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /v1/command/:dispatchId/status | Command status |
| POST | /v1/command/:dispatchId/retry | Retry command |
| DELETE | /v1/command/:dispatchId | Cancel command |
| GET | /v1/telemetry/history | Query telemetry |
| GET | /v1/telemetry/history/export | Export telemetry |
| GET | /v1/telemetry/latest/:deviceId | Latest telemetry |
| GET | /v1/telemetry/stats/:deviceId | Telemetry stats |
| DELETE | /v1/telemetry/cleanup | Cleanup old telemetry |

### Admin & SuperAdmin (Auth + Org + Role Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /v1/admin/clients | List all clients |
| GET/PATCH/DELETE | /v1/admin/clients/:clientId | Manage client |
| POST | /v1/admin/clients/:clientId/rotate-key | Rotate client key |
| GET | /v1/admin/operators | List operators (SuperAdmin) |
| POST | /v1/admin/operators | Create operator (SuperAdmin) |
| GET/PATCH/DELETE | /v1/admin/operators/:id | Manage operator (SuperAdmin) |

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

### Phase 3: Cleanup ✅ DONE
- [x] Remove deprecated operator threshold endpoints (/v1/auth/me/thresholds)
- [x] Remove ThresholdService from GraphQL layer
- [x] Remove operator.thresholds from API responses
- [x] Wire new device/org settings repos to metrics and event processor
- [x] Clean up all backward compatibility code

### Phase 4: Role Management ✅ DONE
- [x] Member role update endpoints (PATCH /v1/organizations/:id/members/:memberId)
- [x] Admin can promote/demote members (role update with permission checks)
- [x] Super_admin protection (cannot change to super_admin, cannot remove last super_admin)
- [x] Suspend member endpoint (POST /v1/organizations/:id/members/:memberId/suspend)
- [x] Reinstate member endpoint (POST /v1/organizations/:id/members/:memberId/reinstate)
- [x] Transfer ownership endpoint (POST /v1/organizations/:id/members/:memberId/transfer)

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

---

## 📊 GRAPHQL API

The GraphQL API provides full parity with the REST API for organization and member management.

### GraphQL Mutations

| Mutation | Description |
|---------|-------------|
| `createOrganization` | Create a new organization |
| `updateOrganization` | Update organization settings |
| `deleteOrganization` | Soft-delete an organization |
| `inviteMember` | Invite a member to an organization |
| `removeMember` | Remove a member from an organization |
| `updateMemberRole` | Promote/demote a member (admin+) |
| `transferOwnership` | Transfer super_admin to another member |
| `suspendMember` | Suspend an active member |
| `reinstateMember` | Reinstate a suspended member |
| `acceptInvitation` | Accept an invitation to join |
| `rejectInvitation` | Reject an invitation |
| `cancelInvitation` | Cancel a pending invitation |
| `transferDevice` | Transfer a device between organizations |

### GraphQL Queries

| Query | Description |
|-------|-------------|
| `myMemberships` | Get current operator's organization memberships |
| `organizations` | List all organizations for the operator |
| `organization` | Get a specific organization |
| `organizationMembers` | List members of an organization |
| `organizationInvitations` | List pending invitations for an org |
| `myInvitations` | List pending invitations for current operator |
| `invitation` | Get invitation by token |

### WebSocket Subscriptions (Real-time Events)

| Subscription | Event Types |
|-------------|-------------|
| `organizationEvents` | `CREATED`, `UPDATED`, `DELETED`, `ACTIVATED`, `DEACTIVATED` |
| `memberEvents` | `member_joined`, `member_invited`, `member_removed`, `member_suspended`, `member_reinstated`, `role_changed`, `ownership_transferred` |
