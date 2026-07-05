# Server Backend - Updates API Specification

> **Version:** 1.0
> **Status:** Draft
> **Created:** 2026-06-24
> **Target:** Production MVP
> **Go Package:** `github.com/VinnsEdesigner/vyzorix/apps/api`

---

## Table of Contents

1. Overview
2. Current State Analysis
3. Required API Endpoints
4. Database Schema
5. Backend File Structure
6. Handler Specifications
7. Service Layer
8. GraphQL Schema
9. Error Handling
10. Rate Limiting & Security
11. File Changes Summary
12. Implementation Order

---

## 1. Overview

### 1.1 Purpose

This document maps out the server-side requirements to support the Updates page as specified in UPDATES_PAGE.md.

### 1.2 Frontend Requirements Summary

| Feature | Description | Required Endpoints |
|---------|-------------|-------------------|
| Status | Current sync status, latest version | GET /v1/updates/status |
| Versions | All available APK versions | GET /v1/updates/versions |
| Changelog | Release notes by version | GET /v1/updates/changelog |
| Push | Push update to devices | POST /v1/updates/push |
| History | Update push history | GET /v1/updates/history |
| Export | Export version data | GET /v1/updates/export |
| Sync | Sync versions from GitHub | POST /v1/updates/sync |

### 1.3 Data Flow

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

---

## 2. Current State Analysis

### 2.1 Existing Related Endpoints

| Endpoint | Status | Handler | Notes |
|----------|--------|---------|-------|
| GET /api/v1/version | EXISTS | VersionHandler | Current APK version |
| GET /api/v1/apk/:filename | EXISTS | APKHandler | Serve APK file |
| POST /v1/device/:imei/command (WAKE_UP_UPDATER) | EXISTS | CommandHandler | Wake device for update |

### 2.2 Missing Endpoints

| Endpoint | Status |
|----------|--------|
| GET /v1/updates/status | MISSING |
| GET /v1/updates/versions | MISSING |
| GET /v1/updates/changelog | MISSING |
| POST /v1/updates/push | MISSING |
| GET /v1/updates/history | MISSING |
| GET /v1/updates/export | MISSING |
| POST /v1/updates/sync | MISSING |

### 2.3 Data Sources

| Data | Source | Status |
|------|--------|--------|
| Version metadata | GitHub bin/version.json | TO BE IMPLEMENTED |
| Changelog | GitHub bin/changelog.json | TO BE IMPLEMENTED |
| APK files | GitHub bin/v{version}/ | TO BE IMPLEMENTED |
| Update history | updates_history table | TO BE IMPLEMENTED |
| Sync status | updates_sync table | TO BE IMPLEMENTED |

---

## 3. Required API Endpoints

### 3.1 GET /v1/updates/status

Get current update system status.

**Response (200 OK):**
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
    "releasedAt": 1718800000000
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

---

### 3.2 GET /v1/updates/versions

Get all available APK versions.

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| status | string | all | Filter: all, latest, previous |
| page | int | 1 | Page number |
| limit | int | 20 | Items per page (max 50) |

