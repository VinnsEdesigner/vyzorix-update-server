/* app.js — Vyzorix UI mock (Grafana-style). Plain JS, no build step.
 *
 * Views mirror the real vyzorix-update-server surface (REST + GraphQL):
 * dashboard, devices, commands, telemetry, updates (OTA), alerts, logs,
 * organizations, api-keys, settings. Data here is mock/illustrative.
 */

const $ = (sel, root = document) => root.querySelector(sel);
const $$ = (sel, root = document) => [...root.querySelectorAll(sel)];
const el = (html) => { const t = document.createElement("template"); t.innerHTML = html.trim(); return t.content.firstElementChild; };

// ---------------------------------------------------------------- mock data
const DEVICES = [
  { imei: "354481100123451", model: "Vyzorix AR-1", fw: "0.4.2", status: "online",  battery: 87, thermal: "nominal", risk: 12, rssi: -58, org: "Acme Ops" },
  { imei: "354481100123452", model: "Vyzorix AR-1", fw: "0.4.2", status: "online",  battery: 64, thermal: "warm",    risk: 38, rssi: -71, org: "Acme Ops" },
  { imei: "354481100123453", model: "Vyzorix AR-2", fw: "0.3.9", status: "offline", battery: 4,  thermal: "—",       risk: 91, rssi: 0,   org: "Northwind" },
  { imei: "354481100123454", model: "Vyzorix AR-2", fw: "0.4.1", status: "online",  battery: 52, thermal: "hot",     risk: 66, rssi: -83, org: "Northwind" },
  { imei: "354481100123455", model: "Vyzorix AR-1", fw: "0.4.2", status: "online",  battery: 95, thermal: "nominal", risk: 5,  rssi: -49, org: "Acme Ops" },
  { imei: "354481100123456", model: "Vyzorix AR-2", fw: "0.3.9", status: "offline", battery: 22, thermal: "—",       risk: 74, rssi: 0,   org: "Globex" },
];

const COMMANDS = [
  { id: "cmd_8f21", imei: "354481100123451", type: "reboot",        status: "delivered", at: "12:41:08", hmac: "ok" },
  { id: "cmd_8f20", imei: "354481100123452", type: "start_capture", status: "acked",     at: "12:39:55", hmac: "ok" },
  { id: "cmd_8f1f", imei: "354481100123454", type: "set_volume",    status: "pending",   at: "12:38:12", hmac: "ok" },
  { id: "cmd_8f1e", imei: "354481100123453", type: "stop_capture",  status: "failed",    at: "12:35:40", hmac: "bad" },
  { id: "cmd_8f1d", imei: "354481100123455", type: "factory_reset", status: "cancelled", at: "12:30:02", hmac: "ok" },
];

const ALERTS = [
  { sev: "err",  rule: "battery_critical",  device: "354481100123453", msg: "Battery 4% — below 10% threshold", at: "12:40", state: "firing" },
  { sev: "err",  rule: "device_offline",    device: "354481100123456", msg: "No telemetry for 22m",             at: "12:39", state: "firing" },
  { sev: "warn", rule: "thermal_high",      device: "354481100123454", msg: "Thermal zone 'hot' for 6m",         at: "12:37", state: "firing" },
  { sev: "warn", rule: "risk_score",        device: "354481100123452", msg: "Risk 38 (> 30)",                    at: "12:31", state: "resolved" },
  { sev: "info", rule: "fw_update_available", device: "354481100123454", msg: "0.4.2 available (on 0.4.1)",       at: "12:20", state: "resolved" },
];

const LOGS = [
  { lvl: "info",  src: "ws-hub",     msg: "device connected 354481100123451",                t: "12:41:22.114", req: "—",               dur: "—" },
  { lvl: "info",  src: "command",    msg: "dispatch cmd_8f21 reboot → 35…451",                 t: "12:41:08.902", req: "req_9f2a1c",      dur: "18ms" },
  { lvl: "warn",  src: "rate-limit", msg: "bucket exhausted for 10.4.2.19",                    t: "12:40:55.031", req: "req_9f2a19",      dur: "1ms" },
  { lvl: "error", src: "hmac",       msg: "signature mismatch cmd_8f1e (device 35…453)",       t: "12:35:40.400", req: "req_9f2a04",      dur: "3ms" },
  { lvl: "info",  src: "ota",        msg: "manifest served v0.4.2 (312 KB)",                    t: "12:20:11.220", req: "req_9f2a01",      dur: "9ms" },
  { lvl: "debug", src: "graphql",    msg: "subscription telemetryStream opened (org acme)",     t: "12:19:58.007", req: "req_9f2a00",      dur: "—" },
  { lvl: "info",  src: "http",       msg: "POST /graphql 200",                                  t: "12:41:12.660", req: "req_9f2a1b",      dur: "24ms" },
  { lvl: "info",  src: "http",       msg: "GET /v1/device/354481100123452/status 200",         t: "12:41:05.221", req: "req_9f2a17",      dur: "6ms" },
  { lvl: "warn",  src: "auth",       msg: "failed login for op@globex.io (attempt 2/5)",        t: "12:39:12.004", req: "req_9f29ff",      dur: "12ms" },
  { lvl: "error", src: "db",         msg: "sqlite busy: retrying WAL write (1/3)",             t: "12:38:44.010", req: "req_9f29fe",      dur: "140ms" },
  { lvl: "info",  src: "ws-hub",     msg: "device disconnected 354481100123456",               t: "12:38:02.331", req: "—",               dur: "—" },
  { lvl: "debug", src: "cache",      msg: "evicted 128 stale entries (ttl)",                    t: "12:36:19.887", req: "—",               dur: "2ms" },
  { lvl: "info",  src: "command",    msg: "acked cmd_8f20 start_capture by 35…452",             t: "12:36:02.114", req: "req_9f29f1",      dur: "31ms" },
  { lvl: "warn",  src: "ota",        msg: "apk checksum slow (2.1s) for canary",                t: "12:33:48.559", req: "req_9f29e0",      dur: "2100ms" },
];

const ORGS = [
  { name: "Acme Ops",   plan: "Scale", members: 14, devices: 128, role: "Owner",  status: "active" },
  { name: "Northwind",  plan: "Pro",   members: 6,  devices: 42,  role: "Admin",  status: "active" },
  { name: "Globex",     plan: "Free",  members: 2,  devices: 9,   role: "Viewer", status: "suspended" },
];

const SESSIONS = [
  { active: true,  seen: "Now",          loggedOn: "September 12, 2026", ip: "10.4.2.19",    browser: "Chrome 153", os: "Linux",           idp: "grafana" },
  { active: false, seen: "2 hours ago",  loggedOn: "September 12, 2026", ip: "10.4.2.51",    browser: "Safari 18",  os: "macOS 15.1",    idp: "google" },
  { active: false, seen: "Yesterday",    loggedOn: "September 11, 2026", ip: "203.0.113.7",  browser: "Firefox 132", os: "Windows 11",   idp: "github" },
  { active: false, seen: "3 days ago",   loggedOn: "September 9, 2026",  ip: "198.51.100.24", browser: "Chrome 152", os: "Android 15",   idp: "google" },
];

const KEYS = [
  { id: "key_9c2a", client: "ci-deploy",       scopes: "devices:write, ota:read", created: "2026-08-01", last: "12:38", status: "active" },
  { id: "key_7b10", client: "grafana-readonly", scopes: "metrics:read",            created: "2026-07-14", last: "12:41", status: "active" },
  { id: "key_5d33", client: "legacy-sync",     scopes: "devices:read",            created: "2026-03-02", last: "—",     status: "revoked" },
];

// ---------------------------------------------------------------- helpers
const sparkline = (values, color) => window.VyzCharts.sparkline(values, { color });
const statusDot = (s) => `<span class="dot dot--${s === "online" ? "ok" : "off"}"></span>`;
const riskBadge = (r) => r > 70 ? `<span class="badge badge--err">${r}</span>` : r > 35 ? `<span class="badge badge--warn">${r}</span>` : `<span class="badge badge--ok">${r}</span>`;
const sevBadge = (s) => s === "err" ? `<span class="badge badge--err">error</span>` : s === "warn" ? `<span class="badge badge--warn">warning</span>` : `<span class="badge badge--info">info</span>`;
const lvlBadge = (l) => ({ error: "badge--err", warn: "badge--warn", info: "badge--info", debug: "" }[l] ?? "");
function bar(pct, tone) { return `<div class="bar bar--${tone}"><i style="width:${pct}%"></i></div>`; }

function statPanel(title, value, label, delta, tone, spark) {
  return `<div class="panel"><div class="panel__head"><span class="panel__title">${title}</span><span class="panel__menu">⋯</span></div>
    <div class="panel__body">
      <div class="stat__value">${value}</div>
      <div class="stat__label">${label}</div>
      ${delta ? `<div class="stat__delta ${tone}">${delta}</div>` : ""}
      ${spark ? `<div class="stat__chart">${sparkline(spark, tone === "down" ? "#f2495c" : "#73bf69")}</div>` : ""}
    </div></div>`;
}

