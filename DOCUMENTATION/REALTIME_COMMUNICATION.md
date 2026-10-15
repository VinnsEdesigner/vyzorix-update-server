# Realtime Communication

The server maintains persistent WebSocket connections with devices and pushes data to dashboards in realtime. This covers how connections work, what flows over them, and how the server handles disconnections.

## WebSocket connections

Devices connect via `GET /v1/device/:imei/stream`. The connection is upgraded to WebSocket and registered with the `Hub` — a central connection manager in `internal/ws/hub.go`.

The Hub tracks which devices are online via a map of `deviceID → *Client`. When a device connects, it calls `h.Register(client)`. When it disconnects, the client's `close` channel fires and the Hub removes it.

### Authentication

WebSocket connections require HMAC authentication (unless `ENFORCE_HMAC=false` in development). The device signs the connection request with its `CommandSecretHash`, and the `SessionSignatureMiddleware` verifies it.

### Online status

`hub.Online(deviceID)` returns true if the device has an active WebSocket connection. This is used by the command execution path to decide whether to send via WebSocket or FCM.

## Message routing

The `MessageRouter` in `internal/api/handlers/websocket/stream_message.go` handles incoming messages from devices. It reads the message type and routes:

- **telemetry** — device sends telemetry data (battery, CPU, memory, temperature). Stored in the telemetry table.
- **command_result** — device reports the result of a command execution (success/failure).
- **subscribe** / **unsubscribe** — dashboard clients subscribe to receive realtime updates for specific devices or all devices.

## Dashboard broadcasting

Dashboard clients (operators) connect over WebSocket and monitor devices in realtime. There are two transports:

- The device stream (`/v1/device/:imei/stream`), where the device itself connects. Inbound telemetry is broadcast to dashboards via `hub.BroadcastTelemetry`.
- The GraphQL subscription endpoint (`/:org/graphql/ws`), where dashboards subscribe to `deviceUpdated`, `telemetryReceived` and `commandStatusChanged`.

GraphQL subscription fan-out is keyed by organization, not by operator: a device going offline, a telemetry frame, or a command reaching `delivered` is published to every subscriber in the device's organization. The Hub resolves a device's organization once, when the device connects, and caches it on the client so the telemetry path needs no per-frame lookup. See [GRAPHQL_API.md](GRAPHQL_API.md) for the subscription-level authorization rules.

## Rate limiting

WebSocket messages are rate-limited per device to prevent a single device from flooding the server. The Hub's `RateLimiter` (set via `h.SetRateLimiter`) enforces a configurable message rate. Messages exceeding the rate are dropped with a warning log.

## Command delivery over WebSocket

When the command execution path sends a command to an online device, it calls `hub.Send(deviceID, frame)`. This pushes the signed `CommandFrame` as JSON to the device's WebSocket connection. If the send fails (connection closed mid-send), the command stays in `pending` and the outbox retries later.

There's also `hub.SendWithDeliveryConfirmation` which waits for the device to acknowledge receipt within a timeout. Used for high-priority commands.

## FCM fallback

If a device is offline (no WebSocket), the server sends a Firebase Cloud Messaging (FCM) silent push notification. The notification tells the device app to wake up, connect to the server, and fetch pending commands.

The FCM notifier in `internal/infrastructure/fcm/notifier.go` has a circuit breaker — if FCM delivery keeps failing, it stops trying temporarily to avoid blocking the command path.

The FCM retry worker (`internal/infrastructure/worker/fcm_retry_worker.go`) periodically retries failed FCM deliveries.

## Telemetry

Telemetry frames from devices contain:

- Battery level and charging state
- CPU and memory usage
- Temperature
- Network type (WiFi/cellular)
- App version
- Custom metrics (device-specific)

These are stored in the `device_telemetry` table and queryable via `GET /v1/devices/:imei/telemetry`. The dashboard shows them in realtime via the WebSocket broadcast.

Telemetry also feeds into threshold alerts — if a device's battery drops below a configured threshold, the server generates a device event.
