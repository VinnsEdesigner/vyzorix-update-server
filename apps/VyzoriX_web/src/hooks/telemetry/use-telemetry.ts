import {
  useGetTelemetryHistory,
  useGetTelemetryLatestDeviceId,
  useGetTelemetryStatsDeviceId,
  useGetTelemetryHistoryExport,
  getGetTelemetryHistoryQueryKey,
  getGetTelemetryLatestDeviceIdQueryKey,
  getGetTelemetryStatsDeviceIdQueryKey,
  getGetTelemetryHistoryExportQueryKey,
} from '@/generated-rq/telemetry/device-telemetry';

export function useTelemetryHistory(deviceId: string | undefined, params?: { start?: number; end?: number; limit?: number }) {
  const keyParams = { deviceId: deviceId ?? '', startTime: params?.start, endTime: params?.end, limit: params?.limit };
  return useGetTelemetryHistory(
    keyParams,
    {
      query: {
        queryKey: getGetTelemetryHistoryQueryKey(keyParams),
        enabled: deviceId !== undefined,
      },
    },
  );
}

export function useLatestTelemetry(deviceId: string | undefined) {
  return useGetTelemetryLatestDeviceId(
    deviceId ?? '',
    {
      query: {
        queryKey: getGetTelemetryLatestDeviceIdQueryKey(deviceId ?? ''),
        enabled: deviceId !== undefined,
      },
    },
  );
}

export function useTelemetryStats(deviceId: string | undefined) {
  return useGetTelemetryStatsDeviceId(
    deviceId ?? '',
    {
      query: {
        queryKey: getGetTelemetryStatsDeviceIdQueryKey(deviceId ?? ''),
        enabled: deviceId !== undefined,
      },
    },
  );
}

export function useExportTelemetry(deviceId: string | undefined) {
  const keyParams = { deviceId: deviceId ?? '' };
  return useGetTelemetryHistoryExport(
    keyParams,
    {
      query: {
        queryKey: getGetTelemetryHistoryExportQueryKey(keyParams),
        enabled: deviceId !== undefined,
      },
    },
  );
}
