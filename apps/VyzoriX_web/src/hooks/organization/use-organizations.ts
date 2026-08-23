import { useQuery, useMutation, useQueryClient, type UseQueryOptions } from '@tanstack/react-query';
import { getOrganizations,
  getMembers,
  getInvitations,
  type Organization,
  type OrganizationMember,
  type Invitation,
  type OrganizationRole,
  type CreateOrganizationRequest,
  type UpdateOrganizationRequest,
  type UpdateMemberRoleRequest,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';

interface CreateInvitationInput {
  email: string;
  role: OrganizationRole;
}

export function useOrganizations(
  options?: Omit<UseQueryOptions<Organization[]>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.organizations,
    queryFn: async () => (await getOrganizations().getOrganizations()).organizations ?? [],
    ...options,
  });
}

export function useOrganization(
  id: string | undefined,
  options?: Omit<UseQueryOptions<Organization>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.organization(id ?? ''),
    queryFn: () => getOrganizations().getOrganizationsId(id!),
    enabled: id !== undefined && id !== '',
    ...options,
  });
}

export function useCreateOrganization() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (request: CreateOrganizationRequest) => getOrganizations().postOrganizations(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations });
    },
  });
}

export function useUpdateOrganization() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, request }: { id: string; request: UpdateOrganizationRequest }) =>
      getOrganizations().patchOrganizationsId(id, request),
    onSuccess: (updated, { id }) => {
      queryClient.setQueryData(queryKeys.organization(id), updated);
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations });
    },
  });
}

export function useDeleteOrganization() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => getOrganizations().deleteOrganizationsId(id),
    onSuccess: (_, id) => {
      queryClient.removeQueries({ queryKey: queryKeys.organization(id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations });
    },
  });
}

export function useOrganizationMembers(
  orgId: string | undefined,
  options?: Omit<UseQueryOptions<OrganizationMember[]>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.organizationMembers(orgId ?? ''),
    queryFn: async () => (await getMembers().getOrganizationsIdMembers(orgId!)).members ?? [],
    enabled: orgId !== undefined && orgId !== '',
    ...options,
  });
}

export function useUpdateMemberRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ orgId, memberId, request }: { orgId: string; memberId: string; request: UpdateMemberRoleRequest }) =>
      getMembers().patchOrganizationsIdMembersMemberIdRole(orgId, memberId, request),
    onSuccess: (updated, { orgId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizationMembers(orgId) });
      queryClient.setQueryData(['organizations', orgId, 'members', updated.id], updated);
    },
  });
}

export function useRemoveMember() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ orgId, memberId }: { orgId: string; memberId: string }) =>
      getMembers().deleteOrganizationsIdMembersMemberId(orgId, memberId),
    onSuccess: (_, { orgId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizationMembers(orgId) });
    },
  });
}

export function useSuspendMember() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ orgId, memberId }: { orgId: string; memberId: string }) =>
      getMembers().postOrganizationsIdMembersMemberIdSuspend(orgId, memberId),
    onSuccess: (_, { orgId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizationMembers(orgId) });
    },
  });
}

export function useReinstateMember() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ orgId, memberId }: { orgId: string; memberId: string }) =>
      getMembers().postOrganizationsIdMembersMemberIdReinstate(orgId, memberId),
    onSuccess: (_, { orgId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizationMembers(orgId) });
    },
  });
}

export function useTransferOwnership() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ orgId, memberId }: { orgId: string; memberId: string }) =>
      getMembers().postOrganizationsIdMembersMemberIdTransfer(orgId, memberId),
    onSuccess: (_, { orgId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizationMembers(orgId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.organization(orgId) });
    },
  });
}

export function useOrganizationInvitations(
  orgId: string | undefined,
  status?: string,
  options?: Omit<UseQueryOptions<Invitation[]>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.organizationInvitations(orgId ?? '', status),
    queryFn: async () =>
      (await getInvitations().getOrganizationsIdInvitations(orgId!, status ? { status } : undefined))
        .invitations ?? [],
    enabled: orgId !== undefined && orgId !== '',
    ...options,
  });
}

export function useCreateInvitation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateInvitationInput) => getInvitations().postInvitations(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
    },
  });
}

export function useAcceptInvitation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ token }: { token: string }) =>
      getInvitations().postInviteTokenAccept(token),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations });
      queryClient.invalidateQueries({ queryKey: queryKeys.myInvitations });
    },
  });
}

export function useRejectInvitation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ token }: { token: string }) =>
      getInvitations().postInviteTokenReject(token),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.myInvitations });
    },
  });
}

export type {
  Organization,
  OrganizationMember,
  Invitation,
  OrganizationRole,
  CreateOrganizationRequest,
  UpdateOrganizationRequest,
};
