import { useQueryClient } from '@tanstack/react-query';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import {
  useGetDevices,
  useGetDeviceImei,
  useGetDeviceCount,
  useGetDevicesImeiSettings,
  usePatchDevicesImeiSettings,
  useGetDevicesImeiSettingsThresholds,
  usePatchDevicesImeiSettingsThresholds,
  useGetDeviceImeiConnectionStatus,
  usePostDeviceImeiDisconnect,
  useDeleteDevicesImei,
  useGetDevicesImeiTags,
  usePutDevicesImeiTags,
  usePostOrganizationsIdDevicesImeiTransfer,
} from '@/generated-rq/devices/device-management';

export interface DeviceParams {
  page?: number;
  limit?: number;
  search?: string;
}

export function useDevices(params?: DeviceParams) {
  const organizationId = useCurrentOrganizationId();
  return useGetDevices(
    { page: params?.page, limit: params?.limit, search: params?.search },
    { query: { queryKey: ['devices', params, organizationId] as const, enabled: organizationId !== null } },
  );
}

export function useDevice(imei: string | undefined) {
  const organizationId = useCurrentOrganizationId();
  return useGetDeviceImei(
    imei ?? '',
    { query: { queryKey: ['devices', imei, organizationId] as const, enabled: imei !== undefined && imei !== '' && organizationId !== null } },
  );
}

export function useDeviceCount() {
  const organizationId = useCurrentOrganizationId();
  return useGetDeviceCount({ query: { queryKey: ['devices', 'count', organizationId] as const, enabled: organizationId !== null } });
}

export function useDeviceSettings(imei: string | undefined) {
  return useGetDevicesImeiSettings(
    imei ?? '',
    { query: { queryKey: ['devices', imei, 'settings'] as const, enabled: imei !== undefined } },
  );
}

export function useUpdateDeviceSettings() {
  const queryClient = useQueryClient();
  return usePatchDevicesImeiSettings({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['devices'] }) },
  });
}

export function useDeviceThresholds(imei: string | undefined) {
  return useGetDevicesImeiSettingsThresholds(
    imei ?? '',
    { query: { queryKey: ['devices', imei, 'thresholds'] as const, enabled: imei !== undefined } },
  );
}

export function useUpdateDeviceThresholds() {
  const queryClient = useQueryClient();
  return usePatchDevicesImeiSettingsThresholds({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['devices'] }) },
  });
}

export function useDeviceConnectionStatus(imei: string | undefined) {
  return useGetDeviceImeiConnectionStatus(
    imei ?? '',
    { query: { queryKey: ['devices', imei, 'connection'] as const, enabled: imei !== undefined } },
  );
}

export function useDisconnectDevice() {
  const queryClient = useQueryClient();
  return usePostDeviceImeiDisconnect({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['devices'] }) },
  });
}

export function useDeviceTags(imei: string | undefined) {
  return useGetDevicesImeiTags(
    imei ?? '',
    { query: { queryKey: ['devices', imei, 'tags'] as const, enabled: imei !== undefined } },
  );
}

export function useSetDeviceTags() {
  const queryClient = useQueryClient();
  return usePutDevicesImeiTags({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['devices'] }) },
  });
}

export function useDeregisterDevice() {
  const queryClient = useQueryClient();
  return useDeleteDevicesImei({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['devices'] }) },
  });
}

export function useTransferDevice() {
  const queryClient = useQueryClient();
  return usePostOrganizationsIdDevicesImeiTransfer({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['devices'] }) },
  });
}
