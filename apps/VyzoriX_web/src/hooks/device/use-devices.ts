import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
} from '@tanstack/react-query';
import { getDevices,
  type DeviceListItem,
  type DeviceListResult,
  type GetTelemetryResponse,
  type DeviceSettingsResult,
  type UpdateDeviceSettingsRequest,
  type ThresholdUpdateRequest,
  type ConnectionStatusResult,
  type Thresholds,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export interface DeviceParams {
  page?: number;
  limit?: number;
  search?: string;
}

export function useDevices(params?: DeviceParams) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.devices({ ...params, organizationId }),
    queryFn: () => getDevices().getDevices({ ...params}),
    enabled: organizationId !== null,
  });
}

export function useDevice(
  imei: string | undefined,
  options?: Omit<UseQueryOptions<DeviceListItem | null>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.device(imei ?? ''),
    queryFn: () => getDevices().getDevicesImei(imei!),
    enabled: imei !== undefined && imei !== '',
    ...options,
  });
}

export function useDeviceConnectionStatus(imei: string | undefined) {
  return useQuery({
    queryKey: queryKeys.deviceConnectionStatus(imei ?? ''),
    queryFn: () => getDevices().getDeviceImeiConnectionStatus(imei!),
    enabled: imei !== undefined && imei !== '',
    refetchInterval: 15_000,
  });
}

export function useDeviceSettings(imei: string | undefined) {
  return useQuery({
    queryKey: queryKeys.deviceSettings(imei ?? ''),
    queryFn: () => getDevices().getDevicesImeiSettings(imei!),
    enabled: imei !== undefined && imei !== '',
  });
}

export function useUpdateDeviceSettings(imei: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (settings: UpdateDeviceSettingsRequest) =>
      getDevices().patchDevicesImeiSettings(imei, settings),
    onSuccess: (updated) => {
      queryClient.setQueryData(queryKeys.deviceSettings(imei), updated);
    },
  });
}

export function useDeviceThresholds(imei: string | undefined) {
  return useQuery({
    queryKey: queryKeys.deviceThresholds(imei ?? ''),
    queryFn: () => getDevices().getDevicesImeiSettingsThresholds(imei!),
    enabled: imei !== undefined && imei !== '',
  });
}

export function useUpdateDeviceThresholds(imei: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (thresholds: ThresholdUpdateRequest) =>
      getDevices().patchDevicesImeiSettingsThresholds(imei, thresholds),
    onSuccess: (updated) => {
      queryClient.setQueryData(queryKeys.deviceThresholds(imei), updated);
      queryClient.invalidateQueries({ queryKey: queryKeys.deviceSettings(imei) });
    },
  });
}

export function useDeviceCount() {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.deviceCount,
    queryFn: () => getDevices().getDeviceCount(),
    enabled: organizationId !== null,
  });
}

export function useGetTelemetryResponse(
  imei: string | undefined,
  options?: Omit<UseQueryOptions<GetTelemetryResponse>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.deviceStats,
    queryFn: () => getDevices().getDashboardDeviceImeiMetrics(imei!),
    enabled: imei !== undefined && imei !== '',
    ...options,
  });
}

export function useDeregisterDevice() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (imei: string) => getDevices().deleteDevicesImei(imei),
    onSuccess: (_, imei) => {
      queryClient.invalidateQueries({ queryKey: ['devices'] });
      queryClient.removeQueries({ queryKey: queryKeys.device(imei) });
    },
  });
}

export function useDisconnectDevice() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (imei: string) => getDevices().postDeviceImeiDisconnect(imei),
    onSuccess: (_, imei) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.deviceConnectionStatus(imei) });
    },
  });
}

export type { DeviceListResult, DeviceListItem, GetTelemetryResponse, DeviceSettingsResult, UpdateDeviceSettingsRequest, ThresholdUpdateRequest, ConnectionStatusResult, Thresholds };
