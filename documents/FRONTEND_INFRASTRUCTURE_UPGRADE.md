# Frontend Infrastructure Upgrade — Design Specification

> **Version:** 1.0
> **Status:** Approved — implementation in progress
> **Created:** 2026-08-15
> **Target:** `apps/VyzoriX_web` (web app only; no API Client changes)
> **Reference baseline:** OpenHands / Agent Canvas frontend (`VinnsEdesigner/OpenHands`, branch `main`)

---

## Table of Contents

1. [Overview & Motivation](#1-overview--motivation)
2. [Architecture Constraints](#2-architecture-constraints)
3. [Feature 1 — Internationalization (i18n)](#3-feature-1--internationalization-i18n)
4. [Feature 2 — Telemetry / Product Analytics](#4-feature-2--telemetry--product-analytics)
5. [Feature 3 — E2E / Integration Testing Infrastructure](#5-feature-3--e2e--integration-testing-infrastructure)
6. [Feature 4 — Error Boundary + Error Store](#6-feature-4--error-boundary--error-store)
7. [Feature 5 — Zustand Devtools Middleware (Factory)](#7-feature-5--zustand-devtools-middleware-factory)
8. [Implementation Order](#8-implementation-order)
9. [Verification Gates](#9-verification-gates)
10. [File Inventory (Complete)](#10-file-inventory-complete)

---

## 1. Overview & Motivation

A cross-cutting comparison of the OpenHands / Agent Canvas frontend against the Vyzorix
web app identified five infrastructure features OpenHands has that Vyzorix lacks. This
document specifies how Vyzorix will adopt each — not by copying OpenHands, but by
improving on its design.

Each feature is designed to be **structurally better** than the OpenHands equivalent:
typed where OpenHands is stringly-typed, provider-agnostic where OpenHands is
vendor-locked, factory-driven where OpenHands is boilerplate, and path-aligned to the
real API contract where OpenHands mocks drift.

### Design principles

- **Explicitly unique file names** — no generic `index.ts` / `config.ts` collisions.
  Every file is prefixed with `vyzor-` (infra) or named for its single responsibility.
- **Layered architecture compliance** — every file respects the enforced
  `UI → Hooks → {API_Client, Lib}` dependency flow (see §2).
- **No comments in implementation code** — per project convention; this doc is the
  design record, code is self-documenting.
- **No bugs** — each phase passes `tsc --noEmit` (0 errors), `eslint` (0 errors),
  and `vitest run` (all green) before moving on.

---

## 2. Architecture Constraints

The Vyzorix web app enforces a layered architecture via a custom ESLint plugin
(`@vyzorix/config` → `vyzo/layer-imports`). The rules are:

```
UI Layer (src/ui, src/routes)
  ↓ only
Hooks Layer (src/hooks)
  ↓ only
{ API_Client (domain, vyzorServer), Lib (src/lib) }
```

| Layer | Path | Can import from | Cannot import from |
|---|---|---|---|
| `ui` | `src/ui/**` | hooks, ui, react | domain, vyzorServer |
| `hooks` | `src/hooks/**` | domain, vyzorServer, hooks, react | ui |
| `domain` | `packages/API_Client/src/domain/**` | domain | hooks, ui, vyzorServer, react |
| `vyzorServer` | `packages/API_Client/src/vyzorServer/**` | domain, vyzorServer, react | hooks, ui |

**Unclassified layers** (not restricted by the rule — can import freely):
- `src/lib/**` — utilities, infra (importable by hooks; not UI-restricted)
- `src/providers/**` — React context providers
- `src/stores/**` — Zustand stores
- `src/routes/**` — TanStack Router route files
- `src/test/**` — test infrastructure (outside lint scope)

All five features place their non-React infra in `src/lib/`, their React glue in
`src/providers/`, their reactive access in `src/hooks/`, and their visual surfaces in
`src/ui/` — fully compliant with the layer rules.

### Other relevant facts

- **Router:** TanStack Router (not React Router) → error boundaries use
  `errorComponent`, not `errorElement`.
- **State:** Zustand v5 → `devtools` + `persist` ship in-box, no new dep.
- **Path alias:** `@/*` → `src/*`.
- **Test runner:** Vitest + jsdom + Testing Library; setup in `src/test/setup.ts`;
  render-hook helper in `src/test/helpers/render-hook.tsx`.
- **Vitest config:** `vite-tsconfig-paths`, `restoreMocks: true`, `mockReset: true`.
- **Query client:** `src/lib/query-client.ts` — retry:false for 4xx, staleTime 30s.
- **`cmdk`** already a dependency (for future command palette — not in this batch).
- **Existing stores (13):** auth, command-dispatch, command-queue, connectivity,
  dashboard, device-selector, diagnostics, log-stream, metrics-realtime, theme,
  timeline-stream, updates, websocket.

---

## 3. Feature 1 — Internationalization (i18n)

### 3.1 How OpenHands does it

- One 35,920-key flat `translation.json` (key → `{ lang: value }` map).
- Manual `I18nKey` type declarations that **drift** from the JSON.
- `i18next-http-backend` → runtime HTTP fetch for every locale (waterfall on first load).
- A Proxy-based default instance (clever but hard to debug).
- Hardcoded to a single `openhands` namespace.

### 3.2 Weaknesses

- Type declarations are hand-maintained → keys can exist in JSON but not in the type,
  or vice versa.
- Runtime HTTP fetch for locales adds latency and a failure mode (locale file 404).
- Flat key space (35K keys in one file) is unmanageable.
- Proxy indirection makes debugging initialization issues hard.

### 3.3 How Vyzorix does it better

1. **Type-safe keys derived from source** —
   `type VyzorTranslationKey = keyof typeof enTranslations`. Add a key to the TS file
   → it is instantly typed. No manual declaration file, zero drift.
2. **TS translation files (not JSON)** — enables i18next's typed interpolation
   params (`t('updates.pushed', { count: 3 })` is type-checked at compile time).
3. **Code-split locales via dynamic `import()`** — no runtime HTTP fetch waterfall;
   Vite bundles each locale as a lazy chunk loaded on demand.
4. **Feature namespaces** (`common`, `updates`, `device`) loaded on demand → small
   initial bundle; a feature's strings load when the feature renders.
5. **Persisted language store** — survives reloads, devtools-visible (via Feature 5),
  testable. OpenHands uses opaque `i18next-browser-languagedetector`.

### 3.4 Dependencies

| Package | Type | Purpose |
|---|---|---|
| `i18next` | dependency | i18n core |
| `react-i18next` | dependency | React bindings (`useTranslation`, `I18nextProvider`) |

### 3.5 File layout

```
src/lib/i18n/
├── vyzor-i18n-config.ts              # i18next init config, supported languages, fallback
├── vyzor-i18n-types.ts               # VyzorTranslationKey = keyof typeof en; VyzorNamespace type
├── vyzor-i18n-loader.ts              # async locale loader: import('./locales/fr/...')
├── locales/
│   ├── en/
│   │   ├── vyzor-common-en.ts        # common.{ok,cancel,error,...}
│   │   └── vyzor-updates-en.ts       # updates.{pushed,version,...}
│   └── fr/
│       ├── vyzor-common-fr.ts
│       └── vyzor-updates-fr.ts
└── index.ts                          # barrel: re-export config + types

src/providers/
└── vyzor-i18n-provider.tsx           # wraps i18next init + I18nextProvider, awaits ready

src/hooks/i18n/
├── use-vyzor-translation.ts          # useVyzorTranslation() → { t, i18n, ready }
├── use-vyzor-language.ts             # useVyzorLanguage() → { locale, setLocale, available }
└── index.ts

src/stores/
└── vyzor-i18n-store.ts               # persisted locale preference (zustand + persist)
```

### 3.6 Layer fit

| File | Layer | Imports from |
|---|---|---|
| `lib/i18n/*` | lib | i18next, react-i18next (external) |
| `providers/vyzor-i18n-provider.tsx` | providers | lib/i18n, react-i18next |
| `hooks/i18n/*` | hooks | lib/i18n, react-i18next, stores |
| `stores/vyzor-i18n-store.ts` | stores | lib/state (Feature 5 factory) |
| `ui/**` (future) | ui | hooks/i18n only ✓ |

### 3.7 Integration points

- `routes/__root.tsx` — wrap `QueryProvider` output in `<VyzorI18nProvider>`.
- All future UI components use `useVyzorTranslation()` instead of hardcoded strings.
- Error fallback UI (Feature 4) uses `t()` for localized error messages.

---

## 4. Feature 2 — Telemetry / Product Analytics

### 4.1 How OpenHands does it

- **Hardcoded to PostHog** — vendor lock-in at the call-site level.
- PostHog SDK bundled always (even when disabled).
- Module-level pub/sub for consent (opaque, untestable).
- Raw string event names (typo-prone).
- Cloud-funnel analytics welded to cloud backend concepts.

### 4.2 Weaknesses

- Swapping analytics vendors requires editing every call site.
- PostHog is always in the bundle even if the user opted out.
- Consent is managed via an untyped module-level subscription — hard to test, hard
  to inspect state.
- Event names are raw strings → `track('updte_pushed')` compiles fine but is wrong.

### 4.3 How Vyzorix does it better

1. **Provider-agnostic adapter interface** — `VyzorAnalyticsAdapter` interface;
   ship a no-op default + lazy PostHog adapter. Swap providers without touching call
   sites.
2. **Typed event catalog** — `VyzorAnalyticsEvent` const catalog catches typos at
   compile time (`track('updte_pushed')` → build error).
3. **No-op default** — zero bundle cost when disabled; PostHog loaded lazily only if
   consent granted + key present.
4. **Persisted consent store** (zustand + persist) — `'pending' | 'granted' | 'denied'`,
   testable, devtools-visible (Feature 5), survives reloads.
5. **DNT + `VITE_DO_NOT_TRACK` respect** — same as OpenHands, but centralized in config.

### 4.4 Dependencies

| Package | Type | Purpose |
|---|---|---|
| `posthog-js` | dependency (lazy) | PostHog adapter; dynamically imported only when consent granted + key present — not in main bundle |

### 4.5 File layout

```
src/lib/analytics/
├── vyzor-analytics-adapter.ts         # interface VyzorAnalyticsAdapter { identify, track, page, flush, reset, setConsent }
├── vyzor-noop-analytics-adapter.ts    # default no-op (zero-cost)
├── vyzor-posthog-analytics-adapter.ts # lazy-loaded PostHog impl (dynamic import)
├── vyzor-analytics-events.ts          # VyzorAnalyticsEvent const catalog (typed)
├── vyzor-analytics-config.ts          # reads VITE_ANALYTICS_KEY, VITE_DO_NOT_TRACK, navigator.doNotTrack
├── vyzor-analytics-context.ts         # module-level adapter registry (get/set active adapter)
└── index.ts                           # barrel

src/stores/
└── vyzor-analytics-consent-store.ts   # persisted 'pending'|'granted'|'denied' + setConsent

src/providers/
└── vyzor-analytics-provider.tsx       # boots adapter on mount, syncs consent → adapter

src/hooks/analytics/
├── use-vyzor-analytics.ts             # { track, identify, page, consent, setConsent }
├── use-vyzor-analytics-consent.ts     # consent-only hook (for banner)
└── index.ts

src/ui/components/analytics/
└── vyzor-analytics-consent-banner.tsx # consent UI (uses useVyzorAnalyticsConsent)
```

### 4.6 Layer fit

| File | Layer | Imports from |
|---|---|---|
| `lib/analytics/*` | lib | posthog-js (lazy), external |
| `providers/vyzor-analytics-provider.tsx` | providers | lib/analytics, stores |
| `hooks/analytics/*` | hooks | lib/analytics, stores |
| `stores/vyzor-analytics-consent-store.ts` | stores | lib/state (Feature 5) |
| `ui/components/analytics/*` | ui | hooks/analytics only ✓ |

### 4.7 Integration points

- `routes/__root.tsx` — mount `<VyzorAnalyticsProvider>` + `<VyzorAnalyticsConsentBanner>`.
- Error reporter hook (Feature 4) calls `track('error_reported', { category })`.
- Consent store is persisted; the provider syncs the adapter's `setConsent` on change.

### 4.8 Consent model

```
true  → opt in  (user explicitly accepted)  → adapter.setConsent(true), PostHog loaded
false → opt out (user explicitly denied)     → adapter.setConsent(false), PostHog never loaded
null  → pending (not yet collected)          → adapter.setConsent(false), banner shown
```

---

## 5. Feature 3 — E2E / Integration Testing Infrastructure (MSW + Playwright)

### 5.1 How OpenHands does it

- **4 Playwright configs** (standard, live, mock-llm, mock-llm-docker) — config sprawl.
- MSW handlers ad-hoc per domain, **not aligned to the API client contract** → mocks drift.
- Hand-rolled mock data, not shared between unit and E2E.
- Stryker mutation testing (expensive CI, marginal value here).

### 5.2 Weaknesses

- Four config files for the same tool is unmaintainable.
- Mocks that aren't derived from the real API contract drift silently — tests pass
  against stale mocks while production breaks.
- Duplicated fixture data between vitest and Playwright.
- Stryker adds significant CI time for low ROI in this codebase.

### 5.3 How Vyzorix does it better

1. **Single Playwright config with projects** — `chromium` / `firefox` + a `mock-api`
   project that boots MSW. One file, not four.
2. **MSW handlers path-aligned to the real API client** — handlers import the same
   endpoint paths the REST client uses, so mocks cannot drift from the contract.
3. **MSW usable by vitest too** — replaces hand-rolled `vi.mock('@vyzorix/api-client')`
   with real HTTP interception, testing the actual fetch layer. (Current updates hook
   tests mock the module; MSW would test the real `updates.getVersions()` HTTP call.)
4. **Shared typed fixtures** — `buildVersion()`, `buildPush()`, `buildDevice()`
   factory functions usable by both vitest and Playwright.
5. **No Stryker** — skip mutation testing (high CI cost, low ROI for this codebase).
   Add later if needed.

### 5.4 Dependencies

| Package | Type | Purpose |
|---|---|---|
| `@playwright/test` | devDependency | E2E browser automation |
| `msw` | devDependency | HTTP request mocking (vitest + Playwright + browser) |

### 5.5 File layout

```
src/test/msw/
├── vyzor-msw-server.ts               # MSW setupServer (vitest) + setupWorker (browser)
├── vyzor-msw-handlers-auth.ts        # /api/auth/* handlers
├── vyzor-msw-handlers-devices.ts     # /api/devices/* handlers
├── vyzor-msw-handlers-updates.ts     # /api/updates/* handlers (aligned to rest/updates-endpoints.ts)
├── vyzor-msw-handlers-index.ts       # combines all handlers
└── vyzor-msw-test-setup.ts           # vitest integration (beforeEach: reset handlers)

src/test/fixtures/
└── vyzor-test-fixtures.ts            # typed factories: buildVersion, buildPush, buildDevice, buildOperator

playwright.config.ts                  # single config: projects = [chromium, firefox, mock-api]
e2e/
├── vyzor-smoke-e2e.spec.ts           # app loads, nav works
└── vyzor-updates-e2e.spec.ts         # updates list → push → history flow
```

### 5.6 Layer fit

| File | Layer | Imports from |
|---|---|---|
| `src/test/msw/*` | test (outside lint scope) | msw, @vyzorix/api-client (for path alignment) |
| `src/test/fixtures/*` | test (outside lint scope) | @vyzorix/api-client (domain types) |
| `playwright.config.ts` | root (outside src) | @playwright/test |
| `e2e/*.spec.ts` | root (outside src) | @playwright/test, test/fixtures |

### 5.7 Integration points

- `vitest.config.ts` — add `src/test/msw/` setup file to `setupFiles`.
- Existing hook tests can opt-in to MSW (gradual migration; `vi.mock` still works).
- `package.json` scripts: add `test:e2e: playwright test`.

### 5.8 MSW handler contract alignment

Each handler file imports the REST endpoint base path from the API client so the mock
URL and the real URL are the same string:

```
vyzor-msw-handlers-updates.ts → imports rest endpoint paths from @vyzorix/api-client
  → http.get('/v1/updates/versions', ...) matches the real updates.getVersions() call
```

This eliminates the class of bug where a mock returns 200 for a path the real client
no longer calls.

---

## 6. Feature 4 — Error Boundary + Error Store

### 6.1 How OpenHands does it

- **Single boundary** (`root-layout.tsx` `ErrorBoundary`) — only catches React Router
  route errors, not arbitrary render crashes.
- Error store is flat (message + type + code) — no recovery strategy, no retry.
- No error classifier — UI logic duplicated to decide what to show.
- No error reporter hook — components improvise.

### 6.2 Weaknesses

- A component that throws during render (not a route error) is not caught by the
  route boundary → white screen.
- Flat error store can't drive recovery UI (retry vs login vs reload).
- Without a classifier, every component re-implements "is this a 401? a 500?"
- No uniform error surface → inconsistent UX.

### 6.3 How Vyzorix does it better

1. **Dual boundary** — TanStack Router `errorComponent` for route errors (404, route
   throw) **+** a React class `ErrorBoundary` for render crashes (a component
   throwing in render). OpenHands only has the route one.
2. **Error classifier** — `classifyVyzorError(error)` →
   `{ category: 'network'|'auth'|'server'|'render'|'unknown', recoverable, retryable }`.
   Drives the UI decision, not duplicated logic.
3. **Error store with recovery** — stores the error, classification, retry count,
   and exposes `retry()`, `reload()`, `goHome()`.
4. **Error reporter hook** — `useVyzorErrorReporter()` lets any hook/component
   surface errors uniformly → store + telemetry.
5. **Telemetry integration** — reported errors auto-flow to the analytics adapter
   (Feature 2).

### 6.4 Dependencies

None new. React class components and TanStack Router's `errorComponent` are built-in.

### 6.5 File layout

```
src/lib/error/
├── vyzor-error-classifier.ts         # classifyVyzorError(error) → VyzorErrorClassification
├── vyzor-error-types.ts              # VyzorErrorCategory, VyzorErrorClassification, VyzorReportedError
└── index.ts

src/stores/
└── vyzor-error-store.ts              # { error, classification, retryCount, report, dismiss, retry }

src/hooks/error/
├── use-vyzor-error-reporter.ts       # reportError(error, context?) → store + telemetry
├── use-vyzor-error-recovery.ts       # { retry, reload, goHome, dismiss }
└── index.ts

src/ui/components/error/
├── vyzor-error-boundary.tsx          # React class boundary (render crashes) — wraps app root
├── vyzor-route-error.tsx             # TanStack Router errorComponent (route errors)
└── vyzor-error-fallback.tsx          # shared fallback UI (classified: retry/login/reload buttons)
```

### 6.6 Layer fit

| File | Layer | Imports from |
|---|---|---|
| `lib/error/*` | lib | (none external) |
| `stores/vyzor-error-store.ts` | stores | lib/error, lib/state (Feature 5) |
| `hooks/error/*` | hooks | lib/error, stores, hooks/analytics (for telemetry) |
| `ui/components/error/*` | ui | hooks/error, hooks/i18n only ✓ |

### 6.7 Integration points

- `routes/__root.tsx`:
  - Set `errorComponent: VyzorRouteError` on the root route.
  - Wrap `<Outlet>` in `<VyzorErrorBoundary>`.
- Error reporter hook calls `track('error_reported', { category, message })` via
  Feature 2's analytics adapter.
- Error fallback UI uses `useVyzorTranslation()` (Feature 1) for localized messages.

### 6.8 Error classification matrix

| Category | Triggered by | Recoverable | Retryable | UI action |
|---|---|---|---|---|
| `network` | fetch failure, timeout, offline | yes | yes | Retry button |
| `auth` | 401, 403, token expired | yes | no | Login redirect |
| `server` | 500, 502, 503 | maybe | yes | Retry + reload |
| `render` | React render throw | no | no | Reload button |
| `route` | 404, route error | no | no | Go home |
| `unknown` | anything else | unknown | yes | Retry + reload |

---

## 7. Feature 5 — Zustand Devtools Middleware (Factory)

### 7.1 How OpenHands does it

- Manually wraps every store: `create<>()(devtools(set, { name: "..." }))` —
  boilerplate, inconsistent naming, easy to forget.
- No prod-disable pattern — devtools overhead in production.
- Doesn't compose `persist + devtools` consistently.

### 7.2 Weaknesses

- Every store author must remember to add devtools + name it correctly.
- Devtools runs in production (overhead: extra Proxy + serialization).
- `persist(devtools(set))` vs `devtools(persist(set))` ordering is a common zustand
  footgun; OpenHands doesn't enforce it.

### 7.3 How Vyzorix does it better

1. **`createVyzorStore` factory** — one entry point:
   `createVyzorStore<T>('AuthStore', initializer, options?)`. Auto-wraps devtools,
   auto-derives name, auto-disables in prod.
2. **Auto prod-disable** — `import.meta.env.PROD` check inside the factory; zero
   overhead in production without per-store conditionals.
3. **Composes `persist + devtools` correctly** — factory handles the
   `persist(devtools(set))` ordering every time.
4. **One-line migration** — each store changes `create<T>(...)` →
   `createVyzorStore<T>('Name', ...)`.

### 7.4 Dependencies

None new. `zustand/middleware` ships with Zustand v5.

### 7.5 File layout

```
src/lib/state/
├── vyzor-store-factory.ts            # createVyzorStore<T>(name, initializer, options?)
├── vyzor-store-devtools.ts           # isDevtoolsEnabled, buildDevtoolsOptions(name)
└── index.ts

# Migration (all 13 existing stores):
# stores/auth-store.ts
# stores/command-dispatch-store.ts
# stores/command-queue-store.ts
# stores/connectivity-store.ts
# stores/dashboard-store.ts
# stores/device-selector-store.ts
# stores/diagnostics-store.ts
# stores/log-stream-store.ts
# stores/metrics-realtime-store.ts
# stores/theme-store.ts
# stores/timeline-stream-store.ts
# stores/updates-store.ts
# stores/websocket-store.ts
# → replace `create<T>(...)` with `createVyzorStore<T>('Name', ...)`
```

### 7.6 Layer fit

| File | Layer | Imports from |
|---|---|---|
| `lib/state/*` | lib | zustand, zustand/middleware |
| `stores/*` | stores | lib/state |

### 7.7 Factory signature

```typescript
interface CreateVyzorStoreOptions {
  persist?: {
    name: string;
    storage?: StateStorage;
    partialize?: (state: any) => unknown;
  };
  devtoolsName?: string;
}

function createVyzorStore<T>(
  name: string,
  initializer: (set: ..., get: ...) => T,
  options?: CreateVyzorStoreOptions
): UseBoundStore<StoreApi<T>>;
```

- If `options.persist` is provided → wraps with `persist(devtools(set))`.
- If `import.meta.env.PROD` → devtools middleware is skipped (no-op pass-through).
- `devtoolsName` defaults to the `name` argument.

### 7.8 Stores that already use `persist`

These stores already import `persist` from `zustand/middleware` and must be migrated
to use the factory's `options.persist` instead:

- `command-queue-store.ts`
- `theme-store.ts`
- `updates-store.ts` (via Feature 5 migration — currently no persist, but future)

The factory's `persist(devtools(set))` ordering ensures persisted state is visible
in devtools while hydrated correctly.

---

## 8. Implementation Order

Dependency-driven ordering. Each phase ends with the verification gates in §9.

| Phase | Feature | Why this order |
|---|---|---|
| **1** | Devtools factory (F5) | Foundation — all stores use it; do first so new stores in F1/F2/F4 inherit it |
| **2** | Error boundary (F4) | Independent, but error reporter hook (F4) feeds into telemetry (F2) |
| **3** | i18n (F1) | Independent; error fallback UI will want `t()` |
| **4** | Telemetry (F2) | Error reporter (F4) integrates with it; consent banner needs i18n `t()` |
| **5** | MSW + Playwright (F3) | Last — tests all the above; MSW replaces hand-rolled mocks in existing tests |

---

## 9. Verification Gates

After each phase:

1. **TypeScript:** `cd apps/VyzoriX_web && tsc --noEmit` → **0 errors**
2. **ESLint:** `cd apps/VyzoriX_web && eslint src/<new-paths>` → **0 errors, 0 warnings**
3. **Vitest:** `cd apps/VyzoriX_web && vitest run` → **all tests pass** (no regressions)
4. **SDK typecheck:** `cd packages/API_Client && tsc --noEmit --ignoreDeprecations 6.0` → **0 errors**
   (should be unaffected — all changes are web-app-only)

After Phase 5 additionally:

5. **Playwright:** `npx playwright test` → **all E2E specs pass**

---

## 10. File Inventory (Complete)

### Phase 1 — Devtools factory

| # | File | Status |
|---|---|---|
| 1 | `src/lib/state/vyzor-store-factory.ts` | new |
| 2 | `src/lib/state/vyzor-store-devtools.ts` | new |
| 3 | `src/lib/state/index.ts` | new |
| 4–16 | `src/stores/*.ts` (13 files) | modified (factory migration) |

### Phase 2 — Error boundary + store

| # | File | Status |
|---|---|---|
| 17 | `src/lib/error/vyzor-error-types.ts` | new |
| 18 | `src/lib/error/vyzor-error-classifier.ts` | new |
| 19 | `src/lib/error/index.ts` | new |
| 20 | `src/stores/vyzor-error-store.ts` | new |
| 21 | `src/hooks/error/use-vyzor-error-reporter.ts` | new |
| 22 | `src/hooks/error/use-vyzor-error-recovery.ts` | new |
| 23 | `src/hooks/error/index.ts` | new |
| 24 | `src/ui/components/error/vyzor-error-boundary.tsx` | new |
| 25 | `src/ui/components/error/vyzor-route-error.tsx` | new |
| 26 | `src/ui/components/error/vyzor-error-fallback.tsx` | new |
| 27 | `routes/__root.tsx` | modified (wire boundaries) |

### Phase 3 — i18n

| # | File | Status |
|---|---|---|
| 28 | `src/lib/i18n/vyzor-i18n-config.ts` | new |
| 29 | `src/lib/i18n/vyzor-i18n-types.ts` | new |
| 30 | `src/lib/i18n/vyzor-i18n-loader.ts` | new |
| 31 | `src/lib/i18n/locales/en/vyzor-common-en.ts` | new |
| 32 | `src/lib/i18n/locales/en/vyzor-updates-en.ts` | new |
| 33 | `src/lib/i18n/locales/fr/vyzor-common-fr.ts` | new |
| 34 | `src/lib/i18n/locales/fr/vyzor-updates-fr.ts` | new |
| 35 | `src/lib/i18n/index.ts` | new |
| 36 | `src/providers/vyzor-i18n-provider.tsx` | new |
| 37 | `src/hooks/i18n/use-vyzor-translation.ts` | new |
| 38 | `src/hooks/i18n/use-vyzor-language.ts` | new |
| 39 | `src/hooks/i18n/index.ts` | new |
| 40 | `src/stores/vyzor-i18n-store.ts` | new |

### Phase 4 — Telemetry

| # | File | Status |
|---|---|---|
| 41 | `src/lib/analytics/vyzor-analytics-adapter.ts` | new |
| 42 | `src/lib/analytics/vyzor-noop-analytics-adapter.ts` | new |
| 43 | `src/lib/analytics/vyzor-posthog-analytics-adapter.ts` | new |
| 44 | `src/lib/analytics/vyzor-analytics-events.ts` | new |
| 45 | `src/lib/analytics/vyzor-analytics-config.ts` | new |
| 46 | `src/lib/analytics/vyzor-analytics-context.ts` | new |
| 47 | `src/lib/analytics/index.ts` | new |
| 48 | `src/stores/vyzor-analytics-consent-store.ts` | new |
| 49 | `src/providers/vyzor-analytics-provider.tsx` | new |
| 50 | `src/hooks/analytics/use-vyzor-analytics.ts` | new |
| 51 | `src/hooks/analytics/use-vyzor-analytics-consent.ts` | new |
| 52 | `src/hooks/analytics/index.ts` | new |
| 53 | `src/ui/components/analytics/vyzor-analytics-consent-banner.tsx` | new |

### Phase 5 — MSW + Playwright

| # | File | Status |
|---|---|---|
| 54 | `src/test/msw/vyzor-msw-server.ts` | new |
| 55 | `src/test/msw/vyzor-msw-handlers-auth.ts` | new |
| 56 | `src/test/msw/vyzor-msw-handlers-devices.ts` | new |
| 57 | `src/test/msw/vyzor-msw-handlers-updates.ts` | new |
| 58 | `src/test/msw/vyzor-msw-handlers-index.ts` | new |
| 59 | `src/test/msw/vyzor-msw-test-setup.ts` | new |
| 60 | `src/test/fixtures/vyzor-test-fixtures.ts` | new |
| 61 | `playwright.config.ts` | new |
| 62 | `e2e/vyzor-smoke-e2e.spec.ts` | new |
| 63 | `e2e/vyzor-updates-e2e.spec.ts` | new |

### Tests (per phase)

Each phase also adds test files under `src/test/` following the existing pattern
(`src/test/**/*.test.{ts,tsx}`):

- Phase 1: `src/test/lib/vyzor-store-factory.test.ts`
- Phase 2: `src/test/lib/vyzor-error-classifier.test.ts`,
  `src/test/stores/vyzor-error-store.test.ts`,
  `src/test/hooks/use-vyzor-error-reporter.test.ts`
- Phase 3: `src/test/stores/vyzor-i18n-store.test.ts`,
  `src/test/hooks/use-vyzor-translation.test.ts`
- Phase 4: `src/test/stores/vyzor-analytics-consent-store.test.ts`,
  `src/test/lib/vyzor-analytics-events.test.ts`
- Phase 5: MSW handlers tested via existing hook tests (gradual migration)

### Dependency additions summary

| Package | Type | Phase |
|---|---|---|
| `i18next` | dependency | 3 |
| `react-i18next` | dependency | 3 |
| `posthog-js` | dependency (lazy-loaded) | 4 |
| `@playwright/test` | devDependency | 5 |
| `msw` | devDependency | 5 |

### Out of scope

- No Stryker mutation testing (low ROI).
- No Electron (not requested).
- No onboarding / command palette (separate batch).
- No cloud / OAuth / MCP / skills (agent-canvas-specific, N/A to Vyzorix).
- No changes to the API Client package (all changes are web-app-only).