**Response (200 OK):**
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
      "status": "latest"
    },
    {
      "version": "2.2.0",
      "apkFilename": "VyzorixAudioRouter-v2.2.0.apk",
      "apkSize": 15204352,
      "sha256": "def789...",
      "releasedAt": 1718200000000,
      "releaseNotes": "New features added",
      "status": "previous"
    },
    {
      "version": "2.1.0",
      "apkFilename": "VyzorixAudioRouter-v2.1.0.apk",
      "apkSize": 14811136,
      "sha256": "ghi012...",
      "releasedAt": 1717600000000,
      "releaseNotes": "Initial release",
      "status": "previous"
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

---

### 3.3 GET /v1/updates/changelog

Get release changelog.

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| version | string | latest | Specific version or "all" |

**Response (200 OK):**
```json
{
  "changelog": [
    {
      "version": "2.2.1",
      "date": "2024-06-19",
      "type": "patch",
      "notes": "Bug fixes and performance improvements\n\n- Fixed audio latency issue\n- Improved stability\n- Updated dependencies"
    },
    {
      "version": "2.2.0",
      "date": "2024-06-05",
      "type": "minor",
      "notes": "New features added\n\n- Added support for new device models\n- Enhanced diagnostic tools\n- Improved logging"
    },
    {
      "version": "2.1.0",
      "date": "2024-05-15",
      "type": "major",
      "notes": "Initial release\n\n- Core functionality\n- Basic device management\n- Telemetry collection"
    }
  ]
}
```

**Release Type Values:**
| Type | Description |
|------|-------------|
| major | Breaking changes |
| minor | New features |
| patch | Bug fixes |

---

### 3.4 POST /v1/updates/push

Push update to devices.

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

**Response (202 Accepted):**
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

---

### 3.5 GET /v1/updates/history

Get update push history.

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| status | string | all | Filter: all, pending, in_progress, completed, failed, cancelled |
| page | int | 1 | Page number |
| limit | int | 20 | Items per page (max 100) |

**Response (200 OK):**
```json
{
  "pushes": [
    {
      "id": "push_abc123",
      "version": "2.2.1",
      "installType": "immediate",
      "status": "completed",
      "initiatedBy": "operator@example.com",
      "initiatedAt": 1718900000000,
      "completedAt": 1718900600000,
      "deviceCount": 2,
      "devices": {
        "acknowledged": 2,
        "failed": 0
      }
    },
    {
      "id": "push_xyz789",
      "version": "2.2.0",
      "installType": "scheduled",
      "scheduledAt": 1718901000000,
      "status": "cancelled",
      "initiatedBy": "operator@example.com",
      "initiatedAt": 1718899000000,
      "cancelledAt": 1718899500000,
      "deviceCount": 5,
      "devices": {
        "pending": 5,
        "acknowledged": 0,
        "failed": 0
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 2,
    "totalPages": 1
  }
}
```

---

### 3.6 GET /v1/updates/history/:pushId

Get detailed status of a specific update push.

**Response (200 OK):**
```json
{
  "id": "push_abc123",
  "version": "2.2.1",
  "installType": "immediate",
  "status": "completed",
  "initiatedBy": "operator@example.com",
  "initiatedAt": 1718900000000,
  "completedAt": 1718900600000,
  "devices": [
    {
      "deviceId": "861234567890123",
      "deviceName": "Pixel 8 Pro",
      "status": "acknowledged",
      "acknowledgedAt": 1718900300000,
      "error": null
    },
    {
      "deviceId": "861234567890124",
      "deviceName": "Pixel 7",
      "status": "acknowledged",
      "acknowledgedAt": 1718900600000,
      "error": null
    }
  ]
}
```

---

### 3.7 POST /v1/updates/history/:pushId/cancel

Cancel a pending update push.

**Response (200 OK):**
```json
{
  "id": "push_xyz789",
  "status": "cancelled",
  "cancelledAt": 1718900500000,
  "cancelledBy": "operator@example.com"
}
```

**Errors:**
- 400 - Push cannot be cancelled (already completed/failed)
- 404 - Push not found

---

### 3.8 GET /v1/updates/export

Export version data.

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| format | string | json | Export format: json, csv |
| version | string | all | Specific version or "all" |
| includeChangelog | bool | true | Include changelog |
| includeApkInfo | bool | true | Include APK metadata |

**Response (200 OK) - JSON:**
```json
{
  "exportedAt": 1718901600000,
  "format": "json",
  "versions": [
    {
      "version": "2.2.1",
      "apkFilename": "VyzorixAudioRouter-v2.2.1.apk",
      "apkSize": 15728640,
      "sha256": "abc123...",
      "releasedAt": 1718800000000,
      "releaseNotes": "Bug fixes and performance improvements"
    }
  ],
  "changelog": [
    {
      "version": "2.2.1",
      "date": "2024-06-19",
      "type": "patch",
      "notes": "Bug fixes and performance improvements"
    }
  ]
}
```

**Response (200 OK) - CSV:**
```
version,apkFilename,apkSize,sha256,releasedAt,releaseType,releaseNotes
2.2.1,VyzorixAudioRouter-v2.2.1.apk,15728640,abc123...,2024-06-19,patch,Bug fixes
2.2.0,VyzorixAudioRouter-v2.2.0.apk,15204352,def789...,2024-06-05,minor,New features
```

---

### 3.9 POST /v1/updates/sync

Trigger manual sync from GitHub.

**Response (202 Accepted):**
```json
{
  "status": "syncing",
  "startedAt": 1718901600000,
  "message": "Syncing versions from GitHub..."
}
```

**Response (200 OK) - If already syncing:**
```json
{
  "status": "syncing",
  "startedAt": 1718901500000,
  "message": "Sync already in progress"
}
```

---

### 3.10 GET /v1/updates/sync/status

Get sync status.

**Response (200 OK):**
```json
{
  "status": "synced",
  "lastSyncAt": 1718900000000,
  "nextSyncAt": 1718986400000,
  "versionsFound": 3,
  "error": null
}
```

---

## 4. Database Schema

### 4.1 Updates Versions Table (NEW)

```sql
CREATE TABLE update_versions (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    version TEXT NOT NULL UNIQUE,
    apk_filename TEXT NOT NULL,
    apk_size BIGINT NOT NULL,
    sha256 TEXT NOT NULL,
    release_date TIMESTAMPTZ NOT NULL,
    release_notes TEXT,
    release_type TEXT NOT NULL DEFAULT 'patch',
    is_latest BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT idx_version UNIQUE (version),
    CONSTRAINT idx_latest UNIQUE (is_latest) WHERE is_latest = TRUE
);

CREATE INDEX idx_versions_date ON update_versions(release_date DESC);
```

### 4.2 Update Pushes Table (NEW)

```sql
CREATE TABLE update_pushes (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    version_id TEXT NOT NULL REFERENCES update_versions(id),
    install_type TEXT NOT NULL DEFAULT 'immediate',
    scheduled_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'pending',
    initiated_by TEXT NOT NULL,
    initiated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    cancelled_by TEXT,
    
    CONSTRAINT idx_status (status),
    CONSTRAINT idx_initiated_at (initiated_at DESC)
);
```

### 4.3 Update Push Devices Table (NEW)

```sql
CREATE TABLE update_push_devices (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    push_id TEXT NOT NULL REFERENCES update_pushes(id) ON DELETE CASCADE,
    device_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    sent_at TIMESTAMPTZ,
    acknowledged_at TIMESTAMPTZ,
    error TEXT,
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_push FOREIGN KEY (push_id) REFERENCES update_pushes(id) ON DELETE CASCADE,
    CONSTRAINT idx_push_device UNIQUE (push_id, device_id),
    CONSTRAINT idx_status (status)
);

CREATE INDEX idx_update_devices_push ON update_push_devices(push_id, status);
```

### 4.4 Sync Status Table (NEW)

```sql
CREATE TABLE update_sync_status (
    id TEXT PRIMARY KEY DEFAULT 'singleton',
    status TEXT NOT NULL DEFAULT 'idle',
    last_sync_at TIMESTAMPTZ,
    last_sync_error TEXT,
    next_sync_at TIMESTAMPTZ,
    versions_found INT DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Initialize singleton row
INSERT INTO update_sync_status (id, status) VALUES ('singleton', 'idle');
```

---

## 5. Backend File Structure

```
apps/api/internal/
├── api/
│   ├── handlers/
│   │   ├── updates/
│   │   │   ├── updates_versions_handler.go     # NEW - versions handlers
│   │   │   ├── updates_push_handler.go         # NEW - push handlers
│   │   │   ├── updates_history_handler.go     # NEW - history handlers
│   │   │   ├── updates_sync_handler.go         # NEW - sync handlers
│   │   │   └── updates_routes.go              # NEW - updates routes
│   │   └── router.go                          # MODIFIED - add updates
│   └── middleware/
│       └── ...
├── application/
│   └── updates/
│       ├── updates_service.go                  # NEW - main updates service
│       ├── updates_versions_service.go         # NEW - version management
│       ├── updates_push_service.go             # NEW - push management
│       ├── updates_history_service.go          # NEW - history management
│       ├── updates_sync_service.go             # NEW - GitHub sync
│       └── updates_dto.go                     # NEW - request/response DTOs
├── domain/
│   └── updates/
│       ├── updates_entity.go                  # NEW - update entities
│       ├── updates_repository.go              # NEW - repository interface
│       └── updates_errors.go                  # NEW - domain errors
├── infrastructure/
│   ├── storage/
│   │   ├── updates_storage.go                 # NEW - updates queries
│   │   └── migrations/                       # NEW - SQL migrations
│   └── github/
│       ├── github_client.go                   # NEW - GitHub API client
│       └── github_sync.go                    # NEW - GitHub sync logic
```

---

## 6. Handler Specifications

### 6.1 Versions Handler

**File:** `api/handlers/updates/updates_versions_handler.go`

```go
package updates

import (
    "net/http"
    "strconv"

    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
    "github.com/gin-gonic/gin"
)

type VersionsHandler struct {
    service *updates.Service
}

func NewVersionsHandler(service *updates.Service) *VersionsHandler {
    return &VersionsHandler{service: service}
}

// GetStatus handles GET /v1/updates/status
func (h *VersionsHandler) GetStatus(c *gin.Context) {
    status, err := h.service.GetStatus(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get status"})
        return
    }
    c.JSON(http.StatusOK, status)
}

// GetVersions handles GET /v1/updates/versions
func (h *VersionsHandler) GetVersions(c *gin.Context) {
    status := c.Query("status")
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 50 {
        limit = 20
    }

    result, err := h.service.GetVersions(c.Request.Context(), status, page, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get versions"})
        return
    }
    c.JSON(http.StatusOK, result)
}

// GetChangelog handles GET /v1/updates/changelog
func (h *VersionsHandler) GetChangelog(c *gin.Context) {
    version := c.Query("version")
    if version == "" {
        version = "all"
    }

    changelog, err := h.service.GetChangelog(c.Request.Context(), version)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get changelog"})
        return
    }
    c.JSON(http.StatusOK, changelog)
}
```

---

### 6.2 Push Handler

**File:** `api/handlers/updates/push.go`

```go
package updates

import (
    "net/http"

    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
    "github.com/gin-gonic/gin"
)

type PushHandler struct {
    service *updates.Service
}

func NewPushHandler(service *updates.Service) *PushHandler {
    return &PushHandler{service: service}
}

// PushUpdate handles POST /v1/updates/push
func (h *PushHandler) PushUpdate(c *gin.Context) {
    var req struct {
        Version     string   `json:"version" binding:"required"`
        DeviceIDs   []string `json:"deviceIds" binding:"required,min=1"`
        InstallType string   `json:"installType" binding:"required,oneof=immediate scheduled"`
        ScheduledAt *int64  `json:"scheduledAt"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
        return
    }

    // Get operator from context
    operatorID := "operator_123" // TODO: Get from auth context

    push, err := h.service.PushUpdate(c.Request.Context(), &updates.PushRequest{
        Version:     req.Version,
        DeviceIDs:   req.DeviceIDs,
        InstallType: req.InstallType,
        ScheduledAt: req.ScheduledAt,
        InitiatedBy: operatorID,
    })
    if err != nil {
        if err == updates.ErrVersionNotFound {
            c.JSON(http.StatusBadRequest, gin.H{"error": "version_not_found", "message": "version not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to push update"})
        return
    }

    c.JSON(http.StatusAccepted, push)
}
```

---

### 6.3 History Handler

**File:** `api/handlers/updates/history.go`

```go
package updates

