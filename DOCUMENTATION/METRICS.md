# Metrics

The server exposes Prometheus metrics at `/metrics`. Everything that matters to
running the thing in production — request rates and latencies, login outcomes,
command throughput, session counts, feature-toggle state — is a labeled series
you can graph, alert on, and correlate.

It's the standard `prometheus/client_golang` registry. Point a scraper at
`/metrics` and you're done; there's no agent to install and no push gateway.

---

## HTTP

| Metric | Type | Labels | What it tells you |
|:-------|:-----|:-------|:------------------|
| `http_requests_total` | Counter | `route`, `method`, `status` | Request volume and error rate, per route. |
| `http_request_duration_seconds` | Histogram | `route`, `method`, `status` | Latency distribution, per route. |
| `http_requests_in_flight` | Gauge | — | Requests currently being served. |

`route` is the templated path (`/v1/devices/:imei`), not the raw URL, so the
series count stays bounded no matter how many device IDs you have.

---

## Authentication and sessions

| Metric | Type | Labels | What it tells you |
|:-------|:-----|:-------|:------------------|
| `login_attempts_total` | Counter | `result` | Login volume. |
| `login_failures_total` | Counter | `reason` | Why logins fail — bad password, MFA, lockout. |
| `security_failures_total` | Counter | `type` | Rejected signatures, CSRF failures, and similar. |
| `active_sessions` | Gauge | — | Sessions currently live. |
| `session_revocations_total` | Counter | — | Sessions revoked, including revoke-all. |

A spike in `login_failures_total` with a flat `login_attempts_total` is someone
guessing. A spike in `security_failures_total` is someone probing the signing
layer.

---

## Commands

| Metric | Type | Labels | What it tells you |
|:-------|:-----|:-------|:------------------|
| `commands_total` | Counter | `result` | Commands dispatched, by outcome. |
| `command_delivery_duration_seconds` | Histogram | — | Time from dispatch to acknowledgement. |
| `command_queue_size` | Gauge | — | Commands waiting in the outbox. |

A growing `command_queue_size` means devices aren't acknowledging — either
they're offline or the delivery path is backed up.

---

## Platform

| Metric | Type | Labels | What it tells you |
|:-------|:-----|:-------|:------------------|
| `alert_rules_active` | Gauge | — | Enabled alert rules across the deployment. |
| `feature_toggle_enabled` | Gauge | `feature` | Which feature toggles are on. |
| `uptime_seconds` | Gauge | — | Process uptime. |
| `sql_connections_total` | Counter | — | Database connections opened. |

---

## Cache

| Metric | Type | Labels | What it tells you |
|:-------|:-----|:-------|:------------------|
| `cache_operations_total` | Counter | `section`, `outcome` | Cache hits and misses, per section. |

See [Caching](CACHING.md) for what the sections mean.

---

## Scraping

```
GET /metrics
```

Standard Prometheus text exposition. The endpoint sits outside the authenticated
routes — scrapers don't carry user credentials, and the SSR proxy treats
`/metrics` like `/health`, so it is served directly rather than redirected to
`/login`. Put it behind your network boundary, not behind the app's auth.

A minimal scrape config:

```yaml
scrape_configs:
  - job_name: vyzorix
    metrics_path: /metrics
    static_configs:
      - targets: ["vyzorix-server:3000"]
```

---

## Why these metrics and not more

Every metric here answers an operational question: is it up, is it fast, is it
failing, is anyone attacking it. Labels are kept low-cardinality on purpose —
routes are templated, not raw — so the series count stays flat as the fleet
grows. Metrics that would explode in cardinality (per-device, per-organization)
are deliberately left to the database and the dashboard, where querying them is
cheap and the answer is a number, not a time series.

## Profiling (pprof)

Metrics tell you *that* something is wrong; pprof tells you *where*. The Go
runtime profilers are exposed under `/v1/admin/pprof` when `ENABLE_PPROF=true`:

| Endpoint | Use |
|:---------|:----|
| `/v1/admin/pprof/heap` | In-use heap - what is retaining memory |
| `/v1/admin/pprof/allocs` | Allocation sites - what churns |
| `/v1/admin/pprof/goroutine?debug=1` | Every goroutine's stack - leaks, blocked work |
| `/v1/admin/pprof/profile?seconds=30` | 30s CPU profile |
| `/v1/admin/pprof/trace?seconds=5` | Execution trace |
| `/v1/admin/pprof/block`, `/mutex` | Contention |

Profiling is **off by default** and, when enabled, sits behind the same cookie
auth + super-admin check as the other `/admin` routes, so a profile is never
reachable anonymously. A heap or CPU profile is a read of process memory, so
treat the endpoint as sensitive and turn it off when you are done.

```bash
# With a super-admin session cookie and the organization selected:
go tool pprof -top http://localhost:3000/v1/admin/pprof/heap
go tool pprof -top -seconds=30 http://localhost:3000/v1/admin/pprof/profile
curl -b cookies.txt 'http://localhost:3000/v1/admin/pprof/goroutine?debug=1'
```

The endpoints are behind auth, so `go tool pprof` needs the cookie (pass
`-header 'Cookie: vyz_session=...'`), or capture with `curl` first and run pprof
against the saved file.
