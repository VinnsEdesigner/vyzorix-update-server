import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
} from '@tanstack/react-query';
import { getAuth, getInvitations } from '@vyzorix/api-client';
import type { MeResult, OrganizationInfo, SelectOrganizationResult, InvitationListResult } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useAuthStore } from '@/stores';

type MeQueryOptions = Omit<UseQueryOptions<MeResult>, 'queryKey' | 'queryFn'>;

export function useMe(options?: MeQueryOptions) {
  return useQuery({
    queryKey: queryKeys.me,
    queryFn: () => getAuth().getAuthMe(),
    ...options,
  });
}

export function useMyOrganizations(
  options?: Omit<UseQueryOptions<{ organizations: OrganizationInfo[] }>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.myOrganizations,
    queryFn: async () => {
      const result = await getAuth().getAuthOrganizations();
      return { organizations: result.organizations ?? [] };
    },
    ...options,
  });
}

export function useMyInvitations(
  options?: UseQueryOptions<InvitationListResult>,
) {
  return useQuery({
    queryKey: queryKeys.myInvitations,
    queryFn: () => getInvitations().getMeInvitations(),
    ...options,
  });
}

export function useSelectOrganization() {
  const queryClient = useQueryClient();
  const setOrganization = useAuthStore((s) => s.setOrganization);

  return useMutation({
    mutationFn: (organizationId: string) =>
      getAuth().postAuthOrganizationsSelect({ organization_id: organizationId }),
    onSuccess: (org: SelectOrganizationResult) => {
      setOrganization(org.organization_id ?? null);
      queryClient.invalidateQueries({ queryKey: queryKeys.me });
      queryClient.removeQueries({ queryKey: ['devices'] });
      queryClient.removeQueries({ queryKey: ['dashboard'] });
      queryClient.removeQueries({ queryKey: ['metrics'] });
      queryClient.removeQueries({ queryKey: ['events'] });
      queryClient.removeQueries({ queryKey: ['commands'] });
      queryClient.removeQueries({ queryKey: ['logs'] });
    },
  });
}
