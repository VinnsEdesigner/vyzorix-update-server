# Plan — Logs redesign, page enrichment, loading polish, missing modules

This is a sizeable refactor. Grouping into 4 workstreams so you can approve the whole batch or trim a section before I build.

---

## 1. Logging subsystem (the core of this turn)

Today `logs` only captures WebSocket frames inside `useDeviceStream`. I will lift logging into its own first-class system.

**New: `src/lib/logger.ts`** — app-wide log bus.

- `log.info/warn/error/debug(source, message, meta?)`
- Sources (typed): `ws`, `api`, `command`, `update`, `device`, `alert`, `auth`, `system`
- Ring buffer (configurable, default 1000) + `subscribe()` for React
- Persists last N to `localStorage` so logs survive reloads
- Wired into: `vyzorix-api.ts` (every request + response/error), `useDeviceStream` (connect / open / close / frame / error), command dispatch, update checks, threshold-derived alerts

**New: `src/lib/logs-context.tsx`** — React provider exposing `useLogs()` with live entries + filter helpers.

**New: `src/components/logs/log-console.tsx`** — the redesigned panel:

- Search input (substring match)
- Filter chips: level (info/warn/error/debug) + source (ws/api/command/…)
- Auto-scroll toggle, pause, clear, copy-all, export `.log`
- Density toggle (compact / comfy), timestamp format toggle (relative / absolute)
- Virtualized list (simple windowing — no extra dep) so 1000+ entries stay smooth
- Color-coded by level using semantic tokens
- **Minimize / Expand**: docked drawer at bottom (collapsed = 36px bar with counts, expanded = ~40vh)
- **"Open in full page"** button → routes to `/logs`

**New route: `src/routes/_app.logs.tsx`** — full-page console, same component in `fullscreen` mode.

**New: `src/components/logs/log-dock.tsx`** — the persistent collapsible footer that hosts `LogConsole`, mounted once in `_app.tsx`. Replaces the inline terminal in `_app.diagnostics.tsx`.

---

## 2. Page enrichment (what each page is missing to feel "pro")

### Dashboard

- Header KPI row: Uptime, Risk (live + 60s avg), Thermal (live + peak), Buffer health
- Status strip: WS state, last frame age, server version, command queue depth
- Mini-sparkline charts (risk, thermal, buffer) next to KPIs
- "Recent commands" mini-feed (last 5)
- "Active alerts" badge linking to /alerts

### Alerts

- Severity grouping (Critical / Warning / Info) with counts
- Acknowledge action (persisted to localStorage; ack hides from default view)
- Filter by source + severity, search by message
- "Mute threshold for 5 min" quick action per rule

### Updates

- Current installed vs latest available diff card
- Per-channel selector (stable / beta) if `version.json` exposes them
- Last 5 update history entries (from logger source=`update`)
- "Check now" button with proper spinner, "Force re-download" advanced action
- APK size + checksum + signature presence indicators

### Diagnostics

- 8 command tiles grouped by category (Audio / Projection / Recovery / System)
- Per-command: last dispatch time, last ack status, cooldown countdown
- Dry-run toggle (sends but flags `dryRun:true`)
- Replaces inline terminal → uses the global LogDock

### Settings (already split — add the missing ones)

- `/settings/notifications` — alert sound on/off, browser notifications toggle
- `/settings/appearance` — theme (system/dark/light), density, accent
- Index page becomes a clean grid of cards linking to each sub-page (already partly done)

### Device

- Registration form already persists — add: "Test command secret" action (signs a no-op locally and shows the canonical string + HMAC for comparison with Go server)
- Show full device metadata block (Android version, last seen, online flag) when registered

---

## 3. Loading states (spinners + skeletons everywhere)

- Use existing `Skeleton` and add a shared `<Spinner/>` (already in radix-less form via small `Loader2` from lucide)
- New `src/components/loading/page-skeleton.tsx` (header + grid skeleton)
- Apply to: Dashboard (until first frame arrives), Updates (during `version.json` fetch), Device (during registration), Diagnostics (per-command pending state), Alerts (initial render)
- Server health badge: show subtle pulsing spinner only on first check, then silent

---

## 4. Missing frontend modules from your tree (pragmatic mapping → TanStack)

Your tree was written for a Vite+RR SPA. I'm translating to our TanStack Start layout — same intent, different filenames. Creating only what adds real value (skipping pure-stub files like `useAuth.ts` until auth is actually wired).

**Will create:**

- `src/components/layout/footer.tsx` — version, build target, API base URL
- `src/components/dashboard/metrics-summary.tsx`
- `src/components/dashboard/system-alerts.tsx` (mini feed used by Dashboard)
- `src/components/device/device-control-panel.tsx` (extracted from diagnostics)
- `src/components/device/route-state-card.tsx`
- `src/components/device/thermal-metrics-card.tsx`
- `src/components/device/update-state-card.tsx`
- `src/components/charts/risk-score-chart.tsx`
- `src/components/charts/thermal-chart.tsx`
- `src/components/charts/buffer-health-chart.tsx`
- `src/components/charts/memory-footprint-chart.tsx`
- `src/components/ui/spinner.tsx` (tiny wrapper around `Loader2`)
- `src/lib/format.ts` — already exists, will extend (hmac truncation, version diff)

**Skipping (with reason):**

- `LoginPage`, `useAuth`, `authService`, `AuthContext` — auth isn't enabled yet; creating empty stubs adds dead code. Will add when you ask for auth.
- `Sidebar/Navbar` — already exist as `app-sidebar.tsx` + topbar in `_app.tsx`
- `WebSocketContext/ThemeContext` — already covered by `device-stream-context.tsx` + next-themes
- `models/*.interface.ts` — types already live in `vyzorix-api.ts`; no need to duplicate
- `NotFoundPage` — TanStack root already has `notFoundComponent`
- Single-device focus → no `DevicesPage`, no `DeviceGrid` (per your earlier "just C22" decision)

---

## Files summary

**New (~15):** `lib/logger.ts`, `lib/logs-context.tsx`, `components/logs/log-console.tsx`, `components/logs/log-dock.tsx`, `routes/_app.logs.tsx`, `routes/_app.settings.notifications.tsx`, `routes/_app.settings.appearance.tsx`, `components/layout/footer.tsx`, `components/dashboard/metrics-summary.tsx`, `components/dashboard/system-alerts.tsx`, `components/device/device-control-panel.tsx`, `components/device/route-state-card.tsx`, `components/device/thermal-metrics-card.tsx`, `components/device/update-state-card.tsx`, `components/charts/{risk,thermal,buffer,memory}-chart.tsx`, `components/ui/spinner.tsx`, `components/loading/page-skeleton.tsx`

**Edited:** `_app.tsx` (mount LogDock), `_app.dashboard.tsx`, `_app.alerts.tsx`, `_app.updates.tsx`, `_app.diagnostics.tsx`, `_app.device.tsx`, `_app.settings.index.tsx`, `vyzorix-api.ts` (logger calls), `use-device-stream.ts` (logger calls), `app-sidebar.tsx` (Logs link), `routeTree.gen.ts`

---

## Open questions before I start

1. **Logs persistence**: ring buffer in localStorage (1000 lines ≈ 200KB) — OK or memory-only?
2. **Settings additions**: want `/settings/notifications` + `/settings/appearance` now, or skip and just enrich the existing four?
3. **Login/auth stubs**: leave out entirely (my plan), or scaffold a real auth pass against the Go server's `middleware/auth.go` JWT now?

Say "go" to build all of it, or tell me which sections to trim.
