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
    { query: { enabled: organizationId !== null } },
  );
}

export function useDevice(imei: string | undefined) {
  const organizationId = useCurrentOrganizationId();
  return useGetDeviceImei(imei ?? '', {
    query: { enabled: imei !== undefined && imei !== '' && organizationId !== null },
  });
}

export function useDeviceCount() {
  const organizationId = useCurrentOrganizationId();
  return useGetDeviceCount({ query: { enabled: organizationId !== null } });
}

export function useDeviceSettings(imei: string | undefined) {
  return useGetDevicesImeiSettings(imei ?? '', {
    query: { enabled: imei !== undefined },
  });
}

export function useUpdateDeviceSettings() {
  return usePatchDevicesImeiSettings();
}

export function useDeviceThresholds(imei: string | undefined) {
  return useGetDevicesImeiSettingsThresholds(imei ?? '', {
    query: { enabled: imei !== undefined },
  });
}

export function useUpdateDeviceThresholds() {
  return usePatchDevicesImeiSettingsThresholds();
}

export function useDeviceConnectionStatus(imei: string | undefined) {
  return useGetDeviceImeiConnectionStatus(imei ?? '', {
    query: { enabled: imei !== undefined },
  });
}

export function useDisconnectDevice() {
  return usePostDeviceImeiDisconnect();
}

export function useDeviceTags(imei: string | undefined) {
  return useGetDevicesImeiTags(imei ?? '', {
    query: { enabled: imei !== undefined },
  });
}

export function useSetDeviceTags() {
  return usePutDevicesImeiTags();
}

export function useDeregisterDevice() {
  return useDeleteDevicesImei();
}

export function useTransferDevice() {
  return usePostOrganizationsIdDevicesImeiTransfer();
}
