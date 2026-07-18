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

### Data Model

```
OPERATOR (Global Identity)
├── id, email, name, password_hash
├── mfa_enabled, mfa_secret, backup_codes
├── last_organization_id (for auto-select)
└── client_settings (server_url, timeouts, etc.)
        │
        │ is member of (via Membership)
        ▼
ORGANIZATION (Tenant - owns all resources)
├── id, name, description, created_by
├── max_members, is_active
└── organization_settings
     ├── default_thresholds (risk, thermal, buffer)
     ├── timezone, date_format
     └── alert_cooldown_minutes
        │
        │ has (via Device.organization_id)
        ▼
DEVICE (Owned by organization)
├── imei, organization_id, registered_by
└── device_settings
     ├── thresholds (NULL = use org defaults)
     ├── custom_name, location, metadata
     └── ...
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

### Organization Selection
```
POST /v1/auth/organizations/select
Body: { "organization_id": "org_xxx" }

→ Updates session.SelectedOrganizationID
→ Updates operator.LastOrganizationID
```

---

## 🏢 ORGANIZATION CREATION

### POST /v1/organizations

**Request:**
```json
{
  "name": "My Company",                    // Optional, defaults to "personal"
  "description": "We build IoT devices",  // REQUIRED
  "maxMembers": 50,                       // Optional, defaults to 100
  "role": "super_admin"                   // REQUIRED: "super_admin" or "admin"
}
```

**Validation:**
| Field | Rules |
|-------|-------|
| name | Min 2 chars, max 255. Default: "personal" |
| description | Required, non-empty string |
| maxMembers | Min 1, default 100 |
| role | Must be exactly "super_admin" or "admin" |

**Response (201):**
```json
{
  "id": "org_abc123...",
  "name": "My Company",
  "description": "We build IoT devices",
  "created_by": "op_xyz789...",
  "created_at": "2026-07-17T01:15:00Z",
  "max_members": 50,
  "is_active": true
}
```

---

## 🛡️ ROUTE PROTECTION

### Middleware Chain (in order)
1. **CookieAuth** - Validates session, sets operator + session in context
2. **OrganizationContext** - Extracts org_id: query → header → context → session.SelectedOrganizationID
3. **OrganizationMembership** - Validates operator is member of organization

### Protected Routes (Require Organization)
- `/dashboard/*` → Dashboard, devices, stats
- `/devices/*` → Device management
- `/v1/organizations/:id/*` → Org management

### Public Routes (No Organization Required)
- `/v1/auth/register`, `/v1/auth/login`
- `/v1/auth/me/settings/client` → Client settings (operator-level)
- `/v1/auth/me/notifications` → Personal notification prefs

---

## ⚙️ SETTINGS HIERARCHY

### Level 1: Operator Settings (Global)
Available without organization:
```
GET/PATCH /v1/auth/me              → Profile (name)
/v1/auth/me/settings/client        → Client settings
/v1/auth/me/notifications          → Personal notification prefs
```

**Client Settings (Operator-Level):**
```json
{
  "serverUrl": "wss://api.vyzorix.com",
  "deviceId": "device-001",
  "requestTimeoutMs": 5000,
  "autoReconnect": true,
  "strictHmac": true,
  "logBufferLimit": 1000,
  "signalHistoryLimit": 500
}
```

**Rationale:** Device/client configuration preferences, not org preferences.

---

### Level 2: Organization Settings
Requires organization context:
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

### Level 3: Device Settings (Per-Device Override)
Requires organization + device:
```
GET/PATCH /v1/devices/:imei/settings
GET/PATCH /v1/devices/:imei/thresholds
```

**Device Settings:**
```json
{
  "customName": "Factory Floor Sensor A",
  "location": "Building A, Floor 2",
  "thresholds": {
    // NULL = use organization defaults
    "riskWarn": 80,
    "riskCrit": 95
  }
}
```

### Threshold Resolution Flow
```
Device Alert Triggered
         │
         ▼
┌────────────────────────────────┐
│ Check device.thresholds        │
│ (device_settings table)        │
└────────────┬───────────────────┘
             │ NULL?
      ┌──────┴──────┐
      Yes          No
         │          ▼
         │    Use device thresholds
         │
         ▼
┌────────────────────────────────┐
│ Check organization defaults     │
└────────────────────────────────┘
```

---

## 📊 THRESHOLDS SPECIFICATION

### Threshold Types
| Threshold | Rule | Description |
|-----------|------|-------------|
| riskWarn | < riskCrit | Device risk score |
| thermalWarn | < thermalCrit | Temperature (°C) |
| bufferWarn | > bufferCrit | Buffer level (inverted) |

### Default Values
```json
{
  "riskWarn": 70, "riskCrit": 90,
  "thermalWarn": 75, "thermalCrit": 85,
  "bufferWarn": 30, "bufferCrit": 10
}
```

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
| GET/PATCH | /v1/auth/me/settings/client | Client settings |
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

### Tables to Create

#### 1. organization_settings (NEW)
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

#### 2. device_settings (NEW)
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

### Phase 2: Settings Foundation ⏳ IN PROGRESS
- [x] Operator client settings (operator level)
- [x] Operator notification preferences (operator level)
- [ ] organization_settings table
- [ ] organization_settings CRUD endpoints
- [ ] Move thresholds to org level defaults
- [ ] device_settings table
- [ ] device_settings CRUD endpoints
- [ ] Threshold resolution: device → org → default

### Phase 3: Cleanup (Deferred)
- [ ] Remove operator.thresholds column
- [ ] Update all API paths to new structure

### Phase 4: Role Management (Deferred - Future Discussion)
- [ ] Member role update endpoints
- [ ] Admin can promote/demote members
- [ ] Super_admin protection

---

## 🔑 ROLE SYSTEM

### Roles (In Organization Context)
| Role | Level | Permissions |
|------|-------|-------------|
| super_admin | 4 | Full access, manage admins, delete org |
| admin | 3 | Manage members, devices, settings |
| operator | 2 | View devices, send commands |
| viewer | 1 | Read-only access |

**Role is NOT stored on Operator.** Determined by:
1. Active session's SelectedOrganizationID
2. Look up OrganizationMember record for (operator_id, organization_id)

---

## 📝 NOTES

### On Client Settings Being Operator-Level
These are device/client configuration preferences:
- serverUrl - Which server to connect to
- requestTimeoutMs - Client network behavior
- autoReconnect - Client persistence
- logBufferLimit - Client local storage

These are NOT user preferences - they are device properties.

### On Thresholds Being Per-Device
Different devices in the same org can have different operating characteristics:
- Temperature sensor in cold storage vs hot factory floor

The hierarchy (device → org → default) allows:
- Org admins set sensible defaults for new devices
- Individual devices can be tuned without affecting others

### Migration Notes (No Users in Production)
Since no users yet:
1. Drop old operator.thresholds column
2. Create new tables
3. Initialize org settings from defaults

---

## 📁 KEY FILES
| Component | Location |
|-----------|----------|
| Operator Entity | internal/domain/operator/operator_entity.go |
| Organization Entity | internal/domain/organization/organization_entity.go |
| Member Entity | internal/domain/organization/member_entity.go |
| Create Org Handler | internal/api/handlers/organization/organization_handler.go |
| Create Org Service | internal/application/organization/organization_service.go |
| Auth Login | internal/application/auth/auth_login_session.go |
| Auth Constructors | internal/application/auth/auth_constructors.go |
| Org Context Middleware | internal/api/middleware/org_context.go |
| Org Membership Middleware | internal/api/middleware/org_membership.go |
| Settings Handler | internal/api/handlers/auth/auth_settings.go |
| Settings Service | internal/application/auth/auth_settings_service.go |
| Threshold Types | internal/domain/operator/settings_types.go |
| Server Routes | internal/api/server_routes.go |
| Auth Routes | internal/api/handlers/auth/auth_routes.go |
| DTOs | internal/application/dto/auth.go |

---

## ✅ VERIFIED IMPLEMENTATIONS

### Create Organization
- ✅ Role validation: only "super_admin" or "admin" accepted
- ✅ Description required (enforced in handler)
- ✅ Name defaults to "personal" if empty
- ✅ MaxMembers defaults to 100 if not positive
- ✅ Operator added as member with specified role
- ✅ LastOrganizationID updated after creation

### Login Auto-Selection
- ✅ 0 memberships → needs_organization: true
- ✅ 1 membership → auto-selected
- ✅ 2+ with valid LastOrganizationID → auto-selected
- ✅ 2+ without LastOrganizationID → needs_organization: true

### Route Protection
- ✅ Dashboard routes require organization context
- ✅ Device routes require organization context
- ✅ Organization routes require organization + membership

### Session Tracking
- ✅ Session.SelectedOrganizationID field exists
- ✅ Updated on login (auto-resolve)
- ✅ Updated on organization switch
- ✅ Read by OrganizationContext middleware