import (
    "net/http"
    "strconv"

    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
    "github.com/gin-gonic/gin"
)

type HistoryHandler struct {
    service *updates.Service
}

func NewHistoryHandler(service *updates.Service) *HistoryHandler {
    return &HistoryHandler{service: service}
}

// GetHistory handles GET /v1/updates/history
func (h *HistoryHandler) GetHistory(c *gin.Context) {
    status := c.Query("status")
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 20
    }

    result, err := h.service.GetHistory(c.Request.Context(), status, page, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get history"})
        return
    }
    c.JSON(http.StatusOK, result)
}

// GetPushDetail handles GET /v1/updates/history/:pushId
func (h *HistoryHandler) GetPushDetail(c *gin.Context) {
    pushID := c.Param("pushId")

    detail, err := h.service.GetPushDetail(c.Request.Context(), pushID)
    if err != nil {
        if err == updates.ErrPushNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "push not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get push detail"})
        return
    }
    c.JSON(http.StatusOK, detail)
}

// CancelPush handles POST /v1/updates/history/:pushId/cancel
func (h *HistoryHandler) CancelPush(c *gin.Context) {
    pushID := c.Param("pushId")

    result, err := h.service.CancelPush(c.Request.Context(), pushID)
    if err != nil {
        if err == updates.ErrPushNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "push not found"})
            return
        }
        if err == updates.ErrPushNotCancellable {
            c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "push cannot be cancelled"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel push"})
        return
    }
    c.JSON(http.StatusOK, result)
}
```

---

### 6.4 Sync Handler

**File:** `api/handlers/updates/sync.go`

```go
package updates

