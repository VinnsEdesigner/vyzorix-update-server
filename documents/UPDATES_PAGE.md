# Updates Page - Enterprise Requirements Specification

> **Version:** 2.0
> **Status:** In Implementation
> **Created:** 2026-06-22
> **Updated:** 2026-08-15
> **Target:** Production MVP
> **Architecture:** Layered (Following `FRONTEND_ARCHITECTURE.md`)

---

## Table of Contents

1. [Overview](#1-overview)
2. [Implementation Status](#2-implementation-status)
3. [Architecture](#3-architecture)
4. [Actual File Structure](#4-actual-file-structure)
5. [Tab 1: Status](#5-tab-1-status)
6. [Tab 2: Versions](#6-tab-2-versions)
7. [Tab 3: Push](#7-tab-3-push)
8. [Tab 4: Changelog](#8-tab-4-changelog)
9. [Tab 5: Export](#9-tab-5-export)
10. [Tab 6: History](#10-tab-6-history)
11. [Domain Layer (DONE)](#11-domain-layer-done)
12. [Data Layer (DONE)](#12-data-layer-done)
13. [Presentation Layer - Hooks (TODO)](#13-presentation-layer---hooks-todo)
14. [UI Layer - Components (TODO)](#14-ui-layer---components-todo)
15. [Remaining Work Summary](#15-remaining-work-summary)
16. [Implementation Order](#16-implementation-order)
17. [Testing Strategy](#17-testing-strategy)
18. [Rollout Checklist](#18-rollout-checklist)

---

> **v2.0 Re-alignment Note**
>
> This document was rewritten to match the **actual** state of the codebase. The
> prior v1.1 draft described a speculative flat layout (`apps/web/src/domain/`,
> `lib/api/`) that does not exist. The real layered architecture lives in:
> - **Domain Layer** — `packages/API_Client/src/domain/updates/` (types + REST mappers)
> - **Data Layer** — `packages/API_Client/src/vyzorServer/graphql/updates/` (GraphQL)
>   and `packages/API_Client/src/vyzorServer/rest/updates/` (REST)
> - **Presentation Layer** — `apps/VyzoriX_web/src/hooks/updates/` (currently empty)
> - **UI Layer** — `apps/VyzoriX_web/src/ui/pages/updates/` (currently empty)
>
> **Organization model:** organization scoping is driven by
> `useAuthStore.organizationId` (see `hooks/_shared/use-current-context.ts`),
> not by a global function call. REST endpoints read it via
> `getOrganizationContext()`; GraphQL is scoped via the org-scoped endpoint URL
> `/v1/orgs/:orgId/graphql` set by `graphqlClient.setOrganization`.
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
GitHub Repository -> Update Server (apps/api) -> API Client (@vyzorix/api-client) -> Frontend (Updates Page)
```

The Update Server exposes both a REST API (`/v1/updates/*`) and an org-scoped
GraphQL endpoint (`/v1/orgs/:orgId/graphql`). The `@vyzorix/api-client` package
is the single source of truth for typed access to both.

### 1.3 Page Structure

```
/updates
 Status      -> Current state, sync status
 Versions    -> All available versions
 Push        -> Push updates to devices
 Changelog   -> Release notes
 Export      -> Export version data
 History     -> Past updates
```

---

## 2. Implementation Status

| Layer | Location | Status |
|-------|----------|--------|
| **Domain** | `packages/API_Client/src/domain/updates/` | ✅ DONE |
| **Data — GraphQL** | `packages/API_Client/src/vyzorServer/graphql/updates/` | ✅ DONE |
| **Data — REST** | `packages/API_Client/src/vyzorServer/rest/updates/` | ✅ DONE |
| **Presentation — Hooks** | `apps/VyzoriX_web/src/hooks/updates/` | ⬜ TODO (only `.gitkeep`) |
| **State — Stores** | `apps/VyzoriX_web/src/stores/` | ⬜ TODO (no updates store) |
| **UI — Components** | `apps/VyzoriX_web/src/ui/pages/updates/` | ⬜ TODO (only `.gitkeep`) |
| **UI — Routes** | `apps/VyzoriX_web/src/routes/` | ⬜ TODO (no updates routes) |

The API client layer (domain + data) is complete and type-checked. The
remaining work is in the web app: hooks, an optional updates store, UI
components, and route files.

### 2.1 Routes (TanStack Router File-Based Routing)

| Route | File | Purpose |
|-------|------|---------|
| `/updates` | `routes/updates.tsx` (or `updates-page.tsx`) | Main layout / tab shell |
| `/updates/status` | `routes/updates.status.tsx` | Status tab |
| `/updates/versions` | `routes/updates.versions.tsx` | Versions tab |
| `/updates/push` | `routes/updates.push.tsx` | Push tab |
| `/updates/changelog` | `routes/updates.changelog.tsx` | Changelog tab |
| `/updates/export` | `routes/updates.export.tsx` | Export tab |
| `/updates/history` | `routes/updates.history.tsx` | History tab |

**Total: 7 NEW routes** (the `__root.tsx` already wires providers; no modification needed).

---

## 3. Architecture

### 3.1 Layered Architecture Overview

```
UI Layer (apps/VyzoriX_web/src/ui/)
  -> Presentation Layer (apps/VyzoriX_web/src/hooks/)
  -> Domain Layer (packages/API_Client/src/domain/)
  -> Data Layer (packages/API_Client/src/vyzorServer/)
```

### 3.2 Dependency Rule

- UI Layer can ONLY import from Presentation Layer (hooks)
- Presentation Layer can import from Domain Layer and Data Layer (`@vyzorix/api-client`)
- Domain Layer can NOT import from any other layer (pure types + transforms)
- Data Layer can import from Domain Layer only

### 3.3 Organization Scoping

All updates queries/mutations are organization-scoped. The current organization
ID is read from the auth store:

```ts
// apps/VyzoriX_web/src/hooks/_shared/use-current-context.ts
export function useCurrentOrganizationId(): string | null {
  return useAuthStore((s) => s.organizationId);
}
export function useRequiredOrganizationId(): string { ... }
```

- **REST** — `getOrganizationContext()` reads the current org and sends it as the
  `organization_id` query param (see `updates-endpoints.ts`).
- **GraphQL** — the org is encoded in the endpoint URL
  (`/v1/orgs/:orgId/graphql`); `graphqlClient.setOrganization(orgId)` selects it.
  Most operations also take an explicit `$organizationId` variable. The
  `syncUpdates` mutation is the one exception: the server defines it with no
  arguments, so it is scoped solely via the endpoint URL.

---

## 4. Actual File Structure

```
packages/API_Client/src/                     # @vyzorix/api-client (DONE)
├── domain/updates/
│   ├── updates-entity.ts        # UpdateVersion, UpdatePush, PushUpdateRequest, ChangelogEntry, ...
│   ├── updates-mappers.ts       # versionFromRaw, updatePushFromRaw, ... (REST raw -> domain)
│   └── index.ts                 # barrel
└── vyzorServer/
    ├── graphql/updates/
    │   ├── graphql-updates-fragments.ts   # UPDATE_VERSION_FRAGMENT, PUSH_HISTORY_ENTRY_FRAGMENT, ...
    │   ├── graphql-updates-queries.ts     # GET_UPDATES, GET_UPDATES_STATUS, GET_UPDATES_CHANGELOG, GET_UPDATES_HISTORY + typed wrappers
    │   ├── graphql-updates-mutations.ts   # PUSH_UPDATE, CANCEL_UPDATE, SYNC_UPDATES + typed wrappers
    │   ├── graphql-updates-mappers.ts     # updateVersionFromRaw, pushHistoryEntryFromRaw, pushUpdateResponseFromRaw, ...
    │   ├── graphql-updates-types.ts       # RawUpdateVersion, RawPushHistoryEntry, RawPushUpdateResponse, ...
    │   └── index.ts                       # barrel (re-exported via graphql/index.ts)
    └── rest/updates/
        └── updates-endpoints.ts           # updates.getStatus / getVersions / pushUpdate / sync / ...

apps/VyzoriX_web/src/                        # web app (TODO)
├── hooks/updates/                           # PRESENTATION LAYER  (only .gitkeep today)
│   ├── use-versions.ts                      # TODO
│   ├── use-update-status.ts                 # TODO
│   ├── use-changelog.ts                     # TODO
│   ├── use-push-update.ts                   # TODO
│   ├── use-update-history.ts                # TODO
│   ├── use-sync-status.ts                   # TODO
│   └── index.ts                             # TODO (barrel)
├── stores/                                  # no updates store yet (TODO if needed)
├── ui/pages/updates/                        # UI LAYER (only .gitkeep today)
│   ├── version-card.tsx                     # TODO
│   ├── push-form.tsx                        # TODO
│   ├── changelog-entry.tsx                  # TODO
│   ├── update-history-card.tsx              # TODO
│   └── ...                                  # TODO
└── routes/                                  # PAGE LAYER (no updates routes yet)
    ├── updates.tsx                          # TODO — main layout / tab shell
    ├── updates.status.tsx                   # TODO
    ├── updates.versions.tsx                 # TODO
    ├── updates.push.tsx                     # TODO
    ├── updates.changelog.tsx                # TODO
    ├── updates.export.tsx                   # TODO
    └── updates.history.tsx                  # TODO
```

### 4.1 Field Naming Contract (server ↔ client)

These names are pinned to the Go server contract in
`apps/api/internal/api/graphql/schema/objects.go` and
`apps/api/internal/application/updates/updates_command_responses.go`. They must
not drift:

| Server field | Domain entity field | Notes |
|--------------|---------------------|-------|
| `releasedAt` | `releaseDate: Date` | GraphQL/REST return `releasedAt`; mappers translate to the domain `releaseDate`. The `UpdateVersion` GraphQL type does **not** expose `updatedAt`, so the mapper falls back to `new Date()`. |
| `version` (string, e.g. `v1.2.0`) | `version: string` | Used by `pushUpdate`, `UpdatePush`, `PushUpdateRequest`. NOT a UUID. The old `versionId`/`version_id` naming was removed. |
| `organizationId` | `organizationId` | Explicit variable on every operation except `syncUpdates` (server takes no args; scoped via endpoint URL). |
| `deviceCount` / `pending` / `acknowledged` / `failed` | `PushDevices` | History entries return aggregate counts, not a `devices[]` array. `pushHistoryEntryFromRaw` reads them directly. |
| `devices[]` (PushDevice) | derived `PushDevices` counts | Only the push **detail** endpoint (`updatesHistoryDetail`) returns a `devices[]` array. |

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

## 11. Domain Layer (DONE)

The domain layer is implemented in `packages/API_Client/src/domain/updates/`.

### 11.1 Files (existing)

| File | Status | Purpose |
|------|--------|---------|
| `domain/updates/updates-entity.ts` | ✅ DONE | `UpdateVersion`, `UpdatePush`, `PushDevices`, `PushUpdateRequest`, `ChangelogEntry`, `VersionListResult`, `UpdateHistoryResult`, enums, helpers (`formatApkSize`, `getReleaseTypeLabel`, ...) |
| `domain/updates/updates-mappers.ts` | ✅ DONE | `versionFromRaw`, `syncStateFromRaw`, `updatePushFromRaw`, `versionListResultFromRaw`, `updateHistoryResultFromRaw`, `changelogEntryFromRaw` + `Raw*` REST shapes |
| `domain/updates/index.ts` | ✅ DONE | Barrel export |
| `domain/_shared/domain-pagination.ts` | ✅ DONE | `Pagination`, `RawPagination`, `paginationFromRaw` (shared) |

### 11.2 Canonical Types

```typescript
// packages/API_Client/src/domain/updates/updates-entity.ts
export type ReleaseType = "major" | "minor" | "patch";
export type UpdateStatus = "pending" | "in_progress" | "completed" | "failed" | "cancelled";
export type InstallType = "immediate" | "scheduled";
export type SyncStatus = "idle" | "syncing" | "synced" | "error";

export interface UpdateVersion {
  id: string;
  version: string;
  apkFilename: string;
  apkSize: number;
  sha256: string;
  releaseType: ReleaseType;
  releaseNotes?: string;
  releaseDate: Date;          // mapped from server `releasedAt`
  isLatest: boolean;
  createdAt: Date;
  updatedAt: Date;
}

export interface PushDevices {
  total: number;
  pending: number;
  sent: number;
  acknowledged: number;
  failed: number;
}

export interface UpdatePush {
  id: string;
  version: string;            // version STRING (e.g. v1.2.0), not a UUID
  installType: InstallType;
  status: UpdateStatus;
  initiatedBy: string;
  initiatedAt: Date;
  scheduledAt?: Date;
  completedAt?: Date;
  cancelledAt?: Date;
  cancelledBy?: string;
  devices: PushDevices;
}

export interface PushUpdateRequest {
  version: string;            // version STRING
  deviceIds: string[];
  installType: InstallType;
  scheduledAt?: Date;
}
```

---

## 12. Data Layer (DONE)

The data layer is implemented in `packages/API_Client/src/vyzorServer/`. It
exposes both a GraphQL and a REST surface; hooks may prefer REST (typed,
eagerly mapped) and fall back to GraphQL (see the pattern in
`hooks/commands/_graphql-fallback.ts`).

### 12.1 GraphQL (existing)

| File | Status | Purpose |
|------|--------|---------|
| `vyzorServer/graphql/updates/graphql-updates-fragments.ts` | ✅ DONE | `UPDATE_VERSION_FRAGMENT`, `PUSH_DEVICE_FRAGMENT`, `UPDATE_PUSH_FRAGMENT`, `SYNC_STATUS_FRAGMENT`, `CHANGELOG_ENTRY_FRAGMENT`, `PUSH_HISTORY_ENTRY_FRAGMENT` |
| `vyzorServer/graphql/updates/graphql-updates-queries.ts` | ✅ DONE | `GET_UPDATES`, `GET_UPDATES_STATUS`, `GET_UPDATES_CHANGELOG`, `GET_UPDATES_HISTORY` + typed wrappers `queryUpdates`, `queryUpdatesStatus`, `queryUpdatesChangelog`, `queryUpdatesHistory` |
| `vyzorServer/graphql/updates/graphql-updates-mutations.ts` | ✅ DONE | `PUSH_UPDATE`, `CANCEL_UPDATE`, `SYNC_UPDATES` + typed wrappers `mutatePushUpdate`, `mutateCancelUpdate`, `mutateSyncUpdates` |
| `vyzorServer/graphql/updates/graphql-updates-mappers.ts` | ✅ DONE | `updateVersionFromRaw`, `updatePushFromRaw`, `pushHistoryEntryFromRaw`, `updateStatusFromRaw`, `pushUpdateResponseFromRaw`, `cancelUpdateResponseFromRaw`, `syncResponseFromRaw`, ... |
| `vyzorServer/graphql/updates/graphql-updates-types.ts` | ✅ DONE | `RawUpdateVersion`, `RawUpdatePush`, `RawPushHistoryEntry`, `RawPushUpdateResponse`, `RawCancelUpdateResponse`, `RawSyncResponse`, `RawUpdateStatusResponse`, ... |
| `vyzorServer/graphql/updates/index.ts` | ✅ DONE | Barrel; re-exported from `graphql/index.ts` via `export * from "./updates"` |

GraphQL fragments are inlined into queries/mutations via `${FRAGMENT}` so
Apollo can resolve `...Fragment` spreads. All wrappers return typed domain
objects (mappers applied), not `unknown`.

### 12.2 REST (existing)

| File | Status | Purpose |
|------|--------|---------|
| `vyzorServer/rest/updates/updates-endpoints.ts` | ✅ DONE | `updates.getStatus`, `getVersions`, `getChangelog`, `getHistory`, `getPushDetail`, `pushUpdate`, `cancelPush`, `sync`, `exportVersions` |

REST API endpoints (server contract):

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/updates/status` | GET | Get update system status |
| `/v1/updates/versions` | GET | Get all available versions |
| `/v1/updates/changelog` | GET | Get release changelog |
| `/v1/updates/push` | POST | Push update to devices (body: `version`, `deviceIds`, `installType`, `scheduledAt`) |
| `/v1/updates/history` | GET | Get update push history |
| `/v1/updates/history/:pushId` | GET | Get push detail |
| `/v1/updates/history/:pushId/cancel` | POST | Cancel pending update |
| `/v1/updates/export` | GET | Export version data (`format=json\|csv`) |
| `/v1/updates/sync` | POST | Sync versions from GitHub |

### 12.3 Usage Example (for hook authors)

```typescript
import { updates, type UpdatePush, type PushUpdateRequest } from '@vyzorix/api-client';

// fetch
const { versions, pagination } = await updates.getVersions({ status: 'latest', page: 1, limit: 20 }, organizationId);

// mutate
const request: PushUpdateRequest = { version: 'v1.2.0', deviceIds: ['dev-1'], installType: 'immediate' };
const pushed: UpdatePush = await updates.pushUpdate(request, organizationId);
```

---

## 13. Presentation Layer - Hooks (TODO)

No hooks exist yet (`hooks/updates/` contains only `.gitkeep`). They should
follow the established pattern in `hooks/commands/use-commands.ts`: use
TanStack Query, read the org via `useCurrentOrganizationId()`, prefer the REST
surface from `@vyzorix/api-client`, and optionally fall back to GraphQL.

### 13.1 Hooks to CREATE

| File | Status | Purpose |
|------|--------|---------|
| `hooks/updates/use-versions.ts` | TODO | Fetch versions (`updates.getVersions`) |
| `hooks/updates/use-update-status.ts` | TODO | Fetch status (`updates.getStatus`) |
| `hooks/updates/use-changelog.ts` | TODO | Fetch changelog (`updates.getChangelog`) |
| `hooks/updates/use-push-update.ts` | TODO | Push update mutation (`updates.pushUpdate`) |
| `hooks/updates/use-update-history.ts` | TODO | Fetch history (`updates.getHistory`) + detail |
| `hooks/updates/use-sync-status.ts` | TODO | Sync status + `mutateSyncUpdates`/`updates.sync` trigger |
| `hooks/updates/use-cancel-update.ts` | TODO | Cancel pending push (`updates.cancelPush`) |
| `hooks/updates/_graphql-fallback.ts` | TODO (optional) | GraphQL fallback wrappers |
| `hooks/updates/index.ts` | TODO | Barrel export |

Query keys should be added to `lib/query-keys.ts` (e.g. `updatesStatus`, `updateVersions`, `updateHistory`, `updateChangelog`). Register the barrel in `hooks/index.ts`.

### 13.2 Example Hook Skeleton

```typescript
// hooks/updates/use-versions.ts
import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import { updates, type VersionListResult } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export function useVersions(
  params?: { status?: 'all' | 'latest' | 'previous'; page?: number; limit?: number },
  options?: Omit<UseQueryOptions<VersionListResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.updateVersions(params ?? {}),
    queryFn: () => updates.getVersions(params ?? undefined, organizationId ?? undefined),
    enabled: organizationId !== null,
    ...options,
  });
}
```

---

## 14. UI Layer - Components (TODO)

No components or routes exist yet (`ui/pages/updates/` contains only `.gitkeep`).
Reuse the existing shared component library under `ui/components/shared/`
(`section`, `export-menu`, `device-selector`, etc.) rather than recreating
generic primitives.

### 14.1 Routes to CREATE (7 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `routes/updates.tsx` | TODO | Main layout / tab shell |
| `routes/updates.status.tsx` | TODO | Status tab |
| `routes/updates.versions.tsx` | TODO | Versions tab |
| `routes/updates.push.tsx` | TODO | Push tab |
| `routes/updates.changelog.tsx` | TODO | Changelog tab |
| `routes/updates.export.tsx` | TODO | Export tab |
| `routes/updates.history.tsx` | TODO | History tab |

### 14.2 Components to CREATE (in `ui/pages/updates/`)

| File | Status | Purpose |
|------|--------|---------|
| `version-card.tsx` | TODO | Version card with badge |
| `push-form.tsx` | TODO | Push update form (version + device multi-select + install type) |
| `changelog-entry.tsx` | TODO | Changelog entry (expandable) |
| `update-history-card.tsx` | TODO | History card with per-device counts + cancel |
| `index.ts` | TODO | Barrel export |

---

## 15. Remaining Work Summary

| Category | Done | Remaining |
|----------|------|-----------|
| Domain Layer | 3 files | 0 |
| Data Layer — GraphQL | 6 files | 0 |
| Data Layer — REST | 1 file | 0 |
| Presentation — Hooks | 0 | ~7 files + query keys |
| State — Stores | 0 | optional `updates-store.ts` (sync status, push draft) |
| UI — Components | 0 | ~5 feature components |
| UI — Routes | 0 | 7 route files |

**Net remaining:** ~19 web-app files (hooks, optional store, components, routes).
The API client contract is frozen and type-checked; hook authors can consume
`@vyzorix/api-client` directly.

---

## 16. Implementation Order

Phases 1–2 (Domain + Data) are **complete**. Remaining work starts at Phase 3.

### Phase 1: Domain Layer ✅ DONE
- `packages/API_Client/src/domain/updates/` (entity, mappers, barrel)

### Phase 2: Data Layer ✅ DONE
- `packages/API_Client/src/vyzorServer/graphql/updates/` (fragments, queries, mutations, mappers, types)
- `packages/API_Client/src/vyzorServer/rest/updates/updates-endpoints.ts`

### Phase 3: Presentation Layer (next)
1. Add update query keys to `lib/query-keys.ts`
2. Create `hooks/updates/*` hooks (follow `hooks/commands/use-commands.ts` pattern)
3. Create `hooks/updates/index.ts` barrel and register in `hooks/index.ts`

### Phase 4: State Layer (optional)
1. If realtime sync status is needed, add `stores/updates-store.ts` (zustand) and register in `stores/index.ts`

### Phase 5: UI Layer - Updates
1. Create `ui/pages/updates/*` feature components (reuse `ui/components/shared/*`)

### Phase 6: Routes
1. Create `routes/updates.tsx` tab shell + per-tab route files

### Phase 7: Polish
1. Tab navigation, loading/error states, mobile responsive, dark mode

---

## 17. Testing Strategy

- Unit tests for GraphQL/REST mappers (`pushHistoryEntryFromRaw`, `updateVersionFromRaw`, ...) — verify `releasedAt` → `releaseDate`, aggregate counts, and missing-`devices` safety
- Integration tests for hooks (TanStack Query)
- E2E tests for each tab
- Visual regression for components
- Run `pnpm --filter @vyzorix/api-client typecheck` to guard the data-layer contract

---

## 18. Rollout Checklist

- All tabs navigate correctly
- Push form submits `version` (string) + `deviceIds` + `installType`
- History renders aggregate device counts (no crash on missing `devices[]`)
- Sync Now triggers `syncUpdates` (org-scoped via endpoint URL)
- Loading/error states work
- Mobile responsive
- Dark mode support

---

*Document Version: 2.0*
*Status: API client DONE; hooks/stores/UI TODO*
*Architecture: Layered (Following FRONTEND_ARCHITECTURE.md)*
