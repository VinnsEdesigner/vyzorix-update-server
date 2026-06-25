# Diagnostics Page - Enterprise Requirements Specification

> **Version:** 1.1
> **Status:** Draft
> **Created:** 2026-06-21
> **Updated:** 2026-06-24
> **Target:** Production MVP
> **Architecture:** Layered (Following `FRONTEND_ARCHITECTURE.md`)

---

## Table of Contents

1. [Overview](#1-overview)
2. [Page Structure](#2-page-structure)
3. [Architecture](#3-architecture)
4. [Target File Structure](#4-target-file-structure)
5. [Tab 1: Inspector](#5-tab-1-inspector)
6. [Tab 2: Timeline](#6-tab-2-timeline)
7. [Domain Layer](#7-domain-layer)
8. [Data Layer](#8-data-layer)
9. [Presentation Layer - Hooks](#9-presentation-layer---hooks)
10. [UI Layer - Components](#10-ui-layer---components)
11. [File Changes Summary](#11-file-changes-summary)
12. [Implementation Order](#12-implementation-order)
13. [Testing Strategy](#13-testing-strategy)
14. [Rollout Checklist](#14-rollout-checklist)

---

---

> ⚠️ **Architecture Alignment Note (v1.1)**
>
> This document has been updated to align with the **Layered Architecture** defined in `FRONTEND_ARCHITECTURE.md`. The file structure below follows the **4-layer architecture**:
> - **UI Layer** (`src/components/`) - Pure UI rendering, imports only from hooks
> - **Presentation Layer** (`src/hooks/`) - UI logic, state management, imports from domain & data
> - **Domain Layer** (`src/domain/` - NEW) - Types, transforms, validation (NO external imports)
> - **Data Layer** (`src/lib/api/`) - API clients (GraphQL/REST), imports only domain types
>
> **Dependency Rule:** UI → Hooks → Domain → API (flow inward only)

---

## 1. Overview

### 1.1 Purpose

The Diagnostics page provides operators with deep visibility into:
1. **Current State** - What the server knows about the device right now
2. **Audit Trail** - What events have occurred over time

### 1.2 Design Principles

- **No fake tests** - Only show real data from the server
- **No filler content** - Every data point is real and actionable
- **Two focused views** - Inspector for now, Timeline for history
- **Premium aesthetic** - Clean, dense, command-center feel

### 1.3 Relation to Other Pages

| Page | Responsibility |
|------|---------------|
| **Dashboard** | Overview + Metrics + Commands + Logs + Alerts (tabs) |
| **Device** | Inbox + Overview + Telemetry + Commands + History (tabs) |
| **Diagnostics** | Inspector + Timeline (tabs) - Deep dive only |
| **Updates** | Version status + Changelog + Push updates |
| **Settings** | Configuration |

---

## 2. Page Structure

### 2.1 Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│  DIAGNOSTICS                                              [Refresh] │
├─────────────────────────────────────────────────────────────────────┤
│  [Inspector]  [Timeline]                                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  TAB CONTENT                                                       │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 Navigation

- Two tabs: **Inspector** (default) and **Timeline**
- Refresh button in header
- Tab state persists in URL

---

## 3. Architecture

### 3.1 Layered Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        FRONTEND ARCHITECTURE                        │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                      UI LAYER                               │   │
│  │                   (src/components/)                        │   │
│  │                                                             │   │
│  │    Pages, Components, Shared UI                            │   │
│  │    ONLY renders UI. Uses hooks for everything.              │   │
│  │    NEVER imports from Data or Domain.                       │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              │ uses                                  │
│                              ▼                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                   PRESENTATION LAYER                          │   │
│  │                      (src/hooks/)                          │   │
│  │                                                             │   │
│  │    UI Logic, State Management, Data Transformation           │   │
│  │    Imports from Domain and Data layers.                      │   │
│  │    NEVER imports UI components.                              │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              │ uses                                  │
│                              ▼                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                      DOMAIN LAYER                           │   │
│  │                       (src/domain/)                         │   │
│  │                                                             │   │
│  │    Types, Transforms, Validation (Pure TypeScript)          │   │
│  │    NO external imports (no React, no API, no i18n)          │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              │                                       │
│                              │ uses                                  │
│                              ▼                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                       DATA LAYER                            │   │
│  │                    (src/lib/api/)                         │   │
│  │                                                             │   │
│  │    GraphQL Queries/Mutations, REST Endpoints               │   │
│  │    Imports Domain types only.                               │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.2 Dependency Rule

```
UI Layer ──────► Presentation Layer ──────► Domain Layer ──────► Data Layer
                  (Hooks)                    (Types)              (API)
                  
• UI Layer can ONLY import from Presentation Layer (hooks)
• Presentation Layer can import from Domain Layer and Data Layer  
• Domain Layer can NOT import from any other layer
• Data Layer can import from Domain Layer only
```

---

## 4. Target File Structure

```
apps/web/src/
│
├── domain/                          # DOMAIN LAYER (NEW)
│   ├── common/
│   │   ├── pagination.ts            # Pagination types & helpers
│   │   ├── error.ts                 # Domain error types
│   │   └── types.ts                 # Shared domain types
│   │
│   ├── diagnostics/
│   │   ├── diagnostic-types.ts       # DeviceInspection, TimelineEvent, EventType
│   │   ├── diagnostic-transforms.ts # inspectionFromRaw(), eventFromRaw()
│   │   └── diagnostic-validation.ts # validateInspection(), validateEvent()
│   │
│   └── devices/                     # Shared device domain types
│       ├── device-types.ts           # Device basic types
│       └── device-transforms.ts      # deviceBasicFromRaw()
│
├── lib/
│   └── api/
│       ├── graphql/
│       │   ├── client.ts            # (EXISTING)
│       │   ├── query-client.ts      # (EXISTING)
│       │   │
│       │   ├── queries/             # DATA LAYER - GraphQL Queries
│       │   │   ├── diagnostic-queries.ts   # GET_DEVICE_INSPECTION, GET_DEVICE_TIMELINE
│       │   │   └── index.ts
│       │   │
│       │   ├── mutations/            # DATA LAYER - GraphQL Mutations
│       │   │   └── index.ts
│       │   │
│       │   ├── fragments/           # GraphQL fragments
│       │   │   ├── inspection.fragment.ts
│       │   │   ├── timeline-event.fragment.ts
│       │   │   └── index.ts
│       │   │
│       │   ├── types.ts             # (EXISTING - add diagnostic types)
│       │   ├── index.ts
│       │   └── hooks.ts            # (EXISTING - add diagnostic hooks)
│       │
│       └── rest/                    # DATA LAYER - REST Fallback
│           ├── diagnostic-rest.ts       # REST endpoints for diagnostics
│           └── index.ts
│
├── hooks/                           # PRESENTATION LAYER
│   │
│   ├── diagnostics/                 # Diagnostics presentation logic
│   │   ├── use-device-inspection.ts # Fetch inspection data
│   │   ├── use-device-timeline.ts   # Fetch timeline events with pagination
│   │   ├── use-timeline-filter.ts  # Timeline filter state
│   │   └── index.ts                # Barrel export
│   │
│   ├── devices/                     # Device presentation logic (shared)
│   │   ├── use-devices.ts          # Device list with filters
│   │   ├── use-device-stream.ts    # WebSocket stream for real-time data
│   │   └── index.ts
│   │
│   └── shared/                     # Shared presentation utilities
│       ├── use-pagination.ts       # Generic pagination hook
│       ├── use-refresh.ts          # Refresh trigger hook
│       └── index.ts
│
├── components/                      # UI LAYER
│   │
│   ├── shared/                     # Shared UI components
│   │   ├── section.tsx             # Bordered section component
│   │   ├── section-header.tsx      # Section header with collapse toggle
│   │   ├── empty-state.tsx         # Empty state component
│   │   ├── loading-skeleton.tsx   # Loading skeleton variants
│   │   ├── data-table.tsx         # Table wrapper with sorting/pagination
│   │   ├── pagination.tsx         # Pagination controls
│   │   ├── status-badge.tsx       # Status badge component
│   │   ├── refresh-button.tsx      # Refresh button with loading state
│   │   ├── tab-nav.tsx            # Tab navigation component
│   │   └── index.ts               # Barrel export
│   │
│   ├── diagnostics/               # Diagnostics feature components
│   │   ├── diagnostics-page.tsx   # Page wrapper with tabs
│   │   ├── diagnostics-inspector.tsx # Inspector tab content
│   │   ├── diagnostics-timeline.tsx # Timeline tab content
│   │   ├── inspector-section.tsx  # Collapsible inspector section
│   │   ├── inspector-field.tsx    # Key-value display field
│   │   ├── inspector-identity.tsx # Identity section
│   │   ├── inspector-software.tsx # Software section
│   │   ├── inspector-registration.tsx # Registration section
│   │   ├── inspector-connection.tsx # Connection section
│   │   ├── inspector-telemetry.tsx # Telemetry stats section
│   │   ├── timeline-event.tsx     # Single event row
│   │   ├── timeline-filters.tsx   # Event type filter dropdown
│   │   ├── timeline-controls.tsx  # Auto-scroll, clear controls
│   │   └── index.ts              # Barrel export
│   │
│   ├── layout/                    # (EXISTING)
│   │   ├── app-layout.tsx
│   │   └── auth-layout.tsx
│   │
│   └── ui/                       # (EXISTING - shadcn/ui)
│       └── ...
│
└── routes/                         # PAGE LAYER (Routes)
    ├── diagnostics-page.tsx         # MODIFIED - layout with tabs
    ├── diagnostics.inspector.tsx   # NEW - Inspector tab
    └── diagnostics.timeline.tsx    # NEW - Timeline tab


---

## 11. File Changes Summary

### 11.1 Total File Count

| Category | New Files | Modified Files |
|----------|-----------|----------------|
| Domain Layer | 6 | 0 |
| Data Layer (GraphQL) | 3 | 1 |
| Data Layer (REST) | 2 | 0 |
| Presentation Layer | 4 | 0 |
| UI Layer (Shared) | 8 | 0 |
| UI Layer (Diagnostics) | 14 | 0 |
| Routes | 2 | 1 |
| **TOTAL** | **39** | **2** |

### 11.2 All Files Listed

#### Domain Layer (6 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `domain/shared/pagination.ts` | **NEW** | Pagination types & helpers |
| `domain/common/error.ts` | **NEW** | Domain error types |
| `domain/shared/types.ts` | **NEW** | Shared domain types |
| `domain/diagnostics/diagnostic-types.ts` | **NEW** | DeviceInspection, TimelineEvent, EventType |
| `domain/diagnostics/diagnostic-transforms.ts` | **NEW** | inspectionFromRaw(), eventFromRaw() |
| `domain/diagnostics/diagnostic-validation.ts` | **NEW** | validateInspection(), validateEvent() |

#### Data Layer - GraphQL (3 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| `lib/api/graphql/queries/diagnostics.ts` | **NEW** | GET_DEVICE_INSPECTION, GET_DEVICE_TIMELINE |
| `lib/api/graphql/fragments/inspection.fragment.ts` | **NEW** | Inspection GraphQL fragment |
| `lib/api/graphql/fragments/timeline-event.fragment.ts` | **NEW** | TimelineEvent fragment |
| `lib/api/graphql/api-response-types.ts` | MODIFIED | Add diagnostic types |
| `lib/api/graphql/hooks.ts` | MODIFIED | Add useDeviceInspection, useDeviceTimeline |

#### Data Layer - REST (2 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `lib/api/rest/diagnostics.ts` | **NEW** | REST endpoints for diagnostics |
| `lib/api/rest/index.ts` | **NEW** | REST barrel export |

#### Presentation Layer - Hooks (4 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `hooks/diagnostics/use-device-inspection.ts` | **NEW** | Fetch inspection data |
| `hooks/diagnostics/use-device-timeline.ts` | **NEW** | Fetch timeline with pagination |
| `hooks/diagnostics/use-timeline-filter.ts` | **NEW** | Timeline filter state |
| `hooks/diagnostics/index.ts` | **NEW** | Barrel export |

#### UI Layer - Shared (8 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `components/shared/section.tsx` | **NEW** | Bordered section component |
| `components/shared/section-header.tsx` | **NEW** | Section header with collapse |
| `components/shared/empty-state.tsx` | **NEW** | Empty state component |
| `components/shared/loading-skeleton.tsx` | **NEW** | Loading skeletons |
| `components/shared/refresh-button.tsx` | **NEW** | Refresh button with loading |
| `components/shared/tab-nav.tsx` | **NEW** | Tab navigation |
| `components/shared/pagination.tsx` | **NEW** | Pagination controls |
| `components/shared/index.ts` | **NEW** | Barrel export |

#### UI Layer - Diagnostics (14 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `components/diagnostics/diagnostics-page.tsx` | **NEW** | Page wrapper with tabs |
| `components/diagnostics/diagnostics-inspector.tsx` | **NEW** | Inspector tab content |
| `components/diagnostics/diagnostics-timeline.tsx` | **NEW** | Timeline tab content |
| `components/diagnostics/inspector-section.tsx` | **NEW** | Collapsible section |
| `components/diagnostics/inspector-field.tsx` | **NEW** | Key-value display |
| `components/diagnostics/inspector-identity.tsx` | **NEW** | Identity section |
| `components/diagnostics/inspector-software.tsx` | **NEW** | Software section |
| `components/diagnostics/inspector-registration.tsx` | **NEW** | Registration section |
| `components/diagnostics/inspector-connection.tsx` | **NEW** | Connection section |
| `components/diagnostics/inspector-telemetry.tsx` | **NEW** | Telemetry section |
| `components/diagnostics/timeline-event.tsx` | **NEW** | Single event row |
| `components/diagnostics/timeline-filters.tsx` | **NEW** | Filter controls |
| `components/diagnostics/timeline-controls.tsx` | **NEW** | Auto-scroll, clear |
| `components/diagnostics/index.ts` | **NEW** | Barrel export |

#### Routes (2 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| `routes/diagnostics-page.tsx` | **MODIFIED** | Main layout with tab navigation |
| `routes/diagnostics.inspector.tsx` | **NEW** | Inspector tab route |
| `routes/diagnostics.timeline.tsx` | **NEW** | Timeline tab route |

---

## 12. Implementation Order

### Phase 1: Domain Layer (Day 1)
1. Create `domain/shared/` types, errors, pagination
2. Create `domain/diagnostics/diagnostic-types.ts` with all diagnostic types
3. Create `domain/diagnostics/diagnostic-transforms.ts`
4. Create `domain/diagnostics/diagnostic-validation.ts`

### Phase 2: Data Layer (Day 1)
1. Create GraphQL queries in `lib/api/graphql/queries/diagnostics.ts`
2. Add types to `lib/api/graphql/api-response-types.ts`
3. Add hooks to `lib/api/graphql/hooks.ts`
4. Create REST fallback in `lib/api/rest/diagnostics.ts`

### Phase 3: Presentation Layer (Day 1-2)
1. Create `hooks/diagnostics/use-device-inspection.ts`
2. Create `hooks/diagnostics/use-device-timeline.ts`
3. Create `hooks/diagnostics/use-timeline-filter.ts`
4. Create shared hooks if needed

### Phase 4: UI Layer - Shared Components (Day 2)
1. Create `components/shared/section.tsx`
2. Create `components/shared/section-header.tsx`
3. Create `components/shared/tab-nav.tsx`
4. Create `components/shared/refresh-button.tsx`
5. Create `components/shared/empty-state.tsx`
6. Create `components/shared/loading-skeleton.tsx`
7. Create `components/shared/pagination.tsx`

### Phase 5: UI Layer - Diagnostics Components (Day 2-3)
1. Create `components/diagnostics/inspector-section.tsx`
2. Create `components/diagnostics/inspector-field.tsx`
3. Create inspector sub-components (identity, software, etc.)
4. Create `components/diagnostics/timeline-event.tsx`
5. Create `components/diagnostics/timeline-filters.tsx`
6. Create `components/diagnostics/timeline-controls.tsx`
7. Create `components/diagnostics/diagnostics-inspector.tsx`
8. Create `components/diagnostics/diagnostics-timeline.tsx`
9. Create `components/diagnostics/diagnostics-page.tsx`

### Phase 6: Route Assembly (Day 3)
1. Update `routes/diagnostics-page.tsx` to use new components
2. Wire up hooks to components
3. Add loading/error states
4. Add refresh functionality

### Phase 7: Polish (Day 3)
1. Add timeline auto-scroll
2. Add event filtering UI
3. Add load more pagination UI
4. Test collapse/expand animations
5. Test empty states

---

## 13. Testing Strategy

### Unit Tests
- Domain transforms (`domain/diagnostics/transforms.test.ts`)
- Domain validation (`domain/diagnostics/validation.test.ts`)
- Hook state management

### Integration Tests
- GraphQL query hooks with mock server
- REST fallback with mock server
- Pagination cursor encoding/decoding

### E2E Tests
- Navigate to Diagnostics page
- View Inspector tab with device data
- View Timeline tab with events
- Filter events by type
- Collapse/expand sections
- Refresh data
- Load more timeline events

### Visual Regression
- Inspector sections layout
- Timeline event cards
- Loading skeletons
- Empty states

---

## 14. Rollout Checklist

### Pre-Launch
- [ ] All new components have loading states
- [ ] All new components have error states
- [ ] All new components have empty states
- [ ] Pagination works correctly
- [ ] Refresh button works
- [ ] Tab navigation persists in URL
- [ ] Mobile responsive layout tested
- [ ] Dark mode tested
- [ ] Performance tested (large timelines)

### Post-Launch
- [ ] Monitor for inspection query errors
- [ ] Monitor for timeline query errors
- [ ] Check for memory leaks in timeline (auto-scroll)
- [ ] Gather user feedback on UX

---

*Document Version: 1.1*
*Status: Ready for Implementation*
*Architecture: Layered (Following FRONTEND_ARCHITECTURE.md)*
