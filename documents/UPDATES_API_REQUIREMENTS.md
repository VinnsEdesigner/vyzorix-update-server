# Updates API Requirements Analysis

> **Document Status:** Analysis In Progress
> **Last Updated:** 2026-07-04
> **Project:** vyzorix-update-server

---

## Executive Summary

This document tracks the requirements for the Updates API implementation and serves as a reference for the frontend client expectations.

---

## 1. Core Functionality Requirements

### 1.1 Data Flow

```
GitHub Repository (bin/) ──────────────────────┐
                                            │
                                            ▼
                                    Update Server
                                            │
    ┌────────────────────────────────────────┤
    │                                        │
    ▼                                        ▼
Versions/Metadata ◄────────────────── GitHub Sync (periodic/webhook)
    │
    │ REST/GraphQL
    ▼
Frontend (Updates Page)
    │
    ▼
Push Update ──────► FCM/WSS ──────► Device
```

### 1.2 Required Features

| Feature | Endpoint | Status | Notes |
|---------|----------|--------|-------|
| Status | GET /v1/updates/status | IMPLEMENTED | Returns sync status, latest version |
| Versions | GET /v1/updates/versions | IMPLEMENTED | Paginated version list |
| Changelog | GET /v1/updates/changelog | IMPLEMENTED | Release notes by version |
| Push | POST /v1/updates/push | IMPLEMENTED | Push update to devices |
| History | GET /v1/updates/history | IMPLEMENTED | Update push history |
| Export | GET /v1/updates/export | IMPLEMENTED | Export version data |
| Sync | POST /v1/updates/sync | IMPLEMENTED | Sync versions from GitHub |

---

## 2. API Response Specifications

### 2.1 GET /v1/updates/status

**Expected Response Structure:**
```json
{
  "sync": {
    "status": "synced",
    "lastSyncAt": 1718900000000,
    "nextSyncAt": 1718986400000,
    "error": null
  },
  "latest": {
    "version": "2.2.1",
    "apkFilename": "VyzorixAudioRouter-v2.2.1.apk",
    "apkSize": 15728640,
    "sha256": "abc123...",
    "releasedAt": 1718800000000,
    "releaseNotes": "Bug fixes..."
  },
  "device": {
    "currentVersion": "2.2.0",
    "needsUpdate": true
  }
}
```

**Sync Status Values:**
| Status | Description |
|--------|-------------|
| idle | No sync in progress |
| syncing | Currently syncing from GitHub |
| synced | Last sync successful |
| error | Last sync failed |

### 2.2 GET /v1/updates/versions

**Expected Response Structure:**
```json
{
  "versions": [
    {
      "version": "2.2.1",
      "apkFilename": "VyzorixAudioRouter-v2.2.1.apk",
      "apkSize": 15728640,
      "sha256": "abc123def456...",
      "releasedAt": 1718800000000,
      "releaseNotes": "Bug fixes and performance improvements",
      "releaseType": "patch",
      "status": "latest"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 3,
    "totalPages": 1
  }
}
```

**Release Type Values:**
| Type | Description |
|------|-------------|
| major | Breaking changes |
| minor | New features |
| patch | Bug fixes |

### 2.3 GET /v1/updates/changelog

**Expected Response Structure:**
```json
{
  "changelog": [
    {
      "version": "2.2.1",
      "date": "2024-06-19",
      "type": "patch",
      "notes": "Bug fixes and performance improvements\n\n- Fixed audio latency issue\n- Improved stability"
    }
  ]
}
```

### 2.4 POST /v1/updates/push

**Request:**
```json
{
  "version": "2.2.1",
  "deviceIds": ["861234567890123", "861234567890124"],
  "installType": "immediate",
  "scheduledAt": null
}
```

**Install Type Values:**
| Type | Description |
|------|-------------|
| immediate | Push now, device installs on next check-in |
| scheduled | Push at specified time |

