// React Query hooks for GraphQL
// These provide caching, loading states, and error handling

import { useQuery, useMutation, UseQueryOptions, UseMutationOptions } from '@tanstack/react-query';
import { request } from 'graphql-request';
import { getGraphQLClient } from './client';
import * as Queries from './queries';
import * as Mutations from './mutations';
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
} from './types';

// Helper to execute GraphQL queries
async function gqlQuery<T>(query: string, variables?: Record<string, unknown>): Promise<T> {
  const client = getGraphQLClient();
  return client.request<T>(query, variables);
}

// Helper to execute GraphQL mutations
async function gqlMutation<T>(mutation: string, variables?: Record<string, unknown>): Promise<T> {
  const client = getGraphQLClient();
  return client.request<T>(mutation, variables);
}

// ============================================================
// DEVICE HOOKS
// ============================================================

export function useDevice(id: string, options?: UseQueryOptions<GetDeviceResponse>) {
  return useQuery({
    queryKey: ['graphql', 'device', id],
    queryFn: () => gqlQuery<GetDeviceResponse>(Queries.GET_DEVICE, { id }),
    enabled: !!id,
    ...options,
  });
}

export function useDevices(limit?: number, offset?: number, options?: UseQueryOptions<GetDevicesResponse>) {
  return useQuery({
    queryKey: ['graphql', 'devices', { limit, offset }],
    queryFn: () => gqlQuery<GetDevicesResponse>(Queries.GET_DEVICES, { limit, offset }),
    ...options,
  });
}

export function useDeviceCount(options?: UseQueryOptions<GetDeviceCountResponse>) {
  return useQuery({
    queryKey: ['graphql', 'deviceCount'],
    queryFn: () => gqlQuery<GetDeviceCountResponse>(Queries.GET_DEVICE_COUNT),
    ...options,
  });
}

// ============================================================
// COMMAND HOOKS
// ============================================================

export function useCommand(dispatchId: string, options?: UseQueryOptions<GetCommandResponse>) {
  return useQuery({
    queryKey: ['graphql', 'command', dispatchId],
    queryFn: () => gqlQuery<GetCommandResponse>(Queries.GET_COMMAND, { dispatchId }),
    enabled: !!dispatchId,
    ...options,
  });
}

export function usePendingCommands(deviceId: string, options?: UseQueryOptions<GetPendingCommandsResponse>) {
  return useQuery({
    queryKey: ['graphql', 'pendingCommands', deviceId],
    queryFn: () => gqlQuery<GetPendingCommandsResponse>(Queries.GET_PENDING_COMMANDS, { deviceId }),
    enabled: !!deviceId,
    ...options,
  });
}

// ============================================================
// TELEMETRY HOOKS
// ============================================================

export function useTelemetryHistory(
  deviceId: string,
  startTime?: number,
  endTime?: number,
  limit?: number,
  options?: UseQueryOptions<GetTelemetryHistoryResponse>
) {
  return useQuery({
    queryKey: ['graphql', 'telemetryHistory', { deviceId, startTime, endTime, limit }],
    queryFn: () =>
      gqlQuery<GetTelemetryHistoryResponse>(Queries.GET_TELEMETRY_HISTORY, {
        deviceId,
        startTime,
        endTime,
        limit,
      }),
    enabled: !!deviceId,
    ...options,
  });
}

export function useLatestTelemetry(deviceId: string, options?: UseQueryOptions<GetLatestTelemetryResponse>) {
  return useQuery({
    queryKey: ['graphql', 'latestTelemetry', deviceId],
    queryFn: () => gqlQuery<GetLatestTelemetryResponse>(Queries.GET_LATEST_TELEMETRY, { deviceId }),
    enabled: !!deviceId,
    ...options,
  });
}

export function useTelemetryStats(deviceId: string, options?: UseQueryOptions<GetTelemetryStatsResponse>) {
  return useQuery({
    queryKey: ['graphql', 'telemetryStats', deviceId],
    queryFn: () => gqlQuery<GetTelemetryStatsResponse>(Queries.GET_TELEMETRY_STATS, { deviceId }),
    enabled: !!deviceId,
    ...options,
  });
}

// ============================================================
// CONNECTION HOOKS
// ============================================================

