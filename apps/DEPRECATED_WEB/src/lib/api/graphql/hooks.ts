// React Query hooks for GraphQL
// These provide caching, loading states, and error handling

import { useQuery, useMutation, UseQueryOptions, UseMutationOptions } from "@tanstack/react-query";

import { getGraphQLClient } from "./client";
import * as Mutations from "./mutations";
import * as Queries from "./queries";
import type {
  GetDeviceResponse,
  GetDevicesResponse,
  GetDeviceCountResponse,
  GetCommandResponse,
  GetPendingCommandsResponse,
  GetTelemetryHistoryResponse,
  GetLatestTelemetryResponse,
  GetTelemetryStatsResponse,
  GetConnectionStatusResponse,
  GetAllConnectionsResponse,
  GetHealthResponse,
  DashboardData,
  DeviceDetail,
  UpdateFCMTokenResponse,
  DeleteDeviceResponse,
  SendCommandResponse,
  RetryCommandResponse,
  CancelCommandResponse,
  DisconnectDeviceResponse,
} from "./types";

// Helper to execute GraphQL queries
const gqlQuery = <T>(query: string, variables?: Record<string, unknown>): Promise<T> => {
  const client = getGraphQLClient();
  return client.request<T>(query, variables);
};

// Helper to execute GraphQL mutations
const gqlMutation = <T>(mutation: string, variables?: Record<string, unknown>): Promise<T> => {
  const client = getGraphQLClient();
  return client.request<T>(mutation, variables);
};

// eslint-disable-next-line func-style
export function useDevice(
  id: string,
  options?: UseQueryOptions<GetDeviceResponse>,
): ReturnType<typeof useQuery<GetDeviceResponse>> {
  return useQuery({
    queryKey: ["graphql", "device", id],
    queryFn: () => gqlQuery<GetDeviceResponse>(Queries.GET_DEVICE, { id }),
    enabled: Boolean(id),
    ...options,
  });
}

// eslint-disable-next-line func-style
export function useDevices(
  limit?: number,
  offset?: number,
  options?: UseQueryOptions<GetDevicesResponse>,
): ReturnType<typeof useQuery<GetDevicesResponse>> {
  return useQuery({
    queryKey: ["graphql", "devices", { limit, offset }],
    queryFn: () => gqlQuery<GetDevicesResponse>(Queries.GET_DEVICES, { limit, offset }),
    ...options,
  });
}

// eslint-disable-next-line func-style
export function useDeviceCount(
  options?: UseQueryOptions<GetDeviceCountResponse>,
): ReturnType<typeof useQuery<GetDeviceCountResponse>> {
  return useQuery({
    queryKey: ["graphql", "deviceCount"],
    queryFn: () => gqlQuery<GetDeviceCountResponse>(Queries.GET_DEVICE_COUNT),
    ...options,
  });
}

// ============================================================
// COMMAND HOOKS
// ============================================================

// eslint-disable-next-line func-style
export function useCommand(
  dispatchId: string,
  options?: UseQueryOptions<GetCommandResponse>,
): ReturnType<typeof useQuery<GetCommandResponse>> {
  return useQuery({
    queryKey: ["graphql", "command", dispatchId],
    queryFn: () => gqlQuery<GetCommandResponse>(Queries.GET_COMMAND, { dispatchId }),
    enabled: Boolean(dispatchId),
    ...options,
  });
}

// eslint-disable-next-line func-style
export function usePendingCommands(
  deviceId: string,
  options?: UseQueryOptions<GetPendingCommandsResponse>,
): ReturnType<typeof useQuery<GetPendingCommandsResponse>> {
  return useQuery({
    queryKey: ["graphql", "pendingCommands", deviceId],
    queryFn: () => gqlQuery<GetPendingCommandsResponse>(Queries.GET_PENDING_COMMANDS, { deviceId }),
    enabled: Boolean(deviceId),
    ...options,
  });
}

