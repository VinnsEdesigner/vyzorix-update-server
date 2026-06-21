# REST to GraphQL Migration Guide

> **Status**: Ready for Migration  
> **Version**: 1.0  
> **Last Updated**: 2026-06-21

---

## Overview

This guide helps you migrate from REST API calls to GraphQL for the Vyzorix dashboard. GraphQL offers more flexibility, reduced network calls, and better typing.

### Benefits of GraphQL

1. **Single Request**: Fetch all required data in one query instead of multiple REST calls
2. **No Over-fetching**: Request only the fields you need
3. **Strong Typing**: Catch errors at build time
4. **Real-time**: Subscriptions for live updates
5. **Introspection**: Explore the API with auto-generated documentation

---

## Quick Reference

### REST → GraphQL Endpoint Mapping

| REST Endpoint | GraphQL Query |
|---------------|---------------|
| `GET /api/v1/devices` | `query { devices { ... } }` |
| `GET /api/v1/devices/:id` | `query { device(id: $id) { ... } }` |
| `GET /api/v1/devices/:id/telemetry` | `query { telemetryHistory(deviceId: $id) { ... } }` |
| `GET /api/v1/connections` | `query { allConnections { ... } }` |
| `POST /api/v1/device/:id/command` | `mutation { sendCommand(input: {...}) { ... } }` |
| `DELETE /api/v1/devices/:id` | `mutation { deleteDevice(id: $id) }` |

---

## Migration Examples

### Example 1: Fetching Devices

#### REST (Before)
```typescript
// Multiple REST calls
const devices = await fetch('/api/v1/devices');
const connections = await fetch('/api/v1/connections');
const commands = await fetch('/api/v1/commands/pending');

// Manually combine data
const dashboard = {
  devices: devices.data,
  connections: connections.data,
  pendingCommands: commands.data
};
```

#### GraphQL (After)
```typescript
// Single query - use the pre-built hook
import { useDashboardData } from '@/lib/api/graphql';

const { data, isLoading } = useDashboardData(50, { enabled: true });

// data.devices, data.connections, data.totalDevices, etc.
```

---

### Example 2: Device Detail

#### REST (Before)
```typescript
// 4 separate REST calls
const device = await fetch(`/api/v1/devices/${deviceId}`);
const telemetry = await fetch(`/api/v1/devices/${deviceId}/telemetry`);
const commands = await fetch(`/api/v1/device/${deviceId}/commands`);
const stats = await fetch(`/api/v1/device/${deviceId}/telemetry/stats`);
```

#### GraphQL (After)
```typescript
// Single query with useDeviceDetail
import { useDeviceDetail } from '@/lib/api/graphql';

const { data, isLoading } = useDeviceDetail(deviceId, { enabled: true });

// data contains everything: device, telemetry, commands, stats
```

---

### Example 3: Sending Commands

#### REST (Before)
```typescript
const response = await fetch(`/api/v1/device/${deviceId}/command`, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  },
  body: JSON.stringify({
    command: 'RESTART_APP',
    args: null
  })
});

const result = await response.json();
```

#### GraphQL (After)
```typescript
// Using the mutation hook
import { useSendCommand } from '@/lib/api/graphql';

const sendCommandMutation = useSendCommand({
  onSuccess: (data) => {
    console.log('Command sent:', data.sendCommand.dispatchId);
  }
});

// Trigger mutation
sendCommandMutation.mutate({
  deviceId: deviceId,
  command: 'RESTART_APP'
});
```

---

### Example 4: Real-time Updates (WebSocket)

#### REST (Before)
```typescript
// Polling approach
const poll = setInterval(async () => {
  const telemetry = await fetch(`/api/v1/devices/${deviceId}/telemetry/latest`);
  updateUI(telemetry.data);
}, 5000);
```

#### GraphQL (After)
```typescript
// Subscriptions - live updates
import { useTelemetrySubscription } from '@/lib/api/graphql';

const { data } = useTelemetrySubscription(deviceId);

// data.telemetryReceived updates automatically
```

---

## Using the React Query Hooks

### Setup

```typescript
// In your route component
import { Route } from '@tanstack/react-router';
import { 
  useDevices, 
  useDevice, 
  useDashboardData,
  useSendCommand 
} from '@/lib/api/graphql';
```

### Dashboard Page

