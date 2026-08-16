import { useQuery, useMutation, useQueryClient, type UseQueryOptions } from '@tanstack/react-query';
import {
  orgSettings,
  type OrganizationSettings,
  type ThresholdUpdateRequest,
  type SettingsUpdateRequest,
  type Thresholds,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';

export function useOrgSettings(
  orgId: string | undefined,
  options?: Omit<UseQueryOptions<OrganizationSettings>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.orgSettings(orgId ?? ''),
    queryFn: () => orgSettings.get(orgId!),
    enabled: orgId !== undefined && orgId !== '',
    ...options,
  });
}

export function useUpdateOrgSettings() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ orgId, request }: { orgId: string; request: SettingsUpdateRequest }) =>
      orgSettings.update(orgId, request),
    onSuccess: (updated, { orgId }) => {
      queryClient.setQueryData(queryKeys.orgSettings(orgId), updated);
      if (updated.defaultThresholds) {
        queryClient.setQueryData(
          queryKeys.orgThresholds(orgId),
          { thresholds: updated.defaultThresholds },
        );
      }
    },
  });
}

export function useOrgThresholds(
  orgId: string | undefined,
  options?: Omit<UseQueryOptions<{ thresholds: Thresholds }>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.orgThresholds(orgId ?? ''),
    queryFn: () => orgSettings.getThresholds(orgId!),
    enabled: orgId !== undefined && orgId !== '',
    ...options,
  });
}

export function useUpdateOrgThresholds(orgId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: ThresholdUpdateRequest) => orgSettings.updateThresholds(orgId, data),
    onSuccess: (updated) => {
      queryClient.setQueryData(queryKeys.orgThresholds(orgId), updated);
      queryClient.invalidateQueries({ queryKey: queryKeys.orgSettings(orgId) });
    },
  });
}

export type { OrganizationSettings, ThresholdUpdateRequest, SettingsUpdateRequest, Thresholds };
