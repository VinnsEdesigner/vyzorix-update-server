import { useEffect, useRef, useState, useCallback, useMemo } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import {
  getCommands,
  validateTelemetry,
  telemetryFromRaw,
  type WSTelemetry,
  type WSEvent,
  type WSEventType,
  type WSCommandType,
} from '@vyzorix/api-client';
import {
  OnTelemetryReceivedDocument as TELEMETRY_RECEIVED_SUBSCRIPTION,
  OnOrganizationEventDocument as ORGANIZATION_EVENT_SUBSCRIPTION,
  OnCommandStatusChangedDocument as COMMAND_STATUS_SUBSCRIPTION,
} from '@vyzorix/api-client/generated-graphql';
import { useWebSocketStore } from '@/stores';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { queryKeys } from '@/lib/query-keys';

const MAX_TELEMETRY_HISTORY = 100;
const MAX_EVENTS = 200;

interface SubscriptionDocument {
  loc?: { source: { body: string } };
}

function toQueryString(doc: unknown): string {
  return (doc as SubscriptionDocument)?.loc?.source.body ?? '';
}

/**
 * Subscribe to a device's live telemetry stream (spec §8.2).
 *
 * Maintains a rolling buffer of the last `MAX_TELEMETRY_HISTORY` frames and
 * exposes the latest frame plus the full history. Also invalidates the REST
 * telemetry query caches so REST-backed components stay in sync. Frames that
 * fail domain validation are dropped (never surface malformed data to the UI).
 */
interface UseDeviceTelemetryOptions {
  imei?: string;
}

interface UseDeviceTelemetryResult {
  telemetry: WSTelemetry | null;
  telemetryHistory: WSTelemetry[];
  isLoading: boolean;
  error: Error | null;
}

interface TelemetryReceivedPayload {
  id: string;
  deviceId: string;
  receivedAt: string;
  riskScore: number;
  bufferLevel: number;
  thermalTemp: number;
  payload?: string;
}

export function useDeviceTelemetry(
  options: UseDeviceTelemetryOptions = {},
): UseDeviceTelemetryResult {
  const { imei } = options;
  const queryClient = useQueryClient();
  const subscribe = useWebSocketStore((s) => s.subscribe);
  const organizationId = useCurrentOrganizationId();

  const historyRef = useRef<WSTelemetry[]>([]);
  const [telemetry, setTelemetry] = useState<WSTelemetry | null>(null);
  const [history, setHistory] = useState<WSTelemetry[]>([]);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!organizationId || !imei) return;

    const unsubscribe = subscribe<TelemetryReceivedPayload>(
      {
        query: toQueryString(TELEMETRY_RECEIVED_SUBSCRIPTION),
        variables: { deviceId: imei },
      },
      {
        next: (data) => {
          if (!data) return;
          // Invalidate REST telemetry caches so REST components stay in sync.
          queryClient.invalidateQueries({
            queryKey: queryKeys.latestTelemetry(organizationId ?? '', data.deviceId),
          });
          queryClient.invalidateQueries({
            queryKey: queryKeys.telemetryStats(organizationId ?? '', data.deviceId),
          });

          const frame = telemetryFromRaw({
            deviceId: data.deviceId,
            timestamp: data.receivedAt ? new Date(data.receivedAt).getTime() : Date.now(),
            riskScore: data.riskScore,
            bufferLevel: data.bufferLevel,
            thermalTemp: data.thermalTemp,
          });
          if (!validateTelemetry(frame)) return;

          historyRef.current = [...historyRef.current, frame].slice(-MAX_TELEMETRY_HISTORY);
          setTelemetry(frame);
          setHistory(historyRef.current);
          setError(null);
        },
        error: (err) => {
          setError(err instanceof Error ? err : new Error(String(err)));
        },
      },
    );
    return () => {
      unsubscribe();
      historyRef.current = [];
      setTelemetry(null);
      setHistory([]);
      setError(null);
    };
  }, [imei, organizationId, subscribe, queryClient]);

  return {
    telemetry,
    telemetryHistory: history,
    isLoading: telemetry === null && error === null,
    error,
  };
}