```typescript
function DashboardPage() {
  // Instead of multiple useQuery calls:
  // const devices = useQuery(['devices'], ...);
  // const connections = useQuery(['connections'], ...);
  
  // Use single hook:
  const { data, isLoading, error } = useDashboardData(50);
  
  if (isLoading) return <DashboardOverviewSkeleton />;
  if (error) return <GraphQLErrorDisplay error={error} />;
  
  return (
    <div>
      <h1>Total Devices: {data?.totalDevices}</h1>
      {/* ... */}
    </div>
  );
}
```

### Device Detail Page

```typescript
function DevicePage({ deviceId }: { deviceId: string }) {
  // Instead of 4 separate queries...
  
  // Use single hook:
  const { data, isLoading } = useDeviceDetail(deviceId);
  
  if (isLoading) return <DeviceDetailSkeleton />;
  
  return (
    <div>
      <h1>{data?.device.model}</h1>
      <TelemetryChart data={data?.telemetry} />
      <CommandsList commands={data?.commands} />
    </div>
  );
}
```

---

## Error Handling

### REST Error Handling
```typescript
if (!response.ok) {
  if (response.status === 401) {
    redirectToLogin();
  }
  throw new Error(`API error: ${response.status}`);
}
```

### GraphQL Error Handling
```typescript
import { GraphQLErrorDisplay } from '@/components/api';

function MyComponent() {
  const { data, isLoading, isError, error } = useDevice(id);
  
  if (isError) {
    return <GraphQLErrorDisplay error={error} onRetry={() => refetch()} />;
  }
  
  // ...
}
```

---

## Loading States

### REST Loading
```typescript
const [loading, setLoading] = useState(true);
// ... fetch data
setLoading(false);
```

### GraphQL Loading
```typescript
import { GraphQLSpinner, GraphQLPageLoading } from '@/components/api';

function MyComponent() {
  const { isLoading } = useDevice(id);
  
  if (isLoading) {
    return <GraphQLPageLoading title="Loading device..." />;
    // or: <GraphQLSpinner label="Loading..." />;
    // or: <DeviceDetailSkeleton />; // for specific skeleton
  }
  
  // ...
}
```

---

## Fetching Only What You Need

### REST - Always Returns Full Objects
```typescript
// This fetches ALL fields, even if you only need name and status
const devices = await fetch('/api/v1/devices');
```

### GraphQL - Request Exactly What You Need
```typescript
// Only fetch the fields you need
const { data } = useQuery(GET_DEVICES, {
  variables: { limit: 10 },
  select: (result) => result.data.devices
});

// GraphQL query:
const GET_DEVICES = gql`
  query GetDevices($limit: Int) {
    devices(limit: $limit) {
      id
      deviceId  # Only these fields
      status    # are fetched
    }
  }
`;
```

---

## Combining REST and GraphQL

You don't need to migrate everything at once:

```typescript
function MyComponent() {
  // Use GraphQL for complex data
  const { data: devices } = useDevices();
  
  // Keep REST for simple, real-time needs
  const { data: stream } = useStream();
  
  // ...
}
```

---

## Migration Checklist

- [ ] Replace `useQuery(['devices'], () => fetch(...))` with `useDevices()`
- [ ] Replace dashboard's 3 REST calls with `useDashboardData()`
- [ ] Replace device detail's 4 REST calls with `useDeviceDetail()`
- [ ] Replace command REST POST with `useSendCommand()` mutation
- [ ] Add `<GraphQLErrorDisplay>` for error handling
- [ ] Add loading skeletons: `<DeviceCardSkeleton />`, `<DashboardOverviewSkeleton />`
- [ ] Update imports from `@/lib/vyzorix-api` to `@/lib/api/graphql`
- [ ] Test authentication still works with GraphQL
- [ ] Verify WebSocket stream still works alongside GraphQL

---

## Troubleshooting

### "Authentication required" error
- Ensure session cookie is being sent
- Check GraphQL client is configured with `credentials: 'include'`

### "Field does not exist" error
- Verify field name in GraphQL schema
- Run introspection: `query { __schema { types { name fields { name } } } }`

### Mutations not working
- Ensure required variables are provided
- Check mutation signature matches schema

---

## Resources

- [GraphQL API Reference](./API.md)
- [Frontend GraphQL Hooks Reference](../web/src/lib/api/graphql/README.md)
- [GraphQL Playground](http://localhost:3000/playground)
