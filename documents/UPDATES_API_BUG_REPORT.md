# Updates API Bug Report

> **Document Version:** 1.0  
> **Status:** IN PROGRESS  
> **Created:** 2026-07-02  
> **Last Updated:** 2026-07-04  
> **Project:** Vyzorix Update Server

---

## Executive Summary

This document catalogs all bugs and missing implementations identified in the Updates API (SERVER_BACKEND_UPDATES_API.md). The Updates API is designed to support the Updates page with features for version management, push updates, and GitHub synchronization.

---

## Current State Analysis

### Existing Implementation Status

| Endpoint | Status | Handler | Notes |
|---------|--------|---------|-------|
| GET /api/v1/version | EXISTS | VersionHandler | Current APK version |
| GET /api/v1/apk/:filename | EXISTS | APKHandler | Serve APK file |
| POST /v1/device/:id/command (WAKE_UP_UPDATER) | EXISTS | CommandHandler | Wake device for update |

### Missing Endpoints (Required)

| Endpoint | Status | Priority |
|----------|--------|----------|
| GET /v1/updates/status | MISSING | HIGH |
| GET /v1/updates/versions | MISSING | HIGH |
| GET /v1/updates/changelog | MISSING | HIGH |
| POST /v1/updates/push | MISSING | HIGH |
| GET /v1/updates/history | MISSING | MEDIUM |
| GET /v1/updates/export | MISSING | MEDIUM |
| POST /v1/updates/sync | MISSING | HIGH |

---

## Missing Data Sources

| Data | Source | Status |
|------|--------|--------|
| Version metadata | GitHub bin/version.json | TO BE IMPLEMENTED |
| Changelog | GitHub bin/changelog.json | TO BE IMPLEMENTED |
| APK files | GitHub bin/v{version}/ | TO BE IMPLEMENTED |
| Update history | updates_history table | TO BE IMPLEMENTED |
| Sync status | updates_sync table | TO BE IMPLEMENTED |

---

## Required Database Schema

### updates_versions Table
```sql
CREATE TABLE updates_versions (
    id TEXT PRIMARY KEY,
    version TEXT UNIQUE NOT NULL,
    apk_filename TEXT NOT NULL,
    apk_size INTEGER NOT NULL,
    sha256 TEXT NOT NULL,
    release_date INTEGER NOT NULL,
    release_notes TEXT,
    release_type TEXT NOT NULL,
    is_latest INTEGER DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
```

### updates_pushes Table
```sql
CREATE TABLE updates_pushes (
    id TEXT PRIMARY KEY,
    version_id TEXT NOT NULL,
    install_type TEXT NOT NULL,
    scheduled_at INTEGER,
    status TEXT NOT NULL,
    initiated_by TEXT NOT NULL,
    initiated_at INTEGER NOT NULL,
    completed_at INTEGER,
    cancelled_at INTEGER,
    device_count INTEGER DEFAULT 0,
    FOREIGN KEY (version_id) REFERENCES updates_versions(id)
);
```

### updates_push_devices Table
```sql
CREATE TABLE updates_push_devices (
    id TEXT PRIMARY KEY,
    push_id TEXT NOT NULL,
    device_id TEXT NOT NULL,
    status TEXT NOT NULL,
    sent_at INTEGER,
    acknowledged_at INTEGER,
    error TEXT,
    FOREIGN KEY (push_id) REFERENCES updates_pushes(id)
);
```

### updates_sync_status Table
```sql
CREATE TABLE updates_sync_status (
    id INTEGER PRIMARY KEY,
    status TEXT NOT NULL,
    last_sync_at INTEGER,
    next_sync_at INTEGER,
    versions_found INTEGER DEFAULT 0,
    error TEXT,
    updated_at INTEGER NOT NULL
);
```

---

## HIGH Priority Bugs

### HIGH-UPD-1: Updates Status Endpoint Missing

**Severity:** HIGH  
**Status:** MISSING  
**Required File:** `internal/api/handlers/updates/updates_status_handler.go`

**Required Implementation:**
```go
type UpdatesStatusHandler struct {
    updatesSvc *updates.UpdatesService
}

func (h *UpdatesStatusHandler) GetStatus(c *gin.Context) {
    // Returns sync status, latest version, device status
}
```

---

### HIGH-UPD-2: Updates Versions Endpoint Missing

**Severity:** HIGH  
**Status:** MISSING  
**Required File:** `internal/api/handlers/updates/updates_versions_handler.go`

**Required Implementation:**
```go
func (h *UpdatesVersionsHandler) ListVersions(c *gin.Context) {
    // Query params: status, page, limit
    // Returns paginated list of versions
}
```

---

### HIGH-UPD-3: Updates Push Endpoint Missing

**Severity:** HIGH  
**Status:** MISSING  
**Required File:** `internal/api/handlers/updates/updates_push_handler.go`

