# GraphQL API Integration Guide

This directory contains the GraphQL client, queries, mutations, and React Query hooks for the Vyzorix API.

## Quick Start

### 1. Using GraphQL Hooks (Recommended)

```tsx
import { useDevices, useDashboardData, useSendCommand } from '@/lib/api/graphql';

// In your component:
function MyComponent() {
  // Get all devices
  const { data: devices, isLoading, error } = useDevices(50);
  
  // Get dashboard data (devices + connections in one query)
  const { data: dashboard } = useDashboardData(50);
  
  // Send a command
  const sendCommand = useSendCommand();
  
  const handleCommand = async () => {
    await sendCommand.mutateAsync({
      deviceId: 'device-123',
      command: 'REQUEST_STATUS',
    });
  };
  
  if (isLoading) return <div>Loading...</div>;
  return <div>{devices?.map(d => <DeviceCard key={d.id} device={d} />)}</div>;
}
```

### 2. Using the Client Directly

```tsx
import { getGraphQLClient } from '@/lib/api/graphql/client';
import { GET_DEVICES } from '@/lib/api/graphql/queries';

async function fetchDevices() {
  const client = getGraphQLClient();
  return await client.request(GET_DEVICES, { limit: 10 });
}
```

## Migration from REST API

### Before (REST)

```tsx
import { useQuery } from '@tanstack/react-query';
import { getDashboardDevices, getDeviceStatus, dispatchCommand } from '@/lib/vyzorix-api';

// Multiple queries for dashboard
const devices = useQuery({
  queryKey: ['devices', serverUrl],
  queryFn: () => getDashboardDevices(serverUrl),
});

const deviceStatus = useQuery({
  queryKey: ['status', deviceId],
  queryFn: () => getDeviceStatus(serverUrl, deviceId),
});

// Sending command via REST
const handleCommand = async () => {
  await dispatchCommand(serverUrl, deviceId, 'FORCE_SPEAKER');
};
```

### After (GraphQL)

```tsx
import { useDashboardData, useSendCommand } from '@/lib/api/graphql';

// Single query fetches devices + connections
const { data: dashboard } = useDashboardData(50);

// Single mutation handles all commands
const sendCommand = useSendCommand();

const handleCommand = async () => {
  await sendCommand.mutateAsync({
    deviceId: deviceId,
    command: 'FORCE_SPEAKER',
  });
};
```

## Available Hooks

### Queries

| Hook | Description |
|------|-------------|
| `useDevice(id)` | Get single device |
| `useDevices(limit, offset)` | List devices |
| `useDeviceCount()` | Get device count |
| `useCommand(dispatchId)` | Get command status |
| `usePendingCommands(deviceId)` | Get pending commands |
| `useTelemetryHistory(deviceId, start, end, limit)` | Get telemetry history |
| `useLatestTelemetry(deviceId)` | Get latest telemetry |
| `useTelemetryStats(deviceId)` | Get telemetry statistics |
| `useConnectionStatus(deviceId)` | Get device connection status |
| `useAllConnections()` | Get all connections |
| `useDashboardData(limit)` | Get dashboard (devices + connections) |
| `useDeviceDetail(deviceId)` | Get device with stats + connections |

### Mutations

| Hook | Description |
|------|-------------|
| `useUpdateFCMToken()` | Update device FCM token |
| `useDeleteDevice()` | Delete a device |
| `useSendCommand()` | Send command to device |
| `useRetryCommand()` | Retry failed command |
| `useCancelCommand()` | Cancel pending command |
| `useDisconnectDevice()` | Force disconnect device |

## Environment Variables

The GraphQL client automatically uses the current origin for API calls. Make sure `ENABLE_GRAPHQL=true` is set on the server.

## Authentication

GraphQL uses the same session cookie authentication as REST:
- Login via REST (`POST /v1/auth/login`) sets the `vyz_session` cookie
- GraphQL automatically sends this cookie with requests
- No additional authentication headers needed

## Benefits

1. **Single request for related data** - Dashboard data in one query
2. **Type-safe** - Full TypeScript support with generated types
3. **Caching** - React Query handles caching automatically
4. **Dev tools** - GraphQL Playground at `/playground`
5. **Flexible queries** - Request only what you need

## Notes

- The GraphQL API is additive - REST continues to work
- For real-time data, WebSocket streaming is still recommended
- Complex queries may need server-side rate limiting adjustments
