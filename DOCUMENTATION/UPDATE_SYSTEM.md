# Update System

The update system manages APK version distribution and push delivery to devices. It integrates with GitHub releases for automatic version sync.

## Update versions

APK versions are stored in the `update_versions` table. Each version has:

- Version string (e.g., "1.2.3")
- APK filename and size
- SHA256 hash
- Release date
- Release notes
- Release type (stable, beta, internal)
- `is_latest` flag (only one version can be latest at a time)

Managed via `GET /v1/updates/versions` (list), `POST /v1/updates/versions` (create), and `DELETE /v1/updates/versions/:id` (delete).

## GitHub webhook sync

The server accepts GitHub webhook events at `POST /v1/updates/webhook`. When a new release is published on GitHub, the webhook fires and the server:

1. Verifies the webhook signature (using `GITHUB_WEBHOOK_SECRET`)
2. Parses the release payload
3. Creates `update_versions` entries for each APK asset in the release
4. Marks the latest version

The handler is in `internal/api/handlers/updates/github_webhook_handler.go`. It ignores non-release events.

## Update push

`POST /v1/updates/push` initiates an update push to a set of devices. The request specifies:

- Version ID (which APK to push)
- Organization ID
- Install type (immediate, scheduled, deferred)
- Device list (or "all" for the org)

The `PushService.PushUpdate` in `internal/application/updates/updates_push_service.go` creates an `update_pushes` record and enqueues commands to each device telling it to download and install the update.

The push status tracks: `in_progress` → `completed` / `failed` / `cancelled`.

## Push status tracking

`GET /v1/updates/pushes/:id/status` returns the current state of a push — how many devices received it, how many succeeded, how many failed.

Devices report their update status back via `POST /v1/updates/device-status` (HMAC-authenticated device endpoint). The status can be: `in_progress`, `completed`, or `failed`.

## Update sync

`POST /v1/updates/sync` triggers a manual sync with GitHub releases. The server fetches all releases from the configured GitHub repository and creates/updates version entries. This is the same logic the webhook uses, but triggered manually.

## Update history

`GET /v1/updates/history` returns a paginated history of update pushes with filtering by status, date range, and version.

## Audit integration

Update pushes and syncs generate audit entries:
- `update_pushed` — when a push is initiated (includes version, device count)
- `update_cancelled` — when a push is cancelled
- `update_sync_started` — when a GitHub sync begins
- `update_sync_failed` — when a sync fails

These go through the `Logger` methods (`UpdatePushed`, `UpdateCancelled`, `UpdateSyncStarted`, `UpdateSyncFailed`) and land in the `audit_logs` table.
