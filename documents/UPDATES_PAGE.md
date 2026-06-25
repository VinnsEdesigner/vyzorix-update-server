# Updates Page - Enterprise Requirements Specification

> **Version:** 1.1
> **Status:** Draft
> **Created:** 2026-06-22
> **Updated:** 2026-06-24
> **Target:** Production MVP
> **Architecture:** Layered (Following `FRONTEND_ARCHITECTURE.md`)

---

## Table of Contents

1. [Overview](#1-overview)
2. [Page Structure](#2-page-structure)
3. [Architecture](#3-architecture)
4. [Target File Structure](#4-target-file-structure)
5. [Tab 1: Status](#5-tab-1-status)
6. [Tab 2: Versions](#6-tab-2-versions)
7. [Tab 3: Push](#7-tab-3-push)
8. [Tab 4: Changelog](#8-tab-4-changelog)
9. [Tab 5: Export](#9-tab-5-export)
10. [Tab 6: History](#10-tab-6-history)
11. [Domain Layer](#11-domain-layer)
12. [Data Layer](#12-data-layer)
13. [Presentation Layer - Hooks](#13-presentation-layer---hooks)
14. [UI Layer - Components](#14-ui-layer---components)
15. [File Changes Summary](#15-file-changes-summary)
16. [Implementation Order](#16-implementation-order)
17. [Testing Strategy](#17-testing-strategy)
18. [Rollout Checklist](#18-rollout-checklist)

---

---

> **Architecture Alignment Note (v1.1)**
>
> This document has been updated to align with the **Layered Architecture** defined in `FRONTEND_ARCHITECTURE.md`. The file structure below follows the **4-layer architecture**:
> - **UI Layer** (`src/components/`) - Pure UI rendering, imports only from hooks
> - **Presentation Layer** (`src/hooks/`) - UI logic, state management, imports from domain & data
> - **Domain Layer** (`src/domain/` - NEW) - Types, transforms, validation (NO external imports)
> - **Data Layer** (`src/lib/api/`) - API clients (GraphQL/REST), imports only domain types
>
> **Dependency Rule:** UI -> Hooks -> Domain -> API (flow inward only)

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
GitHub Repository -> Update Server -> Frontend (Updates Page)
```

### 1.3 Page Structure

```
/updates
├── Status      -> Current state, sync status
├── Versions    -> All available versions
├── Push        -> Push updates to devices
├── Changelog   -> Release notes
├── Export      -> Export version data
└── History     -> Past updates
```

---

## 2. Page Structure

### 2.1 Routes (TanStack Start File-Based Routing)

| Route | File | Purpose |
|-------|------|---------|
| `/updates` | `updates-page.tsx` | Main layout |
| `/updates/status` | `updates.status.tsx` | Status tab |
| `/updates/versions` | `updates.versions.tsx` | Versions tab |
| `/updates/push` | `updates.push.tsx` | Push tab |
| `/updates/changelog` | `updates.changelog.tsx` | Changelog tab |
| `/updates/export` | `updates.export.tsx` | Export tab |
| `/updates/history` | `updates.history.tsx` | History tab |

**Total: 6 NEW routes, 1 MODIFIED route**

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
├── domain/                          # DOMAIN LAYER (NEW)
|   ├── common/
|   |   ├── pagination.ts            # Pagination types
|   |   ├── error.ts                # Domain error types
|   |   └── types.ts                # Shared domain types
|   |
|   └── updates/
|       ├── update-types.ts          # Version, Changelog, UpdatePush
|       ├── update-transforms.ts     # versionFromAPI()
|       └── update-validation.ts     # validateVersion()
|
├── lib/api/
|   └── graphql/
|       ├── queries/                 # GraphQL Queries
|       |   ├── update-queries.ts   # GET_VERSIONS, GET_CHANGELOG, etc.
|       |   └── index.ts
|       ├── mutations/               # GraphQL Mutations
|       |   ├── update-mutations.ts # PUSH_UPDATE, CANCEL_UPDATE
|       |   └── index.ts
|       ├── fragments/               # GraphQL fragments
|       |   ├── version.fragment.ts
|       |   └── index.ts
|       └── api-response-types.ts   # (MODIFIED - add update types)
|
├── hooks/                           # PRESENTATION LAYER
|   ├── updates/                    # Updates hooks
|   |   ├── use-versions.ts
|   |   ├── use-changelog.ts
|   |   ├── use-push-update.ts
|   |   ├── use-update-history.ts
|   |   ├── use-sync-status.ts
|   |   └── index.ts
|   └── shared/
|       └── use-pagination.ts
|
├── components/                      # UI LAYER
|   ├── shared/                    # Shared UI components
|   |   ├── section.tsx
|   |   ├── section-header.tsx
|   |   ├── empty-state.tsx
|   |   ├── loading-skeleton.tsx
|   |   ├── data-table.tsx
|   |   ├── pagination.tsx
|   |   ├── status-badge.tsx
|   |   ├── tab-nav.tsx
|   |   └── index.ts
|   |
|   └── updates/                   # Updates feature components
|       ├── version-card.tsx
|       ├── version-badge.tsx
|       ├── push-form.tsx
|       ├── device-select.tsx
|       ├── changelog-entry.tsx
|       ├── update-history-card.tsx
|       └── index.ts
|
└── routes/                         # PAGE LAYER (Routes)
    ├── updates-page.tsx           # MODIFIED - main layout
    ├── updates.status.tsx          # NEW - Status tab
    ├── updates.versions.tsx        # NEW - Versions tab
    ├── updates.push.tsx            # NEW - Push tab
    ├── updates.changelog.tsx      # NEW - Changelog tab
    ├── updates.export.tsx          # NEW - Export tab
    └── updates.history.tsx        # NEW - History tab
```

---

## 5. - 10. Tab Details (Status, Versions, Push, Changelog, Export, History)

### 5. Tab 1: Status
- Current/Latest version display
- GitHub sync status
- Sync Now action

### 6. Tab 2: Versions
- All available versions list
- Version cards with badges
- Sort/filter functionality

### 7. Tab 3: Push
- Version selector
- Device multi-select
- Install type (immediate/scheduled)
- Push button

### 8. Tab 4: Changelog
- Release notes per version
- Release type badges (major/minor/patch)
- Expandable notes

### 9. Tab 5: Export
- Format selection (JSON/CSV/PDF)
- Export fields checkboxes
- Download button

### 10. Tab 6: History
- Update push history
- Per-device status
- Cancel pending updates

---

## 11. Domain Layer

### 11.1 Files to CREATE

| File | Status | Purpose |
|------|--------|---------|
| `domain/shared/pagination.ts` | NEW | Pagination types |
| `domain/common/error.ts` | NEW | Domain errors |
| `domain/shared/types.ts` | NEW | Shared types |
| `domain/updates/update-types.ts` | NEW | Version, Changelog, UpdatePush types |
| `domain/updates/update-transforms.ts` | NEW | API transforms |
| `domain/updates/update-validation.ts` | NEW | Validation functions |

### 11.2 Types

```typescript
// domain/updates/update-types.ts
export interface Version {
  version: string;
  apkFilename: string;
  apkSize: number;
  sha256: string;
  releasedAt: string;
  releaseNotes?: string;
}

export enum UpdateStatus {
  PENDING = "pending",
  IN_PROGRESS = "in_progress",
  COMPLETED = "completed",
  FAILED = "failed",
  CANCELLED = "cancelled",
}

export enum InstallType {
  IMMEDIATE = "immediate",
  SCHEDULED = "scheduled",
}

export interface UpdatePush {
  id: string;
  version: string;
  deviceIds: string[];
  installType: InstallType;
  scheduledAt?: string;
  status: UpdateStatus;
  initiatedBy: string;
  initiatedAt: string;
}
```

---

## 12. Data Layer

### 12.1 Files to CREATE

| File | Status | Purpose |
|------|--------|---------|
| `lib/api/graphql/queries/update-queries.ts` | NEW | GET_VERSIONS, GET_CHANGELOG, etc. |
| `lib/api/graphql/mutations/update-mutations.ts` | NEW | PUSH_UPDATE, CANCEL_UPDATE |
| `lib/api/graphql/fragments/version.fragment.ts` | NEW | Version fragment |
| `lib/api/graphql/api-response-types.ts` | MODIFIED | Add update types |
| `lib/api/rest/updates.ts` | NEW | REST fallback endpoints |

---

## 13. Presentation Layer - Hooks

### 13.1 Hooks to CREATE

| File | Status | Purpose |
|------|--------|---------|
| `hooks/updates/use-versions.ts` | NEW | Fetch versions |
| `hooks/updates/use-changelog.ts` | NEW | Fetch changelog |
| `hooks/updates/use-push-update.ts` | NEW | Push update mutation |
| `hooks/updates/use-update-history.ts` | NEW | Fetch history |
| `hooks/updates/use-sync-status.ts` | NEW | Sync status |
| `hooks/updates/index.ts` | NEW | Barrel export |

---

## 14. UI Layer - Components

### 14.1 Routes (6 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| `routes/updates-page.tsx` | MODIFIED | Main layout |
| `routes/updates.status.tsx` | NEW | Status tab |
| `routes/updates.versions.tsx` | NEW | Versions tab |
| `routes/updates.push.tsx` | NEW | Push tab |
| `routes/updates.changelog.tsx` | NEW | Changelog tab |
| `routes/updates.export.tsx` | NEW | Export tab |
| `routes/updates.history.tsx` | NEW | History tab |

### 14.2 Components to CREATE

| File | Status | Purpose |
|------|--------|---------|
| `components/shared/section.tsx` | NEW | Bordered section |
| `components/shared/section-header.tsx` | NEW | Section header |
| `components/shared/empty-state.tsx` | NEW | Empty state |
| `components/shared/loading-skeleton.tsx` | NEW | Loading skeleton |
| `components/shared/data-table.tsx` | NEW | Table wrapper |
| `components/shared/pagination.tsx` | NEW | Pagination |
| `components/shared/status-badge.tsx` | NEW | Status badge |
| `components/shared/tab-nav.tsx` | NEW | Tab navigation |
| `components/shared/index.ts` | NEW | Barrel export |
| `components/updates/version-card.tsx` | NEW | Version card |
| `components/updates/version-badge.tsx` | NEW | Version badge |
| `components/updates/push-form.tsx` | NEW | Push form |
| `components/updates/device-select.tsx` | NEW | Device selector |
| `components/updates/changelog-entry.tsx` | NEW | Changelog entry |
| `components/updates/update-history-card.tsx` | NEW | History card |
| `components/updates/index.ts` | NEW | Barrel export |

---

## 15. File Changes Summary

### 15.1 Total File Count

| Category | New Files | Modified Files |
|----------|-----------|----------------|
| Domain Layer | 6 | 0 |
| Data Layer (GraphQL) | 3 | 1 |
| Data Layer (REST) | 1 | 0 |
| Presentation Layer | 6 | 0 |
| UI Layer (Shared) | 9 | 0 |
| UI Layer (Updates) | 7 | 0 |
| Routes | 6 | 1 |
| **TOTAL** | **38** | **2** |

### 15.2 All Files Listed

#### Domain Layer (6 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `domain/shared/pagination.ts` | NEW | Pagination types |
| `domain/common/error.ts` | NEW | Domain error types |
| `domain/shared/types.ts` | NEW | Shared domain types |
| `domain/updates/update-types.ts` | NEW | Version, Changelog, UpdatePush types |
| `domain/updates/update-transforms.ts` | NEW | versionFromAPI(), changelogFromAPI() |
| `domain/updates/update-validation.ts` | NEW | validateVersion(), validateUpdatePush() |

#### Data Layer - GraphQL (3 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| `lib/api/graphql/queries/update-queries.ts` | NEW | GET_VERSIONS, GET_CHANGELOG, etc. |
| `lib/api/graphql/mutations/update-mutations.ts` | NEW | PUSH_UPDATE, CANCEL_UPDATE |
| `lib/api/graphql/fragments/version.fragment.ts` | NEW | Version fragment |
| `lib/api/graphql/api-response-types.ts` | MODIFIED | Add update types |

#### Data Layer - REST (1 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `lib/api/rest/updates.ts` | NEW | REST endpoints for updates |

#### Presentation Layer - Hooks (6 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `hooks/updates/use-versions.ts` | NEW | Fetch all versions |
| `hooks/updates/use-changelog.ts` | NEW | Fetch changelog |
| `hooks/updates/use-push-update.ts` | NEW | Push update mutation |
| `hooks/updates/use-update-history.ts` | NEW | Fetch update history |
| `hooks/updates/use-sync-status.ts` | NEW | GitHub sync status |
| `hooks/updates/index.ts` | NEW | Barrel export |

#### UI Layer - Shared (9 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `components/shared/section.tsx` | NEW | Bordered section |
| `components/shared/section-header.tsx` | NEW | Section header |
| `components/shared/empty-state.tsx` | NEW | Empty state |
| `components/shared/loading-skeleton.tsx` | NEW | Loading skeleton |
| `components/shared/data-table.tsx` | NEW | Table wrapper |
| `components/shared/pagination.tsx` | NEW | Pagination controls |
| `components/shared/status-badge.tsx` | NEW | Status badge |
| `components/shared/tab-nav.tsx` | NEW | Tab navigation |
| `components/shared/index.ts` | NEW | Barrel export |

#### UI Layer - Updates (7 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `components/updates/version-card.tsx` | NEW | Version card |
| `components/updates/version-badge.tsx` | NEW | Version badge |
| `components/updates/push-form.tsx` | NEW | Push update form |
| `components/updates/device-select.tsx` | NEW | Device selector |
| `components/updates/changelog-entry.tsx` | NEW | Changelog entry |
| `components/updates/update-history-card.tsx` | NEW | Update history card |
| `components/updates/index.ts` | NEW | Barrel export |

#### Routes (6 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| `routes/updates-page.tsx` | MODIFIED | Main layout |
| `routes/updates.status.tsx` | NEW | Status tab |
| `routes/updates.versions.tsx` | NEW | Versions tab |
| `routes/updates.push.tsx` | NEW | Push tab |
| `routes/updates.changelog.tsx` | NEW | Changelog tab |
| `routes/updates.export.tsx` | NEW | Export tab |
| `routes/updates.history.tsx` | NEW | History tab |

---

## 16. Implementation Order

### Phase 1: Domain Layer (Day 1)
1. Create domain/shared types
2. Create domain/updates types
3. Create transforms and validation

### Phase 2: Data Layer (Day 1-2)
1. Create GraphQL queries and mutations
2. Add types to existing files
3. Create REST fallback

### Phase 3: Presentation Layer (Day 2)
1. Create hooks for each feature

### Phase 4: UI Layer - Shared (Day 2)
1. Create shared components

### Phase 5: UI Layer - Updates (Day 2-3)
1. Create updates feature components

### Phase 6: Routes (Day 3)
1. Create route files for each tab

### Phase 7: Polish (Day 3)
1. Test tab navigation
2. Add loading states
3. Add error handling

---

## 17. Testing Strategy

- Unit tests for domain transforms
- Integration tests for hooks
- E2E tests for each tab
- Visual regression for components

---

## 18. Rollout Checklist

- All tabs navigate correctly
- Forms submit properly
- Loading/error states work
- Mobile responsive
- Dark mode support

---

*Document Version: 1.1*
*Status: Ready for Implementation*
*Architecture: Layered (Following FRONTEND_ARCHITECTURE.md)*