import (
    "net/http"

    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
    "github.com/gin-gonic/gin"
)

type SyncHandler struct {
    service *updates.Service
}

func NewSyncHandler(service *updates.Service) *SyncHandler {
    return &SyncHandler{service: service}
}

// Sync handles POST /v1/updates/sync
func (h *SyncHandler) Sync(c *gin.Context) {
    result, err := h.service.SyncFromGitHub(c.Request.Context())
    if err != nil {
        if err == updates.ErrSyncAlreadyInProgress {
            c.JSON(http.StatusOK, result)
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed", "message": err.Error()})
        return
    }
    c.JSON(http.StatusAccepted, result)
}

// GetSyncStatus handles GET /v1/updates/sync/status
func (h *SyncHandler) GetSyncStatus(c *gin.Context) {
    status, err := h.service.GetSyncStatus(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sync status"})
        return
    }
    c.JSON(http.StatusOK, status)
}
```

---

## 7. Service Layer

### 7.1 Updates Service

**File:** `application/updates/service.go`

```go
package updates

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "errors"
    "fmt"
    "io"
    "net/http"
    "path"
    "time"

    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/updates"
    "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/github"
)

var (
    ErrVersionNotFound       = errors.New("version not found")
    ErrPushNotFound         = errors.New("push not found")
    ErrPushNotCancellable   = errors.New("push cannot be cancelled")
    ErrSyncAlreadyInProgress = errors.New("sync already in progress")
)

