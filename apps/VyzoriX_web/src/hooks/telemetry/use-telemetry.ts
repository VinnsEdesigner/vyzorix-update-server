import {
  useGetTelemetryHistory,
  useGetTelemetryLatestDeviceId,
  useGetTelemetryStatsDeviceId,
  useGetTelemetryHistoryExport,
} from '@/generated-rq/telemetry/device-telemetry';

export function useTelemetryHistory(deviceId: string | undefined, params?: { start?: number; end?: number; limit?: number }) {
  return useGetTelemetryHistory(
    { deviceId: deviceId ?? '', startTime: params?.start, endTime: params?.end, limit: params?.limit },
    { query: { queryKey: ['telemetry', deviceId, params] as const, enabled: deviceId !== undefined } },
  );
}

export function useLatestTelemetry(deviceId: string | undefined) {
  return useGetTelemetryLatestDeviceId(
    deviceId ?? '',
    { query: { queryKey: ['telemetry', deviceId, 'latest'] as const, enabled: deviceId !== undefined } },
  );
}

export function useTelemetryStats(deviceId: string | undefined) {
  return useGetTelemetryStatsDeviceId(
    deviceId ?? '',
    { query: { queryKey: ['telemetry', deviceId, 'stats'] as const, enabled: deviceId !== undefined } },
  );
}

export function useExportTelemetry(deviceId: string | undefined) {
  return useGetTelemetryHistoryExport(
    { deviceId: deviceId ?? '' },
    { query: { queryKey: ['telemetry', deviceId, 'export'] as const, enabled: deviceId !== undefined } },
  );
}
