import { useAuthStore, useDeviceSelectorStore } from '@/stores';

export function useCurrentOrganizationId(): string | null {
  return useAuthStore((s) => s.organizationId);
}

export function useRequiredOrganizationId(): string {
  const orgId = useAuthStore((s) => s.organizationId);
  if (!orgId) {
    throw new Error('No organization selected. Cannot perform organization-scoped query.');
  }
  return orgId;
}

export function useSelectedImei(): string | null {
  return useDeviceSelectorStore((s) => s.selectedDevice?.imei ?? null);
}

export function useRequiredImei(): string {
  const imei = useDeviceSelectorStore((s) => s.selectedDevice?.imei);
  if (!imei) {
    throw new Error('No device selected. Cannot perform device-scoped query.');
  }
  return imei;
}
