import {
  useGetOrganizations,
  usePostOrganizations,
  useGetOrganizationsId,
  usePatchOrganizationsId,
  useDeleteOrganizationsId,
} from '@/generated-rq/organizations/organization-management';

export function useOrganizations() {
  return useGetOrganizations();
}

export function useCreateOrganization() {
  return usePostOrganizations();
}

export function useOrganization(id: string | undefined) {
  return useGetOrganizationsId(id ?? '', { query: { enabled: id !== undefined } });
}

export function useUpdateOrganization() {
  return usePatchOrganizationsId();
}

export function useDeleteOrganization() {
  return useDeleteOrganizationsId();
}
