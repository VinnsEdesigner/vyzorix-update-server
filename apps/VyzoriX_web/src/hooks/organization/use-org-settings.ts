import { useQuery, useMutation, useQueryClient, type UseQueryOptions } from '@tanstack/react-query';
import {
  getOrganizations,
  type OrganizationSettingsResult,
  type ThresholdUpdateRequest,
  type UpdateOrganizationSettingsRequest,
  type OperatorThresholds,
  type ThresholdsResult,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';

export function useOrgSettings(
  orgId: string | undefined,
  options?: Omit<UseQueryOptions<OrganizationSettingsResult>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.orgSettings(orgId ?? ''),
    queryFn: () => getOrganizations().getOrganizationsIdSettings(orgId!),
    enabled: orgId !== undefined && orgId !== '',
    ...options,
  });
}

export function useUpdateOrgSettings() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ orgId, request }: { orgId: string; request: UpdateOrganizationSettingsRequest }) =>
      getOrganizations().patchOrganizationsIdSettings(orgId, request),
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
  options?: Omit<UseQueryOptions<ThresholdsResult>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.orgThresholds(orgId ?? ''),
    queryFn: () => getOrganizations().getOrganizationsIdSettingsThresholds(orgId!),
    enabled: orgId !== undefined && orgId !== '',
    ...options,
  });
}

export function useUpdateOrgThresholds(orgId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: ThresholdUpdateRequest) =>
      getOrganizations().patchOrganizationsIdSettingsThresholds(orgId, data),
    onSuccess: (updated) => {
      queryClient.setQueryData(queryKeys.orgThresholds(orgId), updated);
      queryClient.invalidateQueries({ queryKey: queryKeys.orgSettings(orgId) });
    },
  });
}

export type { OrganizationSettingsResult, ThresholdUpdateRequest, UpdateOrganizationSettingsRequest, OperatorThresholds, ThresholdsResult };