**Required Implementation:**
```go
func (h *UpdatesPushHandler) PushUpdate(c *gin.Context) {
    // Request: { version, deviceIds, installType, scheduledAt }
    // Returns push status with device counts
    // Requires: operator:write permission
}
```

---

### HIGH-UPD-4: GitHub Sync Endpoint Missing

**Severity:** HIGH  
**Status:** MISSING  
**Required File:** `internal/api/handlers/updates/updates_sync_handler.go`

**Required Implementation:**
```go
func (h *UpdatesSyncHandler) SyncFromGitHub(c *gin.Context) {
    // Syncs versions from GitHub bin/ directory
    // Updates version metadata and changelog
    // Rate limited: 5/hour
}
```

---

## MEDIUM Priority Bugs

### MEDIUM-UPD-1: Updates History Endpoint Missing

**Severity:** MEDIUM  
**Status:** MISSING  
**Required File:** `internal/api/handlers/updates/updates_history_handler.go`

---

### MEDIUM-UPD-2: Updates Export Endpoint Missing

**Severity:** MEDIUM  
**Status:** MISSING  
**Required File:** `internal/api/handlers/updates/updates_export_handler.go`

---

### MEDIUM-UPD-3: Version Changelog Endpoint Missing

**Severity:** MEDIUM  
**Status:** MISSING  
**Required File:** `internal/api/handlers/updates/updates_changelog_handler.go`

---

## Domain Layer Requirements

### Required Files to Create

| File | Purpose |
|------|---------|
| `domain/updates/updates_entity.go` | Update entities (Version, Push, PushDevice, SyncStatus) |
| `domain/updates/updates_repository.go` | Repository interface |
| `domain/updates/updates_errors.go` | Domain errors |

### Entity Definitions

```go
type UpdateVersion struct {
    ID           string
    Version      string
    ApkFilename  string
    ApkSize      int64
    SHA256       string
    ReleaseDate  time.Time
    ReleaseNotes string
    ReleaseType  ReleaseType
    IsLatest     bool
}

type UpdatePush struct {
    ID           string
    VersionID    string
    InstallType  InstallType
    ScheduledAt  *time.Time
    Status       PushStatus
    InitiatedBy  string
    InitiatedAt  time.Time
    CompletedAt  *time.Time
    CancelledAt  *time.Time
    DeviceCount  int
    Devices      []*PushDevice
}
```

---

## Application Layer Requirements

### Required Files to Create

| File | Purpose |
|------|---------|
| `application/updates/updates_service.go` | Main service orchestrator |
| `application/updates/updates_versions_service.go` | Version management |
| `application/updates/updates_push_service.go` | Push management |
| `application/updates/updates_history_service.go` | History management |
| `application/updates/updates_sync_service.go` | GitHub sync logic |
| `application/updates/updates_dto.go` | DTOs |

---

## Infrastructure Requirements

### Required Files to Create

| File | Purpose |
|------|---------|
| `infrastructure/storage/updates_storage.go` | Database queries |
| `infrastructure/github/github_client.go` | GitHub API client |
| `infrastructure/github/github_sync.go` | GitHub sync logic |
| `infrastructure/migrations/` | SQL migrations |

---

## Implementation Status

| Component | Status | Notes |
|-----------|--------|-------|
| Database Schema | NOT STARTED | Tables need to be created |
| Domain Layer | NOT STARTED | Entities and repository interface |
| Application Layer | NOT STARTED | Services need implementation |
| Handler Layer | NOT STARTED | HTTP handlers needed |
| Infrastructure | NOT STARTED | GitHub client and storage |
| GraphQL Schema | NOT STARTED | Resolvers and types |

---

## Bug Report Summary

| Bug ID | Severity | Description | Status |
|--------|----------|-------------|--------|
| HIGH-UPD-1 | HIGH | Updates status endpoint missing | MISSING |
| HIGH-UPD-2 | HIGH | Updates versions endpoint missing | MISSING |
| HIGH-UPD-3 | HIGH | Updates push endpoint missing | MISSING |
| HIGH-UPD-4 | HIGH | GitHub sync endpoint missing | MISSING |
| MEDIUM-UPD-1 | MEDIUM | Updates history endpoint missing | MISSING |
| MEDIUM-UPD-2 | MEDIUM | Updates export endpoint missing | MISSING |
| MEDIUM-UPD-3 | MEDIUM | Version changelog endpoint missing | MISSING |

---

## Next Steps

1. Create database migrations for updates_* tables
2. Implement domain layer (entities and repository interface)
3. Implement infrastructure layer (storage and GitHub client)
4. Implement application layer (services)
5. Implement handler layer (HTTP handlers)
6. Add GraphQL schema and resolvers
7. Integration testing

---

*Document Version: 1.0*  
*Status: In Progress - Implementation Required*  
*Vyzorix-update-server*