type Service struct {
    repo      *updates.Repository
    githubClient *github.Client
}

func NewService(repo *updates.Repository, githubClient *github.Client) *Service {
    return &Service{repo: repo, githubClient: githubClient}
}

// GetStatus returns current update system status
func (s *Service) GetStatus(ctx context.Context) (*StatusResult, error) {
    syncStatus, err := s.repo.GetSyncStatus(ctx)
    if err != nil {
        return nil, err
    }

    latest, err := s.repo.GetLatestVersion(ctx)
    if err != nil {
        return nil, err
    }

    return &StatusResult{
        Sync: SyncStatusInfo{
            Status:     syncStatus.Status,
            LastSyncAt: syncStatus.LastSyncAt,
            NextSyncAt: syncStatus.NextSyncAt,
            Error:      syncStatus.LastSyncError,
        },
        Latest: latest,
    }, nil
}

// GetVersions returns paginated versions
func (s *Service) GetVersions(ctx context.Context, status string, page, limit int) (*VersionsResult, error) {
    offset := (page - 1) * limit
    
    versions, total, err := s.repo.ListVersions(ctx, status, limit, offset)
    if err != nil {
        return nil, err
    }

    totalPages := (total + limit - 1) / limit

    return &VersionsResult{
        Versions: versions,
        Pagination: Pagination{
            Page:       page,
            Limit:      limit,
            Total:      total,
            TotalPages: totalPages,
        },
    }, nil
}

// PushUpdate creates a new update push
func (s *Service) PushUpdate(ctx context.Context, req *PushRequest) (*PushResult, error) {
    // Verify version exists
    version, err := s.repo.GetVersionByVersion(ctx, req.Version)
    if err != nil {
        if err == updates.ErrNotFound {
            return nil, ErrVersionNotFound
        }
        return nil, err
    }

    // Create push record
    push := &updates.Push{
        VersionID:    version.ID,
        InstallType: req.InstallType,
        ScheduledAt:  req.ScheduledAt,
        Status:      updates.StatusPending,
        InitiatedBy: req.InitiatedBy,
        InitiatedAt: time.Now(),
    }

    if err := s.repo.CreatePush(ctx, push); err != nil {
        return nil, err
    }

    // Create device push records and trigger FCM/WSS
    for _, deviceID := range req.DeviceIDs {
        devicePush := &updates.PushDevice{
            PushID:    push.ID,
            DeviceID:  deviceID,
            Status:    updates.DevicePushStatusPending,
            SentAt:    time.Now(),
        }
        if err := s.repo.CreatePushDevice(ctx, devicePush); err != nil {
            // Log error but continue
            continue
        }

        // Send command to device to check for updates
        // This would integrate with the existing command system
    }

    return &PushResult{
        ID:          push.ID,
        Version:     req.Version,
        DeviceIDs:   req.DeviceIDs,
        InstallType: req.InstallType,
        Status:      push.Status,
        InitiatedBy: req.InitiatedBy,
        InitiatedAt: push.InitiatedAt.UnixMilli(),
        DeviceCount: len(req.DeviceIDs),
    }, nil
}