// ============================================================
// TELEMETRY HOOKS
// ============================================================

// eslint-disable-next-line func-style
export function useTelemetryHistory(
  deviceId: string,
  startTime?: number,
  endTime?: number,
  limit?: number,
  options?: UseQueryOptions<GetTelemetryHistoryResponse>,
): ReturnType<typeof useQuery<GetTelemetryHistoryResponse>> {
  return useQuery({
    queryKey: ["graphql", "telemetryHistory", { deviceId, startTime, endTime, limit }],
    queryFn: () =>
      gqlQuery<GetTelemetryHistoryResponse>(Queries.GET_TELEMETRY_HISTORY, {
        deviceId,
        startTime,
        endTime,
        limit,
      }),
    enabled: Boolean(deviceId),
    ...options,
  });
}

// eslint-disable-next-line func-style
export function useLatestTelemetry(
  deviceId: string,
  options?: UseQueryOptions<GetLatestTelemetryResponse>,
): ReturnType<typeof useQuery<GetLatestTelemetryResponse>> {
  return useQuery({
    queryKey: ["graphql", "latestTelemetry", deviceId],
    queryFn: () => gqlQuery<GetLatestTelemetryResponse>(Queries.GET_LATEST_TELEMETRY, { deviceId }),
    enabled: Boolean(deviceId),
    ...options,
  });
}

// eslint-disable-next-line func-style
export function useTelemetryStats(
  deviceId: string,
  options?: UseQueryOptions<GetTelemetryStatsResponse>,
): ReturnType<typeof useQuery<GetTelemetryStatsResponse>> {
  return useQuery({
    queryKey: ["graphql", "telemetryStats", deviceId],
    queryFn: () => gqlQuery<GetTelemetryStatsResponse>(Queries.GET_TELEMETRY_STATS, { deviceId }),
    enabled: Boolean(deviceId),
    ...options,
  });
}

// ============================================================
// CONNECTION HOOKS
// ============================================================

// eslint-disable-next-line func-style
export function useConnectionStatus(
  deviceId: string,
  options?: UseQueryOptions<GetConnectionStatusResponse>,
): ReturnType<typeof useQuery<GetConnectionStatusResponse>> {
  return useQuery({
    queryKey: ["graphql", "connectionStatus", deviceId],
    queryFn: () =>
      gqlQuery<GetConnectionStatusResponse>(Queries.GET_CONNECTION_STATUS, { deviceId }),
    enabled: Boolean(deviceId),
    ...options,
  });
}

// eslint-disable-next-line func-style
export function useAllConnections(
  options?: UseQueryOptions<GetAllConnectionsResponse>,
): ReturnType<typeof useQuery<GetAllConnectionsResponse>> {
  return useQuery({
    queryKey: ["graphql", "allConnections"],
    queryFn: () => gqlQuery<GetAllConnectionsResponse>(Queries.GET_ALL_CONNECTIONS),
    ...options,
  });
}

// ============================================================
// HEALTH HOOK
// ============================================================

// eslint-disable-next-line func-style
export function useHealth(
  options?: UseQueryOptions<GetHealthResponse>,
): ReturnType<typeof useQuery<GetHealthResponse>> {
  return useQuery({
    queryKey: ["graphql", "health"],
    queryFn: () => gqlQuery<GetHealthResponse>(Queries.GET_HEALTH),
    ...options,
  });
}

// ============================================================
// DASHBOARD HOOKS
// ============================================================

// eslint-disable-next-line func-style
export function useDashboardData(
  deviceLimit?: number,
  options?: UseQueryOptions<DashboardData>,
): ReturnType<typeof useQuery<DashboardData>> {
  return useQuery({
    queryKey: ["graphql", "dashboard", { deviceLimit }],
    queryFn: () => gqlQuery<DashboardData>(Queries.GET_DASHBOARD_DATA, { deviceLimit }),
    ...options,
  });
}

