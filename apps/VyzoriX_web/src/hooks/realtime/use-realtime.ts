import { useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import {
  DEVICE_UPDATED_SUBSCRIPTION,
  TELEMETRY_RECEIVED_SUBSCRIPTION,
  COMMAND_STATUS_SUBSCRIPTION,
  ORGANIZATION_EVENT_SUBSCRIPTION,
  MEMBER_EVENT_SUBSCRIPTION,
} from '@vyzorix/api-client';
import { useWebSocketStore } from '@/stores';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { queryKeys } from '@/lib/query-keys';

interface SubscriptionDocument {
  loc?: { source: { body: string } };
}

function toQueryString(doc: unknown): string {
  return (doc as SubscriptionDocument)?.loc?.source.body ?? '';
}

interface UseWebSocketConnectionResult {
  isConnected: boolean;
  isReconnecting: boolean;
  connect: () => void;
  disconnect: () => void;
  lastConnectedAt: Date | null;
  connectionError: { code: number; reason: string } | null;
}

/**
 * Manage the realtime WebSocket connection lifecycle (spec §8.1).
 *
 * Auto-connects when an organization is selected and disconnects on cleanup.
 * Returns the live connection status from the underlying graphql-ws store so
 * the UI can render connection indicators without subscribing to the store
 * directly.
 */
export function useWebSocketConnection(): UseWebSocketConnectionResult {
  const organizationId = useCurrentOrganizationId();
  const connect = useWebSocketStore((s) => s.connect);
  const disconnect = useWebSocketStore((s) => s.disconnect);
  const isConnected = useWebSocketStore((s) => s.isConnected);
  const isReconnecting = useWebSocketStore((s) => s.isReconnecting);
  const lastConnectedAt = useWebSocketStore((s) => s.lastConnectedAt);
  const lastError = useWebSocketStore((s) => s.lastError);

  useEffect(() => {
    if (!organizationId) return;
    connect();
    return () => {
      disconnect();
    };
  }, [organizationId, connect, disconnect]);

  return {
    isConnected,
    isReconnecting,
    connect,
    disconnect,
    lastConnectedAt,
    connectionError: lastError,
  };
}

export interface DeviceUpdatedPayload {
  id: string;
  imei: string;
  deviceName?: string;
  status: string;
  lastSeen?: string;
}

export function useDeviceUpdates(deviceId?: string) {
  const queryClient = useQueryClient();
  const subscribe = useWebSocketStore((s) => s.subscribe);
  const organizationId = useCurrentOrganizationId();

  useEffect(() => {
    if (!organizationId) return;
    const unsubscribe = subscribe<DeviceUpdatedPayload>(
      {
        query: toQueryString(DEVICE_UPDATED_SUBSCRIPTION),
        variables: { deviceId: deviceId ?? null },
      },
      {
        next: (data) => {
          queryClient.invalidateQueries({ queryKey: ['devices'] });
          if (data?.imei) {
            queryClient.invalidateQueries({ queryKey: queryKeys.device(data.imei) });
          }
        },
      },
    );
    return unsubscribe;
  }, [deviceId, organizationId, subscribe, queryClient]);
}

export interface TelemetryReceivedPayload {
  id: string;
  deviceId: string;
  receivedAt: string;
  riskScore: number;
  bufferLevel: number;
  thermalTemp: number;
  payload?: string;
}

export function useTelemetryStream(deviceId?: string) {
  const queryClient = useQueryClient();
  const subscribe = useWebSocketStore((s) => s.subscribe);
  const organizationId = useCurrentOrganizationId();

  useEffect(() => {
    if (!organizationId) return;
    const unsubscribe = subscribe<TelemetryReceivedPayload>(
      {
        query: toQueryString(TELEMETRY_RECEIVED_SUBSCRIPTION),
        variables: { deviceId: deviceId ?? null },
      },
      {
        next: (data) => {
          if (data?.deviceId) {
            queryClient.invalidateQueries({
              queryKey: queryKeys.latestTelemetry(organizationId ?? '', data.deviceId),
            });
            queryClient.invalidateQueries({
              queryKey: queryKeys.telemetryStats(organizationId ?? '', data.deviceId),
            });
            queryClient.invalidateQueries({
              queryKey: queryKeys.deviceMetrics(organizationId ?? '', data.deviceId),
            });
          }
        },
      },
    );
    return unsubscribe;
  }, [deviceId, organizationId, subscribe, queryClient]);
}

export interface CommandStatusPayload {
  dispatchId: string;
  commandId: string;
  deviceId: string;
  command: string;
  status: string;
  createdAt: string;
}

export function useCommandStatusStream(dispatchId?: string) {
  const queryClient = useQueryClient();
  const subscribe = useWebSocketStore((s) => s.subscribe);
  const organizationId = useCurrentOrganizationId();

  useEffect(() => {
    if (!organizationId) return;
    const unsubscribe = subscribe<CommandStatusPayload>(
      {
        query: toQueryString(COMMAND_STATUS_SUBSCRIPTION),
        variables: { dispatchId: dispatchId ?? null },
      },
      {
        next: (data) => {
          if (data?.dispatchId) {
            queryClient.invalidateQueries({ queryKey: queryKeys.command(data.dispatchId) });
          }
          if (data?.deviceId) {
            queryClient.invalidateQueries({ queryKey: queryKeys.pendingCommands(data.deviceId) });
            queryClient.invalidateQueries({ queryKey: queryKeys.commands(data.deviceId) });
          }
        },
      },
    );
    return unsubscribe;
  }, [dispatchId, organizationId, subscribe, queryClient]);
}

export interface OrganizationEventPayload {
  type: string;
  timestamp: string;
  data: unknown;
}

export function useOrganizationEvents(orgId: string | undefined) {
  const queryClient = useQueryClient();
  const subscribe = useWebSocketStore((s) => s.subscribe);

  useEffect(() => {
    if (!orgId) return;
    const unsubscribe = subscribe<OrganizationEventPayload>(
      {
        query: toQueryString(ORGANIZATION_EVENT_SUBSCRIPTION),
        variables: { orgId },
      },
      {
        next: () => {
          queryClient.invalidateQueries({ queryKey: ['events'] });
          queryClient.invalidateQueries({ queryKey: queryKeys.dashboardStats(orgId ?? '') });
          queryClient.invalidateQueries({ queryKey: ['devices'] });
        },
      },
    );
    return unsubscribe;
  }, [orgId, subscribe, queryClient]);
}

export interface MemberEventPayload {
  type: string;
  timestamp: string;
  memberId: string;
  data: unknown;
}

export function useMemberEvents(orgId: string | undefined) {
  const queryClient = useQueryClient();
  const subscribe = useWebSocketStore((s) => s.subscribe);

  useEffect(() => {
    if (!orgId) return;
    const unsubscribe = subscribe<MemberEventPayload>(
      {
        query: toQueryString(MEMBER_EVENT_SUBSCRIPTION),
        variables: { orgId },
      },
      {
        next: () => {
          queryClient.invalidateQueries({ queryKey: queryKeys.organizationMembers(orgId) });
        },
      },
    );
    return unsubscribe;
  }, [orgId, subscribe, queryClient]);
}
