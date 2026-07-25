import { graphqlClient } from '../_shared/graphql-client';
import { ORGANIZATION_FRAGMENT, MEMBERSHIP_FRAGMENT, INVITATION_FRAGMENT } from './graphql-organization-types';

export const CREATE_ORGANIZATION = `
  ${ORGANIZATION_FRAGMENT}
  mutation CreateOrganization($name: String!, $maxMembers: Int) {
    createOrganization(name: $name, maxMembers: $maxMembers) {
      organization {
        ...OrganizationFields
      }
      membership {
        ...MembershipFields
      }
    }
  }
`;

export const UPDATE_ORGANIZATION = `
  ${ORGANIZATION_FRAGMENT}
  mutation UpdateOrganization($id: ID!, $name: String, $maxMembers: Int, $isActive: Boolean) {
    updateOrganization(id: $id, name: $name, maxMembers: $maxMembers, isActive: $isActive) {
      ...OrganizationFields
    }
  }
`;

export const DELETE_ORGANIZATION = `
  mutation DeleteOrganization($id: ID!) {
    deleteOrganization(id: $id)
  }
`;

export const INVITE_MEMBER = `
  ${INVITATION_FRAGMENT}
  mutation InviteMember($organizationId: ID!, $email: String!, $role: OrgRole!, $notes: String) {
    inviteMember(organizationId: $organizationId, email: $email, role: $role, notes: $notes) {
      ...InvitationFields
    }
  }
`;

export const REMOVE_MEMBER = `
  mutation RemoveMember($organizationId: ID!, $memberId: ID!) {
    removeMember(organizationId: $organizationId, memberId: $memberId)
  }
`;

export const UPDATE_MEMBER_ROLE = `
  ${MEMBERSHIP_FRAGMENT}
  mutation UpdateMemberRole($organizationId: ID!, $memberId: ID!, $role: OrgRole!) {
    updateMemberRole(organizationId: $organizationId, memberId: $memberId, role: $role) {
      ...MembershipFields
    }
  }
`;

export const ACCEPT_INVITATION = `
  ${MEMBERSHIP_FRAGMENT}
  mutation AcceptInvitation($token: String!, $notes: String) {
    acceptInvitation(token: $token, notes: $notes) {
      ...MembershipFields
    }
  }
`;

export const REJECT_INVITATION = `
  mutation RejectInvitation($token: String!, $notes: String) {
    rejectInvitation(token: $token, notes: $notes)
  }
`;

export const CANCEL_INVITATION = `
  mutation CancelInvitation($id: ID!) {
    cancelInvitation(id: $id)
  }
`;

export const TRANSFER_OWNERSHIP = `
  mutation TransferOwnership($organizationId: ID!, $memberId: ID!) {
    transferOwnership(organizationId: $organizationId, memberId: $memberId)
  }
`;

export const SUSPEND_MEMBER = `
  mutation SuspendMember($organizationId: ID!, $memberId: ID!) {
    suspendMember(organizationId: $organizationId, memberId: $memberId)
  }
`;

export const REINSTATE_MEMBER = `
  mutation ReinstateMember($organizationId: ID!, $memberId: ID!) {
    reinstateMember(organizationId: $organizationId, memberId: $memberId)
  }
`;

export const TRANSFER_DEVICE = `
  mutation TransferDevice($imei: String!, $sourceOrganizationId: ID!, $targetOrganizationId: ID!) {
    transferDevice(imei: $imei, sourceOrganizationId: $sourceOrganizationId, targetOrganizationId: $targetOrganizationId) {
      success
      deviceId
      sourceOrganizationId
      targetOrganizationId
    }
  }
`;

export async function mutateCreateOrganization(params: { name: string; maxMembers?: number }) {
  return graphqlClient.getClient().mutate({
    mutation: CREATE_ORGANIZATION,
    variables: params,
  });
}

export async function mutateUpdateOrganization(params: { id: string; name?: string; maxMembers?: number; isActive?: boolean }) {
  return graphqlClient.getClient().mutate({
    mutation: UPDATE_ORGANIZATION,
    variables: params,
  });
}

export async function mutateDeleteOrganization(params: { id: string }) {
  return graphqlClient.getClient().mutate({
    mutation: DELETE_ORGANIZATION,
    variables: params,
  });
}

export async function mutateInviteMember(params: { organizationId: string; email: string; role: string; notes?: string }) {
  return graphqlClient.getClient().mutate({
    mutation: INVITE_MEMBER,
    variables: params,
  });
}

export async function mutateRemoveMember(params: { organizationId: string; memberId: string }) {
  return graphqlClient.getClient().mutate({
    mutation: REMOVE_MEMBER,
    variables: params,
  });
}

export async function mutateUpdateMemberRole(params: { organizationId: string; memberId: string; role: string }) {
  return graphqlClient.getClient().mutate({
    mutation: UPDATE_MEMBER_ROLE,
    variables: params,
  });
}

export async function mutateAcceptInvitation(params: { token: string; notes?: string }) {
  return graphqlClient.getClient().mutate({
    mutation: ACCEPT_INVITATION,
    variables: params,
  });
}

export async function mutateRejectInvitation(params: { token: string; notes?: string }) {
  return graphqlClient.getClient().mutate({
    mutation: REJECT_INVITATION,
    variables: params,
  });
}

export async function mutateCancelInvitation(params: { id: string }) {
  return graphqlClient.getClient().mutate({
    mutation: CANCEL_INVITATION,
    variables: params,
  });
}

export async function mutateTransferOwnership(params: { organizationId: string; memberId: string }) {
  return graphqlClient.getClient().mutate({
    mutation: TRANSFER_OWNERSHIP,
    variables: params,
  });
}

export async function mutateSuspendMember(params: { organizationId: string; memberId: string }) {
  return graphqlClient.getClient().mutate({
    mutation: SUSPEND_MEMBER,
    variables: params,
  });
}

export async function mutateReinstateMember(params: { organizationId: string; memberId: string }) {
  return graphqlClient.getClient().mutate({
    mutation: REINSTATE_MEMBER,
    variables: params,
  });
}

export async function mutateTransferDevice(params: { imei: string; sourceOrganizationId: string; targetOrganizationId: string }) {
  return graphqlClient.getClient().mutate({
    mutation: TRANSFER_DEVICE,
    variables: params,
  });
}
