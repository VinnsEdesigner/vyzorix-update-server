import { useQueryClient } from '@tanstack/react-query';
import {
  useGetOrganizationsIdSettings,
  usePatchOrganizationsIdSettings,
  useGetOrganizationsIdSettingsThresholds,
  usePatchOrganizationsIdSettingsThresholds,
} from '@/generated-rq/organizations/organization-management';

export function useOrgSettings(id: string | undefined) {
  return useGetOrganizationsIdSettings(
    id ?? '',
    { query: { queryKey: ['org-settings', id] as const, enabled: id !== undefined && id !== '' } },
  );
}

export function useUpdateOrgSettings() {
  const queryClient = useQueryClient();
  return usePatchOrganizationsIdSettings({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org-settings'] }) },
  });
}

export function useOrgThresholds(id: string | undefined) {
  return useGetOrganizationsIdSettingsThresholds(
    id ?? '',
    { query: { queryKey: ['org-thresholds', id] as const, enabled: id !== undefined && id !== '' } },
  );
}

export function useUpdateOrgThresholds() {
  const queryClient = useQueryClient();
  return usePatchOrganizationsIdSettingsThresholds({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org-thresholds'] }) },
  });
}
