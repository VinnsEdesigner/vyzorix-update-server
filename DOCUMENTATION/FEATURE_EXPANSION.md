# Vyzorix Feature Expansion: Requirements & Design

8 features to implement, each better than Grafana's equivalent. No dead
functions, full wiring, minimal new files (merge into existing where
possible), tests, lint clean.

## 1. Quota System

**Grafana**: Multi-scope (global/org/user) quota with RegisterQuotaReporter.
**Vyzorix improvement**: Org-scoped quotas checked at the application layer
(not middleware), with per-resource-type limits enforced inline at creation
time. No separate quota DB table — quotas live on the org entity as a JSON
blob, queried once per request context.

**Requirements**:
- `QuotaConfig` struct on `Organization` entity (devices, operators, api_keys,
  commands_per_minute, pending_invitations).
- `QuotaService` in `application/` that checks limits before creation ops.
- Checked in: `DeviceService.Register`, `MemberService.AddMember`,
  `APIKeyService.GenerateKey`, `InvitationService.CreateInvitation`.
- No new table — `organization_settings` JSON blob or `organizations` columns.
- Config-driven defaults via env vars (`QUOTA_MAX_DEVICES`, etc.).

**Merge into**: `internal/domain/organization/organization_entity.go` (quota
config field), `internal/application/organization/` (quota check calls).

## 2. Tag System

**Grafana**: Simple `Tag` entity with `EnsureTagsExist`.
**Vyzorix improvement**: Tags are a JSON string array on the device entity
(`Tags []string`), not a separate table. No join tables, no tag entities —
just indexed JSON. Searchable via SQLite JSON functions.

**Requirements**:
- `Tags []string` field on `Device` entity (already has `Metadata` map — tags
  can be a dedicated field for query clarity).
- `DeviceService.SetTags` / `DeviceService.AddTag` / `DeviceService.RemoveTag`.
- Storage: JSON column in `devices` table (migration).
- Query: `WHERE tags LIKE '%"production"%'` or SQLite json_each.
- Merge into: `internal/domain/device/`, `internal/application/device/`.

## 3. Server Lock (Distributed Mutex)

**Grafana**: `pkg/infra/serverlock` — DB-backed advisory lock for single-instance
operations.
**Vyzorix improvement**: SQLite `advisory_lock` table with `acquired_at` +
`expires_at` + `holder` identity. Workers acquire a named lock before running;
release on completion; auto-expire after timeout. No external dependency
(Redis) — works on SQLite and Turso.

**Requirements**:
- `ServerLock` struct: name, holder, acquired_at, expires_at.
- `LockService.Acquire(name, holder, ttl)` / `LockService.Release(name, holder)`.
- Migration: `server_locks` table.
- Used by: `DeviceDeletionWorker`, `InvitationCleanupWorker`, `CommandOutbox`
  (so horizontal scaling doesn't double-execute).
- New file: `internal/infrastructure/serverlock/lock.go` (small, standalone).

## 4. Preferences (Per-Operator)

**Grafana**: `(org_id, user_id, team_id)` triple-keyed preferences table with
hot+cold fields and version-locked PATCH.
**Vyzorix improvement**: Merge into existing `operator_settings` table (already
exists). Add a `preferences` JSON column. No separate table — the operator
settings service already handles get/update. Add `PATCH /v1/me/preferences`.

**Requirements**:
- `Preferences` struct: theme, default_org, notification_frequency, dashboard
  refresh_interval, language, timezone.
- Stored as JSON in `operator_settings.preferences` column (migration).
- `OperatorSettingsService.GetPreferences` / `UpdatePreferences`.
- PATCH nil = leave unchanged, PATCH value = set, PATCH empty string = reset
  to default.
- Merge into: `internal/application/auth/` (operator settings service).

## 5. Support Bundles

**Grafana**: Collector pattern, tar archive, async generation.
**Vyzorix improvement**: Synchronous JSON export (not tar — simpler, API-native).
Collects: config (redacted), DB schema version, recent errors, worker status,
hub metrics, device count, operator count. Returns as a JSON download.

**Requirements**:
- `SupportBundleService` that collects diagnostic data from existing services.
- `GET /v1/admin/support-bundle` (super_admin only).
- Collectors: config, migrations, audit log tail, hub metrics, worker status,
  device/org/operator counts.
- No new table — reads from existing tables/services.
- New file: `internal/api/handlers/admin/support_bundle.go`.

## 6. Update Manager

**Grafana**: Periodic GitHub release check for the server binary + plugins.
**Vyzorix improvement**: Check the Vyzorix GitHub releases API for newer
versions than the running binary's version. Expose via
`GET /v1/admin/updates/check`. No background polling — on-demand only
(admin triggers the check).

**Requirements**:
- `UpdateChecker` that calls `https://api.github.com/repos/VinnsEdesigner/
  vyzorix-update-server/releases/latest`.
- Compares `version` (set via `-X link flag` at build) against latest release
  tag.
- `GET /v1/admin/updates/check` (super_admin).
- No new table — stateless.
- New file: `internal/api/handlers/admin/update_checker.go`.

## 7. Search Service

**Grafana**: Full-text search across dashboards/folders with tag + sort filters.
**Vyzorix improvement**: Cross-resource search (devices + commands + events +
updates) in a single query. Uses SQLite LIKE (no FTS5 dependency). Returns
typed results with resource type + snippet.

**Requirements**:
- `SearchService` that queries devices (by IMEI, name, model, tags), commands
  (by dispatch ID, command type, status), events (by type, device), updates
  (by version).
- `GET /v1/search?q=...&type=device|command|event|update|all&org_id=...`.
- Scoped RBAC: `search.read` on `org:*` (new action).
- Merge into: `internal/application/` (new `search/` package).
- New file: `internal/application/search/search_service.go`.

## 8. Provisioning (Declarative YAML at Boot)

**Grafana**: Reads YAML files at boot, creates resources.
**Vyzorix improvement**: Single `provisioning.yaml` file read at boot that
declares orgs, operators, device groups, API keys, and permission grants.
Idempotent — skips resources that already exist. Runs before the HTTP server
starts.

**Requirements**:
- `provisioning.yaml` schema: orgs, operators, device_groups, api_keys,
  permission_grants.
- `ProvisioningService` that reads the file and creates resources via existing
  application services.
- Runs in `api_main.go` after Wire but before `ListenAndServe`.
- `PROVISIONING_FILE` env var (path to YAML, default: `provisioning.yaml`).
- New file: `internal/application/provisioning/provisioning.go`.

## Implementation order

1. Quota (domain + checks)
2. Tags (device entity + migration)
3. Server lock (infra + worker integration)
4. Preferences (operator settings merge)
5. Support bundle (admin handler)
6. Update manager (admin handler)
7. Search (application service + route)
8. Provisioning (boot-time YAML reader)

## Tests

Each feature gets:
- Domain unit tests (pure logic)
- Service integration tests (if touching DB)
- No mocks for real code paths

## Lint

golangci-lint v2.12.2, 0 issues required before push.