**Expected Response:**
```json
{
  "pushId": "push_abc123",
  "version": "2.2.1",
  "deviceIds": ["861234567890123", "861234567890124"],
  "installType": "immediate",
  "status": "in_progress",
  "initiatedBy": "operator_123",
  "initiatedAt": 1718900000000,
  "devices": {
    "total": 2,
    "pending": 2,
    "sent": 0,
    "acknowledged": 0,
    "failed": 0
  }
}
```

### 2.5 POST /v1/updates/sync

**Expected Response:**
```json
{
  "status": "synced",
  "startedAt": 1718900000000,
  "versionsFound": 5,
  "message": "Successfully synced 5 versions"
}
```

---

## 3. GraphQL Schema Requirements

### 3.1 Types

```graphql
type UpdateVersion {
  id: ID!
  version: String!
  apkFilename: String!
  apkSize: BigInt!
  sha256: String!
  releaseDate: DateTime!
  releaseNotes: String
  releaseType: ReleaseType!
  isLatest: Boolean!
}

enum ReleaseType {
  MAJOR
  MINOR
  PATCH
}

enum UpdateStatus {
  PENDING
  IN_PROGRESS
  COMPLETED
  FAILED
  CANCELLED
}

enum InstallType {
  IMMEDIATE
  SCHEDULED
}

enum DevicePushStatus {
  PENDING
  SENT
  ACKNOWLEDGED
  FAILED
}

type UpdatePush {
  id: ID!
  version: UpdateVersion!
  installType: InstallType!
  scheduledAt: DateTime
  status: UpdateStatus!
  initiatedBy: String!
  initiatedAt: DateTime!
  completedAt: DateTime
  cancelledAt: DateTime
  deviceCount: Int!
  devices: [PushDevice!]!
}

type PushDevice {
  id: ID!
  deviceId: String!
  deviceName: String
  status: DevicePushStatus!
  sentAt: DateTime
  acknowledgedAt: DateTime
  error: String
}

type SyncStatus {
  status: String!
  lastSyncAt: DateTime
  nextSyncAt: DateTime
  versionsFound: Int
  error: String
}

type ChangelogEntry {
  version: String!
  date: String!
  type: ReleaseType!
  notes: String!
}
```

### 3.2 Queries

```graphql
type Query {
  updatesStatus: UpdateStatusResult!
  updatesVersions(status: String, page: Int, limit: Int): [UpdateVersion!]!
  updatesChangelog(version: String): [ChangelogEntry!]!
  updatesHistory(status: String, page: Int, limit: Int): HistoryResult!
  updatesHistoryDetail(pushId: ID!): UpdatePush
  updatesSyncStatus: SyncStatus!
}

type HistoryResult {
  pushes: [UpdatePush!]!
  pagination: Pagination!
}

type Pagination {
  total: Int!
  limit: Int!
  offset: Int!
  hasMore: Boolean!
}
```

### 3.3 Mutations

```graphql
type Mutation {
  pushUpdate(input: PushUpdateInput!): UpdatePush!
  cancelUpdate(pushId: ID!): UpdatePush!
  syncFromGitHub: SyncResult!
}

input PushUpdateInput {
  version: String!
  deviceIds: [String!]!
  installType: InstallType!
  scheduledAt: DateTime
}

type SyncResult {
  status: String!
  startedAt: DateTime!
  versionsFound: Int
  message: String
}
```

---

## 4. Known Issues / Bugs

### 4.1 Updates Resolver Issues

**Issue 1:** `GetUpdatesStatus` response structure mismatch
- **Location:** `internal/api/graphql/resolver/updates_resolver.go` lines 42-46
- **Problem:** `device` field is mapped to a string (`CurrentVersion`) instead of an object with `currentVersion` and `needsUpdate`
- **Expected:** 
  ```json
  "device": {
    "currentVersion": "2.2.0",
    "needsUpdate": true
  }
  ```
