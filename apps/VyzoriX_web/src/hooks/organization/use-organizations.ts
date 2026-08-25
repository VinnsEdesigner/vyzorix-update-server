import { useQueryClient } from '@tanstack/react-query';
import {
  useGetOrganizations,
  usePostOrganizations,
  useGetOrganizationsId,
  usePatchOrganizationsId,
  useDeleteOrganizationsId,
} from '@/generated-rq/organizations/organization-management';

export function useOrganizations() {
  return useGetOrganizations({ query: { queryKey: ['organizations'] as const } });
}

export function useCreateOrganization() {
  const queryClient = useQueryClient();
  return usePostOrganizations({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['organizations'] }) },
  });
}

export function useOrganization(id: string | undefined) {
  return useGetOrganizationsId(
    id ?? '',
    { query: { queryKey: ['organizations', id] as const, enabled: id !== undefined } },
  );
}

export function useUpdateOrganization() {
  const queryClient = useQueryClient();
  return usePatchOrganizationsId({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['organizations'] }) },
  });
}

export function useDeleteOrganization() {
  const queryClient = useQueryClient();
  return useDeleteOrganizationsId({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['organizations'] }) },
  });
}