/**
 * Collect dashboard events (device connect/disconnect, threshold breach,
 * command delivered/failed) with unread tracking (spec §8.3).
 *
 * Events are buffered up to `MAX_EVENTS`. `unreadCount` increments per new
 * event and is reset by `markAsRead` / `clearEvents`. Optional `eventTypes`
 * and `deviceImei` filters narrow the collected set.
 */
interface UseDashboardEventsOptions {
  eventTypes?: WSEventType[];
  deviceImei?: string;
}

interface UseDashboardEventsResult {
  events: WSEvent[];
  unreadCount: number;
  markAsRead: (eventId: string) => void;
  clearEvents: () => void;
}

interface OrganizationEventPayload {
  type: string;
  timestamp: string;
  deviceId?: string;
  data?: unknown;
}

let eventIdCounter = 0;

export function useDashboardEvents(
  options: UseDashboardEventsOptions = {},
): UseDashboardEventsResult {
  const { eventTypes, deviceImei } = options;
  const queryClient = useQueryClient();
  const subscribe = useWebSocketStore((s) => s.subscribe);
  const organizationId = useCurrentOrganizationId();

  // Stabilize eventTypes so the useEffect below doesn't re-subscribe on every
  // render when the caller passes an inline array literal (e.g.
  // useDashboardEvents({ eventTypes: ['DEVICE_CONNECTED'] })). Without this,
  // each render produces a new array reference, the effect deps change, and
  // the hook churns subscriptions — which compounds with the state updates
  // from the next-callback into exponential growth.
  const stableEventTypes = useMemo(
    () => (eventTypes ? [...eventTypes] : null),
    // Re-memoize only when the joined type values change, not the array identity.
    [eventTypes?.join('|')],
  );

  const eventsRef = useRef<WSEvent[]>([]);
  const [events, setEvents] = useState<WSEvent[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);

  useEffect(() => {
    if (!organizationId) return;

    const filterTypes = stableEventTypes ? new Set<WSEventType>(stableEventTypes) : null;

    const unsubscribe = subscribe<OrganizationEventPayload>(
      {
        query: toQueryString(ORGANIZATION_EVENT_SUBSCRIPTION),
        variables: { orgId: organizationId },
      },
      {
        next: (data) => {
          if (!data) return;

          // Keep REST-backed dashboards in sync.
          queryClient.invalidateQueries({ queryKey: queryKeys.dashboardStats(organizationId ?? '') });
          queryClient.invalidateQueries({ queryKey: ['events'] });

          const eventType = data.type as WSEventType;
          if (filterTypes && !filterTypes.has(eventType)) return;
          if (deviceImei && data.deviceId && data.deviceId !== deviceImei) return;

          const event: WSEvent = {
            id: `evt-${Date.now()}-${eventIdCounter++}`,
            type: eventType,
            deviceId: data.deviceId ?? '',
            timestamp: data.timestamp ? new Date(data.timestamp) : new Date(),
            data: data.data as Record<string, unknown> | undefined,
          };

          eventsRef.current = [event, ...eventsRef.current].slice(0, MAX_EVENTS);
          setEvents(eventsRef.current);
          setUnreadCount((count) => count + 1);
        },
      },
    );
    return () => {
      unsubscribe();
      eventsRef.current = [];
      setEvents([]);
      setUnreadCount(0);
    };
  }, [organizationId, subscribe, queryClient, stableEventTypes, deviceImei]);

  const markAsRead = useCallback(() => {
    setUnreadCount(0);
  }, []);

  const clearEvents = useCallback(() => {
    eventsRef.current = [];
    setEvents([]);
    setUnreadCount(0);
  }, []);

  return { events, unreadCount, markAsRead, clearEvents };
}

