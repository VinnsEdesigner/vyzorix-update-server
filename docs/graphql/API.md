# GraphQL API Reference

> **Status**: Enterprise Production Ready  
> **Version**: 1.0  
> **Last Updated**: 2026-06-21

---

## Overview

The Vyzorix GraphQL API provides a flexible, typed interface for managing devices, commands, and telemetry data. It complements the existing REST API and is designed for frontend applications requiring efficient data fetching.

### Key Features

- **Type-safe queries and mutations** with automatic validation
- **Nested data fetching** - get related data in a single request
- **Real-time subscriptions** via WebSocket for live telemetry
- **Session-based authentication** using existing session cookies
- **Playground interface** for interactive exploration

### Endpoint

| Environment | URL |
|-------------|-----|
| Development | `http://localhost:3000/graphql` |
| Production | `https://api.vyzorix.com/graphql` |

### WebSocket for Subscriptions

| Environment | URL |
|-------------|-----|
| Development | `ws://localhost:3000/graphql/ws` |
| Production | `wss://api.vyzorix.com/graphql/ws` |

---

## Authentication

GraphQL uses the same authentication as the REST API:

### Session Cookie (Recommended for Browser)

When logged in via the web dashboard, your session cookie (`vyz_session`) is automatically included in GraphQL requests.

```bash
curl -X POST http://localhost:3000/graphql \
  -H "Content-Type: application/json" \
  -H "Cookie: vyz_session=<your-session-cookie>" \
  -d '{"query":"{ devices { id } }"}'
```

### Authorization Header

For API clients, you can use a Bearer token:

```bash
curl -X POST http://localhost:3000/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{"query":"{ devices { id } }"}'
```

---

## Introspection

Query the schema directly:

```graphql
{
  __schema {
    queryType { name }
    types {
      name
      fields { name }
    }
  }
}
```

---

## Queries

### `devices`

List all devices for the authenticated operator.

```graphql
query {
  devices(limit: Int = 50, offset: Int = 0) {
    id
    deviceId
    model
    manufacturer
    osVersion
    appVersion
    status
    lastSeen
    createdAt
  }
}
```

**Example Response:**
```json
{
  "data": {
    "devices": [
      {
        "id": "1",
        "deviceId": "device-001",
        "model": "Pixel 7",
        "manufacturer": "Google",
        "osVersion": "14",
        "appVersion": "1.4.2",
        "status": "online",
        "lastSeen": "2024-01-15T10:30:00Z",
        "createdAt": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

### `device`

Get a single device by ID.

```graphql
query {
  device(id: ID!) {
    id
    deviceId
    model
    manufacturer
    telemetry {
      timestamp
      riskScore
      thermalTemp
      bufferLevel
      audioMode
    }
    commands(last: Int = 10) {
      id
      dispatchId
      command
      status
      createdAt
    }
  }
}
```

### `telemetryHistory`

Query historical telemetry data.

```graphql
query {
  telemetryHistory(
    deviceId: ID!
    startTime: Int  # Unix timestamp
    endTime: Int    # Unix timestamp
    limit: Int = 100
  ) {
    timestamp
    riskScore
    thermalTemp
    bufferLevel
    audioMode
  }
}
```

### `telemetryStats`

Get aggregated telemetry statistics.

```graphql
query {
  telemetryStats(deviceId: ID!) {
    avgRiskScore
    maxRiskScore
    avgThermalTemp
    maxThermalTemp
    avgBufferLevel
    totalFrames
  }
}
```

### `pendingCommands`

Get pending commands for a device.

```graphql
query {
  pendingCommands(deviceId: ID!) {
    id
    dispatchId
    command
    status
    createdAt
    updatedAt
  }
}
```

### `connectionStatus`

Get WebSocket connection status for a device.

```graphql
query {
  connectionStatus(deviceId: ID!) {
    deviceId
    status
    connectedAt
    lastActivity
    ipAddress
    userAgent
  }
}
```

### `allConnections`

Get all device connections.

```graphql
query {
  allConnections {
    deviceId
    status
    connectedAt
    lastActivity
  }
}
```

---

## Mutations

### `sendCommand`

Send a command to a device.

```graphql
mutation {
  sendCommand(input: {
    deviceId: ID!
    command: String!
    args: JSON  # Optional arguments
  }) {
    dispatchId
    status
    command
    delivery
  }
}
```

**Commands:**
- `RESTART_APP` - Restart the Vyzorix app
- `RESTART_DEVICE` - Reboot the device
- `UPDATE_CONFIG` - Update device configuration
- `FORCE_SYNC` - Force telemetry sync
- `CLEAR_CACHE` - Clear app cache
- `WAKE_UP_UPDATER` - Wake the update daemon

### `cancelCommand`

Cancel a pending command.

```graphql
mutation {
  cancelCommand(dispatchId: ID!): Boolean
}
```

### `retryCommand`

Retry a failed command.

```graphql
mutation {
  retryCommand(dispatchId: ID!): Command
}
```

### `updateFCMToken`

Update Firebase Cloud Messaging token.

```graphql
mutation {
  updateFCMToken(deviceId: ID!, token: String!): Device
}
```

### `deleteDevice`

Delete a device.

```graphql
mutation {
  deleteDevice(id: ID!): Boolean
}
```

---

## Subscriptions

Connect via WebSocket at `/graphql/ws` and use the subscription protocol:

### `deviceUpdated`

Subscribe to device update events.

```graphql
subscription {
  deviceUpdated(deviceId: ID) {
    id
    deviceId
    status
    lastSeen
  }
}
```

### `telemetryReceived`

Subscribe to real-time telemetry.

```graphql
subscription {
  telemetryReceived(deviceId: ID) {
    timestamp
    riskScore
    thermalTemp
    bufferLevel
  }
}
```

### `commandStatusChanged`

Subscribe to command status changes.

```graphql
subscription {
  commandStatusChanged(dispatchId: ID) {
    id
    dispatchId
    status
    updatedAt
  }
}
```

---

## Error Handling

GraphQL returns errors in the `errors` array:

```json
{
  "data": null,
  "errors": [
    {
      "message": "authentication required",
      "extensions": {
        "code": "UNAUTHORIZED"
      }
    }
  ]
}
```

### Error Codes

| Code | Description |
|------|-------------|
| `UNAUTHORIZED` | Authentication required |
| `FORBIDDEN` | Insufficient permissions |
| `NOT_FOUND` | Resource not found |
| `VALIDATION_ERROR` | Invalid input |
| `RATE_LIMITED` | Too many requests |
| `INTERNAL_ERROR` | Server error |

---

## Rate Limiting

GraphQL is subject to the same rate limiting as REST endpoints.

| Plan | Requests/minute |
|------|-----------------|
| Free | 60 |
| Pro | 600 |
| Enterprise | 6000 |

---

## Performance

| Metric | Target |
|--------|--------|
| Query latency (p50) | <50ms |
| Query latency (p99) | <200ms |
| Max query complexity | 1000 |
| Max depth | 10 |
| Max directives | 100 |

---

## Migration from REST

See [MIGRATION_GUIDE.md](./MIGRATION_GUIDE.md) for detailed REST to GraphQL migration instructions.
