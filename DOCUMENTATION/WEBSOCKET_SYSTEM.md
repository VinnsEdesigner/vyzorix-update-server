# WebSocket System

The WebSocket layer is the realtime backbone of Vyzorix. Devices use it to receive commands and send telemetry. Dashboards use it to monitor devices live. This document covers the Hub, clients, message routing, the message queue, rate limiting, and how everything fits together.

## The Hub

The `Hub` in `internal/ws/hub.go` is a singleton that manages all WebSocket connections. It runs a single goroutine (`hub.Run`) that processes events from four channels:

```go
func (h *Hub) Run(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case c := <-h.register:      // new client connecting
            h.handleClientRegistration(ctx, c)
        case c := <-h.unreg:          // client disconnecting
            h.handleClientUnregistration(ctx, c)
        case raw := <-h.broadcast:   // broadcast to all dashboard clients
            h.handleBroadcast(raw)
        case status := <-h.deviceStatus:  // device online/offline change
            h.handleDeviceStatus(ctx, status)
        }
    }
}
```

Everything goes through the select loop — no shared-state access outside it. The Hub holds a `map[string]*Client` (deviceID → client) protected by a `sync.RWMutex`.

The Hub has a panic recovery wrapper. If the run loop panics, it logs the error and restarts itself. This prevents a single bad message from killing the entire realtime layer.

### Hub struct fields

| Field | Type | Purpose |
|------|------|---------|
| `clients` | `map[string]*Client` | Active device connections, keyed by device ID |
| `register` / `unreg` | `chan *Client` | Channels for client connect/disconnect |
| `broadcast` | `chan []byte` | Channel for broadcasting to dashboard clients |
| `deviceStatus` | `chan DeviceStatusUpdate` | Channel for online/offline status changes |
| `messageQueue` | `*MessageQueue` | Offline message storage |
| `rateLimiter` | `*RateLimiter` | Per-device message rate limiting |
| `eventProcessor` | `EventProcessor` | Processes device events (telemetry, commands) |
| `dashboardBroadcaster` | `DashboardBroadcaster` | Pushes updates to dashboard WebSocket clients |
| `telemetryFilter` | `*TelemetryFilter` | Filters/deduplicates telemetry frames |
| `compression` | `*Compression` | WebSocket compression config |
| `latencyConfig` | `*LatencyConfig` | Latency tracking config |
| `metrics` | `HubMetrics` | Connection/message counters |

## Clients

A `Client` in `internal/ws/ws_client.go` represents a single WebSocket connection:

```go
type Client struct {
    Conn        *websocket.Conn       // the raw WebSocket connection
    Send        chan command.CommandFrame  // buffered channel for outbound messages
    Hub         *Hub                  // back-reference to the Hub
    DeviceID    string                // the device this client represents
    ClientID    string                // internal client ID
    Done        chan struct{}         // closed when the client disconnects
    isConnected atomic.Bool           // atomic flag for connection state
    connectedAt int64                 // connection timestamp (UnixNano)
    metrics     ClientMetrics         // connection attempt/success/failure counts
}
```

Each client has a buffered `Send` channel (default size 256). The Hub writes `CommandFrame`s to this channel; a per-client goroutine reads from it and writes to the WebSocket connection. If the channel is full (device is slow), the Hub drops the message and logs a warning.

### Connection lifecycle

1. Device connects to `GET /v1/device/:imei/stream`
2. The `StreamHandler` upgrades HTTP to WebSocket
3. Creates a `Client` and calls `hub.Register(client)`
4. The Hub's `handleClientRegistration` adds the client to the `clients` map
5. The Hub replays any queued messages from the `MessageQueue` (messages that arrived while the device was offline)
6. The Hub emits a "device connected" event via the `EventProcessor`
7. The client's read goroutine starts listening for incoming messages
8. The client's write goroutine starts draining the `Send` channel

On disconnect:

1. The WebSocket connection closes (either side initiates)
2. The `Done` channel closes
3. The Hub's `handleClientUnregistration` removes the client from the map
4. The Hub emits a "device disconnected" event
5. A `DeviceStatusUpdate` with `Online: false` is sent

## Message routing

The `MessageRouter` in `internal/api/handlers/websocket/stream_message.go` handles incoming messages from devices. It reads the `type` field from the JSON envelope and routes:

| Message type | Handler | What it does |
|-------------|---------|-------------|
| `telemetry` | `handleTelemetry` | Parses telemetry frame, stores in DB, forwards to dashboard |
| `SUBSCRIBE` | `handleSubscribe` | Dashboard client subscribes to device updates |
| `UNSUBSCRIBE` | `handleUnsubscribe` | Dashboard client unsubscribes |
| `pong` | `handlePong` | WebSocket keepalive response |
| `status` | `handleStatus` | Device reports status (battery, network, etc.) |

Unknown message types are logged with a warning and dropped.

### Telemetry handling

When a device sends a telemetry frame:

1. The `MessageRouter` unmarshals the `telemetry.TelemetryFrame`
2. The `TelemetryFilter` checks if this frame should be processed (deduplication, rate filtering)
3. The frame is stored in the `device_telemetry` table via the telemetry repository
4. If a `DashboardBroadcaster` is set, the frame is forwarded to all subscribed dashboard clients
5. If any thresholds are configured for the device (in `device_settings`), the server checks if a threshold was crossed and generates a device event