// ---------------------------------------------------------------- views
const VIEWS = {
  dashboard() {
    const online = DEVICES.filter(d => d.status === "online").length;
    const firing = ALERTS.filter(a => a.state === "firing").length;
    return `
      <div class="page-head">
        <div><h1>Fleet Dashboard</h1><p>Real-time device status, risk, thermal &amp; OTA — org <b>Acme Ops</b></p></div>
        <div class="page-head__actions">
          <button class="btn btn--ghost">Last 1h</button>
          <button class="btn btn--ghost">⟳ Refresh</button>
          <button class="btn btn--primary">+ Add device</button>
        </div>
      </div>

      <div class="grid grid--4">
        ${statPanel("Devices online", `${online}<span style="font-size:18px;color:var(--text-secondary)">/${DEVICES.length}</span>`, "reporting in last 5 min", "▲ 2 vs 1h", "up", [4, 4, 5, 5, 5, 6, 6])}
        ${statPanel("Active alerts", `${firing}`, "firing rules", "▲ 1 vs 1h", "down", [1, 1, 2, 1, 2, 2, 3])}
        ${statPanel("Avg battery", "54%", "across fleet", "▼ 6% vs 1h", "down", [72, 68, 63, 60, 58, 55, 54])}
        ${statPanel("OTA success", "98.2%", "last 7 days", "▲ 0.4%", "up", [96, 97, 97, 98, 98, 98, 98])}
      </div>

      <div class="grid grid--3 mt2">
        <div class="panel" style="grid-column: span 2">
          <div class="panel__head"><span class="panel__title">Telemetry — fleet</span><span class="panel__subtitle">rssi · battery · thermal</span><span class="panel__menu">⋯</span></div>
          <div class="panel__body">
            <div class="chart-host" id="chart-telemetry" data-series="all"></div>
          </div>
        </div>

        <div class="panel">
          <div class="panel__head"><span class="panel__title">Alerts by severity</span></div>
          <div class="panel__body">
            <div class="chart-host" id="chart-sev"></div>
          </div>
        </div>
      </div>

      <div class="panel mt2">
        <div class="panel__head"><span class="panel__title">Devices</span><span class="panel__subtitle">${DEVICES.length} registered</span><span class="panel__menu">⋯</span></div>
        <div class="panel__body" style="padding-top:8px">
          ${devicesTable()}
        </div>
      </div>`;
  },

  devices() {
    return `
      <div class="page-head">
        <div><h1>Devices</h1><p>Fleet registry · ${DEVICES.length} devices · transfer, deregister, tags</p></div>
        <div class="page-head__actions"><button class="btn btn--ghost">Filter</button><button class="btn btn--primary">+ Register</button></div>
      </div>
      <div class="panel"><div class="panel__body" style="padding-top:8px">${devicesTable(true)}</div></div>`;
  },

  commands() {
    return `
      <div class="page-head">
        <div><h1>Commands</h1><p>HMAC-SHA256 signed dispatch · retry / cancel · history</p></div>
        <div class="page-head__actions"><button class="btn btn--ghost">Export history</button><button class="btn btn--primary">Dispatch command</button></div>
      </div>
      <div class="grid grid--4">
        ${statPanel("Dispatched (1h)", "48", "signed commands", "+9", "up", [30, 34, 38, 41, 44, 46, 48])}
        ${statPanel("Acked", "44", "device confirmed", "+8", "up", [28, 31, 35, 38, 41, 43, 44])}
        ${statPanel("Failed", "3", "bad signature / timeout", "+1", "down", [1, 1, 2, 2, 3, 3, 3])}
        ${statPanel("P50 latency", "220ms", "dispatch → ack", "−30ms", "up", [280, 260, 250, 240, 230, 225, 220])}
      </div>
      <div class="grid grid--2 mt2">
        <div class="panel"><div class="panel__head"><span class="panel__title">Commands by type</span><span class="panel__subtitle">last 1h</span></div>
          <div class="panel__body"><div class="chart-host" id="chart-cmd"></div></div></div>
        <div class="panel"><div class="panel__head"><span class="panel__title">Dispatch latency</span><span class="panel__subtitle">p50 / p95</span></div>
          <div class="panel__body"><div class="chart-host" id="chart-lat"></div></div></div>
      </div>
      <div class="panel mt2">
        <div class="panel__head"><span class="panel__title">Recent commands</span><span class="panel__menu">⋯</span></div>
        <div class="panel__body" style="padding-top:8px">
          <table class="table">
            <thead><tr><th>Command</th><th>Device (IMEI)</th><th>Type</th><th>Status</th><th>HMAC</th><th>Time</th></tr></thead>
            <tbody>${COMMANDS.map(c => `<tr>
              <td class="mono">${c.id}</td>
              <td class="mono">${c.imei}</td>
              <td>${c.type}</td>
              <td>${cmdBadge(c.status)}</td>
              <td>${c.hmac === "ok" ? '<span class="badge badge--ok">verified</span>' : '<span class="badge badge--err">mismatch</span>'}</td>
              <td class="muted">${c.at}</td>
            </tr>`).join("")}</tbody>
          </table>
        </div>
      </div>`;
  },

  telemetry() {
    return `
      <div class="page-head"><div><h1>Telemetry</h1><p>WebSocket stream · history &amp; export</p></div>
        <div class="page-head__actions"><button class="btn btn--ghost">Export CSV</button><button class="btn btn--primary">Open live stream</button></div></div>
      <div class="grid grid--2">
        <div class="panel"><div class="panel__head"><span class="panel__title">RSSI</span><span class="panel__subtitle">354481100123451 · dBm</span></div>
          <div class="panel__body"><div class="chart-host" id="chart-rssi"></div></div></div>
        <div class="panel"><div class="panel__head"><span class="panel__title">Battery</span><span class="panel__subtitle">354481100123451 · %</span></div>
          <div class="panel__body"><div class="chart-host" id="chart-batt"></div></div></div>
      </div>
      <div class="panel mt2"><div class="panel__head"><span class="panel__title">Live frames</span><span class="panel__subtitle">telemetryStream subscription</span></div>
        <div class="panel__body" style="padding-top:8px">
          <table class="table"><thead><tr><th>Time</th><th>IMEI</th><th>RSSI</th><th>Battery</th><th>Thermal</th></tr></thead>
          <tbody>${DEVICES.slice(0,4).map((d,i)=>`<tr><td class="mono muted">12:41:${(22-i*3).toString().padStart(2,"0")}</td><td class="mono">${d.imei}</td><td>${d.rssi||"—"} dBm</td><td>${d.battery}%</td><td>${d.thermal}</td></tr>`).join("")}</tbody></table>
        </div></div>
      <div class="panel mt2"><div class="panel__head"><span class="panel__title">RSSI heatmap</span><span class="panel__subtitle">per device · last 10 min</span></div>
        <div class="panel__body"><div class="chart-host" id="chart-heat"></div></div></div>`;
  },

  updates() {
    return `
      <div class="page-head"><div><h1>Updates (OTA)</h1><p>Version manifest &amp; APK distribution</p></div>
        <div class="page-head__actions"><button class="btn btn--ghost">Upload APK</button><button class="btn btn--primary">Publish release</button></div></div>
      <div class="grid grid--3">
        ${statPanel("Current release", "0.4.2", "published 2026-09-11", "", "")}
        ${statPanel("Adoption", "76%", "of fleet on latest", "▲ 12%", "up", [40,48,55,62,68,72,76])}
        ${statPanel("Rollouts", "3", "channels: stable, beta, canary", "", "")}
      </div>
      <div class="grid grid--2 mt2">
        <div class="panel"><div class="panel__head"><span class="panel__title">Adoption rollout</span><span class="panel__subtitle">% of fleet over 7 weeks</span></div>
          <div class="panel__body"><div class="chart-host" id="chart-rollout"></div></div></div>
        <div class="panel"><div class="panel__head"><span class="panel__title">Channel distribution</span><span class="panel__subtitle">% of fleet</span></div>
          <div class="panel__body"><div class="chart-host" id="chart-channels"></div></div></div>
      </div>
      <div class="panel mt2"><div class="panel__head"><span class="panel__title">Release history</span><span class="panel__menu">⋯</span></div>
        <div class="panel__body" style="padding-top:8px">
          <table class="table"><thead><tr><th>Version</th><th>Channel</th><th>Size</th><th>Adoption</th><th>Status</th><th>Published</th></tr></thead>
          <tbody>
            <tr><td class="mono">0.4.2</td><td><span class="badge badge--info">stable</span></td><td>312 KB</td><td>${bar(76,"ok")} 76%</td><td><span class="badge badge--ok">active</span></td><td class="muted">2026-09-11</td></tr>
            <tr><td class="mono">0.4.1</td><td><span class="badge">beta</span></td><td>308 KB</td><td>${bar(18,"info")} 18%</td><td><span class="badge badge--ok">active</span></td><td class="muted">2026-08-28</td></tr>
            <tr><td class="mono">0.3.9</td><td><span class="badge">canary</span></td><td>296 KB</td><td>${bar(6,"warn")} 6%</td><td><span class="badge badge--warn">deprecated</span></td><td class="muted">2026-07-30</td></tr>
          </tbody></table>
        </div></div>`;
  },

  alerts() {
    return `
      <div class="page-head"><div><h1>Alerts</h1><p>Threshold-based rule engine · contact points · history</p></div>
        <div class="page-head__actions"><button class="btn btn--ghost">Contact points</button><button class="btn btn--primary">+ New rule</button></div></div>
      <div class="grid grid--4">
        ${statPanel("Firing", "3", "active now", "", "")}
        ${statPanel("Rules", "12", "enabled", "", "")}
        ${statPanel("MTTA", "2m 10s", "mean time to ack", "▲", "up", [3,3,2.6,2.4,2.3,2.2,2.1])}
        ${statPanel("Notifications", "34", "sent today", "", "")}
      </div>
      <div class="grid grid--2 mt2">
        <div class="panel"><div class="panel__head"><span class="panel__title">Alerts by rule (7d)</span><span class="panel__subtitle">firing count</span></div>
          <div class="panel__body"><div class="chart-host" id="chart-alerts"></div></div></div>
        <div class="panel"><div class="panel__head"><span class="panel__title">Mean time to ack</span><span class="panel__subtitle">minutes</span></div>
          <div class="panel__body"><div class="chart-host" id="chart-mtta"></div></div></div>
      </div>
      <div class="panel mt2"><div class="panel__head"><span class="panel__title">Alert rules</span><span class="panel__menu">⋯</span></div>
        <div class="panel__body" style="padding-top:8px">
          <table class="table"><thead><tr><th>Severity</th><th>Rule</th><th>Device</th><th>Message</th><th>State</th><th>Since</th></tr></thead>
          <tbody>${ALERTS.map(a => `<tr>
            <td>${sevBadge(a.sev)}</td><td class="mono">${a.rule}</td><td class="mono">${a.device}</td>
            <td>${a.msg}</td>
            <td>${a.state === "firing" ? '<span class="badge badge--err">firing</span>' : '<span class="badge badge--ok">resolved</span>'}</td>
            <td class="muted">${a.at}</td></tr>`).join("")}</tbody></table>
        </div></div>`;
  },

  logs() {
    const counts = LOGS.reduce((a, l) => ((a[l.lvl] = (a[l.lvl] || 0) + 1), a), {});
    return `
      <div class="page-head">
        <div><h1>Logs</h1><p>Structured server logs · volume · live tail</p></div>
        <div class="page-head__actions">
          <span class="badge badge--err">${counts.error || 0} error</span>
          <span class="badge badge--warn">${counts.warn || 0} warn</span>
          <button class="btn btn--ghost">Export</button>
          <button class="btn btn--primary"><span class="dot dot--ok" style="margin-right:4px"></span>Live tail</button>
        </div>
      </div>

      <div class="log-tabpanel" data-panel="entries">
        <div class="panel">
          <div class="log-toolbar">
            <div class="logs-search">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="7"/><line x1="21" y1="21" x2="16.7" y2="16.7"/></svg>
              <input class="logs-search__input" id="log-search" placeholder="Filter logs (message, source, request id)…" />
              <kbd>/</kbd>
            </div>
            <div class="seg" id="log-levels">
              <button class="seg__btn is-active" data-lvl="all">All</button>
              <button class="seg__btn" data-lvl="error">Error</button>
              <button class="seg__btn" data-lvl="warn">Warn</button>
              <button class="seg__btn" data-lvl="info">Info</button>
              <button class="seg__btn" data-lvl="debug">Debug</button>
            </div>
            <div class="topbar__spacer"></div>
            <span class="tiny muted" id="log-count">${LOGS.length} entries</span>
          </div>
          <div class="panel__body" style="padding-top:0">
            <div class="log-view" id="log-view">${logRows(LOGS)}</div>
          </div>
        </div>
      </div>

      <div class="log-tabpanel" data-panel="volume" hidden>
        <div class="panel">
          <div class="panel__head"><span class="panel__title">Log volume</span><span class="panel__subtitle">stacked by level · last 22 min</span></div>
          <div class="panel__body">
            <div class="chart-legend" style="padding-top:0">
              <span><i style="background:#f2495c"></i>Error</span>
              <span><i style="background:#ff9830"></i>Warn</span>
              <span><i style="background:#fb7185"></i>Info</span>
              <span><i style="background:#5b5b5b"></i>Debug</span>
            </div>
            <div class="chart-host" id="chart-loghist"></div>
          </div>
        </div>
        <div class="grid grid--3 mt2">
          ${statPanel("Errors", String(counts.error || 0), "last 60 min", "", "")}
          ${statPanel("Warnings", String(counts.warn || 0), "last 60 min", "", "")}
          ${statPanel("Total", String(LOGS.length), "entries retained", "", "")}
        </div>
      </div>`;
  },

  orgs() {
    const totalDevices = ORGS.reduce((a, o) => a + o.devices, 0);
    const totalMembers = ORGS.reduce((a, o) => a + o.members, 0);
    const suspended = ORGS.filter((o) => o.status !== "active").length;
    const PLAN_LIMIT = { Scale: 500, Pro: 100, Free: 10 };
    const initials = (n) => n.split(/\s+/).map((w) => w[0]).join("").slice(0, 2).toUpperCase();
    return `
      <div class="page-head">
        <div><h1>Organizations</h1><p>Multi-tenant · members, roles, invitations &amp; quotas</p></div>
        <div class="page-head__actions">
          <button class="btn btn--ghost">Invitations</button>
          <button class="btn btn--primary">+ New organization</button>
        </div>
      </div>

      <div class="grid grid--4">
        ${statPanel("Organizations", String(ORGS.length), "workspaces", "", "")}
        ${statPanel("Total devices", String(totalDevices), "across all orgs", "▲ 12", "up", [140, 150, 158, 165, 170, 175, 179])}
        ${statPanel("Members", String(totalMembers), "seats in use", "", "")}
        ${statPanel("Suspended", String(suspended), "needs attention", "", "")}
      </div>

      <div class="org-grid mt2">
        ${ORGS.map((o) => {
          const limit = PLAN_LIMIT[o.plan] || 100;
          const pct = Math.round((o.devices / limit) * 100);
          const tone = pct > 85 ? "err" : pct > 60 ? "warn" : "ok";
          return `<div class="panel org-card">
            <div class="org-card__top">
              <div class="org-avatar">${initials(o.name)}</div>
              <div class="org-card__id">
                <div class="org-name">${o.name}</div>
                <div class="tiny muted">${o.plan} plan · your role: ${o.role}</div>
              </div>
              ${o.status === "active" ? '<span class="badge badge--ok">active</span>' : '<span class="badge badge--warn">suspended</span>'}
            </div>
            <div class="org-card__stats">
              <div><div class="tiny muted">Members</div><b>${o.members}</b></div>
              <div><div class="tiny muted">Devices</div><b>${o.devices}</b></div>
              <div><div class="tiny muted">Plan</div><b>${o.plan}</b></div>
            </div>
            <div class="org-card__usage">
              <div class="flex between tiny"><span class="muted">Device quota</span><span>${o.devices} / ${limit}</span></div>
              <div style="margin-top:6px">${bar(pct, tone)}</div>
            </div>
            <div class="org-card__actions">
              <button class="btn btn--ghost">Manage members</button>
              <button class="btn btn--ghost">Settings</button>
            </div>
          </div>`;
        }).join("")}
      </div>`;
  },

  "api-keys"() {
    const active = KEYS.filter((k) => k.status === "active").length;
    const revoked = KEYS.filter((k) => k.status !== "active").length;
    const mask = (id) => `${id.slice(0, 6)}••••••••${id.slice(-2)}`;
    return `
      <div class="page-head">
        <div><h1>API Keys</h1><p>Service accounts · scoped tokens · rotation</p></div>
        <div class="page-head__actions">
          <button class="btn btn--ghost">Audit log</button>
          <button class="btn btn--primary">+ Create key</button>
        </div>
      </div>

      <div class="grid grid--4">
        ${statPanel("Active keys", String(active), "in use right now", "", "")}
        ${statPanel("Revoked", String(revoked), "retained for audit", "", "")}
        ${statPanel("Requests (24h)", "12.4k", "across all keys", "▲ 8%", "up", [9, 10, 10.5, 11, 11.6, 12, 12.4])}
        ${statPanel("Rotation due", "1", "older than 90 days", "", "")}
      </div>

      <div class="grid grid--2 mt2">
        <div class="panel"><div class="panel__head"><span class="panel__title">Requests per key</span><span class="panel__subtitle">last 24h</span></div>
          <div class="panel__body"><div class="chart-host" id="chart-keyusage"></div></div></div>
        <div class="panel"><div class="panel__head"><span class="panel__title">Authentication</span><span class="panel__subtitle">success vs rejected · 24h</span></div>
          <div class="panel__body"><div class="chart-host" id="chart-auth"></div></div></div>
      </div>

      <div class="panel mt2">
        <div class="panel__head"><span class="panel__title">Keys</span><span class="panel__subtitle">${KEYS.length} total</span><span class="panel__menu">⋯</span></div>
        <div class="panel__body" style="padding-top:8px">
          <table class="table">
            <thead><tr><th>Key ID</th><th>Client</th><th>Scopes</th><th>Created</th><th>Last used</th><th>Status</th><th></th></tr></thead>
            <tbody>${KEYS.map((k) => `<tr>
              <td>
                <span class="mono key-id">${mask(k.id)}</span>
                <button class="icon-btn key-reveal" data-key="${k.id}" title="Reveal">${eyeIcon()}</button>
              </td>
              <td><b>${k.client}</b></td>
              <td>${k.scopes.split(",").map((s) => `<span class="badge scope">${s.trim()}</span>`).join(" ")}</td>
              <td class="muted">${k.created}</td>
              <td class="muted">${k.last}</td>
              <td>${k.status === "active" ? '<span class="badge badge--ok">active</span>' : '<span class="badge badge--err">revoked</span>'}</td>
              <td class="row-actions">
                <button class="btn btn--ghost tiny">Rotate</button>
                <button class="btn btn--ghost tiny key-revoke${k.status !== "active" ? " is-disabled" : ""}">${k.status === "active" ? "Revoke" : "—"}</button>
              </td>
            </tr>`).join("")}</tbody>
          </table>
        </div>
      </div>`;
  },

  profile() {
    return `
      <div class="profile">
        <aside class="profile-card panel">
          <div class="profile-card__avatar">OP</div>
          <div class="profile-card__name">Oscar Porter</div>
          <div class="profile-card__role">Owner · Acme Ops</div>
          <div class="flex" style="justify-content:center;gap:6px;margin-top:6px">
            <span class="badge badge--ok">MFA on</span>
            <span class="badge badge--info">Google linked</span>
          </div>
          <div class="profile-card__meta">
            <div class="flex between"><span class="muted">Email</span><span>op@acme.io</span></div>
            <div class="flex between"><span class="muted">Phone</span><span>+44 •••• 2210</span></div>
            <div class="flex between"><span class="muted">Member since</span><span>2026-03-02</span></div>
            <div class="flex between"><span class="muted">Last login</span><span>12:41</span></div>
          </div>
          <div class="profile-card__session">
            <div class="tiny muted" style="margin-bottom:4px">Current session</div>
            <div class="flex between"><span class="muted">IP address</span><span class="mono">10.4.2.19</span></div>
            <div class="flex between"><span class="muted">Browser &amp; OS</span><span>Chrome 153 · Linux</span></div>
          </div>
          <div class="profile-card__actions">
            <button class="btn btn--primary" id="profile-upload">Upload photo</button>
            <button class="btn btn--ghost">Copy user ID</button>
          </div>
        </aside>

        <div class="profile-main">
          <div class="profile-panel" data-panel="info">
            <div class="grid grid--4">
              ${statPanel("Sessions", "3", "active devices", "", "")}
              ${statPanel("Commands sent", "1.2k", "all time", "▲ 40", "up", [900, 980, 1040, 1090, 1140, 1180, 1200])}
              ${statPanel("Actions (30d)", "486", "audited events", "", "")}
              ${statPanel("Alerts acked", "112", "mean 2m 10s", "", "")}
            </div>

            <div class="panel mt2">
              <div class="panel__head"><span class="panel__title">Profile</span><span class="panel__subtitle">Your operator identity</span></div>
              <div class="panel__body">
                ${field("Display name", "Shown to other operators in the organization", input("p-name", "Oscar Porter"))}
                ${field("Email", "Used for sign-in and notifications", input("p-email", "op@acme.io"))}
                ${field("Job title", "Optional — shown on your profile", input("p-title", "Fleet Operations Lead"))}
                ${field("Timezone", "Applied to your timestamps", select("p-tz", ["UTC", "Europe/London", "America/New_York"], "UTC"))}
                ${field("Language", "Console display language", select("p-lang", ["English", "Deutsch", "Français"], "English"))}
              </div>
            </div>

            <div class="panel mt2">
              <div class="panel__head"><span class="panel__title">Security</span></div>
              <div class="panel__body">
                ${field("Password", "Last changed 42 days ago", `<button class="btn btn--ghost">Change password</button>`)}
                ${field("Multi-factor auth", "TOTP required on every login", switchEl("p-mfa", true))}
                ${field("Backup codes", "8 of 10 remaining", `<button class="btn btn--ghost">Regenerate</button>`)}
                ${field("Active sessions", "Revoke every other signed-in device", `<button class="btn btn--danger">Revoke all others</button>`)}
              </div>
            </div>

            <div class="panel mt2">
              <div class="panel__head"><span class="panel__title">API tokens</span><span class="panel__subtitle">Personal access tokens</span></div>
              <div class="panel__body" style="padding-top:8px">
                <table class="table">
                  <thead><tr><th>Name</th><th>Scope</th><th>Created</th><th>Last used</th><th></th></tr></thead>
                  <tbody>
                    <tr><td><b>cli-laptop</b></td><td><span class="badge scope">full</span></td><td class="muted">2026-06-01</td><td class="muted">12:30</td><td class="row-actions"><button class="btn btn--ghost tiny">Revoke</button></td></tr>
                    <tr><td><b>ci-readonly</b></td><td><span class="badge scope">metrics:read</span></td><td class="muted">2026-07-14</td><td class="muted">11:58</td><td class="row-actions"><button class="btn btn--ghost tiny">Revoke</button></td></tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          <div class="profile-panel" data-panel="preferences" hidden>
            <div class="panel">
              <div class="panel__head"><span class="panel__title">Preferences</span><span class="panel__subtitle">Your org &amp; UI defaults</span></div>
              <div class="panel__body">
                ${field("Interface theme", "Appearance of the Grafana UI — built-in and experimental themes", themeSelect())}
                ${field("Home Dashboard", "Landing dashboard when you sign in", select("pref-home", ["Default dashboard", "None", "Fleet overview", "OTA adoption", "Alert activity"], "Fleet overview"))}
                ${field("Timezone", "Used for dashboards and alert timestamps", select("pref-timezone", ["Default", "Browser time", "UTC", "Europe/London", "America/New_York"], "Browser time"))}
                ${field("Week start", "First day of the week", select("pref-week", ["Default", "Monday", "Sunday", "Saturday"], "Monday"))}
                ${field("Language", "Console display language", select("pref-language", ["Default", "English", "Deutsch", "Français", "Español", "Português"], "English"))}
                <div class="field">
                  <div class="field__label"></div>
                  <div class="field__control"><button class="btn btn--primary" id="prefs-save">Save preferences</button></div>
                </div>
              </div>
            </div>
          </div>
          <div class="profile-panel" data-panel="sessions" hidden>
            <div class="panel">
              <div class="panel__head"><span class="panel__title">Sessions</span><span class="panel__subtitle">Signed-in devices for this account</span></div>
              <div class="panel__body" style="padding-top:8px">
                <table class="table">
                  <thead><tr><th>Last seen</th><th>Logged on</th><th>IP address</th><th>Browser &amp; OS</th><th>Identity Provider</th><th></th></tr></thead>
                  <tbody>${SESSIONS.map((s) => `<tr>
                    <td>${s.active ? '<span class="badge badge--ok">Now</span>' : `<span class="muted">${s.seen}</span>`}</td>
                    <td class="muted">${s.loggedOn}</td>
                    <td class="mono">${s.ip}</td>
                    <td>${s.browser} on ${s.os}</td>
                    <td><span class="badge scope">${s.idp}</span></td>
                    <td class="row-actions"><button class="btn btn--ghost tiny session-revoke${s.active ? " is-current" : ""}">${s.active ? "This device" : "Revoke"}</button></td>
                  </tr>`).join("")}</tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      </div>`;
  },

  // ----- Grafana "Administration" sub-pages -----

  "admin-defaultprefs"() {
    return `
      <div class="page-head">
        <div><h1>Default preferences</h1><p>Manage preferences across an organization</p></div>
        <div class="page-head__actions"><button class="btn btn--primary" id="prefs-save">Save preferences</button></div>
      </div>
      <div class="admin-form">
        <div class="panel">
          <div class="panel__head"><span class="panel__title">Organization profile</span></div>
          <div class="panel__body">
            ${field("Organization name", "The display name of this organization", input("org-name", "Acme Ops"))}
          </div>
        </div>
        <div class="panel mt2">
          <div class="panel__head"><span class="panel__title">Preferences</span></div>
          <div class="panel__body">
            ${field("Home dashboard", "Default landing dashboard for this organization", select("home-dash", ["None", "Fleet overview", "OTA adoption", "Alert activity"], "Fleet overview"))}
            ${field("Timezone", "Used for dashboards and alert timestamps", select("prefs-tz", ["Browser time", "UTC", "Europe/London", "America/New_York"], "Browser time"))}
            ${field("Week start", "First day of the week", select("week-start", ["Monday", "Sunday", "Saturday"], "Monday"))}
            ${field("Default theme", "Appearance for new sessions", select("theme", ["Dark", "Light", "System"], "Dark"))}
          </div>
        </div>
      </div>`;
  },

  "admin-provisioning"() {
    return `
      <div class="page-head">
        <div><h1>Provisioning</h1><p>Manage your provisioning repositories and connections</p></div>
        <div class="page-head__actions"><button class="btn btn--primary">Connect repository</button></div>
      </div>
      <div class="prov-panel" data-panel="repos">
        <div class="panel"><div class="panel__body">
          <div class="empty-state">
            <div class="empty-state__icon">${plugIcon()}</div>
            <div class="empty-state__title">No repositories yet</div>
            <div class="empty-state__desc">Connect a Git repository to manage your Grafana resources as code.</div>
            <button class="btn btn--primary">Connect repository</button>
          </div>
        </div></div>
      </div>
      <div class="prov-panel" data-panel="connections" hidden>
        <div class="panel"><div class="panel__body" style="padding-top:8px">
          <table class="table"><thead><tr><th>Name</th><th>Type</th><th>Status</th><th></th></tr></thead>
          <tbody>
            <tr><td><b>github-acme</b></td><td>GitHub</td><td><span class="badge badge--ok">connected</span></td><td class="row-actions"><button class="btn btn--ghost tiny">Edit</button></td></tr>
          </tbody></table>
        </div></div>
      </div>
      <div class="prov-panel" data-panel="getting-started" hidden>
        <div class="panel"><div class="panel__body">
          <ol class="steps">
            <li><b>Connect a repository</b><span class="muted"> — link a Git provider account.</span></li>
            <li><b>Choose a branch</b><span class="muted"> — the source of truth for your resources.</span></li>
            <li><b>Set up a sync</b><span class="muted"> — pull changes on a schedule or via webhook.</span></li>
          </ol>
        </div></div>
      </div>`;
  },

  "admin-users"() {
    const USERS = [
      { login: "oscar.porter", email: "op@acme.io", name: "Oscar Porter", orgs: "Acme Ops", role: "Admin", last: "12:41", origin: "Grafana", prov: false },
      { login: "nina.grzelak", email: "nina@acme.io", name: "Nina Grzelak", orgs: "Acme Ops", role: "Editor", last: "2 hours ago", origin: "Google", prov: false },
      { login: "sam.okafor", email: "sam@acme.io", name: "Sam Okafor", orgs: "Acme Ops · Northwind", role: "Viewer", last: "Yesterday", origin: "GitHub", prov: true },
      { login: "lee.chan", email: "lee@globex.io", name: "Lee Chan", orgs: "Globex", role: "Viewer", last: "12 days ago", origin: "Grafana", prov: false },
      { login: "svc-ci", email: "—", name: "ci pipeline", orgs: "Acme Ops", role: "Editor", last: "Never", origin: "Service account", prov: true },
    ];
    return `
      <div class="page-head">
        <div><h1>Users</h1><p>Manage users in Grafana</p></div>
        <div class="page-head__actions"><button class="btn btn--primary">New user</button></div>
      </div>
      <div class="panel">
        <div class="log-toolbar">
          <div class="logs-search">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="7"/><line x1="21" y1="21" x2="16.7" y2="16.7"/></svg>
            <input class="logs-search__input" id="user-search" placeholder="Search users by login, email or name…" />
          </div>
          <div class="seg" id="user-role">
            <button class="seg__btn is-active" data-role="all">All</button>
            <button class="seg__btn" data-role="Admin">Admin</button>
            <button class="seg__btn" data-role="Editor">Editor</button>
            <button class="seg__btn" data-role="Viewer">Viewer</button>
          </div>
          <div class="topbar__spacer"></div>
          <span class="tiny muted" id="user-count">${USERS.length} users</span>
        </div>
        <div class="panel__body" style="padding-top:0">
          <table class="table">
            <thead><tr><th>Login</th><th>Email</th><th>Name</th><th>Belongs to</th><th>Role</th><th>Last active</th><th>Origin</th><th>Provisioned</th></tr></thead>
            <tbody id="user-rows">${USERS.map((u) => userRow(u)).join("")}</tbody>
          </table>
        </div>
      </div>`;
  },

  "admin-serviceaccounts"() {
    const SAS = [
      { name: "ci-deploy", id: 12, roles: "Editor", tokens: 2, created: "2026-06-01" },
      { name: "grafana-readonly", id: 9, roles: "Viewer", tokens: 1, created: "2026-07-14" },
      { name: "legacy-sync", id: 4, roles: "Editor, Viewer", tokens: 0, created: "2026-03-02" },
    ];
    return `
      <div class="page-head">
        <div><h1>Service accounts</h1><p>Use service accounts to run automated workloads in Grafana</p></div>
        <div class="page-head__actions"><button class="btn btn--primary">Add service account</button></div>
      </div>
      <div class="panel">
        <div class="log-toolbar">
          <div class="logs-search">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="7"/><line x1="21" y1="21" x2="16.7" y2="16.7"/></svg>
            <input class="logs-search__input" placeholder="Search service accounts…" />
          </div>
          <div class="topbar__spacer"></div>
          <span class="tiny muted">${SAS.length} service accounts</span>
        </div>
        <div class="panel__body" style="padding-top:0">
          <table class="table">
            <thead><tr><th></th><th>Account</th><th>ID</th><th>Roles</th><th>Tokens</th><th></th></tr></thead>
            <tbody>${SAS.map((s) => `<tr>
              <td><input type="checkbox" class="cb" /></td>
              <td><b>${s.name}</b><div class="tiny muted">created ${s.created}</div></td>
              <td class="mono">${s.id}</td>
              <td>${s.roles.split(",").map((r) => `<span class="badge scope">${r.trim()}</span>`).join(" ")}</td>
              <td>${s.tokens > 0 ? `<span class="badge badge--info">${s.tokens}</span>` : '<span class="muted">0</span>'}</td>
              <td class="row-actions"><button class="btn btn--ghost tiny">Edit</button><button class="btn btn--ghost tiny">Delete</button></td>
            </tr>`).join("")}</tbody>
          </table>
        </div>
      </div>`;
  },

  "admin-auth"() {
    const PROVIDERS = [
      ["Grafana.com", "Sign in with a Grafana.com / Grafana Cloud account", true],
      ["GitHub", "Use GitHub OAuth for authentication", false],
      ["GitLab", "Use GitLab OAuth for authentication", false],
      ["Google", "Use Google OAuth for authentication", true],
      ["Azure AD", "Use Azure Active Directory OAuth", false],
      ["Okta", "Use Okta OAuth for authentication", false],
      ["Generic OAuth", "Configure any OAuth2 provider", false],
      ["SAML", "Sign in via a SAML identity provider", false],
      ["LDAP", "Authenticate users against an LDAP directory", false],
    ];
    return `
      <div class="page-head">
        <div><h1>Authentication</h1><p>Manage your auth settings and configure single sign-on</p></div>
      </div>
      <div class="provider-grid">
        ${PROVIDERS.map(([name, desc, on]) => `<div class="panel provider-card">
          <div class="provider-card__icon">${lockIcon()}</div>
          <div class="provider-card__name">${name}</div>
          <div class="provider-card__desc">${desc}</div>
          <div class="provider-card__foot">
            ${on ? '<span class="badge badge--ok">Enabled</span>' : '<span class="badge badge--info">Not enabled</span>'}
            <button class="btn btn--ghost tiny">Configure</button>
          </div>
        </div>`).join("")}
      </div>`;
  },

  settings() {
    const sections = [
      ["general", "General", "Instance identity, region and data retention"],
      ["connection", "Connection", "Server endpoint, WebSocket and polling"],
      ["access", "Operators & Access", "Accounts, MFA and session policy"],
      ["alerting", "Alerting", "Thresholds that drive the alert engine"],
      ["notifications", "Notifications", "FCM, webhook and email delivery"],
      ["danger", "Danger zone", "Destructive fleet-wide actions"],
    ];
    const nav = sections.map(([id, label, desc], i) =>
      `<button class="set-nav__item${i === 0 ? " is-active" : ""}" data-sec="${id}">
        <span class="set-nav__label">${label}</span><span class="set-nav__desc">${desc}</span></button>`).join("");

    const body = {
      general: `
        ${field("Instance name", "Shown in the top bar and notification emails", input("instance-name", "Acme Vyzorix"))}
        ${field("Region", "Primary data region — cannot be changed after provisioning", select("region", ["eu-west-1", "us-east-1", "ap-south-1"], "eu-west-1"))}
        ${field("Timezone", "Used for timestamps in the audit log", select("tz", ["UTC", "Europe/London", "America/New_York"], "UTC"))}
        ${field("Telemetry retention", "How long device telemetry frames are kept", select("retention", ["7 days", "30 days", "90 days", "1 year"], "30 days"))}`,
      connection: `
        ${field("Server URL", "GraphQL + REST base endpoint (read-only)", `<code class="code">https://api.vyzorix.io</code>`)}
        ${field("WebSocket", "Real-time telemetry + command channel", `<span class="badge badge--ok">connected</span> <span class="tiny muted">wss://api.vyzorix.io/ws</span>`)}
        ${field("Auto-reconnect", "Reconnect the WebSocket automatically on drop", switchEl("autoreconnect", true))}
        ${field("Poll interval", "Fallback REST polling when the socket is down", select("poll", ["5s", "15s", "30s", "60s"], "15s"))}`,
      access: `
        ${field("Operator account", "Your login identity", `<span>op@acme.io</span> <span class="badge badge--info">Owner</span>`)}
        ${field("Multi-factor auth", "Require TOTP on every operator login", switchEl("mfa", true))}
        ${field("Google OAuth", "Link your Google workspace identity", `<span class="badge badge--ok">linked</span> <button class="btn btn--ghost">Unlink</button>`)}
        ${field("Session timeout", "Sign operators out after inactivity", select("session", ["1 hour", "8 hours", "24 hours"], "8 hours"))}
        ${field("IP allow-list", "Restrict console access to these CIDRs", `<input class="input" value="10.0.0.0/8, 203.0.113.0/24" />`)}`,
      alerting: `
        ${field("Battery low", "Fire an alert when a device drops below this level", `<input class="input input--sm" value="10" /> <span class="muted">%</span>`)}
        ${field("Offline after", "Mark a device offline after this silence", `<input class="input input--sm" value="15" /> <span class="muted">min</span>`)}
        ${field("Risk score", "Fire when the device risk score exceeds", `<input class="input input--sm" value="30" />`)}
        ${field("Notify on resolve", "Send a notification when an alert resolves", switchEl("onresolve", false))}`,
      notifications: `
        ${field("FCM push", "Push alerts to the VyzorixAudioRouter app via Firebase", switchEl("fcm", true))}
        ${field("Webhook", "POST alert payloads to your endpoint", `<input class="input" value="https://hooks.acme.io/vyzorix" /> <button class="btn btn--ghost">Test</button>`)}
        ${field("Email digest", "Summary of open alerts and fleet health", select("digest", ["Off", "Daily", "Weekly"], "Daily"))}
        ${field("Escalation contact", "Secondary contact for critical alerts", input("escalation", "oncall@acme.io"))}`,
      danger: `
        <div class="field field--danger">${fieldRow("Deregister all devices", "Remove every device from this organization. Recordings and telemetry are deleted.", `<button class="btn btn--danger">Deregister all</button>`)}</div>
        <div class="field field--danger">${fieldRow("Rotate all API keys", "Immediately revoke and re-issue every service-account key.", `<button class="btn btn--danger">Rotate keys</button>`)}</div>
        <div class="field field--danger">${fieldRow("Delete organization", "Permanently delete this organization and all its data. This cannot be undone.", `<button class="btn btn--danger">Delete organization</button>`)}</div>`,
    };

    return `
      <div class="page-head">
        <div><h1>Settings</h1><p>Instance, connection, access, alerting &amp; notifications</p></div>
        <div class="page-head__actions">
          <button class="btn btn--ghost" id="set-discard">Discard</button>
          <button class="btn btn--primary" id="set-save">Save changes</button>
        </div>
      </div>
      <div class="settings">
        <div class="set-body" id="set-body">
          ${sections.map(([id, label], i) => `<section class="set-section${i === 0 ? " is-active" : ""}" data-sec="${id}"><h2 class="set-h">${label}</h2>${body[id]}</section>`).join("")}
        </div>
      </div>`;
  },
};

function input(id, value) { return `<input class="input" id="${id}" value="${value}" />`; }
function select(id, opts, value) {
  return `<select class="select" id="${id}">${opts.map((o) => `<option${o === value ? " selected" : ""}>${o}</option>`).join("")}</select>`;
}
function switchEl(id, on) {
  return `<button class="switch${on ? " is-on" : ""}" id="${id}" role="switch" aria-checked="${on}"><span class="switch__dot"></span></button>`;
}

// Grafana's selectable themes (registry.ts defaults + getSelectableThemes extras).
function themeSelect() {
  const builtIn = ["System preference", "Dark", "Light"];
  const extras = [
    "Deuteranopia/Protanopia Dark", "Deuteranopia/Protanopia Light", "Tritanopia Dark", "Tritanopia Light",
    "Desert bloom", "Gilded grove", "Gloom", "Sapphire dusk", "Tron",
  ];
  return `<select class="select" id="pref-theme">
    <option>Default</option>
    <optgroup label="Built-in">${builtIn.map((t) => `<option>${t}</option>`).join("")}</optgroup>
    <optgroup label="Experimental">${extras.map((t) => `<option>${t}</option>`).join("")}</optgroup>
  </select>`;
}
function fieldRow(label, desc, control) {
  return `<div class="field__label"><div class="field__title">${label}</div><div class="field__desc">${desc}</div></div><div class="field__control">${control}</div>`;
}
function field(label, desc, control) { return `<div class="field">${fieldRow(label, desc, control)}</div>`; }

function devicesTable(verbose) {
  return `<table class="table">
    <thead><tr><th>Status</th><th>IMEI</th><th>Model</th><th>Firmware</th><th>Battery</th><th>Thermal</th><th>Risk</th>${verbose ? "<th>Org</th>" : ""}<th></th></tr></thead>
    <tbody>${DEVICES.map(d => `<tr>
      <td>${statusDot(d.status)} ${d.status}</td>
      <td class="mono">${d.imei}</td>
      <td>${d.model}</td>
      <td class="mono">${d.fw}${d.fw !== "0.4.2" ? ' <span class="badge badge--warn">update</span>' : ""}</td>
      <td style="min-width:120px">${bar(d.battery, d.battery < 15 ? "err" : d.battery < 40 ? "warn" : "ok")}<span class="tiny muted">${d.battery}%</span></td>
      <td>${d.thermal}</td>
      <td>${riskBadge(d.risk)}</td>
      ${verbose ? `<td>${d.org}</td>` : ""}
      <td><button class="btn btn--ghost tiny">Inspect</button></td>
    </tr>`).join("")}</tbody></table>`;
}

function cmdBadge(s) {
  const map = { delivered: "badge--info", acked: "badge--ok", pending: "badge--warn", failed: "badge--err", cancelled: "" };
  return `<span class="badge ${map[s] ?? ""}">${s}</span>`;
}

function eyeIcon() {
  return `<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-7 11-7 11 7 11 7-4 7-11 7-11-7-11-7z"/><circle cx="12" cy="12" r="3"/></svg>`;
}

function userRow(u) {
  const roleTone = u.role === "Admin" ? "badge--err" : u.role === "Editor" ? "badge--warn" : "badge--info";
  return `<tr data-role="${u.role}" data-text="${(u.login + " " + u.email + " " + u.name).toLowerCase()}">
    <td><b>${u.login}</b></td>
    <td class="muted">${u.email}</td>
    <td>${u.name}</td>
    <td class="muted">${u.orgs}</td>
    <td><span class="badge ${roleTone}">${u.role}</span></td>
    <td class="muted">${u.last}</td>
    <td>${u.origin}</td>
    <td>${u.prov ? '<span class="badge badge--info">yes</span>' : '<span class="muted">no</span>'}</td>
  </tr>`;
}
function plugIcon() { return `<svg width="34" height="34" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6"><path d="M9 2v6M15 2v6M6 8h12v4a6 6 0 0 1-12 0zM12 18v4"/></svg>`; }
function lockIcon() { return `<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="4" y="10" width="16" height="11" rx="2"/><path d="M8 10V7a4 4 0 0 1 8 0v3"/></svg>`; }

const LOG_LVL_CLASS = { error: "is-error", warn: "is-warn", info: "is-info", debug: "is-debug" };
function logRows(rows) {
  if (!rows.length) return `<div class="log-empty">No matching log entries.</div>`;
  return rows.map((l) => `<div class="log-row ${LOG_LVL_CLASS[l.lvl]}">
    <span class="log-time mono">${l.t}</span>
    <span class="log-lvl ${lvlBadge(l.lvl)}">${l.lvl}</span>
    <span class="log-src mono">${l.src}</span>
    <span class="log-msg">${l.msg}</span>
    <span class="log-req mono muted">${l.req}</span>
    <span class="log-dur mono muted">${l.dur}</span>
  </div>`).join("");
}

// ---------------------------------------------------------------- router
function render(view) {
  const fn = VIEWS[view] ?? VIEWS.dashboard;
  $("#content").innerHTML = fn();
  $("#crumb").textContent = {
    dashboard: "Dashboard", devices: "Devices", commands: "Commands", telemetry: "Telemetry",
    updates: "Updates", alerts: "Alerts", logs: "Logs", orgs: "Organizations", "api-keys": "API Keys", settings: "Settings", profile: "Profile",
    "admin-defaultprefs": "Administration › Default preferences",
    "admin-provisioning": "Administration › Provisioning",
    "admin-users": "Administration › Users",
    "admin-serviceaccounts": "Administration › Service accounts",
    "admin-auth": "Administration › Authentication",
  }[view] ?? "Dashboard";
  $$(".nav__item").forEach((b) => b.classList.toggle("is-active", b.dataset.view === view));
  $(".content").scrollTop = 0;
  mountCharts(view);
  renderPageTabs(view);
}

// Per-view in-page tabs shown in the third bar (tabbar).
const PAGE_TABS = {
  dashboard: [["overview", "Overview"]],
  devices: [["overview", "Overview"]],
  commands: [["overview", "Overview"]],
  telemetry: [["overview", "Overview"]],
  updates: [["overview", "Overview"]],
  alerts: [["overview", "Overview"]],
  logs: [["entries", "Entries"], ["volume", "Volume"]],
  orgs: [["overview", "Overview"]],
  "api-keys": [["overview", "Overview"]],
  settings: [["general", "General"], ["connection", "Connection"], ["access", "Operators & Access"], ["alerting", "Alerting"], ["notifications", "Notifications"], ["danger", "Danger zone"]],
  profile: [["info", "Profile"], ["preferences", "Preferences"], ["sessions", "Sessions"]],
  "admin-defaultprefs": [["overview", "Overview"]],
  "admin-provisioning": [["repos", "Repositories"], ["connections", "Connections"], ["getting-started", "Get started"]],
  "admin-users": [["overview", "Overview"]],
  "admin-serviceaccounts": [["overview", "Overview"]],
  "admin-auth": [["overview", "Overview"]],
};

function renderPageTabs(view) {
  const host = document.getElementById("page-tabs");
  if (!host) return;
  const tabs = PAGE_TABS[view] || [["overview", "Overview"]];
  host.innerHTML = tabs
    .map(([id, label], i) => `<button class="appbar__tab${i === 0 ? " is-active" : ""}" data-tab="${id}">${label}</button>`)
    .join("");
  host.querySelectorAll(".appbar__tab").forEach((b) => b.addEventListener("click", () => {
    host.querySelectorAll(".appbar__tab").forEach((x) => x.classList.toggle("is-active", x === b));
    activatePageTab(view, b.dataset.tab);
  }));
  // Ensure the first tab's panel is the visible one.
  activatePageTab(view, tabs[0][0]);
}

// Route a page tab (third bar) to the matching in-page panel.
function activatePageTab(view, tab) {
  const show = (panelSel, attr) => {
    const panels = document.querySelectorAll(panelSel);
    if (!panels.length) return false;
    panels.forEach((p) => { p.hidden = p.dataset[attr] !== tab; });
    return true;
  };
  const showClass = (sectionSel, attr) => {
    const sections = document.querySelectorAll(sectionSel);
    if (!sections.length) return false;
    sections.forEach((s) => s.classList.toggle("is-active", s.dataset[attr] === tab));
    return true;
  };
  if (view === "settings") return showClass(".set-section", "sec");
  if (view === "logs") {
    const ok = show(".log-tabpanel", "panel");
    if (ok && tab === "volume") {
      // Render the histogram once its panel is visible so it sizes correctly.
      const h = document.getElementById("chart-loghist");
      if (h && window.VyzCharts) window.VyzCharts.histogram(h, { height: 240, buckets: logHistogramBuckets() });
    }
    return ok;
  }
  if (view === "profile") return show(".profile-panel", "panel");
  if (view === "admin-provisioning") return show(".prov-panel", "panel");
}

// Shared histogram data (log volume) used by both the page tab and the initial mount.
function logHistogramBuckets() {
  const hist = [];
  for (let i = 20; i < 42; i++) {
    const b = { label: `12:${(i < 10 ? "0" : "") + i}` };
    b.info = i % 6 === 0 ? 3 : i % 3 === 0 ? 2 : 1;
    b.warn = i % 7 === 0 ? 2 : i % 5 === 0 ? 1 : 0;
    b.error = i === 35 || i === 38 ? 1 : 0;
    b.debug = i % 4 === 0 ? 1 : 0;
    hist.push(b);
  }
  return hist;
}

// Mount SVG charts into the freshly rendered view (after layout is available).
function mountCharts(view) {
  const C = window.VyzCharts;
  const host = (id) => document.getElementById(id);
  const t = (n) => Array.from({ length: n }, (_, i) => `${9 + Math.floor(i / 2)}:${i % 2 ? "30" : "00"}`);

  if (view === "dashboard") {
    if (host("chart-telemetry")) C.lineChart(host("chart-telemetry"), {
      height: 210, unit: "", legend: true,
      xLabels: ["12:00", "12:10", "12:20", "12:30", "12:40"],
      series: [
        { name: "Battery %", color: "#73bf69", values: [72, 70, 68, 66, 63, 61, 60, 59, 58, 57, 56, 55, 55, 54] },
        { name: "RSSI (dBm)", color: "#fb7185", values: [58, 60, 63, 61, 59, 57, 56, 58, 60, 62, 61, 59, 58, 57] },
        { name: "Risk", color: "#ff9830", values: [22, 24, 27, 30, 33, 36, 40, 43, 46, 49, 52, 55, 58, 61] },
      ],
    });
    if (host("chart-sev")) C.donutChart(host("chart-sev"), {
      height: 190, centerValue: "5", centerLabel: "active rules",
      segments: [
        { label: "Critical", value: 2, color: "#f2495c" },
        { label: "Warning", value: 2, color: "#ff9830" },
        { label: "Info", value: 1, color: "#fb7185" },
      ],
    });
  }

  if (view === "commands") {
    if (host("chart-cmd")) C.barChart(host("chart-cmd"), {
      height: 190, unit: "",
      data: [
        { label: "reboot", value: 18, color: "#fb7185" },
        { label: "start_cap", value: 12, color: "#73bf69" },
        { label: "stop_cap", value: 8, color: "#73bf69" },
        { label: "set_vol", value: 5, color: "#fade2a" },
        { label: "fw_update", value: 3, color: "#ff9830" },
        { label: "factory", value: 2, color: "#f2495c" },
      ],
    });
    if (host("chart-lat")) C.lineChart(host("chart-lat"), {
      height: 190, unit: "ms", xLabels: ["12:00", "12:10", "12:20", "12:30", "12:40"],
      series: [
        { name: "p50", color: "#fb7185", values: [280, 260, 250, 240, 230, 225, 220, 218, 222, 220, 219, 220] },
        { name: "p95", color: "#ff9830", values: [610, 590, 560, 540, 520, 500, 490, 480, 495, 480, 470, 465] },
      ],
    });
  }

  if (view === "telemetry") {
    if (host("chart-rssi")) C.lineChart(host("chart-rssi"), {
      height: 200, unit: " dBm", legend: false,
      xLabels: ["12:00", "12:10", "12:20", "12:30", "12:40"],
      series: [{ name: "RSSI", color: "#fb7185", values: [-70, -66, -62, -64, -60, -58, -59, -57, -55, -56, -54, -58, -60, -57] }],
    });
    if (host("chart-batt")) C.lineChart(host("chart-batt"), {
      height: 200, unit: "%", legend: false,
      xLabels: ["12:00", "12:10", "12:20", "12:30", "12:40"],
      series: [{ name: "Battery", color: "#73bf69", values: [92, 91, 90, 90, 89, 88, 88, 87, 87, 86, 86, 87, 87, 87] }],
    });
    if (host("chart-heat")) C.heatmap(host("chart-heat"), {
      rows: 4, cols: 12,
      rowLabels: ["451", "452", "453", "454"],
      colLabels: ["12:00", "", "12:10", "", "12:20", "", "12:30", "", "12:40", "", "12:50", ""],
      values: [
        [0.1, 0.15, 0.2, 0.25, 0.2, 0.3, 0.35, 0.3, 0.25, 0.2, 0.15, 0.1],
        [0.2, 0.3, 0.4, 0.5, 0.55, 0.6, 0.5, 0.45, 0.4, 0.35, 0.3, 0.25],
        [0.6, 0.7, 0.85, 0.9, 0.8, 0.7, 0.6, 0.5, 0.45, 0.4, 0.35, 0.3],
        [0.4, 0.5, 0.6, 0.7, 0.75, 0.8, 0.7, 0.6, 0.5, 0.45, 0.4, 0.35],
      ],
    });
  }

  if (view === "updates") {
    if (host("chart-rollout")) C.lineChart(host("chart-rollout"), {
      height: 200, unit: "%", xLabels: ["w1", "w2", "w3", "w4", "w5", "w6", "w7"],
      series: [
        { name: "stable 0.4.2", color: "#73bf69", values: [0, 5, 18, 34, 52, 66, 76] },
        { name: "beta 0.4.1", color: "#fb7185", values: [8, 12, 16, 18, 20, 19, 18] },
      ],
    });
    if (host("chart-channels")) C.barChart(host("chart-channels"), {
      height: 200, unit: "%",
      data: [
        { label: "stable", value: 76, color: "#73bf69" },
        { label: "beta", value: 18, color: "#fb7185" },
        { label: "canary", value: 6, color: "#ff9830" },
      ],
    });
  }

  if (view === "alerts") {
    if (host("chart-alerts")) C.barChart(host("chart-alerts"), {
      height: 200,
      data: [
        { label: "batt", value: 9, color: "#f2495c" },
        { label: "offline", value: 7, color: "#ff9830" },
        { label: "thermal", value: 6, color: "#fade2a" },
        { label: "risk", value: 5, color: "#fb7185" },
        { label: "fw", value: 4, color: "#73bf69" },
      ],
    });
    if (host("chart-mtta")) C.lineChart(host("chart-mtta"), {
      height: 200, unit: "m", xLabels: ["12:00", "12:10", "12:20", "12:30", "12:40"],
      series: [{ name: "MTTA", color: "#b877d9", values: [3, 2.8, 2.6, 2.4, 2.4, 2.3, 2.2, 2.2, 2.1, 2.1, 2.1, 2.1] }],
    });
  }

  if (view === "logs") {
    // Stacked log-volume histogram (error/warn/info/debug per minute).
    const mk = (i) => ({ label: `12:${(i < 10 ? "0" : "") + i}` });
    const hist = [];
    for (let i = 20; i < 42; i++) {
      const b = mk(i);
      b.info = i % 6 === 0 ? 3 : i % 3 === 0 ? 2 : 1;
      b.warn = i % 7 === 0 ? 2 : i % 5 === 0 ? 1 : 0;
      b.error = i === 35 || i === 38 ? 1 : 0;
      b.debug = i % 4 === 0 ? 1 : 0;
      hist.push(b);
    }
    const drawHist = () => { if (host("chart-loghist")) C.histogram(host("chart-loghist"), { height: 240, buckets: logHistogramBuckets() }); };

    // Live filter: level segmented control + text search.
    let level = "all";
    let query = "";
    const apply = () => {
      const rows = LOGS.filter((l) =>
        (level === "all" || l.lvl === level) &&
        (!query || (l.msg + " " + l.src + " " + l.req).toLowerCase().includes(query)),
      );
      const view = document.getElementById("log-view");
      if (view) view.innerHTML = logRows(rows);
      const count = document.getElementById("log-count");
      if (count) count.textContent = `${rows.length} of ${LOGS.length} entries`;
    };
    const levelBox = document.getElementById("log-levels");
    if (levelBox) levelBox.addEventListener("click", (e) => {
      const b = e.target.closest(".seg__btn");
      if (!b) return;
      level = b.dataset.lvl;
      [...levelBox.querySelectorAll(".seg__btn")].forEach((x) => x.classList.toggle("is-active", x === b));
      apply();
    });
    const search = document.getElementById("log-search");
    if (search) search.addEventListener("input", () => { query = search.value.trim().toLowerCase(); apply(); });
  }

  if (view === "api-keys") {
    if (host("chart-keyusage")) C.barChart(host("chart-keyusage"), {
      height: 220, unit: "",
      data: [
        { label: "ci-deploy", value: 6200, color: "#fb7185" },
        { label: "grafana", value: 5400, color: "#73bf69" },
        { label: "legacy", value: 800, color: "#5b5b5b" },
      ],
    });
    if (host("chart-auth")) C.donutChart(host("chart-auth"), {
      height: 200, centerValue: "99.4%", centerLabel: "success rate",
      segments: [
        { label: "Accepted", value: 12344, color: "#73bf69" },
        { label: "Rejected", value: 78, color: "#f2495c" },
      ],
    });
    // Reveal key ids on demand.
    document.querySelectorAll(".key-reveal").forEach((b) => b.addEventListener("click", () => {
      const span = b.previousElementSibling;
      const full = b.dataset.key;
      const masked = full.slice(0, 6) + "••••••••" + full.slice(-2);
      span.textContent = span.textContent.includes("•") ? full : masked;
    }));
  }

  if (view === "profile") {
    document.querySelectorAll(".switch").forEach((s) => s.addEventListener("click", () => {
      const on = s.classList.toggle("is-on");
      s.setAttribute("aria-checked", String(on));
    }));
    const save = document.getElementById("prefs-save");
    if (save) save.addEventListener("click", () => showToast("Preferences saved", "Your preferences were updated"));
    document.querySelectorAll(".session-revoke").forEach((b) => {
      if (b.classList.contains("is-current")) { b.disabled = true; b.style.opacity = "0.5"; b.style.pointerEvents = "none"; return; }
      b.addEventListener("click", () => {
        const tr = b.closest("tr");
        if (tr) tr.style.opacity = "0.4";
        b.disabled = true;
        showToast("Session revoked", "That signed-in device has been signed out");
      });
    });
  }

  if (view === "admin-users") {
    let role = "all", q = "";
    const apply = () => {
      let n = 0;
      document.querySelectorAll("#user-rows tr").forEach((tr) => {
        const ok = (role === "all" || tr.dataset.role === role) && (!q || tr.dataset.text.includes(q));
        tr.hidden = !ok; if (ok) n++;
      });
      const c = document.getElementById("user-count"); if (c) c.textContent = `${n} users`;
    };
    const seg = document.getElementById("user-role");
    if (seg) seg.addEventListener("click", (e) => {
      const b = e.target.closest(".seg__btn"); if (!b) return;
      role = b.dataset.role;
      [...seg.querySelectorAll(".seg__btn")].forEach((x) => x.classList.toggle("is-active", x === b));
      apply();
    });
    const s = document.getElementById("user-search");
    if (s) s.addEventListener("input", () => { q = s.value.trim().toLowerCase(); apply(); });
  }

  if (view === "admin-defaultprefs") {
    const save = document.getElementById("prefs-save");
    if (save) save.addEventListener("click", () => showToast("Preferences saved", "Organization preferences updated"));
  }

  if (view === "settings") {
    // Section switching is driven by the third-bar page tabs; just wire the switches.
    document.querySelectorAll(".switch").forEach((s) => s.addEventListener("click", () => {
      const on = s.classList.toggle("is-on");
      s.setAttribute("aria-checked", String(on));
    }));
    const save = document.getElementById("set-save");
    if (save) save.addEventListener("click", () => {
      const t = document.getElementById("toast");
      if (t) { t.querySelector("b").textContent = "Settings saved"; t.querySelector(".muted").textContent = "Organization preferences updated"; t.classList.add("is-on"); }
    });
  }
}

// Collapsible nav groups (Grafana Administration style).
document.querySelectorAll(".nav__group-head").forEach((h) => h.addEventListener("click", () => {
  const g = h.closest(".nav__group");
  const collapsed = g.classList.toggle("is-collapsed");
  h.setAttribute("aria-expanded", String(!collapsed));
}));

$$(".nav__item").forEach((b) =>
  b.addEventListener("click", () => {
    render(b.dataset.view);
    setMenu(false); // auto-hide the sidebar once a menu item is chosen
  }),
);
render("dashboard");

// Second-bar navigation tabs + global search.
$$(".appbar__tab").forEach((t) => t.addEventListener("click", () => render(t.dataset.view)));
const globalSearch = document.getElementById("global-search");
if (globalSearch) globalSearch.addEventListener("keydown", (e) => {
  if (e.key !== "Enter") return;
  const q = globalSearch.value.trim();
  if (!q) return;
  const v = /alert/i.test(q) ? "alerts" : /log/i.test(q) ? "logs" : /device|imei/i.test(q) ? "devices" : /update|ota/i.test(q) ? "updates" : "dashboard";
  render(v);
  showToast("Search", `Showing results for “${q}”`);
});

// Inject Grafana SVG icons into every [data-icon] placeholder.
if (window.VyzIcons) window.VyzIcons.mountIcons();

// Mega-menu overlay (Grafana-style auto-hiding drawer): open/close, no icon rail.
const chromeEl = $(".chrome");
const setMenu = (open) => chromeEl.classList.toggle("is-menu-open", open);
const isMenuOpen = () => chromeEl.classList.contains("is-menu-open");

$("#menu-toggle").addEventListener("click", () => setMenu(!isMenuOpen()));
$("#menu-close").addEventListener("click", () => setMenu(false));
$(".scrim").addEventListener("click", () => setMenu(false));
const customiseBtn = document.getElementById("menu-customise");
if (customiseBtn) customiseBtn.addEventListener("click", () => showToast("Customise navigation", "Reorder or hide menu items (mock)"));
document.addEventListener("keydown", (e) => {
  if (e.key === "Escape") { setMenu(false); setUserMenu(false); }
});

// Top-bar user menu (profile / preferences / sign out).
const userEl = document.getElementById("user");
const userBtn = document.getElementById("user-btn");
const userMenu = document.getElementById("user-menu");
const setUserMenu = (open) => {
  userMenu.hidden = !open;
  userBtn.setAttribute("aria-expanded", String(open));
};
userBtn.addEventListener("click", (e) => { e.stopPropagation(); setUserMenu(userMenu.hidden); });
document.addEventListener("click", (e) => { if (!userEl.contains(e.target)) setUserMenu(false); });
userMenu.addEventListener("click", (e) => {
  const item = e.target.closest(".user-menu__item");
  if (!item) return;
  setUserMenu(false);
  if (item.dataset.go) render(item.dataset.go);
  else if (item.dataset.action === "signout") showToast("Signed out", "You have been signed out (mock)");
});

function showToast(title, body) {
  const t = document.getElementById("toast");
  if (!t) return;
  t.querySelector("b").textContent = title;
  t.querySelector(".muted").textContent = body;
  t.classList.add("is-on");
  clearTimeout(showToast._t);
  showToast._t = setTimeout(() => t.classList.remove("is-on"), 2600);
}
