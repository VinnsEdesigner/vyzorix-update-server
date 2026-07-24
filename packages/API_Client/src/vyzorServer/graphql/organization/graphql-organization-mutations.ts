import { gql } from "graphql-request";
import { ORGANIZATION_FRAGMENT, MEMBERSHIP_FRAGMENT, INVITATION_FRAGMENT } from "./graphql-organization-types";

export const CREATE_ORGANIZATION = gql`
  ${ORGANIZATION_FRAGMENT}
  mutation CreateOrganization($name: String!, $maxMembers: Int) {
    createOrganization(name: $name, maxMembers: $maxMembers) {
      ...OrganizationFields
    }
  }
`;

export const UPDATE_ORGANIZATION = gql`
  ${ORGANIZATION_FRAGMENT}
  mutation UpdateOrganization($id: ID!, $name: String, $maxMembers: Int, $isActive: Boolean) {
    updateOrganization(id: $id, name: $name, maxMembers: $maxMembers, isActive: $isActive) {
      ...OrganizationFields
    }
  }
`;

export const DELETE_ORGANIZATION = gql`
  mutation DeleteOrganization($id: ID!) {
    deleteOrganization(id: $id)
  }
`;

export const INVITE_MEMBER = gql`
  ${INVITATION_FRAGMENT}
  mutation InviteMember($organizationId: ID!, $email: String!, $role: OrgRole!, $notes: String) {
    inviteMember(organizationId: $organizationId, email: $email, role: $role, notes: $notes) {
      ...InvitationFields
    }
  }
`;

export const REMOVE_MEMBER = gql`
  mutation RemoveMember($organizationId: ID!, $memberId: ID!) {
    removeMember(organizationId: $organizationId, memberId: $memberId)
  }
`;

export const UPDATE_MEMBER_ROLE = gql`
  ${MEMBERSHIP_FRAGMENT}
  mutation UpdateMemberRole($organizationId: ID!, $memberId: ID!, $role: OrgRole!) {
    updateMemberRole(organizationId: $organizationId, memberId: $memberId, role: $role) {
      ...MembershipFields
    }
  }
`;

export const ACCEPT_INVITATION = gql`
  ${MEMBERSHIP_FRAGMENT}
  mutation AcceptInvitation($token: String!, $notes: String) {
    acceptInvitation(token: $token, notes: $notes) {
      ...MembershipFields
    }
  }
`;

export const REJECT_INVITATION = gql`
  mutation RejectInvitation($token: String!, $notes: String) {
    rejectInvitation(token: $token, notes: $notes)
  }
`;

export const CANCEL_INVITATION = gql`
  mutation CancelInvitation($id: ID!) {
    cancelInvitation(id: $id)
  }
`;

export const TRANSFER_OWNERSHIP = gql`
  mutation TransferOwnership($organizationId: ID!, $memberId: ID!) {
    transferOwnership(organizationId: $organizationId, memberId: $memberId)
  }
`;

export const SUSPEND_MEMBER = gql`
  mutation SuspendMember($organizationId: ID!, $memberId: ID!) {
    suspendMember(organizationId: $organizationId, memberId: $memberId)
  }
`;

export const REINSTATE_MEMBER = gql`
  mutation ReinstateMember($organizationId: ID!, $memberId: ID!) {
    reinstateMember(organizationId: $organizationId, memberId: $memberId)
  }
`;

export const TRANSFER_DEVICE = gql`
  mutation TransferDevice($imei: String!, $sourceOrganizationId: ID!, $targetOrganizationId: ID!) {
    transferDevice(imei: $imei, sourceOrganizationId: $sourceOrganizationId, targetOrganizationId: $targetOrganizationId) {
      success
      device {
        id
        imei
      }
    }
  }
`;
