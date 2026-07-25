import { gql } from "@apollo/client";

export const GET_ORGANIZATIONS = gql`
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
  query GetOrganization($id: ID!) {
    organization(id: $id) {
      ...OrganizationFields
    }
  }
`;

export const GET_MY_MEMBERSHIPS = gql`
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
  query GetInvitationByToken($token: String!) {
    invitationByToken(token: $token) {
      ...InvitationFields
    }
  }
`;
