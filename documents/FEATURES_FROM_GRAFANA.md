# Vyzorix Feature Expansion — Lessons from Grafana

Source analysis: head-to-head code inspection of `grafana/pkg/services` + `grafana/pkg/infra`
against `vyzorix-update-server/apps/api/internal` (2026-08-20).

Grafana is an observability platform; Vyzorix is an MDM/OTA update server. The 11 features
below are patterns Grafana has battle-tested that map onto real Vyzorix use cases. Each
entry lists the Grafana reference package, the Vyzorix gap, and the planned integration
point. They are ordered by implementation priority.

## Status legend

- [ ] not started
- [~] in progress
- [x] implemented

---

## Tier 1 — Highest value

### 1. Alerting engine [x]

- **Grafana reference:** `pkg/services/ngalert` (rule store, tick-based evaluator, alert
  state machine, notifier)
- **Vyzorix gap:** nothing. No alert rules, no evaluation loop, no notifications.
- **Vyzorix use cases:** "device offline count > N", "offline share of fleet > X%",
  "command/update failure rate > Y% in window".
- **Plan (lean, DDD-native — not an Alertmanager port):**
  - `internal/domain/alert/` — `Rule` entity (org-scoped, metric, condition, threshold,
    pending duration, optional webhook URL), `Instance` state machine
    (inactive → pending → firing → resolved), `Repository` / `StateRepository` /
    `Notifier` ports.
  - `internal/infrastructure/storage/` — migration 64 (`alert_rules`,
    `alert_instances`) + SQL repositories.
  - `internal/application/alert/` — CRUD service, `MetricSource` (org-scoped SQL over
    devices/commands), `Evaluator` (advances state machine, emits notifications on
    transitions).
  - `internal/infrastructure/webhook/` — notifier adapter over the existing
    `webhook.Client` (SSRF-safe, HMAC-signed, retrying).
  - `internal/infrastructure/worker/` — tick-based evaluation worker, guarded by the
    existing `serverlock.Service` (same as DeviceDeletionWorker).
  - REST: `/v1/alerts/rules` CRUD + manual evaluate; scoped RBAC actions
    `alert.read` / `alert.write` on `org:*`.
- **Status:** implemented.

### 2. Real Prometheus instrumentation [x]

- **Grafana reference:** `pkg/infra/metrics`, `pkg/middleware/request_metrics.go`
- **Vyzorix gap:** `internal/infrastructure/metrics/prometheus.go` is hand-rolled
  atomic counters — no labels, no histograms, not the real `prometheus/client_golang`.
- **Plan:** swap to `client_golang`; `HistogramVec` for HTTP request duration labeled
  by route/method/status; counter for command-delivery latency; expose `/metrics`.
- **Why first after alerting:** you cannot alert on what you cannot measure.

### 3. Notification channels (contact points) [x]

- **Grafana reference:** `pkg/services/notifications`, `ngalert/notifier`
- **Vyzorix gap:** only `infrastructure/webhook/webhook_client.go` and audit entries.
  No Slack/Discord/email-style channel abstraction, no routing, no templates.
- **Plan:** `ContactPoint` entity (type + encrypted config), dispatcher service,
  message templates; becomes the delivery arm of Feature 1 (rules reference contact
  points instead of a single webhook URL).

## Tier 2 — Strong wins

### 4. Service accounts [x]

- **Grafana reference:** `pkg/services/serviceaccounts` (incl. `secretscan`)
- **Vyzorix gap:** API keys hang off operator identities.
- **Plan:** dedicated `service_accounts` entity with scoped tokens, rotation endpoint;
  steal the `secretscan` idea — scan outbound payloads for leaked key patterns.

### 5. Annotations [ ]

- **Grafana reference:** `pkg/services/annotations`
- **Vyzorix gap:** no way to mark "firmware v2.3 rollout started" on the fleet
  timeline and correlate with failure spikes.
- **Plan:** small org-scoped entity (time range + tags + text), repo, routes;
  rendered on dashboard timelines.

### 6. Config versioning [ ]

- **Grafana reference:** `pkg/services/dashboardversion` (versioned saves, diff,
  restore)
- **Vyzorix gap:** dashboard service returns aggregates only; no versioned, restorable
  definitions of org settings or update-campaign configs.
- **Plan:** version table per config resource (who/what/when), restore endpoint;
  bootstrap from existing audit events.

### 7. Query/response caching [ ]

- **Grafana reference:** `pkg/services/caching`, `pkg/infra/localcache`
- **Vyzorix gap:** caching limited to the permission cache and middleware internals.
- **Plan:** generic caching service with TTL + hit-rate metrics; apply to
  `/v1/search`, dashboard stats, telemetry-history queries.

## Tier 3 — Adopt when scale demands it

### 8. Envelope encryption + KMS [ ]

- **Grafana reference:** `pkg/services/encryption`, `pkg/services/kmsproviders`
- **Vyzorix gap:** secrets at rest not envelope-encrypted.
- **Plan:** data-key/envelope-key model for device/app secrets, master key from env,
  pluggable KMS later.

### 9. Background-service lifecycle registry [ ]

- **Grafana reference:** `pkg/modules`, `pkg/registry` (BackgroundService with
  dependency-ordered start/stop + health)
- **Vyzorix gap:** workers wired ad-hoc in `api_main.go`.
- **Plan:** extend `domain/lifecycle` into a manager: named services, ordered
  graceful shutdown, per-service health for `/healthz`.

### 10. Usage stats service [ ]

- **Grafana reference:** `pkg/infra/usagestats`
- **Plan:** periodic self-telemetry (version, feature-toggle state, entity counts);
  feeds the existing admin update-checker.

### 11. Org-scoped live channels [ ]

- **Grafana reference:** `pkg/services/live` (`orgchannel`, `managedstream`)
- **Vyzorix state:** `internal/ws` hub already has compression, rate limiting,
  telemetry filters — ahead of Grafana Live for the device use case.
- **Plan:** adopt only the channel addressing + subscribe-permission pattern:
  `stream/<org>/<scope>` checked through the existing `command.Authorizer`; server-
  pushed managed streams for alert notifications (pairs with Feature 1).

## Already at parity (no action)

- Server lock / distributed mutex (stolen already), feature toggles, scoped RBAC with
  wildcard scopes, boot provisioning (YAML/JSON), quotas, support bundle, update
  checker.
- Follow-up to provisioning: once Features 1+3 land, extend
  `provisioning.example.yaml` to declare alert rules and contact points (mirrors
  Grafana's alerting provisioning).
