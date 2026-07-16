# GraphQL API

This directory contains the GraphQL API implementation for the Vyzorix Update Server. It coexists with the existing REST API, allowing gradual migration.

## Structure

```
graphql/
 context/          # Context utilities for passing auth info
 errors/           # GraphQL-specific error types
 handler/          # HTTP handler for Gin integration
 middleware/       # Authentication middleware
 resolver/         # GraphQL resolvers (Query/Mutation handlers)
 schema/           # Type definitions and schema builders
 validator/        # Input validation
 server.go         # Server initialization
 README.md
```

## Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/graphql` | POST | GraphQL query/mutation endpoint |
| `/graphql` | GET | GraphQL query endpoint (for simple queries) |
| `/playground` | GET | GraphQL Playground UI |

## Authentication

GraphQL uses the same session cookie authentication as REST endpoints:
- Cookie: `vyz_session` - HttpOnly session cookie from login
- Header: `Authorization: Bearer <token>` - API token (not yet implemented)

## GraphQL Schema

### Types

```graphql
type Device {
  id: ID!
  name: String
  online: Boolean!
  lastSeen: DateTime
  fcmToken: String
  version: String
  createdAt: DateTime
}

type Command {
  dispatchId: ID!
  commandId: ID!
  deviceId: ID!
  command: String!
  args: JSON
  status: CommandStatus!
  createdAt: DateTime
  deliveredAt: DateTime
}

enum CommandStatus {
  PENDING
  DELIVERED
  FAILED
  CANCELLED
}

type TelemetryEntry {
  id: ID!
  deviceId: ID!
  receivedAt: DateTime
  riskScore: Int
  bufferLevel: Int
  thermalTemp: Float
  payload: String
}

type ConnectionStatus {
  deviceId: ID!
  connected: Boolean!
  connectedAt: DateTime
  lastMessageAt: DateTime
  uptimeSeconds: Int
}
```

### Queries

```graphql
# Device queries
device(id: ID!): Device
devices(limit: Int, offset: Int): [Device!]!
deviceCount: Int!

# Command queries
command(dispatchId: ID!): Command
pendingCommands(deviceId: ID!): [Command!]!

# Telemetry queries
telemetryHistory(deviceId: ID!, startTime: Int, endTime: Int, limit: Int): [TelemetryEntry!]!
latestTelemetry(deviceId: ID!): TelemetryEntry
telemetryStats(deviceId: ID!): TelemetryStats

# Connection status
connectionStatus(deviceId: ID!): ConnectionStatus
allConnections: [ConnectionStatus!]!
```

### Mutations

```graphql
# Device mutations
updateFCMToken(deviceId: ID!, token: String!): Device!
deleteDevice(id: ID!): Boolean!

# Command mutations
sendCommand(deviceId: ID!, command: String!, args: JSON): CommandResult!
retryCommand(dispatchId: ID!): Command!
cancelCommand(dispatchId: ID!): Boolean!

# Device control
disconnectDevice(deviceId: ID!): Boolean!
```

## Example Usage

### Query: Get Device List

```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "Cookie: vyz_session=<session_cookie>" \
  -d '{
    "query": "query { devices(limit: 10) { id online lastSeen } }"
  }'
```

### Query: Get Telemetry History

```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "Cookie: vyz_session=<session_cookie>" \
  -d '{
    "query": "query { telemetryHistory(deviceId: \"device-123\", limit: 5) { id riskScore thermalTemp receivedAt } }"
  }'
```

### Mutation: Send Command

```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "Cookie: vyz_session=<session_cookie>" \
  -d '{
    "query": "mutation { sendCommand(deviceId: \"device-123\", command: \"restart\", args: {delay: 5}) { dispatchId status deviceOnline } }"
  }'
```

## Integration with Main Server

To enable GraphQL, call `RegisterGraphQL` on the API server:

```go
// In your main.go or server setup
apiServer := api.NewServer(cfg)

// Register GraphQL (after creating services)
err := apiServer.RegisterGraphQL(
    deviceService,
    commandService,
    telemetryRepo,
    hub,
)
if err != nil {
    log.Fatal(err)
}
```

## Production Considerations

1. **Rate Limiting**: GraphQL should use the same rate limiter as REST
2. **Query Depth Limiting**: Add max depth for complex queries
3. **Persisted Queries**: For production, consider persisted queries
4. **Caching**: Add response caching for expensive queries
5. **Monitoring**: Add metrics for query latency and error rates

## Playground

For development, access the GraphQL Playground at `/playground`. It provides:
- Query editor with syntax highlighting
- Schema explorer
- Query history
- Variable editor