// eslint-disable-next-line func-style
export function useDeviceDetail(
  deviceId: string,
  options?: UseQueryOptions<DeviceDetail>,
): ReturnType<typeof useQuery<DeviceDetail>> {
  return useQuery({
    queryKey: ["graphql", "deviceDetail", deviceId],
    queryFn: () => gqlQuery<DeviceDetail>(Queries.GET_DEVICE_DETAIL, { deviceId }),
    enabled: Boolean(deviceId),
    ...options,
  });
}

// ============================================================
// MUTATION HOOKS
// ============================================================

// eslint-disable-next-line func-style
export function useUpdateFCMToken(
  options?: UseMutationOptions<UpdateFCMTokenResponse, Error, { deviceId: string; token: string }>,
): ReturnType<
  typeof useMutation<UpdateFCMTokenResponse, Error, { deviceId: string; token: string }>
> {
  return useMutation({
    mutationKey: ["graphql", "updateFCMToken"],
    mutationFn: ({ deviceId, token }) =>
      gqlMutation<UpdateFCMTokenResponse>(Mutations.UPDATE_FCM_TOKEN, { deviceId, token }),
    ...options,
  });
}

// eslint-disable-next-line func-style
export function useDeleteDevice(
  options?: UseMutationOptions<DeleteDeviceResponse, Error, { id: string }>,
): ReturnType<typeof useMutation<DeleteDeviceResponse, Error, { id: string }>> {
  return useMutation({
    mutationKey: ["graphql", "deleteDevice"],
    mutationFn: ({ id }) => gqlMutation<DeleteDeviceResponse>(Mutations.DELETE_DEVICE, { id }),
    ...options,
  });
}

// eslint-disable-next-line func-style
export function useSendCommand(
  options?: UseMutationOptions<
    SendCommandResponse,
    Error,
    { deviceId: string; command: string; args?: Record<string, unknown> }
  >,
): ReturnType<
  typeof useMutation<
    SendCommandResponse,
    Error,
    { deviceId: string; command: string; args?: Record<string, unknown> }
  >
> {
  return useMutation({
    mutationKey: ["graphql", "sendCommand"],
    mutationFn: ({ deviceId, command, args }) =>
      gqlMutation<SendCommandResponse>(Mutations.SEND_COMMAND, { deviceId, command, args }),
    ...options,
  });
}

// eslint-disable-next-line func-style
export function useRetryCommand(
  options?: UseMutationOptions<RetryCommandResponse, Error, { dispatchId: string }>,
): ReturnType<typeof useMutation<RetryCommandResponse, Error, { dispatchId: string }>> {
  return useMutation({
    mutationKey: ["graphql", "retryCommand"],
    mutationFn: ({ dispatchId }) =>
      gqlMutation<RetryCommandResponse>(Mutations.RETRY_COMMAND, { dispatchId }),
    ...options,
  });
}

// eslint-disable-next-line func-style
export function useCancelCommand(
  options?: UseMutationOptions<CancelCommandResponse, Error, { dispatchId: string }>,
): ReturnType<typeof useMutation<CancelCommandResponse, Error, { dispatchId: string }>> {
  return useMutation({
    mutationKey: ["graphql", "cancelCommand"],
    mutationFn: ({ dispatchId }) =>
      gqlMutation<CancelCommandResponse>(Mutations.CANCEL_COMMAND, { dispatchId }),
    ...options,
  });
}

// eslint-disable-next-line func-style
export function useDisconnectDevice(
  options?: UseMutationOptions<DisconnectDeviceResponse, Error, { deviceId: string }>,
): ReturnType<typeof useMutation<DisconnectDeviceResponse, Error, { deviceId: string }>> {
  return useMutation({
    mutationKey: ["graphql", "disconnectDevice"],
    mutationFn: ({ deviceId }) =>
      gqlMutation<DisconnectDeviceResponse>(Mutations.DISCONNECT_DEVICE, { deviceId }),
    ...options,
  });
}