// SyncFromGitHub syncs versions from GitHub
func (s *Service) SyncFromGitHub(ctx context.Context) (*SyncResult, error) {
    // Check if sync already in progress
    syncStatus, err := s.repo.GetSyncStatus(ctx)
    if err != nil {
        return nil, err
    }

    if syncStatus.Status == "syncing" {
        return &SyncResult{
            Status:    "syncing",
            StartedAt: syncStatus.LastSyncAt.UnixMilli(),
            Message:   "Sync already in progress",
        }, ErrSyncAlreadyInProgress
    }

    // Update status to syncing
    if err := s.repo.UpdateSyncStatus(ctx, "syncing", nil); err != nil {
        return nil, err
    }

    // Fetch version.json from GitHub
    versions, err := s.githubClient.FetchVersions(ctx)
    if err != nil {
        s.repo.UpdateSyncStatus(ctx, "error", &err)
        return nil, err
    }

    // Process and store versions
    for _, v := range versions {
        existing, _ := s.repo.GetVersionByVersion(ctx, v.Version)
        if existing != nil {
            // Update existing
            v.ID = existing.ID
            if err := s.repo.UpdateVersion(ctx, &v); err != nil {
                continue
            }
        } else {
            // Insert new
            if err := s.repo.CreateVersion(ctx, &v); err != nil {
                continue
            }
        }
    }

    // Update latest flag
    if err := s.repo.UpdateLatestFlag(ctx, versions[0].Version); err != nil {
        return nil, err
    }

    // Update sync status to synced
    nextSync := time.Now().Add(24 * time.Hour)
    if err := s.repo.UpdateSyncStatus(ctx, "synced", &nextSync); err != nil {
        return nil, err
    }

    return &SyncResult{
        Status:       "synced",
        StartedAt:    time.Now().UnixMilli(),
        VersionsFound: len(versions),
    }, nil
}
```

---

## 8. GraphQL Schema

### 8.1 Types

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

type UpdateStatusResult {
  sync: SyncStatus!
  latest: UpdateVersion
  device: DeviceUpdateStatus
}

type DeviceUpdateStatus {
  currentVersion: String
  needsUpdate: Boolean!
}
```

### 8.2 Queries

```graphql
type Query {
  updatesStatus: UpdateStatusResult!
  updatesVersions(status: String, page: Int, limit: Int): VersionsResult!
  updatesChangelog(version: String): ChangelogResult!
  updatesHistory(status: String, page: Int, limit: Int): HistoryResult!
  updatesHistoryDetail(pushId: ID!): UpdatePush
  updatesSyncStatus: SyncStatus!
}

type VersionsResult {
  versions: [UpdateVersion!]!
  pagination: Pagination!
}

type ChangelogResult {
  changelog: [ChangelogEntry!]!
}

type HistoryResult {
  pushes: [UpdatePush!]!
  pagination: Pagination!
}
```

### 8.3 Mutations

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

## 9. Error Handling

### 9.1 Error Response Format

```json
{
  "error": "error_code",
  "message": "Human readable message",
  "details": {}
}
```

### 9.2 Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| bad_request | 400 | Invalid request |
| version_not_found | 400 | APK version not found |
| push_not_found | 404 | Update push not found |
| push_not_cancellable | 400 | Push already completed/failed |
| sync_already_in_progress | 409 | Sync in progress |
| internal_error | 500 | Server error |

---

## 10. Rate Limiting & Security

