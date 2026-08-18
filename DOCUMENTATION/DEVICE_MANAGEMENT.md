# Device Management

Devices are the core entity. A device is an Android phone running the Vyzorix agent. The server manages registration, status, command delivery, and deregistration.

## Device registration

Devices register via `POST /v1/device/register` (public, HMAC-signed). The request includes:

```json
{
  "deviceId": "356938035643809",
  "firebaseInstallId": "...",
  "fcmToken": "...",
  "appVersion": "1.2.3",
  "deviceClass": "samsung SM-A505GN"
}
```

The `DeviceService.Register` method in `internal/application/device/device_service.go` handles this. It generates a `CommandSecretHash` (SHA-256 of a random secret) that's used for server→device command signing. The device gets back a `clientID` and `accessToken`.

Devices go through lifecycle states: `pending` → `registered` → `deregistered`. The `Lifecycle` type in `internal/domain/device/device_entity.go` manages transitions. Deregistered devices are soft-deleted (the `DeletionScheduledAt` field is set) and cleaned up by the `DeviceDeletionWorker` after a 30-day retention period.

## Device status

`GET /v1/device/:imei` returns the device's current state: online/offline, last seen, battery, model, OS version, etc. Online status comes from the WebSocket hub — if a device has an active WebSocket connection, it's online.

## Device list

`GET /v1/devices` returns a paginated list of devices for the current organization. The `DevicesHandler` in `internal/api/handlers/device/devices_handler.go` handles this. It's org-scoped — you only see devices in your organization.

## Device deregistration

`DELETE /v1/device/:imei` soft-deletes a device. The device is marked as deregistered, its WebSocket connection is closed, and its data is scheduled for permanent deletion after 30 days.

## Multi-tenancy

Every device belongs to an organization (`OrganizationID` field). The organization context is set by the `NewOrganizationContext` middleware, which extracts the org ID from the URL or session. The `NewOrganizationMembership` middleware verifies the operator is a member of that org.

Handlers call `middleware.GetOrganizationID(c)` to get the org ID and scope their queries accordingly.

## Device settings

Each device has per-device settings stored in the `device_settings` table:

- Custom name (operator-assigned label)
- Location (operator-assigned)
- Metadata (key-value pairs)
- Thresholds (telemetry alert thresholds — battery, temperature, etc.)

Managed via `GET/PATCH /v1/devices/:imei/settings` and `GET/PATCH /v1/devices/:imei/settings/thresholds`.

## Device events

Devices emit events (status changes, telemetry alerts, command results) that are stored in the `device_events` table. These are queryable via `GET /v1/devices/:imei/events` with filtering by event type and time range.