- **Actual:** `"device": "2.2.0"` (string only)

**Issue 2:** Hardcoded `releaseType`
- **Location:** `internal/api/graphql/resolver/updates_resolver.go` line 52, 100
- **Problem:** `releaseType` is hardcoded as `"patch"` instead of using actual `v.ReleaseType`

### 4.2 Frontend Client Data Expectations

The frontend expects:
1. `latest.version` - string
2. `latest.apkFilename` - string  
3. `latest.apkSize` - number
4. `latest.sha256` - string
5. `latest.releaseNotes` - string
6. `latest.releaseType` - enum (MAJOR/MINOR/PATCH)
7. `sync.status` - string
8. `sync.lastSyncAt` - timestamp
9. `sync.nextSyncAt` - timestamp
10. `device.currentVersion` - string
11. `device.needsUpdate` - boolean

---

## 5. Implementation Checklist

### 5.1 Storage Layer
- [x] `CreateVersion`
- [x] `GetVersionByID`
- [x] `GetVersionByVersion`
- [x] `GetLatestVersion`
- [x] `ListVersions`
- [x] `UpdateLatestFlag`
- [x] `CreatePush`
- [x] `GetPushByID`
- [x] `UpdatePushStatus`
- [x] `CompletePush`
- [x] `CancelPush`
- [x] `ListPushes`
- [x] `CreatePushDevice`
- [x] `GetPushDevices`
- [x] `UpdatePushDeviceStatus`
- [x] `GetSyncState`
- [x] `UpdateSyncState`
- [x] `TryAcquireSyncLock`

### 5.2 Service Layer
- [x] `GetStatus` / `GetUpdateStatus`
- [x] `GetVersions`
- [x] `GetChangelog`
- [x] `PushUpdate`
- [x] `GetHistory`
- [x] `GetPushDetail`
- [x] `CancelPush`
- [x] `SyncFromGitHub`
- [x] `ExportVersions`

### 5.3 GraphQL Resolvers
- [x] `GetUpdatesStatus`
- [x] `GetUpdatesVersions`
- [x] `GetUpdatesChangelog`
- [x] `GetUpdatesHistory`
- [x] `GetUpdatesHistoryDetail`
- [x] `GetUpdatesSyncStatus`
- [x] `PushUpdate`
- [x] `CancelUpdate`
- [x] `SyncFromGitHub`

### 5.4 HTTP Handlers
- [ ] Verify all HTTP handler endpoints match spec

---

## 6. Pending Tasks

1. **Fix GraphQL Resolver bugs** - Fix `GetUpdatesStatus` to return proper `device` object
2. **Fix hardcoded releaseType** - Use actual `v.ReleaseType` value
3. **Test sync from GitHub** - Verify binary download works
4. **Verify push notification delivery** - FCM integration test
5. **Frontend integration test** - Verify client receives expected data

---

## 7. Test Scenarios

### 7.1 Version Sync Test
1. Trigger `syncFromGitHub` mutation
2. Verify versions are fetched from GitHub
3. Verify `bin/version.json` and `bin/changelog.json` are parsed
4. Verify `bin/v{version}/` APKs are downloaded
5. Verify database is populated

### 7.2 Push Update Test
1. Select version and devices
2. Choose install type (immediate/scheduled)
3. Trigger `pushUpdate` mutation
4. Verify push record created
5. Verify devices receive FCM notification
6. Verify device acknowledgment updates push status

### 7.3 Version Query Test
1. Query `updatesVersions`
2. Verify pagination works
3. Verify status filter works (all/latest/previous)
4. Verify response matches expected structure

---

## 8. Notes

- FCM push notifications should be implemented (Android APK has FCM infrastructure)
- GitHub webhook for automatic sync on new releases
- Version validation (SHA256) before serving APKs
- Rate limiting as specified in spec

---

*Document Version: 1.0*
*Status: Requirements Analysis Complete*