export function useConnectionStatus(
  deviceId: string,
  options?: UseQueryOptions<GetConnectionStatusResponse>
) {
  return useQuery({
    queryKey: ['graphql', 'connectionStatus', deviceId],
    queryFn: () => gqlQuery<GetConnectionStatusResponse>(Queries.GET_CONNECTION_STATUS, { deviceId }),
    enabled: !!deviceId,
    ...options,
  });
}

export function useAllConnections(options?: UseQueryOptions<GetAllConnectionsResponse>) {
  return useQuery({
    queryKey: ['graphql', 'allConnections'],
    queryFn: () => gqlQuery<GetAllConnectionsResponse>(Queries.GET_ALL_CONNECTIONS),
    ...options,
  });
}

// ============================================================
// HEALTH HOOK
// ============================================================

export function useHealth(options?: UseQueryOptions<GetHealthResponse>) {
  return useQuery({
    queryKey: ['graphql', 'health'],
    queryFn: () => gqlQuery<GetHealthResponse>(Queries.GET_HEALTH),
    ...options,
  });
}

// ============================================================
// DASHBOARD HOOKS
// ============================================================

export function useDashboardData(deviceLimit?: number, options?: UseQueryOptions<DashboardData>) {
  return useQuery({
    queryKey: ['graphql', 'dashboard', { deviceLimit }],
    queryFn: () => gqlQuery<DashboardData>(Queries.GET_DASHBOARD_DATA, { deviceLimit }),
    ...options,
  });
}

export function useDeviceDetail(deviceId: string, options?: UseQueryOptions<DeviceDetail>) {
  return useQuery({
    queryKey: ['graphql', 'deviceDetail', deviceId],
    queryFn: () => gqlQuery<DeviceDetail>(Queries.GET_DEVICE_DETAIL, { deviceId }),
    enabled: !!deviceId,
    ...options,
  });
}

// ============================================================
// MUTATION HOOKS
// ============================================================

export function useUpdateFCMToken(options?: UseMutationOptions<UpdateFCMTokenResponse, Error, { deviceId: string; token: string }>) {
  return useMutation({
    mutationKey: ['graphql', 'updateFCMToken'],
    mutationFn: ({ deviceId, token }) =>
      gqlMutation<UpdateFCMTokenResponse>(Mutations.UPDATE_FCM_TOKEN, { deviceId, token }),
    ...options,
  });
}

export function useDeleteDevice(options?: UseMutationOptions<DeleteDeviceResponse, Error, { id: string }>) {
  return useMutation({
    mutationKey: ['graphql', 'deleteDevice'],
    mutationFn: ({ id }) => gqlMutation<DeleteDeviceResponse>(Mutations.DELETE_DEVICE, { id }),
    ...options,
  });
}

export function useSendCommand(
  options?: UseMutationOptions<SendCommandResponse, Error, { deviceId: string; command: string; args?: Record<string, unknown> }>
) {
  return useMutation({
    mutationKey: ['graphql', 'sendCommand'],
    mutationFn: ({ deviceId, command, args }) =>
      gqlMutation<SendCommandResponse>(Mutations.SEND_COMMAND, { deviceId, command, args }),
    ...options,
  });
}

export function useRetryCommand(
  options?: UseMutationOptions<RetryCommandResponse, Error, { dispatchId: string }>
) {
  return useMutation({
    mutationKey: ['graphql', 'retryCommand'],
    mutationFn: ({ dispatchId }) =>
      gqlMutation<RetryCommandResponse>(Mutations.RETRY_COMMAND, { dispatchId }),
    ...options,
  });
}

export function useCancelCommand(
  options?: UseMutationOptions<CancelCommandResponse, Error, { dispatchId: string }>
) {
  return useMutation({
    mutationKey: ['graphql', 'cancelCommand'],
    mutationFn: ({ dispatchId }) =>
      gqlMutation<CancelCommandResponse>(Mutations.CANCEL_COMMAND, { dispatchId }),
    ...options,
  });
}

export function useDisconnectDevice(
  options?: UseMutationOptions<DisconnectDeviceResponse, Error, { deviceId: string }>
) {
  return useMutation({
    mutationKey: ['graphql', 'disconnectDevice'],
    mutationFn: ({ deviceId }) =>
      gqlMutation<DisconnectDeviceResponse>(Mutations.DISCONNECT_DEVICE, { deviceId }),
    ...options,
  });
}
