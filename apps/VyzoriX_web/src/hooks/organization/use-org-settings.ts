import {
  useGetOrganizationsIdSettings,
  usePatchOrganizationsIdSettings,
  useGetOrganizationsIdSettingsThresholds,
  usePatchOrganizationsIdSettingsThresholds,
} from '@/generated-rq/organizations/organization-management';

export function useOrgSettings(id: string | undefined) {
  return useGetOrganizationsIdSettings(
    id ?? '',
    { query: { enabled: id !== undefined && id !== '' } },
  );
}

export function useUpdateOrgSettings() {
  return usePatchOrganizationsIdSettings();
}

export function useOrgThresholds(id: string | undefined) {
  return useGetOrganizationsIdSettingsThresholds(
    id ?? '',
    { query: { enabled: id !== undefined && id !== '' } },
  );
}

export function useUpdateOrgThresholds() {
  return usePatchOrganizationsIdSettingsThresholds();
}
