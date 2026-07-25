import { graphqlClient } from '../_shared/graphql-client';

export const CREATE_ORGANIZATION = `
  mutation CreateOrganization($name: String!, $maxMembers: Int) {
    createOrganization(name: $name, maxMembers: $maxMembers) {
      organization {
        id
        name
        lifecycle
        maxMembers
        memberCount
        createdAt
        updatedAt
        createdBy
      }
      membership {
        id
        organizationId
        operatorId
        role
        lifecycle
        joinedAt
        operator {
          id
          email
          name
        }
        organization {
          id
          name
        }
      }
    }
  }
`;

export const UPDATE_ORGANIZATION = `
  mutation UpdateOrganization($id: ID!, $name: String, $maxMembers: Int, $isActive: Boolean) {
    updateOrganization(id: $id, name: $name, maxMembers: $maxMembers, isActive: $isActive) {
      id
      name
      lifecycle
      maxMembers
      memberCount
      createdAt
      updatedAt
      deletedAt
      createdBy
    }
  }
`;

export const DELETE_ORGANIZATION = `
  mutation DeleteOrganization($id: ID!) {
    deleteOrganization(id: $id)
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

export async function mutateTransferDevice(params: { imei: string; sourceOrganizationId: string; targetOrganizationId: string }) {
  return graphqlClient.getClient().mutate({
    mutation: TRANSFER_DEVICE,
    variables: params,
  });
}
