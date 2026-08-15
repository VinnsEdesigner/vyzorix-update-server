import { useEffect } from 'react';
import { useWebSocketStore, useDiagnosticsStore, useTimelineStreamStore } from '@/stores';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import type {
  TimelineEvent,
  TimelineEventType,
} from '@vyzorix/api-client';

interface SubscriptionDocument {
  loc?: { source: { body: string } };
}

interface RawDeviceUpdated {
  id?: string;
  imei?: string;
  deviceName?: string;
  status?: string;
  lastSeen?: string | null;
}

interface RawTelemetryReceived {
  id: string;
  deviceId?: string;
  receivedAt?: string | null;
  riskScore?: number;
  bufferLevel?: number;
  thermalTemp?: number;
  payload?: Record<string, unknown>;
}

interface RawOrganizationEvent {
  type: string;
  timestamp: string;
  data?: Record<string, unknown>;
}

function toQueryString(doc: unknown): string {
  return (doc as SubscriptionDocument)?.loc?.source?.body ?? '';
}

function toDate(value?: string | null): Date {
  if (!value) return new Date();
  const d = new Date(value);
  return Number.isNaN(d.getTime()) ? new Date() : d;
}

/**
 * Subscribes to realtime device updates + telemetry for a single device, patching the
 * `diagnostics-store` (live connection/telemetry snapshot) and the `timeline-stream-store`
 * (live event prepend). Org-gated: only subscribes when `organizationId !== null`.
 */
export function useDiagnosticStream(imei: string | undefined) {
  const subscribe = useWebSocketStore((s) => s.subscribe);
  const organizationId = useCurrentOrganizationId();

  const setDiagnosticsOrg = useDiagnosticsStore((s) => s.setActiveOrganization);
  const patchConnection = useDiagnosticsStore((s) => s.patchConnection);
  const patchTelemetry = useDiagnosticsStore((s) => s.patchTelemetry);

  const setStreamOrg = useTimelineStreamStore((s) => s.setActiveOrganization);
  const appendEvent = useTimelineStreamStore((s) => s.append);
  const clearStream = useTimelineStreamStore((s) => s.clear);

  useEffect(() => {
    setDiagnosticsOrg(organizationId);
    setStreamOrg(organizationId);
  }, [organizationId, setDiagnosticsOrg, setStreamOrg]);

  useEffect(() => {
    if (!organizationId || !imei) return;

    const unsubDevice = subscribe<RawDeviceUpdated>(
      {
        query: toQueryString(DEVICE_UPDATED_QUERY),
        variables: { deviceId: imei },
      },
      {
        next: (data) => {
          if (!data) return;
          patchConnection(organizationId, imei, {
            webSocketStatus:
              data.status === 'connected' ? 'connected' : 'disconnected',
            lastSeen: data.lastSeen ? toDate(data.lastSeen) : undefined,
          });
        },
      },
    );

    const unsubTelemetry = subscribe<RawTelemetryReceived>(
      {
        query: toQueryString(TELEMETRY_RECEIVED_QUERY),
        variables: { deviceId: imei },
      },
      {
        next: (data) => {
          if (!data?.id) return;
          const ts = data.receivedAt ? toDate(data.receivedAt) : new Date();
          patchTelemetry(organizationId, imei, {
            lastTimestamp: ts,
            avgLatencyMs: typeof data.bufferLevel === 'number' ? data.bufferLevel : undefined,
          });
          const event: TimelineEvent = {
            id: data.id,
            deviceId: imei,
            type: 'TELEMETRY' as TimelineEventType,
            timestamp: ts,
            data: {
              riskScore: data.riskScore,
              bufferLevel: data.bufferLevel,
              thermalTemp: data.thermalTemp,
              payload: data.payload,
            },
          };
          appendEvent(imei, event);
        },
      },
    );

    return () => {
      unsubDevice();
      unsubTelemetry();
      clearStream(imei);
    };
  }, [
    imei,
    organizationId,
    subscribe,
    patchConnection,
    patchTelemetry,
    appendEvent,
    clearStream,
  ]);
}

// Inline subscription query strings (avoid importing gql-tagged docs that may not be
// re-exported; keep the WS payload minimal and explicit).
const DEVICE_UPDATED_QUERY = `
  subscription OnDeviceUpdated($deviceId: ID) {
    deviceUpdated(deviceId: $deviceId) {
      id
      imei
      deviceName
      status
      lastSeen
    }
  }
`;

const TELEMETRY_RECEIVED_QUERY = `
  subscription OnTelemetryReceived($deviceId: ID) {
    telemetryReceived(deviceId: $deviceId) {
      id
      deviceId
      receivedAt
      riskScore
      bufferLevel
      thermalTemp
      payload
    }
  }
`;

export type { RawOrganizationEvent };
