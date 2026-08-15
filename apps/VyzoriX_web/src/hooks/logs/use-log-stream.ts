import { useEffect } from 'react';
import { useWebSocketStore, useLogStreamStore } from '@/stores';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { SUBSCRIBE_TO_LOGS, type LogEntry, type LogEventType } from '@vyzorix/api-client';

interface SubscriptionDocument {
  loc?: { source: { body: string } };
}

interface RawLogStreamEvent {
  id: string;
  type: string;
  timestamp: number;
  data?: Record<string, unknown>;
  deviceImei?: string;
}

function toQueryString(doc: unknown): string {
  return (doc as SubscriptionDocument)?.loc?.source?.body ?? '';
}

function parseTimestamp(value?: number | null): Date {
  if (!value) return new Date();
  return new Date(value > 1e12 ? value : value * 1000);
}

export interface UseLogStreamOptions {
  types?: LogEventType[];
}

export function useLogStream(deviceId: string | undefined, options?: UseLogStreamOptions) {
  const subscribe = useWebSocketStore((s) => s.subscribe);
  const append = useLogStreamStore((s) => s.append);
  const setActiveOrganization = useLogStreamStore((s) => s.setActiveOrganization);
  const clear = useLogStreamStore((s) => s.clear);
  const organizationId = useCurrentOrganizationId();

  useEffect(() => {
    setActiveOrganization(organizationId);
  }, [organizationId, setActiveOrganization]);

  useEffect(() => {
    if (!organizationId || !deviceId) return;
    const unsubscribe = subscribe<RawLogStreamEvent>(
      {
        query: toQueryString(SUBSCRIBE_TO_LOGS),
        variables: { deviceId, types: options?.types ?? null },
      },
      {
        next: (data) => {
          if (!data?.id) return;
          const entry: LogEntry = {
            id: data.id,
            deviceId: data.deviceImei ?? deviceId,
            eventType: (data.type as LogEventType) ?? 'info',
            timestamp: parseTimestamp(data.timestamp),
            data: data.data,
          };
          append(deviceId, entry);
        },
      },
    );
    return () => {
      unsubscribe();
      clear(deviceId);
    };
  }, [deviceId, organizationId, options?.types, subscribe, append, clear]);
}