### 10.1 Rate Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| GET /v1/updates/status | 60 | 1 minute |
| GET /v1/updates/versions | 30 | 1 minute |
| GET /v1/updates/changelog | 30 | 1 minute |
| POST /v1/updates/push | 10 | 1 minute |
| GET /v1/updates/history | 30 | 1 minute |
| POST /v1/updates/history/:id/cancel | 10 | 1 minute |
| POST /v1/updates/sync | 5 | 1 hour |

### 10.2 Security Requirements

1. **Authentication** - All endpoints require authenticated operator
2. **Authorization** - Only admins can push updates
3. **Audit Logging** - Log all push and sync operations
4. **Version Validation** - Verify APK hash before serving
5. **GitHub Webhook** - Secure webhook endpoint with secret validation

---

## 11. File Changes Summary

### 11.1 Total File Count

| Category | New | Modified | Total |
|----------|-----|---------|-------|
| Domain Layer | 3 | 0 | 3 |
| Application Layer | 6 | 0 | 6 |
| Handler Layer | 5 | 1 | 6 |
| Infrastructure | 4 | 1 | 5 |
| GraphQL | 2 | 2 | 4 |
| Router | 0 | 1 | 1 |
| **TOTAL** | **20** | **5** | **25** |

### 11.2 All Files Listed

#### Domain Layer (3 NEW)

| File | Status | Purpose |
|------|--------|---------|
| domain/updates/updates_entity.go | NEW | Update entities |
| domain/updates/updates_repository.go | NEW | Repository interface |
| domain/updates/updates_errors.go | NEW | Domain errors |

#### Application Layer (6 NEW)

| File | Status | Purpose |
|------|--------|---------|
| application/updates/updates_service.go | NEW | Main service |
| application/updates/updates_versions_service.go | NEW | Version management |
| application/updates/updates_push_service.go | NEW | Push management |
| application/updates/updates_history_service.go | NEW | History management |
| application/updates/updates_sync_service.go | NEW | GitHub sync |
| application/updates/updates_dto.go | NEW | DTOs |

#### Handler Layer (5 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| api/handlers/updates/updates_versions_handler.go | NEW | Versions handlers |
| api/handlers/updates/updates_push_handler.go | NEW | Push handlers |
| api/handlers/updates/updates_history_handler.go | NEW | History handlers |
| api/handlers/updates/updates_sync_handler.go | NEW | Sync handlers |
| api/handlers/updates/updates_routes.go | NEW | Route registration |
| api/handlers/router.go | MODIFIED | Add updates routes |

#### Infrastructure (4 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| infrastructure/storage/updates_storage.go | NEW | Updates queries |
| infrastructure/storage/migrations/ | NEW | SQL migrations |
| infrastructure/github/github_client.go | NEW | GitHub API client |
| infrastructure/github/github_sync.go | NEW | GitHub sync logic |
| infrastructure/storage/device_storage.go | MODIFIED | Add device queries |

#### GraphQL (2 NEW, 2 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| api/graphql/schema/objects.go | MODIFIED | Add updates types |
| api/graphql/schema/resolver.go | MODIFIED | Add resolvers |
| api/graphql/schema/schema.go | MODIFIED | Add root types |

---

## 12. Implementation Order

### Phase 1: Database (Day 1)
1. Create update_versions table
2. Create update_pushes table
3. Create update_push_devices table
4. Create update_sync_status table
5. Test migrations

### Phase 2: Domain Layer (Day 1)
1. Create `domain/updates/updates_entity.go`
2. Create `domain/updates/updates_repository.go`
3. Create `domain/updates/updates_errors.go`

### Phase 3: Infrastructure (Day 1-2)
1. Create `infrastructure/github/github_client.go`
2. Create `infrastructure/github/github_sync.go`
3. Create `infrastructure/storage/updates_storage.go`
4. Test database queries

### Phase 4: Application Layer (Day 2)
1. Create versions service methods
2. Create push service methods
3. Create history service methods
4. Create sync service methods

### Phase 5: Handlers (Day 2-3)
1. Create versions handler
2. Create push handler
3. Create history handler
4. Create sync handler
5. Wire routes
6. Test endpoints

### Phase 6: GraphQL (Day 3)
1. Add schema types
2. Add resolvers
3. Test queries

### Phase 7: Integration (Day 3-4)
1. Integrate with existing command system
2. Test FCM integration
3. Test GitHub webhook
4. E2E tests

---

*Document Version: 1.0*
*Status: Ready for Implementation*