## Command delivery

The Hub's `Send` method is how commands reach devices:

```go
func (h *Hub) Send(deviceID string, frame command.CommandFrame) bool {
    c := h.clients[deviceID]
    if c == nil {
        // Device offline — queue the message
        if h.messageQueue != nil {
            return h.messageQueue.EnqueueWithConfirmation(deviceID, frame)
        }
        return false
    }
    select {
    case c.Send <- frame:  // delivered to the client's write goroutine
        return true
    default:               // channel full, device is slow
        return false
    }
}
```

If the device is online, the frame goes to the `Send` channel and the client's write goroutine pushes it over the WebSocket. If the device is offline, the frame is persisted to the `MessageQueue` for later delivery.

There's also `SendWithDeliveryConfirmation` which waits for the device to acknowledge receipt within a timeout. Used for high-priority commands.

## Message queue

The `MessageQueue` in `internal/ws/message_queue.go` stores messages for offline devices. It's a two-tier system:

- **In-memory channels** — for low-latency access when the device reconnects
- **SQLite persistence** — for durability across server restarts

```go
type MessageQueueConfig struct {
    MaxQueueSize    int           // Max messages per device (default 1000)
    MessageTTL      time.Duration // How long messages live (default 7 days)
    MaxMessageAge   time.Duration // Cleanup threshold (default 7 days)
    CleanupInterval time.Duration // How often to clean expired messages (default 1 hour)
}
```

When a device reconnects, the Hub calls `replayQueuedMessages` which drains the queue and delivers all pending messages. The replay result includes:

- `Count` — how many messages were delivered
- `HasMore` — whether the queue still has messages (buffer was full)
- `Remaining` — estimated count of pending messages

The queue tracks metrics: total enqueued, total delivered, total expired, total dropped.

## Rate limiting

The `RateLimiter` in `internal/ws/ws_rate_limiter.go` caps how many messages a device can send per second. This prevents a malfunctioning or compromised device from flooding the server.

```go
type RateLimiterConfig struct {
    MaxMessagesPerSecond int  // Default: 10
    BurstSize            int  // Default: 20
    WindowSize           time.Duration  // Default: 1 second
}
```

Messages exceeding the rate are dropped with a `WARN` log. The rate limiter tracks per-device metrics: allowed count, dropped count, current rate.

## Dashboard broadcasting

Dashboard clients (operators viewing the dashboard) connect via WebSocket and subscribe to device updates. The `DashboardBroadcaster` interface on the Hub handles this:

```go
type DashboardBroadcaster interface {
    Broadcast(deviceID string, msg []byte)
    BroadcastAll(msg []byte)
}
```

When a device sends telemetry, the Hub calls `Broadcast(deviceID, telemetryFrame)`. The broadcaster forwards it to all dashboard clients subscribed to that device. `BroadcastAll` sends to all connected dashboard clients regardless of subscription.

Dashboard clients subscribe via `SUBSCRIBE` messages:

```json
{"type": "SUBSCRIBE", "deviceId": "356938035643809"}
```

Or subscribe to all devices:

```json
{"type": "SUBSCRIBE", "deviceId": "*"}
```

## Latency tracking

The Hub can track command delivery latency per device. When `LatencyConfig.Enabled` is true, the Hub records:

- Time from `hub.Send` call to the frame being written to the WebSocket
- Whether the send succeeded or was dropped/queued

The `LatencyMetrics` struct tracks: min, max, average, p50, p95, p99 latency. Queryable via the metrics endpoint.

## Hub metrics

The Hub tracks aggregate metrics in `HubMetrics`:

| Metric | What it counts |
|--------|---------------|
| `TotalClientsConnected` | Total connections since startup |
| `TotalMessagesSent` | Messages sent to devices |
| `TotalMessagesReceived` | Messages received from devices |
| `TotalConnectAttempts` | WebSocket upgrade attempts |
| `TotalConnectSuccesses` | Successful upgrades |
| `TotalConnectFailures` | Failed upgrades (auth, network) |

## Connection status

The `Online(deviceID)` method checks if a device has an active WebSocket connection:

```go
func (h *Hub) Online(deviceID string) bool {
    h.mu.RLock()
    defer h.mu.RUnlock()
    _, ok := h.clients[deviceID]
    return ok
}
```

This is used by the command execution path to decide between WebSocket delivery and FCM fallback.

`SetDeviceOnline(deviceID, online)` can force-update the online status (used by the connection-status handler when a disconnect is detected out-of-band).

## WebSocket authentication

WebSocket connections require HMAC authentication (when `ENFORCE_HMAC=true`). The `StreamHandler` in `internal/api/handlers/websocket/websocket_stream.go`:

1. Extracts the device ID from the URL path
2. Verifies the HMAC signature in the query parameters (or headers)
3. Upgrades the HTTP connection to WebSocket
4. Creates a `Client` and registers it with the Hub

In development (`ENFORCE_HMAC=false`), the HMAC check is skipped — any connection is accepted. A warning is logged.

## Compression

The Hub supports WebSocket compression (permessage-deflate). The `Compression` struct configures:

- Enabled/disabled
- Compression level (default: 3)
- Threshold (only compress messages above this size, default: 256 bytes)

Compression reduces bandwidth for telemetry-heavy connections but adds CPU overhead. The threshold prevents compressing small messages where the overhead exceeds the savings.
