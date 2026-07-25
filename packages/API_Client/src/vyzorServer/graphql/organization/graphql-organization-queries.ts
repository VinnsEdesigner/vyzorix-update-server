import { gql } from "graphql-request";
import { ORGANIZATION_FRAGMENT, MEMBERSHIP_FRAGMENT, INVITATION_FRAGMENT } from "./graphql-organization-types";

export const GET_ORGANIZATIONS = gql`
  ${ORGANIZATION_FRAGMENT}
  query GetOrganizations($page: Int, $limit: Int) {
    organizations(page: $page, limit: $limit) {
      items {
        ...OrganizationFields
      }
      pagination {
        page
        limit
        total
        totalPages
        hasMore
      }
    }
  }
`;

export const GET_ORGANIZATION = gql`
  ${ORGANIZATION_FRAGMENT}
  query GetOrganization($id: ID!) {
    organization(id: $id) {
      ...OrganizationFields
    }
  }
`;

export const GET_MY_MEMBERSHIPS = gql`
  ${MEMBERSHIP_FRAGMENT}
  query GetMyMemberships($page: Int, $limit: Int) {
    myMemberships(page: $page, limit: $limit) {
      items {
        ...MembershipFields
      }
      pagination {
        page
        limit
        total
        totalPages
        hasMore
      }
    }
  }
`;

export const GET_ORGANIZATION_MEMBERS = gql`
  ${MEMBERSHIP_FRAGMENT}
  query GetOrganizationMembers($organizationId: ID!, $page: Int, $limit: Int) {
    organizationMembers(organizationId: $organizationId, page: $page, limit: $limit) {
      items {
        ...MembershipFields
      }
      pagination {
        page
        limit
        total
        totalPages
        hasMore
      }
    }
  }
`;

export const GET_ORGANIZATION_INVITATIONS = gql`
  ${INVITATION_FRAGMENT}
  query GetOrganizationInvitations($organizationId: ID!, $page: Int, $limit: Int) {
    organizationInvitations(organizationId: $organizationId, page: $page, limit: $limit) {
      items {
        ...InvitationFields
      }
      pagination {
        page
        limit
        total
        totalPages
        hasMore
      }
    }
  }
`;

export const GET_INVITATION_BY_TOKEN = gql`
  ${INVITATION_FRAGMENT}
  query GetInvitationByToken($token: String!) {
    invitationByToken(token: $token) {
      ...InvitationFields
    }
  }
`;
