import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
} from '@tanstack/react-query';
import { me, type MeResponse, type OrganizationInfo } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useAuthStore } from '@/stores';

type MeQueryOptions = Omit<UseQueryOptions<MeResponse>, 'queryKey' | 'queryFn'>;

export function useMe(options?: MeQueryOptions) {
  return useQuery({
    queryKey: queryKeys.me,
    queryFn: () => me.getMe(),
    ...options,
  });
}

export function useMyOrganizations(
  options?: Omit<UseQueryOptions<{ organizations: OrganizationInfo[] }>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.myOrganizations,
    queryFn: () => me.getOrganizations(),
    ...options,
  });
}

export function useMyInvitations(
  options?: UseQueryOptions<Awaited<ReturnType<typeof me.getInvitations>>>,
) {
  return useQuery({
    queryKey: queryKeys.myInvitations,
    queryFn: () => me.getInvitations(),
    ...options,
  });
}

export function useSelectOrganization() {
  const queryClient = useQueryClient();
  const setOrganization = useAuthStore((s) => s.setOrganization);

  return useMutation({
    mutationFn: (organizationId: string) =>
      me.selectOrganization({ organization_id: organizationId }),
    onSuccess: (org: OrganizationInfo) => {
      setOrganization(org.id);
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
