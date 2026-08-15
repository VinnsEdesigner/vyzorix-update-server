import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
} from '@tanstack/react-query';
import {
  devices,
  type Device,
  type DeviceListResult,
  type DeviceParams,
  type DeviceStats,
  type DeviceSettings,
  type ConnectionStatus,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export function useDevices(params?: DeviceParams) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.devices({ ...params, organizationId }),
    queryFn: () => devices.list({ ...params, organizationId: organizationId ?? undefined }),
    enabled: organizationId !== null,
  });
}

export function useDevice(
  imei: string | undefined,
  options?: Omit<UseQueryOptions<Device | null>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.device(imei ?? ''),
    queryFn: () => devices.get(imei!, organizationId ?? undefined),
    enabled: imei !== undefined && imei !== '',
    ...options,
  });
}

export function useDeviceConnectionStatus(imei: string | undefined) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.deviceConnectionStatus(imei ?? ''),
    queryFn: () => devices.getConnectionStatus(imei!, organizationId ?? undefined),
    enabled: imei !== undefined && imei !== '',
    refetchInterval: 15_000,
  });
}

export function useDeviceSettings(imei: string | undefined) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.deviceSettings(imei ?? ''),
    queryFn: () => devices.getSettings(imei!, organizationId ?? undefined),
    enabled: imei !== undefined && imei !== '',
  });
}

export function useUpdateDeviceSettings(imei: string) {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: (settings: Partial<DeviceSettings>) =>
      devices.updateSettings(imei, settings, organizationId ?? undefined),
    onSuccess: (updated) => {
      queryClient.setQueryData(queryKeys.deviceSettings(imei), updated);
    },
  });
}

export function useDeviceCount() {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.deviceCount,
    queryFn: () => devices.count(organizationId ?? undefined),
    enabled: organizationId !== null,
  });
}

export function useDeviceStats(options?: Omit<UseQueryOptions<DeviceStats>, 'queryKey' | 'queryFn'>) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.deviceStats,
    queryFn: () => devices.stats(organizationId ?? undefined),
    enabled: organizationId !== null,
    ...options,
  });
}

export function useDeregisterDevice() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: (imei: string) => devices.deregister(imei, organizationId ?? undefined),
    onSuccess: (_, imei) => {
      queryClient.invalidateQueries({ queryKey: ['devices'] });
      queryClient.removeQueries({ queryKey: queryKeys.device(imei) });
    },
  });
}

export function useDisconnectDevice() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: (imei: string) => devices.disconnect(imei, organizationId ?? undefined),
    onSuccess: (_, imei) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.deviceConnectionStatus(imei) });
    },
  });
}

export type { DeviceListResult, Device, DeviceStats, DeviceSettings, ConnectionStatus };
