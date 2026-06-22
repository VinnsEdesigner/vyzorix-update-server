# Updates Page - Enterprise Requirements Specification

> **Version:** 1.0  
> **Status:** Draft  
> **Created:** 2026-06-22  
> **Target:** Production MVP  

---

## Table of Contents

1. [Overview](#1-overview)
2. [Architecture](#2-architecture)
3. [Tab 1: Status](#3-tab-1-status)
4. [Tab 2: Versions](#4-tab-2-versions)
5. [Tab 3: Push](#5-tab-3-push)
6. [Tab 4: Changelog](#6-tab-4-changelog)
7. [Tab 5: Export](#7-tab-5-export)
8. [Tab 6: Past Updates](#8-tab-6-past-updates)
9. [REST API Specification](#9-rest-api-specification)
10. [GraphQL Schema](#10-graphql-schema)
11. [Frontend Components](#11-frontend-components)
12. [File Changes Summary](#12-file-changes-summary)
13. [Implementation Order](#13-implementation-order)

---

## 1. Overview

### 1.1 Purpose

The Updates page allows operators to:
- View current and available versions
- Push updates to registered devices
- View changelogs across versions
- Export version data
- View history of past updates

### 1.2 Data Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│  GitHub Repository                                                  │
│  └── bin/                                                           │
│      ├── version.json        (latest version metadata)              │
│      ├── changelog.json      (release notes)                        │
│      ├── v2.2.1/             (version folder)                      │
│      │   └── VyzorixAudioRouter.apk                              │
│      ├── v2.2.0/                                                     │
│      └── v2.1.0/                                                     │
│           └── VyzorixAudioRouter.apk                              │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              │ GitHub Runner Pushes
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│  Update Server                                                     │
│  ├── Syncs from GitHub bin/ folder periodically                    │
│  ├── Stores version metadata in database                           │
│  ├── Serves APK files to devices                                  │
│  └── Tracks update history per device                             │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              │ REST / GraphQL
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│  Frontend (Updates Page)                                           │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 1.3 Page Structure

```
/updates
├── [Tabs]
│   ├── Status      → Current state, sync status
│   ├── Versions    → All available versions
│   ├── Push        → Push updates to devices
│   ├── Changelog   → Release notes
│   ├── Export      → Export version data
│   └── History     → Past updates
```

---

## 2. Architecture

### 2.1 Directory Structure (Following Frontend Architecture)

```
src/
├── domain/
│   └── updates/
│       ├── types.ts           # Version, Changelog, UpdateHistory
│       ├── transforms.ts      # versionFromAPI, etc.
│       └── validation.ts     # validateVersion, etc.
│
├── lib/
│   └── api/
│       ├── graphql/
│       │   ├── queries/
│       │   │   └── updates.ts
│       │   └── mutations/
│       │       └── updates.ts
│       └── rest/
│           └── endpoints.ts   # /v1/updates/*
│
├── hooks/
│   └── updates/
│       ├── use-versions.ts
│       ├── use-push-update.ts
│       ├── use-changelog.ts
│       ├── use-update-history.ts
│       └── use-sync-status.ts
│
└── ui/
    ├── pages/
    │   └── updates/
    │       ├── updates.tsx              # Main page with tabs
    │       ├── updates-status.tsx
    │       ├── updates-versions.tsx
    │       ├── updates-push.tsx
    │       ├── updates-changelog.tsx
    │       ├── updates-export.tsx
    │       └── updates-history.tsx
    │
    └── components/
        └── shared/
            ├── version-card/
            │   ├── version-card.tsx
            │   ├── version-details.tsx
            │   └── version-badge.tsx
            ├── push-form/
            │   ├── push-form.tsx
            │   ├── device-selector.tsx
            │   └── install-type.tsx
            └── changelog-entry/
                └── changelog-entry.tsx
```

### 2.2 Dependency Flow

```
UI Layer (pages/components)
        │
        │ uses
        ▼
Presentation Layer (hooks)
        │
        │ uses
        ▼
Domain Layer (types, transforms)
        │
        │ uses
        ▼
Data Layer (GraphQL/REST clients)
```

---

## 3. Tab 1: Status

### 3.1 Purpose

Show current state of the update system:
- Current device version vs latest available
- Sync status with GitHub
- Quick actions

### 3.2 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  UPDATES                                           [Sync Now] [Refresh]│
├─────────────────────────────────────────────────────────────────────┤
│  [Status] [Versions] [Push] [Changelog] [Export] [History]        │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─ DEVICE VERSION ──────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  VyzorixAudioRouter                                        │   │
│  │  v2.1.0 ──────────────────── [✓] Up to date                 │   │
│  │                                                              │   │
│  │  Installed: Jun 10, 2026  •  Size: 11.8 MB              │   │
│  │  SHA256: a1b2c3d4...                                        │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ AVAILABLE UPDATE ────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  v2.2.1 ────────────────── [!] Update available            │   │
│  │                                                              │   │
│  │  Released: Jun 22, 2026  •  Size: 12.4 MB                │   │
│  │  Changes: 5 features, 3 fixes                             │   │
│  │                                                              │   │
│  │  [View Changes]  [Push to Device]                        │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ SYNC STATUS ─────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  Source: github.com/vyzorix/.../bin                     │   │
│  │  Last Sync: 2 minutes ago  •  Status: ✓ Connected       │   │
│  │  New Versions: 1                                          │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.3 Sections

| Section | Content | Data Source |
|---------|---------|------------|
| **Device Version** | Current installed version on selected device | Device API |
| **Available Update** | Latest version available from GitHub | Updates API |
| **Sync Status** | Connection to GitHub, last sync time | Updates API |

### 3.4 Interactions

| Element | Action |
|---------|--------|
| "Sync Now" | Manually trigger GitHub sync |
| "Refresh" | Refresh all data |
| "View Changes" | Navigate to Changelog tab |
| "Push to Device" | Navigate to Push tab |

---

## 4. Tab 2: Versions

### 4.1 Purpose

Display all available versions with details.

### 4.2 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  VERSIONS                                            [Filter ▾]     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─ v2.2.1 ─────────────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  LATEST                                                   │   │
│  │  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │   │
│  │                                                              │   │
│  │  Released: Jun 22, 2026  •  Size: 12.4 MB  •  SHA256: d4e5f6│   │
│  │                                                              │   │
│  │  ┌───────────────────────────────────────────────────────┐   │   │
│  │  │  ● Security patches (3)                            │   │   │
│  │  │  ● New features (2)                               │   │   │
│  │  │  ● Bug fixes (0)                                   │   │   │
│  │  └───────────────────────────────────────────────────────┘   │   │
│  │                                                              │   │
│  │  [View Full Changelog]  [Push to Device]                 │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ v2.2.0 ─────────────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │   │
│  │                                                              │   │
│  │  Released: Jun 21, 2026  •  Size: 12.0 MB                 │   │
│  │                                                              │   │
│  │  [View Full Changelog]  [Push to Device]                   │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ v2.1.0 ─────────────────────────────── INSTALLED ────────┐   │
│  │                                                              │   │
│  │  Released: Jun 10, 2026  •  Size: 11.8 MB                 │   │
│  │                                                              │   │
│  │  [View Full Changelog]                                      │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ v2.0.0 ─────────────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  Released: May 28, 2026  •  Size: 11.2 MB                 │   │
│  │                                                              │   │
│  │  [View Full Changelog]  [Push to Device]                   │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 4.3 Version Card States

| State | Badge | Actions |
|-------|-------|---------|
| **Latest** | `LATEST` (primary) | View Changes, Push |
| **Current** | `INSTALLED` (success) | View Changes |
| **Previous** | None | View Changes, Push |
| **Deprecated** | `DEPRECATED` (warning) | View Changes |

### 4.4 Filter Options

| Filter | Description |
|--------|-------------|
| All | Show all versions |
| Latest Only | Show only newest version |
| With Updates | Show versions with available updates |
| Security | Show only security releases |

---

## 5. Tab 3: Push

### 5.1 Purpose

Push updates to registered devices.

### 5.2 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  PUSH UPDATE                                                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─ TARGET ───────────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  Select device(s):                                         │   │
│  │                                                              │   │
│  │  ┌─────────────────────────────────────────────────────┐   │   │
│  │  │  [Search devices...]                                │   │   │
│  │  └─────────────────────────────────────────────────────┘   │   │
│  │                                                              │   │
│  │  ┌─────────────────────────────────────────────────────┐   │   │
│  │  │  ● Pixel 8 Pro (861234...)              Online   │   │   │
│  │  │    Current: v2.1.0  •  Status: Compatible       │   │   │
│  │  └─────────────────────────────────────────────────────┘   │   │
│  │  ┌─────────────────────────────────────────────────────┐   │   │
│  │  │  ○ Nokia C22 (352345...)              Offline    │   │   │
│  │  │    Current: v2.0.0  •  Status: Compatible       │   │   │
│  │  └─────────────────────────────────────────────────────┘   │   │
│  │                                                              │   │
│  │  ○ All registered devices (12)                              │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ VERSION ─────────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  Select version to push:                                   │   │
│  │                                                              │   │
│  │  [v2.2.1 ▾]  (Latest)                                    │   │
│  │                                                              │   │
│  │  ┌───────────────────────────────────────────────────────┐   │   │
│  │  │  v2.2.1 Details                                   │   │   │
│  │  │  Released: Jun 22, 2026  •  Size: 12.4 MB        │   │   │
│  │  │  Changes: 5 features, 3 fixes, 2 security         │   │   │
│  │  └───────────────────────────────────────────────────────┘   │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ INSTALL TYPE ───────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  ○ Mandatory                                                │   │
│  │    Device must install immediately, cannot defer              │   │
│  │                                                              │   │
│  │  ● Optional                                                 │   │
│  │    Device can defer installation                            │   │
│  │                                                              │   │
│  │  ○ Silent                                                   │   │
│  │    Install without user interaction (requires confirmation)   │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ SCHEDULE (Optional) ────────────────────────────────────┐   │
│  │                                                              │   │
│  │  ○ Push immediately                                        │   │
│  │  ● Schedule for later                                      │   │
│  │                                                              │   │
│  │  Date: [Jun 23, 2026    ]  Time: [02:00  ]            │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ SUMMARY ─────────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  Devices: 1                                               │   │
│  │  Version: v2.2.1                                         │   │
│  │  Install Type: Optional                                    │   │
│  │  When: Immediately                                        │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│                           [Push Update]                             │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 5.3 Push Form Sections

| Section | Purpose |
|---------|---------|
| **Target** | Select device(s) to push to |
| **Version** | Select version to push |
| **Install Type** | Mandatory, Optional, or Silent |
| **Schedule** | Immediate or scheduled |
| **Summary** | Confirmation before push |

### 5.4 Install Types

| Type | Description | Use Case |
|------|-------------|----------|
| **Mandatory** | Device must install, no defer option | Critical security updates |
| **Optional** | Device user can defer | Feature updates |
| **Silent** | No user interaction required | Automated deployments |

### 5.5 Device Selection States

| State | Checkbox | Details |
|-------|----------|---------|
| **Online** | ● | Can push immediately |
| **Offline** | ○ | Will push when device comes online |
| **Incompatible** | ⊘ (disabled) | Version not compatible |
| **Already on version** | ⊘ (disabled) | Already on this version |

---

## 6. Tab 4: Changelog

### 6.1 Purpose

View detailed release notes for each version.

### 6.2 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  CHANGELOG                                      [Version ▾] [Search] │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  FILTER BY:                                                        │
│  [All ▾] [Features] [Bug Fixes] [Security] [Performance]         │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─ v2.2.1 ─────────────────────────────────────────────────────┐   │
│  │  Released: Jun 22, 2026  •  5 changes                      │   │
│  │                                                              │   │
│  │  ┌───────────────────────────────────────────────────────┐   │   │
│  │  │  ● SECURITY (3)                                   │   │   │
│  │  │                                                       │   │   │
│  │  │  • Updated Argon2id parameters for command signing   │   │   │
│  │  │  • Fixed certificate validation in HTTPS calls        │   │   │
│  │  │  • Patched CVE-2026-1234: Memory corruption          │   │   │
│  │  │                                                       │   │   │
│  │  ├───────────────────────────────────────────────────────┤   │   │
│  │  │  ● FEATURES (2)                                    │   │   │
│  │  │                                                       │   │   │
│  │  │  • Improved WebSocket reconnection with exponential  │   │   │
│  │  │    backoff algorithm                                │   │   │
│  │  │  • Added GZIP compression for telemetry data         │   │   │
│  │  │                                                       │   │   │
│  │  └───────────────────────────────────────────────────────┘   │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ v2.2.0 ─────────────────────────────────────────────────────┐   │
│  │  Released: Jun 21, 2026  •  3 changes                      │   │
│  │                                                              │   │
│  │  ┌───────────────────────────────────────────────────────┐   │   │
│  │  │  ● FEATURES (2)                                       │   │   │
│  │  │                                                       │   │   │
│  │  │  • New telemetry filtering by device ID                │   │   │
│  │  │  • Added device grouping support                       │   │   │
│  │  │                                                       │   │   │
│  │  ├───────────────────────────────────────────────────────┤   │   │
│  │  │  ● BUG FIXES (1)                                     │   │   │
│  │  │                                                       │   │   │
│  │  │  • Fixed memory leak in telemetry buffering           │   │   │
│  │  │                                                       │   │   │
│  │  └───────────────────────────────────────────────────────┘   │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ v2.1.0 ─────────────────────────────────────────────────────┐   │
│  │  Released: Jun 10, 2026  •  4 changes                        │   │
│  │                                                              │   │
│  │  ┌───────────────────────────────────────────────────────┐   │   │
│  │  │  ● FEATURES (4)                                     │   │   │
│  │  │                                                       │   │   │
│  │  │  • Initial release with FCM push notification support │   │   │
│  │  │  • Added device registration via REST API              │   │   │
│  │  │  • Implemented HMAC command signing                   │   │   │
│  │  │  • Added real-time WebSocket streaming               │   │   │
│  │  │                                                       │   │   │
│  │  └───────────────────────────────────────────────────────┘   │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 6.3 Change Categories

| Category | Icon | Description |
|----------|------|-------------|
| **Security** | 🔒 | Security patches and fixes |
| **Features** | ✨ | New functionality |
| **Bug Fixes** | 🐛 | Bug corrections |
| **Performance** | ⚡ | Performance improvements |
| **Breaking** | ⚠️ | Breaking changes (if any) |

### 6.4 Filters

| Filter | Shows |
|--------|-------|
| All | All changes |
| Features | Only new features |
| Bug Fixes | Only bug fixes |
| Security | Only security changes |
| Performance | Only performance changes |

---

## 7. Tab 5: Export

### 7.1 Purpose

Export version data in various formats.

### 7.2 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  EXPORT                                                      [Export] │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─ SELECT DATA TO EXPORT ────────────────────────────────────┐   │
│  │                                                              │   │
│  │  ○ All versions                                             │   │
│  │    Complete version history with changelogs                │   │
│  │                                                              │   │
│  │  ● Current version only                                    │   │
│  │    Latest version with full changelog                       │   │
│  │                                                              │   │
│  │  ○ Version range                                          │   │
│  │    From: [v2.1.0 ▾]  To: [v2.2.1 ▾]                    │   │
│  │                                                              │   │
│  │  ○ Changelog only                                          │   │
│  │    Release notes without version metadata                    │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ INCLUDE ────────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  [✓] Version metadata (version, date, size, SHA256)       │   │
│  │  [✓] Changelog entries                                    │   │
│  │  [ ] APK file (WARNING: Large file, ~12MB each)          │   │
│  │  [ ] Compatibility matrix                                  │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ FORMAT ─────────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  (●) JSON                                                  │   │
│  │      Structured data format, ideal for CI/CD integration    │   │
│  │                                                              │   │
│  │  ( ) CSV                                                    │   │
│  │      Tabular format, ideal for spreadsheets                 │   │
│  │                                                              │   │
│  │  ( ) Markdown                                              │   │
│  │      Human-readable format for documentation               │   │
│  │                                                              │   │
│  │  ( ) XML                                                   │   │
│  │      Enterprise format for legacy systems                   │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│                           [Generate Export]                         │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 7.3 Export Options

| Option | Description |
|--------|-------------|
| **All versions** | Complete version history |
| **Current version** | Only latest version |
| **Version range** | Specific version range |
| **Changelog only** | Release notes without metadata |

### 7.4 Export Formats

| Format | Use Case | Example |
|--------|----------|---------|
| **JSON** | CI/CD, programmatic parsing | API integration |
| **CSV** | Spreadsheets, data analysis | Excel, Google Sheets |
| **Markdown** | Documentation, release notes | GitHub README |
| **XML** | Enterprise legacy systems | SAP integration |

### 7.5 Export Examples

#### JSON Export
```json
{
  "exportedAt": "2026-06-22T10:30:00Z",
  "generator": "Vyzorix Dashboard",
  "version": "1.0",
  "versions": [
    {
      "version": "2.2.1",
      "versionCode": 221,
      "releasedAt": "2026-06-22T00:00:00Z",
      "apkSize": 12400000,
      "apkSha256": "d4e5f6...",
      "changes": [
        {
          "type": "security",
          "description": "Updated Argon2id parameters"
        }
      ]
    }
  ]
}
```

#### CSV Export
```csv
version,version_code,released_at,apk_size,change_type,description
2.2.1,221,2026-06-22,12400000,security,Updated Argon2id parameters
2.2.1,221,2026-06-22,12400000,security,Fixed certificate validation
2.2.1,221,2026-06-22,12400000,feature,Improved WebSocket reconnection
2.2.1,221,2026-06-22,12400000,feature,Added GZIP compression
```

#### Markdown Export
```markdown
# VyzorixAudioRouter Changelog

## v2.2.1 (June 22, 2026)

### Security
- Updated Argon2id parameters for command signing
- Fixed certificate validation in HTTPS calls

### Features
- Improved WebSocket reconnection with exponential backoff
- Added GZIP compression for telemetry data

## v2.2.0 (June 21, 2026)
...
```

---

## 8. Tab 6: Past Updates

### 8.1 Purpose

View history of updates pushed to devices.

### 8.2 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  UPDATE HISTORY                            [Filter ▾] [Search]      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  FILTER BY STATUS:                                                 │
│  [All] [Success] [Pending] [Failed] [Scheduled]                 │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─ TODAY (3) ────────────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  ┌─────────────────────────────────────────────────────┐   │   │
│  │  │  Pixel 8 Pro                                        │   │   │
│  │  │  v2.1.0 → v2.2.1                                   │   │   │
│  │  │  Status: ● Completed  •  12:34:56                    │   │   │
│  │  │  Type: Optional  •  Duration: 45s                    │   │   │
│  │  │  [View Details]                                      │   │   │
│  │  └─────────────────────────────────────────────────────┘   │   │
│  │                                                              │   │
│  │  ┌─────────────────────────────────────────────────────┐   │   │
│  │  │  Nokia C22                                           │   │   │
│  │  │  v2.0.0 → v2.2.0                                   │   │   │
│  │  │  Status: ◐ In Progress  •  Started: 12:30:00         │   │   │
│  │  │  Type: Mandatory  •  Progress: 67%                   │   │   │
│  │  │  [View Details]  [Cancel]                            │   │   │
│  │  └─────────────────────────────────────────────────────┘   │   │
│  │                                                              │   │
│  │  ┌─────────────────────────────────────────────────────┐   │   │
│  │  │  Samsung S24                                          │   │   │
│  │  │  v2.1.0 → v2.2.1                                   │   │   │
│  │  │  Status: ⏱ Scheduled  •  14:00:00                   │   │   │
│  │  │  Type: Silent  •  ETA: ~1h 26m                      │   │   │
│  │  │  [View Details]  [Reschedule]  [Cancel]             │   │   │
│  │  └─────────────────────────────────────────────────────┘   │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ YESTERDAY (5) ──────────────────────────────────────────┐   │
│  │                                                              │   │
│  │  ┌─────────────────────────────────────────────────────┐   │   │
│  │  │  Pixel 8 Pro                                        │   │   │
│  │  │  v2.0.0 → v2.1.0                                   │   │   │
│  │  │  Status: ✓ Completed  •  Jun 21, 14:32:00             │   │   │
│  │  │  Type: Optional  •  Duration: 52s                    │   │   │
│  │  │  [View Details]                                      │   │   │
│  │  └─────────────────────────────────────────────────────┘   │   │
│  │                                                              │   │
│  │  ┌─────────────────────────────────────────────────────┐   │   │
│  │  │  Nokia C22                                           │   │   │
│  │  │  v1.5.2 → v2.0.0                                   │   │   │
│  │  │  Status: ✗ Failed  •  Jun 21, 10:15:00              │   │   │
│  │  │  Error: Device offline during update window           │   │   │
│  │  │  [Retry]  [View Details]                           │   │   │
│  │  └─────────────────────────────────────────────────────┘   │   │
│  │                                                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  [Load More]                                                       │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 8.3 Update Status States

| Status | Icon | Color | Description |
|--------|------|-------|-------------|
| **Completed** | ✓ | Green | Update successfully installed |
| **In Progress** | ◐ | Blue | Currently downloading/installing |
| **Failed** | ✗ | Red | Update failed |
| **Scheduled** | ⏱ | Gray | Waiting for scheduled time |
| **Cancelled** | ⊘ | Gray | Operator cancelled |
| **Pending** | ○ | Yellow | Waiting to start |

### 8.4 Update Card Details

| Field | Description |
|-------|-------------|
| Device | Device name and IMEI |
| Version | From version → To version |
| Status | Current status with icon |
| Type | Mandatory, Optional, or Silent |
| Duration | Time taken to complete (if completed) |
| Progress | Download/install progress (if in progress) |
| Error | Error message (if failed) |

### 8.5 Actions

| Action | Available When | Purpose |
|--------|----------------|--------|
| View Details | Always | Show full update details |
| Cancel | In Progress, Scheduled | Cancel pending update |
| Retry | Failed | Retry failed update |
| Reschedule | Scheduled | Change scheduled time |

### 8.6 Update Details Modal

```
┌─────────────────────────────────────────────────────────────────────┐
│  UPDATE DETAILS                                              [Close] │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Device: Pixel 8 Pro                                           │
│  IMEI: 861234567890123                                         │
│                                                                     │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ │
│                                                                     │
│  VERSION                                                       │
│  From: v2.1.0  →  To: v2.2.1                                  │
│                                                                     │
│  STATUS                                                        │
│  ● Completed successfully                                       │
│  Started: Jun 22, 2026 12:30:00                                │
│  Completed: Jun 22, 2026 12:31:45                                │
│  Duration: 1m 45s                                              │
│                                                                     │
│  METRICS                                                        │
│  Download: 12.4 MB                                              │
│  Install Time: 45s                                              │
│  Reboot Required: Yes                                           │
│                                                                     │
│  LOGS                                                           │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  12:30:00  [INFO] Update initiated                       │   │
│  │  12:30:05  [INFO] Downloading APK...                    │   │
│  │  12:31:00  [INFO] Download complete (12.4 MB)            │   │
│  │  12:31:05  [INFO] Verifying APK signature...           │   │
│  │  12:31:10  [INFO] Signature valid                      │   │
│  │  12:31:15  [INFO] Installing APK...                    │   │
│  │  12:31:40  [INFO] Installation complete                │   │
│  │  12:31:45  [INFO] Update completed successfully        │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
│                           [Close]  [Rollback to Previous]          │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 9. REST API Specification

### 9.1 Endpoints

#### `GET /v1/updates`
**Purpose:** Get all available versions

**Response:**
```json
{
  "versions": [
    {
      "id": "uuid",
      "version": "2.2.1",
      "versionCode": 221,
      "releasedAt": "2026-06-22T00:00:00Z",
      "apkSize": 12400000,
      "apkSha256": "d4e5f6...",
      "isLatest": true,
      "changelog": {
        "security": 3,
        "features": 2,
        "bugFixes": 0
      }
    }
  ],
  "syncStatus": {
    "lastSync": "2026-06-22T10:28:00Z",
    "source": "github.com/vyzorix/...",
    "status": "connected"
  }
}
```

#### `GET /v1/updates/:version`
**Purpose:** Get specific version details

**Response:**
```json
{
  "id": "uuid",
  "version": "2.2.1",
  "versionCode": 221,
  "releasedAt": "2026-06-22T00:00:00Z",
  "apkSize": 12400000,
  "apkSha256": "d4e5f6...",
  "changelog": [
    {
      "type": "security",
      "description": "Updated Argon2id parameters for command signing"
    },
    {
      "type": "feature",
      "description": "Improved WebSocket reconnection with exponential backoff"
    }
  ]
}
```

#### `GET /v1/updates/:version/changelog`
**Purpose:** Get changelog for version

**Response:**
```json
{
  "version": "2.2.1",
  "changes": [
    {
      "type": "security",
      "description": "Updated Argon2id parameters for command signing"
    }
  ]
}
```

#### `POST /v1/updates/push`
**Purpose:** Push update to device(s)

**Request:**
```json
{
  "imei": "861234567890123",
  "version": "2.2.1",
  "installType": "optional",
  "scheduledAt": null
}
```

**Response:**
```json
{
  "updateId": "uuid",
  "status": "pending",
  "device": {
    "imei": "861234567890123",
    "name": "Pixel 8 Pro"
  },
  "version": "2.2.1",
  "scheduledAt": null
}
```

#### `GET /v1/updates/history`
**Purpose:** Get update history

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | string | all | Filter: completed, failed, pending, scheduled |
| `page` | int | 1 | Page number |
| `limit` | int | 20 | Items per page |

**Response:**
```json
{
  "updates": [
    {
      "id": "uuid",
      "device": {
        "imei": "861234567890123",
        "name": "Pixel 8 Pro"
      },
      "fromVersion": "2.1.0",
      "toVersion": "2.2.1",
      "status": "completed",
      "installType": "optional",
      "startedAt": "2026-06-22T12:30:00Z",
      "completedAt": "2026-06-22T12:31:45Z",
      "duration": 105
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

#### `DELETE /v1/updates/:updateId`
**Purpose:** Cancel pending/scheduled update

**Response:**
```json
{
  "id": "uuid",
  "status": "cancelled"
}
```

#### `POST /v1/updates/sync`
**Purpose:** Manually sync from GitHub

**Response:**
```json
{
  "status": "syncing",
  "lastSync": "2026-06-22T10:30:00Z",
  "newVersions": 1,
  "updatedVersions": 0
}
```

---

## 10. GraphQL Schema

### 10.1 Types

```graphql
enum InstallType {
  MANDATORY
  OPTIONAL
  SILENT
}

enum UpdateStatus {
  PENDING
  IN_PROGRESS
  COMPLETED
  FAILED
  CANCELLED
  SCHEDULED
}

enum ChangeType {
  SECURITY
  FEATURE
  BUG_FIX
  PERFORMANCE
  BREAKING
}

type Version {
  id: ID!
  version: String!
  versionCode: Int!
  releasedAt: DateTime!
  apkSize: Int!
  apkSha256: String!
  isLatest: Boolean!
  changelogSummary: ChangelogSummary!
}

type ChangelogSummary {
  security: Int!
  features: Int!
  bugFixes: Int!
  performance: Int!
}

type ChangelogEntry {
  type: ChangeType!
  description: String!
}

type Changelog {
  version: Version!
  changes: [ChangelogEntry!]!
}

type SyncStatus {
  lastSync: DateTime!
  source: String!
  status: String!
}

type DeviceUpdate {
  id: ID!
  device: DeviceSummary!
  fromVersion: String!
  toVersion: String!
  status: UpdateStatus!
  installType: InstallType!
  startedAt: DateTime
  completedAt: DateTime
  duration: Int
  progress: Int
  error: String
}

type UpdateHistoryConnection {
  updates: [DeviceUpdate!]!
  pagination: PaginationInfo!
}

type UpdatePushResult {
  success: Boolean!
  updateId: String
  status: UpdateStatus
  message: String
  error: String
}
```

### 10.2 Queries

```graphql
type Query {
  # Get all versions
  versions: [Version!]!
  
  # Get latest version
  latestVersion: Version!
  
  # Get specific version
  version(version: String!): Version
  
  # Get changelog for version
  changelog(version: String!): Changelog!
  
  # Get update history
  updateHistory(
    status: UpdateStatus
    page: Int = 1
    limit: Int = 20
  ): UpdateHistoryConnection!
  
  # Get sync status
  syncStatus: SyncStatus!
}
```

### 10.3 Mutations

```graphql
type Mutation {
  # Push update to device
  pushUpdate(
    imei: String!
    version: String!
    installType: InstallType!
    scheduledAt: DateTime
  ): UpdatePushResult!
  
  # Push update to all devices
  pushUpdateAll(
    version: String!
    installType: InstallType!
    scheduledAt: DateTime
  ): [UpdatePushResult!]!
  
  # Cancel update
  cancelUpdate(updateId: ID!): DeviceUpdate!
  
  # Retry failed update
  retryUpdate(updateId: ID!): UpdatePushResult!
  
  # Sync from GitHub
  syncUpdates: SyncStatus!
}
```

---

## 11. Frontend Components

### 11.1 Pages

| File | Purpose |
|------|---------|
| `updates.tsx` | Main page with tabs |
| `updates-status.tsx` | Status tab |
| `updates-versions.tsx` | Versions tab |
| `updates-push.tsx` | Push tab |
| `updates-changelog.tsx` | Changelog tab |
| `updates-export.tsx` | Export tab |
| `updates-history.tsx` | History tab |

### 11.2 Shared Components

| Component | Purpose |
|----------|---------|
| `VersionCard` | Version display card |
| `VersionBadge` | Latest/Installed/Previous badge |
| `VersionDetails` | Expandable version details |
| `PushForm` | Update push form |
| `DeviceSelect` | Multi-select for devices |
| `InstallTypeSelect` | Install type radio buttons |
| `ScheduleSelect` | Schedule date/time picker |
| `ChangelogEntry` | Single changelog entry |
| `UpdateHistoryCard` | Update history card |
| `UpdateDetailsModal` | Update details modal |
| `ExportOptions` | Export configuration form |

### 11.3 Hooks

| Hook | Purpose |
|------|---------|
| `useVersions` | Fetch all versions |
| `useLatestVersion` | Fetch latest version |
| `useVersionDetails` | Fetch specific version |
| `useChangelog` | Fetch changelog |
| `usePushUpdate` | Push update mutation |
| `useUpdateHistory` | Fetch update history |
| `useCancelUpdate` | Cancel update mutation |
| `useSyncUpdates` | Sync from GitHub |
| `useExport` | Export data |

---

## 12. File Changes Summary

### 12.1 NEW Files (Frontend)

#### Domain Layer
```
src/domain/updates/
├── types.ts
├── transforms.ts
└── validation.ts
```

#### Data Layer
```
src/lib/api/graphql/
├── queries/
│   └── updates.ts
└── mutations/
    └── updates.ts

src/lib/api/rest/
└── endpoints.ts
```

#### Presentation Layer
```
src/hooks/updates/
├── use-versions.ts
├── use-changelog.ts
├── use-push-update.ts
├── use-update-history.ts
├── use-sync-status.ts
└── use-export.ts
```

#### UI Layer
```
src/ui/pages/updates/
├── updates.tsx
├── updates-status.tsx
├── updates-versions.tsx
├── updates-push.tsx
├── updates-changelog.tsx
├── updates-export.tsx
└── updates-history.tsx

src/ui/components/shared/
├── version-card/
│   ├── version-card.tsx
│   └── version-badge.tsx
├── push-form/
│   ├── push-form.tsx
│   ├── device-select.tsx
│   └── install-type-select.tsx
├── changelog-entry/
│   └── changelog-entry.tsx
└── update-history/
    ├── update-history-card.tsx
    └── update-details-modal.tsx
```

### 12.2 MODIFIED Files (Frontend)

| File | Changes |
|------|---------|
| `src/routes/updates.tsx` | Replace with new tabbed structure |
| `src/lib/api/graphql/queries.ts` | Add updates queries |
| `src/lib/api/graphql/mutations.ts` | Add updates mutations |
| `src/lib/api/rest/endpoints.ts` | Add updates endpoints |

### 12.3 NEW Files (Backend)

| File | Purpose |
|------|---------|
| `internal/api/handlers/updates/versions.go` | GET /v1/updates |
| `internal/api/handlers/updates/push.go` | POST /v1/updates/push |
| `internal/api/handlers/updates/history.go` | GET /v1/updates/history |
| `internal/api/handlers/updates/sync.go` | POST /v1/updates/sync |
| `internal/domain/updates/types.go` | Update types |
| `internal/infrastructure/storage/updates.go` | Update storage |

### 12.4 MODIFIED Files (Backend)

| File | Changes |
|------|---------|
| `internal/api/router.go` | Add updates routes |
| `internal/api/handlers/github/sync.go` | Add GitHub sync logic |

---

## 13. Implementation Order

### Phase 1: Domain & Data (Day 1)
1. Create `src/domain/updates/` types
2. Create GraphQL queries and mutations
3. Create REST endpoints
4. Implement backend handlers

### Phase 2: Presentation Layer (Day 1-2)
1. Create `useVersions` hook
2. Create `useChangelog` hook
3. Create `usePushUpdate` hook
4. Create `useUpdateHistory` hook

### Phase 3: UI Components (Day 2-3)
1. Create `VersionCard` component
2. Create `PushForm` component
3. Create `ChangelogEntry` component
4. Create `UpdateHistoryCard` component

### Phase 4: Pages (Day 3-4)
1. Create `updates-versions.tsx`
2. Create `updates-push.tsx`
3. Create `updates-changelog.tsx`
4. Create `updates-history.tsx`
5. Create `updates-status.tsx`
6. Create `updates-export.tsx`

### Phase 5: Main Page (Day 4)
1. Create `updates.tsx` with tabs
2. Wire all hooks to pages
3. Add loading/error states

### Phase 6: Polish (Day 4-5)
1. Add animations
2. Test all flows
3. Add empty states
4. Add to existing tests

---

*Document Version: 1.0*  
*Status: Ready for Implementation*
