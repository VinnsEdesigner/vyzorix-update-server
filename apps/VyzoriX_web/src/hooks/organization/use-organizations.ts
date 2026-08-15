import { useQuery, useMutation, useQueryClient, type UseQueryOptions } from '@tanstack/react-query';
import {
  organizations,
  members,
  invitations,
  type Organization,
  type OrganizationMember,
  type Invitation,
  type OrganizationRole,
  type CreateOrganizationRequest,
  type UpdateOrganizationRequest,
  type UpdateMemberRoleRequest,
  type InvitationResponseRequest,
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
    queryFn: () => organizations.list(),
    ...options,
  });
}

export function useOrganization(
  id: string | undefined,
  options?: Omit<UseQueryOptions<Organization>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.organization(id ?? ''),
    queryFn: () => organizations.get(id!),
    enabled: id !== undefined && id !== '',
    ...options,
  });
}

export function useCreateOrganization() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (request: CreateOrganizationRequest) => organizations.create(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations });
    },
  });
}

export function useUpdateOrganization() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, request }: { id: string; request: UpdateOrganizationRequest }) =>
      organizations.update(id, request),
    onSuccess: (updated, { id }) => {
      queryClient.setQueryData(queryKeys.organization(id), updated);
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations });
    },
  });
}

export function useDeleteOrganization() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => organizations.delete(id),
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
    queryFn: () => members.list(orgId!),
    enabled: orgId !== undefined && orgId !== '',
    ...options,
  });
}

export function useUpdateMemberRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ orgId, memberId, request }: { orgId: string; memberId: string; request: UpdateMemberRoleRequest }) =>
      members.updateRole(orgId, memberId, request),
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
      members.remove(orgId, memberId),
    onSuccess: (_, { orgId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizationMembers(orgId) });
    },
  });
}

export function useSuspendMember() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ orgId, memberId }: { orgId: string; memberId: string }) =>
      members.suspend(orgId, memberId),
    onSuccess: (_, { orgId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizationMembers(orgId) });
    },
  });
}

export function useReinstateMember() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ orgId, memberId }: { orgId: string; memberId: string }) =>
      members.reinstate(orgId, memberId),
    onSuccess: (_, { orgId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizationMembers(orgId) });
    },
  });
}

export function useTransferOwnership() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ orgId, memberId }: { orgId: string; memberId: string }) =>
      members.transferOwnership(orgId, memberId),
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
    queryFn: () => invitations.listByOrganization(orgId!, status),
    enabled: orgId !== undefined && orgId !== '',
    ...options,
  });
}

export function useCreateInvitation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateInvitationInput) => invitations.create(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
    },
  });
}

export function useAcceptInvitation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ token, request }: { token: string; request?: InvitationResponseRequest }) =>
      invitations.accept(token, request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organizations });
      queryClient.invalidateQueries({ queryKey: queryKeys.myInvitations });
    },
  });
}

export function useRejectInvitation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ token, request }: { token: string; request?: InvitationResponseRequest }) =>
      invitations.reject(token, request),
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