/**
 * Dispatch commands to a device and track their delivery status over the
 * realtime channel (spec §8.4).
 *
 * `sendCommand` dispatches via the REST endpoint (the authoritative path) and
 * returns the `dispatchId`. The hook subscribes to command-status events for
 * the target device over WS and updates `commandStatus` + `pendingCommands` as
 * acks/status updates arrive.
 */
interface UseCommandDispatchOptions {
  imei: string;
}

interface UseCommandDispatchResult {
  sendCommand: (
    command: WSCommandType,
    parameters?: Record<string, unknown>,
  ) => Promise<string>;
  pendingCommands: { dispatchId: string; command: WSCommandType; createdAt: Date }[];
  commandStatus: Map<string, 'pending' | 'delivered' | 'failed'>;
}

interface CommandStatusPayload {
  dispatchId: string;
  commandId: string;
  deviceId: string;
  command: string;
  status: string;
  createdAt: string;
}

export function useCommandDispatch(
  options: UseCommandDispatchOptions,
): UseCommandDispatchResult {
  const { imei } = options;
  const queryClient = useQueryClient();
  const subscribe = useWebSocketStore((s) => s.subscribe);
  const organizationId = useCurrentOrganizationId();

  const [pending, setPending] = useState<
    { dispatchId: string; command: WSCommandType; createdAt: Date }[]
  >([]);
  const [commandStatus, setCommandStatus] = useState<
    Map<string, 'pending' | 'delivered' | 'failed'>
  >(new Map());

  // Subscribe to command-status updates for this device.
  useEffect(() => {
    if (!organizationId || !imei) return;

    const unsubscribe = subscribe<CommandStatusPayload>(
      {
        query: toQueryString(COMMAND_STATUS_SUBSCRIPTION),
        variables: { deviceId: imei },
      },
      {
        next: (data) => {
          if (!data?.dispatchId) return;

          const status = mapCommandStatus(data.status);
          if (!status) return;

          setCommandStatus((prev) => {
            const next = new Map(prev);
            next.set(data.dispatchId, status);
            return next;
          });

          queryClient.invalidateQueries({ queryKey: queryKeys.command(data.dispatchId) });
          queryClient.invalidateQueries({ queryKey: queryKeys.pendingCommands(data.deviceId) });

          // Remove from pending once it reaches a terminal-ish state.
          if (status === 'delivered' || status === 'failed') {
            setPending((prev) => prev.filter((p) => p.dispatchId !== data.dispatchId));
          }
        },
      },
    );
    return () => {
      unsubscribe();
    };
  }, [imei, organizationId, subscribe, queryClient]);

  const sendCommand = useCallback(
    async (
      command: WSCommandType,
      parameters?: Record<string, unknown>,
    ): Promise<string> => {
      if (!imei) throw new Error('Cannot dispatch command: no device IMEI');

      // Dispatch via the authoritative REST path; status arrives over WS.
      const sent = await getCommands().postDeviceImeiCommand(imei, {
        command,
        args: parameters,
      });

      const dispatchId = sent.dispatchId ?? '';
      const createdAt = new Date(sent.serverTime ?? Date.now());

      setPending((prev) => [
        ...prev,
        { dispatchId, command, createdAt },
      ]);
      setCommandStatus((prev) => {
        const next = new Map(prev);
        next.set(dispatchId, 'pending');
        return next;
      });

      return dispatchId;
    },
    [imei, organizationId],
  );

  return { sendCommand, pendingCommands: pending, commandStatus };
}

function mapCommandStatus(
  status: string,
): 'pending' | 'delivered' | 'failed' | null {
  switch (status) {
    case 'pending':
      return 'pending';
    case 'delivered':
      return 'delivered';
    case 'completed':
      return 'delivered';
    case 'failed':
    case 'cancelled':
      return 'failed';
    default:
      return null;
  }
}
