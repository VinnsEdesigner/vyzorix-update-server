# GraphQL API

This directory contains the GraphQL API implementation for the Vyzorix Update Server. It provides a comprehensive API for managing devices, commands, telemetry, organizations, and more in a multi-tenant environment.

## Structure

```
graphql/
 adapters/         # Database adapters for GraphQL
 context/          # Context utilities for passing auth info
 errors/           # GraphQL-specific error types
 handler/          # HTTP handler for Gin integration
 middleware/       # Authentication middleware
 resolver/         # GraphQL resolvers (Query/Mutation handlers)
 schema/           # Type definitions and schema builders
 subscription/     # WebSocket subscriptions for real-time updates
 validator/        # Input validation
 gql_server.go     # Server initialization
 README.md
```

## Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/graphql` | POST | GraphQL query/mutation endpoint |
| `/graphql` | GET | GraphQL query endpoint (for simple queries) |
| `/playground` | GET | GraphQL Playground UI (development only) |

## Authentication

GraphQL uses the same session cookie authentication as REST endpoints:
- Cookie: `vyz_session` - HttpOnly session cookie from login
- Header: `Authorization: Bearer <token>` - API token (for API keys)

## Multi-Tenant Organization Model

All queries and mutations require an `organizationId` parameter to enforce tenant isolation. The authenticated operator must be a member of the specified organization.

### Organization Roles

| Role | Permissions |
|------|-------------|
| `super_admin` | Full access, can manage members and settings |
| `admin` | Can manage devices and send commands |
| `operator` | Can view devices and telemetry |
| `viewer` | Read-only access |

## GraphQL Schema

### Query Categories

- **Settings**: Operator settings, device settings, organization settings
- **Inbox**: Device registration requests
- **Devices**: Device management and status
- **Commands**: Command dispatch and status
- **Telemetry**: Historical telemetry data and statistics
- **Connections**: Device connection status
- **Dashboard**: Device logs, events, and diagnostics
- **Updates**: Software update versions and pushes
- **Organizations**: Organization and membership management

### Example Queries

```graphql
# Get devices in an organization
query {
  devices(organizationId: "org-123", limit: 50, offset: 0) {
    id
    name
    online
    lastSeen
    version
  }
}

# Get device telemetry history
query {
  telemetryHistory(
    organizationId: "org-123"
    deviceId: "device-456"
    limit: 100
  ) {
    id
    receivedAt
    riskScore
    bufferLevel
    thermalTemp
  }
}

# Get organization settings
query {
  organizationSettings(organizationId: "org-123") {
    id
    defaultRiskThreshold
    defaultBufferAlertLevel
    maxDevices
  }
}
```

### Example Mutations

```graphql
# Send a command to a device
mutation {
  sendCommand(
    organizationId: "org-123"
    deviceId: "device-456"
    command: "restart"
    args: { delay: 5 }
  ) {
    dispatchId
    commandId
    status
  }
}

# Create an organization
mutation {
  createOrganization(name: "My Organization", maxMembers: 100) {
    id
    name
    createdAt
    membership {
      role
    }
  }
}

# Invite a member
mutation {
  inviteMember(
    organizationId: "org-123"
    email: "user@example.com"
    role: operator
  ) {
    id
    email
    role
    status
  }
}
```

## Example Usage

### Query: Get Device List

```bash
curl -X POST http://localhost:3000/graphql \
  -H "Content-Type: application/json" \
  -H "Cookie: vyz_session=<session_cookie>" \
  -d '{
    "query": "query { devices(organizationId: \"org-123\", limit: 10) { id name online lastSeen } }"
  }'
```

### Query: Get Telemetry History

```bash
curl -X POST http://localhost:3000/graphql \
  -H "Content-Type: application/json" \
  -H "Cookie: vyz_session=<session_cookie>" \
  -d '{
    "query": "query { telemetryHistory(organizationId: \"org-123\", deviceId: \"device-456\", limit: 5) { id riskScore thermalTemp receivedAt } }"
  }'
```

### Mutation: Send Command

```bash
curl -X POST http://localhost:3000/graphql \
  -H "Content-Type: application/json" \
  -H "Cookie: vyz_session=<session_cookie>" \
  -d '{
    "query": "mutation { sendCommand(organizationId: \"org-123\", deviceId: \"device-456\", command: \"restart\", args: {delay: 5}) { dispatchId status } }"
  }'
```

## Integration with Main Server

The GraphQL server is integrated via wire dependency injection:

```go
// Wire handles dependency injection automatically.
// See internal/api/wire/providers.go for service wiring.
```

## Production Considerations

1. **Rate Limiting**: GraphQL uses the same rate limiter as REST endpoints
2. **Query Depth Limiting**: Consider adding max depth for complex queries
3. **Persisted Queries**: For production, consider persisted queries
4. **Caching**: Add response caching for expensive queries
5. **Monitoring**: Add metrics for query latency and error rates
6. **Security**: All queries are scoped to the operator's organization membership

## Development

### Access Playground

For development, access the GraphQL Playground at `/playground`. It provides:
- Query editor with syntax highlighting
- Schema explorer (Schema tab)
- Query history
- Variable editor

### Run Tests

```bash
cd apps/api
make test
```

### Generate Wire Dependencies

```bash
cd apps/api
make wire
```
